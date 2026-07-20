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
- M4+M5 dark-theme retirement + test inversion — final run-phase commit (i18n.js `theme.aria` ×4 removed, `templ generate` regenerated, comment literals scrubbed)

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

_<pending sync-phase — populated by manager-docs>_
