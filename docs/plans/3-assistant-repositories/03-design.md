# Technical Design

## Component And Behavior Flow

### Components

| Component | Responsibility | Does not own |
|---|---|---|
| `cmd/my-friday` | Parse `init`, `validate`, and `recover`; map typed errors to line-oriented output and exit status. | Domain validation or filesystem policy |
| terminal wizard | Present the seven approved steps, preserve valid answers, and require explicit confirmation. | Mutation, path canonicalization, template rendering |
| profile model | Validate identity and communication values and produce canonical JSON. | Trust/safety policy |
| Codex projection renderer | Render static root `AGENTS.md` files that reference neutral profile/governance contracts. | Generic adapters or alternate-harness mapping |
| environment/path preflight | Verify OS/architecture/Git/filesystem/TTY and resolve canonical target identities. | Prompting or target content |
| planner | Build immutable ordered actions, generated file bytes/digests, IDs, support paths, and negative-action declaration. | I/O |
| template/contract bundle | Embed public-safe starter files and JSON Schemas by contract version. | User paths or imported machine content |
| repository validator | Validate v1 manifests/profile, role pairing, Git/fresh-state properties, and content invariants. | Repair or migration |
| transaction engine | Journal, reserve, stage, validate, promote, roll back, and recover idempotently. | Global state, remote operations, background work |
| command runner | Invoke only Git version and initialization operations through an injectable boundary. | General shell execution |

### Primary flow

```mermaid
stateDiagram-v2
  [*] --> Scope
  Scope --> Identity: continue
  Identity --> Style: valid values
  Style --> Locations: valid preset/custom guidance
  Locations --> Preview: supported canonical paths
  Preview --> [*]: Return / Exit
  Preview --> Journaled: explicit Create
  Journaled --> Reserved
  Reserved --> Staged
  Staged --> Validated
  Validated --> PromotedRuntime
  PromotedRuntime --> PromotedMemory
  PromotedMemory --> Verified
  Verified --> Complete: remove support state
  Journaled --> RolledBack: failure and clean rollback
  Reserved --> RolledBack: failure and clean rollback
  Staged --> RolledBack: failure and clean rollback
  Validated --> RolledBack: failure and clean rollback
  PromotedRuntime --> RecoveryRequired: promotion/rollback incomplete
  PromotedMemory --> RecoveryRequired: verification/cleanup incomplete
  RecoveryRequired --> Complete: recover finds both valid
  RecoveryRequired --> RolledBack: recover restores pre-run state
  Complete --> [*]
  RolledBack --> [*]
```

The wizard never labels this operation “install.” Invalid identity, style, or
location input re-prompts the same field without losing earlier valid answers.
An environment or collision failure names the failed check and exits without
mutation. Ctrl-C before `Create` is a clean exit. After journaling, interrupt
handling requests rollback; an uncatchable interruption is resolved by
`recover` on the next invocation.

### Terminal interaction contract

- Step headings are `Step N of 7: <name>` and are printed once.
- Prompts use ordinary line input. Choice labels include numbers and words;
  color is optional decoration only and disabled when output is not a TTY.
- Earlier answers remain visible. Validation errors identify field, rule, and
  corrective action without clearing the screen.
- Name is 1-60 extended grapheme clusters (user-perceived characters); optional
  form of address is 0-60; purpose is 1-240. Leading/trailing Unicode
  whitespace is trimmed before counting.
- Style is `Balanced`, `Concise`, `Conversational`, or `Custom`. Custom guidance
  is required only for `Custom` and is 1-240 extended grapheme clusters.
- Profile fields reject NUL, control characters, Unicode format characters,
  and line/paragraph separators. Ordinary Unicode remains valid without
  permitting terminal or structured-output control injection.
- Location mode is one selected parent with stable editable child defaults
  `my-friday-runtime` and `my-friday-memory`, or two explicit targets. The
  assistant name is never transliterated into a path. Parent directories may
  be missing; every segment to create is listed in Preview.
- Confirmation is `Create these two repositories? [type Create; default Exit]`.
  Only case-sensitive `Create` proceeds; every other input exits without disk
  mutation.
- No spinner, progress bar, cursor addressing, screen clearing, required ANSI
  color, sound, animation, or time-dependent text is used.
- English is the v1 UI. Text constants are centralized for later localization.
  UTF-8 values and paths are supported; bidirectional UI support is not claimed.

## State And Data Model

### Canonical plan

