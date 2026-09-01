# Discovery: Define The Open Assistant Baseline

- **Status:** Final
- **Discovery issue:** #81
- **Discovery PR:** #82
- **Repository basis:** 08b4eb4cd30df4f05bf15ea6c362acac1d48d814
- **Recommended decision:** approve
- **Gate 1:** awaiting-authority
- **Confidence:** Medium
- **Private evidence:** none

## Decision sought

Decide whether My Friday should evolve from a Codex bootstrap/lifecycle toolkit
into the open, user-owned baseline from which a personal assistant is composed,
installed, extended, remembered, repaired, migrated, and removed.

The selected boundary is deliberately narrower than an execution engine: Codex
is the first **agent harness**, while My Friday owns the assistant's canonical
private repository, bootstrap kernel, portable memory and capability contracts,
harness compilation, verification, and reversal. This preserves the successful
Codex/Apple-silicon starting point without making Codex itself the product
definition or prematurely implementing a second harness.

Two focused reflections make the recommendation implementation-shapable:

- [capability-package.md](capability-package.md) defines a capability as a
  harness-independent semantic package of typed components and specifies the
  compiler/fidelity boundary; and
- [memory-contract.md](memory-contract.md) separates the portable governed
  memory constitution, records, transitions, recall packet, maintenance, and
  tests from the mature reference implementation's replaceable machinery.

## Product thesis

A durable personal assistant begins with two user-owned systems:

1. **Memory** preserves relevant context across tasks with provenance,
   uncertainty, deliberate promotion, and bounded recall.
2. **Capabilities** turn intent into inspectable, testable, permissioned,
   lifecycle-managed abilities.

My Friday should instantiate one private canonical Git repository with
independently governed `/config`, `/memory`, and `/capabilities` modules. Its
bootstrap kernel sets up and reconciles that repository, stewards its Git
history, maintains governed memory, authors capabilities, compiles them for a
selected harness, and launches the assistant. The memory maintainer and
capability author then dogfood the same public capability contract users extend.

Identity, instructions, modes, agents, tools, connectors, transports, and
automation can compose on that foundation; My Friday provides the means to
author them rather than shipping a maintainer's personal integrations.

The product succeeds when a technically capable user can bring up an assistant
on a clean supported machine, understand exactly what owns their data and
execution, use memory across fresh tasks, build and operate another capability,
and safely update or remove the whole baseline.

## Audience and critical tasks

### Initial audience

- Developers, self-hosters, and technical knowledge workers who prefer
  ownership, inspectability, and repairability over an opaque hosted assistant.
- Capability authors who need a stable path from intent through source,
  validation, activation, migration, verification, and reversal.
- Maintainers who must ship official core capabilities through the same
  contracts offered to users.
- Operators of mature private assistants who need reversible dogfood and
  migration rather than a big-bang replacement.

### Critical tasks

1. Create a named, user-owned assistant as one private canonical Git repository
   with independently governed config, memory, and capability modules.
2. Preview and apply an exact lifecycle plan without disturbing unrelated
   harness, credential, workspace, or user state.
3. Capture observations and chronology, stage candidate beliefs, deliberately
   promote durable memory with provenance and uncertainty, and recall relevant
   context in a fresh task.
4. Define, inspect, validate, test, install, enhance, disable, enable, upgrade,
   recover, and remove a typed multi-component capability without silent
   activation or harness-specific canonical source.
5. Know which files, data, permissions, dependencies, projections, and effects
   each capability owns before activation.
6. Verify, diagnose, repair, upgrade, roll back, migrate, and remove the
   baseline and its official core capability set.
7. Compile the same canonical capability for a supported harness and inspect a
   deterministic fidelity report before activation.
8. Automatically commit and push verified assistant-repository changes to its
   configured private remote while refusing divergence, ambiguous conflicts,
   history rewrite, force-push, or publication.
9. Compose private identity and policy without putting private deployment state
   into the public product or requiring a source fork.

## Evidence

### What My Friday already proves

- It is a public MIT-licensed native command with a supported Apple-silicon,
  macOS, APFS, Git, terminal, and Codex pilot boundary.
