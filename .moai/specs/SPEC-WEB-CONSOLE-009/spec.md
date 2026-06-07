---
id: SPEC-WEB-CONSOLE-009
title: "git_convention config redesign — drop custom engine, wire auto-detection + max_length, fix flat→nested schema (honest hybrid)"
version: "0.1.0"
status: draft
created: 2026-06-07
updated: 2026-06-07
author: Goos Kim
priority: P1
phase: "v0.2.0"
module: "internal"
lifecycle: spec-anchored
tags: "web-console, config-redesign, git-convention, config-theater, templ, honest-hybrid, s2-cohort"
era: V3R6
related_specs: [SPEC-WEB-CONSOLE-006, SPEC-WEB-CONSOLE-007, SPEC-WEB-CONSOLE-008]
tier: L
---

# SPEC-WEB-CONSOLE-009 — git_convention config redesign (honest hybrid)

## HISTORY

| Version | Date | Author | Summary |
|---------|------|--------|---------|
| 0.1.0 | 2026-06-07 | Goos Kim | 초안 — `.moai/reports/web-console-statusline-gitconvention-audit.md`(2026-06-07, 42 confirmed defects)의 **git_convention 25-defect inventory + Redesign Proposal — git_convention**을 권위 있는 연구·설계 근거로 사용한다. 처분 철학은 **honest hybrid**(008과 동일): (1) config theater 제거(거짓·dead 표면 삭제) — `custom` 엔진 전체(struct/enum/YAML/웹 위젯/CLI 옵션/dead `LoadFromConfig`), `validation.{enabled,enforce_on_commit}`, `formatting.*`; (2) live lever 배선 — `AutoDetectionConfig`를 `LoadConvention`에 thread(Fix A), `validation.max_length`를 `SetMaxLength`로 forward(Fix B); (3) 구조 drift 교정 — 템플릿 flat→nested 재작성(Fix C) + symmetry CI guard 추가. 2-SPEC 코호트 분할의 둘째 멤버(008=statusline, 009=git_convention). **GCR-5(dormant pre-push 엔진 — 배포된 `.git_hooks/pre-push`가 `moai hook pre-push`를 호출하지 않음)는 maintainer-decision 의존 항목으로 본 SPEC에서 분리 deferred** — §F 참조. 본 SPEC은 `custom` 엔진을 **부활시키지 않는다** — wiring은 feature이지 defect fix가 아니다(audit policy decision 1). |

---

## §A. 배경 및 Ground-Truth (cite, do NOT re-derive)

본 SPEC은 audit 보고서(`.moai/reports/web-console-statusline-gitconvention-audit.md`)의 git_convention 섹션을 권위 있는 ground-truth로 삼는다. 모든 file:line 인용은 adversarial sonnet 검증을 통과했고, **post-007/post-008 현 트리 기준으로 재검증**되었다(아래 라인은 현 트리 실측값 — audit 본문의 일부 라인은 008 이전 값이므로 본 §A의 라인을 사용). run-phase에서 재발견(re-discover)하지 말 것.

### A.1 Schema asymmetry — root defect (GCM-1/GCR-1)

- **TEMPLATE** `internal/template/templates/.moai/config/sections/git-convention.yaml`(전체 1-18줄)은 **FLAT**: `convention`, `enforce_on_push`, `auto_detect`, `max_length: 72`, `sample_size`. nested struct로 절대 unmarshal되지 않음 — 매 `moai init`이 silently-dropped config를 배포(GCM-1/GCR-1 CONFIRMED). `max_length: 72`는 잘못된 path AND 잘못된 값(런타임 nested default 100)(GCM-2).
- **LOCAL** `.moai/config/sections/git-convention.yaml`(1-23줄)은 **NESTED**(007 산출물): `auto_detection.{enabled,sample_size,confidence_threshold,fallback}`, `validation.{enabled,enforce_on_commit,enforce_on_push,max_length}`, `formatting.{show_examples,show_suggestions,verbose}`, `custom.{name,pattern,types,scopes,max_length,examples}`.

### A.2 Config 모델 (struct — DRIFT 없음)

