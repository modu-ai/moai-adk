# SPEC-WEB-001 구현 계획

## 추적성

- **SPEC ID**: SPEC-WEB-001
- **상태**: draft
- **우선순위**: HIGH

---

## 구현 단계

### Phase 1: Primary Goals (핵심 기능)

#### 1.1 프로젝트 구조 설정

- [ ] `src/moai_adk/web/` 디렉토리 구조 생성
- [ ] `pyproject.toml`에 web 의존성 추가 (FastAPI, Uvicorn, websockets)
- [ ] 기본 설정 모듈 (`config.py`) 구현

#### 1.2 CLI 명령어 구현

- [ ] `src/moai_adk/cli/commands/web.py` 생성
- [ ] `moai web` 명령어 등록
- [ ] 포트, 호스트 옵션 처리
- [ ] 서버 시작/종료 로직

#### 1.3 FastAPI 서버 기본 구조

- [ ] `server.py`: FastAPI 애플리케이션 팩토리
- [ ] `routers/health.py`: 헬스 체크 엔드포인트
- [ ] CORS 미들웨어 설정
- [ ] 에러 핸들러 설정

#### 1.4 세션 관리

- [ ] `database.py`: SQLite 비동기 설정
- [ ] `models/session.py`: 세션 Pydantic 모델
- [ ] `services/session_service.py`: 세션 CRUD 서비스
- [ ] `routers/sessions.py`: 세션 REST API

### Phase 2: Secondary Goals (채팅 기능)

#### 2.1 WebSocket 채팅

- [ ] `routers/chat.py`: WebSocket 채팅 엔드포인트
- [ ] 연결 관리 (연결/해제 핸들링)
- [ ] 메시지 프로토콜 구현
- [ ] 인증 미들웨어 (토큰 기반)

#### 2.2 Claude Agent SDK 통합

- [ ] `services/agent_service.py`: Agent SDK 래퍼
- [ ] 스트리밍 응답 처리
- [ ] 메시지 컨텍스트 관리
- [ ] 에러 핸들링

#### 2.3 메시지 영속성

- [ ] `models/message.py`: 메시지 모델
- [ ] 메시지 저장/조회 서비스
- [ ] 대화 히스토리 관리

### Phase 3: Tertiary Goals (Provider 관리)

#### 3.1 Provider 전환

- [ ] `models/provider.py`: Provider 모델
- [ ] `services/provider_service.py`: Provider 관리 서비스
- [ ] `routers/providers.py`: Provider API
- [ ] 환경 변수 동적 업데이트

#### 3.2 Z.AI GLM-4.7 지원

- [ ] GLM Provider 구현
- [ ] API 호환 레이어
- [ ] 프록시 URL 설정

### Phase 4: Final Goals (SPEC 통합)

#### 4.1 SPEC 상태 API

- [ ] `routers/specs.py`: SPEC 조회 API
- [ ] `.moai/specs/` 디렉토리 파싱
- [ ] SPEC 메타데이터 추출

#### 4.2 실시간 이벤트

- [ ] `/ws/events` WebSocket 엔드포인트
- [ ] SPEC 상태 변경 브로드캐스트
- [ ] 이벤트 구독 관리

---

## 기술적 접근 방식

### 아키텍처 설계

```
┌─────────────────────────────────────────────────────────┐
│                    CLI Layer                             │
│  moai web --port 8080                                    │
└─────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│                  FastAPI Application                     │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐    │
│  │ Health  │  │Sessions │  │  Chat   │  │  Specs  │    │
│  │ Router  │  │ Router  │  │ Router  │  │ Router  │    │
│  └─────────┘  └─────────┘  └─────────┘  └─────────┘    │
└─────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│                   Service Layer                          │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐     │
│  │   Session   │  │    Agent    │  │  Provider   │     │
│  │   Service   │  │   Service   │  │   Service   │     │
│  └─────────────┘  └─────────────┘  └─────────────┘     │
└─────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│                   Data Layer                             │
│  ┌─────────────┐  ┌─────────────┐                       │
│  │   SQLite    │  │   Claude    │                       │
│  │  (Sessions) │  │  Agent SDK  │                       │
│  └─────────────┘  └─────────────┘                       │
└─────────────────────────────────────────────────────────┘
```

### 비동기 패턴

- FastAPI의 async/await 전면 활용
- aiosqlite를 통한 비동기 DB 작업
- WebSocket 연결 풀 관리
- 스트리밍 응답을 위한 AsyncGenerator

### 의존성 주입

```python
# 예시: 의존성 주입 패턴
async def get_session_service(
    db: AsyncSession = Depends(get_db)
) -> SessionService:
    return SessionService(db)

@router.get("/sessions")
async def list_sessions(
    service: SessionService = Depends(get_session_service)
):
    return await service.list_all()
```

---

## 리스크 분석

### 기술적 리스크

| 리스크 | 영향도 | 대응 방안 |
|--------|--------|-----------|
| WebSocket 연결 불안정 | 높음 | 자동 재연결 로직, 연결 상태 모니터링 |
| Claude API 레이트 리밋 | 중간 | 요청 큐잉, 백오프 전략 |
| SQLite 동시성 제한 | 낮음 | WAL 모드 활성화, 연결 풀 관리 |

### 보안 리스크

| 리스크 | 영향도 | 대응 방안 |
|--------|--------|-----------|
| API 키 노출 | 높음 | 환경 변수 사용, 응답 필터링 |
| 미인증 접근 | 중간 | 토큰 기반 인증, CORS 제한 |
| 입력 검증 부재 | 중간 | Pydantic 검증, 입력 제한 |

---

## 테스트 전략

### 단위 테스트

- 각 서비스 모듈에 대한 단위 테스트
- Pydantic 모델 검증 테스트
- Mock을 활용한 외부 의존성 격리

### 통합 테스트

- FastAPI TestClient를 활용한 API 테스트
- WebSocket 연결 테스트
- 데이터베이스 트랜잭션 테스트

### E2E 테스트

- CLI에서 서버 시작 테스트
- 실제 Claude API 연동 테스트 (선택적)

---

## 마일스톤 체크리스트

- [ ] Phase 1: Primary Goals 완료
- [ ] Phase 2: Secondary Goals 완료
- [ ] Phase 3: Tertiary Goals 완료
- [ ] Phase 4: Final Goals 완료
- [ ] 테스트 커버리지 85% 이상 달성
- [ ] 문서화 완료

---

## 참조

- **SPEC 문서**: `.moai/specs/SPEC-WEB-001/spec.md`
- **인수 조건**: `.moai/specs/SPEC-WEB-001/acceptance.md`
