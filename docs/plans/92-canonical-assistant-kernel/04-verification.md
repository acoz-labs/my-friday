# Verification And Release Design

## Test Strategy

### Pure contracts and plans

- Add `internal/assistantrepo` table tests for path/remote normalization,
  manifests, schemas, module registry, plan determinism, stable states, preview,
  JSON read model, and secret-shaped input rejection.
- Extend `internal/plan` and `internal/profile` characterization tests before
  changing the default bootstrap layout.
- Add migration fixtures for every released legacy repository/instance contract
  supported by B1. Verify exact field/path mapping, source HEAD/tree provenance,
  Unicode, empty memory, populated instruction-only source, and all collision
  denials.

### Git transaction integration

Use local bare repositories behind an injected test transport to prove the Git
state machine. The production endpoint parser remains active in its own tests
and rejects local/file transports; only the test resolver substitutes transport
execution after that boundary.

- empty-remote initial commit/push and exact `origin/main` tracking;
- clean successor commits that stage only declared paths;
- refusal for dirty/staged/untracked/conflicted/ignored-owned state, detached or
  alternate branch, extra worktree locks, mismatched endpoint, nonempty initial
  remote, behind/ahead/diverged refs, submodules/LFS, hooks, filters, unsafe
  transports, and credential-bearing URLs;
- no pull, merge, rebase, reset, checkout of unrelated paths, force flag, tag,
  prune, provider API, or undeclared remote/ref appears in the Git trace;
- ambiguous push failure rereads the remote and distinguishes predecessor,
  candidate, and foreign commits; and
- commit-success/push-failure recovery is idempotent and never duplicates the
  semantic receipt.

### Filesystem and generated-state integration

Extend the current transaction and named-instance suites with no-follow path
replacement, ownership/mode/link-count, quarantine collision, concurrent
operation, signal at every phase, process-group cleanup, canonical-repository
canaries, and generated-root removal proofs. Every fault injection must verify
the exact allowed predecessor/candidate state and preserve foreign entries.

### Command and experience tests

Golden terminal tests cover 80-column and narrow rendering, no-colour output,
keyboard-only confirmation/cancellation, EOF/INT/TERM, paths containing spaces
and Unicode, actionable degraded states, and redaction. Help tests make the
canonical flow primary and keep legacy migration/recovery discoverable without
presenting split creation as the recommended path.

Accessibility is semantic rather than visual: headings/labels must remain
meaningful when read linearly, action/status may not rely on colour or symbols,
the focus remains at the terminal input, and destructive/generated versus
preserved/canonical effects are explicitly named before confirmation.

## Red/Green Sequence

1. Characterize the released F0 pair, instance, and capability contracts and
   lock the compatibility window.
2. Add failing canonical manifest/module/path/remote plan tests; implement pure
   repository values and preview.
3. Add failing empty-remote Git transaction tests; implement candidate tree,
   exact commit/push, and healthy verification.
4. Add phase-by-phase interruption and ambiguous-push tests; implement journal
   recovery before adding more mutation verbs.
5. Add failing host-binding create/restore/reconcile/repair/remove tests; adapt
   the named-instance boundary so canonical source remains external.
6. Add legacy import and switch-boundary failures; implement migration while
   preserving the old pair/instance.
7. Add migration-chain tests; implement baseline upgrade and forward-history
   rollback.
8. Add complete CLI/error/help/golden tests and native terminal signal cases.
9. Run the immutable-candidate clean-machine journey against a real empty
   private remote and repair all comprehension or recovery failures without
   weakening refusal rules.

## Acceptance Evidence

Automated evidence must include `bin/container bin/ci`, native `bin/ci`, race
tests, the Git trace-denial suite, every operation-journal phase, and complete
repository/remote/generated-state canary equality.

The exact candidate is the nominated Darwin/ARM64 artifact tuple and commit.
Native acceptance on a clean supported Apple-silicon account uses a newly
created private test remote with no refs and exercises:

1. create, inspect, local verify, remote verify, launch, and fresh-task identity,
   including the visible user-attested privacy label;
2. one exact-path canonical change through the shared steward, commit and push;
3. projection drift diagnosis and repair;
4. commit-success/push-failure recovery and push-success/projection-failure
   recovery without duplicate commit;
5. remote divergence refusal with no merge/rewrite;
6. compatible upgrade and forward-commit rollback;
7. migration from a released populated split pair while the old pair remains;
8. generated-state removal while canonical local/remote source and memory
   remain byte/tree equal; and
9. `assistant restore` on a second clean host from the same remote without a new
   semantic commit, proving the host binding and projection are replaceable.

The product owner and independent acceptor must be different from the sole
implementation contributor. B1 acceptance records task completion and boundary
comprehension against the immutable candidate; B4 later owns the three-person
full-MVP retention gate.

No browser or rendered GUI evidence is required. Terminal transcript evidence
must be sanitized and bind scenario outcomes/digests without embedding the
private remote URL, user content, absolute home paths, or credentials.

## Rollout

1. Land the implementation behind the new canonical command path while
   retaining released legacy verify/recover/migrate behavior.
2. Reconcile implementation against this plan, promote durable architecture,
   product, security, development, deployment, and runbook docs, and delete
   this temporary plan before implementation review.
3. Build one immutable candidate after F0 release, run portable and native CI,
   then nominate the exact artifact.
4. Perform clean private-remote acceptance including interruption and migration.
5. Publish through the existing stable Darwin/ARM64 GitHub Release asset only
   after independent acceptance and release-gate verification.
6. Announce the canonical path as the default and the split path as legacy with
   a migration command; do not auto-migrate existing users.

## Rollback And Recovery

Before release, rollback is ordinary PR/release reversal. After release, retain
the previous immutable artifact and its documented recovery commands. A user
who has not migrated can continue the released split lifecycle. A migrated user
can remove the new generated projection and relaunch the preserved legacy
instance; My Friday does not delete the new canonical repository or rewrite its
remote to simulate rollback.

Within B1, compatible baseline rollback creates a new descendant commit through
the same transaction. Interrupted actions use `assistant recover`; remote
divergence requires the user to resolve Git history externally and rerun verify.
The tool never recommends force/reset as recovery.

## Release Prerequisites

- #74/#83 accepted and an immutable instruction-only foundation release exists.
- Every embedded migration targets an actually released input contract.
- A disposable private remote and clean Apple-silicon account are available for
  exact-candidate acceptance; endpoint and credentials stay outside evidence.
- Native Git, APFS, terminal, signal, process, migration, removal, and second-
  host restore evidence passes.
- The stable release archive, checksum, SBOM/notices, and rollback artifact pass
  the existing release gate.

## Production Readiness Preflight

This artifact has no service deployment or runtime secret slot. Production is
the immutable GitHub Release artifact, and publishing that already accepted
artifact is the only production activation. Preflight must verify:

- exact accepted commit and artifact digest;
- CI and native acceptance receipts, including private-remote Git and
  clean-host restore scenarios;
- no credential or private-endpoint value in tracked source, logs, fixtures,
  evidence, archive, checksum ledger, or release notes;
- stable asset name and prior-release rollback availability;
- legacy recovery/migration documentation; and
- release finalization against the same bytes. No rebuild after acceptance is
  permitted.
