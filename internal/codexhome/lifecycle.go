// Package codexhome manages the deliberately tiny installed Codex baseline.
// Its authority is limited to AGENTS.md and .my-friday beneath an injected
// Codex home; it never enumerates or edits unrelated Codex state.
package codexhome

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/acoz-labs/my-friday/internal/profile"
	"github.com/acoz-labs/my-friday/internal/repository"
)

const (
	controlDir             = ".my-friday"
	manifestFile           = "installed-baseline.json"
	previousFile           = "previous-AGENTS.md"
	journalFile            = "transaction.json"
	recoveryProjectionFile = "recovery-AGENTS.md"
	recoveryManifestFile   = "recovery-manifest.json"
)

type Action string

const (
	ActionInstall   Action = "install"
	ActionRepair    Action = "repair"
	ActionUpgrade   Action = "upgrade"
	ActionRollback  Action = "rollback"
	ActionUninstall Action = "uninstall"
)

type State string

const (
	StateNotInstalled State = "not-installed"
	StateHealthy      State = "healthy"
	StateCollision    State = "collision"
	StateDrift        State = "managed-drift"
	StateSourceDrift  State = "source-drift"
	StateInterrupted  State = "interrupted"
)

type Status struct {
	State  State
	Detail string
}
type Change struct{ Operation, Path, SHA256 string }
type PlanResult struct {
	Action             Action
	Runtime, CodexHome string
	Changes            []Change
	NegativeActions    []string
	projection         []byte
	manifest           manifest
}

func (p PlanResult) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Action: %s\nCodex home: %s\n", p.Action, p.CodexHome)
	for _, c := range p.Changes {
		fmt.Fprintf(&b, "- %s %s (sha256:%s)\n", c.Operation, c.Path, c.SHA256)
	}
	for _, n := range p.NegativeActions {
		fmt.Fprintf(&b, "- Will not %s\n", n)
	}
	return b.String()
}

type manifest struct {
	ContractVersion  int    `json:"contract_version"`
	AssistantID      string `json:"assistant_id"`
	ProjectionSHA256 string `json:"projection_sha256"`
	SourcePath       string `json:"source_path"`
	SourceSHA256     string `json:"source_sha256"`
	PreviousSHA256   string `json:"previous_sha256,omitempty"`
}
type runtimeManifest struct {
	ContractVersion int    `json:"contract_version"`
	RepositoryRole  string `json:"repository_role"`
	AssistantID     string `json:"assistant_id"`
}
type journal struct {
	ContractVersion int    `json:"contract_version"`
	Action          Action `json:"action"`
	Phase           string `json:"phase"`
}

