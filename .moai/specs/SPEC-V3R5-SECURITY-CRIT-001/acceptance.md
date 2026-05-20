# Acceptance Criteria — SPEC-V3R5-SECURITY-CRIT-001

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-05-20 | manager-spec | Tier M acceptance 초안. 11개 binary AC (P0-1: 4, P0-2: 4, P0-3: 3) + cross-cutting Definition of Done. |

## REQ ↔ AC Traceability Matrix

| REQ ID | AC IDs | Verification 방식 |
|--------|--------|------------------|
| REQ-SEC-001-001 | AC-SEC-001, AC-SEC-002 | `os.Stat` + grep |
| REQ-SEC-001-002 | AC-SEC-001 | Go test `TestEnsureGLMCredentialsFilePerm` |
| REQ-SEC-001-003 | AC-SEC-001 | (M1 helper coverage 일부) |
| REQ-SEC-001-004 | AC-SEC-002 | Go test `TestSessionEndSettingsPerm` |
| REQ-SEC-001-005 | AC-SEC-003 | Grep guard test (regression lock) |
| REQ-SEC-001-006 | AC-SEC-004 | `~/.moai/.env.glm` source perm preservation test |
| REQ-SEC-002-001 | AC-SEC-005, AC-SEC-006 | Mock tmux argv recorder |
| REQ-SEC-002-002 | AC-SEC-005 | Go test `TestInjectSensitiveEnvNoArgvLeak` |
| REQ-SEC-002-003 | AC-SEC-006 | Code review + grep guard |
| REQ-SEC-002-004 | AC-SEC-005 | InjectEnv signature test |
| REQ-SEC-002-005 | AC-SEC-007 | Error path test |
| REQ-SEC-002-006 | (covered indirectly) | Stale cleanup helper (best-effort) |
| REQ-SEC-002-007 | AC-SEC-008 | `GOOS=windows GOARCH=amd64 go build ./...` |
| REQ-SEC-003-001 | AC-SEC-009, AC-SEC-010 | Mock HTTP test |
| REQ-SEC-003-002 | AC-SEC-009 | Mock release JSON + checksums.txt |
| REQ-SEC-003-003 | AC-SEC-009 | `ErrChecksumUnavailable` sentinel return |
| REQ-SEC-003-004 | AC-SEC-010 | Retry count + backoff assertion |
| REQ-SEC-003-005 | AC-SEC-011 | Defense-in-depth empty-checksum reject |
| REQ-SEC-003-006 | (PRESERVE) | URL whitelist 무변동 (negative grep) |
| REQ-SEC-003-007 | (Optional) | Sentinel export-grade type 유지 |
| REQ-SEC-004-001 | (Manual) | UX 무변동 manual smoke (release note) |
| REQ-SEC-004-002 | AC-SEC-001 ~ AC-SEC-011 | 본 모든 AC가 회귀 테스트 |
| REQ-SEC-004-003 | AC-SEC-DoD | CI 통과 게이트 |
| REQ-SEC-004-004 | AC-SEC-DoD | Coverage threshold |

100% REQ → 최소 1 AC 커버리지. 모든 AC binary PASS/FAIL.

---

## P0-1 Domain (Hook GLM token disk persistence)

### AC-SEC-001 — GLM credential write 시 settings.local.json mode = 0o600

**Given**: 빈 `t.TempDir()` 에 fake `~/.moai/.env.glm` (mode `0o600`) + `.claude/` 디렉터리만 존재.
**When**: `ensureGLMCredentials(projectDir)` (`internal/hook/session_start.go`) 가 호출되어 새로운 `settings.local.json` 을 생성한다.
**Then**: `os.Stat(filepath.Join(projectDir, ".claude", "settings.local.json")).Mode().Perm()` 이 `0o600` 과 정확히 일치한다.

**Verification command**:
```bash
go test -v -run TestEnsureGLMCredentialsFilePerm ./internal/hook/...
```
**Expected output**: `--- PASS: TestEnsureGLMCredentialsFilePerm`

**Binary criterion**: PASS / FAIL.

---

### AC-SEC-002 — session_end settings write-back mode = 0o600

**Given**: GLM keys 가 포함된 `.claude/settings.local.json` (mode `0o600`) + tmux session 종료 시나리오.
**When**: `internal/hook/session_end.go` 의 settings write-back 경로 (line 667 영역) 가 GLM keys 를 제거한 새 페이로드를 쓴다.
**Then**: 쓰기 후 `settings.local.json` 의 mode 가 `0o600` 으로 유지된다 (downgrade 되어선 안 됨).

**Verification command**:
```bash
go test -v -run TestSessionEndSettingsPerm ./internal/hook/...
```
**Expected output**: `--- PASS: TestSessionEndSettingsPerm`

