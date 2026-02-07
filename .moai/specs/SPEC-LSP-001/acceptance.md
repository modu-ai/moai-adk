---
id: SPEC-LSP-001
title: Language Server Protocol Client System - Acceptance Criteria
spec_ref: SPEC-LSP-001/spec.md
plan_ref: SPEC-LSP-001/plan.md
status: Planned
priority: High
---

# SPEC-LSP-001: Acceptance Criteria

## 1. 서버 시작 및 초기화

### AC-01: 단일 Language Server 시작

```gherkin
Feature: Language Server 시작
  Language Server를 시작하고 LSP 핸드셰이크를 완료한다.

  Scenario: 정상적인 서버 시작
    Given "go" 언어의 Language Server(gopls)가 시스템에 설치되어 있다
    And LanguageDef에 "go"의 서버 경로와 인자가 정의되어 있다
    When ServerManager.StartServer(ctx, "go")를 호출한다
    Then 서버 프로세스가 시작된다
    And JSON-RPC 연결이 수립된다
    And "initialize" 요청이 전송된다
    And 서버의 capability 응답이 수신된다
    And "initialized" 알림이 전송된다
    And ActiveServers()에 "go"가 포함된다

  Scenario: 미설치 서버 시작 시도
    Given "haskell" 언어의 Language Server가 시스템에 설치되어 있지 않다
    When ServerManager.StartServer(ctx, "haskell")를 호출한다
    Then ErrServerStartFailed 에러가 반환된다
    And 에러 메시지에 "executable not found" 내용이 포함된다
    And warning 레벨 로그가 기록된다
    And ActiveServers()에 "haskell"이 포함되지 않는다

  Scenario: 이미 실행 중인 서버 시작 시도
    Given "go" 언어의 Language Server가 이미 실행 중이다
    When ServerManager.StartServer(ctx, "go")를 호출한다
    Then 에러 없이 반환된다 (멱등성 보장)
    And 기존 서버 연결이 유지된다
    And ActiveServers()에 "go"가 하나만 포함된다
```

### AC-02: 병렬 서버 시작

```gherkin
Feature: 병렬 Language Server 시작
  여러 Language Server를 동시에 시작한다.

  Scenario: 4개 이하 서버 동시 시작
    Given 4개 언어("go", "python", "typescript", "rust")의 Language Server가 설치되어 있다
    When StartAll(ctx, ["go", "python", "typescript", "rust"])를 호출한다
    Then 4개 서버가 동시에 시작된다
    And 모든 서버의 initialize 핸드셰이크가 완료된다
    And 각 서버 시작에 500ms 이내 소요된다
    And ActiveServers()에 4개 언어가 모두 포함된다

  Scenario: 동시성 한도 초과 (5개 이상 서버)
    Given 8개 언어의 Language Server가 설치되어 있다
    When StartAll(ctx, [8개 언어])를 호출한다
    Then 최대 4개의 서버만 동시에 시작된다
    And 나머지 4개는 대기 후 순차적으로 시작된다
    And 최종적으로 8개 서버 모두 시작된다

  Scenario: 부분 실패 (일부 서버 미설치)
    Given "go"는 설치되어 있고 "zig"는 설치되어 있지 않다
    When StartAll(ctx, ["go", "zig"])를 호출한다
    Then "go" 서버는 정상 시작된다
    And "zig" 서버 실패에 대한 warning 로그가 기록된다
    And ActiveServers()에 "go"만 포함된다
    And 전체 StartAll 호출은 에러 없이 반환된다
```

---

## 2. 진단 수집 (Diagnostic Collection)

### AC-03: 단일 서버 진단 수집

