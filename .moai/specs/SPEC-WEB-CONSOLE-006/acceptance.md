# Acceptance Criteria — SPEC-WEB-CONSOLE-006 (HTMX + Templ migration, port-first / DDD)

Each AC is independently verifiable. The closure gate (AC-WC6-013) is:

```bash
templ generate ./internal/web/... \
  && go test ./internal/web/... ./internal/config/... ./internal/cli/... \
  && go build ./... \
  && GOOS=windows GOARCH=amd64 go build ./...   # all exit 0
```

The characterization gate is the spine: **Class B server-contract tests stay green WITHOUT modification; Class A markup tests stay green with at most §2.2-justified parity adjustments.** Because the migration is behavior-preserving, the AC strategy leans on (1) regression assertions that the existing suite stays green unmodified, (2) structural assertions on the Templ-rendered markup contract, (3) build-wiring assertions for the new `templ generate` dependency, and (4) offline assertions.

## §D AC Matrix (GEARS ↔ AC ↔ verification)

| AC | REQ | Assertion (one-line) | Verification |
|----|-----|----------------------|--------------|
| AC-WC6-001a | REQ-WC6-001 | Every Class B server-contract test (spec.md §2.1) passes WITHOUT test-file modification | `git diff --stat` on the §2.1 files shows zero changes; `go test` green |
| AC-WC6-001b | REQ-WC6-001 | A Class B assertion needing modification triggers a halt + blocker report (not a silent test edit) | code-review of the diff + progress.md blocker-report discipline |
| AC-WC6-002a | REQ-WC6-002 | Every Class A markup test (spec.md §2.2) passes against the Templ render | `go test ./internal/web/...` green |
| AC-WC6-002b | REQ-WC6-002 | Each Class A assertion relaxed to a semantic equivalent has a recorded per-test justification | acceptance.md + progress.md justification entries; no Class A test deleted |
| AC-WC6-003 | REQ-WC6-003 | The Templ render reproduces the 5 fieldsets field-for-field (every `name=`, `id=`, option set, order byte-identical); zero field/section added/removed | `TestNameAttributesPreserved` + `TestProjectFieldsetRendersSelects` green; field-count diff = 0 |
| AC-WC6-004 | REQ-WC6-004 | `validate.go` byte-unchanged; validation via existing predicates/lists; no parallel rule set; no client-side validation SSOT | `git diff --exit-code internal/web/validate.go`; `validate_test.go` green |
| AC-WC6-005 | REQ-WC6-005 | Persistence only via `writePreferences`/`syncToProject`/`writeProjectConfig` seams; no direct YAML marshal in web layer; `app` seams unchanged | `TestSaveValidRoundTrip` + `TestSaveScopeBoundary` green; grep no `yaml.Marshal`/`os.WriteFile` in render path |
| AC-WC6-006 | REQ-WC6-006 | All assets offline via `go:embed` incl. `htmx.min.js`; zero external `https://` font/style/script URL in served assets | `TestConsoleCSSEmbedded` + new htmx-embed test green; `grep -rn 'https://' served-assets` = 0 |
| AC-WC6-007 | REQ-WC6-007 | Loopback bind + `hostCheckMiddleware` byte-unchanged; no auth/token/session/CSRF; no non-loopback bind | `git diff` host-check unchanged; `TestHostCheck*` (3) green |
| AC-WC6-008 | REQ-WC6-008 | Section isolation intact (DO_NOT_TOUCH sentinels) + profile-vs-project scope preserved | `TestGoldenPath_ReadWriteRoundTrip` scope block + `TestWriteProjectConfigSectionIsolation` green |
| AC-WC6-009 | REQ-WC6-009 | `github.com/a-h/templ` in `go.mod` (pinned); `templ generate ./internal/web/...` runs before build/test | `grep a-h/templ go.mod`; `templ generate` exit 0 |
| AC-WC6-010 | REQ-WC6-010 | `//go:generate ... templ generate` directive present; Makefile `generate`/`build`/`test` wire `templ generate`; idempotent | `grep go:generate internal/web`; `make generate` exit 0; re-run is no-op |
| AC-WC6-011 | REQ-WC6-011 | CI mirror runs `templ generate` before build/test; generated `*_templ.go` coexists with `go:embed` (page.html.tmpl dropped, htmx.min.js added) | `scripts/ci-mirror/run.sh` contains `templ generate`; `assets.go` embed updated |
| AC-WC6-012 | REQ-WC6-012 | Host + Windows cross-build exit 0 after `templ generate`; generated Go platform-neutral | `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` exit 0 |
| AC-WC6-013 | REQ-WC6-013 | Closure gate: `templ generate` + `go test ./internal/web/... ./internal/config/... ./internal/cli/...` exit 0 | the closure command exits 0 |
| AC-WC6-014 | REQ-WC6-014 | POST /save returns the same full-page response shape (Templ-rendered); no fragment/partial-swap response | `TestSaveValidRoundTrip` (`"Settings saved"` banner) + `TestGoldenPath_ReadWriteRoundTrip` + EC-2 atomic-reject tests green unmodified |
| AC-WC6-015 | REQ-WC6-015 | `htmx.min.js` embedded + served + linked; progressive-enhancement `hx-*` only; no partial-swap/`hx-target`/section-nav/`/save` fragment | htmx-served test green; grep no `hx-target`/`hx-swap` section-nav; `/save` response-shape unchanged |
| AC-WC6-016 | REQ-WC6-016 | Theme toggle + interface langpick + segment-visibility stay vanilla JS (localStorage); not routed through HTMX; langpick non-submitting | `TestDarkModeAndThemeToggle` + `TestInterfaceLanguageDoesNotAlterPOST` + `TestLangpickNotFormField` green |
| AC-WC6-017 | REQ-WC6-017 | Templ contextual auto-escaping for all dynamic values; `templ.Raw` only on the inline SVG icon set; no user value as raw HTML | code-review: `templ.Raw` call sites = icon markup only; no XSS surface |
| AC-WC6-018 | REQ-WC6-018 | Render-into-buffer-first; render failure → readable inline 500 (not half-written 200) | `TestIndexReadErrorRendersInlineError` + `TestProjectReadSeamFailureRendersInlineError` green |
| AC-WC6-019 | REQ-WC6-001 | Each of the spec.md §2.1.1 twelve Class C source-coupled tests has its mechanism retargeted to the rendered HTTP body / Templ source and is recorded in the §D.3 ledger; NO Class C test is deleted EXCEPT the pure symbol-existence `TestPageTemplateParses` (§2.1.1 #1), which MAY be retired per the E.5.8 carve-out with a §D.3 justification | §D.3 ledger has ≥11 retargeted-mechanism entries (the eleven non-symbol Class C tests) + at most 1 retirement entry; `grep -c 'readEmbeddedAsset(t, "page.html.tmpl")\|pageTemplate()' internal/web/*_test.go` = 0 after migration (no test still reads the deleted source / retired symbol) |

## §D.1 Given-When-Then scenarios

### Scenario 1 — Class B server-contract suite stays green unmodified (AC-WC6-001a, MUST-PASS)

- **Given** the `internal/web` test suite is green before the migration, and the Class B files listed in spec.md §2.1 define the observable server behavior,
- **When** the HTMX+Templ migration is complete (`templ generate` + the render swap + the build wiring),
- **Then** every Class B test passes WITHOUT any modification to the Class B test files (`git diff --stat` on those files is empty), and `go test ./internal/web/... ./internal/config/... ./internal/cli/...` exits 0.

### Scenario 2 — POST /save round-trip is behavior-identical (AC-WC6-014, MUST-PASS)

- **Given** a real Console server (`TestGoldenPath_ReadWriteRoundTrip`) with a seeded profile + project config,
- **When** a valid `POST /save` is submitted (after the migration),
- **Then** the response is HTTP 200 with the same full-page markers the pre-migration server returned (`value="Goos"`, `method="POST"`, `"Settings saved"` banner), the change is persisted via the real seams, the read-back reflects it, an invalid submit is atomically rejected (400, state unchanged), a foreign-Host POST is 403, and the out-of-scope sections (workflow/harness/git-strategy) are byte-for-byte unchanged — ALL without modifying the test.

### Scenario 3 — Class A markup parity preserved (AC-WC6-002a/003)

- **Given** the Templ components port the `langSelect`/`optSelect`/`icon` helpers and the 5 fieldsets,
- **When** `GET /` renders the page via the Templ root component,
- **Then** the rendered body carries the exact class names (`section`, `field__title`, `select`, `select--lang`, `btn btn--primary`, `appbar`, `loopback`, `langpick`), the `data-i18n` attributes (≥25, code-chips excluded), every `name=` POST attribute (15 fields + 15 segments + `__profile` + `__profile_select`), every `id=` (`uiLangSelect`, `themeToggle`, field ids), the `{ view.BindAddr }` loopback indicator, and the canonical option sets — so `TestComponentChromePresent`, `TestAppbarRendered`, `TestNameAttributesPreserved`, `TestDataI18nWiring`, `TestProjectFieldsetRendersSelects`, `TestLangpickRendered` all pass.

### Scenario 4 — Templ build wiring is reproducible (AC-WC6-009/010/011/012)

- **Given** a fresh checkout with `github.com/a-h/templ` in `go.mod` and the committed `*_templ.go`,
- **When** CI (or `scripts/ci-mirror/run.sh`) runs `templ generate ./internal/web/...` then `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` + `go test ./...`,
- **Then** `templ generate` succeeds (idempotent — no diff on regenerate), both builds exit 0, the tests exit 0, golangci-lint does not flag the generated `*_templ.go`, and the `assets.go` `go:embed` directive enumerates `htmx.min.js` but not `page.html.tmpl`.

### Scenario 5 — Offline / zero-network preserved with htmx.min.js (AC-WC6-006/015)

- **Given** `htmx.min.js` is self-hosted under `internal/web/assets/`, embedded via `go:embed`, and linked from the page,
- **When** the served CSS/HTML/JS is scanned for external resource URLs,
- **Then** there is zero external `https://` font/style/script URL (no Google-Fonts `@import`, no unpkg, no jsdelivr, no cdnjs, no HTMX CDN), `GET /static/htmx.min.js` returns 200 from the embed, and the page links `/static/htmx.min.js` before `app.js`.

### Scenario 6 — HTMX foundation degrades gracefully, no partial-swap (AC-WC6-014/015, E.3)

- **Given** the form carries progressive-enhancement `hx-*` attributes,
- **When** a normal (non-HTMX or HTMX-disabled) `POST /save` is submitted,
- **Then** the server returns the full-page response (the existing behavior), there is no `hx-target`/`hx-swap` section-scoped navigation, no fragment `/save` response, and no section-scoped save endpoint — the partial-swap is provably absent (deferred to 007).

### Scenario 7 — Client-only behaviors stay vanilla, interface language non-submitting (AC-WC6-016)

- **Given** the theme toggle and interface langpick are client-only localStorage preferences,
- **When** a form is submitted while the interface language is `ja`,
- **Then** the bound `ProfilePreferences` are byte-identical to an `en`-interface submit (`TestInterfaceLanguageDoesNotAlterPOST`), the langpick carries no `name=` and sits outside `<form>` (`TestLangpickNotFormField`), and the theme/i18n/segment-visibility JS is unchanged (`TestDarkModeAndThemeToggle`, `TestLangAttrUpdatesOnSwitch`).

## §D.2 Edge cases

- **EC-1 (empty project-config preserve)** — POST with empty `development_mode`/`git_convention` keeps existing section values (handler passes through; the write seam is the no-clobber guard). `TestSaveEmptyProjectConfigPasses` green unmodified.
- **EC-2 (atomic reject, full-page)** — one bogus + one valid field → 400, only the bad field's error, NEITHER value persisted, AND a full-page re-render (not a fragment). `TestSaveEC2AtomicReject` green; the full-page re-render is the AC-WC6-014 invariant.
- **EC-3 (render error → readable inline 500)** — a read-seam or Templ-render error surfaces the cause inline (never blank, never panic, never half-written 200). `TestIndexReadErrorRendersInlineError` + `TestProjectReadSeamFailureRendersInlineError` green.
- **EC-4 (custom-preset segment round-trip)** — `statusline_preset=custom` binds all 15 `segment_<key>` (checked → true, unchecked → false, none dropped). `TestSaveCustomSegmentsRoundTrip` green unmodified (handler logic untouched).
- **EC-5 (zero-value profile neutral defaults)** — a fresh profile renders `(unset)` defaults via the Templ render, not an error. `TestIndexNeutralDefaultsForZeroValueProfile` + `TestGoldenPath_EmptyStateNeutralDefaults` green.
- **EC-6 (stale generated file drift)** — if a `.templ` source changes but `*_templ.go` is not regenerated, CI's `templ generate` drift-guard fails (the committed generated file differs from a fresh generate). Mitigation D3: CI regenerates + diffs.
- **EC-7 (source-coupled tests vs vanished page.html.tmpl / retired pageTemplate())** — the **spec.md §2.1.1 twelve source-coupled (Class C) tests** (which `readEmbeddedAsset(t, "page.html.tmpl")` OR call `pageTemplate()`) must have their mechanism retargeted to the Templ source / rendered HTTP body — each recorded in the §D.3 ledger (AC-WC6-019). The pure symbol-existence test `TestPageTemplateParses` (§2.1.1 #1) is retired per the E.5.8 carve-out; the other eleven are retargeted, never deleted.

## §D.3 Class A markup-parity justification ledger (run-phase fills)

[HARD] Per REQ-WC6-002 / AC-WC6-002b: the run-phase MUST record, for each Class A test that required ANY assertion relaxation (not exact-markup parity), a justification entry here AND in progress.md §E.2. Template:

| Test | Original assertion | Why exact parity infeasible | Semantic relaxation applied |
|------|--------------------|-----------------------------|-----------------------------|
| _(run-phase fills — e.g. `TestLoopbackIndicatorShowsRealBindAddr`)_ | _(asserts `page.html.tmpl` source contains `{{.BindAddr}}`)_ | _(page.html.tmpl no longer exists post-migration)_ | _(re-express against the rendered body / Templ source — justified)_ |

[HARD] If this ledger is empty at closure, that asserts EVERY Class A test passed via exact-markup parity with zero relaxation — a stronger result. An empty ledger is valid and preferred.

## §D.4 Severity classification

| Severity | ACs |
|----------|-----|
| MUST-PASS (closure-blocking) | AC-WC6-001a, AC-WC6-004, AC-WC6-013, AC-WC6-014 (behavior-preservation core), AC-WC6-006, AC-WC6-012 |
| MUST-PASS (build-wiring) | AC-WC6-009, AC-WC6-010, AC-WC6-011 |
| SHOULD-PASS (markup parity) | AC-WC6-002a/b, AC-WC6-003, AC-WC6-015, AC-WC6-016 |
| SHOULD-PASS (inherited invariants) | AC-WC6-005, AC-WC6-007, AC-WC6-008, AC-WC6-017, AC-WC6-018 |
| OBSERVE (process) | AC-WC6-001b (blocker-report discipline) |

## §D.5 Traceability

Every REQ-WC6-NNN maps to ≥1 AC; every AC maps to ≥1 verification command or named test. The inherited invariants REQ-WC6-004…008 trace to the verbatim 001–005 REQ-WC IDs cited in spec.md §3 (REQ-WC-002/005/007/008/009/012, REQ-WC2-002/007/008, REQ-WC3-005/007, REQ-WC4-001/009, REQ-WC5-008/010). The build-wiring REQ-WC6-009…013 are NEW (no v3 precedent — this is the first build-system change in the cohort).

## §D.6 Definition of Done

- [ ] `templ generate ./internal/web/...` exit 0 (idempotent)
- [ ] `go test ./internal/web/... ./internal/config/... ./internal/cli/...` exit 0
- [ ] `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` exit 0
- [ ] Every Class B test (spec.md §2.1) green WITHOUT test-file modification (`git diff --stat` empty on those files)
- [ ] `validate.go` byte-unchanged (`git diff --exit-code internal/web/validate.go`)
- [ ] `app.go` host-check + seams byte-unchanged; `console.css`/`app.js`/`i18n.js`/fonts preserved
- [ ] `github.com/a-h/templ` in `go.mod` (pinned); `//go:generate` directive present; Makefile + CI wired
- [ ] `assets.go` `go:embed` drops `page.html.tmpl`, adds `htmx.min.js`; `htmx.min.js` served from `/static/`
- [ ] Zero external `https://` font/style/script URL in served assets (grep = 0)
- [ ] POST /save full-page response shape unchanged; no HTMX partial-swap (AC-WC6-014/015)
- [ ] Zero new settings field / config section (AC-WC6-003); zero S2b field implemented (E.1)
- [ ] `internal/web` coverage ≥ baseline; golangci-lint clean (generated files excluded)
- [ ] Class A markup-parity justification ledger (§D.3) filled (or empty = all exact-parity)
- [ ] progress.md §E.2 run-phase evidence: per-AC PASS matrix + D3 decision + HTMX-depth confirmation
