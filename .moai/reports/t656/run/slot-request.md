# t656 — `internal/cli` slot request (M4 verification + M5 cli mutants)

Card **t656** · SPEC-UPDATE-SETTINGS-BASE-SNAPSHOT-001 · worktree `.claude/worktrees/t656` · branch `WT-update-value-merge`

The `internal/cli` wiring and its tests are written and committed but have **never been compiled into a test binary or run**. The only checks run on the root package are `go vet ./internal/cli/` (exit 0, type-checks the package and its tests) and `gofmt -l` (empty). Everything below needs the lead's slot.

## 0. Pre-flight inside the slot

```bash
git -C .claude/worktrees/t656 rev-parse --short HEAD        # must equal the commit named in the completion report
git -C .claude/worktrees/t656 status --short internal/        # must be empty
```

Run each command below as its own invocation from the worktree root, one at a time (no background, no `go test ./...`). Every `go test` carries `-count=1 -timeout 900s`.

## 1. GREEN run — new tests (AC-USB-005, 007, 008)

```bash
go test ./internal/cli/ -count=1 -timeout 900s -v -run '^(TestCleanReinstall_SettingsSnapshotStagedBeforeMerge|TestSettingsSnapshot_WriteSites|TestSettingsSnapshot_InitLeftoverJudgementPrecedesExecute|TestSettingsSnapshot_WriteFailureDoesNotBlock)$'
```

Must produce `--- PASS` for all 4 top-level tests and these 9 subtests, and exit 0. Count them — a missing line is a gap, not a pass:

| Test / subtest | AC |
|---|---|
| `TestCleanReinstall_SettingsSnapshotStagedBeforeMerge` | AC-USB-005 (i)–(iv) |
| `TestSettingsSnapshot_WriteSites/clean_reinstall` | AC-USB-007 |
| `…/template_sync_backup_empty` | AC-USB-007 |
| `…/template_sync_backup_filled` | AC-USB-007 |
| `…/update_leftover_abort` | AC-USB-007 |
| `…/update_leftover_version_skip` | AC-USB-007 (+ N3-02 retired-entry hardening) |
| `…/init` | AC-USB-007 |
| `TestSettingsSnapshot_InitLeftoverJudgementPrecedesExecute` | AC-USB-007 source-position substitute for the init leftover judgement (reason in the test comment) |
| `TestSettingsSnapshot_WriteFailureDoesNotBlock/clean_reinstall` | AC-USB-008 |
| `…/update_leftover_promote_failure` | N3-06 |

## 2. RED-stub run — the same tests with the lifecycle calls removed

The tests use seams that exist only with the wiring (`preMergeSettingsSnapshotHook`, `applyAutonomyTierBundleFn`), so the pre-implementation tree does not compile them; per acceptance.md §C a compile error is not RED. Observe RED against the wiring with its five lifecycle calls neutralised and the seams kept:

| File:line (HEAD of this commit) | Change |
|---|---|
| `internal/cli/update.go:384` | delete `backup.JudgeLeftoverSettingsSnapshot(cwd, cmd.ErrOrStderr())` (keep the `cwd` lookup; add `_ = cwd`) |
| `internal/cli/update_template_sync.go:371` | delete `backup.StageDeployedSettingsSnapshot(projectRoot, mgr, errOut)` |
| `internal/cli/update_clean_install.go:467` | delete `backup.StageDeployedSettingsSnapshot(projectRoot, mgr, errOut)` |
| `internal/cli/init.go:875`, `:891`, `:921` | delete the three `backup.*SettingsSnapshot(...)` calls |

Then run the §1 command. Expected: exit 1, with value-assertion failures (not build failures) in AC-005 (ii)/(iii), every `WriteSites` subtest, and `WriteFailureDoesNotBlock/clean_reinstall` (0 write-failed lines). `InitLeftoverJudgementPrecedesExecute` fails on "both must be present". Revert with `git -C .claude/worktrees/t656 restore <files>` and re-run §1 to GREEN.

## 3. Regression run — existing tests touched by the wiring

```bash
go test ./internal/cli/ -count=1 -timeout 900s -v -run '^(TestCleanReinstall_SettingsJSONUserKeysPreserved|TestCleanReinstall_MatchesNormalPathProtection|TestMergeUserFiles_.*|TestUpdateSubsystem_HomeSeamReach|TestRunInit_.*AutonomyTier.*|TestRunInit_SemiAutoAndEmptyAreZeroDelta|TestUpdateDryRun_.*|TestCleanReinstall_.*Mirror.*|TestCleanReinstall_SkippedWarningsNotSurfaced|TestCleanReinstall_PlainDeployerStillDeploys|TestTemplateSync_.*|TestSeamDefault.*)$'
```

Expected: exit 0. Why each is in the set: the clean-reinstall and template-sync restore steps now call `mergeUserFilesSettlingSnapshot` unconditionally; init now calls the bundle through `applyAutonomyTierBundleFn`; `TestUpdateDryRun_ZeroMutation` proves the new runUpdate judgement stays below the `--dry-run` return.

## 4. Cross-platform and lint (slot, root package)

```bash
GOOS=windows GOARCH=amd64 go build ./internal/cli/...
golangci-lint run ./internal/cli/
```

## 5. Mutants that need the slot (M5, cli wiring)

