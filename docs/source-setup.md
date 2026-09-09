# Guided source hosting and synchronization

The setup wizard can now create/connect a private GitHub source repository,
choose its owner and verify ongoing synchronization access. This is shared
installation infrastructure, not a GitHub coding/review capability or account-role
system. No service integration is added to the assistant capability inventory.
Local-only setup and existing private credential helpers remain supported.

## Start or resume

For a reusable menu, run `my-friday`, then **Manage an installed agent →
Configure repository and synchronization**. The [management menu](management-menu.md)
also exposes setup, health, repair and toolkit updates. The explicit commands below
remain supported for direct workflows.

`my-friday setup` performs the existing local installation wizard, then offers
source synchronization. To configure an already installed assistant:

```sh
my-friday setup --instance /absolute/path/to/instance
```

Resume does not recreate the source, identity, device, launcher, instance binding,
native settings, authentication or sessions. It requires a clean worktree on
`main`, a valid source and an ordinary source-owned `.git` directory. Checkpoint or
reconcile outstanding work first; close active agent sessions during configuration.
Use an appropriate upgraded toolkit: an old executable bound to the launcher will
not understand the new optional `github_source` field. Resume is not an executable
upgrader or rebind command. It refuses to run against an instance bound to another
executable path. Do not replace a pinned binary silently.

Noninteractive creation flags retain their prior behavior and do not unexpectedly
prompt for remote configuration. Resume is interactive; EOF and declined
confirmation do not authorize the next action. A completed local install remains
usable if a later wizard step is interrupted. Re-run with `--instance`, not create.

Choices:

- `local`: leave synchronization settings unchanged. It does not disconnect an
  already configured remote or stop its future synchronization.
- `github`: choose a stored GitHub CLI account, owner/name and create/connect.
- `existing`: supply an HTTPS URL backed by an already configured private helper,
  or an absolute local Git remote path. Visibility/privacy is user-verified for
  this provider-neutral route; it does not provision or replace credentials.

## GitHub source route

Install and authenticate `gh` separately using its native login flow. The wizard
lists stored github.com account names and asks separately for the repository
setup account and the account authorized for **ongoing source sync**. They may
be the same or different; the initial upload is not an implicit ongoing grant.
It does not change the CLI's active account. Repository ownership can be the
setup account or an organization it is permitted to create repositories under.
Ownership is distinct from authentication and from any future private agent roles.
When accounts differ, source access for the ongoing account must already be
granted/accepted before verification succeeds. Setup does not manage collaborators
or invitations; it retains any newly created repo and reports this prerequisite
so you can grant access separately and resume. A different personal owner cannot
be impersonated for creation; use that owner's setup account or connect to a
repository it has already created.

Choose `create` to create a private, uninitialized repository if absent, or
`connect` to use an existing one without creating it. The wizard prints the exact
owner/name and selected account and requires the owner/name as confirmation.
GitHub repositories must be private, writable, active and match the requested
target. Public/archived/inaccessible repositories are refused, not modified.
Creation is not deletion-on-failure: a created repository is retained if a later
step fails. A repeat first inspects an existing repository instead of repeating
creation. An uncertain network result is not reported as success.

The wizard uploads the assistant's **complete committed history**, not a filtered
snapshot. Review history for secrets first. The wizard is not a secret scanner
or history scrubber and never changes visibility or force-pushes.

An existing `origin` is never replaced. Multiple origin URLs, separate push URLs,
mirror settings, nonstandard fetch mappings and unsupported transports require
manual reconciliation. Before attaching a new
remote, Git checks whether it is empty or has a `main` branch whose `agent.json`
identifies this same assistant. A nonempty foreign repository is refused; import
that assistant separately rather than attaching it to a new identity. Full source
and append-only validation still occur during normal synchronization. A failed
inspection may leave fetched Git objects/FETCH_HEAD, not a changed worktree.
Remote URL cloning/import and SSH transport are not implemented by this wizard.

## Credential storage and portable configuration

The optional source configuration is version 1 with this additional field:

```json
{
  "schema_version": 1,
  "github_source": {"repository": "example-owner/example-assistant"}
}
```

It lives in `.my-friday/sync.json`. Existing Git `author` settings are preserved.
`github_source` and `credential_helper` are mutually exclusive; the wizard refuses
to replace a custom helper with the GitHub adapter. Existing sources without this
field remain compatible. Older toolkit builds reject the new field: upgrade
every writer before enabling it.

The selected account is recorded in Git-ignored
`.my-friday/local/source-github.json`, tied to assistant ID and owner/name. It
contains **no token**. Another checkout must bind an account through setup on that
machine; source descriptors alone do not confer access. Local metadata must be
ignored and untracked, with no symlink binding. Multiple local instances sharing
one source checkout share this source account binding.

The adapter retrieves the explicitly selected account's token from `gh`, verifies
its login through the API, and verifies the target repository's private/write
state before returning credentials to Git. It accepts only HTTPS github.com
credential requests for that exact owner/name. `store` and `erase` do nothing;
native GitHub CLI storage remains authoritative. Ambient GH/GITHUB token, host
and debug overrides are removed for these calls; `GH_CONFIG_DIR`/XDG configuration
locations are preserved. There is no silent fallback to another active account.
Native credential-store restrictions over SSH can still require a local login.

The internal `source-credential` endpoint writes credentials only as Git protocol
output. **Do not invoke its `get` operation manually, log its stdout, or put it in
an agent prompt.** Raw provider diagnostics are suppressed. Each provider process
has a 10-second deadline and a 1 MiB output limit. Tokens exist in process memory
and the child API process environment; full-access same-user processes are not
isolated by this design. The selected token's server-side permissions are not
reduced by the client-side source-path check; provision narrow credentials where
appropriate. This adapter is not a general-purpose secret broker.

## Verification and recovery

Configuration/local binding are saved and checkpointed before remote attachment.
If later verification fails, these settings and any created remote are retained
for inspection/resumption; no success is claimed. Successful completion requires
normal fetch/reconciliation/push to return `synced`. Offline/auth failures remain
pending with local history preserved; conflicts need explicit reconciliation.
Hook budgets may be shorter than full setup time. Never interpret a configured
helper, a successful API login, or structural doctor as proof of successful sync.

Tests use synthetic accounts, a fake provider and local bare Git repositories.
They cover explicit owner/account selection, private-only creation, wrong-account
and public-repository refusal, exact credential scope, no ambient-token fallback,
machine-local bindings, cancellation/redaction, decline/EOF, clean-source checks,
foreign/no-main remotes, origin preservation and compiled CLI sync/resume. Real
native credential access and the owner's interactive experience still require
hands-on verification before considering that installation's remote sync ready.

Primary API/CLI references: [GitHub account status](https://cli.github.com/manual/gh_auth_status),
[explicit account token selection](https://cli.github.com/manual/gh_auth_token),
and [GitHub CLI API requests](https://cli.github.com/manual/gh_api).
Repository creation/privacy fields follow the
[GitHub repository REST API](https://docs.github.com/en/rest/repos/repos).