func digest(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func paths(home string) (string, string, string, string) {
	return filepath.Join(home, "AGENTS.md"), filepath.Join(home, controlDir, manifestFile), filepath.Join(home, controlDir, previousFile), filepath.Join(home, controlDir, journalFile)
}
func safeRegular(path string) ([]byte, error) {
	i, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !i.Mode().IsRegular() || i.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("unsafe non-regular path: %s", path)
	}
	if st, ok := i.Sys().(*syscall.Stat_t); ok && st.Nlink != 1 {
		return nil, fmt.Errorf("unsafe linked path: %s", path)
	}
	if st, ok := i.Sys().(*syscall.Stat_t); ok && st.Uid != uint32(os.Getuid()) {
		return nil, fmt.Errorf("unsafe foreign owner: %s", path)
	}
	if i.Mode().Perm() != 0600 {
		return nil, fmt.Errorf("unsafe file mode: %s", path)
	}
	return os.ReadFile(path)
}
func canonicalHome(home string) (string, error) {
	if home == "" {
		return "", errors.New("Codex home is required")
	}
	abs, err := filepath.Abs(home)
	if err != nil {
		return "", err
	}
	abs = filepath.Clean(abs)
	i, err := os.Lstat(abs)
	if err != nil {
		return "", fmt.Errorf("Codex home must already exist: %w", err)
	}
	if !i.IsDir() || i.Mode()&os.ModeSymlink != 0 {
		return "", errors.New("Codex home must be a non-symlink directory")
	}
	if st, ok := i.Sys().(*syscall.Stat_t); ok && st.Uid != uint32(os.Getuid()) {
		return "", errors.New("Codex home must be owned by the current user")
	}
	return abs, nil
}
func render(runtime string) ([]byte, runtimeManifest, error) {
	abs, err := filepath.Abs(runtime)
	if err != nil {
		return nil, runtimeManifest{}, err
	}
	assistantID, err := repository.ValidateRuntime(abs)
	if err != nil {
		return nil, runtimeManifest{}, fmt.Errorf("runtime repository: %w", err)
	}
	mb, err := safeRegular(filepath.Join(abs, ".my-friday", "manifest.json"))
	if err != nil {
		return nil, runtimeManifest{}, fmt.Errorf("runtime manifest: %w", err)
	}
	var rm runtimeManifest
	if json.Unmarshal(mb, &rm) != nil || rm.ContractVersion != 1 || rm.RepositoryRole != "runtime" || !strings.HasPrefix(rm.AssistantID, "asst-") {
		return nil, rm, errors.New("runtime repository manifest is incompatible")
	}
	if rm.AssistantID != assistantID {
		return nil, rm, errors.New("runtime assistant identity mismatch")
	}
	pb, err := safeRegular(filepath.Join(abs, "assistant", "profile.json"))
	if err != nil {
		return nil, rm, err
	}
	var p profile.Profile
	if json.Unmarshal(pb, &p) != nil || profile.Validate(p) != nil || p.AssistantID != assistantID {
		return nil, rm, errors.New("runtime profile is incompatible")
	}
	var out strings.Builder
	out.WriteString("# Managed Assistant Instructions\n\n")
	fmt.Fprintf(&out, "Your display name is %s.\n\nPurpose: %s\n\n", p.Identity.DisplayName, p.Identity.Purpose)
	if p.Identity.AddressUserAs != nil {
		fmt.Fprintf(&out, "Address the user as %s.\n\n", *p.Identity.AddressUserAs)
	}
	fmt.Fprintf(&out, "Communication style: %s", p.Communication.Preset)
	if p.Communication.CustomGuidance != nil {
		fmt.Fprintf(&out, " — %s", *p.Communication.CustomGuidance)
	}
	out.WriteString(".\n\nThese presentation preferences never override authorization, safety, trust, privacy, or tool policy.\n")
	return []byte(out.String()), rm, nil
}
func loadManifest(path string) (manifest, error) {
	b, err := safeRegular(path)
	if err != nil {
		return manifest{}, err
	}
	var m manifest
	if json.Unmarshal(b, &m) != nil || m.ContractVersion != 1 || m.AssistantID == "" || len(m.ProjectionSHA256) != 64 {
		return m, errors.New("installed manifest is incompatible")
	}
	return m, nil
}
func Inspect(runtime, home string) (Status, error) {
	h, err := canonicalHome(home)
	if err != nil {
		return Status{}, err
	}
	projection, mp, _, jp := paths(h)
	control := filepath.Join(h, controlDir)
	if _, err = os.Lstat(jp); err == nil {
		return Status{StateInterrupted, "transaction journal exists"}, nil
	}
	_, pe := os.Lstat(projection)
	_, me := os.Lstat(mp)
	_, ce := os.Lstat(control)
	if os.IsNotExist(pe) && os.IsNotExist(me) {
		if ce == nil {
			return Status{StateCollision, "foreign .my-friday control namespace"}, nil
		}
		return Status{StateNotInstalled, "no managed projection"}, nil
	}
	if pe != nil || me != nil {
		return Status{StateCollision, "projection and ownership manifest disagree"}, nil
	}
	m, err := loadManifest(mp)
	if err != nil {
		return Status{StateCollision, err.Error()}, nil
	}
	b, err := safeRegular(projection)
	if err != nil {
		return Status{StateCollision, err.Error()}, nil
	}
	if digest(b) != m.ProjectionSHA256 {
		return Status{StateDrift, "projection digest differs from manifest"}, nil
	}
	if runtime == "" {
		runtime = m.SourcePath
	}
	if runtime != "" {
		rb, rm, re := render(runtime)
		if re != nil || rm.AssistantID != m.AssistantID || digest(rb) != m.SourceSHA256 {
			return Status{StateSourceDrift, "installed state is healthy but source differs"}, nil
		}
	}
	return Status{StateHealthy, "manifest and projection agree"}, nil
}

