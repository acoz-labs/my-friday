# Solution Design: Canonical Assistant Repository And Bootstrap Kernel

- **Status:** Final
- **Issue:** #92
- **Planning PR:** #96
- **Repository basis:** 2594147084e8bb6f961971efdb116b02824ecf66
- **Execution envelope:** through-production

## Decision

Replace the split-repository bootstrap as the default with one versioned,
private assistant repository containing independently governed `config/`,
`memory/`, and `capabilities/` modules. A narrow portable kernel will create or
migrate that repository, bind any number of identified installations, refresh
canonical state before a fresh task, reconcile a replaceable Codex projection
per installation, and perform each verified source mutation as one clean-start
Git transaction against an explicitly configured `origin/main`.

The kernel stages and validates an exact candidate before changing the active
repository, writes only declared paths, creates a normal forward commit, and
pushes that exact commit without force or implicit merge. Remote ref
compare-and-swap is the cross-installation serialization boundary: a competing
writer causes the losing operation to preserve its receipt and refuse, except
for a content-addressed append that can be reconstructed and fully revalidated
against the new head without changing meaning. A local operation journal makes
commit-success/push-failure and push-success/promotion-failure recoverable
without repeating the semantic write. Existing split repositories and their
history remain untouched during migration and serve as the rollback source
until the user deliberately retires them.

## Needs Attention

- Implementation and release remain sequenced behind acceptance and release of
  the instruction-only foundation tracked by #74 and its B0/F0 lifecycle issue
  #83. This is a release prerequisite, not an unresolved design decision.
- Git can prove the configured remote and ref state, but cannot prove that an
  arbitrary provider repository is private. Setup therefore requires an
  explicit private-remote attestation, records only a hash of the endpoint in
  receipts, and never creates or publishes a hosted repository.
- B1 establishes multi-installation synchronization, freshness, host identity,
  and safe Git arbitration. B2 and B3 still own capability semantics and
  governed-memory record/promotion semantics; B1 must not pretend an empty
  memory module already provides cross-task continuity.

## Decision Spotlight

- **One canonical repository, three governed modules.** `config/`, `memory/`,
  and `capabilities/` share one Git history for portable recovery, while each
  retains its own schema version, authority, migration, validation, and
  removal rules. This is the approved B1 product boundary, not a filesystem
  convenience.
- **The repository is the assistant; installations are projections.** Every
  host or VM receives a unique installation ID, local capability inventory,
  secret-slot binding, and generated harness projection. Removing or losing an
  installation cannot remove assistant identity, source, or memory.
- **Refresh before fresh-task recall.** The launcher fetches and fast-forwards a
  clean managed checkout before compiling and starting a fresh task. When the
  remote is unreachable, the user sees the last verified commit and age and
  must explicitly accept stale launch; automation cannot silently do so.
- **Git arbitrates concurrent writers; it does not merge their judgment.** A
  non-force push is the remote compare-and-swap. Ordinary config, capability,
  migration, and durable-memory transitions refuse on an advanced remote.
  Only a newly created immutable, content-addressed append may be replanned on
  the new head when its meaning and complete validation are unchanged.
- **Multiple interactive installations may coexist; singleton effects may not.**
  B1 records installation identity and declared operating role. Later
  capabilities must explicitly declare whether an effect is local,
  multi-active, or singleton; B1 provides no distributed lease or hidden
  scheduler.
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
- **A VM is an optional installation profile.** Native machines, headless
  hosts, and VMs use the same repository/binding/projection contract. VM disks,
  browser sessions, and host credentials are never canonical assistant state.
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

The plan is `Final` only after independent maintainer review confirms that the
Git transaction, multi-installation synchronization, migration, interaction,
recovery, and release contracts are complete and no blocking finding remains.
The product authority approved the repository-first, shared-brain flow and
instructed delivery to proceed without repeated ordinary gates; the maintainer
must verify that the exact planning head contains no material choice beyond
that signal before recording approval and merging.