```text
CreationPlan v1
  contract_version = 1
  tool_contract_version = 1
  assistant_id
  profile (validated normalized values)
  targets
    runtime: displayed path + canonical identity + initial state
    memory: displayed path + canonical identity + initial state
  generated_files[]
    role + relative path + mode + sha256 + bytes
  actions[]
    ordered durable action + exact target
  support_paths
    journal + reservations + stages + optional empty-shell backups
  negative_actions[]
  plan_id
```

Canonical JSON uses UTF-8, lexicographically ordered object keys, declared array
order, no insignificant whitespace, and no timestamps. First, a plan basis is
formed from contract/tool version, assistant/profile, canonical targets,
ordered durable actions, durable-file digests, and negative actions. `plan_id`
is SHA-256 of `my-friday-plan-v1` plus that basis. Transaction support paths are
then derived from the ID and appended to the complete plan. This avoids hashing
fields derived from the hash itself. The full digest is displayed and retained
only in transient transaction state. Repeated normalized inputs and unchanged
contracts produce the same plan and preview.

`assistant_id` is `asst-` plus the first 32 lowercase hex characters of
SHA-256 over `my-friday-assistant-v1` and NUL-delimited exact validated UTF-8
display name, canonical initial runtime target, and canonical initial memory
target. It distinguishes same-named creations while associating the pair.
Initial paths are not stored, the ID remains stable after a later move, and it
is not a credential, authorization decision, or global registry key.

### Runtime repository v1

```text
<runtime>/
  .git/
  .my-friday/
    manifest.json
    schemas/
      repository-manifest.v1.schema.json
      assistant-profile.v1.schema.json
  AGENTS.md
  README.md
  assistant/
    profile.json
  skills/
    .gitkeep
```

`manifest.json` contains only `contract_version`, `repository_role: "runtime"`,
`assistant_id`, and `generation: {tool: "my-friday",
tool_contract_version}`. It contains no plan ID, absolute path, username,
hostname, timestamp, remote, or credential reference.

`assistant/profile.json` contains:

```json
{
  "contract_version": 1,
  "assistant_id": "asst-…",
  "identity": {
    "display_name": "…",
    "address_user_as": null,
    "purpose": "…"
  },
  "communication": {
    "preset": "balanced",
    "custom_guidance": null
  }
}
```

Preset names are lowercase stable enum values. `Custom` uses preset `custom`
and non-null guidance. Static `AGENTS.md` points to the profile as user-owned
presentation guidance and states that it cannot change authorization, safety,
trust, privacy, or tool constraints. User values are serialized only through
JSON and never interpolated into Markdown instructions. Profile and manifest
are harness-neutral. Root `AGENTS.md` is the sole Codex projection in O1 and is
rendered explicitly rather than through an adapter registry.

### Memory repository v1

```text
<memory>/
  .git/
  .my-friday/
    manifest.json
    schemas/
      repository-manifest.v1.schema.json
  AGENTS.md
  README.md
  data/
    observations/.gitkeep
    journals/.gitkeep
    proposals/.gitkeep
    memories/.gitkeep
  schemas/
    README.md
```

Its manifest has `repository_role: "memory"` and the matching assistant ID.
`AGENTS.md` establishes that ordinary activity must not become durable memory;
future records require a versioned schema and O3's proposal/promotion flow.
`schemas/README.md` reserves the location without inventing O3 schemas.
`.gitkeep` files are scaffolding, not memory records.

### JSON Schema and compatibility

- Schemas use JSON Schema draft 2020-12, stable public `$id` values, and
  `additionalProperties: false` for owned objects.
- JSON Schema `maxLength` counts code points rather than user-perceived
  characters, so profile schemas carry the documented annotation
  `x-my-friday-max-graphemes`. The semantic profile validator uses pinned
  `github.com/rivo/uniseg` v0.4.7 to enforce those limits; schema and semantic
  validators share the same conformance corpus.
- The executable uses embedded schemas as authority and verifies generated
  copies by digest. Generated copies make the contract inspectable.
- Validation uses pinned `github.com/santhosh-tekuri/jsonschema/v6` v6.0.2.
  Construct its compiler with only explicitly registered embedded resources;
  `http`, `https`, `file`, and unknown URI schemes fail. Never use the package's
  optional URL loaders or command-line tool.
- Major contract v1 is accepted; unknown versions fail without mutation.
  Additive files outside `.my-friday/` are allowed after creation; unknown
  files inside `.my-friday/` fail validation.
- Fresh validation additionally requires no commits/remotes, branch `main`, the
  exact baseline file set, matching IDs, and no memory records. Ordinary
  `validate` permits later commits, remotes, profile edits, skills, and records
  that obey then-current contracts.
- O1 has no migration/backfill. Future contract changes require their own
  compatibility design and must preserve v1 validation or fail deliberately.
