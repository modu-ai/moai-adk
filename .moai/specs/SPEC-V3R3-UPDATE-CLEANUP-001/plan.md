---
id: SPEC-V3R3-UPDATE-CLEANUP-001
version: "0.2.2"
status: draft
created_at: 2026-05-01
updated_at: 2026-05-02
author: manager-spec
priority: High
labels: [cli, update, deployment, cleanup, agency, idempotency]
---

# SPEC-V3R3-UPDATE-CLEANUP-001 — Implementation Plan

## HISTORY

- 2026-05-02 v0.2.2: audit v2 minor patch — D-02-11 LOC 변천 표기를 "estimation refinement, not scope change" 프레이밍으로 명시 (각 버전 업데이트가 새 작업 추가가 아닌 분석 정확도 향상의 결과임을 reader에게 명확히). spec/acceptance와 함께 동기 업데이트.
- 2026-05-01 v0.2.1: audit v2 remediation — D-02-04 LOC reconcile 산식 명시화 (helper 재사용 항목별 deduplication LOC 산출), D-02-05 NFR-P1 benchstat 통계 게이트 (T4.4 갱신), D-02-06 OQ3 해결 반영 (T3.2 인라인), D-02-01 REQ-UPC-018 (`.moai-skip-cleanup` opt-out)을 M2에 신규 task로 추가, D-02-02 `.moai/logs/` self-reference 보호 (T3.6 확장), D-02-03 telemetry permission 처리 (T3.4 확장), D-02-07 case probe 엣지 케이스 (T4.3 확장), D-02-08 symlink 엣지 케이스 (T3.5 확장). M2 LOC 75→약 85, M3 LOC 130→약 145, M4 LOC 150→약 165 (총 ~395 → reconciled 산식 후 ~380 유지).
- 2026-05-01 v0.2.0: 감사 후속 개정 — D-01 LOC reconcile (~95 → ~380으로 일관성 확보), OQ1/OQ2 해결 반영 (T3.1/T3.4 갱신), 신규 REQ-UPC-022~026에 대한 마일스톤 작업 추가 (M3에 telemetry/symlink/backup-self-reference, M4에 case-insensitive E2E 추가), Risk Register에 R6/R7/R8 신규 추가 (symlink, case-insensitive FS, backup self-reference). M3 LOC 90→약 130으로 증가, M4 LOC 120→약 150으로 증가, NFR-UPC-P1 벤치마크 작업을 M4에 명시 추가.
- 2026-05-01 v0.1.0: 최초 작성. 4개 마일스톤(M1~M4) 시퀀셜 진행 결정.

---

## 1. Approach Overview

본 SPEC은 두 결함(멱등성 결함 + 폐기 경로 누수)을 단일 release cycle에 번들링한다. 두 결함 모두 `internal/cli/update.go` + deployer 체인을 건드리며, 회귀 테스트 인프라(`t.TempDir()` 기반 사용자 프로젝트 시뮬레이션)를 공유한다. 마일스톤은 의존성 그래프상 **M1 → M2 → M3 → M4 시퀀셜** 진행이 안전하다 (M1의 atomic write가 M2의 backup 안전성 토대를 제공, M3의 confirmation flow가 M2의 deletion 트리거이므로 통합 테스트 시점에 모두 필요).

**총 LOC 추정 ~380** (spec.md §1.3과 일치). v0.2.1에서 산식 명시화 (D-02-04). **이 추정치 변천(v0.1.0 ~95 → v0.2.0 ~380 → v0.2.1 ~395 simple sum / 380 reconciled)은 estimation refinement이며 SPEC scope 변경이 아니다 (v0.2.2 D-02-11 명시화)** — 각 버전 개정은 분석 깊이를 증가시킨 결과이며, 신규 기능 추가가 아니다. v0.2.0에서 atomic write + manifest provenance + benchstat 인프라가 SPEC scope 분석에 포함되었고, v0.2.1에서 helper deduplication 산식이 정량화되어 ~380 LOC reconciled 값을 도출했다.

