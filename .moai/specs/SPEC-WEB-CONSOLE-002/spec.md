---
id: SPEC-WEB-CONSOLE-002
title: "Web Console — Port 3041 Default + Web↔TUI Validation Parity"
version: "0.1.0"
status: completed
created: 2026-06-03
updated: 2026-06-03
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "internal/web"
lifecycle: spec-anchored
tags: "web, console, validation, parity, port, tui, model-policy"
tier: S
related_specs: [SPEC-WEB-CONSOLE-001]
---

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-06-03 | manager-spec | Initial draft — S1 of web-console-v3-extension cohort. Port 3041 default + 3-field web validation parity (model/effort_level/model_policy) + widget select-ification + TUI model_policy parity. |

---

## §1 Context & Motivation

`SPEC-WEB-CONSOLE-001` shipped a loopback-only, no-auth browser settings editor (`moai web`) that reuses the terminal profile wizard's validation and persistence (`WritePreferences` + `SyncToProjectConfig`). It is implemented and in-progress.

Two correctness gaps and one default-value decision remain after that SPEC:

1. **Validation asymmetry (WHY this matters)** — `internal/web/validate.go` `validatePrefs()` validates `permission_mode`, the four language fields, and the three statusline fields against canonical lists/predicates, but **silently accepts `model`, `effort_level`, and `model_policy` as free text** (verified: `internal/web/validate.go:62-94` has no branch for these three; `bindForm` at `internal/web/handlers.go:164-178` binds them into `ProfilePreferences` but nothing rejects an out-of-list value). The terminal wizard, by contrast, constrains `model` and `effort_level` to fixed `huh.Select` option sets (`internal/cli/profile_setup.go:299-323`). A user typing `gpt-4` or `ultra` into the web `model`/`effort_level` text input persists a meaningless value with no feedback.

2. **Widget asymmetry** — the same three web fields are rendered as `<input type="text">` (`internal/web/assets/page.html.tmpl:62-71`) while every other constrained web field is already a `<select>`. Free-text widgets invite the invalid values gap #1 would now reject, producing a worse UX than a dropdown.

3. **TUI omission** — `model_policy` is web-only. The terminal wizard exposes `model`, `effort_level`, `permission_mode` in its model-settings group but has **no `model_policy` select** (verified: `internal/cli/profile_setup.go:298-337` model-settings group contains exactly those three selects, no `model_policy`). The two editors are not at parity.

4. **Port default** — `SPEC-WEB-CONSOLE-001` REQ-WC-001 set `--port` default as "candidate 8080" (`internal/cli/web.go:47`, plus the Long help text at `:37-43`). 8080 frequently collides with other local dev servers. S1 supersedes that candidate with **3041** (free, user/dynamic port range).

This SPEC closes those gaps with the **smallest** change that reuses existing canonical lists/predicates — duplicating validation logic is the explicit anti-pattern (see §4 Exclusions).

### Cohort scope-fence

This SPEC is **S1** of a 4-SPEC `web-console-v3-extension` cohort. Only S1 is authored here. Siblings S2/S3/S4 are referenced ONLY to delimit what S1 does NOT do (see §4). The cohort's later SPECs cover: S2 — surface the 8 UI-missing v3 settings (development_mode, quality, workflow, git-convention, harness, llm, git-strategy) in both editors; S3 — 4-language i18n + Pretendard/Noto-CJK webfont; S4 — dead-config audit + removal.

---

## §2 GEARS Requirements

### REQ-WC2-001 (Ubiquitous — port default supersede)

The `moai web` subcommand **shall** default `--port` to `3041` (overriding the `8080` candidate set by `SPEC-WEB-CONSOLE-001` REQ-WC-001), and `moai web --help` **shall** present `3041` as the documented default.

> **Supersede note**: This requirement supersedes the `--port` *default value* of `SPEC-WEB-CONSOLE-001` REQ-WC-001 ONLY. The loopback-bind, `--no-open`, and Host-header write-safety invariants of REQ-WC-001..REQ-WC-002 remain in force unchanged.

### REQ-WC2-002 (Event-driven — web model validation parity)

