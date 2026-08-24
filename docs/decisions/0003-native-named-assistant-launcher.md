# ADR 0003: Native named assistant launcher

- Status: Accepted
- Provenance: issue #4; approved planning PR #30

## Context

Named assistants need isolated mutable state while retaining the user's real
home and ordinary macOS identity.

## Decision

Use a manifest-owned root under `$HOME/.my-friday/assistants/<name>` and exactly
one collision-checked native launcher at `$HOME/.local/bin/<name>`. It preserves
`HOME`, sets instance `CODEX_HOME`, fixes Codex `--cd`, and invokes an
instance-owned Codex executable without a shell.

## Consequences

Instances are inspectable and independently reversible without shell edits.
The launcher directory must already exist and is never adopted. Isolation
controls My Friday's mutation boundary; it does not claim confidentiality from
same-UID processes. Caller roots, substituted `HOME`, aliases, OS users, and
background reconciliation remain rejected.
