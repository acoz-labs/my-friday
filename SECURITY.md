# Security

Do not commit secrets or paste secret values into issues, pull requests, logs,
or automation transcripts.

Security-sensitive changes require explicit risk notes, validation evidence, and
review before merge.

Named instances isolate lifecycle authority, not reads or hostile processes
under the same UID. Credentials belong only in the selected instance Codex
home. Manifests, previews, logs, fixtures, and acceptance evidence must never
contain credential values or raw ambient configuration.

Instruction-only capability validation is structural, not semantic safety
certification. It excludes executable files, scripts, dependencies, network
and credential declarations, background work, durable capability data,
publishing, links, devices, and unknown entries. Natural-language instructions
can still be harmful, so users must inspect the complete Git diff and exact CLI
plan. TTY plus exact-token confirmation prevents accidental piping; it does not
authenticate a human or isolate against a malicious same-UID process.
