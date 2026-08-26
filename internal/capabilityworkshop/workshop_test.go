package capabilityworkshop

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/acoz-labs/my-friday/internal/capability"
	"golang.org/x/sys/unix"
)

func TestMain(m *testing.M) {
	if os.Getenv("MY_FRIDAY_WORKSHOP_EXPECT_HELPER") == "1" {
		if len(os.Args) != 5 || os.Args[1] != "capability" || os.Args[2] != "workshop" {
			os.Exit(64)
		}
		instance, slug := os.Args[3], os.Args[4]
		if err := Run(instance, filepath.Join(instance, "runtime", "skills", slug), slug, os.Stdin, os.Stdout); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func createPackage(t *testing.T, instance, slug string, p proposal) string {
	t.Helper()
	source := filepath.Join(instance, "runtime", "skills", slug)
	files, err := render(p)
	if err != nil {
		t.Fatal(err)
	}
	for n, b := range files {
		path := filepath.Join(source, filepath.FromSlash(n))
		if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(path, b, 0600); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = capability.Validate(source); err != nil {
		t.Fatal(err)
	}
	return source
}

func validProposal(slug string) proposal {
	return proposal{Slug: slug, Version: "1.0.0", DisplayName: "Daily brief", Summary: "Prepare a concise daily brief", Purpose: "Prepare a daily brief from supplied topics", Success: "Return a concise brief containing DAILY_BRIEF_READY", Failure: "Explain missing topics", Triggers: []string{"prepare my daily brief"}, NonTriggers: []string{"chat casually"}, Inputs: []string{"topics"}, Outputs: []string{"brief"}, Facts: []string{"explicit invocation only"}, Examples: []example{{"prepare my daily brief", []string{"DAILY_BRIEF_READY"}}}}
}

func TestWorkshopCreatesValidatedInactiveSourceAfterExactConfirmation(t *testing.T) {
	root := t.TempDir()
	instance := filepath.Join(root, "alfred")
	if err := capability.InitializeInstance(instance); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(instance, "runtime", "skills"), 0700); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(instance, "runtime", "skills", "daily-brief")
	input := strings.Join([]string{
		"Daily brief", "Prepare a concise daily brief", "0.1.0",
		"Prepare a daily brief from supplied topics", "Return a concise brief", "Explain missing topics",
		"prepare my daily brief", "chat casually", "topics", "brief", "explicit invocation only",
		"prepare my daily brief for today", "brief", "", "Create source", "",
	}, "\n")
	var out bytes.Buffer
	if err := Run(instance, source, "daily-brief", strings.NewReader(input), &out); err != nil {
		t.Fatal(err)
	}
	status, err := capability.Inspect(instance, source)
	if err != nil {
		t.Fatal(err)
	}
	if status.State != capability.StateReady {
		t.Fatalf("state=%s", status.State)
	}
	for _, want := range []string{"1. capability.json", "2. skill/SKILL.md", "3. tests/cases.json", "Complete source diff:", "Source action: create", "Post-write state: ready", "Source written", "Next: my-friday capability install"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("missing %q", want)
		}
	}
}

func TestWorkshopDefaultAndWrongConfirmationDoNotWrite(t *testing.T) {
	for _, confirm := range []string{"", "Create", " create source "} {
		t.Run(confirm, func(t *testing.T) {
			root := t.TempDir()
			instance := filepath.Join(root, "alfred")
			if err := capability.InitializeInstance(instance); err != nil {
				t.Fatal(err)
			}
			if err := os.MkdirAll(filepath.Join(instance, "runtime", "skills"), 0700); err != nil {
				t.Fatal(err)
			}
			source := filepath.Join(instance, "runtime", "skills", "daily-brief")
			input := strings.Join([]string{"Daily brief", "Prepare a concise daily brief", "0.1.0", "Prepare a daily brief from supplied topics", "Return a concise brief", "Explain missing topics", "prepare my daily brief", "chat casually", "topics", "brief", "explicit invocation only", "prepare my daily brief", "brief", "", confirm, ""}, "\n")
			var out bytes.Buffer
			if err := Run(instance, source, "daily-brief", strings.NewReader(input), &out); err != nil {
				t.Fatal(err)
			}
			if _, err := os.Lstat(source); !os.IsNotExist(err) {
				t.Fatalf("source changed: %v", err)
			}
			if !strings.Contains(out.String(), "No changes made") {
				t.Fatal("missing no-change result")
			}
		})
	}
}

func TestWorkshopRecoversDigestProvenInterruptedCreateBeforeQuestions(t *testing.T) {
	root := t.TempDir()
	instance := filepath.Join(root, "alfred")
	if err := capability.InitializeInstance(instance); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(instance, "runtime", "skills"), 0700); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(instance, "runtime", "skills", "daily-brief")
	input := strings.Join([]string{"Daily brief", "Prepare a concise daily brief", "0.1.0", "Prepare a daily brief from supplied topics", "Return a concise brief", "Explain missing topics", "prepare my daily brief", "chat casually", "topics", "brief", "explicit invocation only", "prepare my daily brief", "brief", "", "Create source", ""}, "\n")
	mutationHook = func(phase string) error {
		if phase == "journal-written" {
			return os.ErrClosed
		}
		return nil
	}
	var first bytes.Buffer
	if err := Run(instance, source, "daily-brief", strings.NewReader(input), &first); err == nil {
		t.Fatal("fault not injected")
	}
	mutationHook = nil
	status, _ := capability.Inspect(instance, source)
	if status.State != capability.StateInterrupted {
		t.Fatalf("state=%s", status.State)
	}
	var recovered bytes.Buffer
	if err := Run(instance, source, "daily-brief", strings.NewReader(""), &recovered); err != nil {
		t.Fatal(err)
	}
	status, err := capability.Inspect(instance, source)
	if err != nil || status.State != capability.StateReady {
		t.Fatalf("state=%s err=%v", status.State, err)
	}
	if !strings.Contains(recovered.String(), "Source workshop recovered") {
		t.Fatal("missing recovery report")
	}
}

