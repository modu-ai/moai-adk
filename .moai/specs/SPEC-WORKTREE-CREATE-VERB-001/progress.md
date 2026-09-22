# SPEC-WORKTREE-CREATE-VERB-001 — Progress

> Plan-phase artifact set authored 2026-09-22 by manager-spec (card t1070) in worktree `.claude/worktrees/t1070` (branch `WT-worktree-verb`, HEAD `0314801c2`). status: draft.

## §E.1 Plan-phase Audit-Ready Signal

```yaml
spec_id: SPEC-WORKTREE-CREATE-VERB-001
phase: plan
status: draft
plan_status: audit-ready
plan_complete_at: 2026-09-22
author: manager-spec (card t1070)
baseline_tree: .claude/worktrees/t1070 @ 0314801c2 (WT-worktree-verb)
artifacts:
  - spec.md
  - plan.md
  - acceptance.md
  - research.md
  - progress.md
self_check:
  spec_id_regex: PASS   # [[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS  → PASS
  id_uniqueness: PASS   # 908-SPEC catalog grepped; no SPEC-WORKTREE-CREATE-VERB-001 collision
  frontmatter_schema: PASS  # canonical 12 fields present; phase: "v3.2.0 target"; status: draft
  gears_notation: PASS  # spec.md §B — Ubiquitous/When/Where patterns; no legacy IF/THEN
  out_of_scope_rule: PASS  # three "### Out of Scope —" H3 sub-headings with "-" bullets
  flat_file_rejection: PASS  # directory layout .moai/specs/SPEC-WORKTREE-CREATE-VERB-001/
amendment_note: "iter-1 plan-audit PASS-WITH-DEBT 0.80 (2026-09-22, .moai/reports/plan-audit/SPEC-WORKTREE-CREATE-VERB-001-iter1.md); D2-D5 micro-amendment applied per re-delegation — REQ-WCV-009 added (path-escape + plain-dir refusal, D3), REQ-WCV-005 gate-independence clause (D2), research.md Fact 1 inventory completed (D4), REQ-WCV-004 relabeled Event-driven (D5)"
open_items:
  - "[NEEDS CLARIFICATION: worktree-verb direction (가) vs (나)] — plan.md §C; resolution precondition: live Codex session observation (t1050 count was 0) must be measured before the orchestrator's AskUserQuestion round"
```

Note for plan-auditor: the `[NEEDS CLARIFICATION]` marker is intentionally present in plan.md (and research.md §C) and intentionally ABSENT from spec.md/acceptance.md, per the marker placement rules. It is the card t1070 [HARD] decision-axis requirement, gated on a measurement, not an omission.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
