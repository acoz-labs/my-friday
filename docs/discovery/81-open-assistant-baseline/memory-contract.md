# Governed Memory Contract Reflection

## Decision sought

Identify which semantics in the mature reference memory system belong in My
Friday's portable governed-memory capability and which are replaceable details
of the current private implementation.

The central finding is encouraging: the reference system already demonstrates
a coherent semantic core. My Friday should standardize its record meanings,
authority transitions, recall packet, maintenance obligations, and conformance
tests. It should not clone its Ruby commands, filesystem layout, local daemon,
source archive, scheduler, or retrieval algorithm.

## Memory constitution

The portable contract begins with six principles:

1. **Capture is cheaper than belief.** Observations, chronology, handoffs, and
   proposals do not become durable guidance merely because they were recorded.
2. **Durable guidance is governed.** Only an authorized review and promotion or
   amendment transition can create active memory.
3. **Current guidance and history are different products.** Default recall
   returns active, currently valid guidance; chronology, superseded memories,
   rejected candidates, and conflict evidence remain explicitly historical.
4. **Provenance travels with the claim.** Durable memory records evidence,
   confidence, sensitivity, validity, volatility, sources, and conflict or
   supersession relationships.
5. **Memory is context, not live-system truth.** Drift-prone guidance must carry
   and satisfy a live-verification requirement before operational action.
6. **Canonical knowledge survives implementation change.** Indexes, summaries,
   embeddings, caches, and harness projections are derived and rebuildable.

## Portable record classes

| Record | Meaning | May guide ordinary recall? |
|---|---|---|
| Observation | Append-only evidence of a user signal, action, result, source read, correction, failure, or success | No |
| Journal entry | Chronological account of completed work, decisions, source consultation, failure, open loop, or maintenance | Historical chronology only |
| Handoff | Quarantined intake from another agent or machine with sender, task, evidence, confidence, sensitivity, and proposed disposition | No |
| Proposal | Candidate durable claim awaiting accepted, rejected, deferred, deduplicated, merged, or conflict disposition | No |
| Memory item | Governed durable fact, preference, decision, procedure, project state, entity knowledge, or research claim | Yes, only while active and valid |
| Source | Citation identity, archive or retrieval reference, consulted time, and verification metadata | Attribution only |
| Reflection | Derived synthesis of a bounded period or corpus slice, including open loops and candidate maintenance work | No, unless separately promoted |
| Transition receipt | Idempotent proof of a governed write, repository reconciliation, migration, conflict resolution, compaction, or recovery | No |
| Verification result | Time-bounded proof that a drift-prone claim was checked against its live authority | Only as a condition on use |

All canonical records need a schema version, stable ID, timestamps, provenance,
sensitivity/visibility policy, and owner. Durable items additionally need type,
scope, entities, confidence, status, validity interval, volatility,
live-verification policy, and conflict/supersession edges. Secret values are
invalid memory content.

## Normative lifecycle

```text
observation ─────────────────────────────────────────── evidence
journal entry ───────────────────────────────────────── chronology
remote input → pending handoff → accepted / rejected / triaged
candidate claim → pending proposal
                 ├── accepted → active durable memory
                 ├── rejected
                 ├── deferred
                 ├── merged/deduplicated
                 └── conflict
active durable memory
                 ├── superseded → historical + active successor
                 ├── archived
                 ├── retracted/redacted
                 └── conflict → resolution → active or historical disposition
```

Promotion is one atomic governed transaction: create or amend the memory item,
record proposal disposition, preserve and merge evidence/source references,
append chronology, validate the resulting corpus, and issue a transition
receipt. Supersession must atomically close the predecessor and create or bind
the active successor; the current reference implementation's manual
predecessor-edit gap must not become part of the portable contract.

Remote or otherwise untrusted material enters through handoff quarantine.
Import and migration follow the same rule: import useful claims with source
identity and disposition, not an old note hierarchy as if it were current
truth.

## Recall contract

The reference system's context packet is a sound starting interface:

- `relevant_memory`: active and currently valid guidance ranked for the task;
- `recent_chronology`: explicitly historical events that illuminate current
  work without masquerading as instructions;
- `pending_or_conflicts`: proposals or records requiring review;
- `unresolved_conflicts`: contradiction evidence kept outside ordinary
  guidance;
- `staleness`: live-verification requirements, age, volatility, and warning
  state;
- `sources`: attribution and verification references; and
- `historical_memory` and `historical_chronology`: returned only when history
  is explicitly requested.

Default recall must never return inactive, expired, superseded, rejected, or
conflicted material as current guidance. An active successor shadows the
predecessor while retaining navigable history. Sensitivity and visibility must
filter by caller and purpose before ranking; this is an identified gap in the
current reference implementation, where sensitivity is metadata but not recall
authorization.

Ranking is an adapter choice. Deterministic lexical search, vector search,
hybrid retrieval, entity indexes, databases, and direct file scans are all
acceptable if they pass the packet and authorization conformance tests.

## Capture, governance, recall, and maintenance ports

The governed-memory capability should advertise semantic operations, not one
daemon or transport:

| Contract | Operations |
|---|---|
| Capture | `observe`, `journal`, `handoff`, `source` |
| Governance | `propose`, `review`, `promote`, `reject`, `defer`, `supersede`, `resolve_conflict`, `retract` |
| Recall | `context`, `history`, `verify_live` |
| Maintenance | `reflect`, `review_queue`, `deduplicate`, `compact`, `migrate`, `validate`, `rebuild_indexes`, `reconcile_remote` |

CLI, MCP, local HTTP, direct library calls, and later ACP-connected clients are
adapters to these operations. No transport gains broader authority than the
caller. The local reference daemon's lack of authentication is implementation
evidence, not a portable precedent.

## Roles and authority

| Role | Permitted transitions |
|---|---|
| Ordinary assistant agent | Recall authorized context; record observations and chronology; submit sources, handoffs, and proposals |
| Memory governor | Review, promote, amend, supersede, resolve conflict, retract/redact, and approve migration or compaction within policy |
| Repository steward | Validate, commit exact touched paths, reconcile and push within configured Git authority; no semantic promotion power |
| User/product owner | Establish policy, grant authority, decide ambiguous sensitivity/conflict/deletion cases, and override or revoke automation |
| Storage/retrieval adapter | Persist and rank only within declared contracts; no authority to promote, widen visibility, or discard canonical history |

Sensitive health, finance, identity, inferred-preference, credentials-adjacent,
and similarly consequential material requires explicit evidence and stricter
review. The schema must express caller, owner, audience/visibility, purpose,
and sensitivity. Recall filters before ranking, and all transitions leave audit
receipts without reproducing secret content.

## Maintenance is part of memory, not an afterthought

The current reference system implements reliable capture, governance, recall,
validation, Git serialization, and daily chronology rollups. Its "reflection"
is presently a deterministic event aggregation, not semantic synthesis. It
also exposes important unfinished work: observation retention is not enforced;
open loops have no closure operation; proposal queues lack dedupe and review
workflow; indexes are not built; schema migrations are absent; and
redaction/right-to-forget behavior is undefined.

My Friday should therefore define memory maintenance as a core contract with
bounded outcomes:

- review and dispose of pending proposals and handoffs;
- identify duplicates, contradictions, stale claims, and orphaned sources;
- close or carry forward open loops explicitly;
- synthesize a proposal from reflection without auto-promoting it;
- enforce retention and compaction dispositions while retaining required audit
  history;
- rebuild all derived indexes from canonical records;
- migrate schemas and storage with preconditions, receipts, rollback, and
  recovery;
- export and selectively import without losing provenance; and
- support governed redaction or retraction without silently rewriting unrelated
  history.

Scheduled execution is optional harness or operating-system composition. The
maintenance operations and their authority are portable.

## Git-backed MVP contract

