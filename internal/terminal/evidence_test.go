package terminal

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/acoz-labs/my-friday/internal/plan"
	"github.com/acoz-labs/my-friday/internal/profile"
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
	write := func(name, transcript string) {
		canonical, _ := filepath.EvalSymlinks(root)
		transcript = strings.ReplaceAll(transcript, canonical, "<TEMP>")
		transcript = strings.ReplaceAll(transcript, root, "<TEMP>")
		if strings.Contains(transcript, "\x1b[") {
			t.Fatalf("ANSI control sequence in %s", name)
		}
		if err := os.WriteFile(filepath.Join(outDir, name), []byte(transcript), 0644); err != nil {
			t.Fatal(err)
		}
	}
	answers := func(parent, confirmation string) string {
		return "\nFridáy 🦇\nBoss\nKeep local work inspectable\n2\n\n" + parent + "\n" + confirmation + "\n"
	}
	var exitOut bytes.Buffer
	_, _ = Run(strings.NewReader(answers(filepath.Join(root, "exit"), "")), &exitOut, root)
	write("01-default-exit.txt", exitOut.String())

	createdRoot := filepath.Join(root, "success")
	var successOut bytes.Buffer
	_, _ = Run(strings.NewReader(answers(createdRoot, "Create")), &successOut, root)
	write("02-unicode-success.txt", successOut.String())

	var rerunOut bytes.Buffer
	_, _ = Run(strings.NewReader(answers(createdRoot, "Create")), &rerunOut, root)
	write("06-already-complete.txt", rerunOut.String())

	collisionRoot := filepath.Join(root, "collision")
	var collisionOut bytes.Buffer
	_, _ = Run(strings.NewReader("\nFriday\n\nHelp\n\n2\n"+collisionRoot+"\n"+filepath.Join(collisionRoot, "child")+"\nq\n"), &collisionOut, root)
	write("03-path-collision.txt", collisionOut.String())

	p, _ := profile.New("Friday", "", "Help", "balanced", "")
	rollbackPlan, _ := plan.Build(p, filepath.Join(root, "rollback/runtime"), filepath.Join(root, "rollback/memory"))
	var rollbackOut strings.Builder
	rollbackOut.WriteString("Preflight\n")
	_, rollbackErr := transaction.ExecuteWithProgress(rollbackPlan, func(phase string) error {
		if phase == "validated" {
			return fmt.Errorf("injected evidence failure")
		}
		return nil
	}, func(status string) { fmt.Fprintln(&rollbackOut, status) })
	fmt.Fprintf(&rollbackOut, "Step 7 of 7: Recovery required\nError: %v\nRollback restored the pre-run state.\n", rollbackErr)
	write("04-rollback.txt", rollbackOut.String())

	recoveryPlan, _ := plan.Build(p, filepath.Join(root, "recover/runtime"), filepath.Join(root, "recover/memory"))
	var recoveryOut strings.Builder
	_, _ = transaction.ExecuteWithProgress(recoveryPlan, func(phase string) error {
		if phase == "verified" {
			return fmt.Errorf("injected interruption")
		}
		return nil
	}, func(status string) { fmt.Fprintln(&recoveryOut, status) })
	journalPath, phase, ok := transaction.Interrupted(recoveryPlan)
	if !ok {
		t.Fatal("expected retained recovery journal")
	}
	fmt.Fprintf(&recoveryOut, "Interrupted creation found at phase %s.\nRecovery command: my-friday recover --transaction %s\n", phase, journalPath)
	result, err := transaction.RecoverWithResult(journalPath)
	fmt.Fprintf(&recoveryOut, "%s\nError: %v\n", result, err)
	write("05-partial-promotion-recovery.txt", recoveryOut.String())
}
