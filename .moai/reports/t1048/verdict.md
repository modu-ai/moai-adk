# SPEC-APPJS-IIFE-GUARD-001 — sync-phase 독립 감사 판정서 (card t1048)

**판정: PASS-WITH-DEBT**

| 항목 | 값 |
|---|---|
| 감사자 | sync-auditor (독립 실행, fresh judgment) |
| 평가 프로필 | `.moai/config/evaluator-profiles/default.md` (spec.md 에 `evaluator_profile` 없음 → harness.yaml `default_profile: "default"`) |
| 측정 트리 | `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1048` |
| 측정 브랜치 / HEAD | `WT-iife-guard` / `b0a738434` |
| 측정 일자 | 2026-09-21 |
| 착지 파일 | 5 (SPEC 산출물 4 + `internal/web/appjs_iife_scope_test.go` 1) |

---

## §1 Baseline-attribution (이 실행에서 잰 좌표)

```
$ git rev-parse --show-toplevel
/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1048
$ git branch --show-current
WT-iife-guard
$ git rev-parse --short HEAD
b0a738434
```

아래 모든 Evidence 는 **이 트리, 이 HEAD 에서 이 실행 중에** 측정했다. run-phase
(`progress.md` §E.2, HEAD `a86d8e12c`) 의 수치를 근거로 옮겨 쓴 곳은 §6 에서 **어느
것이 인계값인지 명시**한다.

---

## §2 차원 점수

| 차원 | 점수 | 판정 | Evidence (요지 — 축자 출력은 §3) |
|---|---|---|---|
| Functionality (40%) | 95/100 | **PASS** | AC-AIG-001~007 전부 독립 재측정 통과. 정방향 `checked=4, violations=0, total=20`, 역방향 `violations=1 line 660: stampRefreshed`, 대수 동일성 `4→4` |
| Security (25%) | 94/100 | **PASS** | 프로덕션 코드 0줄. 테스트 전용, 디스크 쓰기 0, 네트워크 0. 정규식은 Go RE2(선형, ReDoS 불성립) + 식별자 보간에 `regexp.QuoteMeta`. Critical/High 0건 |
| Craft (20%) | 70/100 | **FAIL (임계 미달)** | `gofmt -l internal` 0행 · `go vet` EXIT=0 · `golangci-lint run ./internal/web/...` **0 issues** · 그러나 패키지 커버리지 **67.8% < 프로필 하드 임계 85%** |
| Consistency (15%) | 92/100 | **PASS** | 패키지 관용구 준수(`t.Parallel`, `readEmbeddedAsset`, 명명). 한국어 주석은 이 패키지의 **기존 지배 관행**(§5 F2) |

**가중 조화평균 (프로필 가중 40/25/20/15): 88.0**
**단순 조화평균 (4차원 동등): 86.4**

```
가중:  1 / (0.40/95 + 0.25/94 + 0.20/70 + 0.15/92) = 1 / 0.0113576 = 88.05
단순:  4 / (1/95 + 1/94 + 1/70 + 1/92)             = 4 / 0.0463199 = 86.36
```

### 필수통과 방화벽 (must-pass firewall)

프로필의 must-pass 는 **Functionality + Security** 둘뿐이고 **둘 다 PASS** 다. Craft 의
임계 미달은 전체 FAIL 을 강제하지 않는다(프로필 § Hard Thresholds 의 전체-FAIL 강제는
**Security FAIL** 한 줄뿐이다). 따라서 **PASS-WITH-DEBT** 이며, 그 debt 의 정체는 §4 에
적는다.

---

## §3 AC 별 검증 (Claim / Evidence / Gaps)

### AC-AIG-001 — 정방향: 통과하되 공허하지 않게

**Claim** 위반 0 **그리고** 검사 대수 ≠ 0 이 **함께** 성립한다.

**Evidence**
```
$ go test ./internal/web/ -run 'IIFE|Scope' -count=1 -v
    appjs_iife_scope_test.go:219: forward: bare-identifier handlers checked = 4, violations = 0, total addEventListener calls = 20
--- PASS: TestAppJsBareHandlersDeclaredInSameIIFE (0.00s)
--- PASS: TestAppJsBareHandlerScopeRuleRejectsCrossIIFERegistration (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/web	0.514s
=== EXIT 0 ===
```

**독립 양성 대조 — 비-공허 단언이 실제로 무는가.** 스크래치패드 사본에서 스팬 스캐너를
눈멀게 해(`"(function"` → `"(functionZZZ"`) 표면을 0 으로 떨어뜨렸다:
```
--- FAIL: TestAppJsBareHandlersDeclaredInSameIIFE (0.00s)
    rule_test.go:211: rule matched ZERO bare-identifier handlers — a rule that inspects nothing passes the violations==0 assertion vacuously; this is the failure this guard exists to prevent
--- FAIL: TestAppJsBareHandlerScopeRuleRejectsCrossIIFERegistration (0.00s)
    rule_test.go:241: mutation synthesis precondition failed: need at least 2 top-level IIFEs, found 0
```
단언이 **문다**. 공허 통과는 재현되지 않는다.

