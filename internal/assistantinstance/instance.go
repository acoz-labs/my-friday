// Package assistantinstance manages named, manifest-owned assistant instances.
package assistantinstance

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"syscall"
	"unicode/utf8"

	"github.com/acoz-labs/my-friday/internal/profile"
	"github.com/acoz-labs/my-friday/internal/repository"
)

const ContractVersion = 2
const CapabilityRevision = 3

const legacyBuilderSkill = `---
name: capability-builder
description: Help define, scaffold, inspect, validate, and test a My Friday instruction-only capability through its versioned source workflow.
---

# Capability builder

Clarify purpose, explicit triggers, non-triggers, inputs, outputs, examples, and failure behavior. Edit only runtime source under skills/<slug>/ and its deterministic tests. Show the complete Git diff and unresolved judgments before suggesting activation.

You may run capability inspect, validate, and test. Never run install, upgrade, enable, disable, remove, or recover, and never enter a confirmation token. The user must review the CLI plan and authorize mutation independently. Instruction-only validation is structural and does not certify that natural-language instructions are benign.
`

const builderSkillTemplate = `---
name: capability-builder
description: Help define, scaffold, inspect, validate, and test a My Friday instruction-only capability through its versioned source workflow.
---

# Capability builder

You are helping a user build for assistant instance %q. Its exact private runtime source root is %q. Do not edit capability source directly. The deterministic workshop is the sole source-write authority; this runtime root remains the only writable root outside the trusted workspace.

Use only this manifest-owned My Friday executable: %q. For a capability slug, the allowed command forms are:

- %s capability workshop %s <slug>
- %s capability inspect %s <slug> --plain
- %s capability validate %s <slug>
- %s capability test %s <slug>

The workshop collects purpose, explicit triggers, non-triggers, inputs, outputs, examples, required facts, success, and failure behavior. It shows every generated source byte, the complete source diff, unresolved judgments, and the installed-state effect before requesting exact source-only confirmation.

Never run install, upgrade, enable, disable, remove, or lifecycle recover, and never enter Create source, Update source, Install, Upgrade, Enable, Disable, Remove, Recover, or Rollback as a confirmation token. Launching the workshop does not authorize its source write. The user must answer, review, and authorize source and lifecycle mutations independently. Instruction-only validation is structural and does not certify that natural-language instructions are benign.
`

const builderPolicy = "policy:\n  allow_implicit_invocation: true\n"

