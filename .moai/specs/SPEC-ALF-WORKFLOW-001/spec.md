---
id: ALF-WORKFLOW-001
version: 0.1.0
status: completed
created: 2025-10-29
updated: 2025-10-29
priority: high
category: feature
labels:
  - alfred
  - workflow
  - architecture
  - intent-understanding
  - plan-agent
depends_on: []
blocks: []
related_specs: []
scope:
  packages:
    - .claude/agents/alfred
    - .claude/agents/implementation-planner
    - .claude/agents/tdd-implementer
    - .claude/agents/doc-syncer
    - .claude/agents/git-manager
    - .claude/agents/spec-builder
    - .claude/commands/alfred
  files:
    - CLAUDE.md
    - CLAUDE-RULES.md
    - .claude/commands/alfred/1-plan.md
    - .claude/commands/alfred/2-run.md
    - .claude/commands/alfred/3-sync.md
    - .claude/agents/implementation-planner/main.md
    - .claude/agents/tdd-implementer/main.md
    - .claude/agents/doc-syncer/main.md
    - .claude/agents/git-manager/main.md
    - .claude/agents/spec-builder/main.md
---


## HISTORY

### v0.1.0 (2025-10-29)
- **IMPLEMENTATION COMPLETED**: 4단계 워크플로우 구현 완료 (status: draft → completed)
- **IMPLEMENTATION**:
  - CLAUDE.md: "4-Step Workflow Logic" 섹션 추가
  - CLAUDE-RULES.md: TodoWrite, Report Generation, Git Commit 규칙 추가
  - Commands: 3개 커맨드 템플릿에 워크플로우 통합
  - Agents: 5개 에이전트에 워크플로우 역할 추가
- **COMMITS**:
  - ba542efb: Phase 0 - SPEC 문서 생성
  - 3b130224: feat - 4단계 워크플로우 로직 구현
- **FILES**: 10개 핵심 파일 업데이트

### v0.0.1 (2025-10-29)
- **INITIAL**: Alfred의 새로운 4단계 워크플로우 로직 SPEC 초안 작성
- **SECTIONS**: 환경(Environment), 가정(Assumptions), 요구사항(Requirements), 명세(Specifications), 추적성(Traceability)
- **CONTEXT**: 기존 Alfred 워크플로우를 체계화하여 의도 파악 → 계획 수립 → 작업 실행 → 보고/커밋의 4단계로 구조화

---

## 환경 (Environment)

### 시스템 환경

- **MoAI-ADK 버전**: v0.8.2+
- **Python 버전**: ≥3.13
- **Alfred SuperAgent**: 19명의 팀원 (Alfred + 12개 핵심 서브 에이전트 + 6개 전문가)
- **Claude Skills**: 55개 (Foundation, Essentials, Alfred, Domain, Language, Ops 계층)
- **대화 언어**: 설정 가능 (한국어, 영어, 일본어, 중국어, 스페인어 등)
- **프레임워크 언어**: 영어 (모든 코어 파일: CLAUDE.md, agents, commands, skills, memory)

### 아키텍처 맥락

- **현재 상태**: Alfred는 `/alfred:0-project`, `/alfred:1-plan`, `/alfred:2-run`, `/alfred:3-sync` 명령을 통해 SPEC-first TDD 워크플로우를 조율
- **문제점**:
  1. 사용자 의도가 불명확할 때 가정(assumption)을 통해 진행하는 경우 발생
  2. 작업 계획 없이 즉시 실행하여 병렬화 기회 상실
  3. 작업 진행 상황 추적이 불투명
  4. 명시적 요청 없이 보고서 파일을 과도하게 생성
- **목표**: 4단계 워크플로우를 통해 의도 명확화 → 체계적 계획 → 투명한 실행 → 선택적 보고로 개선

### 의존성

