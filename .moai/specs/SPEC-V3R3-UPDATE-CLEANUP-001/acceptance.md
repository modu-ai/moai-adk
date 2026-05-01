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

# SPEC-V3R3-UPDATE-CLEANUP-001 — Acceptance Criteria

## HISTORY

- 2026-05-02 v0.2.2: audit v2 minor patch — D-02-09 (§5 헤더에 /018 추가하여 cross-reference 표기 일관성 확보), D-02-12 (§7.5 Cross-OS Verification Matrix 신설: REQ-UPC-022/023/024/026 OS별 검증 책임 명시), D-02-13 (AC-UPC-022 외부 네트워크 검증 메커니즘을 `httptest.NewServer 401`에서 custom `http.RoundTripper` interception으로 교체하여 임의 host에 대한 dial 차단 검증 정확화). 카운트 변동 없음.
- 2026-05-01 v0.2.1: audit v2 remediation — AC-UPC-018 신규 (`.moai-skip-cleanup` opt-out marker, REQ-UPC-018 복원), AC-UPC-022b 신규 (telemetry permission denied 처리, D-02-03), AC-UPC-024 확장 (`.moai/logs/` self-reference dual-path, D-02-02), AC-UPC-023 sub-cases (broken symlink, target collision, Windows reparse point — D-02-08), AC-UPC-026 sub-cases (probe failure fallback, frequency, APFS variant — D-02-07), AC-NFR-P1 benchstat 통계 게이트 (delta + p-value, D-02-05). AC 총 개수 26 → 28, E2E 시나리오는 6 유지.
- 2026-05-01 v0.2.0: 감사 후속 개정 — REQ-UPC-015~018을 삼분 분류(015 / 015a / 015b / 015c / 016 / 017)로 재구성, REQ-UPC-022~026 5종에 대한 AC 추가 (telemetry / symlink / backup self-reference / deleted file / case-insensitive FS), Scenario E를 deterministic lock acquisition test로 명시 (D-07 후속), Scenario F (UnverifiedDeprecated) 추가, AC-NFR-P1 벤치마크 검증 항목 신설 (D-11 후속), AC 총 개수 21 → 26, E2E 시나리오 5 → 6.
- 2026-05-01 v0.1.0: 최초 작성. 21개 REQ에 대한 Given/When/Then 시나리오 + 5개 E2E 회귀 시나리오 (A~E).

---

## 1. Idempotent Deployment (REQ-UPC-001~005)

### AC-UPC-001 (REQ-UPC-001): Atomic Write Pattern 적용

- **Given:** 빈 destination 디렉터리와 단일 템플릿 파일이 준비된 상태
- **When:** `Deployer.Deploy()`가 호출됨
- **Then:** destination에 `<file>.moai-tmp` 임시 파일이 일시적으로 생성된 후 `<file>`로 rename되며, 최종 상태에 `.moai-tmp`는 부재한다
- **Test type:** unit (`internal/template/deployer_test.go::TestDeployer_AtomicWrite`)
- **Verification note (D-06):** 검증은 "Deploy 성공 후 destination tree에 `.moai-tmp` 부재" + "Deploy 실패 시뮬레이션 후에도 `.moai-tmp` 부재" 두 deterministic assertion으로 한정. "write 도중 tmp 파일 관찰" 같은 race-prone 검증은 사용하지 않는다.

### AC-UPC-002 (REQ-UPC-002): 연속 실행 멱등성

- **Given:** Deploy가 한 번 완료된 destination 디렉터리
- **When:** 동일 입력으로 두 번째 `Deployer.Deploy()`를 호출
- **Then:** 모든 destination 파일이 byte-identical 유지되며, ` 2` 접미어 또는 `.moai-tmp` 부산물이 발생하지 않는다 (디렉터리 entry 개수 변동 없음)
- **Test type:** unit (`internal/template/deployer_test.go::TestDeployer_Idempotent`)

### AC-UPC-003 (REQ-UPC-003): tmp 파일 잔존 금지

- **Given:** Deploy가 정상 완료된 destination 디렉터리
- **When:** 디렉터리 트리를 walk
- **Then:** 어떤 경로에서도 `.moai-tmp` 접미어 파일이 존재하지 않는다
- **Test type:** unit (`internal/template/deployer_test.go::TestDeployer_NoTmpResidue`)

### AC-UPC-004 (REQ-UPC-004): Rename 실패 시 정리

- **Given:** `os.Rename`이 의도적으로 실패하도록 시뮬레이션된 환경 (예: read-only destination)
- **When:** `Deployer.Deploy()` 호출
- **Then:** 함수가 wrapped error를 반환하고, `.moai-tmp` 파일이 정리되어 부재한다
- **Test type:** unit (`internal/template/deployer_test.go::TestDeployer_TmpCleanupOnFailure`)

### AC-UPC-005 (REQ-UPC-005): 동시 실행 차단

- **Given:** 첫 번째 `moai update` 프로세스가 `.moai/.update.lock`을 보유한 상태
- **When:** 두 번째 `moai update` 프로세스 호출
- **Then:** 두 번째 프로세스가 즉시 종료되며, "another moai update is in progress" 에러 메시지를 출력한다 (exit code != 0)
- **Test type:** integration (`internal/cli/update_test.go::TestUpdate_ConcurrentLock`)

---

## 2. Deprecated Path Detection (REQ-UPC-006~008)

### AC-UPC-006 (REQ-UPC-006): DeprecatedPaths 레지스트리