func TestEnhancementRetainsArbitraryBodyAndOptionalBytesAndMode(t *testing.T) {
	root := t.TempDir()
	instance := filepath.Join(root, "alfred")
	if err := capability.InitializeInstance(instance); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(instance, "runtime", "skills"), 0700); err != nil {
		t.Fatal(err)
	}
	source := createPackage(t, instance, "daily-brief", validProposal("daily-brief"))
	body := []byte("---\nname: daily-brief\ndescription: Prepare a concise daily brief\n---\n\n# Hand-written\n\nKeep **this** exact prose.\n")
	if err := os.WriteFile(filepath.Join(source, "skill", "SKILL.md"), body, 0600); err != nil {
		t.Fatal(err)
	}
	optional := filepath.Join(source, "skill", "references", "guide.txt")
	if err := os.MkdirAll(filepath.Dir(optional), 0750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(optional, []byte("opaque\x00bytes\n"), 0640); err != nil {
		t.Fatal(err)
	}
	input := "retain\n" + strings.Repeat("\n", 13) + "Update source\n"
	var out bytes.Buffer
	if err := Run(instance, source, "daily-brief", strings.NewReader(input), &out); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(source, "skill", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(got, []byte("# Hand-written\n\nKeep **this** exact prose.\n")) {
		t.Fatal("arbitrary instruction body changed")
	}
	opaque, err := os.ReadFile(optional)
	if err != nil || string(opaque) != "opaque\x00bytes\n" {
		t.Fatalf("optional bytes=%q err=%v", opaque, err)
	}
	info, _ := os.Stat(optional)
	if info.Mode().Perm() != 0640 {
		t.Fatalf("optional mode=%o", info.Mode().Perm())
	}
	dirInfo, _ := os.Stat(filepath.Dir(optional))
	if dirInfo.Mode().Perm() != 0750 {
		t.Fatalf("optional dir mode=%o", dirInfo.Mode().Perm())
	}
}

func TestCommitRefusesStalePreviewAndSharedLifecycleLock(t *testing.T) {
	root := t.TempDir()
	instance := filepath.Join(root, "alfred")
	if err := capability.InitializeInstance(instance); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(instance, "runtime", "skills"), 0700); err != nil {
		t.Fatal(err)
	}
	source := createPackage(t, instance, "daily-brief", validProposal("daily-brief"))
	before, err := capability.Inspect(instance, source)
	if err != nil {
		t.Fatal(err)
	}
	files, _ := render(validProposal("daily-brief"))
	manifest := filepath.Join(source, "capability.json")
	b, _ := os.ReadFile(manifest)
	if err = os.WriteFile(manifest, bytes.Replace(b, []byte("1.0.0"), []byte("1.0.1"), 1), 0600); err != nil {
		t.Fatal(err)
	}
	if err = commit(instance, source, "daily-brief", before, files, "update"); err == nil || !strings.Contains(err.Error(), "stale") {
		t.Fatalf("stale err=%v", err)
	}
	before, err = capability.Inspect(instance, source)
	if err != nil {
		t.Fatal(err)
	}
	lock, err := os.Open(filepath.Join(instance, "capabilities"))
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Close()
	if err = unix.Flock(int(lock.Fd()), unix.LOCK_EX|unix.LOCK_NB); err != nil {
		t.Fatal(err)
	}
	defer unix.Flock(int(lock.Fd()), unix.LOCK_UN)
	if err = commit(instance, source, "daily-brief", before, files, "update"); err == nil || !strings.Contains(err.Error(), "already in progress") {
		t.Fatalf("lock err=%v", err)
	}
}

