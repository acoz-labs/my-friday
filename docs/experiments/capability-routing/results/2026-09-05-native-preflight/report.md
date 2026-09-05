# Capability routing comparison

- Corpus revision: `routing-corpus-v1`
- Manifest SHA-256: `c87d185281be47be6e2ec07978803dfb736f06b5adcfa4bd964f72a3834f0ee8`
- Executing runner revision: `` (available: false, modified: false)
- Runner provenance limitation: Go build information did not provide a VCS revision
- Recommendation: **inconclusive**
- Conclusion: At least one held-out harness/mode cell is incomplete or lacks required telemetry; retain the native baseline and make no cross-harness claim.

## Native-driver fidelity

| Harness | Version | State | Missing controls |
| --- | --- | --- | --- |
| claude | 2.1.193 (Claude Code) | unavailable | no OS-enforced fixture-only read boundary was demonstrated without exposing host authentication; native skill body reads were not proven constrained to the staged allowlist; native worker identity/inheritance and pre-launch depth/count enforcement were not demonstrated; enabled built-in tools lack an equivalent pre-dispatch eight-call rejection hook; immediately detached descendants cannot be proven contained between process-table samples |
| codex | codex-cli 0.153.4 | unavailable | no OS-enforced fixture-only read boundary was demonstrated without exposing host authentication; native skill body reads were not proven constrained to the staged allowlist; native worker identity/inheritance and pre-launch depth/count enforcement were not demonstrated; enabled built-in tools lack an equivalent pre-dispatch eight-call rejection hook; immediately detached descendants cannot be proven contained between process-table samples |

## Coverage and scores

| Harness | Mode | Split | Rep | Declared | Complete | Failed | Unavailable | Invalid | Missing | Full tokens | Wall | Peak root input | Window occupancy | Route | Task | Policy | Summary |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| claude | lookup-direct | development | 1 | 12 | 0 | 0 | 12 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-direct | development | 2 | 12 | 0 | 0 | 12 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-direct | held-out | 1 | 12 | 0 | 0 | 12 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-direct | held-out | 2 | 12 | 0 | 0 | 12 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-worker | development | 1 | 12 | 0 | 0 | 12 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-worker | development | 2 | 12 | 0 | 0 | 12 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-worker | held-out | 1 | 12 | 0 | 0 | 12 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-worker | held-out | 2 | 12 | 0 | 0 | 12 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | native-catalogue | development | 1 | 12 | 0 | 0 | 12 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | native-catalogue | development | 2 | 12 | 0 | 0 | 12 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | native-catalogue | held-out | 1 | 12 | 0 | 0 | 12 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | native-catalogue | held-out | 2 | 12 | 0 | 0 | 12 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-direct | development | 1 | 12 | 0 | 0 | 12 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-direct | development | 2 | 12 | 0 | 0 | 12 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-direct | held-out | 1 | 12 | 0 | 0 | 12 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-direct | held-out | 2 | 12 | 0 | 0 | 12 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-worker | development | 1 | 12 | 0 | 0 | 12 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-worker | development | 2 | 12 | 0 | 0 | 12 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-worker | held-out | 1 | 12 | 0 | 0 | 12 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-worker | held-out | 2 | 12 | 0 | 0 | 12 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | native-catalogue | development | 1 | 12 | 0 | 0 | 12 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | native-catalogue | development | 2 | 12 | 0 | 0 | 12 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | native-catalogue | held-out | 1 | 12 | 0 | 0 | 12 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | native-catalogue | held-out | 2 | 12 | 0 | 0 | 12 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |

## Category denominators

