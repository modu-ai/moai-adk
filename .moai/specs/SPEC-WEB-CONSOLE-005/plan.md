# Implementation Plan — SPEC-WEB-CONSOLE-005

> Web interface i18n (en/ko/ja/zh client-side dictionary + appbar langpick + `data-i18n` wiring) + CJK self-host webfont coverage for the ja/zh interface glyphs. Tier **M**. Run-phase `cycle_type=tdd` per `development_mode: tdd`. Plan-phase authored in main checkout on branch `docs/glm-webtool-routing-m1-m5`. Cohort terminator (S3) of web-console-v3.

## §A — Context

### A.1 Position & tier

- **Cohort**: `web-console-v3` final member (005, cohort-internal label "S3"). Siblings 001 (모태) / 002 S1 / 003 S2a / 004 visual-restyle — **all completed**. 005 closes the cohort.
- **Tier M justification**: multi-file cross-asset change spanning template + CSS + two new/extended JS assets + a new font subset + a Go test reconciliation:
  - `page.html.tmpl` — appbar langpick + `data-i18n` attributes on the static chrome elements (≥25 per AC-WC5-002 floor; the ~56 dictionary keys do NOT map 1:1 to DOM elements — `aria` keys feed attributes, banner/error strings are injected/server-rendered) + `<head>` lang-FOUC snippet + `<html lang>` wiring.
  - `assets/i18n.js` (NEW) — the 4-locale dictionary (derivative of the design `i18n.js`, `rv.*` stripped).
  - `assets/app.js` — `applyI18n()` + langpick `change` wiring + load-time apply (mirroring the theme-toggle pattern).
  - `assets/console.css` — CJK `@font-face` block + `--font-sans` stack extension + `langpick` styling.
  - `assets/fonts/` — NEW CJK woff2 subset (~290-glyph (shipped-dictionary measured, low-hundreds) subset) + its OFL-1.1 license.
  - `assets.go` — extend `go:embed` to enumerate `i18n.js`.
  - `restyle_test.go` — INVERT the 004 `TestAppbarRendered` S3-exclusion guard (langpick/`data-i18n` forbidden→expected) + NEW i18n/font tests.
  - LOC is moderate (mostly template/CSS/JS/dictionary + a font asset + test edits; little or no production Go change — the interface language is client-only). The cross-asset coordination (a new client-i18n surface + a new offline font pipeline + a server-contract regression surface + a sibling test reconciliation) exceeds Tier S's "<5 files / <300 LOC" envelope. Not Tier L: no constitutional change, no new persistence model, no nested-config redesign, < ~15 files.
- **Artifact set (Tier M)**: spec.md + plan.md + acceptance.md + progress.md (this is the 3-artifact + progress set; design.md/research.md NOT required at Tier M — the font-strategy research is carried inline in §F of this plan, grounded by the §1.5 glyph measurement in spec.md).

### A.2 SPEC artifacts

- `.moai/specs/SPEC-WEB-CONSOLE-005/spec.md` — 12 GEARS REQs (REQ-WC5-001..012) + 13 AC index + §4 Exclusions (E.1–E.5).
- `.moai/specs/SPEC-WEB-CONSOLE-005/plan.md` — this file.
- `.moai/specs/SPEC-WEB-CONSOLE-005/acceptance.md` — full Given-When-Then AC enumeration.
- `.moai/specs/SPEC-WEB-CONSOLE-005/progress.md` — plan-complete signal + run-phase evidence (created at plan close).

### A.3 PRIMARY SOURCE

`.moai/design/web-console-handoff/IMPLEMENTATION-HANDOFF.md` is the research + design basis. The run-phase implementer reads §5.3 (i18n client-vs-server decision), §6 (cohort decomposition — S3 = i18n + fonts; S2b is a SEPARATE track NOT in scope), §7 (font strategy + the offline-CDN-invalidation discovery), and §9 milestone M5 (i18n wiring deliverable). The design `i18n.js` under `from-claude-design/assets/` is the dictionary source (derivative — strip `rv.*`, reconcile keys to the actual rendered chrome).

### A.4 Existing infrastructure — PRESERVE vs EXTEND

