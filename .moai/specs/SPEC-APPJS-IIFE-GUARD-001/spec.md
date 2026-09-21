---
id: SPEC-APPJS-IIFE-GUARD-001
title: "app.js IIFE 경계 넘는 핸들러 참조 회귀 가드 — 로드 시점 ReferenceError 가 무관한 리스너를 함께 죽이는 계열을 막는다"
version: "0.1.0"
status: completed
created: 2026-09-20
updated: 2026-09-21
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: "internal/web"
lifecycle: spec-anchored
tags: "web, app.js, iife, scope, guard, regression, static-analysis, t1048"
era: V3R6
tier: S
related_specs: [SPEC-WEB-CONSOLE-006, SPEC-DESIGN-MOAIWEBV2-002]
---

# SPEC-APPJS-IIFE-GUARD-001 — app.js IIFE 경계 참조 회귀 가드

## HISTORY

| 날짜 | 버전 | 변경 | 주체 |
|---|---|---|---|
| 2026-09-20 | 0.1.0 | plan-phase 최초 작성 (card t1048) | manager-spec |

---

## §A 배경 — 이미 수리된 결함, 아직 없는 가드

수리는 `18fa0d384` (`fix(web): resolve visual e2e ux regressions`) 로 develop 에 착지했고, 이 SPEC 의 트리(`d8304b49a`)에도 조상으로 들어와 있다. **이 카드는 수리가 아니라 가드다.** `internal/web/assets/app.js` 는 손대지 않는다.

### 결함 계열

`internal/web/assets/app.js` 에는 최상위 IIFE 가 정확히 둘 있다. 수리 이전에는 `stampRefreshed` 의 리스너 등록이 **첫째** IIFE 안에 있었고, 그 식별자의 선언은 **둘째** IIFE 안에 있었다. 식별자가 그 자리에서 스코프에 없으므로 등록 줄이 로드 시점에 `ReferenceError` 를 던졌고, 예외가 그 IIFE 의 최상위 실행을 중단시켜 **그 줄 뒤에 있던 모든 등록이 조용히 일어나지 않았다.**

이 계열의 성질 셋이 이 SPEC 의 설계를 지배한다.

1. **로드 시점에 터진다.** 문법 오류가 아니므로 어떤 정적 문법 검사도 잡지 못한다.
2. **피해가 국소적이지 않다.** 터진 줄 자신이 아니라 **그 뒤의 무관한 등록들**이 죽는다. 그래서 증상이 원인과 멀리 떨어져 나타난다.
3. **어떤 신호도 남기지 않는다.** 콘솔을 열어 보지 않는 한 페이지는 그냥 "일부 버튼이 안 먹는" 상태가 된다.

### 기존 가드가 이 계열을 못 잡는다는 것은 측정된 사실이다

`internal/web/appjs_reinit_test.go` 는 `strings.Contains` 로 (a) `function initConsole(` 의 존재와 (b) 두 이벤트에 대한 `addEventListener("...", initConsole)` 바인딩 두 줄의 존재만 본다. **어느 등록이 어느 IIFE 안에 있는지는 묻지 않는다.** 따라서 `stampRefreshed` 등록을 첫째 IIFE 로 되돌려 놓아도 이 테스트는 초록으로 남는다.

---

## §B 측정된 사실 — 이 SPEC 의 인수조건이 딛고 선 근거

아래 수치는 전부 이 트리(`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1048`, branch `WT-iife-guard`, HEAD `d8304b49a`)에서 이 작성 중에 실측했다. 나중 독자가 다시 유도할 필요가 없도록 명령과 함께 적는다.

### B.1 IIFE 스팬

```bash
grep -n '^[^[:space:]]' internal/web/assets/app.js
```

관측: 컬럼 0 에서 열리는 IIFE 는 **정확히 둘**이고 스팬은 `[(39,660), (673,796)]` 이다. 파일 전체는 796 줄(`wc -l`). 39 줄 앞은 전부 주석이고, 660 과 673 사이는 구분 배너 주석이다.

