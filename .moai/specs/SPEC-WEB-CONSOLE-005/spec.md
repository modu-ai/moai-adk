---
id: SPEC-WEB-CONSOLE-005
title: "Web Console — Interface i18n (en/ko/ja/zh) + CJK self-host webfont coverage (zero server-contract change)"
version: "0.2.0"
status: implemented
created: 2026-06-03
updated: 2026-06-03
author: GOOS
priority: P1
phase: "v3.0.0"
module: "internal/web"
lifecycle: spec-anchored
tags: "web, console, i18n, cjk, webfont, pretendard, noto, offline, localStorage"
tier: M
related_specs: [SPEC-WEB-CONSOLE-001, SPEC-WEB-CONSOLE-002, SPEC-WEB-CONSOLE-003, SPEC-WEB-CONSOLE-004]
---

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-06-03 | GOOS | Initial draft — the FINAL member (005, cohort-internal label "S3") of the web-console-v3 cohort. Two parts: **E.1** web interface i18n (`data-i18n` wiring + client-side `i18n.js` dictionary for en/ko/ja/zh + appbar language picker + `localStorage("moai-console-lang")`), and **E.2** CJK self-host webfont coverage for the ja/zh interface-string glyphs that the 004-embedded Pretendard ko+Latin subset does NOT cover. MUST-PASS invariants: **zero server-contract change** (interface language ≠ content language; the langpick mutates NO server-submitted field), **offline / CDN-free** (all font assets via `go:embed` self-host woff2 subset — no Google-Fonts `@import`, no unpkg, no runtime network fetch), `rv.*` design-review keys EXCLUDED from the dictionary. Inverts 004's `TestAppbarRendered` S3-exclusion guard (langpick / `data-i18n` move from FORBIDDEN to EXPECTED). Cohort siblings 001 (모태) / 002 (S1) / 003 (S2a) / 004 (visual restyle) are ALL completed; 005 is the cohort terminator. |

---

## §1 Context & Motivation

### 1.1 Where this SPEC sits in the cohort

The `web-console-v3` cohort hardens and extends the loopback-only browser settings editor (`moai web`). 005 is the **final** member:

- **001** (모태, completed) — the original loopback-only, no-auth profile/project settings editor.
- **002 / S1** (completed) — port `3041` default + web↔TUI validation parity.
- **003 / S2a** (completed) — flat project-config parity (`development_mode` + `git_convention.convention`).
- **004** (completed) — the **visual layer**: applied the 모두의AI design system (token layer + console.css component port + new appbar + field chrome + dark mode + theme toggle + inline-SVG icon subset + self-hosted **Pretendard ko+Latin** woff2 subset, 5 weights). 004 deliberately deferred web interface i18n + the full CJK webfont to this SPEC (004 §4 E.1/E.2).
- **005 / THIS SPEC (S3)** — the **interface-i18n layer** + the **CJK webfont coverage** that i18n requires.

005 takes the brand look 004 established and makes the console's UI chrome (section legends, field titles, descriptions, banner text, button labels, the loopback subtitle) switchable across **en / ko / ja / zh** entirely client-side, and adds the self-hosted webfont glyph coverage needed so the ja/zh translated strings render in the brand typeface instead of falling back to `system-ui`.

### 1.2 Verified ground-truth — the post-004 implementation target

The console at `internal/web/assets/page.html.tmpl` (post-004) server-renders 5 `<fieldset>` groups + a new appbar + actions via `html/template`, embedded through `internal/web/assets.go:17` (`//go:embed assets/console.css assets/app.js assets/page.html.tmpl assets/fonts`). Verified anchors that 005 builds on or must preserve:

| Concern | Verified ground-truth (post-004) |
|---------|-----------------------------------|
| HTML lang attr | `page.html.tmpl:2` — `<html lang="en" data-theme="light">` (hardcoded `en`; 005 adds client-side lang switching) |
| FOUC theme pattern | `page.html.tmpl:8-19` — a `<head>` inline `<script>` reads `localStorage("moai-console-theme")` and applies `data-theme` before first paint. 005 mirrors this pattern for the persisted interface language |
| Appbar (no langpick) | `page.html.tmpl:24-46` — appbar has brand badge + `모두의AI` + loopback indicator + `#themeToggle`, but **NO** language picker. 004 deliberately excluded it (S3 scope) |
| Hardcoded English chrome | section legends (`Identity`/`Language`/…), field titles/descriptions, `pagehead__sub`, banner text, `Save settings`, `Custom segments` — all hardcoded English in the template; 005 tags them with `data-i18n` |
| Theme toggle JS pattern | `app.js:51-59` `wireThemeToggle()` + `applyTheme()` — the `localStorage`-persist + DOM-apply + DOMContentLoaded-wire pattern 005 mirrors for `applyI18n()` |
| `go:embed` directive | `assets.go:17` — `assets/console.css assets/app.js assets/page.html.tmpl assets/fonts`; 005 extends it for the new `i18n.js` + the CJK font subset |
| Pretendard subset (ko+Latin only) | `assets/fonts/Pretendard-{Regular,Medium,SemiBold,Bold,Black}.subset.woff2` (5 weights, ~18KB each, ~90KB total) + `OFL.txt`. `console.css:18-51` `@font-face` references them by relative `/static/fonts/…`. **These cover Latin + Hangul ONLY** — no hiragana/katakana (ja) or simplified hanzi (zh) |
| Canonical lang SSOT | `internal/web/validate.go:26` `langOptions = []string{"en","ko","ja","zh"}` — the 4 **content-language** values (server-validated). The interface langpick offers the same 4 codes but for UI chrome only |

### 1.3 The design input (already produced)

The same Claude Design session that produced 004's visual prototype also produced the i18n dictionary `i18n.js` (preserved at `.moai/design/web-console-handoff/from-claude-design/assets/i18n.js`), covering en/ko/ja/zh. Verified structure:

- `window.MOAI_I18N = { en: {...}, ko: {...}, ja: {...}, zh: {...} }` — a flat key→string map per locale.
- **Product keys** (wire these): `app.subtitle`, `profile.label`, `profile.selected`, `actions.save`, `actions.meta`, `seg.title`, `seg.note`, `count.*`, `banner.success.*`, `banner.error.*`, `err.user_name`, `sec.<section>.title|desc` (×5 sections), `f.<field>.title|desc` (×16 fields), `theme.aria`, `lang.aria`.
- **`rv.*` design-review keys** (EXCLUDE — `rv.title`/`rv.banner`/`rv.valid`/`rv.profile`/`rv.none`/`rv.success`/`rv.error`/`rv.ok`/`rv.err`/`rv.multi`/`rv.single`/`rv.hint`): the HTML comment in the design markup marks the `.review` "State preview" aside as "NOT part of the product." 004 already excluded the `.review` aside; 005 excludes its companion `rv.*` translation keys from the shipped dictionary.

> The shipped `internal/web/assets/i18n.js` is a **derivative** of this design `i18n.js` with the `rv.*` keys stripped and the key set reconciled to the actual rendered chrome. The design file is the visual SSOT; it is NOT copied verbatim.

### 1.4 Why interface i18n is DISTINCT from content language (the core invariant)

The console already has a **Language fieldset** with 4 server-validated content-language settings: `conversation_lang`, `git_commit_lang`, `code_comment_lang`, `doc_lang` (each `en/ko/ja/zh`, validated by `validatePrefs` against `langOptions`). These are **profile settings the server reads and persists** — they say "what language the assistant replies in / writes commits in / etc."

The 005 **interface language** (the appbar langpick) is a completely different axis: it controls **which language the console's own UI chrome is shown in** — a machine-local cosmetic preference, like the dark/light theme toggle 004 added. It is **NOT a server setting**: it does not appear in any `name=` POST field, it is not validated, it is not persisted to any profile or project config. It lives only in `localStorage("moai-console-lang")` and is applied entirely client-side by `applyI18n()`.

This separation is THE core invariant of 005 (the §3 must-pass): a user could set their content `conversation_lang` to `ko` while viewing the console chrome in `en`, or vice versa — the two are orthogonal. The langpick MUST NOT alter, submit, or interfere with any server-submitted field.

### 1.5 The CJK font gap that i18n exposes (E.2)

