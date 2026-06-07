---
id: SPEC-WEB-CONSOLE-008
title: "statusline config redesign — remove config theater, wire live levers, fix drift (honest hybrid)"
version: "0.1.0"
status: in-progress
created: 2026-06-07
updated: 2026-06-07
author: Goos Kim
priority: P1
phase: "v0.2.0"
module: "internal"
lifecycle: spec-anchored
tags: "web-console, config-redesign, statusline, config-theater, templ, honest-hybrid, s2-cohort"
era: V3R6
related_specs: [SPEC-WEB-CONSOLE-006, SPEC-WEB-CONSOLE-007]
tier: M
---

# SPEC-WEB-CONSOLE-008 — statusline config redesign (honest hybrid)

## HISTORY

| Version | Date | Author | Summary |
|---------|------|--------|---------|
| 0.1.0 | 2026-06-07 | Goos Kim | 초안 — `.moai/reports/web-console-statusline-gitconvention-audit.md`(2026-06-07, 42 confirmed defects)의 **statusline 17-defect inventory + Redesign Proposal — statusline**를 권위 있는 연구·설계 근거로 사용한다. 처분 철학은 **honest hybrid**: (1) config theater 제거(거짓 표면 삭제) — `mode` config surface, `refresh_interval`, dead `loadSegmentConfig`; (2) live lever 배선 — `preset` write-effective; (3) 구조 drift 교정 — 3 divergent struct 통합, seed 15-key 완성, theme seed `default`→`catppuccin-mocha`, doc 정정, symmetry CI guard 추가. 2-SPEC 코호트 분할의 첫 멤버(008=statusline). git_convention(GCR-1..9 등 25-defect)은 **SPEC-WEB-CONSOLE-009로 분리** deferred. 본 SPEC은 `mode=full`/`renderFullV3`(5-line layout)을 **부활시키지 않는다** — 거짓을 제거할 뿐 재구현하지 않는다. |

---

## §A. 배경 및 Ground-Truth (cite, do NOT re-derive)

본 SPEC은 audit 보고서(`.moai/reports/web-console-statusline-gitconvention-audit.md`)의 statusline 섹션을 권위 있는 ground-truth로 삼는다. 모든 file:line 인용은 adversarial sonnet 검증을 통과했다. 아래는 본 SPEC이 의존하는 핵심 seam이며, run-phase에서 재발견(re-discover)하지 말 것.

### A.1 Config 모델 (3 divergent struct — SLM-1/2/SLR-2)

- **canonical**: `models.StatuslineConfig` (pkg/models/config.go:199-203) = `{Preset, Segments, Theme}`. **`Mode` 필드 없음** — canonical 로더가 `mode:`를 drop한다(SLM-1).
- **CLI private**: `statuslineFileConfig` (internal/cli/statusline.go:127-128) = `{Mode, ...}`; 로더 `loadStatuslineFileConfig`(statusline.go:137-163)가 `raw.Statusline.Mode`를 읽는 **유일한** 런타임 경로(SLR-2). statusline.go:50-55가 env/`statuslineCfg.Mode`로 mode를 결정하고 builder에 전달.
- **profile sync private**: `statuslineData` (internal/profile/sync.go:80-85) = `{Mode, Preset, Segments, Theme}` with `yaml:"mode,omitempty"`. `syncStatusline`(sync.go:96-145)가 `current.Statusline.Mode = prefs.StatuslineMode`를 씀(sync.go:120-121).

### A.2 런타임 거동 (두 functional lie — SLR-1/SLR-3)

- **SLR-1**: `mode=full`은 inert. `Renderer.Render(data, mode)`(internal/statusline/renderer.go:55)는 항상 3-line default로 collapse하고, `renderFullV3`(renderer.go:146, 5-line layout)는 production에서 호출되지 않는다(`nolint:unused`). 웹 dropdown + "Full — 5-line…" 설명은 존재하지 않는 layout을 약속한다.
- **SLR-3**: `preset=compact/minimal`은 near-vestigial. per-segment `segments` map(항상 YAML에 존재)이 런타임에서 preset을 **이긴다** — 비-custom preset 선택이 아무것도 바꾸지 않는다. 원 가설(preset=full이 all-on을 강제)을 **반박**한다.

### A.3 Builder API (PRESERVE — config surface와 구분)

- `StatuslineMode` enum(internal/statusline/types.go:13-44: ModeDefault/Full/Minimal/Compact/Verbose) + `NormalizeMode`(types.go:44) + `Render(data, mode StatuslineMode)`(renderer.go:55) + Builder Config의 `Mode StatuslineMode` 필드(builder.go:54) + `SetMode`(builder.go:160)는 **Builder API의 일부**이며 fan-out이 넓다 — **유지**한다. 본 SPEC은 오직 `mode:` YAML surface + 웹 control만 제거한다.