**1단계 — 마일스톤별 단순 합산 (helper 중복 미고려)**:
- M1 (atomic write + lock + idempotency tests): ~95 LOC
- M2 (deprecated cleanup + backup + manifest + `.moai-skip-cleanup` opt-out): ~85 LOC (v0.2.0의 75에서 +10 LOC for REQ-UPC-018 marker scan)
- M3 (confirmation + provenance + telemetry + symlink edges + backup self-reference + `.moai/logs/` 보호 + telemetry permission): ~145 LOC (v0.2.0의 130에서 +15 LOC for D-02-02/D-02-03/D-02-08 확장)
- M4 (E2E regression 6 scenarios + case probe edges + benchstat NFR): ~165 LOC (v0.2.0의 150에서 +15 LOC for D-02-05/D-02-07 확장)
- **단순 합산: 95 + 85 + 145 + 165 = 490 LOC**

**2단계 — Helper 재사용 deduplication 산식 (v0.2.1 명시)**:

| Helper | 정의 마일스톤 | 재사용 마일스톤 | Dedup LOC |
|---|---|---|---|
| `atomicWriteFile()` (write-to-tmp + rename wrapper) | M1 | M3 manifest update path 재사용 | -15 |
| `backupTreeWithManifest()` (recursive copy + MANIFEST emit) | M2 | M3 confirmation flow 재사용 | -25 |
| `classifyDeprecatedFile()` (manifest provenance lookup) | M3 | M2 deprecated path detection 재사용 | -30 |
| `caseSensitivityProbe()` + path matcher | M4 (probe) | M2/M3 path matching 재사용 | -20 |
| Test fixtures (`setupSyntheticProject`, `setupUserProjectWithAgency`) | M4 | M1/M2/M3 unit tests에서 부분 재사용 | -20 |
| **Total deduplication** | | | **-110** |

**3단계 — Reconciled 최종 견적**:
- 490 (단순 합산) − 110 (helper dedup) = **380 LOC** (코드 + 테스트, reconciled).
- 산식 검증 가능성: 각 helper 정의는 plan.md의 Tasks 섹션에서 식별 가능하며, 재사용 위치는 M2~M4의 Tasks에 명시됨. Reader는 위 표를 통해 산식 plausibility를 직접 검증 가능.

---

## 2. Milestones

### M1 — Atomic Write + Idempotency Tests

**목표:** Deployer를 atomic write 패턴으로 강건화하고, 멱등성 회귀 테스트를 추가한다.

**Tasks:**

- T1.1: `internal/template/deployer.go:168` (`os.WriteFile` 호출부)를 atomic pattern으로 교체:
  - Write to `<destPath>.moai-tmp` 우선
  - `os.Rename(tmpPath, destPath)`로 원자적 이동
  - 실패 시 `os.Remove(tmpPath)` defer
- T1.2: `internal/template/deployer_mode.go:128` (transactional mode equivalent)도 동일 패턴 적용
- T1.3: `.moai/.update.lock` 파일 기반 동시 실행 방지 (REQ-UPC-005):
  - Lock 파일 생성 위치: 프로젝트 루트의 `.moai/.update.lock`
  - PID + timestamp + hostname을 JSON으로 기록
  - signal handler (SIGINT/SIGTERM)로 정리 보장
- T1.4: `internal/template/deployer_test.go` 신규 테스트:
  - `TestDeployer_Idempotent` — 동일 destination에 Deploy 두 번 호출 후 byte-identical 검증
  - `TestDeployer_NoTmpResidue` — 성공 후 `.moai-tmp` 파일 부재 검증
  - `TestDeployer_TmpCleanupOnFailure` — Rename 실패 시뮬레이션 후 `.moai-tmp` 정리 검증
  - `TestUpdate_ConcurrentLock` — 두 번째 `moai update` 호출이 lock으로 거부되는지 검증

**Files Modified:**

- `/Users/goos/MoAI/moai-adk-go/internal/template/deployer.go` (~15 LOC)
- `/Users/goos/MoAI/moai-adk-go/internal/template/deployer_mode.go` (~10 LOC)
- `/Users/goos/MoAI/moai-adk-go/internal/cli/update.go` (~20 LOC, lock file logic)
- `/Users/goos/MoAI/moai-adk-go/internal/template/deployer_test.go` (~50 LOC, 4 new tests)

