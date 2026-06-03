---
id: SPEC-WEB-CONSOLE-004
title: "Web Console — 모두의AI Design System Application (visual restyle, zero server-contract change)"
version: "0.2.0"
status: completed
created: 2026-06-03
updated: 2026-06-03
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "internal/web"
lifecycle: spec-anchored
tags: "web, console, design, ui, modu-ai, pretendard, dark-mode, offline, restyle"
tier: M
related_specs: [SPEC-WEB-CONSOLE-001, SPEC-WEB-CONSOLE-002, SPEC-WEB-CONSOLE-003]
---

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-06-03 | manager-spec | Initial draft — visual-restyle member (004) of the web-console-v3 cohort. Applies the 모두의AI design system (token layer + console.css component port + new appbar + field chrome + dark mode + theme toggle + inline-SVG icon subset + self-hosted Pretendard woff2 subset) to the `moai web` console, formalizing `.moai/design/web-console-handoff/IMPLEMENTATION-HANDOFF.md`. MUST-PASS invariant: **zero server-contract change** (§3) — `name=` POST attrs, server-rendered `{{range}}` option lists, `.FieldErrors` server-side rendering, canonical option sets, loopback bind, no-auth posture, `go:embed` delivery all preserved. Web interface i18n + full CJK webfont deferred to cohort S3; deep nested config editing deferred to S2b. |

---

## §1 Context & Motivation

### 1.1 Where this SPEC sits in the cohort

The `web-console-v3` cohort hardens and extends the loopback-only browser settings editor (`moai web`):

- **001** (모태, in-progress) — the original loopback-only, no-auth profile/project settings editor.
- **002 / S1** (completed) — port `3041` default + web↔TUI validation parity (`model` / `effort_level` / `model_policy`).
- **003 / S2a** (in-progress) — flat project-config parity (`development_mode` + `git_convention.convention`).
- **004 / THIS SPEC** — the **visual layer**: apply the 모두의AI design system to the console.

004 is the *visual restyle*. The console's server behavior (form fields, validation, persistence, bind posture) is **frozen**; only the rendered HTML/CSS/asset layer changes. This SPEC formalizes the already-authored handoff guide `.moai/design/web-console-handoff/IMPLEMENTATION-HANDOFF.md` (the research + design basis — it carries the design→Go mapping table §3, the 11-item server-contract preservation list §4, the canonical discrepancy table §5, the cohort decomposition §6, the font/icon offline strategy §7/§8, the step plan §9, and the acceptance basis §10). This SPEC builds the requirements **from** that guide rather than re-deriving them.

### 1.2 The design input

A Claude Design session produced a static HTML/CSS/JS prototype implementing the 모두의AI design system. The prototype's intent is "framework-agnostic, reproduce pixel-for-pixel but do not copy the internal structure" — i.e. mimic the **visual** output while keeping the current Go `html/template` server-render structure intact. The design artifacts are preserved under `.moai/design/web-console-handoff/from-claude-design/`:

- `MoAI-Web-Console.html` — design markup (single page, 5-section form + new appbar + design-review aside).
- `assets/colors_and_type.css` — the 모두의AI **FROZEN token layer** (brand 3-color `#144a46` / `#09110f` / `#f3f3f3`, signature gradient, Pretendard self-host, Inter/JetBrains-Mono, dark-mode `[data-theme="dark"]` overrides, spacing/radius/shadow/motion scales). This is the visual SSOT — token values are not arbitrarily changed.
- `assets/console.css` — component styles (section cards, field chrome, selects, segment cards, banner, buttons).
- `screenshots/` — `console-light.png` (target look), `states.png` (banner/validation/profile states), `segments.png` (segment cards), `appbar.png` (brand strip), `ko.png` (Korean i18n — S3 scope). Visual reference for pixel verification.

### 1.3 The current Go console (the implementation target — verified ground-truth)

The console at `internal/web/assets/page.html.tmpl` server-renders 5 `<fieldset>` groups (Identity / Language / Launch / Statusline / Project) + an actions bar via `html/template`, embedded through the web package's own `go:embed` directive (`internal/web/assets.go:13` — `//go:embed assets/style.css assets/app.js assets/page.html.tmpl`). Static CSS/JS is served from `http.FS(staticFS())` (`internal/web/app.go routes()`). The design has the same 5-section structure, so the structural mapping is close to 1:1. Verified anchors:

