package transaction

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/acoz-labs/my-friday/internal/plan"
	"github.com/acoz-labs/my-friday/internal/repository"
)

type Fault func(string) error
type journal struct {
	PlanID, Phase, Runtime, Memory, RuntimeStage, MemoryStage string
	RuntimeAnchor, MemoryAnchor                               string
	RuntimeExisted, MemoryExisted                             bool
	RuntimeMode, MemoryMode                                   uint32
	CreatedParents                                            []string
	Reservations                                              []string
	Expected                                                  map[string]map[string]string
}

const ownershipMarker = ".my-friday/creation-state.json"

func derivedPaths(planID, runtime, memory string) (string, string, string, []string, error) {
	if len(planID) < 16 || !filepath.IsAbs(runtime) || !filepath.IsAbs(memory) {
		return "", "", "", nil, fmt.Errorf("invalid transaction identity")
	}
	prefix := ".my-friday-" + planID[:16]
	jp := filepath.Join(filepath.Dir(runtime), prefix+".json")
	rs := filepath.Join(filepath.Dir(runtime), prefix+"-runtime")
	ms := filepath.Join(filepath.Dir(memory), prefix+"-memory")
	reservations := []string{reservationPath(runtime), reservationPath(memory)}
	return jp, rs, ms, reservations, nil
}

func reservationPath(target string) string {
	h := sha256.Sum256([]byte(target))
	return filepath.Join(existingAncestor(filepath.Dir(target)), ".my-friday-reservation-"+hex.EncodeToString(h[:8]))
}

func existingAncestor(path string) string {
	for {
		if _, err := os.Lstat(path); err == nil {
			return path
		}
		next := filepath.Dir(path)
		if next == path {
			return path
		}
		path = next
	}
}

func Execute(pl plan.CreationPlan, fault Fault) (string, error) {
	return ExecuteWithProgress(pl, fault, nil)
}

