package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/acoz-labs/my-friday/internal/portable"
)

func TestPortableSetupAndMemoryRoundTrip(t *testing.T) {
	root := filepath.Join(t.TempDir(), "assistant")
	state := filepath.Join(t.TempDir(), "instance")
	var out bytes.Buffer
	err := runPortable([]string{"setup", "--repository", root, "--state", state, "--name", "friday", "--device-label", "Test laptop", "--no-launcher"}, strings.NewReader(""), &out, &out)
	if err != nil {
		t.Fatal(err)
	}
	instance, s, err := portable.LoadInstance(state)
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("MY_FRIDAY_ASSISTANT_ROOT", root)
	t.Setenv("MY_FRIDAY_DEVICE_ID", instance.DeviceID)
	t.Setenv("MY_FRIDAY_HARNESS", "pi")
	out.Reset()
	if err = runPortable([]string{"memory", "template"}, strings.NewReader(""), &out, &out); err != nil {
		t.Fatal(err)
	}
	var r portable.Revision
	if err = json.Unmarshal(out.Bytes(), &r); err != nil {
		t.Fatal(err)
	}
	r.ID = "revision-learning"
	r.RecordID = "record-learning"
	r.Summary = "Prefer short answers"
	r.Body = "The user prefers concise responses."
	r.Authorship.DeviceID = "device-forged"
	data, _ := json.Marshal(r)
	file := filepath.Join(t.TempDir(), "input.json")
	if err = os.WriteFile(file, data, 0600); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if err = runPortable([]string{"memory", "write", "--input", file}, strings.NewReader(""), &out, &out); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if err = runPortable([]string{"memory", "recall", "--query", "short answers"}, strings.NewReader(""), &out, &out); err != nil {
		t.Fatal(err)
	}
	var packet portable.Packet
	if err = json.Unmarshal(out.Bytes(), &packet); err != nil {
		t.Fatal(err)
	}
	if len(packet.Current) != 1 || packet.Current[0].Authorship.DeviceID != instance.DeviceID || packet.Current[0].Authorship.Harness != "pi" {
		t.Fatalf("writer provenance not stamped: %+v", packet)
	}
	if err = s.Validate(); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if err := runPortable([]string{"hook", "--instance", state, "--harness", "codex", "--native", "UserPromptSubmit"}, strings.NewReader(`{"prompt":"short answers","event_id":"event-fixture"}`), &out, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "revision-learning") || !strings.Contains(out.String(), "hookSpecificOutput") {
		t.Fatal("native prompt hook omitted current memory")
	}
	bin := t.TempDir()
	if err := os.WriteFile(filepath.Join(bin, "pi"), []byte("#!/bin/sh\nprintf '%s\\n' \"$PWD\" \"$MY_FRIDAY_ASSISTANT_ROOT\" \"$PI_CODING_AGENT_DIR\" \"$@\"\n"), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	project := t.TempDir()
	t.Chdir(project)
	project, _ = os.Getwd()
	out.Reset()
	if err := runPortable([]string{"agent", "launch", "--instance", state, "--harness", "pi", "--model", "fixture", "hello world"}, strings.NewReader(""), &out, &out); err != nil {
		t.Fatal(err)
	}
	want := strings.Join([]string{project, s.Root, filepath.Join(state, "pi"), "--model", "fixture", "hello world", ""}, "\n")
	if out.String() != want {
		t.Fatalf("launch changed cwd/identity/arguments: %q", out.String())
	}
}

func TestPortableCLIRejectsUnknownFlagsAndTrailingJSON(t *testing.T) {
	var out bytes.Buffer
	if err := runPortable([]string{"setup", "--not-a-real-flag"}, strings.NewReader(""), &out, &out); err == nil {
		t.Fatal("unknown option accepted")
	}
}

func TestPortableScopeDiscoveryCLI(t *testing.T) {
	s, err := portable.Create(filepath.Join(t.TempDir(), "agent"), "friday", "pi", "device-example", "Example")
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("MY_FRIDAY_ASSISTANT_ROOT", s.Root)
	var out bytes.Buffer
	if err := runPortable([]string{"memory", "scopes"}, strings.NewReader(""), &out, &out); err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out.String()) != "[]" {
		t.Fatalf("empty scope inventory: %s", out.String())
	}
	root := filepath.Join(s.Root, "capabilities", "example-capability")
	if err := os.MkdirAll(root, 0700); err != nil {
		t.Fatal(err)
	}
	for name, content := range map[string]string{
		"capability.json": `{"schema_version":1,"id":"example-capability","description":"Synthetic capability","subscriptions":[]}`,
		"instructions.md": "Use the existing implementation.",
	} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}
	out.Reset()
	if err := runPortable([]string{"agent", "capabilities"}, strings.NewReader(""), &out, &out); err != nil {
		t.Fatal(err)
	}
	var caps []portable.CapabilityInfo
	if err := json.Unmarshal(out.Bytes(), &caps); err != nil || len(caps) != 1 || len(caps[0].InstructionFiles) != 1 || caps[0].Directory != root {
		t.Fatalf("CLI omitted capability navigation: %s %v", out.String(), err)
	}
}

