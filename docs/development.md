# Development

Use the repo-local validation entrypoint:

```sh
bin/container bin/ci
```

If container support is not ready for this repo yet:

```sh
bin/ci
```

My Friday uses Go 1.26.4, pinned by `mise.toml`; `go.mod` declares the module's
language baseline. Install the exact host toolchain with `mise install`.

When host-local language execution is supported, commit exact versions in a
root `mise.toml` and install them with:

```sh
mise install
```

The complete check runs solution-plan validation, formatting, vet, race-enabled
tests, acceptance-evidence contract tests, a real no-admin APFS/sandbox
primitive test on Apple silicon macOS, and a static Darwin/ARM64 build with
`bin/ci`. The primitive test skips on non-macOS hosts; it must pass natively
before an installed-baseline candidate is nominated.
`bin/test-acceptance-contract` locks the supervisor's fixture grammar, early
secret removal, bounded runner, working-byte checks, mount authority, profile
placeholder, and durable failure-evidence requirements. Go tests under `tools/`
exercise valid contract-v1 rendering, Scheme escaping, stop-journal validation,
candidate/Codex profile launch, process-start-bound setsid descendant cleanup,
filesystem-linked stop receipts, no-follow root/marker cleanup, strict evidence
semantics/re-fetches, and process-group timeout behavior.

`bin/container bin/ci` proves portable source compilation but cannot prove
APFS, macOS permissions, terminal, or local Git-template behaviour. Run
`bin/ci` natively on Apple silicon before review and artifact nomination.

Focused commands are `go test ./...`, `go test -race ./...`, and
`GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build ./cmd/my-friday`.

Installed-baseline tests must pass explicit temporary Codex-home and runtime
roots to `internal/codexhome`; they must never target a developer's effective
`CODEX_HOME` or `~/.codex`. Only the command layer resolves the production home.
