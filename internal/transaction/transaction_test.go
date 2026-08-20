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
	support := filepath.Join(filepath.Dir(pl.Targets.Runtime), ".my-friday-recovery-test")
	runtimeStage := support + "-runtime"
	memoryStage := support + "-memory"
	if err := repository.Create(pl, runtimeStage, memoryStage); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(runtimeStage, pl.Targets.Runtime); err != nil {
		t.Fatal(err)
	}
	jp := support + ".json"
	j := journal{PlanID: pl.PlanID, Phase: "promoted-runtime", Runtime: pl.Targets.Runtime, Memory: pl.Targets.Memory, RuntimeStage: runtimeStage, MemoryStage: memoryStage}
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
