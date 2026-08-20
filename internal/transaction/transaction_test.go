package transaction

import (
	"github.com/acoz-labs/my-friday/internal/plan"
	"github.com/acoz-labs/my-friday/internal/profile"
	"github.com/acoz-labs/my-friday/internal/repository"
	"os"
	"path/filepath"
	"testing"
)

func testPlan(t *testing.T) plan.CreationPlan {
	t.Helper()
	root := t.TempDir()
	p, _ := profile.New("Friday", "Boss", "Help", "concise", "")
	pl, err := plan.Build(p, filepath.Join(root, "runtime"), filepath.Join(root, "memory"))
	if err != nil {
		t.Fatal(err)
	}
	return pl
}
func TestExecuteAndExactRerun(t *testing.T) {
	pl := testPlan(t)
	result, err := Execute(pl, nil)
	if err != nil {
		t.Fatal(err)
	}
	if result != "Complete" {
		t.Fatal(result)
	}
	result, err = Execute(pl, nil)
	if err != nil || result != "Already complete" {
		t.Fatalf("%s %v", result, err)
	}
}
func TestFailureRollsBack(t *testing.T) {
	pl := testPlan(t)
	_, err := Execute(pl, func(phase string) error {
		if phase == "promoted-runtime" {
			return os.ErrInvalid
		}
		return nil
	})
	if err == nil {
		t.Fatal("expected failure")
	}
	for _, p := range []string{pl.Targets.Runtime, pl.Targets.Memory} {
		if _, e := os.Stat(p); !os.IsNotExist(e) {
			t.Fatalf("target retained: %s", p)
		}
	}
}

func TestFaultMatrixRollsBackEveryPublishedTransition(t *testing.T) {
	for _, phase := range []string{"journaled", "staged", "validated", "promoted-runtime", "promoted-memory"} {
		t.Run(phase, func(t *testing.T) {
			pl := testPlan(t)
			_, err := Execute(pl, func(got string) error {
				if got == phase {
					return os.ErrInvalid
				}
				return nil
			})
			if err == nil {
				t.Fatal("expected failure")
			}
			for _, target := range []string{pl.Targets.Runtime, pl.Targets.Memory} {
				if _, statErr := os.Stat(target); !os.IsNotExist(statErr) {
					t.Fatalf("target retained: %s", target)
				}
			}
		})
	}
}

func TestEmptyShellModeRestoredOnRollback(t *testing.T) {
	pl := testPlan(t)
	for _, p := range []string{pl.Targets.Runtime, pl.Targets.Memory} {
		if err := os.Mkdir(p, 0750); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(p, 0750); err != nil {
			t.Fatal(err)
		}
	}
	_, err := Execute(pl, func(phase string) error {
		if phase == "promoted-runtime" {
			return os.ErrInvalid
		}
		return nil
	})
	if err == nil {
		t.Fatal("expected failure")
	}
	for _, p := range []string{pl.Targets.Runtime, pl.Targets.Memory} {
		info, statErr := os.Stat(p)
		if statErr != nil {
			t.Fatal(statErr)
		}
		if info.Mode().Perm() != 0750 {
			t.Fatalf("mode %o", info.Mode().Perm())
		}
	}
}

func TestMissingParentsRemovedOnRollback(t *testing.T) {
	root := t.TempDir()
	p, _ := profile.New("Friday", "", "Help", "balanced", "")
	pl, _ := plan.Build(p, filepath.Join(root, "new", "pair", "runtime"), filepath.Join(root, "new", "pair", "memory"))
	_, err := Execute(pl, func(phase string) error {
		if phase == "staged" {
			return os.ErrInvalid
		}
		return nil
	})
	if err == nil {
		t.Fatal("expected failure")
	}
	if _, err = os.Stat(filepath.Join(root, "new")); !os.IsNotExist(err) {
		t.Fatal("transaction-owned parent retained")
	}
}

