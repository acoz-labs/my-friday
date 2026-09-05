package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func loadTestBundle(t *testing.T) Bundle {
	t.Helper()
	bundle, err := LoadBundle("testdata")
	if err != nil {
		t.Fatal(err)
	}
	return bundle
}

func TestFrozenBundleValidates(t *testing.T) {
	bundle := loadTestBundle(t)
	if err := ValidateBundle(bundle); err != nil {
		t.Fatal(err)
	}
}

func TestReadJSONRejectsUnknownAndTrailingValues(t *testing.T) {
	for name, body := range map[string]string{
		"unknown":  `{"version":1,"revision":"x","capabilities":[],"surprise":true}`,
		"trailing": `{"version":1,"revision":"x","capabilities":[]} {}`,
	} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "fixture.json")
			if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
				t.Fatal(err)
			}
			var corpus CapabilityCorpus
			if err := ReadJSON(path, &corpus); err == nil {
				t.Fatal("invalid JSON contract accepted")
			}
		})
	}
}

func TestValidationRejectsDuplicatesUnsafePathsRevisionMismatchAndCycles(t *testing.T) {
	tests := map[string]func(*Bundle){
		"duplicate":   func(bundle *Bundle) { bundle.Capabilities.Capabilities[1].ID = bundle.Capabilities.Capabilities[0].ID },
		"unsafe-path": func(bundle *Bundle) { bundle.Tasks.Tasks[0].ReadPaths = []string{"../labels.json"} },
		"stale-dependency": func(bundle *Bundle) {
			bundle.Capabilities.Capabilities[1].Dependencies = []Dependency{{ID: "fixture-reader", Revision: "old"}}
		},
		"cycle": func(bundle *Bundle) {
			bundle.Capabilities.Capabilities[0].Dependencies = []Dependency{{ID: "archive-inspector", Revision: "r1"}}
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			bundle := loadTestBundle(t)
			mutate(&bundle)
			if err := ValidateBundle(bundle); err == nil {
				t.Fatal("invalid bundle accepted")
			}
		})
	}
}

func TestPrepareManifestDeclaresEveryCellAndRotatesModes(t *testing.T) {
	bundle := loadTestBundle(t)
	manifest, err := PrepareManifest(bundle, strings.Repeat("a", 40), []HarnessSpec{{ID: "codex", ExecutableVersion: "c1", Model: "m1", Config: "x"}, {ID: "claude", ExecutableVersion: "c2", Model: "m2", Config: "y"}})
	if err != nil {
		t.Fatal(err)
	}
	if got := len(manifest.Cells); got != 288 {
		t.Fatalf("cells = %d", got)
	}
	if strings.Join([]string{manifest.Cells[0].Mode, manifest.Cells[1].Mode, manifest.Cells[2].Mode}, ",") != "native-catalogue,lookup-direct,lookup-worker" {
		t.Fatal("first repetition order did not start at native baseline")
	}
	if strings.Join([]string{manifest.Cells[3].Mode, manifest.Cells[4].Mode, manifest.Cells[5].Mode}, ",") != "lookup-worker,lookup-direct,native-catalogue" {
		t.Fatal("warm repetition order was not reversed")
	}
	if manifest.Budgets.TrialToolCalls != 8 || manifest.Budgets.BatchTrials != 12 || manifest.Budgets.BatchWallSeconds != 1800 {
		t.Fatal("approved limits changed")
	}
}

func TestManifestValidationRejectsIdentityBudgetsAndIncompleteCartesianMatrix(t *testing.T) {
	bundle := loadTestBundle(t)
	manifest, err := PrepareManifest(bundle, strings.Repeat("a", 40), []HarnessSpec{{ID: "codex", ExecutableVersion: "c1", Model: "m1", Config: "x"}, {ID: "claude", ExecutableVersion: "c2", Model: "m2", Config: "y"}})
	if err != nil {
		t.Fatal(err)
	}
	tests := map[string]func(*Manifest){
		"unknown-harness": func(value *Manifest) { value.Cells[0].HarnessID = "other" },
		"unknown-task":    func(value *Manifest) { value.Cells[0].TaskID = "other" },
		"bad-repetition":  func(value *Manifest) { value.Cells[0].Repetition = 3 },
		"bad-cache":       func(value *Manifest) { value.Cells[0].Cache = "hot" },
		"changed-budget":  func(value *Manifest) { value.Budgets.TrialToolCalls = 9 },
		"duplicate-cell": func(value *Manifest) {
			value.Cells[1].HarnessID = value.Cells[0].HarnessID
			value.Cells[1].TaskID = value.Cells[0].TaskID
			value.Cells[1].Mode = value.Cells[0].Mode
			value.Cells[1].Repetition = value.Cells[0].Repetition
			value.Cells[1].TrialID = "syntactically-unique"
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			candidate := manifest
			candidate.Cells = append([]ManifestCell{}, manifest.Cells...)
			mutate(&candidate)
			bundle.Manifest = candidate
			if err := validateManifest(bundle); err == nil {
				t.Fatal("invalid manifest accepted")
			}
		})
	}
}

func TestBM25RankingAndNoMatch(t *testing.T) {
	bundle := loadTestBundle(t)
	index := NewBM25Index(bundle.Capabilities.Capabilities)
	results := index.Query("normalize CSV headers and row widths", 3)
	if len(results) == 0 || results[0].ID != "csv-normalizer" {
		t.Fatalf("results = %#v", results)
	}
	if results := index.Query("xylophone nebula quantum", 3); len(results) != 0 {
		t.Fatalf("no-match results = %#v", results)
	}
}

func TestBM25TieBreaksByCapabilityID(t *testing.T) {
	index := NewBM25Index([]Capability{{ID: "z-last", Name: "Same", Summary: "term"}, {ID: "a-first", Name: "Same", Summary: "term"}})
	results := index.Query("term", 2)
	if results[0].ID != "a-first" {
		t.Fatalf("tie order = %#v", results)
	}
}
