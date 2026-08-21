# Context

## Problem And Desired Outcome

Issue #5 selects discovery outcome O3: a user must be able to capture ordinary
activity, deliberately promote one durable claim, and recall relevant,
attributed context in a fresh task. The value is continuity the user can inspect
and govern—not autonomous learning. The minimum loop includes observation,
journal, proposal, promotion, validation, and bounded recall while excluding
automatic durable promotion, vectors, and cross-machine synchronization.

The product-experience review further establishes a terminal-first flow:
`memory observe|journal|propose|promote|validate|recall|recover`, explicit
`--memory PATH`, prompt-entered content, exact `Record`/`Propose`/`Promote`,
interactive-only promotion, per-run `Include` for sensitive recall, and a manual
packet for a fresh Codex task.

## Current State

This plan is based on repository commit
`359a2dc4c94f050775b93c4dfc366dfe0976d6a4` and the approved O3 discovery head
`b6db62bf15c8d6ad7a15f7533e6aa5981ae1cd8a`.

- `internal/plan/plan.go` creates a contract-v1 memory repository with
  `data/observations`, `data/journals`, `data/proposals`, and `data/memories`,
  but only `.gitkeep` placeholders. `schemas/README.md` explicitly reserves
  versioned governed-memory schemas for a future outcome.
- `internal/repository/repository.go` authenticates embedded repository-manifest
  schema bytes, role `memory`, assistant ID, and local-Git status. Its owned
  `.my-friday` allowlist must evolve before memory schemas or transaction state
  can be valid.
- `cmd/my-friday/main.go` exposes `init`, pair-level `validate`, `recover`, and
  `version`; it has no memory command family or memory-specific stable errors.
- `internal/transaction` and `docs/architecture/repository-bootstrap.md` provide
  precedent for plan-bound staging, body-free journals, exact-tree ownership,
  atomic promotion, conservative recovery, and owner-only modes.
- `internal/terminal` provides line-oriented prompt and transcript-test
  precedent. The shipped initialization confirmation is case-sensitive and
  safe by default.
- `README.md` declares macOS 14+, Apple silicon, local APFS, an interactive
  UTF-8 terminal, and Git 2.28+ as the supported pilot environment.
- `bin/ci`, artifact nomination, acceptance, production receipt, and release
  tooling define the repository release path. Issue #4's approved plan extends
  that path to one immutable Darwin/ARM64 archive.

There is no existing governed record format, retrieval index, migration burden,
or user data to backfill. Existing generated memory repositories are contract
v1 and contain only reserved empty directories unless users independently
added foreign content.

## Actors And Critical Journeys

### Technical user

1. **Capture observation:** choose a memory repository, enter source,
   uncertainty, sensitivity, and observation text; review the complete preview;
   exact `Record` writes one record.
2. **Capture chronology:** enter when an event occurred, actor, sensitivity,
   uncertainty, and narrative; exact `Record` writes one journal record.
3. **Propose:** enter one claim, rationale, and one or more existing observation
   or journal IDs. The command derives a sensitivity floor, previews attribution,
   and exact `Propose` writes a pending proposal.
4. **Promote:** select a pending proposal by ID, review claim, evidence,
   uncertainty, sensitivity, asserter, and promoter, optionally raise
   sensitivity, then type exact `Promote`. A new immutable durable-memory record
   is written; the proposal remains draft evidence and is never recalled.
5. **Recall in a fresh task:** enter a query, decide per-run whether matched
   sensitive memories may be included, receive a bounded attributed packet,
   and manually paste it into a new Codex task.
6. **Diagnose:** run validation to identify malformed, mismatched, orphaned,
   downgraded, duplicate, or incomplete state without changing it.
7. **Recover:** pass the exact reported transaction journal to a separate
   recovery command. Review the body-free intended action and type exact
   `Recover`; every other input exits without mutation. Recovery proceeds only
   when ownership and expected bytes can be proved; otherwise it preserves
   evidence.

### Independent acceptor

Runs the hands-on loop from the immutable candidate under a disposable
non-admin identity, inspects sanitized transcripts and files, opens a fresh
Codex task manually, verifies attribution and exclusions, and records acceptance
against the exact candidate. The implementer cannot be the sole acceptor.

## Acceptance And Non-Goals

### In scope

- Version-1 JSON Schemas for the governed-memory contract, observation, journal,
  proposal, durable memory, write transaction, and completion receipt.
