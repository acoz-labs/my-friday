# Deployment

## Delivery profile

My Friday is an `artifact`: one native macOS/ARM64 command. It has no staging
service or production environment. Public releases expose the stable URL
`https://github.com/acoz-labs/my-friday/releases/latest/download/my-friday-darwin-arm64.tar.gz`.

Nomination builds one immutable binary from an exact successful commit, uploads
it as the `my-friday-darwin-arm64` Actions artifact, and binds its run, artifact
ID/name, and SHA-256 in an `artifact-v1` authority. Acceptance remains bound to
that executable digest. Release downloads only that exact artifact, verifies
the executable, and packages it without rebuilding.

The public archive contains one executable named `my-friday`. Its metadata and
gzip header are normalized so repeated packaging of accepted bytes is
identical. Every release publishes:

- `my-friday-darwin-arm64.tar.gz`, whose stable name powers the latest URL;
- `SHA256SUMS`, with exactly one archive digest and one contained-executable
  digest.

The archive digest and executable digest prove different layers and must never
be substituted for each other. `bin/finalize-release` uploads missing assets,
treats an existing equal-digest asset as a successful retry, and refuses an
existing mismatched or duplicate name. It never uses clobber.

To verify a downloaded release:

```sh
sha256sum -c SHA256SUMS --ignore-missing
tar -xzf my-friday-darwin-arm64.tar.gz
sha256sum my-friday
```

The first result verifies the archive entry; the second must match the
`my-friday` entry in `SHA256SUMS`.

### Current-release stable-name backfill

Run the `Backfill stable release asset` workflow only against the exact current
latest tag, its retained commit-suffixed archive, and the accepted executable
digest recorded in the release ledger. The command verifies that the release
is published and latest, requires exactly one `- Artifact:` ledger entry in the
closed `sha256:<digest>` form, and binds that authority to the requested digest
and contained executable. It also requires one source archive and checksum
asset, verifies the source archive checksum and its sole safe executable, then
uploads the unchanged archive bytes under the stable name. An equal existing
alias is an idempotent success; a mismatch fails without replacement.

Future archives contain the canonical top-level executable `my-friday`. The
one retained pre-contract archive is also valid for backfill when its sole
top-level regular executable uses the historical
`my-friday-darwin-arm64-<hex>` name. This compatibility does not rename or
repackage the archive: the stable alias receives the original archive bytes.
That legacy macOS archive may also contain one exact top-level AppleDouble
companion named `._<executable-name>`. The guard permits only that paired
regular metadata member, extracts only the executable for digest verification,
and still copies the complete original archive bytes to the stable alias.

If an incorrect stable alias is proven, resolve its exact asset ID and delete
only `my-friday-darwin-arm64.tar.gz` from that release. Retain the tag, release
notes, commit-suffixed archive, `SHA256SUMS`, and acceptance evidence. Re-run the
guarded backfill only after the retained source passes every verification.

Every new candidate still requires successful CI, exact-candidate nomination,
independent Apple silicon acceptance, and the artifact release gate. No
configuration or secrets are used at runtime.

Named-instance acceptance must exercise create, verify, launch, two-instance
coexistence, collision refusal, interrupted-remove recovery, and reversal as
the current non-root user. The accepted `my-friday` bytes become the launcher;
Codex is copied into the instance dependency directory. A separately
credentialed live smoke cannot substitute for credential-free containment.

Installed-baseline acceptance runs from a clean checkout at the freshly
nominated candidate SHA:

```sh
MY_FRIDAY_RUNTIME_PROJECTION=/absolute/path/to/deployed/runtime \
MY_FRIDAY_CODEX_AUTH_FILE=/absolute/path/to/existing/auth.json \
  bin/accept-installed-codex-baseline \
  'artifact-v1:run=…:id=…:name=my-friday-darwin-arm64:sha256=…' 4
```

