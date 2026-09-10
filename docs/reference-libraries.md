# Reference-aware capability building

My Friday can link historical memory, documentation and implementation resources
as **reference libraries**. The current user ask defines the capability; reference
evidence helps sharpen requirements, recover experience and derive regression
tests. This extends the existing agent-driven capability authoring workflow; it
does not start a separate model, import a legacy assistant, or make old rules
current. Provider integrations and personal libraries stay private.

## First slice: external local directories

For guided setup, run `my-friday` → **Manage an installed agent → Reference
sources**. Register a source's ID, title, description and intended purpose; review
the portable metadata and checkpoint/sync effects before confirming. Then open
that source and choose **Bind/rebind local directory**. You can register now and
bind later, or bind an existing description after importing an agent on a second
machine. The menu refuses registration while source has unrelated uncommitted
work; avoid concurrent source edits during registration. A saved descriptor is
retained if synchronization is pending or fails—inspect state rather than retrying
registration. Local-only is not a remote backup.

The source submenu offers **View description**, **Bind/rebind local directory**
and **Check availability**. Checks inspect metadata and open the selected directory,
without listing/reading documents or contacting a remote. States are `available`,
`unbound`, `stale`, `invalid` and `unavailable`. Available does not prove readable
documents, eligible text, absence of secrets or Git freshness. Rebinding explicitly
acknowledges the displayed description; a change during review requires reopening
it. Binding changes do not checkpoint source or edit the external directory.

Agents and scripts use the same core check through:

```sh
my-friday reference status --instance /absolute/instance --library historical-notes
```

The result is JSON with `library_id`, `state`, `detail` and, when known, `root`.
Exit 0 means the check completed; inspect `state` for availability. Unknown IDs or
invalid descriptors are errors. Registration/binding/search/read remain the existing
commands below, not simulated menu input. The TUI does not yet edit/remove portable
descriptors or clone/update remote reference repositories.

Supported sources are existing local directories, including Git working trees.
No remote creation, cloning, fetching, authentication provider, script execution,
vector index or format-specific memory converter is involved. Use a narrowly
scoped directory of material the assistant may read. The original remains external
and is never copied into active memory or the private capability inventory by
these commands. Read-only access is not a secret scanner or an OS sandbox.
Library roots must neither overlap nor contain the assistant source or instance
directory; bindings to a broad parent such as the user's home are rejected.

Register a portable description, then bind a local directory on this instance:

```sh
my-friday reference add --instance /absolute/instance \
  --library historical-notes --title 'Historical notes' \
  --description 'Earlier implementation choices and failures' \
  --purpose 'Historical experience to evaluate against the current ask'
my-friday reference bind --instance /absolute/instance \
  --library historical-notes --path /absolute/external/resources
```

Inside a named agent session, the instance/repository environment supplies those
paths. `add` and `list` can also use an explicit `--repository` without an instance.
`add` writes a descriptor and performs the normal source checkpoint/sync; an
instance supplies its checkpoint device, or an unbound caller may use `--device`.
`bind` changes only local metadata and never synchronizes or mutates the library.
Bind again to explicitly change the local target. Existing descriptors cannot be
overwritten by `add`; reviewed descriptor edits use normal source Git history.

## Storage and portability

Portable descriptors live in `.my-friday/references/<id>.json`:

```json
{
  "schema_version": 1,
  "id": "historical-notes",
  "title": "Historical notes",
  "description": "Earlier implementation choices and failures",
  "purpose": "Historical experience, not current operating policy"
}
```

IDs use the standard 3–128-character lowercase identifier format. Text fields
are required and limited to 8192 bytes each; unknown fields/versions are rejected.
Descriptions are user-supplied metadata, not authority to execute anything. Do
not put credentials or machine-specific paths in them.

Local bindings live in `<instance>/references/<id>.json`, outside source Git,
and contain the canonical absolute root and a SHA-256 fingerprint of the
descriptor's parsed fields serialized as compact JSON in schema field order.
This is not the hash of its on-disk whitespace/formatting. Another
machine receives descriptors, not local paths, and must bind its own resources.
Any descriptor value change invalidates the binding until explicitly rebound. Missing
libraries do not block normal source validation, synchronization or launch.
`list` returns descriptions, not verified local availability; search/read report
unbound, stale or unavailable resources. The optional descriptor directory keeps
existing source format 1 compatible; older toolkit builds do not validate it.

