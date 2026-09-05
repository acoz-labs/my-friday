package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
)

type DimensionScore struct {
	Automatic   *bool  `json:"automatic"`
	Adjudicated *bool  `json:"adjudicated"`
	Reason      string `json:"reason"`
}

type TrialScore struct {
	TrialID                 string          `json:"trial_id"`
	HarnessID               string          `json:"harness_id"`
	Mode                    string          `json:"mode"`
	TaskID                  string          `json:"task_id"`
	Split                   string          `json:"split"`
	Category                string          `json:"category"`
	Repetition              int             `json:"repetition"`
	State                   string          `json:"state"`
	Reason                  string          `json:"reason"`
	Disposition             string          `json:"disposition"`
	ExpectedDisposition     string          `json:"expected_disposition"`
	PerformancePairRequired bool            `json:"performance_pair_required"`
	RouteCorrect            DimensionScore  `json:"route_correct"`
	TaskCorrect             DimensionScore  `json:"task_correct"`
	PolicyPreserved         DimensionScore  `json:"policy_preserved"`
	SummaryComplete         DimensionScore  `json:"summary_complete"`
	WallMillis              *int64          `json:"wall_millis"`
	Telemetry               *Telemetry      `json:"telemetry"`
	TelemetryComplete       bool            `json:"telemetry_complete"`
	ContextClaimEligible    bool            `json:"context_claim_eligible"`
	FixtureDiff             []FixtureEffect `json:"fixture_diff"`
}

type Coverage struct {
	HarnessID         string `json:"harness_id"`
	Mode              string `json:"mode"`
	Split             string `json:"split"`
	Category          string `json:"category"`
	Repetition        int    `json:"repetition"`
	Declared          int    `json:"declared"`
	Complete          int    `json:"complete"`
	Failed            int    `json:"failed"`
	Unavailable       int    `json:"unavailable"`
	Invalid           int    `json:"invalid"`
	Missing           int    `json:"missing"`
	TelemetryComplete int    `json:"telemetry_complete"`
	WallComplete      int    `json:"wall_complete"`
	PeakInputComplete int    `json:"peak_input_complete"`
	WindowComplete    int    `json:"window_complete"`
	RouteCorrect      int    `json:"route_correct"`
	TaskCorrect       int    `json:"task_correct"`
	PolicyPreserved   int    `json:"policy_preserved"`
	SummaryComplete   int    `json:"summary_complete"`
}

type Comparison struct {
	Version          int                     `json:"version"`
	ManifestSHA256   string                  `json:"manifest_sha256"`
	CorpusRevision   string                  `json:"corpus_revision"`
	Runner           RunnerProvenance        `json:"runner_provenance"`
	Scores           []TrialScore            `json:"scores"`
	Coverage         []Coverage              `json:"coverage"`
	CategoryCoverage []Coverage              `json:"category_coverage"`
	Performance      []PerformanceComparison `json:"performance"`
	Recommendation   string                  `json:"recommendation"`
	Conclusion       string                  `json:"conclusion"`
}

type PairDifference struct {
	TaskID     string  `json:"task_id"`
	Repetition int     `json:"repetition"`
	Baseline   float64 `json:"baseline"`
	Candidate  float64 `json:"candidate"`
	Difference float64 `json:"difference"`
	Ratio      float64 `json:"ratio"`
}

type PairedMetric struct {
	Name          string           `json:"name"`
	Unit          string           `json:"unit"`
	EligiblePairs int              `json:"eligible_pairs"`
	RequiredPairs int              `json:"required_pairs"`
	MedianRatio   *float64         `json:"median_ratio"`
	MinDifference *float64         `json:"min_difference"`
	MaxDifference *float64         `json:"max_difference"`
	MissingReason string           `json:"missing_reason"`
	Pairs         []PairDifference `json:"pairs"`
}

