package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestCompiledAgentAPIWorkflow(t *testing.T) {
	dir := t.TempDir()
	binary := filepath.Join(dir, "my-friday")
	build := exec.Command("go", "build", "-o", binary, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("%s %v", out, err)
	}
	project := filepath.Join(dir, "project")
	os.Mkdir(project, 0700)
	bin := filepath.Join(dir, "bin")
	os.Mkdir(bin, 0700)
	os.WriteFile(filepath.Join(bin, "codex"), []byte("#!/bin/sh\nexit 0\n"), 0700)
	run := func(request map[string]any, wantExit int) map[string]any {
		t.Helper()
		input, _ := json.Marshal(request)
		cmd := exec.Command(binary, "api", "--input", "-")
		cmd.Dir = project
		cmd.Env = []string{"PATH=" + bin + ":/usr/bin:/bin", "TERM=dumb"}
		cmd.Stdin = bytes.NewReader(input)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		exit := 0
		if err != nil {
			failure, ok := err.(*exec.ExitError)
			if !ok {
				t.Fatal(err)
			}
			exit = failure.ExitCode()
		}
		if exit != wantExit || stderr.Len() != 0 {
			t.Fatalf("exit %d wanted %d: %s / %s", exit, wantExit, stdout.String(), stderr.String())
		}
		var response map[string]any
		if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
			t.Fatalf("not one JSON response: %s", stdout.String())
		}
		return response
	}
	describe := run(map[string]any{"schema_version": 1, "action": "system.describe"}, 0)
	if len(describe["result"].(map[string]any)["actions"].([]any)) < 10 {
		t.Fatal("missing discovery contract")
	}
	root := filepath.Join(dir, "source")
	instance := filepath.Join(dir, "instance")
	params := map[string]any{"repository": root, "instance": instance, "name": "fixture", "device_label": "Fixture laptop", "no_launcher": true}
	request := map[string]any{"schema_version": 1, "id": "creation-preview", "action": "agent.setup", "params": params}
	if r := run(request, 0); r["state"] != "planned" {
		t.Fatal(r)
	}
	if _, err := os.Stat(root); !os.IsNotExist(err) {
		t.Fatal("preview created source")
	}
	request["apply"] = true
	request["id"] = "creation-apply"
	run(request, 0)
	request = map[string]any{"schema_version": 1, "action": "agent.doctor", "params": map[string]any{"instance": instance}}
	run(request, 0)
	os.Remove(filepath.Join(instance, "codex/hooks.json"))
	if r := run(request, 3); r["state"] != "needs_attention" || r["error"].(map[string]any)["code"] != "installation.unhealthy" {
		t.Fatal(r)
	}
	request["action"] = "agent.repair"
	request["apply"] = true
	if r := run(request, 0); r["requires_fresh_session"] != true {
		t.Fatal("repair omitted context notice")
	}
	request["action"] = "source.sync"
	if r := run(request, 0); r["state"] != "local_only" {
		t.Fatal("local-only falsely reported synced")
	}
	remote := filepath.Join(dir, "remote.git")
	git := exec.Command("git", "init", "--bare", "--initial-branch=main", remote)
	if out, err := git.CombinedOutput(); err != nil {
		t.Fatalf("%s %v", out, err)
	}
	request["action"] = "source.configure"
	request["params"] = map[string]any{"instance": instance, "mode": "existing", "remote": remote}
	if r := run(request, 0); r["state"] != "synced" {
		t.Fatal(r)
	}
	files, err := os.ReadDir(project)
	if err != nil || len(files) != 0 {
		t.Fatal("API wrote to caller project")
	}
	run(map[string]any{"schema_version": 1, "action": "agent.setup", "params": map[string]any{"unknown": "ignored?"}}, 2)
}
