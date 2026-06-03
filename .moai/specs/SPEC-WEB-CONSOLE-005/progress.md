# Progress — SPEC-WEB-CONSOLE-005

## Plan Phase

- **Tier**: M (3 artifacts: spec.md + plan.md + acceptance.md, + progress.md)
- **cycle_type (run-phase)**: tdd (per `development_mode: tdd`)
- **status**: draft
- **Cohort**: web-console-v3 **final** member (005, cohort-internal label "S3" — the cohort terminator); siblings 001 (모태) / 002 S1 / 003 S2a / 004 visual-restyle — **all completed**
- **Primary source**: `.moai/design/web-console-handoff/IMPLEMENTATION-HANDOFF.md` (§5.3 i18n decision / §6 cohort decomposition / §7 font strategy + offline-CDN discovery / §9 M5) + the design `i18n.js` dictionary (derivative source)
- **REQ count**: 12 (REQ-WC5-001 .. REQ-WC5-012)
- **AC count**: 13 logical (15 rows incl. AC-008a/008b + AC-010a/010b split)
- **MUST-PASS invariants**:
  1. **interface language ≠ content language** (REQ-WC5-008 / AC-WC5-008a) — the appbar langpick mutates NO server-submitted field; POST round-trip byte-identical regardless of interface language
  2. **server contract / validate.go byte-unchanged** (REQ-WC5-010 / AC-WC5-010a)
  3. **offline / CDN-free** (REQ-WC5-006/011 / AC-WC5-011) — all font + dictionary assets via `go:embed`; zero external `https://` font/style/script URL

### Scope (two parts)

- **E.1 — Web interface i18n** (greenfield; handoff §9 M5): `data-i18n` attribute wiring on the page chrome + a new client-side `internal/web/assets/i18n.js` dictionary (en/ko/ja/zh, STATIC, no server round-trip) + an appbar `langpick` `<select>` + `localStorage("moai-console-lang")` persistence + load-time apply. `rv.*` design-review keys EXCLUDED.
- **E.2 — CJK self-host webfont coverage** (handoff §7): the 004-embedded Pretendard ko+Latin subset does NOT cover ja hiragana/katakana/kanji or zh hanzi; add a self-host woff2 subset (Option (c) — glyph-subset to exactly the **shipped** ja/zh dictionary string set, a low-hundreds glyph set; estimate ~290, tokenizer-dependent ~284–287 band — run-phase measures the actual shipped-dictionary count and records it here) for the ja/zh interface glyphs.

### Plan-phase artifacts created

- spec.md — 12 GEARS REQs (REQ-WC5-001..012) + 13 AC index + §4 Exclusions (E.1 server-side i18n→out / E.2 content-language fieldset→untouched / E.3 nested config→S2b out / E.4 rv.* + .review aside / E.5 anti-patterns). §1.5 carries the estimated glyph demand (low-hundreds, ~290 estimate — run-phase re-measures against the shipped dictionary) grounding the font decision. §1.4 carries the interface≠content-language invariant rationale. REQ-WC5-003 fixes the appbar picker id to `uiLangSelect` (non-colliding with the live `langSelect` content-language helper template).
- plan.md — Tier M justification + §A Context + §B Known Issues (R1 server-contract-leak-via-langpick HIGHEST + R2 offline CDN + R3 CJK bloat + R4 004-test-guard contradiction + R5 rv.* leak + R6 data-i18n/dictionary drift + R7 html-lang/font interaction) + §C Pre-flight + §D Constraints + §E Self-Verification + **§F Font Strategy Research & Proposal** (3-option trade-off table; recommends Option (c) glyph-subset-to-exact-dictionary) + §G Milestones (M1 CJK font → M2 dictionary + data-i18n → M3 appbar langpick + client apply [server-contract gate] → M4 test reconciliation + a11y + closure) + §H Anti-Patterns + §I Cross-References
- acceptance.md — full Given-When-Then for all 13 AC (15 rows) + §C traceability + §D edge cases (EC-1..EC-10) + §E Definition of Done + §F forward-looking cohort-closure checks
- progress.md — this file

### Tier determination

Tier **M** — multi-file cross-asset change (page.html.tmpl + new i18n.js + app.js + console.css + new CJK woff2 subset font + assets.go go:embed + restyle_test.go reconciliation), little-to-no production Go change (interface language is client-only), no constitutional change, < ~15 files. Exceeds Tier S (<5 files / <300 LOC) due to the new client-i18n surface + new offline font pipeline + server-contract regression surface + sibling-test reconciliation. Not Tier L (no new persistence model, no nested-config redesign).

### Font strategy decision (plan.md §F)

- **Recommended: Option (c)** — glyph-subset the CJK font to EXACTLY the **shipped** ja/zh interface dictionary string set (a low-hundreds glyph set) via `pyftsubset --text=` against `internal/web/assets/i18n.js`, sourced from Noto Sans SC (zh) + Noto Sans JP (ja) under OFL-1.1. Appended to `--font-sans` after Pretendard so en/ko stay on Pretendard (no regression).
- **Why (c) over (a)/(b)**: the dictionary is a FIXED known glyph set in the low hundreds → exact subset is tens of KB (vs Option (b) full Noto multi-MB bloat). Option (a) single-Pretendard rejected because Pretendard has no simplified-SC coverage for zh. Same `pyftsubset` toolchain 004 already used — no new build dependency. The exact glyph count is measured at run-phase against the shipped dictionary (the strategy does not depend on the precise number — auditor D2: the design-file count is provisional + tokenizer-dependent).

### Critical run-phase note (sibling test contradiction)

004's `internal/web/restyle_test.go::TestAppbarRendered` (lines 207-216) asserts the rendered-body literals `class="langpick"`, `id="langSelect"`, and `data-i18n` are ABSENT (the 004 S3-exclusion guard). 005 INTENTIONALLY lands `class="langpick"` + `data-i18n` + the NEW `id="uiLangSelect"` → the `class="langpick"`/`data-i18n` assertions WILL FAIL unless inverted. The run-phase MUST reconcile this test (move `class="langpick"` + `data-i18n` to the expected block AND add an EXPECT for `id="uiLangSelect"`) — it is the planned guard reversal (REQ-WC5-012 / AC-WC5-012), NOT a regression. CRITICAL: do NOT invert the stale `id="langSelect"` forbidden-string into an EXPECT — it referenced the never-landed original id; the appbar picker uses `uiLangSelect` because `langSelect` is the live `{{define "langSelect"}}` content-language helper. This is an AC that inverts an existing passing test — the run-phase TDD RED must account for it explicitly.

plan_complete_at: 2026-06-03
plan_status: audit-ready
