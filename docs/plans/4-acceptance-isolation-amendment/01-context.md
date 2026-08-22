# Context

## Problem And Desired Outcome

Issue #4 shipped the manifest-owned Codex baseline in implementation PR #19,
but product acceptance has not run. Its approved PR #15 plan requires a unique
disposable non-admin macOS account, home, login keychain, and marker-bounded
account teardown. That adds a physical administrator gate and destructive
identity-management authority to acceptance even though the product itself is
nonprivileged and manages neither accounts nor keychains.

Anthony approved replacing that unexecuted boundary in principle with an
administrator-free, same-host containment boundary. The desired outcome is an
executable acceptance path that protects the acceptor's live Codex installation
and deployed runtime projection, exercises the exact nominated artifact through
the full lifecycle and interruption matrix, produces durable sanitized proof,
and cleans up without ambiguous deletion.

This is a material amendment after implementation, not a retroactive rewrite.
PR #15 and PR #19 remain immutable provenance. The amendment supersedes only
the acceptance-host isolation mechanism and the implementation work needed to
operate and bind it.

## Current State

Repository basis `0ad07cf578ed4e7f7f7d8ae7d30aa5013bb9d83b` establishes:

- `cmd/my-friday/main.go` resolves `HOME`, requires `CODEX_HOME` to be a strict
  descendant, and exposes install, verify, repair, upgrade, rollback, uninstall,
  and recover commands.
- `internal/codexhome/lifecycle.go` owns manifest, generation, transaction
  journal, no-follow mutation, and recovery semantics. A production operation
  persists ordinary journal phases; test-only hooks are not present in the
  release artifact.
- `docs/architecture/installed-codex-baseline.md`,
  `docs/decisions/0002-manifest-owned-codex-baseline.md`, and
  `docs/runbook.md` describe the shipped ownership and recovery contract.
- `.github/workflows/nominate-artifact.yml` builds the Darwin/ARM64 archive,
  uploads a named artifact, and emits an `artifact-v1` authority containing run,
  artifact ID, name, and executable SHA-256.
- `bin/prepare-release-artifact`, `bin/record-product-acceptance`,
  `bin/release-gate`, `.github/workflows/product-acceptance.yml`, and
  `.github/workflows/release-artifact.yml` carry exact-byte authority through
  nomination, acceptance status, and GitHub Release.
- Issue #4 records original plan PR #15, merged implementation PR #19, and an
  existing nominated artifact for `0ad07cf...`. Because the acceptance contract
  is changing, that candidate cannot satisfy the amended outcome; implementation
  and a fresh nomination must precede acceptance.

The repository has no shipped local macOS acceptance supervisor and no durable
schema binding the acceptance decision to a separately published evidence
record. The current product-acceptance workflow records operator-supplied
evidence text.

## Actors And Critical Journeys

- **Independent product acceptor:** obtains the nominated authority, runs one
  documented no-admin command on supported Apple Silicon macOS, supplies a
  standard-provider API secret through an approved process boundary, reviews
  sanitized evidence, and records approval or rejection.
- **Acceptance supervisor:** creates and proves the APFS/sandbox boundary,
  downloads and verifies exact bytes, runs controls and scenarios, supervises
  interruption, compares protected state, publishes evidence, and cleans up.
- **My Friday candidate:** runs unmodified under synthetic paths and the sandbox;
  it receives no administrator or account-management authority.
- **Real Codex CLI:** performs one authenticated smoke inside the disposable
  `CODEX_HOME`; its temporary `auth.json` is sensitive disposable state.
- **Release workflow:** accepts only the candidate/artifact and evidence authority
  approved for issue #4, then uploads the same verified bytes.

Critical journeys are boundary setup and refusal, lifecycle success and denial,
drift/rollback/recovery, real-Codex discovery, evidence publication, exact-byte
release, cleanup success, and cleanup mismatch preservation.

## Acceptance And Non-Goals

Acceptance must prove:

1. exact nominated bytes run on the supported macOS/Apple Silicon/APFS/Git
   stack without touching the live installation or deployed runtime projection;
2. lifecycle commands, denials, repair, rollback, uninstall, and externally
   interrupted recovery behave as shipped;
3. lifecycle processes and descendants can write only inside the disposable
   volume and cannot use network; supervisor-owned evidence staging remains
   outside candidate authority;
4. the isolated real-Codex smoke loads the installed baseline, with network and
   auth enabled only for that scenario;
5. protected pre/post state is equal, durable evidence is sanitized and bound
   to acceptance, and the disk image and local evidence staging are removed.

This plan does not redesign installed-baseline behavior, add a daemon, manage
accounts/keychains, claim read confidentiality, test root execution, introduce a
VM, or reopen unrelated PR #15 decisions. UI acceptance is not applicable.

## Constraints, Dependencies, And Risks

- The mount and evidence roots must be unique, absent, marker-bounded directories
  beneath the accepting user's canonical real home. Broad paths, symlinks,
  unresolved variables, and pre-existing entries are refused.
- APFS image creation and attachment must require no administrator authority;
  attachment uses an explicit mountpoint, browsing disabled, and ownership on.
- `sandbox-exec` is deprecated. Availability alone is insufficient: the exact
  profile and controls must pass on the accepting build. There is no permissive
  fallback.
- Same UID means the candidate may retain read access granted by the profile and
  may see the user's login keychain. No secret-bearing path or content enters
  published evidence. This is not a confidentiality boundary.
- The real Codex smoke needs a test-only `OPENAI_API_KEY` source approved by the
  operator. The supervisor pipes it to `codex login --with-api-key`; Codex may
  persist `$CODEX_HOME/auth.json` only inside the disposable volume. Logout and
  image destruction remove it.
- A host crash may leave an attached image or private evidence. The runbook must
  identify and clean only marker-proven leftovers; automatic force detach or
  broad recursive deletion is prohibited.
- The accepted macOS build, Codex CLI contract, APFS behavior, and sandbox profile
  digest form part of evidence. Changed external behavior requires a failed
  preflight or new review, not optimistic acceptance.

## Evidence, Assumptions, And Unknowns

Verified on an Apple Silicon macOS 26.4.1 host as the ordinary accepting user:

- a sparse APFS image was created, attached at an explicit path below the real
  home with ownership enabled, and detached without `sudo`;
- the mounted root was local APFS and owned by the current user;
- a minimal sandbox allowed a write within the volume and denied a write to a
  sibling path; and
- detach and explicit marker-bounded cleanup left no mounted or file residue.

Local manuals confirm program-readable `hdiutil` output, explicit mountpoints,
APFS creation, non-browsing attachment, ownership control, and detach by device
or mountpoint. They also mark `sandbox-exec` deprecated. The probe establishes
feasibility on that build, not a permanent platform guarantee.

Assumptions requiring implementation tests are that the reviewed production
profiles can allow only the device/process reads required by My Friday and the
Codex smoke while preserving write and network restrictions, and that aggregate
metadata manifests can be generated without reading secret contents. Failure of
either assumption reopens design; it does not authorize a weaker boundary.

No product-authority unknown remains before planning review. Exact-head approval
must acknowledge the explicitly weaker distinct-principal proof in exchange for
removing administrator/account-management authority.
