---
id: SPEC-HARNESS-LOOP-REPAIR-001
title: "Harness self-learning loop repair — proposal layout contract + decision-signal observation + lesson-channel unification"
version: "0.2.0"
status: in-progress
created: 2026-07-27
updated: 2026-07-27
author: manager-spec
priority: P1
phase: "v3.0.x"
module: "internal/cli, internal/cli/harness, internal/harness/proposalgen, .claude/rules/moai, .claude/skills/moai"
lifecycle: spec-anchored
tags: "harness-learning, proposal-layout-contract, routing-ledger, lessons-channel, falsifiability, self-learning-loop"
era: V3R6
tier: L
depends_on: [SPEC-HARNESS-LOOP-CLOSURE-001, SPEC-HARNESS-APPLY-EXECUTE-001, SPEC-HARNESS-OUTCOME-CAPTURE-001, SPEC-HARNESS-EVO-PIPE-REPAIR-001]
---

# SPEC-HARNESS-LOOP-REPAIR-001 — Harness Self-Learning Loop Repair

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-27 | manager-spec | Initial draft — full-surface audit of the goal skill, the recursive self-learning subsystem, and the harness CLI. Root cause isolated to a producer/consumer directory-layout contract mismatch that survived four completed predecessor SPECs. |
| 0.2.0 | 2026-07-27 | manager-spec | Post-M1 amendment. §A.3 corrected: the layout mismatch was only the FIRST of TWO independent causes — the fields the consumer requires were never collected anywhere upstream (§A.3.2). §A.4 rewritten around the species distinction (discovery report vs edit instruction) that explains the four-SPEC recurrence. AC-HLR-004 rewritten (success signal is promotion into a SPEC, not an `apply_outcome` record) and REQ-HLR-004 amended to match; AC-HLR-005 rewritten (`applied/` does not materialise from this path). M2 redefined; duplicate-draft defect (REQ-HLR-011) and tier-split disposition scoped. §E converted to an index — `acceptance.md` is now the AC SSOT. |

---

## §A Context (Why)

### A.1 The recurring condition

Four completed SPECs have each diagnosed the same condition — *the harness learning subsystem is fully wired but its loop has never closed*:

| Predecessor | Closed | Recorded diagnosis |
|---|---|---|
| SPEC-HARNESS-LOOP-CLOSURE-001 | 2026-06-14 | "fully wired but has never closed its loop" — zero applies executed |
| SPEC-HARNESS-OUTCOME-CAPTURE-001 | 2026-06-14 | observer captures WHAT, carries no field for apply OUTCOME |
| SPEC-HARNESS-APPLY-EXECUTE-001 | 2026-06-15 | `Applier.Apply()` had zero production callers; apply-outcome telemetry zero records |
| SPEC-HARNESS-EVO-PIPE-REPAIR-001 | 2026-07-03 | recursive self-improvement path structurally at zero coverage |

Each predecessor delivered its stated component. `Applier.Apply()` now has a production caller (`internal/cli/harness/execute.go:189` via `RunExecute`). Yet the loop is still open.

### A.2 Measured baseline (2026-07-27, live tree)

| Signal | Command / path | Observed |
|---|---|---|
| Draft proposals on disk | `find .moai/harness/proposals -name spec.md` | **52** (`status: draft`), oldest 2026-06-17 |
| Pending proposals per CLI | `moai harness status` | **`pending proposals: 0 items`** |
| Next pending per CLI | `moai harness apply` | **`No pending proposals.`** |
| Applies ever executed | `.moai/harness/learning-history/applied/` | **directory absent** |
| Apply telemetry | `grep -c apply_outcome .moai/harness/usage-log.jsonl` | **0** |
| Observation volume | `wc -l .moai/harness/usage-log.jsonl` | 79,060 lines |
| Observation composition | event-type histogram | `agent_invocation` **74,749 (94.5%)**, `tool_failure` 905, `user_prompt` 1,138, `subagent_stop` 1,157, `session_stop` 1,111 |
| Routing decisions recorded | `.moai/state/routing-ledger.jsonl` | **file absent** until this audit wrote one row |
| Falsifiability fields in lessons | `grep -c 'prediction:\|verified:'` across auto-memory | **0 / 102 feedback files + lessons.md** |
| Constitution-designated lesson store | `lessons.md` mtime | **2026-06-17** (40 days stale); live traffic is `feedback_*.md` (102 files) |
| Lesson auto-capture queue | `.moai/lessons-inbox.jsonl` | 845 lines, appended today, **never drained** |

