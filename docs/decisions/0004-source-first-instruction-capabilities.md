# ADR 0004: Source-first instruction capabilities

- Status: accepted
- Provenance: issue #51, planning PR #54

Keep strict instruction-only source in the copied runtime and install validated
copies only into the selected named assistant workspace. Bootstrap a core
builder, render explicit-invocation policy, and retain receipt-bound generations
for reversal.

Source edits do not activate automatically, diffs remain Git-visible, and
disable/remove never delete source. The implementation duplicates a small set
of bytes and promises only fresh-task activation. Global skills, symlinks,
plugins, scripts, dependencies, network, credentials, and registries remain
outside the product boundary.
