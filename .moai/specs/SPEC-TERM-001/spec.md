---
id: SPEC-TERM-001
version: "1.0.0"
status: "draft"
created: "2026-01-09"
updated: "2026-01-09"
author: "MoAI-ADK"
priority: "medium"
depends_on:
  - SPEC-WEB-001
---

# SPEC-TERM-001: Parallel Terminal System - xterm.js + PTY

xterm.js와 PTY 프로세스를 활용한 병렬 터미널 시스템

## HISTORY

| 버전 | 날짜 | 작성자 | 변경 내용 |
|------|------|--------|-----------|
| 1.0.0 | 2026-01-09 | MoAI-ADK | 초기 SPEC 작성 |

---

## 개요

### 목적

MoAI Web UI에서 실제 터미널 환경을 제공하기 위해 xterm.js 프론트엔드와 PTY(Pseudo Terminal) 백엔드를 통합한 병렬 터미널 시스템을 구현한다. 사용자는 최대 6개의 터미널을 동시에 실행하고 관리할 수 있다.

### 범위

- PTY 프로세스 생성 및 관리
- WebSocket 기반 양방향 통신
- 다중 터미널 탭 인터페이스
- 터미널 자동 명령 실행
- 터미널 세션 상태 관리
- 터미널 크기 자동 조정

---

## Environment (환경)

### 기술 스택

```yaml
Backend:
  - ptyprocess: ">=0.7.0"
  - asyncio subprocess
  - WebSocket per terminal

Frontend:
  - xterm.js: "5.3+"
  - @xterm/addon-fit: "0.10+"
  - @xterm/addon-web-links: "0.11+"
  - @xterm/addon-serialize: "0.13+"
```

### 시스템 요구사항

- Python: 3.11 이상
- Node.js: 18 이상
- OS: macOS, Linux (Windows WSL 지원)
- 네트워크: WebSocket 지원

---

## Assumptions (가정)

### 기술적 가정

1. SPEC-WEB-001의 FastAPI 서버가 실행 중임
2. ptyprocess가 운영체제의 PTY 기능을 지원함
3. 클라이언트 브라우저가 WebSocket을 지원함
4. 터미널 명령 실행 권한이 사용자에게 있음

### 비즈니스 가정

1. 사용자는 브라우저에서 터미널 작업을 수행함
2. 다중 터미널이 병렬 작업에 필수적임
3. 자동 명령 실행으로 개발 효율성 향상

---

## Requirements (요구사항)

### Ubiquitous (항상 활성)

| ID | 요구사항 | 검증 방법 |
|----|----------|-----------|
| R-UBI-001 | 시스템은 항상 터미널 출력을 실시간으로 렌더링해야 한다 | 출력 지연 시간 측정 |
| R-UBI-002 | 시스템은 항상 터미널 입력을 PTY에 즉시 전달해야 한다 | 입력 응답 시간 측정 |
| R-UBI-003 | 시스템은 항상 활성 터미널 상태를 UI에 표시해야 한다 | 상태 표시기 확인 |

### Event-Driven (이벤트 기반)

| ID | 트리거 | 동작 | 검증 방법 |
|----|--------|------|-----------|
| R-EVT-001 | WHEN 새 터미널 생성 요청 | THEN PTY 프로세스 생성 + WebSocket 수립 | 터미널 창 표시 확인 |
| R-EVT-002 | WHEN 터미널 입력 수신 | THEN PTY에 데이터 전달 | 명령 실행 확인 |
| R-EVT-003 | WHEN PTY 출력 발생 | THEN xterm.js에 실시간 렌더링 | 화면 출력 확인 |
| R-EVT-004 | WHEN 터미널 종료 요청 | THEN PTY 프로세스 정리 + 리소스 해제 | 프로세스 종료 확인 |
| R-EVT-005 | WHEN 터미널 크기 변경 | THEN PTY 윈도우 크기 업데이트 | SIGWINCH 전달 확인 |
| R-EVT-006 | WHEN 자동 명령 요청 | THEN 지정된 터미널에 명령 주입 | 명령 실행 로그 확인 |

### State-Driven (상태 기반)

| ID | 조건 | 동작 | 검증 방법 |
|----|------|------|-----------|
| R-STA-001 | IF 다중 터미널 활성 | THEN 탭 인터페이스로 전환 | 탭 UI 표시 확인 |
| R-STA-002 | IF PTY 비활성 30분 | THEN 자동 종료 | 타임아웃 후 상태 확인 |
| R-STA-003 | IF 터미널 수 6개 도달 | THEN 새 터미널 생성 차단 | 생성 버튼 비활성화 확인 |

### Unwanted (금지 사항)

| ID | 금지 동작 | 이유 | 검증 방법 |
|----|----------|------|-----------|
| R-UNW-001 | 시스템은 비활성 PTY를 30분 이상 유지하지 않아야 한다 | 리소스 관리 | 메모리 사용량 모니터링 |
| R-UNW-002 | 시스템은 6개 초과 터미널을 생성하지 않아야 한다 | 시스템 안정성 | 터미널 개수 검증 |
| R-UNW-003 | 시스템은 PTY 종료 시 좀비 프로세스를 남기지 않아야 한다 | 리소스 누수 방지 | 프로세스 목록 확인 |

### Optional (선택 사항)

