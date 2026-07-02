---
id: SPEC-WEB-CONSOLE-011
title: "Web Console Redesign — Agent Settings + Profile CRUD + SPEC Board + Full Config Expansion"
version: "0.2.1"
status: draft
created: 2026-07-03
updated: 2026-07-03
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/web"
lifecycle: spec-anchored
tags: "web, console, agent-settings, profile-crud, spec-board, statusline, yaml-node, config-expansion, workflow-agents, i18n"
tier: L
era: V3R6
related_specs: [SPEC-WEB-CONSOLE-010, SPEC-V3R6-WORKFLOW-EFFORT-MAP-001, SPEC-GITSTRATEGY-SAVE-ISOLATION-001, SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001]
---

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.2.1 | 2026-07-03 | manager-spec | plan-audit iter-1 D1-D6 + D8d/D9d 정정. D1 REQ 총계 42→49 (블록 열거는 정확, 총계만 오산 — plan.md/progress.md), D2 REQ-WC11-032/033 delete-guard 전제 정정 (live 코드는 default만 차단, active profile은 stderr 경고 후 삭제 진행 — active-profile 4xx 차단은 신설 웹 경계 로직), D3 AC-WC11-005 vacuous grep → `NewConfigManager\|\.Save\(` + allowlist 검증으로 교체, D4 AC-WC11-002 실존+신설 테스트 함수명 명시 바인딩(-run 무매치 exit 0 함정 제거), D5 AC-WC11-060 주석 제외 필터 추가(app.go:84 기존 주석 false-fail 해소), D6 AC-WC11-023 RoleProfileEntry struct 블록 한정(신설 WorkflowAgentEntry.Effort와의 자기모순 해소), D8d REQ-WC11-016 read-only carve-out, D9d §2.3.1 AC→REQ 매핑 주석. D10d(house-convention orphan AC)는 noticed-only 잔존. |
| 0.2.0 | 2026-07-03 | manager-spec | 사용자 4결정 확정 (2026-07-03): M2 10섹션 전면 확장 + M3 쓰기 전면 지원으로 개정. 결정 3 변경 — config는 선택 확장이 아닌 10개 user-facing 섹션 전면 편집 (Save() 경로 없는 8섹션은 M1 seam이 load-bearing). 결정 4 변경 — agent 설정 4표면 전면 쓰기: sub-agent frontmatter 편집(전용 patch layer + live-only + 지속 경고), dynamic-workflow는 신규 typed `workflow_agents` 표면. 신설 REQ-WC11-016..019, 027..029, 062, 070..074 (기존 ID 유지, 개정 ID: 001, 002, 022, 025, 026, 050, 053). 총 필드 수 하드코딩 assertion(34→35) → 파생(derived) assertion으로 전환. |
| 0.1.0 | 2026-07-03 | manager-spec | 최초 draft. 2026-07-03 사용자 승인 4결정(당시: 선택적 config 확장 + staged write) 반영. 근거 조사 = 2026-07-03 orchestrator survey (research.md). |

---

## §1 Context & Motivation

`moai web` 콘솔은 현재 단일 페이지 HTMX+Templ 앱이다: loopback 전용, 라우트는 `GET /`, `POST /save`, `POST /__shutdown__`, `GET /static/*` (internal/web/app.go:78-88). 필드 정의는 SPEC-WEB-CONSOLE-010이 확립한 공유 SSOT `internal/settings/schema.go`(6 섹션 / 34 필드, schema.go:24-31 + allFields 233-430)에서 파생되며 TUI 위저드와 동일 스키마를 소비한다. 영속화는 `handleSave`(handlers.go:287-316)의 4개 seam — `profile.WritePreferences`, `profile.SyncToProjectConfig`(sync.go:17; statusline은 sync.go:103에서 ConfigManager를 우회하는 직접 yaml.Marshal), `writeProjectConfig`(projectconfig.go:216), `settings.WriteProjectNestedConfig`(nested.go:102) — 로 분기한다. 쓰기 안전 모델은 loopback bind + Host-check 미들웨어뿐이며, app.go:90-92의 @MX:NOTE(REQ-WC-009)가 CSRF/token 인프라 도입을 금지한다 — 본 SPEC은 이 모델을 **보존**한다.