**When** the Console receives a `POST /save` whose `model` field is non-empty and is not one of the 7 canonical model values (`opus` / `opus[1m]` / `sonnet` / `sonnet[1m]` / `haiku` / `opusplan` — the empty string meaning "project default" is always allowed), the Console **shall** reject the mutation, leave persisted state unchanged, and re-render the form with a per-field `model` error — reusing the wizard's canonical model set as the single source of truth (`internal/cli/profile_setup.go:303-310`; reuse `normalizeModel` at `:122-142` for deprecated-ID acceptance where importable, else mirror the 7-value list).

### REQ-WC2-003 (Event-driven — web effort_level validation parity)

**When** the Console receives a `POST /save` whose `effort_level` field is non-empty and is not one of the 6 canonical effort values (`low` / `medium` / `high` / `xhigh` / `max` — empty meaning "runtime default" always allowed), the Console **shall** reject the mutation, leave persisted state unchanged, and re-render with a per-field `effort_level` error — reusing the wizard's canonical effort set (`internal/cli/profile_setup.go:316-322`).

### REQ-WC2-004 (Event-driven — web model_policy validation parity)

**When** the Console receives a `POST /save` whose `model_policy` field is non-empty and is not one of the 3 canonical policy values (`high` / `medium` / `low` — empty always allowed), the Console **shall** reject the mutation, leave persisted state unchanged, and re-render with a per-field `model_policy` error, wiring in the **existing-but-unused** predicate `template.IsValidModelPolicy()` (`internal/template/model_policy.go:30-40`) rather than authoring a new list.

### REQ-WC2-005 (Ubiquitous — web widget select-ification)

The Console page **shall** render the `model`, `effort_level`, and `model_policy` fields as `<select>` dropdowns whose option sets equal the canonical lists of REQ-WC2-002..004 (with an empty-value "(project default)" / "(runtime default)" / "(unset)" option), replacing the three `<input type="text">` widgets currently at `internal/web/assets/page.html.tmpl:62-71`. The option lists **shall** be bound through the existing `newPageView()` view-model mechanism (`internal/web/handlers.go:34-49`).

### REQ-WC2-006 (Ubiquitous — TUI model_policy parity)

The terminal profile wizard **shall** present a `model_policy` `<select>` in its model-settings group (alongside the existing model / effort_level / permission_mode selects at `internal/cli/profile_setup.go:298-337`), offering the 3 canonical policy values plus an empty "(project default)" option, with localized labels added to `internal/cli/profile_setup_translations.go` for all four locales (en / ko / ja / zh).

### REQ-WC2-007 (Ubiquitous — model_policy sync decision: profile-only)

The `model_policy` preference **shall** persist to the profile store only (`preferences.yaml` via `WritePreferences`) and **shall not** be synced to any project `.moai/config` section. No new config section is introduced solely to host `model_policy` (over-engineering avoidance — see §4).

> **Rationale**: `model_policy` is consumed at agent-model-routing time by `template.ApplyModelPolicy` (`internal/template/model_policy.go:243`), reading from the profile, not from a project config section. `SyncToProjectConfig` already scopes to user/language/statusline only (per the `handleSave` MX note, `internal/web/handlers.go:96-100`). Inventing a config section purely to "sync" `model_policy` would add an unconsumed write path.

### REQ-WC2-008 (Ubiquitous — invariant preservation)

The Console **shall** preserve all `SPEC-WEB-CONSOLE-001` invariants unchanged: loopback-only bind (`127.0.0.1`), no-auth / no-token / no-session, Host-header write-safety check, and persistence exclusively through `WritePreferences` + `SyncToProjectConfig` (never a direct YAML marshal from the web layer).

---

## §3 Acceptance Criteria (inline — Tier S)

Each AC is independently verifiable. `go test ./internal/web/... ./internal/cli/...` MUST be green as the closure gate.

### AC-WC2-001 — Port default is 3041 (REQ-WC2-001)

- **Given** a freshly built `moai` binary
- **When** the `web` cobra command's `--port` flag default is inspected (or `moai web --help` is run)
- **Then** the default value is `3041`, and the help text references `3041` (not `8080`)
- **Verify**: `internal/cli/web.go` `IntVar(&webPort, "port", 3041, ...)` + Long help string updated; a table-driven test asserts the flag's `DefValue == "3041"`.

### AC-WC2-002a — Web rejects out-of-list `model` (REQ-WC2-002)