**Binary criterion**: PASS / FAIL.

---

### AC-SEC-003 — settings.local.json write 시 0o644 정적 발생 금지 (grep guard)

**Given**: 본 SPEC 적용된 코드베이스.
**When**: 다음 grep 명령을 실행한다.
**Then**: 출력이 **0 lines** 이다.

**Verification command**:
```bash
grep -nE 'os\.WriteFile\([^,]*settingsPath[^,]*,\s*[^,]+,\s*0o644' internal/hook/session_start.go internal/hook/session_end.go
```
**Expected output**: (no output, exit 1 from grep is fine — assertion is "no match")

또는 회귀 가드 테스트 형태:
```bash
go test -v -run TestNoSettingsLocalJSONWith0o644 ./internal/hook/...
```

**Binary criterion**: PASS (0 matches) / FAIL (≥1 match).

---

### AC-SEC-004 — `~/.moai/.env.glm` source permission PRESERVE

**Given**: `moai glm setup` 으로 생성된 `~/.moai/.env.glm` (mode `0o600`).
**When**: 본 SPEC 의 모든 변경 적용 후 `internal/cli/glm.go` 동등 경로의 source 파일 쓰기 동작.
**Then**: `~/.moai/.env.glm` 의 mode 가 `0o600` 으로 유지된다 (regression 없음).

**Verification command**:
```bash
go test -v -run TestEnvGLMSourcePermissionPreserved ./internal/cli/...
```
**Expected output**: `--- PASS: TestEnvGLMSourcePermissionPreserved`

**Binary criterion**: PASS / FAIL.

---

## P0-2 Domain (tmux IPC sensitive value injection)

### AC-SEC-005 — Sensitive env (ANTHROPIC_AUTH_TOKEN) tmux 주입 시 argv에 미노출

**Given**: mock tmux command recorder (또는 실 tmux ≥3.x test fixture). `ANTHROPIC_AUTH_TOKEN=sk-test-xxx` 입력.
**When**: `ensureTmuxGLMEnv(...)` (or 동등) 가 `InjectSensitiveEnv("ANTHROPIC_AUTH_TOKEN", "sk-test-xxx")` 경로를 호출한다.
**Then**:
1. tmux subprocess 의 argv 어느 원소에도 `"sk-test-xxx"` 가 등장하지 않는다.
2. tmux argv 는 `["-t", sessionName, "source-file", "/path/to/temp"]` 형태이다.
3. 호출 후 임시 파일 (`/tmp/moai-tmux-*` 또는 `~/.moai/run/moai-tmux-*`) 이 unlink 되어 있다 (`os.Stat` → `ErrNotExist`).

**Verification command**:
```bash
go test -v -run TestInjectSensitiveEnvNoArgvLeak ./internal/tmux/...
```
**Expected output**: `--- PASS: TestInjectSensitiveEnvNoArgvLeak`

**Binary criterion**: PASS / FAIL.

---

### AC-SEC-006 — Non-sensitive env (CLAUDE_CONFIG_DIR) argv 경로 유지 (회귀 방지)

**Given**: `launcher.go:183` 의 `CLAUDE_CONFIG_DIR` 주입 경로.
**When**: 코드 검사 (또는 mock command recorder).
**Then**: `CLAUDE_CONFIG_DIR` 은 기존 argv 경로 `exec.Command("tmux", "set-environment", "CLAUDE_CONFIG_DIR", profileDir)` 로 유지된다 (디렉터리 경로는 sensitive 아님; over-engineering 회피).

**Verification command**:
```bash
grep -n 'CLAUDE_CONFIG_DIR' internal/cli/launcher.go | grep 'set-environment'
```
**Expected output**: 1 매치 — line 183 `exec.Command("tmux", "set-environment", "CLAUDE_CONFIG_DIR", profileDir)` 그대로.

**Binary criterion**: PASS (1 match, 변경 없음) / FAIL (변경됨).

---

### AC-SEC-007 — Sensitive injection 실패 시 argv fallback 금지

**Given**: mock filesystem 에서 `os.CreateTemp` 실패 (또는 `tmux source-file` 실패) 시나리오.
**When**: `InjectSensitiveEnv(...)` 가 임시 파일 생성/주입에 실패한다.
**Then**: 함수는 `ErrTmuxSensitiveInjectFailed` (또는 wrap된 sentinel) 를 반환하고, **fallback 으로 argv 경로를 시도하지 않는다** (token 누설 방지).

**Verification command**:
```bash
go test -v -run TestInjectSensitiveEnvFailureNoArgvFallback ./internal/tmux/...
```
**Expected output**: `--- PASS: TestInjectSensitiveEnvFailureNoArgvFallback`