| ID | 요구사항 | 조건 |
|----|----------|------|
| R-OPT-001 | 가능하면 터미널 세션 복원 기능 제공 | 브라우저 새로고침 시 |
| R-OPT-002 | 가능하면 터미널 출력 검색 기능 제공 | 사용자 요청 시 |
| R-OPT-003 | 가능하면 터미널 출력 스크롤백 1만 줄 제공 | 기본 설정 |

---

## Specifications (상세 명세)

### API Endpoints

| Method | Path | 설명 |
|--------|------|------|
| POST | /api/terminals | 새 터미널 세션 생성 |
| GET | /api/terminals | 터미널 목록 조회 |
| GET | /api/terminals/{id} | 터미널 상태 조회 |
| DELETE | /api/terminals/{id} | 터미널 종료 |
| POST | /api/terminals/{id}/resize | 터미널 크기 조정 |
| POST | /api/terminals/{id}/command | 자동 명령 실행 |
| WS | /ws/terminal/{id} | WebSocket 터미널 연결 |

### 백엔드 파일 구조

```
src/moai_adk/web/
├── services/
│   └── terminal_service.py    # TerminalManager class
└── routers/
    └── terminal.py            # /ws/terminal/{id} 엔드포인트
```

### 프론트엔드 파일 구조

```
web-ui/src/components/terminal/
├── TerminalPane.tsx           # 단일 터미널 컴포넌트
├── ParallelTerminals.tsx      # 터미널 그리드 레이아웃
├── TerminalTabs.tsx           # 탭 네비게이션
└── TerminalStatusBar.tsx      # 상태 표시기
```

### 데이터 모델

#### Terminal 모델

```python
class Terminal(BaseModel):
    id: str                    # UUID
    spec_id: str | None        # 연결된 SPEC ID
    created_at: datetime
    last_activity: datetime
    status: Literal["active", "idle", "terminated"]
    rows: int = 24
    cols: int = 80
    shell: str = "/bin/bash"
```

#### TerminalCreate 모델

```python
class TerminalCreate(BaseModel):
    spec_id: str | None = None
    shell: str = "/bin/bash"
    rows: int = 24
    cols: int = 80
    auto_command: str | None = None
```

#### TerminalResize 모델

```python
class TerminalResize(BaseModel):
    rows: int
    cols: int
```

### WebSocket 메시지 프로토콜

#### 클라이언트 -> 서버

```json
{
  "type": "input",
  "data": "키보드 입력 데이터"
}
```

```json
{
  "type": "resize",
  "rows": 24,
  "cols": 80
}
```

#### 서버 -> 클라이언트

```json
{
  "type": "output",
  "data": "터미널 출력 데이터"
}
```

```json
{
  "type": "exit",
  "code": 0
}
```

```json
{
  "type": "error",
  "message": "에러 메시지"
}
```

### TerminalManager 클래스 설계

```python
class TerminalManager:
    """PTY 프로세스 및 WebSocket 연결 관리"""

    def __init__(self, max_terminals: int = 6):
        self.terminals: dict[str, TerminalSession] = {}
        self.max_terminals = max_terminals

    async def create_terminal(
        self,
        spec_id: str | None = None,
        shell: str = "/bin/bash",
        rows: int = 24,
        cols: int = 80,
    ) -> Terminal:
        """새 PTY 프로세스 생성"""
        pass

    async def destroy_terminal(self, terminal_id: str) -> bool:
        """PTY 프로세스 종료 및 리소스 해제"""
        pass

    async def resize_terminal(
        self,
        terminal_id: str,
        rows: int,
        cols: int
    ) -> bool:
        """터미널 크기 조정 (SIGWINCH)"""
        pass

    async def write_to_terminal(
        self,
        terminal_id: str,
        data: str
    ) -> bool:
        """PTY에 데이터 쓰기"""
        pass

    async def execute_command(
        self,
        terminal_id: str,
        command: str
    ) -> bool:
        """자동 명령 실행 (예: claude /moai:2-run)"""
        pass

    async def cleanup_inactive(
        self,
        timeout_minutes: int = 30
    ) -> list[str]:
        """비활성 터미널 정리"""
        pass
```

### 프론트엔드 컴포넌트 인터페이스

#### TerminalPane Props

```typescript
interface TerminalPaneProps {
  terminalId: string;
  specId?: string;
  onClose?: () => void;
  onFocus?: () => void;
  autoCommand?: string;
}
```

#### ParallelTerminals Props

```typescript
interface ParallelTerminalsProps {
  maxTerminals?: number;  // default: 6
  layout?: 'tabs' | 'grid' | 'split';
  specId?: string;
}
```

---

## Dependencies (의존성)

### 의존하는 SPEC

| SPEC ID | 제목 | 의존 내용 |
|---------|------|-----------|
| SPEC-WEB-001 | MoAI Web Backend | FastAPI 서버, WebSocket 인프라 |

### 이 SPEC에 의존하는 SPEC

| SPEC ID | 제목 | 의존 내용 |
|---------|------|-----------|
| SPEC-CMD-001 | Command Extension | 터미널 명령 실행 |
| SPEC-AGENT-001 | Agent Integration | 에이전트 자동 명령 |

---

## Traceability (추적성)

### 관련 문서

- `.moai/project/tech.md`: 기술 스택 정의
- `.moai/project/product.md`: 제품 요구사항
- `SPEC-WEB-001/spec.md`: 웹 백엔드 명세

### 태그

- `SPEC-TERM-001`
- `xterm.js`
- `pty`
- `websocket`
- `terminal`
- `parallel`
