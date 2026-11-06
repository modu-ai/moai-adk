# t2132 progress (#1969 part 2, 3/3): Windows-only test failures, POSIX mode bits

## Claim

The 14 listed Windows-only test failures are fixed in test files only. On Windows each test now asserts only the portable part (existence / readability); POSIX mode-bit equality (0600/0644/0700) is gated with `runtime.GOOS != "windows"`. The one site whose only purpose is a read-only-directory failure is skipped on Windows with a printed reason. POSIX behavior and assertions are unchanged. No production code changed.

## Per-site disposition

| # | Site | Disposition |
|---|---|---|
| 1 | civerdict TestErrorPaths/save-readonly-dir | `t.Skip` on Windows with reason (subtest's only purpose is the read-only-dir failure) |
| 2 | cli/update TestSafeWriteFilePreservesExistingMode | mode equality gated; content assertion still runs on Windows |
| 3 | graph TestBuildPrivateGTDProjectionPrivacy | stat error still fatal; mode equality gated; all other assertions run |
| 4 | graph TestWriteEdgesMeta_FailureBranches | read-only-dir negative step gated (early return after the portable mkdir-failure step) |
| 5 | factory TestBacklogSQLiteArtifactsArePrivateWhileOpen | mode equality gated; stat of db/-wal/-shm still required |
| 6 | factory TestRelocateQueueArtifacts...KeepsRollbackSource | target-db mode equality gated; existence still required |
| 7 | homestate TestHomeLayoutAndProjectRuntimePermissions | 0700 equality gated (both loops); stat still required |
| 8 | homestate TestProjectLayoutPermissions | 0700 equality gated |
| 9 | homestate TestFactorySQLiteArtifactsArePrivateWhileOpen | 0600 equality gated; stat still required |
| 10 | homestate TestProfileLeasesAreGlobalAndPrivate | stat error fatal; 0600 equality gated |
| 11 | yamlpatch TestPatchFileGreenfieldCreation | 0644 equality gated; content assertion runs |
| 12 | yamlpatch TestYAMLPatchAtomicWriteErrors/stat_missing_target | 0644 equality gated; content assertion runs |
| 13 | template TestSyncLocalSettingsPATH/file_mode_stays_0600 | 0600 equality gated; sync success + stat still required |
| 14 | web TestHandleSaveGreenfieldSectionCreation | 0644 equality gated; content assertion runs |

`runtime` import added where missing (civerdict, cli/update, graph x2, homestate x2, web). No shared helper was needed; the inline gate is one condition per site.

## Evidence

Commands run from the worktree, this run (focused, `-count=1 -v`), verbatim result lines:

- `go test ./internal/civerdict -run TestErrorPaths`: `--- PASS: TestErrorPaths/save-readonly-dir` ... `ok  	github.com/modu-ai/moai-adk/internal/civerdict	0.304s`
- `go test ./internal/graph -run 'TestBuildPrivateGTDProjectionPrivacy|TestWriteEdgesMeta_FailureBranches'`: `--- PASS: TestBuildPrivateGTDProjectionPrivacy (0.03s)`, `--- PASS: TestWriteEdgesMeta_FailureBranches (0.17s)`, `ok  	github.com/modu-ai/moai-adk/internal/graph	0.967s`
- `go test ./internal/homestate -run '<4 names>'`: PASS x4 (`TestHomeLayoutAndProjectRuntimePermissions`, `TestProjectLayoutPermissions`, `TestFactorySQLiteArtifactsArePrivateWhileOpen`, `TestProfileLeasesAreGlobalAndPrivate`), `ok  	github.com/modu-ai/moai-adk/internal/homestate	0.905s`
- `go test ./internal/settings/yamlpatch -run 'TestPatchFileGreenfieldCreation|TestYAMLPatchAtomicWriteErrors'`: PASS incl. `read-only_directory` and `stat_missing_target`, `ok  	github.com/modu-ai/moai-adk/internal/settings/yamlpatch	0.260s`
- `go test ./internal/web -run TestHandleSaveGreenfieldSectionCreation`: PASS, `ok  	github.com/modu-ai/moai-adk/internal/web	0.697s`
- `go test ./internal/cli/update -run TestSafeWriteFilePreservesExistingMode`: PASS, `ok  	github.com/modu-ai/moai-adk/internal/cli/update	0.417s`
- `go test ./internal/factory -run 'TestBacklogSQLiteArtifactsArePrivateWhileOpen|TestRelocateQueueArtifactsCopiesLogicalDataAndKeepsRollbackSource'`: PASS x2, `ok  	github.com/modu-ai/moai-adk/internal/factory	0.484s`
- `go test ./internal/template -run TestSyncLocalSettingsPATH`: PASS incl. `file_mode_stays_0600`, `ok  	github.com/modu-ai/moai-adk/internal/template	0.544s`
- `GOOS=windows go vet` on the 8 touched packages (civerdict, cli/update, graph, factory, homestate, settings/yamlpatch, template, web): no output, exit 0.
- `gofmt -l` on the 8 package dirs: no output.

## Baseline-attribution

Measured against the card worktree tree (branch WT-remaining-packages-windo, base HEAD 04d8c27a7e plus the uncommitted edits at measurement time). The tool is the local `go` toolchain; no moai build was used for judging.

## Gaps

- The 14 failures were not reproduced on Windows (this host is darwin). Windows behavior is established by compile (GOOS=windows vet) and by reading the gated conditions; the Windows test run is the CI run on the integration branch (PENDING at report time).
- No whole-suite run (scope policy).
- Production-code mode comparison found, NOT changed per card scope: `internal/graph/gtd_private.go:117` (`mode.Perm() != item.mode` in `AuthorizePrivateGTDProjection`, expecting 0700 dir / 0600 files, fed by `systemPrivateGTDPermissionChecker.Mode` at `gtd_private.go:93`). On Windows this returns `permission_denied` for a genuine projection; needs a separate card. The tests that exercise it with a fake checker are unaffected.
- Other production `Perm()` uses in the touched packages only preserve a mode and do not compare it (`internal/settings/yamlpatch/yamlpatch.go:530`, `internal/cli/update/reconcile.go:377`, `internal/template/skill_mirror.go:319`, `internal/template/retired_model_keys.go:156`, `internal/template/agentemit/sets.go:338`); no finding.
- Production code that SETS 0600/0700 modes was not audited beyond the grep above.

## Residual-risk

Windows may still expose other failures in these tests past the gated lines (the CI report listed only the first failing line per test). The graph meta_tmp test now returns early on Windows, so any step added below it later would be skipped there.