- `models.GitConventionConfig`(pkg/models/config.go:205-221) = `{Convention, AutoDetection, Validation, Formatting, Custom}`. sub-struct: `AutoDetectionConfig`(223-229) `{Enabled, SampleSize, ConfidenceThreshold, Fallback}`, `ConventionValidationConfig`(231-237) `{Enabled, EnforceOnCommit, EnforceOnPush, MaxLength}`, `FormattingConfig`(239-244) `{ShowExamples, ShowSuggestions, Verbose}`, `CustomConventionConfig`(246-254) `{Name, Pattern, Types, Scopes, MaxLength, Examples}`.
- Convention oneof 태그(config.go:208): `validate:"omitempty,oneof=auto conventional-commits angular karma custom"` — `custom` 멤버 포함.

### A.3 Defaults (defaults.go:456-478 — DRIFT 없음)

- `NewDefaultGitConventionConfig`(456-478)는 `Convention`/`AutoDetection`/`Validation`/`Formatting`을 세팅하고 **`Custom` sub-struct는 생략**(GCM-9). 상수: `DefaultGitConvention="auto"`(61), `DefaultGitConventionSampleSize=100`(62), `DefaultGitConventionConfidenceThreshold=0.5`(63), `DefaultGitConventionFallback="conventional-commits"`(64), `DefaultGitConventionMaxLength=100`(65).

### A.4 Validator (validation.go:152-235)

- `validGitConventionNames` map(153-159)에 `"custom"` 포함. `IsValidConvention`(169-171)/`ValidConventions`(177-183)가 이 map을 reuse. `validateGitConventionConfig`(185-235)가 enum + sample_size≥0 + confidence∈[0,1] + max_length≥0 + custom-required-pattern(226-232) 검증.

### A.5 Custom-enum lockstep (4 sites — 전부 "custom" 제거 대상)

| site | file:line | 형태 |
|------|-----------|------|
| oneof 태그 | pkg/models/config.go:208 | `validate:"...oneof=auto conventional-commits angular karma custom"` |
| validator map | internal/config/validation.go:153-159 | `validGitConventionNames["custom"]=true` + 메시지(192) + custom-required block(226-232) |
| web slice | internal/web/validate.go:57 | `conventionCanonical = []string{"auto", "conventional-commits", "angular", "karma", "custom"}` |
| CLI wizard | internal/cli/profile_setup.go:507-511 | `huh.NewOption("custom", "custom")`(511) (블록 507-511) |

### A.6 Runtime (Fix A/B targets — GCM-4/5/6, GCR-3/7)

- `LoadConvention(name string) error`(internal/git/convention/manager.go:21). `name=="auto"`일 때 `Detect(m.repoPath, 100)`(manager.go:23, **100 하드코딩** — GCM-4) 후 실패 시 `ParseBuiltin("conventional-commits")`(**fallback 하드코딩** — GCR-3). `enabled`/`confidence_threshold` 게이트 없음(GCM-5/6).
- `Detect(repoPath string, sampleSize int)`(internal/git/convention/detector.go:12) — `sampleSize` 파라미터 존재하나 LoadConvention이 항상 100을 전달.
- `LoadConvention`의 유일한 production caller: `internal/cli/hook_pre_push.go:54`(`mgr.LoadConvention(convName)`).
- `isEnforceOnPushEnabled()`(internal/cli/hook_pre_push.go:121-134)는 `cfg.GitConvention.Validation.EnforceOnPush`(129)를 **정확히** 읽음(nested에 올바르게 배선). `enforce_on_push`는 그래서 schema-correct + 핸들러-wired — **유일한 dormancy는 배포된 hook이 `runPrePush` 핸들러를 호출하지 않는다는 것(= GCR-5, §F deferred)**.
- `LoadFromConfig(cfg ConventionConfig) error`(manager.go:48-54)는 **production caller 0**(GCR-4 CONFIRMED) — 호출자는 테스트만(parser_test.go, manager_test.go). `ConventionConfig`(internal/git/convention/types.go:44-52)는 convention 패키지 내부 YAML-loadable struct이며 `models.CustomConventionConfig`와 **별개**.
- `Convention`(internal/git/convention/types.go:54-63)에 `MaxLength int`(60) 필드 존재. **`SetMaxLength` setter는 부재** — Fix B에서 신규 추가(`Convention` 또는 `Manager`에 setter).
- ConfigManager `Load`(internal/config/manager.go:49)/`Reload`(:200)가 `Validate`를 호출(audit GCM-10 reject 근거 — enum은 load path에서 이미 검증됨).

