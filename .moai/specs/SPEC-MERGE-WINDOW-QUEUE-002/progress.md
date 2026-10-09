# SPEC-MERGE-WINDOW-QUEUE-002 — Progress

> Card t1582 · created 2026-10-09 by manager-spec (plan phase) · branch `WT-p1-p2-t1576` · base develop tip `f7606c7bc`

## §E.1 Plan-phase Audit-Ready Signal

- Artifacts: spec.md + plan.md + acceptance.md (Tier M set) + decision-index.md (`interview.decision_gate: on`) + progress.md.
- SPEC id regex check (Bash) → `PASS`; uniqueness: `ls .moai/specs/ | grep -i MERGE-WINDOW` → SPEC-MERGE-WINDOW-QUEUE-001 만 존재 (002 자유).
- RED 재현 (plan 단계 실측, 트리 `f7606c7bc`, env-scrub 복합 형태):
  - ④-R7-2: `TestRedT1582MixedSweepWithNoTestPackageCounts` — `runner reported [no test files] — an empty sweep cannot stand for a re-measure` (거짓 거부 관측).
  - ④-R7-3 단위: `TestRedT1582VerifierAdmitsAShortBaseSHA` — 검증기가 3바이트 Base 통과 관측.
  - ④-R7-3 e2e: `TestRedT1582ShortBaseSHAPanicsBeforeWindowRelease` — `runtime error: slice bounds out of range [:12] with length 3`, release 전 사망 + 창 held 관측.
  - item ②: round-trip 관측 — 경계 붕괴 **미재현**(4/5 일치, apostrophe는 escape 바이트 잔류·경계 유지) → 봉인 테스트 GREEN.
  - item ①: probe 관측 — non-test 명령의 실패 의미 출력이 exit 0으로 valid (spec.md §D 잔여, 수리 대상 아님).
  - item ③·R7-1: 재현 설계 확정 — 관측은 run 단계 M3/M4 첫 행위 (plan.md §F).
- RED 오버레이 테스트 파일: `internal/factory/remeasure_red_t1582_test.go`, `internal/cli/shelljoin_red_t1582_test.go` (트리에 작성, 미커밋 — run 단계 M1-M4가 GREEN으로 뒤집는 대상).
- 10 REQ / 8 AC (v0.2.0 — D5 분할: 002/008, 003/009, 006/010). Out of Scope 경계: 보호구역·todo 발행·nominate(리더 큐), non-test 명령 잔여, 호출자 리다이렉션, `merge gate` verdict 설계, complete의 release-failure 출력.
- plan-audit iteration 1: FAIL 0.69 / Tier M 0.80 (`.moai/reports/t1582/plan-audit.md`) — D1-D6 아티팩트 수리 완료(v0.2.0, plan.md·acceptance.md·spec.md).
- plan-audit iteration 2: FAIL 0.85 (동일 파일 Iteration 2 섹션) — 라운드 2 수리 완료(v0.3.0): D19(M1 렌더 12곳 실측 열거+총괄 규칙+cli 스코프 — 감사 prose "14"는 자기 목록 10+2의 오산술, 리더 독립 grep과 본인 측정 일치)·D13(§4 목록 교체)·D14(cause-7 시딩 설비 `internal/factory/mergestep_red_t1582_test.go` + merge-ready RED `internal/cli/factory_merge_ready_red_t1582_test.go` plan 단계 작성·실측, AC-005/006 셀 verbatim 기록, AC-006 실행 계수 가드 `--- PASS` ≥ 2)·D15(스크럽 변수 3종 전체 목록 명기+축약형 전면 제거)·D17(DoD tracked+untracked 수집 대조)·D20(GNU grep `-r`+exit 구분)·D18(version 0.3.0 정합). RED 총 4건 실측: R7-2·R7-3 단위/e2e·merge-ready ③.
- 429 재개 기록: 라운드 2 중 429(요청 한도) 2회 — 트랜스크립트 재개로 잔여 목록(디스크 상태 대조) 수행, 부분 상태는 위 행들이 증언.
- plan_status: audit-ready (pending plan-auditor verdict)

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
