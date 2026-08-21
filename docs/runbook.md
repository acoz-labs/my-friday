# Runbook

## Remove an incorrect stable release alias

First download the stable asset and compare its GitHub asset digest, local
SHA-256, retained commit-suffixed archive, and contained executable against the
release ledger. If the alias is incorrect, identify the one release asset whose
name is exactly `my-friday-darwin-arm64.tar.gz` and delete only that asset by
exact ID. Do not delete or replace the release, tag, `SHA256SUMS`, retained
commit-suffixed archive, or acceptance evidence.

Confirm the permanent latest-download URL no longer serves the bad alias. Then
use the guarded `Backfill stable release asset` workflow with the exact latest
tag, retained source archive name, and accepted executable digest. The workflow
fails closed if the release is no longer latest or any digest/content check is
ambiguous. It is safe to retry after an interrupted upload.

## Recover an interrupted repository creation

Use this only when `my-friday init` reports a retained owner-only transaction
journal. Do not move, edit, or merge either target before diagnosis.

1. Read the reported phase and paths; confirm they are the intended targets.
2. Run `my-friday recover --transaction <reported-journal-path>`.
3. Run `my-friday validate --runtime <runtime-path> --memory <memory-path>`.
4. If recovery refuses because state drifted, preserve the journal and targets
   and ask a maintainer to inspect them. Never manually delete a non-empty
   target.

The command receipt says `Recovered and verified repository pair`, `Rolled
back to the pre-run state`, or `No recovery needed`. Re-running recovery after
validation is safe.

Before promotion, recovery removes only marker-owned partial stages, restores
untouched empty shells and owned parent state, and removes the journal last. A
plan-derived deletion path may briefly remain after interrupted recursive
cleanup. The journal records its exact path and complete-tree proof before
rename; rerunning the same command finishes only that authorized deletion. A
pre-existing foreign collision is preserved and blocks cleanup.
If both the original target and its journal-authorized deletion path are
absent, cleanup had already completed; recovery continues and recreates an
original empty shell with its recorded mode when required.

A pair that still contains `.my-friday/creation-state.json` is incomplete even
when its other files validate. Run the reported recovery command so cleanup can
re-prove the exact pair and remove markers, reservations, and the journal in
their recorded order.
