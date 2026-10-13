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

### M2 bundle F — D.1 status table, D.3 init path check, AC-RC30-28 doctor run (manager-develop, card t1619)

Bundle F of M2 (plan § B.3 steps 1, 2 and 7). Tree before the bundle: HEAD `3e7a64808` on `WT-moai-update-repair` (M2 bundle E complete), clean. Pre-flight, each its own command: `git branch --show-current` gave `WT-moai-update-repair`; `git rev-parse --short HEAD` gave `3e7a64808`; `git status --short` gave no output.

**Refused runs (verification-claim-integrity § 3.1).** The worktree-isolation guard refused every command that set `HOME` inline, which is the scratch-home form this bundle requires. Three commands were refused, all `--help` probes: `env -u CLAUDE_CONFIG_DIR -u MOAI_PROJECT_DIR HOME=<scratch> ./bin/moai init --help`, the same form with `update --help`, and the same form with `doctor --help`. The refusal text, verbatim: "This session is isolated in the worktree /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1619, but this command sets HOME, injecting git configuration whose effect on where git writes can't be verified. Refusing to run it — a worktree-isolated session's git operations must target its own worktree. Run the equivalent from /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1619 without the redirect." The guard's advice is to run without the redirect, which would write into the real home, so no other route was tried. Each refusal is recorded in the primary checkout's `.moai/lessons-inbox.jsonl` as `tool_failure:Bash:WorktreeGuardRefusal` (first row `2026-10-10T12:39:09Z`; three rows carry this worktree's path in the 12:xx window). Blocked: D.3 (step 2 from the init run onward) and AC-RC30-28 (step 7 from the update run onward). The lane does not run either of them. The D.3 init-path check and the AC-RC30-28 scratch run need a session outside the card worktree, and both sit on the operator-return list (operator-terminal items) by board judgment `d-20261010T124923Z-2b09` part 2. AC-RC30-28 is deferred to t1642 by board judgment `d-20261010T124923Z-2b09` part 1 (the leader's defer ruling). The scratch home `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/40f97120-d133-4bd7-a79c-a6466b3a087c/scratchpad/m2f-home.xWSvod` was created with `mktemp -d` and never written to.

**Step 2.1 — build.** `make build` reported no error; the last line is the `go build -ldflags … -o bin/moai ./cmd/moai` command, and the binary reports the build ID below. Its prerequisites ran four drift-check test packages, all passing, verbatim:

```text
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.317s
ok  	github.com/modu-ai/moai-adk/internal/template/commandemit	0.243s
ok  	github.com/modu-ai/moai-adk/internal/config/toolpolicy	0.246s
ok  	github.com/modu-ai/moai-adk/internal/template/embedemit	0.644s
```

The prerequisites also ran `templ generate` and `gen-catalog-hashes --all`, which rewrote `catalog.yaml` with the same content: `git status --short` after the build gave no output. `./bin/moai version`, verbatim:

```text
╭────────────────────────────────────╮
│                                    │
│  moai-adk moai_cp/20261007_195927  │
│                                    │
│                                    │
╰────────────────────────────────────╯
 moai_cp/20261007_195927   moai_cp/20261007_195927-298-g3e7a64808   built 2026-10-10T12:38:49Z
```

Judging build (§ 2.2): the build ID `moai_cp/20261007_195927-298-g3e7a64808` names commit `3e7a64808`, the tree's HEAD.

**Steps 2.2 and 2.3 — D.3 init path check: Gap (blocked; source reading only).** Not run (see Refused runs; operator-return list by board judgment `d-20261010T124923Z-2b09` part 2). Source reading, not a measurement: `internal/cli/init.go:944` calls `resumeInitializedProject` (`internal/cli/init_resume.go:60`), which builds the installer at `init_resume.go:118` through `newUserAssetInstaller` (`internal/cli/user_asset_phase.go:57`). The installer's roots are the four `ResolveRoots` entries (`internal/userassets/paths.go:36-43`); the manifest and backup home sit under `~/.moai` (`paths.go:71`, `paths.go:97`). What init wrote under a scratch home is unobserved.

**Partial evidence (awaited-delegate rule).** The D.1 evidence in Step 1 below was committed at `dd94d9570`. Delegate `t1619-m2f-develop` is reported idle. Under the awaited-delegate rule (`.claude/skills/moai-lane-watchdog/SKILL.md` § 3.5), the disposition for this card is to consume that committed evidence as partial bundle F work, not to re-delegate it. The delegate was not re-delegated or messaged in this run, and D.3 and AC-RC30-28 remain not run.

**Step 1 — D.1 status table (read-only on the real home).** Commands: plain `ls -la` and `ls -1A` on the operator home (`~`), `ls -1 … | grep -c '^moai'`, `diff -rq` per skill name, and `test -d`. This shell's `ls` also prints `.` and `..`, so the counts below exclude them. No moai command ran against the real home, and nothing was written.

- **`~/.claude/skills`** (`ls -la ~/.claude/skills`). 36 entries: 19 `moai*` directories (`moai`, `moai-factory-foreman`, `moai-foundation-core`, `moai-foundation-quality`, `moai-foundation-thinking`, `moai-lane-watchdog`, `moai-plan`, `moai-ref-cross-model-audit`, `moai-ref-owasp-checklist`, `moai-ref-secops`, `moai-ref-testing-pyramid`, `moai-run`, `moai-sync`, `moai-workflow-ddd`, `moai-workflow-project`, `moai-workflow-spec`, `moai-workflow-tdd`, `moai-workflow-testing`, `moai-workflow-worktree`) and 17 other entries, not classified here. The count of 19 matches audit § 2. Against this tree's claude-side template (`internal/template/templates/.claude/skills/<name>`, `diff -rq` per name): 13 identical; 3 differ (`moai/workflows/gtd.md`, `moai-factory-foreman/SKILL.md`, `moai-lane-watchdog/SKILL.md`); 3 have no claude-side template (`moai-plan`, `moai-run`, `moai-sync`) and are identical to `internal/template/templates/.agents/skills/<name>`. So audit § 2's "byte-identical to the template" holds for 13 of 19 against this tree. The claude-side template holds 38 `moai*` directories, of which 19 are installed here.
  - Gap left: profile-session reads (audit § 3) are unobserved; which command wrote these copies is not attributed; the rule that selects the 19 from the 38 (catalog L0 / user-install view) is not re-derived; the content of the 3 differing files is not characterized.
- **`~/.claude/agents`** (`ls -1A ~/.claude/agents`). 6 files: `manager-develop.md`, `manager-docs.md`, `manager-lead.md`, `manager-spec.md`, `plan-auditor.md`, `sync-auditor.md`. Mtimes: 2026-10-08 03:26 for four of them, 2026-10-10 10:29 for `plan-auditor.md`, and 2026-10-10 11:06 for `manager-lead.md`. The five names other than `manager-lead` are `l0CoreAgents` in `internal/template/catalog_user_install_view_test.go`; `manager-lead` is the sixth.
  - Gap left: the audit § A removal on 2026-10-09 (no backup) is not established by a listing; the restoration path (U-25) stays open. Disposition (recorded later, same § E.2): U-25 / AC-RC30-24 is deferred to the operator's rc.30 field run per ruling d-20261010T223350Z-7f6e; see `### AC-RC30-24 (U-25) — deferral record` at the end of § E.2. The "stays open" wording above is kept as dated history at the time of this listing.
- **`~/.agents/skills`** (`ls -la ~/.agents/skills`). 36 entries: the same 19 `moai*` names as above, plus 17 others. Against the claude-side template: 6 identical (`moai-ref-cross-model-audit`, `moai-ref-owasp-checklist`, `moai-ref-secops`, `moai-ref-testing-pyramid`, `moai-workflow-ddd`, `moai-workflow-tdd`); 10 differ; 3 (`moai-plan`, `moai-run`, `moai-sync`) have no claude-side template and are identical to the Codex-side template. The one sampled difference, `moai-foundation-core/SKILL.md`, is a path rewrite (`~/.claude/skills/…` to `~/.agents/skills/…`, `.claude/rules/…` to `.moai/policies/…`). This root is therefore a rendered Codex variant rather than a copy, and audit § 2's byte-identity claim does not apply to it.
  - Gap left: whether `moai init` or `moai update` populated this root is not attributed (blocked run); the rendering rule is confirmed on one file only.
- **`~/.codex/agents`** (`ls -1A ~/.codex/agents`). 6 files: `manager-develop.toml`, `manager-docs.toml`, `manager-lead.toml`, `manager-spec.toml`, `plan-auditor.toml`, `sync-auditor.toml`. All six mtimes are 2026-10-10 11:06.
  - Gap left: the writing command is not attributed; the content is not compared with the `.codex` TOMLs of this tree.
- **`~/.moai`** (`ls -la ~/.moai`: 33 entries). `ls -d ~/.moai/skills ~/.moai/agents` gave `ls: /Users/goos/.moai/agents: No such file or directory` and `ls: /Users/goos/.moai/skills: No such file or directory`, exit status 1. So the row's "no skills or agents directories" holds. Entries the user-asset package owns: `user-assets.json` (104,340 bytes, 2026-10-10 11:48; the manifest, `paths.go:71`), `backups/` (the REQ-023 backup home, `paths.go:97`), and `user-assets.acquire-guard.guard` (0 bytes, 2026-10-08 03:26). `user-assets-journal.json` is absent, so no install journal is pending. The other 30 entries, including `claude-profiles/` (the profile base, `internal/profile/profile.go:55-65`), are not classified here.
  - Gap left: "manifest and backups only" describes the user-asset footprint, not the whole directory; the leader's measurement on main is not compared.

**Step 7 — AC-RC30-28 reach: Gap (G-28 unchanged; blocked; deferred to t1642 by board judgment `d-20261010T124923Z-2b09` part 1).** Not run. Resume condition: t1642 adds the reach judgment to `moai doctor`; the operator-terminal run then records its output in §E.2. The acceptance.md row for AC-RC30-28 already carries this deferral. The `moai update` and `moai doctor` runs under `HOME=<scratch>` and `CLAUDE_CONFIG_DIR=<scratch>/.moai/claude-profiles/work` were refused, so there is no update output, no doctor output, and no exit code. No scratch project, profile directory, or ledger was created, because the run that reads them could not start. Source reading only (not a run):

- Ledger format, for the run that follows: `launch.yaml` under `GetBaseDir()` (`internal/profile/profile.go:55-65`, `~/.moai/claude-profiles`). The mapping sits under `projects:`, keyed by the cleaned, symlink-resolved project root (`normalizeProjectKey`, `profile.go:277-286`), with the profile name as value. It counts only when `<base>/<name>` is an existing directory and the name is traversal-safe (`launchCandidateIsUsable`, `profile.go:435-441`). `last_profile` is write-only on this binary (`profile.go:25-35`). The scratch path under `/private/tmp` is already canonical, so no resolution is needed.
- Doctor: the user-install row is `checkUserInstallIntegrity` (`internal/cli/doctor_user_install.go:25`, name "User Install"). It reads `ManifestPath(homeDir)` and compares the manifest with the four `ResolveRoots(homeDir)` roots. That file has no reference to `profile`, `CLAUDE_CONFIG_DIR`, the launch ledger, or the keep marker (`KeptBy`); its only two matches for "reach" are unrelated comments (lines 135 and 159). `internal/cli/doctor.go` has no reference to `profile` or `CLAUDE_CONFIG_DIR`. The keep marker is written in `internal/cli/migrate_project_assets.go` and defined in `internal/manifest`, and no doctor file references it. `internal/cli/doctor_disk.go` reports per-profile sizes (`doctor_disk.go:4-5`, `:146`), not reach.
- Reading: as read from source, no doctor check computes the reach judgment that AC-RC30-28 names. This is a finding about the code, not the runtime line. G-28 stays open until a profile-session doctor output is recorded.

**Build gate.** `GOOS=windows GOARCH=amd64 go build ./...`: no output.

**Gaps (not observed in this bundle).**
- D.3: the init run to a scratch home, and the four roots plus `.moai` listings under it. Refused (see above); the source reading is not a measurement.
- AC-RC30-28: the update and doctor output with exit codes, and the reach line or its absence. Refused; G-28 stays open.
- The scratch project, profile directory, and ledger were not created, because their only reader was the blocked run.
- D.1, per row: profile-session reads of `~/.claude/skills`; install attribution for every root; the rule behind the 19-of-38 selection; the content of the 3 differing claude-side files; the Codex rendering rule beyond one sampled file.
- No test run beyond the four drift-check packages inside `make build`. No code changed, so no coverage or test claim is made.

**Residual risk.**
- This shell's environment sets `CLAUDE_CONFIG_DIR` to this session's real profile and `MOAI_PROJECT_DIR` to the worktree. Any scratch run must override or unset both, as in the `env -u` form above, or moai can act on the real profile or the worktree.
- The source reading finds no reach judgment in doctor. If the operator's profile-session doctor output also has none, AC-RC30-28 cannot pass on this code, whatever the migration keep (bundle E) does. This is a product gap for the leader to decide; nothing was patched.
- Audit § 2's byte-identity claim fails for 3 of 19 claude-side files against this tree's template, and for 10 of 19 in the Codex mirror. The audit may have used another template revision; that was not checked.
- `make build` ran its drift-check tests and the catalog generator under the real HOME, because a HOME redirect was refused. Go toolchain caches are written under the real home by default. I did not observe moai state written there, and the real `~/.moai` directory's own mtime (2026-10-10 11:48) predates the build. Writes inside its subdirectories were not audited.

**Files changed.** `.moai/specs/SPEC-UPDATE-REPAIR-RC30-001/progress.md` (this record). No other tracked file changed: `git status --short` was empty after the build and after the source reads. `bin/moai` is build output and is not tracked.

### M3 — config data model, restore, bundle listing, AC-RC30-22/23 investigations (manager-develop, card t1619)

Milestone M3 (plan § B.4 steps 1 to 5). Pre-flight, each its own command: `git rev-parse --short HEAD` gave `dd94d9570`; `git branch --show-current` gave `WT-moai-update-repair`; `git status --short` gave no output. Each item ran RED then GREEN, one commit per state. Commit SHAs are listed per item. Nothing was pushed.

**Status.** AC-RC30-05, -06, -07 and -08 pass with their judging commands on the final tree. AC-RC30-22 and AC-RC30-23 were **blocked** at M3, with the cause named below. The judging commands for both selected zero tests on the M3 tree. **F8 correction:** both are PASS at HEAD (see the F8 entry after Step 4).

**Environment.** Test runs touching a home or profile path ran as `unset CLAUDE_CONFIG_DIR MOAI_PROJECT_DIR && go test ...`. Every `internal/cli` run printed `moai-cli-test: userHomeDirFn redirected real home to sandbox`, so the cli HOME sandbox was active.

#### Step 5 — AC-RC30-05, bundle listing (U-04)

- Judging test: `TestBundleList` in `internal/cli/bundle_list_test.go`.
- RED. Command `go test ./internal/cli/ -run TestBundleList -count=1 -v`. Tree `dd94d9570` plus the new untracked test file. Exit status `1`. Verbatim assertion lines (the eleven list lines share one form, and the first and last are shown in full):

```text
    bundle_list_test.go:24: bundle backend: list line fields = [], want [backend 3 skills 0 agents]
    bundle_list_test.go:24: bundle testing: list line fields = [], want [testing 0 skills 0 agents]
    bundle_list_test.go:37: add --help names [], want the catalog bundles [backend commands consult creative delivery deployment design devops frontend ops-tools testing]
--- FAIL: TestBundleList (0.00s)
    --- FAIL: TestBundleList/list_names_each_bundle_with_counts (0.00s)
    --- FAIL: TestBundleList/add_help_names_valid_bundles (0.00s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.867s
```

  The cobra help body that the unknown `list` subcommand printed (17 lines) is elided. Commit `9249644b6`. No stub: the test drives `rootCmd`, so it compiles against the existing tree.
- GREEN. `runBundleList` and `sortedPackNames` in `internal/cli/bundle.go` print `%-10s %3d skills %3d agents` per optional pack, and `bundle add --help` names the packs through a help func (`renderBundleAddHelp`), so no catalog parse runs at process start. Verbatim: `--- PASS: TestBundleList (0.01s)` with both subtests PASS, `PASS`, `ok  	github.com/modu-ai/moai-adk/internal/cli	1.172s`. Exit `0`. Commit `8730ee5ba`.
- Related: `go test ./internal/cli/ -run TestBundle -count=1 -v` passed every `TestBundle*` test (`TestBundleAnnounceForPluginMode`, `TestBundleAddInstallsBundleEntries`, `TestBundleRemoveTakesComplementOnly`, `TestBundleRemoveHonorsDivergence`, `TestBundleUnknownNameRefused`, `TestBundleRemoveDefersDependency`).

#### Step 1 — AC-RC30-06, config round-trip (U-05), Gap G-06

- **G-06 resolved.** The field the round-trip sets is `llm.team_mode` (`LLMConfig.TeamMode`). Evidence from code and test: U-24 names `team_mode` as the field the audit saw reset from `glm` to `""` in a re-marshal; `launcher.go` reads it on the `moai cc` path (`resetTeamModeForCC`); `types.go:281-283` names it the explicit session intent. Rejected: `deployment_mode`, which is not a typed field and is written by `template.ApplyDeployMode`; and `harness`, which init rewrites on every run.
- Judging test: `TestConfigSaveRoundTripLLM` in `internal/config/llm_save_roundtrip_test.go`.
- RED. Command `unset CLAUDE_CONFIG_DIR MOAI_PROJECT_DIR && go test ./internal/config/ -run TestConfigSaveRoundTripLLM -count=1 -v`. Tree `8730ee5ba` plus the test file. Exit `1`. Verbatim excerpt (the got block is the typed re-marshal with the compiled defaults, and it is cut after its first lines):

```text
    llm_save_roundtrip_test.go:54: llm.yaml after a team_mode-only Save changed more than the field:
        got:
        llm:
            mode: ""
            team_mode: glm
            harness: claude
            glm_env_var: GLM_API_KEY
            ...
--- FAIL: TestConfigSaveRoundTripLLM (0.01s)
```

  The got block has no `deployment_mode`, no `future_flag`, and no comment lines. Commit `adc9cdb8b`.
