# Solution Design: Preview and create the assistant repositories

- **Status:** Final
- **Issue:** #3
- **Planning PR:** #10
- **Repository basis:** c669868cb297be51f02e1b6b7824e50b81da318b
- **Execution envelope:** implementation
- **Confidence:** Medium-high

The `implementation` envelope authorizes code, tests, documentation, and a
reviewed implementation pull request. It does not authorize merge, artifact
publication, product acceptance, or release.

## Decision

Build O1 as a single native Go command with embedded, versioned starter
templates and one declarative creation plan. The terminal wizard collects the
approved identity, communication, and location inputs; renders a deterministic
preview; and, only after an explicit `Create`, stages, validates, and promotes
two separate local Git repositories through a recoverable transaction.

The first supported contract is deliberately narrow: macOS 14 or later on
Apple silicon (`arm64`), a local APFS volume, an interactive terminal, and Git
2.28 or later. The executable is shell-independent; documentation uses the
default macOS `zsh`. Intel macOS, non-APFS volumes, cloud/network filesystems,
and other operating systems remain unclaimed until separately exercised.

## Needs Attention

- Public artifact publication under the **My Friday** name remains contingent
  on naming clearance. This does not block implementation or local validation.
- The first-customer acceptance run must verify the actual work machine meets
  the declared macOS, architecture, APFS, and Git contract before a support
  claim or artifact release.

## Decision Spotlight

| Concern | Selected default | Why |
|---|---|---|
| Mutation consent | Preview is read-only; the confirmation defaults to `Exit`, and only an explicit `Create` mutates disk. | Repository creation is consequential and must never follow an accidental Return key. |
| User text normalization and limits | Profile text is trimmed, normalized to NFC, checked for prohibited categories, then counted as grapheme clusters; name/address cap at 60 and purpose/custom guidance at 240. | Canonically equivalent text yields identical profiles, IDs, and previews; combining characters and emoji behave as one visible character. |
| Product boundary | The wizard creates two local repositories only. It performs no Codex installation, remote setup, network call, commit, secret access, import, or global configuration change. | This is the exact O1 outcome and keeps later trust boundaries out of the first slice. |
| Implementation form | One native Go executable with embedded templates and a pinned JSON Schema validator. | It gives macOS users a runtime-independent command while retaining typed filesystem/error handling and executable schemas; shell and Python alternatives create weaker recovery or runtime contracts. |
| Harness boundary | The assistant profile, repository roles, planner, and transaction are harness-neutral domain concepts; root `AGENTS.md` files are the only Codex-first projection in O1. No adapter framework or alternate harness ships. | Future Claude Code, pi, or another harness can receive an explicit capability mapping without forcing the owned data model to become a Codex detail or an unhelpful lowest common denominator. |
| Repository ownership | Runtime and governed memory remain separate and share only a non-secret deterministic assistant identifier. Absolute paths are not written into either repository. | The repositories must be independently movable, shareable, and protectable. |
| Profile authority | Identity and communication preferences live in runtime `assistant/profile.json`; generated instructions explicitly state they cannot override trust, safety, authorization, or tool policy. | Personality is presentation context, not a policy escalation surface. |
| Memory baseline | The memory repository contains governance instructions and empty category directories, but zero memory records. | O1 establishes ownership without pretending the governed-memory loop from O3 already exists. |
| Git posture | Both targets are initialized on branch `main` with an empty Git template; no commits, remotes, credentials, or user Git settings are created. | The user owns history and remote decisions, and private global Git templates must not be imported. |
| Filesystem privacy | Created parents, repositories, and support state are owner-only. An adopted empty shell is explicitly normalized to `0700` on success; rollback restores its original mode. | Memory can be sensitive, and a previewed consistent completion mode avoids inheriting accidental exposure. |
| Collision policy | Existing symlinks and non-empty targets fail before mutation. Distinct, non-nested canonical targets are required; `/` and the user home itself are forbidden. | The tool must never merge with, replace, or ambiguously traverse unrelated content. |
| Location defaults and parents | Combined placement uses stable editable `my-friday-runtime` and `my-friday-memory` names, never the assistant name. Missing parent segments are listed, created `0700`, and rolled back only when transaction-owned and still empty. | Identity must not unexpectedly become a path, while a user-selected new folder remains a supported recoverable flow. |
| Path grammar | Capture the invocation directory once; use it as combined-parent default and as the base for relative paths. Accept absolute paths and current-home `~`; reject named-user tilde and environment-variable expressions. Preview shows normalized absolute targets and resolved symlink ancestors. | Shell-independent parsing must be predictable and must expose the actual mutation targets before consent. |
| Exact-plan reruns | A matching completed pair reports `Already complete` with no write; a matching journal reports the interrupted phase and bounded recovery path; only unrelated content is a collision. | Idempotent retry is part of safe creation and must not mislabel owned success/failure state as foreign. |
| Failure policy | Stage and validate both repos before promotion; retain a minimal owner-only transaction journal only when automatic rollback cannot restore the pre-run state. | A crash cannot be made atomic across two directories, so explicit, idempotent recovery is safer than claiming atomicity. |
| Observability and privacy | The terminal reports phase, plan identifier, paths, and recovery action without telemetry or persistent success logs; failure state excludes profile text. | Users need diagnosis without creating a second store of personal content. |
| Terminal accessibility | Sequential text, numbered choices, explicit status lines, and ordinary line input; no cursor rewrites, spinner, required color, or motion. | The flow remains usable with screen readers, copied transcripts, and reduced visual attention. |

## Plan Map

- [Context](01-context.md)
- [Decision](02-decision.md)
- [Design](03-design.md)
- [Verification](04-verification.md)
- [Implementation handoff](05-handoff.md)

## Final Gate

The complete pack is ready for independent maintainer review. After every
finding is resolved, the product authority must approve this exact planning PR
head and the `implementation` envelope before merge or implementation dispatch.
