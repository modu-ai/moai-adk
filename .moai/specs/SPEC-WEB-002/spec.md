# SPEC-WEB-002: MoAI Web Frontend - React 대시보드

## 메타데이터

| 항목 | 값 |
|------|-----|
| SPEC ID | SPEC-WEB-002 |
| 제목 | MoAI Web Frontend - React 대시보드 |
| 상태 | Planned |
| 우선순위 | HIGH |
| 생성일 | 2026-01-09 |
| 의존성 | SPEC-WEB-001 |
| 라이프사이클 | spec-anchored |

---

## 1. Environment (환경)

### 1.1 기술 스택

```yaml
Frontend:
  React: "19.x"
  Vite: "6.x"
  TypeScript: "5.9+"
  shadcn/ui: "latest"
  Tailwind CSS: "4.x"
  Zustand: "5.x"
  TanStack Query: "5.x"

Communication:
  WebSocket: "native"
  HTTP Client: "ky" 또는 "fetch"

Development:
  Node.js: "22.x LTS"
  pnpm: "9.x"
  Vitest: "3.x"
  Playwright: "latest"
```

### 1.2 프로젝트 구조

```
web-ui/
├── src/
│   ├── components/
│   │   ├── layout/
│   │   │   ├── AppShell.tsx
│   │   │   ├── Sidebar.tsx
│   │   │   └── Header.tsx
│   │   ├── chat/
│   │   │   ├── ChatContainer.tsx
│   │   │   ├── MessageList.tsx
│   │   │   ├── MessageItem.tsx
│   │   │   ├── CodeBlock.tsx
│   │   │   └── MessageInput.tsx
│   │   ├── spec/
│   │   │   ├── SpecMonitor.tsx
│   │   │   ├── SpecCard.tsx
│   │   │   └── SpecProgress.tsx
│   │   ├── provider/
│   │   │   ├── ProviderSwitch.tsx
│   │   │   └── ModelSelector.tsx
│   │   └── cost/
│   │       ├── CostTracker.tsx
│   │       └── CostBreakdown.tsx
│   ├── hooks/
│   │   ├── useWebSocket.ts
│   │   ├── useChat.ts
│   │   └── useSpecs.ts
│   ├── stores/
│   │   ├── appStore.ts
│   │   └── chatStore.ts
│   ├── pages/
│   │   ├── ChatPage.tsx
│   │   ├── SpecPage.tsx
│   │   └── SettingsPage.tsx
│   ├── lib/
│   │   ├── api.ts
│   │   └── websocket.ts
│   ├── types/
│   │   └── index.ts
│   └── App.tsx
├── public/
├── index.html
├── vite.config.ts
├── tailwind.config.ts
├── tsconfig.json
└── package.json
```

### 1.3 외부 의존성

- SPEC-WEB-001 FastAPI 백엔드 서버
- Claude API (via Backend Proxy)
- GLM API (via Backend Proxy)

---

## 2. Assumptions (가정)

### 2.1 기술적 가정

| ID | 가정 | 신뢰도 | 근거 | 실패 시 영향 | 검증 방법 |
|----|------|--------|------|-------------|-----------|
| A1 | SPEC-WEB-001 백엔드가 WebSocket 엔드포인트 제공 | HIGH | SPEC-WEB-001 명세 확인 | WebSocket 연동 불가 | 백엔드 API 테스트 |
| A2 | React 19 Server Components는 Vite에서 미지원 | HIGH | Vite 공식 문서 | CSR 전용 구현 | Vite 설정 확인 |
| A3 | 브라우저 WebSocket API 지원 | HIGH | 모든 최신 브라우저 지원 | 폴백 필요 | caniuse.com 확인 |
| A4 | shadcn/ui가 React 19와 호환 | MEDIUM | 커뮤니티 보고 | UI 라이브러리 변경 | 실제 설치 테스트 |

### 2.2 비즈니스 가정