- Branch RED. `internal/config/llm_save_patch_test.go` holds `TestLLMSavePatchBranches` with six subtests. I observed it with the patch path switched off in the working tree, so `saveLLM` took the pre-change typed rewrite. That path is what `adc9cdb8b` runs. Command `go test ./internal/config/ -run TestLLMSavePatchBranches -count=1 -v`, exit `1`. `--- FAIL` on `unchanged_section_leaves_bytes_alone` (the bytes changed), `nested_scalar_patched_in_place` (the change was not a one-line patch), `upserted_key_keeps_comments_and_unmodelled_keys` (`# llm header comment` lost), and `upserted_map_entry_keeps_comments_and_unmodelled_keys`. `--- PASS` on `quoted_value_falls_back_and_round_trips` and `removed_key_falls_back_and_round_trips`. Those two are regression guards for the fallback, which is unchanged. Commit `70fd61445`.
- GREEN. `saveLLM` in the new `internal/config/llm_save.go` patches a dirty llm section over an existing `llm.yaml` through `yamlpatch.PatchFile`. Each scalar that differs from a baseline captured at Load or Save becomes one edit, and new keys are upserted. It falls back to the typed rewrite for an absent file, a removed key, a sequence, or a value that must be quoted. The baseline is captured in `resetSectionDirtyLocked`. Verbatim: `--- PASS: TestConfigSaveRoundTripLLM (0.01s)`, `PASS`, `ok  	github.com/modu-ai/moai-adk/internal/config	0.192s`; the six branch subtests PASS, with `ok  	github.com/modu-ai/moai-adk/internal/config	0.323s`. Exit `0`. Commit `27af97cbf`.
- Gates: `gofmt -l` no output; `go vet ./internal/config/` no output; `golangci-lint run ./internal/config/...` reported `0 issues.`.
- Regression: `go test ./internal/config/ -run 'Save|LLM|Llm|ConfigManager|DeployMode|AgentOverride|Reload|SetSection|LoadRaw' -count=1` gave `ok  	github.com/modu-ai/moai-adk/internal/config	0.376s`.

#### Step 2 — AC-RC30-07, record-less archive warning (U-06)

- Reading of REQ-RC30-008. The migration still writes the local record, and it now announces the write when a migration archive exists. Rejected: skipping the write. That would change the SPEC-INIT-SHRINK-001 REQ-015 record contract, which is outside this item.
- Judging test: `TestRecordlessArchiveWarning` in `internal/cli/recordless_archive_test.go`.
- RED. Command `go test ./internal/cli/ -run TestRecordlessArchiveWarning -count=1 -v`. Tree `27af97cbf` plus the test and a stub. Exit `1`. Verbatim:

```text
    recordless_archive_test.go:50: record-less stamp over a migration archive printed no visible warning:
        ""
    recordless_archive_test.go:53: record after the stamp = "", want local
    recordless_archive_test.go:67: record after the stamp = "", want local
--- FAIL: TestRecordlessArchiveWarning (0.00s)
```

  Stub: `stampMigrationDeployMode` in `internal/cli/update_migrate.go` with an empty body, so the test compiles. Commit `86d188ec6`.
- GREEN. `stampMigrationDeployMode` owns the stamp. The template-sync call site now calls it, and `migrationArchiveExists` checks both archive roots (`.moai/archive/skills/init-shrink-migration` and `.moai/archive/files/init-shrink-migration`). Verbatim: `--- PASS: TestRecordlessArchiveWarning (0.01s)`, `ok  	github.com/modu-ai/moai-adk/internal/cli	1.078s`. Exit `0`. Commit `d7d71ae23`.
- Related: `go test ./internal/cli/ -run 'Migrat|DeployMode|Recordless' -count=1` gave `ok  	github.com/modu-ai/moai-adk/internal/cli	61.480s`.
- Gates: gofmt clean; `go vet ./internal/cli/` no output; `golangci-lint run ./internal/cli/` reported `0 issues.`.

#### Step 3 — AC-RC30-08, restore and retention (U-07)

- Judging test: `TestRestoreSettingsAndRetention` in `internal/cli/restore_retention_test.go`, subtests a to d. Subtest a (the live `settings.json` stays unchanged) was a regression guard that passed on the RED tree.
- RED 1. Command `go test ./internal/cli/ -run TestRestoreSettingsAndRetention -count=1 -v`. Tree `d7d71ae23` plus the test. Exit `1`. Verbatim assertion lines:

```text
    restore_retention_test.go:106: restore output does not name .claude/settings.json as not restored:
    restore_retention_test.go:111: manual copy instruction lacks the backup path ...
    restore_retention_test.go:114: manual copy instruction lacks the target path ...
    restore_retention_test.go:124: restore output does not list the uncovered backed-up path .gitignore ...
    restore_retention_test.go:155: run 0: namespace directory ... kept; every directory of an older run must be removed
    restore_retention_test.go:155: run 1: namespace directory ... kept; every directory of an older run must be removed
--- FAIL: TestRestoreSettingsAndRetention (0.05s)
```

  The paths are elided with `...`. Commit `034e49501`.
- RED 2, the legacy restore skips the marker. `TestRestoreLegacySkipsRunMarker` in `internal/cli/update/backup/restore_run_marker_test.go`. Command `go test ./internal/cli/update/backup/ -run TestRestoreLegacySkipsRunMarker -count=1 -v`. Tree `034e49501` plus the test, with the skip absent. Exit `1`. Verbatim: `restore_run_marker_test.go:35: the run marker was copied into the config tree (stat err: <nil>)` and `--- FAIL: TestRestoreLegacySkipsRunMarker (0.00s)`. Commit `177eb6426`.
- RED 3, the update writes the marker. `TestUpdateBackupCarriesRunMarker` in `internal/cli/update_run_marker_test.go`. I observed it with the config marker write disabled in the working tree (`error(nil)` in place of the write), which is the committed state of `177eb6426`. Command `unset CLAUDE_CONFIG_DIR MOAI_PROJECT_DIR && go test ./internal/cli/ -run TestUpdateBackupCarriesRunMarker -count=1 -v`. Exit `1`. Verbatim: `update_run_marker_test.go:34: the update's config backup under .../.moai-backups carries no .moai-run marker; retention would count it as its own run` and `--- FAIL: TestUpdateBackupCarriesRunMarker (0.30s)`. The write was restored before the commit. Commit `d8b3ca59c`.
- GREEN. Run identity is in `internal/cli/update/backup/run_marker.go` (`RunMarkerName`, `NewRunID`, `WriteRunMarker`, `readRunMarker`). Run grouping and the name shapes are in `run_retention.go`. `CleanupOldBackups` prunes by run across both backup roots, and a negative keep count removes nothing. The restore output is `reportUncoveredRestorePaths` in `internal/cli/update_restore.go`, with keys `update.restore.*` in `en.yaml` and `ko.yaml`. Marker writes sit in the template-sync Backup step, in the clean reinstall's Step 3 snapshot, and in its Step 4.5 config backup, with one run id per run. The legacy restore skips the marker. The destructive-site registry text matches the rule. Verbatim: `--- PASS: TestRestoreSettingsAndRetention (0.03s)` with subtests a, b, c and d PASS, `ok  	github.com/modu-ai/moai-adk/internal/cli	0.920s`; `TestRestoreLegacySkipsRunMarker` PASS, `ok  	github.com/modu-ai/moai-adk/internal/cli/update/backup	0.393s`; `TestUpdateBackupCarriesRunMarker` PASS, `ok  	github.com/modu-ai/moai-adk/internal/cli	1.328s`. Exit `0`. Commit `6bede8ca3`.
- Existing tests kept green: `TestCleanReinstall_PrunesOldConfigBackups` needs its clean-reinstall directories to share one run id, and that is the reason for the Step 3 and Step 4.5 marker writes. Also green: `TestCleanupOldBackups*` in `update_test.go`, `TestCleanupOldBackups*` in the backup package, and `TestRestoreMoaiConfig_MergeBehavior`.
- Sibling branch. `worktree-agent-a015199acfa2a92e1` holds commits `245c9fbb1`, `fef19384d`, `ac34bbec1` and `447a0a7e2` (card t1619, D9, merge base `2aab5f797`). They implement the same rule that `design.md` § 5 names. Those commits are not absorbed here, and `spec.md` § G.2 records no absorption of them. I used the branch as a design reference and wrote the code on this branch. Its RED test file was not imported. The absorption decision belongs to the leader.
- Related: `go test ./internal/cli/update/backup/ -count=1` gave `ok  	github.com/modu-ai/moai-adk/internal/cli/update/backup	0.675s`. `go test ./internal/tui/ -count=1` gave `ok  	github.com/modu-ai/moai-adk/internal/tui	0.275s`, which covers the catalog checks.
- Gates: gofmt clean; `go vet` on `internal/cli/update/backup/` and `internal/cli/` no output; `golangci-lint` `0 issues.` on both.

#### Step 4 — AC-RC30-22 and AC-RC30-23, investigation (U-23, U-24): BLOCKED at M3 (PASS at HEAD, see F8)

Both judging commands select zero tests on this tree. Verbatim, exit `0` for both:

```text
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/config	0.358s [no tests to run]     (AC-RC30-22, the command as written)
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.070s [no tests to run]     (AC-RC30-23, the command as written)
```

Under the AC ruling (d-20261010T040402Z-94b9, item 5) a run with no tests to run is a failure, not a pass.

**AC-RC30-22, `TestGateYAMLCommentsSurviveRestore` (U-23): blocked at M3, cause named; PASS at HEAD (F8).**

- Reproduction. Probes ran as scratch tests in `internal/cli`, because the cycle lives there. They were not committed, and the file was deleted after the run.
- The update cycle, `runTemplateSyncAt` on a fixture with a gate.yaml that carries 12 user comment lines. Verbatim: `PROBE gate.yaml comment lines before update: 12`, then `PROBE gate.yaml after update: user gate note lines=0, comment lines=66`, then `PROBE backup 20261010_221853 sections/gate.yaml user gate note lines: 12`. The backup keeps all 12. The live file keeps none.
- The restore step alone. The backup holds the user's file, the live tree holds a template-like file, and only `backup.RestoreMoaiConfigRetained` runs. Verbatim: `PROBE backup gate.yaml comment lines=12`, then `PROBE restore-alone: user gate note lines after restore=0`. The user's value `mode: strict` comes back, and the comments do not.
- Cause named. The section restore merge in `internal/cli/update/backup/restore.go` (`RestoreMoaiConfigRetained`, the `sections/*.yaml` walk) rebuilds the file from the template's structure and the user's values. It does not carry comment lines that are not attached to a template key. The restore alone reproduces the loss. The deploy step also writes the template's gate.yaml over the user's file. I did not separate that deploy overwrite from the restore on the full cycle, so the attribution to the restore rests on the isolated probe.
- Why it is blocked. The fix would change the three-way merge policy (`node_merge.go`), which is outside this item and outside `internal/config`. The judging command cannot reach it: a test in package `config` cannot import the backup package without a cycle through `config`. The leader decides the merge policy and the judging package.

**AC-RC30-23, `TestLLMYamlRemarshalCallers` (U-24): blocked at M3, premise does not hold; PASS at HEAD (F8).**

- Premise check. The AC and U-24 name "each `ConfigManager.Save` caller". Production has none. The seven production `.Save()` calls are all on `manifest.Manager`: `internal/core/project/validator.go:355` and `:441`, `internal/core/project/initializer.go:610`, `internal/cli/update_manifest_retrack.go:65`, `internal/cli/update_template_sync.go:652`, `internal/cli/migrate_project_assets.go:246`, `internal/cli/update_reconcile.go:159`, and `internal/codexwiring/write.go:348`. No production code calls `ConfigManager.SetSection`, so `llmDirty` is set only in tests.
- The writer that drops keys. `saveLLMSection` (`internal/cli/glm.go:708`) runs a typed `yaml.Marshal` of the whole llm section. It is reached from `persistTeamMode` (`glm.go:654`), from `disableTeamMode` (`glm.go:694`) through `resetTeamModeForCC` (`internal/cli/launcher.go:409`, the `moai cc` launch), and from `persistTeamMode(root, "glm")` (`launcher.go:304`, the `moai glm` launch). Its change-detection gate skips only unchanged writes.
- Reproduction, as a scratch probe (deleted after the run). The llm.yaml carried `deployment_mode: plugin`, `future_flag: keep-me`, a header comment and `team_mode: glm  # team note`. Verbatim:

```text
PROBE deployment_mode before persistTeamMode: "plugin"
PROBE deployment_mode after persistTeamMode(root, ""): ""
PROBE comment lines before=1 after=0
PROBE future_flag kept=false header kept=false
```

  So the `moai cc` team-mode reset clears the deployment record. The next update then reads the project as record-less (SPEC-INIT-SHRINK-001 REQ-015). This writer is outside `ConfigManager.Save`, so the AC-RC30-06 fix does not cover it.
- Why it is blocked. The AC as written judges an empty set, so no outcome of this run can pass it. The fix belongs to a decision. The natural form is to redirect AC-23 to the `saveLLMSection` callers (`persistTeamMode` with `glm` and with `""`), and then route `saveLLMSection` through the same yamlpatch seam that AC-06 uses. That is a change to the judging surface, so it is the leader's call.

#### Gates and regression (final tree)

- E2 `GOOS=windows GOARCH=amd64 go build ./...`: no output, exit `0`.
- E3 `gofmt -l` over the M3 Go files: no output, exit `0`. `go vet ./internal/config/ ./internal/cli/ ./internal/cli/update/backup/`: no output, exit `0`. `golangci-lint run ./internal/config/... ./internal/cli/ ./internal/cli/update/backup/`: `0 issues.`
- E5 regression, each a plain command: `go test ./internal/config/ -run 'Save|LLM|Llm|ConfigManager|DeployMode|AgentOverride|Reload|SetSection|LoadRaw' -count=1` gave `ok  	github.com/modu-ai/moai-adk/internal/config	0.376s`. `go test ./internal/cli/ -run 'Bundle|Recordless|RestoreSettings|UpdateBackupCarries|Cleanup|CleanReinstall|Restore|Migrat|DeployMode|Destructive|Backup' -count=1` gave `ok  	github.com/modu-ai/moai-adk/internal/cli	247.911s`. `go test ./internal/cli/update/backup/ ./internal/tui/ -count=1` gave `ok  	github.com/modu-ai/moai-adk/internal/cli/update/backup	0.675s` and `ok  	github.com/modu-ai/moai-adk/internal/tui	0.275s`. No test failed.
- Judging commands on the final tree (exit `0` each): AC-RC30-05 `--- PASS: TestBundleList`, AC-RC30-06 `--- PASS: TestConfigSaveRoundTripLLM`, AC-RC30-07 `--- PASS: TestRecordlessArchiveWarning`, AC-RC30-08 `--- PASS: TestRestoreSettingsAndRetention`. Each run was followed by `PASS` and an `ok` line.
- Coverage (M7) and `make tui-snapshot` (M4) were not run in M3.

#### Residual risk

- **The `moai cc` team-mode reset still drops `deployment_mode`.** `saveLLMSection` re-marshals the llm section on any team_mode change, and that drops the deployment record, unmodelled keys, and comments (probe above). This is the most consequential item here. The AC-RC30-06 fix does not reach it.
- U-05's fix covers `ConfigManager.Save` only. D7 describes the defect as a typed save of llm.yaml in general, which includes `saveLLMSection`.
- AC-08 marker coverage is thin. `TestUpdateBackupCarriesRunMarker` checks only the config backup of the template-sync path. The namespace snapshot of that path is marked in code but not asserted. The deploy pre-clean, symlink and legacy-memory backups, and the `ensureRunBackupDir` fallback, carry no marker, so each counts as its own run. The reference branch records the same residual.
- The `update.restore.*` keys exist in `en` and `ko` only. `ja` and `zh` are placeholder catalogs, so those languages show English.
- The gate.yaml comment loss (AC-RC30-22) is reproduced and unfixed. A user who edits gate.yaml loses its comment lines on every update.

#### Gaps (not observed in M3)

- The AC-22 and AC-23 judging commands select zero tests. The probes behind their evidence are scratch files, deleted, and not committed. Their verbatim output is quoted above.
- Two shell commands were refused by the worktree-isolation guard during the investigation. One was a `git` form with a `$(...)` path (`git merge-base` inside a compound), and one was a compound scratch-file append. Neither was a verification basis: the first was rerun as separate plain commands, and the second was replaced by a file write. Recorded here per verification-claim-integrity § 3.1. The guard refusals are in the failure log as `tool_failure:Bash` rows.
- The AC-22 probe did not separate the deploy step from the restore on the full cycle. The restore-alone probe isolates the restore.

#### Files changed (M3)

Go, production: `internal/cli/bundle.go`; `internal/config/manager.go`; `internal/config/llm_save.go` (new); `internal/cli/update_migrate.go`; `internal/cli/update_template_sync.go`; `internal/cli/update_restore.go`; `internal/cli/update_clean_install.go`; `internal/cli/update_destructive_registry.go`; `internal/cli/update/backup/backup.go`; `internal/cli/update/backup/restore.go`; `internal/cli/update/backup/run_marker.go` (new); `internal/cli/update/backup/run_retention.go` (new).

Go, tests: `internal/cli/bundle_list_test.go`; `internal/config/llm_save_roundtrip_test.go`; `internal/config/llm_save_patch_test.go`; `internal/cli/recordless_archive_test.go`; `internal/cli/restore_retention_test.go`; `internal/cli/update_run_marker_test.go`; `internal/cli/update/backup/restore_run_marker_test.go`.

Locale: `internal/tui/messages/en.yaml`, `internal/tui/messages/ko.yaml`.

This progress record: `.moai/specs/SPEC-UPDATE-REPAIR-RC30-001/progress.md` § E.2 (this entry).

#### Commits (M3, in order)

`9249644b6` RED AC-05 · `8730ee5ba` GREEN AC-05 · `adc9cdb8b` RED AC-06 · `70fd61445` RED AC-06 branches · `27af97cbf` GREEN AC-06 · `86d188ec6` RED AC-07 (test and stub) · `d7d71ae23` GREEN AC-07 · `034e49501` RED AC-08 · `177eb6426` RED AC-08 restore marker · `d8b3ca59c` RED AC-08 update marker · `6bede8ca3` GREEN AC-08.

