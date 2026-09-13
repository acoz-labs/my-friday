package memoryinstall

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/acoz-labs/my-friday/internal/memorybank"
	"github.com/acoz-labs/my-friday/internal/portable"
)

func fixture(t *testing.T) Options {
	t.Helper()
	home := t.TempDir()
	bank, err := portable.CreateMemoryBank(filepath.Join(home, "bank"), "Fixture", "device-fixture", "Host")
	if err != nil {
		t.Fatal(err)
	}
	binding := filepath.Join(home, "binding with 'quote.json")
	if _, err := memorybank.Bind(bank.Root, binding, "Host", "Writer"); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(home, "runtime with 'quote")
	if err := os.WriteFile(binary, []byte(fmt.Sprintf("#!/bin/sh\nprintf '%%s\\n' '{\"product\":\"my-friday\",\"memory_protocol\":1,\"os\":%q,\"arch\":%q}'\n", runtime.GOOS, runtime.GOARCH)), 0700); err != nil {
		t.Fatal(err)
	}
	return Options{Home: home, CodexHome: filepath.Join(home, "native"), Codex: "/unused/codex", Binary: binary, Binding: binding}
}

func installedRunner(p Plan, failInstall bool) runner {
	registered := false
	return func(ctx context.Context, o Options, args ...string) ([]byte, error) {
		switch strings.Join(args[:3], " ") {
		case "plugin marketplace list":
			if registered {
				return json.Marshal(map[string]any{"marketplaces": []any{map[string]any{"name": p.Marketplace, "root": p.Root, "marketplaceSource": map[string]string{"sourceType": "local", "source": p.Root}}}})
			}
			return []byte(`{"marketplaces":[]}`), nil
		case "plugin marketplace add":
			registered = true
			return []byte(`{}`), nil
		}
		if args[1] == "add" {
			if failInstall {
				return nil, errors.New("synthetic native failure")
			}
			return []byte(`{}`), nil
		}
		if registered {
			return json.Marshal(map[string]any{"installed": []any{map[string]any{"pluginId": p.PluginID, "name": "my-friday-memory", "version": p.Version, "enabled": true, "marketplaceSource": map[string]string{"source": p.Root}}}})
		}
		return []byte(`{"installed":[]}`), nil
	}
}

func TestApplyDoctorAndIncompleteInstall(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(fmt.Sprint(fail), func(t *testing.T) {
			p, err := Prepare(fixture(t))
			if err != nil {
				t.Fatal(err)
			}
			run := installedRunner(p, fail)
			result, err := apply(context.Background(), p, run)
			if fail {
				if err == nil || result.Installed || !strings.Contains(err.Error(), "incomplete") {
					t.Fatalf("false success: %+v %v", result, err)
				}
				return
			}
			if err != nil || !result.Installed {
				t.Fatalf("%+v %v", result, err)
			}
			files, _ := bundle(p)
			for name, data := range files {
				if !strings.HasPrefix(name, "plugins/my-friday-memory/") {
					continue
				}
				path := filepath.Join(p.CodexHome, "plugins/cache", p.Marketplace, "my-friday-memory", p.Version, strings.TrimPrefix(name, "plugins/my-friday-memory/"))
				if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, data, 0600); err != nil {
					t.Fatal(err)
				}
			}
			if r := doctor(context.Background(), p.Options, run); !r.Healthy {
				t.Fatalf("%+v", r)
			}
			if err := os.WriteFile(filepath.Join(p.Root, "plugins/my-friday-memory/.mcp.json"), []byte("changed"), 0600); err != nil {
				t.Fatal(err)
			}
			if r := doctor(context.Background(), p.Options, run); r.Healthy {
				t.Fatal("doctor ignored modified projection")
			}
		})
	}
}

