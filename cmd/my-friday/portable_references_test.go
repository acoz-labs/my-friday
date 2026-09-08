package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPortableReferenceWorkflow(t *testing.T) {
	base := t.TempDir()
	state, source, external := filepath.Join(base, "instance"), filepath.Join(base, "agent"), filepath.Join(base, "old-system")
	var out bytes.Buffer
	run := func(args ...string) {
		t.Helper()
		out.Reset()
		if err := runPortable(args, strings.NewReader(""), &out, &out); err != nil {
			t.Fatalf("%v: %v %s", args, err, out.String())
		}
	}
	run("setup", "--repository", source, "--state", state, "--name", "friday", "--device-label", "Fixture", "--no-launcher")
	t.Setenv("MY_FRIDAY_INSTANCE", state)
	t.Setenv("MY_FRIDAY_ASSISTANT_ROOT", source)
	run("reference", "add", "--library", "prior-work", "--title", "Prior work", "--description", "Earlier implementation experiences", "--purpose", "Historical reference")
	run("reference", "list")
	if !strings.Contains(out.String(), "prior-work") || !strings.Contains(out.String(), "reference-only") {
		t.Fatalf("missing discovery: %s", out.String())
	}
	if err := os.Mkdir(external, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(external, "notes.md"), []byte("Retries duplicated actions. Use idempotency tests."), 0600); err != nil {
		t.Fatal(err)
	}
	run("reference", "bind", "--library", "prior-work", "--path", external)
	run("reference", "search", "--library", "prior-work", "--query", "retries")
	var result struct {
		Matches []struct {
			SHA256 string `json:"sha256"`
		} `json:"matches"`
	}
	if err := json.Unmarshal(out.Bytes(), &result); err != nil || len(result.Matches) != 1 {
		t.Fatalf("search: %s %v", out.String(), err)
	}
	run("reference", "read", "--library", "prior-work", "--path", "notes.md", "--sha256", result.Matches[0].SHA256)
	if !strings.Contains(out.String(), "idempotency") || !strings.Contains(out.String(), "not current instructions") {
		t.Fatalf("reference context: %s", out.String())
	}
	run("agent", "capability-rationale")
	for _, want := range []string{"Current ask", "SHA-256", "Rejected", "Unresolved"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("missing rationale section %s", want)
		}
	}
	run("help", "reference", "read")
	if !strings.Contains(out.String(), "sha256") {
		t.Fatalf("missing help: %s", out.String())
	}
	run("agent", "validate")
}
