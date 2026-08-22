# Installed Codex baseline

My Friday manages exactly one global Codex projection: the regular file
`$CODEX_HOME/AGENTS.md`. Its private `$CODEX_HOME/.my-friday` directory holds
the ownership manifest, a canonical active generation, transaction journal,
and one previous generation. No other Codex-home content is managed.

## Lifecycle and authority

Mutations first print a path-and-digest preview and require the matching
case-sensitive confirmation word followed by a newline. `verify` and planning
are read-only; verify exits nonzero for every unhealthy state. Every execution
rebuilds the plan and compares the root identity, source, before-state proofs,
desired bytes, and action set with the confirmed preview before mutation.

The effective Codex home must be an existing non-symlink directory below the
current non-root user's real home on a local APFS volume. Every target-path component is opened without
following symlinks, the root device and inode are pinned, and managed access is
descriptor-relative. The opened control directory is rebound to its root entry
before every transition. Managed files must be regular, single-link files. A
foreign `AGENTS.md`, `.my-friday`, or `AGENTS.override.md` refuses installation.
Manifest disagreement or drift refuses upgrade, rollback, and uninstall;
repair restores only the matching assistant identity from its recorded source.

The manifest binds assistant identity, projection and canonical digests, a
canonical absolute source path and digest, and an optional previous-generation
digest. The control tree is a complete allowlist; any foreign entry blocks
mutation and recovery. Reversal is independent of source availability. Source
disagreement is reported separately and gates only operations that render
source bytes.

## Failure and recovery

Transactions are foreground-only. The owner-private journal records the pinned
root identity plus exact before/after proofs for projection, manifest,
canonical, and previous generations. Fixed staging names are part of the
control-tree allowlist, journal publication is exclusive and atomic, and
recovery validates the action-specific relationship between both generations.
Phase publication durably stages the new journal, then swaps it with its proven
predecessor. If interruption leaves both slots, recovery accepts either adjacent
ordering for the same transaction: current-newer/staged-older after the swap, or
current-older/staged-newer before it. Both journals, including their pinned root
authority, are validated before either slot is promoted or removed. Ambiguous,
non-adjacent, malformed, or wrong-root staged authority is retained and refused.
Recovery promotions, swaps, and cleanup first move entries through fixed
allowlisted names, then prove the exact moved and displaced bytes. A concurrent
replacement is restored when that can be proved; otherwise both locations are
retained and recovery refuses without using stale journal bytes. Interrupted
cleanup stages are restored to the normal journal slots before recovery resumes.
The discard-stage names `transaction.json.discard` and
`.my-friday-removal.json.discard` always classify the installation as
interrupted and block planning until recovery removes them.
File replacement uses an atomic swap or exclusive rename so the displaced
entry can be proved before deletion. Each mutation boundary is durably phased,
and the complete result is reproved before commit.
Recovery either completes the committed result or restores every proven
pre-change generation. Incompatible journals and missing proofs are retained
and refused.

Upgrade retains one verified prior projection. Repair never rotates it or
stores drifted bytes. Rollback consumes it. Uninstall removes only a
digest-verified projection and atomically renames the journal-bearing control
directory to a reserved deletion namespace before final cleanup, so interruption
cannot leave an unauthoritative empty control directory. An interrupted initial
install uses the same journal-bearing detached deletion protocol when rollback
returns the control tree to absence.

Implementation: `internal/codexhome/lifecycle.go`; command grammar:
`cmd/my-friday/main.go`; tests: `internal/codexhome/lifecycle_test.go`.

## Acceptance boundary

Release acceptance runs the exact nominated Darwin/ARM64 executable under
`bin/accept-installed-codex-baseline`. The foreground supervisor creates a
marker-owned APFS disk image below the acceptor's canonical home, supplies
synthetic user, Codex, temporary, XDG, runtime, and candidate paths, and runs
candidate processes under fixed deny-default sandbox profiles. Lifecycle
commands and descendants have volume-only write authority and no network. The
fixed Codex login/exec/logout smoke keeps the write boundary but permits broad
outbound network; it makes no endpoint-restriction claim.
The installed Codex command may be a current-user-owned symlink chain; the
supervisor resolves and pins its owned regular executable before mutation.
Profile templates are rendered by literal placeholder replacement rather than
shell substitution, normalized byte-for-byte, and positive-controlled by both
the real Go candidate and resolved Codex executable. The reviewed Go-runtime
`sysctl-read` operation is allowed; broad Mach lookup remains denied and new
diagnostics fail.

The supervisor externally stops and kills exact candidate process groups only
after `tools/acceptance-stop-barrier` observes a schema-valid recoverable journal
behind an all-members-stopped barrier. The barrier binds PID, PGID, process
start identity, pinned Codex-home device/inode, complete proof digests,
action-specific before/after invariants, allowed durable phase, and adjacent
staging authority against descriptor-opened projection, manifest, canonical,
previous, and staged bytes before signaling. Three stable stopped reads and an
identical post-kill filesystem receipt are required. Each operation gets up to
three fresh synthetic homes/fixtures; a missed window carries no evidence into
the next attempt. Ordinary `codex recover` must then
restore install, upgrade, and uninstall cases. No fault-enabled candidate or
background service exists.

Before setup and after ordinary detach, private inventories compare metadata
for live Codex/runtime roots and exact bytes only for schema-allowlisted,
non-sensitive managed files. Secret-bearing and foreign contents are never
opened for evidence. A provisional GitHub comment has no authority. A separate
post-cleanup finalization comment cross-binds its immutable ID and body digest;
acceptance and release re-fetch and verify both comments. This is same-UID
write containment, not distinct-principal isolation, read confidentiality, or
a fresh login keychain.

The supervisor itself is byte-bound to the candidate checkout: clean status is
combined with Git blob checks for the complete repository-owned transitive Go
source closure, module sums, profiles, and supervisor. Evidence records that
closure, pinned build inputs/flags, tree/blob IDs, and an aggregate digest of
the actually executed helper binaries. Every candidate and Codex process runs
from a synthetic cwd under a hard process timeout; process-start-bound lineage
tracking kills descendants even after `setsid` and proves no live descendants.

Image authority binds the sparse-image inode to the image whole disk, APFS
physical store, container, volume device/UUID, mount root, and live `hdiutil`,
`diskutil`, and mount-table views. Cleanup begins only after all four device
identities and the image association disappear. Exact run children are then
removed descriptor-relative under no-follow ancestor walks using matching
root receipts and markers; fixed parents are never recursively deleted.
