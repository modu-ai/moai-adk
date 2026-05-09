# SPEC-V3R2-RT-007 Task Breakdown

> Granular task decomposition of M1-M5 milestones from `plan.md` §2.
> Companion to `spec.md` v0.1.0, `research.md` v0.1.0, `plan.md` v0.1.0, `acceptance.md` v0.1.0.

## HISTORY

| Version | Date       | Author                            | Description                                                            |
|---------|------------|-----------------------------------|------------------------------------------------------------------------|
| 0.1.0   | 2026-05-10 | manager-spec (Plan workflow)      | Initial task breakdown — 30 tasks (T-RT007-01..30) across M1-M5         |

---

## Task ID Convention

- ID format: `T-RT007-NN`
- Priority: P0 (blocker), P1 (required), P2 (recommended), P3 (optional)
- Owner role: `manager-tdd`, `manager-docs`, `expert-backend` (Go), `manager-git` (commit/PR boundary)
- Dependencies: explicit task ID list; tasks with no deps may run in parallel within their milestone
- TDD alignment: per `.moai/config/sections/quality.yaml` `development_mode: tdd`, M1 (RED) precedes M2-M5 (GREEN/REFACTOR)

[HARD] No time estimates per `.claude/rules/moai/core/agent-common-protocol.md` §Time Estimation. Priority + dependencies only.

---

## M1: Test Scaffolding (RED phase) — Priority P0

Goal: Add ~22 failing tests + 3 audit lint tests + 4 resolver tests across 12 new/extended test files. Per `spec-workflow.md` TDD: write failing test first.

