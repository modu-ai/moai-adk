# SPEC-FACTORY-BOOT-MINIMAL-001 — Progress

> 카드 t1923 · 기준 트리 `59959629908cda723c37e26c7c15718194d6b38a` (WT-factory-lane-internal) · status: draft

## §E.1 Plan-phase Audit-Ready Signal

audit_ready: true
plan_status: audit-ready
plan_complete_at: 2026-10-23
plan_artifacts: spec.md 0.1.0 · plan.md · acceptance.md · decision-index.md (Tier M — 3 artifacts + decision index)
plan_author: manager-spec

- 측정: 팩토리 고정분·역할 블록·합본을 실제 렌더러로 4개 언어 × 3개 소스 × 역할별 실측(plan.md §B.2), 카드의 23,145자 넘침을 라이브 훅 출력으로 분해해 원인(10-21 primary checkout의 압축 전 core 18,367)을 확인(§B.3), AGENTS.md Codex 상한 테스트 실행(기준 트리 PASS, 여유 75)과 운영자 패치 검토(§B.5).
- 결정: 0.6.0 예산 수치 유지, 생산자 4,797을 허용분 1,950 + 팩토리 고정분 상한 2,847로 분해(decision-index Q4). 나머지 결정은 decision-index Q1~Q7.
- 채택: AC 9개 — release-blocking 4개(AC-FBM-001~004, 증거 E1~E4 기준 트리 RED 관측), regression-guard 5개.
- Gaps: AC-FBM-009 테스트 4개 plan 단계 미실행.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
