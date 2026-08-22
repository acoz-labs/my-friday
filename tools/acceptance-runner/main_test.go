package main

import (
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
