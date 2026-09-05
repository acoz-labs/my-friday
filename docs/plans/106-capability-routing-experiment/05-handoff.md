# Implementation handoff

## Smallest complete outcome

A standalone Go developer tool validates and stages a frozen synthetic suite,
executes only explicitly opted-in supported native cells, scores reproducibly,
and exports a sanitized comparison that distinguishes failure and missing
evidence. No production code path or default changes.

## Reviewable order and ownership

One task owns `tools/capability-routing-experiment/`, its `testdata/`,
`docs/development.md` experiment entry, and
`docs/experiments/capability-routing/` method/results. Do not modify existing
capability schemas, installed instances, public CLI, #101/#103 plans or #92 work.

1. Red tests for corpus/manifest/score contracts; implement offline preparation,
   deterministic revision-bound retrieval and reporting. Maintainer reviews
   synthetic labels and preregistration before any live data collection.
2. Credential-free subprocess/telemetry fixtures and red isolation/timeout/
   accounting tests; implement only the two named experiment drivers and native
   preflight. Reuse repository primitives where suitable without refactoring
   unrelated acceptance infrastructure.
3. Freeze corpus, prompts, model settings, budgets and hashes in a reviewed
   commit. Perform opt-in live batches only on verified available routes;
   retain unavailable cells and stop on safety failure. Export sanitized
   evidence, independent score review and bounded conclusions.
4. Reconcile implementation to this approved plan, promote method/operation
   docs, and delete this temporary plan before implementation review completes.

Offline tool completion and partial report can be reviewed while access is
unavailable. Keep #106 open with an explicit outstanding real-harness comparison
when necessary; do not label a stubbed/partial comparison complete. No new
product approval is needed for ordinary implementation choices within this
contract; material scope, authority or paid-access changes return upstream.

## Documentation promotion

| Concern | Durable destination | Action |
| --- | --- | --- |
| Commands, offline tests and live opt-in | `docs/development.md` | Update |
| Frozen method, schemas, limits and interpretation | `docs/experiments/capability-routing/README.md` | Create |
| Sanitized machine-readable and readable results | `docs/experiments/capability-routing/results/<run-id>/` | Create after allowlist review |
| Why this did not change production defaults | Experiment method with #106/#105 provenance | Record; no production ADR needed |
| Temporary design pack | This issue plan | Remove during reconciliation |

Results must bind to the exact runner/corpus/preregistration revisions. Raw logs
and private sources never enter tracked results. PR body links the approved plan,
lists exact checks and evidence gaps, and distinguishes runner validation from
real model execution. Maintainer review remains independent.

## Envelope and residual limitations

`implementation`: code/tests/docs and explicitly opted-in bounded synthetic
evaluation only. No staging, release, runtime installation, identity changes,
new credentials, production adapter support or default migration. Confidence is
medium: offline feasibility is high, real-harness availability and complete
telemetry are deliberately reported prerequisites rather than assumed support.
