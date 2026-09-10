# Machine preparation

My Friday owns a harness-neutral check/prepare/verify runner, menu and local
readiness receipts. Private capabilities own prerequisites, installation scripts,
credential storage, enrollment and authentication checks. There is no bundled
provider integration, credential vault, package manager or global installer.

Read `my-friday agent capability-guide` for the embedded, authoritative manifest
and script contract. Existing manifests without `machine_requirements` are
unchanged. Upgrade every participating machine before adding the optional field;
older strict parsers reject it. Removing the field to roll back does not uninstall
software or delete local state. No public release is implied by this feature.

## Human and agent entry points

In the management menu, select an agent → **Prepare this machine**. Inspect a
requirement, then explicitly choose check or prepare. Setup returns to this same
menu. Doctor displays historical readiness alongside structural health; repair
refreshes managed projections only and points to the separate preparation menu.
Neither import, normal launch, source synchronization nor toolkit update invokes
these scripts automatically. No requirements means no provider setup is offered.

Agents use JSON CLI directly, not the TUI:

```sh
my-friday machine status --instance /absolute/instance
my-friday machine check --instance /absolute/instance --capability example-tool --requirement local-runtime
my-friday machine prepare --instance /absolute/instance --capability example-tool --requirement local-runtime
my-friday machine prepare --instance /absolute/instance --capability example-tool --requirement local-runtime --apply --expect-sha256 REVIEWED_HASH
```

The last command requires the fingerprint from the preceding preview; it is not
a permission grant beyond the user's task. Status and preview run no code and
write nothing. `check` explicitly executes private probes and writes local
receipts, but its scripts must avoid external mutations. All command results are
JSON. A nonzero check/prepare exit means not ready or failed; inspect `state` and
`phases`, not exit status alone. Status exit zero only means inspection completed.
The management API remains unchanged; this is a separate agent-ready JSON CLI.

The ordered operation is check → prepare only on exit 10 → verify. Check exit 0
skips preparation, any other check exit fails closed without invoking prepare.
Prepare/verify require exit 0. Selected commands share one source snapshot,
not an OS sandbox. No automatic chain across requirements is supported. A
per-instance nonblocking lock prevents concurrent runner operations. Private
scripts sharing resources across instances must coordinate themselves.

## Local state and evidence

`<instance>/machine/state/<capability>/<requirement>` is a capability-owned local
directory, created with mode 0700. Credentials are not automatically written or
read there by the toolkit. `machine/receipts/<capability>/<requirement>.json`
holds the latest observation; `machine/runs/<run-id>.json` retains each run's
phase record. Files written by the runner are owner-only. Existing unsafe parent
types are refused, not overwritten. These are outside agent Git and preserved
by toolkit adoption/repair/rollback. Copying an instance is not machine import;
bind fresh state on a new machine.

Receipts contain identifiers, device, timestamps, phase status and source hash,
not raw output, secret values or a claim of rollback. `ready` is historical;
expired credentials, changed PATH and external service state need an explicit
live check. Capability changes mark old receipts `stale`. A hard-killed process
can leave `running`; inspect before explicitly retrying. The runner checks again
before preparation but cannot guarantee arbitrary private code is idempotent.

Cancellation/timeout kills ordinary process groups, not deliberately detached
processes. Effects can survive any failure. Source fingerprints cover capability
files and owner permission bits, not external programs, environment or external
resources. Same-user malicious races and deliberately unsafe private code are
not contained. Do not treat receipts as cryptographic authorization or attestations.

## Credential enrollment boundary

V1 preparation is deliberately noninteractive. A capability needing a bootstrap
secret supplies a separate local helper with hidden input and durable local
storage. The user should need to supply a valid secret only once per machine;
ordinary use resolves it only into the processes that need it. Expiry/revocation
requires an explicit replacement, not installer retries or silent overwrites.
This private helper can be run before the preparation command and remains usable
without My Friday. The current TUI does **not** collect secrets or launch that
interactive enrollment helper. Future interactive routing must prove terminal
and transcript behavior before becoming part of the shared contract.

Stdout/stderr from registered phases are discarded. That is not a secret scanner:
scripts can still write unsafe files, open a terminal, invoke external services
or leak through their own logs. They inherit normal host environment; no credential
provider is automatically selected. Hidden input is not protection from terminal
recording. Use an unrecorded local terminal, never chat or shell arguments, and
test secret handling in the private capability before live use.

## Validation scope

Automated synthetic fixtures cover preview/no writes, explicit review, repeat
without reinstall, source drift, check failures without preparation, failed
verification, cancellation, manifest/path validation and plain-menu cancellation.
These are not live package-installation, authentication or new-machine acceptance.
Those belong to the capability author and should be recorded separately.
