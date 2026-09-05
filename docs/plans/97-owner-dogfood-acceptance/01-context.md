# Context

## Problem And Desired Outcome

Issue #97 amends the release authority for the capability workshop delivered
under issue #51. The current contract requires one product-owner receipt and
two design-partner receipts from three distinct GitHub actors. That was useful
as an external-usability hypothesis, but it blocks the approved immediate
outcome: Anthony should be able to accept My Friday for his own Alfred
dogfooding without recruiting unrelated people.

The desired outcome is a narrower claim backed by equally strict technical
integrity: an authorized product owner judges a newly nominated exact candidate,
the complete workshop harness passes against those exact bytes, a separate
acceptor records the evidence, and the owner commits to collect real Alfred
migration evidence during #92. Independent-user evidence remains absent and
must not be implied.

## Current State

Verified at repository basis
`616395439a953100c6c165ccf56ea314c2dd0307`:

- `bin/verify-capability-workshop-acceptance` accepts only issue 51 and exactly
  four fields: workshop evidence, one product-owner receipt, and two design-
  partner receipts. It requires three distinct receipt actors.
- `bin/verify-capability-workshop-evidence` double-fetches provisional and final
  comments, verifies their body digests and author, and binds their strict JSON
  schemas to issue, candidate, artifact, checkout, controls, scenario results,
  preservation, and final cleanup.
- `bin/verify-capability-workshop-receipt` and
  `bin/record-capability-workshop-receipt` use the shared
  `capability-workshop-partner-receipt-v1` schema for `product-owner` and
  `design-partner`; product-owner membership is not independently allowlisted.
- `bin/record-product-acceptance` requires issue-51 approvals to carry the old
  four-part authority, verifies that the workflow actor authored the workshop
  evidence, enforces `ACCEPTANCE_ACTORS`, rejects acceptance by an implementation
  PR author, binds the complete lifecycle-linked implementation PR set, and
  writes issue-, commit-, artifact-, implementation-set-, acceptor-, and
  evidence-specific authority.
- `bin/finalize-release` refetches the latest issue acceptance and re-runs the
  issue-specific evidence verifier before publishing. Its parser currently
  recognizes only the old successful capability-workshop authority.
- `.github/workflows/product-acceptance.yml` injects `ACCEPTANCE_ACTORS`.
  `.github/workflows/release-artifact.yml` invokes finalization but neither
  workflow provides an owner allowlist.
- `bin/test-capability-workshop-evidence` exercises strict schema, author,
  stable-comment, candidate/artifact, cleanup, and duplicate-person checks.
  `bin/test-acceptance-contract` locks the issue-specific command wiring.
- `docs/deployment.md`, `docs/development.md`, and `SECURITY.md` describe the
  current three-person authority as the only issue-51 success path.
- My Friday is a public repository and its artifact delivery profile publishes
  an immutable GitHub Release without staging. Public availability therefore
  cannot itself be used as evidence that an independent user completed the
  workshop.

Live issue #51 nominates candidate
`2594147084e8bb6f961971efdb116b02824ecf66` and artifact
`artifact-v1:run=33946456531:id=9963477209:name=my-friday-darwin-arm64:sha256=d64f3a083cd6af9d1417e563cc29e91ed4666235394a1e78399c2df09a6c39ff`.
Its final workshop evidence is bound to those bytes. Because that commit lacks
the amended verifier, it cannot prove or enforce the new contract.

## Actors And Critical Journeys

### Product owner

Anthony uses the exact nominated artifact, reviews the exact technical
evidence, understands source versus projection and recovery, makes the
retain/remove judgment, accepts only the owner-dogfood claim, and records that
real migration evidence is due under issue #92. The GitHub actor must be in
`PRODUCT_OWNER_ACTORS` and must not be the evidence/product acceptor or an
author of any implementation PR in issue #51's bound set.

### Evidence author and product acceptor

An authorized non-owner actor operates the exact-candidate workshop harness,
publishes provisional and final evidence, and dispatches product acceptance.
Existing `ACCEPTANCE_ACTORS` and implementation-author separation continue to
apply. This role supplies technical and lifecycle independence; it is not
represented as an independent end user.

### Implementation contributors