004 embedded ONLY the Pretendard **Latin + Hangul** woff2 subset (covers en + ko interface text). When 005's langpick switches the interface to **ja** or **zh**, the translated chrome strings contain:

- **ja**: hiragana (e.g. `設定を保存`→`を`), katakana (`ステータスライン`), and kanji (`設定`/`権限`) — **none covered** by the Hangul+Latin subset.
- **zh**: simplified hanzi (`保存设置`/`权限模式`) — **none covered**.

Without added coverage, those glyphs fall back to the OS `system-ui` font, breaking brand typography for the ja/zh interface (the very thing the 모두의AI design system exists to prevent). 005 adds **self-hosted woff2 subset coverage for exactly the ja/zh interface-string glyphs**.

Estimated glyph demand (the i18n dictionary is a FIXED, KNOWN string set — `pyftsubset --text=` can target it exactly): the ja+zh product dictionary (rv.* excluded) uses **on the order of ~290 unique CJK glyphs** (the union of the ja and zh dictionary string values, hiragana + katakana + kanji + simplified hanzi, with kanji/hanzi overlap). This number is an **estimate, not a fixed precision**: it is tokenizer-dependent (different CJK Unicode-range definitions and CJK-punctuation inclusion yield counts in the ~284–287 band) AND it is measured against the **design** `i18n.js`, whereas the **shipped** dictionary is a reconciled derivative (rv.* stripped, keys reconciled to the actual rendered chrome — §1.3/§1.6) so any pre-implementation count is provisional by construction.

[HARD] The run-phase MUST measure the actual **shipped**-dictionary glyph count (not the design file) and record it in progress.md. Reproduce via:

```bash
# extract the union of ja+zh string VALUES from the SHIPPED dictionary, feed to pyftsubset --text=
pyftsubset <NotoSansSC|NotoSansJP source> \
  --text="$(<extract all ja+zh value strings from internal/web/assets/i18n.js, rv.* already absent>)" \
  --flavor=woff2 --output-file=<weight>.subset.woff2
```

