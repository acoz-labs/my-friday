# Portable assistant foundation

This is the locally implemented first slice of the portable rebuild. It does
not install a live personal agent, migrate legacy data, or publish a release. The old
two-repository commands remain available independently.

## Public core versus private agent

My Friday owns shared infrastructure: memory management, capability design and
validation, repository format, setup, synchronization, and harness/lifecycle
adapters. It does not bundle service integrations, password-manager clients,
account-role policies, or personal workflows—even as optional built-in packages.
Those are authored in each user's private agent repository using the shared
capability contract. A fresh agent has no service capabilities or accounts.

The toolkit may define an extension point needed by core infrastructure, such as
the Git credential-helper contract below. The implementation behind that
extension point belongs to the user. Core tests use provider-neutral synthetic
fixtures, and private capability acceptance belongs with the private capability.

## Build and create

Requirements: Go 1.26.4 (pinned by mise), Git with `merge-tree --write-tree`
(Git 2.38+), and a separately installed harness. macOS/ARM64 is the exercised
host; the portable package also builds for Linux. Windows is not supported.

```sh
mise install
mise exec -- go build -o ./my-friday ./cmd/my-friday
./my-friday setup
```

The wizard selects new versus existing **local** repository, a launcher name,
default harness for a new agent, and a readable machine label. Names use
3–128 lowercase letters, digits, or hyphens, beginning with a letter. It creates
a device ID rather than relying on a hostname as identity. Explicit flags allow
noninteractive disposable setup:

```sh
./my-friday setup --name friday --harness codex \
  --repository /absolute/new/agent --state /absolute/new/instance \
  --device-label 'Test laptop' --no-launcher
```

Without `--no-launcher`, setup creates `~/.local/bin/<name>` or the explicit
`--launcher` path. It refuses existing instance/launcher destinations. Put the
binary somewhere stable before installing a launcher: the launcher binds that
absolute binary path. Setup initializes and commits the local source repository.
It does not create a remote repository, install a harness, copy credentials, or
log in. A filesystem error midway through setup can leave a partial installation;
inspect the reported paths before retrying. Existing installations are not removed.

Import currently requires an already-cloned **new-format** repository:

```sh
./my-friday setup --import /absolute/cloned/agent --name friday \
  --state /absolute/new/instance --device-label 'Second computer' --no-launcher
```

The imported agent retains its default harness. Each instance has separate Codex
and Pi homes; authenticate each harness in its own home using its native login
flow. They do not inherit the owner's existing harness login. Authentication and
sessions stay machine-local; only assistant source is synchronized.

```sh
cd /path/to/software/project
friday --harness pi --model YOUR_MODEL 'Work on this project'
```

The caller's cwd is preserved. `--instance` and `--harness` are reserved launcher
arguments; other arguments are forwarded literally. A literal `--` ends launcher
parsing and is also forwarded. Both harnesses receive explicit absolute assistant
paths and generated instructions. Codex launches with approval/sandbox and hook
trust bypass flags. This is deliberate full-access execution, **not OS isolation**.
Project instructions still participate in the harness's instruction hierarchy.

## Source and local state

```text
agent/
  agent.json                    # Format version, stable identity, default harness
  .my-friday/sync.json           # Optional generic Git author/helper configuration
  instructions/                 # Identity and operating instructions
  memory/records/<record-id>/    # Immutable JSON claim revisions
  memory/sources/                # Concise evidence with originating device
  memory/events/YYYY/MM/         # Append-only journal entries
  provenance/devices/           # Stable device IDs and human-readable labels
  provenance/changes/           # Reserved for richer source-change provenance
  capabilities/<capability-id>/ # Manifest, instructions, scripts, checks
  integrations/                 # Nonsecret configuration and secret references
  extensions/                   # Reserved user extension space
  .my-friday/local/              # Git-ignored hook execution receipts
instance/
  binding.json                  # Absolute source/binary paths and device binding
  codex/                        # Generated instructions/config/hooks; native auth/state
  pi/                           # Generated instructions/extension; native auth/state
```

Source format 1 is explicit. Unknown versions fail instead of being guessed at.
There is no version-upgrade engine yet. Generated instance instructions, hook
registrations, and Codex config are regenerated at launch; do not customize those
generated files. Native authentication files are not rewritten.

## Memory and supersession

Inside a launched session, the environment supplies repository and device paths.
Outside it, provide `--repository` and, for writes, `--device`. Commands emit JSON:

```sh
my-friday memory template
my-friday memory source --summary 'User established an ongoing preference'
my-friday memory write --input /path/to/revision.json
my-friday memory recall --query 'response format'
my-friday memory scopes
my-friday memory history --record record-example
my-friday memory event --kind task-completed --summary 'Completed the requested work'
```

