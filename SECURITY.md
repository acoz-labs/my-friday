# Security

Do not commit secrets or paste secret values into issues, pull requests, logs,
or automation transcripts.

Security-sensitive changes require explicit risk notes, validation evidence, and
review before merge.

Named instances isolate lifecycle authority, not reads or hostile processes
under the same UID. Credentials belong only in the selected instance Codex
home. Manifests, previews, logs, fixtures, and acceptance evidence must never
contain credential values or raw ambient configuration.
