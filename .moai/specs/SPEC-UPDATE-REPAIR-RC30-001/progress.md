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

### M1 step 4, bundle B — U-12 (AC-RC30-13) (manager-develop, card t1619)

**GREEN at arrival (no RED observed).** AC-RC30-13 is a regression guard (its effect cell is the notice only). No RED was observed for this item and none is claimed. The judging test is new: `TestKeptOptionalAssetRefresh` in `internal/cli/migrate_kept_optional_test.go`. Fixture: two outdated copies, each template-managed with the hash of its own bytes and with no user counterpart installed (the home holds L0 only): `.claude/agents/moai/e2e-tester.md` (consult-pack agent) and `.agents/skills/moai-clean/SKILL.md` (commands-pack skill). The test requires each copy to be refreshed in place (its bytes changed) or named by a line of the migration output.

Command: `go test ./internal/cli/ -run TestKeptOptionalAssetRefresh -count=1 -v`, exit 0. Verbatim output:

```text
=== RUN   TestKeptOptionalAssetRefresh
--- PASS: TestKeptOptionalAssetRefresh (0.37s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.264s
```

**Why it holds, and the arm it holds on.** `migrateProjectCommonAssets` has no refresh arm: it never rewrites a kept copy. The criterion holds on the notice arm alone. Both copies reach the not-confirmed branch and each gets the line `migration: kept project-side (user counterpart not confirmed — opt in via 'moai bundle add <name>' or re-run update): <path>`. The migration tally also counts both (`3 kept project-side`, in the probe below).

**Mutation probe (test sensitivity; not a RED of this item).** To show that the GREEN is not vacuous, the opt-in report line in `internal/cli/migrate_project_assets.go` was turned into a discarded `fmt.Sprintf` (the message is still computed, nothing is reported). One edit, then the test re-run:

Command: `go test ./internal/cli/ -run TestKeptOptionalAssetRefresh -count=1 -v`, exit 1. Verbatim output:

```text
=== RUN   TestKeptOptionalAssetRefresh
    migrate_kept_optional_test.go:71: kept copy .claude/agents/moai/e2e-tester.md is neither refreshed nor named in the output; lines: ["  migration: unregistered mirror copy (kept — removal needs creation provenance or your approval): .claude/skills/moai-custom/SKILL.md" "  migration: kept project-side (removal is not enabled yet): .claude/skills/moai-foundation-core/SKILL.md" "  migration: preserved (bytes differ from the template — treated as user-modified): .claude/skills/moai-workflow-tdd/SKILL.md" "migration: 0 removed, 1 preserved (user-modified), 3 kept project-side, 1 user-created untouched"]
    migrate_kept_optional_test.go:71: kept copy .agents/skills/moai-clean/SKILL.md is neither refreshed nor named in the output; lines: ["  migration: unregistered mirror copy (kept — removal needs creation provenance or your approval): .claude/skills/moai-custom/SKILL.md" "  migration: kept project-side (removal is not enabled yet): .claude/skills/moai-foundation-core/SKILL.md" "  migration: preserved (bytes differ from the template — treated as user-modified): .claude/skills/moai-workflow-tdd/SKILL.md" "migration: 0 removed, 1 preserved (user-modified), 3 kept project-side, 1 user-created untouched"]
--- FAIL: TestKeptOptionalAssetRefresh (0.81s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.972s
FAIL
```

The probe edit was reverted by the inverse edit. `git status --short` then showed `internal/cli/migrate_project_assets.go` with no difference from HEAD. The GREEN above was re-run on the restored tree (exit 0, same `--- PASS` line).

**Gates (each its own command).**
- `gofmt -l internal/cli/migrate_kept_optional_test.go`: no output.
- `go vet ./internal/cli/`: no output.
- `golangci-lint run ./internal/cli/`: `0 issues.`
- `GOOS=windows GOARCH=amd64 go build ./...` (B1): no output.

**Gaps (not observed in this item).** The judging test calls `migrateProjectCommonAssets` directly, not the whole update command, so the wiring that prints the reporter to the user's output (`update_template_sync.go:1185`) was not re-observed in a real run. No refresh arm exists to observe.

**Residual risk.** A kept copy is named, not refreshed: the user must opt in or update the copy by hand. The notice is the opt-in line; it does not say the copy is outdated. That is accurate for the not-confirmed arm, and the criterion does not require more, but a reader may take the line for a current copy.

### M1 step 4, bundle B — close (manager-develop, card t1619)

