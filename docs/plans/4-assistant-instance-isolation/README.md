# Solution Design: Isolate named assistant instances

- **Status:** Final
- **Issue:** #4
- **Planning PR:** #30
- **Repository basis:** b54ff42f236f0b9cd3438af36f59bdca3ce44c09
- **Execution envelope:** implementation

## Decision

Create manifest-owned named assistant instances beneath
`$HOME/.my-friday/assistants/<name>/`. Each instance exclusively owns its
`codex/`, `runtime/`, `memory/`, `workspace/`, and managed-dependency state.
My Friday creates exactly one managed projection outside the instance root: a
collision-checked regular native launcher at `$HOME/.local/bin/<name>`. Thus an
instance named `alfred` is directly launchable as `$HOME/.local/bin/alfred`
without a shell edit. The launcher sets `CODEX_HOME` to the instance's `codex/`
directory and invokes Codex with `--cd` pointing at the instance's
`workspace/`, while deliberately leaving `HOME` unchanged.

File-backed credentials are configured inside that instance's `CODEX_HOME`.
Existing user Codex configuration, shell startup files, aliases, environment,
home contents, and unrelated instances remain outside My Friday's mutation and
deletion authority. Outside the instance root, lifecycle authority is limited
to the exact manifest-owned launcher leaf; the pre-existing `.local/bin`
directory and every other path remain unmanaged. Multiple instances may
coexist and run independently.

Migration is staged and verified before cleanup. My Friday may delete only the
prior O2 projection whose exact paths and ownership are proven by its earlier
manifest; absent or contradictory proof fails closed and leaves the old state
in place.

## Needs Attention

- Implementation acceptance needs a current-user exact-path containment run
  with canaries around the instance root, the external launcher and its sibling
  entries, user Codex home, shell files, and a second instance.
- The credential-free lifecycle matrix must pass before a separately
  credentialed live Codex smoke is attempted. No credential value may enter
  logs, manifests, fixtures, snapshots, or pull-request evidence.
- Migration cleanup is authorized only by exact prior-manifest proof and only
  after the staged replacement passes verification.

## Decision Spotlight

- **Per-instance root under the existing user's home.** Named instances live at
  `$HOME/.my-friday/assistants/<name>` so ownership is inspectable and backup,
  permissions, and current-user operation remain conventional.
- **`HOME` is invariant.** The launcher scopes Codex through `CODEX_HOME` and
  `--cd`; it never substitutes `HOME`, creates a macOS user, or attempts to
  simulate a second login environment.
- **One external native launcher is managed state.** The sole allowed
  projection is the regular executable `$HOME/.local/bin/<name>`, recorded by
  exact absolute path and artifact identity in the instance manifest. The
  launcher directory must already exist and is never created, chmodded, edited,
  or adopted. An unowned launcher or mismatched manifest is preserved and
  reported, never overwritten; no other outside entry may change.
- **Credentials belong to the instance Codex home.** File-backed credential
  configuration lives beneath `codex/`; My Friday manages the location and
  lifecycle boundary but never records credential values in its manifest or
  evidence.
- **Instances do not share mutable state.** Runtime, memory, workspace, Codex
  configuration, and managed dependencies are contained beneath one instance
  root; only that root's launcher is projected externally. A lifecycle
  operation on one name cannot discover, mutate, or remove another.
- **Migration is promote-before-delete.** The replacement is staged and
  verified first. Cleanup is limited to exact paths proven to belong to the
  prior O2 projection by its manifest; uncertainty preserves data.
- **Existing user state is authoritative.** User Codex state, shell files,
  aliases, environment, and unrelated home contents are never adopted,
  rewritten, or used as a launcher mechanism.

## Plan Map

- [Context](01-context.md)
- [Decision](02-decision.md)
- [Design](03-design.md)
- [Verification](04-verification.md)
- [Implementation handoff](05-handoff.md)

## Final Gate

The plan may become `Final` only after the draft planning PR is linked to issue
#4, the complete pack has no blocking maintainer finding, validation passes,
and the PR number is recorded. Product authority must approve the exact final
head before merge. This planning task performs no implementation, merge,
migration, or release.
