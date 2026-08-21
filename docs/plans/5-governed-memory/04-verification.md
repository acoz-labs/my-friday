# Verification And Release Design

## Test Strategy

### Schema and semantic contract

Add table-driven tests near a new `internal/memory` capability for every schema:
valid canonical bytes; missing/extra fields; wrong constants/types; malformed
IDs/timestamps; assistant mismatch; Unicode normalization; control/format/line
characters; grapheme boundaries; text limits; owner/mode/type violations; and
copied-schema byte drift. Validate filename/ID equality, unique/sorted sources,
source existence/type, proposal sensitivity floor, exact durable copy,
monotonic promotion sensitivity, and one memory per proposal.

Use generated standard/sensitive/restricted fixture text only. Credential-shaped
canaries remain test-generated and assertions prove they do not enter errors,
transaction journals, recall packets without consent, or durable evidence.

### Commands and interaction

Extend `cmd/my-friday/main_test.go` and terminal PTY/transcript tests for exact
usage, explicit path, field prompts, normalized preview, exact `Initialize`,
`Record`, `Propose`, `Promote`, and `Include`, plus Return/EOF/`q`/wrong case/
leading-trailing whitespace safe exits. Prove promotion refuses pipes and every
attempted `--yes`/environment bypass. Success output includes IDs but not bodies;
rejected values never echo.

### Transactions, recovery, and concurrency

Inject filesystem checkpoints before/after journal fsync, stage create/write/
fsync, phase update, no-replace rename, directory fsync, final validation, and
journal removal for each operation. At every point assert either the exact old
valid repository, exact new valid repository, or a retained recoverable journal.
Exercise root inode change, symlinks, devices, wrong owners/modes, stage/final
collision, digest drift, truncated journal, unknown phase/field, cross-root path,
disk-full/permission errors, lock contention, concurrent promotion of the same
proposal, and repeated recovery. Foreign bytes are never overwritten or deleted.

### Recall

Golden tests cover NFC/case normalization, Unicode letter/number tokenization,
distinct-token scoring, zero matches, stable score/time/ID ties, filesystem and
map-order independence, proposals excluded, restricted excluded, sensitive
count-only warning, exact per-run inclusion, consent reset, five-entry cap,
4,000-grapheme cap including envelope, whole-entry skipping, oversized entries,
and deterministic bytes across repeated runs. Validate no reads outside the
explicit root and no writes anywhere.

A performance characterization scans at least 10,000 small valid memory files
on the supported filesystem and records time/memory without creating a pass/fail
product promise; a pathological result reopens the no-index decision before
release.

### Boundary and regression suite

- Existing bootstrap init/validate/recovery tests remain green for uninitialized
  contract-v1 memory repositories.
- `memory validate` distinguishes uninitialized, valid governed, invalid, and
  recovery-pending repositories without changing pair-level `validate` behavior.
- A process observer and before/after filesystem manifest prove no `git`,
  network client, Codex executable, shell, or unrelated child command is invoked.
- Run a loopback listener plus platform network observer/deny harness with a
  positive control where available; candidate memory commands produce zero
  connection attempts. Local-only code also receives static review for network
  imports.
- Run `go test -race ./...`, formatting, vet/static checks, and the complete CI
  entrypoint.

## Red/Green Sequence

1. Add failing closed-schema/canonicalization tests and embedded v1 schemas;
   implement typed records and semantic validation.
2. Add failing initialization/collision tests; implement exact previewed
   governed-contract initialization without changing existing user records.
3. Add failing observation and journal prompt/confirmation tests; implement the
   shared journaled single-file writer and recovery checkpoints.
4. Add failing proposal reference/sensitivity tests; implement proposal preview
   and immutable write.
5. Add failing interactive-only/idempotent/monotonic promotion tests; implement
   distinct durable records.
6. Add failing whole-repository validation and stable-diagnostic matrices;
   implement `memory validate` and exhaustive reference checks.
7. Add failing lexical ranking, consent, bound, and golden-packet tests;
   implement read-only deterministic recall.
8. Add PTY acceptance transcripts, child/network/filesystem negative-effect
   probes, race/concurrency/fault suites, docs, and release-chain evidence.
9. Reconcile the exact implementation head, promote durable documentation,
   delete this temporary plan, and verify reconciliation before review-ready.

## Acceptance Evidence

This change has **no rendered UI impact** under
`docs/operations/ui-acceptance.md`; terminal transcripts are behavioral evidence,
not screenshots. The implementation PR must retain sanitized, openable,
exact-head evidence for:

1. Initialize an existing generated memory repository.
2. Record one observation and one chronological journal entry.
3. Propose a claim citing both, then inspect files and attribution.
4. Attempt noninteractive and wrong-case promotion (denied), then exact
   interactive `Promote` (one durable record).
