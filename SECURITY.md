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

Capability receipts and journals are accepted as mutation authority only after
no-follow regular-file, one-link, owner-mode, canonical-schema, slug, action,
digest, and prior-state validation. Foreign projection, control, and workspace
entries fail closed and are never adopted for cleanup.

Projection lifecycle operations use descriptor-relative no-follow writes,
exclusive no-replace promotion, and identity/digest-bound quarantine before
recursive cleanup. Cleanup unlinks through the opened, identity-bound directory
descriptor; recovery proves quarantine content and retains a deterministic
receipt-derived restoration handle before active-path restoration. A strict
external cleanup manifest binds the root inode and expected per-path content so
partial deletion is resumable while foreign additions fail closed. Manifest
authority itself is sync/close/no-replace promoted; pre-promotion residue is
never authority, and post-root-unlink completion accepts only an absent target.
Same-UID
interference is detected at each ownership boundary; foreign raced bytes are
preserved rather than overwritten or deleted.
