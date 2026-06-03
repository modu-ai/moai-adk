# Acceptance Criteria — SPEC-WEB-CONSOLE-005

> Web interface i18n (en/ko/ja/zh) + CJK self-host webfont coverage. Tier M. Each AC is independently verifiable. MUST-PASS ACs: **AC-WC5-008a** (interface ≠ content language — POST round-trip byte-identical), **AC-WC5-010a** (server contract / `validate.go` byte-unchanged), **AC-WC5-011** (offline / zero external font/style/script URL). Closure gate: **AC-WC5-013** (`go test ./internal/web/... ./internal/cli/... ./internal/config/...` exit 0).

## §A — Verification Approach

i18n + a font subset are hard to pixel-test, so the AC strategy uses three assertion classes:

- **Structural** — the template renders the `data-i18n` markers + the langpick; the dictionary embeds + parses + defines 4 locales; the CJK font file embeds + serves.
- **Server-contract regression** — a POST round-trip is byte-identical regardless of the active interface language; the langpick is not a form field; `validate.go` is byte-unchanged; existing 001–004 invariant tests stay green.
- **Offline** — the served CSS/HTML/JS contains zero external `https://` font/style/script URL; the dictionary + font are `go:embed` self-hosted.

`renderIndexBody(t, …)` / `newTestApp(t)` (the existing `restyle_test.go` / `handlers_test.go` helpers) render the page for structural greps; `httptest`-driven POST round-trips verify the server-contract invariance.

---

## §B — Given-When-Then Acceptance Criteria

### AC-WC5-001 — i18n dictionary embedded + 4 locales (REQ-WC5-001)

**Given** the web package embeds its static assets via `go:embed`,
**When** the server serves `/static/i18n.js` and a test reads the embedded `assets/i18n.js`,
**Then** the file is present + embedded (served 200; readable from the embed FS), defines `window.MOAI_I18N` with all four locale keys (`en`, `ko`, `ja`, `zh`), and is loaded with no network fetch (it is part of the offline embed).

- Verify: `TestI18nDictionaryEmbedded` reads the embedded `assets/i18n.js`, asserts `window.MOAI_I18N` + the 4 locale keys present; `grep -c 'en:\|ko:\|ja:\|zh:' internal/web/assets/i18n.js` ≥ 4; `/static/i18n.js` served 200.

### AC-WC5-002 — `data-i18n` wiring on chrome, NOT on code chips (REQ-WC5-002)

**Given** the page renders the 모두의AI chrome,
**When** the rendered body is inspected,
**Then** the translatable chrome elements carry `data-i18n` attributes (the `pagehead__sub` subtitle, the 5 section legends + descriptions, the field titles + descriptions, the segment group title + note, the banner text, the primary-button label, the profile label, the `actions__meta` text), **and** the `<code class="field__key">` code chips (`user_name`, `model`, `segment_<key>` etc.) carry NO `data-i18n` (they remain English code tokens).

- Verify: `TestDataI18nWiring` asserts `data-i18n="app.subtitle"`, `data-i18n="sec.identity.title"`, `data-i18n="f.user_name.title"`, `data-i18n="actions.save"` (representative set) present; asserts no `data-i18n` on a `field__key` code chip; `grep -c 'data-i18n' page.html.tmpl` ≥ **25** (concrete lower bound — the dictionary has ~56 product keys but not all map 1:1 to a DOM element: `aria` keys feed `aria-label` attributes, and the banner/error strings are injected/server-rendered rather than statically tagged, so the rendered `data-i18n` count is below the key count). The exact attribute set is fixed at run-phase per R6 boundary verification (the `data-i18n` key set used MUST be a subset of the dictionary key set); the ≥25 floor guards against under-wiring without over-pinning the precise count.

### AC-WC5-003 — appbar langpick present, non-form, accessible (REQ-WC5-003)

**Given** the appbar brand strip (introduced in 004),
**When** the page renders,
**Then** the appbar contains a `langpick` `<select>` offering the four interface locales (en / ko / ja / zh) with an `aria-label` (from the `lang.aria` key), **and** the langpick is placed OUTSIDE the `<form>` element and carries NO `name=` attribute.