Contributors implement the amendment and any prior issue-51 code included in
the lifecycle-linked implementation set. No author in that set may be the sole
product acceptor or the product owner for this authority.

### Later independent users

Future participants may provide evidence about usability outside the owner-
operated context. This change neither recruits them nor creates a new external-
validation schema. Controlled Alfred, contributor, maintainer, service, or
automation identities never qualify as independent users.

### Critical journeys

1. A newly merged amendment is built and nominated as one immutable candidate.
2. The separate evidence actor runs the existing complete workshop journey and
   publishes fresh strict evidence against that candidate and artifact.
3. Anthony completes hands-on owner review and records the new owner receipt,
   including the issue-92 migration-evidence promise.
4. Product acceptance verifies both tokens, actor allowlists, distinct roles,
   implementation separation, candidate ancestry, nomination, and checks.
5. Release finalization independently refetches and re-verifies the same
   authority before publishing unchanged bytes.
6. Missing, edited, stale, cross-candidate, unauthorized, duplicate-role, or
   pre-amendment evidence refuses without changing issue or release state.
7. Issue #92 later captures real migration evidence before its own release;
   failure there affects #92 rather than retroactively rewriting F0 evidence.

## Acceptance And Non-Goals

The design satisfies issue #97 by replacing the active three-person success
path with authorized owner judgment plus automated exact-candidate evidence,
retaining all existing integrity and rollback checks, and making claim scope
explicit. Tests must prove refusal of an invalid owner, missing evidence,
mismatched candidate/artifact, edited comments, duplicate roles, an
implementation author, and an absent/malformed migration plan.

Non-goals:

- no weakening of the workshop scenario matrix or cleanup requirements;
- no reuse of candidate `2594147…` or its evidence for the amended release;
- no claim that owner acceptance establishes general external usability;
- no implementation of #92 or collection of migration evidence before it can
  exist;
- no marketplace, telemetry, participant registry, or identity-provider
  integration;
- no automated classification of who is a “real person”; and
- no deletion or semantic reinterpretation of historical comments/tokens.

## Constraints, Dependencies, And Risks

- Issue #51 remains the canonical workshop release authority. Its lifecycle
  must link the amendment implementation so the accepted implementation digest
  includes every relevant merge.
- The implementation PR must associate both #97 and #51 as release-bearing
  issues, and the release set must include both or explicitly close #97 through
  the same verified ledger according to repository lifecycle tooling.
- Repository variables are mutable authorization configuration. Both
  acceptance and finalization must require and consume the same
  `PRODUCT_OWNER_ACTORS` value; missing configuration fails closed.
- GitHub comment bodies are mutable. Existing double-fetch plus stored body
  digest is retained for every component.
- A public release may be downloaded by anyone, but download availability is
  not validation. Claims must remain bounded until later evidence exists.
- No secrets or owner profile data enter receipts. GitHub login, bounded
  booleans/enums, issue numbers, hashes, and artifact identifiers are enough.

## Evidence, Assumptions, And Unknowns

### Evidence

- Issue #97 records Anthony's approved pivot and explicit acceptance criteria.
- Issue #51 and its product-design comment establish the original interaction
  contract; the pivot changes its validation requirement, not the workshop
  behavior.
- The repository paths above show symmetric acceptance/finalization
  verification and the existing exact-candidate safeguards.
- PR #96 at this repository basis establishes issue #92 as the approved
  portable-assistant migration path.
- GitHub reports repository visibility `PUBLIC` and default branch `main`.

### Assumptions

- Anthony's configured GitHub actor can author an issue comment through the
  existing CLI workflow.
- The current workshop harness remains the correct technical evidence matrix;
  implementation changes only authority composition and documentation.
- One repository variable is sufficient because owner authorization is a small
  operational allowlist, like `ACCEPTANCE_ACTORS`.

### Unknowns

None blocks implementation. The specific later independent-user sample,
method, and product claim are intentionally deferred until distribution goals
require them. Issue #92 will define and collect its own exact migration
evidence under its approved plan.

### Decisions

The selected authority grammar, owner allowlist, actor separation, fresh-
candidate requirement, migration-promise timing, and claim boundary are
detailed in `02-decision.md` and `03-design.md`.