- **Given** the Console running
- **When** a `POST /save` submits `model=gpt-4` (a non-canonical value)
- **Then** the response is HTTP 400, `FieldErrors["model"]` is non-empty, and the persisted `preferences.yaml` `model` value is unchanged
- **Verify**: `internal/web` handler/validate test asserts `validatePrefs` returns a `model` key for a bogus value and the empty map for each of the 7 canonical values + empty string.

### AC-WC2-002b — Web persists valid `model` (REQ-WC2-002)

- **Given** the Console running
- **When** a `POST /save` submits `model=sonnet[1m]` (canonical)
- **Then** the response is HTTP 200 ("Settings saved") and the persisted `model` is `sonnet[1m]`
- **Verify**: round-trip handler test (write then read back) using a temp profile dir.

### AC-WC2-003 — Web rejects out-of-list `effort_level`, persists valid (REQ-WC2-003)

- **Given** the Console running
- **When** `POST /save` submits `effort_level=ultra` (bogus) — then separately `effort_level=xhigh` (canonical)
- **Then** the bogus submit returns 400 with a `effort_level` field error and no state change; the canonical submit returns 200 and persists `xhigh`
- **Verify**: `validatePrefs` test covers all 6 canonical values + empty (no error) and ≥1 bogus value (error).

### AC-WC2-004 — Web rejects out-of-list `model_policy` via `IsValidModelPolicy`, persists valid (REQ-WC2-004)

- **Given** the Console running
- **When** `POST /save` submits `model_policy=ultra` (bogus) — then `model_policy=medium` (canonical)
- **Then** the bogus submit returns 400 with a `model_policy` field error and no state change; the canonical submit returns 200 and persists `medium`
- **Verify**: `validatePrefs` delegates to `template.IsValidModelPolicy`; test asserts error for bogus, no error for each of `high`/`medium`/`low`/empty.

### AC-WC2-005 — Web widgets are `<select>` with canonical options (REQ-WC2-005)

- **Given** the rendered Console page (`GET /`)
- **When** the HTML for the `model`, `effort_level`, `model_policy` fields is inspected
- **Then** each is a `<select>` element (no `<input type="text">` for these three), and each contains exactly its canonical option set plus the empty-default option
- **Verify**: a render test (template executed against a view-model) asserts the three field names appear inside `<select` blocks with the expected `<option value="...">` entries; a negative assertion confirms no `type="text"` for `name="model"|"effort_level"|"model_policy"`.

### AC-WC2-006 — TUI exposes `model_policy` select with 4-locale labels (REQ-WC2-006)

- **Given** the terminal profile wizard form
- **When** the model-settings group is constructed
- **Then** it contains a `model_policy` `huh.Select` with options `(project default)`/`high`/`medium`/`low`, and `profile_setup_translations.go` defines the corresponding label strings for en/ko/ja/zh
- **Verify**: a translations test asserts the new label keys are non-empty for all four locales (mirrors any existing `TestProfileTranslations*` parity test); a wizard-construction test (or grep-guard) confirms a `model_policy`-bound select exists.

### AC-WC2-007 — model_policy stays profile-only (REQ-WC2-007)

- **Given** a `POST /save` with `model_policy=high`
- **When** persistence completes
- **Then** `preferences.yaml` contains `model_policy: high` AND no project `.moai/config` section was written for model_policy (no new config key/section introduced)
- **Verify**: handler test asserts the profile file holds the value and that `SyncToProjectConfig`'s touched-section set is unchanged from `SPEC-WEB-CONSOLE-001` (user/language/statusline only).

### AC-WC2-008 — invariants preserved (REQ-WC2-008)

- **Given** the full S1 change set
- **When** the existing `SPEC-WEB-CONSOLE-001` test suite runs
- **Then** loopback-bind, no-auth, Host-header check, and persistence-path tests all still pass (no regression)
- **Verify**: `go test ./internal/web/...` green; no test asserting `0.0.0.0` bind, token auth, or direct-YAML-write is added or weakened.

### AC-WC2-009 — closure gate (all REQs)

- **Given** the complete S1 implementation
- **When** `go test ./internal/web/... ./internal/cli/...` runs
- **Then** the suite is green with zero failures
- **Verify**: command exit 0.

---

## §4 Exclusions (What NOT to Build)