The [runtime schema](../internal/portable/schemas/memory-revision.schema.json)
defines claims. The writer stamps machine, actor, harness, session when available,
and recorded time. Evidence references point to concise source records, not raw
transcripts. Journal entries also carry authorship. Do not put credential values
in source, evidence, journals, or hook output.

A correction uses the same `record_id`, kind, and scope, a new revision ID, and
predecessor IDs in `supersedes`. `change_reason` explains why. A one-task exception
belongs to task scope, not a global replacement. Cycles, missing predecessors,
scope changes inside a chain, and backdated predecessor ordering are rejected.
Sync rejects modification/deletion of already-committed memory or provenance
files. Append a correction instead. This is not tamper-proof against an owner
rewriting the Git history outside My Friday.

Two concurrent successor revisions remain a conflict. Neither becomes current
guidance until an explicit revision supersedes both. Conflict records in the
selected scope surface even when their text does not match a query. Future-effective
revisions do not replace current guidance early.

Recall directly scans and validates source on each read. It uses deterministic
weighted lexical matching, not vectors, BM25, or QMD. An explicit scope selects
account/project/task guidance plus assistant-global guidance. Prompt hooks recall
global guidance; the model must request other scopes explicitly. Search does not
infer permissions or select an account. There is no index that can lag source,
but remote changes are visible only after successful synchronization. Updating a
file also does not erase an earlier statement already in model context.

When the relevant scope is unknown, `memory scopes` lists stored scope kinds,
stable IDs, and distinct record counts, sorted by kind/ID. It validates current
source on every call and includes history/future records in its counts. This is
routing metadata, not a cross-scope guidance packet. Select a relevant ID and
call recall explicitly; IDs need not match a project display name or cwd.
An empty recall is not proof that a fact is absent. Generated instructions and
recall notices direct agents to discover scopes and retry rather than guess.

## Synchronization

`my-friday sync` commits local changes and reconciles `origin/main`. Launch and
session-start/request-received/request-completed hooks also call it; memory writes
checkpoint immediately. Offline failures return `pending` with local commits intact.
No remote returns `local-only`. Git text conflicts preserve the local worktree and
both histories. Candidate trees are validated before adoption; semantic memory
conflicts survive a successful Git merge. Pushes are non-forced. Concurrent local
writers receive a retry error rather than overwrite one another.

The assistant must own a normal `.git` directory on `main`; nested repositories,
worktree bindings, and in-progress Git operations are refused. Sync checkpoints
all nonignored assistant source, so never place unrelated work or raw secrets in
this repository. Git history can recover source but cannot undo external effects.

Local remotes work without credentials. Network sync currently supports
HTTPS remotes with an explicitly configured private Git credential helper. It
does not choose a hosting service, password manager, or account role. SSH and automatic
remote cloning/creation are not implemented. Status is returned per operation;
there is no persistent sync dashboard or background retry daemon. Hook deadlines
can interrupt long synchronization; interrupted operations need a later retry.

## Capabilities and lifecycle subscriptions

The executable carries its own authoring reference and manifest template:

```sh
my-friday --help
my-friday help agent
my-friday agent capability-guide
my-friday agent capability-template --capability example-capability
```

Guide/template commands need no installation, source checkout, or network.
The template prints JSON only; implement or replace its placeholder executable
check before running it. The embedded guide describes every manifest field,
the authoring workflow, portable paths, temporary check execution, and optional
subscriptions. Generated harness instructions point agents to these commands.
Launched sessions and capability handlers expose `MY_FRIDAY_BIN`; use it and
`MY_FRIDAY_ASSISTANT_ROOT` rather than copying installation paths into reusable
instructions. Existing instance instructions refresh on the next launcher run.
Top-level and command help exit successfully. `my-friday help agent launch`
documents the launcher; `--help` forwarded through a launcher still belongs to
the selected harness.

Each capability directory contains `capability.json`, and can include instructions,
scripts, and checks. For example:

```json
{
  "schema_version": 1,
  "id": "daily-context",
  "description": "Supply useful local context before a request",
  "subscriptions": [{
    "id": "add-context",
    "event": "request.received",
    "command": ["./context.sh"],
    "timeout_seconds": 2,
    "failure": "warn"
  }],
  "checks": [["./check.sh"]]
}
```

