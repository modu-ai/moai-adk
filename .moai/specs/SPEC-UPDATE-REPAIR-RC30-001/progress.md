# SPEC-UPDATE-REPAIR-RC30-001 — Progress

Plan-phase skeleton. Run-phase and sync-phase sections are placeholders only.

## §E.1 Plan-phase Audit-Ready Signal

plan_status: bypassed
plan_complete_at: 2026-10-09T18:19:27Z
audit_ready: true

Plan artifacts written: spec.md, plan.md, acceptance.md, design.md, research.md, decision-index.md. Plan-audit iteration 9 (`.moai/reports/t1619/plan-audit-iter9.md`) recorded FAIL with fifteen blocking defects (D1 to D15). Revision 0.1.4 applies D1 to D15 and the planning-session judgments (board ruling d-20261010T101935Z-4bc4). The plan audit is recorded as BYPASSED by operator instruction (f56a, 1353), not as PASS. Q1 is resolved from disk (POLICY-COVERED, decision-index.md Q1). Q2 (profile reach) is decided by operator decision d-20261010T060848Z-dc3b. Q6 is partly measured (decision-index.md Q6). Q8 gates card t1668 only. Q3 (queue overlaps) is recorded with disposition undecided. Q7 carries one applied default: 60 columns (decision-index.md Q7). B1 is decided by the leader's split (d-20261009T235135Z-5e35): 22 criteria and 17 requirements in SPEC A. The revision commit is recorded in the lane's progress record, not here.

## §E.2 Run-phase Evidence

### M1 — plan § B.2 steps 2 and 3 (manager-develop, card t1619)

**Tree.** HEAD `a437e8cf9` on `WT-moai-update-repair` (manager-spec text commit), before the M1 commit. The RED-ledger pin `6f95917a1` is an ancestor of HEAD (`git merge-base --is-ancestor 6f95917a1 a437e8cf9`, exit 0). `git diff --name-only 6f95917a1 a437e8cf9` lists seven files, all under `.moai/specs/SPEC-UPDATE-REPAIR-RC30-001/`; no Go or template path differs, so each 0.1.4 cell applies to this tree.

**Step 2 — RED re-observation, before any implementation.** Each entry is the command string from acceptance.md, run alone from the worktree root by a scratch harness (`eval` of the exact string; stdout, stderr and exit status captured per entry). The comparison against the 0.1.4 cell was mechanical (a scratch script that parses the ledger text). `git status --short` was empty before and after the run (`reobs/env.txt`).

| Entry | 0.1.4 exit | Re-observed exit | Stdout against the 0.1.4 cell |
|---|---|---|---|
| E-01 | 0 | 0 | match |
| E-02 | 0 | 0 | match |
| E-03 | 0 | 0 | match |
| E-04 | 0 | 0 | match |
| E-05 | 0 | 0 | match |
| E-06 | 0 | 0 | match |
| E-07 | 0 | 0 | same 45 paths, different order (traversal order) |
| E-08 | 1 | 1 | match |
| E-09 | 0 | 0 | match |
| E-10 | 0 | 0 | match |
| E-11 | 1 | 1 | match |
| E-12 | 0 | 0 | match |
| E-13 | 0 | 0 | match |
| E-14 | 1 | 1 | match |
| E-15 | 0 | 0 | match |
| E-16 | 0 | 0 | match |
| E-17 | 0 | 0 | match |
| E-18 | 1 | 1 | match |
| E-19 | 0 | 0 | match |
| E-20 | 0 | 0 | match |
| E-21 | 1 | 1 | match |
| E-22 | 0 | 0 | match |
| E-23 | 1 | 1 | match |
| E-24 | 0 | 0 | match |
| E-25 | 1 | 0 | DIFFERS: zero-test selector, Gap G-M1-1 |

Outcome: 23 entries match on exit status and verbatim stdout. E-07 matches on exit status and on the set of 45 paths; only the order differs, and the ledger already records that order as directory-traversal order on this filesystem, so it is order-only and not a blocker. E-25 differs and is recorded as Gap G-M1-1. No entry is a blocker.

Verbatim re-observation (stdout and stderr as captured; the 0.1.4 cell is the ledger text in acceptance.md):

**E-01** `grep -n "0 user customizations modified" internal/cli/update_archive.go` — exit 0 (0.1.4: 0) — match

stdout (verbatim):

```text
282://	total: N skills archived, 0 user customizations modified
367:	_, _ = fmt.Fprintln(out, tui.Pill(tui.PillOpts{Kind: tui.PillOk, Solid: false, Label: fmt.Sprintf("total: %d skills archived, 0 user customizations modified", archived), Theme: &th}))
```

**E-02** `grep -n "preserved: " internal/cli/update/report/outcome.go` — exit 0 (0.1.4: 0) — match

stdout (verbatim):

```text
115:		result += "\n  preserved: " + p
```

**E-03** `grep -n "Current version" internal/cli/update.go` — exit 0 (0.1.4: 0) — match