| Concern | Verified ground-truth |
|---------|------------------------|
| Form fields use `name=` | `page.html.tmpl` — all inputs/selects carry `name=` (`user_name`, `conversation_lang`, `segment_<key>`, `__profile`, `__profile_select`, etc.); the server reads them via `r.PostFormValue` in `handlers.go bindForm` |
| Option lists are server-rendered | `{{range .LangOptions}}` / `{{range .PermissionModes}}` / `{{range .AllSegments}}` / `{{range .ModelOptions}}` etc. populated by `newPageView()` in `handlers.go` — NOT hardcoded `<option>` |
| Errors are server-rendered | `{{with index .FieldErrors .Name}}` per-field error blocks (POST atomic reject re-renders the form server-side) |
| Banner is server-rendered | `{{if .Banner}}<div class="banner {{.BannerKind}}">` — `.BannerKind` is `"ok"` / `"error"` (set in `handlers.go`); the design uses `banner--success` / `banner--error` (class mapping needed) |
| Helpers exist | `{{define "langSelect"}}` (`page.html.tmpl:120`) + `{{define "optSelect"}}` (`page.html.tmpl:133`) — reused across language/launch/project fields |
| Loopback bind | `internal/web/server.go:31` `const loopbackHost = "127.0.0.1"`; `Server.listenerAddr()` knows the real bound `127.0.0.1:<port>` address |
| No-auth + Host-check | `internal/web/app.go hostCheckMiddleware` gates POST/PUT/PATCH on a loopback Host header; no token/session/CSRF |
| Canonical option SSOT | `internal/web/validate.go` mirror lists (`langOptions` = en/ko/ja/zh; `modelCanonical` = 6 values; `allSegments` = 15 snake_case keys; etc.) + `validatePrefs` + `validateProjectConfig` |

### 1.4 The discrepancy: design options vs Go canonical (the canonical risk surface)

The design prototype hardcodes `<option>` sets that diverge from the Go canonical lists in three fields. Adopting a non-canonical option causes the server to reject the mutation in `validatePrefs` / `validateProjectConfig`. The handoff guide §5.1 is the canonical discrepancy table; the three drops are:

| Field | Design options | Go canonical (SSOT) | Action |
|-------|----------------|---------------------|--------|
| Language (×4 selects) | `ko/en/ja/zh/es/fr/de` (7) | `en/ko/ja/zh` (4) — `langOptions` (`validate.go:26`) | **drop** `es/fr/de`; render the 4 canonical + empty `(unset)` |
| Model | `…/haiku/haiku[1m]/opusplan` (7) | `opus/opus[1m]/sonnet/sonnet[1m]/haiku/opusplan` (6) — `modelCanonical` (`validate.go:33`) | **drop** `haiku[1m]`; render the 6 canonical + empty `(project default)` |
| Statusline segments | 15 **kebab-case** keys (`git-branch`, `session-cost`, …) | 15 **snake_case** keys (`git_branch`, `usage_5h`, …) — `allSegments` (`validate.go:75`) | **drop all 15 design keys**; render the 15 Go canonical keys (`name="segment_<key>"`) |

The reverse case also holds: `permission_mode` — the design hardcodes a **subset** (5 values) that omits the canonical `auto` / `dontAsk`. The Go side keeps its full `{{range .PermissionModes}}` set (empty + 6 real values) to preserve web↔TUI parity. All other fields (`model_policy`, `effort_level`, statusline mode/preset/theme, `development_mode`, `git_convention`) are canonical-aligned — adopt as-is.

### 1.5 Offline conflict: the design's CDN dependencies break the loopback invariant

The console is loopback-only (`127.0.0.1`) + no-auth + **zero network dependency** (an 001/S1 design invariant). The design prototype violates this in two places:

- `colors_and_type.css` line 26 — `@import url("https://fonts.googleapis.com/css2?…")` pulls Inter + JetBrains Mono from the **Google Fonts CDN**.
- `MoAI-Web-Console.html` — `<script src="https://unpkg.com/lucide@latest/…">` pulls the **lucide icon library** from the unpkg CDN.

