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

Every mutation takes a non-blocking exclusive lock on the instance capability
root, recomputes the previewed source/projection facts, and writes
`capabilities/<slug>/transaction.json` before changing state. A retained journal
blocks later lifecycle plans. `capability recover NAME SLUG` uses only the
receipt and retained generation to restore installed, disabled, or absent state;
it does not read, rewrite, or remove source.

Observable state precedence is: incompatible ownership or an invalid journal;
valid interrupted transaction; foreign collision; source draft/test state;
then receipt/projection state. The stable vocabulary is `absent`,
`draft-invalid`, `draft-valid`, `test-failed`, `ready`, `installed-healthy`,
`source-changed`, `installed-drift`, `disabled`, `collision`, `interrupted`,
`recovery-required`, and `incompatible`.

Control and journal paths are authority, not mere markers. A journal must be a
canonical, one-link, mode-0600 regular file whose action, slug, source and
projection digests, prior-receipt digest, and control-creation fact match the
observed state. Recovery refuses malformed, linked, drifted, ambiguous, or
foreign entries and removes only digest-proven projections and exact owned
control leaves.

Projection writes use no-follow directory descriptors, exclusive file creation,
and no-replace promotion. Replacement and removal first move the immediately
identity- and digest-verified owned tree to an exclusive quarantine, reverify it,
and refuse cleanup if any raced entry changes the tree. A raced foreign target
or quarantine is preserved for diagnosis.

`capability test` is a bounded structural contract check, not a model run. It
requires nonempty unique declarations, manifest-aligned positive triggers,
disjoint non-triggers, positive-trigger examples with declared output
expectations, instruction-backed required facts, and assertions for every
prohibited effect. Structurally readable but contradictory cases are
`test-failed`, never `ready`.

The bootstrap-owned `capability-builder` is private to each named instance and
is not an ordinary removable package. Its instructions permit source editing
and read-only checks while prohibiting lifecycle mutations and confirmation
tokens. This is an instruction boundary, not an OS boundary.

Standalone runtimes can roll back to the v1 placeholder only while `skills/`
contains no package. Named instances can roll back to v1 only while capability
control is empty and the builder is exact and alone in the managed skill root.
Both paths use an exact `Rollback` preview and preserve source, credentials,
launchers, global skills, and sibling instances.

Runtime initialize/rollback and named-instance upgrade/rollback are serialized
one-root migrations with canonical durable journals and explicit recovery. An
instance upgrade initializes its private copied runtime; it never follows or
mutates the external runtime originally used to create the instance. Rollback
rechecks capability control and workspace-skill observations under the held
instance lock before quarantining exact owned leaves.