**LOC Estimate:** ~45 LOC code + ~50 LOC tests = ~95 LOC

**Dependencies:** None (independent of M2~M4)

**Definition of Done:**
- [ ] `go test ./internal/template/...` 통과
- [ ] `TestDeployer_Idempotent` 두 번 연속 실행해도 ` 2` 접미어 파일 미생성 검증
- [ ] Lock file이 정상/SIGINT 종료 모두에서 정리됨

---

### M2 — Deprecated Path Cleanup + Backup

**목표:** `cleanMoaiManagedPaths`를 확장하여 폐기된 agency 경로를 탐지하고, 안전 백업 후 삭제한다.

**Tasks:**

- T2.1: `internal/defs/dirs.go`에 `DeprecatedPaths` 상수 정의:
  - `.claude/commands/agency/agency.md`
  - `.claude/commands/agency/brief.md`
  - `.claude/commands/agency/build.md`
  - `.claude/commands/agency/evolve.md`
  - `.claude/commands/agency/learn.md`
  - `.claude/commands/agency/profile.md`
  - `.claude/commands/agency/resume.md`
  - `.claude/commands/agency/review.md`
  - `.claude/rules/agency/constitution.md`
  - 각 항목에 `DeprecatedSince`, `DeprecatedBy` (SPEC ID), `RemovalSchedule` 메타데이터 첨부
- T2.2: `internal/cli/update.go:1411-1441`의 `cleanMoaiManagedPaths` 확장:
  - DeprecatedPaths 스캔 phase 추가
  - 발견된 파일 리스트 수집
  - REQ-UPC-025 처리: deprecated 등록되었으나 파일 부재 시 silent skip
- T2.3: 백업 함수 `backupDeprecatedPaths()` 신규 추가:
  - 백업 경로: `.moai/backup/agency-{ISO8601}/`
  - 디렉터리 구조 보존 (`cp -R` 동등 동작)
  - `MANIFEST.json` 생성 (REQ-UPC-011)
- T2.4: `internal/cli/update_test.go` 신규 테스트:
  - `TestCleanup_AgencyPathsDetected` — agency 파일 존재 시 탐지
  - `TestCleanup_NoOpWhenAbsent` — agency 부재 시 백업 미생성
  - `TestCleanup_DeletedFileSkipped` — registry에는 있으나 파일 부재 시 silent (REQ-UPC-025)
- T2.5: **REQ-UPC-018 `.moai-skip-cleanup` opt-out marker (v0.2.1 신규)**:
  - 헬퍼 함수 `hasSkipCleanupMarker(dir string) bool` — 디렉터리 내 `.moai-skip-cleanup` 파일 존재 여부 검사
  - cleanup scan 진입 직후 각 deprecated path의 부모 디렉터리에서 marker 검사
  - marker 발견 시: 해당 디렉터리 전체 skip + stderr `[INFO] cleanup skipped due to .moai-skip-cleanup marker: <path>` 출력
  - skip된 paths 목록을 `cleanupResult.UserOptOutPaths` 필드에 수집 → REQ-UPC-022 telemetry에 포함
  - 단위 테스트: `TestCleanup_SkipMarkerHonored` (marker 존재 시 skip), `TestCleanup_SkipMarkerAbsent` (marker 부재 시 정상 진행)

**Files Modified:**

- `/Users/goos/MoAI/moai-adk-go/internal/defs/dirs.go` (~15 LOC)
- `/Users/goos/MoAI/moai-adk-go/internal/cli/update.go` (~40 LOC, cleanup + backup + skip marker)
- `/Users/goos/MoAI/moai-adk-go/internal/cli/update_test.go` (~40 LOC, 5 new tests)

**LOC Estimate:** ~55 LOC code + ~40 LOC tests = ~95 LOC (단순 합산); helper dedup 후 ~85 LOC

**Dependencies:** M1 (lock file 인프라 활용)

