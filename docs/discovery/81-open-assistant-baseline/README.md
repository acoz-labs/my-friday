# Discovery: Define The Open Assistant Baseline

- **Status:** Draft for product-owner design session
- **Discovery issue:** #81
- **Discovery PR:** #82
- **Repository basis:** `08b4eb4cd30df4f05bf15ea6c362acac1d48d814`
- **Recommended decision:** Re-charter My Friday as an assistant distribution and lifecycle layer above agent harnesses, with governed memory and capability building as the first bundled core capabilities.
- **Gate 1:** not-ready; three product-authority choices remain
- **Confidence:** Medium
- **Private evidence:** none

## Decision sought

Decide whether My Friday should evolve from a Codex bootstrap/lifecycle toolkit
into the open, user-owned baseline from which a personal assistant is composed,
installed, extended, remembered, repaired, migrated, and removed.

The recommended boundary is deliberately narrower than an execution engine:
Codex is the first **agent harness**, while My Friday owns the assistant's
portable domain contracts, official baseline, capability and memory lifecycle,
harness projection, verification, and reversal. This preserves the successful
Codex/Apple-silicon starting point without making Codex itself the product
definition or prematurely implementing a second harness.

## Product thesis

A durable personal assistant begins with two user-owned systems:

1. **Memory** preserves relevant context across tasks with provenance,
   uncertainty, deliberate promotion, and bounded recall.
2. **Capabilities** turn intent into inspectable, testable, permissioned,
   lifecycle-managed abilities.

My Friday should make those systems available out of the box, then use the same
public contracts to build and operate its official extensions. Identity,
instructions, modes, agents, tools, connectors, transports, and automation can
compose on that foundation; they do not replace it.

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

1. Create a named, user-owned assistant from an explicit baseline.
2. Preview and apply an exact lifecycle plan without disturbing unrelated
   harness, credential, workspace, or user state.
3. Capture observations and chronology, stage candidate beliefs, deliberately
   promote durable memory with provenance and uncertainty, and recall relevant
   context in a fresh task.
4. Define, inspect, validate, test, install, enhance, disable, enable, upgrade,
   recover, and remove a capability without silent activation.
5. Know which files, data, permissions, dependencies, projections, and effects
   each capability owns before activation.
6. Verify, diagnose, repair, upgrade, roll back, migrate, and remove the
   baseline and its official bundle.
7. Compose private identity and policy without putting private deployment state
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
| Release | Native artifact plus manifest-owned instance state | Immutable content-addressed runtime releases | Make the complete baseline and official bundle versioned, verifiable, and reversible |
| Capability | Strict instruction-only source and projection lifecycle | Skills plus deterministic commands, tools, services, and expert agents | Generalize through explicit profiles, not one unsafe universal plugin type |
| Capability authoring | Deterministic terminal workshop | Maintained skill/agent authoring procedures | Bundle capability building as a core capability while deterministic lifecycle remains product authority |
| Memory | Separate repository placeholder and approved future outcome | Governed agent-first repository and local recall/write service | Ship governed memory as the first data-bearing core capability |
| Adapters | Codex-specific files and launch behavior | Many private service/tool boundaries | Keep one real Codex adapter; wait for a second harness before extracting a generic framework |
| Secrets | Intentionally out of the first capability profile | Narrow reference-based secret injection | Keep secrets outside the first baseline; later integration profiles require explicit secret contracts |
| Transports and scheduling | Out of scope | Useful optional operations | Maintain later as optional bundles/adapters, never mandatory core |
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

The minimum versioned assistant distribution that My Friday can install,
verify, diagnose, upgrade, roll back, and remove. The baseline declares its
contract version, official core bundle, compatible adapter, and migrations.

### Capability

A versioned source package that declares its purpose, interface, effects,
permissions, dependencies, data ownership, projections, tests, migrations, and
lifecycle. Different trust and data classes use explicit profiles rather than
an ever-expanding implicit format.

### Core capability

A first-party capability maintained and versioned by My Friday, included in the
official baseline because the assistant is not considered complete without it.
For the first public baseline these are:

1. `governed-memory`
2. `capability-builder`

Install/verify/repair/rollback/remove mechanics are baseline infrastructure,
not a third user capability.

### Capability profile

