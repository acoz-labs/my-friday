# Solution Decision

## Decision Drivers

The solution must work without a model or network; preserve one package
contract; display complete source before mutation; keep source and activation
authority independent; support create and enhancement; preserve optional bytes;
serialize with lifecycle mutation; recover from partial filesystem work; remain
testable without a real account home; and fit the existing artifact acceptance
and release path without a new dependency.

## Competing Approaches

1. **Thin scaffold plus direct file editing.** Generate the three files once,
   then require users to edit JSON/Markdown directly for enhancement.
2. **Public declarative proposal plus apply.** Define a JSON/YAML spec that both
   terminal and future agents write, then apply it noninteractively.
3. **Deterministic in-memory workshop over the package contract.** Collect and
   validate answers, render canonical bytes, preview the full diff, and commit
   only after an exact source confirmation.
4. **Resume live-agent authoring with stronger postconditions.** Salvage the #56
   native-exec path and wait for a qualifying filesystem effect.

## Adversarial Comparison

Approach 1 is small but fails the repeatable enhancement outcome and makes the
user memorize the schema after first use. It is a scaffold, which discovery
explicitly rejected.

Approach 2 creates a second public schema whose compatibility and automation
semantics would need long-term support. It asks first-time users to understand
that schema and prebuilds #75 without evidence of reliable agent input. An
internal proposal type is useful; publishing it is not.

Approach 3 adds a bounded UI and source transaction while reusing validation,
tests, lifecycle state, and release authority. Its main cost is implementing
safe tree replacement and a deterministic multi-value editor. Those are direct
requirements, testable with existing safety patterns, and do not expand the
trust boundary.

Approach 4 has already failed across supported live surfaces despite trivial
writes succeeding. Model completion and exit zero cannot satisfy a product
flow that must create source reliably. Keeping it would park #51 and governed
memory on external behavior Anthony chose not to await.

## Selected Approach

Select approach 3 with high confidence. Add an internal `capabilityworkshop`
package for answer state, canonical rendering, diffing, interaction, and a
source-specific transaction. Keep `capability.Validate`, `TestCases`, `Inspect`,
and lifecycle execution authoritative; extract only small exported validation
or rendering helpers when duplication would otherwise create contract drift.

The proposal remains ephemeral. The generated core files are the only durable
answer record. Enhancement parses manifest and case fields into proposal
values, separates SKILL frontmatter from its arbitrary valid instruction body,
retains that body byte-for-byte by default, preserves optional files outside
the proposal, and updates only the three core files. Canonical body regeneration
is an explicit edit, never an inference. No future adapter interface ships
until #75 is reshaped from new evidence.

The prior #56 mechanism is classified as follows:

- **Reuse:** instance verification and confinement, manifest-owned candidate,
  strict package/lifecycle, runner stop/quiescence hardening, false-completion
  principle, typed exact-candidate evidence, private-data redaction, partner
  receipts, nomination, and release.
- **Supersede:** generated complete agent authoring contract, literal
  `$capability-builder` selection as authoring proof, prompt-input catalog/path
  preflight, live builder PTY/native-exec transport, model action/event
  classification, model transcript handling, and `BUILDER_SOURCE_READY`.
- **Defer:** conversational answer proposal, supported Codex adapter/protocol,
  and action/effect qualification to #75.

## Decisions Ledger

| Decision | Rationale | Evidence |
|---|---|---|
| Public command is `capability workshop NAME SLUG` | One discoverable create/enhance entry matching existing capability CLI | Issue #74 and `runCapability` |
| Proposal is internal and in-memory | Avoid second durable/public schema and keep source canonical | Discovery #72 D1/D3 |
| Render exactly three core files | Existing validator and projection contract remain authority | `capability.Validate` |
| Optional references/assets are preserved but not edited | Meets enhancement ownership without adding a file editor | Issue #74 acceptance |
| Existing SKILL body is retained by default | Any valid pre-workshop package must enhance without prose loss or reverse engineering | Existing open Markdown contract |
| Source transaction shares the instance capability lock | Prevent source/lifecycle TOCTOU and keep non-blocking concurrency semantics | `capability.Execute`/`Recover` |
| Source journal is separate from lifecycle receipt/journal | Lifecycle authority must never imply permission to rewrite source | ADR 0004 |
| Exact confirmation occurs after full diff | Consent binds reviewed bytes and action | Product direction and existing confirmation pattern |
| Installed effect is state-specific and never automatic | Preserve explicit Install/Upgrade/Enable authority | `capability.Plan` |
| Core builder becomes workshop guidance, not source-writing agent authority | Existing instances remain useful without claiming model reliability | #72 D1 and #75 |
| No new dependency or public spec | YAGNI and offline deterministic operation | Issue #74 dependencies |
