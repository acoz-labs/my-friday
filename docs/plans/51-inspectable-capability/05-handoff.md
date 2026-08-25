# Implementation Handoff

## Change Tier And Smallest Complete Outcome

Tier 3: broad, security-sensitive product capability with persistent contract
migrations and a release. The smallest complete outcome is one named assistant
that can use the built-in builder to author, review, validate/test, install,
verify, enhance, disable/enable, remove, and recover one instruction-only
capability while preserving source and unrelated state.

Do not ship only package scaffolding, only the builder prompt, or install without
upgrade/reversal/recovery.

## Dependency Order And Reviewable Slices

1. **Pure capability contract** — own `internal/capability` schemas, fixtures,
   package parser, path policy, digest, deterministic cases, state vocabulary.
   Exit: exhaustive pure tests and no mutating code.
2. **Runtime contract v2** — own `internal/repository` creation/validation plus
   explicit initialize/rollback/recover adapters. Exit: v1 compatibility and
   fault-injected v1↔v2 evidence.
3. **Named-instance v2 bootstrap** — own `internal/assistantinstance`, embedded
   builder bytes, create/upgrade/rollback/remove/recover. Exit: builder ownership,
   instance isolation, and migration matrices pass.
4. **Read-only workshop surfaces** — own inspect/validate/test CLI composition,
   plain transcript contract, builder golden instructions, and docs examples.
   Exit: agent can create source but cannot activate it.
5. **Install and verify** — own projection planner, receipt/generation state,
   locks, atomic promotion, verify, collision/drift denial, and fault recovery.
   Exit: exact install and full recovery suite pass.
6. **Enhancement and reversal** — own upgrade, disable, enable, remove, state
   transitions, source preservation, and concurrent instance-removal denial.
   Exit: complete lifecycle and no-adjacent-effect tests pass.
7. **CLI/accessibility/security closure** — exact-token matrices, TTY/EOF,
   stable errors, Unicode/path bounds, content-redaction and no-ANSI evidence.
   Exit: security review has no blocking finding.
8. **Acceptance and release** — extend candidate harness/evidence grammar,
   complete independent/design-partner acceptance, nomination, production
   receipt, GitHub Release, and fresh download verification.

Each slice begins with the failing tests in `04-verification.md`. Parallel work
must not split ownership of shared manifest/transaction code without an explicit
integration owner.

## Acceptance Traceability

| Acceptance group | Slices | Evidence |
|---|---|---|
| Define/scaffold/review without activation | 1, 3, 4 | builder golden tests, Git diff transcript, fresh-instance isolation |
| Deterministic validation/tests | 1, 4 | exhaustive schema/path/case tests; no model/network |
| Exact install/verify | 5, 7 | plan binding, token matrix, projection receipt, healthy state |
| Enhancement | 4–6 | source-changed, stale-plan denial, fresh Upgrade evidence |
| Disable/enable/remove | 6 | state/reversal matrix and preserved source |
| Collision/drift/interruption | 5–7 | foreign canaries, drift refusal, injected-phase recovery |
| v1 migration/rollback | 2, 3 | runtime and instance compatibility matrices |
| Real usefulness and exact release | 8 | immutable artifact, fresh Codex tasks, independent and partner receipts |

## Documentation Promotion

| Destination | Required shipped contract |
|---|---|
| `README.md` | bootstrap, workshop, lifecycle examples, fresh-task caveat |
| `docs/architecture.md` | capability components and source/projection boundary |
| `docs/architecture/capability-workshop.md` | package/state/command/ownership contract |
| `docs/decisions/0004-source-first-instruction-capabilities.md` | selected profile, bootstrap projection, rejected global/symlink/plugin paths |
| `docs/development.md` | fixtures, deterministic-test boundary, fault/security commands |
| `docs/deployment.md` | exact capability acceptance and activation/migration notes |
| `docs/runbook.md` | initialize/upgrade/capability recovery, drift, rollback, compatible-binary retention |
| `SECURITY.md` | structural-profile limits, prompt-instruction risk, same-UID/TTY non-guarantees |

Implementation reconciliation must update these from shipped behavior and remove
`docs/plans/51-inspectable-capability/` before the implementation PR leaves
draft.

## Pull Request And Review Contract

- One issue-linked implementation branch/PR unless a prerequisite must ship
  independently; every release-bearing commit remains traceable to #51.
- Required checks: `bin/ci`, race/fault suites, plan reconciliation, docs link
  validation, artifact/evidence contract tests, and platform-specific acceptance.
- Independent review must cover security/trust claims, no-follow ownership,
  migration/recovery, CLI/accessibility behavior, and exact-candidate release.
- The PR description records plan provenance, contract migrations, Decision
  Spotlight reconciliation, test/acceptance matrix, docs promotion, remaining
  risks, and release envelope.
- No formal product acceptance may reuse contributor evidence or a rebuilt
  artifact. Merge, nomination, acceptance, production receipt, and release must
  follow repository policy in order.

## Explicit Non-Goals And YAGNI Boundary

No executable/script capability, MCP server, plugin, arbitrary dependency,
network/credential grant, background process, durable capability-owned data,
remote source, package registry, marketplace, publishing, signing ecosystem,
automatic update, cross-user installation, global skill management, semantic
safety certification, authenticated-human confirmation, current-session hot
unload, graphical UI, telemetry, capability composition, or governed memory.

Do not generalize the manifest to future profiles, add a dependency framework,
or expose arbitrary `agents/openai.yaml` fields. Unsupported requests fail with
the named profile boundary.

## Exceptions That Reopen Design

Return to Solution Design if implementation evidence requires source and
projection to be the same tree, global Codex mutation, scripts/dependencies,
network or credentials, a cross-root atomic transaction, deletion of repository
source, implicit invocation for user capabilities, an unbounded/customizable
package policy, a new external service or secret, weakening v1 recovery, or an
execution path beyond `through-production` repository policy.
