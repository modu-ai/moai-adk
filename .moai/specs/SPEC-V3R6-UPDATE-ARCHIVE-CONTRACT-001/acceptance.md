---
id: SPEC-V3R6-UPDATE-ARCHIVE-CONTRACT-001
title: "SPEC-V3R6-UPDATE-ARCHIVE-CONTRACT-001 — Acceptance Criteria"
version: "0.1.0"
status: draft
created: 2026-05-23
updated: 2026-05-23
author: manager-spec
priority: P1
phase: "v3.6.0"
module: "internal/cli"
lifecycle: spec-anchored
tags: "v3r6, update, archive, force, skip-sync, contract, acceptance"
tier: M
---

# Acceptance Criteria — SPEC-V3R6-UPDATE-ARCHIVE-CONTRACT-001

본 문서는 SPEC 의 binary 수용 기준 10건을 정의한다. 각 AC 는 PASS/FAIL 로 판정되며 검증 커맨드 또는 Given-When-Then 시나리오를 동반한다. AC 와 REQ 간 traceability 는 §3 트레이스 표 참조.

---

## 1. Acceptance Criteria

### AC-UAC-001 — `archiveLegacySkills` 시그니처 확장 검증

**Given** SPEC 구현 코드가 머지된 main HEAD.
**When** 다음 검증 커맨드를 실행한다:

```bash
grep -n 'func archiveLegacySkills' internal/cli/update_archive.go
```

**Then** 출력에 `func archiveLegacySkills(projectRoot string, out io.Writer, force bool) (int, error)` 가 정확히 1회 포함되어야 한다. force 파라미터 부재 시그니처는 검출되지 않아야 한다.

**Trace**: REQ-UAC-001

---

### AC-UAC-002 — `runUpdate` 호출부의 `--force` 전파 검증

**Given** SPEC 구현 코드가 머지된 main HEAD.
**When** 다음 검증 커맨드를 실행한다:

```bash
grep -rn 'archiveLegacySkills(' internal/cli/ | grep -v '_test.go'
```

**Then** `runUpdate` 호출 사이트가 `archiveLegacySkills(cwd, out, getBoolFlag(cmd, "force"))` 또는 동등한 force 전달 패턴을 포함해야 한다. force 인자 없이 호출하는 라인이 production 코드에 잔존하지 않아야 한다.

**Trace**: REQ-UAC-002

---

### AC-UAC-003 — `TestArchiveForce` force=true drift 해소 검증

**Given** SPEC 구현 코드가 머지된 main HEAD.
**When** 다음 테스트를 실행한다:

```bash
go test -run TestArchiveForce ./internal/cli/ -count=1 -v
```

**Then** subtest `force_with_drift_creates_backup_and_overwrites` 가 PASS 해야 한다.
- 사전 조건: `t.TempDir()` 에 src + 의도적으로 다른 archive 디렉터리 생성.
- 실행 조건: `archiveLegacySkills(tmp, &out, true)` 호출.
- 사후 조건: archive 디렉터리 내용이 src 와 일치 (`computeDirHashes` 비교).

**Trace**: REQ-UAC-003, REQ-UAC-007

---

### AC-UAC-004 — Drift 백업 디렉터리 생성 검증

**Given** `TestArchiveForce` 의 force overwrite 시나리오.
**When** subtest `force_with_drift_creates_backup_and_overwrites` 가 실행된다:

```bash
go test -run TestArchiveForce/force_with_drift_creates_backup_and_overwrites ./internal/cli/ -v
```

**Then** 임시 디렉터리 내에 `archive/skills/v2.16-drift-<timestamp>/<skill-id>/` 패턴의 백업 디렉터리가 정확히 1개 생성되어야 하며, 백업 디렉터리 내용은 overwrite 이전의 archive 와 일치해야 한다. timestamp 는 ISO8601 UTC (`YYYYMMDDTHHMMSSZ`) 형식이어야 한다.

검증 보조 커맨드 (테스트 본체 내):

```go
matches, _ := filepath.Glob(filepath.Join(tmp, ".moai", "archive", "skills", "v2.16-drift-*"))
require.Len(t, matches, 1)
```

**Trace**: REQ-UAC-003

---

### AC-UAC-005 — `TestSkipSyncNoArchive` skip-sync 단락 검증

**Given** SPEC 구현 코드가 머지된 main HEAD.
**When** 다음 테스트를 실행한다:

```bash
go test -run TestSkipSyncNoArchive ./internal/cli/ -count=1 -v
```

**Then** 다음 시나리오가 PASS 해야 한다:
- 사전 조건: `t.TempDir()` 에 임시 프로젝트 + `.moai/config/config.yaml` 의 template_version 을 binary version 과 일치하도록 설정 + archive 디렉터리에 의도적 drift 도입.
- 실행 조건: `--force` 미지정 상태로 update 흐름 실행.
- 사후 조건: skip-sync 분기 활성화 + archive 호출이 일어나지 않음 (ARCHIVE_DRIFT 에러 출력 없음).