- It previews and transactionally creates separate user-owned runtime and
  memory repositories.
- It manages narrow Codex projections and isolated named assistant instances
  with explicit ownership, verification, recovery, rollback, and reversal.
- It has a strict source/projection split and lifecycle for an
  `instruction-only` capability profile.
- It has a deterministic workshop that creates and enhances canonical
  capability source without granting activation authority.
- Discovery #49 and #72 established capability-building before governed memory
  and recovered that direction when natural agent authoring could not be
  honestly accepted.

### What the mature reference deployment proves

A mature private assistant needs more than a prompt and a folder of skills. Its
current baseline demonstrates:

- immutable, content-addressed runtime releases with deployment verification
  and rollback;
- composed global instructions, task modes, specialist agents, reusable
  skills, deterministic commands, and shared libraries;
- routing between judgment-bearing agents and repeatable procedural skills;
- governed long-term memory with observations, journals, handoffs, proposals,
  durable promotion, recall, reflection, provenance, conflicts, and stale-state
  warnings;
- explicit tool, adapter, connector, credential-injection, and secret-handling
  boundaries;
- optional transports, scheduling, services, and software-delivery machinery;
  and
- a private operational overlay that must never become a public product
  dependency.

This is target-shape evidence, not an MVP backlog. Copying the complete private
deployment would produce an unbounded, machine-specific platform and obscure
the two principles the public baseline is meant to establish.

### Current delivery state

- The released public artifact predates the capability workshop.
- Issue #74 and its implementation are close to release but still require
  fresh review, immutable-candidate acceptance, and release.
- Issue #52, governed memory as the first data-bearing core capability, is
  waiting for its Solution Design.
- Issue #75, a conversational adapter for workshop answers, is parked.
- Issue #7, generic remote attachment, is independently shaped but is not part
  of the new foundational sequence.

## Current-state comparison

| Concern | My Friday today | Mature reference deployment | Product implication |
|---|---|---|---|
| Product boundary | Codex bootstrap and lifecycle toolkit | Complete private assistant runtime | Re-charter My Friday as the public distribution/lifecycle foundation, not the whole private deployment |
| Execution | Launches a managed Codex instance | Codex supplies model/tool loop | Call Codex the first agent harness; do not claim My Friday owns execution |
| Identity | Profile rendered into managed instructions | One durable assistant identity with task modes | Keep identity user-owned and overlays composable; modes are optional composition |
| Release | Native artifact plus manifest-owned instance state | Immutable content-addressed runtime releases | Make the complete baseline and official core capability set versioned, verifiable, and reversible |
| Canonical repository | Separate runtime and memory repositories | Several coordinated private repositories and projections | Use one private assistant repository for the MVP with independently governed `/config`, `/memory`, and `/capabilities`; permit later extraction or external references |
| Capability | Strict instruction-only source with Codex compatibility embedded | Skills plus deterministic commands, tools, services, hooks, and expert agents | Preserve the strict source/projection lifecycle but adopt a harness-independent typed package and compiler fidelity contract |
| Capability authoring | Deterministic terminal workshop | Maintained skill/agent authoring procedures | Ship conversational package authoring while deterministic kernel validation, Git, compilation, and activation retain authority |
| Memory | Separate repository placeholder and approved future outcome | Governed capture, proposal, promotion, recall, chronology, handoff, provenance, conflict, and staleness system | Ship the portable capture/governance/recall/maintenance contract without cloning its storage or transport |
| Adapters | Codex-specific files and launch behavior | Many private service/tool boundaries | Keep one real Codex compiler; every projection receives a deterministic receipt and fidelity report; wait for a second harness before extracting shared implementation machinery |
| Git | Lifecycle repositories with bounded transactions | Validated commits and guarded pushes | Auto commit/push verified assistant changes to the configured private remote; never force, rewrite, publish, or resolve ambiguous divergence |
| Secrets | Intentionally out of the first capability profile | Narrow reference-based secret injection | Keep secrets outside the first baseline; later integration profiles require explicit secret contracts |
| Transports and scheduling | Out of scope | Useful optional operations | Maintain later as optional capabilities or capability sets, never mandatory core |
| Software delivery | Product repository SDLC | Private orchestration and source-control machinery | Use it to build My Friday; do not bundle it into every user's assistant |