func TestMalformedJournalPrecedesValidSourceAndWorkshopOnlyRecovery(t *testing.T) {
	root := t.TempDir()
	instance := filepath.Join(root, "alfred")
	if err := capability.InitializeInstance(instance); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(instance, "runtime", "skills"), 0700); err != nil {
		t.Fatal(err)
	}
	source := createPackage(t, instance, "daily-brief", validProposal("daily-brief"))
	jp := filepath.Join(instance, "capabilities", ".workshop-daily-brief.json")
	if err := os.WriteFile(jp, []byte("{}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	status, _ := capability.Inspect(instance, source)
	if status.State != capability.StateRecoveryRequired {
		t.Fatalf("state=%s", status.State)
	}
	if err := capability.Recover(instance, "daily-brief"); err == nil {
		t.Fatal("lifecycle recovery accepted source journal")
	}
	var out bytes.Buffer
	if err := Run(instance, source, "daily-brief", strings.NewReader(""), &out); err == nil {
		t.Fatal("malformed source journal recovered")
	}
}

func TestLinkedWorkshopJournalIsRecoveryRequired(t *testing.T) {
	root := t.TempDir()
	instance := filepath.Join(root, "alfred")
	if err := capability.InitializeInstance(instance); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(instance, "runtime", "skills"), 0700); err != nil {
		t.Fatal(err)
	}
	source := createPackage(t, instance, "daily-brief", validProposal("daily-brief"))
	jp := filepath.Join(instance, "capabilities", ".workshop-daily-brief.json")
	foreign := filepath.Join(root, "journal")
	if err := os.WriteFile(foreign, []byte("{}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Link(foreign, jp); err != nil {
		t.Fatal(err)
	}
	status, _ := capability.Inspect(instance, source)
	if status.State != capability.StateRecoveryRequired {
		t.Fatalf("state=%s", status.State)
	}
}

func TestExactOldTreeCleanupPreservesRacedForeignEntry(t *testing.T) {
	root := t.TempDir()
	instance := filepath.Join(root, "alfred")
	if err := capability.InitializeInstance(instance); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(instance, "runtime", "skills"), 0700); err != nil {
		t.Fatal(err)
	}
	source := createPackage(t, instance, "daily-brief", validProposal("daily-brief"))
	before, err := capability.Inspect(instance, source)
	if err != nil {
		t.Fatal(err)
	}
	p := validProposal("daily-brief")
	p.Version = "1.1.0"
	files, _ := render(p)
	old := filepath.Join(filepath.Dir(source), ".daily-brief.workshop-old")
	mutationHook = func(phase string) error {
		if phase == "source-promoted" {
			return os.WriteFile(filepath.Join(old, "foreign"), []byte("preserve"), 0600)
		}
		return nil
	}
	err = commit(instance, source, "daily-brief", before, files, "update")
	mutationHook = nil
	if err == nil {
		t.Fatal("raced cleanup accepted")
	}
	b, readErr := os.ReadFile(filepath.Join(old, "foreign"))
	if readErr != nil || string(b) != "preserve" {
		t.Fatalf("foreign bytes=%q err=%v", b, readErr)
	}
	if _, readErr = os.ReadFile(filepath.Join(old, "capability.json")); readErr != nil {
		t.Fatalf("cleanup partially removed recoverable old tree: %v", readErr)
	}
}

