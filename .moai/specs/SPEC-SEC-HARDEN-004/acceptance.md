# Acceptance Criteria — SPEC-SEC-HARDEN-004

> GEARS-anchored ACs. SSOT for AC count = this file. 2 mandatory reproduction ACs (F1 write-reject, F2 read-reject) + regression ACs asserting SEC-HARDEN-003 leaf guards still hold.
> Grep-style ACs use `$` end-anchor where applicable + `grep -r` `-E` flags; `while read` greps allow shellcheck `-r` flag idiom (L_ac_grep_read_r_flag).

## D. AC Matrix (10 ACs)

| AC | REQ | Severity | Type | Summary |
|----|-----|----------|------|---------|
| AC-SEC4-001 | REQ-SEC4-001 | MUST-PASS | reproduction (F1) | symlinked-parent-dir 쓰기 거부 |
| AC-SEC4-002 | REQ-SEC4-002 | MUST-PASS | parity | 공유 헬퍼 수정으로 양 walk 동시 봉쇄 |
| AC-SEC4-003 | REQ-SEC4-003 | regression | no-regression | F1 정상 복원 + leaf 가드 보존 |
| AC-SEC4-004 | REQ-SEC4-004 | MUST-PASS | reproduction (F2) | symlink-in-root scan-target 읽기 거부 |
| AC-SEC4-005 | REQ-SEC4-005 | MUST-PASS | contract | 빈 payload 응답 계약 보존 |
| AC-SEC4-006 | REQ-SEC4-006 | regression | no-regression | F2 정상 스캔 + lexical 가드 보존 |
| AC-SEC4-007 | REQ-SEC4-007 | MUST-PASS | boundary | internal/hook specid import 없음 |
| AC-SEC4-008 | REQ-SEC4-008 | MUST-PASS | no-new-surface | 새 패키지/플래그/추상화 없음 |
| AC-SEC4-009 | NFR-SEC4-005 | MUST-PASS | cross-platform | windows build 통과 |
| AC-SEC4-010 | NFR-SEC4-003 | regression | coverage | 커버리지 회귀 없음 |

## D.1 AC 상세 (Given-When-Then)

### AC-SEC4-001 — [MUST-PASS, reproduction F1] symlinked-parent-dir 쓰기 거부

- **Given** configDir 안에 사전 존재하는 symlinked 디렉터리(`configDir/linkdir → <outside-temp-dir>`)가 있고, 백업이 `linkdir/evil.yaml` relPath를 산출함
- **When** 복원이 `restoreTargetContained(configDir, <configDir>/linkdir/evil.yaml)`를 호출함
- **Then** `restoreTargetContained`가 `false`를 반환하고, `os.WriteFile`가 호출되지 않아 `<outside-temp-dir>/evil.yaml`이 생성되지 않음 (수정 전: 쓰기가 outside에 안착 = 취약점 재현)
- **검증**:
  ```bash
  go test -run 'TestRestoreMoaiConfig_RejectsSymlinkedParentDir$' ./internal/cli/ -v 2>&1 | grep -E '^(=== RUN|--- PASS|--- FAIL|PASS|FAIL|ok)'
  # 기대: === RUN + --- PASS (non-vacuous), outside 파일 부재 assert 포함
  ```
- **non-vacuous 확인**: `go test -run 'TestRestoreMoaiConfig_RejectsSymlinkedParentDir$' ./internal/cli/ -v` 출력에 `=== RUN   TestRestoreMoaiConfig_RejectsSymlinkedParentDir`가 존재(`[no tests to run]` 아님).

### AC-SEC4-002 — [MUST-PASS, parity] 공유 헬퍼 수정으로 양 walk 동시 봉쇄

- **Given** `restoreTargetContained`가 레거시 walk(`restoreMoaiConfigLegacy`)와 모던 walk(`restoreMoaiConfig`)의 공유 헬퍼임
- **When** parent-chain 봉쇄를 `restoreTargetContained` 1곳에 추가함
- **Then** 모던 walk와 레거시 walk **둘 다** symlinked-parent-dir 쓰기를 거부함 (AC-SEC4-001 재현 테스트가 두 경로를 모두 커버; 또는 모던/레거시 각각 sub-test로 PASS)
- **검증**:
  ```bash
  # 두 walk 모두 restoreTargetContained 호출 확인 (구조 불변 — 분기 추가 없음)
  grep -nE 'restoreTargetContained\(configDir, targetPath\)' internal/cli/update.go
  # 기대: 2 매치 (L1993 모던 + L2085 레거시) — 공유 헬퍼 호출 지점 보존
  go test -run 'TestRestoreMoaiConfig_RejectsSymlinkedParentDir$|TestRestoreMoaiConfigLegacy' ./internal/cli/ 2>&1 | grep -E '^(ok|FAIL)'
  ```

