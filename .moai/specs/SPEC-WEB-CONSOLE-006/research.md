# Research — SPEC-WEB-CONSOLE-006 (HTMX + Templ migration)

> Codebase + external-tech analysis grounding the migration. Source: the orchestrator's grounding brief (4 parallel Explore reports: settings.json schema / config 29-section modeling map / console as-is seam / SPEC cohort roadmap) + direct reads of `internal/web/*.go`, the `*_test.go` suite, `assets/page.html.tmpl`, `go.mod`, and the `Makefile`.

## §1 As-is rendering architecture (the migration boundary)

### 1.1 Render pipeline (current)

```
GET /  → handlers.go:handleIndex
          → a.readPreferences(selected)           (seam → profile.ReadPreferences)
          → a.readProjectConfig(projectRoot)      (seam → config-manager read)
          → newPageView(prefs, selected)          (assemble view-model + canonical option lists)
          → a.render(w, 200, view)
              → a.tmpl.Execute(&buf, view)         (html/template, parsed once at startup)
              → w.Write(buf)

POST /save → handlers.go:handleSave
          → bindForm(r)                            (form → ProfilePreferences)
          → validatePrefs(prefs) + validateProjectConfig(devMode, convention)   (validate.go)
          → if errs: render(400, view+FieldErrors)  (atomic reject, EC-2, FULL PAGE)
          → else: writePreferences + syncToProject + writeProjectConfig (seams)
          → render(200, view + "Settings saved" banner)  (FULL PAGE)
```

The renderer is `internal/web/assets/page.html.tmpl` (339 lines), parsed via `assets.go:pageTemplate()` (`html/template` + a `dict` FuncMap because `html/template` has no map constructor), embedded via `assets.go:19` `//go:embed`. Three inline `{{define}}` helpers: `langSelect` (~274–288), `optSelect` (~290–304), `icon` (12 inline SVGs, ~311–339).

### 1.2 What is `html/template`-specific (and must be re-expressed in Templ)

- The `dict` map-constructor FuncMap (Templ uses typed args → `dict` retired).
- `{{template "optSelect" dict "Name" … "Options" .DevelopmentModes}}` directives (→ typed Templ component calls).
- `{{range .Options}}…{{end}}` / `{{if …}}…{{end}}` / `{{with index .FieldErrors .Name}}…{{end}}` (→ Templ `for`/`if`).
- `{{.BindAddr}}` / `{{.CurDevelopmentMode}}` interpolation (→ Templ `{ view.BindAddr }`).
- `a.tmpl.Execute(buf, view)` (→ `page(view).Render(ctx, buf)`).

### 1.3 What is renderer-agnostic (PRESERVE byte-for-byte)

- `validate.go` (validatePrefs + validateProjectConfig + all canonical lists/predicates) — the validation SSOT.
- `app.go` (`app` struct seams, `routes()`, `hostCheckMiddleware`, `isLoopbackHost`).
- `handlers.go` handler logic (handleIndex/handleSave/bindForm) — only the render *call* changes.
- `pageView` view-model struct — becomes the typed Templ root input.
- `console.css` / `app.js` / `i18n.js` / fonts — static assets, untouched.

## §2 Characterization gate — verified test classification

The `internal/web` suite is 11 files / ~70 functions. Direct read confirms the A/B split:

### Class B (server-contract SAFE — stay green UNMODIFIED, ~55 tests)

