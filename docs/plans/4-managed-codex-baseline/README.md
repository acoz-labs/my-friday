# Solution Design: Safely manage the installed Codex baseline

- **Status:** Draft
- **Issue:** #4
- **Planning PR:** Pending
- **Repository basis:** 5bc309226d2c40e1473a4011c1bd8552c995919d
- **Execution envelope:** Pending

## Decision

Add a `my-friday codex` lifecycle that renders one self-contained global
`AGENTS.md` projection from a validated runtime repository and manages only
that exact file plus a private `.my-friday` control directory inside the
effective Codex home. Every mutation is planned and confirmed, ownership is
manifest-backed, filesystem changes are journaled and recoverable, foreign or
drifted state fails closed, and one last-known-good generation supports
rollback. My Friday will not edit `config.toml`, authentication, sessions,
logs, skills, packages, project configuration, or the Codex executable.

This is the smallest coherent baseline on the currently documented Codex
surface: user-level `AGENTS.md` is discovered from `CODEX_HOME`, while
`CODEX_HOME` also contains unrelated sensitive and operational state that must
remain outside My Friday's ownership.

## Needs Attention

- Exact-candidate acceptance must run on supported Apple silicon, macOS 14 or
  later, local APFS, and Git 2.28 or later under a disposable non-admin macOS
  user/home. A verified VM or APFS volume boundary is acceptable only if it
  provides an equally fresh user identity, home, keychain, and Codex home.
- Automated lifecycle tests must use injected temporary home and Codex roots.
  Destructive install, repair, uninstall, rollback, or recovery tests must
  never target Alfred's live `~/.codex`, a developer's real Codex home, or a
  deployed `batcomputer-ai` release projection.
- The durable deployment guide still describes public-release prerequisites
  that are inconsistent with the repository's existing accepted artifact
  release. Implementation reconciliation must correct that stale contract
  before this change is released.

## Decision Spotlight

- **One projection, not general configuration management.** The managed
  baseline is only `$CODEX_HOME/AGENTS.md`; `config.toml`, auth, logs, sessions,
  skills, packages, profiles, system config, and project `.codex` layers remain
  user- or Codex-owned. This minimizes collision and credential exposure.
- **Rendered regular file, not a symlink.** The projection is a deterministic,
  self-contained rendering of the validated runtime profile. An atomic regular
  file avoids broken moves, symlink traversal, and ambiguous relative-path
  semantics while the manifest retains source provenance.
- **Collision and drift fail closed.** A foreign `AGENTS.md` or `.my-friday`
  namespace blocks installation. A managed file whose digest differs from the
  manifest blocks upgrade, rollback, and uninstall; only an explicit,
  previewed repair may replace it. My Friday never silently merges prose.
- **Codex home is a distinct trust boundary.** Source checkout/worktree
  isolation does not isolate installed state. Production commands resolve the
  effective `CODEX_HOME` once, require it to be an existing, owner-controlled,
  non-symlink descendant of the current user's real home, and display the
  canonical path before confirmation.
- **Tests inject roots; acceptance changes identities.** Unit and integration
  tests receive temporary roots through internal dependencies and never infer
  the invoking developer's home. Exact-candidate acceptance uses a disposable
  non-admin macOS user/home because a temporary directory alone cannot prove
  separation from keychain, login state, or host-level path discovery.
- **One rollback generation.** Each successful repair or compatible upgrade
  retains exactly one verified prior managed generation. Rollback is explicit,
  previewed, single-step, and idempotent; My Friday is not a general backup or
  configuration-history service.
- **No daemon, privilege escalation, or network.** Lifecycle work is a
  foreground local transaction, refuses root/admin elevation, does not invoke
  `sudo`, and makes no network request. A fresh Codex discovery smoke may use a
  separately authorized test account during independent acceptance, never as
  part of installation.

## Plan Map

- [Context](01-context.md)
- [Decision](02-decision.md)
- [Design](03-design.md)
- [Verification](04-verification.md)
- [Implementation handoff](05-handoff.md)

## Final Gate

The plan becomes final only after the draft planning PR is linked to issue #4,
the complete pack has no blocking maintainer finding, plan validation passes,
and the exact PR number and `through-production` envelope are recorded. Product
authority must then approve the exact final head before merge.