별도 subtest `skip_sync_with_force_does_invoke_archive` 로 `--force` 명시 시 sync + archive 정상 실행 회귀 검증.

**Trace**: REQ-UAC-004, REQ-UAC-007

---

### AC-UAC-006 — BC-V3R3-007 Idempotency 회귀 보호 검증

**Given** SPEC 구현 코드가 머지된 main HEAD.
**When** 기존 archive 회귀 테스트를 실행한다:

```bash
go test -run 'TestArchive|TestIdempotency' ./internal/cli/ -count=1 -v
```

**Then** 기존 force=false 경로의 모든 회귀 테스트가 PASS 해야 한다. 구체적으로:
- source absent → return nil (idempotent)
- archive matches → return nil (idempotent)
- archive differs → return ARCHIVE_DRIFT (기존 컨트랙트 보존)

`update_archive_test.go`, `update_archive_flow_test.go`, `update_idempotency_test.go` 의 PASS 율 100%.

**Trace**: REQ-UAC-006

---

### AC-UAC-007 — 에러 메시지와 코드 경로 정합 검증

**Given** SPEC 구현 코드가 머지된 main HEAD.
**When** 다음 검증 커맨드를 실행한다:

```bash
git grep -n 'Use --force to overwrite' internal/
```

**Then** 출력에 등장하는 메시지가 실제 force 경로 구현과 정합해야 한다. 구체적으로:
- 메시지가 등장하는 모든 위치는 `update_archive.go` 의 `checkArchiveDrift` ARCHIVE_DRIFT 에러 메시지 (`update_archive.go:138-145` + `:151-158`) 에 한정.
- 본 SPEC 구현 후 `archiveLegacySkills(force=true)` 코드 경로가 실제로 overwrite 를 수행함을 코드 리뷰로 검증.
- 메시지가 거짓 안내가 아님을 보증.

**Trace**: REQ-UAC-005

---

### AC-UAC-008 — `~/moai/mo.ai.kr` 수동 reproducer 검증

**Given** SPEC 구현 코드가 머지된 main HEAD + `~/moai/mo.ai.kr` 사용자 프로젝트가 `~/moai/mo.ai.kr/.claude/skills/moai-domain-backend/SKILL.md` 와 `~/moai/mo.ai.kr/.moai/archive/skills/v2.16/moai-domain-backend/SKILL.md` 가 `diff -q` 기준 differ 인 상태.

**When** 사용자가 다음 커맨드를 실행한다:

```bash
cd ~/moai/mo.ai.kr
moai update --force
```

**Then** 다음을 만족해야 한다:
- `"Legacy skill archive failed"` 메시지가 출력에 등장하지 않음.
- `"ARCHIVE_DRIFT"` 코드가 출력에 등장하지 않음.
- 실행 후 `diff -q ~/moai/mo.ai.kr/.claude/skills/moai-domain-backend/SKILL.md ~/moai/mo.ai.kr/.moai/archive/skills/v2.16/moai-domain-backend/SKILL.md` 가 일치 (출력 없음).
- `ls -d ~/moai/mo.ai.kr/.moai/archive/skills/v2.16-drift-*/` 가 정확히 1개 백업 디렉터리를 표시.
- 백업 디렉터리는 overwrite 직전의 archive 내용을 보존.

**Trace**: REQ-UAC-002, REQ-UAC-003

---

### AC-UAC-009 — Help 텍스트 + godoc 정합 검증

**Given** SPEC 구현 코드가 머지된 main HEAD.
**When** 다음 검증 커맨드를 실행한다:

```bash
grep -n '"force"' internal/cli/update.go
grep -n '\-\-force:' internal/cli/update.go
```

**Then**:
- `update.go:85` 의 help 텍스트가 archive overwrite 의미를 명시적으로 포함해야 한다 (예: `"...also forces archive drift overwrite"`).
- `update.go:108` 인근 godoc 주석이 archive 의미와 백업 디렉터리 경로 (`.moai/archive/skills/v2.16-drift-<timestamp>/`) 를 언급해야 한다.
- godoc 주석의 `--force` 의미 표가 spec.md §4.1 (의미 표 SSOT) 와 일치해야 한다.

**Trace**: REQ-UAC-002, REQ-UAC-005

---

### AC-UAC-010 — 테스트 격리 + dev project 무손실 검증

**Given** SPEC 구현 코드가 머지된 main HEAD.
**When** 다음 검증 커맨드를 실행한다:

```bash
grep -n 't.TempDir()' internal/cli/update_archive_force_test.go internal/cli/update_skip_sync_test.go
grep -rn 'TestArchiveForce\|TestSkipSyncNoArchive' internal/cli/ | grep -v _test.go
```

