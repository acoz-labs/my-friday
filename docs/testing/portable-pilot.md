# Portable assistant pilot

Development testing, not release acceptance. Use disposable agent repositories
and synthetic prompts; do not publish session transcripts, local installation
paths, credentials, or private capability source.

## Initial Codex trial — 2026-09-07

- Setup and project-directory launch: user reported correct identity, cwd, and
  explicit memory location.
- Fresh-session recall: a fictional project name survived restarting the harness.
- Correction/history: a renamed project was current and its old name historical.
  Read-only inspection confirmed two immutable revisions, explicit supersession,
  originating device metadata, and successful source validation.
- Working-directory independence: a fresh session from a different directory
  recalled the same remembered entity while reporting its actual new cwd.
- Private capability execution: an agent created instructions, a manifest,
  executable code, and checks, with no automatic subscriptions. Recorded tool
  results showed validation and execution success.

These are scoped observations from one trial, not proof of every harness,
machine, scope, or failure mode. User-visible results and recorded tool traces
are complementary evidence; no claim is made about inaccessible model reasoning.

## Finding: authoring discovery depended on the development checkout

Despite a successful final result, the agent could not obtain the capability
contract from the installed toolkit. Top-level --help returned an error. It
inspected executable strings and ultimately read Go source in a development
checkout to discover the manifest/check format. Reusable instructions also
included installation-specific paths.

Correction: ship successful CLI help, an embedded capability authoring guide,
a JSON manifest template, and MY_FRIDAY_BIN for runtime path resolution.
Point fresh harness sessions to that guide. Tests cover discovery without a
repository binding, template shape/invalid IDs, and projected instructions.

Fresh-session authoring retest passed on candidate `bc407a1` (2026-09-07).
The agent created a different small private capability using the installed
guide and manifest template, without inspecting executable strings or the
development checkout. Its reusable instructions resolve paths through the
runtime environment rather than hardcoded installation paths. It registered
no automatic subscriptions.

Recorded calls showed executable checks covering 13 counting cases, two error
cases, and local no-write assertions. An initial Unicode test expectation was
incorrect; the agent ran the check directly for diagnostics and corrected the
expected count, preserving the assertion. Independent reruns of the checks,
toolkit capability check, and source validation passed. The private source
worktree remained clean after verification.

This closes the observed authoring-discovery finding for this Codex scenario,
not every harness or capability.

## Fresh-session capability reuse — 2026-09-07

A fresh Codex session received a synthetic counting request without the
capability name or implementation path. Its recorded calls listed capability
manifests, read the selected instructions, and executed the existing script
with exact stdin bytes. The script returned the expected counts. No capability
was created or modified: inspection confirmed no capability diff against the
pre-session checkpoint and a clean private source worktree.

Reuse is demonstrated for this scenario. Discovery still involved an extra
agent-help call and a filename search between listing manifests and reading
instructions. Consider making instruction locations explicit in discovery
output; the correct result alone does not prove optimal tool efficiency.

## Pi preflight — 2026-09-07