### AC-SEC4-003 — [regression] F1 정상 복원 + leaf 가드 보존

- **Given** 백업 엔트리가 정규 파일이고 leaf·parent chain 모두 configDir 내부임
- **When** 복원이 `restoreTargetContained`를 호출함
- **Then** `true`를 반환하고 기존 머지·복원 동작이 변경 없이 수행됨; 또한 SEC-HARDEN-003 leaf 가드(leaf `isSymlinkEntry` + `filepath.Rel` `..` 거부)가 여전히 작동함
- **검증**:
  ```bash
  go test -run 'TestRestoreMoaiConfig_LegacyBackup$|TestRestoreMoaiConfigLegacy_WithMerge$|TestRestoreMoaiConfig_SkipsNonYAML$|TestRestoreMoaiConfig_3WayMerge$' ./internal/cli/ 2>&1 | grep -E '^(ok|FAIL)'
  # 기대: ok (전부 GREEN — leaf 가드 + 복원 동작 회귀 없음)
  ```

### AC-SEC4-004 — [MUST-PASS, reproduction F2] symlink-in-root scan-target 읽기 거부

- **Given** project root 안에 root 밖 secret을 가리키는 symlink(`<root>/innocent.go → <outside-temp>/secret.go`, secret.go는 `// @MX:NOTE: secret-tag` 포함)가 있음
- **When** `runMXScan`이 `input.FilePath = <root>/innocent.go`로 호출됨
- **Then** `filepath.EvalSymlinks` 해소 결과가 root를 탈출하므로 `scanner.ScanFile` 없이 early return; secret의 MX-tag가 사이드카에 기록되지 않음 (수정 전: ScanFile이 secret MX-tag 읽어 기록 = 취약점 재현)
- **검증**:
  ```bash
  go test -run 'TestRunMXScan_RejectsSymlinkInRootEscapingTarget$' ./internal/hook/ -v 2>&1 | grep -E '^(=== RUN|--- PASS|--- FAIL|PASS|FAIL|ok)'
  # 기대: === RUN + --- PASS (non-vacuous), 사이드카에 secret MX-tag 부재 assert 포함
  ```
- **non-vacuous 확인**: `-v` 출력에 `=== RUN   TestRunMXScan_RejectsSymlinkInRootEscapingTarget` 존재.

### AC-SEC4-005 — [MUST-PASS, contract] 빈 payload 응답 계약 보존

- **Given** F2 봉쇄가 트리거되거나 트리거되지 않음
- **When** `file_changed` main handler `Handle`가 호출됨
- **Then** 봉쇄 위반 여부와 무관하게 고정 빈 payload(`&HookOutput{}`)를 반환하고, async side-effect 실패가 hook 응답에 전파되지 않음
- **검증**:
  ```bash
  go test -run 'TestFileChanged_AsyncReturn_Under100ms$|TestFileChanged_SideEffectsCompleted$' ./internal/hook/ 2>&1 | grep -E '^(ok|FAIL)'
  # 기대: ok (빈 payload + async 계약 회귀 없음)
  ```

### AC-SEC4-006 — [regression] F2 정상 스캔 + lexical 가드 보존

- **Given** `input.FilePath`가 symlink가 아니거나 해소 후에도 root 내부임
- **When** `runMXScan`이 호출됨
- **Then** 기존 lexical `pathContainedIn` 가드를 통과한 뒤 스캔·사이드카 동작이 변경 없이 수행됨; SEC-HARDEN-003 leaf 가드(lexical FilePath 거부 + sidecar re-root)가 여전히 작동함
- **검증**:
  ```bash
  go test -run 'TestRunMXScan_RejectsUncontainedFilePath$|TestRunMXScan_RejectsUncontainedSidecarCWD$|TestRunMXScan_AllowsInProjectPath$' ./internal/hook/ 2>&1 | grep -E '^(ok|FAIL)'
  # 기대: ok (전부 GREEN — SEC-HARDEN-003 lexical/re-root 가드 회귀 없음)
  ```

### AC-SEC4-007 — [MUST-PASS, boundary] internal/hook specid import 없음

- **Given** F2 봉쇄가 `internal/hook`에 구현됨
- **When** import 그래프를 검사함
- **Then** `internal/hook`가 `internal/cli/specid`를 import하지 않음 (SEC-HARDEN-003 책임 분리 보존)
- **검증**:
  ```bash
  grep -rnE 'internal/cli/specid' internal/hook/
  # 기대: no output (0 매치)
  ```

