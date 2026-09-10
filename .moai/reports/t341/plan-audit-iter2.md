# SPEC 감사 보고 — SPEC-SELECTOR-CENSUS-001 · iter-2

카드 **t341** · 반복 **2/2** (Tier M 상한, 마지막 허용 회차) · 워크트리 `.claude/worktrees/t341` · 브랜치 `WT-selector-census`
감사 트리: **`a6bbbf82b`** (이 세션 `git rev-parse --short HEAD`, 2026-08-29)

**판정: PASS-WITH-DEBT**
**총점 0.894** · Tier M 임계 **0.80** · iter-1 대비 **+0.084** (0.81 → 0.894, 단조 상승 — 점수 회귀 STOP 조항 미발화)

저자의 추론 맥락은 M1 문맥 격리에 따라 무시했다. 이 회차는 **열거된 결함 델타**로 범위를 좁혔다: D1~D8 의 종결 여부, 새로 추가·재작성된 기준(AC-SEC-000·003·006·007)에 대한 자기적용 탐침, 그리고 iter-1 에서 건전하다고 판정한 항목의 회귀 확인.

---

## 0. 독립 상태 확인 (감사 시작 시점)

전달받은 baseline 을 이 세션에서 다시 쟀다.

```
$ git status --porcelain -- .moai/specs/SPEC-TODO-SQLITE-001/acceptance.md
(출력 없음)
$ git diff --name-only HEAD
(출력 없음)
$ git status --porcelain
?? .moai/reports/t341/
?? .moai/specs/SPEC-SELECTOR-CENSUS-001/
```

**결속 파일에 살아남은 편집은 없다.** 저자가 D1 측정 중 편집했다 되돌린 흔적은 작업트리·인덱스 어디에도 남지 않았다. 추적 파일 수정 0건.

---

## 1. 필수 통과 기준 (M5) — 7건 전부 PASS

| # | 기준 | 판정 | 이 회차에 실행한 명령과 출력 |
|---|---|---|---|
| MP-1 | REQ 번호 일관성 | **PASS** | `grep -n '^\*\*REQ-SEC-' spec.md` → `155:REQ-SEC-001 … 167:REQ-SEC-007` 연속 7건, 결번·중복·자릿수 흔들림 없음. AC 는 `AC-SEC-000 … 007` 연속 8건 |
| MP-2 | GEARS 준수 (**요구층 한정**) | **PASS** | REQ 7건 문면 재판독: 001 Unwanted(`…기록하지 않는다`), 002 When, 003 Ubiquitous, 004 When, 005 Ubiquitous, 006 Where, 007 Ubiquitous. **판정은 `REQ-XXX` 층에 대해서만 내렸다** — AC 8건의 Given-When-Then 은 검증층 정형이라 그룹 4 에서 평가했다 |
| MP-3 | 프론트매터 유효성 | **PASS** | `sed -n '2,15p' spec.md \| cut -d: -f1` → `id title version status created updated author priority phase module lifecycle tags tier` — 정규 12필드 전부 + 선택 `tier`. snake_case 별칭 0건 |
| MP-4 | 언어 중립성 | **PASS** | 러너 taxonomy 를 기존 `testCommandSignatures` 에서 물려받고 어느 하나를 PRIMARY 로 두지 않는다. 아래 D9 는 그 목록의 **커버리지** 문제이지 편향 문제가 아니다 |
| MP-5 | D7 교차-SPEC 화해 | **PASS** | `grep '^status:' .moai/specs/SPEC-TODO-SQLITE-001/spec.md` → `status: completed`. retired/superseded/archived 아님 → BLOCKING 없음 |
| MP-6 | D8 크로스플랫폼 규율 | **PASS** | `grep -c 'syscall'` → 네 산출물 모두 `0` → 자동 통과 |
| MP-7 | 해명 게이트 | **PASS** | `grep -rn 'NEEDS CLARIFICATION' .moai/specs/SPEC-SELECTOR-CENSUS-001/` → `rc=1`, 0건 |

**방화벽 미발화.**

---

## 2. 결함 델타 — D1~D8 종결 판정

### D1 (critical) — **종결.** 두 검사가 실제로 상보적이며 빠짐이 없다

저자는 형식 A 가 오늘 비어 있는 것이 커밋 0건 때문이라고 신고하고, 대역 경로로 비공허성을 보였다. **대역이 충분한지를 읽지 않고, 진짜 경로와 격리 저장소 두 곳에서 다시 쟀다.**

