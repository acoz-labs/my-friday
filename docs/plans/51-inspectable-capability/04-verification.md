# Verification And Release Design

## Test Strategy

Likely implementation tests:

- `internal/capability/source_test.go`: package schema, frontmatter agreement,
  canonical paths/UTF-8, size/depth/count limits, unknown entries, modes, links,
  hardlinks, devices, nested repositories, fixed effect declarations, Codex
  compatibility, and deterministic case grammar.
- `internal/capability/state_test.go`: every stable state and precedence,
  source-changed versus installed-drift, disabled/absent distinctions, and
  body-free receipts.
- `internal/capability/lifecycle_test.go`: plan binding, collision refusal,
  exclusive projection, lock/concurrency, stale plan, injected failures at every
  phase, recovery truth table, idempotency, generation retention, and reversal.
- `internal/repository/repository_test.go`: v1 remains valid, v2 creation and
  initialization, placeholder migration, unknown-file denial, interruption and
  rollback.
- `internal/assistantinstance/instance_test.go` and integration tests: v1/v2
  create/verify/remove, explicit upgrade, builder digest ownership, capability
  lock serialization, pending-transaction refusal, and old-instance recovery.
- `cmd/my-friday/main_test.go` and `internal/terminal`: command grammar, stable
  errors, plain output, TTY denial, EOF/interrupt/default exit, exact-token
  matrices, long diff/path output, normalized Unicode, and no content leakage.
- `internal/capability/codex_contract_test.go`: golden official-contract fixtures
  for generated `SKILL.md` and fixed `agents/openai.yaml`; incompatible versions
  fail closed.
- `bin/ci`: formatting, static analysis, unit/integration/race tests, plan/docs
  validation, and acceptance contract tests.

Tests use secure temporary roots and run on Darwin and Linux where platform
semantics permit; descriptor/no-follow and APFS candidate acceptance remain
Darwin-specific. No deterministic test invokes Codex or the network.

## Red/Green Sequence

1. Characterize v1 runtime/instance validation and removal before contract
   changes; add failing v2 schema/layout fixtures.
2. Add failing source-profile and deterministic-case tests, then implement the
   pure parser/validator/digester.
3. Add failing state derivation and inspection tests, then implement read-only
   commands.
4. Add failing runtime initialization recovery tests, then migrate v1→v2 while
   preserving v1 verification.
5. Add failing named-instance upgrade/builder tests, then implement v2 creation
   and explicit upgrade.
6. Add failing install/verify lifecycle tests, then implement projection,
   receipt, locking, and recovery.
7. Add failing upgrade/disable/enable/remove matrices, then complete reversal.
8. Add CLI transcript/accessibility tests and builder golden content tests.
9. Extend artifact acceptance fixtures and release gates; run full `bin/ci` and
   race/fault suites.

## Acceptance Evidence

Terminal impact is behavioral, not rendered, under
`docs/operations/ui-acceptance.md`. Exact-candidate evidence must bind issue #51,
the merged implementation commits, nominated archive SHA-256, workflow/run, and
independent acceptor.

Automated evidence proves all schemas, state transitions, denial matrices,
no-follow ownership, transaction failures/recovery, deterministic package tests,
v1 compatibility, complete reversal, stable plain transcripts, and absence of
unexpected writes/network in sandbox fixtures.

Independent acceptance downloads the immutable Darwin/ARM64 candidate into the
existing disposable no-admin APFS/sandbox harness and:

1. creates a contract-v2 runtime/memory pair and named assistant, confirming the
   builder is present only in that instance;
2. uses the builder in a real fresh Codex task to define one useful
   instruction-only capability, records a redacted complete source diff, and
   confirms no activation occurred;
3. validates/tests, exercises safe-exit and exact `Install`, opens a fresh task,
   explicitly invokes the skill, and observes its declared behavior;
4. changes source, proves `source-changed`, exercises stale-plan denial, then
   tests and upgrades through a fresh exact plan;
