# Verification And Release Design

## Test Strategy

### Pure address and repository-contract tests

Likely files: `internal/remoteaddress/address_test.go` and additions to
`internal/repository/repository_test.go`.

- Table/property/fuzz tests accept representative HTTPS, `ssh://`, and
  SCP-style fixture addresses without normalization.
- Reject empty/oversized/non-ASCII input, every control/whitespace/format/bidi
  class, NUL, percent encoding, query, fragment, userinfo/password, unsafe
  username/host/port/path, plaintext schemes, local/file/UNC/drive paths,
  bundles, unknown schemes, `ext::`, and arbitrary helpers.
- Include leading-option usernames/hosts/paths, malformed DNS labels, numeric-IP
  lookalikes, empty/`.`/`..` segments, SCP ambiguity, and SSH option-injection
  shapes such as a generated `-oProxyCommand` token.
- Generate credential-shaped values only in test memory, assert they never
  reach Git/config/output/evidence, and discard them without printing, hashing,
  snapshotting, or committing a literal sentinel.
- Inspect runtime and memory independently; accept evolved commits, branches,
  content, and other remotes; reject tampered/unsupported contracts, separate or
  bare Git dirs, unsafe `.git`, and changed role/identity.
- Exercise symlinked, spaced, and Unicode entered repository paths while
  preserving a distinct canonical path.

### Git adapter and configuration tests

Likely files: `internal/remoteconfig/config_test.go` and
`internal/gitexec/gitexec_test.go`.

- Start with direct-local config snapshots for absent/canonical/collision,
  duplicate URL/fetch, URL-only/fetch-only, push URL, additional origin keys,
  other remotes, invalid syntax, include/includeIf, unreadable config, unsafe
  remote names, symlink/config-type changes, and global/system canaries.
- Add native Git 2.28/current fixtures containing one empty, one comment-only,
  and duplicate-empty `[remote "origin"]` section. Each is key-semantic absence:
  the preview proceeds, Git creates the canonical pair, exact read-back passes,
  and pre-existing comments/adjacent fixture bytes remain. Duplicate or partial
  URL/fetch keys remain collisions. No production raw-config parser is allowed.
- Prove exact literal argv and environment for every Git call. The observer must
  see only `rev-parse`, `config --local --no-includes`, and one
  `remote add -- origin` operation; no shell, credential, helper, `ls-remote`,
  fetch, push, commit, global/system, or provider command may appear.
- Run the adapter contract suite against Git 2.28 with an owner-only empty
  HOME/XDG isolation root and again against native Git. Put hostile global/XDG/
  system canaries in the invoking test environment and prove they are unread,
  unchanged, and unable to influence origin classification.
- Exact success adds one URL and one canonical fetch refspec while preserving
  every non-origin fixture byte and repository content/ref/index/object/hook
  snapshot. Production inspection requests only adjacent key names; generated
  credential-shaped adjacent values never cross subprocess output or enter
  errors/evidence. System/global canaries remain unread and byte-identical.
- Existing exact state performs no write and preserves config identity/size/
  mtime and the selected-origin state.
- Add local/global `insteadOf` and `pushInsteadOf` canaries. Prove the stored
  value remains exact, no expanded endpoint is queried, and preview/receipt
  explicitly state that future resolved endpoints are unverified.
- Existing different/partial/duplicated key-bearing origin refuses and
  preserves raw config.
- Hold `.git/config.lock` to prove a stable nonzero failure and that My Friday
  does not delete/wait on it; repeat after external release succeeds.
- Make config unwritable, corrupt it, inject Git failures, and replace config or
  repository state at pre/post-confirmation checkpoints. Prove no chmod,
  elevation, speculative cleanup, false success, or unsafe output.
- Fault the process before Git, during externally observed config update, and
  after Git success/before read-back. Rerun must resolve absent, canonical, or
  collision without a product journal.
- Run race tests around concurrent attach attempts. At most one canonical
  attachment succeeds; the other reports race/collision/already state.

### Terminal, privacy, and evidence tests

Likely files: `internal/terminal/remote_wizard_test.go`,
`internal/terminal/remote_evidence_test.go`,
`internal/terminal/production_boundary_test.go`, and
`cmd/my-friday/main_test.go`.

- Exact command grammar; help disclosure; role-first preview; complete runtime
  and memory copy; exact `Attach`; Return/EOF/`q`/`attach`/`Attach ` exits; stable
  error classes; success/already/collision/verification-pending receipts.
- Before accepting confirmation, integration harnesses assert the displayed
  canonical repository equals the explicit temporary fixture. Production must
  refuse noninteractive input; automated scenarios use a PTY.
