package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/acoz-labs/my-friday/internal/portable"
)

func apiFixture(t *testing.T) (portable.Instance, *portable.Store) {
	t.Helper()
	home := t.TempDir()
	s, err := portable.Create(filepath.Join(home, "source"), "fixture", "codex", "device-api", "Fixture")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.InitGit(context.Background()); err != nil {
		t.Fatal(err)
	}
	binary, _ := os.Executable()
	i, err := portable.Bind(s, filepath.Join(home, "instance"), "fixture", binary, "device-api")
	if err != nil {
		t.Fatal(err)
	}
	return i, s
}

func apiCall(t *testing.T, request string) (map[string]any, error) {
	t.Helper()
	var out, errout bytes.Buffer
	err := runPortable([]string{"api", "--input", "-"}, strings.NewReader(request), &out, &errout)
	if errout.Len() != 0 {
		t.Fatalf("unexpected stderr: %s", errout.String())
	}
	var r map[string]any
	if e := json.Unmarshal(out.Bytes(), &r); e != nil {
		t.Fatalf("non-JSON API output: %s / %v", out.String(), e)
	}
	return r, err
}

func TestAPIDiscoveryStrictInputAndNoPrompt(t *testing.T) {
	list, _ := json.Marshal(map[string]any{"schema_version": 1, "action": "agent.list", "params": map[string]any{"directory": t.TempDir()}})
	for _, request := range []string{`{"schema_version":1,"action":"system.describe"}`, string(list)} {
		r, err := apiCall(t, request)
		if err != nil || r["ok"] != true {
			t.Fatalf("%+v %v", r, err)
		}
	}
	for _, request := range []string{
		`{"schema_version":999,"action":"agent.list"}`,
		`{"schema_version":1,"action":"agent.list","surprise":"private-canary"}`,
		`{"schema_version":1,"action":"agent.list","params":{"unknown":"private-canary"}}`,
		`{"schema_version":1,"action":"source-credential"}`,
		`{"schema_version":1,"action":"agent.list"} {}`,
		`{"schema_version":1,"action":"agent.list","apply":false,"apply":true}`,
		`{"schema_version":1,"action":"agent.list","apply":false,"Apply":true}`,
		`{"schema_version":1,"action":"agent.list","params":null}`,
		`{"schema_version":1,"action":"agent.list","apply":null}`,
		`{"schema_version":1,"action":"agent.list","params":{"directory":"/first","directory":"/second"}}`,
	} {
		r, err := apiCall(t, request)
		if err == nil || r["ok"] != false {
			t.Fatalf("accepted invalid request: %+v %v", r, err)
		}
		data, _ := json.Marshal(r)
		if bytes.Contains(data, []byte("private-canary")) {
			t.Fatal("echoed input")
		}
	}
}

func TestAPIPreviewRepairAndHarnessChange(t *testing.T) {
	i, s := apiFixture(t)
	target := filepath.Join(i.Root, "codex/hooks.json")
	os.Remove(target)
	native := filepath.Join(i.Root, "codex/config.toml")
	before, _ := os.ReadFile(native)
	request := map[string]any{"schema_version": 1, "action": "agent.repair", "params": map[string]any{"instance": i.Root}}
	data, _ := json.Marshal(request)
	r, err := apiCall(t, string(data))
	if err != nil || r["state"] != "planned" {
		t.Fatalf("%+v %v", r, err)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatal("preview repaired files")
	}
	request["apply"] = true
	data, _ = json.Marshal(request)
	r, err = apiCall(t, string(data))
	if err != nil || r["state"] != "completed" {
		t.Fatalf("%+v %v", r, err)
	}
	after, _ := os.ReadFile(native)
	if !bytes.Equal(before, after) {
		t.Fatal("native settings changed")
	}
	request["action"] = "agent.harness"
	request["params"] = map[string]any{"instance": i.Root, "harness": "pi"}
	data, _ = json.Marshal(request)
	if _, err := apiCall(t, string(data)); err != nil {
		t.Fatal(err)
	}
	fresh, err := portable.Open(s.Root)
	if err != nil || fresh.Agent.DefaultHarness != "pi" {
		t.Fatal("harness not changed")
	}
}

func TestAPISelfUpgradeAndUnhealthyDoctor(t *testing.T) {
	i, _ := apiFixture(t)
	t.Setenv("MY_FRIDAY_INSTANCE", i.Root)
	data, _ := json.Marshal(map[string]any{"schema_version": 1, "action": "toolkit.use", "apply": true, "params": map[string]any{"instance": i.Root, "sessions_stopped": true}})
	r, err := apiCall(t, string(data))
	if err == nil || r["ok"] != false {
		t.Fatal("live self-upgrade accepted")
	}
	os.Remove(filepath.Join(i.Root, "codex/hooks.json"))
	data, _ = json.Marshal(map[string]any{"schema_version": 1, "action": "agent.doctor", "params": map[string]any{"instance": i.Root}})
	r, err = apiCall(t, string(data))
	if err == nil || r["state"] != "needs_attention" || r["result"] == nil {
		t.Fatalf("%+v %v", r, err)
	}
}
