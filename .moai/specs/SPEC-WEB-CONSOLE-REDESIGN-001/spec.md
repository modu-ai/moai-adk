---
id: SPEC-WEB-CONSOLE-REDESIGN-001
title: "moai web 설정 콘솔 재설계 — 죽은 설정 제거 · 9탭 재편 · 위젯 정책 · GLM 정직 표면 · 프로필 UI 통합"
version: "0.1.0"
status: completed
created: 2026-08-08
updated: 2026-08-13
author: manager-spec
priority: P1
phase: "v3.1.0"
module: "internal/web"
lifecycle: spec-anchored
tags: "web, console, settings, schema-driven, i18n, glm, profile, autonomy, radio, tabs"
tier: L
era: V3R6
related_specs: [SPEC-WEB-CONSOLE-011, SPEC-WEB-CONSOLE-012, SPEC-WEB-CONSOLE-013, SPEC-WEB-CONSOLE-014, SPEC-GLM-KEY-INPUT-001, SPEC-AUTONOMY-TIERS-001, SPEC-MODEL-TIER-PLANTYPE-001]
---

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-08-08 | manager-spec | 최초 draft. 오케스트레이터 실측 브리프(스키마 주도 폼 구조 · 죽은 설정 grep · GLM 단일 전달 채널 제약 · 프로필 UI 중복 · autonomy stub) 기반. Tier L이나 산출물은 3종(spec/plan/acceptance)으로 한정 — §G.1 편차 기록. |

---

## §A Context & Motivation

`moai web` 설정 콘솔의 폼은 **스키마 주도(schema-driven)** 다. `internal/settings/schema_sections.go`의 `FieldDef` 목록이 SSOT이고, `internal/web/fieldsets.templ`의 제네릭 위젯(`schemaTextRow` / `schemaNumberRow` / `schemaToggleRow` / `schemaSelectRow` / `schemaRadioRow`)이 이를 렌더하며, `internal/web/schemaform.go`의 `consoleTabs()`(현재 7탭)와 `schemaSectionMetas()`가 무엇을 렌더할지 결정한다. 영속화는 `settings.ApplySchemaEdits`가 담당한다(`PersistSeam` = yamlpatch, `PersistTypedSection` = typed).

이 구조 위에 다섯 갈래의 결함이 누적됐다.

### §A.1 실측 Findings

전체 명령과 출력은 acceptance.md §B에 있다. 라인 앵커는 실측 시점 기준이며 run-phase에서 content-token으로 재검증한다.

