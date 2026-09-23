---
id: SPEC-APPJS-FIRE-GUARD-001
title: "app.js 버튼 핸들러 런타임 발화 가드 — 정적 경계 가드의 초록이 실제 발화를 함의하지 않음을 브라우저에서 재단다"
version: "0.3.0"
status: in-progress
created: 2026-09-22
updated: 2026-09-23
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: "internal/web"
lifecycle: spec-anchored
tags: "web, app.js, runtime, browser, cdp, fire, guard, e2e, t1060, t1106, t1108"
era: V3R6
tier: M
related_specs: [SPEC-APPJS-IIFE-GUARD-001]
amendment_of: SPEC-APPJS-FIRE-GUARD-001
---

# SPEC-APPJS-FIRE-GUARD-001 — app.js 버튼 핸들러 런타임 발화 가드

## HISTORY

| 날짜 | 버전 | 변경 | 주체 |
|---|---|---|---|
| 2026-09-22 | 0.1.0 | plan-phase 최초 작성 (card t1060) | manager-spec |
| 2026-09-22 | 0.1.0 | plan-audit iter-1 FAIL(0.75) 수리 — D1~D9 (RED-now 4요소 셀·REQ 매핑·인용 정본 일원화·skip 사유 영어) | manager-spec |
| 2026-09-23 | 0.2.0 | in-place amendment — 검증 거부 배너의 **화면 도달**을 이 가드로 편입 (card t1106). REQ-AFG-014·015 신설, §B.6 실측 추가, §C 결정 5 추가. 기존 REQ-AFG-001~013 번호·문언 불변 | manager-spec |
| 2026-09-23 | 0.2.0 | 개정분 plan-audit iter-1 FAIL(0.86) 수리 — D1(사본 base 부재 시 계약 미정의 + AC-001 의 8/9 모호) · D2(무쓰기 제외 목록의 범위 무제한) · D3(AC-008 한계 축 4→7 스테일) · D4/N3(조건 표식 수 3 vs 2 모순). REQ·AC 신설 0건, 기존 번호 불변 | manager-spec |
| 2026-09-23 | 0.2.0 | 개정분 plan-audit iter-2 PASS-WITH-DEBT(0.90) 의 차단급 잔여 2건 인라인 수리 — N-1(AC-AFG-001 (a) 의 판정을 개수 `> 0` 에서 「사본 서빙 표식이 없는 항목 전부」 술어 일치 + 운전 수·제외 이름 보고로 승격) · N-2(§A 의 「AC-001~008 문언 불변」 서사를 실제 개정 세 곳으로 정정). `acceptance.md` 절 편집만 — REQ·AC 신설 0건, `spec.md` 본문 무수정 | manager-spec |
| 2026-09-23 | 0.3.0 | in-place amendment 2 — post-swap 지표의 **전제 교정** (card t1108). REQ-AFG-007 문언 개정(실제 hx-boost 스왑 + `htmx:afterSettle` **이벤트** 대기; 개정 전 문언은 조문 안에 보존), REQ-AFG-016 신설(스왑 자기확인), §B.7 실측 추가, §B.1·§B.2·REQ-AFG-007 근거 서술에 날짜 붙은 사후 정정 주석, §C 개정 결정 2 추가, AC-AFG-014·015·016 신설, AC-AFG-001 (c)·AC-AFG-006·AC-AFG-013 문언 보강. 나머지 REQ-AFG-001~006·008~015 번호·문언 불변 | manager-spec |
| 2026-09-23 | 0.3.0 | 개정 2 plan-audit iter-1 FAIL(0.84) 수리 (card t1108) — D1(차단): AC-AFG-015 에 다리별 역방향 고정물 추가((a)(b)(c) 동시 거짓의 전체 이동 판 + 다리 단독 합성 보고서 + (c) 단독·(d) 단독 라이브 사본). D2: AC-AFG-014 돌연변이를 둘로 정의(`/settings` 경로 폴링 판, 대기 제거 판), 만료 대기 적색 명시. D3: B1 분기 술어를 10/10 으로 고정, 증폭 효과 관측을 필수화. D4: 올린 `-timeout` 이 job `timeout-minutes: 20` 을 넘으면 blocker. D5: 한계로 기록. D6: REQ-AFG-007 (3) 에 exit 1 과 지목 대상. D7: REQ-AFG-007 형식 표기. D8: `app.go:288`. D9: AC-AFG-016 매핑에 REQ-AFG-007·016. REQ·AC 신설 0건 | manager-spec |
| 2026-09-23 | 0.3.0 | 리드 blocker 결정 반영 (card t1108) — B1: AC-AFG-014 에 고정 측정 순서(스로틀 12배 단독 먼저 → 판정되지 않을 때에만 두 판에 동일한 settle 지연 증폭 + 효과 관측 → 증폭 적용 여부·값 기록). B3: AC-AFG-016 에 그린 단계 `-timeout` 조건부 상향(병합 트리 실측이 10m 초과 시 그 한 줄만, 근거 인용). B2: 원문 유지. REQ·AC 신설 0건 | manager-spec |

### Amendments

| 항목 | 값 |
|---|---|
| prior completed version | 0.1.0 |
| prior_completed_sha | `0e2377323` (`docs(SPEC-APPJS-FIRE-GUARD-001): sync-phase artifacts — 3-phase close (card t1060)`) |
| rationale | card t1105 가 `handleSave` 의 검증 거부 경로를 400 → 200 으로 고쳤지만, 그 검증은 **단위 테스트(응답 본문 안에 배너가 있는가)** 뿐이었다. 배너가 **브라우저에서 실제로 칠해지는가**는 t1105 판정서가 Gap 으로 명시 기록했고, 그 sync-audit 이 미측정 구간을 「boost 200 이 스왑되는가」가 아니라 「**이 본문이** 칠해지는가」로 좁혔다. 이 가드가 이미 브라우저에서 발화를 재는 유일한 계측기이므로, 그 구간은 새 러너가 아니라 이 SPEC 의 확장으로 닫는다. |
| scope | 기존 탐침(`internal/web/testdata/appjs_fire_probe.py`)과 드라이버(`internal/web/appjs_fire_guard_test.go`)의 **확장**만. 신규 러너 금지. REQ-AFG-012 는 약화되지 않으며, 그 조문이 이미 이름한 두 경로 중 **둘째 경로(일회용 프로젝트 사본)** 를 처음으로 사용한다. |

#### 개정 2 — 0.3.0 (card t1108)

| 항목 | 값 |
|---|---|
| prior completed version | 0.1.0 (변동 없음 — 0.2.0 개정은 run-phase 까지 develop 에 착지했으나 sync 로 닫히지 않았다. 마지막 완료 판은 여전히 위 표의 0.1.0 / `0e2377323` 이다) |
| amendment base | branch `WT-popover-swap-flake`, HEAD `52a486635`, tree `895ad8954` (0.2.0 run-phase 와 그 F1 수리가 병합된 로컬 develop) |
| rationale | card t1108 판정서(`.moai/reports/t1108/verdict.md`)가 `popover_after_swap` 의 간헐 실패 원인을 확정했다 — 5단계가 클릭하는 `a[href="/todo"]` 에는 `hx-boost` 조상이 없어 클릭이 **메인 프레임 전체 이동**이고, 탐침은 URL 변경만 기다린 채 6단계로 넘어가 새 문서의 초기화보다 먼저 클릭할 수 있다. 따라서 REQ-AFG-007 이 전제한 「hx-boost 스왑 뒤」는 이 SPEC 의 어느 판에서도 실제로 측정된 적이 없다. 리드 결정: 방향 (a) — REQ-AFG-007 의 의도(실제 스왑 뒤 재바인딩)를 유지하고, 전제가 조용히 재발하지 않도록 스왑 자체를 탐침이 확인하게 한다 |
| scope | REQ-AFG-007 문언 개정 + REQ-AFG-016 신설. 매니페스트의 스왑 항목·스왑 뒤 항목 정의 변경(스왑 링크는 `/settings` 위 boost 링크, 스왑 뒤 페이지는 `/settings`). CI 는 `.github/workflows/ci.yml:672` 그린 단계 한 줄만 바꾼다 — `-run` 정규식을 넓히고(`'AppJsHandlersFire'` → `'AppJs.*Fire'`), 병합 트리 실측이 10분을 넘을 때에만 같은 줄의 `-timeout` 을 측정값 + 여유로 올린다(리드 결정 B3). `--primary-entries-only` 3곳(734/745/767)은 유지. `app.js`·제품 템플릿·`INVENTORY_TOTAL` 불변 |

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

