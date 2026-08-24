# Named assistant instances

A canonical lowercase ASCII name maps to one private root at
`$HOME/.my-friday/assistants/<name>` and one executable launcher at
`$HOME/.local/bin/<name>`. Names are 1–32 characters with lowercase letters,
digits, and interior hyphens; reserved names, case folding, Unicode,
separators, and traversal are refused.

The root contains `manifest.json` plus `codex/`, `runtime/`, `memory/`,
`workspace/`, and `dependencies/`. Runtime and memory are copied from a
validated generated pair; `codex/AGENTS.md` comes from that runtime and Codex
is copied to `dependencies/codex`. The manifest fixes the schema, name, root,
owned children, launcher path/digest, managed Codex path, and assistant ID.

Create verifies the existing current-user-owned `$HOME/.local/bin`, refuses
collisions, stages the root, then uses a no-replace launcher projection. Verify
re-derives paths and checks the layout and launcher bytes. Remove requires a
valid manifest and matching launcher, removes that exact leaf, then the selected
root. `assistant recover <name>` finishes an interrupted removal only when the
launcher is absent and the retained manifest still proves the exact root.
Foreign, linked, drifted, unsupported, or contradictory state is preserved.

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
