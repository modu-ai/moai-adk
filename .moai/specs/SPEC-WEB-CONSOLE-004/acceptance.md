# Acceptance Criteria — SPEC-WEB-CONSOLE-004

> Full Given-When-Then enumeration for the 모두의AI design-system application (visual restyle, zero server-contract change). `spec.md §3` is the SSOT index; this file is the authoritative per-AC enumeration. Closure gate: `go test ./internal/web/... ./internal/cli/... ./internal/config/...` exit 0.

## §A — Verification Philosophy

A CSS/template restyle cannot be pixel-asserted in `go test`. The AC strategy therefore combines two assertion classes:

1. **Structural assertions** — parse + render the template (via the existing `pageTemplate()` / `newPageView()` path or an `httptest` `GET /`) and assert that the new structure markers are present (appbar, brand badge SVG, field title / `<code>` key chip / segment card classes, styled-select chrome, inline-SVG icons, the `[data-theme]` token block).
2. **Server-contract regression assertions** — grep the template + served assets, and run the existing 001/002/003 invariant tests, to assert that the restyle did NOT break any server contract (`name=` attrs, `{{range}}` option lists, `.FieldErrors` render, canonical option sets, loopback bind, no-auth posture, no direct YAML marshal, no CDN dependency).

"Zero server-contract change" (the MUST-PASS invariant) is verified primarily through class 2: the existing `internal/web` test suite stays green AND `validate.go` is byte-unchanged AND the canonical-option / `name=` regression greps pass.

## §B — How "zero server-contract change" is verified (cross-cut)

| Contract element | Verification |
|------------------|--------------|
| `name=` POST attrs survive | AC-WC4-009a — grep each canonical `name="…"` present in template; `bindForm`-driven POST round-trip test stays green |
| Option lists server-rendered | AC-WC4-008 / AC-WC4-009a — `{{range}}` still drives option lists; no hardcoded design `<option>`; no non-canonical token present |
| `.FieldErrors` server-side | AC-WC4-009a — `{{with index .FieldErrors …}}` block present; POST atomic-reject re-render test green |
| Canonical option sets | AC-WC4-008 / AC-WC4-009b — `validatePrefs`/`validateProjectConfig` + `validate.go` byte-unchanged; web↔TUI parity tests green |
| Loopback bind / no-auth | AC-WC4-009b — `server.go`/`app.go` invariant + host-check tests green; no `0.0.0.0`/auth/token/session added |
| No direct YAML marshal | AC-WC4-009b — persistence seams untouched (004 writes no config); no `yaml.Marshal`/`os.WriteFile` added to `internal/web` |
| go:embed offline delivery | AC-WC4-001/007/012 — zero external `https://` font/style/script URL; assets served from `go:embed`; build green |

---

## §C — Acceptance Criteria (Given-When-Then)

### AC-WC4-001 — Token layer embedded, no font CDN (REQ-WC4-001)

- **Given** the restyled console with the embedded 모두의AI token CSS layer
- **When** the served console CSS/HTML is inspected (`grep -rn 'fonts.googleapis.com\|@import url("http\|https://fonts' internal/web/assets/`)
- **Then** the grep returns **0 matches** — the Google-Fonts `@import` is removed and no external font/style URL remains
- **And** the token CSS contains the 모두의AI brand tokens (`--color-primary: #144a46`, `--color-bg: #f3f3f3`, `--gradient-signature`) and the type/spacing/radius/shadow/motion scales
- **And** the token CSS is delivered via the web package `go:embed` (not a network fetch)

### AC-WC4-002 — Self-hosted Pretendard woff2 subset + license (REQ-WC4-002)

- **Given** the font assets under the web package tree
- **When** `ls internal/web/assets/fonts/` is run
- **Then** at least one Pretendard **woff2** subset file is present (NOT the 9× OTF set) **and** an OFL-1.1 license file (`OFL.txt` / `LICENSE`) is present
- **And** the `@font-face src` in the token CSS references the font by a **relative** path (no `https://`)
- **And** the font files are enumerated by the `assets.go` `go:embed` directive (build embeds them)
- **Edge — non-default weight**: only the weights the restyle uses are subset-shipped; an unused weight is not required to be present

