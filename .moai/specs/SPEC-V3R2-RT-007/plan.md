# SPEC-V3R2-RT-007 Implementation Plan

> Implementation plan for **Hardcoded Path Fix + Versioned Migration**.
> Companion to `spec.md` v0.1.0, `research.md` v0.1.0, `acceptance.md` v0.1.0, `tasks.md` v0.1.0.
> Authored against branch `plan/SPEC-V3R2-RT-007` at worktree `/Users/goos/.moai/worktrees/moai-adk/SPEC-V3R2-RT-007`. See §7 for cwd resolution rule.

## HISTORY

| Version | Date       | Author                        | Description                                                                                       |
|---------|------------|-------------------------------|---------------------------------------------------------------------------------------------------|
| 0.1.0   | 2026-05-10 | manager-spec (Plan workflow)  | Initial implementation plan per `.claude/skills/moai/workflows/plan.md` Phase 1B. Spec drift acknowledged: 28 wrappers (not 26), 0 hardcoded literals (not all 26). Plan reframes path-fix half as *affirm + harden + retroactive migration* rather than rewrite. |

---

## 1. Plan Overview

### 1.1 Goal restatement

두 개의 reinforcing half으로 구성된 SPEC:

- **Path-fix half (affirm + harden + retroactive)**: 28개 shell wrapper 템플릿은 **이미 깨끗** (research.md §2.2 검증, 0 hardcoded literals). 본 SPEC은 (a) 회귀 방지 CI lint 추가, (b) 두 군데 중복된 `detectGoBinPath` 로직을 단일 helper로 통합, (c) `$HOME` passthrough 토큰 등록을 affirm-only 테스트로 잠금, (d) v2.x 사용자에게 fix를 retroactively 전파하는 마이그레이션 1 제공.
- **Migration framework half (new construction)**: `internal/migration/` 패키지 신규 생성 — `Migration` struct, `MigrationRunner`, version-file guard, registry, session-start 자동 적용, `moai migration {run,status,rollback}` CLI, `moai doctor migration` doctor extension, structured log.

이로써 `spec.md` §1 Goal 두 줄을 모두 충족: P-H04 (hardcoded path 회귀 방지) + P-C06 (silent migration 패턴 부재) 동시 해결.

### 1.2 Implementation Methodology

`.moai/config/sections/quality.yaml`: `development_mode: tdd`. Run phase MUST follow RED → GREEN → REFACTOR per `.claude/rules/moai/workflow/spec-workflow.md` Phase Run. Run-phase agent: `manager-cycle` (cycle_type=tdd) per spec-workflow.md §Run Phase.

- **RED**: M1 에서 32 EARS REQ → 16 AC 매핑을 검증하는 ~22개 신규 테스트 작성. 마이그레이션 러너, version-file guard, idempotency, lock 경합, rollback, doctor 통합, CI lint 모두 RED 상태로 시작.
- **GREEN**: M2 (path-fix 통합) → M3 (migration core) → M4 (registry + m001 + session-start hook) → M5 (CLI + doctor + lint + log + CHANGELOG + MX) 순차 진입.
- **REFACTOR**: M2 에서 두 군데 중복된 `detectGoBinPath` 통합. M3 에서 `migrate_agency.go` 의 `migrationCheckpoint` 패턴 일부를 framework 측으로 일반화 (ownership 분리 유지: agency 명령은 그대로 두고 framework가 별도 cobra group에 거주).

### 1.3 Deliverables

| Deliverable | Path | REQ Coverage |
|-------------|------|--------------|
| 단일 GoBinPath 리졸버 helper | `internal/runtime/gobin/resolver.go` (new) — `Detect(homeDir string) string` | REQ-001 |
| `detectGoBinPath` (initializer) → resolver helper 호출로 교체 | `internal/core/project/initializer.go:286-303` | REQ-001 (DRY) |
| `detectGoBinPathForUpdate` (cli/update) → resolver helper 호출로 교체 | `internal/cli/update.go:2515-2540` | REQ-001 (DRY) |
| CI lint: 절대경로 회귀 방지 audit | `internal/template/hardcoded_path_audit_test.go` (new) | REQ-002, REQ-051 |
| CI lint: `.HomeDir` 폴백 금지 audit | `internal/template/template_homedir_audit_test.go` (new) | REQ-004, REQ-050 |
| affirm test: `$HOME` passthrough 등록 | `internal/template/renderer_passthrough_test.go` (new, ~30 LOC) | REQ-006 |
| `Migration` struct + `MigrationRunner` 인터페이스 | `internal/migration/runner.go` (new) | REQ-010, REQ-012 |
| `Registry` (compile-time static) | `internal/migration/registry.go` (new) | REQ-016, REQ-053 |
| `m001_hardcoded_path` 마이그레이션 | `internal/migration/migrations/m001_hardcoded_path.go` (new) | REQ-022, REQ-023 |
| version-file guard (atomic, advisory lock) | `internal/migration/version.go` (new) | REQ-013, REQ-031 |
| structured log writer | `internal/migration/log.go` (new) — JSONL appender | REQ-014 |
| session-start hook 마이그레이션 호출 | `internal/hook/session_start.go` (extend `Handle`) | REQ-020, REQ-021 |
| `moai migration` cobra group + 3 subcommands | `internal/cli/migration.go` (new) | REQ-015, REQ-040, REQ-041, REQ-024 |
| `moai doctor migration` doctor extension | `internal/cli/doctor.go` (extend `--check migration`) | REQ-015 |
| `system.yaml` `migrations.disabled` 키 지원 | `internal/config/types.go` (add `MigrationsConfig`) + 로더 확장 | REQ-032 |
| advisory lock 재사용 (RT-004 머지 후) | `internal/session/lock.go` import (또는 RT-004 미머지 시 임시 자체 구현) | REQ-031 |
| CHANGELOG entry | `CHANGELOG.md` Unreleased section | Trackable (TRUST 5) |
| MX tags per §6 | 7 files (per §6 below) | mx_plan |