func TestRecoveryRefusesStagedAndOldModeDrift(t *testing.T) {
	for _, phase := range []string{"journal-written", "source-promoted"} {
		t.Run(phase, func(t *testing.T) {
			root := t.TempDir()
			instance := filepath.Join(root, "alfred")
			if err := capability.InitializeInstance(instance); err != nil {
				t.Fatal(err)
			}
			if err := os.MkdirAll(filepath.Join(instance, "runtime", "skills"), 0700); err != nil {
				t.Fatal(err)
			}
			source := createPackage(t, instance, "daily-brief", validProposal("daily-brief"))
			before, err := capability.Inspect(instance, source)
			if err != nil {
				t.Fatal(err)
			}
			p := validProposal("daily-brief")
			p.Version = "1.1.0"
			files, _ := render(p)
			mutationHook = func(got string) error {
				if got == phase {
					return errors.New("stop")
				}
				return nil
			}
			if err = commit(instance, source, "daily-brief", before, files, "update"); err == nil {
				t.Fatal("fault not injected")
			}
			mutationHook = nil
			defer func() { mutationHook = nil }()
			jp := filepath.Join(instance, "capabilities", ".workshop-daily-brief.json")
			j, err := readJournal(jp, "daily-brief")
			if err != nil {
				t.Fatal(err)
			}
			target := filepath.Join(filepath.Dir(source), j.StageRoot, "daily-brief", "skill")
			if phase == "source-promoted" {
				target = filepath.Join(filepath.Dir(source), ".daily-brief.workshop-old", "skill")
			}
			if err = os.Chmod(target, 0755); err != nil {
				t.Fatal(err)
			}
			var out bytes.Buffer
			if err = Run(instance, source, "daily-brief", strings.NewReader(""), &out); err == nil {
				t.Fatal("mode drift recovered")
			}
			if _, err = os.Stat(target); err != nil {
				t.Fatalf("drifted authority removed: %v", err)
			}
		})
	}
}

func TestCreatePromotionNeverReplacesRacedCollision(t *testing.T) {
	root := t.TempDir()
	instance := filepath.Join(root, "alfred")
	if err := capability.InitializeInstance(instance); err != nil {
		t.Fatal(err)
	}
	skills := filepath.Join(instance, "runtime", "skills")
	if err := os.MkdirAll(skills, 0700); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(skills, "daily-brief")
	before, _ := capability.Inspect(instance, source)
	files, _ := render(validProposal("daily-brief"))
	mutationHook = func(phase string) error {
		if phase == "journal-written" {
			if err := os.Mkdir(source, 0700); err != nil {
				return err
			}
			return os.WriteFile(filepath.Join(source, "foreign"), []byte("keep"), 0600)
		}
		return nil
	}
	err := commit(instance, source, "daily-brief", before, files, "create")
	mutationHook = nil
	if err == nil {
		t.Fatal("raced create collision accepted")
	}
	b, readErr := os.ReadFile(filepath.Join(source, "foreign"))
	if readErr != nil || string(b) != "keep" {
		t.Fatalf("foreign=%q err=%v", b, readErr)
	}
}

func TestPreJournalStageFaultsDoNotBlockRetry(t *testing.T) {
	for _, phase := range []string{"stage-created", "stage-written", "stage-synced", "stage-validated"} {
		t.Run(phase, func(t *testing.T) {
			root := t.TempDir()
			instance := filepath.Join(root, "alfred")
			if err := capability.InitializeInstance(instance); err != nil {
				t.Fatal(err)
			}
			if err := os.MkdirAll(filepath.Join(instance, "runtime", "skills"), 0700); err != nil {
				t.Fatal(err)
			}
			source := filepath.Join(instance, "runtime", "skills", "daily-brief")
			before, _ := capability.Inspect(instance, source)
			files, _ := render(validProposal("daily-brief"))
			mutationHook = func(got string) error {
				if got == phase {
					return errors.New("stop")
				}
				return nil
			}
			if err := commit(instance, source, "daily-brief", before, files, "create"); err == nil {
				t.Fatal("fault missing")
			}
			mutationHook = nil
			defer func() { mutationHook = nil }()
			if err := commit(instance, source, "daily-brief", before, files, "create"); err != nil {
				t.Fatalf("retry blocked: %v", err)
			}
		})
	}
}

