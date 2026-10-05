# Card Verdict — t586 (SPEC-INIT-TUX-I18N-001)

Lane: lane-2 (Factory) · Date: 2026-09-13 · Branch: `WT-init-tux-i18n` · HEAD: `db6526daf` · Unpushed: 61 commits over merge-base `03a48b0df` (local develop tip; develop == merge-base, so both `03a48b0df..HEAD` and `develop..HEAD` count 61)

## Claim

Card t586 is COMPLETE through all three phases: plan (audited), run (M1–M8, AC 22/22 PASS), sync (close commit + independent audit PASS 0.92). The branch is ready for the lead's single integration window. Push is the lead's batch act; the lane has not pushed.

## Evidence

### Phase closures (all observed on this tree, this run)

| Phase | Terminal commit | What closed it |
|---|---|---|
| plan | `268cffe2cb` (v0.2.3, scoped iter4 PASS `f6f3f723b`) | Tier L plan audit PASS; §E.1 |
| run | `e561b162e` (M8; M1 `4d985fe48` … M7 `bfaeaf5ee`) | AC 22/22 PASS, §E.3; final gate pty suite exit 0, residual FAIL 0 |
| plan-artifact sync duties | `33d81b8ff` (manager-spec) | research §14 command fix (re-run restores 11 lines, seam hit `init.go:858`); spec D4 ground corrected (80×30 pty re-measure shows `● ○ ○ ○ 1 / 4` visible, frame line 14); design §9 allowlist refreshed to 4 measured lines |
| sync | `533cf5fa4` (manager-docs) | CHANGELOG entry (grep = 1); spec.md `in-progress → completed` (status+updated only); progress.md §E.4 filled; MX:TODO scan of card files = 0 |
| F1 backfill | `db6526daf` (manager-docs) | `sync_commit_sha: 533cf5fa4` replaces the sanctioned placeholder |

### Independent sync-audit

- Verdict **PASS 0.92** (Tier L threshold 0.85). Report: `.moai/reports/t586/sync-audit.md`.
- Scores: Functionality 95 / Security 90 / Craft 90 / Consistency 92 (harmonic 0.92).
- Skeptical sampling: 8 of 22 ACs re-executed green (AC-ITI-008/013/014/018/021/022 + 005 + mechanical guards), run counts == selector counts everywhere; fresh lint `./internal/cli/...` 0 issues; fresh cover `internal/cli/wizard` 93.6%.

### Key evidence paths (untracked, under this tree)

- `.moai/reports/t586/sync-audit.md` — independent audit
- `.moai/reports/t586/sync-research14-recheck.txt` — corrected §14 derivation (11 lines)
- `.moai/reports/t586/sync-d4-recheck.txt` + `sync-d4-recheck-frame/init-first-screen.txt` — D4 re-measurement
- `.moai/reports/t586/m8-pty-capture-suite.txt` — run-phase final gate capture
- Run-phase captures: `m5-*`, `m6-*`, `m7-*`, `ac003-*`, `absorb-t583/slot/`

## Baseline-attribution

Every row above was measured in this run against this tree (worktree `.claude/worktrees/t586`, HEAD `db6526daf`, base `03a48b0df`). Run-phase figures carry their own §E.2 attributions; the audit's fresh measurements are pinned in `sync-audit.md`.

## Gaps

- 14 of 22 ACs were NOT re-executed at sync-audit (sample 8); they rely on run-phase §E.2 evidence (exit-0 full-package suites) — accepted because CI on `origin/develop` is the integration verdict surface.
- No local full-suite run at sync (constraint; CI's job).
- audit_multi not called (git-flow topology would resolve `baseBranch` against the wrong base — fail-open).
- §E.1's plan-audit history cites `init.go:873`-era coordinates pinned to tree `538b56f19`; current-tree coordinate is 858 (corrected in research §14; the §14 step-2 table's other coordinates remain 538b56f19-pinned records).

## Residual-risk (operator-report items, carried in §E.4(b) — not defects in this SPEC)

1. **REQ-ITI-007 option-label locale fixation** — in-form language change re-renders title/description/help but option labels stay at launch locale; en-fixed: 2 model_policy label texts + "(runtime default)"/"(project default)" schema literals. Strict reading = one follow-up translation card.
2. **AC-ITI-011 (3) axis-golden TERM sensitivity** — goldens embed truecolor ANSI; a different color-profile environment can mismatch. Documented; CI pty runs should pin TERM.
3. **@MX tag coverage partial** on some card-authored files (advisory, audit F2).
4. **Conflict-check candidate at the integration window**: `internal/config/manager.go` (ConfigManager.Save git_convention isolation rode inside the M5 absorb commit — hunk-level revert isolation only) and any file other lanes touched in the wizard/i18n area.

## Window procedure (on lead designation)

acquire → `git merge origin/develop` (absorb) → re-measure in merge tree (scoped: `./internal/cli/...` suites + the 8 sampled ACs' selectors; verify golden/pty tests under the merge tree's TERM) → `EnterWorktree(.claude/worktrees/develop)` → `git merge --no-ff WT-init-tux-i18n` → release → report merge SHA. Push stays lead-batch. Worktree disposal only after the remote merge lands.

## Window execution record (2026-09-13, lane-2)

- Window acquired: `moai integration acquire --name lane-2 --card t586` (holder bba81b70, since 08:32Z).
- Correction adopted: absorb target was local develop `4f9025151` (lead's figure; my merge-base `03a48b0df` was 145 commits behind — re-measured `git rev-parse develop` = `4f9025151` = `origin/develop`).
- Absorb merge: `0c369e170` (parents `db6526daf` + `4f9025151`).
- **Conflicts (2) and resolutions:**
  1. `CHANGELOG.md` — union: t586 close entry + develop's SPEC-DOCS-TABCOUNT-DRIFT-001 (t530) entry both kept under `[Unreleased] ### Changed`.
  2. `internal/cli/profile_setup.go` — develop's SPEC-WORKTREE-KEY-WIRING-001 M2 ordering (`sessionExitAutoMerge` BEFORE disposal, REQ-WKW-013) combined with t586 M5's `cleanupSessionWorktreeFn` test seam (AC-ITI-006/007 preservation; seam var still binds to `cleanupSessionWorktree` at `profile.go:83`). `init.go` auto-merged (same pattern resolved by git).
- Post-resolution: `go build ./internal/cli/...` OK, `go vet ./internal/cli` exit 0, conflict markers 0.
- **Merge-tree re-measurement (all exit 0, TERM=xterm-256color pinned):**
  - Sampled AC selectors (wizard pkg): `TestProfileWizardGolden_LocaleFrames|NoEnglishLeak|TestWizardTranslationKeys_ReferencedAndLocaleEquivalent|TestInitRegroup_TwoPages|TestInitStepper_Denominator4|TestGroupLabel_NotRendered` — RUN 4 = top-level PASS 4, FAIL 0.
  - cli pkg (absorbed/seam/t655): `TestProfileSetupAbsorbed|TestHelpLabels_LocaleGoldenSurfaces|TestAutoMerge|TestSessionWorktree|TestSessionExit` — PASS 26, FAIL 0 (61 RUN incl. subtests).
  - Gated pty (`MOAI_PTY_CAPTURE=1`): `TestPtyCapture_InitFirstScreen|TestPtyCapture_DowngradeConfirm*` — PASS 3, FAIL/SKIP 0.
  - Full `./internal/cli/...` suite deliberately NOT run locally (lead directive; CI on origin/develop is the integration verdict).
- Verdict update committed as `f165ccd8e`.
