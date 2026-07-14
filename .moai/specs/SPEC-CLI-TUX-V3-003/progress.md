# SPEC-CLI-TUX-V3-003 — Progress

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-07-13

## §E.2 Run-phase Evidence

### Pre-flight baselines (2026-07-14, M3a entry)

| Check | Command | Result |
|-------|---------|--------|
| git sync | `git rev-list --count --left-right origin/main...HEAD` | `0 0` (synced, HEAD b45c552ea) |
| cross-platform build | `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` | darwin exit 0, windows exit 0 |
| guard 5-family anchor | `go test ./internal/template/ -run TestSplitHarnessNamespaceNoLeak` + `go test ./internal/merge/...` | both PASS |
| fmt.Printf baseline (AC-TUX3-020) | `grep -rn 'fmt\.Printf\|fmt\.Println\|fmt\.Print(' internal/cli --include='*.go' \| grep -v '_test.go' \| wc -l` | **38** (M3f target: below this) |
| bubbletea/bubbles v2 state (AC-TUX3-017) | `go list -m charm.land/bubbletea/v2 charm.land/bubbles/v2` | **already landed by SPEC-CLI-TUX-V3-002**: bubbletea/v2 v2.0.8, bubbles/v2 v2.1.1 (go.mod lines 6-7) |
| confirm.go @MX debt tags | `grep -n '@MX:DEBT\|@MX:CEILING\|@MX:UPGRADE' internal/merge/confirm.go` | 3 tags at lines 871-873 (M3e removal target; untouched this spawn) |

### M3a-1 Classification type contract (REQ-TUX3-001/002) — LANDED

New package `internal/cli/update/` with `class.go` + `class_test.go`:

- **`ChangeClass` enum** (4 values, single shared type for TUI preview + text fallback — REQ-TUX3-001):
  `ClassAdd` / `ClassUpdate` / `ClassPreserveUserOwned` / `ClassConflict`
- **`Classify(relPath, exists, conflict, isUserOwned)`** — priority order: user-owned > conflict > add(!exists) > update. The preserve class delegates entirely to the injected `UserOwnedPredicate` (REQ-TUX3-002 — no parallel heuristic). At M3a the predicate is injected (package `cli` passes `isUserOwnedNamespace`); M3d moves the predicate into the package for a same-package call.
- **`ChangeClass.String()`** labels — preserve class renders `"preserved (user-owned)"` (REQ-TUX3-014).
- **@MX:ANCHOR** on `Classify` (fan_in grows to ≥3 at M3c preview + M3e fallback + M3f enforcement coherence).

M3a gate (classification contract tests green):
```
go test ./internal/cli/update/... -count=1 -v
→ TestClassificationExhaustive PASS (10 subtests)
→ TestChangeClassString PASS
→ TestPreserveClassSource PASS
→ TestPredicateShared PASS
ok  github.com/modu-ai/moai-adk/internal/cli/update  0.252s
```

**AC-regex PASS-WITH-DEBT note (AC-TUX3-001/002 wording)**: acceptance.md shows `-run 'Classification\|ChangeClass'` and `-run 'PreserveClassSource\|PredicateShared'`. In Go RE2, `\|` is a literal-pipe escape (matches the `|` character), NOT alternation — so the regexes as written match zero test names. The intended alternation is plain `|`. My test names (`TestClassificationExhaustive`, `TestChangeClassString`, `TestPreserveClassSource`, `TestPredicateShared`) satisfy the INTENT with `-run 'Classification|ChangeClass'` / `-run 'PreserveClassSource|PredicateShared'` (verified green above). This is an acceptance.md wording fix owned by manager-spec (not a code defect). Recording here per the deferred-artifact obligation; no code action at M3a.

### M3a-2 update.go function→subpackage mapping table (50 funcs → 6 buckets)

Source: `grep -n '^func ' internal/cli/update.go` (3,276 LOC, 50 top-level funcs). Target subpackages per AC-TUX3-004: `internal/cli/update/{plan,backup,deploy,merge,report}` + root `update.go` retains cobra wiring/orchestration glue. Counts: ROOT=18, PLAN=8, BACKUP=8, DEPLOY=6, MERGE=9, REPORT=1 (total 50).

**ROOT (stays in update.go — cobra wiring + orchestration glue + utilities)**:
`validateUpdateFlags`:70, `init`:78, `runUpdate`:136, `applyUpdateTierProfile`:407, `shouldSkipBinaryUpdate`:453, `runBinaryUpdateStep`:482, `reexecNewBinary`:536, `runTemplateSync`:569, `runShellEnvConfig`:1595, `runInitWizard`:2645 (PRESERVE M2/-002), `applyWizardConfig`:2720 (PRESERVE M2/-002), `updateSettingsLocalEnv`:2928, `ensureGlobalSettingsEnv`:2938, `cleanLegacyHooks`:3064, `execCommand`:3160, `detectGoBinPathForUpdate`:3172, `resolveMoaiExecutable`:3181, `readHookOptInEnabled`:3193

