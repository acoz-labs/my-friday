package main

import (
	"crypto/sha256"
	"fmt"
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

func TestNoFollowReadRejectsIntermediateSymlink(t *testing.T) {
	root := t.TempDir()
	realDir := filepath.Join(root, "real")
	if err := os.Mkdir(realDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(realDir, "protected"), []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(realDir, filepath.Join(root, "link")); err != nil {
		t.Fatal(err)
	}
	if _, ok, err := noFollowRead(filepath.Join(root, "link", "protected")); err == nil || ok {
		t.Fatal("intermediate symlink in protected path was accepted")
	}
}

func TestResolveExecutableAcceptsOwnedSymlinkChain(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "codex-real")
	current := filepath.Join(root, "current")
	link := filepath.Join(root, "codex")
	if err := os.WriteFile(target, []byte("binary"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, current); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(current, link); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("go", "run", ".", "resolve-executable", link)
	out, err := cmd.CombinedOutput()
	expected, _ := filepath.EvalSymlinks(target)
	if err != nil || strings.TrimSpace(string(out)) != expected {
		t.Fatalf("safe installed symlink was not resolved: %v %s", err, out)
	}
}

func TestRenderProfileDoesNotInterpretReplacementMetacharacters(t *testing.T) {
	root := t.TempDir()
	template := filepath.Join(root, "profile.in")
	if err := os.WriteFile(template, []byte("(subpath @@VOLUME@@)\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	value := "/tmp/a&b|c\\d\nq\"r"
	cmd := exec.Command("go", "run", ".", "render-profile", "--template", template, "--value", value)
	out, err := cmd.CombinedOutput()
	if err != nil || string(out) != "(subpath "+strconv.Quote(value)+")\n" || strings.Contains(string(out), "@@VOLUME@@") {
		t.Fatalf("profile was not rendered literally: %v %q", err, out)
	}
}

func TestSecureRootsRejectsSymlinkedAncestor(t *testing.T) {
	root := t.TempDir()
	realParent := filepath.Join(root, "real")
	if err := os.Mkdir(realParent, 0o700); err != nil {
		t.Fatal(err)
	}
	linkParent := filepath.Join(root, "link")
	if err := os.Symlink(realParent, linkParent); err != nil {
		t.Fatal(err)
	}
	home := filepath.Join(linkParent, "home")
	if err := os.Mkdir(filepath.Join(realParent, "home"), 0o700); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("go", "run", ".", "secure-roots", "--home", home, "--run-id", "run")
	if err := cmd.Run(); err == nil {
		t.Fatal("secure roots accepted symlinked ancestry")
	}
}

func TestCleanupRootsIsReceiptAndMarkerBound(t *testing.T) {
	home := t.TempDir()
	secure := exec.Command("go", "run", ".", "secure-roots", "--home", home, "--run-id", "run")
	receipt, err := secure.CombinedOutput()
	if err != nil {
		t.Fatalf("secure roots: %v %s", err, receipt)
	}
	marker := []byte("marker\n")
	markerSHA := fmt.Sprintf("%x", sha256.Sum256(marker))
	for _, parent := range []string{".my-friday-acceptance", ".my-friday-acceptance-evidence"} {
		child := filepath.Join(home, parent, "run")
		if err = os.WriteFile(filepath.Join(child, "marker.json"), marker, 0o600); err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(filepath.Join(child, "owned"), []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	cleanup := exec.Command("go", "run", ".", "cleanup-roots", "--home", home, "--run-id", "run", "--receipt", strings.TrimSpace(string(receipt)), "--marker-sha256", markerSHA,
		"--expected-entry", ".my-friday-acceptance:marker.json", "--expected-entry", ".my-friday-acceptance:owned",
		"--expected-entry", ".my-friday-acceptance-evidence:marker.json", "--expected-entry", ".my-friday-acceptance-evidence:owned")
	if out, runErr := cleanup.CombinedOutput(); runErr != nil {
		t.Fatalf("cleanup: %v %s", runErr, out)
	}
	for _, parent := range []string{".my-friday-acceptance", ".my-friday-acceptance-evidence"} {
		if _, err = os.Stat(filepath.Join(home, parent, "run")); !os.IsNotExist(err) {
			t.Fatal("run child survived cleanup")
		}
		if _, err = os.Stat(filepath.Join(home, parent)); err != nil {
			t.Fatal("fixed parent was removed")
		}
	}
}

func TestCleanupRootsPreservesUnexpectedEntry(t *testing.T) {
	home := t.TempDir()
	receipt, err := exec.Command("go", "run", ".", "secure-roots", "--home", home, "--run-id", "run").CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	marker := []byte("marker\n")
	markerSHA := fmt.Sprintf("%x", sha256.Sum256(marker))
	for _, parent := range []string{".my-friday-acceptance", ".my-friday-acceptance-evidence"} {
		child := filepath.Join(home, parent, "run")
		if err = os.WriteFile(filepath.Join(child, "marker.json"), marker, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	unexpected := filepath.Join(home, ".my-friday-acceptance", "run", "unexpected")
	if err = os.WriteFile(unexpected, []byte("preserve"), 0o600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("go", "run", ".", "cleanup-roots", "--home", home, "--run-id", "run", "--receipt", strings.TrimSpace(string(receipt)), "--marker-sha256", markerSHA,
		"--expected-entry", ".my-friday-acceptance:marker.json", "--expected-entry", ".my-friday-acceptance-evidence:marker.json")
	if err = cmd.Run(); err == nil {
		t.Fatal("cleanup accepted an unexpected entry")
	}
	if body, readErr := os.ReadFile(unexpected); readErr != nil || string(body) != "preserve" {
		t.Fatal("unexpected entry was not preserved")
	}
	if body, readErr := os.ReadFile(filepath.Join(home, ".my-friday-acceptance", "run", "marker.json")); readErr != nil || string(body) != "marker\n" {
		t.Fatal("cleanup mutated expected state before refusing the unexpected entry")
	}
}

func TestCleanupRootsPrevalidatesBothRootsBeforeMutation(t *testing.T) {
	home := t.TempDir()
	receipt, err := exec.Command("go", "run", ".", "secure-roots", "--home", home, "--run-id", "run").CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	marker := []byte("marker\n")
	markerSHA := fmt.Sprintf("%x", sha256.Sum256(marker))
	for _, parent := range []string{".my-friday-acceptance", ".my-friday-acceptance-evidence"} {
		if err = os.WriteFile(filepath.Join(home, parent, "run", "marker.json"), marker, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	unexpected := filepath.Join(home, ".my-friday-acceptance-evidence", "run", "unexpected")
	if err = os.WriteFile(unexpected, []byte("preserve"), 0o600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("go", "run", ".", "cleanup-roots", "--home", home, "--run-id", "run", "--receipt", strings.TrimSpace(string(receipt)), "--marker-sha256", markerSHA,
		"--expected-entry", ".my-friday-acceptance:marker.json", "--expected-entry", ".my-friday-acceptance-evidence:marker.json")
	if err = cmd.Run(); err == nil {
		t.Fatal("cleanup accepted an evidence-root surprise")
	}
	for _, parent := range []string{".my-friday-acceptance", ".my-friday-acceptance-evidence"} {
		if body, readErr := os.ReadFile(filepath.Join(home, parent, "run", "marker.json")); readErr != nil || string(body) != "marker\n" {
			t.Fatalf("cleanup mutated %s before both roots validated", parent)
		}
	}
}

func TestSandboxDiagnosticAllowlistV1(t *testing.T) {
	exact := "sandbox-exec: warning: sandbox-exec is deprecated and will be removed in a future release."
	for name, test := range map[string]struct {
		input string
		want  bool
	}{
		"empty": {"", true}, "exact": {exact, true}, "suffix": {exact + " attacker", false},
		"multiline": {exact + "\nunexpected", false}, "duplicate": {exact + "\n" + exact, false},
	} {
		t.Run(name, func(t *testing.T) {
			if got := validSandboxDiagnostic("v1", test.input); got != test.want {
				t.Fatalf("got %v want %v", got, test.want)
			}
		})
	}
}
