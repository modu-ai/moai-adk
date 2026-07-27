---
id: SPEC-HARNESS-LOOP-REPAIR-001
title: "Harness self-learning loop repair — proposal layout contract + decision-signal observation + lesson-channel unification"
version: "0.1.0"
status: draft
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

### A.3 Root cause — producer/consumer layout contract mismatch

The proposal **producer** writes a directory per draft (`internal/harness/proposalgen/scaffolder.go:94,101,106`):

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

This single contract break explains every downstream symptom in §A.2 — zero applies, absent `applied/`, zero `apply_outcome` telemetry — and explains why four component-correct predecessors did not close the loop: **each built a correct part; the parts cannot read each other's data.**

### A.4 Why it survived four SPECs

No predecessor carried an end-to-end acceptance criterion that a generated draft becomes a visible pending proposal and then an applied outcome. Their criteria verified component existence (a caller exists, a field exists, a gate compiles), not reachability across the producer/consumer seam. The MoAI constitution already requires harness edits to record a falsifiable `prediction:` and a later `verified: true|false` (§ Lessons Protocol, Harness Edit Discipline); §A.2 shows that field pair has **never been used**. The verification obligation existed and was never exercised — that is the process-level root cause of the four-SPEC recurrence.

### A.5 Signal-quality gap (independent of A.3)

Even with the seam repaired, the queue content is low-value. 94.5% of observations are `agent_invocation` records naming a tool (`Bash`, `Read`, `Write`, MCP tool ids). Generated drafts inherit that vocabulary — e.g. `Draft proposal — agent_invocation:mcp__chrome-devtools__evaluate_script:`. Raw tool-call frequency carries no decision to learn from. Meanwhile the artifact designed to carry decisions — the routing-ledger — was never written, and the genuinely valuable lessons accumulate in a channel (`feedback_*.md`) that the learning loop never reads.

---

## §B Scope

### B.1 In scope

- The proposal layout contract between `proposalgen` and its three consumers (C1-C3).
- The routing-ledger recording obligation at dispatch time.
- The lesson-channel split between the constitution-designated `lessons.md` and the practiced `feedback_*.md`.
- The `lessons-inbox.jsonl` drain ownership.
- The `prediction:` / `verified:` falsifiability fields for harness edits.
- Two `moai harness` CLI reporting defects (help-text verb omission; `list` vs `doctor` disagreement on thin harnesses).

### B.2 Out of scope

- The `goal` surface. The audit found it fully wired and template-mirrored (`moai goal` CLI, `moai hook stop-goal`, `handle-stop-goal.sh` registered on `Stop`, `goal.md.tmpl`, `handle-stop-goal.sh.tmpl`, `settings.json.tmpl` registration). Its only gap is adoption (`.moai/state/goal/` empty since creation), which is a usage question, not a defect. SPEC-GOAL-DOCS-RETIRE-001 is concurrently active on this surface; this SPEC does not touch it.
- Redesigning the observation event schema wholesale. M4 narrows what is *promoted*, not what is *recorded*.
- Retroactive rewriting of the 52 existing drafts.

---

## §C Requirements (GEARS)

### REQ-HLR-001 — Proposal layout contract (ubiquitous)

The system SHALL define one on-disk layout for a generated proposal, and every producer and consumer SHALL resolve proposals through a single shared accessor rather than re-deriving the path.

### REQ-HLR-002 — Pending discovery (event-driven)

WHERE at least one draft proposal exists on disk, WHEN `moai harness status` runs, the system SHALL report a pending count equal to the number of draft proposals.

### REQ-HLR-003 — Apply reachability (event-driven)

WHERE a draft proposal exists, WHEN `moai harness apply` runs, the system SHALL return that proposal's payload rather than "No pending proposals".

### REQ-HLR-004 — Execute reachability (event-driven)

WHERE a draft proposal with identifier `<ID>` exists, WHEN `moai harness execute --id <ID>` runs, the system SHALL load that proposal and SHALL emit one `apply_outcome` record to `usage-log.jsonl`.

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

---

## §D Milestones

| M | Scope | Audit finding | Verification |
|---|---|---|---|
| **M1** | Shared proposal accessor; repair C1/C2/C3 | L1 | `status` pending == on-disk draft count |
| **M2** | End-to-end reachability: generate → status → apply → execute → `apply_outcome` | L1 | one `apply_outcome` line appears |
| **M3** | Routing-ledger recording obligation at dispatch | L3 | ledger row count increases per dispatch |
| **M4** | Promotion routing by enforceability; narrow `agent_invocation` promotion | L2 | no new draft whose `pattern_key` is a bare tool name |
| **M5** | Lesson-channel unification + inbox drain ownership | L4, L5 | designated store == practiced store; inbox drains |
| **M6** | `prediction:`/`verified:` on harness-edit lessons; CLI reporting fixes | L6, H1, H2 | fields present on new entries; help lists all verbs |

