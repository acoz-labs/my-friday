package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type values []string

func (v *values) String() string     { return strings.Join(*v, ",") }
func (v *values) Set(s string) error { *v = append(*v, s); return nil }

func main() {
	var env values
	cwd := flag.String("cwd", "", "working directory")
	timeout := flag.Duration("timeout", 30*time.Second, "hard timeout")
	flag.Var(&env, "env", "exact KEY=VALUE environment entry")
	flag.Parse()
	args := flag.Args()
	if *cwd == "" || len(args) == 0 || *timeout <= 0 {
		fmt.Fprintln(os.Stderr, "usage: acceptance-runner --cwd DIR --timeout 30s --env K=V -- COMMAND [ARG...]")
		os.Exit(64)
	}
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = *cwd
	cmd.Env = env
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		fatal(err)
	}
	pgid := cmd.Process.Pid
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	var err error
	select {
	case err = <-done:
	case <-time.After(*timeout):
		_ = syscall.Kill(-pgid, syscall.SIGKILL)
		<-done
		fatal(errors.New("command timed out and its process group was killed"))
	}
	if groupExists(pgid) {
		_ = syscall.Kill(-pgid, syscall.SIGKILL)
		fatal(errors.New("command exited with surviving process-group members"))
	}
	if err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			if status, ok := exit.Sys().(syscall.WaitStatus); ok {
				os.Exit(status.ExitStatus())
			}
		}
		fatal(err)
	}
}

func groupExists(pgid int) bool {
	out, err := exec.Command("/bin/ps", "-ax", "-o", "pgid=").Output()
	if err != nil {
		return true
	}
	for _, field := range strings.Fields(string(out)) {
		group, err := strconv.Atoi(field)
		if err == nil && group == pgid {
			return true
		}
	}
	return false
}

func fatal(err error) {
	_, _ = io.WriteString(os.Stderr, err.Error()+"\n")
	os.Exit(1)
}