func ExecuteWithProgress(pl plan.CreationPlan, fault Fault, progress func(string)) (string, error) {
	emit := func(phase string) {
		if progress != nil {
			progress(phase)
		}
	}
	if repository.ExactBaseline(pl, pl.Targets.Runtime, pl.Targets.Memory) {
		return "Already complete", nil
	}
	for _, p := range []string{pl.Targets.Runtime, pl.Targets.Memory} {
		if info, err := os.Lstat(p); err == nil {
			if info.Mode()&os.ModeSymlink != 0 {
				return "", fmt.Errorf("target is a symlink: %s", p)
			}
			entries, _ := os.ReadDir(p)
			if !info.IsDir() || len(entries) > 0 {
				return "", fmt.Errorf("target is not empty: %s", p)
			}
		} else if !os.IsNotExist(err) {
			return "", err
		}
	}
	rExists, rMode := shellState(pl.Targets.Runtime)
	mExists, mMode := shellState(pl.Targets.Memory)
	j := journal{
		PlanID: pl.PlanID, Phase: "journaled", Runtime: pl.Targets.Runtime, Memory: pl.Targets.Memory,
		RuntimeStage:   filepath.Join(filepath.Dir(pl.Targets.Runtime), ".my-friday-"+pl.PlanID[:16]+"-runtime"),
		MemoryStage:    filepath.Join(filepath.Dir(pl.Targets.Memory), ".my-friday-"+pl.PlanID[:16]+"-memory"),
		RuntimeExisted: rExists, MemoryExisted: mExists, RuntimeMode: uint32(rMode), MemoryMode: uint32(mMode),
		CreatedParents: []string{},
		Reservations:   pl.ReservationPaths,
		Expected:       expectedFiles(pl),
	}
	j.RuntimeAnchor = existingAncestor(filepath.Dir(j.Runtime))
	j.MemoryAnchor = existingAncestor(filepath.Dir(j.Memory))
	jp := pl.SupportPaths[0]
	if err := createJournal(jp, j); err != nil {
		return "", err
	}
	emit("Journaled")
	for _, reservation := range j.Reservations {
		if err := createReservation(reservation, pl.PlanID); err != nil {
			return "", rollback(jp, j, err)
		}
	}
	emit("Reserved")
	for _, parent := range pl.MissingParents {
		if _, err := os.Stat(parent); err == nil {
			continue
		} else if !os.IsNotExist(err) {
			return "", rollback(jp, j, err)
		}
		j.CreatedParents = append(j.CreatedParents, parent)
		if err := writeJournal(jp, j); err != nil {
			return "", rollback(jp, j, err)
		}
		if err := os.Mkdir(parent, 0700); err != nil {
			return "", rollback(jp, j, err)
		}
		if err := os.Chmod(parent, 0700); err != nil {
			return "", rollback(jp, j, err)
		}
		if err := syncDir(filepath.Dir(parent)); err != nil {
			return "", rollback(jp, j, err)
		}
	}
	fail := func(phase string) error {
		if fault != nil {
			return fault(phase)
		}
		return nil
	}
	if err := fail("journaled"); err != nil {
		return "", rollback(jp, j, err)
	}
	if err := revalidate(j); err != nil {
		return "", rollback(jp, j, err)
	}
	for _, stage := range []string{j.RuntimeStage, j.MemoryStage} {
		if err := os.Mkdir(stage, 0700); err != nil {
			return "", rollback(jp, j, err)
		}
		if err := os.MkdirAll(filepath.Dir(filepath.Join(stage, ownershipMarker)), 0700); err != nil {
			return "", rollback(jp, j, err)
		}
		if err := os.WriteFile(filepath.Join(stage, ownershipMarker), []byte(pl.PlanID+"\n"), 0600); err != nil {
			return "", rollback(jp, j, err)
		}
		if err := syncDir(filepath.Join(stage, ".my-friday")); err != nil {
			return "", rollback(jp, j, err)
		}
	}
	j.Phase = "stages-owned"
	if err := writeJournal(jp, j); err != nil {
		return "", rollback(jp, j, err)
	}
	if err := repository.CreateWithCheckpoint(pl, j.RuntimeStage, j.MemoryStage, fail); err != nil {
		return "", rollback(jp, j, err)
	}
	for role, stage := range map[string]string{"runtime": j.RuntimeStage, "memory": j.MemoryStage} {
		snapshot, err := treeSnapshot(stage)
		if err != nil {
			return "", rollback(jp, j, err)
		}
		j.Expected[role] = snapshot
	}
	j.Phase = "staged"
	if err := writeJournal(jp, j); err != nil {
		return "", rollback(jp, j, err)
	}
	emit("Staged runtime")
	emit("Staged memory")
	if err := fail("staged"); err != nil {
		return "", rollback(jp, j, err)
	}
	if err := repository.ValidateFreshPair(j.RuntimeStage, j.MemoryStage); err != nil {
		return "", rollback(jp, j, err)
	}
	j.Phase = "validated"
	if err := writeJournal(jp, j); err != nil {
		return "", rollback(jp, j, err)
	}
	emit("Validated")
	if err := fail("validated"); err != nil {
		return "", rollback(jp, j, err)
	}
	if err := revalidate(j); err != nil {
		return "", rollback(jp, j, err)
	}
	if err := promote(j.RuntimeStage, j.Runtime); err != nil {
		return "", rollback(jp, j, err)
	}
	j.Phase = "promoted-runtime"
	if err := writeJournal(jp, j); err != nil {
		return "", rollback(jp, j, err)
	}
	emit("Promoted runtime")
	if err := fail("promoted-runtime"); err != nil {
		return "", rollback(jp, j, err)
	}
	if err := revalidateReservation(j.Reservations[1], j.PlanID); err != nil {
		return "", rollback(jp, j, err)
	}
	if err := promote(j.MemoryStage, j.Memory); err != nil {
		return "", rollback(jp, j, err)
	}
	j.Phase = "promoted-memory"
	if err := writeJournal(jp, j); err != nil {
		return "", rollback(jp, j, err)
	}
	emit("Promoted memory")
	if err := fail("promoted-memory"); err != nil {
		return "", rollback(jp, j, err)
	}
	if err := repository.ValidateFreshPair(j.Runtime, j.Memory); err != nil {
		return "", rollback(jp, j, err)
	}
	if !repository.ExactTransactionBaseline(pl, j.Runtime, j.Memory) {
		return "", rollback(jp, j, fmt.Errorf("final repositories differ from the plan"))
	}
	emit("Verified")
	j.Phase = "verified"
	if err := writeJournal(jp, j); err != nil {
		return "", err
	}
	if err := fail("verified"); err != nil {
		return "", err
	}
	for i, target := range []string{j.Runtime, j.Memory} {
		if err := removeOwnedMarker(target, j.PlanID); err != nil {
			return "", err
		}
		if err := fail([]string{"runtime-marker-removed", "memory-marker-removed"}[i]); err != nil {
			return "", err
		}
	}
	j.Phase = "markers-removed"
	if err := writeJournal(jp, j); err != nil {
		return "", err
	}
	if err := removeReservations(j.Reservations, j.PlanID); err != nil {
		return "", err
	}
	if err := fail("reservations-removed"); err != nil {
		return "", err
	}
	j.Phase = "reservations-removed"
	if err := writeJournal(jp, j); err != nil {
		return "", err
	}
	if err := os.Remove(jp); err != nil {
		return "", err
	}
	if err := syncDir(filepath.Dir(jp)); err != nil {
		return "", err
	}
	emit("Complete")
	return "Complete", nil
}

