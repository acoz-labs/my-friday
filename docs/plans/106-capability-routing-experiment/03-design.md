# Experiment contract

## Surfaces and state

Add `tools/capability-routing-experiment/` with Go command/tests and synthetic
`testdata/`; expose `validate`, `prepare`, `run`, `score`, and `report` subcommands.
These are developer tools, not `my-friday` public CLI additions. Strict version-1
JSON rejects unknown fields, duplicate IDs, unsafe paths and revision mismatch.

Artifacts: corpus (instruction packages and digests), tasks, separate expected
labels, preregistration manifest, per-trial sanitized events, scoring results,
and Markdown/JSON report. Bind each to source commit, corpus/index/rubric hashes,
harness executable version, explicit model/config, mode, trial ID and tool policy.
The manifest defines all cells before execution; never silently drop failures.

State is `prepared -> preflight -> running -> complete|failed|unavailable|invalid`.
Timeout and cancellation are failures, not missing rows. Retry makes a new
linked attempt, preserves the old row, and is excluded from the primary score.
Single-process sequential runs suffice; refuse concurrent reuse of a run root.
Resume only exact manifest matches and never overwrite completed evidence.

## Corpus and preregistration

Use 24 generic instruction capabilities and 24 tasks: 12 development and 12
held-out. Each split contains one task for each category: explicit selection,
informal paraphrase, ambiguous alternatives, no match, dependency loading, stale
index, unsupported required semantics, conflicting instructions, permission
denial, short direct work, complex worker work, and material summary omission.
Specify full task text, permissible capability sets, required result facts,
forbidden effects, and clarification/refusal expectations before any live run.
Use distinct domains/wording in held-out tasks; none reuse the private spike.

Freeze manifest and corpus in a reviewed commit before live results. Development
tasks may guide implementation; held-out tasks and labels must not guide tuning.
Model-visible copies contain only current task, capabilities, essential policy,
and permitted fixture data. Labels, scorer, Git history, implementation source,
other task texts and result rubric remain outside the model filesystem boundary.
The controller scores after model exit. Maintainer audits the staging allowlist
and negative read canaries. A leak invalidates that held-out run; do not relabel
it development and still claim held-out success.

Run all three modes per task with two repetitions in each available harness:
144 trials per harness, 288 for both. Every trial starts a fresh conversation;
first repetition uses a fresh fixture/cache root (cold filesystem), second
reuses only permitted generated fixture/index cache (warm filesystem). No chat,
labels or previous answers survive. Provider prompt-cache state is observed,
not controlled; report cached-token counts separately. Rotate mode order by
task ordinal, reversing it for repetition two; no cross-harness pooling.

## Modes and fidelity

A exposes all 24 skills through the harness's native progressive catalogue,
including normal body-on-demand behavior. B exposes only access instructions;
lookup returns top-three metadata, model confirms applicability and loads the
selected body plus declared dependencies into the root. C uses the identical
lookup policy and delegates only complex or explicitly isolation-required work
to a fresh native worker, otherwise loads directly. Routing prompt specifies
that complex means multi-capability synthesis with intermediate fixture work;
no hidden expected label chooses the worker. At most one worker, depth one,
two capabilities plus one dependency, one broader-metadata fallback, and eight
tool invocations total per trial. Broader fallback exposes at most 24 metadata
entries and its full overhead is counted. A follows the same tool/effect limits.
For A, native workers are available under the same complex-task rule as C, but
selection uses the full native catalogue rather than lookup. Required isolation
must refuse in B because B intentionally does not delegate. This expected refusal
is distinct from failing an achievable task; the rubric records both fidelity
and task completion so modes are not credited with equivalent work on refusals.

## Bounded driver-fidelity checkpoint

Before writing a live runner, implement an offline driver probe and a documented
control matrix for each installed CLI version. The probe reads help/config
schemas and launches only credential-free synthetic process/tool tests. Limit
investigation to these two native harnesses; if no enforceable route is
established, emit `unavailable` and implement the offline comparison/report path.
Inspect at most one installed version/configuration matrix per named harness;
no alternate frameworks, authentication paths or repeated upgrade experiments.
Do not build a sandbox framework to force a pass. This checkpoint resolves live-
driver implementability, not an unbounded implementation assumption or new gate.

The proposed minimal tool channel is a local stdio MCP broker owned by the
controller, exposing only `lookup`, `load`, `read_fixture`, and `write_fixture`.
The broker serves immutable in-memory capabilities and descriptor-bound fixture
files, validates task authority/revision and counts calls BEFORE dispatch. B/C
lookup and selected bodies travel as native tool results, not injected answers.
No broker call accepts arbitrary filesystem paths or executable strings. The
controller generates native configuration pointing at this exact broker, with
other MCP servers and model shell/network tools disabled. Root and worker broker
calls share one task-scoped counter and policy. Confirm native workers inherit
this exact tool surface and cannot add tools; otherwise C is unavailable.

A requires native skill directories and native body reads. Record exactly which
native reader reads bodies, how it is constrained to the staged allowlist and
how catalogue visibility is observed. If disabling broad filesystem tools also
disables native skill loading, do not replace A with broker-fed instructions.
Codex JSON events and Claude structured noninteractive output are transports,
not proof of tool confinement. Claude safe mode is ineligible for A/C. Worker
launch/return identities must come from native events, not a controller-created
subprocess. Satisfy these requirements before marking a live driver supported.

