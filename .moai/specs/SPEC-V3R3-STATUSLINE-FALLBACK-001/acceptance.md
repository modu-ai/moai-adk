---
id: SPEC-V3R3-STATUSLINE-FALLBACK-001
version: "0.2.0"
status: draft
created_at: 2026-05-10
updated_at: 2026-05-10
author: GOOS행님
priority: High
labels: [statusline, fallback, cli]
issue_number: null
---

# Acceptance Criteria: SPEC-V3R3-STATUSLINE-FALLBACK-001

## 1. Core Scenarios

### AC-SF-001: Empty stdin Fallback

**When** `parseStdin()`이 `nil`을 반환하면, `moai statusline` **shall** project directory name(`os.Getwd()`의 basename), git branch(filesystem 기반), MoAI-ADK version, rate limits 정보(API usage collector 경유)를 포함한 statusline을 출력한다.

**Verification**: `echo "" | moai statusline` — 출력에 project name과 version 포함 확인.

### AC-SF-002: `{}` Empty JSON Fallback

**When** stdin으로 `{}`이 전달되어 `parseStdin()`이 빈 `StdinData` 구조체를 반환하면, `moai statusline` **shall** 모든 필드 zero-value 상태에서도 fallback이 정상 동작하여 project directory name(`os.Getwd()` basename)과 model name(env/cache fallback 시도)을 표시한다.

**Verification**: `echo '{}' | moai statusline` — crash 없이 정상 3-line 형식 출력.

### AC-SF-003: Model Name Env Fallback

**When** stdin JSON에 `model` 필드가 없고 환경변수 `MOAI_LAST_MODEL`이 `"claude-opus-4-7-20250514"`로 설정되어 있으면, `moai statusline` **shall** statusline에 `"Opus 4.7"`을 표시한다.

**Verification**: `echo '{}' | MOAI_LAST_MODEL=claude-opus-4-7-20250514 moai statusline` — 출력에 "Opus 4.7" 포함.

### AC-SF-004: Model Name Cache Fallback

**When** stdin JSON에 `model` 필드가 없고 `MOAI_LAST_MODEL` env도 없으며 `~/.moai/state/last-model.txt` 파일에 `"claude-sonnet-4-20250514"`가 저장되어 있으면, `moai statusline` **shall** statusline에 `"Sonnet 4"`를 표시한다.

**Verification**: cache file 생성 후 `echo '{}' | moai statusline` — 출력에 "Sonnet 4" 포함.

### AC-SF-005: Model Name Cache Write

**When** stdin JSON에 유효한 model 정보(`{"model": {"display_name": "Opus"}}`)가 포함되어 `CollectMetrics()`가 model name을 추출하면, `moai statusline` **shall** `~/.moai/state/last-model.txt`에 `"Opus"`를 원자적으로 저장한다.

**Verification**: 정상 stdin 전달 후 cache file 내용 확인.

### AC-SF-006: Cwd Guard — Deleted Directory

**When** process의 현재 working directory가 삭제된 directory일 때, `moai statusline` **shall** working directory를 `$HOME`으로 전환하고 project directory name을 `~`로 표시하며 crash 없이 정상 statusline을 출력한다.

**Verification**: 테스트에서 임시 dir 생성 → `os.Chdir()` → dir 삭제 → `Build()` 호출 → 정상 출력 확인.

### AC-SF-007: Sandbox Cwd Priority

**When** stdin JSON의 `cwd`가 `/sandbox/project`이고 process의 `os.Getwd()`가 `/home/user/project`일 때, `extractProjectDirectory()` **shall** stdin의 `cwd` 값을 우선하여 `/sandbox/project`의 basename을 표시한다.

**Verification**: unit test에서 stdin `cwd`와 `os.Getwd()`를 다르게 설정 후 basename 확인.

### AC-SF-008: Happy Path Regression

**Event-driven**: **When** 정상 stdin JSON(모든 필드 포함)이 전달되면, `moai statusline` **shall** 기존과 동일한 3-line statusline을 출력한다. 출력 diff가 발생하지 않아야 한다.

**Verification**: 기존 golden test 통과.

## 2. Edge Case Scenarios

### EC-SF-001: `null` JSON Literal

**When** stdin으로 `null` literal이 전달되면, `moai statusline` **shall** `parseStdin()`이 `nil`을 반환하고 모든 fallback이 정상 동작한다.

**Verification**: `echo 'null' | moai statusline` — 정상 출력.

### EC-SF-002: Cache File Directory 없음

**When** `~/.moai/state/` directory가 존재하지 않을 때 `WriteModelCache()`가 호출되면, `moai statusline` **shall** directory를 자동 생성하고 cache file을 정상 작성한다.

**Verification**: `t.TempDir()` 환경에서 cache write → file 존재 확인.

### EC-SF-003: Cache File Write 실패

**When** cache directory에 write 권한이 없을 때, `WriteModelCache()` **shall** error를 silent ignore하고 statusline 출력에 영향이 없다.

**Verification**: readonly dir에서 `WriteModelCache()` — error 무시 확인.

### EC-SF-004: `MOAI_LAST_MODEL` 빈 문자열

**When** 환경변수 `MOAI_LAST_MODEL`이 빈 문자열(`""`)로 설정되어 있으면, `moai statusline` **shall** 빈 문자열을 유효하지 않은 것으로 간주하고 cache file fallback으로 진행한다.

**Verification**: `echo '{}' | MOAI_LAST_MODEL="" moai statusline` 동작 확인.

### EC-SF-005: Cache File Corrupted

**When** `~/.moai/state/last-model.txt`에 빈 내용이나 binary data가 있으면, `ReadModelCache()` **shall** 빈 문자열을 반환하고 model name은 미표시된다.

**Verification**: corrupt file로 unit test — crash 없음 확인.

## 3. Quality Gates

### AC-SF-QG-001: 테스트 커버리지

- `internal/statusline/` 패키지 coverage >= 85%
- `internal/cli/statusline.go` coverage >= 85%
- Race detector: `go test -race ./internal/statusline/... ./internal/cli/...` 통과

### AC-SF-QG-002: Lint

- `golangci-lint run ./internal/statusline/... ./internal/cli/...` zero errors
- `go vet ./internal/statusline/... ./internal/cli/...` zero errors

### AC-SF-QG-003: 성능

- Empty stdin fallback path의 추가 latency < 1ms (env read + getcwd + file read)
- 기존 happy path latency 변화 없음 (측정 가능한 regression 없음)

### AC-SF-QG-004: Cross-Platform

- macOS: `os.Getwd()` 정상 동작
- Linux: 동일
- Windows: path separator 차이 고려 (`filepath.Base` 사용으로 이미 대응)

## 4. Definition of Done

- [ ] AC-SF-001 ~ AC-SF-008 모두 통과
- [ ] EC-SF-001 ~ EC-SF-005 모두 통과
- [ ] AC-SF-QG-001 ~ AC-SF-QG-004 모두 통과
- [ ] `go test -race ./internal/statusline/... ./internal/cli/...` 통과
- [ ] 기존 statusline golden test 모두 통과 (regression 없음)
- [ ] MX tags 5개 이상 적용 (plan.md §5 참조)
