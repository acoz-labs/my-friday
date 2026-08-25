# Discovery: capability builder before core capabilities

- **Status:** Final candidate
- **Discovery issue:** #49
- **Discovery PR:** #50
- **Repository basis:** `ae5c10317cfd8d2a6f866fd7ec39cbf1da82da13`
- **Recommended decision:** change the sequence
- **Gate 1:** awaiting-authority
- **Confidence:** Medium
- **Private evidence:** none

## Decision sought

Decide whether My Friday should establish a general, inspectable capability-
building workflow before completing governed memory, then require governed
memory to dogfood that workflow as the first substantial core capability.

This is a product-sequencing and trust-boundary decision. It does not decide a
universal plugin architecture. It must preserve My Friday's promise that users
can understand, version, extend, repair, and remove what the product manages.

## Recommendation

Change the approved sequence.

Build a narrow capability workshop first: an agent-facing authoring skill plus
a deterministic My Friday lifecycle for source packages. The agent may help a
user describe, scaffold, and improve a capability, but it cannot silently
activate one or grant itself new authority. The user reviews the source and an
exact plan; My Friday validates, tests, installs, verifies, upgrades, disables,
and removes only declared owned surfaces.

Use one shared outer package and lifecycle contract with explicit profiles,
rather than pretending every capability has the same internals. Prove the
initial instruction-only profile with a small capability before activation.
Then add the data-bearing profile required by governed memory and build memory
through that same workshop. Memory remains a core My Friday capability, but it
must not introduce a second bespoke authoring, installation, or reversal
lifecycle.

Do not build dynamic code loading, a marketplace, autonomous self-modification,
arbitrary dependencies, background installation, or a universal permissions
system under this decision.

## Audience and critical tasks

The initial audience remains technically capable Codex users who prefer local
ownership and inspectability. The added critical journey serves both users who
want to teach their assistant something new and My Friday maintainers shipping
a useful core capability set.

A user must be able to:

1. Tell their assistant what new ability they want in ordinary language.
2. Be guided through purpose, triggers, inputs, outputs, permissions, data,
   dependencies, failure behavior, verification, and reversal.
3. Receive a version-controlled capability source package and complete change
   preview, not an opaque installed mutation.
4. Inspect the instructions, scripts or owned code, schemas, tests, and declared
   installation surfaces before activation.
5. Run deterministic validation and tests that do not depend on the generating
   model declaring its own work safe.
6. Install only after explicit confirmation, verify the installed projection,
   and repair or remove it without disturbing unrelated Codex or user state.
7. Ask the assistant to enhance the capability through the same source-first,
   diff-reviewed, versioned workflow.
8. Understand when a capability uses local data, executes code, calls a tool or
   network service, or needs authority that My Friday does not grant.

My Friday maintainers must be able to use the same package and lifecycle to
develop core capabilities. Governed memory is the first substantial dogfood
case because it exercises schemas, sensitive local data, validation, recovery,
and a durable lifecycle rather than merely prompt text.

## Evidence

| Key | Sanitized claim | Implication |
|---|---|---|
| E1 | The approved product direction already names inspectable, validated skill extension as critical task 8 and deferred outcome O4. | Capability authoring is not a new product category; its foundational sequence is new. |
| E2 | The product owner now identifies user-directed capability building and enhancement as a must-have, and wants My Friday's core capabilities to dogfood it. | The current O3-before-O4 sequence may optimize for the first feature while weakening the extensibility promise. |
| E3 | Issue #5's approved design defines a complete memory-specific CLI, schemas, transaction model, acceptance path, and reversal constraints. | Memory provides a demanding test of a shared outer lifecycle, but the current plan assumes that lifecycle is bespoke. |
| E4 | Issue #5 paused before any implementation commit or pull request. Eight untracked core-package files pass focused tests but remain incomplete and unauthoritative. | Sequence can change with low sunk cost; useful domain evidence can be retained without preserving its abstraction. |
| E5 | Issue #6 covers scaffolding, inspection, validation, and local installation of one skill, but not capability definition, enhancement, typed profiles, core dogfooding, or full lifecycle management. | Simply moving issue #6 earlier would underspecify the product decision. |
| E6 | O1 and O2 already establish separate version-controlled runtime/memory source, owned projections, preview, collision refusal, verification, repair, rollback, and removal patterns. | A capability workshop can compose proven product principles instead of inventing an unrelated extension platform. |

## Assumptions

- Users will accept human review and explicit confirmation as the price of
  allowing an assistant to help extend itself.
- Natural-language collaboration is most valuable during capability definition
  and iteration; deterministic tooling is most valuable at validation and
  activation boundaries.
- One outer package contract can describe identity, version, profile, owned
  source, install projection, dependencies, permissions, data handling, tests,
  and reversal without standardizing every internal implementation.
