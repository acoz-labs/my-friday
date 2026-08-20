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
| `02-unicode-success.txt` | Combined-parent creation with NFC Unicode profile | Ordered preview/progress and a verified pair |
| `03-path-collision.txt` | Nested separate targets | Field-local rejection before mutation |
| `04-rollback.txt` | Injected post-validation failure | Automatic rollback restores pre-run state |
| `05-partial-promotion-recovery.txt` | Injected verified-phase interruption | Retained journal drives verified cleanup recovery |
| `06-already-complete.txt` | Exact completed rerun | Distinct `Already complete` result with no writes |

Environment evidence: generated natively on Apple Silicon macOS/APFS with the
repository-pinned Go and Git toolchain. `bin/ci` also builds `darwin/arm64` and
scans production imports; the command has no networking package, external
schema loader, telemetry, credential, hosted-account, commit, or remote path.
The Git subprocess boundary is fixed and covered by repository tests.

Accessibility evidence: all scenarios use ordinary line input and keyboard-only
navigation. A byte scan rejects ESC characters, and the interface uses no
cursor addressing, screen clearing, hidden focus, color-required meaning,
animation, or time-dependent text. A hands-on VoiceOver result is intentionally
left for independent candidate acceptance; contributor automation does not
claim VoiceOver certification.

Verdict: pass for contributor terminal behavior, keyboard order, deterministic
text, control-sequence absence, environment boundary, and recovery messaging.
Independent maintainer/product-design review and later immutable-candidate
VoiceOver acceptance remain unreviewed surfaces.
