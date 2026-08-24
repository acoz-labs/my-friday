# Verification And Release Design

## Test Strategy

Start with failing tests at the narrowest useful boundary and retain
credential-free fixtures throughout automated CI.

| Contract | Likely test surface | Required cases |
|---|---|---|
| Canonical name and derived root | new/extended plan and profile unit tests | valid names; empty, reserved, uppercase, Unicode, separators, traversal, length, case collision |
| Exact containment | lifecycle and platform tests | root and `$HOME/.local/bin/<name>` derivation; missing/unsafe launcher directory; symlinked ancestors/children; foreign launcher; case-insensitive collision; changed inode; launcher-directory and sibling canaries |
| Manifest ownership | `internal/codexhome` or new assistant lifecycle tests | complete, missing, unknown schema, altered root path, forged external launcher path, wrong launcher artifact, `..`, foreign manifest, idempotent repeat |
| Transaction safety | `internal/transaction` and lifecycle fault tests | stage, fsync, promote, verify, remove, interrupted and uncertain states |
| Launcher contract | subprocess boundary tests | unchanged `HOME`, fixed `CODEX_HOME`, fixed `--cd`, literal argv, managed executable/dependencies, no shell or startup-file access |
| Multi-instance coexistence | terminal/native integration tests | create A/B, distinct state, concurrent launch, remove/update A with B unchanged |
| O2 migration | lifecycle and native acceptance tests | source proof valid/absent/drifted, absent/foreign/prior-owned launcher leaf, destination collision, verified launcher replacement/rollback, verify-before-delete, partial cleanup, resume, no over-delete |
| Credential boundary | config and evidence tests | no values in manifest/log/receipt/snapshot; per-instance file-backed locations |

Run focused Go tests during red/green work, then `go test -race ./...`, native
platform tests, and the repository-standard `bin/container bin/ci`. Run native
`bin/ci` where macOS-specific filesystem and launcher behavior cannot be proven
inside the container.

## Red/Green Sequence

1. Add failing canonical-name, fixed root-layout, and deterministic external
   launcher-path tests; implement pure planning with no mutation.
2. Add failing manifest-schema, collision, containment, and idempotency tests;
   implement read-only classification.
3. Add failing staged-create, root-promotion, external no-replace launcher, and
   partial rollback tests; implement one complete manifest-owned instance
   transaction with no second outside mutation.
4. Add failing launcher argv/environment tests; implement the native launcher
   with unchanged `HOME`, instance `CODEX_HOME`, and fixed Codex `--cd`.
5. Add failing two-instance and concurrency tests; prove operations on one name
   cannot affect the other.
6. Add failing remove tests for foreign, drifted, symlinked, and partial root or
   launcher state; implement exact launcher-leaf plus root-owned cleanup and
   recovery.
7. Add failing O2 migration tests; implement stage, verify, promote, reverify,
   exact prior-manifest cleanup plan, and interruption recovery.
8. Add credential-leak and exact-path acceptance assertions; then run the
   credential-free lifecycle matrix and separately credentialed live smoke.
9. Promote durable docs, reconcile the implementation against this plan, and
   remove this temporary plan before the implementation PR leaves draft.

## Acceptance Evidence

The implementation PR must retain public-safe, exact-head evidence for:

| Scenario | Observable proof |
|---|---|
| Create one instance | exact root/layout/permissions, supported manifest, external `$HOME/.local/bin/<name>` native launcher, stable receipt |
| Repeat create/verify | read-only canonical result; no metadata/content drift outside documented receipt state |
| Two instances | distinct Codex/runtime/memory/workspace/dependencies; simultaneous launches resolve only their own roots |
| Launcher | exact external location/artifact; real `HOME` unchanged; caller `CODEX_HOME` ignored as authority; child receives instance `CODEX_HOME`; Codex argv contains fixed `--cd` workspace |
| Collision and escape | missing/unsafe `.local/bin`, invalid names, case collisions, symlinks, foreign launcher leaf/files, forged manifest path, and path replacement refuse before mutation |
| Remove one instance | only exact manifest-owned external launcher leaf and selected root paths disappear; launcher-directory/sibling entries, sibling instance, and all canaries remain unchanged |
| Migration success | root and external launcher verified at final paths before exact prior O2 manifest paths are deleted; a prior launcher is replaced only with exact source-manifest proof |
| Migration fault | pre-cleanup failures preserve old state and restore only an exact prior launcher when required; cleanup interruption deletes no unproven path and has deterministic resume/inspect guidance |
| Existing-state preservation | `.local/bin` siblings and content/mode, user Codex home, shell startup files, aliases fixture, unrelated home paths, environment, and source checkout remain unchanged |
| Credential boundary | automated lifecycle completes without credentials and evidence contains no secret value or copied real config |

