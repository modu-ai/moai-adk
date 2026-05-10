---
id: SPEC-V3R3-STATUSLINE-FALLBACK-001
version: "0.1.0"
status: draft
created_at: 2026-05-10
updated_at: 2026-05-10
author: GOOS행님
priority: High
labels: [statusline, fallback, cli]
issue_number: null
---

# Implementation Plan: SPEC-V3R3-STATUSLINE-FALLBACK-001

## 1. Milestone 분해

### M1: Cwd Guard + Project Directory Fallback (P0)

**목표**: `runStatusline()` 시작 부분에 cwd guard 추가, `extractProjectDirectory()` getcwd fallback 구현

**변경 파일**:
- `internal/cli/statusline.go` — `runStatusline()` 함수 시작 부분에 `os.Getwd()` 검증 추가
  - `os.Getwd()` 실패 또는 존재하지 않는 directory → `os.Chdir(os.Getenv("HOME"))` 또는 `os.UserHomeDir()`
  - cwd guard 성공/실패 여부를 `statusline.Options`에 전달 (예: `CwdFallback bool` 또는 별도 flag)
- `internal/statusline/builder.go` — `extractProjectDirectory()` 수정
  - 기존 3-level 우선순위 유지
  - 새로운 4번째 fallback: `os.Getwd()` → `filepath.Base()`
- `internal/cli/statusline_test.go` — cwd guard 테스트 (t.TempDir() + 삭제 시나리오)

**테스트 전략**:
- `TestCwdGuard_DeletedDirectory` — 임시 dir 생성 후 삭제 → HOME fallback 확인
- `TestExtractProjectDirectory_GetwdFallback` — nil input → getcwd basename 확인
- `TestExtractProjectDirectory_SandboxCwdPriority` — stdin cwd ≠ getcwd → stdin 우선

**MX Tag 대상**:
- `internal/cli/statusline.go`: `@MX:ANCHOR` on `runStatusline` (public entry point, fan_in >= 3)
- `internal/statusline/builder.go`: `@MX:NOTE` on `extractProjectDirectory` (fallback chain 설명)

### M2: Model Name Fallback Chain (P0)

**목표**: stdin model 누락 시 env/cache fallback 구현, 정상 수신 시 cache write

**변경 파일**:
- `internal/statusline/model_cache.go` (신규)
  - `ReadModelCache(homeDir string) string` — `~/.moai/state/last-model.txt` 읽기
  - `WriteModelCache(homeDir, modelName string) error` — 원자적 쓰기 (temp + rename)
  - 부모 directory 자동 생성 (`os.MkdirAll`)
- `internal/statusline/metrics.go` — `CollectMetrics()` 수정
  - `input.Model == nil` 또는 model name 비어있을 때 fallback chain:
    1. `os.Getenv("MOAI_LAST_MODEL")` 조회
    2. `ReadModelCache(homeDir)` 조회
    3. 둘 다 없으면 `Available: false` (기존 동작)
  - 정상 model 수신 시 `WriteModelCache()` 호출 (goroutine 없이 동기식)
- `internal/statusline/builder.go` — `New()` 옵션에 `HomeDir` 전달 이미 됨 (기존 `usageProvider`용)
  - `collectAll()`에서 `CollectMetrics(input)` 호출 시 `homeDir` 전달 방식 변경 필요
  - `CollectMetrics` signature 변경: `CollectMetrics(input *StdinData, homeDir string) *MetricsData`
- `internal/statusline/metrics_test.go` — fallback chain 테스트

**테스트 전략**:
- `TestCollectMetrics_NilInput_EnvFallback` — env var 설정 → model name 확인
- `TestCollectMetrics_NilInput_CacheFallback` — cache file 생성 → model name 확인
- `TestCollectMetrics_NilInput_NoFallback` — env/cache 모두 없음 → Available: false
- `TestCollectMetrics_ValidInput_CacheWrite` — 정상 stdin → cache file 생성 확인
- `TestWriteModelCache_AtomicReplace` — 기존 파일 원자적 교체 확인
- `TestWriteModelCache_DirCreation` — 부모 dir 없음 → 자동 생성 확인