**추가 확인 — 통과가 자명하지 않은가.** 규칙이 실제로 두 IIFE 를 가르고 있는지 확인했다:
```
$ grep -n -E '^\(function|^\}\)\(\);' internal/web/assets/app.js
39:(function () {
660:})();
673:(function () {
796:})();
$ grep -n -E '(function|var|let|const)[ \t]+(syncSegmentsVisibility|initConsole|stampRefreshed)([^A-Za-z0-9_$]|$)' internal/web/assets/app.js
45:  function syncSegmentsVisibility() {
526:  function initConsole() {
711:  function stampRefreshed() {
```
스팬 `#0=[39,660]`, `#1=[673,796]`. 등록 4건은 530·548·551(→ #0) 과 727(→ #1) 로
**두 스팬에 걸쳐 분포**하고 선언도 45·526(#0) / 711(#1) 로 분포한다. 즉 "모두 한 스팬
안이라 자동으로 통과" 하는 자명한 초록이 **아니다**.

**판정: PASS**

---

### AC-AIG-002 — 역방향: 그 결함을 거부하고 지목한다

**Claim** 돌연변이에 위반 1건, 식별자와 줄 번호를 모두 담고, 검사 대수는 정방향과 동일.

**Evidence**
```
    appjs_iife_scope_test.go:248: mutant: checked = 4 (forward 4), violations = 1: [line 660: stampRefreshed — declared in IIFE#[1], not IIFE#0]
--- PASS: TestAppJsBareHandlerScopeRuleRejectsCrossIIFERegistration (0.00s)
```
위반 1건 · 식별자 `stampRefreshed` · 줄 번호 660 · 대수 `4 → 4` 유지. 세 단언 모두 성립.

**전제 단언이 실제로 무는가 (AC 가 요구한 `t.Fatal`).** 위 스팬-눈멂 돌연변이에서
`mutation synthesis precondition failed: need at least 2 top-level IIFEs, found 0` 가
발화했다 — 합성 실패가 "위반 0" 으로 조용히 뒤집히지 않는다.

**판정: PASS**

---

### AC-AIG-003 — 기존 계약 회귀 없음

**Claim** 패키지 전체 통과 + 두 명시 테스트가 **수정되지 않은 채로** 통과.

**Evidence**
```
$ go test ./internal/web/ -count=1 -timeout 30m
ok  	github.com/modu-ai/moai-adk/internal/web	25.274s
REAL EXIT=0
$ grep -c -- '--- FAIL' <로그>
0
$ go test ./internal/web/ -run 'TestAppJsInitConsoleBoundToBothEvents|TestAppJsHxBoostPreserved' -count=1 -v
--- PASS: TestAppJsInitConsoleBoundToBothEvents (0.00s)
--- PASS: TestAppJsHxBoostPreserved (0.01s)
ok  	github.com/modu-ai/moai-adk/internal/web	0.670s
$ git log develop..HEAD --name-only --format="" -- internal/web/appjs_reinit_test.go | wc -l
       0
```

판정식은 acceptance.md 의 지시대로 **요약 행 `ok` + 종료 코드**로 갈랐다(`--- FAIL`
세기는 순수 타임아웃에서 0 을 내므로 단독 판별식이 아니다). `REAL EXIT=0` 은 파이프
없이 잡은 값이다.

**판정: PASS**

---

### AC-AIG-004 — 함수 표현식 핸들러는 대상 밖

**Claim** 검사 대수(4)가 전체 `addEventListener`(20)보다 **엄격히 작다**.

**Evidence** — AC-AIG-001 로그의 `checked = 4 … total addEventListener calls = 20`.
테스트가 `checked >= total` 을 오류로 단언한다(`appjs_iife_scope_test.go:222`).

```
$ grep -n -oE '\.addEventListener\(\s*"[^"]*"\s*,\s*[A-Za-z_$][A-Za-z0-9_$]*\s*\)' internal/web/assets/app.js
530:.addEventListener("change", syncSegmentsVisibility)
548:.addEventListener("DOMContentLoaded", initConsole)
551:.addEventListener("htmx:afterSettle", initConsole)
727:.addEventListener("htmx:afterSettle", stampRefreshed)
```
정확히 4건 — 나머지 16건은 표현식 핸들러다.

**판정: PASS**

---

### AC-AIG-005 — 좌표 하드코딩 없음

**Claim** IIFE 스팬(39/660/673/796)과 등록 줄(727)이 판정용 상수로 등장하지 않는다.

**Evidence**
```
$ grep -n -E '\b(39|660|673|727|796|552)\b' internal/web/appjs_iife_scope_test.go
$ echo $?
1
```
적중 **0건** — 주석 안의 설명용 언급조차 없다(AC 는 주석 언급을 허용하지만 실제로는
하나도 없다).

**양성 대조 (빈 결과가 계측기 고장이 아님을 확립)**
```
$ grep -c -E '\b(1|0|2)\b' internal/web/appjs_iife_scope_test.go
41
```
같은 계측기가 같은 파일에서 41행을 낸다. 위 0건은 **부재**이지 미측정이 아니다.

**판정: PASS**

---

### AC-AIG-006 — 범위 침범 없음

**Claim** `internal/web/assets/` · `go.mod` · `go.sum` · `.github/workflows/` 변경 0건.

**Evidence** — 두 끝점 diff 가 아니라 **커밋 열거**로 쟀다(트리 비교는 이 카드의 파일과
develop 의 파일을 가르지 못한다):
```
$ git log develop..HEAD --name-only --format="" | sort -u
.moai/specs/SPEC-APPJS-IIFE-GUARD-001/acceptance.md
.moai/specs/SPEC-APPJS-IIFE-GUARD-001/plan.md
.moai/specs/SPEC-APPJS-IIFE-GUARD-001/progress.md
.moai/specs/SPEC-APPJS-IIFE-GUARD-001/spec.md
internal/web/appjs_iife_scope_test.go

$ git log develop..HEAD --name-only --format="" -- internal/web/assets/ go.mod go.sum .github/workflows/ | sort -u
(출력 없음)
```

**양성 대조 (`-- <경로>` 인자가 실제로 발화하는가)**
```
$ git log develop..HEAD --name-only --format="" -- internal/web/ | sort -u
internal/web/appjs_iife_scope_test.go
```
같은 형태의 명령이 비어 있지 않은 결과를 낸다. 위 빈 출력은 **침범 0**이지 오타가 아니다.

**판정: PASS**

---

### AC-AIG-007 — 한계가 소스에 적혀 있다

**Claim** spec.md §F 의 세 한계 + "완전한 스코프 검사가 아니다" 취지가 테스트 소스
상단 주석에 있다.

**Evidence** — `internal/web/appjs_iife_scope_test.go:16-38` 의 헤더 주석이
`── 이 가드는 완전한 스코프 검사가 아니다 (spec.md §F 한계 선언) ──` 표제 아래 4개
항목을 담는다. spec.md §F 의 세 항목과 1:1 로 대응한다:

| spec.md §F | 테스트 주석 |
|---|---|
| 1. 다른 형태의 최상위 문장 | 항목 1 (동일 문언) |
| 2. 중첩 함수 안의 선언 (거짓 음성) | 항목 2 (동일 문언) |
| 3. 함수 표현식 핸들러 안의 교차 참조 | 항목 3 (동일 문언) |
| — | 항목 4 (런타임 발화 축 → card t1060, `:34`) |

마지막 문단도 있다: "이 테스트의 초록을 '교차 경계 참조가 없다' 로 읽으면 과대
해석이다."

**Gap** — 주석 헤더에 **없는** 네 번째 거짓 음성 방향을 §5 F1 에서 보고한다(인라인
주석 `:119`, `:145` 에는 있으나 AC-AIG-007 이 지배하는 헤더 한계 목록에는 없다).

**판정: PASS** (Gap 은 비차단)

---

## §4 리드가 지목한 세 주장 — 내 판독

### (a) 변이 검증의 비대칭 — **확인함**

run-phase 의 M-b 주장(`found := true` 주입 시 **정방향은 PASS**, 역방향만 FAIL)을
스크래치패드 독립 재현으로 직접 쟀다. 트리는 한 바이트도 건드리지 않았다
(`internal/web/appjs_iife_scope_test.go` + `internal/web/assets/app.js` 를
`$SCRATCH/mut/` 로 복사, `readEmbeddedAsset` → 로컬 파일 읽기로 치환, 무변이 baseline
`ok mut 0.271s` 로 재현 확인 후 변이 주입).

```
=== MUTANT: found := true (rule always passes) ===
=== NAME  TestAppJsBareHandlersDeclaredInSameIIFE
    rule_test.go:220: forward: bare-identifier handlers checked = 4, violations = 0, total addEventListener calls = 20
--- PASS: TestAppJsBareHandlersDeclaredInSameIIFE (0.00s)
=== NAME  TestAppJsBareHandlerScopeRuleRejectsCrossIIFERegistration
    rule_test.go:246: mutant should produce exactly 1 violation, got 0: []
--- FAIL: TestAppJsBareHandlerScopeRuleRejectsCrossIIFERegistration (0.00s)
FAIL	mut	0.283s
```

**정방향 PASS, 역방향만 FAIL.** 비대칭이 관측됐다.

**왜 이것이 양방향 필요성의 근거인가 — 단언을 읽은 결과.** 정방향 테스트의 판정
단언은 `for _, v := range violations { t.Errorf(...) }` 이고, 현재 트리의 위반은 **원래
0** 이다. "항상 통과하는 규칙" 도 위반 0 을 낸다. 즉 **정방향에는 이 변이가 잡힐
표면이 구조적으로 없다** — 잡을 위반이 존재하지 않으니 "규칙이 옳아서 0" 과
"규칙이 눈멀어서 0" 이 정방향에서는 **구별 불가능**하다. 역방향은 위반이 **있어야
하는** 입력을 합성해 만들므로, 이 계열의 유일한 검출 경로다.

### (b) 축소 퇴행(4→3)은 닫히지 않았다 — **정정이 옳다, 실측으로 확인**

먼저 단언을 읽었다. `TestAppJsBareHandlerScopeRuleRejectsCrossIIFERegistration` 은
```go
_, forwardChecked := scanBareHandlerScope(src)      // 같은 실행의 원본
...
violations, mutantChecked := scanBareHandlerScope(mutant)
...
if mutantChecked != forwardChecked { t.Errorf(...) }
```
두 값을 **같은 실행 안에서** 만든다. 표면이 커밋 사이에 줄면 두 값이 **함께** 줄어
동일성은 그대로 성립한다.

**실측으로 확인.** 스크래치패드 app.js 사본에서 530행의 bare 핸들러 하나를 표현식
핸들러로 바꿔 표면을 4→3 으로 축소시켰다:
```
mutated: preset.addEventListener("change", function(){syncSegmentsVisibility();});
=== SHRINK SIM: surface 4 -> 3, does anything fail? ===
    rule_test.go:220: forward: bare-identifier handlers checked = 3, violations = 0, total addEventListener calls = 20
--- PASS: TestAppJsBareHandlersDeclaredInSameIIFE (0.00s)
    rule_test.go:249: mutant: checked = 3 (forward 3), violations = 1: [...]
--- PASS: TestAppJsBareHandlerScopeRuleRejectsCrossIIFERegistration (0.00s)
ok  	mut	0.250s
```
**둘 다 초록.** 대수 동일성은 `3 == 3` 으로 통과하고, `checked < total` 도 `3 < 20` 으로
통과한다. 축소 퇴행을 잡는 장치는 이 트리에 **없다**.

따라서 최초 문장("대수 동일성이 부분적으로 흡수한다")은 **거짓**이었고, 현재 문언
("감수한 것이지 흡수된 것이 아니며, 이 트리에서 그 구멍을 닫는 장치는 없다" —
`appjs_iife_scope_test.go:201,208` · `progress.md:251`)이 **옳다**.

### (c) `$`-포함 식별자 — run-phase 의 "위험을 피한 것이지 잡은 것이 아니다" 는 **옳다**

두 방향으로 쟀다.

**먼저, 현재 트리에서 판정이 갈리는가 — 갈리지 않는다.** 명시 문자류를 Go 의
`\w` / `\W` 로 치환한 사본을 돌렸다:
```
=== MUTANT 3 (corrected): Go \w / \W substituted for explicit classes ===
    rule_test.go:220: forward: bare-identifier handlers checked = 4, violations = 0, total addEventListener calls = 20
--- PASS
    rule_test.go:249: mutant: checked = 4 (forward 4), violations = 1: [line 660: stampRefreshed — declared in IIFE#[1], not IIFE#0]
--- PASS
ok  	mut	0.298s
```
**판정 바이트 동일.** 현재 app.js 의 `$` 4건은 전부 CSS 속성 선택자·정규식 문자열
안이고(`select[name$=".effort"]` 등, `internal/web/assets/app.js:421,452,462,470`)
식별자에 `$` 를 쓰는 곳은 0 이다. 따라서 이 설계 선택은 이 트리에서 **미측정 축**이지
안전성 확인이 아니다 — run-phase 의 서술이 정확하다.

**다음으로, 설계 선택 자체가 실질적으로 옳은가 — 옳다. 그리고 그것은 내가 쟀다.**
`$`-포함 핸들러를 합성해 두 판을 대조했다:

| 정규식 | 합성 입력 | 검사 대수 |
|---|---|---|
| 착지본 `[A-Za-z_$][A-Za-z0-9_$]*` | `preset.addEventListener("change", sync$Seg);` + `function sync$Seg() {` | **4** (규칙이 본다) |
| Go `\w+` | 동일 입력 | **3** (규칙이 조용히 눈먼다) |

`\w` 판은 표면이 4→3 으로 줄었는데도 **두 테스트 모두 초록**이다(위 (b) 의 축소 구멍
그대로). 즉 `$` 처리는 실질적으로 옳지만 **그 올바름을 지키는 가드가 트리에 없다** —
장래에 누군가 `\w` 로 "단순화" 하면 조용히 통과한다. §5 F4 로 보고한다.

---

## §5 Findings (구조화 결함 목록)

프로필 § Finding-Stage Reporting 에 따라 확신도·심각도를 붙여 **거르지 않고** 전부
적는다. 차단 판정은 §2 의 must-pass 방화벽이 한다.

- **F1** [Low] [optional] `internal/web/appjs_iife_scope_test.go:145` — **어느 IIFE 에도
  속하지 않는 최상위 등록은 조용히 건너뛴다.** 브라우저에서 그런 등록이 IIFE 안 식별자를
  참조하면 실제로 ReferenceError 이므로, 이것은 진짜 거짓 음성 방향이다. 인라인 주석
  (`:119`, `:145`)에는 "판정 대상 밖" 으로 적혀 있으나 **AC-AIG-007 이 지배하는 헤더
  한계 목록(`:16-38`)에는 없다.**
  확신도: 높음 — 실측했다. 합성 입력에 파일 최상위 등록 1건을 덧붙이자
  `total addEventListener calls` 는 20→21 로 올랐는데 `checked` 는 **4 에 그대로 머물렀다**.
  Required fix (선택): 헤더 한계 목록에 5번 항목으로 한 줄 추가.

- **F2** [Info] [optional] 파일 주석이 한국어다(`91/103` 행). 프로젝트 설정은
  `code_comments: en` (`.moai/config/sections/language.yaml`).
  확신도: 높음, 그러나 **이 카드가 만든 것이 아니다** — 이 패키지의 지배 관행이다
  (`viewmodel_ops.go` 76/108, `agent_settings_test.go` 54/66, `schemaform.go` 72/163).
  혼자 영어로 쓰는 쪽이 오히려 이질적이므로 Consistency 감점을 **하지 않았다**.
  Required fix: 없음(패키지 차원의 별도 판단 사항).

- **F3** [Medium] [optional — 선언된 debt] 검사 대수 축소 퇴행(4→3)을 닫는 장치가 없다.
  §4(b) 에서 실측 확인. **결함이 아니라 명시적으로 선언·정정된 감수**이므로 비차단으로
  둔다. 다음 독자가 같은 추론을 다시 하지 않도록 여기 기록한다.
  Required fix (선택): 커밋 간 비교가 필요하므로 단위 테스트로는 닫히지 않는다 — 골든
  카운트 파일이나 CI 축 장치가 필요하고, 그것은 이 카드의 Out of Scope 다.

- **F4** [Low] [optional] `$`-포함 식별자 처리의 올바름을 **지키는 테스트가 없다.**
  §4(c) 에서 `\w` 로 치환하면 `$`-포함 핸들러에서 표면이 조용히 줄고 두 테스트가 모두
  초록으로 남는 것을 실측했다. `bareHandlerPattern` (`:74-75`) 과 `declPattern`
  (`:85-88`) 의 주석은 이유를 적고 있으나 주석은 회귀를 막지 못한다.
  확신도: 높음 — 실측했다.
  Required fix (선택): `scanBareHandlerScope` 에 `$`-포함 식별자를 담은 **합성 소스**를
  먹이는 순수 단위 테스트 1개. app.js 를 건드리지 않고 규칙만 겨누므로 값싸다.

- **F5** [Info] [optional] `.moai/reports/t1048/verdict.md` 는 gitignore 대상이다
  (`git check-ignore -v` → `.gitignore:227:.moai/reports/*`). 이 판정서는 워크트리
  유일본이며 병합과 함께 이동하지 않는다. 워크트리 폐기 전 반출 여부는 리드 판단.

**차단(blocking) 결함: 0건.**

---

## §6 거짓 주장 정정의 경위

run-phase 의 manager-develop 은 자기 보고서 안에서 "검사 대수 축소 퇴행을 AC-AIG-002 의
대수 동일성 단언이 부분적으로 흡수한다" 는 문장을 **스스로 반박했다.** 그러나 반박은
보고서에만 남았고, **같은 거짓 문장은 두 산출물에 그대로 서 있었다**:

1. 테스트 소스 주석 (`internal/web/appjs_iife_scope_test.go`, 정정 전 ~195-200행)
2. `progress.md:242`

레인이 이것을 잡아내 **둘 다** 정정했고, 그 정정이 커밋 `b0a738434`
(`docs(SPEC-APPJS-IIFE-GUARD-001): correct a false claim about what the guard protects (t1048)`)
로 착지했다. diff 실측:

```
$ git show b0a738434 --stat
 .moai/specs/SPEC-APPJS-IIFE-GUARD-001/progress.md | 18 ++++++++++++++----
 internal/web/appjs_iife_scope_test.go             | 14 +++++++++++---

-	// 검사 대수가 조용히 줄어드는 퇴행 — 은 아래 역방향 테스트의 **대수 동일성**
-	// 단언이 부분적으로 흡수한다(돌연변이와 정방향의 대수가 달라지면 거기서 걸린다).
+	// 그래서 놓치는 것 — 검사 대수가 커밋 사이에 조용히 줄어드는 퇴행 — 은
+	// **감수한 것이지 흡수된 것이 아니다.** …
+	// 대신 아래 `checked < total` 단언이 다른 축에서 일부를 되찾는다 — 그것도
+	// 축소 구멍을 닫지는 않는다. 이 트리에서 그 구멍을 닫는 장치는 없다.
```

정정 후 문언이 옳다는 것은 §4(b) 에서 **실측**으로 확인했다.

### 왜 이것이 중요한가

이 카드의 [HARD] 는 "**가드가 자기 유효 범위를 스스로 선언한다**" 이다. 그 선언이
거짓 전제 위에 서면, 독자는 **존재하지 않는 보호를 있다고 믿는다.** 가드가 아무 보호도
선언하지 않았다면 독자는 스스로 확인했을 것이다 — 틀린 선언은 무선언보다 나쁘다.
이 경우 구체적으로는 "대수가 줄어도 어딘가에서 걸린다" 는 믿음이었고, §4(b) 의 실측은
그 상황에서 **두 테스트가 모두 초록**임을 보인다.

### 이 배치에서 새로 관측된 실패 형태 — **"보고서는 정정, 산출물은 그대로"**

기존에 기록된 계열들과 다른 축이다:

- **전파 오류**(기억으로 쓴 요약이 본문과 갈림)가 아니다 — 여기서는 **같은 주체가 같은
  세션 안에서** 거짓을 **인지하고 반박까지 했다.**
- **미정정**이 아니다 — 정정은 실제로 일어났다. 다만 **정정이 착지한 표면이 하나뿐**
  이었다.

위험한 이유는 **자기 은폐성**이다. 보고서에 반박이 적혀 있으므로 그 보고서만 읽는
검토자에게는 "이미 처리됨" 으로 보이고, 산출물만 읽는 다음 독자에게는 거짓이 **정정된
적 없는 것과 똑같이** 보인다. 두 독자 어느 쪽도 어긋남을 볼 수 없다. 어긋남을 보려면
**보고서와 산출물을 나란히** 열어야 하는데, 그것은 정정이 끝났다고 믿는 순간 아무도
하지 않는 일이다.

일반 규율: **거짓 주장을 정정할 때는 그 주장이 실린 표면을 전부 세고, 센 개수를
보고한다.** "보고서에서 반박했다" 는 정정의 완료가 아니라 착수다.

---

## §7 변이 검증의 비대칭 (양방향이 필요한 이유)

§4(a) 의 실측이 근거다. 요지를 다시 적는다:

| 변이 | 정방향 테스트 | 역방향 테스트 |
|---|---|---|
| M-b: `found := false` → `found := true` (규칙이 **항상 통과**) | **PASS** | **FAIL** (`got 0` violations) |
| 스팬 스캐너 눈멂 (표면 → 0) | FAIL (`checked == 0`) | FAIL (전제 단언) |

**정방향 단독으로는 M-b 가 구조적으로 잡히지 않는다.** 현재 app.js 의 위반이 원래 0
이므로, 정방향이 관측하는 것은 "위반 0" 한 값뿐이고 그 값은 **올바른 규칙과 항상
통과하는 규칙 양쪽에서 동일하게 나온다.** 잡을 위반이 없는 입력에서는 "잡는 능력" 을
측정할 수 없다 — 이것이 정방향의 눈먼 지점이고, 설정 실수가 아니라 **구성상의 한계**다.

역방향 테스트는 위반이 **있어야 하는** 입력을 메모리 안에서 합성해 그 눈먼 지점을 정확히
겨눈다. 그래서 이 가드에서 역방향은 "있으면 좋은 추가 테스트" 가 아니라 **이 결함 계열의
유일한 검출 경로**다.

두 번째 변이 방향(표면 눈멂)이 별도로 필요한 이유도 같은 구조다. 그쪽은 "규칙이 거부를
검출하는가" 가 아니라 "**규칙이 애초에 무엇인가를 보고 있는가**" 를 묻는다. 한 방향만
돌리면 위험한 변이와 무해한 no-op 이 구별되지 않는다.

---

## §8 RED 정직성

run-phase 의 기록(`progress.md` §E.2):

> Go 는 테스트와 피검 함수가 같은 패키지에 있어야 하므로 이 파일은 한 번에 작성됐고,
> 위 RED 는 **규칙 본문을 스텁으로 되돌려** 생성했다. 즉 "테스트를 먼저 쓴 뒤
> 구현했다" 가 아니라 "**구현을 제거하면 테스트가 실패함을 보였다**" 가 정확한 서술이다.

**이 서술 방식에 동의한다.** 두 가지 이유다.

**첫째, 정확하다.** 주장한 것과 관측한 것이 일치한다. 저작 순서는 관측되지 않았고 —
git 은 한 커밋 안의 저작 순서를 증언하지 못한다 — 그 서술은 하지 않았다. 대신
실행으로 확립되는 명제만 주장했다.

**둘째, 보존되어야 할 성질이 실제로 보존된다.** RED 가 지키려는 성질은 "테스트가 먼저
쓰였다" 라는 연대기가 아니라 "**단언이 실제로 문다**" 이다. 그 성질은 저작 순서가
아니라 **제거-실패 대응**으로 확립되고, 그것이 정확히 관측된 것이다. 나는 이 대응을
독립적으로 재현했다(§4(a) 두 변이 방향, 둘 다 FAIL 발화).

덧붙여, 이 서술은 **약한 주장이지 강한 주장이 아니다.** "테스트를 먼저 썼다" 가 더
강하게 들리지만 검증 불가능하고, 실제로 지키는 것은 더 적다. 약하지만 검증 가능한
주장으로 바꾼 것은 등급 하락이 아니라 **등급 정확화**다.

---

## §9 Gaps (6건) — run-phase 로부터 인계

`progress.md` §E.2 Gaps 전량을 인계하고, 내가 이 실행에서 **닫은 것과 닫지 못한 것을
구분해** 표시한다. 인계값과 내 측정값을 섞지 않는다.

1. **런타임 동작은 검증되지 않았다.** 브라우저를 띄우지 않았고 리스너가 실제로 붙는지
   재지 않았다. 이 가드는 소스 문자열만 본다.
   → **여전히 열려 있다.** **런타임 발화 축은 card t1060 소관으로 넘긴다** (그 카드는
   큐에 있고 이 카드 착지 후 열린다). 테스트 소스 헤더 주석 4번 항목
   (`appjs_iife_scope_test.go:34`) 에도 같은 문장으로 기재돼 있다.

2. **spec.md §F 의 세 한계에 대한 음성 결과는 측정되지 않았다.** 세 방향으로 돌연변이를
   돌려 "잡지 못함" 을 보이지 않았다 — 선언했을 뿐이다.
   → **여전히 열려 있다.** 나도 이 셋에 대한 음성 돌연변이는 돌리지 않았다. 다만
   §5 F1 에서 **선언 목록에 없던 네 번째 방향**(IIFE 밖 최상위 등록)의 음성 결과는
   실측했다.

3. **`golangci-lint` 를 돌리지 않았다.** `gofmt` / `go vet` 만 관측했다.
   → **내가 닫았다.** `golangci-lint run ./internal/web/...` → `EXIT=0`, 출력
   `0 issues.` (golangci-lint 2.10.1). 이 측정은 **내 것**이지 run-phase 의 것이 아니다.

4. **크로스 플랫폼 빌드(`GOOS=windows`)를 재지 않았다.**
   → **내가 닫았다.** `GOOS=windows GOARCH=amd64 go vet ./internal/web/` → `EXIT=0`,
   출력 없음. 이 측정도 **내 것**이다.

5. **다른 정적 자산(`i18n.js` 등)에 같은 계열이 있는지 재지 않았다.**
   → **여전히 열려 있다.** 이 카드의 Out of Scope 이고, 나도 재지 않았다.

6. **`$`-포함 식별자에 대해 규칙이 옳게 동작하는지 재지 않았다.** `\w` 를 피한 것은
   예방이지 관측이 아니다.
   → **부분적으로 닫았다.** §4(c) 에서 합성 입력으로 **규칙이 `$`-포함 식별자를 옳게
   본다**(검사 대수 4 유지)는 것과 **`\w` 판은 조용히 눈먼다**(4→3, 두 테스트 초록)는
   것을 실측했다. 남는 것은 **그 올바름을 지키는 가드가 트리에 없다**는 점이며,
   §5 F4 로 보고한다. run-phase 의 "예방이지 관측이 아니다" 는 **그 시점 서술로서
   정확했다.**

### 내가 추가로 여는 Gap

7. **커버리지의 "변화 없음" 중 한쪽은 내 측정이 아니다.** 나는 신규 파일이 **있는**
   상태에서 `67.8%` 를 쟀다(`go test -cover ./internal/web/ -count=1 -timeout 30m`).
   **없는** 상태의 값은 run-phase 가 같은 실행에서 양방향으로 쟀다고 보고한 인계값이고
   (`progress.md` §E.2, HEAD `a86d8e12c`), 나는 파일을 옮기지 않았으므로 재측정하지
   않았다. 다만 기계적으로는 이동 가능성이 없다 — Go 커버리지는 `_test.go` 를 계측하지
   않고, 이 파일이 부르는 유일한 프로덕션 경로 `staticFS()` 는 같은 패키지의 기존
   테스트(`restyle_test.go:50`)가 이미 통과한다. **이것은 추론이지 측정이 아니다.**

8. **거부된 명령 없음 — 명시.** 이 감사에서 워크트리 가드가 복합 명령 1건을 거부했다
   (스크래치패드 구성 스크립트). 단순 명령으로 쪼개 **전부 실행했고**, 그 거부로 인해
   포기한 측정은 없다. 대체물로 관측을 제시한 곳은 없다.

---

## §10 Residual-risk

§9 의 Gaps(관측하지 **않은** 것)와 구분해, **관측했음에도 여전히 틀릴 수 있는** 것을
적는다.

- **스팬 스캐너는 줄 접두사 기준이다** (`"(function"` / `"})();"`,
  `appjs_iife_scope_test.go:97-108`). 장래에 컬럼 0 에서 다른 형태로 열거나 닫는 IIFE 가
  생기면 스팬 산출이 어긋난다. 다만 그때는 조용히 틀리지 않고 **비-공허 단언이 즉시
  발화**한다(§3 AC-AIG-001 의 양성 대조가 그 시끄러움을 실증한다).
- **정규식은 한 줄 안에서만 매치한다.** `addEventListener(` 호출이 여러 줄로 쪼개지면
  검사 대수에서 빠지고, §4(b) 의 축소 구멍 때문에 **그 이탈은 조용하다.** 현재 20건은
  전부 한 줄이다(실측).
- **선언 탐색은 문자열·주석을 구분하지 않는다.** 주석 안의 `function stampRefreshed`
  같은 문자열이 선언으로 오인될 수 있다 — 거짓 **음성** 방향(위반을 놓침)이다.
- **`declSpans` 는 파일 전역을 훑는다.** 같은 이름이 두 IIFE 에 모두 선언돼 있으면
  등록이 어느 쪽에 있든 통과한다. 이름 중복이 생기면 규칙의 분해능이 떨어지지만,
  현재 세 이름은 각각 한 곳에만 선언돼 있다(실측: 45 / 526 / 711).
- **커버리지 67.8% 는 패키지 선재 부채다.** 이 카드가 만들지도, 움직일 수도 없다
  (§9-7). Craft 임계 미달의 원인이 이것이며, 이 카드에 귀속시키면 오귀속이다.

---

## §11 최종 판정

**PASS-WITH-DEBT** — 병합 가능. 차단 결함 0건.

- **Functionality (40%) 95 — PASS** (must-pass 충족)
- **Security (25%) 94 — PASS** (must-pass 충족)
- **Craft (20%) 70 — FAIL (임계 미달)**: 커버리지 67.8% < 85%. **선재 패키지 부채이며
  이 카드가 움직일 수 없다**(테스트 전용 변경은 계측 대상 문장을 0개 추가한다).
  도구 축은 전부 깨끗하다(gofmt 0행 · vet EXIT=0 · golangci-lint 0 issues).
- **Consistency (15%) 92 — PASS**

**가중 조화평균 88.0 / 단순 조화평균 86.4.**

must-pass 방화벽(Functionality + Security)이 둘 다 성립하므로 Craft 의 임계 미달은
전체 FAIL 을 강제하지 않는다. `debt` 의 정체는 **커버리지 선재 부채 1건**이며, 이
카드에 귀속되는 부채가 아니다. F1·F3·F4 는 전부 optional 이고 차단하지 않는다.

특기할 점: 이 카드의 실제 산출물은 테스트 1개가 아니라 **"가드가 자기 유효 범위를
정직하게 선언한다"** 이며, 거짓 선언이 두 산출물에 남아 있던 것을 레인이 잡아 정정한
것(§6)이 그 산출물의 핵심 부분이다. 그 정정이 옳다는 것을 §4(b) 에서 **실측으로**
확인했다.

---

*판정자: sync-auditor · 트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1048` ·
branch `WT-iife-guard` · HEAD `b0a738434` · 2026-09-21*
