---
id: SPEC-APPJS-FIRE-GUARD-001
title: "app.js 버튼 핸들러 런타임 발화 가드 — 정적 경계 가드의 초록이 실제 발화를 함의하지 않음을 브라우저에서 재단다"
version: "0.1.0"
status: draft
created: 2026-09-22
updated: 2026-09-22
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: "internal/web"
lifecycle: spec-anchored
tags: "web, app.js, runtime, browser, cdp, fire, guard, e2e, t1060"
era: V3R6
tier: M
related_specs: [SPEC-APPJS-IIFE-GUARD-001]
---

# SPEC-APPJS-FIRE-GUARD-001 — app.js 버튼 핸들러 런타임 발화 가드

## HISTORY

| 날짜 | 버전 | 변경 | 주체 |
|---|---|---|---|
| 2026-09-22 | 0.1.0 | plan-phase 최초 작성 (card t1060) | manager-spec |
| 2026-09-22 | 0.1.0 | plan-audit iter-1 FAIL(0.75) 수리 — D1~D9 (RED-now 4요소 셀·REQ 매핑·인용 정본 일원화·skip 사유 영어) | manager-spec |

---

## §A 배경 — 정적 가드의 초록이 함의하지 않는 것

형제 SPEC `SPEC-APPJS-IIFE-GUARD-001`(completed)은 `internal/web/assets/app.js` 의 bare-identifier 핸들러가 IIFE 경계를 넘지 않음을 **정적**으로 보장한다. 이 보장은 "경계를 넘는 참조가 없다"이지 "**버튼을 누르면 실제로 반응한다**"가 아니다. 두 명제 사이에는 적어도 네 갈래 간극이 열려 있다 — 셀렉터가 실제 DOM 과 어긋나는 경우, hx-boost 바디 스왑 뒤 재결합이 일어나지 않는 경우, 등록 이외의 원인으로 초기화 도중 예외가 나는 경우, 그리고 그 어떤 정적 규칙도 문법으로 표현할 수 없는 나머지 전부다.

이 카드의 계기는 형제 SPEC 의 배차 결정에 기록돼 있다 — t1048 진행 기록 §E.1b: 「정적 참조 금지 방식은 채택 안 함 — 버튼 핸들러가 실제로 발화하는지를 런타임으로 잰다」. 이 SPEC 은 그 결정의 구현 SPEC 이다. **정적 규칙을 다시 만들지 않는다** — 형제 SPEC 은 읽기 전용 참조이고, 두 가드는 직교하는 두 계층이다(방어 깊이). 정적 가드가 없어도 이 가드는 성립하고, 그 반대도 마찬가지다.

형제 SPEC 의 인수조건은 이 간극을 정직하게 남겨뒀다 — 그 acceptance.md §D: 「런타임 동작은 검증되지 않는다 … 실제 브라우저에서 리스너가 붙는지는 여전히 수동 E2E 소관이다」. 이 SPEC 이 그 소관을 기계화한다.

---

## §B 측정된 사실 — 이 SPEC 의 인수조건이 딛고 선 근거

아래 수치는 전부 이 트리(`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1060`, branch `WT-appjs-handler-guard`)에서 이 작성 중에 실측했다. 명령과 함께 적는다 — 나중 독자가 다시 유도할 필요가 없도록.

### B.1 정방향 프로토타입 — 이 트리의 오늘 기준선

t1041 의 브라우저 탐침(`.claude/worktrees/t1041/.moai/reports/t1041/browser-probe.py` — 다른 카드의 트리에 있는 읽기 전용 참조, `/tmp` 사본으로 실행)을 이 트리에서 빌드한 바이너리(`go build -o /tmp/t1060-probe/moai ./cmd/moai` → v3.1.3, exit 0)에 대해 돌렸다. 서버는 `moai web --port 18441 --no-open --no-reuse`, 브라우저는 로컬 Chrome `--headless=new --remote-debugging-port=9333`.