**MX Tag 대상**:
- `internal/statusline/model_cache.go`: `@MX:ANCHOR` on `ReadModelCache`/`WriteModelCache`
- `internal/statusline/metrics.go`: `@MX:NOTE` on `CollectMetrics` (fallback chain 설명)

### M3: Builder Fallback 통합 + Template 정리 (P1/P2)

**목표**: `collectAll()` nil input 시 모든 fallback이 올바르게 동작하는지 통합 테스트, template 선택적 정리

**변경 파일**:
- `internal/statusline/builder.go` — `collectAll()` 수정
  - `input == nil`일 때 `os.Getwd()`로 directory fallback
  - `CollectMetrics`에 `homeDir` 전달
- `internal/statusline/builder_test.go` (신규 또는 확장)
  - `TestCollectAll_NilInput` — 모든 collector nil 동작 검증
  - `TestCollectAll_PartialInput` — 일부 필드만 있는 JSON
  - `TestCollectAll_EmptyJSON` — `{}` 입력
  - `TestCollectAll_NullJSON` — `null` 입력
  - `TestBuild_DeletedCwdFallback` — cwd 삭제 → HOME fallback + project name `~`
- `internal/template/templates/.moai/status_line.sh.tmpl` — cwd guard 제거 (선택)
  - Go binary가 cwd guard를 수행하므로 shell-level guard는 중복
  - 단, 기존 배포된 wrapper와의 호환성을 위해 즉시 제거하지 않고 deprecation comment 추가도 고려

**테스트 전략**:
- 위 builder_test.go 테스트 케이스
- `TestBuild_OutputFormat` — nil/partial/normal stdin 모두 3-line 형식 준수 확인
- Race detector: `go test -race ./internal/statusline/...`

**MX Tag 대상**:
- `internal/statusline/builder.go`: `@MX:ANCHOR` on `Build` (core pipeline, fan_in >= 5)

## 2. 기술 제약

| 제약 | 근거 |
|------|------|
| Zero-regression | 기존 happy path 성능/출력 변경 금지 |
| 표준 라이브러리만 사용 | 외부 dependency 도입 금지 |
| 동기식 fallback | env read, getcwd, file read 모두 마이크로초 단위 — goroutine 불필요 |
| Race-safe cache write | Build 파이프라인 내에서 순차 실행 |
| t.TempDir() 필수 | 테스트 임시 파일은 모두 `/tmp` 하위 |

## 3. 위험 분석

| 위험 | 확률 | 영향 | 완화 방안 |
|------|------|------|-----------|
| Cache file write race (동시 moai statusline 프로세스) | 낮음 | 낮음 (마지막 write 승리, data 손실 없음) | Atomic rename으로 corruption 방지 |
| HOME env var 미설정 (CI 등) | 낮음 | 중간 | `os.UserHomeDir()` fallback 사용 |
| `os.Getwd()` 성공했으나 실제로 삭제된 dir인 경우 | 극히 낮음 | 중간 | `os.Stat()`으로 존재 확인 후 Chdir |
| CollectMetrics signature 변경으로 인한 기존 테스트 break | 중간 | 낮음 | 기존 호출부 모두 업데이트 |

## 4. 의존성

- M1과 M2는 독립적 — 병렬 개발 가능
- M3는 M1, M2 완료 후 통합
- Template 변경(M3의 선택 부분)은 `make build` 필요

## 5. MX Tag 요약

| 파일 | Tag Type | 대상 |
|------|----------|------|
| `internal/cli/statusline.go` | `@MX:ANCHOR` | `runStatusline` |
| `internal/statusline/builder.go` | `@MX:ANCHOR` | `Build`, `collectAll` |
| `internal/statusline/builder.go` | `@MX:NOTE` | `extractProjectDirectory` |
| `internal/statusline/metrics.go` | `@MX:NOTE` | `CollectMetrics` |
| `internal/statusline/model_cache.go` | `@MX:ANCHOR` | `ReadModelCache`, `WriteModelCache` |
