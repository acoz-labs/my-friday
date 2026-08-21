# Context

## Problem And Desired Outcome

[Issue #17](https://github.com/acoz-labs/my-friday/issues/17) provides the
release-owned half of the public landing page: one URL that always resolves to
the newest accepted Apple silicon download. `acoz.dev` can then link directly
to GitHub Releases without an API call or a site edit for every version.

## Current State

At repository basis `e0f7c685ec7a0339cb42d945a02b40e4f513b4f2`:

- `.github/workflows/nominate-artifact.yml` verifies an operator-supplied
  immutable artifact identifier but does not build or upload candidate bytes.
- `.github/workflows/release-artifact.yml` verifies release authority and calls
  `bin/finalize-release`; it does not download an accepted asset.
- `bin/finalize-release` requires `RELEASE_ASSET_PATH` for artifact execution,
  but the workflow does not supply it and the finalizer does not currently add
  that path to `gh release create`.
- `docs/deployment.md` correctly establishes that nomination accepts one raw
  Apple silicon executable and release must publish those bytes rather than a
  rebuild.
- The current latest release is `artifact-2026.08.21-5bc3092`. It contains
  `my-friday-darwin-arm64-5bc3092.tar.gz` and `SHA256SUMS`, but no stable-named
  archive. Its release ledger records the accepted executable digest separately
  from the GitHub-reported archive digest.

## Actors And Critical Journeys

- A public visitor follows GitHub's permanent latest-download URL and receives
  a tar archive containing the accepted executable.
- An acceptor exercises the exact nominated executable and records approval
  against its immutable artifact authority.
- The release workflow retrieves those exact bytes, verifies them, packages
  deterministically, publishes the stable archive/checksums, and safely retries
  an interrupted release.
- An authorized operator backfills the current latest release only after
  proving the retained archive and its contained executable; a mismatched alias
  is diagnosed and removed without deleting source evidence.

## Acceptance And Non-Goals

The desired product outcome is preserved with one explicit semantic correction:
the compressed archive cannot itself be byte-identical to the raw accepted
executable. The contained executable must match the accepted digest; the stable
archive must match any commit-suffixed archive byte-for-byte. This distinction
is required for a testable checksum contract and is presented for Gate 2
approval.

Non-goals remain page hosting, GitHub API behavior on `acoz.dev`, new platforms,
installers, signing/notarization, and changing the executable accepted by the
existing artifact ledger.

## Constraints, Dependencies, And Risks

- Release publication must never invoke `go build`; it consumes only nominated
  bytes and fails closed when provenance or digest differs.
- GitHub's `/releases/latest/download/<name>` contract depends on every newest
  non-draft, non-prerelease release containing the stable filename.
- `gh release upload --clobber` could overwrite a mismatched public asset and is
  prohibited in the normal path.
- Current-release backfill is an externally visible mutation and must retain
  the original commit-suffixed archive and checksum evidence.
- GitHub Actions artifacts have finite retention, so accepted-byte transfer must
  complete before expiration; the release remains the durable public record.

## Evidence, Assumptions, And Unknowns

Verified evidence includes the current workflows/finalizer, deployment contract,
and live latest-release asset names and SHA-256 metadata. The design assumes the
current retained archive contains one executable at a known path; backfill
preflight must verify rather than trust that assumption. No blocking product
unknown remains after the archive/executable distinction is made explicit.