## Proposed vocabulary

The vocabulary is part of the product contract. Each term should have one job.

### Agent harness

The external model-and-tool execution environment that runs an assistant.
Codex is the first supported harness. My Friday does not own Codex's internal
model loop, tool protocol, session store, or release cadence.

### Assistant

A user-owned identity and policy composed with one baseline, selected
capabilities, governed memory, and one active harness adapter. The assistant is
not synonymous with its harness, model, repository, or launcher.

### Baseline

The minimum versioned My Friday distribution that can create, reconcile,
verify, diagnose, upgrade, roll back, and remove an assistant. The baseline
declares its contract version, bootstrap kernel, official core capability set,
compatible harness compiler, and migrations.

### Bootstrap kernel

The small trusted control plane that can establish and reconcile the assistant
repository, validate and steward exact Git changes, maintain governed-memory
transactions, author and validate capability source, compile and verify harness
projections, and launch the selected harness. Kernel authority is deterministic
and non-removable from a valid baseline; agent-authored content cannot grant
itself kernel authority.

### Capability

A versioned, self-contained, harness-independent semantic package for one
coherent ability. It declares purpose, typed components, interfaces, effects,
permissions, dependencies, configuration, secret slots, data ownership,
tests, migrations, and lifecycle. A capability may contain instructions,
skills, agents, hooks, services, commands, references, assets, schemas, and
tests; see the dedicated package reflection.

### Core capability set

The digest-bound set of first-party capability packages maintained and tested
with a baseline. Its first dogfood packages expose governed-memory maintenance
and capability authoring to the assistant. Setup/reconciliation, repository
stewardship, compilation, lifecycle authority, and launching remain kernel
mechanics even when an agent-facing capability invokes them.

### Capability profile

A bounded trust, execution, dependency, permission, and data class. The
existing `instruction-only` profile remains valid. `data-bearing` is the next
profile required for governed memory. Executable, integration, background, or
credential-using profiles require later explicit product decisions.

### Capability set

A resolved, digest-bound collection of compatible capability versions.
`core` is the official mandatory set. A single capability is already a bundle
of typed components, so `bundle` is not used for a collection of capabilities.
A public registry or marketplace is not implied.

### Adapter

A maintained compiler and verifier between My Friday's owned contracts and a
specific harness, operating system, or provider boundary. It emits a native
projection, receipt, and fidelity report classifying each required semantic as
native, translated, emulated, optional-and-omitted, or unsupported-and-refused.
It does not redefine assistant identity, memory semantics, or authority.

### Overlay

User- or operator-owned identity, policy, configuration, and selected capabilities
composed with a public baseline without modifying the baseline source. Private
deployments remain overlays, not forks or hidden core requirements.

### Projection

Adapter-produced installed state derived from canonical capability source plus
assistant configuration and granted authority. Projections are replaceable and
verifiable; they are never the only copy of source, configuration, secret
references, or durable memory.

## Decision

### 1. My Friday is the distribution and lifecycle layer

Select a layer above agent harnesses. My Friday owns portable assistant
contracts, official baseline composition, capability and memory lifecycle,
adapters, conformance, migration, verification, and reversal. It does not own
the model/tool execution loop in the current product direction.

This gives My Friday an enduring job even as Codex evolves, while avoiding the
cost and risk of building a model runtime before the baseline itself is proven.

### 2. Ship only the bootstrap kernel and its two dogfood capabilities

The bootstrap kernel owns setup/reconciliation, repository stewardship,
governed-memory transactions, capability authoring and validation, harness
compilation, and the assistant launcher. Its first agent-facing core capability
set contains governed-memory maintenance and capability authoring. Together
they embody the two product principles without importing personal integrations.

Modes, specialist agents, secret injection, connectors, transports,
scheduling, software-delivery orchestration, and background services are
component types users may author into their own capabilities. My Friday ships
the means and contract, not the maintainer's private implementations.