A bounded trust, execution, dependency, permission, and data class. The
existing `instruction-only` profile remains valid. `data-bearing` is the next
profile required for governed memory. Executable, integration, background, or
credential-using profiles require later explicit product decisions.

### Bundle

A signed or digest-bound set of compatible capability versions distributed and
tested together. `core` is the official mandatory bundle. Other maintained or
user bundles may come later; a public registry or marketplace is not implied.

### Adapter

A maintained projection between My Friday's owned contracts and a specific
harness, operating system, or provider boundary. An adapter translates and
verifies; it does not redefine assistant identity, memory semantics, or
capability authority.

### Overlay

User- or operator-owned identity, policy, configuration, and selected bundles
composed with a public baseline without modifying the baseline source. Private
deployments remain overlays, not forks or hidden core requirements.

### Projection

Adapter-produced installed state derived from canonical source. Projections are
replaceable and verifiable; they are never the only copy of user-owned source
or durable memory.

## Recommended product decision

### 1. My Friday is the distribution and lifecycle layer

Select a layer above agent harnesses. My Friday owns portable assistant
contracts, official baseline composition, capability and memory lifecycle,
adapters, conformance, migration, verification, and reversal. It does not own
the model/tool execution loop in the current product direction.

This gives My Friday an enduring job even as Codex evolves, while avoiding the
cost and risk of building a model runtime before the baseline itself is proven.

### 2. Keep the first core deliberately small

Bundle governed memory and capability building. They embody the two product
principles and allow My Friday to dogfood its own extension system. Everything
else must earn core status through evidence.

Modes, specialist agents, secret injection, connectors, transports,
scheduling, software-delivery orchestration, and background services are
useful optional bundles or adapters. Treating them as mandatory would make the
baseline private-deployment-shaped and difficult to understand or remove.

### 3. Preserve Codex and Apple Silicon as the first supported reality

Ship one excellent Codex adapter on Apple Silicon. Design domain contracts so
Codex details do not become canonical assistant semantics, but do not build a
generic adapter framework without a second real harness and a conformance case.

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

## Conceptual architecture

```text
User-owned assistant
├── Overlay
│   ├── identity and communication policy
│   ├── selected optional bundles
│   └── private configuration and references
├── My Friday baseline
│   ├── baseline manifest and migrations
│   ├── lifecycle engine and receipts
│   ├── core bundle
│   │   ├── capability-builder
│   │   └── governed-memory
│   └── conformance and diagnostic contracts
├── Harness adapter
│   └── Codex on Apple Silicon (first supported adapter)
├── Canonical source
│   └── user-owned runtime and capability packages
└── Canonical durable data
    └── governed memory, separate from replaceable projections

Agent harness
└── model loop, tools, sessions, and harness-native state
```

The current two-repository structure is a good privacy boundary but is not
declared permanent merely because it exists. The durable invariant is stronger
and simpler: shareable runtime/capability source and private durable memory must
have independently understandable ownership, versioning, backup, migration,
and remote policies. Discovery should allow later evidence to choose two
repositories, one repository plus a data store, or another composition that
preserves those invariants.

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

## Roadmap and candidate outcome map

### F0 — Stabilize the existing instruction-only foundation

- **Disposition:** selected; complete under already-approved issue #74
- **Outcome:** The existing deterministic workshop receives fresh review,
  immutable-candidate acceptance, and release so the re-charter starts from a
  stable capability source/lifecycle contract rather than stranded work.
- **Acceptance boundary:** The approved #74 contract passes; no new product
  scope or profile is added.
- **Dependencies:** Current implementation and review state.
- **Sequence:** Immediate and parallel only with discovery; no new unapproved
  feature work.

### B1 — Versioned assistant baseline and Codex adapter

- **Disposition:** selected
- **Outcome:** A user can install, inspect, verify, diagnose, repair, upgrade,
  roll back, migrate, and remove one versioned My Friday baseline through the
  supported Codex/Apple-silicon adapter.
- **Acceptance boundary:** The baseline manifest declares compatible adapter,
  core bundle, migrations, canonical source/data ownership, and exact
  projections; clean-machine and existing-Codex-state scenarios prove
  deterministic lifecycle and non-interference.
- **Dependencies:** Accepted #74 lifecycle assets and a separately approved
  Solution Design.
- **Sequence:** First new implementation outcome.

### B2 — Governed memory core capability