- **Given:** `internal/defs/dirs.go` 컴파일 완료
- **When:** `DeprecatedPaths` 슬라이스를 참조
- **Then:** `.claude/commands/agency/{8 files}`, `.claude/rules/agency/constitution.md` 9개 항목이 정확히 등록되어 있으며, 각 항목에 `DeprecatedSince` (= "2026-04-23") 및 `DeprecatedBy` (= "SPEC-AGENCY-ABSORB-001") 메타데이터가 첨부된다
- **Test type:** unit (`internal/defs/dirs_test.go::TestDeprecatedPaths_Registry`)

### AC-UPC-007 (REQ-UPC-007): 사용자 프로젝트 스캔

- **Given:** 사용자 프로젝트 루트에 `.claude/commands/agency/agency.md` 파일이 존재
- **When:** `moai update` 실행
- **Then:** stdout에 "found N deprecated paths" (N >= 1) 메시지가 출력되며, 탐지된 경로 목록이 사용자에게 표시된다
- **Test type:** integration (`internal/cli/update_test.go::TestCleanup_AgencyPathsDetected`)

### AC-UPC-008 (REQ-UPC-008): 부재 시 no-op

- **Given:** 사용자 프로젝트에 어떤 deprecated path도 부재
- **When:** `moai update` 실행
- **Then:** 백업 디렉터리(`.moai/backup/agency-*`)가 생성되지 않으며, 정리 관련 stdout 출력이 없다
- **Test type:** integration (`internal/cli/update_test.go::TestCleanup_NoOpWhenAbsent`)

---

## 3. Backup Before Deletion (REQ-UPC-009~011)

### AC-UPC-009 (REQ-UPC-009): 백업 디렉터리 구조 보존

- **Given:** 사용자 프로젝트에 `.claude/commands/agency/{2 files}` 존재
- **When:** `moai update` 실행 후 cleanup 동의
- **Then:** `.moai/backup/agency-{ISO8601}/` 디렉터리가 생성되며, 그 하위에 `.claude/commands/agency/{2 files}` 트리가 원본과 byte-identical로 복사된다
- **Test type:** integration (`internal/cli/update_test.go::TestBackup_StructurePreserved`)

### AC-UPC-010 (REQ-UPC-010): 백업 실패 시 abort

- **Given:** `.moai/backup/` 디렉터리에 write permission이 없는 환경
- **When:** `moai update` 실행 (deprecated paths 존재)
- **Then:** cleanup phase가 abort되며, deprecated 파일들은 그대로 보존된다. stderr에 backup failure 메시지 출력
- **Test type:** integration (`internal/cli/update_test.go::TestBackup_FailureAborts`)

### AC-UPC-011 (REQ-UPC-011): MANIFEST.json 생성

- **Given:** 백업이 성공적으로 완료된 상태
- **When:** `.moai/backup/agency-{ISO8601}/MANIFEST.json` 파일을 read + parse
- **Then:** 다음 필드가 모두 존재한다: `original_paths` (배열), `file_hashes` (객체, path→SHA-256), `deletion_timestamp` (ISO8601), `authorized_by_spec` (= "SPEC-V3R3-UPDATE-CLEANUP-001"), `entries[*].classification` ("PristineDeprecated" / "UserModifiedDeprecated" / "UnverifiedDeprecated"), `entries[*].symlink_target` (symlink 항목인 경우, 그 외 omit)
- **Test type:** integration (`internal/cli/update_test.go::TestBackup_ManifestSchema`)

---

## 4. User Confirmation (REQ-UPC-012~014)

### AC-UPC-012 (REQ-UPC-012): high-risk 분류로 ConfirmMerge 호출

- **Given:** Deprecated paths 탐지 완료 (분류 결과 무관)
- **When:** Cleanup 단계 진입 시
- **Then:** `merge.ConfirmMerge(analysis)` (`internal/merge/confirm.go:623`)가 다음과 같이 호출된다 — `analysis.RiskLevel == "high"`, `analysis.HasConflicts == true`, 모든 `analysis.Files[i].RiskLevel == "high"`. 사용자에게 표시되는 prompt에 (1) 백업 위치, (2) 삭제 대상 파일 목록, (3) SPEC ID 참조, (4) 각 파일의 분류(Pristine/UserModified/Unverified)가 포함된다
- **Test type:** integration with mock (`internal/cli/update_test.go::TestCleanup_HighRiskPrompt`)

### AC-UPC-013 (REQ-UPC-013): 거부 시 deferral

- **Given:** 사용자가 confirmation prompt에서 거부 응답 (`ConfirmMerge`가 `(false, nil)` 반환)
- **When:** Cleanup phase 종료
- **Then:** Deprecated 파일이 모두 보존되며, "cleanup deferred by user" stdout 메시지 출력. 백업 디렉터리는 생성되었다면 보존(롤백 아님)
- **Test type:** integration (`internal/cli/update_test.go::TestCleanup_UserConfirmationDeclined`)

### AC-UPC-014 (REQ-UPC-014): auto-confirm 플래그 동작 (Pristine만)

- **Given:** `moai update --auto-confirm-cleanup` 호출이고, 모든 deprecated 파일이 `PristineDeprecated`로 분류됨
- **When:** Deprecated paths 탐지됨
- **Then:** Confirmation prompt가 표시되지 않고, 백업 + 삭제가 즉시 진행된다. stdout에 "auto-confirm enabled" 표시. **만약 `UnverifiedDeprecated` 파일이 1개라도 있으면 auto-confirm은 무시되고 prompt가 표시된다 (REQ-UPC-015c, 별도 AC-UPC-015c에서 검증)**
- **Test type:** integration (`internal/cli/update_test.go::TestCleanup_AutoConfirmFlag`)

---

## 5. Customization Safety — 삼분 분류 (REQ-UPC-015 / 015a / 015b / 015c / 016 / 017 / 018)

### AC-UPC-015 (REQ-UPC-015): manifest 조회 + SHA-256 hash 계산