An OS-backed reader/tool boundary is required in addition to broker allowlists.
Name the installed native sandbox configuration and demonstrate read/network
denial canaries for every enabled built-in tool, including native skill reads
and workers, without exposing host auth to those tools. Until verified, the
driver stays unavailable. Default workspace-write is not restricted-read proof;
acceptance-runner isolation has a distinct process model. No credential
relocation or new OS account is an implicit workaround.

Essential root instructions carry task intent, read/change authority, deny-by-
default external effects and required reporting. Workers receive the same
policy, exact task scope and revision-bound selected instructions. Their result
contract contains outcome, capability/revision, changes, verification, failures,
limitations and decisions needed. Scorer checks root answer AND recorded actual
fixture effects; a polished summary cannot hide a failed operation.

Preflight each harness/mode with synthetic canaries proving the expected skill
set, on-demand body visibility, allowed tools, blocked ambient files/services,
and actual native worker creation where needed. A disabled-skills safe mode is
not a native baseline. A CLI child pretending to be a native worker is not C.
Record native child identity/events when supported; otherwise mark C unavailable.
If native catalogue truncates, record the observed visibility and retain the
baseline result; do not manually inject omitted skills to improve it. No ambient
private catalogue, plugins, MCP service, shell credentials or personal rules may
enter a run. Flags alone are insufficient proof of these boundaries.

## Authorization, isolation and failure

Default commands are offline and credential-free. Live `run` requires explicit
`--live` and a manifest that names an already authorized harness, model and
budget. Harness authentication is opaque and externally provisioned; the tool
never discovers, copies, logs, provisions or edits credentials. A preflight
must prove an OS-enforced fixture-only model tool boundary with label/ambient
read-denial canaries, while the harness host retains only its necessary existing
authentication and inference connection. Reuse tested repository sandbox/process
primitives where suitable; unsupported isolation means unavailable, not an
unsandboxed fallback. No new VM manager or cross-platform sandbox framework.

Permit only local synthetic read/write tools within the disposable fixture root;
deny network from model tools, arbitrary external commands and user directories.
Inference transport is distinct from model tool network authority. Stale index
refuses before body loading; dependency cycle, missing dependency, budget excess,
or unsupported required semantics returns a structured refusal. Retrieval does
not enlarge task scope. The controller enforces effects and limits independently
of instructions; adversarial requests test model refusal as well as prevention.

Per trial: 120 seconds wall time, 30,000 observed aggregate input+output tokens,
8 tool invocations, one worker, and a 2,000-token requested output ceiling per
agent where the harness supports enforcement. Tokens are a stop threshold at
event boundaries, not a guaranteed billing cap. Live batch: at most 12 trials,
30 minutes, and 360,000 observed total tokens, including workers. Stop on the
first critical policy loss, boundary violation, malformed telemetry stream or
unreaped process. No automatic retries. A provider-enforced monetary cap is
required before any metered API route with incremental charges; absent that,
that route is unavailable. This plan makes no paid-service commitment. Existing
subscription access still requires operator opt-in and the time/call limits.

The broker independently enforces its eight-call ceiling before executing an
operation. Native built-in calls outside the broker require a native pre-dispatch
limit/hook with equivalent rejection; otherwise call ceilings are only observed
abort thresholds and the cell is ineligible for supported live execution. Worker
count/depth similarly requires a native pre-launch maximum of one/depth one, or
worker cells are unavailable. Output ceilings lacking an enforceable setting
are requested limits only; report their status and rely on the hard wall-clock
supervisor, not fabricated token prevention. Usage stopping may overshoot by an
in-flight response.

Supervisor kills and reaps process groups on timeout/cancel, preserving failure
receipts. Ownership must bind the exact spawned child and its descendants;
the existing acceptance runner's same-executable discovery and blanket descendant
failure policy is not a safe drop-in for native workers or concurrent host tasks.
Tests prove unrelated harness processes survive. Cleanup removes only
manifest-owned disposable roots after proving
their identity; ambiguous ownership is preserved and reported. Raw logs stay
owner-only locally, excluded from Git. Export only allowlisted synthetic facts
and metrics after sanitization; never raw transcript text, environment, auth
paths, session identifiers or private directory names. Operator explicitly
discards local logs after reviewing exports; nothing is uploaded automatically.

## Measurement

Record wall time, selected capabilities, result facts, attempted/actual effects,
worker/handoff events, and per-agent usage where genuinely available. Report
root cumulative input, worker cumulative input, total input/output, cache usage,
and peak per-request root input as DISTINCT quantities. Peak per-request input
is still not actual window occupancy unless the harness proves that semantic.
Deduplicate aggregate/child usage by event identity; do not add an aggregate to
its components. Record provenance and completeness for every metric; missing
values are null with reason. Character/byte sizes describe supplied fixtures,
never token measurements. Missing worker accounting invalidates total-cost
comparisons; missing root per-request context evidence invalidates context-
reduction claims. Partial correctness evidence may still be reported.
