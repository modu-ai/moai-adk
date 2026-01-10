# SPEC-WEB-002: 인수 조건

## 메타데이터

| 항목 | 값 |
|------|-----|
| SPEC ID | SPEC-WEB-002 |
| 제목 | MoAI Web Frontend - React 대시보드 |
| 관련 문서 | [spec.md](./spec.md), [plan.md](./plan.md) |

---

## 1. 기능 테스트 시나리오

### 1.1 대시보드 레이아웃 (AC-LAYOUT)

#### AC-LAYOUT-001: 3-패널 레이아웃 표시

```gherkin
Feature: 대시보드 3-패널 레이아웃
  대시보드는 Sidebar, Main, SpecMonitor 3개 패널로 구성된다

  Scenario: 대시보드 로드 시 3-패널 레이아웃 표시
    Given 사용자가 대시보드 URL에 접속한다
    When 페이지가 완전히 로드된다
    Then 좌측에 Sidebar 패널이 표시된다
    And 중앙에 Main 채팅 패널이 표시된다
    And 우측에 SpecMonitor 패널이 표시된다
    And Header에 로고와 현재 모델명이 표시된다

  Scenario: 반응형 레이아웃 - 태블릿
    Given 사용자가 태블릿 크기 화면(768px)에서 접속한다
    When 페이지가 로드된다
    Then Sidebar는 접힌 상태로 표시된다
    And Main 패널이 전체 너비로 확장된다
    And SpecMonitor는 하단 드로어로 변경된다

  Scenario: 반응형 레이아웃 - 모바일
    Given 사용자가 모바일 크기 화면(375px)에서 접속한다
    When 페이지가 로드된다
    Then Sidebar는 햄버거 메뉴로 숨겨진다
    And Main 패널만 표시된다
    And SpecMonitor는 탭으로 접근 가능하다
```

#### AC-LAYOUT-002: 사이드바 토글

```gherkin
Feature: 사이드바 접기/펼치기
  사용자는 사이드바를 토글할 수 있다

  Scenario: 사이드바 접기
    Given 사이드바가 펼쳐진 상태이다
    When 사용자가 사이드바 접기 버튼을 클릭한다
    Then 사이드바가 아이콘만 표시되는 축소 상태로 변경된다
    And Main 패널이 확장된다

  Scenario: 사이드바 펼치기
    Given 사이드바가 접힌 상태이다
    When 사용자가 사이드바 펼치기 버튼을 클릭한다
    Then 사이드바가 전체 너비로 펼쳐진다
    And 대화 목록이 표시된다
```

### 1.2 채팅 인터페이스 (AC-CHAT)

#### AC-CHAT-001: 메시지 전송

```gherkin
Feature: 채팅 메시지 전송
  사용자는 채팅 메시지를 입력하고 전송할 수 있다

  Scenario: 텍스트 메시지 전송
    Given 사용자가 채팅 페이지에 있다
    And WebSocket이 연결된 상태이다
    When 사용자가 입력란에 "Hello, Claude!"를 입력한다
    And 전송 버튼을 클릭한다
    Then 메시지가 MessageList에 추가된다
    And 입력란이 비워진다
    And WebSocket으로 메시지가 전송된다

  Scenario: 단축키로 메시지 전송
    Given 사용자가 입력란에 메시지를 입력한 상태이다
    When 사용자가 Cmd+Enter(Mac) 또는 Ctrl+Enter(Windows)를 누른다
    Then 메시지가 전송된다

  Scenario: 빈 메시지 전송 방지
    Given 입력란이 비어있다
    Then 전송 버튼이 비활성화 상태이다
    When 사용자가 전송 버튼을 클릭한다
    Then 아무 동작도 발생하지 않는다
```

#### AC-CHAT-002: 스트리밍 응답 표시