- **필수 에이전트**:
  - Alfred (SuperAgent, 오케스트레이터)
  - Plan Agent (built-in Claude Agent, 작업 분석 및 계획 수립)
  - implementation-planner (복잡한 작업 분석)
  - spec-builder (SPEC 문서 작성)
  - tdd-implementer (TDD 구현)
  - doc-syncer (문서 동기화)
  - git-manager (Git 작업 관리)

- **필수 도구**:
  - `AskUserQuestion` (대화형 질문 도구, moai-alfred-ask-user-questions skill에 문서화)
  - `TodoWrite` (작업 목록 추적 도구)
  - `Plan` (Claude 내장 에이전트, 작업 분석 및 계획)

---

## 가정 (Assumptions)

### 사용자 가정

1. **사용자는 자연어로 의도를 표현**
   - 명확한 요청: "JWT 인증 시스템 만들어줘"
   - 모호한 요청: "대시보드 추가해줘", "성능 개선해줘"
   - Alfred는 모호한 경우 반드시 명확화 질문 수행

2. **사용자는 최소한의 개입을 선호**
   - 명시적 질문이 필요한 경우에만 `AskUserQuestion` 사용
   - 기술적 결정은 Alfred가 자율적으로 수행
   - 단, 비즈니스/UX 결정은 반드시 사용자 확인 필요

3. **사용자는 진행 상황을 시각적으로 확인하고 싶어함**
   - TodoWrite를 통한 작업 목록 표시
   - 각 작업의 상태 (pending → in_progress → completed) 실시간 업데이트
   - 차단 요인 발생 시 즉시 알림

### 시스템 가정

1. **Plan Agent는 모든 작업을 정확하게 분석 가능**
   - 단일 작업 vs 병렬 작업 구분
   - 작업 간 의존성 파악
   - 우선순위 결정

2. **TodoWrite는 작업 진행 상태의 단일 진실 공급원 (Single Source of Truth)**
   - 모든 에이전트는 TodoWrite를 통해 진행 상황 보고
   - 동시에 하나의 작업만 in_progress 상태 유지

3. **Git 작업은 모두 git-manager에게 위임**
   - spec-builder, tdd-implementer, doc-syncer는 Git 작업 수행하지 않음
   - git-manager만 branch 생성, commit, PR 생성 책임

### 제약 사항

1. **보고서 파일은 명시적 요청 시에만 생성**
   - ❌ 자동 생성: `IMPLEMENTATION_GUIDE.md`, `EXPLORATION_REPORT.md`, `*_ANALYSIS.md`
   - ✅ 허용: `.moai/docs/`, `.moai/reports/`, `.moai/specs/SPEC-*/` 내부 문서
   - 사용자가 "보고서 만들어줘", "분석 문서 작성해줘" 명시한 경우에만 생성

2. **커밋은 항상 생성**
   - 모든 작업 완료 시 반드시 Git 커밋 생성
   - TDD 단계별 커밋 (RED → GREEN → REFACTOR)
   - Alfred co-authorship 포함

---

## 요구사항 (Requirements)

### Ubiquitous Requirements (기초 요구사항)

**REQ-ALF-WF-001**: Alfred는 모든 사용자 요청에 대해 4단계 워크플로우를 따라야 한다.
- **4단계**: 의도 파악 (Intent Understanding) → 계획 수립 (Plan Creation) → 작업 실행 (Task Execution) → 보고/커밋 (Report & Commit)

**REQ-ALF-WF-002**: Alfred는 사용자의 configured `conversation_language`로 모든 대화를 수행해야 한다.

**REQ-ALF-WF-003**: 시스템은 작업 진행 상황을 TodoWrite를 통해 실시간 추적해야 한다.
- **상태**: pending, in_progress, completed
- **규칙**: 동시에 정확히 하나의 작업만 in_progress 상태

### Event-driven Requirements (이벤트 기반 요구사항)