먼저, 형식 1 을 **결속 파일 그 자체**에 대해, 그 파일을 건드린 커밋이 들어 있는 범위로 실행했다 — 대역이 아니라 실물이다:

```
$ git log --format=%h -1 -- .moai/specs/SPEC-TODO-SQLITE-001/acceptance.md
3030df58b
$ git diff --name-only 3030df58b~1...3030df58b -- .moai/specs/SPEC-TODO-SQLITE-001/acceptance.md
.moai/specs/SPEC-TODO-SQLITE-001/acceptance.md          ← 출력한다
```

다음으로, 세 편집 상태 전부에서 두 형식이 각각 무엇을 보는지 격리 저장소(`/tmp/t341probe`, 이 워크트리를 건드리지 않음)에서 뒤집어 봤다:

| 편집 상태 | `git diff --name-only develop...HEAD -- <path>` | `git status --porcelain -- <path>` |
|---|---|---|
| 작업트리만 수정 | `[]` — **못 본다** | `[ M sub/coupled.md]` — **잡는다** |
| 스테이징됨 | `[]` — **못 본다** | `[M  sub/coupled.md]` — **잡는다** |
| 브랜치에 커밋됨 | `[sub/coupled.md]` — **잡는다** | `[]` — 못 본다 |

**세 상태 전부가 최소 한 검사에 걸린다.** 두 검사는 서로의 사각을 정확히 덮으며 빈틈이 없다. 대역 시연은 대체물로 충분했고, 실물 경로 확인으로 그 결론이 독립적으로 재현됐다. iter-1 D1 이 지적한 구조적 상시-초록은 사라졌다.

### D2 (critical) — **제기된 축에서는 종결. 다만 형제 축이 열려 있다 (→ D9)**

지시받은 탐침을 실제로 돌렸다 — 여덟 기준을 전부 만족하면서 진짜 pass 를 죽이는 뮤턴트를 쓰려 시도했다.

**텍스트 휴리스틱 축에서는 실패했다 — 즉 수정이 통했다.** iter-1 에서 통했던 뮤턴트(`" passed"` 절 삭제)는 이제 표본 (c)(d)(e) 에서 붉어진다. `"0 passed"` 부분 문자열 뮤턴트도 표본 (e) 에서 죽는다 — 그 충돌은 저자가 실측했고 나도 재현했다(`printf 'Tests:       10 passed, 10 total\n' | grep -c '0 passed'` → `1`). 저자가 실측했다고 적은 pytest 문자열도 재현된다(`pytest -q` → `3 passed in 0.05s`).

미측정으로 표시된 두 문자열(cargo, jest/vitest)이 M1 확인 대기인 것은 `spec.md` §5 와 같은 규율이며 받아들인다.

**그러나 exit-code 축은 여전히 비어 있다.** 상세는 D9.

### D3 (major) — **종결.** 탈출구가 사라졌고 근거가 실측이다

`acceptance.md:177` 이 조건 (2) 를 `diff` rc 0 · 출력 0줄로 못 박고 *"판단 탈출구 없음: 사유를 적으면 통과하는 형태가 아니다"* 를 명시한다. 근거를 다시 쟀다:

```
$ diff .claude/rules/moai/development/verification-completeness.md \
       internal/template/templates/.claude/rules/moai/development/verification-completeness.md
diff-exit=0        ← 오늘도 바이트 동일
```

중립화 차이가 실제로 필요해지면 blocker 로 보고한다는 처리도 `acceptance.md:179` · `plan.md:76` 양쪽에 있다. 우회로가 아니라 정지 조건이다.

### D4 (major) — **종결.** 저자의 추론이 참이고, 뮤턴트가 죽는다

저자가 근거로 든 코드 사실을 확인했다:

```
$ grep -n 'return true, false, false' internal/hook/evidence_writer.go
79:	return true, false, false
```

