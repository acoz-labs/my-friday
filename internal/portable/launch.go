package portable

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Instance struct {
	Version     int    `json:"schema_version"`
	Name        string `json:"name"`
	AssistantID string `json:"assistant_id"`
	Repository  string `json:"repository"`
	DeviceID    string `json:"device_id"`
	Binary      string `json:"binary"`
	Root        string `json:"-"`
}
type LaunchPlan struct {
	Executable  string
	Arguments   []string
	Directory   string
	Environment []string
}

func ValidateLauncherName(name string) error {
	if !identifier.MatchString(name) {
		return errors.New("launcher name must be 3-128 lowercase letters, digits, or hyphens, starting with a letter")
	}
	return nil
}

func Bind(s *Store, state, name, binary, deviceID string) (Instance, error) {
	result := Instance{Version: 1, Name: name, AssistantID: s.Agent.ID, Repository: s.Root, DeviceID: deviceID, Binary: binary}
	if err := ValidateLauncherName(name); err != nil {
		return result, err
	}
	if err := s.deviceExists(deviceID); err != nil {
		return result, err
	}
	abs, err := filepath.Abs(state)
	if err != nil {
		return result, err
	}
	result.Root = abs
	result.Binary, err = filepath.Abs(binary)
	if err != nil {
		return result, err
	}
	if err = os.MkdirAll(filepath.Dir(abs), 0700); err != nil {
		return result, err
	}
	if err = os.Mkdir(abs, 0700); err != nil {
		return result, fmt.Errorf("instance already exists or cannot be created: %w", err)
	}
	if err = writeNewJSON(filepath.Join(abs, "binding.json"), result); err != nil {
		return result, err
	}
	return result, result.Project(s)
}

func LoadInstance(path string) (Instance, *Store, error) {
	result := Instance{Root: path}
	if err := readJSON(filepath.Join(path, "binding.json"), &result); err != nil {
		return result, nil, err
	}
	if result.Version != 1 || !identifier.MatchString(result.Name) || !filepath.IsAbs(result.Repository) || !filepath.IsAbs(result.Binary) {
		return result, nil, errors.New("invalid instance binding")
	}
	s, err := Open(result.Repository)
	if err != nil {
		return result, nil, err
	}
	if result.AssistantID != s.Agent.ID {
		return result, nil, errors.New("instance/assistant identity mismatch")
	}
	if err = s.deviceExists(result.DeviceID); err != nil {
		return result, nil, err
	}
	return result, s, nil
}

func shellQuote(value string) string { return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'" }

func (i Instance) InstallLauncher(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0700)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString("#!/bin/sh\nexec " + shellQuote(i.Binary) + " agent launch --instance " + shellQuote(i.Root) + " \"$@\"\n")
	if err != nil {
		return err
	}
	return f.Sync()
}

