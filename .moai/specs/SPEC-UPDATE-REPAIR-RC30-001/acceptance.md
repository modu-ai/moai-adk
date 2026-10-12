# SPEC-UPDATE-REPAIR-RC30-001 — Acceptance Criteria

## Budget, status, and adoption

- Tier L ceiling: 25 acceptance criteria. This file carries **22** criteria: 21 active criteria of revision 0.1.3, plus AC-RC30-28 (new in revision 0.1.4). Retired and moved criteria: AC-RC30-01 [RETIRED], -02, -03 and -11 moved to card t1668 (ruling d-20261009T235135Z-5e35); AC-RC30-26 [RETIRED] moved to card t1668 (ruling d-20261010T033220Z-3c5f); AC-RC30-25 [RETIRED] moved to card t1669 (ruling d-20261010T040402Z-94b9). AC-RC30-27 is new for U-01 (decision d-20261010T084752Z-d4be). Within the ceiling. OPEN LEADER DECISION B1 (decision-index.md) is closed by that split.
- U-22 (owned by card t1578) carries no criterion. U-30 [MOVED to card t1669] was judged by the locale part of AC-RC30-25 [RETIRED] (REQ-RC30-019).
- Requirements with no active criterion in this file, each with its owner card: REQ-RC30-003 (card t1668; AC-RC30-24 here is a cross-reference only), REQ-RC30-004 (card t1668; AC-RC30-03 [REF] and AC-RC30-11 [REF] moved), REQ-RC30-021 (card t1668; AC-RC30-26 [REF] moved), and REQ-RC30-014, -016, -018, -019 and -020 (card t1669; AC-RC30-25 [REF] moved). REQ-RC30-017 has an active row (AC-RC30-09) that judges its literal clause only; its count clause is judged by AC-RC30-25 [REF] in card t1669 (Gap G-17).
- **Tree for every RED cell:** `6f95917a1` (revision 0.1.4; both branches are absorbed there, merges `ff2bc876d` and `6f95917a1`). The revision 0.1.3 cells were pinned to `8fa0dc24b`; the drift from that pin is recorded under the RED ledger. Each RED cell is one single-invocation command (no pipes, redirection, or chaining), its verbatim stdout, its exit status, and the tree SHA.
- **Release gate versus regression-guard** (MP-8):
  - A criterion whose judging command exists on the tree is a release gate only when a RED cell is recorded for it.
  - A criterion that names a test not yet in the tree is judged at its **TDD RED milestone**: the run-stage milestone in which the named test is first run and fails on the absorbed tree. Its RED cell observes the criterion's effect on the current tree (the effect cell). Until its TDD RED milestone, it is a regression-guard, and that status is stated per row.
  - A criterion with no observable effect cell is a regression-guard, and the reason is stated.
- Zero-match selectors are not passes. A `-run` selector that matches no test prints `[no tests to run]`, which is a failure for that criterion.

## RED ledger (re-executed on 6f95917a1 in revision 0.1.4)

Each entry: command (single invocation), verbatim stdout, exit status, tree SHA `6f95917a1`. Each command was run alone from the worktree root on `6f95917a1` in revision 0.1.4; the exit status is observed from that run and recorded as its own field, not inside the command. The revision 0.1.3 line numbers (tree `8fa0dc24b`) are kept in the drift list below.

