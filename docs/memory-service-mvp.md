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

Still pending: actual model recall/save behavior, hook execution/context delivery
in live turns, interruption/compaction/fresh-thread scenarios, actual second-machine
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