### B.2 최상위 문장 형태 전수조사

두 IIFE 안에서 **정확히 2칸 들여쓰기**(= 그 IIFE 의 최상위)인 줄을 형태별로 분류했다. 2칸 들여쓰기가 최상위인 근거는 파일의 들여쓰기 단위가 2칸이라는 것(`sed -n '40,44p' … | cat -et` 로 확인: `  "use strict";`).

| 형태 | 줄 수 |
|---|---|
| comment | 110 |
| declaration (`function`) | 33 |
| declaration (`var`/`let`/`const`) | 7 |
| `document.addEventListener(...)` | 8 |
| `"use strict"` | 2 |
| closer (`}` / `});` 등 여러 줄 문장의 끝) | 38 |
| **그 밖의 형태** | **0** |

즉 **이 파일의 실제 최상위 표면에서, 선언도 주석도 `"use strict"` 도 아닌 문장은 전부 `document.addEventListener(...)` 호출이다.** 이것이 §C 에서 고르는 좁은 규칙이 "완전"한 근거다 — 규칙이 모든 교차 참조를 잡기 때문이 아니라, **이 파일에 실재하는 최상위 문장 형태를 다 덮기 때문이다.**

> 측정의 한계(명시): 이 전수조사는 파서가 아니라 **줄 형태** 기준이다. 각 문장의 첫 줄만 분류하고 이어지는 줄은 더 깊이 들여쓰기되어 집계에서 빠진다. 따라서 "그 밖의 형태 0" 은 "다른 형태의 문장을 시작하는 최상위 줄이 0개" 라는 뜻이다.

### B.3 bare-identifier 핸들러 인벤토리

```bash
grep -n 'addEventListener' internal/web/assets/app.js
```

파일 전체의 `addEventListener` 호출은 20개다. 그중 핸들러 인자가 **벌거벗은 식별자**인 것은 4개이고, 나머지 16개는 함수 표현식을 핸들러로 넘긴다(표현식은 자기 스코프를 데리고 다니므로 이 계열의 대상이 아니다).

| 등록 줄 | 식별자 | 등록 위치 | 선언 줄 | 선언 위치 | 비고 |
|---|---|---|---|---|---|
| 530 | `syncSegmentsVisibility` | IIFE #0 | 45 | IIFE #0 | **중첩 등록**(최상위 아님) |
| 548 | `initConsole` | IIFE #0 | 526 | IIFE #0 | 최상위 |
| 551 | `initConsole` | IIFE #0 | 526 | IIFE #0 | 최상위 |
| 727 | `stampRefreshed` | IIFE #1 | 711 | IIFE #1 | 최상위. **수리된 그 줄** |

선언 줄은 `grep -n -E '(function|var|let|const)[[:space:]]+(syncSegmentsVisibility|initConsole|stampRefreshed)\b'` 로 확인했다.

두 수치를 혼동하지 않도록 적는다: 최상위 `document.addEventListener` 줄은 **8개**지만, 그중 bare-identifier 를 쓰는 것은 **3개**(548·551·727)다. 네 번째(530)는 중첩 등록이다. 규칙이 실제로 검사하는 것은 "최상위 등록"이 아니라 "**어느 깊이에 있든 bare-identifier 를 핸들러로 넘기는 등록**"이고, 그래서 검사 대수가 4다.

### B.4 프로토타입 2방향 실측

규칙의 프로토타입을 이 트리에서 **양방향으로** 돌렸다.

**정방향 — 실제 파일:**

```
IIFE spans: [(39, 660), (673, 796)]
  ok line 530: syncSegmentsVisibility (IIFE#0, decl in [0])
  ok line 548: initConsole (IIFE#0, decl in [0])
  ok line 551: initConsole (IIFE#0, decl in [0])
  ok line 727: stampRefreshed (IIFE#1, decl in [1])
bare-identifier handlers checked = 4, violations = 0
EXIT=0
```