### 3. Preserve Codex and Apple Silicon as the first supported reality

Ship one excellent Codex compiler on Apple Silicon. Design domain contracts so
Codex details do not become canonical assistant semantics. Define the compiler
and fidelity interface now because it protects the source contract; extract
shared adapter implementation only after a second real harness proves it.

### 4. Keep the MVP for technically capable users

The first public proof is ownership and lifecycle correctness, not zero-setup
consumer onboarding. A broader audience requires a later experience decision,
distribution work, and evidence that the technical product is retained.

### 5. Migrate the private reference deployment in stages

“Runs the reference assistant” should mean the public baseline can eventually
host its portable foundation while private policy and operations remain an
overlay. Compatibility proceeds through inventory, contract mapping, export,
shadow operation, parity evidence, and reversible cutover. Full operational
parity is not a prerequisite for the public MVP.

### 6. Use one private canonical assistant repository for the MVP

Create `/config`, `/memory`, and `/capabilities` as independently governed
modules inside one private Git repository. Each module retains its own schema,
authority, migrations, retention, and removal contract. Capabilities may later
be extracted to or referenced from separate repositories without changing the
assistant's canonical dependency and lock semantics.

### 7. Grant bounded autonomous Git stewardship

After validation and verification, the kernel may automatically commit exact
touched paths and push to the configured private remote. It refuses dirty or
ambiguous starting state, upstream divergence, force-push, history rewrite,
ambiguous conflict resolution, and any attempt to publish the repository
without explicit authority. Commit-success/push-failure is a receipt-driven
reconciliation state, not permission to duplicate the semantic write.

## Decision Spotlight

- **Define My Friday above the harness:** Codex runs the model/tool loop; My
  Friday owns the versioned assistant baseline, capability and memory
  lifecycle, adapters, conformance, migration, verification, and reversal.
- **Keep the trusted kernel narrow:** setup/reconciliation, repository
  stewardship, memory transactions, capability authoring/validation, harness
  compilation, and launching are baseline control-plane mechanics.
- **Make the two principles concrete:** governed-memory maintenance and
  capability authoring are the first agent-facing dogfood capabilities.
- **Standardize semantic packages:** a capability bundles typed components;
  source, instance config, data, lock, projection, and catalog are distinct.
- **Fail visibly on portability loss:** every compilation emits a fidelity
  report and refuses unsupported required semantics.
- **Start with one private repository:** independently govern `/config`,
  `/memory`, and `/capabilities`, with later external references permitted.
- **Automate Git within a hard boundary:** verified commit/push is allowed;
  force, rewrite, ambiguous conflict handling, and publication are not.
- **Support one real adapter first:** Codex on Apple Silicon remains the MVP;
  shared adapter machinery waits for a second harness.
- **Dogfood by composition:** the mature private deployment becomes public
  baseline plus private overlay through inventory, shadowing, parity evidence,
  and reversible cutover.
- **Protect the MVP from private breadth:** modes, agents, secrets, connectors,
  transports, scheduling, services, and software-delivery machinery are later
  optional capabilities, capability sets, or private composition.

## Conceptual architecture

```text
User-owned assistant
├── Private canonical Git repository
│   ├── /config
│   │   └── identity, policy, capability instances, secret references
│   ├── /memory
│   │   └── governed canonical records and transition receipts
│   └── /capabilities
│       └── harness-independent semantic packages
├── My Friday baseline
│   ├── baseline manifest, migrations, and bootstrap kernel
│   ├── core capability set
│   │   ├── capability authoring
│   │   └── governed-memory maintenance
│   └── validation, conformance, Git stewardship, and recovery
├── Harness compiler
│   └── Codex on Apple Silicon (first supported target)
└── Generated state
    ├── lock and receipts
    ├── fidelity report
    └── replaceable harness projection

Agent harness
└── model loop, tools, sessions, and harness-native state
```

