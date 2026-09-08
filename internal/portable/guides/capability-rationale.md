# Capability rationale

Keep this completed document as RATIONALE.md in the private capability directory.
It explains design decisions; it does not activate references or replace checks.
Do not retain machine-local paths, secrets, raw transcripts, or unnecessary copies
of source material. Use library IDs, relative paths and content hashes instead.

## Current ask

State the requested outcome, scope, allowed effects, current constraints and
acceptance tests before consulting historical implementations.

## Reference evidence

For each relevant source: library ID, library descriptor SHA-256, relative file
path, file SHA-256, the observed experience, and what it means for this task.
Distinguish verified behavior, old reports, explicit user direction and inference.
Hashes identify inspected bytes; they do not prove truth or preserve the original
file. Record inaccessible libraries and retrieval limits; never invent evidence.

## Reused or adapted

Explain what informed the implementation, why it fits the current ask, which
assumptions changed, and which checks verify it. Reuse sound existing code when
appropriate; comply with its license. Rebuilding does not require rewriting.

## Rejected or superseded

Name relevant obsolete assumptions or policies not adopted, with reasons.
Do not silently treat old constraints as current user requirements.

## Unresolved

Record uncertainty and material decisions needing user input. Ordinary adaptation
and authorized implementation do not require a new approval at every step.

## Verification and active learning

List the actual tests, outcomes, limitations and regression cases derived from
past failures. Only verified new behavior and applicable knowledge belong in
active memory. Use explicit memory supersession when changing a current process.