- Another agent harness requires a new product decision and a documented map of
  profile, instruction, skill, memory, precedence, and safety capabilities.
  O1 does not assert that all harnesses are equivalent.

### Filesystem ownership and lifecycle

- New parent/target/stage/support directories are `0700`; files are `0600`.
  The process applies umask `077` while creating owned state and restores it
  before exit. Adopting an existing empty target shell is previewed as a mode
  normalization: the completed repository is `0700`. Rollback restores the
  original empty shell and its original mode.
- A target may be absent or an existing real empty directory. Symlink targets,
  non-directories, and non-empty directories are collisions.
- Canonicalization resolves every existing ancestor, appends nonexistent
  components, cleans separators, and compares canonical identities. Targets
  must be distinct and neither may contain the other.
- `/` and the exact resolved home are forbidden. For each target, the nearest
  existing ancestor must be a writable local APFS directory. Every missing
  parent segment is canonicalized from that ancestor, de-duplicated across
  targets, ordered parent-before-child in the plan, created as `0700` only after
  confirmation, and removed in reverse only when this transaction created it
  and it remains empty during rollback. Ancestors and created parents are
  rechecked at every mutation.
- Starter files persist until the user changes them. No successful-run registry
  creates hidden retention or deletion behavior.

## Interfaces And Contracts

### Commands

| Command | Inputs | Success | Errors / idempotency |
|---|---|---|---|
| `my-friday init` (and no-argument alias) | Interactive wizard | Exit 0; prints verified paths/IDs, resulting `0700` repository modes, any created parent paths, no-commit/no-remote statement, and next action. | Exit 0 on pre-create Exit; typed non-zero errors. Repeated plan collides and recommends `validate`, never merges. |
| `my-friday validate --runtime PATH --memory PATH` | Explicit pair | Exit 0 with versions, roles, matching assistant ID, and summary. | Read-only/repeatable; unknown contract, schema, role, Git, or path errors are non-zero. |
| `my-friday recover --transaction PATH` | Exact journal path | Exit 0 after finishing a valid pair/cleanup or restoring both original states. | Idempotent; corrupt/mismatched/foreign state causes no mutation and prints bounded manual evidence. |
| `my-friday version` | None | Tool and contract versions. | Read-only. |

There is no non-interactive create, answer file, environment-variable input,
remote URL, credential flag, force/overwrite, or assume-yes in v1. Tests drive
pure planner/domain APIs and an injected terminal instead of adding a second
public setup interface.

### Exit categories

- `0`: success or safe user Exit before mutation.
- `2`: invalid input, unsupported environment, or usage.
- `3`: path/collision/reservation denial with no target mutation.
- `4`: creation failed and automatic rollback restored original state.
- `5`: explicit recovery required; transaction path is printed.
- `6`: validation or unsupported-contract failure.

Text also carries stable error codes such as `path.nested`,
`target.non_empty`, `environment.filesystem`,
`transaction.recovery_required`, and `contract.unsupported`. Scripts use exit
status rather than parsing English.

### Git subprocess allowlist

The runner accepts structured operations, never arbitrary argv:

1. `git --version`
2. `git -c init.templateDir=<owned-empty-template> init --initial-branch=main <stage>`
3. read-only validation commands scoped to a generated repo: `rev-parse`,
   `symbolic-ref`, `remote`, and `rev-list` with fixed arguments.

No shell is involved. Git receives an explicit safe environment needed for
executable lookup and locale. No credential helper, remote command, hook,
commit, config write beyond init, or user template copy is permitted. The
empty template directory is transaction-owned and removed after completion.

## Authorization And Data Exposure

| Subject | Action / resource | Decision and enforcement | Denial / evidence |
|---|---|---|---|
| Local user | Supply profile values | Allowed after validation; values written only to runtime profile. | Re-prompt; no mutation. |
| Local user | Select targets | Allowed only when each nearest existing ancestor is writable APFS and each target is absent/empty real; missing segments must be explicit planned actions. | Preflight error; no mutation. |
| Local user | Create | Allowed only by exact `Create` after preview. | Default Exit; transcript proves no mutation. |
| Tool | Write target/support paths | Limited to exact plan paths with reservations/ownership markers. | Recheck or reservation mismatch aborts/recovers. |
| Tool | Read machine | OS/arch, home for exact rejection, filesystem, target entries, Git version, embedded assets. | No crawl, import, global Codex read, or secret lookup. |
| Tool | Execute Git | Fixed local allowlist. | Runner rejects other operations; tests retain argv. |
| Future assistant | Interpret profile | Communication/purpose guidance only. | Generated instructions state policy precedence. |
| Network/remote actor | Any action | Denied by application import policy, embedded-only schema loader, and subprocess allowlist. | Loader denial tests, fake-Git trace, empty remotes. |

