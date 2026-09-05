package main

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type HarnessName string

const (
	HarnessCodex  HarnessName = "codex"
	HarnessClaude HarnessName = "claude"
)

type DriverControl struct {
	OSFixtureOnlyReadBoundary bool     `json:"os_fixture_only_read_boundary"`
	ModelToolNetworkDenied    bool     `json:"model_tool_network_denied"`
	NativeSkillVisibility     bool     `json:"native_skill_visibility"`
	NativeBodyReadConstrained bool     `json:"native_body_read_constrained"`
	NativeWorkerEvents        bool     `json:"native_worker_events"`
	WorkerPrelaunchLimit      bool     `json:"worker_prelaunch_limit"`
	BuiltinPredispatchLimit   bool     `json:"builtin_predispatch_limit"`
	BrokerPredispatchLimit    bool     `json:"broker_predispatch_limit"`
	OutputLimitEnforced       bool     `json:"output_limit_enforced"`
	Evidence                  []string `json:"evidence"`
}

type DriverProbe struct {
	Version        int           `json:"version"`
	Harness        HarnessName   `json:"harness"`
	Executable     string        `json:"executable"`
	CLIRevision    string        `json:"cli_revision"`
	State          string        `json:"state"`
	Unavailable    []string      `json:"unavailable_reasons"`
	Controls       DriverControl `json:"controls"`
	ProbedCommands []string      `json:"probed_commands"`
}

func ProbeHarness(ctx context.Context, harness HarnessName, path string) (DriverProbe, error) {
	if harness != HarnessCodex && harness != HarnessClaude {
		return DriverProbe{}, fmt.Errorf("unsupported harness %q", harness)
	}
	if path == "" || filepath.Base(path) != string(harness) {
		return DriverProbe{}, errors.New("probe requires the exact named harness executable")
	}
	version, err := credentialFreeProbe(ctx, path, "--version")
	if err != nil {
		return DriverProbe{}, err
	}
	helpArgs := []string{"--help"}
	if harness == HarnessCodex {
		helpArgs = []string{"exec", "--help"}
	}
	help, err := credentialFreeProbe(ctx, path, helpArgs...)
	if err != nil {
		return DriverProbe{}, err
	}
	controls := DriverControl{
		Evidence: []string{
			"version and help text inspected without inference or authentication commands",
			"no broker or model process was activated because the native process boundary is unsupported",
		},
	}
	if harness == HarnessCodex && strings.Contains(help, "--sandbox") {
		controls.Evidence = append(controls.Evidence, "CLI advertises read-only/workspace-write sandbox modes; workspace-write is not restricted-read proof")
	}
	if harness == HarnessClaude && strings.Contains(help, "--allowedTools") {
		controls.Evidence = append(controls.Evidence, "CLI advertises tool allowlists; flags are not an OS-backed read/network denial canary")
	}
	reasons := []string{
		"no OS-enforced fixture-only read boundary was demonstrated without exposing host authentication",
		"native skill body reads were not proven constrained to the staged allowlist",
		"native worker identity/inheritance and pre-launch depth/count enforcement were not demonstrated",
		"enabled built-in tools lack an equivalent pre-dispatch eight-call rejection hook",
	}
	return DriverProbe{
		Version: SchemaVersion, Harness: harness, Executable: filepath.Base(path), CLIRevision: firstLine(version),
		State: "unavailable", Unavailable: reasons, Controls: controls,
		ProbedCommands: []string{string(harness) + " --version", string(harness) + " " + strings.Join(helpArgs, " ")},
	}, nil
}

func (probe DriverProbe) LiveEligible(mode string) bool {
	if probe.State != "supported" || !contains(RoutingModes, mode) {
		return false
	}
	controls := probe.Controls
	common := controls.OSFixtureOnlyReadBoundary && controls.ModelToolNetworkDenied && controls.BuiltinPredispatchLimit
	if !common {
		return false
	}
	switch mode {
	case "native-catalogue":
		return controls.NativeSkillVisibility && controls.NativeBodyReadConstrained && controls.NativeWorkerEvents && controls.WorkerPrelaunchLimit
	case "lookup-direct":
		return controls.BrokerPredispatchLimit
	case "lookup-worker":
		return controls.BrokerPredispatchLimit && controls.NativeWorkerEvents && controls.WorkerPrelaunchLimit
	default:
		return false
	}
}

func credentialFreeProbe(ctx context.Context, path string, args ...string) (string, error) {
	probeCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	command := exec.CommandContext(probeCtx, path, args...)
	command.Env = []string{"PATH=/usr/bin:/bin:/usr/sbin:/sbin"}
	body, err := command.CombinedOutput()
	if probeCtx.Err() != nil {
		return "", fmt.Errorf("probe %s timed out", filepath.Base(path))
	}
	if err != nil {
		return "", fmt.Errorf("probe %s: %w", filepath.Base(path), err)
	}
	return string(body), nil
}

func firstLine(value string) string {
	line, _, _ := strings.Cut(strings.TrimSpace(value), "\n")
	return line
}
