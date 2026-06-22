# Design — SPEC-WEB-CONSOLE-010

> Tier L design artifact. Describes the SSOT package layout, the field-definition shape, the dual-surface render flow, the persistence-target routing, and the i18n key strategy. This is design intent (the shape of the solution), NOT final code — function signatures and exact struct field names are settled in the Run phase against the live tree.

## §A The import-constraint problem and the third-package solution

### A.1 The constraint

```
internal/cli ──(may import)──▶ internal/web      ← legal, one-way
internal/web ──(MUST NOT)────▶ internal/cli      ← illegal, would cycle
```

The wizard's canonical option lists live unexported in `internal/cli/profile_setup.go`. `internal/web/validate.go` cannot import them, so it re-declares them by hand (`modelCanonical`, `effortLevelCanonical`) — documented at `validate.go:17-22`. Hand-mirrored lists drift silently.

### A.2 The solution — a neutral third package

```
            internal/settings   ← NEW, neutral SSOT
           ╱                 ╲
internal/cli              internal/web
 (TUI: huh)                (web: templ)
```

`internal/settings` imports neither `internal/cli` nor `internal/web`. It MAY import the already-neutral packages both surfaces already use: `internal/profile` (for `ProfilePreferences` field names + `IsValidPermissionMode`), `internal/config` (for `IsValidConvention` + the section structs), `pkg/models` (for `ValidDevelopmentModes`, the config section types), `internal/template` (for `ValidModelPolicies` / `IsValidModelPolicy`), and `internal/statusline` (for the segment-key constants). The dependency check AC-WC10-003 enforces no reverse edge.

> Note: `internal/profile` already imports `internal/config`, `internal/statusline`, `pkg/models` (per `sync.go`), so `internal/settings` sits at a layer those neutral packages already form — no new cycle is introduced.

## §B Field-definition shape (design intent)

The schema is **data**, not behavior. Each canonical field is one record. The record's shape (illustrative — final Go names decided at run-phase):

```
FieldDef {
    Name           string        // logical field name, e.g. "model", "quality.test_coverage_target"
    Section        SectionID     // one of the 6 canonical sections
    Type           FieldType     // scalar-select | multi-select | text | int | float | bool
    Options        []OptionDef   // for select/multi-select; nil otherwise. OptionDef{Value, I18nKey}
    EmptyLabel     string        // the ONE canonical empty-option label (resolves the 4 drifts)
    EmptyLabelKey  string        // i18n key for the empty-option label
    Validate       func(string) bool  // OR a reference to an existing exported predicate; nil = always-valid
    I18nKey        string        // the shared key both stores resolve
    Persist        PersistTarget // where the value lands (see §D)
}
```

- `Options[].Value` is the canonical wire value (e.g. `"opus"`, `"xhigh"`); `Options[].I18nKey` resolves the human label per surface.
- `Validate` reuses existing exported predicates wherever importable: `profile.IsValidPermissionMode`, `template.IsValidModelPolicy`, `models.DevelopmentMode(v).IsValid`, `config.IsValidConvention`. For pure membership (model, effort, language) the schema's own `Options` list is the predicate (membership test), eliminating the hand-mirrored lists.
- The schema exposes accessors: `SectionFields(SectionID) []FieldDef`, `AllFields() []FieldDef`, `Field(name) (FieldDef, bool)`, and per-type option accessors so each surface iterates uniformly.

## §C The 6 canonical sections and 34 fields

| # | Section | Fields (count) | Persistence target |
|---|---------|----------------|--------------------|
| 1 | Identity | `user_name` (1) | profile store → synced `user.yaml` |
| 2 | Language | `conversation_lang`, `git_commit_lang`, `code_comment_lang`, `doc_lang` (4) | profile store → synced `language.yaml` |
| 3 | Launch | `model`, `model_policy`, `effort_level`, `permission_mode` (4) | profile store → `settings.local.json` + launch env |
| 4 | Statusline | `statusline_theme` + 15 segments (`claude_version`, `context`, `directory`, `effort_thinking`, `git_branch`, `git_status`, `moai_version`, `model`, `output_style`, `pr`, `session_time`, `task`, `usage_5h`, `usage_7d`, `worktree`) (16) | profile store → synced `statusline.yaml` |
| 5 | Quality | `quality.development_mode` *(the flat development_mode scalar shared with SPEC-WEB-CONSOLE-003)*, `quality.test_coverage_target`, `quality.enforce_quality`, `quality.tdd_settings.min_coverage_per_commit` | `.moai/config/sections/quality.yaml` |
| 6 | Git Convention | `git_convention.convention`, `git_convention.auto_detection.enabled`, `git_convention.auto_detection.confidence_threshold`, `git_convention.auto_detection.sample_size`, `git_convention.validation.enforce_on_push` | `.moai/config/sections/git-convention.yaml` |

