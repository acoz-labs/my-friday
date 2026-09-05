# Context

## Problem And Desired Outcome

Issue #92 materializes B1 from approved discovery #81/#82. A user must be able
to create, inspect, reconcile, verify, diagnose, repair, upgrade, roll back,
migrate, and remove a versioned assistant baseline backed by one configured
private Git repository. The repository must give `config/`, `memory/`, and
`capabilities/` independently understandable governance without making users
coordinate separate repositories.

The desired result is the smallest trustworthy control plane beneath later B2
capability and B3 memory semantics: repository establishment, baseline and
module versioning, exact-path Git stewardship, migration ports, lifecycle
receipts, host binding, Codex launching, diagnosis, and recovery.

## Current State

Evidence at repository basis
`2594147084e8bb6f961971efdb116b02824ecf66`:

- `internal/plan.Build` and `internal/repository.Create` create separate
  runtime and memory Git repositories with a shared assistant ID, embedded
  schemas, owner-only modes, no commit, and no remote.
- `internal/transaction` already demonstrates sibling staging, durable journals,
  exact ownership proofs, atomic promotion, quarantine, rollback, and
  interruption recovery for two filesystem targets.
- `internal/assistantinstance` creates a named root and launcher, copies
  validated runtime and memory source plus exact My Friday, Codex, and Code Mode
  executables, and verifies/removes only manifest-owned state.
- `internal/codexhome` implements previewed install, verify, repair, upgrade,
  rollback, uninstall, and journaled recovery for a narrowly owned projection.
- `internal/capability` and `internal/capabilityworkshop` prove source/projection
  separation, exact confirmation, deterministic validation, and recoverable
  instruction-only lifecycle behavior.
- `cmd/my-friday/main.go` exposes the current split `init`, repository-pair
  validation, named-assistant lifecycle, Codex lifecycle, and capability
  lifecycle. Stable error classes already distinguish input, path, contract,
  transaction, and installed-state failures.
- `docs/architecture/repository-bootstrap.md`,
  `docs/architecture/named-assistant-instances.md`, and
  `docs/architecture/capability-workshop.md` are the durable contracts that B1
  must evolve rather than bypass.

The current system does not have one canonical assistant repository, baseline
or module migration ledgers, bounded commit/push authority, remote divergence
handling, or a host binding whose source of truth remains outside the generated
Codex instance.

## Actors And Critical Journeys

### Actors

- **User/operator:** chooses the assistant name, canonical path, already-created
  private remote, migration source, and explicit mutation confirmations.
- **Bootstrap kernel:** plans, validates, stages, commits, pushes, promotes,
  reconciles, diagnoses, and recovers only within declared authority.
- **Repository steward:** the kernel role allowed to make exact validated Git
  commits and explicit pushes; it has no B2 capability-activation or B3 memory-
  promotion authority.
- **Git remote:** durability/synchronization boundary for one configured branch;
  it is not a source of provider identity, privacy truth, or merge decisions.
- **Codex harness:** executes the assistant from a generated projection. It is a
  consumer of compiled state, not the owner of canonical source.

### Primary create flow

1. The user invokes `my-friday assistant create NAME --repository PATH
   --remote URL --remote-private` from an interactive terminal. The last flag
   explicitly attests that the already-created remote is private.
2. My Friday runs the released profile questions for display name, form of
   address, purpose, and communication style, preserving their normalization,
   limits, and policy boundary. It then validates the name, canonical path, Git
   author, remote syntax and reachability, explicit privacy attestation, empty
   remote, dependencies, and collision-free host binding without changing
   either side.
3. The preview shows normalized non-secret profile values and groups canonical
   repository files, local generated state, Git
   commit/push, preserved state, prohibited effects, and the exact recovery
   location. Only exact `Create` proceeds.
4. The kernel creates and validates a staged repository, commits the complete
   baseline, pushes the exact commit to `origin/main`, atomically promotes the
   staged repository into the empty destination, generates and verifies the
   Codex projection, and removes its journal. The remote-first create order
   ensures a machine lost after push can be restored from the remote.
5. The command reports repository, commit, remote name/ref, baseline version,
   projection state, and the launcher to use.

### Return-use and diagnosis flow

`assistant inspect NAME` gives a stable, read-only state summary even when the
assistant is unhealthy. `assistant verify NAME` is strict and nonzero unless
repository contracts, Git synchronization, host binding, dependencies, and
projection all agree. `assistant diagnose NAME` adds exact refusal reasons and
the one safe next command: reconcile, repair, recover, restore, migrate, or
resolve a remote divergence manually.

### Clean-host restore flow

`assistant restore NAME --repository PATH --remote URL --remote-private`
requires an empty local target and a nonempty remote whose `main` tip is a valid
canonical baseline. It previews the source commit and generated host effects,
then exact `Restore` clones and validates the repository without creating a
semantic commit, builds the host binding and Codex projection, verifies them,
and reports the launcher. A failed restore never changes the remote. This is
also the normal reconstruction path after a host is lost.

### Mutation and recovery flow

