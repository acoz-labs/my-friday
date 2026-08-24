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
`bin/test-acceptance-contract` locks the supervisor's fixture grammar,
file-backed authentication without a login command, named-instance scenario
matrix, genuine empty-argv interactive PTY smoke, bounded manifest-proven
failure cleanup, caller-shell and launcher-sibling canaries, working-byte checks,
mount authority, profile
placeholder, and durable failure-evidence requirements. Go tests under `tools/`
exercise valid contract-v1 rendering, Scheme escaping, stop-journal validation,
candidate/Codex profile launch, process-start-bound setsid descendant cleanup,
filesystem-linked stop receipts, no-follow root/marker cleanup, strict evidence
semantics/re-fetches, legacy-schema refusal, and process-group timeout behavior.
`tools/acceptance-support` cleanup regressions exercise every post-create phase,
including copied credentials and launcher-absent recovery, while proving the
ambient auth source remains unchanged.
They also prove a drifted collision/sibling receipt is reported and preserved
without blocking manifest-proven instance and copied-credential cleanup, while
an out-of-scope receipt remains a fail-closed pre-mutation error.
Contract regressions also require an internal `077` umask before local mutation
and a current APFS mounted-root link count in the safe 1–64 range at every
authority check. The native primitive changes that count by creating root
directories, then proves detach authority still succeeds; only the final
pre-detach observation is evidence-bound.
The PTY contract fixes `TERM=xterm-256color` only for the interactive launcher
smoke; native primitive coverage starts from `TERM=dumb` and proves the
empty-argument smoke still delivers its purpose prompt and observes its token.
The native fixture places its pseudo-terminal in raw mode, emits Codex's
`CSI > 7 u` enhanced-keyboard enablement, discards premature input, then emits
the composer-ready marker and accepts only the exact prompt bytes followed by
`ESC [ 13 ; 1 u`. It also requires the pre-spawn child geometry to be exactly
30×120. This prevents zero-size rendering, blind-delay timing, and
line-discipline CR-to-NL translation from making an invalid PTY driver appear
correct.
Named-instance regressions prove the private Codex config TOML-escapes special
workspace paths, trusts only the exact instance workspace, separates two
instances, and denies verify, remove, and recovery authority after config
tampering. The acceptance contract forbids scripting an onboarding response.
The native primitive performs a fresh valid install under the final lifecycle
profile (including Git/xcrun), while regressions cover immediate double-fork
setsid escape, ambiguous-attach preservation, exact-entry cleanup refusal,
intermediate-symlink protected-read refusal, closure-producer failure, and exact
versioned diagnostic semantics. Evidence adversarial cases also require each
protected metadata/content/canary equality field to be the JSON boolean `true`,
not a truthy string or number.

`bin/container bin/ci` proves portable source compilation but cannot prove
APFS, macOS permissions, terminal, or local Git-template behaviour. Run
`bin/ci` natively on Apple silicon before review and artifact nomination.

Focused commands are `go test ./...`, `go test -race ./...`, and
`GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build ./cmd/my-friday`.

Installed-baseline tests must pass explicit temporary Codex-home and runtime
roots to `internal/codexhome`; they must never target a developer's effective
`CODEX_HOME` or `~/.codex`. Only the command layer resolves the production home.

Named-instance coverage lives in `internal/assistantinstance`. Use
`go test -race ./internal/assistantinstance ./cmd/my-friday` while iterating.
Fixtures create their own roots, launcher siblings, repository pairs, and
credential-free executable stubs; they never inspect or mutate developer
Codex, shell, launcher, or credential state.
Regression coverage includes PATH symlink resolution, caller-`HOME` refusal as
authority, forged managed-executable manifests, same-name concurrency, and
migration success plus injected legacy-cleanup failure.
