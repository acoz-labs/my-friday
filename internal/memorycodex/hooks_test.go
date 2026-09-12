package memorycodex

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/acoz-labs/my-friday/internal/memorybank"
	"github.com/acoz-labs/my-friday/internal/portable"
)

func TestHookReadsMemoryWithoutWritingOrReadingTranscripts(t *testing.T) {
	s, err := portable.CreateMemoryBank(filepath.Join(t.TempDir(), "bank"), "Example", "device-test", "Test machine")
	if err != nil {
		t.Fatal(err)
	}
	binding := filepath.Join(t.TempDir(), "memory.json")
	if _, err := memorybank.Bind(s.Root, binding, "Test machine", "Example user"); err != nil {
		t.Fatal(err)
	}
	service, err := memorybank.OpenBinding(binding, "codex")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Remember(memorybank.Write{Kind: "decision", Summary: "Fictional project", Body: "Silver Heron", Basis: "user-direction", Reason: "Selected"}); err != nil {
		t.Fatal(err)
	}
	transcript := filepath.Join(t.TempDir(), "transcript.json")
	if err := os.WriteFile(transcript, []byte("PRIVATE-TRANSCRIPT-CANARY"), 0600); err != nil {
		t.Fatal(err)
	}
	input, _ := json.Marshal(map[string]any{"hook_event_name": "UserPromptSubmit", "prompt": "What is my fictional project? Read-only; do not journal.", "cwd": t.TempDir(), "transcript_path": transcript, "future_native_field": true})
	var out bytes.Buffer
	if err := Run(binding, bytes.NewReader(input), &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Silver Heron") || strings.Contains(out.String(), "PRIVATE-TRANSCRIPT-CANARY") || len(out.Bytes()) > 12000 {
		t.Fatalf("hook packet: %s", out.String())
	}
	entries, err := service.Journal("", 5)
	if err != nil || len(entries) != 0 {
		t.Fatal("read-only hook created journal")
	}
	for _, event := range []string{"Stop", "Interrupt", "SessionEnd"} {
		out.Reset()
		if err := Run(binding, strings.NewReader(`{"hook_event_name":"`+event+`"}`), &out); err != nil || strings.TrimSpace(out.String()) != "{}" {
			t.Fatalf("unexpected lifecycle action %s: %s %v", event, out.String(), err)
		}
	}
}

func TestMissingBindingWarnsWithoutBlocking(t *testing.T) {
	var out bytes.Buffer
	if err := Run(filepath.Join(t.TempDir(), "missing.json"), strings.NewReader(`{"hook_event_name":"UserPromptSubmit","prompt":"hello"}`), &out); err != nil {
		t.Fatal(err)
	}
	var result map[string]any
	if err := json.Unmarshal(out.Bytes(), &result); err != nil || result["systemMessage"] == nil || result["decision"] != nil {
		t.Fatalf("warning contract: %s %v", out.String(), err)
	}
}
