---
spec_id: SPEC-HOOK-001
title: Compiled Hook System - Implementation Plan
version: "1.0.0"
created: 2026-02-03
updated: 2026-02-03
author: GOOS
tags: "hook, implementation-plan, binary-subcommand, migration"
---

# SPEC-HOOK-001: Implementation Plan

## 1. Overview

### 1.1 Scope

Python 기반 46개 훅 스크립트(21,535 LOC)를 Go 컴파일 바이너리 서브커맨드(~1,500 LOC)로 교체한다. `internal/hook/` 패키지에 9개 파일을 구현하며, Phase 1 Foundation의 P0 Critical 모듈이다.

### 1.2 Implementation Strategy

**Bottom-Up 접근**: 핵심 인프라(Registry, Protocol, Contract)를 먼저 구현한 뒤, 6개 이벤트 핸들러를 순차적으로 구현한다. 각 단계는 독립적으로 테스트 가능하며, 이전 단계의 인터페이스에 의존한다.

### 1.3 Dependencies

| Dependency       | Status  | Blocking | Impact                                     |
|------------------|---------|----------|--------------------------------------------|
| SPEC-CONFIG-001  | Planned | Yes      | ConfigManager 인터페이스 필요                |
| `log/slog`       | Available | No     | Go 1.22 표준 라이브러리                      |
| `encoding/json`  | Available | No     | Go 표준 라이브러리                           |
| `context`        | Available | No     | Go 표준 라이브러리                           |

---

## 2. Task Decomposition

### Milestone 1: Core Infrastructure (Primary Goal)

핵심 인프라 3개 파일 구현. 모든 핸들러의 기반이 되는 레지스트리, 프로토콜, 계약 시스템.

#### Task 1.1: Registry (`registry.go`)

**Priority**: High

**Description**: 타입 안전한 핸들러 레지스트리 구현. EventType을 키로 핸들러 슬라이스를 관리하며, Dispatch 시 순차 실행 및 block 단축 반환(short-circuit) 로직을 포함한다.

**Implementation Details**:
- `registry` 구조체: `handlers map[EventType][]Handler`, `cfg config.ConfigManager`
- `Register(handler Handler)`: 이벤트 타입별 핸들러 슬라이스에 추가
- `Dispatch(ctx, event, input)`: 타임아웃 설정, 순차 실행, block 시 즉시 반환
- `Handlers(event)`: 등록된 핸들러 슬라이스 반환
- 타임아웃: config에서 `HookTimeout` 읽기, 기본값 30초
- `context.WithTimeout`을 통한 데드라인 관리

**Testing**:
- 핸들러 등록/조회 단위 테스트
- 빈 레지스트리 디스패치 테스트
- Block 단축 반환(short-circuit) 테스트
- 타임아웃 초과 시 `ErrHookTimeout` 반환 테스트
- 병렬 안전성(`t.Parallel()`) 테스트

**Covered Requirements**: REQ-HOOK-001, REQ-HOOK-002, REQ-HOOK-003, REQ-HOOK-004

#### Task 1.2: Protocol (`protocol.go`)

**Priority**: High

**Description**: Claude Code JSON stdin/stdout 통신 프로토콜 구현. `io.Reader`에서 JSON 파싱, `io.Writer`로 JSON 직렬화.

**Implementation Details**:
- `ReadInput(r io.Reader) (*HookInput, error)`: `json.NewDecoder(r).Decode()`
- 필수 필드 검증: `session_id`, `cwd`, `hook_event_name`
- `WriteOutput(w io.Writer, output *HookOutput) error`: `json.NewEncoder(w).Encode()`
- stdin EOF 처리: 빈 입력 시 기본 HookInput 반환 또는 오류 반환
- 디코딩 오류 시 `ErrHookInvalidInput` 래핑

**Testing**:
- 유효한 JSON 파싱 테스트
- 필수 필드 누락 시 검증 오류 테스트
- 빈 stdin 처리 테스트
- Malformed JSON 오류 처리 테스트
- JSON 출력 유효성(`json.Valid`) 검증 테스트
- 대용량 JSON 입력 테스트

