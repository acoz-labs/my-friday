# Discovery: public landing page and stable download surface

- **Status:** Final candidate
- **Discovery issue:** #13
- **Discovery PR:** #14
- **Repository basis:** 5bc309226d2c40e1473a4011c1bd8552c995919d
- **Recommended decision:** approve
- **Gate 1:** awaiting-authority
- **Confidence:** High
- **Private evidence:** none

## Decision sought

Decide the smallest trustworthy public landing surface for My Friday: its
canonical public hostname and host, the one-page content boundary, and a
release-backed download link that does not require a website edit for every
accepted release.

## Audience and critical tasks

The first audience is a technically capable macOS user who is curious about a
local-first, inspectable Codex assistant but does not yet know My Friday. They
need to:

1. understand the product in one screenful;
2. confirm that the current pilot supports their machine;
3. understand the important safety and distribution caveats;
4. download the latest accepted artifact without choosing among release assets;
5. reach installation, checksum, source, licence, and release-note details.

Existing users and contributors remain better served by GitHub. The landing
page should route them there rather than duplicating a changelog or
documentation site.

## Evidence

- My Friday is a public, MIT-licensed repository with its first accepted Apple
  silicon artifact already published through GitHub Releases.
- GitHub documents a stable latest-release asset URL of the form
  `/releases/latest/download/<asset-name>`, provided that the latest release
  contains that exact asset name. My Friday's current archive instead includes
  a commit suffix, so its filename is not yet a stable web contract. See
  [GitHub's release-link documentation](https://docs.github.com/en/repositories/releasing-projects-on-github/linking-to-releases).
- `acoz.dev` already uses Cloudflare for public delivery, and the same account
  already operates an Acoz Labs Cloudflare Pages project. Cloudflare Pages
  supports custom subdomains and direct upload from CI. See the official
  [custom-domain](https://developers.cloudflare.com/pages/configuration/custom-domains/)
  and [direct-upload](https://developers.cloudflare.com/pages/get-started/direct-upload/)
  documentation.
- A live DNS check found no existing record for `my-friday.acoz.dev`.
- The current accepted artifact supports macOS 14 or later on Apple silicon and
  is ad-hoc signed but not notarized. Hiding that caveat would make a prominent
  download button misleading.

## Assumptions

- `acoz.dev` is the durable public domain for Acoz Labs product surfaces.
- The hyphenated hostname is acceptable because it matches the repository name
  and unambiguously preserves the product's two-word name.
- GitHub Releases remains the authority for accepted binaries, checksums,
  release notes, and rollback history.
- The page needs no accounts, analytics, runtime API, content management, or
  client-side application framework.

## Unknowns

- The final visual treatment and whether My Friday needs a product mark beyond
  a wordmark belong in product design.
- Solution Design must decide how the Pages deployment participates in the
  repository's acceptance and production controls without weakening its
  artifact-release contract.
- Signing and notarization may improve later; the page needs a durable way to
  state current distribution truth without inventing a second release ledger.

## Competing options

### A. Cloudflare Pages at `my-friday.acoz.dev` — selected

Keep the static page with the My Friday product, use the established Acoz Labs
Pages hosting pattern, and attach the custom subdomain in the existing
Cloudflare zone. This preserves product ownership, keeps the personal
`acoz.dev` application independently releasable, and leaves room for previews
or redirects without requiring them now.

### B. GitHub Pages at `my-friday.acoz.dev` — rejected

GitHub Pages would serve the static page adequately and keep hosting beside the
source and releases. It is not selected because Cloudflare Pages is already an
operated Acoz Labs pattern and keeps custom-domain delivery in the platform
that owns the zone. Reconsider only if the Pages release path proves materially
more complex than GitHub Pages during Solution Design.

### C. Add a route or hostname to the existing `acoz.dev` application — rejected

This avoids a new site deployment, but it couples a product landing page to
Anthony's personal-site repository and release cadence. The My Friday product
should own its public surface.

### D. Resolve the latest asset dynamically at the web edge — rejected

A Worker or browser request could inspect the GitHub API and redirect to the
current commit-suffixed archive. That adds runtime logic and a failure mode for
something GitHub already supports when the release asset name is stable.

## Decision

Approve one static, responsive landing page at
`https://my-friday.acoz.dev`, hosted on Cloudflare Pages and owned by the My
Friday repository.

The page should contain only:

- the My Friday name and a plain-language local-first value proposition;
- a short explanation of what the current pilot creates and what it does not
  touch;
- current platform support and the signing/notarization caveat;
- one primary **Download for Apple silicon** action;
- compact install and SHA-256 verification guidance, or a direct link to the
  canonical repository instructions when that is clearer;
- secondary links to source, the latest GitHub Release, licence, and fuller
  documentation.

GitHub Releases continues to do the heavy lifting for change history and
release detail. The landing page must not carry a copied changelog, version
number, release date, or hand-maintained latest-asset URL.

The primary action should use GitHub's stable latest-release download contract:

`https://github.com/acoz-labs/my-friday/releases/latest/download/my-friday-darwin-arm64.tar.gz`

The release path must publish that stable asset name from the already accepted
bytes on every public release. The current latest release may receive a
byte-identical stable-named alias after its digest is verified, allowing the
page to launch without waiting for another product release. The existing
commit-suffixed artifact and checksums remain authoritative evidence; no
artifact may be rebuilt merely to create the alias.

## Success and stop signals

Success means a first-time visitor can understand My Friday, its supported
environment, and its trust caveats; download the latest accepted archive in one
click; and reach canonical GitHub detail without a future website edit for an
ordinary release.

Change the decision if direct download cannot remain release-owned, if the page
needs dynamic release metadata, or if the deployment design cannot preserve
the repository's acceptance controls. Pause launch if the signing/notarization
or checksum state cannot be stated accurately. Stop rather than grow the page
into a changelog, documentation portal, or CMS.

## Candidate outcome map

### O1 — Ship the public landing page and stable download contract

- **Disposition:** selected
- **Outcome:** A public visitor can evaluate and download the latest accepted
  My Friday artifact from one trustworthy product-owned page.
- **Acceptance:** `https://my-friday.acoz.dev` serves one accessible,
  responsive static page; its primary download resolves without page-managed
  release metadata to a stable-named asset whose bytes match the accepted
  artifact in the latest public GitHub Release; the page accurately states
  current platform and signing/notarization support; source, latest release,
  licence, and canonical instructions are directly reachable; no changelog or
  duplicated release ledger is present; and deployment plus rollback are
  verified under the repository's release policy.
- **Dependencies:** Cloudflare Pages project and custom-domain configuration;
  a release publication path for the stable asset name; product-design review;
  and a Solution Design that reconciles the new public service surface with the
  existing artifact delivery profile.
- **Sequence:** next, before broader public promotion of the pilot.

## Privacy and evidence handling

All evidence in this pack is public or a sanitized claim about live platform
state. No account identifiers, tokens, private URLs, private paths, or provider
access instructions are present. Delivery must use repository-native secrets
and public documentation rather than private discovery context.

## Decision Spotlight

This is intentionally a small front door, not a second product site. Cloudflare
serves the page; GitHub Releases owns the moving release truth; the stable asset
name is the seam between them. If a normal release requires editing the page,
the implementation has missed the decision.

## Gate 1

The final candidate awaits product authority and an authorized approval on the
exact pull-request head.
