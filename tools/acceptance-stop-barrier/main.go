// acceptance-stop-barrier is a repository-owned acceptance supervisor helper.
// It externally stops an unmodified production candidate only after a durable,
// schema-valid lifecycle journal is observable, then kills that exact process
// group while stopped. It is not shipped in the product artifact.
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type journal struct {
	ContractVersion int                        `json:"contract_version"`
	Action          string                     `json:"action"`
	Phase           string                     `json:"phase"`
	Root            map[string]uint64          `json:"root"`
	Before          map[string]json.RawMessage `json:"before"`
	After           map[string]json.RawMessage `json:"after"`
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
	journalPaths := []string{
		filepath.Join(*home, ".my-friday", "transaction.json"),
		filepath.Join(*home, ".my-friday", "transaction.json.next"),
		filepath.Join(*home, ".my-friday-removing", "transaction.json"),
	}
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
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
		if validJournal(journalPaths, *action) {
			if err := syscall.Kill(-pgid, syscall.SIGKILL); err != nil && !errors.Is(err, syscall.ESRCH) {
				fatal(err)
			}
			_ = cmd.Wait()
			if !validJournal(journalPaths, *action) {
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

func validJournal(paths []string, action string) bool {
	for _, path := range paths {
		body, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var value journal
		decoder := json.NewDecoder(bytes.NewReader(body))
		decoder.DisallowUnknownFields()
		if decoder.Decode(&value) == nil && value.ContractVersion == 1 && value.Action == action &&
			value.Phase != "" && value.Phase != "committed" && len(value.Root) == 2 &&
			len(value.Before) == 4 && len(value.After) == 4 {
			return true
		}
	}
	return false
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

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