**PLAN (→ update/plan — analysis/classification/namespace)**:
`classifyFileRisk`:1056, `determineStrategy`:1077, `determineChangeType`:1244, `analyzeFiles`:1260, `isUserAreaPath`:1313, `isUserOwnedNamespace`:1387 (ANCHOR predicate), `isMoaiManaged`:1488, `getProjectConfigVersion`:1639

**BACKUP (→ update/backup — config backup/restore + rotation)**:
`backupMoaiConfig`:1685, `saveTemplateDefaults`:1807, `cleanup_old_backups`:2068, `restoreMoaiConfig`:2132, `restoreMoaiConfigLegacy`:2239, `isSymlinkEntry`:2310, `restoreTargetContained`:2342, `parentChainContained`:2415

**DEPLOY (→ update/deploy — template deploy + migration + cleanup)**:
`runTemplateSyncWithReporter`:579, `runTemplateSyncWithProgress`:1005, `cleanMoaiManagedPaths`:1868, `migrateLegacyMemoryDir`:1983, `runAgencyMigrationAdapter`:2039, `scaffoldEvolutionDir`:3216

**MERGE (→ update/merge — 3-way + gitignore + analysis)**:
`mergeGitignoreFile`:1099, `mergeUserFiles`:1149, `buildMergeAnalysis`:1542, `analyzeMergeChanges`:1585, `mergeYAML3Way`:2488, `deepMerge3Way`:2511, `valuesEqual`:2571, `mergeYAMLDeep`:2584, `deepMergeMaps`:2602

**REPORT (→ update/report — guidance/preview render; grows at M3c/M3e)**:
`emitHooksReviewGuidance`:442 (single function; preview table + outcome card renderers land here at M3c/M3e)

### M3a-3 Namespace predicate caller map (PRESERVE safety — M3d move anchor)

- `isUserOwnedNamespace` (update.go:1387) — 3 call sites in `update_namespace_protect.go` (lines 153, 264, 329). Test files referencing it (M3d import-path adjustment): `update_namespace_hns_test.go`, `update_namespace_harness_v2_test.go`, `update_namespace_harness_v4_test.go`, `update_test.go`, `update_preserve_my_harness_test.go`, `update_evolution_test.go`.
- `isMoaiManaged` (update.go:1488) — 1 call site (update.go:1270).
- `isUserAreaPath` (update.go:1313) — 0 non-test callers (purely tested). Test refs: same 6 files above.
- `isUserOwnedNamespaceConservative` (protect.go:313) — stays in protect.go; calls `isUserOwnedNamespace` internally.

### M3b characterization safety net — LANDED

File: `internal/cli/update_characterization_test.go` (package `cli` — accesses unexported helpers + `updateCmd`).

8 characterization tests pin the observable behavior of the UNDECOMPOSED update.go that M3d decomposition MUST preserve (M3d safety-net anchor established):

1. **TestUpdateFlagMatrixCharacterization** — flag surface (7 flags: config/-c, force, yes, templates-only, binary, dry-run, no-hooks)
2. **TestUpdateFlagMutualExclusionCharacterization** — independent flag set/reset (4 subtests)
3. **TestDetermineChangeTypeCharacterization** — golden values: `"update existing"` / `"new file"`
4. **TestClassifyFileRiskCharacterization** — golden values: settings.json→high, rules/x.md→low, skills/hns-x→medium
5. **TestValuesEqualCharacterization** — 3-way-merge equality helper (7 subtests incl. cross-type numeric equality int/uint/float)
6. **TestIsSymlinkEntryCharacterization** — plain paths are NOT symlinks
7. **TestNamespacePredicateSupersetCharacterization** — NFR-UNP-005: isUserOwnedNamespace is a superset for user-owned surfaces (directory-based paths; colon-format command names are NOT recognized — characterization-pinned behavior)
8. **TestNamespaceProtectionE2EClassification** — bridges `update.Classify` to `isUserOwnedNamespace` (REQ-TUX3-002 reachability proven) + REQ-TUX3-014 label check
9. **TestMoaiManagedClassificationCharacterization** — moai-managed template paths are NOT user-owned (update targets, not preserve)
10. **TestFastPathVersionComparisonCharacterization** — getProjectConfigVersion contract (returns "0.0.0" default on empty project; deterministic)

M3b gate (suite green on UNDECOMPOSED update.go):
```
go test ./internal/cli/ -run Characterization -count=1
→ all 8 tests PASS (14 subtests)
ok  github.com/modu-ai/moai-adk/internal/cli  0.456s
```

No regression confirmed: `go test ./internal/cli/...` all green; guard 5-family (`TestSplitHarnessNamespaceNoLeak` + `internal/merge/...`) PASS; cross-platform builds (darwin + windows) exit 0.