아래 JSON 은 progress.md §E.1 의 정본 출력과 **바이트 동일**하다(iter-2 D7 수리 — 한 측정, 어느 문서에서도 같은 인용).

```json
{
  "label": "t1060-baseline",
  "port": "18441",
  "p1_load_referenceerrors": [],
  "p1_has_glm_btn": true,
  "p2_revealed_hidden_before": true,
  "p2_revealed_hidden_after": false,
  "p2_glm_handler_fired": true,
  "p3_swap_clicked": true,
  "p3_swap_referenceerrors": [],
  "p3_url_after_swap": "/todo",
  "p4_load_referenceerrors": [],
  "p3_has_copy_btn": true,
  "p4_label_after_click": "✓",
  "p4_label_before_click": "Copy",
  "p4_copy_handler_fired": true
}
```

예시 탐침의 5 지표가 전부 발화하고, 로드와 스왑 양쪽에서 ReferenceError 가 **0건**이다. 이것이 이 가드가 앞으로 지킬 이 트리의 기준선이다.

### B.2 역방향 프로토타입 — 역사적 결함 재도입 시 지표가 뒤집힌다

727 행의 등록(`document.addEventListener("htmx:afterSettle", stampRefreshed);`)을 첫째 IIFE 안(합성 좌표 552 행)으로 옮기는 돌연변이를 적용하고 **재빌드**(app.js 는 `//go:embed` 로 바이너리에 실린다)해 같은 프로브를 돌렸다:

아래 JSON 은 progress.md §E.1 및 acceptance.md §B2 E5 의 정본 출력과 **바이트 동일**하다(iter-2 D7 수리 — 편집 주석은 JSON 밖으로 뺐다).

```json
{
  "label": "t1060-mutation",
  "port": "18442",
  "p1_load_referenceerrors": [
    "Uncaught ReferenceError: stampRefreshed is not defined\n    at http://127.0.0.1:18442/static/app.js:552:49\n    at http://127.0.0.1:18442/static/app.js:661:3"
  ],
  "p1_has_glm_btn": true,
  "p2_revealed_hidden_before": true,
  "p2_revealed_hidden_after": true,
  "p2_glm_handler_fired": false,
  "p3_swap_clicked": true,
  "p3_swap_referenceerrors": [
    "Uncaught ReferenceError: stampRefreshed is not defined\n    at http://127.0.0.1:18442/static/app.js:552:49\n    at http://127.0.0.1:18442/static/app.js:661:3"
  ],
  "p3_url_after_swap": "/todo",
  "p4_load_referenceerrors": [
    "Uncaught ReferenceError: stampRefreshed is not defined\n    at http://127.0.0.1:18442/static/app.js:552:49\n    at http://127.0.0.1:18442/static/app.js:661:3"
  ],
  "p3_has_copy_btn": true,
  "p4_label_after_click": "Copy",
  "p4_label_before_click": "Copy",
  "p4_copy_handler_fired": false
}
```

**발화 지표가 전부 뒤집혔다** — glm 공개 버튼과 copy 버튼이 모두 죽고, 로드·스왑·재로드 3Phase 전부에서 ReferenceError 가 찍혔다(세 `p*_referenceerrors` 문자열은 전부 동일한 내용 — `app.js:552` 의 `stampRefreshed is not defined`). 이 재실행은 HEAD `d726ac709` 위에서 이루어졌고, 탐침은 `PROBE_EXIT=0` 을 냈다 — exit 1 을 요구하는 AC-AFG-002 에서 올바른 이유의 적색이다(acceptance.md §B2 E5). 복원 후 `cmp` 로 byte 동일(`RESTORED_BYTE_IDENTICAL`)을 확인했고 작업 트리는 깨끗하다.

