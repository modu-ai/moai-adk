# SPEC-WORKTREE-EXIT-RETURN-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

- Tier: S — plan-phase 산출물 `spec.md` + `plan.md` + `acceptance.md` (카드 지시로 acceptance.md 포함).
- 근거: `.moai/reports/t965/observations.md` (관측 1~4). 이 단계에서 새 측정을 수행하지 않았다.
- status: `draft`.
- plan-audit iter1: **FAIL 0.69** (Tier S 문턱 0.75). must-pass 7/7 통과 — FAIL 은 루브릭 점수와
  열거된 결함이 만들었다. 판정서: `.moai/reports/t965/plan-audit-verdict.md`.
- 반영 v0.2.0 (2026-09-20): D1·D2·D3·D4·D5·D6·D7·D8·D9 전부 문언 층에서 수정. 코드 변경 0.
- plan-audit iter2: **FAIL 0.80** (Tier S 문턱 0.75 를 **넘는다**). must-pass 7/7, 네 축 모두
  상승. FAIL 을 만든 것은 점수가 아니라 새 critical 결함 **D1′** 하나다 — 0.2.0 의 비대칭 결정이
  `internal/template/rule_template_mirror_test.go` 의 바이트 동일성 허용목록과 충돌해,
  **SPEC 문언을 만족시키는 트리 상태가 존재하지 않는** 상태를 만들었다(변이 실험으로 기계 입증).
  판정서: `.moai/reports/t965/plan-audit-verdict-iter2.md`.
- 반영 v0.3.0 (2026-09-20): 리드 결정으로 **Path A 채택**(Path B 는 기각 — `internal/` 테스트
  파일 2개 편집이 필요해 이 SPEC 의 [HARD] 와 충돌하고, 최소 이행이 상시 가드를 0 으로 만든다).
  두 사본을 **바이트 동일**하게 두고 양쪽 모두 **런타임 버전 인라인**만으로 귀속한다.
  - `spec.md` §1.5 전면 재작성(가드 셋 표 + 허용목록 등재 사실 + kanban-dispatch 전례).
  - REQ-WXR-008 → 바이트 파리티, 가드를 이름으로 지목. REQ-WXR-005 → 두 사본 동일 귀속.
  - AC-WXR-007 → 바이트 동일성 / 양쪽 공통 부재 / **세 가드 전부** 완료 이전 관측 / 같은 커밋.
  - D9 분할은 **초과분** 선언(논리는 유효하나 전제가 사라짐) — 증거 경로 인용은 `spec.md`·
    `progress.md` 층으로 이동, 반출 의무와 `git ls-files` 판정은 그 층에서 유지.
  - D2′ 는 **경로 선택으로 닫힘**(편집 아님) — 허용목록을 건드리지 않으므로 상시 가드 생존.
  - D11(REQ-001 표본 명시) → AC-WXR-001 판별식 (4). D12(REQ-005 도구 계약 인용 한정) →
    AC-WXR-005 판별식 (6). **AC 수 8 유지**(Tier S 상한), REQ 수 8 유지.
  - 코드 변경 0. status `draft` 유지.
- 잔여: AC-WXR-007 (2)(b)(저장소 내부 경로 부재)는 **어느 누출 클래스도 받치지 않는다** — 이
  AC 자신의 grep 이 유일한 검사다(판정서 측정). 결함이 아니라 잔여 위험으로 문언에 명시했다.
- 반출(M0) 자체는 run-phase 의 일이며 plan 단계에서 수행하지 않았다(아래 §E.2).

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
