# 인수 기준 — SPEC-CTX-BLIND-DOUBLE-001

> 모든 채택 판정은 **뮤턴트**다. 커버리지 상승은 인수 근거가 아니다.
> 모든 `-run` 셀렉터에는 같은 셀렉터의 `-list` 건수가 함께 기록되어야 한다.

## §D 인수 기준 표

| AC | 마일스톤 | 요구사항 | 판정 방식 | 분류 |
|---|---|---|---|---|
| AC-CBD-001 | M1 | maps REQ-CBD-001, REQ-CBD-002 | 문서 검사 | **regression-guard** |
| AC-CBD-002 | M1 | maps REQ-CBD-004 | 도구 재실행 | **regression-guard** |
| AC-CBD-003 | M1 | maps REQ-CBD-003, REQ-CBD-005 | 문서 검사 | **regression-guard** |
| AC-CBD-004 | M2 | maps REQ-CBD-007 | 뮤턴트 (생존) | release-blocking |
| AC-CBD-005 | M2 | maps REQ-CBD-006, REQ-CBD-007 | 뮤턴트 (사망) | release-blocking |
| AC-CBD-006 | M2 | maps REQ-CBD-008 | 회귀 테스트 | release-blocking |
| AC-CBD-007 | M2 | maps REQ-CBD-008 | 회귀 테스트 + `-list` 로그 보존 | release-blocking |
| AC-CBD-008 | M3 | maps REQ-CBD-009 | 열거 대조 + 로그 존재 | release-blocking |
| AC-CBD-009 | M3 | maps REQ-CBD-011 | 로그 + 후속 카드 요청 기록 | release-blocking |
| AC-CBD-010 | M3 | maps REQ-CBD-010 | 뮤턴트 (사망) + diff 부재 | release-blocking |
| AC-CBD-011 | 전역 | maps REQ-CBD-013 | 공허성 가드 | release-blocking |
| AC-CBD-012 | 전역 | maps REQ-CBD-014 | 위생 검사 | release-blocking |
| AC-CBD-013 | 전역 | maps REQ-CBD-015 | 측정 방식 검사 | release-blocking |
| AC-CBD-014 | 전역 | maps REQ-CBD-012 | 증거 형태 검사 | release-blocking |
| AC-CBD-015 | 전역 | maps REQ-CBD-016 | 추적성 검사 | release-blocking |

### 분류 주석 — M1 의 3건은 regression-guard 다

`AC-CBD-001` / `AC-CBD-002` / `AC-CBD-003` 은 **저작 시점에 이미 초록**이다. M1 은 plan 이전에 끝난
스윕 산출물을 확정하는 마일스톤이므로, 뒤집을 RED-now 셀이 구조적으로 존재하지 않는다 — 이 3건의
시작 관측은 현재 트리에서 재실행해 빨갛게 만들 수 없다.

`verification-completeness.md` §2.1 의 **undecidable disposition** 을 적용한다: 재현 불가능한 RED 를
가진 기준은 release-blocking 자격을 잃고 **regression-guard 로 분류되며, 통과로 기록되지 않는다.**
버리지는 않는다 — `AC-CBD-002` 는 스캐너가 깨지거나 트리가 바뀌면 실제로 빨개지는 진짜 회귀 가드이고,
`AC-CBD-001` / `AC-CBD-003` 은 판별식 문서가 훼손되면 빨개진다. 즉 이 3건은 "이 작업이 무언가를
해냈다" 를 주장하지 않고 "이미 선 것이 무너지지 않았다" 만 주장한다.

`AC-CBD-004` 이하는 진짜 RED-now 를 갖는다 — `AC-CBD-004` 의 뮤턴트 생존이 그 시작 관측이고,
`AC-CBD-005` 가 green path 다.

---

## M1 — 판별식과 스윕

> 이 절의 3건은 **regression-guard** 다(위 분류 주석). 통과로 기록하지 않으며, 깨질 때만 의미가 있다.

### AC-CBD-001 — 판별식이 네 축으로 기록된다

