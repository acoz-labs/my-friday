# Context

## Problem And Desired Outcome

[Issue #3](https://github.com/acoz-labs/my-friday/issues/3) asks for the first
selected product outcome: a technically capable user can understand, preview,
and create two valid local Git repositories for one personalized assistant.
One repository owns runtime identity and future skills/configuration; the other
owns governed memory. Creation must not imply or perform installation.

The approved product-experience contract is a terminal flow:

1. Scope
2. Identity
3. Communication style
4. Locations
5. Preview
6. Creation and verification
7. Result or recovery

Success means both repositories validate against versioned contracts and the
preview exactly describes every durable mutation. Failure means the original
state is restored or a bounded, visible transaction can be recovered without
guessing.

## Current State

Repository state is evidenced at
`c669868cb297be51f02e1b6b7824e50b81da318b`:

- `README.md` and `docs/product.md` define My Friday as a local-first toolkit,
  retain the runtime/memory separation, select O1 first, and explicitly park
  broad operating-system support.
- `docs/architecture.md` is an unfilled managed-repository scaffold; there is
  no application runtime, command, schema, generator, or package manifest.
- `docs/development.md` delegates validation to `bin/container bin/ci`, while
  `bin/ci` currently validates only repository standards and solution plans.
- `docs/deployment.md` has no declared delivery profile or artifact contract.
- `docs/plans/_template/` and `bin/validate-solution-plans` own the planning-pack
  convention used here.
- `.github/workflows/ci.yml` invokes `bin/ci`; artifact nomination, acceptance,
  and release workflows exist as managed scaffolding but are not yet configured
  as a product release path.
- `LICENSE` and the README license link establish MIT licensing.

The Phase 1 authority is the approved exact head
`b6db62bf15c8d6ad7a15f7533e6aa5981ae1cd8a` of
[acoz-labs/.github PR #2](https://github.com/acoz-labs/.github/pull/2), outcome
O1. Issue #3 remains open in `Solution Design` and carries that provenance in
its lifecycle block.

## Actors And Critical Journeys

### Primary user

A technical Codex user at a local interactive terminal who understands file
paths and Git but should not need to reason about transaction internals.

- **Success:** supplies profile and path choices, reviews an exact plan,
  explicitly creates, and receives two verified repository paths plus clear
  next steps.
- **Safe exit:** presses Return at confirmation and leaves with no filesystem
  mutation.
- **Validation denial:** sees the offending field or path and remains at the
  relevant step without losing earlier valid answers.
- **Collision denial:** sees canonical conflicting targets and no target
  mutation.
- **Interrupted creation:** reruns the displayed recovery command and reaches
  either two valid repositories or the exact pre-run state.

### Generated runtime consumer

A future Codex session or installation workflow reads the runtime repository's
manifest, generic instructions, and assistant profile. It must be able to
distinguish communication preferences from safety or authorization policy.

### Generated memory consumer

A future governed-memory workflow reads the memory repository's role and
assistant identifier. O1 supplies governance scaffolding and empty categories,
not memory records or promotion behavior.

### Contributor and reviewer

Contributors must reproduce plan rendering, collision behavior, failure points,
and generated contracts without touching real user directories. Reviewers need
golden terminal transcripts and filesystem fixtures bound to the candidate
commit.

## Acceptance And Non-Goals

### Acceptance groups

1. **Complete deterministic preview:** normalized inputs produce one canonical
   plan identifier, assistant identifier, exact target paths, ordered actions,
   every missing parent segment, exact generated file list, content digests,
   resulting modes, temporary-support policy, and actions that will not occur.
2. **Valid separate repositories:** runtime and memory targets are independent
   Git worktrees on `main`; their embedded v1 manifests and schemas validate;
   assistant identifiers match; runtime profile matches the answers; memory
   contains zero records.
3. **No adjacent effects:** no global Codex or Git configuration, remote,
   commit, credential access, network call, imported content, telemetry, or
   write outside declared targets, planned missing parents, and transaction
   support paths.
4. **Safe paths and collisions:** canonical targets are distinct, non-nested,
   not `/` or the home directory, reside on supported local APFS, and are not
   symlink or non-empty collisions. Path grammar/defaults resolve visibly from
   one captured invocation directory.
5. **Recoverability:** injected failure at every mutating transition ends in
   both valid, neither beyond allowed pre-existing empty shells, or an
   owner-only journal accepted by idempotent recovery.
6. **Accessible terminal behavior:** every state has stable line-oriented text,
   keyboard-only operation, visible focus-by-prompt, non-color status, and no
   animation or cursor rewriting. Unicode profile values and paths work after
   validation.
7. **Idempotent retry:** an exact completed plan reports `Already complete`
   without writes; a matching interrupted plan routes to its existing recovery
   state; only unrelated or ambiguous content is a collision. `b`, `q`, and EOF
   have explicit no-mutation behavior before creation.

### Explicit non-goals

- Codex installation, managed projections, repair, upgrade, uninstall, or
  rollback (O2).
- Memory capture, proposal, promotion, or recall behavior (O3).
- Skill creation or installation (O4).
- Commits, remotes, provider authentication, hosted repositories, or sync
  (O5-O7).
- Intel Mac, non-APFS, Linux, Windows, background services, GUI/web UI, model
  access, secrets management, or telemetry.
- Alternate agent harnesses, a harness adapter framework, or a lowest-common-
  denominator profile. O1 remains Codex-first; each future harness requires a
  product decision and explicit capability/projection mapping.
- Editing, adopting, or merging an existing non-empty repository.
- Marketing/name clearance or artifact publication.

## Constraints, Dependencies, And Risks

### Supported environment

- macOS 14 or later, `arm64`.
- Local APFS; Apple documents APFS as the default Mac filesystem since macOS
  10.13: <https://support.apple.com/guide/disk-utility/file-system-formats-dsku19ed921c/mac>.
- Interactive terminal with UTF-8 input/output. The executable is not sourced
  and has no shell-language dependency; examples use `/bin/zsh`.
- Git 2.28 or later because the contract uses `git init --initial-branch=main`;
  that option is present in Git's 2.28 documentation:
  <https://git-scm.com/docs/git-init/2.28.0>.
- Go 1.26.x for development. Official Go distribution evidence supports
  `darwin/arm64` binaries and macOS 12+, while this product deliberately claims
  the narrower macOS 14+ baseline: <https://go.dev/dl/>.

### Risk summary

- Two directory promotions cannot be one filesystem-atomic transaction.
- Symlink/path checks are vulnerable to hostile concurrent mutation unless all
  operations are directory-descriptor-relative; the initial local single-user
  contract therefore combines canonical checks, reservations, rechecks, and
  recoverability without claiming adversarial sandboxing.
- Terminal text can be spoofed by control/format characters; user-provided
  strings must reject them even while ordinary Unicode remains valid.
- Git templates can import local hooks or private content; initialization must
  use a tool-owned empty template directory.
- Purpose/personality text can be sensitive; success logs and failure journals
  must not duplicate it.
- Generated schema and internal validation can drift; both must be exercised
  over the same conformance corpus.
- A native binary reduces user runtime dependencies but creates future artifact
  signing/notarization and architecture-release work outside this envelope.

There are no secrets, migrations, external APIs, background jobs, privileged
operations, or production environments in O1.

## Evidence, Assumptions, Unknowns, And Decisions

### Evidence

| ID | Evidence | Source |
|---|---|---|
| E1 | Product authority approved O1 as the first selected outcome. | Discovery PR #2 exact head and issue #3 lifecycle |
| E2 | The product owner selected macOS as the first environment and MIT as the license. | Product decision; `LICENSE` at repository basis |
| E3 | The repository has no implementation/runtime and does have managed validation and planning conventions. | Current source paths listed above |
| E4 | Git 2.28 exposes `--initial-branch`; Go supports native Darwin ARM64 binaries; APFS is the modern Mac default. | Official Git, Go, and Apple documentation linked above |
| E5 | No representative user has yet completed the flow independently. | `docs/product.md` validation signals and discovery unknowns |
| E6 | `rivo/uniseg` v0.4.7 implements Unicode extended grapheme segmentation and has no non-standard-library dependencies. | <https://github.com/rivo/uniseg/tree/v0.4.7> |
| E7 | Go's supplementary `x/text/unicode/norm` package provides Unicode normalization and v0.40.0 is BSD-3-Clause. | <https://pkg.go.dev/golang.org/x/text/unicode/norm@v0.40.0> |

### Assumptions

| ID | Assumption | Validation |
|---|---|---|
| A1 | A line-oriented terminal wizard is acceptable to the first technical cohort. | Independent hands-on completion and comprehension interview |
| A2 | The first pilot machine is Apple silicon, local APFS, and meets the Git floor. | Environment preflight before first-customer acceptance |
| A3 | Two repositories remain understandable when their roles and shared identifier are explicit. | Ask users to explain ownership and privacy boundaries after creation |
| A4 | Owner-only default modes do not conflict with expected local use. | Verify on fresh and pre-existing empty target directories |
| A5 | Separating owned assistant data from the Codex `AGENTS.md` projection preserves a future harness seam without implementing one. | Review generated contracts; require a future capability map before adding a harness. |

### Unknowns

| ID | Unknown | Disposition |
|---|---|---|
| U1 | Exact first-customer macOS version, architecture, filesystem, and Git version. | Release prerequisite; preflight discovers it without changing the design. |
| U2 | Public naming clearance. | Artifact-release prerequisite; implementation may proceed. |
| U3 | Whether Intel Mac support is valuable. | Park with O8; do not claim or build a universal artifact in O1. |
| U4 | Final code-signing/notarization and distribution channel. | Future artifact-release design; outside the `implementation` envelope. |
| U5 | Independent setup completion and comprehension. | Product acceptance evidence; not an architecture unknown. |
| U6 | Whether Claude Code, pi, or another harness should be supported and how its capabilities map. | New product decision; not part of O1/O2 Codex-first scope. |

### Decisions

Consequential decisions are centralized in the README Decision Spotlight and
expanded in `02-decision.md` and `03-design.md`. No unknown above prevents an
implementation-ready design.
