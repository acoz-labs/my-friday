# Design

## Delivery sequence and authority

Reviewed inputs → pinned engineering integration → complete #95 laptop-test
readiness → actual exact-candidate owner test → issue-specific acceptance →
public release under existing gates. These are evidence boundaries, not new
Project statuses or a new approval workflow.

For each engineering dependency, the consuming implementation handoff records
the repository and full commit, retained artifact identifier and SHA-256,
successful check URLs for that commit, relevant contract/version compatibility
results, and known-good rollback source with verification. If a dependency is
source-only, record its exact tree and state that no separate executable is
consumed; the integrated executable still needs its artifact identity/digest.
Use existing PR evidence and artifact references, not a new machine schema.

CI failure, unknown contract, missing artifact, mismatched digest, unresolved
divergence or unavailable rollback blocks the affected dependent slice. Refresh
pins only through reviewed evidence; do not follow `latest` or a moving branch.
Expired artifacts require a fresh candidate and evidence, not a claimed reuse
of inaccessible bytes. Do not invoke nomination merely to document an
engineering pin: it changes candidate state. Formal integrated nomination uses
the existing workflow after all required implementation merges are included.

## Exact reconciliation map

| Owner | Required reconciliation | Retained responsibility |
| --- | --- | --- |
| #74/#83/#51 | Add engineering handoff with pins/checks/compatibility and link the amendment; acceptance/release remain open | Complete workshop, isolation, cleanup, immutable evidence and release integrity |
| #101 | Separate completed verifier implementation from deferred owner receipt/publication; remove publication as the event that alone unblocks #92 engineering | Approved #103 receipt schema, actor separation, fresh candidate, finalizer replay; no code duplication in #113 |
| #92 | Amend README/context prerequisite, migration description and command preconditions, verification build order and handoff to permit pinned characterized unreleased F0 engineering inputs | Kernel, private Git stewardship, released legacy compatibility, exact confirmation, refusal, lossless rollback, native macOS/Linux amd64/arm64 proof and real migration before release |
| #93 | Pin verified B1/F0 contracts; carry #106 findings into its own Solution Design | Real package discovery/execution, explicit fidelity/authority, reversible activation and second real harness proof for harness-independent claims |
| #94 | Pin verified kernel/package contracts | Governance, fresh-task attributed context, atomic transitions, sensitivity/secret rejection, schema/crash recovery and migration parity |
| #95 | Own the checklist below and integrated candidate/implementation-set evidence | Complete personal-account journey; keep independent-user claims deferred to #102 |

The #92 change permits engineering fixtures/rehearsals of the exact declared
F0 contract, not an unrestricted runtime migration flag. Its implementation
must preserve released-input tests and refuse unknown versions, ambiguous
mapping and destructive overwrite. If the pinned contract cannot be supported
within its approved schema/migration design, return that technical discrepancy
to #92 review before consuming it. F0 acceptance/publication can occur after the
integrated owner test; final release prerequisites are not waived.

## Complete #95 laptop-readiness checklist

Promote this checklist into `docs/product.md`, with engineering evidence details
in `docs/development.md` and the test/acceptance journey in `docs/deployment.md`.
Each item needs located candidate-bound proof before the readiness claim.

1. One user-owned private Git repository governs `config/`, `memory/` and
   `capabilities/`; private content stays out of public source/evidence.
2. Personal-account instructions cover prerequisites, exact artifact
   verification, binding, declared paths/effects, launch, diagnosis, removal
   and recovery without depending on another account's sessions or paths.
3. Install/restore, update, rollback, remove, interruption recovery and ambient
   preservation pass; removal preserves canonical source and durable data.
4. Fresh-task synchronization, concurrent writers, divergence/conflict
   refusal, explicit offline/stale behavior and recovery have reproducible proof.
5. Packages execute end to end on the chosen supported harness with source/
   projection fidelity, reversible activation and essential identity, policy,
   authority and memory triggers retained by the main agent. Lookup failure,
   unsupported requirements or subagent summaries cannot weaken authority.
   The same package must execute on a second real harness before claiming
   harness-independent execution; fixtures and missing #106 runs do not count.
   Missing proof remains visible as an outstanding full-vision obligation.
6. Memory demonstrates capture/proposal/promotion separation, atomic conflict
   resolution/supersession and attributed current context in a fresh task.
   #94 proves IDs, provenance, validity/history, conflicts and pending-proposal
   migration parity, atomicity, sensitivity/secret rejection, schema conformance
   and crash recovery; #92 proves safe multi-host remote stewardship.
7. Reversible migration inventories declared state privately, preserves
   originals, validates converted capabilities/memory, proves rollback and
   refuses ambiguity/overwrite. Publish only sanitized proof.
8. Credentials remain references with host-local approved bindings. Missing
   access is an explicit setup requirement, never another account's fallback.
9. Existing B1 Linux amd64/arm64 and native restore/synchronization/recovery
   duties remain. A successful Mac test does not establish full portability.
10. The nominated integrated candidate contains every required implementation
    merge, passing checks, verified artifacts, limitations and a guided
    personal-account test/rollback script. Owner judgment is still separate.

An unsupported required behavior fails readiness. A limited supported-harness
laptop test cannot be presented as complete harness-independent portability.
#95 keeps any such outstanding full-vision evidence visible rather than closing
the outcome from a partial demonstration.

## Interfaces, security, recovery and observability

No API, database, permission, secret slot or runtime state changes. Existing
artifact nomination and per-issue receipts remain authoritative. A shared test
session does not merge receipt schemas, transfer evidence between candidates
or authorize a later verifier to accept an older binary. Bind every applicable
issue's complete current implementation PR set to the integrated candidate;
use explicit top-level PR references as the existing nomination contract needs.

Repository docs carry sanitized IDs/digests/check links and known limitations;
private migration records stay private. Revert a bad documentation amendment by
a reviewed revert and reconcile its issue links; do not delete historical
authority or infer that a docs revert rolls back an installed assistant.
