---
id: SPEC-MODEL-MATRIX-UPDATE-001
title: "모델 매트릭스 갱신 — codex 감사 핀 신설, claude 감사 핀 승격, glm effort max 통일, 구형 모델 제공 집합 철수"
version: "0.1.0"
status: draft
created: 2026-09-30
updated: 2026-09-30
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/config + internal/cli + internal/settings + internal/template/templates"
lifecycle: spec-anchored
tags: "model-matrix, audit-pin, codex-backend, claude-opus-5-5, glm-effort-max, glm-5.3, statusline-context-window, card-t1368"
tier: M
---

# SPEC-MODEL-MATRIX-UPDATE-001 — 모델 매트릭스 갱신 (카드 t1368)

## HISTORY

| 날짜 | 변경 | 근거 |
|---|---|---|
| 2026-09-30 | 초판 작성 (plan-phase). 운영자 지시 6개 항목 + 유지 항목 1건을 본 트리에서 검증해 GEARS 요구사항으로 변환 | 카드 t1368 리드 디스패치 (운영자 지시 전사) |
| 2026-09-30 | 2차 개정 (plan-audit iter-1 FAIL 0.875 → wave-1): D2 AC 19→15 병합·재번호, D3 RED 테스트 2파일 추가, D4 i18n 4 로케일(+동급 effort 설명 표면), D5 무효 주석 2표면, D6 llm.yaml 환상 앵커 제거(verify-only 재분류), D8 related_specs 산문 이관 | `.moai/reports/t1368/plan-audit.md` |
| 2026-09-30 | 3차 개정 (wave-2, 운영자 게이트): NC-1/2/3 전건 확정 — plan.md 마커 절을 결정 기록(DR-1..3)으로 전환. NC-2는 권장안(철수 전용)을 기각한 **전면 삭제** 재정: REQ-MMU-004 재작성(삭제+레거시 폴백), Out of Scope 삭제 경계 재편, 소비자 실측 반영 | 운영자 게이트 2026-09-30 (리드 전달) |

## A. 배경

운영자가 모델 매트릭스 전반의 갱신을 지시했다: codex 감사 백엔드에 새 핀을 신설하고, claude 감사 핀을 최신 세대로 승격하며, `moai glm` effort 기본값을 max로 통일하고, 구형 GLM 모델(4.5~4.7대)이 tier 슬롯 제공 집합에 남아 있는지 확인해 정리한다. 아울러 statusline 컨텍스트 윈도우 테이블의 1M 값과 템플릿 `"model": "sonnet"` 별칭은 유지 불변으로 확정한다.

선행 SPEC-MODEL-MATRIX-* 4종(CORE/CONFIG/SURFACES/DOCS, phase v3.1.0)은 모두 `superseded` 상태로, 본 SPEC은 그 계열의 후속 갱신이며 어느 것도 되살리지 않는다. 본 SPEC은 REQ-AMP-005(Audit.Codex의 Go 기본값 EMPTY 유지 중립성)를 운영자 지시로 명시적으로 대체한다 — 이 대체 사실은 audit_models.go 주석에 기록 의무다.

계열 관계 산문: 위 4종 외에 SPEC-V3R6-AUDIT-MODEL-PIN-001(핀 우선순위·어휘 — REQ-AMP-005 소유, completed)과 SPEC-GLM-EFFORT-MAX-001(세션 기본 max, completed)이 선행 계열이다. (구 frontmatter `related_specs` 필드는 스키마 옵션 6필드 밖이라 plan-audit D8에 따라 이 산문으로 이관했다.)

## B. 요구사항 (GEARS)

### REQ-MMU-001 — codex 감사 백엔드 핀 신설 (Ubiquitous)

- **REQ-MMU-001**: The system shall pin the codex audit backend to `{gpt-6.1-sol, high}` across every audit entry path — the `codex_audit` MCP tool, the `audit_multi` codex route, and the moai-mcp codex audit surface — while never applying the pin to the task delegation paths (`codex_task`/`glm_task`).

