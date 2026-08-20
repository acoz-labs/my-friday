package terminal

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/acoz-labs/my-friday/internal/plan"
	"github.com/acoz-labs/my-friday/internal/transaction"
)

func TestGenerateTerminalEvidence(t *testing.T) {
	outDir := os.Getenv("MY_FRIDAY_EVIDENCE_DIR")
	if outDir == "" {
		t.Skip("set MY_FRIDAY_EVIDENCE_DIR to regenerate terminal evidence")
	}
	if err := os.MkdirAll(outDir, 0755); err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	snapshot := func() string { return snapshotTree(t, root) }
	write := func(name, transcript string) {
		canonical, _ := filepath.EvalSymlinks(root)
		transcript = strings.ReplaceAll(transcript, canonical, "<TEMP>")
		transcript = strings.ReplaceAll(transcript, root, "<TEMP>")
		if bytes.ContainsRune([]byte(transcript), '\x1b') {
			t.Fatalf("ESC byte in %s", name)
		}
		if err := os.WriteFile(filepath.Join(outDir, name), []byte(transcript), 0644); err != nil {
			t.Fatal(err)
		}
	}
	answers := func(parent, confirmation string) string {
		return "\nFridáy 🦇\nBoss\nKeep local work inspectable\n2\n\n" + parent + "\n" + confirmation + "\n"
	}
	var exitOut bytes.Buffer
	beforeExit := snapshot()
	exitResult, exitErr := Run(strings.NewReader(answers(filepath.Join(root, "exit"), "")), &exitOut, root)
	if exitErr != nil || exitResult != "Exit" || snapshot() != beforeExit {
		t.Fatalf("default exit result=%q err=%v\nbefore=%s\nafter=%s", exitResult, exitErr, beforeExit, snapshot())
	}
	write("01-default-exit.txt", exitOut.String())

	createdRoot := filepath.Join(root, "success")
	var successOut bytes.Buffer
	successResult, successErr := Run(strings.NewReader(answers(createdRoot, "Create")), &successOut, root)
	if successErr != nil || successResult != "Complete" {
		t.Fatalf("success result=%q err=%v", successResult, successErr)
	}
	write("02-unicode-success.txt", successOut.String())

	var rerunOut bytes.Buffer
	beforeRerun := snapshot()
	rerunResult, rerunErr := Run(strings.NewReader(answers(createdRoot, "Create")), &rerunOut, root)
	if rerunErr != nil || rerunResult != "Already complete" || snapshot() != beforeRerun {
		t.Fatalf("rerun result=%q err=%v\nbefore=%s\nafter=%s", rerunResult, rerunErr, beforeRerun, snapshot())
	}
	write("06-already-complete.txt", rerunOut.String())

	collisionRoot := filepath.Join(root, "collision")
	var collisionOut bytes.Buffer
	beforeCollision := snapshot()
	collisionResult, collisionErr := Run(strings.NewReader("\nFriday\n\nHelp\n\n2\n"+collisionRoot+"\n"+filepath.Join(collisionRoot, "child")+"\nq\n"), &collisionOut, root)
	if collisionErr != nil || collisionResult != "Exit" || snapshot() != beforeCollision {
		t.Fatalf("collision result=%q err=%v", collisionResult, collisionErr)
	}
	write("03-path-collision.txt", collisionOut.String())

	originalExecute := executeWithProgress
	defer func() { executeWithProgress = originalExecute }()
	runWithFault := func(parent, phase string) (string, string, error) {
		executeWithProgress = func(pl plan.CreationPlan, _ transaction.Fault, progress func(string)) (string, error) {
			return transaction.ExecuteWithProgress(pl, func(got string) error {
				if got == phase {
					return fmt.Errorf("injected evidence failure at %s", phase)
				}
				return nil
			}, progress)
		}
		var out bytes.Buffer
		result, err := Run(strings.NewReader(answers(parent, "Create")), &out, root)
		return out.String(), result, err
	}

	rollbackRoot := filepath.Join(root, "rollback")
	rollbackTranscript, rollbackResult, rollbackErr := runWithFault(rollbackRoot, "validated")
	if rollbackErr == nil || rollbackResult != "" {
		t.Fatalf("rollback result=%q err=%v", rollbackResult, rollbackErr)
	}
	for _, target := range []string{filepath.Join(rollbackRoot, "my-friday-runtime"), filepath.Join(rollbackRoot, "my-friday-memory")} {
		if _, err := os.Lstat(target); !os.IsNotExist(err) {
			t.Fatalf("rollback left target %s: %v", target, err)
		}
	}
	write("04-rollback.txt", rollbackTranscript)

	recoveryRoot := filepath.Join(root, "recover")
	recoveryTranscript, recoveryResult, recoveryErr := runWithFault(recoveryRoot, "verified")
	if recoveryErr == nil || recoveryResult != "" {
		t.Fatalf("interruption result=%q err=%v", recoveryResult, recoveryErr)
	}
	executeWithProgress = originalExecute
	journals, err := filepath.Glob(filepath.Join(root, ".my-friday-*.json"))
	if err != nil || len(journals) != 1 {
		t.Fatalf("recovery journals=%v err=%v", journals, err)
	}
	journals[0], err = filepath.EvalSymlinks(journals[0])
	if err != nil {
		t.Fatal(err)
	}
	var resumed bytes.Buffer
	fmt.Fprintf(&resumed, "$ my-friday recover --transaction %s\n", journals[0])
	result, err := Recover(journals[0], &resumed)
	if err != nil || result != "Recovered and verified repository pair" {
		t.Fatalf("recovery result=%q err=%v\n%s", result, err, resumed.String())
	}
	if err := ValidatePair(filepath.Join(recoveryRoot, "my-friday-runtime"), filepath.Join(recoveryRoot, "my-friday-memory")); err != nil {
		t.Fatal(err)
	}
	write("05-partial-promotion-recovery.txt", recoveryTranscript+"\n--- next invocation ---\n"+resumed.String())
}

func TestCommittedTerminalEvidenceHasNoControlBytesAndCoversScenarios(t *testing.T) {
	expected := map[string][]string{
		"01-default-exit.txt":               {"No changes made", "Runtime initial state: absent"},
		"02-unicode-success.txt":            {"Normalized identity: Fridáy 🦇", "Step 7 of 7: Result", "Complete"},
		"03-path-collision.txt":             {"must be distinct and non-nested", "No changes made"},
		"04-rollback.txt":                   {"injected evidence failure at validated", "Rollback restored the pre-run state"},
		"05-partial-promotion-recovery.txt": {"Recovery required", "--- next invocation ---", "Recovered and verified repository pair"},
		"06-already-complete.txt":           {"exact completed repository", "Already complete"},
	}
	for name, fragments := range expected {
		b, err := os.ReadFile(filepath.Join("..", "..", "docs", "evidence", "issue-3-terminal", name))
		if err != nil {
			t.Fatal(err)
		}
		if bytes.ContainsRune(b, '\x1b') {
			t.Fatalf("%s contains ESC", name)
		}
		for _, fragment := range fragments {
			if !bytes.Contains(b, []byte(fragment)) {
				t.Errorf("%s missing %q", name, fragment)
			}
		}
	}
}