### A.3 Root cause — TWO independent causes

The loop is broken in two places, not one. §A.3.1 was diagnosed first and repaired in M1; §A.3.2 was discovered while verifying M1 and is the reason M1 did not close the loop. Repairing the layout makes drafts **visible**; it does not make them **applicable**.

Every file:line below was read on the worktree tree at `feat/SPEC-HARNESS-LOOP-REPAIR-001`. Symbol names are given alongside line numbers because line numbers drift.

#### A.3.1 Cause 1 — producer/consumer layout contract mismatch (repaired in M1)

The proposal **producer** writes a directory per draft (`internal/harness/proposalgen/scaffolder.go:94,101,108`, `WriteProposals`):

```
.moai/harness/proposals/<DRAFT-ID>/
  ├── spec.md          # status: draft
  └── proposal.json    # {tier, pattern_key, confidence, observation_count, ...}
```

Every proposal **consumer** assumes a flat file named `<DRAFT-ID>.json` directly under `proposals/`:

| # | Site | Predicate | Consequence |
|---|---|---|---|
| C1 | `internal/cli/harness.go:186-198` `countProposals` | `!e.IsDir() && strings.HasSuffix(e.Name(), ".json")` | `status` reports 0 pending, always |
| C2 | `internal/cli/harness.go:270-274` apply selector | same predicate | `apply` returns "No pending proposals", always |
| C3 | `internal/cli/harness/execute.go:203` `resolveProposalPath` | `filepath.Join(root, dir, id+".json")` | `execute --id` returns "proposal not found", always |

`!e.IsDir()` excludes every generated draft by construction. The mismatch is total and silent: no error surfaces, the CLI simply reports an empty queue.

This break explains the **visibility** symptoms in §A.2 — `status` reporting 0, `apply` finding nothing, `execute --id` reporting "proposal not found". M1 repaired it via a shared accessor (`internal/harness/proposalgen/layout.go`); `moai harness status` now reports 52 items against the same tree that previously reported 0.

It does **not** explain the remaining §A.2 symptoms (zero applies, absent `applied/`, zero `apply_outcome`). Those have a separate cause.

#### A.3.2 Cause 2 — the required fields were never collected anywhere upstream

The consumer's payload type requires five fields that **no upstream stage ever records**. This is not a serialization mismatch that a mapping layer could bridge; the information does not exist at any point in the pipeline.

Tracing the data chain end to end:

| Stage | Type | Location | What it carries |
|---|---|---|---|
| Observation | `harness.Event` | `internal/harness/types.go:87-217` | 31 fields — timestamp, event type, subject, context hash, tier increment, plus prompt/agent/outcome/weight extras. **No file path, no frontmatter field name, no proposed value.** |
| Aggregation | `harness.Pattern` | `internal/harness/types.go:287-309` | 7 fields — key, event type, subject, context hash, count, confidence, tier |
| Promotion | `harness.Promotion` | `internal/harness/types.go:313-331` | 6 fields — ts, pattern key, from/to tier, observation count, confidence |
| Candidate | `ProposalCandidate` | `internal/harness/proposalgen/types.go:33-56` | 6 fields — pattern key, observation count, confidence, tier, source ts, draft ID |
| On disk | `proposal.json` | `internal/harness/proposalgen/scaffolder.go:168-185`, `marshalProposalJSON` | an 8-key map: `pattern_key`, `observation_count`, `confidence`, `tier`, `source_ts`, `generated_at`, `generator_version`, `draft_id` |
| **Consumer** | `harness.Proposal` | `internal/harness/types.go:342-366` | requires `id`, `target_path`, `field_key`, `new_value`, `created_at` — **none of which has any upstream source** |

Two observations pin the diagnosis:

