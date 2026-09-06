# Decision

## Competing approaches

| Approach | Verified attack / consequence | Decision |
| --- | --- | --- |
| Retain F0 publication before all kernel engineering | #92 README and handoff explicitly block the selected integrated-first outcome | Reject; contradicts approved #109 |
| Edit only issue dependencies | #92 migration and verification sections and #101 final handoff still encode the earlier sequence | Reject; leaves competing authority |
| Add an engineering-candidate service, new receipt or gate | Existing exact SHA/artifact/check evidence already supports reproducibility; nomination does not imply acceptance | Reject; unnecessary mechanism and new authority risk |
| Reconcile durable sequence, affected handoffs and compact issue links | Preserves existing technical work and makes the limited precedence amendment auditable | Select |

The substantive safety risk is calling untested or incompatible bytes ready.
Pinning alone does not solve it: the consuming issue must bind successful
contract tests and rollback evidence to those inputs. Conversely, public
publication alone is not a substitute for integration compatibility tests.

## Consequences and rejected scope

The amended dependency permits engineering work, not false completion. #101
retains conditional release authority and complete workshop obligations, but
its release step is deferred until the integrated owner test. #92 retains
released legacy fixture coverage and supported platform requirements, with
additional pinned unreleased engineering characterization as needed in its
implementation. No release validator is weakened.

Do not rewrite historical approvals or relabel old receipts. Amend active
handoff text explicitly with #113/#109 provenance, retaining the original
technical contracts and authority links. Do not mass-replace every occurrence
of “released”: many refer to compatibility guarantees that remain true.

No new ADR is needed: this is a delivery-order amendment with immutable issue,
discovery and planning provenance, not a new architecture choice.
