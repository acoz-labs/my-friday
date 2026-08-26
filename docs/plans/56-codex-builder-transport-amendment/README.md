# Solution Design Amendment: Complete Builder Contract And Native Exec Acceptance

- **Status:** Draft
- **Issue:** #56
- **Planning PR:** #71
- **Repository basis:** 7cc30e9ca934660bd3c00f62d7745e5adaded7cf
- **Supersedes:** Builder transport and completion portions of #57 at `b48de378ca59e148dd4d4638eb3879d128f1ec2b`
- **Execution envelope:** through-production

## Decision

Keep the capability builder as an agent-authored, instance-local skill, but make
its generated instructions a complete package-authoring contract. Run the live
authoring acceptance through the instance-owned `codex exec` surface and grant
authority only after candidate-owned inspect, validate, and test postconditions.
Retain a fresh PTY task for installed-capability invocation. This is the
smallest repair that addresses the observed `missing-contract` failure without
adding a scaffold command, weakening confinement, or accepting model prose as
evidence.

## Needs Attention

- The implementation must prove the enriched builder contract actually causes
  Codex to create the fixed package before a new candidate is nominated.
- Codex 0.149 exposes no structured skill input for `exec`; the exact literal
  `$capability-builder` prompt plus the existing catalog/path preflight remains
  the explicit-selection proof.
- Every failed candidate and preserved diagnostic root remains historical and
  ineligible for reuse.

## Decision Spotlight

- **Complete contract in the generated skill:** include the three-file layout,
  exact manifest and case fields, invariants, prohibited-effect vocabulary, and
  authoring/check sequence. The builder should not need repository source code
  or private operator knowledge to discover its public package format.
- **No scaffold command:** the agent still authors and enhances capability
  source. A deterministic generator would replace the capability-builder
  behavior this release is meant to prove.
- **`codex exec` for live authoring:** it is the supported Codex surface intended
  to execute an autonomous bounded task and produces a machine-readable private
  event stream. The TUI remains the proof surface for installed capability use.
- **Postconditions, not completion prose:** zero exit or a model completion
  message is insufficient. Only source existence plus exact-candidate
  `inspect` state `ready`, `validate`, and `test` may produce the normalized
  builder receipt.
- **No dependency upgrade gamble:** 0.149.1 is the newest stable release and its
  changelog identifies no relevant behavioral fix; pre-release 0.150 builds are
  not release prerequisites for My Friday.

## Plan Map

- [Context](01-context.md)
- [Decision](02-decision.md)
- [Design](03-design.md)
- [Verification](04-verification.md)
- [Implementation handoff](05-handoff.md)

## Final Gate

The plan may become `Final` only after the planning PR has received independent
maintainer review, validation passes, its exact head and PR number are recorded,
and product authority approves this amendment with the `through-production`
execution envelope.
