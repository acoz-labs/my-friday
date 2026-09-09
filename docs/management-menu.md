# Agent management menu

Run `my-friday` (or `my-friday menu`) at any time, from any working directory.
Opening the menu does not launch an agent, synchronize source or change files.
It uses numbered choices instead of cursor-control dependencies, so it works
over SSH and in plain terminals. Enter `0` for Back/Exit or `:back` at a prompt.
EOF exits without approving an unfinished action. Errors return to the menu;
successful earlier actions are retained, not implicitly undone by Back.

The home menu offers new setup, import, installed-agent management and updates.
Discovery reads only `~/.local/share/my-friday/instances`, showing damaged entries
instead of hiding them. An explicit path can open a nonstandard installation;
these paths are not automatically registered or found by scanning the home tree.
New setup uses the standard source/instance/launcher locations and asks for the
name, harness and machine label, then previews paths before writing anything.
Existing destinations are not overwritten. Explicit `setup` flags still support
custom installation paths. Import currently requires an already-cloned,
new-format My Friday source. Remote cloning, legacy memory migration and native
harness installation/login are not part of this menu increment.

## Managing an agent

- **Status:** identity, paths, pinned toolkit, default harness and configured
  source remote. A remote URL is not evidence that recent synchronization worked.
- **Repository and synchronization:** the existing [source wizard](source-setup.md),
  with human-readable results. It uses the running toolkit only when the selected
  agent is bound to that executable; otherwise adopt it or open the pinned version.
- **Default harness:** validate `codex`/`pi`, refuse unrelated uncommitted source
  work, and checkpoint the setting with the management device's provenance.
  This is a portable default, not just a local override. It propagates on a later
  sync and does not change an active session or perform native login.
- **Health:** read-only structural doctor, with readable findings and remedies.
  This is not an authentication, network or functional capability check.
- **Repair:** confirmed regeneration of managed instructions/hooks, preserving
  source, native config, credentials and sessions. A missing default launcher can
  be recreated with separate confirmation. Existing custom launchers are untouched.
- **Use this toolkit version:** adopt the running executable for this agent,
  with private rollback copies and explicit managed-launcher selection.
- **Rollback:** choose a local adoption checkpoint, verify that its previous
  executable supports today's source, and restore unchanged managed files only.

Close active agent sessions before repair, adoption or rollback and start fresh
afterward. This menu does not claim to detect or terminate every active harness.
Doctor compares with the running toolkit, as documented by the underlying CLI.
Repair does not install a missing executable, repair Git history, fix credentials
or restore deleted memory. Invalid source/bindings may require manual recovery;
the menu preserves them and reports the problem rather than recreating the agent.

## Toolkit installation and agent pins

Updating **My Friday** and synchronizing **agent source** are different actions.
The update menu offers:

1. Check the official latest published release.
2. Install a separately approved local executable with its trusted SHA-256.
3. Switch the management command to a retained compatible executable.

No network check runs until requested. The public release route sends no account
credentials, reads the official GitHub release metadata and requires the explicit
portable-management manifest described below. Legacy releases without it are
refused, even if GitHub calls them latest. Drafts/prereleases, ambiguous/missing
assets, wrong platforms, unsupported schemas and mismatched checksums are refused.
An unavailable compatible release is not reported as “you are up to date.” The
latest-release endpoint is not a branch updater or a search across prereleases.
The user sees the selected release and digest before approving installation.

Downloads are HTTPS-only from the official release path, with constrained GitHub
asset redirects, size limits and timeouts. The SHA-256 is checked before activation.
It establishes integrity relative to the official HTTPS manifest, **not a separate
signature/trust authority**. Publisher compromise is outside this guarantee.
Local artifacts require explicit trust and the expected digest; the updater never
computes a digest and treats that alone as proof that an unknown file is safe.
Approved executables run a bounded compatibility handshake before activation.

Artifacts are installed without overwrite at
`~/.local/share/my-friday/releases/sha256-DIGEST/my-friday`. Only the managed
`~/.local/bin/my-friday` convenience symlink is activated; a custom regular file
or an unexpectedly changed pointer is preserved. A private activation directory
retains the previous symlink. Failed checks may leave a staged artifact, never a
claimed successful activation. No old executable is deleted automatically.
After activation the menu exits: run `my-friday` again for the new version.

Each agent remains pinned until **Manage → Use this toolkit version**. This
permits deliberate per-agent updates, including separate personal/work instances.
Adoption validates the source and sync configuration, requires an exact existing
generated launcher (or explicitly no launcher), then backs up and rewrites only
the binding, managed instructions/hooks and that launcher. The private instance's
`toolkit-updates/TIMESTAMP-…/receipt.json` records before-bytes and after-digests.
No source, credentials, native settings or transcripts are copied into the receipt.
Source writes and adoption use the assistant source lock; this is not isolation
from hostile same-user processes or arbitrary native writers.

Rollback refuses later managed-file edits rather than discarding them, and never
rolls back memory or external actions. Older binaries that cannot run the
compatibility endpoint require deliberate manual recovery with retained files.
A process crash can leave mixed managed files; the receipt remains for inspection.
An ordinary write failure attempts restoration and reports restoration failures.
Automatic repair does not delete/recreate damaged installations.

## Release contract (publishing remains separate)

Publish `my-friday-update.json` alongside **raw executables**, not archives:

```json
{
  "schema_version": 1,
  "portable_format": 1,
  "management_protocol": 1,
  "version": "v1.0.0",
  "artifacts": [
    {
      "os": "darwin",
      "arch": "arm64",
      "name": "my-friday-darwin-arm64",
      "sha256": "<64 lowercase hex characters>"
    }
  ]
}
```

The manifest version must equal the release tag, with exactly one matching
platform entry and matching assets on that release. Existing legacy tarball
assets are not eligible. The explicit development helper:

```sh
my-friday toolkit manifest --binary /approved/my-friday --release v1.0.0
```

checks the local executable and prints single-platform metadata; it does not
publish, rebuild, select a release channel, or modify release automation. Run on
each target platform and combine approved artifact entries for a multi-platform
release. Production publication still requires separate authority and acceptance.
The stable route cannot install this feature until a compatible artifact and
manifest are actually published; the approved local-artifact route permits pilots.

`my-friday version`/`--version` print build revision, platform and compatibility
metadata as JSON. `toolkit check-instance --instance PATH` is a read-only source
compatibility endpoint, not doctor or a harness/network check.
`toolkit use --instance PATH --launcher PATH` is the explicit automation form of
adoption and produces JSON. These commands keep scripting separate from the menu.

Release lookup follows the official [GitHub releases API](https://docs.github.com/en/rest/releases/releases#get-the-latest-release).