In an offline / network-restricted environment these CDN loads fail (fonts fall back to system-ui → brand typography lost; `<i data-lucide>` icons render as empty placeholders). The handoff guide §7/§8 invalidate the CDN approach. This SPEC adopts the recommended offline-safe replacements: **self-hosted Pretendard woff2 subset** (Latin+Hangul, only the used weights) via `go:embed`, and an **inline-SVG icon subset** (~14 icons) embedded directly in the template. Both are network-0, satisfying the loopback invariant.

> **Font scope note**: 004 ships the **Pretendard Latin+Hangul woff2 subset** required for the base brand look. The Inter / JetBrains-Mono Google-Fonts `@import` is **removed** (replaced by self-host or by the system/mono fallback). The full CJK webfont (Noto Sans SC/JP, needed for ja/zh *interface* text) is **S3 scope** — 004's typography covers the base look without the interface-i18n CJK glyph coverage.

### 1.6 Cohort scope-fence

This SPEC is **004** (visual restyle). Siblings are referenced ONLY to delimit what 004 does NOT do (see §4):
- **S3** — web interface i18n (`data-i18n` switching + appbar language picker + `i18n.js` dictionary, en/ko/ja/zh) + full CJK webfont (Noto Sans SC/JP).
- **S2b** — deep nested config editing (quality/workflow/git-strategy/harness + nested git-convention/llm sub-fields).
- **S4** — dead-config audit + removal.

---

## §2 GEARS Requirements

### REQ-WC4-001 (Ubiquitous — 모두의AI token layer embedded, offline-safe)

The Console **shall** ship the 모두의AI design-token layer (brand 3-color palette, signature gradient, type scale, spacing / radius / shadow / motion tokens, and the `[data-theme="dark"]` override block) as an **embedded** CSS sheet delivered via the web package's existing `go:embed` mechanism, with **no** Google-Fonts CDN `@import` and **no** other network font/style fetch — preserving the loopback-only, zero-network invariant of `SPEC-WEB-CONSOLE-001`.

### REQ-WC4-002 (Ubiquitous — self-hosted Pretendard woff2 subset)

The Console **shall** serve the Pretendard brand font as a **self-hosted woff2 subset** (Latin + Hangul coverage, restricted to the weights the restyle actually uses) embedded via `go:embed` and referenced by relative `@font-face src` from the embedded token CSS, with the **OFL-1.1 license file included** alongside the font assets. The Console **shall not** fetch any font over the network at runtime (the Google-Fonts `@import` for Inter / JetBrains-Mono is removed; the base look relies on the Pretendard subset plus system / monospace fallbacks).

### REQ-WC4-003 (Ubiquitous — console.css component port onto the Go template)

The Console page **shall** render the 모두의AI component chrome (section cards with section-icon + legend, field chrome [per-field title + `<code>` key chip + description text], styled `<select>` with chevron affordance, segment checkbox cards, the save banner, and the signature-gradient primary button) by porting `console.css` styles onto the **existing** Go `html/template` structure — preserving the 5-fieldset layout (Identity / Language / Launch / Statusline / Project) and the existing `langSelect` / `optSelect` define blocks (their structure is preserved; only the visual chrome and, where needed, additional title/key/description fields are added).

### REQ-WC4-004 (Ubiquitous — new appbar with brand mark, loopback indicator, theme toggle)

The Console page **shall** render a new top **appbar** containing: the signature-gradient brand badge (Moai-mark inline SVG) + the "모두의AI" brand name, a divider, a **loopback indicator** showing the actual bound address `127.0.0.1:<port>` (server-rendered from a view-model field, NOT a hardcoded `3041`), and a **light/dark theme toggle** button (sun/moon icon). The appbar **shall not** contain an interface-language picker (that widget is web-i18n = cohort S3 scope, excluded here per §4 E.1).

### REQ-WC4-005 (Event-driven — loopback indicator reflects the real bind address)

**When** the Console renders `GET /`, the loopback indicator **shall** display the server's actual bound loopback address (sourced from the existing `Server.listenerAddr()` / bind-address knowledge and injected as a view-model field such as `BindAddr`), so a console started on a non-default port displays its real `127.0.0.1:<port>` rather than the design prototype's placeholder `3041`.

