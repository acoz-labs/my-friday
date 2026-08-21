# Verification And Release Design

## Test Strategy

### Pure plan and contract tests

Likely files: `internal/codexhome/plan_test.go`,
`internal/codexhome/manifest_test.go`, and `internal/codexhome/render_test.go`.

- Runtime contract-v1 validation, role and assistant-ID checks.
- Deterministic self-contained rendering for every profile preset, Unicode NFC,
  null optional address, escaping, size bounds, and safety-policy invariant.
- Manifest/journal schema round trips, rejection of extra fields, wrong
  versions, absolute projection paths, duplicate paths, oversized files, and
  mismatched digests.
- State classification for absent, healthy, shadowed, collision, drift,
  interrupted, source drift/missing, incompatible, and unsupported roots.
- Immutable preview actions/non-actions and exact confirmation verbs.

These tests receive explicit temporary `UserHome` and `CodexHome` roots. A test
helper must reject an omitted root; no test fallback may call `os.UserHomeDir`,
read `HOME`, or inherit `CODEX_HOME`.

### Filesystem transaction tests

Likely files: `internal/codexhome/transaction_test.go` and
`internal/codexhome/recovery_test.go`.

- Install, exact rerun, verify, compatible upgrade, rollback, repair, uninstall,
  and complete install/uninstall reversal inside `t.TempDir()`.
- Full fake-Codex-home tree snapshots before and after every operation, with
  canary config, auth-shaped placeholder, session/log/package/skill paths,
  hidden files, extended attributes where supported, modes, and symlinks. Tests
  assert unrelated entries are byte/mode/link-identical without ever modeling
  real credential values.
- Collision matrix for `AGENTS.md`, non-empty `AGENTS.override.md`,
  `.my-friday` file/directory/symlink, stage/journal names, case-folded names,
  hard links, FIFO/device/socket, and symlinked ancestors.
- Drift matrix for active projection, manifests, generations, journal, runtime
  source, source assistant identity, filesystem device/inode, and ownership.
- Fault injection before and after journal, stage, projection promotion,
  manifest promotion, verification, previous-generation rotation, deletion,
  and cleanup. Each checkpoint proves clean rollback, deterministic recovery,
  or safe refusal with retained evidence.
- Concurrent plan/lock tests and revalidation races, including target creation
  after preview and root replacement before commit.
- Recovery idempotency after every published phase and source disappearance
  after a complete stage.
- Owner-only modes and durable write ordering on Darwin.

No destructive transaction test may receive `/Users/batcomputer/.codex`, the
current user's actual home, a root from `os.UserHomeDir`, or any path under the
deployed `batcomputer-ai` projection. A guard shared by all mutation entrypoints
must fail tests if the explicit fixture-boundary token is absent in test mode.

### CLI integration and evidence tests

Likely files: `cmd/my-friday/main_test.go`,
`internal/terminal/codex_wizard_test.go`, and
`internal/terminal/codex_evidence_test.go`.

- Command grammar, stable exit classes, Return/EOF/back/quit behavior, exact
  confirmation case, progress, recovery command, and sanitized receipts.
- Subprocess tests with temporary `HOME` and pre-created `CODEX_HOME`, never the
  parent process environment. Before sending a confirmation verb, the harness
  asserts the previewed canonical target equals the fixture root.
- Native Darwin/APFS checks for non-root UID, ownership, symlink denial, atomic
  rename, fsync ordering, and Git/Codex version reporting.
- Generated sanitized transcripts for install, collision, interruption/recovery,
  repair, upgrade/rollback, drift denial, and complete uninstall reversal.

### Repository checks

Run the normal `bin/container bin/ci`. Native `bin/ci` on supported Apple
silicon remains mandatory because the container cannot prove APFS, Darwin
ownership, terminal behavior, or installed-state recovery. Run focused
`go test -race ./...` and static Darwin/ARM64 build as already defined in
`docs/development.md`.

## Red/Green Sequence

1. Add failing renderer/manifest tests for the one-projection contract, then
   implement pure values and schemas.
2. Add failing environment/root-boundary tests, including absent injection and
   live-home sentinel rejection, then implement pinned root capabilities.
3. Add failing read-only state and preview tests, then implement classification
   and immutable plans.
4. Add failing first-install transaction and complete reversal tests, then
   implement journal/stage/promotion/verification.