B.1 과 나란히 읽어야 하는 관측 하나: **`p1_has_glm_btn` 은 돌연변이에서도 `true` 였다.** 버튼이 DOM 에 존재하는 것과 발화하는 것은 독립된 축이라는 것이 이 카드의 요지 그 자체다. (glm 버튼이 158 행의 등록 — 돌연변이 지점보다 앞 — 을 두고도 죽은 것은 관측이고, 「552 행 이후의 최상위 효과가 전멸했다」는 기전 가설은 표시일 뿐이며 run-phase 에서 확인 대상이다. 추론을 측정으로 읽지 않는다.)

### B.3 app.js 상호작용 표면 인벤토리

```bash
grep -n -E "addEventListener\((['\"])(click|submit|change|input)" internal/web/assets/app.js
```

관측(이 트리): click/change 계열 등록 줄 **13개** — 73, 83, 109, 158, 327, 352, 406, 414, 473, 513, 530, 603, 648. 전체 `addEventListener` 는 20개(형제 SPEC §B.3 과 일치). 이 13개가 매니페스트의 조사 대상 표면이다 — 각 그룹이 (페이지, 셀렉터, 관측 가능한 효과) 삼중으로 매니페스트에 **들어가거나 명시적으로 제외 사유와 함께 빠진다**.

### B.4 기반시설 실측

| 항목 | 측정 | 명령/출처 |
|---|---|---|
| Chrome | `/Applications/Google Chrome.app` 존재 | `ls -d` |
| python websockets | 15.0.1 | `python3 -c "import websockets; …"` |
| `golang.org/x/net` | v0.58.0, `// indirect` 마커 없음 = 직접 의존 | `sed -n '25,30p' go.mod` |
| 서버 테스트면 | `web.NewServer(cfg)` + `Handler() http.Handler` (server.go:122/162), `Start` for tests | 소스 |
| CLI 플래그 | `web [--port N] [--no-open] [--no-reuse]` — 기본 포트 3041, 자동 브라우저 열기 끌 수 있음 | internal/cli/web.go |
| 브라우저 무관성 확인 | `openDefaultBrowser` 는 열기 헬퍼일 뿐 — 이 SPEC 의 대상 아님 | internal/web/browser.go |
| CI 좌표 | `.github/workflows/ci.yml:125` = `os: [ubuntu-latest]`, `:210` = 단일 `go test -json … ./...` 행; jobs: detect(44)/test(116)/test-race(256)/test-skip-marker(321)/test-integration(373)/lint(427)/build(474)/constitution-check(544) | sed/grep |
| 브라우저 인프라 | `git grep 9333 -- internal/web/` → 0행(exit 1) — 서버 쪽에 CDP 전제 없음 | git grep |

### B.5 탐침의 출처와 신뢰도

탐침의 원형은 t1041 의 `browser-probe.py`(의존성: stdlib asyncio/json/sys/urllib.request + 서드파티 `websockets`)다. 그 탐침은 실제 결함을 잡은 이력이 있다 — t1041 판정서 E3: main 판 재빌드에서 5 지표가 요구대로 죽고 develop 판에서 전부 살았다(양군 대조). **필드에서 검증된 계측기를 그대로 데려오는 것이 이 설계의 기둥이다.** 오늘의 프로토타입(B.1/B.2)은 같은 탐침이 이 트리에서도 양방향으로 올바르게 판정함을 재확인했다.

> 정직한 기록: 예시 탐침은 **보고만 하고 exit 0 으로 끝난다** — B.2 의 돌연변이에서도 `PROBE_EXIT=0` 이었다. 판정은 사람이 JSON 을 읽어야 했다. 이 SPEC 이 요구하는 3값 exit 계약(REQ-AFG-005)은 커밋되는 탐침에 **새로** 더해지는 작업이다.

---

## §C 설계 결정 — 채택된 안과 기각된 안

### 채택: 전용 CI job + 환경 게이트된 Go 드라이버 테스트 + 커밋된 Python 탐침

구조는 네 조각이다.

