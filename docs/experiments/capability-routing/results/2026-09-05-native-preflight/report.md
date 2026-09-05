# Capability routing comparison

- Corpus revision: `routing-corpus-v1`
- Manifest SHA-256: `4f0256b20c8e03ca3913286ceadec85f89b150ca97a791ecbd08f5526f160890`
- Recommendation: **inconclusive**
- Conclusion: At least one held-out harness/mode cell is incomplete or lacks required telemetry; retain the native baseline and make no cross-harness claim.

## Native-driver fidelity

| Harness | Version | State | Missing controls |
| --- | --- | --- | --- |
| claude | 2.1.193 (Claude Code) | unavailable | no OS-enforced fixture-only read boundary was demonstrated without exposing host authentication; native skill body reads were not proven constrained to the staged allowlist; native worker identity/inheritance and pre-launch depth/count enforcement were not demonstrated; enabled built-in tools lack an equivalent pre-dispatch eight-call rejection hook |
| codex | codex-cli 0.153.4 | unavailable | no OS-enforced fixture-only read boundary was demonstrated without exposing host authentication; native skill body reads were not proven constrained to the staged allowlist; native worker identity/inheritance and pre-launch depth/count enforcement were not demonstrated; enabled built-in tools lack an equivalent pre-dispatch eight-call rejection hook |

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

## Paired performance

| Harness | Candidate | Metric | Coverage | Median ratio | Difference range | Missing reason |
| --- | --- | --- | --- | ---: | --- | --- |
| claude | lookup-direct | aggregate input plus output | 0/24 | null | null | no matched correct completed-work cells with the required metric |
| claude | lookup-direct | wall latency | 0/24 | null | null | no matched correct completed-work cells with the required metric |
| claude | lookup-direct | peak root per-request input on complex tasks | 0/2 | null | null | no matched correct completed-work cells with the required metric |
| claude | lookup-worker | aggregate input plus output | 0/24 | null | null | no matched correct completed-work cells with the required metric |
| claude | lookup-worker | wall latency | 0/24 | null | null | no matched correct completed-work cells with the required metric |
| claude | lookup-worker | peak root per-request input on complex tasks | 0/2 | null | null | no matched correct completed-work cells with the required metric |
| codex | lookup-direct | aggregate input plus output | 0/24 | null | null | no matched correct completed-work cells with the required metric |
| codex | lookup-direct | wall latency | 0/24 | null | null | no matched correct completed-work cells with the required metric |
| codex | lookup-direct | peak root per-request input on complex tasks | 0/2 | null | null | no matched correct completed-work cells with the required metric |
| codex | lookup-worker | aggregate input plus output | 0/24 | null | null | no matched correct completed-work cells with the required metric |
| codex | lookup-worker | wall latency | 0/24 | null | null | no matched correct completed-work cells with the required metric |
| codex | lookup-worker | peak root per-request input on complex tasks | 0/2 | null | null | no matched correct completed-work cells with the required metric |

Token, wall-latency, peak-root-input, and actual-window columns are completeness counts over the declared denominator. Their per-trial values and missing reasons are in `report.json`; no missing value is represented as zero. Unavailable, failed, invalid, and missing cells are never treated as cheaper completed work. Null score values mean the cell was not eligible for scoring. Peak root request input is not described as actual context occupancy unless the harness reports that stronger metric separately.
