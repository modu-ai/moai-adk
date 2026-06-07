---
id: SPEC-WEB-CONSOLE-007
title: "moai web console nested config editing — quality + git_convention (S2b, scoped)"
version: "0.1.0"
status: in-progress
created: 2026-06-06
updated: 2026-06-07
author: Goos Kim
priority: P1
phase: "v0.2.0"
module: "internal/web"
lifecycle: spec-anchored
tags: "web-console, config-editing, nested-config, quality, git-convention, templ, htmx, s2b"
era: V3R6
related_specs: [SPEC-WEB-CONSOLE-006, SPEC-WEB-CONSOLE-003]
depends_on: [SPEC-WEB-CONSOLE-006]
---

# SPEC-WEB-CONSOLE-007 — moai web 콘솔 nested config 편집 (quality + git_convention, S2b 스코프 한정)

## HISTORY

| Version | Date | Author | Summary |
|---------|------|--------|---------|
| 0.1.0 | 2026-06-06 | Goos Kim | 초안 — web-console-v4 코호트 "S2b" 멤버. SPEC-WEB-CONSOLE-006(html/template+vanilla → HTMX+Templ 마이그레이션)이 깐 **Templ 섹션 컴포넌트 트리 + HTMX foundation**을 enabler로 활용해, `moai web` config 콘솔의 편집 깊이를 **정확히 두 섹션(`quality` + `git_convention`)의 curated nested 필드 집합**으로 확장한다. 오늘 콘솔은 `writeProjectConfig`(projectconfig.go:37-70)를 통해 두 스칼라 필드(`quality.development_mode`, `git_convention.convention`)만 쓴다. 본 SPEC은 그 두 섹션의 **검증된** nested 필드를 추가로 편집 가능하게 만든다. **CRITICAL SCOPE CONSTRAINT**: 신규 편집 필드는 기존 두 섹션 검증기(`validateQualityConfig` / `validateGitConventionConfig`)를 **확장**해 검증되는 필드로만 한정한다 — 다른 어떤 섹션을 위한 신규 validator 함수도 만들지 않는다. workflow/git-strategy/harness/llm nested 편집은 SPEC-WEB-CONSOLE-008로 명시적 deferred. |

---

## §A. 배경 및 Ground-Truth (cite, do NOT re-derive)

본 SPEC은 Explore 실측(2026-06-06)으로 확정된 ground-truth 위에서 작성된다. 모든 file:line 인용은 research.md에 집약되어 있으며, 아래는 핵심 seam 요약이다.

### A.1 Config 모델 (검증 anchor)
- `quality` → `QualityConfig` (pkg/models/config.go:49), nested: `DDDSettings`/`TDDSettings`/`CoverageExemptions`/`TestQuality`/`LSPQualityGates`/`Principles` 등.
  - 검증기 `validateQualityConfig` (internal/config/validation.go:96-127): `test_coverage_target` 0-100, `tdd_settings.min_coverage_per_commit` 0-100, `coverage_exemptions.max_exempt_percentage` 0-100.
- `git_convention` → `GitConventionConfig` (pkg/models/config.go:206), nested: `AutoDetectionConfig`/`ConventionValidationConfig`/`FormattingConfig`/`CustomConventionConfig`.
  - 검증기 `validateGitConventionConfig` (validation.go:163-212): `convention` enum, `auto_detection.sample_size` ≥ 0, `auto_detection.confidence_threshold` [0.0,1.0], `validation.max_length` ≥ 0, `convention=="custom"`일 때 `custom.pattern` required.