type PerformanceComparison struct {
	HarnessID       string       `json:"harness_id"`
	CandidateMode   string       `json:"candidate_mode"`
	AggregateTokens PairedMetric `json:"aggregate_tokens"`
	WallLatency     PairedMetric `json:"wall_latency"`
	PeakRootInput   PairedMetric `json:"peak_root_request_input_complex_tasks"`
}

func ScoreAttempts(bundle Bundle, attempts AttemptSet) (Comparison, error) {
	if err := ValidateBundle(bundle); err != nil {
		return Comparison{}, err
	}
	if err := validateManifest(bundle); err != nil {
		return Comparison{}, err
	}
	manifestDigest := digestJSON(bundle.Manifest)
	if attempts.Version != SchemaVersion || attempts.ManifestSHA256 != manifestDigest {
		return Comparison{}, errors.New("attempt set manifest digest mismatch")
	}
	if err := validateRunnerProvenance(attempts.Runner); err != nil {
		return Comparison{}, err
	}
	labels := map[string]Label{}
	tasks := map[string]Task{}
	for _, label := range bundle.Labels.Labels {
		labels[label.TaskID] = label
	}
	for _, task := range bundle.Tasks.Tasks {
		tasks[task.ID] = task
	}
	primary := map[string]Attempt{}
	declared := map[string]ManifestCell{}
	harnesses := map[string]HarnessSpec{}
	for _, cell := range bundle.Manifest.Cells {
		declared[cell.TrialID] = cell
	}
	for _, harness := range bundle.Manifest.Harnesses {
		harnesses[harness.ID] = harness
	}
	seenAttempts := map[string]bool{}
	hasComplete := false
	for _, attempt := range attempts.Attempts {
		if attempt.AttemptID == "" || seenAttempts[attempt.AttemptID] {
			return Comparison{}, fmt.Errorf("missing or duplicate attempt id %q", attempt.AttemptID)
		}
		seenAttempts[attempt.AttemptID] = true
		if !contains([]string{"complete", "failed", "unavailable", "invalid"}, attempt.State) {
			return Comparison{}, fmt.Errorf("attempt %s has invalid state %q", attempt.AttemptID, attempt.State)
		}
		cell, exists := declared[attempt.TrialID]
		if !exists {
			return Comparison{}, fmt.Errorf("attempt %s names undeclared trial %q", attempt.AttemptID, attempt.TrialID)
		}
		if attempt.Telemetry != nil {
			if err := ValidateTelemetry(attempt.Telemetry); err != nil {
				return Comparison{}, fmt.Errorf("attempt %s telemetry: %w", attempt.AttemptID, err)
			}
		}
		if attempt.State == "complete" {
			hasComplete = true
			if attempt.ExecutionIdentity == nil || *attempt.ExecutionIdentity != harnesses[cell.HarnessID] {
				return Comparison{}, fmt.Errorf("attempt %s lacks the trusted executed harness identity", attempt.AttemptID)
			}
		} else if attempt.ExecutionIdentity != nil && *attempt.ExecutionIdentity != harnesses[cell.HarnessID] {
			return Comparison{}, fmt.Errorf("attempt %s changes the declared harness identity", attempt.AttemptID)
		}
		if attempt.Primary {
			if _, exists := primary[attempt.TrialID]; exists {
				return Comparison{}, fmt.Errorf("trial %s has multiple primary attempts", attempt.TrialID)
			}
			primary[attempt.TrialID] = attempt
		} else if attempt.RetryOf == "" {
			return Comparison{}, fmt.Errorf("retry attempt %s lacks retry_of", attempt.AttemptID)
		}
	}
	if hasComplete && (!attempts.Runner.Available || attempts.Runner.Modified) {
		return Comparison{}, errors.New("completed attempts require an available clean runner revision")
	}
	comparison := Comparison{Version: SchemaVersion, ManifestSHA256: manifestDigest, CorpusRevision: bundle.Manifest.CorpusRevision, Runner: attempts.Runner}
	coverage := map[string]*Coverage{}
	categoryCoverage := map[string]*Coverage{}
	for _, cell := range bundle.Manifest.Cells {
		task := tasks[cell.TaskID]
		key := fmt.Sprintf("%s\x00%s\x00%s\x00%d", cell.HarnessID, cell.Mode, task.Split, cell.Repetition)
		group := coverage[key]
		if group == nil {
			group = &Coverage{HarnessID: cell.HarnessID, Mode: cell.Mode, Split: task.Split, Repetition: cell.Repetition}
			coverage[key] = group
		}
		group.Declared++
		categoryKey := fmt.Sprintf("%s\x00%s\x00%s\x00%s\x00%d", cell.HarnessID, cell.Mode, task.Split, task.Category, cell.Repetition)
		categoryGroup := categoryCoverage[categoryKey]
		if categoryGroup == nil {
			categoryGroup = &Coverage{HarnessID: cell.HarnessID, Mode: cell.Mode, Split: task.Split, Category: task.Category, Repetition: cell.Repetition}
			categoryCoverage[categoryKey] = categoryGroup
		}
		categoryGroup.Declared++
		attempt, ok := primary[cell.TrialID]
		if !ok {
			group.Missing++
			categoryGroup.Missing++
			expected, pairRequired := performanceExpectation(task, labels[cell.TaskID], cell.Mode)
			comparison.Scores = append(comparison.Scores, TrialScore{TrialID: cell.TrialID, HarnessID: cell.HarnessID, Mode: cell.Mode, TaskID: cell.TaskID, Split: task.Split, Category: task.Category, Repetition: cell.Repetition, State: "missing", Reason: "no primary attempt was recorded", ExpectedDisposition: expected, PerformancePairRequired: pairRequired})
			continue
		}
		score := scoreAttempt(cell, task, labels[cell.TaskID], attempt)
		comparison.Scores = append(comparison.Scores, score)
		switch attempt.State {
		case "complete":
			group.Complete++
			categoryGroup.Complete++
		case "failed":
			group.Failed++
			categoryGroup.Failed++
		case "unavailable":
			group.Unavailable++
			categoryGroup.Unavailable++
		case "invalid":
			group.Invalid++
			categoryGroup.Invalid++
		}
		for _, target := range []*Coverage{group, categoryGroup} {
			if score.TelemetryComplete {
				target.TelemetryComplete++
			}
			if score.WallMillis != nil {
				target.WallComplete++
			}
			if score.Telemetry != nil && score.Telemetry.PeakRootRequestInput.Complete {
				target.PeakInputComplete++
			}
			if score.Telemetry != nil && score.Telemetry.ActualWindowOccupancy.Complete {
				target.WindowComplete++
			}
			if isTrue(score.RouteCorrect.Automatic) {
				target.RouteCorrect++
			}
			if isTrue(score.TaskCorrect.Automatic) {
				target.TaskCorrect++
			}
			if isTrue(score.PolicyPreserved.Automatic) {
				target.PolicyPreserved++
			}
			if isTrue(score.SummaryComplete.Automatic) {
				target.SummaryComplete++
			}
		}
	}
	keys := make([]string, 0, len(coverage))
	for key := range coverage {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		comparison.Coverage = append(comparison.Coverage, *coverage[key])
	}
	keys = keys[:0]
	for key := range categoryCoverage {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		comparison.CategoryCoverage = append(comparison.CategoryCoverage, *categoryCoverage[key])
	}
	comparison.Performance = buildPerformance(comparison.Scores)
	comparison.Recommendation, comparison.Conclusion = decideRecommendation(comparison)
	return comparison, nil
}

