// acceptance-stop-barrier is a repository-owned acceptance supervisor helper.
// It externally stops an unmodified production candidate only after a durable,
// schema-valid lifecycle journal is observable, then kills that exact process
// group while stopped. It is not shipped in the product artifact.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

type fileProof struct {
	Exists bool   `json:"exists"`
	SHA256 string `json:"sha256,omitempty"`
	Bytes  []byte `json:"bytes,omitempty"`
}
type rootProof struct{ Device, Inode uint64 }
type journal struct {
	ContractVersion int                  `json:"contract_version"`
	Action          string               `json:"action"`
	Phase           string               `json:"phase"`
	Root            rootProof            `json:"root"`
	Before          map[string]fileProof `json:"before"`
	After           map[string]fileProof `json:"after"`
}
type installedManifest struct {
	ContractVersion  int    `json:"contract_version"`
	AssistantID      string `json:"assistant_id"`
	ProjectionSHA256 string `json:"projection_sha256"`
	SourcePath       string `json:"source_path"`
	SourceSHA256     string `json:"source_sha256"`
	CanonicalSHA256  string `json:"canonical_sha256"`
	PreviousSHA256   string `json:"previous_sha256,omitempty"`
}

func main() {
	profile := flag.String("profile", "", "sandbox profile")
	candidate := flag.String("candidate", "", "candidate executable")
	runtime := flag.String("runtime", "", "runtime fixture")
	home := flag.String("home", "", "synthetic home")
	action := flag.String("action", "", "install, upgrade, or uninstall")
	flag.Parse()
	if *profile == "" || *candidate == "" || *runtime == "" || *home == "" ||
		(*action != "install" && *action != "upgrade" && *action != "uninstall") {
		fmt.Fprintln(os.Stderr, "usage: acceptance-stop-barrier --profile P --candidate C --runtime R --home H --action install|upgrade|uninstall")
		os.Exit(64)
	}
	args := []string{"-f", *profile, *candidate, "codex", *action}
	if *action != "uninstall" {
		args = append(args, "--runtime", *runtime)
	}
	cmd := exec.Command("/usr/bin/sandbox-exec", args...)
	cmd.Dir = filepath.Dir(*home)
	cmd.Env = []string{
		"HOME=" + filepath.Dir(*home), "CODEX_HOME=" + *home,
		"TMPDIR=" + filepath.Join(filepath.Dir(*home), "tmp"),
		"XDG_CACHE_HOME=" + filepath.Join(filepath.Dir(*home), "xdg-cache"),
		"XDG_CONFIG_HOME=" + filepath.Join(filepath.Dir(*home), "xdg-config"),
		"LANG=en_US.UTF-8", "PATH=/usr/bin:/bin:/usr/sbin:/sbin",
	}
	tokens := map[string]string{"install": "Install", "upgrade": "Upgrade", "uninstall": "Uninstall"}
	cmd.Stdin = strings.NewReader(tokens[*action] + "\n")
	var output bytes.Buffer
	cmd.Stdout, cmd.Stderr = &output, &output
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		fatal(err)
	}
	pgid := cmd.Process.Pid
	var waitOnce sync.Once
	cleanup := func() {
		_ = syscall.Kill(-pgid, syscall.SIGCONT)
		_ = syscall.Kill(-pgid, syscall.SIGKILL)
		waitOnce.Do(func() { _ = cmd.Wait() })
		for retry := 0; retry < 100 && groupPresent(pgid); retry++ {
			_ = syscall.Kill(-pgid, syscall.SIGKILL)
			time.Sleep(10 * time.Millisecond)
		}
	}
	abort := func(err error) { cleanup(); fmt.Fprintln(os.Stderr, err); os.Exit(1) }
	defer cleanup()
	startIdentity := processIdentity(pgid)
	if startIdentity == "" {
		abort(errors.New("could not capture candidate process start identity"))
	}
	journalPaths := []string{
		filepath.Join(*home, ".my-friday", "transaction.json"),
		filepath.Join(*home, ".my-friday", "transaction.json.next"),
		filepath.Join(*home, ".my-friday-removing", "transaction.json"),
	}
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		if processIdentity(pgid) != startIdentity {
			abort(errors.New("candidate PID/start identity changed"))
		}
		if err := syscall.Kill(-pgid, syscall.SIGSTOP); err != nil {
			if errors.Is(err, syscall.ESRCH) {
				break
			}
			abort(err)
		}
		time.Sleep(30 * time.Millisecond)
		if !groupStopped(pgid) {
			_ = syscall.Kill(-pgid, syscall.SIGCONT)
			continue
		}
		receipt, valid := captureJournal(journalPaths, *action, *home)
		if valid && stableJournal(journalPaths, *action, *home, receipt, startIdentity, pgid) {
			if processIdentity(pgid) != startIdentity || !groupStopped(pgid) {
				abort(errors.New("candidate identity or stopped barrier changed"))
			}
			if err := syscall.Kill(-pgid, syscall.SIGKILL); err != nil && !errors.Is(err, syscall.ESRCH) {
				abort(err)
			}
			waitOnce.Do(func() { _ = cmd.Wait() })
			if groupPresent(pgid) {
				abort(errors.New("candidate process group retained surviving members"))
			}
			after, stillValid := captureJournal(journalPaths, *action, *home)
			if !stillValid || after != receipt {
				abort(errors.New("journal changed after stopped-group kill"))
			}
			fmt.Printf("captured recoverable %s interruption\n", *action)
			return
		}
		_ = syscall.Kill(-pgid, syscall.SIGCONT)
		time.Sleep(10 * time.Millisecond)
	}
	_ = syscall.Kill(-pgid, syscall.SIGCONT)
	_ = syscall.Kill(-pgid, syscall.SIGKILL)
	waitOnce.Do(func() { _ = cmd.Wait() })
	fmt.Fprintf(os.Stderr, "no recoverable %s interruption was captured: %s\n", *action, output.String())
	os.Exit(1)
}