## Separate discovery and reading

```sh
my-friday reference list --instance /absolute/instance
my-friday reference search --instance /absolute/instance \
  --library historical-notes --query 'retry failure'
my-friday reference read --instance /absolute/instance \
  --library historical-notes --path notes/retries.md --sha256 HASH_FROM_SEARCH
```

Search returns matching paths, content SHA-256 hashes, byte counts and lexical
scores. It does not return source snippets. Read returns the selected text in a
`reference-only` packet with library metadata, descriptor hash, file path, file
hash and an explicit non-authoritative-use notice. Passing the search hash fails
if content changed; omitting it explicitly reads the current version. No output
from either command is automatically injected by lifecycle hooks or copied into
active memory. AGENTS.md and SKILL.md may be read as evidence but are never
registered with the harness by linking a library. Scripts are only read as text.

Reads are confined to the linked directory through a rooted file handle. Paths
must be canonical and relative. Hidden paths (including .git and .env),
node_modules, symlinks, special files, NUL-containing/non-UTF-8 files and files
over 1 MiB are excluded. Search visits at most 2,000 entries and about 32 MiB of
readable text per call (up to one extra 1 MiB file). It returns 1–100 matches,
default 20; query length is limited to 4096 bytes. Empty queries list eligible
files. Skipped entries and truncation are reported. Large-directory enumeration
and filesystem latency are not hard wall-clock/memory bounded. An empty result
does not prove relevant experience is absent: narrow the library, try other
terms, or report the missing evidence. This is lexical retrieval, not semantic RAG.

File hashes identify the actual returned bytes, not Git commit ancestry,
authorship, truth or automatic archival retention. Git working-tree changes are
read as they are; My Friday does not claim they are committed or take a whole-tree
snapshot. Keep the original repository/history or a separate versioned snapshot
if exact later retrieval matters. Do not retrofit import-time device attribution
as original authorship.

## Capability workflow and rationale

`agent capability-guide` now instructs the agent to establish today's ask first,
consult relevant references, reconcile evidence, implement under the My Friday
contract, and verify actual behavior. `agent capability-rationale` prints a
template for private `capabilities/<id>/RATIONALE.md`. It records:

- Current ask, scope, allowed effects and acceptance tests.
- Consulted library IDs, descriptor/file hashes and relative file paths.
- Useful experiences, reused/adapted implementation and checks proving the fit.
- Rejected or superseded old assumptions, with reasons.
- Uncertainty/material decisions and verified outcomes for active learning.

The rationale is an authoring convention, not a new required manifest field or
an automated judgment that the adaptation is sound. Existing capabilities remain
valid. Its contents do not replace runtime instructions or executable checks.
Sound code can be reused after inspection, license review and testing; 1:1
instruction transplantation and compulsory rewriting are both inappropriate.

Reading a reference does not authorize executing its commands, following links,
installing skills, changing account policy or promoting its facts/preferences.
The agent still runs with full user access, so these are explicit workflow and
loading boundaries—not a guarantee that prose cannot influence the model or that
other tools cannot access the same files. Existing native host/project inheritance
is intentionally unchanged. Do not launch the agent in a reference directory or
place reference trees into native auto-discovery paths.

## Verification and next step

Automated tests cover descriptor validation, source sync without local resources,
instance-local bindings, stale descriptor/file detection, bounded search results,
unsafe paths/content and binding symlinks, non-execution of reference scripts,
absence from normal recall/capabilities/projections, and the CLI authoring flow.
These do not prove model behavior. The initial synthetic behavioral pilot passed
on 2026-09-09: Codex created a private capability using past failure cases while
rejecting obsolete rules; a fresh Pi session discovered/reused that capability
and verified its rationale's reference hashes without modifying source or the
library. See [pilot evidence](testing/portable-pilot.md). This bounded result is
not a general prompt-injection guarantee. Real legacy material still requires
incremental reconciliation against the current ask, one capability at a time.
