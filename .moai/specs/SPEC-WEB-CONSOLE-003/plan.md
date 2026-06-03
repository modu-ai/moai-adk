# Implementation Plan — SPEC-WEB-CONSOLE-003

> S2a of the `web-console-v3` cohort. **Tier M** (3 artifacts: spec.md + plan.md + acceptance.md). Reads with `spec.md` (REQs §2, Exclusions §4, verified paths §5) and `acceptance.md` (full Given-When-Then AC enumeration).

## §A Context

`SPEC-WEB-CONSOLE-001` shipped `moai web` (loopback profile-settings CRUD + a fixed user/language/statusline project-config sync). `SPEC-WEB-CONSOLE-002` (S1) added web↔TUI validation parity for 3 profile fields and set the default port to 3041. S2a extends BOTH editors (web console + TUI wizard) to surface **two flat project-config enum settings** — `development_mode` (`quality.yaml`) and `git_convention.convention` (`git-convention.yaml`) — reaching functional parity. Unlike S1's profile-store settings, these are **project-config** settings, so S2a introduces a thin project-config write path (a bounded extension of the `SyncToProjectConfig` pattern over the existing config-manager `LoadRaw`/`SetSection`/`Save` API). Deeply-nested config (full quality/workflow/harness/git-strategy + nested git-convention/llm sub-fields) is **S2b (Tier L)**; `llm.mode`/`llm.default_model` were narrowed OUT (backend-switch toggle / legacy enum-less string — see spec.md §1).

### A.1 Tier justification (M, not S)

Classified **Tier M** because: (a) the project-config write path is a **genuinely new persistence path** crossing the `internal/web → internal/config` package boundary (not present in `ProfilePreferences`/`SyncToProjectConfig`), (b) ~6-7 files affected (web validate/handlers/app/template + cli profile_setup/translations + one canonical predicate), (c) dual-editor (web + TUI) × (validate + widget + persistence + 4-locale labels + read-on-render) creates enough AC surface (11 ACs) to warrant a separate `acceptance.md`. LOC estimate ~280-400.

### A.2 Branch / artifacts

- Project root: `/Users/goos/MoAI/moai-adk-go`
- Plan-phase branch: `docs/glm-webtool-routing-m1-m5` (HEAD == origin/main, divergence 0 0). Main checkout, NO worktree (Step 1 discipline).
- SPEC artifacts: `.moai/specs/SPEC-WEB-CONSOLE-003/{spec.md, plan.md, acceptance.md, progress.md}`
- development_mode: `tdd` (cycle_type=tdd for run-phase — write the failing validate/handler/render/persist test first, then minimal change).

### A.3 PRESERVE (do not break)

- `internal/web/server.go` loopback-only bind (127.0.0.1)
- `internal/web/app.go` hostCheckMiddleware (no-auth, Host-header write-safety) + the 3 existing injectable seams
- `internal/web/handlers.go` MX:WARN persistence boundary (no direct YAML marshal)
- `internal/web/integration_test.go` DO_NOT_TOUCH sentinels (workflow/harness/git-strategy)
- All S1 (`model`/`effort_level`/`model_policy`) + S1 port-3041 behavior + existing profile sync path
- The config-manager round-trip MUST rewrite the non-targeted fields of quality/git-convention unchanged

### A.4 EXTEND

- `internal/web`: new project-config read+write seams on `app`; new view-model option lists + current-value fields; new "Project" fieldset; new project-config validator
- `internal/cli/profile_setup.go` + `profile_setup_translations.go`: 2 new selects + 4-locale labels + project-config write on save
- ONE canonical predicate: add exported `IsValidConvention`/`ValidConventions` (preferred) to `internal/config` or `pkg/models` (development_mode already has `models.ValidDevelopmentModes()`)

## §B Known Issues / Pre-flight