```gherkin
Feature: 단일 서버 진단 수집
  특정 문서에 대한 진단 정보를 수집한다.

  Scenario: 에러가 있는 파일의 진단
    Given "go" 서버가 실행 중이다
    And "file:///project/main.go" 파일에 구문 에러가 있다
    When Client.Diagnostics(ctx, "file:///project/main.go")를 호출한다
    Then Diagnostic 슬라이스가 반환된다
    And 각 Diagnostic에 Range(Start, End), Severity, Message가 포함된다
    And Severity가 SeverityError(1)인 항목이 존재한다
    And Source 필드에 서버 식별자가 포함된다

  Scenario: 에러가 없는 파일의 진단
    Given "go" 서버가 실행 중이다
    And "file:///project/clean.go" 파일에 에러가 없다
    When Client.Diagnostics(ctx, "file:///project/clean.go")를 호출한다
    Then 빈 Diagnostic 슬라이스가 반환된다
    And 에러는 nil이다

  Scenario: 서버 미실행 시 진단 요청
    Given "python" 서버가 실행 중이지 않다
    When ServerManager.GetClient("python")을 호출한다
    Then ErrServerNotRunning 에러가 반환된다
```

### AC-04: 병렬 진단 수집

```gherkin
Feature: 병렬 진단 수집 (다중 서버)
  모든 활성 서버에서 병렬로 진단을 수집한다.

  Scenario: 16개 서버에서 동시 진단 수집
    Given 16개 언어 서버가 모두 실행 중이다
    When CollectAllDiagnostics(ctx, "file:///project/main.go")를 호출한다
    Then 모든 서버의 진단이 합쳐진 Diagnostic 슬라이스가 반환된다
    And 전체 수집 시간이 2초 이내이다
    And 각 서버의 Source 필드로 진단 출처를 구분할 수 있다

  Scenario: 일부 서버 응답 지연
    Given 16개 언어 서버 중 2개가 응답이 느리다 (> 2초)
    When CollectAllDiagnostics(ctx, uri)를 호출한다
    Then 2초 타임아웃 후 결과가 반환된다
    And 응답한 14개 서버의 진단이 포함된다
    And 느린 2개 서버에 대한 에러는 로그에만 기록된다
    And 반환된 에러는 nil이다

  Scenario: 개별 서버 에러 시 비치명적 처리
    Given 4개 서버가 실행 중이다
    And 그 중 1개 서버가 진단 요청 시 에러를 반환한다
    When CollectAllDiagnostics(ctx, uri)를 호출한다
    Then 나머지 3개 서버의 진단은 정상 수집된다
    And 실패한 서버의 에러는 warning 로그로 기록된다
    And 전체 반환 에러는 nil이다
```

---

## 3. LSP 기능 조회

### AC-05: 참조 검색 (References)

```gherkin
Feature: 심볼 참조 검색
  지정된 위치의 심볼에 대한 모든 참조를 검색한다.

  Scenario: 참조가 존재하는 심볼
    Given "go" 서버가 실행 중이고 프로젝트가 초기화되었다
    And "main.go" 파일에 여러 곳에서 참조되는 함수가 있다
    When Client.References(ctx, "file:///project/main.go", Position{Line: 10, Character: 5})를 호출한다
    Then Location 슬라이스가 반환된다
    And 각 Location에 URI와 Range가 포함된다
    And Location 수가 1개 이상이다

  Scenario: 참조가 없는 위치
    Given "go" 서버가 실행 중이다
    And 빈 공간 위치를 지정한다
    When Client.References(ctx, uri, Position{Line: 0, Character: 0})를 호출한다
    Then 빈 Location 슬라이스가 반환된다
    And 에러는 nil이다
```

### AC-06: Hover 정보 조회

```gherkin
Feature: Hover 정보 조회
  지정된 위치의 심볼에 대한 호버 정보를 반환한다.

  Scenario: 타입 정보가 있는 심볼
    Given "go" 서버가 실행 중이다
    And 변수 위치를 지정한다
    When Client.Hover(ctx, uri, pos)를 호출한다
    Then HoverResult가 반환된다
    And Contents에 타입 정보 문자열이 포함된다
    And Range가 nil이 아니다

  Scenario: Hover 정보가 없는 위치
    Given "go" 서버가 실행 중이다
    And 빈 줄의 위치를 지정한다
    When Client.Hover(ctx, uri, pos)를 호출한다
    Then nil HoverResult가 반환된다
    And 에러는 nil이다
```