- Verify: `TestLangpickRendered` asserts `class="langpick"` + `id="uiLangSelect"` (the non-colliding id — NOT `langSelect`, which is the live content-language helper template) + 4 `<option>` locale values + `aria-label` present in the appbar; asserts the langpick markup appears BEFORE `<form` in the document order (appbar precedes form); asserts the langpick `<select>` has no `name=`.

### AC-WC5-004 — selection applies + persists client-side, no server round-trip (REQ-WC5-004)

**Given** the langpick and the `app.js` wiring,
**When** the user selects an interface language,
**Then** `app.js` wires a `change` listener on the langpick that calls `applyI18n(locale)` (replacing `data-i18n` element text from the active-locale dictionary) and `localStorage.setItem("moai-console-lang", locale)`, with no form submit and no `fetch`/network call.

- Verify: `TestLangpickJSWiring` (or a grep-based assertion on `app.js`) asserts the `change` handler calls `applyI18n` + `localStorage.setItem("moai-console-lang", …)`; asserts no `form.submit()` / `fetch(` in the langpick handler path.

### AC-WC5-005 — persisted language applied on load, default en (REQ-WC5-005)

**Given** a persisted `moai-console-lang` value (or none),
**When** the page loads,
**Then** the init routine applies the persisted locale's translations (mirroring the 004 FOUC theme pattern — an early `<head>` lang apply and/or a DOMContentLoaded `applyI18n`), **and** when no persisted value exists (first visit) or the persisted value is not one of the four valid locales, the interface defaults to `en` (the template's hardcoded English baseline).

- Verify: `TestI18nLoadDefault` asserts the load-time apply path reads `localStorage.getItem("moai-console-lang")` and falls back to `en` for absent/invalid values; the `<head>` lang snippet (or DOMContentLoaded path) is present mirroring the theme FOUC pattern.

### AC-WC5-006 — CJK woff2 subset embedded, OFL-1.1, relative src, subsetted (REQ-WC5-006)

**Given** the ja/zh interface strings require CJK glyph coverage the 004 Latin+Hangul subset lacks,
**When** the assets tree + embed are inspected,
**Then** the CJK woff2 subset font file(s) are present under `internal/web/assets/fonts/` + embedded by `go:embed`, the corresponding OFL-1.1 license file is present, the `@font-face src` is relative (`/static/fonts/…`, no `https://`), and the subset covers exactly the glyph set of the **shipped** `internal/web/assets/i18n.js` ja+zh string values (rv.* stripped) — the glyph set is re-derived from the shipped dictionary at run-phase (NOT from the design file), and the resulting tens-of-KB file (not a multi-MB full CJK font) plus the measured shipped-dictionary glyph count are recorded in progress.md.

- Verify: `TestCJKFontSubsetEmbedded` asserts the CJK woff2 file present + readable from the embed FS + served 200; OFL license present; `@font-face src` in console.css is relative; the CJK file size is within a small bound (e.g. < 200KB total, asserting it is a subset not a full font). The exact glyph set binds to the shipped `i18n.js` ja+zh values; the run-phase records the measured glyph count in progress.md (the spec.md §1.5 number is a provisional estimate).

### AC-WC5-007 — brand typeface covers active language, no en/ko regression (REQ-WC5-007)

**Given** the CJK `@font-face` and the `--font-sans` stack,
**When** the interface language is ja or zh (vs en or ko),
**Then** the CJK `@font-face` is in the active `--font-sans` stack (or a CJK-specific face activated by the `<html lang>`/locale) so ja hiragana/katakana/kanji and zh hanzi render in the embedded webfont rather than `system-ui`, **and** while the interface is en or ko the existing 004 Pretendard Latin+Hangul subset still covers the chrome (the Pretendard `@font-face` block at console.css:18-51 is preserved; en/ko look unchanged).

- Verify: `TestFontStackCoversCJK` asserts the CJK `@font-face` is appended to `--font-sans` (or a lang-activated face) AND the 004 Pretendard `@font-face` block + the 5 Pretendard subset weights are still present (no en/ko regression).

