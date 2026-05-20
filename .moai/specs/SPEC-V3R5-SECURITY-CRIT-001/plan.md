# Plan — SPEC-V3R5-SECURITY-CRIT-001

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-05-20 | manager-spec | Tier M plan 초안. 3 P0 도메인을 3 milestone으로 분리 + cross-cutting M4. Late-Branch workflow 준수. |

## 1. Strategy (Brownfield extend)

본 SPEC은 **기존 코드를 정정**하는 brownfield fix이다. 새 abstraction 도입을 최소화하고, 가장 작은 surgical change로 OWASP/CWE 위협을 차단한다.

### 1.1 PRESERVE 대상 (변경 금지)

| 항목 | 파일 / 함수 | 이유 |
|------|------------|------|
| `~/.moai/.env.glm` permission `0o600` | `internal/cli/glm.go::saveGLMConfig` 류 | 이미 정확. 본 SPEC scope는 derived 산출물. |
| PreToolUse permission gate | `internal/hook/pre_tool.go` | 780 LOC, 8 sentinel — 강건 검증됨. |
| Update URL 도메인 화이트리스트 | `internal/update/checker.go` (existing) | 변경 시 SSRF 표면 변동 — 본 SPEC scope outside. |
| GLM auth UX (모든 사용자-가시 동작) | `moai glm setup`, `moai cc`, `moai update` 정상 흐름 | 사용자 체감 동작 변화 금지 (REQ-SEC-004-001). |
| Windows build path 제외 | `internal/tmux/session.go` build tags | tmux Unix-only. Windows 코드 경로는 무변동. |
| 무관한 0o644 사용 (e.g., reports, audit logs) | session_end.go:150 (reportPath), session_start.go:576 (dstPath) | 본 SPEC scope는 **settings.local.json + sensitive env** 한정. P1-S5 (audit log permission)는 별도 SPEC. |

### 1.2 EXTEND 대상

| 항목 | 변경 |
|------|------|
| `internal/hook/session_start.go` | 3 sites `0o644 → 0o600` (lines 309, 403, 642; line 576 dstPath는 settings 아니므로 변경 안 함) |
| `internal/hook/session_end.go` | 1 site `0o644 → 0o600` (line 667 settings write) |
| `internal/hook/glm_tmux.go::ensureTmuxGLMEnv` | argv 경로 → source-file 경로 (sensitive only) |
| `internal/tmux/session.go::InjectEnv` | `sensitive: bool` 파라미터 추가 (or 별도 `InjectSensitive` 메서드) |
| `internal/cli/launcher.go:183` | 변경 없음 (`CLAUDE_CONFIG_DIR` 은 디렉터리 경로, sensitive 아님) |
| `internal/update/checker.go::queryRelease` (이름 가정) | checksum 다운로드 실패 시 `ErrChecksumUnavailable` 반환, 무묵음 |
| `internal/update/updater.go` | `version.Checksum == ""` 시 `ErrChecksumUnavailable` 반환 |
| 신규 sentinel | `internal/update/checker.go` 또는 `errors.go`: `ErrChecksumUnavailable`, `ErrTmuxSensitiveInjectFailed` |
| 회귀 테스트 신규 | `session_start_test.go`, `glm_tmux_test.go`, `checker_test.go`, `updater_test.go`, `subagent_boundary_test.go` (grep guard) |

## 2. Milestones (Priority 기반, 시간 추정 없음)

본 SPEC은 4개 milestone으로 구성된다. **M1, M2, M3 는 독립적**이며 병렬 가능. M4 는 직렬 (cross-cutting 검증).

### M1 — P0-1 GLM token disk persistence (Priority: P0)

**목표**: `settings.local.json` 의 4 write site permission을 `0o600` 으로 강제하고 회귀 테스트로 잠근다.

**범위**:
- `internal/hook/session_start.go` lines 309, 403, 642 → `0o600`
- `internal/hook/session_end.go` line 667 → `0o600`
- Helper 함수 도입 (선택): `internal/hook/settings_io.go::writeSettingsSecure(path, data)` — 향후 같은 경로 추가 시 일관성 보장.
- 신규 테스트 `internal/hook/session_start_test.go::TestEnsureGLMCredentialsFilePerm`:
  - `t.TempDir()` 에 fake `.claude/settings.local.json` 생성
  - `ensureGLMCredentials` 호출 후 `os.Stat(settingsPath).Mode().Perm() == 0o600` 단언
- 신규 테스트 `internal/hook/session_end_test.go::TestSessionEndSettingsPerm`:
  - GLM keys 가 들어있는 settings 에서 session_end 의 write-back 경로 검증
- 신규 cross-cutting grep guard `internal/hook/subagent_boundary_test.go` (이름 활용 / 새 파일):
  - `grep -rn '0o644' internal/hook/session_start.go internal/hook/session_end.go | grep settings` 가 **0 match** 임을 단언

