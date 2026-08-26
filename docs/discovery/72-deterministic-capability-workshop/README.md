# Discovery: Recover Capability-Builder-First With A Deterministic Workshop

- **Status:** Final
- **Discovery issue:** #72
- **Discovery PR:** Pending
- **Repository basis:** d40950fd54cfd6bfbea7467de6f83af1d4619f08
- **Recommended decision:** approve
- **Gate 1:** awaiting-authority
- **Confidence:** High
- **Private evidence:** none

## Decision sought

Revise C1 so My Friday ships a deterministic, source-first capability workshop
before governed memory, while deferring natural conversational-agent
orchestration until Codex can pass a real action/effect proof. Decide the
smallest workshop that keeps user ownership, exact review, source-write
authority, activation authority, enhancement, and reversal distinct.

This changes the first authoring interface, not the foundational sequence or
the capability package/lifecycle already selected by discovery #49.

## Audience and critical tasks

The first audience remains technically capable Codex users and My Friday
maintainers. They need to:

1. turn an intent into one complete instruction-only package without memorizing
   JSON fields or directory layout;
2. see every generated source byte and unresolved judgment before a write;
3. create or enhance source only after an exact confirmation that does not
   install or activate it;
4. inspect, validate, and deterministically test the result;
5. enter the existing install, invocation, upgrade, disable, remove, and
   recovery lifecycle separately; and
6. reuse the same outer workflow when governed memory adds a reviewed
   data-bearing profile.

The terminal is the current product surface. Keyboard-only operation, plain
text, interruption, return use, and recovery matter more than visual novelty.

## Evidence

- Discovery #49 and approved C1 at
  `4d41ca094ce5aa9c6fd45057763564b027dcc460` established builder-first as the
  extension substrate and sequenced it before governed memory.
- Issue #51 shipped the strict instruction-only package parser, source versus
  projection split, deterministic tests, explicit lifecycle plans, collision
  and drift refusal, reversal, and recovery. Those assets remain valid.
- Issue #56 proved that current Codex can terminate complex capability-authoring
  turns without any qualifying tool/file action. Bounded controls covered PTY,
  app-server, native exec, exact-thread continuation, explicit model selection,
  complete and staged prompts, root grants, alternate working directories,
  workspace drafts, first-action directives, and structured-output attempts.
- The same model performs ordinary trivial writes, so credentials and model
  availability are not the missing dependency. Model completion, prose, and
  exit zero are therefore not source authority.
- Every probe remained confined and completed exact cleanup; no candidate,
  acceptance, or release was authorized.
- Anthony selected `deterministic workshop now` on 2026-08-26 rather than
  parking builder-first delivery.

There is no representative-user study yet. The first exact-candidate workshop
must be exercised by Anthony and two independent design partners through the
existing typed receipt boundary before release.

## Assumptions

- Users will accept a guided terminal workshop when it eliminates schema recall
  and preserves preview, control, reversibility, and local ownership.
- `Create source` and `Update source` are understandable authorities when the
  summary explicitly says `Installed: no` and the later lifecycle retains its
  own `Install` or `Upgrade` confirmation.
- One internal proposal model can power create and enhance without exposing a
  second public package schema.
- A future agent adapter can supply answers to the same workshop boundary; it
  does not require prebuilding an app-server client or model protocol today.

## Unknowns

- Which prompt defaults reduce typing without causing users to overlook a
  consequential behavior. The selected outcome resolves this conservatively:
  safe display formatting may default; triggers, outputs, success/failure
  behavior, non-triggers, examples, and required facts may not.
- Whether design partners prefer question-by-question review or one editable
  specification view. V1 chooses sequential questions plus a complete final
  source preview; acceptance records comprehension and recovery need.
- How a future data-bearing profile will express permissions and storage. It is
  explicitly outside this instruction-only slice and returns to design in #52.

## Competing options

