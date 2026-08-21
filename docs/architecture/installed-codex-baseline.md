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
current user's real home. Every target-path component is opened without
following symlinks, the root device and inode are pinned, and managed access is
descriptor-relative. Managed files must be regular, single-link files. A
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
canonical, and previous generations. Each mutation boundary is durably phased.
Recovery either completes the committed result or restores every proven
pre-change generation. Incompatible journals and missing proofs are retained
and refused.

Upgrade retains one verified prior projection. Repair never rotates it or
stores drifted bytes. Rollback consumes it. Uninstall removes only a
digest-verified projection and the owned control directory.

Implementation: `internal/codexhome/lifecycle.go`; command grammar:
`cmd/my-friday/main.go`; tests: `internal/codexhome/lifecycle_test.go`.
