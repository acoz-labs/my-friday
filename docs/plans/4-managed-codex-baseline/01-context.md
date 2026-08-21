# Context

## Problem And Desired Outcome

Issue [#4](https://github.com/acoz-labs/my-friday/issues/4) selects discovery
outcome O2: a user can install, verify, repair, uninstall, and roll back a
minimal Codex baseline without altering unrelated configuration. The safety
claim is the product. Merely copying an instruction file is insufficient if a
collision, interruption, mistaken root, or later uninstall can damage an
existing Codex installation.

The desired result is a bounded foreground lifecycle for one inspectable global
instruction projection. Every intended write is visible before confirmation;
ownership and exact bytes are machine-verifiable; recovery never guesses; and
an install followed by uninstall restores the pre-install filesystem state.

## Current State

The repository basis is
`5bc309226d2c40e1473a4011c1bd8552c995919d`.

- `cmd/my-friday/main.go` exposes `init`, `validate`, `recover`, and `version`.
  It has no installed-Codex lifecycle.
- `internal/plan/plan.go` creates a runtime repository containing
  `AGENTS.md`, `assistant/profile.json`, a repository manifest, and embedded
  schemas. The runtime instruction file refers to profile presentation data
  and explicitly preserves authorization, safety, trust, privacy, and tool
  policy.
- `internal/transaction/transaction.go` provides useful precedent for durable
  journals, marker ownership, exact-tree proofs, atomic promotion, injected
  faults, and conservative recovery across the two-repository creation flow.
  Installed-state management needs its own smaller transaction because it
  operates inside a pre-existing, mixed-ownership Codex home.
- `internal/transaction/transaction_test.go` and
  `internal/terminal/evidence_test.go` demonstrate temporary-root tests, a
  published fault matrix, adjacent-state snapshots, and repeatable recovery.
- `docs/architecture.md` currently says My Friday has no global installer, and
  `docs/architecture/repository-bootstrap.md` says repository creation performs
  no global Codex change. Both remain true for `init` but the system overview
  must gain a separate installed-baseline capability during implementation.
- `docs/deployment.md` declares the repository an artifact and requires native
  Apple silicon/APFS acceptance. Its release-prerequisite paragraph is stale
  relative to the existing accepted release
  `artifact-2026.08.21-5bc3092` and must be reconciled from shipped behavior.
- `docs/product.md` already records installed-baseline lifecycle as the second
  selected outcome and requires explicit ownership, preview, collision
  handling, verification, and reversal for each projection.

Official OpenAI documentation consulted on 2026-08-21 establishes the external
surface:

- [Config basics](https://learn.chatgpt.com/docs/config-file/config-basic)
  says user configuration lives in `~/.codex/config.toml`, project overrides
  live in trusted `.codex/config.toml` layers, and the CLI and IDE extension
  share those layers.
- [Environment variables](https://learn.chatgpt.com/docs/config-file/environment-variables)
  says `CODEX_HOME` defaults to `~/.codex` and roots config, auth, logs,
  sessions, skills, and standalone package metadata.
- [AGENTS.md discovery](https://learn.chatgpt.com/docs/agent-configuration/agents-md)
  says the global instruction source is `AGENTS.override.md` or `AGENTS.md`
  inside `CODEX_HOME`, with only the first non-empty global file used.

Those facts make Git isolation insufficient: a safe branch or worktree can
still mutate the same live user-level Codex home as Alfred. They also constrain
the minimal useful projection: a managed `AGENTS.md` is effective only when a
foreign `AGENTS.override.md` is absent, and all other state under `CODEX_HOME`
must be treated as unrelated.

## Actors And Critical Journeys

### Technical user

- **Install:** points at a validated My Friday runtime repository, sees the
  exact Codex home, source, projection digest, created control files, and
  prohibited effects, then types `Install`.
- **Verify:** receives a read-only result that distinguishes healthy, not
  installed, collision, managed drift, source drift, incompatible contract,
  interrupted transaction, and shadowing `AGENTS.override.md`.
- **Repair:** previews a replacement of a manifest-owned but drifted projection
  and confirms `Repair`; foreign state remains untouchable.
- **Upgrade:** after the runtime source changes compatibly, previews old and new
  source/generation identities and installs the new managed projection while
  retaining one rollback generation.
- **Rollback:** previews and restores the immediately previous verified managed
  generation, without changing unrelated files.
- **Uninstall:** removes only an exact manifest-owned projection and My Friday
  control state; any drift blocks removal until the user chooses a safe path.
- **Recover:** follows a printed transaction-specific command after interruption
  and receives an idempotent completed, rolled-back, or refused result.

### Independent acceptor

Builds one immutable candidate from the nominated commit, exercises all
lifecycle journeys under a fresh disposable non-admin macOS user/home, proves a
real Codex run discovers the projected instruction, compares unrelated canary
state before and after, records the candidate digest and environment evidence,
then approves or rejects that exact artifact.

### Maintainer and release operator

Review the trust boundary, fault coverage, documentation promotion, artifact
nomination, independent acceptance, rollback evidence, and release receipt.
They must never substitute contributor tests against a live home for fresh
candidate acceptance.

## Acceptance And Non-Goals

The issue acceptance is designable as five groups:

1. one deterministic projection and one ownership/control namespace;
2. preview and explicit confirmation before every mutation;
3. safe denial for collisions, path ambiguity, shadowing, source incompatibility,
   and unowned or drifted state;
4. durable, idempotent recovery and single-generation rollback; and
5. complete install/uninstall reversal with no daemon or background service.

Non-goals are installing or updating the Codex executable; editing or merging
`config.toml`; managing auth, sessions, logs, skills, packages, profiles,
project `.codex` directories, system config, or `/etc`; supporting arbitrary
Unix roots, network homes, non-APFS volumes, Intel Macs, Linux, or Windows;
background reconciliation; privilege escalation; arbitrary runtime templates;
and outcomes O3 through O8.

## Constraints, Dependencies, And Risks

- O1 contract-v1 runtime repositories are the only source input. A source must
  validate as role `runtime`; memory repositories and arbitrary folders fail.
- The pilot environment remains macOS 14+, Apple silicon, local APFS, Git
  2.28+, UTF-8 terminal, and a non-root interactive user.
- The effective `CODEX_HOME` may contain credentials and histories. My Friday
  may inspect only path metadata, the two global instruction filenames, and its
  own control namespace; it must not enumerate or read unrelated content.
- `AGENTS.override.md` shadows `AGENTS.md`. Installation fails before mutation
  if a non-empty override exists rather than claiming the baseline is active.
- A process can race filesystem ancestors after checks. File-descriptor-relative
  operations, non-following opens, an exclusive lock, revalidation immediately
  before rename, and exact inode/device checks reduce the risk. The product
  does not claim an adversarial sandbox against an administrator replacing
  ancestors.
- Paths and file digests are local operational metadata. Journals and manifests
  use owner-only modes and must never contain profile prose, auth material,
  tokens, Codex config, or unrelated file listings.
- Release acceptance requires a clean identity boundary, not just a clean Git
  checkout. Creating and deleting the disposable user is operator work outside
  the My Friday binary and must be documented as a release runbook.

## Evidence, Assumptions, And Unknowns

### Evidence

- Approved discovery PR `acoz-labs/.github#2` at
  `b6db62bf15c8d6ad7a15f7533e6aa5981ae1cd8a`, outcome O2, supplies product
  authority.
- Repository paths and official documentation cited above establish the
  current product and Codex surfaces.
- Local read-only inspection found `codex-cli 0.149.0`; compatibility acceptance
  must record the exact tested version rather than claim all future versions.
- GitHub contains an accepted artifact release for the repository basis and
  active artifact nomination, acceptance, and release workflows.

### Assumptions

- A single global instruction projection is sufficient to prove the initial
  installed-baseline value before skill or memory projections exist.
- Users will accept a safe collision refusal instead of automatic prose merging.
- One rollback generation is enough for the initial lifecycle contract.
- A self-contained renderer can preserve the validated profile's presentation
  semantics without including arbitrary repository prose.

### Known release capability gaps

- The exact Codex CLI patch version present in the disposable acceptance user
  is recorded as evidence; expanding a tested compatibility range is a later
  evidence-backed change, not a blocker to this plan.
- No configured macOS harness currently creates a disposable UID, home,
  keychain, isolated `CODEX_HOME`, named test-credential injection, evidence,
  and teardown; current acceptance automation only records a decision.

These gaps do not block implementation, but they prohibit acceptance/release.
A later design must name the host, commands, secret slot/source, immutable
package store, digest checks, upload/download path, teardown, and receipt.