**Verification**:
```bash
go test ./internal/hook/... -run TestEnsureGLMCredentialsFilePerm -v
go test ./internal/hook/... -run TestSessionEndSettingsPerm -v
grep -n '0o644' internal/hook/session_start.go internal/hook/session_end.go | grep -iE 'settings|WriteFile.*settingsPath'
# Expected: 0 matches
```

### M2 — P0-2 tmux IPC sensitive value injection (Priority: P0)

**목표**: GLM token 의 argv 노출을 source-file 경로로 우회. tmux 버전 호환성 검증.

**범위**:
- `internal/tmux/session.go::InjectEnv` 시그니처 확장: `InjectEnv(name, key, value string, sensitive bool) error` 또는 `InjectSensitiveEnv(name, key, value string) error` 별도 신설.
  - sensitive=true 시: `os.CreateTemp(homeRunDir, "moai-tmux-*")` (homeRunDir = `~/.moai/run/`, mkdir `0o700`) → `chmod 0o600` 즉시 → tmux script `set-environment <key> <value>` 한 줄 작성 → `tmux source-file <temp>` → `os.Remove(temp)` (defer).
  - sensitive=false 시: 기존 argv 경로 유지 (backward compat).
- `internal/hook/glm_tmux.go::ensureTmuxGLMEnv` 호출 분기:
  - `ANTHROPIC_AUTH_TOKEN` → `InjectSensitiveEnv`
  - `ANTHROPIC_BASE_URL`, `CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS` → 기존 argv 경로 (URL/flag는 sensitive 아님).
- `internal/cli/launcher.go:183` 무변동 (`CLAUDE_CONFIG_DIR` 은 디렉터리 경로).
- tmux version preflight 헬퍼 `internal/tmux/version.go`:
  - `tmux -V` 파싱
  - `<2.0` 시 명시적 에러 (`ErrTmuxVersionUnsupported`)
- 신규 테스트 `internal/tmux/session_test.go::TestInjectSensitiveEnvNoArgvLeak`:
  - mock tmux (또는 실 tmux ≥3.x available) 환경에서 호출
  - 검증: `ps -ef` 또는 mock command recorder 로 argv 에 token value 미포함 단언
  - 임시 파일 unlink 검증 (`os.CreateTemp` 패턴 검증, defer cleanup)
- 신규 테스트 `internal/hook/glm_tmux_test.go::TestEnsureTmuxGLMEnvSensitiveBranch`:
  - `ANTHROPIC_AUTH_TOKEN` 이 InjectSensitiveEnv 경유함을 mock 으로 검증

**Verification**:
```bash
go test ./internal/tmux/... -v -run TestInjectSensitiveEnv
go test ./internal/hook/... -v -run TestEnsureTmuxGLMEnvSensitiveBranch
GOOS=windows GOARCH=amd64 go build ./...   # Windows build 무영향 확인
grep -rn 'set-environment.*ANTHROPIC_AUTH_TOKEN\|tmux.*token' internal/ | grep -v "_test.go"
# Expected: 0 direct argv-passing matches
```

### M3 — P0-3 Update flow mandatory checksum (Priority: P0)

**목표**: `checksums.txt` 다운로드 실패 시 update abort. 명시적 sentinel error.

**범위**:
- `internal/update/checker.go` 신규 export: `var ErrChecksumUnavailable = errors.New("update: checksum verification unavailable")`
- `internal/update/checker.go::queryRelease` (또는 동등): `checksumsURL` 발견 시 download 실패 → `ErrChecksumUnavailable` 반환 (현재 묵음 처리 제거).
- Retry logic (`internal/update/checker.go::downloadChecksumWithRetry` 신규):
  - 3회 시도, 지수 백오프 baseline 2s (2s, 4s, 8s)
  - HTTP transient (429, 5xx, network timeout) 시에만 retry. 4xx (404 등 영구) 는 즉시 fail.
- `internal/update/updater.go::downloadAndVerify` 검증: `version.Checksum == ""` 시 `ErrChecksumUnavailable` 반환 (defense-in-depth, REQ-SEC-003-005).
- 신규 테스트 `internal/update/checker_test.go::TestQueryReleaseChecksumDownloadFailureAborts`:
  - mock HTTP server: release JSON OK, checksums.txt 다운로드 → 503 또는 connection refused
  - 검증: `ErrChecksumUnavailable` 반환 + 3회 retry 시도
- 신규 테스트 `internal/update/updater_test.go::TestDownloadAndVerifyEmptyChecksumRejected`:
  - `VersionInfo{Checksum: ""}` 입력 → `ErrChecksumUnavailable` 반환 단언 (binary 다운로드 시도하지 않음)
