# Runbook

## Recover an interrupted repository creation

Use this only when `my-friday init` reports a retained owner-only transaction
journal. Do not move, edit, or merge either target before diagnosis.

1. Read the reported phase and paths; confirm they are the intended targets.
2. Run `my-friday recover <reported-journal-path>`.
3. Run `my-friday validate <runtime-path> <memory-path>`.
4. If recovery refuses because state drifted, preserve the journal and targets
   and ask a maintainer to inspect them. Never manually delete a non-empty
   target.

Success is either one valid paired baseline with no support journal, or the
original absent/empty target state. Re-running recovery after validation is
safe.
