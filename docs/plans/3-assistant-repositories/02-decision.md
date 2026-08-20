# Solution Decision

## Decision Drivers

1. Preview and durable writes must derive from the same immutable plan.
2. Two-target creation must have explicit interrupted-run recovery.
3. Generated JSON must be checked against versioned, inspectable schemas.
4. The user-facing command should not depend on a language runtime already
   being installed on a fresh Mac.
5. No network, remote, secret, import, global setting, commit, or telemetry is
   permitted at runtime.
6. User content and paths must remain portable and must not be baked into
   generated repository identity unnecessarily.
7. Terminal behavior must be accessible and deterministically testable.
8. The first implementation should create extension seams for O2/O3 without
   implementing either outcome.
9. Dependency and maintenance surface should remain small enough for an
   open-source pilot.

## Competing Approaches

### H1 — Native Go command with embedded contracts (selected)

Create one `my-friday` executable. Pure packages own answers, normalized
profile values, path resolution, plan construction, template rendering,
validation, and transaction state. A thin terminal adapter owns prompts and
line output. Versioned templates and JSON Schemas are embedded in the binary.
Pinned JSON Schema, Unicode-normalization, and segmentation libraries validate
generated documents and user-perceived character limits over shared fixtures.

The repository pins Go 1.26.x in `mise.toml`, `go.mod`, and CI. The delivered
runtime binary has no language-runtime or service dependency; official Go
tooling supports Darwin ARM64 builds. Source builds may download declared Go
modules, but wizard execution performs no network access.

### H2 — POSIX shell wizard and template copier

Use `/bin/sh` or macOS Bash plus `git`, `sed`, and filesystem utilities. This
has no compiled artifact and appears easy to inspect.

### H3 — Python package/command

Use Python for strong string/JSON/filesystem support and distribute through a
package manager or isolated environment. Templates and schemas ship as package
data.

### H4 — Templates plus instructions only

Publish two template directories and ask users to copy, edit, and run `git
init` manually.

## Adversarial Comparison

| Attack | H1: Go | H2: shell | H3: Python | H4: manual |
|---|---|---|---|---|
| Fresh-Mac runtime | Native binary; no installed interpreter. Build/release pipeline is future work. | Shell exists, but macOS Bash is old and utility behavior is platform-specific. | macOS does not provide a stable product-owned Python runtime contract; packaging becomes part of first use. | No runtime, but the user performs every risky operation manually. |
| Two-target recovery | Typed state machine, exclusive files, structured JSON, and fault injection are practical. | Signal handling, path quoting, partial renames, and structured journals become fragile. | Practical, but interpreter/package environment adds a second failure boundary. | No enforceable transaction or recovery. |
| Schema enforcement | Embedded schemas plus one pinned validator keep runtime validation executable. | Requires `jq` or another dependency not guaranteed by macOS. | Good library support, again contingent on environment/package resolution. | Documentation can drift from copied content. |
| Unicode and terminal safety | Standard UTF-8/rune handling; explicit rejection of control/format categories. | Byte/locale behavior and length checks are inconsistent. | Strong Unicode support. | Entirely user-dependent. |
| Zero-network proof | Static import/subprocess allowlist and a fake-Git integration boundary are reviewable. | Any utility/function can invoke more commands; still testable but harder to constrain. | Package/runtime setup may require network even when the wizard does not. | Not enforceable. |
| Maintenance | New Go toolchain and one dependency, but clear package/test seams. | Low file count but high hidden behavioral complexity. | Familiar code, larger distribution/runtime obligation. | Lowest maintainer code, highest user risk; already rejected in discovery. |
| O2/O3 evolution | Stable manifest/profile interfaces can be consumed later without coupling. | Scripts tend to couple prompts, rendering, and mutation. | Good module seams. | No executable ownership contract. |

H1's principal cost is introducing a toolchain and future binary distribution
work. That tradeoff is manageable because this plan authorizes implementation
only. H2's apparent simplicity fails under verified recovery, schema, Unicode,
and path requirements. H3 is technically credible but moves interpreter and
package installation into the first-user contract. H4 contradicts the approved
product decision.

## Selected Approach

Select H1 with Medium-high confidence.

The implementation is a small Go module using standard-library packages for
the terminal, JSON, hashing, embedded templates, path handling, subprocesses,
and file operations. Three reviewed dependencies are allowed: the JSON Schema
validator, `golang.org/x/text` v0.40.0 for NFC, and
`github.com/rivo/uniseg` v0.4.7 for extended grapheme clusters. No CLI
framework, telemetry SDK, database, daemon, network client, template fetch, or
plugin system is introduced.

The core boundary is a canonical `CreationPlan`. Prompt answers are normalized
and validated once, then compiled into that plan. Preview, file rendering,
content hashes, repository validation, transaction support paths, and execution
all consume the same value. Mutation code cannot reconstruct intent from raw
prompts.

Consequences:

- O1 introduces My Friday's first language/toolchain and public command
  contract.