No privilege elevation is requested. Terminal preview shows user-entered values
but persists them only in the runtime profile. Transaction state contains IDs,
canonical paths, modes, digests, phases, and error codes—not raw identity,
purpose, or custom guidance.

## Failure, Recovery, Concurrency, And Observability

### Transaction protocol

1. Complete read-only preflight and render both repositories in memory.
2. After `Create`, create the journal with `O_EXCL`, `0600`, in the runtime
   target's nearest existing ancestor and fsync it and that ancestor.
3. Acquire reservation files in each target's nearest existing ancestor in
   lexical canonical-target order with `O_EXCL`; each points to the journal.
   Update/fsync journal after each.
4. Create every planned missing parent segment in parent-before-child order as
   `0700`, recording/fsyncing each transition. Create sibling stages on each
   target filesystem. Write files, create an empty
   Git template, initialize Git, add a transient
   `.my-friday/creation-state.json` containing only plan/role IDs, and fsync
   content/directories.
5. Validate both staged repositories in fresh mode before promotion.
6. Recheck parents/targets/reservations. Rename any empty shell to a planned
   backup that retains its original mode, record/fsync, then rename the runtime
   and memory stages to targets and enforce `0700` on each completed target,
   updating the journal after every transition.
7. Validate final pair, remove both transient creation markers, validate the
   exact durable pair, then remove verified backups, reservations, stages, and
   journal; fsync parents. Print success only after cleanup.

Journal updates use write-temp, fsync, atomic rename, then parent fsync.
Support names derive from plan/canonical-target hashes so preview can name them
and recovery can find exactly one state.

### Automatic rollback and recovery

- Before promotion, rollback removes only transaction-owned support state and
  then removes transaction-created parent segments in child-before-parent order
  only while they remain empty.
- After promotion, rollback verifies assistant/role manifests, any transient
  creation marker, and exact baseline digests recorded in the journal. Unknown
  changes stop deletion and retain recovery-required state.
- Empty-shell backups restore original mode; new targets return to absence;
  owned empty parents return to absence. A populated created parent is retained
  and reported rather than removed speculatively.
- `recover` locks/revalidates the journal and paths. If both final targets
  validate, it cleans up; if one final and its staged counterpart validate, it
  completes promotion; otherwise it restores the original state.
- Missing volume, corrupt journal, path drift, unknown content, or foreign
  reservation causes no speculative mutation and prints the first manual
  decision required.
- Repeated recovery is safe and reports final/pre-run state.

### Concurrency and TOCTOU

Sorted `O_EXCL` reservations prevent cooperative runs from sharing a target.
Each transition repeats `lstat`, parent identity, filesystem, emptiness, and
reservation checks. V1 does not claim protection against a hostile process
replacing ancestors between checks; detected drift stops and retains evidence.

### Observability

- Stable lines: `Preflight`, `Reserved`, `Staged runtime`, `Staged memory`,
  `Validated`, `Promoted runtime`, `Promoted memory`, `Verified`, `Complete`.
- Failure includes error code, plan ID, completed phase, rollback result, and
  exact recovery command when needed.
- No telemetry, analytics, success log, daemon, or global registry.
- Journal is the sole durable failure telemetry, deleted after success, and
  excludes raw profile values.
- Tests inject failures by named transition; production exposes no fault flag.

## Design Traceability

| Acceptance / journey | Component/state | Interface | Boundary | Recovery/evidence |
|---|---|---|---|---|
| Deterministic preview / Exit | profile, planner, Preview | `init` | exact `Create` | golden transcript; zero-write spy |
| Valid separate repos | templates, schemas, validator | `init`, `validate` | planned targets only | schema/Git/pair tests |
| No adjacent effects | bundle, Git allowlist | all | no network/general subprocess/global reads | imports, fake Git, filesystem diff |
| Safe collisions | preflight, Reserved | `init` | canonical APFS policy | symlink/nesting/home/root fixtures |
| Interrupted create | transaction, RecoveryRequired | `recover` | ownership/plan markers | full fault matrix; idempotence |
| Accessible terminal | terminal wizard | `init` | user-controlled flow | plain transcripts; VoiceOver review |
| Personality safety | static AGENTS + JSON profile | generated runtime | profile non-authoritative | malicious/control input tests |
| Zero initial memory | memory template + fresh validator | `init`, `validate` | no record writer | exact tree/record count |