> **사후 정정 (2026-09-23, card t1108) — B.1·B.2 의 「스왑」은 스왑이 아니었다.** 위 두 JSON 과 서술은 관측 당시의 원문이므로 고치지 않는다. 다만 그 해석 하나를 정정한다. 두 프로토타입의 `p3_swap_*` 단계와 `p3_url_after_swap: "/todo"` 는 **hx-boost 바디 스왑이 아니라 메인 프레임 전체 이동**의 관측이다. 클릭하는 `a[href="/todo"]` 에는 `hx-boost` 조상이 없고, 콘솔 어디에도 boost 된 `/todo` 링크가 없다(§B.7.1·§B.7.3). 따라서 B.2 에서 「스왑 Phase 가 같은 ReferenceError 를 찍었다」는 관측은 **새 문서가 `app.js` 를 다시 실행하며 로드 시점 예외를 다시 던진 것**이다. 스왑 뒤 재바인딩을 관측한 것이 아니다. 이 해석에 기대던 REQ-AFG-007 의 근거 서술도 같은 날짜로 정정했다(§D REQ-AFG-007).

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

### B.6 개정분(card t1106)의 실측 — 검증 거부 배너의 미측정 구간

아래는 전부 이 트리(`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1106`, branch `WT-fireguard-reject-submit`, HEAD **`176d8b658`**)에서 2026-09-23 개정 작성 중에 실행한 관측이다.

**B.6.1 — 고쳐진 것과 고쳐졌다고 관측된 것.** card t1105 는 `handleSave` 의 검증 거부 경로를 `http.StatusOK` 로 옮겼다(`internal/web/handlers.go`, 거부 블록 주석이 그 이유를 담고 있다 — 페이지가 `hx-boost="true"` 이고 고정 임베드된 htmx 2.0.4 의 기본 responseHandling 표가 4xx 를 `{swap:false}` 로 답하므로, 400 으로 그린 배너는 클라이언트가 버린다). 그 수리의 검증면은 단위 테스트다(`internal/web/transport400_swap_contract_test.go`, `internal/web/transport400_characterization_test.go`, `internal/web/transport400_htmx_contract_test.go` — 이 트리에 존재함을 `ls` 로 확인). 단위 테스트가 답하는 명제는 「응답 **본문**에 배너와 per-field 에러가 들어 있다」이고, 「그 본문이 **화면에 칠해진다**」가 아니다.

**B.6.2 — 오늘의 매니페스트는 그 경로를 행사하지 않는다.** 탐침의 `--lint-manifest` 자기검증 출력(단일 호출, exit 0):

```
LINT OK: 8 entries + 7 exclusions cover 13 inventory groups; all effects within ['clipboard', 'label', 'swap', 'tab', 'visibility']; post-swap entry present
```

허용목록 5종 어디에도 제출 계열이 없다. 제외 목록 7건 역시 폼 상태·파괴적 동작 사유로 제출을 배제한다.

**B.6.3 — 드라이버는 오늘 실저장소 루트를 서빙한다.** `grep -n 'ProjectRoot:' internal/web/appjs_fire_guard_test.go` → `204:		ProjectRoot:    findRepoRoot(t),` (exit 0). 격리돼 있는 것은 `ProfileBaseDir`(`t.TempDir()`)뿐이다. 즉 오늘의 배선에서 제출 계열을 행사하면 **실제 저장소의 구성 파일이 쓰기 대상이 된다** — REQ-AFG-012 가 막는 바로 그 경로다.

**B.6.4 — 인벤토리는 이 개정으로 변하지 않는다.** `grep -n -E "addEventListener\((['\"])(click|submit|change|input)" internal/web/assets/app.js` → 여전히 **13줄**(73/83/109/158/327/352/406/414/473/513/530/603/648). 검증 거부는 app.js 의 등록 지점이 아니라 htmx boost + 서버 렌더 표면이므로, 신설 항목은 기존 `swap_todo_nav` 와 같이 `line_group: None` 계열이다. 따라서 `INVENTORY_TOTAL = 13` 과 lint 의 커버리지 산식(비-None line_group 집합만 센다)은 **건드리지 않는다**.

**B.6.5 — 정직한 기록: 이 개정의 세 기준은 모두 오늘 빈 스윕이다.** 세 판정 선택자 전부 `[no tests to run]` 으로 exit 0 을 낸다(원문 출력·exit·트리 SHA 는 acceptance.md §B2 증거 장부 E6/E7/E8 에 4요소로 적었다). 빈 스윕의 초록은 통과가 아니라 미측정이다.

### B.7 개정 2(card t1108)의 실측 — post-swap 단계의 전제가 성립하지 않았다

출처는 둘이다. 원인 측정은 card t1108 판정서 `.moai/reports/t1108/verdict.md` 와 그 로그 `.moai/reports/t1108/logs/` 이다. 판정서 §2~§4 는 tree `0c70186fd` 에서 쟀고, 측정 스크립트는 저장소 밖 스크래치에서 실행했다. 드라이버 표면 측정은 `logs/gate-inprocess.log` 이며 HEAD `52a486635`(tree `895ad8954`)에서 쟀다. 코드 좌표는 이 개정의 기준 트리 HEAD `52a486635` 에서 다시 읽었다. **판정서가 인용한 탐침 줄 번호(`:398-403`·`:410-411`·`:474`)는 t1106 병합 전 트리의 좌표이므로 여기서는 쓰지 않는다.**

**B.7.1 — 5단계의 클릭은 전체 이동이다.** 5단계(`internal/web/testdata/appjs_fire_probe.py:531-540`)는 첫 `a[href="/todo"]` 를 클릭한다. 그 뒤로는 `location.pathname == "/todo"` 만 폴링한다(`:538`, `poll(cdp, "location.pathname", "/todo", timeout=8.0)`). 1~4단계가 `/settings` 에 머물러 있으므로 클릭은 `/settings` 위에서 일어난다. 반면 매니페스트 항목 `swap_todo_nav` 는 `page: "/"` 를 선언한다(`:172-178`). 선언된 페이지와 실제로 행사하는 페이지가 다르다는 사실도 함께 기록한다. 판정서 §2 의 시간순 계측 5회는 모두 새 문서로의 이동이었고, `htmx:afterSwap`·`htmx:afterSettle` 은 **0회**였다. 대기 조건이 참이 된 순간에 `DOMContentLoaded` 는 5회 모두 아직 나지 않았다.

