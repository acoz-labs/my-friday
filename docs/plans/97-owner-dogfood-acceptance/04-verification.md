# Verification And Release Design

## Test Strategy

### Authority fixtures

Extend `bin/test-capability-workshop-evidence` or split one focused shell fixture
only if readability requires it. Preserve the existing fake `gh` double-fetch
and mutation controls. Add valid new owner receipt and two-part bundle cases,
then prove refusal for:

- old four-part bundle on the active issue-51 approval path;
- missing/extra/empty token parts and unknown prefixes;
- absent, empty, malformed, or nonmatching `PRODUCT_OWNER_ACTORS`;
- owner comment by an unauthorized actor;
- owner equal to evidence/product acceptor;
- owner or acceptor equal to any lifecycle-linked implementation PR author;
- missing/unmerged implementation PR or candidate excluding its merge;
- wrong issue, candidate, artifact, schema, marker, URL, role, types, or fields;
- edited/disappearing comment between fetches or wrong body digest;
- `completion`, comprehension, or redaction not boolean `true`;
- `recovery_needed` not boolean `false`;
- invalid retention, claim scope, migration issue/due state, or an assertion that
  independent validation was collected; and
- use of a controlled account as an “independent user” field or success claim
  (the strict schema has no such field).

Keep legacy verifier tests to prove historical tokens retain their exact old
meaning when invoked directly. They must not appear in current approval or
finalization parser fixtures.

### Acceptance and release contract

Extend `bin/test-acceptance-contract` and focused fixtures around
`record-product-acceptance`/`finalize-release` to prove:

- approval requires `capability-workshop-owner-dogfood-v1` while rejection
  still requires `capability-workshop-failure-v1`;
- both product-acceptance and release-artifact workflows inject
  `PRODUCT_OWNER_ACTORS`;
- retry parsing binds the exact new evidence token;
- finalization reparses and re-verifies the owner bundle rather than trusting a
  status/comment;
- old or generic evidence cannot authorize issue 51;
- artifact digest, nomination, implementation digest, issue, SHA, acceptor, and
  required checks remain bound; and
- successful no-rebuild packaging still emits the exact accepted executable,
  archive checksum, tag, Release ledger, and issue release record.

### Documentation and claim checks

Repository search/contract tests should ensure active documentation uses
“owner-operated dogfood” for this path, marks independent-user validation as
not collected, and no longer instructs readers to gather two partner receipts
for issue 51. Historical provenance may retain the old terminology when clearly
labelled audit-only.

No rendered UI exists. Terminal accessibility behavior and the workshop
scenario matrix are unchanged; no new screenshot or product-design review is
required.

## Red/Green Sequence

1. Add a failing active-path test showing a valid two-part owner bundle is
   rejected and the old four-part bundle is still accepted.
2. Add strict owner-receipt fixture tests, including owner allowlist, migration
   fields, edited comments, and all type/identity failures; implement recorder
   and verifier.
3. Add role/implementation-author separation failures; implement new bundle
   composition and lifecycle-author checks.
4. Add acceptance-command parser/retry failures; switch issue-51 success to the
   new prefix without changing failure authority.
5. Add release-finalizer and workflow-injection failures; implement symmetric
   release verification and required variable injection.
6. Add documentation claim checks; update durable documentation and help/usage.
7. Run targeted fixtures, `go test -race ./...`, and full native `bin/ci`.
8. Merge, nominate a fresh candidate, rerun the full native Apple-silicon
   workshop harness, obtain Anthony's exact-candidate receipt, approve, and
   release unchanged bytes.

## Acceptance Evidence

Implementation evidence must include:

- exact targeted authority-fixture commands and outcomes;
- complete `bin/ci` on the supported native Apple-silicon runner, including the
  existing APFS/sandbox/PTY workshop journey;
- no-diff or digest evidence proving accepted artifact bytes equal nominated
  and released bytes; and
- reconciliation showing the issue-51 lifecycle-linked implementation set
  contains the amendment merge.

