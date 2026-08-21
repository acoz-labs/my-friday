# ADR 0002: Manifest-owned Codex baseline

- Status: Accepted
- Provenance: issue #4; approved planning PR #15

## Context

`CODEX_HOME` mixes global instructions with configuration, credentials,
sessions, logs, and extensions. Safe reversal requires precise ownership.

## Decision

Render the runtime's `AGENTS.md` as one atomic regular file and bind it to a
private manifest and durable journal. Fail closed on foreign state, links,
drift, shadowing, or ambiguous recovery. Retain one verified rollback
generation. Reject symlinks, `config.toml` merging, whole-home management,
background reconciliation, and general backup history.

## Consequences

The lifecycle is inspectable and source-independent for reversal, but it never
silently merges user instructions. Collisions require user resolution and
managed drift requires explicit repair.