### A.7 Web (internal/web — 006/007 산출물)

- `fieldsetProject`(fieldsets.templ:229-253)는 quality + git_convention을 한 "Project" fieldset에 공유. git_convention 위젯: convention `optSelect`(241, `Empty:"(project default)"` — GCW-7), confidence_threshold `numberField`(245), auto_detection.enabled `toggle`(246), custom.pattern `projectTextField`(247). 섹션 카운트 라벨 `data-i18n="count.project"` = `"8 fields"`(234).
- handlers.go: `Conventions []string`(39), `Conventions: conventionCanonical`(85), view-model `CurConfidenceThreshold`(50)/`CurAutoDetectionEnabled`(51)/`CurCustomPattern`(52) + binding(174-176, 196-202).
- projectconfig.go nested form: `projectNestedForm`(26-48) `{Confidence/ConfidenceSet, AutoEnabled/AutoEnabledSet, CustomPattern/CustomPatternSet}`, `touchesGitConvention()`(55-56) = `ConfidenceSet || AutoEnabledSet || CustomPatternSet`, `parseProjectNestedForm`(66-113) custom.pattern parse(108-110). write seam: `writeProjectConfig`(178, scalar, SetSection :199) + `writeProjectNestedConfig`(227, nested, whole-struct copy :254 + SetSection :264).
- 재사용 위젯: `optSelect`/`numberField`/`toggle`/`projectTextField`(007-added). 폼은 `hx-boost` full-page swap(006/007 계약).

### A.8 Sentinel 계약 (HARD boundary)

- `internal/web/integration_test.go:191-205`(REQ-WC-012 boundary sentinel)은 `workflow.yaml`/`harness.yaml`/`git-strategy.yaml`가 절대 write되지 않음을 단언 — **git_convention은 보호 대상이 아님(in-scope/writable)**. git_convention write는 `SetSection("git_convention", ...)` 단일 섹션만 갱신하므로 sentinel **무수정** byte-identical(GCR boundary). 본 SPEC은 이 sentinel을 변경하지 않는다.

### A.9 GCR-5 context (DOCUMENT only — §F deferred, do NOT wire)

- 배포된 `.git_hooks/pre-push`는 `make ci-local`만 실행 → `moai hook pre-push`(= `runPrePush`)를 절대 호출하지 않음 → `enforce_on_push` + 전체 convention 엔진이 shipped 프로젝트에서 end-to-end 도달 불가(GCR-5).
- hook constant `prePushHookContent`(internal/cli/hook_install.go:27-68)이 템플릿/배포 hook과 byte-match(mirror parity) — GCR-5 후속 SPEC은 이 mirror도 함께 다뤄야 함.
- **GCR-5보다 dormancy가 넓음(ground-truth)**: `git-strategy.yaml`의 `hooks.pre_push`(warn/enforce) 키가 3 site(23/48/80줄)에 존재하나 **어떤 런타임도 읽지 않음**. GCR-5 후속 SPEC scope는 `git_convention.validation.enforce_on_push` end-to-end wiring AND `git_strategy.hooks.pre_push` gap 둘 다 포괄해야 함.

### A.10 코호트 계보

SPEC-WEB-CONSOLE-006(HTMX+Templ 마이그레이션, completed `5714bae97`) → 007(quality+git_convention nested 편집, completed `54c09a61b`) → 008(statusline 재설계, completed `8178ff1a6`) → **009(본 SPEC, git_convention 재설계)**. 008과 009는 동일 audit 보고서의 두 절반(statusline/git_convention) honest-hybrid 분할.

---

## §B. HARD Invariants (must-pass — 모든 AC가 이로부터 도출됨)