```gherkin
Feature: AI 응답 스트리밍
  AI 응답은 실시간으로 스트리밍되어 표시된다

  Scenario: 스트리밍 응답 수신
    Given 사용자가 메시지를 전송한 상태이다
    When 서버로부터 스트리밍 응답이 시작된다
    Then 타이핑 인디케이터가 표시된다
    And 응답 텍스트가 점진적으로 표시된다
    And 메시지 목록이 하단으로 자동 스크롤된다

  Scenario: 스트리밍 완료
    Given AI 응답이 스트리밍 중이다
    When 스트리밍이 완료된다
    Then 타이핑 인디케이터가 사라진다
    And 완성된 메시지가 고정된다
    And 비용 정보가 업데이트된다
```

#### AC-CHAT-003: 코드 블록 렌더링

```gherkin
Feature: 코드 블록 표시
  코드 블록은 문법 하이라이팅과 함께 표시된다

  Scenario: Python 코드 블록 표시
    Given AI 응답에 Python 코드 블록이 포함되어 있다
    When 메시지가 렌더링된다
    Then 코드 블록이 별도 영역에 표시된다
    And Python 문법 하이라이팅이 적용된다
    And 언어 라벨 "python"이 표시된다
    And 복사 버튼이 표시된다

  Scenario: 코드 복사
    Given 코드 블록이 표시된 상태이다
    When 사용자가 복사 버튼을 클릭한다
    Then 코드가 클립보드에 복사된다
    And "복사됨" 토스트 알림이 표시된다
```

### 1.3 WebSocket 연결 (AC-WS)

#### AC-WS-001: 연결 상태 관리

```gherkin
Feature: WebSocket 연결 관리
  시스템은 WebSocket 연결 상태를 관리하고 표시한다

  Scenario: 초기 연결
    Given 사용자가 대시보드에 접속한다
    When 페이지가 로드된다
    Then WebSocket 연결이 시도된다
    And Header에 연결 상태 인디케이터가 표시된다

  Scenario: 연결 성공
    Given WebSocket 연결이 시도 중이다
    When 연결이 성공한다
    Then 인디케이터가 녹색으로 변경된다
    And "연결됨" 상태가 표시된다

  Scenario: 연결 끊김 및 재연결
    Given WebSocket이 연결된 상태이다
    When 네트워크 연결이 끊어진다
    Then 인디케이터가 노란색으로 변경된다
    And "재연결 중..." 상태가 표시된다
    And 지수 백오프로 재연결이 시도된다

  Scenario: 재연결 실패
    Given 재연결을 5회 시도했다
    When 모든 시도가 실패한다
    Then 인디케이터가 빨간색으로 변경된다
    And "연결 실패" 상태가 표시된다
    And "다시 시도" 버튼이 표시된다
```

### 1.4 SPEC 모니터 (AC-SPEC)

#### AC-SPEC-001: SPEC 목록 표시

```gherkin
Feature: SPEC 목록 표시
  SpecMonitor는 현재 프로젝트의 SPEC 목록을 표시한다

  Scenario: SPEC 목록 로드
    Given 사용자가 대시보드에 있다
    When SpecMonitor가 마운트된다
    Then API에서 SPEC 목록을 조회한다
    And 각 SPEC이 SpecCard로 표시된다
    And SPEC ID, 제목, 상태가 표시된다

  Scenario: 빈 SPEC 목록
    Given 프로젝트에 SPEC이 없다
    When SpecMonitor가 로드된다
    Then "등록된 SPEC이 없습니다" 메시지가 표시된다
```

#### AC-SPEC-002: SPEC 상태 실시간 업데이트

```gherkin
Feature: SPEC 상태 실시간 업데이트
  SPEC 상태 변경이 실시간으로 반영된다

  Scenario: SPEC 상태 변경 수신
    Given SPEC-001이 "In Progress" 상태로 표시되어 있다
    When WebSocket으로 상태 업데이트가 수신된다
    And 새 상태가 "Completed"이다
    Then SpecCard의 상태 배지가 "Completed"로 변경된다
    And 진행률이 100%로 표시된다

  Scenario: TDD 단계 업데이트
    Given SPEC-001이 실행 중이다
    When RED 단계에서 GREEN 단계로 전환된다
    Then SpecProgress에 GREEN 상태가 표시된다
    And 진행률 바가 업데이트된다
```