### A.2 Web foundation (internal/web — 006 산출물)
- **STATIC Templ 섹션 레지스트리**: 5 fieldset이 root.templ:105-109에서 수동으로 합성됨(`@fieldsetIdentity`/`Language`/`Launch`/`Statusline`/`Project`). 동적 레지스트리 없음 — 신규 필드 그룹 = root.templ에 추가/확장되는 Templ 컴포넌트.
- **재사용 위젯**: `langSelect`(page.templ:49), `optSelect`(page.templ:90), `optOptionTags`/`langOptionTags`, `statuslineSelect`(fieldsets.templ:157), `permissionOption`(fieldsets.templ:104), `segmentCheckbox`(fieldsets.templ:198), `icon`(icons.templ). **boolean toggle 위젯 없음 + number-input 위젯 없음** — quality nested 필드는 boolean과 int를 포함하므로 본 SPEC은 재사용 가능한 `toggle` + `numberField` Templ 위젯을 **추가**해야 한다(각각 Class A markup-parity 테스트 1개 동반).
- **Project-config write seam**: `writeProjectConfig`(projectconfig.go:37-70) = config manager `LoadRaw` + `SetSection("quality"/"git_convention")` + `Save`; 현재 `DevelopmentMode` + `Convention`만 변경. SetSection은 **섹션 구조체 전체를 교체**하고 Save는 구조체 전체를 직렬화하므로(manager.go:138-196), 확장 시 반드시 **load-modify-write**(LoadRaw로 가져온 전체 섹션 구조체에서 폼이 담은 nested 필드만 변경) 해야 폼에 없는 sibling nested 필드가 byte-identical 보존된다(섹션-레벨보다 더 미세함).
- **Validation entry**: `validatePrefs`(validate.go:101-147, 프로필) + `validateProjectConfig`(validate.go:162-173, 프로젝트). EC-1(empty=보존) projectconfig.go:46,55. EC-2(atomic reject — 한 필드라도 에러면 어떤 write도 하기 전에 return) handlers.go:190.
- **Offline**: `//go:embed`(assets.go) htmx.min.js + self-host fonts, zero CDN. `pageView`(handlers.go:16-55)가 옵션 리스트 + `FieldErrors`를 운반; 에러는 `hasFieldError`/`fieldErrorMsg`로.

### A.3 코호트 계보
SPEC-WEB-CONSOLE-001..005(v3 코호트: i18n/restyle/scalar-config) → 006(HTMX+Templ 마이그레이션, completed v0.2.0, origin/main `5714bae97`) → **007(본 SPEC, S2b nested 편집)**. 006은 자신의 progress.md(라인 21/173/212)에서 nested config 편집과 section-scoped partial-swap을 명시적으로 007로 deferred했고, 006의 Templ 섹션 트리 + HTMX foundation이 007의 enabler임을 기록했다.

---

## §B. HARD Invariants (must-pass — 모든 AC가 이로부터 도출됨)

| # | Invariant | Ground-truth anchor |
|---|-----------|---------------------|
| HARD-1 | REQ-WC-012 boundary UNCHANGED: 콘솔은 workflow/harness/git-strategy/llm 또는 quality+git_convention(+기존 프로필 섹션) 외 어떤 섹션도 쓰지 않는다. | handlers.go:156-159 |
| HARD-2 | 006 scope-boundary sentinel(integration_test.go:197-205)이 **무수정 GREEN** 유지. workflow/harness/git-strategy가 byte-identical(`DO_NOT_TOUCH`)임을 단언. 만약 설계가 이 테스트 변경을 강제하면 → STOP, blocker(= 008 스코프 침범). | integration_test.go:197-205 |
| HARD-3 | Server-canonical 검증: 신규 필드는 `validateQualityConfig`/`validateGitConventionConfig` 확장(+`validateProjectConfig` 배선)으로 server-side 검증. internal/web 패키지 내 직접 YAML marshal 금지. 옵션은 server-rendered `name=`. 에러는 `FieldErrors`. | validation.go:96-212, validate.go:162-173 |
| HARD-4 | NESTED 섹션 격리: quality/git_convention의 한 nested 필드를 쓸 때 **그 섹션의 다른 모든 nested 필드가 byte-identical 보존**(load-modify-write, 폼에서 섹션 전체 교체 금지). sibling nested 필드 생존을 증명하는 테스트 필수. | manager.go:138-196, projectconfig.go:37-70 |
| HARD-5 | EC-1(empty=보존) + EC-2(atomic reject)를 신규 nested 필드에 확장. | projectconfig.go:46,55 / handlers.go:190 |
| HARD-6 | Offline zero-network 보존; 신규 CDN/외부 asset 없음. | assets.go //go:embed |
| HARD-7 | profile vs project scope 분리 보존(두 대상 섹션 모두 project-scope). | handlers.go:178-181, validate.go |
| HARD-8 | zero Node build — Templ `templ generate` codegen만; 신규 위젯은 templ로 컴파일; Vite/npm 없음. | templ.go, fieldsets_templ.go |

---

## §C. GEARS 요구사항

### C.1 Ubiquitous (항상 활성)

