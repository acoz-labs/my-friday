package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"testing"
	"time"
)

func TestRunnerTimesOutProcessGroup(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "--cwd", t.TempDir(), "--timeout", "100ms", "--", "/bin/sh", "-c", "sleep 30 & wait")
	started := time.Now()
	if err := cmd.Run(); err == nil {
		t.Fatal("timed out group succeeded")
	}
	if time.Since(started) > 10*time.Second {
		t.Fatal("timeout did not bound the command")
	}
}

func TestRunnerKillsEscapedSetsidDescendant(t *testing.T) {
	dir := t.TempDir()
	pidfile := filepath.Join(dir, "escaped.pid")
	cmd := exec.Command("go", "run", ".", "--cwd", dir, "--timeout", "2s", "--env", "ESCAPE_HELPER=parent", "--env", "PIDFILE="+pidfile, "--", os.Args[0], "-test.run=TestEscapedDescendantHelper")
	if err := cmd.Run(); err == nil {
		t.Fatal("runner accepted a command that created a setsid descendant")
	}
	body, err := os.ReadFile(pidfile)
	if err != nil {
		t.Fatalf("escaped descendant did not start: %v", err)
	}
	pid, _ := strconv.Atoi(string(body))
	defer syscall.Kill(pid, syscall.SIGKILL)
	if err = syscall.Kill(pid, 0); err == nil {
		t.Fatal("setsid descendant survived runner cleanup")
	}
}

func TestRunnerReapsChildrenOnSignals(t *testing.T) {
	for _, sig := range []syscall.Signal{syscall.SIGINT, syscall.SIGTERM} {
		for _, mode := range []string{"ordinary", "escaped"} {
			t.Run(fmt.Sprintf("%s/%s", sig, mode), func(t *testing.T) {
				t.Parallel()
				dir := t.TempDir()
				runner := filepath.Join(dir, "acceptance-runner")
				build := exec.Command("go", "build", "-o", runner, ".")
				if output, err := build.CombinedOutput(); err != nil {
					t.Fatalf("build runner: %v: %s", err, output)
				}
				rootPIDFile := filepath.Join(dir, "root.pid")
				escapedPIDFile := filepath.Join(dir, "escaped.pid")
				cmd := exec.Command(runner, "--cwd", dir, "--timeout", "30s", "--env", "SIGNAL_HELPER="+mode, "--env", "ROOT_PIDFILE="+rootPIDFile, "--env", "ESCAPED_PIDFILE="+escapedPIDFile, "--", os.Args[0], "-test.run=TestSignalDescendantHelper")
				if err := cmd.Start(); err != nil {
					t.Fatal(err)
				}
				rootPID := waitForPID(t, rootPIDFile)
				pids := []int{rootPID}
				if mode == "escaped" {
					pids = append(pids, waitForPID(t, escapedPIDFile))
				}
				if err := cmd.Process.Signal(sig); err != nil {
					t.Fatalf("signal runner: %v", err)
				}
				err := cmd.Wait()
				var exit *exec.ExitError
				if !errors.As(err, &exit) || exit.ExitCode() != 128+int(sig) {
					t.Fatalf("runner status = %v, want %d", err, 128+int(sig))
				}
				for _, pid := range pids {
					waitForGone(t, pid)
				}
			})
		}
	}
}

func TestRunnerPTYWrappedTimeoutReapsEscapedDescendant(t *testing.T) {
	dir := t.TempDir()
	runner := buildTestRunner(t, dir)
	rootPIDFile := filepath.Join(dir, "root.pid")
	escapedPIDFile := filepath.Join(dir, "escaped.pid")
	transcript := filepath.Join(dir, "timeout.private")
	cmd := exec.Command(runner, "--cwd", dir, "--timeout", "500ms",
		"--env", "SIGNAL_HELPER=escaped", "--env", "ROOT_PIDFILE="+rootPIDFile, "--env", "ESCAPED_PIDFILE="+escapedPIDFile,
		"--", "/usr/bin/expect", launcherCaptureDriver(t), transcript, os.Args[0], "-test.run=TestSignalDescendantHelper")
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	rootPID := waitForPID(t, rootPIDFile)
	escapedPID := waitForPID(t, escapedPIDFile)
	if err := cmd.Wait(); err == nil {
		t.Fatal("PTY-wrapped timeout succeeded")
	}
	waitForGone(t, rootPID)
	waitForGone(t, escapedPID)
	assertPrivateTranscript(t, transcript)
}

