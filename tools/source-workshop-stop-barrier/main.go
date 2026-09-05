// source-workshop-stop-barrier interrupts an unmodified candidate only after
// its durable source-workshop journal and bound staging tree are stable.
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
	ContractVersion int    `json:"contract_version"`
	Action          string `json:"action"`
	Slug            string `json:"slug"`
	Phase           string `json:"phase"`
	StageRoot       string `json:"stage_root"`
	StageInode      uint64 `json:"stage_inode"`
}

func main() {
	expect := flag.String("expect", "", "workshop Expect script")
	candidate := flag.String("candidate", "", "candidate executable")
	instance := flag.String("instance", "", "instance name")
	root := flag.String("instance-root", "", "instance root")
	slug := flag.String("slug", "", "capability slug")
	transcript := flag.String("transcript", "", "private transcript")
	flag.Parse()
	if *expect == "" || *candidate == "" || *instance == "" || !filepath.IsAbs(*root) || *slug == "" || !filepath.IsAbs(*transcript) {
		fatal(errors.New("invalid source-workshop stop-barrier arguments"))
	}
	cmd := exec.Command("/usr/bin/expect", *expect, "enhance", *candidate, *instance, *slug, *transcript)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Start(); err != nil {
		fatal(fmt.Errorf("start isolated Expect process: %w", err))
	}
	pgid := cmd.Process.Pid
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	journalPath := filepath.Join(*root, "capabilities", ".workshop-"+*slug+".json")
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		select {
		case err := <-done:
			fatal(fmt.Errorf("candidate completed before source interruption: %w", err))
		default:
		}
		if err := syscall.Kill(-pgid, syscall.SIGSTOP); err != nil {
			leaderErr := syscall.Kill(pgid, syscall.SIGSTOP)
			if leaderErr == nil {
				_ = syscall.Kill(pgid, syscall.SIGCONT)
			}
			fatal(fmt.Errorf("stop Expect process group %d as euid %d (leader stop: %v; members: %s): %w", pgid, os.Geteuid(), leaderErr, groupSummary(pgid), err))
		}
		body, err := os.ReadFile(journalPath)
		if err == nil && stable(body, *root, *slug) {
			time.Sleep(5 * time.Millisecond)
			next, nextErr := os.ReadFile(journalPath)
			if nextErr == nil && bytes.Equal(body, next) && stable(next, *root, *slug) {
				_ = syscall.Kill(-pgid, syscall.SIGKILL)
				<-done
				fmt.Println("captured real source-workshop journal interruption")
				return
			}
		}
		_ = syscall.Kill(-pgid, syscall.SIGCONT)
		time.Sleep(100 * time.Microsecond)
	}
	_ = syscall.Kill(-pgid, syscall.SIGKILL)
	<-done
	fatal(errors.New("no stable source-workshop interruption captured"))
}

func groupSummary(pgid int) string {
	body, err := exec.Command("/bin/ps", "-axo", "pid=,ppid=,pgid=,uid=,stat=,comm=").Output()
	if err != nil {
		return "unavailable"
	}
	var members []string
	for _, line := range strings.Split(string(body), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 6 || fields[2] != strconv.Itoa(pgid) {
			continue
		}
		members = append(members, strings.Join(fields[:5], ":")+":"+filepath.Base(strings.Join(fields[5:], " ")))
	}
	if len(members) == 0 {
		return "none"
	}
	return strings.Join(members, ",")
}

func stable(body []byte, root, slug string) bool {
	var envelope map[string]any
	if json.Unmarshal(body, &envelope) != nil || envelope["contract_version"] != float64(1) || envelope["slug"] != slug || envelope["phase"] != "staged" {
		return false
	}
	stageRoot, ok := envelope["stage_root"].(string)
	if !ok || filepath.Base(stageRoot) != stageRoot {
		return false
	}
	info, err := os.Lstat(filepath.Join(root, "runtime", "skills", stageRoot, slug))
	return err == nil && info.IsDir()
}
func fatal(err error) { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
