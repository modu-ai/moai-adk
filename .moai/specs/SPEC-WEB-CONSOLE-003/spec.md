---
id: SPEC-WEB-CONSOLE-003
title: "Web Console — Flat Project-Config Parity (development_mode + git_convention.convention)"
version: "0.2.0"
status: implemented
created: 2026-06-03
updated: 2026-06-03
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "internal/web"
lifecycle: spec-anchored
tags: "web, console, config, parity, development_mode, git-convention, project-config, tui"
tier: M
related_specs: [SPEC-WEB-CONSOLE-001, SPEC-WEB-CONSOLE-002]
---

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-06-03 | manager-spec | Initial draft — S2a of web-console-v3 cohort. FLAT/SHALLOW project-config parity: surface `development_mode` + `git_convention.convention` (top-level enum fields only) in BOTH web console and TUI profile wizard. Introduces a project-config write path in the web layer (modest `SyncToProjectConfig`-pattern extension). Deeply-nested config sections deferred to S2b (Tier L). |

---

## §1 Context & Motivation

`SPEC-WEB-CONSOLE-001` shipped a loopback-only, no-auth browser settings editor (`moai web`) that edits **profile preferences** (`preferences.yaml` via `WritePreferences`) and syncs a fixed subset (user/language/statusline) to the project config via `SyncToProjectConfig`. `SPEC-WEB-CONSOLE-002` (S1) hardened web↔TUI validation parity for three profile fields (`model`/`effort_level`/`model_policy`), converted those widgets to `<select>`, added the TUI `model_policy` select, and set the default port to `3041`.

A larger gap remains: the web console (and the TUI wizard) expose **zero project-config settings beyond user/language/statusline**. Research into surfacing the full nested config (quality, workflow, git-strategy, harness, the nested sub-fields of git-convention/llm) revealed that path is **Tier L** — those sections carry deeply-nested structures (dicts, lists, conditional sub-fields) that demand a substantial new persistence-model redesign. The user approved an **incremental split** ("점진 분할"): expose ONLY the *flat / shallow / single-enum / top-level* settings now (S2a), defer the deeply-nested editing to a future **S2b (Tier L)** SPEC.

### Confirmed flat settings (ground-truth verified)

Two settings qualify as flat/shallow (single top-level enum field, no nested sub-structure required for the value):

| Setting | Source yaml | Config type/field (verified) | Canonical enum (verified) | Predicate status |
|---------|-------------|------------------------------|---------------------------|------------------|
| development_mode | `.moai/config/sections/quality.yaml` → `constitution.development_mode` | `pkg/models` `QualityConfig.DevelopmentMode` (`yaml:"development_mode" validate:"omitempty,oneof=ddd tdd"`) | `{ddd, tdd}` | EXPORTED — `models.ValidDevelopmentModes()` returns `[ModeDDD, ModeTDD]` + `DevelopmentMode.IsValid()` method |
| git_convention.convention | `.moai/config/sections/git-convention.yaml` → `git_convention.convention` | `pkg/models` `GitConventionConfig.Convention` (`yaml:"convention" validate:"omitempty,oneof=auto conventional-commits angular karma custom"`) | `{auto, conventional-commits, angular, karma, custom}` | NO exported predicate; canonical map `validGitConventionNames` is **unexported** in `internal/config/validation.go:130` |

> **NOTE — nested fields of git-convention are DEFERRED.** S2a touches ONLY the top-level `git_convention.convention` enum field. The nested `auto_detection`/`validation`/`formatting`/`custom` sub-structures (verified present at `.moai/config/sections/git-convention.yaml`) are **S2b** scope. In particular, `convention: custom` requires `custom.pattern` (validated at `internal/config/validation.go:178-179`) — S2a's enum value `custom` is permitted as a value, but S2a does NOT surface the `custom.*` sub-fields needed to fully configure it. See §4 Exclusion E.1.

### Scope reduction recorded: llm.mode / llm.default_model NARROWED OUT

The cohort scoping prompt named `llm.mode` and `llm.default_model` as **candidate** flat settings, to be verified. Ground-truth verification **rejected both** for S2a:

- **`llm.mode`** (`internal/config/types.go:228`) accepts only `"" | "glm"` — a 2-value toggle. Its only non-empty value (`glm`) flips the entire LLM backend to the z.ai gateway. Exposing a backend-switch toggle in a settings dropdown is a behavior-altering surface that exceeds the "harmless flat enum parity" intent of S2a; it warrants its own deliberate UX decision (deferred).
- **`llm.default_model`** (`internal/config/types.go:241-242`) is explicitly annotated *"Legacy fields (kept for backward compatibility, mapped from tiers)"*, carries **no `validate` tag** and **no canonical enum** (unconstrained string). The modern model-selection path is `llm.claude_models.{high,medium,low}` tiers, not this legacy field. There is no canonical predicate to reuse — surfacing it as a `<select>` would require **authoring a fresh enum**, which is the exact anti-pattern S2a forbids (no parallel/invented rule-set).

