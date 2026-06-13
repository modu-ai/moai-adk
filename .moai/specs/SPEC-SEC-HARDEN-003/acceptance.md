# SPEC-SEC-HARDEN-003 — 인수 기준 (acceptance.md)

> **이 문서가 AC SSOT다.** GEARS 요구사항에서 도출. cycle_type=tdd reproduction-first: 각 결함의 재현(reproduction) AC가 fix 전에 RED로 확인된다.
> 테스트 함수명은 실제 작성 예정 이름이며 sibling 부분문자열 충돌이 없다.

## AC 매트릭스 개요

| AC ID | 결함 | 종류 | REQ | 테스트 함수 |
|-------|------|------|-----|-------------|
| AC-SEC3-001a | C-F1 | reproduction | REQ-SEC3-001/002 | `TestRunMXScan_RejectsUncontainedFilePath` |
| AC-SEC3-001b | C-F1 | reproduction | REQ-SEC3-003 | `TestRunMXScan_RejectsUncontainedSidecarCWD` |
| AC-SEC3-002 | C-F1 | containment | REQ-SEC3-001/002/003/004 | (위 두 테스트의 post-fix GREEN + payload 검증) |
| AC-SEC3-003 | C-F1 | no-regression | REQ-SEC3-004(계약 보존) | `TestRunMXScan_AllowsInProjectPath` |
| AC-SEC3-004a | C-F2 | reproduction | REQ-SEC3-005 | `TestRestoreMoaiConfigLegacy_SkipsSymlinkEntry` |
| AC-SEC3-004b | C-F2 | reproduction | REQ-SEC3-007 | `TestRestoreMoaiConfigLegacy_RejectsTraversalTarget` |
| AC-SEC3-004c | C-F2 | reproduction (sibling) | REQ-SEC3-006 | `TestRestoreMoaiConfig_SkipsSymlinkEntry` |
| AC-SEC3-005 | C-F2 | containment | REQ-SEC3-005/006/007 | (위 세 테스트의 post-fix GREEN) |
| AC-SEC3-006 | C-F2 | no-regression | REQ-SEC3-008 | `TestRestoreMoaiConfigLegacy_AllowsRegularInConfigFile` |
| AC-SEC3-007 | 공통 | no-new-surface | REQ-SEC3-009 | grep 정적 가드 |

---

## C-F1 — MX 사이드카 봉쇄 AC cluster

### AC-SEC3-001a — (reproduction) 비격리 FilePath 재현

**Given** `file_changed` hook 입력에서 `input.FilePath`가 해소된 프로젝트 루트를 탈출하는 절대/`..` 경로일 때
**When** `runMXScan`이 호출되면
**Then** (fix 전 RED) 현재 코드는 루트 밖 파일을 스캔하고, (fix 후 GREEN) `runMXScan`은 스캔 없이 `slog.Warn` 로그 후 early return 한다.

- 테스트: `TestRunMXScan_RejectsUncontainedFilePath`
- RED 확인: fix 전 이 테스트가 봉쇄 부재로 실패함을 확인한다.

### AC-SEC3-001b — (reproduction) 비격리 사이드카 CWD 재현

**Given** `input.CWD`가 프로젝트 루트 밖을 가리켜 `stateDir = filepath.Join(input.CWD, ".moai", "state")`가 루트를 탈출할 때
**When** `runMXScan`이 호출되면
**Then** (fix 전 RED) 현재 코드는 루트 밖에 `.moai/state` 사이드카를 쓰고, (fix 후 GREEN) `runMXScan`은 사이드카 쓰기 없이 `slog.Warn` 로그 후 early return 한다.

- 테스트: `TestRunMXScan_RejectsUncontainedSidecarCWD`

### AC-SEC3-002 — (containment) fail-closed 봉쇄 + 계약 보존

**Given** AC-SEC3-001a/001b의 악의적 입력
**When** fix 적용 후 `runMXScan`이 호출되고 main `file_changed` handler가 응답을 반환하면
**Then**:
- 루트 밖 스캔/쓰기가 **발생하지 않는다**(봉쇄).
- 봉쇄 위반 시 `runMXScan`은 panic 하지 않고 early return 한다(fail-closed, NFR-SEC3-004).
- main handler는 고정 빈 payload를 반환한다 — 비동기 실패가 hook 응답에 전파되지 않는다(REQ-SEC3-004).
- 루트 해소는 `resolveProjectRootFromInputOrEnv`로 수행되며 `os.Getenv` 인라인이 없다.

검증 명령:
```bash
grep -rn "resolveProjectRootFromInputOrEnv" internal/hook/file_changed.go   # ≥1 match
grep -n "os.Getenv" internal/hook/file_changed.go                          # 0 match (NFR-SEC3-002)
```

### AC-SEC3-003 — (no-regression) 정상 in-project 경로

**Given** `input.FilePath`와 `input.CWD`가 모두 해소된 프로젝트 루트 내부의 정상 경로일 때
**When** `runMXScan`이 호출되면
**Then** 기존대로 MX 태그 스캔이 수행되고 `.moai/state` 사이드카가 갱신된다(봉쇄가 정상 경로를 막지 않는다).

- 테스트: `TestRunMXScan_AllowsInProjectPath`

---

## C-F2 — 백업 복원 봉쇄 AC cluster (레거시 + 모던 sibling)

### AC-SEC3-004a — (reproduction) 레거시 심볼릭 링크 엔트리