The current two-repository structure supplied useful evidence, but the MVP
selects one private assistant repository. The module boundary is semantic, not
merely a directory convention: config, memory, and capabilities each retain
independently understandable ownership, schemas, authority, versioning,
backup, migration, retention, and removal. Later external capability
repositories remain dependencies, not a reason to fragment the initial
assistant contract.

## Competing options

### A. Distribution and lifecycle layer above harnesses — recommended

My Friday owns the baseline, capabilities, memory, adapters, conformance, and
reversal; Codex owns execution first. This is large enough to host the portable
foundation of a mature assistant and small enough to test as a public product.

### B. Full agent execution engine

My Friday owns the model loop, tool protocol, sessions, scheduling, and runtime
in addition to distribution, memory, and capabilities. This could reduce
harness dependency eventually, but it multiplies scope and competes with mature
harnesses before the core product is proven. Park unless a future decision is
supported by execution-level user needs that adapters cannot satisfy.

### C. Continue as a Codex-only bootstrap toolkit

This protects the present narrow scope but cannot honestly become a portable
baseline beneath a mature assistant. It makes memory and capabilities Codex
features rather than My Friday's defining contracts. Reject.

### D. Copy the complete mature private deployment into public core

This appears to reach dogfood quickly, but imports private assumptions,
confuses optional operations with universal foundation, and creates an
unmaintainable MVP. Reject.

## Candidate outcome map

### F0 — Stabilize the existing instruction-only foundation

- Disposition: selected
- Existing authority: Complete under already-approved issue #74.
- Outcome: The existing deterministic workshop receives fresh review, immutable-candidate acceptance, and release so the re-charter starts from a stable capability source/lifecycle contract rather than stranded work.
- Acceptance: The approved #74 contract passes; no new product scope or profile is added.
- Dependencies: Current implementation and review state.
- Sequence: Immediate and parallel only with discovery; no new unapproved feature work.

### B1 — Canonical assistant repository and bootstrap kernel

- Disposition: selected
- Outcome: A user can create, inspect, reconcile, verify, diagnose, repair, upgrade, roll back, migrate, and remove a versioned baseline backed by one configured private Git repository with independently governed `/config`, `/memory`, and `/capabilities` modules.
- Acceptance: The kernel owns setup/reconciliation, exact-path Git stewardship, schema/migration validation, lifecycle receipts, recovery, and launching; verified changes commit and push to the configured private remote while divergence, ambiguous conflicts, force-push, rewrite, publication, and unrelated state are refused.
- Scope boundary: B1 establishes repository, lifecycle, Git, launcher, and stable
  kernel port primitives only; B2 supplies the capability package/compiler
  semantics and B3 supplies the governed-memory semantics behind those ports.
- Dependencies: Accepted #74 lifecycle assets and a separately approved Solution Design.
- Sequence: First new implementation outcome.

### B2 — Capability package v1, builder, and Codex compiler

- Disposition: selected
- Outcome: A user can conversationally author one harness-independent capability containing a typed component set, validate and version its canonical source, compile it for Codex, inspect its fidelity and authority plan, activate it separately, and reverse the projection without deleting source.
- Acceptance: The package contract separates source, instance configuration, secret references, durable data, dependency lock, catalog, and projection; the compiler classifies every semantic as native, translated, emulated, omitted-optional, or unsupported-required and fails closed on required loss; the builder cannot grant itself Git or activation authority.
- Dependencies: B1 and the stabilized instruction-only lifecycle from F0.
- Sequence: Second; first full contract and compiler dogfood proof.

### B3 — Governed memory core capability

- Disposition: selected
- Existing authority: Supersede the current shape of #52 while preserving its approved product intent and evidence.
- Outcome: A user can capture observations and chronology, quarantine handoffs, propose and govern durable claims, atomically supersede or resolve conflict, recall current authorized attributed context in a fresh task, and maintain the corpus through the official core capability set.
- Acceptance: The capture, governance, recall, and maintenance contracts are expressed through the capability package; sensitivity-aware recall, secret rejection, atomic transition receipts, current/history separation, live-verification gates, index rebuild, migrations, Git reconciliation, and recovery pass conformance; reflection may propose but never silently promote belief.
- Dependencies: B1, B2, and retained governed-memory evidence.
- Sequence: Third; first data-bearing and multi-component dogfood proof.