stdout (verbatim):

```text
200:		_, _ = fmt.Fprintln(out, tui.KV("Current version", "moai-adk "+currentVersion, tui.KVOpts{Theme: &th, KeyWidth: 16}))
844:	// Card t1527 D1: the "Current version" line is dropped here — the top of
```

**E-04** `grep -n "os.Remove" internal/cli/migrate_project_assets.go` — exit 0 (0.1.4: 0) — match

stdout (verbatim):

```text
83:	// outside-pointing symlink must not route os.Remove to an external
174:			if err := os.Remove(p); err != nil {
```

**E-05** `grep -n "kept project-side" internal/cli/migrate_project_assets.go` — exit 0 (0.1.4: 0) — match

stdout (verbatim):

```text
166:				report("  migration: kept project-side (user counterpart not confirmed — opt in via 'moai bundle add <name>' or re-run update): %s", relSlash)
206:		report("migration: %d removed, %d preserved (user-modified), %d kept project-side, %d user-created untouched", removed, preserved, stayed, untouched)
```

**E-06** `grep -n "WithVersion" internal/cli/update_template_sync.go` — exit 0 (0.1.4: 0) — match

stdout (verbatim):

```text
456:					template.WithVersion(version.GetVersion()),
619:					template.WithVersion(version.GetVersion()),
```

**E-07** `grep -rl "\.claude/skills/moai/" internal/template/templates/.claude/rules` — exit 0 (0.1.4: 0) — same 45 paths, different order (traversal order)

stdout (verbatim):

```text
internal/template/templates/.claude/rules/moai/design/constitution.md
internal/template/templates/.claude/rules/moai/core/askuser-protocol-reference.md
internal/template/templates/.claude/rules/moai/core/moai-mcp-tools-catalogue.md
internal/template/templates/.claude/rules/moai/core/agent-common-protocol-reference.md
internal/template/templates/.claude/rules/moai/core/security-decision-contract.md
internal/template/templates/.claude/rules/moai/core/moai-constitution-detail.md
internal/template/templates/.claude/rules/moai/development/agent-authoring.md
internal/template/templates/.claude/rules/moai/development/spec-frontmatter-schema.md
internal/template/templates/.claude/rules/moai/development/karpathy-quickref.md
internal/template/templates/.claude/rules/moai/development/skill-authoring.md
internal/template/templates/.claude/rules/moai/development/coding-standards.md
internal/template/templates/.claude/rules/moai/development/manager-develop-prompt-template.md
internal/template/templates/.claude/rules/moai/development/branch-origin-protocol.md
internal/template/templates/.claude/rules/moai/workflow/context-clear-policy.md
internal/template/templates/.claude/rules/moai/workflow/orchestration-mode-selection.md
internal/template/templates/.claude/rules/moai/workflow/verification-plan-contract.md
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-mechanics.md
internal/template/templates/.claude/rules/moai/workflow/tdd-result-contract.md
internal/template/templates/.claude/rules/moai/workflow/goal-directive-detail.md
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-cards.md
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md
internal/template/templates/.claude/rules/moai/workflow/language-routing-contract.md
internal/template/templates/.claude/rules/moai/workflow/archived-agent-rejection.md
internal/template/templates/.claude/rules/moai/workflow/cadence-bridge.md
internal/template/templates/.claude/rules/moai/workflow/trace-ledger-contract.md
internal/template/templates/.claude/rules/moai/workflow/skill-routing.md
internal/template/templates/.claude/rules/moai/workflow/mx-tag-protocol.md
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-detail.md
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-gates.md
internal/template/templates/.claude/rules/moai/workflow/goal-directive.md
internal/template/templates/.claude/rules/moai/workflow/read-only-status-contract.md
internal/template/templates/.claude/rules/moai/workflow/resource-budget-contract.md
internal/template/templates/.claude/rules/moai/workflow/phase-id-contract.md
internal/template/templates/.claude/rules/moai/workflow/team-capability-resolver.md
internal/template/templates/.claude/rules/moai/workflow/snapshot-consumer-contract.md
internal/template/templates/.claude/rules/moai/workflow/delivery-policy.md
internal/template/templates/.claude/rules/moai/workflow/native-invocation-model.md
internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md
internal/template/templates/.claude/rules/moai/workflow/graceful-exit-mutation-contract.md
internal/template/templates/.claude/rules/moai/workflow/sync-backup-integrity.md
internal/template/templates/.claude/rules/moai/workflow/sync-coverage-scope-contract.md
internal/template/templates/.claude/rules/moai/workflow/verification-batch-pattern.md
internal/template/templates/.claude/rules/moai/workflow/session-handoff-examples.md
internal/template/templates/.claude/rules/moai/workflow/cross-session-messaging.md
internal/template/templates/.claude/rules/moai/NOTICE.md
```

**E-08** `grep -n "claude/agents/moai" docs-site/content/ko/cli-reference/update.md` — exit 1 (0.1.4: 1) — match