**Covered Requirements**: REQ-HOOK-010, REQ-HOOK-011, REQ-HOOK-012, REQ-HOOK-013

#### Task 1.3: Contract (`contract.go`)

**Priority**: High

**Description**: Hook Execution Contract(ADR-012) 구현. 실행 환경 검증, 보증/비보증 사항 명세.

**Implementation Details**:
- `Validate(ctx context.Context) error`: 작업 디렉터리 존재 확인, 필수 환경 확인
- `Guarantees() []string`: 보증 사항 문자열 목록 반환
- `NonGuarantees() []string`: 비보증 사항 문자열 목록 반환
- 검증 실패 시 `ErrHookContractFail` 반환과 slog 로깅
- 디버그 레벨에서 현재 환경 상태(PATH, CWD, OS) 로깅

**Testing**:
- Contract 검증 성공/실패 단위 테스트
- 보증/비보증 목록 완전성 테스트
- Minimal PATH 환경에서의 계약 테스트 (`exec.Command`)
- 비대화형 셸 환경 시뮬레이션 테스트

**Covered Requirements**: REQ-HOOK-020, REQ-HOOK-021, REQ-HOOK-022

---

### Milestone 2: Event Handlers (Secondary Goal)

6개 이벤트 핸들러 구현. 각 핸들러는 `Handler` 인터페이스를 구현하며 독립적으로 테스트 가능하다.

#### Task 2.1: SessionStart Handler (`session_start.go`)

**Priority**: High

**Description**: 세션 초기화, 설정 로드, 환경 검증.

**Implementation Details**:
- `Handle(ctx, input)`: ConfigManager로 설정 로드, 프로젝트 루트 검증
- 세션 ID 기록(slog)
- 작업 디렉터리 유효성 확인
- 프로젝트 정보(이름, 버전, 설정 상태)를 Data에 JSON 직렬화
- 오류 발생 시 비차단(non-blocking) 처리: 오류 로깅 후 allow 반환

**Testing**: 정상 세션 시작, 설정 누락 시 graceful degradation, 잘못된 CWD 처리

**Covered Requirements**: REQ-HOOK-030

#### Task 2.2: PreToolUse Handler (`pre_tool.go`)

**Priority**: High

**Description**: 도구 사용 권한 검증, 보안 정책 적용.

**Implementation Details**:
- `Handle(ctx, input)`: `tool_name` 기반 보안 정책 조회
- 차단 목록(blocklist) 확인: 금지된 도구 사용 시 block 반환
- 허용 목록(allowlist) 확인: 선택적 허용 모드
- `tool_input` 검증: 위험한 입력 패턴 감지
- block 시 `Decision: "block"`, `Reason`에 차단 사유 포함

**Testing**: 허용 도구 실행, 차단 도구 실행, 알 수 없는 도구 처리, 빈 tool_name 처리

**Covered Requirements**: REQ-HOOK-031, REQ-HOOK-032

#### Task 2.3: PostToolUse Handler (`post_tool.go`)

**Priority**: Medium

**Description**: 도구 실행 결과 수집, 메트릭 기록, Statusline 데이터 업데이트.

**Implementation Details**:
- `Handle(ctx, input)`: `tool_output` 분석, 메트릭 수집
- 실행 시간, 결과 크기, 상태 기록
- Statusline 업데이트를 위한 데이터 구성
- 항상 `Decision: "allow"` 반환 (관찰 전용)

**Testing**: 정상 도구 출력 처리, 빈 tool_output 처리, 대용량 출력 처리

**Covered Requirements**: REQ-HOOK-033

#### Task 2.4: SessionEnd Handler (`session_end.go`)

**Priority**: Medium

**Description**: 메트릭 영속화, 리소스 정리, 랭킹 제출.

**Implementation Details**:
- `Handle(ctx, input)`: 세션 메트릭 집계 및 파일 저장
- 임시 리소스 정리
- 랭킹 활성화 설정 확인 후 선택적 메트릭 제출
- 항상 `Decision: "allow"` 반환

**Testing**: 정상 세션 종료, 메트릭 저장 실패 시 graceful degradation, 랭킹 비활성화 시

