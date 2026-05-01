---
id: SPEC-V3R3-UPDATE-CLEANUP-001
version: "0.2.1"
status: draft
created_at: 2026-05-01
updated_at: 2026-05-01
author: manager-spec
priority: High
labels: [cli, update, deployment, cleanup, agency, idempotency]
issue_number: null
related_specs: [SPEC-AGENCY-ABSORB-001]
---

# SPEC-V3R3-UPDATE-CLEANUP-001: `moai update` — 멱등 배포 + 폐기 경로 정리

## HISTORY

- 2026-05-01 v0.2.1: audit v2 remediation — D-02-01 critical (REQ-UPC-018을 `.moai-skip-cleanup` 사용자 opt-out 메커니즘으로 복원) + D-02-02~D-02-08 major (`.moai/logs/` self-reference 보호, telemetry permission 처리, LOC 공식 명시화, benchstat 기반 NFR-P1 통계 검증, OQ3 해결, case probe 엣지 케이스, symlink 엣지 케이스). REQ 26→27, AC 27→28 + AC-UPC-022b 추가, OQ 2→1.
- 2026-05-01 v0.2.0: 독립 감사 후속 개정 — Critical 결함 D-01~D-04 (LOC 일관성, Bug 1 검증 갭, OQ1/OQ2 코드 인스펙션, dead-path 사용자 처리) + Major 결함 D-08/D-10/D-11 (R6/R7/R8 리스크, 삭제 파일 처리, 벤치마크 검증) 해소. REQ-UPC-022~026 5종 추가 (텔레메트리, symlink, backup self-reference, 삭제 파일, 대소문자 무관 FS). UnverifiedDeprecated 분류 신설 (Pristine / UserModified / UnverifiedDeprecated 삼분). LOC ~95에서 ~380으로 reconcile (atomic write + lock + cleanup + backup + confirmation + provenance + 5 신규 REQ + 테스트 모두 포함). OQ1/OQ2 해결. R6/R7/R8 신규 추가.
- 2026-05-01 v0.1.0: 최초 작성. 사용자 프로젝트에서 발견된 두 결함 번들링 — (1) `moai update` 후 ` 2` 접미어 중복 파일 발생, (2) SPEC-AGENCY-ABSORB-001(`3e8b61e80`, 2026-04-23)에서 템플릿 제거된 agency 디렉터리가 사용자 프로젝트에 잔존. expert-debug 1차 조사 결과를 SPEC 요구사항으로 정형화.

---

## 1. Background & Problem Statement (배경 및 문제 정의)

### 1.1 결함 1 — `moai update` 실행 후 ` 2` 접미어 중복 파일 발생

여러 사용자 프로젝트에서 `moai update` 실행 후 `.claude/agents/manager-spec 2.md`, `.moai/config/sections/quality 2.yaml` 등 ` 2` 접미어가 붙은 중복 파일이 산발적으로 발견되었다. expert-debug 조사 결과 Go의 `os.WriteFile` (`internal/template/deployer.go:168`)은 원자적 덮어쓰기 연산이며 ` 2` 접미어를 생성하지 않는다. 즉, **코드 자체의 직접적 결함이 아니라 환경 요인**(macOS iCloud Drive 동기화 충돌, APFS 확장 속성, Finder 외부 복사 개입, 또는 동시 `moai update` 실행)으로 추정된다.

[HARD] **Bug 1 완화의 한계 인정 (D-02 후속)**: 본 SPEC의 atomic write 패턴(write-to-tmp + rename)은 환경 레이스 조건에 대한 *완화책*이며 *검증된 근본 원인 수정*이 아니다. expert-debug는 macOS iCloud Drive / APFS / 동시 실행을 가장 유력한 가설로 결론지었으나 결정적 증거(deterministic reproduction)를 확보하지 못했다. 따라서 Bug 1은 예상치 못한 환경 컨텍스트(예: 새로운 클라우드 동기화 도구, 외부 백업 솔루션 개입)에서 재발할 가능성을 배제할 수 없다.

