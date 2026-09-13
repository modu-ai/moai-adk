# SPEC-AUT-PERMMODES-001 — Sync-Phase Audit (card t584)

> Auditor: sync-auditor (independent). Tree: `.claude/worktrees/t584` @ `d34ba9ec4`, branch `WT-autonomy-perm-modes` (verified: `git rev-parse --show-toplevel` → `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t584`, `git branch --show-current` → `WT-autonomy-perm-modes`, working tree clean). Audit date 2026-09-14. Skeptical stance: every claim below carries the command actually run in this audit and its verbatim output; the run-phase self-report was NOT trusted — re-measured where cheap.

## Verdict

**PASS — 8.8 / 10** (harmonic mean of dimension scores; must-pass firewall: Functionality PASS, Security PASS).

| Dimension | Weight | Score | Verdict |
|---|---|---|---|
| Functionality | 40% | 9.5 | PASS |
| Security | 25% | 9.0 | PASS |
| Craft | 20% | 8.0 | PASS |
| Consistency | 15% | 9.0 | PASS |

Harmonic mean = 4 / (1/9.5 + 1/9 + 1/8 + 1/9) = **8.84** → reported 8.8. No dimension falls below its must-pass threshold; no finding is blocking.

## Evidence — what this audit ran itself (baseline-attribution: this run, this tree @ d34ba9ec4)

