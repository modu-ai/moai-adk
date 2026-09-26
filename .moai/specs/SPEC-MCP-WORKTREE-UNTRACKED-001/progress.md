# SPEC-MCP-WORKTREE-UNTRACKED-001 — Progress

Card: t1202 | Branch: WT-worktree-moai-root | Base: origin/develop `df526c9a9` | Tier: M

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-09-26
tier: M
spec_version: "0.2.0"
artifacts: [spec.md, plan.md, acceptance.md, progress.md]
spec_id_check: "Bash regex PASS on SPEC-MCP-WORKTREE-UNTRACKED-001; ID absent from .moai/specs (count 0)"
baseline_tree: df526c9a9
premise_evidence: .moai/reports/t1202/verdict.md
plan_audit_iter1: ".moai/reports/t1202/plan-audit.md — FAIL 0.62"
recommendation: "design (a) alone; operator decides at Kickoff (plan.md §C, 7 decisions)"
```

### Revision map for plan-audit iter-1 (spec v0.1.0 → v0.2.0)

| Defect | Change |
|---|---|
| D1 | plan.md §C keeps the three markers as operator decisions 1–3 plus decisions 4–7, each with a recommended default and the REQs/ACs it changes; REQ text is written under the defaults and carries no marker. |
| D2 | Graph tools are class G (spec §3); REQ-MWU-009 keeps them on the tree root with the existing "graph layer absent" error; AC-MWU-012. |
| D3 | Source-root predicate = MoAI configuration (`.moai/config/sections`), not `.moai` existence; REQ-MWU-011; AC-MWU-014. |
| D4 | spec §3 per-tool read/write inventory with code sites (verified against code); REQ-007 split into REQ-MWU-007..010 by class; ACs per class: T AC-002, C AC-011, G AC-012, S AC-013. |
| D5 | AC-MWU-006 (admin entry removed); AC-MWU-004 now also covers the primary-itself case. |
| D6 | AC-MWU-009 now requires a valid `W` to be accepted under hostile `GIT_DIR`/`GIT_WORK_TREE`; the unrelated-dir rejection is the secondary check. |
| D7 | spec §2.2 states the issue's `moai worktree new` reproduction and its acceptance of (b); §2.3 keeps only measured reasons. |
| D8 | AC-MWU-001 carries three RED-first predicates (is-ancestor, `_test.go`-only RED commit, RED-tree test failure). |
| D9 | plan §B.2 forbids unscrubbed helpers and the shape fallback; REQ-MWU-005/006 require scrubbed git and no alternative retry. |
| D10 | REQ-MWU-004 primary-checkout predicate (common-dir parent + porcelain first entry); ambiguous layouts rejected; AC-MWU-007; spec §6. |
| D11 | spec §2.2 separates 2 execution sites from 3 entry points; credential claim marked unmeasured and dropped as a reason. |
| D12 | AC-MWU-015 checks `projectRootDescCommon`. |
| D13 | AC-MWU-002 names `collectReviewDiff`. |
| D14 | Pattern labels changed to Event-driven. |
| D15 | plan §B.3 file estimate and recount-after-M1 rule. |
| D16 | spec §2.1 describes the check as registration + common-dir, not containment; §2.2 names the `WorktreeCreate` hook as a MoAI entry point. |

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
