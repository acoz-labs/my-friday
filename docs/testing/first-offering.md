# Initial portable offering: bounded rollout checklist

This is a development/rollout checklist, not a release announcement or authority
to migrate a live assistant. The [portable guide](../portable-assistant.md) is the
implemented contract; the [pilot record](portable-pilot.md) distinguishes actual
observations from automated coverage. Do not repeat already-passed user tests
without a regression or new behavior that needs them.

## First-use support boundary

- macOS on Apple silicon, Git 2.38+, separately installed/authenticated harnesses.
  Linux cross-builds are useful evidence, not a Linux end-to-end support claim.
- Normal named-launcher Terminal sessions using Codex or Pi, from any project
  directory. The exercised versions are recorded in the pilot, not a promise
  that arbitrary newer or older harness versions work.
- Assistant-owned private source, structured memory, private capabilities, and
  machine-local instance/authentication/session state; native skill inheritance
  is intentional. Full user access, not a sandbox or separate OS identity.
- New-format local setup/import, ordinary Git-backed continuity and explicit
  offline/pending/conflict behavior. Native login and initial cloning/remote
  creation remain manual. Private network authentication helpers are not bundled.
- Codex app-server is not established as a supported automatic-hook transport.
  The observed trust gap remains unresolved. Do not advertise app-server or
  arbitrary remote chat bridges as equivalent to the tested Terminal modes.
- Pi completion is a checkpoint after settling, including abort, not task-success
  evidence. Dedicated interruption subscriptions are not implemented for Pi.

## Already demonstrated

- Named launch, project cwd preservation, recall in fresh sessions and across
  harnesses, explicit supersession, and private capability creation/reuse.
- Cross-clone local/HTTPS synchronization; offline commits, durable correction
  history and conflict preservation; second physical Mac import, new-device
  memory authorship and native-Terminal publication/recall.
- Native host skill inheritance, isolation of machine-local session/auth state,
  and a real doctor → repair → healthy roundtrip.
- Visible warning, interrupted command, and successful next read-only recall in
  both Codex and Pi Terminal. Transport-specific limitations remain explicit.
- Source-checkpoint provenance passed the owner-driven Codex Terminal check on
  2026-09-08: useful private documentation edit, checks, journal/local checkpoint,
  accurate device label and observation/authorship distinction.
- Authenticated executable upgrade and fresh conversation, followed by the
  native-settings/parent-alias regression on 2026-09-08. Doctor stayed healthy
  before, during and after the native session; both parent path spellings agreed.
  Native settings survived repair, normal UI state was retained, and read-only
  recall left assistant source and the empty working project unchanged.

## Remaining before initial rollout

The identified hands-on regression checkpoints are closed, including the
reference-aware capability-building pilot on 2026-09-09. Codex turned useful
historical failures into tests and rejected obsolete rules; Pi discovered and
reused the resulting capability, rationale and hash-verified references in a
fresh read-only session. Both preserved the external library/project. This is
bounded behavioral evidence, not a general instruction-injection guarantee.
No repeat is required without a new finding. See
[reference libraries](../reference-libraries.md) and the detailed pilot evidence.

On 2026-09-09 the owner authorized a pinned durable private installation, leaving
public release separate. The exact tested `864601c` artifact was installed at a
versioned durable path, with a fresh private source/instance and named launcher.
The tested Pi runtime was preserved at a durable path too. Existing systems and
the pilot remain intact. Historical repositories were cloned separately and
registered as reference-only libraries; no old memory claims, native credentials
or provider capabilities were imported. Both harnesses passed structural doctor.
Machine paths and private installation details stay outside this public repo.

The owner subsequently completed scoped ChatGPT login and a first read-only Codex
conversation from outside the assistant repository. Inspection confirmed correct
identity/path reporting, empty private capability/current-memory inventories,
and a clear distinction between historical references and active guidance. No
historical contents or credentials were read; source HEAD/worktree were unchanged
and post-session doctor passed. Some help and metadata discovery was redundant;
no task command failed. The first durable Codex conversation checkpoint is closed.

1. Complete the separate native Pi login when enabling that harness for daily use.
   The synthetic cross-harness behavior is already qualified; this instance has
   not yet had an authenticated Pi conversation.
2. Confirm the private remote destination and configure/verify source-sync
   credentials deliberately; local checkpoints alone do not provide cross-machine
   continuity. A developer bootstrap account is not an implicit agent role.
3. Rebuild one private capability at a time from the current ask and relevant
   references. Facts/preferences/commitments still need separate reconciliation.
   No service credentials, GitHub role policy or email workflow enters this core.

Public merge/release remains separately authorized work. Feature-branch pushes
and this private rollout do not invoke the repository's legacy release automation.

## Follow-ups, not prerequisites for this bounded first use

Remote creation/clone wizard, harness installation/login automation, automatic
binary updates and source-format migrations, richer provenance claims beyond
checkpoint observations, QMD/vector retrieval evaluation, Linux end-to-end
qualification, app-server hook trust, additional lifecycle parity, scheduling,
bridges and dashboards. An import of the existing Alfred is its own private
rollout task, not a reason to bundle that agent's integrations in public core.

## Recovery boundaries

Generated-file damage has a tested repair path. Repair does not relocate/update
the executable, restore native login, resolve Git conflicts, or repair a corrupt
binding. Missing/corrupt bindings and partial setup need explicit inspection and
recovery/import into fresh destinations while preserving the original. These
limits must be visible in the installation guide, not disguised as automatic
recovery. No retrospective provenance or thread portability is promised.
