# Product direction

## Promise

My Friday helps a technically capable Codex user establish an assistant they
can understand, version, extend, repair, and remove.

It is a bootstrap and lifecycle toolkit. It is not a model, hosted assistant
service, secrets manager, or promise of autonomous competence.

## Initial audience

The initial audience is developers, self-hosters, and technical knowledge
workers who:

- use or intend to use Codex;
- are comfortable with git, terminals, and local developer tools;
- want control over assistant instructions, skills, and memory; and
- accept some operational responsibility in exchange for ownership.

The first validation cohort consists of the product owner using a separate
machine and a small group of interested technical design partners. Interest is
an early signal, not proof of independent usability or sustained value.

## Critical tasks

Users must be able to:

1. Understand what My Friday owns, changes, stores, and does not manage.
2. Choose an assistant identity, personality inputs, repository locations, and
   local or remote posture.
3. Preview a deterministic change plan before any mutation.
4. Create separate runtime and governed-memory repositories.
5. Install a narrowly owned Codex baseline without disturbing unrelated
   configuration.
6. Verify, repair, upgrade, uninstall, and roll back the managed installation.
7. Use a guided deterministic workshop to define, preview, create, inspect,
   validate, test, install, verify, enhance, disable, and remove a
   version-controlled capability without memorizing its file schema.
8. Use that capability workflow to record observations and chronology,
   deliberately promote durable claims, and recall relevant context in later
   tasks.
9. Work locally and optionally attach an existing source-control remote.

## Product boundaries

- Local-only operation is the default and requires no hosted source-control
  account.
- Runtime configuration and memory remain separate so users can make different
  sharing and privacy choices for each repository.
- Every installed projection must have explicit ownership, preview, collision
  handling, verification, and reversal.
- Personality is user-defined; trust boundaries and safety policy are not
  personality settings.
- Capability source is version-controlled and profile-bound. The first shipped
  builder is a deterministic terminal workshop: it gathers explicit behavior,
  shows the complete canonical source diff, and requires exact `Create source`
  or `Update source` confirmation without installing. Deterministic My Friday
  tooling separately owns validation, testing, activation, verification,
  upgrade, and reversal, with human review and an explicit confirmation at each
  mutation boundary.
- Conversational-agent input to the workshop is deferred until a supported
  Codex surface reliably produces a terminal structured proposal or a
  qualifying action and filesystem effect. Model prose, catalog discovery, or
  exit zero alone never grants source or lifecycle authority.
- The first capability profile is instruction-only. It permits no executable
  code, arbitrary dependency, network access, credential use, background
  process, or durable user data.
- Governed memory is the first data-bearing core capability and must use the
  shared outer capability lifecycle while retaining stricter domain-specific
  schemas, sensitivity, transaction, recovery, and review requirements.
- Durable memory requires deliberate promotion with provenance and visible
  uncertainty. Ordinary activity must not silently become permanent belief.
- My Friday does not promise autonomous self-modification, generated-code
  activation, third-party capability installation, or a capability marketplace.
- My Friday does not claim complete privacy. Models, remotes, plugins, and tools
  may transmit data and must disclose their own boundaries.
- Employer policy, credentials, permissions, and third-party integrations
  remain outside the core bootstrap promise.

## Public product surface

My Friday's canonical public page is
`https://acoz.dev/projects/my-friday/`. It is part of the existing `acoz.dev`
service and its open-source project collection at
`https://acoz.dev/open-source/`; source availability is project metadata rather
than part of the canonical project URL. `https://my-friday.acoz.dev` is a
permanent redirect to the canonical route, not a second content surface.

The page is a front door for technically capable macOS users to understand the
local-first product, confirm current platform and distribution caveats, and
download the latest accepted Apple silicon artifact. The `acoz` repository
owns page presentation and deployment; this repository owns the binary,
checksums, release notes, licence, documentation, and download contract.

GitHub Releases remains authoritative for binaries, checksums, release notes,
licensing, and rollback history. The primary download uses GitHub's stable
latest-release asset contract:

`https://github.com/acoz-labs/my-friday/releases/latest/download/my-friday-darwin-arm64.tar.gz`

Every public release must publish that stable asset name from the accepted
bytes. The landing page does not duplicate the changelog, version, release
date, or other moving release metadata, and an ordinary release must not
require a website edit. The page is not a documentation portal, hosted product,
account surface, analytics application, or content-management system.

## Approved outcome sequence

### Selected

1. Preview and create valid separate runtime and memory repositories.
2. Install, verify, repair, uninstall, and roll back the managed Codex baseline.
3. Build and operate one inspectable instruction-only capability through the
   source-first capability workshop.
4. Build governed memory through the same outer lifecycle as the first
   data-bearing core capability and complete its loop across fresh tasks.

### Deliberately deferred

5. Let a conversational agent propose or supply workshop answers while My
   Friday and the user retain deterministic preview, source-write, and
   lifecycle authority.
6. Attach either generated repository to an explicitly supplied generic git
   remote without provider-specific authentication.

### Parked

- Local scripted capability profiles.
- Provider-specific remote onboarding.
- Multi-machine memory synchronization.
- Broad operating-system support beyond the first verified environment.

### Not part of the current product direction

Autonomous self-modification, generated or third-party executable capability
activation, remote capability distribution, a marketplace, voice, web
interfaces, messaging transports, connectors, scheduling, hosted services, and
broad autonomous orchestration require a new product decision.

## Validation signals

Continue when the product owner and at least two design partners independently
complete the core lifecycle, build and enhance one instruction-only capability,
understand the ownership and data boundaries, and choose to keep using the
result. Governed memory must reuse the outer capability lifecycle without
concealing a second installer or weakening its data protections.

Change direction when users repeatedly need facilitation, the two-repository or
capability-profile model creates more friction than trust, the builder merely
renames bespoke development, memory cannot honestly reuse the shared lifecycle,
or first-use value requires provider-specific setup.

Pause when capability files, projections, dependencies, permissions, data
effects, or reversal cannot be bounded before activation; when safe collision
handling, rollback, interrupted-run recovery, naming clearance, workplace
constraints, or a bounded Codex compatibility contract cannot be demonstrated;
or when private data cannot remain separate from shareable capability source.

Stop when target users do not retain the product, documentation and templates
or Codex-native tooling solve the complete capability lifecycle adequately, the
abstraction delays useful capabilities without reuse, or global installation
creates unacceptable security or maintenance exposure.