Therefore S2a is **narrowed to `development_mode` + `git_convention.convention`** (the 2 settings with clean, existing canonical enums). This reduction is recorded here per the cohort scoping instruction; `llm.*` exposure (if ever desired) is a separate future SPEC decision, not S2b's nested-editing scope and not S2a's.

### KEY design decision — project-config persistence (differs from S1)

S1's settings (`model`/`effort_level`/`model_policy`) live in the **profile** store (`preferences.yaml`). S1 (REQ-WC2-007) deliberately kept `model_policy` profile-only and did NOT extend `SyncToProjectConfig`.

S2a's settings are **project-config** settings: they live in `.moai/config/sections/quality.yaml` + `.moai/config/sections/git-convention.yaml`, NOT in the profile `preferences.yaml`. Verified: `internal/profile/preferences.go` `ProfilePreferences` has **no field** for `development_mode` or `convention`; `internal/profile/sync.go` `SyncToProjectConfig` scopes to user/language/statusline only.

Therefore S2a MUST introduce a **project-config write path** that persists these two enum values into their respective `.moai/config/sections/*.yaml` files. The canonical mechanism already exists: `config.NewConfigManager()` → `LoadRaw(projectRoot)` → mutate `cfg.Quality.DevelopmentMode` / `cfg.GitConvention.Convention` → `SetSection("quality", ...)` / `SetSection("git_convention", ...)` → `Save()` (verified: `internal/config/manager.go:138 SetSection`, `:155 Save` already persist `quality.yaml` + `git-convention.yaml`). S2a wires a thin function over this existing API — a modest extension of the `SyncToProjectConfig` *pattern* (load → mutate non-empty → SetSection → Save), NOT the large nested-config persistence redesign that S2b will need.

This SPEC closes the parity gap with the **smallest** change that reuses existing canonical enums/predicates and the existing config-manager persistence API — authoring a parallel rule-set or a bespoke YAML marshal is the explicit anti-pattern (see §4).

### Cohort scope-fence

This SPEC is **S2a** of the `web-console-v3` cohort. Only S2a is authored here. Siblings are referenced ONLY to delimit what S2a does NOT do (see §4):
- **S2b** (future, Tier L) — full editing of quality/workflow/git-strategy/harness AND the nested sub-fields of git-convention/llm (auto_detection, validation, claude_models dict, glm dict, custom.* etc.).
- **S3** — 4-language web i18n + Pretendard/Noto-CJK webfont.
- **S4** — dead-config audit + removal.

---

## §2 GEARS Requirements

### REQ-WC3-001 (Event-driven — web `development_mode` validation)

**When** the Console receives a `POST /save` whose `development_mode` field is non-empty and is not one of the 2 canonical values (`ddd` / `tdd` — empty meaning "leave project default unchanged" is always allowed), the Console **shall** reject the mutation, leave persisted state unchanged, and re-render the form with a per-field `development_mode` error — reusing the canonical set from `pkg/models` (`models.ValidDevelopmentModes()` / `models.DevelopmentMode.IsValid()`), NOT a freshly authored list.

### REQ-WC3-002 (Event-driven — web `git_convention.convention` validation)

**When** the Console receives a `POST /save` whose `git_convention` field is non-empty and is not one of the 5 canonical values (`auto` / `conventional-commits` / `angular` / `karma` / `custom` — empty meaning "leave project default unchanged" is always allowed), the Console **shall** reject the mutation, leave persisted state unchanged, and re-render with a per-field `git_convention` error — reusing the canonical convention set as the single source of truth. Because the canonical `validGitConventionNames` map is unexported in `internal/config`, S2a **shall** either (a) add ONE exported predicate `IsValidConvention(string) bool` (+ `ValidConventions() []string`) to the canonical `internal/config` or `pkg/models` package and reuse it, OR (b) mirror the 5-value list in the web layer with an MX:NOTE pointing at the canonical `oneof` validate tag (`pkg/models` `GitConventionConfig.Convention`) as SSOT. Adding one exported predicate (option a) is preferred.

### REQ-WC3-003 (Ubiquitous — web widget select-ification)

