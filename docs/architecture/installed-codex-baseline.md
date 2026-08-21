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
