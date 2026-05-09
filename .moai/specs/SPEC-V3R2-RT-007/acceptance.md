# SPEC-V3R2-RT-007 Acceptance Criteria — Given/When/Then

> Detailed acceptance scenarios for each AC declared in `spec.md` §6.
> Companion to `spec.md` v0.1.0, `research.md` v0.1.0, `plan.md` v0.1.0, `tasks.md` v0.1.0.

## HISTORY

| Version | Date       | Author                        | Description                                                            |
|---------|------------|-------------------------------|------------------------------------------------------------------------|
| 0.1.0   | 2026-05-10 | manager-spec (Plan workflow)  | Initial G/W/T conversion of 16 ACs (AC-V3R2-RT-007-01 through -16)     |

---

## Scope

본 문서는 `spec.md` §6의 16개 AC를 Given/When/Then 형식으로 변환하며, happy-path + edge-case + test-mapping 표기를 포함한다.

Notation:
- **Test mapping** identifies which Go test function (or manual verification step) covers the AC.
- **Sentinel** is the literal error string the test expects on the negative path.

---

## AC-V3R2-RT-007-01 — Generated wrappers contain no absolute user paths

Maps to: REQ-V3R2-RT-007-002, REQ-V3R2-RT-007-025, REQ-V3R2-RT-007-051.

### Happy path

- **Given** `make build` 실행 후 `internal/template/embedded.go` 가 regenerate된 상태
- **And** 28개 hook wrapper 템플릿 + `internal/template/templates/.moai/status_line.sh.tmpl` 모두 embed
- **When** `internal/template/hardcoded_path_audit_test.go::TestNoHardcodedAbsolutePath_HookWrappers` 가 모든 `*.sh.tmpl`에서 정규식 `/Users/[^/]+/go/bin/moai` + `/home/[^/]+/go/bin/moai` 매치 검사
- **Then** 매치 0건
- **And** 테스트 PASS

### Edge case — docs reference allow-list

- **Given** `internal/template/templates/.claude/skills/moai/SKILL.md:141` 에 docs example로 `/Users/goos/MoAI/...` 경로가 있음
- **And** `internal/template/templates/.claude/rules/moai/workflow/session-handoff.md:207, 216` 에 paste-ready 예시로 `/Users/goos/.moai/worktrees/...` 경로가 있음
- **When** audit test가 실행될 때 docs 파일은 allow-list로 분류
- **Then** docs 파일의 경로 매치는 fail 트리거하지 않음
- **And** lint sentinel은 `*.sh.tmpl` 파일에 한정

### Edge case — 미래 contributor의 회귀 시도

- **Given** 누군가 `internal/template/templates/.claude/hooks/moai/handle-XXX.sh.tmpl` 에 `/Users/X/go/bin/moai` 리터럴 추가
- **When** CI가 audit test 실행
- **Then** 테스트 FAIL with sentinel `HARDCODED_ABSOLUTE_PATH: <file> line <N> contains absolute user path; use $HOME or {{posixPath .GoBinPath}} instead. See SPEC-V3R2-RT-007 REQ-002.`

### Test mapping

- `internal/template/hardcoded_path_audit_test.go::TestNoHardcodedAbsolutePath_HookWrappers` (new, M2)
- `internal/template/hardcoded_path_audit_test.go::TestNoHardcodedAbsolutePath_StatusLine` (new, M2)
- `internal/template/hardcoded_path_audit_test.go::TestNoHardcodedAbsolutePath_DocsAllowList` (new, M2)

---

## AC-V3R2-RT-007-02 — Fallback chain order verified in generated wrappers

Maps to: REQ-V3R2-RT-007-003.

### Happy path

- **Given** `make build` 후 모든 hook wrapper 템플릿이 동일한 fallback chain 패턴 따름
- **And** rendered output 검사 (template 변수 치환된 .sh 파일)
- **When** `grep "if command -v moai\|if \[ -f"` 로 fallback 라인 추출
- **Then** 라인 순서가 `command -v moai → {{posixPath .GoBinPath}}/moai → $HOME/go/bin/moai → $HOME/.local/bin/moai → exit 0`
- **And** 28개 wrapper 모두 일관

