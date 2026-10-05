# t1527 — moai init/update TUI/TUX overhaul (D1-D6) — progress record

## Identity

- card: t1527
- lane: lane-24 (factory run tm9i7y)
- branch: WT-tui-tux-overhaul (created in place from the resident tree's branch)
- base SHA: `d0e54b322` (recorded per dispatch; no rebase — integration reconciles)
- worktree: /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1453 (only tree touched)

## Commits (oldest first)

- bd68cdda2 chore(t1527): pin live-before captures and progress record
- 1aea9908a fix(cli): D1 single version identity surface (MOAI_UPDATE_REEXEC)
- c74dbd576 refactor(cli): D2+D3 single progress model and summary-in-outcome
- 191ecf9ae refactor(cli): D4 one severity-prefix vocabulary
- 1995ccfe0 feat(cli): D5 terminal ACTION REQUIRED/Reference block

## Evidence set

- BEFORE captures (d0e54b322 build): live-before-init.txt / live-before-update.txt
  (clean-reinstall branch) / live-before-update2.txt (v3 template-sync surface)
- AFTER captures (1995ccfe0-dirty build): live-after-init.txt / live-after-update.txt /
  live-after-update2.txt
- baseline: test-baseline.txt (small packages full green x9; root package attempt-3
  pre-change transcript = test-baseline-attempt3-30m-timeout.txt, 4 pre-existing FAILs
  recorded; attempts 1-2 preserved as -attempt1-timeout / -attempt2-partial)

## Defect-4 status (approval condition 2)

"Install it yourself" — 0 grep hits in internal/ pkg/ cmd/ on the base tree AND absent
from every before/after capture → **not pinned by capture**. D5 shipped as the skeleton
(updateLedger registry + terminal render; see update_action_block.go header).

## D6 (locale) — measured

- internal/tui/i18n.go: tui.Translate + embedded messages/*.yaml catalogs EXIST.
- init/update emitters: zero tui.Translate uses; raw English strings
  (init.go:614 touches conversation_language only as a wizard config pre-fill).
- Captures: English-only output.
- Decision: the framework exists but wiring the whole init/update surface (catalog
  authoring + CLI language resolution) is not trivial reuse → no i18n introduction in
  this card; recorded for a future card.

## Machine-load context (affects the verification record)

The box ran factory siblings at load 20-48 / 16 cores through this card. Attempts 1-3
of the root-package run hit the 30m binary timeout (attempt 3's binary was compiled
pre-edit, so its transcript IS the pre-change baseline evidence; 4 pre-existing FAILs
recorded, not fixed). The post-change whole-package run launched with -timeout 45m
under the held slot lease. Targeted test families (the tests my diff can affect) all
green: severity/terminal-block/reexec/banner/outcome/worktree-advisory/merge-history +
hook-install families + init/MCP + preview family + race (tui, cli/update/...,
cli targeted) + vet + lint (0 issues) + GOOS=windows build/vet.

## Same-class deferred (recorded, not done)

- migrate_local_instructions shared fan_in-3 "Advisory:" text (REQ-IFU-012 pins ONE
  wording across update/doctor/Codex launcher — a prefix change exceeds this card's
  init/update surface).
- printer.Warn's redundant "Warning: " text prefix (shared by every CLI surface).
- internal/cli/update/backup + update/merge subpackage "advisory:" lines (own writers,
  untouched; their tests still pin the old shape).
