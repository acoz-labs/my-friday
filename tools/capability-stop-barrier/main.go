// capability-stop-barrier externally interrupts an unmodified nominated
// candidate only after its durable capability journal and a post-mutation
// projection state are stable. It is acceptance tooling, not product behavior.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
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
)

type transaction struct {
	ContractVersion    int    `json:"contract_version"`
	Action             string `json:"action"`
	Slug               string `json:"slug"`
	SourceDigest       string `json:"source_digest"`
	ProjectionDigest   string `json:"projection_digest"`
	PriorReceiptDigest string `json:"prior_receipt_digest"`
	CreatedControl     bool   `json:"created_control"`
}

func main() {
	expect := flag.String("expect", "", "confirmation expect script")
	candidate := flag.String("candidate", "", "nominated candidate")
	instance := flag.String("instance", "", "named assistant")
	instanceRoot := flag.String("instance-root", "", "absolute named assistant root")
	slug := flag.String("slug", "", "capability slug")
	flag.Parse()
	if *expect == "" || *candidate == "" || *instance == "" || !filepath.IsAbs(*instanceRoot) || *slug == "" {
		fatal(errors.New("usage: capability-stop-barrier --expect E --candidate C --instance N --instance-root R --slug S"))
	}
	marker := filepath.Join(*instanceRoot, ".acceptance-capability-stop-"+*slug)
	if _, err := os.Lstat(marker); !errors.Is(err, os.ErrNotExist) {
		fatal(errors.New("capability stop-barrier marker already exists"))
	}
	defer os.Remove(marker)
	cmd := exec.Command("/usr/bin/expect", *expect, "Disable", *candidate, "capability", "disable", *instance, *slug)
	cmd.Env = append(os.Environ(), "MY_FRIDAY_CAPABILITY_BARRIER_MARKER="+marker)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		fatal(err)
	}
	pgid := cmd.Process.Pid
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	deadline := time.Now().Add(20 * time.Second)
	stoppedPID := 0
	for time.Now().Before(deadline) {
		select {
		case err := <-done:
			fatal(fmt.Errorf("candidate completed before interruption: %v", err))
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
			fatal(fmt.Errorf("resume stopped capability candidate: %w", err))
		}
		sliceUntil := time.Now().Add(100 * time.Microsecond)
		for time.Now().Before(sliceUntil) {
		}
		if err := syscall.Kill(stoppedPID, syscall.SIGSTOP); err != nil {
			select {
			case completed := <-done:
				fatal(fmt.Errorf("candidate completed during capability interruption race: %v", completed))
			default:
				fatal(fmt.Errorf("stop capability candidate: %w", err))
			}
		}
		time.Sleep(time.Millisecond)
		proof, ok := stablePostMutation(*instanceRoot, *slug)
		if ok {
			for attempt := 0; attempt < 3; attempt++ {
				time.Sleep(5 * time.Millisecond)
				if next, valid := stablePostMutation(*instanceRoot, *slug); !valid || next != proof {
					ok = false
					break
				}
			}
		}
		if ok {
			_ = syscall.Kill(stoppedPID, syscall.SIGKILL)
			_ = syscall.Kill(pgid, syscall.SIGKILL)
			<-done
			fmt.Println("captured real post-mutation capability interruption")
			return
		}
	}
	_ = syscall.Kill(stoppedPID, syscall.SIGKILL)
	_ = syscall.Kill(pgid, syscall.SIGKILL)
	<-done
	fatal(errors.New("no stable post-mutation capability interruption captured"))
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
		return 0, errors.New("invalid stopped-capability marker")
	}
	return pid, nil
}

type processRecord struct {
	pid, ppid, uid int
}

func ownedDescendant(root, target int) bool {
	body, err := exec.Command("/bin/ps", "-axo", "pid=,ppid=,uid=").Output()
	if err != nil {
		return false
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

func stablePostMutation(root, slug string) (string, bool) {
	control := filepath.Join(root, "capabilities", slug)
	receiptBody, err := os.ReadFile(filepath.Join(control, "receipt.json"))
	if err != nil {
		return "", false
	}
	journalBody, err := os.ReadFile(filepath.Join(control, "transaction.json"))
	if err != nil {
		return "", false
	}
	var r map[string]any
	var tx transaction
	if json.Unmarshal(receiptBody, &r) != nil || strict(journalBody, &tx) != nil || r["contract_version"] != float64(1) || tx.ContractVersion != 1 || tx.Action != "disable" || tx.CreatedControl || tx.Slug != slug || r["slug"] != slug || tx.SourceDigest != r["source_digest"] || tx.ProjectionDigest != r["projection_digest"] {
		return "", false
	}
	var compact bytes.Buffer
	if json.Compact(&compact, receiptBody) != nil {
		return "", false
	}
	current := sha256.Sum256(compact.Bytes())
	installedBytes := bytes.Replace(compact.Bytes(), []byte(`"state":"disabled"`), []byte(`"state":"installed-healthy"`), 1)
	installed := sha256.Sum256(installedBytes)
	if (tx.PriorReceiptDigest != fmt.Sprintf("%x", current) && tx.PriorReceiptDigest != fmt.Sprintf("%x", installed)) || !digest(tx.ProjectionDigest) {
		return "", false
	}
	projection := filepath.Join(root, "workspace", ".agents", "skills", slug)
	if _, err = os.Lstat(projection); err == nil || !errors.Is(err, os.ErrNotExist) {
		return "", false
	}
	body := append(append([]byte{}, receiptBody...), journalBody...)
	sum := sha256.Sum256(body)
	return fmt.Sprintf("%x", sum), true
}
func strict(body []byte, target any) error {
	d := json.NewDecoder(bytes.NewReader(body))
	d.DisallowUnknownFields()
	if err := d.Decode(target); err != nil {
		return err
	}
	var extra any
	if err := d.Decode(&extra); err != io.EOF {
		return errors.New("trailing JSON")
	}
	canonical, _ := json.MarshalIndent(target, "", "  ")
	canonical = append(canonical, '\n')
	if !bytes.Equal(body, canonical) {
		return errors.New("noncanonical JSON")
	}
	return nil
}
func digest(value string) bool {
	b, err := hex.DecodeString(value)
	return err == nil && len(b) == sha256.Size && value == strings.ToLower(value)
}
func fatal(err error) { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