func TestApplyRejectsChangedPreviewAndLegacyRuntime(t *testing.T) {
	o := fixture(t)
	p, err := Prepare(o)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(o.Binary, []byte("#!/bin/sh\nprintf '{}\\n'\n"), 0700); err != nil {
		t.Fatal(err)
	}
	never := func(context.Context, Options, ...string) ([]byte, error) {
		t.Fatal("executed native command for changed preview")
		return nil, nil
	}
	if _, err := apply(context.Background(), p, never); err == nil {
		t.Fatal("accepted changed preview")
	}
	p, err = Prepare(o)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := apply(context.Background(), p, installedRunner(p, false)); err == nil || !strings.Contains(err.Error(), "memory protocol") {
		t.Fatalf("accepted legacy runtime: %v", err)
	}
}

func TestNativeCancellationAndOutputBound(t *testing.T) {
	o := fixture(t)
	o.Codex = filepath.Join(o.Home, "fake-codex")
	if err := os.WriteFile(o.Codex, []byte("#!/bin/sh\nsleep 20\n"), 0700); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	start := time.Now()
	if _, err := native(ctx, o, "plugin", "list", "--json"); err == nil {
		t.Fatal("ignored cancellation")
	}
	if time.Since(start) > 3*time.Second {
		t.Fatal("cancellation did not bound descendants")
	}
	var b boundedOutput
	if _, err := b.Write(make([]byte, (1<<20)+1)); err == nil || b.Len() != 0 {
		t.Fatal("output bound failed")
	}
}

func TestPlanRefusesRedirectedManagedDirectory(t *testing.T) {
	o := fixture(t)
	if err := os.Symlink(t.TempDir(), filepath.Join(o.Home, ".local")); err != nil {
		t.Fatal(err)
	}
	if _, err := Prepare(o); err == nil {
		t.Fatal("followed managed path redirection")
	}
}

func TestGeneratedConnectionShellQuotesAndOverrides(t *testing.T) {
	p, err := Prepare(fixture(t))
	if err != nil {
		t.Fatal(err)
	}
	if err = publish(p); err != nil {
		t.Fatal(err)
	}
	if err = os.MkdirAll(filepath.Dir(p.Runtime), 0700); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(p.Runtime, []byte("#!/bin/sh\nprintf '%s\\n' \"$MY_FRIDAY_MEMORY_BINDING\" \"$@\"\n"), 0700); err != nil {
		t.Fatal(err)
	}
	for _, override := range []string{"", "/literal/$(no-command)/binding's.json"} {
		cmd := exec.Command("/bin/sh", filepath.Join(p.Root, "plugins/my-friday-memory/scripts/connection.sh"), "mcp")
		cmd.Env = []string{"PATH=/usr/bin:/bin", "MY_FRIDAY_MEMORY_BINDING=" + override}
		got, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("%s %v", got, err)
		}
		want := override
		if want == "" {
			want = p.Binding
		}
		if string(got) != want+"\nmcp\n--harness\ncodex\n" {
			t.Fatalf("bad literal routing: %q", got)
		}
	}
}

func TestInterruptedRegistrationRetainsCopiesAndCanRetry(t *testing.T) {
	old, err := Prepare(fixture(t))
	if err != nil {
		t.Fatal(err)
	}
	if err = publish(old); err != nil {
		t.Fatal(err)
	}
	o := old.Options
	o.Generation = "repair-test"
	next, err := Prepare(o)
	if err != nil {
		t.Fatal(err)
	}
	registered := old.Root
	fail := true
	removed := 0
	run := func(ctx context.Context, o Options, args ...string) ([]byte, error) {
		if args[1] == "marketplace" {
			switch args[2] {
			case "list":
				if registered == "" {
					return []byte(`{"marketplaces":[]}`), nil
				}
				return json.Marshal(map[string]any{"marketplaces": []any{map[string]any{"name": next.Marketplace, "root": registered, "marketplaceSource": map[string]string{"sourceType": "local", "source": registered}}}})
			case "remove":
				registered = ""
				removed++
				return []byte(`{}`), nil
			case "add":
				if fail {
					return nil, errors.New("synthetic interruption")
				}
				registered = args[3]
				return []byte(`{}`), nil
			}
		}
		if args[1] == "list" {
			if registered != next.Root {
				return []byte(`{"installed":[]}`), nil
			}
			return json.Marshal(map[string]any{"installed": []any{map[string]any{"name": "my-friday-memory", "pluginId": next.PluginID, "version": next.Version, "enabled": true, "marketplaceSource": map[string]string{"source": next.Root}}}})
		}
		return []byte(`{}`), nil
	}
	r, err := apply(context.Background(), next, run)
	if err == nil || r.Installed || r.PreviousRoot != old.Root || registered != "" || removed != 1 {
		t.Fatalf("incorrect partial result: %+v %v", r, err)
	}
	if err = verifyBundle(old); err != nil {
		t.Fatal("previous bundle changed", err)
	}
	if err = verifyBundle(next); err != nil {
		t.Fatal("new bundle not retained", err)
	}
	fail = false
	if r, err = apply(context.Background(), next, run); err != nil || !r.Installed {
		t.Fatalf("retry failed: %+v %v", r, err)
	}
}

