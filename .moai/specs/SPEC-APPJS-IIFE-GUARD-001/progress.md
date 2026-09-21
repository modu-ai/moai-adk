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

## §E.1b Run-phase 진입 블로커 (2026-09-21)

> 이 절은 run-phase **증거가 아니라 run-phase 가 시작되지 않은 이유**다. §E.2 / §E.3 은
> manager-develop 소유이며 run 이 실제로 시작될 때까지 비워 둔다.

**Claim** — Kickoff 승인은 났으나 run-phase 에 진입하지 않았다. 배차문의 `cmd` 와 운영자
결정이 서로 다른 기전을 가리켜, 그대로 실행하면 자기 인수조건에 불합격하는 구현이 나온다.

**Evidence**

배차문(2026-09-21): `cmd: /moai run SPEC-APPJS-IIFE-GUARD-001` · `결정: 정적 참조 금지 방식은
채택 안 함 — 버튼 핸들러가 실제로 발화하는지를 런타임으로 잰다`.

이 SPEC 의 요구사항 10개는 전부 정적 규칙이며, 런타임 탐침 구현과 최소 6개가 충돌한다
(`grep -oE 'REQ-AIG-[0-9]+' spec.md | sort -u` → 001..010):

| 요구사항 | 조항 | 런타임 구현 시 |
|---|---|---|
| REQ-AIG-002 | bare identifier 핸들러의 IIFE 스코프 검사 | 기전 자체가 다름 |
| REQ-AIG-003 | 실패 메시지에 식별자명 + 등록 줄번호 | 런타임은 그 좌표를 못 냄 |
| REQ-AIG-004 | 검사한 핸들러 **개수** 단언 | 해당 개념 부재 |
| REQ-AIG-005 | 함수 표현식 핸들러 검사 제외 | 해당 개념 부재 |
| REQ-AIG-006 | `readEmbeddedAsset` 로 소스 읽기 | 런타임은 소스를 읽지 않고 실행함 |
| REQ-AIG-007 | 신규 `go.mod` 의존성 금지, 표준 라이브러리만 | 브라우저 구동이 이 조항 위반 |

CI 비용 측정(운영자 결정의 실제 입력):

- 워크플로 전수: `ls .github/workflows/` → 18개. 브라우저 grep 적중 1건은
  `test-install.yml:259` 의 주석 문구 `release browser endpoints` 로 **브라우저가 아니다**.
  즉 CI 브라우저 인프라는 **0 에서 시작**한다.
- 테스트 잡: `ci.yml:125` → `os: [ubuntu-latest]` 단독(SPEC-V3R6-CI-PR-SPEEDUP-001 로 좁혀짐),
  본체는 `ci.yml:210` 의 `go test ./...` 한 줄, race 는 `:292`.
- `e2e/` 트리: `e2e/cli/tux3_journeys.sh` 하나뿐 — 셸 CLI 여정, 브라우저 없음.
- t1041 탐침 의존성: 표준 `asyncio`/`json`/`sys`/`urllib.request` + **`websockets`**(서드파티,
  로컬 `15.0.1` 설치 확인) + CDP 가 `127.0.0.1:9333` 에 열린 Chrome + 라이브 서버.
- 따라서 CI 편입 비용 = ubuntu 잡에 Chrome 설치 · python 패키지 설치 · 서버 기동/종료 ·
  포트 할당이 붙고, `go test` 한 줄이던 잡이 별도 실행 계통을 갖는다.

**Baseline-attribution** — 전부 이 트리(`.claude/worktrees/t1048`, `WT-iife-guard`,
HEAD `8628672d8`)에서 2026-09-21 이 실행 중에 측정. 배차문 문면은 수신 메시지 그대로.

**Gaps**

- 런타임 방식을 **실제로 구현해 보지 않았다.** 충돌 판정은 SPEC 요구사항 문면과 탐침
  의존성 대조에서 나온 것이지, 구현해서 인수조건에 걸리는 것을 본 것이 아니다.
