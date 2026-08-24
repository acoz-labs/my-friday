// Package assistantinstance manages named, manifest-owned assistant instances.
package assistantinstance

import (
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
)

const ContractVersion = 1

var namePattern = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,30}[a-z0-9])?$`)
var reserved = map[string]bool{"codex": true, "my-friday": true, "default": true, "current": true, "new": true}

type Manifest struct {
	ContractVersion int      `json:"contract_version"`
	Name            string   `json:"name"`
	Root            string   `json:"root"`
	Owned           []string `json:"owned"`
	Launcher        string   `json:"launcher"`
	LauncherSHA256  string   `json:"launcher_sha256"`
	CodexExecutable string   `json:"codex_executable"`
	CodexSHA256     string   `json:"codex_sha256"`
	AssistantID     string   `json:"assistant_id,omitempty"`
}

type Paths struct {
	Home, Name, Root, Launcher string
}

type rootProof struct{ Device, Inode uint64 }

type Plan struct {
	Action        string
	Paths         Paths
	Items         []string
	RuntimeSource string
	MemorySource  string
	AssistantID   string
	rootDevice    uint64
	rootInode     uint64
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
	return Plan{Action: "create", Paths: p, Items: []string{"create private instance root", "create codex, runtime, memory, workspace, and dependencies", "install exact native launcher with no replacement", "leave HOME, shell files, user Codex state, and launcher siblings unchanged"}}, nil
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

func createLocked(plan Plan, executable, codex string, afterVerify func() error) (resultErr error) {
	if plan.Action != "create" {
		return errors.New("invalid create plan")
	}
	p, err := Derive(plan.Paths.Home, plan.Paths.Name)
	if err != nil || p != plan.Paths {
		return errors.New("create plan path mismatch")
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
	for _, child := range []string{"codex", "runtime", "memory", "workspace", "dependencies"} {
		if err = os.Mkdir(filepath.Join(stage, child), 0700); err != nil {
			clean()
			return err
		}
	}
	owned := []string{"codex", "dependencies", "memory", "runtime", "workspace"}
	sort.Strings(owned)
	managedCodex := filepath.Join(p.Root, "dependencies", "codex")
	codexBytes, _, err := regular(codex)
	if err != nil {
		clean()
		return err
	}
	m := Manifest{ContractVersion: ContractVersion, Name: p.Name, Root: p.Root, Owned: owned, Launcher: p.Launcher, LauncherSHA256: digest(launcher), CodexExecutable: managedCodex, CodexSHA256: digest(codexBytes), AssistantID: plan.AssistantID}
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
	if plan.RuntimeSource != "" {
		if err = copyTree(plan.RuntimeSource, filepath.Join(stage, "runtime")); err != nil {
			clean()
			return err
		}
		if err = copyTree(plan.MemorySource, filepath.Join(stage, "memory")); err != nil {
			clean()
			return err
		}
		agents, _, readErr := regular(filepath.Join(plan.RuntimeSource, "AGENTS.md"))
		if readErr != nil {
			clean()
			return readErr
		}
		if err = os.WriteFile(filepath.Join(stage, "codex", "AGENTS.md"), agents, 0600); err != nil {
			clean()
			return err
		}
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
	want := []string{"codex", "dependencies", "memory", "runtime", "workspace"}
	wantCodex := filepath.Join(p.Root, "dependencies", "codex")
	if m.ContractVersion != ContractVersion || m.Name != p.Name || m.Root != p.Root || m.Launcher != p.Launcher || m.CodexExecutable != wantCodex || !validDigest(m.CodexSHA256) || strings.Join(m.Owned, "\x00") != strings.Join(want, "\x00") {
		return m, errors.New("manifest ownership contract mismatch")
	}
	return m, nil
}

func Verify(home, name string) (Manifest, error) { return verify(home, name) }

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