- **Given:** 사용자 프로젝트에 deprecated 파일이 존재
- **When:** Cleanup 스캔 phase
- **Then:** 각 deprecated 파일에 대해 `manifest.Manager.GetEntry(path)` 호출과 `manifest.HashFile(absPath)` 호출이 정확히 1회 발생한다 (테스트 mock으로 호출 검증)
- **Test type:** unit (`internal/cli/update_test.go::TestCleanup_ManifestLookupAndHash`)

### AC-UPC-015a (REQ-UPC-015a): PristineDeprecated 분류

- **Given:** Deprecated 파일이 존재 + manifest entry 존재 + 현재 hash == `entry.DeployedHash`
- **When:** Cleanup 진행
- **Then:** 파일이 `PristineDeprecated`로 분류되어 backup `MANIFEST.json`의 `classification` 필드에 기록된다. `--auto-confirm-cleanup` 시 자동 삭제 가능. stderr 경고 미출력.
- **Test type:** integration (`internal/cli/update_test.go::TestCleanup_PristineDeprecated`)

### AC-UPC-015b (REQ-UPC-015b): UserModifiedDeprecated 분류 + 경고

- **Given:** Deprecated 파일이 존재 + manifest entry 존재 + 현재 hash != `entry.DeployedHash` (사용자 수정 시뮬레이션)
- **When:** Cleanup 진행
- **Then:** 파일이 `UserModifiedDeprecated`로 분류되고, stderr에 `"WARNING: user-modified deprecated file detected: <path>"` 출력. 백업은 정상 진행되며 backup `MANIFEST.json`의 `classification`에 기록.
- **Test type:** integration (`internal/cli/update_test.go::TestCleanup_UserModifiedDeprecated`)

### AC-UPC-015c (REQ-UPC-015c): UnverifiedDeprecated 분류 + auto-confirm 차단

- **Given:** Deprecated 파일이 존재하나 `manifest.Manager.GetEntry(path)`가 `(nil, false)` 반환 (manifest entry 부재 — 기존 사용자 케이스)
- **When:** Cleanup 진행 (auto-confirm 플래그 유무 모두 검증)
- **Then:**
  - 파일이 `UnverifiedDeprecated`로 분류된다
  - 백업은 mandatory로 수행된다
  - `--auto-confirm-cleanup`가 설정되어 있어도 `ConfirmMerge` prompt가 호출된다 (auto bypass 차단)
  - stderr에 `"WARNING: unverified deprecated file (no provenance record): <path>"` 출력
  - backup `MANIFEST.json`의 `classification`에 기록
- **Test type:** integration (`internal/cli/update_test.go::TestCleanup_UnverifiedDeprecated`)

### AC-UPC-016 (REQ-UPC-016): warning 없는 무음 삭제 금지

- **Given:** `UserModifiedDeprecated` 또는 `UnverifiedDeprecated` 분류된 파일이 1개 이상 존재
- **When:** Cleanup 종료 후 stderr 캡처
- **Then:** stderr 출력에 해당 파일 경로를 포함한 "WARNING" 라벨 라인이 분류된 파일 수만큼 존재한다
- **Test type:** integration (`internal/cli/update_test.go::TestCleanup_WarningRequired`)

### AC-UPC-017 (REQ-UPC-017): MANIFEST.json에 classification 기록

- **Given:** 백업 완료 + 3가지 분류가 모두 포함된 cleanup 결과
- **When:** `.moai/backup/agency-{ts}/MANIFEST.json` parse
- **Then:** `entries` 배열의 각 항목에 `classification` 필드 (3가지 enum 값 중 하나)가 정확히 기록된다
- **Test type:** integration (`internal/cli/update_test.go::TestCleanup_ClassificationInManifest`)

### AC-UPC-018 (REQ-UPC-018, restored v0.2.1): `.moai-skip-cleanup` opt-out marker

- **Given:**
  - 사용자 프로젝트에 `.claude/commands/agency/agency.md` (deprecated path) 존재 + 매니페스트 hash 일치 (Pristine 분류 시뮬레이션)
  - 사용자가 `.claude/commands/agency/.moai-skip-cleanup` marker file을 직접 생성 (manual override 시뮬레이션)
- **When:** `moai update --auto-confirm-cleanup` 실행
- **Then:**
  - `.claude/commands/agency/agency.md`가 backup, deletion 흐름에서 제외된다 (파일 보존)
  - stderr에 `[INFO] cleanup skipped due to .moai-skip-cleanup marker: .claude/commands/agency/agency.md` (또는 부모 디렉터리 경로) 출력
  - `.moai/logs/update-cleanup-{ts}.jsonl`의 `user_opt_out_paths` 필드에 해당 경로가 기록된다
  - 다른 deprecated path가 있으면 정상 처리에 영향 없음 (격리 검증)
  - marker file 자체는 사용자 데이터로 보존됨 (삭제하지 않음)
- **Test type:** integration (`internal/cli/update_test.go::TestCleanup_SkipMarkerHonored`)

---

## 6. Regression Test Suite (REQ-UPC-019~021)

### AC-UPC-019 (REQ-UPC-019): 멱등성 테스트 존재

- **Given:** `internal/template/deployer_test.go` 컴파일 완료
- **When:** `go test -run TestDeployer_Idempotent` 실행
- **Then:** 테스트가 존재하고 PASS한다
- **Test type:** test infrastructure check (CI gate)

### AC-UPC-020 (REQ-UPC-020): cleanup 회귀 테스트 6 시나리오

- **Given:** `internal/cli/update_e2e_test.go` 컴파일 완료
- **When:** `go test -run TestE2E_Scenario` 실행
- **Then:** 시나리오 A–F (§8) 6개 모두 PASS
- **Test type:** test infrastructure check (CI gate)

