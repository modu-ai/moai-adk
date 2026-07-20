---
id: SPEC-DESIGN-MOAIWEBV2-002
title: "moai web console → docs-site v2 alignment: dark-theme retirement, status tokens, Goorm Sans Code self-host, tone"
version: "0.1.1"
status: in-progress
created: 2026-07-21
updated: 2026-07-21
author: manager-spec
priority: P2
phase: "v3.1.0 target"
module: "internal/web"
lifecycle: spec-anchored
tags: "design, web-console, dark-theme, status-tokens, fonts, goorm-sans-code, templ, accessibility"
tier: M
related_specs: [SPEC-DESIGN-DOCSV2-001, SPEC-DESIGN-MOAIWEBV2-001, SPEC-WEB-CONSOLE-004]
---

# SPEC-DESIGN-MOAIWEBV2-002 — moai web console → docs-site v2 design alignment (remaining drift)

> Follow-up to Epic DESIGN-V2. `SPEC-DESIGN-DOCSV2-001` (completed) established the docs-site v2 design system (the FROZEN baseline); `SPEC-DESIGN-MOAIWEBV2-001` (completed) aligned the `moai web` console's core green/canvas tokens to it. This SPEC closes the remaining drift on four user-confirmed axes: (1) dark-theme retirement (light-only, matching docs-site policy), (2) semantic status colors unified to docs-site values with a WCAG AA contrast carve-out, (3) Goorm Sans Code self-hosted woff2 subset as the code/mono family, (4) tone/component-level layout alignment (compact hero + mascot + spacing sensibility) — visual layer only, zero server-contract change.

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-21 | manager-spec | Initial plan-phase authoring (draft). Tier M. GEARS requirements for dark-theme retirement, status-token unification + AA carve-out, Goorm Sans Code self-host subset, tone alignment. |
| 0.1.1 | 2026-07-21 | manager-spec | Plan-audit iter-1 fixes (D1-D7): `board.templ:15` server-rendered `data-theme` attr + `page.templ:9` stale comment in scope; FOUC language-branch preservation guard (AC-MWA-003b); pinned `TestDarkThemeAbsence`; `$BASELINE_SHA` committed-diff comparators; docs-site/ pathspec widened; GEARS relabels (REQ-MWA-002 Unwanted, REQ-MWA-006/011 Event-driven). |

---

## §A. Context & Intent

### A.1 Baseline (SSOT, FROZEN — verification-only, never edited by this SPEC)

docs-site v2 design system per SPEC-DESIGN-DOCSV2-001:

- Neutral canvas `#f4f4f4`, signature green `#3d7d5f` (hover `#316750`, active `#265240`), ink `#060606`
- Light-only single theme (no dark theme)
- Pretendard (KR title + body), Goorm Sans Code (code bodies)
- Token files: `docs-site/static/moai-brand.css` (FROZEN v2 tokens), `moai-design.css`, `moai-docs-tokens.css`, `moai-docs-theme.css`, `moai-docs-layout.css`

### A.2 Target and measured drift (ground truth, measured 2026-07-21 on `fix/docs-layout-collapse` == `origin/main`)

Target: `internal/web` (Go templ + HTMX console, "moai web"). `internal/web/assets/console.css` already carries the v2 green/canvas core tokens. Remaining drift:

| # | Drift axis | Measured current state |
|---|-----------|------------------------|
| 1 | Dark theme | 9 `data-theme` occurrences in `console.css` (token block at ~L245 + overrides at L375/380-381/432-433/485/604 + L7 comment); `themeToggle` buttons in `board.templ:122` + `root.templ:188`; sun/moon SVGs in `icons.templ:30,32`; theme logic + `moai-console-theme` localStorage key + FOUC `<head>` snippet in `app.js` / `root.templ`; `"theme.aria"` i18n key in all 4 locales of `i18n.js`; `restyle_test.go` asserts dark presence (`TestDarkModeAndThemeToggle`, AC-WC4-006 lineage); server-rendered `<html lang="en" data-theme="light">` in `board.templ:15`; stale `data-theme` markup-contract comment in `page.templ:9` |
| 2 | Status colors | `console.css` L130-133: `#2e8a63`/`#c47b2a`/`#c44a3a`/`#2a8a8c` vs docs-site `moai-brand.css` L37-40: `#5db872`/`#d4a017`/`#c64545`/`#5db8a6` |
| 3 | Code/mono font | `--font-mono` is an OS fallback stack; no Goorm Sans Code anywhere in `internal/web/assets/fonts/` (docs-site loads it from the statics.goorm.io CDN — the console MUST self-host per its offline invariant) |
| 4 | Tone/layout | Console header/card/spacing predates the docs-site compact-hero + mascot + 2-column direction (recent docs-site work) |

Pre-measured WCAG contrast of the new status values (WCAG 2.x relative-luminance formula): success `#5db872` 2.45:1 on `#fff` / 2.23:1 on `#f4f4f4`; warning `#d4a017` 2.38:1 / 2.16:1; danger `#c64545` 4.84:1 / 4.40:1; info `#5db8a6` 2.37:1 / 2.15:1. Live status-TEXT usages today: `.banner--success` (fails AA after swap → carve-out required) and `.banner--error` / `.field-error` danger text (passes on white surfaces). `warning`/`info` currently have no text usage sites in `console.css` (definitions only).

