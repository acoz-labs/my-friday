# Discovery: on-demand capabilities and selective delegation

- **Status:** Final
- **Discovery issue:** #104
- **Discovery PR:** pending
- **Repository basis:** b7d8aa154a1f96b2406dc65fd41e23adb5a03ab7
- **Recommended decision:** approve
- **Gate 1:** awaiting-authority
- **Confidence:** Medium
- **Private evidence:** opaque-references

## Decision sought

Refine #81/B2 (#93) from projecting every capability into a harness catalogue
to on-demand discovery through a small harness adapter. Keep canonical packages
portable and choose direct loading or a bounded worker according to the task.
Validate the choice before making it the production default. The decision
horizon is the capability-package design, before its implementation and the
governed-memory migration; the product owner remains accountable for scope.

This amends capability discovery and instruction delivery only. #101/#103
acceptance-policy work, F0 release, and #92 repository portability retain their
existing scope and authority. Their completion must not wait for this experiment.

## Audience and critical tasks

Users with growing capability libraries need accurate routing, compact task
context, preserved governing instructions, and understandable execution results.
Authors need one canonical package whose requirements remain inspectable across
harnesses. Operators must know when an installation cannot execute a capability.

## Evidence

- #81/#82 defines canonical typed capability packages and semantic fidelity;
  #93 is the existing B2 delivery issue and has no planning PR recorded.
- Existing skills use progressive instruction loading. A startup catalogue is
  metadata, not evidence that every capability body is loaded initially.
- Sanitized observation `capability-catalogue-baseline-20260905`: one mature
  installation has 59 custom skills, 19,757 description characters, and 13
  custom agents. This excludes plugin catalogues and is not a token count or
  evidence of degraded performance.
- Sanitized observation `second-harness-access-20260905`: Codex CLI 0.153.4
  and Claude Code 2.1.193 are installed. Claude reports authentication, but a
  no-tool inference probe returned an account-access denial. Installation and
  authentication do not prove a runnable second harness.
- Product authority selected on-demand discovery with selective delegation
  and authorized commencement on 2026-09-05. Performance and portability
  conclusions remain subject to the experiment below.

## Assumptions

Metadata routing costs may grow with a large catalogue. Separating complex
capability work may reduce main-context growth, but can increase total tokens,
latency, and summarization loss. Small capabilities may be cheaper to load
directly. These are hypotheses, not measured benefits or engineering estimates.

## Unknowns

Selection accuracy on ambiguous requests; actual input tokens and latency;
worker-result omissions; capability dependency depth; minimum policy footprint;
and the second harness's supported execution semantics remain unmeasured.
Claude account access must be resolved through an existing authorized route
before its trial. No subscription, new credential, or permission change is
authorized by this discovery.

## Competing options

| Option | Value | Tradeoff | Disposition |
| --- | --- | --- | --- |
| Full native metadata catalogue | Simple routing and existing harness behavior | Startup metadata grows with the library | Retain as measured baseline |
| Lookup with direct loading | Smaller initial catalogue, little delegation overhead | Selected instructions still occupy main context | Select as supported path |
| Lookup with selective bounded workers | Can keep complex operational detail outside main context | Adds handoff cost and possible omissions | Select for experiment |
| Permanent universal capability agent | One visible routing entry | Its own context grows; central routing dependency | Reject as required design |

## Decision

Use a deterministic, rebuildable index derived from canonical capability
definitions and bound to the exact repository revision. Agents may propose
metadata improvements; they are not the authoritative inventory maintainer.
Search and execution must agree on package identity/revision, required
dependencies, execution requirements, and supported semantics. A stale index
must be refreshed or explicitly refused, never silently used for a different
revision.

Each harness receives essential governing instructions and a small capability
access surface. Load only selected capability instructions and their required
dependencies. Keep user authority, task intent, security rules, and essential
memory/routing obligations present in the main context; do not depend on a
search query to discover those obligations.

Use fresh or bounded worker contexts for tasks whose isolation is useful.
Supply a sufficient task brief and applicable policy; return outcome, changes,
verification, limitations, and decisions needed. Direct execution is valid
when it preserves equivalent semantics. Required isolation or other unsupported
semantics must refuse with an actionable explanation, not silently downgrade.

Discovery is not activation or permission. Native hooks, tools, services,
credentials, desktop access, and permissions still require the existing
explicit lifecycle and fidelity checks. The adapter must not grant authority
or make unavailable host abilities appear portable. This is a refinement of
the package/compiler boundary, not a general execution engine.