### A.4 seed / dead code / drift

- **SLM-5**: `defaultStatuslineSegments()`(internal/profile/sync.go:148-162)는 canonical 15 중 **11개만** 반환(누락: `effort_thinking`, `worktree`, `task`, `pr`).
- **SLR-7**: `loadSegmentConfig`(internal/cli/statusline.go:170-186)는 dead production code — production caller 0, 호출자는 테스트만(coverage_improvement_test.go:1289+, statusline_test.go:105+). 내부적으로 `presetToSegments(cfg.Preset, nil)` 호출.
- **presetToSegments**(internal/cli/update.go:2536)는 canonical 15-key 확장기(full/compact/minimal/custom 테스트됨) — DRY anchor.
- **SLR-4**: 템플릿 `statusline.yaml`이 `theme: "default"` ship(존재하는 테마는 catppuccin-mocha/latte 2개뿐) → 묵시적으로 mocha로 coerce + 웹 dropdown에서 non-round-trippable. 같은 파일이 `mode: "default"`, `refresh_interval: 10` ship.
- **SLR-5**: `builder.go:50` ThemeName doc가 존재하지 않는 "default"/"minimal"/"nerd"를 나열.
- **SLM-7**: canonical 카운트는 15; `SegmentRepo`("repo", types.go:321)는 schema 밖 16번째 상수(의도적 제외).
- **SLM-3/SLR-6**: `refresh_interval`는 어떤 struct도 모델하지 않음(consumer 0).
- **SLM-4**: `internal/config/audit_struct_yaml_symmetry_test.go:27-48`의 `symmetryCases`는 정확히 4개 MIG-003 섹션만 — `StatuslineConfig` 미포함.

### A.5 Web (internal/web — 006/007 산출물)

- `fieldsets.templ:128` 카운트 라벨 `"3 fields · segments"`; `fieldsets.templ:134` `@statuslineSelect("statusline_mode", ...)` mode control; preset/theme select(135/136).
- `handlers.go:29` `StatuslineModes []string`; `handlers.go:82` `StatuslineModes: statuslineModeCanonical`; `handlers.go:365` `StatuslineMode: r.PostFormValue("statusline_mode")`.
- `validate.go:40-41` `statuslineModeCanonical = []string{"default", "full"}`; `validate.go:136-137` mode 검증 블록.
- 재사용 위젯: `statuslineSelect`(fieldsets.templ:153), `segmentCheckbox`(fieldsets.templ:198). 폼은 `hx-boost` full-page swap(006/007 계약).

### A.6 Sentinel 계약 (HARD boundary)

- `internal/web/integration_test.go:197-205`(REQ-WC-012 boundary sentinel)은 workflow/harness/git-strategy가 `DO_NOT_TOUCH` byte-identical임을 단언 — **statusline은 보호 대상이 아님**(in-scope). 본 SPEC은 이 sentinel을 **무수정**으로 유지한다.
- `integration_test.go:158-165`은 statusline.yaml이 theme로 갱신됨을 단언 — statusline write는 기대된 거동.

### A.7 코호트 계보

SPEC-WEB-CONSOLE-006(HTMX+Templ 마이그레이션, completed `5714bae97`) → 007(quality+git_convention nested 편집, completed `54c09a61b`) → **008(본 SPEC, statusline 재설계)** + 009(git_convention 재설계, 분리).

---

## §B. HARD Invariants (must-pass — 모든 AC가 이로부터 도출됨)