- Local immutable record creation with explicit confirmations and provenance.
- Monotonic standard/sensitive/restricted handling.
- Full repository validation and exact, conservative recovery.
- Deterministic lexical recall of at most five active durable records and at
  most 4,000 extended grapheme clusters.
- Sanitized no-rendered-impact transcript evidence and exact-artifact release.

### Explicitly out of scope

- Automatic promotion, automatic ingestion, automatic Codex/session injection,
  background capture, and filesystem watchers.
- Semantic retrieval, embeddings, vector databases, model-generated ranking,
  summarization, or claim synthesis.
- Provider integration, remote storage, network access, Git commit/push, merge,
  synchronization, conflict resolution, or cross-machine state.
- Persistent consent, sensitivity downgrade, hidden redaction, encryption,
  secrets management, retention automation, editing, deletion, retraction,
  supersession, bulk import/export, or a GUI.
- Treating prompt text or memory content as instructions with authority.

## Constraints, Dependencies, And Risks

- O1 supplies a recognized memory repository. O2/issue #4 supplies a safely
  managed Codex baseline and the exact-artifact release chain; memory commands
  remain repository-local and do not require the baseline to be installed.
  Nevertheless, issue #5 nomination is authority-gated on issue #4 being
  independently accepted and released with a verified production receipt. The
  issue #4 chain is reused, not reimplemented or replaced in issue #5.
- The on-disk contract becomes durable public API. Schemas use
  `additionalProperties: false`, embedded authoritative bytes, semantic checks,
  explicit versions, and no silently accepted future fields.
- Record bodies can contain personal or secret material. Owner-only directory
  and file modes reduce local exposure but are not encryption. Preview and
  transcripts must not echo rejected or restricted bodies into durable CI logs.
- Command arguments are visible in shell history/process listings, so user
  content and recall queries are read from the terminal only. IDs, paths, and
  enum selections may be arguments or prompts.
- APFS atomic rename protects a single file, not an unjournaled multi-file
  workflow. A durable write transaction is required before staging each record.
- Existing O1 repositories contain empty `data/*/.gitkeep` placeholders. Their
  exact empty regular-file bytes/mode are known generator output, but they may
  now be tracked by Git. Initialization must preview their removal, remove only
  exact generated placeholders within its WAL, and refuse changed or foreign
  entries; My Friday never runs Git or conceals the resulting working-tree diff.
- Path prechecks alone do not contain ancestor replacement races. Every read,
  validation, write, rename, unlink, and recovery action must use a pinned
  memory-root descriptor and descriptor-relative no-follow operations, with
  identity revalidation at mutation boundaries.
- Deterministic retrieval can miss synonyms and contextual similarity. That is
  an accepted transparency tradeoff, not a defect to conceal with fuzzy logic.
- Manual paste can expose selected content to whatever model/service handles
  the fresh task. My Friday must state this at the recall decision point and
  cannot claim local-only processing after the user pastes the packet.
- Record counts may grow. V1 deliberately scans schema-valid local files at
  recall time. An index is deferred until measured scale warrants migration and
  invalidation design.

## Evidence, Assumptions, And Unknowns

### Evidence

- Approved discovery O3 requires the complete named loop, deliberate promotion,
  attribution, and explicit exclusions.
- The repository already generates separate owner-only memory repositories and
  the four record directories (`internal/plan/plan.go`).
- Existing repository transactions prove the project's conservative ownership
  posture (`internal/transaction`, `docs/architecture/repository-bootstrap.md`).
- The product experience packet establishes exact commands, confirmation words,
  sensitivity interaction, recall bounds, and manual fresh-task transfer.

### Assumptions

- Early technical users can copy record IDs and paste a clearly delimited packet.
- A deterministic full scan is adequate for the pilot cohort and bounded local
  repositories.
- User-entered actor labels plus immutable timestamps and source IDs provide
  sufficient first-version attribution without account identity or signatures.
- Exact confirmation words are consistent with the existing initialization
  experience and reduce accidental writes.

### Unknowns retained as measurement, not blockers

- Whether users find proposal/promotion friction proportionate to trust.
- The typical record volume before full-scan latency becomes noticeable.
- Whether lexical matching provides enough usefulness without semantic search.
- Whether users consistently understand that pasting a packet may transmit it.

None changes the approved outcome or prevents a safe first implementation.