**B.7.2 — 6단계는 초기화보다 먼저 클릭할 수 있다.** 6단계(`:542-551`)는 곧바로 패널 `.hidden` 을 읽고 트리거를 클릭한다. 주석(`:543-544`)은 「initConsole re-runs on htmx:afterSettle」이라며 스왑을 전제하지만, 탐침 소스에서 `htmx:afterSettle` 을 담은 줄은 그 주석 하나뿐이다(`grep -n 'htmx:afterSettle' internal/web/testdata/appjs_fire_probe.py` → `544:    # re-runs on htmx:afterSettle).`, exit 0). 즉 기다리는 코드가 없다. `app.js` 는 `defer` 로 로드되고(`internal/web/shell.templ:106`), `wirePopovers` 는 호출 시점 DOM 의 트리거에 직접 핸들러를 붙인다(`internal/web/assets/app.js:65`). `initConsole` 은 `DOMContentLoaded` 에서 호출된다(`app.js:548`). 그래서 파싱이 덜 끝났으면 패널이 없어 「selector matched nothing」이 되고, 파싱은 끝났으나 초기화 전이면 「indicator did not fire」가 된다. 재현율은 판정서 §3 에 있다(탭 한정 CPU 스로틀 12배에서 macOS 30/30, Linux 7/10 실패). `DOMContentLoaded` 까지 기다린 대조군은 같은 조건에서 30/30, 10/10 발화했다. CI 사유 「indicator did not fire」는 판정 결과로 재현되지 않았다 — 해당 상태가 관측된 데서 한 **추론**이다(판정서 §6).

**B.7.3 — 콘솔에는 boost 된 `/todo` 링크가 없다.** 템플릿의 `hx-boost` 는 두 곳뿐이다. 설정 폼 `internal/web/root.templ:56` 과 프로필 팝오버 `internal/web/shell.templ:307` 이다(`grep -n 'hx-boost' internal/web/*.templ` 가 주석 1행과 이 두 행을 낸다). 런타임 인벤토리(판정서 §4)에서도 `/`·`/settings`·`/todo`·`/specs` 어디에도 boost 된 `/todo` 링크는 0개였다.

**B.7.4 — 실제 스왑 경로는 드라이버 표면에 존재한다.** 측정은 `logs/gate-inprocess.log` 이다. 드라이버와 **같은** in-process 표면(`startFireGuardServer`, 비어 있는 `ProfileBaseDir`)에서 임시 게이트 테스트로 쟀고, 테스트 파일은 측정 뒤 삭제했다(소스 사본 `logs/zz_t1108_gate_test.go.txt`). `/settings` 는 설정 폼 안에서 boost 된 `/settings?tab=audit` 링크 4개와 `/settings?tab=mcp` 링크 8개를 렌더한다. `/settings?tab=audit` 클릭의 결과는 다음과 같다. 메인 프레임 이동은 0회였고 같은 문서에 머물렀다. `DOMContentLoaded`·`htmx:afterSwap`·`htmx:afterSettle` 이 기록됐다. `htmx:afterSettle` **이벤트**를 기다린 뒤 팝오버가 CPU 스로틀 12배에서 3/3 발화했다. 실바이너리 표면에서는 판정서 §4(a) 가 macOS 12배 5/5, Linux 12배 3/3 발화를 기록했다.

**B.7.5 — 탭 선택자는 조회 전용이다.** `?tab=` 은 GET 처리기에서 `view.ActiveTab = r.URL.Query().Get("tab")` 로 읽힐 뿐이다(`internal/web/handlers.go:263`). `?profile=` 도 조회 값이다(`internal/web/app.go:288` — 함수 `selectedProfile` 은 `:287`, `internal/web/screens.go:50`). 따라서 boost 링크 클릭은 영속화 부수효과가 없고, REQ-AFG-012 의 비영속 계열(`swap`)에 그대로 머문다. (리드 배차문은 `?tab=` 의 근거로 `app.go:287`·`screens.go:50` 을 댔으나, 두 줄이 읽는 것은 `profile` 이다. `tab` 을 읽는 줄은 `handlers.go:263` 이다.)

**B.7.6 — 인벤토리와 커버리지 산식은 변하지 않는다.** 스왑 항목은 `line_group: None` 이다(`:173`). 스왑 뒤 항목은 `line_group: 73` 인데, `popover_open`(`:133`)과 **같은** 그룹이라 커버리지 집합에 새 원소를 더하지 않는다. 오늘의 자기검증 출력은 다음과 같다(단일 호출, exit 0): `LINT OK: 9 entries + 7 exclusions cover 13 inventory groups; effects within unconditional ['clipboard', 'label', 'swap', 'tab', 'visibility'] or conditional ['validation-reject'] (conditional entries carry ['requires_sandbox_serving', 'requires_no_write_assertion']); post-swap entry present`. 두 항목의 `page`·`selector` 가 바뀌어도 `INVENTORY_TOTAL = 13` 을 바꿀 이유가 없다.

**B.7.7 — CI 는 t1106 의 제출 계열 테스트를 돌리지 않는다.** `.github/workflows/ci.yml:672` 의 그린 단계 선택자는 `-run 'AppJsHandlersFire'` 이다. 이 선택자가 고르는 테스트는 `go test ./internal/web/ -list 'AppJsHandlersFire'` 로 쟀을 때 `TestAppJsHandlersFireRuntime`·`TestAppJsHandlersFireSelectorMiss` 두 개뿐이다. 넓힌 선택자 `'AppJs.*Fire'` 는 여덟 개를 고르며, 여기에 `TestAppJsFireValidationRejectPaints`·`TestAppJsFireValidationRejectNoWrites` 가 들어 있다(원문은 acceptance.md §B2.2 E13·E14). 한편 맨손 탐침 호출 3곳의 `--primary-entries-only`(`grep -c -- '--primary-entries-only' .github/workflows/ci.yml` → `3`)는 t1106 F1 수리의 결과다. 이 플래그를 빼면 표식 항목이 사본 base 없이 운전돼 exit 2 가 난다(progress.md §E.3 개정분 `run_phase_correction_after_f1`).

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

### 채택(개정, card t1106): 제출 1건을 **일회용 사본 위에서만** 행사하고, 무쓰기를 탐침이 단언한다

**문제:** 검증 거부 배너가 브라우저에서 칠해지는지 재려면 실제로 폼을 제출해야 한다. 제출은 영속화 계열이고, REQ-AFG-012 가 정면으로 다루는 데이터 손실 경로다.

**REQ-AFG-012 는 이 길을 이미 이름하고 있다.** 그 조문은 저장 계열을 평평하게 금지하지 않는다 — 두 경로를 제시한다: 매니페스트에서 **제외**하거나, 행사해야 한다면 드라이버가 **일회용 프로젝트 사본** 위에서 서빙한다. 드라이버 소스 상단 주석도 같은 것을 지시문으로 적어 두었다(「if a manifest entry ever needs a persistence-family control, the driver MUST switch it to a one-off project copy instead of loosening the manifest」). 이 개정은 그 둘째 경로를 **처음으로 사용**하는 것이지, 조문을 완화하는 것이 아니다.

**채택안은 두 절반이 함께다. 어느 한쪽도 다른 쪽을 대신하지 못한다.**

1. **일회용 프로젝트 사본 위에서만 서빙한다.** 구성 손실을 「탐지」가 아니라 **구조적으로 불가능**하게 만든다.
2. **무쓰기를 탐침이 스스로 단언한다.** 「거부 경로는 아무것도 쓰지 않는다」는 **주어진 전제가 아니라 시험 대상 명제**다. 회귀해서 실제로 썼다면, 바이트 비교만으로는 **실저장소 구성이 이미 망가진 뒤에야** 잡는다. 사후 탐지는 REQ-AFG-012 가 막으려는 손실 자체를 막지 못한다. (1)이 노출을 없애고 (2)가 명제를 계속 측정한다.