### §1.1 사용자 확정 스코프 (4결정 최종 확정, 2026-07-03 사용자 직접 선택 — 재론 금지)

1. **SPEC 보드**: READ-ONLY 대시보드로 KEEP (drag-drop 없음, 쓰기 경로 없음).
2. **Profile**: CRUD(create/delete/switch UI) 추가; 전제조건 = 웹 쓰기 경로의 profile-name 검증 갭을 reproduction-test-first로 수정.
3. **Config 커버리지 (v0.2.0 변경 — 전면 확장)**: **10개 user-facing 섹션 전부 편집 가능** — git-strategy, llm, workflow, harness, ralph, research, feedback, observability, security, db. 존속 제약: (a) machine/state 섹션 제외(state, system, project, cache, sunset); (b) 대형 정책 파일 제외(tool-policy, lsp, mx); (c) 커버 섹션 내 runtime-managed 키는 READ-ONLY — `llm.mode`/`llm.team_mode`(moai glm/cg가 runtime 기록, 레이스), db.yaml의 5개 system 키(인터뷰 입력 3키 `orm`/`multi_tenant`/`migration_tool`만 편집); (d) form UI에 맞지 않는 map-of-structs 서브블록(harness.yaml `levels` map 내부, workflow.yaml `team.patterns`)은 read-only 또는 collapsed raw view — 스칼라 키는 폼 필드. **Save() 경로가 없는 8개 섹션(workflow, harness, ralph, research, feedback, observability, security, db)은 전부 M1 yaml.Node comment-preserving seam이 load-bearing 의존성이다.**
4. **Agent 설정 (v0.2.0 변경 — 4표면 전면 쓰기, staged 아님)**: (a) llm.yaml tiers 편집; (b) team role profiles — yaml.Node seam 경유 편집 (effort는 opaque-node 유지, design.md §B); (c) sub-agent frontmatter(`.claude/agents/moai/*.md` `model:`/`effort:`) **편집** — frontmatter 전용 read/patch layer 신설(body 무접촉) + live 파일만 기록 + "moai update가 덮어쓸 수 있음" 지속 경고, template dual-write 없음; (d) dynamic-workflow model/effort **편집** — workflow.yaml 신규 typed `workflow_agents:` 블록(7-purpose taxonomy map → {model, effort}) 경유, v4manifest closed set 검증, seam 기록, dynamic-workflows.md(live + template mirror) SSOT 참조 갱신.

### §1.2 기반 차단 요소 (M1이 해소)

- **Scope contract**: REQ-WC-012(internal/web/server.go:10-13) + REQ-WC3-007(projectconfig.go:158 @MX:WARN)이 workflow/harness/git-strategy/llm yaml 접촉을 금지한다. 본 SPEC이 해당 조항을 **10개 user-facing 섹션 전체에 대해** 공식 SUPERSEDE한다(문서 + guard test `projectconfig_scope_test.go`, `coverage_test.go` 갱신). 웹 쓰기 제외군은 이제 machine/state 섹션 + 대형 정책 파일 + 미지명 잔여 섹션이다 (REQ-WC11-018).
- **Typed re-marshal 파괴성**: `ConfigManager.Save()`는 6개 파일만 기록하며(manager.go:166, 207-219) struct 재직렬화가 yaml 주석과 미모델링 키를 파괴한다. 특히 workflow.yaml: `RoleProfileEntry`(internal/config/types.go:373-378)는 명시적 결정(REQ-WEM-006, SPEC-V3R6-WORKFLOW-EFFORT-MAP-001)으로 Effort 필드가 없고, `team.patterns`는 의도적 미모델링(EXCL-WSE-004, types.go:359-360)이다. v0.2.0 전면 확장으로 **Save() 경로 없는 8개 섹션 전부**가 M1의 comment/unknown-key-preserving YAML patch seam(gopkg.in/yaml.v3 node surgery)에 의존한다.

> 본 문서의 모든 파일:라인 앵커는 2026-07-03 orchestrator survey 실측값이다. run-phase 착수 시 content-token 기준 재검증 의무 (plan.md §C). 섹션 파일 실존(10 + 제외군) 및 dynamic-workflows.md / workflow.yaml 템플릿 미러 실존은 2026-07-03 plan-phase에서 `ls` 실측 확인됨 (research.md §I).