| # | Claim | Command | Verbatim output |
|---|---|---|---|
| E1 | Affected packages green | `unset MOAI_KANBAN … && go test ./internal/cli/wizard/... ./internal/config/... ./internal/core/project/... -timeout 8m` | `ok github.com/modu-ai/moai-adk/internal/cli/wizard 6.167s` / `ok … internal/config 9.671s` / `ok … internal/config/atomicfile 1.677s` / `ok … internal/config/toolpolicy 2.276s` / `ok … internal/core/project 7.082s` |
| E2 | Targeted init/autonomy selectors green | `go test ./internal/cli/ -run 'TestRunInit_FlagFullyAutonomousWithoutProofDowngrades\|TestRunInit_SemiAutoAndEmptyAreBoundedDelta\|TestApplyAutonomy\|TestAutonomyTier\|TestRunInit_Autonomy' -count=1 -timeout 8m` | `ok github.com/modu-ai/moai-adk/internal/cli 2.897s` |
| E3 | AC-004 mapping | `go test ./internal/config/ -run 'TestTierDefaultMode' -count=1 -v` | `--- PASS: TestTierDefaultMode_Mapping` / `--- PASS: TestTierDefaultMode_InvalidTierSemiAutoDefault` / `ok … internal/config 0.267s` |
| E4 | AC-008 5-named-set (each run explicitly) | `go test …/internal/{core/project,cli,cli/wizard,config} -run 'TestApplyAutonomyTierBundle_FullyAutonomousDowngradedWithoutProof\|TestApplyAutonomyTierBundle_FullyAutonomousWithProofDeploysBypass\|TestRunInit_FlagFullyAutonomousWithoutProofDowngrades\|TestAppendDowngradeAdvisory\|TestAutonomyTierQuestion_FullyAutonomousNotRecommended' -count=1 -v` | `--- PASS` ×5 across 4 packages (`ok` per package); `TestAppendDowngradeAdvisory_CreatesParentDirs` also PASS |
| E5 | AC-011 real-binary help | `./bin/moai init --help` (binary built 2026-09-13 18:35 from this tree; flag string source `internal/cli/init.go:126` unchanged since `1c30ff881`) | `--autonomy-tier          Session permission mode: accept edits on (semi-auto, default), auto mode (automatic), or bypass permissions (fully-autonomous; requires sandbox proof). Writes user-scope defaultMode: acceptEdits for the default` (exit=0) |
| E6 | AC-012 godocs | `go doc -all ./internal/config` / `go doc -all ./internal/core/project` | TierDefaultMode godoc: "semi-auto → "acceptEdits" … automatic → "auto" … fully-autonomous → "bypassPermissions" … An unknown tier maps to "default" (the MOST restrictive mode)"; ApplyAutonomyTierBundle godoc: "ONLY the USER-scope defaultMode="acceptEdits" record is written … every other file stay byte-identical" |
| E7 | Coverage re-measure | `go test -cover ./internal/cli/wizard/ ./internal/config/ ./internal/core/project/ -count=1` | `wizard … coverage: 93.6%` / `config … coverage: 82.0%` / `core/project … coverage: 88.8%` — matches §E.3 exactly |
| E8 | Lint (changed packages) | `golangci-lint run --timeout=3m ./internal/cli/wizard/... ./internal/config/... ./internal/core/project/...` | `0 issues.` exit=0 |
| E9 | Windows cross-build (E2 re-check) | `GOOS=windows GOARCH=amd64 go build ./...` | `WIN_BUILD_OK exit=0` |
| E10 | AC-009 IAM ref | grep on `internal/template/templates/.claude/skills/moai-foundation-cc/reference/claude-code-iam-official.md` | line 30: "accepts six values (official docs, re-verified 2026-09-13)"; six-row table incl. `auto` and `dontAsk`; lines 43-44 name both kill switches; zero hits for "exactly four" |
| E11 | AC-010 translations | grep on `internal/cli/wizard/translations.go` | ko `편집 자동 수락 (권장)` / `자동 모드` / `권한 우회`; ja `編集を自動承認 (推奨)` / `自動モード` / `権限をバイパス`; zh `自动接受编辑 (推荐)` / `自动模式` / `跳过权限检查`; en labels live in `questions.go` (`Accept edits on (Recommended)` / `Auto mode` / `Bypass permissions`); `TestWizardQuestionTranslationCompleteness` green inside E1 |
| E12 | Security gate priority | source + tests | `EffectiveTierWithGates`: `if killSwitchActive \|\| !sandboxProofPresent { return AutonomyTierAutomatic, true }` (kill-switch trumps proof); covered by `TestEffectiveTierWithGates_KillSwitchWinsOverProof` (proof=true, kill=true → downgrade) and `TestTierToggleOptions_KillSwitchDisablesFullyAutonomousAlways` — both green in E1 |
| E13 | AC-005 bounded delta | test body read (`autonomy_bundle_test.go TestApplyAutonomyTierBundle_SemiAutoIsBoundedDelta`) + wiring test `TestRunInit_SemiAutoAndEmptyAreBoundedDelta` (green in E1/E2) | test asserts USER permissions block "EXACTLY one key", `defaultMode == "acceptEdits"`, and full-file byte compare `projectBefore != projectAfter` → error (full-file bytes, per acceptance §D.3) |
| E14 | AC-001/002/003 | source read (`internal/cli/wizard/questions.go:394-407`) | exactly 3 options; values `semi-auto`/`automatic`/`fully-autonomous`; labels `Accept edits on (Recommended)` / `Auto mode` / `Bypass permissions`; `Default: config.AutonomyTierSemiAuto`; Recommended marker only on the first option; bypass never default/recommended |
| E15 | Official-docs cross-check (live) | `curl -sL https://code.claude.com/docs/en/permission-modes` (2026-09-14, this audit) | `disableBypassPermissionsMode … to "disable" in managed settings` (links `/en/managed-settings`); `disableAutoMode` row speaks of "Any settings file sets disableAutoMode" |
| E16 | M3 comment-only claim | `git show 92f2e4c7f -- internal/config/autonomy_tiers.go` | diff touches only the `ResolveEffectiveTier` godoc comment block — behavior nil |
| E17 | Frontmatter coherence | `sed -n 1,20p spec.md` + progress.md §E.4 | `status: completed`, `updated: 2026-09-14`; §E.4 `sync_commit_sha: 07eed817c` present (backfilled by `d34ba9ec4`); CHANGELOG entry sits under `## [Unreleased] → ### Changed` |

## Findings