| Harness | Mode | Split | Category | Rep | Declared | Complete | Failed | Unavailable | Invalid | Missing | Route | Task | Policy | Summary |
| --- | --- | --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| claude | lookup-direct | development | ambiguous-alternatives | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-direct | development | ambiguous-alternatives | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-direct | development | complex-worker-work | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-direct | development | complex-worker-work | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-direct | development | conflicting-instructions | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-direct | development | conflicting-instructions | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-direct | development | dependency-loading | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-direct | development | dependency-loading | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-direct | development | explicit-selection | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-direct | development | explicit-selection | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-direct | development | informal-paraphrase | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-direct | development | informal-paraphrase | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-direct | development | material-summary-omission | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-direct | development | material-summary-omission | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-direct | development | no-match | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-direct | development | no-match | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-direct | development | permission-denial | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-direct | development | permission-denial | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-direct | development | short-direct-work | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-direct | development | short-direct-work | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-direct | development | stale-index | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-direct | development | stale-index | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-direct | development | unsupported-required-semantics | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-direct | development | unsupported-required-semantics | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-direct | held-out | ambiguous-alternatives | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-direct | held-out | ambiguous-alternatives | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-direct | held-out | complex-worker-work | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-direct | held-out | complex-worker-work | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-direct | held-out | conflicting-instructions | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-direct | held-out | conflicting-instructions | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-direct | held-out | dependency-loading | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-direct | held-out | dependency-loading | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-direct | held-out | explicit-selection | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-direct | held-out | explicit-selection | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-direct | held-out | informal-paraphrase | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-direct | held-out | informal-paraphrase | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-direct | held-out | material-summary-omission | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-direct | held-out | material-summary-omission | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-direct | held-out | no-match | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-direct | held-out | no-match | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-direct | held-out | permission-denial | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-direct | held-out | permission-denial | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-direct | held-out | short-direct-work | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-direct | held-out | short-direct-work | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-direct | held-out | stale-index | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-direct | held-out | stale-index | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-direct | held-out | unsupported-required-semantics | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-direct | held-out | unsupported-required-semantics | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-worker | development | ambiguous-alternatives | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-worker | development | ambiguous-alternatives | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-worker | development | complex-worker-work | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-worker | development | complex-worker-work | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-worker | development | conflicting-instructions | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-worker | development | conflicting-instructions | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-worker | development | dependency-loading | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-worker | development | dependency-loading | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-worker | development | explicit-selection | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-worker | development | explicit-selection | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-worker | development | informal-paraphrase | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-worker | development | informal-paraphrase | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-worker | development | material-summary-omission | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-worker | development | material-summary-omission | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-worker | development | no-match | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-worker | development | no-match | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-worker | development | permission-denial | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-worker | development | permission-denial | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-worker | development | short-direct-work | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-worker | development | short-direct-work | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-worker | development | stale-index | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-worker | development | stale-index | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-worker | development | unsupported-required-semantics | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-worker | development | unsupported-required-semantics | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-worker | held-out | ambiguous-alternatives | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-worker | held-out | ambiguous-alternatives | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-worker | held-out | complex-worker-work | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-worker | held-out | complex-worker-work | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-worker | held-out | conflicting-instructions | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-worker | held-out | conflicting-instructions | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-worker | held-out | dependency-loading | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-worker | held-out | dependency-loading | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-worker | held-out | explicit-selection | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-worker | held-out | explicit-selection | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-worker | held-out | informal-paraphrase | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-worker | held-out | informal-paraphrase | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-worker | held-out | material-summary-omission | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-worker | held-out | material-summary-omission | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-worker | held-out | no-match | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-worker | held-out | no-match | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-worker | held-out | permission-denial | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-worker | held-out | permission-denial | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-worker | held-out | short-direct-work | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-worker | held-out | short-direct-work | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-worker | held-out | stale-index | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-worker | held-out | stale-index | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-worker | held-out | unsupported-required-semantics | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | lookup-worker | held-out | unsupported-required-semantics | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | native-catalogue | development | ambiguous-alternatives | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | native-catalogue | development | ambiguous-alternatives | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | native-catalogue | development | complex-worker-work | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | native-catalogue | development | complex-worker-work | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | native-catalogue | development | conflicting-instructions | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | native-catalogue | development | conflicting-instructions | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | native-catalogue | development | dependency-loading | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | native-catalogue | development | dependency-loading | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | native-catalogue | development | explicit-selection | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | native-catalogue | development | explicit-selection | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | native-catalogue | development | informal-paraphrase | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | native-catalogue | development | informal-paraphrase | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | native-catalogue | development | material-summary-omission | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | native-catalogue | development | material-summary-omission | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | native-catalogue | development | no-match | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | native-catalogue | development | no-match | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | native-catalogue | development | permission-denial | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | native-catalogue | development | permission-denial | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | native-catalogue | development | short-direct-work | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | native-catalogue | development | short-direct-work | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | native-catalogue | development | stale-index | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | native-catalogue | development | stale-index | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | native-catalogue | development | unsupported-required-semantics | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | native-catalogue | development | unsupported-required-semantics | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | native-catalogue | held-out | ambiguous-alternatives | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | native-catalogue | held-out | ambiguous-alternatives | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | native-catalogue | held-out | complex-worker-work | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | native-catalogue | held-out | complex-worker-work | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | native-catalogue | held-out | conflicting-instructions | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | native-catalogue | held-out | conflicting-instructions | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | native-catalogue | held-out | dependency-loading | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | native-catalogue | held-out | dependency-loading | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | native-catalogue | held-out | explicit-selection | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | native-catalogue | held-out | explicit-selection | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | native-catalogue | held-out | informal-paraphrase | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | native-catalogue | held-out | informal-paraphrase | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | native-catalogue | held-out | material-summary-omission | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | native-catalogue | held-out | material-summary-omission | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | native-catalogue | held-out | no-match | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | native-catalogue | held-out | no-match | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | native-catalogue | held-out | permission-denial | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | native-catalogue | held-out | permission-denial | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | native-catalogue | held-out | short-direct-work | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | native-catalogue | held-out | short-direct-work | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | native-catalogue | held-out | stale-index | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | native-catalogue | held-out | stale-index | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | native-catalogue | held-out | unsupported-required-semantics | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| claude | native-catalogue | held-out | unsupported-required-semantics | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-direct | development | ambiguous-alternatives | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-direct | development | ambiguous-alternatives | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-direct | development | complex-worker-work | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-direct | development | complex-worker-work | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-direct | development | conflicting-instructions | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-direct | development | conflicting-instructions | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-direct | development | dependency-loading | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-direct | development | dependency-loading | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-direct | development | explicit-selection | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-direct | development | explicit-selection | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-direct | development | informal-paraphrase | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-direct | development | informal-paraphrase | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-direct | development | material-summary-omission | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-direct | development | material-summary-omission | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-direct | development | no-match | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-direct | development | no-match | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-direct | development | permission-denial | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-direct | development | permission-denial | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-direct | development | short-direct-work | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-direct | development | short-direct-work | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-direct | development | stale-index | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-direct | development | stale-index | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-direct | development | unsupported-required-semantics | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-direct | development | unsupported-required-semantics | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-direct | held-out | ambiguous-alternatives | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-direct | held-out | ambiguous-alternatives | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-direct | held-out | complex-worker-work | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-direct | held-out | complex-worker-work | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-direct | held-out | conflicting-instructions | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-direct | held-out | conflicting-instructions | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-direct | held-out | dependency-loading | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-direct | held-out | dependency-loading | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-direct | held-out | explicit-selection | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-direct | held-out | explicit-selection | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-direct | held-out | informal-paraphrase | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-direct | held-out | informal-paraphrase | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-direct | held-out | material-summary-omission | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-direct | held-out | material-summary-omission | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-direct | held-out | no-match | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-direct | held-out | no-match | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-direct | held-out | permission-denial | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-direct | held-out | permission-denial | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-direct | held-out | short-direct-work | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-direct | held-out | short-direct-work | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-direct | held-out | stale-index | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-direct | held-out | stale-index | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-direct | held-out | unsupported-required-semantics | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-direct | held-out | unsupported-required-semantics | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-worker | development | ambiguous-alternatives | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-worker | development | ambiguous-alternatives | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-worker | development | complex-worker-work | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-worker | development | complex-worker-work | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-worker | development | conflicting-instructions | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-worker | development | conflicting-instructions | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-worker | development | dependency-loading | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-worker | development | dependency-loading | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-worker | development | explicit-selection | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-worker | development | explicit-selection | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-worker | development | informal-paraphrase | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-worker | development | informal-paraphrase | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-worker | development | material-summary-omission | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-worker | development | material-summary-omission | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-worker | development | no-match | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-worker | development | no-match | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-worker | development | permission-denial | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-worker | development | permission-denial | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-worker | development | short-direct-work | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-worker | development | short-direct-work | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-worker | development | stale-index | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-worker | development | stale-index | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-worker | development | unsupported-required-semantics | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-worker | development | unsupported-required-semantics | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-worker | held-out | ambiguous-alternatives | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-worker | held-out | ambiguous-alternatives | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-worker | held-out | complex-worker-work | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-worker | held-out | complex-worker-work | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-worker | held-out | conflicting-instructions | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-worker | held-out | conflicting-instructions | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-worker | held-out | dependency-loading | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-worker | held-out | dependency-loading | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-worker | held-out | explicit-selection | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-worker | held-out | explicit-selection | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-worker | held-out | informal-paraphrase | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-worker | held-out | informal-paraphrase | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-worker | held-out | material-summary-omission | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-worker | held-out | material-summary-omission | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-worker | held-out | no-match | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-worker | held-out | no-match | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-worker | held-out | permission-denial | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-worker | held-out | permission-denial | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-worker | held-out | short-direct-work | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-worker | held-out | short-direct-work | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-worker | held-out | stale-index | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-worker | held-out | stale-index | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-worker | held-out | unsupported-required-semantics | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | lookup-worker | held-out | unsupported-required-semantics | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | native-catalogue | development | ambiguous-alternatives | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | native-catalogue | development | ambiguous-alternatives | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | native-catalogue | development | complex-worker-work | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | native-catalogue | development | complex-worker-work | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | native-catalogue | development | conflicting-instructions | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | native-catalogue | development | conflicting-instructions | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | native-catalogue | development | dependency-loading | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | native-catalogue | development | dependency-loading | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | native-catalogue | development | explicit-selection | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | native-catalogue | development | explicit-selection | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | native-catalogue | development | informal-paraphrase | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | native-catalogue | development | informal-paraphrase | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | native-catalogue | development | material-summary-omission | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | native-catalogue | development | material-summary-omission | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | native-catalogue | development | no-match | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | native-catalogue | development | no-match | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | native-catalogue | development | permission-denial | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | native-catalogue | development | permission-denial | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | native-catalogue | development | short-direct-work | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | native-catalogue | development | short-direct-work | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | native-catalogue | development | stale-index | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | native-catalogue | development | stale-index | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | native-catalogue | development | unsupported-required-semantics | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | native-catalogue | development | unsupported-required-semantics | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | native-catalogue | held-out | ambiguous-alternatives | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | native-catalogue | held-out | ambiguous-alternatives | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | native-catalogue | held-out | complex-worker-work | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | native-catalogue | held-out | complex-worker-work | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | native-catalogue | held-out | conflicting-instructions | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | native-catalogue | held-out | conflicting-instructions | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | native-catalogue | held-out | dependency-loading | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | native-catalogue | held-out | dependency-loading | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | native-catalogue | held-out | explicit-selection | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | native-catalogue | held-out | explicit-selection | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | native-catalogue | held-out | informal-paraphrase | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | native-catalogue | held-out | informal-paraphrase | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | native-catalogue | held-out | material-summary-omission | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | native-catalogue | held-out | material-summary-omission | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | native-catalogue | held-out | no-match | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | native-catalogue | held-out | no-match | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | native-catalogue | held-out | permission-denial | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | native-catalogue | held-out | permission-denial | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | native-catalogue | held-out | short-direct-work | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | native-catalogue | held-out | short-direct-work | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | native-catalogue | held-out | stale-index | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | native-catalogue | held-out | stale-index | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | native-catalogue | held-out | unsupported-required-semantics | 1 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |
| codex | native-catalogue | held-out | unsupported-required-semantics | 2 | 1 | 0 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | 0 |