### AC-WC5-008a — MUST-PASS — POST round-trip language-invariant (REQ-WC5-008)

**Given** the form and the non-form interface langpick,
**When** the same form is submitted while the interface language is `ja` versus while it is `en`,
**Then** the server-submitted fields are byte-identical (same `name=`/value pairs), the langpick value is NOT present anywhere in the POST body, and the content-language fields (`conversation_lang` / `git_commit_lang` / `code_comment_lang` / `doc_lang`) are unaffected by the interface langpick.

- Verify: `TestInterfaceLanguageDoesNotAlterPOST` submits an identical form payload and asserts the `handlers.go bindForm` result is identical regardless of any interface-language state; asserts no `moai-console-lang` / langpick key in `r.PostForm`. **This is the cohort's core invariant (interface ≠ content language) and a must-pass closure gate.**

### AC-WC5-008b — langpick is not a form field, client-only persistence (REQ-WC5-008)

**Given** the interface langpick,
**When** the page + the persistence path are inspected,
**Then** the langpick has NO `name=` attribute and is NOT inside `<form>`, no profile field or project-config field for interface language exists, and the interface language persists ONLY in `localStorage("moai-console-lang")`.

- Verify: `TestLangpickNotFormField` asserts the langpick `<select>` carries no `name=` and sits outside the `<form>…</form>` span; greps `internal/profile` + `internal/web` for any new interface-language config field = 0; the only persistence is `localStorage`.

### AC-WC5-009 — `rv.*` design-review keys excluded (REQ-WC5-009)

**Given** the design `i18n.js` includes `rv.*` design-review keys for the excluded `.review` scaffold,
**When** the shipped `internal/web/assets/i18n.js` + the rendered page are inspected,
**Then** the shipped dictionary contains NO `rv.*` key and the page renders NO element bound to an `rv.*` key (no `.review` aside).

- Verify: `grep -c 'rv\.' internal/web/assets/i18n.js` = 0; `TestNoReviewKeys` asserts no `data-i18n="rv.…"` in the rendered body and no `.review` aside.

### AC-WC5-010a — MUST-PASS — server contract / validate.go byte-unchanged (REQ-WC5-010)

**Given** 005 is i18n + font only (the interface language is client-only),
**When** the run-phase diff is inspected,
**Then** `validatePrefs` / `validateProjectConfig` + the canonical lists in `validate.go` are byte-unchanged (`git diff --exit-code internal/web/validate.go`), every form field retains its `name=` + `{{range}}` server-rendered option, and the `.FieldErrors` per-field errors stay server-side rendered.

- Verify: `git diff --exit-code internal/web/validate.go` → exit 0; `TestServerContractPreserved` asserts each canonical `name="…"` present + option lists `{{range}}`-driven + `{{with index .FieldErrors …}}` blocks present. **Must-pass.**

### AC-WC5-010b — 001–004 invariants green, S1/S2a unchanged (REQ-WC5-010)

**Given** the existing cohort invariant tests,
**When** the full web test suite runs,
**Then** the 001/002/003/004 invariant + DO_NOT_TOUCH sentinel tests stay green (loopback bind, no-auth, host-check, no-direct-marshal, validation parity, the 004 restyle assertions — except the intentionally inverted `TestAppbarRendered` guard, AC-WC5-012), and the web↔TUI parity (S1) + project-config scope (S2a) invariants are unchanged.

- Verify: `go test ./internal/web/...` green (the only intentional test change is the inverted `TestAppbarRendered` per AC-WC5-012); no S1/S2a scope widening.

### AC-WC5-011 — MUST-PASS (offline) — zero external URL, cross-platform build, embed (REQ-WC5-011)

**Given** the loopback / zero-network invariant,
**When** the build + the served assets are inspected,
**Then** `go build ./...` exits 0 + `GOOS=windows GOARCH=amd64 go build ./...` exits 0, the new `i18n.js` is enumerated in the `go:embed` directive (the `assets/fonts` glob covers the new CJK font), and the served CSS/HTML/JS contains zero external `https://` font/style/script URL.