func TestServiceCapabilitiesAreNotCoreCommands(t *testing.T) {
	for _, command := range []string{"github", "secret"} {
		var out bytes.Buffer
		err := runPortable([]string{command}, strings.NewReader(""), &out, &out)
		if err == nil || !strings.HasPrefix(err.Error(), "usage: my-friday <setup|") {
			t.Fatalf("service command %s still routed by core: %v", command, err)
		}
	}
}

func TestPortableLaunchForwardsHarnessFlags(t *testing.T) {
	owned, forwarded, err := splitLaunchArgs([]string{"--instance", "/state", "--harness=pi", "--model", "example", "hello world", "--", "--harness", "literal"})
	if err != nil || !reflect.DeepEqual(owned, []string{"--instance", "/state", "--harness=pi"}) || !reflect.DeepEqual(forwarded, []string{"--model", "example", "hello world", "--", "--harness", "literal"}) {
		t.Fatalf("arguments changed: %q %q %v", owned, forwarded, err)
	}
	if _, _, err := splitLaunchArgs([]string{"--harness"}); err == nil {
		t.Fatal("missing harness accepted")
	}
}

func TestPortableHelpWorksWithoutInstallation(t *testing.T) {
	t.Setenv("MY_FRIDAY_ASSISTANT_ROOT", "/nonexistent-assistant")
	for _, args := range [][]string{{"--help"}, {"-h"}, {"help"}, {"agent"}, {"agent", "--help"}, {"memory", "--help"}, {"help", "agent"}, {"agent", "check", "--help"}, {"setup", "--help"}, {"help", "agent", "launch"}} {
		var out, errs bytes.Buffer
		if err := runPortable(args, strings.NewReader(""), &out, &errs); err != nil {
			t.Errorf("help %v failed: %v", args, err)
		}
		if out.Len() == 0 || errs.Len() != 0 {
			t.Errorf("help %v not on stdout: %q %q", args, out.String(), errs.String())
		}
	}
	var out bytes.Buffer
	if err := runPortable([]string{"help", "agent", "not-real"}, strings.NewReader(""), &out, &out); err == nil {
		t.Fatal("unknown help command accepted")
	}
}

func TestCapabilityAuthoringIsSelfContained(t *testing.T) {
	t.Setenv("MY_FRIDAY_ASSISTANT_ROOT", "/nonexistent-assistant")
	var out, errs bytes.Buffer
	if err := runPortable([]string{"agent", "capability-guide"}, strings.NewReader(""), &out, &errs); err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{"capability.json", "MY_FRIDAY_ASSISTANT_ROOT", "MY_FRIDAY_BIN", "subscriptions", "checks", "schema_version", "temporary", "agent check", "instructions.md"} {
		if !strings.Contains(out.String(), required) {
			t.Errorf("guide omits %s", required)
		}
	}
	out.Reset()
	if err := runPortable([]string{"agent", "capability-template", "--capability", "example-check", "--description", "Synthetic test capability"}, strings.NewReader(""), &out, &errs); err != nil {
		t.Fatal(err)
	}
	var manifest portable.Capability
	if err := json.Unmarshal(out.Bytes(), &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Version != 1 || manifest.ID != "example-check" || manifest.Description != "Synthetic test capability" || len(manifest.Subscriptions) != 0 || len(manifest.Checks) == 0 {
		t.Fatalf("incomplete template: %+v", manifest)
	}
	out.Reset()
	if err := runPortable([]string{"agent", "capability-template", "--capability", "../../outside"}, strings.NewReader(""), &out, &errs); err == nil {
		t.Fatal("invalid template ID accepted")
	}
}