- **Disposition:** selected; supersedes the current shape of #52 while
  preserving its approved product intent and evidence
- **Outcome:** A user can capture observations and chronology, stage and
  deliberately promote a sourced durable claim, and recall relevant attributed
  context in a fresh task through the official core bundle.
- **Acceptance boundary:** A reviewed `data-bearing` profile and the shared
  capability lifecycle own installation, versioning, migration, verification,
  recovery, and reversal. Automatic durable belief promotion, vectors, hosted
  services, and multi-machine synchronization remain excluded.
- **Dependencies:** B1 and retained governed-memory evidence.
- **Sequence:** Second; first data-bearing dogfood proof.

### B3 — Reproducible core bundle and first public baseline acceptance

- **Disposition:** selected
- **Outcome:** A clean supported machine receives the official baseline with
  `capability-builder` and `governed-memory` out of the box, and a user completes
  the full bring-up, fresh-task memory, capability creation/operation, update,
  rollback, and removal journey.
- **Acceptance boundary:** Anthony and two independent design partners complete
  one immutable candidate without operator correction, understand ownership,
  data, permission, and effect boundaries, and choose to retain it after a
  defined review interval.
- **Dependencies:** B1 and B2.
- **Sequence:** Third; MVP release gate rather than a hidden integration chore.

### D1 — Private reference compatibility and shadow migration

- **Disposition:** deferred
- **Outcome:** A mature private assistant can express its portable baseline as
  a public My Friday baseline plus a private overlay, run a sanitized parity
  suite, shadow the existing deployment, and reverse a cutover.
- **Acceptance boundary:** Identity, memory, capabilities, migrations,
  verification, rollback, and recovery meet explicit parity thresholds; private
  services and policy remain optional overlay composition.
- **Dependencies:** Retained B3 use and a sanitized compatibility inventory.
- **Sequence:** After MVP; staged compatibility, not a single parity project.

### D2 — Official optional bundles and richer capability profiles

- **Disposition:** deferred
- **Outcome:** Maintainers can ship selected modes, agents, deterministic tools,
  secret-aware integrations, services, or transports as explicit compatible
  bundles and profiles without enlarging mandatory core.
- **Acceptance boundary:** Each new profile declares permissions, dependencies,
  data, secrets, background behavior, migrations, conformance, and reversal.
- **Dependencies:** B3 dogfood and concrete optional capability demand.
- **Sequence:** Add one evidence-backed profile at a time.

### D3 — Second harness adapter and contract extraction

- **Disposition:** deferred
- **Outcome:** My Friday can install the same assistant-domain baseline through
  a second real harness adapter, and only then extracts common adapter
  interfaces justified by both implementations.
- **Acceptance boundary:** Equivalent identity, memory, capability, lifecycle,
  and reversal conformance passes on both harnesses; harness-specific features
  remain explicit.
- **Dependencies:** Stable B3 and a chosen second harness with users.
- **Sequence:** Later; never blocks the Codex MVP.

### P1 — Public capability registry or marketplace

- **Disposition:** parked
- **Reason:** Distribution trust, signing, provenance, moderation, dependency,
  update, and revocation policy are unsolved and unnecessary for the first core.

### P2 — Full agent execution engine

- **Disposition:** parked
- **Reason:** No current evidence shows that owning the model/tool loop is
  necessary to deliver memory, capabilities, lifecycle, portability, or the
  staged reference migration.

### P3 — Broad audience, hosted service, and broad OS support

- **Disposition:** parked
- **Reason:** Consumer onboarding, accounts, hosting, synchronization, Windows,
  Linux, mobile, voice, and messaging are separate product decisions.

### R1 — Keep the current narrow Codex-toolkit identity indefinitely

- **Disposition:** rejected
- **Reason:** It conflicts with the newly stated product principles and cannot
  become the portable public foundation beneath a mature assistant.

### R2 — Put the complete private deployment into core

- **Disposition:** rejected
- **Reason:** It imports deployment-specific operations and makes the public
  baseline too broad to understand, validate, retain, or remove.

## First-MVP acceptance journey

On a clean supported Apple-silicon Mac with Codex and Git available, but no
preconfigured My Friday state, a technically capable user can:

1. inspect the product boundary and choose source, durable-data, and optional
   remote locations;
