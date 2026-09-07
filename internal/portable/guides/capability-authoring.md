# Designing a private capability

This guide is embedded in My Friday. No source checkout, binary inspection,
network access, or harness-specific skill format is needed to author a capability.
Run `agent capability-guide` to read it again and `--help` for command discovery.

## Find and use existing capabilities first

Run `agent capabilities` to list IDs and descriptions. Read a selected capability's
complete instructions.md (or its existing README.md) and capability.json before
using it. Do not rebuild an existing capability just because memory recall did
not mention it: capability manifests are the authoritative capability inventory.

## Authoring workflow

1. Define the trigger, inputs, output, prerequisites, allowed effects, and failure
   behavior. Keep service choices and personal rules in this private repository.
2. Use `agent capability-template --capability <id> --description TEXT` to obtain
   a manifest. This only prints JSON; it does not create files or activate code.
3. Create the directory below under MY_FRIDAY_ASSISTANT_ROOT. Write reusable
   instructions and executable scripts. The printed check command is a placeholder:
   implement checks/check.sh or replace it with your actual check command.
4. Implement meaningful checks with synthetic fixtures: success, failures,
   argument/path handling, and any promised no-write or account boundaries.
5. Run `agent check --capability <id>` and `agent validate`. Inspect the actual
   check results. Structural validity alone does not prove a capability works.
6. Exercise the requested task, record a concise outcome with `memory event`,
   and call `sync` to checkpoint any remaining source changes. Reuse it in a
   fresh conversation to verify discovery. Do not add it to the public toolkit.

```text
capabilities/<id>/
  capability.json
  instructions.md
  scripts/                       # Implementation, if needed
  checks/                        # Synthetic executable checks, if needed
```

IDs contain 3-128 lowercase letters, digits, or hyphens, starting with a letter.
The directory name and manifest ID must match. Source must use regular files and
directories, not symlinks or special files. No source file may exceed 16 MiB.

## Manifest format (schema_version 1)

```json
{
  "schema_version": 1,
  "id": "example-capability",
  "description": "Describe when to use this capability and its outcome.",
  "subscriptions": [],
  "checks": [["sh", "checks/check.sh"]]
}
```

schema_version, id, and a nonempty description are required. subscriptions is an
array of event subscriptions; an empty array means manual invocation only.
checks is an optional array of command-argument arrays. Each command must be
nonempty. Commands are literal argv, not shell expressions. Use an explicit
interpreter (such as sh or python3) or mark directly invoked scripts executable.
Do not invent additional manifest fields: unknown fields are rejected. Put
requirements, usage, limitations, and side-effect policy in instructions.md.

Checks execute sequentially from a temporary copy of the capability directory,
with a 120-second timeout per command. Locate implementation relative to the
copied check file, not a machine-specific path or the live original source.
Use disposable fixtures outside the user's workspace. A successful check exits
zero; diagnostics must never contain secrets. Raw check output is suppressed by
My Friday. For debugging, run the declared check directly from its capability
directory after reading its instructions and effects. Empty checks mean only
structural validation, not tested behavior. Check scripts have full user access;
the temporary copy is not a security sandbox.

## Portable paths and working directories

Launched sessions receive MY_FRIDAY_ASSISTANT_ROOT (agent source directory),
MY_FRIDAY_BIN (absolute My Friday executable), MY_FRIDAY_DEVICE_ID, and
MY_FRIDAY_INSTANCE (machine-local binding). Resolve paths at execution time:

```sh
"$MY_FRIDAY_BIN" agent capabilities
"$MY_FRIDAY_ASSISTANT_ROOT/capabilities/example-capability/scripts/run.sh"
"$MY_FRIDAY_BIN" agent check --capability example-capability
```

Do not embed /tmp trial paths, home directories, or checkout locations in reusable
instructions, code, or durable memory. Helpers should resolve sibling resources
relative to their own file. The assistant's cwd is the user's working project,
not its memory repository. For a manual workspace-inspection capability, keep
that cwd when invoking its script; do not cd into the capability first.
Outside a launched session, use the installed executable and explicit
--repository / --device options. Missing required bindings should produce clear
errors, not guesses at a user's home directory or source checkout.

## Optional lifecycle subscriptions

Add subscriptions only when the user intends automatic execution. Each object
has id, event, command (argv array), optional after (dependency IDs), optional
timeout_seconds (0 = 10 seconds; maximum 120), and failure (warn or stop).
Dependency IDs are local subscription IDs or capability-id/subscription-id and
must subscribe to the same event; cycles are rejected. Harness deadlines can
be shorter than an individual subscription timeout.

Common events include session.started, request.received, request.completed,
tool.before, tool.after, context.compacting, and context.compacted. Native
unmapped events retain native.<harness>.<event> names. Harness event coverage is
not identical. No current memory.changed/repository.synced event is emitted.

A handler runs from a temporary capability copy, not the user's project. Its
stdin is a JSON event containing schema_version, id, event, assistant_id,
device_id, optional session_id/request_id/native_event/working_directory,
causation_id, and payload. Validate fields needed by the handler. Resolve a
workspace operation from working_directory explicitly. Handlers receive
MY_FRIDAY_ASSISTANT_ROOT, MY_FRIDAY_ASSISTANT_ID, MY_FRIDAY_BIN, and
MY_FRIDAY_DEVICE_ID. Nested lifecycle dispatch is refused.

An empty stdout is accepted; otherwise return JSON with additional_context as a
string (maximum total output 16 KiB). failure: stop stops the subscription chain,
not necessarily the harness action. Retries are not globally exactly-once; design
external effects to reconcile interrupted operations. Do not write credentials or
raw transcripts into capability source, memory, logs, or returned context.