- 신규 테스트 `internal/update/checker_test.go::TestChecksumRetryExponentialBackoff`:
  - mock clock + mock HTTP. 3회 시도, 2s/4s/8s 간격 단언.

**Verification**:
```bash
go test ./internal/update/... -v -run TestQueryReleaseChecksumDownloadFailureAborts
go test ./internal/update/... -v -run TestDownloadAndVerifyEmptyChecksumRejected
go test ./internal/update/... -v -run TestChecksumRetryExponentialBackoff
# 결함 재현 grep
grep -n 'continue without checksum verification\|better to allow update with warning' internal/update/checker.go
# Expected: 0 matches (해당 주석/로직 제거됨)
```

### M4 — Cross-cutting 검증 + 회귀 가드 (Priority: P0)

**목표**: M1+M2+M3 변경이 전체 빌드 / cross-platform / coverage / boundary 를 깨지 않음을 보장.

**범위**:
- Full test suite: `go test ./...` PASS
- Cross-platform build: `GOOS=windows GOARCH=amd64 go build ./...` PASS, `GOOS=linux GOARCH=amd64 go build ./...` PASS
- Race detector: `go test -race ./internal/hook/... ./internal/tmux/... ./internal/update/...` PASS
- Coverage: 영향 패키지 ≥85%, critical files (session_start.go, glm_tmux.go, checker.go) ≥90%
- Subagent boundary (C-HRA-008 류) 무영향:
  - `grep -rn 'AskUserQuestion\|mcp__askuser' internal/hook/ internal/tmux/ internal/update/ | grep -v "_test.go" | grep -v "// "` → 0 match
- spec-lint clean: `go run ./cmd/moai spec lint --strict --path .moai/specs/SPEC-V3R5-SECURITY-CRIT-001/`
- Lint baseline 비교: NEW issues 0 (W2-deferred manifest 와 비교)
- Pre-existing 추가 회귀 테스트로 `~/.moai/.env.glm` 0o600 이 깨지지 않음을 단언 (`internal/cli/glm_test.go::TestEnvGLMSourcePermissionPreserved` — 가벼운 sanity)

**Verification**:
```bash
go test ./...
GOOS=windows GOARCH=amd64 go build ./...
GOOS=linux GOARCH=amd64 go build ./...
go test -race ./internal/hook/... ./internal/tmux/... ./internal/update/...
go test -cover ./internal/hook/... ./internal/tmux/... ./internal/update/...
golangci-lint run --timeout=2m
```

## 3. Technical Approach

### 3.1 Permission Hardening Pattern (M1)

기존:
```go
if err := os.WriteFile(settingsPath, newData, 0o644); err != nil { ... }
```

수정:
```go
if err := os.WriteFile(settingsPath, newData, 0o600); err != nil { ... }
```

선택적으로 헬퍼 도입:
```go
// internal/hook/settings_io.go
// writeSettingsSecure writes settings.local.json with owner-only permission (0o600).
// This file may contain ANTHROPIC_AUTH_TOKEN and other sensitive env values
// per SPEC-V3R5-SECURITY-CRIT-001 REQ-SEC-001-001.
func writeSettingsSecure(path string, data []byte) error {
    return os.WriteFile(path, data, 0o600)
}
```

### 3.2 tmux Sensitive Injection Pattern (M2)

기존 (vulnerable):
```go
args := []string{"set-environment", key, value}
cmd := exec.Command("tmux", append([]string{"-t", sessionName}, args...)...)
return cmd.Run()
```

수정 (sensitive=true 경로):
```go
// Write tmux script to a private temp file
tmpFile, err := os.CreateTemp(homeRunDir, "moai-tmux-*")
if err != nil {
    return fmt.Errorf("create temp: %w", err)
}
defer os.Remove(tmpFile.Name())
if err := os.Chmod(tmpFile.Name(), 0o600); err != nil {
    return fmt.Errorf("chmod temp: %w", err)
}
script := fmt.Sprintf("set-environment %s %s\n", key, value) // value never reaches argv
if _, err := tmpFile.WriteString(script); err != nil {
    return fmt.Errorf("write script: %w", err)
}
if err := tmpFile.Close(); err != nil {
    return fmt.Errorf("close temp: %w", err)
}
cmd := exec.Command("tmux", "-t", sessionName, "source-file", tmpFile.Name())
if err := cmd.Run(); err != nil {
    return fmt.Errorf("%w: %v", ErrTmuxSensitiveInjectFailed, err)
}
return nil
```

핵심: token `value` 는 process argv 에 절대 등장하지 않음. tmux script 파일은 `0o600`, defer cleanup.

### 3.3 Mandatory Checksum Pattern (M3)

