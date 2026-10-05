# t1527 — moai init/update TUI/TUX overhaul (D1-D6) — progress record

## Identity

- card: t1527
- lane: lane-24 (factory run tm9i7y)
- branch: WT-tui-tux-overhaul (created in place from the resident tree's branch)
- base SHA: `d0e54b322` (recorded per dispatch; no rebase — integration reconciles)
- worktree: /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1453 (only tree touched)
- binary for captures: /tmp/t1527-moai-before (fresh-inode copy of bin/moai, built from d0e54b322)

## Milestones

- M0 evidence set-up — DONE
  - baseline test run launched BEFORE any source edit (background, env-scrubbed,
    `-count=1 -timeout 30m ./internal/cli/... ./internal/tui/...`) →
    `.moai/reports/t1527/test-baseline.txt`
  - live capture BEFORE (d0e54b322 build):
    - `.moai/reports/t1527/live-before-init.txt` (init --non-interactive, exit 0, 25 lines)
    - `.moai/reports/t1527/live-before-update.txt` (first update — took the v2 clean-reinstall
      branch because a fresh init deploys the legacy paths; exit 0, 19 lines)
    - `.moai/reports/t1527/live-before-update2.txt` (second update --yes --force — v3 template-sync
      surface, exit 0, 64 lines)
  - defect-4 pin attempt: "Install it yourself" — 0 grep hits in internal/ pkg/ cmd/
    AND absent from all three captures → **not pinned by capture** (recorded; D5 ships skeleton-only)
- M1 D1 banner dedup
- M2 D2 progress model (single ✓-line model, gauge once, labels)
- M3 D3 class counts carried into outcome + preview fallback trim
- M4 D4 severity prefix helper + init/update surface migration
- M5 D5 terminal ACTION-REQUIRED/Reference block + init card move + escalation
- M6 D6 locale measurement recorded (framework exists at tui.Translate; init/update emitters
  bypass it — English-only confirmed by grep + live capture; no i18n wiring in this card)
- M7 verification (tests, race, vet, lint, windows build) + live capture AFTER

## Measured constraints discovered on this base (differences from the dispatch map)

- `runTemplateSyncWithProgress` signature is asserted verbatim by
  update_mirror_heal_test.go:337-338 — signature stays `(cmd) (skipped bool, err error)`.
- update_test.go:111 / coverage_test.go:465 / integration_test.go:98 assert output contains
  "Current version" on a normal pass → the top-of-run KV stays on the normal pass; D1 removes the
  template-sync duplicate + binary-step duplicate and suppresses the top KV only on the re-executed
  pass (new env constant), so exactly one version line remains on every path.
- reporter seam: init.go:783 uses newSpinnerReporter for the init executor — reporter.go untouched.

## D6 measurement (recorded for the verdict)

- internal/tui/i18n.go: tui.Translate + embedded messages/*.yaml catalogs EXIST (fallback lang→en→key).
- internal/cli init/update emitters (init.go, update.go, update_template_sync.go, update_tux.go):
  zero tui.Translate uses — raw English format strings; init.go:614 touches conversation_language
  only as a wizard config pre-fill, never for output selection.
- Live captures: all output lines are English.
- Decision: framework present but wiring the whole init/update surface (catalog authoring +
  language resolution in CLI) is not trivial reuse → no i18n introduction in this card.