1. **탐침** `internal/web/testdata/appjs_fire_probe.py` — B.5 의 예시에서 파생. (페이지, 셀렉터, 관측 효과) 매니페스트 + 3값 exit 계약을 가진다.
2. **드라이버** `internal/web/appjs_fire_guard_test.go` — `MOAI_BROWSER_GUARD=1` 환경 게이트가 있고 Chrome·python3·websockets 가 발견될 때만 전체 사이클을 실행한다. 그 외에는 **무엇이 빠졌는지 이름으로 지목하는 skip** 을 출력한다.
3. **돌연변이기** `internal/web/testdata/appjs_fire_mutation.py` — B.2 의 합성 절차를 기계화하고, 대상 줄을 못 찾으면 즉시 실패한다.
4. **CI job** `.github/workflows/ci.yml` 의 신설 `test-browser` — 기존 job 을 한 줄도 고치지 않고, 그린 단계(게이트 켠 테스트)와 레드 단계(돌연변이 → 재빌드 → 탐침 exit 1 기대 → 복원 → byte 동일 검증)를 모두 돌린다.

**환경 게이트가 존재하는 이유:** 게이트 없이 `go test` 안에 이 테스트를 두면, 그 테스트의 거동이 「러너에 Chrome 이 깔려 있는가」라는 우리가 통제하지 않는 사실에 기생한다. 게이트는 기존 `test` job 의 전제를 러너 이미지 내용과 무관하게 고정한다 — 전제 고립이 우연이 아니라 기계적으로 성립하게 하는 장치다.

### 기각된 대안과 기각 사유

| 대안 | 기각 사유 |
|---|---|
| **Chrome-absent `t.Skip` 만 (새 CI 전제 0 유지)** | 기존 job 의 전제는 안 바뀌지만 가드는 CI 에서 **영원히 발화하지 않는다** — 런타임 발화 축의 절반을 포기하는 형태다(t1048 조사에서도 「목적의 절반 포기」로 기각된 바 있다). 이 SPEC 은 이 형태를 채택하지 않되, 게이트된 skip 이 로컬 열위 환경에서 취하는 **같은 모양**을 정직하게 문서화한다: skip 은 「이 환경에서 재지 않았다」의 기록이지 통과가 아니며, CI 판정면은 전용 job 이 유일하게 운반한다. |
| **기존 ubuntu test job 에 브라우저 단계 편입** | Chrome 공급 + pip 패키지 + 서버 수명 + 포트 할당이 기존 `go test` 전제에 합류한다(t1060 배차문에 실린 agent-13 실측: 18개 워크플로, 브라우저 그렙 적중은 주석 오탐 1건, 브라우저 인프라는 0에서 출발). 기존 판정면의 실패 의미가 흐려진다. |
| **Go 네이티브 CDP 클라이언트 (x/net/websocket)** | `golang.org/x/net` 이 이미 직접 의존이라(B.4) 신규 의존성 없이 가능은 하다. 그러나 필드에서 검증된 탐침 대신 **미실측 계측기 200~300줄**을 새로 쓰는 길이고, 계측기 자체의 버그가 제품 결함으로 위장한다. B.2 로 입증된 도구를 버리고 새로 만들 이유가 없다. |
| **in-process `httptest` 서버로 실바이너리 대체 (드라이버 전면)** | `Handler()` 로 가능하다. 다만 프로토타입이 실측한 표면은 **실바이너리**의 것이고, 드라이버가 바이너리를 스스로 빌드하는 것은 테스트 안의 빌드라는 새로운 부담이다. 절충을 §C 아래 「서버 거점」에 명시한다. |

### 서버 거점 (절충 명시)