| ID | Subject | Owner role | File:line target | Dependency | Touch points | TDD alignment |
|----|---------|-----------|-------------------|------------|--------------|---------------|
| T-RT007-01 | Create `internal/migration/runner_test.go` (new) with 5 tests: `TestRunner_Apply_HappyPath`, `TestRunner_Apply_Idempotent`, `TestRunner_Apply_FreshInstall_AllInOrder`, `TestRunner_Apply_VersionAhead`, `TestRunner_Apply_FailureHaltsAdvance`, `TestRunner_Apply_PartialSuccess`, `TestRunner_Apply_CrashRecovery`. Uses `t.TempDir()` for projectRoot. | expert-backend | new file (~150 LOC) | none | 1 file (create) | RED — runner doesn't exist yet |
| T-RT007-02 | Create `internal/migration/registry_test.go` (new) with `TestRegistry_DuplicateVersion_Panics`, `TestRegistry_Pending`, `TestRegistry_Highest`. | expert-backend | new file (~80 LOC) | none | 1 file (create) | RED — registry doesn't exist |
| T-RT007-03 | Create `internal/migration/version_test.go` (new) with `TestVersionFile_RoundTrip`, `TestVersionFile_AtomicRename`, `TestVersionFile_AbsentMeansZero`, `TestVersionFile_AdvisoryLock_HighContention`. | expert-backend | new file (~120 LOC) | none | 1 file (create) | RED — version helpers don't exist |
| T-RT007-04 | Create `internal/template/hardcoded_path_audit_test.go` (new) with `TestNoHardcodedAbsolutePath_HookWrappers`, `TestNoHardcodedAbsolutePath_StatusLine`, `TestNoHardcodedAbsolutePath_DocsAllowList`, `TestFallbackChainOrder`. Walks `internal/template/templates/` for offending patterns. | expert-backend | new file (~80 LOC) | none | 1 file (create) | GREEN at baseline — current 28 wrappers already clean. Test affirms current state, will fail if regression introduced. |
| T-RT007-05 | Create `internal/template/template_homedir_audit_test.go` (new) with `TestNoHomeDirInFallback`, `TestNoHomeDirInFallback_ContributorRegression`. Synthetic regression test uses `t.TempDir()` to construct violating template snippet. | expert-backend | new file (~60 LOC) | none | 1 file (create) | GREEN at baseline. Detects future contributor regression. |
| T-RT007-06 | Create `internal/template/renderer_passthrough_test.go` (new) with `TestHomeIsRegisteredInPassthroughTokens`. Affirms `claudeCodePassthroughTokens` slice contains `"$HOME"` (verifies REQ-006 invariant). | expert-backend | new file (~30 LOC) | none | 1 file (create) | GREEN at baseline (renderer.go:42 already has `"$HOME"`). |
| T-RT007-07 | Create `internal/migration/log_test.go` (new) with `TestLog_AppendsJSONLine`, `TestLog_PreservesPriorEntries`, `TestLog_HandlesConcurrentWrites`. | expert-backend | new file (~80 LOC) | none | 1 file (create) | RED — log helpers don't exist |
| T-RT007-08 | Create `internal/migration/migrations/m001_hardcoded_path_test.go` (new) with `TestM001_RewritesHardcodedLiteral`, `TestM001_NoOpWhenAlreadyClean`, `TestM001_PreservesExecutableBit`, `TestM001_PreservesOtherContent`, `TestM001_RollbackNotImplemented`, `TestM001_WindowsGitBash`. | expert-backend | new file (~150 LOC) | none | 1 file (create) | RED — m001 doesn't exist |
| T-RT007-09 | Create `internal/cli/migration_test.go` (new) with `TestMigrationStatus_Human`, `TestMigrationStatus_JSON`, `TestMigrationRun_AppliesPending`, `TestMigrationRun_NoPending`, `TestMigrationRollback_NoRollbackable`, `TestMigrationRollback_Succeeds`, `TestMigrationRollback_M001_Rejected`. | expert-backend | new file (~150 LOC) | none | 1 file (create) | RED — CLI doesn't exist |
| T-RT007-10 | Create `internal/cli/doctor_migration_test.go` (new) with `TestDoctorMigration_PrintsCurrentVersion`, `TestDoctorMigration_PrintsPendingCount`. | expert-backend | new file (~50 LOC) | none | 1 file (create) | RED — doctor extension doesn't exist |
| T-RT007-11 | Create `internal/runtime/gobin/resolver_test.go` (new) with `TestDetect_GOBINFirst`, `TestDetect_GOPATHSecond`, `TestDetect_HomeFallback`, `TestDetect_LastResort`. Uses `t.Setenv("GOBIN", ...)` to control fallback chain. | expert-backend | new file (~80 LOC) | none | 1 file (create) | RED — resolver doesn't exist (logic exists in two duplicates but not as helper) |
| T-RT007-12 | Create `internal/hook/session_start_migration_test.go` (new) with `TestSessionStart_InvokesMigrationRunner`, `TestSessionStart_MigrationFailure_SurfacesViaSystemMessage`, `TestSessionStart_DisabledViaSystemYaml`, `TestSessionStart_EnabledByDefault`. Uses stub MigrationRunner injection. | expert-backend | new file (~120 LOC) | none | 1 file (create) | RED — handler doesn't invoke runner yet |
| T-RT007-13 | Run `go test ./internal/migration/... ./internal/cli/... ./internal/template/... ./internal/runtime/... ./internal/hook/...` from worktree root and confirm RED state for all M1 new tests; existing tests still GREEN (regression sentinel). | manager-tdd | n/a (verification only) | T-RT007-01..12 | 0 files | RED gate verification |

**M1 priority: P0** — blocks all subsequent milestones.

T-RT007-01 through T-RT007-12 may execute in parallel (touch independent new files). T-RT007-13 must wait for all 12 to complete.

---

## M2: Path-fix consolidation (GREEN, part 1) — Priority P0

Goal: Single GoBinPath helper + 3 lint test GREEN. Satisfies AC-01, AC-02, AC-10.

