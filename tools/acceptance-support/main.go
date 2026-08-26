package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/acoz-labs/my-friday/internal/assistantinstance"
	"github.com/acoz-labs/my-friday/internal/codexhome"
	"github.com/acoz-labs/my-friday/internal/plan"
	"github.com/acoz-labs/my-friday/internal/profile"
	"github.com/acoz-labs/my-friday/internal/repository"
	"golang.org/x/sys/unix"
)

type stringList []string

var copyAuthHook func(string)

func (v *stringList) String() string     { return strings.Join(*v, ",") }
func (v *stringList) Set(s string) error { *v = append(*v, s); return nil }

type codexCleanupEntry struct {
	Path   string `json:"path"`
	Target string `json:"target,omitempty"`
	Device uint64 `json:"device"`
	Inode  uint64 `json:"inode"`
	Mode   uint32 `json:"mode"`
	UID    uint32 `json:"uid"`
	Nlink  uint64 `json:"nlink"`
	Size   int64  `json:"size"`
	MtimeS int64  `json:"mtime_s"`
	MtimeN int64  `json:"mtime_ns"`
}

type codexCleanupReceipt struct {
	Schema      string              `json:"schema"`
	Candidate   string              `json:"candidate"`
	RunID       string              `json:"run_id"`
	Name        string              `json:"name"`
	RootDevice  uint64              `json:"root_device"`
	RootInode   uint64              `json:"root_inode"`
	CodexDevice uint64              `json:"codex_device"`
	CodexInode  uint64              `json:"codex_inode"`
	Entries     []codexCleanupEntry `json:"entries"`
}

func decodeCodexCleanupReceipts(encoded []string, candidate, runID string) (map[string]codexCleanupReceipt, error) {
	receipts := make(map[string]codexCleanupReceipt, len(encoded))
	for _, value := range encoded {
		var receipt codexCleanupReceipt
		if json.Unmarshal([]byte(value), &receipt) != nil || validateReceiptShape(receipt, candidate, runID) != nil || receipts[receipt.Name].Name != "" {
			return nil, errors.New("invalid or duplicate generated Codex-state receipt")
		}
		receipts[receipt.Name] = receipt
	}
	return receipts, nil
}

func main() {
	if len(os.Args) < 2 {
		fatal("usage: acceptance-support <fixture|update|resolve-executable|render-profile|validate-builder-prompt-input|protected-content|secure-roots|cleanup-named>")
	}
	switch os.Args[1] {
	case "fixture":
		fs := flag.NewFlagSet("fixture", flag.ExitOnError)
		runtime := fs.String("runtime", "", "runtime path")
		memory := fs.String("memory", "", "memory path")
		token := fs.String("token", "", "instruction token")
		padding := fs.Bool("padding", false, "use maximum ordinary profile padding for interruption observation")
		_ = fs.Parse(os.Args[2:])
		purpose := "Return only the exact token " + *token
		if *padding {
			purpose += strings.Repeat(" safely", 24)
		}
		p, err := profile.New("Acceptance Assistant", "", purpose, "concise", "")
		if err != nil {
			fatal(err.Error())
		}
		pl, err := plan.Build(p, *runtime, *memory)
		if err != nil {
			fatal(err.Error())
		}
		if err = repository.Create(pl, pl.Targets.Runtime, pl.Targets.Memory); err != nil {
			fatal(err.Error())
		}
		if err = repository.ValidatePair(pl.Targets.Runtime, pl.Targets.Memory); err != nil {
			fatal(err.Error())
		}
	case "update":
		fs := flag.NewFlagSet("update", flag.ExitOnError)
		runtime := fs.String("runtime", "", "runtime path")
		token := fs.String("token", "", "new instruction token")
		_ = fs.Parse(os.Args[2:])
		path := filepath.Join(*runtime, "assistant", "profile.json")
		body, err := os.ReadFile(path)
		if err != nil {
			fatal(err.Error())
		}
		var p profile.Profile
		decoder := json.NewDecoder(bytes.NewReader(body))
		decoder.DisallowUnknownFields()
		if err = decoder.Decode(&p); err != nil {
			fatal(err.Error())
		}
		p.Identity.Purpose = "Return only the exact token " + *token
		if err = profile.Validate(p); err != nil {
			fatal(err.Error())
		}
		body, _ = json.MarshalIndent(p, "", "  ")
		body = append(body, '\n')
		if err = os.WriteFile(path, body, 0o600); err != nil {
			fatal(err.Error())
		}
		if err = validateRuntimeNoFollow(*runtime); err != nil {
			fatal(err.Error())
		}
	case "scheme-string":
		if len(os.Args) != 3 {
			fatal("usage: acceptance-support scheme-string VALUE")
		}
		fmt.Print(strconv.Quote(os.Args[2]))
	case "resolve-executable":
		if len(os.Args) != 3 {
			fatal("usage: acceptance-support resolve-executable PATH")
		}
		resolved, err := resolveExecutable(os.Args[2])
		if err != nil {
			fatal(err.Error())
		}
		fmt.Println(resolved)
	case "render-profile":
		fs := flag.NewFlagSet("render-profile", flag.ExitOnError)
		template := fs.String("template", "", "profile template")
		value := fs.String("value", "", "literal path value")
		_ = fs.Parse(os.Args[2:])
		body, err := os.ReadFile(*template)
		if err != nil {
			fatal(err.Error())
		}
		placeholder := []byte("@@VOLUME@@")
		if bytes.Count(body, placeholder) != 1 {
			fatal("profile template must contain exactly one volume placeholder")
		}
		body = bytes.Replace(body, placeholder, []byte(strconv.Quote(*value)), 1)
		if _, err = os.Stdout.Write(body); err != nil {
			fatal(err.Error())
		}
	case "validate-sandbox-diagnostic":
		fs := flag.NewFlagSet("validate-sandbox-diagnostic", flag.ExitOnError)
		version := fs.String("allowlist", "", "diagnostic allowlist version")
		_ = fs.Parse(os.Args[2:])
		body, err := io.ReadAll(io.LimitReader(os.Stdin, 16<<10))
		if err != nil || !validSandboxDiagnostic(*version, string(body)) {
			fatal("unreviewed sandbox diagnostic")
		}
	case "validate-builder-prompt-input":
		fs := flag.NewFlagSet("validate-builder-prompt-input", flag.ExitOnError)
		skillRoot := fs.String("skill-root", "", "exact model-visible skill root")
		skill := fs.String("skill", "", "literal skill name")
		promptSHA256 := fs.String("prompt-sha256", "", "exact submitted prompt digest")
		_ = fs.Parse(os.Args[2:])
		if err := validateBuilderPromptInput(os.Stdin, *skillRoot, *skill, *promptSHA256); err != nil {
			fatal(err.Error())
		}
	case "protected-content":
		fs := flag.NewFlagSet("protected-content", flag.ExitOnError)
		codex := fs.String("codex-home", "", "live Codex home")
		runtime := fs.String("runtime", "", "runtime projection")
		_ = fs.Parse(os.Args[2:])
		var entries []string
		status, err := codexhome.Inspect("", *codex)
		if err != nil {
			fatal(err.Error())
		}
		if status.State == codexhome.StateHealthy || status.State == codexhome.StateSourceDrift {
			for _, name := range []string{"AGENTS.md", ".my-friday/installed-baseline.json", ".my-friday/canonical-AGENTS.md", ".my-friday/previous-AGENTS.md"} {
				if digest, ok, err := noFollowDigest(filepath.Join(*codex, name)); err != nil {
					fatal(err.Error())
				} else if ok {
					entries = append(entries, name+":"+digest)
				}
			}
		}
		if err = validateRuntimeNoFollow(*runtime); err != nil {
			fatal(err.Error())
		}
		for _, name := range []string{"AGENTS.md", "assistant/profile.json", ".my-friday/manifest.json"} {
			digest, ok, err := noFollowDigest(filepath.Join(*runtime, name))
			if err != nil || !ok {
				if err != nil {
					fatal(err.Error())
				}
				fatal("missing protected runtime file")
			}
			entries = append(entries, "runtime/"+name+":"+digest)
		}
		statusAfter, err := codexhome.Inspect("", *codex)
		if err != nil || statusAfter != status {
			fatal("protected Codex authority changed during snapshot")
		}
		sort.Strings(entries)
		sum := sha256.Sum256([]byte(strings.Join(entries, "\n") + "\n"))
		body, _ := json.Marshal(map[string]any{"digest": fmt.Sprintf("%x", sum), "count": len(entries)})
		fmt.Println(string(body))
	case "validate-root":
		fs := flag.NewFlagSet("validate-root", flag.ExitOnError)
		path := fs.String("path", "", "directory path")
		ownerMode := fs.Uint("mode", 0, "required owner-only mode; zero permits any mode")
		_ = fs.Parse(os.Args[2:])
		fd, err := openAbsoluteDirNoFollow(*path)
		if err != nil {
			fatal(err.Error())
		}
		defer unix.Close(fd)
		var st unix.Stat_t
		if unix.Fstat(fd, &st) != nil || st.Uid != uint32(os.Getuid()) || (*ownerMode != 0 && uint32(st.Mode)&0o777 != uint32(*ownerMode)) {
			fatal("unsafe directory identity")
		}
		fmt.Printf("%d:%d\n", st.Dev, st.Ino)
	case "mounted-device":
		fs := flag.NewFlagSet("mounted-device", flag.ExitOnError)
		mountpoint := fs.String("mountpoint", "", "exact requested mount point")
		_ = fs.Parse(os.Args[2:])
		device, err := mountedDeviceFromPlist(os.Stdin, *mountpoint)
		if err != nil {
			fatal(err.Error())
		}
		fmt.Println(device)
	case "validate-runtime":
		if len(os.Args) != 3 {
			fatal("usage: acceptance-support validate-runtime PATH")
		}
		fd, err := openAbsoluteDirNoFollow(os.Args[2])
		if err != nil {
			fatal(err.Error())
		}
		unix.Close(fd)
		if err = validateRuntimeNoFollow(os.Args[2]); err != nil {
			fatal(err.Error())
		}
	case "copy-auth":
		fs := flag.NewFlagSet("copy-auth", flag.ExitOnError)
		source := fs.String("source", "", "absolute source auth.json")
		destination := fs.String("destination-dir", "", "absolute destination Codex directory")
		_ = fs.Parse(os.Args[2:])
		receipt, err := copyAuthNoFollow(*source, *destination)
		if err != nil {
			fatal(err.Error())
		}
		body, _ := json.Marshal(receipt)
		fmt.Println(string(body))
	case "protected-metadata":
		if len(os.Args) != 4 || (os.Args[3] != "codex" && os.Args[3] != "runtime") {
			fatal("usage: acceptance-support protected-metadata ROOT codex|runtime")
		}
		fd, err := openAbsoluteDirNoFollow(os.Args[2])
		if err != nil {
			fatal(err.Error())
		}
		defer unix.Close(fd)
		var records []string
		var rootStat unix.Stat_t
		if unix.Fstat(fd, &rootStat) != nil {
			fatal("protected root identity unavailable")
		}
		records = append(records, fmt.Sprintf("%s|%d|%d|%d|%d|%d|%d|%o|%d|%d|%d", os.Args[3], rootStat.Mode&unix.S_IFMT, rootStat.Dev, rootStat.Ino, rootStat.Nlink, rootStat.Uid, rootStat.Gid, rootStat.Mode&0o7777, rootStat.Size, rootStat.Mtim.Sec, rootStat.Ctim.Sec))
		if err = walkMetadata(fd, os.Args[3], &records); err != nil {
			fatal(err.Error())
		}
		sort.Strings(records)
		for _, record := range records {
			fmt.Println(record)
		}
	case "secure-roots":
		fs := flag.NewFlagSet("secure-roots", flag.ExitOnError)
		home := fs.String("home", "", "canonical home")
		runID := fs.String("run-id", "", "run identifier")
		_ = fs.Parse(os.Args[2:])
		result, err := secureRoots(*home, *runID)
		if err != nil {
			fatal(err.Error())
		}
		body, _ := json.Marshal(result)
		fmt.Println(string(body))
	case "cleanup-roots":
		fs := flag.NewFlagSet("cleanup-roots", flag.ExitOnError)
		home := fs.String("home", "", "canonical home")
		runID := fs.String("run-id", "", "run identifier")
		receiptJSON := fs.String("receipt", "", "secure-roots receipt")
		markerSHA := fs.String("marker-sha256", "", "expected marker digest")
		var expectedEntries stringList
		fs.Var(&expectedEntries, "expected-entry", "PARENT:relative/path expected cleanup entry")
		_ = fs.Parse(os.Args[2:])
		var receipt map[string]map[string]uint64
		decodedSHA, shaErr := hex.DecodeString(*markerSHA)
		if json.Unmarshal([]byte(*receiptJSON), &receipt) != nil || len(receipt) != 2 || len(decodedSHA) != 32 || shaErr != nil || *markerSHA != strings.ToLower(*markerSHA) || *runID == "" || *runID != filepath.Base(*runID) || strings.ContainsAny(*runID, "/\\") {
			fatal("invalid cleanup authority")
		}
		hfd, err := openAbsoluteDirNoFollow(*home)
		if err != nil {
			fatal(err.Error())
		}
		if err = cleanupRoots(hfd, *runID, *markerSHA, receipt, expectedEntries); err != nil {
			unix.Close(hfd)
			fatal(err.Error())
		}
		unix.Close(hfd)
	case "cleanup-named":
		fs := flag.NewFlagSet("cleanup-named", flag.ExitOnError)
		home := fs.String("home", "", "account home")
		candidate := fs.String("candidate", "", "exact candidate SHA")
		runID := fs.String("run-id", "", "acceptance run identifier")
		var names, leafReceipts, codexReceipts stringList
		fs.Var(&names, "name", "manifest-owned instance name")
		fs.Var(&leafReceipts, "leaf-receipt", "JSON exact-leaf cleanup receipt")
		fs.Var(&codexReceipts, "codex-receipt", "JSON exact generated Codex-state receipt")
		_ = fs.Parse(os.Args[2:])
		var leaves []exactLeaf
		for _, encoded := range leafReceipts {
			var leaf exactLeaf
			if json.Unmarshal([]byte(encoded), &leaf) != nil {
				fatal("invalid exact-leaf receipt")
			}
			leaves = append(leaves, leaf)
		}
		receipts, receiptErr := decodeCodexCleanupReceipts(codexReceipts, *candidate, *runID)
		if receiptErr != nil {
			fatal(receiptErr.Error())
		}
		if err := cleanupNamed(*home, names, leaves, receipts, *candidate, *runID); err != nil {
			fatal(err.Error())
		}
	case "codex-cleanup-receipt":
		fs := flag.NewFlagSet("codex-cleanup-receipt", flag.ExitOnError)
		home := fs.String("home", "", "account home")
		name := fs.String("name", "", "manifest-owned instance name")
		candidate := fs.String("candidate", "", "exact candidate SHA")
		runID := fs.String("run-id", "", "acceptance run identifier")
		_ = fs.Parse(os.Args[2:])
		receipt, err := captureCodexCleanupReceipt(*home, *name, *candidate, *runID)
		if err != nil {
			fatal(err.Error())
		}
		body, _ := json.Marshal(receipt)
		fmt.Println(string(body))
	case "setsid-probe":
		executable, err := os.Executable()
		if err != nil {
			fatal(err.Error())
		}
		cmd := exec.Command(executable, "setsid-child")
		cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
		if err = cmd.Start(); err != nil {
			fatal(err.Error())
		}
		time.Sleep(500 * time.Millisecond)
		_ = cmd.Wait()
	case "setsid-child":
		time.Sleep(30 * time.Second)
	default:
		fatal("unknown acceptance-support command")
	}
}

