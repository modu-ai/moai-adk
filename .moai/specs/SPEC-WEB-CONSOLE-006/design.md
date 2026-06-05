# Design — SPEC-WEB-CONSOLE-006 (Templ component architecture + HTMX foundation + build/codegen flow)

> Design-level decisions for the rendering-layer migration. The SPEC body (spec.md) owns WHAT/WHY; this document owns the component architecture, the codegen/build flow, and how the architecture seeds 007 (S2b). No implementation code here — structure only.

## §A Component tree (the Templ render graph)

The single `page.html.tmpl` (339 lines) + 3 `{{define}}` helpers map to a small typed Templ component tree. The view-model (`handlers.go:pageView`) is preserved as the typed input to the root.

```
page(view pageView)                          // root: <html> shell + <head> FOUC script + <body>
├── appbar(view)                             // brand SVG + loopback{ view.BindAddr } + langpick(uiLangSelect) + themeToggle
│   ├── icon("sun") / icon("moon")           // themeToggle icons (templ.Raw on the inline SVG)
│   └── langpick                             // class="select langpick" id="uiLangSelect" aria-label, NO name=, OUTSIDE <form>
├── profileSwitch(view)                      // { if view.ShowProfileSwitch } — __profile_select + (current) marker
├── <form method="POST" action="/save"…>     // hidden __profile + hx-* progressive-enhancement attrs (M4)
│   ├── fieldsetIdentity(view)               // user_name text input
│   ├── fieldsetLanguage(view)               // 4× langSelect(conversation_lang / git_commit_lang / code_comment_lang / doc_lang)
│   ├── fieldsetLaunch(view)                 // optSelect(permission_mode / model_policy / model / effort_level)
│   ├── fieldsetStatusline(view)             // statusline_mode/preset/theme optSelects + { for s := range view.AllSegments } segment_<key> checkboxes
│   └── fieldsetProject(view)                // optSelect(development_mode / git_convention) — project-config (SPEC-003)
└── actions(view)                            // Save settings button (btn btn--primary) + actions__meta
                                             // <script src="/static/i18n.js"> + <script src="/static/htmx.min.js"> + <script src="/static/app.js">

// Reusable leaf components (ported 1:1 from the {{define}} helpers):
langSelect(args langSelectArgs)              // field chrome + select--lang + (unset) + {range Options} + error span
optSelect(args optSelectArgs)                // field chrome + select + .Empty option + {range Options} + error span
icon(name string) → templ.Raw(svg)           // 12-icon inline SVG subset (templ.Raw ONLY here — REQ-WC6-017)
fieldError(name string, errs map[string]string) // { if errs[name] != "" } field-error span with alert-circle icon
```

### A.1 Typed args replace the `dict` helper

`html/template` had no map constructor, so the template used a hand-rolled `dict "Name" x "Title" y …` FuncMap. Templ takes typed Go parameters, so the `dict` helper is **retired** and replaced by typed structs:

```
type optSelectArgs struct {
    Name, Title, Key, Desc, Value, Empty string
    Options []string
    Errors  map[string]string
}
type langSelectArgs struct { Name, Title, Key, Desc, Value string; Options []string; Errors map[string]string }
```

A renamed view-model field is now a compile error instead of a render-time `<no value>` — this type safety is a core migration benefit and the prerequisite for 007's section-registry.

## §B Markup-parity contract (the porting SSOT)

[HARD] Each Templ component MUST emit markup that keeps the Class A tests green. The exact contract, derived from the verified `page.html.tmpl` + the Class A tests:

| Component | MUST emit (exact) | Asserted by |
|-----------|-------------------|-------------|
| `optSelect` / `langSelect` | `<div class="field {has-error if error}">`, `<span class="field__title" data-i18n="f.{name}.title">`, `<code class="field__key">{key}</code>`, `<span class="field__desc" data-i18n="f.{name}.desc">`, `<span class="select-wrap">`, `<select class="select[ select--lang]" id="{name}" name="{name}"{ aria-invalid="true" if error}>`, `<option value="" {selected if value==""}>{Empty/(unset)}</option>`, `{for o := range Options}<option value="{o}" {selected if o==value}>{o}</option>`, `{if errs[name]}<span class="field-error">{icon alert-circle}<span>{msg}</span></span>` | `TestComponentChromePresent`, `TestNameAttributesPreserved`, `TestDataI18nWiring`, `TestProjectFieldsetRendersSelects`, `TestProjectSelectsPreselectCurrentValues`, `TestAccessibilityCues` |
| `appbar` | `class="appbar"`, `class="brand__badge"` + brand SVG, `모두의AI`, `class="loopback"` + `{ view.BindAddr }`, `class="select langpick" id="uiLangSelect" aria-label="Interface language"` (NO `name=`, OUTSIDE form), `id="themeToggle"` | `TestAppbarRendered`, `TestLoopbackIndicatorShowsRealBindAddr`, `TestLangpickRendered`, `TestLangpickNotFormField` |
| `icon(name)` | inline `<svg class="icon-{name}" …>` for the 12 names; NO lucide CDN / `data-lucide` / runtime icon JS | `TestInlineSVGIconsNoCDN` |
| root `page` | `<html lang="en" data-theme="light">`, FOUC `<head>` `<script>` reading `moai-console-theme` + `moai-console-lang`, `<form method="POST" action="/save">`, hidden `__profile`, `<script src="/static/i18n.js">` + `<script src="/static/htmx.min.js">` + `<script src="/static/app.js">`, `banner--success`/`banner--error` mapping from `view.BannerKind` | `TestDarkModeAndThemeToggle`, `TestBannerKindMapping`, `TestStaticAssetsServedFromEmbed`, `TestI18nDictionaryEmbedded` |
| statusline segments | `{ for s := range view.AllSegments }<input type="checkbox" name="segment_{s}">`; NO kebab `segment_<a-z>+-<a-z>`; no `value="es"`/`fr`/`de`; no `haiku[1m]` | `TestNoNonCanonicalOptions` |

