# ADR 0001: Native bootstrap command

- Status: Accepted
- Date: 2026-08-20
- Provenance: issue #3 and planning PR #10

## Context and decision

Deterministic preview, Unicode handling, executable JSON Schemas, and truthful
recovery across two directories require more structure than copied templates
or a shell script. Python would make interpreter packaging a first-use concern.

Implement one Go 1.26 command with three pinned direct modules:
`jsonschema/v6` for draft 2020-12 validation, `x/text` for NFC, and `uniseg`
for grapheme clusters. Keep the domain harness-neutral; generated `AGENTS.md`
is the only Codex projection.

## Consequences

The binary has no language-runtime dependency and filesystem failures remain
typed and testable. Go tooling and future signing become artifact concerns. No
CLI framework, Git library, adapter registry, database, or telemetry is added.