`classifyTestCommand` 은 인식 가능한 신호가 없을 때 `isTest=true, isPass=false, isFail=false` 를 반환한다. 따라서 `map[sig]""` 뮤턴트의 빈 문자열은 `isPass=false` 를 내고, **`isPass=false` 만 요구하는 판본에서는 초록이었다** — 저자의 진단이 정확하다. `acceptance.md:158` 이 더한 `detectZeroExecution(항목) == true` 절은 빈 문자열에 0-실행 토큰이 없으므로 발화하지 않아 그 뮤턴트를 붉게 만든다. 조건 (2) 는 또한 corpus 가 AC-SEC-004 와 **같은 하나의 변수**임을 못 박아(`:151`) 사본 분기도 막는다. `plan.md:51`·`:67` 이 같은 제약을 계획층에 반복한다.

### D5 (major) — **종결. M0 은 이제 실제로 게이트다**

AC-SEC-000 의 RED-now 를 재실행했다:

```
$ ls .moai/reports/t341/live-payload.json
ls: .moai/reports/t341/live-payload.json: No such file or directory
ls-exit=1
```

핵심은 이것이다 — **M0 을 건너뛰면 조건 (1) 이 rc≠0 이 되어 완료 판정이 붉어진다.** 종전에는 일곱 기준이 전부 합성 fixture 위에서 돌아 건너뛰어도 전부 초록이었다. 조용한 전제가 시끄러운 전제로 바뀌었고, 그것이 D5 가 요구한 전부다. blocker 분기(조건 3)도 기준 안에서 이진 판정되므로 "stdout 이 안 실리는데 M1 을 그냥 진행" 이 통과하는 경로가 없다.

저자가 스스로 신고한 한계 — 캡처된 payload 와 손으로 쓴 payload 를 기계적으로 가르지 못한다 — 는 참이며, **부채로 명시 신고된 것을 그대로 인정한다.** 그 구분을 기계화하려면 별도 계측기가 필요하고 이 카드 범위 밖이다.

### D6 (minor) — **종결.** 좌표가 맞다

```
$ sed -n '223,224p' internal/hook/evidence_writer.go
	if hasPrecisePass {
		return true, false, true
$ grep -n 'func buildBashRecord\|rec.IsTestPass = isPass' internal/hook/evidence_writer.go
309:func buildBashRecord(...)
328:	rec.IsTestPass = isPass
```

`spec.md:50` 은 이제 `분기 :223, 반환 :224`, `spec.md:56` 은 `:309-330, 대입은 :328` 로 적는다. 둘 다 정확하다.

### D7 · D8 (minor, optional) — **의도적으로 열어 둠. 이의 없음**

D7(요구 문면의 구현 표면 이름)과 D8(AC-SEC-003 의 RED-now 부재)은 iter-1 에서 optional 로 분류했고 저자가 열어 두기로 했다. D8 에 대해서는 저자가 오히려 **오늘 잰 두 사실**(pytest 진짜 pass 의 마커 부재, `10 passed` 의 `0 passed` 포함)을 칸에 적어 "시작 관측이 없는 약속" 을 부분적으로 해소했다 — 요구되지 않은 개선이다. 재론하지 않는다.

---

## 3. 자기적용 탐침 — 새로 추가·재작성된 기준에 같은 렌즈를 다시 걸었다

iter-1 의 D1 은 **공허한 0-히트 검사를 잡으려는 SPEC 안의 공허한 0-히트 검사**였다. 그래서 이번에 추가된 것들을 읽지 않고 각각 뮤턴트를 직접 썼다.

### D9 — exit-code 축의 진짜 pass 가 여덟 기준 어디에도 고정돼 있지 않다 · severity: **major** · class: **blocking** · **신규**

위치: `acceptance.md:88-104` (AC-SEC-003 표본 집합) · `internal/hook/evidence_writer.go:69`

**뮤턴트**: 0-실행 거부권을 옳게 넣으면서, `deriveFromExitCode` 의 pass 경로를 좁힌다 — 예컨대 "텍스트에 인식 가능한 실행 수 근거가 없으면 0-실행으로 본다" 는 형태의 거부권. `plan.md:49` 가 M1 에 지시하는 삽입 위치가 정확히 그 호출 지점 **바로 앞**이다.

**이 뮤턴트는 여덟 기준을 전부 통과한다.** AC-SEC-003 의 표본 다섯은 **전부 텍스트 마커를 담고 있어서**, exit-code 경로가 죽어도 텍스트 휴리스틱으로 살아남아 열 개 payload 가 모두 초록으로 남는다. 나머지 일곱 기준은 발화 방향이거나 무관하다.