**Covered Requirements**: REQ-HOOK-034

#### Task 2.5: Stop Handler (`stop.go`)

**Priority**: Medium

**Description**: 정상 종료, 상태 저장, 루프 컨트롤러 보존.

**Implementation Details**:
- `Handle(ctx, input)`: 현재 작업 상태 저장
- 루프 컨트롤러(Ralph) 상태 직렬화 및 파일 저장
- 진행 중인 작업의 안전한 중단
- 항상 `Decision: "allow"` 반환

**Testing**: 정상 중지, 루프 상태 없을 때 처리, 저장 경로 접근 불가 시

**Covered Requirements**: REQ-HOOK-035

#### Task 2.6: PreCompact Handler (`compact.go`)

**Priority**: Medium

**Description**: 컨텍스트 보존, 세션 스냅샷.

**Implementation Details**:
- `Handle(ctx, input)`: 현재 컨텍스트 캡처
- 세션 상태 스냅샷 생성 및 `.moai/memory/` 저장
- 복구 가능한 형태로 데이터 직렬화
- 항상 `Decision: "allow"` 반환

**Testing**: 정상 컨텍스트 보존, 저장 경로 없을 때 자동 생성, 대용량 컨텍스트 처리

**Covered Requirements**: REQ-HOOK-036

---

### Milestone 3: Integration and Contract Tests (Final Goal)

CLI 통합, 엔드-투-엔드 테스트, 크로스 플랫폼 검증.

#### Task 3.1: CLI Hook Dispatcher (`internal/cli/hook.go`)

**Priority**: High

**Description**: `moai hook <event>` Cobra 서브커맨드 구현.

**Implementation Details**:
- Cobra 커맨드 트리: `moai hook session-start`, `moai hook pre-tool`, 등
- stdin에서 Protocol.ReadInput 호출
- Contract.Validate 실행
- Registry.Dispatch 호출
- Protocol.WriteOutput으로 stdout 출력
- exit code 설정: 0(allow), 2(block), 1(error)

#### Task 3.2: Contract Test Suite

**Priority**: High

**Description**: ADR-012 계약 테스트 전체 스위트.

**Test Cases**:
- Minimal PATH 테스트: PATH=/usr/bin:/bin 환경에서 전체 이벤트 실행
- JSON 왕복(round-trip) 테스트: Marshal -> Unmarshal -> Re-Marshal 동일성
- 비대화형 셸 테스트: SHELL, HOME, USER 환경 변수 없이 실행
- 경로 정규화 테스트: `filepath.Clean()` + 경로 구분자 검증
- 크로스 플랫폼 테스트: darwin, linux, windows CI 매트릭스

#### Task 3.3: Benchmark Test Suite

**Priority**: Medium

**Description**: 성능 기준 충족 검증.

**Benchmarks**:
- `BenchmarkRegistryDispatch`: < 100ms per handler
- `BenchmarkProtocolReadInput`: < 1ms
- `BenchmarkProtocolWriteOutput`: < 1ms
- `BenchmarkContractValidate`: < 1ms
- `BenchmarkEndToEnd`: < 200ms full dispatch

---

## 3. Technology Specifications

### 3.1 Language and Runtime

| Component     | Specification          |
|---------------|------------------------|
| Language      | Go 1.22+               |
| Module        | `github.com/modu-ai/moai-adk-go` |
| Package       | `internal/hook`        |
| Build         | `CGO_ENABLED=0`        |

### 3.2 Standard Library Dependencies

| Package           | Purpose                                |
|-------------------|----------------------------------------|
| `context`         | Cancellation, timeouts                  |
| `encoding/json`   | JSON serialization/deserialization      |
| `errors`          | Sentinel errors, error wrapping         |
| `fmt`             | Error context wrapping (`%w`)           |
| `io`              | Reader/Writer interfaces                |
| `log/slog`        | Structured logging (stderr)             |
| `os`              | stdin, stdout, stderr, exit             |
| `time`            | Timeout duration                        |

### 3.3 Internal Dependencies