### AC-UPC-021 (REQ-UPC-021): t.TempDir() isolation 준수

- **Given:** 모든 신규/수정 테스트 파일
- **When:** 테스트 실행 후 프로젝트 루트 + 사용자 홈 디렉터리 검사
- **Then:** 테스트가 생성한 어떤 파일도 `t.TempDir()` 외부에 잔존하지 않는다 (`go test -count=1`로 캐시 우회 검증)
- **Test type:** manual + CI lint (sweep script)

---

## 7. Observability and Environment Safety (REQ-UPC-022~026, v0.2.0 신규)

### AC-UPC-022 (REQ-UPC-022): cleanup phase 텔레메트리

- **Given:** `moai update` 실행 (cleanup phase 진입 — deprecated paths 존재 또는 부재 모두)
- **When:** Cleanup phase 종료 직후
- **Then:**
  - `.moai/logs/update-cleanup-{ISO8601}.jsonl` 파일이 생성되며, 단일 JSON line이 append된다
  - JSON 객체 키: `atomic_write_used` (bool), `pre_update_suffix2_files` ([string]), `backup_outcome` ("success"/"skipped"/"failed"), `backup_path` (string, optional), `cleanup_outcome` ("completed"/"deferred"/"aborted"), `user_opt_out_paths` ([string], v0.2.1 — REQ-UPC-018 marker로 skip된 경로)
  - 동일 line이 stderr에도 mirror 출력된다
  - **외부 네트워크 호출이 발생하지 않는다** — v0.2.2 D-02-13 정확화: 검증 메커니즘은 custom `http.RoundTripper` interception 패턴 사용. 테스트가 `http.DefaultTransport`를 zero-Dial-허용 round-tripper(모든 host의 `Dial` 시도를 기록 + 즉시 error 반환)로 교체한 뒤 `moai update --enable-telemetry` 실행. assertion: 테스트 종료 시점에 round-tripper의 dial counter가 0임을 검증. (이전 버전의 `httptest.NewServer 401` 방식은 단일 endpoint에 대한 reply만 검증할 뿐 임의 host로의 dial 시도를 차단하지 못했다.)
- **Test type:** integration (`internal/cli/update_test.go::TestCleanup_TelemetryEmitted`)

### AC-UPC-022b (REQ-UPC-022 permission denied, v0.2.1, D-02-03): telemetry log 디렉터리 write 실패 graceful 처리

- **Given:**
  - 사용자 프로젝트에 deprecated paths 존재 (Pristine 분류 시뮬레이션)
  - `.moai/logs/` 디렉터리에 chmod 0500 적용 (또는 부모 `.moai/` chmod 0500) — write permission 부재
- **When:** `moai update --auto-confirm-cleanup` 실행
- **Then:**
  - stderr에 `[WARN] telemetry persistence skipped: <reason>` (예: `permission denied: .moai/logs`) 단일 라인 출력
  - `.moai/logs/update-cleanup-*.jsonl` 파일이 생성되지 **않는다**
  - stderr mirror로 telemetry JSON line 자체는 여전히 emit (관측성 보존)
  - **`moai update` 작업은 정상 완료** (exit code 0) — telemetry persistence 실패가 update 자체를 차단하지 않음
  - cleanup 본체 (backup + deletion)는 정상 진행됨
- **Test type:** integration (`internal/cli/update_test.go::TestCleanup_TelemetryPermissionDenied`)
- **Verification note:** Windows 환경에서는 ACL 또는 read-only attribute로 동등 시뮬레이션. 테스트는 `runtime.GOOS == "windows"` 시 `icacls` 또는 `os.Chmod(0444)` 활용.

### AC-UPC-023 (REQ-UPC-023): symbolic link 안전 처리

#### AC-UPC-023a (기본 케이스): 정상 symbolic link

- **Given:** 사용자 프로젝트에 `.claude/commands/agency/agency.md`가 외부 디렉터리(예: `/tmp/external-rules/agency.md`)를 가리키는 정상 symbolic link로 존재 (target 파일 존재)
- **When:** `moai update --auto-confirm-cleanup` 실행 (이 link는 `PristineDeprecated`로 분류 시뮬레이션)
- **Then:**
  - backup 디렉터리에 link 자체가 보존되거나 (또는) `MANIFEST.json`의 entry에 `symlink_target: "/tmp/external-rules/agency.md"` 필드가 기록된다
  - **링크 target 파일(`/tmp/external-rules/agency.md`)은 backup 디렉터리에 복사되지 않는다** (외부 콘텐츠 누수 방지)
  - 삭제 시 `os.Remove`로 link 자체만 제거되며, target 파일은 보존된다
- **Test type:** integration (`internal/cli/update_test.go::TestCleanup_SymlinkNotFollowed`)

#### AC-UPC-023b (broken symlink, v0.2.1, D-02-08): 끊어진 symbolic link

- **Given:** `.claude/commands/agency/agency.md`가 존재하지 않는 target(`/tmp/missing-target.md`)을 가리키는 broken symlink로 존재
- **When:** `moai update --auto-confirm-cleanup` 실행
- **Then:**
  - link 자체는 backup이 시도되지 않거나, MANIFEST entry만 기록 (콘텐츠 백업 없음)
  - `MANIFEST.json` entry에 `symlink_target: "/tmp/missing-target.md"` AND `symlink_target_status: "broken"` 두 필드 기록
  - link 자체는 `os.Remove`로 제거됨 (missing target은 처음부터 존재하지 않으므로 영향 없음)
  - update 작업은 정상 완료 (broken symlink가 update를 차단하지 않음)
