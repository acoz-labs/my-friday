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
	"runtime"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/acoz-labs/my-friday/internal/codexhome"
	"github.com/acoz-labs/my-friday/internal/plan"
	"github.com/acoz-labs/my-friday/internal/profile"
	"github.com/acoz-labs/my-friday/internal/repository"
	"golang.org/x/sys/unix"
)

func main() {
	if len(os.Args) < 2 {
		fatal("usage: acceptance-support <fixture|update|resolve-executable|render-profile|protected-content|secure-roots>")
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
		defer unix.Close(hfd)
		for _, parent := range []string{".my-friday-acceptance", ".my-friday-acceptance-evidence"} {
			pfd, openErr := unix.Openat(hfd, parent, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
			if openErr != nil {
				fatal(openErr.Error())
			}
			var parentOpened, parentEntry unix.Stat_t
			if unix.Fstat(pfd, &parentOpened) != nil || unix.Fstatat(hfd, parent, &parentEntry, unix.AT_SYMLINK_NOFOLLOW) != nil || parentOpened.Dev != parentEntry.Dev || parentOpened.Ino != parentEntry.Ino {
				unix.Close(pfd)
				fatal("cleanup parent identity changed")
			}
			cfd, openErr := unix.Openat(pfd, *runID, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
			if openErr != nil {
				unix.Close(pfd)
				fatal(openErr.Error())
			}
			var st unix.Stat_t
			expected := receipt[parent]
			if unix.Fstat(cfd, &st) != nil || uint64(st.Dev) != expected["device"] || st.Ino != expected["inode"] || st.Uid != uint32(os.Getuid()) || st.Mode&0o777 != 0o700 {
				unix.Close(cfd)
				unix.Close(pfd)
				fatal("cleanup root identity changed")
			}
			marker, markerErr := readRegularAt(cfd, "marker.json", 0o600)
			if markerErr != nil || fmt.Sprintf("%x", sha256.Sum256(marker)) != *markerSHA {
				unix.Close(cfd)
				unix.Close(pfd)
				fatal("cleanup marker authority changed")
			}
			if err = removeContentsAt(cfd); err != nil {
				unix.Close(cfd)
				unix.Close(pfd)
				fatal(err.Error())
			}
			var childEntry unix.Stat_t
			if unix.Fstatat(pfd, *runID, &childEntry, unix.AT_SYMLINK_NOFOLLOW) != nil || childEntry.Dev != st.Dev || childEntry.Ino != st.Ino {
				unix.Close(cfd)
				unix.Close(pfd)
				fatal("cleanup child binding changed")
			}
			unix.Close(cfd)
			if err = unix.Unlinkat(pfd, *runID, unix.AT_REMOVEDIR); err != nil {
				unix.Close(pfd)
				fatal(err.Error())
			}
			unix.Close(pfd)
		}
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

func removeContentsAt(dirfd int) error {
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
			if err = removeContentsAt(child); err != nil {
				unix.Close(child)
				return err
			}
			unix.Close(child)
			if err = unix.Unlinkat(dirfd, name, unix.AT_REMOVEDIR); err != nil {
				return err
			}
		} else if st.Mode&unix.S_IFMT == unix.S_IFREG && st.Nlink == 1 {
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
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 || info.Sys().(*syscall.Stat_t).Nlink != 1 || info.Sys().(*syscall.Stat_t).Uid != uint32(os.Getuid()) {
		return "", false, fmt.Errorf("unsafe protected file: %s", path)
	}
	fd, err := unix.Open(path, unix.O_RDONLY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if err != nil {
		return "", false, err
	}
	file := os.NewFile(uintptr(fd), path)
	body, err := io.ReadAll(io.LimitReader(file, 16<<20))
	closeErr := file.Close()
	if err != nil {
		return "", false, err
	}
	if closeErr != nil {
		return "", false, closeErr
	}
	after, err := os.Lstat(path)
	if err != nil || !os.SameFile(info, after) {
		return "", false, fmt.Errorf("protected file changed during read: %s", path)
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
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 || info.Sys().(*syscall.Stat_t).Nlink != 1 || info.Sys().(*syscall.Stat_t).Uid != uint32(os.Getuid()) {
		return nil, false, fmt.Errorf("unsafe protected file: %s", path)
	}
	fd, err := unix.Open(path, unix.O_RDONLY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if err != nil {
		return nil, false, err
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
	after, err := os.Lstat(path)
	if err != nil || !os.SameFile(info, after) {
		return nil, false, fmt.Errorf("protected file changed during read: %s", path)
	}
	return body, true, nil
}

func fatal(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