```text
E-01  command: grep -n "0 user customizations modified" internal/cli/update_archive.go
      stdout:  282://	total: N skills archived, 0 user customizations modified
               367:	_, _ = fmt.Fprintln(out, tui.Pill(tui.PillOpts{Kind: tui.PillOk, Solid: false, Label: fmt.Sprintf("total: %d skills archived, 0 user customizations modified", archived), Theme: &th}))
      exit:    0   (two matches: one comment at 282, one code line at 367; unchanged from 8fa0dc24b)

E-02  command: grep -n "preserved: " internal/cli/update/report/outcome.go
      stdout:  115:		result += "\n  preserved: " + p
      exit:    0   (unchanged from 8fa0dc24b)

E-03  command: grep -n "Current version" internal/cli/update.go
      stdout:  200:		_, _ = fmt.Fprintln(out, tui.KV("Current version", "moai-adk "+currentVersion, tui.KVOpts{Theme: &th, KeyWidth: 16}))
               844:	// Card t1527 D1: the "Current version" line is dropped here — the top of
      exit:    0   (drift: comment line 835 at 8fa0dc24b, 844 at 6f95917a1)

E-04  command: grep -n "os.Remove" internal/cli/migrate_project_assets.go
      stdout:  83:	// outside-pointing symlink must not route os.Remove to an external
               174:			if err := os.Remove(p); err != nil {
      exit:    0   (drift: removal call 168 at 8fa0dc24b, 174 at 6f95917a1; the call is not behind a switch, see spec.md § D.2)

E-05  command: grep -n "kept project-side" internal/cli/migrate_project_assets.go
      stdout:  166:				report("  migration: kept project-side (user counterpart not confirmed — opt in via 'moai bundle add <name>' or re-run update): %s", relSlash)
               206:		report("migration: %d removed, %d preserved (user-modified), %d kept project-side, %d user-created untouched", removed, preserved, stayed, untouched)
      exit:    0   (drift: 160 at 8fa0dc24b, 166 at 6f95917a1; 200 at 8fa0dc24b, 206 at 6f95917a1)

E-06  command: grep -n "WithVersion" internal/cli/update_template_sync.go
      stdout:  456:					template.WithVersion(version.GetVersion()),
               619:					template.WithVersion(version.GetVersion()),
      exit:    0   (unchanged from 8fa0dc24b)

E-07  command: grep -rl "\.claude/skills/moai/" internal/template/templates/.claude/rules
      stdout:  (45 lines, verbatim, in directory-traversal order of this filesystem)
               internal/template/templates/.claude/rules/moai/NOTICE.md
               internal/template/templates/.claude/rules/moai/design/constitution.md
               internal/template/templates/.claude/rules/moai/core/askuser-protocol-reference.md
               internal/template/templates/.claude/rules/moai/core/moai-mcp-tools-catalogue.md
               internal/template/templates/.claude/rules/moai/core/agent-common-protocol-reference.md
               internal/template/templates/.claude/rules/moai/core/security-decision-contract.md
               internal/template/templates/.claude/rules/moai/core/moai-constitution-detail.md
               internal/template/templates/.claude/rules/moai/development/agent-authoring.md
               internal/template/templates/.claude/rules/moai/development/karpathy-quickref.md
               internal/template/templates/.claude/rules/moai/development/spec-frontmatter-schema.md
               internal/template/templates/.claude/rules/moai/development/skill-authoring.md
               internal/template/templates/.claude/rules/moai/development/manager-develop-prompt-template.md
               internal/template/templates/.claude/rules/moai/development/coding-standards.md
               internal/template/templates/.claude/rules/moai/development/branch-origin-protocol.md
               internal/template/templates/.claude/rules/moai/workflow/context-clear-policy.md
               internal/template/templates/.claude/rules/moai/workflow/verification-plan-contract.md
               internal/template/templates/.claude/rules/moai/workflow/orchestration-mode-selection.md
               internal/template/templates/.claude/rules/moai/workflow/goal-directive-detail.md
               internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-mechanics.md
               internal/template/templates/.claude/rules/moai/workflow/tdd-result-contract.md
               internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-cards.md
               internal/template/templates/.claude/rules/moai/workflow/archived-agent-rejection.md
               internal/template/templates/.claude/rules/moai/workflow/language-routing-contract.md
               internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md
               internal/template/templates/.claude/rules/moai/workflow/trace-ledger-contract.md
               internal/template/templates/.claude/rules/moai/workflow/cadence-bridge.md
               internal/template/templates/.claude/rules/moai/workflow/mx-tag-protocol.md
               internal/template/templates/.claude/rules/moai/workflow/skill-routing.md
               internal/template/templates/.claude/rules/moai/workflow/goal-directive.md
               internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-detail.md
               internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-gates.md
               internal/template/templates/.claude/rules/moai/workflow/phase-id-contract.md
               internal/template/templates/.claude/rules/moai/workflow/team-capability-resolver.md
               internal/template/templates/.claude/rules/moai/workflow/read-only-status-contract.md
               internal/template/templates/.claude/rules/moai/workflow/resource-budget-contract.md
               internal/template/templates/.claude/rules/moai/workflow/snapshot-consumer-contract.md
               internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md
               internal/template/templates/.claude/rules/moai/workflow/delivery-policy.md
               internal/template/templates/.claude/rules/moai/workflow/sync-backup-integrity.md
               internal/template/templates/.claude/rules/moai/workflow/sync-coverage-scope-contract.md
               internal/template/templates/.claude/rules/moai/workflow/native-invocation-model.md
               internal/template/templates/.claude/rules/moai/workflow/graceful-exit-mutation-contract.md
               internal/template/templates/.claude/rules/moai/workflow/verification-batch-pattern.md
               internal/template/templates/.claude/rules/moai/workflow/session-handoff-examples.md
               internal/template/templates/.claude/rules/moai/workflow/cross-session-messaging.md
      exit:    0   (count 45, matching research.md § 2 E-07-COUNT; the revision 0.1.3 ledger recorded only a summary, and its last line `session-handoff-examples.md` is now line 44 of 45, the list being in traversal order)

E-08  command: grep -n "claude/agents/moai" docs-site/content/ko/cli-reference/update.md
      stdout:  (none)
      exit:    1   (grep no-match; the file exists, so the audit line is not reproduced in this tree)

E-09  command: grep -n "tui-snapshot" Makefile
      stdout:  27:.PHONY: all build test lint fix clean install verify-local-install generate templ-generate help release-local constitution-check ci-local pr-merge ci-disable verify-required-checks tui-snapshot tui-snapshot-verify preflight lint-fast test-race-short agents-emit agents-emit-check commands-emit commands-emit-check embed-manifest embed-manifest-check embed-check fmt-check tool-policy-drift-check
               193:tui-snapshot: ## Regenerate all internal/tui golden snapshots (UPDATE_GOLDEN=1)
               197:tui-snapshot-verify: ## Verify all internal/tui golden snapshots match on-disk state (no update)
      exit:    0   (unchanged from 8fa0dc24b)

E-10  command: grep -n "steps" internal/cli/update_tux.go
      stdout:  134:// "N/M steps" count. It replaces the legacy "N/M steps complete" plain text
               139:	return fmt.Sprintf("  %s %s %d/%d steps", prefix, bar, done, total)
      exit:    0   (unchanged from 8fa0dc24b)

E-11  command: grep -n "settings.json" internal/cli/update_restore.go
      stdout:  (none)
      exit:    1   (restore code does not reference settings.json in this file)

E-12  command: grep -n "Use:" internal/cli/bundle.go
      stdout:  34:	Use:   "bundle",
               44:	Use:   "add <bundle-name>",
               51:	Use:   "remove <bundle-name>",
               64:	cmd := &cobra.Command{Use: "bundle-test"}
      exit:    0   (no `list` subcommand; unchanged from 8fa0dc24b)

E-13  command: grep -n "bundle add" internal/cli/user_asset_phase.go
      stdout:  3:// `moai init` (the first-install trigger, REQ-024), `moai bundle add|remove`
      exit:    0   (comment only; no announce code; unchanged from 8fa0dc24b)

E-14  command: grep -n "deployment_mode" internal/config/manager.go
      stdout:  (none)
      exit:    1

E-15  command: grep -n "Opus" internal/config/types.go
      stdout:  389:// legacy backward-compat alias fields Opus/Sonnet/Haiku are removed. An
      exit:    0   (comment only; no alias fields; unchanged from 8fa0dc24b)

E-16  command: grep -n "stateManifestMatch" internal/userassets/install.go
      stdout:  734:	stateManifestMatch // current == manifest hash, ≠ shipped → refresh (REQ-008)
               789:	case stateManifestMatch:
               845:		return stateManifestMatch
      exit:    0   (no version comparison is visible in these lines; drift: 573 to 734, 612 to 789, 666 to 845)

E-17  command: grep -n "GetCurrentName" internal/cli/update.go
      stdout:  661:	profileName := profile.GetCurrentName()
      exit:    0   (drift: 652 at 8fa0dc24b, 661 at 6f95917a1)

E-18  command: ls -la .moai/logs/config-change-audit.log
      stdout:  (none)
      stderr:  ls: .moai/logs/config-change-audit.log: No such file or directory
      exit:    1

E-19  command: git grep -n "settings.json" e73a7cbf5 -- internal/cli/update_integrity_probe.go
      stdout:  e73a7cbf5:internal/cli/update_integrity_probe.go:60://   - .claude/settings.json: the Claude Code settings merge target; must parse as
               e73a7cbf5:internal/cli/update_integrity_probe.go:70:	{rel: ".claude/settings.json", check: probeJSON, claudeOnly: true},
      exit:    0   (re-pinned in revision 0.1.4 from the moving branch WT-update-migration-fixes to its absorbed tip, the commit e73a7cbf5. Revision 0.1.3 read lines 47 and 52 of the moving branch; they are now lines 60 and 70, and line 70 carries `claudeOnly: true`, so the entry is conditional on the Claude harness and is no longer described as unconditional)

E-20  command: grep -n "migrateProjectCommonAssets(" internal/cli/update_template_sync.go
      stdout:  1185:		removed, err := migrateProjectCommonAssets(projectRoot, homeDir, userAssetsInstalled, nil, func(format string, args ...interface{}) {
      exit:    0   (unchanged from 8fa0dc24b)

E-21  command: grep -rn "conversation_language" internal/cli/update_tux.go internal/cli/update_archive.go internal/cli/update/report/outcome.go
      stdout:  (none)
      exit:    1   (the update renderer files read no locale setting; the locale part of AC-RC30-25 [RETIRED] [MOVED to card t1669] is RED)

E-22  command: grep -n "0 user customizations modified" internal/cli/tuxiu_characterization_test.go
      stdout:  201:		for _, want := range []string{"Updated 26 files", "3 skills archived, 0 user customizations modified"} {
      exit:    0   (excluded by the narrowed AC-RC30-09 scope; its expectation is updated in the same change as U-08; unchanged from 8fa0dc24b)

E-23  command: grep -rn "0 user customizations modified" internal/cli --include='*.go' --exclude='*_test.go' --exclude='update_archive.go'
      stdout:  (none)
      exit:    1   (GREEN reachability: no production match outside update_archive.go)

E-24  command: grep -rn "0 user customizations modified" internal/cli --include='*.go' --exclude='*_test.go'
      stdout:  internal/cli/update_archive.go:282://	total: N skills archived, 0 user customizations modified
               internal/cli/update_archive.go:367:	_, _ = fmt.Fprintln(out, tui.Pill(tui.PillOpts{Kind: tui.PillOk, Solid: false, Label: fmt.Sprintf("total: %d skills archived, 0 user customizations modified", archived), Theme: &th}))
      exit:    0   (the RED state of AC-RC30-09; the GREEN state is expected after U-08 at M4: no output, exit 1)

E-25  command: go test ./internal/cli/ -run '^TestMigrationKeepsSharedCopyUnderProfileReach$' -count=1 -v
      stdout:  (build failure, 14 lines, observed by the lane at tree 6f95917a1 with the test file untracked; the verbatim lines are in .moai/reports/t1619/progress.md, the block that opens "RED 관측 (트리 6f95917a1 + 테스트 파일 미추적)")
               # github.com/modu-ai/moai-adk/internal/cli [github.com/modu-ai/moai-adk/internal/cli.test]
               internal/cli/migrate_profile_reach_test.go:97:11: after.KeptBy undefined (type *manifest.FileEntry has no field or method KeptBy)
               internal/cli/migrate_profile_reach_test.go:97:30: undefined: manifest.KeptByProfileReach
               internal/cli/migrate_profile_reach_test.go:98:55: after.KeptBy undefined (type *manifest.FileEntry has no field or method KeptBy)
               internal/cli/migrate_profile_reach_test.go:98:72: undefined: manifest.KeptByProfileReach
               internal/cli/migrate_profile_reach_test.go:124:63: entry.KeptBy undefined (type *manifest.FileEntry has no field or method KeptBy)
               internal/cli/migrate_profile_reach_test.go:125:58: entry.KeptBy undefined (type *manifest.FileEntry has no field or method KeptBy)
               internal/cli/migrate_profile_reach_test.go:151:62: entry.KeptBy undefined (type *manifest.FileEntry has no field or method KeptBy)
               internal/cli/migrate_profile_reach_test.go:152:90: entry.KeptBy undefined (type *manifest.FileEntry has no field or method KeptBy)
               internal/cli/migrate_profile_reach_test.go:229:3: unknown field KeptBy in struct literal of type manifest.FileEntry
               internal/cli/migrate_profile_reach_test.go:229:26: undefined: manifest.KeptByProfileReach
               internal/cli/migrate_profile_reach_test.go:229:3: too many errors
               FAIL	github.com/modu-ai/moai-adk/internal/cli [build failed]
               FAIL
      exit:    1   (BUILD-RED, not a behavioral RED: `FileEntry.KeptBy` and `manifest.KeptByProfileReach` do not exist on this tree yet. This observation is carried from the lane record, not re-run in revision 0.1.4, which is a text-only revision. The behavioral RED is recorded in M2 after the API stub, plan § B.3 step 6)
```

