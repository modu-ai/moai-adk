# Acceptance Criteria — SPEC-WEB-CONSOLE-003

> S2a of the `web-console-v3` cohort. **Tier M** — full Given-When-Then enumeration. The closure gate for ALL ACs is `go test ./internal/web/... ./internal/cli/... ./internal/config/...` exit 0. Each AC is independently verifiable. Reads with `spec.md` (§2 REQs, §3 AC index) and `plan.md` (§F milestones).

## §A Canonical values (verified ground-truth)

- `development_mode` canonical: `{ddd, tdd}` — source `pkg/models` `models.ValidDevelopmentModes()` / `DevelopmentMode.IsValid()`. Empty string = "keep project default" (always valid).
- `git_convention` canonical: `{auto, conventional-commits, angular, karma, custom}` — source `pkg/models` `GitConventionConfig.Convention` `oneof` tag / new exported `config.IsValidConvention`. Empty string = "keep project default" (always valid).
- Current project values (at SPEC authoring): `development_mode: tdd` (quality.yaml), `convention: auto` (git-convention.yaml).

## §B Severity classification

| Class | ACs |
|-------|-----|
| MUST-PASS (blocking) | AC-WC3-001a, 001b, 002a, 002b, 003, 004, 005, 006a, 006b, 007, 008, 009 |
| (all S2a ACs are MUST-PASS — there are no nice-to-have ACs) | — |

## §C Traceability matrix (AC ↔ REQ)

| AC | REQ | Verification surface |
|----|-----|----------------------|
| AC-WC3-001a/001b | REQ-WC3-001 | `internal/web` validator + handler round-trip (development_mode) |
| AC-WC3-002a/002b | REQ-WC3-002 + M1 predicate | `internal/web` validator + handler round-trip (git_convention) + `internal/config` IsValidConvention |
| AC-WC3-003 | REQ-WC3-003 | `internal/web` template render test |
| AC-WC3-004 | REQ-WC3-004 | `internal/web` read-seam + handleIndex render test |
| AC-WC3-005 | REQ-WC3-005 + REQ-WC3-007 | `internal/web` write-seam + scope test |
| AC-WC3-006a/006b | REQ-WC3-006 | `internal/cli` translations parity + wizard construction + save test |
| AC-WC3-007 | REQ-WC3-007 | section-isolation test (quality + git_convention only) |
| AC-WC3-008 | REQ-WC3-008 | 001/002 suite + integration_test sentinels |
| AC-WC3-009 | all | closure command exit 0 |

---

## §D Acceptance Criteria (Given-When-Then)

### AC-WC3-001a — Web rejects out-of-list `development_mode` (REQ-WC3-001)

- **Given** the Console running with a temp project config (`development_mode: tdd`)
- **When** a `POST /save` submits `development_mode=xyz` (non-canonical)
- **Then** the response is HTTP 400, the re-rendered `FieldErrors["development_mode"]` is non-empty, and the persisted `quality.yaml` `development_mode` value is unchanged (still `tdd`)
- **Verify**: `validateProjectConfig` test returns a `development_mode` key for a bogus value; handler round-trip asserts 400 + field error + `quality.yaml` byte-identical before/after.

### AC-WC3-001b — Web persists valid `development_mode` (REQ-WC3-001/005)

- **Given** the Console running with a temp project config (`development_mode: tdd`)
- **When** a `POST /save` submits `development_mode=ddd` (canonical)
- **Then** the response is HTTP 200 ("Settings saved") and the persisted `quality.yaml` `development_mode` is `ddd`
- **Verify**: round-trip handler test (write then re-read via config manager) using a temp project dir; assert `cfg.Quality.DevelopmentMode == "ddd"`. The validator returns the empty map for `ddd`, `tdd`, and empty string.

### AC-WC3-002a — Web rejects out-of-list `git_convention` (REQ-WC3-002)

- **Given** the Console running with a temp project config (`convention: auto`)
- **When** a `POST /save` submits `git_convention=gitflow` (non-canonical)
- **Then** the response is HTTP 400, `FieldErrors["git_convention"]` is non-empty, and the persisted `git-convention.yaml` `convention` value is unchanged (still `auto`)
- **Verify**: `validateProjectConfig` returns a `git_convention` key for a bogus value; handler round-trip asserts 400 + field error + `git-convention.yaml` unchanged. The validator delegates to `config.IsValidConvention` (not a freshly authored web-local enum).

### AC-WC3-002b — Web persists valid `git_convention` (REQ-WC3-002/005)

