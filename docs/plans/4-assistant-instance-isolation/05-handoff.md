# Implementation Handoff

## Change Tier And Smallest Complete Outcome

Tier: **Risky and broad**. The visible command surface is bounded, but the
change creates and removes directory trees, controls Codex credential and
workspace boundaries, generates an executable launcher, supports concurrent
instances, and migrates prior state with deletion authority.

The smallest complete outcome is one current user creating two named,
manifest-owned instances beneath `$HOME/.my-friday/assistants/`, each with its
sole external native launcher at `$HOME/.local/bin/<name>`, launching each with
unchanged `HOME`, instance-local `CODEX_HOME` and fixed workspace, verifying
independent state and credentials, and safely removing or migrating one without
changing the other, launcher-directory siblings, or any existing user Codex or
shell state.

## Dependency Order And Reviewable Slices

1. **Name, layout, and manifest contract**
   - Likely ownership: `internal/plan/`, `internal/profile/`, a focused new
     assistant-instance package, and unit tests.
   - Exit: pure planning derives only the fixed contained tree and exact
     `$HOME/.local/bin/<name>` launcher leaf; invalid, reserved, ambiguous, and
     colliding names fail without mutation.
2. **Read-only classification and ownership**
   - Likely ownership: assistant lifecycle and platform path helpers.
   - Exit: canonical, absent, foreign, drifted, symlinked, and recovery states
     for both root and launcher are deterministic; only a supported manifest
     grants authority over the root and that one external leaf.
3. **Transactional create, verify, and remove**
   - Likely ownership: `internal/transaction/`, assistant lifecycle, platform
     implementations, and fault tests.
   - Exit: complete root promotion, external launcher no-replace projection,
     exact launcher/root cleanup, and partial rollback are idempotent,
     concurrency-safe, and recoverable without adjacent effects.
4. **Native launcher and managed dependencies**
   - Likely ownership: native command/launcher generation, environment contract,
     managed runtime/dependency resolution, and subprocess tests.
   - Exit: direct exec preserves `HOME`, fixes instance `CODEX_HOME` and Codex
     `--cd`, rejects manifest/path drift, avoids shells, and forwards only the
     documented argument surface.
5. **Multi-instance coexistence and credential boundary**
   - Likely ownership: integration, terminal, and acceptance tests.
   - Exit: two instances run independently; lifecycle operations and
     file-backed credential configuration remain selected-instance-local.
6. **O2 staged migration**
   - Likely ownership: lifecycle/transaction migration code and native fault
     acceptance.
   - Exit: replacement verification precedes cleanup; only exact prior-manifest
     O2 paths can be removed; an external launcher is replaced or restored only
     with exact prior-manifest proof; every interruption preserves a safe
     recovery path.
7. **Evidence, durable docs, and reconciliation**
   - Likely ownership: acceptance support, public-safe evidence, repository
     docs, and implementation PR reconciliation.
   - Exit: container/native CI, exact-path current-user matrix, credential-free
     lifecycle, separately credentialed smoke, independent reviews, docs
     promotion, plan removal, and exact-head reconciliation are complete.

Keep one implementation PR with reviewable logical commits. Do not merge a
deletion path, launcher, or migration slice without its corresponding denial,
fault, containment, and credential-leak tests.

## Acceptance Traceability

| Acceptance group | Slice | Required evidence |
|---|---|---|
| Fixed root and external launcher | 1-3 | name/path/case/symlink/collision tests and exact root plus `$HOME/.local/bin/<name>` receipt |
| Instance-owned state | 1-3 | manifest root/external semantics and lifecycle before/after evidence |
| Unchanged `HOME`; fixed `CODEX_HOME`/`--cd` | 4 | native subprocess environment and literal argv evidence |
| Managed dependency isolation | 4-5 | separate executable/dependency resolution markers per instance |
| Multi-instance coexistence | 5 | concurrent launch and sibling-preservation matrix |
| File-backed credential isolation | 5, 7 | credential-free CI plus separate redacted live smoke |
| Manifest-proven O2 migration | 6 | success and fault matrix with exact cleanup plan/receipt |
| Existing state preservation | 3-7 | current-user canaries for launcher directory/siblings, Codex, shell, home, sibling instance, and checkout paths; no second outside mutation |