- **REQ-WC7-001**: The `moai web` console **shall** persist the curated nested fields of EXACTLY two project-config sections (`quality`, `git_convention`) and no other section beyond the existing profile + project scope (HARD-1).
- **REQ-WC7-002**: The console **shall** validate every newly editable nested field server-side by EXTENDING `validateQualityConfig` and `validateGitConventionConfig` only, and **shall not** introduce a new validator function for any other config section (HARD-3, CRITICAL SCOPE CONSTRAINT).
- **REQ-WC7-003**: The web layer **shall not** marshal YAML or write config files directly; all project-config persistence **shall** route through the config-manager (`LoadRaw` → mutate non-empty → `SetSection` → `Save`) write seam (HARD-3).
- **REQ-WC7-004**: The console **shall** reuse server-rendered option lists with `name=` attributes carried on `pageView`, and **shall** surface validation failures through the existing `FieldErrors` / `hasFieldError` / `fieldErrorMsg` mechanism — no new error channel (HARD-3).

### C.2 Event-driven (When)

- **REQ-WC7-005**: **When** a user submits the form with a changed value for a curated nested field, the console **shall** write only that nested field into its section via load-modify-write, leaving every other nested field of the same section byte-identical (HARD-4).
- **REQ-WC7-006**: **When** a user submits the form with an empty value for a curated nested field, the console **shall** leave the persisted nested value unchanged (EC-1 extended, HARD-5).
- **REQ-WC7-007**: **When** any submitted curated field fails validation, the console **shall** reject the entire submission atomically (no section written) and re-render the form with per-field errors (EC-2 extended, HARD-5).
- **REQ-WC7-008**: **When** `git_convention.convention` is `custom` and `git_convention.custom.pattern` is empty on submit, the console **shall** reject the submission with a `git_convention.custom.pattern` field error (reuses the existing custom-required rule, validation.go:202-209).

### C.3 State-driven (While)

- **REQ-WC7-009**: **While** htmx is loaded, the form submission **shall** remain an `hx-boost` full-page swap returning the SAME full HTML page (006 REQ-WC6-014/015 contract preserved) — **no** section-scoped partial-swap fragment is introduced (that remains deferred per 006).

### C.4 Capability gate (Where)

- **REQ-WC7-010**: **Where** the project has a `quality` section on disk, the console **shall** render the curated quality nested fields populated from the persisted section read seam; **where** absent, fields render neutral/empty defaults (EC-5 read-seam default, projectconfig.go:22-29 pattern).

### C.5 Unwanted behavior

- **REQ-WC7-011**: The console **shall not** write, touch, or re-serialize the `workflow`, `harness`, `git-strategy`, or `llm` config sections, keeping the 006 scope-boundary sentinel byte-identical (HARD-1, HARD-2).
- **REQ-WC7-012**: The console **shall not** replace a whole section struct from form data; it **shall not** drop or zero any nested field absent from the submitted form (HARD-4).
- **REQ-WC7-013**: The console **shall not** introduce any new CDN, external asset, or Node/Vite build step; new widgets **shall** compile via `templ generate` only (HARD-6, HARD-8).

### C.6 Compound

- **REQ-WC7-014**: **Where** a curated nested field is an integer with an existing range rule (e.g., `quality.test_coverage_target` 0-100) **When** a user submits an out-of-range value, the console **shall** reject the submission and re-render with the existing range message from the extended `validateQualityConfig` (HARD-3, HARD-5).

---

## §D. 신규 위젯 (Templ)

REQ-WC7-004의 server-rendered `name=` 계약을 따르되, 기존 `optSelect` chrome 패턴(field/field__label/field__title/field__key/field__desc/has-error/field-error span)을 재사용한다:

- **`toggle`**: boolean nested 필드용 체크박스 위젯. `segmentCheckbox`(fieldsets.templ:198)의 checkbox 패턴 + `optSelect`의 field chrome을 합성. checked/unchecked 두 상태, has-error span 지원.
- **`numberField`**: int/float nested 필드용 `<input type="number">` 위젯. `optSelect` chrome + `<input type="number" name= value= min= max= step=>`. 에러 시 `aria-invalid` + field-error span.

각 위젯은 Class A markup-parity 테스트 1개를 가진다(templ_helpers_test.go의 `TestLangSelectHelperMarkupParity`/`TestOptSelectHelperMarkupParity` 패턴 그대로).

---

## §E. Curated 편집 필드 인벤토리 (정확히 이것만)

HARD-3 제약하에, 기존 검증 anchor를 가진(또는 그것을 확장하는) 고가치 필드만 선정. 각 필드는 §F의 per-field validation map에 매핑된다.

### E.1 `quality` 섹션 (3 필드)

