---
id: SPEC-WEB-CONSOLE-013
title: "moai web New Config Domain Exposure — Model Policy / Handoff / Cache (Track 2)"
version: "0.1.1"
status: draft
created: 2026-07-10
updated: 2026-07-10
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/web"
lifecycle: spec-anchored
tags: "web, console, config-exposure, model-policy, handoff, cache, seam, yamlpatch, i18n"
tier: M
era: V3R6
depends_on: [SPEC-WEB-CONSOLE-012]
related_specs: [SPEC-WEB-CONSOLE-011, SPEC-HANDOFF-AUTORESUME-001, SPEC-V3R6-PROMPT-CACHE-001, SPEC-TOKEN-ROUTING-001, SPEC-AGENT-ARCH-V2-001]
---

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.1 | 2026-07-10 | manager-spec | plan-audit iter-1 (PASS 0.91, skip-eligible) SHOULD-FIX D1-D7 반영: D1 db 귀속 정정(REQ-WC11-019 계보 + 콘솔 표면 제거 — REQ-WC11-018 잔여군 아님; spec/acceptance 양쪽), D2 AC-WC13-015 함수 스코프 awk + 주석 제외로 재작성(서술 주석 false-fail 제거), D3 ApplyPerformanceTier 앵커 :416 정정, D4 012 상태 갱신(authored draft Tier S — completed까지 게이트 유지), D5 AC-WC13-002 셀 깨진 프래그먼트 정리, D6 plan.md M1 주석 스윕 목록 확장(sectionroute.go "7개 섹션" 리터럴 + cache 열거 주석), D7 REQ-WC13-024 EmptyLabelKey 표현 정밀화(FieldDef-less 뷰 설계와 정합). |
| 0.1.0 | 2026-07-10 | manager-spec | 최초 draft. v2.14.0 이후 신설 config 도메인 3종(Model Policy / handoff / cache)의 웹 콘솔 노출. 2026-07-10 orchestrator 검증 findings 기반 + plan-phase 실측 정정 4건 반영 (§1.1). Track 1(SPEC-WEB-CONSOLE-012) 선행 의존. |

---

## §1 Context & Motivation

`moai web` 콘솔의 config 커버리지는 SPEC-WEB-CONSOLE-011이 확립한 9-섹션 편집 계약(`internal/settings/sectionroute.go` SSOT)에 고정되어 있다. v2.14.0 이후 신설된 3개 config 도메인이 콘솔에서 접근 불가하다:

1. **Model Policy** — `workflow.model_routing_profiles`(No-Haiku 3-tier: max/medium/low perfTier × S/M/L×plan/run/sync/mx 12셀, `internal/config/model_routing.go`) + `llm.performance_tier`(활성 tier, `moai init --model-policy`가 기록). 콘솔에 어떤 표면도 없다.
2. **Handoff auto-resume** — `handoff.yaml`의 `handoff.mode`(manual|auto) + `handoff.guide`(bool). `sectionroute.go` 미등재 → zero-value `RouteExcluded`. 런타임 reader는 `internal/hook/handoff_inject.go:146-159`(SessionStart 주입 게이트)로 실존한다. 로컬 dogfood는 `mode: auto`.
3. **Prompt caching** — `cache.yaml`의 `cacheStrategy.enabled`(bool) + `cacheStrategy.session_ttl`(1h|5m|off). 현재 `ExcludedSections()`의 machine/state 그룹에 **명시 등재**되어 있다(REQ-WC11-018) — 본 SPEC이 cache에 한해 부분 SUPERSEDE한다 (§2.1 REQ-WC13-001).

본 SPEC은 기존 영속화 seam(RouteTypedSave / RouteSeam yamlpatch comment-preserving / `sectionroute.go` SSOT)을 재사용하며 새 영속화 경계를 만들지 않는다.

### §1.1 위임 브리프 대비 실측 정정 (2026-07-10 plan-phase grep 실측)

