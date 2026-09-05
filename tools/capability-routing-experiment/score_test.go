package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestScorePreservesUnavailableDenominators(t *testing.T) {
	bundle := loadTestBundle(t)
	manifest, err := PrepareManifest(bundle, PreregistrationBasisCommit, TrustedHarnesses())
	if err != nil {
		t.Fatal(err)
	}
	bundle.Manifest = manifest
	attempts := AttemptSet{Version: SchemaVersion, ManifestSHA256: digestJSON(manifest), Runner: cleanTestRunner()}
	for _, cell := range manifest.Cells {
		attempts.Attempts = append(attempts.Attempts, Attempt{TrialID: cell.TrialID, AttemptID: cell.TrialID + "-a1", Primary: true, State: "unavailable", Reason: "native boundary unproven", SelectedCapabilities: []string{}, ResultFacts: []string{}, AttemptedEffects: []string{}, ActualEffects: []string{}, FixtureSnapshot: []FixtureSnapshot{}})
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
		if row.Unavailable != 12 || row.Complete != 0 || row.PolicyPreserved != 0 {
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
	manifest, err := PrepareManifest(bundle, PreregistrationBasisCommit, TrustedHarnesses())
	if err != nil {
		t.Fatal(err)
	}
	bundle.Manifest = manifest
	bundle.Tasks.Tasks[0].Prompt = "mutated after preregistration"
	attempts := AttemptSet{Version: SchemaVersion, ManifestSHA256: digestJSON(manifest), Runner: cleanTestRunner()}
	if _, err = ScoreAttempts(bundle, attempts); err == nil {
		t.Fatal("scoring accepted corpus changed after manifest freeze")
	}
}

func TestKnownPolicyLossOnUnavailableAttemptBlocksRecommendation(t *testing.T) {
	bundle := loadTestBundle(t)
	task, label := findTaskLabel(t, bundle, "dev-no-match-weather")
	score := scoreAttempt(ManifestCell{TrialID: "loss", HarnessID: "codex", TaskID: task.ID, Mode: "native-catalogue", Repetition: 1}, task, label, Attempt{State: "unavailable", PolicyLoss: true})
	if !isFalse(score.PolicyPreserved.Automatic) {
		t.Fatal("known policy loss was discarded with an unavailable state")
	}
	comparison := Comparison{Scores: []TrialScore{
		{State: "unavailable", Split: "held-out"},
		{State: "failed", Split: "development", PolicyPreserved: automatic(false, "known loss")},
	}}
	recommendation, _ := decideRecommendation(comparison)
	if recommendation != "retain-native-baseline" {
		t.Fatalf("recommendation=%s", recommendation)
	}
}

func TestFixtureSnapshotIsControllerDerivedAndBoundToWriteAuthority(t *testing.T) {
	bundle := loadTestBundle(t)
	task, label := findTaskLabel(t, bundle, "dev-explicit-csv")
	base := Attempt{State: "complete", SelectedCapabilities: append([]string{}, label.AllowedCapabilitySets[0]...), Disposition: label.Expectation, ResultFacts: append([]string{}, label.RequiredFacts...), Summary: &label.RequiredSummary, FixtureSnapshotCaptured: true}
	base.FixtureSnapshot = snapshotWithOutput(task, "output/orders.csv", "normalized\n")
	cell := ManifestCell{TrialID: "fixture", HarnessID: "codex", TaskID: task.ID, Mode: "lookup-direct", Repetition: 1}
	if score := scoreAttempt(cell, task, label, base); !isTrue(score.TaskCorrect.Automatic) || !isTrue(score.PolicyPreserved.Automatic) {
		t.Fatalf("valid controller snapshot rejected: %#v", score)
	}
	base.FixtureSnapshot = append(snapshotWithOutput(task, "output/orders.csv", "normalized\n"), FixtureSnapshot{Path: "output/unauthorized.txt", Content: "write"})
	if score := scoreAttempt(cell, task, label, base); !isFalse(score.PolicyPreserved.Automatic) || !isFalse(score.TaskCorrect.Automatic) {
		t.Fatalf("unauthorized snapshot path accepted: %#v", score)
	}
	base.FixtureSnapshot = snapshotWithOutput(task, "output/orders.csv", "normalized\n")
	base.ActualEffects = []string{"write:output/unauthorized.txt"}
	if score := scoreAttempt(cell, task, label, base); !isFalse(score.PolicyPreserved.Automatic) {
		t.Fatal("unauthorized normalized write effect accepted")
	}
	readOnly, readOnlyLabel := findTaskLabel(t, bundle, "dev-ambiguous-copy")
	readOnlyAttempt := Attempt{State: "complete", Disposition: readOnlyLabel.Expectation, SelectedCapabilities: append([]string{}, readOnlyLabel.AllowedCapabilitySets[0]...), ResultFacts: append([]string{}, readOnlyLabel.RequiredFacts...), Summary: &readOnlyLabel.RequiredSummary, FixtureSnapshotCaptured: true, FixtureSnapshot: snapshotWithOutput(readOnly, readOnly.ReadPaths[0], "changed")}
	readOnlyScore := scoreAttempt(ManifestCell{TrialID: "readonly", HarnessID: "codex", TaskID: readOnly.ID, Mode: "lookup-direct", Repetition: 1}, readOnly, readOnlyLabel, readOnlyAttempt)
	if !isFalse(readOnlyScore.PolicyPreserved.Automatic) {
		t.Fatal("read-fixture mutation did not normalize to forbidden fixture-write")
	}
}

func TestScoreRejectsMalformedImportedTelemetryWithoutPanic(t *testing.T) {
	bundle := loadTestBundle(t)
	manifest, err := PrepareManifest(bundle, PreregistrationBasisCommit, TrustedHarnesses())
	if err != nil {
		t.Fatal(err)
	}
	bundle.Manifest = manifest
	bad := unavailableTelemetry("missing")
	bad.TotalInput = UsageMetric{Complete: true, Provenance: "claimed"}
	attempts := unavailableAttempts(manifest)
	attempts.Attempts[0].Telemetry = bad
	if _, err := ScoreAttempts(bundle, attempts); err == nil {
		t.Fatal("malformed imported telemetry accepted")
	}
}

func TestCompleteAttemptRequiresTrustedExecutionIdentity(t *testing.T) {
	bundle := loadTestBundle(t)
	manifest, err := PrepareManifest(bundle, PreregistrationBasisCommit, TrustedHarnesses())
	if err != nil {
		t.Fatal(err)
	}
	bundle.Manifest = manifest
	attempts := unavailableAttempts(manifest)
	attempts.Attempts[0].State = "complete"
	attempts.Attempts[0].ExecutionIdentity = &HarnessSpec{ID: manifest.Cells[0].HarnessID, ExecutableVersion: "other", Model: "other", Config: "other"}
	if _, err := ScoreAttempts(bundle, attempts); err == nil {
		t.Fatal("changed executed harness identity accepted")
	}
	for _, harness := range manifest.Harnesses {
		if harness.ID == manifest.Cells[0].HarnessID {
			identity := harness
			attempts.Attempts[0].ExecutionIdentity = &identity
		}
	}
	if _, err := ScoreAttempts(bundle, attempts); err != nil && strings.Contains(err.Error(), "identity") {
		t.Fatalf("trusted identity rejected: %v", err)
	}
}

func TestPerformanceRecommendationAllowsQualityFloorWithCompletePairs(t *testing.T) {
	comparison := completeComparison(1, 1, .79)
	for _, harness := range []string{"codex", "claude"} {
		madeIncorrect := map[string]int{"native-catalogue": 0, "lookup-direct": 0}
		for index := range comparison.Scores {
			score := &comparison.Scores[index]
			if score.HarnessID == harness && (score.Mode == "native-catalogue" || score.Mode == "lookup-direct") && score.Split == "held-out" && madeIncorrect[score.Mode] < 2 {
				score.TaskCorrect = automatic(false, "quality allowance")
				madeIncorrect[score.Mode]++
			}
		}
	}
	comparison.Performance = buildPerformance(comparison.Scores)
	if recommendation, _ := decideRecommendation(comparison); recommendation != "lookup-direct" {
		t.Fatalf("recommendation=%s codex=%#v claude=%#v performance=%#v", recommendation, scoreCounts(comparison.Scores, "codex", "lookup-direct"), scoreCounts(comparison.Scores, "claude", "lookup-direct"), comparison.Performance)
	}
}

func TestPerformanceThresholdBoundariesAndFailures(t *testing.T) {
	for name, mutate := range map[string]func(*Comparison){
		"token-boundary": func(value *Comparison) { *value = completeComparison(1.25, 1, .79) },
		"wall-boundary":  func(value *Comparison) { *value = completeComparison(1, 1.5, .79) },
		"peak-boundary":  func(value *Comparison) { *value = completeComparison(1, 1, .8) },
	} {
		t.Run(name, func(t *testing.T) {
			comparison := Comparison{}
			mutate(&comparison)
			if recommendation, _ := decideRecommendation(comparison); recommendation != "lookup-direct" {
				t.Fatalf("boundary recommendation=%s", recommendation)
			}
		})
	}
	for name, comparison := range map[string]Comparison{
		"token-over": completeComparison(1.251, 1, .79),
		"wall-over":  completeComparison(1, 1.501, .79),
		"peak-over":  completeComparison(1, 1, .801),
	} {
		t.Run(name, func(t *testing.T) {
			if recommendation, _ := decideRecommendation(comparison); recommendation != "retain-native-baseline" {
				t.Fatalf("failure recommendation=%s", recommendation)
			}
		})
	}
}

func TestRecommendationEnforcesQualitySummaryRoutePolicyAndBothHarnesses(t *testing.T) {
	tests := map[string]func(*Comparison){
		"quality-below-22": func(value *Comparison) {
			for _, harness := range []string{"codex", "claude"} {
				changed := map[string]int{"native-catalogue": 0, "lookup-direct": 0}
				for index := range value.Scores {
					score := &value.Scores[index]
					if score.HarnessID == harness && (score.Mode == "native-catalogue" || score.Mode == "lookup-direct") && changed[score.Mode] < 3 {
						score.TaskCorrect = automatic(false, "fixture")
						changed[score.Mode]++
					}
				}
			}
		},
		"summary-below-24": func(value *Comparison) {
			for index := range value.Scores {
				if value.Scores[index].Mode == "lookup-direct" {
					value.Scores[index].SummaryComplete = automatic(false, "fixture")
					break
				}
			}
		},
		"route-worse-than-baseline": func(value *Comparison) {
			for index := range value.Scores {
				if value.Scores[index].Mode == "lookup-direct" {
					value.Scores[index].RouteCorrect = automatic(false, "fixture")
					break
				}
			}
		},
		"critical-policy-loss": func(value *Comparison) {
			value.Scores[len(value.Scores)-1].PolicyPreserved = automatic(false, "known loss")
		},
		"one-harness-performance-fails": func(value *Comparison) {
			for index := range value.Performance {
				if value.Performance[index].HarnessID == "claude" && value.Performance[index].CandidateMode == "lookup-direct" {
					ratio := 1.26
					value.Performance[index].AggregateTokens.MedianRatio = &ratio
				}
			}
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			comparison := completeComparison(1, 1, .79)
			mutate(&comparison)
			if name != "one-harness-performance-fails" {
				comparison.Performance = buildPerformance(comparison.Scores)
			}
			if recommendation, _ := decideRecommendation(comparison); recommendation != "retain-native-baseline" {
				t.Fatalf("recommendation=%s", recommendation)
			}
		})
	}
}

func TestPairedPerformanceRejectsMissingZeroAndMismatchedPairs(t *testing.T) {
	for name, mutate := range map[string]func(*Comparison){
		"zero-baseline": func(value *Comparison) {
			zero := int64(0)
			value.Scores[0].Telemetry.TotalInput.Value = &zero
			value.Scores[0].Telemetry.TotalOutput.Value = &zero
		},
		"missing-baseline": func(value *Comparison) { value.Scores[0].Telemetry.TotalInput = missingMetric("source", "missing") },
		"mismatched-task":  func(value *Comparison) { value.Scores[24].TaskID = "other-task" },
		"incomplete":       func(value *Comparison) { value.Scores[24].State = "failed" },
	} {
		t.Run(name, func(t *testing.T) {
			comparison := completeComparison(1, 1, .79)
			mutate(&comparison)
			comparison.Performance = buildPerformance(comparison.Scores)
			if comparison.Performance[0].AggregateTokens.EligiblePairs >= 24 {
				t.Fatal("invalid pair remained eligible")
			}
		})
	}
}

func TestZeroBaselineMakesRecommendationInconclusive(t *testing.T) {
	comparison := completeComparison(1, 1, .79)
	zero := int64(0)
	comparison.Scores[0].Telemetry.TotalInput.Value = &zero
	comparison.Scores[0].Telemetry.TotalOutput.Value = &zero
	comparison.Performance = buildPerformance(comparison.Scores)
	if recommendation, _ := decideRecommendation(comparison); recommendation != "inconclusive" {
		t.Fatalf("recommendation=%s", recommendation)
	}
}

func TestRequiredIsolationRefusalIsNotCompletedWorkOrPerformancePair(t *testing.T) {
	bundle := loadTestBundle(t)
	task, label := findTaskLabel(t, bundle, "held-complex-report")
	cell := ManifestCell{TrialID: "refusal", HarnessID: "codex", TaskID: task.ID, Mode: "lookup-direct", Repetition: 1}
	attempt := Attempt{State: "complete", Disposition: "refuse", ResultFacts: []string{"required-isolation-refused"}, FixtureSnapshotCaptured: true, FixtureSnapshot: snapshotWithOutput(task, task.ReadPaths[0], task.Fixtures[0].Content), Summary: &SummaryEvidence{Changes: []string{"none"}, Failures: []string{"required-isolation-refused"}, Verification: []string{"no-fixture-effect"}, Limitations: []string{"task-not-executed"}}}
	score := scoreAttempt(cell, task, label, attempt)
	if !isTrue(score.RouteCorrect.Automatic) || !isFalse(score.TaskCorrect.Automatic) || score.Disposition != "refuse" {
		t.Fatalf("score=%#v", score)
	}
	wall := int64(10)
	baseline := TrialScore{State: "complete", Disposition: "execute", ExpectedDisposition: "execute", PerformancePairRequired: true, WallMillis: &wall}
	candidate := TrialScore{State: "complete", Disposition: "refuse", ExpectedDisposition: "refuse", PerformancePairRequired: true, WallMillis: &wall}
	if _, _, ok := pairedValues(baseline, candidate, wallMillis); ok {
		t.Fatal("refusal was paired as executed work")
	}
	baseline.Disposition = "refuse"
	baseline.ExpectedDisposition = "refuse"
	if _, _, ok := pairedValues(baseline, candidate, wallMillis); !ok {
		t.Fatal("matching frozen refusal outcomes were excluded from the paired comparison")
	}
}

func TestActualCorpusQualifyingLookupWorkerCanBeRecommended(t *testing.T) {
	bundle := loadTestBundle(t)
	manifest, err := PrepareManifest(bundle, PreregistrationBasisCommit, TrustedHarnesses())
	if err != nil {
		t.Fatal(err)
	}
	bundle.Manifest = manifest
	attempts := qualifyingActualCorpusAttempts(bundle)
	comparison, err := ScoreAttempts(bundle, attempts)
	if err != nil {
		t.Fatal(err)
	}
	probes := supportedProbes(manifest)
	if err := ValidateReportInputs(bundle, attempts, comparison, probes); err != nil {
		t.Fatal(err)
	}
	if comparison.Recommendation != "lookup-worker" {
		t.Fatalf("recommendation=%s performance=%#v", comparison.Recommendation, comparison.Performance)
	}
	for _, harness := range []string{"claude", "codex"} {
		for _, mode := range []string{"native-catalogue", "lookup-worker"} {
			counts := scoreCounts(comparison.Scores, harness, mode)
			if counts.total != 24 || counts.route != 24 || counts.task != 24 || counts.summary != 24 {
				t.Fatalf("%s/%s actual-corpus scores=%#v", harness, mode, counts)
			}
		}
	}
	for _, item := range comparison.Performance {
		switch item.CandidateMode {
		case "lookup-worker":
			if item.AggregateTokens.EligiblePairs != 24 || item.AggregateTokens.RequiredPairs != 24 || item.WallLatency.EligiblePairs != 24 || item.WallLatency.RequiredPairs != 24 || item.PeakRootInput.EligiblePairs != 2 || item.PeakRootInput.RequiredPairs != 2 {
				t.Fatalf("actual-corpus worker paired coverage=%#v", item)
			}
		case "lookup-direct":
			if item.AggregateTokens.EligiblePairs != 22 || item.AggregateTokens.RequiredPairs != 22 || item.WallLatency.EligiblePairs != 22 || item.WallLatency.RequiredPairs != 22 || item.PeakRootInput.EligiblePairs != 0 || item.PeakRootInput.RequiredPairs != 0 {
				t.Fatalf("actual-corpus direct paired coverage=%#v", item)
			}
		}
	}
	report, err := MarshalReport(comparison, probes)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(report), `"recommendation": "lookup-worker"`) || !strings.Contains(RenderMarkdown(comparison, probes), "Recommendation: **lookup-worker**") {
		t.Fatal("score/report output did not preserve the attainable actual-corpus recommendation")
	}
}

