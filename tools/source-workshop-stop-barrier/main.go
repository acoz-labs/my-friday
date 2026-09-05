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
	"sort"
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
	armed := false
	for time.Now().Before(deadline) {
		select {
		case err := <-done:
			fatal(fmt.Errorf("candidate completed before source interruption: %w", err))
		default:
		}
		if !armed {
			armed = confirmationSent(*transcript)
			if !armed {
				time.Sleep(100 * time.Microsecond)
				continue
			}
		}
		stopped, err := stopOwnedTree(pgid)
		if err != nil {
			select {
			case completed := <-done:
				fatal(fmt.Errorf("candidate completed during source interruption race: %w", completed))
			default:
				fatal(fmt.Errorf("stop armed Expect process tree: %w", err))
			}
		}
		body, err := os.ReadFile(journalPath)
		if err == nil && stable(body, *root, *slug) {
			time.Sleep(5 * time.Millisecond)
			next, nextErr := os.ReadFile(journalPath)
			if nextErr == nil && bytes.Equal(body, next) && stable(next, *root, *slug) {
				signalPIDs(stopped, syscall.SIGKILL)
				<-done
				fmt.Println("captured real source-workshop journal interruption")
				return
			}
		}
		signalPIDs(stopped, syscall.SIGCONT)
		time.Sleep(100 * time.Microsecond)
	}
	_ = syscall.Kill(-pgid, syscall.SIGKILL)
	<-done
	fatal(errors.New("no stable source-workshop interruption captured"))
}

func confirmationSent(transcript string) bool {
	body, err := os.ReadFile(transcript)
	return err == nil && bytes.Contains(body, []byte("Type Update source to continue; Return exits: Update source"))
}

type processRecord struct {
	pid, ppid, uid int
}

func stopOwnedTree(root int) ([]int, error) {
	body, err := exec.Command("/bin/ps", "-axo", "pid=,ppid=,uid=").Output()
	if err != nil {
		return nil, err
	}
	records := map[int]processRecord{}
	for _, line := range strings.Split(string(body), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 3 {
			continue
		}
		pid, pidErr := strconv.Atoi(fields[0])
		ppid, ppidErr := strconv.Atoi(fields[1])
		uid, uidErr := strconv.Atoi(fields[2])
		if pidErr == nil && ppidErr == nil && uidErr == nil {
			records[pid] = processRecord{pid: pid, ppid: ppid, uid: uid}
		}
	}
	if _, ok := records[root]; !ok {
		return nil, os.ErrProcessDone
	}
	depth := map[int]int{root: 0}
	for changed := true; changed; {
		changed = false
		for pid, record := range records {
			parentDepth, parentOwned := depth[record.ppid]
			if !parentOwned {
				continue
			}
			if _, known := depth[pid]; !known {
				depth[pid] = parentDepth + 1
				changed = true
			}
		}
	}
	pids := make([]int, 0, len(depth))
	for pid := range depth {
		if records[pid].uid != os.Geteuid() {
			return nil, fmt.Errorf("owned process tree contains uid %d", records[pid].uid)
		}
		pids = append(pids, pid)
	}
	sort.Slice(pids, func(i, j int) bool { return depth[pids[i]] > depth[pids[j]] })
	for _, pid := range pids {
		if err := syscall.Kill(pid, syscall.SIGSTOP); err != nil {
			signalPIDs(pids, syscall.SIGCONT)
			return nil, err
		}
	}
	return pids, nil
}

func signalPIDs(pids []int, signal syscall.Signal) {
	for _, pid := range pids {
		_ = syscall.Kill(pid, signal)
	}
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
