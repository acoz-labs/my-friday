# Deployment

## Delivery profile

My Friday is an `artifact`: one native macOS/ARM64 command. It has no staging
service or production environment. This implementation does not publish an
artifact or change repository release configuration.

Future nomination must build one immutable binary from an exact successful
commit, record its SHA-256 digest and provenance, and independently exercise it
on supported Apple silicon/APFS/Git before acceptance. The accepted bytes—not a
rebuild—must be published.

Public release remains blocked on naming clearance, dependency licence notices,
minimum-macOS verification, signing/notarization and distribution decisions,
configured release variables, rollback instructions, and an independent
acceptor. No configuration or secrets are used at runtime.

Rollback of source is a Git revert. It must never delete repositories already
created by a user. Contract-v1 validation must remain available after a future
release even if creation evolves.