**그리고 그 뮤턴트가 죽이는 진짜 pass 는 가설이 아니라 실측이다.** `testCommandSignatures` 에 있는 `npm test` 를 통해 도달하는 러너 하나의 진짜 pass 출력을 실제로 만들어 네 마커를 전부 셌다:

```
$ node --test /tmp/t341_nt.test.js ; echo "node-exit=$?"
TAP version 13
ok 1 - a
1..1
# tests 1
# pass 1
# fail 0
node-exit=0

$ grep -c "ok  <TAB>" out          → 0
$ grep -c "ok <TAB>" out           → 0
$ grep -c "test result: ok" out    → 0
$ grep -c " passed" out            → 0
$ grep -c " failed" out            → 0
```

네 pass 마커가 **전부 부재**하므로 `deriveFromOutputText` 는 `ok=false` 를 반환한다(`evidence_writer.go:236`). 따라서 이 진짜 pass 가 오늘 `isPass=true` 로 기록되는 유일한 근거는 **`exit_code == 0`** 이다(`:69` → `:163`). 뮤턴트가 그 경로를 좁히면 이 실행은 조용히 `isPass=false` 가 된다 — **REQ-SEC-005 위반이며, 채택된 여덟 기준 중 어느 것도 잡지 못한다.**

부수적으로 이 측정은 `acceptance.md:86` 의 문장 — *"`npm`·`pnpm`·`yarn` 은 그 아래 러너의 출력을 그대로 흘리는 래퍼이므로 (c)(d)(e) 표본에 흡수된다"* — 이 **일반 주장으로는 거짓**임도 보인다. `npm test` 는 node 내장 러너로도 흐르며 그 출력 형태는 (c)(d)(e) 중 어느 것도 아니다.

**요구되는 수정** (한 줄): AC-SEC-003 에 여섯 번째 표본을 더한다 — **네 마커를 하나도 담지 않는 진짜 pass 텍스트 + `{"exit_code": 0}`** 를 담은 payload 가 `isPass=true` 로 남을 것. 그러면 위 뮤턴트가 AC-SEC-003 에서 붉어지고, exit-code 축의 비발화 방향이 처음으로 고정된다. `plan.md` F 절에 대응 안티패턴 한 줄(“0-실행 거부권을 ‘텍스트에 실행 수 근거가 없음’ 으로 구현하기”)을 함께 두면 M1 이 그 형태로 흘러가는 것을 계획층에서도 막는다.

### D10 — AC-SEC-000 조건 (3) 이 D1 이 고친 그 형태를 한 검사만으로 되쓴다 · severity: **minor** · class: **optional**

위치: `acceptance.md:39` · `acceptance.md:193` (DoD)

DoD 의 결속 검사는 D1 을 닫으면서 **두 형식**(`diff` 삼중점 + `status --porcelain`)을 쌍으로 요구한다. 그런데 같은 회차에 새로 쓴 AC-SEC-000 조건 (3) 은 `internal/hook/evidence_writer.go` 에 대해 **`git diff --name-only origin/develop...HEAD` 한 형식만** 쓴다. §2 D1 의 격리 측정표가 그대로 적용된다 — 그 형식은 작업트리·스테이징 편집을 보지 못한다.

**그럼에도 blocking 으로 올리지 않는 이유를 밝힌다.** 조건 (3) 이 겨누는 해악은 "전제가 반증됐는데 M1 설계가 **착지하는 것**" 이고, 착지에는 커밋이 필요하다. 커밋되지 않은 편집은 아무것도 실어 보내지 못하므로, 이 기준의 목적에 한해 커밋 범위가 올바른 범위다. D1 이 blocking 이었던 것은 종전 형식이 **커밋된 편집까지** 놓쳤기 때문이며, 여기서는 그 지배적 경우가 잡힌다. 대칭을 위해 `git status --porcelain -- internal/hook/evidence_writer.go` 를 한 줄 더하는 것은 값싸고 해롭지 않으나, 요구하지는 않는다.

### D11 — 정밀 마커 `"ok \t"`(한 칸) 이 여덟 기준 어디에도 고정돼 있지 않다 · severity: **minor** · class: **optional**

위치: `internal/hook/evidence_writer.go:218` · `acceptance.md:92`

