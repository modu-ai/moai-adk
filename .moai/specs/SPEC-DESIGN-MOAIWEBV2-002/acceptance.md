---
id: SPEC-DESIGN-MOAIWEBV2-002
document: acceptance
version: "0.1.1"
status: in-progress
created: 2026-07-21
updated: 2026-07-21
---

# SPEC-DESIGN-MOAIWEBV2-002 — Acceptance Criteria

All commands run from repo root `/Users/goos/MoAI/moai-adk-go` unless noted. Every PASS claim in progress.md §E.2 must cite the executed command + verbatim output (verification-claim integrity §2).

`$BASELINE_SHA` below is the pre-flight-recorded origin/main SHA (`BASELINE_SHA=$(git rev-parse origin/main)`, captured BEFORE the first run-phase commit — plan.md §C). Committed-diff ACs compare against it so committed edits are caught and per-milestone pushes cannot vacuously empty the diff.

## §D. AC Matrix (machine-verifiable)

| AC | REQ | Command | Expected |
|----|-----|---------|----------|
| AC-MWA-001 | REQ-MWA-001 | `grep -c 'data-theme' internal/web/assets/console.css` | `0` |
| AC-MWA-002a | REQ-MWA-002 | `grep -c 'themeToggle' internal/web/board.templ internal/web/root.templ \| grep -v ':0$'` | no output (both files 0) |
| AC-MWA-002b | REQ-MWA-002 | `grep -cE 'icon-sun\|icon-moon' internal/web/icons.templ` | `0` |
| AC-MWA-002c | REQ-MWA-002 | `grep -c '"theme.aria"' internal/web/assets/i18n.js` | `0` |
| AC-MWA-003a | REQ-MWA-003 | `grep -cE 'data-theme\|moai-console-theme' internal/web/assets/app.js internal/web/root.templ internal/web/board.templ \| grep -v ':0$'` | no output (all 3 files 0 — includes the server-rendered `<html>` attr in board.templ) |
| AC-MWA-003b | REQ-MWA-003 | `grep -c 'moai-console-lang' internal/web/root.templ` | ≥ 1 (FOUC language branch preserved verbatim — REQ-WC5-005 CJK font activation) |
| AC-MWA-004 | REQ-MWA-004 | `go test -v -run 'TestDarkThemeAbsence$' ./internal/web/ 2>&1 \| tee /tmp/ac-mwa-004.log; grep -c -- '--- PASS: TestDarkThemeAbsence' /tmp/ac-mwa-004.log` | test exit 0 AND grep count ≥ 1 (pinned name actually ran — a zero-match vacuous pass is rejected) |
| AC-MWA-005 | REQ-MWA-005 | `for t in success warning danger info; do a=$(grep -oE -- "--color-$t:[[:space:]]*#[0-9a-f]{6}" docs-site/static/moai-brand.css \| grep -oE '#[0-9a-f]{6}'); b=$(grep -oE -- "--color-$t:[[:space:]]*#[0-9a-f]{6}" internal/web/assets/console.css \| grep -oE '#[0-9a-f]{6}'); [ -n "$a" ] && [ "$a" = "$b" ] && echo "$t OK $a" \|\| echo "$t MISMATCH $a/$b"; done` | 4 lines `... OK ...`, zero MISMATCH |
| AC-MWA-006 | REQ-MWA-006 | contrast script (WCAG formula) over final computed colors of every live status-TEXT usage on its actual surface | every live status-text usage ≥ 4.5:1; table recorded in progress.md §E.2; failing raw-token usages carry usage-scoped `color-mix` (token bytes unchanged — cross-check with AC-MWA-005) |
| AC-MWA-007a | REQ-MWA-007/014 | `git diff --name-only "$BASELINE_SHA"..HEAD -- docs-site/ \| wc -l` AND `git status --porcelain docs-site/ \| wc -l` | both `0` (no committed AND no uncommitted change under the FULL `docs-site/` pathspec) |
| AC-MWA-007b | REQ-MWA-014 | `hugo -s docs-site 2>&1 \| tee /tmp/hugo.log; grep -ci 'WARN' /tmp/hugo.log` | build exit 0 AND `0` warnings |
| AC-MWA-008a | REQ-MWA-008 | `ls internal/web/assets/fonts/GoormSansCode*.woff2 \| wc -l` | ≥ 1 |
| AC-MWA-008b | REQ-MWA-008/011 | `ls internal/web/assets/fonts/ \| grep -ci 'goorm'` (license file, e.g. `OFL-GoormSansCode.txt`) counted incl. woff2; explicit: `test -f internal/web/assets/fonts/OFL-GoormSansCode.txt && echo OK` | `OK` (filename may adapt to the actually-verified license type; a license file MUST ship) |
| AC-MWA-009a | REQ-MWA-009 | `grep -c 'Goorm Sans Code' internal/web/assets/console.css` | ≥ 2 (`@font-face` + `--font-mono`) |
| AC-MWA-009b | REQ-MWA-009 | `grep -cE "url\(['\"]?/static/fonts/GoormSansCode" internal/web/assets/console.css` | ≥ 1 |
| AC-MWA-010 | REQ-MWA-010 | `grep -c 'http' internal/web/assets/console.css` | `0` |
| AC-MWA-011 | REQ-MWA-013 | `go test ./internal/web/...` | exit 0 |
| AC-MWA-012 | §C templ discipline | `templ generate && git diff --exit-code -- 'internal/web/*_templ.go'` | exit 0 (generated files in sync) |
| AC-MWA-013 | REQ-MWA-013 | `git diff --name-only "$BASELINE_SHA"..HEAD -- internal/web \| grep -vE '(_templ\.go$\|_test\.go$\|\.templ$\|/assets/)'` | empty output (no server-side Go change vs the pre-flight baseline SHA) |
| AC-MWA-014 | REQ-MWA-012 | tone-alignment decision table (current → target → docs-site ref per row) + before/after evidence present in progress.md §E.2 | table present with ≥ 1 row per axis (header, card density, spacing scale, mascot) |