### 1.5 Provider 전환 (AC-PROVIDER)

#### AC-PROVIDER-001: Provider 변경

```gherkin
Feature: AI Provider 전환
  사용자는 Claude와 GLM 사이에서 Provider를 전환할 수 있다

  Scenario: Claude에서 GLM으로 전환
    Given 현재 Provider가 Claude이다
    And Header에 "Claude"가 표시되어 있다
    When 사용자가 ProviderSwitch에서 GLM을 선택한다
    Then Provider 변경 API가 호출된다
    And Header에 "GLM"이 표시된다
    And 이후 채팅은 GLM을 통해 처리된다

  Scenario: 모델 선택
    Given Provider가 Claude로 설정되어 있다
    When 사용자가 ModelSelector를 클릭한다
    Then 사용 가능한 모델 목록이 표시된다 (opus, sonnet, haiku)
    When 사용자가 "sonnet"을 선택한다
    Then Header에 "Claude Sonnet"이 표시된다
```

### 1.6 비용 추적 (AC-COST)

#### AC-COST-001: 비용 표시

```gherkin
Feature: 비용 추적
  시스템은 API 사용 비용을 실시간으로 추적한다

  Scenario: 세션 비용 업데이트
    Given CostTracker가 표시되어 있다
    When 새 채팅 응답이 완료된다
    Then 입력 토큰 수가 업데이트된다
    And 출력 토큰 수가 업데이트된다
    And 세션 비용이 계산되어 표시된다

  Scenario: 비용 상세 보기
    Given 세션 비용이 $0.15로 표시되어 있다
    When 사용자가 CostTracker를 클릭한다
    Then CostBreakdown 모달이 열린다
    And 모델별 비용이 표시된다
    And 입력/출력별 비용이 분리 표시된다
```

### 1.7 테마 (AC-THEME)

#### AC-THEME-001: 다크 모드 토글

```gherkin
Feature: 다크 모드
  시스템은 다크 모드를 지원한다

  Scenario: 다크 모드로 전환
    Given 현재 라이트 모드이다
    When 사용자가 테마 토글 버튼을 클릭한다
    Then 전체 UI가 다크 모드 색상으로 변경된다
    And 테마 설정이 localStorage에 저장된다

  Scenario: 시스템 테마 따르기
    Given 사용자가 테마를 "시스템"으로 설정했다
    And 시스템이 다크 모드이다
    When 페이지가 로드된다
    Then 다크 모드로 표시된다
```

---

## 2. 비기능 테스트 시나리오

### 2.1 성능 (AC-PERF)

#### AC-PERF-001: 초기 로드 성능

```gherkin
Feature: 초기 로드 성능
  대시보드는 빠르게 로드되어야 한다

  Scenario: First Contentful Paint
    Given 사용자가 대시보드에 처음 접속한다
    When 페이지 로드가 시작된다
    Then FCP(First Contentful Paint)가 1.5초 이내이다

  Scenario: Time to Interactive
    Given 페이지 로드가 진행 중이다
    When 모든 리소스가 로드된다
    Then TTI(Time to Interactive)가 3초 이내이다
```

#### AC-PERF-002: 대량 메시지 성능

```gherkin
Feature: 대량 메시지 렌더링 성능
  많은 메시지가 있어도 성능이 유지되어야 한다

  Scenario: 1000개 메시지 렌더링
    Given 대화에 1000개의 메시지가 있다
    When MessageList가 렌더링된다
    Then 스크롤이 60fps로 부드럽게 동작한다
    And 메모리 사용량이 200MB를 초과하지 않는다
```

### 2.2 보안 (AC-SEC)

#### AC-SEC-001: API 키 보호

```gherkin
Feature: API 키 보호
  API 키가 클라이언트에 노출되지 않아야 한다

  Scenario: 번들 검사
    Given 프로덕션 빌드가 완료되었다
    When JavaScript 번들을 분석한다
    Then API 키 패턴이 발견되지 않는다
    And 환경 변수가 번들에 포함되지 않는다

  Scenario: 네트워크 요청 검사
    Given 사용자가 채팅을 진행한다
    When 네트워크 요청을 모니터링한다
    Then Authorization 헤더에 API 키가 직접 포함되지 않는다
    And 모든 AI API 호출이 백엔드를 경유한다
```

