# SPEC-CODEX-PARTIAL-WIRING-001 — 진행 기록

> 카드 t499 · Tier M · 3-phase(plan → run → sync)

## §E.1 Plan-phase Audit-Ready Signal

- **작성 시점 트리**: worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t499`, branch `WT-codex-partial-wiring`, base `ace1c5440`
- **산출물 4종**: `spec.md`, `plan.md`, `acceptance.md`, `progress.md`
- **Tier 판정**: M — 근거는 `plan.md` §B/§E의 반경(소스 1~2본 + 시험 1본), 요구 10건 / AC 9건(릴리스 게이트 6 + 회귀 가드 3), 그리고 사용자에게 보이는 출력 표면 변경이 포함된다는 점
- **측정 전제의 출처**: `.moai/reports/t499/repro.md`(레인 선행 측정). 이 SPEC은 그 값을 **재측정하지 않고 인용**하며, 인용 사실을 여기에 명시한다
- **plan-phase에서 실행한 확인**: SPEC ID 정규식 검사(`SPEC-CODEX-PARTIAL-WIRING-001` → `PASS`), 기존 SPEC 디렉터리 중복 없음(`ls .moai/specs/ | grep -i PARTIAL` → 무출력), 골든 하네스 재독(`internal/cli/doctor_golden_test.go`)
- **미검증 전제**: `plan.md` §I에 3건 기재 (release 일정 대조 없음 / 위저드 경로 미측정 / `moai update` 거동 미측정)

### plan-audit iter1 수리 라운드 (2026-09-07)

- **판정**: FAIL 0.76 (Tier M 임계 0.80) — 판정서 `.moai/reports/t499/plan-audit-iter1.md`. must-pass 7건 전부 PASS, `spec lint` 무결, FAIL은 Testability 0.60 한 축이 끌었다
- **수리 대응**: D1 → `acceptance.md` AC-CPW-007에 반쪽 경로를 실제로 밟는 신규 시험 추가 / D2 → 릴리스 게이트 5건에 트리 `ace1c5440` 실측 RED-now 셀 부여 + SHA를 `acceptance.md` 머리에 pin / D3 → REQ-CPW-009 신설 + AC-CPW-002에 `initCodexAdvice` 부재 단언 / D4 → AC-CPW-001에 `claude-only` 0회 단언 / D5 → REQ-CPW-010 신설 + AC-CPW-008 재매핑 / D6 → `spec.md` §A.3 좌표 `:105` → `:92` / D7(optional, 수용) → §D-1에 보호 대상 구분 추가
- **레인 이월 2건**: `acceptance.md` 매트릭스의 축약 REQ 표기를 전체 ID로 폄; `plan.md` §I의 `phase` 갭을 리드 측정으로 격하(값은 관측과 일치, 로드맵 확인은 아님)
- **관측했으나 이 카드 범위 밖**: (1) `plan-auditor`의 추적성 검사가 축약 REQ 표기(`REQ-CPW-001, 002`)를 놓쳤다 — D5로 매핑 오류는 찾아냈으니 검사 자체는 돌았고, 축약형만 통과시켰다. (2) `moai spec lint`도 같은 표기를 잡지 못한다(`CoverageIncomplete` 미발화) — lint 규칙 쪽 후속 후보이며 이 카드에서 손대지 않는다
- **RED-now 실측 5건**(전부 이 트리에서 직접 실행, 리다이렉트 후 `echo $?`로 종료코드 판독): 다섯 셀렉터 모두 `exit=0` + `[no tests to run]` — 아직 없는 시험이라 판정 불능이며, 그 상태를 RED로 읽는다. 원문은 `acceptance.md` §D.1 각 항목

_run-phase 진입 전 상태: `status: draft`._

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
