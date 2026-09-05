# Capability routing experiment

This developer-only experiment compares three native capability-routing modes
over one frozen synthetic corpus. It informs later design work; it does not
change My Friday's public CLI, installed assistants, capability format, or
default routing policy.

Current status: the offline corpus, preregistration validator, counted transport
primitives, scorer, and reporter are complete. The live native trial driver and
actual cross-harness comparison are incomplete. The bounded native-driver probe
found no eligible live route for either installed harness, so no inference was
started and the result is `inconclusive`.

## Frozen method

The preregistration contains 24 generic capabilities and 24 tasks, split into
12 development and 12 held-out cases. Each split has one case for explicit
selection, informal paraphrase, ambiguous alternatives, no match, dependency
loading, stale index, unsupported semantics, conflicting instructions,
permission denial, short direct work, complex worker work, and material summary
omission. Labels are separate from model-visible staging and declare permitted
capability sets, execution/clarification/refusal, required facts and fixture
effects, forbidden effects, and required summary material.

The 288 primary cells cover two harnesses, three modes, 24 tasks, and two
repetitions with rotated mode order and cold/warm fixture-cache labels. The
manifest binds the preregistration method/corpus basis commit
`1219e0d8cf892fa02ba80c8b40911648a4c15b58`, corpus revision
`routing-corpus-v1`, exact corpus/rubric hashes, harness versions and models,
budgets, and every cell. The attempt/result binding uses the compact canonical
JSON SHA-256
`c87d185281be47be6e2ec07978803dfb736f06b5adcfa4bd964f72a3834f0ee8`.
The human-formatted file checksum is separately
`b303bdee1061ffeb6bf69fab8ec15d2a921299103373c211d5ece8d22f764fe1`;
whitespace is deliberately not part of the canonical identity.

