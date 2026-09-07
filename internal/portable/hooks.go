package portable

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"
)

type Capability struct {
	Version       int            `json:"schema_version"`
	ID            string         `json:"id"`
	Description   string         `json:"description"`
	Subscriptions []Subscription `json:"subscriptions"`
	Checks        [][]string     `json:"checks,omitempty"`
}
type Subscription struct {
	ID             string   `json:"id"`
	Event          string   `json:"event"`
	Command        []string `json:"command"`
	After          []string `json:"after,omitempty"`
	TimeoutSeconds int      `json:"timeout_seconds,omitempty"`
	Failure        string   `json:"failure,omitempty"`
}
type Event struct {
	Version          int             `json:"schema_version"`
	ID               string          `json:"id"`
	Name             string          `json:"event"`
	AssistantID      string          `json:"assistant_id"`
	DeviceID         string          `json:"device_id"`
	SessionID        string          `json:"session_id,omitempty"`
	RequestID        string          `json:"request_id,omitempty"`
	NativeEvent      string          `json:"native_event,omitempty"`
	WorkingDirectory string          `json:"working_directory,omitempty"`
	CausationID      string          `json:"causation_id,omitempty"`
	Payload          json.RawMessage `json:"payload,omitempty"`
}
type HandlerResult struct {
	ID      string `json:"id"`
	Success bool   `json:"success"`
	Detail  string `json:"detail,omitempty"`
}
type DispatchResult struct {
	EventID   string          `json:"event_id"`
	Duplicate bool            `json:"duplicate"`
	Handlers  []HandlerResult `json:"handlers"`
	Context   []string        `json:"additional_context"`
}

func (s *Store) Capabilities() ([]Capability, error) {
	entries, err := os.ReadDir(filepath.Join(s.Root, "capabilities"))
	if err != nil {
		return nil, err
	}
	result := []Capability{}
	for _, entry := range entries {
		if entry.Name() == ".gitkeep" {
			continue
		}
		if !entry.IsDir() || !identifier.MatchString(entry.Name()) {
			return nil, fmt.Errorf("invalid capability directory: %s", entry.Name())
		}
		var c Capability
		if err = readJSON(filepath.Join(s.Root, "capabilities", entry.Name(), "capability.json"), &c); err != nil {
			return nil, err
		}
		if c.Version != 1 || c.ID != entry.Name() || strings.TrimSpace(c.Description) == "" {
			return nil, errors.New("invalid capability manifest")
		}
		ids := map[string]bool{}
		for _, sub := range c.Subscriptions {
			if !identifier.MatchString(sub.ID) || sub.Event == "" || len(sub.Command) == 0 || sub.Command[0] == "" || sub.TimeoutSeconds < 0 || sub.TimeoutSeconds > 120 {
				return nil, errors.New("invalid subscription")
			}
			if sub.Failure != "" && sub.Failure != "warn" && sub.Failure != "stop" {
				return nil, errors.New("subscription failure must be warn or stop")
			}
			if ids[sub.ID] {
				return nil, errors.New("duplicate subscription ID")
			}
			ids[sub.ID] = true
		}
		result = append(result, c)
	}
	return result, nil
}

type handler struct {
	CapabilityID string
	Subscription
}

