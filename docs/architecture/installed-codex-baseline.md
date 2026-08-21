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
Phase publication swaps the new journal with its proven predecessor. If an
interruption leaves both slots, recovery accepts the current slot only when the
staged slot proves the same transaction at an earlier phase; ambiguous or
malformed staged authority is retained and refused.
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