**Change-scoped regression subset (this tree, after U-12).** One plain command, exit 0: 45 top-level tests. The set is the 35-test migration set of the bundle A record (§ Step 3), the 2 user-asset phase tests (`TestUpdatePhaseRefreshesManifestMatch`, `TestUpdatePhaseHonorsRecordedSelectionAndPrunes`), the 5 bundle command tests (`TestBundleAddInstallsBundleEntries`, `TestBundleRemoveTakesComplementOnly`, `TestBundleRemoveHonorsDivergence`, `TestBundleUnknownNameRefused`, `TestBundleRemoveDefersDependency`), and the 3 judging tests of this bundle. Command: `go test ./internal/cli/ -run '^(TestRunTemplateSyncWithProgress_VersionMatch|TestRunTemplateSyncWithProgress_VersionMismatch|TestFR3A1_MigrationDeletionAnchoredToProjectRoot|TestFR3_1_MigrationPreservesOnUnverifiable|TestMigrationClassifiesUnregisteredMirrorCopy|TestUpdateCancelKeepsProjectAssetsIntact|TestRF2F3_MigrationPreservesPostInstallEdits|TestRF2F3b_MigrationPreservesUntracked|TestRunTemplateSyncWithProgress_VersionMatchSkips|TestRunTemplateSyncWithProgress_ForceFlagBypassesVersionCheck|TestUpdate_Characterize_SkippingSync_NoHooksGuidance|TestUpdateMigratesLegacyProject|TestUpdateOptedOutMigrationPreservesManagedRootLocals|TestMigrationIdempotent|TestUpdateForceKeepsPlantedCommonAssets|TestUpdateLocalModeKeepsFullScope|TestUpdateNeverFlipsModeRecord|TestForceUpdatePreservesUserCommonAssets|TestNotDemonstratedPreservationHoldsAtNextLocalUpdate|TestPreservedSectionSurvivesRestoreUntouched|TestUnknownSkillPreservedAcrossLocalUpdates|TestMigrationPreservesExistingMirrorEntries|TestUpdatePreservesUserModifiedPolicies|TestMigrationRefusesRemovalWhenUserInstallMissing|TestCancelledUpdateKeepsProjectManagedAssets|TestRecordlessBranchTallyCountsReconcileRemovals|TestReconcileDropsStaleManifestRecord|TestPostRemovalUserFileSurvivesNextUpdate|TestTallyCountsMigrationRemovals|TestUserCancelledMerge_LeavesConfigByteIdentical|TestSkipSyncNoArchive|TestMigrationArchivesModifiedBeforeRemoval|TestMigrationArchiveRefusesSymlinkedArchiveDestination|TestUpdateRepair_SyncArchivesLegacySkillsBeforeCleanup|TestMigrationDefaultKeepsConfirmedProjectCopy|TestUpdatePhaseRefreshesManifestMatch|TestUpdatePhaseHonorsRecordedSelectionAndPrunes|TestBundleAddInstallsBundleEntries|TestBundleRemoveTakesComplementOnly|TestBundleRemoveHonorsDivergence|TestBundleUnknownNameRefused|TestBundleRemoveDefersDependency|TestNonCatalogSkillMessage|TestBundleAnnounceForPluginMode|TestKeptOptionalAssetRefresh)$' -count=1 -v`. Result: 45 `--- PASS`, 0 `--- FAIL`, final lines `PASS` and `ok  	github.com/modu-ai/moai-adk/internal/cli	15.529s`. The output was too large to inline (41 KB); the top-level result lines, verbatim (`grep "^--- PASS"` on that output), are:

```text
--- PASS: TestBundleAnnounceForPluginMode (0.72s)
--- PASS: TestBundleAddInstallsBundleEntries (0.63s)
--- PASS: TestBundleRemoveTakesComplementOnly (0.65s)
--- PASS: TestBundleRemoveHonorsDivergence (1.08s)
--- PASS: TestBundleUnknownNameRefused (0.00s)
--- PASS: TestBundleRemoveDefersDependency (0.63s)
--- PASS: TestRunTemplateSyncWithProgress_VersionMatch (0.00s)
--- PASS: TestRunTemplateSyncWithProgress_VersionMismatch (0.22s)
--- PASS: TestFR3A1_MigrationDeletionAnchoredToProjectRoot (0.00s)
--- PASS: TestFR3_1_MigrationPreservesOnUnverifiable (0.00s)
--- PASS: TestMigrationClassifiesUnregisteredMirrorCopy (0.71s)
--- PASS: TestUpdateCancelKeepsProjectAssetsIntact (0.48s)
--- PASS: TestKeptOptionalAssetRefresh (0.60s)
--- PASS: TestNonCatalogSkillMessage (0.43s)
--- PASS: TestMigrationDefaultKeepsConfirmedProjectCopy (0.43s)
--- PASS: TestRF2F3_MigrationPreservesPostInstallEdits (0.00s)
--- PASS: TestRF2F3b_MigrationPreservesUntracked (0.00s)
--- PASS: TestRunTemplateSyncWithProgress_VersionMatchSkips (0.00s)
--- PASS: TestRunTemplateSyncWithProgress_ForceFlagBypassesVersionCheck (0.00s)
--- PASS: TestUpdate_Characterize_SkippingSync_NoHooksGuidance (0.00s)
--- PASS: TestUpdateMigratesLegacyProject (0.48s)
--- PASS: TestUpdateOptedOutMigrationPreservesManagedRootLocals (0.16s)
--- PASS: TestMigrationArchivesModifiedBeforeRemoval (0.00s)
--- PASS: TestMigrationIdempotent (0.33s)
--- PASS: TestUpdateForceKeepsPlantedCommonAssets (0.38s)
--- PASS: TestUpdateLocalModeKeepsFullScope (0.21s)
--- PASS: TestUpdateNeverFlipsModeRecord (0.35s)
--- PASS: TestForceUpdatePreservesUserCommonAssets (0.49s)
--- PASS: TestNotDemonstratedPreservationHoldsAtNextLocalUpdate (0.39s)
--- PASS: TestPreservedSectionSurvivesRestoreUntouched (0.17s)
--- PASS: TestMigrationArchiveRefusesSymlinkedArchiveDestination (0.04s)
--- PASS: TestUnknownSkillPreservedAcrossLocalUpdates (0.41s)
--- PASS: TestMigrationPreservesExistingMirrorEntries (0.54s)
--- PASS: TestUpdatePreservesUserModifiedPolicies (0.16s)
--- PASS: TestMigrationRefusesRemovalWhenUserInstallMissing (0.39s)
--- PASS: TestCancelledUpdateKeepsProjectManagedAssets (0.47s)
--- PASS: TestRecordlessBranchTallyCountsReconcileRemovals (0.18s)
--- PASS: TestReconcileDropsStaleManifestRecord (0.18s)
--- PASS: TestPostRemovalUserFileSurvivesNextUpdate (0.49s)
--- PASS: TestTallyCountsMigrationRemovals (0.69s)
--- PASS: TestUserCancelledMerge_LeavesConfigByteIdentical (0.00s)
--- PASS: TestUpdateRepair_SyncArchivesLegacySkillsBeforeCleanup (0.20s)
--- PASS: TestSkipSyncNoArchive (0.01s)
--- PASS: TestUpdatePhaseRefreshesManifestMatch (0.54s)
--- PASS: TestUpdatePhaseHonorsRecordedSelectionAndPrunes (1.00s)
```

**Scope.** The full `internal/cli` suite and `go test ./...` were not run locally (scope rule, as in bundle A). The repository-wide verdict belongs to the CI run on `origin/develop`, PENDING at the time of this record. `go test -cover` was not measured (M7). The status transition was recorded on the bundle A commit; bundle B changes no status field.

### M1 step 4, bundle C — U-11 (AC-RC30-12) (manager-develop, card t1619)