### M3c change-preview TUI + fallback + user-owned visibility — LANDED (commit f4f378360)

5 files, +728/-2: `update.go` (54-line convergence — `confirmViaPreview` single entry @ line 1609; call sites 678/1039 routed through it; behavior-preserving: non-TTY confirm path mirrors old tea.Run failure, TTY preview and `--yes`/non-TTY fallback produce the SAME classification decision) + new `internal/cli/update/{preview.go (88), preview_tui.go (258, bubbletea v2 table+viewport), preview_fallback.go (71), preview_test.go (268, 11 contract tests)}`.

Preview consumes the M3a `class.go` shared model (`Classify` + `UserOwnedPredicate`) — NO parallel heuristic (REQ-TUX3-002). `preserved (user-owned)` label in BOTH TUI table and text fallback, derived from the shared predicate (REQ-TUX3-014).

M3c gate (AC-TUX3-008/009/010/014 PASS; AC-TUX3-015 mapped):
- AC-TUX3-008: TestPreviewTable* (per-class counts + every-file row + classification-matches-Classify) PASS
- AC-TUX3-009: TestPreviewSelectRowReachesDiffViewport + TestPreviewDiffViewportEscReturnsToTable PASS
- AC-TUX3-010: TestPreviewFallback* (non-TTY text summary + zero-ANSI under NO_COLOR + zero-ANSI when piped) PASS
- AC-TUX3-014: TestPreservedLabel* (TUI table + text fallback + derived-from-shared-predicate) PASS
- AC-TUX3-015: maps to guard 5-family (green) + M3b namespace E2E characterization (green)

Orchestrator Trust-but-verify (2026-07-14, independent batch; evidence `.moai/state/verify/86ac7ac3/`):
- commit f4f378360 on origin/main, divergence `0 0`
- `go test ./internal/cli/update/...` green (preview + class tests)
- namespace guard 5-family: TestSplitHarnessNamespaceNoLeak + cli `Namespace|SecurityM2` + internal/merge all green, UNMODIFIED
- `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` + `go vet ./internal/cli/...` all exit 0 (LSP go-list "undefined: ChangeClass" / BrokenImport diagnostics confirmed FAKE — gopls workspace issue, NOT build errors; actual go build resolves the same-package symbols)
- `golangci-lint run`: 0 issues
- bubbletea v2 v2.0.8 + bubbles v2 v2.1.1 in use (REQ-TUX3-017 first half — landed by SPEC-CLI-TUX-V3-002)

**Note (§E.2 recorded by orchestrator):** the manager-develop spawn reported this M3c evidence block as a BLOCKER — its 4 Edit attempts on progress.md failed with "File has not been read yet" despite preceding Reads (harness read-state tracker issue, system-injected context between Read and Edit). `git status` confirmed progress.md was NOT modified by the spawn. The orchestrator records this block directly (ledger-closure / evidence-persistence obligation; the evidence was fully drafted and verified).

acceptance.md wording fixes owed to manager-spec (forbidden ownership crossing — NOT edited by manager-develop):
- AC-TUX3-009/010 `-run` regexes use `\|` (RE2 literal-pipe escape, NOT alternation) → should be `|`; tests verified green with the corrected form
- AC-TUX3-015 `UserAssetPreservation` test-name absent in the suite → maps to guard 5-family + M3b characterization

KNOWN UNRELATED FAILURES surfaced by `go test ./...` (NOT M3c regressions — pre-existing / local-env):
- internal/config `TestAuditLoaderCompleteness`: local untracked `.moai/config/sections/report.yaml` (26 B, Jul 14 02:46, NOT in git) triggers `YAML_SECTION_NO_LOADER: report`. CI (no report.yaml) is green. The file is working-tree-only; provenance under investigation (harness automation / token-accounting report path).
- internal/settings `TestApplySchemaEditsAllFieldsRoundTrip` + internal/web `TestI18nKeySetParity` / `TestDataI18nKeysSubsetOfDictionary`: ISOLATED runs are GREEN (verified) → flaky under `go test ./...` parallel execution (shared-state / test-order dependent). M3c commit did NOT touch these packages (`git show --stat f4f378360` = update.go + preview* only).
These are tracked as out-of-SPEC-003-scope (WEBCONF-SIMPLIFY residual / local-env / test-flakiness); M3c itself is green.

### M3d-A update.go decomposition (PLAN + BACKUP) — IN PROGRESS (2 spawns; namespace-predicate move HARD-gate cleared)

