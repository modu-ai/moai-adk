---
id: SPEC-WEB-CONSOLE-012
title: "moai web Stale/Dead Config Surface Cleanup (Track 1)"
version: "0.1.0"
status: draft
created: 2026-07-10
updated: 2026-07-10
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "internal/settings, internal/web"
lifecycle: spec-anchored
tags: "web, console, settings-schema, dead-code, cleanup, llm, glm-fable, research-seam, git-convention"
tier: S
era: V3R6
related_specs: [SPEC-WEB-CONSOLE-010, SPEC-WEB-CONSOLE-011]
---

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-07-10 | manager-spec | 최초 draft. 2026-07-10 orchestrator 감사(Track 1: stale/dead config surface) 5개 발견(A1-A5)을 plan-phase 실측 재검증 후 요구사항화. A4c(bindForm statusline residual)는 실측 no-op 판정 — Out of Scope로 격하. A5 5개 필드는 도구 증거로 USED 2 / DEAD 3 분류 완료 (acceptance.md §D.2). |

---

## §A Context & Motivation

`moai web` 콘솔의 설정 스키마 SSOT(`internal/settings/schema.go` + `schema_sections.go`)와 그 주변 라우팅/문서 표면에 2026-07-10 orchestrator 감사가 5개 stale/dead 결함을 확인했다. 본 SPEC은 그 Track 1 정리분이다. 모든 파일:라인 앵커는 2026-07-10 plan-phase에서 실측 재검증되었다 (plan.md §C에 검증 명령 기록; 라인 번호는 드리프트 가능 — run-phase는 content-token 기준 재확인).

**A1 — Ghost GLM tier 필드.** `internal/settings/schema_sections.go` `llmFields()`가 `glm.models.{high, medium, low, opus, sonnet, haiku}` 6개 tier를 노출한다. 그러나 런타임 소비자(`internal/cli/glm.go` `setGLMEnv`, :196-199)는 `Models.{High, Medium, Low, Fable}`만 읽는다 — `ANTHROPIC_DEFAULT_OPUS_MODEL`조차 `Models.High`에서 온다. struct의 `Opus/Sonnet/Haiku` 필드(`internal/config/types.go` GLMModels)는 "Legacy fields for backward compatibility" 주석의 yaml 하위호환 전용이며 행동 reader가 없다(실측: 쓰기 지점만 존재 — `sectionapply.go:203-207` 웹 apply, `glm.go:697-704` defaults backfill). 반면 실제 소비되는 `glm.models.fable`(`ANTHROPIC_DEFAULT_FABLE_MODEL`)은 노출되지 않는다. → ghost 3필드 제거 + fable 노출.

**A2 — Dead `research` seam 쓰기 경로.** `research`는 `SectionResearch`(schema.go:40)로 선언되고 `AllSections()`(schema.go:67)에 등재되며, `RouteSeam` 매핑(sectionroute.go:62) + `SeamSections()`(sectionroute.go:81) + seam 쓰기 whitelist(sectionwrite.go:32)에 배선되어 있다. 그러나 FieldDef 0개, 콘솔 탭(`schemaSectionMetas`, web/schemaform.go:65) 미등재 — 렌더도 편집도 불가능한 유령 쓰기 경로다. `internal/research` 패키지는 삭제됨(실측: 디렉터리 부재; 선행 감사에서 research config 12/12 dead 판정). → research를 seam 라우팅/whitelist/섹션 enum에서 전면 제거 (미등재 → `RouteExcluded`, db 선례와 동일 패턴).

**A3 — Stale doc comment.** `internal/web/server.go`(패키지 doc, :13-25) 및 `internal/web/projectconfig.go`(@MX:REASON, :160 인근)가 여전히 `db`를 편집 가능 10섹션과 8-seam 목록에 열거한다. 실제로는 db가 콘솔 표면에서 제거되어 미등재 → `RouteExcluded`다(sectionroute.go:56-58 주석 실측). A2 완료 후에는 research도 같은 열거에서 빠져야 한다. → 두 doc comment를 최종 scope 계약으로 재작성.