| # | 결함 | 실측 근거 | 처분 |
|---|------|-----------|------|
| F1 | **렌더 표면 누락** | `settings.SchemaSectionIDs()`는 11개 섹션(quality_extras, git_strategy, llm, workflow, harness, ralph, feedback, observability, security, handoff, cache)을 등록하지만 `schemaSectionMetas()`는 3개(LLM / Workflow / Report)만 렌더한다. `git_strategy`(mode + 3× merge_method)를 포함한 8개 섹션은 FieldDef가 존재하나 렌더 표면이 **없다** | git_strategy를 최소 복구 대상으로 승격, 신규 탭에 배치 |
| F2 | **죽은 설정 노출** | `workflow.token_budget.{plan,run,sync}` — 접근자 `WorkflowPlanTokens`/`WorkflowRunTokens`/`WorkflowSyncTokens`(`internal/config/workflow_accessors.go`) 호출자 0건, 산문 소비자 0건. `workflow.auto_clear.{enabled,after_plan,after_run,token_threshold}` — `WorkflowAutoClearEnabled` 호출자 0건, 산문 소비자 0건 | 웹 편집 표면에서 7개 필드 철거 (yaml 키·struct 멤버·접근자는 보존) |
| F3 | **살아 있는 설정** | `workflow.agentic_loop.max_iterations` — Go 호출자 없음이나 `moai.md` 스킬이 산문 소비. `workflow.loop_prevention.*` — Go 호출자 없음이나 `loop.md` + docs-site 우선순위 표가 산문 소비. `workflow.worktree.*` / `workflow.branch_guard.enabled` — Go 소비자 실존(`internal/cli/worktree_advisory.go`, `internal/cli/session_worktree.go`, `internal/hook/pre_tool.go`) | 전부 편집 표면 **유지** — F2와 함께 일괄 제거하지 말 것 |
| F4 | **GLM effort 단일 채널** | `internal/cli/glm.go` `setGLMEnv`가 `llm.glm.models.{high,medium,low,fable}`에서 `ANTHROPIC_DEFAULT_{OPUS,SONNET,HAIKU,FABLE}_MODEL`을 주입한다. 그러나 추론 강도(reasoning effort)의 전달 경로는 `ANTHROPIC_REASONING_EFFORT` 환경변수 **하나뿐**이며(`internal/template/glm_effort_overlay.go` `SessionGLMReasoningState()`에서 파생), 티어별 추론 채널은 존재하지 않는다. 기존 코드 주석은 z.ai의 해당 환경변수 준수 여부를 **UNVERIFIED**로 표기한다 | 4개 티어 effort를 저장은 하되, **어느 티어가 실제 적용되는지 배지로 명시**하고 나머지는 저장 전용(store-only)으로 라벨링 |
| F4a | **적용 원천 정정 (브리프 대비 실측 정정)** | 브리프는 "티어 중 하나만 실제 적용된다"고 기술했으나 실측은 더 강하다: `SessionGLMReasoningState()`는 조건 없이 `reasoning-max`를 반환하고, `SessionGLMReasoningStateForEffort(effort)`는 **세션 단위 effort 환경설정**(LLM 탭의 `effort_level`)을 z.ai 3-상태로 collapse한 값을 반환한다(`internal/cli/launcher.go`가 이 값을 주입). 즉 `llm.glm.models.*` 티어 어느 것의 effort도 런타임에 적용되지 **않는다** | REQ-WCR-033의 배지는 "적용되는 티어"가 아니라 **적용 원천**을 명시한다 |
| F5 | **프로필 UI 중복** | `profileSwitch`(root.templ:202)는 select + "selected: X" 텍스트이고 `app.js:59`가 change 시 이동시킨다. `profileManager`(root.templ:250)는 목록 + 전환 링크 + 삭제 폼 + 생성 폼을 가진 **별도 카드**로 select와 기능이 겹친다 | profileManager 카드 철거, 추가/편집/삭제 컨트롤을 프로필 바로 이동 |
| F6 | **autonomy 미완 stub** | `internal/web/handlers.go:206` 상수 `autonomyToggleLinkHTML`이 `</body>` 앞에 맨 `<a href="/autonomy/tiers">` 링크를 주입한다. `internal/web/autonomy.go` `handleAutonomyTiers`는 GET 전용이고 form/action이 **없어** 선택해도 저장되지 않는다. 디자인 시스템 CSS 클래스도 `data-i18n`도 없다 | 콘솔 정식 필드로 승격하거나 stub 제거 — 둘 중 하나로 **결론을 내고 기록** |

### §A.2 노출 판정 원칙 (SPEC-WEB-CONSOLE-014 §A.2 계승)

1. **편집 가능**: 행동적 소비자(Go 런타임 reader 또는 SPEC/스킬로 문서화된 산문 소비자 계약)가 존재하는 스칼라/enum 키.
2. **읽기 전용 표시**: reader는 있으나 리스트/맵이라 form UI에 부적합하거나, 편집 노출이 오해를 유발하는 키.
3. **비노출**: reader 부재(scaffold), governance frozen, dormant 키.

본 SPEC은 여기에 **네 번째 축**을 추가한다: **위젯 정직성** — 값의 도메인이 닫힌 집합인데 자유 텍스트로 렌더되면 사용자는 유효하지 않은 값을 입력할 수 있고, 서버 검증이 이를 거절하더라도 UI가 이미 거짓말을 한 뒤다. 닫힌 집합은 닫힌 위젯으로 렌더한다.

### §A.3 백워드 호환 하우스룰

웹 **표면**에서 필드를 제거하는 것은 yaml 키, struct 멤버, 로더 경로 중 어느 것도 제거하지 않는다. 이는 SPEC-WEB-CONSOLE-011 M4 다이어트와 SPEC-WEB-CONSOLE-012 REQ-WC12-006이 확립한 선례이며 본 SPEC도 동일하게 따른다.

---

## §B Requirements (GEARS)

### §B.1 M1 — 죽은 설정 철거 + 렌더 표면 복구

