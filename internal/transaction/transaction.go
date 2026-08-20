package transaction

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/acoz-labs/my-friday/internal/plan"
	"github.com/acoz-labs/my-friday/internal/repository"
)

type Fault func(string) error
type journal struct {
	PlanID, Phase, Runtime, Memory, RuntimeStage, MemoryStage string
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
	return filepath.Join(filepath.Dir(target), ".my-friday-reservation-"+hex.EncodeToString(h[:8]))
}

func Execute(pl plan.CreationPlan, fault Fault) (string, error) {
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
	base := filepath.Dir(pl.Targets.Runtime)
	var parents []string
	for _, parent := range pl.MissingParents {
		if _, err := os.Stat(parent); err == nil {
			continue
		} else if !os.IsNotExist(err) {
			return "", err
		}
		if err := os.Mkdir(parent, 0700); err != nil {
			return "", err
		}
		if err := os.Chmod(parent, 0700); err != nil {
			return "", err
		}
		parents = append(parents, parent)
	}
	support := filepath.Join(base, ".my-friday-"+pl.PlanID[:16])
	rExists, rMode := shellState(pl.Targets.Runtime)
	mExists, mMode := shellState(pl.Targets.Memory)
	j := journal{
		PlanID: pl.PlanID, Phase: "journaled", Runtime: pl.Targets.Runtime, Memory: pl.Targets.Memory,
		RuntimeStage:   filepath.Join(filepath.Dir(pl.Targets.Runtime), ".my-friday-"+pl.PlanID[:16]+"-runtime"),
		MemoryStage:    filepath.Join(filepath.Dir(pl.Targets.Memory), ".my-friday-"+pl.PlanID[:16]+"-memory"),
		RuntimeExisted: rExists, MemoryExisted: mExists, RuntimeMode: uint32(rMode), MemoryMode: uint32(mMode),
		CreatedParents: parents,
		Reservations:   pl.ReservationPaths,
		Expected:       expectedFiles(pl),
	}
	jp := support + ".json"
	if err := createJournal(jp, j); err != nil {
		removeParents(parents)
		return "", err
	}
	for _, reservation := range j.Reservations {
		if err := createReservation(reservation, pl.PlanID); err != nil {
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
	if err := repository.Create(pl, j.RuntimeStage, j.MemoryStage); err != nil {
		return "", rollback(jp, j, err)
	}
	for _, stage := range []string{j.RuntimeStage, j.MemoryStage} {
		if err := os.WriteFile(filepath.Join(stage, ownershipMarker), []byte(pl.PlanID+"\n"), 0600); err != nil {
			return "", rollback(jp, j, err)
		}
	}
	j.Phase = "staged"
	if err := writeJournal(jp, j); err != nil {
		return "", rollback(jp, j, err)
	}
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
	if err := fail("promoted-memory"); err != nil {
		return "", rollback(jp, j, err)
	}
	if err := repository.ValidateFreshPair(j.Runtime, j.Memory); err != nil {
		return "", rollback(jp, j, err)
	}
	if !repository.ExactTransactionBaseline(pl, j.Runtime, j.Memory) {
		return "", rollback(jp, j, fmt.Errorf("final repositories differ from the plan"))
	}
	for _, target := range []string{j.Runtime, j.Memory} {
		if err := removeOwnedMarker(target, j.PlanID); err != nil {
			return "", err
		}
	}
	removeReservations(j.Reservations, j.PlanID)
	if err := os.Remove(jp); err != nil {
		return "", err
	}
	return "Complete", nil
}

// Interrupted reports only transaction state whose identity and support paths
// exactly match the supplied immutable plan.
func Interrupted(pl plan.CreationPlan) (string, string, bool) {
	journalPath, runtimeStage, memoryStage, reservations, err := derivedPaths(pl.PlanID, pl.Targets.Runtime, pl.Targets.Memory)
	if err != nil {
		return "", "", false
	}
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
	return os.Rename(stage, target)
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
	for i, p := range []string{j.RuntimeStage, j.MemoryStage, j.Runtime, j.Memory} {
		role := []string{"runtime", "memory", "runtime", "memory"}[i]
		if err := removeOwnedTree(p, j.PlanID, j.Expected[role]); err != nil {
			retained = true
		} else {
			removed[p] = true
		}
	}
	if !retained {
		_ = os.Remove(jp)
		removeReservations(j.Reservations, j.PlanID)
	}
	if j.RuntimeExisted && removed[j.Runtime] {
		_ = os.MkdirAll(j.Runtime, 0700)
		_ = os.Chmod(j.Runtime, os.FileMode(j.RuntimeMode))
	}
	if j.MemoryExisted && removed[j.Memory] {
		_ = os.MkdirAll(j.Memory, 0700)
		_ = os.Chmod(j.Memory, os.FileMode(j.MemoryMode))
	}
	removeParents(j.CreatedParents)
	if retained {
		return fmt.Errorf("creation failed; recovery required and evidence retained: %w", cause)
	}
	return fmt.Errorf("creation failed and was rolled back: %w", cause)
}

func removeOwnedMarker(root, planID string) error {
	b, err := os.ReadFile(filepath.Join(root, ownershipMarker))
	if err != nil || string(b) != planID+"\n" {
		return fmt.Errorf("transaction ownership marker missing or changed: %s", root)
	}
	return os.Remove(filepath.Join(root, ownershipMarker))
}

func removeOwnedTree(root, planID string, expected map[string]string) error {
	if _, err := os.Lstat(root); os.IsNotExist(err) {
		return nil
	}
	marker, err := os.ReadFile(filepath.Join(root, ownershipMarker))
	if err != nil || string(marker) != planID+"\n" {
		return fmt.Errorf("transaction ownership marker missing or changed: %s", root)
	}
	seen := map[string]bool{}
	err = filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, _ := filepath.Rel(root, path)
		if rel == "." || rel == ownershipMarker {
			return nil
		}
		if rel == ".git" {
			return filepath.SkipDir
		}
		if entry.IsDir() {
			return nil
		}
		want, ok := expected[filepath.ToSlash(rel)]
		if !ok {
			return fmt.Errorf("foreign content prevents rollback: %s", path)
		}
		b, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		h := sha256.Sum256(b)
		if hex.EncodeToString(h[:]) != want {
			return fmt.Errorf("changed content prevents rollback: %s", path)
		}
		seen[filepath.ToSlash(rel)] = true
		return nil
	})
	if err != nil {
		return err
	}
	for path := range expected {
		if !seen[path] {
			return fmt.Errorf("missing content prevents rollback: %s", filepath.Join(root, path))
		}
	}
	if err := removeOwnedMarker(root, planID); err != nil {
		return err
	}
	return os.RemoveAll(root)
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

func removeParents(paths []string) {
	for i := len(paths) - 1; i >= 0; i-- {
		_ = os.Remove(paths[i])
	}
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
func removeReservations(paths []string, planID string) {
	for _, path := range paths {
		b, err := os.ReadFile(path)
		if err == nil && string(b) == planID+"\n" {
			_ = os.Remove(path)
		}
	}
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
	derivedJournal, runtimeStage, memoryStage, reservations, err := derivedPaths(j.PlanID, j.Runtime, j.Memory)
	if err != nil || filepath.Clean(journalPath) != derivedJournal || j.RuntimeStage != runtimeStage || j.MemoryStage != memoryStage || !slices.Equal(j.Reservations, reservations) {
		return fmt.Errorf("journal support paths do not match canonical transaction identity")
	}
	if repository.ValidatePair(j.Runtime, j.Memory) == nil {
		if err := removeOwnedTree(j.RuntimeStage, j.PlanID, j.Expected["runtime"]); err != nil {
			return err
		}
		if err := removeOwnedTree(j.MemoryStage, j.PlanID, j.Expected["memory"]); err != nil {
			return err
		}
		_ = removeOwnedMarker(j.Runtime, j.PlanID)
		_ = removeOwnedMarker(j.Memory, j.PlanID)
		removeReservations(j.Reservations, j.PlanID)
		return os.Remove(journalPath)
	}
	if repository.ValidatePair(j.Runtime, j.MemoryStage) == nil {
		if err = promote(j.MemoryStage, j.Memory); err != nil {
			return err
		}
		if err = repository.ValidatePair(j.Runtime, j.Memory); err != nil {
			return err
		}
		if err = removeOwnedTree(j.RuntimeStage, j.PlanID, j.Expected["runtime"]); err != nil {
			return err
		}
		removeReservations(j.Reservations, j.PlanID)
		_ = removeOwnedMarker(j.Runtime, j.PlanID)
		_ = removeOwnedMarker(j.Memory, j.PlanID)
		return os.Remove(journalPath)
	}
	if repository.ValidatePair(j.RuntimeStage, j.MemoryStage) == nil {
		if err = promote(j.RuntimeStage, j.Runtime); err != nil {
			return err
		}
		if err = promote(j.MemoryStage, j.Memory); err != nil {
			return err
		}
		if err = repository.ValidatePair(j.Runtime, j.Memory); err != nil {
			return err
		}
		removeReservations(j.Reservations, j.PlanID)
		return os.Remove(journalPath)
	}
	return fmt.Errorf("automatic recovery cannot prove a complete pair; inspect %s", journalPath)
}