stdout: (none)

**E-09** `grep -n "tui-snapshot" Makefile` — exit 0 (0.1.4: 0) — match

stdout (verbatim):

```text
27:.PHONY: all build test lint fix clean install verify-local-install generate templ-generate help release-local constitution-check ci-local pr-merge ci-disable verify-required-checks tui-snapshot tui-snapshot-verify preflight lint-fast test-race-short agents-emit agents-emit-check commands-emit commands-emit-check embed-manifest embed-manifest-check embed-check fmt-check tool-policy-drift-check
193:tui-snapshot: ## Regenerate all internal/tui golden snapshots (UPDATE_GOLDEN=1)
197:tui-snapshot-verify: ## Verify all internal/tui golden snapshots match on-disk state (no update)
```

**E-10** `grep -n "steps" internal/cli/update_tux.go` — exit 0 (0.1.4: 0) — match

stdout (verbatim):

```text
134:// "N/M steps" count. It replaces the legacy "N/M steps complete" plain text
139:	return fmt.Sprintf("  %s %s %d/%d steps", prefix, bar, done, total)
```

**E-11** `grep -n "settings.json" internal/cli/update_restore.go` — exit 1 (0.1.4: 1) — match

stdout: (none)

**E-12** `grep -n "Use:" internal/cli/bundle.go` — exit 0 (0.1.4: 0) — match

stdout (verbatim):

```text
34:	Use:   "bundle",
44:	Use:   "add <bundle-name>",
51:	Use:   "remove <bundle-name>",
64:	cmd := &cobra.Command{Use: "bundle-test"}
```

**E-13** `grep -n "bundle add" internal/cli/user_asset_phase.go` — exit 0 (0.1.4: 0) — match

stdout (verbatim):

```text
3:// `moai init` (the first-install trigger, REQ-024), `moai bundle add|remove`
```

**E-14** `grep -n "deployment_mode" internal/config/manager.go` — exit 1 (0.1.4: 1) — match

stdout: (none)

**E-15** `grep -n "Opus" internal/config/types.go` — exit 0 (0.1.4: 0) — match

stdout (verbatim):

```text
389:// legacy backward-compat alias fields Opus/Sonnet/Haiku are removed. An
```

**E-16** `grep -n "stateManifestMatch" internal/userassets/install.go` — exit 0 (0.1.4: 0) — match

stdout (verbatim):

```text
734:	stateManifestMatch // current == manifest hash, ≠ shipped → refresh (REQ-008)
789:	case stateManifestMatch:
845:		return stateManifestMatch
```

**E-17** `grep -n "GetCurrentName" internal/cli/update.go` — exit 0 (0.1.4: 0) — match

stdout (verbatim):

```text
661:	profileName := profile.GetCurrentName()
```

**E-18** `ls -la .moai/logs/config-change-audit.log` — exit 1 (0.1.4: 1) — match

stdout: (none)

stderr (verbatim):

```text
ls: .moai/logs/config-change-audit.log: No such file or directory
```

**E-19** `git grep -n "settings.json" e73a7cbf5 -- internal/cli/update_integrity_probe.go` — exit 0 (0.1.4: 0) — match

stdout (verbatim):

```text
e73a7cbf5:internal/cli/update_integrity_probe.go:60://   - .claude/settings.json: the Claude Code settings merge target; must parse as
e73a7cbf5:internal/cli/update_integrity_probe.go:70:	{rel: ".claude/settings.json", check: probeJSON, claudeOnly: true},
```

**E-20** `grep -n "migrateProjectCommonAssets(" internal/cli/update_template_sync.go` — exit 0 (0.1.4: 0) — match

stdout (verbatim):

```text
1185:		removed, err := migrateProjectCommonAssets(projectRoot, homeDir, userAssetsInstalled, nil, func(format string, args ...interface{}) {
```

**E-21** `grep -rn "conversation_language" internal/cli/update_tux.go internal/cli/update_archive.go internal/cli/update/report/outcome.go` — exit 1 (0.1.4: 1) — match

stdout: (none)

**E-22** `grep -n "0 user customizations modified" internal/cli/tuxiu_characterization_test.go` — exit 0 (0.1.4: 0) — match

stdout (verbatim):

```text
201:		for _, want := range []string{"Updated 26 files", "3 skills archived, 0 user customizations modified"} {
```

**E-23** `grep -rn "0 user customizations modified" internal/cli --include='*.go' --exclude='*_test.go' --exclude='update_archive.go'` — exit 1 (0.1.4: 1) — match

stdout: (none)

**E-24** `grep -rn "0 user customizations modified" internal/cli --include='*.go' --exclude='*_test.go'` — exit 0 (0.1.4: 0) — match

stdout (verbatim):

