# Acceptance Criteria: SPEC-SRS-001

## AC-1: 데드코드 완전 제거

**Given** moai-adk-go 코드베이스에 미사용 코드가 존재할 때
**When** 데드코드 정리가 완료되면
**Then**
- `internal/core/temp.go`, `pathutil_nonwindows.go`, `pathutil_windows.go`가 존재하지 않는다
- `internal/defs/timeouts.go`가 존재하지 않는다
- REQ-2에 나열된 모든 미사용 export가 삭제되었다
- `go build ./...`가 성공한다
- `go test ./...`가 전체 통과한다

## AC-2: SPEC ID 정규식 단일 소스

**Given** SPEC ID 정규식이 3곳에 분산되어 있을 때
**When** 정규식 통합이 완료되면
**Then**
- `internal/workflow/specid.go`에 `SpecIDPattern` 변수가 export된다
- `internal/hook/user_prompt_submit.go`가 `workflow.SpecIDPattern`을 사용한다
- `internal/hook/task_completed.go`가 `workflow.SpecIDPattern`을 사용한다
- 로컬 regex 정의가 두 hook 파일에서 제거되었다
- `SPEC-CC297-001`, `SPEC-AUTH-001`, `SPEC-SRS-001` 등 기존 ID가 모두 매칭된다
- `go test ./internal/workflow/... ./internal/hook/...` 전체 통과한다

## AC-3: Eval 엔진 동작

**Given** eval suite YAML 파일이 존재할 때
**When** `EvalEngine.LoadSuite(path)` 호출하면
**Then**
- EvalSuite 구조체가 올바르게 파싱된다
- 모든 criteria의 Weight가 `must_pass` 또는 `nice_to_have`이다
- 잘못된 YAML은 명확한 에러를 반환한다

**Given** eval suite와 출력 데이터가 있을 때
**When** `EvalEngine.Evaluate(suite, outputs)` 호출하면
**Then**
- `Overall`이 0.0~1.0 범위의 float64이다
- `MustPassOK`가 모든 must_pass criteria 통과 여부를 정확히 반영한다
- `PerCriterion` 맵에 모든 criteria 결과가 포함된다

## AC-4: Safety 레이어 동작

**Given** frozen 경로 목록에 `moai-constitution.md`가 포함될 때
**When** `FrozenGuard.IsFrozen("moai-constitution.md")` 호출하면
**Then** `true`를 반환한다

**Given** baseline 점수가 0.85이고 proposed 점수가 0.70일 때 (0.15 하락)
**When** `CanaryChecker.Check(baselines, proposed, 0.10)` 호출하면
**Then** `false`를 반환한다 (0.10 threshold 초과)

**Given** 주간 한도가 3이고 이번 주 2회 실행했을 때
**When** `RateLimiter.CheckLimit(config)` 호출하면
**Then** 에러 없이 통과한다

**Given** 주간 한도가 3이고 이번 주 3회 실행했을 때
**When** `RateLimiter.CheckLimit(config)` 호출하면
**Then** `ErrRateLimitExceeded` 에러를 반환한다

## AC-5: Research 설정 로드

**Given** `.moai/config/sections/research.yaml`이 존재할 때
**When** `ConfigManager.Load()` 호출하면
**Then** `Config.Research` 필드에 research 설정이 로드된다

**Given** `research.yaml`이 존재하지 않을 때
**When** `ConfigManager.Load()` 호출하면
**Then** 기본값으로 초기화되고 에러 없이 로드된다

## AC-6: 전체 품질 게이트

**Given** 모든 변경이 완료된 후
**When** 품질 검증을 실행하면
**Then**
- `go test -race ./...` 전체 통과
- `go vet ./...` 통과
- `golangci-lint run` 통과
- 새 패키지 (`internal/research/eval/`, `internal/research/safety/`) 커버리지 85%+
- `make build` 성공
