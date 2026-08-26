// Package capabilityworkshop implements the deterministic, source-only
// instruction capability workshop. Proposals are ephemeral; only the existing
// capability package format is written.
package capabilityworkshop

import (
	"bufio"
	"bytes"
	cryptorand "crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"unicode/utf8"

	"github.com/acoz-labs/my-friday/internal/capability"
	"golang.org/x/sys/unix"
)

var forbidden = []string{"scripts", "dependencies", "network", "credentials", "background", "durable-data", "publishing"}
var mutationHook func(string) error
var semanticVersionPattern = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)

func semanticVersion(value string) bool { return semanticVersionPattern.MatchString(value) }

type example struct {
	Input  string
	Output []string
}
type proposal struct {
	Slug, Version, DisplayName, Summary, Purpose, Success, Failure string
	Triggers, NonTriggers, Inputs, Outputs, Facts                  []string
	Examples                                                       []example
	Body                                                           []byte
	InputsAnswered, OutputsAnswered, RetainBody                    bool
}

// Run conducts one workshop and returns without mutation unless the exact
// source action shown in the final review is confirmed.
func Run(instance, source, slug string, in io.Reader, out io.Writer) error {
	if err := capability.ValidateSlug(slug); err != nil {
		return err
	}
	status, inspectErr := capability.Inspect(instance, source)
	if status.State == capability.StateInterrupted {
		if err := recoverSource(instance, source, slug); err != nil {
			return err
		}
		status, inspectErr = capability.Inspect(instance, source)
		if inspectErr != nil {
			return inspectErr
		}
		fmt.Fprintf(out, "Source workshop recovered\nState: %s\n", status.State)
		return nil
	}
	if !allowed(status.State) {
		if inspectErr != nil {
			return fmt.Errorf("capability state %s is not workshop-actionable: %w; run capability inspect %s %s --plain", status.State, inspectErr, filepath.Base(instance), slug)
		}
		return fmt.Errorf("capability state %s is not workshop-actionable; run capability inspect %s %s --plain", status.State, filepath.Base(instance), slug)
	}
	creating := status.State == capability.StateAbsent
	p, err := initialProposal(slug, status.Package)
	if err != nil {
		return err
	}
	r := bufio.NewReader(in)
	fmt.Fprintf(out, "Capability workshop: %s\nMode: %s\nEnter q at any prompt to exit without changes. Enter b for the previous item or r to restart the current section. At a section boundary, b repeats that boundary.\n", slug, map[bool]string{true: "create", false: "enhance"}[creating])
	if !creating {
		fmt.Fprintln(out, "Existing instruction body: retained user-authored content")
		for {
			fmt.Fprintln(out, "Current: retain the exact existing SKILL body bytes")
			choice, stop, err := prompt(r, out, "Instruction body [retain/regenerate]", 0, true)
			if err != nil || stop {
				return finishNoChange(out, err)
			}
			if choice == "b" || choice == "r" {
				fmt.Fprintln(out, "Already at the instruction-section boundary; retrying.")
				continue
			}
			if choice != "" && choice != "retain" && choice != "regenerate" {
				fmt.Fprintln(out, "Invalid: instruction body choice must be retain or regenerate. Try again.")
				continue
			}
			if choice == "regenerate" {
				p.Body = nil
				p.RetainBody = false
			}
			break
		}
	}
	fields := []struct {
		label string
		max   int
		dst   *string
	}{
		{"Display name (1-200 UTF-8 bytes; create default shown)", 200, &p.DisplayName}, {"Summary (1-200 UTF-8 bytes)", 200, &p.Summary},
		{"Version (semantic x.y.z)", 32, &p.Version},
		{"Success behavior (1-1000 UTF-8 bytes)", 1000, &p.Success}, {"Failure behavior (1-1000 UTF-8 bytes)", 1000, &p.Failure},
	}
	if !p.RetainBody {
		fields = append(fields[:3], append([]struct {
			label string
			max   int
			dst   *string
		}{{"Purpose (1-1000 UTF-8 bytes)", 1000, &p.Purpose}}, fields[3:]...)...)
	}
	for i := 0; i < len(fields); i++ {
		for {
			if *fields[i].dst != "" {
				fmt.Fprintf(out, "Current/default: %s\n", *fields[i].dst)
			}
			v, stop, e := prompt(r, out, fields[i].label, fields[i].max, *fields[i].dst != "")
			if e != nil {
				if errors.Is(e, io.EOF) || errors.Is(e, io.ErrUnexpectedEOF) {
					return finishNoChange(out, e)
				}
				fmt.Fprintf(out, "Invalid: %v. Try again.\n", e)
				continue
			}
			if stop {
				return finishNoChange(out, nil)
			}
			if v == "b" {
				if i > 0 {
					i -= 2
				}
				break
			}
			if v == "r" {
				if i <= 2 {
					i = -1
				} else {
					i = 2
				}
				break
			}
			if v != "" {
				*fields[i].dst = v
			}
			if *fields[i].dst == "" {
				fmt.Fprintln(out, "Invalid: a consequential answer is required.")
				continue
			}
			if i == 2 && !semanticVersion(*fields[i].dst) {
				fmt.Fprintln(out, "Invalid: version must be semantic x.y.z.")
				continue
			}
			break
		}
	}
	lists := []struct {
		label string
		none  bool
		dst   *[]string
	}{
		{"Triggers", false, &p.Triggers}, {"Non-triggers", false, &p.NonTriggers}, {"Inputs", true, &p.Inputs}, {"Outputs", true, &p.Outputs}, {"Required facts", false, &p.Facts},
	}
	for i := 0; i < len(lists); i++ {
		f := lists[i]
		for {
			answered := len(*f.dst) > 0 || (f.label == "Inputs" && p.InputsAnswered) || (f.label == "Outputs" && p.OutputsAnswered)
			if answered {
				fmt.Fprintf(out, "Current (%d): %s\n", len(*f.dst), listSummary(*f.dst))
			}
			v, stop, e := prompt(r, out, f.label+" (1-16; separate entries with |"+map[bool]string{true: "; 'none' allowed", false: ""}[f.none]+")", 4096, answered)
			if e != nil {
				if errors.Is(e, io.EOF) || errors.Is(e, io.ErrUnexpectedEOF) {
					return finishNoChange(out, e)
				}
				fmt.Fprintf(out, "Invalid: %v. Try again.\n", e)
				continue
			}
			if stop {
				return finishNoChange(out, nil)
			}
			if v == "b" {
				if i > 0 {
					i -= 2
				}
				break
			}
			if v == "r" {
				if i <= 1 {
					i = -1
				} else if i <= 3 {
					i = 1
				} else {
					i = 3
				}
				break
			}
			if v != "" {
				entryMax := 256
				if f.label == "Required facts" {
					entryMax = 512
				}
				values, x := list(v, f.none, entryMax, 16)
				if x != nil {
					fmt.Fprintf(out, "Invalid: %v. Try again.\n", x)
					continue
				}
				*f.dst = values
				if f.label == "Inputs" {
					p.InputsAnswered = true
				}
				if f.label == "Outputs" {
					p.OutputsAnswered = true
				}
			}
			answered = len(*f.dst) > 0 || (f.label == "Inputs" && p.InputsAnswered) || (f.label == "Outputs" && p.OutputsAnswered)
			if !answered {
				fmt.Fprintln(out, "Invalid: a consequential answer is required.")
				continue
			}
			break
		}
	}
	for i, ex := range p.Examples {
		fmt.Fprintf(out, "Current example %d: input=%q; output=%s\n", i+1, ex.Input, listSummary(ex.Output))
	}
	baseExamples := append([]example(nil), p.Examples...)
	for {
		currentInput := ""
		currentOutput := []string(nil)
		if len(baseExamples) > 0 {
			currentInput, currentOutput = baseExamples[0].Input, baseExamples[0].Output
		}
		ex, stop, err := prompt(r, out, "Example 1 input (must contain a trigger; 1-512 UTF-8 bytes)", 512, currentInput != "")
		if err != nil || stop {
			return finishNoChange(out, err)
		}
		if ex == "r" || ex == "b" {
			continue
		}
		outs, stop, err := prompt(r, out, "Expected output fragments for example 1 (1-8; separate with |)", 2048, len(currentOutput) > 0)
		if err != nil || stop {
			return finishNoChange(out, err)
		}
		if outs == "r" || outs == "b" {
			continue
		}
		if ex == "" {
			ex = currentInput
		}
		ov := append([]string(nil), currentOutput...)
		if outs != "" {
			var e error
			ov, e = list(outs, false, 256, 8)
			if e != nil {
				fmt.Fprintf(out, "Invalid: %v. Restarting examples.\n", e)
				continue
			}
		}
		if ex == "" || len(ov) == 0 {
			fmt.Fprintln(out, "Invalid: a complete example is required.")
			continue
		}
		remaining := []example(nil)
		if len(baseExamples) > 1 {
			remaining = baseExamples[1:]
		}
		p.Examples = append([]example{{Input: ex, Output: ov}}, remaining...)
		for len(p.Examples) < 16 {
			more, stop, e := prompt(r, out, "Add another example? [yes/no] (Return keeps current suite)", 3, true)
			if e != nil || stop {
				return finishNoChange(out, e)
			}
			if more == "b" || more == "r" {
				break
			}
			if more == "" || more == "no" {
				break
			}
			if more != "yes" {
				fmt.Fprintln(out, "Invalid: enter yes or no. Try again.")
				continue
			}
			n := len(p.Examples) + 1
			ei, stop, e := prompt(r, out, fmt.Sprintf("Example %d input (1-512 UTF-8 bytes)", n), 512, false)
			if e != nil || stop {
				return finishNoChange(out, e)
			}
			eo, stop, e := prompt(r, out, fmt.Sprintf("Example %d expected output fragments (1-8; separate with |)", n), 2048, false)
			if e != nil || stop {
				return finishNoChange(out, e)
			}
			vals, e := list(eo, false, 256, 8)
			if ei == "" || e != nil {
				fmt.Fprintln(out, "Invalid: a complete bounded example is required. Try again.")
				continue
			}
			p.Examples = append(p.Examples, example{ei, vals})
		}
		break
	}
	if len(p.Examples) == 0 {
		return errors.New("an example is required")
	}
	files, err := render(p)
	if err != nil {
		return err
	}
	if err = validateRendered(slug, files); err != nil {
		return fmt.Errorf("proposal is invalid: %w", err)
	}
	action := "update"
	token := "Update source"
	if creating {
		action = "create"
		token = "Create source"
	}
	if err = preview(out, source, status, files, action); err != nil {
		return err
	}
	fmt.Fprintf(out, "Type %s to continue; Return exits: ", token)
	line, err := r.ReadString('\n')
	if err != nil {
		return finishNoChange(out, err)
	}
	if line != token+"\n" {
		return finishNoChange(out, nil)
	}
	if err = commit(instance, source, slug, status, files, action); err != nil {
		return err
	}
	post, postErr := capability.Inspect(instance, source)
	if postErr != nil {
		return postErr
	}
	if post.Package == nil {
		return errors.New("source write completed but package is not inspectable")
	}
	if err = capability.TestCases(*post.Package); err != nil {
		return err
	}
	fmt.Fprintf(out, "Source written\nState: %s\nSource digest: %s\nDeterministic cases: %d passed\nInstalled: %s\n", post.State, post.Package.SourceDigest, len(post.Package.Cases.Examples), installed(post))
	if post.State == capability.StateReady {
		fmt.Fprintf(out, "Next: my-friday capability install %s %s\n", filepath.Base(instance), slug)
	}
	if post.State == capability.StateSourceChanged {
		fmt.Fprintf(out, "Next: my-friday capability upgrade %s %s\n", filepath.Base(instance), slug)
	}
	if post.State == capability.StateDisabled {
		fmt.Fprintf(out, "Next: enable, then upgrade separately to activate changed source\n")
	}
	return nil
}

