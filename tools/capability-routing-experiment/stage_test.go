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
		mode              string
		skills, transport bool
	}{{"native-catalogue", true, false}, {"lookup-direct", false, true}, {"lookup-worker", false, true}} {
		t.Run(test.mode, func(t *testing.T) {
			root := filepath.Join(t.TempDir(), "model")
			if err := StageTrial(root, bundle, ManifestCell{TaskID: "dev-short-checksum", Mode: test.mode}); err != nil {
				t.Fatal(err)
			}
			_, skillErr := os.Stat(filepath.Join(root, "skills", "checksum-verifier", "SKILL.md"))
			_, indexErr := os.Stat(filepath.Join(root, "index.json"))
			_, transportErr := os.Stat(filepath.Join(root, "transport.json"))
			if indexErr == nil || (skillErr == nil) != test.skills || (transportErr == nil) != test.transport {
				t.Fatalf("unexpected projection skills=%v index=%v transport=%v", skillErr, indexErr, transportErr)
			}
		})
	}
}

func TestCountedTransportLooksUpLoadsAndFallsBackWithoutPreloading(t *testing.T) {
	bundle := loadTestBundle(t)
	task, _ := findTaskLabel(t, bundle, "dev-explicit-csv")
	transport, err := NewCountedTransport(bundle, task)
	if err != nil {
		t.Fatal(err)
	}
	metadata, err := transport.Lookup("normalize CSV headers")
	if err != nil || len(metadata) == 0 || len(metadata) > 3 {
		t.Fatalf("metadata=%#v err=%v", metadata, err)
	}
	loaded, err := transport.Load([]string{metadata[0].ID})
	if err != nil || len(loaded) == 0 || loaded[0].Body == "" {
		t.Fatalf("loaded=%#v err=%v", loaded, err)
	}
	all, err := transport.Fallback()
	if err != nil || len(all) != 24 {
		t.Fatalf("fallback=%d err=%v", len(all), err)
	}
	if _, err = transport.Fallback(); err == nil {
		t.Fatal("second broader-metadata fallback accepted")
	}
	if transport.Calls() != 4 {
		t.Fatalf("calls=%d; rejected call must be counted before dispatch", transport.Calls())
	}
}

func TestCountedTransportEnforcesAuthorityAndCallCeiling(t *testing.T) {
	bundle := loadTestBundle(t)
	task, _ := findTaskLabel(t, bundle, "dev-explicit-csv")
	transport, err := NewCountedTransport(bundle, task)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = transport.ReadFixture("../labels.json"); err == nil {
		t.Fatal("unsafe read accepted")
	}
	if err = transport.WriteFixture("output/not-authorized.txt", "no"); err == nil {
		t.Fatal("unauthorized write accepted")
	}
	for transport.Calls() < 8 {
		_, _ = transport.Lookup("csv")
	}
	if _, err = transport.Lookup("csv"); err == nil {
		t.Fatal("ninth call accepted")
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