- **Given** the Console running with a temp project config (`convention: auto`)
- **When** a `POST /save` submits `git_convention=angular` (canonical)
- **Then** the response is HTTP 200 and the persisted `git-convention.yaml` `convention` is `angular`
- **Verify**: round-trip handler test asserts `cfg.GitConvention.Convention == "angular"`. The validator returns the empty map for each of `auto`/`conventional-commits`/`angular`/`karma`/`custom`/empty.

### AC-WC3-003 — Web widgets are `<select>` in a "Project" fieldset (REQ-WC3-003)

- **Given** the rendered Console page (`GET /`)
- **When** the HTML for `development_mode` and `git_convention` is inspected
- **Then** each is a `<select>` element (no `<input type="text">` for these two), each contains exactly its canonical option set plus the empty "(project default)" option, and both sit inside a `<fieldset>` whose legend is "Project"
- **Verify**: a render test (template executed against a view-model) asserts both field names appear inside `<select` blocks with the expected `<option value="...">` entries (2 options + empty for development_mode; 5 options + empty for git_convention); a negative assertion confirms no `type="text"` for `name="development_mode"|"git_convention"`; an assertion confirms a `Project` legend precedes both fields.

### AC-WC3-004 — `GET /` pre-selects current project-config values (REQ-WC3-004)

- **Given** a temp project config with `development_mode: ddd` and `convention: karma`
- **When** the Console serves `GET /`
- **Then** the `development_mode` select has `ddd` marked `selected` and the `git_convention` select has `karma` marked `selected` (values read from quality.yaml/git-convention.yaml via the config manager, NOT from the profile store)
- **And When** the project-config read seam is injected to fail
- **Then** the response is a readable inline error (REQ-WC-010 invariant), never a blank page or panic
- **Verify**: a handler test injects a read seam returning `("ddd","karma",nil)` and asserts the rendered HTML marks those options selected; a second test injects a read seam returning an error and asserts a non-blank error page (status ≥ 400 or the existing renderError path) with no panic.

### AC-WC3-005 — Project-config write is a new injectable seam; empty keeps existing; no direct marshal (REQ-WC3-005/007)

- **Given** the `app` struct
- **When** the source is inspected
- **Then** there exist injectable `readProjectConfig` and `writeProjectConfig` seams on `app` (mirroring the existing `syncToProject` seam), and the real `writeProjectConfig` persists via `config.NewConfigManager()`/`LoadRaw`/`SetSection`/`Save` — with NO `yaml.Marshal`/`os.WriteFile` of a config section anywhere in `internal/web`
- **And Given** a temp project config (`development_mode: tdd`, `convention: auto`)
- **When** a `POST /save` submits `development_mode=ddd` and `git_convention=` (empty)
- **Then** `quality.yaml` becomes `ddd` AND `git-convention.yaml` `convention` stays `auto` (empty submission leaves the existing value)
- **Verify**: a grep-guard asserts no `yaml.Marshal(` / `os.WriteFile(` targeting a config section in `internal/web/*.go` (excluding `_test.go`); a round-trip test confirms the empty-keeps-existing semantics for each of the two fields independently.

### AC-WC3-006a — TUI exposes both selects with 4-locale labels (REQ-WC3-006)

- **Given** the terminal profile wizard form
- **When** the wizard groups are constructed
- **Then** the form contains a `development_mode` `huh.Select` (options `(project default)`/`ddd`/`tdd`) and a `git_convention` `huh.Select` (options `(project default)`/`auto`/`conventional-commits`/`angular`/`karma`/`custom`), and `profile_setup_translations.go` defines the corresponding label strings for en/ko/ja/zh
- **Verify**: `TestGetProfileText_AllLanguages` is extended to assert `DevelopmentModeTitle` + `GitConventionTitle` (and their option/desc labels) are non-empty for all four locales; a construction/grep guard confirms a `development_mode`-bound and a `git_convention`-bound select exist in `profile_setup.go`.

### AC-WC3-006b — TUI save persists to project config, not preferences.yaml (REQ-WC3-006)

- **Given** the wizard with `development_mode=ddd` and `git_convention=angular` selected, in a temp project
- **When** the wizard save path runs
- **Then** `quality.yaml` `development_mode` becomes `ddd` AND `git-convention.yaml` `convention` becomes `angular`, AND `preferences.yaml` contains NO `development_mode`/`convention` keys (they are not `ProfilePreferences` fields)
- **Verify**: a save test asserts the two project-config files hold the values and that the written `preferences.yaml` (via `ReadPreferences`) has neither a development_mode nor a convention field (the `ProfilePreferences` struct has no such fields, so this is structurally guaranteed — the test documents it).

