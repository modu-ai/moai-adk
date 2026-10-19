---
id: SPEC-LANE-READ-COMPACT-001
title: "Lane read verbs, part (a) — todo compact form and factory status default bound"
version: "0.3.0"
status: draft
created: 2026-10-15
updated: 2026-10-19
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/cli"
lifecycle: spec-anchored
tags: "lane, read-verbs, todo, factory, output-bound, quiet-form, doctrine"
tier: M
related_specs: [SPEC-LANE-READ-VERBS-001, SPEC-LANE-READ-SESSION-001, SPEC-TODO-SURFACE-POLISH-001, SPEC-TODO-CLAIM-LEASE-001, SPEC-TODO-TRANSITION-STAMPS-001, SPEC-TODO-CLASSIFY-DISPATCH-001, SPEC-QUOTA-AWARE-SCHEDULING-001]
---

# SPEC-LANE-READ-COMPACT-001 — Lane Read Verbs (a): Todo Compact Form and Factory Status Default

## A. History

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-10-15 | manager-spec | Split from `SPEC-LANE-READ-VERBS-001` (card t1777) by board ruling `d-20261015T122748Z-6d6b`: this SPEC is part (a), card t1810, the low-risk half. It carries the todo compact form (Q1, Q2), the factory status default (Q3, Q8), their lane-facing consumer reads, and the queue-and-factory doctrine sentence. It carries no session list and no Windows content; those are `SPEC-LANE-READ-SESSION-001` (card t1811). Requirements and criteria are renumbered from 001; the parent ids are listed in plan § A. 10 requirements and 11 criteria; Tier M. |
| 0.2.0 | 2026-10-19 | manager-spec | Card t1810 plan-phase premise refresh (no requirement, criterion, decision, or bound changed): the carried size premises were re-measured on tree 74ad197380 (todo default 2,897,435 bytes, about +10%; factory status 201,400 bytes · 368 cards · 15 run ids, about +72%; `--compact` and `--full` still absent), recorded as ledger rows L-14 to L-16, and the plan's line references were re-anchored by content to tree a3e8d1409c after four days of main drift — the `gtd.md` full-form lines moved to 218/261/262/475 and card t1827 added a fifth naming line at `:47`, which joins the compact change; the `factory_card.go` and `todo.go` anchors moved by roughly 100 lines. |
| 0.3.0 | 2026-10-19 | manager-spec | Plan-audit iter1 repairs (verdict FAIL 0.85, receipt `rcpt-9aaec51f93f1b122971c9d8a`): REQ-LRC-003 and REQ-LRC-006 scope their 2048-byte bounds to content produced by the CLI's default content paths, citing the Q3 anchor (board record `d-20261015T113856Z-6f23`), because the store enforces no id-width invariant (`cardIDPattern` carries no length cap; the run identifier is operator-settable through `MOAI_KANBAN_ID` with no `RecordRun` length validation) — operator-set exotic widths are named out of the bounds' premise, and AC-LRC-003 and AC-LRC-006 gain long-input cells carrying the iter1 overlay observations (3,039 bytes with six 202-byte identifiers; 2,571 bytes with a 600-character run identifier) so neither criterion reads as "the bound holds for all store content". Plan § D.2's four `factory_card.go` anchors are re-pinned to the post-absorb sites (:3335, :3265-3266, :3294, :3396-3397) after merge 5892d48f8d moved the file, and plan § B gains the ledger-row citation clause (the board records' "실측 L-15/L-16" figures are the parent ledger's rows and live here as L-12 and L-13). No cap, bound value, decision, or requirement scope changed otherwise. |

## B. Background

Card t1777 reported that a Codex lane reached 72% of its context before its card was dispatched, because the read verbs a lane calls emit their full records by default. This SPEC bounds two of those verbs: `moai todo list` gains an opt-in `--compact` JSON form, and the default `moai factory status --json` becomes a bounded, active-run-only listing. The measurements are carried from the parent and recorded in `acceptance.md` § D; they were re-measured on 2026-10-19 at tree 74ad197380 as ledger rows L-14 to L-16, and the problem definition they carry strengthened in the four days between.

The decisions it rests on are listed once here: REQ-LRC-001 to REQ-LRC-004 rest on Q1, and REQ-LRC-002 and REQ-LRC-003 also on Q2; REQ-LRC-005 to REQ-LRC-007 rest on Q3, and REQ-LRC-007 also on Q8; REQ-LRC-008 rests on Q7, and Q6 decides at run phase whether its placement holds; the todo and factory half of the scope change Q9 covers REQ-LRC-001, REQ-LRC-003, and REQ-LRC-006. Q1 is DECIDED on board record dfd8, Q2, Q3, Q7, Q8, and Q9 on board record 6f23 (the anchor is the board id under the operator-proxy standing), and Q6 is EVIDENCE-NEEDED. REQ-LRC-009 and REQ-LRC-010 rest on no row of their own.

