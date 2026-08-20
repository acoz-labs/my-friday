# Deployment

Document the repository's actual delivery profile and release path. Do not
invent staging or production for a repository that does not have them.

## Delivery Profile

`service`, `artifact`, or `non-deployable`.

For a non-deployable repository, record why no release target exists and which
verification completes delivery, then remove sections that do not apply.

## Immutable Candidate

Describe what is built, how it is identified, and how the same candidate moves
through staging, acceptance, and production without rebuilding mutable source.
Artifact profiles use `Nominate artifact candidate` after successful main CI;
this records exact-candidate evidence without pretending that staging exists.

## Environments

| Environment | Purpose | Candidate source | Verification | Promotion authority |
|---|---|---|---|---|
| Development preview | | | | |
| Staging | | | | |
| Production | | | | |

## Configuration And Secrets

Name configuration keys, ownership, and injection mechanisms without recording
secret values. Explain which configuration changes require separate review or
rollback preparation.

Set `PRODUCTION_REQUIRED_SECRETS` to the comma-separated environment-variable
names that the release workflow injects, or the explicit value `none`. Never
record values. Declare `PRODUCTION_ACTIVATION_MODE=deploy-command` when deploy
activates the candidate, or `separate-command` with an executable
`bin/activate-production`.

## Production Readiness Preflight

`bin/preflight-production` runs before a production receipt is created. A
production-enabled repository must provide executable `bin/deploy-production`,
`bin/verify-production-candidate`, `bin/verify-production`, and
`bin/preflight-production-project` hooks. The candidate hook proves an exact
artifact was promoted even after an interrupted workflow, without activating or
redeploying it: exit `0` means active, `3` means conclusively inactive, and every
other status is an inconclusive error that stops the retry. The project
preflight verifies repository-specific
infrastructure, configuration, migrations, capacity, and rollback prerequisites
without exposing secrets.

## Release Procedure

1. State the exact release entrypoint.
2. Identify the candidate and required checks.
3. Verify staging or artifact evidence and independent acceptance.
4. Run the production-readiness preflight.
5. Promote the accepted candidate and record an exact-candidate receipt.
6. Verify and activate the release.
7. Finalize the receipt, tag, release ledger, lifecycle, Project fields, and
   linked work without redeploying on an evidence-only retry.

Keep commands current and executable.

## Verification

List automated checks, smoke scenarios, externally visible evidence, and the
conditions that declare the release successful.

## Rollback

Identify the rollback target, command or procedure, data/configuration
constraints, verification steps, and decision authority.

## Failure Recovery

Link `docs/runbook.md` or a focused runbook for recurring deployment and release
failures.
