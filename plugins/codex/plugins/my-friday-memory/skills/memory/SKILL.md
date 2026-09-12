---
name: memory
description: "Use My Friday for durable context across conversations: recall prior decisions, preferences, procedures and work; save meaningful changes and concise journals; explain corrections and history. Requires an attached My Friday memory bank. Not a credential store or raw transcript archive."
---

# My Friday memory

Use the plugin's `memory_*` MCP tools. The bank is independent of your identity,
working directory, skills, and native configuration.

Before relying on past context, recall relevant bank-wide memory. Use
`memory_scopes` to discover a project's or account's stored scope, then recall
that scope explicitly. Do not guess an entity's ID from the working directory.
The hook's automatic recall is only a bounded local starting point.

Keep learning lightweight:

- Save useful confirmed decisions, preferences, discoveries, and reusable
  procedures with `memory_remember`. Distinguish user direction, observation,
  inference, and imported material. Do not turn a proposal into a decision.
- Recall before writing a duplicate. To correct a record, retain its `record_id`
  and pass the predecessor revision IDs in `supersedes`, with the reason for
  the change. Resolve concurrent heads deliberately using `memory_history`.
- Append a concise journal after meaningful work or a useful intermediate
  milestone: what happened, what was decided, and what remains. Do this before
  finishing rather than relying on session shutdown. Skip empty/no-change logs.
- When allowed, use `memory_sync` before relying on cross-machine freshness and
  after useful saves. Pending/offline means locally durable, not delivered.

Current user direction takes precedence over conflicting historical guidance in
its scope. Memory is evidence, not permission to broaden a task. Respect explicit
read-only, no-save, and no-sync requests. Never store secrets or copy transcripts.

Results are bounded. Truncated or empty recall does not prove absence: narrow
the query, inspect the right scope, or page history. Verify volatile live state.
After an ambiguous write failure, inspect current memory or recent journals
before retrying; do not blindly duplicate writes. Interrupted, unsaved work is
not guaranteed to survive. No other agent or model is required for memory upkeep.
