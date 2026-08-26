# Context

## Problem And Desired Outcome

Issue #74 replaces an unreliable live-agent authoring dependency with a usable
deterministic workshop. A technically capable user must be able to create and
enhance one locally owned instruction-only capability without memorizing its
three-file schema, while retaining exact byte review, a separate source-write
authority, deterministic checks, separate activation, reversal, and recovery.

The desired result is not a scaffold. The same guided contract must support
first creation and later enhancement, preserve allowed reference/asset bytes,
and remain a clean input seam for the deliberately deferred agent adapter #75.

## Current State

At repository basis `1724e282393e69e4e023ffc07dd181dee0394bde`:

- `internal/capability/capability.go` owns strict package parsing, the three
  required files, bounded optional files, deterministic cases, stable state
  vocabulary, instance-wide non-blocking mutation lock, lifecycle journals,
  retained generations, collision/drift refusal, and recovery.
- `cmd/my-friday/main.go:runCapability` exposes inspect, validate, test, verify,
  install, upgrade, enable, disable, remove, and recovery. Source creation has
  no deterministic command.
- `internal/assistantinstance/instance.go` installs an instance-bound
  `capability-builder` skill and a manifest-owned My Friday executable. The
  builder currently gives an agent source-write permission and only read-only
  capability commands.
- `internal/terminal/wizard.go` demonstrates injected reader/writer terminal
  flow, preview-first behavior, exact confirmation, and no-change exit, but it
  is repository bootstrap-specific and is not a reusable capability workshop.
- `docs/architecture/capability-workshop.md` and ADR 0004 define source versus
  projection ownership, deterministic checks, source preservation, and
  explicit lifecycle mutations. The architecture document still describes
  live model authoring and PTY marker authority.
- `docs/plans/56-codex-builder-transport-amendment/` selected a complete agent
  contract plus `codex exec`. Subsequent bounded live evidence showed complex
  turns could finish without a qualifying action or source effect, invalidating
  that mechanism while leaving the package/lifecycle and release work useful.
- `bin/accept-capability-workshop` and its evidence/receipt verifiers already
  bind issue #51 to an immutable artifact, strict state facts, ambient
  preservation, one product owner, two design partners, and release. Their
  authoring scenarios currently assume the private agent builder.
- `bin/ci`, `go test -race ./...`, the acceptance contract tests, and native
  Apple-silicon/APFS acceptance are the repository's verification layers.

## Actors And Critical Journeys

- **Capability author:** answers bounded questions, navigates back/exit,
  reviews exact bytes, confirms a source action, and understands that the
  result is inactive or stale until a separate lifecycle action.
- **Returning author:** loads a valid package, explicitly retains or replaces
  each value, sees optional opaque bytes preserved, and reviews only core-file
  changes.
- **My Friday candidate:** derives the named instance from the real account
  home, verifies its manifest and executable, renders canonical bytes, owns the
  source transaction, and runs deterministic postconditions.
- **Lifecycle operator:** separately reviews and confirms Install or Upgrade;
  the workshop never supplies that token or calls lifecycle mutation.
- **Independent acceptors:** exercise the exact artifact through create,
  install, enhance, upgrade, disable/remove, interruption, refusal, recovery,
  and comprehension scenarios without private operator correction.
- **Maintainer:** diagnoses a source journal or collision using stable facts,
  without reading secrets, model transcripts, or user capability content from
  public evidence.

Success covers fresh create and valid enhancement. Denial covers invalid
answers and unsafe existing state. Recovery covers a process or fault after
source confirmation. Reversal covers lifecycle removal with source retained;
the workshop itself never deletes a user capability.

## Acceptance And Non-Goals

All issue #74 criteria are designable with existing repository primitives plus
one new internal workshop/source-transaction vertical. Exact source preview
means the complete canonical bytes and a unified core diff; it does not publish
private capability content as CI or GitHub evidence.

Out of scope are conversational-agent input, a public JSON/`--spec` interface,
a second package profile, direct arbitrary source editing, automatic install or
upgrade, model-evaluated tests, telemetry, dependencies, network, credentials,
scripts, background work, durable data, publishing, and deletion of source.

## Constraints, Dependencies, And Risks

- Supported pilot remains macOS 14+, Apple silicon, APFS, UTF-8 terminal, and
  the artifact release path documented in `docs/deployment.md`.
- The command must derive and verify the named assistant through the same real
  account-home boundary as other capability commands; caller `HOME` is not
  authority.
- Source files are user-owned and potentially private. Preview goes only to the
  user's terminal. Public evidence records digests and typed facts, never
  answers, full source, private paths, or diffs.
- Source commit and lifecycle mutation can race unless both use the existing
  `capabilities/` lock and re-inspect under lock.
- Updating three files is not atomic as a tree without staging and recovery.
  A source-specific journal, deterministic sibling stages, directory syncs,
  no-follow checks, and digest/inode proofs are required.
- Existing optional references/assets may be large within current bounds and
  must remain byte-identical; the workshop does not edit them in v1.
- `disabled` enhancement remains disabled. Enabling restores the prior retained
  projection; the CLI must explain that Enable followed by Upgrade is required
  to activate the new source.
- Existing instances do not receive new behavior silently. Instance upgrade
  and rollback remain explicit, manifest-bound operations.
- The work is release-bearing because it changes the public CLI, source bytes,
  named-instance contract, acceptance evidence, and artifact.

## Evidence, Assumptions, And Unknowns

**Evidence**

- Discovery #72 D1 was approved at
  `cf98e57110642dd0d0945fb55af2fbc66b4e3d49` after bounded live-agent probes.
- Strict parser/test and lifecycle safety behavior are implemented and covered
  in `internal/capability/capability_test.go`.
- Terminal and process/signal contracts have existing deterministic coverage in
  `internal/terminal`, `cmd/my-friday`, and acceptance support/supervisor tests.
- The artifact repository has no staging service; it nominates and accepts one
  immutable macOS artifact before release.

**Assumptions**

- Sequential plain-text questions plus complete final preview are sufficient
  for the first technical cohort.
- A deterministic SKILL template assembled from explicit purpose, behavior,
  inputs/outputs, examples, and facts is useful for creation, while existing
  arbitrary valid instruction bodies must remain retainable without inference.
- Existing value retention is understandable when every retained value is
  displayed and empty input explicitly means retain only in enhancement mode.

**Unknowns resolved by exact-candidate acceptance**

- Whether users understand source write versus activation without correction.
- Whether back/edit/review is efficient enough for multi-value answers.
- Whether users can recover from a deliberate interrupted transaction using
  only the CLI's stable guidance.

These are release evidence questions, not blockers to an implementation-ready
design. Failure returns the issue to implementation; it does not silently add
agent orchestration or another interface.