Installed `@earendil-works/pi-coding-agent@0.85.1` in a disposable local npm
prefix with dependency lifecycle scripts disabled, following the
[upstream package instructions](https://github.com/earendil-works/pi/blob/main/packages/coding-agent/README.md).
No global harness installation or shell configuration was changed.
The existing agent launcher successfully forwarded Pi version/help requests.
Pi's installed extension loader loaded the generated My Friday extension with
all thirteen registered event names and no errors; its context-file loader
found the generated instance instructions from an unrelated project cwd.

This preflight verifies loading, not callback delivery or authenticated model
behavior; the following hands-on run exercises more of the path.

## Finding: Pi recall guessed unknown scopes — 2026-09-07

The user authenticated Pi and ran the same agent from the separate project
directory. Its transcript contains the injected My Friday memory packet,
demonstrating prompt-hook context delivery. It discovered and executed the
existing private capability with correct results and no capability edits.

Memory recall failed: the agent guessed scope IDs from filesystem/account
context, received empty packets, and reported the working directory's name as
the remembered fictional project's name. It also journaled the answer as a
successful retrieval. The stored current revision was unchanged and correct;
the failure was finding and using it. A separate capability-discovery detour
included an invented command and a repository-wide filename scan.

Correction: add `memory scopes` to expose validated routing metadata without
widening scoped guidance. Recall notices and generated instructions direct
agents to discover exact scope IDs, retry relevant recall, and report missing
evidence honestly. Capability inventory now includes direct instruction paths.
Tests cover discovery, unchanged scope filtering, current supersession, invalid
source rejection, CLI output, and projected instructions. Fresh Pi retest is
described below.

Fresh noninteractive Pi retest passed on candidate `e83954c`, using the same
model (`openai-codex/gpt-5.5`, thinking off), exact synthetic prompt, agent, and
project cwd with no continued thread. The trace shows scope inventory followed
by explicit recall of the stored project scope and its current renamed revision.
It read the capability's listed instructions directly and ran the unchanged
implementation successfully. No repository-wide file search or invented
capability command occurred. Source validation passed and capability files
remained unchanged. A journal correction was appended for the earlier false
success report; the original event and memory revisions were preserved.

The incorrect-name failure is resolved in this retest, but one guessed scope
query remained in the first tool batch before the model consumed the inventory.
It also reported the containing assistant repository rather than the precise
memory subdirectory. These remain answer/discovery refinements, not evidence
of lost memory. Loading and this prompt's context delivery do not prove every
lifecycle callback or unattended learning behavior.

Verification: full native `mise exec -- bin/ci` passed, including race tests,
plus a Linux/AMD64 CLI cross-build. The disposable trial binary was updated;
its previous executable was retained for rollback. No source-format migration,
global installation change, or private capability publication was required.

## Pi-authored correction recalled by Codex — 2026-09-07

The user requested another synthetic project rename in a fresh Pi session.
Inspection confirmed a third immutable revision on the same record and stable
scope, explicitly superseding the prior revision. Authorship identifies Pi,
its selected model, the originating device, and session. A concise evidence
record and journal entry were committed; no capability files changed.

A fresh Codex noninteractive session received only a question about the current
and previous names, without any names, scope IDs, or record IDs in the prompt.
Its three tool calls discovered scopes, recalled the correct current revision,
and read the full history. The answer reported the current name and all three
names in chronological order. The run exited successfully and left the private
source unchanged. Source validation passed. This demonstrates cross-harness
continuity on one machine, not thread transfer or cross-machine synchronization.
The verification used the documented
[Codex noninteractive flow](https://learn.chatgpt.com/docs/non-interactive-mode),
without resume and with explicit EOF on stdin for the programmatic runner.

Pi still issued one guessed-scope lookup before reading its discovery result.
It also used a predictable temporary JSON pathname for its write and left that
synthetic scratch file behind. Neither affected the saved revision, but scratch
file isolation/cleanup and needless lookups remain workflow improvements.

## Automated import and offline recovery

`TestPortableImportOfflineRecoveryAcrossDevices` exercises the command layer
through new setup, local Git publication, clone/import into a second instance,
and a second-device correction while that clone's remote is unavailable. It
checks the offline write returns `pending` with a clean committed worktree,
then restores the remote while retaining independent online work from the
first instance. Recovery runs through request-completed and request-received
hook entrypoints, without explicit sync commands. Both clones and the bare remote must converge to the same
commit, retain the correction/history and the independent journal entry, and
contain both device registrations with the correct authorship.

This uses disposable local repositories, synthetic records, and separate
instance bindings on one host. It does not test real network authentication,
an actual second physical machine, or remote cloning through the wizard.
The focused regression passed three consecutive race-enabled runs; the full
Go race suite, command package vet, and diff checks also passed. No installed
executable changes are needed for this test-only addition.

## Authenticated remote pilot — 2026-09-07

With explicit owner approval, created a separate private hosted repository
containing only fresh synthetic agent source and test configuration. Two
disposable installations used the existing `e83954c` executable (SHA-256
`8b0e0cb320625f484f426b4ec1ba188d31c8748786dfb52c2f58e1bd966516d7`).
No prior pilot memory, sessions, authentication files, or instance bindings were
uploaded. The original pilot and live assistant repositories were not changed.

A private test helper restricted credentials to the exact host/repository and
verified the explicitly selected development account. It rejected wrong-host,
wrong-repository, and missing-account inputs without credential output. Tokens
were passed only through the Git credential protocol; the helper stored none.
Neither the helper implementation nor service configuration is bundled in the
public toolkit. Existing development authentication was used without changing
global Git helpers or account selection; separate assistant account roles were
not under test.

Observed outcomes:

- HTTPS publication, authenticated clone, and new-instance import succeeded.
- Unavailable credentials produced `pending` while preserving a committed local
  correction; independent online work continued from the other installation.
- Restoring credential access reconciled both histories. Both worktrees and the
  remote converged, retaining the correction, prior revision, independent journal,
  and both device registrations. Source validation passed.
- A request-start adapter invocation fetched a subsequent remote correction
  before explicit scoped recall. A completion adapter invocation published a
  pending journal entry. A later launch (forwarding harness help, without a
  model session) pulled that entry into the other installation.
- Remote privacy was checked after publication, and the tracked-file inventory
  contained only expected source/configuration. Test artifacts remain available
  privately for further investigation; no automatic deletion was performed.

These are two installations on one physical machine. Credential refusal
simulated the outage; this did not disconnect the host network. Adapter calls
were driven directly, so this test does not establish that every native harness
callback is delivered correctly over a real network. Remote creation and clone
were test orchestration, not new wizard features. A physical second-machine
trial, bootstrap ergonomics, and longer-running failure modes remain open.

## Next checkpoints

1. Exercise bootstrap and continuity on a second physical machine.
2. Improve scratch-file handling and remaining retrieval inefficiencies.
3. Review remaining install/update/recovery and provenance gaps before release.
