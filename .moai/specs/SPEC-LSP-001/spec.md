---
id: SPEC-LSP-001
title: Language Server Protocol Client System
phase: "Phase 2 - Core Domains"
status: Completed
priority: High
created: 2026-02-03
module: "internal/lsp/"
files:
  - client.go
  - server.go
  - protocol.go
  - models.go
estimated_loc: 1500
dependencies:
  - SPEC-CONFIG-001
lifecycle: spec-anchored
tags: lsp, json-rpc, diagnostics, multi-language, concurrency
---

# SPEC-LSP-001: Language Server Protocol Client System

## HISTORY

| 날짜 | 버전 | 변경 내용 |
|------|------|-----------|
| 2026-02-03 | 1.0.0 | 초기 SPEC 작성 |

---

## 1. Environment (환경)

### 1.1 프로젝트 컨텍스트

MoAI-ADK (Go Edition)는 Python 기반 MoAI-ADK(~73,000 LOC)를 Go로 완전 재작성하는 프로젝트이다. LSP 모듈은 16개 이상의 프로그래밍 언어에 대한 코드 진단(Diagnostic) 수집, 심볼 조회, 참조 검색 등의 기능을 제공하며, TRUST 5 Quality Gate 시스템의 핵심 데이터 소스 역할을 한다.

### 1.2 시스템 경계

- **입력**: Language Server로부터의 JSON-RPC 2.0 응답 (stdio/TCP transport)
- **출력**: 구조화된 Diagnostic, Location, HoverResult, DocumentSymbol 데이터
- **의존 모듈**: `internal/config/` (SPEC-CONFIG-001), `pkg/models/`, `pkg/utils/`
- **소비자 모듈**: `internal/core/quality/`, `internal/hook/`, `internal/cli/`

### 1.3 기술 스택

| 구성 요소 | 패키지 | 버전 | 용도 |
|-----------|--------|------|------|
| LSP Type 정의 | `go.lsp.dev/protocol` | v0.12+ | LSP 프로토콜 타입 (Diagnostic, Position, TextDocument) |
| JSON-RPC Transport | `go.lsp.dev/jsonrpc2` | v0.10+ | JSON-RPC 2.0 통신 계층 |
| 동시성 관리 | `golang.org/x/sync/errgroup` | latest | 병렬 서버 시작 및 진단 수집 |
| Context | `context` (stdlib) | Go 1.22+ | 취소, 타임아웃, request-scoped 값 |
| 동기화 | `sync` (stdlib) | Go 1.22+ | RWMutex를 통한 동시 접근 제어 |

### 1.4 지원 언어 (16+)

Go, Python, TypeScript, Java, Rust, C/C++, Ruby, PHP, Kotlin, Swift, Dart, Elixir, Scala, Haskell, Zig 및 추가 언어

### 1.5 성능 예산 (Performance Budget)

| 지표 | 목표 | 측정 방법 |
|------|------|-----------|
| 단일 서버 시작 | < 500ms | Integration test |
| 16개 서버 진단 수집 | < 2s | Parallel benchmark |
| Idle 메모리 사용량 | < 20MB | Runtime profiling |
| Peak 메모리 (16개 서버) | < 200MB | Load testing |

---

## 2. Assumptions (가정)

### 2.1 외부 의존성 가정

- **A-1**: Language Server 바이너리는 사용자 시스템에 사전 설치되어 있으며, 시스템 PATH를 통해 접근 가능하다.
- **A-2**: 각 Language Server는 LSP 3.17 사양을 준수하며, `initialize`, `textDocument/publishDiagnostics`, `textDocument/hover`, `textDocument/definition`, `textDocument/references`, `textDocument/documentSymbol` capability를 지원한다.
- **A-3**: Language Server는 stdio 또는 TCP transport 중 하나 이상을 지원한다.

### 2.2 내부 의존성 가정

- **A-4**: `internal/config/` (SPEC-CONFIG-001)가 구현 완료되어 LSP 서버 설정(서버 경로, 인자, transport 타입)을 제공할 수 있다.
- **A-5**: `internal/foundation/langs.go`에서 `LanguageDef` 구조체를 통해 각 언어의 LSP 서버 실행 경로(`LSPServer`), 인자(`LSPArgs`), 파일 패턴(`FilePatterns`)을 정의한다.

### 2.3 환경 가정