### Edge case — `.HomeDir` 등장 금지

- **Given** wrapper 템플릿 내부
- **When** `internal/template/template_homedir_audit_test.go::TestNoHomeDirInFallback` 가 `{{posixPath .HomeDir}}` 정규식 매치
- **Then** 매치 0건 (REQ-004 금지 사항)

### Test mapping

- `internal/template/hardcoded_path_audit_test.go::TestFallbackChainOrder` (new, M2)
- `internal/template/template_homedir_audit_test.go::TestNoHomeDirInFallback` (new, M2)

---

## AC-V3R2-RT-007-03 — Migration 1 rewrites hardcoded literal in v2.x user wrappers

Maps to: REQ-V3R2-RT-007-022.

### Happy path

- **Given** v2.x 사용자 프로젝트의 `.claude/hooks/moai/handle-session-start.sh` 에 리터럴 `/Users/goos/go/bin/moai` 포함
- **And** 다른 28개 wrapper 중 일부도 동일 리터럴 포함
- **When** session-start hook이 firing → `MigrationRunner.Apply` → `m001_hardcoded_path.Apply(projectRoot)` 호출
- **Then** 각 affected file의 `/Users/goos/go/bin/moai` 리터럴이 `$HOME/go/bin/moai` 로 정확히 치환
- **And** 다른 모든 콘텐츠 (comments, other commands, exec args) 보존
- **And** 파일 mode (`0o755`) 보존
- **And** atomic rename으로 partial-write 없음
- **And** `.moai/state/migration-version` 이 `1` 로 업데이트
- **And** `.moai/logs/migrations.log` 에 `{version: 1, name: "remove_hardcoded_gobin_path", result: "success", details: "rewritten N file(s)"}` JSONL 라인 append

### Edge case — 부분 매칭

- **Given** 어떤 wrapper에 리터럴이 없고 다른 wrapper엔 있음 (믹스)
- **When** Apply 실행
- **Then** 영향받은 파일만 rewrite; 나머지는 미변경
- **And** log entry는 rewritten count 정확히 기록

### Edge case — 사용자 hand-edit 보존

- **Given** wrapper 내부에 사용자가 추가한 주석 라인 `# my custom comment` 가 hardcoded 라인 옆에 있음
- **When** Apply 실행
- **Then** 주석 라인은 그대로; 리터럴만 치환

### Test mapping

- `internal/migration/migrations/m001_hardcoded_path_test.go::TestM001_RewritesHardcodedLiteral` (new, M1/M4) — `t.TempDir()` to construct fake v2.x project tree
- `internal/migration/migrations/m001_hardcoded_path_test.go::TestM001_PreservesOtherContent` (new, M1/M4)
- `internal/migration/migrations/m001_hardcoded_path_test.go::TestM001_PreservesExecutableBit` (new, M1/M4)

---

## AC-V3R2-RT-007-04 — Migration 1 idempotent on already-clean projects

Maps to: REQ-V3R2-RT-007-011, REQ-V3R2-RT-007-023.

### Happy path

- **Given** 사용자 프로젝트의 모든 wrapper가 이미 깨끗 (fresh v3 install 또는 이전 migration 1 적용 완료)
- **When** session-start hook → `m001_hardcoded_path.Apply` 호출
- **Then** 어떤 파일도 수정되지 않음
- **And** log entry: `{version: 1, result: "success", details: "already migrated (28 files scanned, 0 rewritten)"}`
- **And** 첫 적용 시 version 0 → 1, 이후 재적용 시 Pending(1) 이 빈 리스트라 m001은 호출조차 안 됨

### Edge case — version 1로 이미 표기됐으나 한 wrapper만 수동 reset

