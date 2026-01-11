# SPEC-WEB-002: 구현 계획

## 메타데이터

| 항목 | 값 |
|------|-----|
| SPEC ID | SPEC-WEB-002 |
| 제목 | MoAI Web Frontend - React 대시보드 |
| 관련 문서 | [spec.md](./spec.md), [acceptance.md](./acceptance.md) |

---

## 1. 구현 마일스톤

### Primary Goal (1차 목표)

프로젝트 초기화 및 기본 구조 구축

**태스크**:
1. Vite + React 19 + TypeScript 프로젝트 생성
2. Tailwind CSS 4.x 설정
3. shadcn/ui 초기화 및 기본 컴포넌트 설치
4. 디렉토리 구조 생성
5. ESLint + Prettier 설정
6. 기본 라우팅 설정 (React Router 또는 TanStack Router)

**완료 기준**:
- `pnpm dev`로 개발 서버 실행 가능
- 기본 페이지 라우팅 동작
- 다크 모드 토글 동작

### Secondary Goal (2차 목표)

레이아웃 및 상태 관리 구현

**태스크**:
1. AppShell 3-패널 레이아웃 구현
2. Sidebar 컴포넌트 구현
3. Header 컴포넌트 구현
4. Zustand 스토어 설정 (appStore, chatStore)
5. 반응형 브레이크포인트 적용
6. 테마 시스템 구현

**완료 기준**:
- 3-패널 레이아웃 렌더링
- 사이드바 접기/펼치기 동작
- 테마 전환 동작
- 상태 관리 정상 동작

### Tertiary Goal (3차 목표)

채팅 인터페이스 구현

**태스크**:
1. ChatContainer 컴포넌트 구현
2. MessageList 컴포넌트 구현 (가상 스크롤)
3. MessageItem 컴포넌트 구현 (마크다운 렌더링)
4. CodeBlock 컴포넌트 구현 (문법 하이라이팅)
5. MessageInput 컴포넌트 구현
6. useChat 훅 구현

**완료 기준**:
- 메시지 목록 렌더링
- 마크다운 및 코드 블록 렌더링
- 메시지 입력 및 전송 UI 동작
- 자동 스크롤 동작

### Quaternary Goal (4차 목표)

WebSocket 연동 및 실시간 기능

**태스크**:
1. WebSocket 클라이언트 구현
2. useWebSocket 훅 구현
3. 채팅 스트리밍 연동
4. 연결 상태 관리
5. 재연결 로직 구현 (지수 백오프)
6. 에러 처리 및 토스트 알림

**완료 기준**:
- WebSocket 연결/해제 동작
- 실시간 메시지 수신
- 스트리밍 타이핑 효과
- 연결 끊김 시 자동 재연결

### Quinary Goal (5차 목표)

SPEC 모니터 및 Provider 기능

**태스크**:
1. SpecMonitor 컴포넌트 구현
2. SpecCard 컴포넌트 구현
3. SpecProgress 컴포넌트 구현
4. useSpecs 훅 구현
5. ProviderSwitch 컴포넌트 구현
6. ModelSelector 컴포넌트 구현

**완료 기준**:
- SPEC 목록 표시
- 실시간 SPEC 상태 업데이트
- Provider/Model 전환 동작
- 헤더에 현재 모델 표시

### Final Goal (최종 목표)

비용 추적 및 최적화

**태스크**:
1. CostTracker 컴포넌트 구현
2. CostBreakdown 컴포넌트 구현
3. TanStack Query 설정 및 API 연동
4. 성능 최적화 (React.memo, useMemo)
5. 번들 사이즈 최적화
6. 접근성 검사 및 개선

**완료 기준**:
- 비용 실시간 표시
- 모델별 비용 분석
- Lighthouse 성능 점수 90+
- WCAG 2.1 AA 준수

---

## 2. 기술적 접근

### 2.1 프로젝트 초기화