Assert validation / persistence / HTTP / host-check / scope / section-isolation — NOT markup:
- `handlers_test.go` server subset: TestSaveValidRoundTrip, TestSaveInvalidPermissionModeRejected, TestSaveInvalidThemeRejected, TestSaveSyncFailureSurfacesReadableError, TestSaveScopeBoundary, TestSaveCustomSegmentsRoundTrip, TestSaveNonCustomPresetLeavesSegmentsNil, TestHostCheck{Rejects,Allows,DoesNotGate}*, TestIndexNeutralDefaultsForZeroValueProfile, TestIndexReadErrorRendersInlineError, TestIndexProfileQueryParamSelectsProfile.
- `validate_test.go` (10), `projectconfig_test.go` (5), `projectconfig_scope_test.go` (1) — validation/read-write/isolation.
- `projectconfig_handler_test.go` server subset: TestProjectOptionListsMatchCanonical, TestSaveRejectsBogus{DevelopmentMode,Convention}, TestSaveValidProjectConfigPersists, TestSaveEmptyProjectConfigPasses, TestSaveEC2AtomicReject, TestSaveProjectWriteFailureSurfacesError, TestProjectReadSeamFailureRendersInlineError.
- `integration_test.go` (2): golden-path read/write/invalid/host/scope + empty-state.
- `i18n_test.go` server-contract subset: TestInterfaceLanguageDoesNotAlterPOST (MUST-PASS), TestServerContractPreserved, TestLangpickNotFormField.
- `server_test.go` (8), `coverage_test.go` (9): server init, loopback bind, browser-open non-fatal.

### Class A (markup AT-RISK — careful porting, ~15 tests)

Assert rendered HTML / CSS classes / `data-*` / `name=` / `id=` / SVG — Templ output may differ:
- `restyle_test.go`: TestComponentChromePresent, TestAppbarRendered, TestLoopbackIndicatorShowsRealBindAddr, TestNoNonCanonicalOptions, TestNameAttributesPreserved, TestProfileSwitchNameAttrPreserved, TestBannerKindMapping, TestDarkModeAndThemeToggle, TestInlineSVGIconsNoCDN, TestAccessibilityCues (+ TestConsoleCSSEmbedded / TestPretendardFontSubsetEmbedded / TestFontServedFromStatic = CSS/embed, trivially green).
- `i18n_test.go` markup subset: TestDataI18nWiring, TestDataI18nKeysSubsetOfDictionary, TestNoReviewKeys, TestLangpickRendered, TestLangAttrUpdatesOnSwitch (+ TestI18nDictionaryEmbedded = embed).
- `handlers_test.go` markup subset: TestStaticAssetsServedFromEmbed, TestIndexRendersPopulatedForm, TestIndexProfileSelectionShownForMultipleProfiles.
- `projectconfig_handler_test.go` markup subset: TestProjectFieldsetRendersSelects, TestProjectSelectsPreselectCurrentValues.

The porting strategy preserves exact class names + `data-i18n` + `name=`/`id=` + `{{range}}`→`for` semantics so Class A stays green with minimal/no change. Separately, the **§2.1.1 twelve source-coupled (Class C) tests** that `readEmbeddedAsset(t, "page.html.tmpl")` / call `pageTemplate()` have their mechanism retargeted to the Templ source / rendered HTTP body (each recorded in the acceptance.md §D.3 ledger per AC-WC6-019); the pure symbol-existence `TestPageTemplateParses` is the only one retired (E.5.8 carve-out).

## §3 HTMX + Templ technical validation

### 3.1 Templ (`github.com/a-h/templ`) — verified facts

