# SPEC-ROLE-INJECTION-BUDGET-001 — Progress

> 카드 t1617 · 런 tmnboq · 기준 트리 `2aab5f797` (WT-3-2-1)

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-10-10
plan_artifacts: spec.md 0.1.0 · plan.md · acceptance.md (Tier M — 3 artifacts)
plan_author: manager-spec

계획 완결 성명(§E.1 — 리더 처분 2b7b: 감사 실행 전 기입 가능, 감사 준비 완료를 진술):

- 측정: 카드 dispatch 의 전 수치를 본 트리에서 재측정해 확정했다(core 18,114/17,793 · 헤더 123/125 · 포인터 128 · 스텁 8,583/8,072 · 원장 36행 · 갈림 1문장). 원문 `.moai/reports/t1617/measurements.md`. 생산자 기준선 4,376은 lane-15 직접 분해치로 귀속(라이브 자산 재측정 불가, 합본 교차 검산 ±100 — spec.md §B).
- 해석 결정: 「역할별 조립 ≤9,000자」를 조립 합본 판독으로 확정(spec.md §A.4). 축자 판독은 자기 모순(9,000 + 4,376 > 10,000 한도)이라 기각. 파생 core 예산 4,367(leader 구속), 설계 목표 4,000.
- 설계 결정: 리더/레인 역할 분할 기각(완전 분할도 ~9,057 > 4,367 — 산술 기각, spec.md §D), 공유 core 압축 재작성 + 재배치 감사. 구속 절은 core/스텁에만(상위 원장 vocabulary 보존), `buildRoleCore` 표지 검증 무편집.
- 채택: AC 12개 두 칸 채택(acceptance.md) — release-blocking RED 7개는 본 트리에서 실측 관측(E1–E7, exit 별도 기록), regression-guard 4개는 §2 undecidable/보존 처분 명시, 변이 탐침 요지 포함.
- Tier: M (3 artifacts) — 다중 파일이나 단일 서브시스템(주입 경로) + 규칙 2쌍 + 템플릿 미러; 헌법급 아님, 상위 원장 기계 위에 Ride.

Gap(선언): plan-audit 미실행(이 신호는 감사 준비 상태의 진술이다 — 리더 처분 2b7b에 따라 lane이 audit_multi로 실행).
