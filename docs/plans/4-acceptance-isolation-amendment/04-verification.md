# Verification And Release Design

## Test Strategy

Implementation begins with failing tests for the supervisor's pure parsers and
state machine, then exercises real macOS primitives on a disposable image, and
ends with independent exact-candidate acceptance.

Likely ownership:

- `bin/accept-installed-codex-baseline` and focused shell/unit helpers;
- tests/fixtures for artifact authority, `hdiutil -plist`, mount tables, sandbox
  profiles/diagnostics, markers, manifests, evidence, and cleanup;
- `bin/record-product-acceptance`, `bin/release-gate`,
  `.github/workflows/product-acceptance.yml`, and their contract tests for
  evidence binding;
- `docs/deployment.md` and `docs/runbook.md` for operator behavior.

Deterministic tests cover:

- exact `artifact-v1` parse, run/name/ID/digest mismatch, archive shape, and
  digest recheck;
- canonical-home descendant checks, unique absent root creation, symlink/type/
  owner/mode/link/device/inode/marker refusals, and fd-relative cleanup;
- program-readable attach/mount parsing, APFS/local/owner/options enforcement,
  extra-device refusal, detach state, and restart discovery;
- profile generation/escaping/digest, volume-only candidate writes,
  evidence-staging denial, reviewed warning allowlist, unexpected
  diagnostic/nonzero refusal, and lifecycle-versus-smoke network rules;
- protected metadata traversal, exclusions, race/error refusal, aggregate
  comparison, canary proof, and evidence redaction;
- every lifecycle fixture transition and source-freshness-independent
  rollback/uninstall contract;
- observed-journal interruption, PID/process-group targeting, recovery-required
  state, and ordinary recovery;
- auth stdin handling, image-local `auth.json` classification, logout/cleanup,
  and proof that logs/evidence contain no secret or raw provider response;
- issue evidence creation/fetch, author/candidate/artifact/schema/body-digest
  binding, idempotent retry, edit/delete/mismatch rejection, and release gate;
- failures before/after attach, detach failure, publication failure, protected
  mismatch, unexpected entries, and identity drift preserving evidence.

Host integration tests create a tiny sparse APFS image under a disposable test
root and run the actual sandbox positive controls without credentials. They are
macOS-gated, never use the real `HOME`/`CODEX_HOME`, use their own non-sensitive
canary, and must prove detach and zero residue. CI retains platform-independent
parser/state tests; a suitable macOS job runs primitive integration tests.

## Red/Green Sequence

1. Lock current artifact/acceptance/release behavior with characterization tests.
2. Add failing marker/path/identity and protected-manifest tests; implement the
   pure supervisor state model.
3. Add failing APFS output and cleanup tests; implement create/attach/verify/
   ordinary-detach/marker-bound-delete behavior.
4. Add failing profile and diagnostic tests; implement fixed lifecycle/smoke
   profiles, candidate denial of evidence staging, controls, and no-fallback
   refusal.
5. Add failing exact-artifact and environment tests; implement download/copy/
   digest recheck and minimal synthetic environment.
6. Add failing lifecycle and journal-interruption scenario tests; implement the
   complete matrix using only the production candidate interface.
7. Add failing Codex auth/redaction tests; implement login/smoke/logout without
   exposing sensitive disposable state.
8. Add failing evidence mutation/release tests; implement issue-comment ID/body
   digest binding in acceptance and release authority.
9. Run repository CI, macOS primitive integration, a contributor rehearsal with
   a nominated test artifact, then independent acceptance on the fresh final
   nominated candidate.

## Acceptance Evidence

Contributor tests and rehearsals are development evidence only. Product
acceptance is performed by someone other than the sole implementer on the exact
newly nominated Darwin/ARM64 artifact. The acceptor records:

- candidate SHA, full `artifact-v1` authority, archive and executable digest;
- implementation and acceptance-harness commits;
- hardware architecture, macOS build, APFS and Git versions, and Codex version;
- run schema/ID, profile digests, attach/device/volume facts, and all control
  results;
- scenario-by-scenario pass/fail with expected state and exit classification;
- protected manifest aggregate before/after digests/counts/equality and canary
  results;
- real-Codex installed-instruction discovery result without prompt/response or
  auth material;
- child termination, non-forced detach, mount absence, exact cleanup, and
  protected-state postcondition;