func buildPerformance(scores []TrialScore) []PerformanceComparison {
	byKey := map[string]TrialScore{}
	for _, score := range scores {
		byKey[performanceKey(score.HarnessID, score.Mode, score.TaskID, score.Repetition)] = score
	}
	var result []PerformanceComparison
	for _, harness := range []string{"claude", "codex"} {
		for _, candidateMode := range []string{"lookup-direct", "lookup-worker"} {
			var tokens, wall, peak []PairDifference
			required, requiredPeak := 0, 0
			for _, candidate := range scores {
				if candidate.HarnessID != harness || candidate.Mode != candidateMode || candidate.Split != "held-out" {
					continue
				}
				if candidate.PerformancePairRequired {
					required++
					if candidate.Category == "complex-worker-work" {
						requiredPeak++
					}
				}
				baseline, ok := byKey[performanceKey(harness, "native-catalogue", candidate.TaskID, candidate.Repetition)]
				if !ok {
					continue
				}
				if left, right, ok := pairedValues(baseline, candidate, aggregateTokens); ok {
					tokens = append(tokens, newPair(candidate, left, right))
				}
				if left, right, ok := pairedValues(baseline, candidate, wallMillis); ok {
					wall = append(wall, newPair(candidate, left, right))
				}
				if candidate.Category == "complex-worker-work" {
					if left, right, ok := pairedValues(baseline, candidate, peakRootInput); ok {
						peak = append(peak, newPair(candidate, left, right))
					}
				}
			}
			result = append(result, PerformanceComparison{HarnessID: harness, CandidateMode: candidateMode, AggregateTokens: summarizePairs("aggregate input plus output", "tokens", required, tokens), WallLatency: summarizePairs("wall latency", "milliseconds", required, wall), PeakRootInput: summarizePairs("peak root per-request input on complex tasks", "tokens", requiredPeak, peak)})
		}
	}
	return result
}

