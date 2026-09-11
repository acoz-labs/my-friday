package portable

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func fixtureCapability(t *testing.T, s *Store, id string, subscriptions []Subscription, script string) {
	t.Helper()
	root := filepath.Join(s.Root, "capabilities", id)
	if err := os.MkdirAll(filepath.Join(root, "scripts"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "scripts/run.sh"), []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "instructions.md"), []byte("Use this capability when its scope fits the task."), 0600); err != nil {
		t.Fatal(err)
	}
	if err := writeNewJSON(filepath.Join(root, "capability.json"), Capability{Version: 1, ID: id, Description: "Synthetic test capability", Subscriptions: subscriptions}); err != nil {
		t.Fatal(err)
	}
}

func TestCapabilityCheckBindingValidationAndSourceOnlyIsolation(t *testing.T) {
	s := fixtureStore(t)
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	i, err := Bind(s, filepath.Join(t.TempDir(), "instance with spaces"), "fixture", binary, "device-laptop")
	if err != nil {
		t.Fatal(err)
	}
	fixtureCapability(t, s, "context-probe", nil, "set -eu\ntest -z \"${MY_FRIDAY_INSTANCE-}\"\ntest -z \"$MY_FRIDAY_DEVICE_ID\"\n")
	manifest := filepath.Join(s.Root, "capabilities/context-probe/capability.json")
	var capability Capability
	if err := readJSON(manifest, &capability); err != nil {
		t.Fatal(err)
	}
	capability.Checks = [][]string{{"sh", "scripts/run.sh"}}
	if err := writeLocalJSON(manifest, capability); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MY_FRIDAY_INSTANCE", i.Root)
	t.Setenv("MY_FRIDAY_DEVICE_ID", i.DeviceID)
	if _, err := s.CheckCapability("context-probe"); err != nil {
		t.Fatalf("source-only check inherited a binding: %v", err)
	}
	forged := i
	forged.DeviceID = "device-forged"
	if _, err := forged.CheckCapability("context-probe"); err == nil || !strings.Contains(err.Error(), "binding changed") {
		t.Fatalf("forged binding not rejected before execution: %v", err)
	}
	changed := i
	changed.Name = "renamed-fixture"
	if err := writeLocalJSON(filepath.Join(i.Root, "binding.json"), changed); err != nil {
		t.Fatal(err)
	}
	if _, err := i.CheckCapability("context-probe"); err == nil || !strings.Contains(err.Error(), "binding changed") {
		t.Fatalf("stale binding not rejected before execution: %v", err)
	}
}

func TestHookRejectsTrailingOutput(t *testing.T) {
	for _, output := range []string{`{"additional_context":"ok"} {}`, `{"additional_context":"ok"} garbage`, `null`} {
		t.Run(output, func(t *testing.T) {
			s := fixtureStore(t)
			fixtureCapability(t, s, "capability-example", []Subscription{{ID: "handler", Event: "request.received", Command: []string{"sh", "scripts/run.sh"}}}, "#!/bin/sh\nprintf '%s' '"+output+"'\n")
			r, err := s.Dispatch(context.Background(), Event{Version: 1, ID: "event-invalid-output", Name: "request.received", AssistantID: s.Agent.ID, DeviceID: "device-laptop"})
			if err != nil || len(r.Handlers) != 1 || r.Handlers[0].Success || len(r.Context) != 0 {
				t.Fatalf("invalid output accepted: %+v %v", r, err)
			}
		})
	}
}