- Explicit profiles can prevent the first release from granting data-bearing or
  executable capabilities the looser rules of instruction-only skills.
- A small instruction-only capability can prove the workshop mechanics without
  becoming a fake substitute for the memory dogfood requirement.
- The paused memory package contains reusable domain concepts, but no file or
  API is entitled to survive the new design unchanged.

## Unknowns

- Whether Codex's native skill format is stable and expressive enough to be the
  complete instruction-only profile or should be one projection from a My
  Friday source package.
- Which capability kinds are honest for the initial profile set. Candidate
  distinctions are instruction-only, local scripted tool, data-bearing core,
  and externally connected; this discovery selects only the first and the
  minimum data-bearing profile required for memory.
- Whether governed memory is the right first substantial dogfood case or too
  complex to isolate builder defects from capability defects.
- Which validation properties can be proven structurally and which still need
  human review, product-design review, or security review.
- Whether capability source belongs only in the runtime repository or whether a
  data-bearing capability may also own a versioned contract in the memory
  repository without coupling private data to shareable source.
- Whether enhancement should be modeled as source replacement, migrations, or
  a versioned capability upgrade when durable data contracts change.
- How users will compare generated changes when they are not comfortable
  reviewing code despite belonging to the initial technical audience.

These unknowns belong in Solution Design and pilot measurement once the outer
product decision is approved. They do not justify a universal plugin platform.

## Competing options

### A. Continue bespoke governed memory, then add skill building

Resume issue #5 under planning PR #20 and deliver issue #6 afterward. This is
the shortest route to memory and follows the approved sequence. It also allows
the first major capability to establish memory-specific authoring, validation,
installation, upgrade, and reversal patterns before the extension workflow is
defined. Later dogfooding could become a retrofit or a cosmetic wrapper.

Disposition: rejected if this candidate is approved.

### B. Move the existing one-skill workflow ahead of memory unchanged

Deliver issue #6 first, then resume issue #5. This proves native skill
scaffolding but does not define a broader capability contract, enhancement
workflow, typed trust profiles, or a requirement that memory use the same
lifecycle. It improves sequence without answering the product owner's core
concern.

Disposition: rejected as incomplete.

### C. Capability workshop with typed profiles and memory dogfood

Create a source-first capability workshop combining natural agent guidance with
deterministic lifecycle enforcement. Start with an instruction-only profile,
then introduce only the data-bearing core profile needed to build governed
memory. Share the outer lifecycle while allowing profile-specific schemas,
tests, permissions, data, and recovery.

Disposition: selected.

### D. Universal dynamic plugin platform

Define arbitrary executable capability bundles, dependency resolution, runtime
loading, permissions, distribution, remote installation, and a marketplace
before any core capability. This would postpone user value, enlarge the attack
surface, and contradict the narrow local-first product boundary.

Disposition: rejected.

### E. Documentation-only capability authoring

Publish a guide or prompt that teaches the agent to write skills. This is easy
to ship and useful as supporting documentation, but it cannot prove ownership,
validation, activation, upgrade, verification, or reversal. The generating
agent would effectively judge and install its own work.

Disposition: rejected as the product lifecycle, retained as part of the
agent-facing workshop.

## Product model

### Capability source package

A capability begins as source in the generated runtime repository. Its package
declares at least:

- stable identity, human name, purpose, version, and capability profile;
- intended user tasks, triggers, inputs, outputs, and failure behavior;
- source files and exact installed projections owned by the capability;
- tools, executable code, dependencies, permissions, and network/data effects;
- data location, sensitivity, retention, migration, and recovery declarations
  where the selected profile permits data;
- deterministic validators, tests, verification checks, and removal behavior;
- compatibility requirements and explicit non-goals.

The package is not installed state. It is inspectable, version-controlled input
to a deterministic plan. Unknown fields, files, projections, dependencies, or
profile privileges fail closed.

### Agent-facing workshop

My Friday installs a narrow core authoring skill that teaches the assistant to:

1. clarify the requested capability and surface trust/data questions;
2. select only a supported profile;
3. scaffold or modify capability source without activating it;
4. present the source diff, unresolved human judgments, validation command, and
   rollback implications;
5. ask the user to run or authorize the deterministic My Friday lifecycle.

The skill may help write content and tests. It does not approve its own output,
hide diffs, invoke unsupported dependencies, widen profiles, or silently run an
install/upgrade command.

### Deterministic lifecycle

The My Friday CLI owns `inspect`, `validate`, `test`, `install`, `verify`,
`upgrade`, `disable`, and `remove` semantics. Mutating actions reuse O2's
preview, exact ownership, collision refusal, transaction, recovery, and
reversal principles. Enhancement always changes source first, shows the diff
and migration/reversal plan, and re-enters validation; it is never an in-place
mutation of installed files.