### B.1 The §2.1.1 twelve source-coupled (Class C) tests (the expected mechanism updates)

The **spec.md §2.1.1 Class C set — twelve distinct test functions** read the deleted `page.html.tmpl` *source* (via `readEmbeddedAsset(t, "page.html.tmpl")`) OR call the retired `pageTemplate()` symbol (which vanish post-migration). These — NOT a "two-test" subset — are the anticipated mechanism updates. spec.md §2.1.1 is the authoritative inventory with the per-test required mechanism update; the highlights:

1. `TestPageTemplateParses` (§2.1.1 #1) — pure symbol-existence test for `pageTemplate()` → retire (E.5.8 carve-out) OR replace with a `page(view).Render` smoke test.
2. `TestLoopbackIndicatorShowsRealBindAddr` (§2.1.1 #3) — `page.html.tmpl` `{{.BindAddr}}` / no-`127.0.0.1:3041` grep → retarget to the Templ source (`{ view.BindAddr }`, no hardcoded port) / rendered body (the injected-addr body assertion already works).
3. `TestNameAttributesPreserved` (§2.1.1 #5) / `TestServerContractPreserved` (§2.1.1 #12) — `page.html.tmpl` greps for `{{range .AllSegments}}` / `{{with index .FieldErrors}}` / `{{range .LangOptions}}` + `pageTemplate()` `Lookup` → retarget to the Templ source (`for … range view.AllSegments` / `if errs[name]` / option `for`) or the rendered body's structural markers (the body `name=` assertions already work).
4. `TestComponentChromePresent` (§2.1.1 #2), `TestNoNonCanonicalOptions` (#4), `TestDarkModeAndThemeToggle` (#6), `TestInlineSVGIconsNoCDN` (#7), `TestAccessibilityCues` (#8), `TestI18nDictionaryEmbedded` (#9), `TestLangpickNotFormField` (#10), `TestI18nLoadDefault` (#11) — retarget each template-source grep / `Lookup` check to the Templ source / rendered body per §2.1.1.

All twelve are mechanism updates with a per-test justification entry in acceptance.md §D.3. Every Class A test NOT in the §2.1.1 twelve (e.g. `TestAppbarRendered`, `TestBannerKindMapping`, `TestDataI18nWiring`, `TestLangpickRendered`, `TestProjectFieldsetRendersSelects`) asserts the *rendered body* only and is satisfied by exact-markup parity (no test edit).

## §C Render call swap (handlers.go)

`html/template`'s `a.tmpl.Execute(&buf, view)` → Templ's `page(view).Render(ctx, &buf)`. The buffer-first discipline (REQ-WC6-018) is preserved:

```
// before:  var buf bytes.Buffer; err := a.tmpl.Execute(&buf, view); if err { renderError 500 } …
// after:   var buf bytes.Buffer; err := page(view).Render(r.Context(), &buf); if err { renderError 500 } …
```

`assets.go:pageTemplate()` + the `dict` FuncMap are deleted; `a.tmpl *template.Template` field is removed from the `app` struct (or repurposed — but cleaner to remove). The `app` seams (`readPreferences`/`writePreferences`/`syncToProject`/`listProfiles`/`readProjectConfig`/`writeProjectConfig`/`bindAddr`) are UNCHANGED — only the render mechanism changes.

## §D Build / codegen flow

```
.templ sources (committed)
   │  templ generate ./internal/web/...        ← pure-Go tool, run via `go run github.com/a-h/templ/cmd/templ generate`
   ▼
*_templ.go (committed source artifact)          ← compiled like any Go; NOT go:embed'd
   │  go build / go test
   ▼
moai binary  (Templ render compiled in; static assets via go:embed)
```

Wiring points:

- **`//go:generate`** in the `internal/web` package: `//go:generate go run github.com/a-h/templ/cmd/templ generate` → `go generate ./...` regenerates.
- **Makefile**: `generate` target runs `templ generate`; `build` + `test` targets gain a `templ generate` prerequisite (idempotent — no-op on unchanged sources). Existing `build` already runs `gen-catalog-hashes`; `templ generate` is added as a sibling pre-step.
- **CI mirror** (`scripts/ci-mirror/run.sh`) + GitHub CI: `templ generate` before `go build`/`go test`; a drift-guard regenerates and `git diff --exit-code`s the `*_templ.go` (catches a stale committed generate).
- **golangci-lint**: no config change needed. The repo has no `.golangci.yml` (verified); golangci-lint skips generated files by default, and `templ generate` emits the `// Code generated by templ - DO NOT EDIT.` header that triggers that default generated-file exclusion. M1 verifies the emitted `*_templ.go` carries the DO-NOT-EDIT header. IF a `.golangci.yml` is later added, it MUST preserve the default `issues.exclude-generated` behavior.
- **`go:embed` coexistence**: the Templ output is compiled Go, NOT embedded. `assets.go`'s `go:embed` enumerates only static assets — it DROPS `page.html.tmpl` (no longer a runtime asset) and ADDS `htmx.min.js`. The `assets/fonts` glob is unchanged.

### D.1 Decision D3 — commit the generated `*_templ.go`

**RECOMMEND: commit `*_templ.go`** (treat as a source artifact, like other committed generated Go in the repo) + a CI drift-guard. Rationale: `go build ./...` on a bare clone (no `templ` tool invocation) must work — committing the generated files preserves the project's current "clone → `go build` works" posture (today everything is `go:embed`'d source). The drift-guard (CI runs `templ generate` + `git diff --exit-code`) prevents a stale committed generate. Alternative (gitignore + regen-in-CI) is REJECTED: it breaks `go build` on a bare clone, a regression. Record the chosen option in progress.md.

## §E HTMX foundation (M4) — what lands vs what defers

| Lands in 006 | Defers to 007 (S2b) |
|--------------|---------------------|
| `htmx.min.js` self-hosted + `go:embed` + `/static/htmx.min.js` + linked | section-scoped `hx-post` (save one section) |
| progressive-enhancement `hx-*` attrs that degrade to full-page POST | `hx-target` / `hx-swap` section navigation |
| full-page `/save` response shape preserved (REQ-WC6-014) | `/save` fragment / partial response shape |
| (theme/i18n/segment stay vanilla — REQ-WC6-016) | section-scoped save endpoints |

The graceful-degradation contract: with HTMX/JS disabled, the form submits normally and the server returns the full page — exactly the pre-migration behavior. The Class B full-page tests (`TestSaveValidRoundTrip`, `TestGoldenPath_*`, EC-2 atomic-reject) therefore stay green whether or not `hx-*` attrs are present.

## §F How this seeds 007 (S2b) — the architectural payoff

007 must add ~6 nested config sections (quality beyond development_mode, workflow, git-strategy, harness, nested git-convention/llm) with section-scoped editing. 006's architecture makes that tractable:

1. **Typed, composable components.** Each new section in 007 is a new typed Templ component (`fieldsetQuality(view)`, `fieldsetWorkflow(view)`, …) added to the root's component list — not a hand-edited block in a ~339-line monolith. A section registry (a `[]sectionComponent` the root iterates) is the natural 007 extension of the 006 component tree.
2. **HTMX partial-swap ready.** Each section component can render standalone (a Templ component is renderable in isolation), so 007 wires `hx-post` to a section-scoped endpoint that returns just that section's component as a fragment, swapped via `hx-target`/`hx-swap`. 006 embedded `htmx.min.js` and proved the progressive-enhancement wiring; 007 only adds the partial-swap targets.
3. **Server-canonical validation reused.** 007's new sections validate via the same `validate.go` predicate pattern (extended for the new fields) — 006 keeps that boundary intact, so 007 extends rather than re-architects.

006 implements NONE of this (E.1) — it ports the 5 fieldsets 1:1. The section-registry + partial-swap are 007's work; 006 only makes them a small addition instead of a rewrite.

## §G Anti-patterns (design-level)

- **AP-1**: emitting markup that differs from the `html/template` output in class names / `data-i18n` / `name=` — breaks Class A. Mitigation: §B parity contract is the porting SSOT.
- **AP-2**: `templ.Raw` on a view-model value (XSS) — `templ.Raw` is ONLY the inline SVG icon set (REQ-WC6-017).
- **AP-3**: adding `hx-target`/`hx-swap` partial-swap "while we're here" — that is 007 (E.3); 006 keeps full-page `/save`.
- **AP-4**: gitignoring `*_templ.go` → bare-clone `go build` breaks (D3 rejects this).
- **AP-5**: a section-registry that admits a 007 field — 006 ports 5 fieldsets 1:1, zero new field (E.1).