- CI 편입의 **소요 시간·요금을 재지 않았다.** 늘어나는 항목을 셌을 뿐이다.
- `t.Skip` 대안이 실제로 CI 에서 조용히 건너뛰는지 **돌려서 확인하지 않았다.**

**Residual-risk**

- 경로 A/B/C 중 무엇을 고르든 SPEC 재작성이 선행되어야 하며, 재작성 없이 구현하면
  인수조건이 구현을 검증하지 못한다 — 통과가 아무것도 보장하지 않는 상태가 된다.
- 이 기록 시점에 구현 코드는 0줄이다. 판별식은 **커밋 열거**다 —
  `git log develop..HEAD --name-only` → 이 브랜치의 커밋 2개가 건드린 파일은
  `.moai/specs/SPEC-APPJS-IIFE-GUARD-001/` 4개뿐, Go 파일 0개. 되돌리는 비용이 가장 싼
  지점이며, 결정이 늦어질수록 싸지지는 않는다.

  > **쓰면 안 되는 판별식**: `git diff --stat develop..HEAD -- 'internal/**/*.go'`.
  > `git diff` 는 `A..B` 철자를 포함해 전부 **트리 비교**라 내 변경과 남의 변경을 가르지
  > 못한다. 이 기록을 쓰는 동안 실제로 걸렸다 — 그 명령이 8행을 냈고, 확인해 보니 develop 이
  > `d8304b49a` → `db2270127` 로 11커밋 움직이며 들어온 **다른 레인의 변경**이었다.
  > 내 기반과 develop 이 같은 순간에만 우연히 0을 낸다.

**배차문 결함 기록 (리드 요청)** — 리드가 운영자 결정을 전달하면서 그 결정이 기존 SPEC 과
양립하는지 대조하지 않았고, `cmd` 에 기존 SPEC ID 를 그대로 실었다. 리드가 결함을 인정했다.
같은 계열: 배차문의 수치·전제를 열어 보지 않고 옮긴 건. 이 카드에서만 두 번째다(첫 번째는
새 워크트리 기반을 `d8304b49a` 로 적었으나 실제로는 `origin/develop` = `116820f40` 에서
태어난 건, `git rev-list --count HEAD..develop` → 9 로 확인 후 ff 로 교정).

---

## §E.2 Run-phase Evidence

> 운영자 결정으로 경로 C 가 채택되어 §E.1b 의 블로커가 해소됐다 — 이 카드는 **정적
> 가드**로 남는다. SPEC 요구사항 10개는 변경 없이 그대로 유효하다.

**Claim** — `internal/web/appjs_iife_scope_test.go` 가 신설되어 AC-AIG-001..007 이 전부
통과한다. 규칙은 2방향(정방향 + 돌연변이)으로 **Go 에서 다시** 측정됐고, 가드 자신도
3건의 돌연변이로 공허하지 않음이 확인됐다.

### AC 매트릭스

| AC | 상태 | 판정 명령 | 관측된 출력 |
|---|---|---|---|
| AC-AIG-001 | PASS | `go test ./internal/web/ -run 'IIFE\|Scope' -count=1 -v` | `--- PASS: TestAppJsBareHandlersDeclaredInSameIIFE (0.00s)` · 로그 `forward: bare-identifier handlers checked = 4, violations = 0, total addEventListener calls = 20` |
| AC-AIG-002 | PASS | 같은 명령 | `--- PASS: TestAppJsBareHandlerScopeRuleRejectsCrossIIFERegistration (0.00s)` · 로그 `mutant: checked = 4 (forward 4), violations = 1: [line 660: stampRefreshed — declared in IIFE#[1], not IIFE#0]` |
| AC-AIG-003 | PASS | `go test ./internal/web/ -count=1 -timeout 30m` | `ok  github.com/modu-ai/moai-adk/internal/web  25.301s`, EXIT=0, `--- FAIL` 0행. `git diff --stat -- internal/web/appjs_reinit_test.go` → 출력 0바이트 |
| AC-AIG-004 | PASS | 위 정방향 로그 | `checked = 4`, `total addEventListener calls = 20` → 4 ≠ 20, 표현식 핸들러 16건 제외 확인. 테스트 안 단언은 `checked >= total` 이면 실패 |
| AC-AIG-005 | PASS | `grep -n -E '\b(39\|660\|673\|727\|796)\b' internal/web/appjs_iife_scope_test.go` | 출력 없음, grep exit=1 (적중 0). 주석에도 없음 |
| AC-AIG-006 | PASS | `git diff --stat -- internal/web/assets/ go.mod go.sum .github/workflows/` | 세 명령 모두 출력 0바이트. `git status --short` → `?? internal/web/appjs_iife_scope_test.go` 한 줄뿐 |
| AC-AIG-007 | PASS | 소스 상단 주석 판독 | spec.md §F 세 한계(다른 형태의 최상위 문장 / 중첩 함수 안의 선언 / 표현식 핸들러 본문의 교차 참조) + "이 가드는 완전한 스코프 검사가 아니다" 취지 + 네 번째 축(발화 여부, card t1060) 모두 기재 |

