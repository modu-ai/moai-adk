# SPEC-UPDATE-REPAIR-RC30-001 — Acceptance Criteria

## Budget, status, and adoption

- Tier L ceiling: 25 acceptance criteria. This file carries **20** criteria (26 minus AC-RC30-01 [RETIRED], -02, -03, -11, -26, which moved to SPEC B under decision d-20261009T235135Z-5e35, and AC-RC30-25 [RETIRED], which moved out of rc.30 under decision d-20261010T040402Z-94b9). Within the ceiling. OPEN LEADER DECISION B1 (decision-index.md) is closed by that split.
- U-22 (owned by card t1578) carries no criterion. U-30 is judged by the locale part of AC-RC30-25 [RETIRED] (REQ-RC30-019).
- Requirements with no active criterion in this file after the AC-RC30-25 [RETIRED] move (ruling d-20261010T040402Z-94b9): REQ-RC30-014, REQ-RC30-016, REQ-RC30-018, REQ-RC30-019 and REQ-RC30-020. Their disposition (moved with the criterion, as d-20261009T235135Z-5e35 moved REQ-RC30-004 and REQ-RC30-021, or kept with a follow-up marker) is an open operator verdict for the leader's next round. This file does not decide it.
- **Tree for every RED cell:** `8fa0dc24b`. Its code paths equal main `2aab5f797`, and the SPEC-only commits after it leave them unchanged. Each RED cell is one single-invocation command (no pipes, redirection, or chaining), its verbatim stdout, its exit status, and the tree SHA.
- **Release gate versus regression-guard** (MP-8):
  - A criterion whose judging command exists on the tree is a release gate only when a RED cell is recorded for it.
  - A criterion that names a test not yet in the tree is judged at its **TDD RED milestone**: the run-stage milestone in which the named test is first run and fails on the absorbed tree. Its RED cell observes the criterion's effect on the current tree (the effect cell). Until its TDD RED milestone, it is a regression-guard, and that status is stated per row.
  - A criterion with no observable effect cell is a regression-guard, and the reason is stated.
- Zero-match selectors are not passes. A `-run` selector that matches no test prints `[no tests to run]`, which is a failure for that criterion.

## RED ledger (run on the current tree in this revision)

Each entry: command (single invocation), verbatim stdout, exit status, tree SHA `8fa0dc24b`.