Count of RED cells recorded in revision 0.1.4: 25 ledger entries (E-01 to E-25). Eighteen have exit 0 (E-01 to E-07, E-09, E-10, E-12, E-13, E-15, E-16, E-17, E-19, E-20, E-22, E-24). Six are no-match or path-absent observations (E-08, E-11, E-14, E-18, E-21, E-23). E-25 is a build-red observation carried from the lane record (exit 1). E-01 is the earlier single-file form of the AC-RC30-09 judging command, kept for history; E-24 is the judging command itself.

Drift from the revision 0.1.3 pin `8fa0dc24b` (line numbers; the stdout text is unchanged unless stated):
- E-03: `update.go` comment line 835 → 844. Line 200 unchanged.
- E-04: `migrate_project_assets.go` removal call line 168 → 174. Line 83 unchanged.
- E-05: `migrate_project_assets.go` lines 160 → 166 and 200 → 206.
- E-07: count 45 unchanged; the last line changed from `…/workflow/session-handoff-examples.md` to `…/workflow/cross-session-messaging.md` (directory-traversal order on this filesystem). The full list is now recorded.
- E-16: `install.go` lines 573 → 734, 612 → 789, 666 → 845.
- E-17: `update.go` line 652 → 661.
- E-19: re-pinned from the moving branch `WT-update-migration-fixes` to its absorbed tip `e73a7cbf5`. Lines 47 → 60 and 52 → 70. Line 70 now reads `claudeOnly: true`.
- Unchanged: E-01, E-02, E-06, E-09, E-10, E-12, E-13, E-15, E-20, E-22, E-24, and the exit status of all 24 revision 0.1.3 entries.

