# Context

## Problem And Desired Outcome

Issue [#4](https://github.com/acoz-labs/my-friday/issues/4) requires multiple
named assistant instances to coexist under one ordinary macOS user without
cross-contaminating Codex configuration, runtime data, memory, workspaces,
managed dependencies, credentials, or existing user state.

The desired outcome is a native, inspectable instance boundary. A user can
create, inspect, launch, verify, and remove one named instance without changing
`HOME`, shell configuration, the user's existing Codex installation, or any
other instance. The one deliberate projection outside the instance root is a
manifest-owned native launcher at `$HOME/.local/bin/<name>`; no other outside
entry is mutable. A prior O2 projection can be migrated only through a staged,
verified, manifest-proven transition.

## Current State

The repository basis is
`b54ff42f236f0b9cd3438af36f59bdca3ce44c09`.

- `cmd/my-friday/main.go` is the native command entry point.
- `internal/codexhome/lifecycle.go` owns the existing Codex-home lifecycle and
  its manifest-governed cleanup boundary.
- `internal/environment/contract.go` defines environment handling that the new
  launcher must preserve and narrow.
- `internal/plan/plan.go`, `internal/transaction/transaction.go`, and their
  tests provide planning, promotion, rollback, and recovery patterns.
- `docs/decisions/0002-manifest-owned-codex-baseline.md` records the existing
  manifest-ownership precedent. The new design extends that principle to a
  complete named instance rather than treating caller-selected environment
  variables as an ownership boundary.
- `config/acceptance/codex-smoke.sb.in` and
  `config/acceptance/lifecycle.sb.in` provide acceptance boundaries that can be
  extended for exact-path containment and a separately credentialed live smoke.

No existing user Codex or shell state becomes managed merely because it is
present. The repository basis is evidence of current implementation shape, not
authority to inspect or adopt arbitrary user files.

## Actors And Critical Journeys

### Current user

- Creates a valid named instance and receives an exact, inspectable receipt of
  the instance root and external launcher. For `alfred`, the launch path is
  `$HOME/.local/bin/alfred`.
- Creates a second instance and can launch either with separate Codex, runtime,
  memory, workspace, dependency, and credential state.
- Launches through the native launcher; `HOME` remains the real current-user
  home, `CODEX_HOME` is instance-local, and Codex starts in the instance
  workspace through `--cd`.
- Encounters a missing or unsafe launcher directory, pre-existing launcher,
  invalid name, incomplete manifest, or foreign ownership and receives a
  refusal without mutation.
- Removes one manifest-owned instance while unrelated user state and other
  instances remain byte-for-byte and metadata-stable where observable.

### Migrating user

- Stages the named replacement beside the prior O2 projection.
- Verifies structure, manifest, the exact external launcher projection,
  containment, and credential-free launch before cleanup.
- Deletes only old paths exactly enumerated and proven by the prior manifest;
  missing or conflicting proof preserves the prior projection for recovery.

### Independent acceptor

- Runs the credential-free lifecycle and exact-path containment matrix under a
  current non-root user.
- Runs a separate live Codex smoke with independently provisioned file-backed
  credentials inside the instance `CODEX_HOME`, retaining only redacted
  pass/fail evidence.

## Acceptance And Non-Goals

Acceptance is grouped as follows:

1. deterministic, validated names map to roots beneath
   `$HOME/.my-friday/assistants/` and cannot escape by traversal, symlink, case,
   Unicode, or path collision;
2. each instance owns `codex/`, `runtime/`, `memory/`, `workspace/`, managed
   dependencies, and its manifest beneath the root, plus exactly one external
   native launcher at `$HOME/.local/bin/<name>`;
3. the launcher sets instance `CODEX_HOME` and Codex `--cd` while preserving
   `HOME` and avoiding shell startup or alias changes;
4. two or more instances coexist with distinct external launcher names, launch
   independently, and survive lifecycle operations on a sibling;
5. migration stages and verifies the replacement before deleting only prior
   O2 paths proven by the prior manifest; and
6. current-user acceptance proves exact containment, a credential-free
   lifecycle, and a separate credentialed live smoke without credential
   disclosure.

Non-goals are macOS user creation; `HOME` substitution; aliases, shell
functions, shell startup or PATH edits; caller-selected `CODEX_HOME`; creating,
changing, or adopting `$HOME/.local/bin`; any second projection outside the
instance root; shared mutable instance directories; credential acquisition or
secret storage by My Friday; automatic adoption or deletion of unmanifested
state; compatibility with arbitrary pre-O2 layouts; background services;
containers or VMs as the instance boundary; implementation, merge, migration
execution, or release in this planning PR.

## Constraints, Dependencies, And Risks

- The current user owns the instance tree. Normal filesystem permissions are
  necessary but not sufficient; manifest ownership and canonical containment
  decide mutation and deletion authority.
- Names must use a single documented canonical grammar and reserved-name set.
  Validation occurs before path construction; canonicalized target and every
  managed child must remain beneath the assistants root without symlink
  traversal. The sole exception is the separately derived launcher leaf at
  `$HOME/.local/bin/<name>`.
- `$HOME/.local/bin` must already be a real current-user-owned directory with
  safe permissions and no symlink traversal. Creation may mutate only the exact
  launcher directory entry; it does not create, chmod, rewrite, or adopt the
  parent or any sibling. Any existing leaf is a collision unless the same valid
  instance manifest proves ownership and artifact identity.
- The launcher must pass arguments as literal argv. It must not source a shell,
  expand aliases, serialize credentials, or infer configuration from ambient
  caller-selected `CODEX_HOME`.
- `HOME` remains available to Codex and dependencies, so the design does not
  claim that arbitrary third-party programs can never read user-home state.
  The guaranteed product boundary is the instance root, its one exact external
  launcher leaf, launcher variables, invocation directory, managed
  dependencies, and mutation scope.
- Credentials are sensitive mixed-ownership state. Tests use no live value;
  the live smoke injects a separately held credential into the instance Codex
  configuration without exposing it to My Friday receipts or evidence.
- Migration deletion is irreversible enough to require fail-closed ownership,
  staged verification, an explicit deletion plan, and recoverable interruption.

## Evidence, Assumptions, And Unknowns

### Evidence

- Issue #4 supplies the approved product boundary and acceptance intent.
- The repository basis and paths listed above provide native lifecycle,
  manifest, transaction, environment, and acceptance precedents.
- The fixed solution decision requires per-name roots, the sole external
  launcher projection at `$HOME/.local/bin/<name>`, unchanged `HOME`, instance
  `CODEX_HOME`, Codex `--cd`, multi-instance coexistence, and manifest-proven O2
  cleanup.

### Assumptions

- A single current-user installation is the supported operating model.
- File-backed Codex configuration beneath `CODEX_HOME` is sufficient for
  separate instance credentials.
- Managed dependencies can be resolved and invoked from instance-owned runtime
  state without modifying user-global package or shell configuration.
- Existing transaction and manifest primitives can be extended without a new
  daemon or database.

### Validation needs, not blocking unknowns

- Implementation must pin the exact name grammar and manifest schema version
  while preserving the fixed launcher mapping in this plan.
- Native acceptance must prove Codex honors the instance-local file-backed
  credential configuration in a separately credentialed live smoke.
- Fault injection must establish the exact recoverable states for interrupted
  staging, promotion, and cleanup before migration is enabled.
