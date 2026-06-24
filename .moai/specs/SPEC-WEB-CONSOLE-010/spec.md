---
id: SPEC-WEB-CONSOLE-010
title: "Web Console — Shared Field-Schema SSOT for moai web + moai profile setup"
version: "0.1.0"
status: completed
created: 2026-06-22
updated: 2026-06-23
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "internal/settings"
lifecycle: spec-anchored
tags: "web, console, profile, ssot, field-schema, parity, tui, i18n"
tier: L
related_specs: [SPEC-WEB-CONSOLE-002, SPEC-WEB-CONSOLE-003]
---

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-06-22 | manager-spec | Initial draft. Unifies the two MoAI settings surfaces (`moai web` console + `moai profile setup` TUI) onto a shared field-schema SSOT in a NEW neutral package importable by both `internal/cli` and `internal/web`. Consolidates the incremental parity work of SPEC-WEB-CONSOLE-002 (model_policy parity) and SPEC-WEB-CONSOLE-003 (flat project-config parity) onto a single source of truth. Canonical section set = 6 sections (USED-field union), both surfaces identical. Cleanup of unused config is a SEPARATE verified track — this SPEC produces a candidate LIST only, NO deletion. |

---

## §1 Context & Motivation

MoAI exposes user/project settings through **two independent surfaces**:

- **`moai profile setup`** — a terminal TUI wizard (`huh`-based) implemented in `internal/cli/profile_setup.go`. Six form sections render Select/MultiSelect/Input widgets; values persist to the profile store (`~/.moai/claude-profiles/<name>/preferences.yaml`) and sync to project config via `internal/profile/sync.go`.
- **`moai web`** — a loopback-only browser console (`templ`-based) implemented in `internal/web/`. Fieldsets render from `fieldsets.templ`; values validate in `validate.go` and persist via `internal/web/handlers.go` + `internal/web/projectconfig.go`.

The two surfaces were brought toward parity incrementally — `SPEC-WEB-CONSOLE-002` aligned model_policy validation, `SPEC-WEB-CONSOLE-003` aligned the two flat project-config scalars — but they **diverge** because each surface independently hard-codes its own field definitions, option lists, empty-option labels, and validation rules. The comment at `internal/web/validate.go:17-22` documents the structural cause: `internal/cli` cannot import `internal/web` and vice-versa, the wizard's canonical option lists are unexported, so `internal/web/validate.go` **re-declares** the same option lists by hand (`modelCanonical`, `effortLevelCanonical`). Any future drift between the two hand-maintained lists is silent.

This SPEC removes the divergence at its root: a NEW neutral package (`internal/settings`) holds the canonical field schema; **both** surfaces derive their widgets, validation, and persistence targets from it. The TUI builds `huh` widgets from the schema; the web builds `templ` fieldsets and validates from the schema.

### Confirmed user intent (NOT re-litigated)

1. Canonical section set = **6 sections** (USED-field union), both surfaces identical.
2. Cleanup of unused config is a **SEPARATE verified track** — this SPEC produces a cleanup CANDIDATE LIST only; NO deletion of any config in this SPEC.
3. Workflow = formal SPEC plan-phase (this artifact set).

### Established divergence to fix (verified — see research.md §A)

- **Statusline (16 fields): TUI-only.** The web console removed its statusline panel via `SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001` (`internal/web/fieldsets.templ:120-125`). Re-adding to web means **theme + 15 segments ONLY**; the retired `preset` selector MUST NOT be reintroduced.
- **Quality nested (3 fields) + Git nested (4 fields): web-only.** The TUI's Project section (`internal/cli/profile_setup.go:407-431`) persists only 2 scalars (`development_mode` + `git_convention`) through `persistProjectConfig`. The web persists the 7 nested fields via `writeProjectNestedConfig` (`internal/web/projectconfig.go:242`). The TUI must gain the 7 nested fields, persisting through the SAME nested-config write path.
- **Empty-option label drift** (per-field canonical text needed): lang TUI "(not set)" vs web "(unset)"; model TUI "Default (no override)" vs web "(project default)"; effort TUI verbose vs web "(runtime default)"; git_convention TUI "(project default)" vs web "(unchanged)".
- **`permission_mode` normalization**: the TUI normalizes `acceptEdits` → empty string (`profile_setup.go:443`) so the project default is not redundantly overridden; the web renders empty as "(project default)". This normalization semantic MUST be preserved.
- **Canonical-list duplication**: `internal/web/validate.go` hand-mirrors the wizard option lists from `internal/cli/profile_setup.go`. The comment at `validate.go:17-22` notes the one-way `internal/cli → internal/web` import dependency that PREVENTS importing them — which is precisely why the SSOT must live in a THIRD package both can import.