```text
internal/cli/update_archive.go:282://	total: N skills archived, 0 user customizations modified
internal/cli/update_archive.go:367:	_, _ = fmt.Fprintln(out, tui.Pill(tui.PillOpts{Kind: tui.PillOk, Solid: false, Label: fmt.Sprintf("total: %d skills archived, 0 user customizations modified", archived), Theme: &th}))
```

**E-25** `go test ./internal/cli/ -run '^TestMigrationKeepsSharedCopyUnderProfileReach$' -count=1 -v` — exit 0 (0.1.4: 1) — DIFFERS: zero-test selector, Gap G-M1-1

stdout (verbatim):

```text
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.511s [no tests to run]
```

**Step 3 — removal switch, TDD order.**

(a) Regression guard, no criterion (AC-RC30-02 belongs to card t1668): `TestMigrationDefaultKeepsConfirmedProjectCopy` in `internal/cli/migrate_removal_switch_test.go`. It does not reference `removalEnabled`.

(b) RED — `go test ./internal/cli/ -run '^TestMigrationDefaultKeepsConfirmedProjectCopy$' -count=1 -v`, exit 1 (the test ran; the failure is behavioral, not a build error: the confirmed copy is removed on the default path):
```text
=== RUN   TestMigrationDefaultKeepsConfirmedProjectCopy
    migrate_removal_switch_test.go:20: .claude/skills/moai-foundation-core/SKILL.md missing: stat /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestMigrationDefaultKeepsConfirmedProjectCopy2769196026/001/.claude/skills/moai-foundation-core/SKILL.md: no such file or directory
--- FAIL: TestMigrationDefaultKeepsConfirmedProjectCopy (0.37s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.310s
FAIL
```

(c) Implementation: `var removalEnabled = false` in `internal/cli/migrate_project_assets.go` (after `projectCommonAssetRels`). The `os.Remove(p)` call sits behind `if !removalEnabled`, which keeps the copy with the kept-project-side accounting (`stayed++`, `migrationPreservedProjectFiles[p] = true`, one report line) and does not touch `removed` or `removedPaths`.

(d) GREEN — the same command, exit 0:
```text
=== RUN   TestMigrationDefaultKeepsConfirmedProjectCopy
--- PASS: TestMigrationDefaultKeepsConfirmedProjectCopy (0.66s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.783s
```

**Existing tests that assert default-path removal.** With the inert switch, two existing tests failed on the same assertion: `TestMigrationPreservesExistingMirrorEntries` (`internal/cli/update_migrate_test.go:701`, `.claude/skills/moai-foundation-core/SKILL.md still present`) and `TestTallyCountsMigrationRemovals` (`update_migrate_test.go:1286`, `:1290`, `:1295`). REQ-RC30-003 keeps removal off until card t1668, so the default-path expectation is no longer the SPEC's. Each test now calls `enableRemovalForTest(t)` as its first statement; the helper (in `migrate_removal_switch_test.go`) sets `removalEnabled = true` and restores the prior value in `t.Cleanup`. Their assertions are unchanged, so the removal mechanics stay covered.