### REQ-WC4-006 (State-driven — dark mode via `[data-theme]` + persisted toggle)

**While** the dark theme is active (`<html data-theme="dark">`), the Console **shall** apply the 모두의AI dark-mode token overrides (dark background / surface / primary / ink / border tokens). The theme toggle **shall** flip `data-theme` and persist the choice **client-side** (`localStorage`), with a FOUC-prevention inline `<head>` snippet applying the persisted theme before first paint, and a `prefers-reduced-motion` guard preserved. Theme persistence **shall** remain client-side only (no server round-trip, no profile/config write).

### REQ-WC4-007 (Ubiquitous — inline-SVG icon subset, offline-safe)

The Console page **shall** render its ~14 UI icons (user-round / languages / rocket / panel-bottom / folder-git-2 / chevron-down / sun / moon / save / check / alert-circle / x, plus the banner-state icons) as **inline SVG** embedded directly in the template (or a small embedded SVG sprite), with **no** lucide CDN `<script>` and **no** runtime icon-library JS, preserving the zero-network invariant. The lucide ISC license **shall** be acknowledged alongside the icon subset.

### REQ-WC4-008 (Unwanted behavior — no non-canonical design options introduced)

The Console **shall not** introduce the design prototype's non-canonical option values into the rendered form: it **shall not** add the `es` / `fr` / `de` languages (canonical = `en/ko/ja/zh`), **shall not** add the `haiku[1m]` model (canonical = the 6 `modelCanonical` values), and **shall not** substitute the design's kebab-case statusline segment keys for the canonical snake_case `allSegments` keys. All option `<option>` lists **shall** remain server-rendered from the canonical view-model lists (`{{range …}}`), never hardcoded from the design markup.

### REQ-WC4-009 (Ubiquitous — server-contract preservation, zero behavior change)