func orderedHandlers(caps []Capability, event string) ([]handler, error) {
	all := map[string]handler{}
	for _, c := range caps {
		for _, sub := range c.Subscriptions {
			if sub.Event == event {
				key := c.ID + "/" + sub.ID
				all[key] = handler{c.ID, sub}
			}
		}
	}
	keys := []string{}
	for id := range all {
		keys = append(keys, id)
	}
	sort.Strings(keys)
	result := []handler{}
	visiting := map[string]bool{}
	done := map[string]bool{}
	var visit func(string) error
	visit = func(id string) error {
		if visiting[id] {
			return errors.New("subscription ordering cycle")
		}
		if done[id] {
			return nil
		}
		h, ok := all[id]
		if !ok {
			return fmt.Errorf("subscription dependency not registered for this event: %s", id)
		}
		visiting[id] = true
		for _, dependency := range h.After {
			if !strings.Contains(dependency, "/") {
				dependency = h.CapabilityID + "/" + dependency
			}
			if err := visit(dependency); err != nil {
				return err
			}
		}
		visiting[id] = false
		done[id] = true
		result = append(result, h)
		return nil
	}
	for _, id := range keys {
		if err := visit(id); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func (s *Store) Dispatch(ctx context.Context, event Event) (DispatchResult, error) {
	result := DispatchResult{EventID: event.ID, Handlers: []HandlerResult{}, Context: []string{}}
	if event.Version != 1 || !identifier.MatchString(event.ID) || event.Name == "" || event.AssistantID != s.Agent.ID {
		return result, errors.New("invalid lifecycle event")
	}
	if event.CausationID == event.ID || os.Getenv("MY_FRIDAY_DISPATCH_ACTIVE") == s.Agent.ID {
		return result, errors.New("recursive lifecycle dispatch refused")
	}
	if err := s.deviceExists(event.DeviceID); err != nil {
		return result, err
	}
	caps, err := s.Capabilities()
	if err != nil {
		return result, err
	}
	handlers, err := orderedHandlers(caps, event.Name)
	if err != nil {
		return result, err
	}
	if len(handlers) == 0 {
		return result, nil
	}
	receipt := filepath.Join(s.Root, ".my-friday/local/events", event.ID+".json")
	err = s.withLock(func() error {
		if err := os.MkdirAll(filepath.Dir(receipt), 0700); err != nil {
			return err
		}
		if _, err := os.Lstat(receipt); err == nil {
			var status struct {
				EventID  string          `json:"event_id"`
				State    string          `json:"state"`
				Handlers []HandlerResult `json:"handlers"`
			}
			if err = readJSON(receipt, &status); err != nil {
				return err
			}
			if status.State != "completed" {
				return errors.New("event was interrupted or failed; inspect its receipt before replaying external effects")
			}
			result.Duplicate = true
			return nil
		}
		return writeNewJSON(receipt, map[string]any{"event_id": event.ID, "state": "running", "handlers": []HandlerResult{}})
	})
	if err != nil || result.Duplicate {
		return result, err
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return result, err
	}
	for _, h := range handlers {
		var root string
		cleanup := func() {}
		err := s.withLock(func() error {
			var snapshotErr error
			root, cleanup, snapshotErr = s.snapshotCapability(h.CapabilityID)
			return snapshotErr
		})
		if err != nil {
			return result, err
		}
		timeout := h.TimeoutSeconds
		if timeout == 0 {
			timeout = 10
		}
		handlerCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
		output, runErr := runHandler(handlerCtx, root, s, event.DeviceID, h.Command, payload, false)
		cancel()
		cleanup()
		hr := HandlerResult{ID: h.CapabilityID + "/" + h.ID, Success: runErr == nil}
		if runErr != nil {
			hr.Detail = "Handler failed, timed out, or returned invalid output; raw output was not retained."
		} else if output != "" {
			result.Context = append(result.Context, output)
		}
		result.Handlers = append(result.Handlers, hr)
		if runErr != nil && h.Failure == "stop" {
			_ = writeLocalJSON(receipt, map[string]any{"event_id": event.ID, "state": "failed", "handlers": result.Handlers})
			return result, fmt.Errorf("required lifecycle handler failed: %s", hr.ID)
		}
	}
	err = writeLocalJSON(receipt, map[string]any{"event_id": event.ID, "state": "completed", "handlers": result.Handlers})
	return result, err
}

func writeLocalJSON(path string, value any) error {
	b, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".local-write-")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err = tmp.Write(append(b, '\n')); err != nil {
		tmp.Close()
		return err
	}
	if err = tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err = tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}

func (s *Store) snapshotCapability(id string) (string, func(), error) {
	root, err := os.MkdirTemp("", "my-friday-capability-")
	if err != nil {
		return "", func() {}, err
	}
	cleanup := func() { os.RemoveAll(root) }
	source := filepath.Join(s.Root, "capabilities", id)
	err = filepath.WalkDir(source, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := filepath.Join(root, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0700)
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return errors.New("capability contains a symlink or special file")
		}
		if info.Size() > 16<<20 {
			return errors.New("capability source file exceeds 16 MiB")
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, info.Mode().Perm()&0700)
	})
	if err != nil {
		cleanup()
		return "", func() {}, err
	}
	return root, cleanup, nil
}

type boundedOutput struct {
	bytes.Buffer
	Limit int
}

func (b *boundedOutput) Write(p []byte) (int, error) {
	n := len(p)
	left := b.Limit - b.Len()
	if left > 0 {
		if left > n {
			left = n
		}
		_, _ = b.Buffer.Write(p[:left])
	}
	return n, nil
}

func runHandler(ctx context.Context, root string, s *Store, device string, args []string, payload []byte, check bool) (string, error) {
	binary, err := os.Executable()
	if err != nil {
		return "", err
	}
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Dir = root
	env := []string{}
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if strings.HasPrefix(key, "MY_FRIDAY_") {
			continue
		}
		env = append(env, entry)
	}
	cmd.Env = append(env, "MY_FRIDAY_ASSISTANT_ROOT="+s.Root, "MY_FRIDAY_BIN="+binary, "MY_FRIDAY_ASSISTANT_ID="+s.Agent.ID, "MY_FRIDAY_DEVICE_ID="+device, "MY_FRIDAY_DISPATCH_ACTIVE="+s.Agent.ID)
	cmd.Stdin = bytes.NewReader(payload)
	output := &boundedOutput{Limit: 16385}
	cmd.Stdout = output
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return nil
		}
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
	cmd.WaitDelay = time.Second
	if err := cmd.Run(); err != nil {
		return "", err
	}
	if check {
		return "", nil
	}
	if output.Len() > 16384 {
		return "", errors.New("hook output exceeds limit")
	}
	if strings.TrimSpace(output.String()) == "" {
		return "", nil
	}
	var response struct {
		Context string `json:"additional_context"`
	}
	decoder := json.NewDecoder(bytes.NewReader(output.Bytes()))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&response); err != nil {
		return "", err
	}
	return response.Context, nil
}