// Interrupted reports only transaction state whose identity and support paths
// exactly match the supplied immutable plan.
func Interrupted(pl plan.CreationPlan) (string, string, bool) {
	if len(pl.SupportPaths) < 3 {
		return "", "", false
	}
	journalPath, runtimeStage, memoryStage, reservations := pl.SupportPaths[0], pl.SupportPaths[1], pl.SupportPaths[2], pl.ReservationPaths
	b, err := os.ReadFile(journalPath)
	if err != nil {
		return "", "", false
	}
	var j journal
	if json.Unmarshal(b, &j) != nil || j.PlanID != pl.PlanID || j.Runtime != pl.Targets.Runtime || j.Memory != pl.Targets.Memory || j.RuntimeStage != runtimeStage || j.MemoryStage != memoryStage || !slices.Equal(j.Reservations, reservations) {
		return "", "", false
	}
	return journalPath, j.Phase, true
}

func revalidate(j journal) error {
	for _, reservation := range j.Reservations {
		if err := revalidateReservation(reservation, j.PlanID); err != nil {
			return err
		}
	}
	for _, target := range []string{j.Runtime, j.Memory} {
		canonical, err := canonicalCurrent(target)
		if err != nil || canonical != filepath.Clean(target) {
			return fmt.Errorf("target path identity changed: %s", target)
		}
		if info, statErr := os.Lstat(target); statErr == nil {
			if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
				return fmt.Errorf("target type changed: %s", target)
			}
			entries, readErr := os.ReadDir(target)
			if readErr != nil || len(entries) != 0 {
				return fmt.Errorf("target changed before promotion: %s", target)
			}
		} else if !os.IsNotExist(statErr) {
			return statErr
		}
	}
	return nil
}

func revalidateReservation(path, planID string) error {
	b, err := os.ReadFile(path)
	if err != nil || string(b) != planID+"\n" {
		return fmt.Errorf("reservation ownership changed: %s", path)
	}
	return nil
}

