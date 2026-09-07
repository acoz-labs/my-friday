# Contributing

Follow the current development workflow in [AGENTS.md](AGENTS.md). My Friday is
an open-source toolkit, not an installation of a contributor's personal agent.
Private-agent account separation, software-delivery policies, and service
capabilities do not govern contributions to this repository.

## Development workflow

1. Agree on the intended change and keep its scope bounded. Record significant
   decisions in repository docs; use issues when useful for coordination.
2. Implement on a feature branch with meaningful tests. Preserve unrelated work
   and use disposable repositories and synthetic fixtures.
3. Run the checks in [development.md](docs/development.md), review the diff, and
   update documentation to describe actual behavior and limitations.
4. Commit and push using the configured, authorized development account. No
   separate agent account, planning PR, or exact-head product approval is required
   to checkpoint development work.
5. Review before merging. A feature-branch push is not a production release or
   permission to migrate a live agent.

Pull requests should include a summary, tests/checks run, documentation changes,
known limitations, and any relevant compatibility or recovery notes. Link an
issue when one exists; an issue or separate planning PR is not a prerequisite.

Keep the public core provider-neutral. Memory management, capability
design/validation, setup, harness adapters, and synchronization belong here.
Personal workflows and service integrations belong in users' private agent
repositories. Never include private memories, credentials, or transcripts.

The [legacy SDLC and release reference](docs/operations/sdlc.md) preserves earlier
process documentation. It does not reintroduce development gates superseded by
the current workflow. Production releases and live migrations require their own
explicit scope and verification.
