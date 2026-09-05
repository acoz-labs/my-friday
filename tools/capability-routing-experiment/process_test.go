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

func TestSuperviseReapsEscapedReparentedDescendant(t *testing.T) {
	pidFile := filepath.Join(t.TempDir(), "escaped.pid")
	result, err := Supervise(context.Background(), Command{Path: os.Args[0], Args: []string{"-test.run=TestEscapedExperimentHelper"}, Env: []string{"GO_WANT_EXPERIMENT_HELPER=parent", "PIDFILE=" + pidFile, "PATH=/usr/bin:/bin"}, WorkDir: t.TempDir(), Timeout: 150 * time.Millisecond})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("result=%#v err=%v", result, err)
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
	if signalErr := syscall.Kill(pid, 0); signalErr == nil {
		t.Fatalf("escaped descendant %d survived", pid)
	}
}

func TestEscapedExperimentHelper(t *testing.T) {
	mode := os.Getenv("GO_WANT_EXPERIMENT_HELPER")
	if mode == "" {
		return
	}
	if mode == "parent" {
		command := exec.Command(os.Args[0], "-test.run=TestEscapedExperimentHelper")
		command.Env = append(os.Environ(), "GO_WANT_EXPERIMENT_HELPER=child")
		command.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
		if err := command.Start(); err != nil {
			os.Exit(2)
		}
		time.Sleep(30 * time.Second)
	}
	if mode == "child" {
		if err := os.WriteFile(os.Getenv("PIDFILE"), []byte(strconv.Itoa(os.Getpid())), 0o600); err != nil {
			os.Exit(3)
		}
		time.Sleep(30 * time.Second)
	}
}

func TestSuperviseTimesOutAndReapsOwnedProcessGroup(t *testing.T) {
	result, err := Supervise(context.Background(), Command{Path: "/bin/sh", Args: []string{"-c", "sleep 30 & wait"}, Env: []string{"PATH=/usr/bin:/bin"}, WorkDir: t.TempDir(), Timeout: 100 * time.Millisecond})
	if !errors.Is(err, context.DeadlineExceeded) || !result.TimedOut {
		t.Fatalf("result=%#v err=%v", result, err)
	}
	if len(result.Survivors) != 0 {
		t.Fatalf("survivors = %v", result.Survivors)
	}
}

func TestSuperviseDoesNotKillUnrelatedProcess(t *testing.T) {
	unrelated, err := os.StartProcess("/bin/sleep", []string{"sleep", "30"}, &os.ProcAttr{Files: []*os.File{os.Stdin, os.Stdout, os.Stderr}, Sys: &syscall.SysProcAttr{Setsid: true}})
	if err != nil {
		t.Fatal(err)
	}
	defer unrelated.Kill()
	_, _ = Supervise(context.Background(), Command{Path: "/bin/sh", Args: []string{"-c", "sleep 30"}, Env: []string{"PATH=/usr/bin:/bin"}, WorkDir: t.TempDir(), Timeout: 50 * time.Millisecond})
	if err = unrelated.Signal(syscall.Signal(0)); err != nil {
		t.Fatalf("unrelated process was killed: %v", err)
	}
}

func TestSuperviseCancellationReapsOwnedProcess(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() { time.Sleep(40 * time.Millisecond); cancel() }()
	result, err := Supervise(ctx, Command{Path: "/bin/sleep", Args: []string{"30"}, Env: []string{"PATH=/usr/bin:/bin"}, WorkDir: t.TempDir(), Timeout: time.Minute})
	if !errors.Is(err, context.Canceled) || !result.Cancelled || len(result.Survivors) != 0 {
		t.Fatalf("result=%#v err=%v", result, err)
	}
}

func TestRunLockAndImmutableEvidenceRefuseReuse(t *testing.T) {
	root := t.TempDir()
	lock, err := AcquireRunLock(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = AcquireRunLock(root); err == nil {
		t.Fatal("concurrent lock accepted")
	}
	if err = lock.Release(); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "attempts", "one.json")
	if err = CreateImmutableJSON(path, map[string]int{"version": 1}); err != nil {
		t.Fatal(err)
	}
	if err = CreateImmutableJSON(path, map[string]int{"version": 2}); err == nil {
		t.Fatal("immutable evidence overwritten")
	}
}

func TestExactImmutableEvidenceResumeIsIdempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "attempts.json")
	value := map[string]any{"version": float64(1), "state": "unavailable"}
	if err := WriteOrVerifyJSON(path, value); err != nil {
		t.Fatal(err)
	}
	if err := WriteOrVerifyJSON(path, value); err != nil {
		t.Fatalf("exact resume failed: %v", err)
	}
	if err := WriteOrVerifyJSON(path, map[string]any{"version": float64(1), "state": "complete"}); err == nil {
		t.Fatal("changed resume overwrote evidence")
	}
}

func TestRunLockRefusesCleanupAfterPathReplacement(t *testing.T) {
	root := t.TempDir()
	lock, err := AcquireRunLock(root)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Rename(filepath.Join(root, ".run.lock"), filepath.Join(root, "original.lock")); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(root, ".run.lock"), []byte("replacement"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err = lock.Release(); err == nil {
		t.Fatal("replaced lock path was removed")
	}
}

func TestProcessIdentitySurvivesReparentAndStateChange(t *testing.T) {
	start := "Fri Sep 5 12:00:00 2026"
	if !identityMatches(map[int]processIdentity{42: {ppid: 1, state: "S", start: start}}, 42, start) {
		t.Fatal("stable start identity did not survive reparent/state change")
	}
	if identityMatches(map[int]processIdentity{42: {ppid: 99, state: "Z", start: start}}, 42, start) {
		t.Fatal("zombie treated as live owned process")
	}
}

func TestProcessTableFailureLeavesOwnershipUnknown(t *testing.T) {
	tracker := newOwnedTracker(999999, "token")
	tracker.seen[999998] = "Fri Sep 5 12:00:00 2026"
	tracker.err = errors.New("process table unavailable")
	if survivors, err := tracker.kill(); err == nil || len(survivors) == 0 {
		t.Fatalf("survivors=%v err=%v", survivors, err)
	}
}
