package toolkitupdate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestLatestRejectsLegacyAndChecksManifest(t *testing.T) {
	manifest := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/latest" {
			assets := "[]"
			if manifest {
				assets = `[{"name":"my-friday-update.json","browser_download_url":"https://github.com/acoz-labs/my-friday/releases/download/v1/my-friday-update.json"},{"name":"my-friday-test","browser_download_url":"https://github.com/acoz-labs/my-friday/releases/download/v1/my-friday-test"}]`
			}
			fmt.Fprintf(w, `{"tag_name":"v1","assets":%s}`, assets)
		} else {
			fmt.Fprintf(w, `{"schema_version":1,"portable_format":1,"management_protocol":1,"version":"v1","artifacts":[{"os":%q,"arch":%q,"name":"my-friday-test","sha256":%q}]}`, runtime.GOOS, runtime.GOARCH, strings.Repeat("a", 64))
		}
	}))
	defer server.Close()
	c := Client{get: func(ctx context.Context, url string, limit int64) ([]byte, error) {
		target := server.URL + "/manifest"
		if strings.Contains(url, "api.github.com") {
			target = server.URL + "/latest"
		}
		return fetch(ctx, server.Client(), target, limit)
	}}
	if _, err := c.Latest(context.Background()); err == nil {
		t.Fatal("accepted legacy release")
	}
	manifest = true
	a, err := c.Latest(context.Background())
	if err != nil || a.Version != "v1" {
		t.Fatalf("%+v %v", a, err)
	}
}

func TestStageIntegrityAndPointerPreservation(t *testing.T) {
	home := t.TempDir()
	source := filepath.Join(t.TempDir(), "binary")
	os.WriteFile(source, []byte("fixture artifact"), 0700)
	h := sha256.Sum256([]byte("fixture artifact"))
	digest := hex.EncodeToString(h[:])
	if _, err := Stage(home, source, strings.Repeat("0", 64)); err == nil {
		t.Fatal("accepted wrong digest")
	}
	installed, err := Stage(home, source, digest)
	if err != nil {
		t.Fatal(err)
	}
	again, err := Stage(home, source, digest)
	if err != nil || again != installed {
		t.Fatal("not idempotent")
	}
	link := filepath.Join(home, ".local/bin/my-friday")
	os.MkdirAll(filepath.Dir(link), 0700)
	os.WriteFile(link, []byte("custom command"), 0700)
	if _, err := Activate(home, installed, "/previous"); err == nil {
		t.Fatal("overwrote custom command")
	}
	data, _ := os.ReadFile(link)
	if string(data) != "custom command" {
		t.Fatal("modified user command")
	}
}

func TestDownloadVerifiesBytesAndRefusesWrongOrigins(t *testing.T) {
	data := []byte("approved binary")
	sum := sha256.Sum256(data)
	a := Artifact{Name: "my-friday-test", Version: "v1", URL: releasePrefix + "v1/my-friday-test", SHA256: hex.EncodeToString(sum[:])}
	calls := 0
	c := Client{get: func(context.Context, string, int64) ([]byte, error) { calls++; return data, nil }}
	home := t.TempDir()
	if _, err := c.Download(context.Background(), home, a); err != nil {
		t.Fatal(err)
	}
	a.SHA256 = strings.Repeat("0", 64)
	if _, err := c.Download(context.Background(), home, a); err == nil {
		t.Fatal("accepted wrong bytes")
	}
	a.URL = "https://unrelated.example/binary"
	before := calls
	if _, err := c.Download(context.Background(), home, a); err == nil || calls != before {
		t.Fatal("requested foreign URL")
	}
}

func TestFetchIsBoundedAndDoesNotSendCredentials(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" || r.Header.Get("Cookie") != "" {
			t.Error("sent credentials")
		}
		if r.URL.Path == "/slow" {
			<-r.Context().Done()
			return
		}
		fmt.Fprint(w, "too large")
	}))
	defer server.Close()
	if _, err := fetch(context.Background(), server.Client(), server.URL, 2); err == nil {
		t.Fatal("unbounded read")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if _, err := fetch(ctx, server.Client(), server.URL+"/slow", 1024); err == nil {
		t.Fatal("ignored cancellation")
	}
}

func TestActivationRetainsPriorPointerAndRejectsRedirectedDirectories(t *testing.T) {
	home := t.TempDir()
	home, _ = filepath.EvalSymlinks(home)
	previous := filepath.Join(t.TempDir(), "previous")
	next := filepath.Join(t.TempDir(), "next")
	os.WriteFile(previous, []byte("old"), 0700)
	os.WriteFile(next, []byte("new"), 0700)
	previous, _ = filepath.EvalSymlinks(previous)
	next, _ = filepath.EvalSymlinks(next)
	os.MkdirAll(filepath.Join(home, ".local/bin"), 0700)
	link := filepath.Join(home, ".local/bin/my-friday")
	os.Symlink(previous, link)
	backup, err := Activate(home, next, previous)
	if err != nil {
		t.Fatal(err)
	}
	if target, _ := filepath.EvalSymlinks(link); target != next {
		t.Fatal("activation missed")
	}
	if target, _ := filepath.EvalSymlinks(filepath.Join(backup, "previous-my-friday")); target != previous {
		t.Fatal("backup lost")
	}
	another := filepath.Join(t.TempDir(), "another")
	os.WriteFile(another, []byte("another"), 0700)
	if _, err := Activate(home, another, previous); err == nil {
		t.Fatal("stale menu overwrote current pointer")
	}
	redirectHome := t.TempDir()
	external := t.TempDir()
	os.Symlink(external, filepath.Join(redirectHome, ".local"))
	hash, _ := Digest(next)
	if _, err := Stage(redirectHome, next, hash); err == nil {
		t.Fatal("followed managed symlink")
	}
	entries, _ := os.ReadDir(external)
	if len(entries) != 0 {
		t.Fatal("created files through symlink before rejecting")
	}
}