- **A-6**: Go 1.22 이상 환경에서 빌드 및 실행된다.
- **A-7**: CGO_ENABLED=0으로 빌드되며, 순수 Go 의존성만 사용한다.
- **A-8**: 동시 LSP 서버 시작은 최대 4개로 제한하여 시스템 리소스를 보호한다.

### 2.4 신뢰도 분류

| 가정 | 신뢰도 | 검증 방법 |
|------|--------|-----------|
| A-1 | Medium | `exec.LookPath()` 로 서버 존재 여부 확인 |
| A-2 | High | LSP 3.17은 업계 표준; 주요 서버 모두 지원 |
| A-3 | High | stdio가 기본; gopls, pyright 등 모두 지원 |
| A-4 | High | SPEC-CONFIG-001 의존성으로 선행 구현 보장 |
| A-5 | High | design.md에 LanguageDef 인터페이스 정의 완료 |
| A-6 | High | go.mod에 Go 버전 명시 |
| A-7 | High | tech.md ADR에 CGO_ENABLED=0 명시 |
| A-8 | Medium | errgroup.SetLimit(4)로 구현; 향후 설정 가능 |

---

## 3. Requirements (요구사항)

### 3.1 Client Interface 요구사항

#### REQ-LSP-C01: 서버 초기화 (Event-Driven)

**WHEN** Client.Initialize(ctx, rootURI)가 호출되면 **THEN** 시스템은 JSON-RPC `initialize` 요청을 Language Server로 전송하고, 서버의 capability 응답을 파싱하여 `initialized` 알림을 전송해야 한다.

#### REQ-LSP-C02: 진단 수집 (Event-Driven)

**WHEN** Client.Diagnostics(ctx, uri)가 호출되면 **THEN** 시스템은 해당 문서 URI에 대한 현재 Diagnostic 목록을 반환해야 한다. Diagnostic은 Range, Severity, Code, Source, Message 필드를 포함해야 한다.

#### REQ-LSP-C03: 참조 검색 (Event-Driven)

**WHEN** Client.References(ctx, uri, pos)가 호출되면 **THEN** 시스템은 지정된 Position의 심볼에 대한 모든 참조 Location 목록을 반환해야 한다.

#### REQ-LSP-C04: Hover 정보 조회 (Event-Driven)

**WHEN** Client.Hover(ctx, uri, pos)가 호출되면 **THEN** 시스템은 해당 Position의 심볼에 대한 HoverResult(Contents, Range)를 반환해야 한다.

#### REQ-LSP-C05: 정의 위치 조회 (Event-Driven)

**WHEN** Client.Definition(ctx, uri, pos)가 호출되면 **THEN** 시스템은 해당 Position 심볼의 정의 Location 목록을 반환해야 한다.

#### REQ-LSP-C06: 문서 심볼 조회 (Event-Driven)

**WHEN** Client.Symbols(ctx, uri)가 호출되면 **THEN** 시스템은 해당 문서의 DocumentSymbol 계층 구조(Name, Kind, Range, Children)를 반환해야 한다.

#### REQ-LSP-C07: 서버 종료 (Event-Driven)

**WHEN** Client.Shutdown(ctx)이 호출되면 **THEN** 시스템은 `shutdown` 요청을 전송하고 서버 응답을 기다린 후, `exit` 알림을 전송하고 연결을 정리해야 한다.

### 3.2 ServerManager Interface 요구사항

#### REQ-LSP-S01: 서버 시작 (Event-Driven)

**WHEN** ServerManager.StartServer(ctx, lang)가 호출되면 **THEN** 시스템은 `LanguageDef`에서 해당 언어의 서버 경로와 인자를 조회하고, 프로세스를 시작하며, JSON-RPC 연결을 수립하고, `initialize` 핸드셰이크를 완료해야 한다.

#### REQ-LSP-S02: 서버 중지 (Event-Driven)

**WHEN** ServerManager.StopServer(ctx, lang)가 호출되면 **THEN** 시스템은 해당 언어 서버에 `shutdown` 요청을 전송하고, `exit` 알림을 전송하며, 프로세스를 정리하고 내부 레지스트리에서 제거해야 한다.

#### REQ-LSP-S03: Client 조회 (Event-Driven)

**WHEN** ServerManager.GetClient(lang)가 호출되면 **THEN** 시스템은 실행 중인 해당 언어의 LSP Client를 반환해야 한다. **IF** 해당 언어 서버가 실행 중이 아니면 **THEN** `ErrServerNotRunning` 에러를 반환해야 한다.

