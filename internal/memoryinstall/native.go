package memoryinstall

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/acoz-labs/my-friday/internal/memorybank"
	"github.com/acoz-labs/my-friday/internal/toolkitupdate"
)

type runner func(context.Context, Options, ...string) ([]byte, error)
type boundedOutput struct{ bytes.Buffer }

func (b *boundedOutput) Write(p []byte) (int, error) {
	if b.Len()+len(p) > 1<<20 {
		return 0, errors.New("native response exceeds 1 MiB")
	}
	return b.Buffer.Write(p)
}

func native(ctx context.Context, o Options, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, o.Codex, args...)
	cmd.Dir = o.Home
	for _, v := range os.Environ() {
		key, _, _ := strings.Cut(v, "=")
		if key != "CODEX_HOME" {
			cmd.Env = append(cmd.Env, v)
		}
	}
	cmd.Env = append(cmd.Env, "CODEX_HOME="+o.CodexHome)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return nil
		}
		err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		if err == syscall.ESRCH {
			return os.ErrProcessDone
		}
		return err
	}
	cmd.WaitDelay = time.Second
	var out boundedOutput
	cmd.Stdout = &out
	// Do not expose native stderr, which may contain host/private configuration.
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("native Codex command %s failed or timed out; inspect that command in the selected profile (raw output suppressed)", strings.Join(args, " "))
	}
	return out.Bytes(), nil
}

type marketplace struct {
	Name   string `json:"name"`
	Root   string `json:"root"`
	Source struct {
		Type string `json:"sourceType"`
		Path string `json:"source"`
	} `json:"marketplaceSource"`
}
type plugin struct {
	ID      string `json:"pluginId"`
	Name    string `json:"name"`
	Version string `json:"version"`
	Enabled bool   `json:"enabled"`
	Source  struct {
		Source string `json:"source"`
	} `json:"marketplaceSource"`
}

func marketplaces(ctx context.Context, o Options, run runner) ([]marketplace, error) {
	raw, err := run(ctx, o, "plugin", "marketplace", "list", "--json")
	if err != nil {
		return nil, err
	}
	var v struct {
		Items []marketplace `json:"marketplaces"`
	}
	if json.Unmarshal(raw, &v) != nil || v.Items == nil {
		return nil, errors.New("unsupported native marketplace response")
	}
	return v.Items, nil
}
func plugins(ctx context.Context, o Options, run runner) ([]plugin, error) {
	raw, err := run(ctx, o, "plugin", "list", "--json")
	if err != nil {
		return nil, err
	}
	var v struct {
		Items []plugin `json:"installed"`
	}
	if json.Unmarshal(raw, &v) != nil || v.Items == nil {
		return nil, errors.New("unsupported native plugin response")
	}
	return v.Items, nil
}

