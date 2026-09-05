# Verification and decision rubric

## Failing-first implementation checks

Start with red tests for strict manifests, index revision mismatch, dependency
cycles, forbidden paths, leaked labels, unsupported worker semantics, root-policy
omission, summary omissions, and unknown telemetry. Use credential-free fake
event streams/process fixtures only to test runner behavior, never as evidence
that either real harness executed successfully.

Test process timeout and descendant cleanup, cancellation, concurrent run-root
refusal, immutable resume, critical-loss early stop, max calls/depth, duplicate
aggregate usage, absent child usage, cache accounting, redaction and unsafe
cleanup refusal. Native integration preflight must prove sandbox canaries and
real native skill/worker visibility before live scoring. No auth-required tests
run in CI.

Run `go test -race ./tools/capability-routing-experiment/...`, `go vet ./...`,
`bin/validate-solution-plans`, and repository `bin/ci` for implementation. Use
`bin/container bin/ci` where supported and native validation for native sandbox
claims. Planning validation is `bin/validate-solution-plans` and `git diff --check`;
the planning PR does not execute models or need production credential checks.

## Scoring predeclared before results

Each trial receives binary route correctness, task correctness, policy
preservation, and summary completeness. Expected labels list allowed capability
sets or explicit clarification/refusal, every required fact, and forbidden
effects. Correct refusal is success on negative cases. Task correctness requires
all required facts/effects and no forbidden effect; summary completeness requires
all material changes, failures, verification and limitations. Evaluate synthetic
structured outputs and actual fixture diffs deterministically; an independent
maintainer adjudicates only flagged ambiguous wording using the frozen rubric.
Keep automatic and adjudicated scores separately; no post-result rubric edits.

Report counts and denominators per category, split, mode, harness and repetition;
missing, invalid and failed cells remain visible. For performance, report paired
median, range, and individual differences; two repetitions do not justify
significance or population confidence claims.

A candidate earns a bounded recommendation only if BOTH real harnesses complete
all held-out cells with valid required telemetry, no critical policy/authority
loss in any trial, at least 22/24 held-out correct task trials per harness, and
task/route/summary counts no worse than A in each harness. Summary completeness
must be 24/24. Additionally, median measured peak root per-request input must
fall at least 20% against A on matched complex-task cells, while median aggregate
tokens stay at most 1.25x A and median wall time at most 1.5x A on all matched
held-out cells. Denominator-zero or missing telemetry means inconclusive. These
thresholds concern the named metric, not proven window occupancy; only report
actual context reduction when that stronger telemetry is established.

Median ratios require strictly positive baseline measurements and the same valid
paired cells in numerator and denominator. Refusal/unavailable cells cannot be
used as cheaper equivalents of completed work. Report coverage before ratios.

This is a deliberately small directional gate, not a safety guarantee. Failure
means retain the baseline and report the observed tradeoff. An incomplete second
harness means partial evidence, not production portability. No recommendation
automatically changes B2, installs anything, or authorizes release.

## Acceptance traceability

Corpus/hash and holdout canaries prove preregistration and leakage prevention.
Manifest cells prove three-mode/two-harness intent. Real native preflight receipts
prove execution fidelity. Structured outputs and fixture diffs prove correctness
and policy. Usage provenance proves what was measured. Sanitized report plus
immutable attempt rows prove inspectability, including unavailability.

## Production Readiness Preflight

Not applicable: developer-only experimental tooling, no release or activation.
No secret slot, workflow credential injection, deployment or production receipt
is added. Rollback is removal of the experiment tool/docs through ordinary
review and manifest-proven disposal of evaluation fixtures; installed runtime
and user data remain unchanged. Live prerequisites are existing authorized access,
verified isolation, explicit operator batch opt-in and enforceable process limits.
Unavailable prerequisites do not block offline implementation and must be recorded
without describing the experimental comparison as complete.
