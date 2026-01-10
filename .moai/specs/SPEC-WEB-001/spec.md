---
id: SPEC-WEB-001
version: "1.0.0"
status: "draft"
created: "2026-01-09"
updated: "2026-01-09"
author: "MoAI-ADK"
priority: "high"
---

# SPEC-WEB-001: MoAI Web Backend

브라우저 기반 에이전틱 코딩을 위한 FastAPI 백엔드 시스템

## HISTORY

| 버전 | 날짜 | 작성자 | 변경 내용 |
|------|------|--------|-----------|
| 1.0.0 | 2026-01-09 | MoAI-ADK | 초기 SPEC 작성 |

---

## 개요

### 목적

`moai web` CLI 명령어를 통해 FastAPI 서버를 시작하고, Claude Agent SDK를 통합하여 WebSocket 기반 실시간 통신을 제공하는 백엔드 시스템을 구현한다.

### 범위

- FastAPI 기반 REST API 서버
- Claude Agent SDK 통합
- WebSocket 실시간 채팅
- 세션 관리 및 영속성
- Provider 전환 (Claude, GLM-4.7)
- SPEC 상태 조회 API

---

## Environment (환경)

### 기술 스택

```yaml
Backend:
  - FastAPI: ">=0.115.0"
  - Uvicorn: ">=0.34.0"
  - claude-agent-sdk: ">=0.1.19"
  - websockets: ">=14.0"
  - pydantic: ">=2.9.0"
  - aiosqlite: ">=0.20.0"
```

### 시스템 요구사항

- Python: 3.11 이상
- OS: macOS, Linux, Windows
- 네트워크: HTTPS 지원
- 스토리지: SQLite (로컬 세션 저장)

---

## Assumptions (가정)

### 기술적 가정

1. Claude API 키가 환경 변수로 설정되어 있음 (`ANTHROPIC_API_KEY`)
2. 로컬 개발 환경에서 실행됨 (기본 포트: 8080)
3. 단일 사용자 시나리오 우선 지원
4. SQLite가 세션 저장에 충분한 성능을 제공함

### 비즈니스 가정

1. 사용자는 브라우저를 통해 에이전틱 코딩 환경에 접근
2. 실시간 스트리밍 응답이 필수
3. Provider 전환이 런타임에 가능해야 함

---

## Requirements (요구사항)

### Ubiquitous (항상 활성)

| ID | 요구사항 | 검증 방법 |
|----|----------|-----------|
| R-UBI-001 | 시스템은 항상 REST API 요청에 JSON 형식으로 응답해야 한다 | API 응답 Content-Type 검증 |
| R-UBI-002 | 시스템은 항상 WebSocket 연결 상태를 모니터링해야 한다 | 연결 상태 로그 확인 |
| R-UBI-003 | 시스템은 항상 claude-agent-sdk 메시지를 로깅해야 한다 | 로그 파일 분석 |

### Event-Driven (이벤트 기반)

| ID | 트리거 | 동작 | 검증 방법 |
|----|--------|------|-----------|
| R-EVT-001 | WHEN `moai web` CLI 명령이 실행되면 | THEN FastAPI 서버가 지정된 포트에서 시작되어야 한다 | 서버 시작 확인 |
| R-EVT-002 | WHEN WebSocket 연결 요청이 들어오면 | THEN 시스템은 클라이언트 세션을 생성해야 한다 | 세션 생성 로그 확인 |
| R-EVT-003 | WHEN 채팅 메시지가 수신되면 | THEN 시스템은 Claude Agent SDK를 통해 응답을 스트리밍해야 한다 | 스트리밍 응답 검증 |
| R-EVT-004 | WHEN Provider 전환이 요청되면 | THEN 환경 변수가 업데이트되어야 한다 | 환경 변수 확인 |

### State-Driven (상태 기반)

| ID | 조건 | 동작 | 검증 방법 |
|----|------|------|-----------|
| R-STA-001 | IF 에이전트 세션이 활성 상태이면 | THEN 시스템은 SPEC 상태 변경을 실시간으로 브로드캐스트해야 한다 | WebSocket 이벤트 수신 확인 |
| R-STA-002 | IF Z.AI 프록시 모드가 활성화되면 | THEN 요청이 GLM-4.7로 라우팅되어야 한다 | API 요청 헤더 검증 |

