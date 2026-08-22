# Solution Decision

## Decision Drivers

The boundary must protect the live Codex home and deployed runtime projection,
exercise unmodified nominated bytes, need no administrator action, work on the
declared Apple Silicon/macOS/APFS stack, fail closed, support authenticated
Codex discovery, retain reviewable proof, and clean up without broad deletion.
It should be proportionate to a foreground tool that refuses root and owns only
manifest-proven files below `CODEX_HOME`.

## Competing Approaches

### A. Keep the approved disposable account

Create a unique non-admin account/home/login keychain, accept as that UID, then
remove only the marker-proven account and home. This gives the strongest local
principal separation, but requires physical administrator authentication and
adds account creation/deletion risk unrelated to product behavior.

### B. Use only temporary directories and synthetic environment variables

Run under the accepting UID with `HOME`, `CODEX_HOME`, and temporary paths in a
normal temporary directory. This is simple and no-admin, but a path bug or
malicious descendant can write elsewhere with all ordinary user authority.

### C. Use only an APFS disk image

Place every intended path on an attached disposable APFS volume. This provides
a clear filesystem and cleanup boundary and exercises APFS semantics, but it
does not constrain the same-UID process from writing outside the mount.

### D. Use APFS, a fail-closed sandbox, and protected-state proofs

Combine an explicit disposable APFS mount with synthetic paths, a reviewed
macOS sandbox profile, network separation, pre/post protected-state manifests,
write/network positive controls, canaries, and marker-bound cleanup. This is
the selected approach.

### E. Use a fresh macOS VM

A disposable VM supplies strong OS/principal/keychain separation without
changing host accounts. It adds image lifecycle, licensing/cost, virtualization
tooling, artifact transfer, secret injection, and a second operational platform
that this repository does not otherwise own.

## Adversarial Comparison

| Approach | Strongest property | Fatal or material weakness | Judgment |
| --- | --- | --- | --- |
| Disposable account | Distinct UID, login home, fresh keychain | Physical admin gate and destructive identity authority | Secure but disproportionate for this product |
| Temp roots | Lowest implementation cost | No enforced write boundary | Reject |
| APFS only | Disposable filesystem/device boundary | Same-UID process can escape mount | Reject alone |
| APFS + sandbox + proofs | No-admin enforced write containment with observable invariants | Same UID/read authority; deprecated sandbox must be build-tested | Select with explicit limits |
| macOS VM | Strongest isolation and cleanup | New platform/cost/operations far beyond current product | Defer unless selected path becomes unavailable |

The selected mechanism deliberately layers controls that fail independently.
A synthetic path mistake is caught by the APFS/device checks; an attempted
escape is denied by the sandbox; a profile or supervisor error is caught by
positive controls and protected-state comparison. None is represented as a
distinct-UID or confidentiality proof.

## Selected Approach

Implement a repository-owned local acceptance supervisor. It creates one
unique absent run directory and one evidence directory beneath the canonical
real home, creates and attaches a sparse APFS image, builds an exact reviewed
sandbox profile, and runs all candidate commands against synthetic state on the
volume. It downloads and verifies `artifact-v1` exact bytes before use. Lifecycle
scenarios deny network; a separately labeled real-Codex smoke profile permits
only required network while retaining write containment.

The supervisor snapshots protected metadata before setup and after cleanup,
uses non-sensitive canaries for active denial proof, externally interrupts an
ordinary production transaction after observing a durable journal phase, and
publishes sanitized issue evidence whose comment ID/body digest becomes part of
acceptance authority. Cleanup revalidates every recorded identity and refuses
force detach or deletion on mismatch.

Confidence is high for the APFS/no-admin primitive and medium for the deprecated
sandbox contract across future builds. Therefore platform preflight is part of
acceptance, not an installation assumption. If Apple removes or materially
changes the sandbox mechanism, acceptance blocks and design reopens.

## Decisions Ledger

| Decision | Rationale | Evidence or constraint |
| --- | --- | --- |
| Amend PR #15; do not rewrite it | Preserve exact approved and shipped provenance | PR #15 and PR #19 are merged history |
| Re-nominate after harness implementation | Acceptance must include every lifecycle-linked implementation merge | Repository exact-candidate policy |
| Mount below real home | Satisfies product's home-descendant and local APFS checks without admin | Shipped `codexHome`/home-root contract and host probe |
| Separate volume and evidence roots | Supervisor-owned evidence must survive detach while remaining outside candidate write authority | Acceptance/evidence lifecycle |
| Deny lifecycle network; allow smoke network | My Friday lifecycle needs none; real Codex provider smoke does | Least privilege and standard Codex auth |
| Permit no unsandboxed candidate fallback | Otherwise acceptance could silently lose its primary write boundary | Trust-boundary requirement |
| Aggregate protected manifests without content hashes | Prove mutation absence without exposing same-UID secrets | Evidence privacy requirement |
| Bind acceptance to evidence comment ID and body digest | A mutable/deleted comment must not retain release authority | Current workflow accepts opaque evidence text |
| Non-forced, marker-bound cleanup | Ambiguity should preserve evidence, not risk unrelated data | Destructive-action safety |
| Disclose same-UID/keychain limitations | The chosen boundary does not and need not prove account behavior | Product is nonprivileged and account/keychain agnostic |