### A.3 Intent

Make the loopback web console visually indistinguishable in design language from the public docs-site (adk.mo.ai.kr): one brand, one theme (light), one status palette, one code font — while preserving the console's offline invariant, HTMX contracts, and information density.

---

## §B. Requirements (GEARS)

### Group 1 — Dark-theme retirement (light-only)

- **REQ-MWA-001** (Ubiquitous): The console stylesheet `internal/web/assets/console.css` shall contain zero `data-theme` references — all `[data-theme="dark"]` override rules (including the dark token block introduced under REQ-WC4-006 and restyled by SPEC-DESIGN-MOAIWEBV2-001 M3) and the descriptive comment are removed.
- **REQ-MWA-002** (Unwanted): The console UI shall not render a theme-toggle control: the `themeToggle` button is removed from `board.templ` and `root.templ`, the `icon-sun`/`icon-moon` SVGs are removed from `icons.templ`, and the `theme.aria` i18n key is removed from all 4 locales in `i18n.js`.
- **REQ-MWA-003** (Event-driven): **When** a console page loads, the console (`app.js`, the `<head>` FOUC snippet rendered by `root.templ`, AND the server-rendered `<html>` element in `board.templ`) shall neither read nor write any theme persistence key (`moai-console-theme`) nor carry/set a `data-theme` attribute on the document element. The FOUC snippet's language branch (`moai-console-lang` → `<html lang>`, REQ-WC5-005 CJK font-activation lineage) is preserved verbatim — only the theme branch is removed (machine guard: AC-MWA-003b).
- **REQ-MWA-004** (Ubiquitous): The test surface shall assert dark-theme ABSENCE: `restyle_test.go` `TestDarkModeAndThemeToggle` (AC-WC4-006 lineage) and the required-token list entry for `[data-theme="dark"]` are inverted; this SPEC partially supersedes the dark-mode clause of SPEC-WEB-CONSOLE-004 (REQ-WC4-006).

### Group 2 — Semantic status colors (docs-site values + AA carve-out)

- **REQ-MWA-005** (Ubiquitous): The console's semantic status tokens shall be byte-equal to the docs-site baseline: `--color-success: #5db872`, `--color-warning: #d4a017`, `--color-danger: #c64545`, `--color-info: #5db8a6`.
- **REQ-MWA-006** (Event-driven): **When** a status-token TEXT usage in the console fails WCAG AA contrast (≥ 4.5:1 normal text) after the swap — pre-measured: `.banner--success` text at 2.45:1 — the usage site shall darken via `color-mix(in srgb, var(--color-<status>), var(--color-ink) <N>%)` (or equivalent usage-scoped darkening) until AA passes; the token value itself shall not change.
- **REQ-MWA-007** (Unwanted): The alignment shall not modify any file under `docs-site/` — the baseline is verification-only.

### Group 3 — Goorm Sans Code self-host subset

- **REQ-MWA-008** (Ubiquitous): The repository shall ship committed Goorm Sans Code woff2 subset artifact(s) under `internal/web/assets/fonts/` (glyph coverage: Latin + used symbols, mirroring the existing Pretendard subset approach), with the font's license file shipped alongside (pattern: `OFL-NotoSansCJK.txt`).
- **REQ-MWA-009** (Ubiquitous): `console.css` shall register `@font-face` for Goorm Sans Code with relative `/static/fonts/` source URLs and shall place `"Goorm Sans Code"` as the leading family of `--font-mono` (existing OS fallback stack preserved after it).
- **REQ-MWA-010** (Unwanted): The console shall not fetch any font or stylesheet from an external network origin — `grep -c 'http' internal/web/assets/console.css` remains 0 (REQ-WC4 offline-invariant lineage).
- **REQ-MWA-011** (Event-driven): **When** license verification finds that Goorm Sans Code does not permit embedding/subset redistribution, the run shall halt Group 3, keep the current OS fallback mono stack, and return a structured blocker report to the orchestrator (no unverified license claim is committed).

### Group 4 — Tone / component-level layout alignment

- **REQ-MWA-012** (Ubiquitous): The console's visual language — header treatment (compact hero + mascot placement), card density, and spacing scale — shall align with the docs-site v2 current direction; changes are limited to CSS custom properties, component class rules, and templ markup of the visual layer.
- **REQ-MWA-013** (Unwanted): The alignment shall not change any HTMX contract, route, handler signature, server-rendered fragment target, or the console's information density/IA — Go-side diffs are restricted to regenerated `*_templ.go` and `*_test.go` files (zero server-contract change; SPEC-WEB-CONSOLE-004 constraint class).

### Group 5 — Baseline consistency (docs-site side, verification-only)

- **REQ-MWA-014** (Compound): **While** the docs-site baseline is FROZEN, **When** the run-phase completes, the verification set shall confirm the docs-site `hugo` build is warning-free and `docs-site/static/` is byte-unchanged; **When** a token mismatch is discovered that the web side cannot resolve alone, the implementing agent shall return a blocker report instead of editing docs-site.

