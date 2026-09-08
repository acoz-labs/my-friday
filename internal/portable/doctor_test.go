package portable

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDoctorAndRepairPreserveNativeCodexSettings(t *testing.T) {
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
	config := filepath.Join(i.Root, "codex/config.toml")
	native := []byte("# Preserve native choices and comments\nmodel = \"fixture-model\"\n[features]\nhooks = false\n[projects.\"/synthetic/project\"]\ntrust_level = \"trusted\"\n[tui.model_availability_nux]\nfixture = 1\n")
	if err := os.WriteFile(config, native, 0600); err != nil {
		t.Fatal(err)
	}
	if report := i.Doctor(s, "codex"); !report.Healthy {
		t.Fatalf("native settings treated as damage: %+v", report)
	}
	if err := i.Project(s); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(config)
	if err != nil || !bytes.Equal(after, native) {
		t.Fatal("repair erased native settings")
	}
	plan, err := i.Plan(s, "codex", t.TempDir(), nil)
	if err != nil || !strings.Contains(strings.Join(plan.Arguments, "|"), "--enable|hooks") {
		t.Fatalf("hooks depend on native config: %+v %v", plan, err)
	}
	// Malformed native TOML must remain available for explicit native recovery,
	// never silently replaced with defaults by My Friday.
	if err := os.WriteFile(config, []byte("invalid native TOML ["), 0600); err != nil {
		t.Fatal(err)
	}
	if err := i.Project(s); err != nil {
		t.Fatal(err)
	}
	after, _ = os.ReadFile(config)
	if string(after) != "invalid native TOML [" {
		t.Fatal("native configuration was overwritten")
	}
}

func TestInstanceParentAliasesRenderIdentically(t *testing.T) {
	s := fixtureStore(t)
	parent := t.TempDir()
	alias := filepath.Join(t.TempDir(), "alias")
	if err := os.Symlink(parent, alias); err != nil {
		t.Fatal(err)
	}
	i, err := Bind(s, filepath.Join(alias, "state"), "friday", "/fixture/my-friday", "device-laptop")
	if err != nil {
		t.Fatal(err)
	}
	a, _, err := LoadInstance(filepath.Join(alias, "state"))
	if err != nil {
		t.Fatal(err)
	}
	b, _, err := LoadInstance(filepath.Join(parent, "state"))
	if err != nil {
		t.Fatal(err)
	}
	if i.Root != a.Root || a.Root != b.Root || !filepath.IsAbs(a.Root) {
		t.Fatalf("alias-dependent binding: %q %q %q", i.Root, a.Root, b.Root)
	}
	before, _ := a.projection(s)
	after, _ := b.projection(s)
	for name, data := range before {
		if !bytes.Equal(data, after[name]) {
			t.Fatalf("alias-dependent projection: %s", name)
		}
	}
}

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
