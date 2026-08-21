# Deployment

## Delivery profile

My Friday is an `artifact`: one native macOS/ARM64 command. It has no staging
service or production environment. This implementation does not publish an
artifact merely by merging.

Artifact nomination builds one `my-friday-darwin-arm64` binary from the exact
successful commit, uploads it as a named Actions artifact, and records the run,
artifact ID, and SHA-256 digest. Acceptance is bound to that authority string.
Release downloads the same run's artifact, re-verifies its digest, and uploads
those exact bytes as the GitHub Release asset; it never rebuilds the candidate.

Installed-baseline acceptance must use the nominated binary under a disposable
non-admin macOS 14+ Apple-silicon user with a fresh home, keychain, and
`CODEX_HOME`. The operator supplies admin authentication only to create and
remove that marker-bounded user. Acceptance exercises install, collision,
verify, repair, upgrade, rollback, interruption recovery, uninstall reversal,
real Codex instruction discovery, unrelated canaries, logout, and complete
identity/home teardown. Evidence excludes `auth.json` and secret values.

Production publication still requires configured release variables, successful
checks, independent exact-candidate acceptance, and issue-specific release
authority. No configuration or secrets are used by the released lifecycle.

Rollback of source is a Git revert. It must never delete repositories already
created by a user. Contract-v1 validation must remain available after a future
release even if creation evolves.
