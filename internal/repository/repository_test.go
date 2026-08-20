package repository

import (
	"github.com/acoz-labs/my-friday/internal/plan"
	"github.com/acoz-labs/my-friday/internal/profile"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestCreateAndValidatePairWithoutCommitOrRemote(t *testing.T) {
	root := t.TempDir()
	p, _ := profile.New("Friday", "", "Help", "balanced", "")
	pl, _ := plan.Build(p, filepath.Join(root, "runtime"), filepath.Join(root, "memory"))
	if err := Create(pl, pl.Targets.Runtime, pl.Targets.Memory); err != nil {
		t.Fatal(err)
	}
	if err := ValidatePair(pl.Targets.Runtime, pl.Targets.Memory); err != nil {
		t.Fatal(err)
	}
	for _, dir := range []string{pl.Targets.Runtime, pl.Targets.Memory} {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err != nil {
			t.Fatal(err)
		}
		out, _ := exec.Command("git", "-C", dir, "remote").Output()
		if len(out) != 0 {
			t.Fatal("unexpected remote")
		}
		if exec.Command("git", "-C", dir, "rev-parse", "HEAD").Run() == nil {
			t.Fatal("unexpected commit")
		}
	}
}
