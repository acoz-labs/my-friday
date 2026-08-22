# Solution Design Amendment: No-admin installed-baseline acceptance isolation

- **Status:** Draft
- **Issue:** #4
- **Planning PR:** #27
- **Repository basis:** 0ad07cf578ed4e7f7f7d8ae7d30aa5013bb9d83b
- **Execution envelope:** through-production
- **Amends:** original issue #4 Solution Design in PR #15, accepted at
  `ef28270dba982f4e31bf1b2171b3dc28e093b7e4`
- **Implemented basis:** PR #19, merged as
  `0ad07cf578ed4e7f7f7d8ae7d30aa5013bb9d83b`

## Decision

Replace only issue #4's unexecuted disposable-account acceptance boundary. An
independent acceptor will download the exact nominated Darwin/ARM64 artifact,
verify its authority and digest, and exercise it inside a marker-bounded APFS
disk image mounted beneath the acceptor's real home. The supervisor supplies a
synthetic `HOME`, `CODEX_HOME`, temporary directories, fixtures, and Codex auth
state on that volume. Every candidate lifecycle process runs under a reviewed,
fail-closed macOS sandbox that permits candidate writes only to the disposable
volume; lifecycle scenarios deny network,
while the separately identified real-Codex smoke permits only its required
network path. Pre/post protected-state manifests, positive controls, canaries,
and verified detach/delete cleanup prove the intended boundary.

This amendment removes the physical administrator gate without weakening the
exact-artifact, lifecycle, recovery, independent-acceptor, or same-byte release
contracts. It preserves every original PR #15 decision not explicitly replaced
here.

## Needs Attention

- `sandbox-exec` is present but deprecated on supported macOS. Acceptance must
  preflight the exact reviewed profile on the actual OS build, run positive and
  negative controls, and fail closed on unsupported or changed semantics. A
  later OS removal is a fresh design decision, not permission to run unsandboxed.
- This is write-containment for a nonprivileged tool, not a distinct-principal
  security boundary. It does **not** prove a distinct UID, a fresh login session
  or keychain, account teardown, confidentiality from same-UID reads, or
  resistance to a malicious same-UID process.
- Approval authorizes a follow-up implementation PR and a newly nominated
  candidate. The already nominated `0ad07cf...` artifact remains historical
  evidence and must not receive acceptance under the amended contract.

## Decision Spotlight

- **No administrator authority.** Acceptance creates no user, changes no
  account, and uses no `sudo`. My Friday remains entirely unprivileged.
- **Two independent containment proofs.** An APFS mount supplies a disposable
  filesystem boundary; the sandbox denies candidate writes outside that volume.
  Protected-state manifests and canaries demonstrate that the live installation
  and deployed runtime projection did not change.
- **Same UID is disclosed, not disguised.** The accepted boundary shares the
  operator's UID and login keychain. That is proportionate because My Friday
  refuses root and manages files beneath `CODEX_HOME`; it does not manage users,
  login sessions, keychains, or system services.
- **Exact production bytes only.** The harness downloads the nominated artifact
  once, verifies the `artifact-v1` digest, copies it onto the volume, and invokes
  those unmodified bytes for every scenario. There is no test build or fault
  switch.
- **Interruption is externally induced.** The supervisor observes an ordinary
  durable transaction-journal phase on the disposable volume, sends `SIGKILL`
  to the production candidate, then proves ordinary `recover` behavior.
- **Evidence is supervisor-owned, sanitized, and tamper-evident.** The candidate
  and its descendants cannot write evidence staging. An issue comment records
  only exact identities, aggregate manifest equality/digests, controls, outcomes,
  and cleanup proof. Acceptance binds to that comment's immutable ID and body
  digest; secrets, auth files, sensitive paths, and manifest entries stay local
  and are destroyed with the marked run.
- **Cleanup refuses ambiguity.** Detach is non-forced. Deletion occurs only
  after recorded device, volume, image, mount, owner, mode, and run-marker
  proofs still agree; otherwise evidence is preserved for diagnosis.

## Plan Map

- [Context](01-context.md)
- [Decision](02-decision.md)
- [Design](03-design.md)
- [Verification](04-verification.md)
- [Implementation handoff](05-handoff.md)

## Final Gate

The amendment may become `Final` only when issue #4 still owns this acceptance
change, the complete pack has independent maintainer review, the planning PR
and full repository basis are recorded, and product authority approves the
exact final head. That approval supersedes only PR #15's disposable-account
acceptance boundary and authorizes the `through-production` envelope described
here; PR #15 remains the source of all unaffected design decisions.