**REQ-ALF-WF-004**: WHEN 사용자의 의도가 불명확한 경우, Alfred는 `AskUserQuestion` 도구를 사용하여 의도를 명확히 해야 한다.
- **트리거**:
  - 여러 기술 스택 선택지 존재 (PostgreSQL vs MongoDB)
  - 아키텍처 결정 필요 (Microservices vs Monolithic)
  - 모호한 요청 ("대시보드 추가", "성능 최적화")
  - 기존 컴포넌트 영향 불확실 (리팩토링 범위, 하위 호환성)
- **예외**: 사용자가 이미 명확한 지시를 제공한 경우

**REQ-ALF-WF-005**: WHEN 의도가 명확해진 후, Alfred는 Plan Agent를 호출하여 작업을 분석해야 한다.
- **Plan Agent 역할**:
  - 단일 작업 vs 병렬 작업 판단
  - 작업 간 의존성 파악
  - 우선순위 결정
  - 예상 수정 파일 목록 생성

**REQ-ALF-WF-006**: WHEN Plan Agent가 분석을 완료하면, Alfred는 TodoWrite를 통해 작업 목록을 생성해야 한다.
- **작업 항목 형식**:
  - `content`: 명령형 (예: "SPEC 문서 작성", "테스트 코드 구현")
  - `activeForm`: 진행형 (예: "SPEC 문서 작성 중", "테스트 코드 구현 중")
  - `status`: pending (초기 상태)

**REQ-ALF-WF-007**: WHEN 작업 실행 시작 시, Alfred는 TodoWrite를 통해 해당 작업을 `in_progress` 상태로 변경해야 한다.
- **규칙**: 이전 작업이 completed 되기 전에 다음 작업 시작 불가

**REQ-ALF-WF-008**: WHEN 작업이 완료되면, Alfred는 즉시 TodoWrite에서 해당 작업을 `completed` 상태로 변경해야 한다.
- **완료 기준**:
  - ✅ 테스트 통과
  - ✅ 구현 완료
  - ✅ 오류 없음
- **미완료 기준**:
  - ❌ 테스트 실패
  - ❌ 부분 구현
  - ❌ 차단 요인 발생

**REQ-ALF-WF-009**: WHEN 모든 작업이 완료되면, Alfred는 git-manager에게 커밋 생성을 요청해야 한다.
- **커밋 메시지 형식**: TDD 단계별 (test/feat/refactor)
- **Alfred co-authorship 포함**: `Co-Authored-By: 🎩 Alfred@[MoAI](https://adk.mo.ai.kr)`

**REQ-ALF-WF-010**: WHEN 사용자가 명시적으로 보고서를 요청한 경우에만, Alfred는 보고서 파일을 생성해야 한다.
- **명시적 요청 예시**:
  - "보고서 만들어줘"
  - "분석 문서 작성해줘"
  - "구현 가이드 문서화해줘"
- **금지 사항**: 자동으로 루트 디렉토리에 `*_GUIDE.md`, `*_REPORT.md`, `*_ANALYSIS.md` 생성 금지

### State-driven Requirements (상태 기반 요구사항)

**REQ-ALF-WF-011**: WHILE 작업이 in_progress 상태인 동안, Alfred는 진행 상황을 사용자의 `conversation_language`로 보고해야 한다.
- **보고 내용**: 현재 작업 이름, 진행률, 예상 다음 단계

**REQ-ALF-WF-012**: WHILE TodoWrite 작업 목록이 존재하는 동안, Alfred는 정확히 하나의 작업만 in_progress 상태로 유지해야 한다.
- **위반 시**: 시스템 오류 발생 및 작업 중단

**REQ-ALF-WF-013**: WHILE 차단 요인이 발생한 동안, Alfred는 해당 작업을 in_progress 상태로 유지하고 새로운 작업을 생성해야 한다.
- **차단 요인 예시**:
  - 사용자 입력 필요
  - 외부 의존성 오류
  - 설정 변경 필요
- **새 작업 예시**: "차단 요인 해결: <설명>"

### Optional Features (선택적 기능)

