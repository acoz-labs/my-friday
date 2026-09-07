# Portable assistant foundation

- **Status:** Draft
- **Related outcomes:** #92, #93, #94, #95, #104, #106, #113
- **Repository basis:** 949a4171145660d3463994d6f339898bcf3d7315
- **Product direction:** Owner decisions recorded in the design conversation on 2026-09-06
- **Delivery authority:** Owner authorized direct local implementation, testing, and review on 2026-09-06; AGENTS.md records the replacement workflow
- **Private evidence:** Structural observations only; examples are synthetic

## Intended outcome

A person creates or imports a named assistant through My Friday's wizard, then
launches it by name from any working directory. The assistant retains its
identity, memory, capabilities, account roles, and learning behavior across
Codex, Pi, sessions, and machines. Each assistant has one private Git repository.
Machines and harness sessions are replaceable installations of that definition.

Thread migration, a hosted service, scheduling, messaging bridges, and a web
dashboard are outside this foundation. The first host target is macOS on Apple
silicon. Portability of records does not imply availability of every host tool.

## Accepted product decisions

1. A wizard offers creation from scratch or import from an existing assistant
   repository. Its default launcher name is the agent name. It configures a
   default harness; `friday --harness pi` overrides it for one invocation.
2. Launch preserves the caller's working directory and real operating-system
   home. Assistant resources resolve through an explicit instance binding,
   independently of the project root and ambient harness configuration.
3. Each installation has separate harness configuration, session storage,
   credentials, and generated indexes. Separate instances provide organization
   and deliberate account selection, not a security boundary against an agent
   with full access to the same operating-system account.
4. Automatic synchronization occurs at defined checkpoints. An unreachable
   remote does not prevent local work or require a confirmation. Local writes
   remain durable and visibly pending until synchronization succeeds.
5. The assistant learns and can create, modify, test, and activate executable
   capabilities and hook subscriptions in its private repository as part of
   authorized work. Routine memory promotion or code activation does not require
   a second human approval. This is an intentional operating tradeoff.
6. Explicit current user direction supersedes older remembered user guidance
   within its scope. A one-task exception does not become a permanent rule.
   Learning from inference remains distinguishable from explicit user direction.
7. Originating machine and every subsequent change's machine are durable
   provenance. Synchronization never changes authorship.
8. Supersession is structured, traceable, and queryable. Historical guidance
   does not return as current guidance merely because it matches a search well.
9. Import preserves historical evidence while reconciling old behavioral rules
   against the new operating contract. Import is not automatic activation of
   legacy instructions, hooks, scripts, or stale host restrictions.
10. Source files in Git are authoritative. Search projections, lexical indexes,
    vectors, and caches are local, replaceable derivatives.

The public toolkit owns shared capabilities such as memory management and
capability design/validation, not an individual user's agent capabilities.
Service integration implementations, credential providers, account-role models,
and personal workflows belong exclusively in private agent repositories—not in
core code, bundled optional packages, acceptance requirements, or seeded examples.
Core extension contracts and their tests remain provider-neutral.

## What changes from existing plans

| Existing direction | Replacement direction |
| --- | --- |
| Separate runtime/memory repositories in the released bootstrap | One versioned assistant repository |
| Fixed neutral workspace and Codex-only projection | Preserve caller directory; Codex and Pi adapters |
| Explicit `Launch stale` approval or refusal while offline | Continue locally with pending synchronization |
| Canonical repository excludes machine roster | Stable device registry and per-change device authorship |
| Fresh-session-only updates | Refresh memory during sessions; distinguish data from loaded instructions/code |
| Instruction-only capabilities and repeated activation confirmations | Versioned executable capabilities with automatic verification and activation |
| Reflection may propose but never automatically promote | Automatic evidence-aware learning and reconciliation |
| Claude Code as the second routing experiment harness | Actual runnable Pi as the second harness |

This draft does not mark existing issues or PRs superseded remotely. In
particular, #92's merged plan and open PRs #110, #114, and #115 remain untouched.
The product owner subsequently authorized replacing the old delivery gates for
this rebuild. AGENTS.md records that explicit local implementation authority.

## Versioned repository format