5. Publish the full fault matrix as failing tests, then implement conservative
   recovery until every phase is idempotent.
6. Add failing drift/collision/race/concurrency tests, then harden
   file-descriptor-relative operations and revalidation.
7. Add failing repair/upgrade/rollback tests, then implement one-generation
   history without widening ownership.
8. Add failing CLI transcript/evidence tests, then expose the namespaced command
   and stable operator output.
9. Run full container and native suites, then perform exact-candidate acceptance
   under the isolated identity boundary.

## Acceptance Evidence

This is a terminal/filesystem change, not a meaningful rendered browser/native
UI change under `docs/operations/ui-acceptance.md`; screenshot or visual
baseline evidence is not required. Exact terminal transcripts and filesystem
manifests are required.

This envelope stops before candidate acceptance. A later delivery plan must
make the following evidence executable against one packaged Darwin/ARM64 binary
identified by full commit and SHA-256 digest:

- macOS version, Apple silicon architecture, local APFS proof, Git version,
  Codex CLI version, non-admin UID evidence, disposable home identifier in
  sanitized form, candidate SHA, and artifact digest;
- before/after sanitized tree manifests for the disposable Codex home,
  including unrelated canaries but no credential contents;
- install preview/cancel, install/verify, a real Codex discovery smoke using a
  sanitized test profile and separately authorized test account, collision
  denial, production-candidate-safe interruption (for example terminating the
  unmodified process at an externally observed published phase), recovery,
  source upgrade, rollback, drift/repair,
  uninstall, repeated uninstall/verify, and final tree equivalence;
- proof that `/Users/batcomputer/.codex`, the accepting operator's real home,
  and deployed `batcomputer-ai` projections were outside the candidate process
  root and unchanged; and
- nomination, CI, acceptance workflow, evidence artifact, decision, release
  workflow, Git tag, and GitHub Release links.

The preferred boundary is a newly created disposable non-admin macOS user with
a fresh home and keychain. A macOS VM or dedicated APFS test volume is equally
strong only when it also supplies a distinct UID, home, keychain/login state,
and a verified route preventing writes to the host user's Codex home. A plain
temporary directory under Alfred's account is insufficient for acceptance.

## Rollout

1. Complete code, tests, docs, and a reviewed implementation PR.
2. Stop at the `implementation` envelope. Merge, nomination, acceptance, and
   release are not authorized by this plan.
3. A later plan must name executable packaging/upload/download/digest commands,
   the macOS host, disposable-user create/delete commands, fresh home/keychain
   checks, `CODEX_TEST_ACCESS_TOKEN` secret slot and approved injection source,
   sanitized evidence collection, and teardown verification.

There is no feature flag or background activation. Installing the released
artifact does not mutate Codex; the user still invokes the explicit lifecycle.

## Rollback And Recovery

- Before any later release, reject the candidate or revert its implementation.
- A user can run `my-friday codex rollback` for the prior managed generation or
  `uninstall` for complete reversal. Drift or ambiguous state fails closed and
  retains the transaction journal for diagnosis.
- Source rollback does not delete installed state. A compatible older runtime
  may be supplied through explicit `upgrade` only when the assistant identity
  and renderer contract permit it.
- A later release runbook must remove the injected test credential, delete the
  disposable user/home/keychain, and verify no residue after evidence capture.

## Release Prerequisites

- A later exact candidate passes container CI and native Apple silicon/APFS/Git tests.
- All mutation tests prove explicit injected roots and the live-home guard.
- Independent acceptance evidence opens successfully and is bound to the
  candidate SHA and artifact digest.
- `docs/architecture.md`, a new installed-baseline capability document,
  `docs/development.md`, `docs/deployment.md`, and `docs/runbook.md` describe
  shipped behavior and correct stale release text.
- Executable packaging/transport/upload and macOS acceptance automation are
  designed and verified before a broader envelope is requested.

## Production Readiness Preflight

Not applicable to the `implementation` envelope. Runtime installation uses no
secret, but production readiness is not proved: current automation lacks a
digest-verified executable package/transport/upload chain and a macOS harness
for disposable UID/home/keychain provisioning, named test-credential injection,
evidence, and teardown. A broader plan must provide executable deploy or
publication, activation, verification, rollback, and exact-candidate receipt
contracts rather than treating an opaque artifact string as the bytes.