오늘 살아 있는 pass 표면은 네 갈래다(`:217-219` 의 마커 셋 + `:232` 의 카운트 절). AC-SEC-003 표본은 `"ok  \t"`(두 칸, 표본 a), `"test result: ok"`(b), `" passed"`(c·d·e) 를 고정한다. **`"ok \t"`(한 칸, `:218`) 만 어느 표본에도 걸리지 않는다** — 두 칸 형태의 부분 문자열도 아니므로 독립 절이다. 즉 `:218` 을 지우는 뮤턴트는 여덟 기준을 전부 통과한다.

다만 그 뮤턴트가 **진짜 pass 를 죽이는지는 측정되지 않았다.** 이 저장소의 실제 go 출력에서 그 형태는 나오지 않는다:

```
$ go test ./internal/hook/... -count=1 > out
$ grep -c "ok  <TAB>" out   → 11      (두 칸 형태만 나온다)
$ grep -c "ok <TAB>" out    → 0
$ od -c  (kanban 패키지 pass 줄)  →  o  k  ' '  ' '  \t  g  i  t  …
$ grep -rn '"ok \t"' internal/hook/   → evidence_writer.go:218 한 곳뿐 (고정하는 테스트 없음)
```

**어떤 생산자가 그 형태를 내는지 나는 찾지 못했다 — 도달 불가일 수 있다는 것은 가설이다.** 도달 불가라면 이 절을 지워도 해가 없고 결함은 무해하다. 도달 가능하다면 D9 와 같은 종류의 구멍이다. 어느 쪽인지 정하는 것은 M1 이 러너 출력을 실측할 때 공짜로 딸려 오므로, 그때 판정하도록 남긴다.

---

## 4. 회귀 확인 — iter-1 에서 건전하다고 본 것이 약해지지 않았다

지시대로 RED-now 칸을 **두 개 이상** 재실행했다. 네 개 전부 다시 돌렸다.

| 칸 | 실행한 명령 | 관측 출력 | 판정 |
|---|---|---|---|
| AC-SEC-001 토큰 부재 | `grep -rn 'no tests to run\|no test files\|no tests ran\|collected 0 items' internal/hook/` | 출력 0줄, `grep-exit=1` | 오늘도 붉다 ✓ |
| AC-SEC-006 | `grep -c 'testCommandSignatures' internal/hook/evidence_writer_test.go` | `0` | 오늘도 붉다 ✓ |
| AC-SEC-007 | `grep -n 'Evidence footnote (not a rule)' .claude/rules/…/verification-completeness.md` | `45:*Evidence footnote (not a rule):* …` | 오늘도 붉다 ✓ |
| AC-SEC-000 (신규) | `ls .moai/reports/t341/live-payload.json` | `No such file or directory`, `ls-exit=1` | 붉다 ✓ |

**오늘 초록인 칸은 없고, 적힌 사유와 다른 이유로 붉은 칸도 없다.**

**§2.2 순서 주장** — iter-1 에서 이 SPEC 의 가장 값나가는 대목으로 판정했다. 코드에서 재확인했다: `classifyTestCommand` 이 `:69` 에서 `deriveFromExitCode`, `:74` 에서 `deriveFromOutputText` 를 부르고, exit 0 은 `:163` 에서 `return true, false, true` 한다. 순서 제약은 여전히 참이고 약해지지 않았다. (이번 회차의 D9 는 바로 이 대목의 **반대 방향**이 비어 있다는 발견이다 — 순서 주장 자체를 부정하지 않는다.)

**범위 규율 (강조 5)** — 명시된 비목표 셋을 `plan.md` 전체에 대고 다시 훑었다:

```
$ grep -n 't\.Skip\|sg test\|0-히트' .moai/specs/SPEC-SELECTOR-CENSUS-001/plan.md
105:- 범위 밖 세 항목(grep 0-히트 기준, `t.Skip`, `sg test`)을 "겸사겸사" 넣기.
```

유일한 등장이 **금지하는 안티패턴 줄**이다. 재유입 0건. 더해서 iter-1 이 지적한 "DoD 자신이 0-히트 형태를 쓴다" 는 자기적용 실패가 D1 종결로 사라졌으므로, 이 항목은 iter-1 보다 오히려 깨끗해졌다.

---

## 5. 차원별 점수