func allowed(s capability.State) bool {
	return s == capability.StateAbsent || s == capability.StateReady || s == capability.StateInstalledHealthy || s == capability.StateSourceChanged || s == capability.StateDisabled
}
func installed(s capability.Status) string {
	if s.Receipt == nil {
		return "no"
	}
	return "unchanged"
}
func finishNoChange(out io.Writer, err error) error {
	fmt.Fprintln(out, "No changes made")
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return nil
	}
	return err
}
func prompt(r *bufio.Reader, out io.Writer, label string, max int, retain bool) (string, bool, error) {
	if retain {
		fmt.Fprintf(out, "%s [Return retains]: ", label)
	} else {
		fmt.Fprintf(out, "%s: ", label)
	}
	s, e := r.ReadString('\n')
	if e != nil {
		return "", true, e
	}
	s = strings.TrimSuffix(s, "\n")
	s = strings.TrimSuffix(s, "\r")
	if s == "q" {
		return "", true, nil
	}
	if s == "b" {
		return "b", false, nil
	}
	if s == "r" {
		return "r", false, nil
	}
	s = strings.TrimSpace(s)
	if !utf8.ValidString(s) || strings.IndexFunc(s, func(r rune) bool { return r < 32 || r == 127 }) >= 0 {
		return "", false, errors.New("input must be UTF-8 without control characters")
	}
	if max > 0 && len([]byte(s)) > max {
		return "", false, fmt.Errorf("input exceeds %d UTF-8 bytes", max)
	}
	return s, false, nil
}
func list(s string, none bool, maxBytes, maxCount int) ([]string, error) {
	if none && s == "none" {
		return []string{}, nil
	}
	parts := strings.Split(s, "|")
	if len(parts) < 1 || len(parts) > maxCount {
		return nil, fmt.Errorf("entry count must be 1-%d", maxCount)
	}
	seen := map[string]bool{}
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
		if parts[i] == "" || len([]byte(parts[i])) > maxBytes {
			return nil, errors.New("entries must be non-empty and within byte bounds")
		}
		k := strings.ToLower(parts[i])
		if seen[k] {
			return nil, errors.New("duplicate entry")
		}
		seen[k] = true
	}
	return parts, nil
}