```text
assistant/
  .my-friday/
    format.json                  # Format version and minimum compatible writer
    migrations/                  # Applied migration records
  agent.json                     # Stable assistant ID, name, default harness
  instructions/
    identity.md                  # User-owned personality and communication
    operating.md                 # User-owned working preferences and direction
  memory/
    records/<record-id>/<revision-id>.json
    events/YYYY/MM/<event-id>.json
    sources/<source-id>.json
  capabilities/<capability-id>/
    capability.json              # Discovery, requirements, subscriptions
    instructions.md
    scripts/
    checks/
  integrations/<integration-id>.json
  provenance/
    devices/<device-id>.json
    changes/YYYY/MM/<change-id>.json
  extensions/<namespace>/         # Declared user/package extensions
```

The format is shared by all users. Integration configuration is user-authored;
no service integration or account model is shipped by the toolkit. The manifest and record
fields are validated by versioned schemas embedded in My Friday. Scaffold
templates are shipped with the toolkit and may have declared user overrides;
updating the toolkit must not replace customized instructions with new template
defaults. User-editable repository files cannot redefine the core validator.

Host bindings live outside this repository. They contain local paths, executable
locations, device registration binding, harness homes, credential bootstrap
locations, active capability generations, synchronization state, and index state.
Portable integrations contain logical account roles and credential references,
never resolved secrets. No machine-specific absolute path is a portable default.

Version the toolkit, repository format, event protocol, and capability packages
independently. An older writer detects an unsupported newer format before
mutation. A migration is a tested from/to transformation with an affected-path
set, provenance, and a recovery record. Migration preserves record IDs and
authorship. Rollback uses a new corrective commit or a compatible saved
generation, never a force-push or silent loss of post-migration user data.

## Memory, authorship, and supersession

`memory-revision.schema.json` and `examples/` make the first record contract
concrete. They are design fixtures, not an implemented repository validator.
The examples show two revisions, not a complete importable corpus; their source
and device IDs illustrate references whose records will be supplied by the
implementation fixtures.

A record identifies one scoped claim or procedure decision. Its revisions are
immutable. The first revision has no predecessors; a replacement lists the
revision IDs it supersedes and an explicit change reason. Every revision records
its own originating device, actor, harness, model when known, local session
reference when available, evidence basis, recording time, and effective time.
The record's original author is available from its root revision; latest authors
are available from its current head or unresolved heads. Synchronizing a revision
does not append another author or rewrite timestamps.

Sensitivity, volatility, and last verification remain separate from relevance
and confidence. A remembered live-system fact can require current verification
even when it is the applicable revision and its original source was reliable.

Current/superseded/conflicted are derived graph states, not mutable flags copied
into several documents. A scheduled successor takes effect at `effective_from`;
its predecessor remains applicable until then. V1 does not support arbitrary
backdated corrections; recording historical imports and their effective dates
uses an explicit import path. Temporary task exceptions use task scope and are
not successors to the global record.

Semantic validation, in addition to JSON schema validation, must enforce:

- unique revision IDs; referenced predecessors and sources exist;
- no cycles or self-reference;
- successors preserve record identity, kind, and scope;
- one initial root per record, except a deliberately recorded import repair;
- immutable revision bytes after publication;
- device IDs resolve to provenance, or have an explicit unknown legacy origin;
- effective-time consistency between successor and predecessor;
- conflicting concurrent heads are returned as conflicts, not ranked as truth.

Independent additions on two machines can merge mechanically. Two revisions
superseding the same head remain competing heads until a reconciliation revision
names both predecessors and explains the resolution. Timestamp order, Git merge
order, and vector similarity do not decide authority. An explicit user correction
can resolve the conflict automatically when its intended scope is clear.

Device IDs are generated at enrollment rather than derived from serial numbers
or hostnames. Friendly labels are editable. Import on another machine enrolls a
new device; restoring a copied installation must detect/reconcile its binding
before claiming the old device identity. Provenance is operational attribution,
not cryptographic proof against an unrestricted local writer.

Capability and configuration changes use provenance change records containing
the device, actor, reason, source references, and affected paths/content hashes.
Committed records may reference a parent commit, but cannot embed their own
resulting commit SHA because that creates a circular hash dependency. Local
operation receipts can associate a completed change with its resulting commit.

## Lifecycle subscriptions