5. Recall in task A: prove proposal and restricted memory exclusion, standard
   inclusion, sensitive default exclusion and count-only warning.
6. Repeat with exact `Include`; prove no consent survives the next invocation.
7. Demonstrate five-record and 4,000-grapheme bounds plus deterministic output.
8. Paste the exact packet manually into a fresh Codex task B and verify it is
   legible, attributed, explicitly non-authoritative, and sufficient to locate
   source records. Do not require or claim automatic model behavior.
9. Interrupt a write after stage/promotion, run explicit recovery, and prove an
   exact valid final repository with no transaction residue.
10. Run validation, process/network observer, and before/after manifests proving
    no Git, network, global Codex, sibling repository, or unrelated path effect.

Contributor evidence is development evidence only. Independent acceptance
repeats the matrix from the exact nominated Darwin/ARM64 archive under a fresh
disposable non-admin macOS user/home, using generated fixtures and a temporary
memory root. The acceptance record binds issue #5, all lifecycle-linked merged
implementation PRs, merge commit, artifact name, SHA-256, nomination run, OS/
architecture, transcript manifest, and acceptor identity. The implementer is
not the sole acceptor.

## Rollout

1. Merge only after approved-plan reconciliation, independent code/security
   review, all required checks, durable docs promotion, and temporary-plan
   deletion.
2. Nominate the merge commit and one named Darwin/ARM64 archive through the
   artifact-nomination workflow. Build once; record run/artifact/digest.
3. Independent acceptance downloads and verifies that archive, exercises the
   disposable-user matrix, and records acceptance against the exact candidate.
4. The release workflow re-downloads and digest-verifies the accepted archive,
   tags the accepted commit, publishes that same archive in the GitHub Release,
   records the production receipt, and verifies tag/commit/asset/digest.
5. Close issue #5 only after release verification and lifecycle/project state
   are updated. Memory capability has no service activation, migration daemon,
   secret slot, feature flag, or staging environment.

## Rollback And Recovery

- **Interrupted record write:** use `memory recover` with the reported journal.
  It proves and completes/removes only transaction-owned state; ambiguity is
  preserved for diagnosis.
- **Bad unreleased implementation:** do not nominate; correct through a reviewed
  PR. No user data migration exists before release.
- **Bad released binary:** withdraw/mark the release, restore the previous
  immutable release artifact/tag as the recommended version, and ship a new
  forward-fix candidate. Never rebuild under an existing tag or overwrite an
  accepted asset.
- **Existing memory repositories:** immutable v1 records remain ordinary files.
  Rolling back to a binary that predates O3 must not mutate them; it may reject
  the added owned schema/control paths as unsupported. Users retain the newer
  binary for validation/recovery until pending journals are cleared.
- **Schema defects:** v1 schema bytes are immutable after release. Corrections
  requiring different accepted documents use a new schema/record version and a
  separately designed explicit migration; never silently reinterpret v1.
- **User-authored corruption:** validation reports it. V1 provides no automatic
  repair, deletion, redaction, or rollback of user records.

## Release Prerequisites

- Issue #4's exact-byte nomination/acceptance/release chain must be merged and
  operational, or equivalent enabling work must land before nomination.
- Disposable non-admin acceptance tooling must prove unique marker, UID, home,
  ownership, and exact teardown; a mismatch refuses deletion. Physical admin
  authentication may be required only to create/remove that test identity and
  never reaches My Friday.
- A product-design reviewer must inspect the exact-head prompt/transcript matrix,
  especially sensitivity warnings, safe exits, attribution, and packet wording.
- Security review must inspect content leakage, symlink/path handling, journal
  authority, monotonic sensitivity, noninteractive promotion denial, and no
  external effects.

These are implementation/release prerequisites, not unresolved product
decisions.

## Production Readiness Preflight

- **Secrets:** none required. Candidate fixtures contain no real private data or
  credentials; no secret slot or workflow credential is added.
- **Build/injection:** reuse the repository's pinned release workflow extended
  by issue #4 to produce one Darwin/ARM64 archive and pass its name/digest/run to
  acceptance and release.
- **Deploy/activate:** artifact repository; no service deploy or feature toggle.
  Production is the verified Git tag and GitHub Release carrying the accepted
  archive.
- **Verify:** download release asset, compare SHA-256 to nomination/acceptance,
  verify binary version/commit, run isolated smoke `memory validate` and recall
  against generated fixtures, and inspect the production receipt.
- **Rollback:** preserve prior release/tag/assets and recommend the last verified
  version; use a new version for forward fix. Never replace immutable bytes.
- **Receipt:** record accepted commit, tag, release URL/ID, asset name/ID,
  SHA-256, nomination and acceptance record IDs, workflow/run SHA, issue/PR set,
  verifier, and timestamp.