#### AC-SEC-002: XSS 방지

```gherkin
Feature: XSS 방지
  사용자 입력이 안전하게 렌더링되어야 한다

  Scenario: 스크립트 주입 방지
    Given 사용자가 "<script>alert('xss')</script>"를 입력한다
    When 메시지가 렌더링된다
    Then 스크립트가 실행되지 않는다
    And 텍스트가 이스케이프되어 표시된다
```

### 2.3 접근성 (AC-A11Y)

#### AC-A11Y-001: 키보드 접근성

```gherkin
Feature: 키보드 접근성
  모든 기능이 키보드로 접근 가능해야 한다

  Scenario: Tab 네비게이션
    Given 사용자가 키보드만 사용한다
    When Tab 키로 네비게이션한다
    Then 모든 인터랙티브 요소에 접근 가능하다
    And 포커스 인디케이터가 명확히 표시된다

  Scenario: 스크린 리더 지원
    Given 스크린 리더가 활성화되어 있다
    When 페이지를 탐색한다
    Then 모든 요소에 적절한 ARIA 라벨이 있다
    And 동적 콘텐츠 변경이 공지된다
```

---

## 3. Quality Gate

### 3.1 Definition of Done

- [ ] 모든 기능 테스트 시나리오 통과
- [ ] 단위 테스트 커버리지 85% 이상
- [ ] E2E 테스트 주요 시나리오 통과
- [ ] Lighthouse 성능 점수 90점 이상
- [ ] TypeScript strict 모드 에러 없음
- [ ] ESLint 경고/에러 없음
- [ ] 번들 사이즈 500KB 이하 (gzipped)
- [ ] WCAG 2.1 AA 준수
- [ ] 코드 리뷰 완료
- [ ] 문서화 완료 (README, 컴포넌트 스토리)

### 3.2 테스트 매트릭스

| 테스트 유형 | 도구 | 커버리지 목표 | 실행 환경 |
|------------|------|-------------|----------|
| 단위 테스트 | Vitest | 85% | CI |
| 컴포넌트 테스트 | Testing Library | 80% | CI |
| E2E 테스트 | Playwright | 주요 시나리오 | CI |
| 접근성 테스트 | axe-core | WCAG 2.1 AA | CI |
| 성능 테스트 | Lighthouse | 90+ | CI |
| 보안 테스트 | npm audit | 0 high/critical | CI |

### 3.3 브라우저 지원

| 브라우저 | 최소 버전 | 테스트 수준 |
|---------|----------|-----------|
| Chrome | 120+ | Full |
| Firefox | 120+ | Full |
| Safari | 17+ | Full |
| Edge | 120+ | Smoke |

---

## 4. 테스트 파일 구조

```
web-ui/
├── src/
│   └── __tests__/
│       ├── components/
│       │   ├── layout/
│       │   │   └── AppShell.test.tsx
│       │   ├── chat/
│       │   │   ├── ChatContainer.test.tsx
│       │   │   ├── MessageList.test.tsx
│       │   │   └── MessageInput.test.tsx
│       │   └── spec/
│       │       └── SpecMonitor.test.tsx
│       ├── hooks/
│       │   ├── useWebSocket.test.ts
│       │   └── useChat.test.ts
│       └── stores/
│           ├── appStore.test.ts
│           └── chatStore.test.ts
├── e2e/
│   ├── chat.spec.ts
│   ├── spec-monitor.spec.ts
│   ├── provider-switch.spec.ts
│   └── accessibility.spec.ts
└── vitest.config.ts
```

---

## 5. 추적성 태그

- **SPEC**: SPEC-WEB-002
- **의존**: SPEC-WEB-001
- **연관**: [spec.md](./spec.md), [plan.md](./plan.md)

---

**문서 버전**: 1.0.0
**최종 수정**: 2026-01-09
**작성자**: workflow-spec agent
