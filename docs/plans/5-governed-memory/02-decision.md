# Solution Decision

## Decision Drivers

1. Durable claims must require visible, deliberate human promotion.
2. Every recalled claim must retain inspectable attribution and provenance.
3. Sensitive content must not escape through argv, logs, implicit consent, or
   retrieval defaults; restricted content must never be recalled.
4. The repository must remain user-owned, local, portable, and reviewable as
   ordinary files without a service or opaque index.
5. A crash or concurrent writer must not create an apparently valid partial
   record or silently destroy evidence.
6. Retrieval must be bounded, deterministic, dependency-light, and testable.
7. Existing contract-v1 repositories must gain the capability without rewriting
   user data or requiring a global install.
8. The implementation should reuse repository validation, transaction, terminal,
   and release patterns while keeping memory semantics in a coherent package.

## Competing Approaches

### A. Markdown files and convention-only folders

Write human-readable observations, proposals, and memories with frontmatter.
This is approachable in editors and Git diffs, but parsing, canonicalization,
cross-reference validation, sensitivity inheritance, and exact transaction
digests become ambiguous. Free-form Markdown makes it easy for malformed
metadata to look valid.

### B. Immutable schema-bound JSON records plus full-scan lexical recall

Store one canonical JSON document per record, authenticate copied schemas
against embedded bytes, validate all cross-references, and rank active durable
memories directly on each recall. This adds explicit structure but keeps the
source of truth inspectable, portable, and dependency-light.

### C. SQLite state store with full-text search

Use transactions and FTS to obtain strong atomicity and efficient retrieval.
This introduces a binary database, schema migrations, journaling modes, backup
semantics, corruption recovery, and a second representation if human-readable
files are also desired. It is disproportionate before scale is evidenced.

### D. Embeddings or a vector service

Semantic retrieval could find related claims without shared words. It requires
a model/service, new privacy and network boundaries, nondeterministic ranking,
credentials or a local model runtime, index invalidation, and cross-platform
support. The approved outcome explicitly excludes it.

## Adversarial Comparison

| Pressure | A: Markdown | B: JSON/full scan | C: SQLite/FTS | D: vectors |
|---|---|---|---|---|
| Deliberate promotion | Convention can be bypassed accidentally | Distinct typed records and command gate | Enforceable transactionally | Often encourages automated ingestion |
| Attribution | Frontmatter is flexible but ambiguous | Required, closed schemas and references | Strong relational model | Metadata can be stored but ranking remains opaque |
| Inspectability | Excellent prose readability | Good with formatted JSON | Requires tooling | Requires index/model tooling |
| Crash safety | Needs the same journal discipline | Single-file atomic promotion plus journal | Native transactions, more operational state | Database/index plus service failure modes |
| Determinism | Parser conventions are a risk | Exact normalization, scoring, and tie-breakers | FTS version/tokenizer can drift | Inherently model-dependent |
| Privacy | Local, but metadata errors can leak | Local with explicit sensitivity enforcement | Local database, harder casual review | New compute/network boundary likely |
| Pilot scale | Adequate | Adequate | Premature | Out of scope |
| Migration burden | Format ambiguity accumulates | Explicit version boundary | Database migrations immediately | Model/index migrations immediately |

Approach B has one known cost: scanning and parsing all durable records on each
recall. The product cap bounds output, not scan work, so implementation must
measure and test a reasonable fixture volume. That cost is manageable for the
pilot and easier to replace later than a prematurely public database contract.

## Selected Approach

Select B. Introduce versioned, canonical JSON records in the existing data
directories and copied authoritative schemas under `.my-friday/schemas/`.
Records are immutable. A proposal cites observations/journals; a durable memory
cites exactly one proposal and copies the reviewed claim, provenance,
uncertainty, attribution, and a sensitivity at least as high as every input.

Every write uses the same small state machine: validate repository and all
current records, acquire the repository memory lock, revalidate, durably write
a body-free transaction intent, create and fsync a same-filesystem staged file,
validate its exact bytes and relationships, atomically rename to its final
absent path, fsync the parent, mark verification, revalidate, and remove the
journal. Recovery uses only stored IDs, paths, modes, sizes, and digests to
finish or remove an owned stage; it never reconstructs content from metadata.

Recall scans only active durable-memory v1 documents. Query and claim tokens are
NFC-normalized, Unicode-lowercased sequences of letters and numbers. Score is
the number of distinct query tokens present in the claim. Zero-score records
are omitted; ties sort by `promoted_at` descending and then ID ascending. After
the per-run sensitivity decision, records are selected in order only when the
complete attributed entry fits, up to five entries and 4,000 extended grapheme
clusters for the entire packet. Entries are never truncated.

Confidence is **Medium-High**: repository and transaction precedents are strong;
retrieval usefulness and workflow friction remain acceptance and pilot-learning
questions rather than correctness unknowns.

## Decisions Ledger

| Decision | Rationale | Evidence |
|---|---|---|
| JSON, one immutable record per file | Closed validation and atomic file promotion without a database | Existing JSON schemas/manifests and local-file ownership |
| Separate proposal and memory IDs | Promotion is an auditable act, not an in-place mutation | Approved deliberate-promotion outcome |
| Explicit `--memory PATH` always | Avoid hidden defaults and sibling discovery | Product experience packet; two-repository privacy boundary |
| Prompt-entered content/query | Avoid shell history and process-list disclosure | Local threat model |
| Exact `Record`, `Propose`, `Promote` | Safe default and visible mutation boundary | Existing `Create` precedent; product experience packet |
| Interactive-only promotion | Prevent scripts from silently creating durable claims | Central governance promise |
| Monotonic sensitivity | Derived data cannot weaken source handling | Product experience packet and privacy boundary |
| Standard by default; sensitive per-run `Include`; restricted never recalled | Useful local default without persistent consent or restricted disclosure | Product experience packet |
| No proposal recall | Draft claims are not durable knowledge | Approved acceptance criterion |
| Unicode lexical overlap, stable ties | Transparent deterministic retrieval without a model | Explicit no-vector scope |
| Five entries and 4,000 graphemes, no truncation | Bounded context without distorting attributed claims | Product experience packet |
| Full scan, no index | Smallest coherent pilot; no invalidation state | YAGNI and local repository size assumption |
| Body-free recovery journal | Recovery authority without duplicating sensitive content | Existing repository transaction precedent |
| No edit/delete/retract in v1 | Avoid lifecycle semantics not needed to prove the loop | Issue #5's smallest outcome |