`make build` 는 `.claude/`, `.moai/` 템플릿 파일이 변경되지 않으므로 `internal/template/embedded.go` 의 영향이 거의 없다. 그러나 audit lint 추가로 인해 `go test ./internal/template/...` 가 새 테스트를 발견해야 하므로 `go vet ./...` + `go test ./...` 통과는 필수.

### 1.4 Traceability Matrix (REQ → AC → Task)

Per plan-auditor PASS criterion #2 (every REQ maps to at least one AC and at least one Task):

| REQ ID | Category | Mapped AC(s) | Mapped Task(s) |
|--------|----------|--------------|----------------|
| REQ-V3R2-RT-007-001 | Ubiquitous (path) | (covered by detection round-trip via resolver tests) | T-RT007-11, T-RT007-14, T-RT007-15, T-RT007-16 |
| REQ-V3R2-RT-007-002 | Ubiquitous (path) | AC-01 | T-RT007-04, T-RT007-17 |
| REQ-V3R2-RT-007-003 | Ubiquitous (path) | AC-02 | T-RT007-04, T-RT007-17 |
| REQ-V3R2-RT-007-004 | Ubiquitous (path) | AC-10 | T-RT007-05, T-RT007-18 |
| REQ-V3R2-RT-007-005 | Ubiquitous (path) | AC-01 | T-RT007-36 (make build verification) |
| REQ-V3R2-RT-007-006 | Ubiquitous (path) | (passthrough affirm) | T-RT007-06, T-RT007-19 |
| REQ-V3R2-RT-007-010 | Ubiquitous (mig) | AC-08, AC-09 | T-RT007-20 |
| REQ-V3R2-RT-007-011 | Ubiquitous (mig) | AC-04 | T-RT007-01, T-RT007-08, T-RT007-27 |
| REQ-V3R2-RT-007-012 | Ubiquitous (mig) | AC-05, AC-12 | T-RT007-20, T-RT007-24 |
| REQ-V3R2-RT-007-013 | Ubiquitous (mig) | AC-05, AC-06 | T-RT007-22, T-RT007-24 |
| REQ-V3R2-RT-007-014 | Ubiquitous (mig) | AC-06 | T-RT007-23, T-RT007-24 |
| REQ-V3R2-RT-007-015 | Ubiquitous (mig) | AC-07 | T-RT007-25, T-RT007-31, T-RT007-32 |
| REQ-V3R2-RT-007-016 | Ubiquitous (mig) | AC-11 | T-RT007-21 |
| REQ-V3R2-RT-007-020 | Event-Driven | AC-05 | T-RT007-30 |
| REQ-V3R2-RT-007-021 | Event-Driven | AC-06 | T-RT007-24, T-RT007-30 |
| REQ-V3R2-RT-007-022 | Event-Driven | AC-03 | T-RT007-08, T-RT007-27 |
| REQ-V3R2-RT-007-023 | Event-Driven | AC-04 | T-RT007-08, T-RT007-27 |
| REQ-V3R2-RT-007-024 | Event-Driven | AC-08, AC-09 | T-RT007-26, T-RT007-31 |
| REQ-V3R2-RT-007-025 | Event-Driven | AC-01 | T-RT007-36 |
| REQ-V3R2-RT-007-030 | State-Driven | AC-05 | T-RT007-01, T-RT007-24 |
| REQ-V3R2-RT-007-031 | State-Driven | AC-14 | T-RT007-03, T-RT007-22 |
| REQ-V3R2-RT-007-032 | State-Driven | AC-13 | T-RT007-12, T-RT007-29, T-RT007-30 |
| REQ-V3R2-RT-007-040 | Optional | AC-16 | T-RT007-31 |
| REQ-V3R2-RT-007-041 | Optional | AC-07 (json edge) | T-RT007-31 |
| REQ-V3R2-RT-007-042 | Optional | AC-08 | T-RT007-26, T-RT007-31 |
| REQ-V3R2-RT-007-050 | Unwanted | AC-10 | T-RT007-05, T-RT007-18 |
| REQ-V3R2-RT-007-051 | Unwanted | AC-01 (lint side) | T-RT007-04, T-RT007-17 |
| REQ-V3R2-RT-007-052 | Unwanted | (manual: read-only fs) | T-RT007-08, T-RT007-27 (negative test) |
| REQ-V3R2-RT-007-053 | Unwanted | AC-11 | T-RT007-02, T-RT007-21 |
| REQ-V3R2-RT-007-054 | Unwanted | AC-12 | T-RT007-01, T-RT007-24 |
| REQ-V3R2-RT-007-060 | Complex | AC-15 | T-RT007-08, T-RT007-27 |
| REQ-V3R2-RT-007-061 | Complex | AC-14 | T-RT007-03, T-RT007-24 |

Coverage: **32 REQs mapped to 16 ACs and 40 tasks** (see tasks.md for full breakdown). 일부 REQ → multiple AC, 일부 task → multiple REQ. 모든 REQ 가 ≥ 1 AC + ≥ 1 task에 매핑.

---

## 2. Milestone Breakdown (M1-M5)

각 milestone은 **priority-ordered** (no time estimates per `.claude/rules/moai/core/agent-common-protocol.md` §Time Estimation HARD).

