---
id: SPEC-HEAVY-TEST-SLOT-001
title: "Plan — extend the lane slot paragraph with enforcement code, named packages, incident record"
version: "0.2.0"
created: 2026-09-14
updated: 2026-09-14
author: lane-6
tier: S
---

## §A — Approach

The card's design question was answered by measurement: the surface exists (`moai slot`), is enforceable (live-holder second acquire refused, exit 3 — demonstrated on `t774-repro-demo`), carries the card's field list, and is in active use since 2026-09-12. The lane duty text also already exists (§8 of the lane protocol rule, e78fd0ee6). v0.1.0's plan to add a standalone protocol document was therefore wrong — it would parallel an existing procedure copy. v0.2.0 extends the existing §8 slot paragraph with the three genuinely additive items and touches nothing else.

## §B — Milestones

### M1 — extend §8 (one run commit; SPEC close on the sync commit)

1. Edit `.claude/rules/local/gitflow-lane-protocol.md` §8 slot paragraph, appending three sentences to the existing duty text (no restructuring, no second procedure copy):
   - the enforcement fact: second acquirer while a live holder is inside its bound → refused, exit 3, silent, no displacement (measured 2026-09-14, `t774-repro-demo`);
   - the named WHEN list: at minimum `internal/cli` (the 09-10 incident subject), plus `internal/kanban` and `internal/hook` as measured-minute-scale packages;
   - the control-group record: the 2026-09-10 incident (three lanes, load 8–21, verdict flips) as the no-surface evidence, the exit-3 demo as the enforcement evidence, and the prohibition on concurrent heavy-suite re-runs to re-demonstrate.
2. Verify per acceptance.md; commit (run phase) and close the SPEC on the sync commit.

## §C — Risks / Notes

- `.claude/rules/local/gitflow-lane-protocol.md` is dev-only and already carries internal card ids and dates — the additive sentences cite t774/09-10 freely there, and nowhere else.
- `.claude/rules/moai/workflow/resource-slot-lease.md` is template-mirrored: it must NOT receive the incident record or card ids. If its user-facing text ever wants an exit-code table, that is a separate product-surface decision.
- The WHEN list is deliberately a named-examples list, not a duration gate — a mechanical threshold would be a product change this SPEC excludes.
- Plan-audit iteration 1 (FAIL 0.70) findings D1–D6 are all folded here; the audit's own report records what it independently re-verified (slot.go registration/fields/exit code, §8 presence, audit-log first entry).
