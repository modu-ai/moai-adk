# SPEC-FACTORY-BOOT-MINIMAL-001 — Progress

> 카드 t1923 · 기준 트리 `59959629908cda723c37e26c7c15718194d6b38a` (WT-factory-lane-internal) · status: draft

## §E.1 Plan-phase Audit-Ready Signal

audit_ready: true
plan_status: audit-ready
plan_complete_at: 2026-10-23
plan_artifacts: spec.md 0.2.0 · plan.md · acceptance.md · decision-index.md (Tier M — 3 artifacts + decision index)
plan_author: manager-spec
plan_revision: 0.2.0 — plan-audit round 1 FAIL 0.69 (`.moai/reports/t1923/plan-audit.md`, codex `.moai/reports/t1923/boot-minimal-plan-audit.md`) corrected; blocking D1-D5 and optional D6-D10 addressed

- 측정: 팩토리 고정분·역할 블록·합본을 실제 렌더러로 4개 언어 × 3개 소스 × 역할별 실측(plan.md §B.2), 카드의 23,145자 넘침을 라이브 훅 출력으로 분해해 원인(10-21 primary checkout의 압축 전 core 18,367)을 확인(§B.3), AGENTS.md Codex 상한 테스트 실행(기준 트리 PASS, 여유 75)과 운영자 패치 검토(§B.5), 녹색 경로 모의(§B.7 — 영어 레인 startup 2,727, 경계 변이 2,846 관측).
- 결정: 0.6.0 예산 수치 유지, 생산자 4,797을 허용분 1,950 + 결합자 2 + 팩토리 고정분 상한 2,845로 분해(Q4). 리더 공지는 자동 모드 해설만 옮기고 나머지는 그대로(Q8 — SPEC-AUTONOMY-BATCH-GATE-001 REQ-BGS-016/017 보존). 재배치는 R1~R5, 목적지는 기존 문서(Q5·Q10). 나머지는 decision-index Q1~Q10.
- 채택: REQ 11개, AC 12개 — release-blocking 5개(AC-FBM-001·002·003·004·011, 증거 E1~E4·E11 기준 트리 RED 관측), regression-guard 7개.
- Gaps: AC-FBM-009 테스트 13개 plan 단계 미실행(0.1.0의 4개는 감사가 4/4 PASS 관측). ko·ja·zh 녹색 경로는 실행 모의 없음.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §G Override and Refusal Record

- 운영자 결정 2026-10-23: 천장 예외 1회(델타 한정), 리더 세션 3936623d 경유 — recorded by lane-8 on the leader's relayed message (cross-session, 2026-10-23); the lane did not receive the decision from the operator directly. Scope: Tier M plan-audit ceiling (2 rounds, reached at round 2: FAIL 0.76, rcpt-87831eae3fa19ae402a21281) is excepted for ONE delta re-audit limited to the verdict's `fix_scope` anchors (N1-N4); a PASS goes to kickoff; a further FAIL returns to the hold-and-split disposition with no further exception.
