# Solution Decision

## Decision Drivers

1. Isolate mutable assistant state while retaining the ordinary current-user
   macOS environment.
2. Preserve every existing user Codex, shell, and home path unless an exact
   My Friday manifest proves prior ownership.
3. Make two instances independently operable, inspectable, removable, and
   credentialable without shared mutable directories.
4. Give one native launcher a deterministic, testable environment and working
   directory contract without relying on shell behavior.
5. Fail closed on names, paths, symlinks, manifests, launchers, and migration
   collisions.
6. Make interrupted creation, update, migration, and removal diagnosable and
   recoverable.
7. Keep credentials out of My Friday manifests, logs, tests, receipts, and
   retained acceptance evidence.
8. Reuse manifest and transaction patterns; avoid a daemon, OS account manager,
   container runtime, or general profile framework.

## Competing Approaches

### A. Caller-selected `CODEX_HOME`

Allow every invocation to provide an arbitrary `CODEX_HOME` and perhaps a
working directory.

This exposes Codex's mechanism but does not create an owned product object.
Typos, symlinks, shell state, inconsistent dependencies, and arbitrary target
paths defeat collision checks, lifecycle receipts, multi-instance invariants,
and deletion authority. The caller would remain responsible for reconstructing
the full environment on every invocation.

### B. Substitute `HOME`

Launch Codex with `HOME` set to an instance directory so all home-relative state
appears isolated.

This is broad but semantically false: it changes every child process's view of
the user environment, affecting configuration discovery, keychains, tools,
certificates, Git, and system integrations. It creates compatibility and
credential failures far outside the desired Codex boundary.

### C. Generate aliases or shell functions

Write one alias/function per instance into a shell startup file.

Aliases are shell-specific, invisible to noninteractive execution, difficult
to version and remove safely, and require mutation of mixed-ownership user
files. They cannot provide a trustworthy ownership or collision boundary.

### D. Create a macOS user per assistant

Use separate OS accounts and homes as the isolation primitive.

This provides a stronger general-purpose boundary but requires privileged
account lifecycle, separate login/keychain administration, permission and file
sharing policy, and destructive teardown. It solves a larger problem than the
approved current-user instance outcome.

### E. Manifest-owned named instance and native launcher

Map a validated name to `$HOME/.my-friday/assistants/<name>`, create a complete
instance tree and versioned manifest transactionally, and project exactly one
collision-checked regular native launcher to `$HOME/.local/bin/<name>`. The
launcher fixes instance `CODEX_HOME`, passes the instance workspace through
Codex `--cd`, resolves managed dependencies from instance-owned state,
preserves `HOME`, and forwards only documented literal arguments.

This creates one inspectable lifecycle object without changing the operating
system identity or adopting user-global state.

## Adversarial Comparison

| Approach | Isolation and lifecycle failure | Assessment |
|---|---|---|
| Caller `CODEX_HOME` | Arbitrary paths and ambient setup have no manifest-owned lifecycle boundary | Rejected |
| Substitute `HOME` | Changes unrelated child behavior and credential/config discovery | Rejected |
| Aliases/functions | Mutates mixed-ownership shell files and works only in selected shells | Rejected |
| Separate OS users | Adds privilege, account, keychain, and teardown scope | Rejected |
| Named root + one external native launcher | Filesystem boundary is not an OS sandbox and the external leaf adds collision/removal risk; exact containment and canaries are required | Selected, manageable |

The selected approach does not claim process-level secrecy from other processes
running as the same user. It provides product-owned state isolation, exact
mutation containment, deterministic invocation, independent credentials, and
safe lifecycle authority. Stronger adversarial isolation would be a separate
future product outcome.

## Selected Approach

Choose E with high design confidence and an `implementation` execution
envelope.