- `templ generate` is a **pure-Go codegen tool** (`go run github.com/a-h/templ/cmd/templ generate`). No Node, npm, Vite, or JS bundler. Reads `*.templ`, emits `*_templ.go`.
- The generated Go is compiled like any source (NOT a runtime asset → no `go:embed` for it). Only static CSS/JS/font assets are embedded.
- **Contextual auto-escaping** replaces `html/template`'s escaping: `{ expr }` HTML-escaped by default; `templ.Raw(...)` for known-safe markup only. Security parity preserved.
- Components are typed Go functions: `templ optSelect(args optSelectArgs) { … }` → `optSelect(args).Render(ctx, w)`. A renamed field is a compile error (vs `html/template`'s render-time `<no value>`).
- `go.mod` has NO `a-h/templ` today (verified) → this is a genuine new dependency (REQ-WC6-009). Go version is 1.26.4 (Templ supports current Go).

### 3.2 HTMX — verified facts

- `htmx.min.js` (~14.5 KB, v2.x) self-hosted via `go:embed`, served at `/static/htmx.min.js` by the existing FileServer (no new route). Zero CDN → offline preserved (REQ-WC6-006).
- HTMX's model: `hx-get`/`hx-post` + server returns an HTML partial swapped into `hx-target`. This maps cleanly onto Templ (a partial = a standalone-rendered Templ component) — but partial-swap is 007's payoff, NOT 006's (the full-page `/save` Class B tests forbid changing the response shape now).
- Progressive enhancement: `hx-*` attributes degrade gracefully — a non-JS/HTMX-disabled client gets the existing full-page POST round-trip. This is what lets 006 add HTMX without breaking the full-page Class B tests.

### 3.3 The HTMX-depth decision (resolved)

006 lands the Templ render migration + the HTMX *foundation* (embed + link + graceful-degrade `hx-*`) while keeping POST /save observably full-page. Section-scoped partial-swap is 007. Vanilla client-only behaviors (theme/i18n/segment-visibility, all localStorage) stay vanilla — they do NOT fit HTMX's server-driven model and are NOT forced into it. (spec.md §1.5; design.md §E.)

## §4 Build / CI ground-truth

- `go.mod`: `go 1.26.4`, no `a-h/templ`.
- `Makefile`: `build` runs `go run ./internal/template/scripts/gen-catalog-hashes.go` then `go build`; `test` runs `go test -race -coverprofile ./...`; `generate` runs `go generate ./...`; `ci-local` runs `scripts/ci-mirror/run.sh`. → wire `templ generate` into `generate` + ahead of `build`/`test` + the CI mirror (REQ-WC6-010/011).
- The repo already has committed generated artifacts (catalog hashes via `gen-catalog-hashes`) and embeds templates via `go:embed all:templates` (the `embedded.go` retirement noted in project memory). So committing `*_templ.go` (D3) is consistent with the repo's existing generated-source posture.

## §5 Cohort roadmap context

- v3 cohort (001–005) is fully `completed`. 006 opens v4.
- 006 (S2-migrate): HTMX+Templ rendering migration, port 5 fieldsets 1:1, behavior-preserving. THIS SPEC.
- 007 (S2b): nested config-section editing (quality/workflow/git-strategy/harness + nested git-convention/llm), section-scoped partial-swap. Enabled by 006's component tree + HTMX foundation. DEFERRED.
- The grounding brief's user-confirmed scope re-definition: "006 = HTMX+Templ port-first migration ONLY; section expansion deferred to 007 (S2b)." 006 implements zero S2b field.

## §6 Risks surfaced by research

| Risk | Evidence | Mitigation (plan.md) |
|------|----------|----------------------|
| Templ markup drift breaks Class A | `html/template` whitespace/attr-order vs Templ codegen | §B parity contract; port helpers first; per-test justification for any relaxation |
| New build dep + codegen step | `go.mod` has no templ; Makefile/CI need wiring | M1 wires generate→build→test + CI mirror + golangci-lint exclude |
| Fresh-clone `go build` regression | repo currently builds with no external codegen tool | D3: commit `*_templ.go` + CI drift-guard |
| `/save` response-shape change → Class B breakage | golden-path/save/EC-2 assert full-page markers | REQ-WC6-014 pins full page; partial-swap deferred to 007 |
| Scope creep into 007 | the cohort's standing temptation to add a section editor | E.1 HARD fence; 5 fieldsets 1:1, zero new field |

## §7 References

- spec.md §1.3 (migration boundary file:line table), §2 (characterization gate), §3 (REQs), §4 (exclusions).
- design.md (component tree + markup-parity contract + build flow + 007 seeding).
- `internal/web/{assets.go,handlers.go,app.go,validate.go}` + `assets/page.html.tmpl` (verified reads).
- `internal/web/{restyle,i18n,handlers,projectconfig_handler,integration,validate,projectconfig,projectconfig_scope,server,coverage}_test.go` (verified A/B classification).
- `github.com/a-h/templ` (Context7 / official docs — pure-Go codegen, contextual auto-escaping).
- `.moai/specs/SPEC-WEB-CONSOLE-00{1..5}` (cohort invariants REQ-WC-*/REQ-WC2-*/REQ-WC3-*/REQ-WC4-*/REQ-WC5-* ported as REQ-WC6-004…008).