---

## §2 Requirements (GEARS)

### §2.1 M1 — Foundation (scope supersede + yaml.Node patch seam)

**REQ-WC11-001 (Ubiquitous — v0.2.0 개정):** This SPEC shall formally supersede the web-console scope-contract clauses REQ-WC-012 (internal/web/server.go:10-13) and REQ-WC3-007 (internal/web/projectconfig.go:158 @MX:WARN) for ALL 10 user-facing config sections — git-strategy, llm, workflow, harness, ralph, research, feedback, observability, security, db. The machine/state sections (state, system, project, cache, sunset), the large policy files (tool-policy, lsp, mx), and any section not named editable shall remain outside the web write scope.

**REQ-WC11-002 (When — v0.2.0 개정):** When the scope supersede lands, the guard tests `internal/web/projectconfig_scope_test.go` and `internal/web/coverage_test.go` shall be updated to encode the NEW scope contract — permitting exactly the 10 named sections and rejecting the excluded set (state, system, project, cache, sunset, tool-policy, lsp, mx, 및 미지명 섹션) with explicit reject test cases.

**REQ-WC11-003 (Ubiquitous):** The web persistence layer shall gain a comment- and unknown-key-preserving YAML patch seam (gopkg.in/yaml.v3 node surgery) that edits only the targeted keys of a section file while preserving comments, unmodeled keys (e.g. `team.patterns`), and key order in untouched regions.

**REQ-WC11-004 (Where):** Where a config section edited by the web console is NOT fully modeled by its Go struct (workflow.yaml: `team.patterns` per EXCL-WSE-004; role-profile `effort` per REQ-WEM-006), the write path shall route through the yaml.Node patch seam instead of a typed struct re-marshal.

**REQ-WC11-005 (Unwanted):** The web write path shall not invoke a typed-struct re-marshal (the `ConfigManager.Save` family) on workflow.yaml.

### §2.2 M2 — Config 전면 확장 (10섹션)

**REQ-WC11-010 (Ubiquitous):** The web console shall expose the git-strategy section in full as editable fields, persisting through the existing typed Save dirty-flag path (internal/config/manager.go:207-216, per SPEC-GITSTRATEGY-SAVE-ISOLATION-001).

**REQ-WC11-011 (Ubiquitous):** The web console shall expose the remaining un-exposed quality section keys as editable fields following the existing `FieldDef` + `PersistTarget` pattern (구체 키 목록은 run-phase에서 quality.yaml 대비 schema diff로 확정 — design.md §C).

**REQ-WC11-012 (Ubiquitous):** The web console shall expose the llm.yaml safe keys (typed `LLMConfig`, types.go:234-283, validated oneof — including the claude_models tiers) as editable fields.

**REQ-WC11-013 (Ubiquitous):** The web console shall render `llm.mode` and `llm.team_mode` as READ-ONLY display fields (runtime-managed by `moai glm` / `moai cg`; concurrent-write race hazard).

**REQ-WC11-014 (Where):** Where a claude_models tier value is the empty string, the web console shall render the "(runtime default)" empty-value semantics reusing the existing `EmptyLabelKey` pattern (internal/settings/schema.go:225-229).

**REQ-WC11-015 (When):** When a new field is added to the shared schema, i18n keys for all 4 locales (en/ko/ja/zh) shall be added to `internal/web/assets/i18n.js` within the same milestone.

**REQ-WC11-016 (Ubiquitous — v0.2.0 신설, v0.2.1 D8d carve-out):** The web console shall expose ALL 10 user-facing config sections as editable — git-strategy, llm, workflow, harness, ralph, research, feedback, observability, security, db — rendering every scalar key as a form field, **except keys designated read-only (REQ-WC11-013/019)** (map-of-structs 서브블록의 처리는 REQ-WC11-062).

**REQ-WC11-017 (While — v0.2.0 신설):** While a covered section lacks a typed `Save()` path — workflow, harness, ralph, research, feedback, observability, security, db (8개) — the web write path shall persist that section EXCLUSIVELY through the M1 yaml.Node patch seam. **M1 seam은 이 8개 섹션 전부의 load-bearing 의존성이다.**