func loadPlan(root string) (Plan, error) {
	var p Plan
	st, err := os.Lstat(filepath.Join(root, "connection.json"))
	if err != nil {
		return p, err
	}
	if !st.Mode().IsRegular() || st.Size() > 32768 {
		return p, errors.New("invalid connection receipt")
	}
	raw, err := os.ReadFile(filepath.Join(root, "connection.json"))
	if err != nil {
		return p, err
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err = d.Decode(&p); err != nil || p.SchemaVersion != 1 || p.Root != root {
		return Plan{}, errors.New("invalid managed connection receipt")
	}
	var extra any
	if err := d.Decode(&extra); err != io.EOF {
		return Plan{}, errors.New("connection receipt contains trailing data")
	}
	return p, nil
}

func checkCollision(ctx context.Context, p Plan, run runner) error {
	ms, err := marketplaces(ctx, p.Options, run)
	if err != nil {
		return err
	}
	for _, m := range ms {
		if m.Name != p.Marketplace {
			continue
		}
		if m.Source.Type != "local" || m.Source.Path != m.Root || filepath.Dir(m.Root) != managedRoot(p.Options) {
			return errors.New("marketplace name is already owned by another source; nothing installed or replaced")
		}
		prior, err := loadPlan(m.Root)
		if err != nil || prior.CodexHome != p.CodexHome || prior.Home != p.Home || prior.Marketplace != p.Marketplace {
			return errors.New("marketplace name is already owned by another source; nothing installed or replaced")
		}
		// Later source packages may differ. Ownership is determined by the local
		// receipt and managed root; current package bytes are checked after staging.
	}
	ps, err := plugins(ctx, p.Options, run)
	if err != nil {
		return err
	}
	for _, v := range ps {
		if v.Name == "my-friday-memory" && v.ID != p.PluginID {
			return errors.New("another My Friday memory plugin is installed in this profile; choose a separate profile or remove that duplicate explicitly")
		}
	}
	return nil
}

type Result struct {
	Plan                 Plan   `json:"connection"`
	Installed            bool   `json:"installed"`
	RequiresFreshSession bool   `json:"requires_fresh_session"`
	Notice               string `json:"notice"`
	PreviousRoot         string `json:"previous_marketplace_root,omitempty"`
}

func Apply(ctx context.Context, p Plan) (Result, error) { return apply(ctx, p, native) }
func apply(ctx context.Context, p Plan, run runner) (Result, error) {
	result := Result{Plan: p, RequiresFreshSession: true}
	if err := ctx.Err(); err != nil {
		return result, err
	}
	current, err := Prepare(p.Options)
	if err != nil {
		return result, err
	}
	if current != p {
		return result, errors.New("installation inputs changed after preview; preview again")
	}
	if err = os.MkdirAll(p.CodexHome, 0700); err != nil {
		return result, err
	}
	if err = checkCollision(ctx, p, run); err != nil {
		return result, err
	}
	parent := managedRoot(p.Options)
	if err = os.MkdirAll(parent, 0700); err != nil {
		return result, err
	}
	lock := filepath.Join(parent, ".install-lock")
	f, err := os.OpenFile(lock, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return result, errors.New("connection is locked; another installer may be active; inspect before removing a stale lock")
	}
	f.Close()
	defer os.Remove(lock)
	if err = checkCollision(ctx, p, run); err != nil {
		return result, err
	}
	if _, err = toolkitupdate.Stage(p.Home, p.Binary, p.BinarySHA256); err != nil {
		return result, err
	}
	checkOptions := p.Options
	checkOptions.Codex = p.Runtime
	raw, err := native(ctx, checkOptions, "--version")
	var version struct {
		Product        string `json:"product"`
		MemoryProtocol int    `json:"memory_protocol"`
		OS             string `json:"os"`
		Arch           string `json:"arch"`
	}
	if err != nil || json.Unmarshal(raw, &version) != nil || version.Product != "my-friday" || version.MemoryProtocol != 1 || version.OS != runtime.GOOS || version.Arch != runtime.GOARCH {
		return result, errors.New("selected runtime does not support this machine's memory protocol; no native connection was changed")
	}
	if err = publish(p); err != nil {
		return result, err
	}
	// Native Codex refuses to add a second source under an existing name. For
	// our verified local-only registration, remove the registration (not the
	// source/cache files), then add the retained replacement. No config editing.
	ms, err := marketplaces(ctx, p.Options, run)
	if err != nil {
		return result, err
	}
	for _, m := range ms {
		if m.Name == p.Marketplace && m.Root != p.Root {
			if err = checkCollision(ctx, p, run); err != nil {
				return result, err
			}
			result.PreviousRoot = m.Root
			if _, err = run(ctx, p.Options, "plugin", "marketplace", "remove", p.Marketplace, "--json"); err != nil {
				return result, err
			}
		}
	}
	if _, err = run(ctx, p.Options, "plugin", "marketplace", "add", p.Root); err != nil {
		result.Notice = "Connection registration is incomplete. Retained runtime/plugin copies are preserved. Rerun connect with the same bank/runtime; no automatic rollback or authentication changes were attempted."
		return result, err
	}
	if _, err = run(ctx, p.Options, "plugin", "add", p.PluginID, "--json"); err != nil {
		return result, fmt.Errorf("marketplace is registered but plugin installation is incomplete; rerun connect to retry: %w", err)
	}
	ps, err := plugins(ctx, p.Options, run)
	if err != nil {
		return result, err
	}
	for _, v := range ps {
		if v.ID == p.PluginID && v.Enabled && v.Version == p.Version && v.Source.Source == p.Root {
			result.Installed = true
		}
	}
	if !result.Installed {
		return result, errors.New("native installation did not report the selected plugin enabled at the expected version/source; inspect before retrying")
	}
	result.Notice = "Connected for fresh Codex sessions. Review native hook trust and MCP status. No authentication, memory, synchronization, shell startup files or existing sessions were changed. Explicit MY_FRIDAY_MEMORY_BIN/BINDING environment overrides still take precedence."
	return result, nil
}

type Check struct {
	Name   string `json:"name"`
	OK     bool   `json:"ok"`
	Detail string `json:"detail"`
}
type Report struct {
	Healthy    bool    `json:"healthy"`
	Checks     []Check `json:"checks"`
	Notice     string  `json:"notice"`
	Connection *Plan   `json:"connection,omitempty"`
}

func Doctor(ctx context.Context, o Options) Report { return doctor(ctx, o, native) }
func doctor(ctx context.Context, o Options, run runner) Report {
	r := Report{Healthy: true, Checks: []Check{}, Notice: "Read-only local receipt, file-integrity, binding and native plugin inventory checks. Hook trust, live MCP startup, native login, active-session context, remote freshness and model behavior are not tested."}
	add := func(name string, err error) {
		c := Check{Name: name, OK: err == nil, Detail: "Passed"}
		if err != nil {
			c.Detail = err.Error()
			r.Healthy = false
		}
		r.Checks = append(r.Checks, c)
	}
	for _, path := range []*string{&o.Home, &o.CodexHome, &o.Codex} {
		value, err := canonical(*path)
		add("profile-path", err)
		if err != nil {
			return r
		}
		*path = value
	}
	ms, err := marketplaces(ctx, o, run)
	add("native-marketplaces", err)
	if err != nil {
		return r
	}
	var p Plan
	found := false
	for _, m := range ms {
		if m.Source.Type != "local" || m.Source.Path != m.Root || filepath.Dir(m.Root) != managedRoot(o) {
			continue
		}
		v, e := loadPlan(m.Root)
		if e == nil && v.CodexHome == o.CodexHome && filepath.Dir(m.Root) == managedRoot(o) {
			p = v
			found = true
			break
		}
	}
	if !found {
		add("connection", errors.New("no managed memory connection found for this native profile; run bank connect-codex"))
		return r
	}
	r.Connection = &p
	add("bundle-integrity", verifyBundle(p))
	add("native-cache-integrity", verifyCache(p))
	digest, err := toolkitupdate.Digest(p.Runtime)
	if err == nil && digest != p.BinarySHA256 {
		err = errors.New("pinned runtime bytes changed")
	}
	add("runtime-integrity", err)
	s, err := memorybank.OpenBinding(p.Binding, "doctor")
	if err == nil && s.ID() != p.BankID {
		err = errors.New("selected binding now points to a different bank")
	}
	add("bank-binding", err)
	for _, override := range []struct{ Key, Expected string }{{"MY_FRIDAY_MEMORY_BIN", p.Runtime}, {"MY_FRIDAY_MEMORY_BINDING", p.Binding}} {
		if value := os.Getenv(override.Key); value != "" {
			resolved, e := canonical(value)
			if e != nil || resolved != override.Expected {
				add("environment-override", fmt.Errorf("%s overrides this connection in the current process; clear it for the pinned default or deliberately select the other connection", override.Key))
			}
		}
	}
	ps, err := plugins(ctx, o, run)
	add("native-plugin-inventory", err)
	if err == nil {
		found = false
		for _, v := range ps {
			if v.ID == p.PluginID && v.Enabled && v.Version == p.Version && v.Source.Source == p.Root {
				found = true
			}
		}
		if !found {
			add("native-plugin", errors.New("plugin missing, disabled or stale; reconnect after closing affected sessions"))
		} else {
			add("native-plugin", nil)
		}
	}
	return r
}