### B4 — Reproducible core set and first public baseline acceptance

- Disposition: selected
- Outcome: A clean supported machine receives the official baseline, bootstrap kernel, and core capability set and completes full bring-up, fresh-task memory, capability creation/compilation/operation, Git synchronization, update, rollback, migration, and removal.
- Acceptance: The product owner and two independent design partners complete one immutable candidate without operator correction, understand repository/module ownership, memory governance, compilation fidelity, data, permission, secret-reference, and effect boundaries, and choose to retain it after a defined review interval.
- Dependencies: B1, B2, and B3.
- Sequence: Fourth; MVP release gate rather than a hidden integration chore.

### D1 — Private reference compatibility and shadow migration

- Disposition: deferred
- Outcome: A mature private assistant can express its portable baseline as a public My Friday baseline plus a private overlay, run a sanitized parity suite, shadow the existing deployment, and reverse a cutover.
- Acceptance: Identity, memory, capabilities, migrations, verification, rollback, and recovery meet explicit parity thresholds; private services and policy remain optional overlay composition.
- Dependencies: Retained B4 use and a sanitized compatibility inventory.
- Sequence: After MVP; staged compatibility, not a single parity project.

### D2 — Official optional capability sets and richer trust profiles

- Disposition: deferred
- Outcome: Maintainers can ship selected modes, agents, deterministic tools, secret-aware integrations, services, or transports as explicit compatible capabilities or capability sets without enlarging mandatory core.
- Acceptance: Each new profile declares permissions, dependencies, data, secrets, background behavior, migrations, conformance, and reversal.
- Dependencies: B4 dogfood and concrete optional capability demand.
- Sequence: Add one evidence-backed profile at a time.

### D3 — Second harness adapter and contract extraction

- Disposition: deferred
- Outcome: My Friday can install the same assistant-domain baseline through a second real harness adapter, and only then extracts common adapter interfaces justified by both implementations.
- Acceptance: Equivalent identity, memory, capability, lifecycle, and reversal conformance passes on both harnesses; harness-specific features remain explicit.
- Dependencies: Stable B4 and a chosen second harness with users.
- Sequence: Later; never blocks the Codex MVP.

### P1 — Public capability registry or marketplace

- Disposition: parked
- Outcome: A future trusted distribution surface could help users discover and install capabilities from maintainers beyond the local assistant.
- Acceptance: Signing, provenance, moderation, dependency, update, revocation, permission disclosure, and reversal policies are approved and conformance tested before any remote installation.
- Dependencies: Retained core-set use and demonstrated third-party distribution demand.
- Sequence: Reconsider only after B4 and later profile evidence.
- Reason: Distribution trust, signing, provenance, moderation, dependency, update, and revocation policy are unsolved and unnecessary for the first core.

### P2 — Full agent execution engine

- Disposition: parked
- Outcome: A future My Friday runtime could own the model/tool loop when adapter boundaries demonstrably cannot satisfy user needs.
- Acceptance: A separately approved decision proves execution ownership adds necessary user value and can meet model, tool, session, security, lifecycle, and migration obligations.
- Dependencies: Stable public baseline and concrete cross-harness evidence.
- Sequence: Reconsider only after adapter limits are observed in production.
- Reason: No current evidence shows that owning the model/tool loop is necessary to deliver memory, capabilities, lifecycle, portability, or the staged reference migration.

### P3 — Broad audience, hosted service, and broad OS support

- Disposition: parked
- Outcome: Later product lines may make My Friday accessible beyond technical local Apple-silicon users.
- Acceptance: Each audience, operating system, hosted boundary, account model, and support promise receives its own evidence and product decision.
- Dependencies: Retained technical MVP use and prioritized audience demand.
- Sequence: Reconsider after B4 rather than enlarging its acceptance boundary.
- Reason: Consumer onboarding, accounts, hosting, synchronization, Windows, Linux, mobile, voice, and messaging are separate product decisions.

### R1 — Keep the current narrow Codex-toolkit identity indefinitely