**34-field reconciliation (USED union):** Identity 1 + Language 4 + Launch 4 + Statusline 16 + Quality (`development_mode` 1 + 3 nested) + Git Convention (`convention` 1 + 4 nested) = 1+4+4+16+4+5 = **34**. (`development_mode` and `git_convention.convention` are the two flat scalars already shared by SPEC-WEB-CONSOLE-003; the 7 nested fields are the 3 quality + 4 git-convention that are currently web-only.)

The 15 segment keys are sourced from `internal/statusline.Segment*` constants (the SSOT the CLI and sync already use; `SegmentRepo` is the 16th constant, intentionally outside the 15-key schema per `sync.go:154`).

## §D Persistence-target routing

Two distinct persistence boundaries exist (this is intrinsic — not a defect):

```
PersistTarget = ProfileStore{field}              → preferences.yaml (+ SyncToProjectConfig for user/lang/statusline)
              | ProjectConfig{section, key}       → quality.yaml / git-convention.yaml via config manager
```

- **Profile-store fields** (Identity, Language, Launch, Statusline) route through `profile.WritePreferences` + `profile.SyncToProjectConfig` — the existing first boundary (`handlers.go handleSave` boundary 1; TUI `WritePreferences` + `SyncToProjectConfig`).
- **Project-config fields** (Quality, Git Convention) route through the config-manager seam — the existing second boundary (`writeProjectConfig` for the 2 scalars + `writeProjectNestedConfig` for the 7 nested). M2 relocates the nested seam to a neutral location so the TUI can call it without importing `internal/web`.
- `permission_mode` carries a normalization hook in its `Persist`: `acceptEdits → ""` (REQ-WC10-014, preserving `profile_setup.go:443`).

## §E Dual-surface render flow

### E.1 TUI (huh) flow

```
schema.SectionFields(section) → for each FieldDef:
    select  → huh.NewSelect().Options(fieldDef.Options + EmptyLabel option).Value(&binding)
    multi   → huh.NewMultiSelect().Options(fieldDef.Options).Value(&selection)
    text    → huh.NewInput().Value(&binding)
    int/float/bool → huh widget per type
on submit:
    for each FieldDef: route value to fieldDef.Persist
    (profile-store fields → WritePreferences/Sync; project-config → shared config seam)
```

The TUI today hard-codes each `huh.NewOption(...)` block (`profile_setup.go:314-430`). After M3 these blocks iterate the schema's `Options`.

### E.2 Web (templ) flow

```
newPageView() → view-model carries schema-derived option lists (no hand-mirrored modelCanonical)
fieldsets.templ → for each FieldDef in section:
    @optSelect(Options: fieldDef.Options, Empty: fieldDef.EmptyLabel, ...)
    statusline section → theme optSelect + 15 segment toggles (schema-driven; NO preset)
handlers.go handleSave → parse form → validatePrefs/validateProjectNestedConfig
    (validation derives from schema option lists / predicates) → route to fieldDef.Persist
```

`templ generate` re-runs after the `.templ` edit (B2). The web statusline submit reuses `internal/profile/sync.go`'s statusline write (the config-layer plumbing already survives the panel retirement per `fieldsets.templ:124`).

### E.3 Why both can share without a cycle

Each surface depends on `internal/settings` (down-edge). `internal/settings` depends only on the already-neutral leaf packages. No surface depends on the other. The view-model assembly (`newPageView`) and the `huh` form assembly are the per-surface adapters that translate `[]FieldDef` into the surface's native widget vocabulary.

## §F i18n key strategy

### F.1 The two stores have DIFFERENT shapes (verified against the live code)

The two surfaces do NOT share a string-keyed lookup mechanism. This was verified by reading the live stores:

- **Web** (`internal/web/assets/i18n.js`): a FLAT, string-keyed `window.MOAI_I18N[locale]` dictionary with dotted keys under the prefixes `f.` (field), `sec.` (section), `count.` (section field-count), and `seg.` (statusline segment) — e.g. `"f.model.title"`, `"sec.language.title"`, `"count.project"`, `"seg.git_branch"`. The `.templ` files carry `data-i18n="<dotted-key>"` markers; `applyI18n()` in `app.js` swaps element text by key. (Note the prefix is `f.`, **not** `field.` — an earlier draft of this design used `field.` incorrectly.)
- **TUI** (`internal/cli/profile_setup_translations.go`): a `map[string]profileSetupText` (locale → STRUCT), where `profileSetupText` is a Go struct with NAMED fields (`ModelOverrideTitle`, `EffortLevelDesc`, …) accessed via `getProfileText(locale).<NamedField>`. There is NO inner string-keyed lookup — a schema key like `f.model.title` will NEVER appear as a literal in this file. The struct field names, not dotted strings, are the access path.

