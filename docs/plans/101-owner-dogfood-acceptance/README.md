# Solution Design: Owner-Operated Capability Dogfood Acceptance

- **Status:** Draft
- **Issue:** #101
- **Planning PR:** Pending
- **Repository basis:** b7d8aa154a1f96b2406dc65fd41e23adb5a03ab7
- **Execution envelope:** through-production

## Decision

Add a new, versioned `capability-workshop-owner-dogfood-v1` acceptance
authority for issue #51. It combines fresh exact-candidate workshop evidence
with one separately authored, allowlisted product-owner receipt. The owner
receipt records hands-on judgment, source/projection/recovery comprehension,
the retention decision, the internal-dogfood claim scope, and the plan to
collect real migration evidence during issue #92 before that portable-assistant
release.

The evidence/product acceptor, product owner, and every implementation pull-
request author are distinct. Existing candidate, artifact, evidence, acceptance
and release integrity checks remain mandatory. The former four-part
`capability-workshop-acceptance-v1` contract remains parseable for historical
audit but cannot authorize a new issue-51 approval after this amendment ships.
Independent-user validation becomes a later product signal and claim gate; it
is not presented as complete and is not a prerequisite for the product owner's internal
assistant migration.

## Needs Attention

- The repository is public and release remains a public GitHub artifact.
  Release notes and product documentation must describe the validation scope
  as owner-operated dogfooding, not imply independently validated usability.
- The current issue-51 candidate `2594147084e8bb6f961971efdb116b02824ecf66`
  and its evidence predate this verifier. They remain valid historical evidence
  but cannot be released under the amended authority. Implementation must
  produce and nominate a new candidate and rerun the complete workshop harness.
- `PRODUCT_OWNER_ACTORS` must exist before acceptance and release preflight and
  name a configured product-owner GitHub actor. The plan records no private identity in
  product code; repository configuration owns the allowlist.

## Decision Spotlight

- **A new authority token, not a reinterpretation.** The two-part owner-dogfood
  bundle has its own prefix and strict grammar. Existing four-part tokens keep
  their historical meaning, preventing a policy change from silently changing
  old evidence.
- **Fresh amended candidate only.** Acceptance evidence and the owner receipt
  must bind the post-implementation commit and its immutable artifact. Earlier
  evidence cannot cross the candidate/artifact boundary.
- **Three roles remain separated, but only one is the user.** The workflow
  evidence author/product acceptor proves the harness; the product owner judges
  the exact candidate; implementation authors supply neither role. Controlled
  maintainer, contributor, service, or controlled assistant accounts are never labelled as
  independent users.
- **External product-owner allowlist.** `PRODUCT_OWNER_ACTORS` is a required,
  comma-separated repository variable injected into both acceptance and
  release verification. This avoids hard-coding a private owner identity while
  ensuring an arbitrary `product-owner` receipt cannot grant authority.
- **Migration evidence is promised now and collected later.** The owner receipt
  binds issue #92 and states that real assistant migration evidence is due before
  #92 release. Actual migration cannot be an F0 prerequisite because F0 is what
  unblocks that implementation.
- **Distribution and evidence claims are separate.** The GitHub artifact may
  remain publicly downloadable. Until independent users are actually
  validated, public-facing copy may claim owner dogfooding and deterministic
  exact-candidate evidence only.
- **Legacy partner authority is audit-only.** The old verifier remains
  available to authenticate historical records, but active issue-51 approval
  and finalization accept only the new owner-dogfood authority. No new
  independent-user recorder or claim system is added in this change.
- **Fail closed on configuration or identity ambiguity.** Missing/empty owner
  allowlist, duplicate roles, an owner who authored any lifecycle-linked
  implementation PR, edited comments, missing migration-plan fields, or a
  mismatched issue/SHA/artifact all refuse acceptance and release.
- **Artifact rollback remains ordinary.** A failed or rejected candidate is not
  published. After release, the prior immutable release remains the rollback
  artifact; no evidence or comment is rewritten or deleted.

## Plan Map

- [Context](01-context.md)
- [Decision](02-decision.md)
- [Design](03-design.md)
- [Verification](04-verification.md)
- [Implementation handoff](05-handoff.md)

## Final Gate

The plan becomes `Final` only after independent maintainer review confirms that
the new authority cannot reuse pre-amendment evidence, all role-separation and
release-time checks are symmetric, the migration promise is non-circular, and
no blocking finding remains. The product authority approved the outcome at discovery Gate 1;
the exact planning head still requires the repository's ordinary Phase 2
authority record before merge.