func mountedDeviceFromPlist(input io.Reader, mountpoint string) (string, error) {
	if !filepath.IsAbs(mountpoint) {
		return "", errors.New("mount point must be absolute")
	}
	decoder := xml.NewDecoder(input)
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			return "", errors.New("hdiutil plist has no system-entities array")
		}
		if err != nil {
			return "", fmt.Errorf("decode hdiutil plist: %w", err)
		}
		start, ok := token.(xml.StartElement)
		if !ok || start.Name.Local != "key" {
			continue
		}
		var key string
		if err = decoder.DecodeElement(&key, &start); err != nil {
			return "", fmt.Errorf("decode hdiutil plist key: %w", err)
		}
		if key != "system-entities" {
			continue
		}
		for {
			token, err = decoder.Token()
			if err != nil {
				return "", errors.New("hdiutil system-entities value is missing")
			}
			if array, ok := token.(xml.StartElement); ok {
				if array.Name.Local != "array" {
					return "", errors.New("hdiutil system-entities is not an array")
				}
				return mountedDeviceFromEntities(decoder, array, mountpoint)
			}
		}
	}
}

func mountedDeviceFromEntities(decoder *xml.Decoder, array xml.StartElement, mountpoint string) (string, error) {
	match := ""
	for {
		token, err := decoder.Token()
		if err != nil {
			return "", errors.New("invalid hdiutil system-entities array")
		}
		switch value := token.(type) {
		case xml.StartElement:
			if value.Name.Local != "dict" {
				if err = decoder.Skip(); err != nil {
					return "", err
				}
				continue
			}
			entity, err := plistStringDict(decoder, value)
			if err != nil {
				return "", err
			}
			if entity["mount-point"] == mountpoint {
				device := entity["dev-entry"]
				if device == "" || match != "" {
					return "", errors.New("mounted hdiutil entity is missing or ambiguous")
				}
				match = device
			}
		case xml.EndElement:
			if value.Name == array.Name {
				if match == "" {
					return "", errors.New("requested mount point is absent from hdiutil entities")
				}
				return match, nil
			}
		}
	}
}