| ID | 가정 | 신뢰도 | 근거 | 실패 시 영향 |
|----|------|--------|------|-------------|
| B1 | 단일 사용자 환경 (인증 불필요) | HIGH | MVP 요구사항 | 인증 시스템 추가 |
| B2 | 데스크톱 우선 (모바일 후순위) | HIGH | 개발자 도구 특성 | 반응형 범위 축소 |
| B3 | 영어/한국어 2개 언어만 지원 | MEDIUM | 초기 타겟 시장 | i18n 확장 필요 |

---

## 3. Requirements (요구사항)

### 3.1 Ubiquitous Requirements (시스템 전역)

| REQ-ID | 요구사항 | 검증 방법 |
|--------|---------|----------|
| U-001 | UI는 **항상** 반응형으로 동작해야 한다 (breakpoints: sm/md/lg/xl) | Playwright viewport 테스트 |
| U-002 | 시스템은 **항상** 다크 모드를 지원해야 한다 | 테마 토글 테스트 |
| U-003 | 모든 인터랙티브 요소는 **항상** 키보드 접근 가능해야 한다 | a11y 감사 |
| U-004 | 에러 발생 시 **항상** 사용자에게 명확한 피드백을 제공해야 한다 | 에러 시나리오 테스트 |

### 3.2 Event-Driven Requirements (이벤트 기반)

| REQ-ID | 트리거 | 액션 | 검증 방법 |
|--------|--------|------|----------|
| E-001 | **WHEN** 채팅 입력이 제출되면 | **THEN** WebSocket으로 메시지 전송 | WebSocket mock 테스트 |
| E-002 | **WHEN** SPEC 상태가 변경되면 | **THEN** SpecMonitor 패널 실시간 업데이트 | WebSocket 이벤트 테스트 |
| E-003 | **WHEN** Provider가 변경되면 | **THEN** 헤더에 현재 모델명 표시 | UI 상태 테스트 |
| E-004 | **WHEN** WebSocket 연결이 끊어지면 | **THEN** 재연결 시도 (최대 5회, 지수 백오프) | 연결 복구 테스트 |
| E-005 | **WHEN** 새 메시지가 수신되면 | **THEN** MessageList 하단으로 자동 스크롤 | 스크롤 동작 테스트 |
| E-006 | **WHEN** 코드 블록 복사 버튼 클릭 | **THEN** 클립보드에 코드 복사 및 토스트 표시 | 복사 기능 테스트 |

### 3.3 State-Driven Requirements (상태 기반)

| REQ-ID | 조건 | 동작 | 검증 방법 |
|--------|------|------|----------|
| S-001 | **IF** 스트리밍 중이면 | **THEN** 타이핑 인디케이터 표시 | 스트리밍 상태 테스트 |
| S-002 | **IF** 에러 발생 시 | **THEN** 토스트 알림 표시 | 에러 핸들링 테스트 |
| S-003 | **IF** WebSocket 연결 중이면 | **THEN** 연결 상태 인디케이터(녹색) 표시 | 연결 상태 UI 테스트 |
| S-004 | **IF** 메시지가 비어있으면 | **THEN** 전송 버튼 비활성화 | 입력 유효성 테스트 |
| S-005 | **IF** SPEC 실행 중이면 | **THEN** 진행률 바 표시 | SPEC 상태 UI 테스트 |

### 3.4 Unwanted Requirements (금지 사항)

| REQ-ID | 금지 동작 | 이유 | 검증 방법 |
|--------|----------|------|----------|
| N-001 | API 키를 클라이언트에 노출**하지 않아야 한다** | 보안 | 번들 분석, 네트워크 탭 검사 |
| N-002 | 민감한 데이터를 localStorage에 저장**하지 않아야 한다** | 보안 | 스토리지 감사 |
| N-003 | 사용자 입력을 innerHTML로 렌더링**하지 않아야 한다** | XSS 방지 | 코드 리뷰 |
| N-004 | 백엔드 없이 직접 AI API 호출**하지 않아야 한다** | 아키텍처 | 네트워크 요청 검사 |

### 3.5 Optional Requirements (선택 사항)

