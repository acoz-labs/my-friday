# Agent-ready management

Any coding agent with a command-execution tool can manage My Friday without a
terminal UI, MCP server or harness-specific plugin. Installed agents receive
the discovery instructions through the shared generated projection. Existing
memory, reference and capability commands remain their established JSON CLI;
this API completes the **installation/management** surface rather than wrapping
arbitrary shell commands or replacing working memory tooling.

Start with:

```sh
my-friday --version
my-friday api describe
```

`agent_api_version: 1` in version metadata identifies this interface. Older
toolkits without that field should not be assumed to implement it. `api describe`
returns versioned action descriptions, parameters, parameter JSON Schemas,
required fields, side effects, network use and stopped-session requirements.
Schemas describe field structure; action descriptions and runtime checks add
cross-field and operational constraints. Discovery never executes an action or
reads credential values. The private Git credential endpoint is not exposed.

## Requests and outcomes

Send exactly one JSON request through stdin or a request file:

```sh
my-friday api --input /absolute/path/to/request.json
# Or: my-friday api --input -   (JSON on stdin, followed by EOF)
```

Example read-only request:

```json
{
  "schema_version": 1,
  "id": "health-check-01",
  "action": "agent.doctor",
  "params": {"instance": "/absolute/path/to/instance"}
}
```

An installed agent obtains the exact current path from `MY_FRIDAY_INSTANCE` and
the executable from `MY_FRIDAY_BIN`; these variables need normal shell expansion
or JSON serialization by the caller, not literal strings inside JSON. The caller's
project directory is never silently used as an instance/source target. Other
agents can discover standard installations with `agent.list`, or supply a known
custom instances directory. Discovery is local and does not scan the user's home.

Each normal result is one JSON envelope on stdout, without progress chatter or
terminal control codes. Errors retain structured results where useful:

```json
{
  "schema_version": 1,
  "id": "health-check-01",
  "action": "agent.doctor",
  "ok": false,
  "state": "needs_attention",
  "result": {"healthy": false, "checks": []},
  "error": {
    "code": "installation.unhealthy",
    "message": "installation needs attention; see doctor checks and remedies",
    "retryable": false
  }
}
```

That is an abbreviated example; actual doctor results include individual remedies.
Exit 0 means the requested read/preview/operation completed as reported, **not**
that all possible downstream effects succeeded. Exit 2 denotes invalid input or
a required execution prerequisite; exit 3 denotes an operation needing attention.
The normal error path does not append human error text to stderr. Broken pipes,
process termination or OS-level failures can prevent a complete response; always
inspect state before retrying an operation whose outcome was not received.

Use `ok`, `state`, `result` and `error.code` together. Important states include:

- `planned`: parameter-only preview; no operation has run.
- `completed`: the selected local/read operation completed.
- `synced`: normal remote fetch/reconciliation/push confirmed synchronization.
- `local_only`: local checkpoint succeeded but no remote is configured.
- `pending`: remote sync remains pending, with local work retained.
- `conflict`: source reconciliation needs explicit attention.
- `needs_attention`: structural doctor found problems.
- `invalid_request` / `failed`: inspect the error and any partial result.

`sync.pending` is retryable after addressing the reported cause. This is **not**
an instruction to loop or repeat a mutation blindly. Other stable error categories
include `input.invalid`, `input.version`, `input.unreadable`, `action.unknown`,
`params.invalid`, `sessions.active`, `sessions.self_update`, `environment.home`,
`installation.unhealthy`, `sync.conflict`, `operation.interrupted` and the generic
`operation.failed`. Detailed operational failures may still require inspection;
the API is not a semantic resolver for every Git, filesystem or provider error.

## Explicit writes; no interactive permission loop

Mutating actions default to a preview. To execute the same intended operation,
submit its parameters with `"apply": true`:

```json
{
  "schema_version": 1,
  "action": "agent.repair",
  "apply": true,
  "params": {"instance": "/absolute/path/to/instance"}
}
```

The API never asks a question or reads a menu response. `apply` expresses an
explicit execution request—not authority beyond the user's task. A preview is
not a credential check, full preflight, stored permission, state reservation or
signed approval. Parameters may be validated further when execution begins.
Callers must carry the user's actual scope and any review requirements forward.
Ordinary authorized memory/capability work keeps its existing autonomous behavior;
the API does not insert a new confirmation loop into those workflows.

Unknown or wrongly cased request/parameter fields, duplicate JSON keys, trailing
values, null fields, unsupported versions and wrong parameter types are rejected.
Requests are limited to 1 MiB and nesting to 64 levels; paths must be absolute.
There is one request per process, not a persistent JSON-RPC stream. Optional IDs
are correlation only: there is no durable request deduplication or job queue.
Failure may follow partial progress (`side_effects_may_have_occurred`); creation,
remote setup, checkpoints and artifact staging have their documented recovery
rules. No automatic deletion, forced reconciliation or speculative retry occurs.
Subprocess/network operations use their existing bounded contexts plus an API
cancellation context. Local filesystem operations are synchronous; this is not a
hard-deadline transactional scheduler or an OS sandbox.

## Available management operations

- `system.describe`, `system.version`
- `agent.list`, `agent.inspect`, `agent.doctor`
- `agent.setup`, `agent.repair`, `agent.harness`
- `source.status`, `source.accounts`, `source.configure`, `source.sync`
- `toolkit.latest`, `toolkit.install`, `toolkit.install-release`
- `toolkit.use`, `toolkit.backups`, `toolkit.rollback`

The executable's discovery output is authoritative for parameters. Setup requires
explicit paths and a machine label, with either an explicit launcher or
`no_launcher: true`. Import uses an already-cloned new-format source and retains
its harness default. Changing that default is a separate portable source write.
Source hosting has explicit setup/ongoing accounts and never switches a global
active account; applying it can upload the **complete committed source history**.
Review the [source-setup](source-setup.md) boundary and prerequisites first.

Installing a toolkit and changing an agent's pin are separate actions. A local
artifact requires a trusted expected SHA-256. Release installation requires an
explicit version/digest matching the currently eligible latest release. Staging
and compatibility checks happen before optional command activation. Other agent
pins stay unchanged. See [management](management-menu.md) for integrity/trust,
retained artifacts and rollback limits. No provider credentials are accepted as
API parameters, and this is not a secret broker or general GitHub automation API.

## An agent managing itself

An active installed agent can inspect its state, run doctor, perform scoped source
work and stage an approved toolkit using the same API. Repair can refresh managed
files, but returns `requires_fresh_session`; that does not reload instructions
already held by a running model. `toolkit.use` and `toolkit.rollback` require
`sessions_stopped: true`, and refuse a known live self-upgrade when the inherited
`MY_FRIDAY_INSTANCE` matches the target. Stage/inspect during the conversation;
adopt or roll back from a separate management session after it exits.

This guard is operational protection, not process isolation or a claim to detect
all active harnesses. Callers must close other affected sessions too; unsetting
the environment variable is not a substitute. No capability should automate away
the guard or claim that changing files refreshed current model context.

A future MCP adapter can translate these same typed actions. It should not add a
second implementation of management behavior or require an agent to operate the
TUI. No MCP daemon or harness-specific integration is required by this increment.