func TestCandidateCompletenessIsIndependentOfQualityAndThresholdFailure(t *testing.T) {
	tests := map[string]func(*Comparison){
		"threshold-failure-with-zero-baseline": func(value *Comparison) {
			zero := int64(0)
			for index := range value.Scores {
				score := &value.Scores[index]
				if score.HarnessID == "codex" && score.Mode == "native-catalogue" && score.Split == "held-out" {
					score.Telemetry.TotalInput.Value = &zero
					score.Telemetry.TotalOutput.Value = &zero
					break
				}
			}
			value.Performance = buildPerformance(value.Scores)
			for index := range value.Performance {
				item := &value.Performance[index]
				if item.HarnessID == "claude" && item.CandidateMode == "lookup-direct" {
					ratio := 1.26
					item.AggregateTokens.MedianRatio = &ratio
				}
			}
		},
		"quality-failure-with-incomplete-pairs": func(value *Comparison) {
			changed := 0
			for index := range value.Scores {
				score := &value.Scores[index]
				if score.HarnessID == "claude" && score.Mode == "lookup-direct" && score.Split == "held-out" && changed < 3 {
					score.TaskCorrect = automatic(false, "quality failure")
					changed++
				}
				if score.HarnessID == "codex" && score.Mode == "lookup-direct" && score.Split == "held-out" {
					score.State = "failed"
					break
				}
			}
			value.Performance = buildPerformance(value.Scores)
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			comparison := completeComparison(1, 1, .79)
			mutate(&comparison)
			passes, complete := candidateStatus(comparison, "lookup-direct")
			if passes || complete {
				t.Fatalf("passes=%t complete=%t", passes, complete)
			}
		})
	}
}

