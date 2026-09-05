package assistantinstance

import (
	"bytes"
	"encoding/json"
	"errors"
	bootstrap "github.com/acoz-labs/my-friday/internal/plan"
	"github.com/acoz-labs/my-friday/internal/profile"
	"github.com/acoz-labs/my-friday/internal/repository"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func fixture(t *testing.T) (string, string, string) {
	t.Helper()
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, ".local", "bin"), 0755); err != nil {
		t.Fatal(err)
	}
	exe := filepath.Join(home, "my-friday")
	codex := filepath.Join(home, "codex")
	for path, body := range map[string]string{exe: "launcher", codex: "codex"} {
		if err := os.WriteFile(path, []byte(body), 0700); err != nil {
			t.Fatal(err)
		}
	}
	return home, exe, codex
}

func downgradeFixtureToV1(t *testing.T, p Paths) {
	t.Helper()
	manifestPath := filepath.Join(p.Root, "manifest.json")
	body, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	var m Manifest
	if err = json.Unmarshal(body, &m); err != nil {
		t.Fatal(err)
	}
	m.ContractVersion = 1
	m.CapabilityRevision = 0
	m.RollbackContractVersion = 0
	m.RollbackCapabilityRevision = 0
	m.Owned = []string{"codex", "dependencies", "memory", "runtime", "workspace"}
	m.CapabilityBuilder = ""
	m.CapabilityBuilderSHA256 = ""
	m.CapabilityPolicySHA256 = ""
	m.MyFridayExecutable = ""
	m.MyFridaySHA256 = ""
	config, err := managedCodexConfig(p, 0)
	if err != nil {
		t.Fatal(err)
	}
	m.CodexConfigSHA256 = digest(config)
	body, _ = json.MarshalIndent(m, "", "  ")
	if err = os.WriteFile(manifestPath, append(body, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(m.CodexConfig, config, 0o600); err != nil {
		t.Fatal(err)
	}
	if err = os.Remove(filepath.Join(p.Root, "dependencies", "my-friday")); err != nil {
		t.Fatal(err)
	}
}

func downgradeFixtureToLegacyV2(t *testing.T, p Paths) {
	t.Helper()
	manifestPath := filepath.Join(p.Root, "manifest.json")
	body, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	var m Manifest
	if err = json.Unmarshal(body, &m); err != nil {
		t.Fatal(err)
	}
	m.CapabilityRevision = 0
	m.RollbackContractVersion = 0
	m.RollbackCapabilityRevision = 0
	m.MyFridayExecutable = ""
	m.MyFridaySHA256 = ""
	m.CapabilityBuilderSHA256 = digest([]byte(legacyBuilderSkill))
	m.CapabilityPolicySHA256 = digest([]byte(builderPolicy))
	config, err := managedCodexConfig(p, 0)
	if err != nil {
		t.Fatal(err)
	}
	m.CodexConfigSHA256 = digest(config)
	body, _ = json.MarshalIndent(m, "", "  ")
	if err = os.WriteFile(manifestPath, append(body, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(m.CodexConfig, config, 0o600); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(m.CapabilityBuilder, "SKILL.md"), []byte(legacyBuilderSkill), 0o600); err != nil {
		t.Fatal(err)
	}
	if err = os.Remove(filepath.Join(p.Root, "dependencies", "my-friday")); err != nil {
		t.Fatal(err)
	}
}

func downgradeFixtureToRevision2(t *testing.T, p Paths) {
	t.Helper()
	manifestPath := filepath.Join(p.Root, "manifest.json")
	body, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	var m Manifest
	if err = json.Unmarshal(body, &m); err != nil {
		t.Fatal(err)
	}
	m.CapabilityRevision = 2
	m.CapabilityBuilderSHA256 = digest(capabilityRevision2Builder(p))
	config, err := managedCodexConfig(p, 2)
	if err != nil {
		t.Fatal(err)
	}
	m.CodexConfigSHA256 = digest(config)
	body, _ = json.MarshalIndent(m, "", "  ")
	if err = os.WriteFile(manifestPath, append(body, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(m.CodexConfig, config, 0o600); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(m.CapabilityBuilder, "SKILL.md"), capabilityRevision2Builder(p), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestActualRevision2ManifestVerifiesPlansUpgradeAndRefusesDrift(t *testing.T) {
	home, exe, codex := fixture(t)
	p, err := PlanCreate(home, "alfred", exe, codex)
	if err != nil {
		t.Fatal(err)
	}
	if err = Create(p, exe, codex); err != nil {
		t.Fatal(err)
	}
	downgradeFixtureToRevision2(t, p.Paths)
	if _, err = Verify(home, "alfred"); err != nil {
		t.Fatalf("revision 2 compatibility: %v", err)
	}
	managed := filepath.Join(p.Paths.Root, "dependencies", "my-friday")
	original, err := os.ReadFile(managed)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(managed, []byte("revision 2 executable drift"), 0700); err != nil {
		t.Fatal(err)
	}
	if _, err = Verify(home, "alfred"); err == nil || !strings.Contains(err.Error(), "My Friday executable drift") {
		t.Fatalf("revision 2 executable drift accepted: %v", err)
	}
	if err = os.WriteFile(managed, original, 0700); err != nil {
		t.Fatal(err)
	}
	candidate := candidateFixture(t, home, "revision 3 candidate")
	if _, err = PlanUpgrade(home, "alfred", candidate); err != nil {
		t.Fatalf("revision 2 plan: %v", err)
	}
	if err = os.Chmod(filepath.Join(p.Paths.Root, "workspace", ".agents", "skills", "capability-builder", "SKILL.md"), 0o640); err != nil {
		t.Fatal(err)
	}
	if _, err = Verify(home, "alfred"); err == nil || !strings.Contains(err.Error(), "builder drift") {
		t.Fatalf("revision 2 drift accepted: %v", err)
	}
}

func TestActualRevision2UpgradeRecoveryAndRollback(t *testing.T) {
	home, exe, codex := fixture(t)
	p, err := PlanCreate(home, "alfred", exe, codex)
	if err != nil {
		t.Fatal(err)
	}
	if err = Create(p, exe, codex); err != nil {
		t.Fatal(err)
	}
	downgradeFixtureToRevision2(t, p.Paths)
	candidate := candidateFixture(t, home, "revision 3 interrupted candidate")
	upgrade, err := PlanUpgrade(home, "alfred", candidate)
	if err != nil {
		t.Fatal(err)
	}
	upgradeHook = func(phase string) error {
		if phase == "builder-created" {
			return errors.New("stop")
		}
		return nil
	}
	if err = Upgrade(upgrade); err == nil {
		t.Fatal("interruption missing")
	}
	upgradeHook = nil
	defer func() { upgradeHook = nil }()
	if _, err = Recover(home, "alfred"); err != nil {
		t.Fatal(err)
	}
	m, err := Verify(home, "alfred")
	if err != nil || m.CapabilityRevision != CapabilityRevision || m.RollbackCapabilityRevision != 2 {
		t.Fatalf("recovered=%#v err=%v", m, err)
	}
	rollback, err := PlanRollback(home, "alfred")
	if err != nil {
		t.Fatal(err)
	}
	if err = Rollback(rollback); err != nil {
		t.Fatal(err)
	}
	m, err = Verify(home, "alfred")
	if err != nil || m.CapabilityRevision != 2 {
		t.Fatalf("rollback=%#v err=%v", m, err)
	}
}

func TestActualRevision2RollbackResumesExecutableAndManifestPhases(t *testing.T) {
	for _, phase := range []string{"executable-restored", "manifest-promoted"} {
		t.Run(phase, func(t *testing.T) {
			home, exe, codex := fixture(t)
			p, err := PlanCreate(home, "alfred", exe, codex)
			if err != nil {
				t.Fatal(err)
			}
			if err = Create(p, exe, codex); err != nil {
				t.Fatal(err)
			}
			downgradeFixtureToRevision2(t, p.Paths)
			rollbackBytes, err := os.ReadFile(filepath.Join(p.Paths.Root, "dependencies", "my-friday"))
			if err != nil {
				t.Fatal(err)
			}
			candidate := candidateFixture(t, home, "revision 3 candidate for rollback interruption")
			upgrade, err := PlanUpgrade(home, "alfred", candidate)
			if err != nil {
				t.Fatal(err)
			}
			if err = Upgrade(upgrade); err != nil {
				t.Fatal(err)
			}
			plan, err := PlanRollback(home, "alfred")
			if err != nil {
				t.Fatal(err)
			}
			rollbackHook = func(got string) error {
				if got == phase {
					return errors.New("stop")
				}
				return nil
			}
			if err = Rollback(plan); err == nil {
				t.Fatal("rollback fault missing")
			}
			rollbackHook = nil
			t.Cleanup(func() { rollbackHook = nil })
			if _, err = Recover(home, "alfred"); err != nil {
				t.Fatal(err)
			}
			m, err := Verify(home, "alfred")
			if err != nil || m.CapabilityRevision != 2 {
				t.Fatalf("manifest=%#v err=%v", m, err)
			}
			got, err := os.ReadFile(m.MyFridayExecutable)
			if err != nil || !bytes.Equal(got, rollbackBytes) {
				t.Fatalf("restored executable mismatch err=%v", err)
			}
		})
	}
}

func candidateFixture(t *testing.T, home, body string) string {
	t.Helper()
	path := filepath.Join(home, "current-my-friday")
	if err := os.WriteFile(path, []byte(body), 0o700); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestFindCodexResolvesPATHSymlink(t *testing.T) {
	home, _, codex := fixture(t)
	bin := filepath.Join(home, "path-bin")
	if err := os.Mkdir(bin, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(codex, filepath.Join(bin, "codex")); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin)
	got, err := FindCodex()
	if err != nil {
		t.Fatal(err)
	}
	want, err := filepath.EvalSymlinks(codex)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("got %q want resolved %q", got, want)
	}
}

func TestNameAndPaths(t *testing.T) {
	for _, valid := range []string{"alfred", "friday-2", "a"} {
		if err := ValidateName(valid); err != nil {
			t.Errorf("%s: %v", valid, err)
		}
	}
	for _, invalid := range []string{"", "Alfred", "álfred", "-alfred", "alfred-", "a/b", "..", "codex", strings.Repeat("a", 33)} {
		if ValidateName(invalid) == nil {
			t.Errorf("accepted %q", invalid)
		}
	}
	p, err := Derive("/Users/test", "alfred")
	if err != nil {
		t.Fatal(err)
	}
	if p.Root != "/Users/test/.my-friday/assistants/alfred" || p.Launcher != "/Users/test/.local/bin/alfred" {
		t.Fatalf("wrong paths: %#v", p)
	}
}

func TestCreateVerifyRemovePreservesSiblings(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix permission contract")
	}
	home, exe, codex := fixture(t)
	sibling := filepath.Join(home, ".local", "bin", "keep")
	if err := os.WriteFile(sibling, []byte("keep"), 0700); err != nil {
		t.Fatal(err)
	}
	p, err := PlanCreate(home, "alfred", exe, codex)
	if err != nil {
		t.Fatal(err)
	}
	if err = Create(p, exe, codex); err != nil {
		t.Fatal(err)
	}
	m, err := Verify(home, "alfred")
	if err != nil {
		t.Fatal(err)
	}
	if m.Root != p.Paths.Root || m.Launcher != p.Paths.Launcher {
		t.Fatal("manifest paths changed")
	}
	if _, err = PlanCreate(home, "alfred", exe, codex); err == nil {
		t.Fatal("repeat create did not refuse collision")
	}
	rp, err := PlanRemove(home, "alfred")
	if err != nil {
		t.Fatal(err)
	}
	if err = Remove(rp); err != nil {
		t.Fatal(err)
	}
	if b, err := os.ReadFile(sibling); err != nil || string(b) != "keep" {
		t.Fatal("launcher sibling changed")
	}
	if _, err = os.Stat(p.Paths.Root); !os.IsNotExist(err) {
		t.Fatal("root remains")
	}
	entries, err := os.ReadDir(filepath.Join(home, ".my-friday", "assistants"))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.Contains(entry.Name(), "alfred") {
			t.Fatalf("per-name artifact remains after removal: %s", entry.Name())
		}
	}
}

func TestUpgradeV1InstanceProjectsManifestBoundBuilder(t *testing.T) {
	home, exe, codex := fixture(t)
	p, err := PlanCreate(home, "alfred", exe, codex)
	if err != nil {
		t.Fatal(err)
	}
	if err = Create(p, exe, codex); err != nil {
		t.Fatal(err)
	}
	downgradeFixtureToV1(t, p.Paths)
	if err = os.RemoveAll(filepath.Join(p.Paths.Root, "capabilities")); err != nil {
		t.Fatal(err)
	}
	if err = os.RemoveAll(filepath.Join(p.Paths.Root, "workspace", ".agents")); err != nil {
		t.Fatal(err)
	}
	if _, err = Verify(home, "alfred"); err != nil {
		t.Fatalf("v1 compatibility: %v", err)
	}
	candidate := candidateFixture(t, home, "new candidate bytes")
	upgrade, err := PlanUpgrade(home, "alfred", candidate)
	if err != nil {
		t.Fatal(err)
	}
	if err = Upgrade(upgrade); err != nil {
		t.Fatal(err)
	}
	upgraded, err := Verify(home, "alfred")
	if err != nil {
		t.Fatal(err)
	}
	if upgraded.ContractVersion != 2 || upgraded.CapabilityBuilderSHA256 == "" {
		t.Fatalf("not upgraded: %#v", upgraded)
	}
	if upgraded.MyFridaySHA256 != digest([]byte("new candidate bytes")) {
		t.Fatalf("upgrade retained old launcher authority: %#v", upgraded)
	}
	managed, err := os.ReadFile(upgraded.MyFridayExecutable)
	if err != nil || string(managed) != "new candidate bytes" {
		t.Fatalf("managed candidate mismatch: %q %v", managed, err)
	}
	oldLauncher, err := os.ReadFile(exe)
	if err != nil || string(oldLauncher) != "launcher" {
		t.Fatalf("old candidate changed: %q %v", oldLauncher, err)
	}
	if err = os.WriteFile(filepath.Join(upgraded.CapabilityBuilder, "SKILL.md"), []byte("drift"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err = Verify(home, "alfred"); err == nil || !strings.Contains(err.Error(), "builder drift") {
		t.Fatalf("builder drift accepted: %v", err)
	}
}

func TestLegacyV2UpgradeAndRollbackRemainSupported(t *testing.T) {
	home, exe, codex := fixture(t)
	p, err := PlanCreate(home, "alfred", exe, codex)
	if err != nil {
		t.Fatal(err)
	}
	if err = Create(p, exe, codex); err != nil {
		t.Fatal(err)
	}
	downgradeFixtureToLegacyV2(t, p.Paths)
	if _, err = Verify(home, "alfred"); err != nil {
		t.Fatalf("legacy v2 compatibility: %v", err)
	}
	candidate := candidateFixture(t, home, "replacement candidate")
	upgrade, err := PlanUpgrade(home, "alfred", candidate)
	if err != nil {
		t.Fatal(err)
	}
	if err = Upgrade(upgrade); err != nil {
		t.Fatal(err)
	}
	m, err := Verify(home, "alfred")
	if err != nil || m.CapabilityRevision != CapabilityRevision || m.RollbackContractVersion != ContractVersion || m.RollbackCapabilityRevision != 0 || m.MyFridaySHA256 != digest([]byte("replacement candidate")) {
		t.Fatalf("legacy v2 upgrade mismatch: %#v %v", m, err)
	}
	rollback, err := PlanRollback(home, "alfred")
	if err != nil {
		t.Fatal(err)
	}
	if err = Rollback(rollback); err != nil {
		t.Fatal(err)
	}
	m, err = Verify(home, "alfred")
	if err != nil || m.ContractVersion != ContractVersion || m.CapabilityRevision != 0 || m.MyFridayExecutable != "" {
		t.Fatalf("legacy v2 rollback mismatch: %#v %v", m, err)
	}
}

func TestLegacyV2InterruptedUpgradeRecoversAndRollsBack(t *testing.T) {
	home, exe, codex := fixture(t)
	p, err := PlanCreate(home, "alfred", exe, codex)
	if err != nil {
		t.Fatal(err)
	}
	if err = Create(p, exe, codex); err != nil {
		t.Fatal(err)
	}
	downgradeFixtureToLegacyV2(t, p.Paths)
	candidate := candidateFixture(t, home, "interrupted candidate")
	upgrade, err := PlanUpgrade(home, "alfred", candidate)
	if err != nil {
		t.Fatal(err)
	}
	upgradeHook = func(phase string) error {
		if phase == "execution-context-created" {
			return errors.New("stop")
		}
		return nil
	}
	if err = Upgrade(upgrade); err == nil {
		t.Fatal("interruption missing")
	}
	upgradeHook = nil
	defer func() { upgradeHook = nil }()
	if _, err = Recover(home, "alfred"); err != nil {
		t.Fatal(err)
	}
	rollback, err := PlanRollback(home, "alfred")
	if err != nil {
		t.Fatal(err)
	}
	if err = Rollback(rollback); err != nil {
		t.Fatal(err)
	}
	m, err := Verify(home, "alfred")
	if err != nil || m.ContractVersion != ContractVersion || m.CapabilityRevision != 0 {
		t.Fatalf("legacy recovery rollback mismatch: %#v %v", m, err)
	}
}

func TestCapabilityEmptyInstanceCanRollbackToV1(t *testing.T) {
	home, exe, codex := fixture(t)
	p, err := PlanCreate(home, "alfred", exe, codex)
	if err != nil {
		t.Fatal(err)
	}
	if err = Create(p, exe, codex); err != nil {
		t.Fatal(err)
	}
	rollback, err := PlanRollback(home, "alfred")
	if err != nil {
		t.Fatal(err)
	}
	if err = Rollback(rollback); err != nil {
		t.Fatal(err)
	}
	m, err := Verify(home, "alfred")
	if err != nil {
		t.Fatal(err)
	}
	if m.ContractVersion != 1 || m.CapabilityBuilder != "" {
		t.Fatalf("rollback failed: %#v", m)
	}
	if m.MyFridayExecutable != "" || m.MyFridaySHA256 != "" {
		t.Fatalf("rollback retained managed My Friday authority: %#v", m)
	}
	if _, err = os.Stat(filepath.Join(p.Paths.Root, "dependencies", "my-friday")); !os.IsNotExist(err) {
		t.Fatalf("rollback retained managed My Friday executable: %v", err)
	}
	wantConfig, err := managedCodexConfig(p.Paths, 0)
	if err != nil {
		t.Fatal(err)
	}
	gotConfig, err := os.ReadFile(filepath.Join(p.Paths.Root, "codex", "config.toml"))
	if err != nil || !bytes.Equal(gotConfig, wantConfig) {
		t.Fatalf("rollback did not restore v1 config: %v", err)
	}
}

func TestUpgradeFailurePreservesPreexistingWorkspaceAgents(t *testing.T) {
	home, exe, codex := fixture(t)
	p, err := PlanCreate(home, "alfred", exe, codex)
	if err != nil {
		t.Fatal(err)
	}
	if err = Create(p, exe, codex); err != nil {
		t.Fatal(err)
	}
	downgradeFixtureToV1(t, p.Paths)
	if err = os.RemoveAll(filepath.Join(p.Paths.Root, "capabilities")); err != nil {
		t.Fatal(err)
	}
	if err = os.RemoveAll(filepath.Join(p.Paths.Root, "workspace", ".agents")); err != nil {
		t.Fatal(err)
	}
	keep := filepath.Join(p.Paths.Root, "workspace", ".agents", "keep")
	if err = os.MkdirAll(keep, 0o700); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(keep, "foreign"), []byte("keep"), 0o600); err != nil {
		t.Fatal(err)
	}
	pl, err := PlanUpgrade(home, "alfred", exe)
	if err != nil {
		t.Fatal(err)
	}
	upgradeHook = func(string) error { return errors.New("injected upgrade failure") }
	defer func() { upgradeHook = nil }()
	if err = Upgrade(pl); err == nil {
		t.Fatal("upgrade fault missing")
	}
	if b, err := os.ReadFile(filepath.Join(keep, "foreign")); err != nil || string(b) != "keep" {
		t.Fatal("foreign workspace state changed")
	}
	upgradeHook = nil
	result, err := Recover(home, "alfred")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "completed capability upgrade") {
		t.Fatalf("result=%q", result)
	}
	m, err := Verify(home, "alfred")
	if err != nil || m.ContractVersion != 2 {
		t.Fatalf("upgrade recovery failed: %#v %v", m, err)
	}
	if b, err := os.ReadFile(filepath.Join(keep, "foreign")); err != nil || string(b) != "keep" {
		t.Fatal("foreign state changed during recovery")
	}
}

func TestRecoverCompletesUpgradeAfterExecutionContextMutation(t *testing.T) {
	home, exe, codex := fixture(t)
	p, err := PlanCreate(home, "alfred", exe, codex)
	if err != nil {
		t.Fatal(err)
	}
	if err = Create(p, exe, codex); err != nil {
		t.Fatal(err)
	}
	downgradeFixtureToV1(t, p.Paths)
	if err = os.RemoveAll(filepath.Join(p.Paths.Root, "capabilities")); err != nil {
		t.Fatal(err)
	}
	if err = os.RemoveAll(filepath.Join(p.Paths.Root, "workspace", ".agents")); err != nil {
		t.Fatal(err)
	}
	pl, err := PlanUpgrade(home, "alfred", exe)
	if err != nil {
		t.Fatal(err)
	}
	upgradeHook = func(phase string) error {
		if phase == "execution-context-created" {
			return errors.New("injected post-context interruption")
		}
		return nil
	}
	if err = Upgrade(pl); err == nil {
		t.Fatal("post-context fault missing")
	}
	upgradeHook = nil
	defer func() { upgradeHook = nil }()
	if _, err = os.Stat(filepath.Join(p.Paths.Root, "dependencies", "my-friday")); err != nil {
		t.Fatal("managed executable mutation was not reached")
	}
	result, err := Recover(home, "alfred")
	if err != nil {
		t.Fatal(err)
	}
	if result != "completed capability upgrade" {
		t.Fatalf("result=%q", result)
	}
	if _, err = Verify(home, "alfred"); err != nil {
		t.Fatal(err)
	}
}

func TestUpgradeRefusesManagedMyFridayCollisionWithoutChangingBytes(t *testing.T) {
	home, exe, codex := fixture(t)
	p, err := PlanCreate(home, "alfred", exe, codex)
	if err != nil {
		t.Fatal(err)
	}
	if err = Create(p, exe, codex); err != nil {
		t.Fatal(err)
	}
	downgradeFixtureToV1(t, p.Paths)
	if err = os.RemoveAll(filepath.Join(p.Paths.Root, "capabilities")); err != nil {
		t.Fatal(err)
	}
	if err = os.RemoveAll(filepath.Join(p.Paths.Root, "workspace", ".agents")); err != nil {
		t.Fatal(err)
	}
	pl, err := PlanUpgrade(home, "alfred", exe)
	if err != nil {
		t.Fatal(err)
	}
	collision := filepath.Join(p.Paths.Root, "dependencies", "my-friday")
	if err = os.WriteFile(collision, []byte("foreign"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err = Upgrade(pl); err == nil || !strings.Contains(err.Error(), "collision") {
		t.Fatalf("managed executable collision accepted: %v", err)
	}
	if body, readErr := os.ReadFile(collision); readErr != nil || string(body) != "foreign" {
		t.Fatalf("managed executable collision changed: %q %v", body, readErr)
	}
}

func TestRollbackRechecksWorkspaceSkillsUnderLock(t *testing.T) {
	home, exe, codex := fixture(t)
	p, err := PlanCreate(home, "alfred", exe, codex)
	if err != nil {
		t.Fatal(err)
	}
	if err = Create(p, exe, codex); err != nil {
		t.Fatal(err)
	}
	pl, err := PlanRollback(home, "alfred")
	if err != nil {
		t.Fatal(err)
	}
	sibling := filepath.Join(p.Paths.Root, "workspace", ".agents", "skills", "foreign")
	if err = os.Mkdir(sibling, 0o700); err != nil {
		t.Fatal(err)
	}
	if err = Rollback(pl); err == nil {
		t.Fatal("rollback ignored new sibling")
	}
	if _, err = os.Stat(filepath.Join(p.Paths.Root, "workspace", ".agents", "skills", "capability-builder", "SKILL.md")); err != nil {
		t.Fatal("builder damaged")
	}
	if _, err = os.Stat(sibling); err != nil {
		t.Fatal("foreign sibling removed")
	}
}

func TestRollbackFaultRecoversFromQuarantinedOwnedState(t *testing.T) {
	home, exe, codex := fixture(t)
	p, err := PlanCreate(home, "alfred", exe, codex)
	if err != nil {
		t.Fatal(err)
	}
	if err = Create(p, exe, codex); err != nil {
		t.Fatal(err)
	}
	pl, err := PlanRollback(home, "alfred")
	if err != nil {
		t.Fatal(err)
	}
	rollbackHook = func(string) error { return errors.New("injected rollback stop") }
	if err = Rollback(pl); err == nil {
		t.Fatal("fault missing")
	}
	rollbackHook = nil
	result, err := Recover(home, "alfred")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "completed capability rollback") {
		t.Fatalf("result=%q", result)
	}
	m, err := Verify(home, "alfred")
	if err != nil || m.ContractVersion != 1 {
		t.Fatalf("rollback recovery failed: %#v %v", m, err)
	}
}

func TestRollbackRefusesPreexistingQuarantineCollisions(t *testing.T) {
	for _, tc := range []struct {
		name string
		path func(Paths) string
	}{
		{"builder", func(p Paths) string {
			return filepath.Join(p.Root, "workspace", ".agents", "skills", "capability-builder.rollback")
		}},
		{"capabilities", func(p Paths) string { return filepath.Join(p.Root, "capabilities.rollback") }},
		{"executable", func(p Paths) string { return filepath.Join(p.Root, "dependencies", "my-friday.rollback") }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			home, exe, codex := fixture(t)
			p, err := PlanCreate(home, "alfred", exe, codex)
			if err != nil {
				t.Fatal(err)
			}
			if err = Create(p, exe, codex); err != nil {
				t.Fatal(err)
			}
			rollback, err := PlanRollback(home, "alfred")
			if err != nil {
				t.Fatal(err)
			}
			collision := tc.path(p.Paths)
			if err = os.WriteFile(collision, []byte("foreign"), 0o600); err != nil {
				t.Fatal(err)
			}
			if err = Rollback(rollback); err == nil || !strings.Contains(err.Error(), "quarantine collision") {
				t.Fatalf("collision accepted: %v", err)
			}
			body, readErr := os.ReadFile(collision)
			if readErr != nil || string(body) != "foreign" {
				t.Fatalf("foreign collision changed: %q %v", body, readErr)
			}
			if _, statErr := os.Lstat(capabilityMigrationPath(p.Paths.Root)); !errors.Is(statErr, os.ErrNotExist) {
				t.Fatalf("collision created migration journal: %v", statErr)
			}
		})
	}
}

func TestRollbackRecoveryRefusesQuarantineByteSubstitution(t *testing.T) {
	for _, tc := range []struct {
		name string
		path func(Paths) string
	}{
		{"builder", func(p Paths) string {
			return filepath.Join(p.Root, "workspace", ".agents", "skills", "capability-builder.rollback", "SKILL.md")
		}},
		{"capabilities", func(p Paths) string { return filepath.Join(p.Root, "capabilities.rollback", "foreign") }},
		{"executable", func(p Paths) string { return filepath.Join(p.Root, "dependencies", "my-friday.rollback") }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			home, exe, codex := fixture(t)
			p, err := PlanCreate(home, "alfred", exe, codex)
			if err != nil {
				t.Fatal(err)
			}
			if err = Create(p, exe, codex); err != nil {
				t.Fatal(err)
			}
			rollback, err := PlanRollback(home, "alfred")
			if err != nil {
				t.Fatal(err)
			}
			rollbackHook = func(phase string) error {
				if phase == "quarantined" {
					return errors.New("stop")
				}
				return nil
			}
			if err = Rollback(rollback); err == nil {
				t.Fatal("interruption missing")
			}
			rollbackHook = nil
			defer func() { rollbackHook = nil }()
			substitute := tc.path(p.Paths)
			if err = os.WriteFile(substitute, []byte("substituted"), 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err = Recover(home, "alfred"); err == nil || !strings.Contains(err.Error(), "quarantine drift") {
				t.Fatalf("substitution accepted: %v", err)
			}
			body, readErr := os.ReadFile(substitute)
			if readErr != nil || string(body) != "substituted" {
				t.Fatalf("substituted bytes changed: %q %v", body, readErr)
			}
		})
	}
}

func TestUpgradeRecoveryRefusesCandidateStageSubstitution(t *testing.T) {
	home, exe, codex := fixture(t)
	p, err := PlanCreate(home, "alfred", exe, codex)
	if err != nil {
		t.Fatal(err)
	}
	if err = Create(p, exe, codex); err != nil {
		t.Fatal(err)
	}
	downgradeFixtureToV1(t, p.Paths)
	if err = os.RemoveAll(filepath.Join(p.Paths.Root, "capabilities")); err != nil {
		t.Fatal(err)
	}
	if err = os.RemoveAll(filepath.Join(p.Paths.Root, "workspace", ".agents")); err != nil {
		t.Fatal(err)
	}
	candidate := candidateFixture(t, home, "candidate")
	upgrade, err := PlanUpgrade(home, "alfred", candidate)
	if err != nil {
		t.Fatal(err)
	}
	upgradeHook = func(phase string) error {
		if phase == "execution-context-created" {
			return errors.New("stop")
		}
		return nil
	}
	if err = Upgrade(upgrade); err == nil {
		t.Fatal("interruption missing")
	}
	upgradeHook = nil
	defer func() { upgradeHook = nil }()
	stage := filepath.Join(p.Paths.Root, "dependencies", "my-friday.upgrade")
	if err = os.WriteFile(stage, []byte("substituted"), 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err = Recover(home, "alfred"); err == nil || !strings.Contains(err.Error(), "candidate staging drift") {
		t.Fatalf("candidate substitution accepted: %v", err)
	}
	body, readErr := os.ReadFile(stage)
	if readErr != nil || string(body) != "substituted" {
		t.Fatalf("substituted candidate changed: %q %v", body, readErr)
	}
}

func TestUpgradeInitializesOnlyPrivateCopiedRuntime(t *testing.T) {
	home, exe, codex := fixture(t)
	runtimeRoot := filepath.Join(home, "source-runtime")
	memoryRoot := filepath.Join(home, "source-memory")
	configured, _ := profile.New("Friday", "", "Help", "balanced", "")
	rp, _ := bootstrap.Build(configured, runtimeRoot, memoryRoot)
	if err := repository.Create(rp, runtimeRoot, memoryRoot); err != nil {
		t.Fatal(err)
	}
	p, err := PlanCreate(home, "alfred", exe, codex)
	if err != nil {
		t.Fatal(err)
	}
	p, err = WithRepositories(p, runtimeRoot, memoryRoot, rp.AssistantID)
	if err != nil {
		t.Fatal(err)
	}
	if err = Create(p, exe, codex); err != nil {
		t.Fatal(err)
	}
	privateRuntime := filepath.Join(p.Paths.Root, "runtime")
	if err = repository.RollbackCapabilities(privateRuntime); err != nil {
		t.Fatal(err)
	}
	downgradeFixtureToV1(t, p.Paths)
	if err = os.RemoveAll(filepath.Join(p.Paths.Root, "capabilities")); err != nil {
		t.Fatal(err)
	}
	if err = os.RemoveAll(filepath.Join(p.Paths.Root, "workspace", ".agents")); err != nil {
		t.Fatal(err)
	}
	pl, err := PlanUpgrade(home, "alfred", exe)
	if err != nil {
		t.Fatal(err)
	}
	if err = Upgrade(pl); err != nil {
		t.Fatal(err)
	}
	if _, err = os.Stat(filepath.Join(privateRuntime, ".my-friday", "capability-contract.json")); err != nil {
		t.Fatal("private runtime not initialized")
	}
	if _, err = os.Stat(filepath.Join(runtimeRoot, ".my-friday", "capability-contract.json")); err != nil {
		t.Fatal("external source unexpectedly changed")
	}
}

func TestForeignLauncherAndDriftFailClosed(t *testing.T) {
	home, exe, codex := fixture(t)
	foreign := filepath.Join(home, ".local", "bin", "alfred")
	if err := os.WriteFile(foreign, []byte("foreign"), 0700); err != nil {
		t.Fatal(err)
	}
	if _, err := PlanCreate(home, "alfred", exe, codex); err == nil {
		t.Fatal("foreign launcher accepted")
	}
	if _, err := os.Stat(filepath.Join(home, ".my-friday")); !os.IsNotExist(err) {
		t.Fatal("collision mutated root")
	}
	if err := os.Remove(foreign); err != nil {
		t.Fatal(err)
	}
	p, err := PlanCreate(home, "alfred", exe, codex)
	if err != nil {
		t.Fatal(err)
	}
	if err = Create(p, exe, codex); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(p.Paths.Launcher, []byte("drift"), 0700); err != nil {
		t.Fatal(err)
	}
	if _, err = PlanRemove(home, "alfred"); err == nil {
		t.Fatal("drifted launcher granted deletion authority")
	}
	if _, err = os.Stat(p.Paths.Root); err != nil {
		t.Fatal("drift removed root")
	}
}

func TestPostPreviewForeignCollisionLeavesNoInstanceControlArtifact(t *testing.T) {
	home, exe, codex := fixture(t)
	p, err := PlanCreate(home, "alfred", exe, codex)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(p.Paths.Launcher, []byte("foreign"), 0700); err != nil {
		t.Fatal(err)
	}
	if err = Create(p, exe, codex); err == nil {
		t.Fatal("post-preview collision accepted")
	}
	if _, err = os.Stat(filepath.Join(home, ".my-friday")); !os.IsNotExist(err) {
		t.Fatalf("collision created control state: %v", err)
	}
}

func TestRemovePathReplacementPreservesForeignRootExactly(t *testing.T) {
	home, exe, codex := fixture(t)
	p, err := PlanCreate(home, "alfred", exe, codex)
	if err != nil {
		t.Fatal(err)
	}
	if err = Create(p, exe, codex); err != nil {
		t.Fatal(err)
	}
	removePlan, err := PlanRemove(home, "alfred")
	if err != nil {
		t.Fatal(err)
	}
	original := p.Paths.Root + ".original"
	if err = os.Rename(p.Paths.Root, original); err != nil {
		t.Fatal(err)
	}
	if err = os.Mkdir(p.Paths.Root, 0700); err != nil {
		t.Fatal(err)
	}
	canary := filepath.Join(p.Paths.Root, "foreign-canary")
	if err = os.WriteFile(canary, []byte("unchanged"), 0600); err != nil {
		t.Fatal(err)
	}
	if err = Remove(removePlan); err == nil {
		t.Fatal("replaced root accepted")
	}
	entries, err := os.ReadDir(p.Paths.Root)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "foreign-canary" {
		t.Fatalf("foreign entries changed: %v", entries)
	}
	if b, err := os.ReadFile(canary); err != nil || string(b) != "unchanged" {
		t.Fatalf("foreign content changed: %q %v", b, err)
	}
}

func TestRecoverForeignRootPreservesEntriesAndContent(t *testing.T) {
	home, _, _ := fixture(t)
	p, err := Derive(home, "alfred")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.MkdirAll(p.Root, 0700); err != nil {
		t.Fatal(err)
	}
	canary := filepath.Join(p.Root, "foreign-canary")
	if err = os.WriteFile(canary, []byte("unchanged"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = Recover(home, "alfred"); err == nil {
		t.Fatal("foreign recovery root accepted")
	}
	entries, err := os.ReadDir(p.Root)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "foreign-canary" {
		t.Fatalf("foreign entries changed: %v", entries)
	}
	if b, err := os.ReadFile(canary); err != nil || string(b) != "unchanged" {
		t.Fatalf("foreign content changed: %q %v", b, err)
	}
}

func TestForgedManagedExecutableManifestIsDenied(t *testing.T) {
	for _, mutate := range []func(*Manifest){
		func(m *Manifest) { m.CodexExecutable = "/bin/echo" },
		func(m *Manifest) { m.MyFridayExecutable = "/bin/echo" },
		func(m *Manifest) { m.MyFridaySHA256 = strings.Repeat("a", 64) },
	} {
		home, exe, codex := fixture(t)
		p, err := PlanCreate(home, "alfred", exe, codex)
		if err != nil {
			t.Fatal(err)
		}
		if err = Create(p, exe, codex); err != nil {
			t.Fatal(err)
		}
		manifestPath := filepath.Join(p.Paths.Root, "manifest.json")
		b, err := os.ReadFile(manifestPath)
		if err != nil {
			t.Fatal(err)
		}
		var m Manifest
		if err = json.Unmarshal(b, &m); err != nil {
			t.Fatal(err)
		}
		mutate(&m)
		b, err = json.MarshalIndent(m, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(manifestPath, append(b, '\n'), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err = Verify(home, "alfred"); err == nil {
			t.Fatalf("forged executable accepted: %v", err)
		}
	}
}

func TestTwoInstancesAreIndependent(t *testing.T) {
	home, exe, codex := fixture(t)
	for _, name := range []string{"alfred", "robin"} {
		p, err := PlanCreate(home, name, exe, codex)
		if err != nil {
			t.Fatal(err)
		}
		if err = Create(p, exe, codex); err != nil {
			t.Fatal(err)
		}
	}
	rp, err := PlanRemove(home, "alfred")
	if err != nil {
		t.Fatal(err)
	}
	if err = Remove(rp); err != nil {
		t.Fatal(err)
	}
	if _, err = Verify(home, "robin"); err != nil {
		t.Fatal("sibling instance changed:", err)
	}
}

func TestCreateTrustsOnlyExactInstanceWorkspaceWithTOMLEscaping(t *testing.T) {
	base := t.TempDir()
	home := filepath.Join(base, "home \"quoted\" \\ path")
	if err := os.MkdirAll(filepath.Join(home, ".local", "bin"), 0755); err != nil {
		t.Fatal(err)
	}
	exe, codex := filepath.Join(home, "my-friday"), filepath.Join(home, "codex")
	for path, body := range map[string]string{exe: "launcher", codex: "codex"} {
		if err := os.WriteFile(path, []byte(body), 0700); err != nil {
			t.Fatal(err)
		}
	}
	p, err := PlanCreate(home, "alfred", exe, codex)
	if err != nil {
		t.Fatal(err)
	}
	if err = Create(p, exe, codex); err != nil {
		t.Fatal(err)
	}
	config, err := os.ReadFile(filepath.Join(p.Paths.Root, "codex", "config.toml"))
	if err != nil {
		t.Fatal(err)
	}
	escapedWorkspace := strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(filepath.Join(p.Paths.Root, "workspace"))
	escapedRuntime := strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(filepath.Join(p.Paths.Root, "runtime"))
	want := "approval_policy = \"never\"\nsandbox_mode = \"workspace-write\"\n\n[features]\ncode_mode_host = false\n\n[notice]\nhide_rate_limit_model_nudge = true\n\n[sandbox_workspace_write]\nnetwork_access = false\nwritable_roots = [\"" + escapedRuntime + "\"]\n\n[projects.\"" + escapedWorkspace + "\"]\ntrust_level = \"trusted\"\n"
	if string(config) != want {
		t.Fatalf("unexpected instance trust config\nwant: %q\n got: %q", want, config)
	}
	if strings.Count(string(config), "writable_roots") != 1 || strings.Contains(string(config), "danger-full-access") || strings.Contains(string(config), "network_access = true") {
		t.Fatalf("broader sandbox authority rendered: %q", config)
	}
}

func TestVerifyAllowsOnlyCodexModelAvailabilityMetadata(t *testing.T) {
	home, exe, codex := fixture(t)
	p, err := PlanCreate(home, "alfred", exe, codex)
	if err != nil {
		t.Fatal(err)
	}
	if err = Create(p, exe, codex); err != nil {
		t.Fatal(err)
	}
	config := filepath.Join(p.Paths.Root, "codex", "config.toml")
	original, err := os.ReadFile(config)
	if err != nil {
		t.Fatal(err)
	}
	allowed := append(append([]byte{}, original...), []byte("\n[tui.model_availability_nux]\ngpt-6-astra = 1\n\"gpt-5.6-luna\" = 2\n")...)
	if err = os.WriteFile(config, allowed, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err = Verify(home, "alfred"); err != nil {
		t.Fatalf("bounded Codex-owned TUI metadata refused: %v", err)
	}
	bad := append(append([]byte{}, allowed...), []byte("sandbox_mode = \"danger-full-access\"\n")...)
	if err = os.WriteFile(config, bad, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err = Verify(home, "alfred"); err == nil || !strings.Contains(err.Error(), "managed Codex config drift") {
		t.Fatalf("security-relevant config suffix accepted: %v", err)
	}
}

func TestCreateBindsInstanceSpecificBuilderAndMyFridayExecutable(t *testing.T) {
	home, exe, codex := fixture(t)
	paths := make(map[string]Paths)
	builders := make(map[string][]byte)
	for _, name := range []string{"alfred", "robin"} {
		p, err := PlanCreate(home, name, exe, codex)
		if err != nil {
			t.Fatal(err)
		}
		if err = Create(p, exe, codex); err != nil {
			t.Fatal(err)
		}
		paths[name] = p.Paths
		m, err := Verify(home, name)
		if err != nil {
			t.Fatal(err)
		}
		wantExecutable := filepath.Join(p.Paths.Root, "dependencies", "my-friday")
		if m.MyFridayExecutable != wantExecutable || m.MyFridaySHA256 != digest([]byte("launcher")) {
			t.Fatalf("managed executable is not manifest-bound: %#v", m)
		}
		body, err := os.ReadFile(filepath.Join(m.CapabilityBuilder, "SKILL.md"))
		if err != nil {
			t.Fatal(err)
		}
		builders[name] = body
		for _, exact := range []string{p.Paths.Root + "/runtime", wantExecutable, "capability inspect " + name, "capability validate " + name, "capability test " + name} {
			if !strings.Contains(string(body), exact) {
				t.Fatalf("builder lacks %q: %s", exact, body)
			}
		}
		for _, forbidden := range []string{" capability install ", " capability upgrade ", " capability enable ", " capability disable ", " capability remove ", " capability recover "} {
			if strings.Contains(string(body), forbidden) {
				t.Fatalf("builder grants mutation command %q", forbidden)
			}
		}
	}
	if string(builders["alfred"]) == string(builders["robin"]) || strings.Contains(string(builders["alfred"]), paths["robin"].Root) {
		t.Fatal("builder bytes are not instance-specific")
	}
}

func TestManagedMyFridayDriftRefusesVerifyRemoveAndRecovery(t *testing.T) {
	home, exe, codex := fixture(t)
	p, err := PlanCreate(home, "alfred", exe, codex)
	if err != nil {
		t.Fatal(err)
	}
	if err = Create(p, exe, codex); err != nil {
		t.Fatal(err)
	}
	managed := filepath.Join(p.Paths.Root, "dependencies", "my-friday")
	if err = os.WriteFile(managed, []byte("drift"), 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err = Verify(home, "alfred"); err == nil || !strings.Contains(err.Error(), "My Friday executable drift") {
		t.Fatalf("managed executable drift accepted: %v", err)
	}
	if _, err = PlanRemove(home, "alfred"); err == nil {
		t.Fatal("managed executable drift granted removal authority")
	}
	if err = os.Remove(p.Paths.Launcher); err != nil {
		t.Fatal(err)
	}
	if _, err = Recover(home, "alfred"); err == nil || !strings.Contains(err.Error(), "My Friday executable drift") {
		t.Fatalf("managed executable drift granted recovery authority: %v", err)
	}
}

func TestInstanceWorkspaceTrustIsSeparatedAndTamperFailsClosed(t *testing.T) {
	home, exe, codex := fixture(t)
	paths := make(map[string]Paths)
	for _, name := range []string{"alfred", "robin"} {
		p, err := PlanCreate(home, name, exe, codex)
		if err != nil {
			t.Fatal(err)
		}
		if err = Create(p, exe, codex); err != nil {
			t.Fatal(err)
		}
		paths[name] = p.Paths
	}
	alfredConfig, err := os.ReadFile(filepath.Join(paths["alfred"].Root, "codex", "config.toml"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(alfredConfig), paths["robin"].Root) {
		t.Fatal("one instance trusted another instance")
	}
	configPath := filepath.Join(paths["alfred"].Root, "codex", "config.toml")
	if err = os.WriteFile(configPath, append(alfredConfig, []byte("[projects.\"/tmp\"]\ntrust_level = \"trusted\"\n")...), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = Verify(home, "alfred"); err == nil || !strings.Contains(err.Error(), "config") {
		t.Fatalf("tampered trust config accepted: %v", err)
	}
	if _, err = PlanRemove(home, "alfred"); err == nil {
		t.Fatal("tampered trust config granted removal authority")
	}
	if err = os.Remove(paths["alfred"].Launcher); err != nil {
		t.Fatal(err)
	}
	if _, err = Recover(home, "alfred"); err == nil || !strings.Contains(err.Error(), "config") {
		t.Fatalf("tampered trust config granted recovery authority: %v", err)
	}
	if _, err = Verify(home, "robin"); err != nil {
		t.Fatalf("sibling instance changed: %v", err)
	}
}

func TestForgedManagedConfigManifestIsDenied(t *testing.T) {
	for _, mutate := range []func(*Manifest){
		func(m *Manifest) { m.CodexConfig = filepath.Join(m.Root, "codex", "other.toml") },
		func(m *Manifest) { m.CodexConfigSHA256 = strings.Repeat("a", 64) },
	} {
		home, exe, codex := fixture(t)
		p, err := PlanCreate(home, "alfred", exe, codex)
		if err != nil {
			t.Fatal(err)
		}
		if err = Create(p, exe, codex); err != nil {
			t.Fatal(err)
		}
		manifestPath := filepath.Join(p.Paths.Root, "manifest.json")
		b, err := os.ReadFile(manifestPath)
		if err != nil {
			t.Fatal(err)
		}
		var m Manifest
		if err = json.Unmarshal(b, &m); err != nil {
			t.Fatal(err)
		}
		mutate(&m)
		b, err = json.MarshalIndent(m, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(manifestPath, append(b, '\n'), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err = Verify(home, "alfred"); err == nil || !strings.Contains(err.Error(), "manifest ownership contract mismatch") {
			t.Fatalf("forged config manifest accepted: %v", err)
		}
	}
}

func TestRecoverCompletesInterruptedRemoval(t *testing.T) {
	home, exe, codex := fixture(t)
	p, err := PlanCreate(home, "alfred", exe, codex)
	if err != nil {
		t.Fatal(err)
	}
	if err = Create(p, exe, codex); err != nil {
		t.Fatal(err)
	}
	if err = os.Remove(p.Paths.Launcher); err != nil {
		t.Fatal(err)
	}
	result, err := Recover(home, "alfred")
	if err != nil {
		t.Fatal(err)
	}
	if result != "restored absent state" {
		t.Fatal(result)
	}
	if _, err = os.Stat(p.Paths.Root); !os.IsNotExist(err) {
		t.Fatal("recovery retained root")
	}
}

func TestSameNameCreateIsSerialized(t *testing.T) {
	home, exe, codex := fixture(t)
	p, err := PlanCreate(home, "alfred", exe, codex)
	if err != nil {
		t.Fatal(err)
	}
	entered, release := make(chan struct{}, 2), make(chan struct{})
	mutationHook = func(phase string) {
		if phase == "create-preflight" {
			entered <- struct{}{}
			<-release
		}
	}
	defer func() { mutationHook = nil }()
	results := make(chan error, 2)
	go func() { results <- Create(p, exe, codex) }()
	<-entered
	go func() { results <- Create(p, exe, codex) }()
	select {
	case <-entered:
		t.Fatal("second same-name create entered mutation while first held lock")
	case <-time.After(100 * time.Millisecond):
	}
	close(release)
	err1, err2 := <-results, <-results
	if (err1 == nil) == (err2 == nil) {
		t.Fatalf("want one success and one collision, got %v and %v", err1, err2)
	}
	if _, err = Verify(home, "alfred"); err != nil {
		t.Fatal(err)
	}
	if _, err = os.Stat(p.Paths.Root + ".creating"); !os.IsNotExist(err) {
		t.Fatal("shared stage remains")
	}
}

func TestMigrationSuccessAndCleanupFault(t *testing.T) {
	for _, tc := range []struct {
		name       string
		cleanupErr error
	}{{"success", nil}, {"cleanup-fault", errors.New("injected legacy cleanup fault")}} {
		t.Run(tc.name, func(t *testing.T) {
			home, exe, codex := fixture(t)
			p, err := PlanCreate(home, "alfred", exe, codex)
			if err != nil {
				t.Fatal(err)
			}
			cleanupCalled := false
			err = Migrate(p, exe, codex, func() error {
				cleanupCalled = true
				return tc.cleanupErr
			})
			if !cleanupCalled {
				t.Fatal("legacy cleanup was not called")
			}
			if tc.cleanupErr == nil && err != nil {
				t.Fatal(err)
			}
			if tc.cleanupErr != nil && (err == nil || !strings.Contains(err.Error(), tc.cleanupErr.Error())) {
				t.Fatalf("cleanup fault not retained: %v", err)
			}
			if _, verifyErr := Verify(home, "alfred"); verifyErr != nil {
				t.Fatalf("named replacement not retained: %v", verifyErr)
			}
		})
	}
}

func TestLauncherEnvironmentAndArguments(t *testing.T) {
	if os.Getenv("MY_FRIDAY_LAUNCH_HELPER") == "1" {
		if err := Launch(os.Getenv("MY_FRIDAY_TEST_HOME"), "alfred", []string{"hello world", "--cd", "/foreign"}); err != nil {
			t.Fatal(err)
		}
		return
	}
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, ".local", "bin"), 0755); err != nil {
		t.Fatal(err)
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	capture := filepath.Join(home, "capture")
	codex := filepath.Join(home, "codex-stub")
	stub := "#!/bin/sh\nprintf '%s\\n' \"$HOME\" \"$CODEX_HOME\" \"$@\" > \"$MY_FRIDAY_CAPTURE\"\n"
	if err = os.WriteFile(codex, []byte(stub), 0700); err != nil {
		t.Fatal(err)
	}
	p, err := PlanCreate(home, "alfred", exe, codex)
	if err != nil {
		t.Fatal(err)
	}
	if err = Create(p, exe, codex); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(exe, "-test.run=TestLauncherEnvironmentAndArguments")
	cmd.Env = append(os.Environ(), "MY_FRIDAY_LAUNCH_HELPER=1", "MY_FRIDAY_TEST_HOME="+home, "MY_FRIDAY_CAPTURE="+capture, "HOME="+home, "CODEX_HOME=/foreign/codex")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("launch helper: %v: %s", err, out)
	}
	b, err := os.ReadFile(capture)
	if err != nil {
		t.Fatal(err)
	}
	want := strings.Join([]string{home, filepath.Join(p.Paths.Root, "codex"), "--cd", filepath.Join(p.Paths.Root, "workspace"), "--", "hello world", "--cd", "/foreign", ""}, "\n")
	if string(b) != want {
		t.Fatalf("launcher contract\nwant: %q\n got: %q", want, b)
	}
}
