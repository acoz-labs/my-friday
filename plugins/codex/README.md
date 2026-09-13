# Codex memory plugin — development candidate

Keep using native Codex, its skills, authentication and project settings. This
plugin attaches a separately selected My Friday memory bank. Conversation tests
passed on macOS/ARM64 with Codex **0.153.4** and **0.154.0**, including physical-host
continuity. See the [acceptance ledger](../../docs/memory-service-mvp.md) for exact
artifacts, cases and limitations. Linux builds are not Linux runtime acceptance.

## Prepare a bank on this machine

Build the candidate using the repository's pinned toolchain:

```sh
mise exec -- go build -o /absolute/candidate/bin/my-friday ./cmd/my-friday
```

Run that exact executable to open the management menu. Create or connect a bank,
then choose **Connect or update Codex memory**. The connection step embeds the
selected runtime and binding paths in a machine-local plugin copy. Launch Codex
normally afterward; no memory environment exports, agent alias or special working
directory are required for that profile. `my-friday version` identifies the
management executable; its `memory_protocol` must be 1 for this plugin connection.
Examples below assume the intended executable is available as `my-friday`.

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
binding it, then select that file in the native connection step. This changes the
default for fresh sessions in that native profile. For per-session selection,
launch Codex with `MY_FRIDAY_MEMORY_BINDING` set to the alternate file instead.
Use separate bindings for personal/work banks; nothing silently combines them.
The same native Codex authentication and skills can serve either bank.

## Connect, update and repair from the menu or CLI

The executable carries the public plugin package; no source checkout is needed
after building it. Connection preview is read-only:

```sh
my-friday bank connect-codex --binding /absolute/local-binding.json
# Review the JSON preview, then apply the same options:
my-friday bank connect-codex --binding /absolute/local-binding.json --apply
my-friday bank doctor-codex
```

Optional `--codex /absolute/codex`, `--codex-home /absolute/native-profile`,
`--binary /absolute/trusted/my-friday`, and `--installation-home /absolute/home`
make all machine targets explicit for automation. The defaults use the current
native profile (`CODEX_HOME` or `~/.codex`) and running My Friday binary.
`--apply` runs trusted local executables; a checksum is not publisher trust.

The installer retains a checksum-addressed runtime and connection-specific
plugin bundle under `.local/share/my-friday` in the installation home. It uses
native Codex marketplace/plugin commands, never rewrites native configuration
itself or copies login data. Existing marketplace names owned by another source
are refused. Only one My Friday memory plugin should be enabled per profile;
use another profile or an explicit binding override for a concurrent other bank.

For an update, select the new trusted executable and reconnect. For missing or
changed generated files, choose **Repair Codex memory connection**, or rerun
`connect-codex` with `--refresh --apply` and the intended binding. Repair uses the
running toolkit by default and creates a new generated copy; it retains previous
copies and edits. It does not repair or migrate memory, enroll a new device, or
reset authentication. A missing/unverifiable ownership receipt needs manual
inspection rather than automatic takeover. Close affected sessions first, then
start a fresh session and review hook trust/MCP status.

Native Codex requires removing the old local marketplace registration before
registering a different root with the same name. Tested local removal preserves
source/cache files. An interrupted/failed switch may leave registration incomplete;
the JSON outcome exposes the retained previous root. Reconnect with the same
bank/runtime after inspecting the failure. No automatic rollback is claimed.

**Update My Friday** (also `my-friday toolkit`) stages an approved local artifact
or eligible official release and changes only the management command. The memory
menu requires `memory_protocol: 1` in release metadata and the executable, rejecting
assistant-only releases. Existing memory connections and Alfred stay pinned.
Reopen the menu and reconnect a selected native profile to adopt the new runtime.
Retained binaries can be selected explicitly for recovery; no memory history or
external work is rolled back. No compatible public release is implied by this
development implementation.

`doctor-codex` checks managed bundle and native cache bytes, runtime checksum,
binding/bank identity and native plugin inventory. It reports conflicting memory
environment overrides in the checking process. It does not prove native login,
hook trust, live MCP startup, active-session context, remote freshness or model
behavior. It compares generated content with the running toolkit; a newer toolkit
can report an older connection stale. `bank doctor` remains the bank-only check.

## Manual source-plugin installation (development alternative)

The marketplace root is this directory, `plugins/codex`, not the repository root.
The generated development marketplace is currently named `personal`. If your
Codex already has another marketplace with that name, stop and resolve that
collision; do not replace the existing marketplace. Public distribution naming
is not finalized.

This source-package path is unpinned: select `MY_FRIDAY_MEMORY_BIN` as an absolute
executable path when multiple runtimes exist, and use `MY_FRIDAY_MEMORY_BINDING`
for a non-default bank. Native login shells can reorder PATH. Guided installation
above supplies machine defaults instead; explicit environment overrides still win.

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
compaction and interruption have conversation evidence for the documented cases,
not a guarantee that unsaved work survives or model adherence is infallible.
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
Both `ssh://user@host/path/to/bank.git` and `user@host:path/to/bank.git`
are supported. SSH needs no HTTPS credential helper. Host keys and client keys
must already be configured on that machine; My Friday does not enroll either or
bypass host verification. For unattended runs, configure OpenSSH with batch mode
and strict host-key checking. A bank-local Git `core.sshCommand` can select a
specific SSH configuration without changing global settings. `GIT_*` environment
overrides (including `GIT_SSH_COMMAND`) are intentionally not inherited.

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

For guided connections, use the connect/update/repair flow above. For manual
source installs, after updating the binary and packaged plugin, reinstall through
its configured local marketplace and start a fresh thread. Development
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

Native installation/session behavior is described in the official
[Codex plugins documentation](https://learn.chatgpt.com/docs/plugins). My Friday
also tests the installed CLI contract directly; documented availability alone is
not treated as acceptance of a particular machine/version.