2. preview and create a named assistant from the current official baseline;
3. launch a fresh harness task and inspect the installed baseline/core versions;
4. record an observation and journal event, submit a durable-memory proposal,
   deliberately promote it, and recall it with attribution in another fresh
   task;
5. use the core builder to create and test one instruction-only capability,
   install it separately, invoke it in a fresh task, enhance and upgrade it,
   disable/enable it, and remove it while source remains owned;
6. detect and safely refuse drift or collision, repair an owned projection,
   and recover one interrupted lifecycle;
7. upgrade the complete baseline and core bundle, then roll back to the prior
   compatible generation without data loss; and
8. remove the assistant and replaceable projections while preserving canonical
   source and durable memory according to the explicit removal choice.

No step requires private deployment context, provider-specific source-control
authentication, hidden operator intervention, or unreviewed model authority.

## Success and decision signals

### Continue

- The clean-machine journey passes against one immutable candidate.
- Anthony and two design partners complete it without operator correction,
  understand the ownership and trust boundaries, and retain use after a
  product-owner-selected review interval.
- Governed memory reuses the public outer capability lifecycle without a hidden
  installer and never silently promotes durable belief.
- Every baseline/core update has deterministic migration, rollback, recovery,
  and unrelated-state preservation evidence.
- Reference compatibility advances one sanitized parity slice at a time.

### Change

- The two-repository or named-instance structure creates more friction than
  privacy, portability, or recoverability.
- The proposed baseline manifest or bundle model duplicates harness-native
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
| Premature generic adapter framework | Ship one Codex adapter; extract shared interfaces only after a second real implementation |
| Capability privilege escalation | Explicit profiles, declared effects/permissions/dependencies/data, deterministic validation, separate activation, and reversal |
| Memory privacy or belief corruption | Separate durable data, deliberate promotion, provenance, sensitivity/conflict rules, migrations, and bounded recall |
| Bootstrap recursion or unsafe self-modification | Builder proposes source; deterministic My Friday lifecycle and explicit human authority retain activation control |
| Harness drift | Versioned compatibility matrix, adapter conformance fixtures, doctor/repair, and bounded supported versions |
| Public/private leakage | Public baseline plus private overlay; sanitized parity claims only; no private paths or credentials in product evidence |
| Dogfood bias | Anthony plus at least two independent design partners and a retention interval |
| Big-bang reference cutover | Inventory, mapping, export/import, shadow run, parity gates, rollback, then bounded cutover |
| Maintenance burden | Tiny core bundle, evidence-backed optional bundles, explicit support matrix, no marketplace in MVP |

## Assumptions and unknowns ledger

### Assumptions to approve or change

- My Friday remains above agent harnesses rather than owning the execution loop.
- The first replacement target is the portable baseline, not every private
  service and operational workflow.
- The MVP remains for technically capable Apple-silicon/Codex users.
- Governed memory and capability building are the only first core capabilities.
- The existing instruction-only lifecycle should be stabilized and preserved as
  one profile rather than discarded during the re-charter.

### Material unknowns for later outcomes

- Exact baseline/bundle manifest schemas and version negotiation.
- Data-bearing profile permissions, data root, migrations, backup, and recovery.
- Whether source and memory remain two repositories after real baseline and
  migration use.
- Signing, provenance, and update trust beyond exact digests and Git history.
- Adapter compatibility policy for Codex releases.
- Product-owner retention interval and external design-partner recruitment.
- Exact sanitized parity threshold for the private reference deployment.

## Product-authority choices required before Final

1. **Layer boundary:** distribution/lifecycle layer above agent harnesses
   (recommended), or eventual ownership of the model/tool execution loop.
2. **Reference replacement:** public baseline and core first, with services and
   operational machinery as optional/private composition (recommended), or
   full current operational parity before any cutover.
3. **MVP audience:** technically capable Apple-silicon/Codex users
   (recommended), or nontechnical users in the first MVP.

After these choices, the candidate can be made Final, preflighted on its exact
head, presented for Gate 1, and—if approved—materialized into independently
sequenced outcomes. No new re-charter implementation is authorized before that
gate.

## Privacy and evidence handling

The comparison uses sanitized product contracts and public repository history.
No private paths, credentials, identities, transcripts, tool outputs, provider
configuration, or private execution instructions are required. The mature
deployment is represented only as an architectural capability inventory.
