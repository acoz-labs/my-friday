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
Atomic staging serializes competing creates. The transaction holds a
non-blocking advisory lock on the no-follow opened staging directory and carries
that same directory identity through promotion, final verification, and legacy
cleanup during migration. Active-root mutation opens the root without following
links, proves its owner, mode, device/inode, manifest, and expected generation,
then locks the directory descriptor without writing any lock artifact. Path
replacement and foreign recovery roots are refused unchanged. No per-name
control leaf survives outside the root or after removal, and no daemon is
involved.

The launcher is a copy of the native `my-friday` executable. Invocation by its
instance basename selects launcher mode, verifies ownership, preserves `HOME`,
discards ambient `CODEX_HOME`, sets instance `CODEX_HOME`, and executes managed
Codex with `--cd <workspace> --`. No shell, startup file, PATH edit, daemon, or
second outside projection participates. This is lifecycle isolation under one
UID, not a confidentiality boundary.

Release acceptance runs randomized instances beneath the current account's
real root, regardless of caller `HOME`, and proves create, verify, PTY launch,
two-instance isolation, foreign collision preservation, interrupted-remove
recovery, and complete reversal. Credential-free lifecycle evidence is
mandatory. A separate live smoke may copy an existing current-user `auth.json`
byte-for-byte into one disposable instance, but it performs no login, records
no credential-derived value, and must prove both the copy and every disposable
instance leaf absent before its exact-candidate evidence can authorize release.
That smoke invokes the launcher with an empty forwarded argv under a real PTY,
submits the purpose prompt interactively, and requires the installed token in
the response. Failure cleanup reuses manifest verification/removal or recovery
for instances and exact no-follow receipts for harness-created foreign leaves.
The preservation record separately identifies its disposable caller-shell
canary and unrelated launcher sibling; it does not generalize those observations
into an unmeasured byte-for-byte claim about the entire account home.

The legacy `codex` lifecycle remains available for explicit prior-projection
repair, rollback, uninstall, and recovery. `assistant migrate` plans both
halves, creates and verifies the named replacement first, then delegates old
cleanup to that manifest-governed uninstall transaction. Cleanup failure leaves
the replacement active and the old journal recoverable. Ordinary named
creation does not discover, adopt, or delete prior state.
