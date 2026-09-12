package memorycodex

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/acoz-labs/my-friday/internal/memorybank"
	"github.com/acoz-labs/my-friday/internal/portable"
)

// This test installs the real packaged plugin into disposable native state.
// It never copies authentication, starts a model turn, or changes ambient config.
func TestNativeCodexMemoryPlugin(t *testing.T) {
	codex, friday := os.Getenv("FRIDAY_TEST_CODEX"), os.Getenv("FRIDAY_TEST_MEMORY_BINARY")
	if codex == "" || friday == "" {
		t.Skip("set FRIDAY_TEST_CODEX and FRIDAY_TEST_MEMORY_BINARY to absolute executables")
	}
	if !filepath.IsAbs(codex) || !filepath.IsAbs(friday) || filepath.Base(friday) != "my-friday" {
		t.Fatal("absolute executables required; memory binary must be named my-friday")
	}
	root := t.TempDir()
	nativeHome, project := filepath.Join(root, "codex"), filepath.Join(root, "project")
	for _, dir := range []string{nativeHome, project} {
		if err := os.Mkdir(dir, 0700); err != nil {
			t.Fatal(err)
		}
	}
	store, err := portable.CreateMemoryBank(filepath.Join(root, "bank"), "Example", "device-fixture", "Synthetic machine")
	if err != nil {
		t.Fatal(err)
	}
	binding := filepath.Join(root, "binding.json")
	if _, err := memorybank.Bind(store.Root, binding, "Synthetic native host", "Test actor"); err != nil {
		t.Fatal(err)
	}
	service, err := memorybank.OpenBinding(binding, "test")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Remember(memorybank.Write{Kind: "decision", Summary: "Fictional project name", Body: "Silver Heron", Basis: "user-direction", Reason: "Synthetic native acceptance"}); err != nil {
		t.Fatal(err)
	}
	env := []string{}
	for _, v := range os.Environ() {
		key, _, _ := strings.Cut(v, "=")
		if key == "PATH" || strings.HasPrefix(key, "CODEX_") || strings.HasPrefix(key, "OPENAI_") || strings.HasPrefix(key, "MY_FRIDAY_") {
			continue
		}
		env = append(env, v)
	}
	// Deliberately do not add the candidate to PATH: explicit runtime selection
	// must work even when the native shell finds an older installation first.
	env = append(env, "CODEX_HOME="+nativeHome, "MY_FRIDAY_MEMORY_BINDING="+binding, "MY_FRIDAY_MEMORY_BIN="+friday, "PATH="+os.Getenv("PATH"))
	ctx, cancel := context.WithTimeout(context.Background(), 55*time.Second)
	defer cancel()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.CommandContext(ctx, codex, args...)
		cmd.Env, cmd.Dir = env, project
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("native %v: %v\n%s", args, err, output)
		}
	}
	marketplace, err := filepath.Abs("../../plugins/codex")
	if err != nil {
		t.Fatal(err)
	}
	run("plugin", "marketplace", "add", marketplace)
	run("plugin", "add", "my-friday-memory@personal")
	before, err := os.ReadFile(filepath.Join(nativeHome, "config.toml"))
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.CommandContext(ctx, codex, "app-server")
	cmd.Env, cmd.Dir = env, project
	var nativeErrors bytes.Buffer
	cmd.Stderr = &nativeErrors
	in, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { in.Close(); cancel(); _ = cmd.Wait() }()
	enc := json.NewEncoder(in)
	scanner := bufio.NewScanner(out)
	scanner.Buffer(make([]byte, 65536), 4*1024*1024)
	nextID := 0
	request := func(method string, params any) json.RawMessage {
		t.Helper()
		nextID++
		if err := enc.Encode(map[string]any{"id": nextID, "method": method, "params": params}); err != nil {
			t.Fatal(err)
		}
		for scanner.Scan() {
			var reply struct {
				ID     int             `json:"id"`
				Result json.RawMessage `json:"result"`
				Error  json.RawMessage `json:"error"`
			}
			if json.Unmarshal(scanner.Bytes(), &reply) == nil && reply.ID == nextID {
				if len(reply.Error) != 0 {
					t.Fatalf("%s: %s", method, reply.Error)
				}
				return reply.Result
			}
		}
		t.Fatalf("%s ended without reply: %v", method, scanner.Err())
		return nil
	}
	request("initialize", map[string]any{"clientInfo": map[string]string{"name": "friday_memory_test", "version": "1"}, "capabilities": map[string]bool{"experimentalApi": true}})
	if err := enc.Encode(map[string]string{"method": "initialized"}); err != nil {
		t.Fatal(err)
	}
	skills := request("skills/list", map[string]any{"cwds": []string{project}, "forceReload": true})
	if !strings.Contains(string(skills), `"name":"my-friday-memory:memory"`) {
		t.Fatalf("memory skill not discovered: %s", skills)
	}
	hooks := request("hooks/list", map[string]any{"cwds": []string{project}})
	if !strings.Contains(string(hooks), "scripts/run-memory.sh") || !strings.Contains(string(hooks), "userPromptSubmit") {
		t.Fatalf("native hook not discovered: %s", hooks)
	}
	started := request("thread/start", map[string]any{"cwd": project, "approvalPolicy": "never", "ephemeral": true})
	var thread struct {
		Thread struct {
			ID string `json:"id"`
		} `json:"thread"`
	}
	if err := json.Unmarshal(started, &thread); err != nil || thread.Thread.ID == "" {
		t.Fatalf("thread startup: %s %v", started, err)
	}
	status := request("mcpServerStatus/list", map[string]any{"threadId": thread.Thread.ID})
	var inventory struct {
		Data []struct {
			Name  string                     `json:"name"`
			Tools map[string]json.RawMessage `json:"tools"`
		} `json:"data"`
	}
	if err := json.Unmarshal(status, &inventory); err != nil {
		t.Fatal(err)
	}
	server := ""
	for _, v := range inventory.Data {
		if _, ok := v.Tools["memory_recall"]; ok {
			server = v.Name
		}
	}
	if server == "" {
		// Stop before reading captured stderr so the buffer has no live writer.
		in.Close()
		cancel()
		_ = cmd.Wait()
		t.Fatalf("MCP memory server not ready: %s\n%s", status, nativeErrors.String())
	}
	result := request("mcpServer/tool/call", map[string]any{"threadId": thread.Thread.ID, "server": server, "tool": "memory_recall", "arguments": map[string]string{"query": "fictional project"}})
	if !strings.Contains(string(result), "Silver Heron") {
		t.Fatalf("native MCP recall failed: %s", result)
	}
	after, err := os.ReadFile(filepath.Join(nativeHome, "config.toml"))
	if err != nil || string(before) != string(after) {
		t.Fatal("memory runtime changed native configuration")
	}
	if _, err := os.Stat(filepath.Join(nativeHome, "auth.json")); !os.IsNotExist(err) {
		t.Fatal("authentication unexpectedly created")
	}
}