시스템은 codex 감사 백엔드의 {model, effort} 핀을 `{gpt-6.1-sol, high}`로 신설한다. 이 핀은 감사 진입 경로 전체 — `codex_audit` MCP 도구, `audit_multi`의 codex 경로, plan-auditor/sync-auditor가 소비하는 moai-mcp codex 감사 표면 — 에 적용되며, 작업 위임 경로(`codex_task`/`glm_task`)에는 적용되지 않는다(REQ-AMP-008 불변 유지).

근거 표면(본 트리 검증):
- Go 기본값: `internal/config/defaults.go:1200-1211` `NewDefaultWorkflowConfig().Audit` — 현재 `Codex` 필드 미설정(zero). REQ-AMP-005 중립성 주석(`internal/config/audit_models.go:87`)은 대체 사실을 명기하며 갱신한다.
- 해석기 단말 폴백: `internal/cli/mcp_codex.go:201-213` `resolveCodexAuditModelEffort` — 프로젝트 핀 부재 시 현재 zero-value를 반환(`internal/cli/model_backend_default_test.go:36-37`가 이를 단정). 폴백이 새 핀으로 수렴하도록 한다.
- 서빙 가능 접두사 집합: `internal/cli/mcp_codex.go:156` `codexServableModelPrefixes = {"gpt-", "o1", "o3", "o4", "codex"}` — `gpt-6.1-sol`은 `gpt-` 접두사로 이미 통과. 이 집합의 변경은 불필요(검증 완료).
- 배포 템플릿: `internal/template/templates/.moai/config/sections/workflow.yaml:114-116` — 현재 `codex: model: "" / effort: ""`. 새 핀 값으로 채운다(Template-First).
- 로컬 도그푸드: `.moai/config/sections/workflow.yaml:25-27` — 현재 `gpt-5.6-sol/high`. 새 핀으로 갱신한다(미갱신 시 로컬 트리만 구형 모델에 머문다).
- 테스트 픽스처: `gpt-5.6-sol`을 쓰는 테스트 6파일(`internal/config/audit_models_test.go`, `internal/settings/audit_pin_fields_test.go`, `internal/cli/mcp_codex_audit_pin_test.go`, `internal/cli/model_backend_default_test.go`, `internal/cli/factory_operational_fixture_test.go`, `internal/cli/audit_pin_test.go`)을 새 핀 기준으로 갱신한다.

### REQ-MMU-002 — claude 감사 백엔드 핀 승격 (Ubiquitous)

- **REQ-MMU-002**: The system shall pin the claude audit backend to `{claude-opus-5-5, medium}` — replacing the previous `{sonnet, high}` pin — across the Go default, the resolver terminal default, the distributed template, the local dogfood config, and the documentation surfaces that state the default.

시스템은 claude 감사 백엔드의 핀을 `{sonnet, high}`에서 `{claude-opus-5-5, medium}`으로 변경한다. 운영자 지시의 "opus-5-5"는 Opus 5.5 세대(`claude-opus-5-5`)를 지칭하며, 감사 핀 어휘가 백엔드별 모델 id 형태(`gpt-5.6-sol`, `glm-5.3` 선례)이므로 전체 id 형태를 채택한다(값 형태는 plan.md 결정 기록 DR-1 확정).