**REQ-ALF-WF-014**: WHERE 사용자가 명시적으로 요청한 경우, Alfred는 `.moai/docs/` 또는 `.moai/reports/`에 내부 문서를 생성할 수 있다.
- **허용 위치**:
  - `.moai/docs/` (구현 가이드, 전략 문서)
  - `.moai/reports/` (동기화 보고서, TAG 검증)
  - `.moai/analysis/` (기술 분석)
  - `.moai/specs/SPEC-*/` (SPEC 문서)

**REQ-ALF-WF-015**: WHERE Plan Agent가 병렬 작업을 식별한 경우, Alfred는 여러 서브 에이전트를 동시에 호출할 수 있다.
- **병렬 작업 조건**:
  - 작업 간 의존성 없음
  - 파일 충돌 가능성 없음
  - 독립적으로 실행 가능

### Unwanted Behaviors (원치 않는 동작)

**REQ-ALF-WF-016**: Alfred는 사용자 의도가 불명확할 때 가정을 통해 즉시 구현해서는 안 된다.
- ❌ 금지: "대시보드 추가"를 듣고 임의로 React + Recharts로 구현
- ✅ 올바름: `AskUserQuestion`으로 데이터 소스, 차트 유형, 접근 권한 확인 후 구현

**REQ-ALF-WF-017**: Alfred는 명시적 요청 없이 프로젝트 루트에 문서 파일을 생성해서는 안 된다.
- ❌ 금지: `IMPLEMENTATION_GUIDE.md`, `EXPLORATION_REPORT.md`, `ANALYSIS_REPORT.md`
- ✅ 허용: `README.md`, `CHANGELOG.md`, `CONTRIBUTING.md`, `LICENSE` (공식 문서)

**REQ-ALF-WF-018**: Alfred는 작업 완료 전에 다음 작업을 시작해서는 안 된다.
- ❌ 금지: 테스트 실패 상태에서 다음 기능 구현 시작
- ✅ 올바름: 현재 작업 completed로 변경 후 다음 작업 시작

---

## 명세 (Specifications)

### 1단계: Intent Understanding (의도 파악)

#### 1.1 사용자 요청 수신

```
User: "JWT 인증 시스템 만들어줘. 30분 토큰 만료, refresh token 지원"

Alfred (내부 분석):
├─ 명확도 평가: HIGH (기술 스택, 요구사항 명시됨)
├─ AskUserQuestion 필요 여부: NO
└─ 다음 단계: Plan Agent 호출
```

#### 1.2 모호한 요청 처리

```
User: "대시보드 추가해줘"

Alfred (내부 분석):
├─ 명확도 평가: LOW (데이터 소스, 차트 유형 불명확)
├─ AskUserQuestion 필요 여부: YES
└─ 질문 준비:
    ├─ "데이터 소스?" → [REST API | GraphQL | Local state]
    ├─ "주요 차트 유형?" → [Time series | Category comparison | Distribution]
    ├─ "실시간 업데이트 필요?" → [Yes | No | Every 10 seconds]
    └─ "접근 제한?" → [Admin only | Logged-in users | Public]

[사용자 응답 수집]

Alfred (확인):
├─ 선택 사항 요약 표시
├─ "이 선택으로 진행할까요?" 최종 확인
└─ 다음 단계: Plan Agent 호출
```

#### 1.3 AskUserQuestion 사용 기준

**필수 사용 상황**:
| 상황 | 예시 |
|------|------|
| 여러 기술 스택 선택지 | PostgreSQL vs MongoDB, Redux vs Zustand |
| 아키텍처 결정 | Microservices vs Monolithic, CSR vs SSR |
| 모호한 요청 | "대시보드 추가", "성능 최적화" |
| 기존 컴포넌트 영향 | 리팩토링 범위, 하위 호환성 |
| UX/비즈니스 로직 결정 | UI 레이아웃, 데이터 표시 방법 |

**선택 사용 상황**:
- ✅ 사용자가 이미 명확한 지시 제공
- ✅ 표준 규칙이나 모범 사례가 명확함
- ✅ 기술적 제약으로 하나의 선택만 가능
- ✅ 사용자가 "그냥 구현해, 이미 결정했어"라고 명시