### AC-WC4-003 — Component chrome ported onto the Go template (REQ-WC4-003)

- **Given** a `GET /` render of the restyled page
- **When** the rendered HTML is inspected
- **Then** the component chrome markers are present: section card (e.g. `.section`/`.section__head`) per fieldset, field chrome (`.field__title` + `<code class="field__key">` key chip + `.field__desc`), styled `<select>` chrome (chevron affordance), segment checkbox cards, banner chrome, signature-gradient primary button
- **And** the `{{define "langSelect"}}` and `{{define "optSelect"}}` blocks are still present in the template and still used by the Language / Launch / Project fields (structure preserved)
- **And** the 5 fieldsets (Identity / Language / Launch / Statusline / Project) are all rendered in the same logical grouping as before

### AC-WC4-004 — New appbar (brand + loopback + theme toggle, NO langpick) (REQ-WC4-004)

- **Given** a `GET /` render
- **When** the appbar region is inspected
- **Then** it contains the signature-gradient brand badge (Moai-mark inline SVG), the "모두의AI" brand name, the loopback indicator, and the light/dark theme-toggle button
- **And** the appbar contains **NO** interface-language picker (`langpick` / an appbar interface-language `<select>`) — that is S3 scope (§4 E.1)
- **Negative**: `grep -n 'langpick\|id="langSelect"' internal/web/assets/page.html.tmpl` returns 0 matches (the appbar interface-language widget is absent)

### AC-WC4-005 — Loopback indicator shows the real bound address (REQ-WC4-005)

- **Given** a console server bound on a non-default port (e.g. `127.0.0.1:7777`)
- **When** `GET /` is rendered
- **Then** the loopback indicator shows `127.0.0.1:7777` (the real bound address), NOT the design placeholder `3041`
- **And** the address is server-rendered from a view-model field (e.g. `{{.BindAddr}}`) populated from the server's bind knowledge (`Server.listenerAddr()`), NOT hardcoded in the template
- **Test approach**: an `httptest`/handler test that sets a known bind address in the view-model and asserts the rendered indicator contains it; assert the template has no literal `127.0.0.1:3041`

### AC-WC4-006 — Dark mode + persisted client-side theme toggle (REQ-WC4-006)

- **Given** the restyled console
- **When** the token CSS + page + JS are inspected
- **Then** the `[data-theme="dark"]` (or `[data-theme]`) override block is present in the token CSS (dark bg/surface/primary/ink/border tokens)
- **And** the theme-toggle element is present in the appbar
- **And** a FOUC-prevention inline `<head>` snippet applies the persisted theme before first paint
- **And** the `prefers-reduced-motion` motion guard is present
- **And** theme persistence is **client-side only** (`localStorage`) — there is NO server round-trip and NO profile/config field added for theme (negative: no theme key in the persistence path)

### AC-WC4-007 — Inline-SVG icon subset, no icon CDN (REQ-WC4-007)

- **Given** the restyled page
- **When** `grep -rn 'unpkg.com\|lucide@\|data-lucide\|cdn' internal/web/assets/` is run
- **Then** there is **NO** lucide CDN `<script>` and **NO** `data-lucide` runtime-hydration markup (grep = 0 matches for the CDN/runtime patterns)
- **And** the ~14 UI icons render as inline `<svg>` elements (or from a small embedded SVG sprite) — `grep -c '<svg' internal/web/assets/page.html.tmpl` ≥ the icon count, OR the sprite asset is embedded
- **And** no icon-library runtime JS is loaded

### AC-WC4-008 — No non-canonical design options introduced (REQ-WC4-008)