| ID | Subject | Owner role | File:line target | Dependency | Touch points | TDD alignment |
|----|---------|-----------|-------------------|------------|--------------|---------------|
| T-RT007-14 | Create `internal/runtime/gobin/resolver.go` (new) with `Detect(homeDir string) string` function. Body merges logic from `internal/core/project/initializer.go:286-303` and `internal/cli/update.go:2515-2540` (both implementations identical). Order: `go env GOBIN` → `go env GOPATH/bin` → `$HOME/go/bin` → platform-aware last resort. | expert-backend | new file (~30 LOC) | T-RT007-13 | 1 file (create) | GREEN — resolver test (T-RT007-11) turns GREEN |
| T-RT007-15 | Refactor `internal/core/project/initializer.go:286-303`: replace `detectGoBinPath` body with `return gobin.Detect(homeDir)`. Add `import "github.com/modu-ai/moai-adk/internal/runtime/gobin"`. Function signature preserved. | expert-backend | `initializer.go` (edit, ~5 LOC change + 1 import) | T-RT007-14 | 1 file (edit) | REFACTOR — DRY consolidation |
| T-RT007-16 | Refactor `internal/cli/update.go:2515-2540`: replace `detectGoBinPathForUpdate` body with `return gobin.Detect(homeDir)`. Add gobin import. | expert-backend | `update.go` (edit, ~5 LOC change + 1 import) | T-RT007-14 | 1 file (edit) | REFACTOR — DRY consolidation |
| T-RT007-17 | Verify `internal/template/hardcoded_path_audit_test.go::*` 4 tests turn GREEN against current 28 wrappers + status_line. Add allow-list for the 2 known docs files (`SKILL.md:141`, `session-handoff.md:207, 216`) — these are docs examples, not active wrappers. | expert-backend | n/a (verify-only) | T-RT007-13 | 0 files | GREEN — AC-01 |
| T-RT007-18 | Verify `internal/template/template_homedir_audit_test.go::*` 2 tests turn GREEN. Synthetic regression test uses `t.TempDir()` to introduce violating snippet. | expert-backend | n/a (verify-only) | T-RT007-13 | 0 files | GREEN — AC-10 |
| T-RT007-19 | Verify `internal/template/renderer_passthrough_test.go::TestHomeIsRegisteredInPassthroughTokens` turns GREEN (already-passing affirm). | expert-backend | n/a (verify-only) | T-RT007-13 | 0 files | GREEN — REQ-006 affirm |

**M2 priority: P0** — blocks M3-M5. AC-01, AC-02, AC-10 GREEN. resolver test (T-RT007-11) also GREEN.

T-RT007-15, T-RT007-16 may run in parallel (touch independent files). T-RT007-17, T-RT007-18, T-RT007-19 are verify-only and require T-RT007-13 baseline.

---

## M3: Migration core (GREEN, part 2) — Priority P0

Goal: `internal/migration/{runner,registry,version,log}.go` core. Satisfies AC-04, AC-05, AC-11, AC-12, AC-14.