#### Decisions (named)

1. AC-RC30-07 reading of REQ-RC30-008: the migration still writes the local record, and the write is announced.
2. Gap G-06 resolved as `llm.team_mode`, with the evidence above.
3. AC-RC30-06 design: patch per changed scalar against a baseline, and fall back to the typed rewrite for anything a scalar patch cannot express. The fallback still loses comments, and that is named under residual risk.
4. AC-RC30-05: the help is filled by a help func, so the catalog is not parsed at process start.
5. AC-RC30-08: re-implemented from `design.md` § 5 and not absorbed from the sibling branch. Retention counts runs across both roots, and the clean reinstall's two directories share one run id.
6. Locale keys in `en` and `ko` only. `ja` and `zh` fall back to English.
7. AC-RC30-22 and AC-RC30-23 were blocked at M3, with causes named. Nothing was fixed for them at M3; both are PASS at HEAD (F8 entry below).

#### F8 entry — AC-RC30-22 and AC-RC30-23 at HEAD (manager-develop, card t1619)

The blocked status recorded above at M3 is superseded. Both criteria are GREEN at HEAD. Their fixes landed in the commits listed below, which follow the M3 baseline `6bede8ca3`.

Each run uses a materialized copy of HEAD (`git checkout-index -a --prefix=<scratchpad>/head-step3/`), not the working tree. Command form: `go -C <head copy> test -count=1 -v -run '<regex>' ./internal/cli/`.

Command: `-run '^TestGateYAMLCommentsSurviveRestore$'`, exit: 0

```text
=== RUN   TestGateYAMLCommentsSurviveRestore
moai-cli-test: userHomeDirFn redirected real home to sandbox /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/moai-cli-home-3689689345
2026/10/11 02:10:07 INFO SUNSET_CONFIG_DORMANT_NOTICE spec="SPEC-V3R2-MIG-003 REQ-018" yaml_path=.moai/config/sections/sunset.yaml advice="SunsetConfig has no runtime hot path; activation requires a new SPEC"
--- PASS: TestGateYAMLCommentsSurviveRestore (0.35s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.902s
```

Command: `-run '^TestLLMYamlRemarshalCallers$'`, exit: 0

```text
=== RUN   TestLLMYamlRemarshalCallers
=== RUN   TestLLMYamlRemarshalCallers/resetTeamModeForCC
=== RUN   TestLLMYamlRemarshalCallers/persistTeamMode_glm
--- PASS: TestLLMYamlRemarshalCallers (0.00s)
    --- PASS: TestLLMYamlRemarshalCallers/resetTeamModeForCC (0.00s)
    --- PASS: TestLLMYamlRemarshalCallers/persistTeamMode_glm (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	0.705s
```

Commits after the M3 baseline. Each line is the output of `git log --oneline -1 <sha>`. The six listed are the AC-RC30-22 and AC-RC30-23 fix commits. They are a subset of the 24 commits in `6bede8ca3..HEAD` at this entry (`git rev-list --count 6bede8ca3..HEAD`).

```text
66c09f99b fix(SPEC-UPDATE-REPAIR-RC30-001): M3 AC-RC30-23 llm.yaml write keeps unrelated keys (card t1619)
04a11cb6f fix(SPEC-UPDATE-REPAIR-RC30-001): ConfigManager.Save keeps llm keys through the seam (card t1619)
68a70406c feat(SPEC-UPDATE-REPAIR-RC30-001): M3 AC-RC30-22 comment lines survive restore (card t1619)
d118b5c7e feat(SPEC-UPDATE-REPAIR-RC30-001): M3 AC-RC30-22 2-way comments follow the user per slot (card t1619)
fe3caa37c docs(SPEC-UPDATE-REPAIR-RC30-001): AC-RC30-22 scope excludes sequence-item comments; the exception names glm.models (card t1619)
2e54b0a16 test(SPEC-UPDATE-REPAIR-RC30-001): M3 AC-RC30-22 placement notes match the comment rules (card t1619)
```

Code commits of this card under the narrowed gate (scoped checks in the sections below):
- `942c729fb`: M4 remove forbidden archive summary literal (step 1; see the M4 entry).
- `84319245d`: M5 skip on recorded template tree, warn on stamp skew, count preview as run (step 2; see the M5 entry).

Gap: the full `internal/cli` suite and `go test ./...` were not run for this entry. The F8 runs cover the two named tests only. The repository-wide verdict belongs to CI on the integration branch, PENDING at this entry.

### M4 — measured archive summary, AC-RC30-09 (U-08, F1) (manager-develop, card t1619)

Scope of this entry: the literal `0 user customizations modified` only (AC-RC30-09 judges the literal). The rest of M4 (steps 1 to 7) is not in this entry.

#### RED — judging grep on the pre-change tree (exit 0)

Command: `grep -rn "0 user customizations modified" internal/cli --include='*.go' --exclude='*_test.go'`

```text
internal/cli/update_archive.go:282://	total: N skills archived, 0 user customizations modified
internal/cli/update_archive.go:367:	_, _ = fmt.Fprintln(out, tui.Pill(tui.PillOpts{Kind: tui.PillOk, Solid: false, Label: fmt.Sprintf("total: %d skills archived, 0 user customizations modified", archived), Theme: &th}))
```

#### RED — new test, before the change (exit 1)

Test added first: `TestArchiveSummaryCarriesMeasuredCountOnly` in `internal/cli/update_archive_summary_test.go`. It archives two legacy skills and checks the summary line carries the measured count and no customization claim. Command: `go test -count=1 -run '^TestArchiveSummaryCarriesMeasuredCountOnly$' -v ./internal/cli/`

```text
=== RUN   TestArchiveSummaryCarriesMeasuredCountOnly
=== PAUSE TestArchiveSummaryCarriesMeasuredCountOnly
=== CONT  TestArchiveSummaryCarriesMeasuredCountOnly
    update_archive_summary_test.go:37: summary line must not carry an unmeasured customization count, got:
        ✓  archive: moai-domain-db-docs  → .moai/archive/skills/v2.16/moai-domain-db-docs
        ✓  archive: moai-domain-mobile  → .moai/archive/skills/v2.16/moai-domain-mobile
         total: 2 skills archived, 0 user customizations modified 
--- FAIL: TestArchiveSummaryCarriesMeasuredCountOnly (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.007s
FAIL
```

The first assertion (`total: 2 skills archived`) held; the second (no `customizations`) failed on the literal.

#### GREEN — change

- `internal/cli/update_archive.go:282` (comment) and `:367` (Pill label): the summary is `total: %d skills archived`, built from the measured `archived` count (the count of successful archives in the run). The hardcoded clause is removed, not replaced with a number nothing measured.
- `internal/cli/tuxiu_characterization_test.go:201`: the expectation `"3 skills archived, 0 user customizations modified"` becomes `"3 skills archived"`. This is the one expectation change the audit names (E-22). `"Updated 26 files"` is unchanged.

Test-run output after the change (verbatim). exit: 0 for each block below. The step 1 re-run on the staged tree (commit `942c729fb`; the five named tests under `-race`, `go -C <scratchpad>/index-step1 test -count=1 -race -v -run '^(TestArchiveSummaryCarriesMeasuredCountOnly|TestTUXDataValuesPreserved|TestInitUpdateTUXCharacterization|TestTUXChannelPartition|TestTUXPostM4CarriesNewPresentation)$' ./internal/cli/`) exited 0 with all five PASS. The blocks below are the earlier run's text and are not byte-identical to that re-run.

```text
=== RUN   TestArchiveSummaryCarriesMeasuredCountOnly
=== PAUSE TestArchiveSummaryCarriesMeasuredCountOnly
=== CONT  TestArchiveSummaryCarriesMeasuredCountOnly
--- PASS: TestArchiveSummaryCarriesMeasuredCountOnly (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	0.886s
```

```text
=== RUN   TestTUXDataValuesPreserved
--- PASS: TestTUXDataValuesPreserved (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	0.730s
```

#### GREEN — judging grep on the changed tree (exit 1)

Command: `grep -rn "0 user customizations modified" internal/cli --include='*.go' --exclude='*_test.go'` — no output. The same command followed by `; echo "exit=$?"` printed `exit=1`.

#### Checks

- `go build ./...`: no output (success).
- `go vet ./internal/cli/`: no output (success).
- `gofmt -l` on the three Go files of this step: no output.

#### Gaps (not observed in this entry)

- The recorded golden fixtures under `internal/cli/testdata/tuxiu/` (pre-logo and `postm4/`, six files) still carry the old literal. AC-RC30-09 judges `*.go` files only (`--include='*.go'`), so these are outside its judged scope. They were not regenerated in this step. No step names an owner for regenerating these `testdata/tuxiu` fixtures: the golden regeneration of M4 step 7 is limited to `make tui-snapshot`, which covers `internal/tui` (E-09). The leader decides the owner.
- `internal/cli/update_archive.go:396` (`dryRunArchiveLegacySkills`) prints the literal `0 will be archived` in a format string. It is the same class as REQ-RC30-011, but it is not the string AC-RC30-09 judges, and U-08 does not list it. It is not changed in this step.

#### Residual risk

- The summary no longer prints a customization count. The archive step never measured one, so the removed clause reported nothing; the drift backup warning (`archive drift backup: …`) still prints for each archived copy that was moved aside under `--force`.

#### Files changed (M4 step 1)

`internal/cli/update_archive.go`; `internal/cli/tuxiu_characterization_test.go`; `internal/cli/update_archive_summary_test.go` (new); this progress record.

The commit SHA of this step cannot be named in its own commit. It is recorded in the § E.2 F8 entry (the step 3 record).

### M5 — skip decision, stamp skew, preview count, AC-RC30-14/15/19 (U-13, U-14, U-18, F2) (manager-develop, card t1619)

#### Design (named)

1. U-13, AC-RC30-14: a version-matched run skips only when the recorded template tree equals the embedded tree. The recorded tree is `Manifest.TemplateTreeHash` (`internal/manifest/types.go`), written by the deploy step just before its manifest save; the value is `template.ComputeDirTreeHash` over the embedded FS. A project with no recorded tree keeps the version-only decision. That keeps `TestSkipSyncNoArchive/version_match_returns_skipped_true` (a fixture with only `system.yaml`) green, and the first deploy after this change records a tree. A manifest that fails to load does not skip.
2. U-14, AC-RC30-15: after the run's steps, `warnStampSkew` compares `moai.template_version` in the deployed `system.yaml` with `version.GetVersion()` and prints one warning line on disagreement. No `system.yaml` means nothing to compare, so no warning.
3. U-18, AC-RC30-19: the preview builds its deployer through `newTemplateSyncDeployer` with the run's deploy mode. It no longer uses `template.NewDeployerWithForceUpdate`. The run path's confirmation call now goes through the existing `confirmViaPreviewFn` seam, which defaults to the same function. The test needs the seam so that it does not block on the interactive confirmation.
4. RED stage: the API stubs (the manifest field, and a no-op `warnStampSkew`) were added first so the tests compile and fail on behavior, not on a build error.

#### RED — three named tests, before the implementation (exit 1)

Command: `go test -count=1 -run 'TestSkipDecisionUsesTreeHash|TestStampSkewWarning|TestPreviewCountEqualsRunCount' -v ./internal/cli/`

```text
=== RUN   TestSkipDecisionUsesTreeHash
=== RUN   TestSkipDecisionUsesTreeHash/same_version_same_tree_skips
=== RUN   TestSkipDecisionUsesTreeHash/same_version_different_tree_does_not_skip
    update_template_sync_measured_test.go:114: skipped = true, want false; output:
        
         ✓ Up to date · Skipping sync 
--- FAIL: TestSkipDecisionUsesTreeHash (0.02s)
    --- PASS: TestSkipDecisionUsesTreeHash/same_version_same_tree_skips (0.00s)
    --- FAIL: TestSkipDecisionUsesTreeHash/same_version_different_tree_does_not_skip (0.00s)
=== RUN   TestStampSkewWarning
=== PAUSE TestStampSkewWarning
=== RUN   TestPreviewCountEqualsRunCount
    update_template_sync_measured_test.go:193: preview count 268 != run count 423 for the same fixture
--- FAIL: TestPreviewCountEqualsRunCount (0.00s)
=== CONT  TestStampSkewWarning
=== RUN   TestStampSkewWarning/stale_stamp_on_disk_warns
=== PAUSE TestStampSkewWarning/stale_stamp_on_disk_warns
=== RUN   TestStampSkewWarning/matching_stamp_is_silent
=== PAUSE TestStampSkewWarning/matching_stamp_is_silent
=== RUN   TestStampSkewWarning/missing_config_is_silent
=== PAUSE TestStampSkewWarning/missing_config_is_silent
=== CONT  TestStampSkewWarning/stale_stamp_on_disk_warns
=== CONT  TestStampSkewWarning/missing_config_is_silent
=== CONT  TestStampSkewWarning/matching_stamp_is_silent
=== NAME  TestStampSkewWarning/stale_stamp_on_disk_warns
    update_template_sync_measured_test.go:138: a stamp that disagrees with the deployed bytes must warn, got:
--- FAIL: TestStampSkewWarning (0.00s)
    --- PASS: TestStampSkewWarning/missing_config_is_silent (0.00s)
    --- FAIL: TestStampSkewWarning/stale_stamp_on_disk_warns (0.00s)
    --- PASS: TestStampSkewWarning/matching_stamp_is_silent (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.054s
FAIL
```

The positive controls passed on the RED tree (`same_version_same_tree_skips`, `matching_stamp_is_silent`, `missing_config_is_silent`). The failures are the three defects: the skip ignores the tree, the preview counts 268 files against the run's 423 on the same fixture, and a stale stamp prints nothing.

#### GREEN — the same three tests (exit 0)

Command: the same as RED.

```text
=== RUN   TestSkipDecisionUsesTreeHash
=== RUN   TestSkipDecisionUsesTreeHash/same_version_same_tree_skips
=== RUN   TestSkipDecisionUsesTreeHash/same_version_different_tree_does_not_skip
--- PASS: TestSkipDecisionUsesTreeHash (0.04s)
    --- PASS: TestSkipDecisionUsesTreeHash/same_version_same_tree_skips (0.01s)
    --- PASS: TestSkipDecisionUsesTreeHash/same_version_different_tree_does_not_skip (0.01s)
=== RUN   TestStampSkewWarning
=== PAUSE TestStampSkewWarning
=== RUN   TestPreviewCountEqualsRunCount
--- PASS: TestPreviewCountEqualsRunCount (0.00s)
=== CONT  TestStampSkewWarning
=== RUN   TestStampSkewWarning/stale_stamp_on_disk_warns
=== PAUSE TestStampSkewWarning/stale_stamp_on_disk_warns
=== RUN   TestStampSkewWarning/matching_stamp_is_silent
=== PAUSE TestStampSkewWarning/matching_stamp_is_silent
=== RUN   TestStampSkewWarning/missing_config_is_silent
=== PAUSE TestStampSkewWarning/missing_config_is_silent
=== CONT  TestStampSkewWarning/stale_stamp_on_disk_warns
=== CONT  TestStampSkewWarning/missing_config_is_silent
=== CONT  TestStampSkewWarning/matching_stamp_is_silent
--- PASS: TestStampSkewWarning (0.00s)
    --- PASS: TestStampSkewWarning/missing_config_is_silent (0.00s)
    --- PASS: TestStampSkewWarning/matching_stamp_is_silent (0.00s)
    --- PASS: TestStampSkewWarning/stale_stamp_on_disk_warns (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.059s
```

#### Checks and regression (exact-name runs, each a single command)

- `gofmt -l` on the four M5 Go files: no output. `gofmt -w` was applied to `internal/cli/update_template_sync.go` once, to fix the indentation of two hunks this step inserted (the save-line and the preview block); no other line changed.
- `go build ./...`: no output (success).
- `go vet ./internal/cli/`: no output (success).
- `go test -count=1 ./internal/manifest/`: `ok  	github.com/modu-ai/moai-adk/internal/manifest	0.339s`.
- `go test -count=1 -run 'TestSkipSyncNoArchive/version_match_returns_skipped_true' -v ./internal/cli/`: `--- PASS: TestSkipSyncNoArchive/version_match_returns_skipped_true`, `ok  	github.com/modu-ai/moai-adk/internal/cli	0.717s`.
- `go test -count=1 -run 'TestTemplateSync_|TestCleanReinstall_SettingsSnapshotStagedBeforeMerge' -v ./internal/cli/`: six tests run and all PASS (`TestTemplateSync_LeavesNoEmptyManagedSkillDirs`, `TestTemplateSync_ManagedSweepFlagsPlantedEmptyDir`, `TestTemplateSync_MirrorNoticeGoesToStderrNotStdout`, `TestTemplateSync_NoFallbackEmitsNothing`, `TestTemplateSync_PlainDeployerStillDeploys`, `TestCleanReinstall_SettingsSnapshotStagedBeforeMerge`), `ok  	github.com/modu-ai/moai-adk/internal/cli	1.525s`.

#### Gaps (not observed in this step)

- The full `internal/cli` suite and `golangci-lint` were not run (the lead's constraint; the gate that would run them is blocked on the lead's decision, see the F8 record).
- The stamp warning is covered through `warnStampSkew` directly. Its call at the end of the run path is not covered by an end-to-end run of `moai update`.
- The tree record is observed only on test fixtures, not on a real project through `moai update`.
- The cost of `embeddedTemplateTreeHash` on each version-matched run was not measured.

#### Residual risk

- A project whose manifest holds no recorded tree keeps the version-only skip until its first deploy with this binary. A same-version tree change on such a project still skips once.
- The stamp check reports a disagreement; it does not repair one.

#### Files changed (M5 step 2)

`internal/cli/update_template_sync.go`; `internal/cli/update_template_sync_guard.go` (new); `internal/cli/update_template_sync_measured_test.go` (new); `internal/manifest/types.go`; this progress record.

The commit SHA of this step cannot be named in its own commit. It is recorded in the § E.2 F8 entry (the step 3 record).

#### Consumer run 1, M5 identifiers: 60 functions (lane-14, round 2A)

