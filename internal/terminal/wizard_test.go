package terminal

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultExitHasNoMutation(t *testing.T) {
	root := t.TempDir()
	in := strings.NewReader("\nFriday\n\nHelp me work\n\n\n" + root + "\n\n")
	var out bytes.Buffer
	result, err := Run(in, &out, root)
	if err != nil {
		t.Fatal(err)
	}
	if result != "Exit" {
		t.Fatal(result)
	}
	if !strings.Contains(out.String(), "No changes made") || strings.Contains(out.String(), "\x1b[") {
		t.Fatal(out.String())
	}
	if _, err := filepath.Glob(filepath.Join(root, "my-friday-*")); err != nil {
		t.Fatal(err)
	}
}
func TestExplicitCreate(t *testing.T) {
	root := t.TempDir()
	in := strings.NewReader("\nFriday\nBoss\nHelp me work\n2\n\n" + root + "\nCreate\n")
	var out bytes.Buffer
	result, err := Run(in, &out, root)
	if err != nil {
		t.Fatal(err)
	}
	if result != "Complete" {
		t.Fatal(result)
	}
	if err := ValidatePair(filepath.Join(root, "my-friday-runtime"), filepath.Join(root, "my-friday-memory")); err != nil {
		t.Fatal(err)
	}
}