**허용목록이 얻는 것은 넓은 권한이 아니라 좁은 한 종(種)이다.** 새 효과 종별 `validation-reject` 는 **두 조건이 함께 성립할 때만** 유효하다 — 일회용 사본 서빙 + 탐침의 무쓰기 단언. 미래의 어떤 항목도 실저장소 루트를 서빙하면서 이 종별을 주장할 수 없어야 하며, 그 불가분성 자체가 기계 판정 대상이다(REQ-AFG-014, AC-AFG-012).

**기각된 대안**

| 대안 | 기각 사유 |
|---|---|
| 무쓰기 단언만(카드 문면의 최소안) | 위 2번의 이유 — 사후 탐지는 손실을 막지 못한다. 실저장소를 서빙한 채 회귀를 「발견」하는 것은 REQ-AFG-012 의 목적을 달성하지 못한다. |
| 사본 서빙만 | 「거부가 아무것도 쓰지 않는다」가 가정으로 되돌아간다. 사본이 더러워져도 아무도 보지 않으므로 그 명제는 측정에서 빠진다. |
| 신규 러너/프레임워크 | card t1106 이 명시 금지. 이 가드가 이미 브라우저 발화를 재는 계측기이고, 필드 검증된 계측기를 두고 새 계측기를 쓰면 계측기 자신의 버그가 제품 결함으로 위장한다(§C 원 결정과 같은 논거). |
| 응답 본문 단언으로 대신 | 이미 t1105 단위 테스트가 하는 일이고, 정확히 그것이 **남긴** 간극이 이 개정의 대상이다(B.6.1). |

**사본이 무엇을 서빙하는가 — 두 안 중 (a) 를 채택한다.** `startFireGuardServer` 는 서버를 **하나** 띄우고 매니페스트 전 항목이 그 하나를 쓴다. 따라서 `ProjectRoot` 를 사본으로 바꾸는 것은 신설 항목에만 국한되지 않고 **기존 8개 항목 전부의 서빙 루트를 바꾼다** — `glm_reveal`(시드된 자격증명 상태에 의존), `swap_todo_nav`·`popover_after_swap`(`/todo` 로 이동), `copy_button`(`/specs` 로 이동)은 모두 실제 프로젝트 내용을 읽는다.

| 안 | 내용 | 판정 |
|---|---|---|
| **(a) 사본은 신설 제출 항목만 서빙** — 두 번째 서버 인스턴스, 사본 서빙 표식이 라우팅 키 | **채택** | 폭발 반경 0. 기존 AC-AFG-001~009 가 재는 표면이 한 바이트도 달라지지 않는다. 라우팅 키는 이미 REQ-AFG-014 (1) 이 요구하는 표식 그 자체이므로 새 개념을 도입하지 않는다. |
| (b) 사본이 전 항목 서빙 | 기각 | 기존 8항목이 사본 위에서 **여전히 통과함을 먼저 측정**해야 착지할 수 있고, 그 측정 대상(사본에 무엇을 넣어야 8항목이 도는가)은 오늘 전혀 측정돼 있지 않다. 더 무거운 것은 fidelity 다 — 원판 드라이버가 실저장소 루트를 서빙한 것은 **의도**였고(소스 주석: 「identical to what the real binary serves」), (b) 는 이 가드가 재는 대상을 실트리에서 합성 사본으로 바꾼다. 초록이 **틀린 트리를 재는 초록**이 될 위험을 한 항목을 위해 감수할 이유가 없다. |

**(a) 를 택하면 라우팅 기제도 명세해야 한다 — 오늘의 계약으로는 (a) 가 표현 불가능하기 때문이다.** 탐침은 `--base-url` 하나, `base` 하나, `run_scenario` 한 번, 보고서 하나다. 세 후보 중 첫째를 채택한다.

| 라우팅 기제 | 내용 | 판정 |
|---|---|---|
| **두 번째 base 옵션** (`--sandbox-base-url`) + 표식 기반 per-entry base 결정 | **채택** | 프로세스 하나·보고서 하나·exit 하나가 유지돼 REQ-AFG-005 의 3값 계약이 갈라지지 않는다. 매니페스트 스키마에 새로 드는 것도 없다 — 라우팅 키는 조건 (1) 이 이미 요구하는 표식이다. 비용: CLI 옵션 1개, 드라이버의 서버·포트 1개 추가(수명은 `t.Cleanup`), 그리고 `run_scenario` 가 항목마다 base 를 해석하도록 바뀌는 것(오늘은 `base` 를 인라인으로 붙여 쓴다 — 이 개정의 실제 작업량은 여기다). |
| 매니페스트 항목에 base 필드 | 기각 | 포트는 런타임 할당이라 커밋된 매니페스트가 URL 을 알 수 없다. 자리표시자 치환 규약을 새로 만들어야 하고, 그 규약 자체가 검증되지 않은 표면이 된다. |
| 제출 항목만 겨냥한 **두 번째 탐침 호출** | 기각 | 보고서가 둘, exit 이 둘이 되어 **병합 규칙**(어느 exit 이 이기는가, 실패를 어떻게 합치는가)을 새로 정해야 한다. 「한 번 실행 = 한 판정」이라는 원판 계약을 깨는 값이 한 항목의 대가로 너무 크다. Chrome 탭 준비도 두 배다. |

이 선택은 구현에 위임되지 않는다 — (a) 는 계약이고, 라우팅이 어긋나면(표식 없는 항목이 사본으로 가거나, 표식 있는 항목이 실루트로 가면) AC-AFG-013 이 적색이다.

**도달(paint)의 판별식은 「존재」가 아니다.** DOM 에 노드가 있다는 것은 배너가 보인다는 뜻이 아니다 — 판별식은 가시성(레이아웃 박스 존재 + 조상 사슬 미은닉)과 비어 있지 않은 텍스트이며, 거부 응답 **이후**의 문서에서 측정한다. 구체 술식은 run-phase 소관이되, 「존재만으로 통과」는 금지다(REQ-AFG-015).

### 채택(개정 2, card t1108): post-swap 지표는 **실제 boost 스왑** 뒤, **`htmx:afterSettle` 이벤트** 뒤에만 행사한다

**문제:** REQ-AFG-007 의 대상은 「스왑 뒤 재바인딩」이다. 그런데 매니페스트가 행사하던 것은 전체 이동 직후, 초기화 전의 클릭이었다(§B.7.1·§B.7.2). 이 결정이 뒤집히면 탐침의 5·6단계와 두 매니페스트 항목 정의가 함께 뒤집힌다.

**채택안은 리드가 정한 방향 (a) 다.** REQ-AFG-007 의 의도를 유지하고, 매니페스트를 그 의도에 맞게 옮긴다.

1. **스왑 링크는 served surface 위에서 `hx-boost="true"` 조상을 가진 링크다.** 오늘 그 조건을 만족하는 것은 `/settings` 설정 폼 안의 boost 링크다(§B.7.4 — `/settings?tab=audit`·`/settings?tab=mcp`). 스왑 뒤 페이지도 `/settings` 가 된다. 구체 선택자 문자열은 run-phase 소관이다(§E 「구현 세부」). 계약은 「boost 조상 + 실루트 서버 표면」이다.
2. **스왑 뒤 대기는 `htmx:afterSettle` 이벤트다.** 리스너는 클릭 **전에** 같은 문서에 등록한다. 등록이 클릭보다 늦으면 이벤트를 놓치고, 그 누락이 시간 초과로 나타나 원인을 가린다. 시간 연장(sleep·drain 을 늘리는 것)과 URL 변경 폴링은 둘 다 금지다. 앞의 것은 부하가 바뀌면 다시 깨지고, 뒤의 것이 바로 이번 결함의 형태다. 대기에 상한은 두되, 그 상한은 부재를 **이름 붙은 적색**으로 바꾸는 장치이지 통과 경로가 아니다.
3. **스왑이 스왑이었는지를 탐침이 스스로 확인한다(REQ-AFG-016 신설).**

