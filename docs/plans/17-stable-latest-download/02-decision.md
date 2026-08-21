# Solution Decision

## Decision Drivers

The mechanism must preserve accepted executable bytes, make the permanent URL
work on every release, distinguish executable/archive proofs, fail closed on
ambiguity, tolerate retries, avoid mutable rebuilds, work in the repository's
artifact profile, and provide a bounded path for the current release.

## Competing Approaches

1. **Accepted executable plus deterministic release packaging.** Nomination
   uploads the raw executable and records its digest; release downloads it,
   verifies it, packages it once, and publishes the stable archive/checksums.
2. **Make the archive the accepted artifact.** Package before acceptance and
   bind the ledger only to archive bytes.
3. **Publish a raw stable executable.** Change the public URL to omit `.tar.gz`.
4. **Resolve “latest” dynamically elsewhere.** Keep versioned assets and use an
   API, Worker, or redirect that discovers the newest filename.

## Adversarial Comparison

Approach 1 preserves the existing ledger meaning and accepted executable, but
requires two clearly labelled digests. Approach 2 gives literal archive byte
identity yet changes the meaning of the accepted immutable artifact and makes
executable proof indirect. Approach 3 is mechanically simplest but contradicts
the approved URL and offers a poorer macOS distribution format. Approach 4
adds a runtime dependency and moves release truth outside the owning repository.

## Selected Approach

Select approach 1 with high confidence. Nomination produces one raw executable
artifact from the exact successful commit and binds the candidate authority to
its SHA-256. Release downloads that exact Actions artifact by run/name, checks
the digest, creates a reproducible tar.gz without modifying the executable,
checks the archive digest, and passes both assets to a hardened finalizer.

The archive contains one top-level file named `my-friday`. Tar metadata is
normalized (stable path, mode, ownership, ordering, and timestamp) and gzip
metadata is normalized so retries over the same executable yield identical
archive bytes. New releases use only the stable archive name; tag names provide
version identity. The existing commit-suffixed archive is retained solely as
historical/backfill source evidence.

## Decisions Ledger

| Decision | Selected default | Rationale |
| --- | --- | --- |
| Ledger authority | Raw executable SHA-256 | Preserves current acceptance semantics |
| Public format | Deterministic `.tar.gz` | Approved URL and reproducible retries |
| Archive content | One `my-friday` executable | Predictable install/verification path |
| Stable identity | Filename constant, tag versioned | Makes GitHub latest-download durable |
| Checksum file | Explicit archive and executable entries | Avoids false byte-identity claims |
| Existing alias mismatch | Fail; never clobber | Prevents silent public substitution |
| Backfill source | Current retained archive after verification | No rebuild and no deletion of evidence |
