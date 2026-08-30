# SPEC-BINLAG-INVOCATION-001 — 진행 기록

카드 t366 · 브랜치 `WT-lint-binary-lag` · 증거 경로 `.moai/reports/t366/`

## §E.1 Plan-phase Audit-Ready Signal

- 상태: `draft`. plan-phase 산출물 4종 작성 완료(`spec.md` / `plan.md` / `acceptance.md` / `progress.md`).
- Tier: **S** 제안. 근거는 `spec.md` §6의 어느 안을 골라도 예상 변경이 5파일 미만·300 LOC 미만이라는 점 — 비교 로직은 `internal/binlag`에 이미 있고 이 SPEC은 소비 지점만 더한다.
- 요구 8 / 수락 8. Tier S 상한(각 8) 정확히 충족.
- 작성 트리: `d7010f86a`. 이 문서의 모든 RED-now 셀이 그 트리에서 관측됐다.
- **미결(run-phase 진입 차단)**: `plan.md` §F M0의 `[NEEDS CLARIFICATION]` 3건. 수리 안 결정 전에는 run-phase에 들어가지 않는다.
- 이 세션이 실행하지 않은 것: 양방향 재현(설계만), 명령 단위 census, 호출당 비교 비용 측정.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