- **Given** `.moai/state/migration-version` 가 `1` 인데 사용자가 한 wrapper를 수동으로 v2.x 패턴으로 되돌림
- **When** session-start hook
- **Then** Pending(1) 은 빈 리스트 → m001 자동 호출 안 됨
- **And** 사용자가 `moai migration rollback 1` 후 `moai migration run` 으로 강제 재적용 가능 (단, m001은 Rollback nil이라 rollback도 거부됨 → 수동 fix 필요)

### Test mapping

- `internal/migration/migrations/m001_hardcoded_path_test.go::TestM001_NoOpWhenAlreadyClean` (new, M1/M4)
- `internal/migration/runner_test.go::TestRunner_Apply_Idempotent` (new, M1/M3) — runs Apply twice

---

## AC-V3R2-RT-007-05 — Fresh install applies all registered migrations in order

Maps to: REQ-V3R2-RT-007-030.

### Happy path

- **Given** `.moai/state/migration-version` 부재 (fresh v3.0 install OR v2.x → v3 첫 진입)
- **And** registry에 m001 등록 (현 SPEC 단계: m001 only; future: m002+)
- **When** session-start hook → `MigrationRunner.Apply(ctx)`
- **Then** runner는 `readVersion` 에서 0 반환
- **And** `Pending(0)` → `[m001]`
- **And** m001 Apply 실행 + version 1로 업데이트 + log append
- **And** 다음 session-start 에서는 `readVersion` → 1, `Pending(1)` → `[]` → no-op

### Edge case — multiple migrations

- **Given** registry에 m001, m002, m003 등록 (future scenario)
- **And** 사용자 version-file은 0
- **When** Apply 호출
- **Then** m001, m002, m003 차례로 Apply (오름차순)
- **And** 각 성공 후 version 즉시 atomic update (1 → 2 → 3)
- **And** m002 가 실패하면 version은 1에 머무름; m003 시도 안 됨

### Test mapping

- `internal/migration/runner_test.go::TestRunner_Apply_HappyPath` (new, M1/M3)
- `internal/migration/runner_test.go::TestRunner_Apply_FreshInstall_AllInOrder` (new, M1/M3)

---

## AC-V3R2-RT-007-06 — Failed migration: version unchanged + log entry + SystemMessage

Maps to: REQ-V3R2-RT-007-021, REQ-V3R2-RT-007-014.

### Happy path

- **Given** 등록된 migration이 Apply 중 에러 반환 (예: read-only 파일시스템 → REQ-052 `MigrationReadOnly`)
- **When** session-start hook → MigrationRunner.Apply
- **Then** 에러 발생한 migration의 version은 file에 반영되지 않음 (e.g., version still 0 if m001 failed)
- **And** log entry: `{version: 1, result: "failed", details: "MigrationReadOnly: ..."}`
- **And** session은 차단되지 않음 (REQ-021)
- **And** `HookOutput.SystemMessage` 가 `"migration failed: <error> (run 'moai doctor --check migration' for details)"` 패턴 포함 (RT-001 의 HookResponse 머지 후)
- **And** RT-001 미머지 시 `slog.Warn("migration apply failed", ...)` 로 임시 fallback

### Edge case — 부분 성공 (m001 OK, m002 FAIL)

- **Given** registry m001, m002 등록; m001은 정상, m002는 실패
- **When** Apply
- **Then** m001 적용 후 version 0 → 1
- **And** m002 시도 → 실패 → version 1 유지
- **And** log: m001 success entry + m002 failed entry

### Test mapping

- `internal/migration/runner_test.go::TestRunner_Apply_FailureHaltsAdvance` (new, M1/M3)
- `internal/migration/runner_test.go::TestRunner_Apply_PartialSuccess` (new, M1/M3)
- `internal/hook/session_start_migration_test.go::TestSessionStart_MigrationFailure_SurfacesViaSystemMessage` (new, M4)

---

## AC-V3R2-RT-007-07 — `moai doctor migration-status` reports current version + pending + last log

Maps to: REQ-V3R2-RT-007-015.

### Happy path

