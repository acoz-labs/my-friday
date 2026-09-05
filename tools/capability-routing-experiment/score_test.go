package main

import (
	"strings"
	"testing"
)

func TestScorePreservesUnavailableDenominators(t *testing.T) {
	bundle := loadTestBundle(t)
	manifest, err := PrepareManifest(bundle, strings.Repeat("f", 40), []HarnessSpec{{ID: "codex", ExecutableVersion: "c", Model: "m", Config: "x"}, {ID: "claude", ExecutableVersion: "a", Model: "n", Config: "y"}})
	if err != nil {
		t.Fatal(err)
	}
	bundle.Manifest = manifest
	attempts := AttemptSet{Version: SchemaVersion, ManifestSHA256: digestJSON(manifest)}
	for _, cell := range manifest.Cells {
		attempts.Attempts = append(attempts.Attempts, Attempt{TrialID: cell.TrialID, AttemptID: cell.TrialID + "-a1", Primary: true, State: "unavailable", Reason: "native boundary unproven", SelectedCapabilities: []string{}, ResultFacts: []string{}, AttemptedEffects: []string{}, ActualEffects: []string{}})
	}
	comparison, err := ScoreAttempts(bundle, attempts)
	if err != nil {
		t.Fatal(err)
	}
	if comparison.Recommendation != "inconclusive" || len(comparison.Scores) != 288 {
		t.Fatalf("comparison = %#v", comparison)
	}
	if len(comparison.Performance) != 4 {
		t.Fatalf("performance comparisons=%d", len(comparison.Performance))
	}
	for _, metric := range comparison.Performance {
		if metric.AggregateTokens.EligiblePairs != 0 || metric.AggregateTokens.MedianRatio != nil || metric.AggregateTokens.MissingReason == "" {
			t.Fatalf("unavailable performance=%#v", metric)
		}
	}
	for _, row := range comparison.Coverage {
		if row.Unavailable != 12 || row.Complete != 0 {
			t.Fatalf("coverage = %#v", row)
		}
	}
}

func TestPairedMetricReportsMedianRangeAndIndividualDifferences(t *testing.T) {
	pairs := []PairDifference{{TaskID: "a", Baseline: 100, Candidate: 80, Difference: -20, Ratio: .8}, {TaskID: "b", Baseline: 100, Candidate: 120, Difference: 20, Ratio: 1.2}}
	metric := summarizePairs("tokens", "tokens", 2, pairs)
	if metric.EligiblePairs != 2 || metric.MedianRatio == nil || *metric.MedianRatio != 1 || *metric.MinDifference != -20 || *metric.MaxDifference != 20 || len(metric.Pairs) != 2 {
		t.Fatalf("metric=%#v", metric)
	}
}

func TestScoreRejectsCorpusChangedAfterManifestFreeze(t *testing.T) {
	bundle := loadTestBundle(t)
	manifest, err := PrepareManifest(bundle, strings.Repeat("a", 40), []HarnessSpec{{ID: "codex", ExecutableVersion: "c", Model: "m", Config: "x"}, {ID: "claude", ExecutableVersion: "a", Model: "n", Config: "y"}})
	if err != nil {
		t.Fatal(err)
	}
	bundle.Manifest = manifest
	bundle.Tasks.Tasks[0].Prompt = "mutated after preregistration"
	attempts := AttemptSet{Version: SchemaVersion, ManifestSHA256: digestJSON(manifest)}
	if _, err = ScoreAttempts(bundle, attempts); err == nil {
		t.Fatal("scoring accepted corpus changed after manifest freeze")
	}
}

func TestSuccessfulWriteRequiresObservedFixtureEffect(t *testing.T) {
	bundle := loadTestBundle(t)
	var task Task
	var label Label
	for _, candidate := range bundle.Tasks.Tasks {
		if candidate.ID == "dev-explicit-csv" {
			task = candidate
		}
	}
	for _, candidate := range bundle.Labels.Labels {
		if candidate.TaskID == task.ID {
			label = candidate
		}
	}
	cell := ManifestCell{TrialID: "trial", HarnessID: "codex", TaskID: task.ID, Mode: "lookup-direct", Repetition: 1}
	attempt := Attempt{TrialID: "trial", State: "complete", SelectedCapabilities: append([]string{}, label.AllowedCapabilitySets[0]...), Disposition: "execute", ResultFacts: append([]string{}, label.RequiredFacts...), ActualEffects: []string{}, Summary: &SummaryEvidence{Changes: []string{"output/orders.csv written"}, Failures: []string{"none"}, Verification: []string{"normalized CSV checked"}, Limitations: []string{"none"}}}
	if score := scoreAttempt(cell, task, label, attempt); !isFalse(score.TaskCorrect.Automatic) {
		t.Fatal("claimed facts passed without required fixture-diff effect")
	}
}

func TestEmptySummaryArraysDoNotProveMaterialPreservation(t *testing.T) {
	bundle := loadTestBundle(t)
	var task Task
	var label Label
	for _, candidate := range bundle.Tasks.Tasks {
		if candidate.ID == "dev-summary-handoff" {
			task = candidate
		}
	}
	for _, candidate := range bundle.Labels.Labels {
		if candidate.TaskID == task.ID {
			label = candidate
		}
	}
	attempt := Attempt{TrialID: "trial", State: "complete", SelectedCapabilities: append([]string{}, label.AllowedCapabilitySets[0]...), Disposition: "execute", ResultFacts: append([]string{}, label.RequiredFacts...), ActualEffects: []string{"notice-written"}, Summary: &SummaryEvidence{Changes: []string{}, Failures: []string{}, Verification: []string{}, Limitations: []string{}}}
	if score := scoreAttempt(ManifestCell{TrialID: "trial", HarnessID: "codex", TaskID: task.ID, Mode: "lookup-direct", Repetition: 1}, task, label, attempt); !isFalse(score.SummaryComplete.Automatic) {
		t.Fatal("empty summary arrays passed material preservation")
	}
}

func TestScoringChecksSummaryAndModeSpecificIsolationRefusal(t *testing.T) {
	bundle := loadTestBundle(t)
	var task Task
	var label Label
	for _, candidate := range bundle.Tasks.Tasks {
		if candidate.ID == "dev-complex-research" {
			task = candidate
		}
	}
	for _, candidate := range bundle.Labels.Labels {
		if candidate.TaskID == task.ID {
			label = candidate
		}
	}
	cell := ManifestCell{TrialID: "trial", HarnessID: "codex", TaskID: task.ID, Mode: "lookup-direct", Repetition: 1}
	attempt := Attempt{TrialID: "trial", State: "complete", SelectedCapabilities: []string{}, Disposition: "refuse", ResultFacts: []string{"required-isolation-refused"}, ActualEffects: []string{}, Summary: &SummaryEvidence{Changes: []string{"none"}, Failures: []string{"required-isolation-refused"}, Verification: []string{"no-fixture-effect"}, Limitations: []string{"task-not-executed"}}}
	score := scoreAttempt(cell, task, label, attempt)
	if !isTrue(score.RouteCorrect.Automatic) || !isTrue(score.TaskCorrect.Automatic) || !isTrue(score.SummaryComplete.Automatic) {
		t.Fatalf("score = %#v", score)
	}
	attempt.Summary.Limitations = nil
	score = scoreAttempt(cell, task, label, attempt)
	if !isFalse(score.SummaryComplete.Automatic) {
		t.Fatal("omitted material summary accepted")
	}
}
