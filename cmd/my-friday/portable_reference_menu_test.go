package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestReferenceMenuRegistrationBindingAndCheck(t *testing.T) {
	i, s := apiFixture(t)
	external := t.TempDir()
	document := filepath.Join(external, "AGENTS.md")
	canary := "PRIVATE_REFERENCE_CONTENT_NOT_A_MENU_INSTRUCTION"
	if err := os.WriteFile(document, []byte(canary), 0600); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	input := "1\nprior-work\nPrior work\nOld project experiences\nHistorical reference only\nyes\n2\n2\n" + external + "\nyes\n3\n0\n0\n"
	u := newManagementUI(t.TempDir(), strings.NewReader(input), &out, true)
	if err := u.references(i.Root); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Reference registered", "local-only", "Local binding saved", "available", "No source files copied"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("missing %q: %s", want, out.String())
		}
	}
	if strings.Contains(out.String(), canary) || strings.Contains(out.String(), "\x1b") {
		t.Fatal("menu read content or styled plain output")
	}
	status, err := i.CheckReference(s, "prior-work")
	if err != nil || status.State != "available" {
		t.Fatalf("%+v %v", status, err)
	}
	git := exec.Command("git", "status", "--porcelain")
	git.Dir = s.Root
	if data, err := git.Output(); err != nil || len(data) != 0 {
		t.Fatalf("source not checkpointed: %s %v", data, err)
	}
	data, _ := os.ReadFile(document)
	if string(data) != canary {
		t.Fatal("external document changed")
	}
	libs, _ := s.ReferenceLibraries()
	for _, input := range []string{external + "\nno\n", ":back\n", ""} {
		u = newManagementUI(t.TempDir(), strings.NewReader(input), &out, true)
		_ = u.bindReference(i, s, libs[0])
		got, _ := i.CheckReference(s, "prior-work")
		if got.Root != status.Root {
			t.Fatal("cancel changed binding")
		}
	}
	// A deliberate rebind changes only this machine's metadata, not source HEAD.
	git = exec.Command("git", "rev-parse", "HEAD")
	git.Dir = s.Root
	head, _ := git.Output()
	next := t.TempDir()
	u = newManagementUI(t.TempDir(), strings.NewReader(next+"\nyes\n"), &out, true)
	if err := u.bindReference(i, s, libs[0]); err != nil {
		t.Fatal(err)
	}
	git = exec.Command("git", "rev-parse", "HEAD")
	git.Dir = s.Root
	after, _ := git.Output()
	if !bytes.Equal(head, after) {
		t.Fatal("rebind checkpointed source")
	}
}

func TestReferenceMenuCancelledOrDirtyRegistrationDoesNotWrite(t *testing.T) {
	for _, input := range []string{"", ":back\n", "prior-work\nTitle\nDescription\nPurpose\nno\n", "prior-work\nTitle\nDescription\nPurpose\nyes\n"} {
		i, s := apiFixture(t)
		if strings.HasSuffix(input, "yes\n") {
			if err := os.WriteFile(filepath.Join(s.Root, "unrelated.txt"), []byte("user work"), 0600); err != nil {
				t.Fatal(err)
			}
		}
		var out bytes.Buffer
		u := newManagementUI(t.TempDir(), strings.NewReader(input), &out, true)
		_ = u.addReference(i, s)
		libs, err := s.ReferenceLibraries()
		if err != nil || len(libs) != 0 {
			t.Fatalf("registration wrote unexpectedly: %+v %v", libs, err)
		}
	}
}

func TestReferenceMenuOfflineRegistrationRetainsDescriptor(t *testing.T) {
	i, s := apiFixture(t)
	git := exec.Command("git", "remote", "add", "origin", filepath.Join(t.TempDir(), "missing.git"))
	git.Dir = s.Root
	if data, err := git.CombinedOutput(); err != nil {
		t.Fatalf("%s %v", data, err)
	}
	var out bytes.Buffer
	u := newManagementUI(t.TempDir(), strings.NewReader("prior-work\nTitle\nDescription\nPurpose\nyes\n"), &out, true)
	if err := u.addReference(i, s); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "pending") {
		t.Fatal(out.String())
	}
	libs, err := s.ReferenceLibraries()
	if err != nil || len(libs) != 1 {
		t.Fatal("offline descriptor lost")
	}
	// Existing core sync remains recoverable; do not retry registration.
	status, err := s.Sync(context.Background())
	if err != nil || status.State != "pending" {
		t.Fatalf("%+v %v", status, err)
	}
}