Register one My Friday adapter per harness installation. The adapter registers
the native events it supports and dispatches into the shared event protocol.
Capabilities subscribe in their manifests; machine-local overrides can register
local scripts without putting their absolute paths in portable definitions.

Common candidate events are `session.started`, `request.received`,
`tool.before`, `tool.after`, `request.completed`, `context.compacting`,
`context.compacted`, `session.ending`, `memory.changed`, and `repository.synced`.
`task.completed` is a separate semantic event emitted when task completion is
recorded; a harness stop does not prove completion. Preserve raw native names
under namespaced events such as `native.pi.turn_end` when semantics differ.

The versioned event envelope includes event ID, causation ID, assistant ID,
device ID, session/request IDs when available, native event name, working
directory, source generation, and an event-specific payload. Data reaches a
script through structured stdin, not shell interpolation. The subscription
contract declares handler path/arguments, timeout, ordering dependencies,
synchronous/background execution, required semantics, and failure behavior.

The dispatcher validates subscriptions, detects dependency cycles, keeps
foundation ordering explicit, and records bounded outcome metadata. No-op events
need not create processes or durable logs. High-frequency events are not a reason
to run Git or an LLM repeatedly. Capability scripts execute under the user's
chosen full-access posture; My Friday does not claim to sandbox them.

Deduplicate replayed event IDs and detect causal recursion. Pin a complete
capability generation while a handler runs. Update executable generations between
operations so a pull does not replace half of a running capability. Unsupported
event semantics are reported explicitly. New native events require an adapter
version update; there is no claim to intercept events a harness does not expose.

Normal shutdown is only an extra checkpoint. Local writes and a recoverable
outbox must survive a process being killed without a final hook. Local Git
rollback cannot undo an email sent, a secret created, or another external effect;
capabilities record those outcomes and avoid blindly replaying them.

## Synchronization and retrieval freshness

The launcher reconciles before harness startup. Lifecycle events then trigger
the same engine at request and durable-write boundaries. All Git operations
target the explicit assistant repository, independently of the project cwd and
inherited Git selection variables. Local cooperating writers use transactions
and a per-instance lock. Never hold a filesystem lock while waiting on a model.

Remote status distinguishes last successful observation, pending commits,
offline/authentication failure, and semantic conflicts. A successful fetch is
an observation at a time, not a promise of perpetual global freshness. Network
failures preserve the outbox and retry with bounded backoff. Non-fast-forward
pushes trigger fetch/reconciliation; ambiguous push completion is checked against
the remote before retry. Never force-push assistant history.

A new direct read sees the current synchronized files. Search indexes must
catch up too: bind them to the source revision and record content hashes. Refresh
the changed lexical/metadata subset synchronously before promising fresh recall;
embeddings can catch up in the background. Newly changed records remain visible
through lexical/direct lookup while embeddings are pending. Remove obsolete
revisions from current search even if the vector index still contains them.

Already-delivered model context is not changed by Git. A relevant correction
packet identifies the old revision as superseded. Startup instruction changes
and loaded extensions use explicit adapter reload or a new session when
required; ordinary memory reads never require a restart solely because files
changed.

## Retrieval contract and experiment

One CLI/API contract provides recall, record reads, history, and capability
discovery. A hook prepares a bounded initial context packet from request intent;
the agent can retrieve more as its task develops. Essential identity, user
authority, and memory-use obligations remain present independently of search.

Recall packets identify source revision, current claims, applicable capability
IDs/revisions, open commitments, provenance, sources, and separate conflicts.
History is explicit. Selection honors scope and supersession before ranking,
then resolves returned IDs against authoritative current records before content
is used. Machine, entity, kind, time, and scope are structured filters.

Compare structured lookup plus full-text retrieval with QMD hybrid retrieval.
QMD is a candidate backend, not a mandatory repository format or source of truth.
A QMD adapter may generate local Markdown projections from JSON revisions,
preserving record/revision mappings. All projections, databases, vectors, and
model caches are rebuildable and excluded from the assistant Git repository.

Use a predeclared synthetic corpus and held-out prompts covering exact account
names, paraphrases, conflicting scopes, current/history queries, no answer,
recent corrections, and capability discovery. Measure correctness, obsolete
guidance leakage, recall of fresh records, latency, context size, installation
cost, and behavior while embeddings are unavailable. Do not publish private
memory or assume a model-based search backend is better without the comparison.

