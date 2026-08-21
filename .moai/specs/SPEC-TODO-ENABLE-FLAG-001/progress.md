# Progress — SPEC-TODO-ENABLE-FLAG-001

## §E.1 Plan-phase Audit-Ready Signal

- 2026-08-22 (iter2) — plan-audit iter1 **FAIL 0.78**(Tier M 임계 0.80; must-pass 7/7 PASS, 부족분은 Testability 0.65 단일 축) 대응 개정. 블로킹 5건(D1 공허 AC / D5 커밋 후 무조건 통과 / D6 Route 오기 / D2 명시적 호출 미정의 / D4 충돌 규율 단언) + 선택 4건(D3·D7·D8·D9) 처리. `depends_on` 미기재 근거를 "의존 없음"에서 **동시성 대 직렬화 트레이드오프 기록**으로 재작성. version 0.1.0 → 0.2.0, 항목별 처리 내역은 `spec.md` §G. AC 11개 유지(개수 불변, 관측 해상도만 상향).
- 2026-08-22 (iter1) — plan-phase 산출물 3종(Tier M) 작성. 카드 t170에서 AC 예산 초과(32 > 상한)로 분리 신설됐다. 형제 SPEC은 `SPEC-FEEDBACK-AUTO-SUBMIT-001`(Tier L). 근거는 `.moai/reports/t170/lens-web-todo.md`·`lens-init.md`. 카드 전제 P4(전면 억제 불가) 정정과 운영자 결정 D3·D4 반영. AC 11개 / Tier M 상한 16. 남은 결정: M6 템플릿 적재 여부(기본안 = 싣지 않음).

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
