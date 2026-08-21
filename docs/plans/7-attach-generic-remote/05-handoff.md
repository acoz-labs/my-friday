# Implementation Handoff

## Change Tier And Smallest Complete Outcome

Tier: **Risky and broad**. The visible CLI is narrow, but it parses an address
that Git could later execute, writes mixed-ownership `.git/config`, needs
collision/interruption/TOCTOU/privacy guarantees, and must complete the
repository's missing exact-byte release path if not already present.

The smallest complete outcome is a released `remote attach` flow for either one
runtime or memory repository with bounded credential-free HTTPS/SSH input,
role-first disclosure, exact confirmation, one canonical local `origin`, exact
read-back, idempotent repeat, collision refusal, privacy-safe recovery, complete
prohibited-effect evidence, and independent acceptance of the immutable bytes.

## Dependency Order And Reviewable Slices

1. **Address contract**
   - Likely ownership: new `internal/remoteaddress/` package and tests.
   - Start with allowlist/denylist/fuzz failures.
   - Exit: all accepted forms preserve exact bytes; all credential/helper/local/
     unsafe/leading-option forms fail without raw-value output or Git invocation.
2. **Single-repository identity**
   - Likely ownership: `internal/repository/repository.go` and tests.
   - Start with runtime/memory/evolved/symlink/tamper cases.
   - Exit: one repository can be recognized without a pair or fresh-state check,
     while existing pair/fresh validation remains unchanged.
3. **Read-only local-config adapter**
   - Likely ownership: new `internal/remoteconfig/`, `internal/gitexec/`.
   - Start with fixed environment/literal argv and absent/canonical/collision
     snapshots, including includes/corruption/global canaries.
   - Exit: deterministic plans inspect only direct local config and cannot
     invoke shell/network/credential/helper/global Git behavior; Git 2.28 and
     native suites prove private HOME/XDG isolation. Empty/comment-only section
     text is tested as key-semantic absence without a raw config parser.
4. **Terminal plan and safe exits**
   - Likely ownership: `internal/terminal/remote_wizard.go`, CLI routing/tests.
   - Start with help, complete role copy, exact `Attach`, cancel, and collision
     transcripts.
   - Exit: every nonmutating journey is complete, accessible, and privacy-safe.
5. **Git mutation and recovery**
   - Likely ownership: `internal/remoteconfig/`, terminal integration/evidence.
   - Start with exact delta/read-back, then lock/fault/permission/race/TOCTOU.
   - Exit: one canonical add can succeed; repeat is read-only; every ambiguity
     preserves state and returns safe inspect/rerun guidance.
6. **Docs, reconciliation, artifact acceptance, and release**
   - Likely ownership: durable docs, evidence fixtures, nomination/acceptance/
     release workflows and scripts when the exact-byte chain is absent.
   - Exit: container/native CI, independent review, same-byte nomination,
     positive-controlled PF/DTrace disposable-macOS acceptance, GitHub Release
     digest, cleanup receipt, and issue lifecycle all bind the same candidate.

Keep one implementation PR and reviewable logical commits. Do not merge a
parser or config write without its corresponding denial, privacy, fault, and
no-adjacent-effect evidence.

## Acceptance Traceability

| Acceptance group | Slice | Required evidence |
|---|---|---|
| Recognized one-repository role | 2 | Runtime/memory/evolved/tamper/path tests and transcripts |
| Credential-free generic address | 1 | Table/fuzz tests and no-Git boundary proof |
| Consequence disclosure and exact confirmation | 4 | Runtime/memory PTY transcripts and accessibility review |
| Canonical local `origin` and read-back | 3, 5 | Exact argv/env/read-back tests, adjacent-name checks, and native Git evidence |
| Idempotency/collision/failure recovery | 3, 5 | Repeat, lock, corruption, fault, race, and rerun matrix |
| Empty textual origin sections | 3, 5 | Git 2.28/current key-semantic fixtures; comment/adjacent-byte preservation |
| No adjacent effects/privacy leakage | 1-6 | Argv observer, PF/DTrace positive control and zero-event/counter proof, pair/global/home snapshots |
| Exact candidate acceptance/release | 6 | Candidate digest, independent evidence, tag/asset receipt |

Detailed cases and the exact-candidate evidence contract live in
`04-verification.md`.

## Documentation Promotion