- **Test type:** integration (`internal/cli/update_test.go::TestCleanup_SymlinkBroken`)

#### AC-UPC-023c (target collision, v0.2.1, D-02-08): symlink target이 별도 deprecated entry

- **Given:**
  - `.claude/commands/agency/agency.md`가 `.claude/commands/agency/brief.md`를 가리키는 symlink (둘 다 deprecated registry에 등록)
  - 두 entry 모두 manifest hash 일치 (Pristine 시뮬레이션)
- **When:** `moai update --auto-confirm-cleanup` 실행
- **Then:**
  - `agency.md` (link)는 link entry로 처리: backup MANIFEST에 `symlink_target: "...brief.md"` 기록
  - `brief.md` (target)는 자체 deprecated entry로 처리: backup에 단 1회 콘텐츠 복사
  - **`brief.md` 콘텐츠가 backup에 중복 복사되지 않음** (double-count 방지 검증: backup 디렉터리 내 `brief.md` 파일 개수 == 1)
- **Test type:** integration (`internal/cli/update_test.go::TestCleanup_SymlinkTargetCollision`)

#### AC-UPC-023d (Windows fallback, v0.2.1, D-02-08): symlink 미지원 환경

- **Given:** Windows CI 러너 환경 (관리자 권한 부재 시뮬레이션 또는 reparse point 미지원 환경)
- **When:** `moai update` 실행
- **Then:**
  - stderr에 `[INFO] symlink detection unsupported on this platform` 단일 라인 출력
  - 모든 deprecated paths가 일반 파일로 처리됨 (symlink 분기 진입 안 함)
  - update 작업은 정상 완료
- **Test type:** integration (`internal/cli/update_test.go::TestCleanup_SymlinkWindowsFallback`) — runtime.GOOS == "windows" 시에만 실행

### AC-UPC-024 (REQ-UPC-024, extended v0.2.1 D-02-02): backup AND logs self-reference 차단

- **Given:**
  - 사용자 프로젝트에 `.moai/backup/agency-2026-04-30T00-00-00Z/.claude/commands/agency/agency.md` (이전 cleanup의 백업 산출물) 존재
  - 사용자 프로젝트에 `.moai/logs/update-cleanup-2026-04-30T00-00-00Z.jsonl` (이전 telemetry 출력) 존재
  - 두 경로 모두 deprecated registry pattern과 매칭될 가능성 시뮬레이션 (예: registry에 `**/agency.md` 와일드카드 패턴이 있다면 `.moai/backup/.../agency.md`도 매칭될 수 있음)
- **When:** `moai update` 두 번째 실행
- **Then:**
  - 두 번째 cleanup 스캔이 `.moai/backup/` AND `.moai/logs/` 하위 경로를 deprecated path로 인식하지 않는다
  - 새로운 백업 디렉터리가 이전 백업/로그를 백업하는 무한 누적 패턴이 발생하지 않는다
  - `.moai/backup/` 및 `.moai/logs/` 자체는 유지되며 변경되지 않는다
- **Test type:** integration (`internal/cli/update_test.go::TestCleanup_BackupSelfReferenceSkipped` + `TestCleanup_LogsSelfReferenceSkipped`)

### AC-UPC-025 (REQ-UPC-025): 등록되었으나 부재한 파일 silent skip

- **Given:** `internal/defs/dirs.go::DeprecatedPaths`에 `.claude/commands/agency/agency.md`가 등록되어 있으나, 사용자 프로젝트에는 해당 파일이 부재 (원래부터 사용 안 함)
- **When:** `moai update` 실행
- **Then:**
  - 해당 entry에 대해 어떤 분류, backup, warning, telemetry log entry도 발생하지 않는다 (silent skip)
  - 다른 deprecated path 처리에는 영향 없음
- **Test type:** integration (`internal/cli/update_test.go::TestCleanup_DeletedFileSkipped`)

### AC-UPC-026 (REQ-UPC-026): case-insensitive FS 매칭

#### AC-UPC-026a (probe success — case-insensitive): macOS APFS 기본

- **Given:** macOS APFS (case-insensitive) 환경에서 사용자 프로젝트에 `.claude/commands/agency/Agency.md` (대문자) 파일 존재. registry는 `.claude/commands/agency/agency.md` (소문자)로 등록.
- **When:** `moai update` 실행
- **Then:**
  - cleanup phase 시작 시 `.moai-fscase-probe` 파일을 생성하고 대소문자 변형으로 stat 시도 → case-insensitive로 판정
  - `Agency.md`가 deprecated path로 매칭되어 분류 + 백업 + 삭제 흐름 진입
  - probe 파일은 cleanup 종료 시 정리됨
- **Test type:** integration (`internal/cli/update_test.go::TestCleanup_CaseInsensitiveFS_macOS`)

#### AC-UPC-026b (probe success — case-sensitive): Linux ext4 기본

- **Given:** Linux ext4 (case-sensitive) 환경에서 사용자 프로젝트에 `.claude/commands/agency/Agency.md` (대문자) 파일 존재. registry는 `agency.md` (소문자)로 등록.
- **When:** `moai update` 실행
- **Then:**
  - probe가 case-sensitive로 판정
  - `Agency.md`는 registry의 `agency.md`와 매칭되지 않음 (silent skip)
  - probe 파일은 정리됨
- **Test type:** integration (`internal/cli/update_test.go::TestCleanup_CaseInsensitiveFS_Linux`)

#### AC-UPC-026c (probe failure fallback, v0.2.1, D-02-07): probe 자체가 실패하는 환경

