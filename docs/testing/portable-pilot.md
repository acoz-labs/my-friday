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

## Installation separation and generated-file repair — 2026-09-07

Review found that explicit setup paths could nest machine-local instance state
inside versioned source, allowing subsequent checkpoints to include native state.
Regression tests reproduced acceptance of that layout. Setup now rejects
source/state overlap before source creation or imported-device registration;
binding creation/load and launcher installation enforce the same separation.
Existing symlink aliases and prospective descendants are resolved for the check;
macOS conservatively refuses case-only overlaps too. No live installation was
found affected during this work, and no authentication data was used in fixtures.

Generated-file refresh previously followed symlink targets and truncated existing
inodes. Tests cover symlinked directories/files and hard-linked file canaries.
Projection now preflights all managed destinations and replaces files atomically.
Native auth/session files are not managed targets. Multi-file crash atomicity and
hostile same-user races remain outside this guarantee.

`agent repair --instance PATH` regenerates missing/stale managed files without
network synchronization, source mutations, or binding changes. An explicit
`--launcher PATH` can create a missing launcher but never overwrite one. Invalid
source/bindings must be investigated, not silently repaired. This is generated
configuration recovery, not an automatic executable updater or data migration.

Verification: full native `mise exec -- bin/ci` passed, plus Linux/AMD64 CLI
cross-build. A compiled-executable smoke in a separate disposable installation
restored a missing extension and launcher, refused a launcher collision and a
symlink targeting synthetic auth, and rejected nested setup before source
creation. Binding/auth canaries stayed byte-identical; source validation passed
and its Git worktree remained clean. The existing pilot's source/auth state was
not part of these destructive fixtures. Prior executable artifacts are retained
when updating the trial at its stable bound path.

## Physical second-machine offline portability — 2026-09-07

Transferred the tested `f4e10eb` Apple Silicon executable to a second physical
Mac over SSH with an independently verified host key. Its SHA-256 matched
`5fac4770b6fa183c3cf015aab9961d5bff607345c184905e1a0468534cc576af`.
The target ran macOS 26.6.2 and Git 2.55.0. All installation paths were disposable;
existing assistant installations and global SSH/Git configuration were unchanged.

The target's SSH session returned GitHub CLI HTTP 401, including
with token environment overrides removed. Used a Git bundle of the synthetic
remote pilot source to exercise offline import without copying credentials or
native sessions. This was test orchestration, not a new bundle-import wizard.

Observed outcomes:

- Clone/import preserved the assistant identity and registered a distinct device.
- Scoped recall recovered the current synthetic project name and its complete
  three-revision history. The imported record retained its machine provenance.
- A correction authored on the second Mac superseded the current revision and
  carried the new device ID. Intentional credential refusal returned `pending`;
  the correction was committed, the worktree was clean, and validation passed.
- A return bundle brought that correction to the first Mac. Publication through
  its working private test credential helper succeeded, and another installation
  pulled and recalled the correction with the second Mac's provenance intact.
- Fresh target harness homes contained no copied Codex or Pi authentication files.

This proves physical-machine CLI continuity and offline round-trip recovery.
Remote Codex 0.153.0 was inventoried but not used for a model request.

### Direct authenticated sync follow-up

The owner's local Terminal successfully retrieved the selected account's token
and verified its identity, with no token environment overrides. SSH could not
retrieve that token. Local Terminal selected a newer GitHub CLI, but explicitly
running that same executable over SSH still failed. The login itself was valid;
session-specific credential access remains the unresolved SSH limitation.
No credentials were printed, copied, or reset during these checks.

The owner then ran the same My Friday sync command directly in the target's
local Terminal using the private helper's explicit account selection. It returned
`synced` with identical local and remote commit IDs matching the previously
published physical-machine correction. This establishes direct authenticated
reconciliation from the second Mac, but not publication of a new correction
from that Terminal or native model lifecycle delivery. A live harness test is
next; the isolated test Codex home still requires its own login.

## Physical second-machine live Codex pilot — 2026-09-07

The owner authenticated the disposable Codex home locally and launched the
assistant from a fresh, empty project directory outside its source repository.
The private test helper gained a non-reserved account-selector variable because
the launcher clears inherited `MY_FRIDAY_*` variables before supplying runtime
identity. Startup pulled that private helper update; no service-specific code
was added to the public toolkit.

Reviewed the native session's messages and executed commands, then independently
pulled the resulting source on the first physical Mac. Observed:

- Scope discovery selected the existing project ID without guessing from cwd.
- The assistant recovered the current name and four-revision history, then saved
  a fifth revision superseding the previous name under explicit user direction.
- Source evidence, revision, and completion journal were published directly from
  the second Mac. Local and remote heads matched; the first Mac then recalled
  the same new revision with second-device, Codex, and session provenance.
- All five revisions remained available, with no current conflicts. Source
  validation passed and the source worktree was clean. The project stayed empty.
- The memory write used a unique temporary file with cleanup in a finally block,
  rather than a predictable shared scratch filename.
- Native request-hook additional context was visible in the transcript. Explicit
  memory writes performed synchronization, so this result does not isolate the
  completion hook's publication behavior or prove every lifecycle callback.

Two follow-up defects were visible despite the correct final answer:

1. Scoped query `name` did not match the current record's `named` wording. The
   assistant recovered with an empty query in the same scope; lexical retrieval
   still causes avoidable calls. Scope and supersession filtering must remain
   intact when improving matching.
2. The native skill catalog included unrelated user-home skills despite the
   separate Codex home. None were invoked in this task, but generated-home
   separation alone does not isolate inherited capability discovery. The
   [official OpenAI skills documentation](https://learn.chatgpt.com/docs/build-skills)
   describes user-home discovery and per-skill disable configuration. A tested
   inheritance policy is still needed; no existing global skills were modified.

## Discovery exclusions and lexical retest — 2026-09-07

Historical experiment: the exclusion policy below was superseded by the owner's
native-inheritance decision later on the same date. The lexical fix remains.

Implemented Codex launch-time exclusions for discovered user-home skills using
the documented per-skill native configuration. No global skill or home setting
was modified. Regression tests cover linked paths, directory cycles, broken
targets, quoting, and discovery refresh on another launch. An opt-in model-free
Codex 0.153.4 contract test verifies that a synthetic linked user skill changes
from enabled to disabled while a synthetic project skill remains enabled.

A separate disposable installation on the physical second Mac exercised the real
launcher against Codex 0.153.0: native discovery reported all 17 unrelated user
skills disabled and six native skills enabled. This used no copied credentials
and made no model request. The installed app-server schema lacks a newer
extra-user-root parameter in current docs; the committed contract test instead
uses a supported disposable instance-user root.

Added limited English prose inflection matching with lower weight than exact
hits. Tests reproduce `name` versus `named`, exercise silent-e and plural forms,
and reject arbitrary substrings, numeric/compound identifiers, other scopes,
and superseded keyword hits. Existing effective-time and conflict tests pass.
This is a bounded lexical improvement, not general linguistic stemming.

Full native `mise exec -- bin/ci`, an opt-in race-enabled native discovery run,
Linux/AMD64 CLI cross-build, and diff checks passed. Built clean commit `8659297`
as SHA-256 `79e37e37c2d3272623356f60cd9a3b2f27289b34dd4578a7d82c44c22b7fd133`.
Updated both disposable pilot executables at their existing bound paths, retaining
the previous executable beside each as a rollback copy.

A fresh read-only Codex conversation on the second Mac used that executable and
the already-authenticated test home. Its injected skill catalog omitted the
unrelated user-home collection. The model discovered the project scope and
returned the current name from its first `name` query without a fallback. Its
only shell operations were scope discovery, cwd reporting, and scoped recall.
Source HEAD/worktree and the empty project directory remained unchanged. This
SSH session did not establish working GitHub credential access; authenticated
publication remains covered by the earlier local-Terminal pilot.

Exclusions are a startup snapshot, not filesystem isolation or complete control
of administrator/account-provided capabilities. Fresh native sessions are required
to retest changed catalogs. Pi discovery beyond its configured instance home,
installation/update UX, and longer-running lifecycle failures remain separate
checks before release.

## Native inheritance decision supersedes blanket exclusions — 2026-09-07

The owner clarified that host skills should normally supplement the portable
assistant. A daily-use machine can contribute useful tools, while a headless
installation may have few or no additions. Continuity means shared identity,
memory, and assistant-owned capabilities, not identical tool catalogs everywhere.
Inherited skills are not themselves a defect or a security boundary; actual
instruction conflicts and unavailable host dependencies need scoped diagnosis.

Removed the Codex exclusion injection and its scanner-specific tests. Discarded
the uncommitted Pi settings-exclusion experiment before installing it in a pilot.
Neither harness now receives blanket exclusions from My Friday; Pi native
settings remain untouched. No source-format or memory migration is involved.
Generated guidance now distinguishes host/project prerequisites from synchronized
capabilities and asks the assistant to verify them before reusing procedures.

The native Codex regression now exercises the actual launch plan with and
without a synthetic linked user skill; project skills remain enabled in both
cases. A Pi 0.85.1 resource-loader regression verifies linked user and project
skill discovery, generated lifecycle extension registration, and byte-identical
native settings. Both are model-free, opt-in tests using disposable state.
These checks do not claim every host is skill-free when a fixture is absent.

Pi's inspected implementation also discovers user-home `.agents/skills`, plus
trusted project `.pi` resources and ancestor `.agents/skills`. Its default
extension directories are instance/project scoped. Explicit settings, packages,
CLI paths, ancestor instructions, and extensions can contribute more resources;
native trust and instruction precedence remain applicable.

Verification: full native `mise exec -- bin/ci`, race-enabled portable tests with
both opt-in native discovery checks, Linux/AMD64 CLI cross-build, and diff checks
passed. The lexical regression remains green. The existing second-Mac SSH
connection expired during this follow-up; its installed executable still needs
the inheritance update and a fresh native-catalog check after reconnecting.

Built clean commit `92f0a35` as SHA-256
`027db6d2c2ba969e1739d95e7433a5daa3aaec6fff442ad601caab801be03a95`.
The first pilot's bound executable is updated with its previous build retained
for rollback. Offline repair refreshed its generated instructions without
changing source, binding, authentication, or sessions.

## Second-Mac inherited-catalog retest — 2026-09-07

After reconnecting the scoped SSH session, installed the same `92f0a35` artifact
on the second physical Mac. Verified its SHA-256 before replacement, retained the
previous executable for rollback, and atomically replaced the bound executable.
Offline repair refreshed generated files; the binding hash and native auth-file
size, modification time, and inode stayed unchanged. Source validation passed.
An older open conversation was left running; verification used fresh processes
rather than assuming its loaded instructions had changed.

The actual launcher against Codex 0.153.0 returned 17 inherited user skills
enabled, 23 enabled skills total, and no discovery errors through `skills/list`.
A fresh authenticated read-only model conversation listed the inherited catalog
separately from the assistant-owned capability inventory. Its only four shell
operations were scope discovery, cwd reporting, capability inventory, and scoped
recall. The first `name` query returned the current synthetic project name with
no conflicts; no fallback lookup, skill-instruction reads, or assistant-source
or project edits occurred. Native session and runtime state stayed machine-local.

Both checks preserved source HEAD, a clean source worktree, and the empty project
directory. The private pilot helper deliberately reported offline, and launch
continued with visible pending sync. This does not establish GitHub credential
access over SSH or replace the earlier local-Terminal publication evidence.
Codex emitted hook-trust-bypass warnings as error-typed stream items; these were
warnings for the intentionally configured launch mode, not failed operations.
Private transcripts and inherited skill names remain outside this repository.

## Next checkpoints

1. Diagnose/repair/diagnose hands-on flow passed; continue with native interactive
   warning/interruption feedback.
2. Review remaining install/update/recovery and provenance gaps before release.
3. Check native interactive warning/interruption behavior and define the initial
   release's supported boundaries before migrating a live personal assistant.

## Recovery diagnostics and lifecycle failure boundaries — 2026-09-07

Added `agent doctor`: read-only source validation, own-Git-directory presence,
bound-executable metadata, selected-harness PATH lookup, projection path safety,
and byte comparison against the running toolkit's generated files. Findings
include remedies and return a nonzero status. Generated-file remedies contain an
explicit quoted repair command. Doctor never executes the harness or a helper,
reads authentication, initializes native state, or performs synchronization.
Missing/corrupt binding remains a separate refusal, not an inferred new identity.
Launcher placement, native version compatibility, credentials, remote access,
and active-session context are explicitly outside its healthy-result guarantee.

Tests reproduce and correct two lifecycle defects: trailing output/null accepted
as a successful handler response, and a cancelled warn-policy dispatch marked
completed after proceeding through the remaining chain. Adapters now surface
warn-policy failures and synchronization attention, including Pi completion UI
notifications. The Pi request callback preserves available memory context when
also reporting a subscription error. Native callbacks share an internal deadline
shorter than their configured harness timeout, propagated through sync and
subscribers; ordinary Git/helper process groups are cancelled as well.

Regression coverage includes malformed output, warn/stop ordering, cancellation
and replay refusal, no delayed effect from a cancelled ordinary Git child,
Codex/Pi warning response shapes, completion sync warnings, doctor/repair/doctor
round trips, missing Git metadata, inherited resource discovery, and a native Pi
callback invoked with a synthetic hook executable. Generated guidance now prefers
stdin for memory drafts and private temporary directories for scratch files.
This guidance is not a sandbox or an automatic secret/scratch-file filter.

Full native CI passed after narrowly allowing doctor's read-only `LookPath` in
the legacy subprocess-boundary test. Race-enabled portable/CLI tests with native
Codex/Pi checks and a Linux/AMD64 CLI cross-build also passed for the candidate.
No real private capability, production memory, native authentication, global
skill settings, release, or source-format migration is part of these changes.
Deadline handling does not preempt filesystem validation/copying, contain
deliberately detached descendants, or undo external effects already performed.

Built clean code commit `344f660` as SHA-256
`0c48876738181dc01d73992bad0e67f447623398752aa7ddd2172b5a4ea44945`.
An executable-level disposable smoke test created a new local-only installation,
confirmed healthy diagnostics, moved one generated hook file to a retained
backup, confirmed the specific finding without repair, repaired it, and confirmed
healthy diagnostics again. Source repository bytes and binding bytes were
unchanged throughout diagnosis/repair, and the restored hook matched its backup.
A separate equally disposable fixture is left at the missing-hook stage for
hands-on clarity testing. Neither fixture contains native credentials or real
memory; existing authenticated pilots were not modified by this checkpoint.

## Hands-on recovery completed — 2026-09-08

The user ran doctor in Terminal and received the single expected missing-hook
finding. They then ran repair and repeated doctor; the final report was healthy
with every listed check passing. This closes the functional hands-on recovery
scenario, not broader authentication or release acceptance. The repair command
was also repeated in the conversation, so this is not proof that the embedded
remedy alone was sufficiently clear.

The initial failure misleadingly printed `Error [input.invalid]`. Doctor now
returns a typed unhealthy-installation finding, classified as
`installation.unhealthy` with exit status 3. Regression tests cover the actual
doctor/repair command flow, wrapped findings, and retention of input-error
classification for ordinary invalid arguments. The user-tested artifact remains
the earlier recorded build; its executable was not silently replaced.
Independent inspection found a clean source worktree and restored hook bytes
identical to the retained backup. Full native CI, race-enabled portable/CLI
tests, a Linux/AMD64 CLI cross-build, and diff checks passed for the label fix.

## Native warning and interruption probes — 2026-09-08

Built clean `e88effe` as SHA-256
`8d277ba3837722e72d050c1a2dcb31f7fe31ea6220e60c53b337e0122abc0227`
and updated the first authenticated disposable pilot at its bound executable
path, retaining the prior artifact. Repair refreshed generated files without
copying authentication. Added a private synthetic lifecycle canary, gated by an
explicit test-workspace environment marker and matching event cwd. Ordinary
pilot sessions do not activate its intentional failure/slow-interruption paths.
The fixture and private transcripts are not distributed in the public toolkit.

Pi 0.85.1 RPC emitted a native `extension_ui_request` warning for the failed
subscriber and completed read-only recall correctly. After a separate synthetic
shell command reported readiness, RPC abort settled the run in about 0.11 seconds.
The native message reported that the operation was aborted; a subsequent prompt
in the same session recalled the same current name and did not replay the command.
This proves RPC notification/abort/recovery, not visual presentation or an
explicit portable Pi `request.interrupted` event. Pi's `agent_settled` occurs after
aborted runs too; `request.completed` is currently a synchronization checkpoint,
not a success assertion.

Codex 0.153.4 `exec` ran the request subscriber, recorded its deliberate failure,
and the model acknowledged the warning while recalling correctly. A dedicated
warning item was not observed in the captured exec stream; UI visibility remains
a separate hands-on check.

The app-server transport behaved differently: `hooks/list` reported generated
hooks enabled but untrusted, and model turns produced no corresponding canary
dispatch receipt or warning. This is not a passing lifecycle test, despite correct
memory retrieval through explicit commands. A later app-server probe confirmed
an active turn could be interrupted and the next prompt could recall correctly,
but did not prove My Friday interruption-handler delivery. The first interruption
driver awaited a readiness output delta that this version did not emit and timed
out; the corrected driver uses the native command-start event. No claim is made
that its earlier un-interrupted command was cancelled.

After the recorded fixture addition, source HEAD/worktree remained unchanged,
source validation passed, and the test project directory stayed empty. A scoped
Terminal launcher is prepared for the owner's warning/ Escape-interruption check.
The second physical Mac and the separate recovery-only fixture are unchanged.
