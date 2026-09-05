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
	marker := *transcript + ".candidate-stopped"
	if _, err := os.Lstat(marker); !errors.Is(err, os.ErrNotExist) {
		fatal(errors.New("source-workshop barrier marker already exists"))
	}
	defer os.Remove(marker)
	cmd := exec.Command("/usr/bin/expect", *expect, "enhance", *candidate, *instance, *slug, *transcript)
	cmd.Env = append(os.Environ(), "MY_FRIDAY_SOURCE_BARRIER_MARKER="+marker)
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
	stoppedPID := 0
	for time.Now().Before(deadline) {
		select {
		case err := <-done:
			fatal(fmt.Errorf("candidate completed before source interruption: %w", err))
		default:
		}
		if stoppedPID == 0 {
			var markerErr error
			stoppedPID, markerErr = readStoppedPID(marker, pgid)
			if markerErr != nil {
				fatal(markerErr)
			}
			if stoppedPID == 0 {
				time.Sleep(100 * time.Microsecond)
				continue
			}
		}
		if err := syscall.Kill(stoppedPID, syscall.SIGCONT); err != nil {
			fatal(fmt.Errorf("resume stopped candidate: %w", err))
		}
		sliceUntil := time.Now().Add(100 * time.Microsecond)
		for time.Now().Before(sliceUntil) {
		}
		if err := syscall.Kill(stoppedPID, syscall.SIGSTOP); err != nil {
			select {
			case completed := <-done:
				fatal(fmt.Errorf("candidate completed during source interruption race: %v", completed))
			default:
				fatal(fmt.Errorf("stop armed candidate: %w", err))
			}
		}
		time.Sleep(time.Millisecond)
		body, err := os.ReadFile(journalPath)
		if err == nil && stable(body, *root, *slug) {
			time.Sleep(5 * time.Millisecond)
			next, nextErr := os.ReadFile(journalPath)
			if nextErr == nil && bytes.Equal(body, next) && stable(next, *root, *slug) {
				_ = syscall.Kill(stoppedPID, syscall.SIGKILL)
				_ = syscall.Kill(pgid, syscall.SIGKILL)
				<-done
				fmt.Println("captured real source-workshop journal interruption")
				return
			}
		}
	}
	_ = syscall.Kill(stoppedPID, syscall.SIGKILL)
	_ = syscall.Kill(pgid, syscall.SIGKILL)
	<-done
	fatal(errors.New("no stable source-workshop interruption captured"))
}

func readStoppedPID(marker string, root int) (int, error) {
	body, err := os.ReadFile(marker)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return 0, nil
		}
		return 0, err
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(body)))
	if err != nil || !ownedDescendant(root, pid) {
		return 0, errors.New("invalid stopped-candidate marker")
	}
	return pid, nil
}

type processRecord struct {
	pid, ppid, uid int
}

func ownedDescendant(root, target int) bool {
	records := processTable()
	for pid := target; ; {
		record, ok := records[pid]
		if !ok || record.uid != os.Geteuid() {
			return false
		}
		if pid == root {
			return target != root
		}
		pid = record.ppid
	}
}

func processTable() map[int]processRecord {
	body, err := exec.Command("/bin/ps", "-axo", "pid=,ppid=,uid=").Output()
	if err != nil {
		return nil
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
	return records
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
