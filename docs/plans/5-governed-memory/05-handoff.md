# Implementation Handoff

## Change Tier And Smallest Complete Outcome

This is a **Tier 1, release-bearing capability**: it creates and retrieves
user-owned potentially sensitive durable data, establishes public schemas, and
adds crash-recovery and consent boundaries.

The smallest complete shipped outcome is one exact released binary with which a
user can initialize a generated memory repository, record observation and
journal entries, create and deliberately promote one attributed proposal,
validate the whole repository, recover an interrupted write, and produce a
deterministic bounded packet for manual use in a fresh task—with sensitive and
restricted behavior and prohibited external effects proven end to end.

Do not ship capture without promotion governance, promotion without validation/
recovery, or recall without sensitivity enforcement and attribution.

## Dependency Order And Reviewable Slices

1. **Typed contract and validator**
   - Likely ownership: `internal/memory`, embedded schema assets, repository
     role validation tests.
   - Begin with failing schema/semantic/canonical byte tests.
   - Exit: all five v1 document types and complete cross-reference invariants
     validate deterministically; no command writes yet.
2. **Governed-contract initialization**
   - Likely ownership: `internal/memory`, CLI adapter, PTY tests.
   - Exit: separate preview/exact `Initialize` removes only the four exact O1
     placeholders inside one WAL, installs schema/control/contract state, and
     refuses changed or foreign paths.
3. **Shared transaction writer and recovery**
   - Likely ownership: `internal/memory` transaction subpackage or cohesive
     files, fault-injection tests, CLI recovery adapter.
   - Exit: pinned no-follow descriptor operations cover every read/write; the
     3×3 stage/final matrix and receipts make every checkpoint and completed
     retry deterministic; exact `Recover` gates effects; foreign state remains.
4. **Observation and journal capture**
   - Likely ownership: terminal memory flow, typed records, command routing.
   - Exit: exact `Record`, safe exits, prompt-only content, normalized preview,
     immutable owner-only files, and sanitized transcripts pass.
5. **Proposal and promotion**
   - Exit: source selection and sensitivity inheritance are exact; `Propose`
     creates draft only; promotion is interactive-only, monotonic, distinct,
     attributable, and idempotent.
6. **Validation and recall**
   - Exit: deterministic ordered diagnostics; lexical golden matrix; proposal/
     restricted exclusion; per-run sensitive consent; whole-entry dual bounds;
     recorder/reason/matched tokens within cap; byte-stable packet.
7. **Boundary evidence and durable docs**
   - Likely ownership: `cmd/my-friday`, `internal/terminal`, evidence harness,
     architecture/user/runbook/deployment docs and release workflows.
   - Exit: full CI/race/fault/concurrency/no-external-effect suite, exact-head
     transcripts, product/security review, reconciliation, docs promotion, and
     this plan's deletion.
8. **Immutable candidate acceptance and release**
   - Exit: independent disposable-user acceptance, accepted exact archive,
     verified Git tag/GitHub Release/receipt, and issue closure.

Prefer one coherent implementation PR because partial merges would expose a
public contract without the complete governance loop. Commits within it should
follow the slices for reviewability.

## Acceptance Traceability

| Acceptance group | Slice | Required evidence |
|---|---:|---|
| Observation and journal | 1, 3, 4 | Schema/semantic/fault tests; exact-confirmation transcripts |
| Proposal and deliberate promotion | 1, 3, 5 | Reference/sensitivity/noninteractive/idempotency tests; transcripts |
| Validation and recovery | 1–3, 6 | No-follow race suite, 3×3 WAL matrix, receipts, exact-Recover transcripts |
| Bounded attributed recall | 6 | Golden lexical/consent/bound tests; manual fresh-task transcript |
| Automatic promotion excluded | 5 | No bypass tests; process/filesystem manifest |
| Vector/network/Git/sync excluded | 6, 7 | Dependency/static review; child/network observer; before/after manifest |
| Exact-candidate release | 8 | Accepted/released issue #4 receipt gate, nomination, independent acceptance, artifact digest, release receipt |

The detailed scenario matrix and candidate identity are in
`04-verification.md`.

## Documentation Promotion

