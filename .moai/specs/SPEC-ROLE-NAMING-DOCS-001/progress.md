# Progress — SPEC-ROLE-NAMING-DOCS-001

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
spec_id: SPEC-ROLE-NAMING-DOCS-001
spec_version: 0.2.0
card: t1257
tier: L
artifacts: [spec.md, plan.md, acceptance.md, design.md, research.md, progress.md]
requirements: 24
acceptance_criteria: 24
measured_at: "worktree .claude/worktrees/t1257, branch WT-role-naming-docs, base e62c3e183"
inventory: .moai/reports/t1257/inventory.md
gate: "substitution milestones M3-M6 blocked until SPEC-ROLE-NAMING-CODE-001 (t1256) is on develop at implemented/completed with a term table listing no accepted legacy spelling (REQ-RND-002)"
open_questions: 0
body_substitutions_performed: 0
```

### Operator decisions (recorded 2026-09-26)

Relayed to this lane by the coordinator on 2026-09-26. Q1/Q3/Q4/Q5 answered by the operator in the leader window; Q2/Q6/Q7 in this lane window. Full table: research.md §F.

| Q | Decision |
|---|---|
| Q1 | `lane` canonical (`lane-<n>`, `-f lane`); `worker` / `agent` aliases removed immediately, no compatibility alias; docs describe no legacy alias. Conflicts with t1256 design.md §3 at `6fe67c674` ("legacy accepted, hinted"); the leader sent t1256 the same answer, docs follow the operator answer. |
| Q2 | Kanban plan / run / sync companions stay as-is (not lanes). |
| Q3 | Lane self-dispatch allowed up to promoting a queued card; "Promotion is the operator's act, always" and "The lead is the queue's sole producer" plus echoes are amendment targets; production unchanged except the leader rename; no HARD clause silently lost. |
| Q4 | Keep `manager-lead`; prose says leader; B/C/D recorded as rejected alternatives. |
| Q5 | Both leader usages allowed; qualify on first occurrence per file (en factory leader / team lead(er); ko 팩토리 리더 / 팀 리더). |
| Q6 | ko 리더 / 레인; ja リーダー / レーン; zh 主导 (主导会话) / 泳道; zh role-sense 主导·主控·领导·负责人 unify to 主导. |
| Q7 | foreman, deputy, coordinator keep their names; one-line "leader's auxiliary role" definition at each definition site. |

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
