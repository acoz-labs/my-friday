# Context

The owner wants a complete portable assistant to test in a personal laptop
account. Existing handoffs require F0 acceptance and publication before kernel
implementation, preventing that sequence. #113 reconciles that authority; it
does not implement the assistant or claim successful testing.

## Evidence ledger

Repository claims below use full basis
`949a4171145660d3463994d6f339898bcf3d7315`.

| Evidence | Located finding |
| --- | --- |
| [Issue #113](https://github.com/acoz-labs/my-friday/issues/113) | Open, Project 33 Solution Design; O1 from approved discovery #109 |
| [Discovery #109](https://github.com/acoz-labs/my-friday/pull/109#pullrequestreview-5123506311) | Exact approved head `d0821faa76b6928a6db389efd8867c217b913582`; merged `de755a31994e79644b4d19c3db89d84c1bb4c6f2` |
| `docs/discovery/108-integrated-laptop-test-sequence/README.md` | Selected engineering-first sequence, full readiness checklist and existing-outcome map; its historical awaiting-authority text is superseded by the exact-head review |
| `docs/plans/92-canonical-assistant-kernel/{README,01-context,03-design,04-verification,05-handoff}.md` | F0 released-input assumptions affect implementation ordering, migration characterization, candidate construction and final release prerequisites |
| [Plan #103](https://github.com/acoz-labs/my-friday/pull/103), `docs/plans/101-owner-dogfood-acceptance/` | Merged at repository basis; #101 owns new owner receipt/bundle and acceptance/finalizer integration; its handoff currently ends with release unblocking #92 |
| `docs/product.md` / `docs/deployment.md` | Older product sequence and three-person workshop acceptance copy; distinguish approved future direction from currently implemented verifier behavior |
| `.github/workflows/nominate-artifact.yml`, `bin/nominate-release-candidate`, `bin/release-gate` | Existing artifact nomination builds exact main bytes and checks CI; nomination is distinct from acceptance, associates explicit PR reference lines, and can mutate issue candidate state |
| `bin/record-product-acceptance`, `bin/finalize-release`, `bin/test-capability-workshop-evidence` | Existing acceptance/release authority and issue-specific evidence grammar; #113 must not change these |
| `docs/architecture.md`, `go.mod`, `bin/ci` | Native Go command with manifest-owned source/projection lifecycle; CI includes contract fixtures, vet, race tests and builds |
| Live #92/#93/#94/#95/#101/#74/#83/#51 bodies | Original outcomes and provenance remain; stale dependency wording must be reconciled without replacing original authority tuples |

## Assumptions ledger

- Pinned reviewed F0 contracts can be tested as engineering inputs without a
  public release. Each consuming slice must demonstrate this, not assume it.
- #101 may complete independently while this documentation reconciliation is
  reviewed. Its current implementation/durable docs must be re-read before edits.

## Unknowns ledger

No unresolved fact blocks this documentation design. Cross-host compatibility,
complete native/platform evidence, #106 findings, second-harness support,
governed-memory migration and owner retention are unproven implementation
results owned by #92–#95. A failure blocks the affected integration/readiness
claim. Acceptance-tool incompatibility must return to its owning reviewed
technical change; this plan does not silently fix it.

## Decisions ledger

D1: use an explicit documentation amendment and existing issue ownership.
D2: pin engineering inputs and keep them separate from acceptance authority.
D3: assign complete checklist/evidence assembly to #95.
D4: preserve release obligations, privacy and role separation.
D5: perform one reviewed docs implementation, with no new dependency, schema,
workflow, UI, secret or runtime configuration.

## Scope and actors

Product authority supplies intent and actual later judgment; contributors
reconcile documentation and record engineering proof; maintainers independently
review and project maintainers reconcile issue dependencies. #113 is complete
when reviewed repository guidance and linked handoffs agree. Existing
capability, memory, kernel, integration and acceptance implementations remain
their original issues. No product-experience redesign is needed.
