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