| Concern | Destination | Action |
|---|---|---|
| System boundary | `docs/architecture.md` | Update to add optional post-bootstrap remote configuration while preserving local-only `init` |
| End-to-end attach contract | `docs/architecture/remote-attachment.md` | Create from capability template with grammar, state, authority, failure, recovery, and observability |
| User command and privacy boundary | `README.md` | Add concise usage, disclosure, safe examples, and explicit non-effects |
| Git/address test and evidence workflow | `docs/development.md` | Update with focused tests, fixture-address policy, PTY/transcript generation, and no-real-address rule |
| Exact artifact nomination/acceptance/release | `docs/deployment.md` | Update from the actual reused or newly delivered same-byte workflow and remove stale blockers |
| Disposable PF/DTrace acceptance supervisor | `docs/deployment.md` and `docs/runbook.md` | Document privilege separation, positive control, evidence schema, exact cleanup, and mismatch recovery from shipped scripts |
| Collision, lock, verification-pending, undo | `docs/runbook.md` | Add address-free Git inspection/recovery guidance and repository-local manual removal |
| Bounded grammar choice | `docs/decisions/0002-bounded-git-remote-addresses.md` or next free ADR | Create short ADR because security/compatibility tradeoff is likely to be questioned |

Reconciliation chooses the next free ADR number from implementation `main`,
records any plan drift, promotes shipped behavior rather than copying this pack,
and deletes `docs/plans/7-attach-generic-remote/` before the implementation PR
leaves draft.

## Pull Request And Review Contract

- Branch from the exact merged planning head on current `main`; use one
  `feature/attach-generic-remote` implementation PR with top-level `Refs #7`.
- Write failing tests first in the red/green order. Preserve existing init,
  validate, recover, contract-v1, and repository evolution behavior.
- Run focused packages, fuzz seeds, `go test -race ./...`, static Darwin/ARM64
  build, `bin/container bin/ci`, and native `bin/ci` on supported Apple silicon.
- Require independent security review of address parsing, external-helper and
  credential rejection, literal argv/environment, config includes/locking,
  symlink/inode/TOCTOU limits, read-back, uncertain recovery, redaction, and
  prohibited side effects, including the stored-address/future-rewrite boundary.
- Require acceptance-harness review of disposable-UID quiescence, PF anchor and
  enable-token scope, DTrace child/resolver/socket coverage, positive controls,
  counter/event reset, tracer-loss refusal, candidate isolation, and exact
  cleanup without global PF disable/flush.
- Require product-design review of the final exact-head runtime/memory/cancel/
  error transcripts at 80 columns and in screen-reader order.
- Keep the PR draft until reconciliation binds its exact head, durable docs are
  promoted, this plan is removed, and all transcript/config/observer/PF/DTrace
  evidence is openable and sanitized.
- Continue after reviewed merge through deterministic nomination, independent
  Alfred-run disposable-macOS acceptance, same-byte GitHub Release publication,
  verification, and issue closure under the `through-production` envelope.

## Explicit Non-Goals And YAGNI Boundary

Do not add a CLI framework, URL-normalization library, credential scanner,
credential helper, secret store, provider SDK/CLI, provider login, destination
creation, permissions/visibility UI, connection test, fetch/push/commit, custom
remote name, replacement/detach command, local/file/bundle/plaintext transport,
arbitrary helper support, config include engine, direct config writer, general
Git transaction framework, attachment manifest, telemetry, daemon, sync loop,
multi-repository batch, Linux/Windows/Intel support, or fictional staging.

Do not generalize repository bootstrap transactions. Git owns this single
config-file lock/update; the safe recovery model is classify-and-rerun, not a
second journal.

## Exceptions That Reopen Design

Return to Solution Design if implementation evidence shows `git remote add`
cannot provide the canonical pair without reading external config; if common
target destinations cannot be represented without adding a materially riskier
address/credential/helper surface; if one attachment requires modifying both
repositories, content, refs, credentials, global config, or a provider; if a
safe uncertain-write model requires automatic deletion/overwrite; if exact
artifact promotion needs a new secret/service/trust boundary; or if work exceeds
the approved execution envelope.

Ordinary package/method names, error-code numbers, parser implementation detail,
fixture layout, and reuse of an already merged exact-byte chain do not reopen
design when they preserve these contracts.
