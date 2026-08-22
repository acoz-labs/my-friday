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
	"sync"
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
	tracker := newProcessTracker(pgid)
	tracker.start()
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	var err error
	select {
	case err = <-done:
	case <-time.After(*timeout):
		tracker.stopAndKill()
		<-done
		fatal(errors.New("command timed out and its process group was killed"))
	}
	escaped := tracker.stopAndKill()
	if escaped || groupExists(pgid) {
		fatal(errors.New("command created descendants or exited with surviving process-group members"))
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

type processRecord struct {
	pid, ppid, pgid int
	identity, state string
}
type processTracker struct {
	root int
	mu   sync.Mutex
	seen map[int]string
	stop chan struct{}
	done chan struct{}
}

func newProcessTracker(root int) *processTracker {
	return &processTracker{root: root, seen: map[int]string{}, stop: make(chan struct{}), done: make(chan struct{})}
}
func (t *processTracker) start() {
	t.capture()
	go func() {
		defer close(t.done)
		ticker := time.NewTicker(time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				t.capture()
			case <-t.stop:
				return
			}
		}
	}()
}
func (t *processTracker) capture() {
	records := processTable()
	t.mu.Lock()
	defer t.mu.Unlock()
	if root, ok := records[t.root]; ok {
		t.seen[t.root] = root.identity
	}
	changed := true
	for changed {
		changed = false
		for pid, record := range records {
			if pid == t.root {
				continue
			}
			if _, parentKnown := t.seen[record.ppid]; parentKnown {
				if _, known := t.seen[pid]; !known {
					t.seen[pid] = record.identity
					changed = true
				}
			}
		}
	}
}
func (t *processTracker) stopAndKill() bool {
	select {
	case <-t.stop:
	default:
		close(t.stop)
	}
	<-t.done
	for i := 0; i < 5; i++ {
		t.capture()
		time.Sleep(2 * time.Millisecond)
	}
	t.mu.Lock()
	seen := make(map[int]string, len(t.seen))
	for pid, identity := range t.seen {
		seen[pid] = identity
	}
	t.mu.Unlock()
	surviving := false
	for pid, identity := range seen {
		if pid != t.root && currentProcessIdentity(pid) == identity {
			surviving = true
		}
	}
	_ = syscall.Kill(-t.root, syscall.SIGKILL)
	for pid, identity := range seen {
		if pid != t.root && currentProcessIdentity(pid) == identity {
			_ = syscall.Kill(pid, syscall.SIGKILL)
		}
	}
	for retry := 0; retry < 50; retry++ {
		alive := false
		for pid, identity := range seen {
			if currentProcessIdentity(pid) == identity {
				alive = true
				break
			}
		}
		if !alive && !groupExists(t.root) {
			return surviving
		}
		time.Sleep(10 * time.Millisecond)
	}
	return true
}

func processTable() map[int]processRecord {
	out, err := exec.Command("/bin/ps", "-ax", "-o", "pid=,ppid=,pgid=,stat=,lstart=").Output()
	if err != nil {
		return map[int]processRecord{}
	}
	result := map[int]processRecord{}
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 9 {
			continue
		}
		pid, e1 := strconv.Atoi(fields[0])
		ppid, e2 := strconv.Atoi(fields[1])
		pgid, e3 := strconv.Atoi(fields[2])
		if e1 != nil || e2 != nil || e3 != nil {
			continue
		}
		result[pid] = processRecord{pid: pid, ppid: ppid, pgid: pgid, state: fields[3], identity: fields[0] + " " + strings.Join(fields[4:9], " ")}
	}
	return result
}
func currentProcessIdentity(pid int) string {
	if record, ok := processTable()[pid]; ok && !strings.Contains(record.state, "Z") {
		return record.identity
	}
	return ""
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