Bundle C is the single item U-11 (plan § B.2 step 4, user-root downgrade refusal). Tree before the item: HEAD `d511692d7` on `WT-moai-update-repair`, clean (pre-flight: branch, HEAD, clean status, and `go build ./internal/userassets/` with no output). The item is one local commit; nothing is pushed.

**Reading of E-16.** Confirmed on this tree: the `stateManifestMatch` refresh arm of `applyTarget` (`internal/userassets/install.go`) has no version comparison. The version record is per file (`FileEntry.MoaiVersion`, REQ-006); the manifest has no root-level version. The SPEC text does not define how a refusal is reported (see Gaps).

**RED (TDD order).** The judging test is new: `TestUserRootDowngradeRefused` in `internal/userassets/downgrade_refusal_test.go`. It compiles against the tree as it stands; no stub was needed. Fixture: a first `Install` under `v3.2.0-rc.30` records `claude-skills/moai-alpha/SKILL.md`; the shipped copy then changes, so the file on disk still matches its record and `classifyTarget` returns `stateManifestMatch`. The older binary (`v3.2.0-rc.29`) runs a second `Install` on that home. A control subtest runs a newer binary (`v3.2.0-rc.31`) on the same fixture.

Command: `go test ./internal/userassets/ -run TestUserRootDowngradeRefused -count=1 -v`, exit 1. Verbatim output:

```text
=== RUN   TestUserRootDowngradeRefused
=== PAUSE TestUserRootDowngradeRefused
=== CONT  TestUserRootDowngradeRefused
=== RUN   TestUserRootDowngradeRefused/older_binary_refuses_the_refresh
    downgrade_refusal_test.go:46: older binary refreshed the file: bytes "alpha skill body v2\n", want the recorded "alpha skill body v1\n"
    downgrade_refusal_test.go:53: older binary rewrote the record: {SHA256:6e9b25a96b83b3e9b38eed60f9eea0b9ff152dde7a3b6b99208b1b22b1219006 Bundle:core InstalledAt:2026-10-06T00:00:00Z MoaiVersion:v3.2.0-rc.29 installedByJournal:false unknown:map[]}, want the v3.2.0-rc.30 record of the recorded bytes
    downgrade_refusal_test.go:56: Refreshed = 2, want 0: a refused refresh is not counted
    downgrade_refusal_test.go:65: refusal is not observable: no failure for claude-skills/moai-alpha/SKILL.md in []
=== RUN   TestUserRootDowngradeRefused/newer_binary_still_refreshes_the_same_fixture
--- FAIL: TestUserRootDowngradeRefused (0.09s)
    --- FAIL: TestUserRootDowngradeRefused/older_binary_refuses_the_refresh (0.05s)
    --- PASS: TestUserRootDowngradeRefused/newer_binary_still_refreshes_the_same_fixture (0.04s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/userassets	0.393s
FAIL
```

The failure is behavioral. The older binary refreshed the file, rewrote the record to `v3.2.0-rc.29`, and reported no refusal. The control passes, so the fixture reaches the refresh arm and the refusal is the version comparison, not a blanket stop.

An earlier RED run, before the control's count assertion was corrected, also printed `Refreshed = 2, want 1` for the control. That expectation was wrong, not the code: `dirTargets` walks the same source directory (`.claude/skills/moai-alpha`) for both the claude and the agents skill roots, so one changed source file refreshes two records. The control now asserts the bytes and the record only. The RED above is the run on the final test text.