func TestRecoverCompletesPartialPromotion(t *testing.T) {
	pl := testPlan(t)
	jp, runtimeStage, memoryStage, _, err := derivedPaths(pl.PlanID, pl.Targets.Runtime, pl.Targets.Memory)
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.Create(pl, runtimeStage, memoryStage); err != nil {
		t.Fatal(err)
	}
	for _, stage := range []string{runtimeStage, memoryStage} {
		if err := os.WriteFile(filepath.Join(stage, ownershipMarker), []byte(pl.PlanID+"\n"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Rename(runtimeStage, pl.Targets.Runtime); err != nil {
		t.Fatal(err)
	}
	j := journal{PlanID: pl.PlanID, Phase: "promoted-runtime", Runtime: pl.Targets.Runtime, Memory: pl.Targets.Memory, RuntimeStage: runtimeStage, MemoryStage: memoryStage}
	j.Reservations = pl.ReservationPaths
	j.Expected = expectedFiles(pl)
	if err := createJournal(jp, j); err != nil {
		t.Fatal(err)
	}
	if err := Recover(jp); err != nil {
		t.Fatal(err)
	}
	if err := repository.ValidatePair(pl.Targets.Runtime, pl.Targets.Memory); err != nil {
		t.Fatal(err)
	}
	if err := Recover(jp); err != nil {
		t.Fatalf("second recovery must be safe: %v", err)
	}
}

func TestForeignReservationBlocksBeforeStaging(t *testing.T) {
	pl := testPlan(t)
	if err := os.WriteFile(pl.ReservationPaths[0], []byte("foreign\n"), 0600); err != nil {
		t.Fatal(err)
	}
	_, err := Execute(pl, nil)
	if err == nil {
		t.Fatal("expected reservation collision")
	}
	if _, statErr := os.Stat(pl.Targets.Runtime); !os.IsNotExist(statErr) {
		t.Fatal("target mutated")
	}
	if b, readErr := os.ReadFile(pl.ReservationPaths[0]); readErr != nil || string(b) != "foreign\n" {
		t.Fatal("foreign reservation was altered")
	}
}

func TestRollbackPreservesForeignChangesAndJournal(t *testing.T) {
	pl := testPlan(t)
	_, err := Execute(pl, func(phase string) error {
		if phase != "promoted-runtime" {
			return nil
		}
		if writeErr := os.WriteFile(filepath.Join(pl.Targets.Runtime, "foreign.txt"), []byte("keep me"), 0600); writeErr != nil {
			t.Fatal(writeErr)
		}
		return os.ErrInvalid
	})
	if err == nil {
		t.Fatal("expected recovery-required failure")
	}
	if b, readErr := os.ReadFile(filepath.Join(pl.Targets.Runtime, "foreign.txt")); readErr != nil || string(b) != "keep me" {
		t.Fatal("foreign target change was removed")
	}
	journalPath := filepath.Join(filepath.Dir(pl.Targets.Runtime), ".my-friday-"+pl.PlanID[:16]+".json")
	if _, statErr := os.Stat(journalPath); statErr != nil {
		t.Fatal("recovery journal was not retained")
	}
}

func TestRecoverRejectsJournalSuppliedSupportPath(t *testing.T) {
	pl := testPlan(t)
	foreign := filepath.Join(t.TempDir(), "foreign")
	if err := os.Mkdir(foreign, 0700); err != nil {
		t.Fatal(err)
	}
	j := journal{PlanID: pl.PlanID, Runtime: pl.Targets.Runtime, Memory: pl.Targets.Memory,
		RuntimeStage: foreign, MemoryStage: foreign + "-memory"}
	jp := filepath.Join(filepath.Dir(pl.Targets.Runtime), ".my-friday-"+pl.PlanID[:16]+".json")
	if err := createJournal(jp, j); err != nil {
		t.Fatal(err)
	}
	if err := Recover(jp); err == nil {
		t.Fatal("expected untrusted support path rejection")
	}
	if _, err := os.Stat(foreign); err != nil {
		t.Fatal("foreign path was altered")
	}
}

func TestTransitionRevalidationPreservesTargetCreatedAfterPreview(t *testing.T) {
	pl := testPlan(t)
	_, err := Execute(pl, func(phase string) error {
		if phase != "validated" {
			return nil
		}
		if mkdirErr := os.Mkdir(pl.Targets.Runtime, 0700); mkdirErr != nil {
			t.Fatal(mkdirErr)
		}
		if writeErr := os.WriteFile(filepath.Join(pl.Targets.Runtime, "foreign"), []byte("keep"), 0600); writeErr != nil {
			t.Fatal(writeErr)
		}
		return nil
	})
	if err == nil {
		t.Fatal("expected changed-target denial")
	}
	if b, readErr := os.ReadFile(filepath.Join(pl.Targets.Runtime, "foreign")); readErr != nil || string(b) != "keep" {
		t.Fatal("foreign target was altered")
	}
}

func TestInterruptedRecognizesOnlyMatchingPlanState(t *testing.T) {
	pl := testPlan(t)
	jp, runtimeStage, memoryStage, reservations, err := derivedPaths(pl.PlanID, pl.Targets.Runtime, pl.Targets.Memory)
	if err != nil {
		t.Fatal(err)
	}
	j := journal{PlanID: pl.PlanID, Phase: "validated", Runtime: pl.Targets.Runtime, Memory: pl.Targets.Memory, RuntimeStage: runtimeStage, MemoryStage: memoryStage, Reservations: reservations, Expected: expectedFiles(pl)}
	if err = createJournal(jp, j); err != nil {
		t.Fatal(err)
	}
	gotPath, gotPhase, ok := Interrupted(pl)
	if !ok || gotPath != jp || gotPhase != "validated" {
		t.Fatalf("path=%s phase=%s ok=%v", gotPath, gotPhase, ok)
	}
	j.RuntimeStage = filepath.Join(t.TempDir(), "foreign")
	if err = writeJournal(jp, j); err != nil {
		t.Fatal(err)
	}
	if _, _, ok = Interrupted(pl); ok {
		t.Fatal("accepted mismatched support state")
	}
}