Upgrade, rollback, repair, and later B2/B3 semantic writes all use the same
outer transaction. The user sees a deterministic preview and exact confirmation
before mutation. Cancellation and validation denial make no change. An
interruption leaves one authenticated journal; rerunning an ordinary mutation
refuses and names `assistant recover NAME`. Recovery reconstructs the observed
phase and either completes the already-authorized operation or restores the
proven predecessor without creating a second commit.

### Migration flow

The user supplies a validated legacy runtime/memory pair, destination path, and
private remote. Preview shows the exact mapping and states that legacy source
and history will remain. My Friday imports profile and instruction-only source
into `config/` and `capabilities/`, imports memory scaffolding/data into
`memory/`, records legacy HEAD/tree provenance, creates and pushes the new
repository, switches the named projection only after verification, and leaves
the old pair untouched. A failure before verified switch keeps the old launcher
active; a failure after switch is journal-recoverable.

### Removal flow

`assistant remove NAME` previews removal of only the launcher, generated Codex
projection, copied executables, and local binding. It prominently states that
the canonical repository and remote remain. Exact `Remove` detaches the host
after full manifest verification. Drift or an unknown entry is preserved and
refused. B1 provides no flag that recursively deletes canonical data.

## Acceptance And Non-Goals

The issue acceptance is designable through four groups:

1. one repository and independently governed module manifests;
2. full baseline/host lifecycle with explicit inspect, diagnose, and recovery;
3. exact-path, private-remote Git commit/push and idempotent reconciliation; and
4. denial of divergence, ambiguity, force, rewrite, publication, collision, or
   unrelated-state mutation.

Non-goals:

- B2 capability component schemas, compiler feature matrices, dependency
  resolution, or activation semantics;
- B3 memory record schemas, sensitivity authorization, retrieval, or promotion;
- creating a GitHub/GitLab/provider repository, determining remote visibility,
  logging into a provider, or managing credentials;
- multi-writer or distributed locking beyond fail-closed Git ref comparison;
- automatic conflict resolution, pull/rebase, force push, branch management,
  submodules, Git LFS, signed commits, or public distribution;
- a background daemon, VM/container runtime, second harness, or non-Apple-
  silicon production support; and
- deletion of canonical source, durable memory, remote state, or retained
  legacy repositories.

## Constraints, Dependencies, And Risks

- F0/#74 must be accepted and released so migration imports a stable existing
  lifecycle contract.
- B1 continues the Apple-silicon macOS and native Go boundary. Portable tests
  run in the development container, but APFS, terminal, process, and exact
  candidate evidence remain native requirements.
- Git network and credential helpers are external effects. Commands must use
  fixed argument forms, sanitized errors, bounded timeouts, and no URL-embedded
  credentials. The kernel does not suppress legitimate SSH or credential-
  helper prompts during interactive setup, but noninteractive lifecycle writes
  fail rather than hang.
- A remote may become public after attestation. Verification can prove only
  endpoint identity and ref state; documentation must say privacy remains the
  user's/provider's responsibility.
- Atomic filesystem promotion and Git remote publication cannot be one atomic
  transaction. The operation journal and phase-specific reconciliation are
  therefore part of correctness, not cleanup detail.
- Canonical memory can grow. B1 should avoid whole-history rewrites and
  byte-for-byte copies of prior histories; performance thresholds and large-
  object policy need deterministic tests before release.
- Current generated instances copy source into the instance root. B1 changes
  that ownership relation and must migrate without letting remove/recovery
  authority reach the canonical repository.
- Terminal output must remain line-oriented, width-tolerant, keyboard-only,
  screen-reader coherent, and non-colour-dependent. Confirmation tokens are
  English protocol literals; descriptive copy must keep paths and actions on
  separate lines and avoid relying on punctuation or colour.

## Evidence, Assumptions, And Unknowns

### Evidence

- The approved discovery pack defines the one-repository, narrow-kernel,
  verified Git, Codex-first, and source/projection boundaries.
- Existing transaction, instance, Codex, and capability tests prove reusable
  no-follow ownership, manifest, journal, recovery, and preview patterns.
- Existing releases use immutable candidate nomination, independent native
  acceptance, and GitHub release ledgers; B1 can use that path without inventing
  staging for a downloadable artifact.

### Assumptions

- Initial users can create a private empty remote and configure Git/SSH
  credentials before running setup.
- One designated `origin/main` is sufficient for the MVP; other branches and
  remotes are user-owned but cannot participate in automatic stewardship.
- Regular files and directories cover B1-owned paths. Symlinks, submodules,
  special files, and external filters are unnecessary for the baseline modules.
- Preserving legacy repositories rather than combining their histories is an
  acceptable and safer migration result.

### Unknowns to validate during implementation

- The practical upper bound at which staging/validation becomes too slow for a
  technically capable first cohort.
- Which Git/SSH credential prompts appear on clean supported machines and
  whether the preflight copy is sufficiently clear without provider-specific
  onboarding.
- Whether design partners prefer the explicit `diagnose` verb or find
  `inspect` plus actionable failures sufficient. Keep both for acceptance and
  measure comprehension; removing a redundant read-only alias would not alter
  the trust boundary.

None of these unknowns changes the approved outcome or prevents implementation.