### AC-07: 정의 위치 조회

```gherkin
Feature: 정의 위치 조회
  심볼의 정의 위치를 반환한다.

  Scenario: 외부 패키지 함수의 정의
    Given "go" 서버가 실행 중이다
    And 외부 패키지 함수 호출 위치를 지정한다
    When Client.Definition(ctx, uri, pos)를 호출한다
    Then Location 슬라이스가 반환된다
    And Location의 URI가 정의 파일을 가리킨다

  Scenario: 같은 파일 내 정의
    Given "go" 서버가 실행 중이다
    And 같은 파일에 정의된 변수 위치를 지정한다
    When Client.Definition(ctx, uri, pos)를 호출한다
    Then Location의 URI가 같은 파일이다
    And Range가 정의 위치를 가리킨다
```

### AC-08: 문서 심볼 조회

```gherkin
Feature: 문서 심볼 조회
  문서 내 모든 심볼의 계층 구조를 반환한다.

  Scenario: 함수와 타입이 포함된 Go 파일
    Given "go" 서버가 실행 중이다
    And 여러 함수와 struct가 정의된 Go 파일이 있다
    When Client.Symbols(ctx, uri)를 호출한다
    Then DocumentSymbol 슬라이스가 반환된다
    And 각 심볼에 Name, Kind, Range가 포함된다
    And struct 심볼의 Children에 필드 심볼이 포함된다

  Scenario: 빈 파일의 심볼 조회
    Given "go" 서버가 실행 중이다
    And 심볼이 없는 빈 파일이 있다
    When Client.Symbols(ctx, uri)를 호출한다
    Then 빈 DocumentSymbol 슬라이스가 반환된다
```

---

## 4. Health Check

### AC-09: 전체 서버 Health Check

```gherkin
Feature: Language Server Health Check
  모든 활성 서버의 건강 상태를 확인한다.

  Scenario: 모든 서버가 정상
    Given 4개 서버("go", "python", "typescript", "rust")가 실행 중이다
    And 모든 서버가 응답 가능한 상태이다
    When ServerManager.HealthCheck(ctx)를 호출한다
    Then 4개 키를 가진 map[string]error가 반환된다
    And 모든 값이 nil이다

  Scenario: 일부 서버 비정상
    Given 4개 서버 중 "rust" 서버 프로세스가 비정상 종료되었다
    When ServerManager.HealthCheck(ctx)를 호출한다
    Then "go", "python", "typescript"의 값은 nil이다
    And "rust"의 값은 에러 객체이다
    And 에러 메시지에 서버 상태 정보가 포함된다

  Scenario: 서버가 없는 상태
    Given 실행 중인 서버가 없다
    When ServerManager.HealthCheck(ctx)를 호출한다
    Then 빈 map이 반환된다
```

---

## 5. Graceful Shutdown

### AC-10: 단일 서버 종료

```gherkin
Feature: Language Server Graceful Shutdown
  서버를 안전하게 종료한다.

  Scenario: 정상 종료
    Given "go" 서버가 실행 중이다
    When ServerManager.StopServer(ctx, "go")를 호출한다
    Then "shutdown" 요청이 서버로 전송된다
    And 서버의 shutdown 응답이 수신된다
    And "exit" 알림이 서버로 전송된다
    And 서버 프로세스가 종료된다
    And ActiveServers()에 "go"가 포함되지 않는다

  Scenario: 응답 없는 서버 종료
    Given "go" 서버가 실행 중이나 응답이 없다
    When ServerManager.StopServer(ctx, "go")를 호출한다
    Then context 타임아웃 후 프로세스가 강제 종료(Kill)된다
    And ActiveServers()에 "go"가 포함되지 않는다
    And warning 로그에 강제 종료 사유가 기록된다

  Scenario: 미실행 서버 종료 시도
    Given "haskell" 서버가 실행 중이지 않다
    When ServerManager.StopServer(ctx, "haskell")를 호출한다
    Then 에러 없이 반환된다 (멱등성 보장)
```