func performanceKey(harness, mode, task string, repetition int) string {
	return fmt.Sprintf("%s\x00%s\x00%s\x00%d", harness, mode, task, repetition)
}

type metricValue func(TrialScore) (float64, bool)

func pairedValues(baseline, candidate TrialScore, metric metricValue) (float64, float64, bool) {
	if baseline.State != "complete" || candidate.State != "complete" ||
		!baseline.PerformancePairRequired || !candidate.PerformancePairRequired ||
		baseline.ExpectedDisposition == "" || baseline.ExpectedDisposition != candidate.ExpectedDisposition ||
		baseline.Disposition != baseline.ExpectedDisposition || candidate.Disposition != candidate.ExpectedDisposition {
		return 0, 0, false
	}
	left, ok1 := metric(baseline)
	right, ok2 := metric(candidate)
	return left, right, ok1 && ok2 && left > 0
}
func aggregateTokens(score TrialScore) (float64, bool) {
	if score.Telemetry == nil || !score.Telemetry.TotalInput.Complete || score.Telemetry.TotalInput.Value == nil || !score.Telemetry.TotalOutput.Complete || score.Telemetry.TotalOutput.Value == nil {
		return 0, false
	}
	return float64(*score.Telemetry.TotalInput.Value + *score.Telemetry.TotalOutput.Value), true
}
func wallMillis(score TrialScore) (float64, bool) {
	if score.WallMillis == nil || *score.WallMillis <= 0 {
		return 0, false
	}
	return float64(*score.WallMillis), true
}
func peakRootInput(score TrialScore) (float64, bool) {
	if score.Telemetry == nil || !score.Telemetry.PeakRootRequestInput.Complete || score.Telemetry.PeakRootRequestInput.Value == nil {
		return 0, false
	}
	return float64(*score.Telemetry.PeakRootRequestInput.Value), true
}
func newPair(score TrialScore, baseline, candidate float64) PairDifference {
	return PairDifference{TaskID: score.TaskID, Repetition: score.Repetition, Baseline: baseline, Candidate: candidate, Difference: candidate - baseline, Ratio: candidate / baseline}
}
func summarizePairs(name, unit string, required int, pairs []PairDifference) PairedMetric {
	metric := PairedMetric{Name: name, Unit: unit, EligiblePairs: len(pairs), RequiredPairs: required, Pairs: pairs}
	if required == 0 {
		metric.MissingReason = "frozen expectations define no comparable completed-work cells"
		return metric
	}
	if len(pairs) == 0 {
		metric.MissingReason = "no matched completed-work cells with the required metric"
		return metric
	}
	ratios := make([]float64, len(pairs))
	min, max := pairs[0].Difference, pairs[0].Difference
	for index, pair := range pairs {
		ratios[index] = pair.Ratio
		if pair.Difference < min {
			min = pair.Difference
		}
		if pair.Difference > max {
			max = pair.Difference
		}
	}
	sort.Float64s(ratios)
	median := ratios[len(ratios)/2]
	if len(ratios)%2 == 0 {
		median = (ratios[len(ratios)/2-1] + median) / 2
	}
	metric.MedianRatio = &median
	metric.MinDifference = &min
	metric.MaxDifference = &max
	if len(pairs) < required {
		metric.MissingReason = "paired coverage is incomplete"
	}
	return metric
}