- Accepted fixture addresses are visible only where expected. Unsafe/rejected
  values are absent from stdout, stderr, observer args, errors, snapshots, test
  names, CI annotations, and generated transcripts.
- Existing unsafe origin is rendered `<redacted>` or a structurally safe
  scheme/host only; no digest or transformed credential material appears.
- Run output at 80 columns and with long spaced/Unicode repository paths. Review
  line order with a terminal screen reader model; assert no ANSI/cursor/spinner.
- Generate sanitized exact-head transcripts and before/after manifests for the
  complete scenario matrix using only `example.invalid` and temporary aliases.

### Repository and release checks

Run focused packages first, then `go test -race ./...`, static Darwin/ARM64
build, `bin/container bin/ci`, and native `bin/ci` on supported Apple silicon.
Container results cannot substitute for native APFS, Git-lock, terminal, and
PF/DTrace network-guard evidence. The process observer remains useful for argv/
environment assertions but cannot prove syscall or packet non-use.

If the exact-byte artifact chain is absent on implementation base, add failing
workflow/script tests before enabling it. Verify deterministic archive name and
digest, upload/download by run and artifact identity, digest-mismatch refusal,
acceptance binding, same-byte GitHub Release upload, idempotent same-digest
reuse, mismatched existing-asset refusal, and complete receipts. If issue #4
already delivered the chain, reuse those tests and add issue #7 candidate cases;
do not fork a second artifact protocol.

## Red/Green Sequence

1. Add failing address allowlist/denylist and fuzz tests; implement the pure
   non-normalizing parser.
2. Add failing single-repository runtime/memory/evolved-state tests; factor
   ordinary contract inspection without weakening pair validation.
3. Add failing direct-local snapshot and fixed-env/argv observer tests; build
   the read-only adapter and origin state classifier.
4. Add failing preview/cancel/exact-confirmation transcripts; implement the
   terminal plan without mutation.
5. Add failing exact add/read-back/delta/idempotency/collision tests; expose the
   one Git mutation.
6. Add failing lock/permission/corruption/fault/TOCTOU/concurrency tests; harden
   revalidation and verification-pending recovery.
7. Add failing privacy, other-repository isolation, no-network/credential/global
   Git, Unicode path, 80-column, and transcript-generation tests. Add the
   PF/DTrace supervisor positive-control, zero-event, teardown, and mismatch
   refusal tests separately from argv-observer tests.
8. Reconcile or implement the exact-byte artifact transport, run full native/
   container checks, then execute independent exact-candidate acceptance.

## Acceptance Evidence

This is a meaningful terminal experience, not a browser/native graphical UI.
Under `docs/operations/ui-acceptance.md`, retained UTF-8 transcripts and state/
subprocess manifests replace screenshots, browser console, network panel, and
visual baselines.

### Child-inclusive no-network acceptance guard

The subprocess observer is not network evidence; it records scrubbed argv and
environment only. Through-production acceptance adds a separate macOS kernel/
syscall guard around the exact artifact:

1. The root supervisor proves the disposable non-admin UID is marker-bound and
   quiescent, records PF enabled/reference state and existing anchors, and
   verifies the host main ruleset already evaluates direct children of the
   `com.apple/*` anchor. It does not reload or flush the main ruleset or add an
   intermediate dispatcher.
2. Before any privileged mutation, the supervisor renders one rule for the
   numeric UID and bounded run label, then requires unprivileged
   `pfctl -vnf -` to parse the exact bytes successfully. The grammar is:
   `block return out log quick all user <disposable-uid> label
   "my-friday-acceptance-<run-id>"`. Parse warnings, normalized output drift,
   unexpected rules, or unsafe label/UID substitution refuse setup.
3. With physical operator authentication, the supervisor obtains a PF enable-
   reference token using `pfctl -E` and loads the already parsed bytes into one
   unique direct-child anchor:
   `com.apple/my-friday-acceptance-<run-id>`. It first proves that exact anchor
   absent, then reads the loaded rule back and proves the direct child is
   attached/evaluated beneath `com.apple/*`. The token is held by the supervisor;
   it is never passed to the candidate.
4. The supervisor starts DTrace as administrator before the candidate. It uses
   global syscall/process providers filtered by the quiescent disposable UID,
   rather than PID-provider probes that could miss a newly exec'd Git child. It
   records process/child lifecycle plus `socket`, `connect`, `sendto`, and
   `sendmsg` entries for every address family. This catches direct IPv4/IPv6 and
   any socket-based resolver IPC without trying to parse arguments or payloads.
   It records event type, executable identity, PID lineage, and monotonic time
   only—never buffers, addresses, packet payload, environment, or config values.
   UID scope includes the candidate's Git child; the non-admin candidate cannot
   change UID.