func TestHookCancellationStopsChainAndRefusesReplay(t *testing.T) {
	s := fixtureStore(t)
	fixtureCapability(t, s, "capability-example", []Subscription{
		{ID: "slow", Event: "request.received", Command: []string{"sh", "scripts/run.sh", "slow"}, Failure: "warn"},
		{ID: "later", Event: "request.received", Command: []string{"sh", "scripts/run.sh", "later"}, After: []string{"slow"}},
	}, "#!/bin/sh\nif [ \"$1\" = slow ]; then sleep 30; fi\nprintf '%s\\n' \"$1\" >> \"$MY_FRIDAY_ASSISTANT_ROOT/effects\"\n")
	event := Event{Version: 1, ID: "event-cancelled", Name: "request.received", AssistantID: s.Agent.ID, DeviceID: "device-laptop"}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	r, err := s.Dispatch(ctx, event)
	if err == nil || len(r.Handlers) != 1 || r.Handlers[0].Success {
		t.Fatalf("cancelled event claimed completion or continued chain: %+v %v", r, err)
	}
	if _, err := os.Stat(filepath.Join(s.Root, "effects")); !os.IsNotExist(err) {
		t.Fatal("cancelled chain executed later effects")
	}
	if _, err := s.Dispatch(context.Background(), event); err == nil {
		t.Fatal("cancelled event allowed replay")
	}
}

func TestHookWarnContinuesAndStopDoesNot(t *testing.T) {
	for _, failure := range []string{"warn", "stop"} {
		t.Run(failure, func(t *testing.T) {
			s := fixtureStore(t)
			fixtureCapability(t, s, "capability-example", []Subscription{
				{ID: "first", Event: "request.received", Command: []string{"sh", "scripts/run.sh", "first"}, Failure: failure},
				{ID: "second", Event: "request.received", Command: []string{"sh", "scripts/run.sh", "second"}, After: []string{"first"}},
			}, "#!/bin/sh\n[ \"$1\" = first ] && exit 1\nprintf '{\"additional_context\":\"continued\"}'\n")
			event := Event{Version: 1, ID: "event-chain", Name: "request.received", AssistantID: s.Agent.ID, DeviceID: "device-laptop"}
			r, err := s.Dispatch(context.Background(), event)
			if failure == "warn" {
				if err != nil || len(r.Handlers) != 2 || r.Handlers[0].Success || !r.Handlers[1].Success {
					t.Fatalf("warn chain: %+v %v", r, err)
				}
			} else if err == nil || len(r.Handlers) != 1 {
				t.Fatalf("stop chain: %+v %v", r, err)
			}
		})
	}
}

func TestCapabilityDiscoveryIncludesExistingInstructionPaths(t *testing.T) {
	s := fixtureStore(t)
	fixtureCapability(t, s, "capability-example", nil, "#!/bin/sh\nexit 0\n")
	root := filepath.Join(s.Root, "capabilities/capability-example")
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("Additional usage."), 0600); err != nil {
		t.Fatal(err)
	}
	caps, err := s.DiscoverCapabilities()
	if err != nil || len(caps) != 1 {
		t.Fatalf("discovery: %+v %v", caps, err)
	}
	if caps[0].Directory != root || len(caps[0].InstructionFiles) != 2 || caps[0].InstructionFiles[0] != filepath.Join(root, "instructions.md") || caps[0].InstructionFiles[1] != filepath.Join(root, "README.md") {
		t.Fatalf("missing direct instruction locations: %+v", caps[0])
	}
	// The installed runtime inventory must not add fields to persisted manifests.
	if _, err := s.Capabilities(); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(root, "instructions.md")); err != nil {
		t.Fatal(err)
	}
	caps, err = s.DiscoverCapabilities()
	if err != nil || len(caps[0].InstructionFiles) != 1 || caps[0].InstructionFiles[0] != filepath.Join(root, "README.md") {
		t.Fatalf("existing README capability not discoverable: %+v %v", caps, err)
	}
	if err := os.Symlink(filepath.Join(root, "README.md"), filepath.Join(root, "instructions.md")); err != nil {
		t.Fatal(err)
	}
	if _, err := s.DiscoverCapabilities(); err == nil {
		t.Fatal("symlink instructions accepted")
	}
}

