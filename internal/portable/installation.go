package portable

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Resolve existing aliases even when the final directories do not exist yet.
// This is an installation preflight, not protection from hostile same-UID races.
func prospectivePath(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", errors.New("installation path is required")
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	parent := abs
	for {
		_, err := os.Lstat(parent)
		if err == nil {
			resolved, err := filepath.EvalSymlinks(parent)
			if err != nil {
				return "", err
			}
			rel, err := filepath.Rel(parent, abs)
			if err != nil {
				return "", err
			}
			return filepath.Join(resolved, rel), nil
		}
		if !os.IsNotExist(err) {
			return "", err
		}
		next := filepath.Dir(parent)
		if next == parent {
			return "", err
		}
		parent = next
	}
}

func containsInstallationPath(parent, child string) bool {
	// Conservatively reject case-only overlap on macOS, including not-yet-created
	// paths. Most macOS installations use case-insensitive filesystems.
	if runtime.GOOS == "darwin" {
		parent, child = strings.ToLower(parent), strings.ToLower(child)
	}
	rel, err := filepath.Rel(parent, child)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// ValidateInstallationPaths keeps machine-local auth/session state outside
// versioned source. It must run before creating source or registering a device.
func ValidateInstallationPaths(repository, state, launcher string) error {
	repo, err := prospectivePath(repository)
	if err != nil {
		return err
	}
	local, err := prospectivePath(state)
	if err != nil {
		return err
	}
	if containsInstallationPath(repo, local) || containsInstallationPath(local, repo) {
		return errors.New("assistant repository and machine-local instance must be separate, non-nested directories")
	}
	if launcher != "" {
		path, err := prospectivePath(launcher)
		if err != nil {
			return err
		}
		if containsInstallationPath(repo, path) {
			return errors.New("machine-local launcher must be outside the assistant repository")
		}
	}
	return nil
}

var projectionDirectories = []string{".", "codex", "pi", "pi/extensions"}
var projectionFiles = []string{"codex/AGENTS.md", "pi/AGENTS.md", "codex/hooks.json", "pi/extensions/my-friday.ts"}

func (i Instance) preflightProjection(s *Store) error {
	if err := ValidateInstallationPaths(s.Root, i.Root, ""); err != nil {
		return err
	}
	for _, name := range projectionDirectories {
		info, err := os.Lstat(filepath.Join(i.Root, name))
		if os.IsNotExist(err) && name != "." {
			continue
		}
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return fmt.Errorf("projection directory must not be a symlink or file: %s", name)
		}
	}
	// Native config is not managed, but it must not redirect our initial seed.
	for _, name := range append(append([]string{}, projectionFiles...), "codex/config.toml") {
		info, err := os.Lstat(filepath.Join(i.Root, name))
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("projection target must be a regular file: %s", name)
		}
	}
	return nil
}

// Codex owns its settings after first creation. Launch supplies our required
// flags without erasing native model choices, project trust, or UI state.
func seedCodexConfig(path string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if os.IsExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err = f.WriteString("approval_policy = \"never\"\nsandbox_mode = \"danger-full-access\"\n[features]\nhooks = true\n"); err != nil {
		return err
	}
	return f.Sync()
}

// Replace a generated file atomically instead of truncating its old inode.
// A crash can leave mixed projection generations; rerunning repair or launch
// regenerates them. Native authentication and sessions are never targets.
func replaceProjectionFile(path string, data []byte) error {
	f, err := os.CreateTemp(filepath.Dir(path), ".projection-")
	if err != nil {
		return err
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	if _, err = f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	if err = os.Rename(tmp, path); err != nil {
		return err
	}
	dir, err := os.Open(filepath.Dir(path))
	if err != nil {
		return err
	}
	defer dir.Close()
	return dir.Sync()
}