**Definition of Done:**
- [ ] DeprecatedPaths가 `internal/defs/dirs.go`에 single source of truth로 정의
- [ ] 백업 디렉터리에 MANIFEST.json + 원본 파일 트리 보존 검증
- [ ] 백업 실패 시 삭제 phase 진입 차단 검증
- [ ] REQ-UPC-025: 등록은 되었으나 파일 부재 시 silent 검증
- [ ] REQ-UPC-018: `.moai-skip-cleanup` marker 동작 검증 (skip + INFO log + telemetry 필드)

---

### M3 — User Confirmation + Manifest Provenance + Telemetry + Symlink Safety

**목표:** 사용자 확인 흐름 통합, 기존 매니페스트 provenance(`deployed_hash`)로 삼분 분류 (Pristine/UserModified/UnverifiedDeprecated), 텔레메트리 로깅, symlink 안전 처리, backup self-reference 차단.

**OQ1/OQ2 해결 반영:**
- `merge.ConfirmMerge` 시그니처 = `func ConfirmMerge(analysis MergeAnalysis) (bool, error)` (`internal/merge/confirm.go:623`). risk classification은 `MergeAnalysis.RiskLevel`("low"/"medium"/"high") + 각 `FileAnalysis.RiskLevel` 필드로 전달. 신규 wrapper 불필요.
- Provenance hash는 기존 `manifest.FileEntry.DeployedHash` 활용 (`internal/manifest/types.go:52`). 별도 사이드카 또는 schema 마이그레이션 불필요. `manifest.Manager.GetEntry(path)` + `manifest.HashFile(absPath)` 비교로 분류.

**Tasks:**

- T3.1: `internal/cli/update.go`에서 `merge.ConfirmMerge` 호출:
  - `MergeAnalysis{RiskLevel: "high", HasConflicts: true, Files: [...]}` 구성
  - 각 파일에 대해 `FileAnalysis{Path, RiskLevel: "high", Note: classification, Strategy: "backup-then-delete"}` 구성
  - 사용자에게 폐기 경로 리스트 + 백업 위치 + SPEC ID 표시
  - 확인/거부/auto-confirm 분기
- T3.2: `--auto-confirm-cleanup` CLI 플래그 추가 (**OQ3 resolved v0.2.1**: 플래그 명 = `--auto-confirm-cleanup`, default `false`, `PristineDeprecated`만 prompt bypass; `UserModifiedDeprecated` / `UnverifiedDeprecated`는 ALWAYS confirmation 필요). 별도 `--cleanup-policy` 통합 옵션은 채택하지 않음 (UX 단순성 우선).
- T3.3: 매니페스트 provenance check (기존 `manifest` 패키지 활용):
  - 각 deprecated 파일 경로에 대해 `manager.GetEntry(path)` 호출
  - entry 존재 + `entry.DeployedHash == manifest.HashFile(absPath)` → `PristineDeprecated`
  - entry 존재 + hash 불일치 → `UserModifiedDeprecated`
  - entry 부재 → `UnverifiedDeprecated`
  - 분류 결과를 backup `MANIFEST.json`의 entry당 `classification` 필드에 기록
- T3.4: **REQ-UPC-022 텔레메트리** — cleanup phase 종료 시점에 JSON line 이벤트 emit:
  - 헬퍼 함수 `emitCleanupTelemetry(event CleanupEvent)` 신규
  - `.moai/logs/update-cleanup-{ISO8601}.jsonl` append + stderr mirror
  - 필드: `atomic_write_used`, `pre_update_suffix2_files`, `backup_outcome`, `backup_path`, `cleanup_outcome`, `user_opt_out_paths` (REQ-UPC-018 marker로 skip된 경로)
  - cleanup 시작 시 `find . -name "* 2.*"` 동등 스캔으로 ` 2` 접미어 파일 사전 캡처
  - **D-02-03 permission denied 처리 (v0.2.1)**: `.moai/logs/` 디렉터리 생성 또는 write 실패 시 (e.g., chmod 0500) — `[WARN] telemetry persistence skipped: <reason>` stderr 출력 + file persistence skip + update 작업은 정상 계속. stderr mirror는 여전히 emit (관측성 보존).