- **Given:** read-only filesystem 또는 write permission 부재 시뮬레이션 (e.g., chmod 0500 on project root parent)
- **When:** `moai update` 실행 (cleanup phase 진입 시도)
- **Then:**
  - probe (`.moai-fscase-probe` 생성) 실패 → ENOPERM 또는 EROFS
  - stderr에 `[INFO] case-sensitivity probe failed; defaulting to case-sensitive matching: <reason>` 단일 라인 출력
  - cleanup 흐름은 case-sensitive 모드로 fallback하여 정상 진행
  - update 작업은 정상 완료
- **Test type:** integration (`internal/cli/update_test.go::TestCleanup_CaseProbeFailureFallback`)

#### AC-UPC-026d (probe frequency, v0.2.1, D-02-07): invocation당 정확히 1회

- **Given:** `moai update` 단일 invocation 내에서 deprecated path 다수 처리 (3+ 파일)
- **When:** cleanup 실행
- **Then:**
  - probe 함수 호출이 정확히 1회 (counter mock 또는 file creation observer로 검증)
  - 결과가 caching되지 않음 (다음 invocation에서 재실행됨 — 별도 invocation에서 호출 카운트 reset 검증)
- **Test type:** unit (`internal/cli/update_test.go::TestCleanup_CaseProbeRunsOncePerInvocation`)

#### AC-UPC-026e (APFS case-sensitive variant, v0.2.1, D-02-07): macOS의 case-sensitive volume

- **Given:** macOS 환경 + APFS case-sensitive variant volume에서 프로젝트 실행 (hdiutil로 case-sensitive dmg 생성 또는 사전 mount 시뮬레이션)
- **When:** `moai update` 실행
- **Then:**
  - probe가 case-sensitive로 판정 (GOOS == "darwin"임에도 불구하고)
  - `Agency.md` (대문자)는 `agency.md` (소문자)와 매칭되지 않음
  - **GOOS 기반 가정이 아니라 probe 결과가 authoritative임을 검증** (assertion: matching mode == "case-sensitive" on darwin)
- **Test type:** integration (`internal/cli/update_test.go::TestCleanup_AfpsCaseSensitiveVariant`) — runtime.GOOS == "darwin" 시에만 실행, dmg fixture 또는 skip with reason

---

## 8. End-to-End Regression Scenarios (사용자 시나리오)

### Scenario A: 사용자 프로젝트에 agency/ 폴더 존재 → moai update → 백업 + 제거 (Pristine)

- **Given:**
  - 사용자 프로젝트 루트에 `.claude/commands/agency/agency.md`, `.claude/commands/agency/brief.md`, `.claude/rules/agency/constitution.md` 존재 (pristine, 매니페스트 hash 일치)
  - `.moai/manifest.json`에 해당 파일들이 `manifest.Manager.Track(path, manifest.TemplateManaged, hash)`로 등록되어 `DeployedHash`가 현재 hash와 일치
- **When:** `moai update --auto-confirm-cleanup` 실행
- **Then:**
  - `.moai/backup/agency-{timestamp}/.claude/commands/agency/agency.md` (외 2개) 백업본 존재
  - 원본 파일 3개 모두 삭제됨
  - `.moai/backup/agency-{timestamp}/MANIFEST.json` 생성됨, 모든 entry의 `classification == "PristineDeprecated"`
  - stdout에 "Removed 3 deprecated paths (backup at .moai/backup/agency-{ts}/)" 출력
  - `.moai/logs/update-cleanup-{ts}.jsonl` 생성, `cleanup_outcome: "completed"`
- **Test type:** E2E (`internal/cli/update_e2e_test.go::TestE2E_ScenarioA_AgencyCleanup`)

### Scenario B: agency/ 부재 → moai update → no-op

- **Given:** 사용자 프로젝트에 어떤 agency 관련 파일도 부재
- **When:** `moai update` 실행
- **Then:**
  - `.moai/backup/` 디렉터리가 생성되지 않거나, 존재하더라도 새 `agency-*` 하위 디렉터리가 추가되지 않음
  - stdout에 cleanup 관련 메시지 미출력
  - 종료 코드 0
  - `.moai/logs/update-cleanup-{ts}.jsonl`은 여전히 생성됨, `backup_outcome: "skipped"`, `cleanup_outcome: "completed"`
- **Test type:** E2E (`internal/cli/update_e2e_test.go::TestE2E_ScenarioB_NoOp`)

### Scenario C: 사용자가 commands/agency/agency.md 수정 → moai update → 백업 + 경고 (UserModified)

- **Given:**
  - 사용자가 `.claude/commands/agency/agency.md`에 커스텀 라인 추가
  - `.moai/manifest.json`의 `entry.DeployedHash`는 수정 전 hash
- **When:** `moai update --auto-confirm-cleanup` 실행
- **Then:**
  - 백업본이 `.moai/backup/agency-{ts}/.claude/commands/agency/agency.md`에 보존 (사용자 커스텀 라인 포함)
  - stderr에 "WARNING: user-modified deprecated file detected: .claude/commands/agency/agency.md" 출력
  - 원본 파일은 삭제됨 (백업으로 복구 가능)
  - MANIFEST.json의 해당 entry의 `classification == "UserModifiedDeprecated"`
- **Test type:** E2E (`internal/cli/update_e2e_test.go::TestE2E_ScenarioC_UserModified`)

### Scenario D: moai update 두 번 연속 → ` 2` 접미어 파일 미발생

- **Given:** 깨끗한 사용자 프로젝트
- **When:**
  - `moai update` 1차 실행 (성공)
  - `moai update` 2차 실행 (성공)
- **Then:**
  - 프로젝트 트리 walk 결과 ` 2` 접미어 파일이 0개
  - `.moai-tmp` 접미어 파일이 0개
  - 1차/2차 결과 디렉터리가 byte-identical (file hash 비교)
  - 1차/2차 모두 `.moai/logs/update-cleanup-*.jsonl` 생성