| Durable destination | Shipped knowledge to promote |
|---|---|
| `README.md` | Concise governed-memory quick start, manual recall boundary, sensitivity warning |
| `docs/product.md` | O3 capability and explicit automatic/vector/sync exclusions |
| `docs/architecture.md` | Memory capability in the system map and trust/data boundaries |
| `docs/architecture/governed-memory.md` | Commands, record relationships, schemas, invariants, transactions, recall algorithm |
| `docs/decisions/0002-governed-memory-records.md` | Immutable JSON records/full-scan lexical choice versus Markdown/SQLite/vectors |
| `docs/runbook.md` | Validation diagnostics, transaction recovery, retained evidence, rollback compatibility |
| `docs/development.md` | Fixture rules, injected clock/random/filesystem, transcript and no-external-effect tests |
| `docs/deployment.md` | Exact archive acceptance/release proof and memory smoke test |

Reconciliation must locate actual symbols and behavior, explain every material
drift, promote only the contract that shipped, update the current docs, and
remove `docs/plans/5-governed-memory/` before the implementation PR leaves draft.

## Pull Request And Review Contract

- Branch from the approved merged plan on current `main` as
  `feature/governed-memory`; use `Refs #5` as an exact top-level PR line.
- Keep the PR draft through TDD implementation, product-design transcript
  review, security review, complete CI, reconciliation, docs promotion, and
  temporary-plan deletion.
- Required local checks include formatting, unit/integration/PTY tests,
  `go test -race ./...`, transaction fault matrix, deterministic golden recall,
  no-external-effect probes, and `bin/container bin/ci` (or documented
  repository-approved deviation).
- Independent review must be findings-first and focus on schema compatibility,
  path/symlink/ownership enforcement, recovery deletion authority, concurrency,
  placeholder migration, receipt idempotency, content leakage, recorder fields,
  sensitivity monotonicity, promotion bypasses, match explanations, deterministic
  bounds, and issue #4-gated exact-candidate chain.
- The PR description records issue/plan provenance, acceptance matrix, stable
  error and schema changes, docs matrix, no-rendered-impact classification,
  prohibited effects, release impact, and reconciliation bound to exact head.
- Merge does not close issue #5. It remains open through immutable nomination,
  independent acceptance, GitHub Release verification, and lifecycle closure.
- Nomination is mechanically refused until issue #4's lifecycle and production
  receipt prove O2 independently accepted and released; implementation may not
  substitute another workflow or parallel artifact path.

## Explicit Non-Goals And YAGNI Boundary

- No automatic ingestion/promotion or implicit tool/session capture.
- No automatic injection into Codex, prompt rewriting, hooks, or session APIs.
- No embeddings, vectors, FTS database, semantic model, summarization, or index.
- No provider account, network request, remote, Git operation, daemon, watcher,
  scheduler, telemetry, sync, merge/conflict resolution, or cross-machine state.
- No persistent sensitive consent or sensitivity downgrade.
- No encryption/secrets manager, access-control groups, signatures, or claims of
  protection from a compromised local account.
- No editing, deletion, retraction, supersession, bulk import/export, or legacy
  predecessor migration.
- No generalized event store, repository framework, plugin API, or GUI.

## Exceptions That Reopen Design

Return to product design and Solution Design if implementation requires any of:

- automatic ingestion, promotion, or Codex/session injection;
- semantic retrieval, embeddings, vector/FTS service, or model dependency;
- provider integration, credentials, network access, telemetry, daemon, or
  persistent background process;
- persistent sensitive consent, any sensitivity downgrade, restricted recall,
  encryption claims, or a new audience/authorization boundary;
- cross-machine state, synchronization, merge/conflict semantics, or remote
  memory storage;
- mutable/retractable durable records or migration of private predecessor data;
- inability to guarantee body-free recovery authority, exact-file ownership,
  pinned no-follow access, receipt-linked idempotency, deterministic bounds, or
  the issue #4-authorized immutable candidate chain;
- work outside the approved `through-production` envelope.

Measured full-scan performance or lexical usefulness that makes the pilot
unusable also reopens the selected mechanism rather than silently adding an
index or model.
