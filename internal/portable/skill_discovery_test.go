package portable

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUserSkillExclusionsFollowLinksWithoutChangingSources(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(t.TempDir(), "linked skill")
	if err := os.MkdirAll(target, 0700); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(target, "SKILL.md")
	content := []byte("synthetic discovery canary\n")
	if err := os.WriteFile(file, content, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(root, "alias")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(root, filepath.Join(target, "cycle")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(root, "missing"), filepath.Join(root, "broken")); err != nil {
		t.Fatal(err)
	}
	paths, err := userSkillPaths(root)
	if err != nil {
		t.Fatal(err)
	}
	resolved, _ := filepath.EvalSymlinks(file)
	if len(paths) != 1 || paths[0] != resolved {
		t.Fatalf("paths: %v", paths)
	}
	after, _ := os.ReadFile(file)
	if string(after) != string(content) {
		t.Fatal("modified ambient skill")
	}
	paths, err = userSkillPaths(filepath.Join(root, "absent"))
	if err != nil || len(paths) != 0 {
		t.Fatalf("missing root: %v %v", paths, err)
	}
}

func TestCodexSkillOverridesAreScopedAndSafelyEncoded(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "quote\" and space")
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("canary"), 0600); err != nil {
		t.Fatal(err)
	}
	args, err := codexUserSkillOverrides(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(args) != 2 || args[0] != "-c" || !strings.HasPrefix(args[1], "skills.config=[") || !strings.Contains(args[1], `quote\" and space`) || !strings.Contains(args[1], "enabled=false") {
		t.Fatalf("unsafe override: %q", args)
	}
	if err := os.WriteFile(filepath.Join(root, "SKILL.md"), []byte("new canary"), 0600); err != nil {
		t.Fatal(err)
	}
	updated, err := codexUserSkillOverrides(root)
	if err != nil || strings.Count(updated[1], "enabled=false") != 2 {
		t.Fatalf("new skill missed: %v %v", updated, err)
	}
}
