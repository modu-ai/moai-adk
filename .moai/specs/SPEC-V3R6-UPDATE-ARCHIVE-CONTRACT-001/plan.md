---
id: SPEC-V3R6-UPDATE-ARCHIVE-CONTRACT-001
title: "SPEC-V3R6-UPDATE-ARCHIVE-CONTRACT-001 — Implementation Plan"
version: "0.1.0"
status: draft
created: 2026-05-23
updated: 2026-05-23
author: manager-spec
priority: P1
phase: "v3.6.0"
module: "internal/cli"
lifecycle: spec-anchored
tags: "v3r6, update, archive, force, skip-sync, contract, plan"
tier: M
---

# Implementation Plan — SPEC-V3R6-UPDATE-ARCHIVE-CONTRACT-001

## 1. 개요

본 계획은 `moai update` 의 archive 서브시스템 컨트랙트 결함 2건 (`--force` 미전파 + skip-sync 분기 archive 트리거) 을 단일 패치 사이클로 해소한다. Tier M 분류이며 변경 범위는 Go 소스 ~3-5 파일 + 신규 테스트 2개 + 기존 테스트 시그니처 정합 업데이트로 예상된다.

본 plan-phase 산출물은 markdown only 이며, 실제 Go 코드 수정은 후속 run-phase (`/moai run SPEC-V3R6-UPDATE-ARCHIVE-CONTRACT-001`) 에서 manager-develop cycle_type=ddd 위임으로 진행한다.

---

## 2. 영향 범위

### 2.1 수정 대상 파일

| 파일 | 변경 종류 | 추정 LOC | 비고 |
|------|----------|----------|------|
| `internal/cli/update_archive.go` | 수정 | +30 / -5 | `archiveLegacySkills` 시그니처 확장, force 분기 + drift 백업 |
| `internal/cli/update.go` | 수정 | +15 / -3 | force 전파, skip-sync 단락, help/godoc 갱신 |
| `internal/cli/update_archive_test.go` | 수정 | +5 / -2 | 기존 호출자 시그니처 정합 |
| `internal/cli/update_archive_force_test.go` | 신규 | ~80 | `TestArchiveForce` + drift 백업 검증 |
| `internal/cli/update_skip_sync_test.go` | 신규 | ~60 | `TestSkipSyncNoArchive` skip-sync 단락 검증 |
| `internal/cli/update_archive_flow_test.go` | 수정 (옵션) | +5 / -2 | 호출자 시그니처 정합 |

총 추정: +195 / -12, 약 200 LOC 신규.

### 2.2 영향받지 않는 영역 (PRESERVE)

- `legacySkillIDs` 목록 (16개) 변경 없음.
- `archiveVersion` 상수 (`"v2.16"`) 변경 없음.
- `archiveSkill` 본체 (path traversal 가드 포함) 변경 없음.
- `checkArchiveDrift`, `computeDirHashes`, `copyDirAll` 구현 변경 없음 (재사용).
- BC-V3R3-007 의 archive 디렉터리 레이아웃 변경 없음.
- 다른 archive 서브시스템 (design folder, evolution dir) 변경 없음.

---

## 3. 단계별 마일스톤 (Priority-based, 시간 없음)

### M1 — `archiveLegacySkills` 시그니처 확장 + force 분기 (Priority: High)

**범위**: `internal/cli/update_archive.go`

작업:

1. `archiveLegacySkills(projectRoot string, out io.Writer)` 시그니처를 `archiveLegacySkills(projectRoot string, out io.Writer, force bool) (int, error)` 로 확장.
2. 함수 본체 내 `archiveSkill(projectRoot, id)` 호출 직전, `force == true && os.Stat(dstDir) == nil` 인 경우 `checkArchiveDrift` 우선 호출하여 drift 가 감지되면 백업 + overwrite 분기로 진입하는 로직 추가.
3. 백업 디렉터리 경로 생성: `filepath.Join(projectRoot, ".moai", "archive", "skills", archiveVersion + "-drift-" + time.Now().UTC().Format("20060102T150405Z"), id)`.
4. `os.Rename(dstDir, backupDir)` 후 `archiveSkill(projectRoot, id)` 재호출하여 신규 archive 생성.
5. `os.Rename` 또는 백업 디렉터리 생성 실패 시 원본 archive 보존 + 에러 반환.
6. force=false 경로는 기존 동작 (idempotency + drift 감지 + ARCHIVE_DRIFT 에러) 그대로 유지.

