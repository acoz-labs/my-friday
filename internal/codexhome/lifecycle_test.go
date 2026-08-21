package codexhome

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	bootstrap "github.com/acoz-labs/my-friday/internal/plan"
	"github.com/acoz-labs/my-friday/internal/profile"
	"github.com/acoz-labs/my-friday/internal/repository"
	"golang.org/x/sys/unix"
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

}

func installFixture(t *testing.T) (string, string) {
	t.Helper()
	runtime, codex := fixture(t)
	if err := os.MkdirAll(codex, 0700); err != nil {
		t.Fatal(err)
	}
	p, err := Plan(ActionInstall, runtime, codex)
	if err != nil {
		t.Fatal(err)
	}
	if err = Execute(p); err != nil {
		t.Fatalf("install: %v", err)
	}
	return runtime, codex
}

func mutatePurpose(t *testing.T, runtime, purpose string) {
	t.Helper()
	path := filepath.Join(runtime, "assistant", "profile.json")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var p profile.Profile
	if err = json.Unmarshal(b, &p); err != nil {
		t.Fatal(err)
	}
	p.Identity.Purpose = purpose
	b, _ = json.MarshalIndent(p, "", "  ")
	if err = os.WriteFile(path, append(b, '\n'), 0600); err != nil {
		t.Fatal(err)
	}
}

func TestForeignControlCanaryAlwaysFailsClosed(t *testing.T) {
	runtime, codex := installFixture(t)
	canary := filepath.Join(codex, controlDir, "foreign-canary")
	if err := os.WriteFile(canary, []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Plan(ActionUninstall, "", codex); err == nil {
		t.Fatal("uninstall accepted foreign control content")
	}
	if err := Recover(codex); err == nil {
		t.Fatal("recovery accepted foreign control content")
	}
	if got, _ := os.ReadFile(canary); string(got) != "keep" {
		t.Fatalf("foreign canary changed: %q", got)
	}
	_ = runtime
}

func TestSymlinkedHomeAncestorCannotRedirectManagedWrites(t *testing.T) {
	runtime, _ := fixture(t)
	root := t.TempDir()
	realHome := filepath.Join(root, "real", ".codex")
	if err := os.MkdirAll(realHome, 0700); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "link")
	if err := os.Symlink(filepath.Join(root, "real"), link); err != nil {
		t.Fatal(err)
	}
	if _, err := Plan(ActionInstall, runtime, filepath.Join(link, ".codex")); err == nil {
		t.Fatal("symlinked ancestor accepted")
	}
	if _, err := os.Stat(filepath.Join(realHome, "AGENTS.md")); !os.IsNotExist(err) {
		t.Fatalf("redirected write occurred: %v", err)
	}
}

func TestExecuteRefusesAnyPostPreviewChange(t *testing.T) {
	runtime, codex := fixture(t)
	if err := os.MkdirAll(codex, 0700); err != nil {
		t.Fatal(err)
	}
	p, err := Plan(ActionInstall, runtime, codex)
	if err != nil {
		t.Fatal(err)
	}
	mutatePurpose(t, runtime, "Changed after preview")
	if err = Execute(p); err == nil || !strings.Contains(err.Error(), "confirmed preview") {
		t.Fatalf("preview race accepted: %v", err)
	}
	if _, err = os.Stat(filepath.Join(codex, "AGENTS.md")); !os.IsNotExist(err) {
		t.Fatalf("preview race mutated target: %v", err)
	}
}

func TestExecuteRefusesManagedTargetSwapAtMutationBoundary(t *testing.T) {
	runtime, codex := installFixture(t)
	mutatePurpose(t, runtime, "Changed generation")
	p, err := Plan(ActionUpgrade, runtime, codex)
	if err != nil {
		t.Fatal(err)
	}
	faultHook = func(phase string) error {
		if phase == "mutating" {
			return os.WriteFile(filepath.Join(codex, "AGENTS.md"), []byte("concurrent foreign bytes"), 0600)
		}
		return nil
	}
	t.Cleanup(func() { faultHook = nil })
	if err = Execute(p); err == nil || !strings.Contains(err.Error(), "changed before mutation") {
		t.Fatalf("managed target race accepted: %v", err)
	}
	faultHook = nil
	if err = Recover(codex); err == nil || !strings.Contains(err.Error(), "neither journal generation") {
		t.Fatalf("recovery adopted concurrent bytes: %v", err)
	}
	got, _ := os.ReadFile(filepath.Join(codex, "AGENTS.md"))
	if string(got) != "concurrent foreign bytes" {
		t.Fatalf("concurrent bytes changed: %q", got)
	}
}

