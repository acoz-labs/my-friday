package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/acoz-labs/my-friday/internal/portable"
)

func TestManagementMenuNavigationAndEOF(t *testing.T) {
	for _, input := range []string{"", "wrong\n0\n", "3\n0\n0\n", "1\n:back\n0\n"} {
		var out bytes.Buffer
		home := t.TempDir()
		if err := managementMenu(home, strings.NewReader(input), &out); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(out.String(), "Manage an installed agent") || strings.Contains(out.String(), "schema_version") {
			t.Fatal(out.String())
		}
		if _, err := os.Stat(filepath.Join(home, ".local")); !os.IsNotExist(err) {
			t.Fatal("navigation created installation")
		}
	}
}

func TestManagementMenuDiscoversAndRepairs(t *testing.T) {
	home := t.TempDir()
	root := filepath.Join(home, ".local/share/my-friday/repositories/pilot")
	s, err := portable.Create(root, "pilot", "codex", "device-menu", "Fixture")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.InitGit(context.Background()); err != nil {
		t.Fatal(err)
	}
	binary, _ := os.Executable()
	i, err := portable.Bind(s, filepath.Join(home, ".local/share/my-friday/instances/pilot"), "pilot", binary, "device-menu")
	if err != nil {
		t.Fatal(err)
	}
	if err := i.InstallLauncher(filepath.Join(home, ".local/bin/pilot")); err != nil {
		t.Fatal(err)
	}
	native := filepath.Join(i.Root, "codex/config.toml")
	before, _ := os.ReadFile(native)
	os.Remove(filepath.Join(i.Root, "codex/hooks.json"))
	var out bytes.Buffer
	// Select discovered agent, doctor, repair, confirm, back to list, back home, exit.
	if err := managementMenu(home, strings.NewReader("3\n1\n4\n5\nyes\n0\n0\n0\n"), &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Needs attention") || !strings.Contains(out.String(), "Managed files refreshed") {
		t.Fatal(out.String())
	}
	if _, err := os.Stat(filepath.Join(i.Root, "codex/hooks.json")); err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(native)
	if !bytes.Equal(before, after) {
		t.Fatal("native config changed")
	}
}

func TestManagementSetupReturnsToMenuWithoutJSON(t *testing.T) {
	home := t.TempDir()
	var out bytes.Buffer
	err := managementMenu(home, strings.NewReader("1\npilot\npi\nFixture laptop\nyes\n1\n0\n3\n0\n0\n"), &out)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), "schema_version") || !strings.Contains(out.String(), "pilot is installed") || !strings.Contains(out.String(), "Default harness: pi") {
		t.Fatal(out.String())
	}
	i, _, err := portable.LoadInstance(filepath.Join(home, ".local/share/my-friday/instances/pilot"))
	if err != nil || i.Name != "pilot" {
		t.Fatalf("%+v %v", i, err)
	}
}
