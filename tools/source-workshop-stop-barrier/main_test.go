package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestCanStopOwnedExpectProcessGroup(t *testing.T) {
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
	time.Sleep(50 * time.Millisecond)
	if err := syscall.Kill(-pgid, syscall.SIGSTOP); err != nil {
		t.Fatalf("stop owned Expect process group %d: %v", pgid, err)
	}
}
