package main

import "reflect"

// TrustedSourceCommit is the reviewed runner/corpus freeze. It is compiled into
// the validator so a manifest cannot redefine its own identity.
const TrustedSourceCommit = "c8954e9b8a9f620c7349f4d63a236a9c0ac64c77"

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
