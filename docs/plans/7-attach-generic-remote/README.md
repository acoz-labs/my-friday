# Solution Design: Attach an existing generic Git remote

- **Status:** Final
- **Issue:** #7
- **Planning PR:** #16
- **Repository basis:** 5bc309226d2c40e1473a4011c1bd8552c995919d
- **Execution envelope:** through-production

## Decision

Add one foreground terminal flow:
`my-friday remote attach --repository PATH --url REMOTE_ADDRESS`. It validates
one recognized contract-v1 runtime or memory repository, classifies a bounded
credential-free HTTPS or SSH address, previews the repository role and complete
sharing/ownership consequences, and requires exact `Attach`. It then invokes
Git once with literal argv to add repository-local `origin`, re-reads direct
local configuration, and reports success only when the canonical URL and fetch
refspec are exact.

The command performs configuration only. It never contacts the destination,
creates a hosted repository, authenticates, reads or stores credentials,
changes permissions, commits, fetches, pushes, edits global Git configuration,
or discovers or changes the user's other My Friday repository.

## Needs Attention

- Exact-candidate acceptance must run on supported Apple silicon, macOS 14 or
  later, local APFS, and Git 2.28 or later. It must retain sanitized transcripts
  and before/after manifests while using only fixture addresses and paths.
- Acceptance uses a disposable non-admin macOS identity with marker-bounded
  creation and teardown. Creating/removing that identity may require one
  physical administrator authentication. The same operator boundary loads the
  temporary UID-scoped PF rule and DTrace monitor used for child-inclusive
  no-network acceptance; no authority or credential reaches My Friday.
- The current artifact workflows identify a candidate with an input string but
  do not yet carry a built archive through nomination, acceptance, and release.
  Implementation must either reuse the exact-byte chain delivered from issue
  #4's final plan or deliver the equivalent enabling work in this issue. Issue
  #7 does not depend on issue #4 implementation.

## Decision Spotlight

- **One repository per invocation.** Runtime and memory remain independent
  sharing decisions; the command never searches for or mutates a sibling.
- **Fixed `origin`.** The first outcome uses Git's conventional remote name and
  refuses collisions. Custom names and replacement are future features, not an
  extra prompt.
- **Provider-neutral, protocol-bounded address.** Accepted input is HTTPS,
  `ssh://`, or SSH/SCP-style syntax with a strict ASCII host/path grammar.
  Local/file remotes, plaintext protocols, query/fragment data, URL passwords,
  unknown schemes, `ext::`, and remote helpers are refused before Git sees the
  value. Generic means provider-neutral, not every transport Git can execute.
- **Full disclosure on every mutating attempt.** Role appears before the
  consequence text; exact case-sensitive `Attach` is the only mutation signal.
  Return, EOF, `q`, whitespace variants, and every other response are safe exits.
- **Local configuration only; no connectivity claim.** The operation adds the
  two canonical `remote.origin` entries through Git without `-f`, then verifies
  direct local read-back. It never claims the destination exists, is private,
  belongs to the user, or is reachable.
- **Stored address is not a resolved-endpoint guarantee.** My Friday verifies
  the literal repository-local value. Later Git commands may apply user-managed
  `url.*.insteadOf` or `pushInsteadOf` rules from configuration this command
  deliberately does not inspect. The preview makes that boundary explicit;
  My Friday never claims to verify the future transport endpoint.
- **Ambiguity fails closed at Git's key-semantic boundary.** A different,
  duplicated-key, incomplete-key, included, or otherwise non-canonical `origin`
  is never adopted or overwritten. Empty, comment-only, or duplicate-empty
  `[remote "origin"]` text has no Git key/value semantics and is allowed as
  absence; native fixtures prove Git's resulting canonical state and comment
  preservation. Lock and uncertain-write failures preserve state.
- **Accepted values are visible; unsafe values are not.** A validated address is
  shown in the interactive preview so the user can verify it. Rejected input,
  unsafe existing configuration, stable errors, CI annotations, and durable
  evidence never echo or transform a possibly credential-bearing value.
- **No attachment registry.** Git's repository-local config is the only durable
  state. My Friday adds no manifest, journal, telemetry, credential store, or
  record in the runtime/memory content.
- **No-network acceptance uses kernel and syscall evidence.** The argv observer
  proves command selection only. Exact-candidate acceptance separately runs the
  quiescent disposable UID under a temporary outbound-blocking PF anchor while
  a privileged DTrace monitor records resolver and IPv4/IPv6 socket attempts by
  that UID and descendants. A positive control must be blocked and observed;
  candidate counters/events must remain zero; exact cleanup is receipt-bound.

## Plan Map

- [Context](01-context.md)
- [Decision](02-decision.md)
- [Design](03-design.md)
- [Verification](04-verification.md)
- [Implementation handoff](05-handoff.md)

## Final Gate

The plan becomes final only after the draft planning PR is linked to issue #7,
the complete pack has no blocking maintainer finding, validation passes, and
the PR number and `through-production` envelope are recorded. Product authority
must then approve the exact final head before merge.