### 요구사항 대조

| REQ | 충족 지점 |
|---|---|
| REQ-AIG-001 | `scanTopLevelIIFESpans` 가 컬럼 0 의 `(function` / `})();` 로 스팬 산출. 돌연변이 삽입 위치도 스팬에서 **산출**(`spans[0].end - 1`). AC-AIG-005 grep 0적중이 기계적 근거 |
| REQ-AIG-002 | `scanBareHandlerScope` 가 등록 스팬과 선언 스팬을 대조 |
| REQ-AIG-003 | `scopeViolation.String()` 이 식별자명 + 등록 줄 번호를 모두 담고, 역방향 테스트가 두 토큰의 포함을 단언 |
| REQ-AIG-004 | `checked == 0` 이면 `t.Fatal` (RED 에서 실제로 발화 — 아래 RED 증거) |
| REQ-AIG-005 | `bareHandlerPattern` 이 핸들러 인자를 `[A-Za-z_$][A-Za-z0-9_$]*\s*\)` 로만 매치 — 표현식은 구조상 걸리지 않음 |
| REQ-AIG-006 | `readEmbeddedAsset(t, "app.js")` 사용. 두 번째 리더 없음 |
| REQ-AIG-007 | import 는 `fmt` / `regexp` / `strings` / `testing` 표준 라이브러리뿐. `go.mod` / `go.sum` diff 0바이트 |
| REQ-AIG-008 | 파일 상단 주석 (AC-AIG-007) |
| REQ-AIG-009 | `appjs_reinit_test.go` diff 0바이트, 두 테스트 모두 패키지 스위트에서 통과 |
| REQ-AIG-010 | `internal/web/assets/` diff 0바이트. 돌연변이는 메모리 안의 문자열로만 합성 |

### RED 증거 (GREEN 이전, 축자)

`scanBareHandlerScope` 의 본문을 `return nil, 0` 스텁으로 두고 실행:

```
$ go test ./internal/web/ -run 'TestAppJsBareHandler' -count=1 -v
=== NAME  TestAppJsBareHandlersDeclaredInSameIIFE
    appjs_iife_scope_test.go:206: rule matched ZERO bare-identifier handlers — a rule that inspects nothing passes the violations==0 assertion vacuously; this is the failure this guard exists to prevent
--- FAIL: TestAppJsBareHandlersDeclaredInSameIIFE (0.00s)
=== NAME  TestAppJsBareHandlerScopeRuleRejectsCrossIIFERegistration
    appjs_iife_scope_test.go:239: mutant should produce exactly 1 violation, got 0: []
--- FAIL: TestAppJsBareHandlerScopeRuleRejectsCrossIIFERegistration (0.00s)
FAIL	github.com/modu-ai/moai-adk/internal/web	0.708s
EXIT=1
```

