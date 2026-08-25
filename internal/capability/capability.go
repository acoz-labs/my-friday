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
	"slices"
	"sort"
	"strings"
	"syscall"
	"unicode/utf8"

	"golang.org/x/sys/unix"
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
	if err != nil {
		return Package{}, err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
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
	if cases.ContractVersion != 1 {
		return Package{}, errors.New("unsupported deterministic cases contract")
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

// TestCases executes the bounded deterministic assertions declared by a
// structurally valid instruction-only package. It does not run a model.
func TestCases(pkg Package) error {
	normalize := func(value string) string { return strings.ToLower(strings.TrimSpace(value)) }
	unique := func(label string, values []string) (map[string]bool, error) {
		if len(values) == 0 {
			return nil, fmt.Errorf("%s are required", label)
		}
		seen := map[string]bool{}
		for _, value := range values {
			value = normalize(value)
			if value == "" || seen[value] {
				return nil, fmt.Errorf("%s contain an empty or duplicate value", label)
			}
			seen[value] = true
		}
		return seen, nil
	}
	positive, err := unique("positive triggers", pkg.Cases.PositiveTriggers)
	if err != nil {
		return err
	}
	negative, err := unique("non-triggers", pkg.Cases.NonTriggers)
	if err != nil {
		return err
	}
	for value := range positive {
		if negative[value] {
			return fmt.Errorf("trigger %q is both positive and negative", value)
		}
	}
	for _, trigger := range pkg.Manifest.Triggers {
		if !positive[normalize(trigger)] {
			return fmt.Errorf("manifest trigger %q lacks a positive case", trigger)
		}
	}
	manifestTriggers := map[string]bool{}
	for _, trigger := range pkg.Manifest.Triggers {
		manifestTriggers[normalize(trigger)] = true
	}
	for trigger := range positive {
		if !manifestTriggers[trigger] {
			return fmt.Errorf("positive trigger %q is not declared by the manifest", trigger)
		}
	}
	if len(pkg.Cases.Examples) == 0 {
		return errors.New("examples are required")
	}
	searchable := normalize(pkg.Manifest.SuccessBehavior + "\n" + string(pkg.Projection["SKILL.md"]))
	for i, example := range pkg.Cases.Examples {
		input := normalize(example.Input)
		if input == "" {
			return fmt.Errorf("example %d input is empty", i+1)
		}
		covered := false
		for trigger := range positive {
			if strings.Contains(input, trigger) {
				covered = true
				break
			}
		}
		if !covered {
			return fmt.Errorf("example %d does not exercise a positive trigger", i+1)
		}
		outputs, outputErr := unique(fmt.Sprintf("example %d output expectations", i+1), example.OutputContains)
		if outputErr != nil {
			return outputErr
		}
		for expected := range outputs {
			if !strings.Contains(searchable, expected) {
				return fmt.Errorf("example %d output expectation %q is not declared by success behavior or instructions", i+1, expected)
			}
		}
	}
	facts, err := unique("required facts", pkg.Cases.RequiredFacts)
	if err != nil {
		return err
	}
	factText := normalize(pkg.Manifest.Summary + "\n" + pkg.Manifest.SuccessBehavior + "\n" + pkg.Manifest.FailureBehavior + "\n" + string(pkg.Projection["SKILL.md"]))
	for fact := range facts {
		if fact == "explicit invocation only" {
			continue
		}
		if !strings.Contains(factText, fact) {
			return fmt.Errorf("required fact %q is not declared by the capability", fact)
		}
	}
	forbidden, err := unique("forbidden effects", pkg.Cases.ForbiddenEffects)
	if err != nil {
		return err
	}
	for _, effect := range []string{"scripts", "dependencies", "network", "credentials", "background", "durable-data", "publishing"} {
		if !forbidden[effect] {
			return fmt.Errorf("forbidden effect %q is not asserted", effect)
		}
	}
	return nil
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
	StateDraftValid       State = "draft-valid"
	StateTestFailed       State = "test-failed"
	StateReady            State = "ready"
	StateInstalledHealthy State = "installed-healthy"
	StateSourceChanged    State = "source-changed"
	StateInstalledDrift   State = "installed-drift"
	StateDisabled         State = "disabled"
	StateCollision        State = "collision"
	StateInterrupted      State = "interrupted"
	StateRecoveryRequired State = "recovery-required"
	StateIncompatible     State = "incompatible"
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
	Generations      []string `json:"generations"`
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
	ContractVersion    int    `json:"contract_version"`
	Action             Action `json:"action"`
	Slug               string `json:"slug"`
	SourceDigest       string `json:"source_digest"`
	ProjectionDigest   string `json:"projection_digest"`
	PriorReceiptDigest string `json:"prior_receipt_digest"`
	CreatedControl     bool   `json:"created_control"`
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
	path := receiptPath(root, slug)
	b, info, err := regularFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if info.Mode().Perm() != 0o600 {
		return nil, errors.New("unsafe capability receipt mode")
	}
	var r Receipt
	if err = strictJSON(b, &r); err != nil {
		return nil, err
	}
	if r.ContractVersion != 1 || r.Slug != slug || !slugPattern.MatchString(r.Slug) ||
		r.Version == "" || !filepath.IsAbs(r.SourcePath) || !validDigest(r.SourceDigest) ||
		!validDigest(r.ProjectionDigest) || (r.State != StateInstalledHealthy && r.State != StateDisabled) ||
		r.Generation < 1 || !sortedUniqueStrings(r.Files) || !sortedUniqueDigests(r.Generations) ||
		!slices.Contains(r.Generations, r.ProjectionDigest) {
		return nil, errors.New("invalid capability receipt")
	}
	return &r, nil
}
func regularFile(path string) ([]byte, fs.FileInfo, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, nil, err
	}
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 || st.Nlink != 1 {
		return nil, info, errors.New("unsafe owned regular file")
	}
	b, err := os.ReadFile(path)
	return b, info, err
}
func receiptDigest(r *Receipt) string {
	if r == nil {
		return ""
	}
	b, _ := json.Marshal(r)
	return hash(b)
}
func readTransaction(path, slug string, r *Receipt) (transaction, error) {
	b, info, err := regularFile(path)
	if err != nil {
		return transaction{}, err
	}
	if info.Mode().Perm() != 0o600 {
		return transaction{}, errors.New("unsafe transaction journal mode")
	}
	var tx transaction
	if err = strictJSON(b, &tx); err != nil {
		return tx, err
	}
	validAction := tx.Action == ActionInstall || tx.Action == ActionUpgrade || tx.Action == ActionEnable || tx.Action == ActionDisable || tx.Action == ActionRemove
	receiptMatches := tx.PriorReceiptDigest == receiptDigest(r)
	completedDisableReceipt := tx.Action == ActionDisable && r != nil && r.State == StateDisabled && r.SourceDigest == tx.SourceDigest && r.ProjectionDigest == tx.ProjectionDigest
	if tx.ContractVersion != 1 || tx.Slug != slug || !validAction || !validDigest(tx.SourceDigest) || !validDigest(tx.ProjectionDigest) || (!receiptMatches && !completedDisableReceipt) {
		return tx, errors.New("transaction journal ownership mismatch")
	}
	if tx.Action == ActionInstall && (!tx.CreatedControl || r != nil) {
		return tx, errors.New("invalid install journal authority")
	}
	if tx.Action != ActionInstall && (tx.CreatedControl || r == nil) {
		return tx, errors.New("invalid existing-state journal authority")
	}
	canonical, _ := json.MarshalIndent(tx, "", "  ")
	canonical = append(canonical, '\n')
	if !bytes.Equal(b, canonical) {
		return tx, errors.New("transaction journal is not canonical")
	}
	return tx, nil
}
func validDigest(value string) bool {
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == sha256.Size && value == strings.ToLower(value)
}
func sortedUniqueStrings(values []string) bool {
	if len(values) == 0 {
		return false
	}
	for i, value := range values {
		if value == "" || (i > 0 && values[i-1] >= value) {
			return false
		}
	}
	return true
}
func sortedUniqueDigests(values []string) bool {
	if !sortedUniqueStrings(values) {
		return false
	}
	for _, value := range values {
		if !validDigest(value) {
			return false
		}
	}
	return true
}
func validateManagedControl(instance, slug string, r *Receipt, allowJournal bool) error {
	control := filepath.Join(instance, "capabilities", slug)
	info, err := os.Lstat(control)
	if err != nil {
		return err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 || info.Mode().Perm() != 0o700 {
		return errors.New("unsafe managed control directory")
	}
	allowed := map[string]bool{"receipt.json": true, "generations": true}
	if allowJournal {
		allowed["transaction.json"] = true
	}
	entries, err := os.ReadDir(control)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if !allowed[e.Name()] {
			return fmt.Errorf("foreign capability control entry %s", e.Name())
		}
	}
	generations := filepath.Join(control, "generations")
	generationEntries, err := os.ReadDir(generations)
	if err != nil {
		return err
	}
	wantGenerations := make(map[string]bool, len(r.Generations))
	for _, generation := range r.Generations {
		wantGenerations[generation] = true
	}
	foundGenerations := make(map[string]bool, len(generationEntries))
	for _, e := range generationEntries {
		if !e.IsDir() || !wantGenerations[e.Name()] {
			return errors.New("invalid retained generation entry")
		}
		d, digestErr := projectionDigest(filepath.Join(generations, e.Name(), "projection"))
		if digestErr != nil || d != e.Name() {
			return errors.New("retained generation drift")
		}
		foundGenerations[e.Name()] = true
	}
	if len(foundGenerations) != len(wantGenerations) {
		return errors.New("receipt-declared retained generation missing")
	}
	return nil
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
		st, ok := i.Sys().(*syscall.Stat_t)
		if !ok || i.Mode()&os.ModeSymlink != 0 || st.Uid != uint32(os.Getuid()) {
			return errors.New("projection link refused")
		}
		if i.IsDir() {
			if i.Mode().Perm() != 0o700 {
				return errors.New("unsafe projection directory mode")
			}
			return nil
		}
		if !i.Mode().IsRegular() || i.Mode().Perm() != 0o600 || st.Nlink != 1 {
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
	slug := filepath.Base(filepath.Clean(source))
	control := filepath.Join(instance, "capabilities", slug)
	proj := projectionPath(instance, slug)
	_, controlErr := os.Lstat(control)
	controlExists := controlErr == nil
	_, projectionErr := os.Lstat(proj)
	projectionExists := projectionErr == nil
	r, receiptErr := readReceipt(instance, slug)
	if receiptErr != nil {
		return Status{State: StateIncompatible}, receiptErr
	}
	journal := filepath.Join(control, "transaction.json")
	journalExists := false
	if _, journalErr := os.Lstat(journal); journalErr == nil {
		journalExists = true
	}
	if r != nil {
		if controlValidationErr := validateManagedControl(instance, slug, r, journalExists); controlValidationErr != nil {
			return Status{State: StateCollision, Receipt: r}, controlValidationErr
		}
	}
	if _, journalErr := os.Lstat(journal); journalErr == nil {
		if _, err := readTransaction(journal, slug, r); err != nil {
			return Status{State: StateRecoveryRequired, Receipt: r}, err
		}
		return Status{State: StateInterrupted, Receipt: r}, nil
	} else if !errors.Is(journalErr, os.ErrNotExist) {
		return Status{State: StateRecoveryRequired, Receipt: r}, journalErr
	}
	if r == nil && (controlExists || projectionExists) {
		return Status{State: StateCollision}, errors.New("foreign capability projection or control collision")
	}
	pkg, err := Validate(source)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) && r == nil {
			return Status{State: StateAbsent}, nil
		}
		state := StateDraftInvalid
		if strings.Contains(err.Error(), "missing required file tests/cases.json") {
			state = StateDraftValid
		}
		if strings.Contains(err.Error(), "capability cases") || strings.Contains(err.Error(), "deterministic cases") {
			state = StateTestFailed
		}
		return Status{State: state, Receipt: r}, err
	}
	if err = TestCases(pkg); err != nil {
		return Status{State: StateTestFailed, Package: &pkg, Receipt: r}, err
	}
	if r == nil {
		return Status{State: StateReady, Package: &pkg}, nil
	}
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
	if status.State == StateInterrupted || status.State == StateRecoveryRequired {
		return LifecyclePlan{}, errors.New("capability recovery required")
	}
	if status.Package == nil {
		return LifecyclePlan{}, fmt.Errorf("capability state %s is not actionable", status.State)
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
	parent, err := openDirNoFollow(filepath.Dir(root))
	if err != nil {
		return err
	}
	defer unix.Close(parent)
	name := filepath.Base(root)
	stageName := name + ".new"
	stage := filepath.Join(filepath.Dir(root), stageName)
	if err = unix.Mkdirat(parent, stageName, 0o700); err != nil {
		return err
	}
	clean := func() {
		if d, e := projectionDigest(stage); e == nil && d == mapDigest(files) {
			if quarantine, moveErr := quarantineOwnedProjection(stage, d, "stage-cleanup-quarantined"); moveErr == nil {
				_ = removeProvenQuarantine(quarantine, d)
			}
		}
	}
	stageFD, err := unix.Openat(parent, stageName, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if err != nil {
		return err
	}
	defer unix.Close(stageFD)
	for rel, b := range files {
		if err := writeProjectionFileAt(stageFD, rel, b); err != nil {
			clean()
			return err
		}
	}
	if d, digestErr := projectionDigest(stage); digestErr != nil || d != mapDigest(files) {
		return errors.New("staged projection identity changed")
	}
	if err := renameExclusive(parent, stageName, parent, name); err != nil {
		clean()
		return err
	}
	if d, digestErr := projectionDigest(root); digestErr != nil || d != mapDigest(files) {
		return errors.New("promoted projection identity changed")
	}
	return nil
}

func writeProjectionFileAt(rootFD int, rel string, body []byte) error {
	parts := strings.Split(filepath.ToSlash(rel), "/")
	if len(parts) == 0 || len(parts) > MaxDepth {
		return errors.New("invalid projection path")
	}
	fd, err := unix.Dup(rootFD)
	if err != nil {
		return err
	}
	defer unix.Close(fd)
	for _, part := range parts[:len(parts)-1] {
		if part == "" || part == "." || part == ".." {
			return errors.New("invalid projection path")
		}
		if err = unix.Mkdirat(fd, part, 0o700); err != nil && !errors.Is(err, unix.EEXIST) {
			return err
		}
		next, openErr := unix.Openat(fd, part, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
		if openErr != nil {
			return openErr
		}
		unix.Close(fd)
		fd = next
	}
	name := parts[len(parts)-1]
	if name == "" || name == "." || name == ".." {
		return errors.New("invalid projection path")
	}
	fileFD, err := unix.Openat(fd, name, unix.O_WRONLY|unix.O_CREAT|unix.O_EXCL|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0o600)
	if err != nil {
		return err
	}
	f := os.NewFile(uintptr(fileFD), name)
	_, writeErr := f.Write(body)
	closeErr := f.Close()
	if writeErr != nil {
		return writeErr
	}
	return closeErr
}

func openDirNoFollow(path string) (int, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return -1, err
	}
	if abs == "/var" || strings.HasPrefix(abs, "/var/") {
		abs = filepath.Join("/private/var", strings.TrimPrefix(abs, "/var/"))
	}
	fd, err := unix.Open("/", unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if err != nil {
		return -1, err
	}
	for _, part := range strings.Split(strings.TrimPrefix(filepath.Clean(abs), "/"), "/") {
		if part == "" {
			continue
		}
		next, openErr := unix.Openat(fd, part, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
		unix.Close(fd)
		if openErr != nil {
			return -1, openErr
		}
		fd = next
	}
	return fd, nil
}

type projectionProof struct {
	dev, ino uint64
	digest   string
}

type cleanupManifest struct {
	ContractVersion int               `json:"contract_version"`
	RootDev         uint64            `json:"root_dev"`
	RootIno         uint64            `json:"root_ino"`
	ExpectedDigest  string            `json:"expected_digest"`
	Directories     []string          `json:"directories"`
	Files           map[string]string `json:"files"`
}

func cleanupManifestPath(path string) string     { return path + ".cleanup.json" }
func cleanupManifestTempPath(path string) string { return cleanupManifestPath(path) + ".new" }

func snapshotCleanupTree(path, expected string) (cleanupManifest, error) {
	parent, err := openDirNoFollow(filepath.Dir(path))
	if err != nil {
		return cleanupManifest{}, err
	}
	defer unix.Close(parent)
	proof, err := proveProjection(parent, filepath.Base(path), path, expected)
	if err != nil {
		return cleanupManifest{}, err
	}
	m := cleanupManifest{ContractVersion: 1, RootDev: proof.dev, RootIno: proof.ino, ExpectedDigest: expected, Files: map[string]string{}}
	err = filepath.WalkDir(path, func(p string, e fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, _ := filepath.Rel(path, p)
		rel = filepath.ToSlash(rel)
		if rel == "." {
			return nil
		}
		info, statErr := os.Lstat(p)
		if statErr != nil {
			return statErr
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return errors.New("unsafe cleanup snapshot link")
		}
		if info.IsDir() {
			m.Directories = append(m.Directories, rel)
			return nil
		}
		if !info.Mode().IsRegular() {
			return errors.New("unsafe cleanup snapshot entry")
		}
		body, readErr := os.ReadFile(p)
		if readErr != nil {
			return readErr
		}
		m.Files[rel] = hash(body)
		return nil
	})
	sort.Strings(m.Directories)
	return m, err
}

func writeCleanupManifest(path string, m cleanupManifest) error {
	body, _ := json.MarshalIndent(m, "", "  ")
	body = append(body, '\n')
	temp := cleanupManifestTempPath(path)
	f, err := os.OpenFile(temp, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	_, writeErr := f.Write(body)
	syncErr := f.Sync()
	closeErr := f.Close()
	if writeErr != nil {
		return writeErr
	}
	if syncErr != nil {
		return syncErr
	}
	if closeErr != nil {
		return closeErr
	}
	if mutationHook != nil {
		if err = mutationHook("cleanup-manifest-temp-synced"); err != nil {
			return err
		}
	}
	parent, err := openDirNoFollow(filepath.Dir(path))
	if err != nil {
		return err
	}
	defer unix.Close(parent)
	return renameExclusive(parent, filepath.Base(temp), parent, filepath.Base(cleanupManifestPath(path)))
}

func readCleanupManifest(path string) (cleanupManifest, error) {
	return readCleanupManifestFile(cleanupManifestPath(path))
}
func readCleanupManifestFile(manifestPath string) (cleanupManifest, error) {
	body, info, err := regularFile(manifestPath)
	if err != nil {
		return cleanupManifest{}, err
	}
	if info.Mode().Perm() != 0o600 {
		return cleanupManifest{}, errors.New("unsafe cleanup manifest mode")
	}
	var m cleanupManifest
	if err = strictJSON(body, &m); err != nil {
		return m, err
	}
	canonical, _ := json.MarshalIndent(m, "", "  ")
	canonical = append(canonical, '\n')
	if !bytes.Equal(body, canonical) || m.ContractVersion != 1 || !validDigest(m.ExpectedDigest) || !sortedUniqueStrings(m.Directories) || len(m.Files) == 0 {
		return m, errors.New("invalid cleanup manifest")
	}
	for rel, digest := range m.Files {
		if rel == "" || !validDigest(digest) {
			return m, errors.New("invalid cleanup manifest entry")
		}
	}
	return m, nil
}

func validateCleanupSubset(path string, m cleanupManifest) error {
	allowedDirs := map[string]bool{}
	for _, rel := range m.Directories {
		allowedDirs[rel] = true
	}
	return filepath.WalkDir(path, func(p string, e fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, _ := filepath.Rel(path, p)
		rel = filepath.ToSlash(rel)
		if rel == "." {
			return nil
		}
		info, err := os.Lstat(p)
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return errors.New("cleanup subset link refused")
		}
		if info.IsDir() {
			if !allowedDirs[rel] {
				return fmt.Errorf("foreign cleanup directory %s", rel)
			}
			return nil
		}
		expected, ok := m.Files[rel]
		if !ok || !info.Mode().IsRegular() {
			return fmt.Errorf("foreign cleanup file %s", rel)
		}
		body, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		if hash(body) != expected {
			return fmt.Errorf("cleanup file drift %s", rel)
		}
		return nil
	})
}

func proveProjection(parent int, name, path, expected string) (projectionProof, error) {
	fd, err := unix.Openat(parent, name, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if err != nil {
		return projectionProof{}, err
	}
	defer unix.Close(fd)
	var opened, entry unix.Stat_t
	if err = unix.Fstat(fd, &opened); err != nil {
		return projectionProof{}, err
	}
	if opened.Uid != uint32(os.Getuid()) || opened.Mode&unix.S_IFMT != unix.S_IFDIR || opened.Mode&0o777 != 0o700 {
		return projectionProof{}, errors.New("unsafe projection owner or mode")
	}
	if err = unix.Fstatat(parent, name, &entry, unix.AT_SYMLINK_NOFOLLOW); err != nil {
		return projectionProof{}, err
	}
	if opened.Dev != entry.Dev || opened.Ino != entry.Ino {
		return projectionProof{}, errors.New("projection identity changed")
	}
	d, err := projectionDigest(path)
	if err != nil || d != expected {
		return projectionProof{}, errors.New("projection ownership mismatch")
	}
	if err = unix.Fstatat(parent, name, &entry, unix.AT_SYMLINK_NOFOLLOW); err != nil {
		return projectionProof{}, err
	}
	if opened.Dev != entry.Dev || opened.Ino != entry.Ino {
		return projectionProof{}, errors.New("projection identity changed during digest")
	}
	return projectionProof{uint64(opened.Dev), uint64(opened.Ino), d}, nil
}

func quarantineOwnedProjection(path, expected, phase string) (string, error) {
	return quarantineOwnedProjectionAs(path, expected, phase, filepath.Base(path)+".owned-"+expected[:16]+".quarantine")
}

func quarantineOwnedProjectionAs(path, expected, phase, quarantineName string) (string, error) {
	parent, err := openDirNoFollow(filepath.Dir(path))
	if err != nil {
		return "", err
	}
	defer unix.Close(parent)
	name := filepath.Base(path)
	proof, err := proveProjection(parent, name, path, expected)
	if err != nil {
		return "", err
	}
	if err = renameExclusive(parent, name, parent, quarantineName); err != nil {
		return "", err
	}
	if mutationHook != nil {
		if err = mutationHook(phase); err != nil {
			return filepath.Join(filepath.Dir(path), quarantineName), err
		}
	}
	quarantinePath := filepath.Join(filepath.Dir(path), quarantineName)
	after, err := proveProjection(parent, quarantineName, quarantinePath, expected)
	if err != nil || after.dev != proof.dev || after.ino != proof.ino {
		return filepath.Join(filepath.Dir(path), quarantineName), errors.New("quarantined projection identity changed")
	}
	return quarantinePath, nil
}

func removeProvenQuarantine(path, expected string) error {
	manifest, manifestErr := readCleanupManifest(path)
	if errors.Is(manifestErr, os.ErrNotExist) {
		manifest, manifestErr = snapshotCleanupTree(path, expected)
		if manifestErr != nil {
			return manifestErr
		}
		if manifestErr = writeCleanupManifest(path, manifest); manifestErr != nil {
			return manifestErr
		}
	} else if manifestErr != nil {
		return manifestErr
	}
	if manifest.ExpectedDigest != expected {
		return errors.New("cleanup manifest digest mismatch")
	}
	if _, targetErr := os.Lstat(path); errors.Is(targetErr, os.ErrNotExist) {
		return os.Remove(cleanupManifestPath(path))
	} else if targetErr != nil {
		return targetErr
	}
	parent, err := openDirNoFollow(filepath.Dir(path))
	if err != nil {
		return err
	}
	defer unix.Close(parent)
	var entry unix.Stat_t
	if err = unix.Fstatat(parent, filepath.Base(path), &entry, unix.AT_SYMLINK_NOFOLLOW); err != nil {
		return err
	}
	proof := projectionProof{dev: uint64(entry.Dev), ino: uint64(entry.Ino), digest: expected}
	if proof.dev != manifest.RootDev || proof.ino != manifest.RootIno {
		return errors.New("cleanup manifest root identity mismatch")
	}
	if err = validateCleanupSubset(path, manifest); err != nil {
		return err
	}
	if mutationHook != nil {
		if err = mutationHook("final-deletion-boundary"); err != nil {
			return err
		}
	}
	if err = unix.Fstatat(parent, filepath.Base(path), &entry, unix.AT_SYMLINK_NOFOLLOW); err != nil || uint64(entry.Dev) != proof.dev || uint64(entry.Ino) != proof.ino {
		return errors.New("tree identity changed at final boundary")
	}
	fd, err := unix.Openat(parent, filepath.Base(path), unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if err != nil {
		return err
	}
	defer unix.Close(fd)
	var opened unix.Stat_t
	if err = unix.Fstat(fd, &opened); err != nil || uint64(opened.Dev) != proof.dev || uint64(opened.Ino) != proof.ino {
		return errors.New("cleanup descriptor identity changed")
	}
	if mutationHook != nil {
		if err = mutationHook("descriptor-cleanup-started"); err != nil {
			return err
		}
	}
	if err = removeDirectoryContents(fd, 0); err != nil {
		return err
	}
	if err = unix.Fstatat(parent, filepath.Base(path), &entry, unix.AT_SYMLINK_NOFOLLOW); err != nil {
		return err
	}
	if uint64(entry.Dev) != proof.dev || uint64(entry.Ino) != proof.ino {
		return errors.New("cleanup root identity changed")
	}
	if err = unix.Unlinkat(parent, filepath.Base(path), unix.AT_REMOVEDIR); err != nil {
		return err
	}
	if mutationHook != nil {
		if err = mutationHook("cleanup-root-unlinked"); err != nil {
			return err
		}
	}
	return os.Remove(cleanupManifestPath(path))
}

func removeDirectoryContents(dirFD, depth int) error {
	dup, err := unix.Dup(dirFD)
	if err != nil {
		return err
	}
	f := os.NewFile(uintptr(dup), "cleanup")
	names, err := f.Readdirnames(-1)
	closeErr := f.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	sort.Strings(names)
	for _, name := range names {
		var before unix.Stat_t
		if err = unix.Fstatat(dirFD, name, &before, unix.AT_SYMLINK_NOFOLLOW); err != nil {
			return err
		}
		if before.Uid != uint32(os.Getuid()) {
			return errors.New("cleanup entry owner changed")
		}
		if before.Mode&unix.S_IFMT == unix.S_IFDIR {
			child, openErr := unix.Openat(dirFD, name, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
			if openErr != nil {
				return openErr
			}
			if err = removeDirectoryContents(child, depth+1); err != nil {
				unix.Close(child)
				return err
			}
			var opened unix.Stat_t
			if err = unix.Fstat(child, &opened); err != nil {
				unix.Close(child)
				return err
			}
			unix.Close(child)
			var current unix.Stat_t
			if err = unix.Fstatat(dirFD, name, &current, unix.AT_SYMLINK_NOFOLLOW); err != nil {
				return err
			}
			if opened.Dev != current.Dev || opened.Ino != current.Ino {
				return errors.New("cleanup directory identity changed")
			}
			if err = unix.Unlinkat(dirFD, name, unix.AT_REMOVEDIR); err != nil {
				return err
			}
		} else {
			if before.Mode&unix.S_IFMT != unix.S_IFREG || before.Nlink != 1 {
				return errors.New("unsafe cleanup entry")
			}
			fd, openErr := unix.Openat(dirFD, name, unix.O_RDONLY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
			if openErr != nil {
				return openErr
			}
			var opened unix.Stat_t
			if err = unix.Fstat(fd, &opened); err != nil {
				unix.Close(fd)
				return err
			}
			unix.Close(fd)
			var current unix.Stat_t
			if err = unix.Fstatat(dirFD, name, &current, unix.AT_SYMLINK_NOFOLLOW); err != nil {
				return err
			}
			if opened.Dev != current.Dev || opened.Ino != current.Ino {
				return errors.New("cleanup file identity changed")
			}
			if err = unix.Unlinkat(dirFD, name, 0); err != nil {
				return err
			}
		}
		if mutationHook != nil {
			phase := "descriptor-root-entry-unlinked"
			if depth > 0 {
				phase = "descriptor-nested-entry-unlinked"
			}
			if err = mutationHook(phase); err != nil {
				return err
			}
		}
	}
	return nil
}

func restoreProvenQuarantine(quarantine, target, expected string) error {
	parent, err := openDirNoFollow(filepath.Dir(target))
	if err != nil {
		return err
	}
	defer unix.Close(parent)
	restoringName := filepath.Base(target) + ".owned-" + expected[:16] + ".restoring"
	restoring := filepath.Join(filepath.Dir(target), restoringName)
	var proof projectionProof
	if _, restoringErr := os.Lstat(restoring); errors.Is(restoringErr, os.ErrNotExist) {
		proof, err = proveProjection(parent, filepath.Base(quarantine), quarantine, expected)
		if err != nil {
			return err
		}
		if err = renameExclusive(parent, filepath.Base(quarantine), parent, restoringName); err != nil {
			return err
		}
	} else if restoringErr != nil {
		return restoringErr
	} else {
		proof, err = proveProjection(parent, restoringName, restoring, expected)
		if err != nil {
			return err
		}
	}
	afterTransfer, err := proveProjection(parent, restoringName, restoring, expected)
	if err != nil || afterTransfer.dev != proof.dev || afterTransfer.ino != proof.ino {
		return errors.New("restoring handle identity changed")
	}
	if mutationHook != nil {
		if err = mutationHook("restore-handle-transferred"); err != nil {
			return err
		}
	}
	if err = renameExclusive(parent, restoringName, parent, filepath.Base(target)); err != nil {
		return err
	}
	after, err := proveProjection(parent, filepath.Base(target), target, expected)
	if err != nil || after.dev != proof.dev || after.ino != proof.ino {
		return errors.New("restored projection identity changed")
	}
	return nil
}

func resumeCleanupManifests(instance, slug string) error {
	patterns := []string{
		filepath.Join(instance, "workspace", ".agents", "skills", slug+"*.cleanup.json"),
		filepath.Join(instance, "capabilities", slug+"*.cleanup.json"),
	}
	for _, pattern := range patterns {
		tempMatches, err := filepath.Glob(pattern + ".new")
		if err != nil {
			return err
		}
		for _, temp := range tempMatches {
			tree := strings.TrimSuffix(temp, ".cleanup.json.new")
			candidate, readErr := readCleanupManifestFile(temp)
			if readErr != nil {
				return fmt.Errorf("unproven cleanup temp: %w", readErr)
			}
			parent, openErr := openDirNoFollow(filepath.Dir(tree))
			if openErr != nil {
				return openErr
			}
			proof, proofErr := proveProjection(parent, filepath.Base(tree), tree, candidate.ExpectedDigest)
			if proofErr != nil || proof.dev != candidate.RootDev || proof.ino != candidate.RootIno {
				unix.Close(parent)
				return errors.New("cleanup temp root proof mismatch")
			}
			var tempInfo unix.Stat_t
			if statErr := unix.Fstatat(parent, filepath.Base(temp), &tempInfo, unix.AT_SYMLINK_NOFOLLOW); statErr != nil {
				unix.Close(parent)
				return statErr
			}
			if unlinkErr := unix.Unlinkat(parent, filepath.Base(temp), 0); unlinkErr != nil {
				unix.Close(parent)
				return unlinkErr
			}
			unix.Close(parent)
		}
	}
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return err
		}
		sort.Strings(matches)
		for _, manifestPath := range matches {
			tree := strings.TrimSuffix(manifestPath, ".cleanup.json")
			manifest, readErr := readCleanupManifest(tree)
			if readErr != nil {
				return readErr
			}
			if err = removeProvenQuarantine(tree, manifest.ExpectedDigest); err != nil {
				return err
			}
		}
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
	createdControl := false
	if p.Action == ActionInstall {
		if err = os.Mkdir(p.Control, 0o700); err != nil {
			return fmt.Errorf("capability control collision: %w", err)
		}
		createdControl = true
	} else {
		if _, err = os.Lstat(p.Control); err != nil {
			return errors.New("managed capability control missing")
		}
	}
	journalPath := filepath.Join(p.Control, "transaction.json")
	journalBody, _ := json.MarshalIndent(transaction{1, p.Action, p.Package.Manifest.Slug, p.Package.SourceDigest, p.Package.ProjectionDigest, receiptDigest(p.Receipt), createdControl}, "", "  ")
	journalBody = append(journalBody, '\n')
	jf, openErr := os.OpenFile(journalPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if openErr != nil {
		return openErr
	}
	_, err = jf.Write(journalBody)
	closeErr := jf.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	if mutationHook != nil {
		if err = mutationHook("journal-written"); err != nil {
			return err
		}
	}
	if p.Action != ActionInstall {
		if err = validateManagedControl(p.Instance, p.Package.Manifest.Slug, p.Receipt, true); err != nil {
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
		var oldQuarantine string
		if p.Action == ActionUpgrade {
			oldQuarantine, err = quarantineOwnedProjection(p.Projection, p.Receipt.ProjectionDigest, "upgrade-projection-quarantined")
			if err != nil {
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
		if oldQuarantine != "" {
			if err = removeProvenQuarantine(oldQuarantine, p.Receipt.ProjectionDigest); err != nil {
				return err
			}
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
		generations := []string{p.Package.ProjectionDigest}
		if p.Receipt != nil {
			generations = append(generations, p.Receipt.Generations...)
			sort.Strings(generations)
			generations = slices.Compact(generations)
		}
		r := Receipt{ContractVersion: 1, Slug: p.Package.Manifest.Slug, Version: p.Package.Manifest.Version, SourcePath: p.Package.Root, SourceDigest: p.Package.SourceDigest, ProjectionDigest: p.Package.ProjectionDigest, State: StateInstalledHealthy, Generation: generation, Files: sortedKeys(p.Package.Projection), Generations: generations}
		return writeReceipt(receiptPath(p.Instance, p.Package.Manifest.Slug), r)
	case ActionDisable:
		quarantine, moveErr := quarantineOwnedProjection(p.Projection, p.Receipt.ProjectionDigest, "disable-projection-quarantined")
		if moveErr != nil {
			return moveErr
		}
		r := *p.Receipt
		r.State = StateDisabled
		if err = writeReceipt(receiptPath(p.Instance, p.Package.Manifest.Slug), r); err != nil {
			return err
		}
		return removeProvenQuarantine(quarantine, p.Receipt.ProjectionDigest)
	case ActionRemove:
		var projectionQuarantine string
		if p.Receipt.State == StateInstalledHealthy {
			projectionQuarantine, err = quarantineOwnedProjection(p.Projection, p.Receipt.ProjectionDigest, "remove-projection-quarantined")
			if err != nil {
				return err
			}
		}
		controlDigest, digestErr := projectionDigest(p.Control)
		if digestErr != nil {
			return digestErr
		}
		controlQuarantine, moveErr := quarantineOwnedProjectionAs(p.Control, controlDigest, "remove-control-quarantined", filepath.Base(p.Control)+".removing")
		if moveErr != nil {
			return moveErr
		}
		if projectionQuarantine != "" {
			if err = removeProvenQuarantine(projectionQuarantine, p.Receipt.ProjectionDigest); err != nil {
				return err
			}
		}
		return removeProvenQuarantine(controlQuarantine, controlDigest)
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
	if err = resumeCleanupManifests(instance, slug); err != nil {
		return fmt.Errorf("recovery refused: cleanup resume: %w", err)
	}
	control := filepath.Join(instance, "capabilities", slug)
	removingControl := control + ".removing"
	if _, controlErr := os.Lstat(control); errors.Is(controlErr, os.ErrNotExist) {
		if _, removingErr := os.Lstat(removingControl); removingErr == nil {
			parent, openErr := openDirNoFollow(filepath.Dir(control))
			if openErr != nil {
				return openErr
			}
			renameErr := renameExclusive(parent, filepath.Base(removingControl), parent, filepath.Base(control))
			unix.Close(parent)
			if renameErr != nil {
				return fmt.Errorf("recovery refused: removing control collision: %w", renameErr)
			}
		}
	}
	journal := filepath.Join(control, "transaction.json")
	r, err := readReceipt(instance, slug)
	if err != nil {
		return err
	}
	if _, journalErr := os.Lstat(journal); errors.Is(journalErr, os.ErrNotExist) {
		projection := projectionPath(instance, slug)
		if r == nil {
			if _, controlErr := os.Lstat(control); errors.Is(controlErr, os.ErrNotExist) {
				if _, projectionErr := os.Lstat(projection); errors.Is(projectionErr, os.ErrNotExist) {
					return nil
				}
			}
			return errors.New("recovery refused: no owned stable state")
		}
		if err = validateManagedControl(instance, slug, r, false); err != nil {
			return err
		}
		if r.State == StateDisabled {
			if _, projectionErr := os.Lstat(projection); errors.Is(projectionErr, os.ErrNotExist) {
				return nil
			}
			return errors.New("recovery refused: disabled projection exists")
		}
		d, digestErr := projectionDigest(projection)
		if digestErr == nil && d == r.ProjectionDigest {
			return nil
		}
		return errors.New("recovery refused: stable projection drift")
	} else if journalErr != nil {
		return journalErr
	}
	tx, err := readTransaction(journal, slug, r)
	if err != nil {
		return fmt.Errorf("recovery refused: %w", err)
	}
	projection := projectionPath(instance, slug)
	expectedProjection := tx.ProjectionDigest
	if r != nil {
		expectedProjection = r.ProjectionDigest
	}
	if expectedProjection != "" {
		quarantine := projection + ".owned-" + expectedProjection[:16] + ".quarantine"
		restoring := projection + ".owned-" + expectedProjection[:16] + ".restoring"
		if _, projectionErr := os.Lstat(projection); errors.Is(projectionErr, os.ErrNotExist) {
			_, quarantineErr := os.Lstat(quarantine)
			_, restoringErr := os.Lstat(restoring)
			if quarantineErr == nil || restoringErr == nil {
				if restoreErr := restoreProvenQuarantine(quarantine, projection, expectedProjection); restoreErr != nil {
					return fmt.Errorf("recovery refused: projection quarantine ownership mismatch: %w", restoreErr)
				}
			}
		}
	}
	info, err := os.Lstat(control)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 || info.Mode().Perm() != 0o700 {
		return errors.New("recovery refused: unsafe control directory")
	}
	entries, err := os.ReadDir(control)
	if err != nil {
		return err
	}
	allowed := map[string]bool{"transaction.json": true}
	if r != nil {
		allowed["receipt.json"] = true
		allowed["generations"] = true
	}
	for _, entry := range entries {
		if !allowed[entry.Name()] {
			return fmt.Errorf("recovery refused: foreign control entry %s", entry.Name())
		}
	}
	if _, statErr := os.Lstat(projection); statErr == nil {
		d, digestErr := projectionDigest(projection)
		if digestErr != nil || (d != tx.ProjectionDigest && (r == nil || d != r.ProjectionDigest)) {
			return errors.New("recovery refused: projection ownership mismatch")
		}
	} else if !errors.Is(statErr, os.ErrNotExist) {
		return statErr
	}
	if mutationHook != nil {
		if err = mutationHook("recovery-ownership-checked"); err != nil {
			return err
		}
	}
	if tx.CreatedControl {
		if _, statErr := os.Lstat(projection); statErr == nil {
			quarantine, moveErr := quarantineOwnedProjection(projection, tx.ProjectionDigest, "recovery-install-projection-quarantined")
			if moveErr != nil {
				return moveErr
			}
			if err = removeProvenQuarantine(quarantine, tx.ProjectionDigest); err != nil {
				return err
			}
		}
		controlDigest, digestErr := projectionDigest(control)
		if digestErr != nil {
			return digestErr
		}
		controlQuarantine, moveErr := quarantineOwnedProjection(control, controlDigest, "recovery-install-control-quarantined")
		if moveErr != nil {
			return moveErr
		}
		return removeProvenQuarantine(controlQuarantine, controlDigest)
	}
	if r.State == StateDisabled {
		if _, statErr := os.Lstat(projection); statErr == nil {
			quarantine, moveErr := quarantineOwnedProjection(projection, tx.ProjectionDigest, "recovery-disabled-projection-quarantined")
			if moveErr != nil {
				return moveErr
			}
			if err = removeProvenQuarantine(quarantine, tx.ProjectionDigest); err != nil {
				return err
			}
		}
	} else {
		if _, statErr := os.Lstat(projection); statErr == nil {
			d, digestErr := projectionDigest(projection)
			if digestErr != nil {
				return digestErr
			}
			quarantine, moveErr := quarantineOwnedProjection(projection, d, "recovery-existing-projection-quarantined")
			if moveErr != nil {
				return moveErr
			}
			if err = removeProvenQuarantine(quarantine, d); err != nil {
				return err
			}
		}
		files, readErr := readProjection(filepath.Join(control, "generations", r.ProjectionDigest, "projection"))
		if readErr != nil {
			return readErr
		}
		if err = writeProjection(projection, files); err != nil {
			return fmt.Errorf("restore receipt-bound projection: %w", err)
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