```text
E-01  command: grep -n "0 user customizations modified" internal/cli/update_archive.go
      stdout:  282://	total: N skills archived, 0 user customizations modified
               367:	_, _ = fmt.Fprintln(out, tui.Pill(tui.PillOpts{Kind: tui.PillOk, Solid: false, Label: fmt.Sprintf("total: %d skills archived, 0 user customizations modified", archived), Theme: &th}))
      exit:    0   (two matches: one comment at 282, one code line at 367)

E-02  command: grep -n "preserved: " internal/cli/update/report/outcome.go
      stdout:  115:		result += "\n  preserved: " + p
      exit:    0

E-03  command: grep -n "Current version" internal/cli/update.go
      stdout:  200:		_, _ = fmt.Fprintln(out, tui.KV("Current version", "moai-adk "+currentVersion, tui.KVOpts{Theme: &th, KeyWidth: 16}))
               835:	// Card t1527 D1: the "Current version" line is dropped here — the top of
      exit:    0

E-04  command: grep -n "os.Remove" internal/cli/migrate_project_assets.go
      stdout:  83:	// outside-pointing symlink must not route os.Remove to an external
               168:			if err := os.Remove(p); err != nil {
      exit:    0

E-05  command: grep -n "kept project-side" internal/cli/migrate_project_assets.go
      stdout:  160:				report("  migration: kept project-side (user counterpart not confirmed — opt in via 'moai bundle add <name>' or re-run update): %s", relSlash)
               200:		report("migration: %d removed, %d preserved (user-modified), %d kept project-side, %d user-created untouched", removed, preserved, stayed, untouched)
      exit:    0

E-06  command: grep -n "WithVersion" internal/cli/update_template_sync.go
      stdout:  456:					template.WithVersion(version.GetVersion()),
               619:					template.WithVersion(version.GetVersion()),
      exit:    0

E-07  command: grep -rl "\.claude/skills/moai/" internal/template/templates/.claude/rules
      stdout:  (a list of rule files, one path per line; first line internal/template/templates/.claude/rules/moai/NOTICE.md; last line internal/template/templates/.claude/rules/moai/workflow/session-handoff-examples.md; count recorded in research.md § 2 E-07-COUNT)
      exit:    0

E-08  command: grep -n "claude/agents/moai" docs-site/content/ko/cli-reference/update.md
      stdout:  (none)
      exit:    1   (grep no-match; the file exists, so the audit line is not reproduced in this tree)

E-09  command: grep -n "tui-snapshot" Makefile
      stdout:  27:.PHONY: all build test lint fix clean install verify-local-install generate templ-generate help release-local constitution-check ci-local pr-merge ci-disable verify-required-checks tui-snapshot tui-snapshot-verify preflight lint-fast test-race-short agents-emit agents-emit-check commands-emit commands-emit-check embed-manifest embed-manifest-check embed-check fmt-check tool-policy-drift-check
               193:tui-snapshot: ## Regenerate all internal/tui golden snapshots (UPDATE_GOLDEN=1)
               197:tui-snapshot-verify: ## Verify all internal/tui golden snapshots match on-disk state (no update)
      exit:    0

E-10  command: grep -n "steps" internal/cli/update_tux.go
      stdout:  134:// "N/M steps" count. It replaces the legacy "N/M steps complete" plain text
               139:	return fmt.Sprintf("  %s %s %d/%d steps", prefix, bar, done, total)
      exit:    0

E-11  command: grep -n "settings.json" internal/cli/update_restore.go
      stdout:  (none)
      exit:    1   (restore code does not reference settings.json in this file)

E-12  command: grep -n "Use:" internal/cli/bundle.go
      stdout:  34:	Use:   "bundle",
               44:	Use:   "add <bundle-name>",
               51:	Use:   "remove <bundle-name>",
               64:	cmd := &cobra.Command{Use: "bundle-test"}
      exit:    0   (no `list` subcommand)

E-13  command: grep -n "bundle add" internal/cli/user_asset_phase.go
      stdout:  3:// `moai init` (the first-install trigger, REQ-024), `moai bundle add|remove`
      exit:    0   (comment only; no announce code)

E-14  command: grep -n "deployment_mode" internal/config/manager.go
      stdout:  (none)
      exit:    1

E-15  command: grep -n "Opus" internal/config/types.go
      stdout:  389:// legacy backward-compat alias fields Opus/Sonnet/Haiku are removed. An
      exit:    0   (comment only; no alias fields)

E-16  command: grep -n "stateManifestMatch" internal/userassets/install.go
      stdout:  573:	stateManifestMatch // current == manifest hash, ≠ shipped → refresh (REQ-008)
               612:	case stateManifestMatch:
               666:		return stateManifestMatch
      exit:    0   (no version comparison is visible in these lines)

E-17  command: grep -n "GetCurrentName" internal/cli/update.go
      stdout:  652:	profileName := profile.GetCurrentName()
      exit:    0

E-18  command: ls -la .moai/logs/config-change-audit.log
      stdout:  (none)
      stderr:  ls: .moai/logs/config-change-audit.log: No such file or directory
      exit:    1

E-19  command: git grep -n "settings.json" WT-update-migration-fixes -- internal/cli/update_integrity_probe.go
      stdout:  WT-update-migration-fixes:internal/cli/update_integrity_probe.go:47://   - .claude/settings.json: the Claude Code settings merge target; must parse as JSON.
               WT-update-migration-fixes:internal/cli/update_integrity_probe.go:52:	{rel: ".claude/settings.json", check: probeJSON},
      exit:    0   (dated reference, observed 2026-10-10 on the branch tip read in this run: WT-update-migration-fixes carries the unconditional entry; main has no such file; absorbed at run start; the absorption-time tip is recorded in progress.md section G (Absorption Record))

E-20  command: grep -n "migrateProjectCommonAssets(" internal/cli/update_template_sync.go
      stdout:  1185:		removed, err := migrateProjectCommonAssets(projectRoot, homeDir, userAssetsInstalled, nil, func(format string, args ...interface{}) {
      exit:    0   (the call is at 1185, not the 1183 the audit cites)

E-21  command: grep -rn "conversation_language" internal/cli/update_tux.go internal/cli/update_archive.go internal/cli/update/report/outcome.go
      stdout:  (none)
      exit:    1   (the update renderer files read no locale setting; the locale part of AC-RC30-25 [RETIRED] is RED)

E-22  command: grep -n "0 user customizations modified" internal/cli/tuxiu_characterization_test.go
      stdout:  201:		for _, want := range []string{"Updated 26 files", "3 skills archived, 0 user customizations modified"} {
      exit:    0   (excluded by the narrowed AC-RC30-09 scope; its expectation is updated in the same change as U-08)

E-23  command: grep -rn "0 user customizations modified" internal/cli --include='*.go' --exclude='*_test.go' --exclude='update_archive.go'
      stdout:  (none)
      exit:    1   (GREEN reachability: no production match outside update_archive.go)

E-24  command: grep -rn "0 user customizations modified" internal/cli --include='*.go' --exclude='*_test.go'
      stdout:  internal/cli/update_archive.go:282://	total: N skills archived, 0 user customizations modified
               internal/cli/update_archive.go:367:	_, _ = fmt.Fprintln(out, tui.Pill(tui.PillOpts{Kind: tui.PillOk, Solid: false, Label: fmt.Sprintf("total: %d skills archived, 0 user customizations modified", archived), Theme: &th}))
      exit:    0   (the RED state of AC-RC30-09; the GREEN state is expected after U-08 at M4: no output, exit 1)
```

