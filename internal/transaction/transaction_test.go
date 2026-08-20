package transaction

import (
	"github.com/acoz-labs/my-friday/internal/plan"
	"github.com/acoz-labs/my-friday/internal/profile"
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
