package transaction

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/acoz-labs/my-friday/internal/plan"
	"github.com/acoz-labs/my-friday/internal/repository"
)

type Fault func(string) error
type journal struct {
	PlanID, Phase, Runtime, Memory, RuntimeStage, MemoryStage string
	RuntimeExisted, MemoryExisted                             bool
	RuntimeMode, MemoryMode                                   uint32
	CreatedParents                                            []string
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
	}
	jp := support + ".json"
	if err := createJournal(jp, j); err != nil {
		removeParents(parents)
		return "", err
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
	if err := repository.Create(pl, j.RuntimeStage, j.MemoryStage); err != nil {
		return "", rollback(jp, j, err)
	}
	j.Phase = "staged"
	_ = writeJournal(jp, j)
	if err := fail("staged"); err != nil {
		return "", rollback(jp, j, err)
	}
	if err := repository.ValidatePair(j.RuntimeStage, j.MemoryStage); err != nil {
		return "", rollback(jp, j, err)
	}
	j.Phase = "validated"
	_ = writeJournal(jp, j)
	if err := fail("validated"); err != nil {
		return "", rollback(jp, j, err)
	}
	if err := promote(j.RuntimeStage, j.Runtime); err != nil {
		return "", rollback(jp, j, err)
	}
	j.Phase = "promoted-runtime"
	_ = writeJournal(jp, j)
	if err := fail("promoted-runtime"); err != nil {
		return "", rollback(jp, j, err)
	}
	if err := promote(j.MemoryStage, j.Memory); err != nil {
		return "", rollback(jp, j, err)
	}
	j.Phase = "promoted-memory"
	_ = writeJournal(jp, j)
	if err := fail("promoted-memory"); err != nil {
		return "", rollback(jp, j, err)
	}
	if err := repository.ValidatePair(j.Runtime, j.Memory); err != nil {
		return "", rollback(jp, j, err)
	}
	os.Remove(jp)
	return "Complete", nil
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
	if err := os.WriteFile(tmp, b, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
func rollback(jp string, j journal, cause error) error {
	for _, p := range []string{j.RuntimeStage, j.MemoryStage, j.Runtime, j.Memory} {
		_ = os.RemoveAll(p)
	}
	_ = os.Remove(jp)
	if j.RuntimeExisted {
		_ = os.MkdirAll(j.Runtime, 0700)
		_ = os.Chmod(j.Runtime, os.FileMode(j.RuntimeMode))
	}
	if j.MemoryExisted {
		_ = os.MkdirAll(j.Memory, 0700)
		_ = os.Chmod(j.Memory, os.FileMode(j.MemoryMode))
	}
	removeParents(j.CreatedParents)
	return fmt.Errorf("creation failed and was rolled back: %w", cause)
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
	return f.Close()
}

func Recover(journalPath string) error {
	b, err := os.ReadFile(journalPath)
	if err != nil {
		return err
	}
	var j journal
	if err = json.Unmarshal(b, &j); err != nil {
		return err
	}
	if repository.ValidatePair(j.Runtime, j.Memory) == nil {
		_ = os.RemoveAll(j.RuntimeStage)
		_ = os.RemoveAll(j.MemoryStage)
		return os.Remove(journalPath)
	}
	if repository.ValidatePair(j.Runtime, j.MemoryStage) == nil {
		if err = promote(j.MemoryStage, j.Memory); err != nil {
			return err
		}
		if err = repository.ValidatePair(j.Runtime, j.Memory); err != nil {
			return err
		}
		_ = os.RemoveAll(j.RuntimeStage)
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
		return os.Remove(journalPath)
	}
	return fmt.Errorf("automatic recovery cannot prove a complete pair; inspect %s", journalPath)
}
