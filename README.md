# My Friday

My Friday is a local-first toolkit for creating and safely maintaining a
personalized Codex assistant. It gives technically capable users an inspectable,
version-controlled foundation with separate runtime and governed-memory
repositories.

The product is designed for people who want to understand, repair, move, and
extend their assistant rather than depend on an opaque hosted service. Local
operation is the baseline; hosted source control is optional.

The first command previews and creates the two repositories locally. It does
not install Codex, contact a hosted service, read secrets, import private
content, create commits, or configure remotes.

## Quick start

The supported pilot environment is macOS 14 or later on Apple silicon, a local
APFS volume, an interactive UTF-8 terminal, and Git 2.28 or later.

```sh
mise install
go build -o my-friday ./cmd/my-friday
./my-friday init
```

Review the complete plan, then type the case-sensitive word `Create` to write
anything. Return at confirmation safely exits. Validate a generated pair with:

```sh
./my-friday validate --runtime /path/to/my-friday-runtime --memory /path/to/my-friday-memory
```

The broader product direction and deferred outcomes remain in
[docs/product.md](docs/product.md).

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

## License

My Friday is available under the [MIT License](LICENSE).
