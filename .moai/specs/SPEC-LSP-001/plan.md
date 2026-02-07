---
id: SPEC-LSP-001
title: Language Server Protocol Client System - Implementation Plan
spec_ref: SPEC-LSP-001/spec.md
status: Planned
priority: High
---

# SPEC-LSP-001: Implementation Plan

## 1. 구현 전략 개요

### 1.1 접근 방식

Bottom-up 방식으로 구현한다. Protocol 계층(JSON-RPC 2.0 통신)을 먼저 구축하고, 그 위에 Client 계층(단일 서버 통신), 그리고 최상위에 ServerManager 계층(다중 서버 관리)을 쌓는다.

### 1.2 구현 순서 근거

1. **models.go** (데이터 모델) -- 다른 모든 파일이 의존하는 타입 정의
2. **protocol.go** (프로토콜 계층) -- JSON-RPC 2.0 인코딩/디코딩은 Client의 기반
3. **client.go** (클라이언트 계층) -- 단일 서버 통신, Protocol 계층 위에 구축
4. **server.go** (서버 매니저) -- 다중 서버 관리, Client와 Protocol에 의존

### 1.3 DDD 사이클 적용

SPEC-LSP-001은 신규 모듈이므로 ANALYZE-PRESERVE-IMPROVE 중 ANALYZE와 IMPROVE에 집중한다. 기존 코드가 없으므로 PRESERVE(기존 동작 보존)는 해당하지 않는다. 단, 소비자 모듈(`internal/core/quality/`, `internal/hook/`)의 인터페이스 기대치를 ANALYZE하여 호환성을 보장한다.

---

## 2. Milestone (우선순위 기반)

### Primary Goal: 데이터 모델 및 프로토콜 계층

**파일**: `models.go`, `protocol.go`

**models.go 구현 태스크**:

- [ ] Diagnostic, DiagnosticSeverity 타입 정의
- [ ] Position, Range, Location 타입 정의
- [ ] HoverResult, DocumentSymbol 타입 정의
- [ ] sentinel 에러 변수 정의 (ErrServerNotRunning, ErrServerStartFailed, ErrInitializeFailed, ErrConnectionClosed)
- [ ] `go.lsp.dev/protocol` 타입과의 변환 함수 구현
- [ ] JSON 직렬화/역직렬화 검증용 단위 테스트

**protocol.go 구현 태스크**:

- [ ] JSON-RPC 2.0 메시지 구조체 정의 (Request, Response, Notification)
- [ ] `go.lsp.dev/jsonrpc2`를 래핑한 Transport 인터페이스 정의
- [ ] Stdio Transport 구현 (os/exec 프로세스의 stdin/stdout)
- [ ] TCP Transport 구현 (net.Dial 기반)
- [ ] Request ID 자동 생성 및 응답 매칭 로직
- [ ] 단위 테스트: JSON-RPC 메시지 인코딩/디코딩, ID 매칭

**품질 기준**:

- 모든 데이터 모델에 대한 JSON round-trip 테스트 통과
- `go.lsp.dev/protocol` 타입과의 양방향 변환 검증
- 테스트 커버리지 90% 이상

### Secondary Goal: Client 계층

**파일**: `client.go`

**구현 태스크**:

- [ ] Client interface 정의 (spec.md 4.2 참조)
- [ ] `lspClient` 구조체 구현 (Transport, rootURI, capabilities 필드)
- [ ] Initialize: `initialize` request + `initialized` notification 시퀀스
- [ ] Diagnostics: `textDocument/publishDiagnostics` 캐시 기반 반환
- [ ] References: `textDocument/references` request/response 처리
- [ ] Hover: `textDocument/hover` request/response 처리
- [ ] Definition: `textDocument/definition` request/response 처리
- [ ] Symbols: `textDocument/documentSymbol` request/response 처리
- [ ] Shutdown: `shutdown` request + `exit` notification + 연결 정리
- [ ] Context 취소 처리: 모든 메서드에서 ctx.Done() 확인
- [ ] 에러 래핑: `fmt.Errorf("lsp: %s: %w", method, err)` 패턴
- [ ] 단위 테스트: Mock Transport를 사용한 각 메서드 테스트
- [ ] 통합 테스트: 실제 gopls와의 통신 테스트 (build tag로 분리)