| Package                       | Interface Used        | Purpose                        |
|-------------------------------|----------------------|--------------------------------|
| `internal/config`             | `ConfigManager`      | 설정 로드, 타임아웃 값 조회     |
| `internal/core/quality`       | `Gate`               | Quality gate 검증 (PostTool)   |
| `internal/lsp`                | `ServerManager`      | LSP 진단 수집 (PostTool)       |

### 3.4 Testing Dependencies

| Package                        | Purpose                         |
|--------------------------------|---------------------------------|
| `testing` (stdlib)             | 단위 테스트 프레임워크             |
| `github.com/stretchr/testify`  | Assertion helpers, require       |
| `os/exec` (stdlib)             | Contract test binary execution   |

---

## 4. Risk Analysis

### 4.1 Technical Risks

| Risk                                    | Severity | Likelihood | Mitigation                                                |
|-----------------------------------------|----------|------------|-----------------------------------------------------------|
| **stdin/stdout 동시 사용 충돌**           | High     | Medium     | stdout은 JSON 전용, 모든 로그는 stderr로 분리             |
| **플랫폼별 경로 구분자 차이**             | High     | High       | `filepath.Clean()`, `filepath.Join()` 일관 사용           |
| **Claude Code 프로토콜 변경**             | Medium   | Low        | Protocol 인터페이스 추상화로 변경 영향 범위 제한           |
| **타임아웃 내 핸들러 미완료**             | Medium   | Medium     | `context.WithTimeout` + graceful degradation               |
| **ConfigManager 미사용 가능**             | High     | Low        | SPEC-CONFIG-001 선행 구현 보장, mock fallback 준비         |
| **settings.json 형식 불일치**             | High     | Low        | `json.MarshalIndent()` + `json.Valid()` 검증 (ADR-011)    |
| **Windows cmd.exe 실행 환경**             | Medium   | Medium     | `exec.Command` 기반 contract test, CI Windows 매트릭스    |

### 4.2 Process Risks

| Risk                                    | Severity | Likelihood | Mitigation                                                |
|-----------------------------------------|----------|------------|-----------------------------------------------------------|
| **Python 훅 기능 누락**                  | Medium   | Medium     | Python 훅별 기능 체크리스트 작성 및 1:1 매핑 검증          |
| **회귀 테스트 부족**                      | High     | Low        | Contract test suite + CI 6 플랫폼 매트릭스                |
| **성능 기준 미충족**                      | Low      | Low        | Benchmark test suite, compiled binary 특성상 위험 낮음     |

---

## 5. Migration Plan (Python -> Go)

### 5.1 Migration Strategy

**단계적 교체(Phased Replacement)**: Python 훅과 Go 훅을 단계적으로 교체한다. settings.json의 hook command를 Python 스크립트 경로에서 `moai hook <event>`로 변경하면 즉시 전환된다.

### 5.2 Migration Steps

| Step | Action                                                | Verification                              |
|------|-------------------------------------------------------|-------------------------------------------|
| 1    | Go hook 모듈 구현 및 단위 테스트 통과                    | `go test ./internal/hook/...`             |
| 2    | Contract test suite 전체 통과                           | CI 6 플랫폼 매트릭스 green                |
| 3    | settings.json 생성기에서 hook command를 Go 바이너리로 변경 | `moai hook session-start` 수동 실행 확인   |
| 4    | Python 훅 스크립트 기능 매핑 검증                        | 기능 체크리스트 100% 통과                  |
| 5    | 통합 테스트: Claude Code 환경에서 전체 이벤트 사이클 실행  | 전체 세션 라이프사이클 정상 동작           |
| 6    | Python 훅 스크립트 제거                                  | `.claude/hooks/` 디렉터리 삭제             |

### 5.3 Rollback Plan

Go 훅 실행 실패 시 rollback 절차:

1. settings.json에서 hook command를 Python 스크립트 경로로 복원
2. Python 훅 스크립트 재배포 (`templates/.claude/hooks/`)
3. 실패 원인 분석 및 Go 훅 수정 후 재시도

### 5.4 Python Hook Coverage Checklist