| 차원 | iter-1 | iter-2 | 대역 | 근거 |
|---|---|---|---|---|
| 명료성 | 0.85 | **0.88** | 0.75↑ | 표본→코드 경로 대응표(`acceptance.md:90-96`)가 각 표본이 무엇을 고정하는지 명시해 해석 여지가 줄었다. AC-SEC-000 의 (a)(b)(c) 질문이 "어느 키에서 읽혔는지" 까지 요구한다. 잔여: AC-SEC-005 의 "안정된 센티널 토큰" 이 여전히 기준 안에 특정돼 있지 않다(`plan.md:59` 가 예시로만 제시) |
| 완전성 | 0.85 | **0.92** | 0.75↑ | iter-1 의 유일한 감점(M0 을 덮는 기준 부재)이 AC-SEC-000 + DoD 한 줄로 닫혔다. 필수 절 전부 존재, `### Out of Scope —` H3 5개가 각각 구체 불릿을 담는다 |
| 검증가능성 | 0.62 | **0.80** | 0.75↑ | D1 종결(격리 측정으로 상보성 입증), D3 종결(탈출구 삭제), D4 종결(`detectZeroExecution` 절), D5 종결(M0 기계화), D2 의 텍스트 축 종결. 잔여 감점: **D9**(exit-code 축 미고정, 실측된 뮤턴트 생존), AC-SEC-005 센티널 미특정, D11 |
| 추적성 | 1.00 | **1.00** | 1.0 | REQ 7건 전부 최소 1개 AC 를 갖는다(001→AC-001·004, 002→AC-000·002, 003→AC-001·004, 004→AC-000·005, 005→AC-003, 006→AC-006, 007→AC-007). AC 8건 전부 실재하는 REQ 를 가리킨다. 신규 AC-SEC-000 도 REQ-SEC-002·004 로 정상 연결. 고아 없음 |

**총점 = 조화평균(0.88, 0.92, 0.80, 1.00) = 0.8942 → 0.894**

조화평균을 쓰는 이유는 iter-1 과 같다(`agent-common-protocol.md` § Skeptical Evaluation Stance) — 낮은 차원 하나가 상쇄되지 않게 한다. **iter-1 대비 +0.084, 단조 상승.** 점수 회귀 STOP 조항은 발화하지 않는다.

---

## 6. 잔여 결함 (심각도순)

| id | 위치 | 요약 | 심각도 | 분류 | 요구되는 수정 |
|---|---|---|---|---|---|
| **D9** | `acceptance.md:88-104` | exit-code 축의 진짜 pass 가 미고정 — 여덟 기준을 전부 만족하는 뮤턴트가 실측된 진짜 pass(`npm test`→node 내장 러너)를 죽인다 | major | **blocking** | AC-SEC-003 에 여섯 번째 표본 추가: 네 마커 부재 텍스트 + `{"exit_code":0}` → `isPass=true` 단언 |
| D10 | `acceptance.md:39` | AC-SEC-000 (3) 이 검사 한 형식만 쓴다(작업트리·스테이징 사각) | minor | optional | 대칭을 위해 `git status --porcelain -- internal/hook/evidence_writer.go` 한 줄 추가 |
| D11 | `evidence_writer.go:218` | 정밀 마커 `"ok \t"`(한 칸)이 어느 기준에도 미고정. 도달 가능성은 **가설** | minor | optional | M1 의 러너 출력 실측 때 도달 여부를 판정하고, 도달 가능하면 표본 추가 |
| D7 | `spec.md:159`·`:165` | 요구 2건이 구현 표면을 이름으로 담음 | minor | optional | 저자가 의도적으로 열어 둠 — 이의 없음 |
| D8 | `acceptance.md:106-113` | AC-SEC-003 에 RED-now 칸 없음(회귀 고정자) | minor | optional | 정직하게 신고돼 있고 이번 회차에 실측 두 건이 보강됨 — 이의 없음 |

`AC-SEC-000` 의 "캡처 payload 와 손으로 쓴 payload 를 가르지 못한다" 는 저자가 기준 본문에 명시 신고한 **부채**이며, 결함으로 계상하지 않는다.

---

## 7. 판정 근거와 다음 단계

**PASS-WITH-DEBT** 로 읽는 이유:

- 필수 통과 7건 전부 PASS — 방화벽 미발화.
- 총점 0.894 ≫ Tier M 임계 0.80, iter-1 대비 단조 상승.
- iter-1 의 blocking 5건 중 **다섯 건 전부** 닫혔고, 그중 셋(D1·D3·D4)은 저자가 든 근거를 읽지 않고 **다시 재어** 참임을 확인했다. D1 은 격리 저장소에서 세 편집 상태를 뒤집어 상보성을 입증했다 — 저자의 대역 시연보다 강한 근거다.
- 남은 blocking 은 **신규 D9 한 건**이며, 기준 문면 한 줄로 닫힌다.

