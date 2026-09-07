# Repository Instructions

## Current workflow: portable assistant rebuild

On 2026-09-06 the product owner explicitly authorized replacing the legacy
delivery gates for this rebuild with direct local implementation, testing, and
review. The direction in `docs/discovery/portable-assistant-vnext/README.md` is
current. Conflicting old product plans and lifecycle documents are historical
context, not prerequisites for this work.

On 2026-09-07 the product owner clarified that this repository is being developed
directly, not through a personal assistant instance. Private-agent account and
SDLC policies do not govern this development workflow.

- Implement approved work without requiring discovery issues, planning PRs,
  exact-head product approvals, or acceptance cohorts first.
- Use tests first for meaningful behavior and exercise failure boundaries.
- Review the diff and document actual behavior, verification, and limitations.
- Preserve unrelated work and existing installations. Use disposable test
  repositories and synthetic accounts, never ambient harness configuration.
- Keep private memory, credentials, identities, and transcripts out of this
  public repository. Public contributor attribution is normal project metadata.
- Use the configured, authorized development Git/GitHub account for commits and
  feature-branch pushes. Do not require separate agent identities, alternating
  contributor/reviewer accounts, or private-agent delivery gates. Review the
  diff, validate changes, and checkpoint/push completed work without an extra
  identity-approval ceremony. Do not force-push or merge/release by implication.
- This is the shared open-source platform, not any particular user's agent.
  Core capabilities cover memory management, capability design/validation,
  setup, harness adaptation, and synchronization. Service integrations,
  credential-provider implementations, account-role models, inbox rules, and
  personal workflows belong in each user's private agent repository. Core may
  define provider-neutral extension contracts, but must not ship personal
  integrations as built-ins, optional bundled packages, or seeded examples.
  Use synthetic provider-neutral fixtures for core behavior. The repository's
  own development/release tooling is distinct from installed agent capabilities.
- Development authority does not authorize production release,
  live-data migration, or removal of an existing installation.
- Do not reintroduce mandatory confirmations for the private assistant's
  routine learning, capability execution, or offline use.

Legacy lifecycle and release documentation remains available in
`docs/operations/sdlc.md` as reference, not a prerequisite for this rebuild.

## Engineering Rules

- Prefer existing project patterns over new abstractions.
- Preserve framework-native behavior; do not force a shared visual style across
  unrelated products.
- Use YAGNI: build only what the current brief needs.
- Use SOLID as review pressure, not ceremony.
- Keep project knowledge in repo docs, not chat.
- Keep related behavior, data, contracts, authorization, invariants, failure,
  and operation coherent by capability. Do not copy one permanent document per
  solution-design stage.
- Record significant design decisions and remaining limitations in repository
  docs. Use issues and PRs when useful; do not invent mandatory planning gates.
- Do not commit secrets.
- Treat local and development-preview URLs as ephemeral review aids, not
  acceptance or release evidence.
- Build a release candidate once and promote the same immutable artifact through
  staging and production when this repository has deployed environments.
- Do not invent staging for an artifact repository; nominate its exact verified
  commit and artifact before acceptance and release.
- Use container-first validation unless this repo documents a different path.
- When host-local language execution is supported, commit exact runtime versions
  in `mise.toml`; keep ecosystem version files synchronized when other tooling
  still consumes them.

## Validation

Run:

```sh
bin/container bin/ci
```

If container support is not ready yet, run:

```sh
bin/ci
```

Document any required deviation in `docs/development.md`.