var namePattern = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,30}[a-z0-9])?$`)
var reserved = map[string]bool{"codex": true, "my-friday": true, "default": true, "current": true, "new": true}

type Manifest struct {
	ContractVersion            int      `json:"contract_version"`
	Name                       string   `json:"name"`
	Root                       string   `json:"root"`
	Owned                      []string `json:"owned"`
	Launcher                   string   `json:"launcher"`
	LauncherSHA256             string   `json:"launcher_sha256"`
	CodexExecutable            string   `json:"codex_executable"`
	CodexSHA256                string   `json:"codex_sha256"`
	CodexConfig                string   `json:"codex_config"`
	CodexConfigSHA256          string   `json:"codex_config_sha256"`
	CodexInstructions          string   `json:"codex_instructions,omitempty"`
	CodexInstructionsSHA256    string   `json:"codex_instructions_sha256,omitempty"`
	MyFridayExecutable         string   `json:"my_friday_executable,omitempty"`
	MyFridaySHA256             string   `json:"my_friday_sha256,omitempty"`
	AssistantID                string   `json:"assistant_id,omitempty"`
	CapabilityRevision         int      `json:"capability_revision,omitempty"`
	RollbackContractVersion    int      `json:"rollback_contract_version,omitempty"`
	RollbackCapabilityRevision int      `json:"rollback_capability_revision,omitempty"`
	CapabilityBuilder          string   `json:"capability_builder"`
	CapabilityBuilderSHA256    string   `json:"capability_builder_sha256"`
	CapabilityPolicySHA256     string   `json:"capability_policy_sha256"`
}
type capabilityMigration struct {
	ContractVersion              int    `json:"contract_version"`
	Action                       string `json:"action"`
	SourceContractVersion        int    `json:"source_contract_version,omitempty"`
	SourceCapabilityRevision     int    `json:"source_capability_revision,omitempty"`
	TargetContractVersion        int    `json:"target_contract_version,omitempty"`
	TargetCapabilityRevision     int    `json:"target_capability_revision,omitempty"`
	SourceManifestSHA256         string `json:"source_manifest_sha256,omitempty"`
	CandidateSHA256              string `json:"candidate_sha256,omitempty"`
	BuilderQuarantineSHA256      string `json:"builder_quarantine_sha256,omitempty"`
	CapabilitiesQuarantineSHA256 string `json:"capabilities_quarantine_sha256,omitempty"`
	MyFridayQuarantineSHA256     string `json:"my_friday_quarantine_sha256,omitempty"`
}

func capabilityMigrationPath(root string) string {
	return filepath.Join(root, "capability-migration.json")
}
func writeCapabilityMigration(root string, migration capabilityMigration, replace bool) error {
	body, _ := json.MarshalIndent(migration, "", "  ")
	body = append(body, '\n')
	path := capabilityMigrationPath(root)
	if replace {
		path += ".new"
		if _, err := os.Lstat(path); !errors.Is(err, os.ErrNotExist) {
			return errors.New("assistant migration journal staging collision")
		}
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	_, werr := f.Write(body)
	cerr := f.Close()
	if werr != nil {
		return werr
	}
	if cerr != nil {
		return cerr
	}
	if replace {
		return os.Rename(path, capabilityMigrationPath(root))
	}
	return nil
}
func readCapabilityMigration(root string) (capabilityMigration, error) {
	body, info, err := regular(capabilityMigrationPath(root))
	if err != nil {
		return capabilityMigration{}, err
	}
	if info.Mode().Perm() != 0o600 {
		return capabilityMigration{}, errors.New("unsafe assistant migration journal")
	}
	var m capabilityMigration
	d := json.NewDecoder(bytes.NewReader(body))
	d.DisallowUnknownFields()
	if err = d.Decode(&m); err != nil {
		return m, err
	}
	canonical, _ := json.MarshalIndent(m, "", "  ")
	canonical = append(canonical, '\n')
	validTarget := m.Action == "upgrade" && m.TargetContractVersion == 0 && m.TargetCapabilityRevision == 0
	if m.Action == "rollback" {
		validTarget = (m.TargetContractVersion == 1 && m.TargetCapabilityRevision == 0) || (m.TargetContractVersion == ContractVersion && m.TargetCapabilityRevision == 0)
	}
	validQuarantines := m.BuilderQuarantineSHA256 == "" && m.CapabilitiesQuarantineSHA256 == "" && m.MyFridayQuarantineSHA256 == ""
	if m.BuilderQuarantineSHA256 != "" || m.CapabilitiesQuarantineSHA256 != "" || m.MyFridayQuarantineSHA256 != "" {
		validQuarantines = validDigest(m.BuilderQuarantineSHA256) && validDigest(m.CapabilitiesQuarantineSHA256) && validDigest(m.MyFridayQuarantineSHA256)
	}
	if !bytes.Equal(body, canonical) || m.ContractVersion != 2 || (m.Action != "upgrade" && m.Action != "rollback") || !validDigest(m.SourceManifestSHA256) || !validTarget || !validQuarantines {
		return m, errors.New("invalid assistant migration journal")
	}
	return m, nil
}

type Paths struct {
	Home, Name, Root, Launcher string
}

type rootProof struct{ Device, Inode uint64 }

type Plan struct {
	Action          string
	Paths           Paths
	Items           []string
	RuntimeSource   string
	MemorySource    string
	AssistantID     string
	candidatePath   string
	candidateSHA256 string
	rootDevice      uint64
	rootInode       uint64
}

func (p Plan) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Action: %s\nInstance: %s\nRoot: %s\n", p.Action, p.Paths.Name, p.Paths.Root)
	for _, item := range p.Items {
		fmt.Fprintf(&b, "- %s\n", item)
	}
	return b.String()
}

func ValidateName(name string) error {
	if !namePattern.MatchString(name) || reserved[name] {
		return fmt.Errorf("invalid canonical assistant name %q", name)
	}
	return nil
}

func Derive(home, name string) (Paths, error) {
	if err := ValidateName(name); err != nil {
		return Paths{}, err
	}
	abs, err := filepath.Abs(home)
	if err != nil {
		return Paths{}, err
	}
	if filepath.Clean(abs) != abs {
		return Paths{}, errors.New("home must be canonical")
	}
	return Paths{Home: abs, Name: name, Root: filepath.Join(abs, ".my-friday", "assistants", name), Launcher: filepath.Join(abs, ".local", "bin", name)}, nil
}

func digest(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }

func ownedTreeDigest(root string) (string, error) {
	var records []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		info, err := os.Lstat(path)
		if err != nil {
			return err
		}
		st, ok := info.Sys().(*syscall.Stat_t)
		if !ok || int(st.Uid) != os.Getuid() || info.Mode()&os.ModeSymlink != 0 {
			return errors.New("unsafe rollback quarantine entry")
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			if info.Mode().Perm() != 0o700 {
				return errors.New("unsafe rollback quarantine directory mode")
			}
			records = append(records, rel+":dir:0700")
			return nil
		}
		if !info.Mode().IsRegular() || info.Mode().Perm() != 0o600 || st.Nlink != 1 {
			return errors.New("unsafe rollback quarantine file")
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		records = append(records, rel+":file:0600:"+digest(body))
		return nil
	})
	if err != nil {
		return "", err
	}
	sort.Strings(records)
	return digest([]byte(strings.Join(records, "\n") + "\n")), nil
}

func ownedExecutableDigest(path string) (string, error) {
	body, info, err := regular(path)
	if err != nil || info.Mode().Perm() != 0o700 {
		return "", errors.New("unsafe rollback executable quarantine")
	}
	return digest(body), nil
}

func tomlBasicString(value string) (string, error) {
	if !utf8.ValidString(value) {
		return "", errors.New("path is not valid UTF-8")
	}
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range value {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\b':
			b.WriteString(`\b`)
		case '\t':
			b.WriteString(`\t`)
		case '\n':
			b.WriteString(`\n`)
		case '\f':
			b.WriteString(`\f`)
		case '\r':
			b.WriteString(`\r`)
		default:
			if r < 0x20 || r == 0x7f {
				fmt.Fprintf(&b, `\u%04X`, r)
			} else {
				b.WriteRune(r)
			}
		}
	}
	b.WriteByte('"')
	return b.String(), nil
}

func managedCodexConfig(p Paths, capabilityRevision int) ([]byte, error) {
	workspace, err := tomlBasicString(filepath.Join(p.Root, "workspace"))
	if err != nil {
		return nil, err
	}
	if capabilityRevision == 0 {
		return []byte("[projects." + workspace + "]\ntrust_level = \"trusted\"\n"), nil
	}
	runtime, err := tomlBasicString(filepath.Join(p.Root, "runtime"))
	if err != nil {
		return nil, err
	}
	return []byte("approval_policy = \"never\"\nsandbox_mode = \"workspace-write\"\n\n[sandbox_workspace_write]\nnetwork_access = false\nwritable_roots = [" + runtime + "]\n\n[projects." + workspace + "]\ntrust_level = \"trusted\"\n"), nil
}

func capabilityBuilder(p Paths) []byte {
	runtime := filepath.Join(p.Root, "runtime")
	executable := filepath.Join(p.Root, "dependencies", "my-friday")
	command := shellSingleQuote(executable)
	return []byte(fmt.Sprintf(builderSkillTemplate, p.Name, runtime, executable, command, p.Name, command, p.Name, command, p.Name, command, p.Name))
}

func shellSingleQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

func managedCodexInstructions(runtimeRoot, assistantID string) ([]byte, error) {
	id, err := repository.ValidateRuntime(runtimeRoot)
	if err != nil {
		return nil, fmt.Errorf("runtime repository: %w", err)
	}
	if id != assistantID {
		return nil, errors.New("runtime assistant identity mismatch")
	}
	body, _, err := regular(filepath.Join(runtimeRoot, "assistant", "profile.json"))
	if err != nil {
		return nil, fmt.Errorf("runtime profile: %w", err)
	}
	var p profile.Profile
	if json.Unmarshal(body, &p) != nil || profile.Validate(p) != nil || p.AssistantID != id {
		return nil, errors.New("runtime profile is incompatible")
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
	return []byte(out.String()), nil
}

func regular(path string) ([]byte, fs.FileInfo, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, nil, err
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 || info.Sys().(*syscall.Stat_t).Nlink != 1 {
		return nil, nil, fmt.Errorf("unsafe non-regular file %s", path)
	}
	b, err := os.ReadFile(path)
	return b, info, err
}

func safeDir(path string, mode fs.FileMode) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 || int(st.Uid) != os.Getuid() || info.Mode().Perm() != mode {
		return fmt.Errorf("unsafe directory %s", path)
	}
	return nil
}

func safeOwnedDir(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 || int(st.Uid) != os.Getuid() {
		return fmt.Errorf("unsafe directory %s", path)
	}
	return nil
}

func safeLauncherDirectory(home string) error {
	if err := safeOwnedDir(home); err != nil {
		return err
	}
	local := filepath.Join(home, ".local")
	if err := safeOwnedDir(local); err != nil {
		return err
	}
	return safeDir(filepath.Join(local, "bin"), 0755)
}

func prepareAssistantsRoot(home string) error {
	base := filepath.Join(home, ".my-friday")
	assistants := filepath.Join(base, "assistants")
	for _, path := range []string{base, assistants} {
		if err := os.Mkdir(path, 0700); err != nil && !errors.Is(err, os.ErrExist) {
			return err
		}
		if err := safeDir(path, 0700); err != nil {
			return err
		}
	}
	return nil
}

type transactionLock struct {
	file  *os.File
	proof rootProof
}

func acquireTransactionLock(root string, expected *rootProof) (*transactionLock, error) {
	fd, err := syscall.Open(root, syscall.O_RDONLY|syscall.O_DIRECTORY|syscall.O_CLOEXEC|syscall.O_NOFOLLOW, 0)
	if err != nil {
		return nil, err
	}
	f := os.NewFile(uintptr(fd), root)
	var lockStat syscall.Stat_t
	if err = syscall.Fstat(fd, &lockStat); err != nil || lockStat.Mode&syscall.S_IFMT != syscall.S_IFDIR || int(lockStat.Uid) != os.Getuid() || lockStat.Mode&0777 != 0700 {
		f.Close()
		return nil, fmt.Errorf("unsafe instance transaction root %s", root)
	}
	proof := rootProof{Device: uint64(lockStat.Dev), Inode: uint64(lockStat.Ino)}
	if expected != nil && proof != *expected {
		f.Close()
		return nil, errors.New("instance root identity changed before transaction lock")
	}
	if err = matchOpenedRoot(root, proof); err != nil {
		f.Close()
		return nil, err
	}
	if err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		f.Close()
		return nil, fmt.Errorf("instance transaction already active: %w", err)
	}
	if err = matchOpenedRoot(root, proof); err != nil {
		syscall.Flock(fd, syscall.LOCK_UN)
		f.Close()
		return nil, err
	}
	return &transactionLock{file: f, proof: proof}, nil
}

func matchOpenedRoot(path string, proof rootProof) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 || uint64(st.Dev) != proof.Device || uint64(st.Ino) != proof.Inode {
		return errors.New("instance root path changed during transaction lock")
	}
	return nil
}

func (l *transactionLock) release() error {
	if l == nil {
		return nil
	}
	unlockErr := syscall.Flock(int(l.file.Fd()), syscall.LOCK_UN)
	closeErr := l.file.Close()
	if unlockErr != nil {
		return unlockErr
	}
	return closeErr
}

func finishTransaction(lock *transactionLock, operationErr error) error {
	releaseErr := lock.release()
	if operationErr != nil {
		return operationErr
	}
	if releaseErr != nil {
		return fmt.Errorf("transaction cleanup failed: %w", releaseErr)
	}
	return nil
}

func PlanCreate(home, name, executable, codex string) (Plan, error) {
	p, err := Derive(home, name)
	if err != nil {
		return Plan{}, err
	}
	if err = safeLauncherDirectory(p.Home); err != nil {
		return Plan{}, fmt.Errorf("launcher directory refused: %w", err)
	}
	if _, err = os.Lstat(p.Root); !errors.Is(err, os.ErrNotExist) {
		return Plan{}, fmt.Errorf("instance collision at %s", p.Root)
	}
	if _, err = os.Lstat(p.Launcher); !errors.Is(err, os.ErrNotExist) {
		return Plan{}, fmt.Errorf("launcher collision at %s", p.Launcher)
	}
	if _, _, err = regular(executable); err != nil {
		return Plan{}, fmt.Errorf("launcher source refused: %w", err)
	}
	if codex == "" {
		return Plan{}, errors.New("Codex executable is required")
	}
	if _, _, err = regular(codex); err != nil {
		return Plan{}, fmt.Errorf("Codex executable refused: %w", err)
	}
	return Plan{Action: "create", Paths: p, Items: []string{"create private instance root", "create codex, runtime, memory, workspace, and dependencies", "trust the exact workspace and add only the private runtime as workspace-write authority with approvals never and network disabled", "copy and bind the exact My Friday executable for capability-builder read-only checks", "install exact native launcher with no replacement", "leave HOME, shell files, user Codex state, and launcher siblings unchanged"}}, nil
}

func WithRepositories(plan Plan, runtime, memory, assistantID string) (Plan, error) {
	for _, source := range []string{runtime, memory} {
		abs, err := filepath.Abs(source)
		if err != nil {
			return plan, err
		}
		if abs == plan.Paths.Root || strings.HasPrefix(abs, plan.Paths.Root+string(filepath.Separator)) {
			return plan, errors.New("repository source cannot be inside destination instance")
		}
	}
	plan.RuntimeSource, plan.MemorySource, plan.AssistantID = runtime, memory, assistantID
	return plan, nil
}

func copyTree(source, destination string) error {
	return filepath.WalkDir(source, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() && !info.Mode().IsRegular() {
			return fmt.Errorf("repository source contains unsafe entry %s", path)
		}
		target := filepath.Join(destination, rel)
		if info.IsDir() {
			if rel == "." {
				return nil
			}
			return os.Mkdir(target, 0700)
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		mode := fs.FileMode(0600)
		if info.Mode().Perm()&0100 != 0 {
			mode = 0700
		}
		copyErr := CopyFile(target, in, mode)
		closeErr := in.Close()
		if copyErr != nil {
			return copyErr
		}
		return closeErr
	})
}

func Create(plan Plan, executable, codex string) error {
	return createLocked(plan, executable, codex, nil)
}

var mutationHook func(string)
var upgradeHook func(string) error
var rollbackHook func(string) error

func createLocked(plan Plan, executable, codex string, afterVerify func() error) (resultErr error) {
	if plan.Action != "create" {
		return errors.New("invalid create plan")
	}
	p, err := Derive(plan.Paths.Home, plan.Paths.Name)
	if err != nil || p != plan.Paths {
		return errors.New("create plan path mismatch")
	}
	hasRepositoryPlan := plan.RuntimeSource != "" || plan.MemorySource != "" || plan.AssistantID != ""
	if hasRepositoryPlan && (plan.RuntimeSource == "" || plan.MemorySource == "" || plan.AssistantID == "") {
		return errors.New("incomplete create repository plan")
	}
	if _, err = os.Lstat(p.Root); !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("instance collision at %s", p.Root)
	}
	if _, err = os.Lstat(p.Launcher); !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("launcher collision at %s", p.Launcher)
	}
	if err = safeLauncherDirectory(p.Home); err != nil {
		return err
	}
	if err = prepareAssistantsRoot(p.Home); err != nil {
		return err
	}
	launcher, _, err := regular(executable)
	if err != nil {
		return err
	}
	if _, _, err = regular(codex); err != nil {
		return err
	}
	stage := p.Root + ".creating"
	if err = os.Mkdir(stage, 0700); err != nil {
		if errors.Is(err, os.ErrExist) {
			return fmt.Errorf("instance transaction already active or recovery required at %s", stage)
		}
		return err
	}
	clean := func() { _ = os.RemoveAll(stage) }
	lock, err := acquireTransactionLock(stage, nil)
	if err != nil {
		clean()
		return err
	}
	if mutationHook != nil {
		mutationHook("create-preflight")
	}
	defer func() {
		resultErr = finishTransaction(lock, resultErr)
	}()
	for _, child := range []string{"capabilities", "codex", "runtime", "memory", "workspace", "dependencies"} {
		if err = os.Mkdir(filepath.Join(stage, child), 0700); err != nil {
			clean()
			return err
		}
	}
	owned := []string{"capabilities", "codex", "dependencies", "memory", "runtime", "workspace"}
	sort.Strings(owned)
	managedCodex := filepath.Join(p.Root, "dependencies", "codex")
	managedMyFriday := filepath.Join(p.Root, "dependencies", "my-friday")
	managedConfig := filepath.Join(p.Root, "codex", "config.toml")
	codexBytes, _, err := regular(codex)
	if err != nil {
		clean()
		return err
	}
	configBytes, err := managedCodexConfig(p, CapabilityRevision)
	if err != nil {
		clean()
		return err
	}
	var instructionsBytes []byte
	managedInstructions := ""
	if plan.RuntimeSource != "" {
		if err = copyTree(plan.RuntimeSource, filepath.Join(stage, "runtime")); err != nil {
			clean()
			return err
		}
		if err = copyTree(plan.MemorySource, filepath.Join(stage, "memory")); err != nil {
			clean()
			return err
		}
		instructionsBytes, err = managedCodexInstructions(filepath.Join(stage, "runtime"), plan.AssistantID)
		if err != nil {
			clean()
			return err
		}
		managedInstructions = filepath.Join(p.Root, "codex", "AGENTS.md")
		if err = os.WriteFile(filepath.Join(stage, "codex", "AGENTS.md"), instructionsBytes, 0600); err != nil {
			clean()
			return err
		}
	}
	builder := filepath.Join(stage, "workspace", ".agents", "skills", "capability-builder")
	builderBytes := capabilityBuilder(p)
	if err = os.MkdirAll(filepath.Join(builder, "agents"), 0700); err != nil {
		clean()
		return err
	}
	if err = os.WriteFile(filepath.Join(builder, "SKILL.md"), builderBytes, 0600); err != nil {
		clean()
		return err
	}
	policyBytes := []byte(builderPolicy)
	if err = os.WriteFile(filepath.Join(builder, "agents", "openai.yaml"), policyBytes, 0600); err != nil {
		clean()
		return err
	}
	m := Manifest{ContractVersion: ContractVersion, Name: p.Name, Root: p.Root, Owned: owned, Launcher: p.Launcher, LauncherSHA256: digest(launcher), CodexExecutable: managedCodex, CodexSHA256: digest(codexBytes), CodexConfig: managedConfig, CodexConfigSHA256: digest(configBytes), CodexInstructions: managedInstructions, AssistantID: plan.AssistantID, MyFridayExecutable: managedMyFriday, MyFridaySHA256: digest(launcher), CapabilityRevision: CapabilityRevision, RollbackContractVersion: 1, CapabilityBuilder: filepath.Join(p.Root, "workspace", ".agents", "skills", "capability-builder"), CapabilityBuilderSHA256: digest(builderBytes), CapabilityPolicySHA256: digest(policyBytes)}
	if managedInstructions != "" {
		m.CodexInstructionsSHA256 = digest(instructionsBytes)
	}
	mb, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		clean()
		return err
	}
	mb = append(mb, '\n')
	if err = os.WriteFile(filepath.Join(stage, "manifest.json"), mb, 0600); err != nil {
		clean()
		return err
	}
	if err = os.WriteFile(filepath.Join(stage, "dependencies", "codex"), codexBytes, 0700); err != nil {
		clean()
		return err
	}
	if err = os.WriteFile(filepath.Join(stage, "dependencies", "my-friday"), launcher, 0700); err != nil {
		clean()
		return err
	}
	if err = os.WriteFile(filepath.Join(stage, "codex", "config.toml"), configBytes, 0600); err != nil {
		clean()
		return err
	}
	if err = os.Rename(stage, p.Root); err != nil {
		clean()
		return err
	}
	tmp := filepath.Join(filepath.Dir(p.Launcher), "."+p.Name+".my-friday-new")
	f, openErr := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0700)
	if openErr != nil {
		_ = removeOwnedRoot(p, m)
		return openErr
	}
	if err = f.Close(); err != nil {
		_ = os.Remove(tmp)
		_ = removeOwnedRoot(p, m)
		return err
	}
	if err = os.WriteFile(tmp, launcher, 0700); err != nil {
		_ = os.Remove(tmp)
		_ = removeOwnedRoot(p, m)
		return err
	}
	if err = os.Link(tmp, p.Launcher); err != nil {
		_ = os.Remove(tmp)
		_ = removeOwnedRoot(p, m)
		return fmt.Errorf("launcher no-replace promotion failed: %w", err)
	}
	_ = os.Remove(tmp)
	if _, err = verify(p.Home, p.Name); err != nil {
		return fmt.Errorf("recovery required: create verification failed: %w", err)
	}
	if afterVerify != nil {
		if err = afterVerify(); err != nil {
			return fmt.Errorf("named instance active; prior projection cleanup recovery required: %w", err)
		}
	}
	return nil
}

func readManifest(p Paths) (Manifest, error) {
	b, _, err := regular(filepath.Join(p.Root, "manifest.json"))
	if err != nil {
		return Manifest{}, err
	}
	var m Manifest
	if err = json.Unmarshal(b, &m); err != nil {
		return m, err
	}
	want := []string{"capabilities", "codex", "dependencies", "memory", "runtime", "workspace"}
	if m.ContractVersion == 1 {
		want = []string{"codex", "dependencies", "memory", "runtime", "workspace"}
	}
	wantCodex := filepath.Join(p.Root, "dependencies", "codex")
	wantMyFriday := ""
	validMyFridayDigest := m.MyFridaySHA256 == ""
	if m.ContractVersion == ContractVersion && m.CapabilityRevision == CapabilityRevision {
		wantMyFriday = filepath.Join(p.Root, "dependencies", "my-friday")
		validMyFridayDigest = validDigest(m.MyFridaySHA256)
	}
	wantConfig := filepath.Join(p.Root, "codex", "config.toml")
	wantInstructions := ""
	if m.AssistantID != "" {
		wantInstructions = filepath.Join(p.Root, "codex", "AGENTS.md")
	}
	configBytes, configErr := managedCodexConfig(p, m.CapabilityRevision)
	validInstructionsDigest := m.CodexInstructionsSHA256 == ""
	if wantInstructions != "" {
		validInstructionsDigest = validDigest(m.CodexInstructionsSHA256)
	}
	wantBuilder := filepath.Join(p.Root, "workspace", ".agents", "skills", "capability-builder")
	validBuilder := m.ContractVersion == 1 && m.CapabilityRevision == 0 && m.CapabilityBuilder == "" && m.CapabilityBuilderSHA256 == "" && m.CapabilityPolicySHA256 == ""
	validRollback := m.ContractVersion == 1 && m.RollbackContractVersion == 0 && m.RollbackCapabilityRevision == 0
	if m.ContractVersion == ContractVersion && m.CapabilityRevision == 0 {
		validBuilder = m.CapabilityBuilder == wantBuilder && m.CapabilityBuilderSHA256 == digest([]byte(legacyBuilderSkill)) && m.CapabilityPolicySHA256 == digest([]byte(builderPolicy)) && m.MyFridayExecutable == "" && m.MyFridaySHA256 == ""
		validRollback = m.RollbackContractVersion == 0 && m.RollbackCapabilityRevision == 0
	}
	if m.ContractVersion == ContractVersion && m.CapabilityRevision == CapabilityRevision {
		validBuilder = m.CapabilityBuilder == wantBuilder && m.CapabilityBuilderSHA256 == digest(capabilityBuilder(p)) && m.CapabilityPolicySHA256 == digest([]byte(builderPolicy))
		validRollback = (m.RollbackContractVersion == 1 && m.RollbackCapabilityRevision == 0) || (m.RollbackContractVersion == ContractVersion && m.RollbackCapabilityRevision == 0)
	}
	validRevision := (m.ContractVersion == 1 && m.CapabilityRevision == 0) || (m.ContractVersion == ContractVersion && (m.CapabilityRevision == 0 || m.CapabilityRevision == CapabilityRevision))
	if configErr != nil || !validRevision || !validRollback || m.Name != p.Name || m.Root != p.Root || m.Launcher != p.Launcher || m.CodexExecutable != wantCodex || !validDigest(m.CodexSHA256) || m.MyFridayExecutable != wantMyFriday || !validMyFridayDigest || m.CodexConfig != wantConfig || m.CodexConfigSHA256 != digest(configBytes) || m.CodexInstructions != wantInstructions || !validInstructionsDigest || !validBuilder || strings.Join(m.Owned, "\x00") != strings.Join(want, "\x00") {
		return m, errors.New("manifest ownership contract mismatch")
	}
	return m, nil
}

func verifyManagedConfig(p Paths, m Manifest) error {
	b, info, err := regular(m.CodexConfig)
	if err != nil {
		return fmt.Errorf("managed Codex config unavailable: %w", err)
	}
	expected, err := managedCodexConfig(p, m.CapabilityRevision)
	if err != nil {
		return err
	}
	if info.Mode().Perm() != 0600 || digest(b) != m.CodexConfigSHA256 || !bytes.Equal(b, expected) {
		return errors.New("managed Codex config drift")
	}
	return nil
}

func verifyManagedInstructions(p Paths, m Manifest) error {
	if m.CodexInstructions == "" {
		return nil
	}
	body, info, err := regular(m.CodexInstructions)
	if err != nil {
		return fmt.Errorf("managed Codex instructions unavailable: %w", err)
	}
	expected, err := managedCodexInstructions(filepath.Join(p.Root, "runtime"), m.AssistantID)
	if err != nil {
		return fmt.Errorf("managed Codex instructions source invalid: %w", err)
	}
	if info.Mode().Perm() != 0600 || digest(body) != m.CodexInstructionsSHA256 || !bytes.Equal(body, expected) {
		return errors.New("managed Codex instructions drift")
	}
	return nil
}

func verifyCapabilityBuilder(p Paths, m Manifest) error {
	expectedSkill := []byte(legacyBuilderSkill)
	if m.CapabilityRevision == CapabilityRevision {
		expectedSkill = capabilityBuilder(p)
	}
	skill, info, err := regular(filepath.Join(m.CapabilityBuilder, "SKILL.md"))
	if err != nil || info.Mode().Perm() != 0600 || digest(skill) != m.CapabilityBuilderSHA256 || !bytes.Equal(skill, expectedSkill) {
		return errors.New("capability builder drift")
	}
	policy, info, err := regular(filepath.Join(m.CapabilityBuilder, "agents", "openai.yaml"))
	if err != nil || info.Mode().Perm() != 0600 || digest(policy) != m.CapabilityPolicySHA256 || !bytes.Equal(policy, []byte(builderPolicy)) {
		return errors.New("capability builder policy drift")
	}
	return nil
}

func Verify(home, name string) (Manifest, error) { return verify(home, name) }

func PlanUpgrade(home, name, executable string) (Plan, error) {
	p, err := Derive(home, name)
	if err != nil {
		return Plan{}, err
	}
	m, err := Verify(home, name)
	if err != nil {
		return Plan{}, fmt.Errorf("upgrade refused: %w", err)
	}
	if m.ContractVersion == ContractVersion && m.CapabilityRevision == CapabilityRevision {
		return Plan{}, errors.New("assistant already uses capability contract v2")
	}
	candidate, candidateInfo, candidateErr := regular(executable)
	if candidateErr != nil || candidateInfo.Mode().Perm()&0o111 == 0 {
		return Plan{}, errors.New("assistant upgrade candidate executable refused")
	}
	if _, err = os.Lstat(capabilityMigrationPath(p.Root)); err == nil {
		return Plan{}, errors.New("assistant capability migration recovery required")
	} else if !errors.Is(err, os.ErrNotExist) {
		return Plan{}, err
	}
	info, err := os.Lstat(p.Root)
	if err != nil {
		return Plan{}, err
	}
	st := info.Sys().(*syscall.Stat_t)
	return Plan{Action: "upgrade", Paths: p, Items: []string{"add manifest-owned capability control root", "project the instance-specific capability builder into this instance workspace", "copy and bind the exact currently executing My Friday candidate for read-only builder checks", "add only the private runtime to workspace-write authority with approvals never and network disabled", "leave external runtime source, global skills, Codex credentials, and other instances unchanged"}, candidatePath: executable, candidateSHA256: digest(candidate), rootDevice: uint64(st.Dev), rootInode: uint64(st.Ino)}, nil
}

func Upgrade(plan Plan) error {
	if plan.Action != "upgrade" {
		return errors.New("invalid assistant upgrade plan")
	}
	expected := rootProof{Device: plan.rootDevice, Inode: plan.rootInode}
	lock, err := acquireTransactionLock(plan.Paths.Root, &expected)
	if err != nil {
		return err
	}
	return finishTransaction(lock, upgradeLocked(plan))
}

func PlanRollback(home, name string) (Plan, error) {
	p, err := Derive(home, name)
	if err != nil {
		return Plan{}, err
	}
	m, err := Verify(home, name)
	if err != nil {
		return Plan{}, fmt.Errorf("rollback refused: %w", err)
	}
	if m.ContractVersion != ContractVersion || m.CapabilityRevision != CapabilityRevision {
		return Plan{}, errors.New("assistant rollback requires capability contract v2")
	}
	entries, err := os.ReadDir(filepath.Join(p.Root, "capabilities"))
	if err != nil {
		return Plan{}, err
	}
	if len(entries) != 0 {
		return Plan{}, errors.New("assistant rollback refused: capability control state exists")
	}
	skills := filepath.Join(p.Root, "workspace", ".agents", "skills")
	entries, err = os.ReadDir(skills)
	if err != nil {
		return Plan{}, err
	}
	if len(entries) != 1 || entries[0].Name() != "capability-builder" {
		return Plan{}, errors.New("assistant rollback refused: workspace skill entries differ")
	}
	info, err := os.Lstat(p.Root)
	if err != nil {
		return Plan{}, err
	}
	st := info.Sys().(*syscall.Stat_t)
	target := "v1"
	if m.RollbackContractVersion == ContractVersion && m.RollbackCapabilityRevision == 0 {
		target = "legacy capability contract v2"
	}
	return Plan{Action: "rollback", Paths: p, Items: []string{"quarantine the exact manifest-owned capability builder, builder executable, and empty control root", "restore the " + target + " private Codex config and instance manifest", "delete quarantines only after journal-bound type, mode, and digest revalidation", "leave runtime source, Codex credentials, launcher, and other instances unchanged"}, rootDevice: uint64(st.Dev), rootInode: uint64(st.Ino)}, nil
}

func Rollback(plan Plan) error {
	if plan.Action != "rollback" {
		return errors.New("invalid assistant rollback plan")
	}
	expected := rootProof{Device: plan.rootDevice, Inode: plan.rootInode}
	lock, err := acquireTransactionLock(plan.Paths.Root, &expected)
	if err != nil {
		return err
	}
	return finishTransaction(lock, rollbackLocked(plan))
}
func rollbackLocked(plan Plan) error {
	m, err := verify(plan.Paths.Home, plan.Paths.Name)
	if err != nil {
		return err
	}
	if m.ContractVersion != ContractVersion || m.CapabilityRevision != CapabilityRevision {
		return errors.New("assistant rollback requires contract v2")
	}
	for _, quarantine := range []string{m.CapabilityBuilder + ".rollback", filepath.Join(plan.Paths.Root, "capabilities.rollback"), m.MyFridayExecutable + ".rollback", capabilityMigrationPath(plan.Paths.Root) + ".new"} {
		if _, collisionErr := os.Lstat(quarantine); !errors.Is(collisionErr, os.ErrNotExist) {
			return errors.New("assistant rollback quarantine collision")
		}
	}
	entries, err := os.ReadDir(filepath.Join(plan.Paths.Root, "capabilities"))
	if err != nil || len(entries) != 0 {
		return errors.New("assistant rollback refused: capability control state exists")
	}
	skills := filepath.Dir(m.CapabilityBuilder)
	entries, err = os.ReadDir(skills)
	if err != nil || len(entries) != 1 || entries[0].Name() != "capability-builder" {
		return errors.New("assistant rollback refused: workspace skill entries changed")
	}
	manifestBytes, _, manifestErr := regular(filepath.Join(plan.Paths.Root, "manifest.json"))
	if manifestErr != nil {
		return manifestErr
	}
	migration := capabilityMigration{ContractVersion: 2, Action: "rollback", SourceContractVersion: m.ContractVersion, SourceCapabilityRevision: m.CapabilityRevision, TargetContractVersion: m.RollbackContractVersion, TargetCapabilityRevision: m.RollbackCapabilityRevision, SourceManifestSHA256: digest(manifestBytes)}
	if err = writeCapabilityMigration(plan.Paths.Root, migration, false); err != nil {
		return err
	}
	return completeRollback(plan.Paths, m)
}

func validateRollbackQuarantines(p Paths, migration capabilityMigration) error {
	builderDigest, err := ownedTreeDigest(filepath.Join(p.Root, "workspace", ".agents", "skills", "capability-builder.rollback"))
	if err != nil || builderDigest != migration.BuilderQuarantineSHA256 {
		return errors.New("assistant rollback builder quarantine drift")
	}
	capabilities := filepath.Join(p.Root, "capabilities.rollback")
	entries, err := os.ReadDir(capabilities)
	if err != nil || len(entries) != 0 {
		return errors.New("assistant rollback capability quarantine drift")
	}
	capabilitiesDigest, err := ownedTreeDigest(capabilities)
	if err != nil || capabilitiesDigest != migration.CapabilitiesQuarantineSHA256 {
		return errors.New("assistant rollback capability quarantine drift")
	}
	executableDigest, err := ownedExecutableDigest(filepath.Join(p.Root, "dependencies", "my-friday.rollback"))
	if err != nil || executableDigest != migration.MyFridayQuarantineSHA256 {
		return errors.New("assistant rollback executable quarantine drift")
	}
	return nil
}

func finishRollbackQuarantines(p Paths, migration capabilityMigration) error {
	if err := validateRollbackQuarantines(p, migration); err != nil {
		return err
	}
	if err := os.RemoveAll(filepath.Join(p.Root, "workspace", ".agents", "skills", "capability-builder.rollback")); err != nil {
		return err
	}
	if err := os.Remove(filepath.Join(p.Root, "capabilities.rollback")); err != nil {
		return err
	}
	if err := os.Remove(filepath.Join(p.Root, "dependencies", "my-friday.rollback")); err != nil {
		return err
	}
	return os.Remove(capabilityMigrationPath(p.Root))
}

func completeRollback(p Paths, m Manifest) error {
	migration, err := readCapabilityMigration(p.Root)
	if err != nil || migration.Action != "rollback" || migration.SourceContractVersion != m.ContractVersion || migration.SourceCapabilityRevision != m.CapabilityRevision {
		return errors.New("assistant rollback journal invalid")
	}
	manifestBytes, _, manifestErr := regular(filepath.Join(p.Root, "manifest.json"))
	if manifestErr != nil || digest(manifestBytes) != migration.SourceManifestSHA256 {
		return errors.New("assistant rollback source manifest changed")
	}
	builder := m.CapabilityBuilder
	if builder == "" {
		builder = filepath.Join(p.Root, "workspace", ".agents", "skills", "capability-builder")
	}
	skills := filepath.Dir(builder)
	builderQ := builder + ".rollback"
	capabilities := filepath.Join(p.Root, "capabilities")
	capabilitiesQ := capabilities + ".rollback"
	managedMyFriday := filepath.Join(p.Root, "dependencies", "my-friday")
	managedMyFridayQ := managedMyFriday + ".rollback"
	if _, err := os.Lstat(builderQ); errors.Is(err, os.ErrNotExist) {
		entries, readErr := os.ReadDir(skills)
		if readErr != nil || len(entries) != 1 || entries[0].Name() != "capability-builder" {
			return errors.New("assistant rollback recovery refused: workspace skills changed")
		}
		if err = os.Rename(builder, builderQ); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	if _, err := os.Lstat(capabilitiesQ); errors.Is(err, os.ErrNotExist) {
		entries, readErr := os.ReadDir(capabilities)
		if readErr != nil || len(entries) != 0 {
			return errors.New("assistant rollback recovery refused: capability control changed")
		}
		if err = os.Rename(capabilities, capabilitiesQ); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	if _, err := os.Lstat(managedMyFridayQ); errors.Is(err, os.ErrNotExist) {
		if err = os.Rename(managedMyFriday, managedMyFridayQ); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	if migration.BuilderQuarantineSHA256 == "" {
		skill, info, skillErr := regular(filepath.Join(builderQ, "SKILL.md"))
		policy, policyInfo, policyErr := regular(filepath.Join(builderQ, "agents", "openai.yaml"))
		if skillErr != nil || policyErr != nil || info.Mode().Perm() != 0o600 || policyInfo.Mode().Perm() != 0o600 || digest(skill) != m.CapabilityBuilderSHA256 || digest(policy) != m.CapabilityPolicySHA256 || !bytes.Equal(skill, capabilityBuilder(p)) || !bytes.Equal(policy, []byte(builderPolicy)) {
			return errors.New("assistant rollback builder quarantine does not match source manifest")
		}
		migration.BuilderQuarantineSHA256, err = ownedTreeDigest(builderQ)
		if err != nil {
			return err
		}
		entries, err := os.ReadDir(capabilitiesQ)
		if err != nil || len(entries) != 0 {
			return errors.New("assistant rollback capability quarantine is not empty")
		}
		migration.CapabilitiesQuarantineSHA256, err = ownedTreeDigest(capabilitiesQ)
		if err != nil {
			return err
		}
		migration.MyFridayQuarantineSHA256, err = ownedExecutableDigest(managedMyFridayQ)
		if err != nil || migration.MyFridayQuarantineSHA256 != m.MyFridaySHA256 {
			return errors.New("assistant rollback executable quarantine does not match source manifest")
		}
		if err = writeCapabilityMigration(p.Root, migration, true); err != nil {
			return err
		}
	} else if err = validateRollbackQuarantines(p, migration); err != nil {
		return err
	}
	if rollbackHook != nil {
		if err := rollbackHook("quarantined"); err != nil {
			return err
		}
	}
	targetContract, targetRevision := m.RollbackContractVersion, m.RollbackCapabilityRevision
	m.ContractVersion = targetContract
	m.CapabilityRevision = targetRevision
	m.RollbackContractVersion = 0
	m.RollbackCapabilityRevision = 0
	m.MyFridayExecutable = ""
	m.MyFridaySHA256 = ""
	if targetContract == 1 {
		m.Owned = []string{"codex", "dependencies", "memory", "runtime", "workspace"}
		m.CapabilityBuilder = ""
		m.CapabilityBuilderSHA256 = ""
		m.CapabilityPolicySHA256 = ""
	} else if targetContract == ContractVersion && targetRevision == 0 {
		if err = os.MkdirAll(filepath.Join(builder, "agents"), 0o700); err != nil {
			return err
		}
		if err = os.WriteFile(filepath.Join(builder, "SKILL.md"), []byte(legacyBuilderSkill), 0o600); err != nil {
			return err
		}
		if err = os.WriteFile(filepath.Join(builder, "agents", "openai.yaml"), []byte(builderPolicy), 0o600); err != nil {
			return err
		}
		if err = os.Mkdir(capabilities, 0o700); err != nil {
			return err
		}
		m.CapabilityBuilder = builder
		m.CapabilityBuilderSHA256 = digest([]byte(legacyBuilderSkill))
		m.CapabilityPolicySHA256 = digest([]byte(builderPolicy))
	} else {
		return errors.New("assistant rollback target is unsupported")
	}
	configBytes, configErr := managedCodexConfig(p, 0)
	if configErr != nil {
		return configErr
	}
	m.CodexConfigSHA256 = digest(configBytes)
	if err = os.WriteFile(m.CodexConfig, configBytes, 0o600); err != nil {
		return err
	}
	body, _ := json.MarshalIndent(m, "", "  ")
	body = append(body, '\n')
	tmp := filepath.Join(p.Root, "manifest.json.rollback")
	if err = os.WriteFile(tmp, body, 0o600); err != nil {
		return err
	}
	if err = os.Rename(tmp, filepath.Join(p.Root, "manifest.json")); err != nil {
		return err
	}
	if err = finishRollbackQuarantines(p, migration); err != nil {
		return err
	}
	if targetContract == 1 {
		_ = os.Remove(skills)
		_ = os.Remove(filepath.Dir(skills))
	}
	_, err = verify(p.Home, p.Name)
	return err
}

func upgradeLocked(plan Plan) error {
	_, journalErr := os.Lstat(capabilityMigrationPath(plan.Paths.Root))
	resuming := journalErr == nil
	var m Manifest
	var migration capabilityMigration
	var err error
	if resuming {
		migration, err = readCapabilityMigration(plan.Paths.Root)
		if err != nil || migration.Action != "upgrade" || !validDigest(migration.CandidateSHA256) {
			return errors.New("assistant upgrade journal invalid")
		}
		m, err = readManifest(plan.Paths)
	} else {
		m, err = verify(plan.Paths.Home, plan.Paths.Name)
	}
	if err != nil {
		return fmt.Errorf("upgrade refused: %w", err)
	}
	if m.ContractVersion != 1 && !(m.ContractVersion == ContractVersion && m.CapabilityRevision == 0) {
		return errors.New("assistant upgrade requires contract v1 or legacy capability contract v2")
	}
	capabilities := filepath.Join(plan.Paths.Root, "capabilities")
	builder := filepath.Join(plan.Paths.Root, "workspace", ".agents", "skills", "capability-builder")
	managedMyFriday := filepath.Join(plan.Paths.Root, "dependencies", "my-friday")
	candidateStage := managedMyFriday + ".upgrade"
	if !resuming {
		if !errors.Is(journalErr, os.ErrNotExist) {
			return journalErr
		}
		if m.ContractVersion == 1 {
			if _, statErr := os.Lstat(capabilities); !errors.Is(statErr, os.ErrNotExist) {
				return errors.New("capability control collision")
			}
			if _, statErr := os.Lstat(builder); !errors.Is(statErr, os.ErrNotExist) {
				return errors.New("capability builder collision")
			}
		} else {
			entries, readErr := os.ReadDir(capabilities)
			if readErr != nil || len(entries) != 0 {
				return errors.New("legacy v2 capability control is not empty")
			}
		}
		if _, statErr := os.Lstat(managedMyFriday); !errors.Is(statErr, os.ErrNotExist) {
			return errors.New("managed My Friday executable collision")
		}
		if _, statErr := os.Lstat(candidateStage); !errors.Is(statErr, os.ErrNotExist) {
			return errors.New("managed My Friday upgrade staging collision")
		}
		candidate, info, candidateErr := regular(plan.candidatePath)
		if candidateErr != nil || info.Mode().Perm()&0o111 == 0 || digest(candidate) != plan.candidateSHA256 {
			return errors.New("assistant upgrade candidate changed after preview")
		}
		manifestBytes, _, manifestErr := regular(filepath.Join(plan.Paths.Root, "manifest.json"))
		if manifestErr != nil {
			return manifestErr
		}
		migration = capabilityMigration{ContractVersion: 2, Action: "upgrade", SourceContractVersion: m.ContractVersion, SourceCapabilityRevision: m.CapabilityRevision, SourceManifestSHA256: digest(manifestBytes), CandidateSHA256: plan.candidateSHA256}
		if err = CopyFile(candidateStage, bytes.NewReader(candidate), 0o700); err != nil {
			return err
		}
		if err = writeCapabilityMigration(plan.Paths.Root, migration, false); err != nil {
			_ = os.Remove(candidateStage)
			return err
		}
	} else {
		manifestBytes, _, manifestErr := regular(filepath.Join(plan.Paths.Root, "manifest.json"))
		if manifestErr != nil || digest(manifestBytes) != migration.SourceManifestSHA256 || m.ContractVersion != migration.SourceContractVersion || m.CapabilityRevision != migration.SourceCapabilityRevision {
			return errors.New("assistant upgrade source manifest changed")
		}
	}
	candidateBytes, candidateInfo, candidateErr := regular(candidateStage)
	if candidateErr != nil || candidateInfo.Mode().Perm() != 0o700 || digest(candidateBytes) != migration.CandidateSHA256 {
		return errors.New("assistant upgrade candidate staging drift")
	}
	runtimeRoot := filepath.Join(plan.Paths.Root, "runtime")
	if _, runtimeManifestErr := os.Lstat(filepath.Join(runtimeRoot, ".my-friday", "manifest.json")); runtimeManifestErr == nil {
		if _, statErr := os.Lstat(filepath.Join(runtimeRoot, ".my-friday", "capability-migration.json")); statErr == nil {
			if err = repository.RecoverCapabilities(runtimeRoot); err != nil {
				return err
			}
		} else if _, statErr = os.Lstat(filepath.Join(runtimeRoot, ".my-friday", "capability-contract.json")); errors.Is(statErr, os.ErrNotExist) {
			if err = repository.InitializeCapabilities(runtimeRoot); err != nil {
				return err
			}
		}
	}
	if err = os.Mkdir(capabilities, 0o700); err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}
	agentsRoot := filepath.Join(plan.Paths.Root, "workspace", ".agents")
	skillsRoot := filepath.Join(agentsRoot, "skills")
	_, agentsErr := os.Lstat(agentsRoot)
	_, skillsErr := os.Lstat(skillsRoot)
	if err = os.MkdirAll(filepath.Join(builder, "agents"), 0o700); err != nil {
		_ = os.Remove(capabilities)
		return err
	}
	_ = agentsErr
	_ = skillsErr
	if upgradeHook != nil {
		if err = upgradeHook("builder-created"); err != nil {
			return err
		}
	}
	builderBytes := capabilityBuilder(plan.Paths)
	policy := []byte(builderPolicy)
	if err = os.WriteFile(filepath.Join(builder, "SKILL.md"), builderBytes, 0o600); err != nil {
		return err
	}
	if err = os.WriteFile(filepath.Join(builder, "agents", "openai.yaml"), policy, 0o600); err != nil {
		return err
	}
	if existing, info, existingErr := regular(managedMyFriday); existingErr == nil {
		if info.Mode().Perm() != 0o700 || !bytes.Equal(existing, candidateBytes) {
			return errors.New("managed My Friday executable drift")
		}
	} else if !errors.Is(existingErr, os.ErrNotExist) {
		return existingErr
	} else if err = CopyFile(managedMyFriday, bytes.NewReader(candidateBytes), 0o700); err != nil {
		return err
	}
	configBytes, configErr := managedCodexConfig(plan.Paths, CapabilityRevision)
	if configErr != nil {
		return configErr
	}
	if err = os.WriteFile(m.CodexConfig, configBytes, 0o600); err != nil {
		return err
	}
	if upgradeHook != nil {
		if err = upgradeHook("execution-context-created"); err != nil {
			return err
		}
	}
	sourceContract, sourceRevision := m.ContractVersion, m.CapabilityRevision
	m.ContractVersion = ContractVersion
	m.CapabilityRevision = CapabilityRevision
	m.RollbackContractVersion = sourceContract
	m.RollbackCapabilityRevision = sourceRevision
	if sourceContract == 1 {
		m.Owned = append(m.Owned, "capabilities")
		sort.Strings(m.Owned)
	}
	m.CapabilityBuilder = builder
	m.CapabilityBuilderSHA256 = digest(builderBytes)
	m.CapabilityPolicySHA256 = digest(policy)
	m.MyFridayExecutable = managedMyFriday
	m.MyFridaySHA256 = digest(candidateBytes)
	m.CodexConfigSHA256 = digest(configBytes)
	body, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	body = append(body, '\n')
	tmp := filepath.Join(plan.Paths.Root, "manifest.json.upgrade")
	if err = os.WriteFile(tmp, body, 0o600); err != nil {
		return err
	}
	if err = os.Rename(tmp, filepath.Join(plan.Paths.Root, "manifest.json")); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	if _, err = verify(plan.Paths.Home, plan.Paths.Name); err != nil {
		return fmt.Errorf("assistant upgrade recovery required: %w", err)
	}
	if err = os.Remove(candidateStage); err != nil {
		return err
	}
	return os.Remove(capabilityMigrationPath(plan.Paths.Root))
}

func verify(home, name string) (Manifest, error) {
	p, err := Derive(home, name)
	if err != nil {
		return Manifest{}, err
	}
	if err = safeDir(p.Root, 0700); err != nil {
		return Manifest{}, err
	}
	m, err := readManifest(p)
	if err != nil {
		return m, err
	}
	for _, child := range m.Owned {
		if err = safeDir(filepath.Join(p.Root, child), 0700); err != nil {
			return m, err
		}
	}
	lb, info, err := regular(p.Launcher)
	if err != nil {
		return m, err
	}
	if info.Mode().Perm() != 0700 || digest(lb) != m.LauncherSHA256 {
		return m, errors.New("launcher artifact drift")
	}
	cb, cinfo, err := regular(m.CodexExecutable)
	if err != nil {
		return m, fmt.Errorf("managed Codex executable unavailable: %w", err)
	}
	if cinfo.Mode().Perm() != 0700 || digest(cb) != m.CodexSHA256 {
		return m, errors.New("managed Codex executable drift")
	}
	if m.ContractVersion == ContractVersion && m.CapabilityRevision == CapabilityRevision {
		mb, minfo, managedErr := regular(m.MyFridayExecutable)
		if managedErr != nil {
			return m, fmt.Errorf("managed My Friday executable unavailable: %w", managedErr)
		}
		if minfo.Mode().Perm() != 0700 || digest(mb) != m.MyFridaySHA256 {
			return m, errors.New("managed My Friday executable drift")
		}
	}
	if err = verifyManagedConfig(p, m); err != nil {
		return m, err
	}
	if err = verifyManagedInstructions(p, m); err != nil {
		return m, err
	}
	if m.ContractVersion == ContractVersion {
		if err = verifyCapabilityBuilder(p, m); err != nil {
			return m, err
		}
	}
	return m, nil
}

func PlanRemove(home, name string) (Plan, error) {
	p, err := Derive(home, name)
	if err != nil {
		return Plan{}, err
	}
	if _, err = Verify(home, name); err != nil {
		return Plan{}, fmt.Errorf("remove refused: %w", err)
	}
	info, err := os.Lstat(p.Root)
	if err != nil {
		return Plan{}, err
	}
	st := info.Sys().(*syscall.Stat_t)
	return Plan{Action: "remove", Paths: p, Items: []string{"remove exact manifest-owned launcher", "remove exact manifest-owned instance root", "leave launcher directory, siblings, other instances, HOME, shell files, and user Codex state unchanged"}, rootDevice: uint64(st.Dev), rootInode: uint64(st.Ino)}, nil
}

func matchesRoot(p Paths, device, inode uint64) error {
	info, err := os.Lstat(p.Root)
	if err != nil {
		return err
	}
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 || uint64(st.Dev) != device || uint64(st.Ino) != inode {
		return errors.New("instance root identity changed")
	}
	return nil
}

func removeOwnedRoot(p Paths, m Manifest) error {
	current, err := readManifest(p)
	if err != nil {
		return err
	}
	if current.LauncherSHA256 != m.LauncherSHA256 {
		return errors.New("manifest changed before root removal")
	}
	return os.RemoveAll(p.Root)
}

func Remove(plan Plan) error {
	expected := rootProof{Device: plan.rootDevice, Inode: plan.rootInode}
	lock, err := acquireTransactionLock(plan.Paths.Root, &expected)
	if err != nil {
		return err
	}
	return finishTransaction(lock, removeLocked(plan))
}

func removeLocked(plan Plan) error {
	if plan.Action != "remove" {
		return errors.New("invalid remove plan")
	}
	m, err := verify(plan.Paths.Home, plan.Paths.Name)
	if err != nil {
		return fmt.Errorf("remove refused: %w", err)
	}
	if err = matchesRoot(plan.Paths, plan.rootDevice, plan.rootInode); err != nil {
		return fmt.Errorf("remove refused: %w", err)
	}
	if err = os.Remove(plan.Paths.Launcher); err != nil {
		return err
	}
	if err = matchesRoot(plan.Paths, plan.rootDevice, plan.rootInode); err != nil {
		return fmt.Errorf("recovery required: launcher removed but root changed: %w", err)
	}
	if err = removeOwnedRoot(plan.Paths, m); err != nil {
		return fmt.Errorf("recovery required: launcher removed but root retained: %w", err)
	}
	return nil
}

// Recover resolves only an interrupted operation whose retained manifest still
// proves the exact instance root. A missing launcher makes absence the safe
// deterministic result; a complete triple is left active.
func Recover(home, name string) (string, error) {
	p, err := Derive(home, name)
	if err != nil {
		return "", err
	}
	if _, journalErr := os.Lstat(capabilityMigrationPath(p.Root)); journalErr == nil {
		info, statErr := os.Lstat(p.Root)
		if statErr != nil {
			return "", statErr
		}
		st := info.Sys().(*syscall.Stat_t)
		expected := rootProof{Device: uint64(st.Dev), Inode: uint64(st.Ino)}
		lock, lockErr := acquireTransactionLock(p.Root, &expected)
		if lockErr != nil {
			return "", lockErr
		}
		journal, readErr := readCapabilityMigration(p.Root)
		if readErr != nil {
			return "", finishTransaction(lock, readErr)
		}
		if journal.Action == "upgrade" {
			if m, verifyErr := verify(home, name); verifyErr == nil && m.ContractVersion == ContractVersion && m.CapabilityRevision == CapabilityRevision {
				stage := filepath.Join(p.Root, "dependencies", "my-friday.upgrade")
				body, info, stageErr := regular(stage)
				if stageErr != nil || info.Mode().Perm() != 0o700 || digest(body) != journal.CandidateSHA256 {
					return "", finishTransaction(lock, errors.New("assistant upgrade candidate staging drift"))
				}
				removeErr := os.Remove(stage)
				if removeErr == nil {
					removeErr = os.Remove(capabilityMigrationPath(p.Root))
				}
				return "completed capability upgrade", finishTransaction(lock, removeErr)
			}
			upgradeErr := upgradeLocked(Plan{Action: "upgrade", Paths: p, rootDevice: uint64(st.Dev), rootInode: uint64(st.Ino)})
			return "completed capability upgrade", finishTransaction(lock, upgradeErr)
		}
		if journal.Action == "rollback" {
			m, manifestErr := readManifest(p)
			if manifestErr != nil {
				return "", finishTransaction(lock, manifestErr)
			}
			if m.ContractVersion == journal.TargetContractVersion && m.CapabilityRevision == journal.TargetCapabilityRevision {
				if _, verifyErr := verify(home, name); verifyErr != nil {
					return "", finishTransaction(lock, verifyErr)
				}
				return "completed capability rollback", finishTransaction(lock, finishRollbackQuarantines(p, journal))
			}
			rollbackErr := completeRollback(p, m)
			return "completed capability rollback", finishTransaction(lock, rollbackErr)
		}
		return "", finishTransaction(lock, errors.New("unsupported assistant migration journal"))
	} else if !errors.Is(journalErr, os.ErrNotExist) {
		return "", journalErr
	}
	if _, err = Verify(home, name); err == nil {
		return "already healthy", nil
	}
	info, err := os.Lstat(p.Root)
	if err != nil {
		return "", err
	}
	st := info.Sys().(*syscall.Stat_t)
	expected := rootProof{Device: uint64(st.Dev), Inode: uint64(st.Ino)}
	m, err := recoverableManifest(p)
	if err != nil {
		return "", err
	}
	if err = matchOpenedRoot(p.Root, expected); err != nil {
		return "", fmt.Errorf("recovery refused: %w", err)
	}
	lock, err := acquireTransactionLock(p.Root, &expected)
	if err != nil {
		return "", err
	}
	result, operationErr := recoverLocked(p, m, lock.proof)
	return result, finishTransaction(lock, operationErr)
}

func recoverLocked(p Paths, m Manifest, proof rootProof) (string, error) {
	if err := matchOpenedRoot(p.Root, proof); err != nil {
		return "", err
	}
	if _, launcherErr := os.Lstat(p.Launcher); launcherErr == nil {
		return "", errors.New("recovery refused: launcher exists but the instance is not healthy")
	} else if !errors.Is(launcherErr, os.ErrNotExist) {
		return "", launcherErr
	}
	if err := removeOwnedRoot(p, m); err != nil {
		return "", fmt.Errorf("recovery refused: %w", err)
	}
	return "restored absent state", nil
}

func recoverableManifest(p Paths) (Manifest, error) {
	if err := safeDir(p.Root, 0700); err != nil {
		return Manifest{}, fmt.Errorf("recovery refused: %w", err)
	}
	m, err := readManifest(p)
	if err != nil {
		return m, fmt.Errorf("recovery refused: %w", err)
	}
	for _, child := range m.Owned {
		if err = safeDir(filepath.Join(p.Root, child), 0700); err != nil {
			return m, fmt.Errorf("recovery refused: %w", err)
		}
	}
	cb, info, err := regular(m.CodexExecutable)
	if err != nil || info.Mode().Perm() != 0700 || digest(cb) != m.CodexSHA256 {
		return m, errors.New("recovery refused: managed Codex executable drift")
	}
	if err = verifyManagedConfig(p, m); err != nil {
		return m, fmt.Errorf("recovery refused: %w", err)
	}
	if err = verifyManagedInstructions(p, m); err != nil {
		return m, fmt.Errorf("recovery refused: %w", err)
	}
	if m.ContractVersion == ContractVersion && m.CapabilityRevision == CapabilityRevision {
		mb, minfo, managedErr := regular(m.MyFridayExecutable)
		if managedErr != nil || minfo.Mode().Perm() != 0700 || digest(mb) != m.MyFridaySHA256 {
			return m, errors.New("recovery refused: managed My Friday executable drift")
		}
	}
	if m.ContractVersion == ContractVersion {
		if err = verifyCapabilityBuilder(p, m); err != nil {
			return m, fmt.Errorf("recovery refused: %w", err)
		}
	}
	if _, launcherErr := os.Lstat(p.Launcher); launcherErr == nil {
		return m, errors.New("recovery refused: launcher exists but the instance is not healthy")
	} else if !errors.Is(launcherErr, os.ErrNotExist) {
		return m, launcherErr
	}
	return m, nil
}

// Migrate serializes the named replacement through final verification and the
// caller-supplied manifest-governed legacy cleanup transaction.
func Migrate(plan Plan, executable, codex string, cleanup func() error) error {
	return createLocked(plan, executable, codex, cleanup)
}

func Launch(home, name string, args []string) error {
	m, err := Verify(home, name)
	if err != nil {
		return fmt.Errorf("launch refused: %w", err)
	}
	workspace := filepath.Join(m.Root, "workspace")
	argv := append([]string{"codex", "--cd", workspace, "--"}, args...)
	env := make([]string, 0, len(os.Environ())+1)
	for _, item := range os.Environ() {
		if !strings.HasPrefix(item, "CODEX_HOME=") {
			env = append(env, item)
		}
	}
	env = append(env, "CODEX_HOME="+filepath.Join(m.Root, "codex"))
	return syscall.Exec(m.CodexExecutable, argv, env)
}

func CopyFile(dst string, src io.Reader, mode fs.FileMode) error {
	f, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, src)
	return err
}

func FindCodex() (string, error) {
	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		if dir == "" || !filepath.IsAbs(dir) {
			continue
		}
		candidate := filepath.Join(dir, "codex")
		resolved, err := filepath.EvalSymlinks(candidate)
		if err != nil {
			continue
		}
		_, info, err := regular(resolved)
		if err == nil && info.Mode().Perm()&0111 != 0 {
			return resolved, nil
		}
	}
	return "", errors.New("Codex executable not found on absolute PATH entries")
}

func validDigest(value string) bool {
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == sha256.Size && value == strings.ToLower(value)
}
