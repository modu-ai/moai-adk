---
id: SPEC-WEB-TRANSPORT-001
title: "moai web console — validation-400 boosted-response transport: htmx 4xx body-swap discriminator measurement and defect judgment"
version: "0.1.0"
status: completed
created: 2026-09-22
updated: 2026-09-22
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: internal/web
lifecycle: spec-anchored
tags: web-console, htmx, hx-boost, 400, transport, measurement, discriminator
era: V3R6
tier: M
related_specs: [SPEC-WEB-CONSOLE-017]
---

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-09-22 | manager-spec | 최초 draft. 근거는 카드 t1080 (t1051 sync-audit F5 + Residual-risk(3)) — 검증 실패 400 경로(`internal/web/handlers.go:490`)가 저장 실패(비-2xx) 수송 결함과 같은 축의 잠재 재현이라는 가설을, **측정-선행-판정 후행**으로 다룬다. 지금까지 측정된 것은 400 상태 코드뿐이고, htmx 의 4xx 본문 처리는 본 트리에서 미측정이다. 본 SPEC 의 산출물은 판별식(discriminator)과 그에 게이트된 결함 판정 기록이며, 400 경로의 수리는 범위 밖이다. |

---

## §1 Context & Motivation

`moai web` 콘솔의 설정 저장은 폼 전체가 `hx-boost="true"`(`internal/web/root.templ`)라서 htmx AJAX 풀페이지 스왑으로 처리된다. SPEC-WEB-CONSOLE-017(t1051)은 저장 **지속화(persistence)** 실패 경로가 500 으로 답할 때 htmx 가 비-2xx 본문을 스왑하지 않아 실패 사유가 브라우저에 도달하지 않는다는 수송 결함을 확인하고, 해당 경로를 2xx 재렌더(`renderErrorPage`, `internal/web/handlers.go:649-669`)로 바꿔 닫았다.

그러나 같은 핸들러에는 두 번째 오류 경로가 있다. **필드 검증 거부**(validator 가 한 개 이상 실패)일 때 핸들러는 전체 필드 오류를 병합하고 배너("Validation failed — no changes were saved.")와 필드별 오류를 싣은 풀페이지를 **상태 400** 으로 렌더한다(`internal/web/handlers.go:483-491`, `a.render(w, http.StatusBadRequest, view)` 가 490행). 이 400 본문이 hx-boost 아래 브라우저에 실제로 도달하는지는 **아무도 측정하지 않았다** — t1051 sync-audit F5 가 이를 INFO 등급 잠재 재현으로 남겼고 Residual-risk(3) 가 미수리 상태를 명시했다.

카드 t1080 의 지시는 이 결함을 **추측으로 판정하지 말라**는 것이다. 판별식을 먼저 측정하고, 그 측정에 게이트된 판정만을 쓴다. 본 SPEC 은 그 측정-판정 절차를 정의한다.

### §1.1 측정된 사실 (본 레인, 트리 WT-verify-400-path @ 0314801c2)

| # | 측정 | 근거 |
|---|------|------|
| F1 | 핀된 htmx 는 **로컬 임베드**다 — CDN 없음. `//go:embed assets/htmx.min.js` 가 `internal/web/assets/htmx.min.js`(50,917 바이트)를 바이너리에 넣고 `/static/htmx.min.js` 로 서브한다 | `internal/web/assets.go:20-24`, `internal/web/htmx_test.go` |
| F2 | 임베드된 htmx 버전은 **2.0.4**다 — 자산 내부에 `version:"2.0.4"` 문자열 존재 | `grep -o 'version:"[0-9.]*"' internal/web/assets/htmx.min.js` |
| F3 | 2.0.4 빌드에 **`responseHandling` 메커니즘이 존재**한다 — 자산 내부에 해당 식별자 1회 출현. htmx 2.x 의 기본 응답 처리 테이블은 상태 코드 패턴별로 swap/error 를 정의하며, 4xx 는 기본적으로 스왑하지 않는 것이 공식 문서상 기본값이다 (단, **본 트리 자산에서의 실제 기본값은 run 단계에서 추출해야 할 미측정 값**이다) | `grep -c responseHandling internal/web/assets/htmx.min.js` = 1 |
| F4 | 검증 거부 경로는 배너 + 필드별 오류를 실은 풀페이지를 400 으로 렌더한다 (코드 판독; **본문 문자열 포함 여부는 아직 테스트로 증명되지 않은 주장**이다) | `internal/web/handlers.go:483-491` |

F3 이 이 SPEC 의 실현 가능성을 결정한다. htmx 가 CDN 로드가 아니라 **커밋된 로컬 자산**이므로, 4xx 스왑 계약을 브라우저 없이 — 임베드 자산에서 기본 응답 처리 테이블을 기계적으로 추출해 상태 400 에 평가하는 방식으로 — 측정할 수 있다. 이것이 판별식의 server-side 측정 경로다.

