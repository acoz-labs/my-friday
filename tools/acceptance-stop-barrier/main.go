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
	startIdentity := processIdentity(pgid)
	if startIdentity == "" {
		fatal(errors.New("could not capture candidate process start identity"))
	}
	journalPaths := []string{
		filepath.Join(*home, ".my-friday", "transaction.json"),
		filepath.Join(*home, ".my-friday", "transaction.json.next"),
		filepath.Join(*home, ".my-friday-removing", "transaction.json"),
	}
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		if processIdentity(pgid) != startIdentity {
			fatal(errors.New("candidate PID/start identity changed"))
		}
		if err := syscall.Kill(-pgid, syscall.SIGSTOP); err != nil {
			if errors.Is(err, syscall.ESRCH) {
				break
			}
			fatal(err)
		}
		time.Sleep(30 * time.Millisecond)
		if !groupStopped(pgid) {
			_ = syscall.Kill(-pgid, syscall.SIGCONT)
			continue
		}
		if validJournal(journalPaths, *action, *home) {
			if processIdentity(pgid) != startIdentity || !groupStopped(pgid) {
				fatal(errors.New("candidate identity or stopped barrier changed"))
			}
			if err := syscall.Kill(-pgid, syscall.SIGKILL); err != nil && !errors.Is(err, syscall.ESRCH) {
				fatal(err)
			}
			_ = cmd.Wait()
			if groupPresent(pgid) {
				fatal(errors.New("candidate process group retained surviving members"))
			}
			if !validJournal(journalPaths, *action, *home) {
				fatal(errors.New("journal changed after stopped-group kill"))
			}
			fmt.Printf("captured recoverable %s interruption\n", *action)
			return
		}
		_ = syscall.Kill(-pgid, syscall.SIGCONT)
		time.Sleep(10 * time.Millisecond)
	}
	_ = syscall.Kill(-pgid, syscall.SIGCONT)
	_ = syscall.Kill(-pgid, syscall.SIGKILL)
	_ = cmd.Wait()
	fmt.Fprintf(os.Stderr, "no recoverable %s interruption was captured: %s\n", *action, output.String())
	os.Exit(1)
}

func validJournal(paths []string, action, home string) bool {
	var homeStat unix.Stat_t
	if unix.Lstat(home, &homeStat) != nil || homeStat.Mode&unix.S_IFMT != unix.S_IFDIR {
		return false
	}
	var journals []journal
	for _, path := range paths {
		body, err := readNoFollow(path)
		if err != nil {
			continue
		}
		var value journal
		decoder := json.NewDecoder(bytes.NewReader(body))
		decoder.DisallowUnknownFields()
		if decoder.Decode(&value) != nil || decoder.Decode(&struct{}{}) != io.EOF || value.ContractVersion != 1 || value.Action != action ||
			phaseRank(value.Phase) < 0 || value.Phase == "committed" || value.Root.Device != uint64(homeStat.Dev) ||
			value.Root.Inode != homeStat.Ino || !validProofSet(value.Before) || !validProofSet(value.After) || !validTransition(value) {
			return false
		}
		journals = append(journals, value)
	}
	if len(journals) == 0 || len(journals) > 2 {
		return false
	}
	if len(journals) == 2 && (journals[0].Action != journals[1].Action || journals[0].Root != journals[1].Root || abs(phaseRank(journals[0].Phase)-phaseRank(journals[1].Phase)) != 1) {
		return false
	}
	return true
}

func readNoFollow(path string) ([]byte, error) {
	fd, err := unix.Open(path, unix.O_RDONLY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if err != nil {
		return nil, err
	}
	f := os.NewFile(uintptr(fd), path)
	defer f.Close()
	return io.ReadAll(io.LimitReader(f, 16<<20))
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