## C. Requirements (GEARS)

### Todo queue

- REQ-LRC-001 (Ubiquitous): The `moai todo list --json` command shall emit its default output byte-identical to the pre-change full record for the same store snapshot, with the same top-level keys and the same `items`, `findings`, `archived`, and `runtime` fields. This is the frozen contract of REQ-TCL-014 and AC-TST-012.
- REQ-LRC-002 (Event-driven): **When** `moai todo list --json --compact` is invoked, the todo list command shall emit one JSON object with the keys `project_uuid`, `version`, `last_seq`, `live`, `withheld`, `items`, `findings_count`, `archived_count`, and `dropped_count`, where `live` counts the items whose state is not `dropped`, `items` holds the first at most six live items in queue order as objects with `id`, `state`, and `prefix`, and `withheld` equals `live` minus the number of rendered items; `--limit` shall be ignored.
- REQ-LRC-003 (Ubiquitous): The `moai todo list --json --compact` output shall not exceed 2048 bytes for queue content produced by the CLI's default content paths — records whose identifiers are of the width the CLI's own id generator emits. The bound holds because the row cap is six and each prefix is at most 40 runes followed by `...`, and each rune's JSON encoding is at most 6 bytes. Store states written outside those default paths — hand-written or fixture identifiers wider than the generator emits, which the store does not prevent (`cardIDPattern`, `internal/homestate/card_evidence_readers.go:19`, carries no length cap) — are outside the bound's premise, are documented by the long-input cell of AC-LRC-003, and receive no store validation from this SPEC. The scoping is the default-form scope Q3 (board record `d-20261015T113856Z-6f23`) decided for the factory default's cap, applied here to the todo compact form.
- REQ-LRC-004 (Ubiquitous): Every lane-facing read of the todo queue in the lane doctrine and in `.claude/skills/moai/workflows/gtd.md` and its template copy shall name the `--compact` form, and the standing-source duplicate check shall name the full form together with its reason, because a compact row carries only a 40-rune prefix and not the card text the check compares.

### Factory status

- REQ-LRC-005 (Event-driven): **When** `moai factory status --json` is invoked without `--run` and without `--full`, the factory status command shall list only the cards of runs whose `runs.status` is `active`; shall render at most three cards, each with the fields `run_id`, `card_id`, `state`, `stage`, `mode`, `priority`, `lease_holder`, and `lease_expired`, where `lease_holder` is cut at a rune boundary to at most 16 bytes of JSON-encoded width; shall report the active cards not rendered in `withheld` and the count of unavailable records in `unavailable_count`; and, while quota pressure is on, shall carry in its `quota` key the pressure flag, both windows, the unknown-lane count, the no-candidate warning where it applies, and at most two candidate lanes, each label cut as `lease_holder` is, with the number of candidates not rendered in `candidates_withheld`. Keeping `mode` and `priority` keeps REQ-TCD-010 (SPEC-TODO-CLASSIFY-DISPATCH-001) on this path. The candidate cap departs from REQ-QAS-019 (SPEC-QUOTA-AWARE-SCHEDULING-001) on this path only: `--run <id>`, `--full`, and the text view keep the whole candidate list.
- REQ-LRC-006 (Ubiquitous): The `moai factory status --json` default output shall not exceed 2048 bytes in every quota state and for any number of cards, candidate lanes, or holder and label lengths, for records produced by the CLI's default content paths — the fields the default does not cut at the widths the CLI's own generators and records carry: run identifiers of 6 characters, card identifiers of at most 8, states and stages at their longest enumerated values (`needs-decision`, `merge-ready`), counts of at most 6 digits, duration strings of at most 12 characters, timestamps in the 20-character RFC 3339 form, and a used percentage no longer than the 25-byte longest float encoding. Store states with operator-set widths outside those default paths — a run identifier set through `MOAI_KANBAN_ID` beyond the default six characters (`internal/cli/factory.go:913`, key at `internal/config/envkeys.go:202`, no length validation in `RecordRun`, `internal/homestate/runtime.go:28`) or card identifiers wider than the generator emits (no length cap in `cardIDPattern`, `internal/homestate/card_evidence_readers.go:19`) — are outside the bound's premise, are documented by the long-input cell of AC-LRC-006, and receive no store validation from this SPEC. The scoping is Q3's own (board record `d-20261015T113856Z-6f23`): the record caps the CLI's 기본 출력, the default output, not every writable store state.
- REQ-LRC-007 (Event-driven): **When** `moai factory status --json --run <id>` is invoked, the factory status command shall keep its pre-change output for that run, with every card field and no row cap; **when** `moai factory status --json --full` is invoked, the factory status command shall emit the pre-change output for all runs, byte-identical for the same store snapshot.

### Lane doctrine

