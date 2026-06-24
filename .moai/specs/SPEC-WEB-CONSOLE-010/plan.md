# Implementation Plan — SPEC-WEB-CONSOLE-010

> Derived plan for the shared field-schema SSOT unification. WHAT/WHY lives in `spec.md`; HOW lives here. No time estimates — priority-ordered milestones only.

## §A Context

Two settings surfaces (`moai profile setup` TUI, `moai web` console) independently hard-code field definitions, option lists, empty-option labels, and validation rules. The hard import constraint (`internal/cli` ⇎ `internal/web` cannot import each other) forces `internal/web/validate.go` to hand-mirror the wizard's unexported option lists. This plan introduces a THIRD neutral package (`internal/settings`) that both surfaces import, then refactors each surface to derive from it.

The plan is run-phase guidance for `manager-develop`. It names files, seams, and ordering but defers function signatures and struct field names to the Run phase (those are HOW-of-HOW, decided during implementation against the live tree).

## §B Known Issues / Constraints (load-bearing)

- **B1 — Import direction is one-way and immutable.** `internal/cli → internal/web` is the only legal direction (documented at `internal/web/validate.go:17-22`). The SSOT MUST live in a package neither imports — `internal/settings`. Verify with a dependency check that `internal/settings` imports neither `internal/cli` nor `internal/web`.
- **B2 — `templ` generates Go.** `internal/web/fieldsets.templ` compiles to `fieldsets_templ.go`. Any `.templ` edit requires re-running `templ generate` (or `make build` if wired) before `go build`. The generated `*_templ.go` is compiled Go, not an embedded data asset.
- **B3 — Retired preset is a hard prohibition.** `SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001` retired the statusline `preset`. The schema MUST NOT carry a `preset` field and the web re-add MUST NOT add a `preset` control. (Grep guard in acceptance.md.)
- **B4 — Nested write seam already exists and is correct.** `writeProjectNestedConfig` (`internal/web/projectconfig.go:242`) does whole-section-copy + per-field `*Set` mutation through the config manager. The TUI must reuse this seam (relocated to a shared/importable location) — NOT author a parallel YAML writer. The TUI's current `persistProjectConfig` writes only 2 scalars; it must be extended/replaced to also drive the nested path.
- **B5 — `permission_mode` normalization is a semantic, not a cosmetic.** `acceptEdits` → empty string at `profile_setup.go:443`. Preserve it for both surfaces so the saved YAML never carries a redundant `acceptEdits` override.
- **B6 — Two translation stores, one key set.** TUI = Go tables (`profile_setup_translations.go`); web = `window.MOAI_I18N` in `assets/i18n.js`. The schema owns KEYS; each store stays its own format. Do NOT collapse to one file.

## §C Pre-flight (run-phase entry checks)

1. `internal/settings` and `internal/settings/schema` do not yet exist (confirmed plan-phase). Re-confirm at run-phase entry.
2. `go test ./internal/cli/... ./internal/web/... ./internal/profile/...` is green on the baseline before any change (capture the baseline pass count per verification-claim-integrity §2).
3. `templ` toolchain availability (`which templ` or the `make build` templ step) for B2.
4. Capture LSP baseline (errors/warnings) at phase start per the plan-phase LSP gate.

## §D Constraints (TRUST 5 + project)

- Go ≥ 1.23; `context.Context` first-param where blocking; explicit error wrapping (`fmt.Errorf("...: %w", err)`).
- Coverage ≥ 85% per package; `internal/settings` (new) targets ≥ 90% (foundational schema package).
- No hardcoded option strings duplicated across packages (the whole point — single SSOT).
- Cross-platform: `GOOS=windows GOARCH=amd64 go build ./...` must pass.
- Subagent boundary (C-HRA-008): no `AskUserQuestion` / `mcp__askuser` in `internal/cli`, `internal/web`, `internal/settings`.

## §E Self-Verification

See `progress.md` §E for the audit-ready signal skeleton. Run-phase evidence (E1 AC matrix, E2 cross-platform build, E3 coverage, E4 subagent-boundary grep, E5 lint, E6 push) is populated by `manager-develop` at run-phase, NOT here.

## §F Milestones (priority-ordered)

> Ordering is dependency-driven. Each milestone is a `manager-develop` delegation unit. No time estimates.