| # | Invariant | Ground-truth anchor |
|---|-----------|---------------------|
| HARD-1 | **Builder API 보존**: `StatuslineMode` enum + `NormalizeMode` + `Render(data, mode)` 시그니처 + Builder Config `Mode` 필드 + `SetMode`는 변경 없이 유지. config surface(`mode:` YAML + 웹 control)만 제거. | types.go:13-44, renderer.go:55, builder.go:54, builder.go:160 |
| HARD-2 | **renderFullV3 부활 금지**: 5-line layout을 재구현·재호출하지 않는다. `mode=full`은 거짓이며 제거 대상 — wiring 대상 아님. | renderer.go:146 (`renderFullV3`) |
| HARD-3 | **006 scope-boundary sentinel(integration_test.go:197-205) 무수정 GREEN**. statusline은 in-scope이므로 statusline.yaml write는 허용·기대됨. 만약 설계가 이 sentinel 변경을 강제하면 → STOP, blocker. | integration_test.go:197-205 |
| HARD-4 | **CRITICAL SCOPE — git_convention 무관**: git_convention 섹션·struct·검증·웹 위젯을 전혀 건드리지 않는다(009 스코프). | audit git_convention inventory (25 defects) |
| HARD-5 | **canonical 모델에 Mode 추가 금지**: 3-struct 통합은 2개 private struct에서 `Mode`를 **제거**하는 방향. `models.StatuslineConfig`에 `Mode`를 추가하지 않는다. | config.go:199-203 |
| HARD-6 | **runtime precedence 불변**: `segments != nil`이 preset을 이기는 런타임 우선순위는 그대로. SLR-3 fix는 **write 경로가 항상 15-key map을 emit**하게 만드는 것(런타임 우선순위 변경 아님). | sync.go:129-131 |
| HARD-7 | **theme-only round-trip 보존(characterization gate)**: integration_test.go:124-165(theme만 변경, preset 미전송)은 preset write-effective 변경 후에도 GREEN. preset 미전송 시 load-modify-write preserve 경로(sync.go:129 guard)가 유지되어야 함. | integration_test.go:124-165 |
| HARD-8 | **server-canonical 검증 + offline + zero-Node 보존**: 웹은 직접 YAML marshal 금지; 신규 CDN/외부 asset 0; Templ codegen은 `templ generate`만. | assets.go //go:embed, templ.go |
| HARD-9 | **Template-first + neutrality**: 템플릿 YAML 편집은 `internal/template/templates/.moai/config/sections/statusline.yaml` → `make build`(go:embed embedded.go 재생성). 템플릿 주석은 SPEC ID/REQ 토큰/내부 인용 무포함(generic). | CLAUDE.local.md §2/§15/§25 |

---

## §C. GEARS 요구사항

### C.1 Ubiquitous (항상 활성)

- **REQ-WC8-001** — The statusline config schema **shall** expose exactly two live levers — `theme` (2 enum values) and the per-segment `segments` map (15 keys) — and **shall not** model `mode` or `refresh_interval` in any YAML-surface struct (SLM-3/SLR-6).
- **REQ-WC8-002** — The `models.StatuslineConfig` canonical struct **shall** remain `{Preset, Segments, Theme}` with no `Mode` field; the two private structs (`internal/cli/statusline.go` `statuslineFileConfig`, `internal/profile/sync.go` `statuslineData`) **shall** drop their `Mode` field so all three structs converge on the canonical shape (SLM-1/SLM-2/SLR-2, HARD-5).
- **REQ-WC8-003** — The statusline Builder API — the `StatuslineMode` enum, `NormalizeMode`, the `Render(data, mode StatuslineMode)` signature, the Builder Config `Mode` field, and `SetMode` — **shall** remain unchanged (HARD-1).
- **REQ-WC8-004** — `defaultStatuslineSegments()` **shall** return the canonical 15-key segment map, sourced from `presetToSegments("full", nil)` to prevent future drift (SLM-5).
- **REQ-WC8-005** — The template `statusline.yaml` **shall** seed `theme: "catppuccin-mocha"`, **shall not** carry `mode:` or `refresh_interval:`, and its comments **shall** list only the two real themes (SLR-4, HARD-9).
- **REQ-WC8-006** — The struct↔YAML symmetry test (`audit_struct_yaml_symmetry_test.go` `symmetryCases`) **shall** include a `StatuslineConfig` case so model↔YAML drift is CI-guarded (SLM-4); this case **shall** be added LAST, after the struct and template are already symmetric.

### C.2 Event-driven (When)

- **REQ-WC8-007** — **When** a user saves a non-custom preset (`full`/`compact`/`minimal`) via the console or profile sync, the write path **shall** expand the preset into a full 15-key `segments` map so the selection takes effect at render time (SLR-3, HARD-6).
- **REQ-WC8-008** — **When** the form submission carries no `preset` value (theme-only or segment-only change), the write path **shall** preserve the existing persisted `segments` map unchanged (preserve path, HARD-7).
- **REQ-WC8-009** — **When** a user opens the project console, the page **shall not** render a `statusline_mode` select control; the form **shall not** parse a `statusline_mode` form value (SLR-1 lie removed).
- **REQ-WC8-016** — **When** the `statusline_mode` control is removed from the web console, the section count label **shall** change from `"3 fields · segments"` to `"2 fields · segments"` (SLW-4, cosmetic). [D7-fix: 단일 조건 Event-driven으로 재분류 — 이전 §C.6 "Where…When…" Compound 오라벨 정정.]

### C.3 State-driven (While)