### AC-SEC4-008 — [MUST-PASS, no-new-surface] 새 패키지/플래그/추상화 없음

- **Given** 본 SPEC이 외과적 봉쇄만 수행함
- **When** 변경 diff를 검사함
- **Then** 새 Go 패키지·새 cobra 플래그·새 공개 추상화가 도입되지 않음; 변경은 `restoreTargetContained`(update.go) + `runMXScan`(file_changed.go) 함수 본문 + 테스트 파일에 국한됨
- **검증**:
  ```bash
  # 변경된 비테스트 .go 파일이 정확히 2개인지 확인 (update.go, file_changed.go)
  # baseline은 pre-flight C.1에서 캡처한 pre-M1 SHA 사용 (origin/main은 parallel
  # session race 시 이동 가능 → SEC-HARDEN-003 self-reconcile 이력; plan-audit D1)
  BASELINE_SHA="<pre-flight C.1의 git rev-parse HEAD로 캡처한 pre-M1 SHA>"
  git diff --name-only "$BASELINE_SHA" -- 'internal/**/*.go' | grep -vE '_test\.go$' | sort
  # 기대: internal/cli/update.go + internal/hook/file_changed.go (정확히 2)
  # 새 .go 파일 부재 확인
  git diff --name-only --diff-filter=A "$BASELINE_SHA" -- 'internal/**/*.go' | grep -vE '_test\.go$'
  # 기대: no output (신규 비테스트 소스 파일 없음)
  ```

### AC-SEC4-009 — [MUST-PASS, cross-platform] windows build 통과

- **Given** 봉쇄가 `filepath.EvalSymlinks`/`filepath.Rel`/`os.Lstat` 기반임
- **When** windows 타깃으로 빌드함
- **Then** 빌드가 exit 0로 성공함
- **검증**:
  ```bash
  GOOS=windows GOARCH=amd64 go build ./internal/cli/... ./internal/hook/...
  # 기대: exit 0
  ```

### AC-SEC4-010 — [regression, coverage] 커버리지 회귀 없음

- **Given** 변경 패키지가 `internal/cli`, `internal/hook`임
- **When** 커버리지를 측정함
- **Then** 변경 전 baseline 대비 회귀하지 않음(language go.md ≥85% 목표; F1/F2 reproduction 테스트가 새 가드 분기를 커버)
- **검증**:
  ```bash
  go test -cover ./internal/cli/... ./internal/hook/... 2>&1 | grep -E 'coverage:|ok'
  # 기대: 각 패키지 coverage가 pre-flight baseline 이상
  ```

## D.2 Severity / Closure Gates

- **MUST-PASS (closure gate)**: AC-SEC4-001, 002, 004, 005, 007, 008, 009. 하나라도 FAIL이면 4-phase close 불가.
- **Regression (must hold)**: AC-SEC4-003, 006, 010. SEC-HARDEN-003 leaf 가드 + 정상 동작 보존.
- **Definition of Done**: 10/10 AC PASS + cross-platform build exit 0 + C-HRA-008 grep 0 + lint NEW 0 + 두 reproduction 테스트가 수정 전 FAIL(취약점 재현) → 수정 후 PASS(봉쇄 확인).

## D.3 Traceability (REQ → AC)

| REQ | AC |
|-----|-----|
| REQ-SEC4-001 | AC-SEC4-001 |
| REQ-SEC4-002 | AC-SEC4-002 |
| REQ-SEC4-003 | AC-SEC4-003 |
| REQ-SEC4-004 | AC-SEC4-004 |
| REQ-SEC4-005 | AC-SEC4-005 |
| REQ-SEC4-006 | AC-SEC4-006 |
| REQ-SEC4-007 | AC-SEC4-007 |
| REQ-SEC4-008 | AC-SEC4-008 |
| NFR-SEC4-003 | AC-SEC4-010 |
| NFR-SEC4-005 | AC-SEC4-009 |

## D.4 Forward-looking checks (다음 sync-auditor용)

- 본 SPEC 봉쇄 후, F1(symlinked-parent-dir write) + F2(symlink-in-root read) 두 클래스는 spec.md §B에서 **CLOSED**로 선언됨. 다음 sync-auditor가 adversarial bypass로 재시도 시: F1은 EvalSymlinks parent-chain으로, F2는 EvalSymlinks scan-target 재검사로 봉쇄되어야 한다.
- 여전히 OPEN(명시적 이연): TOCTOU window, §F.1 `${IFS}`, §F.2 env-trust — spec.md §F 참조.