검증: `go build ./internal/cli/...` PASS, 기존 호출자 컴파일 에러는 M2 에서 정합.

### M2 — `runUpdate` 호출부의 `--force` 전파 + skip-sync 단락 (Priority: High)

**범위**: `internal/cli/update.go`

작업:

1. `update.go:238` 호출부를 `archiveLegacySkills(cwd, out, getBoolFlag(cmd, "force"))` 로 갱신.
2. skip-sync 단락 방식 채택: **Option A 채택 권장** — `runTemplateSyncWithProgress` 반환 시그니처를 `(skipped bool, err error)` 로 변경하여 호출자가 skip 여부를 인지하도록 한다. `update.go:227-241` 블록에서 `if skipped { goto skipArchive }` 또는 early return 추가.
3. **Option B 대안** — `runUpdate` 가 직접 버전 일치 + force 조합을 판별하여 sync 이전에 archive skip 결정. 호출자 책임 증가하나 시그니처 변경 회피. (Trade-off: A 는 호출자 단순성 / B 는 헬퍼 시그니처 안정성)
4. plan-phase 채택: **Option A**. 이유: 호출자가 skip-sync 의미를 알아야 하는 응집도. 시그니처 변경의 호출 사이트는 1곳 (`runUpdate`) 으로 비용 최소.
5. `update.go:85` help 텍스트 갱신: `"Force update even if version matches (still performs backup, merge, and archive overwrite)"`.
6. `update.go:108` godoc 주석 갱신: `"--force: Force version check bypass + archive drift overwrite (preserves backup via .moai/archive/skills/v2.16-drift-<timestamp>/)"`.

검증: `go build ./...` PASS, `go vet ./...` PASS.

### M3 — 회귀 방어 테스트 신규 작성 (Priority: High)

**범위**: `internal/cli/update_archive_force_test.go` 신규, `internal/cli/update_skip_sync_test.go` 신규

작업:

1. `TestArchiveForce` 작성:
   - Subtest `force_with_drift_creates_backup_and_overwrites`: `t.TempDir()` 에 src/archive 디렉터리 생성, archive 에 의도적 drift 도입, `archiveLegacySkills(tmp, &out, true)` 호출 → 백업 디렉터리 존재 검증 + archive 가 src 와 일치 검증.
   - Subtest `force_without_drift_is_idempotent`: drift 없는 환경에서 force=true 호출 → 백업 디렉터리 생성되지 않음, archive 변경 없음.
   - Subtest `force_false_preserves_drift_error`: 기존 동작 (`ARCHIVE_DRIFT` 에러) 회귀 검증.
   - Subtest `force_with_drift_backup_failure_preserves_original`: 백업 디렉터리 생성 실패 시 원본 archive 가 보존되는지 검증.

2. `TestSkipSyncNoArchive` 작성:
   - 임시 프로젝트 디렉터리 생성 + `config.yaml` 의 template_version 을 현재 binary version 과 일치하도록 설정.
   - `runUpdate` 실행 (또는 호출 가능한 사이즈로 분해된 helper) → skip-sync 분기 활성화.
   - archive 디렉터리에 의도적 drift 가 있더라도 ARCHIVE_DRIFT 에러가 출력되지 않음을 검증.
   - `--force` 명시 시에는 sync + archive 정상 실행됨을 별도 subtest 로 검증.

3. 모든 테스트는 `t.TempDir()` 으로 격리 (CLAUDE.local.md §6 준수). 사용자 프로젝트 (`.claude/`, `.moai/archive/`) 를 절대 변경하지 않음.

검증: `go test -run TestArchiveForce ./internal/cli/` PASS, `go test -run TestSkipSyncNoArchive ./internal/cli/` PASS.

### M4 — 기존 테스트 호출자 시그니처 정합 (Priority: Medium)

**범위**: `internal/cli/update_archive_test.go`, `internal/cli/update_archive_flow_test.go`, 기타 `archiveLegacySkills` 호출 테스트

