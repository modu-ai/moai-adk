---
id: SPEC-DESIGN-MOAIWEBV2-002
document: plan
version: "0.1.1"
status: in-progress
created: 2026-07-21
updated: 2026-07-21
---

# SPEC-DESIGN-MOAIWEBV2-002 — Implementation Plan

> Milestones are ordered by decision-reversibility: the decisions most likely to change under review (tone/layout, color/a11y trade-offs, font-artifact pipeline) lead; mechanical sweeps (dark-theme removal, test inversion) close.

## §A. Context

Branch `fix/docs-layout-collapse` (== `origin/main`). Target module `internal/web` (templ + HTMX loopback console). Baseline `docs-site/static/moai-brand.css` FROZEN. All drift measurements in spec.md §A.2 were taken directly from the working tree on 2026-07-21 (commands + outputs reproduced in §B below where load-bearing).

## §B. Known Issues / Pre-measured Evidence

1. **Contrast pre-measurement** (WCAG 2.x relative luminance, executed python computation):
   - success `#5db872`: 2.45:1 vs `#ffffff`, 2.23:1 vs `#f4f4f4` → FAILS AA as text
   - warning `#d4a017`: 2.38:1 / 2.16:1 → FAILS AA as text (no live text usage today)
   - danger `#c64545`: 4.84:1 / 4.40:1 → PASSES on white surfaces (`.banner--error`, `.field-error` text); on `#f4f4f4` canvas 4.40:1 is marginal — verify actual surface per usage (banners sit on `--color-surface`, i.e. white-mixed)
   - info `#5db8a6`: 2.37:1 / 2.15:1 → FAILS AA as text (no live text usage today)
   - Consequence: `.banner--success` (console.css ~L430, `color: var(--color-success)`) REQUIRES the REQ-MWA-006 usage-site color-mix darkening. `.loopback__dot` (~L322) is a non-text UI dot accompanied by a text label — decide in M2 whether the 3:1 non-text rule applies or it is decorative (record the decision either way).
2. **Test entanglement**: `restyle_test.go` L72 lists `[data-theme="dark"]` in a required-token list, and `TestDarkModeAndThemeToggle` (L398-423) asserts the dark block + `moai-console-theme`/`data-theme` FOUC snippet. Dark removal and test inversion MUST land in the same milestone or the suite goes red mid-flight.
3. **app.js listener coupling**: the htmx-boost re-bind path wires `themeToggle` AND `uiLangSelect` together (app.js ~L33, L77). Removing theme code must not drop the language-select re-bind.
4. **i18n key parity**: `"theme.aria"` exists in all 4 locales (i18n.js L160/539/894/1249); `i18n_test.go` may assert cross-locale key-set parity — remove the key from all 4 locales atomically.
5. **Goorm Sans Code sourcing**: docs-site consumes `https://statics.goorm.io/fonts/GoormSansCode/v1.0.1/GoormSansCode.min.css` (CDN). Self-hosting requires obtaining the upstream font files and verifying the license BEFORE committing artifacts. The license is expected to be OFL but is NOT yet verified — treat as an open verification step (REQ-MWA-011 halt path), not a fact.
6. **Predecessor coupling**: SPEC-DESIGN-MOAIWEBV2-001 M3 deliberately styled the dark override (`--gradient-signature` solid dark). Retiring the block reverses that completed work by design — record the supersession note (spec.md §H) so audit does not flag it as accidental regression.
7. **FOUC dual-branch + server-rendered attr (audit D1/D2/D6)**: the `root.templ` `foucHeadScript` carries BOTH a theme branch (`moai-console-theme` → `data-theme`) AND a language branch (`moai-console-lang` → `<html lang>`, CJK @font-face activation, REQ-WC5-005 lineage). ONLY the theme branch is removed; the language branch is preserved VERBATIM (machine guard AC-MWA-003b). Additionally, `board.templ:15` ships a server-rendered `<html lang="en" data-theme="light">` attribute (remove `data-theme`, keep `lang="en"`), and `page.templ:9` carries a stale `data-theme` markup-contract comment (comment-only update).

## §C. Pre-flight

- [ ] `git rev-list --count --left-right origin/main...HEAD` → `0 0` (or local-ahead only) before run-phase spawn
- [ ] Record `BASELINE_SHA=$(git rev-parse origin/main)` BEFORE the first run-phase commit/push — the committed-diff comparator for AC-MWA-007a / AC-MWA-013 (per-milestone pushes must not vacuously empty the diff; committed edits must be caught)
- [ ] `go test ./internal/web/...` green on baseline (record as pre-change baseline)
- [ ] `hugo -s docs-site` builds warning-free on baseline (record; REQ-MWA-014 comparator)
- [ ] `templ` CLI + `pyftsubset` (fonttools) availability checked; if `pyftsubset` absent, install `fonttools` or perform the subset step on the maintainer machine (artifact is committed either way)
- [ ] Upstream Goorm Sans Code font files + license text obtained and archived alongside the subset

## §D. Constraints

- Offline invariant [HARD]: no external URL enters `console.css` (AC-MWA-010 guard).
- Zero server-contract change [HARD]: Go-side diff restricted to `*_templ.go` (regenerated) + `*_test.go`. AC-MWA-013 machine-checks this.
- Baseline FROZEN [HARD]: `docs-site/**` untouched (AC-MWA-007).
- Token-vs-usage separation [HARD]: AA failures fixed at usage site via color-mix; token bytes stay equal to docs-site.
- No time estimates; priority-ordered milestones only.

