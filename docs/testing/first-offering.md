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

The currently identified hands-on regression checkpoints are closed. No repeat
is required without a new finding. This does not imply every future capability,
transport, harness version or live-data import has been qualified.

1. Choose and explicitly authorize the artifact/release route. Feature-branch
   pushes are development checkpoints, not releases. The repository still has
   legacy release automation; its existence does not make this rebuild released.
2. Separately choose Alfred's private source and instance locations and stage a
   reviewed import. Use durable executable/source/instance paths, never `/tmp`
   pilot bindings. Keep the existing agent recoverable. Establish identity and
   memory first, then validate one private capability at a time. No service
   credentials, GitHub role policy, or email workflow enters this public toolkit.

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
