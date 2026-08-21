# Implementation Handoff

## Change Tier And Smallest Complete Outcome

Tier: **Risky and broad**. The change writes inside a mixed-ownership,
credential-adjacent user state root and adds a lifecycle subsystem, CLI surface,
schemas, recovery model, native acceptance, durable docs, and release evidence.

The smallest complete outcome is not “install works.” It is one self-contained
global `AGENTS.md` projection for a validated runtime repository with preview,
manifest ownership, verify, repair, compatible upgrade, one-step rollback,
uninstall, interruption recovery, unrelated-state preservation, and exact-
candidate acceptance under an isolated macOS identity.

## Dependency Order And Reviewable Slices

1. **Contracts and renderer**
   - Likely ownership: `internal/codexhome/`, embedded JSON schemas and fixtures.
   - Start with failing renderer/manifest tests.
   - Exit: deterministic self-contained projection and strict schemas pass pure
     tests without filesystem or environment access.
2. **Environment capability and read-only plan**
   - Likely ownership: `internal/environment/`, `internal/codexhome/plan.go`.
   - Start with injected-root/live-home/symlink/APFS/ownership failures.
   - Exit: production root discovery is pinned; tests cannot infer a live home;
     all lifecycle states and previews are deterministic and read-only.
3. **Install/uninstall transaction and recovery**
   - Likely ownership: `internal/codexhome/transaction.go`, recovery, tests.
   - Start with reversal and every-phase fault tests.
   - Exit: exact install/uninstall reversal, durable recovery, unrelated-tree
     preservation, and idempotency pass under race testing.
4. **Repair, upgrade, and rollback**
   - Likely ownership: same domain, one-generation store.
   - Start with drift, identity mismatch, compatible generation, and rollback
     consumption tests.
   - Exit: all replacement behavior is explicit and manifest-bounded.
5. **Terminal and command integration**
   - Likely ownership: `cmd/my-friday/main.go`, `internal/terminal/`, evidence
     fixtures.
   - Start with grammar, confirmation, error-class, and transcript tests.
   - Exit: the complete namespaced lifecycle is operable and recovery commands
     are copyable without leaking profile or unrelated state.
6. **Documentation, reconciliation, and release proof**
   - Likely ownership: durable docs, evidence manifest, implementation PR.
   - Exit: native/container CI, independent review, exact-candidate disposable-
     user evidence, docs promotion, plan deletion, reconciliation, nomination,
     acceptance, and release gates all agree on the same artifact.

Each slice may be a logical commit in one implementation PR. Do not merge a
partial slice that exposes mutation without its corresponding recovery and
denial tests.

## Acceptance Traceability

| Acceptance group | Slice | Required evidence |
|---|---|---|
| Manifest ownership and deterministic projection | 1 | Pure contract/render tests and schema fixtures |
| Correct Codex-home boundary and preview | 2 | Injected-root tests, official-doc trace, CLI preview transcript |
| Collision safety and unrelated preservation | 2-3 | Collision matrix and full fake-home snapshots |
| Interrupted recovery and complete reversal | 3 | Every-phase fault matrix, idempotent recovery, install/uninstall equivalence |
| Repair, compatible upgrade, rollback | 4 | Drift/identity/generation tests and sanitized transcripts |
| No daemon/background or privilege/network effects | 2-5 | Negative-action plan, process/network assertions, architecture review |
| Safe real-environment operation | 6 | Exact candidate under disposable non-admin macOS user/home |
| Production completion | 6 | Artifact digest, independent acceptance, tag, release, issue ledger |

Detailed cases and the evidence-boundary requirements live in
`04-verification.md`.

## Documentation Promotion

| Concern | Destination | Action |
|---|---|---|
| Major system boundary and flow | `docs/architecture.md` | Update to add installed baseline without weakening `init` boundaries |
| End-to-end installed lifecycle | `docs/architecture/installed-codex-baseline.md` | Create from the capability template; include contracts, authority, failure, recovery, and observability |
| Consequential regular-file/control-root decision | `docs/decisions/0002-manifest-owned-codex-baseline.md` | Create short ADR with issue #4 and planning-PR provenance |
| Contributor test isolation and native checks | `docs/development.md` | Update with injected-root prohibition and safe native commands |
| Candidate acceptance and artifact promotion | `docs/deployment.md` | Update and correct stale public-release prerequisites from actual workflow |
| Diagnosis/recovery/uninstall | `docs/runbook.md` | Add installed-baseline runbook, including drift refusal and transaction recovery |
| User command contract | `README.md` | Add concise lifecycle quick start and ownership/non-ownership boundary |

Reconciliation decides exact filenames from shipped behavior, records any
located drift, promotes only durable knowledge, and deletes
`docs/plans/4-managed-codex-baseline/` before the implementation PR leaves
draft. The temporary plan is not copied wholesale into permanent docs.

## Pull Request And Review Contract

- Branch from the exact approved planning merge on current `main`; use one
  `feature/managed-codex-baseline` implementation PR with a top-level `Refs #4`.
- Write failing tests first in the red/green order. Preserve reviewable logical
  commits when practical.
- Run focused tests, `go test -race ./...`, `bin/container bin/ci`, and native
  `bin/ci` on supported Apple silicon. Record exact commands and outcomes.
- Require independent security/trust-boundary review of root resolution,
  named-path access, manifest validation, journals, deletion authorization,
  symlink/hard-link/TOCTOU defenses, credential non-access, and test guards.
- Require evidence review proving no lifecycle test targeted Alfred's live
  `~/.codex`, a contributor's actual home, or deployed Batcomputer runtime.
- Keep the PR draft until reconciliation binds the exact head, durable docs are
  promoted, this plan is removed, and evidence is openable.
- Stop after the reviewed implementation PR. Merge, nomination, acceptance,
  and release require later authority backed by executable candidate transport
  and disposable-macOS harness design.

## Explicit Non-Goals And YAGNI Boundary

Do not add a CLI framework, database, daemon, watcher, launch agent, backup
service, general file synchronizer, TOML merge engine, Codex installer/updater,
skill installer, memory projection, hosted account integration, telemetry,
cross-machine sync, multi-generation history, arbitrary template language,
admin/root mode, Linux/Windows support, or a fictional staging environment.

Do not reuse repository-bootstrap transactions by generalizing them into an
all-purpose filesystem framework. Share small proven helpers only when their
invariants are identical; installed-state recovery owns a distinct mixed-root
contract.

## Exceptions That Reopen Design

Return to Solution Design if implementation evidence shows global `AGENTS.md`
is not a supported effective surface; if activation requires editing
`config.toml`, credentials, system config, or paths outside the user's Codex
home; if the product must adopt/merge foreign instructions; if a new daemon,
admin privilege, network service, or data boundary is required; if safe
recovery cannot be proved without broad snapshots/deletion authority; or if
work exceeds the approved execution envelope.

Ordinary method names, internal package splits, error-code numbering, and
fixture layout do not reopen design when they preserve these contracts.
