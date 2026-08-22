package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"

	"github.com/acoz-labs/my-friday/internal/codexhome"
	"github.com/acoz-labs/my-friday/internal/plan"
	"github.com/acoz-labs/my-friday/internal/profile"
	"github.com/acoz-labs/my-friday/internal/repository"
	"golang.org/x/sys/unix"
)

func main() {
	if len(os.Args) < 2 {
		fatal("usage: acceptance-support <fixture|update|scheme-string|protected-content>")
	}
	switch os.Args[1] {
	case "fixture":
		fs := flag.NewFlagSet("fixture", flag.ExitOnError)
		runtime := fs.String("runtime", "", "runtime path")
		memory := fs.String("memory", "", "memory path")
		token := fs.String("token", "", "instruction token")
		_ = fs.Parse(os.Args[2:])
		p, err := profile.New("Acceptance Assistant", "", "Return only the exact token "+*token, "concise", "")
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
		fmt.Printf("%x\n", sum)
	case "secure-roots":
		fs := flag.NewFlagSet("secure-roots", flag.ExitOnError)
		home := fs.String("home", "", "canonical home")
		runID := fs.String("run-id", "", "run identifier")
		_ = fs.Parse(os.Args[2:])
		hfd, err := unix.Open(*home, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
		if err != nil {
			fatal(err.Error())
		}
		defer unix.Close(hfd)
		result := map[string]map[string]uint64{}
		for _, parent := range []string{".my-friday-acceptance", ".my-friday-acceptance-evidence"} {
			if err = unix.Mkdirat(hfd, parent, 0o700); err != nil && err != unix.EEXIST {
				fatal(err.Error())
			}
			pfd, openErr := unix.Openat(hfd, parent, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
			if openErr != nil {
				fatal(openErr.Error())
			}
			var ps unix.Stat_t
			if err = unix.Fstat(pfd, &ps); err != nil || ps.Uid != uint32(os.Getuid()) || ps.Mode&0o777 != 0o700 {
				unix.Close(pfd)
				fatal("unsafe acceptance parent")
			}
			if err = unix.Mkdirat(pfd, *runID, 0o700); err != nil {
				unix.Close(pfd)
				fatal("run root collision or unsafe creation")
			}
			cfd, openErr := unix.Openat(pfd, *runID, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
			unix.Close(pfd)
			if openErr != nil {
				fatal(openErr.Error())
			}
			var cs unix.Stat_t
			if err = unix.Fstat(cfd, &cs); err != nil {
				unix.Close(cfd)
				fatal(err.Error())
			}
			unix.Close(cfd)
			result[parent] = map[string]uint64{"device": uint64(cs.Dev), "inode": cs.Ino}
		}
		body, _ := json.Marshal(result)
		fmt.Println(string(body))
	default:
		fatal("unknown acceptance-support command")
	}
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