#### REQ-LSP-S04: 활성 서버 목록 (Ubiquitous)

시스템은 **항상** ActiveServers() 호출 시 현재 실행 중인 모든 Language Server의 언어 식별자 목록을 반환해야 한다.

#### REQ-LSP-S05: Health Check (Event-Driven)

**WHEN** ServerManager.HealthCheck(ctx)가 호출되면 **THEN** 시스템은 모든 활성 서버에 대해 건강 상태를 확인하고, 언어 식별자를 키로 하고 에러(또는 nil)를 값으로 하는 맵을 반환해야 한다.

### 3.3 병렬 처리 요구사항

#### REQ-LSP-P01: 병렬 서버 시작 (State-Driven)

**IF** 여러 언어 서버를 동시에 시작해야 하는 상태라면 **THEN** 시스템은 `errgroup`을 사용하여 최대 4개의 동시 서버 시작을 수행하고, 개별 서버 시작 타임아웃을 500ms로 제한해야 한다.

#### REQ-LSP-P02: 병렬 진단 수집 (State-Driven)

**IF** 여러 활성 서버에서 진단을 수집해야 하는 상태라면 **THEN** 시스템은 모든 활성 서버에 병렬로 Diagnostics 요청을 전송하고, 전체 수집 타임아웃을 2초로 제한하며, `sync.Mutex`로 결과 슬라이스를 보호해야 한다.

#### REQ-LSP-P03: 비치명적 에러 처리 (State-Driven)

**IF** 병렬 진단 수집 중 개별 서버에서 에러가 발생하면 **THEN** 시스템은 해당 에러를 로깅하되, 다른 서버의 결과 수집을 중단하지 않아야 한다.

### 3.4 Transport 요구사항

#### REQ-LSP-T01: Stdio Transport (Ubiquitous)

시스템은 **항상** Language Server와의 기본 통신 방식으로 stdio(stdin/stdout) transport를 지원해야 한다.

#### REQ-LSP-T02: TCP Transport (Optional)

**가능하면** Language Server와의 TCP 소켓 기반 transport도 제공해야 한다.

#### REQ-LSP-T03: JSON-RPC 2.0 프로토콜 (Ubiquitous)

시스템은 **항상** LSP 3.17 사양에 따른 JSON-RPC 2.0 프로토콜로 통신해야 한다. 모든 요청은 고유한 request ID를 가져야 하며, 응답은 해당 ID로 매칭되어야 한다.

### 3.5 에러 처리 요구사항

#### REQ-LSP-E01: 서버 미실행 에러 (Unwanted)

시스템은 실행 중이지 않은 서버에 대한 요청 시 `ErrServerNotRunning`을 반환**하지 않아야 한다**... 가 아니라, 반드시 명확한 `ErrServerNotRunning` 에러를 반환해야 한다.

#### REQ-LSP-E02: 서버 시작 실패 (Event-Driven)

**WHEN** Language Server 프로세스 시작에 실패하면 **THEN** 시스템은 `ErrServerStartFailed`를 반환하고, 실패 원인(바이너리 미발견, 권한 부족 등)을 에러 메시지에 래핑해야 한다.

#### REQ-LSP-E03: 초기화 실패 (Event-Driven)

**WHEN** `initialize` 핸드셰이크가 타임아웃되거나 실패하면 **THEN** 시스템은 `ErrInitializeFailed`를 반환하고, 시작된 프로세스를 정리해야 한다.

#### REQ-LSP-E04: 연결 끊김 (Event-Driven)

**WHEN** Language Server 프로세스가 예기치 않게 종료되면 **THEN** 시스템은 `ErrConnectionClosed`를 반환하고, 내부 레지스트리에서 해당 서버를 제거해야 한다.

#### REQ-LSP-E05: Context 취소 (Ubiquitous)

시스템은 **항상** `context.Context`의 취소 및 타임아웃 신호를 존중하여, 진행 중인 JSON-RPC 요청을 정리하고 즉시 반환해야 한다.

### 3.6 Graceful Degradation 요구사항

#### REQ-LSP-G01: 서버 미설치 시 동작 (State-Driven)

**IF** 특정 Language Server가 시스템에 설치되지 않은 상태라면 **THEN** 시스템은 해당 언어 서버 시작을 건너뛰고 warning 로그를 남기며, 다른 언어 서버의 동작에 영향을 주지 않아야 한다.