- **Test type:** E2E (`internal/cli/update_e2e_test.go::TestE2E_ScenarioD_DoubleRunIdempotent`)

### Scenario E: 동시 moai update 실행 → file lock으로 중복 write 방지

- **Given:** 사용자 프로젝트
- **When:**
  - 첫 번째 `moai update` 프로세스가 lock acquire에 성공한 직후, lock 파일이 디스크에 존재함을 fsync로 보장
  - 두 번째 `moai update` 프로세스 호출 — lock acquisition 단계에서 deterministic 실패 (timing-based race가 아니라 lock file 존재 자체가 결정적 신호)
- **Then:**
  - 두 번째 프로세스가 즉시 종료 (exit code != 0)
  - stderr에 "another moai update is in progress (PID: <N>)" 메시지 출력
  - 첫 번째 프로세스는 정상 완료
  - 첫 번째 프로세스 종료 후 lock 파일 자동 정리
- **Test type:** E2E (`internal/cli/update_e2e_test.go::TestE2E_ScenarioE_ConcurrentLock`)
- **Verification note (D-07):** 본 시나리오는 deterministic lock acquisition test (lock 파일 존재를 사전 보장한 상태에서 두 번째 호출)로 구현되며, 두 프로세스가 race 상태에서 누가 먼저 lock을 잡는지 timing에 의존하지 않는다. 상세 테스트 전략은 plan-auditor v2 검토 시점에 추가 정형화한다.

### Scenario F: manifest entry 부재 → UnverifiedDeprecated 흐름 (v0.2.0 신규)

- **Given:**
  - 사용자 프로젝트에 `.claude/commands/agency/agency.md` 존재
  - `.moai/manifest.json`이 존재하지 않거나, 존재하더라도 해당 파일에 대한 entry 부재 (기존 사용자의 일반적 케이스)
- **When:** `moai update --auto-confirm-cleanup` 실행
- **Then:**
  - 파일이 `UnverifiedDeprecated`로 분류된다
  - **`--auto-confirm-cleanup` 플래그가 무시되고** `ConfirmMerge` prompt가 호출된다
  - stderr에 "WARNING: unverified deprecated file (no provenance record): .claude/commands/agency/agency.md" 출력
  - 사용자가 prompt에서 승인하면 백업 + 삭제, 거부하면 보존
  - 승인 시 backup `MANIFEST.json`의 해당 entry의 `classification == "UnverifiedDeprecated"`
- **Test type:** E2E (`internal/cli/update_e2e_test.go::TestE2E_ScenarioF_UnverifiedDeprecated`)

---

## 9. NFR Benchmark Verification (NFR-UPC-P1, v0.2.0 신규)

### AC-NFR-P1 (NFR-UPC-P1, v0.2.1 benchstat 통계 강화 — D-02-05): atomic write overhead ≤ 5% with statistical significance

- **Given:**
  - `internal/template/deployer_bench_test.go`에 `Benchmark_UpdateCleanup_Baseline` (atomic write 미적용 — v2.14.0 동작 시뮬레이션) + `Benchmark_UpdateCleanup_AtomicWrite` (atomic write 적용) 두 벤치마크 정의
  - 합성 프로젝트 fixture: 100 files, 5 deprecated paths, 0 symlinks (재현성 우선)
  - **공식 게이트 OS**: `ubuntu-latest` (가장 재현성 높음). macOS / Windows 결과는 정보 제공용 (게이트 제외).
  - `benchstat` 도구 설치: `go install golang.org/x/perf/cmd/benchstat@latest`