---

## §2 Requirements (GEARS)

### §2.1 Shared SSOT package

**REQ-WC10-001 (Ubiquitous):** The settings schema package shall be a NEW neutral Go package (`internal/settings`) importable by BOTH `internal/cli` and `internal/web` without either importing the other.

**REQ-WC10-002 (Ubiquitous):** The settings schema package shall define, per field, all of: logical name, value type, option list (where applicable), canonical empty-option label, validation rule reference, i18n key, and persistence target (a profile-store field name OR a `<yaml-section>.<key>` path).

**REQ-WC10-003 (Ubiquitous):** The settings schema package shall enumerate exactly the 6 canonical sections — Identity, Language, Launch, Statusline, Quality, Git Convention — covering the USED-field union of both surfaces (34 fields total).

**REQ-WC10-004 (Ubiquitous):** The settings schema package shall hold each canonical option list (model, effort level, model policy, permission mode, language, development mode, git convention) exactly once, so neither surface re-declares it.

**REQ-WC10-005 (When):** When `internal/web/validate.go` validates a submitted field, the validation outcome shall derive from the shared schema's option list / predicate rather than from a list re-declared inside `internal/web`.

### §2.2 Dual-surface derivation

**REQ-WC10-006 (When):** When the TUI wizard builds a form widget for a schema field, the widget's options and empty-option label shall be sourced from the shared schema rather than from literals inline in `internal/cli/profile_setup.go`.

**REQ-WC10-007 (When):** When the web console renders a fieldset for a schema field, the fieldset's options and empty-option label shall be sourced from the shared schema rather than from literals inline in `internal/web/fieldsets.templ`.

**REQ-WC10-008 (Ubiquitous):** Both surfaces shall render all 34 canonical fields across the 6 sections, with no field present on one surface and absent on the other (excluding the retired `preset` selector — see Exclusions).

### §2.3 Statusline re-add to web (no preset)

**REQ-WC10-009 (When):** When the web console renders the Statusline section, it shall present the theme field plus the 15 segment toggles sourced from the shared schema.

**REQ-WC10-010 (The system shall not):** The web console shall not reintroduce the retired statusline `preset` selector — neither a `preset` form control, a `preset` option list, nor a `preset` validation rule.

### §2.4 Nested project-config to TUI

**REQ-WC10-011 (When):** When the TUI wizard captures the 7 nested project-config fields (`quality.test_coverage_target`, `quality.enforce_quality`, `quality.tdd_settings.min_coverage_per_commit`, `git_convention.auto_detection.enabled`, `git_convention.auto_detection.confidence_threshold`, `git_convention.auto_detection.sample_size`, `git_convention.validation.enforce_on_push`), it shall persist them through the existing nested-config write path (`internal/web/projectconfig.go` `writeProjectNestedConfig` or a shared seam refactored from it) rather than through a parallel TUI-only YAML writer.

**REQ-WC10-012 (Where):** Where a nested project-config field is submitted empty/unset, the persistence path shall leave the existing on-disk value unchanged (empty = preserve), matching the established `writeProjectNestedConfig` per-field `*Set` semantics.

### §2.5 Label, validation, and persistence normalization

**REQ-WC10-013 (Ubiquitous):** Each schema field shall declare exactly one canonical empty-option label, and both surfaces shall render that single label for that field (eliminating the "(not set)" vs "(unset)" / "(project default)" vs "(unchanged)" drift).

**REQ-WC10-014 (Where):** Where `permission_mode` equals the project default (`acceptEdits`), the persistence layer shall normalize it to the empty string so no redundant override is written — preserving the existing TUI normalization semantic for both surfaces.

**REQ-WC10-015 (When):** When a schema field declares a persistence target, the target shall be either a profile-store `ProfilePreferences` field name OR a `<yaml-section>.<key>` project-config path, and the surface's save action shall route the value to that declared target.

### §2.6 i18n key unification