- **Given** the restyled form
- **When** the template + any rendered `GET /` is inspected
- **Then** the form contains **NO** `es`/`fr`/`de` language option, **NO** `haiku[1m]` model option, and **NO** kebab-case statusline segment key (`segment_git-branch`, `segment_session-cost`, etc.)
- **And** all option lists are still `{{range}}`-driven from the canonical view-model (Language from `.LangOptions`, Model from `.ModelOptions`, segments from `.AllSegments`, etc.) — the design's hardcoded `<option>` markup is NOT copied
- **Regression grep (expect 0)**: `grep -rn 'value="es"\|value="fr"\|value="de"\|haiku\[1m\]\|segment_git-branch\|segment_session-cost\|segment_output-style' internal/web/assets/page.html.tmpl`
- **Positive**: the 4 canonical languages (`en/ko/ja/zh`), the 6 canonical models, and the 15 snake_case segment keys remain renderable through their `{{range}}` lists

### AC-WC4-009a — `name=` attrs + form contract preserved (REQ-WC4-009)

- **Given** the restyled template
- **When** the form structure is inspected
- **Then** every form input/select retains its `name=` POST attribute — `grep` confirms `name="user_name"`, `name="conversation_lang"`, `name="git_commit_lang"`, `name="code_comment_lang"`, `name="doc_lang"`, `name="permission_mode"`, `name="model_policy"`, `name="model"`, `name="effort_level"`, `name="statusline_mode"`, `name="statusline_preset"`, `name="statusline_theme"`, `name="segment_…"` (×15), `name="development_mode"`, `name="git_convention"`, `name="__profile"`, `name="__profile_select"` are all present
- **And** `<form method="POST" action="/save{{if .SelectedProfile}}?profile=…{{end}}">` + `<input type="hidden" name="__profile">` + the `{{if .ShowProfileSwitch}}` profile-switch block are preserved
- **And** the `.FieldErrors` server-side per-field error block (`{{with index .FieldErrors …}}`) is present
- **And** a POST round-trip test (valid submission persists; invalid submission re-renders with a per-field error) stays green

### AC-WC4-009b — 001/002/003 invariants + validators unchanged (REQ-WC4-009)

- **Given** the restyle is complete
- **When** the existing `internal/web` invariant tests run
- **Then** the 001/002/003 invariant + integration tests stay **green** (loopback bind, no-auth, host-check, no-direct-marshal, web↔TUI validation parity, DO_NOT_TOUCH sentinels)
- **And** `git diff --exit-code internal/web/validate.go` confirms the canonical lists + `validatePrefs` + `validateProjectConfig` are **byte-unchanged**
- **And** no `0.0.0.0` bind, no auth/token/session, and no `yaml.Marshal`/`os.WriteFile` were added to `internal/web` (negative grep)

### AC-WC4-010 — Banner class mapping preserves server kind (REQ-WC4-010)

- **Given** a POST that the server completes with `.BannerKind = "ok"` (success) or `.BannerKind = "error"`
- **When** the banner renders
- **Then** the banner shows the corresponding 모두의AI chrome (`banner--success` for `"ok"`, `banner--error` for `"error"`) via a template-local mapping
- **And** the server-set `.BannerKind` values remain `"ok"`/`"error"` (the Go handler is NOT changed to emit `"success"`/`"error"` — the mapping is in the template only)
- **Edge — no banner**: when `.Banner` is empty, no banner element renders (the `{{if .Banner}}` guard is preserved)

### AC-WC4-011 — WCAG 2.1 AA preserved (REQ-WC4-011)

- **Given** the restyled console
- **When** accessibility cues are inspected
- **Then** error states use non-color cues (icon + border + text, not color alone)
- **And** a `focus-visible` outline rule is present in the CSS
- **And** the `prefers-reduced-motion` guard is present
- **And** the theme-toggle button and any icon-only controls carry `aria-label`; errored fields carry `aria-invalid`/`aria-describedby` association where the current console already provides them

### AC-WC4-012 — go:embed delivery + cross-platform build (REQ-WC4-012)

- **Given** the new assets (token+component CSS, woff2 subset font(s), inline-SVG/sprite, theme-toggle JS)
- **When** the build runs
- **Then** `go build ./...` exits 0 **and** `GOOS=windows GOARCH=amd64 go build ./...` exits 0
- **And** the `internal/web/assets.go` `go:embed` directive enumerates the new assets (`grep -n 'go:embed' internal/web/assets.go` shows the font/CSS/SVG/JS coverage)
- **And** the assets are served offline with no network fetch (consistent with AC-WC4-001/007)