### M1: Test scaffolding (RED phase) — Priority P0

Owner role: `manager-cycle` cycle_type=tdd 또는 `expert-backend` 직접 실행.

Scope (12 신규 test files, ~22+ 테스트):

1. `internal/migration/runner_test.go` (new): `TestRunner_Apply_HappyPath`, `TestRunner_Apply_Idempotent`, `TestRunner_Apply_FreshInstall_AllInOrder`, `TestRunner_Apply_VersionAhead`, `TestRunner_Apply_FailureHaltsAdvance`, `TestRunner_Apply_PartialSuccess`, `TestRunner_Apply_CrashRecovery`.
2. `internal/migration/registry_test.go` (new): `TestRegistry_DuplicateVersion_Panics`, `TestRegistry_Pending`, `TestRegistry_HighestVersion`.
3. `internal/migration/version_test.go` (new): `TestVersionFile_RoundTrip`, `TestVersionFile_AtomicRename`, `TestVersionFile_AbsentMeansZero`, `TestVersionFile_AdvisoryLock_HighContention`.
4. `internal/migration/log_test.go` (new): `TestLog_AppendsJSONLine`, `TestLog_PreservesPriorEntries`, `TestLog_HandlesConcurrentWrites`.
5. `internal/migration/migrations/m001_hardcoded_path_test.go` (new): `TestM001_RewritesHardcodedLiteral`, `TestM001_NoOpWhenAlreadyClean`, `TestM001_PreservesExecutableBit`, `TestM001_PreservesOtherContent`, `TestM001_RollbackNotImplemented`, `TestM001_WindowsGitBash`.
6. `internal/cli/migration_test.go` (new): `TestMigrationStatus_Human`, `TestMigrationStatus_JSON`, `TestMigrationRun_AppliesPending`, `TestMigrationRollback_NoRollbackable`, `TestMigrationRollback_Succeeds`.
7. `internal/cli/doctor_migration_test.go` (new): `TestDoctorMigration_PrintsCurrentVersion`, `TestDoctorMigration_PrintsPendingCount`.
8. `internal/template/hardcoded_path_audit_test.go` (new): `TestNoHardcodedAbsolutePath_HookWrappers` (28 wrappers + status_line + others).
9. `internal/template/template_homedir_audit_test.go` (new): `TestNoHomeDirInFallback`.
10. `internal/template/renderer_passthrough_test.go` (new): `TestHomeIsRegisteredInPassthroughTokens`.
11. `internal/runtime/gobin/resolver_test.go` (new): `TestDetect_GOBINFirst`, `TestDetect_GOPATHSecond`, `TestDetect_HomeFallback`, `TestDetect_LastResort`.
12. `internal/hook/session_start_migration_test.go` (new) — invoke handler with stub MigrationRunner, verify Apply is called and SystemMessage propagates on failure.

Run `go test ./internal/migration/... ./internal/cli/... ./internal/template/... ./internal/runtime/... ./internal/hook/...` — confirm RED for new tests; existing tests still pass (regression baseline).

Verification gate before advancing to M2: 22+ new tests fail with documented sentinel messages.

[HARD] No implementation code in M1 outside of test files.

### M2: Path-fix consolidation (GREEN, part 1) — Priority P0

Scope:

1. `internal/runtime/gobin/resolver.go` (new) — `Detect(homeDir string) string` 함수. `go env GOBIN` → `go env GOPATH/bin` → `$HOME/go/bin` → platform-aware last resort. body는 `internal/core/project/initializer.go:286-303`의 verbatim 차용 + `internal/cli/update.go:2515-2540`과 동일 결과 보장.
2. `internal/core/project/initializer.go:286-303` — `detectGoBinPath` 함수 본문을 `gobin.Detect(homeDir)` 호출로 교체. function signature 보존.
3. `internal/cli/update.go:2515-2540` — 동일 처리: `detectGoBinPathForUpdate` 본문을 `gobin.Detect(homeDir)` 호출로 교체.
4. `internal/template/hardcoded_path_audit_test.go` — 28개 hook wrapper + `internal/template/templates/.moai/status_line.sh.tmpl` + docs allow-list 처리.
5. `internal/template/template_homedir_audit_test.go` — `.sh.tmpl` 파일에서 `{{posixPath .HomeDir}}` literal 매칭 0건 assert.
6. `internal/template/renderer_passthrough_test.go` — `claudeCodePassthroughTokens` slice에 `"$HOME"` 존재 assert.

Verification: `TestNoHardcodedAbsolutePath_HookWrappers`, `TestNoHomeDirInFallback`, `TestHomeIsRegisteredInPassthroughTokens`, `TestDetect_*` (4 cases) 모두 GREEN. AC-01, AC-02, AC-10 GREEN.

### M3: Migration core (GREEN, part 2) — Priority P0

Scope:

1. `internal/migration/runner.go` (new) — `Migration` struct, `MigrationRunner` interface, `NewRunner(projectRoot string)`.

   ```go
   // pseudo-Go (interface declaration only; bodies in run-phase)
   type Migration struct {
       Version  int
       Name     string
       Apply    func(projectRoot string) error
       Rollback func(projectRoot string) error  // optional, may be nil
   }

   type MigrationRunner interface {
       Apply(ctx context.Context) (applied []int, err error)
       Status() (current int, pending []int, lastApplied *LogEntry, err error)
       Rollback(version int) error
   }

   func NewRunner(projectRoot string) MigrationRunner { /* ... */ }
   ```