The Console page **shall** render the `development_mode` and `git_convention` fields as `<select>` dropdowns whose option sets equal the canonical lists of REQ-WC3-001/002 (with an empty-value "(project default)" option), bound through the existing `newPageView()` view-model mechanism and rendered via the existing `optSelect` define block (`internal/web/assets/page.html.tmpl:127-138`). A new `<fieldset>` (legend "Project") **shall** host these two project-config selects, visually distinct from the profile-scoped fieldsets.

### REQ-WC3-004 (Ubiquitous — web project-config read on render)

The Console **shall** populate the `development_mode` and `git_convention` selects on `GET /` from the **current project config** (`.moai/config/sections/quality.yaml` / `git-convention.yaml`), reading via the config manager (`LoadRaw`), so the rendered form reflects the project's actual persisted values (not the profile store, which does not hold these fields). A read failure **shall** surface as a readable inline error (REQ-WC-010 invariant), never a blank page or panic.

### REQ-WC3-005 (Event-driven — web project-config persistence)

**When** a `POST /save` carries valid (or empty) `development_mode` and `git_convention` values AND validation passes, the Console **shall** persist each non-empty value into its project-config section via the config-manager API (`LoadRaw` → mutate `cfg.Quality.DevelopmentMode` / `cfg.GitConvention.Convention` → `SetSection` → `Save`), overwriting only non-empty submitted values (empty = "keep existing"), and **shall not** perform a direct `yaml.Marshal`/`os.WriteFile` from the web layer. This persistence path **shall** be added as a new injectable seam on the `app` struct (mirroring the existing `syncToProject` seam) so tests can inject failures without touching the real filesystem.

### REQ-WC3-006 (Ubiquitous — TUI `development_mode` + `git_convention.convention` parity)

The terminal profile wizard **shall** present a `development_mode` `<select>` and a `git_convention` `<select>` (in a new or existing wizard group), offering the canonical values of REQ-WC3-001/002 plus an empty "(project default)" option, with localized labels added to `internal/cli/profile_setup_translations.go` for all four locales (en / ko / ja / zh). On wizard save, the selected non-empty values **shall** persist to the project config via the SAME project-config write path as REQ-WC3-005 (NOT into `ProfilePreferences`, which has no slot for them), so the TUI and the web console reach functional parity.

### REQ-WC3-007 (Ubiquitous — SyncToProjectConfig-pattern extension, bounded)

The project-config write path introduced by REQ-WC3-005 **shall** be a thin extension of the existing `SyncToProjectConfig` *pattern* (load via config manager → mutate only non-empty fields → `SetSection` → `Save`). It **shall** write ONLY the `quality` (development_mode field) and `git_convention` (convention field) sections; it **shall not** touch `workflow`, `harness`, `git-strategy`, `llm`, or any nested sub-field of those sections. No new config section is introduced. The change reuses the config manager's existing atomic `Save()` path (`internal/config/manager.go:155`).

### REQ-WC3-008 (Ubiquitous — invariant preservation)

The Console **shall** preserve all `SPEC-WEB-CONSOLE-001` / `SPEC-WEB-CONSOLE-002` invariants unchanged: loopback-only bind (`127.0.0.1`, `internal/web/server.go`), no-auth / no-token / no-session (`internal/web/app.go` hostCheckMiddleware), Host-header write-safety check, persistence exclusively through profile/sync functions + the config-manager API (never a direct YAML marshal from the web layer, `internal/web/handlers.go` MX:WARN), and the `internal/web/integration_test.go` DO_NOT_TOUCH sentinels for workflow.yaml / harness.yaml / git-strategy.yaml (which S2a touches NONE of).

---

## §3 Acceptance Criteria (summary — full enumeration in acceptance.md)

Each AC is independently verifiable. `go test ./internal/web/... ./internal/cli/... ./internal/config/...` MUST be green as the closure gate. The full Given-When-Then enumeration with edge cases lives in `acceptance.md`; the table below is the SSOT index.

