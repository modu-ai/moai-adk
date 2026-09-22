---
id: SPEC-WEB-CONSOLE-017
title: "moai web console — save-failure observability: inline failure reason and stderr seam logging"
version: "0.1.0"
status: completed
created: 2026-09-22
updated: 2026-09-22
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: internal/web
lifecycle: spec-anchored
tags: web-console, save-failure, observability, htmx, stderr, error-reporting
era: V3R6
tier: M
related_specs: [SPEC-WEB-CONSOLE-016, SPEC-WEB-CONSOLE-011, SPEC-GLM-KEY-INPUT-001, SPEC-JEV-OPTIN-MEASURE-001]
issue_number: 1709
---

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-09-22 | manager-spec | 최초 draft. 근거는 카드 t1051 (issue #1709) — 저장 실패가 「실패했다」는 사실만 알리고(브라우저 콘솔 에러 2건) 「왜」 실패했는지는 어느 표면에도 없다는 측정. 관측 앵커는 본 레인이 트리 `WT-save-observability` @ `616f7451d`(로컬 develop 흡수 직후)에서 이번 실행으로 재측정했다(§1.1). REQ-A(사용자 표면)와 REQ-B(유지보수자 표면)는 리드 지시에 따라 **하나의 REQ 로 합치지 않는 두 요구**다. |

---

## §1 Context & Motivation

`moai web` 콘솔(`internal/web/`)의 설정 저장(`POST /save`, `handleSave`)이 실패하면 브라우저는 **실패했다는 사실**만 보여준다. 카드 분석의 CDP 실측(전제 — §1.1): 실패 시 콘솔 에러 **2건**, 성공 시 **0건**. 그러나 실패 **사유**는 어느 사용자 표면에도 나타나지 않는다 — 사유는 `renderErrorPage`(`internal/web/handlers.go:644`)가 그리는 풀페이지 본문(약 110KB 전제값) 안의 배너 문구인데, 폼이 `hx-boost="true"`(`internal/web/root.templ:56`)라서 htmx 는 **비-2xx 응답 본문을 스왑하지 않는다**. 사용자에게 도달하는 것은 없음이고, 유지보수자에게 도달하는 것도 없음이다.

본 SPEC 은 이 공백을 **서로 다른 두 독자를 위한 두 표면**으로 메운다. 리드 지시에 따라 두 표면은 **절대 하나의 REQ 로 병합되지 않는다**:

- **REQ-A (사용자 표면)** — 저장 실패 시 실패 사유가 기존 인라인 실패 슬롯(`save__msg save__msg--error`, `role="alert"`, `internal/web/shell.templ:274`, 스타일 `internal/web/assets/console.css:155`)에 나타난다. 풀페이지 내비게이션이 아니라 htmx 스왑 가능한 응답으로.
- **REQ-B (유지보수자 표면)** — 저장 실패가 stderr 로 한 줄 남는다. `internal/web` 의 확립된 관용구(`fmt.Fprintf(os.Stderr, ...)`, `internal/web/server.go:252` · `:290`)를 재사용하고, **어느 층/파일이 실패했는지**를 식별한다.

서버측 데이터 경로는 **이미 존재한다** — 본 SPEC 의 REQ-A 는 새 데이터를 만들지 않고 수송만 고친다:

- `settings_shell.go:40-42` — `view.BannerKind == "error"` 이면 `vm.SaveMessage = view.Banner`.
- `settings_shell.go:63-72`(`settingsSaveState`) — 배너 오류 또는 필드 오류 → `SaveState = "error"`.
- `shell.templ:273-281` — `error` 상태에서 `vm.SaveMessage` 가 비어 있지 않으면 그 값을 슬롯에 렌더한다(폴백 문구 `:279`는 서버 메시지가 없을 때만).

즉 실패 풀페이지 자체는 이미 슬롯에 사유를 담고 있다. 결함은 **수송**이다: 그 페이지가 500(`handlers.go:648`)으로 반환되고 hx-boost 가 비-2xx 를 스왑하지 않으므로, 슬롯이 있는 DOM 은 실패 응답을 결코 받지 못한다.

REQ-B 축의 현재 상태: `internal/web` 의 `log/slog` 임포트는 **0건**(본 트리 재측정), `os.Stderr` 기록은 2곳 전부 비-저장 경로다(`:252` 파일 감시, `:290` 브라우저 열기). 저장 실패는 **어떤 stderr 기록도 남기지 않는다**.

### §1.1 측정 귀속

본 문서의 모든 file:line 은 트리 `WT-save-observability` @ `616f7451d`에서 본 레인이 이번 실행으로 재측정한 값이다(배차 전제 재검증 — 9개 `a.renderErrorPage(` call site `handlers.go:497, 518, 526, 535, 545, 555, 565, 578, 591`, 정의 `:644`; `shell.templ:274`; `server.go:252/:290`; `root.templ:31` 단일 배너 클래스; `fieldsets.templ:607` 경고 배너; `console.css:155`). 흡수된 develop 커밋 21건 중 `internal/web` 접촉은 0건이라 좌표가 안정돼 있었다는 배차 전제도 함께 확인했다. **재측정하지 못한 값** — CDP 콘솔 에러 수(실패 2/성공 0), 실패 본문 약 110KB — 는 카드 분석의 전제로서 표기한 것이지 본 레인의 측정이 아니며, run-phase 의 기계 검증 대상이다(§6).

---

## §2 Requirements (GEARS)

REQ 식별 관례(본 SPEC 부터): `REQ-{도메인}-{시리즈}-{일련}` (`REQ-WC-017-NNN`) — `internal/spec/lint.go` `reqLinePattern` 이 수집하는 형태. 형제 SPEC 들의 구형 식별자(`REQ-WC16-001` 등)는 본문 인용 시 그대로 둔다.

### §2.1 REQ-A — 사용자 표면: 실패 사유가 인라인 슬롯에 도달한다

- REQ-WC-017-001: **When** a settings save POST fails at any persistence seam of `handleSave`, the console SHALL present the failure reason — the persistence-failure message identifying the failed seam — inside the existing inline save-failure alert slot (`save__msg save__msg--error`, `role="alert"`, `internal/web/shell.templ:274`) via an htmx-swappable response, without requiring a full-page navigation.
- REQ-WC-017-002: The inline failure slot SHALL carry the seam-identifying failure phrase the server produced for the failing seam (§5.1), and **While** a server-provided failure message exists, the generic fallback text (`Some fields could not be saved`, `shell.templ:279`) SHALL NOT replace it.

### §2.2 REQ-B — 유지보수자 표면: 저장 실패가 stderr 로 층을 식별한다

- REQ-WC-017-003: **When** a settings save POST fails at any persistence seam, the web server SHALL append exactly one stderr line through the established `fmt.Fprintf(os.Stderr, ...)` idiom (`internal/web/server.go:252` / `:290`), naming the failed seam/layer (§5.1).
- REQ-WC-017-004: Every web save-failure stderr line SHALL carry exactly one prefix — `moai web: ` — so that all save-failure lines are uniformly greppable; the pick and its justification are recorded in §5.2. The pre-existing `web: ` site at `server.go:252` is out of scope (§F) and stays untouched.
- REQ-WC-017-005: The stderr failure line SHALL NOT contain raw error values (`err.Error()` output) or any credential/key material — it carries the layer identification and the stable failure phrase only.
- REQ-WC-017-006: A successful save SHALL emit zero save-failure stderr lines and SHALL leave the success surface unchanged (banner `Settings saved.` with `BannerKind` ok, `handlers.go:597-601`; inline saved state).

---

## §3 Acceptance Criteria

전량은 `acceptance.md` (AC-WC17-001 .. AC-WC17-005). 각 AC 는 명령으로 반증 가능하며, 현재 코드에서 무엇 때문에 실패(RED)하는지를 함께 기록한다. 회귀 가드의 판별식은 **§4 HARD-4** (본문 문구) 다.

---

## §4 Constraints (HARD)

- **HARD-1** 영속화 단계의 **순서·개수·존재를 바꾸지 않는다**(§5.1의 9 seam). 본 SPEC 은 관측 표면만 만진다(SPEC-WEB-CONSOLE-016 HARD-1 계승).
- **HARD-2** 주입 하니스 `recordingSeams`(`internal/web/partial_apply_repro_test.go`)의 seam 목록과 시그니처를 **보존하며 확장한다** — 재작성 금지(SPEC-WEB-CONSOLE-016 REQ-WC16-010 이 본 카드를 명시적으로 예속).
- **HARD-3** 비밀값(GLM/Jev 자격증명)은 어떤 새 실패 표면(stderr 로그 포함)에도 나타나지 않는다 — 층 이름과 안정 실패 문구만(REQ-WC-017-005; REQ-WC16-005 · SPEC-WEB-CONSOLE-016 HARD-3 계승). 키의 유일한 거처는 `~/.moai/.env.glm` 등 자격증명 파일이며 웹 핸들러는 그것을 읽지 않는다.
- **HARD-4** 회귀 가드의 판별식은 **실패 문구 본문 텍스트**(어느 층이 실패했는지를 식별하는 안정 문구)다. **CSS 클래스 금지** — `banner banner--warn` 은 제출 전부터 `fieldsets.templ:607` 이 이미 렌더하므로 클래스 단정은 사전에 공허하게 참이고, `banner--error` 는 존재하지 않는다(`root.templ:29-34` 의 `bannerClass` 는 error 를 `banner banner--warn` 으로 사상). **오류값 금지** — `err.Error()` 파생 문자열은 값이라 판별식이 아니다.
- **HARD-5** 저장 실패 stderr 는 접두어 하나(`moai web: `)만 쓴다. 기존 `server.go:252` 의 `web: ` 사이트는 본 SPEC 범위 밖이며 변경하지 않는다(범위 규율).
- **HARD-6** 본 SPEC 은 `banner--error` 신설 여부(표현 전용, 운영자 대기 질문 — plan.md §F.1)에 **의존하지 않는다**. 가드와 요구는 그 결정 어느 쪽에서도 성립한다.
- **HARD-7** 실패 문구는 현재처럼 Go 리터럴이며 i18n 표면으로 옮기지 않는다(SPEC-WEB-CONSOLE-016 HARD-6 계승).

---

## §5 Measured Inventory

### §5.1 9개 저장 seam — call site, 현재 사용자 표면 문구, stderr 층 식별 토큰

| # | seam | call site (`handlers.go`) | 현재 배너 문구(안정부) | stderr 층 식별 토큰 |
|---|------|---------------------------|------------------------|---------------------|
| 1 | `writePreferences` | :497 | `could not save profile preferences:` | `writePreferences` |
| 2 | `syncToProject` | :518 | `profile preferences saved, but project config sync failed:` | `syncToProject` |
| 3 | `writeProjectConfig` | :526 | `profile preferences saved, but project config write failed:` | `writeProjectConfig` |
| 4 | `writeProjectNestedConfig` | :535 | `profile preferences saved, but project nested config write failed:` | `writeProjectNestedConfig` |
| 5 | `applySchemaEdits` | :545 | `profile preferences saved, but section config write failed:` | `applySchemaEdits` |
| 6 | `applyPerfTierEdits` | :555 | `profile preferences saved, but performance_tier apply failed:` | `applyPerfTierEdits` |
| 7 | `patchAgentFM` | :565 | `settings saved, but agent override write failed:` | `patchAgentFM` |
| 8 | `glmcred.Save` | :578 | `settings saved, but GLM credential write failed:` | `glmcred.Save` |
| 9 | `jevcred.Save` | :591 | `settings saved, but Jev credential write failed:` | `jevcred.Save` |

주: 7-9번의 `settings saved, but` 도입부는 SPEC-WEB-CONSOLE-016 REQ-WC16-007 이 과잉 주장으로 판정해 두었고 그 수리는 미착수(draft)다. 본 SPEC 의 REQ-WC-017-002는 **서버가 산출하는 문구의 충실한 전달**을 요구하므로 016 의 개정이 문구를 고치면 본 SPEC 의 전달도 따라가며, 가드는 표의 안정부를 016 개정과 같은 커밋에서 갱신한다(§6-3).

### §5.2 stderr 접두어 결정 — `moai web: `

두 기존 사이트는 접두어가 갈라져 있다: `server.go:252` `web: ` / `server.go:290` `moai web: ` (1:1, 다수파 없음). 저장 실패 로그의 소비자는 **이슈에 붙여넣는 사용자와 그것을 읽는 유지보수자**다(issue #1709 자체가 그 경로로 왔다). `moai web: ` 를 택한다:

1. **생산자 자기식별** — `moai web: ` 는 바이너리+서브시스템을 한 줄에서 식별한다. 다른 도구가 남긴 일반적인 `web: ...` 행과 같은 스트림·파일에 섞여도 분리 수집된다.
2. **grep 안정성** — issue 보고에서 `moai web:` 은 이 콘솔 고유의 토큰이라 필터링이 정확하다; `web: ` 은 넓어 오탈 행을 당긴다.
3. **선례 정합** — `:290` 은 이미 「사용자가 관측해 보고하는 진단」 클래스(`브라우저 열기 실패`)에 이 접두어를 쓴다. 저장 실패 로그는 같은 클래스다.

비용은 없다(길이 4바이트). `:252` 를 함께 통일하는 것은 인접 수리라 본 SPEC 이 손대지 않는다(§F, HARD-5).

### §5.3 관측된 서버측 데이터 경로 (REQ-A 가 재사용하는 것)

- `settings_shell.go:40-42` — `BannerKind == "error"` → `vm.SaveMessage = view.Banner`.
- `settings_shell.go:63-72` — `settingsSaveState`: 배너/필드 오류 → `SaveState = "error"`.
- `shell.templ:273-281` — `error` 분기: `role="alert"` 슬롯 + `vm.SaveMessage` (폴백 `:279`).
- `root.templ:48-56` — `#settings-form` + `hx-boost="true"` (수송 결함의 한쪽 끝).
- `handlers.go:644-649` — `renderErrorPage`: 풀페이지 + `http.StatusInternalServerError` (수송 결함의 다른 끝).

---

## §6 Gaps carried forward — 미측정이지 부재가 아니다

1. **브라우저 실측.** 콘솔 에러 수(실패 2/성공 0)와 본문 크기(~110KB)는 카드 분석의 CDP 전제다. 본 레인은 재측정하지 않았고 REQ-A 의 인수는 렌더/DOM 계층의 기계 검증으로 세운다(AC-WC17-001) — 브라우저 재측정은 run-phase 의 선택 검증이다.
2. **htmx 비-2xx 스왑 동작의 버전별 세부.** 폼이 hx-boost 라는 것과 실패 응답이 500 풀페이지라는 것은 재측정했으나, htmx 버전별 응답 처리 규칙의 세부는 검증하지 않았다. **어떤 기구(2xx 재응답, 스왑 헤더, `htmx:responseError` 핸들러 등)로 REQ-A 를 달성할지는 run-phase 의 설계 결정이다** — 본 SPEC 은 관측 가능한 결과(사유가 슬롯에 보인다)만 요구한다.
3. **016 개정과의 문구 연동.** §5.1 각주 — 016 이 배너 문구를 고치면 본 SPEC 의 전달 표면과 가드 리터럴이 같은 커밋에서 갱신되어야 한다. 어느 쪽이 먼저 착수하는지에 따라 가드 리터럴의 초기 값이 달라질 수 있다.
4. **stderr 층 식별 토큰의 최종 문면.** §5.1의 토큰(Go 심볼명)은 식별성의 하한이며, run-phase 가 구현 형태에 맞게 문면을 다듬을 수 있다. 단층 식별이 가능하기만 하면 REQ-WC-017-003 은 충족된다.

---

## §F Exclusions (What NOT to Build)

본 절은 이 SPEC 이 **만들지 않는 것**을 열거한다. 아래 항목은 전부 out of scope 이며, 삭제 대상이 아니라 후속 판단의 입력이다.

### Out of Scope — banner 변형 분리 (`banner--error` 신설)

- `banner--error` 클래스를 신설하고 오류/경고 배너 변형을 나누는 일은 **단일 강조 설계 결정을 뒤집는 일**이다(`root.templ:26-28` 주석 — 설계 시스템의 위험 계열 강조는 하나뿐이며 오류 배너가 그 유일 변형이다).
- 리드가 운영자에게 상신한 **운영자 대기 질문**이다(plan.md §F.1). 본 SPEC 은 그 결정과 무관하게 성립한다(HARD-6) — 가드는 본문 문구 판별식이라 클래스 변화에 영향받지 않는다.
- 결정이 「신설」로 나면 표현 계층의 후속 작업이며, 「유지」로 나면 아무 일도 없다. 어느 쪽이든 본 SPEC 의 요구는 변하지 않는다.

### Out of Scope — 저장 경로 원자성/트랜잭션

- 9 seam 의 부분 적용 상태를 원자화하거나 되돌리는 일. SPEC-WEB-CONSOLE-016 §F 가 별건으로 판정·분리한 축이며 본 SPEC 은 선취하지 않는다.

### Out of Scope — internal/web 로그 기구 전환

- `log/slog` 등 구조화 로거의 도입. 본 SPEC 은 기존 `fmt.Fprintf(os.Stderr, ...)` 관용구를 **재사용**하는 것이 요구다(REQ-WC-017-003). 로거 전환은 별도의 설계 결정이다.

### Out of Scope — 실패 문구의 i18n 화 및 신규 사용자 문구 작성

- 배너 문구를 i18n 키로 옮기는 일(HARD-7), 그리고 REQ-A 전달을 위한 새 문구 작성. 본 SPEC 은 서버가 이미 산출하는 문구의 전달만 요구한다.

### Out of Scope — 기존 비-저장 stderr 사이트의 정비

- `server.go:252` 의 `web: ` 접두어를 `moai web: ` 로 맞추는 일을 포함한 기존 2곳의 변경(§5.2, HARD-5). 범위 규율 — 인접 수리는 별도 판단 대상이다.