```bash
# 프로젝트 생성
pnpm create vite@latest web-ui --template react-ts

# 의존성 설치
pnpm add react@19 react-dom@19
pnpm add zustand @tanstack/react-query
pnpm add react-router-dom
pnpm add react-markdown remark-gfm
pnpm add shiki # 코드 하이라이팅

# shadcn/ui 설정
pnpm dlx shadcn@latest init
pnpm dlx shadcn@latest add button card input toast scroll-area

# 개발 의존성
pnpm add -D tailwindcss@4 @tailwindcss/vite
pnpm add -D vitest @testing-library/react @playwright/test
pnpm add -D @types/react @types/react-dom
```

### 2.2 Vite 설정

```typescript
// vite.config.ts
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import tailwindcss from '@tailwindcss/vite';
import path from 'path';

export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    port: 3000,
    proxy: {
      '/api': 'http://localhost:8000',
      '/ws': {
        target: 'ws://localhost:8000',
        ws: true,
      },
    },
  },
});
```

### 2.3 상태 관리 아키텍처

```
┌─────────────────────────────────────────────────────┐
│                    Components                        │
├──────────────┬──────────────┬──────────────────────┤
│   appStore   │  chatStore   │  TanStack Query      │
│   (Zustand)  │  (Zustand)   │  (Server State)      │
├──────────────┴──────────────┴──────────────────────┤
│                  WebSocket Layer                     │
├─────────────────────────────────────────────────────┤
│                  REST API Layer                      │
└─────────────────────────────────────────────────────┘
```

- **Zustand**: 클라이언트 상태 (UI, 연결, 스트리밍)
- **TanStack Query**: 서버 상태 (대화 목록, SPEC 목록, 설정)
- **WebSocket**: 실시간 상태 (채팅 스트리밍, SPEC 업데이트)

### 2.4 컴포넌트 설계 원칙

1. **Compound Components**: 관련 컴포넌트 그룹화
2. **Render Props / Hooks**: 로직과 UI 분리
3. **Controlled/Uncontrolled**: 폼 상태 일관성
4. **Error Boundaries**: 에러 격리
5. **Suspense**: 로딩 상태 처리

### 2.5 WebSocket 전략

```typescript
// 연결 관리
const wsManager = {
  connect: () => void,
  disconnect: () => void,
  send: (message: WebSocketMessage) => void,
  subscribe: (handler: MessageHandler) => () => void,
};

// 재연결 전략 (지수 백오프)
const RECONNECT_DELAYS = [1000, 2000, 4000, 8000, 16000]; // ms
const MAX_RECONNECT_ATTEMPTS = 5;
```

---

## 3. 아키텍처 설계

### 3.1 컴포넌트 계층

```
App
├── ThemeProvider
│   └── QueryClientProvider
│       └── Router
│           ├── ChatPage
│           │   └── AppShell
│           │       ├── Sidebar
│           │       │   └── ConversationList
│           │       ├── Main
│           │       │   └── ChatContainer
│           │       │       ├── MessageList
│           │       │       │   └── MessageItem[]
│           │       │       │       └── CodeBlock
│           │       │       └── MessageInput
│           │       └── Aside
│           │           └── SpecMonitor
│           │               └── SpecCard[]
│           ├── SpecPage
│           └── SettingsPage
└── Toaster
```

### 3.2 데이터 흐름

```
User Action → Component → Store/Hook → WebSocket/API → Backend
     ↑                                                    │
     └────────────────── State Update ←──────────────────┘
```

### 3.3 파일 구조 상세