This tiny, bounded glyph set (a few hundred glyphs, tens of KB) makes exact-string subsetting overwhelmingly the most byte-efficient strategy (see plan.md §F font-strategy decision). ko adds no new glyphs (already covered by 004's Hangul subset). The subset target + AC-WC5-006 bind to the SHIPPED dictionary's glyph set, NOT this estimate.

### 1.6 Cohort scope-fence

005 is the cohort terminator. It does NOT widen the S1 (port + web↔TUI parity) or S2a (project-config scope = `development_mode` + `git_convention` only) invariants, and it does NOT touch the S2b track (deep nested config editing — a separate, non-cohort track). See §4.

---

## §2 GEARS Requirements

### REQ-WC5-001 (Ubiquitous — client-side i18n dictionary, 4 locales, offline)

The Console **shall** ship an embedded client-side i18n dictionary (`internal/web/assets/i18n.js`) covering the four interface locales **en / ko / ja / zh**, delivered via the web package's existing `go:embed` mechanism and served from `/static/`, with **no** server round-trip required to switch the interface language and **no** network fetch of the dictionary at runtime (it is part of the offline embed).

### REQ-WC5-002 (Ubiquitous — `data-i18n` attribute wiring on the page chrome)

The Console page (`page.html.tmpl`) **shall** carry `data-i18n` attributes on every translatable UI-chrome element (the `pagehead__sub` subtitle, the 5 section legends + descriptions, the field titles + descriptions, the segment group title + note, the banner text, the primary-button label, the profile label, and the `actions__meta` text), so the client `applyI18n()` routine can replace their text content from the active-locale dictionary. The **code-chip key identifiers** (`<code class="field__key">user_name</code>` etc.) **shall not** be tagged for translation — they remain English code tokens.

### REQ-WC5-003 (Ubiquitous — appbar language picker)

The Console appbar **shall** render an interface-language picker (`class="langpick"` `<select>` with `id="uiLangSelect"`) offering the four interface locales (en / ko / ja / zh), placed in the appbar brand strip (which 004 introduced) alongside the existing loopback indicator and theme toggle. The picker **shall** carry an `aria-label` (the `lang.aria` dictionary key) and **shall not** carry a `name=` POST attribute (it is not a form field — it is excluded from the `<form>` and submits nothing to the server).

> **Identifier-collision constraint (do NOT reuse `langSelect`)**: `langSelect` is ALREADY a live `{{define "langSelect"}}` **content-language** helper template (`page.html.tmpl:249`), invoked 4× for the content-language fieldset. The new appbar **interface**-language picker MUST use the non-colliding id `uiLangSelect` (keep `class="langpick"` — that class marker does NOT collide). Reusing `id="langSelect"` would overload an identifier that already means "content-language select," violating this SPEC's interface≠content invariant at the identifier level. See AC-WC5-012 / R4 for the test-inversion implication.

### REQ-WC5-004 (Event-driven — language selection applies + persists client-side)

**When** the user selects an interface language from the appbar `langpick`, the Console **shall** apply the chosen locale's translations to all `data-i18n` elements client-side (`applyI18n()`) and persist the selection in `localStorage("moai-console-lang")`, with **no** server round-trip, **no** form submission, and **no** mutation of any profile or project setting.

### REQ-WC5-005 (State-driven — persisted interface language applied on load)

**While** a persisted interface language exists in `localStorage("moai-console-lang")` (one of en / ko / ja / zh), the Console **shall** apply that locale's translations on page load — mirroring 004's FOUC-prevention theme pattern — so the interface renders in the persisted language from first interaction. **Where** no persisted value exists (first visit) **or** the persisted value is not one of the four valid locales, the Console **shall** default to the `en` interface dictionary (the template's hardcoded English text is the `en` baseline).

### REQ-WC5-006 (Ubiquitous — self-hosted CJK webfont subset, offline)

The Console **shall** serve CJK glyph coverage for the ja/zh interface strings as a **self-hosted woff2 subset** embedded via `go:embed` and referenced by relative `@font-face src` from the embedded console CSS, subsetted to **exactly** the glyph set used by the ja/zh interface dictionary (the fixed, known string set), with the corresponding **OFL-1.1 license file included** alongside the font assets. The Console **shall not** fetch any CJK font over the network at runtime (no Google-Fonts `@import`, no unpkg, no other CDN).

### REQ-WC5-007 (State-driven — brand typeface covers the active interface language)

**While** the interface language is `ja` or `zh`, the Console **shall** render the translated UI-chrome strings in the brand typeface via the self-hosted CJK subset `@font-face` (added to the `--font-sans` stack), such that ja hiragana/katakana/kanji and zh simplified hanzi render in the embedded webfont rather than falling back to `system-ui`. **While** the interface language is `en` or `ko`, the existing 004 Pretendard Latin+Hangul subset **shall** continue to cover the chrome (no regression to the en/ko base look).

### REQ-WC5-008 (Unwanted behavior — interface language MUST NOT alter server-submitted state)

The interface langpick **shall not** be a form field and **shall not** alter, submit, or interfere with any server-submitted value: the content-language settings (`conversation_lang` / `git_commit_lang` / `code_comment_lang` / `doc_lang`) and every other `name=` POST field **shall** be byte-identical in a POST round-trip regardless of the active interface language. The interface language **shall** be entirely client-side (`localStorage` only); the Console **shall not** add any profile field, project-config field, or server round-trip to persist the interface language.

### REQ-WC5-009 (Unwanted behavior — `rv.*` design-review keys excluded)

The shipped i18n dictionary **shall not** include the `rv.*` design-review keys (`rv.title` / `rv.banner` / `rv.valid` / `rv.profile` / `rv.none` / `rv.success` / `rv.error` / `rv.ok` / `rv.err` / `rv.multi` / `rv.single` / `rv.hint`); they belong to the `.review` "State preview" design-tool scaffold ("NOT part of the product") that 004 already excluded. The Console page **shall not** render any element bound to an `rv.*` key.

### REQ-WC5-010 (Ubiquitous — server contract preservation, zero behavior change)

The Console **shall** preserve every server contract unchanged through the i18n + font work: (a) every form input/select keeps its `name=` POST attribute and server-rendered `{{range}}` option list; (b) `.FieldErrors` per-field errors stay **server-side** rendered; (c) the `<form method="POST" action="/save…">` + hidden `__profile` + `{{if .ShowProfileSwitch}}` + `name="__profile_select"` are preserved; (d) `validatePrefs` / `validateProjectConfig` and the canonical lists in `validate.go` are **byte-unchanged** (005 touches templates + assets + at most a non-validating view-model wiring, never a validator); (e) the loopback-only bind, no-auth / no-token / no-session posture, and Host-header write-safety check are unchanged; (f) the S1 (port + web↔TUI parity) and S2a (project-config scope fence) invariants are not widened.

### REQ-WC5-011 (Ubiquitous — `go:embed` delivery + cross-platform build)

All new assets (the `i18n.js` dictionary, the CJK woff2 subset font file(s), the corresponding OFL-1.1 license, and the CJK `@font-face` additions to the console CSS) **shall** be delivered through the web package's `go:embed` directive (extending `internal/web/assets.go`), such that `go build ./...` and the cross-platform build (`GOOS=windows GOARCH=amd64 go build ./...`) succeed and the assets are served offline with no network fetch.

### REQ-WC5-012 (Ubiquitous — accessibility + 004 test-guard reconciliation)

The Console **shall** preserve accessibility through the i18n addition: the langpick carries an `aria-label`, switching the interface language updates the `<html lang>` attribute to the active locale (so assistive tech announces the correct language), and the existing 004 accessibility cues (focus-visible, prefers-reduced-motion, error-cue ARIA) are unchanged. The 004 `TestAppbarRendered` S3-exclusion guard (which currently asserts the rendered-body literals `class="langpick"`, `id="langSelect"`, and `data-i18n` are ABSENT) **shall** be reconciled because 005 intentionally lands the elements 004 forbade.

> **Inversion semantics (do NOT invert the wrong marker)**: 005 adds the rendered-body literals `class="langpick"`, `id="uiLangSelect"`, and `data-i18n` (the appbar picker uses id `uiLangSelect` per REQ-WC5-003, NOT `langSelect`). The run-phase reconciliation therefore inverts the guard to **EXPECT** `class="langpick"` + `data-i18n` + the NEW `id="uiLangSelect"`. The 004 guard's stale forbidden-string `id="langSelect"` was the WRONG handle for the new element — it referenced the originally-planned id, never landed. That forbidden entry MUST be either (a) removed, or (b) left in place harmlessly: `langSelect` only ever appears in the rendered body as a content-language `<select>`'s own id (`page.html.tmpl:254`, `id="{{.Name}}"` where `.Name` is `conversation_lang`/`git_commit_lang`/… — never the literal string `langSelect`), so a residual `id="langSelect"`-absent assertion stays satisfied and does not collide with the appbar picker. The run-phase MUST NOT invert the `id="langSelect"` forbidden entry into an EXPECT (that would assert a literal that 005 never renders).

---

## §3 Acceptance Criteria (summary — full enumeration in acceptance.md)

Each AC is independently verifiable. Because i18n + a font subset are hard to unit-test for pixels, the AC strategy leans on **structural assertions** (template renders the `data-i18n` markers + langpick; the dictionary embeds + parses; the font file embeds), **server-contract regression assertions** (POST round-trip byte-identical regardless of interface language; `validate.go` byte-unchanged), and **offline assertions** (zero external font/script URL in served assets). The closure gate is `go test ./internal/web/... ./internal/cli/... ./internal/config/...` exit 0. The full Given-When-Then enumeration with edge cases lives in `acceptance.md`; the table below is the SSOT index.

| AC | REQ | Assertion (one-line) |
|----|-----|----------------------|
| AC-WC5-001 | REQ-WC5-001 | `i18n.js` present under web assets + embedded via `go:embed` + served from `/static/`; defines `window.MOAI_I18N` with all 4 locale keys (en/ko/ja/zh); no network fetch of the dictionary |
| AC-WC5-002 | REQ-WC5-002 | Rendered page carries `data-i18n` attributes on the chrome elements (subtitle, 5 section legends+descs, field titles+descs, seg title+note, banner text, save button, profile label, actions meta); code-chip `<code class="field__key">` keys carry NO `data-i18n` |
| AC-WC5-003 | REQ-WC5-003 | Appbar renders the `class="langpick"` `<select>` with `id="uiLangSelect"` (NOT `langSelect`), the 4 locale options + `aria-label`; the langpick is OUTSIDE the `<form>` and carries NO `name=` attribute |
| AC-WC5-004 | REQ-WC5-004 | `app.js` (or a sibling) wires a `change` listener on `langpick` → `applyI18n(locale)` + `localStorage.setItem("moai-console-lang", locale)`; no form submit, no fetch |
| AC-WC5-005 | REQ-WC5-005 | Page-load init applies the persisted `moai-console-lang` locale; invalid/absent persisted value defaults to `en`; mirrors the 004 FOUC theme pattern (early `<head>` lang application or DOMContentLoaded apply) |
| AC-WC5-006 | REQ-WC5-006 | CJK woff2 subset font file(s) present under the web assets fonts tree + embedded by `go:embed`; OFL-1.1 license present; `@font-face src` is relative (no `https://`); subset covers exactly the glyph set of the **shipped** `internal/web/assets/i18n.js` ja+zh values (rv.* stripped) — verified by re-deriving the glyph set from the shipped dictionary at run-phase, NOT from the design file. The run-phase records the actual shipped-dictionary glyph count in progress.md (the §1.5 estimate is provisional — see §1.5) |
| AC-WC5-007 | REQ-WC5-007 | The CJK `@font-face` is added to the `--font-sans` stack (or a CJK-specific face) so ja/zh chrome renders in the webfont; the 004 Pretendard Latin+Hangul subset still covers en/ko (no en/ko regression) |
| AC-WC5-008a | REQ-WC5-008 | **MUST-PASS** — POST round-trip: a form submitted while interface=ja produces byte-identical server-submitted fields to interface=en; the langpick value is NOT in the POST body; content-language fields (`conversation_lang` etc.) are unaffected by the interface langpick |
| AC-WC5-008b | REQ-WC5-008 | langpick has NO `name=` attribute and is NOT inside `<form>`; no profile/project-config field for interface language exists; interface language persists ONLY in `localStorage` |
| AC-WC5-009 | REQ-WC5-009 | Shipped `i18n.js` contains NO `rv.*` keys (grep = 0); rendered page binds NO element to an `rv.*` key; no `.review` aside |
| AC-WC5-010a | REQ-WC5-010 | **MUST-PASS** — `validatePrefs`/`validateProjectConfig` + canonical lists in `validate.go` byte-unchanged (`git diff --exit-code internal/web/validate.go`); every form field retains its `name=` + `{{range}}` server-rendered option; `.FieldErrors` server-side render intact |
| AC-WC5-010b | REQ-WC5-010 | Existing 001/002/003/004 invariant tests stay green (loopback bind, no-auth, host-check, no-direct-marshal, validation parity, restyle assertions); web↔TUI parity (S1) + project-config scope (S2a) unchanged |
| AC-WC5-011 | REQ-WC5-011 | **MUST-PASS (offline)** — `go build ./...` exit 0 + `GOOS=windows GOARCH=amd64 go build ./...` exit 0; new `i18n.js` + CJK font enumerated in `go:embed`; served CSS/HTML/JS contains 0 external `https://` font/style/script URL (grep = 0) |
| AC-WC5-012 | REQ-WC5-012 | langpick carries `aria-label`; switching language updates `<html lang>`; 004 `TestAppbarRendered` forbidden-element guard inverted to EXPECT `class="langpick"` + `data-i18n` + the NEW `id="uiLangSelect"` (NOT `id="langSelect"` — that stale forbidden-string referenced the never-landed original id; do NOT invert it into an EXPECT) — the test reversal is in the diff |
| AC-WC5-013 | all | Closure gate: `go test ./internal/web/... ./internal/cli/... ./internal/config/...` exit 0 |

---

## §4 Exclusions (What NOT to Build)

### Out of Scope

The following are explicitly excluded or belong to other tracks; they MUST NOT be implemented in 005:

- **S2b track** — deep nested config editing (full quality / workflow / git-strategy / harness sections + nested git-convention / llm sub-fields). 005 changes NO settings fields, adds NO settings fields, removes NO settings fields. It adds interface i18n + font coverage only — it does NOT touch the server-known config surface.
- **Server-side content translation** — 005 does NOT translate any server-submitted value, does NOT localize validation error message *content* on the server (the `.FieldErrors` server-rendered text is unchanged; only the client `data-i18n` chrome around it is switchable), and does NOT add a server locale negotiation / `Accept-Language` mechanism.
- **Design-review `.review` aside + `rv.*` keys** — already excluded by 004; 005 reaffirms the `rv.*` dictionary keys are stripped (REQ-WC5-009).
- **New content-language values** — the content-language fieldset stays exactly `en/ko/ja/zh` (the `langOptions` canonical set). 005 adds NO new content-language option and does NOT alter `validate.go`.

[HARD] The following are explicitly **out of scope** for 005 and MUST NOT be implemented:

### E.1 Out of Scope — server-side i18n / content translation

005's i18n is **client-side interface chrome only**. It MUST NOT: translate server-submitted values; localize the server-rendered `.FieldErrors` message text; add `Accept-Language` negotiation; persist the interface language to any profile/project config; or add a server round-trip to switch the interface language. The interface language lives ONLY in `localStorage("moai-console-lang")` and is applied ONLY by client JS.

### E.2 Out of Scope — content-language fieldset changes (S2a invariant)

The Language fieldset's 4 content-language selects (`conversation_lang` / `git_commit_lang` / `code_comment_lang` / `doc_lang`) are server settings validated against `langOptions`. 005 MUST NOT add, remove, or re-scope any of them, MUST NOT change their option set, and MUST NOT touch `validate.go` `langOptions` / `validatePrefs` / `validateProjectConfig` (byte-unchanged). The interface langpick is a SEPARATE, non-form widget — it is NOT a 5th content-language setting.

### E.3 Out of Scope — deep nested config editing (S2b)

No new settings fields, no field-set changes, no nested config editing. 005 is i18n + font coverage only.

### E.4 Out of Scope — `rv.*` design-review keys + `.review` aside

The `rv.*` translation keys and the `.review` "State preview" design-review chrome are a design-tool scaffold ("NOT part of the product"), already excluded by 004. 005 MUST strip the `rv.*` keys from the shipped dictionary and MUST NOT render any `rv.*`-bound element.

### E.5 Out of Scope — additional anti-patterns

[HARD] The following are forbidden regardless of sibling-SPEC scope:

1. **No CDN font / dictionary / script fetch** — the CJK font, the `i18n.js` dictionary, and any script MUST be self-hosted via `go:embed`. The served CSS/HTML/JS MUST contain ZERO external `https://` font/style/script URL (loopback / zero-network invariant — a cohort lesson: offline CDN invalidation). No Google-Fonts `@import`, no unpkg, no jsdelivr, no cdnjs.
2. **No full unsubsetted CJK font ship** — the CJK font MUST be subsetted to exactly the **shipped** ja/zh interface dictionary glyph set (a low-hundreds glyph set, measured at run-phase against `internal/web/assets/i18n.js`). Shipping a full Noto Sans SC/JP (tens of thousands of glyphs, multi-MB) is forbidden binary bloat. The dictionary is a fixed known string set — `pyftsubset --text=` targets it exactly.
3. **No interface-language → server mutation** — the langpick MUST NOT be a `name=` form field, MUST NOT be inside `<form>`, MUST NOT submit anything, and MUST NOT alter any server-submitted value (REQ-WC5-008). Interface language ≠ content language is THE core invariant.
4. **No server-contract change** — `name=` attrs, `{{range}}` server-rendered options, `.FieldErrors` server-side render, form method/action/hidden-profile, loopback bind, no-auth + Host-check, the `langSelect`/`optSelect` helper structure, and the S1/S2a invariants are all preserved (REQ-WC5-010). `validate.go` byte-unchanged.
5. **No client validation as SSOT** — the server's `.FieldErrors` (atomic POST reject re-render) remains the validation source of truth. i18n switches the chrome text around errors, NOT the validation authority. `name=` POST attributes remain mandatory on all inputs/selects.
6. **No auth / token / session / non-loopback bind** — the no-auth loopback-only posture of 001 is invariant. 005 adds zero security surface.
7. **No `rv.*` key in the shipped dictionary** (E.4) — strip them; render no `rv.*`-bound element.
8. **No template mirroring / `make build`** — `internal/web/assets/*` is embedded via the web package's own `go:embed` (verified `assets.go` pattern), NOT a deployed asset under `internal/template/templates/`. No `make build` / embedded-mirror parity step applies. The new `i18n.js` + CJK font assets live under `internal/web/assets/` and are added to the web package's `go:embed` directive only.

---

## §5 References (verified ground-truth)

| Path | Role |
|------|------|
| `.moai/design/web-console-handoff/IMPLEMENTATION-HANDOFF.md` | PRIMARY SOURCE — §5.3 (i18n client-vs-server decision), §6 (cohort decomposition: S3 = i18n + fonts; S2b separate), §7 (font strategy + offline CDN invalidation discovery), §9 milestone M5 (i18n wiring deliverable). The run-phase implementer reads §5/§6/§7/§9 |
| `.moai/design/web-console-handoff/from-claude-design/assets/i18n.js` | design i18n dictionary (en/ko/ja/zh) — the derivative source for `internal/web/assets/i18n.js` (strip `rv.*`, reconcile keys; NOT copied verbatim) |
| `.moai/specs/SPEC-WEB-CONSOLE-004/spec.md` | sibling format reference + §4 E.1/E.2 (the deferral of i18n + CJK font to this SPEC) + REQ-WC4-009 server-contract preservation invariant 005 inherits |
| `internal/web/assets/page.html.tmpl` | implementation target — appbar (add langpick), chrome elements (add `data-i18n`), `<head>` (mirror FOUC pattern for lang); `name=`/`{{range}}`/`.FieldErrors`/helpers PRESERVE |
| `internal/web/assets/console.css` | current console CSS (501 lines, post-004) — add CJK `@font-face` + langpick styling; `--font-sans` stack extension; the 004 Pretendard Latin+Hangul `@font-face` (lines 18-51) PRESERVED |
| `internal/web/assets/app.js` | current console JS (post-004) — add `applyI18n()` + langpick `change` wiring + load-time apply, mirroring the `wireThemeToggle`/`applyTheme`/DOMContentLoaded pattern; theme + segment-visibility logic preserved |
| `internal/web/assets/i18n.js` (NEW) | the shipped dictionary (derivative of the design `i18n.js`, `rv.*` stripped) |
| `internal/web/assets/fonts/` | 004's Pretendard Latin+Hangul subset (5 woff2 + OFL.txt) PRESERVED; 005 ADDS the CJK woff2 subset + its OFL-1.1 license |
| `internal/web/assets.go:17` | `//go:embed assets/console.css assets/app.js assets/page.html.tmpl assets/fonts` — extend to enumerate `assets/i18n.js` (the `assets/fonts` glob already covers new font files) (REQ-WC5-011) |
| `internal/web/validate.go:26` | `langOptions = []string{"en","ko","ja","zh"}` + `validatePrefs` + `validateProjectConfig` — **byte-unchanged** by 005 (E.2 / E.5.4) |
| `internal/web/handlers.go` | `pageView` + `newPageView()` + `bindForm` — unchanged for server contract; 005 does NOT add a server-validated field (the interface language is client-only) |
| `internal/web/restyle_test.go:192-217` | 004 `TestAppbarRendered` — the S3-exclusion guard asserting `class="langpick"`/`id="langSelect"`/`data-i18n` are ABSENT. 005 INVERTS this guard (REQ-WC5-012 / AC-WC5-012) — the elements 004 forbade are now expected |
| `internal/web/integration_test.go` | 001/002/003 invariant + DO_NOT_TOUCH sentinels — MUST stay green (AC-WC5-010b) |
| Noto Sans SC/JP OR Pretendard JP | CJK woff2 subset source (OFL-1.1) — subsetted to the ja/zh dictionary glyph set (REQ-WC5-006); font-family choice is a plan.md §F decision |