**REQ-WC11-018 (Unwanted — v0.2.0 신설):** The web console shall not expose the machine/state sections (state, system, project, cache, sunset), the large policy files (tool-policy, lsp, mx), nor any section not named in REQ-WC11-016 or the existing 34-field coverage (실측 잔여: constitution, context, design, interview) — 폼·라우트·persist 대상에서 전부 제외.

**REQ-WC11-019 (Where — v0.2.0 신설):** Where the db section is rendered, only the 3 interview-input keys (`orm`, `multi_tenant`, `migration_tool`) shall be editable; the remaining 5 system keys shall render READ-ONLY, and write attempts against them shall leave the file values unchanged (5개 system 키의 명칭은 run-phase pre-flight에서 db.yaml 실측 열거).

### §2.3 M3 — Agent Settings 페이지 (4표면 전면 쓰기)

**REQ-WC11-020 (Ubiquitous):** The web console shall present a single agent-settings view exposing ALL FOUR agent model/effort surfaces: (a) llm.yaml tiers, (b) workflow.yaml `team.role_profiles` (7 profiles, L81-125 + `default_model` L18 + `role_profile_keys` L29-32), (c) sub-agent frontmatter (`.claude/agents/moai/*.md`, 7 agents), (d) dynamic-workflow per-purpose model/effort (§2.3.1 `workflow_agents`).

**REQ-WC11-021 (Where):** Where the surface is llm.yaml tiers, the agent-settings view shall allow editing through the typed llm persistence path.

**REQ-WC11-022 (Where — v0.2.0 개정, staged 표현 제거):** Where the surface is workflow.yaml `team.role_profiles`, the agent-settings view shall route ALL edits through the M1 yaml.Node patch seam (M1 의존 — typed 경로 금지, REQ-WC11-005).

**REQ-WC11-023 (Ubiquitous):** The role-profile editor shall treat the `effort` key as an opaque yaml.Node (Go-invisible), patching it in place WITHOUT adding an `Effort` field to `RoleProfileEntry` — the REQ-WEM-006 decision (SPEC-V3R6-WORKFLOW-EFFORT-MAP-001) is NOT reversed (design.md §B — v0.2.0 전면 쓰기 하에서도 유지됨을 확인).

**REQ-WC11-024 (When):** When a role-profile enum value (effort or model tier) is submitted, the validator shall reuse the closed sets in `internal/harness/v4manifest/schema.go:41-73` (5 efforts, 4 model tiers) rather than re-declaring option lists.

**REQ-WC11-025 (Ubiquitous — v0.2.0 개정, read-only → 편집):** The agent-settings view shall allow editing the `model:` / `effort:` frontmatter values of the 7 sub-agent files through the NEW frontmatter-only patch layer (REQ-WC11-027..029).

**REQ-WC11-026 (Ubiquitous — v0.2.0 개정, read-only → 편집):** The agent-settings view shall allow editing the dynamic-workflow per-purpose model/effort defaults through the NEW typed `workflow_agents` config surface (§2.3.1), retaining a reference link to the 7-purpose taxonomy prose (`.claude/rules/moai/workflow/dynamic-workflows.md` L82-103).

**REQ-WC11-027 (Ubiquitous — v0.2.0 신설):** The frontmatter patch layer shall parse and patch ONLY the YAML frontmatter block of an agent file, reassembling the file with the original body bytes verbatim (body 무접촉); writes shall target the LIVE file only, and the layer shall not perform any automatic dual-write to `internal/template/templates` (해당 미러는 moai-adk-go dev repo에만 존재하며 사용자 프로젝트에는 없다 — design.md §C.1 template-mirror policy).

**REQ-WC11-028 (While — v0.2.0 신설):** While a template-managed agent file is presented as editable, the view shall surface a persistent warning that `moai update` may overwrite live edits (i18n ×4).

**REQ-WC11-029 (When — v0.2.0 신설):** When a frontmatter `model:` / `effort:` value is submitted, validation shall use the closed sets from `internal/harness/v4manifest/schema.go:41-73` — model ∈ {inherit, haiku, sonnet, opus} (4-tier), effort ∈ {low, medium, high, xhigh, max} (5-value); absence of the `effort` key shall be VALID (manager-docs / manager-git 선례) and the editor shall support leaving or setting the key absent (부재 시 빈 값 주입 금지).