- T3.5: **REQ-UPC-023 symlink 안전** — backup/delete 직전 `os.Lstat`로 symlink 감지:
  - symlink일 경우: backup은 link target 경로 기록만 (`MANIFEST.json`의 `symlink_target` 필드), 실제 따라가지 않음
  - 삭제는 `os.Remove`로 link 자체만 제거 (target은 외부 위치이므로 보존)
  - **D-02-08 엣지 케이스 (v0.2.1)**:
    - **Broken symlink**: `os.Stat(linkTarget)` ENOENT 시 `MANIFEST.json` entry에 `symlink_target_status: "broken"` 기록, link 자체만 제거 (target 백업 시도 안 함)
    - **Target collision**: link target 자체가 다른 deprecated entry와 매칭되는 경우, link와 target을 독립 처리. target은 자체 entry로 단 1회만 backup (double-count 방지)
    - **Windows junction/reparse point**: POSIX symlink와 동일하게 처리. symlink 감지 미지원 환경(예: 관리자 권한 부재)에서는 `[INFO] symlink detection unsupported on this platform` 출력 + 일반 파일로 처리
- T3.6: **REQ-UPC-024 backup self-reference + logs self-reference (v0.2.1 D-02-02 확장)** — deprecated 스캔 함수에서 `.moai/backup/` AND `.moai/logs/` 접두사 path는 모두 무조건 skip. 단위 테스트로 두 경로 모두 검증 (`TestCleanup_BackupSelfReferenceSkipped` + `TestCleanup_LogsSelfReferenceSkipped`).
- T3.7: `internal/cli/update_test.go` 추가 테스트:
  - `TestCleanup_UserConfirmationDeclined` — 확인 거부 시 미삭제
  - `TestCleanup_AutoConfirmFlag` — 플래그로 prompt bypass (Pristine만)
  - `TestCleanup_PristineDeprecated` — 동일 hash → 자동 삭제 가능
  - `TestCleanup_UserModifiedDeprecated` — 다른 hash → 백업 + 경고
  - `TestCleanup_UnverifiedDeprecated` — manifest entry 없음 → auto-confirm 차단 + 경고
  - `TestCleanup_TelemetryEmitted` — JSON line 로그 파일 생성 검증 (user_opt_out_paths 필드 포함)
  - `TestCleanup_TelemetryPermissionDenied` (v0.2.1, D-02-03) — `.moai/logs/` chmod 0500 → WARN 출력 + update 정상 종료 + .jsonl 파일 부재
  - `TestCleanup_SymlinkNotFollowed` — symlink는 target follow 안 함, link만 삭제
  - `TestCleanup_SymlinkBroken` (v0.2.1) — broken symlink 처리 + symlink_target_status: "broken" 기록
  - `TestCleanup_SymlinkTargetCollision` (v0.2.1) — target이 별도 deprecated entry인 경우 단 1회 backup
  - `TestCleanup_BackupSelfReferenceSkipped` — `.moai/backup/` 하위는 스캔 대상 아님
  - `TestCleanup_LogsSelfReferenceSkipped` (v0.2.1, D-02-02) — `.moai/logs/` 하위는 스캔 대상 아님

**Files Modified:**

- `/Users/goos/MoAI/moai-adk-go/internal/cli/update.go` (~60 LOC, confirmation + provenance + telemetry + symlink edges + self-ref + permission)
- `/Users/goos/MoAI/moai-adk-go/internal/cli/flags.go` 또는 동등 위치 (~5 LOC, 플래그)
- `/Users/goos/MoAI/moai-adk-go/internal/cli/update_test.go` (~95 LOC, 12 new tests)

**LOC Estimate:** ~65 LOC code + ~95 LOC tests = ~160 LOC (단순 합산); helper dedup 후 ~145 LOC

**Dependencies:** M2 (cleanup 함수 + 백업 인프라 + skip marker 헬퍼 필요)

