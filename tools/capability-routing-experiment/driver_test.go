package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"testing"
	"time"
)

func TestUnavailableProbeNeverBecomesLiveEligible(t *testing.T) {
	probe := DriverProbe{State: "unavailable", Controls: DriverControl{OSFixtureOnlyReadBoundary: true, ModelToolNetworkDenied: true, NativeSkillVisibility: true, NativeBodyReadConstrained: true, NativeWorkerEvents: true, WorkerPrelaunchLimit: true, BuiltinPredispatchLimit: true, BrokerPredispatchLimit: true}}
	for _, mode := range RoutingModes {
		if probe.LiveEligible(mode) {
			t.Fatalf("unavailable probe eligible for %s", mode)
		}
	}
}

func TestRunAndScoreAllUnavailableEvidenceIsInconclusive(t *testing.T) {
	bin := t.TempDir()
	writeHarness := func(name, version string) {
		t.Helper()
		body := "#!/bin/sh\nif [ \"$1\" = \"--version\" ]; then echo '" + version + "'; else echo 'credential-free help'; fi\n"
		if err := os.WriteFile(filepath.Join(bin, name), []byte(body), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	writeHarness("codex", "codex-cli 0.153.4")
	writeHarness("claude", "2.1.193 (Claude Code)")
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	runRoot := t.TempDir()
	if err := runCommand([]string{"--live", "--data", "testdata", "--run-root", runRoot}); err != nil {
		t.Fatal(err)
	}
	if err := scoreCommand([]string{"--data", "testdata", "--run-root", runRoot}); err != nil {
		t.Fatal(err)
	}
	var comparison Comparison
	if err := ReadJSON(filepath.Join(runRoot, "scores.json"), &comparison); err != nil {
		t.Fatal(err)
	}
	if comparison.Recommendation != "inconclusive" {
		t.Fatalf("recommendation=%s", comparison.Recommendation)
	}
}

func TestCredentialFreeProbeSupervisesEscapedDescendants(t *testing.T) {
	if os.Getenv("GO_WANT_PROBE_HELPER") != "" {
		if os.Getenv("GO_WANT_PROBE_HELPER") == "child" {
			time.Sleep(30 * time.Second)
			return
		}
		child := exec.Command(os.Args[0], "-test.run=TestCredentialFreeProbeSupervisesEscapedDescendants")
		child.Env = []string{"GO_WANT_PROBE_HELPER=child", "PROBE_PIDFILE=" + os.Getenv("PROBE_PIDFILE"), "PATH=/usr/bin:/bin"}
		child.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
		if err := child.Start(); err != nil {
			os.Exit(2)
		}
		if err := os.WriteFile(os.Getenv("PROBE_PIDFILE"), []byte(strconv.Itoa(child.Process.Pid)), 0o600); err != nil {
			os.Exit(3)
		}
		time.Sleep(30 * time.Second)
		return
	}
	pidFile := filepath.Join(t.TempDir(), "probe.pid")
	t.Setenv("GO_WANT_PROBE_HELPER", "parent")
	t.Setenv("PROBE_PIDFILE", pidFile)
	oldTimeout := credentialFreeProbeTimeout
	oldEnv := credentialFreeProbeExtraEnv
	credentialFreeProbeTimeout = 100 * time.Millisecond
	credentialFreeProbeExtraEnv = []string{"GO_WANT_PROBE_HELPER=parent", "PROBE_PIDFILE=" + pidFile}
	t.Cleanup(func() { credentialFreeProbeTimeout = oldTimeout; credentialFreeProbeExtraEnv = oldEnv })
	_, err := credentialFreeProbe(context.Background(), os.Args[0], "-test.run=TestCredentialFreeProbeSupervisesEscapedDescendants")
	if err == nil {
		t.Fatal("timed out probe succeeded")
	}
	body, readErr := os.ReadFile(pidFile)
	if readErr != nil {
		t.Fatal(readErr)
	}
	pid, parseErr := strconv.Atoi(string(body))
	if parseErr != nil {
		t.Fatal(parseErr)
	}
	defer syscall.Kill(pid, syscall.SIGKILL)
	if signalErr := syscall.Kill(pid, 0); signalErr == nil || !errors.Is(signalErr, os.ErrProcessDone) && signalErr != syscall.ESRCH {
		t.Fatalf("probe descendant %d survived: %v", pid, signalErr)
	}
}

func TestModeSpecificDriverControls(t *testing.T) {
	probe := DriverProbe{State: "supported", Controls: DriverControl{OSFixtureOnlyReadBoundary: true, ModelToolNetworkDenied: true, BuiltinPredispatchLimit: true, BrokerPredispatchLimit: true}}
	if !probe.LiveEligible("lookup-direct") {
		t.Fatal("broker-only direct mode should be eligible")
	}
	if probe.LiveEligible("lookup-worker") || probe.LiveEligible("native-catalogue") {
		t.Fatal("native modes eligible without native controls")
	}
}
