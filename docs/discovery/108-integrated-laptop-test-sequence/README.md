# Discovery: integrated laptop testing before owner acceptance

- **Status:** Final
- **Discovery issue:** #108
- **Discovery PR:** pending
- **Repository basis:** f1e6f28fe2db6f4d3ec0e0292e3f03f87e7297f2
- **Recommended decision:** approve
- **Gate 1:** awaiting-authority
- **Confidence:** High
- **Private evidence:** none

## Decision sought

Allow reviewed implementation and integration of the complete portable assistant
before the product owner performs personal-account laptop testing. Replace the
F0-release-before-implementation dependency with a pinned, CI-verified engineering
dependency. Do not replace owner judgment, mark an untested candidate accepted,
or authorize public release before the existing acceptance gates pass.

The target is an installable integrated assistant, not a routing benchmark,
documentation-only plan, unsupported-only harness adapter, or partial kernel.

## Audience and critical tasks

The product owner wants to install and test the whole system in a personal
laptop account after its implementation and automated verification are ready.
Maintainers need reproducible inputs and a truthful distinction between code
completion, test readiness, owner acceptance, and public release.

The repository remains the portable source of configuration, governed memory,
and capabilities. Hosts and optional virtual machines remain replaceable
installation targets; this amendment does not introduce a VM requirement.

## Evidence

