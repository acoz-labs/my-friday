package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/acoz-labs/my-friday/internal/memorybank"
	"github.com/acoz-labs/my-friday/internal/portable"
)

func TestMemoryMenuIsDefaultAndLegacyIsExplicit(t *testing.T) {
	for _, args := range [][]string{nil, {"menu", "--plain"}} {
		var out bytes.Buffer
		if err := runPortable(args, strings.NewReader("0\n"), &out, &out); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(out.String(), "Create a memory bank") || strings.Contains(out.String(), "Set up a new agent") {
			t.Fatalf("wrong front door: %s", out.String())
		}
	}
	var out bytes.Buffer
	if err := runPortable([]string{"menu", "--legacy", "--plain"}, strings.NewReader("0\n"), &out, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Set up a new agent") {
		t.Fatal("previous management menu unavailable")
	}
}

func TestMemoryMenuConnectsExistingCloneWithoutAssistantSetup(t *testing.T) {
	home := t.TempDir()
	bank, err := portable.CreateMemoryBank(filepath.Join(home, "existing-bank"), "Existing", "device-original", "Original host")
	if err != nil {
		t.Fatal(err)
	}
	binding := filepath.Join(home, "new-binding.json")
	t.Setenv("MY_FRIDAY_MEMORY_BINDING", binding)
	input := strings.Join([]string{"2", bank.Root, "New host", "Test writer", binding, "yes", "0", ""}, "\n")
	var out bytes.Buffer
	if err := memoryMenuMode(home, strings.NewReader(input), &out, true); err != nil {
		t.Fatal(err)
	}
	s, err := memorybank.OpenBinding(binding, "test")
	if err != nil || s.ID() != bank.Agent.ID {
		t.Fatalf("existing identity not preserved: %v\n%s", err, out.String())
	}
}

func TestMemoryMenuRejectsBindingInsideNewBankBeforeCreation(t *testing.T) {
	home := t.TempDir()
	root := filepath.Join(home, "bank")
	binding := filepath.Join(root, "local.json")
	t.Setenv("MY_FRIDAY_MEMORY_BINDING", binding)
	input := strings.Join([]string{"1", "Example", root, "Host", "Writer", binding, "yes", "0", ""}, "\n")
	var out bytes.Buffer
	if err := memoryMenuMode(home, strings.NewReader(input), &out, true); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(root); !os.IsNotExist(err) {
		t.Fatal("invalid local binding destination created bank")
	}
}

func TestMemoryMenuCreateInspectAndSync(t *testing.T) {
	home := t.TempDir()
	root, binding := filepath.Join(home, "bank"), filepath.Join(home, "memory.json")
	t.Setenv("MY_FRIDAY_MEMORY_BINDING", binding)
	input := strings.Join([]string{"1", "Example memory", root, "Test laptop", "Test writer", binding, "yes", "3", binding, "4", binding, "yes", "0", ""}, "\n")
	var out bytes.Buffer
	if err := memoryMenuMode(home, strings.NewReader(input), &out, true); err != nil {
		t.Fatal(err)
	}
	s, err := memorybank.OpenBinding(binding, "test")
	if err != nil {
		t.Fatalf("menu did not produce usable binding: %v\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), "Memory bank connected") || !strings.Contains(out.String(), "Memory structure is healthy") || !strings.Contains(out.String(), "local-only") || strings.Contains(out.String(), "schema_version") {
		t.Fatal(out.String())
	}
	for _, forbidden := range []string{"agent.json", "instructions", "capabilities", "codex", "pi"} {
		if _, err := os.Lstat(filepath.Join(s.Root(), forbidden)); !os.IsNotExist(err) {
			t.Fatalf("memory menu created %s", forbidden)
		}
	}
}

func TestMemoryMenuCancellationAndBindingCollisionPreserveState(t *testing.T) {
	for _, action := range []string{":back", "no", "yes"} {
		t.Run(action, func(t *testing.T) {
			home := t.TempDir()
			root, binding := filepath.Join(home, "bank"), filepath.Join(home, "existing.json")
			if err := os.WriteFile(binding, []byte("PRESERVE"), 0600); err != nil {
				t.Fatal(err)
			}
			t.Setenv("MY_FRIDAY_MEMORY_BINDING", binding)
			input := strings.Join([]string{"1", "Example", root, "Test host", "Test writer", binding, action, "0", ""}, "\n")
			var out bytes.Buffer
			if err := memoryMenuMode(home, strings.NewReader(input), &out, true); err != nil {
				t.Fatal(err)
			}
			if _, err := os.Stat(root); !os.IsNotExist(err) {
				t.Fatal("cancelled or colliding setup created bank")
			}
			got, _ := os.ReadFile(binding)
			if string(got) != "PRESERVE" {
				t.Fatal("existing binding changed")
			}
		})
	}
}

func TestMemoryConnectionPreviewAndMenuCancellation(t *testing.T) {
	home := t.TempDir()
	bank, err := portable.CreateMemoryBank(filepath.Join(home, "bank"), "Fixture", "device-fixture", "Host")
	if err != nil {
		t.Fatal(err)
	}
	binding := filepath.Join(home, "binding.json")
	if _, err := memorybank.Bind(bank.Root, binding, "Host", "Writer"); err != nil {
		t.Fatal(err)
	}
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	args := []string{"bank", "connect-codex", "--installation-home", home, "--codex-home", filepath.Join(home, "native"), "--codex", "/not/executed/codex", "--binding", binding, "--binary", binary}
	if err := runPortable(args, strings.NewReader(""), &out, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), `"state": "preview"`) || !strings.Contains(out.String(), `"binding"`) {
		t.Fatal(out.String())
	}
	t.Setenv("MY_FRIDAY_MEMORY_BINDING", binding)
	input := strings.Join([]string{"5", filepath.Join(home, "native"), "/not/executed/codex", binding, "no", "7", "0", "0", ""}, "\n")
	out.Reset()
	if err := memoryMenuMode(home, strings.NewReader(input), &out, true); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Repair Codex memory connection") || !strings.Contains(out.String(), "Update My Friday") {
		t.Fatal(out.String())
	}
	for _, dir := range []string{".local", "native"} {
		if _, err := os.Stat(filepath.Join(home, dir)); !os.IsNotExist(err) {
			t.Fatalf("cancelled preview wrote %s", dir)
		}
	}
}

func TestMemoryMenuInstallationHomeIsExplicitAndDoesNotChangeEnvironment(t *testing.T) {
	home := t.TempDir()
	before := os.Getenv("HOME")
	var out bytes.Buffer
	if err := runPortable([]string{"menu", "--plain", "--installation-home", home}, strings.NewReader("0\n"), &out, &out); err != nil {
		t.Fatal(err)
	}
	if os.Getenv("HOME") != before {
		t.Fatal("menu changed HOME")
	}
	for _, args := range [][]string{{"menu", "--installation-home", "relative"}, {"menu", "--legacy", "--installation-home", home}} {
		if err := runPortable(args, strings.NewReader("0\n"), &out, &out); err == nil {
			t.Fatal("invalid home selection accepted")
		}
	}
}