### AC-11: 전체 서버 종료 (프로세스 정리)

```gherkin
Feature: 전체 서버 Graceful Shutdown
  모든 활성 서버를 안전하게 종료한다.

  Scenario: 모든 서버 순차 종료
    Given 4개 서버가 실행 중이다
    When 각 서버에 대해 StopServer를 호출한다
    Then 모든 서버 프로세스가 종료된다
    And ActiveServers()가 빈 슬라이스를 반환한다
    And 시스템에 Language Server 고아 프로세스가 없다
```

---

## 6. 에러 처리

### AC-12: Context 취소 처리

```gherkin
Feature: Context 취소 및 타임아웃 처리
  context.Context의 취소 신호를 존중한다.

  Scenario: 진단 수집 중 Context 취소
    Given "go" 서버가 실행 중이다
    And 타임아웃이 100ms인 Context를 생성한다
    When 오래 걸리는 Diagnostics 요청을 시작한다
    And 100ms 후 Context가 취소된다
    Then context.DeadlineExceeded 에러가 반환된다
    And 진행 중인 JSON-RPC 요청이 정리된다

  Scenario: 서버 시작 중 Context 취소
    Given 500ms 타임아웃 Context가 설정되어 있다
    And Language Server 시작이 500ms 이상 소요된다
    When StartServer(ctx, lang)를 호출한다
    Then context.DeadlineExceeded 에러가 반환된다
    And 부분적으로 시작된 프로세스가 정리된다
```

### AC-13: 비정상 서버 종료 감지

```gherkin
Feature: Language Server 비정상 종료 감지
  서버 프로세스의 예기치 않은 종료를 감지한다.

  Scenario: 서버 프로세스 크래시
    Given "go" 서버가 실행 중이다
    When 서버 프로세스가 예기치 않게 종료된다
    Then 프로세스 감시 goroutine이 종료를 감지한다
    And 내부 레지스트리에서 해당 서버가 제거된다
    And warning 로그에 비정상 종료 사유가 기록된다
    And 이후 GetClient("go") 호출 시 ErrServerNotRunning이 반환된다
```

---

## 7. Transport 계층

### AC-14: Stdio Transport

```gherkin
Feature: Stdio Transport
  stdin/stdout을 통한 JSON-RPC 통신을 수행한다.

  Scenario: Stdio 기반 서버 연결
    Given Language Server가 stdio transport를 지원한다
    When os/exec.Cmd로 서버 프로세스를 시작한다
    Then stdin 파이프로 JSON-RPC 요청을 전송할 수 있다
    And stdout 파이프로 JSON-RPC 응답을 수신할 수 있다
    And Content-Length 헤더가 올바르게 처리된다
```

### AC-15: TCP Transport

```gherkin
Feature: TCP Transport
  TCP 소켓을 통한 JSON-RPC 통신을 수행한다.

  Scenario: TCP 기반 서버 연결
    Given Language Server가 TCP transport를 지원한다
    And 서버가 지정된 포트에서 리스닝 중이다
    When net.Dial로 서버에 연결한다
    Then TCP 소켓으로 JSON-RPC 요청을 전송할 수 있다
    And TCP 소켓으로 JSON-RPC 응답을 수신할 수 있다
```

---

## 8. 동시성 안전 (Concurrency Safety)

### AC-16: Race Condition 없음

