// Package codexhome manages the deliberately tiny installed Codex baseline.
package codexhome

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/acoz-labs/my-friday/internal/profile"
	"github.com/acoz-labs/my-friday/internal/repository"
	"golang.org/x/sys/unix"
)

const (
	controlDir    = ".my-friday"
	manifestFile  = "installed-baseline.json"
	previousFile  = "previous-AGENTS.md"
	canonicalFile = "canonical-AGENTS.md"
	journalFile   = "transaction.json"
)

var managedControlNames = map[string]bool{manifestFile: true, previousFile: true, canonicalFile: true, journalFile: true}

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
type fileProof struct {
	Exists bool   `json:"exists"`
	SHA256 string `json:"sha256,omitempty"`
	Bytes  []byte `json:"bytes,omitempty"`
}
type rootProof struct {
	Device uint64 `json:"device"`
	Inode  uint64 `json:"inode"`
}
type PlanResult struct {
	Action               Action
	Runtime, CodexHome   string
	Changes              []Change
	NegativeActions      []string
	projection, previous []byte
	manifest             manifest
	root                 rootProof
	before               map[string]fileProof
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
	CanonicalSHA256  string `json:"canonical_sha256"`
	PreviousSHA256   string `json:"previous_sha256,omitempty"`
}
type runtimeManifest struct {
	ContractVersion int    `json:"contract_version"`
	RepositoryRole  string `json:"repository_role"`
	AssistantID     string `json:"assistant_id"`
}
type journal struct {
	ContractVersion int                  `json:"contract_version"`
	Action          Action               `json:"action"`
	Phase           string               `json:"phase"`
	Root            rootProof            `json:"root"`
	Before          map[string]fileProof `json:"before"`
	After           map[string]fileProof `json:"after"`
}

// faultHook is test-only deterministic crash injection after durable phases.
var faultHook func(string) error

