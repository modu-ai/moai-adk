# SPEC-TODO-HOLD-STATE-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

```yaml
spec: SPEC-TODO-HOLD-STATE-001
phase: plan
plan_status: audit-ready
tier: M
measured_at:
  head: 8a969dfc0
  worktree: .moai/worktrees/t1308
  branch: WT-todo-hold-state
  date: 2026-09-29
artifacts:
  - spec.md
  - plan.md
  - acceptance.md
  - progress.md
requirements: 16
acceptance_criteria: 16
tier_ceiling_disposition: |
  plan-audit iter1 D2 (19/19 vs Tier M 16/16): resolved by CONSOLIDATION (auditor option a).
  Folds: REQ-THS-007→REQ-THS-006, REQ-THS-009→REQ-THS-008 (refusal halves absorbed as second
  When-clauses), REQ-THS-015→REQ-THS-011 (autodone covered by the actionable-surface enumeration);
  AC-THS-007→AC-THS-006, AC-THS-009→AC-THS-008, AC-THS-014→AC-THS-011(c, regression guard).
  Freed numbers (REQ/AC 007, 009, REQ 015, AC 014) are deliberate consolidation gaps recorded in
  spec.md HISTORY 0.2.0 — not authoring errors. Tier-up and split were rejected: the content fits
  M without the two additional Tier L artifacts.
lead_memo_survival:
  rebuild_safety_contract: "REQ-THS-002/003/004 — ids and content intact (AC-THS-002/003/004/019)"
  negative_predicate_sweep: "REQ-THS-011/012 — ids and content intact; mutation-drift AC = AC-THS-011 (id intact), autodone regression-guard row absorbed as (c)"
premise_corrections:
  - "카드 전제 「internal/kanban 이력상 ALTER TABLE 0건」 기각 — ensureLandingColumn (backlog_sqlite.go:439-455, t359)의 ADD COLUMN 선례 1건 존재. 마이그레이션 결정(리빌드+스탬프)은 불변."
  - "iter1 D1 정정 — todo_autodone.go:283 의 부정 disjunction 은 미래 상태를 삼키는 것이 아니라 이미 건너뛴다 (행동 정상·형태만 결함). 행동 red-now 는 todo.go:920 픽 게이트 (dropped 만 거절). autodone AC 는 green-at-M1 회귀 가드로 재분류."
lead_memos_folded:
  - "부정 술어 전수 전환 REQ-THS-011/012 (mutation-testable drift AC-THS-011)"
  - "리빌드 안전 계약 REQ-THS-002/003/004 (JSON→SQLite 규율 계승, SPEC-TODO-SQLITE-001)"
```

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## 진행 기록

- 2026-09-29 (plan, card t1308, lane worker-72): Tier M artifact set authored in worktree
  `.moai/worktrees/t1308`. Measured basis exported to `.moai/reports/t1308/predicate-sweep.md`
  (state vocabulary, CHECK/DDL comment, schema_version machinery, predicate inventory S1-S13/D1-D4,
  actor precedent, doc surfaces). Two lead memos folded as REQs. One card premise falsified
  (ALTER TABLE 0건 → 1건 ADD COLUMN 선례), decision unchanged. Lint evidence: see
  `.moai/reports/t1308/lint.txt`.
- 2026-09-29 (plan-audit iter1 remediation, v0.2.0): FAIL 0.90 (2 blocking) — D1 autodone
  filter-direction 정정 (4개 표면: acceptance RED-now 셀·plan §A/§B.5·sweep S5·spec §A.1/§B.4,
  회귀 가드 재분류), D2 16/16 통합 처분 (위 tier_ceiling_disposition), D3 Event-detected →
  Event-driven 재표기 (REQ-THS-003/005/014), D4 P3 전제 인라인 인용+primary-checkout-local 표기.
  재측정: REQ 16 / AC 16, lint 0 error. 커밋 없음 — 리드 검토 후 커밋.
