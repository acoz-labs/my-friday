# Verification And Release Design

## Test Strategy

### Schema and semantic contract

Add table-driven tests near a new `internal/memory` capability for every schema:
valid canonical bytes; missing/extra fields; wrong constants/types; malformed
IDs/timestamps; assistant mismatch; Unicode normalization; control/format/line
characters; grapheme boundaries; text limits; owner/mode/type violations; and
copied-schema byte drift. Validate filename/ID equality, unique/sorted sources,
source existence/type, proposal sensitivity floor, exact durable copy,
monotonic promotion sensitivity, explicit recorder attribution, unique
transaction IDs, completion-receipt cross-links, and one memory per proposal.

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

Prove uninitialized `observe|journal|propose|promote|recall` refuse read-only and
name `memory validate --initialize`; no command combines `Initialize` with
`Record`. Initialization transcripts show four exact placeholder removals, the
Git working-tree consequence, exact `Initialize`, and all safe exits.

### Transactions, recovery, and concurrency

Inject checkpoints before/after journal fsync, stage create/write/fsync, phase
update, no-replace rename, directory fsync, final validation, completion-receipt
fsync, and journal removal. Run the complete 3×3 stage/final `A|E|X` matrix at
every journal phase, explicitly including `stage absent + exact final + phase
staged` (rename before phase update), exact duplicate stage/final, advanced
phase with missing bytes, and conflicting/missing receipts. Assert the truth
table's sole action, exact `Recover` gate and safe exits, or retained refusal.
Repeat after cleanup for committed and aborted receipts and prove transaction/
record linkage.

Exercise root inode replacement before preview and every mutation, symlinked
ancestors/entries, hard links, case-fold collisions, device/mount change,
FIFO/device/socket, wrong owner/mode, descriptor/path disagreement, ambient
absolute-path substitution, collision/digest drift, malformed journal, cross-
root journal, disk-full, lock contention, concurrency, and repeated recovery.
Instrument filesystem calls to assert every read, enumeration, rename, and
unlink is pinned-descriptor-relative and no-follow. Foreign bytes are never
opened through a followed link, overwritten, or deleted.

Initialization fault tests cover the four data directories independently as
empty or exact empty-mode-`0600` `.gitkeep`, mixed exact/absent placeholders,
changed bytes/mode/owner/link/type, extra entries, and interruption before/after
each schema addition, placeholder deletion, `memory-contract.json`, receipt,
and cleanup. Only exact generator placeholders are removed; after any visible
effect recovery completes the manifest or refuses on drift.

### Recall

Golden tests cover NFC/case normalization, Unicode letter/number tokenization,
distinct-token scoring, zero matches, stable score/time/ID ties, filesystem and
map-order independence, proposals excluded, restricted excluded, sensitive
count-only warning, exact per-run inclusion, consent reset, five-entry cap,
4,000-grapheme cap including envelope, whole-entry skipping, oversized entries,
explicit recorder, fixed match reason, sorted matched tokens, those fields
inside the packet cap, and deterministic bytes. Validate no reads outside the
pinned root descriptor and no writes anywhere.

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
2. Add failing placeholder/collision tests; implement separately confirmed
   initialization WAL, exact placeholder migration, and completion receipt.
3. Add failing observation/journal tests; implement the pinned-root writer,
   transaction-linked records, exhaustive recovery table, receipts, and exact
   `Recover` interaction.
4. Add failing proposal reference/sensitivity tests; implement proposal preview
   and immutable write.
5. Add failing interactive-only/idempotent/monotonic promotion tests; implement
   distinct durable records.
6. Add failing whole-repository validation and stable-diagnostic matrices;
   implement `memory validate` and exhaustive reference checks.
7. Add failing lexical ranking, consent, bound, and golden-packet tests;
   implement recall with recorder, reason, and matched tokens inside the cap.
8. Add PTY acceptance transcripts, child/network/filesystem negative-effect
   probes, race/concurrency/fault suites, docs, and release-chain evidence.
9. Reconcile the exact implementation head, promote durable documentation,
   delete this temporary plan, and verify reconciliation before review-ready.

## Acceptance Evidence

This change has **no rendered UI impact** under
`docs/operations/ui-acceptance.md`; terminal transcripts are behavioral evidence,
not screenshots. The implementation PR must retain sanitized, openable,
exact-head evidence for:

1. Initialize an existing generated memory repository; prove separate exact
   `Initialize`, four-placeholder migration, safe exits, and no Git invocation.
2. Record one observation and one chronological journal entry.
3. Propose a claim citing both, then inspect files and attribution.
4. Attempt noninteractive and wrong-case promotion (denied), then exact
   interactive `Promote` (one durable record).
5. Recall in task A: prove proposal and restricted memory exclusion, standard
   inclusion, sensitive default exclusion and count-only warning.
6. Repeat with exact `Include`; prove no consent survives the next invocation.
7. Demonstrate five-record and 4,000-grapheme bounds, recorder attribution,
   lexical reason, sorted matched tokens, and deterministic output.
8. Paste the exact packet manually into a fresh Codex task B and verify it is
   legible, attributed, explicitly non-authoritative, and sufficient to locate
   source records. Do not require or claim automatic model behavior.
9. Interrupt before/after rename and before phase/receipt/journal updates;
   preview recovery, prove safe exit, exact `Recover`, and then receipt-linked
   read-only retry after journal deletion.
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
2. Verify issue #4/O2 is independently accepted and released and its production
   receipt/tag/asset/digest prove the exact-byte chain. Without that completed
   authority, issue #5 cannot be nominated.
3. Nominate the merge commit and one named Darwin/ARM64 archive through the
   artifact-nomination workflow. Build once; record run/artifact/digest.
4. Independent acceptance downloads and verifies that archive, exercises the
   disposable-user matrix, and records acceptance against the exact candidate.
5. The release workflow re-downloads and digest-verifies the accepted archive,
   tags the accepted commit, publishes that same archive in the GitHub Release,
   records the production receipt, and verifies tag/commit/asset/digest.
6. Close issue #5 only after release verification and lifecycle/project state
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

- Issue #4/O2 must be independently accepted and released. Its production
  receipt must bind accepted commit, tag, GitHub Release asset/digest, workflow/
  run SHA, and acceptance record; issue #5 verifies and reuses that authority
  before nomination. No substitute artifact path is authorized.
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
- **Build/injection:** only after verifying issue #4's accepted/released receipt,
  reuse its pinned workflow to produce one Darwin/ARM64 archive and pass its
  name/digest/run to acceptance and release.
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
