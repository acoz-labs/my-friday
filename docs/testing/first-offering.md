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

## Remaining before initial rollout

1. Complete the conversational source-checkpoint provenance/`agent changes`
   check. Automated creation/edit/delete, cross-clone, offline retry, malformed
   remote and device-binding checks passed on 2026-09-08; they do not close the
   human-facing discovery check.
2. Apply the rehearsed executable-update flow to the disposable authenticated
   pilot and start a fresh conversation. The separate black-box rehearsal passed
   on 2026-09-08: preserved binary, atomic bound-path replacement, doctor drift,
   repair, healthy doctor, unchanged launcher/cwd, and retained synthetic auth.
   Actual native login/session continuity still needs the conversational check.
   Production placement must use durable paths, never a `/tmp` pilot binding.
3. Choose and explicitly authorize the artifact/release route. Feature-branch
   pushes are development checkpoints, not releases. The repository still has
   legacy release automation; its existence does not make this rebuild released.
4. Separately choose Alfred's private source and instance locations and stage a
   reviewed import. Keep the existing agent recoverable. Establish identity and
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