| # | 브리프 주장 | 실측 결과 | 본 SPEC 반영 |
|---|------------|----------|--------------|
| 1 | "14 workflow_agents 필드가 agent-settings 섹션에 존재" | **STALE** — M5-a B1이 workflow_agents를 웹 렌더에서 제거함 (`internal/settings/schema_sections.go:262`; struct/yaml 키는 보존, dynamic-workflow JS가 yaml 직접 읽음). 현 agent-settings는 role_profiles 7×4=28 필드만 | workflow_agents는 hidden 유지 (REQ-WC13-023) — M5-a B1 결정 존중, 재노출은 별도 사용자 결정 |
| 2 | "`MOAI_MODEL_POLICY` env가 profile 선택" | **미실존** — `internal/config/envkeys.go`에 MODEL_POLICY 항목 없음. 선택자는 `moai init --model-policy max\|medium\|low` 플래그 → `template.ApplyPerformanceTier`가 `llm.yaml performance_tier`에 영속 (`internal/cli/init.go:94`, `internal/template/model_policy.go:416` ApplyPerformanceTier) | 활성 perfTier 표시는 `llm.performance_tier` READ-ONLY 표시로 설계 (REQ-WC13-020) |
| 3 | "model_routing_profiles는 model_routing.go + resolver.go가 backing" | resolver.go에 perfTier/model_routing_profiles 참조 0건 — backing은 model_routing.go + types.go + loader | 문서 앵커 정정 |
| 4 | "cache.yaml — enabled + session_ttl 2키" | 파일에는 4키 존재 (`spec_ttl`, `min_cacheable_tokens` 추가) | 노출 스코프는 브리프대로 2키; 나머지 2키는 seam unmodeled-key 보존으로 무손상 (REQ-WC13-006/015) |

### §1.2 편집 가능성 증거 요약 (verification-claim-integrity §1.1 surface 3)

키별 editable 판정은 "런타임 reader가 편집을 소비하는가"의 grep 실측(비테스트 코드 한정)에 근거한다. 전체 증거 표는 acceptance.md §D.1이 SSOT다. 핵심:

- **handoff.mode / guide → editable.** `internal/hook/handoff_inject.go:154,159`가 `cfg.Handoff.Mode/Guide`를 SessionStart마다 읽는다 (강한 런타임 소비).
- **cacheStrategy.enabled / session_ttl → editable (caveat 기록).** `LoadCacheConfig`의 유일한 비테스트 호출자는 `internal/cli/doctor_cache.go:48`(doctor 지표 경로). `InjectCacheControl`(실제 cache_control 주입기, `internal/runtime/cache_control.go:97`)은 **비테스트 호출자 0건** — 편집의 현재 유효 반경은 doctor 표시에 국한. 사용자 opt-in 표면이라는 도메인 의도(PROMPT-CACHE-001) + 실존 reader로 editable 판정; caveat는 UI 힌트 없이 acceptance.md 증거 표에만 기록.
- **model_routing_profiles 12×3셀 + llm.performance_tier → READ-ONLY.** 접근자 `RouteModelFor`는 internal/config 외부 비테스트 호출자 0건 (SPEC-TOKEN-ROUTING-001 D1 wiring 부채). 편집을 소비하는 런타임 reader 부재 → 증거 규칙상 편집 금지, 표시만.
- **legacy `workflow.model_routing` flat 블록 → 완전 숨김.** DEPRECATED alias; internal/config 밖 소비자 0건. read-only 표시조차 노이즈.

### §1.3 기존 인프라 (PRESERVE + EXTEND)

- PRESERVE: `internal/settings/yamlpatch` seam(주석/미모델링 키 보존 노드 수술), `WriteSectionViaSeam`의 sectionRootKeys 최상위 키 가드, loopback+Host-check 쓰기 안전 모델(REQ-WC-009 CSRF 금지 유지), TUI 브리지의 persist-kind 제외 술어(`internal/cli/schema_bridge_test.go:33`).
- EXTEND: `sectionRoutes` 맵 + `SeamSections()` + `ExcludedSections()`, `sectionRootKeys`, `SectionID` enum + `AllSections()`, generic schema fieldset 렌더 경로, `internal/web/assets/i18n.js` 4-locale 사전.

