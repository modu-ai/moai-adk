# SPEC 감사 보고 — SPEC-SELECTOR-CENSUS-001

카드 **t341** · 반복 **1/2** (Tier M 상한) · 워크트리 `.claude/worktrees/t341` · 브랜치 `WT-selector-census`
감사 트리: **`a6bbbf82b`** (이 세션 `git rev-parse --short HEAD`, 2026-08-29)

**판정: PASS-WITH-DEBT**
**총점 0.81** · Tier M 임계 **0.80** (`spec-workflow.md` § SPEC Complexity Tier)
**여유 0.006.** 편한 통과가 아니다 — 아래 blocking 5건 중 하나만 더 있었으면 임계 아래였다.

저자의 추론 맥락은 M1 문맥 격리에 따라 무시했다. 판정은 네 산출물과 `discovery.md`, 그리고 이 트리에서 직접 잰 명령의 출력만을 근거로 한다.

---

## 1. 반증 시도 — 무엇을 직접 재었나

이 SPEC은 **검증 계측기를 짓는 SPEC**이라 `verification-completeness.md`가 자기 자신에게 걸린다. 그래서 RED-now 칸을 읽지 않고 **다시 실행**했다.

### 1.1 RED-now 칸 재실행 (전부 오늘도 붉다)

| 칸 | 실행한 명령 | 관측 출력 | 판정 |
|---|---|---|---|
| AC-SEC-001 러너 재현 | `go test ./internal/kanban -run TestT341NoSuchTestXYZ -count=1` | `ok  github.com/modu-ai/moai-adk/internal/kanban 0.440s [no tests to run]` / `exit=0` | 붉다 ✓ |
| AC-SEC-001 토큰 부재 | `grep -rn 'no tests to run\|no test files\|no tests ran\|collected 0 items' internal/hook/` | 출력 0줄, `grep-exit=1` | 붉다 ✓ |
| AC-SEC-006 | `grep -c 'testCommandSignatures' internal/hook/evidence_writer_test.go` | `0` | 붉다 ✓ |
| AC-SEC-007 | `grep -n 'Evidence footnote (not a rule)' .claude/rules/moai/development/verification-completeness.md` | `45:*Evidence footnote (not a rule):* a selector that swept nothing still prints ok …` | 붉다 ✓ |

**오늘 초록인 칸은 없었고, 적힌 사유와 다른 이유로 붉은 칸도 없었다.**

### 1.2 §2.2의 순서 주장 — 이 SPEC의 가장 값나가는 대목이며, 참이다

`discovery.md`는 이음매를 `deriveFromOutputText`로 지목했다. `spec.md` §2.2는 그것만으로는 **살아 있는 훅 경로에서 아무 일도 일어나지 않는다**고 고쳐 읽었다. 소스로 확인했다.

```
$ grep -n 'func classifyTestCommand\|deriveFromExitCode(result)\|deriveFromOutputText(result)' internal/hook/evidence_writer.go
56:func classifyTestCommand(command string, result []byte) (isTest, isPass, isFail bool) {
69:	if pass, fail, ok := deriveFromExitCode(result); ok {
74:	if pass, fail, ok := deriveFromOutputText(result); ok {
```

그리고 `deriveFromExitCode` 본문(`:151-175`)은 `exit_code == 0` 에서 `return true, false, true`(`:163`) 한다. 0-매치 실행의 종료코드가 정확히 0 임은 §1.1의 `exit=0` 으로 실측됐다. **exit-code 신호를 담은 payload 에서는 텍스트가 읽히지도 않는다** — §2.2의 주장 그대로다. 이 한 줄이 `discovery.md` 위에서 이 판이 실제로 더 읽어낸 것이고, 감사 대상 중 가장 정확한 부분이다.

### 1.3 인용 좌표 정확도