- **The loss point is identifiable.** `internal/cli/hook.go:606-610` (`runHarnessObserve`) receives the full PostToolUse payload — `hook.HookInput` carries `ToolInput json.RawMessage` (`internal/hook/types.go:217`), which holds the edited file path — and extracts **only** `hookInput.ToolName` into `Event.Subject`, discarding the rest. Aggregation then collapses further: `internal/harness/learner.go:72-84` reduces N events to a `Count`, keeping nothing else.
- **No production code constructs a populated `harness.Proposal`.** The only non-test `harness.Proposal{` sites in the tree are three empty error returns in `internal/cli/harness/execute.go:218,220,224`. Every populated instance lives in test fixtures or in the documented example payload at `.claude/skills/moai-harness-learner/SKILL.md`.

So the `Applier` has always been waiting for a payload shape that nothing in the system produces.

### A.4 Why it survived four SPECs — the two artifacts are different species

The deeper finding is that `proposalgen` and `Applier` were never two ends of one pipeline. They are **two different machines**, and the wiring between them was an assumption nobody stated.

**What `proposalgen` produces is a pattern-DISCOVERY REPORT.** Its own package doc names the intended consumer: the rendered `spec.md` is the "downstream manager-spec authoring target" (`internal/harness/proposalgen/scaffolder.go:7-8`). Each draft is a directory holding a human-readable `spec.md` with EARS-style placeholders (`renderSpecMd`, `scaffolder.go:120-165`) beside machine metadata. It answers *"this pattern recurred — is it worth a SPEC?"* and its natural consumer is a human or an authoring agent.

**What `Applier` consumes is a frontmatter EDIT INSTRUCTION.** `applyFileModification` (`internal/harness/applier.go:463-479`) switches on `proposal.FieldKey`: `"description"` dispatches to `EnrichDescription(TargetPath, NewValue)` (line 466), `"triggers"` to `InjectTrigger(TargetPath, NewValue)` (line 472), and any other value hard-errors with `applier: unsupported fieldKey %q` (line 476). It answers *"apply exactly this edit to exactly this file"* and needs a caller that already decided both.

A discovery report cannot be mechanically converted into an edit instruction, because deciding *which file to edit and what to write* is the authoring judgment the report exists to prompt. The missing fields in §A.3.2 are not an oversight in the schema — they are the output of a step that has no automated implementation.

At some point `moai harness execute --id <ID>` was wired to read `proposalgen`'s drafts (`internal/cli/harness/execute.go`, `resolveProposalPath` → `loadProposalByID` → `Applier.Apply`). That wiring joined two machines that were never the same pipeline, and it is why each predecessor could deliver a correct component and still leave the loop open:

| Predecessor | Delivered | Why the loop stayed open |
|---|---|---|
| SPEC-HARNESS-LOOP-CLOSURE-001 | lineage logging | logs a transition that never occurs |
| SPEC-HARNESS-OUTCOME-CAPTURE-001 | `apply_outcome` field | a field for an outcome nothing produces |
| SPEC-HARNESS-APPLY-EXECUTE-001 | a production `Apply()` caller | a caller wired to the wrong producer |
| SPEC-HARNESS-EVO-PIPE-REPAIR-001 | mapper vocabulary repair | repaired candidate→draft, not draft→apply |

Each verified its own component in isolation. None carried a criterion that a generated draft reaches a consumer that can act on it — so the seam between the two machines was never tested, and its absence never surfaced as a failure.

The process-level counterpart is unchanged and still holds: no predecessor carried an end-to-end reachability criterion, and the MoAI constitution's harness-edit obligation to record a falsifiable `prediction:` with a later `verified: true|false` (§ Lessons Protocol, Harness Edit Discipline) has **never been used** (§A.2). The verification obligation existed and was never exercised.

### A.5 Signal-quality gap (independent of A.3)

Even with the seam repaired, the queue content is low-value. 94.5% of observations are `agent_invocation` records naming a tool (`Bash`, `Read`, `Write`, MCP tool ids). Generated drafts inherit that vocabulary — e.g. `Draft proposal — agent_invocation:mcp__chrome-devtools__evaluate_script:`. Raw tool-call frequency carries no decision to learn from. Meanwhile the artifact designed to carry decisions — the routing-ledger — was never written, and the genuinely valuable lessons accumulate in a channel (`feedback_*.md`) that the learning loop never reads.