이 한계를 보완하기 위해 **REQ-UPC-022 (텔레메트리)**를 도입하여 `moai update` cleanup phase 종료 시점에 (a) atomic write 사용 플래그, (b) cleanup 진입 시점에 탐지된 ` 2` 접미어 파일 목록과 경로, (c) 백업 결과를 구조화 로그(JSON line)로 `.moai/logs/update-cleanup-{timestamp}.jsonl`에 기록한다. 외부 네트워크 호출 없음. 출시 후 사용자 프로젝트의 jsonl 로그 수집을 통해 완화 효과를 사후 검증할 수 있다.

또한 현재 `deployer_test.go`에는 멱등성 테스트(연속 두 번 Deploy 호출)가 부재하여 회귀 방지 장치가 없다.

### 1.2 결함 2 — 폐기된 agency 폴더가 사용자 프로젝트에서 정리되지 않음

SPEC-AGENCY-ABSORB-001 완료 시점(2026-04-23 commit `3e8b61e80`)에 템플릿에서 `.claude/commands/agency/`(8 파일)와 `.claude/rules/agency/constitution.md`가 제거되었다. 그러나 `internal/cli/update.go:1411-1441`의 `cleanMoaiManagedPaths` 함수는 `commands/moai`만 정리 대상에 포함하고 agency 경로는 미포함이다. 또한 `isMoaiManaged()` (`internal/cli/update.go:1064-1068`)는 `moai*` 접두사만 매칭하므로 agency 파일들은 `UserCreated`로 분류되어 영구 보존된다. 결과적으로 사용자는 Claude Code 슬래시 커맨드 팔레트에서 폐기된 `/agency`, `/brief`, `/build` 등을 계속 보게 되며, 호출 시 깨진 동작 또는 혼동을 경험한다.

### 1.3 두 결함을 하나의 SPEC으로 묶는 근거

두 결함 모두 `internal/cli/update.go` 및 deployer 체인을 건드린다. 동일한 회귀 테스트 인프라(임시 사용자 프로젝트 시뮬레이션)를 공유하므로 단일 SPEC에서 해결하는 것이 통합 테스트 일관성 측면에서 유리하다. **LOC 추정은 ~380 LOC** — atomic write 패턴 + transactional mode + lock file + deprecated cleanup + backup with manifest + user confirmation + provenance check + telemetry + symlink/case-insensitive/backup-self-reference 처리 + 회귀 테스트 5 시나리오를 모두 포함한다. (참고: v0.1.0에서 명시한 "~95 LOC"는 deprecated path 추가만 고려한 초기 협소 추정치였으며, atomic write와 provenance 작업을 포함하지 않은 부분 견적이었다. v0.2.0에서 plan.md §2 마일스톤 합산값(M1=95 + M2=75 + M3=90 + M4=120 = 380 LOC)으로 정정한다.)

---

## 2. Goals (목표)

- G1: `moai update`를 멱등하게 만든다 — 동일 입력으로 N회 실행해도 ` 2` 접미어 등 부산물이 생성되지 않는다.
- G2: 템플릿에서 제거된 폐기 경로(agency 디렉터리)를 `moai update` 실행 시 안전하게 정리한다.
- G3: 사용자 커스터마이징을 보호한다 — 폐기 파일이 사용자에 의해 수정되었거나 provenance를 알 수 없는 경우 자동 삭제 대신 백업 + 경고를 제공한다.
- G4: 모든 삭제 작업은 `.moai/backup/agency-{timestamp}/`에 백업본을 보존한다.
- G5: `deployer_test.go`와 `update_test.go`에 회귀 방지 테스트를 추가한다.
- G6: Bug 1 완화 효과를 사후 검증할 수 있는 구조화 텔레메트리 로그를 출시한다 (D-02 보강).

---

## 3. Non-Goals (비범위)