- **When:**
  - CI (ubuntu-latest)가 다음 절차 실행:
    1. `go test -bench=Benchmark_UpdateCleanup -benchmem -count=10 ./internal/template/... -run=^$ | tee bench-result.txt` (10 independent runs per benchmark)
    2. baseline + atomic 결과를 별도 파일로 split
    3. `benchstat -alpha 0.05 baseline.txt atomic.txt | tee bench-report.txt` (Welch's t-test 또는 Mann-Whitney U test, p-value 자동 계산)
    4. `scripts/benchmark-overhead-check.sh bench-report.txt` — delta_pct + p_value 파싱 후 게이트 검증
- **Then:**
  - **게이트 통과 조건 (둘 중 하나)**:
    - Case A: `delta_pct ≤ 5%` (overhead 허용 범위 내) → 통과
    - Case B: `p_value ≥ 0.05` (통계적으로 유의하지 않음 — noise 범위) → 통과 (delta 무관)
  - **게이트 실패 조건**: `delta_pct > 5%` AND `p_value < 0.05` 동시 만족 → exit 1
  - `benchstat` 출력 (`bench-report.txt`)이 PR 본문에 자동 코멘트로 attach (별도 GitHub Actions 워크플로우)
  - macOS / Windows 벤치마크 결과는 PR 코멘트에 별도 섹션으로 표시되나 게이트 영향 없음
- **Test type:** benchmark + CI statistical gate (`scripts/benchmark-overhead-check.sh`)
- **Verification note:** p-value < 0.05 임계값은 표준 통계적 유의성 기준. 10회 반복은 benchstat 권장 최소값 (작은 샘플 사이즈일수록 outlier에 취약하므로 향후 30회까지 확장 가능 — 현재는 CI 시간 vs 정확성 trade-off로 10회 유지).

---

## 10. Cross-OS Verification Matrix (v0.2.2 신설, D-02-12)

REQ별 OS 검증 책임 매트릭스. plan.md §5 DoD가 요구하는 CI 3-플랫폼(`ubuntu-latest`, `macos-latest`, `windows-latest`) 그린 조건과 결합하여 실행.

| REQ | ubuntu (POSIX) | macos (POSIX + APFS) | windows (NTFS + reparse) | 비고 |
|---|---|---|---|---|
| REQ-UPC-022 (telemetry) | 필수 | 필수 | 필수 | jsonl 파일 생성 + stderr mirror 모두 OS 무관 검증 |
| REQ-UPC-022b (permission denied) | 필수 (`chmod 0500`) | 필수 (`chmod 0500`) | 필수 (`icacls /deny` 또는 read-only attribute) | Windows는 POSIX chmod 미지원이므로 ACL 기반 동등 시뮬레이션 |
| REQ-UPC-023a (정상 symlink) | 필수 | 필수 | 정보 제공용 (junction point) | POSIX symlink는 ubuntu/macos에서 full coverage |
| REQ-UPC-023b (broken symlink) | 필수 | 필수 | 정보 제공용 | broken symlink 동작은 OS 별 ENOENT 처리 일관성 검증 |
| REQ-UPC-023c (target collision) | 필수 | 필수 | 정보 제공용 | double-count 방지는 OS 무관 로직 |
| REQ-UPC-023d (Windows fallback) | skip | skip | **필수 (실제 검증)** | `runtime.GOOS == "windows"` 게이트, junction/reparse 환경 또는 권한 부재 시뮬레이션 |
| REQ-UPC-024 (backup + logs self-ref) | 필수 | 필수 | 필수 | path prefix matching은 OS 무관 |
| REQ-UPC-025 (deleted file silent) | 필수 | 필수 | 필수 | OS 무관 |
| REQ-UPC-026a (case-insensitive APFS) | skip | **필수** | skip | macOS APFS 기본 case-insensitive 환경 |
| REQ-UPC-026b (case-sensitive ext4) | **필수** | skip | skip | ubuntu ext4 기본 case-sensitive 환경 |
| REQ-UPC-026c (probe failure fallback) | 필수 | 필수 | 필수 | read-only FS 시뮬레이션 (`chmod 0500` 또는 동등) |
| REQ-UPC-026d (probe frequency once) | 필수 | 필수 | 필수 | counter mock 기반, OS 무관 |
| REQ-UPC-026e (APFS case-sensitive variant) | skip | **필수 (조건부)** | skip | macOS의 case-sensitive APFS variant 환경 (hdiutil dmg fixture 필요, fixture 부재 시 `t.Skip` with reason) |

**Verification rules:**
- "필수" = 해당 OS에서 테스트 실행 + PASS 필수 (CI gate fail-closed)
- "정보 제공용" = 해당 OS에서 실행되더라도 결과는 PR 코멘트에만 첨부, 게이트 영향 없음
- "skip" = 해당 OS에서 `t.Skip("not applicable on <OS>")` 명시적 호출
- "조건부" = fixture (예: hdiutil dmg) 또는 환경 변수 존재 시에만 실행, 부재 시 skip

**Constitution C1 정합성**: 본 매트릭스는 16-language neutrality (CLAUDE.local.md §15)와 OS neutrality를 분리한다. 16-language는 사용자 프로젝트 언어 동등 취급, OS neutrality는 `moai update` 자체의 cross-OS 거동 보장.

---

## 11. Quality Gate Summary

- 총 AC 개수: **28개** (AC-UPC-001~005, 006~008, 009~011, 012~014, 015 / 015a / 015b / 015c / 016 / 017 / **018**, 019~021, 022 / **022b** / 023 / 023b / 023c / 023d / 024 / 025 / 026 / 026b / 026c / 026d / 026e + AC-NFR-P1)
  - 주요 카운트: AC-UPC-NNN 27개 + AC-NFR-P1 1개 = 28개. (sub-cases 023a~d, 026a~e는 단일 REQ에 종속된 시나리오 분할로 각각 독립 검증되나 상위 AC ID로 집계)
- E2E 시나리오: **6개** (A~F, 변경 없음)
- NFR 벤치마크: **1개** (AC-NFR-P1, benchstat 통계 게이트)
- 테스트 파일 분포 (v0.2.1 기준):
  - `internal/template/deployer_test.go`: 4 tests (atomic write 관련)
  - `internal/template/deployer_bench_test.go`: 2 benchmarks (NFR-UPC-P1)
  - `internal/defs/dirs_test.go`: 1 test (registry)
  - `internal/cli/update_test.go`: 약 22 tests (cleanup, backup, confirmation, provenance 삼분, opt-out marker, telemetry + permission, symlink + 3 edges, dual self-reference, deleted, case-insensitive + 4 edges)
  - `internal/cli/update_e2e_test.go`: 6 tests (시나리오 A~F)
  - `scripts/benchmark-overhead-check.sh`: CI statistical gate (delta + p-value)
- TRUST 5 매핑:
  - **Tested:** 모든 REQ에 최소 1개 AC 매핑 + 6개 E2E 시나리오 + 1개 NFR 벤치마크 (statistical) + opt-out marker AC + 7개 edge case sub-AC
  - **Readable:** EARS keyword 일관 사용 (SHALL/WHEN/IF-THEN/WHILE/WHERE)
  - **Unified:** Given-When-Then 형식 일관
  - **Secured:** 사용자 데이터 보호 (백업 의무, 커스터마이징 감지, 동시 실행 차단, symlink follow 금지, broken symlink 안전 처리, backup + logs self-reference 차단, telemetry permission graceful, 외부 네트워크 호출 금지, 사용자 manual opt-out 메커니즘)
  - **Trackable:** 각 AC에 SPEC ID 및 REQ ID 명시, telemetry jsonl로 출시 후 사후 검증 가능 (user_opt_out_paths 필드 포함)