**역방향 — 돌연변이**(727 줄의 등록을 첫째 IIFE 안으로 되돌림):

```
IIFE spans: [(39, 661), (674, 796)]
  ok line 530: syncSegmentsVisibility (IIFE#0, decl in [0])
  ok line 548: initConsole (IIFE#0, decl in [0])
  ok line 551: initConsole (IIFE#0, decl in [0])
VIOLATION line 552: stampRefreshed — declared in IIFE#[1] not #0
bare-identifier handlers checked = 4, violations = 1
EXIT=1
```

**검사 대수가 양쪽 모두 4로 유지된 것이 양성 대조다.** 돌연변이에서 위반이 1로 올라간 것만으로는 "규칙이 핸들러를 실제로 들여다본다"가 서지 않는다 — 아무것도 매치하지 않으면서 통과하는 규칙도 정방향에서 위반 0을 낸다. 검사 대수가 0이 아니고 두 방향에서 같다는 사실이 그 공허한 통과를 배제한다.

### B.5 의존성 실측

```bash
awk '/^require \(/{f=1;next} /^\)/{f=0} f && !/\/\/ indirect/ && NF' go.mod | wc -l   # → 30
grep -c '^\t' go.mod                                                                   # → 89
grep -n -E '(dop251/goja|robertkrimen/otto|evanw/esbuild|rogchap/v8go|quickjs)' go.mod # → 0 hits
```

`go.mod` 의 require 항목은 총 **89**개이고 그중 **직접** 의존은 **30**개다. JS 엔진은 **없다**.

> 주의(다음 사람이 같은 grep 을 다시 돌릴 때): 단어 경계 없이 `otto` 만 찾으면 `github.com/atotto/clipboard` 가 걸려 **거짓 양성**이 난다. 모듈 경로로 정확히 재야 한다.

---

## §C 설계 결정 — 채택된 안과 그 근거 (재논의 대상 아님)

네 가지 안을 재고 하나를 골랐다. 이 절은 **결정의 기록**이며, 이 SPEC 은 대안을 열린 질문으로 되돌리지 않는다.

### 채택: package `web` 안의 Go 테스트로 구현하는 소스 수준 정적 규칙 — 신규 의존성 0

**규칙:** `addEventListener` 호출 중 핸들러 인자가 **벌거벗은 식별자**인 것은, 그 식별자가 **같은 최상위 IIFE 안 어딘가에** 선언되어 있어야 한다. 핸들러가 함수 표현식인 등록은 규칙 대상 밖이다.

**근거는 §B 의 측정 그 자체다.**

- **B.2** — 이 파일의 실제 최상위 표면에 다른 형태의 문장이 0개이므로, 좁은 규칙이 **현재 표면에 대해 완전하다.**
- **B.3/B.4** — 실재하는 4건 전부에 대해 규칙이 올바른 답을 내고, 실제로 발생했던 그 결함을 돌연변이로 재현하면 규칙이 식별자와 줄 번호를 지목하며 거부한다.
- **B.5** — 신규 의존성이 0이라는 것이 이 안의 결정적 이점이다.

### 기각된 대안과 기각 사유

| 대안 | 기각 사유 |
|---|---|
| 교차 경계 참조 **전면** 금지 | 진짜 스코프 분석이 필요하고, 그것은 JS 파서 의존성을 뜻한다. 정규식 근사는 문자열·주석·중첩 스코프에서 거짓 양성을 낸다 |
| in-process 런타임 실행 (goja + DOM 스텁) | 문장 형태와 무관하게 계열을 잡지만 **JS 엔진 의존성**을 들인다. B.5 기준 `go.mod` 에 JS 엔진은 없다 |
| 브라우저 프로브 (CDP, 실제 Chrome) | 일회성 진단으로는 유효하다. CI 가드로는 Chrome·python websockets·라이브 서버를 **새 CI 전제**로 들인다. 더구나 인용된 프로브(`.claude/worktrees/t1041/.moai/reports/t1041/browser-probe.py`)는 **다른 워크트리**에 있어 이 트리에서 도달조차 하지 않는다 |

