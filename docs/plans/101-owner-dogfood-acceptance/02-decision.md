# Solution Decision

## Decision Drivers

1. Remove an unrelated participant-recruitment dependency from internal assistant
   dogfooding.
2. Preserve exact issue/candidate/artifact/evidence binding and immutable
   release bytes.
3. Keep product judgment separate from automated evidence and implementation.
4. Make the narrower evidence claim honest on a publicly distributed artifact.
5. Prevent arbitrary or controlled accounts from being presented as product
   owners or independent users.
6. Avoid making future migration evidence a circular prerequisite for the F0
   release that enables that migration.
7. Keep acceptance and release-time verification symmetric, fail-closed, and
   covered by the existing fixture architecture.
8. Preserve historical auditability without supporting two simultaneous active
   interpretations of the same token.

## Competing Approaches

### A. Reinterpret the existing four-part token

Make partner fields optional or accept duplicate/controlled actors for internal
dogfood.

### B. New owner-dogfood bundle with fresh owner receipt (selected)

Introduce a strict two-part bundle containing existing exact-candidate workshop
evidence and a new owner-dogfood receipt. Require a configured owner actor,
three-way role separation, fresh amended candidate, and an issue-92 migration-
evidence promise.

### C. Product-owner workflow dispatch alone

Treat the product owner's workflow dispatch and acceptance summary as sufficient; retain
only the automated evidence token.

### D. Release the existing candidate and amend policy in later control code

Use the already passing candidate/evidence while relying on the default branch
to interpret it under the new rules.

### E. Keep three-person validation but permit assistant-controlled accounts

Satisfy the cardinality check with distinct GitHub accounts already under the
same operator's control.

## Adversarial Comparison

| Approach | Advantage | Fatal flaw or tradeoff |
|---|---|---|
| A | Minimal edits | Changes the meaning of a versioned historical token; optional fields and ambiguous arity weaken fail-closed parsing |
| B | Explicit policy version, bounded scope, strong audit trail | Requires one new receipt schema, configuration variable, and fresh candidate/evidence run |
| C | Simplest user flow | Conflates GitHub workflow authority with hands-on product judgment and cannot prove owner comprehension or the migration commitment |
| D | Avoids rebuilding/rerunning | The released bytes do not contain the verifier that grants authority; finalization would rely on code different from the artifact being accepted |
| E | Satisfies syntax without recruiting | Fabricates independence and produces no evidence about an external user's comprehension |

Approach B is the only path that both removes the artificial gate and preserves
the exact-candidate trust model. Its additional schema is deliberate policy
versioning, not an abstraction layer.

## Selected Approach

Create these versioned authorities:

```text
capability-workshop-owner-dogfood-v1:<workshop-evidence>|<owner-receipt>

capability-workshop-owner-receipt-v1:comment=<id>:<sha256(body)>
```

The receipt comment is strict, redacted, and bound to issue 51, candidate SHA,
artifact, role `product-owner`, completed hands-on judgment, source/projection
comprehension, no outstanding recovery, retain/remove decision, acceptance
scope `owner-operated-dogfood`, `migration_evidence_issue: 92`,
`migration_evidence_due: "before-release"`, and
`independent_user_validation: "not-collected"`.

The bundle verifier:

- accepts issue 51 only;
- double-fetches both evidence and owner comment through the component
  verifiers and validates body digests;
- requires the evidence author to equal the authorized product-acceptance actor;
- requires the owner comment author to be listed in `PRODUCT_OWNER_ACTORS`;
- requires owner and evidence/product-acceptance actors to differ;
- resolves every lifecycle-linked implementation PR and rejects an owner or
  acceptor matching any implementation author;
- requires identical issue, candidate, and artifact across all authority;
- accepts no extra fields or token parts; and
- emits no independent-user success claim.

`record-product-acceptance` accepts only the new success token for issue 51
after the amendment. Failure continues to use
`capability-workshop-failure-v1`; weakening failure evidence is unnecessary.
`finalize-release` recognizes and re-verifies the new success token with the
same owner allowlist. The legacy four-part verifier remains executable for
historical audit and fixture compatibility, but it is removed from active
success-token parsing so it cannot authorize a new acceptance or release.

Implementation and documentation must call the old design-partner contract
historical validation evidence, not independent-user success. A later public-
claim initiative can design a real external-validation authority once a claim,
sample, and distribution decision exist.

Confidence is high. This is a narrow authority-composition change built on
existing strict comment, evidence, actor, candidate, artifact, implementation-
set, acceptance-status, and release machinery. Residual risk lies chiefly in
parser asymmetry and lifecycle association, both directly testable.

## Decisions Ledger

| Decision | Rationale | Evidence |
|---|---|---|
| New two-part `owner-dogfood-v1` bundle | Version policy rather than weakening or reinterpreting v1 | Existing verifier has strict four-part grammar |
| New strict owner receipt | Owner judgment and migration promise need typed immutable authority | Existing partner receipt lacks scope and migration fields |
| Preserve existing workshop evidence schema | Technical matrix and cleanup proof do not change | Issue #101 changes validation participants, not workshop behavior |
| Require fresh amended candidate and evidence | Candidate must contain its own active verification path | Immutable artifact and release-finalization model |
| `PRODUCT_OWNER_ACTORS` repository variable | Authorize owner without hard-coded identity or profile data | Existing `ACCEPTANCE_ACTORS` pattern |
| Owner, acceptor/evidence author, implementation authors all distinct | Maintains judgment, technical evidence, and implementation separation | Repository SDLC and issue #101 integrity criteria |
| Issue #92 migration promise in receipt; actual proof later | Makes the commitment durable without creating a circular dependency | #92 is sequenced after F0 release |
| Owner-operated claim only | Public availability is not external usability evidence | Live public repository and absent independent-user evidence |
| Legacy verifier audit-only | Preserve historical meaning while eliminating an active bypass | Immutable comment/token audit requirement |
| No future external-validation schema now | Claim/sample requirements are not yet shaped | YAGNI and issue #101 non-goals |