- N1: 결함 1의 환경 요인 근본 해결(예: macOS iCloud Drive 동기화 차단). 본 SPEC은 atomic write 패턴으로 환경 레이스 조건에 강건한 코드 측 완화 + 텔레메트리만 다룬다.
- N2: `.agency/` 프로젝트 루트 디렉터리 마이그레이션. 별도 `moai migrate agency` 커맨드(SPEC-AGENCY-ABSORB-001 산출물)가 이 영역을 담당한다.
- N3: 폐기 경로 정리 정책의 일반화 프레임워크 도입. 본 SPEC은 agency 케이스만 하드코딩 처리하며, 향후 폐기 정책은 별도 SPEC에서 다룬다.
- N4: macOS 외 플랫폼(Linux, Windows)에서의 ` 2` 접미어 재현 시도. atomic write 패턴은 모든 플랫폼에 동일 적용되므로 효과는 자동 확장된다.
- N5: 텔레메트리 데이터 외부 전송. JSON line 로그는 사용자 로컬 디스크에만 보존되며, 사용자가 수동으로 GitHub Issue에 첨부하는 것이 출시 후 검증의 유일한 채널이다.

---

## 4. Functional Requirements (EARS Format)

### 4.1 멱등 배포 (Idempotent Deployment)

- **REQ-UPC-001 (Event-Driven):** WHEN the deployer writes a template file to a destination, THE deployer SHALL use an atomic temporary-file pattern: write to `<destPath>.moai-tmp` first, then rename to `<destPath>` via `os.Rename`.
- **REQ-UPC-002 (Event-Driven):** WHEN `moai update` is invoked twice consecutively without intervening modifications, THE deployer SHALL produce byte-identical destination files with no auxiliary files (no ` 2` suffix, no `.moai-tmp` residue).
- **REQ-UPC-003 (Unwanted):** THE deployer SHALL NOT leave `.moai-tmp` files in the destination tree after successful Deploy completion.
- **REQ-UPC-004 (If-Then):** IF a `.moai-tmp` file rename fails, THEN THE deployer SHALL remove the temporary file and return an error wrapping the underlying cause.
- **REQ-UPC-005 (State-Driven):** WHILE a previous `moai update` is in progress (detectable via lock file `.moai/.update.lock`), THE deployer SHALL refuse to start a second concurrent execution and return a clear error message.

### 4.2 폐기 경로 탐지 (Deprecated Path Detection)

- **REQ-UPC-006 (Ubiquitous):** THE update command SHALL maintain a deprecated paths registry in `internal/defs/dirs.go` listing canonical paths removed from templates by previous SPECs.
- **REQ-UPC-007 (Event-Driven):** WHEN `moai update` runs in an existing user project, THE update command SHALL scan for the presence of every path in the deprecated paths registry.
- **REQ-UPC-008 (State-Driven):** WHILE no deprecated paths are present in the user project, THE update command SHALL skip the cleanup phase silently with no backup directory creation.

### 4.3 백업 (Backup Before Deletion)

- **REQ-UPC-009 (Event-Driven):** WHEN deprecated paths are detected, THE update command SHALL create a backup at `.moai/backup/agency-{ISO8601-timestamp}/` preserving the original directory structure before any deletion occurs.
- **REQ-UPC-010 (If-Then):** IF backup creation fails for any reason (disk full, permission denied), THEN THE update command SHALL abort the cleanup phase, leave deprecated paths intact, and report the backup failure to the user.
- **REQ-UPC-011 (Ubiquitous):** THE backup directory SHALL include a `MANIFEST.json` file recording original paths, file hashes, deletion timestamp, the SPEC ID that authorized the cleanup, and (when applicable) `symlink_target` for symbolic-link entries.

### 4.4 사용자 확인 (User Confirmation)

- **REQ-UPC-012 (Event-Driven):** WHEN deprecated paths are detected, THE update command SHALL invoke `merge.ConfirmMerge(analysis MergeAnalysis) (bool, error)` (defined at `internal/merge/confirm.go:623`) with a `MergeAnalysis` whose top-level `RiskLevel` is `"high"` and whose every `FileAnalysis.RiskLevel` is `"high"`, presenting the user with the list of paths to be backed up and removed.
- **REQ-UPC-013 (If-Then):** IF `ConfirmMerge` returns `(false, nil)` (user declined), THEN THE update command SHALL skip cleanup, leave the deprecated paths intact, and emit an informational message indicating cleanup was deferred.
- **REQ-UPC-014 (Optional, OQ3 resolved):** WHERE the `--auto-confirm-cleanup` flag is provided (default `false`, requires explicit user opt-in), THE update command SHALL bypass the `ConfirmMerge` call and proceed directly with backup + deletion **only for files classified as `PristineDeprecated`**. `UserModifiedDeprecated` and `UnverifiedDeprecated` files SHALL ALWAYS require explicit confirmation regardless of this flag (safety-by-design — cannot be bypassed). The flag name follows existing `moai update` conventions (`--force`, `--dry-run`).