Measured composition of the draft queue (52 draft directories, each carrying both `proposal.json` and `spec.md`; measured 2026-07-27 against the primary checkout, where the gitignored runtime directory lives):

| Event type | Drafts | Unique `pattern_key` |
|---|---:|---:|
| `agent_invocation` | 29 (56%) | 28 |
| `tool_failure` | 19 (37%) | **13** |
| `user_prompt` | 2 | 2 |
| `subagent_stop` | 1 | 1 |
| `session_stop` | 1 | 1 |
| **total** | **52** | **45** |

The 13 unique `tool_failure` patterns are the genuinely actionable subset (`Bash:TimeoutError`, `Bash:SandboxViolation`, `Bash:OOMKilled`, `Write:PermissionDenied`, `Read:ContextCancelled`, `Agent:UnknownFailure`, and 7 more): each names a recurring failure mode with a subject and an error class. The 29 `agent_invocation` drafts are bare tool names and carry no decision — exactly the low-value class this section predicts.

### A.6 Draft identity is date-scoped, so one pattern yields repeat drafts

`buildDraftID` (`internal/harness/proposalgen/mapper.go:130-135`) derives the draft identifier as `PROPOSAL-<YYYYMMDD>-<sha256(pattern_key)[:8]>`, taking the date from the originating promotion's timestamp. The same `pattern_key` promoted on a different day therefore produces a **different** draft ID and a second directory.

Measured on the same 52-draft queue: 52 drafts collapse to 45 unique `pattern_key` values, and the 7-draft difference is exactly 7 duplicate pairs, each sharing a hash suffix and differing only in the date prefix:

| `pattern_key` | Draft IDs |
|---|---|
| `tool_failure:Bash:TimeoutError` | `PROPOSAL-20260713-bdba6c03`, `PROPOSAL-20260714-bdba6c03` |
| `tool_failure:Bash:SandboxViolation` | `PROPOSAL-20260713-59fa9c72`, `PROPOSAL-20260714-59fa9c72` |
| `tool_failure:Bash:OOMKilled` | `PROPOSAL-20260713-5ddb5047`, `PROPOSAL-20260714-5ddb5047` |
| `tool_failure:Write:PermissionDenied` | `PROPOSAL-20260713-0bd239a8`, `PROPOSAL-20260714-0bd239a8` |
| `tool_failure:Read:ContextCancelled` | `PROPOSAL-20260713-d4347ddd`, `PROPOSAL-20260714-d4347ddd` |
| `tool_failure:Agent:UnknownFailure` | `PROPOSAL-20260713-bab10eb9`, `PROPOSAL-20260714-bab10eb9` |
| `agent_invocation:mcp__chrome-devtools__navigate_page:` | `PROPOSAL-20260714-dbc0e2ec`, `PROPOSAL-20260720-dbc0e2ec` |

The scaffolder documents itself as byte-idempotent on re-run (`scaffolder.go:26-29`), and it is — but only **within a single date**. Across dates the identity changes and the idempotence guarantee does not hold. Left unaddressed, the queue accumulates one duplicate per pattern per active day, and duplicates are indistinguishable from genuine new findings at the point of review.

### A.7 The safety pipeline approves a content-free proposal

This subsection records a hazard that is **latent today and becomes live the moment the payload is made parseable**. It is stated here because the obvious next step after §A.3.2 — making the producer's JSON decode into `harness.Proposal` — would trigger it.

Today, `moai harness execute --id <ID>` against a producer-written draft fails early: `loadProposalByID` (`internal/cli/harness/execute.go:211-227`) fails at `json.Unmarshal` (line 223) because `proposal.json` carries `"tier": "auto_update"` (a string) while `harness.Proposal.Tier` is the numeric `harness.Tier` (`internal/harness/types.go:228,359`). The verb exits 1 before `Apply` is ever called, so nothing is written.

If only that decode were repaired — leaving `target_path` / `field_key` / `new_value` absent, as §A.3.2 establishes they must be — the resulting content-free proposal passes **all five safety layers**:

| Layer | Behaviour on an empty `TargetPath` | Evidence |
|---|---|---|
| L1 Frozen Guard | returns `false` — empty path is explicitly not frozen | `internal/harness/safety/frozen_guard.go:42-44` |
| L2 Canary | `defaultProjectedScorer` treats an empty target as a "meaningless change" and returns the baseline unchanged → delta 0 → below the −0.10 rejection threshold | `internal/harness/safety/canary.go:56-59` |
| L3 Contradiction | `FindFrozenRule("")` returns nil → empty report → no contradiction | `internal/harness/safety/frozen_rules.go:66-68`, `internal/harness/safety/contradiction.go:169-172` |
| L4 Rate Limiter | content-independent | `internal/harness/safety/pipeline.go:140-153` |
| L5 Oversight | auto-approved: `RunExecute` constructs the pipeline with `AutoApply=true` | `internal/cli/harness/execute.go:102-152` |

`Apply` therefore reaches Step 2 and calls `createSnapshot` (`internal/harness/applier.go:361`, defined at 642), which creates the dated snapshot directory (`os.MkdirAll`, line 648) **before** attempting to read the target file (`os.ReadFile(proposal.TargetPath)`, line 653). With an empty path the read fails and `Apply` aborts at line 363 — leaving an empty snapshot directory behind on every attempt.

Two corrections to the intuitive reading follow, and both matter for M2:

- The `unsupported fieldKey` hard-error (`applier.go:476`) is **not** the failure point for a content-free proposal. `createSnapshot` fails first, so that branch is unreachable and cannot be relied on as the guard.
- The real defect is that the 5-layer pipeline does not screen for **applicability** at all. It checks whether an edit is *safe*, never whether an edit was actually *specified*. A guard that rejects a non-applicable proposal must therefore run **before** `createSnapshot`, not inside the apply switch.

---

## §B Scope

### B.1 In scope

- The proposal layout contract between `proposalgen` and its three consumers (C1-C3). **(M1 — complete)**
- The routing of `proposalgen` drafts to their designed consumer (manager-spec SPEC authoring), and the retirement of the `execute`→draft wiring that joined two different machines (§A.4).
- Draft identity: one `pattern_key` yields one draft, not one per active date (§A.6).
- The disposition of the `tier` string/numeric split (§A.7) — dissolved by de-wiring rather than repaired, with the fallback recorded.
- The routing-ledger recording obligation at dispatch time.
- The lesson-channel split between the constitution-designated `lessons.md` and the practiced `feedback_*.md`.
- The `lessons-inbox.jsonl` drain ownership.
- The `prediction:` / `verified:` falsifiability fields for harness edits.
- Two `moai harness` CLI reporting defects (help-text verb omission; `list` vs `doctor` disagreement on thin harnesses).

### B.2 Exclusions

### Out of Scope — adjacent surfaces

- The `goal` surface. The audit found it fully wired and template-mirrored (`moai goal` CLI, `moai hook stop-goal`, `handle-stop-goal.sh` registered on `Stop`, `goal.md.tmpl`, `handle-stop-goal.sh.tmpl`, `settings.json.tmpl` registration). Its only gap is adoption (`.moai/state/goal/` empty since creation), which is a usage question, not a defect. SPEC-GOAL-DOCS-RETIRE-001 is concurrently active on this surface; this SPEC does not touch it.
- Redesigning the observation event schema wholesale. M4 narrows what is *promoted*, not what is *recorded*.
- The 5-layer safety pipeline itself. §A.7's finding is that it does not check *applicability*; the remedy is a guard ahead of it (REQ-HLR-004c), not a change inside it.

### Out of Scope — existing runtime data

- Retroactive rewriting of the 52 existing drafts. This includes the 7 duplicate pairs in §A.6: the identity fix is **forward-looking only**, and the existing duplicates are grandfathered.

### Out of Scope — the Applier payload gap

- Building a producer for the `Applier` frontmatter-enrichment path. That path is preserved and left unfed; whether it ever gets a producer is §G open question 4.
- Adding the missing `target_path` / `field_key` / `new_value` fields to the observation schema so that drafts could become edit instructions. §A.4 establishes these are the output of an authoring judgment, not of an observation gap.
- Adding a `harness.Tier` JSON codec. Its only consumer is the path this SPEC removes; the split is dissolved, not repaired (REQ-HLR-012).

