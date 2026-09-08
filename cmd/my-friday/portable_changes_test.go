package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/acoz-labs/my-friday/internal/portable"
)

func TestPortableSourceChangesUseBoundAndExplicitDevice(t *testing.T) {
	base := t.TempDir()
	root, state := filepath.Join(base, "agent"), filepath.Join(base, "instance")
	var out bytes.Buffer
	run := func(args ...string) {
		t.Helper()
		out.Reset()
		if err := runPortable(args, strings.NewReader(`{"session_id":"native-session","event_id":"event-provenance"}`), &out, &out); err != nil {
			t.Fatalf("%v: %v\n%s", args, err, out.String())
		}
	}
	run("setup", "--repository", root, "--state", state, "--name", "friday", "--device-label", "Fixture", "--no-launcher")
	instance, s, err := portable.LoadInstance(state)
	if err != nil {
		t.Fatal(err)
	}
	changes, err := s.SourceChanges("")
	if err != nil || len(changes) != 1 || changes[0].Observer == nil || changes[0].Observer.DeviceID != instance.DeviceID || changes[0].Observer.Harness != "setup" {
		t.Fatalf("setup provenance: %+v %v", changes, err)
	}
	for _, checkpoint := range []string{"explicit", "environment", "hook", "unbound"} {
		rel := "instructions/" + checkpoint + ".md"
		if err := os.WriteFile(filepath.Join(root, rel), []byte("Synthetic configuration"), 0600); err != nil {
			t.Fatal(err)
		}
		switch checkpoint {
		case "explicit":
			run("sync", "--repository", root, "--device", instance.DeviceID)
		case "environment":
			t.Setenv("MY_FRIDAY_ASSISTANT_ROOT", root)
			t.Setenv("MY_FRIDAY_DEVICE_ID", instance.DeviceID)
			t.Setenv("MY_FRIDAY_HARNESS", "pi")
			t.Setenv("MY_FRIDAY_SESSION_ID", "launcher-session")
			run("sync")
		case "hook":
			// Native adapter identity must take precedence over incoming env.
			t.Setenv("MY_FRIDAY_DEVICE_ID", "device-forged")
			run("hook", "--instance", state, "--harness", "pi", "--native", "agent_settled")
		case "unbound":
			t.Setenv("MY_FRIDAY_DEVICE_ID", "")
			run("sync", "--repository", root)
		}
		run("agent", "changes", "--repository", root, "--path", rel)
		var report struct {
			Changes []portable.SourceChange `json:"changes"`
			Notice  string                  `json:"notice"`
		}
		if err := json.Unmarshal(out.Bytes(), &report); err != nil {
			t.Fatal(err)
		}
		if len(report.Changes) != 1 || !strings.Contains(report.Notice, "not proof of authorship") {
			t.Fatalf("report: %s", out.String())
		}
		observer := report.Changes[0].Observer
		if checkpoint == "unbound" {
			if observer != nil {
				t.Fatalf("invented observer: %+v", observer)
			}
			continue
		}
		if observer == nil || observer.DeviceID != instance.DeviceID {
			t.Fatalf("wrong observer: %+v", observer)
		}
		if checkpoint == "environment" && (observer.Harness != "pi" || observer.SessionID == nil || *observer.SessionID != "launcher-session") {
			t.Fatalf("lost environment: %+v", observer)
		}
		if checkpoint == "hook" && (observer.SessionID == nil || *observer.SessionID != "native-session") {
			t.Fatalf("lost native session: %+v", observer)
		}
	}
}