5. A disposable-UID child positive control attempts one high-entropy `.invalid`
   resolution plus direct reserved-address IPv4 and IPv6 TCP/UDP operations.
   Resolver and direct-network phases use distinct supervisor timestamps.
   Acceptance proceeds only when the resolver phase produces the expected
   socket/IPC syscall class on that macOS build, direct phases produce their
   expected socket classes, labeled PF counters increase, all operations fail,
   and supervisor/tracer health remains good. If the system resolver uses an
   unobserved non-socket path, the positive control fails and no no-network claim
   is permitted. The trace/counters are then reset for the candidate window.
6. The exact candidate runs all scenarios with only high-entropy `.invalid` or
   numeric fixture hosts. Approval requires zero resolver/network events for the
   disposable UID/process tree and zero PF counter delta, alongside the separate
   expected Git argv trace. Any event, counter, tracer loss, PID/UID ambiguity,
   or incomplete window is rejection—not a warning.
7. After the candidate exits, the supervisor proves no descendant or other
   disposable-UID process survives the monitored window. Cleanup then stops and
   verifies DTrace, flushes only the exact marker-bound anchor, releases only the
   recorded PF enable token using `pfctl -X`, and proves the initial PF enabled/
   reference/anchor state is restored. It never uses global `pfctl -d` or
   flushes another ruleset. A process, marker, UID, anchor, token, counter, or
   baseline mismatch blocks acceptance and preserves evidence for operator
   diagnosis rather than broad cleanup.

The admin boundary belongs only to the acceptance supervisor. My Friday and
its Git child always run as the disposable non-admin UID with no privileged
descriptor, token, executable, or environment. Unit tests use a fake supervisor;
one native preflight and exact-candidate run must prove the real PF/DTrace
capabilities on the recorded supported macOS build. If SIP, permissions, PF
layout, or probe availability prevents the positive control, through-production
acceptance cannot claim no-network behavior and the candidate is rejected.

The evidence bundle contains policy/trace-script digests, the exact enabled
DTrace probe manifest for the recorded macOS build, run ID and sanitized UID,
initial/final PF state, enable-reference acquired/released status (the token
value is excluded), dry-run parse transcript/digest, direct-child anchor
attachment/read-back, rule label, positive-control phase/event-class and
counter totals, candidate zero-event/counter totals, tracer start/stop health,
process-tree quiescence, command window timestamps, and exact cleanup verdict.
It contains no packet payload, unrelated host event, private hostname, address,
or credential.

Implementation retains sanitized exact-head evidence for:

| Scenario | Observable proof | Retained evidence |
|---|---|---|
| Runtime + HTTPS fixture | Role, full disclosure, exact add/read-back, receipt | Transcript; before/after config/tree/ref manifest; observer trace |
| Memory + SSH/SCP fixture | Memory disclosure and exact config | Transcript and manifests |
| Return, `q`, EOF, wrong case/space | `No changes made`; byte-identical state | Combined PTY transcript and snapshots |
| Exact repeat | `Already attached`; identity/size/mtime unchanged | Transcript and config metadata |
| Empty/comment-only/duplicate-empty origin sections | Key-semantic absence; canonical add; comments preserved | Native Git 2.28/current fixture transcript and byte comparison |
| Different/partial origin | Collision; original preserved and unsafe value redacted | Transcript and selected-state/metadata proof, never raw value or digest |
| Other remote only | Names disclosed; only `origin` added | Transcript and exact config delta |
| Unsafe/credential/helper/local inputs | Reject before Git; generated value absent everywhere | Boundary assertion and sanitized category counts |
| Invalid/tampered/non-My-Friday repo | Refusal before mutation | Transcript and filesystem snapshot |
| Symlinked/spaced/Unicode path | Entered/canonical mapping; target only | Transcript and target/adjacent manifests |
| Config lock/permission/corruption | Stable failure; no lock deletion/chmod/repair | Transcript and metadata snapshot |
| Change after preview/interruption | Revalidation or verification pending; rerun-safe | Deterministic fault transcript |
| External-effects boundary | Expected local argv; zero child-inclusive resolver/socket attempt; no credential/global/content/ref/other repo effect | Scrubbed argv trace, PF rule counters, DTrace event-class manifest, pair snapshots |
| User Git rewrite rules present | Literal local value verified; later endpoint explicitly unverified | Transcript plus untouched rewrite canaries |
| 80-column/accessibility | Logical order, natural wrap, no ANSI/timing state | Pinned PTY transcript and checklist |

Independent product acceptance freshly reruns both roles, cancellation, repeat,
collision, unsafe-address, lock/interruption, and prohibited-effect scenarios
against the nominated immutable Darwin/ARM64 artifact. The manifest records:

- candidate commit, Actions run/artifact identity, SHA-256 digest, and control
  workflow commit;