#### REQ-LSP-G02: 부분 실패 시 동작 (State-Driven)

**IF** 16개 서버 중 일부가 시작에 실패한 상태라면 **THEN** 시스템은 성공한 서버들에 대해서만 정상적으로 진단 수집 및 조회 기능을 제공해야 한다.

---

## 4. Specifications (세부 사양)

### 4.1 파일 구조

```
internal/lsp/
  client.go     -- Client interface 구현, JSON-RPC 요청/응답 처리
  server.go     -- ServerManager 구현, 프로세스 라이프사이클, errgroup 병렬 처리
  protocol.go   -- JSON-RPC 2.0 메시지 인코딩/디코딩, request ID 관리
  models.go     -- Diagnostic, Position, Range, Location, HoverResult, DocumentSymbol 타입 정의
```

### 4.2 인터페이스 정의

```go
// Client -- 단일 Language Server와 통신하는 인터페이스
type Client interface {
    Initialize(ctx context.Context, rootURI string) error
    Diagnostics(ctx context.Context, uri string) ([]Diagnostic, error)
    References(ctx context.Context, uri string, pos Position) ([]Location, error)
    Hover(ctx context.Context, uri string, pos Position) (*HoverResult, error)
    Definition(ctx context.Context, uri string, pos Position) ([]Location, error)
    Symbols(ctx context.Context, uri string) ([]DocumentSymbol, error)
    Shutdown(ctx context.Context) error
}

// ServerManager -- 여러 Language Server의 라이프사이클을 관리하는 인터페이스
type ServerManager interface {
    StartServer(ctx context.Context, lang string) error
    StopServer(ctx context.Context, lang string) error
    GetClient(lang string) (Client, error)
    ActiveServers() []string
    HealthCheck(ctx context.Context) map[string]error
}
```

### 4.3 데이터 모델

```go
type Diagnostic struct {
    Range    Range              `json:"range"`
    Severity DiagnosticSeverity `json:"severity"`
    Code     string             `json:"code,omitempty"`
    Source   string             `json:"source,omitempty"`
    Message  string             `json:"message"`
}

type DiagnosticSeverity int

const (
    SeverityError   DiagnosticSeverity = 1
    SeverityWarning DiagnosticSeverity = 2
    SeverityInfo    DiagnosticSeverity = 3
    SeverityHint    DiagnosticSeverity = 4
)

type Position struct {
    Line      int `json:"line"`
    Character int `json:"character"`
}

type Range struct {
    Start Position `json:"start"`
    End   Position `json:"end"`
}

type Location struct {
    URI   string `json:"uri"`
    Range Range  `json:"range"`
}

type HoverResult struct {
    Contents string `json:"contents"`
    Range    *Range `json:"range,omitempty"`
}

type DocumentSymbol struct {
    Name     string           `json:"name"`
    Kind     int              `json:"kind"`
    Range    Range            `json:"range"`
    Children []DocumentSymbol `json:"children,omitempty"`
}
```

### 4.4 에러 타입

```go
var (
    ErrServerNotRunning  = errors.New("lsp: server not running")
    ErrServerStartFailed = errors.New("lsp: server failed to start")
    ErrInitializeFailed  = errors.New("lsp: initialization failed")
    ErrConnectionClosed  = errors.New("lsp: connection closed")
)
```

### 4.5 동시성 설계

| 연산 | 패턴 | 동시 한도 | 타임아웃 |
|------|------|----------|---------|
| 서버 일괄 시작 | `errgroup.Group` | 4개 동시 | 500ms/서버 |
| 진단 일괄 수집 | `errgroup.Group` + `sync.Mutex` | 제한 없음 | 2s 전체 |
| 서버 레지스트리 접근 | `sync.RWMutex` | N/A | N/A |
| 개별 요청 | `context.WithTimeout` | 1 | 요청별 설정 |

### 4.6 의존성 그래프

```
internal/lsp/ -----> go.lsp.dev/protocol (LSP type 정의)
                  -> go.lsp.dev/jsonrpc2 (JSON-RPC 2.0 transport)
                  -> golang.org/x/sync/errgroup (병렬 처리)
                  -> internal/config/ (서버 설정 조회)
                  -> internal/foundation/ (LanguageDef 조회)
                  -> pkg/models/ (공유 모델)
                  -> pkg/utils/ (로거, 타임아웃 유틸리티)
```

### 4.7 소비자 통합

