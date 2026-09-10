package portable

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// MachineRequirement is private capability code, never an automatic session hook.
// Check: 0 already prepared, 10 preparation needed, any other exit an error.
// Verify: 0 ready. Prepare must reconcile existing state, not overwrite it.
type MachineRequirement struct {
	ID             string   `json:"id"`
	Description    string   `json:"description"`
	Check          []string `json:"check"`
	Prepare        []string `json:"prepare"`
	Verify         []string `json:"verify"`
	TimeoutSeconds int      `json:"timeout_seconds,omitempty"`
}

type MachinePlan struct {
	CapabilityID   string             `json:"capability_id"`
	Requirement    MachineRequirement `json:"requirement"`
	SHA256         string             `json:"capability_sha256"`
	StateDirectory string             `json:"state_directory"`
}

type MachineReceipt struct {
	Version                int            `json:"schema_version"`
	ID                     string         `json:"id"`
	AssistantID            string         `json:"assistant_id"`
	DeviceID               string         `json:"device_id"`
	CapabilityID           string         `json:"capability_id"`
	RequirementID          string         `json:"requirement_id"`
	SHA256                 string         `json:"capability_sha256"`
	Action                 string         `json:"action"`
	State                  string         `json:"state"`
	StartedAt              string         `json:"started_at"`
	CheckedAt              string         `json:"checked_at,omitempty"`
	Phases                 []MachinePhase `json:"phases"`
	EffectsMayHaveOccurred bool           `json:"effects_may_have_occurred"`
}

type MachinePhase struct {
	Name  string `json:"name"`
	State string `json:"state"`
}

type MachineStatus struct {
	MachinePlan
	State     string `json:"state"`
	CheckedAt string `json:"checked_at,omitempty"`
	ReceiptID string `json:"receipt_id,omitempty"`
}

func sameMachineDefinition(a, b any) bool {
	left, e1 := json.Marshal(a)
	right, e2 := json.Marshal(b)
	return e1 == nil && e2 == nil && bytes.Equal(left, right)
}

func validateMachineRequirements(requirements []MachineRequirement) error {
	ids := map[string]bool{}
	for _, r := range requirements {
		if !identifier.MatchString(r.ID) || ids[r.ID] || strings.TrimSpace(r.Description) == "" || r.TimeoutSeconds < 0 || r.TimeoutSeconds > 3600 {
			return errors.New("invalid or duplicate machine requirement; timeout must be 0–3600 seconds")
		}
		ids[r.ID] = true
		for _, argv := range [][]string{r.Check, r.Prepare, r.Verify} {
			if len(argv) == 0 || strings.TrimSpace(argv[0]) == "" {
				return errors.New("machine check, prepare and verify require nonempty command arrays")
			}
			for _, arg := range argv {
				if strings.ContainsRune(arg, 0) {
					return errors.New("machine command contains NUL")
				}
			}
		}
	}
	return nil
}