**3번을 채택하는 이유.** 이번 결함은 SPEC 의 첫 판(t1060)부터 t1106 개정까지 한 번도 드러나지 않았다. 5단계는 URL 이 `/todo` 가 됐는지만 봤고, 전체 이동도 그 조건을 만족하기 때문이다. 검사가 전제를 재지 않으면 `hx-boost` 배치가 바뀔 때마다 같은 오류가 조용히 되돌아온다(판정서 §7). 링크에서 boost 조상이 빠지거나 폼에서 `hx-boost` 가 제거될 때가 그렇다. 자기확인은 그 변화를 셀렉터 스테일과 같은 급의 적색으로 바꾼다(REQ-AFG-004 와 같은 논리 — verification-completeness §1.3 continued firing). 비용은 탐침 안의 판정 몇 개와, 셀렉터 미달과 나란한 실패 사유 하나다. 계측기 전제가 틀렸다는 것을 CI 가 매번 알려 주는 값으로 싸다.

**자기확인이 exit 2 가 아니라 exit 1 인 이유.** exit 2 는 CDP·서버 도달 불가 같은 계측 기반 결함이다(REQ-AFG-005). 스왑 전제의 붕괴는 제품 표면의 변화, 즉 boost 배치 변경으로 생기며, 그 결과 매니페스트가 더는 자기 대상을 재지 못하게 된다. 이것은 셀렉터 스테일(REQ-AFG-004, exit 1)과 같은 부류다.

**기각한 대안**

| 대안 | 기각 사유 |
|---|---|
| (b) 지금 링크를 그대로 두고 `DOMContentLoaded` 뒤의 초기화 완료를 기다린다 | 판정서 §3 의 대조군이 이 방향의 안정성을 보였다. 그러나 이 방향에서는 검사 대상이 「전체 이동 뒤 초기화」로 바뀌고, 이는 1·3·7단계의 전체 로드 검사와 같은 성격이다. **REQ-AFG-007 이 보려던 스왑 뒤 재바인딩(`htmx:afterSettle` 리스너)을 어떤 단계도 검사하지 않게 된다.** 리드가 (a) 를 택했다 |
| boost 된 `/todo` 링크를 제품에 새로 만든다 | 이 SPEC 은 `internal/web/` 제품 표면을 바꾸지 않는다(REQ-AFG-011 의 정신, §E). 가드를 통과시키려고 제품을 바꾸는 것은 순서가 뒤집힌 것이다 |
| 대기를 고정 시간으로 늘린다 | 부하가 커지면 다시 깨진다. 스로틀 12배에서 30/30 실패한 형태의 연장선이다 |
| 자기확인 없이 링크만 바꾼다 | 이번 결함의 재발 경로가 그대로 남는다(위 「3번을 채택하는 이유」) |

**스왑 항목의 id 를 바꾼다.** `swap_todo_nav` → `swap_boosted_tab`. 새 정의에서 이 항목은 `/todo` 로도, nav 링크로도 가지 않는다. 옛 이름을 두면 이름이 스스로에 대해 거짓을 말하게 되고, 이번 결함이 바로 이름과 실제가 어긋난 채 몇 개 판을 버틴 사례다. `popover_after_swap` 은 새 정의에서도 정확하므로 유지한다. 보고서 키(`p5_*`·`p6_*`)는 드라이버의 JSON 태그(`appjs_fire_guard_test.go:70`·`:73`)가 고정하는 계약이라 **바꾸지 않는다**. 이름을 바꾸는 파급 범위는 plan.md §E M8.6 에 전부 열거했다.

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

### REQ-AFG-007 (ubiquitous 머리 문장 + event-driven 절 — 개정 2, card t1108 에서 문언 개정)

> **형식 표기(plan-audit iter-1 D7).** 이 조항은 두 GEARS 형식을 담는다. 머리 문장은 주어가 매니페스트인 **ubiquitous** 문장이고, 번호 붙은 세 절은 주어가 탐침인 **event-driven** 절이다. 두 번호로 나누지 않은 이유는 Tier M 의 REQ 상한이 16 이고 이 SPEC 이 이미 16건이기 때문이다. 조항을 나누면 상한을 넘는다. 대신 각 부분의 형식을 여기 밝힌다.

**(ubiquitous)** 매니페스트는 hx-boost 바디 스왑 **뒤에** 최소 1개 지표를 행사해야 한다(shall).

**(event-driven)** **When** 탐침이 스왑 뒤 지표를 행사할 때, 탐침은 다음을 지켜야 한다(shall):

1. **스왑은 실제 hx-boost 스왑이어야 한다.** 스왑을 일으키는 클릭 대상은 served surface 에서 `hx-boost="true"` 조상을 가진 링크여야 한다(shall). 그 표면은 실루트 서버다 — 스왑 항목과 스왑 뒤 항목은 사본 서빙 표식을 갖지 않는다. 그 스왑이 스왑이었음은 REQ-AFG-016 의 자기확인이 판정한다(shall). URL 이 바뀌었다는 사실만으로 스왑을 추정해서는 안 된다(shall not).
2. **스왑과 지표 행사 사이의 대기는 `htmx:afterSettle` 이벤트여야 한다.** 스왑된 문서에서 그 이벤트가 관측되는 것을 기다려야 하고(shall), 리스너는 클릭 **전에** 같은 문서에 등록돼 있어야 한다(shall). 대기는 시간 연장(고정 sleep·drain 의 증량)이어서는 안 되고(shall not), URL 변경 폴링(`location.pathname` 등)이어서도 안 된다(shall not).
3. **대기에 둔 상한은 부재를 적색으로 바꾸는 장치다.** 상한 안에 `htmx:afterSettle` 이 관측되지 않으면 탐침은 **exit 1** 로 실패해야 한다(shall). 이는 REQ-AFG-005 의 exit 1 범주(지표 붕괴)에 드는 사유이며 3값 계약에 값을 더하지 않는다. 보고서는 **스왑 뒤 항목(`popover_after_swap`)** 을 사유 「afterSettle 대기 만료」와 함께 지목해야 한다(shall). 같은 실행에서 REQ-AFG-016 (c) 도 거짓이면 스왑 항목도 함께 지목하며, 두 지목은 모두 보고서에 남는다. 탐침은 그 만료를 통과로 읽어서는 안 된다(shall not). 구체적으로, 새 대기는 오늘 탐침의 `poll` 이 쓰는 「상한이 지나면 그 시점의 현재값을 돌려주고 진행한다」는 의미론(`internal/web/testdata/appjs_fire_probe.py:434-443`, 마지막 줄 `return await ev(cdp, expr)`)을 **재사용해서는 안 된다**(shall not). 만료는 값이 아니라 실패 사건이다.

역사적 결함 가족은 스왑 뒤의 무관한 등록을 죽였다. 스왑 뒤 발화는 이 가드가 정적 가드와 구별되는 축 중 하나다. `app.js` 는 스왑 뒤 재바인딩을 `htmx:afterSettle` 리스너로 수행하므로(§B.7.2), 그 이벤트를 기다리지 않는 검사는 재바인딩이 아니라 경주를 잰다.