| ID | Subject | Owner role | File:line target | Dependency | Touch points | TDD alignment |
|----|---------|-----------|-------------------|------------|--------------|---------------|
| T-RT007-20 | Create `internal/migration/runner.go` (new): `Migration` struct (`Version int`, `Name string`, `Apply func(string) error`, `Rollback func(string) error`). `MigrationRunner` interface (`Apply`, `Status`, `Rollback`). `NewRunner(projectRoot string) MigrationRunner` factory. `runner` concrete struct holds `projectRoot`. | expert-backend | new file (~120 LOC) | T-RT007-13 | 1 file (create) | GREEN — runner.go skeleton + Apply body |
| T-RT007-21 | Create `internal/migration/registry.go` (new): package-level `var registry []Migration`. `init()` validates duplicate Versions → `panic("DuplicateMigrationVersion: ...")`. `Pending(current int) []Migration`, `Highest() int`. Returns sorted by Version asc. | expert-backend | new file (~40 LOC) | T-RT007-20 | 1 file (create) | GREEN — AC-11 + REQ-016 |
| T-RT007-22 | Create `internal/migration/version.go` (new): `readVersion(projectRoot string) (int, error)`, `writeVersion(projectRoot string, v int) error`. `writeVersion` uses `<path>.tmp` + `os.Rename` atomicity + advisory lock. RT-004 머지 시 `internal/session/lock.go` import; 미머지 시 inline `flock`/`LockFileEx` 임시 구현 (build-tag separated `version_unix.go` + `version_windows.go`). | expert-backend | new file (~80 LOC + 2 build-tag files ~50 LOC each = ~180 LOC) | T-RT007-20 | 3 files (create) | GREEN — AC-14 |
| T-RT007-23 | Create `internal/migration/log.go` (new): `LogEntry struct {Version int, Name string, StartedAt, CompletedAt time.Time, Result string, Details string}`. `Append(projectRoot string, entry LogEntry) error` — JSONL append-only to `.moai/logs/migrations.log`. `LastApplied(projectRoot string) (*LogEntry, error)` — reads file, returns last `Result: "success"` entry. | expert-backend | new file (~50 LOC) | T-RT007-20 | 1 file (create) | GREEN — REQ-014 |
| T-RT007-24 | Wire `Apply()` body in `runner.go`: walk `Pending(current)` → call each `Apply(projectRoot)` → on success: `writeVersion(current+1)` + `log.Append(success)`. On failure: `log.Append(failed)`, return error WITHOUT advancing version. | expert-backend | `runner.go` (extend Apply body) | T-RT007-20, T-RT007-21, T-RT007-22, T-RT007-23 | 1 file (edit) | GREEN — REQ-012, REQ-013 |
| T-RT007-25 | Wire `Status()` body in `runner.go`: `(current int, pending []int, last *LogEntry, err error)`. Reads version-file + registry + log.LastApplied. | expert-backend | `runner.go` (extend Status body) | T-RT007-24 | 1 file (edit) | GREEN — REQ-015 partial |
| T-RT007-26 | Wire `Rollback(version int)` body in `runner.go`: find Migration by Version in registry → if `Rollback == nil` return `ErrMigrationNotRollbackable` → else call `Rollback(projectRoot)` → on success: `writeVersion(version-1)` + `log.Append(rolled-back)`. | expert-backend | `runner.go` (extend Rollback body) | T-RT007-24 | 1 file (edit) | GREEN — REQ-024, REQ-042 |

**M3 priority: P0** — blocks M4 (m001 + session-start hook). AC-04, AC-05, AC-11, AC-12, AC-14 GREEN.

T-RT007-21, T-RT007-22, T-RT007-23 may run in parallel after T-RT007-20. T-RT007-24 depends on all 3.

---

## M4: m001 + session-start hook integration (GREEN, part 3) — Priority P0

Goal: m001 마이그레이션 + session-start 자동 적용 + system.yaml `migrations.disabled` 지원. Satisfies AC-03, AC-04, AC-06, AC-13, AC-15.

