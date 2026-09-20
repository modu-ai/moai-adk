# SPEC-APPJS-IIFE-GUARD-001 — 진행 기록

- SPEC: SPEC-APPJS-IIFE-GUARD-001
- card: t1048
- worktree: `.claude/worktrees/t1048` (branch `WT-iife-guard`)
- plan-phase base: `d8304b49a`

---

## §E.1 Plan-phase Audit-Ready Signal

**Claim** — plan-phase 산출물 4종(spec.md / plan.md / acceptance.md / progress.md)이 작성됐고, SPEC 이 딛고 선 실측값은 이 트리에서 이번 실행으로 측정됐다.

**Evidence**

- SPEC ID 정규식 자가검사: `[[ "SPEC-APPJS-IIFE-GUARD-001" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]]` → `PASS`
- IIFE 스팬: `grep -n '^[^[:space:]]' internal/web/assets/app.js` → 컬럼 0 IIFE 둘, `[(39,660),(673,796)]`
- 핸들러 인벤토리: `grep -n 'addEventListener' internal/web/assets/app.js` → 20건, 그중 bare identifier 4건 (530·548·551·727)
- 선언 위치: `grep -n -E '(function|var|let|const)[[:space:]]+(syncSegmentsVisibility|initConsole|stampRefreshed)\b'` → 45 / 526 / 711
- 최상위 문장 형태 전수조사 → 그 밖의 형태 **0** (상세는 spec.md §B.2)
- 프로토타입 2방향: 정방향 `checked=4, violations=0, EXIT=0` / 돌연변이 `checked=4, violations=1, EXIT=1` (`line 552: stampRefreshed — declared in IIFE#[1] not #0`)
- 의존성: 직접 30 / 전체 require 89, JS 엔진 모듈 0건
- 수리 커밋 조상 확인: `git merge-base --is-ancestor 18fa0d384 HEAD` → `YES`

**Baseline-attribution** — 전부 worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1048`, branch `WT-iife-guard`, HEAD `d8304b49a` 에서 이번 plan-phase 실행 중에 측정.

**Gaps**

- `go test ./internal/web/` 를 **돌리지 않았다** — plan-phase 는 저작만 하며, 기존 패키지 테스트의 현재 상태는 이 실행에서 측정되지 않았다.
- 돌연변이는 스크래치패드의 사본에 대해서만 합성했다. 작업 트리의 `app.js` 는 읽기만 했고 바이트 하나 바뀌지 않았다.
- 규칙이 spec.md §F 의 세 한계를 못 잡는다는 것은 **선언**이며, 그 방향으로 돌연변이를 돌려 확인하지는 않았다.

**Residual-risk**

- 최상위 문장 형태 전수조사는 파서가 아니라 줄 형태 기준이다. 여러 줄 문장의 이어지는 줄이 다른 형태를 숨기고 있을 가능성은 배제되지 않았다.
- 프로토타입은 Python 이고 구현은 Go 다. 정규식 방언 차이(특히 `$` 의 단어 문자 취급)가 이식 과정에서 결과를 바꿀 수 있다 — run-phase 에서 2방향을 **Go 로 다시** 재야 한다.

---

## §E.2 Run-phase Evidence

_<pending run-phase>_

---

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

---

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
