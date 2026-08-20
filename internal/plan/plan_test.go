package plan

import (
	"github.com/acoz-labs/my-friday/internal/profile"
	"os"
	"path/filepath"
	"testing"
)

func TestBuildCanonicalizesSymlinkAncestorsAndCaseFoldCollisions(t *testing.T) {
	root := t.TempDir()
	real := filepath.Join(root, "real")
	if err := os.Mkdir(real, 0700); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "link")
	if err := os.Symlink(real, link); err != nil {
		t.Fatal(err)
	}
	p, _ := profile.New("Friday", "", "Help", "balanced", "")
	pl, err := Build(p, filepath.Join(link, "runtime"), filepath.Join(real, "memory"))
	if err != nil {
		t.Fatal(err)
	}
	canonicalReal, _ := filepath.EvalSymlinks(real)
	if pl.Targets.Runtime != filepath.Join(canonicalReal, "runtime") {
		t.Fatalf("canonical runtime = %s", pl.Targets.Runtime)
	}
	if _, err = Build(p, filepath.Join(real, "Pair"), filepath.Join(real, "pair", "memory")); err == nil {
		t.Fatal("expected case-folded nesting denial")
	}
}

func TestMemoryBaselineMatchesV1Contract(t *testing.T) {
	p, _ := profile.New("Friday", "", "Help", "balanced", "")
	pl, err := Build(p, filepath.Join(t.TempDir(), "runtime"), filepath.Join(t.TempDir(), "memory"))
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{"data/observations/.gitkeep": true, "data/journals/.gitkeep": true, "data/proposals/.gitkeep": true, "data/memories/.gitkeep": true, "schemas/README.md": true}
	for _, file := range pl.Files {
		if file.Role == "memory" {
			delete(want, file.Path)
		}
	}
	if len(want) != 0 {
		t.Fatalf("missing memory baseline files: %v", want)
	}
}

func TestBuildIsDeterministicAndSeparate(t *testing.T) {
	p, _ := profile.New("Friday", "Boss", "Keep work inspectable", "concise", "")
	a, err := Build(p, "/tmp/runtime", "/tmp/memory")
	if err != nil {
		t.Fatal(err)
	}
	b, err := Build(p, "/tmp/runtime", "/tmp/memory")
	if err != nil {
		t.Fatal(err)
	}
	if a.PlanID != b.PlanID || a.AssistantID != b.AssistantID {
		t.Fatal("ids are not deterministic")
	}
	if a.Targets.Runtime == a.Targets.Memory {
		t.Fatal("targets must be separate")
	}
	if len(a.Files) == 0 || len(a.NegativeActions) == 0 {
		t.Fatal("preview must declare files and exclusions")
	}
}

func TestBuildRejectsNestedTargets(t *testing.T) {
	p, _ := profile.New("Friday", "", "Help", "balanced", "")
	if _, err := Build(p, "/tmp/a", "/tmp/a/memory"); err == nil {
		t.Fatal("expected nesting rejection")
	}
}

func TestBuildRejectsExactTargetSymlink(t *testing.T) {
	root := t.TempDir()
	real := filepath.Join(root, "real")
	if err := os.Mkdir(real, 0700); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "runtime")
	if err := os.Symlink(real, link); err != nil {
		t.Fatal(err)
	}
	p, _ := profile.New("Friday", "", "Help", "balanced", "")
	if _, err := Build(p, link, filepath.Join(root, "memory")); err == nil {
		t.Fatal("exact target symlink accepted")
	}
}
