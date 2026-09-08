package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/acoz-labs/my-friday/internal/portable"
)

func TestPortableHookWarningsAndDeadline(t *testing.T) {
	for _, harness := range []string{"codex", "pi"} {
		for _, slow := range []bool{false, true} {
			t.Run(harness+map[bool]string{false: "/warn", true: "/deadline"}[slow], func(t *testing.T) {
				base := t.TempDir()
				s, err := portable.Create(filepath.Join(base, "agent"), "friday", harness, "device-example", "Fixture")
				if err != nil {
					t.Fatal(err)
				}
				i, err := portable.Bind(s, filepath.Join(base, "instance"), "friday", "/fixture/my-friday", "device-example")
				if err != nil {
					t.Fatal(err)
				}
				root := filepath.Join(s.Root, "capabilities/example")
				if err := os.Mkdir(root, 0700); err != nil {
					t.Fatal(err)
				}
				script := "#!/bin/sh\nprintf sensitive-fixture-output\nexit 1\n"
				if slow {
					script = "#!/bin/sh\nsleep 30\n"
				}
				if err := os.WriteFile(filepath.Join(root, "run.sh"), []byte(script), 0700); err != nil {
					t.Fatal(err)
				}
				// tool.before avoids network work; this isolates dispatcher behavior.
				manifest := portable.Capability{Version: 1, ID: "example", Description: "Synthetic failure", Subscriptions: []portable.Subscription{{ID: "handler", Event: "tool.before", Command: []string{"sh", "run.sh"}, Failure: "warn"}}}
				data, _ := json.Marshal(manifest)
				if err := os.WriteFile(filepath.Join(root, "capability.json"), data, 0600); err != nil {
					t.Fatal(err)
				}
				native := "PreToolUse"
				if harness == "pi" {
					native = "tool_call"
				}
				ctx := context.Background()
				if slow {
					var cancel context.CancelFunc
					ctx, cancel = context.WithTimeout(ctx, 150*time.Millisecond)
					defer cancel()
				}
				var out bytes.Buffer
				if err := portableHookContext(ctx, []string{"--instance", i.Root, "--harness", harness, "--native", native}, strings.NewReader(`{"event_id":"event-fixture"}`), &out, &out); err != nil {
					t.Fatal(err)
				}
				var response map[string]any
				if err := json.Unmarshal(out.Bytes(), &response); err != nil {
					t.Fatal(err)
				}
				if !strings.Contains(out.String(), "example/handler") || strings.Contains(out.String(), "sensitive-fixture-output") {
					t.Fatalf("missing/redaction failure: %s", out.String())
				}
				if harness == "pi" && response["warnings"] == nil {
					t.Fatal("Pi warning missing")
				}
				if harness == "codex" && response["systemMessage"] == nil {
					t.Fatal("Codex warning missing")
				}
				if slow && !strings.Contains(out.String(), "deadline or cancellation") {
					t.Fatal("deadline status missing")
				}
			})
		}
	}
}

func TestPortableCompletionSyncWarningsAreVisible(t *testing.T) {
	for _, harness := range []string{"codex", "pi"} {
		t.Run(harness, func(t *testing.T) {
			s, err := portable.Create(filepath.Join(t.TempDir(), "agent"), "friday", harness, "device-example", "Fixture")
			if err != nil {
				t.Fatal(err)
			}
			i, err := portable.Bind(s, filepath.Join(t.TempDir(), "instance"), "friday", "/fixture/my-friday", "device-example")
			if err != nil {
				t.Fatal(err)
			}
			// Intentionally no Git initialization: checkpoint must report attention.
			native := "Stop"
			if harness == "pi" {
				native = "agent_settled"
			}
			var out bytes.Buffer
			if err := portableHookContext(context.Background(), []string{"--instance", i.Root, "--harness", harness, "--native", native}, strings.NewReader(`{"event_id":"event-sync-warning"}`), &out, &out); err != nil {
				t.Fatal(err)
			}
			var response map[string]json.RawMessage
			if err := json.Unmarshal(out.Bytes(), &response); err != nil {
				t.Fatal(err)
			}
			key := "systemMessage"
			if harness == "pi" {
				key = "warnings"
			}
			if !strings.Contains(string(response[key]), "synchronization needs attention") {
				t.Fatalf("completion warning hidden: %s", out.String())
			}
		})
	}
}
