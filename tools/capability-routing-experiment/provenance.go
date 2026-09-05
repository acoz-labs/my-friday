package main

import (
	"errors"
	"runtime/debug"
)

func currentRunnerProvenance() RunnerProvenance {
	result := RunnerProvenance{Reason: "Go build information did not provide a VCS revision"}
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return result
	}
	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			result.Revision = setting.Value
		case "vcs.modified":
			result.Modified = setting.Value == "true"
		}
	}
	if commitPattern.MatchString(result.Revision) {
		result.Available = true
		result.Reason = ""
		if result.Modified {
			result.Reason = "Go build information reports modified VCS state"
		}
	}
	return result
}

func validateRunnerProvenance(result RunnerProvenance) error {
	if result.Available {
		if !commitPattern.MatchString(result.Revision) {
			return errors.New("available runner provenance lacks a full VCS revision")
		}
		if result.Modified && result.Reason == "" {
			return errors.New("modified runner provenance lacks an explicit reason")
		}
		if !result.Modified && result.Reason != "" {
			return errors.New("clean runner provenance carries a contradictory reason")
		}
		return nil
	}
	if result.Revision != "" || result.Reason == "" {
		return errors.New("unavailable runner provenance must omit revision and state a reason")
	}
	return nil
}