- macOS/architecture/APFS/Git versions, locale, terminal width, sanitized
  disposable UID/home identity, and acceptor;
- PF/DTrace policy and script digests, exact enabled-probe manifest, positive-
  control proof, candidate zero-event/counter window, supervisor health, and
  exact cleanup receipt;
- scenario/action/expected/result/verdict rows and openable transcript/manifest
  artifact links;
- proof that Alfred's live Codex home, operator home, source checkout, global
  Git config, and deployed runtime projections were outside candidate roots and
  unchanged; and
- disposable identity marker/teardown proof and final approval or rejection.

No real remote, network service, credential, private path, or provider account
is needed. Fixture addresses are configuration strings only. The subprocess
observer proves expected Git command selection; the PF/DTrace guard separately
proves child-inclusive network denial and zero candidate resolver/socket attempt.

## Rollout

1. Merge only the reviewed implementation whose current head is reconciled to
   this plan, promotes durable docs, removes the temporary plan, and passes all
   required checks.
2. From the exact successful main SHA, nomination builds one deterministic
   Darwin/ARM64 archive once, computes SHA-256, uploads a named Actions artifact,
   and records structured run/artifact/digest/commit/issue identity.
3. Alfred downloads that artifact onto supported macOS, verifies the digest,
   provisions or reuses the marker-bounded disposable non-admin acceptance
   identity, starts and positively verifies the marker-bound PF/DTrace
   supervisor, executes the fresh matrix without credentials, captures
   sanitized evidence, and verifies policy/identity teardown plus protected-
   live-state canaries.
4. Product acceptance records issue #7, the exact candidate/digest, independent
   acceptor, verdict, and openable evidence.
5. Release downloads the same named artifact, re-verifies acceptance and digest,
   and publishes those exact bytes as the GitHub Release asset. It records tag,
   asset URL/digest, workflow/control SHA, issue, and rollback target.
6. Verify the public asset digest and issue lifecycle, then close issue #7.

There is no staging service or feature flag. The released command remains
inactive until the user explicitly invokes `remote attach` and confirms.

## Rollback And Recovery

- Before release, reject/revert the candidate and publish no asset.
- A failed or uncertain attach is never auto-undone. The user inspects
  repository-local `origin`; rerun classifies absent, exact, or collision.
- For a confirmed successful attachment the documented undo is Git's explicit
  repository-local `remote remove origin`, performed by the user. This issue
  does not add a detach command or invoke undo automatically.
- Release rollback restores the prior immutable GitHub Release as the
  recommended download and reverts source in a new commit/release. It does not
  edit repositories already configured by users.
- Artifact/digest/evidence mismatch fails closed. Published bytes are never
  overwritten; a correction receives a new immutable candidate and release.

## Release Prerequisites

- Address, config, terminal, privacy, fault, race, and production-boundary tests
  pass under focused, race, container, and native suites.
- Native Apple silicon/APFS/Git evidence proves real config locking, literal
  argv, PF/DTrace positive control and zero candidate network events/counters,
  no credential/global-Git effects, and protected live roots.
- Durable docs accurately describe shipped grammar, disclosure, recovery, and
  artifact promotion; reconciliation removes this temporary plan.
- Exact-byte build/transport/upload and marker-bounded macOS acceptance exist
  on the implementation head, whether reused from issue #4 or delivered here.
- Independent exact-candidate evidence is openable and accepted before release.

## Production Readiness Preflight

- **Secrets:** runtime attachment and acceptance require no secret slots,
  credentials, provider accounts, or credential helpers. The address grammar
  rejects credential channels; acceptance uses fixture strings only.
- **Deploy/promote:** nomination builds and uploads one named Darwin/ARM64
  archive and records run ID, artifact ID/name, digest, commit, and issue.
  Release downloads and verifies that exact artifact before publication.
- **Activation:** the digest-verified GitHub Release asset is production. User
  behavior remains explicit and dormant until `remote attach` plus `Attach`.
- **Verification:** CI proves parser/config/boundary behavior; macOS acceptance
  proves the fresh terminal matrix, local Git effects, PF/DTrace network guard,
  zero candidate events/counters, live-state isolation, and exact policy/user
  teardown; release verifies the public asset digest.
- **Rollback:** failed acceptance releases nothing. A bad release is superseded
  by the previous asset/source revert; user repositories are never changed by
  release rollback. Mismatched assets are refused, not overwritten.
- **Receipt:** commit, control workflow SHA, Actions run/artifact ID/name,
  SHA-256, issue #7, independent acceptor and evidence, tag/release/asset URL and
  digest, PF/DTrace policy/positive-control/zero-event/cleanup proof, protected-
  state proof, teardown proof, and rollback target.
