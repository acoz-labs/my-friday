package main

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
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
	var tokenBytes [16]byte
	if _, err := rand.Read(tokenBytes[:]); err != nil {
		fatal(err)
	}
	processToken := hex.EncodeToString(tokenBytes[:])
	cmd.Env = append(env, "MY_FRIDAY_ACCEPTANCE_PROCESS_TOKEN="+processToken)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(signals)
	baseline := processTable()
	if err := cmd.Start(); err != nil {
		fatal(err)
	}
	pgid := cmd.Process.Pid
	tracker := newProcessTracker(pgid, processToken, baseline, args[0])
	tracker.start()
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	var err error
	select {
	case err = <-done:
	case received := <-signals:
		result := tracker.stopAndKill()
		<-done
		if !result.quiescent {
			fatal(errors.New("command processes remained after signal cleanup"))
		}
		signal.Stop(signals)
		os.Exit(128 + int(received.(syscall.Signal)))
	case <-time.After(*timeout):
		tracker.stopAndKill()
		<-done
		fatal(errors.New("command timed out and its process group was killed"))
	}
	result := tracker.stopAndKill()
	if result.descendants || !result.quiescent {
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
	pid, ppid, pgid          int
	identity, state, command string
}
type processTracker struct {
	root       int
	token      string
	executable string
	baseline   map[int]string
	mu         sync.Mutex
	seen       map[int]string
	stop       chan struct{}
	done       chan struct{}
}

func newProcessTracker(root int, token string, baseline map[int]processRecord, executable string) *processTracker {
	identities := map[int]string{}
	for pid, record := range baseline {
		identities[pid] = record.identity
	}
	return &processTracker{root: root, token: token, executable: filepath.Base(executable), baseline: identities, seen: map[int]string{}, stop: make(chan struct{}), done: make(chan struct{})}
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
	for pid, record := range records {
		fields := strings.Fields(record.command)
		sameExecutable := len(fields) > 0 && filepath.Base(fields[0]) == t.executable
		_, existedBefore := t.baseline[pid]
		if strings.Contains(record.command, "MY_FRIDAY_ACCEPTANCE_PROCESS_TOKEN="+t.token) || (!existedBefore && sameExecutable) {
			t.seen[pid] = record.identity
		}
	}
}

type cleanupResult struct {
	descendants bool
	quiescent   bool
}

func (t *processTracker) stopAndKill() cleanupResult {
	_ = syscall.Kill(-t.root, syscall.SIGSTOP)
	for i := 0; i < 5; i++ {
		t.capture()
		t.mu.Lock()
		for pid, identity := range t.seen {
			if pid != t.root && currentProcessIdentity(pid) == identity {
				_ = syscall.Kill(pid, syscall.SIGSTOP)
			}
		}
		t.mu.Unlock()
		time.Sleep(2 * time.Millisecond)
	}
	select {
	case <-t.stop:
	default:
		close(t.stop)
	}
	<-t.done
	t.capture()
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
			return cleanupResult{descendants: surviving, quiescent: true}
		}
		time.Sleep(10 * time.Millisecond)
	}
	return cleanupResult{descendants: surviving, quiescent: false}
}

func processTable() map[int]processRecord {
	out, err := exec.Command("/bin/ps", "-axeww", "-o", "pid=,ppid=,pgid=,stat=,lstart=,command=").Output()
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
		result[pid] = processRecord{pid: pid, ppid: ppid, pgid: pgid, state: fields[3], identity: fields[0] + " " + strings.Join(fields[4:9], " "), command: strings.Join(fields[9:], " ")}
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