| # | Invariant | Ground-truth anchor |
|---|-----------|---------------------|
| HARD-1 | **`custom` 엔진 완전 제거 — 부활 금지**: `custom` enum 멤버(4 lockstep site), `CustomConventionConfig` struct 필드, YAML `custom.*` 블록, 웹 `custom.pattern` 위젯, CLI wizard `custom` 옵션, dead `LoadFromConfig` + custom 테스트를 모두 삭제한다. `custom` wiring(feature)은 본 SPEC scope 밖(audit policy decision 1). | config.go:208/220/246-254, validation.go:153-159/226-232, web/validate.go:57, profile_setup.go:511, manager.go:48-54 |
| HARD-2 | **CRITICAL SCOPE — statusline 무관**: statusline 섹션·struct·검증·웹 위젯을 전혀 건드리지 않는다(008 스코프, completed). | audit statusline inventory (17 defects) |
| HARD-3 | **006 scope-boundary sentinel(integration_test.go:191-205) 무수정 GREEN**. git_convention은 in-scope이므로 `SetSection("git_convention", ...)` write는 허용·기대됨. workflow/harness/git-strategy는 byte-identical. 설계가 이 sentinel 변경을 강제하면 → STOP, blocker. | integration_test.go:191-205 |
| HARD-4 | **GCR-5 wiring 금지**: 배포된 `.git_hooks/pre-push`에 `moai hook pre-push` 배선(hook_install.go `prePushHookContent` mirror 포함), `git_strategy.hooks.pre_push` 배선을 본 SPEC에서 수행하지 않는다. `enforce_on_push`는 schema-correct + web-visible로 유지하되 end-to-end firing은 후속 SPEC. | .git_hooks/pre-push, hook_install.go:27-68, git-strategy.yaml:23/48/80 |
| HARD-5 | **runtime precedence/계약 보존**: `isEnforceOnPushEnabled()`(hook_pre_push.go:121-134)가 nested `Validation.EnforceOnPush`를 읽는 경로는 그대로 유지(이미 올바름). `Detect()` 시그니처는 변경 없음(`sampleSize` 파라미터 이미 존재) — Fix A는 LoadConvention이 100 대신 config 값을 전달하게 만드는 것. | hook_pre_push.go:121-134, detector.go:12 |
| HARD-6 | **`LoadConvention` 시그니처 변경 cascade 통제**: Fix A로 `LoadConvention`이 `AutoDetectionConfig`(enabled/sample_size/confidence_threshold/fallback)를 받도록 시그니처 변경 시, 유일 production caller(hook_pre_push.go:54)와 convention 패키지 테스트(manager_test.go)를 모두 새 시그니처로 갱신. Builder 외 다른 production caller 0 확인 필수. | manager.go:21, hook_pre_push.go:54, manager_test.go |
| HARD-7 | **`enforce_on_push` web-visible + schema-correct 유지**: 본 SPEC은 `validation.enforce_on_push`를 삭제하지 않는다 — 웹 toggle로 노출(GCW-3) + nested schema 유지. 삭제 대상은 `validation.{enabled,enforce_on_commit}`만. | config.go:233-235, hook_pre_push.go:129 |
| HARD-8 | **server-canonical 검증 + offline + zero-Node 보존**: 웹은 직접 YAML marshal 금지; 신규 CDN/외부 asset 0; Templ codegen은 `templ generate`만; 모든 git_convention 영속은 기존 write seam(`writeProjectConfig`/`writeProjectNestedConfig` → `SetSection`/`Save`)을 통한다. | assets.go //go:embed, projectconfig.go SetSection |
| HARD-9 | **Template-first + neutrality**: 템플릿 YAML 편집은 `internal/template/templates/.moai/config/sections/git-convention.yaml` → `make build`(go:embed embedded.go 재생성). 템플릿 주석은 SPEC ID/REQ 토큰/내부 인용 무포함(generic). | CLAUDE.local.md §2/§15/§25 |
| HARD-10 | **symmetry test 순서 게이트**: `GitConventionConfig` symmetryCases는 **template flat→nested 재작성(M4) + struct trim(M1)이 완료되어 struct↔YAML이 이미 대칭일 때 마지막에** 추가한다. 순서 위반 시 RED for the wrong reason(audit sequence note). | audit_struct_yaml_symmetry_test.go:31-57 |

---

## §C. GEARS 요구사항

### C.1 Ubiquitous (항상 활성)

