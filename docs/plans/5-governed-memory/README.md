# Solution Design: Complete the governed-memory loop

- **Status:** Final
- **Issue:** #5
- **Planning PR:** #20
- **Repository basis:** 359a2dc4c94f050775b93c4dfc366dfe0976d6a4
- **Execution envelope:** through-production

## Decision

Add a local, foreground `my-friday memory` command family for one explicitly
selected contract-v1 memory repository. `observe` and `journal` capture ordinary
activity as attributed JSON records; `propose` turns selected records into a
reviewable claim; interactive-only `promote` creates a distinct immutable
durable-memory record after exact `Promote`; `validate` checks the complete
owned memory contract; `recall` deterministically selects at most five active
durable memories into a 4,000-grapheme, attributed packet that the user may
paste into a fresh Codex task; and `recover` completes or reverses a proven
interrupted write.

All content is entered at terminal prompts rather than command arguments.
Proposals never appear in recall, durable promotion is never automatic, and
the implementation performs no network, Git, global Codex, vector, daemon, or
cross-machine operation. Standard memories may appear in recall; sensitive
memories require exact `Include` on every recall invocation; restricted
memories never appear.

## Needs Attention

- Exact-candidate acceptance must use the immutable Darwin/ARM64 artifact on a
  supported disposable non-admin macOS identity and a generated fixture memory
  repository. It must never read or mutate an operator's live Codex home,
  installed runtime projection, or private memory.
- Issue #5 nomination is blocked until issue #4/O2 is accepted and released and
  its production receipt proves the working exact-byte nomination, acceptance,
  and GitHub Release chain. Issue #5 must reuse that released chain; no
  substitute enabling work or rebuild between acceptance and release is allowed.

## Decision Spotlight

- **A promotion is a new record, not a proposal status flip.** Observations,
  journals, proposals, and durable memories remain immutable and attributable.
  This preserves the evidence a user reviewed and avoids making draft material
  durable by editing it in place.
- **Promotion is interactive-only.** `memory promote` accepts no non-interactive
  confirmation flag and exact `Promote` is the sole mutation signal. Automation
  may capture or propose, but cannot silently grant durable status.
- **Content stays out of argv and transaction metadata.** Claim, narrative,
  rationale, and recall query are prompt-entered. Recovery journals contain
  paths, IDs, phases, modes, and digests, never record bodies.
- **Sensitivity is monotonic.** `standard < sensitive < restricted`; derived
  records inherit at least the highest sensitivity of every cited source.
  Promotion may raise sensitivity but never lower it.
- **Sensitive consent is per run; restricted means excluded.** Recall may
  include standard memories immediately. It discloses only the count of matched
  sensitive records before exact `Include`, remembers no consent, and never
  emits restricted content.
- **Recall is lexical, deterministic, and deliberately small.** Unicode-aware
  token overlap, stable tie-breakers, five-record and 4,000-grapheme caps produce
  an inspectable packet without vectors, embeddings, model calls, or hidden
  ranking state.
- **Recall is manual context transfer.** My Friday prints an attributed packet;
  it does not inject context into Codex, discover a session, or claim that the
  receiving model will use the packet correctly.
- **One memory repository per invocation.** Every command requires `--memory
  PATH`, validates role and assistant identity, and never searches for a sibling
  repository or a global default.
- **Local append-only files are the source of truth.** There is no database,
  index, cache, telemetry, credential store, background worker, or hidden
  synchronization state. Validation recomputes relationships from files.
- **Writes fail closed and recover explicitly.** An owner-only lock serializes
  writers. A durable body-free transaction journal and same-filesystem stage
  precede atomic rename; ambiguous ownership or drift retains evidence and
  refuses mutation.
- **Initialization is its own confirmed prerequisite.** An uninitialized memory
  repository refuses capture and directs the user to read-only preview plus
  exact `Initialize`. That one recoverable transaction installs schemas/control
  state and removes only the four exact generated empty `.gitkeep` placeholders;
  it never hides initialization inside `Record`.
- **Every filesystem capability is pinned and no-follow.** The memory root is
  opened once as an owner-controlled directory descriptor, all traversal and
  mutation is descriptor-relative, and root identity plus entry type/link/
  owner/mode/digest are re-proved immediately before each effect.
- **Completed transactions remain provable.** Final records carry their
  transaction ID and a body-free completion receipt survives journal cleanup,
  making committed and cleanly aborted recovery retries read-only and
  deterministic.
- **Recall explains each match inside the bound.** Every emitted entry includes
  explicit recorder attribution, the fixed lexical reason, and sorted matched
  query tokens; those bytes count toward the 4,000-grapheme cap.

## Plan Map

- [Context](01-context.md)
- [Decision](02-decision.md)
- [Design](03-design.md)
- [Verification](04-verification.md)
- [Implementation handoff](05-handoff.md)

## Final Gate

This plan is final after resolution of the independent findings and validation
on the current PR head. Product authority must approve that exact final head
with the `through-production` envelope before merge.