func TestHookOrderStructuredInputAndDeduplication(t *testing.T) {
	s := fixtureStore(t)
	script := "#!/bin/sh\nread payload\nprintf '%s\\n' \"$1\" >> \"$MY_FRIDAY_ASSISTANT_ROOT/order\"\nprintf '{\"additional_context\":\"done\"}'\n"
	subscriptions := []Subscription{
		{ID: "second", Event: "request.received", Command: []string{"sh", "scripts/run.sh", "second"}, After: []string{"first"}},
		{ID: "first", Event: "request.received", Command: []string{"sh", "scripts/run.sh", "first"}},
	}
	fixtureCapability(t, s, "capability-example", subscriptions, script)
	event := Event{Version: 1, ID: "event-request-one", Name: "request.received", AssistantID: s.Agent.ID, DeviceID: "device-laptop", Payload: json.RawMessage(`{"prompt":"literal $(touch evil)"}`)}
	r, err := s.Dispatch(context.Background(), event)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Handlers) != 2 || len(r.Context) != 2 {
		t.Fatalf("missing dispatch results: %+v", r)
	}
	r, err = s.Dispatch(context.Background(), event)
	if err != nil || !r.Duplicate {
		t.Fatalf("duplicate not skipped: %+v %v", r, err)
	}
	b, err := os.ReadFile(filepath.Join(s.Root, "order"))
	if err != nil || string(b) != "first\nsecond\n" {
		t.Fatalf("order: %q %v", b, err)
	}
}

func TestHookCycleRefusedBeforeExecution(t *testing.T) {
	s := fixtureStore(t)
	fixtureCapability(t, s, "capability-example", []Subscription{
		{ID: "one", Event: "request.received", Command: []string{"sh", "scripts/run.sh"}, After: []string{"two"}},
		{ID: "two", Event: "request.received", Command: []string{"sh", "scripts/run.sh"}, After: []string{"one"}},
	}, "exit 0\n")
	_, err := s.Dispatch(context.Background(), Event{Version: 1, ID: "event-cycle", Name: "request.received", AssistantID: s.Agent.ID, DeviceID: "device-laptop"})
	if err == nil || !strings.Contains(err.Error(), "cycle") {
		t.Fatalf("cycle accepted: %v", err)
	}
}

func TestHookTimeoutAndFailureDoNotClaimSuccess(t *testing.T) {
	s := fixtureStore(t)
	fixtureCapability(t, s, "capability-example", []Subscription{
		{ID: "slow", Event: "request.received", Command: []string{"sh", "scripts/run.sh"}, TimeoutSeconds: 1, Failure: "stop"},
	}, "#!/bin/sh\nsleep 30\n")
	r, err := s.Dispatch(context.Background(), Event{Version: 1, ID: "event-timeout", Name: "request.received", AssistantID: s.Agent.ID, DeviceID: "device-laptop"})
	if err == nil || len(r.Handlers) != 1 || r.Handlers[0].Success {
		t.Fatalf("failure hidden: %+v %v", r, err)
	}
}

func TestHookReceiptsDoNotContainPayloadOrOutput(t *testing.T) {
	s := fixtureStore(t)
	fixtureCapability(t, s, "capability-example", []Subscription{{ID: "handler", Event: "request.received", Command: []string{"sh", "scripts/run.sh"}}}, "#!/bin/sh\nprintf 'unstructured-sensitive-value'\n")
	_, _ = s.Dispatch(context.Background(), Event{Version: 1, ID: "event-private", Name: "request.received", AssistantID: s.Agent.ID, DeviceID: "device-laptop", Payload: json.RawMessage(`{"prompt":"private-user-text"}`)})
	b, err := os.ReadFile(filepath.Join(s.Root, ".my-friday/local/events/event-private.json"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "private-user-text") || strings.Contains(string(b), "unstructured-sensitive-value") {
		t.Fatalf("raw hook data persisted: %s", b)
	}
}
