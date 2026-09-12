# Native memory integrations

My Friday plugins connect existing agents to the shared memory service. They do
not own the agent's identity, skills, provider credentials or orchestration.

Each harness owns one directory:

- `codex/`: current MVP; prove this integration before implementing others.
- `pi/`: follow-up tracked in issue #117, after Codex acceptance.
- `claude-code/`: follow-up tracked in issue #118, after Pi acceptance.

Only Codex is implemented in this goal. Future adapters reuse shared storage,
retrieval, provenance, supersession and synchronization semantics. A plugin may
adapt native tool/hook interfaces, but must document any reduced lifecycle or
context-delivery support rather than claiming all harnesses behave identically.

See [the memory MVP acceptance ledger](../docs/memory-service-mvp.md).
