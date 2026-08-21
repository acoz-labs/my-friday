# Verification And Release Design

## Test Strategy

Add failing-first shell/fixture coverage beside the repository's existing
script tests:

- nomination workflow/config tests require an exact-commit arm64 build, named
  Actions artifact upload, and versioned authority output;
- packaging tests run twice with varied filesystem mtimes and assert identical
  archive SHA-256, one safe entry, normalized metadata, executable mode, and
  contained bytes matching the accepted fixture;
- release workflow tests prohibit build commands, require exact artifact
  download/digest verification, and pass explicit asset paths;
- finalizer fixtures cover new release, missing-asset retry, identical existing
  asset, mismatched asset refusal, partial publication, and exact checksum text;
- backfill fixtures cover verified source, non-latest tag, draft/prerelease,
  wrong archive digest, wrong executable digest, unsafe/multiple entries,
  identical alias, mismatched alias, and bounded alias removal.

Run `bin/ci`, `go test ./...`, shell lint/format checks already selected by CI,
and workflow static tests. Network-writing behavior is mocked until the
authorized exact-release operation.

## Red/Green Sequence

1. Fail authority parsing and artifact-transfer tests; implement nomination
   upload and exact release download without packaging.
2. Fail reproducibility/content tests; add deterministic archive and checksum
   generation.
3. Fail finalizer retry/mismatch tests; add digest-aware asset publication.
4. Fail backfill safety cases; add guarded current-release alias management.
5. Update durable documentation and run the complete CI suite before any public
   release mutation.

## Acceptance Evidence

Automated evidence binds the exact implementation SHA, accepted authority,
executable SHA-256, two independently produced archive SHA-256 results, archive
listing, and finalizer/backfill fixture results. Independent acceptance on Apple
silicon extracts the exact nominated candidate archive, verifies both checksum
layers, executes the contained binary's safe preview/help path, and records the
artifact-specific approval required by the release gate.

For the one-time public backfill, fresh evidence records the latest tag, source
asset ID/name/digest, checksum asset, contained executable digest, new alias
ID/digest, and a successful non-following request to the permanent URL. Values
are identifiers and digests, never credentials.

## Rollout

1. Merge the implementation only after CI and independent exact-head review.
2. Run the guarded backfill against the then-current latest release. Confirm the
   stable URL downloads the source-identical archive before `acoz` promotion.
3. Nominate the next exact successful commit using the new transfer authority;
   independently accept its downloaded artifact.
4. Release the exact accepted candidate. Verify the newest release contains the
   stable archive and checksums and that `/releases/latest/download/...` now
   resolves to it.

The execution envelope authorizes the guarded existing-release asset addition
and one normal exact-candidate artifact release; it does not authorize changing
historical source assets, tags, or accepted evidence.

## Rollback And Recovery

Code rollback is a Git revert, but published releases remain historical records.
If backfill uploaded an incorrect alias, the operator first proves the mismatch,
resolves its exact asset ID, deletes only
`my-friday-darwin-arm64.tar.gz`, and confirms the original archive and checksum
remain. A future release regression is corrected by a new accepted release; an
incorrect latest stable asset is removed until a verified replacement can be
published. Never delete or rewrite accepted source artifacts, tags, or ledgers.

## Release Prerequisites

- Current latest tag/assets and their GitHub SHA-256 metadata remain readable.
- The existing artifact release workflow has `contents: write`; no new secret
  slot is required.
- CI runner supports the repository's Go toolchain and reproducible tar/gzip
  commands chosen by implementation.
- Independent Apple silicon acceptance remains available.

## Production Readiness Preflight

The existing `GITHUB_TOKEN` secret injection and `contents: write` permission
are the only deployment/publication credential path. Implementation must
demonstrate exact
Actions artifact retrieval, executable digest verification, deterministic
packaging, retry-safe release lookup/upload, and dry fixture rollback before
the final implementation review. Activation is an explicit guarded workflow
dispatch after its named release preflight, never an implicit merge side effect.
Production verification and receipts are the workflow URL,
candidate SHA/authority, release tag, asset IDs/digests, checksum evidence, and
successful permanent-URL probe.
