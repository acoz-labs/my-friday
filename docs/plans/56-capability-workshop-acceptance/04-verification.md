# Verification And Release Design

## Test Strategy

- Extend `bin/test-acceptance-contract` and `bin/test-acceptance-evidence` with
  issue-51 schema, cross-schema rejection, edited-comment, wrong actor,
  candidate/artifact/PR mismatch, provisional/failure, and partner-receipt
  matrices.
- Add shell/helper fixture tests for every CLI state and transition, builder
  mutation refusal, no early projection, exact tokens, stale plans,
  collision/drift preservation, complete reversal, migrations, and faults.
- Add cleanup faults at APFS attach/detach, process stop, auth copy/removal,
  source/projection mutation, evidence publication, and final root removal.
- Keep deterministic CI network/model-free; run the real Codex journey only in
  independent Darwin acceptance.
- Run `bin/ci`, shell validation, Go race tests, acceptance contract/evidence,
  and release preflight tests.

## Red/Green Sequence

1. Prove current issue-4 supervisor rejects 51 and release verification lacks a
   valid issue-51 schema.
2. Add failing issue-51 typed evidence and cross-authorization tests.
3. Add failing deterministic workshop fixture and lifecycle transcript tests.
4. Implement APFS/native supervisor slices through authoring, lifecycle,
   recovery, migration, and cleanup.
5. Add live PTY builder/invocation checks and partner receipts.
6. Integrate product acceptance/finalization and run full CI/fault matrices.

## Acceptance Evidence

Automated PR evidence proves schemas, fixture states, denials, faults, cleanup,
and release routing. Independent acceptance downloads the nominated artifact,
runs the complete issue-51 supervisor on Apple silicon, and publishes the strict
provisional/final pair only after cleanup. The product owner and two design
partners then run the same approved workshop against that artifact and publish
typed receipts. No screenshot or UI baseline is required; terminal output is
plain-text protocol evidence.

## Rollout

Execution envelope is `through-production`:

1. Merge the reviewed implementation and mark the prior nomination historical.
2. Nominate the new merged main SHA; build one immutable ARM64 executable.
3. Run independent issue-51 acceptance and verify final evidence.
4. Record the product-owner and two design-partner receipts.
5. Record independent product acceptance for issues 51 and 56 as required by
   the implementation PR set.
6. Run `release-artifact.yml` with the exact SHA, artifact authority, issues,
   and summary; no rebuild occurs.
7. Download the GitHub Release anew and verify archive checksum, contained
   executable digest, version/help, and latest stable URL.
8. Use `complete-software-release` for both issues.

## Rollback And Recovery

Before release, revert the acceptance implementation and nominate a new
artifact; never reuse or relabel the historical candidate. After release, a bad
artifact is superseded by a new release. Existing user roots are never mutated
by acceptance or publication. Interrupted acceptance follows the existing
marked-run recovery runbook; ambiguous APFS or credential state is preserved.
Successful acceptance/finalization evidence is resumable without rerunning live
Codex or redeploying an artifact.

## Release Prerequisites

- Current-user mode-0600 one-link Codex auth file exists and is injected only by
  absolute path.
- Apple silicon/APFS and the reviewed `sandbox-exec` behavior are available.
- Independent acceptor differs from the sole implementation contributor.
- One product-owner and two distinct design-partner receipts bind the candidate.
- No unresolved P0/P1 review finding or protected-state cleanup ambiguity.

## Production Readiness Preflight

- **Secrets:** no new GitHub or runtime secret; acceptance copies an existing
  local auth file through the reviewed no-content path boundary.
- **Candidate:** full main SHA, artifact-v1 run/id/name/executable digest, helper
  closure, and implementation PR set are exact.
- **Deploy/activation:** artifact release publishes accepted bytes only;
  publication does not activate or migrate user roots.
- **Verification:** release workflow, fresh download, checksums, executable
  digest, version/help, stable URL, and immutable issue ledger are executable.
- **Rollback:** supersede with a newly accepted artifact; retain prior release
  and compatible binaries.
- **Receipt:** final GitHub Release/tag and issue comments bind SHA, artifact,
  acceptance, partner receipts, included issues, and release URL.
