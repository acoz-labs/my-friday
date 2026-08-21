# Architecture

My Friday is a local native command that compiles normalized wizard answers
into deterministic, previewed plans. It creates two separate user-owned Git
repositories and can project one manifest-owned runtime `AGENTS.md` into the
effective Codex home. It has no service, database, network integration, Codex
executable installer, or background process.

```mermaid
flowchart LR
  T[Line-oriented terminal] --> P[Profile and path validation]
  P --> C[Canonical creation plan]
  C --> V[Read-only preview]
  C --> X[Recoverable transaction]
  X --> R[Runtime repository]
  X --> M[Memory repository]
  R --> Z[Pair validator]
  M --> Z
  R --> B[Managed Codex baseline plan]
  B --> H[CODEX_HOME/AGENTS.md]
```

| Component | Responsibility | Boundary |
|---|---|---|
| `cmd/my-friday` and `internal/terminal` | Commands and sequential prompts | No repository policy |
| `internal/profile` and `internal/plan` | NFC text, grapheme limits, IDs, files/actions, plan digest | Pure values; no I/O |
| `internal/repository` | Render, empty-template Git init, schema/Git validation | Git is the only subprocess |
| `internal/transaction` | Stage, validate, promote, roll back, recover | No overwrite of non-empty targets |
| `internal/codexhome` | Inspect, preview, mutate, verify, and recover one projection | Authority limited to `AGENTS.md` and `.my-friday` |

Runtime identity and governed memory share only a deterministic non-secret
assistant identifier. Absolute paths and plan IDs are not written into either
repository. Profile values remain JSON data and never enter generated Markdown
policy. See [repository bootstrap](architecture/repository-bootstrap.md) and
[ADR 0001](decisions/0001-native-bootstrap-command.md). The installed-state
boundary is specified in [installed Codex baseline](architecture/installed-codex-baseline.md)
and [ADR 0002](decisions/0002-manifest-owned-codex-baseline.md).