func TestRunnerPTYWrappedSignalsReapDescendants(t *testing.T) {
	for _, sig := range []syscall.Signal{syscall.SIGINT, syscall.SIGTERM} {
		t.Run(sig.String(), func(t *testing.T) {
			dir := t.TempDir()
			runner := buildTestRunner(t, dir)
			rootPIDFile := filepath.Join(dir, "root.pid")
			escapedPIDFile := filepath.Join(dir, "escaped.pid")
			transcript := filepath.Join(dir, "signal.private")
			cmd := exec.Command(runner, "--cwd", dir, "--timeout", "30s",
				"--env", "SIGNAL_HELPER=escaped", "--env", "ROOT_PIDFILE="+rootPIDFile, "--env", "ESCAPED_PIDFILE="+escapedPIDFile,
				"--", "/usr/bin/expect", launcherCaptureDriver(t), transcript, os.Args[0], "-test.run=TestSignalDescendantHelper")
			if err := cmd.Start(); err != nil {
				t.Fatal(err)
			}
			rootPID := waitForPID(t, rootPIDFile)
			escapedPID := waitForPID(t, escapedPIDFile)
			if err := cmd.Process.Signal(sig); err != nil {
				t.Fatal(err)
			}
			err := cmd.Wait()
			var exit *exec.ExitError
			if !errors.As(err, &exit) || exit.ExitCode() != 128+int(sig) {
				t.Fatalf("runner status = %v, want %d", err, 128+int(sig))
			}
			waitForGone(t, rootPID)
			waitForGone(t, escapedPID)
			assertPrivateTranscript(t, transcript)
		})
	}
}

func buildTestRunner(t *testing.T, dir string) string {
	t.Helper()
	runner := filepath.Join(dir, "acceptance-runner")
	build := exec.Command("go", "build", "-o", runner, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build runner: %v: %s", err, output)
	}
	return runner
}

func launcherCaptureDriver(t *testing.T) string {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Clean(filepath.Join(cwd, "..", "..", "config", "acceptance", "launcher-capture.exp"))
	if _, err = os.Stat(path); err != nil {
		t.Fatal(err)
	}
	return path
}

func assertPrivateTranscript(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("transcript mode = %o", info.Mode().Perm())
	}
}

func TestSignalDescendantHelper(t *testing.T) {
	mode := os.Getenv("SIGNAL_HELPER")
	if mode == "" {
		return
	}
	if err := os.WriteFile(os.Getenv("ROOT_PIDFILE"), []byte(strconv.Itoa(os.Getpid())), 0o600); err != nil {
		os.Exit(3)
	}
	if mode == "escaped" {
		cmd := exec.Command(os.Args[0], "-test.run=TestSignalEscapedHelper")
		cmd.Env = []string{"SIGNAL_ESCAPED_HELPER=1", "ESCAPED_PIDFILE=" + os.Getenv("ESCAPED_PIDFILE"), "MY_FRIDAY_ACCEPTANCE_PROCESS_TOKEN=" + os.Getenv("MY_FRIDAY_ACCEPTANCE_PROCESS_TOKEN")}
		cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
		if err := cmd.Start(); err != nil {
			os.Exit(4)
		}
	}
	time.Sleep(30 * time.Second)
}

func TestSignalEscapedHelper(t *testing.T) {
	if os.Getenv("SIGNAL_ESCAPED_HELPER") == "" {
		return
	}
	if err := os.WriteFile(os.Getenv("ESCAPED_PIDFILE"), []byte(strconv.Itoa(os.Getpid())), 0o600); err != nil {
		os.Exit(3)
	}
	time.Sleep(30 * time.Second)
}

func waitForPID(t *testing.T, path string) int {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		body, err := os.ReadFile(path)
		if err == nil {
			pid, parseErr := strconv.Atoi(string(body))
			if parseErr == nil {
				return pid
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("PID file %s was not created", path)
	return 0
}

func waitForGone(t *testing.T, pid int) {
	t.Helper()
	defer syscall.Kill(pid, syscall.SIGKILL)
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if syscall.Kill(pid, 0) != nil {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("process %d survived runner exit", pid)
}

func TestEscapedDescendantHelper(t *testing.T) {
	mode := os.Getenv("ESCAPE_HELPER")
	if mode == "" {
		return
	}
	if mode == "parent" {
		cmd := exec.Command(os.Args[0], "-test.run=TestEscapedDescendantHelper")
		cmd.Env = []string{"ESCAPE_HELPER=child", "PIDFILE=" + os.Getenv("PIDFILE"), "MY_FRIDAY_ACCEPTANCE_PROCESS_TOKEN=" + os.Getenv("MY_FRIDAY_ACCEPTANCE_PROCESS_TOKEN")}
		if err := cmd.Start(); err != nil {
			os.Exit(2)
		}
		return
	}
	if mode == "child" {
		cmd := exec.Command(os.Args[0], "-test.run=TestEscapedDescendantHelper")
		cmd.Env = []string{"ESCAPE_HELPER=grandchild", "PIDFILE=" + os.Getenv("PIDFILE"), "MY_FRIDAY_ACCEPTANCE_PROCESS_TOKEN=" + os.Getenv("MY_FRIDAY_ACCEPTANCE_PROCESS_TOKEN")}
		cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
		if err := cmd.Start(); err != nil {
			os.Exit(2)
		}
		return
	}
	if err := os.WriteFile(os.Getenv("PIDFILE"), []byte(strconv.Itoa(os.Getpid())), 0o600); err != nil {
		os.Exit(3)
	}
	time.Sleep(30 * time.Second)
}
