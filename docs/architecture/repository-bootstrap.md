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
The executable's embedded, stable-`$id` schemas are the authoritative v1
contracts; generated copies under `.my-friday/schemas/` must match those bytes.
Validation authenticates copied schema bytes before compilation, then applies
semantic NFC, annotated grapheme, style, nullability, and custom-guidance rules
in addition to JSON Schema. Ordinary validation also
requires each target to remain a local Git repository, while permitting later
commits, branches, and remotes.

Targets must be distinct, non-nested, non-symlink, and empty or absent. Root
and the current home are prohibited. Existing exact fresh pairs report
`Already complete` without writes. Evolved Git metadata, including config,
hooks, refs, or objects, is not an exact rerun and is never rewritten. A
retained creation marker is recovery state, not an ordinary-valid pair.

## Transaction and recovery

The transaction writes an owner-only journal containing the original support
anchors before creating parents, then creates sibling stages with plan-bound
creation markers before any repository files or Git metadata,
initializes and validates both, then promotes runtime followed by memory. A
handled failure removes transaction-owned state; pre-existing empty shells are
recreated with their original modes. A journal is retained only when safe
automatic completion cannot be proven. Rollback removes a promoted repository
only after its marker and an exact snapshot of the complete staged tree,
including Git configuration, refs, hooks, objects, file types, modes, and
digests, prove transaction ownership; foreign or changed content is preserved
with the journal. An owned tree is atomically renamed to a plan-derived
deletion path before recursive removal, preserving retry authority if cleanup
is interrupted after deleting only part of the tree. Untouched pre-existing
empty shells remain original state. Recovery applies the same proof before
every promotion, re-proves the promoted pair, and rejects support paths that do
not derive from the stored original anchors and canonical targets. Marker removal and
reservation removal are separate durable, idempotent cleanup phases; original
shells and empty owned parents are restored before support authority is
removed. Use `my-friday recover --transaction <journal>`; recovery refuses
speculative mutation.

The implementation validates ownership but does not claim an adversarial
filesystem sandbox against an ancestor being replaced between checks.

## Privacy and external effects

Execution performs no network request, credential read, telemetry, global
Git/Codex change, import, commit, remote creation, or hosted-account setup. The
journal stores identifiers, paths, phase, and shell modes—not profile text.