func TestDescriptorRelativeStageRaceNeverWritesThroughSymlink(t *testing.T) {
	root := t.TempDir()
	instance := filepath.Join(root, "alfred")
	if err := capability.InitializeInstance(instance); err != nil {
		t.Fatal(err)
	}
	skills := filepath.Join(instance, "runtime", "skills")
	if err := os.MkdirAll(skills, 0700); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(skills, "daily-brief")
	before, _ := capability.Inspect(instance, source)
	files, _ := render(validProposal("daily-brief"))
	external := filepath.Join(root, "external")
	if err := os.Mkdir(external, 0700); err != nil {
		t.Fatal(err)
	}
	mutationHook = func(phase string) error {
		if phase != "stage-created" {
			return nil
		}
		entries, err := os.ReadDir(skills)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), ".daily-brief.workshop-new-") {
				if err = os.Rename(filepath.Join(skills, entry.Name()), filepath.Join(skills, entry.Name()+"-held")); err != nil {
					return err
				}
				return os.Symlink(external, filepath.Join(skills, entry.Name()))
			}
		}
		return errors.New("stage absent")
	}
	if err := commit(instance, source, "daily-brief", before, files, "create"); err == nil {
		t.Fatal("ancestor race accepted")
	}
	mutationHook = nil
	entries, err := os.ReadDir(external)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("external write: %v", entries)
	}
}

func TestInstalledStateSummariesNeverPromiseActivation(t *testing.T) {
	root := t.TempDir()
	instance := filepath.Join(root, "alfred")
	if err := capability.InitializeInstance(instance); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(instance, "runtime", "skills"), 0700); err != nil {
		t.Fatal(err)
	}
	source := createPackage(t, instance, "daily-brief", validProposal("daily-brief"))
	pl, err := capability.Plan(instance, source, capability.ActionInstall)
	if err != nil {
		t.Fatal(err)
	}
	if err = capability.Execute(pl); err != nil {
		t.Fatal(err)
	}
	files, _ := render(validProposal("daily-brief"))
	for _, tc := range []struct {
		state capability.State
		post  string
	}{{capability.StateInstalledHealthy, "source-changed"}, {capability.StateDisabled, "disabled"}} {
		status, err := capability.Inspect(instance, source)
		if err != nil {
			t.Fatal(err)
		}
		if tc.state == capability.StateDisabled && status.State != capability.StateDisabled {
			disable, err := capability.Plan(instance, source, capability.ActionDisable)
			if err != nil {
				t.Fatal(err)
			}
			if err = capability.Execute(disable); err != nil {
				t.Fatal(err)
			}
			status, err = capability.Inspect(instance, source)
			if err != nil {
				t.Fatal(err)
			}
		}
		var out bytes.Buffer
		if err = preview(&out, source, status, files, "update"); err != nil {
			t.Fatal(err)
		}
		text := out.String()
		if !strings.Contains(text, "Installed: unchanged") || !strings.Contains(text, "Post-write state: "+tc.post) {
			t.Fatalf("summary=%s", text)
		}
		if strings.Contains(text, "automatically") || strings.Contains(text, "activated") {
			t.Fatalf("activation promised: %s", text)
		}
	}
}

