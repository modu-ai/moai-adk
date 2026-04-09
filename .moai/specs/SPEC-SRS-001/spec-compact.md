# SPEC-SRS-001 Compact

## Requirements

### REQ-1: 죽은 패키지 삭제
- DELETE: `internal/core/temp.go`, `pathutil_nonwindows.go`, `pathutil_windows.go`
- KEEP: `internal/core/git/`, `internal/core/project/`, `internal/core/quality/`

### REQ-2: 미사용 export 삭제
- `internal/core/project/validator.go`: BackupTimestampFormat, BackupsDir
- `internal/foundation/timeouts.go`: DefaultCLITimeout, DefaultSearchTimeout, DefaultLSPTimeout
- `internal/foundation/errors.go`: ErrInvalidRequirementType, ErrInvalidPillar, ErrAssessmentFailed, ErrInvalidPhaseTransition, ErrUnsupportedLanguage, RequirementNotFoundError, LanguageNotFoundError
- `internal/shell/errors.go`: ErrUnsupportedShell, ErrConfigNotFound
- `internal/defs/perms.go`: CredDirPerm, CredFilePerm
- `internal/defs/files.go`: GithubSpecRegistryJSON, MCPJSON
- `internal/defs/paths.go`: StatusLinePath
- `internal/defs/dirs.go`: SpecsSubdir, ReportsSubdir

### REQ-3: 중복 타임아웃 상수 통합
- DELETE: `internal/defs/timeouts.go` (전체)
- Foundation 패키지에 이미 사용중인 값 존재

### REQ-4: SPEC ID 정규식 통합
- EXPORT: `workflow.SpecIDPattern` in `internal/workflow/specid.go`
- MIGRATE: `internal/hook/user_prompt_submit.go`, `internal/hook/task_completed.go`
- PATTERN: `SPEC-[A-Z][A-Z0-9]*-\d+`

### REQ-5: Binary Eval 엔진
- CREATE: `internal/research/eval/` (engine.go, suite.go, criterion.go, result.go, types.go)
- Interface: EvalEngine { LoadSuite, Evaluate }
- Types: EvalSuite, EvalCriterion (must_pass/nice_to_have), EvalResult

### REQ-6: Safety 레이어
- CREATE: `internal/research/safety/` (frozen.go, canary.go, limiter.go, types.go)
- FrozenGuard: IsFrozen, ValidateWrite
- CanaryChecker: Check (threshold 0.10)
- RateLimiter: CheckLimit, RecordAction

### REQ-7: Research 설정 스키마
- ADD: ResearchConfig to `internal/config/types.go`
- CREATE: `.moai/config/sections/research.yaml` template
- Optional section (missing = default values)

## Files to Modify

DELETE: `internal/core/temp.go`, `internal/core/pathutil_nonwindows.go`, `internal/core/pathutil_windows.go`, `internal/defs/timeouts.go`

MODIFY: `internal/core/project/validator.go`, `internal/foundation/timeouts.go`, `internal/foundation/errors.go`, `internal/shell/errors.go`, `internal/defs/perms.go`, `internal/defs/files.go`, `internal/defs/paths.go`, `internal/defs/dirs.go`, `internal/hook/user_prompt_submit.go`, `internal/hook/task_completed.go`, `internal/workflow/specid.go`, `internal/config/types.go`, `internal/config/defaults.go`

CREATE: `internal/research/eval/*.go`, `internal/research/safety/*.go`, `internal/research/types.go`, `.moai/config/sections/research.yaml`, `internal/template/templates/.moai/config/sections/research.yaml`

## Exclusions
- Experiment loop (Phase 2)
- Passive observation (Phase 3)
- Dashboard (Phase 5)
- Agency integration (Phase 4)
- CLI command `moai research` (Phase 2)
- Researcher agent/skill definitions (Phase 2)