Issue [#106](https://github.com/acoz-labs/my-friday/issues/106) owns the
experiment. Approved discovery [#105](https://github.com/acoz-labs/my-friday/pull/105)
owns outcome O1, and planning pull request
[#107](https://github.com/acoz-labs/my-friday/pull/107) preserves design
provenance.

## Modes and limits

- `native-catalogue` stages all 24 valid native skill packages and relies on
  native progressive discovery/body loading. Complex work may use one native
  worker only when its identity and inherited boundary are proven.
- `lookup-direct` stages only revision-bound access instructions. The counted
  controller transport returns top-three metadata on lookup and bodies only on
  an explicit load. Frozen BM25 uses `k1=1.2`,
  `b=0.75`, Unicode lowercase letter/digit tokens, the declared IDF,
  ID-ascending tie breaks, top three, and no result for zero positive scores.
  Selected bodies execute in the root; isolation-required work must refuse.
- `lookup-worker` uses the same counted lookup/load/fallback transport, but complex or
  isolation-required work requires one real native worker at depth one.

The counted transport is an offline controller primitive and staging contract;
it is not wired to either native harness as a live MCP tool while the required
sandbox, tool, worker, and process controls remain unproven. No live metadata or
capability body delivery is claimed by the current evidence.

Limits are two selected capabilities plus one dependency, one broader-metadata
fallback, eight tool calls, one worker, 120 seconds, 30,000 observed aggregate
input/output tokens, and a requested 2,000 output-token ceiling per agent. A
live batch is at most 12 trials, 30 minutes, and 360,000 observed tokens.
Observed token thresholds may overshoot one in-flight event and are not
provider billing caps.

## Contracts and scoring

All JSON is strict version 1. Unknown fields, trailing values, duplicates,
unsafe paths, stale revisions, dependency cycles, changed corpus hashes, and a
non-Cartesian manifest refuse. Model-visible staging contains only the current
sanitized task, essential root policy, permitted fixture bytes, and the mode's
native skills or counted transport access instructions. Split/category fields,
catalogue metadata, other tasks, labels, scorer, Git history, and implementation
source remain outside that root.

Completed trials receive separate binary route, task, policy, and summary
scores. Claimed facts do not prove writes: the controller captures complete
post-trial fixture content, the scorer compares it with the frozen inputs, and
every derived change must belong to the task's declared write paths. Reported
digests are computed from that evidence rather than accepted as input. Generic
write effects normalize to the `fixture-write` policy class. Summaries
must contain the frozen material changes, failures, verification, and
limitations in the corresponding section; non-null empty arrays do not pass.

Usage keeps root cumulative input, worker cumulative input, total input/output,
cached input, peak root per-request input, and actual window occupancy distinct.
Aggregate events are never added to components. Missing child usage fails total
cost completeness. Cache, peak, and occupancy are complete only when every
contributing event reports them. Worker starts, usage, and returns retain
matching event and agent identities. Imported metrics are revalidated for
nullability, provenance, nonnegative values, counts, and aggregate accounting.
Missing values are `null` with provenance and a reason, never zero. Peak request input is not called context-window occupancy
unless the harness reports that stronger metric.

Unavailable, failed, invalid, retry, and missing rows stay visible and are not
treated as cheaper completed work. Only primary attempts enter automatic
scores. The recommendation thresholds approved in #107 remain the decision
rule; missing held-out cells or telemetry make the result `inconclusive`.
Each score records its frozen expected disposition and whether that task/mode
belongs in the performance denominator. Matching expected execution,
clarification, and refusal outcomes can be paired; an observed disposition must
match that expectation in both cells. Correctly refusing isolation-required
work in `lookup-direct` can pass routing and summary checks, but its native
baseline expects execution, so the frozen task/mode contract excludes that cell
from performance rather than treating refusal as cheaper completed work. The
denominators therefore come from frozen expectations, never observed outcomes.
Performance eligibility is independent of the 22/24 task-quality floor and the
24/24 summary requirement. Recommendation completeness is evaluated across all
required metrics in both harnesses before a quality or threshold failure can be
reported; missing or nonpositive baselines remain `inconclusive`.

## Driver boundary and current evidence

Live `run` requires explicit `--live`. Authentication is opaque and externally
provisioned; the tool does not inspect, copy, edit, or initiate it. The current
probe invokes only `--version` and help for one installed version of each CLI.
It does not start inference.

| Harness | Version | Current state | Why live cells are unavailable |
| --- | --- | --- | --- |
| Codex | `codex-cli 0.153.4` | `unavailable` | No demonstrated OS-enforced fixture-only read boundary, constrained native body reader, native worker inheritance/pre-launch limit, built-in pre-dispatch rejection, or guaranteed immediately-detached-descendant containment. Workspace-write is not restricted-read proof. |
| Claude Code | `2.1.193` | `unavailable` | The same controls are unproven, including guaranteed immediately-detached-descendant containment. Tool allowlist flags are not OS-backed read/network denial, and disabling native skills/agents is not an eligible A/C baseline. |

The controller includes a stable PID/start-time supervisor for both trials and
credential-free version/help probes, timeout/cancel
cleanup, sampled escaped-descendant tracking, unrelated-process canaries, strict
telemetry parsing, and immutable exact-match resume. These primitives do not
prove an immediately detached child cannot escape between process-table samples
and therefore do not themselves prove a harness boundary. A supported driver still needs the native
canaries and pre-dispatch controls above; flags or simulated workers are
insufficient.

A `supported` probe must report the exact preregistered executable version.
Only an `unavailable` probe may preserve an observed missing or different
version with an explicit reason; that exception can never authorize completed
attempts.

## Commands

From the repository root:

```sh
go run ./tools/capability-routing-experiment validate \
  --data tools/capability-routing-experiment/testdata

go run ./tools/capability-routing-experiment prepare \
  --data tools/capability-routing-experiment/testdata \
  --source-commit 1219e0d8cf892fa02ba80c8b40911648a4c15b58 \
  --out /path/to/new/manifest.json

go run ./tools/capability-routing-experiment prepare \
  --data tools/capability-routing-experiment/testdata \
  --source-commit 1219e0d8cf892fa02ba80c8b40911648a4c15b58 \
  --out /path/to/new/manifest.json \
  --trial <manifest-trial-id> \
  --stage-root /path/to/new/model-root

go run ./tools/capability-routing-experiment run --live \
  --data tools/capability-routing-experiment/testdata \
  --run-root /path/to/new/evidence-root
go run ./tools/capability-routing-experiment score \
  --data tools/capability-routing-experiment/testdata \
  --run-root /path/to/evidence-root
go run ./tools/capability-routing-experiment report \
  --data tools/capability-routing-experiment/testdata \
  --run-root /path/to/evidence-root

# Verify canonical identity first, formatted-file checksum second.
jq -cj . tools/capability-routing-experiment/testdata/manifest.json | shasum -a 256
shasum -a 256 tools/capability-routing-experiment/testdata/manifest.json
```

`prepare` accepts only the compiled preregistration-basis commit and trusted
harness identity. Attempt evidence separately records the executing runner's Go
VCS revision and modified state. Missing or dirty provenance is explicit and
cannot support a completed live trial. Evidence writes are create-or-verify:
exact immutable resumes are
idempotent, while changed content refuses. Use a fresh owner-only model/run root.
Raw model streams, environment data, authentication paths, session identifiers,
and private paths are not allowlisted report fields and must not be committed.

## Results

The frozen preflight lives in
[`results/2026-09-05-native-preflight/`](results/2026-09-05-native-preflight/):

- `probes.json` is the per-version native control matrix;
- `attempts.json` preserves all 288 unavailable rows and per-metric null reasons;
- `scores.json` is deterministic per-cell scoring output;
- `report.json` is the sanitized machine-readable comparison; and
- `report.md` is the readable coverage summary.

`report` reloads the frozen bundle and attempts, deterministically recomputes
the scores, and validates the exact two probe identities before freezing either
format. Coverage includes aggregate and per-category denominators.

This preflight was invoked from clean source-control head
`22bb6a6d01d691460d7d496ab90258aa181343a4`, but the `go run` build did not
expose a VCS revision through Go build information. The attempt receipt records
that limitation rather than inferring the revision, which also makes it
ineligible for completed live evidence. No live trial was attempted.

All 288 rows are unavailable, including the second harness. There are no model
task, route, policy, summary, token, latency, or context observations. This is
an inspectable partial result, not the actual comparison required to choose a
portable default. It authorizes no installation, adapter, production change,
or release.