| REQ-ID | 기능 | 우선순위 | 조건 |
|--------|------|---------|------|
| O-001 | **가능하면** 대화 내역 내보내기 제공 | LOW | 시간 여유 시 |
| O-002 | **가능하면** 단축키 지원 제공 (Cmd+Enter 전송) | MEDIUM | UX 개선 |
| O-003 | **가능하면** 마크다운 미리보기 제공 | LOW | 시간 여유 시 |

---

## 4. Specifications (상세 명세)

### 4.1 컴포넌트 명세

#### 4.1.1 Layout Components

**AppShell**
```typescript
interface AppShellProps {
  children: React.ReactNode;
}
// 3-패널 레이아웃: Sidebar(좌) | Main(중앙) | SpecMonitor(우)
// 반응형: lg 이하에서 Sidebar 접힘
```

**Header**
```typescript
interface HeaderProps {
  currentModel: string;
  isConnected: boolean;
}
// 표시: 로고, 현재 모델명, 연결 상태, 설정 버튼
```

**Sidebar**
```typescript
interface SidebarProps {
  conversations: Conversation[];
  activeId: string;
  onSelect: (id: string) => void;
}
// 대화 목록, 새 대화 버튼
```

#### 4.1.2 Chat Components

**ChatContainer**
```typescript
interface ChatContainerProps {
  conversationId: string;
}
// MessageList + MessageInput 조합
// WebSocket 연결 관리
```

**MessageList**
```typescript
interface MessageListProps {
  messages: Message[];
  isStreaming: boolean;
}
// 가상 스크롤 (대량 메시지 성능)
// 자동 스크롤
```

**MessageItem**
```typescript
interface MessageItemProps {
  message: Message;
  isStreaming?: boolean;
}
// 역할별 스타일 (user/assistant/system)
// 마크다운 렌더링
// 코드 블록 하이라이팅
```

**CodeBlock**
```typescript
interface CodeBlockProps {
  code: string;
  language: string;
}
// Shiki 또는 Prism 문법 하이라이팅
// 복사 버튼
// 언어 라벨
```

**MessageInput**
```typescript
interface MessageInputProps {
  onSend: (content: string) => void;
  disabled: boolean;
}
// 텍스트 영역 (자동 높이 조절)
// 전송 버튼
// Cmd+Enter 단축키
```

#### 4.1.3 SPEC Components

**SpecMonitor**
```typescript
interface SpecMonitorProps {
  specs: Spec[];
}
// SPEC 목록 표시
// 실시간 상태 업데이트
```

**SpecCard**
```typescript
interface SpecCardProps {
  spec: Spec;
  onClick: () => void;
}
// SPEC ID, 제목, 상태 배지
// 진행률 표시
```

**SpecProgress**
```typescript
interface SpecProgressProps {
  status: SpecStatus;
  progress: number; // 0-100
}
// 단계별 진행 표시
// RED/GREEN/REFACTOR 상태
```

#### 4.1.4 Provider Components

**ProviderSwitch**
```typescript
interface ProviderSwitchProps {
  providers: Provider[];
  current: string;
  onChange: (providerId: string) => void;
}
// Claude/GLM 전환
// 모델 선택 드롭다운
```

**ModelSelector**
```typescript
interface ModelSelectorProps {
  models: Model[];
  selected: string;
  onSelect: (modelId: string) => void;
}
// opus/sonnet/haiku 등 모델 선택
```

#### 4.1.5 Cost Components

**CostTracker**
```typescript
interface CostTrackerProps {
  totalCost: number;
  sessionCost: number;
}
// 현재 세션 비용
// 누적 비용
```

**CostBreakdown**
```typescript
interface CostBreakdownProps {
  breakdown: CostEntry[];
}
// 모델별 비용
// 입력/출력 토큰 분리
```

### 4.2 상태 관리 명세

#### 4.2.1 App Store (Zustand)