작업:

1. `grep -rn 'archiveLegacySkills(' internal/cli/ | grep _test.go` 로 호출 사이트 전수 조사.
2. 각 호출자에 `false` 인자 (force=false, 기본 의미) 추가.
3. 의도적으로 force=true 를 검증해야 하는 테스트는 M3 의 신규 테스트에서만 다루며, 기존 테스트는 force=false 의미를 명시한다.

검증: `go test ./internal/cli/...` 전체 PASS, 기존 idempotency 회귀 테스트 영향 없음.

### M5 — `dryRunArchiveLegacySkills` 시그니처 정합 (옵션) (Priority: Low)

**범위**: `internal/cli/update_archive.go:270+`

작업:

1. `dryRunArchiveLegacySkills` 가 force 정보를 dry-run 출력에 포함해야 하는지 결정.
2. 결정 시: 시그니처 확장 `dryRunArchiveLegacySkills(projectRoot string, out io.Writer, force bool)`, force=true 일 때 "would backup to v2.16-drift-... and overwrite" 출력.
3. 결정 보류 시: 별도 follow-up SPEC.

판단: run-phase 진입 후 manager-develop 이 dry-run UX 검토와 함께 결정. plan-phase 에서는 옵션으로 표시.

### M6 — 검증 + chore (Priority: Medium)

**범위**: `acceptance.md` 의 검증 커맨드 일괄 실행, status 갱신

작업:

1. `acceptance.md` 의 10개 AC 검증 커맨드 전수 실행.
2. `~/moai/mo.ai.kr` (또는 동등 reproducer 디렉터리) 에서 `moai update --force` 수동 검증 (AC-UAC-008).
3. `spec.md` / `plan.md` / `acceptance.md` 의 `status: draft` → `status: implemented` 갱신, `version: 0.1.0` → `version: 0.2.0`.
4. `progress.md` 작성: M1~M5 증거 보존 + AC 결과 표.

---

## 4. 기술적 접근

### 4.1 `--force` 의미 단일 진실 공급원 (R2 완화)

본 SPEC 후 `--force` 의미 표 (godoc 단일 출처):

| 의미 측면 | force=true 동작 | force=false 동작 |
|----------|-----------------|-------------------|
| 버전 일치 check | 우회 (sync 실행) | 일치 시 skip-sync |
| backup/merge | 사용자 confirmation 우회하여 강제 진행 | 사용자 confirmation 요구 |
| archive drift | 기존 archive 백업 후 overwrite | ARCHIVE_DRIFT 에러 반환 |
| backup 디렉터리 | `.moai/archive/skills/v2.16-drift-<ts>/` 자동 생성 | 해당 없음 |

godoc 의 모순적 안내 (`update.go:85` help vs `update.go:108` godoc) 를 본 표로 정렬하고, 두 곳을 표 내용으로 동기화한다.

### 4.2 Drift 백업 디렉터리 명명

형식: `archive/skills/v2.16-drift-<YYYYMMDDTHHMMSSZ>/<skill-id>/`

예: `archive/skills/v2.16-drift-20260523T143022Z/moai-domain-backend/`

장점: 1) BC-V3R3-007 의 `archive/skills/<version>/` 패턴과 분리 유지 (drift 백업은 별도 version-like prefix), 2) timestamp 정렬로 시간순 복구 가능, 3) UTC ISO8601 으로 cross-locale 일관성.

### 4.3 Skip-sync 단락 — Option A 구현 스케치

`runTemplateSyncWithProgress` 의 반환 변경:

```
func runTemplateSyncWithProgress(cmd *cobra.Command) (skipped bool, err error)
```

호출자 변경 (`update.go:227-241`):

```
skipped, err := runTemplateSyncWithProgress(cmd)
if err != nil {
    return err
}
if skipped {
    // skip-sync 분기: archive 검사도 단락
    return nil
}
// (이후 archive 블록 진입)
```

기존 `return nil` (skip-sync) 을 `return true, nil` 로, `return runTemplateSyncWithReporter(...)` 를 `return false, runTemplateSyncWithReporter(...)` 로 정합.

### 4.4 호환성 보존

