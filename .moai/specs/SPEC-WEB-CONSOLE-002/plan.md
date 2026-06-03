# Implementation Plan — SPEC-WEB-CONSOLE-002

> S1 of the `web-console-v3-extension` cohort. Tier S. Reads with `spec.md` (ACs inline in §3).

## §A Context

`SPEC-WEB-CONSOLE-001` shipped `moai web` (loopback settings CRUD). S1 hardens web↔TUI parity for three fields that the web layer currently accepts as free text, converts those widgets to dropdowns, adds the missing `model_policy` select to the TUI, and changes the default port from the colliding `8080` candidate to `3041`. Everything reuses existing canonical lists/predicates — no new validation rule-set, no new config section.

Development mode: per `.moai/config/sections/quality.yaml` `development_mode` (TDD by default — write the failing validate/handler/render test first, then the minimal change).

## §B Known Issues / Pre-flight

- **Import direction constraint** — `internal/cli → internal/web` is the only legal direction (web cannot import cli; documented in `internal/web/validate.go:15-19`). The wizard's `model`/`effort_level` option strings are unexported in `internal/cli`. Therefore the web `model`/`effort_level` canonical lists must be **mirrored** (re-declared) in `internal/web`, exactly as the existing statusline/lang lists already are. `model_policy` is the exception: `template.IsValidModelPolicy` / `template.ValidModelPolicies` ARE exported and importable from `internal/web` — wire those in directly (no mirror).
- **`normalizeModel`** lives in `internal/cli` (unexported package-level func). It is NOT importable from `internal/web`. For S1, web validation only needs membership of the 7 canonical aliases; deprecated-ID *migration* on the web path is optional and may be deferred (the wizard already normalizes on its own read path). Decision: mirror the 7-value canonical list in web; do not attempt to reuse `normalizeModel` across the package boundary.
- **`page.html.tmpl` embedding** — confirm the web assets are embedded via the web package's own `go:embed` (not the `internal/template/templates` deploy system) so no `make build` mirror step applies. Pre-flight: `grep -rn "go:embed" internal/web/`.
- **TUI translations parity test** — check whether a `TestProfileSetupTranslations*` parity test exists that enforces all locales define every key; if so, the new `model_policy` keys must be added to all 4 locales or that test fails (this is the AC-WC2-006 guard, working in our favor).

## §C Pre-flight verification batch (run before M1)

```bash
# 1. Confirm port default is currently 8080
grep -n '"port", 8080\|default 8080\|:8080' internal/cli/web.go
# 2. Confirm validatePrefs has no model/effort/model_policy branch
grep -n 'model\|effort\|policy' internal/web/validate.go
# 3. Confirm IsValidModelPolicy is exported + currently unused outside tests
grep -rn 'IsValidModelPolicy' internal/ | grep -v _test.go
# 4. Confirm web assets embed mechanism (not template-deploy)
grep -rn 'go:embed' internal/web/
# 5. Confirm TUI model-settings group lacks model_policy
grep -n 'model_policy\|ModelPolicy\|ModelSettingsTitle' internal/cli/profile_setup.go
# 6. Locate translations struct + a parity test if any
grep -rn 'ModelOverrideTitle\|ModelSettingsTitle' internal/cli/profile_setup_translations.go
go test ./internal/web/... ./internal/cli/... 2>&1 | tail -5   # baseline green
```

## §D Constraints

- [HARD] Reuse canonical lists/predicates; mirror ONLY where the source value is unexported (model, effort_level). Import `template.IsValidModelPolicy` for model_policy. No parallel rule-set.
- [HARD] No new config section for model_policy (profile-only — REQ-WC2-007).
- [HARD] Anti-over-engineering: smallest change for port + 3-field validation + select-ification + TUI model_policy. No new abstractions, no helper packages.
- [HARD] Preserve all SPEC-WEB-CONSOLE-001 invariants (loopback, no-auth, Host-header, persistence path).
- [HARD] Scope-fence: i18n (S3), webfont (S3), 8 UI-missing settings (S2), dead-config removal (S4) are OUT.
- Go ≥ 1.23 conventions (`.claude/rules/moai/languages/go.md`): explicit error wrapping, table-driven tests, ≥85% package coverage on changed packages.

## §E Self-Verification (closure gate)

- [ ] `go test ./internal/web/... ./internal/cli/...` exit 0 (AC-WC2-009)
- [ ] `go vet ./internal/web/... ./internal/cli/...` clean
- [ ] `golangci-lint run ./internal/web/... ./internal/cli/...` clean
- [ ] Port default asserted == "3041" (AC-WC2-001)
- [ ] `validatePrefs` rejects bogus + accepts canonical for all 3 fields (AC-WC2-002a/002b/003/004)
- [ ] Render test: 3 fields are `<select>`, no `type="text"` for them (AC-WC2-005)
- [ ] TUI model_policy select present + 4-locale labels non-empty (AC-WC2-006)
- [ ] model_policy persists profile-only, no config-section write (AC-WC2-007)
- [ ] SPEC-WEB-CONSOLE-001 suite still green — no invariant regression (AC-WC2-008)

## §F Milestones (priority order — no time estimates)

### M1 — Port default 3041 (Priority: High; AC-WC2-001)

