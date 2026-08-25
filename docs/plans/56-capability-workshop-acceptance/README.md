# Solution Design: Exact-Candidate Capability Workshop Acceptance

- **Status:** Draft
- **Issue:** #56
- **Planning PR:** Pending
- **Repository basis:** 8e2371b433f4f6e4f28fe5c3491cc40b697d680b
- **Execution envelope:** `through-production`

## Decision

Add a separate issue-51 acceptance supervisor and typed evidence contract that
reuse the existing immutable-artifact, APFS isolation, file-backed Codex
credential, and finalization machinery. The new journey proves the complete
capability workshop against one nominated executable. Existing issue-4 evidence
remains valid only for issue 4 and cannot satisfy issue 51.

The already nominated `8e2371b...` artifact is historical because it predates
the acceptance implementation. After this implementation merges, a new exact
main commit and newly built artifact must be nominated and accepted.

## Needs Attention

- Acceptance requires the existing current-user-owned mode-0600 Codex
  `auth.json`, supplied by absolute path and never read into argv or evidence.
- Independent hands-on acceptance requires Apple silicon macOS, APFS,
  `sandbox-exec`, GitHub access, and a bounded live Codex session.

## Decision Spotlight

- **Separate supervisor:** preserve issue 4 unchanged; issue 51 receives its own
  command, schemas, and verifier routing so evidence cannot cross-authorize.
- **One immutable artifact:** every lifecycle and live-Codex observation uses
  the nominated executable digest; no rebuild occurs during acceptance.
- **Source-first proof:** the live agent may edit only disposable runtime
  source and run read-only checks. Lifecycle confirmation tokens are supplied
  by the acceptance driver, never by the builder agent.
- **Redacted evidence:** record paths, declarations, diffs by digest and a
  secret-free behavioral summary; never publish instruction bodies, auth bytes,
  foreign collision content, prompts beyond fixed fixtures, or private paths.
- **Destructive cleanup boundary:** cleanup uses marker/manifest authority,
  no-follow identity proofs, ordinary detach, and exact run roots; ambiguity is
  preserved and reported rather than recursively removed.
- **Product acceptance:** one independent authorized acceptor may record the
  repository gate; issue #51's product-owner plus two design-partner workshop
  receipts are separate typed evidence inputs, not GitHub labels or prose.

## Plan Map

- [Context](01-context.md)
- [Decision](02-decision.md)
- [Design](03-design.md)
- [Verification](04-verification.md)
- [Implementation handoff](05-handoff.md)

## Final Gate

The plan becomes Final only after the planning PR number is recorded, plan
validation and maintainer review pass, and Anthony approves the exact final
head with the `through-production` envelope.
