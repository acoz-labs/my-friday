# Portable assistant pilot

Development testing, not release acceptance. Use disposable agent repositories
and synthetic prompts; do not publish session transcripts, local installation
paths, credentials, or private capability source.

## Initial Codex trial — 2026-09-07

- Setup and project-directory launch: user reported correct identity, cwd, and
  explicit memory location.
- Fresh-session recall: a fictional project name survived restarting the harness.
- Correction/history: a renamed project was current and its old name historical.
  Read-only inspection confirmed two immutable revisions, explicit supersession,
  originating device metadata, and successful source validation.
- Working-directory independence: a fresh session from a different directory
  recalled the same remembered entity while reporting its actual new cwd.
- Private capability execution: an agent created instructions, a manifest,
  executable code, and checks, with no automatic subscriptions. Recorded tool
  results showed validation and execution success.

These are scoped observations from one trial, not proof of every harness,
machine, scope, or failure mode. User-visible results and recorded tool traces
are complementary evidence; no claim is made about inaccessible model reasoning.

## Finding: authoring discovery depended on the development checkout

Despite a successful final result, the agent could not obtain the capability
contract from the installed toolkit. Top-level --help returned an error. It
inspected executable strings and ultimately read Go source in a development
checkout to discover the manifest/check format. Reusable instructions also
included installation-specific paths.

Correction: ship successful CLI help, an embedded capability authoring guide,
a JSON manifest template, and MY_FRIDAY_BIN for runtime path resolution.
Point fresh harness sessions to that guide. Tests cover discovery without a
repository binding, template shape/invalid IDs, and projected instructions.

Fresh-session authoring retest passed on candidate `bc407a1` (2026-09-07).
The agent created a different small private capability using the installed
guide and manifest template, without inspecting executable strings or the
development checkout. Its reusable instructions resolve paths through the
runtime environment rather than hardcoded installation paths. It registered
no automatic subscriptions.

Recorded calls showed executable checks covering 13 counting cases, two error
cases, and local no-write assertions. An initial Unicode test expectation was
incorrect; the agent ran the check directly for diagnostics and corrected the
expected count, preserving the assertion. Independent reruns of the checks,
toolkit capability check, and source validation passed. The private source
worktree remained clean after verification.

This closes the observed authoring-discovery finding for this Codex scenario,
not every harness or capability. Natural-language discovery and reuse in
another fresh session remain pending; inspect recorded calls as well as the
answer to distinguish reuse from reimplementation.

## Next checkpoints

1. Validate capability reuse without reimplementation.
2. Repeat the core learning/capability experience through Pi.
3. Exercise remote sync, a second installation, and offline recovery.
4. Review remaining install/update/recovery and provenance gaps before release.
