# Codex memory plugin — development candidate

Keep using native Codex, its skills, authentication and project settings. This
plugin attaches a separately selected My Friday memory bank. Tested native
discovery and MCP tool calls with Codex **0.153.4**; real model behavior remains
an acceptance gate, not a claim implied by those checks.

## Prepare a bank on this machine

Build the candidate using the repository's pinned toolchain:

```sh
mise exec -- go build -o /absolute/candidate/bin/my-friday ./cmd/my-friday
```

Make that directory available on the PATH used to start Codex. `my-friday version`
identifies the selected executable. An older assistant-platform executable does
not supply the new `bank`, `mcp`, or `codex-memory-hook` commands.

When more than one version is installed, also select the exact executable before
starting Codex:

```sh
export MY_FRIDAY_MEMORY_BIN=/absolute/candidate/bin/my-friday
```

Both MCP and hooks honor this machine-local override. It must be an absolute
executable path (spaces are supported). Without it they use `my-friday` on PATH.
Native login-shell setup can reorder PATH, so merely prepending the candidate
directory at launch is not enough to pin a trial. No global shell changes are
required; the variable can be supplied only to the Codex process.

Create a new memory bank; do not point this command at existing memory:

```sh
my-friday bank create --repository /absolute/private-memory \
  --name "Personal memory" --device-label "My laptop"
my-friday bank bind --repository /absolute/private-memory \
  --device-label "My laptop" --actor "My name"
my-friday bank doctor
```

The bank contains `bank.json`, `memory/{records,events,sources}` and
`provenance/{devices,changes}`. It needs no `agent.json`, capabilities, instruction
files or harness home. Procedures can be remembered as knowledge without
installing executable tools.

`bind` creates a new machine identity and a small **local** JSON binding. The
default is `my-friday/memory.json` under the operating system's user configuration
directory: `~/Library/Application Support` on macOS; `$XDG_CONFIG_HOME` or
`~/.config` on Linux. It is not Git-backed and contains no credential values.
Both commands preserve existing destinations rather than overwriting them.

To select a different bank, pass `--binding /absolute/local-binding.json` when
binding it, and launch Codex with `MY_FRIDAY_MEMORY_BINDING` set to that file.
Use separate bindings for personal/work banks; nothing silently combines them.
The same native Codex authentication and skills can serve either bank.

## Install the plugin

The marketplace root is this directory, `plugins/codex`, not the repository root.
The generated development marketplace is currently named `personal`. If your
Codex already has another marketplace with that name, stop and resolve that
collision; do not replace the existing marketplace. Public distribution naming
is not finalized.

```sh
codex plugin marketplace add /absolute/my-friday/plugins/codex
codex plugin add my-friday-memory@personal
```

These are native Codex installation commands. They add the selected plugin to
Codex; My Friday does not replace its configuration or copy authentication.
Open a **fresh thread**, use `/hooks` to review and trust the two bundled hooks,
and `/mcp` to confirm the memory server connected. Hook trust is a native setup
step; no bypass flag is required or recommended for everyday use.

The plugin's small POSIX shell runner starts the selected executable with
`mcp --harness codex`. It explicitly forwards `MY_FRIDAY_MEMORY_BIN`,
`MY_FRIDAY_MEMORY_BINDING`, `XDG_CONFIG_HOME`, and `SSH_AUTH_SOCK`, not arbitrary
credentials from the shell. Default binding selection needs no export. The
runner requires `/bin/sh` on the current macOS/Linux targets.

## What happens during work

- `SessionStart` supplies a short memory-use orientation.
- `UserPromptSubmit` reads bounded bank-wide evidence matching the prompt and
  a small scope inventory. It neither journals, synchronizes, nor reads the
  native transcript. Scoped recall and deeper retrieval use MCP tools.
- The agent uses its existing model to decide what is worth learning. It writes
  explicit facts/decisions and concise semantic journals through memory tools,
  incrementally and before finishing when allowed.
- `memory_sync` checkpoints and exchanges changes with an already configured
  Git remote. Offline/pending means local durability, not confirmed delivery.

There is no hidden background summarizer, second inference service, or end-hook
transcript capture. Abruptly interrupted **unsaved** work can be lost. Native
compaction/resume and model adherence still require real conversation testing.
Explicit read-only/no-save requests prohibit writes and sync. Hooks read the
local clone, so automatic recall alone does not establish remote freshness.

Recall uses scoped lexical ranking, not vectors. Compact recall defaults to
8 KiB and supports a 1–32 KiB budget; hooks use 6 KiB for their recall packet.
The complete hook also includes short orientation and routing metadata. Skills
load progressively; the memory bank is never loaded wholesale. History, scope
and journal responses are paged/bounded. Truncation and conflicts are explicit.
Current user direction outranks conflicting historical guidance within its scope.

Writes are append-only, not idempotent. If a result was interrupted or lost,
inspect current memory/recent journals before retrying. A rejected proposed
revision writes neither its evidence nor its record. Disk failure between
evidence and revision writes can leave unused evidence, not a dangling revision.

## Another machine and Git synchronization

Use a private Git remote you control. Configure its existing `origin` through
normal Git tooling, then run `my-friday bank sync`. Sync uses branch `main`,
never force-pushes, and preserves competing semantic revisions as conflicts.
Transport credentials remain machine-local; configuring a remote is not handled
by the MCP tools. The current engine ignores global Git configuration and its
credential helpers: do not assume ambient HTTPS authentication will work.
An SSH remote can use normal SSH configuration and an available agent socket.
Test source access from the actual Codex process, particularly over SSH.

On the next machine, clone the bank and bind the clone:

```sh
git clone <your-private-memory-remote> /absolute/private-memory
my-friday bank bind --repository /absolute/private-memory \
  --device-label "Second machine" --actor "My name"
my-friday bank sync
```

Install the same plugin there. Native threads and credentials are not portable
memory and are not copied. Existing assistant-format repositories are not
silently adopted; migration is a separate explicit task.

## Update and diagnose

Use `my-friday bank doctor` for read-only structure and binding checks. It does
not prove remote freshness, authentication, plugin loading or model behavior.
Missing binding/server: check the selected executable, PATH, binding path and
`/mcp`. Missing automatic recall: inspect `/hooks` for disabled/untrusted hooks.
`unknown command "codex-memory-hook"` indicates an older executable was selected.
Set `MY_FRIDAY_MEMORY_BIN` to the intended candidate and start a fresh session.
The packaged runner converts a missing/failed hook executable into a visible
nonblocking warning; it suppresses failed raw output. This does not make memory
available: diagnose the warning before relying on automatic recall. MCP startup
failures remain visible as an unavailable server, not a successful connection.

After updating the candidate binary and packaged plugin, reinstall the plugin
through its configured local marketplace and start a fresh thread. Development
plugin changes use a version cachebuster to avoid stale native caches. No live
Alfred runtime or previously installed assistant is upgraded automatically.

Repository layout:

```text
plugins/codex/
  .agents/plugins/marketplace.json
  plugins/my-friday-memory/
    .codex-plugin/plugin.json
    .mcp.json
    hooks/hooks.json
    scripts/run-memory.sh
    skills/memory/SKILL.md
```

Native lifecycle adaptation lives in `internal/memorycodex`; it delegates to
the same `internal/memorybank` service as MCP and CLI.
