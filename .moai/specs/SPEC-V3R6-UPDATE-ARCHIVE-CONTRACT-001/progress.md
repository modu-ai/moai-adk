---
id: SPEC-V3R6-UPDATE-ARCHIVE-CONTRACT-001
title: "SPEC-V3R6-UPDATE-ARCHIVE-CONTRACT-001 — Progress"
version: "0.2.0"
status: implemented
created: 2026-05-23
updated: 2026-05-23
author: manager-develop
priority: P1
phase: "v3.6.0"
module: "internal/cli"
lifecycle: spec-anchored
tags: "v3r6, update, archive, force, skip-sync, contract, progress"
tier: M
---

# Progress — SPEC-V3R6-UPDATE-ARCHIVE-CONTRACT-001

본 progress 문서는 manager-develop cycle_type=ddd Section A-E Tier M 위임으로 진행된 run-phase 산출 증거를 정리한다. DDD 메소드 (ANALYZE → PRESERVE → IMPROVE) 적용.

## 마일스톤 결과

### M1 — `archiveLegacySkills` 시그니처 확장 + force 분기 (완료)

`internal/cli/update_archive.go`:
- import 에 `"time"` 추가.
- `archiveLegacySkills(projectRoot string, out io.Writer) (int, error)` → `archiveLegacySkills(projectRoot string, out io.Writer, force bool) (int, error)`.
- `time.Now().UTC().Format("20060102T150405Z")` 으로 호출당 단일 drift 타임스탬프 생성.
- `force == true && alreadyArchived` 시 `checkArchiveDrift` 사전 호출 → drift 검출되면 `archive/skills/v2.16-drift-<ts>/<id>/` 백업 디렉터리로 `os.Rename` → `archiveSkill` 재호출.
- 백업 부모 디렉터리 생성 실패 또는 `os.Rename` 실패 시 wrapped error 반환 (원본 archive 보존).
- 출력에 `archive drift backup: <id> → <relpath>` 라인 추가 (R3/R4 추적성).
- @MX:NOTE 갱신 (force 추가 + Windows-safe ISO8601 명시).

### M2 — `runUpdate` 호출부의 `--force` 전파 + skip-sync 단락 (완료)

`internal/cli/update.go`:
- `runTemplateSyncWithProgress(cmd) error` → `runTemplateSyncWithProgress(cmd) (skipped bool, err error)` (Option A, plan §4.3 채택).
  - 버전 일치 + `!force` → `return true, nil`.
  - 사용자 머지 취소 → `return true, nil` (no-op이므로 archive skip).
  - 정상 진행 → `return false, runTemplateSyncWithReporter(...)`.
- `runUpdate` (`update.go:233`) 에서 `syncSkipped, err := runTemplateSyncWithProgress(cmd)` 수신 → `if syncSkipped { return nil }` 로 archive 블록 단락.
- `archiveLegacySkills(cwd, out, getBoolFlag(cmd, "force"))` 호출로 --force 전파.
- `--force` help text (`update.go:85`) 갱신: "Force update: bypass version-match skip, force backup+merge, and overwrite archive drift (backed up to .moai/archive/skills/v2.16-drift-<UTC-timestamp>/)".
- `runUpdate` godoc (`update.go:104` 부근) 갱신: --force 의 3가지 효과 명시 (bypass version-match / force backup+merge / overwrite archive drift with lossless backup directory).
- `runTemplateSyncWithProgress` godoc 갱신: skipped 의미 명문화.

### M3 — 회귀 방어 테스트 신규 작성 (완료)