func TestInterruptedUpdateRecoversNewTreeAndReportsSourceChangedWhenInstalled(t *testing.T) {
	root := t.TempDir()
	instance := filepath.Join(root, "alfred")
	if err := capability.InitializeInstance(instance); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(instance, "runtime", "skills"), 0700); err != nil {
		t.Fatal(err)
	}
	source := createPackage(t, instance, "daily-brief", validProposal("daily-brief"))
	install, err := capability.Plan(instance, source, capability.ActionInstall)
	if err != nil {
		t.Fatal(err)
	}
	if err = capability.Execute(install); err != nil {
		t.Fatal(err)
	}
	projectionBefore := install.Package.ProjectionDigest
	p := validProposal("daily-brief")
	p.Version = "1.1.0"
	p.Success = "Return an updated concise brief containing DAILY_BRIEF_UPDATED"
	p.Examples[0].Output = []string{"DAILY_BRIEF_UPDATED"}
	files, _ := render(p)
	before, err := capability.Inspect(instance, source)
	if err != nil {
		t.Fatal(err)
	}
	mutationHook = func(phase string) error {
		if phase == "source-promoted" {
			return os.ErrClosed
		}
		return nil
	}
	if err = commit(instance, source, "daily-brief", before, files, "update"); err == nil {
		t.Fatal("fault not injected")
	}
	mutationHook = nil
	status, _ := capability.Inspect(instance, source)
	if status.State != capability.StateInterrupted {
		t.Fatalf("state=%s", status.State)
	}
	var out bytes.Buffer
	if err = Run(instance, source, "daily-brief", strings.NewReader(""), &out); err != nil {
		t.Fatal(err)
	}
	status, err = capability.Inspect(instance, source)
	if err != nil || status.State != capability.StateSourceChanged {
		t.Fatalf("state=%s err=%v", status.State, err)
	}
	if status.Receipt.ProjectionDigest != projectionBefore {
		t.Fatal("workshop activated changed source")
	}
	if !strings.Contains(out.String(), "State: source-changed") {
		t.Fatal("missing installed recovery state")
	}
}

func TestBoundsRepromptAndNavigationExitWithoutWrite(t *testing.T) {
	for _, input := range []string{"q\n", "", "b\nq\n", "r\nq\n", strings.Repeat("x", 201) + "\nDaily brief\nq\n"} {
		t.Run(fmt.Sprintf("%d", len(input)), func(t *testing.T) {
			root := t.TempDir()
			instance := filepath.Join(root, "alfred")
			if err := capability.InitializeInstance(instance); err != nil {
				t.Fatal(err)
			}
			if err := os.MkdirAll(filepath.Join(instance, "runtime", "skills"), 0700); err != nil {
				t.Fatal(err)
			}
			source := filepath.Join(instance, "runtime", "skills", "daily-brief")
			var out bytes.Buffer
			if err := Run(instance, source, "daily-brief", strings.NewReader(input), &out); err != nil {
				t.Fatal(err)
			}
			if _, err := os.Lstat(source); !os.IsNotExist(err) {
				t.Fatal("navigation changed source")
			}
			if !strings.Contains(out.String(), "No changes made") {
				t.Fatal("missing no-change result")
			}
		})
	}
}

func TestExplicitNoneRendersEmptyAnsweredInputsAndOutputs(t *testing.T) {
	p := validProposal("daily-brief")
	p.Inputs = nil
	p.Outputs = nil
	p.InputsAnswered = true
	p.OutputsAnswered = true
	files, err := render(p)
	if err != nil {
		t.Fatal(err)
	}
	var manifest capability.Manifest
	if err = json.Unmarshal(files["capability.json"], &manifest); err != nil {
		t.Fatal(err)
	}
	if len(manifest.Inputs) != 0 || len(manifest.Outputs) != 0 {
		t.Fatalf("none rendered as values: %#v %#v", manifest.Inputs, manifest.Outputs)
	}
	if values, err := list("none", true, 256, 16); err != nil || len(values) != 0 {
		t.Fatalf("none=%#v err=%v", values, err)
	}
}

func TestRetainedSkillBodySuffixPreservesTerminalNewlineExactly(t *testing.T) {
	for _, suffix := range [][]byte{[]byte("\n# Exact\n"), []byte("\n# Exact")} {
		t.Run(fmt.Sprintf("newline-%t", bytes.HasSuffix(suffix, []byte("\n"))), func(t *testing.T) {
			p := validProposal("daily-brief")
			p.Body = append([]byte(nil), suffix...)
			p.RetainBody = true
			files, err := render(p)
			if err != nil {
				t.Fatal(err)
			}
			marker := []byte("---\n")
			idx := bytes.Index(files["skill/SKILL.md"], marker)
			idx = bytes.Index(files["skill/SKILL.md"][idx+len(marker):], marker) + idx + 2*len(marker)
			if got := files["skill/SKILL.md"][idx:]; !bytes.Equal(got, suffix) {
				t.Fatalf("suffix=%q want=%q", got, suffix)
			}
		})
	}
}

