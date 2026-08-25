# Context

## Problem And Desired Outcome

[Issue #51](https://github.com/acoz-labs/my-friday/issues/51) requires one
inspectable capability-building loop before governed memory: a user collaborates
with the assistant to define, scaffold, review, validate, test, install, enhance,
disable, and remove an instruction-only capability. Source must remain visible
and versioned; the agent may help author it but may not activate it.

## Current State

Repository basis: `f35f08fb07e51299657841ac626ebdfba492e80e`.

- `internal/repository` creates and validates a manifest-owned runtime repository
  containing `assistant/profile.json`, schemas, and reserved `skills/.gitkeep`.
- `internal/assistantinstance` creates a private named instance at
  `$HOME/.my-friday/assistants/<name>`, copies validated runtime and memory
  repositories into it, pins five owned directories plus its manifest, installs
  a native launcher, and fixes Codex's working directory to `workspace`.
- `internal/codexhome`, `internal/transaction`, and the named-instance lifecycle
  establish no-follow traversal, exact ownership/digest checks, advisory locks,
  durable recovery evidence, collision refusal, source-independent reversal,
  and exact-token mutation precedents.
- `cmd/my-friday/main.go` is the CLI composition root; terminal tests already
  prove case-sensitive newline-terminated confirmation and `No changes made`.
- `bin/ci` is the required local/CI entry point. The artifact nomination,
  independent acceptance, production receipt, and GitHub Release machinery is
  mature and documented in `docs/deployment.md` and `docs/runbook.md`.
- There is no staging service. Production is the immutable Darwin/ARM64 release
  artifact accepted and published through the existing release workflow.

The supported Codex skill contract requires `SKILL.md` with `name` and
`description`, supports optional `scripts/`, `references/`, `assets/`, and
`agents/openai.yaml`, scans `.agents/skills` from the working directory through
the repository root, and supports explicit skill invocation. Source:
[OpenAI Codex Skills](https://developers.openai.com/codex/skills/), consulted
2026-08-25. This design deliberately permits only a strict subset.

## Actors And Critical Journeys

- **User/product authority:** asks naturally for an ability, reviews the complete
  source diff and effects, then independently authorizes each mutation.
- **Named assistant/builder skill:** clarifies intent, edits only capability
  source and tests, explains validation failures, and never activates a package.
- **My Friday CLI:** validates, plans, owns projections and receipts, serializes
  mutation, reports stable states, and recovers interrupted work.
- **Codex:** discovers the installed repository-scoped skill from the instance
  workspace and executes instructions in a fresh task.
- **Release operator and independent acceptor:** preserve exact artifact identity,
  exercise the full lifecycle in disposable roots, and authorize release under
  existing repository policy.

Critical journeys are bootstrap, first authoring, safe denial, install/verify,
source-first enhancement, collision/drift diagnosis, disable/re-enable/remove,
interruption recovery, and exact-candidate acceptance.

## Acceptance And Non-Goals

The issue is designable without changing its outcome. C1 ships one profile:
`instruction-only`. It includes the built-in builder and one complete locally
installed user capability lifecycle.

Excluded: executable code or scripts, arbitrary dependencies, network or
credential use, background work, durable capability-owned user data, publishing,
registries, marketplace/plugin distribution, automatic updates, remote package
sources, authenticated-human confirmation claims, general plugin architecture,
and the governed-memory capability in #52.

## Constraints, Dependencies, And Risks

- O1 runtime repositories and O2 named-instance ownership are shipped inputs;
  both require forward-compatible v1-to-v2 migration without weakening v1
  verification or recovery.
- Named instances contain copied runtime source. Editing the original runtime
  does not mutate an installed assistant; upgrade must be explicit.
- A user-authored `SKILL.md` can contain harmful instructions even without code.
  Structural validation reduces effects but cannot certify semantics.
- Codex may retain a skill within a running task. Install/upgrade/disable/remove
  guarantees apply to a fresh task, with restart guidance if discovery lags.
- My Friday owns only its instance-scoped projections. It does not inspect or
  resolve same-named skills in unrelated global/user scopes and must not claim
  ecosystem-wide uniqueness.
- Capability state crosses two trust surfaces—the Git source repository and the
  private installed instance—but no single mutation spans both roots.
- No secrets or new external service are required. Acceptance reuses the
  repository's existing file-backed Codex credential boundary without recording
  credential contents.

## Evidence, Assumptions, And Unknowns

**Evidence:** the source paths and official Codex contract above; approved
discovery PR #50 at `4d41ca094ce5aa9c6fd45057763564b027dcc460`;
the product-design contract on issue #51; existing release documentation and
tests at the repository basis.

**Assumptions to validate:** users understand source versus projection after one
guided pass; explicit invocation is acceptable for the first profile; static
tests plus exact-candidate Codex acceptance form an honest evidence boundary.

**Unknowns:** usefulness and comprehension across users are pilot learnings, not
implementation blockers. The approved design-partner exercise must record them
without expanding C1.