**품질 기준**:

- 모든 7개 Client 메서드에 대한 단위 테스트
- Context 취소 시 정상 반환 확인
- 에러 래핑 체인 검증
- 테스트 커버리지 90% 이상

### Tertiary Goal: ServerManager 계층

**파일**: `server.go`

**구현 태스크**:

- [ ] ServerManager interface 정의 (spec.md 4.2 참조)
- [ ] `serverManager` 구조체 구현 (servers map, mu sync.RWMutex, config, langRegistry)
- [ ] StartServer: LanguageDef 조회 -> exec.LookPath -> 프로세스 시작 -> Client 생성 -> Initialize
- [ ] StopServer: Client.Shutdown -> 프로세스 종료 대기 -> 레지스트리 제거
- [ ] GetClient: RWMutex.RLock 기반 안전한 클라이언트 조회
- [ ] ActiveServers: 현재 실행 중인 서버 언어 목록 반환
- [ ] HealthCheck: 모든 활성 서버에 대한 건강 상태 확인
- [ ] StartAll: errgroup.SetLimit(4) 기반 병렬 서버 시작
- [ ] CollectAllDiagnostics: errgroup + sync.Mutex 기반 병렬 진단 수집
- [ ] Graceful Degradation: 서버 미설치/시작 실패 시 warning 로그 + 계속 진행
- [ ] 프로세스 감시: 비정상 종료 감지 및 레지스트리 정리
- [ ] 단위 테스트: Mock Client를 사용한 ServerManager 테스트
- [ ] 벤치마크 테스트: 16개 서버 동시 시작/진단 수집 성능 측정

**품질 기준**:

- 병렬 시작 4개 동시 한도 준수 확인
- 16개 서버 진단 수집 2초 이내 확인 (벤치마크)
- Race condition 없음 확인 (`go test -race`)
- 테스트 커버리지 85% 이상

### Optional Goal: 확장 및 최적화

- [ ] Diagnostic 캐시 시스템 구현 (TTL 기반, `lsp_quality_gates.cache_ttl_seconds` 설정 연동)
- [ ] Connection pooling: 재사용 가능한 LSP 연결 풀
- [ ] Lazy initialization: 첫 번째 요청 시에만 서버 시작
- [ ] Retry 로직: 일시적 통신 실패에 대한 재시도 (최대 3회)
- [ ] Metrics 수집: 서버별 응답 시간, 에러율 추적
- [ ] `log/slog` 구조화 로깅: 모든 주요 연산에 structured logging 적용

---

## 3. 기술적 접근 방식

### 3.1 Protocol 계층 설계

`go.lsp.dev/jsonrpc2` 패키지를 래핑하여 MoAI-ADK 전용 추상화를 제공한다.

```
Transport (interface)
  |
  +-- StdioTransport: os/exec.Cmd의 stdin/stdout 파이프
  |
  +-- TCPTransport: net.Dial 기반 TCP 소켓
```

Request ID는 atomic counter로 관리하여 goroutine-safe한 고유 ID를 보장한다.

### 3.2 Client 계층 설계

단일 Language Server와의 전체 통신 라이프사이클을 관리한다.

```
lspClient struct {
    transport   Transport
    conn        jsonrpc2.Conn
    rootURI     string
    capabilities ServerCapabilities
    mu          sync.Mutex    // 직렬화된 요청 (선택적)
}
```

각 LSP 메서드는 JSON-RPC request를 구성하고, 응답을 파싱하여 MoAI 내부 타입으로 변환한다. `go.lsp.dev/protocol` 타입을 직접 노출하지 않고, `models.go`의 자체 타입으로 변환하여 외부 의존성 격리를 달성한다.

### 3.3 ServerManager 계층 설계

다중 Language Server의 라이프사이클을 관리한다.