### AC-WC4-013 — Closure gate (all)

- **Given** all milestones complete
- **When** `go test ./internal/web/... ./internal/cli/... ./internal/config/...` runs
- **Then** it exits **0** (all packages green)
- **And** the §E regression batch (offline grep = 0, non-canonical grep = 0, `name=` survival, `validate.go` unchanged) all pass

---

## §D — Edge Cases & Negative Assertions

| # | Scenario | Expected |
|---|----------|----------|
| EC-1 | Console started on default port (3041) | loopback indicator shows `127.0.0.1:3041` from `BindAddr` (coincidentally matches the design placeholder, but is server-sourced, not hardcoded) |
| EC-2 | Console started on a non-default port | loopback indicator shows the real port, not 3041 (AC-WC4-005) |
| EC-3 | Offline / no network at render time | page renders fully — fonts (woff2 subset) + icons (inline SVG) load from embed; no broken-glyph / empty-icon placeholders (AC-WC4-001/007) |
| EC-4 | A field submitted with a non-canonical value (e.g. `model=haiku[1m]`) | server still rejects in `validatePrefs` and re-renders the per-field error (the restyle did not weaken validation — `validate.go` unchanged) |
| EC-5 | preset != "custom" | segment checkboxes still present in markup (always POSTed); server binds segments only when preset==custom (existing behavior preserved) |
| EC-6 | Dark theme persisted, page reloaded | FOUC `<head>` snippet applies dark theme before first paint; no flash of light theme |
| EC-7 | `prefers-reduced-motion: reduce` set | motion/transition durations collapse per the guard (AC-WC4-006/011) |
| EC-8 | `.review` "State preview" aside | ABSENT from the rendered page (design-tool scaffold excluded — §4 E.4); grep `class="review"` = 0 |
| EC-9 | Interface-language picker | ABSENT from the appbar (S3 scope — §4 E.1); the content-language Language fieldset (`conversation_lang` etc.) remains, unchanged |
| EC-10 | Single profile (no multi-profile) | profile-switch block omitted via `{{if .ShowProfileSwitch}}` (existing REQ-WC-011 behavior preserved) |

---

## §E — Definition of Done

All of the following MUST hold for 004 to close:

- [ ] All 13 AC (AC-WC4-001 .. AC-WC4-013) PASS per the Given-When-Then above.
- [ ] Closure gate `go test ./internal/web/... ./internal/cli/... ./internal/config/...` exit 0.
- [ ] `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- [ ] Offline regression: zero external `https://` font/style/script URL in served assets (Google-Fonts `@import` + lucide CDN `<script>` both removed).
- [ ] Non-canonical option regression: no `es/fr/de`, no `haiku[1m]`, no kebab segment keys in the form; option lists `{{range}}`-driven.
- [ ] `name=` survival: every canonical POST attribute present; `.FieldErrors` server-side render present; form method/action/hidden-profile preserved.
- [ ] `internal/web/validate.go` byte-unchanged (validators + canonical lists untouched).
- [ ] 001/002/003 invariant + DO_NOT_TOUCH sentinel tests green; no `0.0.0.0`/auth/token/session/direct-YAML-marshal added.
- [ ] Pretendard woff2 subset + OFL-1.1 license shipped; lucide ISC acknowledged.
- [ ] `.review` aside, interface i18n / language picker / `i18n.js` / Noto-CJK all ABSENT (S3/E.4 exclusions honored).
- [ ] WCAG 2.1 AA cues (non-color error, focus-visible, prefers-reduced-motion, ARIA) preserved.

## §F — Forward-Looking Checks (not 004 closure, cohort tracking)

These are NOT 004 acceptance criteria — recorded so the cohort tracker knows what 004 deliberately leaves for S3:

- Web interface i18n (`data-i18n` + `i18n.js` + appbar `langpick` + interface-language `localStorage`) → S3.
- Full CJK webfont (Noto Sans SC/JP) for ja/zh interface text → S3.
- Deep nested config editing → S2b; dead-config audit/removal → S4.
