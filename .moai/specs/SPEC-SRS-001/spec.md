---
id: SPEC-SRS-001
version: "1.0.0"
status: completed
created: "2026-04-09"
updated: "2026-04-09"
author: GOOS
priority: high
issue_number: 0
---

# SPEC-SRS-001: 데드코드 정리 + Self-Research 기반 패키지

## HISTORY

| Date | Version | Change |
|------|---------|--------|
| 2026-04-09 | 1.0.0 | Initial SPEC from DESIGN-SRS-002 Phase 0 + Phase 1 |

## Overview

moai-adk-go 코드베이스의 데드코드를 정리하고, Self-Research System의 기반 Go 패키지(`internal/research/eval`, `internal/research/safety`)를 TDD로 구현한다.

Source: `.moai/plans/DESIGN-SRS-002-unified.md` Part 2 + Part 3.3

## What NOT to Build (Exclusions)

- Experiment loop (`internal/research/experiment/`) - Phase 2에서 구현
- Passive observation (`internal/research/observe/`) - Phase 3에서 구현
- Dashboard (`internal/research/dashboard/`) - Phase 5에서 구현
- Agency 통합 - Phase 4에서 구현
- CLI 명령 (`moai research`) - Phase 2에서 구현
- Researcher 에이전트/스킬 정의 - Phase 2에서 구현

---

## Requirements

### REQ-1: 죽은 패키지 삭제

**[EARS: Ubiquitous]**

The system SHALL NOT contain the `internal/core` top-level package (`temp.go`, `pathutil_nonwindows.go`, `pathutil_windows.go`) as it has zero importers.

Acceptance:
- `internal/core/temp.go`, `internal/core/pathutil_nonwindows.go`, `internal/core/pathutil_windows.go` 삭제
- 하위 패키지 (`internal/core/git/`, `internal/core/project/`, `internal/core/quality/`)는 유지
- `go build ./...` 성공
- `go test ./...` 전체 통과

### REQ-2: 미사용 export 삭제

**[EARS: Ubiquitous]**

The system SHALL NOT export identifiers that have zero callers outside their package.

삭제 대상:

| 파일 | 식별자 |
|------|--------|
| `internal/core/project/validator.go` | `BackupTimestampFormat`, `BackupsDir` |
| `internal/foundation/timeouts.go` | `DefaultCLITimeout`, `DefaultSearchTimeout`, `DefaultLSPTimeout` |
| `internal/foundation/errors.go` | `ErrInvalidRequirementType`, `ErrInvalidPillar`, `ErrAssessmentFailed`, `ErrInvalidPhaseTransition`, `ErrUnsupportedLanguage`, `RequirementNotFoundError`, `LanguageNotFoundError` |
| `internal/shell/errors.go` | `ErrUnsupportedShell`, `ErrConfigNotFound` |
| `internal/defs/perms.go` | `CredDirPerm`, `CredFilePerm` |
| `internal/defs/files.go` | `GithubSpecRegistryJSON`, `MCPJSON` |
| `internal/defs/paths.go` | `StatusLinePath` |
| `internal/defs/dirs.go` | `SpecsSubdir`, `ReportsSubdir` |

Acceptance:
- 모든 대상 식별자가 삭제됨
- 해당 식별자를 참조하는 테스트가 있으면 함께 업데이트
- `go build ./...` 성공
- `go test ./...` 전체 통과

### REQ-3: 중복 타임아웃 상수 통합

**[EARS: Ubiquitous]**

The system SHALL have a single source of timeout constants. `internal/defs/timeouts.go` SHALL be deleted as all used timeouts are in `internal/foundation/timeouts.go`.

Acceptance:
- `internal/defs/timeouts.go` 파일 삭제 (`HookDefaultTimeout`, `HookPostToolTimeout`, `GitShortTimeout`, `GitLongTimeout`)
- `defs.HookDefaultTimeout` 등을 참조하는 코드가 있으면 `foundation` 패키지로 마이그레이션
- `go test ./...` 전체 통과