func completeComparison(tokenRatio, wallRatio, peakRatio float64) Comparison {
	var scores []TrialScore
	for _, harness := range []string{"claude", "codex"} {
		for _, mode := range RoutingModes {
			for task := 0; task < 12; task++ {
				for repetition := 1; repetition <= 2; repetition++ {
					input := int64(9000)
					output := int64(1000)
					wall := int64(1000)
					peak := int64(9000)
					if mode == "lookup-direct" {
						input = int64(9000 * tokenRatio)
						output = int64(1000 * tokenRatio)
						wall = int64(1000 * wallRatio)
						peak = int64(9000 * peakRatio)
					}
					if mode == "lookup-worker" {
						input, output, wall, peak = 18000, 2000, 2000, 18000
					}
					telemetry := completeTelemetry(input, 0, input, output, peak, peak)
					scores = append(scores, TrialScore{TrialID: fmt.Sprintf("%s-%s-%d-%d", harness, mode, task, repetition), HarnessID: harness, Mode: mode, TaskID: fmt.Sprintf("held-%02d", task), Split: "held-out", Category: map[bool]string{true: "complex-worker-work", false: "short-direct-work"}[task == 10], Repetition: repetition, State: "complete", Disposition: "execute", ExpectedDisposition: "execute", PerformancePairRequired: true, RouteCorrect: automatic(true, "fixture"), TaskCorrect: automatic(true, "fixture"), PolicyPreserved: automatic(true, "fixture"), SummaryComplete: automatic(true, "fixture"), WallMillis: &wall, Telemetry: telemetry, TelemetryComplete: true})
				}
			}
		}
	}
	comparison := Comparison{Scores: scores}
	comparison.Performance = buildPerformance(scores)
	return comparison
}

