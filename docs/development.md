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
Before production removal, the acceptance-only named cleanup proves the
manifest-owned instance, permits only the exact private `codex/auth.json`
addition, and holds one no-follow Codex-directory descriptor through validation
and mutation. It atomically moves `auth.json` to a no-replace same-directory
quarantine name, proves that moved entry is the already-opened file, then
atomically transfers that exact inode into a no-replace quarantine entry held
by the manifest-verified instance-root descriptor. The Codex pathname is
re-proved before the transfer is accepted; a replacement restores the file
through the still-open original directory and is reported. Final unlink is not
used: credential contents are truncated and synced through the already-open
verified file descriptor, and the neutralized root quarantine remains for
ordinary manifest-gated instance-root removal. Later pathname replacement is
preserved and reported rather than deleted. Retries recognize and revalidate
the exact deterministic quarantine after either rename or neutralization;
multiple entries, destination collisions, or mismatched identity remain
untouched. The file must also be a
current-user-owned, mode-`0600`, single-link regular file. Deterministic file-
and directory-replacement races, symlinks, hardlinks, wrong metadata, alternate
credential paths, and unrelated private Codex entries are preserved and
reported.
Immediately before each rename or descriptor-bound neutralization, cleanup
re-verifies the manifest, every required owned root, managed Codex config and
instructions, exact root/Codex entry sets with every expected name observed
once and no extras, and held directory identities.
Acceptance requires the current user's disposable instance tree to remain
quiescent after those checks return. Consistent with ADR 0002, this detects
ordinary replacement but does not claim isolation from an actively malicious
same-UID process that can continuously rewrite owner-controlled directories.
For the credentialed TUI smoke, the supervisor captures a metadata-only,
no-follow receipt only after the bounded Codex process exits, descendants are
reaped, and no cooperating writer remains. The receipt binds the candidate,
run, instance-root and Codex-directory identities plus every generated cache,
session, log, database, and other non-managed entry. Cleanup requires that exact
device/inode/type/owner/mode/link-count/size/mtime tree on every revalidation.
Regular generated files are limited to the observed `0600`, `0644`, `0664`,
and `0755` modes beneath the separately verified owner-only instance and Codex
directories; special bits and every other mode refuse. The private ancestors
make group-readable or group-writable plugin-cache metadata unreachable to
other users while the exact captured mode remains drift-bound.
Codex may create single-link argument-zero helper symlinks; the receipt accepts
only `tmp/arg0/codex-<alphanumeric>/apply_patch`, `applypatch`, and
`codex-execve-wrapper` links whose captured target is the manifest-bound
instance Codex executable, and binds that target on every replay. A new,
missing, replaced,
hard-linked, alternate-target, or otherwise unbound symlink is preserved and
refused. The
helper receives the same reviewed Git-capable PATH as candidate lifecycle
commands so complete manifest verification cannot depend on ambient shell PATH.
The complete receipt and manifest authority are replayed once more immediately
before unchanged production removal planning.
Ambient preservation evidence compares the stable no-follow Codex tree and the
protected runtime tree. It deliberately omits `sessions/` and the versioned
`logs_<n>.sqlite*` and `state_<n>.sqlite*` operational databases from whole-tree
metadata equality because the acceptor may itself be an active Codex session;
those paths have a legitimate writer that is unrelated to the candidate.
Ambient `auth.json` retains a separate exact metadata identity check, protected
config/instructions/skills and unrelated state remain in the stable snapshot,
and launcher/CODEX_HOME isolation tests prove the candidate never receives the
ambient Codex home.
The machine authority names this proof
`ambient_codex_stable_subset_equal` and carries the exact ordered
`ambient_codex_metadata_excluded` classes; the superseded broad
`ambient_codex_equal` field is invalid evidence.
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
an early composer, an MCP boot-progress state, repeated boot-time composers,
spinner titles, the post-progress plain `OSC 0;workspace` title, and finally the
stable composer. Persistent harmless redraw output continues after readiness,
so a quiescence-based driver cannot pass. The fixture accepts the exact prompt,
echoes its exact prefix and `token.` suffix around an ANSI cursor-position
sequence, discards any Enter bytes queued before the suffix, and accepts
`ESC [ 13 ; 1 u` only after both fragments. It also requires the
pre-spawn child geometry to be exactly 30×120. This prevents zero-size
rendering, first-composer timing, blind-delay or quiescence timing, fragmented
prompt rendering, adjacent prompt/Enter ordering, and
line-discipline CR-to-NL translation from making an invalid PTY driver appear
correct.
Named-instance regressions prove the private Codex config TOML-escapes special
workspace/runtime paths, trusts only the exact instance workspace, adds only
the private runtime as extra workspace-write authority, fixes approvals/network,
and separates two instances. They bind an instance-specific builder and exact
currently executing My Friday candidate copy, refuse executable/config/builder
drift, exercise v1 and legacy-v2 revision upgrades and bounded rollback
reversal, and recover an interruption after execution-context mutation.
Adversarial migration tests substitute candidate staging and each rollback
quarantine, and prove foreign quarantine collisions remain untouched. Purpose
instructions remain manifest-bound, and verify,
remove, and recovery authority is denied after managed-state tampering. Reversal tests preserve ambient user Codex state. The
acceptance contract forbids scripting an onboarding response.
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
migration success plus injected legacy-cleanup failure. A subprocess flow
against a fixture `$HOME` is intentionally not used because the production CLI
derives authority from the operating-system account home; acceptance exercises
the exact managed executable and read-only builder commands in a disposable
real-home instance.

Capability contract and lifecycle tests live under `internal/capability`.
They use deterministic temporary fixtures and never invoke Codex or the
network. Run `go test -race ./internal/capability` while iterating and the full
`bin/container bin/ci` before review.

Issue-51 release authority has a separate deterministic contract suite:

```sh
bin/test-capability-workshop-evidence
bin/test-acceptance-contract
```

These checks lock the strict provisional, final, failure, and three-person
workshop receipt schemas; exact issue/candidate/artifact bindings; stable-
comment re-fetches; cross-schema refusal; and product-acceptance/release
routing. They are network/model-free and do not replace the native Apple-
silicon run of `bin/accept-capability-workshop` against nominated bytes.
The shared acceptance-support and runner suites additionally lock no-follow
auth copy/source-swap refusal, secure-root collision and ambiguous cleanup
preservation, timeout/escaped-child reaping, `INT`/`TERM` exit status and
ordinary/escaped-descendant quiescence, and the capability stop barrier's
canonical post-mutation journal boundary.
`bin/test-launcher-pty-capture` additionally proves that private launcher tasks
receive a real stdout TTY, capture no public transcript, retain owner-only
transcript permissions and child exit status, and remain inside the runner's
timeout/signal descendant-reaping boundary.
It also locks empty-forwarded-argv launch, the CSI-u/composer/MCP/title/composer
readiness sequence, separate prompt typing and protocol-Enter submission,
pre-submission marker refusal, output-only marker observation, prompt self-match
refusal, missing-marker and nonzero failure, bytewise observation through
invalid UTF-8/control bytes, marker-triggered TUI/descendant closure, and the
split 180-second deterministic-command / 600-second named-launcher timeout
contract.
This integration runs on the required Apple-silicon acceptance host and skips
on portable CI hosts without `/usr/bin/expect`; the remaining contract and
cross-platform runner tests still run there.