```
serverManager struct {
    servers     map[string]*managedServer  // lang -> server
    mu          sync.RWMutex
    config      config.Manager
    langReg     foundation.LanguageRegistry
    logger      *slog.Logger
}

managedServer struct {
    client   Client
    process  *os.Process
    lang     string
    startedAt time.Time
}
```

### 3.4 에러 처리 전략

모든 에러는 sentinel error + context wrapping 패턴을 따른다:

```go
// 에러 반환 예시
return fmt.Errorf("starting %s server: %w", lang, ErrServerStartFailed)

// 소비자 검사 예시
if errors.Is(err, lsp.ErrServerNotRunning) {
    // 서버가 실행 중이지 않음
}
```

### 3.5 테스트 전략

| 테스트 유형 | 위치 | 접근 방식 |
|------------|------|-----------|
| Unit Test | `*_test.go` (same package) | Mock Transport, Mock Client |
| Integration Test | `*_integration_test.go` | 실제 gopls 사용, build tag `integration` |
| Benchmark Test | `*_bench_test.go` | 16개 서버 시뮬레이션 |
| Fuzz Test | `*_fuzz_test.go` | JSON-RPC 메시지 파싱 |
| Race Detection | 모든 테스트 | `go test -race` |

**Mock 생성**: `mockery`를 사용하여 Client, ServerManager, Transport interface의 mock 자동 생성

---

## 4. 리스크 및 대응 방안

### 4.1 Risk: Language Server 가용성

- **설명**: 통합 테스트 환경(CI)에서 16개 이상의 Language Server가 모두 설치되어 있지 않을 수 있다.
- **영향**: High -- 통합 테스트 실행 불가
- **대응 방안**:
  - 단위 테스트는 Mock Transport로 완전 격리하여 Language Server 없이도 실행 가능
  - 통합 테스트는 `//go:build integration` build tag로 분리
  - CI에서는 gopls(Go)와 pyright(Python)만 설치하여 핵심 경로 검증
  - 나머지 Language Server는 선택적 통합 테스트로 분류

### 4.2 Risk: JSON-RPC 호환성 문제

- **설명**: 특정 Language Server가 LSP 3.17 사양을 완전히 준수하지 않을 수 있다.
- **영향**: Medium -- 일부 언어에서 기능 동작 실패
- **대응 방안**:
  - 필수 capability(`textDocument/publishDiagnostics`)만 요구하고, 나머지는 optional로 처리
  - 서버 capability 응답을 파싱하여 지원하지 않는 기능 호출 방지
  - Graceful degradation: 미지원 기능 호출 시 빈 결과 + warning 로그

### 4.3 Risk: 동시성 관련 버그

- **설명**: 16개 이상의 서버와 동시 통신 시 race condition, deadlock 발생 가능
- **영향**: High -- 데이터 손상, 프로세스 행 발생
- **대응 방안**:
  - 모든 테스트에 `-race` 플래그 적용
  - `sync.RWMutex`로 서버 레지스트리 보호
  - `errgroup`으로 goroutine 라이프사이클 관리
  - `context.WithTimeout`으로 모든 외부 호출에 시간 제한 적용

### 4.4 Risk: 프로세스 누수

- **설명**: Language Server 프로세스가 정리되지 않아 시스템 리소스를 소모할 수 있다.
- **영향**: Medium -- 메모리 및 프로세스 누수
- **대응 방안**:
  - `defer process.Kill()` 패턴으로 프로세스 정리 보장
  - `ServerManager.StopServer`에서 프로세스 종료 대기 타임아웃 (5초)
  - 비정상 종료 감지: `process.Wait()` goroutine으로 프로세스 상태 모니터링
  - `ServerManager` 소멸 시 모든 활성 서버 강제 종료

### 4.5 Risk: go.lsp.dev 패키지 안정성

- **설명**: `go.lsp.dev/protocol`과 `go.lsp.dev/jsonrpc2`는 community-maintained로 breaking change 가능성이 있다.
- **영향**: Low -- API 변경 시 adapter 수정 필요
- **대응 방안**:
  - `models.go`에서 자체 타입 정의로 외부 의존성 격리
  - 변환 함수를 통한 adapter 패턴으로 교체 용이성 확보
  - `go.sum` 핀으로 버전 고정

