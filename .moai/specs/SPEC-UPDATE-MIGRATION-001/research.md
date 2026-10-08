# research.md — SPEC-UPDATE-MIGRATION-001 (card t1547)

Read-only codebase study performed 2026-10-07 in worktree `t1547` (branch `WT-update-migration`, HEAD `cad44a751` at session start).

## 1. The destruction site — `internal/cli/update/deploy/deploy.go`

- `ManagedCleanTargets(projectRoot)` (deploy.go:56) — the fixed clean list: `.claude/settings.json`, `.claude/commands/moai`, `.claude/agents/moai`, `.claude/skills/moai*` (glob), `.claude/rules/moai`, `.claude/output-styles/moai`, `.claude/hooks/moai`. `CleanMoaiManagedPaths` additionally removes `.moai/config/` inline (deploy.go:201-215).
- `CleanMoaiManagedPathsWithTargets` (deploy.go:119) — the generalized form (SPEC-INIT-SHRINK-001): per-target machinery (progress lines, t111 pre-clean backup, symlink dispositions, backup-then-remove ordering) is identical; only the target list differs. **This is the seam the reconciliation pipeline can enter without duplicating machinery.**
- `backupThenRemove` (deploy.go:437): backs up only files the template does NOT carry, then removes everything. Files the template DOES carry are removed without backup (deployment rewrites them) — including user-modified content. This is the silent-overwrite hazard: a user-edited template-carried file is destroyed and replaced by the new render with no merge and no report.
- `InventoryManagedPathsWithTargets` (deploy.go:241) — read-only snapshot of what the clean will touch; already exists for preview/accounting.
- Symlink discipline (SPEC-CLI-CLEAN-SYMLINK-001): Lstat-first classification, link-dedicated branch, never follows links. Must be carried into any new walk.
- REQ-UDS-008 ordering (SPEC-UPDATE-DATA-SURVIVAL-001): backup failure aborts before removal. Load-bearing invariant to preserve.

## 2. The classification machinery — already built

- `internal/cli/update/migrate_classify.go` (SPEC-INIT-SHRINK-001 REQ-010/013): `ClassifyMigration` walks dropped roots and classifies every regular file identical / modified / foreign. Class gate is TEMPLATE CARRIAGE, not managed-name matching. `classifyCarried` uses the manifest: healthy provenance (`template_managed` + matching `CurrentHash`) + content equal to render → identical; absent/stale record or diverging content → modified (conservative). Foreign (template does not carry) → preserved byte-for-byte, never archived. Symlinks recorded, never classified.
- This classifier is scoped to the DROPPED component roots (`.claude/skills`, `.claude/commands`, `.agents/skills`) for the init-shrink migration. The t1547 pipeline needs the same three-way logic over the MANAGED roots — an extension of scope, not a new algorithm.
- `internal/manifest` — `FileEntry{Provenance, CurrentHash}`; provenance values `template_managed` / `user_modified` / `user_created` (verify exact enum names in run phase); `manifest.HashBytes`.

## 3. The merge machinery — already built

- `internal/cli/update/plan/plan.go`: `DetermineStrategy(filename)` → `merge.SectionMerge` (CLAUDE.md), `EntryMerge` (.gitignore), `JSONMerge`, `YAMLDeep`, `LineMerge`; `ClassifyFileRisk`. `IsUserOwnedNamespace` (plan.go:152, fan_in 3) — hns-/harness-/user skills/agents namespace predicates, platform-normalized. `IsMoaiManaged` (plan.go:236) — also protects `.moai/evolution/`.
- `internal/cli/update/merge/merge.go`: 3-way merge orchestration over backups (`FileBackup`), imports `internal/merge` engine (aliased `mrg`), `MergeGitignoreFile` with `UserPatternsMarker` verbatim-carry. settings/gitignore/MCP snapshot flows have characterization tests already.
- `internal/merge` — the engine (`FileAnalysis`, `MergeAnalysis`, `Engine`, conflict reporting).

## 4. The counting blindspot

- `internal/cli/update/plan/plan.go:73`: `if IsMoaiManaged(...) { continue }` — managed files are excluded from the `Updated N files` count. Deletions appear nowhere in the summary. `internal/cli/update/report/outcome.go` / `report.go` — the existing outcome-report structure to extend.

## 5. Config wipe + merge asymmetry

- `.moai/config` is wiped (deploy.go:201-215) then the Backup step's restore merges sections back (`internal/cli/update/backup/`). The merge restores values for keys the TEMPLATE also carries; user-only keys (`git_strategy.manual.workflow: git-flow`, `worktree_base_branch: develop` — verified absent from the template default in the 2026-09-24 incident, `.moai/reports/t1159/measurement.md`) are lost. `mrg.YAMLDeep` exists and is the natural preservation vehicle — the flow just never uses it in the preserve direction.

## 6. Callers / flow wiring

- `internal/cli/update.go` — the update command orchestration; `archiveLegacySkills` historically called after the wipe (t750 defect ②, since restructured by SPEC-INIT-SHRINK-001 — re-verify call order in run phase).
- `--dry-run` does not preview `CleanMoaiManagedPaths` deletions (t750 defect ③); `InventoryManagedPaths` exists as the preview primitive.

## 7. Test landscape to build on

- `deploy_test.go` (20K), `deploy_preclean_backup_test.go`, `deploy_symlink_*_test.go` (3 files), `deploy_contract_test.go`, `deploy_error_test.go` — characterization base for the deploy package.
- `migrate_classify_test.go` (10K) — classifier behavior base.
- `merge/merge_test.go` (27K), settings_snapshot/base tests — merge base.
- All tests use `t.TempDir()` isolation per repo convention.

## 8. Key findings for the design

1. The reconciliation pipeline is mostly an ORCHESTRATION change: classify first (extend `ClassifyMigration` semantics to managed roots), then route each class to refresh / merge / preserve / archive-remove. The per-path machinery (backup ordering, symlink handling, progress lines) exists in deploy.
2. `CleanMoaiManagedPaths` must not be deleted outright — the v1→v2 legacy fresh-install path and `CleanMoaiManagedPathsWithTargets` callers depend on it. The update-on-existing-project flow routes around it; the wholesale path is retained behind REQ-UPM-040/015.
3. The `.moai-new` conflict sidecar has no existing precedent; it is new surface. Alternative considered (conflict → `.rej`-style markers inside the file) rejected: it mutates the user's file, violating the preserve-untouched rule.
4. The `Updated N files` fix touches `plan.go:73` — the count exclusion was deliberate (avoid double counting wiped files). Once managed files are no longer wiped, the exclusion's premise is gone and managed files join the count with their own outcome categories.