### AC sub-ID note

`AC-MWA-002a/002b/002c` etc. are paired sub-criteria of one logical AC group (acceptance-criteria sub-ID convention; SPEC IDs never carry alpha suffixes).

## Given-When-Then Scenarios

### Scenario 1 — Light-only console (dark retirement)

- **Given** a fresh browser profile with `localStorage` entry `moai-console-theme=dark` left over from a prior version
- **When** the user opens `moai web` and the console page loads
- **Then** the page renders in the light theme, no `data-theme` attribute is set on `<html>`, no theme-toggle button exists in the header, and the stale localStorage key is simply ignored (no read, no write, no crash)

### Scenario 2 — Status banner after token swap

- **Given** the console renders a success banner (`.banner--success`) after a config save
- **When** the status palette is the docs-site v2 set (`#5db872` family)
- **Then** the banner's tint/border derive from the new token AND the banner TEXT color measures ≥ 4.5:1 against its actual surface (usage-scoped darkening applied), while `--color-success` in `console.css` remains byte-equal to `moai-brand.css`

### Scenario 3 — Offline code font

- **Given** a machine with no network access
- **When** the console renders a `.mono` / code element
- **Then** Goorm Sans Code loads from the self-hosted `/static/fonts/` woff2 subset with zero external requests; a missing glyph (e.g. CJK comment) falls through to the OS mono fallback stack without layout breakage

### Scenario 4 — Baseline untouched

- **Given** the run-phase is complete
- **When** the verification batch runs
- **Then** `docs-site/` shows zero diff and its hugo build is warning-free — any needed baseline change would have surfaced as a blocker report, not an edit

## Edge Cases

- Stale `moai-console-theme` localStorage key on returning users (Scenario 1 — must be inert).
- htmx-boosted navigation re-binding: removing the theme listener must not break `uiLangSelect` re-bind (plan §B.3); i18n language switch still works after boost navigation.
- ja/zh interface-language users: the preserved FOUC language branch still sets `<html lang>` pre-paint (CJK @font-face activation, REQ-WC5-005 — no system-ui flash); only the theme branch disappears (AC-MWA-003b).
- `warning` / `info` tokens have no live text usage today — the AA carve-out rule (REQ-MWA-006) still binds FUTURE usages; record this in the contrast table so the rule is discoverable.
- `danger` text on the `#f4f4f4` canvas measures 4.40:1 (< 4.5) — if any danger-text usage sits directly on canvas (not on a white-mixed surface), it needs the carve-out too; verify per actual surface.
- woff2 subset missing a symbol glyph actually used by console output → fallback renders it; visual check that mixed-font lines do not jump line-height.
- License outcome ≠ OFL (REQ-MWA-011): Group 3 halts with a blocker report; AC-MWA-008/009 are then N/A-blocked, not silently skipped.

## Quality Gates

- `go test ./internal/web/...` exit 0 (AC-MWA-011); full-suite `go test ./...` green before sync commit (CLAUDE.local.md §6 full-suite rule)
- `golangci-lint run` clean on touched packages
- `templ generate` idempotent (AC-MWA-012)
- No new external URL in any `internal/web/assets/*.css` (AC-MWA-010)
- `moai spec lint` clean for this SPEC directory's artifacts

## Definition of Done

1. All AC-MWA-001..014 PASS with verbatim command outputs recorded in progress.md §E.2 (evidence-bearing, no summarized claims)
2. Frontmatter transitions follow ownership: `draft → in-progress` (manager-develop), `→ implemented/completed` (manager-docs)
3. Supersession bookkeeping noted for SPEC-WEB-CONSOLE-004 REQ-WC4-006 (sync-phase, manager-docs)
4. CHANGELOG + docs sync via manager-docs (sync-phase; not this SPEC's run scope)