2. `internal/migration/version.go` (new) — version-file 읽기/쓰기 + advisory lock + atomic rename. `readVersion(projectRoot string) (int, error)` 부재면 0 반환. `writeVersion(projectRoot string, v int) error` write `.tmp` → `os.Rename`.

3. `internal/migration/log.go` (new) — JSONL appender. `LogEntry struct { Version int, Name string, StartedAt, CompletedAt time.Time, Result string, Details string }`. `Append(projectRoot, entry)`, `LastApplied(projectRoot)`.

4. `internal/migration/registry.go` (new) — compile-time registry. `var registry []Migration`. `init()` validates duplicate Versions → panic.

Verification: M3 tests GREEN.

### M4: m001 + session-start hook 통합 (GREEN, part 3) — Priority P0

Scope:

1. `internal/migration/migrations/m001_hardcoded_path.go` (new) — m001 마이그레이션 함수 본문 (research.md §8.3 참조).
2. `internal/migration/registry.go` — m001을 registry에 등록.
3. `internal/migration/runner.go::Apply` 본문 — `Pending(current)` walk → 각 Migration의 `Apply(projectRoot)` 호출 → 성공 시 `writeVersion(current+1)` → log append.
4. `internal/hook/session_start.go::Handle` 확장 (project config load 직후) — MigrationRunner.Apply 호출.
5. `migrationsDisabled(cfg)` helper — system.yaml `migrations.disabled: true` check.
6. `internal/config/types.go` — `MigrationsConfig{Disabled bool}` 추가; `Config.Migrations MigrationsConfig` 필드.
7. `internal/config/loader.go` — `system.yaml` `migrations:` 섹션 로딩.

Verification: AC-03, AC-04, AC-05, AC-06, AC-12, AC-13, AC-14, AC-15 GREEN.

### M5: CLI surface + doctor + lint affirm + log + CHANGELOG + MX (GREEN, part 4 + Trackable) — Priority P1

#### M5a: `moai migration` cobra group (REQ-015, REQ-040, REQ-041, REQ-024)

`internal/cli/migration.go` (new) — `moai migration {run, status [--json], rollback <version>}` cobra group + 3 subcommands.

#### M5b: `moai doctor --check migration` 통합 (REQ-015)

`internal/cli/doctor.go` 확장 — `--check migration` 옵션 추가. `runDoctorMigration(cmd, args)` — `migration.NewRunner(cwd).Status()` 호출 후 doctor format 출력.

#### M5c: AskUserQuestion-routing for SystemMessage

session-start handler 의 `hookOut.SystemMessage` 는 RT-001 의 HookResponse 스키마를 따라 orchestrator로 전달. RT-007의 코드는 SystemMessage 필드 작성까지만; orchestrator-side AskUserQuestion 처리는 out-of-scope.

#### M5d: CHANGELOG + MX tags + final verification

1. CHANGELOG entry under `## [Unreleased] / ### Added`:

   ```
   ### Added
   - SPEC-V3R2-RT-007: Versioned migration framework + hardcoded-path regression lock.
     New `internal/migration/` package (Migration, MigrationRunner, Registry, version-file
     guard, JSONL log). New `moai migration {run,status,rollback}` CLI subcommands. New
     `moai doctor --check migration`. Session-start hook silently applies pending
     migrations (REQ-020). Migration 1 (`m001_hardcoded_path`) retroactively rewrites
     `/Users/goos/go/bin/moai` literals to `$HOME/go/bin/moai` in v2.x user wrappers
     (idempotent on already-clean projects). Two CI lints prevent absolute-path and
     `.HomeDir` regressions in shell templates.
   ```

2. MX tags per §6 (7 tags across 7 files).

3. `make build` from worktree root.

4. `go test ./...` from worktree root — verify ALL tests pass.

5. `go vet ./...` + `golangci-lint run` — zero warnings.

6. Update `progress.md`.

[HARD] No new documents in `.moai/specs/` or `.moai/reports/` during M5.

---

## 3. File:line Anchors (concrete edit targets)

### 3.1 To-be-modified (existing files, verified to exist via Grep/Read 2026-05-10)

| File | Anchor | Edit type | Reason |
|------|--------|-----------|--------|
| `internal/core/project/initializer.go:286-303` | `detectGoBinPath` 함수 | Replace body with `return gobin.Detect(homeDir)` | M2 / REQ-001 (DRY) |
| `internal/cli/update.go:2515-2540` | `detectGoBinPathForUpdate` 함수 | Replace body with `return gobin.Detect(homeDir)` | M2 / REQ-001 (DRY) |
| `internal/hook/session_start.go:23-60` | `sessionStartHandler.Handle` | Insert MigrationRunner.Apply call after config load | M4 / REQ-020, REQ-021 |
| `internal/cli/doctor.go:43-58` | `doctorCmd` flag handling | Add `--check migration` branch | M5b / REQ-015 |
| `internal/config/types.go` | top-level `Config` struct | Add `Migrations MigrationsConfig` field with `Disabled bool` | M4 / REQ-032 |
| `internal/config/loader.go` | yaml loader | Read `system.yaml` `migrations:` section | M4 / REQ-032 |
| `internal/template/renderer.go:39-50` | `claudeCodePassthroughTokens` | (NO CHANGE — affirm-only via M2 test) | M2 / REQ-006 |
| `CHANGELOG.md` | `## [Unreleased]` section | Add Added entry per §M5d | M5d / Trackable |

### 3.2 To-be-created (new files)