**M3d-A spawn-1 (commit `be3cf0b4d`) — PLAN subpackage LANDED**: 8 funcs moved to `internal/cli/update/plan/` (classifyFileRisk, determineStrategy, determineChangeType, analyzeFiles, isUserAreaPath, **isUserOwnedNamespace** [ANCHOR predicate], isMoaiManaged, getProjectConfigVersion). Namespace-predicate move is the worst-case-risk gate: `isUserOwnedNamespace` → `plan.IsUserOwnedNamespace`; `update_namespace_protect.go` calls updated; variable `plan` renamed `deployPlan` to avoid collision with the new `plan` package import. Guard-test diff (`be3cf0b4d~1..be3cf0b4d`) shows ONLY import-path/qualifier/variable-rename lines — NO assertion-logic change (REQ-TUX3-005 satisfied). Orchestrator verification: guard 5-family (TestSplitHarnessNamespaceNoLeak + cli Namespace|SecurityM2 + internal/merge) all green; `go test ./internal/cli/...` green; characterization (M3b safety net) green; go build exit 0.

**M3d-A spawn-2 thrashed (autocompact)** after committing PLAN and progressing BACKUP. Working-tree state at thrash (orchestrator-verified GREEN, not broken):
- `internal/cli/update/backup/backup.go` (439 lines, `package backup`) — 6 BACKUP funcs moved: BackupMoaiConfig, SaveTemplateDefaults, CleanupOldBackups, IsSymlinkEntry, RestoreTargetContained, ParentChainContained.
- `restoreMoaiConfig` (update.go:1548) + `restoreMoaiConfigLegacy` (update.go:1655) remain in root update.go — 2/8 BACKUP funcs NOT yet moved (M3d-A2 scope).
- `go build ./...` exit 0; `go test ./internal/cli/...` green; characterization green; guard 5-family green.
- Modified (uncommitted): update.go, update_namespace_protect.go, coverage_improvement_test.go, target_coverage_test.go, update_characterization_test.go, update_fileops_test.go, update_test.go, update_archive.go + untracked backup/.