func plistStringDict(decoder *xml.Decoder, dict xml.StartElement) (map[string]string, error) {
	result := map[string]string{}
	key := ""
	for {
		token, err := decoder.Token()
		if err != nil {
			return nil, errors.New("invalid hdiutil entity dictionary")
		}
		switch value := token.(type) {
		case xml.StartElement:
			switch value.Name.Local {
			case "key":
				if err = decoder.DecodeElement(&key, &value); err != nil {
					return nil, err
				}
			case "string":
				var text string
				if err = decoder.DecodeElement(&text, &value); err != nil {
					return nil, err
				}
				if key != "" {
					result[key] = text
					key = ""
				}
			default:
				if err = decoder.Skip(); err != nil {
					return nil, err
				}
				key = ""
			}
		case xml.EndElement:
			if value.Name == dict.Name {
				return result, nil
			}
		}
	}
}

func copyAuthNoFollow(source, destinationDir string) (map[string]any, error) {
	if !filepath.IsAbs(source) || filepath.Base(source) != "auth.json" || !filepath.IsAbs(destinationDir) {
		return nil, errors.New("auth copy requires absolute auth.json and destination directory")
	}
	sourceParent, err := openAbsoluteDirNoFollow(filepath.Dir(source))
	if err != nil {
		return nil, err
	}
	defer unix.Close(sourceParent)
	sourceFD, err := unix.Openat(sourceParent, "auth.json", unix.O_RDONLY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if err != nil {
		return nil, err
	}
	defer unix.Close(sourceFD)
	var opened, entry unix.Stat_t
	if unix.Fstat(sourceFD, &opened) != nil || unix.Fstatat(sourceParent, "auth.json", &entry, unix.AT_SYMLINK_NOFOLLOW) != nil ||
		opened.Dev != entry.Dev || opened.Ino != entry.Ino || opened.Mode&unix.S_IFMT != unix.S_IFREG || opened.Uid != uint32(os.Getuid()) || opened.Mode&0o777 != 0o600 || opened.Nlink != 1 {
		return nil, errors.New("auth source identity or metadata is unsafe")
	}
	if copyAuthHook != nil {
		copyAuthHook("source-opened")
	}
	destination, err := openAbsoluteDirNoFollow(destinationDir)
	if err != nil {
		return nil, err
	}
	defer unix.Close(destination)
	destinationFD, err := unix.Openat(destination, "auth.json", unix.O_WRONLY|unix.O_CREAT|unix.O_EXCL|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0o600)
	if err != nil {
		return nil, err
	}
	destinationFile := os.NewFile(uintptr(destinationFD), "auth-copy")
	sourceFile := os.NewFile(uintptr(sourceFD), "auth-source")
	h := sha256.New()
	_, copyErr := io.Copy(io.MultiWriter(destinationFile, h), sourceFile)
	sourceCloseErr := sourceFile.Close()
	sourceFD = -1
	syncErr := destinationFile.Sync()
	closeErr := destinationFile.Close()
	if copyErr != nil || sourceCloseErr != nil || syncErr != nil || closeErr != nil {
		_ = unix.Unlinkat(destination, "auth.json", 0)
		return nil, errors.Join(copyErr, sourceCloseErr, syncErr, closeErr)
	}
	if copyAuthHook != nil {
		copyAuthHook("destination-synced")
	}
	var sourceAfter, destinationEntry unix.Stat_t
	if unix.Fstatat(sourceParent, "auth.json", &sourceAfter, unix.AT_SYMLINK_NOFOLLOW) != nil || sourceAfter.Dev != opened.Dev || sourceAfter.Ino != opened.Ino || sourceAfter.Mtim != opened.Mtim || sourceAfter.Size != opened.Size ||
		unix.Fstatat(destination, "auth.json", &destinationEntry, unix.AT_SYMLINK_NOFOLLOW) != nil || destinationEntry.Mode&unix.S_IFMT != unix.S_IFREG || destinationEntry.Uid != uint32(os.Getuid()) || destinationEntry.Mode&0o777 != 0o600 || destinationEntry.Nlink != 1 {
		_ = unix.Unlinkat(destination, "auth.json", 0)
		return nil, errors.New("auth source changed or destination is unsafe")
	}
	return map[string]any{"schema": "auth-copy-receipt-v1", "source_device": uint64(opened.Dev), "source_inode": opened.Ino, "destination_device": uint64(destinationEntry.Dev), "destination_inode": destinationEntry.Ino, "sha256": fmt.Sprintf("%x", h.Sum(nil))}, nil
}

type exactLeaf struct {
	Path   string `json:"path"`
	Device uint64 `json:"device"`
	Inode  uint64 `json:"inode"`
	SHA256 string `json:"sha256"`
}

func inspectExactLeaf(leaf exactLeaf) error {
	if !filepath.IsAbs(leaf.Path) || len(leaf.SHA256) != 64 {
		return errors.New("invalid exact-leaf authority")
	}
	info, err := os.Lstat(leaf.Path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm() != 0o700 || info.Sys().(*syscall.Stat_t).Uid != uint32(os.Getuid()) || info.Sys().(*syscall.Stat_t).Nlink != 1 {
		return errors.New("exact cleanup leaf is unsafe")
	}
	st := info.Sys().(*syscall.Stat_t)
	body, ok, err := noFollowRead(leaf.Path)
	if err != nil || uint64(st.Dev) != leaf.Device || st.Ino != leaf.Inode || fmt.Sprintf("%x", sha256.Sum256(body)) != leaf.SHA256 {
		return errors.New("exact cleanup leaf identity changed")
	}
	if !ok {
		return nil
	}
	return nil
}

func removeExactLeaf(leaf exactLeaf) error {
	parentFD, err := openAbsoluteDirNoFollow(filepath.Dir(leaf.Path))
	if err != nil {
		return err
	}
	defer unix.Close(parentFD)
	fd, err := unix.Openat(parentFD, filepath.Base(leaf.Path), unix.O_RDONLY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if errors.Is(err, unix.ENOENT) {
		return nil
	}
	if err != nil {
		return err
	}
	defer unix.Close(fd)
	var opened, entry unix.Stat_t
	if unix.Fstat(fd, &opened) != nil || unix.Fstatat(parentFD, filepath.Base(leaf.Path), &entry, unix.AT_SYMLINK_NOFOLLOW) != nil || opened.Dev != entry.Dev || opened.Ino != entry.Ino || uint64(opened.Dev) != leaf.Device || opened.Ino != leaf.Inode {
		return errors.New("exact cleanup leaf changed before unlink")
	}
	return unix.Unlinkat(parentFD, filepath.Base(leaf.Path), 0)
}

func disposableAuthStatSafe(mode uint32, uid uint32, nlink uint64) bool {
	return mode&unix.S_IFMT == unix.S_IFREG && mode&0o777 == 0o600 && uid == uint32(os.Getuid()) && nlink == 1
}

func openedDirectoryMatchesPath(fd int, path string) bool {
	var opened unix.Stat_t
	info, err := os.Lstat(path)
	if unix.Fstat(fd, &opened) != nil || err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return false
	}
	entry := info.Sys().(*syscall.Stat_t)
	return uint64(opened.Dev) == uint64(entry.Dev) && opened.Ino == entry.Ino
}

func verifyCodexCleanupEntriesAt(fd int, codexRoot string, manifest assistantinstance.Manifest, allowedExtras ...string) error {
	duplicate, err := unix.Openat(fd, ".", unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if err != nil {
		return err
	}
	dir := os.NewFile(uintptr(duplicate), codexRoot)
	defer dir.Close()
	entries, err := dir.Readdirnames(-1)
	if err != nil {
		return err
	}
	allowed := map[string]bool{"config.toml": true}
	if manifest.CodexInstructions != "" {
		allowed["AGENTS.md"] = true
	}
	for _, extra := range allowedExtras {
		allowed[extra] = true
	}
	seen := make(map[string]bool, len(entries))
	for _, entry := range entries {
		if !allowed[entry] || seen[entry] {
			return fmt.Errorf("unexpected disposable Codex entry preserved: %s", entry)
		}
		seen[entry] = true
	}
	for expected := range allowed {
		if !seen[expected] {
			return fmt.Errorf("required disposable Codex entry missing: %s", expected)
		}
	}
	return nil
}

func verifyRootCleanupEntriesAt(fd int, root string, manifest assistantinstance.Manifest, allowedExtras ...string) error {
	duplicate, err := unix.Openat(fd, ".", unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if err != nil {
		return err
	}
	dir := os.NewFile(uintptr(duplicate), root)
	defer dir.Close()
	entries, err := dir.Readdirnames(-1)
	if err != nil {
		return err
	}
	allowed := map[string]bool{"manifest.json": true}
	for _, owned := range manifest.Owned {
		allowed[owned] = true
	}
	for _, extra := range allowedExtras {
		allowed[extra] = true
	}
	seen := make(map[string]bool, len(entries))
	for _, entry := range entries {
		if !allowed[entry] || seen[entry] {
			return fmt.Errorf("unexpected disposable instance-root entry preserved: %s", entry)
		}
		seen[entry] = true
	}
	for expected := range allowed {
		if !seen[expected] {
			return fmt.Errorf("required disposable instance-root entry missing: %s", expected)
		}
	}
	return nil
}

var cleanupMutationHook func(string)

const authQuarantinePrefix = ".auth.json.my-friday-cleanup-"

func validCodexArg0HelperSymlink(path string) bool {
	parts := strings.Split(path, "/")
	if len(parts) != 4 || parts[0] != "tmp" || parts[1] != "arg0" || !strings.HasPrefix(parts[2], "codex-") || len(parts[2]) == len("codex-") {
		return false
	}
	for _, character := range parts[2][len("codex-"):] {
		if (character < 'a' || character > 'z') && (character < 'A' || character > 'Z') && (character < '0' || character > '9') {
			return false
		}
	}
	switch parts[3] {
	case "apply_patch", "applypatch", "codex-execve-wrapper":
		return true
	default:
		return false
	}
}

func validGeneratedRegularMode(mode uint32) bool {
	switch mode {
	case 0o600, 0o644, 0o664, 0o755:
		return true
	default:
		return false
	}
}

func volatileAmbientCodexMetadata(path string) bool {
	if path == "codex/sessions" || strings.HasPrefix(path, "codex/sessions/") {
		return true
	}
	if strings.Count(path, "/") != 1 || !strings.HasPrefix(path, "codex/") {
		return false
	}
	name := strings.TrimPrefix(path, "codex/")
	for _, prefix := range []string{"logs_", "state_"} {
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		versioned := strings.TrimPrefix(name, prefix)
		for _, suffix := range []string{".sqlite", ".sqlite-shm", ".sqlite-wal"} {
			if !strings.HasSuffix(versioned, suffix) {
				continue
			}
			version := strings.TrimSuffix(versioned, suffix)
			if version == "" {
				return false
			}
			for _, character := range version {
				if character < '0' || character > '9' {
					return false
				}
			}
			return true
		}
	}
	return false
}

func validateReceiptShape(receipt codexCleanupReceipt, candidate, runID string) error {
	decodedCandidate, candidateErr := hex.DecodeString(candidate)
	if receipt.Schema != "generated-codex-cleanup-receipt-v1" || receipt.Candidate != candidate || receipt.RunID != runID || receipt.Name == "" || candidateErr != nil || len(decodedCandidate) != 20 || candidate != strings.ToLower(candidate) || runID == "" || runID != filepath.Base(runID) || strings.ContainsAny(runID, "/\\") || receipt.RootDevice == 0 || receipt.RootInode == 0 || receipt.CodexDevice == 0 || receipt.CodexInode == 0 {
		return errors.New("generated Codex-state receipt authority is malformed")
	}
	seen := make(map[string]bool, len(receipt.Entries))
	for _, entry := range receipt.Entries {
		clean := filepath.ToSlash(filepath.Clean(entry.Path))
		kind, permissions := entry.Mode&unix.S_IFMT, entry.Mode&0o777
		unsafePermissions := entry.Mode&0o7000 != 0 || (kind == unix.S_IFREG && !validGeneratedRegularMode(permissions)) || (kind == unix.S_IFDIR && (permissions&0o022 != 0 || permissions&0o700 != 0o700))
		targetMismatch := (kind == unix.S_IFLNK) != (entry.Target != "")
		if entry.Path == "" || clean != entry.Path || strings.HasPrefix(clean, "../") || strings.HasPrefix(clean, "/") || seen[entry.Path] || entry.Device == 0 || entry.Inode == 0 || entry.UID != uint32(os.Getuid()) || entry.Nlink == 0 || entry.Size < 0 || (kind != unix.S_IFREG && kind != unix.S_IFDIR && kind != unix.S_IFLNK) || ((kind == unix.S_IFREG || kind == unix.S_IFLNK) && entry.Nlink != 1) || unsafePermissions || targetMismatch || (kind == unix.S_IFLNK && (!filepath.IsAbs(entry.Target) || !validCodexArg0HelperSymlink(entry.Path))) {
			return errors.New("generated Codex-state receipt entry is malformed")
		}
		seen[entry.Path] = true
	}
	return nil
}

func captureCodexCleanupReceipt(home, name, candidate, runID string) (codexCleanupReceipt, error) {
	paths, err := assistantinstance.Derive(home, name)
	if err != nil {
		return codexCleanupReceipt{}, err
	}
	manifest, err := assistantinstance.Verify(home, name)
	if err != nil {
		return codexCleanupReceipt{}, err
	}
	codexRoot := filepath.Join(paths.Root, "codex")
	rootInfo, err := os.Lstat(paths.Root)
	if err != nil || !rootInfo.IsDir() || rootInfo.Mode()&os.ModeSymlink != 0 {
		return codexCleanupReceipt{}, errors.New("unsafe generated Codex receipt root")
	}
	codexInfo, err := os.Lstat(codexRoot)
	if err != nil || !codexInfo.IsDir() || codexInfo.Mode()&os.ModeSymlink != 0 {
		return codexCleanupReceipt{}, errors.New("unsafe generated Codex receipt directory")
	}
	rootStat := rootInfo.Sys().(*syscall.Stat_t)
	codexStat := codexInfo.Sys().(*syscall.Stat_t)
	var entries []codexCleanupEntry
	err = filepath.WalkDir(codexRoot, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, relErr := filepath.Rel(codexRoot, path)
		if relErr != nil || relative == "." {
			return relErr
		}
		relative = filepath.ToSlash(relative)
		top := strings.Split(relative, "/")[0]
		if top == "config.toml" || top == "AGENTS.md" || top == "auth.json" || strings.HasPrefix(top, authQuarantinePrefix) {
			if strings.Contains(relative, "/") {
				return fmt.Errorf("managed Codex cleanup leaf became a directory: %s", relative)
			}
			return nil
		}
		info, statErr := os.Lstat(path)
		if statErr != nil {
			return fmt.Errorf("unsafe generated Codex entry: %s", relative)
		}
		st := info.Sys().(*syscall.Stat_t)
		isSymlink := info.Mode()&os.ModeSymlink != 0
		if st.Uid != uint32(os.Getuid()) || (!info.IsDir() && !info.Mode().IsRegular() && !isSymlink) || ((info.Mode().IsRegular() || isSymlink) && st.Nlink != 1) {
			return fmt.Errorf("unsafe generated Codex metadata: %s", relative)
		}
		target := ""
		if isSymlink {
			target, statErr = os.Readlink(path)
			if statErr != nil || target != manifest.CodexExecutable || !validCodexArg0HelperSymlink(relative) {
				return fmt.Errorf("unsafe generated Codex symlink: %s", relative)
			}
		}
		mtimeS, mtimeN := statMtime(st)
		entries = append(entries, codexCleanupEntry{Path: relative, Target: target, Device: uint64(st.Dev), Inode: st.Ino, Mode: uint32(st.Mode), UID: st.Uid, Nlink: uint64(st.Nlink), Size: st.Size, MtimeS: mtimeS, MtimeN: mtimeN})
		return nil
	})
	if err != nil {
		return codexCleanupReceipt{}, err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Path < entries[j].Path })
	receipt := codexCleanupReceipt{Schema: "generated-codex-cleanup-receipt-v1", Candidate: candidate, RunID: runID, Name: name, RootDevice: uint64(rootStat.Dev), RootInode: rootStat.Ino, CodexDevice: uint64(codexStat.Dev), CodexInode: codexStat.Ino, Entries: entries}
	if err = validateReceiptShape(receipt, candidate, runID); err != nil {
		return codexCleanupReceipt{}, err
	}
	return receipt, nil
}

func validateCodexCleanupReceipt(home, name, candidate, runID string, receipt *codexCleanupReceipt) ([]string, error) {
	if receipt == nil {
		return nil, nil
	}
	if validateReceiptShape(*receipt, candidate, runID) != nil || receipt.Name != name {
		return nil, errors.New("generated Codex-state receipt names another instance")
	}
	current, err := captureCodexCleanupReceipt(home, name, candidate, runID)
	if err != nil {
		return nil, err
	}
	if current.RootDevice != receipt.RootDevice || current.RootInode != receipt.RootInode || current.CodexDevice != receipt.CodexDevice || current.CodexInode != receipt.CodexInode || !reflect.DeepEqual(current.Entries, receipt.Entries) {
		return nil, errors.New("generated Codex state changed after receipt capture")
	}
	tops := make(map[string]bool)
	for _, entry := range receipt.Entries {
		tops[strings.Split(entry.Path, "/")[0]] = true
	}
	result := make([]string, 0, len(tops))
	for top := range tops {
		result = append(result, top)
	}
	sort.Strings(result)
	return result, nil
}

func restoreQuarantinedAuth(fromFD int, quarantine string, codexFD int) error {
	return renameNoReplace(fromFD, quarantine, codexFD, "auth.json")
}

func quarantineNamesAt(fd int, label string) ([]string, error) {
	duplicate, err := unix.Openat(fd, ".", unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if err != nil {
		return nil, err
	}
	dir := os.NewFile(uintptr(duplicate), label)
	defer dir.Close()
	names, err := dir.Readdirnames(-1)
	if err != nil {
		return nil, err
	}
	var found []string
	for _, name := range names {
		if strings.HasPrefix(name, authQuarantinePrefix) {
			found = append(found, name)
		}
	}
	return found, nil
}

func openVerifiedQuarantine(dirFD int, name string) (int, unix.Stat_t, error) {
	fd, err := unix.Openat(dirFD, name, unix.O_RDWR|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if err != nil {
		return -1, unix.Stat_t{}, err
	}
	var opened, entry unix.Stat_t
	if unix.Fstat(fd, &opened) != nil || unix.Fstatat(dirFD, name, &entry, unix.AT_SYMLINK_NOFOLLOW) != nil || opened.Dev != entry.Dev || opened.Ino != entry.Ino || name != fmt.Sprintf("%s%x-%x", authQuarantinePrefix, uint64(opened.Dev), opened.Ino) || !disposableAuthStatSafe(uint32(opened.Mode), opened.Uid, uint64(opened.Nlink)) || !disposableAuthStatSafe(uint32(entry.Mode), entry.Uid, uint64(entry.Nlink)) {
		unix.Close(fd)
		return -1, unix.Stat_t{}, errors.New("disposable auth quarantine identity or ownership is unsafe")
	}
	return fd, opened, nil
}

func neutralizeVerifiedAuth(dirFD int, name string, beforeMutation func() error) error {
	fd, opened, err := openVerifiedQuarantine(dirFD, name)
	if err != nil {
		return err
	}
	defer unix.Close(fd)
	if cleanupMutationHook != nil {
		cleanupMutationHook("auth-before-neutralize")
	}
	if err = beforeMutation(); err != nil {
		return err
	}
	if err = unix.Ftruncate(fd, 0); err != nil {
		return err
	}
	if err = unix.Fsync(fd); err != nil {
		return err
	}
	var entry unix.Stat_t
	if unix.Fstatat(dirFD, name, &entry, unix.AT_SYMLINK_NOFOLLOW) != nil || entry.Dev != opened.Dev || entry.Ino != opened.Ino || !disposableAuthStatSafe(uint32(entry.Mode), entry.Uid, uint64(entry.Nlink)) {
		return errors.New("neutralized disposable auth pathname was replaced; replacement preserved")
	}
	if cleanupMutationHook != nil {
		cleanupMutationHook("auth-after-neutralize")
	}
	return nil
}

func verifyDisposableAuthAuthority(home, name, candidate, runID string, paths assistantinstance.Paths, rootFD, codexFD int, codexExtras, rootExtras []string, receipt *codexCleanupReceipt) (assistantinstance.Manifest, error) {
	manifest, err := assistantinstance.Verify(home, name)
	if err != nil {
		return manifest, fmt.Errorf("disposable auth cleanup lacks manifest authority: %w", err)
	}
	if cleanupMutationHook != nil {
		cleanupMutationHook("authority-after-manifest-verify")
	}
	generatedExtras, err := validateCodexCleanupReceipt(home, name, candidate, runID, receipt)
	if err != nil {
		return manifest, err
	}
	codexExtras = append(append([]string(nil), codexExtras...), generatedExtras...)
	if !openedDirectoryMatchesPath(rootFD, paths.Root) || !openedDirectoryMatchesPath(codexFD, filepath.Join(paths.Root, "codex")) {
		return manifest, errors.New("disposable auth cleanup directory identity changed")
	}
	if err = verifyRootCleanupEntriesAt(rootFD, paths.Root, manifest, rootExtras...); err != nil {
		return manifest, err
	}
	if err = verifyCodexCleanupEntriesAt(codexFD, filepath.Join(paths.Root, "codex"), manifest, codexExtras...); err != nil {
		return manifest, err
	}
	return manifest, nil
}

func cleanupDisposableAuth(home, name, candidate, runID string, receipt *codexCleanupReceipt) error {
	paths, err := assistantinstance.Derive(home, name)
	if err != nil {
		return err
	}
	codexRoot := filepath.Join(paths.Root, "codex")
	rootFD, openErr := openAbsoluteDirNoFollow(paths.Root)
	if openErr != nil {
		return openErr
	}
	defer unix.Close(rootFD)
	parentFD, openErr := openAbsoluteDirNoFollow(codexRoot)
	if openErr != nil {
		return openErr
	}
	defer unix.Close(parentFD)
	var initialAuth unix.Stat_t
	authErr := unix.Fstatat(parentFD, "auth.json", &initialAuth, unix.AT_SYMLINK_NOFOLLOW)
	manifest, verifyErr := assistantinstance.Verify(home, name)
	if verifyErr != nil {
		if errors.Is(authErr, unix.ENOENT) {
			return nil
		}
		return fmt.Errorf("disposable auth cleanup lacks manifest authority: %w", verifyErr)
	}
	if !openedDirectoryMatchesPath(parentFD, codexRoot) {
		return errors.New("disposable auth directory identity changed during verification")
	}
	if !openedDirectoryMatchesPath(rootFD, paths.Root) {
		return errors.New("disposable auth instance-root identity changed during verification")
	}
	codexQuarantines, err := quarantineNamesAt(parentFD, codexRoot)
	if err != nil {
		return err
	}
	rootQuarantines, err := quarantineNamesAt(rootFD, paths.Root)
	if err != nil {
		return err
	}
	if len(codexQuarantines) > 1 || len(rootQuarantines) > 1 || (!errors.Is(authErr, unix.ENOENT) && (len(codexQuarantines) != 0 || len(rootQuarantines) != 0)) || (len(codexQuarantines) != 0 && len(rootQuarantines) != 0) {
		return errors.New("ambiguous disposable auth quarantine state preserved")
	}
	if err = verifyRootCleanupEntriesAt(rootFD, paths.Root, manifest, rootQuarantines...); err != nil {
		return err
	}
	var codexExtras []string
	generatedExtras, err := validateCodexCleanupReceipt(home, name, candidate, runID, receipt)
	if err != nil {
		return err
	}
	codexExtras = append(codexExtras, generatedExtras...)
	if !errors.Is(authErr, unix.ENOENT) {
		codexExtras = append(codexExtras, "auth.json")
	}
	codexExtras = append(codexExtras, codexQuarantines...)
	if err = verifyCodexCleanupEntriesAt(parentFD, codexRoot, manifest, codexExtras...); err != nil {
		return err
	}
	if len(rootQuarantines) == 1 {
		return neutralizeVerifiedAuth(rootFD, rootQuarantines[0], func() error {
			_, verifyErr := verifyDisposableAuthAuthority(home, name, candidate, runID, paths, rootFD, parentFD, nil, rootQuarantines, receipt)
			return verifyErr
		})
	}
	if errors.Is(authErr, unix.ENOENT) && len(codexQuarantines) == 0 {
		return nil
	}
	if authErr != nil && len(codexQuarantines) == 0 {
		return authErr
	}
	quarantine := ""
	if len(codexQuarantines) == 1 {
		quarantine = codexQuarantines[0]
	} else {
		fd, openErr := unix.Openat(parentFD, "auth.json", unix.O_RDWR|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
		if openErr != nil {
			return fmt.Errorf("disposable auth is unsafe: %w", openErr)
		}
		var opened, entry unix.Stat_t
		if unix.Fstat(fd, &opened) != nil || unix.Fstatat(parentFD, "auth.json", &entry, unix.AT_SYMLINK_NOFOLLOW) != nil || opened.Dev != entry.Dev || opened.Ino != entry.Ino || !disposableAuthStatSafe(uint32(opened.Mode), opened.Uid, uint64(opened.Nlink)) || !disposableAuthStatSafe(uint32(entry.Mode), entry.Uid, uint64(entry.Nlink)) {
			unix.Close(fd)
			return errors.New("disposable auth identity or ownership is unsafe")
		}
		if cleanupMutationHook != nil {
			cleanupMutationHook("auth-before-quarantine")
		}
		if _, err = verifyDisposableAuthAuthority(home, name, candidate, runID, paths, rootFD, parentFD, []string{"auth.json"}, nil, receipt); err != nil {
			unix.Close(fd)
			return err
		}
		quarantine = fmt.Sprintf("%s%x-%x", authQuarantinePrefix, uint64(opened.Dev), opened.Ino)
		if err = renameNoReplace(parentFD, "auth.json", parentFD, quarantine); err != nil {
			unix.Close(fd)
			return err
		}
		unix.Close(fd)
		if cleanupMutationHook != nil {
			cleanupMutationHook("auth-after-codex-quarantine")
		}
		if _, err = verifyDisposableAuthAuthority(home, name, candidate, runID, paths, rootFD, parentFD, []string{quarantine}, nil, receipt); err != nil {
			return err
		}
	}
	qfd, moved, err := openVerifiedQuarantine(parentFD, quarantine)
	if err != nil {
		if restoreErr := restoreQuarantinedAuth(parentFD, quarantine, parentFD); restoreErr != nil {
			return fmt.Errorf("disposable auth quarantine refused and preserved after restore failure: %w", restoreErr)
		}
		return err
	}
	unix.Close(qfd)
	if cleanupMutationHook != nil {
		cleanupMutationHook("auth-quarantine-verified")
	}
	if _, err = verifyDisposableAuthAuthority(home, name, candidate, runID, paths, rootFD, parentFD, []string{quarantine}, nil, receipt); err != nil {
		if restoreErr := restoreQuarantinedAuth(parentFD, quarantine, parentFD); restoreErr != nil {
			return fmt.Errorf("verified disposable auth preserved at quarantine after authority refusal: %w", restoreErr)
		}
		return err
	}
	if !openedDirectoryMatchesPath(parentFD, codexRoot) {
		if restoreErr := restoreQuarantinedAuth(parentFD, quarantine, parentFD); restoreErr != nil {
			return fmt.Errorf("verified disposable auth preserved at quarantine after directory replacement: %w", restoreErr)
		}
		return errors.New("disposable auth directory changed before deletion")
	}
	rootQuarantine := quarantine
	if err = renameNoReplace(parentFD, quarantine, rootFD, rootQuarantine); err != nil {
		return err
	}
	if cleanupMutationHook != nil {
		cleanupMutationHook("auth-after-root-quarantine")
	}
	if _, err = verifyDisposableAuthAuthority(home, name, candidate, runID, paths, rootFD, parentFD, nil, []string{rootQuarantine}, receipt); err != nil {
		return err
	}
	var rootMoved unix.Stat_t
	if unix.Fstatat(rootFD, rootQuarantine, &rootMoved, unix.AT_SYMLINK_NOFOLLOW) != nil || rootMoved.Dev != moved.Dev || rootMoved.Ino != moved.Ino || !disposableAuthStatSafe(uint32(rootMoved.Mode), rootMoved.Uid, uint64(rootMoved.Nlink)) {
		if restoreErr := restoreQuarantinedAuth(rootFD, rootQuarantine, parentFD); restoreErr != nil {
			return fmt.Errorf("verified disposable auth preserved in root quarantine after identity refusal: %w", restoreErr)
		}
		return errors.New("disposable auth changed during root quarantine transfer")
	}
	if !openedDirectoryMatchesPath(rootFD, paths.Root) || !openedDirectoryMatchesPath(parentFD, codexRoot) {
		if restoreErr := restoreQuarantinedAuth(rootFD, rootQuarantine, parentFD); restoreErr != nil {
			return fmt.Errorf("verified disposable auth preserved in root quarantine after directory replacement: %w", restoreErr)
		}
		return errors.New("disposable auth directory changed during root quarantine transfer")
	}
	if err = neutralizeVerifiedAuth(rootFD, rootQuarantine, func() error {
		_, verifyErr := verifyDisposableAuthAuthority(home, name, candidate, runID, paths, rootFD, parentFD, nil, []string{rootQuarantine}, receipt)
		return verifyErr
	}); err != nil {
		return err
	}
	return verifyCodexCleanupEntriesAt(parentFD, codexRoot, manifest, generatedExtras...)
}

func cleanupNamed(home string, names []string, leaves []exactLeaf, receipts map[string]codexCleanupReceipt, authority ...string) error {
	candidate, runID := "", ""
	if len(authority) == 2 {
		candidate, runID = authority[0], authority[1]
	} else if len(authority) != 0 {
		return errors.New("invalid generated Codex cleanup authority")
	}
	requested := make(map[string]bool, len(names))
	for _, name := range names {
		if requested[name] {
			return errors.New("duplicate cleanup instance name")
		}
		requested[name] = true
	}
	if candidate != "" || runID != "" || len(receipts) != 0 {
		if candidate == "" || runID == "" || len(receipts) != len(requested) {
			return errors.New("cleanup names and generated Codex receipts differ")
		}
		for name := range requested {
			if receipts[name].Name != name {
				return errors.New("requested cleanup instance lacks exactly one receipt")
			}
		}
		for name := range receipts {
			if !requested[name] {
				return errors.New("unused generated Codex cleanup authority")
			}
		}
	}
	for _, leaf := range leaves {
		if filepath.Dir(leaf.Path) != filepath.Join(home, ".local", "bin") || !strings.HasPrefix(filepath.Base(leaf.Path), "mfac-") {
			return errors.New("exact cleanup leaf escaped acceptance launcher scope")
		}
	}
	var failures []string
	for _, name := range names {
		if len(name) < 1 || len(name) > 32 || strings.Trim(name, "abcdefghijklmnopqrstuvwxyz0123456789-") != "" || strings.HasPrefix(name, "-") || strings.HasSuffix(name, "-") || strings.Contains(name, "--") {
			failures = append(failures, "invalid cleanup instance name")
			continue
		}
		root := filepath.Join(home, ".my-friday", "assistants", name)
		launcher := filepath.Join(home, ".local", "bin", name)
		_, rootErr := os.Lstat(root)
		_, launcherErr := os.Lstat(launcher)
		if errors.Is(rootErr, os.ErrNotExist) && errors.Is(launcherErr, os.ErrNotExist) {
			continue
		}
		if rootErr == nil && errors.Is(launcherErr, os.ErrNotExist) {
			if _, recoverErr := assistantinstance.Recover(home, name); recoverErr == nil {
				continue
			}
			failures = append(failures, "manifest-proven interrupted-state cleanup refused for "+name)
			continue
		}
		var receipt *codexCleanupReceipt
		if value, ok := receipts[name]; ok {
			receipt = &value
		}
		if err := cleanupDisposableAuth(home, name, candidate, runID, receipt); err != nil {
			failures = append(failures, err.Error())
			continue
		}
		// Re-run the complete authority check at the latest practical point. In
		// the credential case this revalidates the neutralized root quarantine
		// and every receipt-bound generated entry immediately before PlanRemove.
		if cleanupMutationHook != nil {
			cleanupMutationHook("auth-before-final-removal-validation")
		}
		if err := cleanupDisposableAuth(home, name, candidate, runID, receipt); err != nil {
			failures = append(failures, err.Error())
			continue
		}
		plan, planErr := assistantinstance.PlanRemove(home, name)
		if planErr == nil {
			if err := assistantinstance.Remove(plan); err == nil {
				continue
			}
		}
		if errors.Is(launcherErr, os.ErrNotExist) {
			if _, err := assistantinstance.Recover(home, name); err == nil {
				continue
			}
		}
		failures = append(failures, "manifest-proven cleanup refused for "+name)
	}
	for _, leaf := range leaves {
		if _, err := os.Lstat(leaf.Path); errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err := inspectExactLeaf(leaf); err != nil {
			failures = append(failures, err.Error())
			continue
		}
		if err := removeExactLeaf(leaf); err != nil {
			failures = append(failures, err.Error())
		}
	}
	if len(failures) > 0 {
		return errors.New(strings.Join(failures, "; "))
	}
	return nil
}

func validSandboxDiagnostic(version, diagnostic string) bool {
	if version != "v1" {
		return false
	}
	if diagnostic == "" {
		return true
	}
	return diagnostic == "sandbox-exec: warning: sandbox-exec is deprecated and will be removed in a future release."
}

type cleanupBinding struct {
	parent            string
	parentFD, childFD int
	parentStat        unix.Stat_t
	childStat         unix.Stat_t
	expected          map[string]bool
}

func cleanupRoots(homeFD int, runID, markerSHA string, receipt map[string]map[string]uint64, expectedEntries []string) error {
	bindings := make([]cleanupBinding, 0, 2)
	defer func() {
		for _, binding := range bindings {
			unix.Close(binding.childFD)
			unix.Close(binding.parentFD)
		}
	}()
	// Phase one opens, identity-binds, and completely validates both roots.
	// No unlink is permitted until this loop succeeds for both.
	for _, parent := range []string{".my-friday-acceptance", ".my-friday-acceptance-evidence"} {
		pfd, openErr := unix.Openat(homeFD, parent, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
		if openErr != nil {
			return openErr
		}
		var parentOpened, parentEntry unix.Stat_t
		if unix.Fstat(pfd, &parentOpened) != nil || unix.Fstatat(homeFD, parent, &parentEntry, unix.AT_SYMLINK_NOFOLLOW) != nil || parentOpened.Dev != parentEntry.Dev || parentOpened.Ino != parentEntry.Ino {
			unix.Close(pfd)
			return errors.New("cleanup parent identity changed")
		}
		cfd, openErr := unix.Openat(pfd, runID, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
		if openErr != nil {
			unix.Close(pfd)
			return openErr
		}
		var st unix.Stat_t
		expected := receipt[parent]
		if unix.Fstat(cfd, &st) != nil || uint64(st.Dev) != expected["device"] || st.Ino != expected["inode"] || st.Uid != uint32(os.Getuid()) || st.Mode&0o777 != 0o700 {
			unix.Close(cfd)
			unix.Close(pfd)
			return errors.New("cleanup root identity changed")
		}
		marker, markerErr := readRegularAt(cfd, "marker.json", 0o600)
		if markerErr != nil || fmt.Sprintf("%x", sha256.Sum256(marker)) != markerSHA {
			unix.Close(cfd)
			unix.Close(pfd)
			return errors.New("cleanup marker authority changed")
		}
		expectedPaths := map[string]bool{}
		for _, entry := range expectedEntries {
			prefix := parent + ":"
			if strings.HasPrefix(entry, prefix) {
				rel := strings.TrimPrefix(entry, prefix)
				if rel == "" || filepath.Clean(rel) != rel || filepath.IsAbs(rel) || strings.HasPrefix(rel, "../") {
					unix.Close(cfd)
					unix.Close(pfd)
					return errors.New("invalid expected cleanup entry")
				}
				expectedPaths[rel] = true
			}
		}
		if !expectedPaths["marker.json"] {
			unix.Close(cfd)
			unix.Close(pfd)
			return errors.New("cleanup marker is not expected")
		}
		seenPaths := map[string]bool{}
		if validationErr := validateExpectedContentsAt(cfd, "", expectedPaths, seenPaths); validationErr != nil || len(seenPaths) != len(expectedPaths) {
			unix.Close(cfd)
			unix.Close(pfd)
			if validationErr != nil {
				return validationErr
			}
			return errors.New("expected cleanup entry is missing")
		}
		bindings = append(bindings, cleanupBinding{parent: parent, parentFD: pfd, childFD: cfd, parentStat: parentOpened, childStat: st, expected: expectedPaths})
	}
	// Rebind both parent/child directory entries after all validation and before
	// any mutation, so a rename cannot redirect one side of the transaction.
	for _, binding := range bindings {
		var parentEntry, childEntry unix.Stat_t
		if unix.Fstatat(homeFD, binding.parent, &parentEntry, unix.AT_SYMLINK_NOFOLLOW) != nil || parentEntry.Dev != binding.parentStat.Dev || parentEntry.Ino != binding.parentStat.Ino ||
			unix.Fstatat(binding.parentFD, runID, &childEntry, unix.AT_SYMLINK_NOFOLLOW) != nil || childEntry.Dev != binding.childStat.Dev || childEntry.Ino != binding.childStat.Ino {
			return errors.New("cleanup root binding changed before mutation")
		}
	}
	// Phase two mutates only the jointly validated and rebound roots.
	for _, binding := range bindings {
		if err := removeExpectedContentsAt(binding.childFD, "", binding.expected); err != nil {
			return err
		}
		var childEntry unix.Stat_t
		if unix.Fstatat(binding.parentFD, runID, &childEntry, unix.AT_SYMLINK_NOFOLLOW) != nil || childEntry.Dev != binding.childStat.Dev || childEntry.Ino != binding.childStat.Ino {
			return errors.New("cleanup child binding changed")
		}
		if err := unix.Unlinkat(binding.parentFD, runID, unix.AT_REMOVEDIR); err != nil {
			return err
		}
	}
	return nil
}

func secureRoots(home, runID string) (map[string]map[string]uint64, error) {
	if runID == "" || runID != filepath.Base(runID) || strings.ContainsAny(runID, "/\\") {
		return nil, errors.New("invalid run identifier")
	}
	hfd, err := openAbsoluteDirNoFollow(home)
	if err != nil {
		return nil, err
	}
	defer unix.Close(hfd)
	var hs unix.Stat_t
	if unix.Fstat(hfd, &hs) != nil {
		return nil, errors.New("cannot bind canonical home")
	}
	parents := []string{".my-friday-acceptance", ".my-friday-acceptance-evidence"}
	pfds := make([]int, 0, 2)
	defer func() {
		for _, fd := range pfds {
			unix.Close(fd)
		}
	}()
	for _, parent := range parents {
		if err = unix.Mkdirat(hfd, parent, 0o700); err != nil && err != unix.EEXIST {
			return nil, err
		}
		pfd, openErr := unix.Openat(hfd, parent, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
		if openErr != nil {
			return nil, openErr
		}
		var ps unix.Stat_t
		if unix.Fstat(pfd, &ps) != nil || ps.Uid != uint32(os.Getuid()) || ps.Mode&0o777 != 0o700 || ps.Dev != hs.Dev {
			unix.Close(pfd)
			return nil, errors.New("unsafe acceptance parent")
		}
		pfds = append(pfds, pfd)
	}
	created := 0
	complete := false
	defer func() {
		if !complete {
			for rollback := 0; rollback < created; rollback++ {
				_ = unix.Unlinkat(pfds[rollback], runID, unix.AT_REMOVEDIR)
			}
		}
	}()
	for index, pfd := range pfds {
		if err = unix.Mkdirat(pfd, runID, 0o700); err != nil {
			return nil, errors.New("run root collision or unsafe creation")
		}
		created = index + 1
	}
	result := map[string]map[string]uint64{}
	for index, pfd := range pfds {
		cfd, openErr := unix.Openat(pfd, runID, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
		if openErr != nil {
			return nil, openErr
		}
		var cs unix.Stat_t
		statErr := unix.Fstat(cfd, &cs)
		unix.Close(cfd)
		if statErr != nil {
			return nil, statErr
		}
		if cs.Uid != uint32(os.Getuid()) || cs.Mode&0o777 != 0o700 || cs.Dev != hs.Dev {
			return nil, errors.New("unsafe acceptance child")
		}
		result[parents[index]] = map[string]uint64{"device": uint64(cs.Dev), "inode": cs.Ino}
	}
	complete = true
	return result, nil
}

func walkMetadata(dirfd int, relative string, records *[]string) error {
	dup, err := unix.Dup(dirfd)
	if err != nil {
		return err
	}
	dir := os.NewFile(uintptr(dup), relative)
	names, err := dir.Readdirnames(-1)
	_ = dir.Close()
	if err != nil {
		return err
	}
	sort.Strings(names)
	for _, name := range names {
		var st unix.Stat_t
		if err = unix.Fstatat(dirfd, name, &st, unix.AT_SYMLINK_NOFOLLOW); err != nil {
			return err
		}
		childRelative := filepath.Join(relative, name)
		if volatileAmbientCodexMetadata(filepath.ToSlash(childRelative)) {
			continue
		}
		*records = append(*records, fmt.Sprintf("%s|%d|%d|%d|%d|%d|%d|%o|%d|%d|%d", childRelative, st.Mode&unix.S_IFMT, st.Dev, st.Ino, st.Nlink, st.Uid, st.Gid, st.Mode&0o7777, st.Size, st.Mtim.Sec, st.Ctim.Sec))
		if st.Mode&unix.S_IFMT == unix.S_IFDIR {
			child, openErr := unix.Openat(dirfd, name, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
			if openErr != nil {
				return openErr
			}
			var opened unix.Stat_t
			if unix.Fstat(child, &opened) != nil || opened.Dev != st.Dev || opened.Ino != st.Ino {
				unix.Close(child)
				return fmt.Errorf("metadata directory changed: %s", childRelative)
			}
			if err = walkMetadata(child, childRelative, records); err != nil {
				unix.Close(child)
				return err
			}
			unix.Close(child)
		}
	}
	return nil
}

func readRegularAt(dirfd int, name string, mode uint32) ([]byte, error) {
	fd, err := unix.Openat(dirfd, name, unix.O_RDONLY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if err != nil {
		return nil, err
	}
	f := os.NewFile(uintptr(fd), name)
	defer f.Close()
	var st unix.Stat_t
	if unix.Fstat(fd, &st) != nil || st.Mode&unix.S_IFMT != unix.S_IFREG || uint32(st.Mode)&0o777 != mode || st.Uid != uint32(os.Getuid()) || st.Nlink != 1 {
		return nil, fmt.Errorf("unsafe regular file: %s", name)
	}
	return io.ReadAll(io.LimitReader(f, 32<<20))
}

func validateExpectedContentsAt(dirfd int, prefix string, expected, seen map[string]bool) error {
	if _, err := unix.Seek(dirfd, 0, 0); err != nil {
		return err
	}
	dup, err := unix.Dup(dirfd)
	if err != nil {
		return err
	}
	dir := os.NewFile(uintptr(dup), "cleanup-validation")
	names, err := dir.Readdirnames(-1)
	_ = dir.Close()
	if err != nil {
		return err
	}
	for _, name := range names {
		relative := name
		if prefix != "" {
			relative = prefix + "/" + name
		}
		if !expected[relative] {
			return fmt.Errorf("unexpected cleanup entry preserved: %s", relative)
		}
		seen[relative] = true
		var st unix.Stat_t
		if unix.Fstatat(dirfd, name, &st, unix.AT_SYMLINK_NOFOLLOW) != nil || st.Uid != uint32(os.Getuid()) || st.Mode&unix.S_IFMT == unix.S_IFLNK {
			return fmt.Errorf("unsafe cleanup entry: %s", relative)
		}
		if st.Mode&unix.S_IFMT == unix.S_IFDIR {
			child, openErr := unix.Openat(dirfd, name, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
			if openErr != nil {
				return openErr
			}
			err = validateExpectedContentsAt(child, relative, expected, seen)
			unix.Close(child)
			if err != nil {
				return err
			}
		} else if st.Mode&unix.S_IFMT != unix.S_IFREG || st.Nlink != 1 {
			return fmt.Errorf("unsupported cleanup entry: %s", relative)
		}
	}
	return nil
}

func removeExpectedContentsAt(dirfd int, prefix string, expected map[string]bool) error {
	if _, err := unix.Seek(dirfd, 0, 0); err != nil {
		return err
	}
	dup, err := unix.Dup(dirfd)
	if err != nil {
		return err
	}
	dir := os.NewFile(uintptr(dup), "cleanup-root")
	names, err := dir.Readdirnames(-1)
	_ = dir.Close()
	if err != nil {
		return err
	}
	sort.Strings(names)
	for _, name := range names {
		if name == "." || name == ".." {
			return fmt.Errorf("unsafe cleanup entry")
		}
		relative := name
		if prefix != "" {
			relative = prefix + "/" + name
		}
		if expected != nil && !expected[relative] {
			return fmt.Errorf("unexpected cleanup entry preserved: %s", relative)
		}
		var st unix.Stat_t
		if err = unix.Fstatat(dirfd, name, &st, unix.AT_SYMLINK_NOFOLLOW); err != nil {
			return err
		}
		if st.Uid != uint32(os.Getuid()) || st.Mode&unix.S_IFMT == unix.S_IFLNK {
			return fmt.Errorf("unsafe cleanup entry: %s", name)
		}
		if st.Mode&unix.S_IFMT == unix.S_IFDIR {
			child, openErr := unix.Openat(dirfd, name, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
			if openErr != nil {
				return openErr
			}
			if err = removeExpectedContentsAt(child, relative, expected); err != nil {
				unix.Close(child)
				return err
			}
			var rebound unix.Stat_t
			if unix.Fstatat(dirfd, name, &rebound, unix.AT_SYMLINK_NOFOLLOW) != nil || rebound.Dev != st.Dev || rebound.Ino != st.Ino {
				unix.Close(child)
				return fmt.Errorf("cleanup directory identity changed: %s", relative)
			}
			unix.Close(child)
			if err = unix.Unlinkat(dirfd, name, unix.AT_REMOVEDIR); err != nil {
				return err
			}
		} else if st.Mode&unix.S_IFMT == unix.S_IFREG && st.Nlink == 1 {
			fd, openErr := unix.Openat(dirfd, name, unix.O_RDONLY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
			if openErr != nil {
				return openErr
			}
			var opened, rebound unix.Stat_t
			if unix.Fstat(fd, &opened) != nil || unix.Fstatat(dirfd, name, &rebound, unix.AT_SYMLINK_NOFOLLOW) != nil || opened.Dev != st.Dev || opened.Ino != st.Ino || rebound.Dev != st.Dev || rebound.Ino != st.Ino {
				unix.Close(fd)
				return fmt.Errorf("cleanup file identity changed: %s", relative)
			}
			unix.Close(fd)
			if err = unix.Unlinkat(dirfd, name, 0); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("unsupported cleanup entry: %s", name)
		}
	}
	return unix.Fsync(dirfd)
}

func resolveExecutable(path string) (string, error) {
	if !filepath.IsAbs(path) {
		return "", fmt.Errorf("executable path must be absolute")
	}
	clean := filepath.Clean(path)
	if runtime.GOOS == "darwin" && (clean == "/var" || strings.HasPrefix(clean, "/var/")) {
		clean = filepath.Join("/private/var", strings.TrimPrefix(clean, "/var/"))
	}
	resolved := clean
	for hop := 0; hop < 32; hop++ {
		current := string(filepath.Separator)
		parts := strings.Split(strings.TrimPrefix(resolved, "/"), "/")
		followed := false
		for index, part := range parts {
			current = filepath.Join(current, part)
			entry, lerr := os.Lstat(current)
			if lerr != nil {
				return "", lerr
			}
			if entry.Mode()&os.ModeSymlink != 0 {
				st, valid := entry.Sys().(*syscall.Stat_t)
				if !valid || st.Uid != uint32(os.Getuid()) {
					return "", fmt.Errorf("unowned executable symlink: %s", current)
				}
				target, readErr := os.Readlink(current)
				if readErr != nil {
					return "", readErr
				}
				if !filepath.IsAbs(target) {
					target = filepath.Join(filepath.Dir(current), target)
				}
				remaining := parts[index+1:]
				resolved = filepath.Clean(filepath.Join(append([]string{target}, remaining...)...))
				if runtime.GOOS == "darwin" && (resolved == "/var" || strings.HasPrefix(resolved, "/var/")) {
					resolved = filepath.Join("/private/var", strings.TrimPrefix(resolved, "/var/"))
				}
				followed = true
				break
			}
		}
		if !followed {
			break
		}
		if hop == 31 {
			return "", errors.New("executable symlink chain is too deep")
		}
	}
	info, err := os.Lstat(resolved)
	if err != nil {
		return "", err
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok || !info.Mode().IsRegular() || info.Mode()&0o111 == 0 || stat.Uid != uint32(os.Getuid()) || stat.Nlink != 1 {
		return "", fmt.Errorf("resolved executable is not an owned regular executable")
	}
	return resolved, nil
}

func openAbsoluteDirNoFollow(path string) (int, error) {
	if !filepath.IsAbs(path) {
		return -1, fmt.Errorf("path must be absolute")
	}
	clean := filepath.Clean(path)
	if runtime.GOOS == "darwin" && (clean == "/var" || strings.HasPrefix(clean, "/var/")) {
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
		_ = unix.Close(fd)
		if openErr != nil {
			return -1, fmt.Errorf("unsafe path component %q: %w", part, openErr)
		}
		fd = next
	}
	return fd, nil
}

func noFollowDigest(path string) (string, bool, error) {
	body, ok, err := noFollowRead(path)
	if err != nil || !ok {
		return "", ok, err
	}
	sum := sha256.Sum256(body)
	return fmt.Sprintf("%x", sum), true, nil
}

func validateRuntimeNoFollow(root string) error {
	manifestBody, ok, err := noFollowRead(filepath.Join(root, ".my-friday/manifest.json"))
	if err != nil || !ok {
		return fmt.Errorf("unsafe runtime manifest")
	}
	profileBody, ok, err := noFollowRead(filepath.Join(root, "assistant/profile.json"))
	if err != nil || !ok {
		return fmt.Errorf("unsafe runtime profile")
	}
	repositorySchema, ok, err := noFollowRead(filepath.Join(root, ".my-friday/schemas/repository-manifest.v1.schema.json"))
	if err != nil || !ok || !bytes.Equal(repositorySchema, []byte(plan.ManifestSchema())) {
		return fmt.Errorf("runtime manifest schema mismatch")
	}
	profileSchema, ok, err := noFollowRead(filepath.Join(root, ".my-friday/schemas/assistant-profile.v1.schema.json"))
	if err != nil || !ok || !bytes.Equal(profileSchema, []byte(plan.ProfileSchema())) {
		return fmt.Errorf("runtime profile schema mismatch")
	}
	var manifest struct {
		ContractVersion int             `json:"contract_version"`
		RepositoryRole  string          `json:"repository_role"`
		AssistantID     string          `json:"assistant_id"`
		Generation      json.RawMessage `json:"generation"`
	}
	decoder := json.NewDecoder(bytes.NewReader(manifestBody))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&manifest) != nil || manifest.ContractVersion != 1 || manifest.RepositoryRole != "runtime" {
		return fmt.Errorf("runtime manifest invalid")
	}
	var p profile.Profile
	decoder = json.NewDecoder(bytes.NewReader(profileBody))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&p) != nil || profile.Validate(p) != nil || p.AssistantID != manifest.AssistantID {
		return fmt.Errorf("runtime profile invalid")
	}
	return nil
}

func noFollowRead(path string) ([]byte, bool, error) {
	dirfd, err := openAbsoluteDirNoFollow(filepath.Dir(path))
	if err != nil {
		return nil, false, err
	}
	defer unix.Close(dirfd)
	fd, err := unix.Openat(dirfd, filepath.Base(path), unix.O_RDONLY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if errors.Is(err, unix.ENOENT) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	var st unix.Stat_t
	if unix.Fstat(fd, &st) != nil || st.Mode&unix.S_IFMT != unix.S_IFREG || st.Nlink != 1 || st.Uid != uint32(os.Getuid()) {
		unix.Close(fd)
		return nil, false, fmt.Errorf("unsafe protected file: %s", path)
	}
	file := os.NewFile(uintptr(fd), path)
	body, err := io.ReadAll(io.LimitReader(file, 16<<20))
	closeErr := file.Close()
	if err != nil {
		return nil, false, err
	}
	if closeErr != nil {
		return nil, false, closeErr
	}
	var after unix.Stat_t
	if unix.Fstatat(dirfd, filepath.Base(path), &after, unix.AT_SYMLINK_NOFOLLOW) != nil || after.Dev != st.Dev || after.Ino != st.Ino {
		return nil, false, fmt.Errorf("protected file changed during read: %s", path)
	}
	return body, true, nil
}

func fatal(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
