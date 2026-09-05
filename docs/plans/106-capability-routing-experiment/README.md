# Solution Design: comparative capability-routing experiment

- **Status:** Draft
- **Issue:** #106
- **Planning PR:** #107
- **Repository basis:** 1277b7924f22dcd55128dfa006d0ec0ae3e841e1
- **Execution envelope:** implementation

## Decision

Build a developer-only, offline-first Go experiment under `tools/`, comparing
native progressive skill discovery, lookup/direct loading, and lookup/selective
native workers. Commit the synthetic benchmark and rubric before live results.
Publish an inspectable comparison, including unavailable and invalid cells;
do not change a production capability default.

## Needs Attention

Actual second-harness access is presently unavailable. This is a live-evidence
prerequisite, not a blocker to implementing the runner or reporting partial
results. Missing isolation, native-worker fidelity, or telemetry similarly
invalidates affected claims rather than inviting a simulated substitute.

## Decision Spotlight

- Offline operation is the default. Each live batch requires explicit opt-in,
  existing authorized access, and a declared bounded run manifest; no login,
  credential copying, account changes, new subscription, or spending commitment.
- Essential policy stays in the root. Retrieval supplies candidates, never
  authority; a required worker cannot silently become direct execution.
- Only synthetic fixtures enter model-visible workspaces. Held-out labels and
  the scorer are outside the sandbox, not merely hidden in a repository folder.
- Measure supplied context separately from actual context occupancy and total
  billed tokens. Unavailable telemetry means unknown, never zero or improvement.
- Adopt nothing automatically. A positive experiment only informs B2 design;
  missing second-harness evidence prevents a cross-harness recommendation.
- The implementation envelope covers the offline runner, tests, docs and
  bounded explicitly opted-in evaluation; it excludes releases and installation.

## Plan Map

- [Context](01-context.md)
- [Decision](02-decision.md)
- [Design](03-design.md)
- [Verification](04-verification.md)
- [Implementation handoff](05-handoff.md)

## Final Gate

Independent maintainer review precedes Final status and the one product-authority
approval of this plan and its implementation envelope. Implementation remains
separate from this planning-only PR.