func Plan(action Action, runtime, home string) (PlanResult, error) {
	h, err := canonicalHome(home)
	if err != nil {
		return PlanResult{}, err
	}
	projection, mp, pp, _ := paths(h)
	s, err := Inspect(runtime, h)
	if err != nil {
		return PlanResult{}, err
	}
	p := PlanResult{Action: action, Runtime: runtime, CodexHome: h, NegativeActions: []string{"edit config.toml, auth, sessions, logs, skills, packages, or project configuration", "start a daemon, use the network, or request privilege"}}
	switch action {
	case ActionInstall:
		if s.State != StateNotInstalled {
			return p, fmt.Errorf("install refused: %s", s.State)
		}
		if _, e := os.Lstat(filepath.Join(h, "AGENTS.override.md")); e == nil {
			return p, errors.New("install refused: AGENTS.override.md would shadow the managed projection")
		} else if !os.IsNotExist(e) {
			return p, e
		}
		p.projection, _, err = render(runtime)
		if err != nil {
			return p, err
		}
		var rm runtimeManifest
		_, rm, _ = render(runtime)
		p.manifest = manifest{1, rm.AssistantID, digest(p.projection), runtime, digest(p.projection), ""}
		p.Changes = []Change{{"create", projection, digest(p.projection)}, {"create", mp, "manifest"}}
	case ActionRepair:
		if s.State != StateDrift {
			return p, fmt.Errorf("repair requires managed drift; found %s", s.State)
		}
		old, e := loadManifest(mp)
		if e != nil {
			return p, e
		}
		if runtime == "" {
			runtime = old.SourcePath
			p.Runtime = runtime
		}
		var rm runtimeManifest
		p.projection, rm, e = render(runtime)
		if e != nil {
			return p, e
		}
		if rm.AssistantID != old.AssistantID {
			return p, errors.New("assistant identity mismatch")
		}
		p.manifest = old
		p.manifest.ProjectionSHA256 = digest(p.projection)
		p.manifest.SourceSHA256 = digest(p.projection)
		p.manifest.SourcePath = runtime
		p.Changes = []Change{{"replace drifted managed file", projection, digest(p.projection)}}
	case ActionUpgrade:
		if s.State != StateSourceDrift {
			return p, fmt.Errorf("upgrade requires healthy installed state and changed source; found %s", s.State)
		}
		old, e := loadManifest(mp)
		if e != nil {
			return p, e
		}
		current, e := safeRegular(projection)
		if e != nil {
			return p, e
		}
		var rm runtimeManifest
		p.projection, rm, e = render(runtime)
		if e != nil {
			return p, e
		}
		if rm.AssistantID != old.AssistantID {
			return p, errors.New("assistant identity mismatch")
		}
		p.manifest = old
		p.manifest.PreviousSHA256 = digest(current)
		p.manifest.ProjectionSHA256 = digest(p.projection)
		p.manifest.SourceSHA256 = digest(p.projection)
		p.manifest.SourcePath = runtime
		p.Changes = []Change{{"store previous", pp, digest(current)}, {"replace", projection, digest(p.projection)}}
	case ActionRollback:
		if s.State != StateHealthy && s.State != StateSourceDrift {
			return p, fmt.Errorf("rollback refused: %s", s.State)
		}
		old, e := loadManifest(mp)
		if e != nil {
			return p, e
		}
		if old.PreviousSHA256 == "" {
			return p, errors.New("no rollback generation")
		}
		p.projection, e = safeRegular(pp)
		if e != nil || digest(p.projection) != old.PreviousSHA256 {
			return p, errors.New("rollback generation is not manifest-verified")
		}
		p.manifest = old
		p.manifest.ProjectionSHA256 = old.PreviousSHA256
		p.manifest.SourceSHA256 = old.PreviousSHA256
		p.manifest.PreviousSHA256 = ""
		p.Changes = []Change{{"restore", projection, digest(p.projection)}, {"remove", pp, ""}}
	case ActionUninstall:
		if s.State != StateHealthy && s.State != StateSourceDrift {
			return p, fmt.Errorf("uninstall refused: %s", s.State)
		}
		p.manifest, err = loadManifest(mp)
		if err != nil {
			return p, err
		}
		p.Changes = []Change{{"remove", projection, p.manifest.ProjectionSHA256}, {"remove", filepath.Join(h, controlDir), ""}}
	default:
		return p, fmt.Errorf("unknown action %q", action)
	}
	return p, nil
}

func atomicWrite(path string, b []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".my-friday-write-")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer os.Remove(name)
	if err = tmp.Chmod(0600); err == nil {
		_, err = tmp.Write(b)
	}
	if err == nil {
		err = tmp.Sync()
	}
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	return os.Rename(name, path)
}
func writeManifest(path string, m manifest) error {
	b, _ := json.MarshalIndent(m, "", "  ")
	return atomicWrite(path, append(b, '\n'))
}

func writeExclusive(path string, b []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return fmt.Errorf("another transaction or recovery is active: %w", err)
	}
	if _, err = f.Write(b); err == nil {
		err = f.Sync()
	}
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	return err
}