### 2단계: Plan Creation (계획 수립)

#### 2.1 Plan Agent 호출

```
Alfred → Plan Agent:
  Task: "JWT 인증 시스템 구현 작업 분석"
  Context: {
    - 사용자 요청: "JWT 인증, 30분 만료, refresh token 지원"
    - 프로젝트 언어: Python
    - 기존 아키텍처: FastAPI backend
  }

Plan Agent (분석):
├─ 작업 유형: 단일 기능 (복잡도 중간)
├─ 예상 수정 파일:
│   ├─ src/auth/jwt_service.py (신규)
│   ├─ src/auth/middleware.py (신규)
│   ├─ tests/auth/test_jwt_service.py (신규)
│   └─ .moai/specs/SPEC-AUTH-001/ (신규 디렉토리)
├─ 단계 분해:
│   ├─ Phase 1: SPEC 문서 작성 (spec-builder)
│   ├─ Phase 2: 테스트 코드 작성 (tdd-implementer, RED)
│   ├─ Phase 3: 구현 (tdd-implementer, GREEN)
│   ├─ Phase 4: 리팩토링 (tdd-implementer, REFACTOR)
│   └─ Phase 5: 문서 동기화 (doc-syncer)
└─ 실행 방식: 순차 (의존성 있음)
```

#### 2.2 TodoWrite 작업 목록 생성

```
Alfred → TodoWrite:
  [
    {
      content: "SPEC-AUTH-001 문서 작성 (spec.md, plan.md, acceptance.md)",
      activeForm: "SPEC-AUTH-001 문서 작성 중",
      status: "pending"
    },
    {
      content: "JWT 서비스 테스트 코드 작성 (RED 단계)",
      activeForm: "JWT 서비스 테스트 코드 작성 중",
      status: "pending"
    },
    {
      content: "JWT 서비스 구현 (GREEN 단계)",
      activeForm: "JWT 서비스 구현 중",
      status: "pending"
    },
    {
      content: "코드 리팩토링 및 품질 개선 (REFACTOR 단계)",
      activeForm: "코드 리팩토링 및 품질 개선 중",
      status: "pending"
    },
    {
      content: "문서 동기화 및 TAG 검증 (/alfred:3-sync)",
      activeForm: "문서 동기화 및 TAG 검증 중",
      status: "pending"
    },
    {
      content: "Git 커밋 생성 및 PR 업데이트",
      activeForm: "Git 커밋 생성 및 PR 업데이트 중",
      status: "pending"
    }
  ]
```

### 3단계: Task Execution (작업 실행)

#### 3.1 순차 실행 패턴

```
Alfred (Task 1 시작):
├─ TodoWrite 업데이트: Task 1 → in_progress
├─ spec-builder 호출: "SPEC-AUTH-001 작성"
│   └─ spec-builder: 3개 파일 생성 (spec.md, plan.md, acceptance.md)
├─ spec-builder 완료
├─ TodoWrite 업데이트: Task 1 → completed
└─ 다음 작업 시작

Alfred (Task 2 시작):
├─ TodoWrite 업데이트: Task 2 → in_progress
├─ tdd-implementer 호출: "RED 단계 - 테스트 작성"
│   └─ tdd-implementer: tests/auth/test_jwt_service.py 생성
│       └─ pytest 실행 → FAIL (예상됨)
├─ tdd-implementer 완료
├─ TodoWrite 업데이트: Task 2 → completed
└─ 다음 작업 시작

[Task 3~5 동일 패턴]
```

#### 3.2 병렬 실행 패턴 (의존성 없는 경우)

```
Plan Agent (분석 결과):
├─ 작업 A: README.md 업데이트 (doc-syncer)
├─ 작업 B: CHANGELOG.md 생성 (doc-syncer)
└─ 의존성: 없음 (병렬 실행 가능)

Alfred (병렬 실행):
├─ TodoWrite 업데이트: Task A, Task B → in_progress (동시)
├─ 병렬 호출:
│   ├─ Thread 1: doc-syncer → README.md 업데이트
│   └─ Thread 2: doc-syncer → CHANGELOG.md 생성
├─ 두 작업 완료 대기
└─ TodoWrite 업데이트: Task A, Task B → completed (동시)
```

