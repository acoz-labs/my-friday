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
	cmd := exec.Command("/usr/bin/expect", *expect, "Disable", *candidate, "capability", "disable", *instance, *slug)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		fatal(err)
	}
	pgid := cmd.Process.Pid
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		select {
		case err := <-done:
			fatal(fmt.Errorf("candidate completed before interruption: %w", err))
		default:
		}
		if err := syscall.Kill(-pgid, syscall.SIGSTOP); err != nil {
			fatal(err)
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
			_ = syscall.Kill(-pgid, syscall.SIGKILL)
			<-done
			for retry := 0; retry < 100 && groupPresent(pgid); retry++ {
				_ = syscall.Kill(-pgid, syscall.SIGKILL)
				time.Sleep(5 * time.Millisecond)
			}
			if groupPresent(pgid) {
				fatal(errors.New("interrupted candidate retained process-group members"))
			}
			fmt.Println("captured real post-mutation capability interruption")
			return
		}
		_ = syscall.Kill(-pgid, syscall.SIGCONT)
		time.Sleep(100 * time.Microsecond)
	}
	_ = syscall.Kill(-pgid, syscall.SIGCONT)
	_ = syscall.Kill(-pgid, syscall.SIGKILL)
	<-done
	fatal(errors.New("no stable post-mutation capability interruption captured"))
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
func groupPresent(pgid int) bool { return syscall.Kill(-pgid, 0) == nil }
func fatal(err error)            { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