**Binary criterion**: PASS / FAIL.

---

### AC-SEC-008 — Windows cross-platform build 무영향

**Given**: 본 SPEC의 모든 변경이 적용된 코드베이스.
**When**: Windows target 빌드 명령 실행.
**Then**: 빌드가 exit 0 으로 통과한다 (tmux 경로는 build tag 로 이미 제외).

**Verification command**:
```bash
GOOS=windows GOARCH=amd64 go build ./...
echo "exit=$?"
```
**Expected output**: `exit=0`

**Binary criterion**: PASS (exit 0) / FAIL (exit != 0).

---

## P0-3 Domain (Update flow mandatory checksum)

### AC-SEC-009 — checksums.txt 다운로드 실패 시 update abort + ErrChecksumUnavailable

**Given**: Mock HTTP server. Release JSON 응답 OK, `checksums.txt` URL 응답 503 (또는 connection refused, 3 회 retry 후 모두 실패).
**When**: `checker.QueryRelease(...)` (또는 동등) 호출.
**Then**: 함수는 `errors.Is(err, ErrChecksumUnavailable) == true` 인 에러를 반환한다. Binary URL 다운로드는 시도되지 않는다.

**Verification command**:
```bash
go test -v -run TestQueryReleaseChecksumDownloadFailureAborts ./internal/update/...
```
**Expected output**: `--- PASS: TestQueryReleaseChecksumDownloadFailureAborts`

**Binary criterion**: PASS / FAIL.

---

### AC-SEC-010 — checksum 다운로드 retry 3회 + 지수 백오프 (2s/4s/8s)

**Given**: Mock HTTP server + mock clock. checksums.txt 응답 = 503 (영구).
**When**: `downloadChecksumWithRetry(url, archiveName, 3)` 호출.
**Then**:
1. 정확히 3 회 HTTP 요청이 발생한다.
2. 시도 간 간격은 mock clock 기준 2s, 4s, 8s (지수 백오프) 이다.
3. 최종 반환 에러는 `ErrChecksumUnavailable` wrap.

**Verification command**:
```bash
go test -v -run TestChecksumRetryExponentialBackoff ./internal/update/...
```
**Expected output**: `--- PASS: TestChecksumRetryExponentialBackoff`

**Binary criterion**: PASS / FAIL.

---

### AC-SEC-011 — version.Checksum 이 empty 인 상태로 downloadAndVerify 호출 시 거부

**Given**: `VersionInfo{Checksum: ""}` (defense-in-depth 시나리오 — checker 우회 가정).
**When**: `updater.downloadAndVerify(version)` 호출.
**Then**: 함수는 `errors.Is(err, ErrChecksumUnavailable) == true` 를 반환한다. Binary HTTP 요청은 발생하지 않는다 (mock HTTP server request count = 0).

**Verification command**:
```bash
go test -v -run TestDownloadAndVerifyEmptyChecksumRejected ./internal/update/...
```
**Expected output**: `--- PASS: TestDownloadAndVerifyEmptyChecksumRejected`

**Binary criterion**: PASS / FAIL.

---

## Definition of Done (AC-SEC-DoD)

본 SPEC은 다음 게이트를 모두 통과해야 완료된다 (REQ-SEC-004-003, REQ-SEC-004-004 충족):

### DoD-1 — 모든 AC PASS
```bash
go test -v -run 'TestEnsureGLMCredentialsFilePerm|TestSessionEndSettingsPerm|TestNoSettingsLocalJSONWith0o644|TestEnvGLMSourcePermissionPreserved|TestInjectSensitiveEnvNoArgvLeak|TestInjectSensitiveEnvFailureNoArgvFallback|TestQueryReleaseChecksumDownloadFailureAborts|TestChecksumRetryExponentialBackoff|TestDownloadAndVerifyEmptyChecksumRejected' ./...
```
**Expected**: 모든 9 named test PASS (AC-SEC-006, AC-SEC-008 은 grep / build 검증).

### DoD-2 — 전체 테스트 스위트 PASS
```bash
go test ./...
```
**Expected**: exit 0, FAIL count = 0.

### DoD-3 — Race detector PASS
```bash
go test -race ./internal/hook/... ./internal/tmux/... ./internal/update/...
```
**Expected**: exit 0.

### DoD-4 — Cross-platform build PASS
```bash
GOOS=windows GOARCH=amd64 go build ./...
GOOS=linux GOARCH=amd64 go build ./...
GOOS=darwin GOARCH=arm64 go build ./...
```
**Expected**: 3 exit 0.