> **개정 전 문언 (0.1.0 ~ 0.2.0, 이력 보존).** 「매니페스트는 hx-boost 바디 스왑 **뒤에** 최소 1개 지표를 행사해야 한다(shall). 역사적 결함 가족이 스왑 뒤의 무관한 등록을 죽였고(B.2 에서 스왑 Phase 가 같은 ReferenceError 를 찍은 것이 그 재확인이다), 스왑 뒤 발화야말로 이 가드가 정적 가드와 구별되는 축 중 하나다.」
>
> **사후 정정 (2026-09-23, card t1108).** 괄호 안 근거는 틀렸다. B.2 의 「스왑 Phase」는 전체 이동이었고, 거기 찍힌 ReferenceError 는 새 문서가 `app.js` 를 다시 실행하며 난 로드 시점 예외다(§B.2 사후 정정, §B.7.1). 조문의 **의도**(스왑 뒤 재바인딩을 잰다)는 유지한다. 그 의도가 어느 판에서도 실제로 측정되지 않았다는 사실이 이 개정의 계기다.

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

### REQ-AFG-014 (capability gate — card t1106, REQ-AFG-012 를 약화하지 않는 좁은 확장)

**Where** 매니페스트 항목의 효과 종별이 `validation-reject` 일 때, 그리고 **오직 그때만**, 탐침은 영속화 계열 제어(폼 제출)를 행사해도 된다(may) — 단 다음 **세 조건이 함께** 성립해야 한다(shall):

1. **사본 서빙 — 그리고 사본은 그 항목만 서빙한다(안 (a)). 라우팅 기제는 이 조항의 일부다.** 그 항목은 **별도의 두 번째 서버 인스턴스**가 일회용 프로젝트 사본을 `ProjectRoot` 로 삼아 서빙한다(shall). 실저장소 루트(`findRepoRoot`) 위에서 이 종별을 행사해서는 안 된다(shall not).

   오늘의 탐침 계약으로는 이것이 **표현 불가능하다** — 탐침은 `--base-url` 을 하나만 받고(`appjs_fire_probe.py` 의 `main()` optparse), `main_async` 가 `base` 하나를 결정해 `run_scenario(cdp, base)` 를 한 번 호출하며, 시나리오는 그 하나의 `base` 에 경로를 붙여 모든 항목을 운전한다. 드라이버 쪽도 `startFireGuardServer` 가 서버 하나를 띄워 URL 하나를 돌려준다. 따라서 조항이 만족 가능해지려면 계약이 바뀌어야 하고, **어떻게 바뀌는지가 이 조항에 속한다**:

   - 탐침은 **두 번째 base 옵션**(`--sandbox-base-url <URL>`)을 받아야 한다(shall). 시나리오는 항목마다 base 를 **표식으로 결정한다**(shall) — 사본 서빙 표식이 있는 항목은 sandbox base 로, 없는 항목은 primary base 로. 표식 없는 항목을 sandbox base 로 운전해서는 안 되고(shall not), 표식 있는 항목을 primary base 로 운전해서도 안 된다(shall not).
   - **탐침 실행은 여전히 한 번이고 보고서도 하나다**(shall) — REQ-AFG-005 의 3값 exit 계약은 갈라지지 않는다. 두 base 의 관측은 같은 보고서·같은 판정에 합류한다.
   - 매니페스트에 **URL 을 적지 않는다**(shall not) — 포트는 런타임 할당이고 커밋된 매니페스트가 알 수 없다. 매니페스트가 운반하는 것은 표식이고, URL 은 호출자가 준다.
   - **사본 base 가 주어지지 않은 경우의 계약(card t1106 iter-2 — 이 경우는 미정의로 남길 수 없다).** 커밋된 매니페스트에 사본 서빙 표식을 가진 항목이 있는데 `--sandbox-base-url` 이 주어지지 않으면, 탐침은 **exit 2(기계결함)** 로 실패하고 그 표식 항목을 **이름으로** 보고해야 한다(shall). 그 항목을 **조용히 건너뛰어서는 안 되고**(shall not — 침묵 스킵 금지), primary base 로 대체 운전해서도 안 된다(shall not). 침묵 스킵은 이 가드의 논지 자체(빈 스윕의 초록은 통과가 아니라 미측정이다)를 가드 안에서 뒤집는 형태이므로 이름을 붙여 금지한다. exit **1** 이 아니라 **2** 인 이유는 이것이 제품 결함이 아니라 호출자 배선 결함이기 때문이다 — REQ-AFG-005 의 기계결함 값이 그 구분을 운반한다.
   - **한 실행이 매니페스트의 일부만 운전하려면, 그 축소가 적극적 선언이어야 한다(card t1106 iter-2).** 호출자는 **실루트 계열만 운전한다**고 명시 선언할 수 있고(may), 그 선언이 있을 때에 **한해** 탐침은 표식 항목을 운전하지 않으며 위 불릿의 exit 2 에도 걸리지 않는다. 선언의 구체 플래그 이름은 run-phase 소관이되 세 가지는 계약이다(shall): (i) 선언은 **적극적**이어야 한다 — 사본 base 의 **부재**가 축소를 함의해서는 안 된다(shall not; 선언 없는 부재는 위 불릿대로 exit 2 다), (ii) 보고서는 **운전한 항목 수**와 **선언으로 제외된 항목의 이름**을 함께 담아야 한다 — 제외는 보고되는 사건이지 침묵이 아니다, (iii) 제외 집합은 **정확히 표식 계열**이어야 하고 운전 집합이 비면 exit 1 이다 — 「아무것도 운전하지 않는 선언」으로 초록을 만들 수 없다. 이 선언이 있는 이유는 AC-AFG-001(실루트 계열 전체 사이클)과 AC-AFG-010(제출 사이클)이 **서로 다른 실행**이기 때문이다 — 두 계열을 한 판정에 섞지 않는 것이 이 개정의 결정이다.

   기존 항목의 서빙 루트는 **한 바이트도 바뀌지 않는다**(shall not) — `startFireGuardServer` 의 `ProjectRoot: findRepoRoot(t)` 는 불변이고, 사본 서버는 그 옆에 추가된다.
2. **무쓰기 단언.** 탐침이 제출 전후로 사본 루트의 **바이트 불변**을 스스로 단언하고, 한 바이트라도 달라지면 exit 1 로 실패하며 달라진 경로를 이름으로 보고한다(shall). 「거부 경로는 아무것도 쓰지 않는다」는 전제가 아니라 이 조항이 계속 측정하는 명제다. 비교에서 제외하는 경로가 있다면 각각 사유와 함께 열거돼야 한다(shall) — 사유 없는 제외는 단언을 조용히 공허하게 만든다. **그리고 사유의 존재는 어느 경로를 뺄 수 있는지를 제한하지 않으므로, 범위 자체를 따로 묶는다(card t1106 iter-2):** 제외는 **행사되는 요청의 쓰기 이음매가 도달할 수 있는 경로를 덮어서는 안 된다**(shall not). `/save` 경로의 경우 그 이음매가 닿는 곳은 사본 루트 안의 `.moai/config/sections/**` 다(`internal/web/handlers.go` 의 `handleSave` 가 `SyncToProjectConfig` 와 `writeProjectConfig` 로 쓰는 자리 — spec.md §B.6.3). 그 하위 트리는 사유가 붙어 있어도 제외 대상이 아니다 — 제외하면 단언이 **구성상 초록**이 되어, 회귀를 잡으라고 둔 조항이 회귀를 못 보는 자리에서만 돌게 된다. 제외가 허용되는 곳은 행사 경로가 닿지 못하는 자리뿐이다.
3. **수명은 테스트 프레임워크가 보증한다.** 사본의 수명은 `t.TempDir()` 에 묶여야 하고(shall), 따라서 panic·조기 실패를 포함한 **모든 종료 경로**에서 제거된다. `defer` 나 함수 말미의 제거문에 수명을 맡겨서는 안 된다(shall not) — 뒤에 붙은 정리는 프로세스가 도달하지 못할 수 있는 줄이며, 이 저장소는 정리가 프레임워크 등록이어야 한다는 규율을 따로 갖고 있다.