---

## §C Requirements (GEARS)

### REQ-HLR-001 — Proposal layout contract (ubiquitous)

The system SHALL define one on-disk layout for a generated proposal, and every producer and consumer SHALL resolve proposals through a single shared accessor rather than re-deriving the path.

### REQ-HLR-002 — Pending discovery (event-driven)

WHERE at least one draft proposal exists on disk, WHEN `moai harness status` runs, the system SHALL report a pending count equal to the number of draft proposals.

### REQ-HLR-003 — Apply reachability (event-driven)

WHERE a draft proposal exists, WHEN `moai harness apply` runs, the system SHALL return that proposal's payload rather than "No pending proposals".

### REQ-HLR-004 — Draft delivery to its designed consumer (event-driven)

> Amended in v0.2.0. The prior formulation required `moai harness execute --id <ID>` to load a draft and emit one `apply_outcome` record. §A.3.2 and §A.4 establish that a `proposalgen` draft is a discovery report and can never satisfy `Applier.Apply()`, which consumes a frontmatter edit instruction. Requiring an `apply_outcome` from this path required an impossible conversion.

WHERE a draft proposal with identifier `<ID>` exists, WHEN the promotion path for `<ID>` runs, the system SHALL create a SPEC directory under `.moai/specs/`, SHALL record `<ID>` in that SPEC as its provenance, and SHALL stop offering `<ID>` as a pending draft.

### REQ-HLR-004b — Applier path not fed by proposalgen (ubiquitous)

The system SHALL NOT route `proposalgen` drafts into `Applier.Apply()`. The frontmatter-enrichment Applier path SHALL be preserved and SHALL remain unfed until a producer of fully-populated `harness.Proposal` values exists.

### REQ-HLR-004c — Applicability guard precedes snapshot (event-driven)

WHERE a proposal lacking `target_path`, `field_key`, or `new_value` is submitted to the apply path, WHEN that path runs, the system SHALL reject it with a diagnostic naming the missing field, and SHALL do so before any snapshot directory is created.

### REQ-HLR-005 — Dispatch observation (state-driven)

WHILE the routing observation opt-in is enabled, the orchestrator SHALL record each `/moai` dispatch to the routing-ledger before executing the routed workflow.

### REQ-HLR-006 — Single lesson channel (ubiquitous)

The system SHALL designate exactly one lesson store, and the constitution's Lessons Protocol SHALL name that store. Divergence between the designated and the practiced store SHALL be resolved in favour of the practiced store.

### REQ-HLR-007 — Inbox drain ownership (state-driven)

WHILE `.moai/lessons-inbox.jsonl` holds undrained entries, the system SHALL name the actor and the trigger that drains them.

### REQ-HLR-008 — Harness-edit falsifiability (event-driven)

WHEN a lesson motivates a harness edit, the lesson entry SHALL record a falsifiable `prediction:` and SHALL later record `verified: true|false` with observed evidence.

### REQ-HLR-009 — Promotion routing by enforceability (ubiquitous)

The system SHALL route a promotion candidate by whether a script can mechanically detect the condition. WHERE it can, the promotion target SHALL be a hook; WHERE it cannot, the target SHALL be a rule or the lesson store.

### REQ-HLR-010 — CLI reporting accuracy (ubiquitous)

The `moai harness` help text SHALL enumerate every shipped verb, and `list` SHALL NOT render a state that `doctor` classifies as expected in defect-suggesting language.

### REQ-HLR-011 — Draft identity is pattern-scoped (ubiquitous)

The system SHALL derive a draft's identity from its `pattern_key` alone. WHERE the same `pattern_key` is promoted on more than one date, the generator SHALL converge on a single draft rather than emitting one draft per date.

### REQ-HLR-012 — Tier-split disposition is recorded (ubiquitous)

The system SHALL record the disposition of the producer/consumer `tier` representation split (string on disk, numeric in `harness.Proposal`). WHERE the `execute`→draft wiring is removed, the split SHALL be recorded as dissolved — no `harness.Tier` JSON codec is added — and the characterization test pinning the mismatch SHALL be retired with its rationale stated rather than deleted silently.