### DoD-5 — Coverage threshold
```bash
go test -cover ./internal/hook/... ./internal/tmux/... ./internal/update/...
```
**Expected**:
- `internal/hook` ≥ 85%
- `internal/tmux` ≥ 85%
- `internal/update` ≥ 85%
- `internal/hook/session_start.go` ≥ 90% (critical)
- `internal/hook/glm_tmux.go` ≥ 90% (critical)
- `internal/update/checker.go` ≥ 90% (critical)

### DoD-6 — Subagent boundary 무영향 (C-HRA-008 회귀 방지)
```bash
grep -rn 'AskUserQuestion\|mcp__askuser' internal/hook/ internal/tmux/ internal/update/ | grep -v "_test.go" | grep -v "// "
```
**Expected**: 0 matches.

### DoD-7 — spec-lint clean
```bash
go run ./cmd/moai spec lint --strict --path .moai/specs/SPEC-V3R5-SECURITY-CRIT-001/
```
**Expected**: `No findings` (또는 동등 PASS).

### DoD-8 — Lint baseline 무회귀
```bash
golangci-lint run --timeout=2m
```
**Expected**: NEW findings 0 (W2-deferred manifest 와 비교 시 증분 0).

### DoD-9 — Grep regression locks
```bash
# P0-1 lock
grep -nE 'os\.WriteFile\([^,]*settingsPath[^,]*,\s*[^,]+,\s*0o644' internal/hook/session_start.go internal/hook/session_end.go
# Expected: 0 matches

# P0-3 lock (묵음 우회 주석 제거됨)
grep -n 'continue without checksum verification\|better to allow update with warning' internal/update/checker.go
# Expected: 0 matches
```

### DoD-10 — Frontmatter status 갱신
- `spec.md` frontmatter `status: draft → implemented` (run-phase 완료 시)
- `spec.md` frontmatter `version: 0.1.0 → 0.2.0` (M1-M4 모두 완료)
- HISTORY 0.2.0 entry 추가

### DoD-11 — Late-Branch PR & Sync
- PR 생성: `feat/SPEC-V3R5-SECURITY-CRIT-001` branch
- PR 머지 (admin override 적용 가능)
- Local main sync (`git reset --hard origin/main` + `git pull --ff-only`)
- branch cleanup + `git fetch --prune`

### DoD-12 — Release blocker 게이트 해소 (v2.20.0-rc1)
본 SPEC 머지 후, `.moai/reports/review-v214-to-HEAD.md` 의 P0-1, P0-2, P0-3 가 release blocker 목록에서 제거된 상태가 됨을 확인.

---

## Edge Cases (테스트 시 고려)

| EC | 시나리오 | 처리 |
|----|---------|------|
| EC-1 | `~/.moai/.env.glm` 부재 | ensureGLMCredentials 가 no-op (기존 동작) — perm 변경 미적용 |
| EC-2 | 기존 `settings.local.json` mode = `0o600` (사용자가 수동 보호) | 변경 후 `0o600` 유지 (downgrade 금지) |
| EC-3 | tmux 미설치 / 미실행 | `ensureTmuxGLMEnv` 가 graceful skip (기존 동작) |
| EC-4 | tmux version < 2.0 (드물지만 존재) | `ErrTmuxVersionUnsupported` 명시적 에러 (argv fallback 금지) |
| EC-5 | `os.CreateTemp` 실패 (disk full) | `ErrTmuxSensitiveInjectFailed` 반환, token 누설 없음 |
| EC-6 | `checksums.txt` 응답이 200 OK 지만 archive name entry 부재 | `ErrChecksumUnavailable` (parsed but missing) |
| EC-7 | release JSON 에 `checksums.txt` asset 자체 부재 | `ErrChecksumUnavailable: release missing checksums.txt asset` |
| EC-8 | HTTP 4xx (404) — 영구 실패 | retry 없이 즉시 `ErrChecksumUnavailable` |
| EC-9 | HTTP 5xx 또는 timeout (transient) | 3회 retry, 모두 실패 시 `ErrChecksumUnavailable` |
| EC-10 | 본 SPEC 변경 후 `moai update` 정상 흐름 (checksum.txt 정상 다운로드) | 기존 동작 동일 — 사용자 체감 변화 없음 (REQ-SEC-004-001) |

각 EC 는 최소 1 test case 로 커버.

---

## Notes

- 본 acceptance criteria 는 모두 **binary PASS/FAIL** 이며 각 항목마다 단일 verification command 가 명시되었다.
- AC-SEC-001 ~ AC-SEC-011 은 자동화 테스트로 잠긴다 (REQ-SEC-004-002).
- DoD-1 ~ DoD-12 는 PR 머지 직전 CI 통과 게이트로 작동한다 (REQ-SEC-004-003).
- 본 SPEC 머지 = v2.20.0-rc1 release blocker 3건 (P0-1/P0-2/P0-3) 해소.