- `internal/cli/web.go`: change `cmd.Flags().IntVar(&webPort, "port", 8080, ...)` → `3041`; update the Long help block (`Flags:` line and the two `Examples:` lines that print `:8080`).
- Test: table-driven test asserting the constructed `web` command's `--port` flag `DefValue == "3041"`.
- Smallest possible change; no other file touched in M1.

### M2 — Web validation parity (Priority: High; AC-WC2-002a/002b/003/004)

- `internal/web/validate.go`:
  - Add mirrored canonical lists `modelCanonical` (7 values) and `effortLevelCanonical` (6 values) alongside the existing mirror lists, with an MX:NOTE pointing at `internal/cli/profile_setup.go:303-310` / `:316-322` as SSOT (mirror rationale = unidirectional import constraint, same as the existing statusline mirrors).
  - Import `github.com/modu-ai/moai-adk/internal/template` and call `template.IsValidModelPolicy` for the `model_policy` branch (NO mirror — it's exported).
  - Extend `validatePrefs()` with three branches: `model` (empty allowed, else `inList(modelCanonical,...)`), `effort_level` (empty allowed, else `inList(effortLevelCanonical,...)`), `model_policy` (empty allowed, else `template.IsValidModelPolicy(...)`). Mirror the existing empty-allowed pattern verbatim.
- Tests (write first — RED): `validatePrefs` returns error key for each bogus value and empty map for each canonical value + empty string, across all 3 fields. Handler round-trip: bogus → 400 + field error + state unchanged; canonical → 200 + persisted.

### M3 — Web widget select-ification (Priority: Medium; AC-WC2-005)

- `internal/web/handlers.go`: add `ModelOptions []string`, `EffortLevels []string`, `ModelPolicies []string` to `pageView`; populate in `newPageView()` from the M2 mirror lists + `template.ValidModelPolicies()`.
- `internal/web/assets/page.html.tmpl:62-71`: replace the 3 `<input type="text">` with `<select>` blocks. Reuse the structure of the existing `langSelect` define (empty-default `<option>` + `range` over options + `has-error` class + `field-error` div). Each select keeps `{{if eq . $.Prefs.<Field>}}selected{{end}}`.
- Test: render the template against a view-model; assert each of the 3 field names appears inside a `<select` block with expected `<option value=...>`; negative assert no `type="text"` for those names.

### M4 — TUI model_policy parity (Priority: Medium; AC-WC2-006)

- `internal/cli/profile_setup_translations.go`: add `ModelPolicyTitle` + `ModelPolicyDesc` (+ option labels if the existing model/effort labels use per-option translated strings; otherwise raw `high`/`medium`/`low` values suffice with a translated `(project default)` reusing `ModelDefault`-style key) for en/ko/ja/zh. Follow the exact pattern of the existing `ModelOverrideTitle`/`EffortLevelTitle` entries.
- `internal/cli/profile_setup.go`: add a `model_policy` value var (init from `existingPrefs.ModelPolicy`) and a `huh.NewSelect[string]()` bound to it inside the model-settings group (after effort or after permission_mode), with options `("", high, medium, low)`. Persist the value into the written `ProfilePreferences{ModelPolicy: ...}` in the save path (locate where the form values are assembled into prefs and add the field — verify the wizard's prefs-assembly currently omits ModelPolicy and add it).
- Test: translations parity test sees the new keys for all 4 locales (or add to the existing parity test's key set); a construction/grep guard confirms a model_policy-bound select exists.

### M5 — Sync decision documentation + invariant regression (Priority: Low; AC-WC2-007/008)

- No code for the sync decision beyond confirming `SyncToProjectConfig` scope is untouched (REQ-WC2-007 is "do nothing new"). Add a one-line MX:NOTE near the model_policy bindForm/persist path stating "profile-only by design — see SPEC-WEB-CONSOLE-002 REQ-WC2-007" if a natural anchor exists.
- Run the full `SPEC-WEB-CONSOLE-001` web suite to confirm no invariant regression (AC-WC2-008).
- Final closure batch (§E).

## §G Anti-Patterns to avoid

- AP-1: Declaring a fresh `model`/`effort`/`policy` list in web that drifts from the wizard SSOT → mirror with an MX:NOTE pointing at the SSOT line; import where exported.
- AP-2: Adding a `model_policy:` key to a config section "so it syncs" → forbidden by REQ-WC2-007.
- AP-3: Adding any auth/token/session or `0.0.0.0` bind "while in here" → invariant violation (REQ-WC2-008).
- AP-4: Touching the 8 UI-missing settings, i18n, webfont, or dead-config → wrong SPEC (S2/S3/S4).
- AP-5: Running `make build` / chasing an embedded-mirror parity failure → not applicable; web assets are not template-deploy assets (§B / §D8 of spec.md).

## §H Cross-References

- `spec.md` §2 (REQs), §3 (inline ACs), §4 (Exclusions), §5 (verified paths)
- `.moai/specs/SPEC-WEB-CONSOLE-001/` — sibling (parent feature; invariants inherited)
- `.claude/rules/moai/languages/go.md` — Go conventions, ≥85% coverage
- `internal/web/CLAUDE.md` / `internal/cli/CLAUDE.md` — module conventions (subagent boundary, import direction)
