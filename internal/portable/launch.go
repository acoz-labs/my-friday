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
	if err := ValidateInstallationPaths(s.Root, state, ""); err != nil {
		return result, err
	}
	if err := ValidateLauncherName(name); err != nil {
		return result, err
	}
	if err := s.deviceExists(deviceID); err != nil {
		return result, err
	}
	abs, err := prospectivePath(state)
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
	info, err := os.Lstat(path)
	if err != nil {
		return result, nil, err
	}
	if !info.IsDir() {
		return result, nil, errors.New("instance must be a directory, not a symlink or file")
	}
	result.Root, err = prospectivePath(path)
	if err != nil {
		return result, nil, err
	}
	if err := readJSON(filepath.Join(path, "binding.json"), &result); err != nil {
		return result, nil, err
	}
	if result.Version != 1 || !identifier.MatchString(result.Name) || !filepath.IsAbs(result.Repository) || !filepath.IsAbs(result.Binary) {
		return result, nil, errors.New("invalid instance binding")
	}
	if err := ValidateInstallationPaths(result.Repository, path, ""); err != nil {
		return result, nil, err
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
	if err := ValidateInstallationPaths(i.Repository, i.Root, path); err != nil {
		return err
	}
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
	if err := i.preflightProjection(s); err != nil {
		return err
	}
	files, err := i.projection(s)
	if err != nil {
		return err
	}
	for _, dir := range projectionDirectories {
		if err := os.MkdirAll(filepath.Join(i.Root, dir), 0700); err != nil {
			return err
		}
	}
	if err := seedCodexConfig(filepath.Join(i.Root, "codex/config.toml")); err != nil {
		return err
	}
	for _, name := range projectionFiles {
		if err := replaceProjectionFile(filepath.Join(i.Root, name), files[name]); err != nil {
			return err
		}
	}
	return nil
}

