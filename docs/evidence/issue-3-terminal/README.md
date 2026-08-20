# Issue 3 Terminal Evidence

This directory is the contributor evidence package for the implementation pull
request head that contains it. Git binds these generated artifacts to that
exact source tree; temp-root prefixes are consistently replaced with `<TEMP>`.

Regenerate the six scenarios from the repository root:

```sh
MY_FRIDAY_EVIDENCE_DIR="$PWD/docs/evidence/issue-3-terminal" \
  go test ./internal/terminal -run TestGenerateTerminalEvidence -count=1 -v
```

| File | Scenario | Expected result |
|---|---|---|
| `01-default-exit.txt` | Return at confirmation | `No changes made`; no target or support writes |
| `02-unicode-success.txt` | Combined-parent creation with NFC Unicode profile | Ordered preview/progress, exactly two valid targets, and no support residue |
| `03-path-collision.txt` | Nested separate targets | Field-local rejection before mutation |
| `04-rollback.txt` | Injected post-validation failure through the wizard adapter | Automatic rollback restores the full adjacent-root snapshot, including hidden paths |
| `05-partial-promotion-recovery.txt` | Injected verified-phase interruption through the wizard, then shipped `recover` adapter | Retained journal drives recovery to exactly two valid targets with no transaction residue |
| `06-already-complete.txt` | Exact completed rerun | Distinct `Already complete` result with no writes |

Environment evidence: generated natively on Apple Silicon macOS/APFS with the
repository-pinned Go and Git toolchain. Tests exercise the supported contract
and each architecture, OS, terminal, filesystem, and Git denial. The committed
evidence test scans every transcript byte for ESC and checks scenario-specific
receipts. A production-source AST check enforces an exact import boundary and
permits only literal `git` subprocesses. One shared observer captures every
shipped Git argv/environment, including the environment preflight version
probe; tests enumerate each permitted argv shape and value and reject every
other operation. The command has no external schema loader,
telemetry, credential, hosted-account, commit, or remote-creation path.

Accessibility evidence: all scenarios use ordinary line input and keyboard-only
navigation. A table-driven test exercises `b`, `q`, and EOF at every applicable
pre-mutation prompt, including Scope and retained-transaction recovery, proves
that Back re-emits its destination prompt, continues with sentinel answers to
creation, and compares the complete adjacent filesystem state. A byte scan rejects every ESC byte, and the
interface uses no cursor addressing, screen clearing, hidden focus,
color-required meaning, animation, or time-dependent text. A hands-on VoiceOver
result is intentionally left for independent candidate acceptance; contributor
automation does not claim VoiceOver certification.

Verdict: pass for contributor terminal behavior, keyboard order, deterministic
text, control-sequence absence, environment boundary, and recovery messaging.
Independent maintainer/product-design review and later immutable-candidate
VoiceOver acceptance remain unreviewed surfaces.