**Definition of Done:**
- [ ] `merge.ConfirmMerge` 호출이 high-risk 분류로 표시됨 (signature 일치 검증)
- [ ] `manifest.FileEntry.DeployedHash` 기반 삼분 분류 동작 (Pristine/UserModified/Unverified)
- [ ] `--auto-confirm-cleanup` 플래그 (OQ3 resolved): default false, Pristine만 bypass, UnverifiedDeprecated/UserModifiedDeprecated는 ALWAYS confirmation
- [ ] REQ-UPC-022 telemetry: `.moai/logs/update-cleanup-*.jsonl` 생성 + 6개 필드 검증 (`user_opt_out_paths` 포함)
- [ ] REQ-UPC-022 permission denied (v0.2.1): chmod 0500 환경에서 WARN + update 정상 종료
- [ ] REQ-UPC-023 symlink: link follow 차단 + `symlink_target` MANIFEST 기록
- [ ] REQ-UPC-023 broken symlink (v0.2.1): `symlink_target_status: "broken"` 기록
- [ ] REQ-UPC-023 target collision (v0.2.1): 단 1회 backup 검증
- [ ] REQ-UPC-024 backup self-reference: `.moai/backup/` 스캔 제외
- [ ] REQ-UPC-024 logs self-reference (v0.2.1): `.moai/logs/` 스캔 제외

---

### M4 — Regression Test Suite + Case-Insensitive FS + NFR Benchmark

**목표:** End-to-end 회귀 시나리오 A–E를 자동화 테스트로 고정하고, 케이스 무관 FS 처리(REQ-UPC-026)와 NFR-UPC-P1 벤치마크 게이트를 추가한다.

**Tasks:**

- T4.1: `internal/cli/update_e2e_test.go` 신규 파일 생성 (또는 기존 update_test.go 확장):
  - 시나리오 A: 사용자 프로젝트에 agency/ 존재 → moai update → 백업 + 삭제 (Pristine)
  - 시나리오 B: agency/ 부재 → moai update → no-op, 백업 미생성
  - 시나리오 C: 사용자가 agency.md 수정 → moai update → 백업 + 경고 (UserModified)
  - 시나리오 D: moai update 두 번 연속 → ` 2` 접미어 파일 미발생
  - 시나리오 E: 동시 moai update → lock 차단 (deterministic lock acquisition test, not timing-based race)
  - 추가 시나리오 F: manifest entry 부재 → UnverifiedDeprecated 경로 (REQ-UPC-015c)
- T4.2: 테스트 헬퍼 작성:
  - `setupUserProjectWithAgency(t)` — agency 파일 시뮬레이션
  - `setupManifestWithProvenance(t, hash)` — provenance 매니페스트 작성 (`manifest.Manager.Track` 활용)
  - `setupUserProjectWithoutManifest(t)` — Unverified 케이스용
- T4.3: **REQ-UPC-026 케이스 무관 FS 테스트** (`TestCleanup_CaseInsensitiveFS`):
  - case-insensitive FS probe 헬퍼 (`.moai-fscase-probe` 생성 → 대문자 stat → 정리)
  - macOS APFS 환경에서 `Agency.md` (대문자)도 deprecated registry의 `agency.md`와 매칭됨을 검증
  - Linux ext4 환경(case-sensitive)에서는 매칭 안 됨을 검증
  - 두 환경 모두 CI에서 자동 실행
  - **D-02-07 엣지 케이스 (v0.2.1)**:
    - **Probe failure fallback** (`TestCleanup_CaseProbeFailureFallback`): read-only filesystem 시뮬레이션 → probe 실패 → case-sensitive 매칭으로 fallback + INFO 출력 검증
    - **Probe frequency** (`TestCleanup_CaseProbeRunsOncePerInvocation`): 단일 invocation 내 probe 호출 정확히 1회 (caching 금지) — counter mock으로 검증
    - **APFS case-sensitive variant volume** (`TestCleanup_AfpsCaseSensitiveVariant`): macOS에서도 probe가 case-sensitive로 판정 시 case-sensitive 매칭 (GOOS만으로 결정 안 함) — case-sensitive APFS variant 생성 시뮬레이션 또는 hdiutil dmg fixture
