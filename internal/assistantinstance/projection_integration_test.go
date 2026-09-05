package assistantinstance_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/acoz-labs/my-friday/internal/assistantinstance"
	bootstrap "github.com/acoz-labs/my-friday/internal/plan"
	"github.com/acoz-labs/my-friday/internal/profile"
	"github.com/acoz-labs/my-friday/internal/repository"
)

func TestForgedManagedInstructionsManifestIsDenied(t *testing.T) {
	for _, mutate := range []func(*assistantinstance.Manifest){
		func(m *assistantinstance.Manifest) { m.CodexInstructions = filepath.Join(m.Root, "codex", "other.md") },
		func(m *assistantinstance.Manifest) { m.CodexInstructionsSHA256 = strings.Repeat("a", 64) },
	} {
		home := filepath.Join(t.TempDir(), "home")
		if err := os.MkdirAll(filepath.Join(home, ".local", "bin"), 0o755); err != nil {
			t.Fatal(err)
		}
		executable, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		codex := filepath.Join(home, "codex-stub")
		if err = os.WriteFile(codex, []byte("codex"), 0o700); err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(filepath.Join(home, "codex-code-mode-host"), []byte("host"), 0o700); err != nil {
			t.Fatal(err)
		}
		paths := createNamedWithProfile(t, home, "alfred", executable, codex, filepath.Join(home, "source"), "PURPOSE_BOUND")
		manifestPath := filepath.Join(paths.Root, "manifest.json")
		body, err := os.ReadFile(manifestPath)
		if err != nil {
			t.Fatal(err)
		}
		var m assistantinstance.Manifest
		if err = json.Unmarshal(body, &m); err != nil {
			t.Fatal(err)
		}
		mutate(&m)
		body, err = json.MarshalIndent(m, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(manifestPath, append(body, '\n'), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err = assistantinstance.Verify(home, "alfred"); err == nil {
			t.Fatal("forged managed-instructions manifest verified")
		}
	}
}

func createProfilePair(t *testing.T, base, purpose string) (string, string, string) {
	t.Helper()
	runtimeRoot := filepath.Join(base, "runtime")
	memoryRoot := filepath.Join(base, "memory")
	configured, err := profile.New("Friday", "Anthony", purpose, "balanced", "")
	if err != nil {
		t.Fatal(err)
	}
	p, err := bootstrap.Build(configured, runtimeRoot, memoryRoot)
	if err != nil {
		t.Fatal(err)
	}
	if err = repository.Create(p, runtimeRoot, memoryRoot); err != nil {
		t.Fatal(err)
	}
	return runtimeRoot, memoryRoot, p.AssistantID
}

func createNamedWithProfile(t *testing.T, home, name, executable, codex, source, purpose string) assistantinstance.Paths {
	t.Helper()
	runtimeRoot, memoryRoot, assistantID := createProfilePair(t, source, purpose)
	p, err := assistantinstance.PlanCreate(home, name, executable, codex)
	if err != nil {
		t.Fatal(err)
	}
	p, err = assistantinstance.WithRepositories(p, runtimeRoot, memoryRoot, assistantID)
	if err != nil {
		t.Fatal(err)
	}
	if err = assistantinstance.Create(p, executable, codex); err != nil {
		t.Fatal(err)
	}
	return p.Paths
}

func TestManagedInstructionsProjectValidatedPurposeAndRemainManifestBound(t *testing.T) {
	home := filepath.Join(t.TempDir(), "home")
	if err := os.MkdirAll(filepath.Join(home, ".local", "bin"), 0o755); err != nil {
		t.Fatal(err)
	}
	ambient := filepath.Join(home, ".codex", "ambient-canary")
	if err := os.MkdirAll(filepath.Dir(ambient), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ambient, []byte("ambient-unchanged"), 0o600); err != nil {
		t.Fatal(err)
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	codex := filepath.Join(home, "codex-stub")
	if err = os.WriteFile(codex, []byte("#!/bin/sh\nexit 0\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(home, "codex-code-mode-host"), []byte("host"), 0o700); err != nil {
		t.Fatal(err)
	}

	paths := map[string]assistantinstance.Paths{
		"alfred": createNamedWithProfile(t, home, "alfred", executable, codex, filepath.Join(home, "alfred-source"), "Return only PURPOSE_ALFRED"),
		"robin":  createNamedWithProfile(t, home, "robin", executable, codex, filepath.Join(home, "robin-source"), "Return only PURPOSE_ROBIN"),
	}
	manifests := make(map[string]assistantinstance.Manifest)
	for name, p := range paths {
		body, readErr := os.ReadFile(filepath.Join(p.Root, "codex", "AGENTS.md"))
		if readErr != nil {
			t.Fatal(readErr)
		}
		wantPurpose := "PURPOSE_" + strings.ToUpper(name)
		if !strings.Contains(string(body), wantPurpose) || strings.Contains(string(body), "assistant/profile.json") {
			t.Fatalf("%s instructions do not contain the projected purpose: %q", name, body)
		}
		m, verifyErr := assistantinstance.Verify(home, name)
		if verifyErr != nil {
			t.Fatal(verifyErr)
		}
		if m.CodexInstructions != filepath.Join(p.Root, "codex", "AGENTS.md") || m.CodexInstructionsSHA256 == "" {
			t.Fatalf("%s instructions are not manifest-bound: %#v", name, m)
		}
		manifests[name] = m
	}
	if manifests["alfred"].CodexInstructionsSHA256 == manifests["robin"].CodexInstructionsSHA256 {
		t.Fatal("distinct instance purposes produced the same instruction binding")
	}

	alfredInstructions := paths["alfred"].Root + "/codex/AGENTS.md"
	original, err := os.ReadFile(alfredInstructions)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(alfredInstructions, []byte("tampered\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err = assistantinstance.Verify(home, "alfred"); err == nil || !strings.Contains(err.Error(), "instructions") {
		t.Fatalf("tampered instructions verified: %v", err)
	}
	if _, err = assistantinstance.PlanRemove(home, "alfred"); err == nil {
		t.Fatal("tampered instructions granted removal authority")
	}
	if err = os.Remove(paths["alfred"].Launcher); err != nil {
		t.Fatal(err)
	}
	if _, err = assistantinstance.Recover(home, "alfred"); err == nil || !strings.Contains(err.Error(), "instructions") {
		t.Fatalf("tampered instructions granted recovery authority: %v", err)
	}
	if err = os.WriteFile(alfredInstructions, original, 0o600); err != nil {
		t.Fatal(err)
	}
	if result, recoverErr := assistantinstance.Recover(home, "alfred"); recoverErr != nil || result != "restored absent state" {
		t.Fatalf("manifest-proven recovery failed: %q %v", result, recoverErr)
	}
	removeRobin, err := assistantinstance.PlanRemove(home, "robin")
	if err != nil {
		t.Fatal(err)
	}
	if err = assistantinstance.Remove(removeRobin); err != nil {
		t.Fatal(err)
	}
	for _, p := range paths {
		if _, err = os.Lstat(p.Root); !os.IsNotExist(err) {
			t.Fatalf("instance root survived reversal: %s", p.Root)
		}
	}
	if body, readErr := os.ReadFile(ambient); readErr != nil || string(body) != "ambient-unchanged" {
		t.Fatalf("ambient Codex state changed: %q %v", body, readErr)
	}
}
