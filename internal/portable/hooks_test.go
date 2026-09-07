package portable

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