## Acceptance matrix

Columns: AC | Items | REQ | Judging command | Expected result | RED cell | Status.

| AC | Items | REQ | Judging command | Expected result | RED cell | Status |
|----|-------|-----|-----------------|-----------------|----------|--------|
AC-RC30-01 [RETIRED] moved to card t1668 (split, d-20261009T235135Z-5e35).
AC-RC30-02 [RETIRED] moved to card t1668 (split, d-20261009T235135Z-5e35).
AC-RC30-03 [RETIRED] moved to card t1668 (split, d-20261009T235135Z-5e35).
| AC-RC30-04 | U-03 | REQ-RC30-005 | `go test ./internal/cli/ -run TestBundleAnnounceForPluginMode -count=1 -v` | `--- PASS`. Output carries the exact `moai bundle add <name>` line for each optional pack not installed. | E-13 (`user_asset_phase.go` has only a comment about `bundle add`) | Release gate from its TDD RED milestone (M1). Regression-guard until then. |
| AC-RC30-05 | U-04 | REQ-RC30-006 | `go test ./internal/cli/ -run TestBundleList -count=1 -v` | `--- PASS`. Output names each catalog bundle with its skill and agent counts, and `moai bundle add --help` names the valid bundles. | E-12 (no `list` subcommand) | Release gate from its TDD RED milestone (M3). Regression-guard until then. |
| AC-RC30-06 | U-05 | REQ-RC30-007 | `go test ./internal/config/ -run TestConfigSaveRoundTripLLM -count=1 -v` | `--- PASS`. A fixture `llm.yaml` with `deployment_mode`, an unmodelled key, and comments is unchanged by Save, except for the single field the test sets before Save. That field is not named in this SPEC text (Gap G-06): the run-stage test names it at M3, and the criterion does not pass until it is named. The production writer `saveLLMSection` is covered by AC-RC30-23, not by this row. | E-14 (no `deployment_mode` in `manager.go`), E-15 (alias fields absent) | Release gate from its TDD RED milestone (M3). Regression-guard until then. |
| AC-RC30-07 | U-06 | REQ-RC30-008 | `go test ./internal/cli/ -run TestRecordlessArchiveWarning -count=1 -v` | `--- PASS`. A record-less project with a migration archive prints a visible warning. | none: the silent re-stamp path is audit-cited (U-06) and not yet re-read | Regression-guard (reason: no effect cell yet; re-read in M3). |
| AC-RC30-08 | U-07 | REQ-RC30-009, REQ-RC30-010 | `go test ./internal/cli/ -run TestRestoreSettingsAndRetention -count=1 -v` | `--- PASS`. The user-facing restore does not write the live `.claude/settings.json`. The restore output lists `.claude/settings.json` as a not-restored path and tells the user how to copy it by hand. Restore output lists uncovered paths. Retention keeps the five newest runs and removes every directory of an older run in both backup roots. | E-11 (restore code has no `settings.json` reference) | Release gate from its TDD RED milestone (M3). Regression-guard until then. |
| AC-RC30-09 | U-08 | REQ-RC30-011, REQ-RC30-017 | `grep -rn "0 user customizations modified" internal/cli --include='*.go' --exclude='*_test.go'` (production code; the characterization test at line 201 is excluded from the scope and is updated in the same change, E-22) | After the fix: no output, exit 1. Both production lines in update_archive.go (282 and 367) are removed, and the same change removes the literal from the test expectation at `tuxiu_characterization_test.go:201` (E-22 then prints no output). The GREEN reachability check (E-23) is empty outside update_archive.go. | E-24: production matches at update_archive.go lines 282 and 367, exit 0; E-22: excluded test match at line 201; E-23: no other production match | Release gate (RED cell recorded). Judges the literal string only. The count clause of REQ-RC30-017 is judged by AC-RC30-25 [REF] in card t1669 (Gap G-17). |
| AC-RC30-10 | U-09 | REQ-RC30-005 | `go test ./internal/cli/ -run TestNonCatalogSkillMessage -count=1 -v` | `--- PASS`. A managed skill with no catalog entry gets a distinct message, not "opt in via moai bundle add". | E-05 (the "opt in via" message at line 166 at `6f95917a1`; 160 at `8fa0dc24b`) | Release gate from its TDD RED milestone (M1). Regression-guard until then. |
AC-RC30-11 [RETIRED] moved to card t1668 (split, d-20261009T235135Z-5e35).
| AC-RC30-12 | U-11 | REQ-RC30-012 | `go test ./internal/userassets/ -run TestUserRootDowngradeRefused -count=1 -v` | `--- PASS`. An older binary does not refresh a user root with a newer recorded version. | E-16 (`stateManifestMatch` has no version comparison in the visible lines) | Release gate from its TDD RED milestone (M1, after absorption). Regression-guard until then. |
| AC-RC30-13 | U-12 | REQ-RC30-001 | `go test ./internal/cli/ -run TestKeptOptionalAssetRefresh -count=1 -v` | `--- PASS`. A kept optional agent or `.agents/skills` copy is refreshed in place or gets a notice. | E-05 (the kept message at line 166; no refresh call is visible at that site) | Regression-guard (effect cell is the message only; the refresh absence is not yet shown by grep). |
| AC-RC30-14 | U-13 | REQ-RC30-022 | `go test ./internal/cli/ -run TestSkipDecisionUsesTreeHash -count=1 -v` | `--- PASS`. A same-version run with a different template tree does not skip. | none: the skip lines are audit-cited (`update_template_sync.go:247`, `:1143`) and not yet re-read | Regression-guard. |
| AC-RC30-15 | U-14 | REQ-RC30-023 | `go test ./internal/cli/ -run TestStampSkewWarning -count=1 -v` | `--- PASS`. A stamp that disagrees with the deployed bytes produces a visible warning. | E-06 (two stamp writes at 456 and 619; no comparison call is visible) | Release gate from its TDD RED milestone (M5). Regression-guard until then. |
| AC-RC30-16 | U-15 | REQ-RC30-024 | `go test ./internal/cli/update/merge/ -run TestPermissionsAskPreserved -count=1 -v` (test written first from the U-15 fixture) | `--- PASS` after the cause is named. | none yet: the starting observation is the RED-now fixture (U-15 next step, spec.md § C), not yet run; the cause is inferred, not established | Regression-guard (undecidable starting observation, verification-completeness.md § 2.1). |
| AC-RC30-17 | U-16 | REQ-RC30-015 | `go test ./internal/cli/update/report/ -run TestPreservedLinesCountOnly -count=1 -v` | `--- PASS`. Default output prints per-category counts in place of the file list (one count line per category, REQ-RC30-015). `--verbose` prints the full file list. | E-02 (one `preserved:` line appended per path at line 115) | Release gate from its TDD RED milestone (M4). Regression-guard until then. |
| AC-RC30-18 | U-17 | REQ-RC30-013 | `go test ./internal/cli/ -run TestUpdateSingleBanner -count=1 -v` | `--- PASS`. Exactly one version line appears on a run. The identity band (REQ-TUXIU-015) is kept. | E-03 (the "Current version" line at 200; the comment at 844 at `6f95917a1`, 835 at `8fa0dc24b`) | Release gate from its TDD RED milestone (M1, after absorption). Regression-guard until then. |
| AC-RC30-19 | U-18 | REQ-RC30-011 | `go test ./internal/cli/ -run TestPreviewCountEqualsRunCount -count=1 -v` | `--- PASS`. Preview count equals run count for the same fixture. | none: the 266-against-423 split is audit-cited and not reproduced | Regression-guard. |
| AC-RC30-20 | U-19, U-20 | REQ-RC30-025 | Investigation with a recorded result: `grep -rl "\.claude/skills/moai/" internal/template/templates/.claude/rules` (the hits, E-07), and `grep -n "claude/agents/moai" docs-site/content/ko/cli-reference/update.md` (E-08) | Each rule-file hit gets exactly one disposition from the fixed set {`deployed` (names a path the template deploys), `stale` (names a path the template does not deploy; corrected under REQ-RC30-025), `not-a-reference`}, recorded in research.md. The docs line gets exactly one of {`found-at-real-path`, `not-present`}. The criterion fails for a hit with no disposition, for a `stale` hit without a recorded correction, and for a docs line with neither value. | E-07 (rule file list), E-08 (docs: no match) | GAP-INVESTIGATE. Regression-guard; no release gate. |
| AC-RC30-21 | U-21 | REQ-RC30-002 | `go test ./internal/cli/ -run TestProfileSyncResolution -count=1 -v` | `--- PASS`. Profile sync reads the project's profile mapping from the profile launch ledger, not from the environment alone. The ledger's location is not named in this SPEC text (Gap G-21); the audit cites `~/.moai/claude-profiles/launch.yaml` (audit §3, not re-read). | E-17 (`profile.GetCurrentName()` at 661 at `6f95917a1`, 652 at `8fa0dc24b`) | Release gate from its TDD RED milestone (M1, after absorption). Regression-guard until then. |
| AC-RC30-22 | U-23 | REQ-RC30-007 | `go test ./internal/cli/ -run '^TestGateYAMLCommentsSurviveRestore$' -count=1 -v` (fixture: 12 comment lines) | `--- PASS`. All comment lines survive backup, wipe, deploy, and restore, except comments on sequence items, which REQ-UYP-014 of SPEC-UPDATE-YAML-PRESERVE-001 excludes (sequences are replaced wholesale; ruling d-20261010T143542Z-5d19, a delegated judgment under the operator's delegation, option (가): REQ-UYP-014 stays). | Reproduced by the M3 delegate with a scratch probe (progress.md section E.2, M3 entry, recorded in commit `c5327202c`). The probe ran after commit `6bede8ca3` and before `c5327202c`, whose only change is the SPEC progress.md. The lane's RED test will reproduce it again; no re-run by the lane is claimed. | GAP-INVESTIGATE. Regression-guard. Test written first in the run stage; it does not exist on this tree yet, so the command selects 0 until then; a run with no tests to run is a failure (ruling d-20261010T040402Z-94b9, item 5). |
| AC-RC30-23 | U-24 | REQ-RC30-007 | `go test ./internal/cli/ -run '^TestLLMYamlRemarshalCallers$' -count=1 -v` (one case per real caller of `persistTeamMode`, which writes through `saveLLMSection`: the cc reset path through `resetTeamModeForCC` and the glm path) | `--- PASS` only when no caller drops a key outside the legacy alias keys `opus:`, `sonnet:`, and `haiku:` (the named exception of REQ-RC30-007, decision DR-2 of SPEC-MODEL-MATRIX-UPDATE-001). A run in which any caller drops a key outside that exception fails, and its output names that caller; the named list is diagnostic output of a failing run, not a pass. | E-14 (`saveLLMSection` re-marshals the typed `LLMConfig`, which has no `deployment_mode` field) | GAP-INVESTIGATE. Regression-guard. Test written first in the run stage; it does not exist on this tree yet, so the command selects 0 until then; a run with no tests to run is a failure (ruling d-20261010T040402Z-94b9, item 5). |
| AC-RC30-24 | U-25 | REQ-RC30-003 | Investigation: `ls -la .moai/logs/config-change-audit.log` is the audit's path. The log is absent from this tree (E-18). The log is located in the operator profile first, and the result is recorded in research.md with its command and output. | PASS only when the untraced removal is attributed to a command whose output is recorded in research.md. An unresolved result is a Gap in research.md, not a PASS outcome: the criterion fails while no attributing command and output are recorded, and the missing evidence is named in the Gap. | E-18 (no such file in this tree) | deferred: excluded from the landing verdict; the operator checks the log at the 3.2-6-2 rc.30 field run (mo.ai.kr). Regression-guard. REQ-RC30-003 is owned by card t1668; this row is its cross-reference. |
| AC-RC30-25 [RETIRED] [MOVED to card t1669] | U-26, U-27, U-28, U-29, U-30 | REQ-RC30-014, REQ-RC30-016, REQ-RC30-017, REQ-RC30-018, REQ-RC30-019, REQ-RC30-020 | `make tui-snapshot-verify` (target exists, E-09) and `go test ./internal/cli/ -run 'TestUpdateOutputGolden\|TestUpdatePlainFallback' -count=1 -v` (TTY golden shows N of M, bar, one ✓/!/✗ line per step, one-screen summary, one conflict count above the summary; plain fallback carries the same counts; locale part: `go test ./internal/cli/ -run TestUpdateLocaleKeyParity -count=1 -v`, which requires the four locale tables ko, en, ja, and zh to define the same keys and the summary to show the same counts in each locale) | `--- PASS` for each. The golden is changed only through `make tui-snapshot`, on a reviewed diff. The one conflict count above the summary equals the conflict count of the Reconciliation line (`update/report/outcome.go:106`); this equality is the count-level judge of REQ-RC30-017. | E-10 (the step line renderer at 134-139); E-21 (no `conversation_language` reference in the renderer files) | Release gate from its TDD RED milestone (M4) in card t1669. Regression-guard until then. Not a condition of this SPEC. |
| AC-RC30-27 | U-01 | REQ-RC30-002 | `go test ./internal/cli/ -run '^TestMigrationKeepsSharedCopyUnderProfileReach$' -count=1 -v`. The test file `internal/cli/migrate_profile_reach_test.go` is untracked on this tree; it is taken from branch `worktree-agent-adee9fb2106290d8a` at commit `30c838d72` and is committed with the M2 change, not before | `--- PASS`. A project mapped to a profile keeps its project-side shared-asset copy and marks it (option (c)). The mapping is the one the test fixture sets through the profile launch ledger (Gap G-27); the reach judgment for the same state is AC-RC30-28. | E-25 is a build-red observation (undefined `FileEntry.KeptBy` and `manifest.KeptByProfileReach`), not a behavioral RED. The behavioral RED is recorded in M2 after the API stub (plan § B.3 step 6). | Release gate from its behavioral RED in M2 (the owning milestone; M2 exit condition). Regression-guard until then. The keep half only: the removal half (TestMigrationRemovesSharedCopyWithoutProfileReach) moves with AC-RC30-01 [REF] to card t1668 (ruling d-20261009T235135Z-5e35). |
| AC-RC30-28 | U-01 | REQ-RC30-002 | On a project launched under a profile config directory (`CLAUDE_CONFIG_DIR` set to the profile): with the binary built from this tree and invoked by path, run `moai update`, then `moai doctor`. | `moai doctor` reports the shared assets as reachable from the profile session. This criterion is the single reach judge of REQ-RC30-002; the "profile detected" predicate of revision 0.1.3 is replaced by this judgment. | none recorded: the judgment needs a profile session on the operator side; Q6 is partly measured (decision-index.md Q6). | Deferred to t1642 (3.2-2-5) by judgment d-20261010T124923Z-2b09 part 1 (the leader's defer ruling; plan § B.3 step 7 is deferred with it). Resume condition: t1642 adds the reach judgment to `moai doctor`; then this criterion runs as an operator-terminal item (profile session: `moai update`, then `moai doctor`, binary built from this tree), and its output is recorded in progress.md § E.2. Interim: the lane's own reach test (M2 bundle G) confirms reach and does not replace this criterion. Regression-guard until then. Gap G-28. |
AC-RC30-26 [RETIRED] moved to card t1668 (ruling d-20261010T033220Z-3c5f; decision d-20261010T084752Z-d4be); the repair stays with card t1578 (leader correction).

## Gaps named in revision 0.1.4

- **G-06 (AC-RC30-06, the intended field).** The SPEC text does not determine which field `ConfigManager.Save` is to change in the round-trip test. The run-stage test names it at M3, and the name is recorded in progress.md § E.2. The criterion does not pass until then. This revision does not guess the field.
- **G-17 (REQ-RC30-017 count clause).** AC-RC30-09 judges the literal string only. The count-level judge is AC-RC30-25 [REF] in card t1669, which SPEC A does not adopt. Owner: card t1669.
- **G-21 (AC-RC30-21, the project ledger).** The SPEC text names the profile launch ledger only through audit §3 (`~/.moai/claude-profiles/launch.yaml`), which was not re-read in this revision. REQ-RC30-015 does not define a ledger. The run-stage test names the artifact at M1.
- **G-27 (AC-RC30-27, the RED cell).** The first observation is build-red (E-25), recorded by the lane at `6f95917a1`. It was not re-run in revision 0.1.4, which is text-only. The behavioral RED is recorded in M2 after the API stub.
- **G-28 (AC-RC30-28, the doctor output).** The exact `moai doctor` line that reports profile reach is not established by this text, and whether the current `moai doctor` reports it is unverified. The first operator-side profile run records the line.

## Mutant probes (verification-completeness § 2)

- AC-RC30-09 (literal): a relocated literal in another file passes a single-file grep. The judge is the repo-wide grep above, and both lines 282 and 367 must go.
- AC-RC30-18 (one banner): a mutant that removes the identity band and keeps `update.go:200` passes "exactly one line". The judge must also assert the band is present (REQ-TUXIU-015).
- AC-RC30-25 [RETIRED] [MOVED to card t1669] (golden): a mutant that regenerates the golden from the current output passes. Golden updates must be made only through `make tui-snapshot` on a reviewed diff.

## Definition of Done (for the run phase)

- Every criterion has its RED cell (or its stated regression-guard reason) recorded in progress.md § E.2 on the then-current tree.
- Every test-named criterion has its effect cell recorded now (E-01 to E-25) and its TDD RED milestone named in the matrix.
- Each criterion in this SPEC passes at its owning milestone, with the -v output recorded in progress.md, or is recorded in progress.md as blocked or deferred with a named trigger.
- The coverage target of 85% or more per touched package is measured by `go test -cover` on those packages (measurement Gap until M7).