func digest(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func validDigest(s string) bool {
	_, e := hex.DecodeString(s)
	return len(s) == 64 && e == nil && s == strings.ToLower(s)
}
func paths(home string) (string, string, string, string) {
	return filepath.Join(home, "AGENTS.md"), filepath.Join(home, controlDir, manifestFile), filepath.Join(home, controlDir, previousFile), filepath.Join(home, controlDir, journalFile)
}

type homeRoot struct {
	fd    int
	path  string
	proof rootProof
}

func (r *homeRoot) close() {
	if r != nil && r.fd >= 0 {
		_ = unix.Close(r.fd)
		r.fd = -1
	}
}
func openHome(home string) (*homeRoot, error) {
	if home == "" {
		return nil, errors.New("Codex home is required")
	}
	abs, e := filepath.Abs(home)
	if e != nil {
		return nil, e
	}
	abs = filepath.Clean(abs)
	// macOS exposes /var as a system-owned compatibility symlink to
	// /private/var. Resolve only that fixed platform alias; user-controlled
	// ancestor symlinks remain forbidden by the component walk below.
	if abs == "/var" || strings.HasPrefix(abs, "/var/") {
		abs = filepath.Join("/private/var", strings.TrimPrefix(abs, "/var/"))
	}
	fd, e := unix.Open("/", unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if e != nil {
		return nil, e
	}
	for _, part := range strings.Split(strings.TrimPrefix(abs, "/"), "/") {
		if part == "" {
			continue
		}
		next, x := unix.Openat(fd, part, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
		_ = unix.Close(fd)
		if x != nil {
			return nil, fmt.Errorf("Codex home path component %q is unsafe: %w", part, x)
		}
		fd = next
	}
	var st unix.Stat_t
	if e = unix.Fstat(fd, &st); e != nil {
		_ = unix.Close(fd)
		return nil, e
	}
	if st.Uid != uint32(os.Getuid()) {
		_ = unix.Close(fd)
		return nil, errors.New("Codex home must be owned by current user")
	}
	return &homeRoot{fd, abs, rootProof{uint64(st.Dev), uint64(st.Ino)}}, nil
}
func canonicalHome(home string) (string, error) {
	r, e := openHome(home)
	if e != nil {
		return "", e
	}
	defer r.close()
	return r.path, nil
}
func openControl(r *homeRoot, create bool) (int, error) {
	fd, e := unix.Openat(r.fd, controlDir, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if e == nil {
		var st unix.Stat_t
		if x := unix.Fstat(fd, &st); x != nil || st.Uid != uint32(os.Getuid()) || st.Mode&0777 != 0700 {
			_ = unix.Close(fd)
			return -1, errors.New("unsafe control directory owner or mode")
		}
		return fd, nil
	}
	if !create {
		return fd, e
	}
	if !errors.Is(e, unix.ENOENT) {
		return -1, e
	}
	if e = unix.Mkdirat(r.fd, controlDir, 0700); e != nil && !errors.Is(e, unix.EEXIST) {
		return -1, e
	}
	return openControl(r, false)
}
func readRegularAt(dirfd int, name string) ([]byte, error) {
	fd, e := unix.Openat(dirfd, name, unix.O_RDONLY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if e != nil {
		return nil, e
	}
	f := os.NewFile(uintptr(fd), name)
	defer f.Close()
	var st unix.Stat_t
	if e = unix.Fstat(fd, &st); e != nil {
		return nil, e
	}
	if st.Mode&unix.S_IFMT != unix.S_IFREG || st.Nlink != 1 || st.Uid != uint32(os.Getuid()) || st.Mode&0777 != 0600 {
		return nil, fmt.Errorf("unsafe managed file: %s", name)
	}
	return io.ReadAll(f)
}
func safeRegular(path string) ([]byte, error) {
	d, e := os.Open(filepath.Dir(path))
	if e != nil {
		return nil, e
	}
	defer d.Close()
	return readRegularAt(int(d.Fd()), filepath.Base(path))
}
func lstatAt(fd int, name string) (unix.Stat_t, error) {
	var st unix.Stat_t
	e := unix.Fstatat(fd, name, &st, unix.AT_SYMLINK_NOFOLLOW)
	return st, e
}
func proofAt(fd int, name string) (fileProof, error) {
	b, e := readRegularAt(fd, name)
	if errors.Is(e, unix.ENOENT) {
		return fileProof{}, nil
	}
	if e != nil {
		return fileProof{}, e
	}
	return fileProof{true, digest(b), b}, nil
}
func controlNames(fd int) ([]string, error) {
	dup, e := unix.Dup(fd)
	if e != nil {
		return nil, e
	}
	f := os.NewFile(uintptr(dup), controlDir)
	defer f.Close()
	names, e := f.Readdirnames(-1)
	if e != nil {
		return nil, e
	}
	sort.Strings(names)
	for _, n := range names {
		if !managedControlNames[n] {
			return nil, fmt.Errorf("foreign control-tree entry %q", n)
		}
	}
	return names, nil
}
func validateControl(r *homeRoot) error {
	fd, e := openControl(r, false)
	if errors.Is(e, unix.ENOENT) {
		return nil
	}
	if e != nil {
		return fmt.Errorf("unsafe control tree: %w", e)
	}
	defer unix.Close(fd)
	_, e = controlNames(fd)
	return e
}
func strictJSON(b []byte, v any) error {
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if e := d.Decode(v); e != nil {
		return e
	}
	if d.Decode(&struct{}{}) != io.EOF {
		return errors.New("trailing JSON data")
	}
	return nil
}

func render(runtime string) ([]byte, runtimeManifest, string, error) {
	abs, e := filepath.Abs(runtime)
	if e != nil {
		return nil, runtimeManifest{}, "", e
	}
	abs, e = filepath.EvalSymlinks(filepath.Clean(abs))
	if e != nil {
		return nil, runtimeManifest{}, "", e
	}
	abs = filepath.Clean(abs)
	id, e := repository.ValidateRuntime(abs)
	if e != nil {
		return nil, runtimeManifest{}, "", fmt.Errorf("runtime repository: %w", e)
	}
	mb, e := safeRegular(filepath.Join(abs, ".my-friday", "manifest.json"))
	if e != nil {
		return nil, runtimeManifest{}, "", fmt.Errorf("runtime manifest: %w", e)
	}
	var rm runtimeManifest
	if json.Unmarshal(mb, &rm) != nil || rm.ContractVersion != 1 || rm.RepositoryRole != "runtime" || !strings.HasPrefix(rm.AssistantID, "asst-") {
		return nil, rm, "", errors.New("runtime repository manifest is incompatible")
	}
	if rm.AssistantID != id {
		return nil, rm, "", errors.New("runtime assistant identity mismatch")
	}
	pb, e := safeRegular(filepath.Join(abs, "assistant", "profile.json"))
	if e != nil {
		return nil, rm, "", e
	}
	var p profile.Profile
	if json.Unmarshal(pb, &p) != nil || profile.Validate(p) != nil || p.AssistantID != id {
		return nil, rm, "", errors.New("runtime profile is incompatible")
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
	return []byte(out.String()), rm, abs, nil
}
func validateManifest(m manifest) error {
	if m.ContractVersion != 1 || !strings.HasPrefix(m.AssistantID, "asst-") || len(m.AssistantID) <= 5 {
		return errors.New("identity invariant")
	}
	if !validDigest(m.ProjectionSHA256) || !validDigest(m.SourceSHA256) || !validDigest(m.CanonicalSHA256) {
		return errors.New("digest invariant")
	}
	if m.ProjectionSHA256 != m.SourceSHA256 || m.ProjectionSHA256 != m.CanonicalSHA256 {
		return errors.New("active invariant")
	}
	if m.PreviousSHA256 != "" && (!validDigest(m.PreviousSHA256) || m.PreviousSHA256 == m.ProjectionSHA256) {
		return errors.New("previous invariant")
	}
	if !filepath.IsAbs(m.SourcePath) || filepath.Clean(m.SourcePath) != m.SourcePath {
		return errors.New("source path invariant")
	}
	return nil
}
func loadManifestAt(fd int) (manifest, error) {
	b, e := readRegularAt(fd, manifestFile)
	if e != nil {
		return manifest{}, e
	}
	var m manifest
	if strictJSON(b, &m) != nil || validateManifest(m) != nil {
		return manifest{}, errors.New("installed manifest is incompatible")
	}
	return m, nil
}
func loadManifest(path string) (manifest, error) {
	b, e := safeRegular(path)
	if e != nil {
		return manifest{}, e
	}
	var m manifest
	if strictJSON(b, &m) != nil || validateManifest(m) != nil {
		return manifest{}, errors.New("installed manifest is incompatible")
	}
	return m, nil
}

func inspectRoot(r *homeRoot, runtime string, checkSource bool) (Status, manifest, error) {
	if e := validateControl(r); e != nil {
		return Status{StateCollision, e.Error()}, manifest{}, nil
	}
	if _, e := lstatAt(r.fd, "AGENTS.override.md"); e == nil {
		return Status{StateCollision, "AGENTS.override.md shadows the managed projection"}, manifest{}, nil
	} else if !errors.Is(e, unix.ENOENT) {
		return Status{}, manifest{}, e
	}
	cfd, ce := openControl(r, false)
	_, pe := lstatAt(r.fd, "AGENTS.md")
	if ce == nil {
		defer unix.Close(cfd)
		if _, je := lstatAt(cfd, journalFile); je == nil {
			return Status{StateInterrupted, "transaction journal exists"}, manifest{}, nil
		}
	}
	if errors.Is(pe, unix.ENOENT) && errors.Is(ce, unix.ENOENT) {
		return Status{StateNotInstalled, "no managed projection"}, manifest{}, nil
	}
	if pe != nil || ce != nil {
		return Status{StateCollision, "projection and ownership control tree disagree"}, manifest{}, nil
	}
	m, e := loadManifestAt(cfd)
	if e != nil {
		return Status{StateCollision, e.Error()}, manifest{}, nil
	}
	projection, e := readRegularAt(r.fd, "AGENTS.md")
	if e != nil {
		return Status{StateCollision, e.Error()}, manifest{}, nil
	}
	canonical, e := readRegularAt(cfd, canonicalFile)
	if e != nil || digest(canonical) != m.CanonicalSHA256 {
		return Status{StateCollision, "canonical generation disagrees with manifest"}, manifest{}, nil
	}
	prev, e := proofAt(cfd, previousFile)
	if e != nil || prev.Exists != (m.PreviousSHA256 != "") || (prev.Exists && prev.SHA256 != m.PreviousSHA256) {
		return Status{StateCollision, "previous generation disagrees with manifest"}, manifest{}, nil
	}
	if digest(projection) != m.ProjectionSHA256 {
		return Status{StateDrift, "projection digest differs from manifest"}, m, nil
	}
	if checkSource {
		if runtime == "" {
			runtime = m.SourcePath
		}
		rb, rm, source, re := render(runtime)
		if re != nil || rm.AssistantID != m.AssistantID || source != m.SourcePath || digest(rb) != m.SourceSHA256 {
			return Status{StateSourceDrift, "installed state is healthy but source differs"}, m, nil
		}
	}
	return Status{StateHealthy, "manifest, canonical generation, and projection agree"}, m, nil
}
func Inspect(runtime, home string) (Status, error) {
	r, e := openHome(home)
	if e != nil {
		return Status{}, e
	}
	defer r.close()
	s, _, e := inspectRoot(r, runtime, true)
	return s, e
}
func snapshot(r *homeRoot) (map[string]fileProof, error) {
	out := map[string]fileProof{}
	p, e := proofAt(r.fd, "AGENTS.md")
	if e != nil {
		return nil, e
	}
	out["projection"] = p
	cfd, e := openControl(r, false)
	if errors.Is(e, unix.ENOENT) {
		for _, k := range []string{"manifest", "canonical", "previous"} {
			out[k] = fileProof{}
		}
		return out, nil
	}
	if e != nil {
		return nil, e
	}
	defer unix.Close(cfd)
	for k, n := range map[string]string{"manifest": manifestFile, "canonical": canonicalFile, "previous": previousFile} {
		p, e = proofAt(cfd, n)
		if e != nil {
			return nil, e
		}
		out[k] = p
	}
	return out, nil
}
func proofsEqual(a, b map[string]fileProof) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		w, ok := b[k]
		if !ok || v.Exists != w.Exists || v.SHA256 != w.SHA256 || !bytes.Equal(v.Bytes, w.Bytes) {
			return false
		}
	}
	return true
}
func planEqual(a, b PlanResult) bool {
	return a.Action == b.Action && a.Runtime == b.Runtime && a.CodexHome == b.CodexHome && a.root == b.root && bytes.Equal(a.projection, b.projection) && bytes.Equal(a.previous, b.previous) && a.manifest == b.manifest && proofsEqual(a.before, b.before) && fmt.Sprint(a.Changes) == fmt.Sprint(b.Changes)
}

func Plan(action Action, runtime, home string) (PlanResult, error) {
	r, e := openHome(home)
	if e != nil {
		return PlanResult{}, e
	}
	defer r.close()
	check := action == ActionInstall || action == ActionRepair || action == ActionUpgrade
	s, m, e := inspectRoot(r, runtime, check)
	if e != nil {
		return PlanResult{}, e
	}
	p := PlanResult{Action: action, Runtime: runtime, CodexHome: r.path, root: r.proof, NegativeActions: []string{"edit config.toml, auth, sessions, logs, skills, packages, or project configuration", "start a daemon, use the network, or request privilege"}}
	p.before, e = snapshot(r)
	if e != nil {
		return p, e
	}
	projection, mp, pp, _ := paths(r.path)
	switch action {
	case ActionInstall:
		if s.State != StateNotInstalled {
			return p, fmt.Errorf("install refused: %s", s.State)
		}
		var rm runtimeManifest
		p.projection, rm, p.Runtime, e = render(runtime)
		if e != nil {
			return p, e
		}
		sha := digest(p.projection)
		p.manifest = manifest{1, rm.AssistantID, sha, p.Runtime, sha, sha, ""}
		p.Changes = []Change{{"create", projection, sha}, {"create", mp, "manifest"}, {"create", filepath.Join(r.path, controlDir, canonicalFile), sha}}
	case ActionRepair:
		if s.State != StateDrift {
			return p, fmt.Errorf("repair requires managed drift; found %s", s.State)
		}
		cfd, x := openControl(r, false)
		if x != nil {
			return p, x
		}
		p.projection, x = readRegularAt(cfd, canonicalFile)
		unix.Close(cfd)
		if x != nil {
			return p, x
		}
		rb, rm, source, x := render(runtime)
		if x != nil || rm.AssistantID != m.AssistantID || source != m.SourcePath || digest(rb) != m.SourceSHA256 {
			return p, errors.New("repair refused: recorded source changed")
		}
		p.Runtime = source
		p.manifest = m
		p.Changes = []Change{{"restore manifest-proven bytes", projection, m.ProjectionSHA256}}
	case ActionUpgrade:
		if s.State != StateSourceDrift {
			return p, fmt.Errorf("upgrade requires healthy installed state and changed source; found %s", s.State)
		}
		current, x := readRegularAt(r.fd, "AGENTS.md")
		if x != nil {
			return p, x
		}
		var rm runtimeManifest
		p.projection, rm, p.Runtime, x = render(runtime)
		if x != nil {
			return p, x
		}
		if rm.AssistantID != m.AssistantID {
			return p, errors.New("assistant identity mismatch")
		}
		p.previous = current
		sha := digest(p.projection)
		p.manifest = m
		p.manifest.PreviousSHA256 = digest(current)
		p.manifest.ProjectionSHA256 = sha
		p.manifest.SourceSHA256 = sha
		p.manifest.CanonicalSHA256 = sha
		p.manifest.SourcePath = p.Runtime
		p.Changes = []Change{{"store previous", pp, digest(current)}, {"replace", projection, sha}, {"replace", filepath.Join(r.path, controlDir, canonicalFile), sha}}
	case ActionRollback:
		if s.State != StateHealthy {
			return p, fmt.Errorf("rollback refused: %s", s.State)
		}
		if m.PreviousSHA256 == "" {
			return p, errors.New("no rollback generation")
		}
		cfd, x := openControl(r, false)
		if x != nil {
			return p, x
		}
		p.projection, x = readRegularAt(cfd, previousFile)
		unix.Close(cfd)
		if x != nil || digest(p.projection) != m.PreviousSHA256 {
			return p, errors.New("rollback generation is not manifest-verified")
		}
		p.manifest = m
		p.manifest.ProjectionSHA256 = m.PreviousSHA256
		p.manifest.SourceSHA256 = m.PreviousSHA256
		p.manifest.CanonicalSHA256 = m.PreviousSHA256
		p.manifest.PreviousSHA256 = ""
		p.Changes = []Change{{"restore", projection, digest(p.projection)}, {"replace", filepath.Join(r.path, controlDir, canonicalFile), digest(p.projection)}, {"remove", pp, ""}}
	case ActionUninstall:
		if s.State != StateHealthy {
			return p, fmt.Errorf("uninstall refused: %s", s.State)
		}
		p.manifest = m
		p.Changes = []Change{{"remove", projection, m.ProjectionSHA256}, {"remove proven managed entries", filepath.Join(r.path, controlDir), ""}}
	default:
		return p, fmt.Errorf("unknown action %q", action)
	}
	return p, nil
}

func atomicWriteAt(fd int, name string, b []byte) error {
	random := make([]byte, 12)
	if _, e := rand.Read(random); e != nil {
		return e
	}
	tmp := ".my-friday-write-" + hex.EncodeToString(random)
	fno, e := unix.Openat(fd, tmp, unix.O_WRONLY|unix.O_CREAT|unix.O_EXCL|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0600)
	if e != nil {
		return e
	}
	f := os.NewFile(uintptr(fno), tmp)
	_, e = f.Write(b)
	if e == nil {
		e = f.Sync()
	}
	if x := f.Close(); e == nil {
		e = x
	}
	if e != nil {
		_ = unix.Unlinkat(fd, tmp, 0)
		return e
	}
	if e = unix.Renameat(fd, tmp, fd, name); e != nil {
		_ = unix.Unlinkat(fd, tmp, 0)
		return e
	}
	return unix.Fsync(fd)
}
func writeExclusiveAt(fd int, name string, b []byte) error {
	n, e := unix.Openat(fd, name, unix.O_WRONLY|unix.O_CREAT|unix.O_EXCL|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0600)
	if e != nil {
		return fmt.Errorf("another transaction or recovery is active: %w", e)
	}
	f := os.NewFile(uintptr(n), name)
	_, e = f.Write(b)
	if e == nil {
		e = f.Sync()
	}
	if x := f.Close(); e == nil {
		e = x
	}
	if e == nil {
		e = unix.Fsync(fd)
	}
	return e
}
func manifestBytes(m manifest) []byte {
	b, _ := json.MarshalIndent(m, "", "  ")
	return append(b, '\n')
}
func journalBytes(j journal) []byte { b, _ := json.MarshalIndent(j, "", "  "); return append(b, '\n') }
func afterProofs(p PlanResult) map[string]fileProof {
	mk := func(b []byte) fileProof { return fileProof{true, digest(b), b} }
	if p.Action == ActionUninstall {
		return map[string]fileProof{"projection": {}, "manifest": {}, "canonical": {}, "previous": {}}
	}
	prev := fileProof{}
	if p.manifest.PreviousSHA256 != "" {
		prev = mk(p.previous)
	}
	return map[string]fileProof{"projection": mk(p.projection), "manifest": mk(manifestBytes(p.manifest)), "canonical": mk(p.projection), "previous": prev}
}
func setPhase(fd int, j *journal, phase string) error {
	j.Phase = phase
	if e := atomicWriteAt(fd, journalFile, journalBytes(*j)); e != nil {
		return e
	}
	if faultHook != nil {
		return faultHook(phase)
	}
	return nil
}
func applyPhase(cfd int, j *journal, phase string, apply func() error) error {
	if e := apply(); e != nil {
		return e
	}
	return setPhase(cfd, j, phase)
}
func removeAt(fd int, name string) error {
	e := unix.Unlinkat(fd, name, 0)
	if errors.Is(e, unix.ENOENT) {
		return nil
	}
	return e
}
func applyProof(fd int, name string, p fileProof) error {
	if !p.Exists {
		return removeAt(fd, name)
	}
	if digest(p.Bytes) != p.SHA256 {
		return errors.New("journal byte proof failed")
	}
	return atomicWriteAt(fd, name, p.Bytes)
}
func proofMatches(fd int, name string, want fileProof) bool {
	got, e := proofAt(fd, name)
	return e == nil && got.Exists == want.Exists && got.SHA256 == want.SHA256 && bytes.Equal(got.Bytes, want.Bytes)
}
func applyTransition(fd int, name string, before, after fileProof) error {
	if !proofMatches(fd, name, before) {
		return fmt.Errorf("managed state changed before mutation: %s", name)
	}
	return applyProof(fd, name, after)
}
func applyRecovery(fd int, name string, before, after, target fileProof) error {
	if !proofMatches(fd, name, before) && !proofMatches(fd, name, after) {
		return fmt.Errorf("recovery refused: %s matches neither journal generation", name)
	}
	return applyProof(fd, name, target)
}
func Execute(p PlanResult) error {
	fresh, e := Plan(p.Action, p.Runtime, p.CodexHome)
	if e != nil {
		return e
	}
	if !planEqual(p, fresh) {
		return errors.New("execution refused: current state no longer matches confirmed preview")
	}
	r, e := openHome(p.CodexHome)
	if e != nil {
		return e
	}
	defer r.close()
	if r.proof != p.root {
		return errors.New("execution refused: Codex home identity changed")
	}
	current, e := snapshot(r)
	if e != nil || !proofsEqual(current, p.before) {
		return errors.New("execution refused: managed state changed after preview")
	}
	cfd, e := openControl(r, true)
	if e != nil {
		return e
	}
	defer unix.Close(cfd)
	if _, e = controlNames(cfd); e != nil {
		return e
	}
	j := journal{1, p.Action, "prepared", p.root, p.before, afterProofs(p)}
	if e = writeExclusiveAt(cfd, journalFile, journalBytes(j)); e != nil {
		return e
	}
	if faultHook != nil {
		if e = faultHook("prepared"); e != nil {
			return fmt.Errorf("transaction interrupted; recovery required: %w", e)
		}
	}
	if e = setPhase(cfd, &j, "mutating"); e != nil {
		return fmt.Errorf("transaction interrupted; recovery required: %w", e)
	}
	if e = applyPhase(cfd, &j, "projection-written", func() error {
		return applyTransition(r.fd, "AGENTS.md", j.Before["projection"], j.After["projection"])
	}); e == nil {
		e = applyPhase(cfd, &j, "canonical-written", func() error {
			return applyTransition(cfd, canonicalFile, j.Before["canonical"], j.After["canonical"])
		})
	}
	if e == nil {
		e = applyPhase(cfd, &j, "previous-written", func() error {
			return applyTransition(cfd, previousFile, j.Before["previous"], j.After["previous"])
		})
	}
	if e == nil {
		e = applyPhase(cfd, &j, "manifest-written", func() error {
			return applyTransition(cfd, manifestFile, j.Before["manifest"], j.After["manifest"])
		})
	}
	if e != nil {
		return fmt.Errorf("transaction interrupted; recovery required: %w", e)
	}
	if e = setPhase(cfd, &j, "committed"); e != nil {
		return fmt.Errorf("transaction interrupted; recovery required: %w", e)
	}
	if e = removeAt(cfd, journalFile); e != nil {
		return e
	}
	if e = unix.Fsync(cfd); e != nil {
		return e
	}
	if p.Action == ActionUninstall {
		names, x := controlNames(cfd)
		if x != nil {
			return x
		}
		if len(names) != 0 {
			return errors.New("uninstall refused: control tree is not empty")
		}
		if x = unix.Unlinkat(r.fd, controlDir, unix.AT_REMOVEDIR); x != nil {
			return x
		}
		return unix.Fsync(r.fd)
	}
	return nil
}
func validJournal(j journal) error {
	validPhase := map[string]bool{
		"prepared": true, "mutating": true, "projection-written": true,
		"canonical-written": true, "previous-written": true,
		"manifest-written": true, "committed": true,
	}
	if j.ContractVersion != 1 || !validPhase[j.Phase] {
		return errors.New("incompatible phase")
	}
	switch j.Action {
	case ActionInstall, ActionRepair, ActionUpgrade, ActionRollback, ActionUninstall:
	default:
		return errors.New("incompatible action")
	}
	for _, set := range []map[string]fileProof{j.Before, j.After} {
		if len(set) != 4 {
			return errors.New("incomplete generation set")
		}
		for _, k := range []string{"projection", "manifest", "canonical", "previous"} {
			p, ok := set[k]
			if !ok {
				return errors.New("missing generation proof")
			}
			if p.Exists {
				if !validDigest(p.SHA256) || digest(p.Bytes) != p.SHA256 {
					return errors.New("invalid generation proof")
				}
			} else if p.SHA256 != "" || len(p.Bytes) != 0 {
				return errors.New("invalid absent proof")
			}
		}
	}
	return nil
}
func Recover(home string) error {
	r, e := openHome(home)
	if e != nil {
		return e
	}
	defer r.close()
	if e = validateControl(r); e != nil {
		return e
	}
	cfd, e := openControl(r, false)
	if errors.Is(e, unix.ENOENT) {
		return nil
	}
	if e != nil {
		return e
	}
	defer unix.Close(cfd)
	b, e := readRegularAt(cfd, journalFile)
	if errors.Is(e, unix.ENOENT) {
		return nil
	}
	if e != nil {
		return e
	}
	var j journal
	if strictJSON(b, &j) != nil || validJournal(j) != nil {
		return errors.New("recovery refused: incompatible or ambiguous journal")
	}
	if j.Root != r.proof {
		return errors.New("recovery refused: Codex home identity changed")
	}
	target := j.Before
	if j.Phase == "committed" {
		target = j.After
	}
	if e = applyRecovery(r.fd, "AGENTS.md", j.Before["projection"], j.After["projection"], target["projection"]); e == nil {
		e = applyRecovery(cfd, canonicalFile, j.Before["canonical"], j.After["canonical"], target["canonical"])
	}
	if e == nil {
		e = applyRecovery(cfd, previousFile, j.Before["previous"], j.After["previous"], target["previous"])
	}
	if e == nil {
		e = applyRecovery(cfd, manifestFile, j.Before["manifest"], j.After["manifest"], target["manifest"])
	}
	if e != nil {
		return e
	}
	if e = removeAt(cfd, journalFile); e != nil {
		return e
	}
	if e = unix.Fsync(cfd); e != nil {
		return e
	}
	names, e := controlNames(cfd)
	if e != nil {
		return e
	}
	if len(names) == 0 {
		if e = unix.Unlinkat(r.fd, controlDir, unix.AT_REMOVEDIR); e != nil {
			return e
		}
		return unix.Fsync(r.fd)
	}
	return nil
}