```
src/
├── components/
│   ├── layout/
│   │   ├── AppShell.tsx        # 3-패널 레이아웃
│   │   ├── Sidebar.tsx         # 좌측 사이드바
│   │   ├── Header.tsx          # 상단 헤더
│   │   └── index.ts            # 배럴 익스포트
│   ├── chat/
│   │   ├── ChatContainer.tsx   # 채팅 컨테이너
│   │   ├── MessageList.tsx     # 메시지 목록
│   │   ├── MessageItem.tsx     # 개별 메시지
│   │   ├── CodeBlock.tsx       # 코드 블록
│   │   ├── MessageInput.tsx    # 입력 영역
│   │   └── index.ts
│   ├── spec/
│   │   ├── SpecMonitor.tsx     # SPEC 모니터
│   │   ├── SpecCard.tsx        # SPEC 카드
│   │   ├── SpecProgress.tsx    # 진행률 표시
│   │   └── index.ts
│   ├── provider/
│   │   ├── ProviderSwitch.tsx  # Provider 전환
│   │   ├── ModelSelector.tsx   # 모델 선택
│   │   └── index.ts
│   ├── cost/
│   │   ├── CostTracker.tsx     # 비용 추적
│   │   ├── CostBreakdown.tsx   # 비용 분석
│   │   └── index.ts
│   └── ui/                     # shadcn/ui 컴포넌트
├── hooks/
│   ├── useWebSocket.ts         # WebSocket 관리
│   ├── useChat.ts              # 채팅 로직
│   ├── useSpecs.ts             # SPEC 관리
│   └── useTheme.ts             # 테마 관리
├── stores/
│   ├── appStore.ts             # 앱 전역 상태
│   ├── chatStore.ts            # 채팅 상태
│   └── index.ts
├── lib/
│   ├── api.ts                  # REST API 클라이언트
│   ├── websocket.ts            # WebSocket 클라이언트
│   └── utils.ts                # 유틸리티 함수
├── types/
│   ├── chat.ts                 # 채팅 관련 타입
│   ├── spec.ts                 # SPEC 관련 타입
│   ├── api.ts                  # API 관련 타입
│   └── index.ts
├── pages/
│   ├── ChatPage.tsx            # 채팅 페이지
│   ├── SpecPage.tsx            # SPEC 페이지
│   └── SettingsPage.tsx        # 설정 페이지
├── App.tsx                     # 루트 컴포넌트
├── main.tsx                    # 엔트리 포인트
└── index.css                   # 글로벌 스타일
```

---

## 4. 리스크 및 대응

### 4.1 기술적 리스크

| 리스크 | 영향도 | 발생 확률 | 대응 전략 |
|--------|--------|----------|----------|
| React 19 호환성 이슈 | HIGH | MEDIUM | React 18.3 폴백 준비 |
| WebSocket 연결 불안정 | HIGH | LOW | 재연결 로직, 폴링 폴백 |
| shadcn/ui 버전 충돌 | MEDIUM | LOW | 버전 고정, 대체 컴포넌트 |
| 대량 메시지 성능 저하 | MEDIUM | MEDIUM | 가상 스크롤, 페이지네이션 |

### 4.2 의존성 리스크

| 의존성 | 리스크 | 대응 |
|--------|--------|------|
| SPEC-WEB-001 백엔드 | 지연 시 프론트엔드 테스트 불가 | Mock 서버 구축 |
| 외부 라이브러리 | 보안 취약점 | npm audit 자동화 |

### 4.3 완화 전략

1. **Mock 서버**: MSW (Mock Service Worker)로 백엔드 독립 개발
2. **Feature Flags**: 불안정 기능 점진적 롤아웃
3. **Error Boundaries**: 컴포넌트별 에러 격리
4. **Monitoring**: 클라이언트 에러 로깅 (Sentry 등)

---

## 5. 개발 환경

### 5.1 필수 도구

- Node.js 22.x LTS
- pnpm 9.x
- VS Code + 확장
  - ESLint
  - Prettier
  - Tailwind CSS IntelliSense
  - TypeScript Vue Plugin (Volar)

### 5.2 스크립트

```json
{
  "scripts": {
    "dev": "vite",
    "build": "tsc && vite build",
    "preview": "vite preview",
    "test": "vitest",
    "test:e2e": "playwright test",
    "lint": "eslint src --ext ts,tsx",
    "format": "prettier --write src"
  }
}
```

### 5.3 환경 변수

```bash
# .env.development
VITE_API_URL=http://localhost:8000
VITE_WS_URL=ws://localhost:8000

# .env.production
VITE_API_URL=/api
VITE_WS_URL=/ws
```

---

## 6. 추적성 태그

- **SPEC**: SPEC-WEB-002
- **의존**: SPEC-WEB-001
- **연관**: [spec.md](./spec.md), [acceptance.md](./acceptance.md)

---

**문서 버전**: 1.0.0
**최종 수정**: 2026-01-09
**작성자**: workflow-spec agent
