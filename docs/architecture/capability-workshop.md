# Capability workshop

My Friday supports one package profile: `instruction-only`. User-owned source
is versioned under `runtime/skills/<slug>/`; installed state is a copied,
managed projection under the selected named assistant's
`workspace/.agents/skills/<slug>/`. Source and installed bytes are never the
same tree.

Each package contains canonical `capability.json`, `skill/SKILL.md`, and
`tests/cases.json`, plus optional regular files beneath `skill/references/` and
`skill/assets/`. Paths, UTF-8, depth, counts, sizes, link count, file type, and
non-executable modes are bounded. Unknown fields and entries fail closed.
Scripts, user policy, dependencies, network, credentials, background work,
durable data, and publishing are forbidden.

The projection contains allowed skill bytes plus fixed policy disabling
implicit invocation. Deterministic tests check declared triggers,
non-triggers, examples, facts, and forbidden effects without a model or
network. Structural validation cannot prove instruction semantics are benign.

Receipts and retained generations live under `capabilities/<slug>/`. Disable
removes the active projection but retains exact bytes for enable. Remove proves
managed state, clears projection and control state, and never deletes source.
Changed or missing installed bytes are drift and are preserved. Lifecycle
effects are guaranteed for fresh Codex tasks only.

The bootstrap-owned `capability-builder` is private to each named instance and
is not an ordinary removable package. Its instructions permit source editing
and read-only checks while prohibiting lifecycle mutations and confirmation
tokens. This is an instruction boundary, not an OS boundary.