Git is the selected MVP canonical backend for the one private assistant
repository. Preserve these reference-system safeguards:

- serialize writes and validate the complete affected contract before commit;
- require a clean, unambiguous starting state and stage only exact touched
  paths;
- automatically commit and push a verified change to the configured private
  remote;
- fetch first and refuse divergence, ambiguous conflicts, history rewrite,
  force-push, or repository publication;
- preserve the local commit if a push fails, then reconcile idempotently from a
  transition receipt rather than duplicating the semantic write;
- never treat host-local locking as distributed synchronization; and
- preserve unrelated work and report the exact recovery state.

The repository backend is replaceable later, but these consistency,
auditability, and authority semantics are not.

## Capability-package mapping

The governed-memory capability dogfoods the general package contract:

### Skills

- recall task context;
- capture observation;
- record chronology;
- propose durable memory;
- review and disposition proposals;
- supersede and resolve conflicts;
- submit and triage handoffs;
- cite or register a source;
- reflect, compact, and audit staleness; and
- migrate and import.

### Agent

One memory-governor specialist owns durable promotion and amendment judgment.
Ordinary assistants remain capture/proposal clients. The compiler must preserve
this authority separation even if a target harness models the governor as an
isolated task rather than a native specialist-agent file.

### Hooks

- pre-task context recall;
- observation capture after substantive source/tool reads;
- end-of-task chronology and proposal check;
- live-verification gate before drift-prone action;
- scheduled reflection, proposal review, stale audit, and index rebuild; and
- post-governance repository validation and bounded commit/push.

Hooks that cannot run natively may be emulated by the trusted kernel only when
the same timing and enforcement contract can be proven. A missing required
live-verification or authorization hook is compilation failure, not a warning.

### Services, references, schemas, and tests

Services expose the semantic ports above. References include the memory
constitution, promotion policy, recall packet, sensitivity policy,
live-verification boundary, conflict/supersession rules, retention, migration,
and storage-adapter contract. Schemas cover every canonical record and
transition receipt.

Conformance tests must prove at least:

- no inactive guidance in default recall and history only on request;
- conflicts remain separate from guidance;
- sensitivity-aware caller authorization and secret-content rejection;
- source and provenance preservation;
- atomic promotion, supersession, and conflict resolution;
- duplicate-safe and idempotent retry behavior;
- commit-success/push-failure reconciliation;
- stale/live-verification blocking for required actions;
- remote input quarantine;
- lossless derived-index deletion and rebuild; and
- migration preserving IDs, provenance, status, and history.

## Invariants selected for Gate 1

- The canonical private assistant repository contains `/memory` as an
  independently governed module alongside `/config` and `/capabilities`.
- Observations, journals, handoffs, proposals, durable memories, sources,
  reflections, transition receipts, and verification results have distinct
  meanings and schemas.
- Only governed transitions create or amend durable guidance.
- Default recall is current, authorized, attributed, and validity-aware;
  history and conflict evidence are explicit separate views.
- Sensitivity is enforced during recall and transition authorization, not merely
  recorded as metadata.
- Secret values are rejected from canonical memory.
- Promotion, supersession, conflict resolution, migration, and compaction are
  atomic or receipt-driven recoverable transactions.
- Derived indexes and summaries can be destroyed and rebuilt without losing
  canonical knowledge.
- Reflection may propose but never silently promote durable belief.
- Memory warns or blocks according to declared live-verification policy; the
  live system remains operational truth.
- Git commit/push automation is bounded by validation, private-remote
  configuration, no divergence, and idempotent reconciliation.

## Deferred implementation choices

- Exact JSON field names, storage tree, serialization, and migration engine.
- Initial lexical ranking and the threshold for adding semantic retrieval.
- Retention durations and governed redaction mechanics by sensitivity class.
- Whether recall and write ports run in-process, through MCP, or through a
  local service in the first Codex projection.
- Reflection cadence, review SLA, and compaction thresholds.
- Multi-machine synchronization beyond Git's selected MVP semantics.