| 인용 | 실제 | 판정 |
|---|---|---|
| `evidence_writer.go:217` 정밀 pass 마커 | `217: hasPrecisePass := strings.Contains(text, "ok  \t") \|\|` | 정확 |
| `evidence_writer.go:223` `return true, false, true` | `223: if hasPrecisePass {` / `224: return true, false, true` | 1줄 어긋남 (D6) |
| `evidence_writer.go:69` / `:74` 분기 순서 | 정확 | 정확 |
| `evidence_writer.go:229` / `:232` 카운트 휴리스틱 | `228 if " failed"` → `229 return`, `231 if " passed"` → `232 return` | 정확 |
| `buildBashRecord :296-330` | `309:func buildBashRecord`, `328: rec.IsTestPass = isPass` | 시작점 어긋남 (D6) |
| `post_tool.go:219-224` "never alters HookOutput" | `222: // Best-effort, additive — never blocks, never alters HookOutput.` | 정확 |
| `post_tool.go:239` navigator 자문 선례 | `239: systemMessage = emitNavigatorDetectAdvisory(input, ndResult, systemMessage)` | 정확 |
| `post_tool.go:280` SystemMessage 반환 | `280: SystemMessage: systemMessage,` | 정확 |
| `types.go:366` 필드 계약 | `366: SystemMessage  string \`json:"systemMessage,omitempty"\`  // Warning message shown to user` | 정확 |

§3.3의 채널 선택 근거 세 가지(a·b·c)는 전부 실재한다. 다만 (b) navigator 선례는 **Write/Edit/NotebookEdit 분기**에 있고 같은 주석이 `Bash 는 dispatch 하지 않는다(AC-NS2-001b negative)`를 명시한다 — 상충은 아니지만 M2가 Bash 분기에 **새 경로를 만드는** 일임을 뜻한다. 선례가 있다는 §3.3의 서술은 여전히 참이다.

---

## 2. 필수 통과 기준 (M5)

| # | 기준 | 판정 | 근거 |
|---|---|---|---|
| MP-1 | REQ 번호 일관성 | **PASS** | REQ-SEC-001…007 연속, 결번·중복·자릿수 흔들림 없음 (`spec.md:153-165`) |
| MP-2 | GEARS 준수 (**요구층에 한정**) | **PASS** | 7건 전부 다섯 패턴 중 하나: 001 Unwanted(`…기록하지 않는다`), 002 When, 003 Ubiquitous, 004 When, 005 Ubiquitous, 006 Where, 007 Ubiquitous. AC 7건은 검증층이라 Given-When-Then 이 정형이며 여기서 평가하지 않았다(그룹 4 소관) |
| MP-3 | 프론트매터 유효성 | **PASS** | `spec.md:2-15` 에 정규 12필드 전부 + 선택 `tier: M`. snake_case 별칭 0건 |
| MP-4 | 언어 중립성 | **PASS** | 러너 taxonomy를 **기존** `testCommandSignatures`(`evidence_writer.go:25-32`, 9개 signature)에서 그대로 물려받고 어느 하나를 PRIMARY로 두지 않는다. 그 목록이 16개 언어 중 4계열만 덮는 것은 선재 상태이며 이 SPEC 범위 밖이다. REQ-SEC-006 은 오히려 그 격차를 **자기교정 방향**으로 돌린다 — 러너가 추가되면 표본 없이는 테스트가 붉어진다 |
| MP-5 | D7 교차-SPEC 화해 | **PASS** | 참조 SPEC은 `SPEC-TODO-SQLITE-001` 하나. `grep '^status:'` → `completed`. retired/superseded/archived 아님 → BLOCKING 없음 |
| MP-6 | D8 크로스플랫폼 규율 | **PASS** | `grep -c 'syscall'` → 네 산출물 모두 `0` → 자동 통과 |
| MP-7 | 해명 게이트 | **PASS** | `grep -rn 'NEEDS CLARIFICATION'` → 0건 (`research.md` 는 Tier M 이라 부재, `plan.md` 는 존재하므로 검증 동사는 실행 가능) |

**필수 통과 7건 전부 PASS. 방화벽은 발화하지 않았다.**

---

## 3. 차원별 점수 (루브릭 기준)