- Merged plan #96 for #92 explicitly sequences both implementation and release
  after accepted/released F0 (#74/#83).
- Approved discovery #100 removes independent-partner recruitment but still
  expects an owner-approved F0 release before #92 migration.
- Open plan #103 implements that owner-dogfood authority with a fresh candidate,
  an allowlisted owner receipt, separate acceptance, and immutable release
  evidence. Its through-production envelope does not remove those prerequisites.
- #93 and #94 own portable capability and governed-memory outcomes; #95 owns
  complete baseline integration. Their implementation remains unfinished.
- #106 supplies comparative routing evidence before B2 design; it does not
  deliver those runtime outcomes or prove the complete personal-account journey.
- The current requested outcome is complete implementation followed by owner
  laptop testing, not sequential owner tests of each unfinished foundation.

## Assumptions and unknowns

Assumptions: reviewed, immutable candidate inputs with recorded compatibility
checks can support engineering integration without being publicly released;
the current release machinery remains the authority for final publication.

Unknowns: complete migration compatibility, personal-account integration,
cross-host behavior, and owner retention are not yet proven. Engineering must
verify feasibility and exact candidate handling in the affected solution plans.
No claim of external-user validation or automatic migration safety is made.

## Options and decision

1. Retain a separate owner-approved F0 release before any B1 implementation.
   Preserves the old sequence but prevents the requested integrated-first test.
2. Treat automation as owner acceptance or publish an unaccepted release.
   Rejected: it fabricates judgment or bypasses immutable release authority.
3. **Selected:** build on pinned, tested unreleased inputs, then offer one complete
   integrated candidate for personal-account testing. Preserve all subsequent
   acceptance and public-release gates.

Option 3 is not an untracked local build. Each dependency and integrated
candidate records its commit, artifact identity/digest, checks, compatibility
evidence, and rollback source. Candidate artifacts are not labelled released,
accepted, or independently validated. No production tag, GitHub Release, or
default production activation is created as a convenience for integration.

## Three distinct completion boundaries

### Engineering readiness

Reviewed code and tests may merge while owner acceptance remains pending.
Dependent work uses exact pinned inputs whose relevant contract and safety
checks pass. Missing tests or incompatible contracts block the dependent slice;
absence of an owner receipt does not block unreleased engineering integration.
Ordinary Solution Design, contributor review, and CI requirements remain.

### Laptop-test readiness

The complete integrated candidate must satisfy every item below before asking
the owner to test. An explicitly unsupported required behavior fails readiness.

- One user-owned private Git repository governs `config/`, `memory/`, and
  `capabilities/`; no private content is copied into the public product repo.
- Personal-account setup instructions identify prerequisites, exact candidate
  verification, repository binding, declared paths/effects, launch, diagnosis,
  removal, and recovery. Installation does not depend on an existing dedicated
  account's private paths, sessions, credentials, or runtime projections.
- Install/restore, update, rollback, remove, interruption recovery, and ambient
  preservation are exercised. Removing a projection preserves canonical source
  and durable data; replacement never overwrites unrelated account state.
- Fresh-task synchronization, concurrent writers, divergence/conflict refusal,
  offline/stale behavior, and recovery are tested with reproducible evidence.
- Capability packages support real end-to-end discovery and execution through
  the selected supported harness path, with source/projection fidelity and
  reversible activation. Essential identity, policy, authority and memory
  triggers remain available to the main agent. Lookup failures, unsupported
  requirements, and subagent summaries cannot silently weaken task authority.
- Governed memory captures observations and proposals, separates promotion,
  resolves/supersedes claims, and returns attributed current context in a fresh
  task. Shared-repository refresh and conflict behavior preserve governance.
  #94 owns migration parity for identifiers, provenance, validity/history,
  conflicts and pending proposals, plus explicit checks for atomic transitions,
  sensitivity/secret rejection, schema conformance and crash recovery. Copying
  host-bound scripts or relying on a local lock is not multi-host proof; #92
  owns safe remote stewardship. Its solution design supplies the mechanisms.
- A reversible migration rehearsal preserves originals, inventories declared
  state without exposing secrets, validates converted capabilities/memory,
  proves rollback, and refuses ambiguous or destructive overwrite. Private
  migration records remain private; only sanitized proof enters product docs.
- Credentials remain references and host-local approved access bindings, never
  secret values in Git, packages, artifacts, logs, or evidence. Missing access
  yields a precise setup requirement, not another account's implicit fallback.
- Existing B1 platform obligations remain, including Linux amd64/arm64 and
  required native restore/synchronization/recovery evidence. A successful Mac
  demonstration does not silently reduce the approved portability contract.
- The nominated test candidate includes all required implementation merges,
  passing checks, exact artifact verification, known limitations, and a guided
  personal-account test/rollback script. Actual owner judgment remains pending.

### Owner acceptance and public release

The owner tests the exact integrated candidate in the personal account and
supplies actual judgment. The evidence must satisfy every relevant issue's
unchanged exact-candidate, artifact, implementation-set, identity-separation,
and release rules. A single test session may exercise several outcomes, but
does not invent a generic receipt that bypasses their verifiers.

Fresh owner receipts cannot be authored in advance or inferred from approval
to implement. Any necessary acceptance-tool compatibility change requires its
own reviewed technical reconciliation. No earlier evidence is borrowed across
candidate boundaries. Independent-user claims remain deferred to #102.

## Existing-outcome reconciliation

| Existing owner | Sequence adjustment; retained responsibility |
| --- | --- |
| #101 / plan #103 | Implement owner-dogfood authority and tests; defer hands-on receipt and publication until the integrated test. Through-production is conditional authority, not a deadline to publish first. |
| #74 / #83 / #51 | Preserve complete workshop evidence and release integrity; supply pinned verified engineering inputs while acceptance/release remain pending. |
| #92 / merged plan #96 | Replace F0 acceptance/release as an implementation prerequisite with verified pinned F0 compatibility; retain kernel, platform, synchronization, recovery and migration obligations. |
| #93 | Reconcile routing findings from #106 in Solution Design; deliver the actual portable package and supported execution path, not benchmark infrastructure alone. |
| #94 | Implement governed memory on verified kernel/package contracts and prove fresh-task shared context. |
| #95 | Own complete integrated test-readiness checklist and candidate evidence; distinguish owner testing from subsequent public/external claims. |
| #102 | Remains later independent-user validation, never a blocker for this owner test. |

Do not create replacement implementation issues or mark these outcomes done
from this discovery. Preserve existing authority tuples and add explicit links
to the reviewed sequencing amendment. Reconcile affected issue dependencies,
open/merged plan handoffs, and current product/deployment documentation through
reviewed changes before dispatching work that contradicts their old sequence.

## Candidate outcome map

### O1 — Reconcile integrated-test-first delivery authority

- Disposition: selected
- Outcome: Existing foundation, kernel, capability, memory and integration outcomes share an explicit engineering-before-owner-test sequence without duplicate implementations or weakened acceptance.
- Acceptance: Reviewed reconciliation of #92/#93/#94/#95/#101 and affected F0 handoffs permits pinned CI-verified unreleased integration, assigns the complete laptop-test checklist to #95, and preserves exact-candidate personal-account judgment and all public-release gates.
- Dependencies: Existing approved discovery and plans; technical review of compatibility and candidate handling; no dependency on recruiting independent users.
- Sequence: Now; one bounded authority/documentation reconciliation before otherwise-blocked dependent implementation, while #106 continues independently.

## Success, change, pause and stop signals

Continue when dependencies are pinned and verified and each outcome advances
toward the complete checklist. Change the engineering plan when compatibility
or measured execution contradicts assumptions. Pause the affected slice for
missing required platform/harness proof or unsafe migration. Pause acceptance
and release until real owner judgment and required evidence exist. Stop any
attempt to call a partial implementation laptop-ready, invent receipts, copy
private state into public artifacts, or overwrite original state.

## Privacy and promotion

This pack contains public-safe product constraints only, no private evidence.
After authority, promote the sequence and checklist into the existing product,
development and deployment documentation through reviewed reconciliation, then
retire this temporary discovery pack under the ordinary lifecycle. This PR
changes no code, account permissions, installed runtime, acceptance, or release.

## Gate 1

This Final candidate awaits exact-head authority recording and independent
maintainer review. The standing implementation request supplies the intended product outcome;
this document makes its changed sequence auditable rather than fabricating
completed personal-account testing or claiming that the owner inspected a
particular commit. No additional product choice is proposed.