func (i Instance) Project(s *Store) error {
	instructions := []string{"# My Friday assistant\n\nAssistant repository: " + s.Root + "\n"}
	for _, name := range []string{"identity.md", "operating.md"} {
		path := filepath.Join(s.Root, "instructions", name)
		info, err := os.Lstat(path)
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return errors.New("instructions must be regular files")
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		instructions = append(instructions, string(b))
	}
	instructions = append(instructions, fmt.Sprintf(`## Memory and capabilities

Your working directory is the user's project. Your assistant repository is %s.
Use the absolute My Friday command %s; do not infer assistant paths from cwd.
The launch environment supplies MY_FRIDAY_ASSISTANT_ROOT and MY_FRIDAY_DEVICE_ID.
MY_FRIDAY_BIN is the absolute toolkit executable; use it to keep reusable
instructions and scripts portable rather than embedding installation paths.

Before substantial work, use "memory recall --query TEXT" and select an explicit
--scope-kind/--scope-id for account, project, or task-specific guidance. Use
"memory history --record ID" to explain changes. Scope is not inferred from a
similarity score. A conflicting revision is evidence to reconcile, not guidance.

To learn, use "memory write --input FILE" with a memory revision document. Use
"memory template" for its shape; the writer stamps device, actor, harness, and
time. Preserve record/scope IDs and name predecessor revision IDs in supersedes.
Use "memory source --summary TEXT --kind user-direction" to record concise
evidence, and "memory event --summary TEXT --kind task-completed" for chronology.
Record meaningful outcomes and reusable learning before finishing the task.
One-task exceptions use task scope. Never copy secrets or raw transcripts.

Use "agent capabilities" to list portable capabilities, then read the selected
capability's complete instructions and requirements. Capabilities can contain
scripts and hook subscriptions. Before authoring a capability, read the built-in
"agent capability-guide" and use "agent capability-template --capability ID".
These commands work without a source checkout. Use --help for CLI discovery;
do not inspect executable strings or search for a development checkout to learn
the format. Use "agent check --capability ID" after changing
one. Repository Git history is versioned; external effects require their own
reconciliation. User corrections override older remembered user guidance within
their actual scope. Native project instructions still apply to project work.
`, s.Root, i.Binary))
	text := strings.Join(instructions, "\n\n")
	for _, harness := range []string{"codex", "pi"} {
		root := filepath.Join(i.Root, harness)
		if err := os.MkdirAll(root, 0700); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte(text), 0600); err != nil {
			return err
		}
	}
	config := "approval_policy = \"never\"\nsandbox_mode = \"danger-full-access\"\n[features]\nhooks = true\n"
	if err := os.WriteFile(filepath.Join(i.Root, "codex/config.toml"), []byte(config), 0600); err != nil {
		return err
	}
	hooks := map[string]any{}
	for _, event := range CodexEvents {
		timeout := 20
		if event == "SessionEnd" || event == "Interrupt" {
			timeout = 3
		}
		command := shellQuote(i.Binary) + " hook --instance " + shellQuote(i.Root) + " --harness codex --native " + shellQuote(event)
		hooks[event] = []any{map[string]any{"hooks": []any{map[string]any{"type": "command", "command": command, "timeout": timeout}}}}
	}
	if err := writeLocalJSON(filepath.Join(i.Root, "codex/hooks.json"), map[string]any{"hooks": hooks}); err != nil {
		return err
	}
	dir := filepath.Join(i.Root, "pi/extensions")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	binary, _ := json.Marshal(i.Binary)
	instance, _ := json.Marshal(i.Root)
	extension := `import { execFile } from "node:child_process";
import { randomUUID } from "node:crypto";

export default function (pi: any) {
  const events = ["session_start", "session_shutdown", "before_agent_start", "agent_start", "agent_end", "agent_settled", "turn_start", "turn_end", "tool_call", "tool_result", "session_before_compact", "session_compact", "model_select"];
  for (const name of events) {
    pi.on(name, async (event: any, ctx: any) => {
      const input = JSON.stringify({ event_id: "event-" + randomUUID(), session_id: ctx.sessionManager?.getSessionId?.(), prompt: event.prompt, tool_name: event.toolName, cwd: ctx.cwd });
      const result: any = await new Promise((resolve) => {
        const child = execFile(BINARY, ["hook", "--instance", INSTANCE, "--harness", "pi", "--native", name], { timeout: 20000, maxBuffer: 65536 }, (error, stdout) => {
          if (error) { resolve({ error: "My Friday hook failed; local changes remain available." }); return; }
          try { resolve(JSON.parse(stdout || "{}")); } catch { resolve({ error: "Invalid My Friday hook response." }); }
        });
        child.stdin?.end(input);
      });
      if (result.error) { ctx.ui?.notify?.(result.error, "warning"); return; }
      if (name === "before_agent_start" && result.additional_context) {
        return { message: { customType: "my-friday-memory", content: result.additional_context, display: false } };
      }
    });
  }
}
`
	extension = strings.ReplaceAll(extension, "BINARY", string(binary))
	extension = strings.ReplaceAll(extension, "INSTANCE", string(instance))
	return os.WriteFile(filepath.Join(dir, "my-friday.ts"), []byte(extension), 0600)
}

func (i Instance) Plan(s *Store, harness, cwd string, args []string) (LaunchPlan, error) {
	if harness == "" {
		harness = s.Agent.DefaultHarness
	}
	if harness != "codex" && harness != "pi" {
		return LaunchPlan{}, errors.New("harness must be codex or pi")
	}
	env := []string{}
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if strings.HasPrefix(key, "MY_FRIDAY_") || strings.HasPrefix(key, "GIT_") || key == "CODEX_HOME" || key == "PI_CODING_AGENT_DIR" || key == "PI_CODING_AGENT_SESSION_DIR" {
			continue
		}
		env = append(env, entry)
	}
	env = append(env, "MY_FRIDAY_ASSISTANT_ROOT="+s.Root, "MY_FRIDAY_BIN="+i.Binary, "MY_FRIDAY_DEVICE_ID="+i.DeviceID, "MY_FRIDAY_INSTANCE="+i.Root, "MY_FRIDAY_HARNESS="+harness, "MY_FRIDAY_SESSION_ID="+NewID("session"))
	if harness == "codex" {
		env = append(env, "CODEX_HOME="+filepath.Join(i.Root, "codex"))
		args = append([]string{"--dangerously-bypass-approvals-and-sandbox", "--dangerously-bypass-hook-trust"}, args...)
	} else {
		env = append(env, "PI_CODING_AGENT_DIR="+filepath.Join(i.Root, "pi"))
	}
	return LaunchPlan{Executable: harness, Arguments: args, Directory: cwd, Environment: env}, nil
}

var CodexEvents = []string{"SessionStart", "SessionEnd", "UserPromptSubmit", "PreToolUse", "PermissionRequest", "PostToolUse", "PreCompact", "PostCompact", "SubagentStart", "SubagentStop", "Stop", "Interrupt"}

func NormalizeEvent(harness, native string) string {
	mappings := map[string]map[string]string{
		"codex": {"SessionStart": "session.started", "SessionEnd": "session.ending", "UserPromptSubmit": "request.received", "PreToolUse": "tool.before", "PostToolUse": "tool.after", "PreCompact": "context.compacting", "PostCompact": "context.compacted", "Stop": "request.completed", "Interrupt": "request.interrupted"},
		"pi":    {"session_start": "session.started", "session_shutdown": "session.ending", "before_agent_start": "request.received", "agent_settled": "request.completed", "tool_call": "tool.before", "tool_result": "tool.after", "session_before_compact": "context.compacting", "session_compact": "context.compacted"},
	}
	if name := mappings[harness][native]; name != "" {
		return name
	}
	return "native." + harness + "." + native
}
