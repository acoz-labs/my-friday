# Implementation handoff

## Smallest complete outcome

Tier: narrow documentation reconciliation with authority-sensitive review.
One implementation PR makes the approved sequence and full readiness ownership
coherent. No new runtime logic, tests, receipt grammar, candidate workflow,
dependencies or replacement implementation issues.

1. Re-read current main, #101 implementation and existing outcome handoffs.
   Inventory contradictory ordering in the exact reconciliation map. Keep
   concurrent author changes and original authority records intact.
2. Update `docs/product.md` with approved engineering/test/acceptance/release
   boundaries and complete #95 checklist, `docs/development.md` with pin and
   compatibility evidence rules, and `docs/deployment.md` with the deferred
   integrated owner test and preserved nomination/acceptance/publication path.
   Describe implemented verifier behavior accurately; link #101 for unfinished
   tooling rather than copying its schema implementation.
3. Amend active #92 plan sections listed in `03-design.md` explicitly with
   #113/#109 precedence, without resetting its original basis or approval
   history. Annotate #101 handoff if still present; if its concurrent
   implementation has removed the plan, amend the promoted durable sequence
   instead. Preserve original accepted/released legacy compatibility duties.
4. Hand project maintainers exact compact dependency/handoff updates for
   #92/#93/#94/#95/#101/#74/#83/#51 linking the merged reconciliation. Preserve
   existing issue ownership, acceptance/release states and authority tuples.
   This is projection of the reviewed amendment, not another product gate.
5. Validate, independently review and reconcile the exact implementation PR;
   remove this temporary issue plan before the PR leaves draft. After merge,
   verify the issue projections and report documentation completion separately
   from downstream engineering or test readiness.

## Documentation promotion

| Concern | Destination | Action |
| --- | --- | --- |
| Approved sequence, scope and #95 complete readiness checklist | `docs/product.md` | Update; clearly separate current shipped product from approved target |
| Pinned inputs, compatibility and reproducible evidence | `docs/development.md` | Update |
| Integrated test, exact nomination, acceptance and public release boundary | `docs/deployment.md` | Update; coordinate with #101 |
| Existing kernel and owner-acceptance handoffs | Active #92/#101 plans or their promoted docs | Amend only sequence conflicts with provenance |
| Issue dependencies and handoffs | Existing issues | Maintainer projection; preserve original authority |
| Architecture, security, ADR, runbook | Existing documents | No new contract; update only an active contradictory sequence reference discovered during reconciliation |
| Discovery/plan trail | Git history and merged PRs | Retire `docs/discovery/108-integrated-laptop-test-sequence/` after its checklist/sequence is promoted; remove this issue plan; do not archive duplicate permanent documents |

## PR contract and residual constraints

Use `Refs #113`; avoid implying this docs PR implements or closes #92–#95 or
F0. Document the amendment map, validation, concurrent #101 reconciliation and
promotion/removal result. No production deploy. Independent review must catch
weakened acceptance, lost portability requirements and planned behavior
misrepresented as shipped. The execution envelope remains `implementation`.

#106 continues independently; its results feed #93 design. Missing engineering
proof blocks its consuming slice; missing owner judgment blocks acceptance and
release. Neither creates a new product decision for #113. Return upstream only
for an actual conflict requiring a different outcome, unsafe migration or an
acceptance-tool contract change beyond its owning approved plan.