func listSummary(values []string) string {
	if len(values) == 0 {
		return "none"
	}
	parts := make([]string, len(values))
	for i, value := range values {
		parts[i] = fmt.Sprintf("%d=%q", i+1, value)
	}
	return strings.Join(parts, "; ")
}

func initialProposal(slug string, pkg *capability.Package) (proposal, error) {
	p := proposal{Slug: slug, Version: "0.1.0", DisplayName: strings.Title(strings.ReplaceAll(slug, "-", " "))}
	if pkg == nil {
		return p, nil
	}
	m := pkg.Manifest
	p.Version = m.Version
	p.DisplayName = m.DisplayName
	p.Summary = m.Summary
	p.Success = m.SuccessBehavior
	p.Failure = m.FailureBehavior
	p.Triggers = append([]string(nil), m.Triggers...)
	p.Inputs = append([]string(nil), m.Inputs...)
	p.Outputs = append([]string(nil), m.Outputs...)
	p.InputsAnswered, p.OutputsAnswered, p.RetainBody = true, true, true
	p.NonTriggers = append([]string(nil), pkg.Cases.NonTriggers...)
	p.Facts = append([]string(nil), pkg.Cases.RequiredFacts...)
	for _, e := range pkg.Cases.Examples {
		p.Examples = append(p.Examples, example{e.Input, append([]string(nil), e.OutputContains...)})
	}
	b := pkg.Projection["SKILL.md"]
	idx := bytes.Index(b, []byte("\n---\n"))
	if idx < 0 {
		return p, errors.New("existing SKILL.md frontmatter is invalid")
	}
	p.Body = append([]byte(nil), b[idx+5:]...)
	return p, nil
}
func render(p proposal) (map[string][]byte, error) {
	m := capability.Manifest{ContractVersion: 1, Slug: p.Slug, Version: p.Version, DisplayName: p.DisplayName, Summary: p.Summary, Profile: "instruction-only", CodexCompatibility: "skills-v1", Triggers: p.Triggers, Inputs: p.Inputs, Outputs: p.Outputs, SuccessBehavior: p.Success, FailureBehavior: p.Failure, Scripts: "none", Dependencies: "none", Network: "none", Credentials: "none", Background: "none", DurableData: "none", Publishing: "none"}
	c := capability.Cases{ContractVersion: 1, PositiveTriggers: append([]string(nil), p.Triggers...), NonTriggers: p.NonTriggers, RequiredFacts: p.Facts, ForbiddenEffects: append([]string(nil), forbidden...)}
	for _, e := range p.Examples {
		c.Examples = append(c.Examples, struct {
			Input          string   `json:"input"`
			OutputContains []string `json:"output_contains"`
		}{e.Input, e.Output})
	}
	mb, _ := json.MarshalIndent(m, "", "  ")
	cb, _ := json.MarshalIndent(c, "", "  ")
	mb = append(mb, '\n')
	cb = append(cb, '\n')
	body := p.Body
	if body == nil {
		body = []byte(fmt.Sprintf("\n# %s\n\n## Purpose\n\n%s\n\n## Inputs\n\n%s\n\n## Outputs\n\n%s\n\n## Success\n\n%s\n\n## Failure\n\n%s\n\n## Required facts\n\n- %s\n", p.DisplayName, p.Purpose, strings.Join(p.Inputs, ", "), strings.Join(p.Outputs, ", "), p.Success, p.Failure, strings.Join(p.Facts, "\n- ")))
	}
	skill := append([]byte(fmt.Sprintf("---\nname: %s\ndescription: %s\n---\n", p.Slug, strconv.Quote(p.Summary))), body...)
	if !p.RetainBody && (len(skill) == 0 || skill[len(skill)-1] != '\n') {
		skill = append(skill, '\n')
	}
	return map[string][]byte{"capability.json": mb, "skill/SKILL.md": skill, "tests/cases.json": cb}, nil
}
func validateRendered(slug string, files map[string][]byte) error {
	d, e := os.MkdirTemp("", "my-friday-workshop-")
	if e != nil {
		return e
	}
	root := filepath.Join(d, slug)
	defer func() {
		for _, name := range []string{"capability.json", "skill/SKILL.md", "tests/cases.json"} {
			_ = os.Remove(filepath.Join(root, filepath.FromSlash(name)))
		}
		_ = os.Remove(filepath.Join(root, "skill"))
		_ = os.Remove(filepath.Join(root, "tests"))
		_ = os.Remove(root)
		_ = os.Remove(d)
	}()
	for n, b := range files {
		p := filepath.Join(root, filepath.FromSlash(n))
		if e = os.MkdirAll(filepath.Dir(p), 0700); e != nil {
			return e
		}
		if e = os.WriteFile(p, b, 0600); e != nil {
			return e
		}
	}
	pkg, e := capability.Validate(root)
	if e != nil {
		return e
	}
	return capability.TestCases(pkg)
}
func digest(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func readOwnedRegular(path string) ([]byte, error) {
	fd, e := unix.Open(path, unix.O_RDONLY|unix.O_NOFOLLOW, 0)
	if e != nil {
		return nil, e
	}
	f := os.NewFile(uintptr(fd), path)
	defer f.Close()
	var before unix.Stat_t
	if e = unix.Fstat(fd, &before); e != nil || before.Uid != uint32(os.Getuid()) || before.Nlink != 1 || before.Mode&unix.S_IFMT != unix.S_IFREG {
		return nil, errors.New("unsafe source file")
	}
	body, e := io.ReadAll(f)
	if e != nil {
		return nil, e
	}
	var after unix.Stat_t
	if e = unix.Fstat(fd, &after); e != nil || before.Dev != after.Dev || before.Ino != after.Ino || before.Size != after.Size || before.Mode != after.Mode || before.Nlink != after.Nlink {
		return nil, errors.New("source file changed during review")
	}
	return body, nil
}

func preview(out io.Writer, source string, status capability.Status, files map[string][]byte, action string) error {
	names := []string{"capability.json", "skill/SKILL.md", "tests/cases.json"}
	for i, n := range names {
		fmt.Fprintf(out, "\n%d. %s\n%s", i+1, n, files[n])
	}
	fmt.Fprintln(out, "\nComplete source diff:")
	for _, n := range names {
		old, e := readOwnedRegular(filepath.Join(source, filepath.FromSlash(n)))
		if errors.Is(e, os.ErrNotExist) {
			old = nil
		} else if e != nil {
			return e
		}
		fmt.Fprintf(out, "--- a/%s\n+++ b/%s\n", n, n)
		writeDiffBytes(out, '-', old)
		writeDiffBytes(out, '+', files[n])
	}
	if status.Package != nil {
		var optional []string
		for _, n := range status.Package.Files {
			if n != "capability.json" && n != "skill/SKILL.md" && n != "tests/cases.json" {
				optional = append(optional, n)
			}
		}
		sort.Strings(optional)
		for _, n := range optional {
			b, e := readOwnedRegular(filepath.Join(source, filepath.FromSlash(n)))
			if e != nil {
				return e
			}
			fmt.Fprintf(out, "Unchanged opaque file: %s (%d bytes, sha256:%s)\n", n, len(b), digest(b))
		}
	}
	fmt.Fprintln(out, "Unresolved or invalid answers: none")
	post := "ready"
	if status.Receipt != nil {
		post = "source-changed"
	}
	if status.State == capability.StateDisabled {
		post = "disabled"
	}
	fmt.Fprintf(out, "Source action: %s\nInstalled: %s\nCurrent state: %s\nPost-write state: %s\n", action, installed(status), status.State, post)
	return nil
}

func writeDiffBytes(out io.Writer, prefix byte, body []byte) {
	if len(body) == 0 {
		return
	}
	lines := bytes.Split(body, []byte{'\n'})
	terminated := body[len(body)-1] == '\n'
	limit := len(lines)
	if terminated {
		limit--
	}
	for _, line := range lines[:limit] {
		fmt.Fprintf(out, "%c%s\n", prefix, line)
	}
	if !terminated {
		fmt.Fprintln(out, "\\ No newline at end of file")
	}
}

type journal struct {
	ContractVersion int                     `json:"contract_version"`
	Action          string                  `json:"action"`
	Slug            string                  `json:"slug"`
	OldDigest       string                  `json:"old_digest"`
	NewDigest       string                  `json:"new_digest"`
	Phase           string                  `json:"phase"`
	SourceInode     uint64                  `json:"source_inode"`
	StageInode      uint64                  `json:"stage_inode"`
	StageRoot       string                  `json:"stage_root"`
	StageTree       map[string]journalEntry `json:"stage_tree"`
	OldTree         map[string]journalEntry `json:"old_tree,omitempty"`
}

type journalEntry struct {
	Device    uint64 `json:"device"`
	Inode     uint64 `json:"inode"`
	Mode      uint32 `json:"mode"`
	Owner     uint32 `json:"owner"`
	Links     uint64 `json:"links"`
	SHA256    string `json:"sha256,omitempty"`
	Directory bool   `json:"directory"`
}

func journalTree(entries map[string]snapshotEntry) map[string]journalEntry {
	out := make(map[string]journalEntry, len(entries))
	for path, entry := range entries {
		out[path] = journalEntry{entry.dev, entry.ino, uint32(entry.mode.Perm()), entry.uid, entry.nlink, entry.digest, entry.dir}
	}
	return out
}

func snapshotAuthority(entries map[string]journalEntry) map[string]snapshotEntry {
	out := make(map[string]snapshotEntry, len(entries))
	for path, entry := range entries {
		out[path] = snapshotEntry{entry.Device, entry.Inode, os.FileMode(entry.Mode), entry.Owner, entry.Links, entry.SHA256, entry.Directory}
	}
	return out
}

func readJournal(path, slug string) (journal, error) {
	var j journal
	info, statErr := os.Lstat(path)
	if statErr != nil || info.Mode().Perm() != 0600 || !info.Mode().IsRegular() {
		return j, errors.New("source workshop recovery required")
	}
	b, err := readOwnedRegular(path)
	if err != nil {
		return j, err
	}
	if err = json.Unmarshal(b, &j); err != nil {
		return j, errors.New("source workshop recovery required")
	}
	canonical, _ := json.MarshalIndent(j, "", "  ")
	canonical = append(canonical, '\n')
	if !bytes.Equal(b, canonical) || j.ContractVersion != 1 || j.Slug != slug || (j.Action != "create" && j.Action != "update") || j.Phase != "staged" || len(j.NewDigest) != 64 || j.StageInode == 0 || !validStageRoot(slug, j.StageRoot) || len(j.StageTree) == 0 || (j.Action == "update" && (len(j.OldDigest) != 64 || j.SourceInode == 0 || len(j.OldTree) == 0)) || (j.Action == "create" && (j.SourceInode != 0 || len(j.OldTree) != 0)) {
		return j, errors.New("source workshop recovery required")
	}
	return j, nil
}

func validStageRoot(slug, name string) bool {
	prefix := "." + slug + ".workshop-new-"
	if !strings.HasPrefix(name, prefix) || len(name) != len(prefix)+32 {
		return false
	}
	_, err := hex.DecodeString(strings.TrimPrefix(name, prefix))
	return err == nil && filepath.Base(name) == name
}

func packageDigest(path, slug string) string {
	p, e := capability.ValidateForSlug(path, slug)
	if e != nil {
		return ""
	}
	if capability.TestCases(p) != nil {
		return ""
	}
	return p.SourceDigest
}

func inode(path string) uint64 {
	i, e := os.Lstat(path)
	if e != nil {
		return 0
	}
	s, ok := i.Sys().(*syscall.Stat_t)
	if !ok {
		return 0
	}
	return s.Ino
}

type snapshotEntry struct {
	dev, ino uint64
	mode     os.FileMode
	uid      uint32
	nlink    uint64
	digest   string
	dir      bool
}

func snapshotTree(root, slug string) (map[string]snapshotEntry, error) {
	pkg, e := capability.ValidateForSlug(root, slug)
	if e != nil {
		return nil, e
	}
	if e = capability.TestCases(pkg); e != nil {
		return nil, e
	}
	entries := map[string]snapshotEntry{}
	e = filepath.WalkDir(root, func(path string, de os.DirEntry, we error) error {
		if we != nil {
			return we
		}
		rel, _ := filepath.Rel(root, path)
		rel = filepath.ToSlash(rel)
		if rel == "." {
			rel = ""
		}
		i, x := os.Lstat(path)
		if x != nil {
			return x
		}
		st, ok := i.Sys().(*syscall.Stat_t)
		if !ok || st.Uid != uint32(os.Getuid()) || i.Mode()&os.ModeSymlink != 0 {
			return errors.New("unsafe source snapshot entry")
		}
		se := snapshotEntry{uint64(st.Dev), st.Ino, i.Mode().Perm(), st.Uid, uint64(st.Nlink), "", i.IsDir()}
		if !i.IsDir() {
			if !i.Mode().IsRegular() || st.Nlink != 1 {
				return errors.New("unsafe source snapshot file")
			}
			b, x := os.ReadFile(path)
			if x != nil {
				return x
			}
			se.digest = digest(b)
		}
		entries[rel] = se
		return nil
	})
	return entries, e
}

func sameStat(st *unix.Stat_t, w snapshotEntry) bool {
	return uint64(st.Dev) == w.dev && st.Ino == w.ino && os.FileMode(st.Mode).Perm() == w.mode && st.Uid == w.uid && uint64(st.Nlink) == w.nlink
}
func sameDirAfter(st *unix.Stat_t, w snapshotEntry) bool {
	return uint64(st.Dev) == w.dev && st.Ino == w.ino && os.FileMode(st.Mode).Perm() == w.mode && st.Uid == w.uid && st.Nlink >= 1 && uint64(st.Nlink) <= w.nlink
}

func readSnapshotFile(path string, w snapshotEntry) ([]byte, error) {
	fd, e := unix.Open(path, unix.O_RDONLY|unix.O_NOFOLLOW, 0)
	if e != nil {
		return nil, e
	}
	file := os.NewFile(uintptr(fd), path)
	defer file.Close()
	var before unix.Stat_t
	if e = unix.Fstat(fd, &before); e != nil || !sameStat(&before, w) || w.dir {
		return nil, errors.New("optional source identity changed")
	}
	body, e := io.ReadAll(file)
	if e != nil {
		return nil, e
	}
	var after unix.Stat_t
	if e = unix.Fstat(fd, &after); e != nil || !sameStat(&after, w) || digest(body) != w.digest {
		return nil, errors.New("optional source bytes changed")
	}
	return body, nil
}

func writeExclusiveCanonical(path string, body []byte) error {
	fd, e := unix.Open(path, unix.O_WRONLY|unix.O_CREAT|unix.O_EXCL|unix.O_NOFOLLOW, 0600)
	if e != nil {
		return e
	}
	f := os.NewFile(uintptr(fd), path)
	_, we := f.Write(body)
	if we == nil {
		we = f.Sync()
	}
	ce := f.Close()
	if we != nil {
		return we
	}
	return ce
}
func writeExclusiveMode(path string, body []byte, mode os.FileMode) error {
	fd, e := unix.Open(path, unix.O_WRONLY|unix.O_CREAT|unix.O_EXCL|unix.O_NOFOLLOW, uint32(mode.Perm()))
	if e != nil {
		return e
	}
	f := os.NewFile(uintptr(fd), path)
	_, we := f.Write(body)
	if we == nil {
		we = f.Sync()
	}
	ce := f.Close()
	if we != nil {
		return we
	}
	return ce
}

func createStage(parentFD int, slug string) (string, int, error) {
	for attempts := 0; attempts < 8; attempts++ {
		random := make([]byte, 16)
		if _, err := cryptorand.Read(random); err != nil {
			return "", -1, err
		}
		name := "." + slug + ".workshop-new-" + hex.EncodeToString(random)
		if err := unix.Mkdirat(parentFD, name, 0700); errors.Is(err, unix.EEXIST) {
			continue
		} else if err != nil {
			return "", -1, err
		}
		fd, err := unix.Openat(parentFD, name, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW, 0)
		if err != nil {
			return "", -1, err
		}
		return name, fd, nil
	}
	return "", -1, errors.New("source workshop could not allocate staging")
}

func openOrCreateDirAt(rootFD int, rel string, mode os.FileMode) (int, error) {
	fd, err := unix.Dup(rootFD)
	if err != nil {
		return -1, err
	}
	components := strings.Split(filepath.ToSlash(rel), "/")
	finalCreated := false
	for index, component := range components {
		if component == "" || component == "." || component == ".." {
			unix.Close(fd)
			return -1, errors.New("unsafe staging directory")
		}
		next, openErr := unix.Openat(fd, component, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW, 0)
		if errors.Is(openErr, unix.ENOENT) {
			if err = unix.Mkdirat(fd, component, uint32(mode.Perm())); err != nil {
				unix.Close(fd)
				return -1, err
			}
			next, openErr = unix.Openat(fd, component, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW, 0)
			if index == len(components)-1 {
				finalCreated = true
			}
		}
		unix.Close(fd)
		if openErr != nil {
			return -1, openErr
		}
		fd = next
	}
	if finalCreated {
		err = unix.Fchmod(fd, uint32(mode.Perm()))
	}
	if err != nil {
		unix.Close(fd)
		return -1, err
	}
	return fd, nil
}

func writeExclusiveAt(rootFD int, rel string, body []byte, mode os.FileMode) error {
	dir, base := filepath.ToSlash(filepath.Dir(rel)), filepath.Base(rel)
	fd := rootFD
	owned := false
	var err error
	if dir != "." {
		fd, err = openOrCreateDirAt(rootFD, dir, 0700)
		if err != nil {
			return err
		}
		owned = true
	}
	if owned {
		defer unix.Close(fd)
	}
	fileFD, err := unix.Openat(fd, base, unix.O_WRONLY|unix.O_CREAT|unix.O_EXCL|unix.O_NOFOLLOW, uint32(mode.Perm()))
	if err != nil {
		return err
	}
	file := os.NewFile(uintptr(fileFD), base)
	_, err = file.Write(body)
	if err == nil {
		err = file.Sync()
	}
	if err == nil {
		err = unix.Fchmod(fileFD, uint32(mode.Perm()))
	}
	closeErr := file.Close()
	if err != nil {
		return err
	}
	return closeErr
}
func syncDir(path string) error {
	f, e := os.Open(path)
	if e != nil {
		return e
	}
	defer f.Close()
	return f.Sync()
}
func syncTreeDirs(root string) error {
	var dirs []string
	e := filepath.WalkDir(root, func(path string, de os.DirEntry, we error) error {
		if we != nil {
			return we
		}
		if de.IsDir() {
			dirs = append(dirs, path)
		}
		return nil
	})
	if e != nil {
		return e
	}
	sort.Slice(dirs, func(i, j int) bool {
		return strings.Count(dirs[i], string(os.PathSeparator)) > strings.Count(dirs[j], string(os.PathSeparator))
	})
	for _, d := range dirs {
		if e = syncDir(d); e != nil {
			return e
		}
	}
	return nil
}

func renameNoReplace(from, to string) error {
	fromDir, toDir := filepath.Dir(from), filepath.Dir(to)
	ffd, e := unix.Open(fromDir, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW, 0)
	if e != nil {
		return e
	}
	defer unix.Close(ffd)
	tfd, e := unix.Open(toDir, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW, 0)
	if e != nil {
		return e
	}
	defer unix.Close(tfd)
	return renameNoReplaceAt(ffd, filepath.Base(from), tfd, filepath.Base(to))
}

func removeExactTree(root, slug string, want map[string]snapshotEntry, wantDigest string) error {
	if err := validateExactTree(root, slug, want, wantDigest); err != nil {
		return err
	}
	if mutationHook != nil {
		if err := mutationHook("cleanup-validated"); err != nil {
			return err
		}
	}
	if err := validateExactTree(root, slug, want, wantDigest); err != nil {
		return err
	}
	rfd, e := unix.Open(root, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW, 0)
	if e != nil {
		return e
	}
	defer unix.Close(rfd)
	var rs unix.Stat_t
	if e = unix.Fstat(rfd, &rs); e != nil || !sameStat(&rs, want[""]) {
		return errors.New("source cleanup root identity mismatch")
	}
	var walk func(int, string) error
	walk = func(fd int, rel string) error {
		dup, e := unix.Dup(fd)
		if e != nil {
			return e
		}
		f := os.NewFile(uintptr(dup), "snapshot")
		names, e := f.ReadDir(-1)
		f.Close()
		if e != nil {
			return e
		}
		sort.Slice(names, func(i, j int) bool { return names[i].Name() < names[j].Name() })
		for _, de := range names {
			child := de.Name()
			cr := child
			if rel != "" {
				cr = rel + "/" + child
			}
			w, ok := want[cr]
			if !ok {
				return fmt.Errorf("foreign source cleanup entry %s", cr)
			}
			flags := unix.O_RDONLY | unix.O_NOFOLLOW
			if w.dir {
				flags |= unix.O_DIRECTORY
			}
			cfd, e := unix.Openat(fd, child, flags, 0)
			if e != nil {
				return e
			}
			var before unix.Stat_t
			if e = unix.Fstat(cfd, &before); e != nil || !sameStat(&before, w) {
				unix.Close(cfd)
				return fmt.Errorf("source cleanup identity changed %s", cr)
			}
			if w.dir {
				if e = walk(cfd, cr); e != nil {
					unix.Close(cfd)
					return e
				}
				var after unix.Stat_t
				if e = unix.Fstat(cfd, &after); e != nil || !sameDirAfter(&after, w) {
					unix.Close(cfd)
					return fmt.Errorf("source cleanup directory changed %s", cr)
				}
				unix.Close(cfd)
				if e = unix.Unlinkat(fd, child, unix.AT_REMOVEDIR); e != nil {
					return e
				}
			} else {
				file := os.NewFile(uintptr(cfd), child)
				body, e := io.ReadAll(file)
				if e != nil {
					file.Close()
					return e
				}
				var after unix.Stat_t
				if e = unix.Fstat(cfd, &after); e != nil || !sameStat(&after, w) || digest(body) != w.digest {
					file.Close()
					return fmt.Errorf("source cleanup file changed %s", cr)
				}
				file.Close()
				if e = unix.Unlinkat(fd, child, 0); e != nil {
					return e
				}
			}
		}
		return nil
	}
	if e = walk(rfd, ""); e != nil {
		return e
	}
	var final unix.Stat_t
	if e = unix.Fstat(rfd, &final); e != nil || !sameDirAfter(&final, want[""]) {
		return errors.New("source cleanup root changed")
	}
	return unix.Rmdir(root)
}

func validateExactTree(root, slug string, want map[string]snapshotEntry, wantDigest string) error {
	pkg, packageErr := capability.ValidateForSlug(root, slug)
	if packageErr != nil || capability.TestCases(pkg) != nil || pkg.SourceDigest != wantDigest {
		return errors.New("source cleanup digest mismatch")
	}
	got, err := snapshotTree(root, slug)
	if err != nil {
		return err
	}
	if len(got) != len(want) {
		return errors.New("source cleanup entry set mismatch")
	}
	for path, expected := range want {
		actual, ok := got[path]
		if !ok || actual != expected {
			return fmt.Errorf("source cleanup authority changed %s", path)
		}
	}
	return nil
}

func recoverSource(instance, source, slug string) error {
	lock, e := os.Open(filepath.Join(instance, "capabilities"))
	if e != nil {
		return e
	}
	defer lock.Close()
	if e = unix.Flock(int(lock.Fd()), unix.LOCK_EX|unix.LOCK_NB); e != nil {
		return errors.New("capability mutation already in progress")
	}
	defer unix.Flock(int(lock.Fd()), unix.LOCK_UN)
	parent := filepath.Dir(source)
	old := filepath.Join(parent, "."+slug+".workshop-old")
	jp := filepath.Join(instance, "capabilities", ".workshop-"+slug+".json")
	j, e := readJournal(jp, slug)
	if e != nil {
		return e
	}
	stageRoot := filepath.Join(parent, j.StageRoot)
	stage := filepath.Join(stageRoot, slug)
	sd, od, nd := packageDigest(source, slug), packageDigest(old, slug), packageDigest(stage, slug)
	si, oi, ni := inode(source), inode(old), inode(stage)
	if nd != "" {
		if e = validateExactTree(stage, slug, snapshotAuthority(j.StageTree), j.NewDigest); e != nil {
			return e
		}
	}
	if od != "" && len(j.OldTree) > 0 {
		if e = validateExactTree(old, slug, snapshotAuthority(j.OldTree), j.OldDigest); e != nil {
			return e
		}
	}
	if j.Action == "create" {
		switch {
		case sd == j.NewDigest && si == j.StageInode:
			if e = os.Remove(stageRoot); e != nil && !errors.Is(e, os.ErrNotExist) {
				return e
			}
		case sd == "" && nd == j.NewDigest && ni == j.StageInode:
			if e = renameNoReplace(stage, source); e != nil {
				return e
			}
			_ = os.Remove(stageRoot)
		default:
			return errors.New("source workshop recovery required: create authority is ambiguous")
		}
	} else {
		switch {
		case sd == j.NewDigest && si == j.StageInode && od == j.OldDigest && oi == j.SourceInode:
			oldSnap := snapshotAuthority(j.OldTree)
			if e = removeExactTree(old, slug, oldSnap, j.OldDigest); e != nil {
				return e
			}
			if e = os.Remove(stageRoot); e != nil && !errors.Is(e, os.ErrNotExist) {
				return e
			}
		case sd == j.OldDigest && si == j.SourceInode && nd == j.NewDigest && ni == j.StageInode && od == "":
			if e = renameNoReplace(source, old); e != nil {
				return e
			}
			if e = renameNoReplace(stage, source); e != nil {
				_ = renameNoReplace(old, source)
				return e
			}
			_ = os.Remove(stageRoot)
			oldSnap := snapshotAuthority(j.OldTree)
			if e = removeExactTree(old, slug, oldSnap, j.OldDigest); e != nil {
				return e
			}
		case sd == "" && od == j.OldDigest && oi == j.SourceInode && nd == j.NewDigest && ni == j.StageInode:
			if e = renameNoReplace(stage, source); e != nil {
				return e
			}
			_ = os.Remove(stageRoot)
			oldSnap := snapshotAuthority(j.OldTree)
			if e = removeExactTree(old, slug, oldSnap, j.OldDigest); e != nil {
				return e
			}
		default:
			return errors.New("source workshop recovery required: update authority is ambiguous")
		}
	}
	if e = syncDir(parent); e != nil {
		return e
	}
	if e = os.Remove(jp); e != nil {
		return e
	}
	return syncDir(filepath.Dir(jp))
}

func commit(instance, source, slug string, before capability.Status, files map[string][]byte, action string) error {
	lock, e := os.Open(filepath.Join(instance, "capabilities"))
	if e != nil {
		return e
	}
	defer lock.Close()
	if e = unix.Flock(int(lock.Fd()), unix.LOCK_EX|unix.LOCK_NB); e != nil {
		return errors.New("capability mutation already in progress")
	}
	defer unix.Flock(int(lock.Fd()), unix.LOCK_UN)
	now, ie := capability.Inspect(instance, source)
	if ie != nil || now.State != before.State || ((now.Package == nil) != (before.Package == nil)) || (now.Package != nil && now.Package.SourceDigest != before.Package.SourceDigest) {
		return errors.New("stale workshop preview: source or control state changed")
	}
	var priorSnapshot map[string]snapshotEntry
	if before.Package != nil {
		priorSnapshot, e = snapshotTree(source, slug)
		if e != nil {
			return e
		}
		if priorSnapshot[""].ino == 0 {
			return errors.New("source snapshot authority missing")
		}
	}
	parent := filepath.Dir(source)
	old := filepath.Join(parent, "."+slug+".workshop-old")
	jp := filepath.Join(instance, "capabilities", ".workshop-"+slug+".json")
	for _, p := range []string{old, jp} {
		if _, x := os.Lstat(p); !errors.Is(x, os.ErrNotExist) {
			return errors.New("source workshop collision; recovery required")
		}
	}
	parentFD, e := unix.Open(parent, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW, 0)
	if e != nil {
		return e
	}
	defer unix.Close(parentFD)
	stageName, stageRootFD, e := createStage(parentFD, slug)
	if e != nil {
		return e
	}
	defer unix.Close(stageRootFD)
	stageRoot := filepath.Join(parent, stageName)
	stage := filepath.Join(stageRoot, slug)
	stageFD, e := openOrCreateDirAt(stageRootFD, slug, 0700)
	if e != nil {
		return e
	}
	defer unix.Close(stageFD)
	if mutationHook != nil {
		if e = mutationHook("stage-created"); e != nil {
			return e
		}
	}
	if before.Package != nil {
		var optionalDirs []string
		for n, se := range priorSnapshot {
			if se.dir && (strings.HasPrefix(n, "skill/references") || strings.HasPrefix(n, "skill/assets")) {
				optionalDirs = append(optionalDirs, n)
			}
		}
		sort.Slice(optionalDirs, func(i, j int) bool { return strings.Count(optionalDirs[i], "/") < strings.Count(optionalDirs[j], "/") })
		for _, n := range optionalDirs {
			dfd, dirErr := openOrCreateDirAt(stageFD, n, priorSnapshot[n].mode)
			if dirErr != nil {
				return dirErr
			}
			unix.Close(dfd)
		}
		for _, n := range before.Package.Files {
			if n == "capability.json" || n == "skill/SKILL.md" || n == "tests/cases.json" {
				continue
			}
			b, x := readSnapshotFile(filepath.Join(source, filepath.FromSlash(n)), priorSnapshot[n])
			if x != nil {
				return x
			}
			mode := priorSnapshot[n].mode
			if x = writeExclusiveAt(stageFD, n, b, mode); x != nil {
				return x
			}
		}
	}
	for n, b := range files {
		if e = writeExclusiveAt(stageFD, n, b, 0600); e != nil {
			return e
		}
	}
	if mutationHook != nil {
		if e = mutationHook("stage-written"); e != nil {
			return e
		}
	}
	if e = syncTreeDirs(stage); e != nil {
		return e
	}
	if mutationHook != nil {
		if e = mutationHook("stage-synced"); e != nil {
			return e
		}
	}
	pkg, e := capability.Validate(stage)
	if e != nil {
		return e
	}
	if e = capability.TestCases(pkg); e != nil {
		return e
	}
	if mutationHook != nil {
		if e = mutationHook("stage-validated"); e != nil {
			return e
		}
	}
	stageSnapshot, e := snapshotTree(stage, slug)
	if e != nil {
		return e
	}
	oldDigest := ""
	if before.Package != nil {
		oldDigest = before.Package.SourceDigest
	}
	j := journal{ContractVersion: 1, Action: action, Slug: slug, OldDigest: oldDigest, NewDigest: pkg.SourceDigest, Phase: "staged", StageInode: stageSnapshot[""].ino, StageRoot: stageName, StageTree: journalTree(stageSnapshot)}
	if before.Package != nil {
		j.SourceInode = priorSnapshot[""].ino
		j.OldTree = journalTree(priorSnapshot)
	}
	jb, _ := json.MarshalIndent(j, "", "  ")
	jb = append(jb, '\n')
	if e = writeExclusiveCanonical(jp, jb); e != nil {
		return e
	}
	if e = syncDir(filepath.Dir(jp)); e != nil {
		return e
	}
	if mutationHook != nil {
		if e = mutationHook("journal-written"); e != nil {
			return e
		}
	}
	if action == "create" {
		if e = renameNoReplace(stage, source); e != nil {
			return e
		}
		if e = syncDir(parent); e != nil {
			return e
		}
	} else {
		if e = renameNoReplace(source, old); e != nil {
			return e
		}
		if e = syncDir(parent); e != nil {
			return e
		}
		if e = renameNoReplace(stage, source); e != nil {
			_ = renameNoReplace(old, source)
			return e
		}
		if e = syncDir(parent); e != nil {
			return e
		}
		if mutationHook != nil {
			if e = mutationHook("source-promoted"); e != nil {
				return e
			}
		}
		if e = removeExactTree(old, slug, priorSnapshot, oldDigest); e != nil {
			return e
		}
		if e = syncDir(parent); e != nil {
			return e
		}
	}
	_ = os.Remove(stageRoot)
	if e = os.Remove(jp); e != nil {
		return e
	}
	return syncDir(filepath.Dir(jp))
}