### 4.5 커스터마이징 안전성 (Customization Safety) — 삼분 분류 (D-04 보강)

본 SPEC은 폐기 파일을 다음 세 가지 분류로 구분한다 (Pristine / UserModified / UnverifiedDeprecated). 각 분류의 정의와 처리 정책:

| 분류 | 정의 | 처리 |
|---|---|---|
| `PristineDeprecated` | manifest에 `deployed_hash` 기록 존재 + 현재 파일 hash가 `deployed_hash`와 일치 | 백업 후 자동 삭제 가능 (auto-confirm 가능) |
| `UserModifiedDeprecated` | manifest에 `deployed_hash` 기록 존재 + 현재 파일 hash가 `deployed_hash`와 불일치 | 백업 + 사용자 확인 필수, stderr 경고 출력 |
| `UnverifiedDeprecated` | manifest에 해당 파일 entry 부재 (provenance 미기록 — 기존 사용자의 일반적 케이스) | 백업 필수, `--auto-confirm-cleanup`로도 자동 삭제 불가, 사용자 확인 필수, stderr 경고 출력 |

- **REQ-UPC-015 (Event-Driven):** WHEN scanning deprecated paths, THE update command SHALL look up each file in `.moai/manifest.json` via `manifest.Manager.GetEntry(path)` and compute the SHA-256 hash of the current file via `manifest.HashFile(absPath)`.
- **REQ-UPC-015a (If-Then):** IF a manifest entry exists for the deprecated file AND the current hash equals the recorded `deployed_hash`, THEN THE update command SHALL classify the file as `PristineDeprecated`.
- **REQ-UPC-015b (If-Then):** IF a manifest entry exists for the deprecated file AND the current hash does NOT equal the recorded `deployed_hash`, THEN THE update command SHALL classify the file as `UserModifiedDeprecated`, back it up, and emit an stderr warning naming the file before deletion.
- **REQ-UPC-015c (If-Then):** IF NO manifest entry exists for the deprecated file (provenance unknown — typical for users who installed before manifest tracking was complete), THEN THE update command SHALL classify the file as `UnverifiedDeprecated`. Backup is mandatory. The file SHALL NOT be removed even when `--auto-confirm-cleanup` is set; explicit user confirmation via `ConfirmMerge` is required. An stderr warning naming the file SHALL be emitted before any removal attempt.
- **REQ-UPC-016 (Unwanted):** THE update command SHALL NOT silently delete a file classified as `UserModifiedDeprecated` or `UnverifiedDeprecated` without the corresponding stderr warning being emitted.
- **REQ-UPC-017 (Ubiquitous):** THE backup `MANIFEST.json` SHALL record each entry's classification (`PristineDeprecated` / `UserModifiedDeprecated` / `UnverifiedDeprecated`) so post-hoc analysis can distinguish the three cases.
- **REQ-UPC-018 (Optional, restored v0.2.1, addresses D-02-01):** WHERE the user has placed a `.moai-skip-cleanup` marker file inside a deprecated path directory (e.g., `.claude/commands/agency/.moai-skip-cleanup`), THE update command SHALL skip cleanup of that directory regardless of file classification, emit a single `[INFO] cleanup skipped due to .moai-skip-cleanup marker: <path>` line to stderr, and record the skipped paths in the telemetry log entry (REQ-UPC-022) under a new field `user_opt_out_paths: [string]`. This provides users an explicit manual override mechanism complementing the automatic three-way classification.

### 4.6 회귀 방지 (Regression Prevention)

- **REQ-UPC-019 (Ubiquitous):** THE test suite SHALL include an idempotency test in `internal/template/deployer_test.go` verifying that two consecutive `Deploy` calls on the same destination produce byte-identical results with no auxiliary files.
- **REQ-UPC-020 (Ubiquitous):** THE test suite SHALL include a deprecated-path-cleanup regression test in `internal/cli/update_test.go` covering the five scenarios A–E enumerated in `acceptance.md`.
- **REQ-UPC-021 (Ubiquitous):** THE test suite SHALL run with `t.TempDir()` isolation per CLAUDE.local.md §6 and MUST NOT modify files outside the temporary directory.

