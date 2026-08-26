# Implementation Handoff

## Change Tier And Smallest Complete Outcome

Risky, release-bound verification and generated-agent-contract change. The
smallest complete outcome is an instance-local builder that contains the full
capability package contract, a bounded native-exec authoring path whose receipt
requires exact candidate postconditions, retained PTY installed invocation, and
a fresh accepted/released artifact.

## Dependency Order And Reviewable Slices

1. **Generated contract:** own `internal/assistantinstance/instance.go` and its
   tests; exit when a deployed builder is self-sufficient and retains every
   mutation prohibition.
2. **Fail-closed completion:** reconcile the preserved helper/test work in
   `bin/test-builder-completion` and `config/acceptance/builder-completion`; exit
   when all false-positive paths are rejected.
3. **Native exec driver:** update `bin/accept-capability-workshop`, focused
   helpers, and supervisor/runner tests; exit when prompt/event/process handling
   is bounded and private and installed-invocation PTY tests remain green.
4. **Live disposable proof:** run the real enriched builder without touching
   preserved roots; exit only on source plus exact inspect-ready/validate/test
   success and full cleanup.
5. **Reconciliation/docs:** update shipped architecture, development,
   deployment, runbook, and security text; remove this temporary plan and
   verify the exact implementation head.
6. **Release:** merge, nominate fresh bytes, run full independent acceptance,
   collect receipts, publish, download, and verify.

## Acceptance Traceability

Slices 1 and 4 prove a user can have the agent author the package from its own
builder contract. Slices 2 and 3 prove no model assertion becomes authority and
preserve the no-activation boundary. The unchanged #57 implementation covers
the remaining lifecycle, recovery, migration, preservation, partner, evidence,
and release criteria; `04-verification.md` defines the amended evidence.

## Documentation Promotion

- Update `docs/architecture/capability-workshop.md` with the complete builder
  information boundary, native-exec authoring, postcondition authority, and PTY
  invocation split.
- Update `docs/development.md` with deterministic builder/exec/completion tests.
- Update `docs/deployment.md` with the amended exact-candidate acceptance flow.
- Update `docs/runbook.md` with no-action/missing-contract diagnosis and private
  exec transcript handling.
- Update `SECURITY.md` only if its current credential/evidence boundary names
  PTY-only builder transport; no new ADR is needed because the merged amendment
  is the decision record and no durable subsystem/dependency is added.

## Pull Request And Review Contract

Resume the existing issue-56 implementation branch after this amendment merges.
Use failing-first commits, reconcile the final implementation against both #57
and this amendment, explain the superseded PTY-builder detail, update durable
docs from shipped behavior, and remove both temporary plan directories before
review readiness. Require `bin/ci`, targeted live-probe evidence, exact-head CI,
independent maintainer/security review, and a newly nominated post-merge
artifact. Neither issue closes at merge.

## Explicit Non-Goals And YAGNI Boundary

No scaffold command, second capability profile, general agent protocol adapter,
Codex fork, app-server client, remote-TUI bridge, Codex pre-release pin,
marketplace, public transcript, model-evaluated test, or change to lifecycle
authorization/evidence schemas beyond the builder transport facts required by
this amendment.

## Exceptions That Reopen Design

Return to Solution Design if the complete builder contract still cannot produce
source through native exec; a supported Codex change removes literal explicit
skill selection; exec requires a new credential/network/trust boundary; event
handling cannot remain private and bounded; candidate postconditions cannot
distinguish agent and driver authorship; or release would need rebuilt bytes.