Count of RED cells recorded in this revision: 24 ledger entries (E-01 to E-24). Eighteen have exit 0 (E-07 is a file list whose count, 45, is in research.md § 2 E-07-COUNT). Six are no-match or path-absent observations (E-08, E-11, E-14, E-18, E-21, E-23). Each is a single-invocation command. E-24 is the narrowed judging command of AC-RC30-09; E-01 is the earlier single-file form, kept for history.

## Acceptance matrix

Columns: AC | Items | REQ | Judging command | Expected result | RED cell | Status.

| AC | Items | REQ | Judging command | Expected result | RED cell | Status |
|----|-------|-----|-----------------|-----------------|----------|--------|
AC-RC30-01 [RETIRED] moved to SPEC B (split, d-20261009T235135Z-5e35).
AC-RC30-02 [RETIRED] moved to SPEC B (split, d-20261009T235135Z-5e35).
AC-RC30-03 [RETIRED] moved to SPEC B (split, d-20261009T235135Z-5e35).
| AC-RC30-04 | U-03 | REQ-RC30-005 | `go test ./internal/cli/ -run TestBundleAnnounceForPluginMode -count=1 -v` | `--- PASS`. Output carries the exact `moai bundle add <name>` line for each optional pack not installed. | E-13 (`user_asset_phase.go` has only a comment about `bundle add`) | Release gate from its TDD RED milestone (M1). Regression-guard until then. |
| AC-RC30-05 | U-04 | REQ-RC30-006 | `go test ./internal/cli/ -run TestBundleList -count=1 -v` | `--- PASS`. Output names each catalog bundle with its skill and agent counts. | E-12 (no `list` subcommand) | Release gate from its TDD RED milestone (M3). Regression-guard until then. |
| AC-RC30-06 | U-05 | REQ-RC30-007 | `go test ./internal/config/ -run TestConfigSaveRoundTripLLM -count=1 -v` | `--- PASS`. A fixture `llm.yaml` with `deployment_mode`, an unmodelled key, and comments is unchanged by Save except for the intended field. | E-14 (no `deployment_mode` in `manager.go`), E-15 (alias fields absent) | Release gate from its TDD RED milestone (M3). Regression-guard until then. |
| AC-RC30-07 | U-06 | REQ-RC30-008 | `go test ./internal/cli/ -run TestRecordlessArchiveWarning -count=1 -v` | `--- PASS`. A record-less project with a migration archive prints a visible warning. | none: the silent re-stamp path is audit-cited (U-06) and not yet re-read | Regression-guard (reason: no effect cell yet; re-read in M3). |
| AC-RC30-08 | U-07 | REQ-RC30-009, REQ-RC30-010 | `go test ./internal/cli/ -run TestRestoreSettingsAndRetention -count=1 -v` | `--- PASS`. The user-facing restore does not write the live `.claude/settings.json`. The restore output lists `.claude/settings.json` as a not-restored path and tells the user how to copy it by hand. Restore output lists uncovered paths. Retention keeps the five newest runs and removes every directory of an older run in both backup roots. | E-11 (restore code has no `settings.json` reference) | Release gate from its TDD RED milestone (M3). Regression-guard until then. |
| AC-RC30-09 | U-08 | REQ-RC30-011, REQ-RC30-017 | `grep -rn "0 user customizations modified" internal/cli --include='*.go' --exclude='*_test.go'` (production code; the characterization test at line 201 is excluded from the scope and is updated in the same change, E-22) | After the fix: no output, exit 1. Both production lines in update_archive.go (282 and 367) are removed, and the same change removes the literal from the test expectation at `tuxiu_characterization_test.go:201` (E-22 then prints no output). The GREEN reachability check (E-23) is empty outside update_archive.go. | E-24: production matches at update_archive.go lines 282 and 367, exit 0; E-22: excluded test match at line 201; E-23: no other production match | Release gate (RED cell recorded). |
| AC-RC30-10 | U-09 | REQ-RC30-005 | `go test ./internal/cli/ -run TestNonCatalogSkillMessage -count=1 -v` | `--- PASS`. A managed skill with no catalog entry gets a distinct message, not "opt in via moai bundle add". | E-05 (the "opt in via" message at line 160) | Release gate from its TDD RED milestone (M1). Regression-guard until then. |
AC-RC30-11 [RETIRED] moved to SPEC B (split, d-20261009T235135Z-5e35).
| AC-RC30-12 | U-11 | REQ-RC30-012 | `go test ./internal/userassets/ -run TestUserRootDowngradeRefused -count=1 -v` | `--- PASS`. An older binary does not refresh a user root with a newer recorded version. | E-16 (`stateManifestMatch` has no version comparison in the visible lines) | Release gate from its TDD RED milestone (M1, after absorption). Regression-guard until then. |
| AC-RC30-13 | U-12 | REQ-RC30-001 | `go test ./internal/cli/ -run TestKeptOptionalAssetRefresh -count=1 -v` | `--- PASS`. A kept optional agent or `.agents/skills` copy is refreshed in place or gets a notice. | E-05 (the kept message at line 160; no refresh call is visible at that site) | Regression-guard (effect cell is the message only; the refresh absence is not yet shown by grep). |
| AC-RC30-14 | U-13 | REQ-RC30-022 | `go test ./internal/cli/ -run TestSkipDecisionUsesTreeHash -count=1 -v` | `--- PASS`. A same-version run with a different template tree does not skip. | none: the skip lines are audit-cited (`update_template_sync.go:247`, `:1143`) and not yet re-read | Regression-guard. |
| AC-RC30-15 | U-14 | REQ-RC30-023 | `go test ./internal/cli/ -run TestStampSkewWarning -count=1 -v` | `--- PASS`. A stamp that disagrees with the deployed bytes produces a visible warning. | E-06 (two stamp writes at 456 and 619; no comparison call is visible) | Release gate from its TDD RED milestone (M5). Regression-guard until then. |
| AC-RC30-16 | U-15 | REQ-RC30-024 | `go test ./internal/cli/update/merge/ -run TestPermissionsAskPreserved -count=1 -v` (test written first from the U-15 fixture) | `--- PASS` after the cause is named. | none yet: the starting observation is the RED-now fixture (U-15 next step, spec.md § C), not yet run; the cause is inferred, not established | Regression-guard (undecidable starting observation, verification-completeness.md § 2.1). |
| AC-RC30-17 | U-16 | REQ-RC30-015 | `go test ./internal/cli/update/report/ -run TestPreservedLinesCountOnly -count=1 -v` | `--- PASS`. Default output has one count line. `--verbose` lists the paths. | E-02 (one `preserved:` line appended per path at line 115) | Release gate from its TDD RED milestone (M4). Regression-guard until then. |
| AC-RC30-18 | U-17 | REQ-RC30-013 | `go test ./internal/cli/ -run TestUpdateSingleBanner -count=1 -v` | `--- PASS`. Exactly one version line appears on a run. The identity band (REQ-TUXIU-015) is kept. | E-03 (the "Current version" line at 200; the comment at 835) | Release gate from its TDD RED milestone (M1, after absorption). Regression-guard until then. |
| AC-RC30-19 | U-18 | REQ-RC30-011 | `go test ./internal/cli/ -run TestPreviewCountEqualsRunCount -count=1 -v` | `--- PASS`. Preview count equals run count for the same fixture. | none: the 266-against-423 split is audit-cited and not reproduced | Regression-guard. |
| AC-RC30-20 | U-19, U-20 | REQ-RC30-025 | Investigation with a recorded result: `grep -rl "\.claude/skills/moai/" internal/template/templates/.claude/rules` (count), and `grep -n "claude/agents/moai" docs-site/content/ko/cli-reference/update.md` | Each hit is classified in research.md against the § D.1 outcome. The docs line is either found at its real path or recorded as not present. | E-07 (rule file list), E-08 (docs: no match) | GAP-INVESTIGATE. Regression-guard; no release gate. |
| AC-RC30-21 | U-21 | REQ-RC30-002 | `go test ./internal/cli/ -run TestProfileSyncResolution -count=1 -v` | `--- PASS`. Profile sync reads the project ledger, not only the environment. | E-17 (`profile.GetCurrentName()` at 652) | Release gate from its TDD RED milestone (M1, after absorption). Regression-guard until then. |
| AC-RC30-22 | U-23 | REQ-RC30-007 | `go test ./internal/config/ -run TestGateYAMLCommentsSurviveRestore -count=1 -v` (fixture: 12 comment lines) | `--- PASS`. All comment lines survive backup, wipe, deploy, and restore. | none: the loss is audit-cited and not reproduced in this tree | GAP-INVESTIGATE. Regression-guard. |
| AC-RC30-23 | U-24 | REQ-RC30-007 | `go test ./internal/cli/ -run TestLLMYamlRemarshalCallers -count=1 -v` (one case per `ConfigManager.Save` caller) | `--- PASS` for each caller. Each caller that drops a key is named. | E-14 (Save path has no `deployment_mode` reference) | GAP-INVESTIGATE. Regression-guard. |
| AC-RC30-24 | U-25 | REQ-RC30-003 | Investigation: `ls -la .moai/logs/config-change-audit.log` is the audit's path. The log is absent from this tree (E-18). The log is located in the operator profile first, and the result is recorded in research.md with its command and output. | The untraced removal is attributed to a command with evidence, or recorded as unresolved with the missing evidence named. | E-18 (no such file in this tree) | GAP-INVESTIGATE. Regression-guard. |
| AC-RC30-25 [RETIRED] | U-26, U-27, U-28, U-29, U-30 | REQ-RC30-014, REQ-RC30-016, REQ-RC30-017, REQ-RC30-018, REQ-RC30-019, REQ-RC30-020 | `make tui-snapshot-verify` (target exists, E-09) and `go test ./internal/cli/ -run 'TestUpdateOutputGolden\|TestUpdatePlainFallback' -count=1 -v` (TTY golden shows N of M, bar, one ✓/!/✗ line per step, one-screen summary, one conflict count above the summary; plain fallback carries the same counts; locale part: `go test ./internal/cli/ -run TestUpdateLocaleKeyParity -count=1 -v`, which requires the four locale tables ko, en, ja, and zh to define the same keys and the summary to show the same counts in each locale) | `--- PASS` for each. The golden is changed only through `make tui-snapshot`, on a reviewed diff. | E-10 (the step line renderer at 134-139); E-21 (no `conversation_language` reference in the renderer files) | Release gate from its TDD RED milestone (M4). Regression-guard until then. |
AC-RC30-26 [RETIRED] moved to SPEC B; the repair stays with card t1578 (leader correction).

## Mutant probes (verification-completeness § 2)

- AC-RC30-09 (literal): a relocated literal in another file passes a single-file grep. The judge is the repo-wide grep above, and both lines 282 and 367 must go.
- AC-RC30-18 (one banner): a mutant that removes the identity band and keeps `update.go:200` passes "exactly one line". The judge must also assert the band is present (REQ-TUXIU-015).
- AC-RC30-25 [RETIRED] (golden): a mutant that regenerates the golden from the current output passes. Golden updates must be made only through `make tui-snapshot` on a reviewed diff.

## Definition of Done (for the run phase)

- Every criterion has its RED cell (or its stated regression-guard reason) recorded in progress.md § E.2 on the then-current tree.
- Every test-named criterion has its effect cell recorded now (E-01 to E-24) and its TDD RED milestone named in the matrix.
- Each criterion in this SPEC passes at its owning milestone, with the -v output recorded in progress.md, or is recorded in progress.md as blocked or deferred with a named trigger.
- The coverage target of 85% or more per touched package is measured by `go test -cover` on those packages (measurement Gap until M7).
