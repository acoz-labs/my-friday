package transaction

import (
	"github.com/acoz-labs/my-friday/internal/plan"
	"github.com/acoz-labs/my-friday/internal/profile"
	"github.com/acoz-labs/my-friday/internal/repository"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

func TestExactRerunRejectsEvolvedGitMetadata(t *testing.T) {
	pl := testPlan(t)
	if _, err := Execute(pl, nil); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("git", "-C", pl.Targets.Runtime, "config", "test.evolved", "yes")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git config: %v: %s", err, out)
	}
	if result, err := Execute(pl, nil); err == nil || result == "Already complete" {
		t.Fatalf("result=%q err=%v", result, err)
	}
}

func TestExactRerunRejectsChangedAllowedGitValueAndDirectory(t *testing.T) {
	for _, mutate := range []func(plan.CreationPlan) error{
		func(pl plan.CreationPlan) error {
			return exec.Command("git", "-C", pl.Targets.Runtime, "config", "core.bare", "true").Run()
		},
		func(pl plan.CreationPlan) error {
			return os.Mkdir(filepath.Join(pl.Targets.Runtime, ".git", "foreign-empty"), 0700)
		},
	} {
		pl := testPlan(t)
		if _, err := Execute(pl, nil); err != nil {
			t.Fatal(err)
		}
		if err := mutate(pl); err != nil {
			t.Fatal(err)
		}
		if result, err := Execute(pl, nil); err == nil || result == "Already complete" {
			t.Fatalf("result=%q err=%v", result, err)
		}
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
	for _, phase := range []string{"journaled", "runtime-files", "runtime-git", "memory-files", "memory-git", "staged", "validated", "promoted-runtime", "promoted-memory"} {
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

func TestCleanupFaultsRecoverIdempotently(t *testing.T) {
	for _, phase := range []string{"verified", "runtime-marker-removed", "memory-marker-removed", "reservations-removed"} {
		t.Run(phase, func(t *testing.T) {
			pl := testPlan(t)
			_, err := Execute(pl, func(got string) error {
				if got == phase {
					return os.ErrInvalid
				}
				return nil
			})
			if err == nil {
				t.Fatal("expected injected cleanup failure")
			}
			jp := pl.SupportPaths[0]
			if err := Recover(jp); err != nil {
				t.Fatalf("recover: %v", err)
			}
			if err := Recover(jp); err != nil {
				t.Fatalf("second recover: %v", err)
			}
			if err := repository.ValidatePair(pl.Targets.Runtime, pl.Targets.Memory); err != nil {
				t.Fatal(err)
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

func TestUntouchedEmptyShellsAreOriginalStateAtEarlyFailure(t *testing.T) {
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
		if phase == "journaled" {
			return os.ErrInvalid
		}
		return nil
	})
	if err == nil || strings.Contains(err.Error(), "recovery required") {
		t.Fatalf("expected clean rollback, got %v", err)
	}
	for _, p := range []string{pl.Targets.Runtime, pl.Targets.Memory} {
		info, statErr := os.Stat(p)
		if statErr != nil || info.Mode().Perm() != 0750 {
			t.Fatalf("shell %s not preserved: info=%v err=%v", p, info, statErr)
		}
	}
}

func TestRecoverRollsBackPrePromotionState(t *testing.T) {
	pl := testPlan(t)
	_, err := Execute(pl, func(phase string) error {
		if phase == "journaled" {
			return os.ErrInvalid
		}
		return nil
	})
	if err == nil {
		t.Fatal("expected injected failure")
	}
	// Recreate a crash-like journal with a marker-owned partial stage.
	j := journal{
		PlanID: pl.PlanID, Phase: "journaled", Runtime: pl.Targets.Runtime, Memory: pl.Targets.Memory,
		RuntimeStage: pl.SupportPaths[1], MemoryStage: pl.SupportPaths[2],
		RuntimeAnchor: existingAncestor(filepath.Dir(pl.Targets.Runtime)), MemoryAnchor: existingAncestor(filepath.Dir(pl.Targets.Memory)),
		Reservations: pl.ReservationPaths, Expected: expectedFiles(pl),
	}
	if err := createJournal(pl.SupportPaths[0], j); err != nil {
		t.Fatal(err)
	}
	for _, reservation := range j.Reservations {
		if err := createReservation(reservation, pl.PlanID); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(filepath.Join(j.RuntimeStage, filepath.Dir(ownershipMarker)), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(j.RuntimeStage, ownershipMarker), []byte(pl.PlanID+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(j.RuntimeStage, "partial"), []byte("owned"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := Recover(pl.SupportPaths[0]); err != nil {
		t.Fatal(err)
	}
	for _, p := range append([]string{pl.SupportPaths[0], j.RuntimeStage}, j.Reservations...) {
		if _, statErr := os.Lstat(p); !os.IsNotExist(statErr) {
			t.Fatalf("recovery residue retained: %s", p)
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
	j.RuntimeAnchor = existingAncestor(filepath.Dir(j.Runtime))
	j.MemoryAnchor = existingAncestor(filepath.Dir(j.Memory))
	j.Reservations = pl.ReservationPaths
	for _, reservation := range j.Reservations {
		if err := createReservation(reservation, pl.PlanID); err != nil {
			t.Fatal(err)
		}
	}
	runtimeExpected, _ := treeSnapshot(pl.Targets.Runtime)
	memoryExpected, _ := treeSnapshot(memoryStage)
	j.Expected = map[string]map[string]string{"runtime": runtimeExpected, "memory": memoryExpected}
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

func TestRollbackPreservesForeignGitChangesAndJournal(t *testing.T) {
	pl := testPlan(t)
	_, err := Execute(pl, func(phase string) error {
		if phase != "promoted-runtime" {
			return nil
		}
		cmd := exec.Command("git", "-C", pl.Targets.Runtime, "config", "test.foreign", "keep")
		if output, runErr := cmd.CombinedOutput(); runErr != nil {
			t.Fatalf("git config: %v: %s", runErr, output)
		}
		return os.ErrInvalid
	})
	if err == nil || !strings.Contains(err.Error(), "recovery required") {
		t.Fatalf("err=%v", err)
	}
	cmd := exec.Command("git", "-C", pl.Targets.Runtime, "config", "--get", "test.foreign")
	if output, runErr := cmd.Output(); runErr != nil || strings.TrimSpace(string(output)) != "keep" {
		t.Fatal("foreign Git state was removed")
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
	j.RuntimeAnchor = existingAncestor(filepath.Dir(j.Runtime))
	j.MemoryAnchor = existingAncestor(filepath.Dir(j.Memory))
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

func TestRecoverRejectsShortIdentityWithoutPanicking(t *testing.T) {
	pl := testPlan(t)
	jp := filepath.Join(filepath.Dir(pl.Targets.Runtime), ".my-friday-invalid.json")
	j := journal{PlanID: "short", Runtime: pl.Targets.Runtime, Memory: pl.Targets.Memory, RuntimeAnchor: filepath.Dir(pl.Targets.Runtime), MemoryAnchor: filepath.Dir(pl.Targets.Memory)}
	if err := createJournal(jp, j); err != nil {
		t.Fatal(err)
	}
	if err := Recover(jp); err == nil || !strings.Contains(err.Error(), "invalid journal transaction identity") {
		t.Fatalf("err=%v", err)
	}
}

func TestDeletionQuarantineRequiresJournalAuthorization(t *testing.T) {
	pl := testPlan(t)
	root := pl.Targets.Runtime
	if err := os.MkdirAll(filepath.Join(root, filepath.Dir(ownershipMarker)), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ownershipMarker), []byte(pl.PlanID+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	expected, err := treeSnapshot(root)
	if err != nil {
		t.Fatal(err)
	}
	quarantine := root + ".my-friday-delete-" + pl.PlanID[:16]
	if err := os.Mkdir(quarantine, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(quarantine, "foreign"), []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}
	j := journal{PlanID: pl.PlanID, DeletionPaths: map[string]string{}}
	if err := createJournal(pl.SupportPaths[0], j); err != nil {
		t.Fatal(err)
	}
	if err := removeOwnedTree(pl.SupportPaths[0], &j, root, expected); err == nil {
		t.Fatal("expected foreign quarantine collision")
	}
	if b, err := os.ReadFile(filepath.Join(quarantine, "foreign")); err != nil || string(b) != "keep" {
		t.Fatal("foreign quarantine was altered")
	}
}

func TestAuthorizedDeletionRetryRestoresOriginalEmptyShell(t *testing.T) {
	pl := testPlan(t)
	root := pl.Targets.Runtime
	if err := os.MkdirAll(filepath.Join(root, filepath.Dir(ownershipMarker)), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ownershipMarker), []byte(pl.PlanID+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	expected, err := treeSnapshot(root)
	if err != nil {
		t.Fatal(err)
	}
	quarantine := root + ".my-friday-delete-" + pl.PlanID[:16]
	if err := os.Rename(root, quarantine); err != nil {
		t.Fatal(err)
	}
	j := journal{
		PlanID: pl.PlanID, Runtime: root, Memory: pl.Targets.Memory,
		RuntimeExisted: true, RuntimeMode: 0750,
		RuntimeStage: pl.SupportPaths[1], MemoryStage: pl.SupportPaths[2],
		Reservations: []string{}, Expected: map[string]map[string]string{"runtime": expected, "memory": {}},
		DeletionPaths:    map[string]string{root: quarantine},
		DeletionExpected: map[string]map[string]string{root: expected},
	}
	if err := createJournal(pl.SupportPaths[0], j); err != nil {
		t.Fatal(err)
	}
	if err := rollback(pl.SupportPaths[0], j, os.ErrInvalid); err == nil || !strings.Contains(err.Error(), "rolled back") {
		t.Fatalf("rollback: %v", err)
	}
	info, err := os.Stat(root)
	if err != nil || info.Mode().Perm() != 0750 {
		t.Fatalf("shell not restored: info=%v err=%v", info, err)
	}
	if _, err := os.Lstat(quarantine); !os.IsNotExist(err) {
		t.Fatal("authorized quarantine retained")
	}
}

func TestAuthorizedDeletionRetryCompletesRenameBoundary(t *testing.T) {
	pl := testPlan(t)
	root := pl.Targets.Runtime
	if err := os.MkdirAll(filepath.Join(root, filepath.Dir(ownershipMarker)), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ownershipMarker), []byte(pl.PlanID+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	expected, err := treeSnapshot(root)
	if err != nil {
		t.Fatal(err)
	}
	quarantine := root + ".my-friday-delete-" + pl.PlanID[:16]
	j := journal{
		PlanID:           pl.PlanID,
		DeletionPaths:    map[string]string{root: quarantine},
		DeletionExpected: map[string]map[string]string{root: expected},
	}
	if err := createJournal(pl.SupportPaths[0], j); err != nil {
		t.Fatal(err)
	}
	if err := removeOwnedTree(pl.SupportPaths[0], &j, root, expected); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(root); !os.IsNotExist(err) {
		t.Fatal("root retained after authorized retry")
	}
	if _, err := os.Lstat(quarantine); !os.IsNotExist(err) {
		t.Fatal("quarantine retained after authorized retry")
	}
}

func TestAuthorizedDeletionRetryCompletesAfterQuarantineRemoval(t *testing.T) {
	pl := testPlan(t)
	root := pl.Targets.Runtime
	quarantine := root + ".my-friday-delete-" + pl.PlanID[:16]
	j := journal{
		PlanID: pl.PlanID, Runtime: root, Memory: pl.Targets.Memory,
		RuntimeExisted: true, RuntimeMode: 0750,
		RuntimeStage: pl.SupportPaths[1], MemoryStage: pl.SupportPaths[2],
		Reservations: []string{}, Expected: map[string]map[string]string{"runtime": {}, "memory": {}},
		DeletionPaths:    map[string]string{root: quarantine},
		DeletionExpected: map[string]map[string]string{root: {}},
	}
	if err := createJournal(pl.SupportPaths[0], j); err != nil {
		t.Fatal(err)
	}
	if err := rollback(pl.SupportPaths[0], j, os.ErrInvalid); err == nil || !strings.Contains(err.Error(), "rolled back") {
		t.Fatalf("rollback: %v", err)
	}
	info, err := os.Stat(root)
	if err != nil || info.Mode().Perm() != 0750 {
		t.Fatalf("shell not restored after completed deletion: info=%v err=%v", info, err)
	}
	if _, err := os.Lstat(pl.SupportPaths[0]); !os.IsNotExist(err) {
		t.Fatal("journal retained after completed deletion retry")
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