---

## 5. 아키텍처 설계 방향

### 5.1 계층 구조

```
+------------------------------------------+
|           ServerManager (server.go)       |
|  - 다중 서버 관리, errgroup 병렬 처리     |
|  - sync.RWMutex 기반 레지스트리           |
+------------------------------------------+
                    |
+------------------------------------------+
|             Client (client.go)            |
|  - 단일 서버 통신, LSP 메서드 호출        |
|  - context 기반 타임아웃/취소             |
+------------------------------------------+
                    |
+------------------------------------------+
|           Protocol (protocol.go)          |
|  - JSON-RPC 2.0 인코딩/디코딩            |
|  - Stdio/TCP Transport                   |
+------------------------------------------+
                    |
+------------------------------------------+
|            Models (models.go)             |
|  - Diagnostic, Position, Range 등 타입    |
|  - Sentinel error 정의                    |
+------------------------------------------+
```

### 5.2 외부 패키지 격리 원칙

`go.lsp.dev/protocol`의 타입을 직접 노출하지 않는다. `models.go`에 자체 타입을 정의하고, `client.go` 내부에서만 변환 함수를 통해 `go.lsp.dev` 타입과 상호 변환한다. 이를 통해:

- 소비자 모듈(`quality/`, `hook/`)은 `go.lsp.dev`에 대한 의존성이 없다
- 향후 `go.lsp.dev` 패키지 교체 시 `client.go`만 수정하면 된다
- `models.go`의 타입은 MoAI-ADK 전체에서 안정적인 계약 역할을 한다

### 5.3 설정 연동

`internal/config/`에서 제공하는 `QualityConfig.LSPQualityGates` 설정과 연동:

- `cache_ttl_seconds`: Diagnostic 캐시 TTL
- `timeout_seconds`: 개별 요청 타임아웃
- Phase별 threshold (plan/run/sync): 소비자인 `quality/` 모듈에서 적용

---

## 6. 의존성 및 선행 조건

| 의존성 | 유형 | 상태 | 영향 |
|--------|------|------|------|
| SPEC-CONFIG-001 | SPEC 의존성 | Planned | config.Manager 인터페이스 필요 |
| `internal/foundation/langs.go` | 코드 의존성 | 구현 예정 | LanguageDef, LanguageRegistry 필요 |
| `go.lsp.dev/protocol` v0.12+ | 외부 패키지 | go.mod 등록 필요 | LSP 타입 정의 |
| `go.lsp.dev/jsonrpc2` v0.10+ | 외부 패키지 | go.mod 등록 필요 | JSON-RPC transport |
| `golang.org/x/sync` | 외부 패키지 | go.mod 등록 필요 | errgroup |
| `github.com/vektra/mockery` | 개발 도구 | 설치 필요 | Mock 생성 |

---

## 7. 검증 체크리스트

- [ ] Client interface의 7개 메서드 모두 구현 완료
- [ ] ServerManager interface의 5개 메서드 모두 구현 완료
- [ ] 병렬 서버 시작 (errgroup.SetLimit(4)) 구현 및 검증
- [ ] 병렬 진단 수집 (2초 타임아웃) 구현 및 검증
- [ ] Stdio Transport 구현 및 테스트
- [ ] TCP Transport 구현 및 테스트
- [ ] 모든 sentinel error 타입 정의 및 에러 래핑 적용
- [ ] Graceful degradation: 서버 미설치, 부분 실패 처리
- [ ] Context 취소 및 타임아웃 처리
- [ ] `go test -race` 통과 (race condition 없음)
- [ ] 단위 테스트 커버리지 85% 이상
- [ ] gopls 통합 테스트 통과 (build tag: integration)
- [ ] 벤치마크: 16개 서버 진단 수집 < 2초
- [ ] `golangci-lint run` 통과 (lint 에러 없음)
- [ ] godoc 주석 완료 (모든 exported 타입 및 함수)
