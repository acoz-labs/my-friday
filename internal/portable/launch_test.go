package portable

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLaunchKeepsProjectAndSeparatesHarnessHomes(t *testing.T) {
	s := fixtureStore(t)
	state := filepath.Join(t.TempDir(), "instance")
	binary := filepath.Join(t.TempDir(), "my friday's binary")
	instance, err := Bind(s, state, "friday", binary, "device-laptop")
	if err != nil {
		t.Fatal(err)
	}
	project := t.TempDir()
	t.Setenv("CODEX_HOME", "/ambient/codex")
	t.Setenv("PI_CODING_AGENT_DIR", "/ambient/pi")
	t.Setenv("GIT_DIR", "/ambient/git")
	plan, err := instance.Plan(s, "pi", project, []string{"literal $(touch nope)", "two words"})
	if err != nil {
		t.Fatal(err)
	}
	if plan.Directory != project || plan.Executable != "pi" || strings.Join(plan.Arguments, "|") != "literal $(touch nope)|two words" {
		t.Fatalf("launch changed request: %+v", plan)
	}
	env := strings.Join(plan.Environment, "\n")
	if strings.Contains(env, "/ambient/") || !strings.Contains(env, "PI_CODING_AGENT_DIR="+filepath.Join(state, "pi")) || !strings.Contains(env, "MY_FRIDAY_ASSISTANT_ROOT="+s.Root) {
		t.Fatalf("wrong instance environment: %s", env)
	}
	if err = instance.Project(s); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(state, "codex/hooks.json"))
	if err != nil {
		t.Fatal(err)
	}
	var hooks struct {
		Hooks map[string]any `json:"hooks"`
	}
	if err = json.Unmarshal(b, &hooks); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"SessionStart", "UserPromptSubmit", "Stop", "PreToolUse", "PostToolUse", "SessionEnd"} {
		if _, ok := hooks.Hooks[name]; !ok {
			t.Fatalf("missing native event %s", name)
		}
	}
	instructions, _ := os.ReadFile(filepath.Join(state, "codex/AGENTS.md"))
	if !strings.Contains(string(instructions), "memory recall") || !strings.Contains(string(instructions), s.Root) {
		t.Fatal("memory unavailable outside assistant directory")
	}
	for _, expected := range []string{"agent capability-guide", "agent capability-template", "MY_FRIDAY_BIN", "MY_FRIDAY_ASSISTANT_ROOT", "memory scopes", "Do not guess scope IDs", "instruction_files", "An empty recall is not evidence", "memory write --input -", "fresh private temporary directory", "clean up only files you created"} {
		if !strings.Contains(string(instructions), expected) {
			t.Errorf("missing authoring guidance: %s", expected)
		}
	}
	if !strings.Contains(env, "MY_FRIDAY_BIN="+instance.Binary) {
		t.Fatal("portable CLI path unavailable to capabilities")
	}
	for _, expected := range []string{"Native user-wide and project-local skills", "host/project dependencies", "verify its required tools are available"} {
		if !strings.Contains(string(instructions), expected) {
			t.Errorf("missing inheritance guidance: %s", expected)
		}
	}
	codexPlan, err := instance.Plan(s, "codex", project, []string{"app-server"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(codexPlan.Arguments, "|") != "--dangerously-bypass-approvals-and-sandbox|--dangerously-bypass-hook-trust|app-server" {
		t.Fatalf("unexpected native discovery override: %q", codexPlan.Arguments)
	}
	if _, err = Bind(s, state, "other", binary, "device-laptop"); err == nil {
		t.Fatal("overwrote another instance")
	}
}

func TestNativeEventNormalizationDoesNotConfuseTurnsAndTasks(t *testing.T) {
	if NormalizeEvent("codex", "Stop") != "request.completed" {
		t.Fatal("Codex stop mapping")
	}
	if NormalizeEvent("pi", "turn_end") == "task.completed" || NormalizeEvent("pi", "turn_end") == "request.completed" {
		t.Fatal("Pi model turn promoted into task completion")
	}
	if NormalizeEvent("pi", "before_agent_start") != "request.received" {
		t.Fatal("Pi request mapping")
	}
	if NormalizeEvent("pi", "future_event") != "native.pi.future_event" {
		t.Fatal("unknown native semantics lost")
	}
}

func TestLauncherQuotesPathsAndForwardsArguments(t *testing.T) {
	s := fixtureStore(t)
	state := filepath.Join(t.TempDir(), "instance with ' quote")
	instance, err := Bind(s, state, "friday", "/path with ' quote/my-friday", "device-laptop")
	if err != nil {
		t.Fatal(err)
	}
	launcher := filepath.Join(t.TempDir(), "friday")
	if err = instance.InstallLauncher(launcher); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(launcher)
	if !strings.Contains(string(b), `"$@"`) || !strings.Contains(string(b), `'\''`) {
		t.Fatalf("unsafe launcher: %s", b)
	}
	if err = instance.InstallLauncher(launcher); err == nil {
		t.Fatal("launcher collision overwritten")
	}
}