Sequencing: M1 → M2 gate the rest (nothing downstream is observable until the seam is repaired). M3-M6 are independent of each other.

---

## §E Acceptance Criteria

Each criterion below is stated so that reverting the corresponding change makes it fail. Criteria that only assert token presence in a file are explicitly rejected for this SPEC — the predecessor recurrence in §A.4 is attributed to exactly that weakness.

### AC-HLR-001 (M1) — pending count matches disk
- **Given** N draft proposal directories under `.moai/harness/proposals/`
- **When** `moai harness status` runs
- **Then** it reports `pending proposals: N items` where N > 0
- **Falsification** reverting the accessor restores `0 items` with the same fixture

### AC-HLR-002 (M1) — apply returns a payload
- **Given** at least one draft proposal
- **When** `moai harness apply` runs
- **Then** stdout contains that proposal's identifier, not "No pending proposals"
- **Falsification** reverting restores the "No pending proposals" branch

### AC-HLR-003 (M1) — execute resolves by ID
- **Given** a draft proposal with identifier `<ID>`
- **When** `moai harness execute --id <ID>` runs
- **Then** it does not fail with "proposal not found"
- **Falsification** reverting restores `proposal not found`

### AC-HLR-004 (M2) — first apply outcome
- **Given** the repaired seam and one draft proposal
- **When** the execute path completes
- **Then** `grep -c apply_outcome .moai/harness/usage-log.jsonl` returns ≥ 1 (baseline: 0)
- **Note** this is the criterion no predecessor carried

### AC-HLR-005 (M2) — applied history materialises
- **Given** one completed apply
- **Then** `.moai/harness/learning-history/applied/` exists and holds one record (baseline: directory absent)

### AC-HLR-006 (M1) — single accessor
- **Given** the repaired code
- **Then** C1, C2, C3 resolve proposals through one shared function; no call site re-derives `id + ".json"`
- **Falsification** a regression test fails if any site reintroduces an independent path derivation

### AC-HLR-007 (M3) — dispatch recorded
- **Given** the routing observation opt-in enabled
- **When** a `/moai` subcommand dispatches
- **Then** the routing-ledger line count increases by one

### AC-HLR-008 (M4) — promotion excludes bare tool names
- **Given** the narrowed promotion rule
- **When** the generator runs against the existing usage log
- **Then** no newly generated draft carries a `pattern_key` whose subject is a bare tool name
- **Falsification** reverting reproduces an `agent_invocation:<Tool>` draft

### AC-HLR-009 (M5) — one designated lesson store
- **Given** the reconciled constitution
- **Then** the Lessons Protocol names the practiced store, and no rule names a store that has not been written to in 30 days

### AC-HLR-010 (M5) — inbox drain named
- **Given** the reconciled doctrine
- **Then** the drain actor and trigger are named, and a drain run reduces the undrained entry count

### AC-HLR-011 (M6) — falsifiability recorded
- **Given** a harness edit made under this SPEC
- **Then** its lesson entry carries `prediction:` at edit time and `verified:` after observation

### AC-HLR-012 (M6) — help enumerates every verb
- **Given** `moai harness --help`
- **Then** its description lists every verb present in the command table (baseline: 6 omitted — `clusters`, `propose`, `install`, `execute`, `doctor`, `ledger`)

### AC-HLR-013 (M6) — list and doctor agree
- **Given** a command-only thin harness
- **Then** `list` does not describe it in defect-suggesting terms while `doctor` classifies it as expected

---

## §F Risks

| Risk | Mitigation |
|---|---|
| Repairing the seam surfaces 52 low-value drafts at once | M4 narrows promotion before any bulk apply; existing drafts stay unapplied and are not retroactively rewritten |
| First real `Applier.Apply()` execution mutates harness files | The 5-layer safety pipeline and snapshot/rollback path already exist; M2 exercises exactly one apply behind the human gate |
| Concurrent session owns the goal surface | Goal is out of scope (§B.2); this SPEC's worktree branches from `origin/main` |
| Choosing the flat layout instead of the nested one would orphan 52 drafts | Layout decision is deferred to plan-phase; the nested form is the one with live data |

---

## §G Open Questions

1. **Layout direction** — normalise consumers to the nested producer layout, or change the producer to flat? The nested form holds 52 live drafts and carries `spec.md` alongside `proposal.json`; the flat form is assumed by three consumers. Resolution belongs in plan.md.
2. **Promotion narrowing rule** — which observation subjects remain promotable once bare tool names are excluded.
3. **Lesson store direction** — migrate `lessons.md` content into the topic-file convention, or restore `lessons.md` as an index over it.