- **Given** `.moai/state/migration-version` 가 `1`, registry에 m001만 등록
- **And** `.moai/logs/migrations.log` 에 m001 success entry 존재
- **When** 사용자가 `moai doctor --check migration` (research §9.7 결정 패턴)
- **Then** stdout 포함:
  ```
  Migration Status:
    Current version: 1
    Registered versions: [1]
    Pending versions: []
    Last applied:
      Version: 1
      Name: remove_hardcoded_gobin_path
      Result: success
      Completed: 2026-MM-DD HH:MM:SS
      Details: rewritten 5 file(s)
  ```
- **And** exit code 0

### Edge case — `--json` 플래그

- **Given** 동일 상태
- **When** 사용자가 `moai migration status --json`
- **Then** stdout이 valid JSON: `{"current_version": 1, "registered_versions": [1], "pending_versions": [], "last_applied": {...}}`
- **And** JSON 파싱 가능

### Edge case — pending migrations 존재

- **Given** version-file이 1인데 registry에 m001+m002 (m002는 가상의 future migration)
- **When** doctor migration-status
- **Then** `Pending versions: [2]` 출력

### Test mapping

- `internal/cli/migration_test.go::TestMigrationStatus_Human` (new, M1/M5)
- `internal/cli/migration_test.go::TestMigrationStatus_JSON` (new, M1/M5)
- `internal/cli/doctor_migration_test.go::TestDoctorMigration_PrintsCurrentVersion` (new, M1/M5)
- `internal/cli/doctor_migration_test.go::TestDoctorMigration_PrintsPendingCount` (new, M1/M5)

---

## AC-V3R2-RT-007-08 — `moai migration rollback <v>` calls Rollback if declared

Maps to: REQ-V3R2-RT-007-024, REQ-V3R2-RT-007-042.

### Happy path

- **Given** registry에 가상의 m_rev (Rollback 함수 declared)
- **And** version-file은 그 m_rev 의 버전 수치
- **When** 사용자가 `moai migration rollback <v>`
- **Then** `m_rev.Rollback(projectRoot)` 호출
- **And** 성공 시 version-file → `<v>-1`
- **And** log entry: `{version: <v>, result: "rolled-back", ...}`
- **And** exit code 0

### Edge case — Rollback nil

- **Given** m001의 `Rollback: nil`
- **And** version-file은 1
- **When** `moai migration rollback 1`
- **Then** 명령 실패; exit code 1
- **And** stderr: `MigrationNotRollbackable: m001 (remove_hardcoded_gobin_path) is non-rollback-able by design (CRITICAL bug-fix)`
- **And** version-file 미변경

### Test mapping

- `internal/cli/migration_test.go::TestMigrationRollback_NoRollbackable` (new, M1/M5) — sentinel `MigrationNotRollbackable`
- `internal/cli/migration_test.go::TestMigrationRollback_Succeeds` (new, M1/M5) — uses test-only registered migration with Rollback declared

---

## AC-V3R2-RT-007-09 — `moai migration rollback` returns `MigrationNotRollbackable` for m001

Maps to: REQ-V3R2-RT-007-024.

### Happy path

- **Given** version-file은 `1` (m001 적용된 상태)
- **And** m001은 `Rollback: nil`
- **When** `moai migration rollback 1`
- **Then** stderr 출력: `MigrationNotRollbackable: m001 (remove_hardcoded_gobin_path) is non-rollback-able by design`
- **And** exit code != 0
- **And** version-file 그대로 1
- **And** log entry: `{version: 1, result: "rolled-back-failed", details: "MigrationNotRollbackable"}`

### Test mapping

- `internal/cli/migration_test.go::TestMigrationRollback_M001_Rejected` (new, M5)
- `internal/migration/migrations/m001_hardcoded_path_test.go::TestM001_RollbackNotImplemented` (new, M1/M4)

---

## AC-V3R2-RT-007-10 — Template author adding `.HomeDir` in fallback fails CI

Maps to: REQ-V3R2-RT-007-050.

### Happy path