| 차원 | 점수 | 대역 | 근거 |
|---|---|---|---|
| 명료성 | 0.85 | 0.75↑ | 표본을 축자로 싣고 사유를 함께 적어 해석 여지가 좁다. 국소 모호 2건: AC-SEC-005의 "안정된 센티널 토큰"이 어느 토큰인지 기준 안에 없고(plan M2가 예시로만 제시), AC-SEC-004의 러너 문자열이 M1로 미뤄져 있다. 둘 다 주인이 명시돼 있다 |
| 완전성 | 0.85 | 0.75↑ | 필수 절 전부 존재, `### Out of Scope — <주제>` H3 5개가 각각 구체 불릿을 담는다(`spec.md:129-149`). §5가 미검증 전제 3건을 스스로 드러낸 것은 드문 미덕이다. 감점: 계획이 게이트로 세운 M0을 덮는 수락 기준이 하나도 없다(D5) |
| 검증가능성 | 0.62 | 0.50↔0.75 | AC-SEC-005는 이진 판정 불가(센티널 미지정), AC-SEC-007은 "의도된 차이면 사유를 적는다"는 판단 탈출구를 담고, AC-SEC-006은 세는 corpus와 실제로 돌리는 corpus를 묶지 않으며, DoD의 결속 검사는 **오늘 이미 공허하다**(D1, 실측) |
| 추적성 | 1.00 | 1.0 | REQ 7건 전부 최소 1개 AC를 갖고, AC 7건 전부 실재하는 REQ를 가리킨다. 고아 없음 |

**총점 = 조화평균(0.85, 0.85, 0.62, 1.00) = 0.8055 → 0.81**

조화평균을 쓴 이유는 `agent-common-protocol.md` § Skeptical Evaluation Stance 가 명시하기 때문이며, 낮은 차원 하나가 상쇄되지 않게 한다.

---

## 4. 발견된 결함

각 항목은 **실행한 명령과 그 출력**을 함께 적는다. 읽기만으로 세운 것은 그렇게 표시했다.

### D1 — 교차 카드 결속의 [HARD] 제약이 실제로는 강제되지 않는다 · severity: critical · class: **blocking**

위치: `acceptance.md:147` (DoD) · `plan.md:89` (§E5)

`spec.md` §2.4는 `SPEC-TODO-SQLITE-001/acceptance.md:13` 을 **[HARD] 읽기 전용 증거**로 못 박는다. 그것을 지키는 기계는 딱 하나 — DoD의 `git diff --name-only | grep SPEC-TODO-SQLITE-001` → 0줄. **이 검사는 커밋되거나 스테이징된 편집을 보지 못한다.**

실행 (이 워크트리, `a6bbbf82b`):

```
# 1) 손대지 않은 상태
$ git diff --name-only | grep SPEC-TODO-SQLITE-001
grep-exit=1        ← 0줄 = DoD가 읽는 PASS

# 2) 결속 파일을 실제로 편집 + 스테이징
$ echo "" >> .moai/specs/SPEC-TODO-SQLITE-001/acceptance.md
$ git add .moai/specs/SPEC-TODO-SQLITE-001/acceptance.md
$ git diff --name-only | grep SPEC-TODO-SQLITE-001
DoD-grep-exit=1    ← 여전히 0줄 = DoD가 여전히 PASS 를 읽는다
```

(검사 후 `git restore --staged` + `git checkout --` 로 되돌렸고, `git diff HEAD --name-only -- .moai/specs/SPEC-TODO-SQLITE-001/` → 0줄로 원상 확인.)

인수 없는 `git diff` 는 **작업트리 미스테이징 변경만** 본다. M4가 편집하고 커밋하면 — 그것이 실제 시나리오다 — 검사는 언제나 초록이다. 이것은 `verification-completeness.md` §1.2(a) 가 이름 붙인 **구조적 상시-초록 검사**이며, 동시에 이 SPEC이 스스로 범위 밖으로 선언한 **"0-히트 통과 조건을 가진 grep 기반 기준"** 바로 그 형태다(`spec.md:129-132`). 자기가 겨눈 병을 자기 DoD가 앓고 있다.

요구되는 수정: 기준 트리에 대고 재라 — `git diff --name-only a6bbbf82b -- .moai/specs/SPEC-TODO-SQLITE-001/` 가 0줄이거나, `git diff a6bbbf82b --stat -- <path>` 가 빈 출력임을 단언한다. 그리고 **양방향을 관측하라**: 그 파일에 한 바이트를 넣었을 때 검사가 실제로 붉어지는 것을 한 번 보고 나서 채택한다(§1.1).

### D2 — 비발화 방향이 5개 러너 계열 중 1개에만 못 박혀 있고, 전 기준을 만족하는 뮤턴트가 쓰인다 · severity: critical · class: **blocking**

위치: `acceptance.md:61-72` (AC-SEC-003) · `spec.md:123` (§3.5)