5. proves foreign collision and owned-drift refusal without exposing contents;
6. disables, verifies a fresh task cannot discover the projection, enables,
   removes, and proves source remains while all instance control/projection is
   gone;
7. injects one post-mutation interruption and completes exact recovery;
8. separately migrates a standalone v1 runtime and a disposable v1 instance,
   proving that instance upgrade uses only its private copy and supports rollback;
9. verifies ambient HOME, global Codex skills/config, credential source, sibling
   instance, and protected canaries are unchanged.

The product owner and at least two design partners then repeat the approved
workshop journey against the exact candidate and record completion,
source/projection comprehension, recovery needs, and retention. This pilot
evidence is required for acceptance but does not become telemetry or durable
user profiling.

## Rollout

Execution envelope: `through-production`.

1. Implement on one issue-linked branch and keep the feature unavailable until
   all lifecycle slices are coherent.
2. Run `bin/ci`, security review, and contributor artifact-contract checks.
3. Merge only after required review/checks and plan reconciliation.
4. Nominate the exact Darwin/ARM64 artifact through the existing release
   workflow; no rebuild may occur after nomination.
5. Independent acceptance and design-partner evidence bind to that artifact.
6. Record product acceptance and production deployment receipt, finalize the
   GitHub Release, download it anew, verify digest/help/version, and close the
   lifecycle only after release evidence is complete.

No existing runtime or named assistant changes merely because a new binary is
released. Users explicitly run runtime initialization and assistant upgrade;
newly created roots use v2.

## Rollback And Recovery

- Before release, revert the implementation commit set and nominate a new
  artifact; never relabel an already accepted archive.
- Runtime initialization retains the exact v1 placeholder/prestate until commit
  and can restore it through its journal. After successful migration, repository
  rollback is an explicit tested v2→v1 operation only while no capability source
  exists; otherwise it refuses and preserves source.
- Assistant upgrade retains a proven v1 generation. `assistant rollback NAME`
  restores v1 only when no user capability receipt/transaction exists and the
  builder projection is exact; otherwise it refuses.
- Ordinary capability recovery finishes or restores the journaled projection.
  Disable/remove never delete source. Release rollback does not auto-downgrade
  user data; retain the last compatible binary until all pending journals clear.
- A bad published release is superseded by a new fixed artifact under repository
  policy. Existing healthy v1 and v2 roots remain operable by compatible binaries.

## Release Prerequisites

- Independent security review of instruction-injection claims, path ownership,
  migration/recovery tables, and confirmation wording.
- Official Codex skill contract golden fixtures captured without vendoring
  proprietary documentation; any incompatible upstream change reopens design.
- Exact acceptance harness support for repository-scoped skills and fresh-task
  discovery.
- Product owner plus two design-partner workshop receipts on the exact artifact.

These are implementation/acceptance prerequisites, not unresolved plan choices.

## Production Readiness Preflight

- **Secrets:** no new slot. Acceptance reuses the existing file-backed Codex
  authentication injection; values never enter evidence.
- **Workflow:** existing nomination, acceptance, product-acceptance, production
  receipt, finalization, and release-asset scripts remain authoritative and are
  extended only for issue #51 evidence grammar.
- **Deploy:** this artifact product has no service deployment; publishing the
  exact accepted GitHub Release through the existing production workflow is the
  deploy action, and no user root is mutated automatically.
- **Activation:** release publication does not mutate users. A standalone runtime
  and an existing named assistant each require an explicit migration; the latter
  migrates its private runtime copy rather than following external source.
- **Verification:** exact archive digest, CLI version/help, v1 compatibility,
  v2 bootstrap, complete capability lifecycle, fresh-task Codex behavior,
  ambient preservation, and reversal are named above.
- **Rollback:** artifact supersession plus bounded v2→v1 rollback and retained
  compatible binary are testable; automatic data downgrade is forbidden.
- **Receipt:** the existing release receipt must bind artifact digest,
  implementation closure, independent acceptance, design-partner evidence,
  workflow/run SHA, issue/PR set, and release URL.
