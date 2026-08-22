package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/acoz-labs/my-friday/internal/codexhome"
)

func TestFixtureIsValidAndRendersToken(t *testing.T) {
	root := t.TempDir()
	runtime := filepath.Join(root, "runtime")
	memory := filepath.Join(root, "memory")
	cmd := exec.Command("go", "run", ".", "fixture", "--runtime", runtime, "--memory", memory, "--token", "TOKEN_ONE")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("fixture: %v: %s", err, out)
	}
	codex := filepath.Join(root, "home", ".codex")
	if err := os.MkdirAll(codex, 0o700); err != nil {
		t.Fatal(err)
	}
	p, err := codexhome.Plan(codexhome.ActionInstall, runtime, codex)
	if err != nil {
		t.Fatal(err)
	}
	if err = codexhome.Execute(p); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join(codex, "AGENTS.md"))
	if err != nil || !strings.Contains(string(body), "TOKEN_ONE") {
		t.Fatalf("rendered token missing: %v %q", err, body)
	}
	cmd = exec.Command("go", "run", ".", "update", "--runtime", runtime, "--token", "TOKEN_TWO")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("update: %v: %s", err, out)
	}
	p, err = codexhome.Plan(codexhome.ActionUpgrade, runtime, codex)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(p.String(), "replace") {
		t.Fatal("valid profile update did not plan an upgrade")
	}
}

func TestSchemeStringEscapesQuotesBackslashesAndNewlines(t *testing.T) {
	got := strconv.Quote("/tmp/a\"b\\c\n")
	if got != `"/tmp/a\"b\\c\n"` {
		t.Fatalf("unsafe Scheme string: %s", got)
	}
}

func TestSecureRootsRefusesCollision(t *testing.T) {
	home := t.TempDir()
	cmd := exec.Command("go", "run", ".", "secure-roots", "--home", home, "--run-id", "run")
	if out, err := cmd.CombinedOutput(); err != nil || !strings.Contains(string(out), "device") {
		t.Fatalf("secure roots: %v %s", err, out)
	}
	cmd = exec.Command("go", "run", ".", "secure-roots", "--home", home, "--run-id", "run")
	if err := cmd.Run(); err == nil {
		t.Fatal("secure root collision was accepted")
	}
}

func TestNoFollowReadRejectsSymlink(t *testing.T) {
	root := t.TempDir()
	secret := filepath.Join(root, "secret")
	link := filepath.Join(root, "link")
	if err := os.WriteFile(secret, []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(secret, link); err != nil {
		t.Fatal(err)
	}
	if _, ok, err := noFollowRead(link); err == nil || ok {
		t.Fatal("symlinked protected content was accepted")
	}
}