- **B1 — Package boundary (import direction)**: `internal/web` already imports `internal/profile` and `internal/template`. It does NOT yet import `internal/config`. The project-config write path requires `internal/web → internal/config`. Verify this import is legal (no cycle): `internal/config` does NOT import `internal/web` (confirmed — config is a leaf consumed by everything). Add the import to `internal/web/handlers.go` (or a new `internal/web/projectconfig.go`).
- **B2 — `validatePrefs` signature is ProfilePreferences-bound**: the existing `validatePrefs(p profile.ProfilePreferences)` cannot validate `development_mode`/`git_convention` because they are not `ProfilePreferences` fields. Author a **separate** validator (e.g. `validateProjectConfig(devMode, convention string) map[string]string`) — do NOT shoehorn the two project-config fields into `ProfilePreferences` (that would corrupt the profile schema and break `SyncToProjectConfig`). The handler runs BOTH validators and merges the FieldErrors maps before deciding 400-vs-persist.
- **B3 — convention predicate is unexported**: `validGitConventionNames` (`internal/config/validation.go:130`) is unexported. Preferred fix: add `func IsValidConvention(string) bool` + `func ValidConventions() []string` in `internal/config` (next to the map, reusing it) OR in `pkg/models` (next to `GitConventionConfig`). `development_mode` already has the exported `models.ValidDevelopmentModes()` + `DevelopmentMode.IsValid()` — reuse directly, no new predicate.
- **B4 — empty-value semantics**: empty submitted value = "keep existing project-config value" (mirrors `SyncToProjectConfig`'s non-empty-only overwrite). The load-mutate-save cycle MUST NOT clobber a non-empty persisted value with an empty submission. Both validators treat empty as valid (no error).
- **B5 — TUI persistence is profile-only today**: `internal/cli/profile_setup.go:447-474` writes `ProfilePreferences` via `WritePreferences` + `SyncToProjectConfig`. The 2 new TUI values are project-config, so the wizard must call the SAME project-config write path (the new `internal/config` round-trip helper) AFTER the existing profile write — NOT add them to the `ProfilePreferences{...}` literal at line 447.
- **B6 — translations parity test guards us**: `TestGetProfileText_AllLanguages` (`profile_setup_translations_test.go:5`) asserts label fields are non-empty for all 4 locales. Add the new label fields to all 4 locale blocks (en/ko/ja/zh) AND extend the test's non-empty assertions — this is the AC-WC3-006a guard working in our favor.
- **B7 — web assets embed (no make build)**: `page.html.tmpl` is embedded via the web package's own `go:embed` (S1 §4.8 verified). No `make build`/embedded-mirror parity step. Pre-flight: `grep -rn "go:embed" internal/web/`.
- **B8 — config-manager state machine**: `SetSection` returns `ErrNotInitialized` if `Load`/`LoadRaw` was not called first. The write helper MUST `LoadRaw(projectRoot)` before `SetSection`. Verify the helper handles a fresh/absent config dir gracefully (LoadRaw behavior).
- **B11 — AskUserQuestion forbidden**: both `internal/web` and `internal/cli` run in subagent context. No `AskUserQuestion`/`mcp__askuser`. Blocker → structured report (orchestrator handles). The CLI subagent-boundary guard (`TestNew_NoAskUserQuestion` style) applies — no interactive prompt added.

## §C Pre-flight verification batch (run before M1)

```bash
# 1. Confirm current project-config values
grep -n "development_mode" .moai/config/sections/quality.yaml
grep -n "convention" .moai/config/sections/git-convention.yaml
# 2. Confirm exported development_mode predicate + git-convention enum SSOT
grep -n "ValidDevelopmentModes\|func.*IsValid\|oneof=auto conventional" pkg/models/config.go
# 3. Confirm no exported convention predicate yet (we add one)
grep -rn "func IsValidConvention\|func ValidConventions" internal/ pkg/ | grep -v _test.go || echo "none — add one"
# 4. Confirm SetSection supports quality + git_convention
grep -n 'case "quality"\|case "git_convention"' internal/config/manager.go
# 5. Confirm web does NOT yet import internal/config (we add it)
grep -rn "modu-ai/moai-adk/internal/config" internal/web/ || echo "not imported yet"
# 6. Confirm web assets embed mechanism (not template-deploy)
grep -rn "go:embed" internal/web/
# 7. Confirm integration_test DO_NOT_TOUCH sentinels exist
grep -n "DO_NOT_TOUCH\|workflow.yaml\|harness.yaml\|git-strategy" internal/web/integration_test.go | head
# 8. Confirm TUI translations parity test + persistence path
grep -n "TestGetProfileText_AllLanguages\|WritePreferences\|SyncToProjectConfig" internal/cli/profile_setup_translations_test.go internal/cli/profile_setup.go | head
go test ./internal/web/... ./internal/cli/... ./internal/config/... 2>&1 | tail -5   # baseline green
```

## §D Constraints

- [HARD] Reuse canonical enums/predicates: `models.ValidDevelopmentModes()` for development_mode; add ONE exported `IsValidConvention`/`ValidConventions` for git-convention (or mirror with MX:NOTE → pkg/models oneof SSOT). No parallel/invented rule-set. No fresh enum for any enum-less field (this is why llm.default_model is excluded).
- [HARD] No new config section. Write ONLY `quality` (development_mode) + `git_convention` (convention) via `SetSection`/`Save` (REQ-WC3-007).
- [HARD] No direct `yaml.Marshal`/`os.WriteFile` of a config section from `internal/web` — persistence via `config.NewConfigManager()`/`LoadRaw`/`SetSection`/`Save` only (REQ-WC3-008).
- [HARD] Do NOT add development_mode/convention to `ProfilePreferences` — they are project-config, not profile (B2/B5). Author a separate validator + separate persistence helper.
- [HARD] Preserve all 001/002 invariants (loopback, no-auth, Host-header, integration_test DO_NOT_TOUCH sentinels intact).
- [HARD] Scope-fence: nested git-convention/llm sub-fields + full quality/workflow/harness/git-strategy (S2b), web i18n + webfont (S3), dead-config (S4), llm.mode/llm.default_model (narrowed out) are ALL OUT.
- [HARD] No `--no-verify`, no `--amend`, no settings.json touch, no parallel-session deletion touch.
- Go ≥ 1.23 conventions (`.claude/rules/moai/languages/go.md`): explicit error wrapping, table-driven tests, ≥85% package coverage on changed packages.

## §E Self-Verification (closure gate — mirrors acceptance.md §D)

- [ ] `go test ./internal/web/... ./internal/cli/... ./internal/config/...` exit 0 (AC-WC3-009)
- [ ] `go vet ./internal/web/... ./internal/cli/... ./internal/config/...` clean
- [ ] `golangci-lint run ./internal/web/... ./internal/cli/... ./internal/config/...` clean
- [ ] `GOOS=windows GOARCH=amd64 go build ./...` exit 0
- [ ] Web rejects bogus + persists valid `development_mode` (AC-WC3-001a/001b)
- [ ] Web rejects bogus + persists valid `git_convention` (AC-WC3-002a/002b)
- [ ] Both fields render `<select>` in a "Project" fieldset, no `type="text"` (AC-WC3-003)
- [ ] `GET /` pre-selects current project-config values; read failure → readable inline error (AC-WC3-004)
- [ ] Project-config write is a new injectable seam; empty value keeps existing; no direct marshal (AC-WC3-005)
- [ ] TUI exposes both selects + 4-locale labels non-empty (AC-WC3-006a)
- [ ] TUI save persists to project config, not preferences.yaml (AC-WC3-006b)
- [ ] Write path touches ONLY quality + git_convention; others untouched (AC-WC3-007)
- [ ] 001/002 invariants green; integration_test DO_NOT_TOUCH sentinels intact (AC-WC3-008)
- [ ] C-HRA-008: no AskUserQuestion in internal/web or internal/cli new code

## §F Milestones (priority order — no time estimates)

### M1 — Canonical convention predicate (Priority: High; REQ-WC3-002 support)

- Add exported `func IsValidConvention(string) bool` + `func ValidConventions() []string` to the canonical package (preferred: `internal/config` next to `validGitConventionNames`, reusing the map; or `pkg/models` next to `GitConventionConfig`). `development_mode` needs NO new predicate (`models.ValidDevelopmentModes()` exists).
- Test (RED first): `IsValidConvention` returns true for each of the 5 canonical values, false for a bogus value; `ValidConventions()` returns the 5 values.
- Smallest possible change; reuses the existing canonical map/oneof tag — no new enum.

### M2 — Web project-config validator + read/write seams (Priority: High; REQ-WC3-001/002/004/005)

- `internal/web/app.go`: add 2 injectable seams to `app` — `readProjectConfig func(projectRoot string) (devMode, convention string, err error)` and `writeProjectConfig func(projectRoot, devMode, convention string) error` — defaulting to real functions over the config manager.
- New `internal/web/projectconfig.go` (or add to handlers.go): implement the real read (`config.NewConfigManager().LoadRaw` → return `cfg.Quality.DevelopmentMode`/`cfg.GitConvention.Convention`) and the real write (LoadRaw → mutate only non-empty → `SetSection("quality")`/`SetSection("git_convention")` → `Save`). Import `internal/config`. Add an MX:WARN documenting this as the second persistence boundary (config-manager only, no direct marshal).
- New validator `validateProjectConfig(devMode, convention string) map[string]string` in validate.go: empty allowed; else `development_mode` via `models.DevelopmentMode(devMode).IsValid()` (or `inList` over `ValidDevelopmentModes()`), `git_convention` via `config.IsValidConvention`.
- Tests (RED first): validator returns error key for each bogus value, empty map for each canonical + empty; read seam returns the persisted values; write seam updates quality.yaml/git-convention.yaml round-trip (temp project dir) and leaves a non-targeted section field (e.g. quality.test_coverage_target) unchanged.

### M3 — Web handler wiring + widget select-ification (Priority: High; REQ-WC3-003/004/005)

- `internal/web/handlers.go`: add `DevelopmentModes []string`, `Conventions []string`, `CurDevelopmentMode string`, `CurConvention string` to `pageView`; populate in `newPageView()` (options from M1/models, current values from the new read seam). `handleIndex` calls the read seam; on error → existing `renderError` path.
- `bindForm` (or a sibling bind): read `development_mode` + `git_convention` from the POST form (NOT into `ProfilePreferences`). `handleSave` runs `validatePrefs` AND `validateProjectConfig`, merges FieldErrors; on any error → 400 + re-render; on success → existing profile persist THEN the new project-config write seam.
- `page.html.tmpl`: add a `<fieldset><legend>Project</legend>` with 2 `optSelect` instances (`development_mode` empty "(project default)", `git_convention` empty "(project default)"), bound to `.CurDevelopmentMode`/`.CurConvention` + `.DevelopmentModes`/`.Conventions` + `.FieldErrors`.
- Tests: render test asserts both fields are `<select>` inside the Project fieldset (no `type="text"` for them) with canonical options; handler round-trip bogus→400+field error+no write, valid→200+persisted; empty value→existing value preserved.

### M4 — TUI development_mode + git_convention parity (Priority: Medium; REQ-WC3-006)

- `internal/cli/profile_setup_translations.go`: add label fields `DevelopmentModeTitle`/`DevelopmentModeDesc`/`DevelopmentModeDDD`/`DevelopmentModeTDD` + `GitConventionTitle`/`GitConventionDesc` (+ option labels or raw values + a translated "(project default)") for en/ko/ja/zh, following the `ModelPolicyTitle` pattern.
- `internal/cli/profile_setup.go`: add 2 `huh.NewSelect[string]()` (development_mode, git_convention) in a wizard group (init from the CURRENT project config — read via the config manager, NOT from `existingPrefs`, since they are not profile fields). On save (after the existing `WritePreferences`/`SyncToProjectConfig` at :463/:474), call the project-config write path (the M2 real write function, or a shared helper) to persist the 2 selected non-empty values to quality.yaml + git-convention.yaml.
- Tests: extend `TestGetProfileText_AllLanguages` to assert the new labels non-empty for all 4 locales; a construction/grep guard confirms development_mode + git_convention selects exist; a save test asserts the 2 values land in quality.yaml + git-convention.yaml and NOT in preferences.yaml.

### M5 — Invariant regression + scope verification + closure (Priority: Low; REQ-WC3-007/008)

- Confirm the project-config write touches ONLY quality + git_convention sections (assert workflow/harness/git-strategy/llm files unchanged after a save).
- Run the full 001/002 web suite + integration_test.go to confirm no invariant regression and the DO_NOT_TOUCH sentinels stay green.
- C-HRA-008 grep guard (no AskUserQuestion in new internal/web or internal/cli code).
- Final closure batch (§E).

## §G Anti-Patterns to avoid

- AP-1: Adding `development_mode`/`convention` to `ProfilePreferences` "so they sync like the others" → corrupts the profile schema; they are project-config (B2/B5). Use a separate validator + separate persistence helper.
- AP-2: Authoring a fresh enum for `git_convention` in `internal/web` → mirror the pkg/models `oneof` SSOT with an MX:NOTE, or (preferred) add ONE exported `IsValidConvention` and reuse it.
- AP-3: Direct `yaml.Marshal`/`os.WriteFile` of quality.yaml/git-convention.yaml from `internal/web` → forbidden; use config-manager `SetSection`/`Save` (REQ-WC3-008).
- AP-4: Exposing `llm.mode`/`llm.default_model` "while in here" → narrowed out (backend-switch toggle / legacy enum-less string).
- AP-5: Touching the nested git-convention/llm sub-fields, or the full quality/workflow/harness/git-strategy sections → wrong SPEC (S2b).
- AP-6: Clobbering a persisted non-empty value with an empty submission → empty = "keep existing" (B4).
- AP-7: Running `make build` / chasing an embedded-mirror parity failure → not applicable; web assets are not template-deploy assets (B7).
- AP-8: Modifying the integration_test.go DO_NOT_TOUCH sentinels → they MUST stay green; S2a writes none of those sections.

## §H Cross-References

- `spec.md` §2 (REQs), §3 (AC index), §4 (Exclusions E.1-E.5), §5 (verified paths)
- `acceptance.md` — full Given-When-Then AC enumeration + edge cases + closure gate
- `.moai/specs/SPEC-WEB-CONSOLE-001/` + `.moai/specs/SPEC-WEB-CONSOLE-002/` — siblings (invariants + S1 patterns inherited)
- `.claude/rules/moai/languages/go.md` — Go conventions, ≥85% coverage
- `internal/web/CLAUDE.md` / `internal/cli/CLAUDE.md` / `internal/config/CLAUDE.md` — module conventions (subagent boundary, import direction, SetSection/Save persistence)
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — frontmatter 12-field schema + status transition ownership