변경 표면(본 트리 검증, 4+2 곳):
- Go 기본값: `internal/config/defaults.go:1202-1205` — `Model: "sonnet" / Effort: "high"` → `{claude-opus-5-5, medium}`.
- 해석기 단말 기본값: `internal/cli/mcp_claude.go:20-21` — `claudeAuditDefaultModel = "sonnet"` / `claudeAuditDefaultEffort = "high"` 상수 2개.
- 배포 템플릿: `internal/template/templates/.moai/config/sections/workflow.yaml:111-113` + 상단 주석(:108 "Claude ships at sonnet/high").
- 로컬 도그푸드: `.moai/config/sections/workflow.yaml:22-24`.
- 검증 문서: `.claude/rules/moai/core/moai-mcp-tools-catalogue.md:70`("defaults to `sonnet/high`") — 템플릿 미러(`internal/template/templates/.claude/rules/moai/core/moai-mcp-tools-catalogue.md`)와 같은 커밋에 동시 갱신(byte-parity 검사 `rule_template_mirror_test.go` 구속).
- web i18n: `internal/web/assets/i18n.js` — `f.workflow.audit.claude.model.desc` 4 로케일(en :586, ko :1481, ja :2248, zh :3015)의 "sonnet 기본" 문구 + `f.workflow.audit.claude.effort.desc` 4 로케일(en :588, ko :1483, ja :2250, zh :3017)의 "high 기본" 문구 — 모두 새 기본값을 서술하도록 갱신(plan-audit D4 + 감사 발견 누락분인 동급 effort 설명 표면 포함).
- effort 유효성: `internal/cli/mcp_claude.go:199-204` `validClaudeAuditEffort`는 `medium`을 이미 수용 — 집합 변경 불필요.

### REQ-MMU-003 — glm effort 기본값 max 통일 (Ubiquitous)

- **REQ-MMU-003**: The system shall default every GLM tier effort (high/medium/low/fable) to `max`, a value compatible with both glm-5.3 and glm-5.3-flash (flash accepts reasoning_effort max only).

`moai glm`의 티어별 effort 기본값은 4티어(high/medium/low/fable) 모두 `max`로 통일된다. glm-5.3과 glm-5.3-flash 양쪽 모두 max와 양립한다 — flash는 max 전용 모델이며(defaults.go:245-248 주석, 본 트리 검증), 세션 기본값은 이미 max다(`internal/template/glm_effort_overlay.go:208-210` `SessionGLMReasoningState() → glmReasoningMax`).

변경 표면(본 트리 검증):
- web 위젯 티어 기본값: `internal/settings/schema_sections.go:174-183` `glmDefaultTierEffort` — 현재 high/medium→`high`, low→`low`, fable→`max`. 전 티어 `GLMStateMax`로 통일하고 근거 주석을 갱신한다.
- 템플릿 llm.yaml: `internal/template/templates/.moai/config/sections/llm.yaml:55-95`의 effort 블록은 Claude-effort→z.ai 상태 collapse map으로, tier 기본값 서술을 담지 않는다(plan-audit D6 재확인) — 내용 변화 없이 유효성만 확인한다(verify-only). 실제 티어 기본값 서술은 `internal/settings/schema_sections.go:169-172` 주석이 소유하며 M2 항목 1이 커버한다.
- 선택 옵션 집합: `GLMReasoningStateNames()`(= [max, high, low])은 유지 — 저장 상태로 high/low 선택은 여전히 유효하며, 바뀌는 것은 기본 선택값뿐이다.
- 설정 파일 키: llm.yaml 템플릿에는 실체 `effort:` 키가 없고(주석 전용), Go `NewDefaultLLMConfig`도 GLM.Effort를 설정하지 않는다 — 신설하지 않는다(현재 구조 유지, plan.md §F M2 참조).

### REQ-MMU-004 — 구형 GLM 모델의 전면 삭제와 레거시 값 폴백 (State-driven)

- **REQ-MMU-004**: **While** glm-5.3 and glm-5.3-flash are the default models of every tier slot, the system shall offer exactly these two models in the tier-slot model closed set (`config.ValidGLMModels()`), shall fully delete the legacy old-model surfaces — the named constants (glm-4.5/4.5-air/4.6/4.7/5/5.1/5.2/5-turbo), the statusline context-window entries, and the legacy struct alias fields — and shall resolve a stored config value naming a removed model to the tier's default model with a one-line warning (fail-open: never a silent pass-through, never a hard error).

While glm-5.3과 glm-5.3-flash가 모든 tier 슬롯의 기본 모델인 동안, 시스템은 tier 슬롯 제공 집합을 이 두 모델만으로 제한하고, 구형 모델 표면 — 이름 상수(glm-4.5/4.5-air/4.6/4.7/5/5.1/5.2/5-turbo), statusline 컨텍스트 윈도우 항목, legacy struct 별칭 필드 — 를 전부 삭제하며, 삭제된 모델을 지명하는 저장 설정값은 티어 기본 모델로 폴백 + 1행 경고로 처리한다(fail-open — 침묵 통과도 하드 오류도 아님).

