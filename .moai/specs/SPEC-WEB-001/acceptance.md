# SPEC-WEB-001 인수 조건

## 추적성

- **SPEC ID**: SPEC-WEB-001
- **상태**: draft
- **우선순위**: HIGH

---

## 인수 테스트 시나리오

### Scenario 1: CLI 서버 시작

```gherkin
Feature: CLI를 통한 Web 서버 시작
  MoAI Web Backend를 CLI 명령어로 시작할 수 있어야 한다

  Scenario: 기본 포트로 서버 시작
    Given MoAI-ADK가 설치되어 있고
      And 유효한 Claude API 키가 환경 변수에 설정되어 있을 때
    When 사용자가 "moai web" 명령을 실행하면
    Then FastAPI 서버가 http://localhost:8080에서 시작되어야 하고
      And 콘솔에 "MoAI Web running at http://localhost:8080" 메시지가 표시되어야 한다
      And /health 엔드포인트가 {"status": "healthy"}를 반환해야 한다

  Scenario: 사용자 지정 포트로 서버 시작
    Given MoAI-ADK가 설치되어 있고
      And 유효한 Claude API 키가 환경 변수에 설정되어 있을 때
    When 사용자가 "moai web --port 3000" 명령을 실행하면
    Then FastAPI 서버가 http://localhost:3000에서 시작되어야 하고
      And 콘솔에 "MoAI Web running at http://localhost:3000" 메시지가 표시되어야 한다

  Scenario: API 키 없이 서버 시작 시도
    Given MoAI-ADK가 설치되어 있고
      And ANTHROPIC_API_KEY 환경 변수가 설정되어 있지 않을 때
    When 사용자가 "moai web" 명령을 실행하면
    Then 시스템은 에러 메시지를 표시해야 하고
      And 서버가 시작되지 않아야 한다
```

### Scenario 2: WebSocket 채팅 세션

```gherkin
Feature: WebSocket 기반 실시간 채팅
  사용자가 WebSocket을 통해 Claude와 실시간 대화할 수 있어야 한다

  Scenario: 채팅 세션 생성 및 메시지 전송
    Given Web 서버가 http://localhost:8080에서 실행 중이고
      And 클라이언트가 /ws/chat/session-123으로 WebSocket 연결을 수립했을 때
    When 사용자가 {"type": "chat", "content": "Hello"}를 전송하면
    Then 시스템은 Claude Agent SDK를 통해 응답을 생성하고
      And {"type": "token", "content": "..."}을 스트리밍으로 전송해야 한다
      And 응답 완료 시 {"type": "done"}을 전송해야 한다

  Scenario: 메시지 컨텍스트 유지
    Given 활성화된 채팅 세션이 있고
      And 이전에 "나는 홍길동이야"라는 메시지를 보냈을 때
    When "내 이름이 뭐야?"라는 메시지를 전송하면
    Then 시스템은 "홍길동"이 포함된 응답을 반환해야 한다

  Scenario: WebSocket 연결 해제 처리
    Given 활성화된 WebSocket 연결이 있을 때
    When 클라이언트가 연결을 종료하면
    Then 시스템은 세션 상태를 업데이트해야 하고
      And 연결 종료 로그를 기록해야 한다
```

### Scenario 3: Provider 전환

```gherkin
Feature: AI Provider 런타임 전환
  사용자가 런타임에 AI Provider를 전환할 수 있어야 한다

  Scenario: Claude에서 GLM으로 전환
    Given 현재 Provider가 "claude"로 설정되어 있을 때
    When POST /api/providers/switch {"provider": "glm"}을 호출하면
    Then 환경 변수 ANTHROPIC_BASE_URL이 "https://api.z.ai/api/anthropic"로 설정되어야 하고
      And 응답이 {"provider": "glm", "status": "switched"}를 반환해야 한다

  Scenario: GLM에서 Claude로 전환
    Given 현재 Provider가 "glm"로 설정되어 있을 때
    When POST /api/providers/switch {"provider": "claude"}을 호출하면
    Then 환경 변수 ANTHROPIC_BASE_URL이 기본값으로 복원되어야 하고
      And 응답이 {"provider": "claude", "status": "switched"}를 반환해야 한다

  Scenario: 현재 Provider 조회
    Given Provider가 설정되어 있을 때
    When GET /api/providers/current를 호출하면
    Then 응답에 현재 활성화된 Provider 정보가 포함되어야 한다
```

### Scenario 4: SPEC 상태 조회

```gherkin
Feature: SPEC 상태 조회 API
  사용자가 REST API를 통해 SPEC 상태를 조회할 수 있어야 한다

  Scenario: SPEC 목록 조회
    Given .moai/specs/에 SPEC-AUTH-001이 존재할 때
    When GET /api/specs를 호출하면
    Then 응답이 [{"id": "SPEC-AUTH-001", "status": "draft", ...}]를 포함해야 한다

  Scenario: SPEC 상세 조회
    Given .moai/specs/SPEC-AUTH-001/spec.md가 존재할 때
    When GET /api/specs/SPEC-AUTH-001을 호출하면
    Then 응답에 SPEC의 전체 메타데이터와 내용이 포함되어야 한다

  Scenario: 존재하지 않는 SPEC 조회
    Given .moai/specs/에 SPEC-NONE-001이 존재하지 않을 때
    When GET /api/specs/SPEC-NONE-001을 호출하면
    Then 404 Not Found 응답을 반환해야 한다
```

