# Solution Design: Inspectable Capability Builder

- **Status:** Draft
- **Issue:** #51
- **Planning PR:** Pending
- **Repository basis:** f35f08fb07e51299657841ac626ebdfba492e80e
- **Execution envelope:** through-production

## Decision

Ship a narrow, source-first `instruction-only` capability profile and a built-in
capability-builder skill. Existing runtime repositories and named assistants
move to contract v2 through two explicit, independently recoverable migrations:
`capability initialize` evolves the runtime source contract, then `assistant
upgrade` installs the builder into that assistant's repository-scoped Codex
skill projection. User-authored packages remain versioned under
`runtime/skills/<slug>/`; only their reviewed `skill/` subtree is copied into
the named assistant workspace after deterministic checks and exact confirmation.

This is the smallest coherent route that lets an assistant help build its first
capability without making source and installed state the same thing or granting
the agent activation authority.

## Needs Attention

None.

## Decision Spotlight

- **The builder is a bootstrap-owned core skill, not the first ordinary user
  capability.** New contract-v2 instances receive it at creation; existing
  instances receive it only through explicit `assistant upgrade`. This resolves
  the bootstrap loop without special-casing user packages.
- **Runtime and instance upgrades are separate transactions.** The runtime is
  initialized first and remains the source of truth; the instance upgrade then
  consumes that verified source. A failure cannot leave two roots under one
  ambiguous transaction.
- **User capabilities install into
  `<instance>/workspace/.agents/skills/<slug>`.** Codex discovers this scope from
  the launcher's fixed workspace, preserving named-instance isolation instead of
  writing the user's global `$HOME/.agents/skills` directory.
- **Installed user capabilities are explicit-invocation-only in C1.** My Friday
  renders a fixed `agents/openai.yaml` with implicit invocation disabled. The
  built-in builder may match ordinary authoring requests, but newly authored
  instructions cannot silently enter unrelated conversations.
- **“Instruction-only” is a structural profile, not a safety certification.**
  My Friday forbids scripts, dependencies, credentials, network declarations,
  executable entries, links, devices, and unknown files, but cannot prove that
  natural-language instructions are benign. The complete diff and exact user
  confirmation remain essential.
- **Tests are deterministic package-contract tests.** They verify schemas,
  triggers/non-triggers, examples, projections, and prohibited effects without
  invoking a nondeterministic model. Exact-candidate acceptance separately
  exercises the installed skill with Codex.
- **Mutation authority stays in the CLI.** The builder can scaffold, edit,
  inspect, validate, and test, but its instructions forbid mutating lifecycle
  commands or entering confirmation tokens. A TTY and exact token prevent
  accidents; they are not proof that a human typed the token.
- **Disable and remove affect managed projections, never source.** Disable
  retains the receipt for re-enable and audit; remove clears instance control
  state only after ownership proof. Repository source remains user-owned.
- **Lifecycle changes become effective for a fresh Codex task.** My Friday does
  not claim it can unload or replace instructions already resident in a running
  model session.

## Plan Map

- [Context](01-context.md)
- [Decision](02-decision.md)
- [Design](03-design.md)
- [Verification](04-verification.md)
- [Implementation handoff](05-handoff.md)

## Final Gate

The final plan requires Anthony's approval for the `through-production`
execution envelope at the exact planning PR head. That approval authorizes a
later implementation task to implement, independently accept, merge, and
release only this plan; it does not itself dispatch implementation.