The 17:15Z watchdog entry in `.moai/reports/t1619/progress.md` recorded that this entry lacked the M5 consumer run. The lane ran the set itself under the heavy-run slot. This entry records that run with its provenance per field. Round 2A did not repeat the run (leader instruction: repeat only if the record is wrong).

- Consumer set: `.moai/reports/t1619/m5-consumer-tests.txt`, line 6 (the `-run` regex). Lines 2-3 name the M5 tokens (`TemplateTreeHash`, `warnStampSkew`, `confirmViaPreviewFn`, `newTemplateSyncDeployer`, `runTemplateSyncWith`, `TestSkipSyncNoArchive`); line 5 selects 60 test functions.
- Command shape: `go test -count=1 -race -run '<line 6 regex>' ./internal/cli/`, the form recorded by the lane in the 17:15Z watchdog entry (`.moai/reports/t1619/progress.md`, the entry whose step reads "보완 (레인 실행)").
- Slot: `internal-cli-tests`, foreground. lane-14 ran the set under the heavy-run slot and acquired and released the slot itself.
- Result line (as recorded by the lane, same source): `ok  	github.com/modu-ai/moai-adk/internal/cli	131.178s`
exit: 0

Gap G-M5-C1: the stdout of that run is not retained on disk. The result line and exit code above are carried from the lane's recorded result. They were not re-observed in round 2A, and no verbatim output is quoted here. A repeat costs about 131 s under the same slot.

Residual risk: the run covers the 60 named functions only, not the full `internal/cli` suite (local scope rule). The repository-wide verdict is owned by CI on `origin/develop` and is PENDING at the time of this record.

#### Consumer run 2, M5 identifiers: 105 functions (writer's run, relayed)

**Provenance first.** This result was not observed in round 2A. The writer's run is session scratch and is not in the repository. team-lead relayed its counts and final line for this record (round-2A addition, instruction: do not re-run the 105 set unless a record is wrong). The names set is verified here against its file.

- Consumer set: `.moai/reports/t1619/m5-consumers-105.txt`, last line (the `-run` regex). The file has 6 lines. Its last line names 105 test functions, all unique. Round 2A verified this by counting the names on that line (`grep -o "Test[A-Za-z0-9_]*"`, then `wc -l`: 105 and 105 after `sort -u`).
- Counts (relayed): 105 PASS, 0 FAIL, 0 SKIP.
- Final line (relayed, as given by team-lead): `ok  github.com/modu-ai/moai-adk/internal/cli  270.581s`
- Exit code: not relayed separately. The `ok` final line implies a passing package run (exit 0); this is an inference, not a relayed value.

Gap G-M5-C2: the writer's log is not in the repository, so the counts and the final line are carried from the relay and were not re-observed in round 2A. The names set is observed (105 names in the file).

Residual risk: the run covers the 105 named functions only, not the full `internal/cli` suite (local scope rule). The repository-wide verdict is owned by CI on `origin/develop` and is PENDING at the time of this record.

### Scoped gate evidence, M4 and M5 (narrowed gate, card t1619)

Basis: the lead's narrowed-gate decision. The scope is the named tests, plus every test whose assertions reference a changed identifier or output string. Runs are in the foreground, one `go test` per command. Exit codes are separate fields. The tool printed no exit-code error for the runs below, so each is recorded as `exit=0`.

#### Repo-wide grep (narrowed gate, item 2)

Command: `grep -rln -E 'skills archived|archiveLegacySkills|runTemplateSyncWith|confirmViaPreviewFn|versionMatchSkipsSync|warnStampSkew|recordTemplateTree|embeddedTemplateTreeHash|TemplateTreeHash|template_tree_hash|template stamp|Up to date|Already up-to-date|Found [0-9]+ files|total: ' internal --include='*_test.go'`

Output: 35 files. Piped to `wc -l`, the same command printed `35`.

```text
internal/cli/update_idempotency_test.go
internal/cli/update_home_seam_test.go
internal/cli/doctor_codex_test.go
internal/cli/update_characterization_test.go
internal/cli/update_archive_flow_test.go
internal/cli/update_codex_wiring_test.go
internal/cli/update_archive_summary_test.go
internal/cli/update_integrity_probe_test.go
internal/cli/tuxiu_characterization_test.go
internal/cli/update_hooks_guidance_test.go
internal/cli/update_template_sync_measured_test.go
internal/cli/update_claude_to_both_test.go
internal/cli/update_identity_preserve_test.go
internal/cli/coverage_improvement_test.go
internal/cli/update_mirror_notice_test.go
internal/cli/update_fileops_test.go
internal/cli/update_integrity_probe_codex_test.go
internal/cli/update_archive_force_test.go
internal/cli/harness_profile_transition_test.go
internal/cli/update_settings_snapshot_test.go
internal/cli/update_archive_continue_test.go
internal/cli/update_model_key_strip_test.go
internal/cli/m0_surface_red_test.go
internal/cli/update_observability_test.go
internal/cli/update_repair_test.go
internal/cli/update_skip_sync_test.go
internal/cli/update_identity_unrenderable_test.go
internal/cli/update_llm_preserve_test.go
internal/cli/update_dry_run_test.go
internal/cli/update_archive_test.go
internal/cli/update_migrate_test.go
internal/cli/target_coverage_test.go
internal/cli/update/report/outcome_test.go
internal/auditreceipt/store_edge_test.go
internal/guardstate/axis_test.go
```

#### Named tests, foreground, race detector

Command: `go test -count=1 -race -run '^(TestArchiveSummaryCarriesMeasuredCountOnly|TestTUXDataValuesPreserved|TestSkipDecisionUsesTreeHash|TestStampSkewWarning|TestPreviewCountEqualsRunCount)$' -v ./internal/cli/` — exit=0

```text
=== RUN   TestTUXDataValuesPreserved
--- PASS: TestTUXDataValuesPreserved (0.00s)
=== RUN   TestArchiveSummaryCarriesMeasuredCountOnly
=== PAUSE TestArchiveSummaryCarriesMeasuredCountOnly
=== RUN   TestSkipDecisionUsesTreeHash
=== RUN   TestSkipDecisionUsesTreeHash/same_version_same_tree_skips
=== RUN   TestSkipDecisionUsesTreeHash/same_version_different_tree_does_not_skip
--- PASS: TestSkipDecisionUsesTreeHash (0.23s)
    --- PASS: TestSkipDecisionUsesTreeHash/same_version_same_tree_skips (0.07s)
    --- PASS: TestSkipDecisionUsesTreeHash/same_version_different_tree_does_not_skip (0.09s)
=== RUN   TestStampSkewWarning
=== PAUSE TestStampSkewWarning
=== RUN   TestPreviewCountEqualsRunCount
--- PASS: TestPreviewCountEqualsRunCount (0.06s)
=== CONT  TestArchiveSummaryCarriesMeasuredCountOnly
=== CONT  TestStampSkewWarning
=== RUN   TestStampSkewWarning/stale_stamp_on_disk_warns
=== PAUSE TestStampSkewWarning/stale_stamp_on_disk_warns
=== RUN   TestStampSkewWarning/matching_stamp_is_silent
=== PAUSE TestStampSkewWarning/matching_stamp_is_silent
=== RUN   TestStampSkewWarning/missing_config_is_silent
=== PAUSE TestStampSkewWarning/missing_config_is_silent
=== CONT  TestStampSkewWarning/stale_stamp_on_disk_warns
=== CONT  TestStampSkewWarning/missing_config_is_silent
=== CONT  TestStampSkewWarning/matching_stamp_is_silent
--- PASS: TestStampSkewWarning (0.00s)
    --- PASS: TestStampSkewWarning/missing_config_is_silent (0.00s)
    --- PASS: TestStampSkewWarning/stale_stamp_on_disk_warns (0.00s)
    --- PASS: TestStampSkewWarning/matching_stamp_is_silent (0.00s)
--- PASS: TestArchiveSummaryCarriesMeasuredCountOnly (0.01s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	3.271s
```

Command: `go test -count=1 -race -run 'TestSkipSyncNoArchive/^version_match_returns_skipped_true$' -v ./internal/cli/` — exit=0

```text
=== RUN   TestSkipSyncNoArchive
=== RUN   TestSkipSyncNoArchive/version_match_returns_skipped_true
--- PASS: TestSkipSyncNoArchive (0.01s)
    --- PASS: TestSkipSyncNoArchive/version_match_returns_skipped_true (0.01s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	2.985s
```

#### Static checks (foreground)

- `gofmt -l` on the seven touched Go files (`internal/cli/update_archive.go`, `internal/cli/tuxiu_characterization_test.go`, `internal/cli/update_archive_summary_test.go`, `internal/cli/update_template_sync.go`, `internal/cli/update_template_sync_guard.go`, `internal/cli/update_template_sync_measured_test.go`, `internal/manifest/types.go`): no output.
- `go build ./...`: no output.
- `go vet ./internal/cli/ ./internal/manifest/`: no output.
- `golangci-lint run ./internal/cli/ ./internal/manifest/`: `0 issues.`

#### Gaps

- The 35 files above are not run. Their test functions were not run, so the grep-derived set is incomplete. The `TestSkipSyncNoArchive/skip_sync_with_force_does_invoke_archive` subtest is in that set, and it was not run: its own comment says it reaches the interactive confirmation.
- The full-suite `go test` for the touched packages was not run locally (lane protocol §8). Its verdict comes from CI after the leader's batch push.
- `/verify` was not invoked whole. Its Go test row is `go test -race` with no package scope, launched in the background (`workflows/gate.md` Phase 1 and Phase 2). Whether the `--file` scope narrows the Go test run was not verified. The static checks above were run by hand.
- Phase 1B snapshot: absent. `moai verify check --key-current` exited 1, with `"fresh": false` and reason `no snapshot recorded for the current working-tree key`.
- Hook: whether a pre-commit hook runs is not relied on. `git config --get core.hooksPath` printed `/dev/null`, and the lane record says otherwise. A commit that prints hook output will quote it verbatim with its exit code.

### M6 — round 2A: U-15 (AC-RC30-16) and U-16 (AC-RC30-17) (manager-develop, card t1619)

**Tree and scope.** HEAD `51f957864` (the M5 record) on `WT-moai-update-repair`, clean before this round. The lead assigned U-15 and U-16 to this round as M6. Plan § B.7 places U-15 under M6; plan § B.5 step 5 places U-16 under M4 (AC-RC30-17 names M4 as its TDD RED milestone). The round follows the lead's assignment, and this entry keeps the plan's milestone labels.

**Commit order.** This record commits first, on its own (docs only: the root cause and the RED outputs). The two test files and the U-16 API stub are the working-tree state these RED outputs were measured on. They commit with the code commit that follows, together with the GREEN implementation, so no commit carries a failing test. The GREEN record commits last. A baseline measured before a change must land in its own earlier commit (`verification-claim-integrity.md` §2.3); this record is that commit, and the code commit follows it.

**Status at the close of round 2A.** U-15 is GREEN and committed (`c9ab1903f`). U-16 is BLOCKED: its GREEN breaks four existing consumer expectations, so it is not committed (see § U-16 GREEN attempt and blocker). The M5 consumer-run record is `51f957864`; the step 2 M5 code is `84319245d`.

#### U-15 root cause (AC-RC30-16, REQ-RC30-024), recorded before the test

Commands and verbatim output, in the order run.

`git log --all --oneline -S TestPermissionsAskPreserved` (exit 0)

```text
18d9f6087 test(SPEC-UPDATE-REPAIR-RC30-001): pin the permissions retired-array restore scope (card t1619)
edb327fc0 fix(SPEC-UPDATE-REPAIR-RC30-001): D11 permissions.ask preserved on merge (card t1619)
520e2e0df docs(SPEC-UPDATE-REPAIR-RC30-001): add plan-phase SPEC draft (card t1619)
```

`git branch -a --contains edb327fc0` (exit 0)

```text
+ worktree-agent-a51849d65be0c3627
```