```typescript
interface AppState {
  // 연결 상태
  isConnected: boolean;
  connectionStatus: 'connecting' | 'connected' | 'disconnected' | 'error';

  // Provider 설정
  currentProvider: 'claude' | 'glm';
  currentModel: string;

  // UI 상태
  theme: 'light' | 'dark';
  sidebarOpen: boolean;

  // Actions
  setConnectionStatus: (status: ConnectionStatus) => void;
  setProvider: (provider: string, model: string) => void;
  toggleTheme: () => void;
  toggleSidebar: () => void;
}
```

#### 4.2.2 Chat Store (Zustand)

```typescript
interface ChatState {
  // 대화
  conversations: Map<string, Conversation>;
  activeConversationId: string | null;

  // 메시지
  messages: Map<string, Message[]>;
  streamingMessage: string | null;

  // Actions
  createConversation: () => string;
  setActiveConversation: (id: string) => void;
  addMessage: (conversationId: string, message: Message) => void;
  updateStreamingMessage: (content: string) => void;
  clearStreamingMessage: () => void;
}
```

### 4.3 WebSocket 명세

```typescript
interface WebSocketMessage {
  type: 'chat' | 'spec_update' | 'cost_update' | 'error';
  payload: unknown;
  timestamp: string;
}

interface ChatPayload {
  conversationId: string;
  role: 'user' | 'assistant';
  content: string;
  isStreaming: boolean;
}

interface SpecUpdatePayload {
  specId: string;
  status: SpecStatus;
  progress: number;
  message?: string;
}

interface CostUpdatePayload {
  inputTokens: number;
  outputTokens: number;
  cost: number;
  model: string;
}
```

### 4.4 API 통신 명세

```typescript
// REST API (비스트리밍 작업)
const api = {
  // Provider
  getProviders: () => GET('/api/providers'),
  setProvider: (data: SetProviderRequest) => POST('/api/provider', data),

  // Conversations
  getConversations: () => GET('/api/conversations'),
  createConversation: () => POST('/api/conversations'),
  deleteConversation: (id: string) => DELETE(`/api/conversations/${id}`),

  // SPECs
  getSpecs: () => GET('/api/specs'),
  getSpec: (id: string) => GET(`/api/specs/${id}`),

  // Settings
  getSettings: () => GET('/api/settings'),
  updateSettings: (data: Settings) => PUT('/api/settings', data),
};

// WebSocket (실시간 작업)
// ws://localhost:8000/ws/chat - 채팅 스트리밍
// ws://localhost:8000/ws/specs - SPEC 상태 업데이트
```

---

## 5. Traceability (추적성)

### 5.1 요구사항-컴포넌트 매핑

| 요구사항 | 관련 컴포넌트 | 테스트 파일 |
|---------|-------------|------------|
| U-001 | AppShell, Sidebar | layout.spec.ts |
| U-002 | AppShell, Header | theme.spec.ts |
| E-001 | MessageInput, useChat | chat.spec.ts |
| E-002 | SpecMonitor, useSpecs | spec.spec.ts |
| E-003 | ProviderSwitch, Header | provider.spec.ts |
| S-001 | MessageList, MessageItem | streaming.spec.ts |
| S-002 | Toast (shadcn/ui) | error.spec.ts |
| N-001 | api.ts, 환경 변수 | security.spec.ts |

### 5.2 관련 SPEC

- **SPEC-WEB-001**: FastAPI 백엔드 (의존)
- **SPEC-WEB-003**: E2E 테스트 (후속, 계획)

---

## 6. 참고 자료

### 6.1 기술 문서

- [React 19 Documentation](https://react.dev)
- [Vite 6 Guide](https://vite.dev)
- [shadcn/ui Components](https://ui.shadcn.com)
- [Zustand Documentation](https://zustand-demo.pmnd.rs)
- [TanStack Query](https://tanstack.com/query)

### 6.2 디자인 참조

- Claude Code CLI 인터페이스
- VS Code 채팅 패널
- GitHub Copilot Chat

---

**문서 버전**: 1.0.0
**최종 수정**: 2026-01-09
**작성자**: workflow-spec agent
