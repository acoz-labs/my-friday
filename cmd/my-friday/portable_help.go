package main

import (
	"fmt"
	"io"
)

var portableHelpTopics = map[string]string{
	"bank": `Usage: my-friday bank <command> [options]
Memory-only banks do not need an assistant installation or launcher.
  create          Create --repository PATH --name NAME --device-label LABEL
  bind            Enroll --repository PATH --device-label LABEL --actor NAME
  doctor          Read-only structure and local binding checks
  recall          Retrieve compact current evidence with --query TEXT
  remember        Append knowledge/corrections from bounded JSON on stdin
  history         Inspect --record ID, optionally --offset N --limit N
  scopes          Discover stored scopes, optionally --offset N --limit N
  journal         Search semantic journals with --query TEXT
  journal-append  Append {"kind":"session","summary":"..."} from stdin
  sync            Checkpoint locally and sync the bank's configured Git remote

bind and all memory operations accept --binding FILE. Default: the operating
system's user config directory / my-friday / memory.json, or
MY_FRIDAY_MEMORY_BINDING. The binding stays outside the Git bank and pins its
identity and this machine's device ID. Use a separate binding for another bank.
Clone an existing bank with Git, then bind its local path on the new machine.
Nothing targets the working directory implicitly. See each command's --help.
`,
	"menu": `Usage: my-friday [menu [--plain] [--legacy]]
Open memory-bank management: create a bank, connect an existing local clone,
check its structure, or synchronize its configured Git remote. No assistant
launcher, capability setup or native login is required to manage a bank.
Use --legacy for previous assistant-platform setup, doctor/repair and updates.
Interactive terminals use arrows or j/k, Enter/l,
Esc/h and gg/G. Text fields use normal typing, Tab to edit the default, and
Ctrl+U to clear. Ctrl+C exits. The UI restores terminal input modes on exit.
Use --plain or MY_FRIDAY_PLAIN=1 for numbered prompts; non-TTY/dumb terminals
automatically use plain mode. There, 0 is Back/Exit and :back cancels a prompt.
Opening the menu does not change files or launch an agent.
`,
	"api": `Usage: my-friday api describe
       my-friday api --input FILE (or - for stdin)
One bounded JSON request/response. Discover versioned actions and parameter
schemas with api describe; never automate the human TUI.
Request: {"schema_version":1,"action":"agent.doctor","params":{"instance":"/absolute/instance"}}
Mutations default to preview; apply=true explicitly executes within the user's
actual authority. A preview is not full preflight or a stored approval.
Inspect ok, state, error, result and process exit status. Partial work may survive
failure; request IDs correlate, not deduplicate. Do not blindly retry mutations.
Known active self-upgrade is refused; use a separate management session after
the affected agent exits. Memory/reference/capability CLI commands remain available.
`,
	"version": `Usage: my-friday version (or --version)
Print build revision, platform and portable-management compatibility as JSON.
`,
	"toolkit": `Usage: my-friday toolkit [command] [options]
No command opens the update menu: latest release, approved local artifact,
or retained toolkit. Updates preserve per-agent version pins.
  check-instance --instance PATH    Read-only source compatibility check
  use --instance PATH [--launcher PATH]
                                    Adopt the running toolkit with backups
  manifest --binary PATH --release TAG
                                    Print approved release metadata; no publishing
Close active agent sessions before adoption. An omitted launcher stays unchanged.
Use menu --legacy for guided assistant adoption, repair and rollback.
`,
	"": `My Friday — durable memory for your agent

Usage: my-friday <command> [options]
  bank        Create, bind, and use a memory-only bank (memory-service MVP)
  mcp         Serve the selected memory bank over stdio MCP
  menu        Open memory-bank management; --legacy manages previous assistants
  api         Discover and execute structured, noninteractive management actions
  toolkit     Open updates; use/check-instance/manifest support explicit tooling
  version     Print toolkit build and portable compatibility metadata
  setup       Create/import an agent, or resume remote setup with --instance PATH
  agent       Inspect, validate, launch, and design private capabilities
  memory      Recall, revise, and explain persistent memory
  reference   Link and consult external reference-only libraries
  machine     Inspect readiness and explicitly prepare private prerequisites
  sync        Checkpoint and synchronize an explicit agent repository
  hook        Dispatch a harness lifecycle event (adapter entrypoint)
  help        Show help, optionally for a command

Running without arguments opens memory-bank management. Commands accept --help / -h.
Assistant/capability commands above are retained compatibility, not memory setup.
Resume source hosting/account setup with: my-friday setup --instance PATH
Legacy init, assistant, capability, codex, validate, and recover commands remain
separate from this portable workflow; use agent commands for portable capabilities.
`,
	"agent": `Usage: my-friday agent <command> [options]
  launch                Start the configured harness, preserving project cwd
  doctor                Read-only installation checks for --instance PATH
  repair                Rebuild generated files for --instance PATH; no network
  inspect               Print the agent identity
  validate              Validate agent source and memory structure
  changes               List source checkpoint provenance; optional --path FILE
  capabilities          List private capabilities with directory/instruction paths
  capability-guide      Print the built-in capability design/format guide
  capability-template   Print a manifest; requires --capability <id>
  capability-rationale  Print a private design/evidence rationale template
  check                 Validate and execute checks; requires --capability <id>

Most commands accept --repository PATH (or MY_FRIDAY_ASSISTANT_ROOT).
Check also accepts --instance PATH (or MY_FRIDAY_INSTANCE), validates the binding
and supplies instance/device context to its temporary capability copy. An explicit
instance overrides ambient source defaults; an explicit repository must match.
Use --instance '' --repository PATH for intentional source-only checks.
The guide and template need no repository, installation, or source checkout.
Repair preserves credentials/sessions and the binding; optional --launcher PATH
creates a missing launcher only. It does not replace or relocate the executable.
Doctor compares generated files with the running toolkit and checks the selected
harness on PATH. It does not test authentication/network or change files.
Use my-friday help agent launch for launcher help; a launcher's --help is
forwarded to its selected harness. Use <command> --help for options.
`,
	"machine": `Usage: my-friday machine <command> --instance PATH [options]
  status    List requirements and historical local readiness; no scripts or writes
  check     Run selected private check/verify and save local receipts
  prepare   Preview selected requirement; execute with --apply --expect-sha256 HASH

check/prepare require --capability ID --requirement ID. Instance may come from
MY_FRIDAY_INSTANCE; working directory is never an implicit target. Commands
return JSON; never automate the TUI. Apply requires the preview fingerprint.
Only exit 10 from check requests preparation; other failures do not install.
Successful prepare is verified, and already-satisfied installers are skipped.
Commands are noninteractive; private secret enrollment is a separate local step.
No credential values belong in arguments, output, source, memory or receipts.
Scripts have full user access, not a sandbox. Stdout/stderr are discarded.
Status exit 0 means inspection completed, not all requirements are ready.
check/prepare exit nonzero when not ready; inspect JSON state and phases before
retrying. Cancellation does not roll back effects. No implicit sync or updates.
See agent capability-guide for the manifest and script contract.
`,
	"memory": `Usage: my-friday memory <command> [options]
  template   Print a revision document to edit
  write      Save a revision with --input FILE (or - for stdin)
  recall     Retrieve current scoped guidance with --query TEXT
  scopes     List stored scope IDs and record counts (not guidance)
  history    Explain a record's revisions with --record ID
  source     Save concise evidence with --summary TEXT
  event      Save a journal entry with --summary TEXT

Use --repository PATH / MY_FRIDAY_ASSISTANT_ROOT and --device ID /
MY_FRIDAY_DEVICE_ID for writes. Select --scope-kind and --scope-id for scoped
recall. Discover IDs with memory scopes; do not infer them from cwd. Scope counts
include stored history/future records, not just currently effective guidance.
Corrections append a new revision with explicit supersedes IDs.
Use <command> --help for options.
`,
	"reference": `Usage: my-friday reference <command> [options]
  add       Register a portable descriptor: --library ID --title TEXT
            --description TEXT --purpose TEXT; then checkpoint source
  list      List descriptions, not local availability or current guidance
  bind      Bind/rebind --library ID to an external local directory with --path
  status    Check --library ID on this instance without reading any documents
  search    Discover files in --library ID with --query TEXT (empty lists files)
  read      Read --library ID --path FILE; use --sha256 HASH from search

Add/list use --repository PATH or --instance PATH (or their launch environment).
Bind/status/search/read require an instance; bindings do not travel in source Git.
Status returns available/unbound/stale/invalid/unavailable in JSON; exit 0 means
the check completed, not that the directory or its documents are usable.
Existing local directories and Git working trees are read as text. Nothing is
cloned, fetched, executed, or automatically promoted into memory/capabilities.
Search results contain paths and hashes, not source text. A changed descriptor
requires review and an explicit rebind. Read requires canonical relative paths;
hidden paths, node_modules, symlinks, binary and oversized files are excluded.
Use agent capability-guide for reference-aware building, and
agent capability-rationale to document reuse/adaptation/rejection with evidence.
Use <command> --help for options.
`,
	"agent launch": `Usage: my-friday agent launch --instance PATH [--harness codex|pi] [harness arguments]

Uses MY_FRIDAY_INSTANCE when --instance is omitted. Preserves caller cwd.
Only --instance and --harness are reserved; other arguments pass to the harness.
A literal -- ends launcher parsing and is also forwarded. --help after launch
belongs to the harness. This command synchronizes source and regenerates the
instance projection before launch. Harness execution has full user access.
`,
}

func printPortableHelp(topic string, out io.Writer) error {
	text, ok := portableHelpTopics[topic]
	if !ok {
		return fmt.Errorf("unknown help topic %q; use my-friday --help", topic)
	}
	_, err := io.WriteString(out, text)
	return err
}

func helpFlag(arg string) bool { return arg == "--help" || arg == "-h" }
