# Implementation Handoff

## Change Tier And Smallest Complete Outcome

**Tier: Risky.** The code change is small but modifies release authority and
public evidence interpretation.

The smallest complete outcome is one new strict owner-dogfood acceptance
authority that product acceptance and release finalization both verify, plus
the owner recorder, external allowlist injection, failing-first adversarial
tests, and durable documentation. Shipping only the recorder, only acceptance,
only finalization, or only documentation would leave an authorization bypass or
an unusable contract.

## Dependency Order And Reviewable Slices

1. **Characterize and fail the active old contract**
   - Paths: `bin/test-capability-workshop-evidence`, focused acceptance fixtures.
   - Exit: new bundle fails; old bundle succeeds on active path; historical
     direct verifier behavior is locked.
2. **Strict owner receipt**
   - Paths: new `bin/record-capability-workshop-owner-receipt`, new
     `bin/verify-capability-workshop-owner-receipt`, authority fixtures.
   - Exit: valid allowlisted receipt round-trips; every schema, digest, author,
     issue, SHA, artifact, and migration-plan negative case refuses.
3. **Owner-dogfood bundle and actor separation**
   - Paths: new `bin/verify-capability-workshop-owner-dogfood`, shared narrowly
     extracted implementation-set helper only if required, fixture API stubs.
   - Exit: valid two-part bundle passes; owner/acceptor equality and either role
     matching any implementation author fail.
4. **Acceptance integration**
   - Paths: `bin/record-product-acceptance`,
     `.github/workflows/product-acceptance.yml`, acceptance contract tests.
   - Exit: issue-51 approval accepts only new success authority; failure path,
     nomination, checks, implementation digest, statuses, and retries remain
     exact.
5. **Release integration**
   - Paths: `bin/finalize-release`, `.github/workflows/release-artifact.yml`,
     release fixtures.
   - Exit: finalization reparses/reverifies the new bundle with the same owner
     allowlist and preserves no-rebuild artifact finalization.
6. **Durable documentation and complete verification**
   - Paths: `docs/deployment.md`, `docs/development.md`, `SECURITY.md`, and only
     other existing product copy that actively states the old requirement.
   - Exit: current contract, historical boundary, public claim scope, operation,
     failure, and rollback are accurate; full `bin/ci` passes.
7. **Exact-candidate release journey**
   - GitHub lifecycle/workflows, no code in a separate convenience candidate.
   - Exit: new candidate nominated, fresh evidence and owner receipt accepted,
     unchanged artifact released, both issues reconciled, and #92 unblocked.

## Acceptance Traceability

| Acceptance group | Slices | Evidence |
|---|---|---|
| Owner judgment, automation, migration plan | 2–4, 7 | Strict receipt/bundle fixtures and fresh exact-candidate tokens |
| Remove two-partner prerequisite | 1, 3–5 | Active parser rejects old bundle and accepts new bundle |
| No controlled-account independent claim | 2, 3, 6 | Strict schema, role language, claim-search tests |
| Documentation scope distinction | 6 | Repository docs and release-summary inspection |
| Integrity, separation, rollback preserved | 1–5, 7 | Adversarial fixtures, full CI, finalizer replay, release ledger |
| Invalid owner/evidence/candidate/artifact refuses | 2–5 | Negative fixture matrix in `04-verification.md` |

## Documentation Promotion

| Design concern | Durable destination | Action |
|---|---|---|
| Operator commands, bundle grammar, candidate/acceptance/release sequence | `docs/deployment.md` | Update current issue-51 contract and mark prior path historical |
| Fixture coverage and focused commands | `docs/development.md` | Update |
| Actor, evidence, privacy, public-claim, and comment-integrity boundaries | `SECURITY.md` | Update |
| General product quick start | `README.md` | Update only if it currently claims three-person or externally validated release |
| Operational recovery from interrupted evidence/acceptance | `docs/runbook.md` | Update only where active issue-51 recovery names old bundle |
| Architecture | `docs/architecture/capability-workshop.md` | Update only its acceptance boundary; workshop source/lifecycle architecture is unchanged |
| ADR | Not needed | Versioned schema and this issue/PR provide sufficient provenance; no runtime architecture changes |

Reconciliation must describe what actually shipped, promote only needed
contract text, and remove `docs/plans/101-owner-dogfood-acceptance/` before the
implementation PR leaves draft.

## Pull Request And Review Contract

- Use one task-scoped `feature/101-owner-dogfood-acceptance` implementation
  branch from the exact approved `origin/main`.
- Associate the PR with #101 and #51 using explicit top-level reference lines;
  update the issue-51 lifecycle implementation set before acceptance.
- Begin with failing authority tests and keep commits logically reviewable.
- Required checks: targeted shell fixtures, `go test -race ./...`, native
  Apple-silicon `bin/ci`, solution reconciliation, documentation promotion, and
  repository-required CI.
- Security review must inspect parser grammar, double fetch/digest behavior,
  actor normalization/allowlists, implementation-author resolution, workflow
  injection symmetry, retry semantics, and finalizer replay.
- The implementation PR must contain no credentials, private owner profile,
  workshop source, transcript, home path, or migration content.
- Independent maintainer review is required before merge. The implementer may
  not accept the resulting candidate.
- Reconcile the exact current implementation head against this plan, explain
  any drift, promote durable docs, and delete the temporary plan.

## Explicit Non-Goals And YAGNI Boundary

Do not implement:

- the portable assistant kernel or actual migration (#92);
- a participant database, telemetry, identity proofing, or “real person” API;
- a new independent-user receipt, sample policy, public-launch gate, or broad-
  usability badge;
- configurable arbitrary migration targets or free-form claim scopes;
- weakening or shortening the workshop harness;
- support for generic evidence on issue 51;
- acceptance of an old candidate under newer verifier code;
- rewriting/deleting historical comments, tokens, tags, or releases; or
- a generalized policy/rules framework.

## Exceptions That Reopen Design

Return to Solution Design if implementation evidence shows that GitHub cannot
provide stable comment authors/digests, the product owner must also be an
implementation author or evidence acceptor, issue #92 is no longer the migration
owner, the artifact would be rebuilt after acceptance, a new secret or external
identity service is required, or the desired claim expands from owner dogfood
to independently validated general usability.
