// Package capabilityworkshop implements the deterministic, source-only
// instruction capability workshop. Proposals are ephemeral; only the existing
// capability package format is written.
package capabilityworkshop

import (
	"bufio"
	"bytes"
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
	fmt.Fprintf(out, "Capability workshop: %s\nMode: %s\nEnter q at any prompt to exit without changes. Enter b to go back or r to restart the current section.\n", slug, map[bool]string{true: "create", false: "enhance"}[creating])
	if !creating {
		fmt.Fprintln(out, "Existing instruction body: retained user-authored content")
		choice, stop, err := prompt(r, out, "Instruction body [retain/regenerate] (Return retains)", 0, true)
		if err != nil || stop {
			return finishNoChange(out, err)
		}
		if choice != "" && choice != "retain" && choice != "regenerate" {
			return errors.New("instruction body choice must be retain or regenerate")
		}
		if choice == "regenerate" {
			p.Body = nil
		}
	}
	fields := []struct {
		label string
		max   int
		dst   *string
	}{
		{"Display name (1-200 UTF-8 bytes; create default shown)", 200, &p.DisplayName}, {"Summary (1-200 UTF-8 bytes)", 200, &p.Summary},
		{"Version (semantic x.y.z; create default 0.1.0)", 32, &p.Version}, {"Purpose (1-1000 UTF-8 bytes)", 1000, &p.Purpose},
		{"Success behavior (1-1000 UTF-8 bytes)", 1000, &p.Success}, {"Failure behavior (1-1000 UTF-8 bytes)", 1000, &p.Failure},
	}
	for i := 0; i < len(fields); i++ {
		for {
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
			v, stop, e := prompt(r, out, f.label+" (1-16; separate entries with |"+map[bool]string{true: "; 'none' allowed", false: ""}[f.none]+")", 4096, len(*f.dst) > 0)
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
			}
			if len(*f.dst) == 0 {
				fmt.Fprintln(out, "Invalid: a consequential answer is required.")
				continue
			}
			break
		}
	}
	for {
		ex, stop, err := prompt(r, out, "Example input (must contain a trigger; 1-512 UTF-8 bytes)", 512, !creating)
		if err != nil || stop {
			return finishNoChange(out, err)
		}
		if ex == "r" {
			continue
		}
		outs, stop, err := prompt(r, out, "Expected output fragments (1-8; separate with |)", 2048, !creating)
		if err != nil || stop {
			return finishNoChange(out, err)
		}
		if outs == "r" {
			continue
		}
		if ex != "" || outs != "" {
			ov, e := list(outs, false, 256, 8)
			if e != nil {
				fmt.Fprintf(out, "Invalid: %v. Restarting examples.\n", e)
				continue
			}
			p.Examples = []example{{Input: ex, Output: ov}}
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
		return []string{"none"}, nil
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
	skill := append([]byte(fmt.Sprintf("---\nname: %s\ndescription: %s\n---\n", p.Slug, p.Summary)), body...)
	if len(skill) == 0 || skill[len(skill)-1] != '\n' {
		skill = append(skill, '\n')
	}
	return map[string][]byte{"capability.json": mb, "skill/SKILL.md": skill, "tests/cases.json": cb}, nil
}
func validateRendered(slug string, files map[string][]byte) error {
	d, e := os.MkdirTemp("", "my-friday-workshop-")
	if e != nil {
		return e
	}
	defer os.RemoveAll(d)
	root := filepath.Join(d, slug)
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
		for _, l := range strings.Split(string(old), "\n") {
			if l != "" {
				fmt.Fprintf(out, "-%s\n", l)
			}
		}
		for _, l := range strings.Split(string(files[n]), "\n") {
			if l != "" {
				fmt.Fprintf(out, "+%s\n", l)
			}
		}
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

type journal struct {
	ContractVersion int    `json:"contract_version"`
	Action          string `json:"action"`
	Slug            string `json:"slug"`
	OldDigest       string `json:"old_digest"`
	NewDigest       string `json:"new_digest"`
	Phase           string `json:"phase"`
	SourceInode     uint64 `json:"source_inode"`
	StageInode      uint64 `json:"stage_inode"`
}

func readJournal(path, slug string) (journal, error) {
	var j journal
	b, err := readOwnedRegular(path)
	if err != nil {
		return j, err
	}
	if err = json.Unmarshal(b, &j); err != nil {
		return j, errors.New("source workshop recovery required")
	}
	canonical, _ := json.MarshalIndent(j, "", "  ")
	canonical = append(canonical, '\n')
	if !bytes.Equal(b, canonical) || j.ContractVersion != 1 || j.Slug != slug || (j.Action != "create" && j.Action != "update") || j.Phase != "staged" || len(j.NewDigest) != 64 || j.StageInode == 0 || (j.Action == "update" && (len(j.OldDigest) != 64 || j.SourceInode == 0)) || (j.Action == "create" && j.SourceInode != 0) {
		return j, errors.New("source workshop recovery required")
	}
	return j, nil
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

func removeExactTree(root string, want map[string]snapshotEntry, wantDigest string) error {
	_ = wantDigest
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
	stageRoot := filepath.Join(parent, "."+slug+".workshop-new")
	stage := filepath.Join(stageRoot, slug)
	old := filepath.Join(parent, "."+slug+".workshop-old")
	jp := filepath.Join(instance, "capabilities", ".workshop-"+slug+".json")
	j, e := readJournal(jp, slug)
	if e != nil {
		return e
	}
	sd, od, nd := packageDigest(source, slug), packageDigest(old, slug), packageDigest(stage, slug)
	si, oi, ni := inode(source), inode(old), inode(stage)
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
			oldSnap, x := snapshotTree(old, slug)
			if x != nil {
				return x
			}
			if e = removeExactTree(old, oldSnap, j.OldDigest); e != nil {
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
			oldSnap, x := snapshotTree(old, slug)
			if x != nil {
				return x
			}
			if e = removeExactTree(old, oldSnap, j.OldDigest); e != nil {
				return e
			}
		case sd == "" && od == j.OldDigest && oi == j.SourceInode && nd == j.NewDigest && ni == j.StageInode:
			if e = renameNoReplace(stage, source); e != nil {
				return e
			}
			_ = os.Remove(stageRoot)
			oldSnap, x := snapshotTree(old, slug)
			if x != nil {
				return x
			}
			if e = removeExactTree(old, oldSnap, j.OldDigest); e != nil {
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
	stageRoot := filepath.Join(parent, "."+slug+".workshop-new")
	stage := filepath.Join(stageRoot, slug)
	old := filepath.Join(parent, "."+slug+".workshop-old")
	jp := filepath.Join(instance, "capabilities", ".workshop-"+slug+".json")
	for _, p := range []string{stageRoot, old, jp} {
		if _, x := os.Lstat(p); !errors.Is(x, os.ErrNotExist) {
			return errors.New("source workshop collision; recovery required")
		}
	}
	if e = os.MkdirAll(stage, 0700); e != nil {
		return e
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
			p := filepath.Join(stage, filepath.FromSlash(n))
			if e = os.MkdirAll(p, priorSnapshot[n].mode); e != nil {
				return e
			}
			if e = os.Chmod(p, priorSnapshot[n].mode); e != nil {
				return e
			}
		}
		for _, n := range before.Package.Files {
			if n == "capability.json" || n == "skill/SKILL.md" || n == "tests/cases.json" {
				continue
			}
			b, x := readSnapshotFile(filepath.Join(source, filepath.FromSlash(n)), priorSnapshot[n])
			if x != nil {
				return x
			}
			p := filepath.Join(stage, filepath.FromSlash(n))
			if x = os.MkdirAll(filepath.Dir(p), 0700); x != nil {
				return x
			}
			mode := priorSnapshot[n].mode
			if x = writeExclusiveMode(p, b, mode); x != nil {
				return x
			}
			if x = os.Chmod(p, mode); x != nil {
				return x
			}
		}
	}
	for n, b := range files {
		p := filepath.Join(stage, filepath.FromSlash(n))
		if e = os.MkdirAll(filepath.Dir(p), 0700); e != nil {
			return e
		}
		if e = writeExclusiveMode(p, b, 0600); e != nil {
			return e
		}
	}
	if e = syncTreeDirs(stage); e != nil {
		return e
	}
	pkg, e := capability.Validate(stage)
	if e != nil {
		return e
	}
	if e = capability.TestCases(pkg); e != nil {
		return e
	}
	stageSnapshot, e := snapshotTree(stage, slug)
	if e != nil {
		return e
	}
	oldDigest := ""
	if before.Package != nil {
		oldDigest = before.Package.SourceDigest
	}
	j := journal{ContractVersion: 1, Action: action, Slug: slug, OldDigest: oldDigest, NewDigest: pkg.SourceDigest, Phase: "staged", StageInode: stageSnapshot[""].ino}
	if before.Package != nil {
		j.SourceInode = priorSnapshot[""].ino
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
		if e = removeExactTree(old, priorSnapshot, oldDigest); e != nil {
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