### §2.3.1 M3 — `workflow_agents` 신규 typed 표면 (v0.2.0 신설)

**REQ-WC11-070 (Ubiquitous):** workflow.yaml shall gain a NEW `workflow_agents:` block — a purpose-taxonomy map (the 7 purposes of `.claude/rules/moai/workflow/dynamic-workflows.md` L82-103) → `{model, effort}` — serving as the SSOT for dynamic-workflow model/effort DEFAULTS, while per-script literals remain overrides.

**REQ-WC11-071 (Ubiquitous):** The `workflow_agents` block shall be typed in Go — a new struct + loader wiring in `internal/config` (읽기는 typed, 쓰기는 REQ-WC11-073 seam).

**REQ-WC11-072 (When):** When a `workflow_agents` model/effort value is submitted, the validator shall reuse the same v4manifest closed sets as REQ-WC11-024/029 and reject out-of-set values with a 4xx.

**REQ-WC11-073 (Ubiquitous):** `workflow_agents` writes shall route through the M1 yaml.Node patch seam; 최초 기록 시(블록 부재 — 2026-07-03 grep 0 실측) 누락 경로의 upsert 생성을 허용한다 (design.md §A.2 seam upsert 확장).

**REQ-WC11-074 (Ubiquitous):** `.claude/rules/moai/workflow/dynamic-workflows.md` (live + template mirror — 양쪽 실존 2026-07-03 확인) shall be updated to reference the `workflow_agents` config surface as the defaults SSOT (per-script literals = overrides), and the template workflow.yaml mirror shall gain the default block — 별도 work item으로 관리하며 Template-First(`make build` 재임베드) + §25 neutrality(SPEC ID 주석 금지)를 준수한다.

### §2.4 M4 — Profile CRUD (repro-test-first 보안 수정 → routes/UI)

**REQ-WC11-030 (Ubiquitous):** A failing reproduction test shall FIRST establish (or refute) the hypothesized defect: the web write path accepts arbitrary `?profile=` / `__profile` values (app.go:133-141, handlers.go:252-254) flowing into `GetPreferencesPath` / `WritePreferences` (preferences.go:82-88, 150-165) WITHOUT `isValidProfileName` (profile.go:126-132) — plausible path traversal (`__profile=../../x`) plus undocumented implicit-create via MkdirAll. 본 결함은 UNVERIFIED HYPOTHESIS이며(verification-claim-integrity §1.1 surface 3), repro test가 기계적 검증 그 자체다.

**REQ-WC11-031 (When):** When the reproduction test confirms the defect, the web write path shall enforce `isValidProfileName` (at the web boundary AND/OR inside `GetPreferencesPath`/`WritePreferences` — 배치 결정은 design.md §D) such that the reproduction test turns green.

**REQ-WC11-032 (Ubiquitous — v0.2.1 D2 정합):** The web console shall provide profile CRUD UI — create / switch / delete — via approximately 2 new POST routes + Templ fragments + 4-locale i18n, reusing the existing backend primitives `profile.List` / `profile.EnsureDir` / `profile.Delete`. 단, active-profile 삭제를 차단하는 4xx guard는 기존 primitive의 재사용이 **아니라** REQ-WC11-033이 신설하는 웹 경계 로직이다.

**REQ-WC11-033 (Unwanted — v0.2.1 D2 전제 정정):** The profile delete route shall not delete the `default` profile nor the currently active profile. 전제 정정(plan-audit iter-1 D2): live 코드의 기존 guard는 `default`만 차단하며, ACTIVE profile은 stderr 경고 후 삭제가 **진행**된다(profile.go:98-105 — RemoveAll 수행). 따라서 default 차단은 기존 guard 재사용이지만, **active-profile 삭제 차단(4xx)은 본 SPEC이 웹 경계에 NEW 로직으로 신설**한다 — 기존 delete guards의 재사용이 아니다.