- **로컬/게이트된 드라이버**: in-process 서버(`NewServer` + `Handler()`, 임시 포트) — 테스트가 스스로 수명을 `t.Cleanup` 으로 정리할 수 있어야 하기 때문이다. 바이너리 빌드를 테스트 안에서 요구하지 않는다.
- **CI 레드 단계(돌연변이)**: **실바이너리** — embed 는 컴파일타임이라 돌연변이 자산을 재려면 재빌드가 원리상 필수다(B.2 절차 그대로).
- 두 서면이 같은 app.js 표면을 서빙하는지는 run-phase M1 이 등가를 **측정**해서 확인한다(같은 탐침, 같은 지표). 측정 전까지 이 절차는 가정이 아니라 계획이다.

---

## §D 요구사항 (GEARS)

### REQ-AFG-001 (capability gate)

**Where** `MOAI_BROWSER_GUARD=1` 이 설정돼 있고 Chrome·python3·websockets 가 모두 발견될 때, 드라이버 테스트는 서버 기동 → headless Chrome 기동 → 탐침 실행 → 보고 판정의 전체 사이클을 실행해야 한다(shall). 그 전제 중 하나라도 빠지면, 테스트는 건너뛰되 **무엇이 빠졌는지 이름으로 지목하는 skip 사유를 출력해야 한다**(shall) — 사유 메시지는 영어로 출력되며(`error_messages: en` 정책) 게이트 변수명 `MOAI_BROWSER_GUARD` 또는 빠진 전제의 이름 `chrome` / `python3` / `websockets` 이 그대로 보여야 한다(shall). 사유 없는 skip 은 이 가드의 생존을 읽는 사람에게 「통과」와 구분되지 않는 침묵을 준다.

### REQ-AFG-002 (ubiquitous)

탐침은 커밋된 Python 스크립트(`internal/web/testdata/appjs_fire_probe.py`)여야 하고(shall), t1041 의 검증된 탐침에서 파생해야 하며(shall), `go.mod` 에 신규 의존성을 추가해서는 안 된다(shall not). 탐침의 서드파티 의존은 `websockets` 하나뿐이다.

### REQ-AFG-003 (ubiquitous)

탐침은 (페이지, 셀렉터, 관측 가능한 효과) 삼중의 **명시적 매니페스트**를 가져야 하고(shall), B.3 의 인벤토리 13줄 각 그룹이 매니페스트 항목이거나 명시된 사유의 제외 항목이어야 한다(shall). 매니페스트 항목 수가 0인 상태로 판정이 통과해서는 안 된다(shall not) — 0행 매니페스트의 초록은 아무것도 재지 않은 초록이다.

### REQ-AFG-004 (event-driven)

**When** 매니페스트의 셀렉터가 서빙된 페이지에서 아무 원소도 매치하지 않을 때, 탐침은 실패 지표로 **그 셀렉터 이름을 보고해야 한다**(shall). 셀렉터의 스테일은 red 여야지, 조용한 0 이 아니어야 한다 — 이 가드의 매니페스트가 시간이 지나도 살아 있는지를 이 조항이 보증한다.

### REQ-AFG-005 (ubiquitous)

탐침은 3값 exit 계약을 가져야 한다(shall): 0 = 전 지표 발화 + ReferenceError 0건, 1 = 지표 붕괴 또는 셀렉터 미달, 2 = 기계결함(CDP·서버 도달 불가 등). 판정이 JSON 육안 해석에 묶여 있으면 CI 에서 red 는 red 로 보이지 않는다 — B.5 의 정직한 기록이 그 증거다.

### REQ-AFG-006 (ubiquitous)

드라이버 테스트는 실제 서버 표면을 띄우고(§C 서버 거점), headless Chrome 을 띄우고, 탐침을 실행해, 모든 지표의 발화와 load 시점 ReferenceError 0건을 단언해야 한다(shall). 서버·Chrome·탐침 프로세스의 정리는 `t.Cleanup` 에 등록돼야 한다(shall) — 뒤에 붙은 kill 은 정리가 아니며, 어떤 종료 경로에서도 프로세스가 남으면 안 된다.

### REQ-AFG-007 (ubiquitous)

