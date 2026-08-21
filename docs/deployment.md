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

If an incorrect stable alias is proven, resolve its exact asset ID and delete
only `my-friday-darwin-arm64.tar.gz` from that release. Retain the tag, release
notes, commit-suffixed archive, `SHA256SUMS`, and acceptance evidence. Re-run the
guarded backfill only after the retained source passes every verification.

Every new candidate still requires successful CI, exact-candidate nomination,
independent Apple silicon acceptance, and the artifact release gate. No
configuration or secrets are used at runtime.

Rollback of source is a Git revert. It must never delete repositories already
created by a user. Contract-v1 validation must remain available after a future
release even if creation evolves.
