# Runbook

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

## Recover or remove an installed Codex baseline

Run `my-friday codex verify` first. A healthy or source-drifted installation
can be removed with `my-friday codex uninstall` after reviewing its exact
paths. Managed drift refuses deletion; use `my-friday codex repair` only when
the recorded runtime source is still the intended assistant.

For interruption, run only the exact command printed by the failure:

```sh
my-friday codex recover --transaction "$CODEX_HOME/.my-friday/transaction.json"
```

Recovery recognizes a complete manifest-consistent operation or restores the
stored projection, manifest, canonical, and previous generations. If exact
proof is unavailable,
preserve `.my-friday` and `AGENTS.md` for diagnosis. Never manually adopt or
delete a foreign file, shadowing override, symlink, hard link, or control tree.
Any unrecognized entry inside `.my-friday` blocks both mutation and recovery;
preserve it for diagnosis rather than deleting the control directory.