**Given** run-phase 가 시작되고 `research.md` 가 SPEC 디렉터리에 존재하며,
**When** 판별식 절을 읽으면,
**Then** (a) 실물이 컨텍스트를 보는가, (b) 대역이 보는가, (c) `(a) ∧ ¬(b)`,
(d) 대역이 컨텍스트를 보는 유일한 층인가 — 네 축이 각각 판정 방법과 함께 기술되어 있고,
"수리 대상 = `(c) ∧ (d)`" 가 명시되어 있으며, (d) 축의 근거로 M1 검출 사례가 인용되어 있다.

### AC-CBD-002 — 스캐너가 같은 트리에서 같은 집계를 낸다

**Given** 트리가 기준 커밋에 있고 작업 트리가 깨끗하며(`git diff --stat` 무출력),
**When** `go run .moai/reports/t539/ctxsweep/main.go internal` 을 실행하면,
**Then** 집계가 `method blank=127 / no=26 / yes=29`(합 182),
`funclit blank=196 / no=52 / yes=12`(합 260)로 재현되고,
스캐너 파일은 `.moai/reports/t539/ctxsweep/` 아래에 있으며 `internal/` 이하에는 존재하지 않는다
(부재는 `/usr/bin/grep -rl ctxsweep internal/` 무출력으로 확인).

### AC-CBD-003 — 눈멂이 옳은 대역이 근거와 함께 분리된다

**Given** `research.md` 의 후보 표를 읽고,
**When** "옳게 눈멂" 으로 분류된 항목을 확인하면,
**Then** 각 항목에 (a) 판정 근거(실물의 컨텍스트 참조 건수 또는 참조 0건 사실)가 붙어 있고,
`internal/goal/*.go` 러너 대역과 hook 핸들러 대역이 그 분류에 들어 있으며,
위임 예외(`teeGLMDoer`)가 "기계 스캔은 안 본다이나 충실함" 으로 별도 표기되어 있다.

---

## M2 — GLM audit 경로

### AC-CBD-004 — 수리 전, 뮤턴트가 생존한다 (RED-now 수립)

**Given** 가드 테스트를 아직 추가하지 않은 상태에서
`internal/cli/mcp_glm.go:305` 의 `http.NewRequestWithContext(ctx, …)` 를 `http.NewRequest(…)` 로 변이하고,
**When** `go test ./internal/cli/ -run 'GLM|Audit|Converg' -count=1 -timeout 1800s` 를 실행하면,
**Then** 결과가 `ok` 이고(= 뮤턴트 생존 = 현재 가드 부재),
같은 셀렉터의 `-list` 건수가 0보다 크게 기록되며,
되돌림 후 `git diff --stat` 이 무출력이다.

### AC-CBD-005 — 수리 후, 같은 뮤턴트가 죽는다

**Given** audit 경로용 충실한 대역과 취소 가드 테스트가 들어간 상태에서
AC-CBD-004 와 **동일한** 뮤턴트를 재주입하고,
**When** 같은 셀렉터로 테스트를 실행하면,
**Then** 결과가 `FAIL` 이며 실패한 테스트 이름에 새 가드 테스트가 포함되고,
뮤턴트를 되돌린 뒤 같은 셀렉터가 다시 `ok` 가 되며 `git diff --stat` 이 무출력이다.

### AC-CBD-006 — task 경로 가드가 초록을 유지한다

**Given** M2 수리가 적용된 트리에서,
**When** `go test ./internal/cli/ -run 'TestGLMTask' -count=1 -timeout 1800s` 를 실행하면,
**Then** 결과가 `ok` 이고, `-list` 건수가 수리 전과 같거나 크며,
`internal/cli/glm_task_bg_context_test.go` 의 기존 단언이 수정되지 않았다
(`git diff` 에 해당 파일의 단언 변경이 없다).

### AC-CBD-007 — 나머지 GLM 테스트의 의미가 바뀌지 않는다