---

## §C. Constraints & Invariants

- **Offline invariant** [HARD]: zero external network font/CSS fetches in the console (REQ-WC4 lineage). Self-hosting is the only path for Goorm Sans Code.
- **Zero server-contract change** [HARD]: visual layer only. No handler, route, HTMX attribute-contract, or fragment-id changes consumed by the server.
- **Baseline FROZEN** [HARD]: `docs-site/static/moai-brand.css` values are the SSOT; the web side conforms to them, never the reverse. Contrast failures are resolved at the usage site (color-mix), never by changing the token.
- **Committed artifacts**: font subsets enter the repo as committed woff2 files (build-time/manual `pyftsubset` step documented in plan.md), like the existing Pretendard subsets.
- **templ discipline**: every `.templ` edit is followed by `templ generate`; generated `*_templ.go` stays in sync (CI-verifiable via clean regeneration diff).

---

## §D. Acceptance Criteria (summary — full matrix in acceptance.md)

| AC | REQ | One-line gate |
|----|-----|---------------|
| AC-MWA-001 | REQ-MWA-001 | `grep -c 'data-theme' console.css` → 0 |
| AC-MWA-002 | REQ-MWA-002 | `themeToggle` / sun-moon icons / `theme.aria` all absent |
| AC-MWA-003 | REQ-MWA-003 | theme key + `data-theme` absent from `app.js` / `root.templ` / `board.templ` (003a); `moai-console-lang` language branch preserved in `root.templ` (003b guard) |
| AC-MWA-004 | REQ-MWA-004 | pinned `TestDarkThemeAbsence` runs non-vacuously (`--- PASS` grep ≥ 1) and exits 0 |
| AC-MWA-005 | REQ-MWA-005 | 4 status values byte-equal to `moai-brand.css` |
| AC-MWA-006 | REQ-MWA-006 | contrast table recorded; failing text usages carry usage-scoped darkening ≥ 4.5:1 |
| AC-MWA-007 | REQ-MWA-007/014 | `git diff --name-only $BASELINE_SHA..HEAD -- docs-site/` empty + clean working tree; hugo build warning-free |
| AC-MWA-008 | REQ-MWA-008 | `GoormSansCode*.woff2` ≥ 1 + license file committed |
| AC-MWA-009 | REQ-MWA-009 | `@font-face` registered, `/static/fonts/` relative src, `--font-mono` leads with Goorm Sans Code |
| AC-MWA-010 | REQ-MWA-010 | `grep -c 'http' console.css` → 0 |
| AC-MWA-011 | REQ-MWA-013 | `go test ./internal/web/...` exit 0 |
| AC-MWA-012 | §C templ discipline | `templ generate` → clean diff |
| AC-MWA-013 | REQ-MWA-013 | Go-side diff vs `$BASELINE_SHA` restricted to `*_templ.go` / `*_test.go` / `.templ` / `assets/` |
| AC-MWA-014 | REQ-MWA-012 | tone-alignment decision table + before/after evidence recorded in progress.md §E.2 |

---

## Exclusions (Out of Scope)

The following are explicitly out of scope for this SPEC:

### Out of Scope — docs-site edits

- Any modification to `docs-site/**` (tokens, layouts, content). The docs-site is the FROZEN verification-only baseline; an unresolvable token mismatch produces a blocker report, not an edit.

### Out of Scope — console IA / HTMX restructuring

- Restructuring the console's information density, navigation, panel composition, or any HTMX request/response contract. This SPEC is visual-layer only (SPEC-WEB-CONSOLE-004 constraint class).

### Out of Scope — full CJK glyph coverage for Goorm Sans Code

- Korean/Japanese/Chinese glyph subsets for the code font. The committed subset covers Latin + used symbols only (code bodies are ASCII-dominant); CJK code-comment rendering falls through to the existing fallback stack.

### Out of Scope — dark theme re-introduction in any form

- `prefers-color-scheme` auto-dark media queries, per-user theme preference storage, or any dark-mode successor mechanism. Light-only is the docs-site v2 policy this SPEC adopts.

### Out of Scope — non-mono font family changes

- Pretendard / Noto Sans CJK body-font subsets and their `@font-face` registrations remain untouched; only the code/mono family changes.

---

## §H. Cross-References

- `SPEC-DESIGN-DOCSV2-001` — docs-site v2 design system (baseline SSOT, completed)
- `SPEC-DESIGN-MOAIWEBV2-001` — predecessor console alignment (completed; its M3 dark-override styling is retired by Group 1)
- `SPEC-WEB-CONSOLE-004` — REQ-WC4 offline invariant (preserved) + REQ-WC4-006 dark mode (partially superseded by this SPEC; sync-phase MAY record `partially_superseded_by` bookkeeping on that SPEC via manager-docs)
- `plan.md` — implementation plan, milestone ordering, font-subset pipeline
- `acceptance.md` — full AC matrix, GWT scenarios, DoD