---

## §D Milestones

| M | Scope | Source | Verification | Status |
|---|---|---|---|---|
| **M1** | Shared proposal accessor; repair C1/C2/C3 | §A.3.1 | `status` pending == on-disk draft count | **complete** — `c996eb294` |
| **M2** | Route drafts to their designed consumer: promotion path draft → SPEC; de-wire `execute`→draft; applicability guard before snapshot; retire the tier tripwire | §A.3.2, §A.4, §A.7 | a named draft becomes a SPEC carrying its provenance, and leaves the pending queue | not started |
| **M3** | Routing-ledger recording obligation at dispatch | §A.5 | ledger row count increases per dispatch | not started |
| **M4** | Generator quality: promotion routing by enforceability, narrow `agent_invocation` promotion, pattern-scoped draft identity | §A.5, §A.6 | no new draft with a bare-tool-name `pattern_key`; a two-date fixture yields one draft | not started |
| **M5** | Lesson-channel unification + inbox drain ownership | §A.2 | designated store == practiced store; inbox drains | not started |
| **M6** | `prediction:`/`verified:` on harness-edit lessons; CLI reporting fixes | §A.2, §B.1 | fields present on new entries; help lists all verbs | not started |

Sequencing: M1 gates M2 (nothing downstream was observable until the seam was repaired). M2 is the only milestone carrying a reversible architectural decision and is therefore sequenced first among the remainder. M3-M6 are independent of each other and of M2.

M4 depends on M2 only in ordering preference, not correctness: narrowing what gets promoted is more useful once the promotion destination exists, but neither blocks the other.

---

## §E Acceptance Criteria

> **`acceptance.md` is the SSOT for acceptance criteria.** As of v0.2.0 the full Given / When / Then bodies, falsification conditions, and evidence live in `.moai/specs/SPEC-HARNESS-LOOP-REPAIR-001/acceptance.md`. This section is an index only — it exists so a reader of `spec.md` can see the criterion set without opening a second file, and it deliberately carries no detail that could drift from the SSOT.

Every criterion is stated so that **reverting the corresponding change makes it fail**. Criteria that only assert token presence in a file are explicitly rejected for this SPEC — the predecessor recurrence in §A.4 is attributed to exactly that weakness.

| AC | M | Intent | Status |
|---|---|---|---|
| AC-HLR-001 | M1 | `status` pending count equals the on-disk draft count | **PASS** |
| AC-HLR-002 | M1 | `apply` returns a draft payload, not "No pending proposals" | **PASS** |
| AC-HLR-003 | M1 | `execute --id` resolves to the draft rather than "proposal not found" | **PASS** |
| AC-HLR-004 | M2 | a named draft becomes a SPEC carrying its provenance and leaves the pending queue | open — **rewritten in v0.2.0** |
| AC-HLR-005 | M2 | each promotion leaves one auditable record linking draft → SPEC | open — **rewritten in v0.2.0** |
| AC-HLR-006 | M1 | one shared accessor; no call site re-derives `id + ".json"` | **PASS** |
| AC-HLR-007 | M3 | a `/moai` dispatch appends one routing-ledger row | open |
| AC-HLR-008 | M4 | no newly generated draft carries a bare-tool-name `pattern_key` | open |
| AC-HLR-009 | M5 | exactly one designated lesson store, and it is the practiced one | open |
| AC-HLR-010 | M5 | the inbox drain actor and trigger are named, and a drain reduces the backlog | open |
| AC-HLR-011 | M6 | harness-edit lessons carry `prediction:` then `verified:` | open |
| AC-HLR-012 | M6 | `harness --help` enumerates every shipped verb | open |
| AC-HLR-013 | M6 | `list` and `doctor` agree on a command-only thin harness | open |
| AC-HLR-014 | M2 | `execute` no longer accepts a `proposalgen` draft, and says why | open — new in v0.2.0 |
| AC-HLR-015 | M2 | a non-applicable proposal is rejected before any snapshot directory is created | open — new in v0.2.0 |
| AC-HLR-016 | M2 | the tier-split disposition is recorded and the tripwire retired with rationale | open — new in v0.2.0 |
| AC-HLR-017 | M4 | one `pattern_key` yields one draft across dates | open — new in v0.2.0 |

