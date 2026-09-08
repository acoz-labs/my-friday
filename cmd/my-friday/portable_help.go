package main

import (
	"fmt"
	"io"
)

var portableHelpTopics = map[string]string{
	"": `My Friday — portable assistant toolkit

Usage: my-friday <command> [options]
  setup       Create/import a private agent and machine-local instance
  agent       Inspect, validate, launch, and design private capabilities
  memory      Recall, revise, and explain persistent memory
  reference   Link and consult external reference-only libraries
  sync        Checkpoint and synchronize an explicit agent repository
  hook        Dispatch a harness lifecycle event (adapter entrypoint)
  help        Show help, optionally for a command

Start capability authoring with: my-friday agent capability-guide
Print a manifest with: my-friday agent capability-template --capability <id>
Running without arguments opens setup. Commands accept --help / -h.
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
The guide and template need no repository, installation, or source checkout.
Repair preserves credentials/sessions and the binding; optional --launcher PATH
creates a missing launcher only. It does not replace or relocate the executable.
Doctor compares generated files with the running toolkit and checks the selected
harness on PATH. It does not test authentication/network or change files.
Use my-friday help agent launch for launcher help; a launcher's --help is
forwarded to its selected harness. Use <command> --help for options.
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
  search    Discover files in --library ID with --query TEXT (empty lists files)
  read      Read --library ID --path FILE; use --sha256 HASH from search

Add/list use --repository PATH or --instance PATH (or their launch environment).
Bind/search/read require an instance; bindings do not travel in source Git.
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