- redaction/sanitization assertion and evidence comment ID/body digest.

The acceptor must inspect the sanitized record before approval. The product-
acceptance workflow independently fetches and verifies it; a pasted URL or
opaque statement is insufficient. Since this is terminal behavior with no
rendered UI, `docs/operations/ui-acceptance.md` classification is
`no-rendered-impact`.

The acceptance boundary explicitly does not evidence a distinct UID, fresh
login keychain/session, account teardown, read confidentiality, or malicious
same-UID resistance. The evidence record repeats those limitations.

## Rollout

1. Merge this exact approved amendment plan.
2. Implement the supervisor, evidence binding, tests, and durable docs in one
   follow-up issue #4 implementation PR; reconcile against PR #15 plus this
   amendment and append the PR to the issue's implementation lifecycle.
3. Merge only after required CI, macOS primitive integration, independent
   review, and reconciliation. The prior nomination is stale for acceptance.
4. Run artifact nomination on the new main commit, producing one new exact
   Darwin/ARM64 archive and `artifact-v1` authority.
5. Independent acceptance downloads those bytes, runs the complete amended
   boundary/matrix, publishes sanitized evidence, verifies cleanup, and records
   the decision bound to the evidence comment.
6. On approval, release downloads the same run/name artifact, re-verifies the
   executable digest and still-valid evidence authority, and uploads those exact
   bytes to the GitHub Release. Existing mismatched assets fail closed.
7. Verify the release receipt, asset digest, acceptance status, issue lifecycle,
   and closure under repository policy.

No configuration flag activates a weaker path. The old disposable-account
instructions are replaced in durable operator docs only when the implementation
ships.

## Rollback And Recovery

Before acceptance, failure releases nothing and leaves issue #4 awaiting a new
candidate. A harness failure is repaired through a new implementation merge and
new nomination; acceptance never carries across candidate digests.

During a run, ordinary non-forced detach and marker-bound cleanup are attempted
only after quiescence and proof. Ambiguous mounts, images, markers, evidence, or
protected-state differences are retained privately for runbook diagnosis. The
recovery runbook never force-detaches automatically, removes an unmarked path,
or modifies the real Codex installation to make a test pass.

After release, artifact rollback follows the repository's immutable-release
policy: publish a corrected later version rather than replacing accepted bytes.
The installed baseline itself retains the PR #15/PR #19 verify, repair,
rollback, recover, and uninstall contracts; this amendment changes only how
those contracts are accepted.

## Release Prerequisites

- The exact sandbox templates and diagnostic allowlist pass review and actual
  positive controls on the supported macOS build.
- A repository-supported macOS job can execute APFS/sandbox primitive tests
  without administrator authority.
- The local acceptor has GitHub read/comment authority and an approved secret
  source for a test-only provider key; no secret value is stored in the repo.
- Evidence comment fetch/digest/author binding and release revalidation ship
  before product acceptance.
- Durable docs explicitly replace the old disposable-account path and retain
  the same-UID/keychain/read-boundary disclosure.

## Production Readiness Preflight

- **Secret slot/source:** one operator-approved test-only `OPENAI_API_KEY` source
  is injected locally into stdin for `codex login --with-api-key`. The value is
  never a workflow input, argument, log, evidence field, or committed fixture.
  Temporary `auth.json` remains sensitive image-local state and is destroyed.
- **Artifact transport:** nomination builds once and records run/artifact/name/
  digest; acceptance and release download that exact artifact and independently
  verify the executable SHA-256.
- **Host path:** the no-admin command, required binaries, synthetic environment,
  profile preflight, controls, full matrix, evidence publication, and cleanup
  are executable on supported Apple Silicon macOS.
- **Acceptance authority:** issue #4, current lifecycle-linked implementation
  set, candidate, artifact, acceptor, evidence comment ID, and body digest are
  mutually bound and revalidated at release.
- **Activation:** no service deployment exists. GitHub Release publication is
  the production action after exact-candidate approval.
- **Rollback:** failed or ambiguous acceptance publishes no release; published
  immutable assets are superseded, never silently replaced.
- **Receipt:** release records candidate SHA, complete artifact authority,
  executable digest, issue/implementation set, acceptance/evidence authority,
  workflow/run identity, tag, asset URL, and post-upload digest verification.