운영자 결정(2026-09-30 gate): 권장안이던 glm-5.2 선례식 철수 전용을 기각하고 **전면 삭제**를 재정했다 — 호환 보존의 위험을 운영자가 수용하되 침묵시키지 않고 관리하며, 폴백+경고가 그 관리 장치다(plan.md 결정 기록 DR-2). 삭제 범위의 소비자 실측(본 트리):
1. **제공 집합 (축소)**: `internal/config/closed_sets.go:82-84` — 5멤버 {5.3-flash, 5.3, 5.1, 4.7, 4.5-air} → 2멤버. glm-5.2 철수 선례(`:76-81`)의 주석 문체는 철수 기록에 준용하되, 상수 보존 절은 삭제로 대체된다.
2. **이름 상수 (삭제)**: `internal/config/defaults.go:255-261` `DefaultGLM45/46/47/45Air/51/52/5Turbo` + 보존 근거 주석(`:249-254`). 소비자 실측: `closed_sets.go`(2적중), `config/defaults_test.go`(7), `cli/glm_persist_gate_test.go`(2), `web/schema_select_preserve_test.go`(13), `template/glm_effort_overlay_test.go` — 전부 갱신 대상.
3. **statusline 항목 (삭제)**: `internal/statusline/memory.go:36-42` 구형 모델 7항목 + `:96` "glm-4.5 masking glm-4.5-air" 주석. `:34-35`(두 5.3 모델, 1M)는 REQ-MMU-005로 유지.
4. **legacy struct 별칭 필드 (삭제)**: `internal/config/types.go:361-364`(Opus/Sonnet/Haiku) + `defaults.go:971-974`(기본값 배정) + 소비자 `internal/cli/glm.go:746-753`·`:847`(레거시 별칭 폴백 사슬) + `internal/settings/schema_sections.go:200-202` REQ-WC12-006 "무접촉 보존" 주석. 레거시 별칭 **키**(yaml의 `opus:` 등)는 비엄격 로더의 기존 의미대로 무오류 무시된다(§D 삭제 경계 참조).

폴백 동작의 근거는 plan.md 결정 기록 DR-2에, 검증은 AC-MMU-012에 있다.

### REQ-MMU-005 — statusline 컨텍스트 윈도우 1M 불변 (Ubiquitous, 검증 전용)

- **REQ-MMU-005**: The system shall keep both glm-5.3 and glm-5.3-flash at 1,000,000 tokens in the statusline glmContextWindows table, unchanged by this SPEC (verify-only invariant).

시스템은 statusline의 glmContextWindows 테이블에서 glm-5.3과 glm-5.3-flash 모두 1,000,000 토큰을 유지한다. 본 트리 검증 완료: `internal/statusline/memory.go:34-35` — `"glm-5.3-flash": 1_000_000`(명시적 분기 방어 항목), `"glm-5.3": 1_000_000`. 변경 없음. run-phase는 이 값을 재확인하고 증거를 남긴다.

### REQ-MMU-006 — 템플릿 sonnet 별칭 유지 (Ubiquitous, 보존 불변)

- **REQ-MMU-006**: The system shall keep the template settings `"model": "sonnet"` alias unchanged, so coding lanes auto-benefit from the sonnet-5.5 upgrade (preserved invariant, operator-evidenced).

템플릿 settings의 `"model": "sonnet"` 별칭은 변경하지 않는다. 본 트리 검증 완료: `internal/template/templates/.claude/settings.json.tmpl:417`. 코딩 레인이 sonnet-5.5 승격을 자동 수혜한다(운영자 Terminal-Bench 4.0 70.6% 근거). preserved-invariant로 기록하며, run-phase가 이 줄의 불변을 재확인한다.

### REQ-MMU-007 — 감사 핀 갱신의 문서·테스트 동반 (Ubiquitous)