- **Given** 미래 contributor가 `internal/template/templates/.claude/hooks/moai/handle-XXX.sh.tmpl` 에 `if [ -f "{{posixPath .HomeDir}}/go/bin/moai" ]; then` 라인 추가
- **When** CI가 `go test ./internal/template/ -run TestNoHomeDirInFallback`
- **Then** 테스트 FAIL with sentinel `HOMEDIR_IN_FALLBACK_CHAIN: <file> line <N> contains {{posixPath .HomeDir}}; use $HOME (shell-expanded) instead. See SPEC-V3R2-RT-007 REQ-004 / CLAUDE.local.md §14.`
- **And** PR is blocked from merging

### Test mapping

- `internal/template/template_homedir_audit_test.go::TestNoHomeDirInFallback` (new, M2)
- `internal/template/template_homedir_audit_test.go::TestNoHomeDirInFallback_ContributorRegression` (new, M2) — synthetic test using `t.TempDir()`

---

## AC-V3R2-RT-007-11 — Two migrations with same Version panic at init

Maps to: REQ-V3R2-RT-007-053.

### Happy path

- **Given** registry에 두 마이그레이션이 동일 Version 1 등록 (가상의 잘못된 contributor 변경)
- **When** `go test ./internal/migration/...` 실행 → `init()` 가 호출되며 registry validate
- **Then** panic: `DuplicateMigrationVersion: version 1 declared by both 'remove_hardcoded_gobin_path' and '<other>'`
- **And** 어떤 다른 테스트도 실행되지 않음 (init panic이 테스트 바이너리 init 단계 차단)

### Test mapping

- `internal/migration/registry_test.go::TestRegistry_DuplicateVersion_Panics` (new, M1/M3) — uses test-local registry copy with intentional duplicate; asserts `recover()` produces the sentinel string

---

## AC-V3R2-RT-007-12 — Version-file ahead of known migrations: graceful no-op

Maps to: REQ-V3R2-RT-007-054.

### Happy path

- **Given** `.moai/state/migration-version` 가 `99` (사용자가 v3.5에서 v3.0으로 다운그레이드)
- **And** registry max version은 `1`
- **When** session-start hook → MigrationRunner.Apply
- **Then** runner는 `Pending(99)` 호출 → `[]` (registered versions 모두 ≤ 99)
- **And** `HookOutput.SystemMessage` 에 `"Version file ahead of known migrations; treating as no-op"` 포함
- **And** version-file 미변경
- **And** log entry: `{version: -1, result: "skipped", details: "version file ahead (99 > registry max 1)"}`
- **And** session 진행 정상

### Test mapping

- `internal/migration/runner_test.go::TestRunner_Apply_VersionAhead` (new, M1/M3)

---

## AC-V3R2-RT-007-13 — `migrations.disabled: true` skips runner with no SystemMessage

Maps to: REQ-V3R2-RT-007-032.

### Happy path

- **Given** `.moai/config/sections/system.yaml` 가 `migrations:\n  disabled: true` 포함
- **And** version-file 부재 (즉, fresh v3.0 install — 일반적 상황이라면 m001이 적용되어야 할 상태)
- **When** session-start hook
- **Then** `migrationsDisabled(cfg)` → true → MigrationRunner.Apply 호출 자체 skip
- **And** `HookOutput.SystemMessage` 비어있음 (REQ-032 NO SystemMessage 명시)
- **And** log entry: `{version: 0, result: "skipped", details: "disabled via system.yaml"}`
- **And** version-file 미변경

### Edge case — `disabled: false` (또는 미설정)

- **Given** `migrations.disabled: false` 또는 키 미존재
- **When** session-start hook
- **Then** runner 정상 호출 (기본 경로)

### Test mapping

- `internal/hook/session_start_migration_test.go::TestSessionStart_DisabledViaSystemYaml` (new, M4)
- `internal/hook/session_start_migration_test.go::TestSessionStart_EnabledByDefault` (new, M4)

---

## AC-V3R2-RT-007-14 — Crash mid-Apply: re-run completes idempotently

Maps to: REQ-V3R2-RT-007-031, REQ-V3R2-RT-007-061.

### Happy path