**요구:**

1. **D9 를 M1 진입 전에 닫는 것을 권한다** — 표본 한 개 추가이며, 고치는 자리가 M1 이 손댈 바로 그 호출 지점이라 나중에 발견되면 M1 을 되돌려야 한다. 이 회차가 Tier M 상한(2회)의 마지막이므로 재감사는 없다: 오케스트레이터는 (a) 기준 한 줄 수정 후 진행, 또는 (b) D9 를 문서화된 부채로 수용하고 Kickoff 게이트에서 밝히는 것 중 하나를 택한다.
2. D10·D11·D7·D8 은 **optional** 이다. 이것들을 이유로 수정 라운드를 돌리지 않기를 권한다 — 요구하지 않은 기준을 늘리면 이 SPEC 이 스스로 경계한 과설계가 된다.
3. **Implementation Kickoff Approval 은 이 판정과 무관하게 여전히 필수이며, 감사 PASS 가 그 게이트를 열지 않는다.** 배차문의 `cmd:` 나열도 승인이 아니다.

---

## 8. 증거 무결성

**Claim** — 위 판정과 점수.
**Evidence** — 각 절에 실행한 명령과 그 출력을 함께 실었다. 읽기만으로 세운 것(D11 의 도달 가능성)은 **가설**로 명시했다.
**Baseline-attribution** — 전부 이 워크트리 `a6bbbf82b`(`git rev-parse --short HEAD`, 2026-08-29)에서, 이 실행에서 잰 값이다. 격리 측정은 `/tmp/t341probe` 의 별도 저장소에서 했고 이 트리를 건드리지 않았다.

**Gaps (관측하지 않은 것)**

- cargo · jest/vitest 의 실제 0-실행 및 진짜-pass 출력 문자열은 **실행해 보지 않았다**. `spec.md` §5 가 미검증 전제로 신고했고 M1 몫이다.
- 살아 있는 PostToolUse payload 를 잡아 읽지 않았다 — AC-SEC-000 이 겨누는 그 전제는 나도 관측하지 못했다.
- D9 뮤턴트를 **코드로 구현해 여덟 기준을 실제로 돌려 보지는 않았다.** 생존 판정은 여덟 기준 문면 전수 판독 + 마커 부재 실측 + `:69`/`:163` 소스 확인의 조합이다. 기준이 아직 코드로 존재하지 않으므로 실행 검증은 M1 이후에만 가능하다.
- `"ok \t"` 를 내는 생산자를 찾지 못했다(D11). 부재를 도달 불가의 증거로 읽지 않았다.

**Residual-risk**

- D9 를 부채로 넘기면, M1 이 거부권을 "텍스트에 실행 수 근거 없음" 형태로 구현할 때 여덟 기준이 전부 초록인 채로 `npm test` 계열의 진짜 pass 가 조용히 죽는다. 이 SPEC 이 겨눈 병(조용한 오판정)과 **같은 종류의 병**이 반대 방향으로 생긴다.
- D1 이 못 박은 기준 트리는 `a6bbbf82b` 다. 이 워크트리가 `develop` 을 흡수하면 검사 1 의 `origin/develop...HEAD` 는 **그때의 병합 기준점**으로 다시 재야 한다.
- AC-SEC-005 의 센티널 토큰이 기준에 특정돼 있지 않아, M2 가 고른 토큰과 테스트가 기대하는 토큰이 갈릴 수 있다.

**감사 중 SPEC 산출물 수정 0건.** 다만 회귀 확인용 `go test ./internal/hook/... -count=1` 이 `.moai/specs/SPEC-HOOK-PRETOOL-PERF-001/{baseline,postchange}.md` 두 개를 재작성했다(알려진 부작용). 같은 턴에 `git restore` 로 되돌렸고, `git status --porcelain` 이 감사 시작 시점과 동일하게 미추적 2건(`.moai/reports/t341/`, `.moai/specs/SPEC-SELECTOR-CENSUS-001/`)만 남음을 확인했다. 결속 파일 `SPEC-TODO-SQLITE-001/acceptance.md` 는 빈 출력이다.