- BC-V3R3-007 archive 레이아웃 무변경.
- force=false 경로의 idempotency 컨트랙트 무변경.
- 기존 dry-run (`--dry-run`) 동작 무변경 (M5 결정에 따라 변경 가능).
- 외부 도구 호출 인터페이스 무변경 (`moai update` CLI 시그니처 유지).

---

## 5. 검증 전략

### 5.1 단위 테스트 (M3 신규)

- `TestArchiveForce` 4 subtest 로 force 경로 4 시나리오 커버.
- `TestSkipSyncNoArchive` 로 skip-sync 단락 검증.
- 모든 테스트 `t.TempDir()` 격리.

### 5.2 통합 / 수동 검증 (AC-UAC-008)

- `~/moai/mo.ai.kr` reproducer 디렉터리에서 `moai update --force` 실행.
- 기대 결과:
  - "Skipping sync" 출력 없음 (force 활성 시 sync 진행).
  - "Legacy skill archive failed" 출력 없음.
  - `.moai/archive/skills/v2.16-drift-<ts>/moai-domain-backend/` 백업 디렉터리 생성.
  - `.moai/archive/skills/v2.16/moai-domain-backend/SKILL.md` 가 `.claude/skills/moai-domain-backend/SKILL.md` 와 `diff -q` 일치.

### 5.3 회귀 방어

- 기존 `update_archive_test.go` (모든 force=false 케이스) PASS.
- 기존 `update_archive_flow_test.go` PASS.
- `update_idempotency_test.go` PASS.
- `go test ./internal/cli/...` 전체 PASS, 신규 회귀 없음.
- `golangci-lint run ./internal/cli/...` 0 NEW issues.

### 5.4 Cross-platform

- darwin/amd64, darwin/arm64, linux/amd64, windows/amd64 빌드 정합.
- timestamp 백업 디렉터리명에 ISO8601 형식 사용으로 cross-platform 안전 (콜론 미사용).

---

## 6. Risks (요약, 상세는 spec.md §E)

- **R1**: BC-V3R3-007 idempotency 침해 → REQ-UAC-006 회귀 테스트.
- **R2**: `--force` 의미 분절화 → §4.1 의미 표 SSOT 정합.
- **R3**: 테스트 flakiness (timestamp 의존) → glob 패턴 검증 또는 injectable seam.
- **R4**: 사용자 customization 의도 파괴 → REQ-UAC-003 자동 백업.
- **R5**: skip-sync 단락 Option A vs B 선택 → §4.3 Option A 채택.

---

## 7. 후속 SPEC 후보 (Out of Scope)

- **SPEC-V3R6-UPDATE-NOISE-001**: reserved filename / 3-way merge 노이즈 정리.
- **SPEC-V3R6-UPDATE-PROGRESS-001**: spinner / progress bar 출력 손상 정리.
- **SPEC-V3R6-UPDATE-RESTORE-001** (예약): `moai migrate restore-archive-drift` 명령으로 백업 디렉터리 복원 UX.

---

## 8. 머지 후 작업

- `progress.md` 작성 (M1~M6 증거).
- `.moai/specs/SPEC-V3R6-UPDATE-ARCHIVE-CONTRACT-001/` status `draft` → `implemented`.
- 후속 SPEC 후보 (§7) 의 plan-phase 진입 여부 결정.
- Wave 1 (skill consolidation) 진행 중이므로 본 SPEC 은 Wave 외 정합 작업으로 분류 (cross-Wave caveat — 사용자 동의 후 진입).

---

## Out of Scope

### Out of Scope — 본 plan 범위 외 작업

- SPEC-V3R6-UPDATE-NOISE-001 — reserved filename ack ledger + 3-strike merge-history (별도 SPEC)
- SPEC-V3R6-UPDATE-PROGRESS-001 — `\r` overwrite 출력 깨짐 정정 (별도 SPEC)
- SPEC-V3R6-UPDATE-RESTORE-001 (후보) — `moai migrate restore-archive-drift` UX (drift 백업 복원)
- BC-V3R3-007 archive contract 자체의 v2.16 → v2.17 마이그레이션 (legacy skill ID 목록 변경)
- Design folder + evolution dir 등 다른 archive 서브시스템
