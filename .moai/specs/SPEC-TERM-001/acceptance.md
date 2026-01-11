---
id: SPEC-TERM-001
type: acceptance
version: "1.0.0"
created: "2026-01-09"
updated: "2026-01-09"
---

# SPEC-TERM-001: Acceptance Criteria

Parallel Terminal System 인수 조건

## TAG

`SPEC-TERM-001` | `xterm.js` | `pty` | `websocket` | `terminal`

---

## 테스트 시나리오

### Feature: 터미널 생성 및 연결

#### Scenario: 새 터미널 생성 시 PTY 프로세스 생성

```gherkin
Given MoAI Web 서버가 실행 중이고
And 활성 터미널이 6개 미만일 때
When 사용자가 새 터미널 생성을 요청하면
Then PTY 프로세스가 생성되고
And WebSocket 연결이 수립되고
And xterm.js에 터미널 프롬프트가 표시된다
```

#### Scenario: 최대 터미널 수 도달 시 생성 차단

```gherkin
Given 활성 터미널이 6개일 때
When 사용자가 새 터미널 생성을 요청하면
Then 터미널 생성이 거부되고
And "최대 터미널 수에 도달했습니다" 메시지가 표시된다
```

---

### Feature: 터미널 입출력

#### Scenario: 키보드 입력이 PTY로 전달

```gherkin
Given 활성 터미널이 연결되어 있을 때
When 사용자가 "ls -la" 명령을 입력하면
Then 입력이 WebSocket을 통해 서버로 전송되고
And PTY 프로세스에 명령이 전달되고
And 명령 실행 결과가 xterm.js에 표시된다
```

#### Scenario: PTY 출력이 실시간 렌더링

```gherkin
Given 활성 터미널에서 장시간 실행 명령 수행 중일 때
When PTY가 출력을 생성하면
Then 출력이 100ms 이내에 xterm.js에 렌더링된다
```

#### Scenario: 특수 키 입력 처리

```gherkin
Given 활성 터미널이 연결되어 있을 때
When 사용자가 Ctrl+C를 입력하면
Then SIGINT 시그널이 PTY 프로세스에 전달되고
And 실행 중인 명령이 종료된다
```

---

### Feature: 터미널 크기 조정

#### Scenario: 브라우저 창 크기 변경 시 터미널 크기 조정

```gherkin
Given 활성 터미널이 표시되어 있을 때
When 사용자가 브라우저 창 크기를 변경하면
Then FitAddon이 xterm.js 크기를 조정하고
And 새 rows/cols 값이 서버로 전송되고
And PTY 윈도우 크기가 업데이트된다
```

---

### Feature: 다중 터미널 관리

#### Scenario: 탭을 통한 터미널 전환

```gherkin
Given 2개 이상의 터미널이 활성화되어 있을 때
When 사용자가 다른 터미널 탭을 클릭하면
Then 해당 터미널이 활성화되고
And 이전 터미널의 입력 포커스가 해제된다
```

#### Scenario: 터미널 종료

```gherkin
Given 활성 터미널이 있을 때
When 사용자가 터미널 종료 버튼을 클릭하면
Then PTY 프로세스가 종료되고
And WebSocket 연결이 종료되고
And 터미널 탭이 제거된다
```

---

### Feature: 자동 명령 실행

#### Scenario: SPEC 연결 터미널에서 자동 명령 실행

```gherkin
Given SPEC-001에 연결된 터미널이 있을 때
When 시스템이 "claude /moai:2-run SPEC-001" 자동 명령을 실행하면
Then 명령이 터미널에 입력되고
And 명령 실행이 시작되고
And 실행 결과가 실시간으로 표시된다
```

---

### Feature: 비활성 터미널 정리

#### Scenario: 30분 비활성 터미널 자동 종료

```gherkin
Given 터미널이 30분 동안 입력이 없을 때
When 정리 태스크가 실행되면
Then PTY 프로세스가 종료되고
And 리소스가 해제되고
And 사용자에게 종료 알림이 표시된다
```