- T4.4: **NFR-UPC-P1 벤치마크 게이트 (v0.2.1 benchstat 통계 강화, D-02-05)** (`internal/template/deployer_bench_test.go`):
  - `Benchmark_UpdateCleanup_Baseline` — atomic write 미적용 (현재 v2.14.0 동작 시뮬레이션)
  - `Benchmark_UpdateCleanup_AtomicWrite` — atomic write 적용
  - 합성 프로젝트 fixture: `setupSyntheticProject(t, 100, 5, 0)` — 100 files, 5 deprecated paths, 0 symlinks (재현성 우선)
  - **CI 통계 게이트 절차**:
    1. `go install golang.org/x/perf/cmd/benchstat@latest` (CI step)
    2. `go test -bench=Benchmark_UpdateCleanup -benchmem -count=10 ./internal/template/... -run=^$ | tee bench-result.txt` (10 independent runs)
    3. baseline + atomic 결과를 별도 파일로 split (sed/awk)
    4. `benchstat -alpha 0.05 baseline.txt atomic.txt | tee bench-report.txt`
    5. `scripts/benchmark-overhead-check.sh bench-report.txt` — delta_pct ≤ 5% AND p_value < 0.05 검증, 실패 시 exit 1
  - **공식 게이트 OS**: `ubuntu-latest` (재현성 우선). macOS/Windows 결과는 PR 코멘트에만 첨부, 게이트에 영향 없음
  - **결과 PR 자동 코멘트**: `bench-report.txt`를 PR body에 attach (별도 GitHub Action 워크플로우)
- T4.5: `make test` 통합 후 CI 그린 확인 (3 플랫폼)

**Files Modified:**

- `/Users/goos/MoAI/moai-adk-go/internal/cli/update_e2e_test.go` (신규, ~165 LOC, 6 시나리오 + 3 case probe edge tests)
- `/Users/goos/MoAI/moai-adk-go/internal/template/deployer_bench_test.go` (신규, ~30 LOC, 2 benchmarks)
- `/Users/goos/MoAI/moai-adk-go/scripts/benchmark-overhead-check.sh` (신규, ~25 LOC, benchstat parsing + p-value 검증)
- `/Users/goos/MoAI/moai-adk-go/.github/workflows/test.yml` (~10 LOC, benchstat install + benchmark step 추가)

**LOC Estimate:** ~165 LOC tests + ~30 LOC benchmark + ~30 LOC CI = ~225 LOC (단순 합산); 헬퍼 재사용 dedup 후 실효 ~165 LOC

**Dependencies:** M1 + M2 + M3 (모든 기능 구현 완료 후 통합 검증)

**Definition of Done:**
- [ ] 6개 시나리오(A~F) 모두 통과
- [ ] case-insensitive FS test가 macOS/Linux 양쪽에서 의도대로 동작
- [ ] case probe 3개 엣지 케이스(probe failure / frequency / APFS variant) 검증 (v0.2.1)
- [ ] NFR-UPC-P1 benchstat 게이트가 CI에 통합되어 (delta > 5% AND p < 0.05) 시 실패 (v0.2.1)
- [ ] benchstat 결과가 PR body에 자동 코멘트로 attach
- [ ] CI ubuntu-latest, macos-latest, windows-latest에서 그린 (단, NFR 게이트는 ubuntu-latest만 공식)
- [ ] 기존 deployer_test.go 및 update_test.go 회귀 없음

---

## 3. Sequencing Decision

**시퀀셜 진행 (M1 → M2 → M3 → M4)** 채택. 근거:

- M2의 backup 흐름이 M1의 atomic write 안전성에 의존 (백업 도중 중단 시 .moai-tmp 정리 보장)
- M3의 manifest provenance check / telemetry / symlink가 M2의 cleanup 분기에 통합되므로 M2 우선 완료 필요
- M4 E2E 테스트 + NFR 벤치마크는 M1+M2+M3 모두 작동하는 상태에서만 의미 있음