- **Given** Apply가 `migration-version.tmp` 파일 생성 후 `os.Rename` 직전 프로세스 사망 (예: kill -9)
- **When** 다음 session-start hook
- **Then** runner는 `version-file` 존재 여부 확인 (정상 위치)
- **And** `version-file.tmp` 잔재 발견 시 (in-flight detection): `os.Remove(*.tmp)` 후 m001 재실행 (m001은 idempotent — 이미 일부 wrappers는 깨끗할 수 있음)
- **And** Apply 재실행 → 미완료된 wrapper rewrite 완성 → version-file atomic update → `.tmp` 정리
- **And** 최종 상태: version 1, 모든 wrapper 깨끗

### Edge case — write-during-rename

- **Given** `os.Rename` 자체가 atomic이라 실패 시 source 그대로 (Go stdlib 보장)
- **When** rename 실패 시
- **Then** 최종 version-file 미변경; m001은 재실행 시 다시 Apply 시도 (idempotent)

### Test mapping

- `internal/migration/version_test.go::TestVersionFile_AtomicRename` (new, M1/M3) — simulates crash via partial write + verifies recovery
- `internal/migration/runner_test.go::TestRunner_Apply_CrashRecovery` (new, M1/M3) — uses fault-injection

---

## AC-V3R2-RT-007-15 — Windows Git Bash environment: $HOME expansion at runtime

Maps to: REQ-V3R2-RT-007-060.

### Happy path

- **Given** Windows Git Bash 환경에서 사용자 프로젝트의 `.claude/hooks/moai/handle-session-start.sh` 가 `/Users/goos/go/bin/moai` 리터럴 포함 (예: 누군가 macOS에서 만든 프로젝트를 Windows로 가져옴)
- **When** session-start hook → m001 Apply
- **Then** 파일 내 리터럴이 `$HOME/go/bin/moai` 문자열로 치환 (마이그레이션은 *문자열 치환만* 수행, 변수 expansion 안 함)
- **And** 다음 session-start 시 Bash가 `$HOME` → user home expand → 실제 경로 변환
- **And** 마이그레이션 자체는 platform-independent (Windows/macOS/Linux 동일 동작)

### Edge case — WSL vs Git Bash

- **Given** WSL과 Git Bash의 `$HOME` 값이 다름 (WSL: `/home/X`, Git Bash: `/c/Users/X`)
- **When** Bash가 expand
- **Then** 각 환경의 `$HOME` 값이 정확히 사용 (마이그레이션과 무관)

### Test mapping

- `internal/migration/migrations/m001_hardcoded_path_test.go::TestM001_WindowsGitBash` (new, M1/M4)

---

## AC-V3R2-RT-007-16 — `moai migration run` manually applies pending migrations

Maps to: REQ-V3R2-RT-007-040.

### Happy path

- **Given** version-file 부재 (fresh state)
- **And** session 활성 안 됨 (CI 또는 manual 환경)
- **When** 사용자가 `moai migration run` 실행
- **Then** runner가 `Pending(0)` → `[m001]` 으로 m001 Apply
- **And** version-file 1로 업데이트
- **And** stdout: `applied migration 1 (remove_hardcoded_gobin_path): rewritten N file(s)`
- **And** exit code 0

### Edge case — pending 없음

- **Given** version-file 1, registry max 1
- **When** `moai migration run`
- **Then** stdout: `no pending migrations (current version: 1)`
- **And** exit code 0

### Edge case — `--dry-run` 플래그 (optional)

- **Given** pending migrations 있음
- **When** `moai migration run --dry-run`
- **Then** stdout: 실행될 마이그레이션 목록 (Version, Name) 출력하지만 Apply 안 함
- **And** version-file 미변경

### Test mapping

- `internal/cli/migration_test.go::TestMigrationRun_AppliesPending` (new, M1/M5)
- `internal/cli/migration_test.go::TestMigrationRun_NoPending` (new, M1/M5)
- `internal/cli/migration_test.go::TestMigrationRun_DryRun` (new, M1/M5) — optional, only if `--dry-run` 플래그 결정됨 in run-phase

---

## Summary table — AC → REQ → Test