---

## §2 Requirements (GEARS)

### §2.1 M1 — 라우팅 기반 (scope 부분 supersede + seam 라우팅)

**REQ-WC13-001 (Ubiquitous):** This SPEC shall partially supersede REQ-WC11-018 (SPEC-WEB-CONSOLE-011) for the `cache` section ONLY — reclassifying cache from the machine/state exclusion group to a user-facing seam-writable section with a 2-key exposure scope. The remaining REQ-WC11-018 exclusions (state, system, project, sunset; tool-policy, lsp, mx; constitution, context, design, interview) shall remain intact; `db` shall likewise remain non-writable — 단 db의 제외 근거는 REQ-WC11-018 잔여군이 아니라 REQ-WC11-019(3-key editable carve-out) 이후의 콘솔 표면 제거(settings SSOT)다.

**REQ-WC13-002 (Ubiquitous):** The section-routing SSOT (`internal/settings/sectionroute.go`) shall register `"handoff"` and `"cache"` as `RouteSeam`; `ExcludedSections()` shall no longer list `"cache"`; `SeamSections()` shall include both new sections.

**REQ-WC13-003 (Ubiquitous):** The seam write whitelist (`internal/settings/sectionwrite.go` `sectionRootKeys`) shall gain `"handoff": {handoff}` and `"cache": {cacheStrategy}` so that seam edits targeting any other top-level key are rejected without touching the file.

**REQ-WC13-004 (When):** When the routing entries land, the guard tests encoding the section scope (`internal/settings/sectionroute_test.go`, `internal/settings/sectionwrite_test.go`, and the `internal/web` coverage/scope test family) shall be updated to accept exactly the new scope and to keep explicit reject cases for the remaining excluded set.

**REQ-WC13-005 (Unwanted):** The web write path shall not invoke a typed-struct re-marshal (the `ConfigManager.Save` family) on handoff.yaml or cache.yaml — both sections persist EXCLUSIVELY through the yamlpatch seam; the typed `HandoffConfig` / `CacheConfig` structs remain read-side only.

**REQ-WC13-006 (While):** While a seam write targets cache.yaml, the un-exposed keys (`spec_ttl`, `min_cacheable_tokens`) and all YAML comments shall be preserved verbatim outside the edited scalar (yamlpatch unmodeled-key/comment preservation invariant).

### §2.2 M2 — handoff + cache 섹션 노출

**REQ-WC13-010 (Ubiquitous):** The web console shall expose `handoff.mode` (select) and `handoff.guide` (boolean) as editable fields in a new handoff section, reusing the generic schema fieldset render path (evidence: runtime reader `internal/hook/handoff_inject.go:146-159`).

**REQ-WC13-011 (When):** When a `handoff.mode` value is submitted, the validator shall accept exactly the closed set {`manual`, `auto`} and reject any other value with a 4xx response leaving the file unchanged.

**REQ-WC13-012 (Ubiquitous):** The web console shall expose `cacheStrategy.enabled` (boolean) and `cacheStrategy.session_ttl` (select) as editable fields in a new cache section (evidence + caveat per acceptance.md §D.1).

**REQ-WC13-013 (When):** When a `cacheStrategy.session_ttl` value is submitted, the validator shall accept exactly the closed set {`1h`, `5m`, `off`} sourced from `internal/config/cache_config.go` `validSessionTTLs` — either by exporting the set or by a mirrored declaration guarded by a symmetry test; a divergent re-declaration without the symmetry guard is prohibited.

**REQ-WC13-014 (When):** When a new field or section is added to the shared schema, i18n keys for all 4 locales (en/ko/ja/zh) shall be added to `internal/web/assets/i18n.js` within the same milestone, following the existing `sec.<id>.title` / `sec.<id>.desc` + field-label key patterns.

**REQ-WC13-015 (Unwanted):** The cache section shall not render `spec_ttl` or `min_cacheable_tokens` as form fields (신설 FieldDef·persist 대상에서 제외; 파일 내 키는 REQ-WC13-006이 보존).

