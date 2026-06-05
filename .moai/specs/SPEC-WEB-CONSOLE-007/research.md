# SPEC-WEB-CONSOLE-007 — Research (Ground-Truth Citations)

Explore 실측 2026-06-06. 모든 인용은 file:line. **재도출 금지 — 본 파일이 단일 진실 출처.**

---

## 1. Config 모델 (검증 anchor)

### 1.1 QualityConfig
- `pkg/models/config.go:49` — `type QualityConfig struct` (nested: line 50 DevelopmentMode, 51 EnforceQuality, 52 TestCoverageTarget, 57 DDDSettings, 58 TDDSettings, 59 CoverageExemptions, 60 TestQuality, 61 LSPQualityGates, 62 Principles).
- `pkg/models/config.go:146-152` — `type TDDSettings struct` (line 150 `MinCoveragePerCommit int`).
- `pkg/models/config.go:154-159` — `type CoverageExemptions struct` (line 158 `MaxExemptPercentage int`).
- 검증기 `internal/config/validation.go:96-127` `validateQualityConfig`:
  - line 99-106: `test_coverage_target` < 0 || > 100 → error "must be between 0 and 100".
  - line 108-115: `quality.tdd_settings.min_coverage_per_commit` < 0 || > 100.
  - line 117-124: `quality.coverage_exemptions.max_exempt_percentage` < 0 || > 100.

### 1.2 GitConventionConfig
- `pkg/models/config.go:206-221` — `type GitConventionConfig struct` (208 Convention, 211 AutoDetection, 214 Validation, 217 Formatting, 220 Custom).
- `pkg/models/config.go:223-229` — `type AutoDetectionConfig struct` (225 Enabled bool, 226 SampleSize int, 227 ConfidenceThreshold float64, 228 Fallback string).
- `pkg/models/config.go:231-237` — `type ConventionValidationConfig struct` (235 EnforceOnPush, 236 MaxLength).
- `pkg/models/config.go:246-254` — `type CustomConventionConfig struct` (249 Pattern string).
- 검증기 `internal/config/validation.go:163-212` `validateGitConventionConfig`:
  - line 166-173: `git_convention.convention` enum (auto/conventional-commits/angular/karma/custom).
  - line 175-182: `auto_detection.sample_size` < 0 → "must be non-negative".
  - line 184-191: `auto_detection.confidence_threshold` < 0 || > 1 → "must be between 0.0 and 1.0".
  - line 193-200: `validation.max_length` < 0 → "must be non-negative".
  - line 202-209: `convention == "custom" && custom.pattern == ""` → "pattern is required when convention is 'custom'".
- 예측자 `internal/config/validation.go:146-160`: `IsValidConvention(name)` + `ValidConventions()` (canonical 5 enum SSOT, empty 비포함).

---

## 2. Web foundation (internal/web)

### 2.1 Templ 섹션 합성 (static registry)
- `internal/web/root.templ:72-121` — `templ page(view pageView)`; line 105-109 = 5 fieldset 수동 합성 `@fieldsetIdentity/Language/Launch/Statusline/Project`.
- `internal/web/root.templ:103` — `<form ... hx-boost="true">` (HTMX foundation; full-page swap, NO partial-swap attr — 006 주석 line 95-102 명시).
- `internal/web/fieldsets.templ:209-226` — `templ fieldsetProject(view)` = 2x `optSelect`(development_mode, git_convention).

### 2.2 재사용 위젯
- `internal/web/page.templ:49-69` — `templ langSelect(a langSelectArgs)`.
- `internal/web/page.templ:90-110` — `templ optSelect(a optSelectArgs)` (field chrome 패턴: field/field__label/field__title data-i18n/field__key/field__desc/select-wrap/select/has-error/field-error span). 신규 위젯이 재사용할 chrome 원천.
- `internal/web/page.templ:114-127` — `templ optOptionTags(value, empty, options)`.
- `internal/web/fieldsets.templ:104-118` — `templ permissionOption(value, current)`.
- `internal/web/fieldsets.templ:157-195` — `templ statuslineSelect(...)` + `statuslineSelectInner`.
- `internal/web/fieldsets.templ:198-207` — `templ segmentCheckbox(key, checked)` (checkbox 패턴: `<input type="checkbox" name= value="1" checked/>` — 신규 `toggle` 위젯의 checkbox 원천).
- `internal/web/icons.templ` — `icon(name)` 닫힌 switch (가용: user-round, languages, rocket, panel-bottom, folder-git, sun, moon, save, check, check-circle, alert-circle, chevron-down, x). 신규 위젯의 에러 아이콘은 `alert-circle` 재사용.
- **부재 확인**: boolean toggle 위젯 없음, number-input 위젯 없음 (page.templ/fieldsets.templ grep 결과 `<input type="number"` 0개, generic toggle 0개). → 본 SPEC이 `toggle` + `numberField` 추가.