| Item | Action | Detail |
|------|--------|--------|
| `internal/web/validate.go` (`langOptions` + validators) | PRESERVE (byte-unchanged) | 005 is i18n + font only; the interface language is client-only and touches NO validator. `langOptions` stays the 4 **content-language** values |
| `internal/web/handlers.go` (`newPageView` / `bindForm`) | PRESERVE | the interface language is NOT a view-model field and NOT a POST field; no server-side change needed |
| `internal/web/server.go` / `app.go` (bind + middleware) | PRESERVE | no-auth loopback posture + `/static/` serving unchanged |
| `internal/web/assets/fonts/Pretendard-*.subset.woff2` (004) | PRESERVE | the 5 Latin+Hangul weights cover en/ko chrome; 005 ADDS CJK coverage, does not replace them |
| `internal/web/assets/console.css` `@font-face` (004, lines 18-51) | PRESERVE + EXTEND | keep the Pretendard subset faces; ADD the CJK `@font-face` + extend `--font-sans` |
| `internal/web/assets/page.html.tmpl` | EXTEND | appbar langpick + `data-i18n` attrs + `<head>` lang FOUC + `<html lang>`; PRESERVE `name=`/`{{range}}`/`.FieldErrors`/helpers/form method/action/hidden-profile |
| `internal/web/assets/app.js` | EXTEND | `applyI18n()` + langpick wiring + load-time apply; PRESERVE theme + segment-visibility logic |
| `internal/web/assets.go` `go:embed` | EXTEND | add `assets/i18n.js` (the `assets/fonts` glob already covers new font files) |
| `internal/web/assets/i18n.js` (NEW) | CREATE | derivative dictionary (rv.* stripped) |
| `internal/web/assets/fonts/` CJK subset (NEW) | CREATE | low-hundreds-glyph woff2 subset (shipped-dictionary measured) + OFL-1.1 license |
| `internal/web/restyle_test.go:192-217` (`TestAppbarRendered`) | RECONCILE | invert the S3-exclusion guard — `class="langpick"`/`data-i18n` move from FORBIDDEN to EXPECTED + add `id="uiLangSelect"` EXPECT; do NOT invert the stale `id="langSelect"` forbidden-string (never-landed id — see R4) |

### A.5 PRESERVE list (do NOT modify outside SPEC scope)

- `internal/web/validate.go` — canonical lists + `validatePrefs` + `validateProjectConfig` (byte-unchanged — `git diff --exit-code` at run-phase end).
- `internal/web/server.go` bind/host logic; `internal/web/app.go` middleware/auth posture; `internal/web/handlers.go` server-contract surface (`bindForm`, `newPageView` POST handling).
- The 004 Pretendard Latin+Hangul `@font-face` block + the 5 woff2 subset files (extend, do not replace).
- The S1/S2a form field set (no settings field added/removed/re-scoped).
- All parallel-session in-flight files; unrelated modified working-tree files (`.claude/settings.json`, `.moai/config/sections/{statusline,user}.yaml`, the deleted `SPEC-V3R5-WORKFLOW-SCHEMA-EXTEND-001/progress.md`); `.moai/design/web-console-handoff/` (leave the design input untracked — read-only source); `.moai/docs/harness-delivery-strategy.md`.

---

## §B — Known Issues (auto-injected for run-phase delegation)

Filtered to the categories relevant to this `internal/web` i18n + font SPEC:

- **B3 (Subagent Boundary)** — `internal/web` is product code; no `AskUserQuestion` in the web layer. The implementer MUST NOT add user-interaction surfaces.
- **B4 (Frontmatter schema)** — N/A to run-phase (spec.md frontmatter already canonical: `created`/`updated`/`tags`).
- **B5 (CI 3-tier)** — spec-lint, golangci-lint, Test(per OS) fail independently. The i18n addition is mostly template/CSS/JS/dictionary + a font asset; if any Go change is introduced (it should not be needed), it must compile clean and cross-platform.
- **B8 (Working tree hygiene)** — `git add` ONLY `internal/web/**` run-phase changes; do NOT stage the parallel-session files, the design handoff dir, or unrelated modified config. Runtime-managed files untouched.
- **B10 (Untouched paths PRESERVE)** — §A.5 PRESERVE list. Parallel `manager-develop`/sibling-SPEC instances may be in flight on other paths; stay inside `internal/web/`.
- **B11 (AskUserQuestion forbidden)** — subagent returns a structured blocker report on any blocker; never prompts the user.
- **B1 (cross-platform build)** — `embed` uses forward-slash paths (cross-platform-safe). The font/dictionary assets are static bytes — platform-neutral by construction. If any Go change is added, the run-phase self-verification MUST include `GOOS=windows GOARCH=amd64 go build ./...` (REQ-WC5-011 / AC-WC5-011).