| # | dot-path | 타입 | 위젯 | 검증 (확장 위치) |
|---|----------|------|------|-----------------|
| Q1 | `quality.test_coverage_target` | int | numberField | 기존 0-100 (validation.go:99-106) — 추가 규칙 불요, 폼 배선만 |
| Q2 | `quality.enforce_quality` | bool | toggle | bool, 검증 무관(불리언은 누락/존재 양값만) — `validateQualityConfig`에 명시적 no-op 주석 |
| Q3 | `quality.tdd_settings.min_coverage_per_commit` | int | numberField | 기존 0-100 (validation.go:108-115) — 추가 규칙 불요, 폼 배선만 |

### E.2 `git_convention` 섹션 (3 필드 + 기존 convention 스칼라 유지)

| # | dot-path | 타입 | 위젯 | 검증 (확장 위치) |
|---|----------|------|------|-----------------|
| G0 | `git_convention.convention` | enum | optSelect (기존) | 기존 enum (validation.go:166-173) — 006/003에서 이미 편집 가능, 유지 |
| G1 | `git_convention.auto_detection.confidence_threshold` | float [0,1] | numberField (step=0.01) | 기존 [0.0,1.0] (validation.go:184-191) — 추가 규칙 불요, 폼 배선만 |
| G2 | `git_convention.auto_detection.enabled` | bool | toggle | bool, no-op 검증 |
| G3 | `git_convention.custom.pattern` | string | text input (기존 input 패턴) | 기존 custom-required (validation.go:202-209) — `convention=="custom"`일 때 required, 폼 배선만 |

**선정 근거**: 7개 편집 필드(스칼라 G0 포함) 전부 기존 검증 anchor에 직접 대응하거나(Q1/Q3/G1/G3) 검증 무관 불리언(Q2/G2)이다. **신규 range/enum 규칙을 새로 만들지 않으며**, 신규 validator 함수도 만들지 않는다 — 두 기존 검증기에 dot-path 배선만 추가한다. 인벤토리는 Tier-M-sized로 한정(섹션당 3 nested + 스칼라 1).

**명시적 제외**: `coverage_exemptions.max_exempt_percentage`(기존 검증 anchor 있으나 인벤토리 bloat 회피로 제외), LSPQualityGates/TestQuality/Principles/DDDSettings/Formatting/CustomConventionConfig의 나머지 필드, `auto_detection.sample_size`/`validation.max_length`(≥0 규칙 있으나 인벤토리 한정으로 제외). 추가 노출은 후속 SPEC.

---

## §F. Exclusions (What NOT to Build)

### §F.1 Out of Scope

[HARD] 본 SPEC이 명시적으로 **만들지 않는** 것:

- **workflow/git-strategy/harness/llm nested 편집 → SPEC-WEB-CONSOLE-008로 deferred.** 이는 REQ-WC-012 boundary **LIFT** + workflow & git-strategy를 위한 **신규 validator 함수**(현재 0개) + integration_test.go:197-205 sentinel **retargeting**(Class C 변경)을 요구한다. 007은 이 중 어느 것도 하지 않는다.
- **section-scoped partial-swap / `/save` fragment 응답** — 006에서도 deferred했고 007도 도입하지 않는다. 폼은 `hx-boost` full-page swap을 유지한다(REQ-WC7-009).
- **동적 섹션 레지스트리** — 섹션 합성은 root.templ에서 수동(static)으로 유지. 동적 레지스트리는 비-목표.
- **신규 config 섹션, 신규 settings.json 필드, 신규 validator 함수** — 두 기존 검증기 확장만.
- **E.2 명시적 제외 필드들** — `coverage_exemptions.*`, LSP/TestQuality/Principles/DDD/Formatting의 나머지, `auto_detection.sample_size`, `validation.max_length`.
- **profile-scope 변경** — 프로필 섹션(user/language/statusline/launch)은 무변경. 본 SPEC은 project-scope 두 섹션만.
- **관찰 사항(고치지 말 것)**: internal/config/audit_registry.go에 5개 stale YAML-only 예외(constitution/context/design/harness/interview — 지금은 구조체 존재)가 있으나, 이는 별도 cleanup SPEC 대상이며 007 스코프 아님.

---

## §G. Cross-References

- SPEC-WEB-CONSOLE-006 (enabler — Templ 섹션 트리 + HTMX foundation; completed v0.2.0 `5714bae97`)
- SPEC-WEB-CONSOLE-003 (현 scalar project-config 편집의 원천 — writeProjectConfig/validateProjectConfig)
- research.md (전체 file:line ground-truth 인용)
- design.md (nested serialization + per-field validation map + 신규 위젯 + write-seam load-modify-write + nested isolation 메커니즘)
- acceptance.md (AC-WC7-NNN, 각 GREEN-verifiable)
- plan.md (M1..Mn + Tier 정당화)
