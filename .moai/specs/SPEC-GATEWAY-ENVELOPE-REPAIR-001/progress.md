---
id: SPEC-GATEWAY-ENVELOPE-REPAIR-001
title: "Reasoning-envelope repair — progress"
version: "0.1.0"
created: 2026-09-13
updated: 2026-09-13
author: manager-spec (card t708)
---

# Progress — SPEC-GATEWAY-ENVELOPE-REPAIR-001

## §E.1 Plan-phase Audit-Ready Signal

```yaml
spec: SPEC-GATEWAY-ENVELOPE-REPAIR-001
card: t708
phase: plan
status: draft
tier: M
artifacts: [spec.md, plan.md, acceptance.md, progress.md, research.md]
base: local develop 7a7a08f20
branch: WT-envelope-persist
symbol_pins_verified_at_plan_time: true
security_determination: spec.md §3 (byte-exact re-injection satisfies the t672 binding)
sibling_seam: spec.md §4 (t700 REQ-WRR-007 vs REQ-WRR-008(d); Reading B adopted, adjudication gated at M0)
needs_clarification_count: 2  # plan.md §H — both bounded by the refusal path (REQ-EVR-007)
```

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

- plan_complete_at: 2026-09-13T13:44:15Z
- plan_status: audit-ready
- plan_audit: PASS 0.975 (iteration 2/2) — .moai/reports/t708/plan-audit.md

## §E.1 Addendum — M0 Adjudication Record (2026-09-14)

Two-input adjudication per plan.md §F M0 / §H Resolution Record row 3. Verdict-only card; zero code changes in this step.

**Input 1 — t700 seam (Reading A/B)**: t700 SPEC-GATEWAY-WEDGE-REROOT-001 remains unlanded (draft in `.claude/worktrees/t700`). Its drafted text bans envelope re-issue/synthesis/transplant with a rationale that bans MANUFACTURED envelopes ("the Opaque digest must always reference something the gateway itself issued"); verbatim gateway-issued re-injection satisfies that rationale. t708 spec.md §4 Reading B position stands with no adverse finding. Final confirmation re-rides the M1 develop absorb, where §A step 2 re-reads t700's landed status (merge order is the lead's call at landing time).

**Input 2 — t707 verdict** (`.claude/worktrees/t707/.moai/reports/t707/verdict.md`, soon merged to develop): no reproducible serialization defect (8 synthetic publish→stream→client-model→Check round-trips all GREEN; tamper negative tests pass — Check strength unchanged). Live CauseChain 400s localized only to a common pattern (first request ~80-90ms after an Edit tool run) with transcript≠request bytes PROVEN (prefix recompute mismatch from boundary[0] on a request pair that previously passed). Final localization awaits one live instrumentation run (t707 observability.patch, MOAI_RECEIPT_DEBUG=1) — operator decision pending, owned by t707's follow-up, NOT assigned to this SPEC. Per §H row 3: scope stays CauseReasoning (REQ-EVR-004); M3+ proceeds unblocked on this axis.

**Design directive folded (lead)**: this SPEC's design rests on client-side persistence / replay byte preservation without presupposing any server-side fix. t707's transcript≠request-bytes finding CORROBORATES the design premise: the family transcript retains the correct issued bytes (t708 research addendum: 36 complete carriers measured) while the outgoing request diverges — the launcher-side verbatim re-injection path is exactly the byte-preservation mechanism, and it presupposes no server-side change (REQ-EVR-001 validator lock).

**M0 outcome: PASSED — no adverse scope finding from either input. M1 (develop absorb + pin re-verification) is ready; awaiting the lead's window signal so the absorb carries the t707 merge.**
