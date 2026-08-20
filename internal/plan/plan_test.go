package plan

import (
	"github.com/acoz-labs/my-friday/internal/profile"
	"testing"
)

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