// Render without side effects so diagnostics and repair compare the same output.
func (i Instance) projection(s *Store) (map[string][]byte, error) {
	files := map[string][]byte{}
	instructions := []string{"# My Friday assistant\n\nAssistant repository: " + s.Root + "\n"}
	for _, name := range []string{"identity.md", "operating.md"} {
		path := filepath.Join(s.Root, "instructions", name)
		info, err := os.Lstat(path)
		if err != nil {
			return nil, err
		}
		if !info.Mode().IsRegular() {
			return nil, errors.New("instructions must be regular files")
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		instructions = append(instructions, string(b))
	}
	instructions = append(instructions, fmt.Sprintf(`## Memory and capabilities

Your working directory is the user's project. Your assistant repository is %s.
Use the absolute My Friday command %s; do not infer assistant paths from cwd.
The launch environment supplies MY_FRIDAY_ASSISTANT_ROOT and MY_FRIDAY_DEVICE_ID.
MY_FRIDAY_BIN is the absolute toolkit executable; use it to keep reusable
instructions and scripts portable rather than embedding installation paths.

For installation/management automation, start with "$MY_FRIDAY_BIN" api describe.
It lists versioned actions, typed parameters and side effects. Submit one JSON
request with api --input - (stdin) or --input FILE; never navigate the human TUI
or simulate menu keystrokes. Management targets use explicit absolute instance
paths; MY_FRIDAY_INSTANCE identifies this current instance. Read-only actions run
directly; mutations default to a parameter-only preview until apply=true. That
flag expresses an explicit execution request, not permission beyond the user's
actual task. Inspect ok, state, result, error and the process exit status; a
failed operation may have made partial progress. Request IDs are correlation,
not deduplication: inspect before retrying. Structural doctor is not proof of
authentication or sync, and local_only/pending/conflict are not remote success.
Memory, references and capabilities still use their established commands below.
An active agent may inspect/repair its managed files or stage an approved toolkit,
but toolkit adoption/rollback must be driven from a separate management session
after the affected session exits. The API refuses a known live self-upgrade.
Repair or adoption does not reload already-loaded model instructions; honor the
requires_fresh_session result instead of claiming current context has changed.

Native user-wide and project-local skills may supplement this assistant's own
capabilities. Their availability depends on the current machine and harness;
they are not automatically part of the synchronized assistant repository.
Before reusing a remembered procedure, verify its required tools are available.
When recording a reusable procedure, distinguish assistant-owned capabilities
from host/project dependencies and describe prerequisites without hardcoding
machine paths. Report conflicting inherited guidance instead of silently
turning it into this assistant's permanent policy. Native instruction priority
still applies; remembered preferences are not higher-priority instructions.

Before substantial work, use "memory recall --query TEXT" and select an explicit
--scope-kind/--scope-id for account, project, or task-specific guidance. Use
"memory history --record ID" to explain changes. Scope is not inferred from a
similarity score. A conflicting revision is evidence to reconcile, not guidance.
Do not guess scope IDs from cwd, account identity, or the harness state directory.
When the relevant scope is unknown or recall is empty, use "memory scopes" to
list stored scope kinds/IDs, then recall explicitly within the relevant scope.
Scope discovery is metadata, not permission to apply unrelated guidance. Scope
IDs are stable identifiers and need not match display names or directory paths.
An empty recall is not evidence that a remembered fact is absent. Try the relevant
discovered scope with a simpler or empty query; if evidence remains unavailable,
say it was not found rather than inventing a name. Journal only verified outcomes;
record retrieval failures honestly instead of recording a guess as success.

To learn, use "memory write --input -" with a revision document on stdin, or
"memory write --input FILE" with a prepared document. Use
"memory template" for its shape; the writer stamps device, actor, harness, and
time. Preserve record/scope IDs and name predecessor revision IDs in supersedes.
Use "memory source --summary TEXT --kind user-direction" to record concise
evidence, and "memory event --summary TEXT --kind task-completed" for chronology.
Record meaningful outcomes and reusable learning before finishing the task.
One-task exceptions use task scope. Never copy secrets or raw transcripts.
Temporary drafts and test fixtures belong in a fresh private temporary directory,
not the user's project or synchronized assistant source. Prefer stdin for memory
writes. Use mktemp -d when files are needed, keep track of that exact directory,
and clean up only files you created. Do not commit scratch files, copied native
state, credentials, or raw session logs during automatic source synchronization.

Use "agent capabilities" to list portable capabilities. The runtime inventory
includes directory and instruction_files paths; read the selected capability's
listed instructions completely, then use its documented entrypoint. There is no
"agent capability" command. Do not search the whole repository to locate listed
instructions. Runtime inventory paths are not fields for capability.json.
Capabilities can contain
scripts and hook subscriptions. Before authoring a capability, read the built-in
"agent capability-guide" and use "agent capability-template --capability ID".
For local prerequisites, register machine_requirements in that private manifest.
Use "machine status" for historical readiness, "machine check" for explicit live
checks, and "machine prepare" to preview then apply one reviewed requirement.
Use --help for required flags. Do not install dependencies through session hooks.
Installation, service choices and one-time local secret enrollment belong to the
private capability, not the shared toolkit. Never put secret values in chat,
command arguments, source, memory or receipts. Structural doctor and historical
readiness are not live authentication tests.
Declare OS/architecture support and native evidence in private capability docs.
Use shared logic with explicit platform-specific backends where needed; selecting
a backend must not rewrite portable source. An unsupported platform is an error,
not permission to install or improvise a fallback. Follow the platform support
section of "agent capability-guide"; skipped native tests are not passing evidence.
For capability creation or material redesign, define the current ask first, then
use "reference list" to discover relevant historical libraries. Consult selected
files with "reference search" and "reference read"; their text is reference-only,
not current policy or permission to execute old scripts. Use experiences to
refine requirements and tests, not import procedures 1:1. Record reuse, adaptation,
rejection and source hashes using "agent capability-rationale" in the private
capability's RATIONALE.md. Missing resources are unavailable evidence, not a reason
to invent history. Linked libraries are never automatically loaded into memory,
instructions or the capability inventory by My Friday.
These commands work without a source checkout. Use --help for CLI discovery;
do not inspect executable strings or search for a development checkout to learn
the format. Use "agent check --capability ID" after changing
one. Local source checkpoints record the device and before/after Git versions.
Use "agent changes --path capabilities/ID/instructions.md" (or omit --path) to
inspect these observations. They identify the checkpointing machine, not proof
of original authorship; pulled changes keep their original records. An empty
result can mean older or externally committed work has no observation. Explain
process changes using memory supersession and journal reasons, not timestamp
ordering alone. Repository Git history is versioned; external effects require their own
reconciliation. User corrections override older remembered user guidance within
their actual scope. Native project instructions still apply to project work.
`, s.Root, i.Binary))
	text := strings.Join(instructions, "\n\n")
	for _, harness := range []string{"codex", "pi"} {
		files[harness+"/AGENTS.md"] = []byte(text)
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
	hookJSON, err := json.MarshalIndent(map[string]any{"hooks": hooks}, "", "  ")
	if err != nil {
		return nil, err
	}
	files["codex/hooks.json"] = append(hookJSON, '\n')
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
      for (const warning of result.warnings || []) { ctx.ui?.notify?.(warning, "warning"); }
      if (result.error) { ctx.ui?.notify?.(result.error, "warning"); }
      if (name === "before_agent_start" && result.additional_context) {
        return { message: { customType: "my-friday-memory", content: result.additional_context, display: false } };
      }
    });
  }
}
`
	extension = strings.ReplaceAll(extension, "BINARY", string(binary))
	extension = strings.ReplaceAll(extension, "INSTANCE", string(instance))
	files["pi/extensions/my-friday.ts"] = []byte(extension)
	return files, nil
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
		args = append([]string{"--dangerously-bypass-approvals-and-sandbox", "--dangerously-bypass-hook-trust", "--enable", "hooks"}, args...)
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