| 소비자 모듈 | 사용 방식 | 주요 메서드 |
|-------------|-----------|-------------|
| `internal/core/quality/` | TRUST 5 검증 시 진단 수집 | `CollectAllDiagnostics()`, `Diagnostics()` |
| `internal/hook/post_tool.go` | PostToolUse Hook에서 LSP 진단 보고 | `Diagnostics()` |
| `internal/cli/doctor.go` | `moai doctor` 명령에서 LSP 서버 상태 확인 | `HealthCheck()`, `ActiveServers()` |

---

## 5. Constraints (제약 조건)

### 5.1 기술적 제약

- **C-1**: CGO_ENABLED=0 빌드 필수 -- 모든 의존성은 순수 Go여야 한다
- **C-2**: LSP 3.17 사양 준수 필수
- **C-3**: `go.lsp.dev` 패키지 사용 (ADR in tech.md)
- **C-4**: 모든 장기 실행 연산은 첫 번째 매개변수로 `context.Context` 수용
- **C-5**: Error wrapping은 `fmt.Errorf("context: %w", err)` 패턴 사용

### 5.2 호환성 제약

- **C-6**: Python MoAI-ADK와 동일한 16개 이상 언어 지원 필수
- **C-7**: 기존 `.moai/config/sections/quality.yaml`의 `lsp_quality_gates` 설정과 호환

### 5.3 보안 제약

- **C-8**: Language Server 프로세스 경로는 `exec.LookPath()`로 검증 후 실행
- **C-9**: LSP 서버 토큰은 환경 변수를 통해서만 전달 (`os.Getenv()` + 검증)

---

## 6. Traceability (추적성)

| 요구사항 | 파일 | 테스트 |
|----------|------|--------|
| REQ-LSP-C01 | client.go | client_test.go: TestInitialize |
| REQ-LSP-C02 | client.go | client_test.go: TestDiagnostics |
| REQ-LSP-C03 | client.go | client_test.go: TestReferences |
| REQ-LSP-C04 | client.go | client_test.go: TestHover |
| REQ-LSP-C05 | client.go | client_test.go: TestDefinition |
| REQ-LSP-C06 | client.go | client_test.go: TestSymbols |
| REQ-LSP-C07 | client.go | client_test.go: TestShutdown |
| REQ-LSP-S01 | server.go | server_test.go: TestStartServer |
| REQ-LSP-S02 | server.go | server_test.go: TestStopServer |
| REQ-LSP-S03 | server.go | server_test.go: TestGetClient |
| REQ-LSP-S04 | server.go | server_test.go: TestActiveServers |
| REQ-LSP-S05 | server.go | server_test.go: TestHealthCheck |
| REQ-LSP-P01 | server.go | server_test.go: TestParallelStartup |
| REQ-LSP-P02 | server.go | server_test.go: TestParallelDiagnostics |
| REQ-LSP-P03 | server.go | server_test.go: TestNonFatalError |
| REQ-LSP-T01 | protocol.go | protocol_test.go: TestStdioTransport |
| REQ-LSP-T02 | protocol.go | protocol_test.go: TestTCPTransport |
| REQ-LSP-T03 | protocol.go | protocol_test.go: TestJSONRPC |
| REQ-LSP-E01~E05 | client.go, server.go | *_test.go: TestError* |
| REQ-LSP-G01 | server.go | server_test.go: TestServerNotInstalled |
| REQ-LSP-G02 | server.go | server_test.go: TestPartialFailure |

---

## Implementation Notes

**Status**: Completed
**Implementation Date**: 2026-02-03
**Development Methodology**: Hybrid (TDD for new code, DDD for existing code)
**Test Coverage**: 91.3%

### Summary

LSP integration package implemented with multi-language server management supporting 16+ languages. Includes LSP 3.17 compliant client with initialize, diagnostics, references, hover, definition, symbols, and shutdown operations. Server lifecycle management with parallel startup, health checks, and graceful shutdown. Supports both stdio and TCP transport protocols via JSON-RPC 2.0. Pure Go implementation with CGO_ENABLED=0 compatibility.

### Files Created

- `internal/lsp/client.go`
- `internal/lsp/client_test.go`
- `internal/lsp/doc.go`
- `internal/lsp/models.go`
- `internal/lsp/models_test.go`
- `internal/lsp/protocol.go`
- `internal/lsp/protocol_test.go`
- `internal/lsp/server.go`
- `internal/lsp/server_test.go`