## Paired performance

| Harness | Candidate | Metric | Coverage | Median ratio | Difference range | Missing reason |
| --- | --- | --- | --- | ---: | --- | --- |
| claude | lookup-direct | aggregate input plus output | 0/24 | null | null | no matched completed-work cells with the required metric |
| claude | lookup-direct | wall latency | 0/24 | null | null | no matched completed-work cells with the required metric |
| claude | lookup-direct | peak root per-request input on complex tasks | 0/2 | null | null | no matched completed-work cells with the required metric |
| claude | lookup-worker | aggregate input plus output | 0/24 | null | null | no matched completed-work cells with the required metric |
| claude | lookup-worker | wall latency | 0/24 | null | null | no matched completed-work cells with the required metric |
| claude | lookup-worker | peak root per-request input on complex tasks | 0/2 | null | null | no matched completed-work cells with the required metric |
| codex | lookup-direct | aggregate input plus output | 0/24 | null | null | no matched completed-work cells with the required metric |
| codex | lookup-direct | wall latency | 0/24 | null | null | no matched completed-work cells with the required metric |
| codex | lookup-direct | peak root per-request input on complex tasks | 0/2 | null | null | no matched completed-work cells with the required metric |
| codex | lookup-worker | aggregate input plus output | 0/24 | null | null | no matched completed-work cells with the required metric |
| codex | lookup-worker | wall latency | 0/24 | null | null | no matched completed-work cells with the required metric |
| codex | lookup-worker | peak root per-request input on complex tasks | 0/2 | null | null | no matched completed-work cells with the required metric |

Token, wall-latency, peak-root-input, and actual-window columns are completeness counts over the declared denominator. Their per-trial values and missing reasons are in `report.json`; no missing value is represented as zero. Unavailable, failed, invalid, and missing cells are never treated as cheaper completed work. Null score values mean the cell was not eligible for scoring. Peak root request input is not described as actual context occupancy unless the harness reports that stronger metric separately.