Detailed cases and evidence boundaries live in `04-verification.md`.

## Documentation Promotion

| Concern | Destination | Implementation action |
|---|---|---|
| System and instance boundary | `docs/architecture.md` | Add named-instance root, sole `$HOME/.local/bin/<name>` projection, and unchanged-`HOME` boundary |
| Complete capability contract | new focused architecture document under `docs/architecture/` | Document root/external state model, manifest, launcher collision rules, authority, failure, and recovery |
| Durable tradeoff | next available `docs/decisions/` ADR | Record native launcher with `CODEX_HOME`/`--cd` versus caller env, substituted `HOME`, aliases, and OS users |
| User lifecycle | `README.md` | Add create, inspect, launch, coexistence, remove, and credential-location guidance |
| Contributor verification | `docs/development.md` | Add focused tests, fault injection, native containment fixtures, and redaction rules |
| Migration/recovery operations | `docs/runbook.md` | Add O2 preflight, staged verification, exact cleanup, interruption, and manual-preservation guidance |
| Installation and release boundary | `docs/deployment.md` | Document managed dependencies and later acceptance/release prerequisites from shipped behavior |
| Security and reporting | `SECURITY.md` where warranted | Clarify same-user isolation claims and credential evidence exclusions |

Reconciliation selects exact filenames and the next free ADR number from
implementation `main`, documents material drift, promotes shipped behavior
rather than copying this pack, and deletes
`docs/plans/4-assistant-instance-isolation/` before the implementation PR leaves
draft.

## Pull Request And Review Contract

- Branch the implementation from the exact merged planning head on `main`; use
  one feature PR with top-level `Refs #4`.
- Preserve existing lifecycle and installed Codex baseline behavior unless the
  named-instance contract explicitly supersedes it through the staged migration.
- Write failing tests first in the sequence from `04-verification.md`.
- Run focused packages, `go test -race ./...`, `bin/container bin/ci`, native
  `bin/ci`, and the public-safe exact-path acceptance matrix.
- Require independent security review of canonicalization, symlink and
  case-insensitive collision handling, the fixed external launcher and
  no-second-projection invariant, manifest authority, recursive deletion,
  launcher environment/argv, same-user claims, credential handling, and
  migration fault recovery.
- Require independent product/terminal review of create, collision, verify,
  launch, coexistence, migration, recovery, and removal transcripts.
- Keep the PR draft until durable docs are promoted, this temporary plan is
  removed, exact-head reconciliation is current, and credential-free plus
  separately credentialed evidence is openable and sanitized.
- Stop at an implementation-ready reviewed head. The approved envelope does not
  authorize merge, user-state migration, nomination, release, or activation.

## Explicit Non-Goals And YAGNI Boundary

Do not add caller-selected managed roots, launcher paths, or `CODEX_HOME`;
substitute `HOME`; edit `.zshrc`, `.bashrc`, profiles, aliases, functions, or
global `PATH`; create, chmod, or adopt `$HOME/.local/bin`; create a second
outside projection; create macOS users; add containers, VMs, sandbox claims, a
daemon, registry service, database, multi-user sharing, remote synchronization,
secret manager, credential acquisition, credential copying, arbitrary
environment profiles, shared mutable instance directories, automatic
migration, best-effort cleanup, unmanifested state adoption, or generalized
plugin/dependency management.

Do not use successful live credential authentication as proof of containment.
Containment is proven credential-free; the live smoke proves only that one
separately configured instance can use its own supported Codex credentials.

## Exceptions That Reopen Design

Return to Solution Design if implementation evidence shows Codex cannot use
file-backed credentials from instance `CODEX_HOME` while real `HOME` remains
unchanged; if `$HOME/.local/bin/<name>` cannot be the sole safe external
projection without shell or parent-directory changes; if a native launcher
cannot fix `--cd` without exposing an unsafe argument override; if managed
dependencies require user-global mutation; if case/symlink semantics prevent
reliable containment; if safe O2 migration requires deletion beyond exact
prior-manifest ownership; if multiple instances must share mutable state; if a
stronger OS security boundary becomes a product requirement; or if work must
exceed the `implementation` envelope.

Ordinary package names, error codes, and exact manifest serialization do not
reopen design when they preserve the fixed ownership, external launcher,
containment, environment, credential, coexistence, and migration contracts.