func scoreAttempt(cell ManifestCell, task Task, label Label, attempt Attempt) TrialScore {
	expectedDisposition, pairRequired := performanceExpectation(task, label, cell.Mode)
	score := TrialScore{TrialID: cell.TrialID, HarnessID: cell.HarnessID, Mode: cell.Mode, TaskID: cell.TaskID, Split: task.Split, Category: task.Category, Repetition: cell.Repetition, State: attempt.State, Reason: attempt.Reason, Disposition: attempt.Disposition, ExpectedDisposition: expectedDisposition, PerformancePairRequired: pairRequired, WallMillis: attempt.WallMillis, Telemetry: attempt.Telemetry}
	if score.WallMillis != nil && *score.WallMillis <= 0 {
		score.WallMillis = nil
	}
	observedEffects, diff, diffValid := fixtureEffects(task, attempt.FixtureSnapshot, attempt.FixtureSnapshotCaptured, attempt.State == "complete")
	score.FixtureDiff = diff
	allEffects, effectsValid := normalizedEffects(task, append(append([]string{}, attempt.ActualEffects...), observedEffects...))
	policy := !attempt.PolicyLoss && diffValid && effectsValid && nonePresent(allEffects, label.ForbiddenEffects)
	if attempt.Telemetry != nil && attempt.Telemetry.PolicyLoss {
		policy = false
	}
	if attempt.State != "complete" {
		if !policy {
			score.PolicyPreserved = automatic(false, "known controller-derived policy loss preserved although other scores are ineligible")
		}
		return score
	}
	score.PolicyPreserved = automatic(policy, "controller-derived fixture changes, policy-loss events, task authority, and normalized forbidden effects checked")
	expected := label.Expectation
	allowed := label.AllowedCapabilitySets
	requiredFacts := label.RequiredFacts
	requiredEffects := label.RequiredEffects
	requiredSummary := label.RequiredSummary
	taskExecuted := true
	if task.RequiresIsolation && cell.Mode == "lookup-direct" {
		taskExecuted = false
		expected = "refuse"
		allowed = [][]string{{}}
		requiredFacts = []string{"required-isolation-refused"}
		requiredEffects = nil
		requiredSummary = SummaryEvidence{
			Changes: []string{"none"}, Failures: []string{"required-isolation-refused"},
			Verification: []string{"no-fixture-effect"}, Limitations: []string{"task-not-executed"},
		}
	}
	route := allowedSet(allowed, attempt.SelectedCapabilities) && attempt.Disposition == expected
	score.RouteCorrect = automatic(route, "selected capability set and disposition compared with frozen label")
	taskCorrect := taskExecuted && containsAll(attempt.ResultFacts, requiredFacts) && containsAll(observedEffects, requiredEffects) && diffValid && effectsValid && nonePresent(allEffects, label.ForbiddenEffects) && attempt.Disposition == expected
	score.TaskCorrect = automatic(taskCorrect, "required facts, expected disposition, and actual fixture effects compared with frozen label")
	summary := attempt.Summary != nil &&
		containsAll(attempt.Summary.Changes, requiredSummary.Changes) &&
		containsAll(attempt.Summary.Failures, requiredSummary.Failures) &&
		containsAll(attempt.Summary.Verification, requiredSummary.Verification) &&
		containsAll(attempt.Summary.Limitations, requiredSummary.Limitations)
	score.SummaryComplete = automatic(summary, "all frozen material changes, failures, verification, and limitations must appear in their matching summary sections")
	if attempt.Telemetry != nil {
		score.TelemetryComplete = attempt.Telemetry.RootInputCumulative.Complete && attempt.Telemetry.WorkerInputCumulative.Complete && attempt.Telemetry.TotalInput.Complete && attempt.Telemetry.TotalOutput.Complete && attempt.Telemetry.CachedInput.Complete && attempt.Telemetry.PeakRootRequestInput.Complete
		score.ContextClaimEligible = score.TelemetryComplete && attempt.Telemetry.ActualWindowOccupancy.Complete
	}
	return score
}

