# Solution Design: Deterministic Instruction-Only Capability Workshop

- **Status:** Draft
- **Issue:** #74
- **Planning PR:** Pending
- **Repository basis:** 1724e282393e69e4e023ffc07dd181dee0394bde
- **Execution envelope:** through-production

## Decision

Add `my-friday capability workshop NAME SLUG` as a deterministic, TTY-only
create-and-enhance flow over the existing instruction-only package contract.
One internal proposal model renders the canonical three core files, the user
reviews their complete diff, and exact `Create source` or `Update source`
authorizes a journaled source transaction. The workshop never installs or
upgrades. It finishes with the existing inspect, validate, and deterministic
test contracts and reports the separate lifecycle command, if any.

This replaces the agent-authored source mechanism planned in #56. It reuses the
strict package parser, state inspection, instance confinement, capability lock,
lifecycle commands, acceptance schema, and release path from #51/#56, while
retiring model completion, prompt-input catalog proof, builder-event transport,
and agent-action evidence from the authoring authority chain.

## Needs Attention

- Before implementation, preserve the separate uncommitted #56 worktree and
  salvage only reviewed generic completion/supervisor hardening that remains
  useful after the live-agent path is removed. Do not merge it wholesale.
- Exact-candidate acceptance must replace agent-authorship facts with
  deterministic workshop completion and comprehension facts without weakening
  the existing three-person receipt, ambient-preservation, or artifact binding.
- Existing named instances need an explicit assistant upgrade to receive the
  revised core builder guidance and candidate executable; source is never
  migrated automatically.

## Decision Spotlight

- **One public interface, no proposal file:** answers exist only in memory and
  render the existing package schema. This avoids a second automation contract
  before #75 has reliable agent evidence.
- **Source authority is not activation authority:** exact `Create source` or
  `Update source` changes canonical source only. Install and Upgrade keep their
  existing independent plans and confirmations.
- **No consequential defaults:** display formatting and an existing value may
  be retained explicitly; behavior, triggers, non-triggers, outputs, examples,
  failures, and required facts are never invented.
- **Complete bytes before consent:** the final review shows canonical
  `capability.json`, `skill/SKILL.md`, and `tests/cases.json`, the entire core
  diff, unchanged opaque-file inventory, source action, and installed effect.
- **Same lock, separate journal:** source mutation serializes with lifecycle
  mutation but uses source-specific authority. Lifecycle receipts never grant
  permission to rewrite user-owned source.
- **Signals respect the authority boundary:** exit, EOF, `INT`, and `TERM`
  before exact confirmation write nothing. After confirmation, termination is
  deferred across the bounded commit so recovery observes a complete old or
  complete new package, never partial source.
- **Valid enhancement only:** `ready`, `installed-healthy`, `source-changed`,
  and `disabled` source may be enhanced. Invalid, drifted, collided,
  interrupted, recovery-required, and incompatible states refuse before
  questions begin.
- **Agent adapter remains genuinely deferred:** no Codex client, model prompt,
  dormant adapter protocol, or public spec surface ships in #74.

## Plan Map

- [Context](01-context.md)
- [Decision](02-decision.md)
- [Design](03-design.md)
- [Verification](04-verification.md)
- [Implementation handoff](05-handoff.md)

## Final Gate

The plan may become `Final` only when issue #74 remains the approved D1 outcome,
the complete pack has received independent maintainer review, no blocking
finding remains, validation passes, the planning PR number and current exact
head are recorded, and Anthony approves this plan with the
`through-production` execution envelope.
