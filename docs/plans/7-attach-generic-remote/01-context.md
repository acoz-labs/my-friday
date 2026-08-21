# Context

## Problem And Desired Outcome

Issue [#7](https://github.com/acoz-labs/my-friday/issues/7) selects discovery
outcome O5: after the local bootstrap succeeds, a user may deliberately attach
either generated repository to an explicitly supplied generic Git remote.

The useful outcome is narrower than hosted onboarding. A technically capable
user already has a destination and wants one inspectable local configuration
change. The product must make the runtime-versus-memory sharing decision clear,
avoid implying that attachment transmits data, and keep destination creation,
authentication, visibility, retention, permissions, and provider policy with
the user.

## Current State

The repository basis is
`5bc309226d2c40e1473a4011c1bd8552c995919d`.

- `cmd/my-friday/main.go` exposes `init`, `validate`, `recover`, and `version`.
  It has no `remote` command.
- `internal/repository/repository.go` validates a pair by factoring a private
  per-repository contract check. Ordinary validation requires a local `.git`
  directory but intentionally permits later commits, branches, and remotes.
  `ValidateFreshPair` is creation-only and rejects remotes, so attachment needs
  a public single-repository validator based on the ordinary contract rather
  than the fresh-pair predicate.
- `internal/gitexec/gitexec.go` already centralizes Git execution through
  literal `exec.Command` argv, a minimal environment, and a scrubbed observer.
  The new capability must strengthen this boundary for local-config-only reads
  by disabling system/global configuration and rejecting repository include
  directives; it must not introduce shell interpolation.
- `internal/terminal/wizard.go` and its tests establish a plain UTF-8,
  line-oriented preview, safe default exit, exact confirmation, stable receipt,
  and no-ANSI precedent. The remote flow is separate from `init` because local
  bootstrap must remain complete without a hosted service.
- `internal/terminal/evidence_test.go` and
  `internal/terminal/production_boundary_test.go` provide transcript and
  subprocess-boundary precedents. Exact-head remote evidence needs additional
  config/tree manifests and a network/credential/global-Git negative observer.
- `docs/architecture/repository-bootstrap.md` defines two separate generated
  repositories and says `init` has no network, credential, commit, or remote
  effect. That remains true; remote attachment is a later capability.
- `docs/product.md` makes local-only operation the default, preserves separate
  runtime and memory sharing choices, and deliberately defers generic remote
  attachment after the core local outcomes.
- `.github/workflows/nominate-artifact.yml`, `product-acceptance.yml`, and
  `release-artifact.yml` pass a caller-supplied artifact identifier through the
  lifecycle but do not build, upload, download, or republish exact archive
  bytes. Issue #4's final planning PR #15 at
  `ef28270dba982f4e31bf1b2171b3dc28e093b7e4` defines the repository-wide
  exact-byte chain. This plan adopts that contract but remains independently
  executable if its implementation has not landed.

Official Git documentation consulted on 2026-08-21 establishes the mechanism:

- [`git remote add`](https://git-scm.com/docs/git-remote) adds a named remote;
  only its optional `-f` performs an immediate fetch. Omitting `-f` makes the
  selected command a local configuration operation.
- [`git config`](https://git-scm.com/docs/git-config) can constrain reads and
  writes to repository-local configuration, reports invalid/unwritable config
  with nonzero status, and changes one config file at a time.
- Git's [URL syntax](https://git-scm.com/docs/git-clone#_git_urls) includes
  HTTPS, SSH, SCP-style, local/file, and helper forms. It also documents that
  `<transport>::<address>` explicitly invokes a remote helper.
- [Remote helpers](https://git-scm.com/docs/gitremote-helpers) run independent
  processes and may transfer refs and objects. That extensibility is precisely
  why unknown schemes and helper syntax must not enter this bounded attach flow.

## Actors And Critical Journeys

### Technical user

- **Runtime attachment:** supplies the runtime path and HTTPS fixture, sees the
  canonical repository and Runtime disclosure, types `Attach`, and receives an
  exact local-config receipt.
- **Memory attachment:** independently supplies the memory path and SSH/SCP
  address, sees the stronger Memory disclosure, and decides whether that
  repository should be shareable later.
- **Safe exit:** presses Return, sends EOF, types `q`, or mistypes the case and
  receives `No changes made` with byte-identical repository state.
- **Repeat:** reruns the exact command and receives `Already attached` without
  a lock or write.
- **Collision/recovery:** an existing or incomplete `origin`, invalid config,
  lock, permission failure, or concurrent change is preserved and explained;
  the user inspects with Git and reruns only after resolving it.

### Independent acceptor

Downloads the nominated immutable Darwin/ARM64 artifact, verifies its digest,
exercises both roles and denial/retry cases under a disposable non-admin macOS
identity, proves prohibited adjacent effects with a subprocess observer and
before/after manifests, and records fresh candidate-bound transcripts.

### Maintainer and release operator

Verify the grammar and injection boundary, Git config semantics, privacy-safe
evidence, exact-byte candidate chain, documentation promotion, independent
acceptance, GitHub Release digest, and issue closure.

## Acceptance And Non-Goals

The approved criterion is designable as six groups:

1. accept one explicit recognized repository and one bounded credential-free
   provider-neutral address;
2. explain repository role, future data/history transmission, destination
   ownership, and all prohibited effects before mutation;
3. require exact `Attach`, then revalidate and add only local `origin`;
4. verify exact read-back, idempotent repeat, collision refusal, and recoverable
   lock/write/interruption behavior;
5. prevent raw unsafe-address disclosure and prove no network, credential,
   global Git, content, commit, fetch, push, or other-repository effect; and
6. accept and release the exact nominated artifact with transcript evidence.

Non-goals are provider selection/login/API use; remote repository creation;
permission or visibility administration; credential capture or testing;
connectivity checks; `fetch`, `push`, commit, branch, tag, or content changes;
custom remote names; overwrite, rename, detach, or removal UX; local/file or
bundle remotes; plaintext HTTP/Git/FTP transports; arbitrary remote helpers;
automatic synchronization; multiple repositories per invocation; telemetry;
background work; Linux, Windows, or Intel support; and changing `init`.

## Constraints, Dependencies, And Risks

- The pilot remains macOS 14+, Apple silicon, local APFS, Git 2.28+, an
  interactive UTF-8 terminal, and a non-root user.
- The repository may have evolved with commits, branches, content, and other
  remotes. Contract-v1 manifests and schema/profile semantics remain the
  recognition boundary; attachment must not reapply creation-only freshness.
- Accepted repository paths may traverse a symlink. The flow shows entered and
  canonical paths and operates on a revalidated canonical root. Path/inode
  pinning narrows TOCTOU but does not claim an adversarial sandbox against an
  administrator replacing ancestors between an external Git process's checks.
- `.git/config` is mixed Git/user state. My Friday may add only the canonical
  `origin` entries when no `origin` subsection exists. It never edits config
  directly, removes a lock, repairs syntax, follows includes, or overwrites a
  partial/different state.
- An accepted address is intentionally visible in terminal output and may be
  retained by shell history or process listings because it is an argument.
  Help and preview warn users never to supply credentials. Rejected and
  pre-validation values are never echoed.
- The direct local value is not necessarily the endpoint a later ordinary Git
  command resolves. User-owned local, global, or system `url.*.insteadOf` and
  `pushInsteadOf` rules may rewrite it. Attachment verifies only the stored
  local address; the preview and receipt must not imply effective-endpoint
  verification.
- Network non-use must be executable evidence, not an inference from fixture
  hosts. The production Git observer and acceptance process monitor must prove
  only local inspection/config argv were attempted.
- The repository has no staging service; production for this artifact is a
  digest-verified GitHub Release asset. The same bytes must move from nomination
  through acceptance to release.

## Evidence, Assumptions, And Unknowns

### Evidence

- Approved discovery PR `acoz-labs/.github#2` at
  `b6db62bf15c8d6ad7a15f7533e6aa5981ae1cd8a`, outcome O5, supplies Phase 1
  product authority.
- Issue #7 is open, contains the immutable discovery marker, and has no prior
  solution-design or implementation link.
- The repository source and official Git documentation cited above establish
  the current contract and supported mechanism.
- Product-experience review classifies this as a net-new terminal flow and
  requires role-first disclosure, exact `Attach`, safe defaults, privacy-safe
  transcripts, and fresh exact-candidate evidence. No further product choice is
  needed before Solution Design.

### Assumptions

- Users already possess a credential-free remote address and have separately
  arranged destination ownership, permissions, visibility, and authentication.
- Fixed `origin` and one repository per invocation are sufficient for the
  initial outcome.
- A deliberately strict HTTPS/SSH subset serves common hosted Git destinations
  without exposing Git's executable helper or local-path surface.
- Standard Git inspection/removal commands are sufficient recovery/undo
  guidance; a product detach command is not required for this issue.

### Validation needs, not blocking unknowns

- Product owner and at least two design partners should independently explain
  what changed, what did not, whether data has left the machine, and who owns
  destination permissions after using the candidate.
- Evidence should test whether the runtime/memory disclosure is read and
  whether users can recover from an existing-`origin` collision without
  facilitation.
- Sustained use may show demand for custom names, local remotes, or provider
  assistance. Those are later product decisions, not implementation guesses.
