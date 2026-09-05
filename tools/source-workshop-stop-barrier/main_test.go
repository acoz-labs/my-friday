package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestCanStopOwnedExpectCandidate(t *testing.T) {
	if _, err := os.Stat("/usr/bin/expect"); err != nil {
		t.Skip("native Expect unavailable")
	}
	script := filepath.Join(t.TempDir(), "wait.exp")
	if err := os.WriteFile(script, []byte("#!/usr/bin/expect -f\nspawn -noecho /bin/sleep 30\nexpect eof\n"), 0600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("/usr/bin/expect", script)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	pgid := cmd.Process.Pid
	t.Cleanup(func() {
		_ = syscall.Kill(-pgid, syscall.SIGCONT)
		_ = syscall.Kill(-pgid, syscall.SIGKILL)
		_, _ = cmd.Process.Wait()
	})
	var candidatePID int
	for deadline := time.Now().Add(time.Second); time.Now().Before(deadline); {
		for pid, record := range processTable() {
			if record.ppid == pgid {
				candidatePID = pid
				break
			}
		}
		if candidatePID != 0 {
			break
		}
		time.Sleep(time.Millisecond)
	}
	if candidatePID == 0 || !ownedDescendant(pgid, candidatePID) || ownedDescendant(pgid, pgid) {
		t.Fatalf("candidate=%d ownership validation failed", candidatePID)
	}
	if err := syscall.Kill(candidatePID, syscall.SIGSTOP); err != nil {
		t.Fatal(err)
	}
	_ = syscall.Kill(candidatePID, syscall.SIGCONT)
}

func TestReadStoppedPIDRequiresOwnedNumericCandidate(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "candidate-stopped")
	if err := os.WriteFile(marker, []byte("not-a-pid\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := readStoppedPID(marker, os.Getpid()); err == nil {
		t.Fatal("accepted invalid marker")
	}
}