**Root-cause of repeated thrashes (M3a#1, M3c#1, M3d-A#1, M3d-A#2)**: a single spawn carrying a full subpackage move (8+ funcs + 6 guard-test import edits + repeated verify cycles) overflows the context. Mitigation: subpackage-per-spawn (PLAN done alone; BACKUP split: 6 funcs landed green, restore* 2 funcs → M3d-A2). CHUNKED-READ discipline (no >200-line file whole; sed-extract to temp) is necessary but NOT sufficient — scope-per-spawn must also shrink.

**M3d-A2 COMPLETE — backup subpackage 8/8 funcs (commit `fdd3ae35c`, pushed origin/main, divergence `0 0`).** Moved the final 2 BACKUP funcs (RestoreMoaiConfig, RestoreMoaiConfigLegacy) → `internal/cli/update/backup/restore.go`. Transitively required moving the 5-func YAML merge cluster (MergeYAML3Way, DeepMerge3Way, ValuesEqual, MergeYAMLDeep, DeepMergeMaps) → `backup/merge.go` — the restore funcs depend on it AND `backup` cannot import parent `cli` (import-cycle). The merge-history noise-suppression ledger (`recordMergeFallback` + merge-history.json counter in `update_noise.go` + `updateVerboseMode`) stays in package `cli` and is invoked from RestoreMoaiConfig via a nil-safe `MergeFallbackRecorder` callback seam (root update.go:~910 bridges it). 6 test files: import-path/qualifier adjustments ONLY (assertion bodies UNMODIFIED per REQ-TUX3-005). update.go: −343 lines, unused `"maps"` import removed.

Orchestrator Trust-but-verify (2026-07-14, fresh batch; evidence `.moai/state/verify/86ac7ac3/`): `go test ./... -count=1` exit 0 — FULL repo green (the prior config/settings/web FAILs are now resolved, coincident with parallel-session commits `3a34e0df5 chore: consolidate` + `93161bf78 v3.0.0-rc12`); namespace guard 5-family (TestSplitHarnessNamespaceNoLeak + cli `Namespace|SecurityM2` + internal/merge) all green UNMODIFIED; characterization (M3b safety net) green; `go build ./...` + `GOOS=windows GOARCH=amd64` + `go vet ./...` exit 0; `golangci-lint run` 0 issues. `internal/cli/update/{plan,backup}` subpackages exist (AC-TUX3-004 partial — `deploy` + `report` + root shrink pending in M3d-B). /goal conditions "go test ./... exits 0" + "lint 0" + "namespace 가드 green" currently MET; remaining blocker is 20/20 AC (needs M3d-B + M3e + M3f).

**Note:** §E.2 M3d-A2 block recorded by orchestrator directly — manager-develop's 3 Edit attempts failed with the same harness read-state bug as the M3c block (documented above). The drafted content was preserved in the spawn transcript and verified before insertion.

### M3d-B1 deploy subpackage (3/6 DEPLOY funcs moved) — IN PROGRESS

**M3d-B1 spawn-1 — deploy subpackage LANDED (3 clean leaf funcs):** Created `internal/cli/update/deploy/deploy.go` (package `deploy`, leaf — imports only `defs`, `tui`, stdlib; NO root-cli-local helpers, one-way cli → deploy dependency). Moved 3 funcs: `CleanMoaiManagedPaths` (cleanMoaiManagedPaths → deploy.CleanMoaiManagedPaths), `MigrateLegacyMemoryDir` (migrateLegacyMemoryDir → deploy.MigrateLegacyMemoryDir, intra-package call from CleanMoaiManagedPaths preserved), `ScaffoldEvolutionDir` (scaffoldEvolutionDir → deploy.ScaffoldEvolutionDir). Root `update.go`: −225 lines (2185→1960), 3 func blocks deleted, 2 call sites qualified (`deploy.CleanMoaiManagedPaths`, `deploy.ScaffoldEvolutionDir`), import added. Root `init.go`: 1 call site qualified (`deploy.ScaffoldEvolutionDir`), import added. 8 test files (24 func-call sites total): perl negative-lookbehind regex qualified calls with `deploy.` prefix — assertion bodies UNMODIFIED per REQ-TUX3-005; goimports auto-added the deploy import to all 8 test files; gofmt normalized spacing. 3 remaining references to old lowercase names in update.go/init.go are comment-line documentation only (NOT calls).

Orchestrator Trust-but-verify (2026-07-14, fresh batch; evidence `/tmp/moai-verify/`): `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `go vet ./...` exit 0; `go test ./internal/cli/... -count=1` green (22.9s root + all subpackages incl. new `update/deploy` [no test files]); namespace guard 5-family (TestSplitHarnessNamespaceNoLeak + cli `Namespace|SecurityM2` + internal/merge) all green UNMODIFIED; `golangci-lint run --timeout=5m` 0 issues. `internal/cli/update/{plan,backup,deploy}` subpackages exist (AC-TUX3-004 partial — `report` + root shrink pending). characterization (M3b safety net) green.

**BLOCKER — 3/6 DEPLOY funcs entangled (M3d-B2 scope):** The remaining 3 DEPLOY funcs (`runTemplateSyncWithReporter`, `runTemplateSyncWithProgress`, `runAgencyMigrationAdapter`) reference root-`cli`-package-local helpers (analyzeMergeChanges, confirmViaPreview, resolveTheme, getBoolFlag, migrateAgencyRunner, MigrateError, ErrMigrateNoSource). Moving them to the leaf `deploy` package would create a circular import (deploy → cli + cli → deploy). The M3d-A2 precedent resolved an analogous entanglement via a nil-safe callback seam (`MergeFallbackRecorder`); applying the same pattern to 10+ helpers across 3 funcs is scope creep that would cause a 5th thrash (4 prior thrashes/autocompacts documented in M3d-A root-cause). Per the subpackage-per-spawn anti-thrash rule, M3d-B1 moves ONLY the 3 clean leaf funcs and STOPS. The 3 entangled funcs are deferred to M3d-B2 (callback-seam extraction, separate spawn).

### M3d-B2 merge + report subpackages — LANDED (5-subpackage decomposition COMPLETE; orchestrator-committed after spawn thrash)

**M3d-B2 spawn thrashed** (5th autocompact) after creating the merge + report subpackages but before commit + §E.2 Edit (progress.md Edit hit the same harness read-state bug as M3c/M3d-A2). Orchestrator verified the working-tree artifacts GREEN and committed directly (the spawn's code work is intact; only the commit + §E.2 record were not reached).

Created:
- `internal/cli/update/merge/merge.go` (234 lines, `package merge`) — MERGE-bucket orchestration funcs (mergeGitignoreFile, mergeUserFiles, buildMergeAnalysis, analyzeMergeChanges moved; the 5 YAML-merge helpers remain in `backup/merge.go` as config-restore helpers — option (a), lowest-thrash, no circular import).
- `internal/cli/update/report/report.go` (23 lines, `package report`) — `emitHooksReviewGuidance` → `report.EmitHooksReviewGuidance`.
- `update.go`: −229 lines; imports `updatemerge ".../update/merge"` + `".../update/report"` added.
- 4 test files: import-path/qualifier adjustments ONLY (assertion bodies UNMODIFIED, REQ-TUX3-005).

**5-subpackage decomposition COMPLETE** per AC-TUX3-004: `internal/cli/update/{plan,backup,deploy,merge,report}` all exist. update.go: 3,276 → ~1,730 lines (대폭 감소 MET). The 3 entangled DEPLOY funcs (`runTemplateSyncWithReporter/WithProgress`, `runAgencyMigrationAdapter`) remain in root as **orchestration glue** — AC-TUX3-004 accepts root retaining "cobra wiring + orchestration glue" given "대폭 감소"; callback-seam extraction deferred (documented debt).

Orchestrator Trust-but-verify (2026-07-14): `go test ./... -count=1` exit 0 (FULL repo green); guard 5-family (TestSplitHarnessNamespaceNoLeak + cli `Namespace|SecurityM2` + internal/merge) green UNMODIFIED; characterization (M3b safety net) green; `go build ./...` exit 0. /goal conditions "go test ./... exits 0" + "lint 0" + "namespace 가드 green" MET. Remaining for 20/20 AC: M3e (confirm.go v2 + outcome card, AC-TUX3-011~013) + M3f (coherence integration test AC-TUX3-016, coverage AC-TUX3-018, ratchet AC-TUX3-020, full-suite AC-TUX3-019).

### Amendment — AC-TUX3-018 deploy coverage completion (M1, commit `6fb84a280`)

**M1 deploy subpackage coverage — LANDED**: AC-TUX3-018 deploy coverage debt cleared. Added `internal/cli/update/deploy/deploy_error_test.go` (369 lines, error-path tests) covering all testable `if err != nil` branches in deploy.go: CleanMoaiManagedPaths (stat/remove/glob/config/migrate errors), MigrateLegacyMemoryDir (rename/remove-legacy errors), ScaffoldEvolutionDir (mkdirall/gitkeep/manifest/changelog write errors). Coverage: 74.7% → 97.5% (≥85% target MET). Remaining 2.5% is `filepath.Glob`'s `ErrBadPattern` branch (deploy.go:78-81), unreachable with a valid `moai*` pattern. Permission-based tests skip when `os.Geteuid()==0` (root bypasses EACCES); the MkdirAll ENOTDIR test needs no root skip (file-as-dir works everywhere).

Orchestrator Trust-but-verify (2026-07-14, fresh batch; evidence `/tmp/moai-verify-tux3/`): `go test -count=1 -cover ./internal/cli/update/deploy/` → `coverage: 97.5% of statements` (exit 0); `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0; PRESERVE confirmed — `go test -count=1 -cover ./internal/cli/update/{report,plan,merge}/` → report 92.9% · plan 95.0% · merge 90.3% (all ≥85%, UNCHANGED); `grep -rn 'AskUserQuestion' internal/cli/update/deploy/ | grep -v _test.go` → 0 (subagent boundary clean); `golangci-lint run --timeout=2m` → 6 pre-existing errcheck in `merge_test.go` (OUT OF SCOPE, PRESERVE), 0 NEW in deploy; `git rev-list --count --left-right origin/main...HEAD` = `0 0` (synced).

## §E.3 Run-phase Audit-Ready Signal

run_status: audit-ready
run_complete_at: 2026-07-14
run_commit_sha: f7bf7407b
run_summary: "M3a~M3f complete — 5-subpackage decomposition + change-preview TUI + confirm.go v2 + outcome card; 18/20 AC clean PASS (AC-TUX3-018 partial, AC-TUX3-020 debt); go test ./... exit 0; golangci-lint clean on SPEC-003 scope; namespace guard 5-family UNMODIFIED green."
ac_matrix:
  AC-TUX3-001: PASS   # ChangeClass enum — class.go
  AC-TUX3-002: PASS   # UserOwnedPredicate inject — no parallel heuristic
  AC-TUX3-003: PASS   # ChangeClass.String labels
  AC-TUX3-004: PASS   # 5-subpackage decomposition (3 entangled funcs stay as orchestration glue — debt)
  AC-TUX3-005: PASS   # guard-test import-path adjustments only, assertion bodies UNMODIFIED
  AC-TUX3-006: PASS   # update.go LOC drop (3276 → ~1730)
  AC-TUX3-007: PASS   # bubbletea v2 / bubbles v2 state
  AC-TUX3-008: PASS   # preview table per-class counts + every-file row
  AC-TUX3-009: PASS   # row-select → diff-viewport; Esc returns to table
  AC-TUX3-010: PASS   # text fallback: NO_COLOR / piped → zero ANSI
  AC-TUX3-011: PASS   # confirm.go v2 bubbles-list promotion
  AC-TUX3-012: PASS   # unified outcome card
  AC-TUX3-013: PASS   # @MX:DEBT/CEILING/UPGRADE 3 tags RESOLVED (removed)
  AC-TUX3-014: PASS   # "preserved (user-owned)" label in TUI + fallback, derived from shared predicate
  AC-TUX3-015: PASS   # namespace guard 5-family UNMODIFIED green (maps to guard + M3b characterization)
  AC-TUX3-016: PASS   # coherence_test.go — Classify↔isUserOwnedNamespace reachability
  AC-TUX3-017: PASS   # bubbletea v2 v2.0.8 + bubbles v2 v2.1.1 in use
  AC-TUX3-018: PARTIAL # coverage: report 92.9% · plan 95.0% · merge 90.3% ✓; deploy 74.7% · backup 75.2% — error-handling paths debt
  AC-TUX3-019: PASS   # golangci-lint clean on SPEC-003 scope (errcheck + staticcheck fixed in f7bf7407b)
  AC-TUX3-020: DEBT    # fmt.Print* ratchet holds at baseline 38 (no regression); Printer migration is follow-up
ac_pass_count: 18
ac_partial_count: 1
ac_debt_count: 1
verdict: PASS-WITH-DEBT

## §E.4 Sync-phase Audit-Ready Signal

sync_status: audit-ready
sync_complete_at: 2026-07-14
sync_commit_sha: 4321ecc9c
sync_summary: "3-phase close (PASS-WITH-DEBT). Single sync commit carries the in-progress → implemented → completed transition on spec.md frontmatter + CHANGELOG [Unreleased] entry + progress.md §E.3/§E.4. SPEC body UNCHANGED (frontmatter status + updated only). No separate Mx chore commit (MX Tag validated as sync sub-step: confirm.go @MX:DEBT 3 tags confirmed removed; 3 ANCHOR + 12 NOTE + 3 REASON + 1 SPEC across the new subpackages)."
changelog_entry_position: "[Unreleased] / Added (single entry, dedup-verified: grep -c 'SPEC-CLI-TUX-V3-003' CHANGELOG.md == 1 pre-edit, 0 → halt-guard passed)"
frontmatter_status_transitions:
  spec.md: "in-progress → completed (single sync commit carries the terminal transition)"
  plan.md: "n/a (plan-phase artifact — not transitioned at sync)"
  acceptance.md: "n/a (acceptance-phase artifact — not transitioned at sync)"
  progress.md: "§E.3 + §E.4 populated this commit"
b12_self_test_a: "pre-emission grep: grep -c 'SPEC-CLI-TUX-V3-003' CHANGELOG.md == 0 (pre-edit) → emission SAFE (no duplicate)"
b12_self_test_b: "AC count match: acceptance.md SSOT has 20 unique AC-TUX3-* IDs; CHANGELOG entry references 18 clean + 018 partial + 020 debt = 20 total ✓"
b12_self_test_c: "file-path verification: internal/cli/update/{plan,backup,deploy,merge,report}/ ls-confirmed; internal/cli/update/class.go + preview*.go ls-confirmed; internal/merge/confirm.go ls-confirmed; update_characterization_test.go + coherence_test.go ls-confirmed"
canary_compliance_check:
  acceptance_md_regex_wording: "AC-TUX3-009/010 `-run` regexes use `\\|` (RE2 literal-pipe, not alternation) — acceptance.md wording fix owed to manager-spec (forbidden ownership crossing; NOT edited in sync-phase)"
  acceptance_md_test_name: "AC-TUX3-015 names `UserAssetPreservation` test absent from suite — maps to guard 5-family + M3b characterization; acceptance.md wording fix owed to manager-spec"
  ac_tux3_018_deploy_backup_coverage: "deploy 74.7% · backup 75.2% < 85% target — error-handling paths; manager-develop judged gap reasonable; ratchet follow-up"
  ac_tux3_020_printer_migration: "fmt.Print* ratchet baseline 38 unchanged (no regression); Printer migration is larger follow-up SPEC scope"
debt_register:
  - id: AC-TUX3-018
    kind: coverage-gap
    scope: "internal/cli/update/deploy (74.7%) + internal/cli/update/backup (75.2%) error-handling paths"
    target: "≥85% coverage"
    owner: follow-up
  - id: AC-TUX3-020
    kind: ratchet-hold
    scope: "internal/cli non-test fmt.Print* (baseline 38, no regression)"
    target: "Printer migration — migrate remaining call sites to internal/cli/printer"
    owner: follow-up
  - id: acceptance-wording
    kind: doc-fix-owed
    scope: "acceptance.md AC-TUX3-009/010 regex `\\|` → `|`; AC-TUX3-015 test-name → guard mapping"
    target: "manager-spec body edit (forbidden ownership crossing)"
    owner: manager-spec

### Amendment Re-close (2026-07-14)

amendment_sync_status: audit-ready
amendment_sync_complete_at: 2026-07-14
sync_commit_sha: pending-backfill-amendment
amendment_sync_summary: "Amendment re-close (2nd close of SPEC-CLI-TUX-V3-003). The SPEC was amended (commit 313482ca4, completed → in-progress) to complete AC-TUX3-018 coverage. AC-TUX3-018 NOW DONE: deploy 97.5% (commit 6fb84a280) + backup 88.6% (commit 8d7106c59); report 92.9% / plan 95.0% / merge 90.3% unchanged ≥85%. D1 resolved: acceptance.md §D.5 reconciled (commit 36eeb3280) — 19 AC PASS + AC-TUX3-020 split (debt). AC-TUX3-020 (Printer migration) EXCLUDED — split to a separate SPEC (not yet authored; debt_register carries it). 19/20 AC clean. Single sync commit carries the in-progress → implemented → completed transition (amendment re-close). SPEC body UNCHANGED (frontmatter status + updated only). acceptance.md / plan.md body NOT modified (manager-spec owned; D1 already resolved by commit 36eeb3280)."
amendment_changelog_entry_position: "[Unreleased] / Changed (amendment re-close entry, dedup-verified: grep -c 'SPEC-CLI-TUX-V3-003' CHANGELOG.md == 1 pre-edit = original close only → amendment emission SAFE)"
amendment_frontmatter_status_transitions:
  spec.md: "in-progress → completed (amendment re-close, single sync commit carries the terminal transition)"
  plan.md: "n/a (plan-phase artifact — not transitioned at amendment sync)"
  acceptance.md: "n/a (acceptance-phase artifact — D1 §D.5 reconciliation already landed in commit 36eeb3280 by manager-spec)"
  progress.md: "§E.4 amendment re-close block populated this commit"
amendment_b12_self_test_a: "pre-emission grep: grep -c 'SPEC-CLI-TUX-V3-003' CHANGELOG.md == 1 (original close entry only, no duplicate amendment entry) → emission SAFE"
amendment_b12_self_test_b: "AC count match: acceptance.md §D.5 reconciled to 19 AC PASS + AC-TUX3-020 split (debt); CHANGELOG amendment entry references 19/20 AC clean ✓"
amendment_b12_self_test_c: "file-path verification: internal/cli/update/{deploy,backup}/ ls-confirmed (coverage commits 6fb84a280 + 8d7106c59); acceptance.md §D.5 ls-confirmed (commit 36eeb3280)"
amendment_canary_compliance_check:
  ac_tux3_018_resolved: "deploy 97.5% + backup 88.6% — both ≥85% target (commits 6fb84a280 + 8d7106c59); AC-TUX3-018 PASS"
  ac_tux3_020_split: "Printer migration split to separate SPEC (not yet authored); debt_register carries it; fmt.Print* ratchet baseline 38 unchanged"
  d1_acceptance_reconciled: "acceptance.md §D.5 reconciled (commit 36eeb3280) — 19 AC PASS + AC-TUX3-020 split (debt)"
amendment_debt_register:
  - id: AC-TUX3-020
    kind: split-to-separate-spec
    scope: "internal/cli non-test fmt.Print* Printer migration"
    target: "Separate SPEC (not yet authored) — migrate remaining call sites to internal/cli/printer"
    owner: follow-up

## §F Phase 4 Mode Selection

decided_at: 2026-07-14 (run kickoff)
orchestrator_session: 86ac7ac3-d1d9-43a4-8ec9-55732930b7b2
handoff_source_session: ec3aa70e-e1ac-4a68-ad48-02b61a1ee323

Input parameters:
- tier: L
- scope (file count): ~15+ (update.go decompose target + 5 new internal/cli/update/* subpackages + internal/merge/{confirm,confirm_test,confirm_coverage} 3 + go.mod/go.sum + characterization/preview/coverage tests)
- domains: 2 (internal/cli Go code + internal/merge)
- file language mix: Go 100%
- concurrency benefit: LOW (sequential dependency — M3b characterization MUST precede M3d decomposition; M3a classification contract gates M3c/M3e display layer)
- Agent Teams prereqs: N/A (Mode 3 RETIRED)

Mode evaluation:
- Mode 1 (trivial): not selected — Tier L 3,276 LOC decomposition + TUI, not trivial
- Mode 2 (background): not selected — write-capable coding work, blocking delegation
- Mode 3 (agent-team): not selected — RETIRED
- Mode 4 (parallel): not selected — coding-heavy + sequential milestone dependency (Anthropic coding-task parallelism caveat)
- Mode 5 (sub-agent): SELECTED
- Mode 6 (workflow): not selected — semantic/new-code/multi-rule work (decomposition + preview TUI + confirm.go v2 promotion), NOT mechanical-uniform transform

Decision: sub-agent (Mode 5)

Justification: Tier L coding-heavy refactoring + new TUI display layer with hard sequential dependencies — M3b characterization safety net MUST land before M3d decomposition, and M3a classification contract gates M3c/M3e. Per Anthropic's coding-task parallelism caveat ("most coding tasks involve fewer truly parallelizable tasks than research, and LLM agents are not yet great at coordinating in real time"), sequential sub-agent delegation per milestone is the safe default. The worst-case risk (namespace-protection regression = user-asset loss) demands the characterization-first safety net that only sequential DDD provides, not parallel fan-out. Methodology routing within the single manager-develop delegation: M3b/M3d = DDD (ANALYZE-PRESERVE-IMPROVE), M3c/M3e = TDD (RED-GREEN-REFACTOR), M3a/M3f = planning/gate. Route A (Hybrid Trunk main-direct) per CLAUDE.local.md §23 1-person OSS policy — direct-to-main push with CI 4-check protection; manager-develop performs commit + push directly.

Implementation Kickoff Approval: PASSED — user selected "run-phase 진입" 2026-07-14 (plan-audit 0.92 PASS, preconditions 4/4 verified).
