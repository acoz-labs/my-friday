# Solution Design: integrated-test-first delivery authority

- **Status:** Draft
- **Issue:** #113
- **Planning PR:** Pending
- **Repository basis:** 949a4171145660d3463994d6f339898bcf3d7315
- **Execution envelope:** implementation

## Decision

Reconcile existing delivery documentation and handoffs to permit reviewed,
pinned, CI-verified unreleased engineering integration before personal-account
owner testing. #95 owns the complete laptop-readiness checklist. Existing
implementation owners and exact-candidate acceptance/release rules remain.
This implements approved discovery #109 outcome O1 without a new runtime or
acceptance mechanism.

## Needs Attention

No blocking design unknown. Independent maintainer review and final plan
approval remain required. Actual integration compatibility, native platform
proof, real harness execution, migration rehearsal, owner judgment, and public
release are later delivery evidence; this documentation change certifies none.

## Decision Spotlight

- A dependency need not be publicly released for engineering use, but its
  exact commit, artifact digest, successful checks, compatibility evidence and
  rollback source must be recorded. A floating branch or failed check is not
  an eligible substitute.
- Migration rehearsals may use pinned, characterized unreleased F0 inputs;
  shipped legacy compatibility remains required. This grants no arbitrary
  schema acceptance or overwrite authority.
- Engineering-ready, laptop-test-ready, owner-accepted and released are
  separate claims. One owner session can cover multiple outcomes only through
  their own exact-candidate evidence contracts; no generic approval receipt.
- #95 inherits the whole approved checklist, including Linux obligations,
  real capability execution, governed memory, reversible migration and the
  second-real-harness full-vision requirement. Partial success is not complete
  portability. External-user validation remains #102.
- The envelope is `implementation`: documentation and handoff reconciliation
  only. Installation, personal-account testing, publication and production
  activation are outside this issue's authority.

## Plan Map

- [Context and ledgers](01-context.md)
- [Decision and alternatives](02-decision.md)
- [Sequence and ownership design](03-design.md)
- [Verification and readiness](04-verification.md)
- [Implementation handoff](05-handoff.md)

## Final Gate

The maintainer reviews the complete pack before it is marked Final. Final
product-authority approval must bind the actual planning head and the
`implementation` envelope; neither discovery approval nor CI supplies it.
