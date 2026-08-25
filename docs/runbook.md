# Runbook

## Capability lifecycle diagnosis

Start with `my-friday capability inspect NAME SLUG --plain`, then run
`validate`, `test`, and `verify`. `source-changed` requires a fresh tested
upgrade plan. Drift or collision is preserved; restore the receipt-bound
projection or resolve a foreign collision only after identifying its owner.
Disable retains a generation for enable. Remove clears only the selected
instance projection and control state and leaves runtime source intact.
Lifecycle changes require a fresh Codex task.

`capability recovery required` means a durable transaction journal remains.
Review the named instance and run `my-friday capability recover NAME SLUG`;
after exact confirmation it serializes on the instance lock and restores the
receipt-declared projection from retained bytes. If no receipt exists, recovery
restores absence. Do not delete the journal or generation manually.

Use `my-friday capability rollback --runtime PATH` or `my-friday assistant
rollback NAME` only for an unused v2 boundary. Runtime rollback refuses any
package source; assistant rollback refuses any receipt, generation,
transaction, or additional workspace skill. Preserve the compatible v2 binary
if either refusal occurs.

Runtime migration recovery is `my-friday capability recover --runtime PATH`.
Named-instance migration recovery is `my-friday assistant recover NAME`. Both
require exact `Recover` confirmation and accept only canonical owned journals;
do not repair, replace, or delete a malformed journal by hand. Preserve foreign
projection, control, and workspace entries for ownership diagnosis.

## Recover a named assistant instance

Run `my-friday assistant verify <name>` first. If removal stopped after the
exact launcher disappeared, run `my-friday assistant recover <name>`; a valid
manifest authorizes removal of only that instance root. If a launcher remains
but verification fails, or any manifest, root, managed executable, or launcher
proof drifted, recovery fails closed. Preserve both paths for diagnosis; do not
replace the launcher or recursively delete the root manually.

The prior single-home projection remains governed by the separate `my-friday
codex` commands below. Named creation never deletes or adopts it.

For explicit migration, set the same effective legacy `CODEX_HOME` used by the
old lifecycle and run `assistant migrate <name> --runtime <path> --memory
<path>`. Review both plans. The named instance is promoted and verified before
legacy uninstall begins. If cleanup fails, keep the working named instance and
run only the legacy recovery command printed by the failure.

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

The retained historical archive may contain its sole top-level executable as
`my-friday-darwin-arm64-<hex>`; future archives use `my-friday`. No other entry
name, nested path, symlink, non-executable file, or additional entry is valid.
The workflow copies the verified source archive unchanged rather than renaming
its internal executable.

The same retained archive may expose one top-level AppleDouble metadata member
named exactly `._<executable-name>` on GNU tar runners. It is valid only as the
sole companion to that executable and must be a regular member. The guard
extracts only the executable; it neither extracts nor uploads the AppleDouble
member separately.

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
During a committed uninstall, recovery may report the reserved
`.my-friday-removing` namespace. Do not rename or delete it manually: the
embedded committed journal authorizes the same recovery command to finish
deletion. The same namespace can hold an interrupted initial-install rollback;
its embedded journal must remain present until cleanup completes. When both
`transaction.json` and `transaction.json.next` exist, recovery accepts only
adjacent phases of the same pinned-root transaction. Either slot may contain the
newer phase depending on whether interruption happened before or after the
atomic swap. Recovery validates both slots before promoting or removing either.
A malformed, non-adjacent, wrong-root staging file, journal, or deletion
namespace is retained and refused for maintainer diagnosis.
Recovery also re-proves every journal entry after an atomic promotion, swap, or
move to cleanup. If a same-user process replaces an entry between validation
and mutation, recovery restores the moved entry when safe or retains both
locations and refuses. Preserve `transaction.json.discard` or
`.my-friday-removal.json.discard` if reported; rerunning recovery restores a
valid interrupted cleanup stage before proceeding. Either name makes verify
report `interrupted` and prevents another lifecycle plan until recovery finishes.

