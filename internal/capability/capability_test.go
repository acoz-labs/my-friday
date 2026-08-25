package capability

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

func writePackage(t *testing.T, root, slug, version string) string {
	t.Helper()
	p := filepath.Join(root, "skills", slug)
	for _, d := range []string{"skill", "tests"} {
		if err := os.MkdirAll(filepath.Join(p, d), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	manifest := `{"contract_version":1,"slug":"` + slug + `","version":"` + version + `","display_name":"Daily brief","summary":"Prepare a concise daily brief","profile":"instruction-only","codex_compatibility":"skills-v1","triggers":["prepare my daily brief"],"inputs":["topics"],"outputs":["brief"],"success_behavior":"Return a concise brief","failure_behavior":"Explain missing input","scripts":"none","dependencies":"none","network":"none","credentials":"none","background":"none","durable_data":"none","publishing":"none"}` + "\n"
	skill := "---\nname: " + slug + "\ndescription: Prepare a concise daily brief when explicitly requested.\n---\n\n# Daily brief\n\nAsk for missing topics, then return a concise brief.\n"
	cases := `{"contract_version":1,"positive_triggers":["prepare my daily brief"],"non_triggers":["hello"],"examples":[{"input":"prepare my daily brief","output_contains":["brief"]}],"required_facts":["explicit invocation only"],"forbidden_effects":["scripts","dependencies","network","credentials","background","durable-data","publishing"]}` + "\n"
	for name, body := range map[string]string{"capability.json": manifest, "skill/SKILL.md": skill, "tests/cases.json": cases} {
		if err := os.WriteFile(filepath.Join(p, name), []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return p
}

func TestValidateStrictInstructionOnlyPackage(t *testing.T) {
	p := writePackage(t, t.TempDir(), "daily-brief", "1.0.0")
	pkg, err := Validate(p)
	if err != nil {
		t.Fatal(err)
	}
	if pkg.Manifest.Slug != "daily-brief" || len(pkg.Files) != 3 || pkg.SourceDigest == "" || pkg.ProjectionDigest == "" {
		t.Fatalf("unexpected package: %#v", pkg)
	}
	if pkg.Projection["agents/openai.yaml"] == nil || !strings.Contains(string(pkg.Projection["agents/openai.yaml"]), "allow_implicit_invocation: false") {
		t.Fatal("fixed explicit-invocation policy missing")
	}
}

func TestDeterministicCasesRejectContradictionsAndEmptyDeclarations(t *testing.T) {
	for _, tc := range []struct {
		name, cases string
	}{
		{"empty trigger", `{"contract_version":1,"positive_triggers":[""],"non_triggers":["hello"],"examples":[{"input":"prepare my daily brief","output_contains":["brief"]}],"required_facts":["explicit invocation only"],"forbidden_effects":["scripts","dependencies","network","credentials","background","durable-data","publishing"]}`},
		{"contradictory trigger", `{"contract_version":1,"positive_triggers":["prepare my daily brief"],"non_triggers":["prepare my daily brief"],"examples":[{"input":"prepare my daily brief","output_contains":["brief"]}],"required_facts":["explicit invocation only"],"forbidden_effects":["scripts","dependencies","network","credentials","background","durable-data","publishing"]}`},
		{"unaligned trigger", `{"contract_version":1,"positive_triggers":["unrelated"],"non_triggers":["hello"],"examples":[{"input":"unrelated","output_contains":["brief"]}],"required_facts":["explicit invocation only"],"forbidden_effects":["scripts","dependencies","network","credentials","background","durable-data","publishing"]}`},
		{"extra positive trigger", `{"contract_version":1,"positive_triggers":["prepare my daily brief","unrelated"],"non_triggers":["hello"],"examples":[{"input":"prepare my daily brief","output_contains":["brief"]}],"required_facts":["explicit invocation only"],"forbidden_effects":["scripts","dependencies","network","credentials","background","durable-data","publishing"]}`},
		{"uncovered example", `{"contract_version":1,"positive_triggers":["prepare my daily brief"],"non_triggers":["hello"],"examples":[{"input":"unrelated","output_contains":["brief"]}],"required_facts":["explicit invocation only"],"forbidden_effects":["scripts","dependencies","network","credentials","background","durable-data","publishing"]}`},
		{"empty output", `{"contract_version":1,"positive_triggers":["prepare my daily brief"],"non_triggers":["hello"],"examples":[{"input":"prepare my daily brief","output_contains":[""]}],"required_facts":["explicit invocation only"],"forbidden_effects":["scripts","dependencies","network","credentials","background","durable-data","publishing"]}`},
		{"missing fact", `{"contract_version":1,"positive_triggers":["prepare my daily brief"],"non_triggers":["hello"],"examples":[{"input":"prepare my daily brief","output_contains":["brief"]}],"required_facts":["not in instructions"],"forbidden_effects":["scripts","dependencies","network","credentials","background","durable-data","publishing"]}`},
		{"missing forbidden effect", `{"contract_version":1,"positive_triggers":["prepare my daily brief"],"non_triggers":["hello"],"examples":[{"input":"prepare my daily brief","output_contains":["brief"]}],"required_facts":["explicit invocation only"],"forbidden_effects":["network"]}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			p := writePackage(t, root, "daily-brief", "1.0.0")
			if err := os.WriteFile(filepath.Join(p, "tests", "cases.json"), []byte(tc.cases+"\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			pkg, err := Validate(p)
			if err != nil {
				t.Fatalf("suite should be structurally valid: %v", err)
			}
			if err = TestCases(pkg); err == nil {
				t.Fatal("contradictory deterministic suite passed")
			}
			instance := filepath.Join(root, "instance")
			if err = InitializeInstance(instance); err != nil {
				t.Fatal(err)
			}
			status, _ := Inspect(instance, p)
			if status.State != StateTestFailed {
				t.Fatalf("state=%s", status.State)
			}
			if _, err = Plan(instance, p, ActionInstall); err == nil {
				t.Fatal("test-failed suite became installable")
			}
		})
	}
}

func TestProjectionMutationRacesPreserveForeignBytes(t *testing.T) {
	for _, action := range []Action{ActionInstall, ActionUpgrade, ActionDisable, ActionRemove} {
		t.Run(string(action), func(t *testing.T) {
			root := t.TempDir()
			p := writePackage(t, root, "daily-brief", "1.0.0")
			instance := filepath.Join(root, "instance")
			if err := InitializeInstance(instance); err != nil {
				t.Fatal(err)
			}
			if action != ActionInstall {
				pl, err := Plan(instance, p, ActionInstall)
				if err != nil {
					t.Fatal(err)
				}
				if err = Execute(pl); err != nil {
					t.Fatal(err)
				}
				if action == ActionUpgrade {
					if err = os.WriteFile(filepath.Join(p, "capability.json"), []byte(strings.ReplaceAll(string(mustRead(t, filepath.Join(p, "capability.json"))), `"version":"1.0.0"`, `"version":"1.0.1"`)), 0o600); err != nil {
						t.Fatal(err)
					}
				}
			}
			pl, err := Plan(instance, p, action)
			if err != nil {
				t.Fatal(err)
			}
			foreign := filepath.Join(root, "foreign")
			mutationHook = func(phase string) error {
				if phase != "journal-written" {
					return nil
				}
				if action != ActionInstall {
					if err := os.RemoveAll(pl.Projection); err != nil {
						return err
					}
				}
				if err := os.MkdirAll(pl.Projection, 0o700); err != nil {
					return err
				}
				if err := os.WriteFile(filepath.Join(pl.Projection, "foreign"), []byte("keep"), 0o600); err != nil {
					return err
				}
				foreign = filepath.Join(pl.Projection, "foreign")
				return nil
			}
			defer func() { mutationHook = nil }()
			if err = Execute(pl); err == nil {
				t.Fatal("raced foreign projection accepted")
			}
			if b, readErr := os.ReadFile(foreign); readErr != nil || string(b) != "keep" {
				t.Fatalf("foreign bytes lost: %q %v", b, readErr)
			}
		})
	}
}

func TestQuarantineAndRecoveryRacesPreserveForeignBytes(t *testing.T) {
	root := t.TempDir()
	p := writePackage(t, root, "daily-brief", "1.0.0")
	instance := filepath.Join(root, "instance")
	if err := InitializeInstance(instance); err != nil {
		t.Fatal(err)
	}
	pl, err := Plan(instance, p, ActionInstall)
	if err != nil {
		t.Fatal(err)
	}
	if err = Execute(pl); err != nil {
		t.Fatal(err)
	}
	pl, err = Plan(instance, p, ActionDisable)
	if err != nil {
		t.Fatal(err)
	}
	var foreign string
	mutationHook = func(phase string) error {
		if phase == "disable-projection-quarantined" {
			foreign = pl.Projection + ".owned-" + pl.Receipt.ProjectionDigest[:16] + ".quarantine/foreign"
			return os.WriteFile(foreign, []byte("keep"), 0o600)
		}
		return nil
	}
	if err = Execute(pl); err == nil {
		t.Fatal("quarantine drift accepted")
	}
	mutationHook = nil
	if b, readErr := os.ReadFile(foreign); readErr != nil || string(b) != "keep" {
		t.Fatalf("quarantine foreign bytes lost: %q %v", b, readErr)
	}
	if err = Recover(instance, "daily-brief"); err == nil {
		t.Fatal("drifted quarantine recovery accepted")
	}
	if _, activeErr := os.Lstat(pl.Projection); !errors.Is(activeErr, os.ErrNotExist) {
		t.Fatalf("foreign quarantine exposed at active projection: %v", activeErr)
	}
	if b, readErr := os.ReadFile(foreign); readErr != nil || string(b) != "keep" {
		t.Fatalf("recovery lost drift evidence: %q %v", b, readErr)
	}

	// A separate interrupted install exercises the recovery post-check window.
	root2 := t.TempDir()
	p2 := writePackage(t, root2, "daily-brief", "1.0.0")
	instance2 := filepath.Join(root2, "instance")
	if err = InitializeInstance(instance2); err != nil {
		t.Fatal(err)
	}
	pl2, err := Plan(instance2, p2, ActionInstall)
	if err != nil {
		t.Fatal(err)
	}
	mutationHook = func(phase string) error {
		if phase == "projection-written" {
			return errors.New("stop")
		}
		return nil
	}
	if err = Execute(pl2); err == nil {
		t.Fatal("fault not injected")
	}
	foreign = filepath.Join(pl2.Projection, "foreign")
	mutationHook = func(phase string) error {
		if phase != "recovery-ownership-checked" {
			return nil
		}
		if err := os.RemoveAll(pl2.Projection); err != nil {
			return err
		}
		if err := os.Mkdir(pl2.Projection, 0o700); err != nil {
			return err
		}
		return os.WriteFile(foreign, []byte("keep"), 0o600)
	}
	defer func() { mutationHook = nil }()
	if err = Recover(instance2, "daily-brief"); err == nil {
		t.Fatal("recovery race accepted")
	}
	if b, readErr := os.ReadFile(foreign); readErr != nil || string(b) != "keep" {
		t.Fatalf("recovery foreign bytes lost: %q %v", b, readErr)
	}
}

func TestFinalDeletionBoundaryPreservesRacedReplacement(t *testing.T) {
	root := t.TempDir()
	p := writePackage(t, root, "daily-brief", "1.0.0")
	instance := filepath.Join(root, "instance")
	if err := InitializeInstance(instance); err != nil {
		t.Fatal(err)
	}
	pl, err := Plan(instance, p, ActionInstall)
	if err != nil {
		t.Fatal(err)
	}
	if err = Execute(pl); err != nil {
		t.Fatal(err)
	}
	pl, err = Plan(instance, p, ActionDisable)
	if err != nil {
		t.Fatal(err)
	}
	quarantine := pl.Projection + ".owned-" + pl.Receipt.ProjectionDigest[:16] + ".quarantine"
	foreign := filepath.Join(quarantine, "foreign")
	mutationHook = func(phase string) error {
		if phase != "final-deletion-boundary" {
			return nil
		}
		if err := os.RemoveAll(quarantine); err != nil {
			return err
		}
		if err := os.Mkdir(quarantine, 0o700); err != nil {
			return err
		}
		return os.WriteFile(foreign, []byte("keep"), 0o600)
	}
	defer func() { mutationHook = nil }()
	if err = Execute(pl); err == nil {
		t.Fatal("final deletion race accepted")
	}
	if b, readErr := os.ReadFile(foreign); readErr != nil || string(b) != "keep" {
		t.Fatalf("raced replacement deleted: %q %v", b, readErr)
	}
}

func TestRemoveQuarantineFaultIsRecoverable(t *testing.T) {
	root := t.TempDir()
	p := writePackage(t, root, "daily-brief", "1.0.0")
	instance := filepath.Join(root, "instance")
	if err := InitializeInstance(instance); err != nil {
		t.Fatal(err)
	}
	pl, err := Plan(instance, p, ActionInstall)
	if err != nil {
		t.Fatal(err)
	}
	if err = Execute(pl); err != nil {
		t.Fatal(err)
	}
	pl, err = Plan(instance, p, ActionRemove)
	if err != nil {
		t.Fatal(err)
	}
	mutationHook = func(phase string) error {
		if phase == "remove-control-quarantined" {
			return errors.New("stop")
		}
		return nil
	}
	if err = Execute(pl); err == nil {
		t.Fatal("fault not injected")
	}
	mutationHook = nil
	if err = Recover(instance, "daily-brief"); err != nil {
		t.Fatal(err)
	}
	status, err := Inspect(instance, p)
	if err != nil || status.State != StateInstalledHealthy {
		t.Fatalf("status=%s err=%v", status.State, err)
	}
}

func TestDescriptorCleanupFaultsRecoverIdempotently(t *testing.T) {
	for _, faultPhase := range []string{"descriptor-nested-entry-unlinked", "descriptor-root-entry-unlinked"} {
		for _, action := range []Action{ActionUpgrade, ActionDisable, ActionRemove} {
			t.Run(faultPhase+"/"+string(action), func(t *testing.T) {
				root := t.TempDir()
				p := writePackage(t, root, "daily-brief", "1.0.0")
				instance := filepath.Join(root, "instance")
				if err := InitializeInstance(instance); err != nil {
					t.Fatal(err)
				}
				pl, err := Plan(instance, p, ActionInstall)
				if err != nil {
					t.Fatal(err)
				}
				if err = Execute(pl); err != nil {
					t.Fatal(err)
				}
				if action == ActionUpgrade {
					body := strings.ReplaceAll(string(mustRead(t, filepath.Join(p, "capability.json"))), `"version":"1.0.0"`, `"version":"1.0.1"`)
					if err = os.WriteFile(filepath.Join(p, "capability.json"), []byte(body), 0o600); err != nil {
						t.Fatal(err)
					}
					skill := append(mustRead(t, filepath.Join(p, "skill", "SKILL.md")), []byte("\nUse the updated brief format.\n")...)
					if err = os.WriteFile(filepath.Join(p, "skill", "SKILL.md"), skill, 0o600); err != nil {
						t.Fatal(err)
					}
				}
				pl, err = Plan(instance, p, action)
				if err != nil {
					t.Fatal(err)
				}
				mutationHook = func(phase string) error {
					if phase == faultPhase {
						return errors.New("stop")
					}
					return nil
				}
				if err = Execute(pl); err == nil {
					t.Fatal("cleanup fault not injected")
				}
				mutationHook = nil
				if err = Recover(instance, "daily-brief"); err != nil {
					t.Fatal(err)
				}
				if err = Recover(instance, "daily-brief"); err != nil {
					t.Fatalf("idempotent recovery failed: %v", err)
				}
				for _, pattern := range []string{filepath.Join(instance, "workspace", ".agents", "skills", "daily-brief.*"), filepath.Join(instance, "capabilities", "daily-brief.*")} {
					matches, _ := filepath.Glob(pattern)
					if len(matches) != 0 {
						t.Fatalf("cleanup residue: %v", matches)
					}
				}
				status, inspectErr := Inspect(instance, p)
				want := StateInstalledHealthy
				if action == ActionUpgrade {
					want = StateSourceChanged
				}
				if action == ActionDisable {
					want = StateDisabled
				}
				if action == ActionRemove {
					want = StateInstalledHealthy
				}
				if inspectErr != nil || status.State != want {
					t.Fatalf("stable state=%s want=%s err=%v", status.State, want, inspectErr)
				}
			})
		}
	}
}

func TestRecoveryCleanupFaultAndRestoreCollisionRetainHandle(t *testing.T) {
	root := t.TempDir()
	p := writePackage(t, root, "daily-brief", "1.0.0")
	instance := filepath.Join(root, "instance")
	if err := InitializeInstance(instance); err != nil {
		t.Fatal(err)
	}
	pl, err := Plan(instance, p, ActionInstall)
	if err != nil {
		t.Fatal(err)
	}
	if err = Execute(pl); err != nil {
		t.Fatal(err)
	}
	pl, err = Plan(instance, p, ActionDisable)
	if err != nil {
		t.Fatal(err)
	}
	mutationHook = func(phase string) error {
		if phase == "disable-projection-quarantined" {
			return errors.New("stop")
		}
		return nil
	}
	if err = Execute(pl); err == nil {
		t.Fatal("quarantine transfer fault not injected")
	}
	mutationHook = func(phase string) error {
		if phase == "restore-handle-transferred" {
			return errors.New("stop")
		}
		return nil
	}
	if err = Recover(instance, "daily-brief"); err == nil {
		t.Fatal("restore transfer fault not injected")
	}
	mutationHook = nil
	restoring := pl.Projection + ".owned-" + pl.Package.ProjectionDigest[:16] + ".restoring"
	if _, err = os.Lstat(restoring); err != nil {
		t.Fatal("deterministic recovery handle missing")
	}
	if err = os.Mkdir(pl.Projection, 0o700); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(pl.Projection, "foreign"), []byte("keep"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err = Recover(instance, "daily-brief"); err == nil {
		t.Fatal("restore target collision accepted")
	}
	if _, err = os.Lstat(restoring); err != nil {
		t.Fatal("recovery handle lost on collision")
	}
	if err = os.RemoveAll(pl.Projection); err != nil {
		t.Fatal(err)
	}
	if err = Recover(instance, "daily-brief"); err != nil {
		t.Fatal(err)
	}
	if err = Recover(instance, "daily-brief"); err != nil {
		t.Fatal(err)
	}
	if _, err = os.Lstat(restoring); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("restoring residue remains")
	}
}

func TestRecoveryPartialCleanupFaultsResume(t *testing.T) {
	for _, faultPhase := range []string{"descriptor-nested-entry-unlinked", "descriptor-root-entry-unlinked"} {
		t.Run(faultPhase, func(t *testing.T) {
			root := t.TempDir()
			p := writePackage(t, root, "daily-brief", "1.0.0")
			instance := filepath.Join(root, "instance")
			if err := InitializeInstance(instance); err != nil {
				t.Fatal(err)
			}
			pl, err := Plan(instance, p, ActionInstall)
			if err != nil {
				t.Fatal(err)
			}
			mutationHook = func(phase string) error {
				if phase == "projection-written" {
					return errors.New("stop")
				}
				return nil
			}
			if err = Execute(pl); err == nil {
				t.Fatal("install fault not injected")
			}
			mutationHook = func(phase string) error {
				if phase == faultPhase {
					return errors.New("stop")
				}
				return nil
			}
			if err = Recover(instance, "daily-brief"); err == nil {
				t.Fatal("partial recovery cleanup fault not injected")
			}
			mutationHook = nil
			if err = Recover(instance, "daily-brief"); err != nil {
				t.Fatal(err)
			}
			if err = Recover(instance, "daily-brief"); err != nil {
				t.Fatal(err)
			}
			status, inspectErr := Inspect(instance, p)
			if inspectErr != nil || status.State != StateReady {
				t.Fatalf("state=%s err=%v", status.State, inspectErr)
			}
			for _, pattern := range []string{filepath.Join(instance, "workspace", ".agents", "skills", "daily-brief.*"), filepath.Join(instance, "capabilities", "daily-brief.*")} {
				matches, _ := filepath.Glob(pattern)
				if len(matches) != 0 {
					t.Fatalf("cleanup residue: %v", matches)
				}
			}
		})
	}
}

func TestCleanupManifestCommitGapsRecover(t *testing.T) {
	for _, tc := range []struct {
		name       string
		action     Action
		phase      string
		occurrence int
		want       State
	}{
		{"projection short write", ActionDisable, "cleanup-manifest-short-write", 1, StateDisabled},
		{"control short write", ActionRemove, "cleanup-manifest-short-write", 2, StateInstalledHealthy},
		{"projection before sync", ActionDisable, "cleanup-manifest-before-sync", 1, StateDisabled},
		{"control before sync", ActionRemove, "cleanup-manifest-before-sync", 2, StateInstalledHealthy},
		{"projection temp", ActionDisable, "cleanup-manifest-temp-synced", 1, StateDisabled},
		{"control temp", ActionRemove, "cleanup-manifest-temp-synced", 2, StateInstalledHealthy},
		{"projection root unlinked", ActionDisable, "cleanup-root-unlinked", 1, StateDisabled},
		{"control root unlinked", ActionRemove, "cleanup-root-unlinked", 2, StateReady},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			p := writePackage(t, root, "daily-brief", "1.0.0")
			instance := filepath.Join(root, "instance")
			if err := InitializeInstance(instance); err != nil {
				t.Fatal(err)
			}
			pl, err := Plan(instance, p, ActionInstall)
			if err != nil {
				t.Fatal(err)
			}
			if err = Execute(pl); err != nil {
				t.Fatal(err)
			}
			pl, err = Plan(instance, p, tc.action)
			if err != nil {
				t.Fatal(err)
			}
			seen := 0
			mutationHook = func(phase string) error {
				if phase == tc.phase {
					seen++
					if seen == tc.occurrence {
						return errors.New("stop")
					}
				}
				return nil
			}
			if err = Execute(pl); err == nil {
				t.Fatal("commit-gap fault not injected")
			}
			mutationHook = nil
			if err = Recover(instance, "daily-brief"); err != nil {
				t.Fatal(err)
			}
			if err = Recover(instance, "daily-brief"); err != nil {
				t.Fatal(err)
			}
			status, inspectErr := Inspect(instance, p)
			if inspectErr != nil || status.State != tc.want {
				t.Fatalf("state=%s want=%s err=%v", status.State, tc.want, inspectErr)
			}
			for _, pattern := range []string{filepath.Join(instance, "workspace", ".agents", "skills", "daily-brief.*"), filepath.Join(instance, "capabilities", "daily-brief.*")} {
				matches, _ := filepath.Glob(pattern)
				if len(matches) != 0 {
					t.Fatalf("cleanup residue: %v", matches)
				}
			}
		})
	}
}

func TestUnsafeCleanupTempIsPreserved(t *testing.T) {
	for _, kind := range []string{"symlink", "mode"} {
		t.Run(kind, func(t *testing.T) {
			instance := filepath.Join(t.TempDir(), "instance")
			if err := InitializeInstance(instance); err != nil {
				t.Fatal(err)
			}
			temp := filepath.Join(instance, "workspace", ".agents", "skills", "daily-brief.owned-deadbeef.cleanup.json.new")
			var err error
			if kind == "symlink" {
				err = os.Symlink("foreign", temp)
			} else {
				err = os.WriteFile(temp, []byte("{}"), 0o644)
			}
			if err != nil {
				t.Fatal(err)
			}
			if err = resumeCleanupManifests(instance, "daily-brief"); err == nil {
				t.Fatal("unsafe temp accepted")
			}
			if _, err = os.Lstat(temp); err != nil {
				t.Fatal("unsafe temp was removed")
			}
		})
	}
}

func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestValidateRejectsProhibitedAndUnsafeEntries(t *testing.T) {
	for _, tc := range []struct {
		name, path string
		body       []byte
		mode       os.FileMode
	}{
		{"scripts", "skill/scripts/run.sh", []byte("exit 0"), 0o600},
		{"user policy", "skill/agents/openai.yaml", []byte("policy: {}"), 0o600},
		{"executable", "skill/assets/run", []byte("x"), 0o700},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := writePackage(t, t.TempDir(), "daily-brief", "1.0.0")
			if err := os.MkdirAll(filepath.Dir(filepath.Join(p, tc.path)), 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(p, tc.path), tc.body, tc.mode); err != nil {
				t.Fatal(err)
			}
			if _, err := Validate(p); err == nil {
				t.Fatal("unsafe package accepted")
			}
		})
	}
	t.Run("symlink", func(t *testing.T) {
		p := writePackage(t, t.TempDir(), "daily-brief", "1.0.0")
		if err := os.Symlink("SKILL.md", filepath.Join(p, "skill", "linked")); err != nil {
			t.Fatal(err)
		}
		if _, err := Validate(p); err == nil {
			t.Fatal("symlink accepted")
		}
	})
	t.Run("hardlink", func(t *testing.T) {
		root := t.TempDir()
		instance := filepath.Join(root, "instance")
		if err := os.MkdirAll(filepath.Join(instance, "workspace"), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := InitializeInstance(instance); err != nil {
			t.Fatal(err)
		}
		control := filepath.Join(instance, "capabilities", "daily-brief")
		if err := os.Mkdir(control, 0o700); err != nil {
			t.Fatal(err)
		}
		foreign := filepath.Join(root, "foreign")
		if err := os.WriteFile(foreign, []byte("foreign"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.Link(foreign, filepath.Join(control, "transaction.json")); err != nil {
			t.Fatal(err)
		}
		if err := Recover(instance, "daily-brief"); err == nil {
			t.Fatal("hardlinked journal accepted")
		}
	})
}

func TestLifecyclePreservesSourceAndRefusesDrift(t *testing.T) {
	root := t.TempDir()
	p := writePackage(t, root, "daily-brief", "1.0.0")
	instance := filepath.Join(root, "instance")
	if err := os.MkdirAll(filepath.Join(instance, "workspace"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := InitializeInstance(instance); err != nil {
		t.Fatal(err)
	}
	plan, err := Plan(instance, p, ActionInstall)
	if err != nil {
		t.Fatal(err)
	}
	if err = Execute(plan); err != nil {
		t.Fatal(err)
	}
	status, err := Inspect(instance, p)
	if err != nil || status.State != StateInstalledHealthy {
		t.Fatalf("status=%#v err=%v", status, err)
	}
	if err = os.WriteFile(filepath.Join(p, "skill", "SKILL.md"), []byte("changed"), 0o600); err != nil {
		t.Fatal(err)
	}
	status, err = Inspect(instance, p)
	if err == nil || status.State != StateDraftInvalid {
		t.Fatalf("invalid source status=%#v err=%v", status, err)
	}
	// Restore source, then prove projection drift is never overwritten.
	p = writePackage(t, root, "daily-brief", "1.1.0")
	projection := filepath.Join(instance, "workspace", ".agents", "skills", "daily-brief", "SKILL.md")
	if err = os.WriteFile(projection, []byte("foreign drift"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err = Plan(instance, p, ActionUpgrade); err == nil {
		t.Fatal("managed drift accepted")
	}
	if err = os.Remove(projection); err != nil {
		t.Fatal(err)
	}
	if _, err = Plan(instance, p, ActionDisable); err == nil {
		t.Fatal("missing managed projection accepted")
	}
	if _, err = os.Stat(p); err != nil {
		t.Fatal("source was not preserved")
	}
}

func TestDisableEnableRemoveCompleteReversal(t *testing.T) {
	root := t.TempDir()
	p := writePackage(t, root, "daily-brief", "1.0.0")
	instance := filepath.Join(root, "instance")
	if err := os.MkdirAll(filepath.Join(instance, "workspace"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := InitializeInstance(instance); err != nil {
		t.Fatal(err)
	}
	for _, action := range []Action{ActionInstall, ActionDisable, ActionEnable, ActionRemove} {
		pl, err := Plan(instance, p, action)
		if err != nil {
			t.Fatalf("%s: %v", action, err)
		}
		if err = Execute(pl); err != nil {
			t.Fatalf("%s: %v", action, err)
		}
	}
	if _, err := os.Stat(filepath.Join(instance, "workspace", ".agents", "skills", "daily-brief")); !os.IsNotExist(err) {
		t.Fatal("projection remains")
	}
	if _, err := os.Stat(filepath.Join(instance, "capabilities", "daily-brief")); !os.IsNotExist(err) {
		t.Fatal("control state remains")
	}
	if _, err := os.Stat(p); err != nil {
		t.Fatal("source removed")
	}
}

func TestRecoverRestoresReceiptBoundProjection(t *testing.T) {
	root := t.TempDir()
	p := writePackage(t, root, "daily-brief", "1.0.0")
	instance := filepath.Join(root, "instance")
	if err := os.MkdirAll(filepath.Join(instance, "workspace"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := InitializeInstance(instance); err != nil {
		t.Fatal(err)
	}
	pl, err := Plan(instance, p, ActionInstall)
	if err != nil {
		t.Fatal(err)
	}
	if err = Execute(pl); err != nil {
		t.Fatal(err)
	}
	if err = os.RemoveAll(projectionPath(instance, "daily-brief")); err != nil {
		t.Fatal(err)
	}
	journal := filepath.Join(instance, "capabilities", "daily-brief", "transaction.json")
	r, err := readReceipt(instance, "daily-brief")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := json.MarshalIndent(transaction{1, ActionUpgrade, "daily-brief", r.SourceDigest, r.ProjectionDigest, receiptDigest(r), false}, "", "  ")
	body = append(body, '\n')
	if err = os.WriteFile(journal, body, 0o600); err != nil {
		t.Fatal(err)
	}
	if err = Recover(instance, "daily-brief"); err != nil {
		t.Fatal(err)
	}
	status, err := Inspect(instance, p)
	if err != nil || status.State != StateInstalledHealthy {
		t.Fatalf("status=%#v err=%v", status, err)
	}
}

func TestExecuteRefusesStaleSourcePlan(t *testing.T) {
	root := t.TempDir()
	p := writePackage(t, root, "daily-brief", "1.0.0")
	instance := filepath.Join(root, "instance")
	if err := os.MkdirAll(filepath.Join(instance, "workspace"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := InitializeInstance(instance); err != nil {
		t.Fatal(err)
	}
	pl, err := Plan(instance, p, ActionInstall)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(p, "skill", "SKILL.md"), []byte("changed after preview"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err = Execute(pl); err == nil || !strings.Contains(err.Error(), "stale capability plan") {
		t.Fatalf("stale plan accepted: %v", err)
	}
	if _, err = os.Stat(projectionPath(instance, "daily-brief")); !os.IsNotExist(err) {
		t.Fatal("stale plan mutated projection")
	}
}

func TestExecuteSerializesInstanceMutations(t *testing.T) {
	root := t.TempDir()
	p := writePackage(t, root, "daily-brief", "1.0.0")
	instance := filepath.Join(root, "instance")
	if err := os.MkdirAll(filepath.Join(instance, "workspace"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := InitializeInstance(instance); err != nil {
		t.Fatal(err)
	}
	pl, err := Plan(instance, p, ActionInstall)
	if err != nil {
		t.Fatal(err)
	}
	lock, err := os.Open(filepath.Join(instance, "capabilities"))
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Close()
	if err = syscall.Flock(int(lock.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		t.Fatal(err)
	}
	defer syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)
	if err = Execute(pl); err == nil || !strings.Contains(err.Error(), "transaction already active") {
		t.Fatalf("concurrent mutation accepted: %v", err)
	}
}

func TestFaultAfterMutationLeavesJournalAndRecoveryRestoresStableState(t *testing.T) {
	root := t.TempDir()
	p := writePackage(t, root, "daily-brief", "1.0.0")
	instance := filepath.Join(root, "instance")
	if err := os.MkdirAll(filepath.Join(instance, "workspace"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := InitializeInstance(instance); err != nil {
		t.Fatal(err)
	}
	pl, err := Plan(instance, p, ActionInstall)
	if err != nil {
		t.Fatal(err)
	}
	mutationHook = func(phase string) error {
		if phase == "projection-written" {
			return errors.New("injected stop")
		}
		return nil
	}
	defer func() { mutationHook = nil }()
	if err = Execute(pl); err == nil || !strings.Contains(err.Error(), "injected stop") {
		t.Fatalf("fault not injected: %v", err)
	}
	mutationHook = nil
	if _, err = Plan(instance, p, ActionInstall); err == nil || !strings.Contains(err.Error(), "recovery required") {
		t.Fatalf("journal did not block plan: %v", err)
	}
	if err = Recover(instance, "daily-brief"); err != nil {
		t.Fatal(err)
	}
	if _, err = os.Stat(projectionPath(instance, "daily-brief")); !os.IsNotExist(err) {
		t.Fatal("pre-receipt fault did not restore absence")
	}
	if _, err = os.Stat(filepath.Join(instance, "capabilities", "daily-brief")); !os.IsNotExist(err) {
		t.Fatal("pre-receipt control remains")
	}
}

func TestInstallRefusesForeignProjectionAndControl(t *testing.T) {
	for _, target := range []string{"projection", "control"} {
		for _, populated := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s-populated-%v", target, populated), func(t *testing.T) {
				root := t.TempDir()
				p := writePackage(t, root, "daily-brief", "1.0.0")
				instance := filepath.Join(root, "instance")
				if err := os.MkdirAll(filepath.Join(instance, "workspace"), 0o700); err != nil {
					t.Fatal(err)
				}
				if err := InitializeInstance(instance); err != nil {
					t.Fatal(err)
				}
				foreign := projectionPath(instance, "daily-brief")
				if target == "control" {
					foreign = filepath.Join(instance, "capabilities", "daily-brief")
				}
				if err := os.MkdirAll(foreign, 0o700); err != nil {
					t.Fatal(err)
				}
				if populated {
					if err := os.WriteFile(filepath.Join(foreign, "keep"), []byte("foreign"), 0o600); err != nil {
						t.Fatal(err)
					}
				}
				status, _ := Inspect(instance, p)
				if status.State != StateCollision {
					t.Fatalf("state=%s", status.State)
				}
				if _, err := Plan(instance, p, ActionInstall); err == nil {
					t.Fatal("foreign state accepted")
				}
				if _, err := os.Stat(foreign); err != nil {
					t.Fatal("foreign state removed")
				}
			})
		}
	}
}

func TestExecuteRefusesCollisionCreatedAfterPreview(t *testing.T) {
	root := t.TempDir()
	p := writePackage(t, root, "daily-brief", "1.0.0")
	instance := filepath.Join(root, "instance")
	if err := os.MkdirAll(filepath.Join(instance, "workspace"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := InitializeInstance(instance); err != nil {
		t.Fatal(err)
	}
	pl, err := Plan(instance, p, ActionInstall)
	if err != nil {
		t.Fatal(err)
	}
	foreign := projectionPath(instance, "daily-brief")
	if err = os.Mkdir(foreign, 0o700); err != nil {
		t.Fatal(err)
	}
	if err = Execute(pl); err == nil {
		t.Fatal("post-preview collision replaced")
	}
	if _, err = os.Stat(foreign); err != nil {
		t.Fatal("collision not preserved")
	}
}

func TestManagedMutationPreservesForeignControlSibling(t *testing.T) {
	root := t.TempDir()
	p := writePackage(t, root, "daily-brief", "1.0.0")
	instance := filepath.Join(root, "instance")
	if err := os.MkdirAll(filepath.Join(instance, "workspace"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := InitializeInstance(instance); err != nil {
		t.Fatal(err)
	}
	pl, err := Plan(instance, p, ActionInstall)
	if err != nil {
		t.Fatal(err)
	}
	if err = Execute(pl); err != nil {
		t.Fatal(err)
	}
	foreign := filepath.Join(instance, "capabilities", "daily-brief", "keep")
	if err = os.WriteFile(foreign, []byte("foreign"), 0o600); err != nil {
		t.Fatal(err)
	}
	status, _ := Inspect(instance, p)
	if status.State != StateCollision {
		t.Fatalf("state=%s", status.State)
	}
	if _, err = Plan(instance, p, ActionRemove); err == nil {
		t.Fatal("foreign control sibling accepted")
	}
	if b, err := os.ReadFile(foreign); err != nil || string(b) != "foreign" {
		t.Fatal("foreign control changed")
	}
}

func TestRecoverRejectsUnownedJournal(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(string) error
	}{
		{"malformed", func(p string) error { return os.WriteFile(p, []byte("{"), 0o600) }},
		{"mismatched", func(p string) error {
			return os.WriteFile(p, []byte(`{"contract_version":1,"action":"install","slug":"other","source_digest":"x","projection_digest":"y","prior_receipt_digest":"","created_control":true}`+"\n"), 0o600)
		}},
		{"wrong-mode", func(p string) error { return os.WriteFile(p, []byte(`{}`), 0o644) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			instance := filepath.Join(root, "instance")
			if err := os.MkdirAll(filepath.Join(instance, "workspace"), 0o700); err != nil {
				t.Fatal(err)
			}
			if err := InitializeInstance(instance); err != nil {
				t.Fatal(err)
			}
			control := filepath.Join(instance, "capabilities", "daily-brief")
			if err := os.Mkdir(control, 0o700); err != nil {
				t.Fatal(err)
			}
			journal := filepath.Join(control, "transaction.json")
			if err := tc.mutate(journal); err != nil {
				t.Fatal(err)
			}
			if err := Recover(instance, "daily-brief"); err == nil {
				t.Fatal("foreign journal authorized recovery")
			}
			if _, err := os.Stat(control); err != nil {
				t.Fatal("foreign control removed")
			}
		})
	}
	t.Run("symlink", func(t *testing.T) {
		root := t.TempDir()
		instance := filepath.Join(root, "instance")
		if err := os.MkdirAll(filepath.Join(instance, "workspace"), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := InitializeInstance(instance); err != nil {
			t.Fatal(err)
		}
		control := filepath.Join(instance, "capabilities", "daily-brief")
		if err := os.Mkdir(control, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink("foreign", filepath.Join(control, "transaction.json")); err != nil {
			t.Fatal(err)
		}
		if err := Recover(instance, "daily-brief"); err == nil {
			t.Fatal("linked journal accepted")
		}
	})
}

func TestObservableStatePrecedence(t *testing.T) {
	root := t.TempDir()
	instance := filepath.Join(root, "instance")
	if err := os.MkdirAll(filepath.Join(instance, "workspace"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := InitializeInstance(instance); err != nil {
		t.Fatal(err)
	}
	missing := filepath.Join(root, "skills", "missing")
	status, err := Inspect(instance, missing)
	if err != nil || status.State != StateAbsent {
		t.Fatalf("absent=%s %v", status.State, err)
	}
	p := writePackage(t, root, "daily-brief", "1.0.0")
	if err = os.Remove(filepath.Join(p, "tests", "cases.json")); err != nil {
		t.Fatal(err)
	}
	status, _ = Inspect(instance, p)
	if status.State != StateDraftValid {
		t.Fatalf("draft-valid=%s", status.State)
	}
	p = writePackage(t, root, "daily-brief", "1.0.0")
	if err = os.WriteFile(filepath.Join(p, "tests", "cases.json"), []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	status, _ = Inspect(instance, p)
	if status.State != StateTestFailed {
		t.Fatalf("test-failed=%s", status.State)
	}
	p = writePackage(t, root, "daily-brief", "1.0.0")
	status, err = Inspect(instance, p)
	if err != nil || status.State != StateReady {
		t.Fatalf("ready=%s %v", status.State, err)
	}
	pl, err := Plan(instance, p, ActionInstall)
	if err != nil {
		t.Fatal(err)
	}
	mutationHook = func(phase string) error {
		if phase == "journal-written" {
			return errors.New("stop")
		}
		return nil
	}
	if err = Execute(pl); err == nil {
		t.Fatal("fault missing")
	}
	mutationHook = nil
	status, err = Inspect(instance, p)
	if err != nil || status.State != StateInterrupted {
		t.Fatalf("interrupted=%s %v", status.State, err)
	}
	journal := filepath.Join(instance, "capabilities", "daily-brief", "transaction.json")
	if err = os.WriteFile(journal, []byte("foreign"), 0o600); err != nil {
		t.Fatal(err)
	}
	status, _ = Inspect(instance, p)
	if status.State != StateRecoveryRequired {
		t.Fatalf("recovery-required=%s", status.State)
	}
	if err = os.Remove(journal); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(instance, "capabilities", "daily-brief", "receipt.json"), []byte("foreign"), 0o600); err != nil {
		t.Fatal(err)
	}
	status, _ = Inspect(instance, p)
	if status.State != StateIncompatible {
		t.Fatalf("incompatible=%s", status.State)
	}
}