```gherkin
Feature: Race Condition 검증
  동시 접근 시 데이터 무결성을 보장한다.

  Scenario: 동시 GetClient 호출
    Given 4개 서버가 실행 중이다
    When 10개의 goroutine이 동시에 GetClient를 호출한다
    Then 모든 호출이 올바른 Client를 반환한다
    And "go test -race"에서 race condition이 감지되지 않는다

  Scenario: 서버 시작과 동시 조회
    Given StartServer와 GetClient가 동시에 호출된다
    When 새 서버가 시작되는 중에 GetClient가 호출된다
    Then GetClient는 ErrServerNotRunning 또는 유효한 Client를 반환한다
    And 프로그램이 패닉하지 않는다
    And "go test -race"에서 race condition이 감지되지 않는다

  Scenario: 동시 진단 수집과 서버 종료
    Given CollectAllDiagnostics와 StopServer가 동시에 호출된다
    When 진단 수집 중 서버가 종료된다
    Then 해당 서버의 진단은 건너뛰어진다
    And 다른 서버의 진단은 정상 수집된다
    And race condition이 감지되지 않는다
```

---

## 9. 성능 기준

### AC-17: 성능 목표 달성

```gherkin
Feature: 성능 기준 충족
  정의된 성능 예산을 충족한다.

  Scenario: 단일 서버 시작 성능
    Given Language Server 바이너리가 로컬에 존재한다
    When StartServer를 호출한다
    Then 500ms 이내에 서버 시작 및 초기화가 완료된다

  Scenario: 16개 서버 병렬 진단 수집 성능
    Given 16개 서버가 모두 실행 중이다
    When CollectAllDiagnostics를 호출한다
    Then 2초 이내에 모든 서버의 진단이 수집된다

  Scenario: Idle 메모리 사용량
    Given LSP 모듈이 로드되었으나 서버가 시작되지 않았다
    When 메모리 사용량을 측정한다
    Then 20MB 이하이다

  Scenario: Peak 메모리 사용량
    Given 16개 서버가 모두 실행 중이고 진단을 수집하고 있다
    When 메모리 사용량을 측정한다
    Then 200MB 이하이다
```

---

## 10. Quality Gate (품질 게이트)

### Definition of Done

SPEC-LSP-001이 완료된 것으로 간주되려면 다음 조건을 모두 충족해야 한다:

**코드 품질**:

- [ ] `golangci-lint run ./internal/lsp/...` 통과 (에러 0개)
- [ ] 모든 exported 타입 및 함수에 godoc 주석 작성
- [ ] `go vet ./internal/lsp/...` 경고 없음

**테스트 품질**:

- [ ] `go test -race ./internal/lsp/...` 통과
- [ ] 단위 테스트 커버리지 85% 이상 (`go test -cover`)
- [ ] Client interface 7개 메서드 각각에 대한 테스트 존재
- [ ] ServerManager interface 5개 메서드 각각에 대한 테스트 존재
- [ ] 병렬 서버 시작 테스트 통과
- [ ] 병렬 진단 수집 테스트 통과
- [ ] Context 취소 테스트 통과
- [ ] 에러 처리 테스트 (sentinel error 검증) 통과

**통합 테스트**:

- [ ] gopls와의 통합 테스트 통과 (build tag: integration)
- [ ] Initialize -> Diagnostics -> Shutdown 전체 라이프사이클 검증
- [ ] JSON-RPC round-trip 테스트 통과

**성능 테스트**:

- [ ] 벤치마크: 단일 서버 시작 < 500ms
- [ ] 벤치마크: 16개 서버 진단 수집 < 2s
- [ ] 벤치마크 결과가 `testdata/benchmarks/` 에 저장

**아키텍처 검증**:

- [ ] `go.lsp.dev/protocol` 타입이 `models.go` 외부로 노출되지 않음
- [ ] `internal/lsp/` 패키지가 `pkg/models/`와 `pkg/utils/`만 의존
- [ ] 순환 의존성 없음 (`go vet` 검증)

**문서화**:

- [ ] 모든 exported interface에 사용 예제 godoc 주석
- [ ] `internal/lsp/doc.go` 패키지 레벨 문서 작성
- [ ] 에러 처리 가이드 (소비자 모듈 참고용)
