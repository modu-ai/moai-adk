# SPEC-CLIFIX-HYGIENE-001 — Design

## §A update.go Decomposition Seam Map

Basis: live function inventory (`grep -n '^func ' internal/cli/update.go`, 3,181 lines, captured 2026-07-10). Moves are mechanical (same package, no signature or export changes); line anchors drift — re-derive the inventory at run time before moving.

| Target file | Concern | Function group (verified symbols) |
|---|---|---|
| update_sync.go | Template sync pipeline, path classification, managed-path cleanup | runTemplateSync, runTemplateSyncWithReporter, runTemplateSyncWithProgress, classifyFileRisk, determineChangeType, analyzeFiles, analyzeMergeChanges, buildMergeAnalysis, isUserAreaPath, isUserOwnedNamespace, isMoaiManaged, cleanMoaiManagedPaths, migrateLegacyMemoryDir, runAgencyMigrationAdapter, cleanup_old_backups, scaffoldEvolutionDir |
| update_merge.go | 3-way / deep merge machinery | determineStrategy, mergeGitignoreFile, mergeUserFiles, mergeYAML3Way, deepMerge3Way, valuesEqual, mergeYAMLDeep, deepMergeMaps |
| update_wizard.go | Interactive (re)configuration | runInitWizard, applyWizardConfig, runShellEnvConfig |
| update_settings.go | settings/config persistence, backup/restore, env wiring | updateSettingsLocalEnv, ensureGlobalSettingsEnv, cleanLegacyHooks, saveTemplateDefaults, getProjectConfigVersion, backupMoaiConfig, restoreMoaiConfig, restoreMoaiConfigLegacy, isSymlinkEntry, restoreTargetContained, parentChainContained, readHookOptInEnabled |
| update.go (residual core) | Command wiring + orchestration + binary self-update | init, runUpdate, shouldSkipBinaryUpdate, runBinaryUpdateStep, reexecNewBinary, emitHooksReviewGuidance, execCommand, detectGoBinPathForUpdate, resolveMoaiExecutable |

Move rules:

- One seam per commit, mechanical move only (no logic edits in the move commit — plan.md §D).
- Characterization suite runs after every move; per-file ceiling 1,200 lines (AC-HYG-001-001).
- The path-classifier quartet (isUserAreaPath/isUserOwnedNamespace/isMoaiManaged + shared prefix table extraction, audit cluster-1 Major row 8) lands in update_sync.go so the shared-table refactor has a single home.
- Sequencing note: this SPEC runs last (P0→P4); the P0 lock wiring (runUpdate ← acquireUpdateLock) and any P0 edits to isSymlinkEntry-adjacent restore code are already merged before moves start.

## §B deepMerge3Way Key-Retirement Design (REQ-HYG-001-007)

The current behavior of dropping keys absent from the new template is PARTLY intentional: it is how template-retired keys leave user configs. The defect is only the loss of user-ADDED keys. Disposition is decided against the 3-way inputs (base = template defaults snapshot at last sync, old = user's current config, new = incoming template):

| in new | in base | in old | Classification | Disposition |
|---|---|---|---|---|
| yes | any | any | template-owned / shared key | existing merge semantics (unchanged) |
| no | yes | yes | template-retired key | DROP (current behavior preserved — intentional retirement) |
| no | no | yes | user-added key | PRESERVE from old (the fix) |
| no | yes | no | retired key already absent | no-op |

Fallback — missing base snapshot: legacy trees that predate saveTemplateDefaults have no base. Retirement vs user-addition is then indistinguishable; the conservative default is PRESERVE old-only keys (no data loss) plus a stderr WARN naming the undecidable keys. Run phase adds a fixture for this path; the fallback choice is confirmed at Implementation Kickoff.

Reintroduction collision: a preserved user-added key later reintroduced by the template becomes template-owned on the next sync (row 1 semantics; new wins for template-owned keys — matches the AC-HYG-001-006/007 edge-case note in acceptance.md).

## §C Cross-References

- research.md — dead-code caller-graph inventory feeding REQ-HYG-001-002 (M5 deletion milestone input).
- acceptance.md AC-HYG-001-001 / AC-HYG-001-007 — the machine checks over this design.
- Audit SSOT: `.moai/reports/cli-improvement-audit-20260710.html` §3 cluster 1 (update.go structure rows), §4 row 1 (closed-struct round-trip family — deepMerge3Way listed there).
