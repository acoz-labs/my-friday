# Capability Package Contract Reflection

## Decision sought

Define the portable semantic boundary of a My Friday capability before designing
the capability builder or any second harness adapter. The decision is not a
promise that every harness can express every component. It is a promise that My
Friday will describe the capability once, compile it deliberately, and report
what was preserved, transformed, omitted, or refused.

## Evidence from existing ecosystems

OpenAI and Anthropic have independently converged on an installable directory
that can contain several kinds of agent extension:

- OpenAI plugins package skills and optional MCP connections, assets, and
  lifecycle hooks behind `.codex-plugin/plugin.json`.
- Claude Code plugins package skills, agents, hooks, MCP and LSP servers,
  monitors, scripts, and related resources behind
  `.claude-plugin/plugin.json`.
- Both ecosystems distribute packages through marketplace catalogs rather than
  making the marketplace record itself the package.
- Agent Skills already standardizes a useful portable leaf: `SKILL.md` plus
  optional `scripts/`, `references/`, and `assets/` with progressive loading.
- MCP standardizes negotiated tool, resource, and prompt services, not complete
  assistant composition.

The overlap is real, but it is not a universal plugin protocol. OpenAI's own
Claude-plugin conversion guidance must translate agents and commands into
skills, adapt hooks, and reject unsupported configuration. Claude Code also
supports runtime features that OpenAI plugins do not promise. Treating either
vendor directory as My Friday's source of truth would therefore encode silent
loss into migration.

Sources:

- [OpenAI plugin architecture](https://developers.openai.com/plugins/concepts/plugins)
- [OpenAI plugin packaging](https://developers.openai.com/plugins/build/plugins)
- [OpenAI Claude-plugin conversion guidance](https://developers.openai.com/plugins/guides/submit-claude-plugin)
- [Claude Code plugin reference](https://code.claude.com/docs/en/plugins-reference)
- [Claude Code marketplace reference](https://code.claude.com/docs/en/plugin-marketplaces)
- [Agent Skills specification](https://agentskills.io/specification)
- [MCP lifecycle and capability negotiation](https://modelcontextprotocol.io/specification/2025-11-25/basic/lifecycle)

## Recommended vocabulary

### Capability

A versioned, self-contained semantic package that gives an assistant one
coherent ability. It may contain instructions, skills, specialist agents,
event reactions, service declarations, deterministic code, references, assets,
schemas, migrations, and tests. Components are implementation ingredients; the
capability is the user-facing unit of purpose, lifecycle, authority, and
portability.

### Capability source

The canonical package committed under `/capabilities/<id>/` in the assistant
repository. It is harness-independent and never contains generated vendor
projections or secret values.

### Capability instance

One assistant's configuration and authority grant for a capability version.
Instance state lives under `/config`, separately from shareable package source.
It binds user choices, secret references, data roots, enabled state, and granted
permissions without modifying the package.

### Capability set

A resolved, compatible, digest-bound collection of capability versions. `core`
is the mandatory set shipped with a baseline. This replaces the overloaded word
`bundle` when referring to several capabilities; a single capability is already
a bundle of components.

### Adapter

A compiler and verifier for one harness. It consumes capability source plus an
instance and emits a replaceable harness-native projection, receipt, and
fidelity report. It may not change the canonical semantics or grant authority.

### Projection

Generated harness-native files or registrations. A projection is disposable,
reproducible, receipt-bound, and never the only copy of source, configuration,
or durable data.

### Catalog

Optional discovery metadata that points to capability packages or capability
sets. A marketplace is one catalog distribution mechanism. Catalog records do
not own package semantics, installation authority, or user configuration.

## Canonical package shape

The v1 contract should reserve this semantic shape. Exact encodings, limits,
and JSON Schema are Solution Design work, but the component and trust boundaries
belong in Product Discovery.

```text
capabilities/<capability-id>/
├── capability.json            # strict identity, inventory, contracts, requirements
├── README.md                  # human explanation; never runtime authority
├── instructions/             # reusable policy or context blocks
├── skills/                    # Agent Skills-compatible directories
│   └── <skill-id>/
│       ├── SKILL.md
│       ├── scripts/
│       ├── references/
│       └── assets/
├── agents/                    # canonical specialist-agent definitions
├── hooks/                     # canonical event/reaction definitions
├── services/                  # MCP or other protocol service declarations
├── commands/                  # deterministic executables not owned by one skill
├── references/                # package-wide schemas and documentation
├── assets/                    # package-wide templates or static data
├── migrations/                # versioned durable-data/config migrations
└── tests/                     # structural, policy, contract, and behavior fixtures
```

Directories are optional, but every functional entry is explicitly inventoried
by `capability.json`. Unknown functional files and unknown manifest fields fail
validation. A package cannot reach outside its root. Large or shared
dependencies are resolved by declared capability or service dependencies, not
by relative paths into another package.

The existing `instruction-only` package becomes a valid narrow subset: one
skill, deterministic cases, no service, executable, secret, network,
background, or durable-data requirements. Its current source/projection split,
strict parsing, bounded tree, tests, receipts, drift detection, recovery, and
source-preserving removal remain valuable invariants.

## Minimum manifest semantics

The manifest must declare enough information for a builder to create a complete
package and for an adapter to compile or refuse it deterministically.

| Section | Required meaning |
|---|---|
| Identity | schema version, stable ID, human name, package version, description, authorship/license and provenance |
| Purpose | triggers, non-triggers, inputs, outputs, success, failure, and explicit non-goals |
| Components | typed IDs and package-relative entry paths for instructions, skills, agents, hooks, services, commands, references, assets, migrations, and tests |
| Interfaces | named contracts the capability provides and consumes so composition does not depend on paths or vendor names |
| Dependencies | capability version ranges, runtime/tool requirements, protocols, operating-system constraints, and external services |
| Configuration | a typed configuration schema, defaults, mutability, and sensitivity; user values live in `/config`, not the package |
| Authority | declared permissions, effects, approval requirements, destructive boundaries, and actions that may never be automated |
| Secrets | symbolic secret slots, purpose, consuming component, required/optional status, and allowed injection mechanism; never a value |
| Data | owned stores, sensitivity, retention, backup, migration, deletion, export, and concurrency expectations |
| Lifecycle | install, reconcile, enable, disable, upgrade, rollback, recover, export, and remove obligations, including migration compatibility |
| Compatibility | required semantic features and optional enhancements; no claim that a named harness supports them |
| Verification | structural tests, policy tests, conformance cases, behavior fixtures, health checks, and acceptance evidence requirements |

The manifest should be strict and machine-validatable. Descriptive prose may
explain a contract but cannot substitute for typed effects, permissions,
requirements, or tests.

## Component semantics

### Instructions

Context or policy intended to shape the main assistant. Instructions declare
scope and precedence. They are not enforcement. If a rule must fire or block
deterministically, it belongs in a hook or trusted kernel policy.

### Skills

Reusable instructions, references, workflows, and optional scripts loaded into
an agent context. My Friday should preserve the Agent Skills directory format
where possible and add package-level authority outside `SKILL.md` rather than
forking that useful common denominator.

### Agents

Specialist execution roles with an activation description, system
instructions, tool/service access, capability dependencies, isolation needs,
resource bounds, and return contract. A harness without specialist agents
cannot silently flatten a required agent into a skill; that is a semantic
downgrade requiring explicit policy or refusal.

### Hooks

Event-driven reactions with a canonical event, matcher, action, timing,
blocking semantics, input/output schema, timeout, failure policy, and
idempotency rule. Canonical events describe intent such as `session.started`,
`tool.before`, `tool.after`, `agent.started`, `agent.finished`, and
`session.finished`; adapters map them to native events only when behavior is
equivalent.

Prompt instructions are not a safe fallback for blocking hooks. If the target
harness cannot enforce a required guard, compilation fails.

### Services

External functionality exposed through a declared protocol such as MCP. The
package declares the interface, transport options, trust boundary, data
exposure, configuration, and health contract. Credentials remain instance
bindings. MCP should be reused for its actual primitives and capability
negotiation rather than reimplemented inside the package schema.

### Commands

Deterministic executables shared by package components. They declare runtime,
arguments, environment allowlist, filesystem/network effects, expected exits,
timeouts, and platform support. Merely placing a script in a package does not
grant execution authority.

### References and assets

Inspectable non-executable context, schemas, templates, examples, and static
data. The inventory distinguishes material that may enter model context from
material used only by commands or adapters.

### Migrations and tests

Migrations are ordered transformations of capability-owned config or durable
data with preconditions, postconditions, rollback boundaries, and recovery.
Tests include schema validation, package closure, permission/effect coverage,
component contract tests, adapter conformance, and user-visible behavior cases.

## Source, instance, resolution, and projection are separate records

```text
capability source + assistant config + granted authority
                         │
                         ▼
              dependency resolution/lock
                         │
                         ▼
                 harness compilation
                         │
             ┌───────────┴───────────┐
             ▼                       ▼
     native projection       receipt + fidelity report
```

The assistant repository should therefore keep distinct artifacts:

1. source packages under `/capabilities`;
2. user configuration and authority under `/config`;
3. a generated lock/receipt recording exact resolved package and adapter
   digests;
4. replaceable generated projections outside canonical package source; and
5. capability-owned durable records under `/memory` or another manifest-bound
   data root, never inside the package.

This separation permits a user to move the repository, bind machine-local
secrets again, recompile for another harness, and retain the same assistant
semantics and durable data.

## Compilation and fidelity contract

Every adapter exposes a versioned feature matrix. Each component or requirement
is classified during compilation as:

- `native`: emitted with equivalent target semantics;
- `translated`: emitted differently but passes the same conformance contract;
- `emulated`: supplied by trusted My Friday runtime support rather than the
  harness;
- `omitted-optional`: absent with an explicit reason and no required behavior
  lost; or
- `unsupported-required`: compilation fails before projection mutation.

The compiler emits a deterministic plan before activation and a receipt after
activation. Both bind source, instance configuration schema, granted authority,
resolved dependencies, adapter/compiler version, output digests, migrations,
tests, and the complete fidelity classification. Reconciliation compares those
facts with current source and installed state; it never infers ownership from a
pathname alone.

Harness-specific overrides are a last-resort, explicitly nonportable escape
hatch. They must be isolated from canonical semantics, included in the fidelity
report, and cannot broaden authority beyond the instance grant.

## Capability builder contract

The core builder should author this schema conversationally without asking the
user to memorize it. Natural-language collection is input, not authority. The
builder must:

1. identify one coherent user outcome and propose the smallest component set;
2. collect success, failure, triggers, non-triggers, inputs, outputs, data,
   permissions, effects, dependencies, secret slots, and reversal;
3. choose standard components before inventing harness-specific ones;
4. render the complete package and tests into an isolated proposal;
5. validate package closure, schema, authority coverage, and selected adapter
   fidelity;
6. show the exact source diff and activation plan separately;
7. commit and push verified source to the configured private remote under the
   repository-steward policy; and
8. require separate lifecycle authority before installing or changing an
   active projection.

The builder may enhance its own source package, but the trusted kernel owns
validation, repository mutation, compilation, activation, recovery, and
rollback. This prevents a prompt from granting itself execution authority.

## Invariants selected for Gate 1

- A capability is canonical semantic source, never a vendor projection.
- A capability is one coherent ability and may bundle several typed components.
- A capability set, not a capability, groups independently versioned
  capabilities such as the official `core` set.
- Package source, instance configuration, secret bindings, durable data,
  dependency lock, and generated projection are separately owned.
- Secret declarations contain references and policy only, never values.
- Every effect, permission, dependency, data store, migration, and background
  behavior is declared before activation.
- Required semantics either compile with conformance evidence or fail closed.
- Optional loss is visible in the plan, receipt, and fidelity report.
- Generated projections can be deleted and rebuilt without losing source,
  configuration, or durable data.
- The capability builder proposes source; trusted lifecycle machinery retains
  activation and repository authority.
- A marketplace is optional distribution metadata and is not required for the
  MVP or for repository portability.

## Deferred implementation choices

- JSON field spelling, schema modularization, canonical serialization, and
  package size/count limits.
- The exact canonical agent and hook file formats.
- Dependency solver and lockfile encoding.
- Signing and public catalog trust beyond private Git provenance and digests.
- Which hook events the first Codex adapter can express natively or requires My
  Friday to emulate.
- ACP and later protocol bindings, which should be added only against a real
  supported harness or service boundary.
- Whether executable, background, secret-consuming, and remote-integration
  components ship in one later profile or several narrower trust profiles.