Retired in v0.2.0: the former AC-HLR-004 (`grep -c apply_outcome … ≥ 1`) and AC-HLR-005 (`learning-history/applied/` materialises). Both presumed `proposalgen` drafts reach `Applier.Apply()`; §A.4 establishes they cannot. Their replacements carry the same falsifiability discipline against the corrected consumer. `apply_outcome` telemetry and the `applied/` directory remain the success signals of the **Applier** path, which this SPEC deliberately leaves unfed (§G open question 4).

---

## §F Risks

| Risk | Mitigation |
|---|---|
| Repairing the seam surfaces 52 low-value drafts at once | M4 narrows promotion before any bulk promotion; existing drafts stay unpromoted and are not retroactively rewritten |
| ~~First real `Applier.Apply()` execution mutates harness files~~ | **Withdrawn in v0.2.0.** M2 no longer drives `Applier.Apply()` at all (§A.4). The residual apply-path risk is now the §A.7 hazard, below |
| Making the producer payload parseable would let a content-free proposal pass all five safety layers and leave empty snapshot directories | §A.7 records the full trace; REQ-HLR-004c requires the applicability guard to run **before** `createSnapshot`. A tier codec MUST NOT be added ahead of that guard |
| Promotion creates SPEC directories from low-value drafts, polluting `.moai/specs/` | Promotion is explicit and per-draft (never a bulk sweep); M4's narrowing reduces the candidate pool; the §A.5 measurement identifies the 13 `tool_failure` patterns as the actionable subset |
| Retiring the `execute`→draft wiring leaves `moai harness execute` with no valid input | AC-HLR-014 requires the verb to fail with a diagnostic naming the absent producer, rather than a confusing unmarshal error. Whether the verb is removed outright is §G open question 4 |
| Concurrent session owns the goal surface | Goal is out of scope (§B.2); this SPEC's worktree branches from `origin/main` |
| ~~Choosing the flat layout instead of the nested one would orphan 52 drafts~~ | **Resolved in M1.** The nested producer layout was adopted; consumers were normalised to it. See §G question 1 |

---

## §G Open Questions

1. ~~**Layout direction**~~ — **RESOLVED (M1).** Consumers were normalised to the nested producer layout. The nested form held all 52 live drafts and carries `spec.md` alongside `proposal.json`; changing the producer to flat would have orphaned them. Recorded in `progress.md` § Decision taken this session.
2. **Promotion narrowing rule** — which observation subjects remain promotable once bare tool names are excluded. §A.5 measures the candidate pool (13 unique `tool_failure` patterns vs 28 `agent_invocation`), but does not settle the rule. Belongs to M4.
3. **Lesson store direction** — migrate `lessons.md` content into the topic-file convention, or restore `lessons.md` as an index over it. Belongs to M5.
4. **Does the `Applier` frontmatter-enrichment path ever get a producer?** — The path is fully built (5-layer pipeline, snapshot/rollback, regression gate, lineage, outcome telemetry) and has never had a caller supplying a populated `harness.Proposal` (§A.3.2). Three dispositions are open, and this SPEC picks none of them:
   - **(a) Leave dormant** — preserve the code unfed; accept that `apply_outcome` telemetry stays at zero. Cheapest; leaves a large unexercised subsystem in the tree.
   - **(b) Give it a producer** — a distinct upstream that decides target file + field + value. This is the authoring judgment §A.4 identifies as unautomated; a producer would need to make it.
   - **(c) Retire it** — remove the path and its telemetry. Reverses four predecessor SPECs' deliverables and should not be done inside this SPEC.
   The choice determines whether `moai harness execute` survives as a verb (see AC-HLR-014). Deferring is deliberate: this SPEC's mandate is to stop the mis-wiring, not to decide the Applier's future.
5. **Where does the applicability guard live?** — REQ-HLR-004c requires rejection before `createSnapshot`. Two placements are viable: at load time (reject a proposal that decodes but carries no edit), or as a pre-flight inside `Apply` ahead of Step 2. The load-time placement keeps `Apply` unchanged; the `Apply` placement protects every future caller including ones that bypass the loader. Belongs to M2 design.