매니페스트는 hx-boost 바디 스왑 **뒤에** 최소 1개 지표를 행사해야 한다(shall). 역사적 결함 가족이 스왑 뒤의 무관한 등록을 죽였고(B.2 에서 스왑 Phase 가 같은 ReferenceError 를 찍은 것이 그 재확인이다), 스왑 뒤 발화야말로 이 가드가 정적 가드와 구별되는 축 중 하나다.

### REQ-AFG-008 (event-driven)

**When** B.2 의 돌연변이(등록 줄을 다른 최상위 IIFE 로 이동)가 재도입되고 바이너리가 재빌드될 때, 탐침은 exit 1 로 실패해야 하며 그 보고에 최소 1개 지표의 미발화 또는 load 시점 ReferenceError 를 담아야 한다(shall). 돌연변이기가 대상 줄을 찾지 못하면 **즉시 실패해야 한다**(shall) — 합성 실패가 「위반 0」으로 나타나 조용한 초록이 되는 것을 막는 전제 단언이다.

### REQ-AFG-009 (ubiquitous)

돌연변이 방향이 끝나면 원본 `app.js` 가 byte 동일하게 복원됐음을 `cmp`(또는 동등한 byte 비교)와 `git status` 로 검증해야 하고(shall), 복원·검증 없이 job 이 끝나서는 안 된다(shall not).

### REQ-AFG-010 (ubiquitous)

CI 는 신설 전용 job(`test-browser`)에서 이 가드를 실행해야 하고(shall), 기존 job 과 기존 step 의 전제를 변경해서는 안 된다(shall not). Chrome 공급은 버전 고정 다운로드(Chrome-for-Testing 고정본 또는 명시된 동등 수단 — 러너 이미지에 사전 설치된 Chrome 을 신뢰하지 않는다)여야 하고(shall), `websockets` 버전은 고정돼야 한다(shall).

### REQ-AFG-011 (ubiquitous)

이 SPEC 의 구현은 `internal/web/assets/` 아래 어떤 파일도 수정한 채로 남겨서는 안 된다(shall not) — 돌연변이는 job 실행 안에서만 존재하고, 트리에 남는 것은 탐침·드라이버·돌연변이기·job 정의뿐이다.

### REQ-AFG-012 (ubiquitous)

매니페스트가 행사하는 상호작용은 **영속화 부수효과가 없어야 한다**(shall) — 가시성 전환, 라벨 변경, 탭 전환 같은 관측 가능하되 되돌아가는 효과만 행사한다. 저장 버튼처럼 프로젝트 설정을 쓰는 제어는 매니페스트에서 제외하거나, 행사해야 한다면 드라이버가 **일회용 프로젝트 사본** 위에서 서빙해야 한다(shall). 탐침이 누른 버튼이 실제 프로젝트의 구성 파일을 바꾸는 일은 데이터 손실 경로다.

### REQ-AFG-013 (ubiquitous)

가드 소스(드라이버·탐침 상단 주석)는 이 가드가 덮지 못하는 것을 명시해야 한다(shall) — §F 의 한계 목록과, 정적 가드와의 직교성(어느 쪽 초록도 다른 쪽을 함의하지 않는다)을 포함해.

---

## §E 제외 범위

이 절은 무엇을 **만들지 않는지** 선언한다.

### Out of Scope — 정적 가드의 재저작

- `SPEC-APPJS-IIFE-GUARD-001` 의 규칙·테스트(`internal/web/appjs_iife_scope_test.go`)를 수정·재작성하지 않는다. 두 가드는 직교 계층이다.
- IIFE 스코프 분석을 런타임 가드 안에 다시 구현하지 않는다.

### Out of Scope — app.js 자체의 수정

- `internal/web/assets/` 아래 파일은 한 줄도 바꾸지 않는다(REQ-AFG-011). 핸들러 재배치·리팩터링은 이 카드의 소관이 아니다.

### Out of Scope — 일반 브라우저 테스트 인프라