## Recover an interrupted acceptance run

Acceptance state lives only under the canonical-home parents
`.my-friday-acceptance` and `.my-friday-acceptance-evidence`. Diagnose one exact
run-ID child at a time. Read its owner-only `marker.json` and require the schema,
run ID, nonce, candidate SHA, canonical home, UID/GID, and directory identity to
match before cleanup.

If the image remains attached, verify the device, mountpoint, backing image,
APFS volume, and current owner against `attach.plist` and `diskutil info`. Stop
only the recorded acceptance process group, then use ordinary `hdiutil detach
<device>`—never `-force`. On any mismatch or detach failure, preserve the run
and evidence roots for maintainer diagnosis.
Attachment start is itself a preservation barrier. If attach output is partial
or ambiguous, do not remove the image, mountpoint, or anything beneath either
run root. Retain the attach receipt until an operator proves the complete graph,
performs ordinary detach, and proves graph absence.

After proven detach, remove only marker-listed files and the empty mountpoint.
Remove the exact run directory only when empty. Private before/after manifests
may be removed only after their marker and protected-state equality are proven;
then remove the exact evidence directory only when empty. Never recursively
delete either fixed parent or an unmarked child. A lone provisional GitHub
comment has no acceptance authority; failed finalization requires a fresh run.
Before detach or deletion, revalidate the backing image device/inode/owner/mode,
the complete image-whole/physical-store/container/volume graph, mounted device
and UUID, mount-table entry, writable local APFS identity, live `hdiutil`
association, and exact run/evidence markers. After ordinary detach, require all
device lookups, mount entries, and image associations absent. Cleanup reopens
the canonical home and each ancestor without following symlinks, matches both
root receipts and marker digests, unlinks owned entries descriptor-relative,
and removes only the two exact run-ID children. The failure trap attempts
ordinary detach only while all authority still matches; substituted or
ambiguous state is preserved.
Descriptor-relative cleanup requires the exact expected transitive entry set.
Any extra owner-controlled file, directory, link, or identity mismatch refuses
cleanup and preserves both roots for diagnosis. The helper must prevalidate and
hold bindings for both roots before unlinking from either; do not regenerate the
expected set from the live tree.

Interruption observation never reuses a missed attempt. Install, upgrade, and
uninstall each receive at most three fresh synthetic homes and valid padded
contract-v1 fixtures. The barrier requires three identical stopped receipts,
then kills and reaps the exact group and compares the post-kill projection,
manifest, canonical, previous, journal, and staging state to the captured
receipt before ordinary recovery.

## Recover an interrupted capability-workshop run

Capability-workshop state uses the same reviewed secure-root pair as native
acceptance: one random child beneath `~/.my-friday-acceptance/` and
`~/.my-friday-acceptance-evidence/`, plus randomized target and valid sibling
assistant/launcher pairs. Failure cleanup disarms signal traps, stops and reaps
bounded child groups, revalidates no-follow markers, manifests, auth receipts,
and APFS authority, then attempts manifest-owned named cleanup and ordinary
detach. It publishes non-approving
`capability-workshop-failure-v1` authority. Unproven cleanup preserves the run.

Identify the exact run from failure evidence and inspect `marker.json`. It must
bind schema `capability-workshop-run-v1`, issue 51, candidate SHA, artifact,
run ID, nonce, and current UID. Never recursively remove the parent or an
ambiguous assistant. Recover a proven interrupted capability with:

```sh
my-friday capability recover <random-instance> daily-brief
```

Then remove that exact instance through `assistant remove` and ordinarily
detach the marker-bound image without `-force`. If marker, manifest, inode,
foreign collision, credential-copy, or mount facts are ambiguous, retain them
for maintainer diagnosis and retry with a new run. Failure and provisional
comments cannot authorize product acceptance or release.