| ID | Subject | Owner role | File:line target | Dependency | Touch points | TDD alignment |
|----|---------|-----------|-------------------|------------|--------------|---------------|
| T-RT007-27 | Create `internal/migration/migrations/m001_hardcoded_path.go` (new): export `var M001_HardcodedPath = migration.Migration{Version: 1, Name: "remove_hardcoded_gobin_path", Apply: m001Apply, Rollback: nil}`. `m001Apply(projectRoot string) error` body per research.md §8.3: `filepath.Glob(.claude/hooks/moai/handle-*.sh)` → for each: read → check literal `/Users/goos/go/bin/moai` → if found `bytes.ReplaceAll` to `$HOME/go/bin/moai` → atomic write preserving mode. Log details: rewritten count or "already migrated". | expert-backend | new file (~50 LOC) | T-RT007-26 | 1 file (create) | GREEN — AC-03, AC-04, REQ-022, REQ-023 |
| T-RT007-28 | Register m001 in `internal/migration/registry.go`: import `migrations` package, append `migrations.M001_HardcodedPath` to `registry`. | expert-backend | `registry.go` (extend, ~3 LOC + 1 import) | T-RT007-27 | 1 file (edit) | GREEN — m001 reachable via runner |
| T-RT007-29 | Add `MigrationsConfig{Disabled bool}` to `internal/config/types.go`; add `Migrations MigrationsConfig` field to `Config` struct. Extend `internal/config/loader.go` to parse `system.yaml` `migrations:` section. Default: `Disabled: false`. | expert-backend | 2 files (edit, ~30 LOC) | T-RT007-26 | 2 files (edit) | GREEN — REQ-032 |
| T-RT007-30 | Extend `internal/hook/session_start.go::Handle` (around line 30, after project config load): insert MigrationRunner.Apply invocation. Pseudo-pattern (final body in run-phase): if `cfg.Migrations.Disabled` → log+skip; else `runner := migration.NewRunner(input.ProjectDir); applied, err := runner.Apply(ctx)`; on err: `slog.Warn` + (if RT-001 머지) `hookOut.SystemMessage = ...`; on success with applied > 0: `slog.Info`. NEVER block session. | expert-backend | `session_start.go` (extend Handle body, ~25 LOC + 1 import) | T-RT007-28, T-RT007-29 | 1 file (edit) | GREEN — REQ-020, REQ-021, AC-06, AC-13 |

**M4 priority: P0** — blocks M5 (CLI consumes runner). AC-03, AC-04, AC-06, AC-13, AC-15 GREEN.

T-RT007-27, T-RT007-29 may run in parallel after T-RT007-26. T-RT007-28 depends on T-RT007-27. T-RT007-30 depends on T-RT007-28 + T-RT007-29.

---

## M5: CLI surface + doctor + lint affirm + log + CHANGELOG + MX (GREEN, part 4 + Trackable) — Priority P1

Goal: User-facing surface + observability + trackability + MX tag insertions.