### Typed profiles

The shared contract is an outer lifecycle, not proof that all capabilities are
equally trusted.

- **Instruction-only v1:** native Codex skill content and declared supporting
  resources, with no executable code, external dependencies, network access,
  credentials, background process, or durable user data.
- **Data-bearing core v1:** selected only for governed memory after separate
  Solution Design. It adds declared private data roots, immutable schemas,
  sensitivity, migrations, recovery, and stronger acceptance/security review.

Scripted tools, external connectors, arbitrary binaries, background services,
credential use, marketplace distribution, and remotely fetched packages remain
outside this decision and require independent discovery.

## Experience sketch

The intended interaction is conversational but not opaque:

```text
User: Teach Friday to prepare a weekly project retrospective.

Friday: I can draft that as an instruction-only capability. It will read only
the context you give the task, produce a Markdown retrospective, install one
reviewable skill, use no network service, and store no durable data. I still
need your preferred sections and success check.

...source package and tests are drafted...

Friday: The source diff is ready. My Friday validation passes. The install plan
would add one owned skill projection and no dependencies. Review the diff, then
run the displayed install command; installation will require its exact
confirmation word.
```

For an enhancement, the assistant starts from the installed capability's
versioned source, explains changed behavior and trust/data effects, adds or
updates tests, and produces a new plan. If the requested change needs a broader
profile, new dependency class, network access, credentials, or data migration,
the workflow stops and identifies the missing authority rather than smuggling
it into an ordinary update.

## Decision and sequencing

If approved, this discovery supersedes the current O3-before-O4 sequence while
preserving O1 and O2 as shipped foundations.

1. Establish the capability package, agent-facing workshop, and deterministic
   local lifecycle for the instruction-only profile.
2. Prove that lifecycle with one deliberately small capability and independent
   user completion; the proof is not counted as the governed-memory outcome.
3. Design the minimum data-bearing core profile and governed memory together.
4. Build governed memory through the workshop, reuse the outer capability
   lifecycle, and retain its stricter domain-specific schemas, sensitivity,
   transaction, recovery, and acceptance requirements.
5. Use evidence from those two profiles before considering scripted tools,
   external integrations, distribution, or broader capability types.

Issue #5 remains paused and its merged Solution Design PR #20 is not
implementation authority after this candidate is approved. Its memory-domain
decisions and partial uncommitted evidence may inform the replacement design,
but the implementation must not resume until a new delivery outcome and plan
bind memory to the capability contract.

Issue #6 remains valid evidence of the earlier, narrower skill outcome but is
superseded as a delivery contract by the selected capability-workshop outcome.
Neither issue is closed or rewritten before exact-head Gate 1 approval.

## Candidate outcome map

### C1 — Build and operate one inspectable capability

- **Disposition:** selected
- **Outcome:** A user can collaborate naturally with their assistant to define,
  scaffold, inspect, validate, test, locally install, verify, enhance, disable,
  and remove one instruction-only capability through a source-first workflow.
- **Acceptance:** The capability source and complete diff are inspectable and
  versioned; deterministic validation and tests pass; installation mutates only
  declared owned projections after exact preview and confirmation; enhancement
  re-enters the same source/review/test lifecycle; complete reversal is proven.
  No executable code, arbitrary dependency, network access, credentials,
  background process, durable user data, publishing, or marketplace is allowed.
- **Dependencies:** Shipped O1 repository source and O2 installed-baseline
  ownership/lifecycle patterns; a supported Codex skill contract; product-
  design and security review of the natural authoring and approval boundary.
- **Sequence:** First and P1. This replaces the narrower deferred O4 contract
  and establishes the product's extension substrate before another core
  capability is defined.

### C2 — Build governed memory as the first data-bearing core capability

- **Disposition:** selected
- **Outcome:** A user can use the capability workshop to create and operate My
  Friday's governed-memory capability, capture activity, deliberately promote a
  durable claim, and recall relevant attributed context in a fresh task.
- **Acceptance:** The capability package passes the shared lifecycle and its
  stricter data-bearing profile; observation, journal, proposal, deliberate
  promotion, validation, recovery, bounded recall, upgrade, verification, and
  reversal work without a parallel bespoke installation system. Automatic
  durable promotion, vectors, external services, and synchronization remain
  excluded.
- **Dependencies:** Accepted C1, a separately approved data-bearing profile and
  memory Solution Design, and the retained provenance/sensitivity/recovery
  evidence from issue #5 without assuming its current implementation shape.
- **Sequence:** Second. This supersedes the current issue #5 delivery contract
  and is the first substantial dogfood proof.

