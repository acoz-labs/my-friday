# Architecture

Describe the system that exists now. Design history belongs in linked issues
and ADRs; this document and its capability pages describe the shipped contract.

## Purpose And Scope

State what the system does, its primary actors, and the boundary of this
overview.

## System Context

Add the smallest useful Mermaid context or component diagram. Explain any
boundary, dependency, or invariant that is not visible in the diagram.

## Components And Boundaries

| Component | Responsibility | Owns | Depends on |
|---|---|---|---|
| | | | |

## Data And State

Summarize the domain model, storage systems, ownership rules, background work,
and cross-component consistency boundaries.

## External Services

Record durable integration contracts and failure boundaries without placing
credentials, volatile inventory, or environment-specific secret values here.

## Deployment Boundaries

Describe which components are built and deployed together. Link
`docs/deployment.md` for the release and rollback procedure.

## Capability Architecture

Keep an end-to-end capability vertical when it needs more detail than this
overview can carry. Copy
`docs/architecture/0000-capability-template.md` to a descriptive filename and
remove sections that do not apply.

| Capability | Documentation |
|---|---|
| | |

## Decisions And Tradeoffs

Link `docs/decisions/` for consequential decisions likely to be questioned
later. Do not duplicate the complete decision history here.
