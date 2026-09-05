package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

func RenderMarkdown(comparison Comparison, probes []DriverProbe) string {
	var output strings.Builder
	fmt.Fprintf(&output, "# Capability routing comparison\n\n")
	fmt.Fprintf(&output, "- Corpus revision: `%s`\n", comparison.CorpusRevision)
	fmt.Fprintf(&output, "- Manifest SHA-256: `%s`\n", comparison.ManifestSHA256)
	fmt.Fprintf(&output, "- Recommendation: **%s**\n", comparison.Recommendation)
	fmt.Fprintf(&output, "- Conclusion: %s\n\n", comparison.Conclusion)
	output.WriteString("## Native-driver fidelity\n\n")
	output.WriteString("| Harness | Version | State | Missing controls |\n| --- | --- | --- | --- |\n")
	sort.Slice(probes, func(i, j int) bool { return probes[i].Harness < probes[j].Harness })
	for _, probe := range probes {
		fmt.Fprintf(&output, "| %s | %s | %s | %s |\n", probe.Harness, escapeCell(probe.CLIRevision), probe.State, escapeCell(strings.Join(probe.Unavailable, "; ")))
	}
	output.WriteString("\n## Coverage and scores\n\n")
	output.WriteString("| Harness | Mode | Split | Rep | Declared | Complete | Failed | Unavailable | Invalid | Missing | Full tokens | Wall | Peak root input | Window occupancy | Route | Task | Policy | Summary |\n")
	output.WriteString("| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |\n")
	for _, row := range comparison.Coverage {
		fmt.Fprintf(&output, "| %s | %s | %s | %d | %d | %d | %d | %d | %d | %d | %d | %d | %d | %d | %d | %d | %d | %d |\n", row.HarnessID, row.Mode, row.Split, row.Repetition, row.Declared, row.Complete, row.Failed, row.Unavailable, row.Invalid, row.Missing, row.TelemetryComplete, row.WallComplete, row.PeakInputComplete, row.WindowComplete, row.RouteCorrect, row.TaskCorrect, row.PolicyPreserved, row.SummaryComplete)
	}
	output.WriteString("\n## Paired performance\n\n")
	output.WriteString("| Harness | Candidate | Metric | Coverage | Median ratio | Difference range | Missing reason |\n| --- | --- | --- | --- | ---: | --- | --- |\n")
	for _, item := range comparison.Performance {
		for _, metric := range []PairedMetric{item.AggregateTokens, item.WallLatency, item.PeakRootInput} {
			fmt.Fprintf(&output, "| %s | %s | %s | %d/%d | %s | %s | %s |\n", item.HarnessID, item.CandidateMode, metric.Name, metric.EligiblePairs, metric.RequiredPairs, formatFloat(metric.MedianRatio), formatRange(metric.MinDifference, metric.MaxDifference, metric.Unit), escapeCell(metric.MissingReason))
		}
	}
	output.WriteString("\nToken, wall-latency, peak-root-input, and actual-window columns are completeness counts over the declared denominator. Their per-trial values and missing reasons are in `report.json`; no missing value is represented as zero. Unavailable, failed, invalid, and missing cells are never treated as cheaper completed work. Null score values mean the cell was not eligible for scoring. Peak root request input is not described as actual context occupancy unless the harness reports that stronger metric separately.\n")
	return output.String()
}

func MarshalReport(comparison Comparison, probes []DriverProbe) ([]byte, error) {
	value := struct {
		Version    int           `json:"version"`
		Comparison Comparison    `json:"comparison"`
		Probes     []DriverProbe `json:"driver_probes"`
	}{Version: SchemaVersion, Comparison: comparison, Probes: probes}
	return json.MarshalIndent(value, "", "  ")
}

func escapeCell(value string) string { return strings.ReplaceAll(value, "|", "\\|") }
func formatFloat(value *float64) string {
	if value == nil {
		return "null"
	}
	return fmt.Sprintf("%.3f", *value)
}
func formatRange(min, max *float64, unit string) string {
	if min == nil || max == nil {
		return "null"
	}
	return fmt.Sprintf("%.0f to %.0f %s", *min, *max, unit)
}
