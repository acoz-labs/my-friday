# Repository Bootstrap Capability

## User contract

`my-friday init` gathers identity, communication style, and one parent
location; defaults create `my-friday-runtime` and `my-friday-memory`. The
line-oriented preview declares the plan ID, assistant ID, normalized absolute
targets, generated files, and prohibited adjacent effects. Only exact `Create`
mutates disk; every other confirmation input exits.

Profile text is trimmed, NFC-normalized, screened for control/format/line
separator characters, and counted as extended grapheme clusters. Display name
and address are limited to 60; purpose and custom guidance to 240. Personality
is presentation data and cannot override authorization, safety, trust, privacy,
or tool policy.

## Repository contracts

Both targets are owner-only Git repositories on unborn branch `main`, with no
commit or remote. Git initialization supplies a tool-owned empty template.
Each `.my-friday/manifest.json` declares contract version 1, role, shared
`assistant_id`, and generator version. Runtime additionally owns
`assistant/profile.json`; memory starts with empty `data/observations`,
`data/journals`, `data/proposals`, and `data/memories` scaffolds plus a reserved
`schemas/README.md`, and no records.
Generated copies under `.my-friday/schemas/` are the authoritative v1 contracts.

Targets must be distinct, non-nested, non-symlink, and empty or absent. Root
and the current home are prohibited. Existing exact pairs report `Already
complete` without writes.

## Transaction and recovery

The transaction writes an owner-only journal, creates sibling stages with
plan-bound creation markers,
initializes and validates both, then promotes runtime followed by memory. A
handled failure removes transaction-owned state; pre-existing empty shells are
recreated with their original modes. A journal is retained only when safe
automatic completion cannot be proven. Rollback removes a promoted repository
only after its marker, planned file set, and content digests prove transaction
ownership; foreign or changed content is preserved with the journal. Recovery
rejects journal-supplied paths that do not derive from the plan and canonical
targets. Use `my-friday recover --transaction <journal>`; recovery refuses
speculative mutation.

The implementation validates ownership but does not claim an adversarial
filesystem sandbox against an ancestor being replaced between checks.

## Privacy and external effects

Execution performs no network request, credential read, telemetry, global
Git/Codex change, import, commit, remote creation, or hosted-account setup. The
journal stores identifiers, paths, phase, and shell modes—not profile text.