### 4.7 관측성 및 환경 안전 (Observability and Environment Safety) — v0.2.0 신규

- **REQ-UPC-022 (Event-Driven, addresses D-02; extended v0.2.1 D-02-03):** WHEN the `moai update` cleanup phase finishes (success or failure), THE update command SHALL emit a structured JSON-line event to stderr AND persist the same line to `.moai/logs/update-cleanup-{ISO8601-timestamp}.jsonl`. The event SHALL contain: (a) `atomic_write_used: bool`, (b) `pre_update_suffix2_files: [string]` — list of paths matching the ` 2` suffix detected at cleanup-phase start, (c) `backup_outcome: "success" | "skipped" | "failed"`, (d) `backup_path: string` (when applicable), (e) `cleanup_outcome: "completed" | "deferred" | "aborted"`, (f) `user_opt_out_paths: [string]` — directories skipped due to `.moai-skip-cleanup` marker (REQ-UPC-018). NO external network call SHALL be made.
  - **Permission denied handling (v0.2.1):** IF the telemetry log directory `.moai/logs/` cannot be created or written due to filesystem permissions (e.g., chmod 0500 on parent), THEN THE update command SHALL emit a single `[WARN] telemetry persistence skipped: <reason>` line to stderr, skip the file persistence step, and continue the update operation without failure. The stderr mirror of the JSON line SHALL still be emitted for observability.
- **REQ-UPC-023 (If-Then, addresses D-08 R6; extended v0.2.1 D-02-08):** IF a deprecated path entry resolves to a symbolic link, THEN THE update command SHALL NOT follow the link for backup or deletion purposes. THE update command SHALL remove only the symbolic link itself, and the backup `MANIFEST.json` entry SHALL include a `symlink_target: string` field recording the link's original target path.
  - **Broken symlink (v0.2.1):** IF a deprecated path is a symlink with a non-existent target (broken symlink, `os.Stat` on target returns ENOENT), THEN THE update command SHALL remove only the link itself (no backup attempt of the missing target file content) and record `symlink_target_status: "broken"` in the corresponding MANIFEST.json entry. The `symlink_target` field SHALL still record the dangling target path string.
  - **Symlink target collision (v0.2.1):** IF a symlink target itself matches another deprecated path entry in the registry, THEN THE update command SHALL classify and process the link and the target independently; the target file SHALL be backed up exactly once (under its own deprecated entry), not double-counted via the symlink entry.
  - **Windows symlinks (v0.2.1):** ON Windows platforms, junction points and reparse points SHALL be treated identically to POSIX symlinks for cleanup purposes. IF the platform does not support symlink detection (e.g., older Windows without administrator privileges, `os.Lstat` returns equivalent regular-file info), THEN THE update command SHALL log `[INFO] symlink detection unsupported on this platform` to stderr and treat all paths as regular files for that invocation.
- **REQ-UPC-024 (Ubiquitous, addresses D-08 R8; extended v0.2.1 D-02-02):** THE update command SHALL exclude `.moai/backup/` AND `.moai/logs/` (and all subdirectories thereof) from any deprecated-path scan, classification, or cleanup operation. Subsequent `moai update` invocations SHALL NOT attempt to clean up their own previous backup output, AND SHALL NOT clean up their own previously emitted telemetry log files (REQ-UPC-022 self-reference protection).
- **REQ-UPC-025 (If-Then, addresses D-10):** IF a deprecated path entry exists in the registry (`internal/defs/dirs.go::DeprecatedPaths`) but the corresponding file is absent from the user project, THEN THE update command SHALL skip the entry silently — no classification, no backup, no warning, no log entry.
- **REQ-UPC-026 (Event-Driven, addresses D-08 R7; extended v0.2.1 D-02-07):** WHEN running on a case-insensitive filesystem (detected at start of cleanup phase via a test-file creation probe — create `.moai-fscase-probe`, attempt to stat with the uppercase variant, then remove), THE update command SHALL match deprecated paths case-insensitively (e.g., `agency.md` and `Agency.md` resolve to the same registry entry). On case-sensitive filesystems, matching SHALL remain exact-case.
  - **Probe failure fallback (v0.2.1):** IF the case-sensitivity probe fails (read-only filesystem, write permission denied, parent directory missing, etc.), THEN THE update command SHALL default to case-sensitive matching and emit a single `[INFO] case-sensitivity probe failed; defaulting to case-sensitive matching: <reason>` line to stderr.
  - **Probe frequency (v0.2.1):** THE probe SHALL execute exactly once per `moai update` invocation, scoped to the project root filesystem. Probe results SHALL NOT be cached across invocations (avoids stale state when filesystem is remounted with different case settings between runs).
  - **APFS variant volumes (v0.2.1):** WHERE the project root resides on an APFS case-sensitive variant volume (probe-detected), THE update command SHALL use case-sensitive matching even on macOS. The probe is the authoritative signal — platform identity (GOOS) alone SHALL NOT determine matching mode.

