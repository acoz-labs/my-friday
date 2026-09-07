package portable

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Inventory paths, not skill contents. Follow directory aliases like Codex does,
// but visit each canonical directory once and bound unexpected directory trees.
// This is launch-time discovery policy, not a filesystem security boundary.
func userSkillPaths(root string) ([]string, error) {
	seen := map[string]bool{}
	files := map[string]bool{}
	var walk func(string) error
	walk = func(path string) error {
		resolved, err := filepath.EvalSymlinks(path)
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			return err
		}
		if seen[resolved] {
			return nil
		}
		seen[resolved] = true
		if len(seen) > 10000 {
			return fmt.Errorf("user skill discovery exceeds 10000 paths")
		}
		info, err := os.Stat(resolved)
		if err != nil {
			return err
		}
		if !info.IsDir() {
			if filepath.Base(path) == "SKILL.md" && info.Mode().IsRegular() {
				files[resolved] = true
			}
			return nil
		}
		entries, err := os.ReadDir(resolved)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if entry.IsDir() || entry.Type()&os.ModeSymlink != 0 || entry.Name() == "SKILL.md" {
				if err := walk(filepath.Join(resolved, entry.Name())); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if err := walk(root); err != nil {
		return nil, fmt.Errorf("cannot inventory user-wide skills: %w", err)
	}
	paths := make([]string, 0, len(files))
	for path := range files {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths, nil
}

func codexUserSkillOverrides(root string) ([]string, error) {
	paths, err := userSkillPaths(root)
	if err != nil {
		return nil, err
	}
	if len(paths) == 0 {
		return nil, nil
	}
	entries := make([]string, 0, len(paths))
	for _, path := range paths {
		quoted, _ := json.Marshal(path)
		entries = append(entries, "{path="+string(quoted)+",enabled=false}")
	}
	value := "skills.config=[" + strings.Join(entries, ",") + "]"
	if len(value) > 65536 {
		return nil, fmt.Errorf("user skill exclusions exceed the 64 KiB launch argument budget")
	}
	return []string{"-c", value}, nil
}