**Step 3 verification (this tree, this run).**
- Named subset on unchanged HEAD (the 31 tests that reach the migration path, found by a scan of `internal/cli/*_test.go` for calls to `migrateProjectCommonAssets`, `runTemplateSyncWithProgress`, `runUpdateCobraCmd`): command `go test ./internal/cli/ -run '^(TestRunTemplateSyncWithProgress_VersionMatch|TestRunTemplateSyncWithProgress_VersionMismatch|TestFR3A1_MigrationDeletionAnchoredToProjectRoot|TestFR3_1_MigrationPreservesOnUnverifiable|TestMigrationClassifiesUnregisteredMirrorCopy|TestUpdateCancelKeepsProjectAssetsIntact|TestRF2F3_MigrationPreservesPostInstallEdits|TestRF2F3b_MigrationPreservesUntracked|TestRunTemplateSyncWithProgress_VersionMatchSkips|TestRunTemplateSyncWithProgress_ForceFlagBypassesVersionCheck|TestUpdate_Characterize_SkippingSync_NoHooksGuidance|TestUpdateMigratesLegacyProject|TestUpdateOptedOutMigrationPreservesManagedRootLocals|TestMigrationIdempotent|TestUpdateForceKeepsPlantedCommonAssets|TestUpdateLocalModeKeepsFullScope|TestUpdateNeverFlipsModeRecord|TestForceUpdatePreservesUserCommonAssets|TestNotDemonstratedPreservationHoldsAtNextLocalUpdate|TestPreservedSectionSurvivesRestoreUntouched|TestUnknownSkillPreservedAcrossLocalUpdates|TestMigrationPreservesExistingMirrorEntries|TestUpdatePreservesUserModifiedPolicies|TestMigrationRefusesRemovalWhenUserInstallMissing|TestCancelledUpdateKeepsProjectManagedAssets|TestRecordlessBranchTallyCountsReconcileRemovals|TestReconcileDropsStaleManifestRecord|TestPostRemovalUserFileSurvivesNextUpdate|TestTallyCountsMigrationRemovals|TestUserCancelledMerge_LeavesConfigByteIdentical|TestSkipSyncNoArchive)$' -count=1 -v`, `--- PASS` 31, exit 0.
- Same 31-test subset after the switch, before the adjustment above: 29 PASS, 2 FAIL (the two tests named above), exit 1.
- Named set after the adjustment: the 31 + 3 tests found by a widened scan (`TestMigrationArchivesModifiedBeforeRemoval`, `TestMigrationArchiveRefusesSymlinkedArchiveDestination`, `TestUpdateRepair_SyncArchivesLegacySkillsBeforeCleanup`) + the guard = 35 tests. Command: `go test ./internal/cli/ -run '^(TestRunTemplateSyncWithProgress_VersionMatch|TestRunTemplateSyncWithProgress_VersionMismatch|TestFR3A1_MigrationDeletionAnchoredToProjectRoot|TestFR3_1_MigrationPreservesOnUnverifiable|TestMigrationClassifiesUnregisteredMirrorCopy|TestUpdateCancelKeepsProjectAssetsIntact|TestRF2F3_MigrationPreservesPostInstallEdits|TestRF2F3b_MigrationPreservesUntracked|TestRunTemplateSyncWithProgress_VersionMatchSkips|TestRunTemplateSyncWithProgress_ForceFlagBypassesVersionCheck|TestUpdate_Characterize_SkippingSync_NoHooksGuidance|TestUpdateMigratesLegacyProject|TestUpdateOptedOutMigrationPreservesManagedRootLocals|TestMigrationIdempotent|TestUpdateForceKeepsPlantedCommonAssets|TestUpdateLocalModeKeepsFullScope|TestUpdateNeverFlipsModeRecord|TestForceUpdatePreservesUserCommonAssets|TestNotDemonstratedPreservationHoldsAtNextLocalUpdate|TestPreservedSectionSurvivesRestoreUntouched|TestUnknownSkillPreservedAcrossLocalUpdates|TestMigrationPreservesExistingMirrorEntries|TestUpdatePreservesUserModifiedPolicies|TestMigrationRefusesRemovalWhenUserInstallMissing|TestCancelledUpdateKeepsProjectManagedAssets|TestRecordlessBranchTallyCountsReconcileRemovals|TestReconcileDropsStaleManifestRecord|TestPostRemovalUserFileSurvivesNextUpdate|TestTallyCountsMigrationRemovals|TestUserCancelledMerge_LeavesConfigByteIdentical|TestSkipSyncNoArchive|TestMigrationArchivesModifiedBeforeRemoval|TestMigrationArchiveRefusesSymlinkedArchiveDestination|TestUpdateRepair_SyncArchivesLegacySkillsBeforeCleanup|TestMigrationDefaultKeepsConfirmedProjectCopy)$' -count=1 -v`. All 35 `--- PASS`, exit 0. Top-level result lines, verbatim:
```text
--- PASS: TestRunTemplateSyncWithProgress_VersionMatch (0.01s)
--- PASS: TestRunTemplateSyncWithProgress_VersionMismatch (0.57s)
--- PASS: TestFR3A1_MigrationDeletionAnchoredToProjectRoot (0.01s)
--- PASS: TestFR3_1_MigrationPreservesOnUnverifiable (0.00s)
--- PASS: TestMigrationClassifiesUnregisteredMirrorCopy (1.07s)
--- PASS: TestUpdateCancelKeepsProjectAssetsIntact (1.00s)
--- PASS: TestMigrationDefaultKeepsConfirmedProjectCopy (1.00s)
--- PASS: TestRF2F3_MigrationPreservesPostInstallEdits (0.01s)
--- PASS: TestRF2F3b_MigrationPreservesUntracked (0.02s)
--- PASS: TestRunTemplateSyncWithProgress_VersionMatchSkips (0.01s)
--- PASS: TestRunTemplateSyncWithProgress_ForceFlagBypassesVersionCheck (0.01s)
--- PASS: TestUpdate_Characterize_SkippingSync_NoHooksGuidance (0.01s)
--- PASS: TestUpdateMigratesLegacyProject (1.22s)
--- PASS: TestUpdateOptedOutMigrationPreservesManagedRootLocals (0.29s)
--- PASS: TestMigrationArchivesModifiedBeforeRemoval (0.01s)
--- PASS: TestMigrationIdempotent (0.67s)
--- PASS: TestUpdateForceKeepsPlantedCommonAssets (0.60s)
--- PASS: TestUpdateLocalModeKeepsFullScope (0.22s)
--- PASS: TestUpdateNeverFlipsModeRecord (0.88s)
--- PASS: TestForceUpdatePreservesUserCommonAssets (1.01s)
--- PASS: TestNotDemonstratedPreservationHoldsAtNextLocalUpdate (0.76s)
--- PASS: TestPreservedSectionSurvivesRestoreUntouched (0.59s)
--- PASS: TestMigrationArchiveRefusesSymlinkedArchiveDestination (0.07s)
--- PASS: TestUnknownSkillPreservedAcrossLocalUpdates (1.09s)
--- PASS: TestMigrationPreservesExistingMirrorEntries (1.09s)
--- PASS: TestUpdatePreservesUserModifiedPolicies (0.71s)
--- PASS: TestMigrationRefusesRemovalWhenUserInstallMissing (1.21s)
--- PASS: TestCancelledUpdateKeepsProjectManagedAssets (0.96s)
--- PASS: TestRecordlessBranchTallyCountsReconcileRemovals (0.32s)
--- PASS: TestReconcileDropsStaleManifestRecord (0.44s)
--- PASS: TestPostRemovalUserFileSurvivesNextUpdate (0.85s)
--- PASS: TestTallyCountsMigrationRemovals (1.03s)
--- PASS: TestUserCancelledMerge_LeavesConfigByteIdentical (0.00s)
--- PASS: TestUpdateRepair_SyncArchivesLegacySkillsBeforeCleanup (0.32s)
--- PASS: TestSkipSyncNoArchive (0.01s)
```
- `gofmt -l` on the three touched Go files: no output.
- `GOOS=windows GOARCH=amd64 go build ./...`: exit 0, no output.
- `git grep -n "AskUserQuestion" -- internal/cli/migrate_project_assets.go`: no output, exit 1.
- `golangci-lint run --timeout=10m ./internal/cli/`: `0 issues.`, exit 0.

