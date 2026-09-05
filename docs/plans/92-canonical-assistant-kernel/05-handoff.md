# Implementation Handoff

## Change Tier And Smallest Complete Outcome

**Tier: broad, release-bearing foundation change.**

The smallest complete outcome is not merely a new directory layout. One
immutable artifact must create a remote-backed canonical assistant repository,
verify and diagnose it, launch from a separately owned generated projection,
perform one exact commit/push transaction with interruption recovery, migrate a
released split pair without altering it, roll a compatible baseline forward
and back through descendant commits, and remove generated state while source
and memory survive locally and remotely.

## Dependency Order And Reviewable Slices

1. **Canonical contracts and read model**
   - Ownership: new `internal/assistantrepo`, embedded schemas, stable states,
     preview/JSON fixtures, command parsing tests.
   - Exit: deterministic fresh plan and complete validation with no I/O
     mutation; released contracts remain characterized.
2. **Bounded Git adapter and operation journal**
   - Ownership: extend `internal/gitexec`; add repository-steward transaction
     package and local bare-remote fixtures.
   - Exit: create/change commits and explicit pushes pass, unsafe Git traces
     refuse, and every partial phase recovers idempotently.
3. **Canonical create plus host binding**
   - Ownership: adapt `internal/assistantinstance`, retain reusable executable,
     config, launcher, and no-follow projection primitives.
   - Exit: `assistant create/inspect/verify/diagnose/reconcile/repair/remove`
     work end to end; remove plans cannot name canonical paths.
4. **Legacy migration**
   - Ownership: repository-pair adapter plus instance switch transaction.
   - Exit: every supported released pair imports with provenance, old state is
     unchanged, and injected switch failures leave exactly one usable path.
5. **Baseline migration chain and rollback**
   - Ownership: embedded migration registry, affected-module validation,
     canonical receipts.
   - Exit: upgrade and lossless inverse rollback make descendant commits;
     incompatibility/divergence refuse before source mutation.
6. **CLI experience, documentation, and immutable acceptance**
   - Ownership: `cmd/my-friday`, terminal goldens, help/error mapping,
     acceptance tooling, durable docs and release scripts.
   - Exit: container/native CI, real private-remote journey, clean-host restore,
     migration, removal, artifact nomination, and independent acceptance pass.

Each slice begins with the failing tests in `04-verification.md`. The
implementation may use multiple internal commits but should remain one
reconciled implementation PR because the external contract is not useful or
safe in partially released form.

## Acceptance Traceability

| Acceptance group | Slices | Required evidence |
|---|---|---|
| One repository/three governed modules | 1, 3 | schema/semantic tests; fresh exact-candidate creation |
| Setup, inspect, reconcile, verify, diagnose, repair, launch, remove | 3, 6 | command goldens; native full lifecycle; preservation canaries |
| Exact commit/push and refusal boundaries | 2, 6 | Git trace suite; real private-remote success/divergence/ambiguity |
| Upgrade/rollback/migrate/recover | 2, 4, 5, 6 | phase matrix; released legacy fixtures; native interruption journey |
| Portability after host failure | 3, 6 | second clean-host restore from remote with rebuilt projection |

See `04-verification.md` for the complete matrix and candidate-binding rules.

## Documentation Promotion

Implementation reconciliation should update, based on shipped behavior:

- `docs/product.md` — canonical-repository promise, critical tasks, supported
  compatibility window, and B1 boundaries;
- `docs/architecture.md` — one-repository/kernel/generated-projection overview;
- `docs/architecture/repository-bootstrap.md` — canonical layout, Git
  transaction, manifests, remote, migration, and recovery contract;
- `docs/architecture/named-assistant-instances.md` — host binding and external
  canonical-source ownership;
- a new ADR replacing the split-repository default and recording rejected
  branch/submodule/instance-owned approaches;
- `SECURITY.md` — remote/credential, same-user, hooks/filters, secret rejection,
  and non-publication boundaries;
- `docs/development.md` — fixtures, focused tests, native requirements;
- `docs/deployment.md` — candidate, migration compatibility, acceptance, and
  release prerequisites; and
- `docs/runbook.md`/README — create, inspect, diagnose, recover, migrate,
  rollback, remove, and clean-host restore commands.

The implementation PR must delete `docs/plans/92-canonical-assistant-kernel/`
after promoting these contracts and record the exact promotion matrix in its
reconciliation.

## Pull Request And Review Contract

- Branch from then-current `origin/main` only after this plan is approved and
  merged; use one task-scoped worktree and a neutral `feature/...` branch.
- Open one draft implementation PR with `Refs #92`; do not close the issue at
  merge because artifact acceptance and release remain.
- Require `bin/container bin/ci`, native `bin/ci`, Git transaction trace tests,
  migration fixtures, and exact-candidate acceptance evidence.
- Maintainer review is findings-first with special attention to deletion
  authority, Git invocation traces, URL/error redaction, journal ambiguity,
  migration preservation, and compatibility.
- Reconcile the exact draft head against this plan, explain any drift, promote
  durable docs, remove the plan, and pass
  `solution-design-plan verify-implementation` before leaving draft.
- The `through-production` envelope permits merge, exact artifact nomination,
  independent acceptance, and release only under existing repository gates and
  after the F0 dependency is released.

## Explicit Non-Goals And YAGNI Boundary

Do not add a database, daemon, server, VM, container runtime, provider SDK,
OAuth flow, secrets manager, merge engine, sync service, second harness,
submodule/LFS support, public registry, signed catalog, B2 component/compiler
semantics, or B3 memory-record semantics. Do not combine legacy Git histories,
delete legacy/canonical repositories, auto-resolve divergence, broaden Codex
filesystem authority, or promise malicious same-UID resistance.

Do not generalize the adapter until a second real Git transport or harness
forces a proven abstraction. Internal interfaces should follow the selected
ports, but one Codex/Apple-silicon implementation is sufficient.

## Exceptions That Reopen Design

Return to Solution Design for a required provider-specific privacy/API check,
credential storage, destructive canonical-data cleanup, history rewrite,
automatic conflict resolution, multi-writer synchronization, background
service, broader OS/harness target, or any discovery that same-UID policy
cannot enforce a required authority boundary. Ordinary implementation choices
inside the described manifests, transaction phases, commands, tests, and
release path do not require another product gate.
