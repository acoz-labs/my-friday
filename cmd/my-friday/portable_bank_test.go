package main

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/acoz-labs/my-friday/internal/memorybank"
)

func TestMemoryOnlyCLI(t *testing.T) {
	root := filepath.Join(t.TempDir(), "memory bank")
	binding := filepath.Join(t.TempDir(), "memory.json")
	var out bytes.Buffer
	call := func(input string, args ...string) {
		t.Helper()
		out.Reset()
		if err := runPortable(args, strings.NewReader(input), &out, &out); err != nil {
			t.Fatalf("%v: %v %s", args, err, out.String())
		}
	}
	call("", "bank", "create", "--repository", root, "--name", "Example", "--device-label", "Test host")
	call("", "bank", "bind", "--repository", root, "--binding", binding, "--device-label", "Test host", "--actor", "Example user")
	call(`{"kind":"decision","summary":"Project name","body":"Silver Heron","basis":"user-direction","reason":"User selected"}`, "bank", "remember", "--binding", binding)
	call("", "bank", "recall", "--binding", binding, "--query", "project")
	var p memorybank.RecallPacket
	if err := json.Unmarshal(out.Bytes(), &p); err != nil || len(p.Current) != 1 || p.Current[0].Body != "Silver Heron" {
		t.Fatalf("recall: %s %v", out.String(), err)
	}
	call("", "bank", "sync", "--binding", binding)
	if !strings.Contains(out.String(), "local-only") {
		t.Fatal(out.String())
	}
	call("", "bank", "doctor", "--binding", binding)
	if !strings.Contains(out.String(), `"healthy": true`) {
		t.Fatal(out.String())
	}
	for _, input := range []string{`{"kind":"fact"} {}`, `{"kind":"fact","unknown":true}`, strings.Repeat("x", 32769)} {
		if err := runPortable([]string{"bank", "remember", "--binding", binding}, strings.NewReader(input), &out, &out); err == nil {
			t.Fatal("invalid JSON accepted")
		}
	}
}