### Unwanted (금지 사항)

| ID | 금지 동작 | 이유 | 검증 방법 |
|----|----------|------|-----------|
| R-UNW-001 | 시스템은 인증되지 않은 WebSocket 연결을 허용하지 않아야 한다 | 보안 | 미인증 연결 시도 테스트 |
| R-UNW-002 | 시스템은 API 키를 응답 본문에 포함하지 않아야 한다 | 보안 | 응답 검사 자동화 |

### Optional (선택 사항)

| ID | 요구사항 | 조건 |
|----|----------|------|
| R-OPT-001 | 가능하면 CORS 설정을 통해 지정된 도메인만 접근 허용 | CORS 미들웨어 구성 |

---

## Specifications (상세 명세)

### API Endpoints

| Method | Path | 설명 |
|--------|------|------|
| GET | /health | 헬스 체크 |
| POST | /api/sessions | 채팅 세션 생성 |
| GET | /api/sessions | 세션 목록 조회 |
| GET | /api/sessions/{id} | 세션 상세 조회 |
| DELETE | /api/sessions/{id} | 세션 삭제 |
| POST | /api/chat | 메시지 전송 (SSE 스트리밍) |
| GET | /api/specs | SPEC 목록 조회 |
| GET | /api/specs/{id} | SPEC 상세 조회 |
| POST | /api/providers/switch | Provider 전환 |
| GET | /api/providers/current | 현재 Provider 조회 |
| WS | /ws/chat/{session_id} | WebSocket 채팅 |
| WS | /ws/events | 실시간 이벤트 |

### 생성할 파일 구조

```
src/moai_adk/
├── cli/commands/web.py           # moai web CLI 명령어
├── web/
│   ├── __init__.py
│   ├── server.py                 # FastAPI 애플리케이션 팩토리
│   ├── config.py                 # 서버 설정
│   ├── database.py               # SQLite 비동기 설정
│   ├── routers/
│   │   ├── __init__.py
│   │   ├── chat.py               # 채팅 엔드포인트
│   │   ├── sessions.py           # 세션 관리
│   │   ├── specs.py              # SPEC 상태 엔드포인트
│   │   ├── providers.py          # Provider 전환
│   │   └── health.py             # 헬스 체크
│   ├── services/
│   │   ├── __init__.py
│   │   ├── agent_service.py      # Claude Agent SDK 래퍼
│   │   ├── provider_service.py   # Provider 관리
│   │   └── session_service.py    # 세션 영속성
│   └── models/
│       ├── __init__.py
│       ├── session.py            # 세션 모델
│       ├── message.py            # 메시지 모델
│       └── provider.py           # Provider 모델
```

### 데이터 모델

#### Session 모델

```python
class Session(BaseModel):
    id: str                    # UUID
    created_at: datetime
    updated_at: datetime
    name: str | None = None
    provider: str = "claude"
    message_count: int = 0
```

#### Message 모델

```python
class Message(BaseModel):
    id: str                    # UUID
    session_id: str
    role: Literal["user", "assistant"]
    content: str
    created_at: datetime
    metadata: dict | None = None
```

#### Provider 모델

```python
class Provider(BaseModel):
    name: str                  # "claude" | "glm"
    base_url: str | None
    is_active: bool = False
```

### WebSocket 메시지 프로토콜

#### 클라이언트 -> 서버

```json
{
  "type": "chat",
  "content": "사용자 메시지"
}
```

#### 서버 -> 클라이언트

```json
{
  "type": "token",
  "content": "응답 토큰"
}
```

```json
{
  "type": "done",
  "message_id": "uuid"
}
```

```json
{
  "type": "error",
  "message": "에러 메시지"
}
```

---

## Dependencies (의존성)

### 의존하는 SPEC

- 없음 (Foundation SPEC)

### 이 SPEC에 의존하는 SPEC

- SPEC-WEB-002: MoAI Web Frontend
- SPEC-TERM-001: Terminal Integration
- SPEC-MODEL-001: Model Provider Abstraction
- SPEC-CMD-001: Command Extension

---

## Traceability (추적성)

### 관련 문서

- `.moai/project/tech.md`: 기술 스택 정의
- `.moai/project/product.md`: 제품 요구사항

### 태그

- `SPEC-WEB-001`
- `fastapi`
- `websocket`
- `claude-agent-sdk`
- `real-time`
