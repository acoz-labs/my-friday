# Solution Design: Stable latest-download asset

- **Status:** Draft
- **Issue:** #17
- **Planning PR:** #23
- **Repository basis:** e0f7c685ec7a0339cb42d945a02b40e4f513b4f2
- **Execution envelope:** through-production

## Decision

Make the release lane transfer the accepted Apple silicon executable into a
deterministic `my-friday-darwin-arm64.tar.gz`, publish that archive under the
same filename on every release, and publish checksums that separately prove the
accepted executable and distribution archive. Add a guarded, retry-safe
backfill command for the current latest release so the permanent GitHub
`latest/download` URL begins working without rebuilding anything.

## Needs Attention

- The issue's phrase “stable asset bytes are exactly the accepted artifact” is
  technically impossible when the accepted artifact is the raw executable and
  the public asset is a gzip-compressed tar archive. The plan preserves the
  approved invariant as: the contained executable is byte-identical to the
  accepted artifact, while the stable archive is byte-identical to any retained
  commit-suffixed archive. Both SHA-256 values are published unambiguously.
- The one-time backfill mutates the current public GitHub Release. It is allowed
  only after exact source-archive and contained-executable verification and has
  a bounded removal recovery path.

## Decision Spotlight

- **Artifact authority:** the existing accepted executable digest remains the
  ledger authority; packaging never rebuilds or transforms that executable.
- **Stable filename:** every future release uploads exactly
  `my-friday-darwin-arm64.tar.gz`; the release tag supplies version identity.
- **Checksum semantics:** `SHA256SUMS` labels archive and executable entries
  explicitly so users and operators cannot confuse the two layers.
- **Current release:** an explicit guarded backfill copies verified existing
  archive bytes to the stable name; it does not create a new product build.
- **Retry behavior:** an absent alias is created, an identical alias is a
  success, and a mismatched or ambiguous alias fails closed.

## Plan Map

- [Context](01-context.md)
- [Decision](02-decision.md)
- [Design](03-design.md)
- [Verification](04-verification.md)
- [Implementation handoff](05-handoff.md)

## Final Gate

The plan becomes `Final` only after independent maintainer review, the planning
PR and exact head are recorded, and Anthony approves that exact head for the
`through-production` envelope, including the bounded current-release mutation.