---

## 5. Non-Functional Requirements

### 5.1 Performance (성능)

- NFR-UPC-P1: Atomic write 패턴(write-then-rename)은 직접 write 대비 ≤ 5% latency overhead 이내여야 한다 (statistically significant, p-value < 0.05). **검증 방법 (v0.2.1 통계 강화, D-02-05)**: `internal/template/deployer_bench_test.go`에 Go 벤치마크 (`Benchmark_UpdateCleanup_Baseline` vs `Benchmark_UpdateCleanup_AtomicWrite`)를 추가한다. Baseline은 v2.14.0 update duration on synthetic project (100 files, 5 deprecated paths, no symlinks)로 고정. **CI 게이트 통계 방법론**:
  - 실행: `go test -bench=Benchmark_UpdateCleanup -benchmem -count=10 ./internal/template/...` (10 independent runs per benchmark)
  - 분석: `benchstat baseline.txt atomic.txt` (mean ± stddev + Welch's t-test 또는 Mann-Whitney U test, p-value 자동 계산)
  - 게이트 통과 조건: **`delta_pct ≤ 5%` AND `p_value < 0.05`** 두 조건 모두 만족 시 통과. p-value ≥ 0.05이면 통계적으로 유의하지 않음 → 통과 판정 (noise 범위로 간주). delta > 5% AND p_value < 0.05이면 실패 (CI exit 1).
  - **공식 게이트 OS**: `ubuntu-latest` (가장 재현성 높음). macOS / Windows 벤치마크는 정보 제공용으로만 수집되며 게이트에 영향 없음.
  - **benchstat 호출 예시**: `go install golang.org/x/perf/cmd/benchstat@latest && benchstat -alpha 0.05 baseline.txt atomic.txt | tee bench-report.txt`
- NFR-UPC-P2: 폐기 경로 스캔은 사용자 프로젝트 1000 파일 기준 200ms 이내에 완료되어야 한다.
- NFR-UPC-P3: 백업 디렉터리 생성은 `cp -R` 또는 Go의 `io/fs.CopyFS` 등 최적화된 경로를 사용한다 (개별 파일 read/write 루프 금지).

### 5.2 Safety (안전성)

- NFR-UPC-S1: 백업이 성공적으로 완료되기 전에는 어떤 폐기 파일도 삭제되지 않는다 (원자적 백업-삭제 시퀀스).
- NFR-UPC-S2: 동시 실행 방지를 위한 `.moai/.update.lock` 파일은 프로세스 종료 시(정상/비정상 모두) 자동 정리된다 (defer + signal handler).
- NFR-UPC-S3: `.moai-tmp` 임시 파일은 같은 디렉터리에 생성되어 cross-device rename 실패를 방지한다.
- NFR-UPC-S4: Symbolic link는 절대 follow하지 않는다 (REQ-UPC-023). 외부 위치로의 backup leak 방지.
- NFR-UPC-S5: `.moai/backup/` self-reference 차단 (REQ-UPC-024). 무한 루프 방지.

### 5.3 Observability (관측성)

- NFR-UPC-O1: 백업 경로, 삭제된 파일 수, 사용자 수정 감지 건수, UnverifiedDeprecated 건수를 `moai update` stdout에 명시적으로 출력한다.
- NFR-UPC-O2: `MANIFEST.json`은 향후 `moai restore` 커맨드(out of scope)로 복원 가능한 형태로 작성한다.
- NFR-UPC-O3: 모든 삭제 작업은 `.moai/logs/update-{timestamp}.log`에 구조화 로그로 기록한다.
- NFR-UPC-O4: REQ-UPC-022의 cleanup phase 종료 이벤트는 별도 `.moai/logs/update-cleanup-{timestamp}.jsonl`에 기록되어 사후 분석을 단순화한다 (Bug 1 완화 효과 검증용).

---

## 6. Open Questions (Plan/Run으로 이연)

v0.2.0에서 OQ1과 OQ2는 코드 인스펙션을 통해 해소되었다 (`/Users/goos/MoAI/moai-adk-go/internal/merge/confirm.go:623`, `/Users/goos/MoAI/moai-adk-go/internal/manifest/types.go:11-39`).

**OQ1 (resolved 2026-05-01):** `merge.ConfirmMerge` 시그니처는 `func ConfirmMerge(analysis MergeAnalysis) (bool, error)` (`internal/merge/confirm.go:623`). risk classification은 추가 인자가 아니라 `MergeAnalysis` 구조체의 두 필드를 통해 전달된다 — top-level `RiskLevel` (`"low"`/`"medium"`/`"high"`)는 전체 머지의 위험도를, 개별 `FileAnalysis.RiskLevel`은 파일별 위험도를 나타낸다 (`confirm.go:24, 32`). Validation은 `validateAnalysis` (`confirm.go:558-588`)에서 `low`/`medium`/`high` 세 값만 허용한다. 따라서 deprecated cleanup은 `MergeAnalysis{RiskLevel: "high", HasConflicts: true, Files: [...각 파일 RiskLevel "high"...]}`로 구성하여 호출한다. 신규 wrapper 함수는 불필요하다. → REQ-UPC-012/013/014 갱신 완료.

**OQ2 (resolved 2026-05-01):** 매니페스트 provenance hash는 **이미 기존 `manifest.json` 스키마에 존재**한다. `internal/manifest/types.go:42-54`의 `FileEntry` 구조체에 `template_hash` / `deployed_hash` / `current_hash` 3-hash 필드가 정의되어 있으며, `manifest.go:130-149`의 `Track` 메서드와 `manifest.go:165-201`의 `DetectChanges` 메서드가 이를 활용한다. 또한 `Provenance` enum에 `TemplateManaged` / `UserModified` / `UserCreated` / `Deprecated` 4가지 분류가 이미 존재한다 (`types.go:14-30`). **결정**: 별도 `provenance.json` 사이드카를 두지 않고, 기존 manifest에 이미 존재하는 `deployed_hash`를 provenance hash로 활용한다. 본 SPEC의 deprecated cleanup은 `manifest.Manager.GetEntry(path)`로 entry 조회 후 `entry.DeployedHash` vs `manifest.HashFile(currentPath)` 비교로 분류를 결정한다. 추가 schema 마이그레이션 불필요. → REQ-UPC-015~018 갱신 완료.

**OQ3 (resolved 2026-05-01, v0.2.1):** `--auto-confirm-cleanup` 플래그 명명 및 동작 결정 — 플래그 명: `--auto-confirm-cleanup` (기존 `moai update` 플래그 관례 `--force`/`--dry-run`과 일관). 적용 범위: **`PristineDeprecated`만 prompt bypass**. `UserModifiedDeprecated`와 `UnverifiedDeprecated`는 본 플래그가 set이어도 ALWAYS confirmation prompt 호출 (safety-by-design — 안전성 우선). 기본값: `false` (사용자 명시적 opt-in 필요). → REQ-UPC-014에 인라인 반영, 의존성 태그 제거 완료.

남아 있는 Open Questions:

- OQ4: Lock file (`.moai/.update.lock`) PID 기록 형식 — 단순 PID인지, JSON(pid + timestamp + hostname)인지 결정. (구현 상세이므로 Plan/Run 단계로 이연.)

---

## 7. Constraints (제약)

- C1: **16-language neutrality** — `internal/template/templates/` 하위는 16개 언어 동등 취급 (CLAUDE.local.md §15). 본 SPEC의 deprecated paths 정의는 모든 언어 프로젝트에 동일 적용된다.
- C2: **Template-First** — 새로운 deprecated 경로 추가 시, 반드시 `internal/defs/dirs.go`(코드)와 SPEC `acceptance.md`(문서) 양쪽에 등록한다 (CLAUDE.local.md §2). [HARD] **빌드 타임 enforcement는 본 SPEC 범위 밖**이며 (D-09 후속), 현재는 **코드 리뷰 게이트**로 보장한다. 향후 별도 SPEC에서 `make verify-deprecated-registry` lint를 자동화한다.
- C3: **No hardcoding** — Backup root path, lock file path, manifest filename은 `internal/config/defaults.go`에 단일 원천 정의하고 코드 전반에서 참조 (CLAUDE.local.md §14).
- C4: **t.TempDir() isolation** — 모든 테스트는 `t.TempDir()` 사용, 프로젝트 루트 또는 사용자 홈 디렉터리에 파일 생성 금지 (CLAUDE.local.md §6).
- C5: **No breaking API changes** — `moai update`의 기존 CLI 시그니처는 유지. 새 동작은 기본 활성(safe defaults: prompt + backup)이며 opt-out 플래그로만 제어.
- C6: **Backward compatibility** — agency 파일이 사용자 커스터마이징을 포함할 수 있으므로 매니페스트 provenance check + 백업이 의무. 더 나아가 manifest entry 부재 케이스(기존 사용자의 일반적 상황)는 `UnverifiedDeprecated`로 별도 분류하여 안전성을 우선시한다 (REQ-UPC-015c).
- C7: **No telemetry exfiltration** — REQ-UPC-022의 텔레메트리 로그는 사용자 로컬 디스크에만 보존된다. 외부 네트워크 호출 금지. CLAUDE.md §10 (Web Search Protocol)와 SPEC-AGENCY-ABSORB-001의 안전성 원칙을 따른다.

---

## 8. Acceptance Cross-Reference

각 REQ-UPC-NNN에 대한 Given/When/Then 시나리오 및 시나리오 A–E (regression scenarios)는 `acceptance.md` 참조.

- REQ-UPC-001~005 → acceptance.md §1 Idempotent Deployment scenarios
- REQ-UPC-006~008 → acceptance.md §2 Deprecated Path Detection scenarios
- REQ-UPC-009~011 → acceptance.md §3 Backup scenarios
- REQ-UPC-012~014 → acceptance.md §4 User Confirmation scenarios
- REQ-UPC-015 / 015a / 015b / 015c / 016 / 017 / 018 → acceptance.md §5 Customization Safety scenarios (삼분 분류 + opt-out marker)
- REQ-UPC-019~021 → acceptance.md §6 Regression Test scenarios
- REQ-UPC-022~026 → acceptance.md §7 Observability and Environment Safety scenarios (AC-UPC-022b telemetry permission 포함)
- 통합 시나리오 A–F → acceptance.md §8 End-to-End Regression Scenarios
- NFR-UPC-P1 벤치마크 → acceptance.md §9 NFR Benchmark Verification (benchstat 통계 검증)

## 9. Summary

- 총 REQ 개수: **27** (UPC-001~005, 006~008, 009~011, 012~014, 015 / 015a / 015b / 015c / 016 / 017 / **018**, 019~021, 022~026) — v0.2.1에서 REQ-UPC-018 (`.moai-skip-cleanup` opt-out) 복원
- 신규 분류 (v0.2.0): `PristineDeprecated`, `UserModifiedDeprecated`, `UnverifiedDeprecated`
- 신규 사용자 메커니즘 (v0.2.1): `.moai-skip-cleanup` marker file (REQ-UPC-018)
- LOC 추정: ~380 (atomic write + lock + cleanup + backup + provenance + telemetry + symlink/case-insensitive/self-reference 처리 + opt-out marker + 회귀 테스트 6 시나리오 + NFR 벤치마크). 상세 산식은 plan.md §1 참조.