| AC | REQ | Assertion (one-line) |
|----|-----|----------------------|
| AC-WC3-001a | REQ-WC3-001 | Web rejects out-of-list `development_mode` (e.g. `xyz`) → 400 + field error + state unchanged |
| AC-WC3-001b | REQ-WC3-001 | Web persists valid `development_mode` (`ddd`) → 200 + quality.yaml updated |
| AC-WC3-002a | REQ-WC3-002 | Web rejects out-of-list `git_convention` (e.g. `gitflow`) → 400 + field error + state unchanged |
| AC-WC3-002b | REQ-WC3-002 | Web persists valid `git_convention` (`angular`) → 200 + git-convention.yaml updated |
| AC-WC3-003 | REQ-WC3-003 | Both fields render as `<select>` (no `type="text"`) with exactly the canonical options + empty-default, inside a "Project" fieldset |
| AC-WC3-004 | REQ-WC3-004 | `GET /` pre-selects the current project-config values (read from quality.yaml/git-convention.yaml, not the profile store); read failure → readable inline error |
| AC-WC3-005 | REQ-WC3-005 | Project-config write path is a new injectable `app` seam; empty submitted value leaves the existing section value unchanged; no direct `yaml.Marshal` in web layer |
| AC-WC3-006a | REQ-WC3-006 | TUI exposes a `development_mode` select + a `git_convention` select; 4-locale labels non-empty (parity test) |
| AC-WC3-006b | REQ-WC3-006 | TUI save persists the two values to project config (quality.yaml + git-convention.yaml), NOT to preferences.yaml |
| AC-WC3-007 | REQ-WC3-007 | The write path touches ONLY quality + git_convention sections; workflow/harness/git-strategy/llm untouched |
| AC-WC3-008 | REQ-WC3-008 | 001/002 invariant tests still green; integration_test.go DO_NOT_TOUCH sentinels intact; no `0.0.0.0`/auth/token added |
| AC-WC3-009 | all | Closure gate: `go test ./internal/web/... ./internal/cli/... ./internal/config/...` exit 0 |

---

## §4 Exclusions (What NOT to Build)

### Out of Scope

The following are deferred to sibling SPECs and MUST NOT be implemented in S2a:

- **S2b scope** — full editing of quality / workflow / git-strategy / harness sections AND the nested sub-fields of git-convention/llm.
- **S3 scope** — 4-language web i18n + Pretendard/Noto-CJK webfont.
- **S4 scope** — dead-config audit + removal.
- **Anti-patterns** — duplicate/invented validation enums, a new config section, direct YAML marshal in the web layer, exposing legacy/backend-switching llm fields (detailed in the numbered list below).

[HARD] The following are explicitly **out of scope** for S2a and MUST NOT be implemented:

### E.1 Out of Scope — nested git-convention / llm sub-fields (S2b)

The nested sub-structures of `git_convention` (`auto_detection.{enabled,sample_size,confidence_threshold,fallback}`, `validation.{enabled,enforce_on_commit,enforce_on_push,max_length}`, `formatting.{show_examples,show_suggestions,verbose}`, `custom.{name,pattern,types,scopes,max_length,examples}`) and of `llm` (`claude_models.{high,medium,low}`, `glm.{base_url,models,context_windows}`, `mode`, `team_mode`, `default_model`, `quality_model`, `speed_model`, `performance_tier`) are **S2b (Tier L)** scope. S2a surfaces ONLY the top-level `development_mode` and `git_convention.convention` enum fields. The enum value `convention: custom` MAY be selected but S2a does NOT surface the `custom.pattern` sub-field required to fully configure a custom convention — selecting `custom` without a pre-existing `custom.pattern` is a user responsibility deferred to S2b (S2a does not add custom.* widgets).

### E.2 Out of Scope — full project-config sections (S2b)

The full `quality` section (test_coverage_target, coverage_threshold, ddd_settings, tdd_settings, lsp_quality_gates, etc.), the entire `workflow`, `harness`, and `git-strategy` sections are **S2b**. S2a writes ONLY the single `development_mode` field of the quality section and the single `convention` field of the git-convention section — all other fields of those sections are read-then-rewritten unchanged by the load-mutate-save cycle (the config manager round-trips them; S2a never sets them).

### E.3 Out of Scope — web i18n + webfont (S3)

4-language web UI translation and Pretendard / Noto-CJK CDN webfont loading are **S3**, not S2a. The web page stays English. (TUI `development_mode`/`git_convention` labels in REQ-WC3-006 are TUI-only and follow the existing `profile_setup_translations.go` 4-locale pattern — this is parity with already-localized TUI fields, NOT new web i18n.)

### E.4 Out of Scope — dead-config audit/removal (S4)

Auditing or removing unused config keys is **S4**.

### E.5 Out of Scope — additional anti-patterns

[HARD] The following are forbidden regardless of sibling-SPEC scope:

1. **No duplicate/invented validation enums** — reuse `models.ValidDevelopmentModes()` for `development_mode`. For `git_convention.convention`, add ONE exported `IsValidConvention`/`ValidConventions` predicate to the canonical package (preferred) OR mirror the 5-value list with an MX:NOTE pointing at the `pkg/models` `oneof` SSOT. Do NOT author a third parallel rule-set, and do NOT author a fresh enum for any field that lacks one (this is why `llm.default_model` is excluded — no canonical enum exists to reuse).
2. **No new config section** — REQ-WC3-007 writes ONLY the existing `quality` + `git_convention` sections via `SetSection`. Do NOT add a new `.moai/config/sections/*.yaml` file or a new top-level config key.
3. **No direct YAML marshal in the web layer** — persistence goes through `config.NewConfigManager()`/`SetSection`/`Save` (and the existing `WritePreferences`/`SyncToProjectConfig` for unchanged profile fields). A direct `yaml.Marshal`/`os.WriteFile` of a config section from `internal/web` is the forbidden anti-pattern (mirrors `SPEC-WEB-CONSOLE-001` REQ-WC-007 + the handlers.go MX:WARN).
4. **No llm.mode / llm.default_model exposure** — narrowed out in §1; `llm.mode` is a backend-switch toggle, `llm.default_model` is a legacy enum-less string. Neither is surfaced in S2a.
5. **No auth / token / session / non-loopback bind** — the no-auth loopback-only posture of `SPEC-WEB-CONSOLE-001` is invariant (REQ-WC3-008). S2a adds zero security surface.
6. **No template mirroring / make build** — this is a Go binary feature under `internal/`, NOT a deployed asset under `internal/template/templates/`. The `page.html.tmpl` is embedded via the web package's own `go:embed` (verified pattern from S1 §4.8). No `make build`/embedded-mirror parity step applies.
7. **No touching the integration_test.go DO_NOT_TOUCH sentinels** — `internal/web/integration_test.go` asserts workflow.yaml/harness.yaml/git-strategy.yaml are NOT written. S2a writes NONE of those sections, so those sentinels MUST remain green and unmodified.

---

## §5 References (verified ground-truth)

| Path | Role |
|------|------|
| `internal/web/validate.go:14-64,75-121` | mirror lists + `validatePrefs(ProfilePreferences)` — add a sibling project-config validator (NOT inside validatePrefs — different value source) |
| `internal/web/handlers.go:12-56` | `pageView` + `newPageView()` — add `DevelopmentModes`/`Conventions` option lists + current project-config values to the view-model |
| `internal/web/handlers.go:108-155,171-198` | `handleSave` + `bindForm` — bind 2 new form fields (NOT into ProfilePreferences) + invoke the project-config write path |
| `internal/web/handlers.go:103-107` | MX:WARN persistence boundary — extend rationale to cover the new config-manager write seam |
| `internal/web/app.go:16-40` | `app` struct injectable seams (`readPreferences`/`writePreferences`/`syncToProject`) — add `readProjectConfig`/`writeProjectConfig` seams |
| `internal/web/assets/page.html.tmpl:50-64,127-138` | `optSelect` define block + Launch fieldset — add a "Project" fieldset with 2 `optSelect` instances |
| `internal/config/manager.go:138,155,295-318` | `SetSection("quality"/"git_convention")` + `Save()` — the persistence API to wrap |
| `internal/config/manager.go:86 LoadRaw` | load current config to read existing values + round-trip unchanged sections |
| `pkg/models/config.go:4-29,49-50` | `DevelopmentMode` type + `ModeDDD`/`ModeTDD` + `ValidDevelopmentModes()` + `QualityConfig.DevelopmentMode` |
| `pkg/models/config.go:206-208` | `GitConventionConfig.Convention` + `oneof=auto conventional-commits angular karma custom` validate tag (SSOT for the 5-value enum) |
| `internal/config/validation.go:130-136` | unexported `validGitConventionNames` map — canonical 5 values; add an exported predicate OR mirror |
| `internal/cli/profile_setup.go:241-340,447-474` | wizard groups + persistence (WritePreferences + SyncToProjectConfig) — add 2 selects + project-config write |
| `internal/cli/profile_setup_translations.go:29-44,150-461` | 4-locale label struct (ModelPolicyTitle pattern) — add DevelopmentMode + GitConvention label fields |
| `internal/cli/profile_setup_translations_test.go:5-51` | `TestGetProfileText_AllLanguages` parity guard — extend with the new label keys |
| `internal/web/integration_test.go` | DO_NOT_TOUCH sentinels for workflow/harness/git-strategy — MUST stay intact |
| `internal/cli/web.go:47` | `--port 3041` default (already set by S1; unchanged by S2a) |
| `.moai/config/sections/quality.yaml` | `constitution.development_mode: tdd` (current value) |
| `.moai/config/sections/git-convention.yaml` | `git_convention.convention: auto` (current value) + nested sub-fields (S2b) |