The Console **shall** preserve every server contract unchanged through the restyle: (a) every form input/select keeps its `name=` POST attribute (the design's `id=`-only markup is augmented, not adopted); (b) all option lists stay server-rendered via `{{range}}` from the canonical view-model; (c) `.FieldErrors` per-field errors stay **server-side** rendered (`{{with index .FieldErrors …}}`) — the design's client-side validation preview is NOT adopted as the source of truth; (d) the `<form method="POST" action="/save…">` + `<input type="hidden" name="__profile">` + the `{{if .ShowProfileSwitch}}` conditional + `name="__profile_select"` are preserved; (e) persistence remains exclusively through `WritePreferences` / `SyncToProjectConfig` / the config-manager API (no direct `yaml.Marshal`/`os.WriteFile` from the web layer); (f) the loopback-only bind, no-auth / no-token / no-session posture, and Host-header write-safety check are unchanged; (g) the `langSelect` / `optSelect` helpers are reused (structure preserved); (h) the S1 (port + web↔TUI parity) and S2a (project-config scope fence) invariants are not widened.

### REQ-WC4-010 (Event-driven — banner class mapping preserves server-set kind)

**When** the server sets `.BannerKind` to `"ok"` or `"error"` on a POST round-trip, the Console **shall** render the banner with the corresponding 모두의AI banner chrome (success / error variant) by mapping the server-set kind to the design's `banner--success` / `banner--error` visual classes in the template — **without** changing the server-set `.BannerKind` values (`"ok"`/`"error"` remain the server contract; the mapping is template-local).

### REQ-WC4-011 (Ubiquitous — WCAG 2.1 AA preservation)

The Console **shall** preserve accessibility through the restyle: non-color error cues (icon + border + text, not color alone), `focus-visible` outlines on interactive elements, the `prefers-reduced-motion` motion guard, and ARIA attributes (`aria-label` on the theme toggle and icon-only controls; `aria-invalid` / `aria-describedby` association between an errored field and its `.FieldErrors` message where the current console already provides them).

### REQ-WC4-012 (Ubiquitous — go:embed delivery + cross-platform build)

All new assets (the embedded token + component CSS, the Pretendard woff2 subset font files, the inline-SVG icon subset or sprite, any restyle-specific JS for the theme toggle) **shall** be delivered through the web package's `go:embed` directive (extending `internal/web/assets.go`), such that `go build ./...` and the cross-platform build (`GOOS=windows GOARCH=amd64 go build ./...`) succeed and the assets are served offline with no network fetch.

---

## §3 Acceptance Criteria (summary — full enumeration in acceptance.md)

Each AC is independently verifiable. Because a CSS/template restyle is hard to unit-test for pixels, the AC strategy leans on **structural assertions** (template parses + the new markers render) and **server-contract regression assertions** (grep that the contract elements survive). The closure gate is `go test ./internal/web/... ./internal/cli/... ./internal/config/...` exit 0. The full Given-When-Then enumeration with edge cases lives in `acceptance.md`; the table below is the SSOT index.

| AC | REQ | Assertion (one-line) |
|----|-----|----------------------|
| AC-WC4-001 | REQ-WC4-001 | Rendered page links/embeds the 모두의AI token layer; **no** `fonts.googleapis.com` `@import` and no external style/font URL anywhere in served CSS/HTML (grep = 0 matches) |
| AC-WC4-002 | REQ-WC4-002 | Pretendard woff2 subset font file(s) present under the web assets tree + embedded by `go:embed`; OFL-1.1 license file present; `@font-face src` is relative (no `https://`) |
| AC-WC4-003 | REQ-WC4-003 | Page renders the new component chrome markers (section card / field title / `<code>` key chip / segment card / styled select); `langSelect` + `optSelect` define blocks still present and used |
| AC-WC4-004 | REQ-WC4-004 | Page renders the appbar with brand badge SVG + "모두의AI" + loopback indicator + theme-toggle button; **no** interface-language `<select>` / langpick element in the appbar (S3 exclusion) |
| AC-WC4-005 | REQ-WC4-005 | `GET /` loopback indicator shows the real bound `127.0.0.1:<port>` from a view-model field (e.g. `BindAddr`), not a hardcoded `3041`; a non-default port renders correctly |
| AC-WC4-006 | REQ-WC4-006 | `[data-theme]` override block present in token CSS; theme-toggle element present; FOUC `<head>` inline init present; `prefers-reduced-motion` guard present; no server-side theme write |
| AC-WC4-007 | REQ-WC4-007 | Icons render as inline SVG (`<svg>` markers present); **no** `unpkg.com`/lucide CDN `<script>` and no icon-library runtime JS (grep = 0 matches) |
| AC-WC4-008 | REQ-WC4-008 | Rendered form contains NO `es`/`fr`/`de` language option, NO `haiku[1m]` model option, NO kebab-case `segment_git-branch`-style key; all option lists still `{{range}}`-driven (regression grep) |
| AC-WC4-009a | REQ-WC4-009 | Every form field retains its `name=` POST attribute (grep the rendered/template for each canonical `name="…"`); `.FieldErrors` server-side render block present; form `method="POST" action="/save…"` + hidden `__profile` present |
| AC-WC4-009b | REQ-WC4-009 | Existing 001/002/003 invariant tests stay green (loopback bind, no-auth, host-check, no-direct-marshal, validation parity); `validatePrefs`/`validateProjectConfig` byte-unchanged; persistence still via profile/sync/config-manager seams |
| AC-WC4-010 | REQ-WC4-010 | Server-set `.BannerKind` (`"ok"`/`"error"`) maps to `banner--success`/`banner--error` chrome in the template; the server-set kind values are unchanged (still `"ok"`/`"error"`) |
| AC-WC4-011 | REQ-WC4-011 | Error cues non-color (icon+border+text); `focus-visible` rule present; `prefers-reduced-motion` guard present; theme-toggle + icon-only controls carry `aria-label` |
| AC-WC4-012 | REQ-WC4-012 | `go build ./...` exit 0 + `GOOS=windows GOARCH=amd64 go build ./...` exit 0; new assets enumerated in the `go:embed` directive; assets served with no network fetch |
| AC-WC4-013 | all | Closure gate: `go test ./internal/web/... ./internal/cli/... ./internal/config/...` exit 0 |

---

## §4 Exclusions (What NOT to Build)

### Out of Scope

The following are deferred to sibling SPECs or are explicitly excluded; they MUST NOT be implemented in 004:

- **S3 scope** — web interface i18n (`data-i18n` attribute switching + `i18n.js` dictionary + appbar language picker `langpick` + interface-language `localStorage`) AND the full CJK webfont (Noto Sans SC/JP for ja/zh interface text). 004 ships only the Pretendard Latin+Hangul subset for the base look.
- **S2b scope** — deep nested config editing (full quality / workflow / git-strategy / harness sections + nested git-convention / llm sub-fields). 004 changes NO settings fields, adds NO fields, and removes NO fields — it is visual-only.
- **S4 scope** — dead-config audit / removal.
- **Design-review `.review` aside** — the prototype's "State preview" design-review chrome is a design-tool scaffold ("NOT part of the product"), not a product element.
- **Anti-patterns** — adopting non-canonical design options, CDN font/icon fetch, client-side validation as SSOT, direct YAML marshal in the web layer, template-mirroring (detailed in the numbered list below).

[HARD] The following are explicitly **out of scope** for 004 and MUST NOT be implemented:

### E.1 Out of Scope — web interface i18n + appbar language picker (S3)

The `data-i18n` attribute switching mechanism, the `i18n.js` translation dictionary (en/ko/ja/zh interface strings), the appbar interface-language picker (`langpick` `<select>` + `localStorage("moai-console-lang")`), and the FOUC-style early interface-language application are **cohort S3** scope. The 004 appbar carries a brand badge + loopback indicator + theme toggle ONLY — **no** language picker. The web page interface text stays English. (Note: the **content-language settings** — `conversation_lang` etc. in the Language fieldset — are the server-known profile settings and remain in the form unchanged; they are NOT the same thing as interface i18n.)

### E.2 Out of Scope — full CJK webfont (S3)

The full Noto Sans SC / Noto Sans JP webfonts (needed for Chinese / Japanese **interface** glyph coverage once S3 i18n lands) are **S3** scope. 004 ships ONLY the Pretendard Latin+Hangul woff2 subset for the base brand look; ja/zh interface glyph coverage is not a 004 deliverable.

### E.3 Out of Scope — deep nested config editing (S2b)

No new settings fields, no field-set changes, no nested config editing (quality / workflow / git-strategy / harness / nested git-convention / llm sub-fields). 004 is **visual / layout only** — it restyles the exact field set the console already renders (the S1/S2a field set). It does NOT add, remove, or re-scope any setting.

### E.4 Out of Scope — design-review `.review` aside

The `<aside class="review">` "State preview" design-review chrome in `MoAI-Web-Console.html` (and its companion `console.js` `pressGroup`/`.seg-ctrl` logic, `console.css` `.review*`/`.seg-ctrl*` styles, and `i18n.js` `rv.*` keys) is a design-tool scaffold explicitly marked "NOT part of the product." It MUST be entirely excluded from the port.

### E.5 Out of Scope — additional anti-patterns

[HARD] The following are forbidden regardless of sibling-SPEC scope:

1. **No non-canonical option adoption** — the rendered form MUST NOT contain `es`/`fr`/`de` languages, `haiku[1m]` model, or kebab-case segment keys (per the §1.4 / handoff §5.1 discrepancy table). Option lists stay server-rendered via `{{range}}` from the canonical view-model lists; the design's hardcoded `<option>` markup is NOT copied. `validatePrefs` / `validateProjectConfig` and the canonical lists in `validate.go` MUST be **byte-unchanged** by 004 (visual restyle touches templates + assets, not validators).
2. **No CDN font / icon fetch** — the Google-Fonts `@import` and the lucide unpkg `<script>` MUST be removed. Fonts are self-hosted woff2 subset via `go:embed`; icons are inline SVG. The served CSS/HTML MUST contain zero external `https://` font/style/script URL (loopback / zero-network invariant).
3. **No client-side validation as SSOT** — the server's `.FieldErrors` (atomic POST reject re-render) remains the validation source of truth. A client-side validation preview MAY exist as a convenience but MUST NOT replace or bypass server validation. `name=` POST attributes are mandatory on all inputs/selects.
4. **No direct YAML marshal in the web layer** — 004 writes NO config (it is visual-only); the existing persistence seams (`WritePreferences` / `SyncToProjectConfig` / config-manager) are untouched. A `yaml.Marshal`/`os.WriteFile` of any config from `internal/web` remains the forbidden anti-pattern (it was already forbidden in 001 — 004 introduces no new write path).
5. **No template mirroring / `make build`** — `internal/web/assets/*` is embedded via the web package's own `go:embed` (verified `assets.go` pattern), NOT a deployed asset under `internal/template/templates/`. No `make build` / embedded-mirror parity step applies. The new font / CSS / SVG assets live under `internal/web/assets/` (or a sibling embedded dir) and are added to the web package's `go:embed` directive only.
6. **No server-contract change** — `name=` attrs, `{{range}}` server-rendered options, `.FieldErrors` server-side render, form method/action/hidden-profile, loopback bind, no-auth + Host-check, the `langSelect`/`optSelect` helper structure, and the S1/S2a invariants are all preserved (REQ-WC4-009). The restyle is purely the rendered HTML/CSS/asset layer.
7. **No auth / token / session / non-loopback bind** — the no-auth loopback-only posture of 001 is invariant. 004 adds zero security surface.
8. **No theme persistence on the server** — the dark/light theme choice is client-side `localStorage` only; 004 MUST NOT add a profile or project-config field for theme, and MUST NOT add a server round-trip to persist theme (REQ-WC4-006).

---

## §5 References (verified ground-truth)

| Path | Role |
|------|------|
| `.moai/design/web-console-handoff/IMPLEMENTATION-HANDOFF.md` | PRIMARY SOURCE — design→Go mapping (§3), 11-item server-contract preservation (§4), discrepancy table (§5), cohort decomposition (§6), font (§7) / icon (§8) offline strategy, step plan (§9), acceptance basis (§10) |
| `.moai/design/web-console-handoff/from-claude-design/MoAI-Web-Console.html` | design markup — appbar / banner / 5-fieldset / `.review` aside structure (the visual target; structure NOT copied) |
| `.moai/design/web-console-handoff/from-claude-design/assets/colors_and_type.css` | 모두의AI FROZEN token layer (brand 3-color, signature gradient, Pretendard self-host `@font-face`, Google-Fonts `@import` to REMOVE, `[data-theme="dark"]` overrides, spacing/radius/shadow/motion) |
| `.moai/design/web-console-handoff/from-claude-design/assets/console.css` | component styles (section cards, field chrome, selects, segment cards, banner, button) — port target |
| `.moai/design/web-console-handoff/from-claude-design/screenshots/` | `console-light.png` (target look), `states.png`, `segments.png`, `appbar.png` — pixel-verification reference |
| `internal/web/assets/page.html.tmpl` | implementation target — 5-fieldset form + `langSelect`/`optSelect` define blocks (restyle target; `name=` + `{{range}}` + `.FieldErrors` PRESERVE) |
| `internal/web/assets/style.css` | current console CSS — replaced/extended by the embedded token + component layer |
| `internal/web/assets/app.js` | current console JS — extended for the theme toggle (client-side `localStorage`); segment visibility logic preserved |
| `internal/web/assets.go:13` | `//go:embed assets/style.css assets/app.js assets/page.html.tmpl` — extend to enumerate the new font / CSS / SVG / JS assets (REQ-WC4-012) |
| `internal/web/handlers.go` | `pageView` + `newPageView()` (add a `BindAddr` view-model field) + `bindForm` (unchanged) + `.Banner`/`.BannerKind` set sites |
| `internal/web/server.go:31,142` | `const loopbackHost = "127.0.0.1"` + `Server.listenerAddr()` — the real bound address source for the loopback indicator (REQ-WC4-005) |
| `internal/web/app.go` | `hostCheckMiddleware` (no-auth loopback posture PRESERVE) + `routes()` `/static/` `http.FS(staticFS())` (asset serving) |
| `internal/web/validate.go` | canonical lists (`langOptions`/`modelCanonical`/`allSegments`/…) + `validatePrefs` + `validateProjectConfig` — **byte-unchanged** by 004 (E.5.1) |
| `internal/web/integration_test.go` | 001/002/003 invariant + DO_NOT_TOUCH sentinels — MUST stay green (AC-WC4-009b) |
| Pretendard release | `github.com/orioncactus/pretendard` (OFL-1.1) — woff2 subset source (REQ-WC4-002); OR the user's design bundle `project/fonts/` OTF set |
| lucide | inline-SVG icon subset source (ISC license) — ~14 icons (REQ-WC4-007) |