`internal/cli/update_archive_force_test.go` 신규:
- `TestArchiveForce/force_with_drift_creates_backup_and_overwrites` PASS — 드리프트 백업 디렉터리 정확히 1개 생성 (`filepath.Glob` 기반, R3 timestamp-flakiness 회피), 백업은 pre-drift 내용 보존, 라이브 archive 는 src 와 일치, "archive drift backup:" 출력 라인 확인.
- `TestArchiveForce/force_without_drift_is_idempotent` PASS — drift 없을 시 백업 0개.
- `TestArchiveForce/force_false_preserves_drift_error` PASS — BC-V3R3-007 `ARCHIVE_DRIFT` 에러 보존.
- `TestArchiveForce/force_with_drift_backup_failure_preserves_original` PASS — chmod 0o555 parent → 백업 실패 시 원본 archive 보존 + wrapped error. Windows 는 chmod-based denial 불안정성으로 skip.

`internal/cli/update_skip_sync_test.go` 신규:
- `TestSkipSyncNoArchive/version_match_returns_skipped_true` PASS — version match + !force → `skipped=true`, "Legacy skill archive" 출력 부재 확인 (REQ-UAC-004 contract).
- `TestSkipSyncNoArchive/skip_sync_with_force_does_invoke_archive` PASS — --force 시 `skipped=false` (skip-sync 단락 우회).

모든 테스트 `t.TempDir()` 격리 — dev project 의 `.claude/`, `.moai/archive/` 절대 수정 안 함.

### M4 — 기존 테스트 호출자 시그니처 정합 (완료)

`archiveLegacySkills(` 호출 4개 + `runTemplateSyncWithProgress(` 호출 4개 모두 새 시그니처로 정합:
- `internal/cli/update_idempotency_test.go` (2 sites) — `false` arg 추가.
- `internal/cli/update_archive_flow_test.go` (2 sites) — `false` arg 추가.
- `internal/cli/coverage_improvement_test.go` (2 sites) — `(skipped, err)` 수신 패턴, version match path 는 `skipped=true` 명시 검증.
- `internal/cli/target_coverage_test.go` (2 sites) — `(skipped, err)` 수신, --force 시 `skipped=false` 검증.

### M5 — `dryRunArchiveLegacySkills` 정합 (Deferred per S4)

Plan §M5 + Section D S4 결정: 본 SPEC 범위 밖. `dryRunArchiveLegacySkills` 시그니처 무변경. 후속 SPEC 후보 (`SPEC-V3R6-UPDATE-RESTORE-001` 또는 별도 dry-run 정합 SPEC) 로 위임.

### M6 — chore (완료)

- spec.md / plan.md / acceptance.md: `status: draft` → `implemented`, `version: 0.1.0` → `0.2.0`, author `manager-spec` → `manager-develop` (spec.md only).
- 본 progress.md 작성.

---

## Acceptance Criteria 결과 매트릭스 (10/10)

| AC | Status | 검증 방법 | 증거 |
|----|--------|-----------|------|
| AC-UAC-001 | PASS | `grep -n 'func archiveLegacySkills' internal/cli/update_archive.go` | `update_archive.go:240: func archiveLegacySkills(projectRoot string, out io.Writer, force bool) (int, error)` |
| AC-UAC-002 | PASS | `grep -rn 'archiveLegacySkills(' internal/cli/ \| grep -v _test.go` | `update.go:251: archiveLegacySkills(cwd, out, getBoolFlag(cmd, "force"))` |
| AC-UAC-003 | PASS | `go test -run TestArchiveForce/force_with_drift_creates_backup_and_overwrites ./internal/cli/ -v` | `--- PASS: TestArchiveForce/force_with_drift_creates_backup_and_overwrites (0.06s)` |
| AC-UAC-004 | PASS | glob `v2.16-drift-*` in test asserts `len(matches) == 1` | `filepath.Glob(...) → 1 match, backup file content == oldContent (pre-drift)` |
| AC-UAC-005 | PASS | `go test -run TestSkipSyncNoArchive ./internal/cli/ -v` | `--- PASS: TestSkipSyncNoArchive (0.00s)` (both subtests) |
| AC-UAC-006 | PASS | `go test -run 'TestArchive\|TestIdempotency' ./internal/cli/ -v` | 모든 archive/idempotency 기존 테스트 PASS (force=false 경로 보존) |
| AC-UAC-007 | PASS | `git grep -n 'Use --force to overwrite' internal/` + 코드리뷰 | `update_archive.go:141, :153` 메시지 그대로, force 경로 (`update_archive.go:264-285`) 실제 구현됨 → 메시지가 거짓말 아님 |
| AC-UAC-008 | DEFERRED | Manual reproducer in `~/moai/mo.ai.kr` (post-merge user action) | 사용자가 merge 후 `cd ~/moai/mo.ai.kr && moai update --force` 실행 → diff -q clean 확인 + `ls .moai/archive/skills/v2.16-drift-*/` 1 backup 확인 |
| AC-UAC-009 | PASS | `grep -n '"force"' + '--force:' internal/cli/update.go` | `update.go:85` help text + `update.go:107-115` godoc 모두 archive drift overwrite 의미 명시 |
| AC-UAC-010 | PASS | `grep -n 't.TempDir()' internal/cli/update_archive_force_test.go internal/cli/update_skip_sync_test.go` + `go test -race` | 모든 신규 테스트 함수 `t.TempDir()` 사용, race PASS (0.627s clean run) |

