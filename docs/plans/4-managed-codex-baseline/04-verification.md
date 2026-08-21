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

Add shell/fixture tests for deterministic archive creation, parseable candidate
IDs, artifact download/digest mismatch, acceptance evidence binding, secret
redaction, idempotent same-digest Release asset reuse, mismatched-asset refusal,
and receipt completeness. The macOS harness has dry preflight tests for UID/home/
keychain/CODEX_HOME isolation and teardown plus one native end-to-end run; it
must refuse when admin provisioning, secret injection, or protected-home guards
are not satisfied.

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

Implementation makes the following evidence executable against one packaged Darwin/ARM64 binary
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

1. Nomination builds once from the exact CI SHA, creates a deterministic
   Darwin/ARM64 archive, computes SHA-256, uploads a named Actions artifact, and
   records a parseable `run-id/artifact-name/digest` candidate ID.
2. A local macOS harness downloads by run/name and verifies digest. With one-time
   physical admin authentication its runbook creates a dedicated non-admin user
   with fresh home/keychain/CODEX_HOME, installs supported Codex, and injects
   test-only `OPENAI_API_KEY` from an operator-approved secret source into the
   smoke process only, tracing disabled and without persistence.
3. It runs the matrix, externally terminates the unmodified candidate at an
   observed durable phase, captures sanitized evidence, verifies protected live
   homes, clears the credential, removes user/home/keychain after physical admin
   authentication, and proves teardown.
4. Acceptance records the exact ID/digest/evidence. Release downloads the same
   artifact, re-verifies digest and authority, and uploads those exact bytes to
   GitHub Release. Mismatches fail closed; an existing asset is reused only when
   its digest matches.

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

- **Secrets:** runtime install uses none. Acceptance consumes only the named
  `OPENAI_API_KEY` slot from an operator-approved source, process-scoped with
  tracing disabled; no value enters logs, files, evidence, or My Friday.
- **Deploy/promote:** nomination uploads one named archive and records run ID,
  artifact ID/name, digest, commit, and issue. Release downloads that exact
  artifact and verifies all fields before uploading the same bytes.
- **Activation:** the GitHub Release asset is activation; user installation
  remains explicit and automation never writes an operator Codex home.
- **Verification:** CI verifies packaging; macOS acceptance verifies digest,
  lifecycle, Codex discovery, isolation, and teardown; release verifies accepted
  authority and published-asset digest.
- **Rollback:** failed acceptance publishes nothing. Mismatch fails closed. A
  correction is a new immutable artifact/release; assets are never overwritten.
- **Receipt:** commit, run ID, Actions artifact ID/name, digest, accepted issue,
  evidence and acceptor, release/tag/asset URL and digest, control workflow SHA,
  teardown proof, and rollback target.
