# Implementation Handoff

## Change Tier And Smallest Complete Outcome

This is a high-risk Tier 3 release-integrity change because it publishes
executable bytes and mutates an existing public release. The smallest complete
outcome is a verified stable alias on the current latest release plus an
automated exact-byte path that supplies the same filename on every future
release.

## Dependency Order And Reviewable Slices

1. **Immutable transfer:** failing authority/workflow tests, nomination build
   and upload, strict authority, exact release download, and executable digest
   verification. Exit when accepted bytes move without a release rebuild.
2. **Distribution format:** reproducible packaging and explicit two-layer
   checksums. Exit when repeated packaging is byte-identical and archive content
   matches the accepted executable.
3. **Publication safety:** finalizer asset upload plus missing/equal/mismatch and
   partial-retry behavior. Exit when no normal path silently overwrites an
   asset.
4. **Current-release backfill:** guarded command/workflow, fixture coverage,
   dry evidence, authorized execution, and permanent-URL probe. Exit when the
   current latest release exposes the verified stable alias.
5. **Normal release proof:** nominate, accept, and release the exact candidate
   through the governed artifact lane. Exit when the next latest release also
   satisfies the stable URL contract without manual renaming.

## Acceptance Traceability

Future stable naming and no-rebuild criteria map to slices 1–3; executable and
archive digest clarity maps to slice 2; retry/fail-closed behavior maps to slice
3; current latest coverage maps to slice 4; operational documentation and
end-to-end permanence map to slices 4–5. Detailed failure cases and evidence
identities are in `04-verification.md`.

## Documentation Promotion

Update `docs/deployment.md` with artifact authority, packaging, checksum,
backfill, verification, and rollback contracts. Update `docs/operations/sdlc.md`
only if the generic artifact-profile contract changes rather than this product's
adapter. Add an ADR only if implementation must alter the repository-wide
meaning of an accepted artifact, which would first reopen design.

## Pull Request And Review Contract

Use one implementation PR linked to #17. Require failing-first fixture evidence,
full CI, security/release review of all untrusted archive and GitHub API inputs,
and independent exact-head acceptance before any public write. Promote durable
documentation and remove this temporary plan before leaving draft. The backfill
run and normal release must each be tied to explicit immutable receipts.

## Explicit Non-Goals And YAGNI Boundary

Do not add other platforms, installers, signing/notarization, auto-updates,
package managers, CDN/proxy infrastructure, a GitHub release API on `acoz.dev`,
semantic versioning changes, or historical asset renaming. Do not rebuild during
release or replace the accepted executable digest with the archive digest.

## Exceptions That Reopen Design

Return to Solution Design if GitHub cannot provide a trustworthy exact Actions
artifact identity, deterministic packaging cannot be achieved on the runner,
the accepted artifact must change from executable to archive, the current source
archive fails verification, a mismatched stable alias already exists, or safe
publication requires broader credentials or irreversible historical changes.
