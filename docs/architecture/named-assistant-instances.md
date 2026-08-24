# Named assistant instances

A canonical lowercase ASCII name maps to one private root at
`$HOME/.my-friday/assistants/<name>` and one executable launcher at
`$HOME/.local/bin/<name>`. Names are 1–32 characters with lowercase letters,
digits, and interior hyphens; reserved names, case folding, Unicode,
separators, and traversal are refused.

The account home comes from the current operating-system account record, never
from caller-supplied `HOME`; the launcher's child environment nevertheless
preserves the caller's `HOME` value unchanged.

The root contains `manifest.json` plus `codex/`, `runtime/`, `memory/`,
`workspace/`, and `dependencies/`. Runtime and memory are copied from a
validated generated pair; `codex/AGENTS.md` comes from that runtime and Codex
is copied to `dependencies/codex`. The manifest fixes the schema, name, root,
owned children, launcher path/digest, the exact
`<root>/dependencies/codex` path and digest, and assistant ID. PATH discovery
may traverse a symlink, but creation resolves it first and copies only the
resolved current-user regular executable into managed state.

Create verifies the existing current-user-owned `$HOME/.local/bin`, refuses
collisions, stages the root, then uses a no-replace launcher projection. Verify
re-derives paths and checks the layout and launcher bytes. Remove requires a
valid manifest and matching launcher, removes that exact leaf, then the selected
root. `assistant recover <name>` finishes an interrupted removal only when the
launcher is absent and the retained manifest still proves the exact root.
Foreign, linked, drifted, unsupported, or contradictory state is preserved.
Atomic staging serializes competing creates. Other mutations use a non-blocking,
transaction-owned advisory lock inside the selected instance root; create
carries that in-root lock through promotion, final verification, and legacy
cleanup during migration. Refused waiters never retain an unlinked lock inode,
and successful completion removes the transient lock while still holding it.
No per-name control leaf survives outside the root or after removal, and no
daemon is involved.

The launcher is a copy of the native `my-friday` executable. Invocation by its
instance basename selects launcher mode, verifies ownership, preserves `HOME`,
discards ambient `CODEX_HOME`, sets instance `CODEX_HOME`, and executes managed
Codex with `--cd <workspace> --`. No shell, startup file, PATH edit, daemon, or
second outside projection participates. This is lifecycle isolation under one
UID, not a confidentiality boundary.

The legacy `codex` lifecycle remains available for explicit prior-projection
repair, rollback, uninstall, and recovery. `assistant migrate` plans both
halves, creates and verifies the named replacement first, then delegates old
cleanup to that manifest-governed uninstall transaction. Cleanup failure leaves
the replacement active and the old journal recoverable. Ordinary named
creation does not discover, adopt, or delete prior state.