| File | Reason | LOC estimate |
|------|--------|--------------|
| `internal/runtime/gobin/resolver.go` | `Detect` helper consolidating two duplicate impls | ~30 |
| `internal/runtime/gobin/resolver_test.go` | 4 test cases covering fallback chain | ~80 |
| `internal/migration/runner.go` | `Migration` struct, `MigrationRunner` interface, `NewRunner`, `Apply` body | ~120 |
| `internal/migration/runner_test.go` | 7 test cases | ~150 |
| `internal/migration/registry.go` | Compile-time static registry + `init()` duplicate panic | ~40 |
| `internal/migration/registry_test.go` | 3 test cases | ~80 |
| `internal/migration/version.go` | version-file read/write + advisory lock + atomic rename | ~80 |
| `internal/migration/version_unix.go` | `//go:build unix` flock helper (planned for M3) | ~50 |
| `internal/migration/version_windows.go` | `//go:build windows` LockFileEx helper (planned for M3) | ~60 |
| `internal/migration/version_test.go` | 4 test cases incl. high-contention | ~120 |
| `internal/migration/log.go` | JSONL log appender + LastApplied reader | ~50 |
| `internal/migration/log_test.go` | 3 test cases | ~80 |
| `internal/migration/migrations/m001_hardcoded_path.go` | m001 Apply 본문 (research §8.3) | ~50 |
| `internal/migration/migrations/m001_hardcoded_path_test.go` | 6 test cases | ~150 |
| `internal/cli/migration.go` | cobra group + 3 subcommands | ~150 |
| `internal/cli/migration_test.go` | 5 test cases | ~150 |
| `internal/cli/doctor_migration.go` | `--check migration` 구현 | ~40 |
| `internal/cli/doctor_migration_test.go` | 2 test cases | ~50 |
| `internal/template/hardcoded_path_audit_test.go` | 28 wrappers + status_line audit | ~80 |
| `internal/template/template_homedir_audit_test.go` | `{{posixPath .HomeDir}}` audit | ~60 |
| `internal/template/renderer_passthrough_test.go` | `$HOME` passthrough affirm | ~30 |
| `internal/hook/session_start_migration_test.go` | hook ↔ MigrationRunner integration test | ~120 |

Total new: ~1,800 LOC (planned). Total modified (existing): ~30 LOC. Net: ~1,830 LOC across 22 new files + 6 modified files.

### 3.3 NOT to be touched (preserved by reference)

- 28개 `*.sh.tmpl` 파일 (`internal/template/templates/.claude/hooks/moai/handle-*.sh.tmpl`) — 이미 깨끗 (research.md §2.2 검증). 본 SPEC은 추가 lint만 추가하며 wrappers 자체는 미변경.
- `internal/template/templates/.moai/status_line.sh.tmpl` — fallback chain 사용 패턴 동일, 미변경 (회귀 lint만 적용).
- `.moai/state/` 디렉터리 contents at runtime — runtime artifacts. `.gitkeep` 만 존재.
- `internal/cli/migrate_agency.go` — 기존 `moai migrate agency` 일회성 명령 그대로 보존; 본 SPEC은 별도 cobra group `moai migration` 신설.
- `.claude/rules/moai/core/agent-common-protocol.md` — load-bearing 규칙 verbatim 유지.
- `internal/template/templates/.claude/...` — skill, command, agent 파일 모두 미변경.

### 3.4 Reference citations (verified file:line anchors)

다음 anchor들은 Grep/Read로 검증되었으며, plan-audit가 verify한다:

1. `spec.md:50-66` — In-scope items (plan + migration halves).
2. `spec.md:122-175` — 32 EARS REQs (verified by `grep -c "^- REQ-V3R2-RT-007-" spec.md` → 32).
3. `spec.md:179-194` — 16 ACs (verified by `grep -c "^- AC-V3R2-RT-007-" spec.md` → 16).
4. `internal/template/templates/.claude/hooks/moai/handle-session-start.sh.tmpl:13-34` — Current fallback chain (verified 0 hardcoded literals 2026-05-10).
5. `internal/template/templates/.claude/hooks/moai/handle-session-end.sh.tmpl:13-34` — Same chain pattern.
6. `internal/template/renderer.go:39-50` — `claudeCodePassthroughTokens` (`$HOME` already at index 3).
7. `internal/template/context.go:51` — `TemplateContext.GoBinPath` field.
8. `internal/template/context.go:209-211` — `WithGoBinPath` setter.
9. `internal/core/project/initializer.go:257` — `goBinPath := detectGoBinPath(homeDir)` (init path).
10. `internal/core/project/initializer.go:286-303` — `detectGoBinPath` definition.
11. `internal/cli/update.go:528, 560` — `detectGoBinPathForUpdate` call sites (update path).
12. `internal/cli/update.go:2515-2540` — `detectGoBinPathForUpdate` definition (duplicate — to be unified).
13. `internal/cli/migrate_agency.go:573-607` — Existing `migrateCmd` cobra wiring (preserved unchanged).
14. `internal/cli/migrate_agency.go:1-700+` — Reference for migration framework pattern (one-shot, preserved).
15. `internal/cli/doctor.go:43-58` — `doctorCmd` definition (extension target).
16. `internal/hook/session_start.go:23-60` — `sessionStartHandler.Handle` (extension target).
17. `internal/template/renderer_test.go:359-370` — Existing 'windows_gobinpath_converted_in_shell' test (regression baseline).
18. `internal/template/templates/.claude/skills/moai/SKILL.md:141` — Doc-level path reference (audit allow-list candidate).
19. `internal/template/templates/.claude/rules/moai/workflow/session-handoff.md:207, 216` — Doc-level path examples (audit allow-list candidate).
20. `.claude/rules/moai/core/agent-common-protocol.md` §User Interaction Boundary — Subagent ↔ orchestrator routing rule.
21. `.claude/rules/moai/workflow/spec-workflow.md` Phase 0.5 — Plan Audit Gate.
22. `CLAUDE.local.md:§14` — `.HomeDir` 금지 / `$HOME` 사용 규칙.
23. `CLAUDE.local.md:§2` — 보호 디렉터리 (`.moai/state/`, `.moai/logs/` 포함).
24. `CLAUDE.local.md:§6` — Test isolation (`t.TempDir()`).

