# Solution Decision

## Decision Drivers

1. Never alter unrelated configuration or credential-bearing state.
2. Make every write attributable to a manifest and exact pre-change preview.
3. Survive interruption without speculative deletion or overwrite.
4. Provide meaningful verify, repair, uninstall, and one-step rollback behavior.
5. Match documented Codex discovery while keeping compatibility bounded and
   observable.
6. Remain a dependency-light local Go artifact with no daemon, database,
   network, privilege escalation, or hidden background work.
7. Be testable without any dependency on the developer's home or live Codex
   installation.

## Competing Approaches

### A. Documentation-only manual copy

Tell users to copy the runtime `AGENTS.md` to `~/.codex/AGENTS.md` and manually
back it up before changes.

This is small but cannot prove ownership, detect source or target drift,
recover an interrupted write, distinguish a foreign collision, or completely
reverse an installation. It repeats the primary approach rejected in discovery.

### B. Edit or merge `config.toml`

Install the runtime through user-level configuration, perhaps using
`model_instructions_file`, while preserving other TOML keys.

This would enter a mixed-ownership, credential-adjacent contract with parsing,
format/comment preservation, precedence, profile, and future-key hazards.
Official documentation already provides a direct global `AGENTS.md` discovery
surface, so editing configuration buys no necessary capability for O2.

### C. Symlink the runtime repository into Codex home

Create `$CODEX_HOME/AGENTS.md` as a symlink to the runtime repository and keep
the source live.

This makes edits immediate but couples installed behavior to source moves and
worktree state, complicates relative references, expands symlink/ancestor race
risks, and makes exact rollback ambiguous. A broken or replaced source changes
Codex behavior outside a My Friday transaction.

### D. Manifest-owned rendered file with transactional control state

Render one regular `AGENTS.md` from the validated runtime profile, install it
atomically, record exact ownership and provenance under a private control
namespace, and retain one prior generation for rollback.

This duplicates a small rendered file but gives the strongest bounded
ownership, testability, recovery, and reversal contract. It is selected.

### E. Manage the entire Codex home

Treat `CODEX_HOME` as a generated directory or mirror, including configuration,
auth, sessions, skills, and packages.

This is incompatible with the issue's unrelated-configuration promise and the
official state layout. It would turn My Friday into a Codex distribution and
credential/state migration tool, which discovery did not authorize.

## Adversarial Comparison

| Approach | Fatal flaw | Manageable tradeoff | Evidence |
|---|---|---|---|
| Manual copy | No durable ownership or recovery; cannot meet acceptance | Minimal implementation | Discovery rejected templates/docs as primary solution |
| Config edit | Mixed ownership and future-key/comment preservation create avoidable destructive risk | Could centralize configuration | Official config docs place many unrelated settings in user config |
| Symlink | Source movement or replacement changes installed behavior outside a transaction | Immediate source updates | Current runtime instruction contains relative source semantics; Codex discovers the destination path |
| Rendered file | No automatic reflection of source edits | User runs verify/upgrade; exact bytes are inspectable | Existing repository already uses deterministic plans, hashes, and atomic promotion |
| Whole-home management | Encompasses auth, logs, sessions, skills, and package metadata | Unified snapshot | Official `CODEX_HOME` contract proves the trust boundary is too broad |

An atomic rename alone is not enough for option D: without a durable journal,
an interruption between projection and manifest publication could orphan
ownership. Conversely, journaling without revalidating target inode/digest
could overwrite a raced foreign file. The selected path requires both.

## Selected Approach

Implement a dedicated installed-baseline domain with injected environment and
filesystem boundaries, deterministic plan and manifest schemas, an exclusive
transaction lock, staged regular files, atomic replacement, fault checkpoints,
and conservative recovery. Expose it through `my-friday codex install|verify|
repair|upgrade|rollback|uninstall|recover`.

The installed projection is a stable My Friday-owned template rendered from
the validated runtime `assistant/profile.json`; arbitrary edits to source
`AGENTS.md` are not installed in O2. The rendering is self-contained and
contains only presentation identity/purpose/style plus the invariant that those
preferences do not override authorization, safety, trust, privacy, or tool
policy. This avoids both dangling relative references and turning repository
prose into an unchecked global instruction channel.

Confidence is high in the architecture and medium-high in the Codex
compatibility claim until exact-candidate acceptance proves discovery under a
fresh macOS identity. The release envelope should be `through-production`
because this artifact repository already has nomination, independent
acceptance, and release workflows; the plan adds the candidate-specific
disposable-user precondition rather than inventing staging.

## Decisions Ledger

| Decision | Rationale | Evidence |
|---|---|---|
| Manage only global `AGENTS.md` plus private control state | Smallest effective projection; avoids mixed config/auth ownership | Official Codex config, environment, and AGENTS discovery docs |
| Render a regular file from validated profile data | Self-contained, deterministic, atomic, and move-safe | `internal/plan/plan.go`; symlink alternative analysis |
| Refuse foreign global override or projection | Cannot merge global instructions safely or claim activation while shadowed | Official precedence and discovery behavior |
| Resolve and pin effective Codex home once | Prevent preview/execute root drift | `CODEX_HOME` official contract; transaction precedent |
| Restrict production root to current user's home and local APFS | Narrows destructive authority to the supported pilot boundary | `README.md`, `docs/development.md`, issue safety outcome |
| Use internal root injection for tests, not inherited `HOME` | Makes live-home mutation structurally unavailable to test helpers | Existing `t.TempDir` patterns; trust-boundary analysis |
| Keep one prior verified generation | Meets rollback without building a backup manager | YAGNI and selected outcome |
| Drift blocks uninstall/upgrade/rollback; repair is explicit | Prevents deleting or replacing user-changed bytes accidentally | Ownership/digest acceptance criterion |
| Run exact candidate under disposable non-admin user/home | Git checkout isolation does not isolate Codex state, keychain, or UID | Official `CODEX_HOME`; acceptance risk analysis |
| No staging for the artifact | Repository policy explicitly forbids fictional staging | `AGENTS.md`, `docs/operations/sdlc.md`, `docs/deployment.md` |