- Playwright·Puppeteer·selenium 같은 브라우저 프레임워크를 도입하지 않는다. 탐침은 stdlib + `websockets` 로 CDP 를 직접 말한다.
- 시각적 회귀(스크린샷 비교), 접근성 감사, 다른 브라우저(WebKit/Firefox) 지원을 하지 않는다.
- `e2e/` 의 셸 CLI 저니를 확장하지 않는다 — 이 가드의 거점은 `internal/web` + `test-browser` job 이다.

### Out of Scope — 기존 CI 전제 변경

- 기존 8개 job(detect/test/test-race/test-skip-marker/test-integration/lint/build/constitution-check)의 step·전제를 변경하지 않는다(REQ-AFG-010).
- 필수 검사(required check)로의 승격은 이 SPEC 이 정하지 않는다 — 운영자 판정 사항이다.

### Out of Scope — 구현 세부

- 셀렉터 문자열·매니페스트의 최종 항목 목록·Chrome 고정 버전 숫자·정확한 워크플로 YAML 은 run-phase 소관이다. 이 SPEC 은 계약과 판별식만 정한다.

---

## §F 이 가드가 덮지 못하는 것 (정직한 한계 선언)

이 가드는 **완전한 발화 보증이 아니다.** 네 방향으로 놓친다.

1. **매니페스트에 없는 상호작용.** 13줄 인벤토리(B.3) 각 그룹이 항목이거나 명시적 제외일 뿐, 제외된 것은 재지 않는다. 제외 사유가 취약하면 가드도 취약하다.
2. **탐침이 행사하지 않는 경로.** 클릭·스왑·재클릭의 순서를 바꾸면 나타날 수 있는 상태는 재지 않는다 — 탐침은 정해진 시나리오 하나를 운반한다.
3. **Chrome 한 종류.** WebKit·Firefox 에서만 나타나는 발화 차이는 이 가드의 밖이다.
4. **정적 축의 결함을 런타임 그림자로만 본다.** IIFE 경계 결함은 B.2 처럼 런타임에서도 관측되지만, 그 축의 정밀 판정(식별자·줄 번호 지목)은 여전히 정적 가드의 몫이다. 이 가드의 red 가 곧 원인 보고는 아니다.

이 가드가 덮는 것은 **사용자가 버튼을 눌렀을 때 반응하는가**의 최소 표면이고, 그것은 어떤 정적 규칙도 대신 봐 주지 않는 축이다. 이 한계는 REQ-AFG-013 에 따라 소스 주석에도 그대로 들어간다.

---

## §G 인수조건

Tier M 이므로 AC 정본은 별도 `acceptance.md` 다. 요약: (1) 게이트 켠 드라이버가 실트리에서 전 지표 발화로 통과하되 매니페스트 0행·ReferenceError 를 함께 거부, (2) 돌연변이 재도입 시 탐침 exit 1 — 양방향 모두 관측된 출력으로, (3) 게이트 없는 실행은 사유를 이름으로 대는 skip, (4) 기존 job·assets·go.mod 무변경.

---

## §H 교차 참조

- `SPEC-APPJS-IIFE-GUARD-001` — 정적 형제 가드(completed). 직교 계층. 읽기 전용 참조.
- `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1041/.moai/reports/t1041/browser-probe.py` — 탐침 원형. **다른 워크트리의 읽기 전용 파일** — 참조만 하고 수정·이동하지 않는다.
- t1041 판정서 E3 — 탐침이 실제 결함을 잡은 양군 대조 기록.
- t1048 진행 기록 §E.1b — 이 카드의 계기가 된 배차 결정.
- `internal/web/server.go:122/162` — `NewServer`/`Handler()` 서버 테스트면.
- `internal/cli/web.go` — `web [--port] [--no-open] [--no-reuse]` 플래그 표면.
- `.github/workflows/ci.yml` — `test-browser` job 의 합류 지점(기존 job 8개 무변경).
- card t1060 — 이 SPEC 의 배차 카드.
