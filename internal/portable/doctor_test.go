package portable

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDoctorReportsProjectionDriftWithoutRepair(t *testing.T) {
	s := fixtureStore(t)
	if err := s.InitGit(context.Background()); err != nil {
		t.Fatal(err)
	}
	bin := t.TempDir()
	for _, name := range []string{"my-friday", "codex"} {
		if err := os.WriteFile(filepath.Join(bin, name), []byte("#!/bin/sh\nexit 99\n"), 0700); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", bin)
	i, err := Bind(s, filepath.Join(t.TempDir(), "state"), "friday", filepath.Join(bin, "my-friday"), "device-laptop")
	if err != nil {
		t.Fatal(err)
	}
	r := i.Doctor(s, "codex")
	if !r.Healthy {
		t.Fatalf("healthy fixture: %+v", r)
	}
	path := filepath.Join(i.Root, "codex/AGENTS.md")
	if err := os.WriteFile(path, []byte("outdated fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	missing := filepath.Join(i.Root, "pi/extensions/my-friday.ts")
	if err := os.Remove(missing); err != nil {
		t.Fatal(err)
	}
	r = i.Doctor(s, "codex")
	if r.Healthy {
		t.Fatal("drift reported healthy")
	}
	issues := map[string]bool{}
	for _, check := range r.Checks {
		if !check.OK {
			issues[check.Name] = true
		}
	}
	if !issues["projection:codex/AGENTS.md"] || !issues["projection:pi/extensions/my-friday.ts"] {
		t.Fatalf("missing specific repair diagnostics: %+v", r)
	}
	data, _ := os.ReadFile(path)
	if string(data) != "outdated fixture" {
		t.Fatal("doctor repaired without authorization")
	}
	if _, err := os.Lstat(missing); !os.IsNotExist(err) {
		t.Fatal("doctor recreated missing file")
	}
	if err := i.Project(s); err != nil {
		t.Fatal(err)
	}
	if r := i.Doctor(s, "codex"); !r.Healthy {
		t.Fatalf("repair did not clear findings: %+v", r)
	}
	if r := i.Doctor(s, "pi"); r.Healthy {
		t.Fatal("missing selected harness reported healthy")
	}
	if err := os.Rename(filepath.Join(s.Root, ".git"), filepath.Join(s.Root, ".git-saved")); err != nil {
		t.Fatal(err)
	}
	if r := i.Doctor(s, "codex"); r.Healthy {
		t.Fatal("missing Git metadata reported healthy")
	}
}

func TestDoctorDoesNotFollowProjectionSymlinks(t *testing.T) {
	s := fixtureStore(t)
	i, err := Bind(s, filepath.Join(t.TempDir(), "state"), "friday", "/missing/binary", "device-laptop")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(i.Root, "codex/AGENTS.md")
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("/nonexistent-fixture", path); err != nil {
		t.Fatal(err)
	}
	r := i.Doctor(s, "codex")
	if r.Healthy {
		t.Fatal("unsafe projection reported healthy")
	}
	found := false
	for _, c := range r.Checks {
		if c.Name == "projection-paths" && !c.OK && strings.Contains(c.Detail, "regular file") {
			found = true
		}
	}
	if !found {
		t.Fatalf("missing unsafe-path diagnostic: %+v", r)
	}
	if target, err := os.Readlink(path); err != nil || target != "/nonexistent-fixture" {
		t.Fatal("doctor touched symlink")
	}
}
