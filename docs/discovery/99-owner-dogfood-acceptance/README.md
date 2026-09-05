# Discovery: owner-operated dogfood acceptance authority

- **Status:** Draft
- **Discovery issue:** #99
- **Discovery PR:** #100
- **Repository basis:** `616395439a953100c6c165ccf56ea314c2dd0307`
- **Recommended decision:** approve
- **Gate 1:** awaiting-authority
- **Confidence:** High
- **Private evidence:** none

## Decision sought

What evidence should authorize My Friday's first owner-operated Alfred dogfood
release without blocking the migration on recruiting unrelated design partners?

The decision concerns product validation, not artifact integrity. The release
must still be built, tested, nominated, accepted, and published as the same
immutable bytes through the existing release machinery.

## Audience and critical tasks

The immediate user is the product owner operating and migrating his own Alfred
installation. He needs a stable released instruction-only capability contract
before the repository-first portable assistant kernel in #92 can migrate real
state onto it.

Independent technical users remain a later audience. Their use is valuable
evidence of comprehension and product fit, but recruiting them is not part of
the product owner's critical path for dogfooding his own system.

Critical tasks are:

1. preserve an exact, inspectable capability candidate and artifact;
2. prove lifecycle, isolation, recovery, cleanup, and ambient-state
   preservation automatically on the nominated bytes;
3. obtain explicit product-owner judgment about retaining or removing that
   exact result;
4. have an authorized acceptor, separate from implementation and the product
   owner, record and verify the release decision;
5. release the stable F0 contract so #92 can migrate Alfred and collect real
   operational dogfood evidence; and
6. avoid claiming that owner dogfooding proves independent usability.

## Evidence

- `bin/accept-capability-workshop` already exercises creation without
  activation, complete source review, deterministic checks, install/use,
  enhancement, collision and drift refusal, interruption recovery, reversal,
  cleanup, and ambient-state preservation against an immutable artifact.
- `bin/verify-capability-workshop-evidence` re-fetches typed provisional and
  final issue evidence and binds it to issue, candidate, artifact, author, and
  cleanup proof.
- `bin/record-product-acceptance` already restricts acceptance actors and
  refuses the implementing pull-request author as sole acceptor.
- `bin/verify-capability-workshop-acceptance` additionally requires one
  product-owner and two design-partner receipts from three distinct GitHub
  actors. This is the only part that makes unrelated participant recruitment a
  release prerequisite.
- `docs/product.md` correctly says interest is not proof of independent
  usability, but its current continue signal conflates owner dogfooding with
  external validation.
- The repository and its releases are public. Therefore an owner-dogfood
  release cannot honestly be described as independently validated merely
  because it is downloadable.
- Product authority approved removing the two-design-partner blocker on
  2026-09-05 while retaining automated evidence and owner acceptance.

## Assumptions

- The product owner's immediate purpose is to migrate and operate Alfred, not
  to assert broad market validation.
- An allowlisted product-owner identity can remain repository configuration
  rather than a personal identity hardcoded in product source.
- Real migration evidence can be recorded after F0 release without becoming a
  circular prerequisite for the release that enables migration.
- Historical authority formats should remain verifiable for audit but need not
  authorize new releases after the policy changes.

## Unknowns

- Independent users' comprehension, retention, and support burden remain
  unknown until external validation occurs.
- The first real multi-machine migration may expose usability work not present
  in the deterministic workshop.

Neither unknown prevents owner-operated dogfooding. Both constrain the claims
that may be made afterward.

## Competing options

### A. Retain the three-person release prerequisite

This supplies external comprehension evidence before release, but makes the
owner's migration depend on recruiting two unrelated people. It delays the
stronger real-world evidence the migration itself can produce and is rejected.

### B. Accept automated workshop evidence alone

This removes all human friction, but loses product judgment and permits a
technically green candidate to advance without anyone deciding whether the
experience is worth retaining. It is rejected.

### C. Owner judgment plus independent release recording

Require fresh automated exact-candidate evidence, one allowlisted
product-owner decision, and an authorized release acceptor who is distinct
from both the owner and every implementation author. Record a new versioned
authority rather than changing the meaning of the historical bundle. This is
selected.