func canonicalCurrent(target string) (string, error) {
	ancestor := filepath.Clean(target)
	var suffix []string
	for {
		if _, err := os.Lstat(ancestor); err == nil {
			break
		} else if !os.IsNotExist(err) {
			return "", err
		}
		suffix = append(suffix, filepath.Base(ancestor))
		ancestor = filepath.Dir(ancestor)
	}
	resolved, err := filepath.EvalSymlinks(ancestor)
	if err != nil {
		return "", err
	}
	for i := len(suffix) - 1; i >= 0; i-- {
		resolved = filepath.Join(resolved, suffix[i])
	}
	return filepath.Clean(resolved), nil
}
func promote(stage, target string) error {
	if _, err := os.Stat(target); err == nil {
		entries, e := os.ReadDir(target)
		if e != nil || len(entries) > 0 {
			return fmt.Errorf("target changed before promotion: %s", target)
		}
		if e = os.Remove(target); e != nil {
			return e
		}
	}
	if err := os.Rename(stage, target); err != nil {
		return err
	}
	return syncDir(filepath.Dir(target))
}
func writeJournal(path string, j journal) error {
	b, _ := json.Marshal(j)
	tmp := path + ".tmp"
	f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	if _, err = f.Write(b); err != nil {
		_ = f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		_ = f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	if err = os.Rename(tmp, path); err != nil {
		return err
	}
	return syncDir(filepath.Dir(path))
}
func rollback(jp string, j journal, cause error) error {
	retained := false
	removed := map[string]bool{}
	for _, p := range []string{j.RuntimeStage, j.MemoryStage} {
		if err := removeMarkedStage(p, j.PlanID, nil); err != nil {
			retained = true
		}
	}
	for i, p := range []string{j.Runtime, j.Memory} {
		role := []string{"runtime", "memory"}[i]
		existed := []bool{j.RuntimeExisted, j.MemoryExisted}[i]
		changed, err := removeFinalTarget(p, j.PlanID, j.Expected[role], existed)
		if err != nil {
			retained = true
		} else if changed {
			removed[p] = true
		}
	}
	if j.RuntimeExisted && removed[j.Runtime] {
		if err := restoreShell(j.Runtime, os.FileMode(j.RuntimeMode)); err != nil {
			retained = true
		}
	}
	if j.MemoryExisted && removed[j.Memory] {
		if err := restoreShell(j.Memory, os.FileMode(j.MemoryMode)); err != nil {
			retained = true
		}
	}
	if err := removeParents(j.CreatedParents); err != nil {
		retained = true
	}
	if !retained {
		if err := removeReservations(j.Reservations, j.PlanID); err != nil {
			retained = true
		}
	}
	if !retained {
		if err := os.Remove(jp); err != nil && !os.IsNotExist(err) {
			retained = true
		}
		if !retained {
			retained = syncDir(filepath.Dir(jp)) != nil
		}
	}
	if retained {
		return fmt.Errorf("creation failed; recovery required and evidence retained: %w", cause)
	}
	return fmt.Errorf("creation failed and was rolled back: %w", cause)
}

func removeFinalTarget(root, planID string, expected map[string]string, existed bool) (bool, error) {
	if _, err := os.Lstat(root); os.IsNotExist(err) {
		return false, nil
	}
	if marker, err := os.ReadFile(filepath.Join(root, ownershipMarker)); err == nil && string(marker) == planID+"\n" {
		return true, removeOwnedTree(root, planID, expected)
	}
	if existed {
		info, err := os.Lstat(root)
		if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return false, fmt.Errorf("original empty shell changed: %s", root)
		}
		entries, err := os.ReadDir(root)
		if err == nil && len(entries) == 0 {
			return false, nil
		}
	}
	return false, fmt.Errorf("foreign or changed target prevents rollback: %s", root)
}

func restoreShell(path string, mode os.FileMode) error {
	if err := os.Mkdir(path, 0700); err != nil && !os.IsExist(err) {
		return err
	}
	if err := os.Chmod(path, mode); err != nil {
		return err
	}
	return syncDir(filepath.Dir(path))
}

func removeOwnedMarker(root, planID string) error {
	b, err := os.ReadFile(filepath.Join(root, ownershipMarker))
	if err != nil || string(b) != planID+"\n" {
		return fmt.Errorf("transaction ownership marker missing or changed: %s", root)
	}
	return os.Remove(filepath.Join(root, ownershipMarker))
}