func TestPlanIsReadOnlyAndPinsRuntimeAndBinding(t *testing.T) {
	o := fixture(t)
	p, err := Prepare(o)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(p.Runtime, p.BinarySHA256) || p.BankID == "" || p.PluginID == "" {
		t.Fatalf("incomplete plan: %+v", p)
	}
	if _, err := os.Stat(filepath.Join(o.Home, ".local")); !os.IsNotExist(err) {
		t.Fatal("planning wrote installation")
	}
	files, err := bundle(p)
	if err != nil {
		t.Fatal(err)
	}
	wrapper := string(files["plugins/my-friday-memory/scripts/connection.sh"])
	if !strings.Contains(wrapper, "'\\''") || !strings.Contains(wrapper, p.Runtime) {
		t.Fatal(wrapper)
	}
	if !strings.Contains(string(files["plugins/my-friday-memory/.mcp.json"]), "connection.sh") {
		t.Fatal("MCP not pinned")
	}
	for _, v := range files {
		if strings.Contains(string(v), "OPENAI_API_KEY") {
			t.Fatal("copied auth")
		}
	}
}

func TestInstallationCollisionStopsBeforeStaging(t *testing.T) {
	o := fixture(t)
	p, err := Prepare(o)
	if err != nil {
		t.Fatal(err)
	}
	calls := 0
	run := func(ctx context.Context, o Options, args ...string) ([]byte, error) {
		calls++
		if args[1] == "marketplace" {
			return json.Marshal(map[string]any{"marketplaces": []any{map[string]any{"name": p.Marketplace, "root": "/someone/elses/marketplace"}}})
		}
		return []byte(`{"installed":[]}`), nil
	}
	if _, err := apply(context.Background(), p, run); err == nil {
		t.Fatal("overwrote another marketplace")
	}
	if calls != 1 {
		t.Fatalf("unexpected native calls: %d", calls)
	}
	if _, err := os.Stat(filepath.Join(o.Home, ".local")); !os.IsNotExist(err) {
		t.Fatal("collision staged files")
	}
}

func TestBundlePublicationIsImmutableAndDetectsEdits(t *testing.T) {
	p, err := Prepare(fixture(t))
	if err != nil {
		t.Fatal(err)
	}
	if err := publish(p); err != nil {
		t.Fatal(err)
	}
	if err := publish(p); err != nil {
		t.Fatal("idempotent publication", err)
	}
	file := filepath.Join(p.Root, "plugins/my-friday-memory/scripts/connection.sh")
	if err := os.WriteFile(file, []byte("preserve user edit"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := publish(p); err == nil {
		t.Fatal("accepted edited immutable bundle")
	}
	data, _ := os.ReadFile(file)
	if string(data) != "preserve user edit" {
		t.Fatal("overwrote edit")
	}
}

func TestPlanRefusesBindingOrManagedPathsInsideBank(t *testing.T) {
	o := fixture(t)
	o.Home = filepath.Join(o.Home, "bank")
	if _, err := Prepare(o); err == nil {
		t.Fatal("accepted installation inside bank")
	}
}
