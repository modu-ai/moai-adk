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

The box ran factory siblings at load 20-56 / 16 cores through this card. Attempts 1-3
of the root-package run hit the 30m binary timeout (attempt 3's binary was compiled
pre-edit, so its transcript IS the pre-change baseline evidence; 4 pre-existing FAILs
recorded, not fixed). Attempt 4 (the post-change whole-package run, -timeout 45m under
the held lease) exited 1 after ~44 min with zero output lines written — no panic dump,
no package verdict; recorded as-is, not retried into load 56.

Verification that DID land, all green:
- small packages whole, post-change: tui x3 + cli/update x6 (exit 0, re-run after the
  final edits)
- targeted root-package families covering every test the diff can touch: severity,
  terminal-block, reexec/banner, outcome/classify, worktree-advisory, merge-history,
  hook-install (push+commit+disclosure+attribution+preserve), init/MCP-provision,
  preview family + regenerated golden, skip-sync, retained-advisory, identity,
  worktree-migration, harness, dry-run, e2e, mode, legacy/current snapshot,
  clean-legacy-hooks, agency-adapter, flag-matrix, fast-path, archive-skill,
  clean-install
- go test -race: internal/tui/... + internal/cli/update/... whole, + cli targeted
- go vet (cli, cli/update, config) darwin + GOOS=windows; GOOS=windows build ./...
- golangci-lint run ./internal/cli/... ./internal/config/ → 0 issues
- the repository-wide verdict belongs to the CI run on the integration branch (PENDING
  at report time, per the run-phase contract)

## Same-class deferred (recorded, not done)

- migrate_local_instructions shared fan_in-3 "Advisory:" text (REQ-IFU-012 pins ONE
  wording across update/doctor/Codex launcher — a prefix change exceeds this card's
  init/update surface).
- printer.Warn's redundant "Warning: " text prefix (shared by every CLI surface).
- internal/cli/update/backup + update/merge subpackage "advisory:" lines (own writers,
  untouched; their tests still pin the old shape).

## Repair round (card review: 4xP2 + 2xP3, raw at card-review-codex-raw.txt)

- F1 (P2) MOAI_UPDATE_REEXEC consume-and-clear: reexecPassActive unsets the marker as
  it reads it (update.go; test rewritten as TestReexecPassActive_ConsumeAndClear).
- F2 (P2) remaining raw "warning:" emitters migrated: wireCodexUnlessClaude (init.go)
  and reportUndeployedCodexTemplates (update_codex_wiring.go, · note — retained files
  are advisories).
- F3 (P2) terminal block renders via defer in runUpdate — EVERY exit path carries it
  (clean-reinstall success and version-match skip previously exited before the
  render). Re-captured: live-after-update.txt (clean-reinstall branch) now ends with
  the block.
- F4 (P2) init MCP-provision failure printed 3x (helper line + caller Warn + tail
  panel) → helper returns the error silently, caller Collect()s once, the exit panel
  is the single surface; signature dropped the unused errOut param.
- F5 (P2, adjudicated — reviewer was RIGHT): init_mcp_provision_test.go's
  FailureIsNonFatal asserted stderr contains "warning" while the helper had emitted
  "✗ MCP server entry provisioning failed" since the D4 helper landed. My earlier
  "green" was a -run pattern ('TestInitMCP') that matched NO test — a no-match run
  reports ok. Assertion rewritten to the return-value + no-print contract; the
  family now runs and passes (verified with -v showing the three PASS lines).
- F6 (P3) one surface per failure: caller-level mid-run duplicates removed (binary
  update, re-exec, worktree migration, retired model keys, legacy-skill archive,
  model-key strip, harness re-assert, .gitignore/file merge, global settings env,
  evolution scaffold, profile read/sync); installer-internal immediate lines stay as
  the live feed. Capture check: "worktree migration failed" appears exactly once.
- F7 (P3) init order: flushUpdateNotice moved BEFORE the completion card; the card is
  followed only by the deferred warning summary panel (REQ-TUX2-013 terminal surface,
  which the card's pointer text presupposes). Documented in init.go.

Repair verification: affected families re-run green (provision family -v 3 PASS,
reexec/banner/severity/terminal-block, skip-sync, retained, identity,
worktree-migration, harness, dry-run, e2e, mode, snapshots, hooks, clean-install),
race targeted ok, vet darwin+windows clean, golangci-lint 0 issues, GOOS=windows
build ok. Live-after captures re-taken on the e24ae0298+repair build
(/tmp/t1527-moai-after2).