### §1.2 카드 분할 결정 (t1080 vs t1081) — 설계 결정, 명시적 기록

형제 카드 t1081 은 실패 제출 시 `save__msg--error` 슬롯의 브라우저 CDP 실측(REQ-A 갭)을 담당하고, 같은 측정에서 관측 가능하면 htmx 버전별 비-2xx 처리를 **부수 기록**으로 남긴다. 중복을 막기 위해 본 SPEC 은 분할을 다음과 같이 확정한다:

| | t1080 (본 SPEC) | t1081 |
|---|---|---|
| 측정 대상 | 서버 측 400 본문 특성화 + **임베드 htmx 자산의 비-2xx 스왑 계약** (기계적, 브라우저 불요) | 실제 브라우저에서의 2xx 실패 재렌더 전달 (REQ-A 갭) + 관측 가능 시 4xx 처리 부수 기록 |
| 산출물 | 판별식 판정값(D-400) + 결함 판정 기록 (계약 수준 신뢰도) | 브라우저 관측 기록 (browser-observed 신뢰도) |
| 판정 지위 | 독립 판정 — t1051 판정을 인용하지 않고 자기 측정에만 근거 | t1080 판정의 런타임 확인은 담당하지 않아도 되며(선택), 자기 REQ-A 갭은 독립 |

분할 근거: htmx 가 로컬 임베드(F1)라서 클라이언트 계약이 서버 트리 안에서 기계적으로 측정 가능하므로, 브라우저 세션 없이도 판별식과 판정을 낼 수 있다. 브라우저 런타임 실행의 실측은 t1081 의 고유 축으로 남긴다 — 두 카드가 같은 CDP 세션을 요구하지 않고, 어느 쪽이 먼저 착수해도 상대의 전제를 삼키지 않는다.

---

## §2 Requirements (GEARS)

### REQ-TR400-001 — 400 본문 특성화 (server-side characterization)

검증 거부 시 `handleSave` 가 돌려주는 400 응답은 오류 배너 문구와 병합된 모든 필드별 오류 메시지를 본문에 담은 풀페이지여야 하며, run 단계는 httptest 기반 특성화 테스트로 이 특성을 **본문 문자열 수준에서 검증**해야 한다. 현재 이 주장은 코드 판독에 근거한 미증명 주장이다(§1.1 F4) — 측정으로 승격해야 한다.

### REQ-TR400-002 — 판별식 측정 (discriminator, 측정-선행)

**When** run 단계가 판별식 D-400 을 측정할 때, 측정 하위시스템은 (a) 상태 400 에 대한 스왑 결과를 **커밋된 임베드 자산**(`internal/web/assets/htmx.min.js`, htmx 2.0.4)에서 기본 응답 처리 기본값을 기계적으로 추출해 평가하여 기록하고(D-400b), (b) 그 기록이 완료되기 **전에는** 결함 존재 판정을 작성해선 안 된다. 외부 문서만으로 판정하는 것은 측정이 아니다 — 자산 추출 증거(매칭된 테이블 조각 전문)가 함께 기록되어야 한다.

**D-400 판별식 정의(운용화):** "4xx 부스트 응답에서 응답 본문이 페이지로 스왑되는가." 두 하위 측정으로 나뉜다:

- **D-400a** (서버): 400 본문이 사용자 가시 피드백(배너 + 필드별 오류)을 운반한다 — REQ-TR400-001 의 측정.
- **D-400b** (클라이언트 계약): 임베드 htmx 2.0.4 빌드의 기본 응답 처리 테이블이 상태 400 을 `swap: false`(본문 폐기, 오류 이벤트)로 사상한다 — REQ-TR400-002 의 측정.

### REQ-TR400-003 — 게이트된 결함 판정 (judgment procedure)

**While** D-400b 가 임베드 자산에서 판정 불가(미확정 추출)인 상태일 때, 측정 하위시스템은 추론으로 판정을 쓰지 말고 판정을 브라우저 수준 실행자(t1081)에게 이연해야 하며, 그 이연 기록이 곧 판정 기록이다. D-400a 와 D-400b 가 모두 측정됐을 때 판정은 다음 사상에 게이트된다:

| D-400a | D-400b | 판정 등급 |
|--------|--------|-----------|
| true (본문이 피드백 운반) | swap: false | **defect-present (contract-level)** — 서버가 내보낸 피드백이 클라이언트에서 폐기됨 |
| true | swap: true | **defect-absent** — 부스트 아래에서도 400 본문이 스왑됨 |
| false (본문이 피드백을 운반하지 않음 — AC-TR400-001 FAIL) | 측정값과 무관 | **measurement-invalid (서버 계약 발산)** — 판별식의 서버 축이 성립하지 않으므로 어떤 결함 판정(defect-present/absent 포함)도 쓰지 않는다. 발산 사실을 판정 기록에 기록하고 특성화를 재측정한다(프로브 결함 — 페이로드·경로·고정값 오류 — 우선 배제, 실행자 manager-develop). 재측정 후에도 false 가 유지되면 코드 판독(§1.1 F4)과 실측의 모순이므로 카드 전제의 재판정을 운영자에게 에스컬레이션한다 |
| true (D-400a 가 성립한 상태 — D-400a=false 인 측정은 위 measurement-invalid 칸이 우선한다) | 미판정 | **deferred-to-browser** — t1081 가 실행자 |

