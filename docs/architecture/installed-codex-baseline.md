# Installed Codex baseline

My Friday manages exactly one global Codex projection: the regular file
`$CODEX_HOME/AGENTS.md`. Its private `$CODEX_HOME/.my-friday` directory holds
the ownership manifest, transaction journal, one previous generation, and
short-lived recovery copies. No other Codex-home content is managed.

## Lifecycle and authority

Mutations first print a path-and-digest preview and require the matching
case-sensitive confirmation word. `verify` and planning are read-only. Every
execution rebuilds the plan immediately before mutation.

The effective Codex home must be an existing non-symlink directory below the
current user's real home. Managed files must be regular, single-link files. A
foreign `AGENTS.md`, `.my-friday`, or `AGENTS.override.md` refuses installation.
Manifest disagreement or drift refuses upgrade, rollback, and uninstall;
repair restores only the matching assistant identity from its recorded source.

The manifest binds assistant identity, projection digest, source path and
digest, and an optional previous-generation digest. Reversal is independent of
source availability. Source disagreement is reported separately and gates only
operations that render source bytes.

## Failure and recovery

Transactions are foreground-only. Before changing an installation, My Friday
stores exact pre-change projection and manifest bytes, then writes an
owner-private journal. Recovery either recognizes a fully committed result or
restores those proven bytes. Incompatible journals and missing proofs are
retained and refused.

Upgrade retains one verified prior projection. Repair never rotates it or
stores drifted bytes. Rollback consumes it. Uninstall removes only a
digest-verified projection and the owned control directory.

Implementation: `internal/codexhome/lifecycle.go`; command grammar:
`cmd/my-friday/main.go`; tests: `internal/codexhome/lifecycle_test.go`.
