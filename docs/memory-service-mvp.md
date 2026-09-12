# My Friday memory service MVP

Owner direction, 2026-09-12: bring your agent; My Friday supplies durable memory.
This replaces assistant/capability ownership as the product's core. Existing
assistant installations remain intact while the memory-only path is proven.

## Required outcome

An ordinary Codex installation connects to an explicitly selected memory bank,
recalls relevant knowledge, journals meaningful outcomes and writes traceable
decisions/corrections. Another installation on another machine uses the same
bank without copying threads, native credentials, capabilities or agent homes.
The host agent performs semantic extraction using its existing model access;
the memory service does not require a separate inference subscription.

The shared engine owns durable records, journals, provenance, scoped retrieval,
supersession/conflicts and local-first Git synchronization. The CLI and MCP server
use the same engine. Native plugins supply only discovery, connection and memory
lifecycle integration; native tools, skills and agent behavior remain native.

MVP deployment is local stdio MCP with an explicit machine-local binding to a
Git-backed bank. Remote hosting is not required. Supported host targets must be
verified and documented; transport compatibility is not proof of native behavior.
Pi follows fully accepted Codex support (#117); Claude Code follows Pi (#118). Their
implementation is excluded from this MVP goal.

## Repository boundaries

- `internal/portable`: existing immutable memory/sync primitives and legacy
  assistant compatibility; extend bank-only persistence without loading legacy
  capabilities, instructions or subscriptions.
- `internal/memorybank`: harness-neutral memory operations and bounded retrieval.
- `internal/memorymcp`: MCP transport, delegating to the memory engine.
- `internal/memorycodex`: compiled thin native event adapter, with read-only
  bounded context delivery and no transcript access.
- `plugins/codex`: distributable Codex plugin, connection, skill and hook wiring.
- Future `plugins/pi` and `plugins/claude-code`: native adapters over that same
  engine, not new memory implementations. Follow-up issues define their work.

No provider-specific capability or local credential manager is introduced.
Personal/work banks are explicitly connected, never silently combined. Device
and harness provenance describes the originating client, not just a server host.

## Acceptance ledger

Completion requires evidence for every item, not just a working MCP handshake:

- Memory-only create/import, machine enrollment and inspection without assistant
  instructions, a harness home or capability manifests.
- Durable journals and structured knowledge, scoped search/recall, history,
  corrections and explicit concurrent-conflict reporting.
- Same shared semantics through CLI and MCP, with validated typed inputs, bounded
  output, predictable errors and safe retry/concurrent-write behavior.
- Git checkpoint/pull/push, offline persistence and visible pending/conflict state;
  real fresh installations preserve original provenance and receive corrections.
- Native Codex plugin installation/discovery, memory tools and automatic recall
  integration without replacing native configuration or requiring an agent alias.
- Meaningful journal/decision persistence during work, fresh-thread continuity,
  and documented interruption/compaction limits. No end-hook-only durability.
- Relevant current memory arrives with bounded context; old proposals do not
  override current user direction. Empty search is not proof of absent history.
- Synthetic end-to-end create/recall/correct/history/isolation acceptance using
  actual Codex plus a second machine; preserve unrelated host/project resources.
- User-facing setup, update and troubleshooting documentation, reproducible
  validation, clean retained artifact and explicit installation handoff.
- Pi and Claude Code follow-up issues tracked in the My Friday GitHub project.

Journals retain semantic summaries, not lossless raw conversations. No automatic
transcript upload or secret retention. Integration must honor task-specific
read-only/no-memory-write directions; remembered text is evidence, not authority.

## Current evidence

- Existing memory primitives inspected: immutable revisions, graph supersession,
  device provenance, lexical recall and Git sync can be reused.
- Codex CLI installed for acceptance: 0.153.4. Current public documentation may
  describe newer behavior; verify the installed native contract before relying on it.
- GitHub repository access works. Project listing currently fails for missing
  `read:project` scope; board attachment needs authorized project access or an
  owner action. Do not expand credentials silently.

Implemented and exercised on macOS/ARM64, 2026-09-12:

- Memory-only bank creation, independently bound clones, journals, scoped recall,
  correction history, original device provenance, and concurrent semantic heads.
- Compact byte-budgeted recall, bounded history/journal/scope pages, invalid
  writes rejected before evidence mutation, and replacement-bank detection.
- `bank` CLI and seven typed MCP tools over the same service; strict input
  validation and no journal writes from reads. Official MCP Go SDK pinned at 1.7.0.
- Codex native plugin install in disposable state, namespaced skill and hook
  discovery, connected MCP server and successful native MCP recall. No model
  turn, authentication copy, or ambient configuration mutation in this test.
- The native test exposed missing binding environment forwarding; the plugin
  now explicitly forwards it. Native discovery names the skill
  `my-friday-memory:memory`, not simply `memory`.
- `mise exec -- bin/ci` passed (including race tests and native APFS primitive).
  Docker daemon was unavailable, so the documented pinned host fallback was used.
  The final automated pass after the identity/scope changes passed as well.
- Linux AMD64 and ARM64 static builds passed. Linux runtime acceptance remains
  unverified; a build is not proof of platform behavior.
- Plugin and skill validators pass using an isolated `uv --with pyyaml` environment.

Retrieval scale investigation after the first candidate:

- Added `BenchmarkRecallCorpus` with validated on-disk synthetic corpora of 100,
  1,000 and 5,000 records, approximately 540 bytes of body text per record. It
  measures a broad matching query returning five hits within an 8 KiB budget.
- A single local measurement on Apple M1 Pro showed the 5,000-record recall
  taking 3.85 seconds and allocating 872 MB. Relevance was repeatedly recomputed
  inside the sorting comparator.
- Computing each candidate score once per fresh read reduced that measurement
  to 0.65 seconds and 182 MB allocated. The 1,000-record measurement fell from
  0.64 to 0.13 seconds. These are synthetic benchmark samples, not latency SLOs,
  peak retained memory, cold-disk measurements, or full hook timings.
- Ordering/tie regression coverage preserves the existing lexical results,
  while supersession and scope tests retain their semantics. No persisted index
  or cross-call cache was added; later disk changes still require fresh reads.
- Focused `go vet` and race-enabled tests passed for `internal/portable`,
  `internal/memorybank`, `internal/memorymcp` and `internal/memorycodex`.
- The owner-assisted trial remains pinned to its original candidate. Do not
  replace its executable while the first conversation test is pending.

Memory-first front door after the initial candidate:

- Empty argv and `menu` now open memory-bank create/connect/check/sync, using
  the existing arrow/Vim UI and numbered fallback. Machine-local binding paths
  are explicit; no agent identity, capability setup or native configuration is
  created. Existing assistant management remains available through `menu --legacy`.
- The UI uses the same memory storage, binding and sync operations as explicit
  commands. It does not automate native plugin installation or authentication,
  and it does not change the already-handed-off pilot binary.
- Tests cover the new default route, explicit legacy route, create/connect,
  health/sync, cancellation, connection collisions and an unresolved symlink
  hiding an in-bank binding. Full host CI and Linux AMD64/ARM64 builds passed.
- At the follow-up check, no process matching the handed-off native pilot was
  running and its scope inventory remained empty. GitHub Project listing still
  failed for missing `read:project` scope. Continue native acceptance with owner
  input; do not treat these unperformed checks as successful acceptance.

First owner-assisted hook trial exposed a runtime-selection bug:

- Native shell initialization reordered PATH and selected an older installed
  assistant-platform executable. `codex-memory-hook` was unknown; exit code 2
  blocked the prompt before a model response. Discovery/MCP-only testing had
  not covered this path. This is failed acceptance, not a memory-save result.
- The plugin now uses a small POSIX runner and optional absolute
  `MY_FRIDAY_MEMORY_BIN` for both hooks and MCP, independent of shell PATH order.
  Failed hook execution produces a nonblocking warning without failed raw output.
  It does not silently fall back from a missing explicit runtime to another one.
- Regression tests first reproduced the failure, then passed for both lifecycle
  hooks, old/missing runtimes, relative overrides, paths with spaces and MCP stdio.
  Native Codex 0.153.4 install/discovery and MCP recall passed with the pinned
  executable deliberately absent from PATH. Live trusted-hook retest is pending.
- Direct execution through the same host login shell returned valid recall
  context with the explicit runtime. Full host `bin/ci` and plugin validation
  passed after the repair; no live model-save success is implied.
- The trial's compiled executable remains unchanged. Only its plugin and local
  test wrapper need updating, at a fresh-session boundary; Alfred stays intact.

First successful owner-assisted conversation after the hook repair:

- The retained synthetic native transcript contains the prompt-hook orientation,
  empty initial recall packet and scope inventory, confirming context delivery.
- The model read the installed memory skill, recalled existing records, saved
  Copper Finch as a fact and concise answers as a preference through MCP, then
  appended one short setup journal and called sync. No tool failures or retries
  appeared in this turn; no capability-building workflow was invoked.
- Independent read-only CLI inspection confirmed exactly two current records
  and one journal, with the bound device and Codex provenance. Sync returned
  `local-only`, and the final answer accurately stated that cross-machine sync
  was not configured. This proves the basic save path, not remote delivery.
- The native turn took about 38 seconds, including a roughly 14.5-second batch
  containing writes and checkpoint. This sample is not a performance target;
  unnecessary reads and checkpoint overhead remain worth observing.

Fresh-thread read-only acceptance passed:

- A distinct native session received both saved facts through the prompt hook,
  then read the memory skill and made one read-only MCP recall. Its answer named
  Copper Finch and the concise-answer preference correctly without being given
  either answer in the new prompt.
- No remember, journal-append or sync call occurred. Independent inspection
  found the same two records, one journal, unchanged Git HEAD and clean bank.
- The additional MCP recall repeated evidence already present in hook context;
  harmless here, but a concrete efficiency opportunity to revisit after core
  behavioral acceptance. No plugin/runtime change was made for this test.

Owner-assisted rename/supersession acceptance passed:

- The model updated the existing project record, explicitly superseding its
  Copper Finch revision with Silver Heron. Independent history inspection
  confirmed both immutable revisions, the supersession edge, user-directed
  reason and original machine/harness provenance.
- Current recall contains Silver Heron and the unchanged original preference
  revision, with no duplicates or conflicts. One rename journal was added and
  the bank was checkpointed locally with a clean working tree. Fresh-session
  recall of the corrected name is the next behavioral check.

Fresh-thread corrected recall passed:

- A distinct native session received only the current Silver Heron revision
  for its name query, including Copper Finch as historical context. It answered
  correctly without the prompt supplying either name.
- The model read the skill and performed one MCP recall; no write or sync tools
  were called. Git HEAD and the clean bank were unchanged; the same two journals
  remained. Automatic context delivery and subsequent MCP recall agreed.

Ordinary-task learning did not pass on its first trial:

- In the same session immediately after a read-only recall request, the owner
  stated an offline/local-storage prototype decision and excluded cloud sync,
  asking for a two-item checklist without explicitly requesting a memory save.
- Hook context arrived, but the model answered without any tool calls. No
  decision or journal was saved; the bank remained at the rename checkpoint.
- The earlier read-only request is a possible confound, not an established
  cause. Repeat the same ordinary-task prompt in a fresh session before changing
  learning instructions. Do not insert the missing decision on the model's
  behalf or count its correct checklist as durable-memory acceptance.

Fresh-session ordinary-task control passed:

- Repeating the same prototype/checklist prompt in a distinct session saved one
  user-directed decision for offline operation, local storage and excluded cloud
  sync, without an explicit request to remember. One semantic journal stated
  that no implementation had occurred, and sync checkpointed the local bank.
- Suggested CRUD tests, asset bundling and other checklist details were not
  promoted to user-approved requirements. Existing records remained unchanged.
- This supports, but does not prove, prior read-only scope carryover as the
  explanation for the first miss. Test an explicitly one-answer no-save request
  followed by another ordinary decision in the same session before resolving it.

Explicit one-reply no-save boundary passed:

- The hypothetical cloud-sync question produced no tool calls or memory writes;
  its answer retained the actual offline-only decision. The following macOS/Linux
  target decision was automatically saved and journaled in the same session.
- The offline decision, project identity and answer preference remained unchanged.
  The new confirmed record contains only the user's platform choice; the agent's
  storage-interface suggestion appears in historical journal context, not as an
  approved decision. No cloud-sync requirement was introduced by the hypothetical.
- This proves the explicitly scoped boundary in this sample. The earlier bare
  read-only carryover ambiguity remains documented rather than reclassified as
  a proven implementation defect or universal model behavior.

Current-user scope change acceptance passed:

- The model accepted optional cloud backup without an approval loop or objection
  based on the earlier exclusion. It revised the existing scope record with an
  explicit supersession edge and reason, preserving the original revision.
- Current knowledge retains local data, full offline operation when backup is
  disabled, and unchanged macOS/Linux targets. It distinguishes optional backup
  from general cloud synchronization rather than expanding the user's request.
- A semantic journal describes the actual change; no implementation was claimed.
  No duplicate or conflicting current decision was introduced.

Different-working-directory native acceptance passed:

- A fresh native session launched in an empty second project directory. Its
  `pwd` and final answer retained that directory while both prompt-hook context
  and MCP recall supplied Silver Heron's current prototype scope from the bound
  bank. No assistant-repository cwd or project-local instructions were needed.
- It retrieved macOS/Linux targets, local storage, offline operation and optional
  backup without reverting to the earlier exclusion. Only read operations were
  invoked; the second directory remained empty and bank HEAD/tree were unchanged.
- A separate empty synthetic bank/binding is prepared for the next native
  isolation test, using the same plugin and native profile in a fresh session.

Separate-bank native selection passed:

- A fresh session using the same native profile/plugin but a separate empty
  binding received the second bank ID in hook context. MCP recall and scope
  inventory were empty, and the model did not invent a project or preference.
- It did not inspect other banks or native transcripts and made no write/sync
  calls. The second bank remained empty; the original bank's clean tree and
  checkpoint were unchanged. This is evidence for explicit bank routing and
  observed task compliance, not filesystem isolation from an unrestricted agent.

Still pending: broader read-only/no-save scenarios,
interruption/compaction scenarios, actual second-machine
acceptance, final user-facing management/distribution experience, and GitHub board
attachment. Current clone tests simulate two machines; they are not physical-host
acceptance. Writes are not idempotent: clients must inspect after ambiguous failure
before retrying. The generated development marketplace is named `personal`; a
collision-safe public distribution name is not finalized.

An isolated native Codex pilot and blank bank are prepared for owner-assisted
conversation acceptance; no ambient authentication was copied. The previously
used second Mac currently refuses SSH with `Host key verification failed`.
Do not bypass host verification to continue that test; establish its identity
with the owner first.

The MVP is not complete. No production release or private-memory migration has
been performed.