- Disposition: rejected
- Outcome: No delivery outcome; the existing toolkit assets remain inputs to the approved baseline direction rather than the permanent product boundary.
- Acceptance: Not applicable; this option conflicts with the decision sought.
- Dependencies: None.
- Sequence: Do not materialize.
- Reason: It conflicts with the newly stated product principles and cannot become the portable public foundation beneath a mature assistant.

### R2 — Put the complete private deployment into core

- Disposition: rejected
- Outcome: No delivery outcome; private operations may later compose as optional bundles or overlays only when separately evidenced.
- Acceptance: Not applicable; copying private breadth is not an MVP strategy.
- Dependencies: None.
- Sequence: Do not materialize.
- Reason: It imports deployment-specific operations and makes the public baseline too broad to understand, validate, retain, or remove.

## First-MVP acceptance journey

On a clean supported Apple-silicon Mac with Codex and Git available, but no
preconfigured My Friday state, a technically capable user can:

1. inspect the product boundary and configure one private remote for the
   assistant repository;
2. preview and create a named assistant with independently governed `/config`,
   `/memory`, and `/capabilities` modules from the official baseline;
3. launch a fresh harness task and inspect the installed baseline/core versions;
4. record an observation and journal event, submit a durable-memory proposal,
   deliberately promote it, atomically supersede it, and recall only the active
   attributed successor in another fresh task;
5. use the core builder to create and test one multi-component capability,
   compile it for Codex, inspect the fidelity report, install it separately,
   invoke it in a fresh task, enhance and upgrade it, disable/enable it, and
   remove it while canonical source remains owned;
6. verify that each source or memory change commits and pushes to the configured
   private remote, then safely refuse an upstream divergence without rewriting
   or ambiguously merging history;
7. detect and safely refuse drift or collision, repair an owned projection,
   and recover one interrupted lifecycle or commit-success/push-failure state;
8. upgrade the complete baseline and core capability set, then roll back to the prior
   compatible generation without data loss; and
9. remove the assistant and replaceable projections while preserving canonical
   source and durable memory according to the explicit removal choice.

No step requires private deployment context, provider-specific source-control
authentication, hidden operator intervention, or unreviewed model authority.

## Success and stop signals

### Continue

- The clean-machine journey passes against one immutable candidate.
- The product owner and two design partners complete it without operator correction,
  understand the ownership and trust boundaries, and retain use after a
  product-owner-selected review interval.
- Governed memory reuses the public capability package/compiler lifecycle,
  enforces sensitivity and current/history separation, and never silently
  promotes durable belief.
- Every baseline/core update has deterministic migration, rollback, recovery,
  and unrelated-state preservation evidence.
- Reference compatibility advances one sanitized parity slice at a time.

### Change

- The one-repository module or named-instance structure creates more friction
  than privacy, portability, or recoverability.
- The proposed baseline manifest or capability-set model duplicates harness-native
  lifecycle without adding user-visible ownership.
- Data-bearing memory cannot fit the shared outer lifecycle without obscuring
  domain-specific safety.
- A second harness proves supposedly portable contracts are Codex projections
  in disguise.

### Pause

- Capability effects, permissions, dependencies, data, migrations, or
  projections cannot be bounded before activation.
- Public/private composition needs private paths, credentials, identities, or
  execution context in public source or release evidence.
- Baseline upgrade or rollback can orphan or silently downgrade durable memory.
- Existing #74 release work cannot be completed or cleanly superseded.

### Stop

- Target users do not retain the baseline after dogfood.
- Native harness features solve the complete memory and capability lifecycle
  with equal ownership, portability, conformance, and reversal.
- My Friday requires bespoke installers for each official capability and the
  shared contract provides no real reuse.
- Replacing a mature private baseline requires an irreversible big-bang move.

## Risks and mitigations

