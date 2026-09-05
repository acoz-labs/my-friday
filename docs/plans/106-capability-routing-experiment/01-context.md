# Context

## Outcome and scope

Issue #106 materializes O1 of approved discovery #105. Maintainers need evidence
to decide whether on-demand discovery and selective delegation merit a later
production design. This is broad experimental tooling with privacy and process
risks, not a new public command or production adapter framework. There is no UI,
database migration, external integration, or release in this outcome.

## Evidence ledger

All repository observations below use the full README basis commit.

- `docs/discovery/104-on-demand-capabilities/README.md` and `evidence.md` own
  the selected outcome, three modes, negative cases, and exploratory lexical
  results. PR #105 was merged with an independent approval of exact head
  `dfdedeedb9e2cc1488185175620a34759cb5fe81`; #106 carries its O1 authority.
- `internal/capability/capability.go` owns strict source-first instruction-only
  manifests and projection digests. Its current contract is not a general
  dependency/execution system. The experiment must not extend that contract.
- `tools/acceptance-runner/` and `tools/acceptance-support/` demonstrate standalone
  Go tools and process/evidence tests. `go.mod`, `mise.toml`, and `bin/ci` own the
  Go toolchain, formatting, vet, race tests, and native validation conventions.
- `README.md` and `docs/development.md` distinguish installed named instances,
  private runtime state, and credential-free fixtures. No experimental run may
  reuse an installed instance as its disposable workspace.
- Discovery's sanitized second-harness observation reports installed CLIs but
  actual Claude Code inference denial. An authentication status is not execution
  proof. Access must be retested only through an already authorized route.
- Maintainer-supplied installed-help inspection: Codex has noninteractive JSON
  output and configuration/rules switches; Claude safe mode disables skills
  and agents. Neither flag inventory proves clean native skill visibility.
  [Codex noninteractive documentation](https://learn.chatgpt.com/docs/non-interactive-mode)
  documents turn usage events, not complete worker aggregation or per-call
  context occupancy. Live adapter preflight owns those proofs.

## Assumptions ledger

- Synthetic tasks can expose routing and policy losses without private skills.
- Existing Go tooling is sufficient; no embedding service or new dependency is
  necessary for this bounded experiment.
- Model behavior varies; paired repetitions support descriptive comparison,
  not statistical proof of universal portability.

## Unknowns ledger

- Runnable Claude access, native worker fidelity, effective ambient-skill
  exclusion, and complete per-call telemetry are runtime prerequisites. Their
  defined outcome is `unavailable` or `invalid`, not a blocking design unknown.
- Performance and correctness differences are the experiment's output.
- Repo-linked Project synchronization cannot currently be verified because the
  configured credential lacks Project read permission. Issue lifecycle remains
  the navigation surface; no permission expansion is part of this work.

## Decisions ledger

D1: bounded developer tool, not production runtime. D2: native fidelity before
measurement. D3: held-out isolation enforced outside the model process. D4:
predeclared rubric and budgets; incomplete evidence preserved. D5: no default
rollout or private-corpus publication. Detailed contracts follow.

## Actors and acceptance

Contributors implement and preregister; an independent maintainer reviews corpus
labels, manifests and scored evidence; an authorized operator opts into live
batches. Acceptance is an inspectable, reproducible three-mode comparison with
honest gaps, not a requirement that the preferred mode win. #101/#103 and #92
retain independent sequencing and scope.