---

## Quality Gates (품질 게이트)

### 기능 완성도

| 기준 | 목표 | 측정 방법 |
|------|------|-----------|
| 핵심 기능 구현 | 100% | EARS 요구사항 체크리스트 |
| 테스트 시나리오 통과 | 100% | Gherkin 시나리오 실행 |
| 엣지 케이스 처리 | 90% 이상 | 엣지 케이스 테스트 |

### 코드 품질

| 기준 | 목표 | 측정 방법 |
|------|------|-----------|
| 테스트 커버리지 | 85% 이상 | pytest --cov |
| 린터 경고 | 0개 | ruff check |
| 타입 검사 통과 | 100% | mypy --strict |

### 성능 기준

| 기준 | 목표 | 측정 방법 |
|------|------|-----------|
| 입력 지연 | 50ms 이하 | 네트워크 지연 측정 |
| 출력 렌더링 | 100ms 이하 | 프레임 타이밍 |
| 메모리 사용 (터미널당) | 50MB 이하 | 메모리 프로파일링 |
| 동시 터미널 | 6개 안정 동작 | 부하 테스트 |

### 보안 기준

| 기준 | 목표 | 검증 방법 |
|------|------|-----------|
| 명령 주입 방지 | 통과 | 보안 테스트 |
| WebSocket 인증 | 필수 | 인증 검증 |
| 프로세스 격리 | 확인 | 권한 검증 |

---

## Definition of Done (완료 정의)

### 필수 조건

- [ ] 모든 EARS 요구사항 구현
- [ ] 모든 Gherkin 시나리오 통과
- [ ] 테스트 커버리지 85% 이상
- [ ] 린터/타입 체크 통과
- [ ] API 문서화 완료
- [ ] 코드 리뷰 승인

### 성능 조건

- [ ] 입력 지연 50ms 이하
- [ ] 출력 렌더링 100ms 이하
- [ ] 6개 동시 터미널 안정 동작
- [ ] 메모리 누수 없음

### 문서화 조건

- [ ] API 엔드포인트 문서화
- [ ] WebSocket 프로토콜 문서화
- [ ] 컴포넌트 Props 문서화
- [ ] 트러블슈팅 가이드

---

## Test Matrix (테스트 매트릭스)

### 브라우저 호환성

| 브라우저 | 버전 | 테스트 결과 |
|----------|------|-------------|
| Chrome | 120+ | - |
| Firefox | 120+ | - |
| Safari | 17+ | - |
| Edge | 120+ | - |

### OS 호환성

| OS | 버전 | PTY 지원 | 테스트 결과 |
|----|------|----------|-------------|
| macOS | 13+ | Native | - |
| Ubuntu | 22.04+ | Native | - |
| Windows | 10+ | WSL | - |

---

## Traceability (추적성)

### 요구사항-테스트 매핑

| 요구사항 ID | 테스트 시나리오 | 우선순위 |
|-------------|-----------------|----------|
| R-EVT-001 | 새 터미널 생성 시 PTY 프로세스 생성 | High |
| R-EVT-002 | 키보드 입력이 PTY로 전달 | High |
| R-EVT-003 | PTY 출력이 실시간 렌더링 | High |
| R-EVT-004 | 터미널 종료 | High |
| R-EVT-005 | 브라우저 창 크기 변경 시 터미널 크기 조정 | Medium |
| R-EVT-006 | SPEC 연결 터미널에서 자동 명령 실행 | Medium |
| R-STA-001 | 탭을 통한 터미널 전환 | Medium |
| R-STA-002 | 30분 비활성 터미널 자동 종료 | Low |
| R-STA-003 | 최대 터미널 수 도달 시 생성 차단 | Medium |
| R-UNW-001 | 30분 비활성 터미널 자동 종료 | Low |
| R-UNW-002 | 최대 터미널 수 도달 시 생성 차단 | Medium |

### 관련 SPEC

- `SPEC-WEB-001`: 의존 (FastAPI 서버)