## Private capability boundary

Core supplies capability manifests, instructions, execution, lifecycle
subscriptions, and verification tools. Users author their own integrations,
credential operations, account selection, and workflow rules in their agent
repository. A default installation must not imply a chosen service, provider,
number of identities, or domain-specific workflow.

Core synchronization needs authentication extensibility, not a password-manager
client or hosting-service policy. A private executable can implement Git's
credential-helper protocol; the helper owns provider selection and identity
verification. The toolkit owns literal argument forwarding, explicit repository
selection, bounded execution, and durable offline behavior. Provider acceptance
tests stay with the private capability, not the public core.

API/CLI, browser, and computer-use adapters are all candidate execution routes.
Availability is measured on each installation. A failure in one route should
lead to investigation of another appropriate route, rather than a remembered
claim that the task is impossible. Account selection remains consistent across
routes, including browser profiles.

## Migration without inherited behavioral restrictions

Import first into an isolated candidate repository and preserve the original
source. Record source commit and record IDs, original provenance when known,
and import device/time separately. Missing original authorship remains unknown.

Classify facts, preferences, decisions, open commitments, chronology,
procedures, and executable content. Preserve the evidence while reconciling
current applicability. Old machine assumptions, obsolete identities, refusal
rules, and delivery gates are not automatically promoted into new instructions.
Preserve still-valid user preferences and factual context. The assistant may
perform routine reconciliation automatically under this approved direction;
only unresolved material ambiguity needs the owner.

Validate a representative sample and recall scenarios before a full import.
Port each capability through its new manifest, dependency, account, and check
contracts. Switch the launcher only after the new installation passes its
acceptance scenarios. Keep the existing installation and repository available
for recovery; do not operate both against the same effectful task by default.

## Acceptance scenarios and implementation order

| Slice | Required evidence |
| --- | --- |
| Format and memory core | Valid create/read/revise/history; machine provenance; immutable revisions; conflict and scope cases; upgrade preservation |
| Wizard and launch | Create/import distinct instances; name/default harness; override; preserve cwd; both harnesses recall outside the assistant repo |
| Sync and event core | Two clones plus concurrent sessions; offline writes; divergent updates; killed process; duplicate hook; no task-project Git mutation |
| Retrieval | Shared contract; fresh correction visibility; current/history separation; measured full-text versus QMD comparison |
| Private extension boundary | No seeded service capability; generic helper protocol; literal arguments; explicit configuration; no ambient Git helper fallback |
| Migration | Representative import preserves evidence without activating obsolete rules; complete cross-machine learning loop |

Each slice requires a bounded implementation plan and meaningful tests. Use
synthetic fixtures and disposable repositories for development. Live
integration verification needs configured agent-owned credentials; successful
unit tests alone do not establish working authentication or computer use.

The first implementation slice is the format/memory core: schema-backed
creation, revision, supersession, provenance, history, and migrations. Hook,
QMD, and authentication experiments can then consume the same contracts rather
than inventing their own memory representation. No runtime behavior or schema
compatibility is claimed by this design-only change.

## Sources

- [Existing kernel plan](../../plans/92-canonical-assistant-kernel/03-design.md)
- [Existing capability discovery experiment](../104-on-demand-capabilities/README.md)
- [Codex instruction loading](https://learn.chatgpt.com/docs/agent-configuration/agents-md)
- [Codex lifecycle hooks](https://learn.chatgpt.com/docs/hooks)
- [Pi configuration and launch](https://github.com/badlogic/pi-mono/tree/main/packages/coding-agent)
- [Pi lifecycle extensions](https://github.com/badlogic/pi-mono/blob/main/packages/coding-agent/docs/extensions.md)
- [QMD retrieval](https://github.com/tobi/qmd)

## Execution scope

The product choices are resolved and the owner authorized direct development,
testing, and review. On 2026-09-07 the owner clarified that the configured
development account may be used for commits and feature-branch pushes. Private
assistant account separation and SDLC policies do not govern this repository's
development. Live service capabilities are configured and verified separately
in private agent repositories. Production releases and live migrations remain
outside this development scope.