// Hash names, modes and bytes; no file contents or script output enter receipts.
func machineFingerprint(root string) (string, error) {
	h := sha256.New()
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if !info.IsDir() && !info.Mode().IsRegular() {
			return errors.New("machine capability contains a symlink or special file")
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if info.Size() > 16<<20 {
			return errors.New("capability source file exceeds 16 MiB")
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		fmt.Fprintf(h, "%d:%s:%o:%d:", len(rel), rel, info.Mode().Perm()&0700, len(data))
		_, _ = h.Write(data)
		return nil
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func (i Instance) machinePlans(s *Store) ([]MachinePlan, error) {
	if i.AssistantID != s.Agent.ID || i.Repository != s.Root {
		return nil, errors.New("machine preparation instance/source mismatch")
	}
	caps, err := s.Capabilities()
	if err != nil {
		return nil, err
	}
	plans := []MachinePlan{}
	for _, c := range caps {
		if len(c.MachineRequirements) == 0 {
			continue
		}
		hash, err := machineFingerprint(filepath.Join(s.Root, "capabilities", c.ID))
		if err != nil {
			return nil, err
		}
		var current Capability
		if err := readJSON(filepath.Join(s.Root, "capabilities", c.ID, "capability.json"), &current); err != nil {
			return nil, err
		}
		if !sameMachineDefinition(current, c) {
			return nil, errors.New("capability manifest changed during inspection; retry a fresh plan")
		}
		for _, r := range c.MachineRequirements {
			plans = append(plans, MachinePlan{CapabilityID: c.ID, Requirement: r, SHA256: hash, StateDirectory: filepath.Join(i.Root, "machine/state", c.ID, r.ID)})
		}
	}
	return plans, nil
}

func (i Instance) MachinePlan(s *Store, capability, requirement string) (MachinePlan, error) {
	plans, err := i.machinePlans(s)
	if err != nil {
		return MachinePlan{}, err
	}
	for _, p := range plans {
		if p.CapabilityID == capability && p.Requirement.ID == requirement {
			return p, nil
		}
	}
	return MachinePlan{}, errors.New("machine requirement not registered; use machine status")
}

// Refuse existing symlink/file parents. Same-user hostile races are not contained.
func (i Instance) machineDirectory(relative string, create bool) (string, error) {
	path := i.Root
	for _, part := range strings.Split("machine/"+relative, "/") {
		if part == "" {
			continue
		}
		if part == "." || part == ".." {
			return "", errors.New("invalid machine state path")
		}
		path = filepath.Join(path, part)
		info, err := os.Lstat(path)
		if os.IsNotExist(err) && create {
			if err = os.Mkdir(path, 0700); err != nil {
				return "", err
			}
			info, err = os.Lstat(path)
		}
		if err != nil {
			return "", err
		}
		if !info.IsDir() {
			return "", errors.New("machine state requires directories, not symlinks or files")
		}
	}
	return path, nil
}

// Metadata-only: no private scripts, credentials, network or local writes.
// Ready means a matching historical receipt, not continuously verified readiness.
func (i Instance) MachineStatus(s *Store) ([]MachineStatus, error) {
	plans, err := i.machinePlans(s)
	if err != nil {
		return nil, err
	}
	result := []MachineStatus{}
	for _, p := range plans {
		st := MachineStatus{MachinePlan: p, State: "unknown"}
		dir, err := i.machineDirectory("receipts/"+p.CapabilityID, false)
		if err == nil {
			var r MachineReceipt
			err = readJSON(filepath.Join(dir, p.Requirement.ID+".json"), &r)
			if err == nil {
				if r.Version != 1 || r.AssistantID != i.AssistantID || r.DeviceID != i.DeviceID || r.CapabilityID != p.CapabilityID || r.RequirementID != p.Requirement.ID || !identifier.MatchString(r.ID) {
					return nil, errors.New("invalid machine readiness receipt")
				}
				st.State, st.CheckedAt, st.ReceiptID = r.State, r.CheckedAt, r.ID
				if r.SHA256 != p.SHA256 {
					st.State = "stale"
				}
			}
		}
		if err != nil && !os.IsNotExist(err) {
			return nil, errors.New("machine readiness metadata is unreadable or unsafe; preserve and inspect local state")
		}
		result = append(result, st)
	}
	return result, nil
}

// RunMachine explicitly executes one reviewed requirement. It never syncs source,
// installs other requirements, invokes a native harness, or replays on its own.
func (i Instance) RunMachine(ctx context.Context, s *Store, reviewed MachinePlan, action string) (MachineReceipt, error) {
	var r MachineReceipt
	if action != "check" && action != "prepare" {
		return r, errors.New("machine action must be check or prepare")
	}
	if os.Getenv("MY_FRIDAY_MACHINE_ACTIVE") != "" {
		return r, errors.New("recursive machine preparation refused")
	}
	if err := ctx.Err(); err != nil {
		return r, err
	}
	p, err := i.MachinePlan(s, reviewed.CapabilityID, reviewed.Requirement.ID)
	if err != nil {
		return r, err
	}
	if p.SHA256 != reviewed.SHA256 {
		return r, errors.New("capability changed since review; inspect a fresh machine plan")
	}
	if _, err := i.MachineStatus(s); err != nil {
		return r, err
	}
	var root string
	var cleanup func()
	err = s.withLock(func() error { var e error; root, cleanup, e = s.snapshotCapability(p.CapabilityID); return e })
	if err != nil {
		return r, err
	}
	defer cleanup()
	hash, err := machineFingerprint(root)
	if err != nil || hash != p.SHA256 {
		return r, errors.New("capability changed during preparation snapshot; review again")
	}
	var captured Capability
	if err := readJSON(filepath.Join(root, "capability.json"), &captured); err != nil {
		return r, err
	}
	matched := false
	for _, requirement := range captured.MachineRequirements {
		if sameMachineDefinition(requirement, p.Requirement) {
			matched = true
		}
	}
	if !matched {
		return r, errors.New("captured requirement differs from reviewed plan")
	}
	dir, err := i.machineDirectory("", true)
	if err != nil {
		return r, err
	}
	fd, err := syscall.Open(filepath.Join(dir, "run.lock"), syscall.O_CREAT|syscall.O_RDWR|syscall.O_NOFOLLOW, 0600)
	if err != nil {
		return r, err
	}
	defer syscall.Close(fd)
	var lockInfo syscall.Stat_t
	if err := syscall.Fstat(fd, &lockInfo); err != nil || lockInfo.Mode&syscall.S_IFMT != syscall.S_IFREG {
		return r, errors.New("machine lock must be a regular file")
	}
	if err = syscall.Flock(fd, syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		return r, errors.New("another machine check/preparation is active; inspect status before retrying")
	}
	defer syscall.Flock(fd, syscall.LOCK_UN)
	if _, err = i.machineDirectory("state/"+p.CapabilityID+"/"+p.Requirement.ID, true); err != nil {
		return r, err
	}
	receipts, err := i.machineDirectory("receipts/"+p.CapabilityID, true)
	if err != nil {
		return r, err
	}
	runs, err := i.machineDirectory("runs", true)
	if err != nil {
		return r, err
	}
	r = MachineReceipt{Version: 1, ID: NewID("machine"), AssistantID: i.AssistantID, DeviceID: i.DeviceID, CapabilityID: p.CapabilityID, RequirementID: p.Requirement.ID, SHA256: p.SHA256, Action: action, State: "running", StartedAt: time.Now().UTC().Format(time.RFC3339Nano), Phases: []MachinePhase{}}
	save := func() error {
		for _, target := range []string{filepath.Join(runs, r.ID+".json"), filepath.Join(receipts, p.Requirement.ID+".json")} {
			info, err := os.Lstat(target)
			if err != nil && !os.IsNotExist(err) {
				return err
			}
			if err == nil && !info.Mode().IsRegular() {
				return errors.New("machine receipt target must be a regular file; existing target preserved")
			}
		}
		if err := writeLocalJSON(filepath.Join(runs, r.ID+".json"), r); err != nil {
			return err
		}
		return writeLocalJSON(filepath.Join(receipts, p.Requirement.ID+".json"), r)
	}
	if err := save(); err != nil {
		return r, err
	}
	run := func(phase string, argv []string) (int, error) {
		if err := ctx.Err(); err != nil {
			return -1, err
		}
		r.Phases = append(r.Phases, MachinePhase{Name: phase, State: "running"})
		// Private scripts have full user access, even checks. No rollback promised.
		r.EffectsMayHaveOccurred = true
		if err := save(); err != nil {
			return -1, err
		}
		code, err := runMachineCommand(ctx, root, s, i, p, phase, argv)
		state := "passed"
		if code == 10 && phase == "check" && err == nil {
			state = "needs-preparation"
		} else if err != nil || code != 0 {
			state = "failed"
		}
		r.Phases[len(r.Phases)-1].State = state
		if e := save(); e != nil {
			return -1, e
		}
		return code, err
	}
	finish := func(state string, cause error) (MachineReceipt, error) {
		r.State, r.CheckedAt = state, time.Now().UTC().Format(time.RFC3339Nano)
		if err := save(); err != nil {
			return r, errors.New("machine receipt could not be saved; inspect local state before retrying")
		}
		return r, cause
	}
	fail := func() (MachineReceipt, error) {
		return finish("failed", errors.New("machine phase failed or was cancelled; raw output suppressed; inspect receipt and private instructions before retrying"))
	}
	code, err := run("check", p.Requirement.Check)
	if err != nil || (code != 0 && code != 10) {
		return fail()
	}
	if code == 10 {
		if action == "check" {
			return finish("needs-preparation", nil)
		}
		if code, err = run("prepare", p.Requirement.Prepare); err != nil || code != 0 {
			return fail()
		}
	}
	if code, err = run("verify", p.Requirement.Verify); err != nil || code != 0 {
		return fail()
	}
	return finish("ready", nil)
}

func runMachineCommand(parent context.Context, root string, s *Store, i Instance, p MachinePlan, phase string, argv []string) (int, error) {
	seconds := p.Requirement.TimeoutSeconds
	if seconds == 0 {
		seconds = 300
	}
	ctx, cancel := context.WithTimeout(parent, time.Duration(seconds)*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.Dir = root
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if !strings.HasPrefix(key, "MY_FRIDAY_") {
			cmd.Env = append(cmd.Env, entry)
		}
	}
	binary, err := os.Executable()
	if err != nil {
		return -1, err
	}
	cmd.Env = append(cmd.Env, "MY_FRIDAY_BIN="+binary, "MY_FRIDAY_ASSISTANT_ROOT="+s.Root, "MY_FRIDAY_ASSISTANT_ID="+s.Agent.ID, "MY_FRIDAY_DEVICE_ID="+i.DeviceID, "MY_FRIDAY_INSTANCE="+i.Root, "MY_FRIDAY_MACHINE_ACTIVE=1", "MY_FRIDAY_MACHINE_STATE="+p.StateDirectory, "MY_FRIDAY_MACHINE_INTERACTIVE=false")
	payload, _ := json.Marshal(map[string]any{"schema_version": 1, "phase": phase, "capability_id": p.CapabilityID, "requirement_id": p.Requirement.ID, "device_id": i.DeviceID, "state_directory": p.StateDirectory, "interactive": false})
	cmd.Stdin = strings.NewReader(string(payload))
	cmd.Stdout, cmd.Stderr = io.Discard, io.Discard
	configureCommandCancellation(cmd)
	err = cmd.Run()
	if ctx.Err() != nil {
		return -1, ctx.Err()
	}
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		return exit.ExitCode(), nil
	}
	if err != nil {
		return -1, err
	}
	return 0, nil
}