func stableJournal(paths []string, action, home, want, identity string, pgid int) bool {
	for attempt := 0; attempt < 3; attempt++ {
		time.Sleep(25 * time.Millisecond)
		if processIdentity(pgid) != identity || !groupStopped(pgid) {
			return false
		}
		got, ok := captureJournal(paths, action, home)
		if !ok || got != want {
			return false
		}
	}
	return true
}

func validJournal(paths []string, action, home string) bool {
	_, ok := captureJournal(paths, action, home)
	return ok
}

func captureJournal(paths []string, action, home string) (string, bool) {
	_ = paths
	root, err := openAbsoluteDirNoFollow(home)
	if err != nil {
		return "", false
	}
	defer unix.Close(root)
	var homeStat unix.Stat_t
	if unix.Fstat(root, &homeStat) != nil || homeStat.Mode&unix.S_IFMT != unix.S_IFDIR || homeStat.Uid != uint32(os.Getuid()) {
		return "", false
	}
	control, err := unix.Openat(root, ".my-friday", unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if err != nil {
		return "", false
	}
	defer unix.Close(control)
	var journals []journal
	for _, name := range []string{"transaction.json", "transaction.json.next"} {
		proof, proofErr := proofAt(control, name)
		if proofErr != nil {
			return "", false
		}
		if !proof.Exists {
			continue
		}
		body := proof.Bytes
		var value journal
		decoder := json.NewDecoder(bytes.NewReader(body))
		decoder.DisallowUnknownFields()
		if decoder.Decode(&value) != nil || decoder.Decode(&struct{}{}) != io.EOF || value.ContractVersion != 1 || value.Action != action ||
			phaseRank(value.Phase) < 0 || value.Phase == "committed" || value.Root.Device != uint64(homeStat.Dev) ||
			value.Root.Inode != homeStat.Ino || !validProofSet(value.Before) || !validProofSet(value.After) || !validTransition(value) {
			return "", false
		}
		journals = append(journals, value)
	}
	if len(journals) == 0 || len(journals) > 2 {
		return "", false
	}
	if len(journals) == 2 && (journals[0].Action != journals[1].Action || journals[0].Root != journals[1].Root || abs(phaseRank(journals[0].Phase)-phaseRank(journals[1].Phase)) != 1) {
		return "", false
	}
	actual, ok := actualProofsAt(root, control)
	if !ok {
		return "", false
	}
	effective := journals[0]
	for _, candidate := range journals[1:] {
		if phaseRank(candidate.Phase) > phaseRank(effective.Phase) {
			effective = candidate
		}
	}
	if !actualMatchesPhase(actual, effective) || !validStagingAt(control, journals) {
		return "", false
	}
	body, _ := json.Marshal(struct {
		Journals []journal            `json:"journals"`
		Actual   map[string]fileProof `json:"actual"`
	}{journals, actual})
	return fmt.Sprintf("%x", sha256.Sum256(body)), true
}

func actualProofs(home string) (map[string]fileProof, bool) {
	root, err := openAbsoluteDirNoFollow(home)
	if err != nil {
		return nil, false
	}
	defer unix.Close(root)
	control, err := unix.Openat(root, ".my-friday", unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if err != nil {
		return nil, false
	}
	defer unix.Close(control)
	return actualProofsAt(root, control)
}

func actualProofsAt(root, control int) (map[string]fileProof, bool) {
	result := map[string]fileProof{}
	for name, location := range map[string]struct {
		fd   int
		file string
	}{
		"projection": {root, "AGENTS.md"}, "manifest": {control, "installed-baseline.json"},
		"canonical": {control, "canonical-AGENTS.md"}, "previous": {control, "previous-AGENTS.md"},
	} {
		proof, proofErr := proofAt(location.fd, location.file)
		if proofErr != nil {
			return nil, false
		}
		result[name] = proof
	}
	return result, true
}

func proofAt(fd int, name string) (fileProof, error) {
	opened, err := unix.Openat(fd, name, unix.O_RDONLY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if errors.Is(err, unix.ENOENT) {
		return fileProof{}, nil
	}
	if err != nil {
		return fileProof{}, err
	}
	f := os.NewFile(uintptr(opened), name)
	defer f.Close()
	var st unix.Stat_t
	if err = unix.Fstat(opened, &st); err != nil || st.Mode&unix.S_IFMT != unix.S_IFREG || st.Uid != uint32(os.Getuid()) || st.Nlink != 1 || st.Mode&0o777 != 0o600 {
		return fileProof{}, errors.New("unsafe managed file")
	}
	body, err := io.ReadAll(io.LimitReader(f, 16<<20))
	if err != nil {
		return fileProof{}, err
	}
	return fileProof{Exists: true, SHA256: fmt.Sprintf("%x", sha256.Sum256(body)), Bytes: body}, nil
}

func actualMatchesPhase(actual map[string]fileProof, j journal) bool {
	rank := phaseRank(j.Phase)
	threshold := map[string]int{"projection": 2, "canonical": 3, "previous": 4, "manifest": 5}
	for name, at := range threshold {
		want := j.Before[name]
		if rank >= at {
			want = j.After[name]
		}
		if !proofEqual(actual[name], want) {
			return false
		}
	}
	return true
}
func proofEqual(a, b fileProof) bool {
	return a.Exists == b.Exists && a.SHA256 == b.SHA256 && bytes.Equal(a.Bytes, b.Bytes)
}

func validStagingAt(control int, journals []journal) bool {
	allowed := map[string]bool{"transaction.json": true, "transaction.json.next": true, "installed-baseline.json": true, "canonical-AGENTS.md": true, "previous-AGENTS.md": true,
		"installed-baseline.json.next": true, "canonical-AGENTS.md.next": true, "previous-AGENTS.md.next": true}
	dup, err := unix.Dup(control)
	if err != nil {
		return false
	}
	dir := os.NewFile(uintptr(dup), "control")
	entries, err := dir.ReadDir(-1)
	_ = dir.Close()
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if !allowed[entry.Name()] {
			return false
		}
	}
	for _, name := range []string{"installed-baseline.json.next", "canonical-AGENTS.md.next", "previous-AGENTS.md.next"} {
		proof, err := proofAt(control, name)
		if err != nil {
			return false
		}
		if !proof.Exists {
			continue
		}
		sum := proof.SHA256
		matched := false
		key := strings.TrimSuffix(strings.TrimSuffix(name, ".next"), "-AGENTS.md")
		if name == "canonical-AGENTS.md.next" {
			key = "canonical"
		}
		if name == "previous-AGENTS.md.next" {
			key = "previous"
		}
		if name == "installed-baseline.json.next" {
			key = "manifest"
		}
		for _, j := range journals {
			if j.Before[key].SHA256 == sum || j.After[key].SHA256 == sum {
				matched = true
			}
		}
		if !matched {
			return false
		}
	}
	return true
}

func openAbsoluteDirNoFollow(path string) (int, error) {
	if !filepath.IsAbs(path) {
		return -1, errors.New("path must be absolute")
	}
	clean := filepath.Clean(path)
	if clean == "/var" || strings.HasPrefix(clean, "/var/") {
		clean = filepath.Join("/private/var", strings.TrimPrefix(clean, "/var/"))
	}
	fd, err := unix.Open("/", unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if err != nil {
		return -1, err
	}
	for _, part := range strings.Split(strings.TrimPrefix(clean, "/"), "/") {
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
func validProofSet(set map[string]fileProof) bool {
	if len(set) != 4 {
		return false
	}
	for _, name := range []string{"projection", "manifest", "canonical", "previous"} {
		p, ok := set[name]
		if !ok {
			return false
		}
		if !p.Exists {
			if p.SHA256 != "" || len(p.Bytes) != 0 {
				return false
			}
			continue
		}
		d := fmt.Sprintf("%x", sha256.Sum256(p.Bytes))
		if p.SHA256 != d {
			return false
		}
	}
	return true
}
func validTransition(j journal) bool {
	allAbsent := func(s map[string]fileProof) bool {
		for _, p := range s {
			if p.Exists {
				return false
			}
		}
		return true
	}
	if !validGeneration(j.Before, true) || !validGeneration(j.After, j.Action == "uninstall") {
		return false
	}
	switch j.Action {
	case "install":
		return allAbsent(j.Before) && !allAbsent(j.After) && j.After["projection"].SHA256 == j.After["canonical"].SHA256
	case "uninstall":
		return !allAbsent(j.Before) && allAbsent(j.After) && j.Before["projection"].SHA256 == j.Before["canonical"].SHA256
	case "upgrade":
		return j.Before["projection"].Exists && j.After["projection"].Exists && j.After["previous"].SHA256 == j.Before["projection"].SHA256 && j.After["projection"].SHA256 == j.After["canonical"].SHA256
	}
	return false
}

func validGeneration(set map[string]fileProof, allowAbsent bool) bool {
	allAbsent := true
	for _, p := range set {
		if p.Exists {
			allAbsent = false
		}
	}
	if allAbsent {
		return allowAbsent
	}
	manifestProof := set["manifest"]
	if !manifestProof.Exists {
		return false
	}
	var m installedManifest
	decoder := json.NewDecoder(bytes.NewReader(manifestProof.Bytes))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&m) != nil || decoder.Decode(&struct{}{}) != io.EOF || m.ContractVersion != 1 || m.AssistantID == "" || m.SourcePath == "" || len(m.SourceSHA256) != 64 {
		return false
	}
	return set["projection"].Exists && set["canonical"].Exists && set["projection"].SHA256 == m.ProjectionSHA256 && set["canonical"].SHA256 == m.CanonicalSHA256 && set["previous"].Exists == (m.PreviousSHA256 != "") && set["previous"].SHA256 == m.PreviousSHA256
}
func phaseRank(p string) int {
	for i, v := range []string{"prepared", "mutating", "projection-written", "canonical-written", "previous-written", "manifest-written", "final-verified", "committed"} {
		if p == v {
			return i
		}
	}
	return -1
}
func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
func processIdentity(pid int) string {
	out, err := exec.Command("/bin/ps", "-p", strconv.Itoa(pid), "-o", "pid=,pgid=,lstart=").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func groupStopped(pgid int) bool {
	out, err := exec.Command("/bin/ps", "-ax", "-o", "pgid=,stat=").Output()
	if err != nil {
		return false
	}
	found := false
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		group, err := strconv.Atoi(fields[0])
		if err == nil && group == pgid {
			found = true
			if !strings.Contains(fields[1], "T") {
				return false
			}
		}
	}
	return found
}

func groupPresent(pgid int) bool {
	out, err := exec.Command("/bin/ps", "-ax", "-o", "pgid=").Output()
	if err != nil {
		return true
	}
	for _, field := range strings.Fields(string(out)) {
		group, err := strconv.Atoi(field)
		if err == nil && group == pgid {
			return true
		}
	}
	return false
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