**REQ-WC13-016 (Where):** Where a new FieldDef carries the `PersistSeam` persist kind, the TUI-side schema bridge shall require NO code change — the existing persist-kind exclusion predicate (`internal/cli/schema_bridge_test.go:33`) scopes web-only fields; run-phase shall verify by test run that every new field rides the predicate.

**REQ-WC13-017 (Ubiquitous):** The two new sections shall reuse the existing generic schema fieldset pipeline (SectionID 등재 → `AllSections()` 렌더 순서 → FieldDefs → sectionvalues/저장 경로) — no bespoke per-section template is introduced for handoff/cache.

### §2.3 M3 — Model Policy 통합 READ-ONLY 뷰

**REQ-WC13-020 (Ubiquitous):** The web console shall present a consolidated Model Policy view rendering (a) the active performance tier from `llm.performance_tier` and (b) the `workflow.model_routing_profiles` map as a 3 perfTier × 12 cell (S/M/L × plan/run/sync/mx) table — BOTH as READ-ONLY displays with no form inputs.

**REQ-WC13-021 (Unwanted):** The Model Policy view shall not create any write path, PersistTarget, FieldDef persist binding, or form POST field for `model_routing_profiles` cells or `llm.performance_tier` (evidence: `RouteModelFor` has 0 non-test callers outside internal/config — TOKEN-ROUTING-001 D1 wiring debt; no runtime reader consumes edits).

**REQ-WC13-022 (Unwanted):** The legacy flat `workflow.model_routing` block shall not be rendered at all — neither editable nor read-only (deprecated alias; 0 non-config consumers).

**REQ-WC13-023 (Ubiquitous):** `workflow.workflow_agents` shall remain hidden from web render per the prior M5-a B1 decision (struct fields and yaml keys preserved; dynamic-workflow scripts read the yaml directly); the Model Policy view MAY reference the 7-purpose taxonomy in prose but shall not re-add workflow_agents FieldDefs.

**REQ-WC13-024 (Where):** Where `llm.performance_tier` is the empty string (현 로컬 상태), the Model Policy view shall render the "(runtime default: medium)" empty-value label — following the `EmptyLabelKey` display convention (`internal/settings/schema.go`) as an i18n-label precedent ONLY, without creating any FieldDef binding (뷰는 FieldDef-less — plan.md §A.4-2와 정합).

**REQ-WC13-025 (When):** When the Model Policy view lands, i18n keys for all 4 locales shall be added in the same milestone, including a disambiguation note distinguishing `model_policy` (launch 섹션 기존 필드 — init-time agent frontmatter 정책, high|medium|low) from `performance_tier` (routing profile 선택자, max|medium|low).

**REQ-WC13-026 (Where):** Where the `model_routing_profiles` block is absent from workflow.yaml, the view shall render the documented fallback state (all lookups fall back to `inherit`/`medium` per `defaultRoutingEntry`) instead of an error.

### §2.4 M4 — 검증 및 무접촉 계약

**REQ-WC13-030 (Unwanted):** This SPEC's run-phase shall not modify any file under `internal/statusline/` (병렬 세션이 `renderer.go` + `cache_hit_test.go` 수정 소유 중 — 2026-07-10 git status 실측).

**REQ-WC13-031 (Ubiquitous):** The template tree (`internal/template/templates/`) shall require no modification — the handoff.yaml / cache.yaml template mirrors already carry all exposed keys (2026-07-10 실측; template handoff default `mode: manual` 유지). When a run-phase discovery contradicts this, the change shall follow Template-First (`make build` 재임베드) + §25 neutrality (SPEC ID 주석 금지) and be reported as a scope delta.

**REQ-WC13-032 (When):** When run-phase completes, the full verification batch shall pass: `go build ./...`, `GOOS=windows GOARCH=amd64 go build ./...`, `go test ./internal/settings/... ./internal/web/... ./internal/cli/...`, and `golangci-lint run` with zero NEW issues (pre-existing baseline separately noted).

---

## §3 Exclusions (Out of Scope)

### Out of Scope — Statusline 세그먼트 노출 (16-segment map)

