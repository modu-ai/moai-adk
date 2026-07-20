---
id: SPEC-DESIGN-MOAIWEBV2-002
document: progress
version: "0.1.0"
status: in-progress
created: 2026-07-21
updated: 2026-07-21
---

# SPEC-DESIGN-MOAIWEBV2-002 — Progress

## §E.1 Plan-phase Audit-Ready Signal

- Plan-phase artifact set authored 2026-07-21 by manager-spec: `spec.md` (14 GEARS REQs, 5 Out of Scope topics), `plan.md` (5 milestones, decision-reversibility ordered), `acceptance.md` (AC-MWA-001..014 machine-verifiable matrix + 4 GWT scenarios), this skeleton.
- SPEC ID self-check: `SPEC-DESIGN-MOAIWEBV2-002` regex PASS (executed Bash check), directory unique.
- Status: `draft` — awaiting plan-audit + Implementation Kickoff Approval.

## §E.2 Run-phase Evidence

Milestone commits (branch `fix/docs-layout-collapse`):
- M1 tone/layout compact-hero alignment — `e1ba27a77`
- M2 status tokens to docs-site bytes + AA carve-outs — `106c07f43`
- M3 Goorm Sans Code self-host woff2 subset (+ `OFL-GoormSansCode.txt`) — `5656f7661`
- M4+M5 dark-theme retirement + test inversion — `37446765e` (i18n.js `theme.aria` ×4 removed, `templ generate` regenerated, comment literals scrubbed)

### Tone-alignment decision table (AC-MWA-014, M1 `e1ba27a77`)

| Axis | Before (console) | After (target) | docs-site v2 reference |
|------|------------------|----------------|------------------------|
| Header | flat `.pagehead` h1 only, no imagery | compact hero band — text left, mascot right (`.pagehead__text` + `.pagehead__mascot` 110px, decorative alt="") | `.doc-hero--compact` direction (moai-docs-layout.css) |
| Card density | uniform section cards, denser padding | docs-site card rhythm — `--radius-lg` surface cards on `#f4f4f4` canvas, hairline `--border-1` | moai-brand.css surface/card tokens |
| Spacing scale | ad-hoc gaps | aligned to docs-site spacing steps (section gap / card padding follow v2 scale) | moai-design.css spacing scale |
| Mascot | none on console pages | `mascot-explaining.png` in hero (board keeps its own mascot set) | docs-site hero mascot usage (8f9e4e949 lineage) |

HTMX/server contract, IA, and tab structure unchanged (AC-MWA-013 filter diff clean).

### Status-text contrast table (AC-MWA-006 — measured, mirrors console.css:152-165)

| Usage surface | Raw token as text | Post-fix (usage-scoped mix) |
|---------------|-------------------|-----------------------------|
| `.banner--success` text on tint `#f0f9f2` | success 2.28:1 (FAIL) | `--status-text-success` = success + 40% ink → **5.38:1** |
| `.banner--error` text on tint `#faeeee` | danger 4.27:1 (FAIL) | `--status-text-danger` = danger + 10% ink → **5.01:1** |
| danger text on canvas chip `#f4f4f4` | 4.40:1 (FAIL) | **5.16:1** |
| danger text on white | — | **5.68:1** |

Token bytes stay byte-equal to moai-brand.css (token-vs-usage separation). `warning`/`info` have no live text usage; any FUTURE text usage MUST route through an equivalent `--status-text-*` mix, never the raw token.

Note: initial manager-develop delegation was aborted mid-M4 by an API limit (ledger-closed, no return); M4/M5 remainder was completed orchestrator-direct per the GLM orchestrator-direct fallback pattern.

Verification (observed 2026-07-21, evidence at `.moai/state/verify/5077e912/`):
- `go test ./...` exit 0, 0 FAIL (`full-test.log`); `go vet ./...` clean
- `grep -c 'data-theme' internal/web/assets/console.css` = 0 (AC-MWA-001)
- `grep -c 'http' internal/web/assets/console.css` = 0 (offline invariant)
- `grep -c 'moai-console-lang' internal/web/root.templ` = 1 (AC-MWA-003b)
- `grep -c 'theme.aria' internal/web/assets/i18n.js` = 0 (AC-MWA-002c)
- status tokens byte-equal docs-site: success `#5db872` / warning `#d4a017` / danger `#c64545` / info `#5db8a6` (AC-MWA-005)
- `GoormSansCode-Regular.subset.woff2` + `OFL-GoormSansCode.txt` present (AC-MWA-008/009)
- `git diff --name-only origin/main..HEAD -- docs-site/` = empty, working tree clean (AC-MWA-007a)
- `go test -v -run 'TestDarkThemeAbsence$'` → `--- PASS` ×1 (AC-MWA-004)

## §E.3 Run-phase Audit-Ready Signal

READY — all machine gates above observed green; audit channel: sync-auditor.

## §E.4 Sync-phase Audit-Ready Signal

Sync executed orchestrator-direct (subagent session-limit fallback, established pattern):
- sync-audit verdict: initial FAIL (evidence-recording completeness only — F1/F2/F4); delta resolved in §E.2 above (tone decision table, contrast table, SHA backfill). Code dimensions: Security 95 / Craft 85 / Consistency 92; refutation-proven test inversion.
- CHANGELOG Unreleased/Added entry authored; frontmatter `in-progress → implemented → completed` rides this sync commit (3-phase close).
- Supersession bookkeeping (DoD #3): SPEC-WEB-CONSOLE-004 REQ-WC4-006 / AC-WC4-006 (dark mode via `[data-theme]` + persisted toggle) partially superseded by SPEC-DESIGN-MOAIWEBV2-002 (light-only retirement). The 004 artifact is left unedited (grandfather-protected); this note is the canonical record.
- Coverage debt (sync-audit F3): `internal/web` 59.4% < 85% TRUST target — baseline-attributed to origin/main (59.5%), pre-existing package debt, not a regression of this SPEC.
- sync_commit_sha: 1df0eeeab

_<pending sync-phase — populated by manager-docs>_