**Implementation.** `internal/userassets/install.go` and `go.mod`:
- `applyTarget`, `stateManifestMatch` arm: before the refresh, the file's own recorded `MoaiVersion` is compared with `in.MoaiVersion`. A newer record returns an error. Both callers already record such an error as a per-file `FileOutcome` in `Result.Failures`. Nothing is written, the record is untouched, and `Refreshed` does not count the file.
- `versionNewer`: normalizes both sides to a leading `v` (the Makefile's `VERSION` default is `git describe --tags`, which is `v`-prefixed; `.goreleaser.yml:22` passes `{{.Version}}`, which omits it), then compares with `golang.org/x/mod/semver`. That module is already in the build list and `go.sum`. A value that is not valid semver is never judged newer, so the refresh proceeds as before.
- `go.mod`: `golang.org/x/mod v0.41.0` loses its `// indirect` comment. This is the only change `go mod tidy -diff` reports for the new direct import.

**GREEN.** Same command, exit 0:

```text
=== RUN   TestUserRootDowngradeRefused
=== PAUSE TestUserRootDowngradeRefused
=== CONT  TestUserRootDowngradeRefused
=== RUN   TestUserRootDowngradeRefused/older_binary_refuses_the_refresh
=== RUN   TestUserRootDowngradeRefused/newer_binary_still_refreshes_the_same_fixture
--- PASS: TestUserRootDowngradeRefused (0.02s)
    --- PASS: TestUserRootDowngradeRefused/older_binary_refuses_the_refresh (0.01s)
    --- PASS: TestUserRootDowngradeRefused/newer_binary_still_refreshes_the_same_fixture (0.01s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/userassets	0.316s
```

**Regression checks (this tree, after the change).**
- `go test ./internal/userassets/ -count=1`: `ok  	github.com/modu-ai/moai-adk/internal/userassets	3.261s`.
- Other importers of `internal/userassets`: `go test ./internal/cli/ -count=1 -run '^(…)$'` over the 59 test functions of the 13 cli test files that import the package: `ok  	github.com/modu-ai/moai-adk/internal/cli	77.075s`. `go test ./internal/hook/ -count=1 -run '^(TestUserRootSlugDirsMatchUserassetsRoots|TestUserRootFormsEmitNamespacedForm|TestUserRootFormTrackedScoping)$'`: `ok  	github.com/modu-ai/moai-adk/internal/hook	0.787s`.

**Gates (each its own command).**
- `gofmt -l internal/userassets/install.go internal/userassets/downgrade_refusal_test.go`: no output.
- `go vet ./internal/userassets/`: no output.
- `golangci-lint run ./internal/userassets/`: `0 issues.`
- `go mod tidy -diff`, after the go.mod change: no output.
- `GOOS=windows GOARCH=amd64 go build ./...` (B1): no output.

**Pre-commit gate (`moai gate` workflow, scoped run).** Phase 1B: `moai verify check --key-current` exited 1 with "no snapshot recorded for the current working-tree key" (key `d511692d7ce7458b912bc8b63a4618190c361b8a:b6f388707a828e36`), so no result was reused. Scope: the changed package and the two changed Go files. Lint PASS (`golangci-lint run ./internal/userassets/`: `0 issues.`). Format PASS (`gofmt -l` on the two files: no output). Type PASS (`go vet ./internal/userassets/`: no output). Test PASS (`go test -race -count=1 ./internal/userassets/`: `ok  	github.com/modu-ai/moai-adk/internal/userassets	212.430s`). The test run is above the gate's 60 s per-check budget; that budget was not enforced in this run, and under the gate's timeout rule it would have been a WARNING. A scoped run is not recorded as a snapshot.

**Gaps (not observed in this item).**
- How a refusal is reported is not defined by the SPEC text; AC-RC30-12 names only the outcome. Chosen: a per-file `FileOutcome` in `Result.Failures`, reason `refresh refused: recorded moai version … is newer than this binary (…) — REQ-RC30-012`. `writeInstallSummary` prints it as a `failure:` line. No counter or exit code was added. `runUserAssetUpdatePhase` returns nil on per-file failures, so the update's exit status is unchanged.
- The comparison is per file, because the manifest records no root-level version (REQ-006). A root whose files carry mixed versions is judged file by file.
- The `stateAbsent` install arm (a file missing from a root, reinstalled by an older binary) is not refused. AC-RC30-12 names the refresh only.
- The full `internal/cli` suite and `go test ./...` were not run; `.claude/rules/local/gitflow-lane-protocol.md` § 8 forbids a local full-suite run in a lane. The 59 test functions above are the ones in the test files that import `internal/userassets`. The full-suite verdict belongs to CI on `origin/develop` (PENDING at the time of this record).
- `go test -cover` was not measured (M7 owns coverage).

**Residual risk.**
- A version that is neither `vX.Y.Z[-pre]` nor `X.Y.Z[-pre]` (a `dev` build, for example) is never judged newer, so such a binary still refreshes. That is the pre-change behavior, and it is an open fail-open path.
- The refused file's journal entry stays staged with `WriteCompleted` false until the manifest save clears the journal. The existing per-file failure arms leave the same state. The test does not inspect the journal.

**Files changed.** `internal/userassets/install.go`, `internal/userassets/downgrade_refusal_test.go` (new), `go.mod`, `.moai/specs/SPEC-UPDATE-REPAIR-RC30-001/progress.md`.

### M1 step 4, bundle D — U-17 (AC-RC30-18) (manager-develop, card t1619)

Bundle D is two items, U-17 then U-21 (plan § B.2 step 4). This entry is U-17; U-21 has its own entry. Tree before the item: HEAD `26b3fc31f` on `WT-moai-update-repair`, clean (pre-flight: branch, HEAD, `git status --short` empty, `go build ./internal/cli/` no output). The item is one local commit; nothing is pushed.

**Reading of E-03 (re-read on this tree).** `internal/cli/update.go` prints the "Current version" KV on every non-re-exec pass, ahead of every mode branch. The identity band (`renderIdentityBand`, REQ-TUXIU-015) prints only in `runTemplateSyncWithReporter` (`internal/cli/update_template_sync.go`, the sync header). A default run therefore printed two version lines. The band renders only on the sync path: `--check`, `--shell-env`, `--binary`, `--restore`, `--dry-run`, the non-project refusal, the lock error and the v2 clean-reinstall print no band.

**Design.** The band moves to the top of `runUpdate`, where the KV was, and the sync header no longer repeats it. Removing only the KV would strip the sole version line from every path listed above. The re-executed pass keeps skipping the band through the existing `reexecPassActive` marker.

**RED (TDD order).** New judging test `TestUpdateSingleBanner` in `internal/cli/update_single_banner_test.go`. It runs the default `updateCmd.RunE` in a moai project (`.moai/config/sections/system.yaml`), with `version.Version` pinned to `v9.9.9-banner`, `yes` set, and a mock update checker that reports no update. It counts the output lines carrying the version and asserts the band marker. The test compiles against the tree as it stood; no stub was needed.

Command: `go test ./internal/cli/ -run TestUpdateSingleBanner -count=1 -v`, exit 1. Verbatim (second run). The first run printed the whole run output in the failure message, so the message was narrowed and the RED re-run; the same assertion fired both times.

```text
=== RUN   TestUpdateSingleBanner
moai-cli-test: userHomeDirFn redirected real home to sandbox /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/moai-cli-home-3832687276
2026/10/10 20:58:05 INFO SUNSET_CONFIG_DORMANT_NOTICE spec="SPEC-V3R2-MIG-003 REQ-018" yaml_path=.moai/config/sections/sunset.yaml advice="SunsetConfig has no runtime hot path; activation requires a new SPEC"
    update_single_banner_test.go:60: a run must print exactly one version line, got 2: ["Current version   moai-adk v9.9.9-banner" "◆ MoAI-ADK  v9.9.9-banner  go1.26.8 · claude"]
--- FAIL: TestUpdateSingleBanner (1.12s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	2.599s
FAIL
```

**Implementation.** `internal/cli/update.go` (the `currentVersion` block in `runUpdate`): `tui.KV("Current version", …)` becomes `renderIdentityBand(currentVersion, th)`, under the same `reexecPassActive` guard. `internal/cli/update_template_sync.go`: the band print and the now-unused `currentVersion` local are removed from `runTemplateSyncWithReporter`; the "Syncing templates" line stays.

**Test assertions changed.** REQ-RC30-013 requires one banner, so the literal "Current version" no longer appears in any run. Three existing assertions that required it now require the band marker `◆ MoAI-ADK` (on `stripSGR` output): `internal/cli/update_test.go` `TestUpdateCmd_CheckOnly_NoDeps`, `internal/cli/coverage_test.go` `TestRunUpdate_DefaultIsTemplateSync`, `internal/cli/integration_test.go` `TestUpdateCmd_DefaultIsTemplateSync`. In `internal/cli/update_reexec_banner_test.go` two comments that described the removed KV and the sync-header band were corrected; no assertion there changed.

**GREEN.** Same command, exit 0, verbatim:

```text
=== RUN   TestUpdateSingleBanner
moai-cli-test: userHomeDirFn redirected real home to sandbox /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/moai-cli-home-2695936028
2026/10/10 21:02:27 INFO SUNSET_CONFIG_DORMANT_NOTICE spec="SPEC-V3R2-MIG-003 REQ-018" yaml_path=.moai/config/sections/sunset.yaml advice="SunsetConfig has no runtime hot path; activation requires a new SPEC"
--- PASS: TestUpdateSingleBanner (0.74s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.569s
```

**Regression run (one plain command, exit 0).** `go test ./internal/cli/ -run 'TestUpdateSingleBanner|TestUpdateCmd_CheckOnly_NoDeps|TestUpdateCmd_DefaultIsTemplateSync|TestRunUpdate_DefaultIsTemplateSync|TestUpdateBanner_IdentityBandIsTheSingleSyncSurface|TestReexecPassActive_FlagGated|TestReexecChildArgv_InsertsMarker|TestReexecMarkerFlag_IsHiddenAndPersistent|TestRunTemplateSync_Timeout' -count=1 -v`: all nine `--- PASS`, last line `ok  	github.com/modu-ai/moai-adk/internal/cli	2.905s`. This run preceded one comment-only reflow of `update.go`; the judging test was re-run after it (GREEN above is the post-reflow run).

**Gates (each its own command).**
- `gofmt -l` on the seven touched Go files: no output. Re-run on `update.go` after the reflow: no output.
- `go vet ./internal/cli/`: no output.
- `golangci-lint run ./internal/cli/`: `0 issues.` Re-run after the reflow: `0 issues.`
- `GOOS=windows GOARCH=amd64 go build ./...` (B1): no output.

**Gaps (not observed in this item).**
- The full `internal/cli` suite was not run locally (`.claude/rules/local/gitflow-lane-protocol.md` § 8). The CI run on `origin/develop` owns the full-suite verdict, PENDING at the time of this record.
- The `testdata/tuxiu/**` goldens are static fixtures. `TestInitUpdateTUXCharacterization` and its siblings read them from disk and never run `runUpdate`, so they are not regenerated here. They already differed from HEAD's live output before this item (the pre-logo golden shows two "Current version" lines and a sync header with no band). Not observed: any other test that asserts band or KV text through a live run. A package-wide grep found only the files named above.
- On an update that installs a new binary, the band shows the version the pre-exec pass started with, and the re-executed pass prints no band. The `New version` and `Updated A → B` lines carry the change. Before this item the post-update band showed the new version. This is a behavior change for review.
- The binary-update path was not run end to end. The judging test covers the no-update default run.

**Pre-commit gate (`moai gate` workflow, scoped to the seven touched Go files).** Phase 1B: `moai verify check --key-current` exited 1 (`no snapshot recorded for the current working-tree key`, key `26b3fc31fb1f21667e566e59d64c9bfeab8fafef:abf8028cdf0bb689`), so nothing was reused and every check ran. Lint PASS (`golangci-lint run ./internal/cli/`: `0 issues.`). Format PASS (`gofmt -l` on the seven files: no output). Type PASS (`go vet ./internal/cli/`: no output). Test PASS (`go test -race ./internal/cli/ -run` over the nine affected tests above, `-count=1 -v`: all nine `--- PASS`; `ok  	github.com/modu-ai/moai-adk/internal/cli	7.202s`). Scope: the full `internal/cli` package and `./...` were not run locally (`.claude/rules/local/gitflow-lane-protocol.md` § 8); the full-suite verdict belongs to CI on `origin/develop` (PENDING). A scoped run is not recorded as a snapshot (gate.md § Snapshot Recording). The routing-ledger record of the gate's Step 2.2 was not written: it writes under `.moai/state/`, which B8 keeps out of this lane.

**Files changed.** `internal/cli/update.go`, `internal/cli/update_template_sync.go`, `internal/cli/update_single_banner_test.go` (new), `internal/cli/update_test.go`, `internal/cli/coverage_test.go`, `internal/cli/integration_test.go`, `internal/cli/update_reexec_banner_test.go`, `.moai/specs/SPEC-UPDATE-REPAIR-RC30-001/progress.md`.

### M1 step 4, bundle D — U-21 (AC-RC30-21) (manager-develop, card t1619)

Second item of bundle D (plan § B.2 step 4). Tree before the item: HEAD `e3f004896` on `WT-moai-update-repair` (the U-17 commit), clean (pre-flight: branch, HEAD, `git status --short` empty). The item is one local commit; nothing is pushed.

**G-21 — the ledger location, determined from the tree.** The SPEC text does not name the ledger (G-21), and the audit citation was not re-read. The tree defines it:
- `internal/profile/profile.go:14` `profilesDir = ".moai/claude-profiles"`; `:23` `launchLedgerFile = "launch.yaml"`.
- `internal/profile/profile.go:55-65` `GetBaseDir()`: `BaseDirOverride` when set, otherwise `$HOME/.moai/claude-profiles`.
- Readers and writers join `GetBaseDir()` with `launchLedgerFile` (`profile.go:420`, `:571`, `:695`). The project mapping is `projects[normalizeProjectKey(root)]`, a profile-name string, accepted only when `<baseDir>/<name>` is a directory (`ResolveLaunchProfileForProject`, `profile.go:461`; `launchCandidateIsUsable`, `profile.go:435`).

So the ledger is `<GetBaseDir()>/launch.yaml`, which is `~/.moai/claude-profiles/launch.yaml` by default. This agrees with the audit citation, and it is now read from code rather than carried over. The tree is unambiguous, so the item proceeds and no blocker is recorded.

**Reading of E-17 (re-read on this tree).** `internal/cli/update.go` calls `profileName := profile.GetCurrentName()` (line 661 at `6f95917a1`; line 662 at `e3f004896`). `GetCurrentName()` is `GetCurrentNameForProject("")`, and an empty root skips the project-scoped ledger lookup (`profile.go:94-111`). With `CLAUDE_CONFIG_DIR` unset, the sync therefore read the `default` profile's preferences whatever the ledger says for the project. The `projects` entry was never consulted.

**RED (TDD order).** New judging test `TestProfileSyncResolution` in `internal/cli/update_profile_sync_test.go`. The fixture uses the existing sandbox helpers (`lpSandbox`, `lpMkProfile`, `lpWriteLedger`, `lpProjectKey`): a moai project with `system.yaml`, a `work` profile whose preferences set `conversation_lang: ja`, and `launch.yaml` mapping the project root to `work`, with `CLAUDE_CONFIG_DIR` unset. A fixture check asserts `profile.GetCurrentNameForProject(root) == "work"` before the run, so a failure cannot come from a broken ledger. The judged step is the default `updateCmd.RunE` (`yes` set, a mock checker that reports no update), followed by the assertion that the project config's `conversation_language` is `ja`. The test compiles against the tree as it stood; no stub was needed.

A first RED run failed at the assertion without the fixture check; the fixture check was then added and the RED re-run. The recorded RED is the second run.

Command: `go test ./internal/cli/ -run TestProfileSyncResolution -count=1 -v`, exit 1. Verbatim:

```text
=== RUN   TestProfileSyncResolution
moai-cli-test: userHomeDirFn redirected real home to sandbox /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/moai-cli-home-634819596
2026/10/10 21:08:02 INFO SUNSET_CONFIG_DORMANT_NOTICE spec="SPEC-V3R2-MIG-003 REQ-018" yaml_path=.moai/config/sections/sunset.yaml advice="SunsetConfig has no runtime hot path; activation requires a new SPEC"
    update_profile_sync_test.go:67: profile sync did not resolve the project's ledger profile: conversation_language = "en", want "ja" (work profile)
--- FAIL: TestProfileSyncResolution (0.52s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.456s
```

**Implementation.** `internal/cli/update.go`, the profile sync (lines 661-664):

```go
	// Sync profile preferences to project config (after template deployment).
	// The profile is the project's mapping in the launch ledger (AC-RC30-21), not
	// the environment alone: GetCurrentName() never reads the ledger.
	profileRoot, _ := filepath.Abs(".")
	profileName := profile.GetCurrentNameForProject(profileRoot)
```

The precedence in `GetCurrentNameForProject` is unchanged: `CLAUDE_CONFIG_DIR` still wins when it is set, and the project's ledger entry decides when it is unset. `filepath.Abs(".")` supplies the absolute root that `normalizeProjectKey` resolves for the lookup. Passing `.` directly would have left the key relative, and `normalizeProjectKey` does not make it absolute.

**GREEN.** Same command, exit 0, verbatim:

```text
=== RUN   TestProfileSyncResolution
moai-cli-test: userHomeDirFn redirected real home to sandbox /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/moai-cli-home-1438482385
2026/10/10 21:08:29 INFO SUNSET_CONFIG_DORMANT_NOTICE spec="SPEC-V3R2-MIG-003 REQ-018" yaml_path=.moai/config/sections/sunset.yaml advice="SunsetConfig has no runtime hot path; activation requires a new SPEC"
--- PASS: TestProfileSyncResolution (0.62s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.468s
```

**Regression runs (one plain command each).**
- `go test ./internal/cli/ -run 'TestProfileSyncResolution|TestUpdateForce_UserOwnedKeysSurvive|TestCleanReinstall_UserOwnedKeysSurvive|TestUpdateForce_UserOwnedKeys_UnrenderableValueFallsBack|TestUpdateVersionDowngradeConfirm_Localized|TestUpdateParticipationStep|TestUpdateSingleBanner|TestRunUpdate_DefaultIsTemplateSync' -count=1 -v`: all `--- PASS` (the participation subtests included); last line `ok  	github.com/modu-ai/moai-adk/internal/cli	4.990s`.
- The same eight names with `go test -race ./internal/cli/ -count=1` (inside the gate below): `ok  	github.com/modu-ai/moai-adk/internal/cli	32.368s`.

**Gates (each its own command).**
- `gofmt -l internal/cli/update.go internal/cli/update_profile_sync_test.go`: no output.
- `go vet ./internal/cli/`: no output.
- `golangci-lint run ./internal/cli/`: `0 issues.`
- `GOOS=windows GOARCH=amd64 go build ./...` (B1): no output.

**Pre-commit gate (`moai gate` workflow, scoped to the two touched Go files).** Phase 1B: `moai verify check --key-current` exited 1 (`no snapshot recorded for the current working-tree key`, key `e3f00489618fd0bd7cb5e953f45b133a3d5399a5:e1fb7eeb0482d0f4`), so every check ran. Lint PASS (`0 issues.`). Format PASS (no output). Type PASS (no output). Test PASS (`go test -race` over the eight affected tests above: `ok  	github.com/modu-ai/moai-adk/internal/cli	32.368s`). Scope and recording as in the U-17 entry: the full package and `./...` were not run locally (`.claude/rules/local/gitflow-lane-protocol.md` § 8); a scoped run is not recorded as a snapshot; the routing-ledger record was not written (B8).

**Gaps (not observed in this item).**
- The full `internal/cli` suite and `./...` were not run locally; the full-suite verdict belongs to CI on `origin/develop` (PENDING).
- Not observed: a run with `CLAUDE_CONFIG_DIR` set. The env-wins branch of `GetCurrentNameForProject` is unchanged by construction and is not exercised by the judging test.
- Other environment-only readers remain in the update flow: `internal/cli/update_participation.go:49` and `internal/cli/downgrade_locale.go:22`, both `profile.ReadPreferences(profile.GetCurrentName())`. AC-RC30-21 names the profile sync only, so they are unchanged. The same defect class applies to them; this needs a leader decision.
- The up-to-date path returns before the profile sync (`internal/cli/update.go`, the `syncSkipped` return near line 632). A run that skips the sync does not apply the ledger profile. This is pre-existing and unchanged.

**Residual risk.**
- Behavior change for review: on a project whose ledger entry names a profile, with `CLAUDE_CONFIG_DIR` unset, `moai update` now applies that profile's preferences (language, user name) to the project config. Before this change it read the `default` profile's preferences, not the mapped profile's.
- The judging test writes only under the sandboxed profile base (`lpSandbox` redirects `BaseDirOverride`); nothing is written under the real `~/.moai`.

**Files changed.** `internal/cli/update.go`, `internal/cli/update_profile_sync_test.go` (new), `.moai/specs/SPEC-UPDATE-REPAIR-RC30-001/progress.md`.

### M2 bundle E — profile-reach keep, AC-RC30-27, option (c) (manager-develop, card t1619)

Bundle E of M2 (plan § B.3 steps 5 and 6). Tree before the bundle: HEAD `f085909fb` on `WT-moai-update-repair` (M1 complete), clean. Pre-flight: branch, HEAD, `git status --short` (empty), `go build ./internal/cli/` (no output), `git cat-file -t 30c838d72` (`commit`). The bundle is two code commits and this record; nothing is pushed.

**Test file — kept and dropped.** `internal/cli/migrate_profile_reach_test.go` was restored from `30c838d72` with `git checkout 30c838d72 -- internal/cli/migrate_profile_reach_test.go` (plain git, no branch change).

Kept:
- `TestMigrationKeepsSharedCopyUnderProfileReach` — the judging test (AC-RC30-27).
- `TestMigrationReachLeavesUserModifiedCopy` — the user-modified guard for the same keep behavior. It passes after GREEN.
- `TestFileEntryKeptBy` — the `kept_by` JSON contract of the new field.

Dropped:
- `TestMigrationRemovesSharedCopyWithoutProfileReach` — the removal half, owned by card t1668 (ruling d-20261009T235135Z-5e35).
- `TestDryRunPreviewReachKeepsSharedCopy` — it needs the dry-run preview to print a kept line for a profile-mapped copy. That is a change to `internal/cli/update_dryrun_preview.go`, outside the migration keep. The preview still prints `preserved` for such a copy (see Gaps). Routed to the leader.

The imports `bytes`, `io/fs`, and `template` went with the dropped functions, and the file header was edited to match. The judging test and the kept functions are unchanged from `30c838d72`.

**RED (TDD order).** Stub first, with no behavior: `internal/manifest/types.go` gains `FileEntry.KeptBy string` (`json:"kept_by,omitempty"`) and `const KeptByProfileReach = "profile-reach"`. The judging test compiles against the stub and fails on its assertions, so the failure is behavioral, not a build failure.

Command: `go test ./internal/cli/ -run '^TestMigrationKeepsSharedCopyUnderProfileReach$' -count=1 -v`, exit 1. Verbatim:

```text
=== RUN   TestMigrationKeepsSharedCopyUnderProfileReach
    migrate_profile_reach_test.go:84: no kept line for .claude/skills/moai-foundation-core/SKILL.md in the migration output: ["  migration: unregistered mirror copy (kept — removal needs creation provenance or your approval): .claude/skills/moai-custom/SKILL.md" "  migration: kept project-side (removal is not enabled yet): .claude/skills/moai-foundation-core/SKILL.md" "  migration: preserved (bytes differ from the template — treated as user-modified): .claude/skills/moai-workflow-tdd/SKILL.md" "migration: 0 removed, 1 preserved (user-modified), 1 kept project-side, 1 user-created untouched"]
    migrate_profile_reach_test.go:96: KeptBy after reload = "", want "profile-reach"
--- FAIL: TestMigrationKeepsSharedCopyUnderProfileReach (0.66s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.951s
```

Tree SHA of the RED state: `745999cd36d72a977d4c59a3771f900961b9faef` (RED commit `ca631aab7`). The RED was observed on the uncommitted tree. `git status --short` before staging listed exactly `AM internal/cli/migrate_profile_reach_test.go` and ` M internal/manifest/types.go`, and the commit recorded exactly those two paths.

**Implementation.** Commit `22dc423ff` (tree `65c459257b9a11fc30940154282eb7aa2e4860ec`):
- `internal/manifest/manifest.go`: `Manager.SetKeptBy(path, by string) error` sets `KeptBy` on one existing entry and leaves its provenance and hashes unchanged. `Track` replaces the whole entry and `GetEntry` returns a copy, so neither can write the mark. An unknown path returns `ErrEntryNotFound`.
- `internal/cli/migrate_project_assets.go`:
  - `profileMapped` is computed once per run as `profile.ResolveLaunchProfileForProject(rootAbs, "") != ""`, where `rootAbs` is `filepath.Abs(projectRoot)`. The predicate is the launch-ledger mapping only, as AC-RC30-27 states.
  - The keep branch runs after the template-bytes and user-modified checks and before the counterpart check, and only for `provenance == template_managed`. It calls `SetKeptBy`, increments `stayed`, records the path in `migrationPreservedProjectFiles`, and prints `kept (profile-mapped project — …): <path>`. Nothing is removed or refreshed.
  - Unmapped projects take the unchanged path, and `removalEnabled` stays `false`.

**Why the root is absolute.** The production caller passes `"."` (`internal/cli/update_template_sync.go`, `runTemplateSyncWithProgress`). A relative root misses an ancestor's mapping, because the subtree lookup walks up from an absolute path. Temporary probes, deleted before the commit, observed: with the mapping on an ancestor and the process in the project, `ResolveLaunchProfileForProject(".")` returned `""` and `ResolveLaunchProfileForProject(root)` returned `work`. The migration called with `"."` then kept the copy and set `KeptBy`. A first draft of the code comment also claimed that a relative root fails for a direct mapping; the probe showed it does not, so the comment states the subtree reason only.

**Deprecated copies are not kept.** A probe with `provenance=deprecated` and bytes equal to the template, in a profile-mapped project, printed `kept project-side (removal is not enabled yet)` and left `KeptBy` empty. This matches spec § D.2 (retired copies are not in the kept set).

**GREEN.** Same command, exit 0, verbatim:

```text
=== RUN   TestMigrationKeepsSharedCopyUnderProfileReach
--- PASS: TestMigrationKeepsSharedCopyUnderProfileReach (0.57s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.445s
```

The two other kept functions, `go test ./internal/cli/ -run '^(TestMigrationReachLeavesUserModifiedCopy|TestFileEntryKeptBy)$' -count=1 -v`: both `--- PASS`, `ok  	github.com/modu-ai/moai-adk/internal/cli	1.112s`.

**Gates (each its own command).**
- `gofmt -l` on the four changed Go files: no output.
- `go vet ./internal/cli/ ./internal/manifest/`: no output.
- `golangci-lint run ./internal/cli/ ./internal/manifest/`: `0 issues.`
- `GOOS=windows GOARCH=amd64 go build ./...`: no output.

**Pre-commit gate (`moai gate` workflow, scoped).** Phase 1B: `moai verify check --key-current` exited 1 (`fresh: false`, key `ca631aab7ed8aa99b76089a7aaf29f5349587b45:47970260e27af8dd`, reason `no snapshot recorded for the current working-tree key`), so every check ran. The checks ran scoped to the two changed packages, with the budgets of `.moai/config/sections/gate.yaml` (vet 30 s, lint 60 s, test 120 s):
- Format: `gofmt -l` on the two changed files, no output. PASS.
- Type: `go vet ./internal/cli/ ./internal/manifest/`, no output, 1.030 s. PASS.
- Lint: `golangci-lint run ./internal/cli/ ./internal/manifest/`, `0 issues.`, 1.728 s. PASS.
- Test: `go test -race ./internal/manifest/` gave `ok  	github.com/modu-ai/moai-adk/internal/manifest	1.848s`. `go test -race ./internal/cli/` with the `-run` pattern of the three profile-reach tests gave three `--- PASS` and `ok  	github.com/modu-ai/moai-adk/internal/cli	8.229s` (36.224 s wall, including the race-instrumented build). PASS (scoped).

Scope: the whole-project gate (standalone `moai gate`) and `go test ./...` were not run locally (lane protocol § 8). Load average was 11.51 / 12.90 / 13.81 (`uptime`). A scoped run is not recorded as a snapshot (`gate.md` § Snapshot Recording).

**Regression (one plain command each).**
- `go test ./internal/cli/ ./internal/manifest/ -run 'Migrat|migrat|Reconcil|reconcil|Preview|preview|KeptBy' -count=1 -v`: internal/cli `ok  	github.com/modu-ai/moai-adk/internal/cli	64.590s`. The manifest half selected no tests (`testing: warning: no tests to run`).
- The cli set has 123 top-level tests: 116 PASS, 7 SKIP, 0 FAIL, counted by hand from the `-v` output of the final-state run. The 7 SKIPs are the `migration_test.go` placeholders (`waiting for migration CLI implementation`), which this bundle does not touch.
- The manifest package was run in full, since the filter selected nothing there: `go test ./internal/manifest/ -count=1 -v` gave 23 top-level tests, all `--- PASS`, and `ok  	github.com/modu-ai/moai-adk/internal/manifest	0.113s`.

**Gaps (not observed in this bundle).**
- The full `internal/cli` suite, `go test ./...`, and the standalone whole-project `moai gate` were not run locally. The repository-wide verdict belongs to CI on `origin/develop` (PENDING at this record).
- The race run covered the three profile-reach tests and the manifest package, not the whole migration set.
- Not observed: a run with `CLAUDE_CONFIG_DIR` set to a profile directory in a project the ledger does not map. The predicate is the launch-ledger mapping only, as AC-RC30-27 states. REQ-RC30-002's trigger text names `CLAUDE_CONFIG_DIR` (see Residual risk).
- `SetKeptBy`'s error branches (nil manifest, unknown path) have no direct unit test. The judging test drives the success branch. The migration checks `hasEntry` before it calls `SetKeptBy`, so the unknown-path branch is not reachable from the migration.
- The dry-run preview (`internal/cli/update_dryrun_preview.go`, `previewReconciliation`) still prints `[dry-run] preserved: <path>` for a profile-mapped shipped copy, with no kept line. `TestDryRunPreviewReachKeepsSharedCopy` was dropped for this reason.
- Not recorded: the routing-ledger entry (`moai harness ledger record`, Step 2.2). It is opt-in and fail-open, and this run was not given the tier, level, or session values.
- Coverage (`go test -cover`) was not measured (M7 owns it).
- Not touched, recorded as Gaps per the lane brief: `internal/cli/update_participation.go` and `internal/cli/downgrade_locale.go` both read the environment-only profile (the U-21 defect class).

**Residual risk.**
- Predicate scope: a session launched with `CLAUDE_CONFIG_DIR` into a profile, in a project the ledger does not map, is not a keep trigger. With `removalEnabled` still `false`, nothing is removed in this build, so the gap matters only once card t1668 turns removal on. This needs a leader decision before then.
- Stale mark: `KeptBy` is not cleared when the ledger later stops mapping the project. A kept copy keeps `kept_by: profile-reach` until its file is removed, because the migration drops entries whose files are gone. Card t1668 must define how a stale mark is read.
- Opt-out: `MOAI_NO_PROFILE_FALLBACK=1` makes the ledger resolution return `""`, so the predicate reads the project as unmapped under that opt-out.
- Message change for a mapped project: a shipped copy now reports `kept (profile-mapped project …)` instead of `kept project-side (user counterpart not confirmed …)` or `kept project-side (removal is not enabled yet)`. The summary's `kept project-side` count includes these copies.
- Working directory: the predicate uses `filepath.Abs(projectRoot)`, which needs the process working directory. An `Abs` failure aborts the migration with `resolve project root`, the same conservative direction as the existing `rootErr` path.

**Files changed.** `internal/manifest/types.go` (stub) and `internal/cli/migrate_profile_reach_test.go` (restored and trimmed), in the RED commit `ca631aab7`. `internal/manifest/manifest.go` and `internal/cli/migrate_project_assets.go`, in the GREEN commit `22dc423ff`. `.moai/specs/SPEC-UPDATE-REPAIR-RC30-001/progress.md`, in this record's commit.

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
