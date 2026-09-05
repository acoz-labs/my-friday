# Solution Design: Canonical Assistant Repository And Bootstrap Kernel

- **Status:** Draft
- **Issue:** #92
- **Planning PR:** #96
- **Repository basis:** 2594147084e8bb6f961971efdb116b02824ecf66
- **Execution envelope:** through-production

## Decision

Replace the split-repository bootstrap as the default with one versioned,
private assistant repository containing independently governed `config/`,
`memory/`, and `capabilities/` modules. A narrow native kernel will create or
migrate that repository, reconcile a replaceable named Codex projection, and
perform each verified source mutation as one clean-start Git transaction
against an explicitly configured `origin/main`.

The kernel stages and validates an exact candidate before changing the active
repository, writes only declared paths, creates a normal forward commit, and
pushes that exact commit without force or implicit merge. A local operation
journal makes commit-success/push-failure and push-success/promotion-failure
recoverable without repeating the semantic write. Existing split repositories
and their history remain untouched during migration and serve as the rollback
source until the user deliberately retires them.

## Needs Attention

- Implementation and release remain sequenced behind acceptance and release of
  the instruction-only foundation tracked by #74 and its B0/F0 lifecycle issue
  #83. This is a release prerequisite, not an unresolved design decision.
- Git can prove the configured remote and ref state, but cannot prove that an
  arbitrary provider repository is private. Setup therefore requires an
  explicit private-remote attestation, records only a hash of the endpoint in
  receipts, and never creates or publishes a hosted repository.

## Decision Spotlight

- **One canonical repository, three governed modules.** `config/`, `memory/`,
  and `capabilities/` share one Git history for portable recovery, while each
  retains its own schema version, authority, migration, validation, and
  removal rules. This is the approved B1 product boundary, not a filesystem
  convenience.
- **Remote-backed creation is the B1 default.** The first canonical commit must
  push to an already-created, explicitly attested private remote. Local-only
  and hosted-provider creation are outside B1 because the approved outcome
  requires recoverable remote ownership without provider-specific policy. The
  required `--remote-private` flag is an attestation, not a provider check.
- **Restore is distinct from create.** `assistant restore` accepts an existing
  valid canonical `origin/main`, creates no semantic commit, and rebuilds only
  the local checkout, binding, dependencies, launcher, and Codex projection.
  This is the clean-host recovery path and prevents nonempty remotes from being
  misclassified as create collisions.
- **Source survives removal.** Ordinary `assistant remove` deletes only the
  manifest-proven launcher and generated harness state; it detaches but never
  deletes the canonical repository or remote. Destructive repository deletion
  is excluded from this outcome.
- **Rollback moves history forward.** Baseline rollback applies a compatible
  inverse migration and commits it as a new descendant. The kernel never
  resets, rebases, rewrites, force-pushes, or guesses through divergence.
- **Git authority is narrower than filesystem access.** Every mutation starts
  clean, fetches the designated ref, stages only plan-declared paths, validates
  the complete affected contracts, and pushes `HEAD:refs/heads/main` without
  force. Untracked files, another branch, detached HEAD, hooks/filters on owned
  control paths, or ref disagreement refuse before mutation.
- **Migration is copy-and-prove, not history surgery.** A validated legacy
  runtime/memory pair is imported into the new module layout with source
  HEAD/tree provenance. Old repositories are preserved; combining unrelated
  Git histories or deleting old state would add risk without improving the
  accepted journey.
- **Generated Codex state is replaceable.** The canonical repository owns
  source and policy; `$HOME/.my-friday/assistants/<name>` owns a host binding,
  launcher, copied executable dependencies, and generated Codex projection.
  Deleting generated state cannot delete canonical capability source or memory.
- **No hidden daemon or secret store.** B1 remains an invoked native kernel.
  It delegates transport authentication to Git's existing credential/SSH
  mechanisms, rejects credential-bearing URLs, and never reads or persists
  credential values.
- **Same-user isolation remains honest.** Manifests, modes, no-follow checks,
  directory locks, receipts, and exact-path plans prevent accidental or stale
  mutation; they do not claim resistance to a malicious process running as the
  same operating-system user.

## Plan Map

- [Context](01-context.md)
- [Decision](02-decision.md)
- [Design](03-design.md)
- [Verification](04-verification.md)
- [Implementation handoff](05-handoff.md)

## Final Gate

The plan may become `Final` after independent maintainer review confirms that
the Git transaction, migration, interaction, recovery, and release contracts
are complete and no blocking finding remains. Product authority must then
approve this exact planning head and its `through-production` envelope before
the plan merges or implementation begins.