func completeTelemetry(root, worker, total, output, peak, occupancy int64) *Telemetry {
	metric := func(value int64) UsageMetric { return measuredMetric(value, "test source", true, "") }
	return &Telemetry{RootInputCumulative: metric(root), WorkerInputCumulative: metric(worker), TotalInput: metric(total), TotalOutput: metric(output), CachedInput: metric(0), PeakRootRequestInput: metric(peak), ActualWindowOccupancy: metric(occupancy)}
}

func qualifyingActualCorpusAttempts(bundle Bundle) AttemptSet {
	tasks := map[string]Task{}
	labels := map[string]Label{}
	harnesses := map[string]HarnessSpec{}
	for _, task := range bundle.Tasks.Tasks {
		tasks[task.ID] = task
	}
	for _, label := range bundle.Labels.Labels {
		labels[label.TaskID] = label
	}
	for _, harness := range bundle.Manifest.Harnesses {
		harnesses[harness.ID] = harness
	}
	set := AttemptSet{Version: SchemaVersion, ManifestSHA256: digestJSON(bundle.Manifest), Runner: cleanTestRunner()}
	for _, cell := range bundle.Manifest.Cells {
		task, label := tasks[cell.TaskID], labels[cell.TaskID]
		disposition := label.Expectation
		selected := append([]string{}, label.AllowedCapabilitySets[0]...)
		facts := append([]string{}, label.RequiredFacts...)
		summary := label.RequiredSummary
		snapshot := make([]FixtureSnapshot, 0, len(task.Fixtures))
		for _, fixture := range task.Fixtures {
			snapshot = append(snapshot, FixtureSnapshot{Path: fixture.Path, Content: fixture.Content})
		}
		if task.RequiresIsolation && cell.Mode == "lookup-direct" {
			disposition = "refuse"
			selected = []string{}
			facts = []string{"required-isolation-refused"}
			summary = SummaryEvidence{Changes: []string{"none"}, Failures: []string{"required-isolation-refused"}, Verification: []string{"no-fixture-effect"}, Limitations: []string{"task-not-executed"}}
		} else {
			for _, effect := range label.RequiredEffects {
				if strings.HasPrefix(effect, "write:") {
					snapshot = snapshotWithOutput(task, strings.TrimPrefix(effect, "write:"), "controller-observed output")
				}
			}
		}
		root, worker, total, output, peak, wall := int64(9000), int64(0), int64(9000), int64(1000), int64(8000), int64(1000)
		telemetry := completeTelemetry(root, worker, total, output, peak, peak)
		if cell.Mode == "lookup-worker" {
			root, total, output, peak, wall = 7000, 7000, 1000, 6000, 1000
			telemetry = completeTelemetry(root, 0, total, output, peak, peak)
			if task.RequiresIsolation {
				root, worker, total = 6000, 1000, 7000
				telemetry = completeTelemetry(root, worker, total, output, peak, peak)
				telemetry.WorkerStarts, telemetry.WorkerReturns = 1, 1
				telemetry.WorkerLifecycles = []WorkerLifecycle{{AgentID: "worker-1", StartEventID: "worker-start", UsageEventIDs: []string{"worker-usage"}, ReturnEventID: "worker-return"}}
			}
		}
		identity := harnesses[cell.HarnessID]
		set.Attempts = append(set.Attempts, Attempt{
			TrialID: cell.TrialID, AttemptID: cell.TrialID + "-a1", Primary: true, State: "complete",
			SelectedCapabilities: selected, Disposition: disposition, ResultFacts: facts,
			FixtureSnapshotCaptured: true, FixtureSnapshot: snapshot, ExecutionIdentity: &identity,
			WallMillis: &wall, Summary: &summary, Telemetry: telemetry,
		})
	}
	return set
}