§3.5와 AC-SEC-003은 "양방향을 모두 관측"한다고 선언한다. 발화 방향(AC-SEC-001·002·004)은 다섯 러너 계열 전부를 덮는다. **비발화 방향은 go 의 정밀 마커 경로 하나만 덮는다** — AC-SEC-003의 진짜-pass 표본이 `ok  \tgithub.com/modu-ai/moai-adk/internal/hook\t0.590s` 뿐이기 때문이다.

`verification-completeness.md` §2 가 요구하는 뮤턴트를, SPEC의 자기 신고를 받아들이지 않고 직접 써 봤다:

> **뮤턴트**: 0-실행 토큰 6종 거부권을 올바르게 넣고, "안전을 위해" 카운트 휴리스틱 (b)의 `" passed"` 절(`evidence_writer.go:231-233`)을 함께 삭제한다.

- AC-SEC-001 — go 0-매치 표본은 토큰 보유 → 거부권 발화 → `isPass=false` **만족**
- AC-SEC-002 — 거부권이 exit-code 앞이므로 **만족**
- AC-SEC-003 — go 진짜 pass 는 토큰 없음 → 정밀 마커(`:217`)로 살아남음 → `isPass=true` **만족**
- AC-SEC-004 — 다섯 표본 전부 `isPass=false` **만족**
- AC-SEC-005·006·007 — 무관 **만족**

그런데 **pytest 의 진짜 pass `3 passed`** 는 정밀 마커가 없고 0-실행 토큰도 없다. `" passed"` 절이 사라지면 `deriveFromOutputText` 는 신호 없음을 반환하고 `isPass=false` 가 된다 — **REQ-SEC-005("살아 있는 pass 신호를 좁히지 않는다") 위반이며, 채택된 7개 기준 중 어느 것도 이를 잡지 못한다.** (현재 동작은 소스로 확인: `226-233` 행, `" failed"` 없음 → `" passed"` 있음 → `return true, false, true`.)

§2는 이 경우를 명시적으로 판정한다 — *"그런 뮤턴트가 쓰이면 그 기준은 채택하기에 너무 얕다."*

요구되는 수정: AC-SEC-003에 **정밀 마커가 없는 진짜 pass 표본**을 최소 하나 추가한다 — pytest `3 passed`(그리고 가능하면 jest `5 passed`) — 그것이 `isPass=true` 로 남는 것을 단언한다. 그러면 위 뮤턴트가 AC-SEC-003 에서 붉어진다.

### D3 — AC-SEC-007의 판단 탈출구가 뮤턴트를 허용하며, 오늘 그럴 이유도 없다 · severity: major · class: **blocking**

위치: `acceptance.md:131`

조건 (2)는 두 파일의 해당 절이 *"의도된 중립화 차이 외에는 동일"* 하고 *"차이가 있으면 그 사유를 진행 기록에 적는다"* 를 요구한다. 사유를 적으면 어떤 차이든 통과한다 — `moai update` 가 로컬 판을 템플릿판으로 되돌리는 경로(`CLAUDE.local.md` §2.3)를 막겠다는 이 기준의 목적을 정확히 무력화하는 뮤턴트다.

게다가 그 탈출구가 필요 없다. 실측:

```
$ diff .claude/rules/moai/development/verification-completeness.md \
       internal/template/templates/.claude/rules/moai/development/verification-completeness.md
diff-exit=0        ← 오늘 두 파일은 바이트 동일
```

승격될 조항은 SPEC-ID·카드 id·날짜를 담지 않는 일반 산문이므로 중립화 차이가 생길 이유가 없다. 기준은 `diff` rc 0 을 그대로 단언할 수 있다.

요구되는 수정: 조건 (2)를 `diff <로컬> <미러>` rc 0 으로 바꾼다. 중립화 차이가 정말 필요해지면 그때 기준을 고치되, 탈출구를 미리 열어 두지 않는다.

### D4 — AC-SEC-006이 세는 corpus 와 실제로 돌리는 corpus 가 묶여 있지 않다 · severity: major · class: **blocking**

위치: `acceptance.md:108-121`

기준은 "signature 집합의 모든 러너 계열이 corpus 에 최소 1개 표본을 가진다"만 단언한다. 표본이 **분류기에 실제로 먹여져 `isPass=false` 를 내야 한다**는 요구가 없다.