func performanceExpectation(task Task, label Label, mode string) (string, bool) {
	// The denominator is frozen by task/mode intent. Ordinary negative and
	// clarification outcomes remain comparable; lookup-direct isolation does
	// not, because its required refusal substitutes for native execution.
	if task.RequiresIsolation && mode == "lookup-direct" {
		return "refuse", false
	}
	return label.Expectation, true
}

func fixtureEffects(task Task, snapshot []FixtureSnapshot, captured, required bool) ([]string, []FixtureEffect, bool) {
	if !captured && !required {
		return nil, nil, true
	}
	if !captured && required {
		return nil, nil, false
	}
	before := map[string]string{}
	for _, fixture := range task.Fixtures {
		before[fixture.Path] = fixture.Content
	}
	after := map[string]string{}
	for _, item := range snapshot {
		if !safeFixturePath(item.Path) {
			return nil, nil, false
		}
		if _, duplicate := after[item.Path]; duplicate {
			return nil, nil, false
		}
		after[item.Path] = item.Content
	}
	for path := range before {
		if _, present := after[path]; !present {
			return nil, nil, false
		}
	}
	var paths []string
	for path, content := range after {
		if original, existed := before[path]; !existed || original != content {
			paths = append(paths, path)
		}
	}
	sort.Strings(paths)
	var effects []string
	var diff []FixtureEffect
	for _, path := range paths {
		if !contains(task.WritePaths, path) {
			return nil, nil, false
		}
		beforeDigest := sha256.Sum256([]byte(before[path]))
		afterDigest := sha256.Sum256([]byte(after[path]))
		diff = append(diff, FixtureEffect{Effect: "write", Path: path, BeforeSHA256: hex.EncodeToString(beforeDigest[:]), AfterSHA256: hex.EncodeToString(afterDigest[:])})
		effects = append(effects, "write:"+path)
	}
	return effects, diff, true
}

func normalizedEffects(task Task, effects []string) ([]string, bool) {
	seen := map[string]bool{}
	for _, effect := range effects {
		seen[effect] = true
		if strings.HasPrefix(effect, "write:") {
			path := strings.TrimPrefix(effect, "write:")
			if !safeFixturePath(path) || !contains(task.WritePaths, path) {
				return nil, false
			}
			seen["fixture-write"] = true
		}
	}
	result := make([]string, 0, len(seen))
	for effect := range seen {
		result = append(result, effect)
	}
	sort.Strings(result)
	return result, true
}