### C3 — Evaluate local scripted capabilities

- **Disposition:** parked
- **Outcome:** Evidence may later determine whether locally executable tools can
  fit a bounded capability profile without turning My Friday into a general
  plugin runtime.
- **Acceptance:** No generated or third-party executable code is activated under
  C1 or C2 authority.
- **Dependencies:** Sustained use of C1 and C2, explicit executable/dependency
  trust design, and separate product/security approval.
- **Sequence:** Later, only if instruction-only and data-bearing core profiles
  demonstrate real demand for a third type.

### C4 — Distribute or install third-party capabilities

- **Disposition:** rejected for the current direction
- **Outcome:** My Friday remains local and source-first rather than becoming a
  marketplace, remote registry, dependency resolver, or trust broker.
- **Acceptance:** No publishing, discovery catalog, remote package retrieval,
  signature authority, ratings, payments, or third-party automatic update
  mechanism is introduced.
- **Dependencies:** A wholly new product decision and security/distribution
  model.
- **Sequence:** Not planned.

### C5 — Allow autonomous self-modification

- **Disposition:** rejected
- **Outcome:** The assistant may help author capabilities but cannot silently
  expand its own instructions, tools, data access, permissions, or installed
  state.
- **Acceptance:** Human-visible source, diff, validation, and explicit mutation
  confirmation remain mandatory; generating work never grants approval.
- **Dependencies:** None; this is a product trust boundary.
- **Sequence:** Not planned.

## Success, change, pause, and stop signals

### Continue

- The product owner and at least two design partners independently create,
  install, enhance, verify, and remove an instruction-only capability while
  correctly explaining source versus installed state.
- Governed memory uses the same outer package and lifecycle without concealing
  a second installer or weakening its stricter data protections.
- Users experience conversational authoring as helpful while still noticing and
  understanding permissions, data, dependencies, tests, and reversal.
- Capability enhancement produces reviewable source changes and reliable
  upgrades rather than configuration drift.

### Change

- The outer contract is useful but memory proves that more explicit profiles
  are required.
- Users can operate source packages but cannot review code-level diffs; the
  workshop must improve explanations or restrict generated surfaces.
- Native Codex skill compatibility changes enough that the source package needs
  a projection adapter rather than direct format ownership.
- The small instruction-only proof passes but does not exercise enough of the
  lifecycle to predict memory dogfooding risk.

### Pause

- Validation cannot deterministically bound package files, projections,
  dependencies, permissions, data effects, or reversal before activation.
- The assistant can cause activation without a distinct human confirmation.
- O2 ownership and recovery primitives cannot be composed without unsafe
  duplication or broadening.
- A data-bearing profile cannot keep private data separate from shareable
  capability source.

### Stop

- The workshop merely renames normal bespoke development and adds no user-
  visible authoring or lifecycle value.
- Codex-native tooling already supplies the complete source, validation,
  installation, enhancement, verification, and reversal experience within the
  required trust boundary.
- Building the abstraction materially delays useful capabilities without reuse
  across instruction-only and governed-memory cases.
- Users prefer curated core capabilities and do not attempt or retain their own
  extensions.

## Privacy and trust handling

This discovery uses public, sanitized product evidence and no private evidence
references. The paused implementation is described only by repository-safe
state and must not be copied into the discovery branch.

Capability definitions must disclose data and transmission effects. Local
source does not imply private execution: models, tools, remotes, plugins, and
future connectors may transmit content. Instruction-only v1 forbids durable
user data and external effects; governed memory's data-bearing profile remains
local, owner-only, and separately reviewed. Capability source never stores
credentials or user memory.

The assistant is an authoring collaborator, not the authority that approves its
own capability. Deterministic validation is necessary but not a claim that
generated content is correct, non-malicious, or fit for every user.

## Decision Spotlight

The recommended abstraction is intentionally asymmetrical:

- natural language and model assistance help the user define and improve a
  capability;
- versioned source and explicit profiles make the result inspectable;
- deterministic My Friday tooling governs activation and reversal;
- human review remains the approval boundary;
- capability-specific design still owns domain behavior and risk.

This preserves the compelling part of the product owner's vision—an assistant
that can help grow its own useful abilities—without quietly promising an
assistant that may rewrite or re-authorize itself.

## Gate 1

This candidate recommends approving C1 and C2 as selected outcomes, parking C3,
and rejecting C4 and C5. Approval changes the product sequence and authorizes
materializing replacement delivery issues only after an authorized maintainer
approves the exact final head and the discovery PR merges.

Approval does not itself authorize implementation, install generated code,
publish capabilities, resume issue #5, close issues #5 or #6, or discard the
paused implementation evidence. Those actions require the verified outcome
map and the repository's normal Solution Design and delivery gates.