**Status transition.** `spec.md` frontmatter `status: draft` → `in-progress` on this M1 commit (manager-develop's one transition, per `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix). `updated:` is already 2026-10-10 and is unchanged.

**Gaps (not observed in this record).**
- G-M1-1 (AC-RC30-27 keep half, E-25): the test file `internal/cli/migrate_profile_reach_test.go` is parked for M2, so the behavioral RED cannot run on HEAD. The re-run selects zero tests (`testing: warning: no tests to run`, `PASS`, exit 0). That exit is NOT a pass. The behavioral RED is recorded in M2 (plan § B.3 step 6).
- G-M1-2: `go test -cover` was not measured in M1 (M7 owns coverage).
- G-M1-3: the full `internal/cli` suite and `go test ./...` were not run locally (scope rule). The repository-wide test verdict belongs to the CI run on `origin/develop`, PENDING at the time of this record.
- G-M1-4: `moai spec lint` / `moai spec audit` over the `draft → in-progress` transition was not run; lint and audit verdicts belong to CI and the sync phase.

**Residual risk.**
- R-M1-1: with the switch off, `moai update` no longer removes a template-managed project copy whose user counterpart is confirmed. Such copies stay project-side with the report line `removal is not enabled yet` until card t1668 adopts AC-RC30-02.
- R-M1-2 [RESOLVED by commit e4979fd34: the dry-run follows the removal switch]: `internal/cli/update_dryrun_preview.go:81-83` still prints `removed by the asset migration once your user-side copy is confirmed` for those copies. With the inert switch that statement is false. M1 does not change the preview (scope); the fix belongs to card t1668 or to U-18 (plan § B.6, M5). Lead decision needed.

### M1 step 4, bundle B — U-09 (AC-RC30-10) (manager-develop, card t1619)

Bundle B runs the plan § B.2 step 4 items in order: U-09, U-03, U-12. Tree before bundle B: HEAD `19accc920` on `WT-moai-update-repair`, clean (pre-flight). Each item is its own local commit; nothing is pushed.

**RED (TDD order).** The judging test is new: `TestNonCatalogSkillMessage` in `internal/cli/migrate_noncatalog_message_test.go`. It compiles against the tree as it stands; no stub was needed. The fixture is a template-managed `.claude/skills/moai-retired-skill/SKILL.md` with no catalog entry and no user counterpart.

Command: `go test ./internal/cli/ -run TestNonCatalogSkillMessage -count=1 -v`, exit 1. Verbatim output:

```text
=== RUN   TestNonCatalogSkillMessage
    migrate_noncatalog_message_test.go:52: non-catalog skill gets the opt-in message:   migration: kept project-side (user counterpart not confirmed — opt in via 'moai bundle add <name>' or re-run update): .claude/skills/moai-retired-skill/SKILL.md
    migrate_noncatalog_message_test.go:55: non-catalog skill gets no distinct no-catalog message:   migration: kept project-side (user counterpart not confirmed — opt in via 'moai bundle add <name>' or re-run update): .claude/skills/moai-retired-skill/SKILL.md
--- FAIL: TestNonCatalogSkillMessage (0.55s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.446s
FAIL
```

The failure is behavioral: the test runs, and both assertions fail on the opt-in line that the current code prints for this copy.

**Implementation.** `internal/cli/migrate_project_assets.go`, in the not-confirmed arm: a skill path with no catalog entry gets a distinct line (`no catalog entry for this managed skill — no bundle carries it, so there is no opt-in`). New helper `skillNotInCatalog`, placed beside `skillSourceRel`. Every other not-confirmed copy keeps the opt-in line. `removalEnabled` and the removal path are untouched.

**GREEN.** Same command, exit 0:

```text
=== RUN   TestNonCatalogSkillMessage
--- PASS: TestNonCatalogSkillMessage (0.40s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.233s
```

**Gates (each its own command).**
- `gofmt -l internal/cli/migrate_project_assets.go internal/cli/migrate_noncatalog_message_test.go`: no output.
- `go vet ./internal/cli/`: no output.
- `golangci-lint run ./internal/cli/`: `0 issues.`
- `GOOS=windows GOARCH=amd64 go build ./...` (B1): no output.

**Gaps (not observed in this item).** The full `internal/cli` suite was not run for this item; the change-scoped migration subset runs once after bundle B (see the bundle B close). `go test -cover` was not measured (M7 owns coverage).

**Residual risk.** The change alters one output line for a catalog-less managed skill. Removal behavior is unchanged; the removal of such a copy stays with card t1668.

### M1 step 4, bundle B — U-03 (AC-RC30-04) (manager-develop, card t1619)

**RED (TDD order).** The judging test is new: `TestBundleAnnounceForPluginMode` in `internal/cli/bundle_announce_test.go`. It compiles against the tree as it stands. Fixture: `.moai/config/sections/llm.yaml` with `deployment_mode: plugin`, an empty user selection (L0 only), and a temporary HOME. The plugin-mode subtest is the RED. The local-mode subtest pins the gate (a local-mode project gets no announce); it passes on arrival, which is the REQ's intended behavior for that arm.

Command: `go test ./internal/cli/ -run TestBundleAnnounceForPluginMode -count=1 -v`, exit 1. Verbatim output:

```text
=== RUN   TestBundleAnnounceForPluginMode
=== RUN   TestBundleAnnounceForPluginMode/plugin_mode_names_each_uninstalled_pack
    bundle_announce_test.go:24: update output lacks the add command for 11 optional pack(s) [backend commands consult creative delivery deployment design devops frontend ops-tools testing]:
        user-asset update: 394 installed, 0 refreshed, 0 removed, 0 collision-skipped, 0 divergence-preserved
          unconverted reference (no Codex-side mapping — report to the harness maintainers): agents-skills/moai-workflow-project/references/navigator.md
          unconverted reference (no Codex-side mapping — report to the harness maintainers): agents-skills/moai-workflow-worktree/SKILL.md
          unconverted reference (no Codex-side mapping — report to the harness maintainers): agents-skills/moai-workflow-worktree/modules/moai-adk-integration.md
          unconverted reference (no Codex-side mapping — report to the harness maintainers): agents-skills/moai-workflow-worktree/modules/tools-integration.md
          unconverted reference (no Codex-side mapping — report to the harness maintainers): agents-skills/moai-workflow-worktree/references/examples.md
          unconverted reference (no Codex-side mapping — report to the harness maintainers): agents-skills/moai/SKILL.md
          unconverted reference (no Codex-side mapping — report to the harness maintainers): agents-skills/moai/workflows/codemaps.md
          unconverted reference (no Codex-side mapping — report to the harness maintainers): agents-skills/moai/workflows/design.md
          unconverted reference (no Codex-side mapping — report to the harness maintainers): agents-skills/moai/workflows/gtd.md
          unconverted reference (no Codex-side mapping — report to the harness maintainers): agents-skills/moai/workflows/harness-build-entry.md
          unconverted reference (no Codex-side mapping — report to the harness maintainers): agents-skills/moai/workflows/harness-builder.md
          unconverted reference (no Codex-side mapping — report to the harness maintainers): agents-skills/moai/workflows/harness.md
          unconverted reference (no Codex-side mapping — report to the harness maintainers): agents-skills/moai/workflows/plan.md
          unconverted reference (no Codex-side mapping — report to the harness maintainers): agents-skills/moai/workflows/plan/spec-assembly.md
          unconverted reference (no Codex-side mapping — report to the harness maintainers): agents-skills/moai/workflows/project/doc-generation.md
          unconverted reference (no Codex-side mapping — report to the harness maintainers): agents-skills/moai/workflows/project/meta-harness.md
          unconverted reference (no Codex-side mapping — report to the harness maintainers): agents-skills/moai/workflows/review.md
          unconverted reference (no Codex-side mapping — report to the harness maintainers): agents-skills/moai/workflows/run.md
          unconverted reference (no Codex-side mapping — report to the harness maintainers): agents-skills/moai/workflows/run/context-loading.md
          unconverted reference (no Codex-side mapping — report to the harness maintainers): agents-skills/moai/workflows/run/task-decomposition.md
          unconverted reference (no Codex-side mapping — report to the harness maintainers): agents-skills/moai/workflows/sync.md
          unconverted reference (no Codex-side mapping — report to the harness maintainers): agents-skills/moai/workflows/sync/delivery.md
          unconverted reference (no Codex-side mapping — report to the harness maintainers): agents-skills/moai/workflows/sync/doc-execution.md
          unconverted reference (no Codex-side mapping — report to the harness maintainers): agents-skills/moai/workflows/sync/quality-gates-quality.md
          unconverted reference (no Codex-side mapping — report to the harness maintainers): codex-agents/manager-develop.toml
          unconverted reference (no Codex-side mapping — report to the harness maintainers): codex-agents/manager-docs.toml
          unconverted reference (no Codex-side mapping — report to the harness maintainers): codex-agents/manager-lead.toml
          unconverted reference (no Codex-side mapping — report to the harness maintainers): codex-agents/manager-spec.toml
          unconverted reference (no Codex-side mapping — report to the harness maintainers): codex-agents/plan-auditor.toml
          unconverted reference (no Codex-side mapping — report to the harness maintainers): codex-agents/sync-auditor.toml
        user-asset prune: 0 removed, 0 divergence-preserved
=== RUN   TestBundleAnnounceForPluginMode/local_mode_names_no_pack
--- FAIL: TestBundleAnnounceForPluginMode (1.49s)
    --- FAIL: TestBundleAnnounceForPluginMode/plugin_mode_names_each_uninstalled_pack (0.56s)
    --- PASS: TestBundleAnnounceForPluginMode/local_mode_names_no_pack (0.92s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	2.778s
```

The failure is behavioral. The phase runs, and its output lacks `moai bundle add <name>` for all 11 optional packs of the embedded catalog. The 30 `unconverted reference` lines are the phase's existing output (`writeInstallSummary`), not part of this change.

**Implementation.** `internal/cli/user_asset_phase.go`: `runUserAssetUpdatePhase` calls a new helper, `announceUnselectedOptionalPacks(".", saved.Bundles, out)`, after the prune summary. The helper prints one line per optional pack the recorded selection lacks: `  optional pack not installed: <name> — run: moai bundle add <name>`. It returns at once unless `config.ReadDeployMode(".")` is `plugin`. The `config` import is new in this file. The project root is `"."` as in the rest of the update flow (`resolveUpdateDeployMode(".", ...)`).

**GREEN.** Same command, exit 0:

```text
=== RUN   TestBundleAnnounceForPluginMode
=== RUN   TestBundleAnnounceForPluginMode/plugin_mode_names_each_uninstalled_pack
=== RUN   TestBundleAnnounceForPluginMode/local_mode_names_no_pack
--- PASS: TestBundleAnnounceForPluginMode (1.15s)
    --- PASS: TestBundleAnnounceForPluginMode/plugin_mode_names_each_uninstalled_pack (0.68s)
    --- PASS: TestBundleAnnounceForPluginMode/local_mode_names_no_pack (0.47s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	2.474s
```

**Gates (each its own command).**
- `gofmt -l internal/cli/user_asset_phase.go internal/cli/bundle_announce_test.go`: no output.
- `go vet ./internal/cli/`: no output.
- `golangci-lint run ./internal/cli/`: `0 issues.`
- `GOOS=windows GOARCH=amd64 go build ./...` (B1): no output.

**Gaps (not observed in this item).** The announce was not observed in a real `moai update` run on an operator profile; the judging test drives `runUserAssetUpdatePhase` directly, not the whole update command. Coverage not measured (M7).

**Residual risk.** A plugin-mode project sees one line per unselected pack on every update, for as long as the pack stays unselected; the phase keeps no once-only record. A project whose `deployment_mode` record is absent or out of set is not announced, which follows `ReadDeployMode`'s record-less reading.

## §E.3 Run-phase Audit-Ready Signal

_pending run-phase (manager-develop)_

## §E.4 Sync-phase Audit-Ready Signal

_pending sync-phase (manager-docs)_

## §G Absorption Record

Absorption is done for revision 0.1.4 (spec.md § G.2). Both tips were ancestors of HEAD `6f95917a1` when this record was written.

- WT-user-asset-bundle (t1591): tip `a372a984c`, merged as `ff2bc876d` ("Merge branch 'WT-user-asset-bundle' into WT-moai-update-repair (card t1619)"). Earlier merges of the same branch: `f5e039f2b` and `cbb6c7940`, both ancestors of HEAD.
- WT-update-migration-fixes (t1578): tip `e73a7cbf5`, merged as `6f95917a1` ("Merge branch 'WT-update-migration-fixes' into WT-moai-update-repair (card t1619)"). Earlier merge: `763720170` ("Merge branch 'WT-update-migration-fixes' into WT-moai-update-repair"), an ancestor of HEAD.
- Ranges are measured from the merge-base with the absorbed ref (`git merge-base HEAD <absorbed-tip>`), not from a literal SHA. If a branch moves, the lane merges it again and records the new tip here.
- Tree of the RED cells in acceptance.md: `6f95917a1`.