**Given** M2 수리가 적용된 트리에서,
**When** `-run GLM` / `-run Audit` / `-run Converg` 를 각각 실행하고 `-list` 건수를 기록하면,
**Then** 세 셀렉터 모두 `ok` 이고,
`-list` 건수가 기준선(각각 223 / 99 / 32)보다 **줄지 않으며**,
`stubGLMDoer` 타입 자체를 컨텍스트 인식형으로 바꾼 diff 가 없다
(눈먼 대역을 쓰는 다른 테스트의 의미를 바꾸지 않았다는 증거),
**그리고** 세 셀렉터의 `-list` 출력이
`.moai/reports/t539/list-GLM.log` / `list-Audit.log` / `list-Converg.log` 로 보존되어
건수가 파일에서 재유도된다
(`/usr/bin/grep -c '^Test' .moai/reports/t539/list-<셀렉터>.log`).

> **plan 단계 간극(소급 보강 대상)**: 223 / 99 / 32 는 plan 단계에서 본문 숫자로만 기록됐고
> `-list` 출력을 담은 파일이 없다(`m2-run-*.log` 3개는 `ok …` 한 줄뿐). 이 AC 의 로그 보존 조항이
> 그 간극을 run-phase 에서 닫는다 — plan-audit D8.

---

## M3 — 남은 다섯 후보

> **수리의 소속(운영자 결정 2026-09-08)**: 이 카드는 다섯 후보군을 **측정만** 한다.
> 수리하는 자리는 M2 의 GLM audit 경로 하나뿐이며, 그 밖의 생존은 로그와 후속 카드 요청으로 남긴다.

### AC-CBD-008 — 다섯 후보군의 84건 전부에 뮤턴트 판정이 있다

**Given** M3 가 끝난 시점에,
**When** `.moai/reports/t539/` 의 뮤턴트 로그를 열거하고 `plan.md` §F M3 의 열거와 대조하면,
**Then** 다섯 후보군 각각에 대해 뮤턴트 로그 파일이 하나 이상 존재하고,
각 로그에 주입 지점(`파일:줄`), 변이 내용, 실행 셀렉터, `-list` 건수, 판정(생존/사망)이 담겨 있으며,
`plan.md` §F M3 이 `파일:줄` 로 열거한 **84건 전부**가 어느 후보군의 뮤턴트 판정 아래 귀속되어
**미측정으로 남은 항목이 0건**이다.

**반증 가능성**: 이 종결 조건은 열거된 84건이라는 **닫힌 집합** 위에서만 성립한다.
집합의 원소가 확정되지 않으면 여집합이 비었다는 주장은 어떤 관측으로도 거짓이 될 수 없고,
그것은 이 SPEC 이 `plan.md` §G 에서 스스로 금지한 "공허한 초록" 이다.
따라서 판정 절차는 열거를 먼저 고정한다 — `plan.md` §F M3 의 항목 수를 세어 84 임을 확인한 뒤,
각 항목이 어느 로그에 귀속되는지 대조한다. 84 가 아니거나 귀속되지 않은 항목이 있으면 **FAIL** 이다.

### AC-CBD-009 — 생존한 후보는 로그와 후속 카드 요청으로 남는다

**Given** 어떤 후보의 뮤턴트가 생존한 것으로 기록되고(= 그 자리의 결함이 가려진다),
**When** 그 후보에 대한 최종 트리 상태와 `verdict.md` 를 확인하면,
**Then** 해당 대역 타입 정의에 대한 **수리 diff 가 없고**(이 카드는 M2 밖을 수리하지 않는다),
그 후보의 뮤턴트 로그가 `.moai/reports/t539/` 에 실재하며,
`verdict.md` 에 해당 패키지의 **후속 카드 요청**이 — 대상 패키지, 생존한 뮤턴트의 주입 지점,
로그 경로와 함께 — 한 항목으로 기록되어 있다.

> M2 의 GLM audit 경로는 이 AC 가 아니라 `AC-CBD-004` → `AC-CBD-005` 의 삼단(생존 → 수리 → 사망)이
> 판정한다. 두 자리를 섞지 않는다.

### AC-CBD-010 — 검출된 후보는 수리 없이 닫힌다

