package codexhome

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	bootstrap "github.com/acoz-labs/my-friday/internal/plan"
	"github.com/acoz-labs/my-friday/internal/profile"
	"github.com/acoz-labs/my-friday/internal/repository"
)

func fixture(t *testing.T) (string, string) {
	t.Helper()
	root := t.TempDir()
	runtime := filepath.Join(root, "runtime")
	memory := filepath.Join(root, "memory")
	codex := filepath.Join(root, "home", ".codex")
	p, err := profile.New("Friday", "Anthony", "Help with careful work", "balanced", "")
	if err != nil {
		t.Fatal(err)
	}
	pl, err := bootstrap.Build(p, runtime, memory)
	if err != nil {
		t.Fatal(err)
	}
	if err = repository.Create(pl, runtime, memory); err != nil {
		t.Fatal(err)
	}
	return runtime, codex
}

func TestInstallVerifyAndCompleteReversal(t *testing.T) {
	runtime, codex := fixture(t)
	if err := os.MkdirAll(codex, 0700); err != nil {
		t.Fatal(err)
	}
	canary := filepath.Join(codex, "auth.json")
	if err := os.WriteFile(canary, []byte("unrelated"), 0600); err != nil {
		t.Fatal(err)
	}
	p, err := Plan(ActionInstall, runtime, codex)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Changes) == 0 || !strings.Contains(p.String(), "AGENTS.md") {
		t.Fatalf("missing preview: %#v", p)
	}
	if err = Execute(p); err != nil {
		t.Fatal(err)
	}
	s, err := Inspect(runtime, codex)
	if err != nil || s.State != StateHealthy {
		t.Fatalf("state=%s err=%v", s.State, err)
	}
	p, err = Plan(ActionUninstall, "", codex)
	if err != nil {
		t.Fatal(err)
	}
	if err = Execute(p); err != nil {
		t.Fatal(err)
	}
	if _, err = os.Stat(filepath.Join(codex, "AGENTS.md")); !os.IsNotExist(err) {
		t.Fatalf("projection remains: %v", err)
	}
	if got, _ := os.ReadFile(canary); string(got) != "unrelated" {
		t.Fatalf("canary changed: %q", got)
	}
}

func TestCollisionAndDriftFailClosedUntilRepair(t *testing.T) {
	runtime, codex := fixture(t)
	if err := os.MkdirAll(codex, 0700); err != nil {
		t.Fatal(err)
	}
	projection := filepath.Join(codex, "AGENTS.md")
	if err := os.WriteFile(projection, []byte("foreign"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Plan(ActionInstall, runtime, codex); err == nil {
		t.Fatal("foreign collision accepted")
	}
	if err := os.Remove(projection); err != nil {
		t.Fatal(err)
	}
	p, _ := Plan(ActionInstall, runtime, codex)
	if err := Execute(p); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(projection, []byte("drift"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Plan(ActionUninstall, "", codex); err == nil {
		t.Fatal("drifted uninstall accepted")
	}
	p, err := Plan(ActionRepair, runtime, codex)
	if err != nil {
		t.Fatal(err)
	}
	if err = Execute(p); err != nil {
		t.Fatal(err)
	}
	if s, _ := Inspect(runtime, codex); s.State != StateHealthy {
		t.Fatalf("state=%s", s.State)
	}
}

func TestForeignControlAndShadowingOverrideFailClosed(t *testing.T) {
	runtime, codex := fixture(t)
	if err := os.MkdirAll(filepath.Join(codex, controlDir), 0700); err != nil {
		t.Fatal(err)
	}
	if _, err := Plan(ActionInstall, runtime, codex); err == nil {
		t.Fatal("foreign control namespace accepted")
	}
	if err := os.RemoveAll(filepath.Join(codex, controlDir)); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(codex, "AGENTS.override.md"), []byte("foreign"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Plan(ActionInstall, runtime, codex); err == nil {
		t.Fatal("shadowing override accepted")
	}
}

func TestUnsafeProjectionTypesFailClosed(t *testing.T) {
	runtime, codex := fixture(t)
	if err := os.MkdirAll(codex, 0700); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(t.TempDir(), "target")
	if err := os.WriteFile(target, []byte("foreign"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(codex, "AGENTS.md")); err != nil {
		t.Fatal(err)
	}
	if _, err := Plan(ActionInstall, runtime, codex); err == nil {
		t.Fatal("symlink projection accepted")
	}
}

func TestUpgradeRollbackAndRecovery(t *testing.T) {
	runtime, codex := fixture(t)
	if err := os.MkdirAll(codex, 0700); err != nil {
		t.Fatal(err)
	}
	p, _ := Plan(ActionInstall, runtime, codex)
	if err := Execute(p); err != nil {
		t.Fatal(err)
	}
	pb, err := os.ReadFile(filepath.Join(runtime, "assistant", "profile.json"))
	if err != nil {
		t.Fatal(err)
	}
	var configured profile.Profile
	if json.Unmarshal(pb, &configured) != nil {
		t.Fatal("profile decode")
	}
	configured.Identity.Purpose = "New purpose"
	pb, _ = json.MarshalIndent(configured, "", "  ")
	pb = append(pb, '\n')
	if err := os.WriteFile(filepath.Join(runtime, "assistant", "profile.json"), pb, 0600); err != nil {
		t.Fatal(err)
	}
	p, err = Plan(ActionUpgrade, runtime, codex)
	if err != nil {
		t.Fatal(err)
	}
	if err = Execute(p); err != nil {
		t.Fatal(err)
	}
	p, err = Plan(ActionRollback, "", codex)
	if err != nil {
		t.Fatal(err)
	}
	if err = Execute(p); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(filepath.Join(codex, "AGENTS.md"))
	if !strings.Contains(string(got), "Help with careful work") || strings.Contains(string(got), "New purpose") {
		t.Fatalf("wrong rollback bytes: %s", got)
	}

	journal := filepath.Join(codex, controlDir, journalFile)
	if err := os.WriteFile(journal, []byte(`{"contract_version":1,"action":"repair","phase":"prepared"}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := Recover(codex); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(journal); !os.IsNotExist(err) {
		t.Fatalf("journal remains: %v", err)
	}
}