세 조건은 **분리 불가능하다**(shall) — 하나라도 빠진 항목이 이 종별을 주장할 수 없어야 하고, 그 불가분성은 탐침의 매니페스트 자기검증(1·2 의 표식)과 드라이버 측 기계 판정(1 의 라우팅, 3 의 수명)이 함께 판정해야 한다(shall).

이 조항은 REQ-AFG-012 를 대체하지 않는다. REQ-AFG-012 가 이미 이름한 두 경로 중 둘째(일회용 사본 서빙)를 사용하는 구체 조건이며, 제출 계열의 **기본값은 여전히 제외**다.

### REQ-AFG-015 (event-driven — card t1106)

**When** `validation-reject` 항목이 검증에 실패하는 값으로 폼을 제출할 때, 탐침은 거부 응답 **이후의 문서에서** 거부 배너가 **칠해졌음**을 관측해야 한다(shall) — 즉 (a) 배너 노드가 레이아웃 박스를 가지고 조상 사슬 어디서도 은닉되지 않았으며, (b) 그 텍스트가 비어 있지 않아야 한다. 노드의 **존재만으로 통과해서는 안 된다**(shall not) — 응답 본문 안에 배너가 있다는 명제는 단위 테스트 층이 이미 덮고 있고, 이 가드가 더하는 축은 그것이 화면에 도달하는가다.

같은 관측 창에서 load/swap ReferenceError 0건 계약(REQ-AFG-005·006)은 그대로 적용된다(shall).

### REQ-AFG-016 (event-driven — 개정 2, card t1108: 스왑 자기확인)

**When** 탐침이 스왑 항목을 행사할 때, 탐침은 그 클릭에 대해 다음 네 다리를 스스로 확인하고 결과를 보고서에 다리별로 담아야 한다(shall):

- (a) **boost 조상** — 클릭 시점에 클릭 대상이 `hx-boost="true"` 조상을 가졌다.
- (b) **같은 문서** — 클릭부터 스왑 뒤 지표 행사까지 새 메인 프레임 문서가 생기지 않았다. 판정은 클릭 전에 문서에 심은 표지가 행사 시점에도 남아 있는가로 한다.
- (c) **스왑 이벤트** — 그 같은 문서에서 `htmx:afterSwap` 과 `htmx:afterSettle` 이 모두 관측됐다.
- (d) **스왑이 만든 트리거** — 스왑 뒤에 행사한 트리거가 스왑이 삽입한 노드다. 판정은 클릭 전에 옛 트리거 노드에 표지를 달고, 스왑 뒤 같은 선택자가 가리키는 노드에 그 표지가 없는가로 한다. 옛 노드가 핸들러를 가진 채 살아남았다면, 발화는 재바인딩이 아니라 옛 결합을 잰 것이다.

**When** 네 다리 중 하나라도 거짓일 때, 탐침은 **exit 1** 로 실패해야 한다(shall). 보고서는 스왑 항목을 이름으로 지목하고, 거짓이 된 다리를 사유로 담아야 한다(shall) — 예: `swap premise not met: full navigation`. 그 실행에서 스왑 뒤 지표를 발화로 판정해서는 안 된다(shall not). 이 실패는 REQ-AFG-005 의 exit 1 범주(지표 붕괴 또는 셀렉터 미달)에 드는 새 **사유**일 뿐이며, 3값 exit 계약에 값을 더하지 않는다.

이 조항은 스왑의 **전제**를 재고, REQ-AFG-007 은 그 전제 위의 **발화**를 잰다. 둘 중 하나만 있으면 판단이 서지 않는다. 자기확인이 없으면 전체 이동도 스왑으로 통과하고, 발화 판정이 없으면 스왑이 일어난 것만 확인할 뿐 재바인딩은 재지 않는다.

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

### Out of Scope — 개정분(card t1106)이 넓히지 않는 것

- REQ-AFG-012 의 완화를 하지 않는다. 허용목록이 얻는 것은 두 조건이 함께 성립할 때의 `validation-reject` 한 종뿐이며, 저장·제출 계열의 **기본값은 제외**로 남는다.
- 신규 러너·신규 브라우저 프레임워크를 만들지 않는다 — 기존 탐침과 드라이버의 확장만이다.
- 거부 경로의 서버 측 계약(상태 코드 선택, 필드 에러 병합, atomic reject)을 이 SPEC 이 다시 정하지 않는다 — 그것은 card t1105 가 이미 고쳤고 단위 테스트가 소유한다.
- 기존 항목의 서빙 루트를 바꾸지 않는다 — 사본은 신설 제출 항목 하나만 서빙한다(안 (a)). 전 항목을 사본 위로 옮기는 안 (b) 는 §C 개정 결정에서 기각됐다.
- 성공 저장 경로(배너가 `ok` 로 칠해지는 경우)를 행사하지 않는다 — 성공 저장은 실제 영속화이고, 사본 위에서라도 이 가드의 목적 밖이다.
- 배너의 시각적 스타일·대비·접근성을 재지 않는다(§E 일반 브라우저 테스트 인프라 제외와 같은 이유).

### Out of Scope — 개정 2(card t1108)가 넓히지 않는 것

- **판정 규칙의 관측 층 불일치는 기록만 하고 고치지 않는다.** `internal/web/testdata/appjs_fire_probe.py:789` 는 `popover_after_swap` 의 「selector matched nothing」을 **패널** `.hidden` 값이 null 인지로 판정한다(`rep.get("p6_panel_hidden_before") is not None`). 그런데 보고서의 `selector` 칸에는 **트리거** 선택자 `[data-pop="profile"]` 이 찍힌다(`:798`·`:802` 가 매니페스트 항목의 `selector` 를 그대로 싣고, 그 값은 `:183` 의 트리거 선택자다). 트리거는 있고 패널이 없는 경우와 그 반대를 보고서만으로는 가를 수 없다. 이 개정은 그 규칙의 변경을 요구하지 않는다. 판정서가 인용한 좌표 `:474` 는 t1106 병합 전 트리의 것이며, 같은 규칙이 이 트리에서는 `:789` 에 있다.
- **`app.js`·제품 템플릿을 바꾸지 않는다.** boost 된 `/todo` 링크를 만들어 옛 매니페스트를 살리는 길은 §C 개정 결정 2 에서 기각했다.
- **CI 는 `.github/workflows/ci.yml:672` 그린 단계 한 줄만 바꾼다.** 바꾸는 것은 `-run` 정규식이다. `-timeout 10m` 은 병합 트리 실측이 그 상한을 넘을 때에만, 측정 소요 시간 + 여유로 올린다(리드 결정 B3, AC-AFG-016). 올린 값이 job 상한 `timeout-minutes: 20`(`ci.yml:608`) 안에 들어가지 않으면 job 상한은 바꾸지 않고 run-phase 가 멈추며 blocker 로 보고한다(plan-audit iter-1 D4). `--primary-entries-only` 3곳(734/745/767)은 유지한다 — 빼면 t1106 의 F1(표식 항목이 사본 base 없이 운전돼 exit 2)이 되살아난다. 그 밖의 단계와 그 상한은 건드리지 않는다.
- **`INVENTORY_TOTAL` 과 lint 커버리지 산식을 바꾸지 않는다**(§B.7.6).
- **보고서 키(`p5_*`·`p6_*`)의 이름을 바꾸지 않는다.** 드라이버의 JSON 태그가 고정하는 계약이다. 스왑 창 이름 `p5_swap_referenceerrors` 의 **의미**가 바뀌는 문제는 §F 가 아니라 plan.md §F 잔여 위험에 기록했다.
- **스왑 대기와 자기확인 (c) 는 이벤트가 「이 클릭의 요청」에서 나왔는지는 묻지 않는다 — 알려진 한계로 기록만 한다(plan-audit iter-1 D5).** `app.js` 에는 같은 문서 안에서 스왑을 일으키는 실시간 갱신 경로가 있다(`internal/web/assets/app.js:686-695` — `htmx.ajax("GET", window.location.href, {target: ".body", select: ".body", swap: "outerHTML"})`). 설정 화면에 실시간 영역이 생기면 무관한 settle 이 대기를 풀고 (c) 를 참으로 만들 수 있다. 오늘 `/settings` 템플릿에는 실시간 영역이 없다(`grep -c 'data-live=' internal/web/root.templ` → `0`). 그래서 이 개정은 요구사항을 바꾸지 않고, plan.md §F 잔여 위험과 acceptance.md §D 에 기록한다.
- **실바이너리 레드 단계의 돌연변이 대상을 바꾸지 않는다.** 스왑이 실제 스왑이 되면서 돌연변이 아래에서 어느 지표가 무너지는지는 달라질 수 있다. 그 결과는 run-phase 가 AC-AFG-002 재측정으로 **관측**하며, 미리 가정하지 않는다.