- **F1 [S3] [optional]** `internal/template/templates/.claude/skills/moai-foundation-cc/reference/claude-code-iam-official.md:43-44` — the refreshed kill-switch note reads "Kill switches (settable to `"disable"` in **any settings file**)" for BOTH switches, but the official page (E15, live fetch this audit) qualifies `disableBypassPermissionsMode` with **managed settings**; only the `disableAutoMode` row speaks of "any settings file". The card's own M1 research (progress.md §E.2 item 4) recorded "settable in managed settings" — so the doc edit overstates the bypass kill-switch scope against both the official page and the card's own research. Security-adjacent reference inaccuracy (an operator could place the bypass kill switch where Claude Code does not enforce it); no AC-009 cell is violated (it requires six values + both switch names + no "exactly four" — all hold), and the Go-side gate reads the env seam, unaffected. Required fix: split the note — `disableAutoMode`: any settings file; `disableBypassPermissionsMode`: managed settings (per official docs), and mirror the fix wherever the note is quoted.
- **F2 [S3] [optional]** `internal/config` package coverage 82.0% < 85% target (E7, re-measured — matches §E.3); baseline at the pre-card tree was NOT measured, so pre-existing-vs-new attribution remains unestablished. Touched functions are covered (`TierDefaultMode` two tests, E3). Required fix (follow-up card): measure the config package baseline at `162b6e…` (pre-M1) once and record the delta attribution; no code change required.
- **F3 [S3] [optional]** `.moai/specs/SPEC-AUT-PERMMODES-001/progress.md:151` — `sync_commit_sha` recorded as short SHA `07eed817c` (the `…` suffix notes the truncation); the t688 precedent recorded the full 40-hex value. era.go H-4 needs only a non-empty value, so the mechanical predicate passes; consistency nit. Required fix: none (historical record); prefer full SHA in future §E.4 slots.
- **F4 [S3] [optional, cross-tree]** docs-site `getting-started/init-wizard` pages (4 locales) still describe the pre-t583 question set. Declared out of scope in the CHANGELOG entry and §E.4 (`separate docs-site tree; reported as a finding, not edited`) — same disposition as SPEC-INIT-QUIET-WIZARD-001. Not verifiable from this worktree (Gaps). Required fix: separate docs card; do NOT fold into this card.
- **F5 [note, closed by re-measurement]** AC-011's evidence binary (`bin/moai`, mtime 18:35) predates the M3 commit (18:39). Closed: the M3 Go delta is comment-only (E16), the flag help string lives in `internal/cli/init.go:126` untouched since `1c30ff881`, and this audit re-ran `./bin/moai init --help` itself (E5) with output matching source verbatim. No action.

## Gaps (explicitly NOT observed by this audit)

1. Full `internal/cli` package suite (`ok 1262.131s` per §E.2) — not re-run (verification load discipline: targeted selectors only, E2). The run-phase full-package export at `.moai/state/verify/t584/t584-green-cli-full.txt` stands un-re-executed; its attribution caveat (compiled before the M3 comment edit) is immaterial per E16 and the config package re-verified green in E1.
2. Interactive TTY wizard end-to-end run — not exercised by a human-path run; AC-001/002/003 verified via source (E14), golden file (diff shows the intended 6-line re-render), and green unit tests.
3. docs-site page currency (F4) — separate tree, not read.
4. `darwin/arm64` full build re-run — not repeated; only the Windows cross-build was re-checked (E9); CI matrix is the integration verdict per repo policy.
5. `TestRunInit_WizardAutonomyTierAppliesBundle` / `TestRunInit_FlagAutonomyTierNonInteractive` bodies were read for existence only; their PASS is attested by package-green runs (E1/E2), not individually quoted.

## Residual-risk

- The `bin/moai` binary this audit probed (E5) was `make build` output from this tree, not the installed `~/go/bin` binary — same caveat the run phase declared; mitigated by E5 output matching source exactly.
- The IAM-reference kill-switch scope inaccuracy (F1) ships inside the embedded template until a follow-up edit lands; users reading the bundled reference could under-enforce the bypass kill switch. Blast radius: documentation only; the wizard/CLI gating code paths are correct (E12).
- config package 82.0% package coverage could hide untested branches in files this card did not touch; touched-surface coverage is complete (E3), so residual is pre-existing debt, not card-introduced (unproven only because the baseline was never measured — F2).

## Disposition (what the lead/lane must do before merge)

1. **None blocking.** The card may merge as-is: 12/12 ACs re-verified green by independent measurement, security gates intact, sync artifacts coherent.
2. Recommended follow-ups (non-blocking, do not gate this merge):
   - F1 — one-file doc fix for the bypass kill-switch scope wording (fast follow-up; touches `internal/template/templates/**`, so it owes `make build` + template-neutrality checklist per CLAUDE.local.md §2.1).
   - F2 — config package coverage baseline measurement (one command at the pre-card tree, recorded in a note).
   - F4 — docs card for the 4-locale `getting-started/init-wizard` pages (already queued by the CHANGELOG note).
   - F3 — convention only: full SHA in future §E.4 slots.

---

Audit method note: per `verification-claim-integrity.md` §2, every score above is attributed to the commands and verbatim outputs in the Evidence table, run in this audit against tree `d34ba9ec4`. The run-phase self-report's claims that were NOT re-measured are listed under Gaps, not carried as PASS.

🗿 MoAI