### A. Deterministic interactive workshop over the existing package contract

`my-friday capability workshop NAME SLUG` asks bounded questions, renders the
canonical three-file package, shows a complete diff, requires exact source-write
confirmation, commits source atomically, and runs inspect/validate/test. On an
existing valid package it loads current structured values, preserves allowed
opaque reference/asset files, previews only intended core-file changes, and
requires a distinct update confirmation.

This is the smallest path that users can operate now, keeps one package
authority, supports enhancement, and can later accept agent-supplied answers.

### B. Public JSON proposal file plus noninteractive apply

A declarative file is automation-friendly but introduces another public
contract, makes first use depend on schema knowledge, and invites an agent
adapter before agent reliability exists. The internal proposal may be designed
for later reuse, but no public `--spec` surface ships in this slice.

### C. Deterministic scaffold with direct source editing for enhancement

This creates the initial files but abandons the promised repeatable enhancement
workflow and forces users back into schema/manual-file expertise. Rejected.

### D. Keep agent-first acceptance parked

This preserves the original interface claim but blocks the extension substrate
and governed-memory dogfood on an external behavior with no reliable current
path. Anthony explicitly chose not to park.

### E. Accept driver-authored fixtures or model prose

This would make release evidence green without proving a usable builder. It is
rejected as false authority.

## Decision

Select option A.

The workshop is a deterministic terminal interaction over the existing strict
package contract. It owns collection and canonical rendering, but the user owns
the decision to write source. It never installs or activates. After a successful
write it automatically reports exact candidate inspect state and runs structural
validation/tests; failures leave source inspectable and inactive.

### Conceptual model and information architecture

- **Capability source:** the canonical user-owned package under the private
  runtime.
- **Workshop answers:** in-memory, bounded values used to render a proposed
  source change; they are not a second durable record.
- **Preview:** complete source diff plus explicit unresolved/invalid answers.
- **Source write:** one separately confirmed create or update transaction.
- **Activation:** the existing later install/upgrade lifecycle, never part of
  workshop confirmation.
- **Agent adapter:** a deferred input adapter to workshop answers, not a source
  writer or lifecycle authority.

### Primary create flow

1. Preflight exact named instance/runtime/slug and require a TTY.
2. Ask purpose/display/summary, explicit triggers and non-triggers, inputs and
   outputs, success and failure behavior, examples, and required facts.
3. Show the instruction-only profile's seven prohibited effects as fixed
   `none`, rather than asking users to approve unavailable powers.
4. Validate each answer locally and allow `b` to go back, `q`/EOF/interrupt to
   exit without a write, and a final review to restart a section.
5. Render all three canonical files and a complete source diff. Clearly state
   `Source action: create` and `Installed: no`.
6. Require exact `Create source`; default Enter exits.
7. Journal and atomically write only the absent source package, then run
   inspect/validate/test and report the exact inactive state and next command.

### Enhancement flow

An existing valid package opens the same workshop with current structured
values shown as defaults. The user may retain or change each value. Optional
allowed `skill/references/` and `skill/assets/` bytes are preserved exactly and
listed as unchanged. The final diff states `Source action: update` and
`Installed: unchanged`. Exact `Update source` authorizes only the core source
replacement. If an installed projection exists, the result becomes
`source-changed`; the existing stale-plan-safe `Upgrade` flow remains separate.

Invalid, drifted, interrupted, collided, or incompatible source/control state
does not enter the workshop. It reports the current state and existing
inspect/recovery action without attempting repair.

### Interaction contract

- Plain text with stable numbered sections; color is optional and never carries
  meaning.
- Every prompt names its format and bound. Secret input is never requested.
- Empty/default behavior is explicit; consequential answers have no invented
  defaults.
- `b`, `q`, EOF, `INT`, and `TERM` preserve source. Focus is the terminal cursor;
  no mouse, motion, or screen-size dependency exists.
