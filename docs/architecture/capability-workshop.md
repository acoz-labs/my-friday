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

Before prompting, the source workshop walks the supplied already-canonical
absolute instance path descriptor-relatively from `/` without resolving it and
refuses every symlink component. It records every instance-path ancestor's
device, inode, directory type, owner, and mode. Commit reopens that same chain,
refuses any replacement, and derives both
the shared `capabilities/` lock and `runtime/skills` from the same verified root
descriptor. The workshop creates a random, exclusive staging root descriptor-
relatively beneath that opened `runtime/skills` directory. Every staged
directory and file is created through no-follow descriptors, so replacing the
pathname with a symlink cannot redirect a write. The canonical source journal
records that stage name and the complete staged and prior-tree device, inode,
owner, mode, link-count, type, path, and content-digest authority. Recovery
revalidates those facts before promotion or cleanup and refuses chmod or entry-
set drift. Cleanup validates the complete old tree and package digest twice
before its first unlink, journals each file and directory unlink, and on
re-entry treats a missing expected entry only as completed work while every
survivor still matches its original inode and metadata. Additions and
substitutions are preserved and refused. Pre-journal staging records exact
entry inode, digest, type, owner, mode, and link authority after each successful
construction step. Failure removes only that exact descriptor-bound tree;
same-owner pathname substitution or other ambiguous residue is preserved and
refused rather than deleted.

Retained `SKILL.md` bodies preserve their suffix bytes, including whether the
last byte is a newline. Generated frontmatter renders the summary as a JSON-
quoted YAML scalar. Complete source diffs retain blank lines and emit the
standard missing-final-newline marker, so review accounts for every source
byte.

`capability test` is a bounded structural contract check, not a model run. It
requires nonempty unique declarations, an exact normalized match between
manifest and positive triggers,
disjoint non-triggers, positive-trigger examples with declared output
expectations, instruction-backed required facts, and assertions for every
prohibited effect. Structurally readable but contradictory cases are
`test-failed`, never `ready`.

`capability workshop NAME SLUG` is the guided source-authoring interface. It
verifies the real-home named instance, refuses unsafe source/control states,
and holds answers only in memory. One proposal renders the existing strict
package contract; no proposal file or second public schema exists. The fixed
sequence collects identity, purpose, success and failure, triggers and
non-triggers, inputs and outputs, examples, and required facts. All seven
prohibited effects remain fixed to `none`.

Final review prints the complete canonical three-file package, the complete
core diff, unchanged optional-file digests, source action, current state, and
post-write state. Existing valid instruction bodies are retained byte-for-byte
unless regeneration is explicit. Only exact `Create source` or `Update source`
authorizes one source transaction. Return, EOF, `q`, interruption, and every
other token leave source unchanged. Source confirmation never authorizes or
calls Install, Upgrade, Enable, Disable, Remove, or lifecycle recovery.
If INT or TERM arrives during a confirmed transaction, its transaction or
recovery error remains primary; after a successful durable flush the command
returns a stable interruption error rather than reporting ordinary success.

Source mutation shares the non-blocking instance `capabilities/` lock with
lifecycle mutation and uses a separate mode-0600
`capabilities/.workshop-<slug>.json` journal. It re-inspects previewed facts
while locked, stages and validates the complete package with byte-identical
optional files, and promotes only the exact staged source. A retained canonical
journal is `interrupted` with interruption kind `source-workshop`; a valid
lifecycle journal is separately `interrupted` with kind `lifecycle`. Malformed
or contradictory authority is `recovery-required`. Re-entering the workshop
performs only digest-proven source-workshop recovery and exits before collecting
answers; lifecycle interruption directs the existing `capability recover`
command and is not consumed by the workshop.

Successful create reports `ready` and a separate Install command. Updating an
active projection reports `source-changed` and a separate Upgrade command;
disabled source remains disabled. Postconditions call inspect, validation, and
deterministic tests directly and never invoke lifecycle mutation.

The bootstrap-owned `capability-builder` is private to each named instance and
is not an ordinary removable package. Revision 3 names the instance and its
manifest-owned `dependencies/my-friday` executable, directs users to the
deterministic workshop, and prohibits direct source edits, lifecycle mutations,
and every confirmation token. Managed Codex retains workspace-write with only
that private runtime as an additional writable root, approvals never, and
network disabled, but model output is not source or acceptance authority.

Standalone runtimes can roll back to the v1 placeholder only while `skills/`
contains no package. Named instances use an explicit capability revision inside
contract v2. Revision 3 is the deterministic-workshop contract; supported older
manifests remain accepted only for a bounded explicit upgrade and rollback to
their exact prior revision. New revision-3 instances
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
including after interrupted rollback. Revision-2 verification also binds the
managed executable bytes and mode. A rollback to revision 2 journals executable
restoration before manifest promotion, so recovery can resume on either side of
that promotion without requiring the already-restored rollback-source name.
