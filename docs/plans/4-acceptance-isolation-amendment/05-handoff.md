# Implementation Handoff

## Change Tier And Smallest Complete Outcome

This is a high-risk operational/security-boundary change with no product UI.
The smallest complete outcome is a reviewed no-admin macOS supervisor that
accepts the exact nominated installed-baseline artifact inside an APFS/sandbox
boundary, publishes tamper-evident sanitized evidence, proves cleanup and live-
state preservation, binds that evidence through release, and replaces the old
disposable-account operator contract.

The execution envelope is `through-production`: implementation, fresh
nomination, independent exact-candidate acceptance, same-byte GitHub Release,
receipt verification, and issue lifecycle completion are one coherent outcome.

## Dependency Order And Reviewable Slices

1. **Characterization and pure boundary model**
   - Ownership: acceptance script tests/fixtures and helper modules.
   - Exit: artifact, path, marker, identity, manifest, and failure-state tests
     fail usefully before implementation and preserve current release contracts.
2. **No-admin APFS lifecycle and cleanup**
   - Ownership: `bin/accept-installed-codex-baseline`, APFS fixtures/integration.
   - Exit: create/attach/verify/detach/delete works without admin and every
     mismatch preserves exact marked state.
3. **Fail-closed sandbox and protected-state proof**
   - Ownership: reviewed profile templates, diagnostic parser, controls,
     manifest/canary implementation and macOS integration tests.
   - Exit: allowed writes work, protected writes/lifecycle network fail,
     smoke network works, unexpected semantics refuse, and pre/post equality is
     proven without secret-content reads.
4. **Exact candidate and lifecycle matrix**
   - Ownership: artifact downloader/environment builder, fixtures, PTY runner,
     journal observer, process supervision.
   - Exit: unmodified digest-verified bytes pass the complete lifecycle and
     externally induced interruption/recovery matrix on the disposable volume.
5. **Isolated real-Codex smoke and sanitization**
   - Ownership: local secret injection, Codex login/discovery/logout adapter,
     redaction tests.
   - Exit: image-local auth enables one discovery proof, never appears in
     evidence/logs, and teardown removes it.
6. **Durable evidence and release authority**
   - Ownership: issue-comment publisher, `bin/record-product-acceptance`,
     `bin/release-gate`, acceptance/release workflows and contract tests.
   - Exit: comment ID/body digest/author/candidate/artifact bind acceptance and
     release; edit/deletion/mismatch fails closed.
7. **Documentation, reconciliation, nomination, acceptance, release**
   - Ownership: architecture/deployment/runbook/SDLC docs, PR reconciliation,
     exact artifact workflows and release receipt.
   - Exit: temporary plan removed after promotion, new candidate accepted by an
     independent acceptor, same bytes released, receipt verified, issue closed.

## Acceptance Traceability

| Acceptance group | Slice | Required evidence |
| --- | --- | --- |
| No-admin APFS boundary and cleanup | 2 | Unit plus real macOS primitive integration |
| Sandbox/network/protected-state containment | 3 | Parser tests, positive controls, manifest equality |
| Exact bytes and full lifecycle/recovery | 4 | Automated fixtures and fresh nominated-artifact matrix |
| Real Codex instruction discovery | 5 | Redacted independent smoke result and auth teardown |
| Tamper-evident acceptance/release | 6 | Workflow contract tests and comment mutation denials |
| Through-production completion | 7 | Evidence authority, accepted status, release asset/receipt digest |

Detailed cases and acceptance limitations are in `04-verification.md`.

## Documentation Promotion

| Destination | Promoted contract |
| --- | --- |
| `docs/architecture/installed-codex-baseline.md` | Acceptance containment/evidence relationship to installed lifecycle |
| `docs/decisions/0002-manifest-owned-codex-baseline.md` or a scoped successor ADR | Why same-UID APFS/sandbox write containment is proportionate and its limits |
| `docs/deployment.md` | Fresh nomination, local exact-artifact acceptance, evidence binding, same-byte release |
| `docs/runbook.md` | Preconditions, command, controls, secrets, failure preservation, crash recovery, exact cleanup |
| `docs/operations/sdlc.md` if generic behavior changes | Evidence comment ID/body digest acceptance and release authority |
| CLI/help where appropriate | Supported platform/refusal and exit classifications |

Reconciliation decides the smallest durable destinations from shipped behavior.
It must remove this temporary amendment pack together with the implementation
PR's plan cleanup while leaving PR #15 and this merged planning PR as historical
authority.

## Pull Request And Review Contract

- Branch from the exact approved amendment merge on current `main`; use one
  issue #4 follow-up implementation PR and append it to the existing Phase 3
  lifecycle without removing PR #19.
- Begin with failing tests and keep APFS/sandbox, acceptance authority, and docs
  slices independently reviewable. No implementation work enters this planning
  PR.
- Require container/CI, macOS primitive integration, shell/security review,
  contributor rehearsal, exact-head reconciliation, and independent maintainer
  review before merge.
- The draft implementation PR explains reconciliation against original PR #15
  plus this amendment, the stale prior nomination, documentation promotion,
  secret handling, same-UID limitations, and recovery behavior.
- After merge, nominate a fresh candidate. The implementer cannot be its sole
  acceptor. Do not mark issue #4 complete until acceptance, same-byte release,
  receipt verification, and lifecycle closure finish.

## Explicit Non-Goals And YAGNI Boundary

Do not add user/account creation, administrator helpers, PF/DTrace, a daemon,
privileged helper, container/VM platform, generalized sandbox framework, secret
manager, arbitrary evidence store, broad filesystem snapshotter, live-home
repair, test-only candidate build, or production fault switch. Do not claim
distinct UID, fresh keychain, read confidentiality, or general malware
containment. Do not alter unrelated installed-baseline lifecycle decisions.

## Exceptions That Reopen Design

Return to Solution Design if the supported macOS build cannot enforce and
positive-control the reviewed sandbox without admin; APFS attachment requires
elevation; exact production-byte interruption cannot be safely observed;
protected manifests require reading secret contents; standard-provider Codex
cannot authenticate entirely inside the disposable volume; evidence cannot be
bound immutably enough for release; or acceptance needs a distinct UID,
keychain, VM, privileged helper, broader secrets, or a different execution
envelope.