### B-extra: i18n/font-specific risks

- **R1 — Server-contract leak via langpick (HIGHEST)**: the single most dangerous failure mode. If the langpick is placed INSIDE `<form>` or given a `name=`, it submits to the server and either pollutes the POST body or collides with a field. AC-WC5-008a is the must-pass guard: a POST round-trip with interface=ja must be byte-identical to interface=en, and the langpick value must NOT appear in the POST body. The langpick MUST be a non-form widget in the appbar (the appbar is OUTSIDE `<form>` — verify the form boundary in `page.html.tmpl`).
- **R2 — Offline CDN regression**: the cohort's defining lesson (offline CDN invalidation). The CJK font + `i18n.js` + any script MUST be self-hosted via `go:embed`. AC-WC5-011 greps served CSS/HTML/JS for `fonts.googleapis.com` / `unpkg.com` / `jsdelivr` / `cdnjs` / any external font/style/script `https://` URL = 0 matches. A single leaked CDN reference fails the zero-network invariant.
- **R3 — CJK font bloat**: shipping a full Noto Sans SC/JP (multi-MB) instead of a low-hundreds-of-glyphs subset is forbidden binary bloat. The subset MUST be generated from the FIXED **shipped**-dictionary ja/zh string set via `pyftsubset --text=<exact glyphs from internal/web/assets/i18n.js>`. Keep the embedded CJK footprint to tens of KB.
- **R4 — 004 test guard contradiction (+ id-collision)**: `restyle_test.go::TestAppbarRendered` (lines 207-216) currently asserts the rendered-body literals `class="langpick"`, `id="langSelect"`, and `data-i18n` are ABSENT (the 004 S3-exclusion guard). 005 INTENTIONALLY lands `class="langpick"` + `data-i18n` + the NEW `id="uiLangSelect"` → the `class="langpick"`/`data-i18n` assertions WILL FAIL unless inverted. The run-phase reconciliation inverts the guard to **EXPECT** `class="langpick"` + `data-i18n` + `id="uiLangSelect"`. CRITICAL: do NOT invert the stale `id="langSelect"` forbidden-string into an EXPECT — it referenced the never-landed original id (the appbar picker uses `uiLangSelect`, NOT `langSelect`, because `langSelect` is already the live `{{define "langSelect"}}` content-language helper at `page.html.tmpl:249`). The `id="langSelect"` forbidden entry MUST be removed or left harmlessly absent-asserting (the literal `id="langSelect"` is never rendered — the content-language selects render `id="conversation_lang"` etc. via `id="{{.Name}}"`). It is NOT a regression — it is the planned guard reversal (AC-WC5-012). The TDD RED for the i18n appbar test must account for this existing-test contradiction (a plan-auditor-style trap: an AC that inverts an existing passing test).
- **R5 — `rv.*` key leak**: the design `i18n.js` includes `rv.*` design-review keys; copying it verbatim ships dead keys for an excluded scaffold. The shipped `i18n.js` MUST strip them (AC-WC5-009 greps the shipped dictionary for `rv.` = 0).
- **R6 — `data-i18n` key/render drift**: a `data-i18n="sec.identity.title"` attribute whose key is absent from the dictionary leaves an element untranslated (or blanks it if `applyI18n` overwrites with empty). The dictionary key set MUST exactly cover the `data-i18n` attribute set used in the template (boundary verification — cross-read both sides). `applyI18n` MUST leave the element's existing (English baseline) text intact when a key is missing rather than blanking it.
- **R7 — `<html lang>` + font-stack interaction**: switching interface to ja/zh should update `<html lang>` (a11y) AND ensure the CJK `@font-face` is in the active `--font-sans` stack. If the CJK face is only conditionally applied via a `[lang="ja"]`/`[lang="zh"]` selector, the `<html lang>` update is what activates it — the two must be wired consistently (R7 is why REQ-WC5-012 ties the lang attribute update to the font activation).

