# Implementation Plan: SPEC-SRS-001

## Task Decomposition

### Group A: 데드코드 삭제 (REQ-1, REQ-2, REQ-3) - 병렬 가능

**Task A1: 죽은 패키지 삭제 (REQ-1)**
- 삭제: `internal/core/temp.go`, `pathutil_nonwindows.go`, `pathutil_windows.go`
- 검증: `go build ./...`, `go test ./...`
- 예상: 15분

**Task A2: 미사용 export 삭제 (REQ-2)**
- 각 파일에서 대상 식별자 삭제
- 해당 식별자를 사용하는 테스트 코드 업데이트
- 검증: `go build ./...`, `go test ./...`
- 예상: 30분

**Task A3: 중복 타임아웃 삭제 (REQ-3)**
- `internal/defs/timeouts.go` 전체 삭제
- 혹시 참조하는 코드가 있으면 `foundation` 패키지로 마이그레이션
- 검증: `go build ./...`, `go test ./...`
- 예상: 10분

### Group B: SPEC ID 정규식 통합 (REQ-4)

**Task B1: SpecIDPattern export**
- `internal/workflow/specid.go`에 `SpecIDPattern = regexp.MustCompile(...)` 추가
- 패턴: `SPEC-[A-Z][A-Z0-9]*-\d+` (대문자 시작, 이후 대문자+숫자)
- 테스트 추가

**Task B2: Hook 파일 마이그레이션**
- `internal/hook/user_prompt_submit.go`: 로컬 regex → `workflow.SpecIDPattern`
- `internal/hook/task_completed.go`: 동일
- 기존 테스트 통과 확인
- circular import 없음 검증 (workflow는 hook에 의존하지 않음)

### Group C: 설정 스키마 (REQ-7)

**Task C1: ResearchConfig struct**
- `internal/config/types.go`에 `ResearchConfig` struct 추가
- `Config` root struct에 `Research` 필드 추가
- `defaults.go`에 기본값 추가

**Task C2: research.yaml 템플릿**
- `internal/template/templates/.moai/config/sections/research.yaml` 생성
- `make build`로 embedded 재생성

### Group D: Eval 엔진 (REQ-5)

**Task D1: types.go + suite.go**
- 기본 타입 정의 (EvalSuite, EvalCriterion, TargetSpec, TestInput, CriterionWeight)
- YAML 로드/파싱 구현
- 테이블 드리븐 테스트

**Task D2: criterion.go + result.go**
- CriterionResult 판정 로직
- EvalResult 점수 집계 (Overall, MustPassOK)
- 테이블 드리븐 테스트

**Task D3: engine.go**
- EvalEngine 인터페이스 + 기본 구현
- LoadSuite, Evaluate 메서드
- 통합 테스트

### Group E: Safety 레이어 (REQ-6)

**Task E1: frozen.go**
- FrozenGuard 구현
- 기본 frozen 경로 목록
- 테스트

**Task E2: canary.go**
- CanaryChecker 구현
- baseline 비교 로직
- 테스트

**Task E3: limiter.go**
- RateLimiter 구현
- JSONL 액션 기록
- 테스트

## Dependency Graph

```
A1, A2, A3 (병렬) ─→ B1, B2 (순차) ─→ C1, C2 (순차) ─→ D1, D2, D3 (순차)
                                                        ─→ E1, E2, E3 (순차)
                                                        (D와 E는 병렬 가능)
```

## Technology Stack

- Go 1.26 (go.mod)
- gopkg.in/yaml.v3 (YAML 파싱, 이미 의존성)
- 추가 외부 의존성 없음

## Quality Gates

- `go test -race ./...` 전체 통과
- `go vet ./...` 통과
- `golangci-lint run` 통과
- 새 패키지 테스트 커버리지 85%+
- `make build` 성공 (템플릿 재생성)