**Then**:
- 신규 테스트 파일에 `t.TempDir()` 호출이 각 테스트 함수 진입부에 존재해야 한다.
- 테스트는 `dev project` 의 `.claude/`, `.moai/archive/`, `.moai/specs/` 어떤 경로도 직접 수정하지 않아야 한다 (Read/Glob 만 허용, Write 절대 금지).
- 두 번째 grep 명령은 production 코드 호출이 없음을 확인 (테스트만 사용).
- `go test -race ./internal/cli/...` 실행 시 race condition 없음.

**Trace**: REQ-UAC-007 + CLAUDE.local.md §6 테스트 격리 원칙

---

## 2. Definition of Done

본 SPEC 의 Definition of Done 은 다음 모든 조건의 동시 충족이다:

1. AC-UAC-001 ~ AC-UAC-010 의 10개 AC 모두 PASS.
2. `go test ./internal/cli/... -count=1` 전체 PASS, 0 NEW failures vs main baseline.
3. `golangci-lint run ./internal/cli/...` 0 NEW issues.
4. `go vet ./...` PASS.
5. Cross-platform 빌드 PASS (darwin/amd64, darwin/arm64, linux/amd64, windows/amd64).
6. C-HRA-008 sentinel grep clean (`grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/ | grep -v _test.go | grep -v "^[^:]*:[0-9]*:[ \t]*//"` 결과 0 라인).
7. `progress.md` 작성 완료 (M1~M6 evidence + AC 결과 표).
8. `spec.md` / `plan.md` / `acceptance.md` 의 `status` 가 `draft` → `implemented` 갱신.
9. `version` 이 `0.1.0` → `0.2.0` 갱신.
10. PR description 또는 commit message 에 `~/moai/mo.ai.kr` reproducer 의 before/after `diff -q` 출력 첨부.

---

## 3. Traceability Matrix

각 AC ↔ REQ ↔ Plan 마일스톤 매핑. 100% 양방향 traceability 보장.

| AC ID | REQ ID | Plan 마일스톤 | 검증 방식 |
|-------|--------|---------------|-----------|
| AC-UAC-001 | REQ-UAC-001 | M1 | grep |
| AC-UAC-002 | REQ-UAC-002 | M2 | grep |
| AC-UAC-003 | REQ-UAC-003, REQ-UAC-007 | M1, M3 | go test |
| AC-UAC-004 | REQ-UAC-003 | M1, M3 | go test (glob) |
| AC-UAC-005 | REQ-UAC-004, REQ-UAC-007 | M2, M3 | go test |
| AC-UAC-006 | REQ-UAC-006 | M4 | go test (회귀) |
| AC-UAC-007 | REQ-UAC-005 | M2 | git grep + 코드 리뷰 |
| AC-UAC-008 | REQ-UAC-002, REQ-UAC-003 | M6 | 수동 reproducer |
| AC-UAC-009 | REQ-UAC-002, REQ-UAC-005 | M2 | grep + 코드 리뷰 |
| AC-UAC-010 | REQ-UAC-007 | M3 | grep + race 검증 |

REQ 의 100% 커버리지 확인:
- REQ-UAC-001 → AC-UAC-001 ✓
- REQ-UAC-002 → AC-UAC-002, AC-UAC-008, AC-UAC-009 ✓
- REQ-UAC-003 → AC-UAC-003, AC-UAC-004, AC-UAC-008 ✓
- REQ-UAC-004 → AC-UAC-005 ✓
- REQ-UAC-005 → AC-UAC-007, AC-UAC-009 ✓
- REQ-UAC-006 → AC-UAC-006 ✓
- REQ-UAC-007 → AC-UAC-003, AC-UAC-005, AC-UAC-010 ✓

모든 REQ 가 최소 1개 AC 에 매핑되며, 모든 AC 가 최소 1개 REQ 에 매핑됨.

---

## 4. Out of Scope

### 4.1 Out of Scope — 본 acceptance 범위 외 시나리오 (미검증)

다음은 본 SPEC 의 acceptance 대상이 아니며 후속 SPEC 에서 다룬다:

- v2.16 → v2.17 archive version 마이그레이션.
- `legacySkillIDs` 목록 변경.
- `moai migrate restore-archive-drift` (백업 디렉터리 복원 UX) — SPEC-V3R6-UPDATE-RESTORE-001 후보.
- Spinner / progress bar 출력 손상 — SPEC-V3R6-UPDATE-PROGRESS-001 후보.
- Reserved filename / 3-way merge 노이즈 — SPEC-V3R6-UPDATE-NOISE-001 후보.
- Design folder, evolution dir 등 다른 archive 서브시스템 컨트랙트.