func TestSummaryUsesYAMLSafeScalarAcceptedByPackageConsumer(t *testing.T) {
	p := validProposal("daily-brief")
	p.Summary = "Brief: #1 with \"quoted\" punctuation"
	files, err := render(p)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(files["skill/SKILL.md"], []byte(`description: "Brief: #1 with \"quoted\" punctuation"`)) {
		t.Fatalf("frontmatter=%s", files["skill/SKILL.md"])
	}
	root := t.TempDir()
	for name, body := range files {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(path, body, 0600); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = capability.ValidateForSlug(root, "daily-brief"); err != nil {
		t.Fatal(err)
	}
}

func TestCompleteDiffPreservesBlankLinesAndNoFinalNewlineGolden(t *testing.T) {
	var out bytes.Buffer
	writeDiffBytes(&out, '-', []byte("first\n\nlast"))
	writeDiffBytes(&out, '+', []byte("first\n\nlast\n"))
	want := "-first\n-\n-last\n\\ No newline at end of file\n+first\n+\n+last\n"
	if out.String() != want {
		t.Fatalf("diff:\n%s\nwant:\n%s", out.String(), want)
	}
}

func TestEnhancementSummariesPartialExampleRetentionAndMultiExamplePreservation(t *testing.T) {
	root := t.TempDir()
	instance := filepath.Join(root, "alfred")
	if err := capability.InitializeInstance(instance); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(instance, "runtime", "skills"), 0700); err != nil {
		t.Fatal(err)
	}
	p := validProposal("daily-brief")
	p.Success += " or TOMORROW"
	p.Examples = append(p.Examples, example{"prepare my daily brief tomorrow", []string{"TOMORROW"}})
	source := createPackage(t, instance, "daily-brief", p)
	input := "retain\n" + strings.Repeat("\n", 5) + strings.Repeat("\n", 5) + "\nTOMORROW\n\nUpdate source\n"
	var out bytes.Buffer
	if err := Run(instance, source, "daily-brief", strings.NewReader(input), &out); err != nil {
		t.Fatal(err)
	}
	pkg, err := capability.Validate(source)
	if err != nil {
		t.Fatal(err)
	}
	if len(pkg.Cases.Examples) != 2 || pkg.Cases.Examples[0].Input != p.Examples[0].Input || pkg.Cases.Examples[0].OutputContains[0] != "TOMORROW" || pkg.Cases.Examples[1].Input != p.Examples[1].Input {
		t.Fatalf("examples=%#v", pkg.Cases.Examples)
	}
	for _, want := range []string{"Current/default: Daily brief", "Current (1): 1=\"topics\"", "Current example 2:"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("missing %q", want)
		}
	}
	if strings.Contains(out.String(), "Purpose (1-1000") {
		t.Fatal("retained body requested unused purpose")
	}
}

func TestCheckedInEnhancementExpectRunsAgainstCandidateProcess(t *testing.T) {
	if _, err := os.Stat("/usr/bin/expect"); err != nil {
		t.Skip("native Expect unavailable")
	}
	root := t.TempDir()
	instance := filepath.Join(root, "alfred")
	if err := capability.InitializeInstance(instance); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(instance, "runtime", "skills"), 0700); err != nil {
		t.Fatal(err)
	}
	source := createPackage(t, instance, "daily-brief", validProposal("daily-brief"))
	transcript := filepath.Join(root, "enhance.transcript")
	script := filepath.Join("..", "..", "config", "acceptance", "capability-workshop.exp")
	cmd := exec.Command("/usr/bin/expect", script, "enhance", os.Args[0], instance, "daily-brief", transcript)
	cmd.Env = append(os.Environ(), "MY_FRIDAY_WORKSHOP_EXPECT_HELPER=1", "MY_FRIDAY_EXPECT_TIMEOUT=5")
	if output, err := cmd.CombinedOutput(); err != nil {
		transcriptBody, _ := os.ReadFile(transcript)
		t.Fatalf("native enhancement journey: %v\n%s\ntranscript:\n%s", err, output, transcriptBody)
	}
	pkg, err := capability.Validate(source)
	if err != nil {
		t.Fatal(err)
	}
	if pkg.Manifest.Version != "1.1.0" || pkg.Cases.Examples[0].Input != "prepare my daily brief" || pkg.Cases.Examples[0].OutputContains[0] != "DAILY_BRIEF_UPDATED" {
		t.Fatalf("enhancement=%#v %#v", pkg.Manifest, pkg.Cases.Examples)
	}
}