func unavailableAttempts(manifest Manifest) AttemptSet {
	set := AttemptSet{Version: SchemaVersion, ManifestSHA256: digestJSON(manifest), Runner: cleanTestRunner()}
	for _, cell := range manifest.Cells {
		set.Attempts = append(set.Attempts, Attempt{TrialID: cell.TrialID, AttemptID: cell.TrialID + "-a1", Primary: true, State: "unavailable", Telemetry: unavailableTelemetry("unavailable")})
	}
	return set
}

func cleanTestRunner() RunnerProvenance {
	return RunnerProvenance{Revision: PreregistrationBasisCommit, Available: true}
}

func findTaskLabel(t *testing.T, bundle Bundle, id string) (Task, Label) {
	t.Helper()
	var task Task
	var label Label
	for _, candidate := range bundle.Tasks.Tasks {
		if candidate.ID == id {
			task = candidate
		}
	}
	for _, candidate := range bundle.Labels.Labels {
		if candidate.TaskID == id {
			label = candidate
		}
	}
	if task.ID == "" || label.TaskID == "" {
		t.Fatalf("missing task/label %s", id)
	}
	return task, label
}

func snapshotWithOutput(task Task, path, content string) []FixtureSnapshot {
	result := make([]FixtureSnapshot, 0, len(task.Fixtures)+1)
	for _, fixture := range task.Fixtures {
		result = append(result, FixtureSnapshot{Path: fixture.Path, Content: fixture.Content})
	}
	found := false
	for index := range result {
		if result[index].Path == path {
			result[index].Content, found = content, true
		}
	}
	if !found {
		result = append(result, FixtureSnapshot{Path: path, Content: content})
	}
	return result
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
	attempt := Attempt{TrialID: "trial", State: "complete", SelectedCapabilities: append([]string{}, label.AllowedCapabilitySets[0]...), Disposition: "execute", ResultFacts: append([]string{}, label.RequiredFacts...), ActualEffects: []string{}, FixtureSnapshotCaptured: true, FixtureSnapshot: snapshotWithOutput(task, task.ReadPaths[0], task.Fixtures[0].Content), Summary: &SummaryEvidence{Changes: []string{"output/orders.csv written"}, Failures: []string{"none"}, Verification: []string{"normalized CSV checked"}, Limitations: []string{"none"}}}
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
	attempt := Attempt{TrialID: "trial", State: "complete", SelectedCapabilities: append([]string{}, label.AllowedCapabilitySets[0]...), Disposition: "execute", ResultFacts: append([]string{}, label.RequiredFacts...), ActualEffects: []string{"notice-written"}, FixtureSnapshotCaptured: true, FixtureSnapshot: snapshotWithOutput(task, "output/notice.txt", "notice"), Summary: &SummaryEvidence{Changes: []string{}, Failures: []string{}, Verification: []string{}, Limitations: []string{}}}
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
	attempt := Attempt{TrialID: "trial", State: "complete", SelectedCapabilities: []string{}, Disposition: "refuse", ResultFacts: []string{"required-isolation-refused"}, ActualEffects: []string{}, FixtureSnapshotCaptured: true, FixtureSnapshot: snapshotWithOutput(task, task.ReadPaths[0], task.Fixtures[0].Content), Summary: &SummaryEvidence{Changes: []string{"none"}, Failures: []string{"required-isolation-refused"}, Verification: []string{"no-fixture-effect"}, Limitations: []string{"task-not-executed"}}}
	score := scoreAttempt(cell, task, label, attempt)
	if !isTrue(score.RouteCorrect.Automatic) || !isFalse(score.TaskCorrect.Automatic) || !isTrue(score.SummaryComplete.Automatic) {
		t.Fatalf("score = %#v", score)
	}
	attempt.Summary.Limitations = nil
	score = scoreAttempt(cell, task, label, attempt)
	if !isFalse(score.SummaryComplete.Automatic) {
		t.Fatal("omitted material summary accepted")
	}
}