- Unicode input follows existing profile validation; parsing and confirmation
  tokens are locale-independent. V1 product copy is English and does not claim
  localization.
- A narrow terminal repeats/wraps text without truncating source or hiding the
  action/installed summary.
- No analytics or model transcript is introduced. Typed release receipts retain
  only comprehension, completion, recovery, and retention facts already
  approved for #51.

## Success and stop signals

- **Success:** a first-time user can create `daily-brief`, understand that it is
  inactive, pass inspect/validate/test, install it separately, return to enhance
  it, observe `source-changed`, upgrade separately, and remove it while source
  remains. Anthony and two design partners complete this against one immutable
  candidate without operator correction.
- **Change:** users cannot understand source-write versus activation, or
  enhancement cannot preserve allowed opaque bytes. Revise the workshop before
  adding an agent adapter.
- **Pause:** atomic source create/update and interruption recovery cannot be
  proven without granting deletion or activation authority.
- **Stop:** the implementation forks package validation, hides generated bytes,
  adds a public proposal schema solely for future automation, or claims natural
  agent authorship without an action/effect proof.

## Candidate outcome map

### D1 — Deterministic instruction-only capability workshop

- Disposition: selected
- Outcome: A user can create and enhance a canonical instruction-only
  capability through one guided terminal workshop with complete diff, explicit
  source-write confirmation, automatic deterministic checks, and no activation.
- Acceptance: Create and update paths, back/exit/interruption, invalid/collision/
  drift refusal, opaque-byte preservation, exact diff, separate install/upgrade,
  full reversal, ambient preservation, and three-person exact-candidate workshop
  receipts all pass.
- Dependencies: Existing #51 package/lifecycle; no new external dependency,
  credential, model, network, or public schema.
- Sequence: P1 and immediate; replace the failed live-agent authoring portion of
  #56, then complete #51 release before #52 governed-memory dogfood.

### D2 — Agent adapter for workshop answers

- Disposition: deferred
- Outcome: A conversational agent can propose or supply workshop answers while
  deterministic preview, source-write confirmation, and lifecycle authority
  remain owned by My Friday and the user.
- Acceptance: A supported Codex surface must produce a terminal structured
  proposal or qualifying action/effect reliably across bounded repeated runs;
  prose, exit zero, or catalog discovery alone never qualifies.
- Dependencies: Shipped D1 and verified upstream agent behavior.
- Sequence: Reassess after D1 dogfood and Codex compatibility evidence; it does
  not block #51 or #52.

### D3 — Public proposal-file automation surface

- Disposition: rejected
- Outcome: None in this discovery.
- Acceptance: Not applicable.
- Dependencies: Would require demonstrated noninteractive user demand and a
  separately reviewed public contract.
- Sequence: Do not materialize.

## Privacy and evidence handling

All evidence is sanitized repository behavior and typed boolean results already
recorded on issue #56. No private paths, transcripts, prompt bodies, auth bytes,
credentials, or preserved diagnostic roots enter this pack. The workshop stores
no answers outside the canonical source transaction and adds no telemetry.

## Decision Spotlight

- **Ship a deterministic workshop now:** capability ownership and dogfooding
  progress without claiming unreliable agent execution.
- **Keep one package contract:** workshop answers are ephemeral input to the
  existing strict parser/renderer, not a second durable schema.
- **Separate source write from activation:** exact `Create source` or
  `Update source` never installs; existing lifecycle confirmations remain.
- **No consequential defaults:** unavailable powers are fixed `none`; behavior,
  triggers, examples, and failures require explicit user answers.
- **Preserve enhancement ownership:** valid existing source is loaded, optional
  references/assets are retained exactly, and the full core-file diff is shown.
- **Defer the agent adapter honestly:** future compatibility is an input seam and
  evidence gate, not dormant protocol code in D1.

## Gate 1

The final candidate awaits product authority and an authorized approval on the
exact pull-request head.
