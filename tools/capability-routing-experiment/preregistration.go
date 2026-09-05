package main

import "reflect"

// TrustedSourceCommit is the reviewed runner/corpus freeze. It is compiled into
// the validator so a manifest cannot redefine its own identity.
const TrustedSourceCommit = "1219e0d8cf892fa02ba80c8b40911648a4c15b58"

func TrustedHarnesses() []HarnessSpec {
	return []HarnessSpec{
		{ID: "codex", ExecutableVersion: "codex-cli 0.153.4", Model: "gpt-5.3-codex", Config: "experiment-v1-unavailable-until-control-preflight"},
		{ID: "claude", ExecutableVersion: "2.1.193 (Claude Code)", Model: "claude-sonnet-4-6", Config: "experiment-v1-unavailable-until-control-preflight"},
	}
}

func validateTrustedIdentity(manifest Manifest) error {
	if manifest.SourceCommit != TrustedSourceCommit {
		return errUntrustedPreregistration("source commit")
	}
	if !reflect.DeepEqual(manifest.Harnesses, TrustedHarnesses()) {
		return errUntrustedPreregistration("harness version/model/config")
	}
	return nil
}

type errUntrustedPreregistration string

func (value errUntrustedPreregistration) Error() string {
	return "manifest changes the trusted preregistration " + string(value)
}
