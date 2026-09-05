package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

type Command struct {
	Path    string
	Args    []string
	Env     []string
	WorkDir string
	Timeout time.Duration
}

type ProcessResult struct {
	ExitCode       int    `json:"exit_code"`
	TimedOut       bool   `json:"timed_out"`
	Cancelled      bool   `json:"cancelled"`
	HadDescendants bool   `json:"had_descendants"`
	Survivors      []int  `json:"survivors"`
	WallMillis     int64  `json:"wall_millis"`
	Stdout         string `json:"stdout"`
	Stderr         string `json:"stderr"`
}

func Supervise(ctx context.Context, spec Command) (ProcessResult, error) {
	if spec.Path == "" || spec.WorkDir == "" || spec.Timeout <= 0 {
		return ProcessResult{}, errors.New("path, work directory, and positive timeout are required")
	}
	var random [16]byte
	if _, err := rand.Read(random[:]); err != nil {
		return ProcessResult{}, err
	}
	token := hex.EncodeToString(random[:])
	commandCtx, cancel := context.WithTimeout(ctx, spec.Timeout)
	defer cancel()
	command := exec.Command(spec.Path, spec.Args...)
	command.Dir = spec.WorkDir
	command.Env = append(append([]string{}, spec.Env...), "MY_FRIDAY_EXPERIMENT_PROCESS_TOKEN="+token)
	var stdout, stderr strings.Builder
	command.Stdout, command.Stderr = &stdout, &stderr
	command.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	started := time.Now()
	if err := command.Start(); err != nil {
		return ProcessResult{}, err
	}
	tracker := newOwnedTracker(command.Process.Pid, token)
	tracker.start()
	done := make(chan error, 1)
	go func() { done <- command.Wait() }()
	var waitErr error
	result := ProcessResult{ExitCode: 0}
	select {
	case waitErr = <-done:
	case <-commandCtx.Done():
		result.TimedOut = errors.Is(commandCtx.Err(), context.DeadlineExceeded)
		result.Cancelled = !result.TimedOut
		_, _ = tracker.kill()
		waitErr = <-done
	}
	tracker.stopTracking()
	result.HadDescendants = tracker.hadDescendants()
	var cleanupErr error
	result.Survivors, cleanupErr = tracker.kill()
	result.WallMillis = time.Since(started).Milliseconds()
	result.Stdout, result.Stderr = stdout.String(), stderr.String()
	if waitErr != nil {
		var exit *exec.ExitError
		if errors.As(waitErr, &exit) {
			result.ExitCode = exit.ExitCode()
		} else {
			result.ExitCode = -1
		}
	}
	if result.TimedOut || result.Cancelled {
		if cleanupErr != nil {
			return result, cleanupErr
		}
		return result, commandCtx.Err()
	}
	if cleanupErr != nil {
		return result, cleanupErr
	}
	if len(result.Survivors) > 0 {
		return result, errors.New("owned descendants survived cleanup")
	}
	return result, waitErr
}

type processIdentity struct {
	ppid    int
	state   string
	start   string
	command string
}
type ownedTracker struct {
	root int
	mu   sync.Mutex
	seen map[int]string
	stop chan struct{}
	done chan struct{}
	err  error
}

func newOwnedTracker(root int, _ string) *ownedTracker {
	return &ownedTracker{root: root, seen: map[int]string{}, stop: make(chan struct{}), done: make(chan struct{})}
}
func (tracker *ownedTracker) start() {
	tracker.capture()
	go func() {
		defer close(tracker.done)
		ticker := time.NewTicker(2 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				tracker.capture()
			case <-tracker.stop:
				return
			}
		}
	}()
}
func (tracker *ownedTracker) capture() {
	table, err := experimentProcessTable()
	tracker.mu.Lock()
	defer tracker.mu.Unlock()
	if err != nil {
		tracker.err = err
		return
	}
	if root, ok := table[tracker.root]; ok {
		tracker.seen[tracker.root] = root.start
	}
	changed := true
	for changed {
		changed = false
		for pid, process := range table {
			if _, ok := tracker.seen[process.ppid]; ok {
				if _, known := tracker.seen[pid]; !known {
					tracker.seen[pid] = process.start
					changed = true
				}
			}
		}
	}
}
func (tracker *ownedTracker) stopTracking() {
	select {
	case <-tracker.stop:
	default:
		close(tracker.stop)
	}
	<-tracker.done
	tracker.capture()
}
func (tracker *ownedTracker) hadDescendants() bool {
	tracker.mu.Lock()
	defer tracker.mu.Unlock()
	return len(tracker.seen) > 1
}
func (tracker *ownedTracker) kill() ([]int, error) {
	tracker.capture()
	tracker.mu.Lock()
	seen := map[int]string{}
	for pid, identity := range tracker.seen {
		seen[pid] = identity
	}
	trackingErr := tracker.err
	tracker.mu.Unlock()
	_ = syscall.Kill(-tracker.root, syscall.SIGKILL)
	if trackingErr != nil {
		return sortedPIDs(seen), fmt.Errorf("process ownership is unknown: %w", trackingErr)
	}
	table, err := experimentProcessTable()
	if err != nil {
		return sortedPIDs(seen), fmt.Errorf("process ownership is unknown: %w", err)
	}
	for pid, identity := range seen {
		if pid != tracker.root && identityMatches(table, pid, identity) {
			_ = syscall.Kill(pid, syscall.SIGKILL)
		}
	}
	for retry := 0; retry < 100; retry++ {
		table, err = experimentProcessTable()
		if err != nil {
			return sortedPIDs(seen), fmt.Errorf("process ownership is unknown: %w", err)
		}
		var alive []int
		for pid, identity := range seen {
			if identityMatches(table, pid, identity) {
				alive = append(alive, pid)
			}
		}
		if len(alive) == 0 {
			return nil, nil
		}
		time.Sleep(5 * time.Millisecond)
	}
	table, err = experimentProcessTable()
	if err != nil {
		return sortedPIDs(seen), fmt.Errorf("process ownership is unknown: %w", err)
	}
	var survivors []int
	for pid, identity := range seen {
		if identityMatches(table, pid, identity) {
			survivors = append(survivors, pid)
		}
	}
	return survivors, nil
}