- **REQ-MMU-007**: **While** any pin or default named by this SPEC is being changed, the system shall update in the same change set every test and document that asserts or states the changed value — a pin change landing without its test update is a defect.

While 본 SPEC의 어떤 핀·기본값이 변경되는 동안, 해당 값을 단정하거나 서술하는 테스트와 문서는 같은 변경 집합 안에서 함께 갱신된다. 핀 변경이 테스트 갱신 없이 착지하는 것은 결함이다. 최소 표면: REQ-MMU-001/002에 열거한 테스트 6파일, `internal/config/mcp_audit_config_test.go`(로컬 workflow.yaml sonnet/high 단정), `internal/settings/schema_sections_test.go`(티어 effort 기본값), `internal/web/glm_tier_test.go`와 `internal/config/defaults_test.go`(ValidGLMModels 5멤버 집합 하드코드 — plan-audit D3), 구형 상수 소비 테스트(`internal/web/schema_select_preserve_test.go` 13적중, `internal/cli/glm_persist_gate_test.go`, `internal/template/glm_effort_overlay_test.go` — DR-2), REQ-MMU-002의 문서·i18n 표면(4 로케일), 무효화되는 주석 표면(`internal/config/defaults.go:249-254` 제공 집합 서술, 로컬 workflow.yaml:20 "Claude defaults to sonnet/high" — plan-audit D5, `internal/settings/schema_sections.go:200-202` REQ-WC12-006 보존 서술, `internal/statusline/memory.go:96` 마스킹 주석 — DR-2).

### REQ-MMU-008 — Template-First 착지 (Ubiquitous)

- **REQ-MMU-008**: The system shall land every template-surface change in `internal/template/templates/` first and regenerate the embedded copies via `make build`, updating each local mirror in the same change set.

템플릿 표면의 모든 변경은 `internal/template/templates/`에서 먼저 이뤄지고 `make build`로 임베드에 반영된다. 로컬 미러(`.moai/config/sections/`, `.claude/rules/`)는 템플릿 원본과 같은 변경 집합에서 갱신된다.

## C. 채택 결정 (운영자 지시의 충실 전사)

1. REQ-MMU-001의 {gpt-6.1-sol, high}는 REQ-V3R6-AUDIT-MODEL-PIN-001 REQ-AMP-005("Go 기본값 EMPTY 유지")를 운영자 지시로 대체한다. 대체 사실은 `internal/config/audit_models.go`의 해당 주석에 기록한다.
2. REQ-MMU-002의 effort `medium`은 Opus 5.5 세대의 기본 effort이기도 하다(model-policy.md:21 "default effort `medium`") — 값 자체는 운영자 지시 그대로다.
3. REQ-MMU-004의 정리 범위는 운영자 게이트(2026-09-30)에서 **전면 삭제로 재정**됐다 — 권장안이던 glm-5.2 선례식 철수 전용 기각(plan.md 결정 기록 DR-2). 수용된 위험은 폴백+1행 경고로 관리한다.
4. REQ-MMU-005/006은 검증·보존 항목으로, 코드 변경을 수반하지 않는다.

## D. Out of Scope — 본 SPEC이 만지지 않는 것

### Out of Scope — 템플릿 sonnet 별칭 및 agent 모델 정책

- `internal/template/templates/.claude/settings.json.tmpl:417`의 `"model": "sonnet"` — REQ-MMU-006으로 유지 불변.
- `.claude/agents/moai/**`와 `model-policy.md`의 agent 모델 상속 규칙(에이전트 무선언·상속) — 감사 핀과 별개 축.
- plan-auditor/sync-auditor 에이전트 파일 — model/effort 무선언이 확정돼 있어 변경 대상 아니다.

### Out of Scope — 삭제 경계와 수용된 위험의 관리 방식