---

## §C — Pre-flight Checklist (run-phase, before any code change)

```bash
# 1. Branch + baseline
git branch --show-current          # docs/glm-webtool-routing-m1-m5 (or run-phase feat branch)
git rev-parse HEAD

# 2. Cross-platform build baseline (must already be green)
go build ./...
GOOS=windows GOARCH=amd64 go build ./...

# 3. Web package test baseline (NEW vs pre-existing distinction)
go test ./internal/web/... 2>&1 | tail -10

# 4. Lint baseline
golangci-lint run ./internal/web/... --timeout=2m 2>&1 | tail -5

# 5. Read the primary source sections FULLY + the design dictionary
#    .moai/design/web-console-handoff/IMPLEMENTATION-HANDOFF.md §5.3 / §6 / §7 / §9
#    from-claude-design/assets/i18n.js  (the dictionary source; note rv.* keys to strip)

# 6. Confirm the validators are the byte-unchanged baseline + the 004 font subset present
git diff --stat internal/web/validate.go              # expect NO change at run-phase end
ls internal/web/assets/fonts/                          # 5 Pretendard subset woff2 + OFL.txt (004)

# 7. Confirm the 004 test guard that must be inverted
grep -n 'langpick\|data-i18n' internal/web/restyle_test.go   # TestAppbarRendered forbidden block

# 8. Confirm the form boundary (langpick must be OUTSIDE <form>)
grep -n '<form\|</form>\|class="appbar"' internal/web/assets/page.html.tmpl
```

---

## §D — Constraints (DO NOT VIOLATE)

- **PRESERVE** the §A.5 list verbatim. `internal/web/validate.go` byte-unchanged.
- **Interface language ≠ content language** (REQ-WC5-008): the langpick is NOT a form field, has NO `name=`, is NOT inside `<form>`, submits nothing, and mutates no server-submitted value. Interface language lives ONLY in `localStorage("moai-console-lang")`.
- **Server contract** (REQ-WC5-010): `name=` on every input/select; `{{range}}` server-rendered options; `.FieldErrors` server-side render; `form method/action` + hidden `__profile`; `langSelect`/`optSelect` helper structure; loopback bind; no-auth + Host-check — all preserved.
- **Zero network** (REQ-WC5-006/011): no Google-Fonts `@import`, no unpkg, no jsdelivr/cdnjs, no external font/style/script `https://` URL in served assets. CJK font = self-host woff2 subset via `go:embed`; `i18n.js` = self-host via `go:embed`.
- **CJK subset, not full font** (REQ-WC5-006): subset to exactly the **shipped** ja/zh dictionary glyph set (a few hundred glyphs) via `pyftsubset --text=` against `internal/web/assets/i18n.js`. No multi-MB full CJK font.
- **`rv.*` excluded** (REQ-WC5-009): strip the design-review keys from the shipped dictionary; render no `rv.*`-bound element.
- **No new settings field** (E.2/E.3): i18n + font only — no settings field added/removed/re-scoped; the content-language fieldset is untouched.
- **Licensing**: the CJK font's OFL-1.1 license file shipped alongside the subset; the 004 Pretendard OFL.txt preserved.
- **Test reconciliation** (REQ-WC5-012): invert the 004 `TestAppbarRendered` S3-exclusion guard (langpick/`data-i18n` forbidden→expected) — committed in the same change as the appbar langpick.
- **Git**: Conventional Commits (`feat(SPEC-WEB-CONSOLE-005): M{N} …`); `Authored-By-Agent: manager-develop` + `🗿 MoAI` trailer; `git add` specific `internal/web/**` paths only; NO `--no-verify`, NO `--amend`, NO force-push to main.
- **No template mirroring / make build** — web assets are package-local `go:embed`, not `internal/template/templates/` deployed assets.

---

## §E — Self-Verification (run-phase deliverables)

The run-phase implementer reports the AC PASS/FAIL matrix (acceptance.md SSOT) + the following verification batch (parallel single-turn):