- **REQ-WCR-001**: 설정 스키마는 `workflow.token_budget.plan` / `.run` / `.sync`에 대한 편집 가능 필드를 제공하지 **않아야 한다(shall not)**. **When** 저장 요청이 이 세 이름 중 하나를 담은 것이 감지되면, 폼 파서는 해당 이름을 편집 대상으로 수집하지 않아야 한다.
- **REQ-WCR-002**: 설정 스키마는 `workflow.auto_clear.enabled` / `.after_plan` / `.after_run` / `.token_threshold`에 대한 편집 가능 필드를 제공하지 **않아야 한다(shall not)**.
- **REQ-WCR-003**: REQ-WCR-001과 REQ-WCR-002의 필드 철거는 `.moai/config/sections/workflow.yaml`의 해당 yaml 키, `internal/config`의 struct 멤버, 그리고 그 접근자 함수를 제거하지 **않아야 한다(shall not)**. **When** 철거 이후 기존 `workflow.yaml`이 로드되면, 로더는 해당 키를 오류 없이 읽어야 한다(shall).
- **REQ-WCR-004**: 콘솔은 `git_strategy` 섹션을 렌더 표면에 복구해야 한다(shall) — `git_strategy.mode` 및 3개 profile(manual/personal/team)의 `merge_method`. **Where** 해당 FieldDef가 이미 `settings.SchemaSectionIDs()`에 등록되어 있는 경우, 복구는 신규 FieldDef 선언이 아니라 렌더 메타(`schemaSectionMetas()`) 등록으로 수행해야 한다.

### §B.2 M2 — 탭 7→9 재편

- **REQ-WCR-010**: 콘솔 탭 네비게이션은 정확히 9개 탭을 다음 순서로 제공해야 한다(shall): 사용자 정보 / 언어 / LLM / 서드파티 LLM / 워크플로우 / Git·워크트리 / 감사 / 에이전트 / 리포트. 순서는 렌더 순서와 일치해야 한다.
- **REQ-WCR-011**: 워크플로우 탭은 `workflow.execution_mode`, `workflow.default_mode`, `workflow.agentic_loop.max_iterations`, `workflow.loop_prevention.*`를 편집 필드로 유지해야 한다(shall). 이 키들은 산문 소비자가 실존하므로 REQ-WCR-001/002의 철거 대상이 **아니다**.
- **REQ-WCR-012**: Git·워크트리 탭(신규)은 `git_strategy.mode`, 3× `git_strategy.<profile>.merge_method`, `workflow.worktree.{auto_cleanup,auto_create,auto_merge,tmux_preferred}`, `workflow.branch_guard.enabled`를 렌더해야 한다(shall).
- **REQ-WCR-013**: 감사 탭(신규)은 `workflow.audit.model`과 `workflow.audit.gates.{claude,codex,glm}`를 렌더해야 한다(shall).

### §B.3 M3 — 위젯 정책

- **REQ-WCR-020**: 모든 bool 타입 필드는 체크박스가 아니라 **사용 / 미사용** 2옵션 라디오 그룹으로 렌더되어야 한다(shall). 라벨은 4개 로케일 전부에 대응 키를 가져야 한다. **Where** 변경 지점이 공용 위젯인 경우, 변경은 `schemaToggleRow` 한 곳에서 이루어져 모든 탭에 적용되어야 한다.
- **REQ-WCR-021**: REQ-WCR-020의 위젯 변경은 hidden companion 입력(`<name>__present`)을 보존해야 한다(shall). **When** bool 필드가 제출되지 않은 상태로 저장 요청이 도착하면, 파서는 이를 "미제출 → 값 보존"으로 해석해야 하며 `parseSchemaForm`의 파싱 로직은 변경되지 않아야 한다(shall not).
- **REQ-WCR-022**: 값 도메인이 닫힌 집합인데 현재 자유 텍스트로 렌더되는 모든 필드 — `workflow.execution_mode`, `workflow.default_mode`, `workflow.audit.model`, `workflow.audit.gates.{claude,codex,glm}`, `harness.*` — 는 select 또는 라디오로 렌더되고 멤버십 검증을 가져야 한다(shall). **When** 저장 요청이 닫힌 집합 밖의 값을 담은 것이 감지되면, 설정 적용 계층은 해당 쓰기를 거부해야 한다.
- **REQ-WCR-023**: 값 도메인이 진정으로 열려 있는 필드 — `feedback.repository`, `observability.{report_dir,trace_dir}`, GLM API 키 — 만 자유 텍스트 입력으로 잔류해야 한다(shall). 그 외 자유 텍스트 잔류는 허용되지 않는다(shall not).

