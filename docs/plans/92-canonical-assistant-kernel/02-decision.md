# Solution Decision

## Decision Drivers

1. Canonical user data must survive host loss and generated-state removal.
2. Module governance must be real without recreating three repositories.
3. A Git write must be exact, reviewable, forward-only, and recoverable across
   local and remote partial failure.
4. Existing instruction-only source and named-instance users need a reversible
   migration path.
5. The kernel must be small enough that B2 and B3 reuse it instead of bypassing
   it, but must not pre-design their domain semantics.
6. The first-use flow must explain ownership and failure states to a technical
   user without requiring knowledge of internal manifests.
7. Implementation must reuse proven Go, manifest, no-follow, journal, and
   immutable-artifact patterns and add no runtime service or dependency.

## Competing Approaches

### A. One repository plus a transactional native kernel — selected

Keep canonical source in one remote-backed repository, independently version
the three modules, generate Codex state separately, and route all governed
mutations through a shared plan/journal/Git transaction.

### B. Keep the current two repositories and add an orchestration manifest

This minimizes migration code and preserves existing boundaries. A third local
manifest could link runtime and memory and add Git push to each.

### C. Put canonical source inside the generated named instance

The existing instance already contains runtime, memory, workspace, and copied
dependencies. It could become the Git repository and reduce path count.

### D. Use Git branches, submodules, or nested repositories as module governance

Each module could carry independent history or be referenced as a submodule,
while a superproject binds versions.

### E. Run every mutation directly in the active working tree

Require a clean checkout, write files, validate, commit, and push; use Git to
restore on failure.

## Adversarial Comparison

| Approach | Strongest benefit | Failure under B1 pressure |
|---|---|---|
| A | Matches approved ownership; one backup/remote; reuses current transaction patterns | Requires explicit reconciliation for the unavoidable local/remote atomicity gap; manageable with a durable phase journal |
| B | Least code churn | Directly contradicts the approved one-repository outcome and preserves coordination friction |
| C | Fewest visible roots | Generated removal or projection drift could threaten canonical memory/source; violates source/projection separation |
| D | Maximum module independence | Multiplies synchronization, authentication, migration, and failure modes before B2/B3 prove a need; submodule state is poor first-use UX |
| E | Simple happy path | A crash exposes partial canonical files; rollback risks discarding unrelated work; Git push failure lacks an authenticated semantic-write receipt |

A pure clone-and-swap transaction was also considered. Swapping the complete
repository gives strong local atomicity but scales poorly with growing memory
and complicates open working directories. The selected implementation instead
stages the exact owned changes and their Git tree in an operation area, then
publishes files/ref through a journaled sequence with compare-and-swap checks.
The operation journal makes each visible intermediate state explicit and
recoverable.

## Selected Approach

Use a contract-v1 canonical assistant repository and a contract-v3 named host
binding. The repository owns baseline/module manifests, configuration,
capability source, governed-memory data, migrations, and canonical lifecycle
receipts. The host binding owns only endpoint hash/ref, canonical path identity,
current source commit, generated Codex state, copied executable dependencies,
and the launcher. A separately derivable local operation area exists before
either repository or binding and owns the one active journal.

The Git adapter creates candidate blobs/tree with filters and hooks disabled,
validates the complete candidate through a read-only staged view, writes the
planned filesystem entries with no-follow and compare-before-replace checks,
and advances local `refs/heads/main` with `git update-ref <new> <old>`. It pushes
only `HEAD:refs/heads/main` to the configured `origin`. Before planning, the
adapter fetches that exact ref and requires local HEAD, upstream, host binding,
and the canonical manifest to agree. It never invokes pull, merge, rebase,
reset, checkout of unrelated paths, force, prune, tag publication, or a
provider API.

Fresh create uses `prepared`, `candidate-committed`, `remote-pushed`,
`repository-promoted`, `projection-promoted`, and `verified`, because the empty
remote can safely become the recovery source before the local target exists.
Existing-repository mutation uses `prepared`, `source-index-promoted`,
`ref-committed`, `remote-pushed`, `projection-promoted`, and `verified`.
Recovery validates the same plan, predecessor proofs, candidate commit, remote
ref, active index, and host binding. It completes the unique safe next
transition or stops without mutation. A remote that advanced to an unknown
commit is divergence even when the semantic files appear equal.

Confidence is medium-high. Existing components prove the hardest local
ownership/recovery primitives; new risk is concentrated in Git candidate/ref
construction and cross-boundary recovery, for which the plan requires
adversarial fixtures and real private-remote acceptance.

## Decisions Ledger

| Decision | Rationale | Evidence |
|---|---|---|
| One Git repository with module manifests | Delivers portability while preserving domain governance | Discovery #81 B1 and conceptual architecture |
| Canonical source outside the named instance | Removal and rebuild must never threaten source or memory | Existing source/projection lifecycle and B1 acceptance |
| Required pre-existing empty private remote for create | Provides host-failure recovery without provider-specific account creation | Approved remote-backed outcome; provider setup is non-goal |
| Keep the endpoint URL only in Git config; store its normalized hash in binding/receipts | Recovery needs endpoint identity without unnecessarily copying private coordinates into history/logs | Privacy boundary and Git's native ownership |
| Forward commits for rollback | Preserves audit history and enforces no rewrite/force | Approved Git stewardship decision |
| Preserve legacy repositories after migration | Gives reversible cutover without multi-history surgery or destructive cleanup | Current split contract and migration risk |
| Stable read-only `inspect`, strict `verify`, actionable `diagnose` | Supports healthy return use and degraded recovery without hiding state | B1 critical tasks and terminal experience gate |
| Exact English confirmation only for mutations | Reuses released safety contract; cancellation is any other input | Existing CLI and capability workshop evidence |
| No submodules, LFS, symlinks, special files, hooks, or filters in B1-owned paths | Avoids executable/indirect mutation and ambiguous ownership | Narrow-kernel and no arbitrary dependency boundary |
| Through-production envelope | Repository already has immutable artifact, native acceptance, rollback, and release gates; no service staging applies | Existing release tooling and repository instructions |
