# Verification And Release Design

## Test Strategy

Testing follows TDD at the lowest useful layer. No test points at a real home,
Codex configuration, or unrelated repository. Filesystem integration tests use
explicit temporary APFS roots and assert the complete before/after tree.

### Unit and contract tests

- `internal/profile/*_test.go`: Unicode lengths, trimming, style/custom rules,
  control/format rejection, deterministic assistant ID.
- `internal/plan/*_test.go`: canonical serialization, ordered actions, content
  digests, negative actions, stable plan ID, slug fallback, preview model.
- `internal/contract/*_test.go`: schema compilation, conformance corpus, schema
  copy digests, unknown-version rejection, no profile interpolation.
- Codex projection tests: root `AGENTS.md` references neutral contracts,
  contains no user-value interpolation, and no generic adapter or alternate
  harness ships.
- `internal/paths/*_test.go`: root/home, absent descendants, same/nested targets,
  symlink ancestors/targets, Unicode paths, empty/non-empty targets, injected
  filesystem/environment results.
- `internal/terminal/*_test.go`: line transcripts, default Exit, exact Create,
  answer preservation, no ANSI/cursor control, stable English messages.

### Integration and acceptance tests

- `internal/repository/*_test.go`: render exact trees, initialize Git through
  fake/real runners, validate roles/fresh state, prove no commits/remotes/templates.
- `internal/transaction/*_test.go`: absent and empty targets, modes, journals,
  reservations, concurrency, promotion, rollback, and recovery.
- `test/acceptance/init_test.go`: drive Scope through Result, retain transcript,
  validate both repos, compare every mutation with `CreationPlan`.
- `test/acceptance/fault_matrix_test.go`: fail immediately before/after every
  journal/reservation/stage/init/validate/backup/promote/final-validate/cleanup
  transition; assert both-valid, pre-run, or recoverable, then recover twice.
- `test/acceptance/zero_network_test.go`: source import allowlist, injected
  runner rejecting non-allowlisted Git operations, captured argv/environment,
  schema compiler rejecting `http`, `https`, `file`, and unknown resource
  schemes, empty remotes, and full temp-root diff.
- `test/acceptance/unsupported_environment_test.go`: macOS version, architecture,
  Git version, filesystem, non-TTY, and permission failures stop before journal.

### Security and privacy cases

- ESC, NUL, newlines, bidi/format controls, Markdown headings, JSON delimiters,
  shell syntax, long combining sequences, and invalid UTF-8.
- Symlink target/parent; ancestor replacement; target creation/population after
  preview; case-insensitive equivalent APFS paths; nesting both directions.
- Global Git template with sentinel private file/hook; generated `.git` must
  not contain it because initialization supplies an empty template.
- Journal/stderr scans proving purpose, address, and custom guidance are absent.
- Cleanup refusal after manifest, transient creation marker, or baseline digest
  change.
- Owner-only modes under permissive caller umask and empty-shell mode restore.

## Red/Green Sequence

1. Environment/profile tests, then typed inputs/errors.
2. Schema conformance and starter-tree tests, then embedded contracts/templates.
3. Canonical ID/action/hash tests, then pure planner.
4. Complete preview and zero-write Exit transcript, then terminal adapter.
5. Staged Git pair and empty-template isolation, then runner/renderer/validator.
6. Every fault transition for absent targets, then transaction/rollback.
7. Empty-shell modes, crash states, idempotent recovery, changed-target refusal.
8. End-to-end Unicode success/Exit/collision/failure/recovery command wiring.
9. Zero-network/privacy allowlists and output scans.
10. Full formatting, vet, race/unit/integration/acceptance, ARM64 build, and docs
    reconciliation.

## Acceptance Evidence

### Deterministic automated evidence

The implementation PR retains:

- passing unit/race/integration/acceptance checks;
- schema corpus and generated-tree golden fixtures;
- fault matrix naming every transition/result;
- zero-network/subprocess trace and Git-template sentinel result;
- a `darwin/arm64` source/build check;
- `bin/ci` extended for format, vet, tests, and build;
- plan validation until reconciliation removes this directory.

