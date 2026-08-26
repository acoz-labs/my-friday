# Verification And Release Design

## Test Strategy

- Add `internal/capabilityworkshop` unit tests for answer editing, enhancement
  loading, deterministic rendering, escaping, full diff, exact confirmation,
  navigation, every numeric input bound, validation feedback, arbitrary valid
  SKILL-body retention, explicit canonical-body regeneration, and
  state-specific current/post summaries.
- Add source-transaction tests with fault hooks at journal/stage/quarantine/
  promotion/sync/cleanup boundaries. Prove create/update success, no-replace,
  stale preview, shared-lock refusal, old-or-new recovery, unsafe-entry refusal,
  `Inspect`/lifecycle state precedence, workshop-only source recovery, and no
  deletion without exact inode/digest authority.
- Extend `internal/capability` tests only for narrow exported helpers or state
  integration; keep the strict parser/test/lifecycle regression matrix green.
- Extend `cmd/my-friday/main_test.go` for usage, real-home authority, TTY gate,
  injected interaction, stable error classification, and proof that workshop
  success never calls lifecycle mutation.
- Replace live-agent authoring paths in acceptance support, supervisor,
  evidence, deployment, and receipt tests with deterministic workshop paths.
  Retain runner signal/quiescence, auth non-exposure where still needed for
  fresh Codex invocation, exact artifact, ambient canaries, lifecycle, and
  three-person receipt coverage.
- Add native Expect scenarios on Apple silicon/APFS for narrow terminal,
  Unicode, create, default exit, `b`, `q`, EOF, `INT`, `TERM`, invalid answer,
  exact confirmation, update, optional-byte preservation, separate
  install/upgrade, and interrupted source recovery.

No screenshot baseline applies. This is a plain-text terminal protocol, but it
is a meaningful rendered interaction: exact-candidate evidence must retain an
owner-only transcript or typed step receipt proving headings, action/installed
summary, keyboard-only navigation, wrapping without truncation, and no color-
only meaning. Public artifacts remain redacted.

## Red/Green Sequence

1. Lock current absence: `capability workshop` is unsupported and no renderer
   or source transaction exists.
2. Write failing proposal/render tests, then render a valid three-file package
   whose tests pass through the existing authority.
3. Write failing create transaction and fault/recovery tests, then implement
   lock, journal, stage, no-replace promotion, and cleanup.
4. Write failing update/opaque/stale/concurrency tests, then implement exact old
   quarantine, new promotion, preservation, and recovery.
5. Write failing terminal navigation/confirmation/state tests, then connect the
   injected workshop flow to `runCapability` with the TTY and real-home gates.
6. Replace generated builder guidance and deterministic tests; remove live
   authoring authority while keeping explicit fresh-task capability invocation.
7. Replace acceptance authoring scenarios/receipts and run the complete local
   and portable matrices.
8. Reconcile docs and the superseded #56 plan, merge, nominate new bytes, and
   run independent exact-candidate acceptance from the beginning.

## Acceptance Evidence

Deterministic PR evidence proves canonical bytes, every input/refusal state,
full diff, exact token, no activation calls, source transaction fault recovery,
opaque equality, shared-lock serialization, ambient preservation, and removal
of model completion from authority.

One immutable nominated artifact must then prove:

1. first-time `daily-brief` create, inactive `ready`, and automatic deterministic
   checks;
2. separate Install and fresh-task explicit invocation;
3. enhancement with complete diff and byte-identical optional files;
   an arbitrary valid pre-workshop SKILL body is retained unless regeneration
   is explicitly selected;
4. `source-changed`, separate Upgrade, and fresh-task changed behavior;
5. back/exit/default/EOF/signal no-write behavior and one fault-injected source
   recovery journey;
6. invalid/collision/drift/incompatible refusal;
7. disable/enable/remove and source preservation; and
8. unchanged sibling assistant, runtime, workspace, credential receipt,
   launcher, global state, and all unowned paths.

Anthony and two distinct design partners each record typed comprehension,
completion, recovery, and retention facts against the same issue, candidate,
and artifact. The acceptance schema may gain a versioned workshop-mechanism
field, but it must retain the established identity/distinct-actor checks and
must not accept legacy model-builder receipts for the new candidate.

## Rollout

The execution envelope is `through-production`:

1. merge this reviewed plan after Phase 2 approval;
2. implement from current `origin/main` in a new task-scoped worktree and
   reconcile any reusable #56 dirty work by reviewed cherry-pick/reimplementation,
   never by wholesale copy;
3. run CI, race, shell, plan/reconciliation, portable container, and native
   terminal/APFS checks;
4. merge the implementation and nominate a new immutable artifact containing
   every lifecycle-linked implementation PR;
5. run independent exact-candidate workshop acceptance and collect one product
   owner plus two distinct design-partner receipts;
6. publish the accepted artifact without rebuilding through the existing
   release workflow; and
7. freshly download and verify checksums, contained executable digest,
   version/help, workshop help/denial behavior, and stable latest URL before
   completing #51, #56, and #74 as the release ledger permits.

There is no staging service and no automatic activation in an existing user
root. Named-instance upgrade remains explicit.

## Rollback And Recovery

Before release, revert the implementation and nominate new bytes; never reuse a
failed artifact. After release, publish a newly accepted superseding artifact.
Existing source and installed projections are not mutated by release.

An existing instance may explicitly roll back its capability revision only
under the current empty-control constraints. A workshop source update is not
automatically reversible by My Friday because source is user-owned; the source
transaction restores the old tree only for an interrupted/uncommitted update.
After a committed update, users use their version-control history for content
reversal, then the existing Upgrade flow if they choose to activate it.

Acceptance-run interruption uses the existing exact marked-root recovery and
stop barrier. Source-journal ambiguity preserves only the exact bound paths for
diagnosis; no broad recursive cleanup or force detach is authorized.

## Release Prerequisites

- Green `bin/container bin/ci` and native `bin/ci` on Apple silicon.
- Green targeted `go test -race` for capability, workshop, command, assistant
  instance, runner, and stop-barrier packages.
- Reviewed reconciliation of the old #56 live-agent mechanism and preserved
  uncommitted work; no inaccessible diagnostic root is reused.
- One new merged candidate and immutable artifact; all failed prior candidates
  remain historical and ineligible.
- Independent acceptor, Anthony's product-owner receipt, two distinct design-
  partner receipts, and no unresolved P0/P1 finding.

## Production Readiness Preflight

- **Secrets:** no new secret slot. Existing Codex auth-by-path remains needed
  only for fresh installed-capability invocation acceptance, not authoring.
- **Candidate:** exact main SHA, artifact run/ID/name/digest, implementation PR
  set, workshop/evidence helper closure, and acceptance bundle are bound by the
  existing nomination and release tools.
- **Deploy:** artifact repository with no staging; publish accepted bytes using
  the existing release workflow and stable asset name, without rebuild.
- **Activation:** release and workshop create no user activation. Install,
  Upgrade, Enable, and named-instance Upgrade retain exact plans/tokens.
- **Verification:** evidence verifier, product acceptance, release workflow,
  fresh download, checksums, executable digest, version/help, workshop smoke,
  and stable URL are executable repository paths.
- **Rollback:** explicit assistant rollback, source version-control reversal,
  and artifact supersession are bounded; source is never deleted.
- **Receipt:** final GitHub Release and issue lifecycle bind the exact planning
  PR, implementation set, candidate, artifact, acceptors, and included issues.