| AC | REQs covered | Test files |
|----|--------------|------------|
| AC-01 | REQ-002, REQ-025, REQ-051 | `template/hardcoded_path_audit_test.go::*` (3 cases) |
| AC-02 | REQ-003 | `template/hardcoded_path_audit_test.go::TestFallbackChainOrder`, `template_homedir_audit_test.go` |
| AC-03 | REQ-022 | `migration/migrations/m001_hardcoded_path_test.go::TestM001_RewritesHardcodedLiteral`, `_PreservesOtherContent`, `_PreservesExecutableBit` |
| AC-04 | REQ-011, REQ-023 | `migration/migrations/m001_hardcoded_path_test.go::TestM001_NoOpWhenAlreadyClean`, `migration/runner_test.go::TestRunner_Apply_Idempotent` |
| AC-05 | REQ-030 | `migration/runner_test.go::TestRunner_Apply_FreshInstall_AllInOrder`, `_HappyPath` |
| AC-06 | REQ-021, REQ-014 | `migration/runner_test.go::TestRunner_Apply_FailureHaltsAdvance`, `_PartialSuccess`, `hook/session_start_migration_test.go::TestSessionStart_MigrationFailure_SurfacesViaSystemMessage` |
| AC-07 | REQ-015 | `cli/migration_test.go::TestMigrationStatus_*`, `cli/doctor_migration_test.go::TestDoctorMigration_*` (4 cases) |
| AC-08 | REQ-024, REQ-042 | `cli/migration_test.go::TestMigrationRollback_NoRollbackable`, `_Succeeds` |
| AC-09 | REQ-024 | `cli/migration_test.go::TestMigrationRollback_M001_Rejected`, `migrations/m001_hardcoded_path_test.go::TestM001_RollbackNotImplemented` |
| AC-10 | REQ-050 | `template/template_homedir_audit_test.go::TestNoHomeDirInFallback`, `_ContributorRegression` |
| AC-11 | REQ-053 | `migration/registry_test.go::TestRegistry_DuplicateVersion_Panics` |
| AC-12 | REQ-054 | `migration/runner_test.go::TestRunner_Apply_VersionAhead` |
| AC-13 | REQ-032 | `hook/session_start_migration_test.go::TestSessionStart_DisabledViaSystemYaml`, `_EnabledByDefault` |
| AC-14 | REQ-031, REQ-061 | `migration/version_test.go::TestVersionFile_AtomicRename`, `migration/runner_test.go::TestRunner_Apply_CrashRecovery` |
| AC-15 | REQ-060 | `migration/migrations/m001_hardcoded_path_test.go::TestM001_WindowsGitBash` |
| AC-16 | REQ-040 | `cli/migration_test.go::TestMigrationRun_AppliesPending`, `_NoPending`, `_DryRun` (optional) |

Total new test functions: **~40 across 12 new test files**.

---

## Definition of Done

This SPEC is considered done when ALL of the following are true:

1. All 16 ACs above pass under `go test ./internal/migration/... ./internal/cli/... ./internal/template/... ./internal/hook/... ./internal/runtime/...`.
2. Full `go test ./...` from worktree root passes with zero failures and zero cascading regressions.
3. `make build` succeeds and `internal/template/embedded.go` regenerates cleanly.
4. `go vet ./...` and `golangci-lint run` pass with zero warnings.
5. `progress.md` is updated with `run_complete_at: <timestamp>` and `run_status: implementation-complete`.
6. CHANGELOG entry is present under `## [Unreleased] / ### Added`.
7. 7 MX tags are inserted per `plan.md` §6 (3 ANCHOR, 2 NOTE, 2 WARN).
8. 3 CI lint tests (hardcoded-path audit, homedir audit, passthrough affirm) added and passing.
9. `moai migration run`, `moai migration status`, `moai migration rollback`, `moai doctor --check migration` CLI surfaces functional.
10. PR opened by `manager-git` has all required CI checks green (Lint, Test ubuntu/macos/windows, Build all 5, CodeQL).

---

End of acceptance.md.