- `statusline.yaml`의 16-세그먼트 맵(cache_hit, usage_5h/7d, task, pr 포함)의 웹 콘솔 재노출은 본 SPEC 범위 밖이다.
- 근거: 2026-07-06 사용자 피드백이 statusline 섹션을 웹 콘솔에서 제거했다. 재도입은 유예된 사용자 결정이며 본 SPEC이 선점하지 않는다.
- `RouteStatusline` 전용 경로 및 `internal/statusline/*` 파일은 무접촉 (REQ-WC13-030).

### Out of Scope — model_routing_profiles / performance_tier 편집 경로

- 3×12 라우팅 셀과 `llm.performance_tier`의 웹 편집(쓰기 경로)은 범위 밖 — READ-ONLY 표시만 (REQ-WC13-020/021).
- 근거: `RouteModelFor` 비테스트 호출자 0건 (TOKEN-ROUTING-001 D1 wiring 부채). 편집을 소비하는 런타임이 없는 키의 편집 노출은 config theater다.
- 재검토 트리거: D1 wiring이 별도 SPEC으로 착지해 `RouteModelFor` 호출자가 생기면 편집 승격을 후속 SPEC으로 검토.

### Out of Scope — workflow_agents 웹 렌더 재노출

- `workflow.workflow_agents`(7-purpose 맵)의 웹 폼 재노출은 범위 밖 — M5-a B1 결정(웹 렌더 숨김, dynamic-workflow JS 직접 읽기)을 유지한다 (REQ-WC13-023).

### Out of Scope — cache.yaml 잔여 2키 (spec_ttl / min_cacheable_tokens)

- `cacheStrategy.spec_ttl`과 `cacheStrategy.min_cacheable_tokens`의 폼 노출은 범위 밖 (REQ-WC13-015). seam의 unmodeled-key 보존이 파일 값을 무손상 유지한다 (REQ-WC13-006).

### Out of Scope — 캐시/라우팅 런타임 배선 부채 해소

- `InjectCacheControl`(cache_control 주입기)의 요청 경로 배선, `RouteModelFor`의 spawn call-site 배선은 각각 SPEC-V3R6-PROMPT-CACHE-001 / SPEC-TOKEN-ROUTING-001 계보의 별도 부채이며 본 SPEC이 해소하지 않는다.
- `internal/config/types.go:245`의 `PerformanceTier` validator `oneof=high medium low` vs `ValidPerformanceTiers()` {max, medium, low} 불일치(선재 결함 후보)의 수정도 범위 밖 — plan.md §B Known Issues로 기록, 표시가 차단되면 blocker report.

### Out of Scope — legacy model_routing flat 블록 표시

- DEPRECATED `workflow.model_routing` flat 12-key 블록은 어떤 형태로도 렌더하지 않는다 (REQ-WC13-022). yaml 파일 내 블록 자체의 제거(정리)도 범위 밖.

---

## §4 Traceability

- 선행(blocking): **SPEC-WEB-CONSOLE-012** (Track 1 cleanup — `internal/settings/schema_sections.go` + `sectionroute.go` 동일 파일 접촉; 선행 착지 후 본 SPEC run 진입). 2026-07-10 현재 012는 authored 상태(`status: draft`, Tier S) — depends_on 게이트는 012가 `completed`가 될 때까지 run 진입을 차단하며(정상 동작), plan.md §C pre-flight가 이를 기계 확인한다.
- 부분 supersede: SPEC-WEB-CONSOLE-011 REQ-WC11-018 (cache 한정, REQ-WC13-001).
- 도메인 원천: SPEC-HANDOFF-AUTORESUME-001 (handoff.yaml), SPEC-V3R6-PROMPT-CACHE-001 (cache.yaml), SPEC-TOKEN-ROUTING-001 + SPEC-AGENT-ARCH-V2-001 M3 (model_routing_profiles / performance_tier).
- 본 문서의 모든 파일:라인 앵커는 2026-07-10 plan-phase grep 실측값이다. run-phase 착수 시 content-token 기준 재검증 의무 (plan.md §C).