Apply one at a time, run the named test with the §1 flags (`-count=1 -timeout 900s -v -run '<regex>'`), record the four elements, revert with `git -C .claude/worktrees/t656 restore <file>`. 15 mutants.

| ID | File:line | Change | Killing test (`-run`) and expected RED |
|---|---|---|---|
| M-05 | `update_clean_install.go:467`, `:515` | remove the Stage call at :467; after the `mergeUserFilesSettlingSnapshot` call at :515 copy the live `.claude/settings.json` bytes to `backup.SettingsSnapshotPendingPath(projectRoot)` with `os.WriteFile` | `^TestCleanReinstall_SettingsSnapshotStagedBeforeMerge$` — (ii) pending at hook = "" ≠ render |
| M-06c-w | `update_clean_install.go:467` | after the Stage call, `os.Rename(pending, canonical)` | same test — (i) canonical at hook = render, not `{"marker":"prior"}` |
| M-06t-w | `update_template_sync.go:371` | replace the Stage call with a copy of the live file straight to `backup.SettingsSnapshotPath(projectRoot)` | `^TestSettingsSnapshot_WriteSites$/^template_sync_backup_(empty\|filled)$` — `a` = 1 (want 2), `K` absent |
| M-07a | `init.go:891` | move the Stage call below the bundle block (after :919) | `^TestSettingsSnapshot_WriteSites$/^init$` — canonical absent (the gate sees the rewritten file) |
| M-07b | `update_template_sync.go:371` | delete the Stage call | `…/^template_sync_backup_(empty\|filled)$` — canonical = r1, want r2 |
| M-07c | `update_clean_install.go:467` | move the Stage call below `stripRetiredV2DenyEntries` (after :540) | `…/^clean_reinstall$` — canonical = r1, want r2 |
| M-07d | `update_template_sync.go:551` | move the `mergeUserFilesSettlingSnapshot` block inside `if configBackupPath != ""` (:501) | `…/^template_sync_backup_empty$` — `a` = 1 / canonical = r1 |
| M-07e | `init.go:891` | move the Stage call into the `if err != nil` branch after `executor.Execute` | `…/^init$` — canonical absent |
| M-08 | `update_clean_install.go:467` | after the Stage call: `if _, err := os.Stat(backup.SettingsSnapshotPendingPath(projectRoot)); err != nil { return result, recovery.fail("mutant M-08", err) }` | `^TestSettingsSnapshot_WriteFailureDoesNotBlock$/^clean_reinstall$` — runCleanReinstall returns non-nil |
| M-08L | `update.go:384` | after the judgement: `if _, err := os.Stat(backup.SettingsSnapshotPendingPath(cwd)); err == nil { return fmt.Errorf("mutant M-08L") }` | `…/^update_leftover_promote_failure$` — runUpdate returns non-nil |
| M-D5g-wb (N3-01) | `update.go:384` → `update_template_sync.go:425` | move the judgement to the top of `case "Backup":` (use `projectRoot`, `errOut`) | `…/^update_leftover_version_skip$` — canonical = r1, pending left. `update_leftover_abort` stays GREEN (expected: the strip is a no-op on r2) |
| M-D5g-wd (N3-01) | `update.go:384` → `update_template_sync.go:371` | move the judgement to just before the Stage call in Deploy Templates | `…/^update_leftover_abort$` — `a` = 2 (want 3); `…/^update_leftover_version_skip$` — canonical = r1 |
| M-D5g-s (N3-02) | `update.go:384` | move the judgement block below the deny-rule strip block (after :407) | `…/^update_leftover_version_skip$` — canonical = r1 (live stripped ≠ leftover) |
| M-D5f | `update.go:384` → `update_template_sync.go:371` | move the judgement to just after the Stage call | `…/^update_leftover_abort$` — `a` = 2, `L` absent |
| M-D5g-init | `init.go:875` | move the judgement below `executor.Execute` (after the error branch) | `^TestSettingsSnapshot_InitLeftoverJudgementPrecedesExecute$` — offset order failure |

M-D5g-init is killed by a static source-order check, but that check compiles into the root package's test binary, so it needs the slot too.

Mutants already observed without the slot (merge/backup packages): M-01, M-02, M-04, M-06c (= M-06t = M-D5e in the stateless helper), M-08p, M-08s, M-09, M-11, M-14, M-D5a, M-D5b, M-D5c, M-D5d, M-D5i — evidence in `progress.md` §E.2.7 and `mut-*.txt` beside this file.

## 6. Isolation notes and residual risk for the slot

- No test here calls `t.Setenv("HOME", …)`; each cell injects the home through `homeSeamSpy` and asserts `userHomeDirFn` returns the sandbox before driving a flow. The package `TestMain` sandbox (card t661) redirects the real home only for code that routes through `userHomeDirFn`.
- **Residual risk:** `runUpdate` and `runInit` reach code outside `internal/cli` that may resolve the home without the seam (t661 recorded 6 production sites calling `paths.Home`/`os.UserHomeDir` directly, e.g. `runAgencyMigrationAdapter` at `update.go:879`, which the clean-reinstall cells bypass by injecting `RunMigrateAgency`). The `update_leftover_*` and `init` cells were not traced through every callee. The lead may want to snapshot `~/.claude/settings.json` and `~/.moai` mtimes before §1 and compare after.
- The `update_leftover_*` and `init` cells chdir into their project and replace package-level seams; they are not parallel-safe and must not be run with `-parallel` overrides.