두 단언이 **각각** 발화한 것이 이 RED 의 값이다. 특히 첫 번째는 "아무것도 매치하지
않는 규칙이 위반 0 을 내는" 공허한 통과를 실제로 재현해 잡았다.

> 저작 순서 기록(정직성): Go 는 테스트와 피검 함수가 같은 패키지에 있어야 하므로 이
> 파일은 한 번에 작성됐고, 위 RED 는 **규칙 본문을 스텁으로 되돌려** 생성했다. 즉
> "테스트를 먼저 쓴 뒤 구현했다"가 아니라 "구현을 제거하면 테스트가 실패함을 보였다"가
> 정확한 서술이다. 단언이 실제로 문다는 성질은 동일하게 확립된다.

### 가드 자신에 대한 돌연변이 검증 (3방향)

| 돌연변이 | 주입 지점 | 관측 |
|---|---|---|
| M-a 규칙이 아무것도 매치하지 않음 | `scanBareHandlerScope` → `return nil, 0` | 정방향 FAIL(`checked == 0`), 역방향 FAIL(위반 0). EXIT=1 (위 RED 블록) |
| M-b 스코프 검사가 **항상 통과** | `found := false` → `found := true` | **정방향은 PASS**, 역방향만 FAIL: `mutant should produce exactly 1 violation, got 0: []`. EXIT=1 |
| M-c 합성이 조용히 no-op | 제거·삽입 분기를 죽임 | 전제 단언이 발화: `mutation synthesis produced a byte-identical copy — the mutant would prove nothing`. EXIT=1 |

M-b 가 이 가드의 핵심 근거다 — **정방향만으로는 잡히지 않고 역방향 테스트만 잡았다.**
정방향은 어차피 위반이 0이므로 "항상 통과하는 규칙"과 구별되지 않는다. M-c 는
AC-AIG-002 가 요구한 `t.Fatal` 전제 단언이 실제로 무는지를 보인다.

> M-c 에서 관측된 부수 사실: 합성 함수의 **이른 `return src`** 는 byte-identity 단언에
> 도달하기 전에 빠져나가므로, 그 경우 실패 메시지는 전제 단언이 아니라 위반 개수
> 단언에서 난다. 테스트는 두 경우 모두 실패하지만 진단성이 다르다 — 남겨 둔다.

### 기계 검증 배치 (축자)

```
$ go build ./...                                  → EXIT=0, 출력 없음
$ gofmt -l internal                               → 출력 0바이트
$ go vet ./internal/web/                          → EXIT=0, 출력 없음
$ go test ./internal/web/ -run 'IIFE|Scope' -count=1 -v
                                                  → EXIT=0
   --- PASS: TestAppJsBareHandlersDeclaredInSameIIFE (0.00s)
   --- PASS: TestAppJsBareHandlerScopeRuleRejectsCrossIIFERegistration (0.00s)
   (동반 매치 5건도 전부 PASS: TestGLMEffortScopeBadge / TestSaveScopeBoundary /
    TestD1HaikuRelaxationIsScopedToOverrides / TestScopeContractEditableSections /
    TestScopeContractExclusions)
   ok  github.com/modu-ai/moai-adk/internal/web  0.482s
$ go test ./internal/web/ -count=1 -timeout 30m   → EXIT=0
   ok  github.com/modu-ai/moai-adk/internal/web  25.301s
   grep -c -- '--- FAIL' → 0        (요약 행 `ok` + EXIT=0 으로 판정; 타임아웃 아님)
$ go test -cover ./internal/web/ -count=1 -timeout 30m
   ok  github.com/modu-ai/moai-adk/internal/web  25.581s  coverage: 67.8% of statements
$ (같은 실행에서, 신규 파일을 패키지 밖으로 잠시 옮겨 측정한 pre-change baseline)
   ok  github.com/modu-ai/moai-adk/internal/web  25.822s  coverage: 67.8% of statements
```

### Python 프로토타입 대비 Go 측정값

