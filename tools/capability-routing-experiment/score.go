package main

import (
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
	TrialID              string         `json:"trial_id"`
	HarnessID            string         `json:"harness_id"`
	Mode                 string         `json:"mode"`
	TaskID               string         `json:"task_id"`
	Split                string         `json:"split"`
	Category             string         `json:"category"`
	Repetition           int            `json:"repetition"`
	State                string         `json:"state"`
	Reason               string         `json:"reason"`
	RouteCorrect         DimensionScore `json:"route_correct"`
	TaskCorrect          DimensionScore `json:"task_correct"`
	PolicyPreserved      DimensionScore `json:"policy_preserved"`
	SummaryComplete      DimensionScore `json:"summary_complete"`
	WallMillis           *int64         `json:"wall_millis"`
	Telemetry            *Telemetry     `json:"telemetry"`
	TelemetryComplete    bool           `json:"telemetry_complete"`
	ContextClaimEligible bool           `json:"context_claim_eligible"`
}

type Coverage struct {
	HarnessID         string `json:"harness_id"`
	Mode              string `json:"mode"`
	Split             string `json:"split"`
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
	Version        int                     `json:"version"`
	ManifestSHA256 string                  `json:"manifest_sha256"`
	CorpusRevision string                  `json:"corpus_revision"`
	Scores         []TrialScore            `json:"scores"`
	Coverage       []Coverage              `json:"coverage"`
	Performance    []PerformanceComparison `json:"performance"`
	Recommendation string                  `json:"recommendation"`
	Conclusion     string                  `json:"conclusion"`
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
	labels := map[string]Label{}
	tasks := map[string]Task{}
	for _, label := range bundle.Labels.Labels {
		labels[label.TaskID] = label
	}
	for _, task := range bundle.Tasks.Tasks {
		tasks[task.ID] = task
	}
	primary := map[string]Attempt{}
	seenAttempts := map[string]bool{}
	for _, attempt := range attempts.Attempts {
		if attempt.AttemptID == "" || seenAttempts[attempt.AttemptID] {
			return Comparison{}, fmt.Errorf("missing or duplicate attempt id %q", attempt.AttemptID)
		}
		seenAttempts[attempt.AttemptID] = true
		if !contains([]string{"complete", "failed", "unavailable", "invalid"}, attempt.State) {
			return Comparison{}, fmt.Errorf("attempt %s has invalid state %q", attempt.AttemptID, attempt.State)
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
	comparison := Comparison{Version: SchemaVersion, ManifestSHA256: manifestDigest, CorpusRevision: bundle.Manifest.CorpusRevision}
	coverage := map[string]*Coverage{}
	for _, cell := range bundle.Manifest.Cells {
		task := tasks[cell.TaskID]
		key := fmt.Sprintf("%s\x00%s\x00%s\x00%d", cell.HarnessID, cell.Mode, task.Split, cell.Repetition)
		group := coverage[key]
		if group == nil {
			group = &Coverage{HarnessID: cell.HarnessID, Mode: cell.Mode, Split: task.Split, Repetition: cell.Repetition}
			coverage[key] = group
		}
		group.Declared++
		attempt, ok := primary[cell.TrialID]
		if !ok {
			group.Missing++
			comparison.Scores = append(comparison.Scores, TrialScore{TrialID: cell.TrialID, HarnessID: cell.HarnessID, Mode: cell.Mode, TaskID: cell.TaskID, Split: task.Split, Category: task.Category, Repetition: cell.Repetition, State: "missing", Reason: "no primary attempt was recorded"})
			continue
		}
		score := scoreAttempt(cell, task, labels[cell.TaskID], attempt)
		comparison.Scores = append(comparison.Scores, score)
		switch attempt.State {
		case "complete":
			group.Complete++
		case "failed":
			group.Failed++
		case "unavailable":
			group.Unavailable++
		case "invalid":
			group.Invalid++
		}
		if score.TelemetryComplete {
			group.TelemetryComplete++
		}
		if score.WallMillis != nil {
			group.WallComplete++
		}
		if score.Telemetry != nil && score.Telemetry.PeakRootRequestInput.Complete {
			group.PeakInputComplete++
		}
		if score.Telemetry != nil && score.Telemetry.ActualWindowOccupancy.Complete {
			group.WindowComplete++
		}
		if isTrue(score.RouteCorrect.Automatic) {
			group.RouteCorrect++
		}
		if isTrue(score.TaskCorrect.Automatic) {
			group.TaskCorrect++
		}
		if isTrue(score.PolicyPreserved.Automatic) {
			group.PolicyPreserved++
		}
		if isTrue(score.SummaryComplete.Automatic) {
			group.SummaryComplete++
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
			for _, candidate := range scores {
				if candidate.HarnessID != harness || candidate.Mode != candidateMode || candidate.Split != "held-out" {
					continue
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
			result = append(result, PerformanceComparison{HarnessID: harness, CandidateMode: candidateMode, AggregateTokens: summarizePairs("aggregate input plus output", "tokens", 24, tokens), WallLatency: summarizePairs("wall latency", "milliseconds", 24, wall), PeakRootInput: summarizePairs("peak root per-request input on complex tasks", "tokens", 2, peak)})
		}
	}
	return result
}

func performanceKey(harness, mode, task string, repetition int) string {
	return fmt.Sprintf("%s\x00%s\x00%s\x00%d", harness, mode, task, repetition)
}

type metricValue func(TrialScore) (float64, bool)

func pairedValues(baseline, candidate TrialScore, metric metricValue) (float64, float64, bool) {
	if baseline.State != "complete" || candidate.State != "complete" || !isTrue(baseline.TaskCorrect.Automatic) || !isTrue(candidate.TaskCorrect.Automatic) {
		return 0, 0, false
	}
	left, ok1 := metric(baseline)
	right, ok2 := metric(candidate)
	return left, right, ok1 && ok2 && left > 0
}
func aggregateTokens(score TrialScore) (float64, bool) {
	if score.Telemetry == nil || !score.Telemetry.TotalInput.Complete || !score.Telemetry.TotalOutput.Complete {
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
	if score.Telemetry == nil || !score.Telemetry.PeakRootRequestInput.Complete {
		return 0, false
	}
	return float64(*score.Telemetry.PeakRootRequestInput.Value), true
}
func newPair(score TrialScore, baseline, candidate float64) PairDifference {
	return PairDifference{TaskID: score.TaskID, Repetition: score.Repetition, Baseline: baseline, Candidate: candidate, Difference: candidate - baseline, Ratio: candidate / baseline}
}
func summarizePairs(name, unit string, required int, pairs []PairDifference) PairedMetric {
	metric := PairedMetric{Name: name, Unit: unit, EligiblePairs: len(pairs), RequiredPairs: required, Pairs: pairs}
	if len(pairs) == 0 {
		metric.MissingReason = "no matched correct completed-work cells with the required metric"
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
	score := TrialScore{TrialID: cell.TrialID, HarnessID: cell.HarnessID, Mode: cell.Mode, TaskID: cell.TaskID, Split: task.Split, Category: task.Category, Repetition: cell.Repetition, State: attempt.State, Reason: attempt.Reason, WallMillis: attempt.WallMillis, Telemetry: attempt.Telemetry}
	if score.WallMillis != nil && *score.WallMillis <= 0 {
		score.WallMillis = nil
	}
	if attempt.State != "complete" {
		return score
	}
	expected := label.Expectation
	allowed := label.AllowedCapabilitySets
	requiredFacts := label.RequiredFacts
	requiredEffects := label.RequiredEffects
	requiredSummary := label.RequiredSummary
	if task.RequiresIsolation && cell.Mode == "lookup-direct" {
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
	observedEffects, diffValid := fixtureEffects(attempt.FixtureDiff)
	allEffects := append(append([]string{}, attempt.ActualEffects...), observedEffects...)
	taskCorrect := containsAll(attempt.ResultFacts, requiredFacts) && containsAll(observedEffects, requiredEffects) && diffValid && nonePresent(allEffects, label.ForbiddenEffects) && attempt.Disposition == expected
	score.TaskCorrect = automatic(taskCorrect, "required facts, expected disposition, and actual fixture effects compared with frozen label")
	policy := !attempt.PolicyLoss && diffValid && nonePresent(allEffects, label.ForbiddenEffects)
	if attempt.Telemetry != nil && attempt.Telemetry.PolicyLoss {
		policy = false
	}
	score.PolicyPreserved = automatic(policy, "policy-loss event and forbidden actual effects checked")
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

func fixtureEffects(diff []FixtureEffect) ([]string, bool) {
	result := make([]string, 0, len(diff))
	for _, effect := range diff {
		if effect.Effect != "write" || !safeFixturePath(effect.Path) || !sha256Pattern.MatchString(effect.BeforeSHA256) || !sha256Pattern.MatchString(effect.AfterSHA256) || effect.BeforeSHA256 == effect.AfterSHA256 {
			return nil, false
		}
		result = append(result, effect.Effect+":"+effect.Path)
	}
	return result, true
}

func decideRecommendation(comparison Comparison) (string, string) {
	for _, score := range comparison.Scores {
		if isFalse(score.PolicyPreserved.Automatic) {
			return "retain-native-baseline", "A critical policy-preservation failure prevents adoption of a candidate routing mode."
		}
		if score.Split == "held-out" && (score.State != "complete" || !score.TelemetryComplete || score.WallMillis == nil) {
			return "inconclusive", "At least one held-out harness/mode cell is incomplete or lacks required telemetry; retain the native baseline and make no cross-harness claim."
		}
	}
	var passing []string
	for _, mode := range []string{"lookup-direct", "lookup-worker"} {
		if candidatePasses(comparison, mode) {
			passing = append(passing, mode)
		}
	}
	if len(passing) == 0 {
		return "retain-native-baseline", "All required cells completed, but no candidate met the frozen correctness, summary, token, peak-root-input, and wall-latency thresholds in both harnesses."
	}
	return strings.Join(passing, "+"), "The named candidate mode(s) met the frozen bounded thresholds in both harnesses; independent maintainer adjudication remains required before using this directional result in later solution design."
}

func candidatePasses(comparison Comparison, mode string) bool {
	performance := map[string]PerformanceComparison{}
	for _, item := range comparison.Performance {
		if item.CandidateMode == mode {
			performance[item.HarnessID] = item
		}
	}
	for _, harness := range []string{"claude", "codex"} {
		candidate, baseline := scoreCounts(comparison.Scores, harness, mode), scoreCounts(comparison.Scores, harness, "native-catalogue")
		if candidate.total != 24 || baseline.total != 24 || candidate.task < 22 || candidate.summary != 24 || candidate.route < baseline.route || candidate.task < baseline.task || candidate.summary < baseline.summary {
			return false
		}
		item, ok := performance[harness]
		if !ok || !metricAtMost(item.AggregateTokens, 24, 1.25) || !metricAtMost(item.WallLatency, 24, 1.5) || !metricAtMost(item.PeakRootInput, 2, .8) {
			return false
		}
	}
	return true
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
func metricAtMost(metric PairedMetric, required int, threshold float64) bool {
	return metric.EligiblePairs == required && metric.MedianRatio != nil && *metric.MedianRatio <= threshold
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
