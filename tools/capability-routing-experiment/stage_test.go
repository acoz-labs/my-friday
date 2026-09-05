package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStageTrialKeepsLabelsAndOtherTasksOutsideModelRoot(t *testing.T) {
	bundle := loadTestBundle(t)
	root := filepath.Join(t.TempDir(), "model")
	cell := ManifestCell{TaskID: "held-explicit-archive", Mode: "lookup-direct"}
	if err := StageTrial(root, bundle, cell); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "labels.json")); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("held-out labels entered model root")
	}
	body, err := os.ReadFile(filepath.Join(root, "task.json"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(body), "held-paraphrase-duplicates") || strings.Contains(string(body), "required_facts") || strings.Contains(string(body), "held-out") || strings.Contains(string(body), "explicit-selection") {
		t.Fatal("other tasks or rubric leaked into staged task")
	}
	policy, err := os.ReadFile(filepath.Join(root, "policy.txt"))
	if err != nil {
		t.Fatal(err)
	}
	for _, marker := range []string{"TASK INTENT", "READ AUTHORITY", "WRITE AUTHORITY", "DENY BY DEFAULT", "REQUIRED REPORT"} {
		if !strings.Contains(string(policy), marker) {
			t.Fatalf("root policy omits %s", marker)
		}
	}
}

func TestStageTrialProjectsNativeBodiesOnlyInNativeMode(t *testing.T) {
	bundle := loadTestBundle(t)
	for _, test := range []struct {
		mode          string
		skills, index bool
	}{{"native-catalogue", true, false}, {"lookup-direct", false, true}, {"lookup-worker", false, true}} {
		t.Run(test.mode, func(t *testing.T) {
			root := filepath.Join(t.TempDir(), "model")
			if err := StageTrial(root, bundle, ManifestCell{TaskID: "dev-short-checksum", Mode: test.mode}); err != nil {
				t.Fatal(err)
			}
			_, skillErr := os.Stat(filepath.Join(root, "skills", "checksum-verifier", "SKILL.md"))
			_, indexErr := os.Stat(filepath.Join(root, "index.json"))
			if (skillErr == nil) != test.skills || (indexErr == nil) != test.index {
				t.Fatalf("unexpected projection skills=%v index=%v", skillErr, indexErr)
			}
		})
	}
}

func TestSelectionRefusesStaleRevisionAndBudgetsDependencies(t *testing.T) {
	bundle := loadTestBundle(t)
	var stale Task
	for _, task := range bundle.Tasks.Tasks {
		if task.ID == "dev-stale-index" {
			stale = task
		}
	}
	if _, err := SelectCapabilitySet(bundle, stale, []string{"inventory-counter"}); !errors.Is(err, ErrStaleIndex) {
		t.Fatalf("stale selection err=%v", err)
	}
	stale.IndexRevision = bundle.Capabilities.Revision
	if _, err := SelectCapabilitySet(bundle, stale, []string{"chart-builder", "incident-reporter"}); err == nil {
		t.Fatal("two selected capabilities loaded two dependencies")
	}
}