Because the TUI store is struct-field-addressed and the web store is string-key-addressed, the schema cannot simply "hold a key both stores look up." Instead:

### F.2 The schema owns a canonical key set; the TUI binds via a named bridge resolver

- The schema owns the canonical dotted KEY namespace, matching the web's live prefixes: section keys (`sec.identity.title`, …), field keys (`f.model.title`, `f.model.desc`, `f.model.empty`, …), section field-count keys (`count.project`, …), and segment keys (`seg.claude_version`, …). The web store consumes these keys directly (string lookup) — no adapter needed.
- The TUI cannot string-look-up these keys, so M3 adds a small **named bridge resolver** — `schemaKeyToTUIField(schemaKey, profileSetupText) string` (or an equivalent `map[schemaKey]func(profileSetupText) string` table) in `internal/cli` — that maps each schema field key to the corresponding `getProfileText(locale).<NamedField>` accessor. The TUI builds its widget titles/descriptions THROUGH this bridge, so the schema key set remains the single source of the key namespace while the TUI keeps its existing struct-of-strings translation store unchanged.

### F.3 Two stores, one key set (no shared translation file)

- This SPEC does NOT introduce a single shared translation FILE (REQ-WC10-017). The TUI store stays a Go struct-of-strings (`profileSetupText`); the web store stays a flat JS dict (`window.MOAI_I18N`). Forcing one cross-format file (Go struct ⇄ JS dict) would couple two unrelated render runtimes.
- The schema's canonical key set is the contract both stores satisfy: the web satisfies it by string lookup; the TUI satisfies it through the named bridge resolver (F.2). AC-WC10-016 (`TestI18nKeySetParity`) asserts every schema key resolves to a non-empty value in BOTH stores for all 4 locales — for the TUI it asserts resolution THROUGH the bridge (not by expecting a `f.*` literal in the struct file), and for the web it asserts a matching flat dotted-key entry.

## §G Risks and mitigations

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Schema refactor breaks an existing TUI/web test that pinned an inline literal | Medium | M1 builds the schema first with its own tests; M3/M4 migrate consumers incrementally; full suite green per milestone |
| `templ generate` skipped → stale generated Go | Medium | M4 explicit re-generate step; AC-WC10-004/007 catch staleness |
| Nested-seam relocation (M2) subtly changes the whole-section-copy semantics | Low-Med | Preserve `writeProjectNestedConfig` behavior byte-for-behavior; existing web nested tests must stay green (B4) |
| Accidental `preset` reintroduction during statusline re-add | Low | AC-WC10-011b grep guard; design forbids a `preset` FieldDef entirely |
| i18n key added to schema but missing from one store | Medium | AC-WC10-016 (`TestI18nKeySetParity`) asserts every key resolves in BOTH stores × 4 locales — web by direct string lookup, TUI through the named bridge resolver (F.2) |
| Re-key-only migration leaves new keys unauthored → AC-WC10-016 fails | Medium-High | Plan M3/M4 schedule AUTHORING the net-new keys, not just re-keying: web side authors the 15 `seg.<segment>` keys × 4 locales in `i18n.js` (only `seg.note`/`seg.title` exist today); TUI side authors the 7 nested-field `profileSetupText` struct fields × 4 locales + their `schemaKeyToTUIField` bridge mappings (the TUI never had a Project nested section). The 16 nested-field WEB keys already exist (verified) — M4 verifies their 4-locale completeness rather than re-authoring. |
| TUI store is struct-addressed, not key-addressed → schema key never a literal there | (design-fidelity) | M3 named bridge resolver (`schemaKeyToTUIField`) maps each schema key to a `getProfileText().<Field>` accessor; AC-WC10-016 asserts resolution through the bridge, not by grepping for an `f.*` literal in the struct file |
| A field's persistence target mis-routed (profile vs project-config) | Low | AC-WC10-015 + per-field `Persist` declaration is the single routing authority |

## §H Cross-references

- `research.md` §A (established inventory), §C (import-constraint rationale), §D (Cleanup Candidates).
- `plan.md` §F (milestones M1-M6 mapping to this design's package/seam/flow).
- `internal/web/validate.go:17-22` — the documented import-constraint comment this design resolves.
- `internal/web/projectconfig.go:242` (`writeProjectNestedConfig`) — the nested seam relocated in M2.
- `internal/profile/sync.go` (`syncStatusline`, `defaultStatuslineSegments`) — the config-layer statusline plumbing the web re-add reuses.