## Smallest experiment and success signals

Engineering should pre-register a shared, synthetic capability corpus and task
suite before collecting results. Compare three modes in both Codex and actual
Claude Code: full metadata catalogue, lookup/direct loading, and lookup/selective
delegation. Keep corpus revision, essential policy, task brief, environment,
model configuration within each harness, and tool permissions fixed. Record
model/harness versions, repeats, cold/warm conditions, and tool availability;
report cross-harness differences rather than attributing them to routing alone.

Include easy selection, ambiguous alternatives, no matching capability,
dependency loading, stale index, unsupported required semantics, conflicting
instructions, permission denial, short direct tasks, complex worker tasks,
and worker summaries that omit a material failure or change. Score against
predeclared expected actions/results and refusal behavior.

Measure selection accuracy, task success, policy preservation, main-context
input/growth, total input/output tokens including workers and lookup, elapsed
latency, and summary completeness. Report raw counts, variation, and missing
telemetry. Character counts must never substitute for measured tokens.

Continue to default rollout only when the held-out suite has no observed
critical policy/authority loss, task correctness meets the baseline acceptance
boundary, and measured main-context reduction justifies total-token/latency
costs. Engineering must define the numerical budgets and scoring rubric before
the run; do not pick thresholds after seeing the results. Change the routing
policy if workers add cost without benefit. Pause second-harness claims until
the actual second harness completes the suite. Stop rollout on required semantic
loss or unrecoverable lifecycle behavior. A small suite is bounded evidence,
not proof of universal agent independence.

## Candidate outcome map

### O1 — Comparative capability-routing experiment

- Disposition: selected
- Outcome: Maintainers can choose a capability discovery/execution default from reproducible Codex and actual Claude Code evidence rather than catalogue-size assumptions.
- Acceptance: A predeclared synthetic corpus, expected-result rubric, three routing modes, held-out tasks, negative cases, complete token/latency/context reporting, and evidence of policy and summary preservation produce an inspectable comparison; blocked second-harness runs remain explicitly incomplete.
- Dependencies: Existing authorized harness access and isolated disposable evaluation surfaces; no dependency on completion of #101/#103 or #92 for synthetic evaluation.
- Sequence: Now, before B2 solution design; prepare and run available portions while an unavailable harness is resolved without widening account authority.

### O2 — On-demand capability access as the B2 delivery contract

- Disposition: deferred
- Outcome: A user authors one canonical package and discovers/executes it through a small harness access layer with direct loading or bounded delegation while preserving semantic fidelity and lifecycle control.
- Acceptance: Incorporate O1 findings into the B2 plan; specify revision-bound index consistency, dependency loading, essential main-context policy, bounded-worker result contract, equivalent direct fallback, refusal of required semantic loss, and deterministic preview/activation/update/reversal; reconcile #93 as the existing B2 authority instead of creating duplicate execution work.
- Dependencies: O1 comparison and the existing B2 prerequisites in #93, including #92 and F0; materialization must reconcile the supersession relationship explicitly before any implementation dispatch.
- Sequence: Next, amend B2 after comparative evidence; retain the current package contract and native component lifecycle where needed.

### O3 — Broader production harness support

- Disposition: parked
- Outcome: Users can run supported capability contracts on additional production harnesses with declared fidelity and verified lifecycle behavior.
- Acceptance: A separately shaped support matrix and installation/recovery suite justify each production compatibility claim; experiment participation alone does not grant production support.
- Dependencies: Successful O1 and delivered B2 access contract.
- Sequence: Later; the real second-harness experiment is immediate evidence work, while broad production support remains uncommitted.

## Privacy and evidence handling

Use generic synthetic capability contents and tasks in tracked experiment data.
Private instructions, paths, credentials, raw transcripts, and user records are
excluded. The two opaque observations above carry only sanitized inventory and
availability facts. Experimental logs must be scrubbed before repository use.

## Decision Spotlight

The preferred direction is authorized; its performance benefit is unproven.
The index is generated, delegation is selective, essential policy remains in
the main agent, and unavailable semantics never gain permission through lookup.
The experiment precedes B2 design without delaying the independent acceptance
policy or portable repository work. Reassess at O1 completion or an observed
policy failure, whichever occurs first.

## Gate 1

The product owner's direction approval is recorded above. The concrete final
candidate still requires the repository's authorized maintainer review on its
exact head before merge and outcome materialization. No additional product
choice is requested unless evidence changes the approved direction materially.
