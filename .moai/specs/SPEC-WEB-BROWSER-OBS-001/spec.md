---
id: SPEC-WEB-BROWSER-OBS-001
title: "moai web console — REQ-A real-browser (CDP) observation: save__msg--error swap discriminator and premise re-measurement"
version: "0.1.1"
status: in-progress
created: 2026-09-22
updated: 2026-09-22
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: internal/web
lifecycle: spec-anchored
tags: "web-console, cdp, browser-observation, htmx, hx-boost, discriminator, measurement"
era: V3R6
tier: M
related_specs: [SPEC-WEB-CONSOLE-017, SPEC-WEB-TRANSPORT-001]
---

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-09-22 | manager-spec | 최초 draft. 근거는 카드 t1081 — SPEC-WEB-CONSOLE-017 (t1051) 이 명시 허용으로 남겨 둔 갭(§6-1 브라우저 실측, §6-2 htmx 버전별 비-2xx 처리)을 **실제 브라우저**로 닫는 측정-판정 SPEC. t1051 sync-audit F3·Gaps(1)(2) 가 원천이다. 판별식은 페이지 본문 문자열 탐침이 아니라 **살아 있는 브라우저**다 — 문자열 탐침 판별식은 t1051 카드 본문이 폐기 판정했다. |
| 0.1.1 | 2026-09-22 | manager-spec | plan-audit iter-1 FAIL(0.79) 수리. **D1 [BLOCKING]** — REQ-BO-002 신호 (c) 「`location.href` 불변」은 기대 결과(성공 스왑)에 의해 위반된다: hx-boost + 임베드 htmx 2.0.4 `historyEnabled:true` + `hx-push-url` override 0건 → 성공 스왑조차 pushState 로 href 를 action URL 로 바꾼다(판정 트리 측정, `plan-audit.md` §4 D1). 내비게이션 증거를 메인 프레임 CDP 마커로 고정하고 href 는 기록 전용으로 강등. **D2** — swap-failed-no-nav 제3 결과값 신설(원인 라벨 분리). **D3** — plan pre-flight 에 Chrome 기동 절차·websockets 의존성 프루브 추가. **D4** — AC-BO-001 에 슬롯 문구 일치 비교(폴백 문구 거부) 추가. **D5** — plan §A.1 의 t1080 상태 서술을 종결 순서-무관으로 수정. |

---

## §1 Context & Motivation

SPEC-WEB-CONSOLE-017 (카드 t1051) 은 저장 실패 시 실패 사유가 인라인 슬롯(`save__msg save__msg--error`, `role="alert"`, `internal/web/shell.templ`, 트리 cd99336bf 기준 :274)에 htmx 스왑 가능한 응답으로 도달한다는 REQ-A 를 정의하고, 수송 결함(500 풀페이지 + `hx-boost` 가 비-2xx 본문을 스왑하지 않음)을 2xx 재렌더(`renderErrorPage`, `internal/web/handlers.go`)로 닫는 구현을 감사 종료했다. 그러나 그 SPEC 은 두 갭을 **명시 허용**으로 남겼다:

- **갭 1 (t1051 §6-1)** — 콘솔 에러 수(실패 2 / 성공 0)와 실패 본문 크기(약 110KB)는 **카드 분석의 전제**로서 표기된 것이지 그 레인의 측정이 아니다. REQ-A 인수는 렌더/DOM 계층 기계 검증(AC-WC17-001)으로 세웠고 브라우저 재측정은 선택이었다.
- **갭 2 (t1051 §6-2)** — htmx 버전별 비-2xx 응답 처리 세부는 미측정이다. 어떤 기구로 REQ-A 가 전달되는지는 관측 가능한 결과만 요구했으므로, 브라우저가 실제로 스왑하는지는 **아무도 보지 못했다**.

본 SPEC 은 이 갭을 **코드를 더 읽는 것으로** 닫지 않는다. 판별식은 **실제 브라우저가 페이지를 실행하는 것**이다 — 페이지 본문 문자열 탐침(grep/파싱)은 판별식이 아니며, t1051 카드 본문이 명시적으로 폐기한 판별식이다. 작동하는 본보기는 t1041 의 `browser-probe.py`(`.moai/reports/t1041/browser-probe.py`)다 — python3 + websockets CDP 클라이언트로 원격 디버깅 포트를 연 Chrome 에 탭을 열고, 콘솔·DOM 이벤트를 캡처하고, 탭을 닫는 기구. 본 SPEC 은 그 기구의 모양을 재사용하고 표적을 바꾼다.