| 항목 | Python 프로토타입 | Go 실측 | 일치 |
|---|---|---|---|
| IIFE 스팬 (정방향) | `[(39,660),(673,796)]` | 동일 (컬럼 0 스캔으로 산출) | 예 |
| 정방향 검사 대수 | 4 | 4 | 예 |
| 정방향 위반 | 0 | 0 | 예 |
| 돌연변이 검사 대수 | 4 | 4 | 예 |
| 돌연변이 위반 | 1 (`stampRefreshed`) | 1 (`stampRefreshed`) | 예 |
| 돌연변이 등록 줄 번호 | 552 | 660 | **아니오 — 아래 설명** |
| 전체 `addEventListener` | 20 | 20 | 예 |

유일한 불일치는 **돌연변이의 등록 줄 번호**(552 vs 660)이고, 이것은 규칙의 차이가
아니라 **삽입 지점의 차이**다. 프로토타입은 `initConsole` 의 `htmx:afterSettle` 등록
바로 뒤에 끼워 넣었고, 이 구현은 첫째 IIFE 의 **닫는 줄 바로 앞**에 끼워 넣는다.
후자를 고른 이유는 삽입 위치를 스팬 스캐너에서 **산출**할 수 있어 다른 테스트가
지키는 성질(`appjs_reinit_test.go` 의 `initConsole` 바인딩 줄)에 결합하지 않기
때문이다. 두 위치 모두 첫째 IIFE 의 최상위이므로 규칙이 내는 판정은 같다.

**Python `re` → Go RE2 이식 위험은 실현되지 않았다.** 프로토타입의 `\w` 를 그대로
옮기지 않고 명시적 문자류 `[A-Za-z_$][A-Za-z0-9_$]*` 로 썼기 때문이다 — Go 의 `\w` 는
`$` 를 포함하지 않아 `$`-포함 JS 식별자를 조용히 놓쳤을 것이다. 선언 탐색의 경계도
같은 이유로 `\b` 가 아니라 명시적 non-identifier 문자류를 쓴다. 다만 현재 app.js 의
식별자 4개에는 `$` 가 없어 **이 차이가 이 트리에서 판정을 가르지는 않았다** — 위험을
피한 것이지 위험이 발화한 것을 잡은 것이 아니다.

### 검사 대수 고정 vs `> 0` — 채택된 판단

`> 0` (느슨하게) 를 채택했고, 이유는 테스트 주석에 기록됐다. 요지: 정확한 값에
고정하면 핸들러가 **정당하게** 추가될 때마다 테스트가 깨지고, 그러면 테스트는 "고칠
것"이 아니라 "숫자를 맞춰 줄 것"이 된다 — 숫자를 습관적으로 올리는 순간 그 단언의
탐지 가치는 0이 된다. 여기에 더해 AC-AIG-004 를 `checked < total` 로 단언해 표현식
제외를 값 고정 없이 기계적으로 보인다.

[정정 — 레인 검증에서 잡음] 이 판단의 근거로 처음 적혔던 문장 "느슨하게 두어 놓치는
것(검사 대수가 조용히 3으로 줄어드는 퇴행)은 AC-AIG-002 의 대수 동일성 단언이
부분적으로 흡수한다" 는 **거짓이다.** 대수 동일성은 **한 실행 안에서** 정방향과
돌연변이의 대수를 비교한다. 표면이 커밋 사이에 4→3 으로 줄면 돌연변이 쪽 대수도 함께
3이 되므로 동일성은 그대로 통과한다. 그 단언이 잡는 것은 "돌연변이가 규칙이 보는 표면
자체를 바꿨는가" 이지 "표면이 지난 커밋보다 줄었는가" 가 아니다.

따라서 축소 퇴행은 **감수한 것이지 흡수된 것이 아니며**, 이 트리에 그 구멍을 닫는
장치는 없다. 같은 거짓 문장이 테스트 주석에도 실려 있었고 함께 정정했다
(`internal/web/appjs_iife_scope_test.go`). 이 카드의 [HARD] 가 "가드가 자기 유효 범위를
스스로 선언한다" 이므로, 그 선언에 남은 거짓 근거는 없는 보호를 있다고 믿게 만든다.

