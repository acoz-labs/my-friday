# Context

## Problem And Desired Outcome

Issue [#56](https://github.com/acoz-labs/my-friday/issues/56) must prove that a
fresh My Friday assistant can use its private capability builder to author a
useful, inactive capability before issue #51 can release. The approved plan
required a live PTY builder and explicitly reopened Solution Design if real
Codex could not demonstrate explicit skill discovery without relaxing the
no-activation boundary.

Candidate `7cc30e9ca934660bd3c00f62d7745e5adaded7cf` proved native prompt delivery
and exact completion-marker observation, but Codex created no source. Subsequent
disposable probes separated transport from behavior: typed app-server skill
input, equivalent text-only app-server input, and `codex exec` all reached a
normal terminal completion without filesystem or command action. The `exec`
result classified the missing prerequisite as the capability package contract.

The desired outcome is unchanged: an actual Codex agent authors the fixed
capability within the disposable private runtime, the candidate independently
proves it ready, and no activation occurs until the driver presents and confirms
the lifecycle plan.

## Current State

At repository basis `7cc30e9ca934660bd3c00f62d7745e5adaded7cf`:

- `internal/assistantinstance/instance.go` generates an instance-bound builder
  skill with paths, allowed check commands, and mutation prohibitions, but it
  does not define the package files, JSON fields, allowed values, or cross-file
  invariants.
- `internal/capability/capability.go` is the authoritative strict parser and
  deterministic test contract.
- `internal/capability/capability_test.go` contains complete valid package
  fixtures that are not available to a deployed instance's builder.
- `docs/architecture/capability-workshop.md` documents the package at a human
  overview level and currently specifies PTY authoring.
- `bin/accept-capability-workshop` performs the native issue-51 journey and
  currently treats a model marker as builder completion before candidate
  postconditions.
- `config/acceptance/launcher-capture.exp` is appropriate for the later fresh
  PTY invocation but model prose is not sufficient authoring authority.

The official Codex release list identifies 0.149.1 as the current stable
release, with only a full-diff link from 0.149.0 and no relevant documented fix:
https://github.com/openai/codex/releases/tag/rust-v0.149.1. The 0.150 releases
visible on the same page are pre-releases.

## Actors And Critical Journeys

- **User/builder agent:** asks the named assistant to create or enhance an
  instruction-only capability and reviews the resulting source before any
  lifecycle mutation.
- **Acceptance driver:** supplies the fixed authoring request, contains the
  process, verifies exact source postconditions, and separately owns mutation
  confirmations.
- **Candidate executable:** parses and tests the package and remains the sole
  authority for `ready` state.
- **Independent acceptor:** runs nominated bytes on Apple silicon/APFS and
  publishes only redacted typed evidence after complete cleanup.

Critical journeys are successful authoring, no-op or malformed authoring
refusal, postcondition failure, timeout/interruption cleanup, explicit installed
use in a fresh PTY, and preservation of ambient and sibling state.

## Acceptance And Non-Goals

The issue criteria remain designable without changing the capability profile or
activation model. The amendment changes only the missing builder contract,
authoring transport, and completion authority.

Non-goals remain executable capabilities, network/credential access, a
marketplace, automatic migration, a general scaffolding API, a Codex fork, a
pre-release Codex dependency, or weakening issue-4 evidence.

## Constraints, Dependencies, And Risks

- Builder instructions are manifest-owned per instance and must remain bounded
  enough for Codex context while complete enough to author without repository
  access.
- `codex exec` accepts prompt text/stdin but no structured `UserInput::Skill`;
  explicit selection therefore composes literal mention, exact prompt digest,
  and existing `debug prompt-input` catalog/path proof.
- The private event stream can contain model text or paths and must remain
  owner-only, unpublished, and removed after successful cleanup.
- Codex behavior is external and nondeterministic; candidate postconditions are
  deterministic and must be the only positive authority.
- Existing uncommitted fail-closed completion-gate work may be reused only after
  reconciliation with this amendment; it is not approved production code yet.

## Evidence, Assumptions, And Unknowns

### Evidence

- Candidate 13 failure record on issue #51 binds the no-source builder failure
  to exact candidate `7cc30e9` and immutable artifact authority.
- Native positional prompt delivery and cleanup passed before that failure.
- App-server typed skill catalog/path resolution was exact and the turn
  completed, but source and candidate checks were false.
- Equivalent text-only app-server input produced the same
  `acknowledged-without-action` result, excluding typed transport as the cause.
- `codex exec` has only literal-text skill selection; it exited zero with no
  action and classified the missing package contract while confinement and
  cleanup passed.

### Assumptions

- A complete generated contract will let the agent author the existing strict
  package without a new executable scaffold command.
- `codex exec` remains compatible with the instance's existing Codex config,
  auth-by-path copy, workspace trust, sandbox, and network denial.

### Unknowns

- Exact wording and size of the smallest sufficient builder contract; this is
  resolved during implementation by a failing-first disposable live probe and
  must not be guessed into release authority.