### §1.1 카드 분할 경계 (t1080 vs t1081) — 형제 SPEC 의 구속력 있는 분할표 인용

형제 SPEC-WEB-TRANSPORT-001 (t1080, 미병합) 의 §1.2 「카드 분할 결정」 이 본 카드의 축 소유를 확정한다:

| | t1080 (SPEC-WEB-TRANSPORT-001) | **t1081 (본 SPEC)** |
|---|---|---|
| 측정 대상 | 서버 측 400 본문 특성화 + 임베드 htmx 자산의 비-2xx 스왑 계약 (기계적, 브라우저 불요) | **실제 브라우저에서의 2xx 실패 재렌더 전달 (REQ-A 갭)** + 관측 가능 시 4xx 처리 **부수 기록** |
| 산출물 | 판별식 판정값(D-400) + 결함 판정 기록 (계약 수준 신뢰도) | **브라우저 관측 기록 (browser-observed 신뢰도)** |
| 판정 지위 | 독립 판정 — t1051 판정 비인용 | t1080 판정의 런타임 확인 불요(선택), 자기 REQ-A 갭은 독립 |

**신뢰도 어휘** — 본 SPEC 의 산출물 등급은 `browser-observed` 다. t1080 의 상한은 `contract-level`이며, **계약 수준 주장을 browser-observed 로 표기하는 것은 관측되지 않은 검증 주장이다**(`.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 1). 역방향도 같다 — 본 SPEC 이 계약 수준 추론만으로 판정을 쓰면 그것도 브라우저 관측이 아니다.

### §1.2 측정 귀속

본 문서의 심볼 좌표(`renderErrorPage`, `save__msg save__msg--error` 슬롯, `hx-boost` 폼, `settingsSaveState` 데이터 경로)는 트리 **cd99336bf**(배차된 로컬 develop, t1051 의 REQ-A 구현이 착지돼 있는 트리)에서 본 레인이 이번 실행으로 재확인한 값이다 — 슬롯 선택자와 `hx-boost` 속성, `renderErrorPage` 정의 존재를 grep 으로 관측했다. **본 레인이 측정하지 않은 값** — 실패 제출 시 콘솔 에러 수, 성공 제출 시 콘솔 에러 수, 실패 응답 본문 크기 — 는 t1051 카드 분석의 전제이며, 본 SPEC 의 run 단계가 **측정으로 승격**하는 대상이다(REQ-BO-003). 브라우저 실측 전까지 그 값들은 어디에도 측정값으로 인용되지 않는다.

---

## §2 Requirements (GEARS)

REQ 식별자: `REQ-BO-NNN`. 요구의 주체는 대부분 **측정 하위시스템**(run 단계의 브라우저 관측 절차 전반)이다 — 본 SPEC 은 운용 코드를 만지지 않는다.

### REQ-BO-001 — 판별식: 살아 있는 브라우저만이 판별식이다

**When** run 단계가 REQ-A 갭(실패 제출 후 `save__msg save__msg--error` 슬롯 도달 여부)을 판정할 때, the browser-observation subsystem shall judge the gap only from live-DOM state changes observed in a real browser session (CDP-connected Chrome against a `moai web` server built at cd99336bf) and shall not use page-source string probing (grep·파싱·정적 HTML 판독) as the discriminator. 성공 판정의 최소 조건: 실패 사유 문구가 `role="alert"` 슬롯 노드에 **보이는(visible)** 텍스트로 도달한다. 렌더됐으나 화면에 표시되지 않는 노드의 존재만으로는 REQ-A 전달의 관측이 아니다.

### REQ-BO-002 — 스왑 vs 풀페이지 내비게이션 판별의 조작화

The browser-observation subsystem shall classify the failed-submit delivery into one of the following three outcomes and record the classification:

- **htmx-swap** — 제출 교환 동안 (a) 메인 프레임 **내비게이션 마커**(`Page.frameNavigated` 의 메인 프레임 도달, 신규 로드의 `Page.loadEventFired`)가 관측되지 **않고**, (b) 부스트된 컨테이너(`hx-boost="true"` 폼이 속한 본문 영역) 안의 DOM 변이가 슬롯 노드의 신설 또는 내용 갱신을 전달한다.
- **full-page-navigation** — (a) 위배: 메인 프레임 내비게이션 마커가 관측된다. 새 문서 렌더의 검출은 (a) 하나로 충분하다 — htmx 의 `history.pushState` 는 CDP 메인 프레임 내비게이션 이벤트를 내지 않는다.
- **swap-failed-no-nav** — (a)는 유지되나 (b)가 관측되지 않는다(예: `htmx:responseError` 발화 후 본문이 폐기되는 형태). 전달 계약의 불합격 축이며, 그 **원인 라벨은 full-page-navigation 과 구분**된다 — 새 문서가 아니다.

판정 기록은 두 판별 신호 (a)·(b)의 **관측 출력 전문**과 함께 `location.href` 값을 **기록 전용 신호**로 남긴다 — href 는 내비게이션 증거가 **아니다**: 임베드 htmx 2.0.4 빌드는 `historyEnabled:true` 이고 boost 응답에 명시 `hx-push-url` override 가 없어 **성공 스왑조차 `history.pushState` 로 href 를 action URL(`action=/save?profile=…`)로 바꾼다** — 이것이 본 트리(cd99336bf)의 측정이다(측정 사슬: `/settings` 라우트 → 폼 `hx-boost` + action → override 0건 → 자산 기본값 push; `plan-audit.md` §4 D1). 기대 변화가 관측되지 않으면 그 사실도 함께 기록한다. 이 분류가 곧 판별식의 심장이다 — 「슬롯에 사유가 있다」는 사실만으로는 REQ-A 의 전달 양식(hx-boost 스왑 계약)이 검증되지 않는다.

### REQ-BO-003 — 미측정 전제의 측정 승격

The browser-observation subsystem shall promote the t1051 card-analysis premises to measured values recorded in the same session: (a) 실패 제출 시 브라우저 콘솔 에러 수, (b) 성공 제출 시 브라우저 콘솔 에러 수, (c) 실패 응답의 본문 크기(바이트). 각 값은 그것을 산출한 명령·CDP 이벤트 근거와 함께 기록되며, 기록은 「측정값」과 「t1051 전제값(실패 2 / 성공 0 / 약 110KB)」을 **구분해 표기**한다. 전제값을 측정값으로 인용하는 것은 baseline 귀속 위반이다(`verification-claim-integrity.md` §2).

### REQ-BO-004 — htmx 버전별 비-2xx 처리의 부수 기록 (갭 2)

**When** the same browser session makes htmx non-2xx (특히 4xx 검증 거부) response handling observable, the browser-observation subsystem shall record the observed behavior (스왑 여부, `htmx:responseError` 계열 이벤트 발화 여부) as an incidental record. 이 기록은 **게이팅 아니다** — 관측 불가능 세션이면 「본 세션에서 관측 불가」라는 기록 자체가 합격이고, 그 결함으로 run 을 실패 처리하지 않는다. 부수 기록은 REQ-TR400-003 의 deferred 판정 이연을 수령하는 판정 문서가 **아니다**: 관측을 기록하고, t1080 의 이연 판정에 결정적이 되는 관측이 있으면 그 사실을 리드 라우팅 메모 한 줄로 남기는 것까지다(§1.1 분할 경계).

### REQ-BO-005 — t1080 에 대한 독립

The browser-observation subsystem shall neither wait for the SPEC-WEB-TRANSPORT-001 verdict record (`.moai/reports/t1080/verdict.md`) nor cite it as a premise or basis. 분할(t1080 §1.2)은 어느 쪽이 먼저 착수해도 상대의 전제를 삼키지 않게 설계됐다. t1080 판정이 존재하면 인용하지 않고 병존하며, 본 SPEC 의 기록은 자기 측정만으로 자립해 읽혀야 한다.

### REQ-BO-006 — t1051 판정의 비소급

The browser-observation subsystem shall not retro-apply the completed SPEC-WEB-CONSOLE-017 audit verdict to this observation, nor cite it as the record's basis — 본 기록은 **자기 측정에만** 선다. REQ-A 가 이미 닫혔다는 사실은 예상일 뿐이며, 관측이 그것을 반증하면 REQ-BO-007 이 따른다.

### REQ-BO-007 — 모순 관측의 결함 라우팅 (수리 아님)

**When** the browser observation produces a result contradicting the REQ-A contract (사유가 슬롯에 도달하지 않음, 또는 풀페이지 내비게이션으로만 도달), the browser-observation subsystem shall record it as a NEW defect finding for a follow-up repair card and shall not modify `internal/web` production code. 본 SPEC 은 측정 전용이다 — 발견의 수리는 판정 기록을 근거로 하는 별도 카드의 몫이다.

### REQ-BO-008 — 증거 거처와 프로세스 정리

The browser-observation subsystem shall keep all measurement artifacts (프로브 스크립트, CDP 원본 캡처 JSON, 분류 출력) under the local evidence directory `.moai/reports/t1081/` (gitignored) — 추적 대상 소스 트리에 두지 않는다 — and shall launch and tear down the browser and server with cleanup guarantees (`timeout` 래퍼, 등록된 정리 경로, 종료 후 고아 프로세스 프로브). 브라우저와 서버는 cleanup 보장으로 기동·해체한다: `timeout` 래퍼로 밖에서 묶이고, 종료는 정리 경로가 실행하며, run 종료 시 고아 프로세스(원격 디버깅 포트를 쥔 Chrome, `moai web` 서버)가 남지 않음을 프로브로 확인한다. 판정 기록(추적되는 run 산출물)은 `.moai/reports/t1081/verdict.md` 다.

---

## §3 Success Criteria

요약 — 상세 시나리오는 `acceptance.md`:

1. REQ-BO-001/002: 실패 제출 세션에서 삼원 분류(htmx-swap / full-page-navigation / swap-failed-no-nav)가 두 판별 신호(메인 프레임 내비게이션 마커, 슬롯 DOM 변이)와 기록 전용 `location.href`의 관측 전문과 함께 기록된다 (AC-BO-001, AC-BO-002).
2. REQ-BO-003: 세 전제(콘솔 에러 수 ×2, 본문 크기)가 명령 근거와 함께 측정값으로 기록된다 (AC-BO-003).
3. REQ-BO-004: 부수 기록 또는 명시적 관측 불가 기록이 존재한다 (AC-BO-004).
4. REQ-BO-005, REQ-BO-006, REQ-BO-007: 독립성 — t1080 미인용·미대기, t1051 비소급, 운용 코드 0변경, 모순 시 결함 라우팅 (AC-BO-005).
5. REQ-BO-008: 고아 프로세스 0 + 증거 거처 준수 (AC-BO-006).

---

## §4 Constraints (HARD)

- **HARD-1** `internal/web` 의 운용 소스(`*_test.go` 포함해 어떤 파일도)는 변경하지 않는다. 측정 대상 서버는 **이 워크트리 cd99336bf 에서 빌드한 바이너리**다(REQ-BO-007, AC-BO-005).
- **HARD-2** 판별식은 살아 있는 브라우저 전용이다. 본문 소스 문자열 탐침, `curl` 본문 판독, `renderErrorPage` 소스 판독은 REQ-BO-001 의 판정 입력이 될 수 없다 — 보조 맥락은 가능하다.
- **HARD-3** 측정 전제값(t1051 카드 분석의 실패 2 / 성공 0 / 약 110KB)은 기대값으로만 쓰고 측정값으로 인용하지 않는다(HARD 에의 승격 — REQ-BO-003).
- **HARD-4** 부수 기록(htmx 비-2xx)과 임시 관측은 t1080 의 판정 지위를 침범하지 않는다 — deferred 판정의 작성은 t1080 의 소관이고 본 SPEC 은 관측 기록과 리드 라우팅 메모까지만 남긴다(§1.1, REQ-BO-004).
- **HARD-5** 원격 디버깅 포트·웹 서버 포트는 run 단계에서 **전용 값**을 택한다(t1041 이 9333 을 쓴 것은 선례일 뿐 상속 의무가 아니다). 다른 레인의 라이브 Chrome 세션과 충돌하는 포트 재사용은 금지한다.
- **HARD-6** 문서 좌표 인용은 심볌 + 트리 SHA(cd99336bf)로 앵커한다. 행 번호를 트리 간 안정 좌표처럼 인용하지 않는다.
- **HARD-7** run 단계에서 `mcp__moai__*` 도구는 MCP 서버의 project root 가 다른 트리에 spawn-frozen 돼 있으므로 `project_root` 에 **본 워크트리의 `git rev-parse --show-toplevel` 절대 경로**를 명시 전달한다.

---

## §5 판별식 관측 인벤토리 (run 단계가 관측하는 것)

| # | 관측 항목 | CDP 근거 | 충족 의미 |
|---|-----------|----------|-----------|
| 1 | 실패 제출 교환 동안 메인 프레임 내비게이션 이벤트 | `Page.enable` + `Page.frameNavigated` / `Page.loadEventFired` | REQ-BO-002 신호 (a) |
| 2 | 부스트 컨테이너 안 DOM 변이로 슬롯 신설/갱신 | 제출 전 주입한 `MutationObserver`(`Runtime.evaluate`) 또는 `DOM.setChildNodes` | REQ-BO-002 신호 (b) — 슬롯 `save__msg save__msg--error` + `role="alert"` + 실패 문구 텍스트 |
| 3 | `location.href` 값 (**기록 전용** — 내비게이션 증거 아님) | `Runtime.evaluate` | REQ-BO-002 — 기대 변화: 성공 스왑에서도 pushState 로 action URL 로 바뀜(plan-audit §4 D1 측정) |
| 4 | 제출 POST 의 응답 상태·본문 크기 | `Network.enable` + `Network.responseReceived` / `Network.getResponseBody` | REQ-BO-003 (c), 2xx 재렌더 기대의 관측 |
| 5 | 콘솔 에러 이벤트 | `Runtime.exceptionThrown`, `Runtime.consoleAPICalled(type=error)`, `Log.entryAdded(level=error)` — t1041 프로브와 동일 축 | REQ-BO-003 (a)(b) |
| 6 | 슬롯의 가시성 | 제출 후 `getComputedStyle`(display/visibility) + 오프셋 크기 | REQ-BO-001 — 「렌더됨」과 「보인다」의 구분 |
| 7 | (부수) 검증 거부 4xx 제출의 스왑 여부 | 같은 세션, 빈 필드 제출력 | REQ-BO-004 |

실패 유도 방법(어느 영속화 seam 을 실패시킬지)과 페이로드 구성은 run 단계의 설계 결정이다 — t1051 §5.1 의 9 seam 문구 표가 문구 판별의 입력이며, 유도 방식(예: 설정 파일 경로를 비가록으로 만드는 방식)은 그 문구 표를 건드리지 않는 한 run 단계가 택한다. 그 선택과 근거는 판정 기록에 한 줄 남긴다.

---

## §6 Gaps — 본 SPEC 이 닫지 않는 것

1. **서버 측 400 본문 특성화와 임베드 htmx 계약 추출** — t1080 의 고유 축이다(SPEC-WEB-TRANSPORT-001 §1.2). 본 SPEC 의 부수 기록은 그 판정을 대체하지 않는다.
2. **t1051 §6-3·§6-4 (016 개정 문구 연동, stderr 층 식별 토큰 문면)** — 본 SPEC 의 범위 밖. 브라우저 관측은 서버가 산출한 문구를 **그대로** 관측한다.
3. **다중 htmx 버전 비교** — 관측은 **현재 핀된 빌드**(임베드 자산, cd99336bf)에서의 동작만 담는다. 버전 간 비교는 범위 밖이다.

---

## §F Exclusions (What NOT to Build)

본 절은 이 SPEC 이 **만들지 않는 것**을 열거한다.

### Out of Scope — 운용 코드 변경

- `internal/web/` 의 어떤 소스(핸들러, 템플릿, CSS, JS 자산, 테스트 포함)도 변경하지 않는다(HARD-1). 브라우저 관측이 REQ-A 계약에 모순돼도 수리하지 않고 결함 발견으로 라우팅한다(REQ-BO-007).
- t1051 §F 이 배제한 축(`banner--error` 신설, 저장 경로 원자성, `log/slog` 전환, 실패 문구 i18n, 기존 비-저장 stderr 사이트 정비)을 **모두 계승 배제**한다 — 브라우저 관측 SPEC 이라 해서 그 판정이 뒤집히지 않는다.

### Out of Scope — t1080 계약의 재유도

- 서버 측 400 본문 특성화(httptest), 임베드 자산(`internal/web/assets/htmx.min.js`)의 응답 처리 테이블 기계 추출 — SPEC-WEB-TRANSPORT-001 의 소관이다. 본 SPEC 은 임베드 자산을 분석하지 않고 브라우저 **동작**을 관측한다.
- D-400 판별식 판정값의 작성 — REQ-TR400-003 사상표의 네 칸 판정은 t1080 의 산출물이다(HARD-4).

### Out of Scope — htmx 업그레이드·재핀

- 임베드 htmx 자산을 다른 버전으로 올리거나 재생성하지 않는다. 관측은 현재 핀된 빌드에 대해 성립한다(§6-3).

### Out of Scope — SPEC-WEB-CONSOLE-017 의 소급 변경

- t1051 의 산출물·판정·감사 기록을 재개·수정하지 않는다(REQ-BO-006). `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1051/` 워크트리는 보존된 읽기 전용 참조다.

### Out of Scope — 추적 대상 소스에 측정 코드 상주

- CDP 프로브 스크립트와 원본 캡처를 `internal/`·`.claude/` 등 추적 대상 트리에 커밋하지 않는다(REQ-BO-008) — t1041 선례처럼 로컬 증거 디렉터가 그 거처이다.
