# Progress — SPEC-DOCS-LOCALE-PARITY-REPAIR-001

- 카드: t538 · 브랜치: `WT-docs-v313-locales` · 베이스: `bce6d7e08`
- 상태: **draft** (plan 페이즈 아티팩트 4종 생성 완료 — run 미개시)

## §E.1 Plan-phase Audit-Ready Signal

```yaml
phase: plan
status: draft
spec: SPEC-DOCS-LOCALE-PARITY-REPAIR-001
tier: M
artifacts:
  - spec.md
  - plan.md
  - acceptance.md
  - progress.md
measurement_ssot: .moai/reports/t538/plan-phase.md
base: bce6d7e08
scope: G1 (e2e en/zh stale) + G2 (doctor ja/zh examples) + G3 (legacy emphasis spacing, 3 lines) + G4 (skill-guide SVG rules x4) + CHANGELOG
ac_count: 11
red_now_acs: [AC-001, AC-002, AC-003, AC-006, AC-007, AC-008]
milestones: [M1-G1, M2-G2G3, M3-G4-CHANGELOG]
kickoff_gate: pending (Implementation Kickoff Approval — HUMAN GATE)
```

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## Phase Log

| 일시 | 페이즈 | 내용 |
|---|---|---|
| 2026-09-08 | plan | 측정 SSOT(`plan-phase.md`) 인용 기반 SPEC 4종 생성. baseline 재관측 완료(en:170 deferral 1건, zh:170 1건, ja/zh 예시 4행, 강조위반 ko:41·ja:39·zh:39 각 1건/en 0건, skill-guide SVG0 언급 0건). 리드 추가지시 2건(전제-정정 기록·G4 1행 근거)과 코디네이터 정정 2회(G3 = 3행 한정·괄호 마커 밖, t535 신규 절 불가침) 반영. |
