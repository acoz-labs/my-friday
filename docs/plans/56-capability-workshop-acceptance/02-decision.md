# Solution Decision

## Decision Drivers

The solution must bind real behavior to one executable, remain independent of
the contributor, preserve ambient state and credentials, cover every #51 trust
boundary, produce machine-verifiable release authority, and avoid weakening the
mature issue-4 acceptance contract.

## Competing Approaches

1. Extend `accept-installed-codex-baseline` to accept issue 51 and conditionally
   run capability steps.
2. Add an issue-51 supervisor that reuses reviewed helper primitives and defines
   a distinct evidence schema/verifier route.
3. Accept unit/CI evidence plus a manual prose comment.

## Adversarial Comparison

Conditionalizing the issue-4 supervisor couples unrelated journeys and risks
cross-issue evidence acceptance. A separate supervisor duplicates orchestration
but keeps authority and failures legible. Manual prose cannot prove candidate,
artifact, cleanup, or fresh-task behavior and is incompatible with the release
gate. Therefore approach 2 is the smallest safe path.

## Selected Approach

Create `bin/accept-capability-workshop` and an issue-51 typed evidence family,
using the existing artifact retrieval, runner, APFS graph, protected snapshots,
Codex auth-by-path, PTY, and cleanup helpers where their contracts already fit.
Add issue-aware verification and finalization routing without making schemas
interchangeable. Confidence is high because the isolation and release substrate
is shipped; the new risk is journey completeness, handled by fixture and
fault-matrix tests.

## Decisions Ledger

| Decision | Rationale | Evidence |
|---|---|---|
| Separate issue-51 command/schema | Prevent cross-authorization and keep issue 4 stable | Current hard scope and verifier contracts |
| Fixed useful fixture plus live builder | Repeatable assertions while proving natural source authoring | #51 product-design journey |
| Driver owns mutation tokens | Builder cannot activate its own work | #54 authorization design |
| Digest/redaction evidence | Prove complete diff and behavior without publishing bodies or private paths | Existing evidence-v1 boundary |
| New nomination after merge | Acceptance code is part of the executable/helper closure | Artifact-v1 release policy |
| Through-production envelope | Existing artifact release is reversible and independently gated | `release-artifact.yml`, release ledger |