#### 3.3 차단 요인 처리

```
Alfred (Task 3 실행 중):
├─ tdd-implementer 호출: "GREEN 단계 - 구현"
├─ tdd-implementer: 오류 발생 → "PyJWT 패키지 필요"
├─ 차단 요인 감지
├─ TodoWrite 업데이트: Task 3 → in_progress (유지)
└─ 새 작업 생성:
    {
      content: "차단 요인 해결: PyJWT 패키지 설치",
      activeForm: "차단 요인 해결 중",
      status: "in_progress"
    }

Alfred (차단 요인 해결):
├─ pip install PyJWT 실행
├─ TodoWrite 업데이트: 차단 작업 → completed
└─ Task 3 재개
```

### 4단계: Report & Commit (보고/커밋)

#### 4.1 커밋 생성 (항상 수행)

```
Alfred → git-manager:
  Action: "Create TDD commits for SPEC-AUTH-001"
  Context: {
    - SPEC ID: AUTH-001
    - Modified files: [src/auth/*, tests/auth/*, .moai/specs/SPEC-AUTH-001/*]
    - TDD stages: RED, GREEN, REFACTOR
  }

git-manager (실행):
├─ Commit 1 (RED):
│   Message: "test: add failing tests for JWT authentication
│
│   - Add test_jwt_service.py with token generation tests
│   - Add test cases for token expiration and refresh
│
│
│   🤖 Generated with [Claude Code](https://claude.com/claude-code)
│
│   Co-Authored-By: 🎩 Alfred@[MoAI](https://adk.mo.ai.kr)"
│
├─ Commit 2 (GREEN):
│   Message: "feat: implement JWT authentication service
│
│   - Implement jwt_service.py with token generation
│   - Add middleware for token validation
│   - Support 30-minute token expiration and refresh
│
│
│   🤖 Generated with [Claude Code](https://claude.com/claude-code)
│
│   Co-Authored-By: 🎩 Alfred@[MoAI](https://adk.mo.ai.kr)"
│
└─ Commit 3 (REFACTOR):
    Message: "refactor: improve JWT service code quality

    - Extract token generation logic to separate function
    - Add comprehensive error handling
    - Improve type hints for better IDE support


    🤖 Generated with [Claude Code](https://claude.com/claude-code)

    Co-Authored-By: 🎩 Alfred@[MoAI](https://adk.mo.ai.kr)"
```

#### 4.2 보고서 생성 (명시적 요청 시에만)

```
Case 1: 보고서 미요청 (기본)
User: "JWT 인증 시스템 만들어줘"
Alfred: [4단계 워크플로우 실행]
       [커밋 생성]
       [보고서 파일 생성 안 함] ✅

Case 2: 보고서 명시적 요청
User: "JWT 인증 시스템 만들고, 구현 가이드 문서도 작성해줘"
Alfred: [4단계 워크플로우 실행]
       [커밋 생성]
       [보고서 생성: .moai/docs/implementation-AUTH-001.md] ✅

Case 3: 분석 문서 요청
User: "기존 인증 시스템 분석하고 보고서 만들어줘"
Alfred: [분석 수행]
       [보고서 생성: .moai/analysis/auth-system-analysis.md] ✅
       [커밋 생성]
```

### 파일 수정 범위

#### Phase 1: 기초 문서 업데이트
- `CLAUDE.md`: Alfred 핵심 지침에 4단계 워크플로우 추가
- `CLAUDE-RULES.md`: Skill 호출 규칙, 대화형 질문 규칙 업데이트

#### Phase 2: 명령 템플릿 업데이트
- `.claude/commands/alfred/1-plan.md`: Plan Agent 호출 로직 추가
- `.claude/commands/alfred/2-run.md`: TodoWrite 통합, 차단 요인 처리
- `.claude/commands/alfred/3-sync.md`: 최종 보고 로직 업데이트