### REQ-4: SPEC ID 정규식 통합

**[EARS: Ubiquitous]**

The system SHALL use a single canonical SPEC ID regex defined in `internal/workflow/specid.go`.

현재 중복:
- `internal/hook/user_prompt_submit.go:21` → `SPEC-[A-Z0-9]+-\d+`
- `internal/hook/task_completed.go:15` → `SPEC-[A-Z]+-\d+`
- `internal/workflow/specid.go` (기존)

Acceptance:
- `internal/workflow/specid.go`에 `SpecIDPattern` 변수 export
- `user_prompt_submit.go`와 `task_completed.go`가 `workflow.SpecIDPattern`을 import하여 사용
- 정규식은 `[A-Z][A-Z0-9]*` (대문자로 시작, 이후 대문자+숫자 허용)로 통일
- `go test ./...` 전체 통과

### REQ-5: Binary Eval 엔진 (`internal/research/eval`)

**[EARS: Event-driven]**

WHEN an eval suite YAML file is loaded, the system SHALL parse it into an `EvalSuite` struct and validate all criteria are binary (pass/fail).

```go
// 핵심 인터페이스
type EvalEngine interface {
    LoadSuite(path string) (*EvalSuite, error)
    Evaluate(suite *EvalSuite, outputs map[string][]byte) (*EvalResult, error)
}

type EvalSuite struct {
    Target   TargetSpec
    Inputs   []TestInput
    Criteria []EvalCriterion
    Settings EvalSettings
}

type EvalCriterion struct {
    Name     string
    Question string           // Yes/No 질문
    Pass     string           // 통과 조건 설명
    Fail     string           // 실패 조건 설명
    Weight   CriterionWeight  // must_pass | nice_to_have
}

type EvalResult struct {
    Overall      float64
    PerCriterion map[string]CriterionResult
    MustPassOK   bool
    Timestamp    time.Time
}
```

Acceptance:
- `internal/research/eval/` 패키지 생성 (`engine.go`, `suite.go`, `criterion.go`, `result.go`, `types.go`)
- `EvalSuite` YAML 로드/파싱 구현
- `EvalResult` 점수 집계 구현: `Overall = (pass_count / total_count)`, `MustPassOK = all must_pass criteria passed`
- 테이블 드리븐 테스트 85%+ 커버리지
- `go vet ./internal/research/...` 통과
- eval suite YAML 스키마 예시 파일 1개 포함

### REQ-6: Safety 레이어 (`internal/research/safety`)

**[EARS: State-driven]**

WHILE the research system is active, the system SHALL enforce frozen file protection, canary regression checks, and rate limiting.

```go
// FrozenGuard
type FrozenGuard interface {
    IsFrozen(path string) bool
    ValidateWrite(path string) error
}

// CanaryChecker
type CanaryChecker interface {
    Check(baselines []Baseline, proposed EvalResult, threshold float64) (bool, error)
}

// RateLimiter
type RateLimiter interface {
    CheckLimit(config RateLimitConfig) error
    RecordAction(actionType string) error
}
```

Acceptance:
- `internal/research/safety/` 패키지 생성 (`frozen.go`, `canary.go`, `limiter.go`, `types.go`)
- `FrozenGuard`: constitution.md, constitution 관련 파일, researcher 자신 등 frozen 경로 목록 관리
- `CanaryChecker`: baseline 배열과 proposed result를 비교, threshold(기본 0.10) 초과 하락 시 false 반환
- `RateLimiter`: 세션당/주간 한도 체크, JSONL 파일로 액션 기록
- 테이블 드리븐 테스트 85%+ 커버리지

### REQ-7: Research 설정 스키마 (`research.yaml`)

**[EARS: Ubiquitous]**

The system SHALL define a `research.yaml` configuration section loadable by `internal/config`.