`agent capabilities` lists manifest fields plus runtime-only `directory` and
`instruction_files` paths. The latter lists existing regular `instructions.md`
and `README.md` files, in that order; an empty list means neither was found.
These paths are computed for the current installation and must not be copied
into source manifests. Agents can read the listed files directly without a
repository-wide search. `agent check --capability daily-context`
validates ordering and executes declared checks. There is no separate activation
approval: a valid committed capability can execute at its subscribed event.
Commands run from a temporary copy of their capability source, receive a structured
event on stdin, and can return `{"additional_context":"..."}` on stdout. Use explicit
source paths for durable edits; edits inside the temporary copy are discarded.
Hooks have full user access and may execute external effects. Never subscribe an
irreversible action that assumes exactly-once delivery.

`after` orders handlers by local subscription ID or `capability/subscription`.
Timeouts terminate the process group. `failure: stop` stops the subscription chain;
it does **not** promise to block the harness's action. Failed/interrupted receipts
require inspection before replay. Deduplication is local and applies only when
the same event ID is supplied. Native events without stable IDs cannot be
deduplicated across redelivery. Receipts retain status, not raw event payloads or
returned context. The trusted consumer remains responsible for its own logs.

The Codex projection registers the twelve native events listed in `launch.go`.
The Pi projection currently subscribes to thirteen session/agent/turn/tool/model
events. Events without a common semantic mapping stay native-namespaced; Pi
`turn_end` is not treated as task completion. Pi injects returned context at
`before_agent_start`; other callbacks currently only report errors to the UI.
Background handlers, every Pi extension event, source-change events, persistent
outboxes, and guaranteed end-of-task learning capture are not implemented.

## Provider-neutral Git authentication

An optional `.my-friday/sync.json` selects a **user-supplied** executable and Git
commit attribution. No helper implementation is distributed or seeded:

```json
{
  "schema_version": 1,
  "credential_helper": ["./capabilities/source-auth/scripts/git-credentials"],
  "author": {
    "name": "Example Agent",
    "email": "agent@example.invalid"
  }
}
```

The array is literal argv, not shell source. Relative executable paths containing
a slash resolve against the agent repository; bare command names use PATH.
The executable uses Git's standard credential-helper protocol: an appended
`get`, `store`, or `erase` argument, a credential description on stdin, and
protocol output on stdout. HTTPS request descriptions include the repository
path. The private helper must validate its intended host/repository, resolve its
chosen identity, and handle storage/erasure appropriately. Provider-specific
identity verification is not a core promise. Git attribution is metadata, not
proof of authentication. Never embed credential values in argv or configuration.

Synchronization ignores global/system Git configuration, resets credential
helpers, disables interactive prompting, and uses only the explicitly selected
helper. Other process environment variables are inherited; this is not credential
isolation or a scrubbed environment. The private helper owns its bootstrap and
must not fall back to an unintended account. Core launchers reserve harness and
My Friday variables, but do not interpret service-specific environment variables.

Without a helper, HTTPS sync remains visibly pending after committing locally.
Without an author override, commits use the agent's name and
`my-friday@localhost`. Configuration is snapshotted per sync operation. Imported
helpers may need machine-local dependencies or bootstrap before network sync can
work. No core CLI commands read/create provider secrets or execute named account
roles; those operations belong to private capabilities.

## Verification and remaining milestones

Tests use disposable repositories, fake harnesses, and synthetic credentials.
They exercise supersession/conflicts, cross-clone synchronization, offline commits,
history rewrite refusal, hook ordering/timeouts/deduplication, native prompt
injection, literal launch arguments/cwd, optional generic attribution, and real Git
credential-helper plumbing against a fake credential provider.

```sh
mise exec -- go test -race ./internal/portable ./cmd/my-friday
mise exec -- bin/ci
```

Actual authenticated Codex/Pi conversations and live network-provider effects have
not been validated. Installed Codex 0.153.4's CLI flags were inspected; Pi is not
installed on the development host. This is not yet a production-ready runtime.

Local verification on 2026-09-06: `mise exec -- bin/ci` passed, including native
acceptance primitives, `go vet`, race-enabled tests, and Darwin builds. An
additional Linux/AMD64 CLI cross-build passed. The initial CI run exposed the
legacy subprocess allowlist (updated narrowly for portable runtime files) and a
one-off killed legacy launcher helper (passed standalone and on the full rerun).
That verification used disposable fixtures; it did not install a live agent,
import private memory, bootstrap credentials, mutate services, or publish a release.

Next milestones are live harness parity, richer configuration/capability machine
provenance (currently only memory/source/journal authorship is stamped), source
migrations, remote bootstrap, QMD retrieval evaluation, lifecycle completeness,
and staged legacy import. User service integrations are outside the public-core
roadmap. Legacy operational instructions and scripts must be reconciled
before activation, not imported wholesale as current authority.

See [the hands-on pilot record](testing/portable-pilot.md) for user-experience
checks, discovered defects, and the next retest. Automated passing checks alone
do not close a hands-on finding.
