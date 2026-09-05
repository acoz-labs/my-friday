# Sanitized exploratory evidence, 2026-09-05

These observations motivate a controlled experiment. They do not establish a
production architecture, performance advantage, or cross-harness compatibility.

## Catalogue inventory

The observed custom catalogue contained 59 skills and 13 agent definitions.
Description-only text contained 19,757 characters; skill names plus descriptions
contained 20,219 characters / 20,229 UTF-8 bytes. Plugin skills, paths, wrapper
framing, tools, policy and references were excluded. Full instruction sources
contained 269,594 characters / 269,934 bytes, but progressive instruction loading
means this is **not** startup overhead or a claimed token saving.

## Lexical probe method

Read one name and description per installed skill, following installed links.
Reject unsupported multiline frontmatter. Search only those two metadata fields;
instruction bodies contribute lengths and file digests, never retrieval terms.
Tokenize lowercase ASCII letters/digits, split punctuation, and remove this fixed
stop list: `a an and are as at be by for from in into is it of on or that the this
to use when with without you your`. Use unmodified BM25 with `k1=1.2`, `b=0.75`,
and IDF `log(1 + (N - df + 0.5) / (df + 0.5))`. Sum over unique query terms,
retain positive-score candidates, sort descending by score and then by name,
and return at most three results. No aliases, semantic reranker, or model calls.

The manually labelled convenience suite contained ten explicit domain queries,
two ambiguous requests, three requests outside the catalogue, and eight later
paraphrases. Retrieval was unchanged between the initial and paraphrase groups.
Examples contrasted explicit existing-credential injection with “the deployment
needs the saved token passed safely to its process”; ambiguous memory updates
admitted several capture/governance capabilities; out-of-catalogue requests
included translation, science explanation, and exercise planning.

| Group | Expected top result | Expected result in top three |
| --- | --- | --- |
| Explicit, n=10 | 10/10 | 10/10 |
| Paraphrase, n=8 | 0/8 | 3/8 |
| Ambiguous, n=2 | 0/2 | 2/2 |

Only 1/3 out-of-catalogue queries correctly returned no result. Generic word
overlap produced false candidates. Two identical runs agreed; their reported
output digest was `df0390f837a91cacea7e01f82eb8c9f1d9232ed4d0062d30afb85f46d7e679bd`.

This is a sanitized observation record, not a reproducible public benchmark:
the underlying private catalogue is deliberately excluded. The explicit queries
share vocabulary with descriptions, the suite is small and not blind, and some
paraphrases need conversation context. O1 must replace this exploratory evidence
with a committed synthetic corpus, complete query labels, and reproducible
runner. Do not use these counts to select a numerical production threshold.

## Second harness availability

Codex CLI 0.153.4 and Claude Code 2.1.193 were installed. Claude authentication
status was positive but actual no-tool inference returned an account-access
denial. Narrow existing credential-route probes did not establish inference.
No second-harness task, token, latency, or compatibility result was produced.
