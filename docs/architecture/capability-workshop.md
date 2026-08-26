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
and refuse cleanup if any raced entry changes the tree. Final cleanup recursively
unlinks entries through the opened, reverified directory descriptor and removes
the root only while its inode remains bound. Recovery proves deterministic
quarantine bytes, uses a receipt-derived `.restoring` handle, and can resume
cleanup or restoration without an unjournaled random path. Before the first
unlink, a strict sidecar cleanup manifest outside the payload binds the root
inode, exact directory set, and per-file digests. Missing expected entries are
completed work; every surviving entry must still match, so foreign additions
cannot gain deletion authority. Control cleanup retains this external authority
even if its journal-bearing payload is partly removed. A raced foreign target or
quarantine is preserved for diagnosis.

Cleanup authority is written as canonical bytes to an exclusive same-directory
temporary file, synced and closed, then promoted with no replacement. Recovery
removes a pre-promotion temp only after exact root inode/digest proof and never
uses it as authority. A valid final manifest whose bound root is absent records
a committed unlink; recovery removes the manifest and continues, while any
reappeared different inode is preserved and refused.

Short-write and pre-sync temporary residue is reconstructed only from the live
transaction/receipt plus an exact whole-target proof. Unsafe temp type, link,
mode, owner, or target identity is preserved and refused. The containing
directory is synced after authority promotion and after final authority removal.

`capability test` is a bounded structural contract check, not a model run. It
requires nonempty unique declarations, an exact normalized match between
manifest and positive triggers,
disjoint non-triggers, positive-trigger examples with declared output
expectations, instruction-backed required facts, and assertions for every
prohibited effect. Structurally readable but contradictory cases are
`test-failed`, never `ready`.

The bootstrap-owned `capability-builder` is private to each named instance and
is not an ordinary removable package. Its instructions permit source editing
only beneath that instance's exact private runtime. They name the instance and
manifest-owned `dependencies/my-friday` executable, and allow only exact
`inspect`, `validate`, and `test` command forms while prohibiting lifecycle
mutations and confirmation tokens. Managed Codex enforces workspace-write with
only that private runtime as an additional writable root, approvals never, and
network disabled. The trusted workspace remains writable Codex workspace state;
no sibling instance, ambient runtime, or broader root is granted.
Exact-candidate acceptance starts the builder prompt with the literal
`$capability-builder` invocation. Before opening the private PTY task, a bounded
managed-Codex `debug prompt-input` preflight proves that the exact prompt digest,
builder description, and exact skill file path are model-visible in one unique
builder catalog record and that its reported skill-root alias resolves uniquely
to the instance's exact private workspace skill root.
The preflight emits no transcript or prompt content into public evidence.
Under the private PTY, the builder drains pre-prompt output, waits for a fresh
builder-name, description, and insert-hint autocomplete event, and drains that
event before sending one CSI-u Enter to select the literal skill mention. It
then requires the fresh selected-composer redraw with the extra separator that
Codex 0.149 inserts before an existing suffix, refusing completion or a missing
selection state at that boundary, before sending a separate CSI-u Enter to
submit the composed task. Non-mention workshop prompts retain one-key submission.
Every output-drain boundary consumes incrementally with rolling marker detection;
a marker split across bytes or coalesced with a selected-composer redraw fails
before later output or a submit key can hide it.

Standalone runtimes can roll back to the v1 placeholder only while `skills/`
contains no package. Named instances use an explicit capability revision inside
contract v2. Revision 2 is the instance-specific execution contract; an
unversioned revision-0 v2 manifest remains accepted only for a bounded upgrade
to revision 2 and rollback to its exact prior revision. New revision-2 instances
roll back to v1. Either rollback requires empty capability control and the exact
builder alone in the managed skill root.
Both paths use an exact `Rollback` preview and preserve source, credentials,
launchers, global skills, and sibling instances.

Runtime initialize/rollback and named-instance upgrade/rollback are serialized
one-root migrations with canonical durable journals and explicit recovery. An
instance upgrade initializes its private copied runtime and stages a digest of
the currently executing candidate before installing that exact executable with
the manifest-bound builder/config; it never derives builder authority from the
old launcher or follows or mutates the external runtime originally used to
create the instance. The journal binds source manifest, source/target revision,
and candidate digest for recovery. Rollback restores the recorded prior config
and revision and rechecks capability control and workspace-skill observations
under the held instance lock before quarantining exact owned leaves. Foreign
quarantine names are refused. Quarantine tree type, owner, mode, link count, and
digest remain journal-bound and are revalidated immediately before deletion,
including after interrupted rollback.