**A4 — Dead code 잔재.** (a) `internal/web/assets.go:27-34`의 `errDictKey` keep-alive sentinel — retired html/template `dict` helper의 오류 sentinel(`validate.go:13-15`)을 `unused` linter로부터 지키는 blank reference. 당시 근거였던 "validate.go BYTE-UNCHANGED"(REQ-WC6-004/AC-WC5-010a)는 완결된 선행 SPEC의 시점 제약이며, 상시 byte-guard 테스트 실존 여부는 run-phase 검증 게이트로 둔다(§B.4 Where-gate). (b) `schema_sections.go:62` `WorkflowAgentPurposes()` — 전 리포(테스트 포함) 호출자 0 실측. 함수 주석은 "canonical taxonomy 참조로 유지"를 주장하나(Chesterton's fence 검토됨), taxonomy의 canonical SSOT는 dynamic-workflows.md 표 + `config.Workflow.WorkflowAgents` 맵 키이지 zero-caller Go 함수가 아니다. → 제거.

**A5 — 의심-dead 노출 필드 (도구 증거로 분류 완료).** verification-claim-integrity.md §1.1 surface 3에 따라 5개 필드 각각의 런타임 reader를 grep 실측했다 (증거 verbatim: acceptance.md §D.2):

| 필드 | 분류 | 근거 (non-test, non-web 행동 reader) |
|------|------|--------------------------------------|
| `quality.tdd_settings.min_coverage_per_commit` | **USED** | `internal/core/quality/trust.go:788` `g.config.TDDSettings.MinCoveragePerCommit` |
| `git_convention.validation.enforce_on_push` | **USED** | `internal/cli/hook_pre_push.go:251-258` (env override 후 config 필드 읽기) — 선행 감사의 dead 반증과 일치 |
| `git_convention.auto_detection.enabled` | **DEAD** | 행동 reader 0 (config 인프라의 validation/defaults/loader-key 인식만 존재) |
| `git_convention.auto_detection.confidence_threshold` | **DEAD** | 상동 |
| `git_convention.auto_detection.sample_size` | **DEAD** | 상동 |

→ DEAD 3필드만 웹 스키마(FieldDef, schema.go:440-458)에서 제거; USED 2필드는 잔류. struct/yaml 로드/defaults/validation은 전부 보존(011 M4 다이어트 선례 패턴).

---

## §B Requirements (GEARS)

### §B.1 M2 — A1: LLM ghost tier 제거 + fable 노출

**REQ-WC12-001 (Ubiquitous):** The web settings schema (`llmFields()`) shall expose exactly the four GLM model tiers `glm.models.{high, medium, low, fable}` as editable fields.

**REQ-WC12-002 (Ubiquitous — shall not):** The web settings schema shall not expose `glm.models.{opus, sonnet, haiku}` — legacy backward-compat struct aliases with zero runtime behavior readers.

**REQ-WC12-003 (When):** When a console save submits `llm.glm.models.fable`, the typed apply path (`applyLLMKey`) shall persist the value to `LLMConfig.GLM.Models.Fable`, and shall no longer carry apply branches for the three removed ghost keys.

**REQ-WC12-004 (Ubiquitous):** The interface-i18n dictionary (`internal/web/assets/i18n.js`) shall carry `f.llm.glm.models.fable.{title,desc}` entries in every locale block that carries the sibling `f.llm.glm.models.high.{title,desc}` entries, and shall not carry orphan `f.llm.glm.models.{opus,sonnet,haiku}.*` entries.

**REQ-WC12-005 (Ubiquitous):** The LLM section display meta (`schemaSectionMetas()`, internal/web/schemaform.go) shall describe the real tier set (high/medium/low/fable), not the retired six-tier enumeration.

### §B.2 M1 — A2: research seam 쓰기 경로 폐선

**REQ-WC12-010 (Ubiquitous):** The section routing SSOT shall classify `research` as `RouteExcluded` by non-registration — removed from the `sectionRoutes` map, from `SeamSections()`, and from the `sectionRootKeys` seam write whitelist.

**REQ-WC12-011 (Ubiquitous — shall not):** The settings schema shall not declare a `SectionResearch` section ID nor list it in `AllSections()`; surrounding comments enumerating "7개 seam 섹션" shall be corrected to the remaining six.

**REQ-WC12-012 (When):** When a write to section `research` is attempted via `WriteSectionViaSeam`, the seam shall reject it with the not-seam-writable error and shall not touch `research.yaml` — the same guard behavior already exercised for `db` and other unregistered sections.

### §B.3 M3 — A5: dead 노출 필드 제거 (USED 필드 잔류)

**REQ-WC12-020 (Ubiquitous — shall not):** The web settings schema shall not expose `git_convention.auto_detection.{enabled, confidence_threshold, sample_size}` (FieldDefs at schema.go:440-458) — classified DEAD by the per-field runtime-reader evidence in acceptance.md §D.2.

**REQ-WC12-021 (Ubiquitous):** The web settings schema shall retain `quality.tdd_settings.min_coverage_per_commit` and `git_convention.validation.enforce_on_push` — classified USED by the same evidence protocol.

**REQ-WC12-022 (Ubiquitous):** The removal shall be schema-surface-only: `AutoDetectionConfig` struct members, yaml load, `internal/config/defaults.go` defaults, and `internal/config/validation.go` range checks shall be preserved (backward compat — the SPEC-WEB-CONSOLE-011 M4 diet precedent pattern).

**REQ-WC12-023 (Ubiquitous — shall not):** The cleanup shall not touch the distinct `harness.auto_detection.enabled` field (schema_sections.go:215) — a live harness-section field that shares only the key substring.

### §B.4 M4 — A4: dead code 잔재 제거

**REQ-WC12-030 (Where):** Where the run-phase pre-flight confirms that no standing test asserts byte-content or hash of `internal/web/validate.go` (the historical REQ-WC6-004 constraint being a closed-SPEC point-in-time assertion), the `errDictKey` sentinel (validate.go:13-15) and its keep-alive blank reference (assets.go:27-34) shall both be removed.

**REQ-WC12-031 (When):** When the pre-flight instead finds a standing byte-content guard on validate.go, the implementation shall retain the sentinel, update the assets.go keep-alive comment to cite the found guard by file:test-name, and report the constraint in the completion report (no silent skip).

**REQ-WC12-032 (Ubiquitous):** The zero-caller function `WorkflowAgentPurposes()` (schema_sections.go:62) shall be removed after a run-phase re-grep confirms the caller count is still zero; the 7-purpose taxonomy remains documented by its canonical SSOT (dynamic-workflows.md + `config.Workflow.WorkflowAgents` map keys).

### §B.5 M5 — A3: stale doc comment 정정

**REQ-WC12-040 (Ubiquitous):** The `internal/web/server.go` package doc and the `internal/web/projectconfig.go` @MX:REASON block shall enumerate the post-cleanup scope contract — typed sections (git-strategy, llm), the six seam sections (workflow, harness, ralph, feedback, observability, security), and shall name `db` and `research` among the excluded set instead of the editable set.

### §B.6 Cross-cutting guards

**REQ-WC12-050 (When):** When any schema FieldDef change lands, both the web-side parity tests (`internal/web` + `internal/settings` schema/i18n tests) and the TUI-side `internal/cli/schema_bridge_test.go` shall be updated together and pass — the SPEC-WEB-CONSOLE-011 M2b regression (web-only update broke 3 TUI tests) shall not recur.

**REQ-WC12-051 (Ubiquitous — shall not):** The implementation shall not modify any file under `internal/statusline/` (owned by a parallel session) nor any file under `internal/template/templates/` (the schema lives in Go code; the template llm.yaml already carries the correct key set).

---

## §C Scope Exclusions

The following are explicitly out of scope for this SPEC.

### Out of Scope — struct/yaml 하위호환 표면

- `GLMModels`의 legacy `Opus/Sonnet/Haiku` struct 필드, `glm.go:697-704` defaults backfill, `AutoDetectionConfig` struct/defaults/validation — 전부 보존 (yaml 하위호환; REQ-WC12-022).
- 라이브 `.moai/config/sections/llm.yaml`에 이미 기록된 ghost 키 값 정리 — 런타임 config 파일 무접촉 (struct가 legacy 필드를 유지하므로 typed re-marshal이 기존 키를 파괴하지 않음).

### Out of Scope — A4c bindForm statusline residual

- plan-phase 실측 결과 no-op: `internal/web/handlers.go` `bindForm`(:449-471) 본문에 statusline 바인딩 분기가 이미 존재하지 않으며, doc comment도 정확하다(M3 redesign 반영). `ProfilePreferences`의 statusline struct 필드는 TUI/CLI + statusline.yaml sync 소비자가 있어 무접촉.

### Out of Scope — git_convention 엔진 배선

- `auto_detection` 필드의 미래 엔진 배선(SPEC-WEB-CONSOLE-009 후보 영역, GCR-5 maintainer 결정 종속)은 본 SPEC 소관이 아니다. 본 SPEC은 웹 스키마 노출만 제거하며 struct/yaml이 보존되므로 미래 배선을 차단하지 않는다.

### Out of Scope — 병렬 세션 소유 파일 및 템플릿 미러

- `internal/statusline/*` (병렬 세션이 수정 중 — renderer.go, cache_hit_test.go).
- `internal/template/templates/**` 전체 (llm.yaml 템플릿은 이미 정답 키셋 — 변경 불요 실측).

---

## §D Acceptance Criteria

정식 AC 매트릭스(기계 검증 명령 + 기대 출력 + A5 per-field 증거 verbatim)는 `acceptance.md`가 SSOT다. 요약: AC-WC12-001..017, Given-When-Then 시나리오 3건, edge case 3건.
