# Implementation Handoff

## Change Tier And Smallest Complete Outcome

Tier 3, release-bound verification and credential-handling change. The smallest
complete outcome is an issue-51-specific native acceptance command, strict
evidence/partner schemas and verifiers, product-acceptance/release integration,
complete cleanup/recovery, tests, and durable operator documentation. Partial
journey or evidence-only shipping is not useful.

## Dependency Order And Reviewable Slices

1. Characterize issue-4 isolation and release-gate rejection.
2. Add issue-51 schema/parser/verifier and cross-schema tests.
3. Add deterministic capability fixture and transcript driver.
4. Add APFS/native supervisor authoring and lifecycle journey.
5. Add live Codex builder and explicit-invocation PTY checks.
6. Add fault/cleanup/migration/partner-receipt matrices.
7. Integrate product acceptance and artifact finalization.
8. Promote docs, reconcile plan, run CI, security review, and exact acceptance.

Likely ownership includes a new `bin/accept-capability-workshop`, focused
`tools/acceptance-*` additions, acceptance evidence verifiers/tests,
product-acceptance/finalization scripts, workflow inputs only if required, and
deployment/runbook/security documentation. Reuse helpers only where their
existing contracts remain unchanged.

## Acceptance Traceability

Slices 2-3 cover typed authority and deterministic workshop states; slices 4-6
cover every issue criterion and failure/recovery path; slice 7 proves candidate
and release binding; slice 8 supplies independent exact-candidate and partner
evidence described in `04-verification.md`.

## Documentation Promotion

Reconciliation updates `docs/deployment.md` with the executable issue-51
acceptance contract, `docs/runbook.md` with interrupted-run recovery,
`SECURITY.md` with credential/evidence boundaries, and `docs/development.md`
with acceptance tests. Update architecture docs only if implementation changes a
durable component boundary; remove this plan before review readiness.

## Pull Request And Review Contract

Use one issue-56 implementation PR with `Refs #56` and an explicit linkage to
parent #51; neither issue closes at merge. Begin with failing tests, keep issue 4
green, obtain independent security/maintainer review, run `bin/ci`, reconcile
against this plan, remove the temporary plan, and verify the exact head. Merge
only with required checks green. Acceptance and release remain within the
approved envelope but use a newly nominated post-merge artifact.

## Explicit Non-Goals And YAGNI Boundary

No new capability profile, general acceptance framework rewrite, cloud service,
credential provider, telemetry, marketplace, automated user migration, shared
human-study database, or change to issue-4 evidence semantics. Do not expose
private instruction/diff/model/auth content to GitHub.

## Exceptions That Reopen Design

Return to Solution Design only if real Codex cannot demonstrate explicit skill
discovery without relaxing the no-activation boundary, partner evidence would
require durable personal profiling, credential handling needs a new secret or
trust boundary, cleanup cannot preserve ambient state, or release needs to
rebuild accepted bytes.