| ID | Subject | Owner role | File:line target | Dependency | Touch points | TDD alignment |
|----|---------|-----------|-------------------|------------|--------------|---------------|
| T-RT007-31 | Create `internal/cli/migration.go` (new): cobra `migrationCmd` group + 3 subcommands `migrationRunCmd`, `migrationStatusCmd` (with `--json` flag), `migrationRollbackCmd` (`<version>` arg). Wire to `rootCmd.AddCommand(migrationCmd)`. Each subcommand calls into `migration.NewRunner(cwd).{Apply,Status,Rollback}`. Format human-readable + JSON output for status. | expert-backend | new file (~150 LOC) | T-RT007-30 | 1 file (create) | GREEN — AC-07, AC-08, AC-09, AC-16 |
| T-RT007-32 | Create `internal/cli/doctor_migration.go` (new): `runDoctorMigration(cmd, args)` — calls `migration.NewRunner(cwd).Status()` and prints in doctor section format (Status: ok/warn/fail, Message). Extend `internal/cli/doctor.go` `runDoctor` switch (or `--check` flag handling) to add `case "migration": return runDoctorMigration(cmd, args)`. | expert-backend | 2 files (1 new ~40 LOC + 1 edit ~5 LOC) | T-RT007-31 | 2 files (create + edit) | GREEN — AC-07 (doctor side) |
| T-RT007-33 | Update `internal/template/templates/.claude/skills/moai/SKILL.md:141` and `.claude/rules/moai/workflow/session-handoff.md:207, 216` audit allow-list metadata: add `// audit-allow: docs-example` comment OR add to test allow-list slice in `hardcoded_path_audit_test.go`. Document why these strings exist (paste-ready examples, cross-reference paths). | manager-docs | 2-3 files (edit, audit metadata) | T-RT007-31 | 2-3 files (edit) | Trackable — audit allow-list documented |
| T-RT007-34 | Insert 7 MX tags per `plan.md` §6 verbatim across 7 distinct files: 3 ANCHOR (runner.go::MigrationRunner.Apply, registry.go::registry, gobin/resolver.go::Detect), 2 NOTE (version.go top, m001_hardcoded_path.go::Rollback), 2 WARN (session_start.go::Handle near Apply, runner.go::Apply near version-file write). At run-phase post-implementation: measure ANCHOR `fan_in=N` via `grep -rn "<symbol>(" internal/ pkg/ \| wc -l` and write the EXACT measured count (no `fan_in=N` literal). | manager-docs | 7 files (edit, MX tag insertion only) | T-RT007-33 | 7 files (edit) | Trackable — plan §6 |
| T-RT007-35 | Add CHANGELOG entry under `## [Unreleased] / ### Added` per `plan.md` §M5d text. | manager-docs | `CHANGELOG.md` (extend) | T-RT007-34 | 1 file (edit) | Trackable |
| T-RT007-36 | Run `make build` from worktree root to regenerate `internal/template/embedded.go`. Expected diff: minimal or empty (no `.claude/` template content changes; only audit test files added in `internal/template/`, which are not embedded). Verify diff. | manager-docs | `internal/template/embedded.go` (regenerated) | T-RT007-35 | 1 file (regenerated) | Build verification |
| T-RT007-37 | Run full `go test ./...` from worktree root. Verify ALL audit tests pass + 0 cascading failures (per `CLAUDE.local.md` §6 HARD rule). Run `go vet ./...` + `golangci-lint run` — zero warnings. | manager-tdd | n/a (verification only) | T-RT007-36 | 0 files | GREEN gate (final) |
| T-RT007-38 | Update `progress.md` with `run_complete_at: <timestamp>` and `run_status: implementation-complete` + populate `acs_passed: 16/16`, `tests_added: ~40`, `mx_tags_inserted: 7`. | manager-docs | `progress.md` (extend) | T-RT007-37 | 1 file (edit) | Trackable closure |
| T-RT007-39 | (Optional, deferred to manager-git) Create branch `feature/SPEC-V3R2-RT-007-migration-framework` from `plan/SPEC-V3R2-RT-007`, push, open PR with template `release-drafter` autolabeling. PR title format: `feat(migration): SPEC-V3R2-RT-007 versioned migration framework + hardcoded-path regression lock`. | manager-git | git only | T-RT007-38 | 0 files (git) | Trackable — PR open |
| T-RT007-40 | (After PR open) Update `progress.md` with `pr_number: <N>`. After merge: `merged_commit: <hash>`. | manager-git | `progress.md` (extend) | T-RT007-39 | 1 file (edit) | Trackable closure |

**M5 priority: P1** — completes the SPEC. AC-07, AC-08, AC-09, AC-16 GREEN.

T-RT007-31 and T-RT007-32 sequential (32 depends on 31). T-RT007-33 may run in parallel with T-RT007-31. T-RT007-34 must wait for all M5 source code edits to land (MX tag positions reference final code structure).

---

## Task summary by milestone

| Milestone | Task IDs | Total tasks | Priority | Owner role mix |
|-----------|----------|-------------|----------|----------------|
| M1 (RED) | T-RT007-01..13 | 13 | P0 | expert-backend (12) + manager-tdd (1 verification) |
| M2 (GREEN part 1 — DRY consolidation) | T-RT007-14..19 | 6 | P0 | expert-backend (3 source) + verify (3) |
| M3 (GREEN part 2 — migration core) | T-RT007-20..26 | 7 | P0 | expert-backend |
| M4 (GREEN part 3 — m001 + hook) | T-RT007-27..30 | 4 | P0 | expert-backend |
| M5 (GREEN part 4 + Trackable) | T-RT007-31..40 | 10 | P1 | expert-backend (2) + manager-docs (5) + manager-tdd (1 verification) + manager-git (2) |
| **TOTAL** | T-RT007-01..40 | **40 tasks** | — | — |