- REQ-LRC-008 (Ubiquitous): The factory dispatch rule shall state, in one sentence, that lanes read queue and factory state only through the quiet forms and that a full dump of either is an operator-requested exception. The sentence shall appear exactly once, inside a role-core region, in `.claude/rules/moai/workflow/factory-dispatch.md` and in its template copy `internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md`, the two copies from which the role injection extracts regions; the identical sentence shall appear exactly once in the shared-policy copy `.moai/policies/workflow/factory-dispatch.md`, which carries no role-core region. The local and template copies shall keep the one-line live-only difference that `TestAutoPickMirrorParity` requires.

### Compatibility and discoverability

- REQ-LRC-009 (Ubiquitous): The frozen contracts REQ-TCL-014 (`TestTodoListJSON_GoldenByteIdentity`), AC-TST-012 (`internal/cli/testdata/ac_tst_012_golden.json`), and AC-TSP-021 (`TestTodoListJSONIgnoresLimit`) shall remain unchanged in their asserted bytes and shall remain passing; `TestFactoryStatusReportsLeaseKeys` shall keep its asserted keys; and `TestTodoListJSONShapeMatchesDoc`, `TestSessionListDefaultNoProbe`, `TestAutoPickMirrorParity`, and `TestRuleTemplateMirrorDrift` shall remain green.
- REQ-LRC-010 (Ubiquitous): The help text of `moai todo list` and of `moai factory status` shall name the new flag that applies to each verb (`--compact` and `--full`, respectively) and shall state the default row cap.

## D. Acceptance Criteria (summary)

The Given-When-Then criteria, the RED-now cells, and the evidence ledger are in `acceptance.md`.

- AC-LRC-001 — todo default is unchanged in bytes and keys (REQ-LRC-001).
- AC-LRC-002 — todo `--compact` shape, live-only rows, prefix, and `withheld` arithmetic (REQ-LRC-002).
- AC-LRC-003 — todo `--compact` is at most 2048 bytes on the worst-case fixture (REQ-LRC-003).
- AC-LRC-004 — lane-facing reads name `--compact`; the duplicate check names the full form with the literal `40-rune prefix` (REQ-LRC-004).
- AC-LRC-005 — factory default is active-run only, eight fields, capped at three, with `withheld`, `unavailable_count`, and at most two quota candidates (REQ-LRC-005).
- AC-LRC-006 — factory default is at most 2048 bytes with quota pressure on (REQ-LRC-006).
- AC-LRC-007 — factory `--run` keeps its pre-change output, and `--full` is byte-identical to the pre-change output (REQ-LRC-007).
- AC-LRC-008 — the queue-and-factory doctrine sentence is present once in each of the three copies, inside the role-core region of the two `.claude/rules` copies (REQ-LRC-008).
- AC-LRC-009 — the frozen goldens, the contract tests, and the mirror guard stay green (REQ-LRC-009).
- AC-LRC-010 — the todo and factory help texts name their new flags and the default row cap (REQ-LRC-010).
- AC-LRC-011 — `moai spec lint SPEC-LANE-READ-COMPACT-001 --json` exits 0 (quality gate G1).

## E. Non-Functional Constraints

- Byte bounds: `moai todo list --json --compact` and the default `moai factory status --json` are each at most 2048 bytes. The default `moai todo list --json` has no byte bound: it is the full record.
- Determinism: for a fixed store snapshot, each output is byte-for-byte deterministic. The tests use fixtures, and never the live store.
- Exit contract: no existing exit status or sentinel string changes. The new flags add no refusal path beyond the existing unknown-flag refusal.
- Doctrine: one sentence. No new always-loaded file is created.
- Code comments follow the `code_comments` setting in `.moai/config/sections/language.yaml` (English by default).

## F. Out of Scope

The items below are out of scope for this SPEC. Each is named so that a later change cannot claim it silently.

### Out of Scope — session list and the Windows probe

- `moai session list`, its live filter, the PID liveness probe on any platform, and the REQ-SCV-006 reversal belong to `SPEC-LANE-READ-SESSION-001` (card t1811).

### Out of Scope — todo text view

- `moai todo list` without `--json` keeps its present 100-row text bound and its withheld line.

### Out of Scope — todo filters

- `moai todo list` gains no `--state` or `--fields` flag.

### Out of Scope — MCP surfaces

- The MCP `todo_list` path (`internal/cli/mcp_todo.go`) and the MCP factory tools keep their current output.

### Out of Scope — registry and queue contents

- Factory runs and queue items are neither purged, archived, nor edited by this SPEC. Only what the read verbs render by default changes.

### Out of Scope — mutation verbs

- `factory next`, `factory complete`, the lease verbs, and every queue mutation verb keep their current behavior.

### Out of Scope — lane join notice

- The SessionStart lane join notice in `internal/hook` belongs to SPEC-LANE-NOTICE-DIET-001 and is not changed here.