func experimentProcessTable() (map[int]processIdentity, error) {
	body, err := exec.Command("/bin/ps", "-axo", "pid=,ppid=,stat=,lstart=,command=").Output()
	if err != nil {
		return nil, err
	}
	result := map[int]processIdentity{}
	for _, line := range strings.Split(string(body), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 8 {
			continue
		}
		pid, e1 := strconv.Atoi(fields[0])
		ppid, e2 := strconv.Atoi(fields[1])
		if e1 != nil || e2 != nil {
			continue
		}
		result[pid] = processIdentity{ppid: ppid, state: fields[2], start: strings.Join(fields[3:8], " "), command: strings.Join(fields[8:], " ")}
	}
	return result, nil
}
func identityMatches(table map[int]processIdentity, pid int, start string) bool {
	process, ok := table[pid]
	return ok && !strings.Contains(process.state, "Z") && process.start == start
}
func sortedPIDs(seen map[int]string) []int {
	result := make([]int, 0, len(seen))
	for pid := range seen {
		result = append(result, pid)
	}
	sort.Ints(result)
	return result
}

type RunLock struct {
	path string
	file *os.File
	info os.FileInfo
}

func AcquireRunLock(root string) (*RunLock, error) {
	if err := os.MkdirAll(root, 0o700); err != nil {
		return nil, err
	}
	path := filepath.Join(root, ".run.lock")
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return nil, fmt.Errorf("run root is already in use: %w", err)
	}
	if _, err = fmt.Fprintf(file, "%d\n", os.Getpid()); err != nil {
		file.Close()
		return nil, err
	}
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}
	return &RunLock{path: path, file: file, info: info}, nil
}
func (lock *RunLock) Release() error {
	if lock == nil || lock.file == nil {
		return nil
	}
	pathInfo, err := os.Lstat(lock.path)
	if err != nil || !os.SameFile(lock.info, pathInfo) {
		return errors.New("run lock identity changed; refusing cleanup")
	}
	if err := lock.file.Close(); err != nil {
		return err
	}
	lock.file = nil
	return os.Remove(lock.path)
}

func CreateImmutableJSON(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(value)
	closeErr := file.Close()
	if err != nil {
		return err
	}
	return closeErr
}

func WriteOrVerifyJSON(path string, value any) error {
	wanted, err := json.Marshal(value)
	if err != nil {
		return err
	}
	existing, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return CreateImmutableJSON(path, value)
	}
	if err != nil {
		return err
	}
	var existingValue any
	if err = decodeStrictJSON(filepath.Base(path), existing, &existingValue); err != nil {
		return err
	}
	existingCanonical, err := json.Marshal(existingValue)
	if err != nil {
		return err
	}
	var wantedValue any
	if err = json.Unmarshal(wanted, &wantedValue); err != nil {
		return err
	}
	wantedCanonical, err := json.Marshal(wantedValue)
	if err != nil {
		return err
	}
	if string(existingCanonical) != string(wantedCanonical) {
		return errors.New("existing immutable evidence does not match requested resume")
	}
	return nil
}

func WriteOrVerifyBytes(path string, body []byte) error {
	existing, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return createImmutableBytes(path, body)
	}
	if err != nil {
		return err
	}
	if string(existing) != string(body) {
		return errors.New("existing immutable evidence does not match requested resume")
	}
	return nil
}