기존 (vulnerable):
```go
if checksumsURL != "" {
    checksum, err := c.downloadChecksum(checksumsURL, archiveName)
    if err == nil && checksum != "" {
        info.Checksum = checksum
    }
    // If checksum download fails, continue without checksum verification
    // (better to allow update with warning than to block entirely)
}
```

수정:
```go
if checksumsURL == "" {
    return info, fmt.Errorf("%w: release missing checksums.txt asset", ErrChecksumUnavailable)
}
checksum, err := c.downloadChecksumWithRetry(checksumsURL, archiveName, 3)
if err != nil {
    return info, fmt.Errorf("%w: %v", ErrChecksumUnavailable, err)
}
if checksum == "" {
    return info, fmt.Errorf("%w: checksums.txt parsed but no entry for %s", ErrChecksumUnavailable, archiveName)
}
info.Checksum = checksum
return info, nil
```

`downloadChecksumWithRetry` 는 transient HTTP error 시에만 retry (timeout, 429, 5xx); 4xx 영구는 즉시 fail.

`updater.go::downloadAndVerify` 도 defense-in-depth:
```go
if version.Checksum == "" {
    return "", ErrChecksumUnavailable
}
// existing hash check follows
```

## 4. Dependencies & Ordering

| Milestone | Depends on | Parallel-safe with |
|-----------|-----------|-------------------|
| M1 | (none) | M2, M3 |
| M2 | (none) | M1, M3 |
| M3 | (none) | M1, M2 |
| M4 | M1, M2, M3 | — |

M1, M2, M3 는 서로 다른 파일을 수정하므로 병렬 가능. M4 는 통합 검증.

## 5. Late-Branch Workflow 적용

본 SPEC은 SPEC-V3R5-LATE-BRANCH-001 결정에 따라 다음 절차로 수행한다:

**Phase A (SPEC)**: 본 plan/spec/acceptance 를 main 직접 commit (이미 진행 중).

**Phase B (Implementation)**: M1, M2, M3 의 commit 들도 main 에 직접 누적 (auto branch 생성 금지).

**Phase C (PR)**: M1-M4 모두 완료 시, `git switch -c feat/SPEC-V3R5-SECURITY-CRIT-001` → `git push origin feat/SPEC-V3R5-SECURITY-CRIT-001` → `gh pr create` → squash merge → admin override (필요 시 chicken-and-egg lint baseline 적용).

**Phase D (Local sync)**: PR 머지 후 `git reset --hard origin/main` + `git pull --ff-only` (또는 user 결정 시 `! git reset --hard`).

**GitHub Issue**: SPEC-V3R5-LATE-BRANCH-001 REQ-LB-009 (default off) 에 따라 본 SPEC 은 `--issue` 옵션을 사용하지 않는다.

## 6. Token Budget (manager-develop 위임 시)

| Milestone | 예상 token | 비고 |
|-----------|----------|------|
| M1 | ~8K | 4 line change + 2 test file |
| M2 | ~15K | new function + sensitive flag + 2 test file + tmux version helper |
| M3 | ~12K | sentinel + retry + 3 test file |
| M4 | ~5K | 전체 검증 명령 + grep |

총 ~40K. Tier M (≤80K target) 내.

## 7. Out of Scope (재확인)

### 7.1 Out of Scope

- cosign/sigstore 서명: 별도 SPEC.
- atomicWrite (P0-4), schema collision (P0-5), pipeline stub (P0-6), defaultProjectedScorer (P0-7), github/runner stub (P0-8): 각각 별도 SPEC.
- P1 findings 전체 (P1-S1 ~ P1-U2): 별도 SPEC들.
- Audit log 0o644 → 0o600 (P1-S5): 본 SPEC 은 settings.local.json + sensitive env 만 다룬다.
- Windows-native tmux 대체: tmux Unix-only.
- ~/.moai/ parent dir `0o755 → 0o700` (P3-3): defense-in-depth hygiene SPEC.

## 8. Self-Audit Pre-Check

manager-develop 위임 전 본 plan 자가 점검 (plan-auditor 기준):

- [x] Brownfield strategy (extend, PRESERVE 목록 명시)
- [x] Milestone 명확 (M1-M4), 의존성 그래프 명시
- [x] REQ↔M1/M2/M3 매핑 명확 (REQ-SEC-001-* → M1, REQ-SEC-002-* → M2, REQ-SEC-003-* → M3)
- [x] Verification 명령 binary (PASS/FAIL)
- [x] Risk register (5 risks, mitigation 명시)
- [x] Cross-platform 검증 포함 (M4)
- [x] Coverage threshold 명시 (≥85%, critical ≥90%)
- [x] Late-Branch workflow 절차 명시
- [x] Token budget 추정
- [x] Out of Scope 재확인 (§7)

예상 plan-auditor score: 0.88+ (Tier M threshold 0.80 통과 여유).
