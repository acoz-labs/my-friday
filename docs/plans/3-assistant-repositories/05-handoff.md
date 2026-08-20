# Implementation Handoff

## Change Tier And Smallest Complete Outcome

**Tier: Broad with risky local-data operations.** This introduces the first
toolchain, public terminal workflow, generated contracts, and recoverable
two-target filesystem transaction. It touches no production service, secrets,
remote API, or existing user data by design.

The smallest complete outcome is one reviewed `my-friday init` flow that can
default-exit without mutation or explicitly create/verify both v1 repositories
on supported Apple Silicon macOS/APFS/Git, with deterministic preview and
exhaustive recovery evidence. Templates alone, one repo, or untested recovery
is incomplete.

## Dependency Order And Reviewable Slices

All slices remain one implementation PR because safety depends on one plan and
contract. Each begins with the red tests in `04-verification.md`.

| Slice | Likely paths | Depends | Exit condition |
|---|---|---|---|
| 1. Toolchain/environment | `go.mod`, `go.sum`, `mise.toml`, `internal/environment/`, `docs/development.md`, `bin/ci` | None | Pinned Go/dependency; typed macOS/ARM64/APFS/Git/TTY tests pass. |
| 2. Profile/contracts | `internal/profile/`, `internal/contract/`, embedded assets/schemas | 1 | Schema/grapheme corpus passes at 60/240 limits; profile cannot supply trust policy. |
| 3. Plan/rendering | `internal/plan/`, `internal/templates/`, `internal/projection/codex/`, `testdata/` | 2 | Stable IDs/actions/files/hashes and golden trees; Codex projection consumes neutral contracts without an adapter framework. |
| 4. Wizard/preview | `internal/terminal/`, `cmd/my-friday/` | 3 | Seven steps, retries, Exit/Create, ANSI-free transcripts, zero writes. |
| 5. Git/validation | `internal/git/`, `internal/repository/` | 3 | Fixed allowlist creates valid stages with no templates/commits/remotes. |
| 6. Transaction/recovery | `internal/transaction/`, `internal/paths/` | 3, 5 | Reservations, planned parent creation, empty-shell `0700` normalization/original-mode rollback, full fault matrix, recovery. |
| 7. End to end | `cmd/my-friday/`, `test/acceptance/` | 4, 6 | Success/Exit/collision/failure/recovery prove plan-to-mutation trace. |
| 8. Reconciliation/docs | durable docs and PR reconciliation; delete this plan | 1-7 | Docs match code, all checks pass, plan removed, no drift. |

Prompting, domain validation, planning, filesystem mutation, and command
execution remain separate packages so tests do not require terminal/disk
effects. This is a safety seam, not a plugin framework.

## Acceptance Traceability

| Acceptance group | Slices | Evidence |
|---|---|---|
| Deterministic preview | 2-4, 7 | canonical/golden tests and Exit transcript |
| Valid repos | 2, 3, 5, 7 | schemas, trees, Git/fresh-pair validation |
| No adjacent effects | 4-7 | spy, allowlists, filesystem diff, privacy scan |
| Safe paths | 1, 6, 7 | APFS symlink/case/nesting/collision matrix |
| Recoverability | 6, 7 | full fault matrix and idempotent recovery |
| Accessible terminal | 4, 7 | exact-head transcripts and design review |

## Documentation Promotion

Reconciliation describes shipped behavior rather than copying this plan.

| Concern | Destination | Action |
|---|---|---|
| Quick start, supported environment, boundary | `README.md` | Update user overview |
| Components, repo boundary, plan/transaction | `docs/architecture.md` | Update system overview |
| V1 contracts, creation, invariants, recovery | `docs/architecture/repository-bootstrap.md` | Create capability doc |
| Go/mise and exact tests | `docs/development.md` | Update development guide |
| Current artifact/no-release boundary | `docs/deployment.md` | Update actual delivery profile |
| Recovery diagnosis | `docs/runbook.md` | Update if recover ships |
| Native command/transaction choice | `docs/decisions/0001-native-bootstrap-command.md` | Create ADR |
| Manifest/profile fields | Embedded/copied JSON Schemas | Link as authoritative contracts |

A separate security reference is unnecessary unless implementation discovers a
larger trust surface; capability docs own local authorization/data/failure.

## Pull Request And Review Contract

- Branch `feature/assistant-repositories` from merged planning head; one draft
  implementation PR with top-level `Refs #3`.
- TDD in the slice order; do not weaken fault/collision tests.
- Run `bin/container bin/ci` plus native Apple Silicon macOS checks and document
  container limitations.
- Attach exact-head transcripts; obtain product-design review.
- Review/license the JSON Schema and grapheme-segmentation dependencies.
- Reconcile to this plan, explain drift, promote docs, delete
  `docs/plans/3-assistant-repositories/`, and run
  `solution-design-plan verify-implementation` before leaving draft.
- Independent maintainer review is required. `implementation` stops before
  merge/release unless authority expands it.

## Explicit Non-Goals And YAGNI Boundary

Refuse O2 install/global Codex writes; O3 record/promotion/search; starter skills;
commit/remote/provider/sync/credential/plugin/model setup; non-interactive or
force/adopt/import creation; Intel/non-macOS/non-APFS abstractions; GUI/TUI
framework, telemetry, daemon, database, package-manager integration, auto-update,
signing/notarization/release; general transaction frameworks; and personality
fields that influence trust or safety. Also refuse alternate harnesses, generic
adapter registries, or a lowest-common-denominator schema; Claude Code, pi, or
another harness needs a future product decision and capability mapping.

## Exceptions That Reopen Design

Return to Solution Design for a one-repository or installation outcome; global
Codex/import/secret reads; network/remote/daemon/privilege boundaries;
non-empty overwrite or unproven deletion; personality-controlled safety;
non-deterministic preview/execution; weaker recovery; another platform; or
merge/release beyond `implementation`.

Package names, internal methods, pinned patch versions, and copy edits within
the interaction contract do not reopen design.