**Given** `backupDir` 안에 심볼릭 링크 백업 엔트리가 있을 때
**When** `restoreMoaiConfigLegacy`가 호출되면
**Then** (fix 전 RED) 현재 코드는 링크를 따라 `os.ReadFile` 하고, (fix 후 GREEN) `os.Lstat`로 심볼릭 링크를 검출하여 `os.ReadFile` 없이 스킵한다(REQ-SEC3-005).

- 테스트: `TestRestoreMoaiConfigLegacy_SkipsSymlinkEntry`
- 픽스처는 `t.TempDir()` 내에서 `os.Symlink`로 생성하고 자동 정리된다.

### AC-SEC3-004b — (reproduction) 레거시 traversal target 거부

**Given** 백업 `relPath`가 `targetPath := filepath.Join(configDir, relPath)`를 `configDir` 밖으로 탈출시킬 때
**When** `restoreMoaiConfigLegacy`가 쓰기를 시도하면
**Then** (fix 전 RED) 현재 코드는 `configDir` 밖에 `os.WriteFile` 하고, (fix 후 GREEN) `filepath.Rel(configDir, targetPath)`가 `..` prefix/탈출을 산출하면 쓰기 없이 거부한다(REQ-SEC3-007).

- 테스트: `TestRestoreMoaiConfigLegacy_RejectsTraversalTarget`

### AC-SEC3-004c — (reproduction, sibling) 모던 심볼릭 링크 엔트리

**Given** 모던 백업 트리(`sectionsBackupDir`) 안에 심볼릭 링크 엔트리가 있을 때
**When** `restoreMoaiConfig`이 호출되면
**Then** (fix 전 RED) 현재 코드는 링크를 따라 `os.ReadFile` 하고, (fix 후 GREEN) `os.Lstat`로 검출하여 `os.ReadFile` 없이 스킵한다(REQ-SEC3-006 — 모던 경로는 in-scope sibling).

- 테스트: `TestRestoreMoaiConfig_SkipsSymlinkEntry`

### AC-SEC3-005 — (containment) fail-closed 봉쇄

**Given** AC-SEC3-004a/004b/004c의 악의적 백업
**When** fix 적용 후 복원이 실행되면
**Then**:
- 심볼릭 링크 추종 읽기/쓰기가 **발생하지 않는다**(레거시·모던 양쪽).
- `configDir` 탈출 쓰기가 **발생하지 않는다**.
- 거부는 fail-closed(스킵/거부 + 로그)이며 panic 하지 않는다(NFR-SEC3-004).

검증 명령:
```bash
# symlink 가드: os.Lstat는 fix 전 0회 등장 → 신규 도입을 discriminate함
grep -rn "os.Lstat" internal/cli/update.go          # ≥1 match (symlink 가드)
# traversal 가드: filepath.Rel는 fix 전 이미 4곳(L778/L1538/L1977/L2049) 존재 →
# symbol-presence grep은 non-discriminating(vacuous)이므로 제외. 봉쇄는 행위 테스트로 검증:
#   AC-SEC3-004b TestRestoreMoaiConfigLegacy_RejectsTraversalTarget — configDir 탈출 target 거부
```

### AC-SEC3-006 — (no-regression) 정규 파일 + configDir 내부

**Given** 백업 엔트리가 정규 파일(심볼릭 링크 아님)이고 `targetPath`가 `configDir` 내부일 때
**When** `restoreMoaiConfigLegacy`(및 모던)가 호출되면
**Then** 기존 머지·복원 동작이 변경 없이 수행된다 — 정상 백업 복원이 봉쇄에 막히지 않는다(REQ-SEC3-008).

- 테스트: `TestRestoreMoaiConfigLegacy_AllowsRegularInConfigFile`

---

## 공통 AC

### AC-SEC3-007 — (no-new-surface) 새 표면 미도입

**Given** 본 SPEC 구현
**When** 변경 diff를 검사하면
**Then** 새 패키지·새 공개 cobra 플래그·새 공개 추상화가 도입되지 않는다 — 봉쇄는 기존 함수 내 additive guard로만 구현된다(REQ-SEC3-009).

검증 명령:
```bash
# specid 패키지 무변경 (C-F2가 specid를 import/수정하지 않음)
grep -rn "internal/cli/specid" internal/cli/update.go   # 0 match
# 새 cobra 플래그 미추가 (update.go에 신규 .Flags() 표면 없음)
git diff --stat internal/cli/update.go internal/hook/file_changed.go
```

---

## 품질 게이트 / Definition of Done

- [ ] M1·M2 모든 reproduction AC가 fix 전 RED로 확인됨.
- [ ] 모든 containment·no-regression AC가 fix 후 GREEN.
- [ ] `go test ./internal/hook/... ./internal/cli/...` exit 0.
- [ ] `go test -race ./internal/hook/...` exit 0 (비동기 goroutine).
- [ ] 변경 패키지 커버리지 회귀 없음(NFR-SEC3-003).
- [ ] `golangci-lint run` 0 finding.
- [ ] `GOOS=windows GOARCH=amd64 go build ./...` exit 0 (NFR-SEC3-005).
- [ ] `grep -n "os.Getenv" internal/hook/file_changed.go` → 0 match (NFR-SEC3-002).
- [ ] 비동기 구조·5s asyncDeadline·빈 payload 계약 무변경(NFR-SEC3-001) — 코드 리뷰 확인.
