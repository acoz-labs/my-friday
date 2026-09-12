package memorycodex

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestPackagedRunnerPinsRuntimeAndDoesNotBlockOnFailure(t *testing.T) {
	plugin, err := filepath.Abs("../../plugins/codex/plugins/my-friday-memory")
	if err != nil {
		t.Fatal(err)
	}
	var hooks struct {
		Hooks map[string][]struct{ Hooks []struct{ Command string } }
	}
	b, err := os.ReadFile(filepath.Join(plugin, "hooks/hooks.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, &hooks); err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	old := filepath.Join(root, "old")
	if err := os.Mkdir(old, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(old, "my-friday"), []byte("#!/bin/sh\necho LEGACY-OUTPUT >&2\nexit 2\n"), 0700); err != nil {
		t.Fatal(err)
	}
	good := filepath.Join(root, "pinned runtime with spaces")
	if err := os.WriteFile(good, []byte("#!/bin/sh\n[ \"$1\" = codex-memory-hook ] || exit 9\nprintf '%s\\n' '{\"systemMessage\":\"PINNED-RUNTIME\"}'\n"), 0700); err != nil {
		t.Fatal(err)
	}
	for _, event := range []string{"SessionStart", "UserPromptSubmit"} {
		for _, tc := range []struct{ name, binary, want string }{
			{"pinned despite old PATH", good, "PINNED-RUNTIME"},
			{"old binary fails open", filepath.Join(old, "my-friday"), "unavailable"},
			{"missing binary fails open", filepath.Join(root, "missing"), "unavailable"},
			{"relative override rejected", "my-friday", "absolute"},
			{"default old PATH fails open", "", "unavailable"},
		} {
			t.Run(event+"/"+tc.name, func(t *testing.T) {
				cmd := exec.Command("/bin/sh", "-c", hooks.Hooks[event][0].Hooks[0].Command)
				cmd.Env = []string{"PATH=" + old, "PLUGIN_ROOT=" + plugin, "MY_FRIDAY_MEMORY_BIN=" + tc.binary}
				cmd.Stdin = strings.NewReader(`{"hook_event_name":"` + event + `"}`)
				out, err := cmd.CombinedOutput()
				if err != nil {
					t.Fatalf("hook blocked: %v %s", err, out)
				}
				var reply map[string]any
				if err := json.Unmarshal(out, &reply); err != nil || !strings.Contains(string(out), tc.want) || strings.Contains(string(out), "LEGACY-OUTPUT") || reply["decision"] != nil {
					t.Fatalf("unexpected hook output: %s (%v)", out, err)
				}
			})
		}
	}
}

func TestPackagedMCPRunnerUsesSamePinnedRuntime(t *testing.T) {
	plugin, err := filepath.Abs("../../plugins/codex/plugins/my-friday-memory")
	if err != nil {
		t.Fatal(err)
	}
	var config struct {
		Servers map[string]struct {
			Command string
			Args    []string
			Cwd     string
			EnvVars []string `json:"env_vars"`
		} `json:"mcpServers"`
	}
	b, err := os.ReadFile(filepath.Join(plugin, ".mcp.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, &config); err != nil {
		t.Fatal(err)
	}
	spec := config.Servers["my-friday-memory"]
	if !slices.Contains(spec.EnvVars, "MY_FRIDAY_MEMORY_BIN") {
		t.Fatal("runtime selection is not forwarded")
	}
	runtime := filepath.Join(t.TempDir(), "pinned runtime")
	if err := os.WriteFile(runtime, []byte("#!/bin/sh\n[ \"$1 $2 $3\" = 'mcp --harness codex' ] || exit 9\nIFS= read -r line\nprintf '%s\\n' \"$line\"\n"), 0700); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(spec.Command, spec.Args...)
	cmd.Dir = filepath.Join(plugin, spec.Cwd)
	cmd.Env = []string{"PATH=/nonexistent", "MY_FRIDAY_MEMORY_BIN=" + runtime}
	cmd.Stdin = strings.NewReader("STDIO-PRESERVED\n")
	out, err := cmd.CombinedOutput()
	if err != nil || string(out) != "STDIO-PRESERVED\n" {
		t.Fatalf("MCP launch: %s %v", out, err)
	}
}