### §B.4 M4 — 서드파티 LLM 탭

- **REQ-WCR-030**: `llm.glm.models.{high,medium,low,fable}` 4개 필드는 닫힌 집합 `{glm-5.2, glm-5.1, glm-4.7, glm-4.5-air}`에 대한 select로 렌더되어야 한다(shall).
- **REQ-WCR-031**: 콘솔은 4개 티어 각각에 대해 추론 강도 select를 `{Max, High, None}` 집합으로 제공해야 하며(shall), 기본값은 Fable=Max, Opus(high)=Max, Sonnet(medium)=High, Haiku(low)=None이어야 한다.
- **REQ-WCR-032**: REQ-WCR-030/031의 티어 라벨은 Claude 대응 이름(Opus / Sonnet / Haiku / Fable)으로 표시되어야 하며(shall), 내부 키(`high` / `medium` / `low` / `fable`)는 변경되지 않아야 한다(shall not).
- **REQ-WCR-033**: **Where** 추론 강도의 런타임 전달 채널이 세션 전역 단일 환경변수 하나뿐인 동안, 서드파티 LLM 탭은 실제 적용 원천(§A.1 F4a — 세션 단위 effort 환경설정에서 파생되며 티어별 필드에서 파생되지 **않는다**)을 명시하는 배지/라벨을 렌더해야 하며(shall), 4개 티어 추론 강도 값이 저장 전용(store-only)임을 표시해야 한다. 콘솔은 티어별 추론 강도 값이 런타임에 적용된다고 암시하지 **않아야 한다(shall not)**.
- **REQ-WCR-034**: GLM API 키 입력은 기본 렌더에서 저장된 값을 절대 에코하지 않아야 하며(shall not — `value=""` 무조건), 명시적 "표시(reveal)" 토글이 조작된 경우에만 서버에서 평문을 1회 조회해 표시해야 한다(shall). **When** reveal 요청이 루프백이 아닌 출처에서 감지되면, 서버는 해당 요청을 거부해야 한다.

### §B.5 M5 — 프로필 UI 통합

- **REQ-WCR-040**: 콘솔은 `profileManager` 카드를 렌더하지 **않아야 한다(shall not)**.
- **REQ-WCR-041**: 프로필 생성 / 이름변경 / 삭제 컨트롤은 프로필 바에 인접해 배치되어야 한다(shall). **Where** 생성·삭제·이름변경이 각각 독립 POST 폼인 경우, 해당 `<form>` 요소는 메인 설정 폼 **바깥**에 위치해야 한다 — 중첩 form은 유효하지 않은 HTML이다.
- **REQ-WCR-042**: 서버는 프로필 이름변경(rename) 핸들러를 제공해야 한다(shall). **When** 이름변경 요청이 기본 프로필 / 현재 활성 프로필 / 이미 존재하는 이름을 대상으로 감지되면, 서버는 해당 요청을 거부하고 사유를 배너로 표시해야 한다.

### §B.6 M6 — autonomy stub 결말

- **REQ-WCR-050**: 콘솔은 렌더된 HTML에 form/action 없는 맨 autonomy 링크 조각을 주입하지 **않아야 한다(shall not)**. autonomy tier의 영속화 대상이 config yaml에 존재하지 않으므로(plan.md §B D3 실측 — env reader 1건 + init 전용 writer만 존재), stub(`internal/web/autonomy.go`의 `handleAutonomyTiers` / `renderAutonomyToggle`)과 링크 주입부(`internal/web/handlers.go`의 `autonomyToggleLinkHTML` 상수 · `injectAutonomyToggleLink`)와 `/autonomy/tiers` 라우트는 제거되어야 한다(shall). 제거 사유는 plan.md §B D3에 기록되며, init 경로의 `ApplyAutonomyTierBundle`과 `MOAI_AUTONOMY_TIER` 환경변수 seam은 보존한다(콘솔 표면만 제거).

### §B.7 횡단 제약