---

## §D 요구사항 (GEARS)

### REQ-AIG-001 (Ubiquitous)

가드 테스트는 `internal/web/assets/app.js` 의 최상위 IIFE 스팬을, 컬럼 0 에서 열리고 닫히는 형태를 기준으로 스스로 산출해야 한다(shall). 스팬을 상수로 하드코딩해서는 안 된다(shall not) — 하드코딩된 줄 번호는 파일이 한 줄만 움직여도 조용히 다른 영역을 재게 된다.

### REQ-AIG-002 (Ubiquitous)

가드 테스트는, `addEventListener` 호출의 핸들러 인자가 벌거벗은 식별자인 모든 등록에 대해, 그 식별자의 선언이 **등록과 같은 최상위 IIFE 안에** 있는지 검사해야 한다(shall).

### REQ-AIG-003 (event-driven)

**When** 벌거벗은 식별자 핸들러의 선언이 등록과 다른 IIFE 에 있을 때, 가드 테스트는 실패하며 **식별자 이름과 등록 줄 번호를 모두** 담은 메시지를 내야 한다(shall). 줄 번호 없는 실패 메시지는 수리자를 파일 전체 탐색으로 돌려보낸다.

### REQ-AIG-004 (Ubiquitous)

가드 테스트는 자신이 실제로 검사한 bare-identifier 핸들러의 **개수를 단언해야 한다**(shall), 그 값이 0이 아님을 포함하여. 이것이 §B.4 의 양성 대조를 테스트 안으로 들여오는 조항이며, 아무것도 매치하지 않는 규칙이 공허하게 통과하는 것을 막는 유일한 장치다.

### REQ-AIG-005 (capability gate)

**Where** 핸들러 인자가 함수 표현식(`function (…) {…}` 또는 화살표 함수)인 경우, 가드 테스트는 그 등록을 검사 대상에서 제외해야 한다(shall) — 표현식은 자기 스코프를 데리고 다니므로 이 계열에 해당하지 않는다.

### REQ-AIG-006 (Ubiquitous)

가드 테스트는 `internal/web/restyle_test.go` 의 기존 헬퍼 `readEmbeddedAsset(t, "app.js")` 를 통해 소스를 읽어야 한다(shall). 두 번째 asset 리더를 추가해서는 안 된다(shall not). 이 헬퍼는 디스크가 아니라 **임베드된** 사본을 읽으므로, 가드가 재는 대상이 실제로 배포되는 바이트라는 성질이 딸려 온다.

### REQ-AIG-007 (Ubiquitous)

가드 테스트는 신규 `go.mod` 의존성을 도입해서는 안 된다(shall not). 표준 라이브러리(`regexp`, `strings`, `bufio`)만 쓴다.

### REQ-AIG-008 (Ubiquitous)

가드 테스트 소스는, 이 규칙이 **덮지 못하는 범위를 명시하는 주석**을 담아야 한다(shall) — §F 의 한계 문장 그대로.

### REQ-AIG-009 (Ubiquitous)

`internal/web/appjs_reinit_test.go` 의 기존 테스트 두 개(`TestAppJsInitConsoleBoundToBothEvents`, `TestAppJsHxBoostPreserved`)는 삭제되거나 약화되어서는 안 된다(shall not). 그 둘은 다른 성질(initConsole 이 두 이벤트에 **바인딩되어 있는가**)을 지키며, 이 SPEC 의 규칙(등록이 **어느 스코프에** 있는가)과 직교한다.

### REQ-AIG-010 (Ubiquitous)

이 SPEC 의 구현은 `internal/web/assets/` 아래 어떤 파일도 수정해서는 안 된다(shall not).

---

## §E 제외 범위

이 절은 무엇을 **만들지 않는지** 선언한다.

### Out of Scope — app.js 자체의 수정