- Verify: both builds exit 0; `grep -n 'go:embed' internal/web/assets.go` enumerates `assets/i18n.js`; `grep -rn 'fonts.googleapis.com\|unpkg.com\|cdnjs\|jsdelivr\|https://fonts\|@import url("http' internal/web/assets/` → 0 matches. **Must-pass (offline).**

### AC-WC5-012 — a11y + 004 test-guard reconciliation (REQ-WC5-012)

**Given** 005 intentionally lands the langpick + `data-i18n` that 004's `TestAppbarRendered` forbade,
**When** the langpick is added + the a11y is verified,
**Then** the langpick carries an `aria-label`, switching the interface language updates the `<html lang>` attribute to the active locale, the 004 a11y cues (focus-visible, prefers-reduced-motion, error ARIA) are unchanged, **and** the 004 `restyle_test.go::TestAppbarRendered` S3-exclusion guard is reconciled — its forbidden-element assertions for `class="langpick"` + `data-i18n` (ABSENT) are inverted to EXPECT those now-present elements PLUS the NEW `id="uiLangSelect"`. The stale `id="langSelect"` forbidden-string is NOT inverted (it referenced the never-landed original id; the appbar picker uses `uiLangSelect`) — it is removed or left harmlessly absent-asserting.

- Verify: `TestAppbarRendered` diff shows `class="langpick"` + `data-i18n` + `id="uiLangSelect"` moved into / added to the expected block (and the stale `id="langSelect"` forbidden entry NOT inverted into an EXPECT); `TestLangAttrUpdatesOnSwitch` asserts the switch path updates `<html lang>`; the langpick `aria-label` present; 004 a11y cues unchanged.

### AC-WC5-013 — closure gate (all REQ)

**Given** all milestones complete,
**When** the closure gate runs,
**Then** `go test ./internal/web/... ./internal/cli/... ./internal/config/...` exits 0.

- Verify: the three-package test invocation exits 0.

---

## §C — AC → REQ Traceability

| AC | REQ | Class | Must-pass? |
|----|-----|-------|------------|
| AC-WC5-001 | REQ-WC5-001 | Structural | — |
| AC-WC5-002 | REQ-WC5-002 | Structural | — |
| AC-WC5-003 | REQ-WC5-003 | Structural | — |
| AC-WC5-004 | REQ-WC5-004 | Structural | — |
| AC-WC5-005 | REQ-WC5-005 | Structural | — |
| AC-WC5-006 | REQ-WC5-006 | Structural / offline | — |
| AC-WC5-007 | REQ-WC5-007 | Structural | — |
| AC-WC5-008a | REQ-WC5-008 | Server-contract regression | **MUST-PASS** |
| AC-WC5-008b | REQ-WC5-008 | Structural | — |
| AC-WC5-009 | REQ-WC5-009 | Structural | — |
| AC-WC5-010a | REQ-WC5-010 | Server-contract regression | **MUST-PASS** |
| AC-WC5-010b | REQ-WC5-010 | Server-contract regression | — |
| AC-WC5-011 | REQ-WC5-011 | Offline / build | **MUST-PASS** |
| AC-WC5-012 | REQ-WC5-012 | Structural / test-reconciliation | — |
| AC-WC5-013 | all | Closure gate | gate |

All 12 REQs are covered by at least one AC (REQ-WC5-008 → AC-008a/008b; REQ-WC5-010 → AC-010a/010b; the rest 1:1). No orphan REQ, no orphan AC.

---

## §D — Edge Cases