### M1 — SSOT schema package skeleton (Priority: High)

- Create `internal/settings` (package `settings`). Define the field-definition type carrying: logical name, type, option list, canonical empty-option label, validation-rule reference, i18n key, persistence target. Define the 6-section enumeration and the 34-field schema as data.
- Move the canonical option lists (model, effort, model policy, language, development mode, git convention) into the schema as the single source. Reuse existing exported predicates where importable (`profile.IsValidPermissionMode`, `template.IsValidModelPolicy`, `models.ValidDevelopmentModes`, `config.IsValidConvention`) — the schema references them, does not re-implement them.
- Tests: schema completeness (34 fields / 6 sections), per-field invariants (every field has a non-empty i18n key + persistence target), option-list uniqueness.
- Dependency check: `internal/settings` imports neither `internal/cli` nor `internal/web`.

### M2 — Relocate the nested-config write seam to a shared location (Priority: High)

- Extract / relocate the nested project-config read+write seams (`readProjectNestedConfig` / `writeProjectNestedConfig` + the `projectNestedForm` shape's persistence logic) so BOTH the TUI and the web can call them without `internal/cli` importing `internal/web`. Target: `internal/settings` (or a sibling neutral package) for the persistence seam; keep the HTTP-form parsing (`parseProjectNestedForm`) in `internal/web` (it is web-request-specific).
- Preserve the whole-section-copy + per-field `*Set` semantics (B4) byte-for-behavior. Existing web tests for the nested write path must stay green.

### M3 — TUI derives from schema + gains 7 nested fields (Priority: High)

- Refactor `internal/cli/profile_setup.go` to build the `huh` Select/MultiSelect/Input widgets from the M1 schema (options + empty-option label sourced from schema, not inline literals).
- Add the 7 nested project-config fields to the TUI Project section; persist them through the M2 shared seam (replacing/extending `persistProjectConfig`'s 2-scalar-only behavior). Preserve empty=preserve (REQ-WC10-012) and `permission_mode` normalization (B5).
- Add the named i18n bridge resolver `schemaKeyToTUIField(schemaKey, profileSetupText) string` (or an equivalent `map[schemaKey]func(profileSetupText) string`) in `internal/cli` (REQ-WC10-017a, design §F.2): the TUI store is the struct-field-addressed `profileSetupText` accessed via `getProfileText(locale).<NamedField>` and CANNOT string-look-up the schema's dotted keys. The TUI builds widget titles/descriptions THROUGH this bridge. `profile_setup_translations.go` keeps its struct-of-strings shape unchanged (translation values stay Go-struct form); only the access path is bridged.
- **AUTHOR the 7 new nested-field translations in the TUI store (D7 — TUI side).** The TUI has never had a Project nested section, so `profileSetupText` lacks struct fields for the 7 nested project fields. Add the new `profileSetupText` struct fields (title + desc per field) for all 7 (`quality.test_coverage_target`, `quality.enforce_quality`, `quality.tdd_settings.min_coverage_per_commit`, `git_convention.auto_detection.{enabled,confidence_threshold,sample_size}`, `git_convention.validation.enforce_on_push`), populate them in ALL 4 locale tables (en/ko/ja/zh) in `profile_setup_translations.go`, and register their bridge-key mappings in `schemaKeyToTUIField`. Without this, AC-WC10-016 (`TestI18nKeySetParity`) fails for the TUI side. Re-key-only (no authoring) is insufficient — these struct fields do not exist yet.

### M4 — Web derives from schema + re-adds statusline (no preset) (Priority: High)

- Refactor `internal/web/fieldsets.templ` to source options + empty-option labels from the M1 schema; re-run `templ generate` (B2).
- Re-add the Statusline section to the web console: theme + 15 segment toggles, sourced from the schema. NO `preset` control (B3).
- Refactor `internal/web/validate.go` to validate from the schema's option lists / predicates — remove ALL 5 hand-mirrored declarations (`langOptions`, `modelCanonical`, `effortLevelCanonical`, `developmentModeCanonical`, `conventionCanonical`) so the schema is the sole source (REQ-WC10-004/005; AC-WC10-012 greps all 5). Route the web statusline submit through the existing `internal/profile/sync.go` statusline write (the config-layer plumbing already exists per `fieldsets.templ:124`).
- Update `assets/i18n.js` (`window.MOAI_I18N`) field/section/segment keys to match the schema's canonical key set (live prefixes `f.` / `sec.` / `count.` / `seg.` — NOT `field.`); the web consumes these by direct string lookup.
- **AUTHOR the 15 `seg.<segment>` keys in `i18n.js` × 4 locales (D7 — web side).** The statusline panel was removed from the web (PRESET-RETIRE-001), so `i18n.js` currently has ONLY `seg.note` + `seg.title` — the 15 per-segment label keys (`seg.claude_version`, `seg.context`, `seg.directory`, `seg.effort_thinking`, `seg.git_branch`, `seg.git_status`, `seg.moai_version`, `seg.model`, `seg.output_style`, `seg.pr`, `seg.session_time`, `seg.task`, `seg.usage_5h`, `seg.usage_7d`, `seg.worktree`) are MISSING. Author all 15 in ALL 4 locale blocks (en/ko/ja/zh). Without this, AC-WC10-016 fails for the web statusline segment labels.
- **VERIFY (do NOT re-author) the nested Project field keys in `i18n.js` × 4 locales (D7 — web side).** The 16 nested-field web keys (`f.quality.test_coverage_target.title/desc`, `f.quality.enforce_quality.title/desc`, `f.quality.tdd_settings.min_coverage_per_commit.title/desc`, and the 4 `f.git_convention.*` nested `.title/.desc`) ALREADY EXIST in `i18n.js` (verified plan-phase). M4 confirms their presence across all 4 locales rather than re-authoring; if a locale is missing one, backfill only the gap. (This corrects the audit-relayed assumption that these keys were absent — only the `seg.*` keys and the TUI struct fields are genuinely net-new.)

### M5 — Label/persistence normalization sweep (Priority: Medium)

- Verify per-field canonical empty-option labels render identically on both surfaces (REQ-WC10-013). Resolve the four documented drifts (lang, model, effort, git_convention) by single-sourcing through the schema.
- Verify each field's persistence target (profile-store field vs yaml section.key) is honored by both surfaces' save actions (REQ-WC10-015).

### M6 — Verification + research.md cleanup-candidate appendix finalization (Priority: Medium)

- Run the full read-only verification batch (test, coverage, subagent-boundary grep, statusline-preset-absence grep, duplicated-list-absence grep, cross-platform build, lint).
- Confirm the research.md Cleanup Candidates appendix (Tier A/B/C) is accurate against the live tree (the cleanup itself is a SEPARATE SPEC — this milestone only confirms the candidate list, deletes nothing).

## §G Anti-Patterns (do NOT do)

- **AP-1**: Putting the SSOT in `internal/web` or `internal/cli` — defeats the import constraint (B1). It MUST be the third neutral package.
- **AP-2**: Authoring a parallel TUI-only YAML writer for the 7 nested fields — reuse the M2 shared seam (B4).
- **AP-3**: Re-adding any `preset` control to the web statusline (B3).
- **AP-4**: Collapsing the two translation stores into one shared file — only the KEY set is shared (B6).
- **AP-5**: Deleting/altering any Tier C cleanup candidate in this SPEC — cleanup is a separate verified track with a grep gate (spec.md §4).
- **AP-6**: Editing `internal/template/templates/**` — out of scope; this SPEC is moai-adk's own `internal/` code.
- **AP-7**: Leaving a duplicated canonical option list in `internal/web/validate.go` after M4 — the schema must be the sole source (REQ-WC10-004/005).

## §H Cross-References

- `spec.md` — REQ-WC10-001..019 (WHAT/WHY).
- `acceptance.md` — AC-WC10-NNN mechanically-verifiable criteria.
- `design.md` — SSOT package layout, field-def struct shape, dual-surface render flow, i18n key strategy.
- `research.md` — established inventory, import-constraint rationale, Cleanup Candidates appendix (Tier A/B/C).
- `SPEC-WEB-CONSOLE-002` — model_policy parity (consolidated here).
- `SPEC-WEB-CONSOLE-003` — flat project-config parity (consolidated here); its own forward-reference to "S2b (Tier L)" anticipates this SPEC.
- `internal/cli/CLAUDE.md` — CLI module conventions (subagent boundary, exit codes, cross-platform).