### 2.3 Project-config write seam (load-modify-write 격리 메커니즘)
- `internal/web/projectconfig.go:22-29` — `readProjectConfig` (LoadRaw → 현재 devMode + convention; 부재 시 empty, EC-5).
- `internal/web/projectconfig.go:37-70` — `writeProjectConfig(projectRoot, devMode, convention)`:
  - line 39: `cfg, _ := mgr.LoadRaw(projectRoot)`.
  - line 46-53: devMode 변경 시 `quality := cfg.Quality`(전체 구조체 복사) → `quality.DevelopmentMode = ...` → `SetSection("quality", quality)`.
  - line 55-62: convention 변경 시 `gc := cfg.GitConvention`(전체 복사) → `gc.Convention = ...` → `SetSection("git_convention", gc)`.
  - line 64-68: `if changed { mgr.Save() }`.
- `internal/config/manager.go:138-147` — `SetSection(name, value)` (전체 섹션 구조체를 in-memory 교체).
- `internal/config/manager.go:155-196` — `Save()` (각 섹션을 wrapper로 전체 직렬화, temp+rename atomic). **격리 crux**: SetSection이 구조체 전체를 교체하므로, LoadRaw로 가져온 전체 구조체에서 **타깃 필드만 변경**하면 폼에 없는 sibling nested 필드가 LoadRaw값 그대로 라이드 → byte-identical 보존.

### 2.4 Validation entry + EC
- `internal/web/validate.go:101-147` — `validatePrefs(p ProfilePreferences)` (프로필 검증, 기존 예측자 재사용, empty allowed).
- `internal/web/validate.go:162-173` — `validateProjectConfig(devMode, convention)` (별도 검증기, `models.DevelopmentMode.IsValid` + `config.IsValidConvention` 재사용, empty allowed, exact match). 본 SPEC이 nested 필드로 확장.
- `internal/web/handlers.go:16-55` — `pageView` (line 38-41 DevelopmentModes/Conventions/CurDevelopmentMode/CurConvention 옵션+현재값; line 54 FieldErrors map). 신규 nested 옵션/현재값/에러가 여기에 추가됨.
- `internal/web/handlers.go:178-197` — handleSave: line 180-181 devMode/convention bind, line 186-189 두 검증기 병합, line 190-197 **atomic reject**(한 필드라도 에러면 view 재렌더 후 return, write 전). EC-2 원천.
- EC-1(empty=보존): projectconfig.go:46,55 (`!= ""` 가드).

### 2.5 Scope boundary sentinel (HARD-2 — 무수정 유지)
- `internal/web/integration_test.go:191-205` — `for _, name := range []string{"workflow.yaml", "harness.yaml", "git-strategy.yaml"}` → `DO_NOT_TOUCH` sentinel 존재 단언.
- **핵심**: sentinel 집합은 workflow/harness/git-strategy **3개만**. quality/git-convention은 **sentinel에 없음**(in-scope, Save()가 소유). 따라서 quality/git_convention nested 쓰기 확장은 이 sentinel을 건드리지 않으며 무수정 GREEN 유지가 자연 성립한다.

### 2.6 Offline / zero Node
- `internal/web/assets.go` — `//go:embed` (htmx.min.js + self-host fonts, CDN 0). 신규 위젯은 marshal/CDN 불요.
- `internal/web/templ.go` + `*_templ.go` — `templ generate` codegen(pure-Go, zero Node). 신규 위젯은 .templ 추가 → `templ generate` → `*_templ.go` 재생성.

### 2.7 Class A markup-parity 테스트 패턴
- `internal/web/templ_helpers_test.go:28-87` — `TestLangSelectHelperMarkupParity` (clean + errored 렌더, `strings.Contains` 단언, aria-invalid/field-error 음성 단언).
- `internal/web/templ_helpers_test.go:89-123` — `TestOptSelectHelperMarkupParity`. 신규 `toggle`/`numberField` 위젯의 Class A 테스트가 이 패턴을 그대로 따른다.
- `renderTempl(t, c)` helper (line 14-21) — 신규 위젯 테스트가 재사용.

---

## 3. 코호트 deferred 기록 (007 = S2b)
- `.moai/specs/SPEC-WEB-CONSOLE-006/progress.md:21` — `deferred_to: SPEC-WEB-CONSOLE-007 # S2b — nested config-section editing (006 is the enabler)`.
- `.moai/specs/SPEC-WEB-CONSOLE-006/progress.md:173` — `deferred_to: SPEC-WEB-CONSOLE-007 # section-scoped partial-swap + nested config-section editing`.
- `.moai/specs/SPEC-WEB-CONSOLE-006/progress.md:212` — `deferred_to: SPEC-WEB-CONSOLE-007 # S2b ... 006 component tree + HTMX foundation = enabler`.

---

## 4. 008 deferred 근거 (007 비-목표)
- workflow/git-strategy validator 부재: `grep -n "func validateWorkflowConfig\|func validateGitStrategyConfig" internal/config/validation.go` → 0건(현재 0 validator). 008은 신규 validator 함수 생성 필요 → 007 CRITICAL SCOPE CONSTRAINT 위반이므로 deferred.
- integration_test.go:197-205 sentinel이 workflow/harness/git-strategy를 byte-identical로 단언 → 008은 이 sentinel retarget(Class C) 필요 → 007 HARD-2 위반.
