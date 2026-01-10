---
id: SPEC-TERM-001
type: plan
version: "1.0.0"
created: "2026-01-09"
updated: "2026-01-09"
---

# SPEC-TERM-001: Implementation Plan

Parallel Terminal System 구현 계획

## TAG

`SPEC-TERM-001` | `xterm.js` | `pty` | `websocket` | `terminal`

---

## 마일스톤

### Primary Goal (최우선)

백엔드 PTY 서비스 구현

#### 작업 내역

1. **TerminalManager 클래스 구현**
   - PTY 프로세스 생성/종료
   - 비동기 읽기/쓰기
   - 세션 상태 관리
   - 파일: `src/moai_adk/web/services/terminal_service.py`

2. **Terminal 라우터 구현**
   - REST API 엔드포인트
   - WebSocket 핸들러
   - 파일: `src/moai_adk/web/routers/terminal.py`

3. **Terminal 모델 정의**
   - Pydantic 모델
   - 파일: `src/moai_adk/web/models/terminal.py`

#### 완료 조건

- PTY 프로세스 생성/종료 정상 동작
- WebSocket 통신 양방향 확인
- 단위 테스트 통과 (커버리지 85% 이상)

---

### Secondary Goal (2차)

프론트엔드 터미널 컴포넌트 구현

#### 작업 내역

1. **TerminalPane 컴포넌트**
   - xterm.js 초기화
   - WebSocket 연결 관리
   - 입출력 처리
   - 파일: `web-ui/src/components/terminal/TerminalPane.tsx`

2. **ParallelTerminals 컴포넌트**
   - 다중 터미널 레이아웃
   - 터미널 생성/삭제
   - 파일: `web-ui/src/components/terminal/ParallelTerminals.tsx`

3. **TerminalTabs 컴포넌트**
   - 탭 네비게이션
   - 활성 터미널 전환
   - 파일: `web-ui/src/components/terminal/TerminalTabs.tsx`

4. **TerminalStatusBar 컴포넌트**
   - 연결 상태 표시
   - 터미널 정보 표시
   - 파일: `web-ui/src/components/terminal/TerminalStatusBar.tsx`

#### 완료 조건

- xterm.js 렌더링 정상
- 키보드 입력 전달 확인
- 탭 전환 동작 확인

---

### Final Goal (최종)

통합 및 최적화

#### 작업 내역

1. **자동 명령 실행 통합**
   - `claude /moai:2-run` 자동 실행
   - SPEC ID 기반 터미널 연결
   - 명령 결과 피드백

2. **비활성 터미널 정리**
   - 30분 타임아웃 구현
   - 백그라운드 정리 태스크
   - 리소스 모니터링

3. **성능 최적화**
   - 출력 버퍼링 최적화
   - 메모리 사용량 최적화
   - WebSocket 재연결 처리

#### 완료 조건

- 통합 테스트 통과
- 6개 동시 터미널 안정 동작
- 메모리 누수 없음

---

### Optional Goal (선택)

고급 기능 구현

#### 작업 내역

1. **세션 복원**
   - 터미널 상태 직렬화
   - 브라우저 새로고침 복원

2. **출력 검색**
   - 스크롤백 버퍼 검색
   - 검색 UI

3. **테마 지원**
   - 다크/라이트 테마
   - 커스텀 색상 설정

---

## Technical Approach (기술적 접근)

### 백엔드 아키텍처

```
[FastAPI Server]
    │
    ├── [REST API]
    │     └── Terminal CRUD
    │
    └── [WebSocket Handler]
          │
          └── [TerminalManager]
                │
                ├── [TerminalSession 1]
                │     └── [ptyprocess]
                │
                ├── [TerminalSession 2]
                │     └── [ptyprocess]
                │
                └── [TerminalSession N]
                      └── [ptyprocess]
```

### 프론트엔드 아키텍처

```
[ParallelTerminals]
    │
    ├── [TerminalTabs]
    │     └── Tab Navigation
    │
    └── [TerminalPane] x N
          │
          ├── [xterm.js Terminal]
          │
          ├── [FitAddon]
          │
          └── [WebLinksAddon]
```

### 데이터 흐름

```
User Input → xterm.js → WebSocket → FastAPI → ptyprocess → Shell
                                                    ↓
User Screen ← xterm.js ← WebSocket ← FastAPI ← PTY Output
```

### 핵심 기술 결정

| 영역 | 선택 | 이유 |
|------|------|------|
| PTY 라이브러리 | ptyprocess | 크로스 플랫폼, 비동기 지원 |
| 터미널 UI | xterm.js 5.3+ | 산업 표준, 풍부한 애드온 |
| 통신 | WebSocket | 양방향 실시간 통신 |
| 상태 관리 | In-memory | 터미널 수 제한(6개)으로 충분 |

---

## Risks (리스크)

### 기술적 리스크

| 리스크 | 영향 | 완화 방안 |
|--------|------|-----------|
| PTY 플랫폼 호환성 | 높음 | Windows WSL 테스트, 폴백 구현 |
| WebSocket 연결 불안정 | 중간 | 재연결 로직, 상태 복구 |
| 메모리 누수 | 중간 | 명시적 리소스 해제, 모니터링 |

### 비즈니스 리스크

| 리스크 | 영향 | 완화 방안 |
|--------|------|-----------|
| 성능 요구사항 미충족 | 중간 | 조기 벤치마크, 점진적 최적화 |
| 사용자 경험 저하 | 중간 | 지연 시간 모니터링, UX 피드백 |

---

## Dependencies (의존성)

### 선행 작업

| SPEC/작업 | 상태 | 필요 시점 |
|-----------|------|-----------|
| SPEC-WEB-001 | Draft | Primary Goal 시작 전 |

### 외부 의존성

| 라이브러리 | 버전 | 용도 |
|------------|------|------|
| ptyprocess | >=0.7.0 | PTY 프로세스 관리 |
| xterm.js | 5.3+ | 터미널 UI 렌더링 |
| @xterm/addon-fit | 0.10+ | 터미널 크기 자동 조정 |
| @xterm/addon-web-links | 0.11+ | 링크 클릭 지원 |
| @xterm/addon-serialize | 0.13+ | 터미널 상태 직렬화 |

---

## Traceability (추적성)

### 요구사항 매핑

| 요구사항 ID | 마일스톤 | 구현 파일 |
|-------------|----------|-----------|
| R-EVT-001 | Primary | terminal_service.py |
| R-EVT-002 | Primary | terminal.py (router) |
| R-EVT-003 | Secondary | TerminalPane.tsx |
| R-EVT-004 | Primary | terminal_service.py |
| R-STA-001 | Secondary | ParallelTerminals.tsx |
| R-STA-002 | Final | terminal_service.py |
| R-UNW-001 | Final | terminal_service.py |

### 관련 SPEC

- `SPEC-WEB-001`: 의존 (FastAPI 서버)
- `SPEC-CMD-001`: 피의존 (명령 실행)