func Execute(p PlanResult) error {
	// Rebuild the plan immediately before mutation to close the preview/use gap.
	fresh, err := Plan(p.Action, p.Runtime, p.CodexHome)
	if err != nil {
		return err
	}
	projection, mp, pp, jp := paths(p.CodexHome)
	j, _ := json.Marshal(journal{1, p.Action, "prepared"})
	if err = writeExclusive(jp, j); err != nil {
		return err
	}
	recoveryProjection := filepath.Join(p.CodexHome, controlDir, recoveryProjectionFile)
	recoveryManifest := filepath.Join(p.CodexHome, controlDir, recoveryManifestFile)
	if p.Action != ActionInstall {
		current, copyErr := safeRegular(projection)
		if copyErr == nil {
			copyErr = atomicWrite(recoveryProjection, current)
		}
		manifestBytes, manifestErr := safeRegular(mp)
		if copyErr == nil && manifestErr == nil {
			copyErr = atomicWrite(recoveryManifest, manifestBytes)
		}
		if copyErr != nil || manifestErr != nil {
			return fmt.Errorf("transaction preparation failed; recovery required at %s", jp)
		}
	}
	complete := false
	defer func() {
		if complete {
			_ = os.Remove(jp)
		}
	}()
	switch p.Action {
	case ActionInstall:
		if err = atomicWrite(projection, fresh.projection); err == nil {
			err = writeManifest(mp, fresh.manifest)
		}
	case ActionRepair:
		if err = atomicWrite(projection, fresh.projection); err == nil {
			err = writeManifest(mp, fresh.manifest)
		}
	case ActionUpgrade:
		var current []byte
		current, err = safeRegular(projection)
		if err == nil {
			err = atomicWrite(pp, current)
		}
		if err == nil {
			err = atomicWrite(projection, fresh.projection)
		}
		if err == nil {
			err = writeManifest(mp, fresh.manifest)
		}
	case ActionRollback:
		if err = atomicWrite(projection, fresh.projection); err == nil {
			err = writeManifest(mp, fresh.manifest)
		}
		if err == nil {
			err = os.Remove(pp)
		}
	case ActionUninstall:
		if err = os.Remove(projection); err == nil {
			err = os.RemoveAll(filepath.Join(p.CodexHome, controlDir))
		}
	}
	if err != nil {
		return fmt.Errorf("transaction interrupted; recovery required at %s: %w", jp, err)
	}
	if err = os.Remove(jp); err != nil && !os.IsNotExist(err) {
		return err
	}
	_ = os.Remove(recoveryProjection)
	_ = os.Remove(recoveryManifest)
	complete = true
	return nil
}

// Recover is idempotent. Prepared journals precede mutation and can be removed;
// a later ambiguous state is retained for explicit diagnosis.
func Recover(home string) error {
	h, err := canonicalHome(home)
	if err != nil {
		return err
	}
	_, _, _, jp := paths(h)
	b, err := safeRegular(jp)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	var j journal
	if json.Unmarshal(b, &j) != nil || j.ContractVersion != 1 || j.Phase != "prepared" {
		return errors.New("recovery refused: incompatible or ambiguous journal")
	}
	projection, mp, _, _ := paths(h)
	if j.Action == ActionInstall {
		if m, manifestErr := loadManifest(mp); manifestErr == nil {
			if installed, fileErr := safeRegular(projection); fileErr == nil && digest(installed) == m.ProjectionSHA256 {
				return os.Remove(jp)
			}
		}
		if err := os.Remove(projection); err != nil && !os.IsNotExist(err) {
			return err
		}
		return os.RemoveAll(filepath.Join(h, controlDir))
	}
	if m, manifestErr := loadManifest(mp); manifestErr == nil {
		if installed, fileErr := safeRegular(projection); fileErr == nil && digest(installed) == m.ProjectionSHA256 {
			_ = os.Remove(filepath.Join(h, controlDir, recoveryProjectionFile))
			_ = os.Remove(filepath.Join(h, controlDir, recoveryManifestFile))
			return os.Remove(jp)
		}
	}
	priorProjection, projectionErr := safeRegular(filepath.Join(h, controlDir, recoveryProjectionFile))
	priorManifest, manifestErr := safeRegular(filepath.Join(h, controlDir, recoveryManifestFile))
	if projectionErr != nil || manifestErr != nil {
		return errors.New("recovery refused: exact pre-change generation is unavailable")
	}
	var prior manifest
	if json.Unmarshal(priorManifest, &prior) != nil || digest(priorProjection) != prior.ProjectionSHA256 {
		return errors.New("recovery refused: pre-change generation proof failed")
	}
	if err := atomicWrite(projection, priorProjection); err != nil {
		return err
	}
	if err := atomicWrite(mp, priorManifest); err != nil {
		return err
	}
	_ = os.Remove(filepath.Join(h, controlDir, recoveryProjectionFile))
	_ = os.Remove(filepath.Join(h, controlDir, recoveryManifestFile))
	return os.Remove(jp)
}