**REQ-WC11-034 (When — undesired condition detected):** When a profile name failing `isValidProfileName` is submitted to any profile route, the handler shall reject the request with a 4xx response and shall not create any directory (no implicit MkdirAll side effect).

### §2.5 M5 — SPEC READ-ONLY Board

**REQ-WC11-040 (Ubiquitous):** The web console shall provide a READ-ONLY SPEC dashboard rendering the status distribution, an implemented-not-completed close-debt column, and MUST-FIX SyncStatusDrift badges (audit.go:255-300), sourced from `spec.Audit` (internal/spec/audit.go:94 — pure FS scan, 실측 0.14s / 412 SPECs).

**REQ-WC11-041 (Ubiquitous):** The `internal/spec` package shall gain an exported `ListDocs(baseDir)` wrapper around the unexported `discoverSPECs` / `parseSPECDoc` (lint.go:224 / 311).

**REQ-WC11-042 (Ubiquitous):** `SPECFrontmatter` (lint.go:257-287) shall gain a `Tier string yaml:"tier"` field; the board shall render tier as an OPTIONAL badge (tier 보유 177/412 — 부재 시 badge 생략, 오류 아님).

**REQ-WC11-043 (Ubiquitous):** The board shall render each MUST-FIX finding's remediation string as COPYABLE TEXT (e.g. `moai spec close <ID> --backfill-only`).

**REQ-WC11-044 (Unwanted):** The board handler shall not execute any remediation command server-side.

**REQ-WC11-045 (Unwanted):** The synchronous board render shall not invoke the git-dependent `DetectDrift` path (실측 7.9s) — only the pure-FS `Audit` scan.

**REQ-WC11-046 (Unwanted):** The board shall not provide any status write path — status transitions are agent-owned per the Status Transition Ownership Matrix (`.claude/rules/moai/development/spec-frontmatter-schema.md`; hook `status-transition-ownership.sh` exits 2 on owner mismatch).

### §2.6 M6 — Statusline cache_hit delta + segment-list SSOT

**REQ-WC11-050 (Ubiquitous — v0.2.0 개정, 총계 하드코딩 제거):** The `cache_hit` segment (`SegmentCacheHit`, internal/statusline/types.go:329 — renderer에서 toggleable, default-on, hand-edited yaml honored) shall be exposed across ALL segment-list surfaces where it is currently absent: `settings.statuslineSegmentKeys` (settings/schema.go:106-122), `statusline.CanonicalSegments` (preset.go:12-18), profile `defaultStatuslineSegments` (sync.go:155-178), TUI `statuslineAllSegments`, both statusline.yaml files (live + template mirror), and i18n.js — statusline segment count 15→16. 스키마 **총** 필드 수는 M2 전면 확장으로 유동이므로 하드코딩 총계(구 34→35 pin)를 요구하지 않는다 — 총계 assertion은 파생(derived) 방식 (REQ-WC11-053).

**REQ-WC11-051 (Ubiquitous):** A set-equality test shall bind the 3+ hardcoded segment lists to one SSOT, so that a future segment addition cannot silently orphan again (design.md §F).

**REQ-WC11-052 (Where):** Where the live `.moai/config/sections/statusline.yaml` gains the `cache_hit` key, the template mirror `internal/template/templates/.moai/config/sections/statusline.yaml` shall gain the same key within the same milestone (Template-First rule; 템플릿 편집 내용은 §25 neutrality 준수 — SPEC ID 주석 금지).

**REQ-WC11-053 (When — v0.2.0 개정):** When the cache_hit exposure lands, the count labels and tests shall be updated consistently — `internal/web/fieldsets.templ:135`의 카운트 라벨은 schema 길이에서 **파생**되도록 전환(하드코딩 총계 리터럴 금지 — M2 확장으로 필드 총수 유동), `internal/web/statusline_test.go:53-54`는 statusline-국소 카운트(want 15→16)로 갱신 — and the stale "11-segment" comment at `internal/cli/profile_setup.go:333` shall be corrected in passing.

### §2.7 Cross-cutting

**REQ-WC11-060 (Ubiquitous + Unwanted):** The write-safety model shall remain loopback bind + Host-check middleware; the console shall not introduce CSRF/token infrastructure (REQ-WC-009 / @MX:NOTE app.go:90-92 PRESERVED).

