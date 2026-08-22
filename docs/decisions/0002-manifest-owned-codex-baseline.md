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

Exact-candidate acceptance uses a no-admin APFS image plus fail-closed macOS
sandbox write containment instead of creating and deleting a disposable user.
This proportionate boundary is same-UID and therefore deliberately claims no
read confidentiality, fresh keychain, or malicious same-UID resistance.
Acceptance authority requires a no-authority provisional evidence comment and
a separately published post-cleanup finalization comment. Their IDs, SHA-256
body digests, author, candidate, artifact, and issue bindings are revalidated
before release.