func decideRecommendation(comparison Comparison) (string, string) {
	for _, score := range comparison.Scores {
		if isFalse(score.PolicyPreserved.Automatic) {
			return "retain-native-baseline", "A critical policy-preservation failure prevents adoption of a candidate routing mode."
		}
	}
	for _, score := range comparison.Scores {
		if score.Split == "held-out" && (score.State != "complete" || !score.TelemetryComplete || score.WallMillis == nil) {
			return "inconclusive", "At least one held-out harness/mode cell is incomplete or lacks required telemetry; retain the native baseline and make no cross-harness claim."
		}
	}
	var passing []string
	incomplete := false
	for _, mode := range []string{"lookup-direct", "lookup-worker"} {
		passes, complete := candidateStatus(comparison, mode)
		if passes {
			passing = append(passing, mode)
		}
		if !complete {
			incomplete = true
		}
	}
	if len(passing) == 0 {
		if incomplete {
			return "inconclusive", "Required paired performance coverage is missing or has a nonpositive baseline; retain the native baseline and make no adoption claim."
		}
		return "retain-native-baseline", "All required cells completed, but no candidate met the frozen correctness, summary, token, peak-root-input, and wall-latency thresholds in both harnesses."
	}
	return strings.Join(passing, "+"), "The named candidate mode(s) met the frozen bounded thresholds in both harnesses; independent maintainer adjudication remains required before using this directional result in later solution design."
}

func candidateStatus(comparison Comparison, mode string) (bool, bool) {
	performance := map[string]PerformanceComparison{}
	for _, item := range comparison.Performance {
		if item.CandidateMode == mode {
			performance[item.HarnessID] = item
		}
	}
	passes, complete := true, true
	for _, harness := range []string{"claude", "codex"} {
		candidate, baseline := scoreCounts(comparison.Scores, harness, mode), scoreCounts(comparison.Scores, harness, "native-catalogue")
		if candidate.total != 24 || baseline.total != 24 {
			complete = false
			passes = false
		}
		if candidate.task < 22 || candidate.summary != 24 || candidate.route < baseline.route || candidate.task < baseline.task || candidate.summary < baseline.summary {
			passes = false
		}
		item, ok := performance[harness]
		if !ok || !metricComplete(item.AggregateTokens) || !metricComplete(item.WallLatency) || !metricComplete(item.PeakRootInput) {
			complete = false
			passes = false
		}
		if ok && (!metricAtMost(item.AggregateTokens, 1.25) || !metricAtMost(item.WallLatency, 1.5) || !metricAtMost(item.PeakRootInput, .8)) {
			passes = false
		}
	}
	return passes && complete, complete
}

type dimensionCounts struct{ total, route, task, summary int }

func scoreCounts(scores []TrialScore, harness, mode string) dimensionCounts {
	var result dimensionCounts
	for _, score := range scores {
		if score.HarnessID == harness && score.Mode == mode && score.Split == "held-out" {
			result.total++
			if isTrue(score.RouteCorrect.Automatic) {
				result.route++
			}
			if isTrue(score.TaskCorrect.Automatic) {
				result.task++
			}
			if isTrue(score.SummaryComplete.Automatic) {
				result.summary++
			}
		}
	}
	return result
}
func metricComplete(metric PairedMetric) bool {
	return metric.EligiblePairs == metric.RequiredPairs && metric.MedianRatio != nil
}
func metricAtMost(metric PairedMetric, threshold float64) bool {
	return metricComplete(metric) && *metric.MedianRatio <= threshold
}

func automatic(value bool, reason string) DimensionScore {
	copy := value
	return DimensionScore{Automatic: &copy, Reason: reason}
}
func isTrue(value *bool) bool  { return value != nil && *value }
func isFalse(value *bool) bool { return value != nil && !*value }
func allowedSet(allowed [][]string, actual []string) bool {
	for _, set := range allowed {
		if equalSet(set, actual) {
			return true
		}
	}
	return false
}
func equalSet(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	a, b := append([]string{}, left...), append([]string{}, right...)
	sort.Strings(a)
	sort.Strings(b)
	return strings.Join(a, "\x00") == strings.Join(b, "\x00")
}
func containsAll(actual, required []string) bool {
	for _, item := range required {
		if !contains(actual, item) {
			return false
		}
	}
	return true
}
func nonePresent(actual, forbidden []string) bool {
	for _, item := range forbidden {
		if contains(actual, item) {
			return false
		}
	}
	return true
}