Total: **24 distinct file:line anchors** (exceeds the §Hard-Constraints minimum of 10 for plan.md).

---

## 4. Technology Stack Constraints

Per `spec.md` §7 Constraints, **minimal new technology**:

- Go 1.22+ (already required by `go.mod`).
- `golang.org/x/sys/unix` (advisory lock for version-file) — already in indirect deps.
- `golang.org/x/sys/windows` (advisory lock for version-file) — already in indirect deps.
- No new external dependencies.

Reuse opportunities:

- RT-004 의 `internal/session/lock.go` (`fileLock` interface) — RT-004 머지 후 import; 미머지 시 `internal/migration/version_unix.go` + `version_windows.go` 임시 자체 구현 (RT-004 머지 시 swap).
- RT-004 의 `writeAtomic` helper — 미export 시 `internal/migration/version.go`에 자체 보유.

새로 등장하는 surfaces:

- 22 new Go files under `internal/migration/`, `internal/runtime/gobin/`, `internal/cli/`, `internal/template/`, `internal/hook/`.
- 6 modified existing files (per §3.1).
- 1 new cobra group (`moai migration`) + 3 subcommands.
- 1 new doctor `--check migration` branch.
- 1 new session-start hook integration point.
- 1 new system.yaml section (`migrations:`).

**No new directory structures** under user project — `.moai/state/`, `.moai/logs/` 모두 보호 디렉터리 (CLAUDE.local.md §2).

---

## 5. Risk Analysis & Mitigations

`spec.md` §8 risks를 file-path mitigation으로 확장.

| Risk | Probability | Impact | Mitigation Anchor |
|------|-------------|--------|-------------------|
| spec.md (drafted 2026-04-23) 의 "26 wrappers with hardcoded path" 명제가 invalid | H | M | research.md §2.2 명시. spec.md v0.1.0 본문이 본 plan-phase에서 이미 코드 현실에 맞게 보정됨. |
| RT-001 (HookResponse SystemMessage) 미머지 상태에서 RT-007 run phase 진입 | M | M | session-start handler가 RT-001 미머지 시 `slog.Warn` 임시 사용; RT-001 머지 후 single-line swap to `hookOut.SystemMessage = ...`. plan-audit gate에서 RT-001 PR 상태 verify. |
| RT-006 (handler completeness) 미머지 상태에서 핸들러 변경이 RT-006와 충돌 | M | M | RT-006 머지 후 RT-007 run-phase 진입 권장. RT-006 미머지 시 RT-007 변경을 stash/rebase로 합류. |
| `go env GOBIN`이 일부 머신에서 빈 문자열 반환 (GOPATH로 fall-through) | L | L | `gobin.Detect` 다단 fallback (REQ-001). 4-단 검증. |
| Migration 1 이 사용자 hand-edit한 wrapper 변경 덮어씀 | L | H | `bytes.ReplaceAll`로 *literal exact match* 만 replace. AC-04 (no-op when clean) regression baseline. |
| Concurrent session-start 두 개가 version-file race | L | M | advisory lock + 3-retry / 10ms backoff (RT-004 lock primitive 재사용). atomic rename으로 partial state 차단. Test: `TestVersionFile_AdvisoryLock_HighContention`. |
| Future m002+ 마이그레이션의 idempotency bug가 version 진행 후 작업 누락 | M | M | Test harness가 모든 마이그레이션을 *2회 실행* 비교. CI gate에서 `TestRunner_Apply_Idempotent`. m001은 자체로 idempotent (literal absence check). |
| Rollback이 CRITICAL bug-fix 마이그레이션 재도입 | H | H | m001은 `Rollback: nil` (declared non-rollback-able per REQ-024). `moai migration rollback 1` → `MigrationNotRollbackable` 명시 에러. |
| 사용자가 `moai migrate agency` 와 `moai migration run` 혼동 | M | L | `moai migration --help` 첫 줄에 cross-reference 포함. |
| Windows shell wrapper의 executable bit 손실 | M | L | atomic rename 시 `os.Stat`으로 기존 mode 읽고 새 파일에 적용. Test: `TestM001_PreservesExecutableBit`. |
| `migrations.disabled: true` 사용자가 critical fix 누락 | M | M | `moai doctor --check migration` 에서 disabled 상태 + pending count 표시. Release notes에 명시. |
| Cobra command group naming clash (`migrate` vs `migration`) | L | L | 두 그룹의 분리된 Use 문자열 ("migrate" vs "migration") cobra 단순 disambiguation. `--help` cross-reference. |
| `migrate agency` 구현 패턴 (`migrationCheckpoint` struct)이 framework로 새도 안 된 채 leak | L | L | 본 SPEC은 별도 패키지 (`internal/migration/`) 생성; `migrate_agency.go` 의 internal types 미사용. |
| `internal/template/embedded.go` 변경 없는데 `make build`가 실패 | L | L | `make build`는 idempotent — embed regenerator는 컨텐트 일치 시 no-op. plan-audit가 commit 전 `make build` 통과 verify. |
| `migrations.log` JSONL 파일이 무한히 커짐 | L | L | retention은 본 SPEC scope 외 — 별도 마이그레이션이 retention 정책 도입 가능. 사용자당 1-5건 수준. |