- **REQ-WC8-010** — **While** the selected preset is `custom`, the 15 `segment_<key>` checkboxes **shall** be editable and honored on save; **while** the preset is non-custom (`full`/`compact`/`minimal`), the checkboxes **shall** render display-only/disabled and the server **shall** expand the preset via `presetToSegments` (SLR-3 UX).

### C.4 Capability gate (Where)

- **REQ-WC8-011** — **Where** the statusline theme is set to an unrecognized value, the model layer's existing coercion safety net **shall** remain (coerce to mocha); the template seed fix removes the source of the non-round-trippable `default` value (SLR-4) without removing the runtime safety net.

### C.5 Unwanted behavior

- **REQ-WC8-012** — The codebase **shall not** retain the dead production function `loadSegmentConfig` (internal/cli/statusline.go:170-186) nor its test references (SLR-7).
- **REQ-WC8-013** — The console **shall not** re-introduce, re-call, or re-implement `renderFullV3` (the 5-line layout); the `mode=full` lie is removed, not wired (HARD-2).
- **REQ-WC8-014** — The change **shall not** touch, re-serialize, or alter any git_convention config surface, struct, validator, or web widget (HARD-4, 009 scope).
- **REQ-WC8-015** — The web layer **shall not** marshal YAML or write config files directly; all statusline persistence **shall** route through the existing write seams (HARD-8).

### C.6 Compound

- **REQ-WC8-017** — **Where** the `builder.go` `ThemeName` doc comment lists themes **When** the doc is corrected, it **shall** list exactly the two real themes (catppuccin-mocha, catppuccin-latte) and remove the non-existent default/minimal/nerd entries (SLR-5); a 1-line comment **shall** mark `SegmentRepo` (types.go:321) as intentionally outside the config schema (SLM-7).

---

## §D. 변경 표면 인벤토리 (정확히 이것만)

audit "Runtime Wiring Fixes" + "Implementation Plan (statusline)"의 9-step에 1:1 대응.

### D.1 REMOVE (config theater 제거)

| target | file:line | defect |
|--------|-----------|--------|
| `mode:` YAML surface | template statusline.yaml | SLR-1/SLM-1/SLR-2 |
| `Mode` field (CLI private struct) | internal/cli/statusline.go:127-128 + loader 137-163 + config-derived mode resolution 50-55 (line 72는 아래 D6 note 참조) | SLM-1/SLR-2 |
| `Mode` field (profile sync private struct) | internal/profile/sync.go:80-85 (`statuslineData.Mode`) + 사용 70,120-121 | SLM-2 |
| web `statusline_mode` select + parse + validation | fieldsets.templ:134, handlers.go:29/82/365, validate.go:40-41/136-137 | SLR-1 |
| `refresh_interval:` YAML | template statusline.yaml | SLM-3/SLR-6 |
| dead `loadSegmentConfig` + test refs | internal/cli/statusline.go:170-186 + coverage_improvement_test.go:1289+, statusline_test.go:105+ | SLR-7 |

> **D6-fix (line 72 구분)**: internal/cli/statusline.go의 mode 사용처 중 **50-55(env > config > default mode 계산/resolution)는 제거** 대상이나, **line 72 `opts.Mode = ...`(builder hand-off)는 Builder Config `Mode` 필드를 PRESERVE**한다(HARD-1/HARD-5) — config-derived 값 계산만 제거하고 builder 호출은 `ModeDefault` 전달(또는 mode 인자 생략). Builder Config `Mode` 필드·`SetMode`·`NormalizeMode`·`Render(data, mode)` 시그니처는 절대 삭제 금지(over-removal 방지).

### D.2 WIRE/FIX (live lever 배선 + drift 교정)

| target | file:line | defect |
|--------|-----------|--------|
| preset write-effective (비-custom → 15-key map) | internal/profile/sync.go `syncStatusline`(96-145) + 프로젝트-config write 경로 | SLR-3 |
| `defaultStatuslineSegments()` → `presetToSegments("full", nil)` | internal/profile/sync.go:148-162 | SLM-5 |
| theme seed `default`→`catppuccin-mocha` + 주석 정정 | template statusline.yaml | SLR-4 |
| `builder.go:50` ThemeName doc 정정(2 real themes) | internal/statusline/builder.go:50 | SLR-5 |
| `SegmentRepo` intentional-exclusion 주석 1줄 | internal/statusline/types.go:321 | SLM-7 |
| web 조건부 segment 노출(preset=custom일 때만 편집) | fieldsets.templ segment 블록 | SLR-3 UX |
| 카운트 라벨 `"3 fields"`→`"2 fields"` | fieldsets.templ:128 | SLW-4 |
| `StatuslineConfig` symmetryCases 추가(LAST) | internal/config/audit_struct_yaml_symmetry_test.go:27-48 | SLM-4 |