| Risk | Mitigation |
|---|---|
| Vocabulary collapse between assistant, harness, skill, tool, and capability | Adopt the proposed glossary and require architecture/docs to use each term consistently |
| Vendor feature loss hidden as portability | Canonical component semantics, versioned adapter feature matrices, deterministic fidelity reports, and fail-closed required behavior |
| Premature generic adapter framework | Ship one Codex compiler; extract shared implementation only after a second real target |
| Capability privilege escalation | Explicit profiles, declared effects/permissions/dependencies/data, deterministic validation, separate activation, and reversal |
| Memory privacy or belief corruption | Deliberate promotion, authorization-aware recall, secret rejection, provenance, sensitivity/conflict/supersession rules, atomic receipts, migrations, and bounded recall |
| Bootstrap recursion or unsafe self-modification | Builder proposes source; deterministic My Friday lifecycle and explicit human authority retain activation control |
| Harness drift | Versioned compatibility matrix, adapter conformance fixtures, doctor/repair, and bounded supported versions |
| Private-repository divergence or duplicate retry | Exact-path staging, fetch-before-push, no ambiguous merge, and idempotent transition/reconciliation receipts |
| Public/private leakage | Public baseline plus private overlay; sanitized parity claims only; no private paths or credentials in product evidence |
| Dogfood bias | The product owner plus at least two independent design partners and a retention interval |
| Big-bang reference cutover | Inventory, mapping, export/import, shadow run, parity gates, rollback, then bounded cutover |
| Maintenance burden | Tiny bootstrap kernel and core capability set, evidence-backed optional packages, explicit support matrix, no marketplace in MVP |

## Assumptions

- My Friday remains above agent harnesses rather than owning the execution loop.
- The first replacement target is the portable baseline, not every private
  service and operational workflow.
- The MVP remains for technically capable Apple-silicon/Codex users.
- One private canonical repository with independently governed `/config`,
  `/memory`, and `/capabilities` is the MVP topology.
- The bootstrap kernel is limited to setup/reconciliation, repository
  stewardship, memory transactions, capability authoring/validation, harness
  compilation, and launching.
- Governed-memory maintenance and capability authoring are the first
  agent-facing dogfood capabilities.
- The existing instruction-only lifecycle should be stabilized and preserved as
  one profile rather than discarded during the re-charter.
- Verified assistant-repository changes may automatically commit and push only
  to the configured private remote within the selected refusal boundaries.

## Unknowns

- Exact JSON Schema field names, canonical serialization, limits, dependency
  solver, lockfile, and migration encodings for capability package v1.
- Exact canonical agent and hook formats and the first Codex feature/fidelity
  matrix.
- Initial memory retrieval implementation, retention/redaction policies,
  reflection cadence, and review/compaction thresholds.
- Signing, provenance, and update trust beyond exact digests and Git history.
- Adapter compatibility policy for Codex releases.
- Product-owner retention interval and external design-partner recruitment.
- Exact sanitized parity threshold for the private reference deployment.

## Product-authority decisions incorporated

- My Friday remains the distribution/lifecycle and semantic-contract layer
  above harnesses; Codex owns the first execution loop.
- Public baseline and portable core precede optional/private operational parity.
- The MVP remains for technically capable Apple-silicon/Codex users.
- One private assistant repository contains independently governed config,
  memory, and capabilities; later external capability repositories are allowed.
- The shipped core is the narrow bootstrap kernel and two dogfood capability
  surfaces, not a collection of personal integrations.
- Verified changes automatically commit and push to the configured private
  remote inside the explicit no-force/no-rewrite/no-ambiguous-conflict/no-publication
  boundary.
- A capability is a harness-independent semantic package of typed components;
  adapters compile projections and report fidelity rather than silently
  flattening unsupported behavior.

## Gate 1

This candidate is Final for Gate 1 review once repository and GitHub preflight
passes on the exact pushed head. An authorized maintainer approval must target
that exact final head; any later commit makes the signal stale. Approval selects
the semantic boundaries, invariants, and outcome sequence above. It does not
approve implementation, exact JSON field spelling, or Solution Design.

## Privacy and evidence handling

The comparison uses sanitized product contracts and public repository history.
No private paths, credentials, identities, transcripts, tool outputs, provider
configuration, or private execution instructions are required. The mature
deployment is represented only as an architectural capability inventory.