- glm-5.3·glm-5.3-flash와 `DefaultGLMHaiku/Sonnet/Opus`(`internal/config/defaults.go:262-265`, 값이 glm-5.3-flash인 티어 별칭 상수)는 삭제 대상이 아니다 — 지워지는 것은 구형 "모델 id" 표면뿐이다.
- `llm.glm.context_windows` 사용자 오버라이드 키(statusline)는 유지 — 사용자가 직접 창을 지정하는 경로는 구형 id 지명에서도 동작한다.
- 레거시 별칭 **키**(yaml `opus:`/`sonnet:`/`haiku:`)의 마이그레이션 도구나 오류 전환은 만들지 않는다 — 비엄격 로더의 기존 의미(무오류 무시)가 그대로 수용된 위험이다(plan.md 결정 기록 DR-2).
- 삭제된 모델을 위한 설정 자동 재작성 도구는 만들지 않는다 — 폴백+1행 경고가 관리 장치의 전부다.

### Out of Scope — 감사 게이트·작업 위임 경로·유효성 집합

- `workflow.audit.gates.*`(claude required / codex required / glm advisory) — 게이트 토큰 불변.
- `codex_task`/`glm_task` 작업 위임 경로의 모델 해석 — REQ-AMP-008(감사 전용 핀) 불변.
- `codexServableModelPrefixes` 집합 — `gpt-6.1-sol`이 이미 통과(검증 완료), 변경 불필요.
- `ValidAuditModels()`/`ValidAuditGates()`/`validClaudeAuditEffort()` 닫힌 집합 — 새 값이 모두 기존 집합에 수용된다.
- moai web 감사 codex 모델 필드의 닫힌 집합화 — 현재 TypeText 자유 입력이며(검증: `internal/settings/schema_sections.go:388`), gpt-6.1-sol은 이미 타이핑 가능하다. 셀렉트 전환은 신규 기능으로 본 SPEC 범위 밖(plan.md 결정 기록 DR-3 확정).

### Out of Scope — 통합·배포

- 직렬 통합 창을 통한 develop 병합과 push — 레인 소관 밖(리드 일괄).
- `moai web` 런타임 동작의 실측(브라우저 기반 확인) — 스키마·소스 수준 검증으로 충분하다.

## E. 영향 표면 요약

| 표면 | 파일 | 변경 |
|---|---|---|
| Go 감사 기본값 + 구형 상수 삭제 | internal/config/defaults.go | REQ-MMU-001/002/004 |
| 감사 어휘 주석 | internal/config/audit_models.go | REQ-AMP-005 대체 기록, sonnet/high 서술 갱신 |
| codex 해석기 폴백 | internal/cli/mcp_codex.go | REQ-MMU-001 |
| claude 해석기 기본 상수 | internal/cli/mcp_claude.go | REQ-MMU-002 |
| GLM 제공 집합 | internal/config/closed_sets.go | REQ-MMU-004 |
| legacy struct 필드·소비자 | internal/config/types.go + internal/cli/glm.go | REQ-MMU-004 |
| web 티어 effort 기본값 + WC12-006 주석 | internal/settings/schema_sections.go | REQ-MMU-003/004 |
| web i18n 감사 설명 | internal/web/assets/i18n.js | REQ-MMU-002/007 |
| statusline 구형 항목 삭제 + 1M 불변 | internal/statusline/memory.go | REQ-MMU-004/005 |
| 템플릿 workflow.yaml | internal/template/templates/.moai/config/sections/workflow.yaml | REQ-MMU-001/002/008 |
| 템플릿 llm.yaml | internal/template/templates/.moai/config/sections/llm.yaml | REQ-MMU-003 (verify-only) |
| 룰 카탈로그 + 미러 | .claude/rules/moai/core/moai-mcp-tools-catalogue.md (+templates 미러) | REQ-MMU-002/007/008 |
| 로컬 도그푸드 workflow.yaml | .moai/config/sections/workflow.yaml | REQ-MMU-001/002 |
| 테스트 | REQ-MMU-001의 6파일 + mcp_audit_config_test.go + schema_sections_test.go + glm_tier_test.go + defaults_test.go + schema_select_preserve_test.go + glm_persist_gate_test.go + glm_effort_overlay_test.go | REQ-MMU-007 |
| 불변 | settings.json.tmpl | REQ-MMU-006 (변경 없음) |