- **REQ-WC9-001** — The `git_convention` config schema **shall** model exactly these live levers — `convention`(4-value enum: auto / conventional-commits / angular / karma), `auto_detection.{enabled, sample_size, confidence_threshold, fallback}`, and `validation.{enforce_on_push, max_length}` — and **shall not** model `validation.enabled`, `validation.enforce_on_commit`, `formatting.*`, or any `custom.*` field (GCM-7/GCW-6/GCR-6/GCR-8, audit Corrected Schema).
- **REQ-WC9-002** — The `models.GitConventionConfig` struct **shall** drop the `Custom`, `Formatting` fields and the `Validation.{Enabled, EnforceOnCommit}` fields; the convention `oneof` validate tag **shall** read `oneof=auto conventional-commits angular karma`(no `custom`); range/oneof tags **shall** be tightened where applicable (GCM-8/GCM-9, audit struct trim).
- **REQ-WC9-003** — The `custom` convention engine **shall** be fully removed across all four lockstep sites — the `models` oneof tag, the `internal/config/validation.go` `validGitConventionNames` map + custom-required block, the `internal/web/validate.go` `conventionCanonical` slice, and the `internal/cli/profile_setup.go` wizard option — so no surface accepts, persists, or offers `custom` (GCW-1/GCR-2, audit Custom removal).
- **REQ-WC9-004** — The dead production function `LoadConvention`'s sibling `LoadFromConfig`(internal/git/convention/manager.go:48-54) **shall not** be retained, since it has zero production callers; its test references **shall** be removed (GCR-4, audit Custom removal).
- **REQ-WC9-005** — `NewDefaultGitConventionConfig`(internal/config/defaults.go) **shall** return only the retained fields — `Convention`, `AutoDetection.{Enabled, SampleSize, ConfidenceThreshold, Fallback}`, `Validation.{EnforceOnPush, MaxLength}` — and **shall not** seed `Validation.{Enabled, EnforceOnCommit}`, `Formatting`, or `Custom` (GCM-9, audit step 2).
- **REQ-WC9-006** — The template `git-convention.yaml` **shall** be rewritten from the FLAT schema to the nested shape matching the trimmed struct (root-cause fix for the silent-drop on every `moai init`), `max_length` **shall** standardize on `100`(dropping the template's incorrect `72`), and its comments **shall** carry no SPEC ID/REQ token/internal citation (GCM-1/GCM-2/GCR-1, HARD-9, audit Fix C).
- **REQ-WC9-007** — The struct↔YAML symmetry test (`audit_struct_yaml_symmetry_test.go` `symmetryCases`) **shall** include a `GitConventionConfig` case so model↔YAML drift is CI-guarded (GCM-3); this case **shall** be added LAST, after the struct trim and template nested rewrite have already made the two symmetric (HARD-10).

### C.2 Event-driven (When)

- **REQ-WC9-008** — **When** `LoadConvention` resolves a convention with `convention == "auto"`, it **shall** honor the configured `auto_detection.enabled` flag(skip auto-detection when disabled), pass the configured `auto_detection.sample_size` to `Detect()`(instead of the hardcoded `100`), gate the detection result on `auto_detection.confidence_threshold`(fall back when below), and use the configured `auto_detection.fallback` value(instead of the hardcoded `"conventional-commits"`) — Fix A (GCM-4/GCM-5/GCM-6, GCR-3).
- **REQ-WC9-009** — **When** a convention is loaded for validation, the configured `validation.max_length` **shall** be forwarded to the loaded `Convention` via a new `SetMaxLength` setter applied after `LoadConvention`, so the documented `max_length` is enforced rather than silently overridden by the built-in's own value — Fix B (GCR-7).
- **REQ-WC9-010** — **When** a user opens the project console, the form **shall not** render the `git_convention.custom.pattern` widget; it **shall** render an `auto_detection.sample_size` number field and a `validation.enforce_on_push` toggle in addition to the existing convention select, confidence_threshold, and auto_detection.enabled controls (GCW-3/GCW-5, audit Web UX).
- **REQ-WC9-011** — **When** the `custom.pattern` widget is removed and the `sample_size` + `enforce_on_push` widgets are added, the Project section count label **shall** update from `"8 fields"` to the corrected count `"9 fields"`(dev_mode + git_convention + 3 quality + confidence_threshold + auto_detection.enabled + sample_size + enforce_on_push) (cosmetic, audit Web UX).

### C.3 State-driven (While)

- **REQ-WC9-012** — **While** the selected convention is not `auto`, the console **shall** present the auto-detection sub-fields(confidence_threshold, sample_size, enabled) as a greyed/de-emphasized client hint(no server-side cross-field logic); **while** the convention is `auto`, those sub-fields **shall** be presented as active (GCR-3 UX, audit "conditional SHOULD").

### C.4 Capability gate (Where)

- **REQ-WC9-013** — **Where** the console git_convention dropdown renders its empty option, the label **shall** read `"(unchanged)"`(not `"(project default)"`), because there is no configuration layer below the project file — empty maps to preserve (GCW-7, audit one-string fix; on-save preserve mapping is already correct and **shall** remain unchanged).

### C.5 Unwanted behavior

- **REQ-WC9-014** — The codebase **shall not** wire the deployed `.git_hooks/pre-push` hook to invoke `moai hook pre-push`, **shall not** modify the `prePushHookContent` mirror in `internal/cli/hook_install.go`, and **shall not** wire `git_strategy.hooks.pre_push` — these constitute the GCR-5 deferred boundary carrying a maintainer-decision dependency (HARD-4, §F).
- **REQ-WC9-015** — The change **shall not** touch, re-serialize, or alter any statusline config surface, struct, validator, or web widget (HARD-2, 008 scope, completed).
- **REQ-WC9-016** — The web layer **shall not** marshal YAML or write config files directly; all git_convention persistence **shall** route through the existing `writeProjectConfig` / `writeProjectNestedConfig` → `SetSection` / `Save` seams (HARD-8).
- **REQ-WC9-017** — The change **shall not** delete or alter `validation.enforce_on_push`(the one live gate) nor the `isEnforceOnPushEnabled()` read path; only `validation.{enabled, enforce_on_commit}` are removed (HARD-7/HARD-5).

### C.6 Compound

- **REQ-WC9-018** — **Where** the `LoadConvention` signature changes to accept `AutoDetectionConfig`(Fix A) **When** the cascade is propagated, the change **shall** update exactly the one production caller(`internal/cli/hook_pre_push.go:54`) and the convention-package tests(`manager_test.go`, `parser_test.go`) to the new signature, and **shall** confirm via grep that no other production `LoadConvention`/`LoadFromConfig` caller exists (HARD-6, audit Fix A cascade).

---

## §D. 변경 표면 인벤토리 (정확히 이것만)

audit "Runtime Wiring Fixes" + "Implementation Plan (git_convention)"의 1-8 step에 1:1 대응(step 9 = GCR-5 = §F deferred).

### D.1 REMOVE (config theater / dead 제거)

| target | file:line | defect |
|--------|-----------|--------|
| `custom` enum 멤버 (oneof) | pkg/models/config.go:208 | GCW-1/GCR-2 |
| `custom` validator map + 메시지 + custom-required block | internal/config/validation.go:153-159/192/226-232 | GCW-1/GCR-2 |
| `custom` web slice 멤버 | internal/web/validate.go:57 | GCW-1 |
| `custom` CLI wizard 옵션 | internal/cli/profile_setup.go:511 | GCW-1 |
| `CustomConventionConfig` struct + `Custom` 필드 | pkg/models/config.go:220, 246-254 | GCM-8 |
| `FormattingConfig` struct + `Formatting` 필드 | pkg/models/config.go:217, 239-244 | GCM-7/GCR-8 |
| `Validation.{Enabled, EnforceOnCommit}` 필드 | pkg/models/config.go:233-234 | GCW-6/GCR-6 |
| defaults: `Validation.{Enabled, EnforceOnCommit}`, `Formatting` seed 제거 | internal/config/defaults.go:467-468, 472-476 | GCM-9 |
| dead `LoadFromConfig` + custom 테스트 refs | internal/git/convention/manager.go:48-54 + parser_test.go/manager_test.go | GCR-4 |
| web `custom.pattern` 위젯 + view-model + nested-form parse | fieldsets.templ:247, handlers.go:52/176/202, projectconfig.go:42-43/108-110/168/261-262 (CustomPattern 경로) | GCR-9 |
| YAML `custom.*`, `validation.{enabled,enforce_on_commit}`, `formatting.*` 블록 | template git-convention.yaml(재작성으로 자연 제거) + local YAML(재작성) | GCM-7/GCW-6 |

### D.2 WIRE/FIX (live lever 배선 + drift 교정)

| target | file:line | defect |
|--------|-----------|--------|
| **Fix A**: `LoadConvention`이 `AutoDetectionConfig`를 honor (enabled/sample_size→Detect/confidence_threshold gate/fallback) | internal/git/convention/manager.go:21-45 + caller hook_pre_push.go:54 | GCM-4/5/6, GCR-3 |
| **Fix B**: `SetMaxLength` setter 신규 + LoadConvention 후 적용 | internal/git/convention/manager.go(or convention.go) + hook_pre_push.go | GCR-7 |
| **Fix C**: 템플릿 flat→nested 재작성 + `max_length`=100 + 주석 generic | template git-convention.yaml | GCM-1/GCM-2/GCR-1 |
| local YAML nested 재작성(trimmed shape) | .moai/config/sections/git-convention.yaml | GCM-1 (live align) |
| web `sample_size` number field 추가 | fieldsets.templ + handlers.go view-model + projectconfig.go nested-form | GCW-5 |
| web `enforce_on_push` toggle 추가 | fieldsets.templ + handlers.go view-model + projectconfig.go nested-form | GCW-3 |
| 카운트 라벨 `"8 fields"`→`"9 fields"` | fieldsets.templ:234 | GCW UX |
| 빈 옵션 라벨 `"(project default)"`→`"(unchanged)"` (git_convention select만) | fieldsets.templ:241 | GCW-7 |
| 조건부 detection sub-field grey (convention≠auto) | fieldsets.templ git_convention 블록 | GCR-3 UX |
| `GitConventionConfig` symmetryCases 추가(LAST) | internal/config/audit_struct_yaml_symmetry_test.go:31-57 | GCM-3 |

### D.3 codegen / build

- 템플릿 YAML 편집 → `make build`(go:embed embedded.go 재생성).
- `.templ` 편집 → `templ generate`(fieldsets_templ.go 갱신, zero-Node).

---

## §E. 영향받는 테스트 / sentinel (cite, do NOT re-discover)

| 테스트 | 위치 | 영향 |
|--------|------|------|
| 006 boundary sentinel | integration_test.go:191-205 | **무수정 GREEN** — git_convention 미보호(in-scope). HARD-3. |
| `TestNewDefaultGitConventionConfig` | internal/config/defaults_test.go:570+ (Validation.Enabled/EnforceOnCommit/Formatting 단언) | struct trim에 맞춰 **갱신**(fail-loud 신호) — 제거된 필드 단언 삭제. |
| `LoadFromConfig` / custom 테스트 | internal/git/convention/parser_test.go, manager_test.go | dead 함수·custom 경로와 함께 **삭제/갱신**(Class C source-coupled). |
| `LoadConvention` 테스트 | internal/git/convention/manager_test.go | Fix A 신규 시그니처로 **갱신**(AutoDetectionConfig 전달). |
| `GitConventionConfig` symmetry (신규) | audit_struct_yaml_symmetry_test.go | 신규 추가(REQ-WC9-007), struct trim + 템플릿 nested 재작성 후 마지막에. |
| validator custom 테스트 | internal/config/validation_test.go (custom enum/custom-required) | custom 제거에 맞춰 갱신. |
| web characterization (nested round-trip) | internal/web/integration_test.go (git_convention nested round-trip) | sample_size/enforce_on_push 추가 + custom.pattern 제거 반영하여 갱신; nested isolation 보존. |
| Template neutrality | internal/template/internal_content_leak_test.go + TestTemplateNeutralityAudit | 재작성 YAML 주석이 SPEC/REQ 토큰 0 — 통과 유지. |

---

## §F. Exclusions (What NOT to Build)

### §F.1 Out of Scope

[HARD] 본 SPEC이 명시적으로 **만들지 않는** 것:

- **GCR-5 dormant pre-push engine wiring → 별도 후속 SPEC으로 deferred.** 배포된 `.git_hooks/pre-push`(현재 `make ci-local`만 실행)에 `moai hook pre-push`(`runPrePush`) 호출 배선, `internal/cli/hook_install.go:27-68` `prePushHookContent` mirror 갱신을 본 SPEC에서 수행하지 않는다. 이는 hook placement maintainer-decision 의존 항목이다.
  - **후속 SPEC scope 메모(ground-truth 확장)**: dormancy는 GCR-5 단독보다 넓다 — `git-strategy.yaml`의 `hooks.pre_push`(warn/enforce) 키가 3 site(23/48/80줄)에 존재하나 어떤 런타임도 읽지 않는다. GCR-5 후속 SPEC scope는 **(a) `git_convention.validation.enforce_on_push` end-to-end wiring**(deployed hook → `runPrePush` → `isEnforceOnPushEnabled` 경로 완성, hook_install.go mirror 포함) **AND (b) `git_strategy.hooks.pre_push` gap** 둘 다 포괄해야 한다.
  - 009는 `enforce_on_push`를 **schema-correct + web-visible**로 유지한다(REQ-WC9-010, HARD-7). 핸들러 `isEnforceOnPushEnabled()`는 이미 nested 값을 올바르게 읽는다(GCR-5의 유일한 dormancy는 배포된 hook이 핸들러를 호출하지 않는다는 것). 009는 hook 배선만 건드리지 않는다.
- **`custom` convention 엔진 wiring/부활** — `custom`은 production caller 0인 dead 표면이며 `LoadConvention`으로 도달 불가하다. wiring은 feature이지 defect fix가 아니다(audit policy decision 1). 009는 `custom`을 **제거**할 뿐 wiring하지 않는다(REQ-WC9-003, HARD-1).
- **statusline config redesign** — 008 스코프(completed `8178ff1a6`). statusline 섹션·struct·검증·웹 위젯 무변경(HARD-2).
- **workflow/git-strategy/harness/llm config 섹션** — 본 SPEC은 git_convention만. 이 네 섹션은 무변경(006 boundary sentinel byte-identical 유지, HARD-3). 단, GCR-5 후속 메모로 `git-strategy.hooks.pre_push` gap을 **기록**만 한다(읽기 전용 — 코드 변경 0).
- **`validation.enforce_on_push` 삭제** — 이것은 유일한 live gate이며 웹 노출 대상(GCW-3). 삭제 대상은 `validation.{enabled, enforce_on_commit}`만(REQ-WC9-001, HARD-7).
- **새로운 git_convention validator 함수 신설** — 본 SPEC은 기존 `validateGitConventionConfig`를 **trim**한다(custom-required block + custom enum 멤버 제거). 신규 validator 함수(`validateWorkflowConfig` 류)를 생성하지 않는다(scope creep 회피).
- **section-scoped partial-swap / `/save` fragment 응답** — 006/007/008에서 deferred; 009도 도입하지 않는다. 폼은 `hx-boost` full-page swap 유지.
- **`Detect()` 시그니처 변경** — `Detect(repoPath, sampleSize)`는 이미 `sampleSize` 파라미터를 받는다. Fix A는 LoadConvention이 100 대신 config 값을 **전달**하게 만드는 것이지 Detect 시그니처를 바꾸는 것이 아니다(HARD-5).

---

## §G. Cross-References

- `.moai/reports/web-console-statusline-gitconvention-audit.md` (권위 있는 연구·설계 근거 — git_convention 25-defect inventory + Redesign Proposal — git_convention + 8-step Implementation Plan + Test & Sentinel Impact. **본 audit 보고서가 Tier L의 design.md/research.md 역할을 대체** — plan.md §F 참조)
- SPEC-WEB-CONSOLE-006 (enabler — Templ 섹션 트리 + HTMX foundation; completed `5714bae97`)
- SPEC-WEB-CONSOLE-007 (nested config 편집 write-seam(`writeProjectNestedConfig`) + 위젯 재사용; completed `54c09a61b`)
- SPEC-WEB-CONSOLE-008 (sibling — statusline config redesign, honest hybrid; completed `8178ff1a6`)
- plan.md (M1..Mn + Tier 정당화 + TDD RED→GREEN→REFACTOR + 순서 게이트)
- acceptance.md (AC-WC9-NNN, 각 GREEN-verifiable + characterization/sentinel gate)