The acceptor supplies the regular, current-user-owned `auth.json` from an
existing Codex login. The supervisor does not perform or refresh a login and
does not read credential bytes into shell variables, argv, output, or evidence.
It first proves an ambient bounded model call, copies the file byte-for-byte to
one disposable named instance's `CODEX_HOME`, then launches the native instance
with no forwarded arguments under a bounded interactive PTY. The PTY submits a
purpose question and must observe the unique installed-purpose token from the
instance-owned Codex executable. That smoke receives the reviewed
`TERM=xterm-256color` explicitly, so a hostile or `dumb` caller terminal cannot
divert the purpose prompt into Codex's terminal-safety confirmation. Other
candidate lifecycle commands retain their existing sanitized environment. The
named instance's manifest-bound private `codex/config.toml` trusts only its
exact absolute workspace, so the smoke sends no first-run workspace-trust
answer. After Codex explicitly enables CSI-u enhanced keyboard reporting, the
driver submits the literal purpose prompt with the protocol Enter sequence
`ESC [ 13 ; 1 u`; a raw carriage return is not accepted as equivalent. The
driver then closes its PTY, with the bounded process runner retaining descendant
cleanup authority. Approval and sandbox policy remain Codex defaults. The
copy is removed with the instance. The
same file-backed OAuth credential is deliberately not routed through
`codex login --with-api-key`.

The command requires
ordinary-user Apple silicon macOS, APFS, GitHub comment authority, Codex, Go,
and the reviewed `sandbox-exec` behavior. It uses randomized leaves under the
current account's real named-instance root and launcher directory, independently
of caller `HOME`. It exercises create, verify, PTY launch, two-instance
isolation, collision preservation, interrupted-remove recovery, complete
reversal, ambient-state preservation, and real-Codex instruction discovery. It
prints an `evidence-v1` authority only after every instance root, launcher,
credential copy, APFS helper image, run root, and evidence root is proven absent.
The supervisor sets umask `077` before local mutation, so caller umask cannot
weaken sparse-image or evidence permissions. Mounted APFS root identity checks
the currently observed link count as a positive integer no greater than 64 at
each proof; it is not immutable because top-level directory creation changes it.
The final pre-detach observation is retained in strict evidence.
Failure cleanup uses manifest authority for instance roots and launchers and
no-follow device/inode/digest receipts for the temporary collision and unrelated
launcher-sibling leaves. It is bounded, runs at every post-create failure phase,
never removes the ambient auth source, and reports the cleanup facts it actually
proved rather than converting a partial cleanup into approval authority.
Receipt path and acceptance-namespace scope are validated before mutation. Once
that scope is valid, a drifted foreign leaf is preserved and reported but cannot
prevent independent manifest-proven instance cleanup or deletion of the copied
credential; other exact leaves are still attempted and refusals are aggregated.

Product acceptance accepts that authority, not an opaque evidence URL. Release
finalization re-fetches both evidence comments and rejects edits, deletion,
author/candidate/artifact mismatch, missing cross-binding, or a provisional
record without final cleanup proof. A candidate nominated before this
acceptance contract remains historical and cannot use the amended path.
An acceptance failure publishes a redacted, strict `failure-v1` issue comment
authority when GitHub is reachable. That authority permits an independent
acceptor to durably record `changes-required`; it can never authorize approval
or release. It distinguishes candidate behavior from harness/environment
failure and records only a redacted failure class, run binding, preservation,
and safe-detach result.

The successful `named-instance-acceptance-evidence-v1` authority uses exact
typed provisional/final schemas. It binds
archive/executable hashes, transitive helper build closure, platform and APFS
graph, normalized-profile controls, expected state/exit class for every
named-instance scenario, the separate file-backed smoke semantics, process
quiescence, protected metadata/content counts and digests, and the cleanup set.
Preservation claims are limited to measured state: live Codex/runtime snapshots,
a disposable caller-`HOME` shell canary, an unrelated real launcher sibling,
the foreign collision leaf, and unchanged ambient-auth metadata.
The verifier rejects the superseded single-home evidence schema. Both comments
are fetched twice for stability; final
protected-state and provisional body digests must agree exactly.

Rollback of source is a Git revert. It must never delete repositories already
created by a user. Contract-v1 validation must remain available after a future
release even if creation evolves.