### D.3 codegen / build

- 템플릿 YAML 편집 → `make build`(go:embed embedded.go 재생성).
- `.templ` 편집 → `templ generate`(fieldsets_templ.go 갱신, zero-Node).

---

## §E. 영향받는 테스트 / sentinel (cite, do NOT re-discover)

| 테스트 | 위치 | 영향 |
|--------|------|------|
| 006 boundary sentinel | integration_test.go:197-205 | **무수정 GREEN** — statusline 미보호(in-scope). HARD-3. |
| theme-only round-trip (characterization) | integration_test.go:124-165 | preset write-effective(REQ-WC8-007) 후에도 GREEN — preset 미전송 시 preserve 경로(HARD-7). |
| "every mode collapses to 3-line / full retired" | builder_test.go:248, ~620-641 | enum + Render 시그니처 유지로 unaffected(HARD-1). |
| `loadSegmentConfig` tests | coverage_improvement_test.go:1289+, statusline_test.go:105+ | dead 함수와 함께 **삭제**(REQ-WC8-012). |
| mode YAML/parse tests | statusline.go/sync.go 관련 mode 테스트 | mode surface 제거에 맞춰 갱신(Class C source-coupled). |
| `presetToSegments` tests | update_test.go:2779+, wizard_config_test.go:410+ | 무변경(presetToSegments는 PRESERVE anchor). |
| StatuslineConfig symmetry (신규) | audit_struct_yaml_symmetry_test.go | 신규 추가(REQ-WC8-006), struct+template 대칭 확보 후 마지막에. |

---

## §F. Exclusions (What NOT to Build)

### §F.1 Out of Scope

[HARD] 본 SPEC이 명시적으로 **만들지 않는** 것:

- **git_convention config redesign → SPEC-WEB-CONSOLE-009로 deferred.** audit의 git_convention 25-defect inventory(GCM-1..9, GCW-1..7, GCR-1..9) 전부 — 009 스코프. 008은 git_convention 섹션·struct·검증·웹 위젯을 전혀 건드리지 않는다.
- **GCR-5 dormant-engine wiring** — 배포된 `.git_hooks/pre-push`에 `moai hook pre-push` 배선(hook_install.go mirror 포함)은 maintainer-decision 의존 항목으로 008·009 어느 쪽도 강제하지 않는다.
- **workflow/git-strategy/harness/llm config 섹션** — 본 SPEC은 statusline만. 이 네 섹션은 무변경(006 boundary sentinel byte-identical 유지).
- **renderFullV3 5-line mode 부활** — `mode=full`은 거짓이며 제거 대상; 5-line layout을 재구현·재호출하지 않는다(REQ-WC8-013, HARD-2).
- **canonical 모델에 Mode 필드 추가** — 3-struct 통합은 private struct에서 Mode를 **제거**하는 방향; `models.StatuslineConfig`에 Mode를 추가하지 않는다(HARD-5).
- **신규 statusline 검증기/validator 함수** — theme/preset enum의 `validate:"oneof=..."` struct 태그(SLM-6)와 segment-map server validation(SLW-4)은 **deferrable SHOULD**로만 표기; 본 SPEC MUST 스코프에 강제하지 않는다(scope creep 회피). 기존 coercion 안전망(REQ-WC8-011)으로 충분.
- **section-scoped partial-swap / `/save` fragment 응답** — 006/007에서 deferred; 008도 도입하지 않는다. 폼은 `hx-boost` full-page swap 유지.
- **SLR-8(working-tree drift) 코드 변경** — live config가 HEAD/template와 다른 working-tree drift는 코드 변경 불요; 템플릿 seed fix(REQ-WC8-005)가 source를 교정한다.

---

## §G. Cross-References

- `.moai/reports/web-console-statusline-gitconvention-audit.md` (권위 있는 연구·설계 근거 — statusline 17-defect inventory + Redesign Proposal — statusline + 9-step plan + Test & Sentinel Impact)
- SPEC-WEB-CONSOLE-006 (enabler — Templ 섹션 트리 + HTMX foundation; completed `5714bae97`)
- SPEC-WEB-CONSOLE-007 (nested config 편집 write-seam + 위젯 재사용; completed `54c09a61b`)
- SPEC-WEB-CONSOLE-009 (sibling — git_convention config redesign, 분리)
- plan.md (M1..Mn + Tier 정당화 + TDD RED→GREEN→REFACTOR)
- acceptance.md (AC-WC8-NNN, 각 GREEN-verifiable + 3 characterization/sentinel gate)
