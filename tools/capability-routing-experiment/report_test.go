package main

import "testing"

func TestReportInputsAreRecomputedAndCrossBound(t *testing.T) {
	bundle := loadTestBundle(t)
	manifest, err := PrepareManifest(bundle, TrustedSourceCommit, TrustedHarnesses())
	if err != nil {
		t.Fatal(err)
	}
	bundle.Manifest = manifest
	attempts := unavailableAttempts(manifest)
	comparison, err := ScoreAttempts(bundle, attempts)
	if err != nil {
		t.Fatal(err)
	}
	probes := unavailableProbes(manifest)
	if err = ValidateReportInputs(bundle, attempts, comparison, probes); err != nil {
		t.Fatal(err)
	}

	t.Run("tampered-score", func(t *testing.T) {
		candidate := comparison
		candidate.Recommendation = "lookup-direct"
		if err := ValidateReportInputs(bundle, attempts, candidate, probes); err == nil {
			t.Fatal("tampered score accepted")
		}
	})
	t.Run("stale-probe", func(t *testing.T) {
		candidate := append([]DriverProbe{}, probes...)
		candidate[0].CLIRevision = "other"
		if err := ValidateReportInputs(bundle, attempts, comparison, candidate); err == nil {
			t.Fatal("stale probe accepted")
		}
	})
	t.Run("duplicate-probe", func(t *testing.T) {
		candidate := append(append([]DriverProbe{}, probes...), probes[0])
		if err := ValidateReportInputs(bundle, attempts, comparison, candidate); err == nil {
			t.Fatal("duplicate probe accepted")
		}
	})
}

func TestCategoryCoveragePreservesDeclaredDenominators(t *testing.T) {
	bundle := loadTestBundle(t)
	manifest, err := PrepareManifest(bundle, TrustedSourceCommit, TrustedHarnesses())
	if err != nil {
		t.Fatal(err)
	}
	bundle.Manifest = manifest
	comparison, err := ScoreAttempts(bundle, unavailableAttempts(manifest))
	if err != nil {
		t.Fatal(err)
	}
	if len(comparison.CategoryCoverage) != 288 {
		t.Fatalf("category coverage rows=%d", len(comparison.CategoryCoverage))
	}
	for _, row := range comparison.CategoryCoverage {
		if row.Category == "" || row.Declared != 1 || row.Unavailable != 1 {
			t.Fatalf("row=%#v", row)
		}
	}
}

func unavailableProbes(manifest Manifest) []DriverProbe {
	result := make([]DriverProbe, 0, len(manifest.Harnesses))
	for _, harness := range manifest.Harnesses {
		commands := []string{harness.ID + " --version", harness.ID + " --help"}
		if harness.ID == "codex" {
			commands[1] = "codex exec --help"
		}
		result = append(result, DriverProbe{Version: SchemaVersion, Harness: HarnessName(harness.ID), Executable: harness.ID, CLIRevision: harness.ExecutableVersion, State: "unavailable", Unavailable: []string{"controls unproven"}, ProbedCommands: commands})
	}
	return result
}
