# Decision and adversarial comparison

| Hypothesis | Strongest case | Adversarial check | Disposition |
| --- | --- | --- | --- |
| Native progressive catalogue is enough | Simple, native routing already loads bodies on demand | Measure actual visible catalogue and failures rather than assuming every body is startup context | Baseline A |
| Lookup/direct saves context cheaply | Small initial surface; selected instructions remain local | Paraphrases, no-match, and dependencies may defeat lexical rank | Candidate B |
| Selective native workers reduce root growth | Complex details remain in worker contexts | Handoff omissions and total-token overhead can erase benefits | Candidate C |
| External subprocess posing as native worker | Easy to implement uniformly | Does not test the claimed native delegation semantics | Reject as benchmark substitute |
| Full adapter framework or permanent routing agent | Broad future extensibility | Expands runtime scope before evidence and recreates accumulating context | Reject |

Select a standalone Go runner with small explicit Codex and Claude experiment
drivers, not a generalized plugin API. The same generic synthetic instruction
source renders into native fixtures and an experiment-only revision-bound index.
Fixtures representing dependencies or required isolation are experimental
metadata, not changes to the existing instruction-only package schema.

Use deterministic lexical top-three retrieval as the first candidate generator,
with model scope confirmation and one bounded broader-metadata fallback. Freeze
tokenization, ranking, tie-breaking and fallback in preregistration. Avoid an
embedding service and its model/cost variables. This compares the complete
declared routing policies, not lexical ranking in isolation.

Attacks already supported by evidence: the discovery spike's paraphrase failures
invalidate blind first-hit dispatch; native progressive loading invalidates full
body-size startup claims; Claude inference denial invalidates login-only proof.
Unverified claims about actual context savings or native worker availability
remain hypotheses until the live preflight and scored runs establish them.

Reject automatic architecture promotion. Negative results, abstentions, invalid
telemetry and unavailable harnesses are useful outcomes. A later B2 plan owns
any shipped package/compiler or routing-default change.
