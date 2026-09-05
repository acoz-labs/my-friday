package main

import "testing"

func TestRunnerProvenanceValidation(t *testing.T) {
	for name, candidate := range map[string]RunnerProvenance{
		"missing-reason":       {},
		"bad-revision":         {Revision: "short", Available: true},
		"dirty-without-reason": {Revision: PreregistrationBasisCommit, Modified: true, Available: true},
		"clean-with-reason":    {Revision: PreregistrationBasisCommit, Available: true, Reason: "contradiction"},
	} {
		t.Run(name, func(t *testing.T) {
			if err := validateRunnerProvenance(candidate); err == nil {
				t.Fatal("invalid runner provenance accepted")
			}
		})
	}
	if err := validateRunnerProvenance(RunnerProvenance{Revision: PreregistrationBasisCommit, Available: true}); err != nil {
		t.Fatal(err)
	}
	if err := validateRunnerProvenance(RunnerProvenance{Reason: "unavailable"}); err != nil {
		t.Fatal(err)
	}
}