func TestRepairRefusesChangedSourceAndRestoresCanonicalBytes(t *testing.T) {
	runtime, codex := installFixture(t)
	projection := filepath.Join(codex, "AGENTS.md")
	if err := os.WriteFile(projection, []byte("drift"), 0600); err != nil {
		t.Fatal(err)
	}
	mutatePurpose(t, runtime, "Changed source")
	if _, err := Plan(ActionRepair, runtime, codex); err == nil || !strings.Contains(err.Error(), "recorded source changed") {
		t.Fatalf("repair upgraded changed source: %v", err)
	}
	mutatePurpose(t, runtime, "Help with careful work")
	p, err := Plan(ActionRepair, runtime, codex)
	if err != nil || Execute(p) != nil {
		t.Fatalf("repair failed: %v", err)
	}
	got, _ := os.ReadFile(projection)
	if string(got) == "drift" {
		t.Fatal("repair did not restore canonical generation")
	}
}

func TestManifestAuthorityIsStrict(t *testing.T) {
	_, codex := installFixture(t)
	path := filepath.Join(codex, controlDir, manifestFile)
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	b = bytes.Replace(b, []byte("\n}"), []byte(",\n  \"foreign\": true\n}"), 1)
	if err = os.WriteFile(path, b, 0600); err != nil {
		t.Fatal(err)
	}
	if s, _ := Inspect("", codex); s.State != StateCollision {
		t.Fatalf("unknown manifest authority accepted: %s", s.State)
	}
}

func TestPostInstallOverrideIsUnhealthy(t *testing.T) {
	_, codex := installFixture(t)
	if err := os.WriteFile(filepath.Join(codex, "AGENTS.override.md"), []byte("shadow"), 0600); err != nil {
		t.Fatal(err)
	}
	if s, _ := Inspect("", codex); s.State != StateCollision || !strings.Contains(s.Detail, "shadows") {
		t.Fatalf("shadowing state=%s detail=%q", s.State, s.Detail)
	}
}

func TestRollbackAndUninstallDoNotRequireSource(t *testing.T) {
	runtime, codex := installFixture(t)
	mutatePurpose(t, runtime, "Second generation")
	p, err := Plan(ActionUpgrade, runtime, codex)
	if err != nil || Execute(p) != nil {
		t.Fatalf("upgrade: %v", err)
	}
	if err = os.Rename(runtime, runtime+"-gone"); err != nil {
		t.Fatal(err)
	}
	p, err = Plan(ActionRollback, "", codex)
	if err != nil || Execute(p) != nil {
		t.Fatalf("source-independent rollback: %v", err)
	}
	p, err = Plan(ActionUninstall, "", codex)
	if err != nil || Execute(p) != nil {
		t.Fatalf("source-independent uninstall: %v", err)
	}
}

func TestRecoveryAtEveryDurablePhase(t *testing.T) {
	phases := []string{"prepared", "mutating", "projection-written", "canonical-written", "previous-written", "manifest-written", "committed", "final-verified"}
	for _, phase := range phases {
		t.Run(phase, func(t *testing.T) {
			runtime, codex := installFixture(t)
			mutatePurpose(t, runtime, "New generation")
			p, err := Plan(ActionUpgrade, runtime, codex)
			if err != nil {
				t.Fatal(err)
			}
			faultHook = func(got string) error {
				if got == phase {
					return errors.New("injected interruption")
				}
				return nil
			}
			t.Cleanup(func() { faultHook = nil })
			if err = Execute(p); err == nil {
				t.Fatal("fault did not interrupt")
			}
			faultHook = nil
			if err = Recover(codex); err != nil {
				t.Fatal(err)
			}
			s, _ := Inspect("", codex)
			if s.State != StateHealthy && s.State != StateSourceDrift {
				t.Fatalf("unrecovered state: %s", s.State)
			}
			if _, err = os.Stat(filepath.Join(codex, controlDir, journalFile)); !os.IsNotExist(err) {
				t.Fatalf("journal remains: %v", err)
			}
		})
	}
}