## §E. Self-Verification (run-phase exit)

Run the full AC matrix (acceptance.md §D) as a single-turn verification batch; record verbatim outputs in `progress.md` §E.2. Minimum set: dark grep 0 · 4-value byte-equal loop · font artifact ls + license test · http grep 0 · `go test ./internal/web/...` · pinned `TestDarkThemeAbsence` non-vacuous PASS · `templ generate` clean diff · `git diff --name-only $BASELINE_SHA..HEAD -- docs-site/` empty (+ clean working tree) · hugo warning-free · contrast table with post-fix ratios ≥ 4.5:1 for all live status-text usages.

## §F. Milestones (decision-reversibility order)

### M1 — Tone/layout alignment decisions (highest change-likelihood; user-facing)

- Audit console header/card/spacing against docs-site current direction (compact hero + mascot + 2-column sensibility; ref. docs-site commit `8f9e4e949`).
- Produce a decision table: header vertical rhythm, mascot placement, card padding/radius/shadow deltas, spacing-scale token deltas — each row: current value → target value → docs-site reference.
- Apply as CSS custom-property + component-rule + visual-layer templ edits only. NO IA/density restructuring (REQ-MWA-013).
- Evidence: decision table + before/after notes into progress.md §E.2 (AC-MWA-014).

### M2 — Status-token swap + AA carve-out (user-facing color + a11y decisions)

- Swap the 4 token values in `console.css` L130-133 to docs-site bytes (REQ-MWA-005).
- Re-measure contrast for every live status-text usage on its ACTUAL surface color; apply `color-mix(... , var(--color-ink) N%)` at failing usage sites until ≥ 4.5:1 (`.banner--success` confirmed; re-check `.banner--error` on its mixed surface). Decide + record the `.loopback__dot` non-text ruling.
- Evidence: post-fix contrast table (AC-MWA-006).

### M3 — Goorm Sans Code self-host subset (artifact pipeline)

- Verify license permits embedding + subset redistribution; on failure → REQ-MWA-011 halt + blocker report.
- Build-time/manual step (documented, artifact committed — same pattern as existing Pretendard subsets): `pyftsubset` with Latin + used-symbol unicode ranges (mirror the Pretendard subset range approach); weights: Regular + Bold (add Medium only if a live usage needs it — keep artifact count minimal).
- Commit `GoormSansCode-*.subset.woff2` + license file (e.g. `OFL-GoormSansCode.txt`) under `internal/web/assets/fonts/`; add `@font-face` blocks with `/static/fonts/` relative src; set `"Goorm Sans Code"` as leading `--font-mono` family (REQ-MWA-008/009/010).

### M4 — Dark-theme retirement sweep (mechanical)

- `console.css`: remove all 9 `data-theme` references (token block ~L245, overrides L375/380-381/432-433/485/604, L7 comment).
- `board.templ`: remove the `themeToggle` button AND the server-rendered `data-theme="light"` attribute from `<html>` (L15; keep `lang="en"`).
- `root.templ`: remove the `themeToggle` button; in the FOUC `<head>` snippet remove ONLY the theme branch (`moai-console-theme` read + `data-theme` set) and its comment mention — preserve the `moai-console-lang` / `<html lang>` language branch VERBATIM (CJK font activation, REQ-WC5-005 lineage; machine guard AC-MWA-003b).
- `page.templ`: update the stale L9 markup-contract comment to drop `data-theme` from the attribute enumeration (comment-only edit).
- `icons.templ`: remove sun/moon SVGs; `app.js`: remove theme read/write/toggle logic WITHOUT disturbing the shared htmx-boost re-bind path for `uiLangSelect` (§B.3); `i18n.js`: remove `theme.aria` from all 4 locales atomically (§B.4).
- `templ generate` after every `.templ` edit.

### M5 — Test-surface inversion + full verification batch (mechanical)

- Invert `restyle_test.go`: required-token list drops `[data-theme="dark"]`; `TestDarkModeAndThemeToggle` becomes the dark-ABSENCE assertion renamed to the PINNED name `TestDarkThemeAbsence` (AC-MWA-004 runs it by exact name with a non-vacuous `--- PASS` grep — a silent rename cannot match zero tests); FOUC assertions inverted to assert theme-branch absence AND language-branch (`moai-console-lang`) presence. Update `i18n_test.go` if key-set parity asserts `theme.aria`. Confirm `mascots_test.go` unaffected.
- Add/extend a stylesheet test asserting the 4 docs-site status bytes and the Goorm `@font-face` presence (regression lock for AC-MWA-005/009).
- Execute §E verification batch; populate progress.md §E.2/§E.3.

## §G. Anti-Patterns

- Blind `sed` sweeps across templ/css (platform + collateral risk — use targeted edits).
- Changing the FROZEN token bytes to "fix" contrast (violates baseline SSOT; carve-out is usage-scoped only).
- Committing font artifacts before license verification (unobserved-claim violation; REQ-MWA-011).
- Editing `*_templ.go` by hand (always regenerate).
- Drive-by IA/HTMX changes while touching templ visual markup (REQ-MWA-013).

## §H. Cross-References

- spec.md §B (GEARS requirements), §C (invariants), Exclusions
- acceptance.md §D (full AC matrix + commands)
- SPEC-WEB-CONSOLE-004 (offline invariant lineage + dark-mode partial supersession)
- `.claude/rules/moai/core/verification-claim-integrity.md` (evidence discipline for §E)