- `internal/web/assets/app.js` 의 어떤 줄도 바꾸지 않는다. 그 디렉터리의 결함은 card t1041(agent-12) 소관이다.
- 리스너 등록의 재배치·리팩터링·중복 제거도 하지 않는다. 이 카드는 테스트만 더한다.

### Out of Scope — 일반 JS 스코프 분석

- IIFE 경계를 넘는 **모든** 참조를 잡는 검사는 만들지 않는다. §C 에서 기각됐다.
- JS 파서·AST·린터를 도입하지 않는다.
- 다른 정적 자산(`i18n.js`, CSS 등)으로 규칙을 확장하지 않는다.

### Out of Scope — 런타임·브라우저 검증

- goja 등 in-process JS 엔진을 도입하지 않는다.
- CDP 브라우저 프로브를 CI 경로에 넣지 않는다. 라이브 서버를 띄우지 않는다.

### Out of Scope — CI 설정 변경

- `.github/workflows/**` 를 건드리지 않는다. 가드는 기존 `go test ./internal/web/` 경로 안에서 돈다.
- 새 make 타깃·새 스크립트를 만들지 않는다.

### Out of Scope — 구현 세부

- 함수 이름·헬퍼 시그니처·정규식 문자열은 이 SPEC 이 정하지 않는다. run-phase 소관이다.

---

## §F 규칙이 덮지 못하는 것 (정직한 한계 선언)

이 가드는 **완전한 스코프 검사가 아니다.** 세 방향으로 놓친다.

1. **다른 형태의 최상위 문장.** 장래에 최상위에 `someIdentifier();` 나 `obj.method(otherIdentifier)` 같은 다른 형태의 문장이 생기고 그것이 IIFE 경계를 넘으면, 이 가드는 잡지 못한다. §B.2 의 전수조사는 **지금** 그런 문장이 0개임을 보인 것이지, 앞으로도 0이리라는 보장이 아니다.
2. **중첩 함수 안의 선언.** 규칙은 "같은 IIFE 안 **어느 깊이든**" 선언되어 있으면 통과시킨다. 따라서 어떤 중첩 함수 본문 안에만 선언된 식별자를 그 IIFE 의 최상위에서 등록하면, 실제로는 스코프 밖인데도 규칙은 통과시킨다. 이것은 편의가 아니라 **거짓 음성**이며, 의도한 맞교환이다 — 정확히 가르려면 §C 가 기각한 스코프 분석이 필요하다.
3. **함수 표현식 핸들러 안의 교차 참조.** REQ-AIG-005 로 명시적으로 제외했다. 표현식 본문이 다른 IIFE 의 식별자를 참조하면 잡히지 않는다.

가드가 덮는 것은 **bare-identifier 핸들러 계열**이고, 그것이 **실제로 일어난 계열**이다. 이 한계는 REQ-AIG-008 에 따라 테스트 소스 주석에도 그대로 들어간다.

---

## §G 인수조건

Tier S 이므로 AC 는 본래 여기 인라인이 정본이나, 이 카드의 검증이 2방향 성질에 통째로 걸려 있어 별도 `acceptance.md` 에 Given-When-Then 으로 펼쳐 둔다. AC 정본은 `acceptance.md` 다.

요약: (1) 실제 app.js 에 대해 위반 0 이며 **검사 대수가 0이 아님**을 함께 단언, (2) stampRefreshed 등록을 첫째 IIFE 로 옮긴 돌연변이가 식별자와 줄 번호를 지목하며 실패, (3) `internal/web` 의 기존 테스트가 계속 통과.

---

## §H 교차 참조

- `internal/web/assets/app.js` — 가드의 대상. **수정 금지**
- `internal/web/appjs_reinit_test.go` — 직교하는 기존 가드. 보존
- `internal/web/restyle_test.go:48` — `readEmbeddedAsset` 헬퍼의 정의 자리
- `18fa0d384` — 수리 커밋(이 SPEC 의 대상이 아님)
- card t1041 — `internal/web/assets/` 의 소관 카드