func TestConcurrentControlDirectoryReplacementIsRefused(t *testing.T) {
	runtime, codex := installFixture(t)
	mutatePurpose(t, runtime, "Replacement race")
	p, err := Plan(ActionUpgrade, runtime, codex)
	if err != nil {
		t.Fatal(err)
	}
	faultHook = func(phase string) error {
		if phase != "mutating" {
			return nil
		}
		owned := filepath.Join(codex, controlDir)
		if err := os.Rename(owned, owned+"-detached"); err != nil {
			return err
		}
		return os.Mkdir(owned, 0700)
	}
	t.Cleanup(func() { faultHook = nil })
	if err = Execute(p); err == nil || !strings.Contains(err.Error(), "control directory identity changed") {
		t.Fatalf("control replacement accepted: %v", err)
	}
}

func TestForgedActionJournalIsRefused(t *testing.T) {
	_, codex := installFixture(t)
	r, err := openHome(codex)
	if err != nil {
		t.Fatal(err)
	}
	before, err := snapshot(r)
	root := r.proof
	r.close()
	if err != nil {
		t.Fatal(err)
	}
	forged := journal{ContractVersion: 1, Action: ActionUninstall, Phase: "committed", Root: root, Before: before, After: before}
	b, _ := json.MarshalIndent(forged, "", "  ")
	if err = os.WriteFile(filepath.Join(codex, controlDir, journalFile), append(b, '\n'), 0600); err != nil {
		t.Fatal(err)
	}
	if err = Recover(codex); err == nil || !strings.Contains(err.Error(), "incompatible or ambiguous journal") {
		t.Fatalf("forged uninstall journal accepted: %v", err)
	}
}

func TestFinalVerificationDetectsPostMutationRace(t *testing.T) {
	runtime, codex := installFixture(t)
	mutatePurpose(t, runtime, "Final verification race")
	p, err := Plan(ActionUpgrade, runtime, codex)
	if err != nil {
		t.Fatal(err)
	}
	faultHook = func(phase string) error {
		if phase == "before-final-verification" {
			return os.WriteFile(filepath.Join(codex, "AGENTS.md"), []byte("foreign final bytes"), 0600)
		}
		return nil
	}
	t.Cleanup(func() { faultHook = nil })
	if err = Execute(p); err == nil || !strings.Contains(err.Error(), "final managed state") {
		t.Fatalf("final race accepted: %v", err)
	}
	if got, _ := os.ReadFile(filepath.Join(codex, "AGENTS.md")); string(got) != "foreign final bytes" {
		t.Fatalf("foreign bytes overwritten: %q", got)
	}
}

func TestCommittedUninstallDeletionIsRecoverable(t *testing.T) {
	for _, interrupted := range []string{"control-detached", "removal-marked"} {
		t.Run(interrupted, func(t *testing.T) {
			_, codex := installFixture(t)
			p, err := Plan(ActionUninstall, "", codex)
			if err != nil {
				t.Fatal(err)
			}
			faultHook = func(phase string) error {
				if phase == interrupted {
					return errors.New("injected interruption")
				}
				return nil
			}
			t.Cleanup(func() { faultHook = nil })
			if err = Execute(p); err == nil {
				t.Fatal("fault did not interrupt uninstall")
			}
			faultHook = nil
			if err = Recover(codex); err != nil {
				t.Fatal(err)
			}
			if _, err = os.Stat(filepath.Join(codex, removingDir)); !os.IsNotExist(err) {
				t.Fatalf("deletion namespace remains: %v", err)
			}
			if _, err = os.Stat(filepath.Join(codex, removalMarker)); !os.IsNotExist(err) {
				t.Fatalf("deletion marker remains: %v", err)
			}
			if s, _ := Inspect("", codex); s.State != StateNotInstalled {
				t.Fatalf("state=%s", s.State)
			}
		})
	}
}

func TestSupportedHomeEnvironmentRequiresNonRootLocalAPFS(t *testing.T) {
	for _, tc := range []struct {
		name  string
		euid  int
		fs    string
		local bool
	}{
		{"root", 0, "apfs", true},
		{"network apfs", 501, "apfs", false},
		{"local non-apfs", 501, "nfs", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := validateHomeEnvironment(tc.euid, tc.fs, tc.local); err == nil {
				t.Fatal("unsupported environment accepted")
			}
		})
	}
	if err := validateHomeEnvironment(501, "apfs", true); err != nil {
		t.Fatalf("supported environment refused: %v", err)
	}
}