**Given** 어떤 후보의 뮤턴트가 검출된 것(= 사망)으로 기록되고,
**When** 그 후보에 대한 최종 트리 상태를 확인하면,
**Then** 해당 대역 타입 정의에 대한 수리 diff 가 **없고**,
`research.md` 또는 `verdict.md` 에 "옳게 눈멂 / 위층이 컨텍스트를 본다" 판정과
그 판정을 세운 로그 경로가 기록되어 있다.

---

## 전역 검증 규율

### AC-CBD-011 — 공허성 가드: 모든 셀렉터의 `-list` 건수가 기록된다

**Given** `verdict.md` 가 작성되고,
**When** 그 안의 모든 `go test … -run <셀렉터>` 인용을 열거하면,
**Then** 각 인용마다 같은 셀렉터의 `-list` 건수가 함께 적혀 있고,
그 값이 전부 0보다 크며, 0건 셀렉터를 근거로 삼은 판정이 하나도 없다.

### AC-CBD-012 — 뮤턴트 위생: 트리에 잔재가 없다

**Given** 각 마일스톤이 끝난 시점에,
**When** `git diff --stat` 과 `git status --short` 를 실행하면,
**Then** 뮤턴트로 변이했던 프로덕션 파일이 변경 목록에 없고,
`git log --oneline` 의 어느 커밋도 뮤턴트를 담고 있지 않다.

### AC-CBD-013 — 부재 주장은 `/usr/bin/grep` 으로 측정된다

**Given** `verdict.md` 또는 `research.md` 에 "이 토큰/픽스처가 없다" 는 주장이 있고,
**When** 그 주장의 근거 명령을 확인하면,
**Then** 명령이 `/usr/bin/grep` 을 절대 경로로 호출하거나 `-a` 를 붙였고,
셸의 맨 `grep` 만으로 세운 부재 주장이 0건이다.

### AC-CBD-014 — 채택 증거가 커버리지가 아니다

**Given** `verdict.md` 의 각 수리 항목에 대해,
**When** 그 항목의 채택 근거를 확인하면,
**Then** 근거가 "뮤턴트 생존 → 수리 → 뮤턴트 사망" 삼단 측정이며,
커버리지 수치만으로 채택을 주장한 항목이 0건이다.

### AC-CBD-015 — 추적성 (maps REQ-CBD-016)

**Given** run-phase 가 끝난 시점에,
**When** `git log --format=%s <base>..HEAD` 를 실행하고 증거 경로를 확인하면,
**Then** 모든 커밋 제목에 `t539` 가 들어 있고,
`.moai/reports/t539/verdict.md` 가 존재하며, 그 안에서 인용된 로그 경로가 전부 실재한다.

---

## Definition of Done

- [ ] AC-CBD-004 ~ 015 전부 PASS, 각 AC 에 실행 명령과 관측 출력이 귀속되어 있다
- [ ] AC-CBD-001 ~ 003 은 regression-guard 로서 **깨지지 않았음**만 확인한다 (통과로 기록하지 않는다)
- [ ] `go vet` 이 건드린 패키지 전부에서 통과
- [ ] `go test ./internal/cli/... -count=1 -timeout 1800s` 초록 (+ M3 가 건드린 패키지 각각)
- [ ] 프로덕션 코드 변경이 M2 에서 0건임을 `git diff --stat` 로 확인 (가드의 부재를 고치는 카드이지
      프로덕션 전파를 다시 쓰는 카드가 아니다)
- [x] `plan.md` §F M3 의 미해소 질문(수리의 소속)이 해소됨 — 운영자 결정 2026-09-08,
      "측정만 이 카드, 수리는 후속 카드". `plan.md` §F "수리의 소속 — 해소됨" 참조.
      해소 게이트 마커는 산출물 전체에서 0건이다.
- [ ] M2 밖에서 생존한 뮤턴트마다 후속 카드 요청이 `verdict.md` 에 한 항목으로 기록됨 (AC-CBD-009)
- [ ] `-list` 출력이 셀렉터별 로그 파일로 `.moai/reports/t539/` 에 보존됨 (AC-CBD-007)
- [ ] `.moai/reports/t539/verdict.md` 에 5절 형식(Claim / Evidence / Baseline-attribution / Gaps / Residual-risk)으로 기록
