# Solution Decision

## Decision Drivers

The design must preserve versioned inspectable source, named-instance isolation,
explicit mutation authority, exact reversal, deterministic offline checks,
Codex compatibility, v1 compatibility, and the repository's proven transaction
and release patterns. It must solve the builder bootstrap loop without creating
a general extension runtime.

## Competing Approaches

### A. Install user skills globally

Write packages directly to `$HOME/.agents/skills` and use Codex's native skill
discovery. This is mechanically small and supports all Codex sessions.

### B. Treat source as the installed projection

Point `.agents/skills/<slug>` at the version-controlled runtime package with a
symlink. Review, enhancement, and activation would share one tree.

### C. Bootstrap a core builder and copy strict projections per instance

Add explicit runtime and named-instance contract-v2 migrations. Keep complete
packages in `runtime/skills/<slug>`, copy only the validated `skill/` subtree
into `<instance>/workspace/.agents/skills/<slug>`, and own installed receipts
under the private instance. New instances start at v2; existing v1 state remains
valid until the user explicitly upgrades.

### D. Build a general plugin/package system first

Model code, dependencies, remote registries, permissions, and multiple execution
profiles before shipping the first capability.

## Adversarial Comparison

| Approach | Strength | Fatal flaw or tradeoff |
|---|---|---|
| A: global skills | Minimal Codex integration | Breaks named-instance isolation, broadens collision/removal scope, and mutates unrelated assistants |
| B: source equals projection | No copy or receipt drift | Source edits activate immediately, defeating review/confirmation and source-versus-installed evidence |
| C: core builder + copies | Preserves authority, isolation, inspectability, and reversal using current patterns | Requires two explicit migrations and duplicate bytes; current-session unloading is not guaranteed |
| D: general plugin system | Future breadth | Violates C1's trust boundary and adds code, dependency, registry, permission, and publishing questions before evidence warrants them |

Using Codex configuration to disable skills was also rejected for C1 lifecycle:
it changes a broader managed configuration and requires restart. Removing and
restoring the exact instance-owned projection gives clearer ownership, reversal,
and fresh-task semantics.

## Selected Approach

Select C with high design confidence. It extends current manifest/transaction
patterns and the official repository-scoped Codex skill surface. The built-in
builder is embedded release content and projected during named-instance v2
creation/upgrade. It is not represented as user-authored source and is reserved
from ordinary removal.

Runtime source contract v2 adds typed instruction-only package schemas and
replaces the placeholder with zero or more packages. Instance contract v2 adds
builder and managed-capability state. The migrations are ordered but independent:
runtime first, instance second. Ordinary user capability operations thereafter
mutate only the instance.

The design deliberately separates deterministic package tests from model
behavior. `capability test` never calls a model or network service. Independent
candidate acceptance uses Codex to demonstrate actual discovery and behavior.

## Decisions Ledger

| Decision | Rationale | Evidence |
|---|---|---|
| Package source under `runtime/skills/<slug>` | Reuses the reserved O1 source surface | `internal/repository`; bootstrap evidence |
| Installed target under instance workspace | Codex discovers repo-scoped skills from fixed launcher CWD | Official Codex skill docs; `internal/assistantinstance` |
| Copy, never symlink | Prevents source edits from activating and supports digest ownership | Existing copied instance model |
| Built-in builder as v2 bootstrap projection | Resolves the builder-before-first-capability loop | Approved C1 product contract |
| Separate runtime and instance migrations | Avoids ambiguous cross-root transactions | Existing single-root transaction precedent |
| Explicit invocation for user skills | Limits unexpected instruction injection in C1 | `agents/openai.yaml` official contract; security review |
| Fixed rendered `agents/openai.yaml` | Users cannot add dependencies or broaden invocation policy | Strict instruction-only profile |
| Static deterministic tests plus live acceptance | Honest boundary between repeatable validation and model behavior | No deterministic offline model executor exists |
| Preserve source on disable/remove | Source is user-owned; projections are My Friday-owned | Product-design contract |
| Fresh-task activation semantics | Avoids false claims about in-session Codex cache/state | Official docs advise restart if skill is not detected |