**REQ-WC11-061 (Ubiquitous):** Every new user-facing UI string shall carry i18n keys for all 4 locales (en/ko/ja/zh); no new string ships English-only.

**REQ-WC11-062 (Where — v0.2.0 신설):** Where a covered section contains a map-of-structs sub-block unsuited to form UI (예: harness.yaml `levels` map 내부, workflow.yaml `team.patterns`), the web console shall render that sub-block read-only or as a collapsed raw view — 같은 섹션의 스칼라 키는 여전히 폼 필드로 렌더한다.

---

## §3 제외 범위 (Exclusions)

본 SPEC이 만들지 **않는** 것. 각 항목은 후속 SPEC 또는 영구 제외. (v0.2.0: 구 "10-section 전체 확장" / "sub-agent frontmatter WRITE 레이어" / "dynamic-workflow typed config surface" 3개 제외 항목은 스코프 IN으로 이동되어 삭제됨.)

### Out of Scope — 보드 쓰기 경로 (drag-drop / status 전이)

- SPEC 보드의 drag-drop, status 전이, frontmatter 편집 등 일체의 쓰기 경로. status 전이는 agent-owned (Status Transition Ownership Matrix).

### Out of Scope — GLM API key 관리 UI

- secrets는 `~/.moai/.env.glm`에 유지. 웹 UI에서 키 표시/편집/저장 일체 금지.

### Out of Scope — CSRF/token 인프라

- REQ-WC-009 보존. loopback bind + Host-check가 유일한 쓰기 안전 모델로 유지된다.

### Out of Scope — machine/state 섹션 · 대형 정책 파일 · 미지명 섹션

- machine/state: state, system, project, cache, sunset — 웹 노출·쓰기 전부 제외.
- 대형 정책 파일: tool-policy, lsp, mx — form UI 제외.
- 미지명 잔여 섹션(2026-07-03 실측: constitution, context, design, interview): REQ-WC11-016의 10-목록에도 기존 34-필드 커버리지에도 없는 섹션은 노출하지 않는다.

### Out of Scope — Agent frontmatter의 template dual-write

- sub-agent frontmatter 편집은 **live 파일만** 기록한다. `internal/template/templates`로의 자동 동기 기록 없음 — 해당 미러는 moai-adk-go dev repo 전용이며 사용자 프로젝트에 존재하지 않는다. (dev repo에서의 미러 정합은 메인테이너 수동 판단.)

### Out of Scope — env-only statusline knobs

- `CLAUDE_AUTOCOMPACT_PCT_OVERRIDE`, `MOAI_STATUSLINE_CONTEXT_SIZE` 등 환경변수 전용 노브의 웹 노출 제외.

### Out of Scope — render-constant 노출

- statusline 렌더 상수(bar width, handoff thresholds — context-window-management.md HARD rule에 고정)의 설정화/노출 제외.

---

## §4 Cross-References

- `SPEC-WEB-CONSOLE-010` — 공유 field-schema SSOT(internal/settings) 확립. 본 SPEC의 모든 신규 필드는 이 SSOT 패턴을 따른다.
- `SPEC-V3R6-WORKFLOW-EFFORT-MAP-001` REQ-WEM-006 — RoleProfileEntry에 Effort 필드를 두지 않는 결정. 본 SPEC은 이를 뒤집지 않는다 (REQ-WC11-023).
- `SPEC-GITSTRATEGY-SAVE-ISOLATION-001` — git-strategy Save dirty-flag 경로 (REQ-WC11-010의 영속화 기반).
- `SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001` — 은퇴한 `preset` 셀렉터. 본 SPEC의 statusline 작업에서 재도입 금지.
- `.claude/rules/moai/workflow/dynamic-workflows.md` — 7-purpose taxonomy prose (REQ-WC11-070 원천; live + template mirror 실존 2026-07-03 확인, REQ-WC11-074 갱신 대상).
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — Status Transition Ownership Matrix (REQ-WC11-046 근거).
- `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 3 — M4 결함 가설의 repro-test-first 의무 근거.
- plan.md / acceptance.md / design.md / research.md — 본 SPEC 디렉터리 내 자매 산출물.