뮤턴트(직접 작성): corpus 를 `map[string]string{sig: ""}` 로 signature 목록에서 파생시킨다. 망라 테스트는 정의상 항상 초록이고, 아무것도 단언하지 않는다. AC-SEC-004가 다섯 표본을 돌리지만 그 표본 집합이 AC-SEC-006이 세는 집합과 같은 객체라는 요구가 어디에도 없다.

요구되는 수정: AC-SEC-006의 Then 에 "corpus 의 각 표본은 AC-SEC-004의 분류기 호출에 실제로 사용되며 `isPass=false` 를 낸다"를 더한다 — 두 기준이 하나의 corpus 를 공유하게 못 박는다.

### D5 — M0이 게이트로 선언돼 있으나 어떤 수락 기준도 그것을 강제하지 않는다 · severity: major · class: **blocking**

위치: `plan.md:33-40` · `acceptance.md:144-148` (DoD)

`plan.md` M0은 "이 계획의 나머지가 여기에 걸려 있다"고 적고, stdout 이 실려 오지 않으면 M1에 들어가지 않고 blocker 보고라고 명시한다. 옳은 판단이다. 그러나:

- 수락 기준 7건 중 M0을 덮는 것이 **없다**.
- DoD 5줄 중 M0 관측이나 `.moai/reports/t341/live-payload.json` 의 존재를 요구하는 줄이 **없다**.
- AC-SEC-001~004는 `classifyTestCommand` 를 직접 호출하고 AC-SEC-005는 구성된 입력으로 핸들러를 부른다 — **전부 합성 fixture 다.** M0을 통째로 건너뛰어도 7건 전부 초록이 된다.

즉 M0은 산문으로 선언된 게이트이고, 완료 판정에는 걸리지 않는다. `spec.md` §5가 정직하게 "가설"이라 적은 그 전제가, 기계적으로는 여전히 **조용한 전제**다.

요구되는 수정: DoD에 한 줄을 더한다 — "`.moai/reports/t341/live-payload.json` 이 존재하고, 그 판독이 progress.md §E.2 에 적혀 있으며, (a)(b)(c) 세 질문에 답한다." 또는 AC-SEC-000 을 신설해 M0 관측을 기준으로 승격한다.

### D6 — 인용 좌표 2건이 어긋난다 · severity: minor · class: optional

- `spec.md:50` — `:223` 을 `return true, false, true` 로 인용하나 실제 `223` 은 `if hasPrecisePass {` 이고 반환은 `224` 다.
- `spec.md:56` — `buildBashRecord` 를 `:296-330` 으로 인용하나 함수는 `309` 에서 시작하고 `IsTestPass` 대입은 `328` 이다.

읽는 사람을 잘못된 줄로 보내지만 주장 자체는 참이다. `a6bbbf82b` 기준으로 좌표만 고치면 된다.

### D7 — 요구 2건이 구현 표면을 이름으로 담는다 · severity: minor · class: optional

`REQ-SEC-004`(`spec.md:159`)는 `HookOutput.SystemMessage` 를, `REQ-SEC-007`(`:165`)은 `make build` 를 요구 문면에 담는다. 요구는 WHAT/WHY 이고 채널 선택은 §3.3의, 빌드 절차는 plan M4의 몫이다. 다만 §3.3이 그 선택을 근거와 함께 이미 논증했으므로 실질 해악은 작다.

### D8 — AC-SEC-003에 RED-now 칸이 없다 · severity: minor · class: optional

`verification-completeness.md` §2는 [HARD] 로 두 칸 쌍을 요구하고 *"green path 만 있는 기준은 시작 관측이 없는 약속"* 이라 못 박는다. AC-SEC-003은 스스로 "오늘 초록이다"라고 적고 §2를 인용하며 짝으로만 채택된다고 선언한다.

**이 부채는 정직하게 신고돼 있고, 대안(비발화 방향 기준을 아예 두지 않는 것)이 더 나쁘다.** 규칙 문면에 회귀-고정자 예외가 없다는 사실만 기록해 둔다 — 규칙을 고칠지 기준을 고칠지는 이 카드 밖의 결정이다. D2의 수정을 하면 이 기준의 값어치는 실질적으로 올라간다.

---