**Baseline-attribution** — 전부 worktree
`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1048`, branch `WT-iife-guard`,
HEAD `a86d8e12c` 에서 2026-09-21 이 실행 중에 측정. app.js 는 796줄로 §E.1 측정 당시와
동일(develop 흡수가 이 파일을 건드리지 않음). 커버리지 67.8% 는 **이 실행에서 양방향**
(신규 파일 포함 / 제외)으로 측정한 값이라, 변화 없음이 추론이 아니라 관측이다.

**Gaps**

- **런타임 동작은 검증되지 않았다.** 브라우저를 띄우지 않았고 리스너가 실제로 붙는지
  재지 않았다. 이 가드는 소스 문자열만 본다.
- **spec.md §F 의 세 한계에 대한 음성 결과는 측정되지 않았다.** 세 방향으로 돌연변이를
  돌려 "잡지 못함"을 보이지 않았다 — 선언했을 뿐이다.
- **`golangci-lint` 를 돌리지 않았다.** `gofmt` / `go vet` 만 관측했다. 린트 baseline
  대비 NEW 여부는 이 실행에서 측정되지 않았다.
- **크로스 플랫폼 빌드(`GOOS=windows`)를 재지 않았다.** 변경이 `_test.go` 한 파일이고
  빌드 태그·syscall 을 쓰지 않으나, 그것은 추론이지 측정이 아니다.
- **다른 정적 자산(`i18n.js` 등)에 같은 계열이 있는지 재지 않았다.**
- **`$`-포함 식별자에 대해 규칙이 옳게 동작하는지 재지 않았다.** 현재 app.js 에 그런
  식별자가 없어 합성 입력으로 확인하지 않았다. `\w` 를 피한 것은 예방이지 관측이 아니다.

**Residual-risk**

- 스팬 스캐너는 **줄 접두사** 기준(`(function` / `})();`)이다. 장래에 컬럼 0 에서
  다른 형태로 열거나 닫는 IIFE 가 생기면 스팬 산출이 어긋난다. 다만 그때는 조용히
  틀리지 않고 스캐너가 즉시 다른 답을 내므로 시끄러운 실패다.
- 정규식은 **한 줄 안에서만** 매치한다. `addEventListener(` 호출이 여러 줄로 쪼개져
  있으면 검사 대수에서 빠진다. 현재 20건은 전부 한 줄이다.
- 선언 탐색은 문자열·주석을 구분하지 않는다. 주석 안의 `function stampRefreshed` 같은
  문자열이 선언으로 오인될 수 있다 — 거짓 **음성** 방향(위반을 놓침)이다.

> **축 인계**: "경계를 넘는 참조가 없다"는 "핸들러가 실제로 발화한다"를 함의하지 않는다.
> 발화 여부 축은 **card t1060** 소관으로 넘긴다 — 이 카드에서 닫지 않으며, 그 사실은
> 테스트 소스 상단 주석 4번 항목에도 같은 문장으로 기재돼 있다.

---

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-21
run_commit_sha: pending-backfill-run
run_status: audit-ready
ac_pass_count: 7
ac_fail_count: 0
preserve_list_post_run_count: 0        # internal/web/assets/ + appjs_reinit_test.go 모두 diff 0바이트
l44_pre_commit_fetch: not-performed    # 레인-로컬 워크트리, 공유 체크아웃 편집 없음
l44_post_push_fetch: not-applicable    # push 하지 않음 (리드 일괄 소관)
new_warnings_or_lints_introduced: none-observed  # gofmt 0행 · go vet EXIT=0 (golangci-lint 미측정 — §E.2 Gaps)
cross_platform_build:
  darwin_arm64: pass                   # go build ./... EXIT=0
  windows_amd64: not-measured          # §E.2 Gaps
total_run_phase_files: 1               # internal/web/appjs_iife_scope_test.go (신규)
m1_to_mN_commit_strategy: single-commit  # plan.md §E 단일 마일스톤
```

---

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