- The implementation PR must classify the project as a buildable artifact in
  durable docs, but artifact publication remains outside this plan's envelope.
- Generated repository contracts start at version 1 and must reject unknown
  major versions.
- macOS-specific APFS preflight remains behind a narrow environment interface,
  but no abstraction for other operating systems is implemented.
- A transaction journal is an exceptional failure artifact, not routine local
  state.
- The domain model and transaction remain independent of Codex, while a small
  Codex projection renders root `AGENTS.md`. This is a seam, not an adapter
  framework; every additional harness requires a future product decision and
  capability mapping.

## Decisions Ledger

| ID | Decision | Rationale | Evidence |
|---|---|---|---|
| D1 | Use a native Go 1.26.x command and embedded templates. | Runtime independence plus typed recovery and testability. | H1-H4 comparison; official Go Darwin ARM64 distribution |
| D2 | Support macOS 14+ ARM64 on local APFS with Git 2.28+. | Narrow first environment; `git init -b`; local rename semantics. | Product decision, Apple APFS docs, Git 2.28 docs |
| D3 | Use JSON Schema 2020-12 with `github.com/santhosh-tekuri/jsonschema/v6` pinned initially at v6.0.2, configured with embedded in-memory resources only. | It supports draft 2020-12 under Apache-2.0; a restricted loader prevents its optional URL features from becoming a runtime network path. | Issue acceptance; <https://github.com/santhosh-tekuri/jsonschema/tree/v6.0.2> |
| D4 | Compile one canonical plan; do not let preview and execution re-derive answers. | Prevent preview/execution drift. | Acceptance group 1 |
| D5 | Derive `assistant_id` from a domain-separated SHA-256 of the exact validated UTF-8 display name plus both initial canonical target paths and retain 128 bits. | Deterministic association distinguishes same-named assistants at different locations. Paths are not stored; the ID remains stable after a later move and is not a credential or registry key. | Deterministic preview, pairing correctness, portability |
| D6 | Derive `plan_id` from canonical JSON containing contract/tool version, normalized answers, canonical targets, ordered actions, and durable-content digests; exclude it from durable generated files. | Same inputs produce the same recovery namespace without a circular file-digest definition or permanent profile hash in repository metadata. | Deterministic preview and privacy requirements |
| D7 | Initialize Git with an empty tool-owned template and branch `main`; create no commits/remotes. | Prevent private template import and reserve history/remote choices for the user. | O1 boundary and Git template risk |
| D8 | Stage and validate both repos before ordered promotion, with reservations and an owner-only journal. | Cross-directory atomicity is impossible; explicit recovery is truthful and testable. | Failure requirement |
| D9 | Keep profile values out of generated Markdown instructions and transaction state. | Prevent structural/prompt injection and avoid a second sensitive-data store. | Trust and privacy requirements |
| D10 | Keep UI line-oriented and ANSI-independent. | Accessibility and deterministic evidence. | Supplied product-experience contract |
| D11 | Use `implementation` execution envelope. | No declared artifact/release contract, naming clearance, or exact-machine evidence yet. | Current `docs/deployment.md` and unknowns U1-U4 |
| D12 | Keep profile/manifests/planner harness-neutral and render only a Codex `AGENTS.md` projection in O1. | Preserve an extension seam without alternate harnesses, adapter machinery, or lowest-common-denominator semantics. | Product-authority guidance; O1/O2 remain Codex-first |
| D13 | Count user text limits with `github.com/rivo/uniseg` v0.4.7 extended grapheme clusters; cap custom guidance at 240. | User-perceived characters treat combining sequences/emoji coherently and preserve the approved profile boundary. | Product-experience contract; <https://github.com/rivo/uniseg/tree/v0.4.7> |
| D14 | Normalize profile text with `golang.org/x/text/unicode/norm` pinned at v0.40.0 in the order trim → NFC → prohibited-category rejection → grapheme count. | Canonically equivalent input must generate identical IDs, plan bytes, previews, and profiles. | Product-experience contract; <https://pkg.go.dev/golang.org/x/text/unicode/norm@v0.40.0> |

## Rejected Mechanisms And Scope

- Cobra/other CLI framework: three commands and one wizard do not justify it.
- Embedded Git library: system Git is an explicit prerequisite and the library
  would broaden dependency and compatibility burden.
- SQLite/database transaction: local directories and JSON state are sufficient.
- `sudo`, installer packages, launch agents, or background recovery: O1 has no
  privileged or persistent service behavior.
- Automatically committing generated files: the user owns authorship and
  history; O5 owns future remote posture.
- Storing absolute paths in repository manifests: breaks portability and leaks
  machine layout.
- Auto-detecting/importing existing assistant settings: violates O1 privacy and
  collision boundaries.
- Supporting Intel and non-macOS targets “because Go can compile them”: build
  capability is not support evidence.
- Generic harness adapter registry or alternate harness template: no approved
  capability mapping or outcome exists yet.