**REQ-WC10-016 (Ubiquitous):** The shared schema shall own the canonical i18n KEY set for the 6 sections and 34 fields, using the web store's live dotted prefixes (`sec.` / `f.` / `count.` / `seg.`), such that the web store consumes the keys by direct string lookup and the TUI store resolves them through the bridge resolver of REQ-WC10-017a.

**REQ-WC10-017 (Ubiquitous):** Each surface shall retain its own translation STORE in its native shape — the TUI's locale-keyed Go struct `profileSetupText` accessed via `getProfileText(locale).<NamedField>` (`internal/cli/profile_setup_translations.go`) and the web's flat `window.MOAI_I18N[locale]` dotted-key dictionary (`internal/web/assets/i18n.js`) — and this SPEC shall not force a single shared translation file across CLI and HTML.

**REQ-WC10-017a (When):** When the TUI builds a widget title or description for a schema field, it shall resolve the schema's dotted i18n key to the corresponding `getProfileText(locale).<NamedField>` value through a named bridge resolver (`schemaKeyToTUIField` or equivalent map) in `internal/cli`, because the TUI store is struct-field-addressed and cannot perform a string-keyed lookup of the schema's dotted keys directly.

### §2.7 Cleanup candidate list (LIST ONLY)

**REQ-WC10-018 (Ubiquitous):** The research artifact shall record a "Cleanup Candidates" appendix classifying config into three tiers (A keep-and-unify, B Go-unread-but-agent-consumed → KEEP, C true cleanup CANDIDATES), with NO config deleted by this SPEC.

**REQ-WC10-019 (The system shall not):** This SPEC shall not delete, rename, or alter any configuration field, section, or file outside the 6 canonical UI sections; the Tier C candidate list is advisory only and carries a mandatory pre-deletion grep gate.

---

## §3 Success Criteria (summary)

- A new `internal/settings` package exists and is imported by BOTH `internal/cli` and `internal/web`; neither imports the other.
- Both surfaces render all 34 canonical fields; the web Statusline section has theme + 15 segments and NO `preset` control.
- `internal/web/validate.go` no longer declares a model/effort/etc. option list that the schema also owns (no duplicated canonical list).
- The TUI persists the 7 nested fields through the shared/nested write seam (not a parallel writer).
- Per-field canonical empty-option labels are single-sourced; both surfaces render identical labels.
- `go test ./...` passes; `go vet ./...` and `golangci-lint run` clean; `GOOS=windows GOARCH=amd64 go build ./...` succeeds.

Full mechanically-verifiable criteria are enumerated in `acceptance.md` (AC-WC10-NNN).

---

## §4 Exclusions

This SPEC is scoped to the WHAT/WHY of unifying the two settings surfaces onto a shared schema. The following are explicitly excluded.

### Out of Scope — Configuration deletion

- No deletion, renaming, or value change of ANY configuration field, section, or file. The unused-config cleanup is a SEPARATE verified track (a future SPEC). This SPEC emits a candidate LIST only (research.md appendix), with a mandatory pre-deletion grep gate. The cleanup track is deferred precisely because "Go has no consumer" is a hypothesis, not proof of dead config — see research.md §D.

### Out of Scope — Retired statusline preset selector

- No reintroduction of the statusline `preset` selector (form control, option list, or validation rule) retired by `SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001`. The web Statusline section re-add is theme + 15 segments ONLY.

### Out of Scope — Agent/skill-consumed configuration

- No change to the Tier B configuration that Go does not read but the agent/skill/rule (markdown) layer consumes (`workflow.*` token_budget + team role_profiles, `harness.*`, `llm.glm.*`, `language.yaml` directives, `design.*`, `interview.*`, `context_search.*`). Surfacing these in a UI is out of scope.

### Out of Scope — docs-site and user documentation

- No docs-site work. This SPEC modifies internal CLI/web Go code under `internal/`, not user-facing product documentation. Template-neutrality constraints (CLAUDE.local.md §15/§25) apply only to `internal/template/templates/` and therefore do NOT apply to this SPEC's `internal/settings`, `internal/cli`, `internal/web` targets.

### Out of Scope — new settings beyond the USED union

- No introduction of new settings fields beyond the 34-field USED union. Adding entirely new configurable surfaces (e.g. exposing `workflow.*` levers in the UI) is a future SPEC, not this one.