### Out of Scope

The following are deferred to sibling SPECs and MUST NOT be implemented in S1:

- **S2 scope** — 8 UI-missing v3 settings (development_mode, quality, workflow, git-convention, harness, llm, git-strategy)
- **S3 scope** — 4-language web i18n + Pretendard/Noto-CJK webfont (CDN-loaded)
- **S4 scope** — dead-config audit + removal
- **Anti-patterns** — duplicate validation lists, a new config section for model_policy, template mirroring (detailed in the numbered list below)

[HARD] The following are explicitly **out of scope** for S1 and MUST NOT be implemented:

1. **No duplicate validation lists** — do NOT author a new canonical list for `model`, `effort_level`, or `model_policy` inside `internal/web`. Reuse the wizard's option sets and `template.IsValidModelPolicy()`. A second, parallel rule-set is the primary anti-pattern this SPEC exists to avoid (consistent with `SPEC-WEB-CONSOLE-001` REQ-WC-008 "no parallel validation rule set"). The `internal/web` validate.go MX note (`:15-19`) already documents that canonical lists are mirrored from the wizard SSOT under the unidirectional `internal/cli → internal/web` import constraint; where `normalizeModel`/`IsValidModelPolicy` are importable, import them; mirroring is permitted ONLY where the wizard value is unexported.
2. **No new config section for model_policy** — REQ-WC2-007 keeps it profile-only. Do NOT add a `model_policy:` section to any `.moai/config/sections/*.yaml` or extend `SyncToProjectConfig`'s scope.
3. **No i18n of the web console** — 4-language web UI translation is **S3**, not S1. The web page stays English. (TUI `model_policy` labels in REQ-WC2-006 are TUI-only and follow the existing `profile_setup_translations.go` pattern — this is parity with already-localized TUI fields, not new web i18n.)
4. **No webfont work** — Pretendard / Noto-CJK CDN webfont loading is **S3**.
5. **No new settings surfaced** — the 8 UI-missing v3 settings (development_mode, quality, workflow, git-convention, harness, llm, git-strategy) are **S2**. S1 touches ONLY the three already-present-but-unvalidated fields plus the TUI `model_policy` addition.
6. **No dead-config removal** — auditing/removing unused config keys is **S4**.
7. **No auth / token / session / non-loopback bind** — the no-auth loopback-only posture of `SPEC-WEB-CONSOLE-001` is invariant (REQ-WC2-008). S1 adds zero security surface.
8. **No template mirroring** — this is a Go binary feature under `internal/`, NOT a deployed asset under `internal/template/templates/`. No `make build` / embedded-mirror parity step is required for the validation/widget/port changes (the `page.html.tmpl` lives under `internal/web/assets/`, embedded via the web package's own `go:embed`, not the template-deploy system).

---

## §5 References (verified ground-truth)

| Path | Role |
|------|------|
| `internal/web/validate.go:23-41,62-94` | existing canonical lists + `validatePrefs()` — extend with 3 field branches |
| `internal/web/assets/page.html.tmpl:62-71` | 3 text inputs → convert to `<select>` |
| `internal/web/assets/page.html.tmpl:123-134` | `langSelect` define block — pattern template for new selects |
| `internal/web/handlers.go:34-49` | `newPageView()` option binding — add Model/Effort/ModelPolicy option lists |
| `internal/web/handlers.go:164-191` | `bindForm()` — already binds the 3 fields; no change needed |
| `internal/cli/profile_setup.go:299-323` | TUI model + effort selects — parity template for model_policy |
| `internal/cli/profile_setup.go:122-142` | `normalizeModel()` — deprecated-ID migration, reuse if importable |
| `internal/cli/profile_setup_translations.go` | 4-locale label strings — add model_policy labels |
| `internal/template/model_policy.go:18-40` | `ModelPolicy` constants + `ValidModelPolicies()` + unused `IsValidModelPolicy()` |
| `internal/cli/web.go:47,37-43` | `--port` default 8080 → 3041 + Long help text |
| `internal/profile/preferences.go:25` | `ModelPolicy` field (already exists) |
| `.moai/specs/SPEC-WEB-CONSOLE-001/spec.md:67-69` | REQ-WC-001 "default candidate 8080" — superseded by REQ-WC2-001 |