func removeOwnedMarkerIfPresent(root, planID string) error {
	path := filepath.Join(root, ownershipMarker)
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil || string(b) != planID+"\n" {
		return fmt.Errorf("transaction ownership marker changed: %s", root)
	}
	return os.Remove(path)
}

func removeOwnedTree(root, planID string, expected map[string]string) error {
	deleting := root + ".my-friday-delete-" + planID[:16]
	if _, err := os.Lstat(deleting); err == nil {
		if _, rootErr := os.Lstat(root); !os.IsNotExist(rootErr) {
			return fmt.Errorf("rollback deletion identity is ambiguous: %s", root)
		}
		if err := os.RemoveAll(deleting); err != nil {
			return err
		}
		return syncDir(filepath.Dir(deleting))
	} else if !os.IsNotExist(err) {
		return err
	}
	if _, err := os.Lstat(root); os.IsNotExist(err) {
		return nil
	}
	marker, err := os.ReadFile(filepath.Join(root, ownershipMarker))
	if err != nil || string(marker) != planID+"\n" {
		return fmt.Errorf("transaction ownership marker missing or changed: %s", root)
	}
	actual, err := treeSnapshot(root)
	if err != nil {
		return err
	}
	if !mapsEqual(actual, expected) {
		return fmt.Errorf("foreign or changed content prevents rollback: %s", root)
	}
	// The atomic rename is the durable deletion-authority transition. A retry
	// may finish removing this derived path even if recursive deletion stopped
	// after removing the marker or part of the tree.
	if err := os.Rename(root, deleting); err != nil {
		return err
	}
	if err := syncDir(filepath.Dir(root)); err != nil {
		return err
	}
	if err := os.RemoveAll(deleting); err != nil {
		return err
	}
	return syncDir(filepath.Dir(deleting))
}

func removeMarkedStage(root, planID string, _ map[string]string) error {
	deleting := root + ".my-friday-delete-" + planID[:16]
	if _, err := os.Lstat(deleting); err == nil {
		if _, rootErr := os.Lstat(root); !os.IsNotExist(rootErr) {
			return fmt.Errorf("stage deletion identity is ambiguous: %s", root)
		}
		if err := os.RemoveAll(deleting); err != nil {
			return err
		}
		return syncDir(filepath.Dir(deleting))
	}
	if _, err := os.Lstat(root); os.IsNotExist(err) {
		return nil
	}
	marker, err := os.ReadFile(filepath.Join(root, ownershipMarker))
	if err != nil || string(marker) != planID+"\n" {
		return fmt.Errorf("stage ownership marker missing or changed: %s", root)
	}
	if err := os.Rename(root, deleting); err != nil {
		return err
	}
	if err := syncDir(filepath.Dir(root)); err != nil {
		return err
	}
	if err := os.RemoveAll(deleting); err != nil {
		return err
	}
	return syncDir(filepath.Dir(root))
}

func treeSnapshot(root string) (map[string]string, error) {
	out := map[string]string{}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, _ := filepath.Rel(root, path)
		rel = filepath.ToSlash(rel)
		if rel == "." || rel == ownershipMarker {
			return nil
		}
		info, err := os.Lstat(path)
		if err != nil {
			return err
		}
		h := sha256.New()
		fmt.Fprintf(h, "%s:%o:", info.Mode().Type().String(), info.Mode().Perm())
		if info.Mode().IsRegular() {
			b, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			_, _ = h.Write(b)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			target, err := os.Readlink(path)
			if err != nil {
				return err
			}
			_, _ = h.Write([]byte(target))
		}
		out[rel] = hex.EncodeToString(h.Sum(nil))
		return nil
	})
	return out, err
}