```bash
# Functional / build
go test ./internal/web/... ./internal/cli/... ./internal/config/...   # AC-WC5-013 closure gate
go build ./...                                                        # AC-WC5-011
GOOS=windows GOARCH=amd64 go build ./...                              # AC-WC5-011

# Offline / zero-network regression (AC-WC5-011) — expect 0 matches
grep -rn 'fonts.googleapis.com\|unpkg.com\|cdnjs\|jsdelivr\|https://fonts\|@import url("http' internal/web/assets/

# rv.* design-review keys excluded from the shipped dictionary (AC-WC5-009) — expect 0
grep -c 'rv\.' internal/web/assets/i18n.js

# langpick is NOT a form field (AC-WC5-008b) — the langpick line must carry NO name= and sit before <form>
grep -n 'langpick' internal/web/assets/page.html.tmpl
grep -n '<form\|</form>' internal/web/assets/page.html.tmpl   # langpick must precede <form> (in appbar)

# data-i18n wiring present (AC-WC5-002) — expect chrome elements tagged
grep -c 'data-i18n' internal/web/assets/page.html.tmpl

# i18n.js defines all 4 locales (AC-WC5-001)
grep -c 'en:\|ko:\|ja:\|zh:' internal/web/assets/i18n.js

# Validators byte-unchanged (AC-WC5-010a / E.2)
git diff --exit-code internal/web/validate.go && echo "validate.go unchanged"

# go:embed enumerates the new dictionary (AC-WC5-011)
grep -n 'go:embed' internal/web/assets.go

# CJK font + license present + subset (AC-WC5-006)
ls internal/web/assets/fonts/   # 004 Pretendard 5× + OFL.txt + NEW CJK subset woff2 + its OFL license
```

E-deliverables: AC binary matrix, cross-platform build result, web/cli/config test result, the offline + rv.* + langpick-not-form + data-i18n + locale-count grep outputs, `validate.go` unchanged confirmation, the CJK subset byte size, commit SHAs + push status, blocker report (if any).

---

## §F — Font Strategy Research & Proposal (the §E.2 plan-phase decision)

The CJK webfont strategy is the one genuine research decision in this SPEC. The handoff §7 invalidated the original "CDN" S3 plan (offline-unsafe); below is the trade-off analysis grounding the recommended option. The decisive input is the **bounded glyph demand** (spec.md §1.5): the ja+zh product dictionary (rv.* excluded) uses **on the order of ~290 unique CJK glyphs** (the union of ja and zh dictionary string values — hiragana + katakana + kanji + simplified hanzi, with kanji/hanzi overlap). This count is an **estimate, not fixed precision**: it is tokenizer-dependent (different CJK Unicode-range / CJK-punctuation definitions land it in the ~284–287 band) and it is measured against the **design** file, whereas the **shipped** dictionary is a reconciled derivative — so the run-phase MUST re-measure against `internal/web/assets/i18n.js` (the shipped dictionary) and record the actual count in progress.md. The strategy does not depend on the exact number: the set is a FIXED, KNOWN string set in the low hundreds of glyphs, and `pyftsubset --text=` targets the shipped dictionary's exact glyph set.

### F.1 Options evaluated

| Option | Family / coverage | Binary size (5 weights or 1) | Build-pipeline complexity | Visual consistency | License |
|--------|-------------------|------------------------------|---------------------------|--------------------|---------|
| **(a) Extend Pretendard subset to ja/zh glyphs** | Pretendard (single family). Pretendard's base release covers Latin+Hangul; **Pretendard JP** exists and covers hiragana/katakana/kanji, but Pretendard has **no full simplified-SC** coverage — zh hanzi would be missing or incomplete | Small if subset to the few-hundred shipped-dictionary glyphs (~tens of KB), but needs Pretendard JP for ja AND a separate SC source for zh → not truly "single family" | High for en/ko/ja (one type voice); **fails for zh** (no SC family) | Pretendard OFL-1.1 |
| **(b) Add Noto Sans JP + Noto Sans SC (full, unsubsetted)** | Two Noto families, full coverage | **MULTI-MB each** (~6-9MB unsubsetted per family) — forbidden bloat (R3) | Low (drop-in fonts) but **huge binary** | Medium (Noto ≠ Pretendard voice; mixing Pretendard for Latin/Hangul + Noto for CJK) | Noto OFL-1.1 |
| **(c) Glyph-subset to EXACTLY the i18n dictionary string set** (applied to whichever family) | Subset of Noto Sans SC (zh) + Noto Sans JP (ja) — OR a single Noto Sans CJK — restricted to the **shipped**-dictionary glyph set (a few hundred glyphs) | **~tens of KB total** (few-hundred glyphs × used weights; far smaller than the 90KB Pretendard ko subset 004 already ships) | Moderate — `pyftsubset --text=<exact shipped ja+zh dictionary glyphs>` per weight; same `pyftsubset` toolchain 004 already used for the Pretendard subset | Medium (Noto CJK voice for ja/zh chrome alongside Pretendard for en/ko) — acceptable, and far better than the `system-ui` fallback it replaces | Noto OFL-1.1 |