#### Phase 3: 에이전트 업데이트
- `.claude/agents/implementation-planner/main.md`: Plan Agent 통합 지침
- `.claude/agents/tdd-implementer/main.md`: TodoWrite 진행 상황 보고
- `.claude/agents/doc-syncer/main.md`: 선택적 보고서 생성 규칙
- `.claude/agents/git-manager/main.md`: 커밋 생성 필수화
- `.claude/agents/spec-builder/main.md`: SPEC 작성 시 TodoWrite 업데이트

### EARS 요구사항 매핑

| EARS 유형 | 요구사항 ID | 설명 |
|-----------|-------------|------|
| Ubiquitous | REQ-ALF-WF-001, REQ-ALF-WF-002, REQ-ALF-WF-003 | 4단계 워크플로우 필수, 대화 언어 준수, TodoWrite 추적 |
| Event-driven | REQ-ALF-WF-004~010 | 의도 불명확 시 질문, Plan Agent 호출, TodoWrite 상태 변경, 커밋 생성, 선택적 보고서 |
| State-driven | REQ-ALF-WF-011~013 | 진행 중 보고, 하나의 in_progress 유지, 차단 요인 처리 |
| Optional | REQ-ALF-WF-014~015 | 내부 문서 생성, 병렬 작업 실행 |
| Unwanted | REQ-ALF-WF-016~018 | 가정 금지, 자동 보고서 생성 금지, 작업 완료 전 다음 작업 시작 금지 |

---

## 추적성 (Traceability)

### TAG 체계


### 영향받는 컴포넌트

| 컴포넌트 | 파일 경로 | 수정 유형 | 우선순위 |
|----------|-----------|-----------|----------|
| Alfred 지침 | CLAUDE.md | 4단계 워크플로우 추가 | High |
| 규칙 문서 | CLAUDE-RULES.md | AskUserQuestion, TodoWrite 규칙 | High |
| 1-plan 명령 | .claude/commands/alfred/1-plan.md | Plan Agent 통합 | High |
| 2-run 명령 | .claude/commands/alfred/2-run.md | TodoWrite 통합 | High |
| 3-sync 명령 | .claude/commands/alfred/3-sync.md | 보고 로직 | Medium |
| implementation-planner | .claude/agents/implementation-planner/main.md | Plan 패턴 | Medium |
| tdd-implementer | .claude/agents/tdd-implementer/main.md | TodoWrite 보고 | Medium |
| doc-syncer | .claude/agents/doc-syncer/main.md | 선택적 보고서 | Medium |
| git-manager | .claude/agents/git-manager/main.md | 필수 커밋 | Medium |
| spec-builder | .claude/agents/spec-builder/main.md | TodoWrite 업데이트 | Low |

### 검증 방법

```bash
# 1. TAG 체인 검증
rg '@(SPEC|CODE):ALF-WORKFLOW-001' -n CLAUDE.md .claude/

# 2. 수정 파일 확인
git diff --name-only feature/SPEC-ALF-WORKFLOW-001

# 3. 필수 키워드 존재 확인
rg 'AskUserQuestion|TodoWrite|Plan Agent' CLAUDE.md CLAUDE-RULES.md

# 4. 커밋 메시지 검증
```

---

## 참고 문서

- **CLAUDE.md**: Alfred SuperAgent 핵심 지침
- **CLAUDE-RULES.md**: 필수 규칙 및 표준
- **CLAUDE-AGENTS-GUIDE.md**: 에이전트 선택 기준
- **CLAUDE-PRACTICES.md**: 실전 워크플로우 예시
- **.moai/memory/spec-metadata.md**: SPEC 메타데이터 표준
- **.moai/memory/language-config-schema.md**: 언어 설정 스키마

---

**마지막 업데이트**: 2025-10-29
**문서 버전**: v0.0.1