`git merge-base --is-ancestor edb327fc0 HEAD` (exit 1: the fix commit is not in this branch's history)

`git show --stat --format=%H%n%s%n%b edb327fc0` (stat lines, verbatim)

```text
 .../update/merge/permissions_ask_preserved_test.go | 67 ++++++++++++++++++
 internal/merge/retired_user_array_test.go          | 81 ++++++++++++++++++++++
 internal/merge/strategies.go                       | 61 +++++++++++++++-
 3 files changed, 207 insertions(+), 2 deletions(-)
```

`grep -rn "permissions" internal/merge/ --include='*.go' --exclude='*_test.go'` (no output: HEAD's merge engine has no permission-specific handling)

`grep -rn "mergeJSON(\|NewEngine()\|JSONMerge" internal/merge --include='*.go' --exclude='*_test.go'` (exit 0)

```text
internal/merge/types.go:23:	// JSONMerge performs object-level merge for JSON files.
internal/merge/types.go:24:	JSONMerge MergeStrategy = "json_merge"
internal/merge/strategies.go:83:		return JSONMerge
internal/merge/strategies.go:302:func mergeJSON(base, current, updated []byte) (*MergeResult, error) {
internal/merge/strategies.go:326:		Strategy:     JSONMerge,
internal/merge/three_way.go:14:func NewEngine() Engine {
internal/merge/three_way.go:58:	case JSONMerge:
internal/merge/three_way.go:59:		return mergeJSON(base, current, updated)
```

**Code path (HEAD `51f957864`).** `internal/cli/update/merge/merge.go:239` `MergeUserFilesWithOutcome` → `merge.go:257` `engine := mrg.NewEngine()` → `internal/merge/three_way.go:58-59` (JSONMerge calls `mergeJSON`) → `internal/merge/strategies.go:315` `deepMergeMap(baseMap, currentMap, updatedMap, "")`. `.claude/settings.json` is a JSON file, so `strategies.go:83` selects JSONMerge for it.

**Cause, read from the code.** `deepMergeMap` (`internal/merge/strategies.go:366`) decides each top-level key by its three-way state:

- `strategies.go:421-424` (`inBase && inCurrent && !inUpdated`, "Template removed key") keeps the user's value and records the key in `retained`. This is the per-key rule that keeps a key the template removed.
- `strategies.go:439-441` (`!baseChanged && updChanged`, "Only template changed") runs `result[key] = updVal`. For `permissions`, the user left the object unchanged and the template changed it, so the whole template object replaces the user's. The template object has no `ask`, so `ask` is dropped. The recursion that would apply the per-key rule to the children of `permissions` runs only in the both-changed branch (`strategies.go:454-455`), not in this branch.

The cause is the wholesale assignment at `strategies.go:441` for a map the user did not change. The three-way decision is sound for top-level keys. This turns the "inferred" cause in spec.md § U-15 into an established one: the mechanism is the 3-way merge, and the defect is the wholesale branch.

**Corroboration, not a source of the fix.** The F-2 commit `edb327fc0` states the same cause ("The wholesale replacement for a key the user left unchanged skipped the per-key rule that keeps a key the template removed"). Its stat is recorded above. Its diff was not read; the fix below is written from HEAD's code and the U-15 fixture.

**Fixture (spec.md § U-15, next step).** The base is the canonical settings snapshot with `permissions.ask: [sudo, chmod, chown]`. The template render has no `ask` key. The user copy carries the same `ask` list, unchanged from the base. The test runs the production path (`MergeUserFiles`, via `mergeSettings`), the same route as `settings_snapshot_base_test.go`.

#### U-15 RED (AC-RC30-16), before the fix

Test: `TestPermissionsAskPreserved` in `internal/cli/update/merge/permissions_ask_preserved_test.go` (new file). Fixture as in the root cause above. It runs through `mergeSettings` (`settings_snapshot_base_test.go`), so it reaches the production `MergeUserFiles` path, not the engine alone.

Command: `go test -count=1 -race -run '^TestPermissionsAskPreserved$' -v ./internal/cli/update/merge/`

```text
=== RUN   TestPermissionsAskPreserved
    permissions_ask_preserved_test.go:29: permissions.ask = <nil>, want [sudo chmod chown]
--- FAIL: TestPermissionsAskPreserved (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli/update/merge	0.417s
FAIL
```

exit: 1 (this is the re-run on the same tree; the first run printed the same lines with duration 0.512s)

The failure is behavioral. The test ran, and the merged file has no `permissions.ask`. It is not a build error. The cause is the branch at `internal/merge/strategies.go:439-441`.

#### U-16 RED (AC-RC30-17), before the change

**Scope reading.** REQ-RC30-015 says the update prints per-category counts "in place of the file list". This round reads that as the preserved list (the 369-line list of D13), which is what the test name and D13 name. Archived removals stay listed by path, because REQ-UPM-031 keeps deletions visible. Conflicts stay listed, because each names a sidecar the user must act on. The alternative reading (collapse every list) is named in the round report for the leader to decide.

**API stub (working tree; it commits with the code commit that follows this record).** `RenderReconciliation` (`internal/cli/update/report/outcome.go`) gains a `verbose bool` parameter that the stub ignores, so the output is unchanged. Call sites: `internal/cli/update_tux.go` passes `updateVerboseMode`; `TestRenderReconciliation` (`internal/cli/update/report/outcome_test.go`) and `internal/cli/update/reconcile_test.go` pass `true`, which keeps their full-list expectations. No assertion in those tests changed. The stub makes the RED behavioral: without it, the test would not compile.

Test: `TestPreservedLinesCountOnly` in `internal/cli/update/report/preserved_lines_test.go` (new file).

Command: `go test -count=1 -race -run '^TestPreservedLinesCountOnly$' -v ./internal/cli/update/report/`

```text
=== RUN   TestPreservedLinesCountOnly
    preserved_lines_test.go:22: 
        	Error Trace:	/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1619/internal/cli/update/report/preserved_lines_test.go:22
        	Error:      	Not equal: 
        	            	expected: 1
        	            	actual  : 3
        	Test:       	TestPreservedLinesCountOnly
    preserved_lines_test.go:23: 
        	Error Trace:	/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1619/internal/cli/update/report/preserved_lines_test.go:23
        	Error:      	"Reconciliation: 1 refreshed, 0 merged, 0 conflict(s), 3 preserved, 1 archived-removed\n  archived-removed: .claude/rules/moai/old-rule.md\n  preserved: .claude/rules/moai/a.md\n  preserved: .claude/rules/moai/b.md\n  preserved: .claude/rules/moai/c.md" does not contain "\n  preserved: 3 file(s) (run with --verbose to list)"
        	Test:       	TestPreservedLinesCountOnly
    preserved_lines_test.go:25: 
        	Error Trace:	/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1619/internal/cli/update/report/preserved_lines_test.go:25
        	Error:      	"Reconciliation: 1 refreshed, 0 merged, 0 conflict(s), 3 preserved, 1 archived-removed\n  archived-removed: .claude/rules/moai/old-rule.md\n  preserved: .claude/rules/moai/a.md\n  preserved: .claude/rules/moai/b.md\n  preserved: .claude/rules/moai/c.md" should not contain ".claude/rules/moai/a.md"
        	Test:       	TestPreservedLinesCountOnly
    preserved_lines_test.go:25: 
        	Error Trace:	/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1619/internal/cli/update/report/preserved_lines_test.go:25
        	Error:      	"Reconciliation: 1 refreshed, 0 merged, 0 conflict(s), 3 preserved, 1 archived-removed\n  archived-removed: .claude/rules/moai/old-rule.md\n  preserved: .claude/rules/moai/a.md\n  preserved: .claude/rules/moai/b.md\n  preserved: .claude/rules/moai/c.md" should not contain ".claude/rules/moai/b.md"
        	Test:       	TestPreservedLinesCountOnly
    preserved_lines_test.go:25: 
        	Error Trace:	/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1619/internal/cli/update/report/preserved_lines_test.go:25
        	Error:      	"Reconciliation: 1 refreshed, 0 merged, 0 conflict(s), 3 preserved, 1 archived-removed\n  archived-removed: .claude/rules/moai/old-rule.md\n  preserved: .claude/rules/moai/a.md\n  preserved: .claude/rules/moai/b.md\n  preserved: .claude/rules/moai/c.md" should not contain ".claude/rules/moai/c.md"
        	Test:       	TestPreservedLinesCountOnly
--- FAIL: TestPreservedLinesCountOnly (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli/update/report	0.170s
FAIL
```

exit: 1 (this is the re-run on the same tree; the first run printed the same lines with duration 0.326s)

The failure is behavioral. The verbose assertions (the full list under `verbose=true`) already pass on the stub, which is expected: the stub prints the full list in every mode, so the RED sits on the default-mode assertions only.

#### U-15 GREEN (AC-RC30-16), the fix (commit `c9ab1903f`)

**Fix.** `internal/merge/strategies.go`. In `deepMergeMap`, the branch "Only template changed" keeps the user's arrays under the top-level `permissions` object that the template no longer carries (lines 442-448, which call the helper). The constant `permissionsKey` (line 489) and the helper `keepRetiredArrays` (line 496) are new. Objects and hook arrays keep the template's value, as before (REQ-RC30-024). **Test:** `internal/cli/update/merge/permissions_ask_preserved_test.go` (new file, RED recorded above).

Targeted GREEN, command `go test -count=1 -race -run '^TestPermissionsAskPreserved$' -v ./internal/cli/update/merge/`:

```text
=== RUN   TestPermissionsAskPreserved
--- PASS: TestPermissionsAskPreserved (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli/update/merge	1.590s
```

exit: 0

**Consumers, repo-wide grep over `internal/**/*_test.go` (verbatim).**

- `grep -rn "permissions" internal/merge --include='*_test.go'`: no output. No existing test pins the permissions merge.
- `grep -rln "deepMergeMap\|mergeJSON" internal --include='*_test.go'`: `internal/merge/coverage_extra_test.go`, `internal/merge/strategies_test.go`, `internal/cli/update/backup/node_merge_edge_test.go`, `internal/cli/update/backup/node_merge_twoway_comments_test.go`. The two backup files match the substring `deepMergeMapsTo`, a different function in `backup/node_merge.go`; `grep -rln "internal/merge\"" internal/cli/update/backup/ --include='*.go'` returns nothing, so they do not consume this change.
- `grep -rn "keepRetiredArrays\|permissionsKey" internal --include='*_test.go'`: no output.

**Package runs (exit 0 each).** `go test -count=1 -race ./internal/merge/` → `ok  	github.com/modu-ai/moai-adk/internal/merge	2.104s`. `go test -count=1 -race ./internal/cli/update/merge/` → `ok  	github.com/modu-ai/moai-adk/internal/cli/update/merge	1.434s`. Neither package imports the report package, so the U-16 working-tree state does not change these results.

**Gate-equivalent checks (lint, vet, gofmt, build).**

- `go build ./internal/cli/... ./internal/merge/...`: exit 0.
- `go vet ./internal/merge/ ./internal/cli/update/...`: exit 0. `go vet ./internal/cli/`: exit 0.
- `gofmt -l` on the seven touched Go files: no output.
- `golangci-lint run --timeout=10m ./internal/merge/ ./internal/cli/update/merge/ ./internal/cli/update/report/`: `0 issues.`, exit 0.
- `golangci-lint run --timeout=10m ./internal/cli/ ./internal/cli/update/`: `0 issues.`, exit 0. This run includes the U-16 working-tree files.

**Consumer run, the 60-function M5 set under the heavy-run slot.** `moai slot acquire --resource internal-cli-tests --max-duration 30m` exit 0; the run; `moai slot release --resource internal-cli-tests` exit 0. Command: `go test -count=1 -race -v -run '<the 60-name list, line 6 of .moai/reports/t1619/m5-consumer-tests.txt>' ./internal/cli/ > <scratchpad>/m6-consumer-run.txt 2>&1`. The file is machine-local scratch, so the deciding lines are carried here. Result: exit 1, `FAIL	github.com/modu-ai/moai-adk/internal/cli	110.248s`. Top-level counts: 56 `--- PASS`, 4 `--- FAIL`, so all 60 selected names ran. The settings-family tests in the set passed: `--- PASS: TestUpdateForce_UserCustomizationSurvivesTwoUpdates (6.43s)`, `--- PASS: TestUpdateForce_TemplateChangedKeyStillPropagates (4.40s)`, `--- PASS: TestUpdateLLMYAMLCommentsSurvive (0.50s)`, `--- PASS: TestUpdatePreservesUserModifiedPolicies (0.48s)`, `--- PASS: TestSettingsSnapshot_WriteSites (11.80s)`.

The four failures are not U-15. Each asserts that the default summary names a preserved path, which is the U-16 default change (see the next section):

```text
update_migrate_test.go:453: force-run summary did not list the preserved foreign skill:        (TestForceUpdatePreservesUserCommonAssets)
update_migrate_test.go:490: run 2 summary did not list the preserved foreign skill:            (TestNotDemonstratedPreservationHoldsAtNextLocalUpdate)
update_migrate_test.go:520: summary did not report the local-only section as preserved:        (TestPreservedSectionSurvivesRestoreUntouched)
update_migrate_test.go:1133: the preserved post-removal file never reached the summary:        (TestPostRemovalUserFileSurvivesNextUpdate)
```

#### U-16 GREEN attempt and blocker (AC-RC30-17), NOT committed

**What passes.** `go test -count=1 -race -run '^TestPreservedLinesCountOnly$' -v ./internal/cli/update/report/` → `--- PASS: TestPreservedLinesCountOnly (0.00s)`, `ok  	github.com/modu-ai/moai-adk/internal/cli/update/report	1.343s`, exit 0. The whole report package, `go test -count=1 -race -v ./internal/cli/update/report/`, exits 0 with `ok  	github.com/modu-ai/moai-adk/internal/cli/update/report	1.184s`, and `TestRenderReconciliation` passes. `TestUpdate_SummaryHonesty` (`internal/cli/update/reconcile_test.go`, its only consumer of the new call shape) passes: `ok  	github.com/modu-ai/moai-adk/internal/cli/update	1.349s`.

**What breaks.** The 60-function run above: four existing tests in `internal/cli/update_migrate_test.go` assert that the default update summary names a preserved path. The GREEN prints `  preserved: 3 file(s) (run with --verbose to list)` in its place.

**The conflict, read from the requirements.**

- REQ-RC30-015 (spec.md line 56, item U-16): without `--verbose`, "the update shall print per-category counts in place of the file list". AC-RC30-17 expects the count line by default.
- SPEC-UPDATE-MIGRATION-001 REQ-UPM-031/032: preserved files are listed, so the summary cannot claim a deletion-free run over what it kept. The rationale is written in `internal/cli/update/report/outcome.go` and in the `TestRenderReconciliation` comment. plan.md § A binds REQ-UPM-030 to -032, and the four failing tests encode that requirement.
- The U-16 row in spec.md (line 98) does not mention REQ-UPM-031.

**Decision needed (leader, then operator).** (A) Accept that REQ-RC30-015 supersedes the path listing of REQ-UPM-031 for preserved files only. Then the four expectations move to the count line and the `--verbose` list, in a separate commit that states the reason, and the SPEC gets an amendment note. Or (B) keep REQ-UPM-031 (preserved paths listed by default), and send AC-RC30-17 and REQ-RC30-015 back to manager-spec. The lane does not choose between them and does not change the four expectations (leader rule: never change an expectation to make a test pass).

**Left in the working tree, uncommitted.** Tracked: `internal/cli/update/report/outcome.go` (the `verbose` parameter and the count line), `internal/cli/update/report/outcome_test.go` and `internal/cli/update/reconcile_test.go` (the `true` argument), and `internal/cli/update_tux.go` (the `updateVerboseMode` argument and its comment). Untracked: `internal/cli/update/report/preserved_lines_test.go` (TestPreservedLinesCountOnly). The tracked diff is saved as `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/40f97120-d133-4bd7-a79c-a6466b3a087c/scratchpad/u16-tracked.patch` (4 files, 18 insertions, 11 deletions). Its RED evidence is committed in `30de29d36`.

#### Commit order actually used (supersedes the "Commit order" paragraph above)

- `51f957864`: the M5 consumer-run record (docs; round 2A task 1).
- `30de29d36`: the M6 root cause and RED record (docs only).
- `c9ab1903f`: the U-15 fix and its test (code: `internal/merge/strategies.go`, `internal/cli/update/merge/permissions_ask_preserved_test.go`).
- U-16: not committed in the blocked round (blocked, see above). Resolved by operator ruling `d-20261012T221654Z-9ced`: the summary surface landed as `024e740a6e` (RED) and `6ecfb4a1e0` (GREEN). The `--dry-run` preview surface is round 3 below (`0ddb588ce9` RED, `e5b11c04e` GREEN).
- This GREEN record commits last.

The U-15 code commit is separate from U-16 on purpose: the merge packages do not import the report package, so the U-15 commit stands alone, and the U-16 change stays out of the commit until the decision is made.

#### Gate scope (deviation, recorded)

The pre-commit gate (`/verify`, delegating to the moai skill's `gate --staged`) runs the full Go test suite under `--staged` (gate.md, Scoping: "Tests always run full suite"). The lane rules forbid a repository-wide `go test` locally, so the gate was not run as specified. The gate-equivalent checks above were run instead: lint, vet, gofmt, build, the change-scoped test packages, and the 60-function M5 consumer set under the heavy-run slot. The repository-wide test verdict is owned by CI on `origin/develop` and is PENDING.

#### Gaps (not observed in this round)

- G-M6-1: the repository-wide test verdict (the gate's full `--staged` suite) was not run locally. CI on `origin/develop` owns it, PENDING.
- G-M6-2: the 60-function run was made on the combined working tree (U-15 committed, U-16 uncommitted). U-15's contribution to the 56 PASS is not isolated by a separate run.
- G-M6-3: the full `internal/cli` suite was not run. Only the 60-function consumer set ran, plus lint and vet of the package.
- G-M6-4: AC-RC30-16's starting observation is recorded above (RED, exit 1). The cause is established from the code (`strategies.go:439-441`) and corroborated by the F-2 commit message. The F-2 diff was not read.

#### Residual risk

- R-M6-1: the U-15 fix applies to any top-level JSON or YAML object named `permissions` that the user left unchanged and the template changed. The merge engine does not receive the file path, so the scope is wider than `.claude/settings.json`. Retired permission arrays now also appear in the retained-key advisory (`permissions.<array>`), with the same semantics as other template-removed keys.
- R-M6-2: under decision (A) the four expectations change, and under (B) the default summary keeps the path lists and U-16 needs re-scoping.

### M6 round 3 — `--dry-run` preview counts preserved files by default (AC-RC30-17, U-16) (manager-develop, card t1619)

**Scope.** The U-16 summary surface had already landed (see the resolution note in the commit-order list above). This round applies the same rule to the `--dry-run` reconciliation preview (`previewReconciliation`, `internal/cli/update_dryrun_preview.go`), which still listed every preserved path.

**Commits.** `0ddb588ce9` is the RED (test only, prior round). `e5b11c04e44217e8cdfc5c8138025b0a4f6a1472` is the GREEN (implementation plus the five changed tests). This record is the commit after `e5b11c04e` and cannot name its own SHA.

**Decision.** (1) By default the preserved user-owned files are one line, `[dry-run] preserved: N file(s)`. Each path is listed only under `--verbose` (ruling `d-20261012T221654Z-9ced`). (2) Header option (c): the header line `N files: … N preserved …` is unchanged. It counts every user-owned file, migration candidates included. (3) The count line appends `; migration candidates are not counted here (listed above)` only when at least one migration candidate is listed above it. The round's decision text asked for the note on the default count line. It is narrowed to that case, because on the common path no candidate is listed, and the note would describe a candidate that is not on screen.

**Starting observation (RED at HEAD `0ddb588ce9`, re-observed in this run).** Command: `go test ./internal/cli/ -run '^TestPreviewPreservedLinesCountByDefault$' -count=1 -v`. Exit 1. Assertion lines, verbatim:

```text
    update_dryrun_preview_count_test.go:44: default preview lists the preserved path .claude/skills/moai-domain-db-docs/SKILL.md; the count is the default, got:
    update_dryrun_preview_count_test.go:44: default preview lists the preserved path .claude/skills/moai-domain-mobile/SKILL.md; the count is the default, got:
    update_dryrun_preview_count_test.go:48: default preview must carry the preserved count line, got:
    update_dryrun_preview_count_test.go:51: default preview must name the --verbose route to the list, got:
--- FAIL: TestPreviewPreservedLinesCountByDefault (0.00s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.998s
```

The echoed preview body shows the HEAD per-path format (`[dry-run] preserved: .claude/skills/… user-owned — left untouched`), so the binary was built from HEAD. Method: the GREEN file was saved as a patch, HEAD was restored with `git checkout`, the test ran, and the patch was re-applied. The re-apply was issued in the same batch as the test run, and the output format fixes which tree was compiled. The blob hash of the restored file (`1103579f09…`) matched the hash the patch recorded.

**Failures before the test edits (GREEN preview, old default-mode assertions).** Command: `go test ./internal/cli/ -run TestPreview -count=1 -v`. Exit 1. Five tests failed. Assertion lines, verbatim:

```text
    update_migrate_test.go:1258: the control (untracked local-only skill) must stay plain preserved:
--- FAIL: TestPreviewFlagsMigrationRemovalCandidates (0.00s)
    update_migrate_test.go:1297: the preview dropped the copy instead of naming it:
--- FAIL: TestPreviewOmitsRemovalCandidateWhileSwitchOff (0.00s)
    update_observability_test.go:319: preview must still classify (conservative no-record route), got:
--- FAIL: TestPreviewReconciliation_ReadOnlyOnCorruptManifest (0.00s)
    update_observability_test.go:350: local-mode preview lost the preserved skill listing:
--- FAIL: TestPreviewReconciliation_ModeScoped (0.00s)
    update_observability_test.go:254: preview must name the preserved local-only skill path, got:
--- FAIL: TestPreviewReconciliation (0.00s)
```

`TestPreviewPreservedLinesCountByDefault`, `TestPreviewCountEqualsRunCount`, and `TestPreviewReconciliation_EmptyProject` passed in that run.

**Changed tests.** Each per-path assertion that the count contract replaces by default moved into a serial `--verbose` twin, placed next to its original. A twin sets `updateVerboseMode` through the helper `withVerboseMode(t)` (in `update_dryrun_preview_count_test.go`, restored in `t.Cleanup`) and does not call `t.Parallel`. Go runs parallel top-level tests only after the sequential ones finish, so the package flag is not raced.

| Original test (default mode) | Default-mode expectation now | Verbose twin keeps the path check |
|---|---|---|
| `TestPreviewFlagsMigrationRemovalCandidates` | `[dry-run] preserved: 1 file(s)`, control path absent, count note present | `TestPreviewFlagsMigrationRemovalCandidates_VerboseListsControl` |
| `TestPreviewOmitsRemovalCandidateWhileSwitchOff` | `[dry-run] preserved: 1 file(s)`, copy path absent | `TestPreviewOmitsRemovalCandidateWhileSwitchOff_VerboseNamesCopy` |
| `TestPreviewReconciliation_ReadOnlyOnCorruptManifest` | `[dry-run] preserved: 1 file(s)`, `.claude/rules/moai/local.md` absent | `TestPreviewReconciliation_ReadOnlyOnCorruptManifest_VerboseListsLocalRule` |
| `TestPreviewReconciliation_ModeScoped` | `[dry-run] preserved: 1 file(s)`, `.claude/skills/moai-custom/SKILL.md` absent | `TestPreviewReconciliation_ModeScoped_VerboseListsSkill` |
| `TestPreviewReconciliation` | `[dry-run] preserved: 1 file(s)`, skill path absent | `TestPreviewReconciliation_VerboseListsPreservedSkill` |

Unchanged: `TestPreviewPreservedLinesCountByDefault` (the RED, from `0ddb588ce9`; it passes now), `TestPreviewCountEqualsRunCount` (it compares merge-review counts and never reads the preview text), and `TestPreviewReconciliation_EmptyProject`. Each default-mode test keeps its non-path assertions (read-only contract, migration candidate line, no-overlap check, refreshed count).

**Targeted runs after the edits (exact names, `-count=1`).** Each changed or new test passed on its own:

```text
--- PASS: TestPreviewPreservedLinesCountByDefault (0.00s)
--- PASS: TestPreviewFlagsMigrationRemovalCandidates (0.00s)
--- PASS: TestPreviewFlagsMigrationRemovalCandidates_VerboseListsControl (0.00s)
--- PASS: TestPreviewOmitsRemovalCandidateWhileSwitchOff (0.00s)
--- PASS: TestPreviewOmitsRemovalCandidateWhileSwitchOff_VerboseNamesCopy (0.00s)
--- PASS: TestPreviewReconciliation (0.00s)
--- PASS: TestPreviewReconciliation_VerboseListsPreservedSkill (0.00s)
--- PASS: TestPreviewReconciliation_ReadOnlyOnCorruptManifest (0.00s)
--- PASS: TestPreviewReconciliation_ReadOnlyOnCorruptManifest_VerboseListsLocalRule (0.00s)
--- PASS: TestPreviewReconciliation_ModeScoped (0.00s)
--- PASS: TestPreviewReconciliation_ModeScoped_VerboseListsSkill (0.00s)
```

The whole `TestPreview` family (13 tests, including the unchanged tests named above) ran as one command: `PASS` and `ok  	github.com/modu-ai/moai-adk/internal/cli	0.759s`, exit 0. Neighbouring tests that drive the real update or dry-run path all passed, exit 0: `TestUpdateDryRun_EmitsCleanReinstallPlan`, `TestUpdateDryRun_EmitsResidueCleanupPlan`, `TestUpdateDryRun_ZeroMutation` (both subtests), `TestUpdateDryRun_PreservesExistingOutput`, `TestForceUpdatePreservesUserCommonAssets`, `TestNotDemonstratedPreservationHoldsAtNextLocalUpdate`, `TestPreservedSectionSurvivesRestoreUntouched`, and `TestPostRemovalUserFileSurvivesNextUpdate`. Before commit B, `gofmt -l` on the four touched Go files produced no output.

**Mutant probes (observed red for the round-3 tests).** Each probe put a temporary implementation in place, ran `go test ./internal/cli/ -run 'TestPreview' -count=1 -v`, and then restored the committed `e5b11c04e` file. The restored blob is `cf49c5ca7e…` (`git hash-object`), and `git status --short` is clean afterwards.

- Probe A, HEAD implementation (`0ddb588ce9`) under the round-3 tests. Exit 1. Six default-mode tests fail on their count assertions: `TestPreviewPreservedLinesCountByDefault`, `TestPreviewFlagsMigrationRemovalCandidates`, `TestPreviewOmitsRemovalCandidateWhileSwitchOff`, `TestPreviewReconciliation_ReadOnlyOnCorruptManifest`, `TestPreviewReconciliation_ModeScoped`, and `TestPreviewReconciliation`. The five verbose twins pass, because HEAD already printed each path in every mode. The twins are therefore characterization tests of the verbose listing. They are not red-first tests of this change, and the red for them is probe B.
- Probe B, the committed implementation with the verbose listing disabled (`if false && updateVerboseMode` in place of `if updateVerboseMode`). Exit 1. All five twins fail on their path assertions: `update_migrate_test.go:1308`, `update_migrate_test.go:1398`, `update_observability_test.go:303`, `update_observability_test.go:395`, and `update_observability_test.go:457`. The default-mode tests pass, as expected.
- Restored state: the `TestPreview` family gives 13 `--- PASS` lines, `ok  	github.com/modu-ai/moai-adk/internal/cli	0.961s`, exit 0.

**Files changed (round 3, commit `e5b11c04e`).** `internal/cli/update_dryrun_preview.go` (count, candidate counter, note); `internal/cli/update_dryrun_preview_count_test.go` (`withVerboseMode` helper); `internal/cli/update_observability_test.go` (three default-mode assertions, three twins); `internal/cli/update_migrate_test.go` (two default-mode assertions, two twins). `internal/cli/update_clean_install.go` is untouched.

**Gaps (not observed in round 3).**

- G-R3-1: `/verify` was not run before commit `e5b11c04e`. `.claude/skills/moai/workflows/gate.md` documents `--fix`, `--staged`, `--file PATH`, and `--fresh`, and no package-scope option. Its `--file PATH` form is documented as checking the named files, with tests run "only matching test files if discoverable". `gate.md` does not say which Go tests that selects, so that scope was not observed, and the gate was not run: this lane does not run the `internal/cli` suite locally, and CI owns it. The first full verdict is the push-time CI run on `origin/develop`.
- G-R3-2: the repository-wide test verdict is PENDING (CI on `origin/develop` owns it).
- G-R3-3: the full `internal/cli` suite was not run, and `internal/cli` coverage was not measured.
- G-R3-4: for `e5b11c04e`, `git log -1 --format=%(trailers:key=Authored-By-Agent,valueonly)` returns empty. The `🗿 MoAI` marker is last, per the attribution mandate. The `Authored-By-Agent:` line sits in its own paragraph, and the ownership consumer's predicate matches it (body line 7). This is the known limitation of git's trailer parser for a marker-last layout, not a missing attribution.

**Residual risk.**

- R-R3-1: when every user-owned file is a migration candidate, the count line does not print (`preserved > 0` guard). The header still reads `N preserved` and the note cannot appear. Header option (c) leaves that case unexplained, and no test covers it.
- R-R3-2: each verbose twin repeats the fixture of its default-mode original. A fixture edit in one must be mirrored in the other.
- R-R3-3: the count assertions match exact text (`[dry-run] preserved: 1 file(s)` and the note). A wording change needs a test edit in the same commit.
- R-R3-4: the summary surface (`internal/cli/update/report/outcome.go`) and the preview surface now carry the same count rule through two code paths. A later change to one may not reach the other.

### CI fix — Linux CI run 38330633013, four failing tests (manager-develop, card t1619)

**Scope.** Job `Test (ubuntu-latest)` (job id 115079438840) of CI run 38330633013, head branch `WT-moai-update-repair`, head SHA `ccf9aa6ae96716e29e068f9b536d18f8d387307c`, concluded `failure` in the step "Run tests with coverage (fast — no race detector)". The run was still in progress when it was read, so only the job log was available.

**Commits.** Commit 1 `507844b67aad90777b4694e9039467afc0d606c6` carries the code and test-hermeticity fix. This record is commit 2 and cannot name its own SHA. The after-fix results below were measured on the tree of commit 1; the staged diff was committed unchanged.

**The failing lines, verbatim (CI job log of job 115079438840).**

```text
    register_test.go:191: recover site internal/cli/update_integrity_probe.go:runManagedSurfaceIntegrityProbe neither calls the capture entry point nor is allowlisted
    update_integrity_probe_test.go:184: REQ-UMF-002: a probe fault must never return an error: load manifest: manifest load: read .moai/manifest.json: is a directory
--- FAIL: TestProfileSyncResolution (0.00s)
    update_profile_sync_test.go:67: profile sync did not resolve the project's ledger profile: conversation_language = "en", want "ja" (work profile)
    update_single_banner_test.go:60: a run must print exactly one version line, got 0: []
    update_single_banner_test.go:63: the identity band must be kept (REQ-TUXIU-015)
--- FAIL: TestUpdateSingleBanner (0.00s)
```

**Cause and evidence, per test.**

1. `TestEveryRecoverSiteReportsOrIsAllowlisted`. Confirmed cause: the deferred recover in `runManagedSurfaceIntegrityProbe` (`internal/cli/update_integrity_probe.go`, introduced on this branch) neither calls the capture entry point nor is on the allowlist. Not Linux-specific: it fails identically on macOS.
   - Before. Command: `go test ./internal/bugreport/ -run '^TestEveryRecoverSiteReportsOrIsAllowlisted$' -count=1`. Output: `--- FAIL: TestEveryRecoverSiteReportsOrIsAllowlisted (1.83s)`, the register line above verbatim, `FAIL`. Exit 1.
   - Fix: the recover calls `bugreport.Capture(bugreport.KindPanic, nil, "", nil)`, the convention for recovered panics, and `internal/bugreport/register.go` lists the site. The allowlist is not extended.
   - After. `--- PASS: TestEveryRecoverSiteReportsOrIsAllowlisted (2.55s)`, `ok  	github.com/modu-ai/moai-adk/internal/bugreport	2.731s`. `TestEmitSiteRegister`: `--- PASS: TestEmitSiteRegister (0.00s)`.

2. `TestIntegrityProbe_FailOpen`. Confirmed cause: not the probe. The version-match predicate `versionMatchSkipsSync` (`internal/cli/update_template_sync_guard.go`, RC30 M5 design item 1) returned false when `manifest.Load` failed, so the run went into the template sync. The sync's own manifest load failed on the same directory (`update_template_sync.go:270-275`), and `runUpdate` returned that error (`update.go:577-580`) before the probe, which runs only on the skip path (`update.go:613`). Not Linux-specific: the macOS and CI traces match line for line.
   - Before. Command: `go test ./internal/cli/ -run '^TestIntegrityProbe_FailOpen$' -count=1 -v`. Output: the `update_integrity_probe_test.go:184` line above verbatim, `✓ Version check complete`, `✗ Manifest load failed: manifest load: read .moai/manifest.json: is a directory`, `--- FAIL: TestIntegrityProbe_FailOpen (2.24s)`. Exit 1.
   - Fix: on a load error the predicate returns `!errors.Is(err, manifest.ErrManifestCorrupt)`. Observed with a temporary diagnostic test (deleted before the commit): `missing: versionMatchSkipsSync=true corrupt-sidecar-exists=false`, `corrupt: versionMatchSkipsSync=false corrupt-sidecar-exists=true`, `directory: versionMatchSkipsSync=true corrupt-sidecar-exists=false`.
   - After. `--- PASS: TestIntegrityProbe_FailOpen (0.56s)`.
   - **Decision (narrows M5 item 1 above).** The sentence "A manifest that fails to load does not skip." now applies to corrupt manifests only. Reasons: (a) REQ-UMF-002 and AC-UMF-002 (release-blocking in SPEC-UPDATE-MIGRATION-FIX-001) require a version-matched run with a damaged representative path to exit 0 with one integrity row; (b) the sync cannot rebuild a path it cannot read, so the sentence as written turned the probe's warning into a hard failure on every version-matched run with such a manifest; (c) a corrupt manifest still syncs, because `Load` moves it aside and the sync rebuilds it, and that branch is unchanged.

3. `TestProfileSyncResolution`. Confirmed cause: a test-order leak, not Linux. `TestUpdateCmd_TemplatesOnlyAndBinaryMutuallyExclusive` (`internal/cli/update_mode_test.go`) set `templates-only` and `binary` on the package-level `updateCmd` and never restored them. The first validation in `runUpdate` (`update.go:179-181`) then returns "--binary and --templates-only are mutually exclusive" before the identity band (`update.go:202`), before the sync, and before the profile step. Every later run in the package returns at once, which is the 0.00s in the CI job. The test passes alone (`--- PASS: TestProfileSyncResolution (0.50s)`), which is why a single-name run cannot show the cause.
   - Before, leaker first. Command: `go test ./internal/cli/ -run '^(TestUpdateCmd_TemplatesOnlyAndBinaryMutuallyExclusive|TestProfileSyncResolution)$' -count=1 -v`. Output: `--- PASS: TestUpdateCmd_TemplatesOnlyAndBinaryMutuallyExclusive (0.00s)`, the `update_profile_sync_test.go:67` line above verbatim, `--- FAIL: TestProfileSyncResolution (0.00s)`. Exit 1.
   - Observed early return, with a temporary diagnostic test (deleted): `leaked binary=true templates-only=true`, `RunE err=--binary and --templates-only are mutually exclusive`, `RunE output=""`.
   - Fix: the leaking test snapshots and restores `templates-only`, `binary`, and `yes` through `t.Cleanup`, the pattern `runUpdateFixtureErr` already uses for the shared command.
   - After, in CI order. Command: `go test ./internal/cli/ -run '^(TestRunUpdate_CodexOnlyVersionMatch_PrintsNoIntegrityRow|TestUpdateCmd_TemplatesOnlyAndBinaryMutuallyExclusive|TestProfileSyncResolution)$' -count=1 -v`. The codex test is the earlier test that leaves `force=true` set in CI order. Output: `--- PASS: TestProfileSyncResolution (0.50s)`, `PASS`, `ok  	github.com/modu-ai/moai-adk/internal/cli	2.111s`.

4. `TestUpdateSingleBanner`. Same cause as item 3. Before, leaker first: `update_single_banner_test.go:60` and `:63` lines above verbatim, `--- FAIL: TestUpdateSingleBanner (0.00s)`. After, alone: `--- PASS: TestUpdateSingleBanner (0.51s)`. After, in CI order (same three-name chain with `TestUpdateSingleBanner`): `--- PASS: TestUpdateSingleBanner (0.49s)`, `ok  	github.com/modu-ai/moai-adk/internal/cli	4.362s`.

**Attribution scan (read with `git grep`, not run).** Every write to the shared update command's flags in `internal/cli` tests was listed. `update_mode_test.go:101-103` is the only writer of `binary` or `templates-only` without cleanup. `update_harness_test.go` (via `resetUpdateFlags`, then `force` and `yes`) and `update_integrity_probe_codex_test.go` (via `runCodexOnlyProjectThenUpdate`) leave `force` and `yes` set. The victims pass with that state, as the CI-order chain shows.

**Regression checks after the fix (tree of commit 1).** Each command was run alone, one name per invocation, `-count=1`.
- `TestSkipDecisionUsesTreeHash`: `--- PASS: TestSkipDecisionUsesTreeHash (0.04s)`, both subtests pass.
- `TestSkipSyncNoArchive`: `--- PASS: TestSkipSyncNoArchive (0.01s)`.
- `TestRunUpdate_VersionMatch_RunsIntegrityProbe`: `--- PASS: TestRunUpdate_VersionMatch_RunsIntegrityProbe (0.66s)`.
- `TestRunUpdate_UserCancelled_SkipsIntegrityProbe`: `--- PASS: TestRunUpdate_UserCancelled_SkipsIntegrityProbe (0.33s)`.
- `TestIntegrityProbeSet_EveryMemberObservableOnVersionMatchedPath`: `--- PASS` for both members.
- `TestIntegrityProbeEntry_DamageReasons`: `--- PASS` for all eight subtests.
- `TestReadProbeFile_HandleRecheckRefusesNonRegular`: `--- PASS (0.01s)`.
- `TestUpdateCmd_TemplatesOnlyAndBinaryMutuallyExclusive`: `--- PASS (0.00s)`.
- `go vet ./internal/bugreport/ ./internal/cli/`: no output, exit 0.
- `golangci-lint run ./internal/bugreport/ ./internal/cli/`: `0 issues.`
- `gofmt -l` on the four changed files: no output.
- `go test -count=1 -run '^$' ./internal/bugreport/ ./internal/cli/`: `ok  	github.com/modu-ai/moai-adk/internal/bugreport	0.519s [no tests to run]` and `ok  	github.com/modu-ai/moai-adk/internal/cli	0.956s [no tests to run]`.

**Files changed (commit 1).** `internal/cli/update_integrity_probe.go` (import; `KindPanic` capture in the recover). `internal/bugreport/register.go` (one `KindPanic` lead). `internal/cli/update_mode_test.go` (restore of the three flags the test sets). `internal/cli/update_template_sync_guard.go` (predicate on a manifest load error, and its comment).

**Gaps (not observed).**
- G-1: Linux was not run locally. The Linux evidence is the CI log: the test-2 trace matches the macOS trace. The Linux verdict for commit 1 is the next CI run on this branch, which the lane dispatches.
- G-2: The full `internal/cli` suite and `go test ./...` were not run, per the brief. The victims were checked in the CI-order chain of three named tests. The full-package order was not observed.
- G-3: `/verify` was not run. `.claude/skills/moai/workflows/gate.md` documents `--fix`, `--staged`, `--file PATH`, and `--fresh`, with no package scope. Its Go test step is `go test -race` over the module, and `gate.md` does not define the `--file` test scope for Go, so running it would run the module suite. The lane decides. The package-scoped checks are the vet, gofmt, lint, and targeted test lines above.
- G-4: The corrupt-manifest branch of the predicate has no targeted test on the update path. It was observed only through the temporary diagnostic.
- G-5: The trailer-parse query `git log -1 --format=%(trailers:key=Authored-By-Agent,valueonly)` was refused by the worktree guard as too complex, so the trailer was not parsed by git. The layout was checked by reading the body with `git log -1 --format=%B`. It matches commit `e5b11c04e`.
- G-6: Two- and three-name `go test -run` invocations were used only for the order-dependent reproduction. Every other run named one test.

**Residual risk.**
- R-1: For a corrupt manifest on a version-matched run, the predicate still does not skip, so the probe does not name the damage. REQ-UMF F3 ("every member observable on the version-matched path") is not met for that class. This is unchanged from RC30, and the sync rebuilds the file.
- R-2: Behaviour change. A version-matched project with an unreadable, non-corrupt manifest now exits 0 with one integrity row. Before this commit it exited 1 with `load manifest`.
- R-3: Tests 3 and 4 depend on the leaking test's cleanup. A future test that sets `binary` or `templates-only` on `updateCmd` without cleanup reopens them. The `force` and `yes` leaks in the harness tests are a hermeticity debt, left as they are.
- R-4: The probe's recover now emits a `KindPanic` signal. The walk checks that the call exists, not what it does at runtime. Capture is consent-gated and fail-open.
- R-5: The full CI job log and the local outputs are in machine-local scratch under `.moai/state/verify/t1619-ci-fix/`, which is not committed. The lines above are the carried evidence. The rest is a known loss.

### AC-RC30-24 (U-25) — deferral record (manager-develop, card t1619)

**Criterion.** AC-RC30-24 (acceptance.md line 237; REQ-RC30-003, owned by card t1668, cross-referenced here). Judging step: locate `.moai/logs/config-change-audit.log` in the operator profile, and attribute the untraced removal of 2026-10-09 11:26:48 (`.claude/skills/moai/**`, 46 files, and six core agents) to a command whose output is recorded in research.md. PASS only when that attribution is recorded.

**Disposition.** Deferred and excluded from the landing verdict, per ruling d-20261010T223350Z-7f6e (plan.md § B.7). This card does not close the criterion and claims no attribution. The E-18 observation stands: `ls -la .moai/logs/config-change-audit.log` exits 1 on this tree (M1 entry above). The research.md removal-attribution verdict still reads that the removing command is not attributed.

**Trigger (verbatim, acceptance.md line 237, Status column).**

> deferred: excluded from the landing verdict; the operator checks the log at the 3.2-6-2 rc.30 field run (mo.ai.kr). Regression-guard. REQ-RC30-003 is owned by card t1668; this row is its cross-reference.

The trigger is the operator's check of the log at the 3.2-6-2 rc.30 field run (mo.ai.kr). The operator-profile search named in the judging step was not performed by this card.

**Decision commits.** Both exist in this run (`git log -1 --format=%h%x20%s <sha>`):

- `bd190dd294 docs(SPEC-UPDATE-REPAIR-RC30-001): defer AC-RC30-24 U-25 closure to operator field run (card t1619)`. Edits plan.md § B.7 and the spec.md U-25 row (Next step marked operator-terminal).
- `277c9eddc9 docs(SPEC-UPDATE-REPAIR-RC30-001): align AC-RC30-24 stale wording with its deferral (card t1619)`. Edits the research.md removal-attribution verdict sentence and the spec.md U-25 status cell.

**Gap carried.** The removing command is not attributed, and the audit log is absent from this tree. The criterion stays unclosed until the operator field run records the attribution.

## §E.3 Run-phase Audit-Ready Signal

**Scope.** This record is the M7 measurement that finding N1 of the sync audit at HEAD `30659e8e1a` asks for (plan.md § B.8 step 2; acceptance.md Definition of Done, line 262). It covers M7 only. It is not a run-phase completion claim.

**Claim.** At HEAD `30659e8e1aba4332bb1d20a6271c9df09b3a6708`, M7 is not met. Of the 15 impacted packages, 8 meet 85% and pass their test run; 5 pass their test run but fall below 85%; 2 do not pass their test run: `internal/cli` (timed out, no valid figure) and `internal/hook` (one failing test; its 87.6% comes from a completed run).

**Evidence.** Command, run once while the `go-heavy` slot was held. Output was redirected to `.moai/state/verify/t1619-m7/go-test-cover.txt` (machine-local scratch, gitignored):

```text
go test -cover -count=1 ./internal/bugreport ./internal/cli ./internal/cli/update ./internal/cli/update/backup ./internal/cli/update/merge ./internal/cli/update/report ./internal/cli/wizard ./internal/config ./internal/hook ./internal/manifest ./internal/merge ./internal/settings/yamlpatch ./internal/template ./internal/userassets ./internal/web
```

Package result lines, verbatim:

```text
ok  	github.com/modu-ai/moai-adk/internal/bugreport	2.357s	coverage: 71.7% of statements
FAIL	github.com/modu-ai/moai-adk/internal/cli	601.501s
ok  	github.com/modu-ai/moai-adk/internal/cli/update	3.341s	coverage: 85.2% of statements
ok  	github.com/modu-ai/moai-adk/internal/cli/update/backup	2.766s	coverage: 85.3% of statements
ok  	github.com/modu-ai/moai-adk/internal/cli/update/merge	2.701s	coverage: 93.1% of statements
ok  	github.com/modu-ai/moai-adk/internal/cli/update/report	0.415s	coverage: 95.1% of statements
ok  	github.com/modu-ai/moai-adk/internal/cli/wizard	13.437s	coverage: 94.2% of statements
ok  	github.com/modu-ai/moai-adk/internal/config	70.345s	coverage: 84.8% of statements
FAIL	github.com/modu-ai/moai-adk/internal/hook	594.339s
ok  	github.com/modu-ai/moai-adk/internal/manifest	0.523s	coverage: 89.2% of statements
ok  	github.com/modu-ai/moai-adk/internal/merge	0.682s	coverage: 87.9% of statements
ok  	github.com/modu-ai/moai-adk/internal/settings/yamlpatch	0.490s	coverage: 83.2% of statements
ok  	github.com/modu-ai/moai-adk/internal/template	590.183s	coverage: 82.7% of statements
ok  	github.com/modu-ai/moai-adk/internal/userassets	17.731s	coverage: 86.4% of statements
ok  	github.com/modu-ai/moai-adk/internal/web	151.180s	coverage: 73.5% of statements
```

Two coverage lines carry no package name. Output line 387 sits inside the `internal/cli` block and is followed by the timeout panic. Output line 1240 directly precedes the `internal/hook` FAIL line.

```text
coverage: 21.9% of statements
coverage: 87.6% of statements
```

**Impacted package set.** `git diff --name-only 2aab5f797 HEAD -- '*.go'` lists 162 Go files in 15 directories (saved as `.moai/state/verify/t1619-m7/touched-go-files.txt`). No Go file is deleted in that range. The 15 directories are the package lines above.

**Threshold comparison.** The threshold is 85% per touched package: plan.md line 98 (Tested pillar), acceptance.md line 262 (Definition of Done), and the SSOT key `test_coverage_target: 85` (`.moai/config/sections/quality.yaml` line 13).

| Package | Test verdict | Coverage figure | Against 85% |
|---|---|---|---|
| internal/bugreport | ok | 71.7% | below by 13.3 points |
| internal/cli | FAIL, timed out | no valid figure (Gap G-M7-1) | no number |
| internal/cli/update | ok | 85.2% | meets |
| internal/cli/update/backup | ok | 85.3% | meets |
| internal/cli/update/merge | ok | 93.1% | meets |
| internal/cli/update/report | ok | 95.1% | meets |
| internal/cli/wizard | ok | 94.2% | meets |
| internal/config | ok | 84.8% | below by 0.2 points |
| internal/hook | FAIL, one test | 87.6% (completed run) | figure meets; test verdict FAIL (Gap G-M7-2) |
| internal/manifest | ok | 89.2% | meets |
| internal/merge | ok | 87.9% | meets |
| internal/settings/yamlpatch | ok | 83.2% | below by 1.8 points |
| internal/template | ok | 82.7% | below by 2.3 points |
| internal/userassets | ok | 86.4% | meets |
| internal/web | ok | 73.5% | below by 11.5 points |

**Heavy-run slot, verbatim.**

```text
slot go-heavy acquired by b6daac3a-0119-4eb1-a070-c751d00d25b2 until 2026-10-13T04:35:10Z
slot go-heavy released (was b6daac3a-0119-4eb1-a070-c751d00d25b2)
```

No refusal was returned by either call.

**gofmt and go vet, verbatim.**

- `gofmt -l` over the touched Go files (the diff list, passed as arguments): no output. Cross-check, `gofmt -l internal/bugreport internal/cli internal/config internal/hook internal/manifest internal/merge internal/settings/yamlpatch internal/template internal/userassets internal/web` (every Go file in the 15 directories): no output. Positive control, `gofmt -l .moai/state/verify/t1619-m7/gofmt-control.go`, printed `.moai/state/verify/t1619-m7/gofmt-control.go`.
- `go vet ./internal/bugreport ./internal/cli ./internal/cli/update ./internal/cli/update/backup ./internal/cli/update/merge ./internal/cli/update/report ./internal/cli/wizard ./internal/config ./internal/hook ./internal/manifest ./internal/merge ./internal/settings/yamlpatch ./internal/template ./internal/userassets ./internal/web`: no output and no error (`.moai/state/verify/t1619-m7/go-vet.txt` is empty). Positive control, `go vet .moai/state/verify/t1619-m7/vet-control.go`, printed `.moai/state/verify/t1619-m7/vet-control.go:6:14: fmt.Printf format %d has arg "not-an-int" of wrong type string` and failed.

**Baseline-attribution.** Tree: HEAD `30659e8e1aba4332bb1d20a6271c9df09b3a6708` (`git rev-parse HEAD`). `git status --short` was empty before the run. Toolchain: `go version go1.26.8 darwin/arm64`. One run, `-count=1`, default per-binary timeout (`-test.timeout=10m0s`, visible in the process list). The record commit changes only this progress file (staged by explicit path), so the Go tree this record measures is the Go tree at the record commit.

**Gaps (not observed in this record).**

- G-M7-1 `internal/cli`: no valid figure, and the test verdict is FAIL. The test binary hit its 10-minute limit (`panic: test timed out after 10m0s`; running test `TestTreeScopeSkip_WTSessionsStillReviewed`), and `FAIL	github.com/modu-ai/moai-adk/internal/cli	601.501s` carries no figure. The `coverage: 21.9% of statements` line (output line 387) comes from the killed binary, so it reflects an incomplete run and is not a measurement of the package. Four top-level tests failed before the timeout: `TestCodexGoalUnmeasuredCapBoundsContinuation` (sub-test `a_measured_unmet_evaluation_resets_the_streak_(same_tree_key)`), `TestCountCodexAgentTOMLs_BadPatternRoot`, `TestCountCodexAgentTOMLs`, and `TestCodexReviewGatePrimaryCheckoutSkip`. Causes were not investigated.
- G-M7-2 `internal/hook`: the test verdict is FAIL. `--- FAIL: TestSessionStart_DeferredScanJoinsWithinBound`, sub-test `slow_scan_drops_advisory`: `Handle blocked 773.396ms; expected return near the 250ms bound`. The 87.6% figure above is from a completed run, but the package does not pass its test run, so M7 is not met for it. Cause not investigated.
- G-M7-3 Scope of "touched package". Plan.md line 98 names `internal/backup` and `internal/tui`. `internal/backup` is not a package path in this tree; the touched backup package is `internal/cli/update/backup` (85.3%, measured above). `internal/tui` has no Go change: `git diff --name-only 2aab5f797 HEAD -- internal/tui` lists four locale catalogs (`internal/tui/messages/en.yaml`, `ja.yaml`, `ko.yaml`, `zh.yaml`). It is outside the impacted Go set and was not measured in this run. Of the five packages below 85%, `internal/config` is named in plan.md line 98; `internal/bugreport`, `internal/settings/yamlpatch`, `internal/template`, and `internal/web` are not. Which set governs the DoD is a decision for the leader.
- G-M7-4 The run-level signal fields (`run_status`, `ac_pass_count`, and the other fields of the § E.3 signal schema) are not populated. This pass measured M7 only.
- G-M7-5 The coverage step of CI run `38334739075` (headSha `0d322db302`, cited in § E.4) was not read. Its log may carry per-package figures for G-M7-1 and G-M7-2. That SHA is not this record's baseline.

**Residual risk.**

- R-M7-1 `internal/template` finished at 590.183s against the 600s per-binary limit. A slower run would time out and lose its figure.
- R-M7-2 The host was shared during the run. `pgrep` showed another session's `vitest` process (a different repository) active while the run was in progress. The timing-bound failure in G-M7-2 and the timeout in G-M7-1 may depend on load; that is not established here.
- R-M7-3 `TestCountCodexAgentTOMLs` reported `count = 6`, and the operator's `~/.codex/agents` holds six `.toml` files (the § M2 bundle F entry, `~/.codex/agents`). Whether the test reads the real home was not checked; the match is a hypothesis only.

### E.3 top-up measurement

**Scope.** Top-up of the M7 measurement above, under the coordinator's refinement of the coverage criterion. Package-wide 85% applies to `internal/config` only. Every other changed package is judged by changed-line coverage: each added Go line of the card is checked against the coverage profile of its package. Package-wide percentages are reference only. This subsection is not a run-phase completion claim; the factory stage stays `sync`.

**Method.** Changed lines are the added lines of `git diff -U0 2aab5f797b75983e132af451da68f69e3426557b HEAD -- '*.go'`, non-test files only (test lines are not coverage targets). A changed line is **executed (E)** when an instrumented coverage block covering it has count > 0 in the package's profile. A line is **non-executable (X)** when no instrumented block covers it: declarations, struct fields, signatures, comments, braces, case labels, package and import lines, and data literals. A line in a `*_windows.go` file is **Windows-only (P)**: it is not compiled on the darwin/arm64 host and is not measured here. A line that is executable and covered by no profile is **unexecuted (U)**. The classifier is an awk program over the diff and the profile; its output is kept in the machine-local scratch directory and is not a citation target.

**Package list (the 15 directories of the diff, verbatim).** DoD, package-wide 85%: `internal/config`. Changed-line criterion: `internal/bugreport`, `internal/cli`, `internal/cli/update`, `internal/cli/update/backup`, `internal/cli/update/merge`, `internal/cli/update/report`, `internal/cli/wizard`, `internal/hook`, `internal/manifest`, `internal/merge`, `internal/settings/yamlpatch`, `internal/template`, `internal/userassets`, `internal/web`. `internal/cli/update` and `internal/cli/update/merge` carry test-file changes only (N = 0).

**Config (DoD), package-wide, verbatim.**

```text
before (HEAD 0f9ac47081, go test -count=1 -coverprofile ./internal/config/):
ok  	github.com/modu-ai/moai-adk/internal/config	3.330s	coverage: 84.8% of statements   (2321 of 2736 statements)
after (commit d4ba3234e tree, same command):
ok  	github.com/modu-ai/moai-adk/internal/config	3.682s	coverage: 85.4% of statements   (2337 of 2736 statements)
```

The after figure is from a final re-run at the same tree (`ok  	github.com/modu-ai/moai-adk/internal/config	3.682s	coverage: 85.4% of statements`). The added test file is `internal/config/llm_save_branches_test.go`; no production code changed.

**Changed-line table.** N = changed non-test lines; E, U, X, P as defined above. Package-wide figures in the last column are reference only.

| Package | N | E | U | X | P | Status | Package-wide (reference) |
|---|---:|---:|---:|---:|---:|---|---|
| internal/bugreport | 1 | 0 | 0 | 1 | 0 | closed: the one line is a data literal in a package-level registry | 71.7% |
| internal/cli (narrowed union, see below) | 1637 | 479 | 110 | 1021 | 27 | open: 110 U | not measured (full package not run) |
| internal/cli/update | 0 | 0 | 0 | 0 | 0 | no production change (test file only) | 85.2% |
| internal/cli/update/backup | 345 | 124 | 21 | 200 | 0 | open: 21 U | 85.3% |
| internal/cli/update/merge | 0 | 0 | 0 | 0 | 0 | no production change (test files only) | 93.1% |
| internal/cli/update/report | 12 | 5 | 0 | 7 | 0 | closed | 95.1% |
| internal/cli/wizard | 34 | 11 | 0 | 23 | 0 | closed | 94.2% |
| internal/config (DoD) | 321 | 132 | 3 | 186 | 0 | 3 U, each with a reason (below) | 85.4% (criterion met) |
| internal/hook | 597 | 207 | 25 | 337 | 28 | open: 25 U; 28 P | 87.6% |
| internal/manifest | 31 | 9 | 0 | 22 | 0 | closed | 89.2% |
| internal/merge | 40 | 20 | 0 | 20 | 0 | closed by 152bbcad8 (was U = 19) | 91.1% (was 87.9%) |
| internal/settings/yamlpatch | 133 | 66 | 0 | 67 | 0 | closed by 6cdd488e4 (was U = 8) | 87.4% (was 83.2%) |
| internal/template | 8 | 1 | 0 | 7 | 0 | closed | 82.7% |
| internal/userassets | 859 | 262 | 30 | 455 | 112 | open: 30 U; 112 P | 86.4% |
| internal/web | 16 | 8 | 0 | 8 | 0 | closed | 73.5% |
| **Total** | **4034** | **1324** | **189** | **2354** | **167** | 186 open U, 3 reasoned U | |

Per-file unexecuted counts of the open packages: `internal/cli` 110 = init_resume.go 31, doctor_lock_markers.go 16, codex_skills_disable.go 14, doctor_user_install.go 6, update_integrity_probe.go 6, update_disk_backup.go 5, doctor_harness.go 4, update_template_sync_guard.go 4, doctor.go 3, update_migrate.go 3, update_template_sync.go 3, glm.go 2, init_resume_read_unix.go 2, init.go 2, migrate_project_assets.go 2, update_clean_install.go 2, update_restore.go 2, bundle.go 1, init_update_notice.go 1, user_asset_phase.go 1. `internal/cli/update/backup` 21 = backup.go 11, run_marker.go 9, restore.go 1. `internal/hook` 25 = user_root_zone.go 23, zone_user_read_unix.go 2. `internal/userassets` 30 = install.go 20, remove.go 5, lock_guard_unix.go 3, journal.go 1, lock.go 1. Windows-only: `internal/cli/init_resume_read_windows.go` 27, `internal/hook/zone_user_read_windows.go` 28, `internal/userassets/lock_guard_windows.go` 112.

The packages closed in this run were measured at the tree where they stood: the test files added in d4ba3234e, 152bbcad8, and 6cdd488e4 are in packages other than the open ones, so the open figures hold at HEAD 0f9ac47081 and at the commit tree.

**Config: the three unexecuted changed lines, with reasons.** All three are error returns that the typed rendering cannot produce:

- `llm_save.go` line 56 (`return nil, fmt.Errorf("llm baseline: %w", err)` in `LLMSectionEdits`) propagates `llmMappingOf(llmBaselineOf(*persisted))`. `llmBaselineOf` returns nil only on the `yaml.Marshal` failure branch (line 69), and `llmMappingOf` finds the llm mapping in every document `llmBaselineOf` renders.
- `llm_save.go` line 69 (`return nil` in `llmBaselineOf`) is the `yaml.Marshal` failure branch. No input reaches it. The type tree of `LLMConfig` holds only string, bool, int, struct, and string-keyed map fields (`internal/config/types.go` lines 274-340 and 394-399; `ModelEffort` at `internal/config/audit_models.go` lines 65-68). yaml.v3 v3.0.1 encodes an untagged string with invalid UTF-8 as `!!binary` instead of failing (observed, see Evidence). The encoder's refusal sites (`encode.go` lines 331, 334, 546, 549, 575) need a tagged scalar or an unknown node kind, and the typed rendering produces neither.
- `llm_save.go` line 81 (`return nil, fmt.Errorf("llm section: %w", err)` in `llmSectionEditsFrom`) propagates `llmMappingOf(llmBaselineOf(cur))`, which fails for the same reason as line 56.

**Hook, four-way triage (`TestSessionStart_DeferredScanJoinsWithinBound`, file `internal/hook/session_start_parallel_test.go`, not changed by the card).**

```text
(1) package run, this tree, under the go-heavy slot:
ok  	github.com/modu-ai/moai-adk/internal/hook	250.088s	coverage: 87.6% of statements
    (no FAIL line)
(2) this tree, alone: go test ./internal/hook/ -run '^TestSessionStart_DeferredScanJoinsWithinBound$' -count=1 -v
--- PASS: TestSessionStart_DeferredScanJoinsWithinBound (0.58s)
    --- PASS: TestSessionStart_DeferredScanJoinsWithinBound/fast_scan_lands_advisory (0.26s)
    --- PASS: TestSessionStart_DeferredScanJoinsWithinBound/slow_scan_drops_advisory (0.32s)
ok  	github.com/modu-ai/moai-adk/internal/hook	1.336s
(3) main worktree (commit 79d0a0a47d), alone: go -C /Users/goos/MoAI/moai-adk-go/.claude/worktrees/main test ./internal/hook/ -run '^TestSessionStart_DeferredScanJoinsWithinBound$' -count=1 -v
    (not refused; the go invocation ran)
--- PASS: TestSessionStart_DeferredScanJoinsWithinBound (0.52s)
    --- PASS: TestSessionStart_DeferredScanJoinsWithinBound/fast_scan_lands_advisory (0.19s)
    --- PASS: TestSessionStart_DeferredScanJoinsWithinBound/slow_scan_drops_advisory (0.33s)
ok  	github.com/modu-ai/moai-adk/internal/hook	1.119s
(4) failure recorded in the M7 measurement (HEAD 30659e8e1a, under host load, R-M7-2), not reproduced in this run:
--- FAIL: TestSessionStart_DeferredScanJoinsWithinBound
    --- FAIL: TestSessionStart_DeferredScanJoinsWithinBound/slow_scan_drops_advisory
        Handle blocked 773.396ms; expected return near the 250ms bound
```

Decision under the leader's rule: the test passes alone in this tree, so it is unrelated to the card. It is recorded as a Gap and not fixed. No hook code changed.

**Narrowed `internal/cli` coverage (each line is narrowed; none is a package figure).** Each invocation was `go test ./internal/cli/ -run '<one prefix>' -count=1 -v -coverprofile=...`. The changed-line union above is the union of these families.

```text
TestInit               ok  	github.com/modu-ai/moai-adk/internal/cli	185.287s	coverage: 7.1% of statements   (83 top-level PASS)
TestDoctor             ok  	github.com/modu-ai/moai-adk/internal/cli	130.654s	coverage: 10.5% of statements   (94 top-level PASS)
TestUpdate             ok  	github.com/modu-ai/moai-adk/internal/cli	91.026s	coverage: 11.3% of statements   (87 top-level PASS)
TestMigrat             ok  	github.com/modu-ai/moai-adk/internal/cli	8.187s	coverage: 8.3% of statements   (55 top-level PASS)
TestBundle             ok  	github.com/modu-ai/moai-adk/internal/cli	8.106s	coverage: 5.0% of statements   (7 top-level PASS)
TestInitResume         ok  	github.com/modu-ai/moai-adk/internal/cli	37.744s	coverage: 6.5% of statements   (29 top-level PASS)
TestPreview            ok  	github.com/modu-ai/moai-adk/internal/cli	1.195s	coverage: 5.2% of statements   (13 top-level PASS)
TestIntegrityProbe     ok  	github.com/modu-ai/moai-adk/internal/cli	1.344s	coverage: 7.7% of statements   (4 top-level PASS)
TestSaveLLM            ok  	github.com/modu-ai/moai-adk/internal/cli	0.934s	coverage: 4.8% of statements   (13 top-level PASS)
TestRestoreSettings    ok  	github.com/modu-ai/moai-adk/internal/cli	0.827s	coverage: 4.9% of statements   (2 top-level PASS)
TestCleanReinstall     ok  	github.com/modu-ai/moai-adk/internal/cli	8.303s	coverage: 8.8% of statements   (23 top-level PASS)
TestRecordless         ok  	github.com/modu-ai/moai-adk/internal/cli	0.920s	coverage: 6.4% of statements   (2 top-level PASS)
TestDeployedWiring     ok  	github.com/modu-ai/moai-adk/internal/cli	0.754s	coverage: 4.7% of statements   (1 top-level PASS)
TestLLMYaml            ok  	github.com/modu-ai/moai-adk/internal/cli	0.749s	coverage: 4.8% of statements   (1 top-level PASS)
TestProfileSync        ok  	github.com/modu-ai/moai-adk/internal/cli	3.951s	coverage: 7.4% of statements   (1 top-level PASS)
TestSkipSync           ok  	github.com/modu-ai/moai-adk/internal/cli	0.904s	coverage: 4.8% of statements   (1 top-level PASS)
TestCodexInstall       ok  	github.com/modu-ai/moai-adk/internal/cli	1.027s	coverage: 4.7% of statements   (1 top-level PASS)
TestGateYAML           ok  	github.com/modu-ai/moai-adk/internal/cli	1.738s	coverage: 6.2% of statements   (1 top-level PASS)
TestTemplateSync       ok  	github.com/modu-ai/moai-adk/internal/cli	1.477s	coverage: 7.4% of statements   (5 top-level PASS)
TestSkipDecision       ok  	github.com/modu-ai/moai-adk/internal/cli	3.295s	coverage: 4.8% of statements   (1 top-level PASS)
TestLegacyLLM          ok  	github.com/modu-ai/moai-adk/internal/cli	0.997s	coverage: 4.7% of statements   (1 top-level PASS)
TestKeptOptional       ok  	github.com/modu-ai/moai-adk/internal/cli	2.797s	coverage: 5.1% of statements   (1 top-level PASS)
TestNonCatalog         ok  	github.com/modu-ai/moai-adk/internal/cli	1.597s	coverage: 5.1% of statements   (1 top-level PASS)
TestUpdateSingle       ok  	github.com/modu-ai/moai-adk/internal/cli	1.287s	coverage: 7.4% of statements   (1 top-level PASS)
TestStampSkew          ok  	github.com/modu-ai/moai-adk/internal/cli	0.805s	coverage: 4.7% of statements   (1 top-level PASS)
TestPreviewCount       ok  	github.com/modu-ai/moai-adk/internal/cli	0.771s	coverage: 5.0% of statements   (1 top-level PASS)
TestInitQuestions      ok  	github.com/modu-ai/moai-adk/internal/cli	0.833s	coverage: 4.6% of statements [no tests to run]   (EXCLUDED: zero tests swept)
exact TestInitQuestionsWithAgentWiring_PrefillsRecordedDefault in ./internal/cli/   ok  	github.com/modu-ai/moai-adk/internal/cli	1.155s	coverage: 4.6% of statements [no tests to run]   (EXCLUDED: zero tests swept; the test is in internal/cli/wizard)
```

**Commits (one concern each).** `d4ba3234e5ed4eba99124d70683f0106f6d310f5` test(config) (`internal/config/llm_save_branches_test.go`, 183 insertions); `152bbcad87d63ee208770e24d9a9142b3a787fd0` test(merge) (`internal/merge/permissions_retired_arrays_test.go`, 67 insertions); `6cdd488e4068b14dda9bc2e4eed9b9aefb33462f` test(yamlpatch) (`internal/settings/yamlpatch/branches_test.go`, 161 insertions). This subsection is the commit after them; its own SHA is not cited here (D3 backfill).

**Evidence.** Commands and outputs are in the lines above. Three corrections were observed during the run and are kept rather than rewritten. (a) The first yamlpatch invalid-UTF-8 test asserted a refusal of an untagged value; it failed with `PatchFile err = <nil>, want an encode refusal`, because yaml.v3 writes that value as `!!binary`. The test now uses the quoted form, which pins the `!!str` tag, and passes. (b) Two narrowed cli runs named tests that are not in `internal/cli` and swept zero tests (both print `[no tests to run]`). They are excluded above. (c) The first config commit message says lines 56, 69, and 81 are returns of `yaml.Marshal`; in fact 56 and 81 propagate `llmMappingOf` errors, whose root is the line 69 branch. This record is the authority for that wording.

**Baseline-attribution.** Every figure is attributed to the command above run in this session against the tree named there. The coverage tests added in this run are tests written after the code they cover (closing coverage on existing code); the RED state for each is the zero count of its line in the baseline profile, observed before the test was added.

**Gaps (not observed, or not closed).**

- The per-line rule is not met for 186 executable changed lines: `internal/hook` 25, `internal/userassets` 30, `internal/cli` 110, `internal/cli/update/backup` 21. No test and no reason was written for them in this run. Each is a real statement that the families run here did not reach, so none qualifies for a per-line reason. This is the open item for the leader's decision.
- 167 Windows-only changed lines (`init_resume_read_windows.go` 27, `zone_user_read_windows.go` 28, `lock_guard_windows.go` 112) are not measurable on darwin/arm64. A `GOOS=windows` run is required to measure them.
- The full `internal/cli` package was not run, per the instruction. Its figure is the narrowed union only (26 families: TestInit, TestDoctor, TestUpdate, TestMigrat, TestBundle, TestInitResume, TestPreview, TestIntegrityProbe, TestSaveLLM, TestRestoreSettings, TestCleanReinstall, TestRecordless, TestDeployedWiring, TestLLMYaml, TestProfileSync, TestSkipSync, TestCodexInstall, TestGateYAML, TestTemplateSync, TestSkipDecision, TestLegacyLLM, TestKeptOptional, TestNonCatalog, TestUpdateSingle, TestStampSkew, TestPreviewCount).
- The M7 hook failure was not reproduced in this run, and its cause was not investigated. The package passed in this run.
- `internal/cli/wizard` was measured as a full package (no `-run` filter). Its changed lines are closed on that basis.

**Residual risk.**

- The changed-line method marks a line executed when a covering block ran. A multi-line statement shares one block range, so a continuation line can be counted executed when only a sibling line ran. This can overstate E.
- The narrowed cli union can only reduce U when more families are added, so 110 is an upper bound on cli unexecuted lines under these families, not a claim about the package.
- An untagged value with invalid UTF-8 written through `PatchFile` is stored as `!!binary`, not refused. This card did not change that behavior, and the new tests do not assert it. It is noted because the llm save path writes through the same seam.
- The three config reasons depend on yaml.v3 v3.0.1 (`go.mod` line 33) not failing on the `LLMConfig` type tree. A field of a chan, func, or complex type added later would make line 69 reachable and invalidate its reason.

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-10-13T02:46:08Z
sync_commit_sha: pending-backfill-sync   # a commit cannot name its own SHA; backfilled in a following commit (spec-frontmatter-schema.md § SHA placeholder backfill exemption)
sync_status: audit-ready   # sync-audit verdict at HEAD not observed (see Gaps); card-review.md absent (see Gaps)
b12_self_test_a: pass   # grep -c 'SPEC-UPDATE-REPAIR-RC30-001' CHANGELOG.md → 0 before the entry was appended
b12_self_test_b: pass   # acceptance.md counter: stdout 22, stderr live=22 excluded=6 ambiguous=0; the entry carries no AC count (acceptance.md line 5 declares 22)
b12_self_test_c: pass   # ls .moai/specs/SPEC-UPDATE-REPAIR-RC30-001/spec.md (the only path the entry names)
changelog_entry_position: "CHANGELOG.md:28 — [Unreleased] → Changed, after the SPEC-UPDATE-MIGRATION-FIX-001 entry"
frontmatter_status_transitions:
  spec_md: "in-progress → implemented → completed (this sync commit, manager-docs); updated: 2026-10-11 → 2026-10-13"
  plan_md: unchanged body; updated not refreshed (outside this commit's scope; no status field, spec-frontmatter-schema.md § Artifact Statelessness)
  acceptance_md: unchanged body; updated not refreshed (outside this commit's scope; no status field)
  progress_md: "§E.4 updated in this sync commit; §E.1 to §E.3 and §G unchanged"
```

Scope: card t1619 lands on the current SPEC scope only (leader scope decision, 2026-10-13). The subsection at the end of this section records the card-body items outside the SPEC and where each one is handed off.

t1591 closes together with this card's tip when the tip lands (leader instruction, 2026-10-13).

Run-phase commits covered (`git log --oneline -6`, read at HEAD `0d322db302` before this sync commit):

```text
0d322db302 docs(SPEC-UPDATE-REPAIR-RC30-001): record E.2 for Linux CI fix (card t1619)
507844b67a fix(update): make card t1619 integrity, profile, and banner tests pass on Linux CI (card t1619)
ccf9aa6ae9 docs(SPEC-UPDATE-REPAIR-RC30-001): record twin mutant probes in E.2 (card t1619)
5c523fd804 docs(SPEC-UPDATE-REPAIR-RC30-001): correct E.2 gate-scope gap to what gate.md documents (card t1619)
4c78154f7f docs(SPEC-UPDATE-REPAIR-RC30-001): record E.2 for dry-run preview count-by-default (card t1619)
e5b11c04e4 fix(update): dry-run preview counts preserved files by default (card t1619)
```

CI evidence. Command: `gh run view 38334739075 --json conclusion,headSha,status,name,url,headBranch,event`. Observed output (2026-10-13):

```text
{"conclusion":"success","event":"workflow_dispatch","headBranch":"WT-moai-update-repair","headSha":"0d322db302a4074c01a556b79cfa7844f15b1a73","name":"CI","status":"completed","url":"https://github.com/modu-ai/moai-adk/actions/runs/38334739075"}
```

CHANGELOG.md: not edited by this commit. `grep -n -i changelog` over plan.md, acceptance.md, and spec.md printed no lines, so the SPEC requires no entry. The entry at CHANGELOG.md:28 comes from the earlier sync commit `ab0adb3830`, the last commit that touched CHANGELOG.md (see the b12 lines above).

**Gaps (not observed in this record).**
- Sync-auditor verdict at HEAD `0d322db302`: not observed. The sync audit on disk (`.moai/reports/t1619/sync-audit.md`, `verdict: FAIL`, `audited_sha: ab0adb38308dc7a008b5d73af222bb4e0847f377`) predates HEAD. Of its 17 findings (F1 to F17: 7 High, 5 Medium, 5 Low), this record re-checked only F10: the decision rows it named (`decision-index.md` lines 39, 75, 88, 115) now carry operator verdicts. F6 still holds at HEAD: §E.3 reads `_pending run-phase (manager-develop)_`.
- `card-review.md`: absent. `ls .moai/reports/t1619` lists no `card-review.md`. Not yet run.
- Coverage: not measured. This is a docs-only commit with no test run.

**Residual risk.**
- The SPEC is recorded `completed` before a sync-audit verdict at HEAD exists.
- `sync_commit_sha` holds the placeholder `pending-backfill-sync`. The real SHA lands in a follow-up commit.
- The CHANGELOG.md:28 entry is not required by the SPEC. This commit leaves it in place.

### 카드 본문 중 SPEC 밖 항목 → 후속 이관

| 카드 본문 항목 | SPEC 착지 조건 | 후속 이관 | 상태 |
|---|---|---|---|
| 출력 TUI/TUX 전면 개편 (a)–(g) | 밖 (U-16·U-17 일부만 SPEC 안) | t1668 · t1669 · t1681 (겹치는 부분, 리더 지정) | 겹치지 않는 나머지는 아래 미배정 |
| update 후 스킬 트리 소실 (git D 55건, ~/.claude/skills 이전 36개, 나머지 소실) | 밖 | t1668 (update 제거 가족, 리더 지정) | 리더 확인 요청 |
| 에이전트 명부 두 루트 분리 (6개 ~/.claude/agents, 6개 프로젝트 측 보존) | 밖 | t1668 | 리더 확인 요청 |
| 프로필 설정 폴더에서 스킬 가시성 (가설) | 밖 | 미배정 | 검증 필요 |
| 빈 껍데기 스킬 폴더 3개, 옛 이름 moai-kanban-foreman, 소문자 파일명 | 밖 | 미배정 | — |
| 백업 두 루트, "Cleaned up 1 old backup"의 유일본 삭제 여부 | 밖 | 미배정 | 미확인 |
| 출력 문제: preserved 369줄 무페이징 (U-16으로 SPEC 안 착지), 버전 배너 2회 (U-17로 SPEC 안 착지), 분류표 잘림·모순 문장 | 일부 안 | 분류표 잘림·모순 문장은 미배정 | 리더 확인 요청 |
| config restore가 .moai/config/sections/*.yaml만 복원 (evaluator-profiles/, astgrep-rules/ 미복원) | 밖 (복원 범위 확인 필요) | 미배정 | 확인 필요 |
| 설치 바이너리와 배포 규칙의 버전 스큐가 조용함 | 밖 | 미배정 | — |
| 요구: 자산 소실·복구 불가 경로 0, 성공 메시지 사실 일치, 이동·제외·복구 명령 출력, 프로필 설정 폴더 사용자 동작 | SPEC 항목과 대응 확인 필요 | 본 카드 SPEC (항목별 대응은 관문에서 확인) | — |
| 템플릿 범용 제약 (16개 언어, Claude Code·Codex 두 하네스) | 공통 제약 | 미배정 (전 카드 공통) | 리더 확인 요청 |
| 겹침: t1591 (SPEC-USERASSET-DEPLOY-GUARD-001), t1578 (SPEC-UPDATE-MIGRATION-FIX-001), t1527·t1594·t1612 (큐) | 관계 | t1591은 이 카드 팁과 함께 닫힘 (리더 지시) | t1578·t1527·t1594·t1612는 리더 확인 요청 |

## §G Absorption Record

Absorption is done for revision 0.1.4 (spec.md § G.2). Both tips were ancestors of HEAD `6f95917a1` when this record was written.

- WT-user-asset-bundle (t1591): tip `a372a984c`, merged as `ff2bc876d` ("Merge branch 'WT-user-asset-bundle' into WT-moai-update-repair (card t1619)"). Earlier merges of the same branch: `f5e039f2b` and `cbb6c7940`, both ancestors of HEAD.
- WT-update-migration-fixes (t1578): tip `e73a7cbf5`, merged as `6f95917a1` ("Merge branch 'WT-update-migration-fixes' into WT-moai-update-repair (card t1619)"). Earlier merge: `763720170` ("Merge branch 'WT-update-migration-fixes' into WT-moai-update-repair"), an ancestor of HEAD.
- Ranges are measured from the merge-base with the absorbed ref (`git merge-base HEAD <absorbed-tip>`), not from a literal SHA. If a branch moves, the lane merges it again and records the new tip here.
- Tree of the RED cells in acceptance.md: `6f95917a1`.