func mapsEqual(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

func expectedFiles(pl plan.CreationPlan) map[string]map[string]string {
	out := map[string]map[string]string{"runtime": {}, "memory": {}}
	for _, file := range pl.Files {
		out[file.Role][file.Path] = file.SHA256
	}
	return out
}

func shellState(path string) (bool, os.FileMode) {
	info, err := os.Stat(path)
	if err != nil {
		return false, 0
	}
	return true, info.Mode().Perm()
}

func removeParents(paths []string) error {
	for i := len(paths) - 1; i >= 0; i-- {
		if err := os.Remove(paths[i]); err != nil && !os.IsNotExist(err) {
			return err
		}
		if err := syncDir(filepath.Dir(paths[i])); err != nil {
			return err
		}
	}
	return nil
}
func createJournal(path string, j journal) error {
	b, _ := json.Marshal(j)
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return fmt.Errorf("another creation owns the plan journal: %w", err)
	}
	if _, err = f.Write(b); err != nil {
		_ = f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		_ = f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return syncDir(filepath.Dir(path))
}
func createReservation(path, planID string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return fmt.Errorf("target is reserved by another creation: %w", err)
	}
	if _, err = f.WriteString(planID + "\n"); err != nil {
		_ = f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		_ = f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return syncDir(filepath.Dir(path))
}

func syncDir(path string) error {
	dir, err := os.Open(path)
	if err != nil {
		return err
	}
	defer dir.Close()
	return dir.Sync()
}
func removeReservations(paths []string, planID string) error {
	for _, path := range paths {
		b, err := os.ReadFile(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return err
		}
		if string(b) != planID+"\n" {
			return fmt.Errorf("reservation ownership changed: %s", path)
		}
	}
	for _, path := range paths {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
		if err := syncDir(filepath.Dir(path)); err != nil {
			return err
		}
	}
	return nil
}

func finishRecovery(journalPath string, j journal) error {
	if j.Phase != "markers-removed" && j.Phase != "reservations-removed" {
		for _, target := range []string{j.Runtime, j.Memory} {
			if err := removeOwnedMarkerIfPresent(target, j.PlanID); err != nil {
				return err
			}
			if err := syncDir(filepath.Join(target, ".my-friday")); err != nil {
				return err
			}
		}
		j.Phase = "markers-removed"
		if err := writeJournal(journalPath, j); err != nil {
			return err
		}
	}
	if j.Phase != "reservations-removed" {
		if err := removeReservations(j.Reservations, j.PlanID); err != nil {
			return err
		}
		j.Phase = "reservations-removed"
		if err := writeJournal(journalPath, j); err != nil {
			return err
		}
	}
	if err := os.Remove(journalPath); err != nil {
		return err
	}
	return syncDir(filepath.Dir(journalPath))
}

func Recover(journalPath string) error {
	b, err := os.ReadFile(journalPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var j journal
	if err = json.Unmarshal(b, &j); err != nil {
		return err
	}
	if len(j.PlanID) < 16 || !filepath.IsAbs(j.Runtime) || !filepath.IsAbs(j.Memory) || !filepath.IsAbs(j.RuntimeAnchor) || !filepath.IsAbs(j.MemoryAnchor) {
		return fmt.Errorf("invalid journal transaction identity")
	}
	runtimeStage := filepath.Join(filepath.Dir(j.Runtime), ".my-friday-"+j.PlanID[:16]+"-runtime")
	memoryStage := filepath.Join(filepath.Dir(j.Memory), ".my-friday-"+j.PlanID[:16]+"-memory")
	derivedJournal := filepath.Join(j.RuntimeAnchor, ".my-friday-"+j.PlanID[:16]+".json")
	reservations := []string{reservationAt(j.RuntimeAnchor, j.Runtime), reservationAt(j.MemoryAnchor, j.Memory)}
	if filepath.Clean(journalPath) != derivedJournal || j.RuntimeStage != runtimeStage || j.MemoryStage != memoryStage || !slices.Equal(j.Reservations, reservations) {
		return fmt.Errorf("journal support paths do not match canonical transaction identity")
	}
	if j.Phase == "reservations-removed" {
		if err := os.Remove(journalPath); err != nil && !os.IsNotExist(err) {
			return err
		}
		return syncDir(filepath.Dir(journalPath))
	}
	if j.Phase == "markers-removed" {
		if err := repository.ValidateFreshPair(j.Runtime, j.Memory); err != nil {
			return err
		}
		for root, role := range map[string]string{j.Runtime: "runtime", j.Memory: "memory"} {
			actual, snapshotErr := treeSnapshot(root)
			if snapshotErr != nil || !mapsEqual(actual, j.Expected[role]) {
				return fmt.Errorf("completed repository changed during cleanup: %s", root)
			}
		}
		return finishRecovery(journalPath, j)
	}
	if j.Phase == "journaled" || j.Phase == "stages-owned" {
		err := rollback(journalPath, j, fmt.Errorf("interrupted before promotion"))
		if strings.Contains(err.Error(), "was rolled back") {
			return nil
		}
		return err
	}
	owned := func(root, role string) bool {
		marker, markerErr := os.ReadFile(filepath.Join(root, ownershipMarker))
		actual, e := treeSnapshot(root)
		return markerErr == nil && string(marker) == j.PlanID+"\n" && e == nil && mapsEqual(actual, j.Expected[role])
	}
	cleanupOwned := func(root, role string) bool {
		marker, markerErr := os.ReadFile(filepath.Join(root, ownershipMarker))
		if markerErr != nil && !os.IsNotExist(markerErr) {
			return false
		}
		if markerErr == nil && string(marker) != j.PlanID+"\n" {
			return false
		}
		actual, snapshotErr := treeSnapshot(root)
		return snapshotErr == nil && mapsEqual(actual, j.Expected[role])
	}
	if j.Phase == "verified" && repository.ValidateFreshPair(j.Runtime, j.Memory) == nil && cleanupOwned(j.Runtime, "runtime") && cleanupOwned(j.Memory, "memory") {
		return finishRecovery(journalPath, j)
	}
	if repository.ValidateFreshPair(j.Runtime, j.Memory) == nil && owned(j.Runtime, "runtime") && owned(j.Memory, "memory") {
		if err := removeOwnedTree(j.RuntimeStage, j.PlanID, j.Expected["runtime"]); err != nil {
			return err
		}
		if err := removeOwnedTree(j.MemoryStage, j.PlanID, j.Expected["memory"]); err != nil {
			return err
		}
		return finishRecovery(journalPath, j)
	}
	if repository.ValidateFreshPair(j.Runtime, j.MemoryStage) == nil && owned(j.Runtime, "runtime") && owned(j.MemoryStage, "memory") {
		if err = promote(j.MemoryStage, j.Memory); err != nil {
			return err
		}
		if err = repository.ValidateFreshPair(j.Runtime, j.Memory); err != nil || !owned(j.Runtime, "runtime") || !owned(j.Memory, "memory") {
			if err == nil {
				err = fmt.Errorf("promoted pair no longer matches transaction snapshot")
			}
			return err
		}
		if err = removeOwnedTree(j.RuntimeStage, j.PlanID, j.Expected["runtime"]); err != nil {
			return err
		}
		return finishRecovery(journalPath, j)
	}
	if repository.ValidateFreshPair(j.RuntimeStage, j.MemoryStage) == nil && owned(j.RuntimeStage, "runtime") && owned(j.MemoryStage, "memory") {
		if err = promote(j.RuntimeStage, j.Runtime); err != nil {
			return err
		}
		if err = promote(j.MemoryStage, j.Memory); err != nil {
			return err
		}
		if err = repository.ValidateFreshPair(j.Runtime, j.Memory); err != nil || !owned(j.Runtime, "runtime") || !owned(j.Memory, "memory") {
			if err == nil {
				err = fmt.Errorf("promoted pair no longer matches transaction snapshot")
			}
			return err
		}
		return finishRecovery(journalPath, j)
	}
	return fmt.Errorf("automatic recovery cannot prove a complete pair; inspect %s", journalPath)
}
func reservationAt(anchor, target string) string {
	h := sha256.Sum256([]byte(target))
	return filepath.Join(anchor, ".my-friday-reservation-"+hex.EncodeToString(h[:8]))
}