---

## 6. mx_plan — @MX Tag Strategy

Per `.claude/rules/moai/workflow/mx-tag-protocol.md` and `.claude/skills/moai/workflows/plan.md` mx_plan MANDATORY rule.

### 6.1 @MX:ANCHOR targets (high fan_in / contract enforcers)

fan_in measurement methodology: `grep -rn "<symbol>(" internal/ pkg/ | wc -l`. Plan-time projection used; final fan_in measured at run-phase post-implementation.

| Target file:line | Tag content (fan_in MEASURED at implementation, NOT a literal `N`) | Rationale |
|------------------|-------------|-----------|
| `internal/migration/runner.go::MigrationRunner.Apply()` | `@MX:ANCHOR fan_in=<measured> - SPEC-V3R2-RT-007 REQ-012, REQ-020 entry point. Called from session-start hook (silent path) and 'moai migration run' CLI (manual path). Idempotency contract + version-file atomic update + log entry must hold for both call sites.` Plan-time projection: 2 callers (hook + CLI). | Two upstream callers. Apply 본문 변경은 양쪽 영향. |
| `internal/migration/registry.go::registry` (package-level var) | `@MX:ANCHOR fan_in=<measured> - SPEC-V3R2-RT-007 REQ-016 compile-time static registry. Read by Pending(), Highest(), and init()-time DuplicateMigrationVersion check. Future migration additions (m002+) must register here; runtime modification forbidden.` Plan-time projection: 3 reads + 1 init validation. | 단일 진실 소스. 향후 EXT-004 / MIG-001 SPEC들이 registry에 append. |
| `internal/runtime/gobin/resolver.go::Detect()` | `@MX:ANCHOR fan_in=<measured> - SPEC-V3R2-RT-007 REQ-001 GoBinPath single source of truth. Replaces duplicate logic at internal/core/project/initializer.go:286 (init path) and internal/cli/update.go:2515 (update path). Adding a new caller without aligning fallback chain breaks REQ-002 absolute-path-free guarantee.` Plan-time projection: 2 callers (init + update). | 두 upstream caller. 폴백 체인 일관성 보장. |

### 6.2 @MX:NOTE targets (intent / context delivery)

| Target file:line | Tag content | Rationale |
|------------------|-------------|-----------|
| `internal/migration/version.go` (top of file) | `@MX:NOTE - SPEC-V3R2-RT-007 .moai/state/migration-version is the single source of truth for "what migrations have been applied". Atomic rename + advisory lock guard. Version absent = 0 (fresh install or v2.x first encounter). Never written by any code outside this package.` | 다른 패키지가 직접 version-file 쓰는 것 차단. |
| `internal/migration/migrations/m001_hardcoded_path.go::Rollback` (set to nil) | `@MX:NOTE - SPEC-V3R2-RT-007 REQ-024 m001 is intentionally NON-rollback-able. Rolling back reintroduces the CRITICAL absolute-path literal that breaks every user except 'goos'. CRITICAL-bug-fix migrations SHOULD declare Rollback: nil; rollback callers receive MigrationNotRollbackable.` | 사용자가 rollback 시도 시 의도된 거부. |

### 6.3 @MX:WARN targets (danger zones)

| Target file:line | Tag content | Rationale |
|------------------|-------------|-----------|
| `internal/hook/session_start.go::Handle` near MigrationRunner.Apply call | `@MX:WARN @MX:REASON - SPEC-V3R2-RT-007 REQ-020 silent migration apply at session start. Errors must NOT block session (REQ-021). Surface via SystemMessage but allow handler to return success. Bypassing this preserves migration-version-file unchanged → next session retries. NEVER let migration error abort session-start handler.` | 가장 회귀 위험 큰 지점. |
| `internal/migration/runner.go::Apply` near version-file update | `@MX:WARN @MX:REASON - SPEC-V3R2-RT-007 REQ-013, REQ-021 ATOMICITY CONTRACT. version-file update must occur AFTER successful migration apply, never before. Reordering causes "version advanced but work incomplete" — same idempotency-bug class that REQ-011 prevents at the apply layer.` | 미래 contributor가 "효율" 핑계로 version-file을 미리 쓸 수 있음. |

### 6.4 @MX:TODO targets (intentionally NONE for this SPEC)

본 SPEC은 완결된 마이그레이션 framework + path-fix 회귀 락을 산출. `@MX:TODO` 마커 없음 — 모든 작업이 M1-M5 안에서 GREEN 도달.

### 6.5 MX tag count summary

- @MX:ANCHOR: 3 targets
- @MX:NOTE: 2 targets
- @MX:WARN: 2 targets
- @MX:TODO: 0 targets
- **Total**: 7 MX tag insertions planned across 7 distinct files

---

## 7. Worktree Mode Discipline

[HARD] All run-phase work for SPEC-V3R2-RT-007 executes in:

```
/Users/goos/.moai/worktrees/moai-adk/SPEC-V3R2-RT-007
```

Branch: `plan/SPEC-V3R2-RT-007` (current). Run-phase에서는 spec-workflow.md updated convention 에 따라 Step 2: 별도 fresh worktree + `feat/SPEC-V3R2-RT-007` 브랜치를 `moai worktree new SPEC-V3R2-RT-007 --base origin/main` 으로 생성 후 진입.

[HARD] Worktree mode is in effect. All Read/Write/Edit tool invocations use absolute paths under the worktree root.

[HARD] `make build` and `go test ./...` execute from the worktree root.

