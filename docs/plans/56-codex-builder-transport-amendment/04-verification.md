# Verification And Release Design

## Test Strategy

- Extend `internal/assistantinstance/instance_test.go` with a failing assertion
  that a generated builder contains the complete three-file package contract,
  strict field vocabulary, cross-file invariants, prohibited effects, and no
  lifecycle mutation authority.
- Replace builder-marker cases in `bin/test-launcher-pty-capture` and
  `bin/test-capability-workshop-supervisor` with native-exec argv/stdin safety,
  private event-stream, timeout, nonzero, malformed/empty event, no-action,
  no-source, non-ready, validate-failure, test-failure, and postcondition-success
  cases. Retain all PTY tests for installed invocation.
- Keep the focused adversarial `bin/test-builder-completion` gate: no model
  completion or zero exit may create `BUILDER_SOURCE_READY` without all exact
  source and candidate postconditions.
- Run `bin/ci`, Go race tests, acceptance contracts/evidence, shell syntax, and
  a disposable live native-exec probe before opening the implementation PR.
- Run the full Apple-silicon/APFS supervisor only against a newly nominated
  post-merge candidate.

## Red/Green Sequence

1. Lock the current generated builder's missing-contract failure.
2. Add failing generated-skill contract assertions, then enrich the template.
3. Add failing native-exec driver and event/postcondition tests.
4. Implement bounded native exec while preserving auth, sandbox, and runner
   contracts.
5. Make the disposable live probe create `daily-brief` and pass exact candidate
   inspect, validate, and test.
6. Integrate the fail-closed receipt with the full supervisor and rerun CI/fault
   matrices.
7. Merge, nominate a fresh artifact, and rerun independent acceptance from the
   beginning.

## Acceptance Evidence

Deterministic PR evidence proves the builder text contract, invocation
construction, event classification, every false-completion denial, confinement,
and cleanup. The disposable pre-PR live probe proves that the actual installed
Codex can use the enriched builder to author the fixed package; it is diagnostic
evidence, not release authority.

Independent exact-candidate acceptance proves native-exec authorship and
candidate postconditions, then the unchanged fresh PTY installed invocation and
remaining issue-51 scenario matrix. Public evidence records only typed scenario
facts and the redacted complete-diff digest. Product-owner and two distinct
design-partner receipts remain required.

This change has no rendered graphical interface. The terminal interaction is a
plain-text protocol and follows the existing terminal acceptance contract; no
screenshot baseline applies.

## Rollout

The execution envelope remains `through-production`:

1. merge the reviewed amendment before resuming implementation;
2. implement and reconcile the exact amendment, including the disposable live
   green proof;
3. merge the reviewed implementation and mark candidate 13 historical;
4. nominate one new immutable artifact from merged main;
5. run the complete independent issue-51 supervisor and verify final evidence;
6. record partner and product acceptance for the exact candidate;
7. publish the already-built artifact through `release-artifact.yml`;
8. freshly download and verify checksums, executable digest, version/help, and
   stable URL; and
9. complete issues #51 and #56 through the repository release workflow.

## Rollback And Recovery

Before release, revert the builder/acceptance implementation and nominate new
bytes; do not reuse a prior artifact. After release, supersede a bad artifact
with a newly accepted release. The generated builder is instance-manifest-owned,
so existing instances receive it only through the already designed explicit
instance upgrade/rollback path. Source remains user-owned and is never removed.

Interrupted or failed live acceptance follows the existing marked-run recovery
contract. Raw exec transcripts and auth copies are removed only under exact
run-owned authority; ambiguity is preserved for diagnosis.

## Release Prerequisites

- Current-user mode-0600 one-link Codex auth file injected only by absolute path.
- Apple silicon/APFS independent acceptance host.
- Successful disposable native-exec proof using the implementation's generated
  contract before nomination.
- New merged candidate and artifact authority; all 13 prior failed candidates
  remain historical.
- Independent acceptor, one product-owner receipt, two distinct design-partner
  receipts, green CI, and no unresolved P0/P1 finding.

## Production Readiness Preflight

- **Secrets:** no new slot; existing auth-by-path copy remains no-content and
  private.
- **Candidate:** exact main SHA, artifact run/id/name/digest, generated builder
  digest, native-exec helper closure, and all implementation PRs are bound.
- **Deploy:** this artifact repository has no staging service; the production
  action is the existing `release-artifact.yml` publication of accepted bytes,
  with no rebuild or automatic user-root activation.
- **Activation:** builder cannot activate; only driver-confirmed lifecycle
  commands mutate the disposable instance.
- **Verification:** issue-51 evidence verifier, product acceptance,
  `release-artifact.yml`, fresh download, checksum/executable digest,
  version/help, and stable URL remain executable.
- **Rollback:** explicit instance rollback and artifact supersession remain
  bounded; source is preserved.
- **Receipt:** final evidence and GitHub Release bind the amended planning PR,
  implementation set, candidate, artifact, partner receipts, and included
  issues.
