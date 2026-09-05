# Capability routing experiment

This developer-only experiment compares three native capability-routing modes
over one frozen synthetic corpus. It informs later design work; it does not
change My Friday's public CLI, installed assistants, capability format, or
default routing policy.

Current status: the offline runner, corpus, rubric, manifest, scorer, and report
are complete. The actual cross-harness comparison is incomplete. The bounded
native-driver probe found no eligible live route for either installed harness,
so no inference was started and the result is `inconclusive`.

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
manifest binds runner/corpus source commit
`b1056c968de49ff8643853015226350f1a218e17`, corpus revision
`routing-corpus-v1`, exact corpus/rubric hashes, harness versions and models,
budgets, and every cell. Its SHA-256 is
`4f0256b20c8e03ca3913286ceadec85f89b150ca97a791ecbd08f5526f160890`.

Issue [#106](https://github.com/acoz-labs/my-friday/issues/106) owns the
experiment. Approved discovery [#105](https://github.com/acoz-labs/my-friday/pull/105)
owns outcome O1, and planning pull request
[#107](https://github.com/acoz-labs/my-friday/pull/107) preserves design
provenance.

## Modes and limits

- `native-catalogue` stages all 24 valid native skill packages and relies on
  native progressive discovery/body loading. Complex work may use one native
  worker only when its identity and inherited boundary are proven.
- `lookup-direct` exposes revision-bound metadata. Frozen BM25 uses `k1=1.2`,
  `b=0.75`, Unicode lowercase letter/digit tokens, the declared IDF,
  ID-ascending tie breaks, top three, and no result for zero positive scores.
  Selected bodies execute in the root; isolation-required work must refuse.
- `lookup-worker` uses the same lookup policy, but complex or
  isolation-required work requires one real native worker at depth one.

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
native skills or metadata. Split/category fields, other tasks, labels, scorer,
Git history, and implementation source remain outside that root.

Completed trials receive separate binary route, task, policy, and summary
scores. Claimed facts do not prove writes: every required write needs a
safe-path before/after fixture diff with distinct SHA-256 digests. Summaries
must contain the frozen material changes, failures, verification, and
limitations in the corresponding section; non-null empty arrays do not pass.

Usage keeps root cumulative input, worker cumulative input, total input/output,
cached input, peak root per-request input, and actual window occupancy distinct.
Aggregate events are never added to components. Missing child usage fails total
cost completeness. Cache, peak, and occupancy are complete only when every
contributing event reports them. Missing values are `null` with provenance and
a reason, never zero. Peak request input is not called context-window occupancy
unless the harness reports that stronger metric.

Unavailable, failed, invalid, retry, and missing rows stay visible and are not
treated as cheaper completed work. Only primary attempts enter automatic
scores. The recommendation thresholds approved in #107 remain the decision
rule; missing held-out cells or telemetry make the result `inconclusive`.

## Driver boundary and current evidence

Live `run` requires explicit `--live`. Authentication is opaque and externally
provisioned; the tool does not inspect, copy, edit, or initiate it. The current
probe invokes only `--version` and help for one installed version of each CLI.
It does not start inference.

| Harness | Version | Current state | Why live cells are unavailable |
| --- | --- | --- | --- |
| Codex | `codex-cli 0.153.4` | `unavailable` | No demonstrated OS-enforced fixture-only read boundary, constrained native body reader, native worker inheritance/pre-launch limit, or built-in pre-dispatch rejection. Workspace-write is not restricted-read proof. |
| Claude Code | `2.1.193` | `unavailable` | The same controls are unproven. Tool allowlist flags are not OS-backed read/network denial, and disabling native skills/agents is not an eligible A/C baseline. |

The controller includes a stable PID/start-time supervisor, timeout/cancel
cleanup, escaped-descendant tracking, unrelated-process canaries, strict
telemetry parsing, and immutable exact-match resume. These primitives do not
themselves prove a harness boundary. A supported driver still needs the native
canaries and pre-dispatch controls above; flags or simulated workers are
insufficient.

## Commands

From the repository root:

```sh
go run ./tools/capability-routing-experiment validate \
  --data tools/capability-routing-experiment/testdata

go run ./tools/capability-routing-experiment prepare \
  --data tools/capability-routing-experiment/testdata \
  --source-commit <frozen-runner-corpus-commit> \
  --out /path/to/new/manifest.json

go run ./tools/capability-routing-experiment prepare \
  --data tools/capability-routing-experiment/testdata \
  --source-commit <frozen-runner-corpus-commit> \
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
  --run-root /path/to/evidence-root
```

`prepare` and evidence writes are create-or-verify: exact immutable resumes are
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

All 288 rows are unavailable, including the second harness. There are no model
task, route, policy, summary, token, latency, or context observations. This is
an inspectable partial result, not the actual comparison required to choose a
portable default. It authorizes no installation, adapter, production change,
or release.