병렬화 가능 영역: M1의 unit tests(T1.4)와 M2의 dirs.go 정의(T2.1)는 동일 시점 작업 가능하나 효과 미미하므로 시퀀셜 유지.

---

## 4. Risk Register

| ID | 리스크 | 영향도 | 완화 |
|----|--------|--------|------|
| R1 | macOS APFS의 `os.Rename` 동작이 cross-volume에서 비원자적일 수 있음 | High | `.moai-tmp`를 destination과 동일 디렉터리에 생성 (NFR-UPC-S3); 추가로 `runtime.GOOS == "darwin"` 시 fsync 호출 검토 |
| R2 | Lock file이 비정상 종료(SIGKILL)로 정리 안 될 경우 영구 차단 | Medium | Lock file에 PID 기록 + 시작 시 `kill -0 <PID>`로 stale 감지 후 자동 정리 |
| R3 | `merge.ConfirmMerge` API 시그니처 — v0.1에서 미확인 | Medium | **v0.2에서 해소**: 시그니처 = `func ConfirmMerge(analysis MergeAnalysis) (bool, error)` (`internal/merge/confirm.go:623`). risk는 `MergeAnalysis.RiskLevel` + `FileAnalysis.RiskLevel` 필드로 전달. wrapper 불필요. |
| R4 | 사용자가 agency 파일을 .gitignore에 의존하여 보존 중일 수 있음 | Low | manifest provenance check + UnverifiedDeprecated 분류로 대응; 백업 보존으로 복구 가능 |
| R5 | Windows에서 path separator 차이로 DeprecatedPaths 매칭 실패 | Medium | `filepath.ToSlash()` 정규화 + Windows CI 테스트 (CLAUDE.local.md §lessons #7) |
| R6 (NEW) | Symbolic link 폐기 경로 처리 — link target이 외부 위치(`/Users/.../external-rules`)일 경우 backup 시 외부 콘텐츠가 같이 백업되거나 link target이 삭제될 수 있음 | High | REQ-UPC-023: `os.Lstat`로 symlink 감지 후 link 자체만 삭제. backup `MANIFEST.json`에 `symlink_target` 기록. **target은 절대 follow 안 함**. |
| R7 (NEW) | macOS APFS는 기본 case-insensitive — `agency.md`와 `Agency.md`가 동일 파일로 인식되어 매칭 누락 또는 중복 처리 가능 | Medium | REQ-UPC-026: cleanup phase 시작 시 case-insensitive FS probe (test file 생성 + 대소문자 변형 stat) → case-insensitive FS이면 deprecated 매칭도 case-insensitive로 전환 |
| R8 (NEW) | `.moai/backup/` 디렉터리 자체가 다음 `moai update` 시 스캔 대상에 포함되어 무한 누적 또는 자기 백업 시도 가능 | Medium | REQ-UPC-024: deprecated 스캔 함수에서 `.moai/backup/` 접두사 경로는 무조건 skip. 단위 테스트 (`TestCleanup_BackupSelfReferenceSkipped`)로 회귀 방지. |

---

## 5. Definition of Done (전체 SPEC)

- [ ] M1~M4 모든 마일스톤의 DoD 충족
- [ ] `go test ./...` 전체 통과 (race detection 포함)
- [ ] `golangci-lint run` zero warnings
- [ ] `make build` 성공 (embedded templates 재생성)
- [ ] CI 3 플랫폼(ubuntu/macos/windows) 그린
- [ ] NFR-UPC-P1 벤치마크 게이트 통과 (≤ 5% overhead)
- [ ] PR 본문에 6개 시나리오 A~F 검증 결과 첨부
- [ ] CHANGELOG.md에 한글+영문 엔트리 추가
- [ ] 기존 사용자 프로젝트 1개에서 수동 dry-run 검증 (백업 디렉터리 생성 + agency/ 정리 + telemetry jsonl 확인)
- [ ] [HARD] 출시 후 텔레메트리 jsonl 로그 수집 계획 합의 (Bug 1 완화 검증 — GitHub Issue 템플릿에 첨부 안내 추가)
