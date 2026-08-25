// Package capability implements the strict, source-first instruction-only
// capability contract and its instance-scoped projection lifecycle.
package capability

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"syscall"
	"unicode/utf8"
)

const (
	MaxFiles        = 64
	MaxFileBytes    = 256 * 1024
	MaxPackageBytes = 1024 * 1024
	MaxDepth        = 8
)

var slugPattern = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$`)
var versionPattern = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)

type Manifest struct {
	ContractVersion    int      `json:"contract_version"`
	Slug               string   `json:"slug"`
	Version            string   `json:"version"`
	DisplayName        string   `json:"display_name"`
	Summary            string   `json:"summary"`
	Profile            string   `json:"profile"`
	CodexCompatibility string   `json:"codex_compatibility"`
	Triggers           []string `json:"triggers"`
	Inputs             []string `json:"inputs"`
	Outputs            []string `json:"outputs"`
	SuccessBehavior    string   `json:"success_behavior"`
	FailureBehavior    string   `json:"failure_behavior"`
	Scripts            string   `json:"scripts"`
	Dependencies       string   `json:"dependencies"`
	Network            string   `json:"network"`
	Credentials        string   `json:"credentials"`
	Background         string   `json:"background"`
	DurableData        string   `json:"durable_data"`
	Publishing         string   `json:"publishing"`
}

type Cases struct {
	ContractVersion  int      `json:"contract_version"`
	PositiveTriggers []string `json:"positive_triggers"`
	NonTriggers      []string `json:"non_triggers"`
	Examples         []struct {
		Input          string   `json:"input"`
		OutputContains []string `json:"output_contains"`
	} `json:"examples"`
	RequiredFacts    []string `json:"required_facts"`
	ForbiddenEffects []string `json:"forbidden_effects"`
}

type Package struct {
	Root             string
	Manifest         Manifest
	Cases            Cases
	Files            []string
	SourceDigest     string
	ProjectionDigest string
	Projection       map[string][]byte
}

func strictJSON(body []byte, out any) error {
	d := json.NewDecoder(bytes.NewReader(body))
	d.DisallowUnknownFields()
	if err := d.Decode(out); err != nil {
		return err
	}
	if d.More() {
		return errors.New("multiple JSON values")
	}
	return nil
}

func Validate(root string) (Package, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return Package{}, err
	}
	info, err := os.Lstat(abs)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return Package{}, fmt.Errorf("invalid capability root")
	}
	var files []string
	total := int64(0)
	bodies := map[string][]byte{}
	err = filepath.WalkDir(abs, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(abs, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if rel == "." {
			return nil
		}
		if !utf8.ValidString(rel) || strings.Contains(rel, "\\") || len(strings.Split(rel, "/")) > MaxDepth {
			return fmt.Errorf("invalid capability path %q", rel)
		}
		i, err := os.Lstat(path)
		if err != nil {
			return err
		}
		st, ok := i.Sys().(*syscall.Stat_t)
		if !ok || i.Mode()&os.ModeSymlink != 0 || (!i.IsDir() && !i.Mode().IsRegular()) {
			return fmt.Errorf("unsafe capability entry %s", rel)
		}
		if i.IsDir() {
			if rel == ".git" || strings.HasPrefix(rel, ".git/") || rel == "skill/scripts" || rel == "skill/agents" || !allowedDir(rel) {
				return fmt.Errorf("unsupported capability directory %s", rel)
			}
			return nil
		}
		if st.Nlink != 1 || i.Mode().Perm()&0o111 != 0 {
			return fmt.Errorf("unsafe capability file %s", rel)
		}
		if !allowedFile(rel) {
			return fmt.Errorf("unsupported capability file %s", rel)
		}
		if i.Size() > MaxFileBytes {
			return fmt.Errorf("capability file too large %s", rel)
		}
		total += i.Size()
		if total > MaxPackageBytes || len(files) >= MaxFiles {
			return errors.New("capability package bounds exceeded")
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		files = append(files, rel)
		bodies[rel] = body
		return nil
	})
	if err != nil {
		return Package{}, err
	}
	for _, required := range []string{"capability.json", "skill/SKILL.md", "tests/cases.json"} {
		if bodies[required] == nil {
			return Package{}, fmt.Errorf("missing required file %s", required)
		}
	}
	var m Manifest
	if err = strictJSON(bodies["capability.json"], &m); err != nil {
		return Package{}, fmt.Errorf("capability manifest: %w", err)
	}
	if m.ContractVersion != 1 || !slugPattern.MatchString(m.Slug) || filepath.Base(abs) != m.Slug || !versionPattern.MatchString(m.Version) || m.Profile != "instruction-only" || m.CodexCompatibility != "skills-v1" {
		return Package{}, errors.New("unsupported capability manifest identity or profile")
	}
	for name, value := range map[string]string{"scripts": m.Scripts, "dependencies": m.Dependencies, "network": m.Network, "credentials": m.Credentials, "background": m.Background, "durable_data": m.DurableData, "publishing": m.Publishing} {
		if value != "none" {
			return Package{}, fmt.Errorf("%s must be none", name)
		}
	}
	if m.DisplayName == "" || m.Summary == "" || len(m.Triggers) == 0 || m.SuccessBehavior == "" || m.FailureBehavior == "" {
		return Package{}, errors.New("capability manifest is incomplete")
	}
	name, description, err := parseFrontmatter(bodies["skill/SKILL.md"])
	if err != nil {
		return Package{}, err
	}
	if name != m.Slug || description == "" {
		return Package{}, errors.New("SKILL.md frontmatter does not match capability manifest")
	}
	var cases Cases
	if err = strictJSON(bodies["tests/cases.json"], &cases); err != nil {
		return Package{}, fmt.Errorf("capability cases: %w", err)
	}
	if cases.ContractVersion != 1 || len(cases.PositiveTriggers) == 0 || len(cases.NonTriggers) == 0 || len(cases.Examples) == 0 {
		return Package{}, errors.New("deterministic cases are incomplete")
	}
	sort.Strings(files)
	sourceDigest := treeDigest(files, bodies)
	projection := map[string][]byte{}
	for path, body := range bodies {
		if path == "skill/SKILL.md" {
			projection["SKILL.md"] = body
		}
		if strings.HasPrefix(path, "skill/references/") {
			projection[strings.TrimPrefix(path, "skill/")] = body
		}
		if strings.HasPrefix(path, "skill/assets/") {
			projection[strings.TrimPrefix(path, "skill/")] = body
		}
	}
	projection["agents/openai.yaml"] = []byte("policy:\n  allow_implicit_invocation: false\n")
	return Package{Root: abs, Manifest: m, Cases: cases, Files: files, SourceDigest: sourceDigest, ProjectionDigest: mapDigest(projection), Projection: projection}, nil
}

func allowedDir(rel string) bool {
	return rel == "skill" || rel == "tests" || rel == "skill/references" || strings.HasPrefix(rel, "skill/references/") || rel == "skill/assets" || strings.HasPrefix(rel, "skill/assets/")
}
func allowedFile(rel string) bool {
	return rel == "capability.json" || rel == "skill/SKILL.md" || rel == "tests/cases.json" || strings.HasPrefix(rel, "skill/references/") || strings.HasPrefix(rel, "skill/assets/")
}
func parseFrontmatter(body []byte) (string, string, error) {
	if !utf8.Valid(body) {
		return "", "", errors.New("SKILL.md must be UTF-8")
	}
	lines := strings.Split(string(body), "\n")
	if len(lines) < 5 || lines[0] != "---" {
		return "", "", errors.New("SKILL.md frontmatter missing")
	}
	var n, d string
	for _, line := range lines[1:] {
		if line == "---" {
			break
		}
		if strings.HasPrefix(line, "name: ") {
			n = strings.TrimSpace(strings.TrimPrefix(line, "name: "))
		}
		if strings.HasPrefix(line, "description: ") {
			d = strings.TrimSpace(strings.TrimPrefix(line, "description: "))
		}
	}
	return n, d, nil
}
func hash(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func treeDigest(files []string, bodies map[string][]byte) string {
	var b bytes.Buffer
	for _, p := range files {
		b.WriteString(p)
		b.WriteByte(0)
		b.WriteString(hash(bodies[p]))
		b.WriteByte(0)
	}
	return hash(b.Bytes())
}
func mapDigest(files map[string][]byte) string {
	names := make([]string, 0, len(files))
	for n := range files {
		names = append(names, n)
	}
	sort.Strings(names)
	return treeDigest(names, files)
}

type Action string

const (
	ActionInstall Action = "install"
	ActionUpgrade Action = "upgrade"
	ActionEnable  Action = "enable"
	ActionDisable Action = "disable"
	ActionRemove  Action = "remove"
)

type State string

const (
	StateAbsent           State = "absent"
	StateDraftInvalid     State = "draft-invalid"
	StateReady            State = "ready"
	StateInstalledHealthy State = "installed-healthy"
	StateSourceChanged    State = "source-changed"
	StateInstalledDrift   State = "installed-drift"
	StateDisabled         State = "disabled"
)

type Receipt struct {
	ContractVersion  int      `json:"contract_version"`
	Slug             string   `json:"slug"`
	Version          string   `json:"version"`
	SourcePath       string   `json:"source_path"`
	SourceDigest     string   `json:"source_digest"`
	ProjectionDigest string   `json:"projection_digest"`
	State            State    `json:"state"`
	Generation       int      `json:"generation"`
	Files            []string `json:"files"`
}
type Status struct {
	State   State
	Package *Package
	Receipt *Receipt
}
type LifecyclePlan struct {
	Action     Action
	Instance   string
	Package    Package
	Receipt    *Receipt
	Projection string
	Control    string
	Summary    []string
}

type transaction struct {
	ContractVersion  int    `json:"contract_version"`
	Action           Action `json:"action"`
	Slug             string `json:"slug"`
	SourceDigest     string `json:"source_digest"`
	ProjectionDigest string `json:"projection_digest"`
}

var mutationHook func(string) error

func InitializeInstance(root string) error {
	for _, p := range []string{filepath.Join(root, "capabilities"), filepath.Join(root, "workspace", ".agents", "skills")} {
		if err := os.MkdirAll(p, 0o700); err != nil {
			return err
		}
	}
	return nil
}
func receiptPath(root, slug string) string {
	return filepath.Join(root, "capabilities", slug, "receipt.json")
}
func projectionPath(root, slug string) string {
	return filepath.Join(root, "workspace", ".agents", "skills", slug)
}
func readReceipt(root, slug string) (*Receipt, error) {
	b, err := os.ReadFile(receiptPath(root, slug))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var r Receipt
	if err = strictJSON(b, &r); err != nil {
		return nil, err
	}
	if r.ContractVersion != 1 || r.Slug != slug {
		return nil, errors.New("invalid capability receipt")
	}
	return &r, nil
}
func projectionDigest(path string) (string, error) {
	bodies := map[string][]byte{}
	var names []string
	err := filepath.WalkDir(path, func(p string, e fs.DirEntry, we error) error {
		if we != nil {
			return we
		}
		rel, _ := filepath.Rel(path, p)
		rel = filepath.ToSlash(rel)
		if rel == "." {
			return nil
		}
		i, err := os.Lstat(p)
		if err != nil {
			return err
		}
		if i.Mode()&os.ModeSymlink != 0 {
			return errors.New("projection link refused")
		}
		if i.IsDir() {
			return nil
		}
		if !i.Mode().IsRegular() {
			return errors.New("projection special file refused")
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		names = append(names, rel)
		bodies[rel] = b
		return nil
	})
	if err != nil {
		return "", err
	}
	sort.Strings(names)
	return treeDigest(names, bodies), nil
}

func Inspect(instance, source string) (Status, error) {
	pkg, err := Validate(source)
	if err != nil {
		return Status{State: StateDraftInvalid}, err
	}
	r, err := readReceipt(instance, pkg.Manifest.Slug)
	if err != nil {
		return Status{}, err
	}
	if r == nil {
		return Status{State: StateReady, Package: &pkg}, nil
	}
	proj := projectionPath(instance, pkg.Manifest.Slug)
	_, statErr := os.Lstat(proj)
	if r.State == StateDisabled {
		if statErr == nil {
			return Status{State: StateInstalledDrift, Package: &pkg, Receipt: r}, errors.New("disabled projection unexpectedly exists")
		}
		return Status{State: StateDisabled, Package: &pkg, Receipt: r}, nil
	}
	if statErr != nil {
		return Status{State: StateInstalledDrift, Package: &pkg, Receipt: r}, errors.New("installed projection missing")
	}
	d, err := projectionDigest(proj)
	if err != nil || d != r.ProjectionDigest {
		return Status{State: StateInstalledDrift, Package: &pkg, Receipt: r}, errors.New("installed projection drift")
	}
	if pkg.SourceDigest != r.SourceDigest {
		return Status{State: StateSourceChanged, Package: &pkg, Receipt: r}, nil
	}
	return Status{State: StateInstalledHealthy, Package: &pkg, Receipt: r}, nil
}

func Plan(instance, source string, action Action) (LifecyclePlan, error) {
	status, err := Inspect(instance, source)
	if err != nil {
		return LifecyclePlan{}, err
	}
	p := *status.Package
	journal := filepath.Join(instance, "capabilities", p.Manifest.Slug, "transaction.json")
	if _, statErr := os.Lstat(journal); statErr == nil {
		return LifecyclePlan{}, errors.New("capability recovery required")
	} else if !errors.Is(statErr, os.ErrNotExist) {
		return LifecyclePlan{}, statErr
	}
	switch action {
	case ActionInstall:
		if status.Receipt != nil {
			return LifecyclePlan{}, errors.New("install requires absent control state")
		}
	case ActionUpgrade:
		if status.State != StateSourceChanged {
			return LifecyclePlan{}, fmt.Errorf("upgrade requires source-changed state, got %s", status.State)
		}
	case ActionDisable:
		if status.State != StateInstalledHealthy {
			return LifecyclePlan{}, fmt.Errorf("disable requires installed-healthy state, got %s", status.State)
		}
	case ActionEnable:
		if status.State != StateDisabled {
			return LifecyclePlan{}, fmt.Errorf("enable requires disabled state, got %s", status.State)
		}
	case ActionRemove:
		if status.State != StateInstalledHealthy && status.State != StateDisabled {
			return LifecyclePlan{}, fmt.Errorf("remove requires healthy or disabled state, got %s", status.State)
		}
	default:
		return LifecyclePlan{}, errors.New("unsupported capability action")
	}
	return LifecyclePlan{Action: action, Instance: instance, Package: p, Receipt: status.Receipt, Projection: projectionPath(instance, p.Manifest.Slug), Control: filepath.Join(instance, "capabilities", p.Manifest.Slug), Summary: []string{"source " + p.SourceDigest, "projection " + p.ProjectionDigest, "fresh Codex task required"}}, nil
}

func (p LifecyclePlan) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Action: %s\nCapability: %s %s\nSource: %s\nSource digest: %s\nProjection: %s\n", p.Action, p.Package.Manifest.Slug, p.Package.Manifest.Version, p.Package.Root, p.Package.SourceDigest, p.Projection)
	for _, s := range p.Summary {
		fmt.Fprintf(&b, "- %s\n", s)
	}
	return b.String()
}
func writeProjection(root string, files map[string][]byte) error {
	stage := root + ".new"
	if err := os.Mkdir(stage, 0o700); err != nil {
		return err
	}
	clean := func() { _ = os.RemoveAll(stage) }
	for rel, b := range files {
		dst := filepath.Join(stage, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(dst), 0o700); err != nil {
			clean()
			return err
		}
		if err := os.WriteFile(dst, b, 0o600); err != nil {
			clean()
			return err
		}
	}
	if err := os.Rename(stage, root); err != nil {
		clean()
		return err
	}
	return nil
}

func readProjection(root string) (map[string][]byte, error) {
	files := map[string][]byte{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, _ := filepath.Rel(root, path)
		rel = filepath.ToSlash(rel)
		if rel == "." {
			return nil
		}
		info, err := os.Lstat(path)
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 || (!info.IsDir() && !info.Mode().IsRegular()) {
			return errors.New("unsafe retained generation")
		}
		if info.IsDir() {
			return nil
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		files[rel] = body
		return nil
	})
	return files, err
}
func writeReceipt(path string, r Receipt) error {
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	tmp := path + ".new"
	if err = os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
func Execute(p LifecyclePlan) (resultErr error) {
	lockFile, err := os.Open(filepath.Join(p.Instance, "capabilities"))
	if err != nil {
		return err
	}
	defer lockFile.Close()
	if err = syscall.Flock(int(lockFile.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		return errors.New("capability transaction already active")
	}
	defer syscall.Flock(int(lockFile.Fd()), syscall.LOCK_UN)
	fresh, err := Plan(p.Instance, p.Package.Root, p.Action)
	if err != nil {
		return fmt.Errorf("stale capability plan: %w", err)
	}
	if fresh.Package.SourceDigest != p.Package.SourceDigest || fresh.Package.ProjectionDigest != p.Package.ProjectionDigest {
		return errors.New("stale capability plan: source changed")
	}
	if err = os.MkdirAll(filepath.Dir(p.Control), 0o700); err != nil {
		return err
	}
	if err = os.MkdirAll(p.Control, 0o700); err != nil {
		return err
	}
	journalPath := filepath.Join(p.Control, "transaction.json")
	journalBody, _ := json.MarshalIndent(transaction{1, p.Action, p.Package.Manifest.Slug, p.Package.SourceDigest, p.Package.ProjectionDigest}, "", "  ")
	journalBody = append(journalBody, '\n')
	if err = os.WriteFile(journalPath, journalBody, 0o600); err != nil {
		return err
	}
	if mutationHook != nil {
		if err = mutationHook("journal-written"); err != nil {
			return err
		}
	}
	defer func() {
		if resultErr == nil {
			_ = os.Remove(journalPath)
		}
	}()
	switch p.Action {
	case ActionInstall, ActionUpgrade, ActionEnable:
		if p.Action == ActionUpgrade {
			if err = os.RemoveAll(p.Projection); err != nil {
				return err
			}
		}
		projection := p.Package.Projection
		if p.Action == ActionEnable {
			projection, err = readProjection(filepath.Join(p.Control, "generations", p.Receipt.ProjectionDigest, "projection"))
			if err != nil {
				return fmt.Errorf("retained generation unavailable: %w", err)
			}
		}
		if err = writeProjection(p.Projection, projection); err != nil {
			return err
		}
		if mutationHook != nil {
			if err = mutationHook("projection-written"); err != nil {
				return err
			}
		}
		generation := 1
		if p.Receipt != nil {
			generation = p.Receipt.Generation + 1
		}
		generationRoot := filepath.Join(p.Control, "generations", p.Package.ProjectionDigest, "projection")
		if p.Action != ActionEnable {
			if err = os.MkdirAll(filepath.Dir(generationRoot), 0o700); err != nil {
				return err
			}
			if _, statErr := os.Lstat(generationRoot); errors.Is(statErr, os.ErrNotExist) {
				if err = writeProjection(generationRoot, p.Package.Projection); err != nil {
					return err
				}
			}
		}
		if err = os.MkdirAll(filepath.Dir(generationRoot), 0o700); err != nil {
			return err
		}
		r := Receipt{1, p.Package.Manifest.Slug, p.Package.Manifest.Version, p.Package.Root, p.Package.SourceDigest, p.Package.ProjectionDigest, StateInstalledHealthy, generation, sortedKeys(p.Package.Projection)}
		return writeReceipt(receiptPath(p.Instance, p.Package.Manifest.Slug), r)
	case ActionDisable:
		if err = os.RemoveAll(p.Projection); err != nil {
			return err
		}
		r := *p.Receipt
		r.State = StateDisabled
		return writeReceipt(receiptPath(p.Instance, p.Package.Manifest.Slug), r)
	case ActionRemove:
		if p.Receipt.State == StateInstalledHealthy {
			if err = os.RemoveAll(p.Projection); err != nil {
				return err
			}
		}
		return os.RemoveAll(p.Control)
	}
	return nil
}

// Recover restores the receipt-declared stable state after an interrupted
// lifecycle mutation. It never consults or deletes source.
func Recover(instance, slug string) error {
	if !slugPattern.MatchString(slug) {
		return errors.New("invalid capability slug")
	}
	lockFile, err := os.Open(filepath.Join(instance, "capabilities"))
	if err != nil {
		return err
	}
	defer lockFile.Close()
	if err = syscall.Flock(int(lockFile.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		return errors.New("capability transaction already active")
	}
	defer syscall.Flock(int(lockFile.Fd()), syscall.LOCK_UN)
	control := filepath.Join(instance, "capabilities", slug)
	journal := filepath.Join(control, "transaction.json")
	if _, err = os.Lstat(journal); err != nil {
		return fmt.Errorf("capability recovery not required: %w", err)
	}
	r, err := readReceipt(instance, slug)
	if err != nil {
		return err
	}
	projection := projectionPath(instance, slug)
	if r == nil {
		if err = os.RemoveAll(projection); err != nil {
			return err
		}
		return os.RemoveAll(control)
	}
	if r.State == StateDisabled {
		if err = os.RemoveAll(projection); err != nil {
			return err
		}
	} else {
		if err = os.RemoveAll(projection); err != nil {
			return err
		}
		files, readErr := readProjection(filepath.Join(control, "generations", r.ProjectionDigest, "projection"))
		if readErr != nil {
			return readErr
		}
		if err = writeProjection(projection, files); err != nil {
			return err
		}
	}
	return os.Remove(journal)
}
func sortedKeys(m map[string][]byte) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
