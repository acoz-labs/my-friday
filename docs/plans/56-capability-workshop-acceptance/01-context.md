# Context

## Problem And Desired Outcome

Issue #51 implemented the capability workshop, but its nominated release cannot
be accepted: `bin/accept-installed-codex-baseline` refuses every issue except 4
and does not exercise capability authoring or lifecycle behavior. Issue #56
closes only that verification and release-authority gap.

## Current State

At repository basis `8e2371b433f4f6e4f28fe5c3491cc40b697d680b`:

- `internal/capability` and the CLI implement the instruction-only package,
  state vocabulary, lifecycle, migration, and recovery contracts.
- `bin/accept-installed-codex-baseline` provides mature Apple-silicon/APFS
  isolation, exact artifact download, live Codex credential injection, typed
  provisional/final evidence, and complete cleanup, but is issue-4-specific.
- `tools/acceptance-*`, `bin/verify-acceptance-evidence`,
  `bin/record-product-acceptance`, and `bin/finalize-release` enforce the
  existing acceptance ledger.
- `docs/deployment.md` names the issue-51 workshop journey but no executable
  harness implements it.

## Actors And Critical Journeys

- The independent acceptor runs one exact candidate in a disposable APFS and
  named-instance boundary and publishes evidence only after complete cleanup.
- The builder-enabled Codex task clarifies and authors a fixed useful fixture,
  shows its complete diff, and performs read-only validation/testing only.
- The terminal driver supplies exact lifecycle confirmation and observes every
  state transition, denial, interruption, migration, and reversal.
- The product owner and two design partners repeat the human workshop and
  publish bounded completion/comprehension receipts for the same artifact.
- Release workflows accept issue-51 evidence only when candidate, artifact,
  implementation PR set, and issue bindings all agree.

## Acceptance And Non-Goals

The full criteria are in issue #56. The slice ships no new capability behavior,
profile, dependency, service, runtime credential, or automatic migration. It
does not refactor the issue-4 supervisor merely to share code, weaken live-Codex
checks, or use contributor-local output as product acceptance.

## Constraints, Dependencies, And Risks

- Tests must prove script/schema behavior on CI without invoking Codex or APFS;
  the exact-candidate run remains Darwin/ARM64-only.
- Live acceptance consumes an existing auth file by path. No secret slot or
  value enters GitHub evidence.
- Model output is bounded by fixed prompts and exact tokens; deterministic CLI
  evidence remains authoritative for lifecycle correctness.
- Interruption or ambiguous detach/cleanup preserves the exact marked roots and
  emits failure authority that can never approve a release.
- The acceptance implementation changes the artifact bytes, so the current
  nomination cannot be reused.

## Evidence, Assumptions, And Unknowns

Evidence is the merged #54 plan, #55 implementation, current harness and release
scripts, and the successful but insufficient issue-51 nomination. We assume the
current Codex skills contract and file-backed authentication path remain usable
for a bounded fresh task. No product decision is unknown; live execution may
reveal an implementation or environment failure, which records
`changes-required` rather than reopening product scope.