> NOTE: 40 tasks span the 5 milestones. M1 (RED) opens with 13 test-only tasks; M2-M4 (GREEN core) deliver the framework + path-fix lock in 17 tasks; M5 (final GREEN + REFACTOR + Trackable) closes with 10 tasks of CLI, audit allow-list, MX tags, CHANGELOG, build verify, and git/PR closure.

---

## Dependency graph (critical path)

```
T-RT007-01..12 (M1 tests, parallel)
   ↓
T-RT007-13 (M1 verification gate)
   ↓
T-RT007-14 (gobin.Detect helper) ← M2 start
   ↓
T-RT007-15, 16 (initializer.go, update.go refactor — parallel)
   ↓
T-RT007-17, 18, 19 (M2 verify lint tests — parallel)
   ↓
T-RT007-20 (runner.go skeleton) ← M3 start
   ↓
T-RT007-21, 22, 23 (registry, version, log — parallel)
   ↓
T-RT007-24 (Apply body wires all)
   ↓
T-RT007-25, 26 (Status, Rollback bodies — parallel)
   ↓
T-RT007-27 (m001) ← M4 start
   ↓
T-RT007-28 (registry m001 register) ┐
T-RT007-29 (config.Migrations field)│ — parallel
   ↓
T-RT007-30 (session-start hook integration)
   ↓
T-RT007-31 (CLI migration.go) ← M5 start
   ↓
T-RT007-32 (doctor migration extension) ┐
T-RT007-33 (audit allow-list metadata)  │ — parallel
   ↓
T-RT007-34 (MX tags)
   ↓
T-RT007-35 (CHANGELOG) → T-RT007-36 (make build) → T-RT007-37 (go test ./... + vet + lint)
   ↓
T-RT007-38 (progress.md closure)
   ↓
T-RT007-39 (manager-git: branch + PR)
   ↓
T-RT007-40 (progress.md PR/merge metadata)
```

Critical path: **18 sequential gates** from T-RT007-13 → T-RT007-40 (M1 → M5 closure).

---

## Cross-task constraints

[HARD] All file edits use absolute paths under the worktree root `/Users/goos/.moai/worktrees/moai-adk/SPEC-V3R2-RT-007` per CLAUDE.md §Worktree Isolation Rules.

[HARD] All tests use `t.TempDir()` per `CLAUDE.local.md` §6 — no test creates files in the project root.

[HARD] All filesystem operations use `filepath.Join` / `filepath.Abs`; no `filepath.Join(cwd, absPath)` patterns per `CLAUDE.local.md` §6.

[HARD] No new direct module dependencies. Only `golang.org/x/sys/{unix,windows}` (already in go.mod indirect) used for advisory lock.

[HARD] No `internal/template/templates/.claude/...` files modified by this SPEC (templates already clean per research.md §2.2). Only the 3 audit test files added in `internal/template/` (not embedded) + 2 doc allow-list metadata edits.

[HARD] Code comments in Korean (per `.moai/config/sections/language.yaml` `code_comments: ko`). Godoc and exported identifier docstrings remain English (industry standard).

[HARD] Commit messages in Korean (per `.moai/config/sections/language.yaml` `git_commit_messages: ko`).

[HARD] Subagent never invokes `AskUserQuestion`. SystemMessage is the migration's only user-facing surface; orchestrator (out-of-scope for this SPEC) handles user prompts.

[HARD] m001 atomic-write preserves executable bit (`0o755`). Test `TestM001_PreservesExecutableBit` verifies.

[HARD] RT-001 (HookResponse SystemMessage) and RT-006 (handler completeness) blocker dependencies verified at plan-audit gate. If unmerged at run-phase entry, plan-auditor blocks or run-phase agent stubs the SystemMessage write.

---

End of tasks.md.