| Python Script                           | Go Handler         | Status  |
|-----------------------------------------|--------------------|---------|
| `session_start__show_project_info.py`   | `session_start.go` | Planned |
| `pre_tool__security_guard.py`           | `pre_tool.go`      | Planned |
| `post_tool__linter.py`                  | `post_tool.go`     | Planned |
| `post_tool__code_formatter.py`          | `post_tool.go`     | Planned |
| `post_tool__lsp_diagnostic.py`          | `post_tool.go`     | Planned |
| `session_end__auto_cleanup.py`          | `session_end.go`   | Planned |
| `session_end__rank_submit.py`           | `session_end.go`   | Planned |
| `stop__loop_controller.py`             | `stop.go`          | Planned |
| `pre_compact__save_context.py`          | `compact.go`       | Planned |
| 기타 37개 보조 스크립트                   | (통합)             | Planned |

---

## 6. Architecture Design Direction

### 6.1 Package Structure

```
internal/hook/
    registry.go          # Handler registration and event dispatch
    registry_test.go
    protocol.go          # Claude Code JSON stdin/stdout protocol
    protocol_test.go
    contract.go          # Hook Execution Contract (ADR-012)
    contract_test.go
    session_start.go     # SessionStart event handler
    session_start_test.go
    pre_tool.go          # PreToolUse event handler
    pre_tool_test.go
    post_tool.go         # PostToolUse event handler
    post_tool_test.go
    session_end.go       # SessionEnd event handler
    session_end_test.go
    stop.go              # Stop event handler
    stop_test.go
    compact.go           # PreCompact event handler
    compact_test.go
    errors.go            # Sentinel errors
    testdata/            # Test fixtures (JSON payloads)
```

### 6.2 Dependency Flow

```
internal/cli/hook.go
    |
    v
internal/hook/protocol.go    -- Parse JSON from stdin
    |
    v
internal/hook/contract.go    -- Validate execution contract
    |
    v
internal/hook/registry.go    -- Dispatch to handlers
    |
    +-> session_start.go -> internal/config/
    +-> pre_tool.go      -> internal/config/ (security policies)
    +-> post_tool.go     -> internal/config/, internal/lsp/
    +-> session_end.go   -> internal/config/, internal/rank/
    +-> stop.go          -> internal/loop/
    +-> compact.go       -> (standalone)
    |
    v
internal/hook/protocol.go    -- Write JSON to stdout
```

### 6.3 Constructor Pattern

```go
// NewRegistry creates a hook registry with all dependencies.
func NewRegistry(
    cfg config.ConfigManager,
    quality quality.Gate,
    lspMgr lsp.ServerManager,
) Registry {
    r := &registry{
        cfg:      cfg,
        quality:  quality,
        lspMgr:   lspMgr,
        handlers: make(map[EventType][]Handler),
    }
    // Register all built-in handlers
    r.Register(NewSessionStartHandler(cfg))
    r.Register(NewPreToolHandler(cfg))
    r.Register(NewPostToolHandler(cfg, lspMgr))
    r.Register(NewSessionEndHandler(cfg))
    r.Register(NewStopHandler(cfg))
    r.Register(NewCompactHandler())
    return r
}
```

---

## 7. Quality Criteria

### 7.1 Coverage Target

| Scope                    | Target | Rationale                              |
|--------------------------|--------|----------------------------------------|
| `internal/hook/` 전체    | 95%    | 핵심 인프라, 계약 기반 테스트 필수       |
| Registry                 | 95%    | 디스패치 로직 완전 검증                  |
| Protocol                 | 95%    | 입출력 경계 완전 검증                    |
| Contract                 | 100%   | 계약 위반 탐지 완전성 보장               |
| Event Handlers           | 90%    | 핵심 경로 + 오류 경로 검증               |

### 7.2 TRUST 5 Compliance

| Principle   | Hook Module Application                                      |
|-------------|--------------------------------------------------------------|
| Tested      | 95%+ coverage, contract tests, benchmark tests, fuzz tests   |
| Readable    | Go naming conventions, godoc comments, slog structured logging |
| Unified     | gofumpt formatting, golangci-lint compliance                  |
| Secured     | JSON injection prevention (ADR-011), input validation         |
| Trackable   | Conventional commits, SPEC-HOOK-001 reference in all commits |
