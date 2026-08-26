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

## Named assistant instances

Create a private named instance from a validated runtime/memory pair. The
launcher directory must already exist as `$HOME/.local/bin`:

```sh
./my-friday assistant create alfred --runtime /path/to/runtime --memory /path/to/memory
$HOME/.local/bin/alfred
./my-friday assistant verify alfred
./my-friday assistant remove alfred
```

Mutations print a plan and require the exact action word. Each instance owns
separate `codex`, `runtime`, `memory`, `workspace`, and `dependencies`
directories. Its native launcher is the sole outside projection. It preserves
real `HOME`, sets instance `CODEX_HOME`, fixes Codex `--cd` to the workspace,
and forwards literal arguments. The private `codex/config.toml` trusts that
instance's exact absolute workspace. Contract v2 also uses workspace-write,
adds only the private runtime as a writable root, sets approvals to never, and
disables network for the managed builder context. Put file-backed credentials only in that
instance's `codex` directory; My Friday does not read, copy, or report them.
To migrate a manifest-owned legacy projection, use the same arguments with
`assistant migrate`; My Friday verifies the named replacement before executing
the legacy projection's separately previewed, manifest-proven uninstall.

## Legacy managed Codex baseline

Install the generated runtime's global instructions after reviewing the full
preview:

```sh
./my-friday codex install --runtime /path/to/my-friday-runtime
./my-friday codex verify
```

These commands retain the prior single-home projection's repair and rollback
lifecycle for explicit recovery. The confirmation word is the exact case-sensitive action name followed by
Return; surrounding whitespace and unterminated input are refused. My Friday owns only
`$CODEX_HOME/AGENTS.md` and `$CODEX_HOME/.my-friday`; it never edits Codex
configuration, authentication, sessions, logs, skills, packages, or project
configuration. See [the installed-baseline contract](docs/architecture/installed-codex-baseline.md).

## Instruction-only capabilities

Each named assistant includes a private capability builder. It may edit and
review versioned source under its exact private `runtime/skills/<slug>/`, and
uses its manifest-owned My Friday copy only for inspect, validate, and test;
activation remains
an independent CLI action:

```sh
my-friday capability inspect alfred daily-brief --plain
my-friday capability validate alfred daily-brief
my-friday capability test alfred daily-brief
my-friday capability install alfred daily-brief
my-friday capability verify alfred daily-brief
my-friday capability disable alfred daily-brief
my-friday capability enable alfred daily-brief
my-friday capability remove alfred daily-brief
```

Mutations preview exact digests and paths, require an interactive terminal and
case-sensitive action word, and affect only the selected instance. Source
survives disable and remove. Lifecycle changes apply to a fresh Codex task.
See [the capability workshop contract](docs/architecture/capability-workshop.md).

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