### F.2 Recommendation — Option (c): glyph-subset to exactly the dictionary string set

**Recommended: Option (c).** Rationale:

1. **Byte efficiency (decisive)**: because the ja/zh interface dictionary is a fixed, known glyph set in the low hundreds, subsetting to exactly those glyphs yields a CJK font footprint of tens of KB — comparable to or smaller than the ~90KB Pretendard ko subset 004 already ships. Option (b)'s full Noto families would add multiple MB and violate R3 / E.5.2. This is the single most important factor: a full CJK font (tens of thousands of glyphs) is almost entirely wasted bytes for a UI that only ever renders a few hundred distinct CJK characters. The exact subset count is measured at run-phase against the shipped dictionary (the strategy does not depend on the precise number).
2. **Offline-safe (satisfies the cohort invariant)**: self-hosted via `go:embed`, zero network — satisfies REQ-WC5-006 and the offline-CDN-invalidation lesson that killed the S3 "CDN" original plan.
3. **Toolchain already proven**: `pyftsubset` (fonttools) is the exact toolchain 004 used to produce the Pretendard subset — no new build dependency. The subset target is `--text=` fed the exact ja+zh dictionary characters (extractable from `i18n.js` programmatically, as the §1.5 measurement script demonstrates).
4. **Family choice within (c)**: use **Noto Sans SC** (zh) + **Noto Sans JP** (ja), each subsetted to its locale's glyph subset, OR a single **Noto Sans CJK** subsetted to the union — the run-phase implementer's call based on which yields the smaller embedded total. (Option (a)'s "single Pretendard family" is rejected because Pretendard has no SC coverage for zh — it cannot cover the zh hanzi, so it cannot be the single-family answer.) The CJK face is added to the `--font-sans` stack after Pretendard, so en/ko continue to resolve to Pretendard (no regression — REQ-WC5-007) and only the ja/zh glyphs (absent from Pretendard) fall through to the Noto CJK subset.
5. **Visual trade-off accepted**: a Noto CJK voice for ja/zh chrome sits alongside the Pretendard voice for en/ko. This minor mixed-family inconsistency is far preferable to the `system-ui` fallback (which varies per OS and breaks brand typography entirely). If a future SPEC wants a single-voice ja look, Pretendard JP could be subset-added for ja specifically — but that is a refinement, not a 005 requirement.

### F.3 Run-phase font deliverable (M3 below)

- Generate the CJK subset: `pyftsubset <NotoSansSC|NotoSansJP source> --text="<exact ja+zh dictionary glyphs>" --flavor=woff2 --output-file=…` for each needed weight (match the weights actually used by the chrome — likely Regular/Medium/SemiBold, not all 9).
- Ship under `internal/web/assets/fonts/` (the `assets/fonts` `go:embed` glob already covers it) + the Noto OFL-1.1 license file.
- Add the CJK `@font-face` to `console.css` (relative `/static/fonts/…` src) + extend `--font-sans` so ja/zh glyphs resolve to the subset.

---

## §G — Milestones (priority-ordered; no time estimates)

Per the handoff §9 step plan (M5 = i18n wiring) + §7 (font). Each milestone names its preservation verification. `cycle_type=tdd` — for the testable surface (template render assertions for `data-i18n`/langpick, the dictionary embed+parse, the font embed, the POST round-trip server-contract regression), write the failing test first, then implement.

### M1 — CJK font subset layer (offline-safe foundation)

- Generate the CJK woff2 subset (Option (c) — `pyftsubset --text=<ja+zh dictionary glyphs>`) for the used weights from a Noto Sans SC/JP (OFL-1.1) source. Ship under `internal/web/assets/fonts/` + the Noto OFL-1.1 license.
- Add the CJK `@font-face` block to `console.css` (relative src) + extend `--font-sans` so ja/zh glyphs resolve to the subset. The 004 Pretendard Latin+Hangul `@font-face` PRESERVED (en/ko unchanged).
- The `assets/fonts` `go:embed` glob already covers new font files (verify).
- **Preservation check**: `go build ./...` + `GOOS=windows … build` green; served CSS has 0 external font/style URL (AC-WC5-011); CJK subset is tens of KB (AC-WC5-006); en/ko still resolve to Pretendard (AC-WC5-007).
- **REQ coverage**: REQ-WC5-006, REQ-WC5-007, (partial) REQ-WC5-011.

### M2 — i18n dictionary + `data-i18n` wiring (the chrome-translation layer)

- Create `internal/web/assets/i18n.js` as a derivative of the design `i18n.js` with the `rv.*` keys STRIPPED and the key set reconciled to the actual rendered chrome (REQ-WC5-001 / REQ-WC5-009). Extend the `assets.go` `go:embed` directive to enumerate `assets/i18n.js`.
- Tag the page chrome elements with `data-i18n` attributes (subtitle, 5 section legends+descs, field titles+descs, seg title+note, banner text, save button, profile label, actions meta) — NOT the `<code class="field__key">` code chips (REQ-WC5-002). Boundary-verify: the `data-i18n` key set used in the template exactly matches the dictionary key set (R6).
- **Preservation check**: `i18n.js` embeds + defines all 4 locales (AC-WC5-001); `data-i18n` markers render (AC-WC5-002); shipped dictionary has 0 `rv.` keys (AC-WC5-009); NO server-contract change yet — `name=`/`{{range}}`/`.FieldErrors` intact; `validate.go` byte-unchanged.
- **REQ coverage**: REQ-WC5-001, REQ-WC5-002, REQ-WC5-009.

### M3 — Appbar langpick + client apply/persist (server-contract gate)

- Add the `langpick` `<select>` (4 locales + `aria-label` from `lang.aria`) to the appbar brand strip — **OUTSIDE `<form>`, NO `name=`** (R1). Verify the form boundary: the appbar precedes `<form>` in `page.html.tmpl`.
- Wire `applyI18n(locale)` + the langpick `change` listener + `localStorage("moai-console-lang")` persist + load-time apply in `app.js`, mirroring the `wireThemeToggle`/`applyTheme`/DOMContentLoaded pattern. Add the `<head>` FOUC-style early lang-apply (or DOMContentLoaded apply) for the persisted locale; default to `en` when absent/invalid (REQ-WC5-004/005). Update `<html lang>` on switch (REQ-WC5-012). `applyI18n` leaves English baseline text intact for missing keys (R6).
- **Preservation check (the critical gate)**: **MUST-PASS** POST round-trip byte-identical regardless of interface language (AC-WC5-008a); langpick has NO `name=` + is NOT in `<form>` (AC-WC5-008b); `validatePrefs`/`validateProjectConfig` byte-unchanged (AC-WC5-010a); 001/002/003/004 invariant tests green (AC-WC5-010b); content-language fields unaffected.
- **REQ coverage**: REQ-WC5-003, REQ-WC5-004, REQ-WC5-005, REQ-WC5-008, REQ-WC5-010, REQ-WC5-012 (langpick a11y + `<html lang>`).

### M4 — Test reconciliation + accessibility + closure verification

- INVERT the 004 `restyle_test.go::TestAppbarRendered` S3-exclusion guard: move `class="langpick"` + `data-i18n` from the FORBIDDEN block to the EXPECTED block AND add an EXPECT for the NEW `id="uiLangSelect"` (REQ-WC5-012 / AC-WC5-012). Do NOT invert the stale `id="langSelect"` forbidden-string — it referenced the never-landed original id (the appbar picker uses `uiLangSelect`; `langSelect` is the live content-language helper template) — remove it or leave it harmlessly absent-asserting (see R4). Add NEW tests: dictionary embed+parse, `data-i18n` coverage vs dictionary keys, langpick-not-a-form-field, POST round-trip language-invariance, CJK font embed, `rv.*` exclusion, `<html lang>` update.
- Verify accessibility: langpick `aria-label`; `<html lang>` updates on switch; 004 a11y cues (focus-visible, reduced-motion, error ARIA) unchanged.
- Run the full closure gate `go test ./internal/web/... ./internal/cli/... ./internal/config/...` + the §E regression batch.
- **REQ coverage**: REQ-WC5-012 (test reversal), closure (AC-WC5-013), final REQ-WC5-011.

### Milestone ordering rationale

M1 (CJK font) is the prerequisite for ja/zh chrome to render in-brand once i18n switches. M2 (dictionary + `data-i18n`) wires the translation targets without yet touching the server-contract surface. M3 is the server-contract preservation gate — the langpick MUST NOT leak into the form (highest risk), so it is isolated as its own milestone and gates closure. M4 reconciles the sibling 004 test guard and verifies the end state. M3 is the gate; if the langpick contaminates the POST body, the cohort's core invariant (interface ≠ content language) breaks.

---

## §H — Anti-Patterns (do NOT do)

- Placing the langpick INSIDE `<form>` or giving it a `name=` (leaks the interface language into the POST body → server-contract violation, the cohort's core invariant break). Keep it a non-form appbar widget.
- Copying the design `i18n.js` verbatim (ships the `rv.*` design-review keys for an excluded scaffold). Strip `rv.*`.
- Shipping a full Noto Sans SC/JP (multi-MB) instead of the ~290-glyph (shipped-dictionary measured, low-hundreds) subset (binary bloat). Subset to the fixed dictionary glyph set via `pyftsubset --text=`.
- Leaving any Google-Fonts `@import` / unpkg / jsdelivr / cdnjs reference (breaks offline). Self-host the font + dictionary via `go:embed`.
- Translating server-submitted values or the server-rendered `.FieldErrors` message text (server-side content translation is out of scope — E.1). i18n switches the CHROME around errors, not the validation authority.
- Touching `validate.go` / `langOptions` / adding a 5th content-language setting (the interface langpick is NOT a content-language field — E.2).
- Adding a profile/project-config field or server round-trip to persist the interface language (it is client `localStorage` only — REQ-WC5-008).
- Leaving the 004 `TestAppbarRendered` S3-exclusion guard intact (it will fail once the langpick lands). Invert it as part of the same change.
- `applyI18n` blanking an element when its `data-i18n` key is missing from the dictionary (leaves UI empty). Leave the English baseline text intact for missing keys.
- Mirroring assets into `internal/template/templates/` or running `make build` (web assets are package-local `go:embed`).

---

## §I — Cross-References

- `.moai/design/web-console-handoff/IMPLEMENTATION-HANDOFF.md` — PRIMARY design + research basis (§5.3 i18n decision, §6 cohort decomposition, §7 font strategy + offline CDN discovery, §9 M5).
- `.moai/specs/SPEC-WEB-CONSOLE-004/` — visual-restyle sibling (completed); §4 E.1/E.2 deferred i18n + CJK font to this SPEC; REQ-WC4-009 server-contract invariant 005 inherits; `restyle_test.go::TestAppbarRendered` S3-exclusion guard 005 inverts.
- `.moai/specs/SPEC-WEB-CONSOLE-001/` — original loopback console (zero-network + no-auth invariant source).
- `.moai/specs/SPEC-WEB-CONSOLE-002/` — port 3041 + web↔TUI parity (S1 invariant).
- `.moai/specs/SPEC-WEB-CONSOLE-003/` — flat project-config parity (S2a; same `internal/web` module, sibling scope fence).
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — 12-field frontmatter schema (canonical).
- `.claude/rules/moai/development/manager-develop-prompt-template.md` — Tier M Section A-E delegation template.
- `.claude/rules/moai/quality/boundary-verification.md` — the `data-i18n` key ↔ dictionary key cross-read (R6 boundary verification).