| # | Edge case | Expected behavior |
|---|-----------|-------------------|
| EC-1 | `localStorage` unavailable / disabled | langpick still switches the current page view (in-memory `applyI18n`); persist silently no-ops in a `try/catch` (mirrors the 004 theme `applyTheme` try/catch). No exception surfaces. |
| EC-2 | Persisted `moai-console-lang` is a stale/invalid value (e.g. `es`) | Init falls back to `en` (REQ-WC5-005); the invalid value does not break rendering. |
| EC-3 | A `data-i18n` key is absent from the active locale's dictionary | `applyI18n` leaves the element's existing (English baseline) text intact — it does NOT blank the element (R6). |
| EC-4 | JS disabled entirely | The page renders in the hardcoded English baseline (the `data-i18n` attributes are inert without JS); the form still works (plain HTML POST round-trip — no regression to the no-JS baseline 001 established). |
| EC-5 | A POST submitted while interface=zh | Server-submitted fields byte-identical to interface=en; langpick value absent from the POST body (AC-WC5-008a). |
| EC-6 | A ja string contains a glyph outside the subset (dictionary edited later without re-subsetting) | That single glyph falls back to `system-ui`; the subset MUST be regenerated whenever the dictionary changes (a build-time discipline noted in M1). For the shipped dictionary the subset covers 100% of glyphs. |
| EC-7 | Switching language while a server-rendered `.FieldErrors` is shown (post-POST reject) | The chrome around the error switches language; the server-rendered error message TEXT is unchanged (server content translation is out of scope — E.1); the validation authority is unaffected. |
| EC-8 | A non-default port console (e.g. `127.0.0.1:7777`) | i18n + font behavior is port-independent; the loopback indicator (004 `BindAddr`) is unaffected by 005. |
| EC-9 | `<html lang>` update + CJK font activation consistency | Switching to ja/zh updates `<html lang>` AND the CJK glyphs resolve to the subset — the two are wired consistently (R7); a lang switch never leaves ja/zh text in `system-ui` while `<html lang>` claims the CJK locale. |
| EC-10 | The design `i18n.js` is re-imported in a future refresh | The `rv.*` strip + key reconciliation MUST be re-applied; the shipped dictionary is a derivative, never a verbatim copy (AC-WC5-009). |

---

## §E — Definition of Done

- [ ] All 13 ACs PASS (the 3 must-pass — AC-008a interface≠content, AC-010a server-contract byte-unchanged, AC-011 offline — explicitly verified).
- [ ] `go test ./internal/web/... ./internal/cli/... ./internal/config/...` exit 0 (AC-WC5-013).
- [ ] `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` exit 0 (AC-WC5-011).
- [ ] Offline regression grep = 0 external font/style/script URL (AC-WC5-011).
- [ ] `git diff --exit-code internal/web/validate.go` → exit 0 (validators byte-unchanged, AC-WC5-010a).
- [ ] `i18n.js` ships with 0 `rv.*` keys (AC-WC5-009) + all 4 locales (AC-WC5-001).
- [ ] langpick is a non-form appbar widget (no `name=`, outside `<form>`) — the cohort core invariant (AC-WC5-008a/008b).
- [ ] CJK woff2 subset is tens of KB (subset, not full font) + OFL-1.1 license shipped (AC-WC5-006).
- [ ] The 004 `TestAppbarRendered` S3-exclusion guard inverted (langpick/`data-i18n` forbidden→expected) (AC-WC5-012).
- [ ] `go:embed` enumerates the new `i18n.js`; the `assets/fonts` glob covers the new CJK font (AC-WC5-011).
- [ ] No new settings field; content-language fieldset untouched; S1/S2a invariants not widened (AC-WC5-010b / E.2/E.3).

---

## §F — Forward-Looking Checks (cohort closure)

005 is the cohort terminator — there is no S4 successor SPEC queued at authoring time. Two forward-looking notes for any future refresh:

- **Dictionary ↔ subset coupling**: any future edit to `i18n.js` that adds ja/zh strings MUST re-run the `pyftsubset --text=` step against the updated glyph set, or new glyphs fall back to `system-ui` (EC-6). This coupling is a build-time discipline, not a runtime check.
- **Single-voice ja refinement (optional, NOT 005)**: if a future SPEC wants the ja chrome in the Pretendard voice (rather than Noto), Pretendard JP could be subset-added for ja specifically. This is a visual refinement on top of 005's offline-safe baseline, not a 005 requirement (plan.md §F.2 note).
- **Server-side content translation (explicitly NOT this cohort)**: localizing the server-rendered `.FieldErrors` message text or adding `Accept-Language` negotiation is a separate, larger effort outside the web-console-v3 cohort (E.1). 005 deliberately keeps i18n client-side.