## 5. 범위 규율 (강조 5) — 통과

`plan.md` 전체를 명시된 비목표 셋에 대고 훑었다.

- **grep 0-히트 통과 조건 기준**: 계획에 재유입 없음. 단 **DoD 자신이 그 형태를 쓴다**(D1) — 밀수는 아니고 자기적용 실패다.
- **`t.Skip`**: `plan.md` 어디에도 없음.
- **`sg test`**: 없음. `testCommandSignatures` 에 추가하지 않는다는 §3.6의 선언도 계획에서 지켜진다.
- **규칙층 작업의 크기**: M4는 "문단 한 개 → 조항 한 개, 그 파일의 다른 절을 다시 쓰지 않는다"(`plan.md:68`)로 묶여 있고 `F` 절이 재유입을 안티패턴으로 못 박는다. 재작성 아님 ✓
- **셸 훅 쌍 규율**: `plan.md:75` 가 `.sh`/`.sh.tmpl` 선회 자체를 blocker 사유로 둔다 — `CLAUDE.local.md` §2.3의 드리프트 사고를 정확히 겨눈 제약이다 ✓

---

## 6. 판정 근거와 다음 단계

**PASS-WITH-DEBT** 로 읽는 이유:

- 필수 통과 7건 전부 PASS — 방화벽 미발화.
- 총점 0.81 ≥ Tier M 임계 0.80.
- 핵심 통찰(§2.2 순서 제약)이 **실측으로 참**이며, 이것이 `discovery.md` 위에서 이 판이 더 읽어낸 대목이다. RED-now 칸 4개를 재실행해 전부 오늘도 붉음을 확인했다 — 공허한 기준이 없다.
- 그러나 blocking 5건(D1~D5)이 전부 **기준 문면의 한두 줄 수정**으로 닫히고, 셋(D1·D2·D4)은 `verification-completeness.md` 의 [HARD] 조항 위반이다. 계측기를 짓는 SPEC에서 자기적용 실패를 부채로 넘기면 안 된다.

**요구:**

1. D1~D5를 닫은 뒤 **iter-2 재감사**(Tier M 상한 2회 중 남은 1회). 재감사는 열거된 결함 델타로 한정한다 — 전면 재감사가 아니다.
2. 그때까지 이 판정을 **skip-eligible 로 취급하지 않는다**(`spec-workflow.md` § Plan Audit Gate 세 조건 중 2번을 형식적으로는 만족하나, blocking 부채가 열려 있다).
3. Implementation Kickoff Approval 은 이 판정과 무관하게 여전히 필수이며, 감사 PASS 가 그 게이트를 열지 않는다.

**감사 중 트리 변경 없음.** D1 검증을 위한 임시 편집은 같은 턴에서 되돌렸고 `git status --short` 는 미추적 2건(`.moai/reports/t341/`, `.moai/specs/SPEC-SELECTOR-CENSUS-001/`)만 남는다 — 감사 시작 시점과 동일하다. SPEC 산출물은 한 바이트도 수정하지 않았다.

**미검증으로 남긴 것 (Gaps):**

- jest / vitest / cargo 의 실제 0-실행 출력 문자열은 **실행해 보지 않았다**. AC-SEC-004의 RED 사유는 "`" passed"` 를 담으면 pass 로 읽힌다"는 **소스 판독**으로만 확인했고, 그 러너들이 정말 `0 passed` 를 내는지는 확인하지 않았다. `spec.md` §5가 이를 미검증 전제로 이미 신고했으며, M1이 확인할 몫이다.
- 살아 있는 PostToolUse payload 를 잡아 읽지 않았다 — D5가 지적하는 그 전제는 나도 관측하지 못했다.
- 같은 "missing test → red" 전제를 담은 다른 착지 SPEC 의 수는 세지 않았다.

**잔여 위험:**

- D2의 수정이 AC-SEC-003을 넓히므로, M1이 거부권을 넣을 때 pytest/jest 진짜 pass 경로를 함께 봐야 한다. 넓힌 기준이 M1 설계를 바꿀 수 있다.
- D1의 수정이 기준 트리 SHA 를 새로 못 박는다. 이 워크트리가 `develop` 을 흡수하면 `a6bbbf82b` 가 아니라 **그때의 병합 기준점**으로 다시 재야 한다(§4 evidence pinning).
