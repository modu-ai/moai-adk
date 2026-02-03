---
id: SPEC-HOOK-001
title: Compiled Hook System
version: "1.0.0"
status: Completed
created: 2026-02-03
updated: 2026-02-03
author: GOOS
priority: P0 Critical
phase: "Phase 1 - Foundation"
module: "internal/hook/"
dependencies:
  - SPEC-CONFIG-001
adr_references:
  - ADR-006 (Hooks as Binary Subcommands)
  - ADR-012 (Hook Execution Contract)
resolves_issues: 28
lifecycle: spec-anchored
tags: "hook, claude-code, binary-subcommand, event-system, P0"
---

# SPEC-HOOK-001: Compiled Hook System

## HISTORY

| Version | Date       | Author | Description                            |
|---------|------------|--------|----------------------------------------|
| 1.0.0   | 2026-02-03 | GOOS   | Initial SPEC creation                  |

---

## 1. Environment (E)

### 1.1 Project Context

MoAI-ADK Go Edition은 기존 Python 기반 MoAI-ADK(~73,000 LOC, 220+ files)를 Go 언어로 완전 재작성하는 프로젝트이다. 이 SPEC은 Python 기반 46개 훅 스크립트(21,535 LOC)를 단일 컴파일 바이너리의 서브커맨드(~1,500 LOC)로 교체하는 Compiled Hook System을 정의한다.

### 1.2 Problem Statement

Python 기반 훅 시스템은 5개월 동안 41회 이상의 PATH 관련 커밋, 4회의 회귀 사이클, 7가지 상이한 접근 방식 시도를 유발하였다. 근본 원인은 다음과 같다:

- **Python 런타임 의존성**: PyYAML import 오류, 가상 환경 미설정 (#278)
- **PATH 해석 실패**: `$CLAUDE_PROJECT_DIR` 미확장, 혼합 경로 구분자 (#259, #265)
- **플랫폼 비호환**: SIGALRM 미지원(Windows), cp949 인코딩 오류 (#129, #269)
- **훅 포맷 오류**: settings.json 내 템플릿 변수(`${SHELL}`, `{{VAR}}`) 미확장 (#288)
- **Claude Code 비대화형 셸 환경**: .bashrc/.zshrc 미로드, PATH 미상속 (#5202, #7490)

### 1.3 Target Module

- **경로**: `internal/hook/`
- **파일 구성**: `registry.go`, `protocol.go`, `contract.go`, `session_start.go`, `pre_tool.go`, `post_tool.go`, `session_end.go`, `stop.go`, `compact.go`
- **예상 LOC**: ~1,500
- **해결 이슈**: 28개 (#129, #259, #265, #269, #278, #288 외 22개)

### 1.4 Dependencies

| Dependency       | Type     | Description                                    |
|------------------|----------|------------------------------------------------|
| SPEC-CONFIG-001  | Internal | Configuration Manager (`internal/config/`)     |
| Claude Code      | External | Hook event protocol (stdin JSON + exit codes)  |
| Go 1.22+         | Runtime  | context, encoding/json, io standard packages   |

### 1.5 Architecture Reference

- **ADR-006**: Hooks as Binary Subcommands -- 훅을 외부 스크립트가 아닌 `moai hook <event>` 컴파일 서브커맨드로 구현
- **ADR-012**: Hook Execution Contract -- 실행 환경 보증/비보증 사항의 공식 명세

---

## 2. Assumptions (A)

### 2.1 Technical Assumptions

| ID    | Assumption                                                    | Confidence | Risk if Wrong                          |
|-------|---------------------------------------------------------------|------------|----------------------------------------|
| A-001 | Claude Code는 hooks를 `type: command`로 subprocess 실행한다    | High       | 전체 훅 아키텍처 재설계 필요             |
| A-002 | stdin으로 유효한 JSON 페이로드가 전달된다                        | High       | 입력 검증 실패로 훅 동작 불가            |
| A-003 | exit code 0은 allow, 2는 block으로 해석된다                     | High       | Claude Code와의 프로토콜 불일치          |
| A-004 | `moai` 바이너리가 시스템 PATH에 존재한다                        | Medium     | 바이너리 발견 불가로 훅 실행 실패         |
| A-005 | 훅 실행 환경에 .bashrc/.zshrc가 로드되지 않는다                  | High       | 환경 변수 의존 코드의 예기치 않은 동작    |
| A-006 | SPEC-CONFIG-001이 선행 구현되어 ConfigManager를 사용 가능하다    | High       | 설정 로드 불가로 훅 기능 제한             |

### 2.2 Business Assumptions

| ID    | Assumption                                                    | Confidence | Risk if Wrong                          |
|-------|---------------------------------------------------------------|------------|----------------------------------------|
| B-001 | 6개 이벤트 타입이 현재 Claude Code 훅 프로토콜의 전체 집합이다   | Medium     | 누락 이벤트에 대한 추가 핸들러 필요       |
| B-002 | Python 훅의 모든 기능이 Go 핸들러로 1:1 매핑 가능하다            | High       | 기능 손실 또는 우회 구현 필요             |

---

## 3. Requirements (R)

### Module 1: Registry (핸들러 등록 및 디스패치)

**REQ-HOOK-001** [Ubiquitous]
시스템은 **항상** 타입 안전한 핸들러 레지스트리를 통해 이벤트별 핸들러를 관리해야 한다.

- `Register(handler Handler)` 메서드로 핸들러 등록
- `Handlers(event EventType) []Handler`로 이벤트별 핸들러 조회
- `Dispatch(ctx, event, input) (*HookOutput, error)`로 이벤트 디스패치

**REQ-HOOK-002** [Event-Driven]
**WHEN** Claude Code가 훅 이벤트를 발생시키면 **THEN** Registry는 해당 EventType에 등록된 모든 핸들러를 순차 실행하고 결과를 반환해야 한다.

**REQ-HOOK-003** [Event-Driven]
**WHEN** 핸들러가 `Decision: "block"`을 반환하면 **THEN** Registry는 남은 핸들러 실행을 중단하고 block 결과를 즉시 반환해야 한다.

**REQ-HOOK-004** [State-Driven]
**IF** 모든 핸들러가 성공적으로 완료되면 **THEN** Registry는 `Decision: "allow"` 결과를 반환해야 한다.

### Module 2: Protocol (Claude Code JSON stdin/stdout 통신)

**REQ-HOOK-010** [Ubiquitous]
시스템은 **항상** Claude Code의 JSON stdin/stdout 프로토콜을 준수해야 한다.

- `ReadInput(r io.Reader) (*HookInput, error)`: stdin에서 JSON 파싱
- `WriteOutput(w io.Writer, output *HookOutput) error`: stdout으로 JSON 직렬화

**REQ-HOOK-011** [Event-Driven]
**WHEN** stdin으로 JSON 페이로드가 수신되면 **THEN** Protocol은 `HookInput` 구조체로 역직렬화하고 필수 필드(`session_id`, `cwd`, `hook_event_name`)를 검증해야 한다.

**REQ-HOOK-012** [Unwanted]
시스템은 문자열 연결(string concatenation)로 JSON을 생성**하지 않아야 한다**. 모든 JSON 출력은 `json.Marshal()`을 통한 Go 구조체 직렬화만 허용한다. (ADR-011)

**REQ-HOOK-013** [Unwanted]
시스템은 stdout에 JSON 외의 출력(로그, 디버그 메시지 등)을 기록**하지 않아야 한다**. 진단 출력은 반드시 stderr로 전송한다.

### Module 3: Contract (Hook Execution Contract, ADR-012)

**REQ-HOOK-020** [Ubiquitous]
시스템은 **항상** Hook Execution Contract(ADR-012)를 준수하며 실행 환경을 검증해야 한다.

**보증 사항(Guarantees)**:
- stdin: Claude Code 훅 프로토콜을 준수하는 유효한 JSON
- exit code: 0(allow/success), 2(block), 기타(non-blocking error)
- timeout: `context.WithTimeout`을 통한 구성 가능한 타임아웃(기본 30초)
- config access: CLI 커맨드와 동일한 바이너리, 동일한 설정 로더
- working directory: 프로젝트 루트(`$CLAUDE_PROJECT_DIR`)

**비보증 사항(Non-Guarantees)**:
- 사용자 PATH(시스템 PATH에 `moai`가 존재해야 함)
- 셸 환경 변수(.bashrc/.zshrc 미로드)
- 셸 함수 또는 별칭(alias)
- Python/Node.js/uv 가용성

**REQ-HOOK-021** [Event-Driven]
**WHEN** Contract 검증이 실패하면 **THEN** 시스템은 `ErrHookContractFail` 오류를 반환하고 실패 원인을 stderr에 로그로 기록해야 한다.

**REQ-HOOK-022** [State-Driven]
**IF** 훅 실행이 설정된 타임아웃을 초과하면 **THEN** `context.DeadlineExceeded`를 감지하여 `ErrHookTimeout` 오류를 반환해야 한다.

### Module 4: Event Handlers (6개 이벤트 핸들러)

#### 4.1 SessionStart Handler

**REQ-HOOK-030** [Event-Driven]
**WHEN** `SessionStart` 이벤트가 발생하면 **THEN** 시스템은 다음을 수행해야 한다:
1. 세션 초기화 및 세션 ID 기록
2. 프로젝트 설정(`ConfigManager`) 로드 및 검증
3. 실행 환경(작업 디렉터리, 필수 경로) 유효성 확인
4. 프로젝트 정보를 Data 필드에 포함하여 반환

#### 4.2 PreToolUse Handler

**REQ-HOOK-031** [Event-Driven]
**WHEN** `PreToolUse` 이벤트가 발생하면 **THEN** 시스템은 다음을 수행해야 한다:
1. `tool_name`과 `tool_input`을 기반으로 도구 사용 권한 검증
2. 보안 정책(허용/차단 도구 목록)에 따른 접근 제어
3. 차단 시 `Decision: "block"`과 `Reason` 반환
4. 허용 시 `Decision: "allow"` 반환

**REQ-HOOK-032** [Unwanted]
시스템은 보안 정책에 의해 차단된 도구 실행을 허용**하지 않아야 한다**. 차단된 도구에 대해 반드시 exit code 2를 반환한다.

#### 4.3 PostToolUse Handler

**REQ-HOOK-033** [Event-Driven]
**WHEN** `PostToolUse` 이벤트가 발생하면 **THEN** 시스템은 다음을 수행해야 한다:
1. 도구 실행 결과(`tool_output`) 수집 및 분석
2. 실행 메트릭(수행 시간, 결과 상태) 기록
3. Statusline 데이터 업데이트를 위한 정보 수집
4. 수집된 메트릭을 Data 필드에 포함하여 반환

#### 4.4 SessionEnd Handler

**REQ-HOOK-034** [Event-Driven]
**WHEN** `SessionEnd` 이벤트가 발생하면 **THEN** 시스템은 다음을 수행해야 한다:
1. 세션 동안 수집된 메트릭 영속화(persistence)
2. 임시 리소스 정리(cleanup)
3. 랭킹 시스템으로 세션 메트릭 제출(설정된 경우)
4. 정리 결과를 Data 필드에 포함하여 반환

#### 4.5 Stop Handler

**REQ-HOOK-035** [Event-Driven]
**WHEN** `Stop` 이벤트가 발생하면 **THEN** 시스템은 다음을 수행해야 한다:
1. 현재 실행 상태의 정상 종료(graceful shutdown)
2. 진행 중인 작업의 상태 저장
3. 루프 컨트롤러(Ralph) 상태 보존
4. 저장 결과를 Data 필드에 포함하여 반환

#### 4.6 PreCompact Handler

**REQ-HOOK-036** [Event-Driven]
**WHEN** `PreCompact` 이벤트가 발생하면 **THEN** 시스템은 다음을 수행해야 한다:
1. 현재 컨텍스트 정보 캡처 및 저장
2. 세션 상태 스냅샷 생성
3. 컴팩션 이후 복구에 필요한 데이터 보존
4. 보존 결과를 Data 필드에 포함하여 반환

### Module 5: Cross-Cutting Concerns (횡단 관심사)

**REQ-HOOK-040** [Ubiquitous]
시스템은 **항상** `log/slog`를 통한 구조화된 로깅을 수행해야 한다. 로그 출력은 반드시 stderr로 전송하며 stdout은 JSON 응답 전용이다.

**REQ-HOOK-041** [Ubiquitous]
시스템은 **항상** `context.Context`를 첫 번째 매개변수로 받아 취소(cancellation) 및 타임아웃을 지원해야 한다.

**REQ-HOOK-042** [State-Driven]
**IF** 핸들러에서 복구 가능한 오류가 발생하면 **THEN** `fmt.Errorf("context: %w", err)` 패턴으로 오류를 래핑하여 반환해야 한다.

**REQ-HOOK-043** [Optional]
**가능하면** 핸들러 실행 메트릭(실행 시간, 성공/실패 횟수)을 수집하여 성능 모니터링을 제공한다.

---

## 4. Specifications (S)

### 4.1 Interface Definitions

```go
type Handler interface {
    Handle(ctx context.Context, input *HookInput) (*HookOutput, error)
    EventType() EventType
}

type Registry interface {
    Register(handler Handler)
    Dispatch(ctx context.Context, event EventType, input *HookInput) (*HookOutput, error)
    Handlers(event EventType) []Handler
}

type Protocol interface {
    ReadInput(r io.Reader) (*HookInput, error)
    WriteOutput(w io.Writer, output *HookOutput) error
}

type Contract interface {
    Validate(ctx context.Context) error
    Guarantees() []string
    NonGuarantees() []string
}
```

### 4.2 Data Structures

```go
type EventType string

const (
    EventSessionStart EventType = "SessionStart"
    EventPreToolUse   EventType = "PreToolUse"
    EventPostToolUse  EventType = "PostToolUse"
    EventSessionEnd   EventType = "SessionEnd"
    EventStop         EventType = "Stop"
    EventPreCompact   EventType = "PreCompact"
)

type HookInput struct {
    SessionID     string          `json:"session_id"`
    CWD           string          `json:"cwd"`
    HookEventName string          `json:"hook_event_name"`
    ToolName      string          `json:"tool_name,omitempty"`
    ToolInput     json.RawMessage `json:"tool_input,omitempty"`
    ToolOutput    json.RawMessage `json:"tool_output,omitempty"`
    ProjectDir    string          `json:"project_dir"`
}

type HookOutput struct {
    Decision string          `json:"decision,omitempty"`
    Reason   string          `json:"reason,omitempty"`
    Data     json.RawMessage `json:"data,omitempty"`
}
```

### 4.3 Sentinel Errors

```go
var (
    ErrHookTimeout      = errors.New("hook: execution timed out")
    ErrHookContractFail = errors.New("hook: execution contract violated")
    ErrHookInvalidInput = errors.New("hook: invalid JSON input")
    ErrHookBlocked      = errors.New("hook: action blocked by hook")
)
```

### 4.4 settings.json Integration

Claude Code `settings.json`에서의 훅 등록 형태:

```json
{
  "hooks": {
    "SessionStart": [
      { "hooks": [{ "type": "command", "command": "moai hook session-start" }] }
    ],
    "PreToolUse": [
      { "matcher": "Write|Edit|Bash", "hooks": [{ "type": "command", "command": "moai hook pre-tool" }] }
    ],
    "PostToolUse": [
      { "matcher": "Write|Edit", "hooks": [{ "type": "command", "command": "moai hook post-tool" }] }
    ],
    "SessionEnd": [
      { "hooks": [{ "type": "command", "command": "moai hook session-end" }] }
    ],
    "Stop": [
      { "hooks": [{ "type": "command", "command": "moai hook stop" }] }
    ],
    "PreCompact": [
      { "hooks": [{ "type": "command", "command": "moai hook compact" }] }
    ]
  }
}
```

### 4.5 Exit Code Semantics

| Exit Code | Meaning          | Claude Code Behavior                |
|-----------|------------------|-------------------------------------|
| 0         | Allow / Success  | 도구 실행 허용, 정상 진행            |
| 2         | Block            | 도구 실행 차단, Reason 표시          |
| Other     | Non-blocking Error | 오류 로깅 후 정상 진행 (비차단)     |

### 4.6 File-to-Python Replacement Mapping

| Go File              | Purpose                    | Replaces (Python)                                              |
|----------------------|----------------------------|----------------------------------------------------------------|
| `registry.go`        | 핸들러 등록 및 디스패치       | `jit_enhanced_hook_manager.py` (1,988 LOC)                     |
| `protocol.go`        | JSON stdin/stdout 프로토콜   | `lib/models.py`, `lib/common.py`                               |
| `contract.go`        | 실행 환경 계약 검증           | -- (신규: 회귀 방지)                                            |
| `session_start.go`   | 프로젝트 정보, 설정 검증      | `session_start__show_project_info.py`                          |
| `pre_tool.go`        | 보안 가드, 입력 검증          | `pre_tool__security_guard.py`                                  |
| `post_tool.go`       | 린터, 포매터, LSP 진단        | `post_tool__linter.py`, `post_tool__code_formatter.py`, `post_tool__lsp_diagnostic.py` |
| `session_end.go`     | 정리, 랭크 제출              | `session_end__auto_cleanup.py`, `session_end__rank_submit.py`  |
| `stop.go`            | 루프 컨트롤러                | `stop__loop_controller.py`                                     |
| `compact.go`         | 컨텍스트 보존                | `pre_compact__save_context.py`                                 |

### 4.7 Performance Requirements

| Metric                        | Target    | Measurement Method                     |
|-------------------------------|-----------|----------------------------------------|
| 단일 핸들러 실행 시간           | < 100ms   | Benchmark test (`go test -bench`)       |
| 전체 디스패치 (모든 핸들러)     | < 200ms   | End-to-end benchmark                    |
| JSON 파싱 (stdin)              | < 1ms     | Benchmark test                          |
| JSON 직렬화 (stdout)           | < 1ms     | Benchmark test                          |
| Contract 검증                  | < 1ms     | Benchmark test                          |
| 메모리 사용량 (hook 실행 중)    | < 10MB    | Runtime profiling                       |

---

## 5. Traceability

### 5.1 Requirements to Files

| Requirement      | Implementation File            |
|------------------|-------------------------------|
| REQ-HOOK-001~004 | `registry.go`                 |
| REQ-HOOK-010~013 | `protocol.go`                 |
| REQ-HOOK-020~022 | `contract.go`                 |
| REQ-HOOK-030     | `session_start.go`            |
| REQ-HOOK-031~032 | `pre_tool.go`                 |
| REQ-HOOK-033     | `post_tool.go`                |
| REQ-HOOK-034     | `session_end.go`              |
| REQ-HOOK-035     | `stop.go`                     |
| REQ-HOOK-036     | `compact.go`                  |
| REQ-HOOK-040~043 | All files (cross-cutting)     |

### 5.2 ADR References

| ADR     | Related Requirements       | Impact                                      |
|---------|---------------------------|---------------------------------------------|
| ADR-006 | All REQ-HOOK-*             | 훅을 바이너리 서브커맨드로 구현               |
| ADR-011 | REQ-HOOK-012               | JSON 출력 시 구조체 직렬화만 허용             |
| ADR-012 | REQ-HOOK-020~022           | 실행 환경 계약 공식 명세                      |

### 5.3 Issue Resolution Mapping

| Issue Category              | Count | Representative Issues        |
|-----------------------------|-------|-------------------------------|
| Python 런타임 의존성          | 12    | #278, #288                    |
| PATH 해석 실패               | 8     | #259, #265                    |
| 플랫폼 비호환                | 5     | #129, #269                    |
| settings.json 생성 오류      | 3     | #288                          |
| **Total**                   | **28**|                               |

---

## Implementation Notes

**Status**: Completed
**Implementation Date**: 2026-02-03
**Development Methodology**: Hybrid (TDD for new code, DDD for existing code)
**Test Coverage**: 90.0%

### Summary

Hook system implemented as a pure Go replacement for the Python hook infrastructure. Includes JSON stdin/stdout protocol for Claude Code integration, handler registry with event-type dispatching, and execution contract validation. Covers all six hook points: SessionStart, PreToolUse, PostToolUse, SessionEnd, Stop, and PreCompact. Eliminates Python runtime dependency and resolves 28 known issues from the Python implementation.

### Files Created

- `internal/hook/compact.go`
- `internal/hook/compact_test.go`
- `internal/hook/contract.go`
- `internal/hook/contract_test.go`
- `internal/hook/doc.go`
- `internal/hook/errors.go`
- `internal/hook/post_tool.go`
- `internal/hook/post_tool_test.go`
- `internal/hook/pre_tool.go`
- `internal/hook/pre_tool_test.go`
- `internal/hook/protocol.go`
- `internal/hook/protocol_test.go`
- `internal/hook/registry.go`
- `internal/hook/registry_test.go`
- `internal/hook/session_end.go`
- `internal/hook/session_end_test.go`
- `internal/hook/session_start.go`
- `internal/hook/session_start_test.go`
- `internal/hook/stop.go`
- `internal/hook/stop_test.go`
- `internal/hook/types.go`
- `internal/hook/types_test.go`
