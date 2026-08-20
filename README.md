# My Friday

My Friday is a local-first toolkit for creating and safely maintaining a
personalized Codex assistant. It gives technically capable users an inspectable,
version-controlled foundation with separate runtime and governed-memory
repositories.

The product is designed for people who want to understand, repair, move, and
extend their assistant rather than depend on an opaque hosted service. Local
operation is the baseline; hosted source control is optional.

My Friday is currently in product shaping. The approved product direction and
boundaries are documented in [docs/product.md](docs/product.md). Delivery issues
track the selected and deliberately deferred outcomes.

## Development

Use the repo-local commands first:

```sh
bin/container bin/ci
```

If the project does not yet need a container, `bin/ci` should still be the
single local validation entrypoint.

## SDLC

This repository follows a role-based SDLC:

- issues define product intent and lifecycle state
- contributors author temporary Solution Design packs on `design/<issue>-<topic>`
  branches and implementation on `feature/<topic>` branches
- planning pull requests carry the complete design and execution envelope;
  implementation pull requests carry code, validation, reconciliation, durable
  docs promotion, and deploy impact
- maintainers and reviewers approve before merge
- merged release-bearing changes receive independent product acceptance
- accepted immutable candidates are promoted rather than rebuilt
- production releases are recorded by a Git tag and GitHub Release
- staging and production policy live in `docs/operations/sdlc.md`
- rendered verification and evidence policy lives in
  `docs/operations/ui-acceptance.md`

Current system shape lives in `docs/architecture.md`; durable capability
details use `docs/architecture/0000-capability-template.md` as a scaffold.
Temporary issue plans use `docs/plans/_template/` and are removed after their
shipped knowledge is promoted.
Repository-specific release and rollback truth lives in `docs/deployment.md`.
