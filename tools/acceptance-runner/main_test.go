package main

import (
	"os/exec"
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