Product acceptance requires one new candidate built after this implementation
lands. Against its exact SHA and `artifact-v1` identifier:

1. a non-owner authorized evidence/product acceptor runs the full existing
   capability-workshop matrix and publishes provisional/final evidence;
2. Anthony uses the same exact artifact, completes the bounded hands-on owner
   journey, reviews the automated evidence, and records the strict owner
   receipt with the issue-92 migration promise;
3. the acceptor dispatches approval with the new two-part token; and
4. the workflow proves owner/acceptor/implementer separation and writes exact
   bound statuses and issue evidence.

The existing candidate `2594147…` and evidence remain visible historical proof
but do not satisfy any step above. There is no migration evidence yet and no
independent-user evidence; acceptance and release summaries must say so.

## Rollout

1. Land the new authority alongside historical verifiers, update current-path
   parsers and documentation, and configure `PRODUCT_OWNER_ACTORS` before
   acceptance.
2. Link the implementation merge into both #97 provenance and issue #51's
   complete implementation set.
3. Run native CI and nominate a new immutable artifact from the merged commit.
4. Run fresh workshop evidence and owner acceptance against that artifact.
5. Release exactly those bytes through the existing artifact workflow, with
   #51 and #97 represented in the release lifecycle as required by the final
   implementation association.
6. Verify release tag, GitHub Release assets, `SHA256SUMS`, issue acceptance and
   release records, and public copy's owner-dogfood scope.
7. Unblock #92. Its own acceptance later captures actual Alfred migration
   evidence before #92 release.

There is no feature flag or staged service. Activation is the new immutable
GitHub Release and the current issue-51 acceptance parser.

## Rollback And Recovery

Before publication, a failed candidate is abandoned through ordinary corrective
PR and renomination; no acceptance evidence is deleted. After publication, keep
the previous immutable release available and correct the default with a new
forward commit, candidate, acceptance, and release. Never retag, overwrite
assets, edit receipts, reinterpret the old token, or accept a rebuilt binary.

A partial owner-recorder attempt may be retried by producing a new comment;
only the token whose digest and body verify is authority. Acceptance or release
workflow retry is idempotent only for the exact same SHA, artifact,
implementation digest, acceptor, and evidence bundle.

## Release Prerequisites

- Planning PR merged and issue #97 Ready under the repository verifier.
- `PRODUCT_OWNER_ACTORS` configured and injected in both workflows.
- Implementation PR merged, linked to #97 and added to issue #51's complete
  lifecycle implementation set.
- Full CI succeeds on the exact merged commit.
- A fresh post-amendment candidate is nominated; prior evidence is not reused.
- New workshop evidence and Anthony's owner receipt pass independently.
- Acceptance/release summaries state the owner-operated validation scope and
  do not claim independent-user evidence.
- Previous release remains available for rollback.

## Production Readiness Preflight

Production is the public immutable GitHub artifact release; no service staging
or runtime secret exists.

- **Configuration:** required non-secret variable `PRODUCT_OWNER_ACTORS` is
  read by product acceptance and artifact release. `ACCEPTANCE_ACTORS` and
  `CI_RUNNER` remain unchanged.
- **Permissions:** existing workflow permissions suffice; no new secret,
  environment, write scope, or external service is introduced.
- **Build/injection:** the candidate is built once; owner configuration is
  injected only into control verification, never into the artifact.
- **Activation:** explicit product acceptance followed by artifact release.
- **Verification:** exact SHA/artifact, checks, implementation set, evidence,
  owner receipt, actor separation, statuses, archive/executable digests, tag,
  Release, assets, and ledger are all executable checks.
- **Rollback:** retain prior immutable release and publish only a corrected
  forward candidate; no overwrite or evidence rewrite.
- **Receipt:** the existing product-acceptance marker and release ledger gain
  the exact new bundle while preserving issue/SHA/artifact/implementation/
  acceptor binding.

Result: ready for the `through-production` envelope once the named release
prerequisites pass.