- **REQ-WCR-060**: 신규로 도입되는 모든 사용자 대면 문자열은 `internal/web/assets/i18n.js`의 4개 로케일(en/ko/ja/zh) 전부에 키를 가져야 한다(shall). **When** i18n 키 집합 동등성 테스트가 실행되면, 예외(exemption) 0건으로 통과해야 한다.
- **REQ-WCR-061**: **When** `.moai/config/sections/*.yaml`에 대한 편집이 발생하면, 동일 커밋이 `internal/template/templates/` 하위의 대응 미러 편집과 `make build` 재생성 산출물을 함께 담아야 한다(shall).
- **REQ-WCR-062**: `internal/template/templates/` 하위로 배포되는 산출물은 SPEC ID, REQ/AC 토큰, 내부 작업 날짜, commit SHA를 포함하지 **않아야 한다(shall not)**.
- **REQ-WCR-063**: `internal/web` 패키지의 구문 커버리지는 90% 이상이어야 한다(shall).

---

## §C Out of Scope

본 SPEC이 **만들지 않는** 것. 각 항목은 의도적 배제이며 누락이 아니다.

### Out of Scope — 설정 저장소 스키마 변경

- yaml 키 삭제, struct 멤버 삭제, config 로더 경로 변경 — REQ-WCR-003의 보존 계약과 정면 충돌한다. 본 SPEC은 **웹 렌더 표면**만 변경한다.
- `settings.ApplySchemaEdits`의 영속화 라우팅(PersistSeam / PersistTypedSection) 재설계.

### Out of Scope — GLM 런타임 채널 신설

- 티어별 추론 강도를 실제로 전달하는 신규 런타임 채널(요청 단위 `reasoning_effort` 주입, per-agent 환경 분기)의 구현. REQ-WCR-033은 현재 제약을 **정직하게 표시**할 뿐 제약을 해소하지 않는다.
- z.ai가 `ANTHROPIC_REASONING_EFFORT`를 실제로 준수하는지에 대한 실증 검증. 기존 코드 주석의 UNVERIFIED 표기는 그대로 유지된다.

### Out of Scope — autonomy tier 런타임 배선

- `MOAI_AUTONOMY_TIER` 환경변수 seam이나 `ApplyAutonomyTierBundle` 경로 자체의 재설계. REQ-WCR-050은 stub의 **표면**만 결말짓는다.

### Out of Scope — 다른 미렌더 섹션

- `harness` / `ralph` / `feedback` / `observability` / `security` / `handoff` / `cache` / `quality_extras` 섹션의 전면 렌더 복구. M1은 `git_strategy`를 **최소** 복구 대상으로 한정하며, 나머지 섹션은 M2의 신규 탭이 요구하는 개별 필드 단위로만 등장한다. 전면 복구는 후속 SPEC 소관이다.

### Out of Scope — 콘솔 비설정 표면

- `/specs` 보드(`board.templ`), 실시간 대시보드, 디자인 토큰/CSS 재작업, 마스코트·히어로 배너 변경.

---

## §D 성공 기준

1. 9개 탭이 §B.2 순서대로 렌더되고, 각 탭이 §B.2가 지정한 필드 집합을 보유한다.
2. 죽은 설정 7개 필드가 렌더 표면에서 사라지고, 동일 이름의 yaml 키를 담은 기존 config가 오류 없이 로드된다.
3. bool 필드에 체크박스 위젯이 하나도 남지 않는다.
4. 닫힌 집합 필드에 자유 텍스트 입력이 남지 않는다(§B.3 REQ-WCR-023 예외 목록 제외).
5. 서드파티 LLM 탭이 실제 적용 티어를 명시한다.
6. `profileManager` 카드가 사라지고 프로필 CRUD가 프로필 바에서 가능하다.
7. autonomy stub의 맨 링크가 사라진다.
8. i18n 키 동등성 테스트가 예외 0건으로 통과하고, `internal/web` 커버리지가 90% 이상이다.

---

## §E 참조

- `internal/settings/schema_sections.go` — FieldDef SSOT
- `internal/web/schemaform.go` — `consoleTabs()` / `schemaSectionMetas()` / `parseSchemaForm`
- `internal/web/fieldsets.templ` — 제네릭 위젯 5종
- `internal/web/root.templ` — 탭 nav + 프로필 UI
- `internal/web/handlers.go` / `internal/web/autonomy.go` — autonomy stub
- `internal/web/glmkey.go` — GLM 키 보안 계약
- `internal/cli/glm.go` / `internal/template/glm_effort_overlay.go` — GLM 전달 채널
- `CLAUDE.local.md` §2(Template-First) / §6(커버리지) / §25(템플릿 중립성)