Current-user native acceptance runs beneath a disposable test root only where
the product contract permits root injection for tests; the shipped command must
still derive its real assistants root from the unchanged current-user `HOME`.
Before/after manifests enumerate the instance root and its sole allowed
external launcher leaf plus public-safe canaries: launcher-directory siblings,
a foreign same-name launcher collision fixture, a sibling instance, and
representative existing Codex and shell files. They prove no second outside
directory entry changed and that `.local/bin` was not created, chmodded, or
otherwise adopted. They record paths, types, permissions, sizes, and fixture
hashes, never real user-file contents. The expected parent-directory metadata
effect of adding or removing its one child entry is recorded separately and
does not authorize any other entry or directory-property mutation.

After the credential-free lifecycle matrix passes, an independent acceptor
configures a dedicated non-production credential through Codex's supported
file-backed mechanism inside one instance `CODEX_HOME` and runs one bounded live
Codex smoke through the generated launcher. Evidence records candidate commit,
platform, instance name, launcher/manifest identity, expected workspace,
success/failure, and redaction checks. It records no credential, token-shaped
output, raw config, private prompt content, or ambient environment dump. A
second uncredentialed instance must remain uncredentialed and unchanged.

This is terminal and filesystem work, not a meaningful graphical interface.
`docs/operations/ui-acceptance.md` should classify it as no rendered UI impact;
plain terminal order, wrapping, error clarity, and non-color-only status remain
part of transcript review.

## Rollout

The approved execution envelope is `implementation` only:

1. Implement on a feature branch from the merged, approved planning head.
2. Keep the implementation PR draft through failing-first tests, native
   acceptance harness work, documentation promotion, and reconciliation.
3. Produce exact-head credential-free and separately credentialed evidence for
   independent review.
4. Stop before merge, any migration against a user's prior projection,
   nomination, release, installation, or activation. Those actions require the
   repository's later acceptance and release authority.

No feature flag or staging service is introduced. Migration remains an
explicit user-invoked lifecycle operation and must not run during installation,
upgrade, validation, or ordinary launch.

## Rollback And Recovery

- Before root promotion, discard only manifest-proven staging paths; the
  external launcher leaf remains untouched.
- If root promotion succeeds but external launcher installation fails, remove
  the new root only when its manifest and artifact identity are still exact;
  otherwise preserve it recovery-required. Never clear a foreign launcher.
- After new-instance promotion and launcher installation but before old cleanup,
  retain both states; the new manifest records recovery status and verification
  outcome. Migration rollback may restore only the exact prior launcher
  artifact proven by the source manifest.
- After cleanup begins, never attempt a broad reverse copy. Preserve the new
  active instance and any remaining old manifest-proven paths, then resume or
  inspect the same exact cleanup plan.
- Removing a newly created instance is allowed only through its valid manifest:
  verify and remove the exact external launcher leaf, then root-owned paths. It
  must not alter `.local/bin` itself, siblings, user-global Codex, or shell state.
- No rollback deletes or rewrites an unmanifested path. Manual recovery guidance
  may identify a path and invariant but must not recommend broad recursive
  deletion.

## Release Prerequisites

- Independent review of path canonicalization, symlink/case behavior, external
  launcher derivation/collision and sole-projection authority, manifest schema
  and deletion authority, launcher argv/environment construction, credential
  redaction, and migration interruption recovery.
- Native macOS current-user acceptance of exact containment and coexistence.
- A dedicated credential and safe external endpoint for the separate live
  smoke, provisioned outside My Friday and excluded from retained evidence.
- Reconciliation must promote shipped contracts to durable architecture,
  security, development, deployment/runbook, and user documentation and delete
  this temporary plan.
- Merge, migration execution, nomination, acceptance, and release require
  separate authority beyond this `implementation` envelope.

## Production Readiness Preflight

Not applicable to this planning task because the `implementation` execution
envelope cannot merge, deploy, activate, migrate a user's state, or release a
production candidate. Implementation must nevertheless define non-secret
credential slots for the isolated live smoke, executable verification commands,
bounded rollback/recovery behavior, and exact-head evidence receipts so a later
release decision has usable inputs.