---

## TRUST 5 + Cross-Platform 검증

| 항목 | 결과 |
|------|------|
| `go build ./...` (darwin/arm64 host) | exit 0 |
| `GOOS=linux GOARCH=amd64 go build ./...` | exit 0 |
| `GOOS=windows GOARCH=amd64 go build ./...` | exit 0 (timestamp ISO8601 no-colon → path-safe) |
| `go vet ./...` | exit 0 |
| `golangci-lint run --timeout=2m ./internal/cli/...` | 17 issues (pre-existing baseline) = 0 NEW lint regression |
| `go test -cover ./internal/cli/...` | 70.5% (identical to pre-implementation baseline) |
| `go test -race -run 'TestArchiveForce\|TestSkipSyncNoArchive' ./internal/cli/` | PASS, 0 race |
| C-HRA-008 boundary grep | 0 production-code AskUserQuestion calls introduced (existing matches are agent_lint rule docstrings + godoc descriptions, all PRE-existing) |

`internal/cli` 패키지 전체 테스트는 `TestDoctor_*` / `TestStatus_*` (6개) 가 pre-existing version-string baseline drift (v2.17.0 hardcoded) 로 FAIL — 본 SPEC scope 밖, 0 NEW failure.

---

## DDD Methodology Evidence

ANALYZE (pre-implementation):
- `update.go:227-241` archive 블록 + `update.go:765-772` skip-sync 분기 + `update_archive.go:233` 시그니처 + `update_archive.go:138-158` 에러 메시지 모두 읽어 컨트랙트 결함 2건 (force 미전파 + skip-sync 분기 누락) 확인.
- 6 production + test caller sites 의 fan-out 매핑.

PRESERVE:
- BC-V3R3-007 idempotency 컨트랙트 (REQ-UAC-006): force=false 경로의 4 시나리오 (source absent / archive matches / archive differs / 16-skill walk) 모두 기존 테스트로 회귀 차단.
- `legacySkillIDs`, `archiveVersion`, `archiveSkill`, `checkArchiveDrift`, `computeDirHashes`, `copyDirAll` 모두 unchanged.

IMPROVE:
- M1 → M2 → M3 → M4 단계적 적용, 각 단계 후 `go build ./...` 검증.
- M1 후 시그니처 변경으로 인한 컴파일 에러는 M4 호출자 정합으로 해소.
- 신규 테스트 (M3) PASS 확인 후 M4 진행.

---

## 다음 단계

1. 사용자 결정: feat 브랜치 vs main 직진 push (Hybrid Trunk Tier M doctrine).
2. 머지 후 AC-UAC-008 사용자 수동 검증 (`~/moai/mo.ai.kr` reproducer).
3. SPEC-V3R6-UPDATE-NOISE-001 + SPEC-V3R6-UPDATE-PROGRESS-001 run-phase 진입.