Expected initial commands, finalized in `docs/development.md`:

```sh
bin/container bin/ci
go test ./...
go test -race ./...
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build ./cmd/my-friday
```

Container validation cannot prove APFS/macOS behavior. A native Apple Silicon
macOS job or reviewer run is mandatory for path, permission, Git-template,
transaction, and terminal acceptance. Document any container deviation.

### Terminal experience evidence

This is a new material terminal experience. The browser/native screenshot
matrix in `docs/operations/ui-acceptance.md` does not directly apply, but its
exact-head, scenario, accessibility, and independent-review principles do.
Attach or commit machine-generated plain-text transcripts for:

1. default Exit at Preview;
2. successful combined-parent creation with Unicode profile data;
3. separate-path symlink/nesting collision;
4. injected failure with successful rollback;
5. partial promotion followed by recovery.

Transcripts bind to exact PR head, consistently redact only temp-root prefixes,
contain no ANSI cursor sequences, and note terminal dimensions only as context.
Product-design review checks vocabulary, answer preservation, default safety,
result/recovery comprehension, VoiceOver-readable order, keyboard-only flow,
and no color/motion dependence. First-customer acceptance later repeats happy
path and safe Exit on the immutable candidate; contributor self-check is not
product acceptance.

| Acceptance group | Required evidence |
|---|---|
| Deterministic preview | repeated-plan golden, action/file/hash comparison, zero-write Exit transcript |
| Valid repositories | schema corpus, exact trees, matching IDs, branch/no-commit/no-remote assertions |
| No adjacent effects | temp-root diff, imports, subprocess trace, template sentinel, privacy scan |
| Safe paths | canonical/symlink/home/root/nesting/case-fold/emptiness tests on APFS |
| Recoverability | full fault matrix, two-pass recovery, foreign-change refusal |
| Accessible terminal | transcripts, control-sequence scan, keyboard/VoiceOver review |

## Rollout

Within `implementation`:

1. Implement with no feature flag; no prior runtime exists.
2. Exercise isolated unit/integration tests.
3. Run native Apple Silicon macOS acceptance against exact implementation head.
4. Reconcile, promote docs, remove this plan, and obtain independent engineering
   and product-design review.
5. Stop at a reviewed implementation PR. Merge, nomination, acceptance, and
   release require separate authority.

There is no database/config migration, server, staging service, secret
injection, activation, or user-data backfill.

## Rollback And Recovery

Tests use marked temporary roots; failed tests report retained fixtures and
cleanup removes only marker-owned content. User creation restores absent
targets to absence and empty shells to original mode/state. It never removes a
foreign-modified target without ownership/digest proof. The retained journal
authorizes idempotent `recover`.

No previous application version exists. Reverting implementation removes the
command from source but never deletes user-owned generated repos. Once an
artifact is released, v1 validation must remain available; release is not
authorized here.

## Release Prerequisites

Non-blocking for implementation, blocking before public artifact release:

1. Naming clearance for “My Friday.”
2. Exact first-customer macOS version, Apple Silicon architecture, APFS, and Git
   preflight plus native acceptance evidence. This does not assume an OS version
   the product owner has not supplied.
3. Durable artifact delivery profile, immutable binary digest, build provenance,
   signing/notarization/distribution decision, and rollback procedure.
4. Configured nomination/acceptance/release variables and independent acceptor.
5. Dependency review and license notice for JSON Schema validator.
6. Verified minimum-macOS build target and an ARM64 candidate from the approved
   release path, not a contributor-local binary.

Future harness support is not an O1 release prerequisite. It requires a new
product decision and capability map rather than a speculative compatibility
claim.

## Production Readiness Preflight

Not applicable. The `implementation` envelope cannot merge, publish, stage, or
promote an artifact and O1 has no production service. There are no secret slots,
deploy commands, activation, production verification, or production receipt.