D-400a=false 는 결함 부재의 증거가 아니라 **발산 사실**이다 — 특성화 기대가 깨졌다는 것은 판별식의 전제(서버가 피드백을 내보낸다)가 측정되지 않았다는 뜻이지, 수송 결함이 없다는 뜻이 아니다. 그래서 이 칸은 판정이 아니라 재측정·에스컬레이션으로 닫힌다. AC-TR400-001 의 모든 관측 결과는 위 네 칸 중 하나에 분류된다.

### REQ-TR400-004 — 판정 기록의 신뢰도 등급

판정 기록(`.moai/reports/t1080/verdict.md`)은 판정 등급과 함께 **신뢰도 등급**(confidence class)을 명시해야 한다: `contract-level`(임베드 자산 추출 + 핀 버전 공식 문서 의미론 — 본 SPEC 의 상한) 또는 `browser-observed`(t1081 수준). 계약 수준 판정을 브라우저 관측으로 표기하는 것은 관측되지 않은 검증 주장이다.

### REQ-TR400-005 — t1051 판정의 비소급 독립

**When** 결함 판정이 작성될 때, 판정 하위시스템은 SPEC-WEB-CONSOLE-017(t1051)의 판정을 근거로 인용하거나 소급 적용해선 안 되고, 운용 코드를 수정해선 안 되며, 감사 종료된 SPEC-WEB-CONSOLE-017 을 재개해선 안 된다. defect-present 판정은 **후속 수리 카드**를 위한 새 발견으로 기록된다 — 400 경로의 수리 자체는 본 SPEC 범위 밖이다.

---

## §3 Success Criteria

요약 (상세 시나리오는 acceptance.md):

1. REQ-TR400-001: httptest 특성화 테스트가 400 본문의 배너 + 필드별 오류 포함을 증명 (AC-TR400-001).
2. REQ-TR400-002: 임베드 자산 기계 추출로 D-400b 가 측정·기록되고, 추출 증거가 남는다 (AC-TR400-002).
3. REQ-TR400-003/004: 판정 기록이 측정값에 게이트되고 판정 등급(defect-present(contract-level) / defect-absent / measurement-invalid / deferred-to-browser 중 하나 — measurement-invalid 시 판정 대신 재측정·에스컬레이션 기록으로 적용)과 신뢰도 등급을 명시한다 (AC-TR400-003).
4. REQ-TR400-005: 운용 코드 무변경 + SPEC-WEB-CONSOLE-017 무접촉 + t1051 판정 비인용 (AC-TR400-004).
5. 품질 게이트: 영향 패키지 환경스크럽 테스트 통과 (AC-TR400-005).

## §4 Out of Scope

### Out of Scope — 400 경로의 수리 (remediation)

- 판정이 defect-present 여도 `internal/web/handlers.go` 의 검증 거부 경로(400 렌더)를 수정하지 않는다. 수리는 판정 기록을 근거로 하는 **후속 카드**의 몫이다.
- 2xx 재렌더로의 전환(t1051 이 500 경로에 한 것과 같은 수선)을 이 SPEC 에서 시도하지 않는다.

### Out of Scope — 브라우저 런타임 확인 (browser-level confirmation)

- 실제 브라우저에서의 4xx 스왑 관측은 t1081 (CDP 실측)의 고유 축이다. 본 SPEC 은 계약 수준 측정까지만 담당하고, 브라우저 세션·CDP·DOM 검사를 요구하지 않는다.
- 임베드 htmx 자산을 JS 런타임에서 실행하는 방식(headless 실행)도 채택하지 않는다 — 자산에서 기본 테이블을 추출·평가하는 기계적 경로가 존재하므로(F1-F3) 더 무거운 실행은 불요다.

### Out of Scope — SPEC-WEB-CONSOLE-017 소급 변경

- t1051 의 산출물, 판정, 감사 기록을 재개·수정·재적용하지 않는다. `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1051/` 워크트리는 보존된 읽기 전용 참조다.
- REQ-A(브라우저 표면) 갭 자체는 t1081 의 범위다.

### Out of Scope — htmx 업그레이드·재핀

- htmx 2.0.4 를 다른 버전으로 올리거나 자산을 재생성하지 않는다. 판별식은 **현재 핀된 빌드**에 대해 성립한다.