### AC-WC3-007 — Write path touches ONLY quality + git_convention sections (REQ-WC3-007)

- **Given** a temp project with workflow.yaml / harness.yaml / git-strategy.yaml / llm.yaml present
- **When** a web `POST /save` (or a TUI save) persists `development_mode=ddd` + `git_convention=angular`
- **Then** ONLY `quality.yaml` and `git-convention.yaml` are modified; `workflow.yaml`, `harness.yaml`, `git-strategy.yaml`, `llm.yaml` are byte-identical before and after (the config-manager round-trip rewrites all sections, but their CONTENT is unchanged)
- **Verify**: a test snapshots the mtimes/contents of workflow/harness/git-strategy/llm yaml files before and after a save, asserting their content is unchanged (the load-mutate-save cycle must not alter non-targeted section fields). NOTE: the config manager's `Save()` rewrites all 5 sections it owns; the assertion is on CONTENT equality, not file-untouched.

### AC-WC3-008 — Invariants preserved (REQ-WC3-008)

- **Given** the full S2a change set
- **When** the existing `SPEC-WEB-CONSOLE-001` + `SPEC-WEB-CONSOLE-002` web test suites + `integration_test.go` run
- **Then** loopback-bind, no-auth, Host-header check, profile-persistence-path tests, and the `integration_test.go` DO_NOT_TOUCH sentinels (workflow.yaml/harness.yaml/git-strategy.yaml NOT written by the web layer's profile path) all still pass — no regression
- **Verify**: `go test ./internal/web/...` green; no test asserting `0.0.0.0` bind, token auth, session store, or a direct-YAML-write profile path is added or weakened; the DO_NOT_TOUCH sentinels remain unmodified and green.

### AC-WC3-009 — Closure gate (all REQs)

- **Given** the complete S2a implementation
- **When** `go test ./internal/web/... ./internal/cli/... ./internal/config/...` runs
- **Then** the suite is green with zero failures
- **Verify**: command exit 0. Supplementary: `go vet` clean, `golangci-lint run` clean on the three packages, `GOOS=windows GOARCH=amd64 go build ./...` exit 0.

---

## §D.1 Edge cases (must be covered by tests)

- **EC-1 — both fields empty**: `POST /save` with `development_mode=` and `git_convention=` both empty → 200, both project-config values unchanged (no clobber).
- **EC-2 — one bogus, one valid**: `development_mode=xyz` (bogus) + `git_convention=angular` (valid) → 400, FieldErrors has `development_mode` only, NEITHER value is persisted (atomic reject — validation failure blocks the whole save, mirroring S1's all-or-nothing handler behavior).
- **EC-3 — convention `custom` value accepted as enum**: `git_convention=custom` is a valid enum value and persists; S2a does NOT additionally require/validate `custom.pattern` (that sub-field is S2b). The persisted `convention: custom` round-trips the existing `custom.*` sub-structure unchanged.
- **EC-4 — case sensitivity**: `development_mode=TDD` (uppercase) is NOT canonical (`ddd`/`tdd` are lowercase) → 400 + field error (validator does not lowercase-normalize; matches the existing `inList`/`oneof` exact-match behavior).
- **EC-5 — absent config file**: if `quality.yaml`/`git-convention.yaml` is absent in the temp project, the read seam returns the default/empty value (LoadRaw behavior) and the write seam creates the section via `Save()` — no panic.
- **EC-6 — whitespace value**: `development_mode=" "` (a space) is non-empty and non-canonical → 400 + field error (not treated as empty).

## §D.2 Closure gate (Definition of Done)

S2a is DONE when ALL of the following hold:

1. All 12 MUST-PASS ACs (AC-WC3-001a..009) verified PASS with their stated verification commands.
2. All 6 edge cases (EC-1..EC-6) covered by tests.
3. `go test ./internal/web/... ./internal/cli/... ./internal/config/...` exit 0.
4. `go vet` + `golangci-lint run` clean on the three packages.
5. `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
6. Package coverage ≥ 85% on changed packages (internal/web, internal/cli, internal/config).
7. C-HRA-008: no `AskUserQuestion`/`mcp__askuser` in new `internal/web` or `internal/cli` code.
8. The `internal/web/integration_test.go` DO_NOT_TOUCH sentinels are unmodified and green.
9. No direct `yaml.Marshal`/`os.WriteFile` of a config section in `internal/web` (grep-guard clean).
10. status frontmatter transitions `draft → in-progress` (M1 commit, manager-develop) per the ownership matrix.