func TestPartialInitialJournalPublicationRecoversWithoutMutation(t *testing.T) {
	_, codex := fixture(t)
	if err := os.MkdirAll(filepath.Join(codex, controlDir), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(codex, controlDir, journalNext), []byte("{\"contract_version\":"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := Recover(codex); err == nil || !strings.Contains(err.Error(), "unproven staged journal") {
		t.Fatalf("malformed staged authority was not retained: %v", err)
	}
	if got, err := os.ReadFile(filepath.Join(codex, controlDir, journalNext)); err != nil || string(got) != "{\"contract_version\":" {
		t.Fatalf("staged authority changed: %q, %v", got, err)
	}
	if _, err := os.Stat(filepath.Join(codex, "AGENTS.md")); !os.IsNotExist(err) {
		t.Fatalf("projection was mutated: %v", err)
	}
}

func TestRecoveryUsesNewerJournalWhenPriorPhaseRemainsStaged(t *testing.T) {
	runtime, codex := installFixture(t)
	mutatePurpose(t, runtime, "Journal swap interruption")
	p, err := Plan(ActionUpgrade, runtime, codex)
	if err != nil {
		t.Fatal(err)
	}
	faultHook = func(phase string) error {
		if phase == "journal-committed-swapped" {
			return errors.New("injected journal swap interruption")
		}
		return nil
	}
	t.Cleanup(func() { faultHook = nil })
	if err = Execute(p); err == nil {
		t.Fatal("journal swap interruption did not fire")
	}
	faultHook = nil
	if err = Recover(codex); err != nil {
		t.Fatal(err)
	}
	if err = Recover(codex); err != nil {
		t.Fatalf("second recovery reversed the committed result: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(codex, "AGENTS.md"))
	if err != nil || digest(got) != p.manifest.ProjectionSHA256 {
		t.Fatalf("committed projection was not retained: %v", err)
	}
}

func TestRecoveryUsesNewerStagedJournalWhenCurrentPhasePrecedesIt(t *testing.T) {
	runtime, codex := installFixture(t)
	mutatePurpose(t, runtime, "Journal pre-swap interruption")
	p, err := Plan(ActionUpgrade, runtime, codex)
	if err != nil {
		t.Fatal(err)
	}
	faultHook = func(phase string) error {
		if phase == "journal-committed-staged" {
			return errors.New("injected journal staging interruption")
		}
		return nil
	}
	t.Cleanup(func() { faultHook = nil })
	if err = Execute(p); err == nil {
		t.Fatal("journal staging interruption did not fire")
	}
	faultHook = nil
	if err = Recover(codex); err != nil {
		t.Fatal(err)
	}
	if err = Recover(codex); err != nil {
		t.Fatalf("second recovery reversed the committed result: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(codex, "AGENTS.md"))
	if err != nil || digest(got) != p.manifest.ProjectionSHA256 {
		t.Fatalf("committed projection was not retained: %v", err)
	}
}

func TestRecoveryPreservesStagedJournalWhenPinnedRootDiffers(t *testing.T) {
	for _, tc := range []struct {
		name      string
		bothSlots bool
	}{
		{"stage only", false},
		{"both slots", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			runtime, codex := fixture(t)
			if err := os.MkdirAll(codex, 0700); err != nil {
				t.Fatal(err)
			}
			p, err := Plan(ActionInstall, runtime, codex)
			if err != nil {
				t.Fatal(err)
			}
			j := journal{ContractVersion: 1, Action: ActionInstall, Phase: "prepared", Root: p.root, Before: p.before, After: afterProofs(p)}
			j.Root.Inode++
			if err = os.Mkdir(filepath.Join(codex, controlDir), 0700); err != nil {
				t.Fatal(err)
			}
			next := journalBytes(j)
			if err = os.WriteFile(filepath.Join(codex, controlDir, journalNext), next, 0600); err != nil {
				t.Fatal(err)
			}
			if tc.bothSlots {
				j.Root = p.root
				j.Phase = "mutating"
				if err = os.WriteFile(filepath.Join(codex, controlDir, journalFile), journalBytes(j), 0600); err != nil {
					t.Fatal(err)
				}
			}
			before, err := os.ReadFile(filepath.Join(codex, controlDir, journalNext))
			if err != nil {
				t.Fatal(err)
			}
			if err = Recover(codex); err == nil {
				t.Fatal("mismatched pinned root was accepted")
			}
			after, readErr := os.ReadFile(filepath.Join(codex, controlDir, journalNext))
			if readErr != nil || !bytes.Equal(after, before) {
				t.Fatalf("staged evidence changed: %v", readErr)
			}
		})
	}
}

func TestDetachedRecoveryPreservesJournalWhenPinnedRootDiffers(t *testing.T) {
	runtime, codex := fixture(t)
	if err := os.MkdirAll(codex, 0700); err != nil {
		t.Fatal(err)
	}
	p, err := Plan(ActionInstall, runtime, codex)
	if err != nil {
		t.Fatal(err)
	}
	j := journal{ContractVersion: 1, Action: ActionInstall, Phase: "prepared", Root: p.root, Before: p.before, After: afterProofs(p)}
	j.Root.Inode++
	if err = os.Mkdir(filepath.Join(codex, removingDir), 0700); err != nil {
		t.Fatal(err)
	}
	want := journalBytes(j)
	journalPath := filepath.Join(codex, removingDir, journalFile)
	if err = os.WriteFile(journalPath, want, 0600); err != nil {
		t.Fatal(err)
	}
	if err = Recover(codex); err == nil {
		t.Fatal("mismatched detached authority was accepted")
	}
	got, readErr := os.ReadFile(journalPath)
	if readErr != nil || !bytes.Equal(got, want) {
		t.Fatalf("detached journal evidence changed: %v", readErr)
	}
	if _, statErr := os.Stat(filepath.Join(codex, removalMarker)); !os.IsNotExist(statErr) {
		t.Fatalf("detached journal was promoted before validation: %v", statErr)
	}
}

func TestStageOnlyRecoveryPreservesConcurrentJournalReplacement(t *testing.T) {
	runtime, codex := fixture(t)
	if err := os.MkdirAll(codex, 0700); err != nil {
		t.Fatal(err)
	}
	p, err := Plan(ActionInstall, runtime, codex)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Mkdir(filepath.Join(codex, controlDir), 0700); err != nil {
		t.Fatal(err)
	}
	nextPath := filepath.Join(codex, controlDir, journalNext)
	j := journal{ContractVersion: 1, Action: ActionInstall, Phase: "prepared", Root: p.root, Before: p.before, After: afterProofs(p)}
	if err = os.WriteFile(nextPath, journalBytes(j), 0600); err != nil {
		t.Fatal(err)
	}
	foreign := []byte("concurrent foreign staged journal")
	recoveryHook = func(point string) error {
		if point == "stage-only-before-promote" {
			if err := os.Remove(nextPath); err != nil {
				return err
			}
			return os.WriteFile(nextPath, foreign, 0600)
		}
		return nil
	}
	t.Cleanup(func() { recoveryHook = nil })
	if err = Recover(codex); err == nil {
		t.Fatal("concurrent stage-only replacement was accepted")
	}
	got, readErr := os.ReadFile(nextPath)
	if readErr != nil || !bytes.Equal(got, foreign) {
		t.Fatalf("foreign staged journal was not retained: %q, %v", got, readErr)
	}
	if _, statErr := os.Stat(filepath.Join(codex, controlDir, journalFile)); !os.IsNotExist(statErr) {
		t.Fatalf("stale authority was promoted: %v", statErr)
	}
}

func TestDualSlotRecoveryPreservesConcurrentSwapReplacement(t *testing.T) {
	runtime, codex := installFixture(t)
	mutatePurpose(t, runtime, "Dual-slot race")
	p, err := Plan(ActionUpgrade, runtime, codex)
	if err != nil {
		t.Fatal(err)
	}
	faultHook = func(phase string) error {
		if phase == "journal-committed-staged" {
			return errors.New("leave newer staged journal")
		}
		return nil
	}
	if err = Execute(p); err == nil {
		t.Fatal("journal staging interruption did not fire")
	}
	faultHook = nil
	nextPath := filepath.Join(codex, controlDir, journalNext)
	foreign := []byte("concurrent foreign dual-slot journal")
	recoveryHook = func(point string) error {
		if point == "dual-slot-before-swap" {
			if err := os.Remove(nextPath); err != nil {
				return err
			}
			return os.WriteFile(nextPath, foreign, 0600)
		}
		return nil
	}
	t.Cleanup(func() { faultHook = nil; recoveryHook = nil })
	if err = Recover(codex); err == nil {
		t.Fatal("concurrent dual-slot replacement was accepted")
	}
	got, readErr := os.ReadFile(nextPath)
	if readErr != nil || !bytes.Equal(got, foreign) {
		t.Fatalf("foreign dual-slot journal was not retained: %q, %v", got, readErr)
	}
}

func TestDualSlotRecoveryPreservesConcurrentDiscardReplacement(t *testing.T) {
	runtime, codex := installFixture(t)
	mutatePurpose(t, runtime, "Dual-slot discard race")
	p, err := Plan(ActionUpgrade, runtime, codex)
	if err != nil {
		t.Fatal(err)
	}
	faultHook = func(phase string) error {
		if phase == "journal-committed-swapped" {
			return errors.New("leave prior journal staged")
		}
		return nil
	}
	if err = Execute(p); err == nil {
		t.Fatal("journal swap interruption did not fire")
	}
	faultHook = nil
	discardPath := filepath.Join(codex, controlDir, journalDiscard)
	foreign := []byte("concurrent foreign discard journal")
	recoveryHook = func(point string) error {
		if point == "dual-slot-before-discard-unlink" {
			if err := os.Remove(discardPath); err != nil {
				return err
			}
			return os.WriteFile(discardPath, foreign, 0600)
		}
		return nil
	}
	t.Cleanup(func() { faultHook = nil; recoveryHook = nil })
	if err = Recover(codex); err == nil {
		t.Fatal("concurrent discard replacement was accepted")
	}
	got, readErr := os.ReadFile(filepath.Join(codex, controlDir, journalNext))
	if readErr != nil || !bytes.Equal(got, foreign) {
		t.Fatalf("foreign discard journal was not restored: %q, %v", got, readErr)
	}
}

func TestInterruptedRecoveryDiscardStageResumes(t *testing.T) {
	runtime, codex := installFixture(t)
	mutatePurpose(t, runtime, "Interrupted recovery discard")
	p, err := Plan(ActionUpgrade, runtime, codex)
	if err != nil {
		t.Fatal(err)
	}
	faultHook = func(phase string) error {
		if phase == "journal-committed-swapped" {
			return errors.New("leave prior journal staged")
		}
		return nil
	}
	if err = Execute(p); err == nil {
		t.Fatal("journal swap interruption did not fire")
	}
	faultHook = nil
	recoveryHook = func(point string) error {
		if point == "dual-slot-before-discard-unlink" {
			return errors.New("interrupt discard cleanup")
		}
		return nil
	}
	if err = Recover(codex); err == nil {
		t.Fatal("discard cleanup interruption did not fire")
	}
	recoveryHook = nil
	t.Cleanup(func() { faultHook = nil; recoveryHook = nil })
	if err = Recover(codex); err != nil {
		t.Fatalf("interrupted discard cleanup did not resume: %v", err)
	}
	if err = Recover(codex); err != nil {
		t.Fatalf("repeat recovery failed: %v", err)
	}
}

func TestDetachedRecoveryPreservesConcurrentJournalReplacement(t *testing.T) {
	runtime, codex := fixture(t)
	if err := os.MkdirAll(codex, 0700); err != nil {
		t.Fatal(err)
	}
	p, err := Plan(ActionInstall, runtime, codex)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Mkdir(filepath.Join(codex, removingDir), 0700); err != nil {
		t.Fatal(err)
	}
	journalPath := filepath.Join(codex, removingDir, journalFile)
	j := journal{ContractVersion: 1, Action: ActionInstall, Phase: "prepared", Root: p.root, Before: p.before, After: afterProofs(p)}
	if err = os.WriteFile(journalPath, journalBytes(j), 0600); err != nil {
		t.Fatal(err)
	}
	foreign := []byte("concurrent foreign detached journal")
	recoveryHook = func(point string) error {
		if point == "detached-before-promote" {
			if err := os.Remove(journalPath); err != nil {
				return err
			}
			return os.WriteFile(journalPath, foreign, 0600)
		}
		return nil
	}
	t.Cleanup(func() { recoveryHook = nil })
	if err = Recover(codex); err == nil {
		t.Fatal("concurrent detached replacement was accepted")
	}
	got, readErr := os.ReadFile(journalPath)
	if readErr != nil || !bytes.Equal(got, foreign) {
		t.Fatalf("foreign detached journal was not retained: %q, %v", got, readErr)
	}
	if _, statErr := os.Stat(filepath.Join(codex, removalMarker)); !os.IsNotExist(statErr) {
		t.Fatalf("foreign detached journal was promoted: %v", statErr)
	}
}

func TestCommittedUninstallRecoversWithPriorJournalPhaseStaged(t *testing.T) {
	_, codex := installFixture(t)
	p, err := Plan(ActionUninstall, "", codex)
	if err != nil {
		t.Fatal(err)
	}
	faultHook = func(phase string) error {
		if phase == "journal-committed-swapped" {
			return errors.New("injected journal swap interruption")
		}
		return nil
	}
	t.Cleanup(func() { faultHook = nil })
	if err = Execute(p); err == nil {
		t.Fatal("journal swap interruption did not fire")
	}
	faultHook = nil
	if err = Recover(codex); err != nil {
		t.Fatal(err)
	}
	if s, _ := Inspect("", codex); s.State != StateNotInstalled {
		t.Fatalf("state=%s", s.State)
	}
}

func TestInterruptedInitialInstallRollbackCannotLeaveEmptyControlDirectory(t *testing.T) {
	runtime, codex := fixture(t)
	if err := os.MkdirAll(codex, 0700); err != nil {
		t.Fatal(err)
	}
	p, err := Plan(ActionInstall, runtime, codex)
	if err != nil {
		t.Fatal(err)
	}
	faultHook = func(phase string) error {
		if phase == "prepared" || phase == "rollback-control-detached" {
			return errors.New("injected interruption")
		}
		return nil
	}
	t.Cleanup(func() { faultHook = nil })
	if err = Execute(p); err == nil {
		t.Fatal("install interruption did not fire")
	}
	faultHook = nil
	faultHook = func(phase string) error {
		if phase == "rollback-control-detached" {
			return errors.New("injected cleanup interruption")
		}
		return nil
	}
	if err = Recover(codex); err == nil {
		t.Fatal("cleanup interruption did not fire")
	}
	faultHook = nil
	if err = Recover(codex); err != nil {
		t.Fatal(err)
	}
	if _, err = os.Stat(filepath.Join(codex, controlDir)); !os.IsNotExist(err) {
		t.Fatalf("empty control directory remains: %v", err)
	}
	if s, _ := Inspect("", codex); s.State != StateNotInstalled {
		t.Fatalf("state=%s", s.State)
	}
}

func TestFailedSwapRestorationPreservesDisplacedBytes(t *testing.T) {
	_, codex := installFixture(t)
	fd, err := unix.Open(filepath.Join(codex, controlDir), unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer unix.Close(fd)
	before, err := proofAt(fd, canonicalFile)
	if err != nil {
		t.Fatal(err)
	}
	afterBytes := []byte("replacement generation")
	after := fileProof{Exists: true, SHA256: digest(afterBytes), Bytes: afterBytes}
	transitionHook = func(point string) error {
		switch point {
		case "after-swap":
			if err := removeAt(fd, canonicalFile+".next"); err != nil {
				return err
			}
			return writeStageAt(fd, canonicalFile+".next", []byte("displaced foreign bytes"))
		case "before-swap-restore":
			return removeAt(fd, canonicalFile)
		}
		return nil
	}
	t.Cleanup(func() { transitionHook = nil })
	if err = applyTransition(fd, canonicalFile, before, after); err == nil {
		t.Fatal("concurrent displaced-entry replacement was accepted")
	}
	got, err := os.ReadFile(filepath.Join(codex, controlDir, canonicalFile+".next"))
	if err != nil || string(got) != "displaced foreign bytes" {
		t.Fatalf("displaced foreign bytes were not preserved: %q, %v", got, err)
	}
}
