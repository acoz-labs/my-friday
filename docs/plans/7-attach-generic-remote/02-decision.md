# Solution Decision

## Decision Drivers

1. Keep runtime and memory sharing choices independent and explicit.
2. Make attachment a transparent local Git configuration operation, never a
   connectivity, authentication, privacy, or publication claim.
3. Reject credential-bearing, executable-helper, plaintext, and local-path
   address forms before Git receives them.
4. Preserve arbitrary repository content/history and all non-`origin` Git
   configuration while failing closed on ambiguous `origin` state.
5. Use Git's own config parser, lock, and canonical remote setup without shell
   interpolation or bespoke `.git/config` editing.
6. Make cancellation, collision, retry, interruption, and exact repetition
   deterministic and evidence-friendly.
7. Retain no duplicate registry, credential, remote value, or telemetry.
8. Ship and accept the exact immutable macOS artifact without waiting on a
   separate issue's implementation.

## Competing Approaches

### A. Documentation-only `git remote add`

Document a command for users to run directly.

This is mechanically minimal and Git-native, but it cannot identify the
repository role, disclose runtime/memory consequences at the decision point,
screen credential-bearing/helper forms, make Return the safe default, or retain
candidate-bound evidence for the product contract. It leaves the entire value
of outcome O5 to documentation and is therefore incomplete.

### B. Provider-specific onboarding

Ask for a host, log in through a provider CLI/API, create the destination, set
visibility/permissions, attach it, and perhaps push.

This could reduce setup friction but adds accounts, credentials, network calls,
provider policy, destructive creation choices, and data transmission. Discovery
explicitly parks provider-specific onboarding. It is a different trust boundary
and product outcome.

### C. Parse and atomically edit `.git/config` in Go

Open the config file, add `[remote "origin"]`, write a lock, fsync, and rename.

This would offer descriptor-relative control, but it duplicates Git's config
grammar, include semantics, quoting, locking, formatting preservation, and
future compatibility. A correct writer is a new subsystem for one standard Git
operation and would still need Git read-back.

### D. Bounded input plus one literal `git remote add`

Validate a single My Friday repository and strict network-address grammar,
inspect direct local config with includes disabled, preview and confirm, then
execute `git -C <canonical> remote add -- origin <address>` as literal argv.
Omit `-f`; read direct local `remote.origin.url` and `.fetch` back exactly.

This delegates config syntax and locking to Git while keeping all dangerous
interpretation outside the boundary. It adds no shell, network, provider,
credential, persistence, or general config framework.

## Adversarial Comparison

| Approach | Primary failure under the approved outcome | Assessment |
|---|---|---|
| Documentation only | No product-owned role disclosure, confirmation, screening, or evidence | Fatal incompleteness |
| Provider onboarding | Creates new network, credential, permission, and data-transfer authority | Fatal scope/trust expansion |
| Direct config writer | Reimplements a security-sensitive Git grammar and lock protocol | Avoidable correctness risk |
| Bounded Git command | External process and narrow TOCTOU remain; strict pre/post checks required | Manageable with tests and explicit limits |

Git accepts local paths, file URLs, unknown transport helpers, and explicit
`<transport>::<address>` helper dispatch. Passing arbitrary user input to
`remote add` would therefore store values whose later use can run helper
programs or access an unintended local path. The selected grammar is a product
allowlist, not a claim that Git itself considers other syntax invalid.

Git also supports configuration includes and multi-valued keys. Following an
include would read an external file and make the state/write boundary ambiguous.
The selected flow rejects direct local include directives for this operation,
disables system/global config for every inspection and write subprocess, and
requires exactly absent or canonical direct-local `origin` state.

## Selected Approach

Choose D with high design confidence.

The smallest coherent implementation introduces a pure `remoteaddress` parser,
a public single-repository contract inspection, a `remoteconfig` adapter over
the existing observable Git runner, and one terminal flow. The adapter owns a
typed snapshot of direct local config, not a new durable state model.

The accepted grammar is deliberately smaller than Git's:

| Form | Accepted shape | Refused within the form |
|---|---|---|
| HTTPS | `https://host[:port]/path` | userinfo, query, fragment, percent escapes, empty host/path, nonnumeric/out-of-range port |
| SSH URL | `ssh://[user@]host[:port]/path` | password userinfo, query, fragment, percent escapes, empty host/path, unsafe username/port |
| SCP-style SSH | `[user@]host:path` | slash before separator, empty host/path, ambiguous/local form, unsafe username/path |

Accepted input is ASCII. DNS labels must start and end with an alphanumeric
character and may contain only interior ASCII alphanumerics or hyphens; IPv4
and bracketed IPv6 literals must pass numeric parsing. SSH usernames must start
with an alphanumeric character and then contain only ASCII alphanumerics, dot,
underscore, or hyphen. No host or username can begin with an option marker.

Paths use slash-separated ASCII segments containing letters, digits, dot,
underscore, hyphen, and tilde. Every segment is nonempty, is neither `.` nor
`..`, and starts with an alphanumeric, dot, or tilde rather than hyphen. URL
paths begin with `/`; SCP-style paths begin with a safe segment, `/`, or a
bounded `~user/` prefix. They contain no whitespace, controls, bidi/format
characters, percent encoding, shell metacharacters, query, or fragment. The
original accepted string is stored and read back byte-for-byte; it is never
normalized or rewritten.

Everything else fails before Git sees it: `http`, `git`, `ftp`, `ftps`, `file`,
absolute/relative paths, Windows/UNC paths, bundles, unknown schemes,
`<transport>::`, `ext::`, credential/userinfo forms, control/line separators,
NUL, leading-option shapes, and empty values.

This restriction does not mathematically prove that a user chose a non-secret
repository path. It removes all supported credential channels and opaque
encoding, warns that the argument/output are visible, and never stores values
outside `.git/config`. Accepting arbitrary token-shaped paths and attempting to
classify secrets heuristically would create false confidence; users remain
responsible for supplying a destination address rather than a credential.

No implementation dependency on issue #4 is introduced. At implementation
time, the branch inspects current `main`: if the exact-byte artifact chain from
PR #15 exists, it reuses and tests it; otherwise the issue #7 implementation PR
includes the same deterministic build/upload/download/digest/release enabling
slice before nomination. No code is copied speculatively and no string-only
candidate may be accepted.

## Decisions Ledger

| Decision | Rationale | Evidence |
|---|---|---|
| Separate `remote attach` command | Local bootstrap remains independently useful | Issue #7 sequencing; `docs/product.md` |
| One repository, explicit path | Runtime and memory are independent privacy choices | Product boundaries and experience review |
| Fixed `origin` | Small conventional first contract; collisions remain visible | Git remote convention; approved safe default |
| HTTPS and SSH subset only | Covers common hosted destinations without plaintext, local, or helper execution surfaces | Official Git URL/helper docs |
| Literal `git remote add --` without `-f` | Git owns config locking; end-of-options blocks option injection; no fetch occurs | Official `git remote` docs and local argv verification |
| Direct local config, includes disabled | No global/user/include access or ambiguous inherited state | Official `git config` scopes |
| Stored value distinct from future resolved endpoint | User Git rewrites remain outside My Friday authority and may alter later transport | Official `insteadOf`/`pushInsteadOf` behavior |
| Git-key-semantic absence/canonical/collision | Empty/comment-only section text has no remote semantics; partial/duplicated keys are never repaired implicitly | Native Git behavior and idempotency criteria |
| No durable My Friday registry | Git config already owns the fact; duplication adds drift/privacy risk | YAGNI and data minimization |
| Full accepted address only in foreground output | User must verify destination; unsafe/durable surfaces must not leak it | Product-experience privacy rules |
| Through-production with enabling work | User asked for released completion; artifact profile has no staging | Repository release policy and PR #15 contract |
| PF deny plus UID-scoped DTrace acceptance | Argv cannot prove network non-use; parsed kernel rules/counters and global UID-filtered socket syscalls cover the candidate and Git children | macOS `pfctl` enable references/direct-child anchors and DTrace syscall probes, verified by positive control |