```yaml
research:
  enabled: true
  passive:
    enabled: true
    correction_window_seconds: 60
    pattern_thresholds:
      heuristic: 3
      rule: 5
      high_confidence: 10
  active:
    defaults:
      runs_per_experiment: 3
      max_experiments: 20
      pass_threshold: 0.80
      target_score: 0.95
      budget_cap_tokens: 500000
  safety:
    worktree_isolation: true
    canary_regression_threshold: 0.10
    rate_limits:
      max_experiments_per_session: 20
      max_accepted_per_session: 5
      max_auto_research_per_week: 3
  dashboard:
    default_mode: terminal
    html_open_browser: true
```

Acceptance:
- `internal/config/types.go`에 `ResearchConfig` struct 추가
- `.moai/config/sections/research.yaml` 기본 템플릿 생성
- `internal/template/templates/.moai/config/sections/research.yaml` 템플릿 파일 생성
- `ConfigManager.Load()`에서 research 섹션 로드 가능
- 설정 로드 테스트 통과

---

## Technical Approach

### 실행 순서

1. REQ-1, REQ-2, REQ-3 (데드코드 삭제) - 독립적, 병렬 가능
2. REQ-4 (SPEC ID 통합) - hook 패키지 수정 필요
3. REQ-7 (설정 스키마) - REQ-5, REQ-6의 전제
4. REQ-5 (Eval 엔진) - 새 패키지 생성
5. REQ-6 (Safety 레이어) - 새 패키지 생성

### 파일 수정 범위

**삭제:**
- `internal/core/temp.go`
- `internal/core/pathutil_nonwindows.go`
- `internal/core/pathutil_windows.go`
- `internal/defs/timeouts.go`

**수정:**
- `internal/core/project/validator.go` (deprecated export 삭제)
- `internal/foundation/timeouts.go` (미사용 export 삭제)
- `internal/foundation/errors.go` (미사용 에러 타입 삭제)
- `internal/shell/errors.go` (미사용 에러 삭제)
- `internal/defs/perms.go`, `files.go`, `paths.go`, `dirs.go` (미사용 상수 삭제)
- `internal/hook/user_prompt_submit.go` (SPEC ID regex → workflow 패키지 참조)
- `internal/hook/task_completed.go` (SPEC ID regex → workflow 패키지 참조)
- `internal/workflow/specid.go` (SpecIDPattern export 추가)
- `internal/config/types.go` (ResearchConfig 추가)
- `internal/config/defaults.go` (research 기본값)

**생성:**
- `internal/research/eval/` (5개 파일 + 5개 테스트)
- `internal/research/safety/` (4개 파일 + 4개 테스트)
- `internal/research/types.go`
- `.moai/config/sections/research.yaml`
- `internal/template/templates/.moai/config/sections/research.yaml`

### [DELTA] 기존 코드 변경

- [REMOVE] `internal/core/temp.go`, `pathutil_*.go` - 0 callers, 안전하게 삭제
- [REMOVE] `internal/defs/timeouts.go` - foundation으로 대체됨
- [MODIFY] `internal/hook/user_prompt_submit.go` - SPEC ID regex를 workflow 패키지에서 import
- [MODIFY] `internal/hook/task_completed.go` - 동일
- [MODIFY] `internal/workflow/specid.go` - SpecIDPattern export 추가
- [MODIFY] `internal/config/types.go` - ResearchConfig struct 추가
- [NEW] `internal/research/eval/` - Binary Eval 엔진
- [NEW] `internal/research/safety/` - Safety 레이어

### 위험 분석

| 위험 | 영향 | 완화 |
|------|------|------|
| 미사용으로 판단한 export가 실제 사용중 | 빌드 실패 | 삭제 전 `go build ./...` 확인 |
| SPEC ID regex 변경으로 기존 동작 변경 | 훅 동작 변경 | 기존 테스트 + 새 테스트로 커버 |
| circular import (hook → workflow) | 컴파일 실패 | workflow 패키지는 hook에 의존하지 않음 확인 |
| config 변경으로 기존 설정 로드 실패 | 초기화 실패 | research는 optional section, 없어도 로드 성공 |