### Out of Scope — 구현 세부

- 셀렉터 문자열·매니페스트의 최종 항목 목록·Chrome 고정 버전 숫자·정확한 워크플로 YAML 은 run-phase 소관이다. 이 SPEC 은 계약과 판별식만 정한다.

---

## §F 이 가드가 덮지 못하는 것 (정직한 한계 선언)

이 가드는 **완전한 발화 보증이 아니다.** 일곱 방향으로 놓친다(1~4 는 원판, 5~7 은 card t1106 개정분).

1. **매니페스트에 없는 상호작용.** 13줄 인벤토리(B.3) 각 그룹이 항목이거나 명시적 제외일 뿐, 제외된 것은 재지 않는다. 제외 사유가 취약하면 가드도 취약하다.
2. **탐침이 행사하지 않는 경로.** 클릭·스왑·재클릭의 순서를 바꾸면 나타날 수 있는 상태는 재지 않는다 — 탐침은 정해진 시나리오 하나를 운반한다.
3. **Chrome 한 종류.** WebKit·Firefox 에서만 나타나는 발화 차이는 이 가드의 밖이다.
4. **정적 축의 결함을 런타임 그림자로만 본다.** IIFE 경계 결함은 B.2 처럼 런타임에서도 관측되지만, 그 축의 정밀 판정(식별자·줄 번호 지목)은 여전히 정적 가드의 몫이다. 이 가드의 red 가 곧 원인 보고는 아니다.

5. **칠해짐의 판별식이 덮는 범위.** REQ-AFG-015 는 배너가 보이는 자리에 비지 않은 텍스트로 도달했는가까지만 잰다 — 문구의 정확성, 대비, 스크롤 위치, 포커스 이동은 이 가드 밖이다.
6. **무쓰기 단언의 범위는 사본 루트다.** 사본 밖(프로필 스토어·임시 디렉터리·프로세스 환경)에 남는 흔적은 이 단언이 보지 않는다. 제외 경로를 두면 그 사유의 품질이 곧 단언의 품질이다(REQ-AFG-014).

7. **두 서버 서면의 등가는 주장하지 않는다.** 사본 서버와 실루트 서버가 같은 표면을 서빙한다고 이 가드는 주장하지 않는다 — 사본 서버는 거부 경로 한 갈래만 운반하고, 나머지 전 항목은 실루트 서버가 그대로 운반한다(REQ-AFG-014 (1)). 그것이 폭발 반경을 0 으로 두는 방식이다.

이 가드가 덮는 것은 **사용자가 버튼을 눌렀을 때 반응하는가**의 최소 표면이고, 그것은 어떤 정적 규칙도 대신 봐 주지 않는 축이다. 이 한계는 REQ-AFG-013 에 따라 소스 주석에도 그대로 들어간다.

---

## §G 인수조건

Tier M 이므로 AC 정본은 별도 `acceptance.md` 다. 요약: (1) 게이트 켠 드라이버가 실트리에서 전 지표 발화로 통과하되 매니페스트 0행·ReferenceError 를 함께 거부, (2) 돌연변이 재도입 시 탐침 exit 1 — 양방향 모두 관측된 출력으로, (3) 게이트 없는 실행은 사유를 이름으로 대는 skip, (4) 기존 job·assets·go.mod 무변경. 개정분(card t1106): (5) 검증 실패 제출 뒤 거부 배너가 **칠해졌음**을 관측, (6) 그 제출이 일회용 사본 위에서만 일어나고 사본이 바이트 불변임을 탐침이 단언, (7) 매니페스트의 **두 조건 표식**(사본 전용 서빙 · 무쓰기 단언) 없이 `validation-reject` 를 주장하는 항목은 매니페스트 자기검증이 거부하고, 셋째 조건(프레임워크 등록 수명)은 매니페스트가 운반할 수 없는 Go 쪽 수명 속성이므로 드라이버 측 기계 판정(AC-AFG-011 (d))이 진다, (8) 기존 항목의 서빙 루트 불변 — 사본은 신설 항목만 서빙한다. 개정 2(card t1108): (9) CPU 스로틀 12배의 in-process 표면에서 스왑 뒤 지표가 실제 boost 스왑과 `htmx:afterSettle` 이벤트 대기 뒤에 발화하고, 두 돌연변이 — 대기를 `location.pathname == "/settings"` 폴링으로 되돌린 판(스왑 대상 경로라 settle 전에 이미 참이다)과 afterSettle 대기를 제거한 판 — 는 같은 조건에서 적색이며, 만료된 대기는 통과가 아니라 적색이다, (10) 자기확인 네 다리가 **다리마다** 제 고정물에서 거짓이 되고, 그때 그 다리를 지목하며 적색이다((a)·(b)·(c) 는 전체 이동 돌연변이에서 동시에, 각 다리 단독은 합성 보고서 고정물로, (c) 단독은 리스너를 클릭 뒤에 붙인 사본으로, (d) 단독은 표지를 스왑 뒤 노드에 다는 사본으로), (11) 병합 트리에서 넓힌 CI 선택자 `'AppJs.*Fire'` 가 실루트 계열과 제출 계열을 둘 다 skip 없이 돌린다.

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
- card t1105 — `handleSave` 검증 거부 경로의 400 → 200 수리와, 그 판정서가 Gap 으로 남긴 「브라우저에서 칠해지는가」 구간.
- card t1106 — 이 개정(REQ-AFG-014·015)의 배차 카드.
- `internal/web/handlers.go` — `handleSave` 의 검증 거부 블록(t1105 주석이 boost·htmx 2.0.4 responseHandling 근거를 담고 있다).
- `internal/web/transport400_swap_contract_test.go` / `transport400_characterization_test.go` / `transport400_htmx_contract_test.go` — 응답 **본문** 층의 기존 검증면. 이 개정은 그 위층(도달)만 더한다.
- card t1108 — 개정 2(REQ-AFG-007 문언 개정, REQ-AFG-016)의 배차 카드. 원인 확정 판정서 `.moai/reports/t1108/verdict.md`, 측정 로그 `.moai/reports/t1108/logs/`(드라이버 표면 측정 `gate-inprocess.log`).
- `internal/web/root.templ:56` / `internal/web/shell.templ:307` — 콘솔의 `hx-boost` 두 자리(설정 폼, 프로필 팝오버).
- `internal/web/handlers.go:263` — `?tab=` 을 읽는 GET 경로(조회 전용).