### D. Release privately or as an untracked local build

The public repository has no private GitHub Release, and an untracked build
would not satisfy #92's need for a stable immutable input contract. This is
rejected.

## Decision

Select option C.

Create a new versioned owner-dogfood acceptance authority that combines:

- fresh final automated workshop evidence produced by the amended candidate;
- one complete product-owner decision bound to the same issue, candidate, and
  artifact;
- an externally configured product-owner allowlist; and
- an authorized product acceptor who differs from the product owner and from
  every lifecycle-linked implementation author.

Do not reinterpret the existing four-part, three-person authority. Preserve it
for historical verification. A new candidate containing the amended verifier
must be built, nominated, and exercised; the old candidate cannot borrow code
or policy added after its SHA.

The resulting GitHub Release remains public and stable. Product copy and
release notes must label the evidence accurately as owner dogfooding, not
independent-user validation. External design-partner use becomes a later
validation signal and claim gate, not a prerequisite for Anthony's migration.

The #92 migration records operational evidence after F0 release. That evidence
can guide changes or external validation, but it is not required to release the
contract needed to perform the migration.

## Success and stop signals

Continue when:

- the new candidate passes CI and emits fresh typed workshop evidence;
- the configured product owner approves that exact candidate and artifact;
- a separate authorized acceptor records acceptance;
- release verification re-fetches every authority and preserves the same-byte
  chain; and
- F0 is published with claims limited to owner dogfooding.

Change direction when the real Alfred migration exposes repeated recovery,
comprehension, or maintenance failures.

Pause when product-owner configuration is absent or ambiguous, owner and
acceptor are not distinct, any implementation author attempts to accept, or
fresh evidence cannot be produced by the amended candidate.

Stop broad-distribution claims when independent users do not understand or
retain the product. Do not stop owner dogfooding merely because that external
evidence has not yet been collected.

## Candidate outcome map

### O1 — Versioned owner-dogfood acceptance contract

- Disposition: selected
- Outcome: F0 can be accepted and released for owner-operated Alfred
  dogfooding using fresh automated evidence, an allowlisted owner decision,
  and a separate authorized acceptor.
- Acceptance: tests prove exact binding, identity separation, configuration
  refusal, compatibility behavior, new-candidate evidence, and unchanged
  release integrity; durable product/deployment/security documentation states
  the scoped claim.
- Dependencies: existing F0 workshop and immutable artifact release machinery.
- Sequence: first; replaces the manually created, non-authorizing #97 brief
  with a gate-backed delivery issue.

### O2 — Independent-user validation cohort

- Disposition: deferred
- Outcome: independent technical users exercise and retain a future immutable
  candidate so My Friday can make evidence-backed external-usability claims.
- Acceptance: separately shaped participant, comprehension, correction,
  retention-interval, and evidence criteria; controlled accounts never count
  as independent users.
- Dependencies: stable owner-dogfood release and findings from the real Alfred
  migration.
- Sequence: after O1 and initial #92 dogfooding; it must not block them.

## Privacy and evidence handling

No private evidence is required. Tracked files and GitHub records contain only
repository paths, public issue/PR identifiers, policy decisions, and typed
redacted release authorities. Secret values, private paths, prompts, capability
contents, credentials, and model transcripts remain excluded.

Product-owner allowlisting is runtime repository configuration. Product source
defines the generic contract and never hardcodes a private person's identity.

## Decision Spotlight

- **Human judgment remains:** owner dogfooding is not automation-only; the
  product owner still decides retain or remove for the exact candidate.
- **Acceptance remains independent of implementation:** a separate authorized
  acceptor records the release, and no implementation author may do so.
- **Policy changes create new bytes:** a new authority version and a fresh
  candidate prevent retroactive acceptance of an older artifact.
- **Public availability is not product validation:** release notes scope the
  claim; external usability remains unknown until independently tested.
- **Migration evidence is deliberately non-circular:** #92 records the first
  operational dogfood evidence after the F0 release it depends upon.
- **Identity is configured, not embedded:** the generic verifier consumes an
  allowlist supplied by repository configuration and fails closed when absent.

## Gate 1

The final candidate awaits product authority and an authorized approval on the
exact pull-request head.