### Scenario 5: 세션 관리

```gherkin
Feature: 채팅 세션 관리
  사용자가 채팅 세션을 생성, 조회, 삭제할 수 있어야 한다

  Scenario: 새 세션 생성
    Given Web 서버가 실행 중일 때
    When POST /api/sessions {"name": "Test Session"}을 호출하면
    Then 새 세션이 생성되고
      And 응답에 session_id가 포함되어야 한다

  Scenario: 세션 목록 조회
    Given 하나 이상의 세션이 존재할 때
    When GET /api/sessions를 호출하면
    Then 모든 세션 목록이 반환되어야 한다

  Scenario: 세션 삭제
    Given session-123 세션이 존재할 때
    When DELETE /api/sessions/session-123을 호출하면
    Then 세션이 삭제되어야 하고
      And 204 No Content 응답을 반환해야 한다
```

---

## Edge Case 시나리오

### Edge Case 1: 대용량 메시지 처리

```gherkin
Scenario: 매우 긴 메시지 전송
  Given 활성화된 채팅 세션이 있을 때
  When 100,000자 이상의 메시지를 전송하면
  Then 시스템은 적절한 에러 메시지를 반환해야 하고
    And WebSocket 연결이 유지되어야 한다
```

### Edge Case 2: 동시 연결 처리

```gherkin
Scenario: 다중 WebSocket 연결
  Given Web 서버가 실행 중일 때
  When 10개의 동시 WebSocket 연결이 수립되면
  Then 모든 연결이 독립적으로 동작해야 하고
    And 각 세션의 메시지가 올바른 클라이언트에게 전달되어야 한다
```

### Edge Case 3: 네트워크 지연

```gherkin
Scenario: 느린 네트워크 환경
  Given 네트워크 지연이 5초 이상인 환경에서
  When 메시지를 전송하면
  Then 시스템은 타임아웃 처리를 해야 하고
    And 클라이언트에게 적절한 상태 정보를 제공해야 한다
```

---

## 에러 처리 시나리오

### Error Scenario 1: Claude API 오류

```gherkin
Scenario: Claude API 레이트 리밋
  Given Claude API가 레이트 리밋 상태일 때
  When 채팅 메시지를 전송하면
  Then 시스템은 {"type": "error", "message": "Rate limit exceeded"}를 반환해야 하고
    And 재시도 정보를 포함해야 한다
```

### Error Scenario 2: 잘못된 요청

```gherkin
Scenario: 잘못된 JSON 형식
  Given 활성화된 WebSocket 연결이 있을 때
  When 잘못된 JSON 형식의 메시지를 전송하면
  Then 시스템은 {"type": "error", "message": "Invalid message format"}를 반환해야 한다
```

### Error Scenario 3: 인증 실패

```gherkin
Scenario: 유효하지 않은 세션 ID
  Given 유효하지 않은 세션 ID로 WebSocket 연결을 시도할 때
  Then 연결이 거부되어야 하고
    And 4003 에러 코드가 반환되어야 한다
```

---

## 성능 기준

### 응답 시간

| 항목 | P50 | P95 | P99 |
|------|-----|-----|-----|
| 헬스 체크 | < 10ms | < 50ms | < 100ms |
| 세션 생성 | < 50ms | < 100ms | < 200ms |
| SPEC 목록 조회 | < 100ms | < 200ms | < 500ms |
| WebSocket 메시지 수신 | < 5ms | < 20ms | < 50ms |

### 동시성

| 항목 | 최소 요구 |
|------|----------|
| 동시 WebSocket 연결 | 100개 |
| 동시 REST API 요청 | 50 req/sec |

### 리소스 사용

| 항목 | 제한 |
|------|------|
| 메모리 사용량 | < 500MB |
| CPU 사용률 (대기) | < 5% |
| SQLite DB 크기 | < 100MB |

---

## Definition of Done

### 기능 완료 조건

- [ ] 모든 API 엔드포인트 구현 완료
- [ ] WebSocket 채팅 기능 구현 완료
- [ ] Provider 전환 기능 구현 완료
- [ ] SPEC 상태 조회 기능 구현 완료
- [ ] CLI 명령어 구현 완료

### 품질 조건

- [ ] 테스트 커버리지 85% 이상
- [ ] 모든 Gherkin 시나리오 통과
- [ ] ruff 린트 경고 0개
- [ ] 보안 취약점 스캔 통과
- [ ] API 문서 자동 생성 (OpenAPI)

### 문서화 조건

- [ ] API 엔드포인트 문서화
- [ ] WebSocket 프로토콜 문서화
- [ ] 설치 및 사용 가이드

---

## 참조

- **SPEC 문서**: `.moai/specs/SPEC-WEB-001/spec.md`
- **구현 계획**: `.moai/specs/SPEC-WEB-001/plan.md`
