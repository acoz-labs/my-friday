# Implementation Handoff

## Change Tier And Smallest Complete Outcome

Broad and risky: a new public terminal workflow, user-owned source transaction,
named-instance contract revision, acceptance-authority change, and artifact
release. The smallest coherent shipped outcome is create plus enhancement with
complete preview, exact source-only confirmation, deterministic postchecks,
safe interruption/recovery, preserved optional bytes, separate activation, and
the full independently accepted/released lifecycle.

## Dependency Order And Reviewable Slices

1. **Proposal and canonical renderer** — own new
   `internal/capabilityworkshop` model/render/diff tests and the narrowest helper
   exports from `internal/capability`; exit on byte-stable valid packages whose
   existing deterministic cases pass.
2. **Source transaction** — own workshop plan/journal/stage/quarantine/recovery
   and fault tests; exit on create/update old-or-new atomicity, shared-lock
   serialization, stale refusal, and exact cleanup authority.
3. **Terminal interaction and CLI** — own injected workshop reader/writer,
   navigation, review, confirmation, TTY/real-home command wiring, stable
   messages, and command tests; exit on all create/enhance/no-write journeys.
4. **Named-instance contract** — revise core builder guidance and capability
   revision so it directs users to the deterministic workshop without granting
   agent source/lifecycle authority; retain explicit instance migration and
   rollback tests.
5. **Acceptance authority** — replace live model-builder authoring and its
   prompt/event/completion proof with exact workshop interaction, transaction,
   comprehension, and recovery evidence; retain fresh Codex invocation,
   runner/stop barriers, ambient canaries, lifecycle, typed receipts, and
   immutable artifact binding.
6. **Reconciliation and docs** — promote shipped behavior, classify/reconcile
   every #56 mechanism, remove the temporary #56 and #74 plan directories, and
   bind reconciliation to the final implementation head.
7. **Artifact release** — merge, nominate, independently accept, collect three
   receipts, publish unchanged bytes, download, verify, and close release-
   bearing issues through the existing ledger.

Each behavioral slice begins with the failing tests in `04-verification.md` and
lands in one reviewable logical commit. Do not allow the public CLI before its
transaction recovery and no-activation invariants exist.

## Acceptance Traceability

- Slices 1 and 3 prove bounded answers, complete bytes/diff, navigation, exact
  consent, Unicode/narrow terminal behavior, and deterministic checks.
- Slice 2 proves create/update atomicity, interruption recovery, collision,
  drift, concurrency, stale preview, and opaque-byte preservation.
- Slices 3 and 4 prove named-instance confinement and that source write never
  installs or upgrades.
- Slice 5 proves the three-person exact-candidate journey, full reversal,
  ambient equality, and that no model statement becomes authority.
- Slices 6 and 7 prove durable documentation, immutable artifact promotion,
  production verification, and rollback.

## Documentation Promotion

- Update `README.md` with workshop create/enhance commands and explicit
  source-versus-activation sequence.
- Rewrite `docs/architecture/capability-workshop.md` around the deterministic
  proposal/renderer/source transaction, state transitions, authority,
  interruption recovery, and remaining fresh-task invocation boundary.
- Supersede ADR 0004 with a short ADR only if implementation changes its core
  source-first decision; otherwise update its provenance note and keep the
  accepted decision intact.
- Update `docs/development.md` with workshop/render/transaction/terminal tests
  and removal of model-authoring proof.
- Update `docs/deployment.md` with deterministic workshop acceptance, revised
  typed receipt, candidate binding, and unchanged artifact release path.
- Update `docs/runbook.md` with source-journal diagnosis/recovery and retire
  prompt/catalog/model-transcript troubleshooting from active guidance.
- Update `SECURITY.md` with source-preview privacy and source-transaction
  authority, and remove live-agent authoring claims that no longer ship.
- Update `docs/product.md` only if implementation evidence changes the already
  promoted D1 promise; ordinary implementation detail belongs in architecture.

## Pull Request And Review Contract

Implementation starts in a new task-scoped worktree from the exact fetched
`origin/main`, on a neutral `feature/deterministic-capability-workshop` branch.
The draft PR uses `Refs #74` and also links #51/#56 without closing release-
bearing issues at merge. It must include failing-first commits, focused and full
validation, native terminal/APFS evidence, findings-first independent
maintainer/security review, and product-design review of the exact terminal
flow.

Before leaving draft, reconcile the implementation against this plan and the
superseded #56 plan, explain every material drift, promote durable docs, delete
temporary plan directories, and verify reconciliation on the exact head. No
preserved diagnostic root or private transcript enters the repository or PR.

## Explicit Non-Goals And YAGNI Boundary

No agent adapter, Codex client, public proposal file, `--spec`, noninteractive
mode, general form engine, TUI framework, second capability profile, arbitrary
file editor, template marketplace, scripts, dependency management, network,
credentials, background execution, durable capability data, analytics,
automatic Git commit, automatic Install/Upgrade/Enable, or source deletion.

Do not generalize the workshop before governed memory supplies a second real
profile. Do not expose internal proposal or source-journal structs as public
compatibility contracts.

## Exceptions That Reopen Design

Return to Solution Design if safe source replacement requires deletion
authority broader than exact journal/inode/digest proofs; source and lifecycle
cannot share one lock without breaking existing recovery; valid enhancement
cannot round-trip the canonical package without content loss; terminal
acceptance shows users cannot distinguish source write from activation; a new
credential/network/data boundary becomes necessary; the release would rebuild
accepted bytes; or implementation requires any #75 agent-adapter behavior.