[HARD] Run-phase agent MUST verify base commit `496595c3f` (HEAD of `origin/main` at plan-time) is the ancestor; rebase onto latest `origin/main` only via `manager-git`.

---

## 8. Plan-Audit-Ready Checklist

These criteria are checked by `plan-auditor` at `/moai run` Phase 0.5 (Plan Audit Gate per `spec-workflow.md` Phase 0.5).

각 항목은 **measured evidence** 를 cite (per Critical Lessons #5 in run prompt).

- [x] **C1: Frontmatter v0.1.0 schema** — `spec.md:1-23` frontmatter has 19 fields (id/title/version/status/created/updated/author/priority/phase/module/dependencies/bc_id/related_principle/related_pattern/related_problem/related_theme/breaking/lifecycle/tags). Verified by `head -23 spec.md`.
- [x] **C2: HISTORY entry for v0.1.0** — `spec.md:30-32` HISTORY table has v0.1.0 row with description.
- [x] **C3: 32 EARS REQs across 7 categories** — measured by `grep -E "^- REQ-V3R2-RT-007-[0-9]+:" spec.md | wc -l` → **32**. Categories: Ubiquitous (path) 6 + Ubiquitous (mig) 7 + Event-Driven 6 + State-Driven 3 + Optional 3 + Unwanted 5 + Complex 2 = 32. Verified.
- [x] **C4: 16 ACs all map to REQs (100% coverage)** — measured by `grep -E "^- AC-V3R2-RT-007-[0-9]+:" spec.md | wc -l` → **16**. Each AC explicitly cites the REQ(s) it maps to. plan §1.4 confirms 32 REQ → 16 AC → 40 task mapping.
- [x] **C5: BC scope clarity** — `spec.md:21` (`breaking: true`) + `bc_id: [BC-V3R2-008]` + spec.md §1 (master §8 BC-V3R2-008 commits to AUTO migration). issue-body.md "Breaking Change — BC-V3R2-008" section provides the user-impact matrix.
- [x] **C6: File:line anchors ≥10** — research.md §10 cites 29 anchors; this plan §3.4 cites 24 anchors. Each anchor verified via Grep/Read in plan-time.
- [x] **C7: Exclusions section present** — `spec.md:67-73` Out-of-scope (5 entries explicitly mapped to other SPECs: MIG-001, RT-006, MIG-003, RT-005, CON-001).
- [x] **C8: TDD methodology declared** — this plan §1.2 + `.moai/config/sections/quality.yaml` `development_mode: tdd` (verified by `head -2 quality.yaml` → `development_mode: tdd`).
- [x] **C9: mx_plan section** — this plan §6 with 7 MX tag insertions across 4 categories. fan_in values are *plan-time projections*; final fan_in measured at run-phase post-implementation via `grep -rn "<symbol>(" internal/ \| wc -l` and the EXACT integer count is written (no `fan_in=N` literal).
- [x] **C10: Risk table with mitigations** — `spec.md:209-219` (9 risks) + this plan §5 (15 risks, file-anchored mitigations).
- [x] **C11: Worktree mode path discipline** — this plan §7 (4 HARD rules, worktree path absolute, base commit `496595c3f` anchored).
- [x] **C12: No implementation code in plan documents** — verified self-check: this plan, research.md, acceptance.md, tasks.md contain only natural-language descriptions, regex patterns, file paths, code-block templates (illustrative pseudo-Go for interface declarations). No executable Go function bodies as final implementation.
- [x] **C13: Acceptance.md G/W/T format with edge cases** — verified in acceptance.md §1-16 (Given/When/Then for all 16 ACs + happy + edge cases + test mapping).
- [x] **C14: tasks.md owner roles aligned with TDD methodology** — verified in tasks.md §M1-M5 (manager-cycle / expert-backend / manager-docs / manager-git assignments).
- [x] **C15: Cross-SPEC consistency** — blocked-by dependencies verified: SPEC-V3R2-CON-001, SPEC-V3R2-RT-001, SPEC-V3R2-RT-006 (status check at plan-audit gate). RT-007 blocks EXT-004, MIG-001, MIG-002, MIG-003 per `spec.md` §9.2.

All 15 criteria PASS → plan is **audit-ready**.

---

## 9. Implementation Order Summary

Run-phase agent executes in this order (P0 first, dependencies resolved):

1. **M1 (P0)**: ~22 새 테스트 추가 across 12 신규 test files. Confirm RED for all; existing tests still GREEN.
2. **M2 (P0)**: `gobin.Detect` helper 작성 + 두 호출처 swap + 3 lint test (path / homedir / passthrough). Confirm AC-01, AC-02, AC-10 GREEN.
3. **M3 (P0)**: `internal/migration/{runner,registry,version,log}.go` core. Confirm registry+version+log tests GREEN.
4. **M4 (P0)**: `m001_hardcoded_path.go` + session-start hook 통합 + system.yaml `migrations.disabled` config. Confirm AC-03, AC-04, AC-05, AC-06, AC-12, AC-13, AC-14, AC-15 GREEN.
5. **M5 (P1)**: `moai migration {run,status,rollback}` cobra group + `moai doctor --check migration` + CHANGELOG + 7 MX tags + final `make build` + `go test ./...` + `go vet` + `golangci-lint`. Update `progress.md` with `run_complete_at` and `run_status: implementation-complete`.

Total milestones: 5. Total file edits (existing): ~6. Total file creations (new): 22. Total CHANGELOG entries: 1. Total MX tag insertions: 7.

Critical path: M1 → M2 → M3 → M4 → M5 (sequential gates).

---

End of plan.md.