Each instance is a versioned manifest-governed aggregate rooted at
`$HOME/.my-friday/assistants/<name>/`. Its required root-owned children are
`codex/`, `runtime/`, `memory/`, `workspace/`, and `dependencies/`. Its sole
owned external projection is the regular native executable
`$HOME/.local/bin/<name>`. For example, `alfred` maps exactly to
`$HOME/.local/bin/alfred`. The manifest records schema, canonical name,
canonical root, fixed managed relative paths, the derived absolute launcher
path and artifact identity, lifecycle state, and migration provenance. It
records no credential value or other arbitrary user path.

Creation plans the complete tree and external launcher leaf, validates every
ancestor and target, and refuses foreign or ambiguous state. The launcher
directory must pre-exist as a real safe current-user directory; My Friday never
creates, chmods, edits, or adopts it. Creation stages and verifies the root,
promotes it without overwrite, then installs the external launcher with an
atomic no-replace operation after a final collision check and verifies the
root/manifest/launcher triple. Failure removes only artifacts whose identity
the new manifest proves; uncertainty preserves them for recovery. Repetition
succeeds read-only only when the complete managed state is canonical; drift
fails with inspect/recover guidance.

The external launcher derives paths from its verified owning manifest rather
than ambient `CODEX_HOME`. It preserves the process's real `HOME`, sets
`CODEX_HOME` to `<root>/codex`, resolves the managed Codex executable and
dependencies from the instance-owned runtime/dependency state, and invokes
Codex with literal `--cd <root>/workspace` plus bounded forwarded arguments. It
never invokes a shell or edits shell files, PATH, the launcher directory, or a
second outside path.

Migration treats the prior O2 projection as a separate manifest-owned source.
It preflights source, destination root, and external launcher. An existing
launcher leaf is replaceable only when the prior O2 manifest proves exact
ownership and artifact identity; otherwise it is a collision. Migration stages
the full named root and candidate launcher, promotes and verifies the root,
atomically replaces only that proven prior launcher (or creates an absent leaf),
then verifies credential-free launch and containment before creating an exact
cleanup plan intersecting the prior manifest's remaining owned paths with the
known O2 schema. Only those proven paths may be deleted. Any missing, invalid,
unexpected, symlinked, changed, or shared path preserves the prior projection
and reports the remaining manual decision. Rollback may restore the exact
verified prior launcher artifact when replacement fails; it may not synthesize
or touch another outside entry.

## Decisions Ledger

| Decision | Rationale | Evidence |
|---|---|---|
| Root at `$HOME/.my-friday/assistants/<name>` | Deterministic, current-user-owned, inspectable namespace | Issue #4 fixed decision |
| Versioned manifest owns the aggregate | Mutation and deletion require durable exact-path authority | Existing manifest lifecycle precedent |
| Fixed root-owned child layout | Enables exact containment, coexistence, verification, and recovery | Acceptance contract |
| External launcher at `$HOME/.local/bin/<name>` | Makes `alfred` directly executable as `$HOME/.local/bin/alfred` without shell edits while retaining a deterministic collision boundary | Approved issue contract |
| Exactly one external projection | Keeps mutation and rollback authority narrow; the launcher directory and siblings remain unmanaged | Preservation requirement |
| Preserve real `HOME` | Avoids changing unrelated tool, keychain, Git, and system behavior | Adversarial comparison |
| Set only instance `CODEX_HOME` and Codex `--cd` | Uses Codex's supported configuration/workspace boundaries narrowly | Fixed Codex contract |
| File-backed credentials under `codex/` | Credentials can differ per instance without entering My Friday state | Multi-instance acceptance |
| Caller cannot select managed roots | Prevents arbitrary-path ownership and deletion authority | Containment requirement |
| Stage and verify before cleanup | Failure leaves the old projection available | Migration safety requirement |
| Delete only prior-manifest-proven O2 paths | Unmanifested and contradictory state is user-owned | Preserve-existing-state requirement |
| No macOS users, aliases, or `HOME` substitution | Avoids privilege and mixed-ownership expansion | YAGNI and fixed exclusions |
| `implementation` envelope | This approval may authorize code and tests, but not merge, migration execution, or release | User-supplied execution boundary |
