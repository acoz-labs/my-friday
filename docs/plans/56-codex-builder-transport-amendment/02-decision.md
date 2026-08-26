# Solution Decision

## Decision Drivers

The repair must preserve real agent authorship, explicit private builder
selection, no activation by the builder, exact-candidate authority, ambient
confinement, private transcript handling, deterministic CI, and the existing
through-production release path. It must address observed behavior rather than
depending on an undocumented Codex upgrade.

## Competing Approaches

1. Wait for or pin a different Codex version and retain PTY marker acceptance.
2. Add a My Friday `capability scaffold` command and have the builder fill a
   generated package.
3. Enrich the generated builder skill with the complete contract and use
   `codex exec` for bounded authoring, followed by candidate postconditions.
4. Accept a fixture or driver-authored package and retain Codex only for
   installed invocation.

## Adversarial Comparison

Approach 1 has no supporting release note or successful probe. The latest
stable is only 0.149.1, while 0.150 is pre-release; changing dependencies would
still leave the missing My Friday contract.

Approach 2 is deterministic but changes the product from agent-authored source
to generator-assisted source and adds a public command not required by issue
#51. It is a credible later usability enhancement, not the smallest repair.

Approach 3 fixes the deployed information boundary: the builder receives the
same strict contract the candidate already enforces. `exec` is designed for a
bounded autonomous task, while deterministic postconditions prevent a zero
exit or model assertion from becoming authority. Its tradeoff is a split proof:
headless Codex authors source, then a fresh PTY proves installed user behavior.

Approach 4 would make the acceptance green while abandoning the core product
claim that users can ask their agent to build a capability. It is rejected.

## Selected Approach

Select approach 3 with medium-high confidence. Expand only the instance-generated
builder instructions; do not add a command or package profile. Replace the live
builder PTY marker path with bounded `codex exec` plus a private machine-readable
event transcript and driver-issued receipt after filesystem and candidate
checks. Preserve the existing PTY path for fresh installed invocation.

If the enriched contract still produces no source in a disposable live probe,
stop before nomination and return to Solution Design. Do not fall back to a
fixture, pre-release dependency, direct driver write, or model prose.

## Decisions Ledger

| Decision | Rationale | Evidence |
|---|---|---|
| Put the complete package contract in the generated builder | The deployed agent cannot inspect repository test fixtures and reported `missing-contract` | `instance.go` versus `capability.go` and live probes |
| Use native `codex exec` for authoring | Supported autonomous surface with bounded exit and private JSON events | Installed 0.149 CLI help and disposable probe |
| Keep literal mention plus catalog preflight | `exec` exposes no structured skill argument | Disposable interface inspection |
| Driver issues completion receipt only after candidate checks | Model completion previously preceded all required effects | Candidate 13 failure and adversarial completion-gate tests |
| Retain PTY for installed invocation | Preserves direct fresh-task, user-facing behavior evidence | Approved #57 journey |
| No scaffold command or Codex upgrade | Avoid product expansion and speculative dependency churn | Existing strict parser; official 0.149.1 release |
