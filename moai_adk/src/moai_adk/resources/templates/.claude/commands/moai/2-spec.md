---
description: EARS 형식 명세 작성 - 비즈니스 요구사항을 엔지니어링 관점에서 구조화된 명세로 변환
argument-hint: <feature-name> [description] [--clarification-only]
allowed-tools: Read, Write, Edit, MultiEdit, Task
---

# MoAI-ADK EARS 형식 명세 작성

MoAI-ADK 파이프라인의 최초 진입점으로, 비즈니스 요구사항을 엔지니어링 관점에서 인간이 읽기 쉬우면서도 컴퓨터가 처리하기 적합한 EARS(Easy Approach to Requirements Syntax) 형식으로 변환합니다. 불명확한 부분은 [NEEDS CLARIFICATION] 마커로 표시하여 구현 전에 소통 비용을 최소화합니다.

## 실행 플로우

```mermaid
flowchart TD
    A[사용자 설명 파싱] --> B{설명 존재?}
    B -->|비어있음| C[ERROR: No feature description]
    B -->|존재| D[핵심 개념 추출]
    
    D --> E[actors, actions, data, constraints]
    E --> F[불명확한 부분 마킹]
    F --> G[[NEEDS CLARIFICATION: 구체적 질문]]
    
    G --> H[User Stories 작성]
    H --> I[US-XXX 형식]
    I --> J[EARS 형식 요구사항 생성]
    
    J --> K[WHEN/IF/WHILE/WHERE/UBIQUITOUS]
    K --> L[수락 기준 (Given-When-Then)]
    L --> M[Review Checklist 실행]
    
    M --> N{[NEEDS CLARIFICATION] 존재?}
    N -->|존재| O[WARN: 명확화 필요]
    N -->|없음| P[명세 완성]
    
    P --> Q[컨텍스트 초기화 권장]
```

## 🤖 자연어 체이닝 오케스트레이션

🤖 **EARS 형식 명세 작성을 전문 에이전트 체인으로 완전 자동화합니다.**

**요구사항 분석 단계**: Task tool을 사용하여 requirements-analyst 에이전트를 호출하여 사용자 입력을 구조화된 EARS 형식(WHEN/IF/WHILE/WHERE/UBIQUITOUS)으로 변환하고 [NEEDS CLARIFICATION] 마커를 통한 불명확 요소를 식별합니다.

**사용자 스토리 생성 단계**: Task tool을 사용하여 user-story-generator 에이전트를 호출하여 EARS 요구사항을 바탕으로 US-XXX 형식의 체계적인 사용자 스토리를 생성하고 우선순위를 설정합니다.

**수락 기준 작성 단계**: Task tool을 사용하여 acceptance-criteria-writer 에이전트를 호출하여 각 사용자 스토리에 대한 Given-When-Then 형식의 구체적이고 테스트 가능한 수락 기준을 작성합니다.

## 핵심 개념 추출

### 1. Actors (주체) 식별
```markdown
- 사용자 (User)
- 관리자 (Admin)  
- 시스템 (System)
- 외부 서비스 (External Service)
- 타이머/스케줄러 (Timer/Scheduler)
```

### 2. Actions (행동) 분류
```markdown
- CRUD 작업: Create, Read, Update, Delete
- 비즈니스 로직: 계산, 검증, 변환, 승인
- 시스템 이벤트: 알림, 로깅, 모니터링
- 통합 작업: API 호출, 데이터 동기화
```

### 3. Data (데이터) 정의
```markdown
- 입력 데이터: 사용자 입력, API 요청
- 출력 데이터: 응답, 결과, 리포트
- 저장 데이터: 엔티티, 설정, 로그
- 임시 데이터: 세션, 캐시, 큐
```

### 4. Constraints (제약사항) 식별
```markdown
- 성능 제약: 응답시간, 처리량, 메모리
- 보안 제약: 인증, 권한, 암호화
- 비즈니스 제약: 규정, 정책, 승인 절차
- 기술 제약: 브라우저, 디바이스, 네트워크
```

## [NEEDS CLARIFICATION] 마커 시스템

### 자동 마킹 기준
불명확한 부분을 자동으로 감지하고 구체적인 질문으로 표시합니다:

```markdown
[NEEDS CLARIFICATION: 사용자 권한 체계가 명확하지 않습니다. 
일반 사용자, 관리자, 슈퍼 관리자의 구체적인 권한 범위를 정의해주세요.]

[NEEDS CLARIFICATION: "빠른 응답"이 구체적이지 않습니다. 
목표 응답시간을 ms 단위로 명시해주세요. (예: 500ms 이하)]

[NEEDS CLARIFICATION: 데이터 보관 정책이 누락되었습니다.
개인정보 보관기간과 삭제 정책을 명시해주세요.]
```

### 마킹 카테고리
```markdown
- 성능 기준 모호: 응답시간, 처리량, 동시접속 수
- 보안 정책 미정의: 인증 방식, 권한 체계, 암호화 수준
- 데이터 정책 누락: 보관기간, 백업, 복구 절차
- 사용자 역할 불명확: 권한, 책임, 접근 범위
- 비즈니스 로직 모호: 계산 방식, 승인 절차, 예외 처리
```

## User Stories 자동 생성 (US-XXX 형식)

### User Story 템플릿
```markdown
US-001: 사용자 로그인
As a 일반 사용자
I want to 이메일과 패스워드로 로그인
So that 개인화된 서비스를 이용할 수 있다

수락 기준:
- 올바른 이메일 형식 검증
- 패스워드 최소 8자리 이상
- 3회 실패 시 계정 임시 잠금
- 로그인 성공 시 대시보드 리다이렉트

US-002: 관리자 사용자 관리
As a 관리자
I want to 사용자 목록을 조회하고 관리
So that 시스템 운영을 효율적으로 할 수 있다

수락 기준:
- 페이지네이션 지원 (페이지당 50명)
- 검색 및 필터링 기능
- 사용자 상태 변경 (활성/비활성)
- 권한 수정 기능
```

## EARS 형식 요구사항 생성

### EARS 키워드별 사용법

#### WHEN - 조건 발생 시
```markdown
WHEN 사용자가 올바른 이메일과 패스워드를 입력하면,
시스템은 3초 이내에 JWT 토큰을 생성하고 대시보드로 리디렉션해야 한다.
```

#### IF - 상태 조건
```markdown
IF 잘못된 인증 정보가 3회 연속 입력되면,
시스템은 해당 계정을 15분간 일시적으로 잠그고 관리자에게 알림 메일을 발송해야 한다.
```

#### WHILE - 진행 중
```markdown
WHILE 사용자 세션이 활성 상태인 동안,
시스템은 30분마다 자동으로 JWT 토큰을 갱신하고 무효한 토큰을 감지해야 한다.
```

#### WHERE - 위치/컨텍스트
```markdown
WHERE 모바일 환경에서,
시스템은 Touch ID 또는 Face ID를 통한 생체 인증을 지원하고 앱 백그라운드 전환 시 화면을 보호해야 한다.
```

#### UBIQUITOUS - 항상
```markdown
UBIQUITOUS 모든 API 요청에 대해,
시스템은 구조화된 로그를 생성하고 응답시간, 에러율, 사용 패턴을 실시간으로 모니터링해야 한다.
```

## 수락 기준 (Given-When-Then)

### Given-When-Then 템플릿
각 User Story에 대한 구체적이고 테스트 가능한 수락 기준을 작성합니다:

```markdown
**Scenario 1: 성공적인 로그인**
Given 등록된 사용자가 존재하고
  And 올바른 이메일 "user@example.com"과 패스워드 "password123"을 가지고 있을 때
When 사용자가 로그인 폼에 올바른 정보를 입력하고 "로그인" 버튼을 클릭하면
Then 시스템은 3초 이내에 JWT 토큰을 생성하고
  And 대시보드 페이지로 리디렉션하며
  And "환영합니다, [사용자명]님" 메시지를 표시한다

**Scenario 2: 잘못된 로그인 정보**
Given 등록된 사용자가 존재할 때
When 사용자가 잘못된 패스워드를 입력하고 "로그인" 버튼을 클릭하면
Then 시스템은 "이메일 또는 패스워드가 잘못되었습니다" 메시지를 표시하고
  And 로그인 실패 횟수를 증가시키며
  And 로그인 폼을 초기화한다

**Scenario 3: 계정 잠금**
Given 사용자가 2회 연속으로 잘못된 정보를 입력한 상태일 때
When 사용자가 3번째로 잘못된 정보를 입력하면
Then 시스템은 해당 계정을 15분간 잠그고
  And "계정이 일시적으로 잠겼습니다. 15분 후 다시 시도해주세요" 메시지를 표시하며
  And 관리자에게 알림 이메일을 발송한다
```

## Review Checklist 실행

### 완결성 검증 체크리스트
명세 작성 완료 후 다음 항목들을 자동으로 검증합니다:

```markdown
명세 완결성 체크:
- 모든 User Story에 수락 기준이 정의되었는가?
- 모든 EARS 요구사항이 테스트 가능한 형태인가?
- [NEEDS CLARIFICATION] 마커가 모두 해결되었는가?
- 비기능 요구사항(성능, 보안, 가용성)이 포함되었는가?

품질 기준 체크:
- User Story당 평균 수락 기준 개수: 3-5개
- EARS 키워드 분포: WHEN(40%), IF(25%), WHILE(15%), WHERE(10%), UBIQUITOUS(10%)
- 명확성 점수: [NEEDS CLARIFICATION] 비율 10% 이하
- 추적성: 모든 기능에 @REQ 태그 매핑

추적성 매핑 체크:
- Steering 문서 연결: @VISION → User Story 매핑
- 태그 시스템: @REQ-XXX → EARS 요구사항 연결
- 향후 단계: 계획(PLAN) → 작업(TASKS) → 구현(DEV) 준비
```

### 검증 실패 시 경고
```markdown
**경고**: 명세 품질 기준 미달

미완료 항목:
- US-003: 수락 기준 누락 (0개/최소 3개)
- REQ-005: EARS 형식 부정확 (키워드 누락)
- [NEEDS CLARIFICATION: 데이터 보관 정책] 미해결

권장 조치:
1. 누락된 수락 기준 추가
2. EARS 형식 재검토
3. 명확화 마커 해결 후 재실행
```

## 단계 완료 후 컨텍스트 초기화 권장

### 컨텍스트 관리 안내
```markdown
**완료**: SPEC 문서 작성이 완료되었습니다!

**컨텍스트 최적화 권장**:
현재 컨텍스트가 명세 작성으로 포화 상태입니다.
다음 단계 진행 전 컨텍스트를 초기화하는 것을 권장합니다:

방법 1: 새 탭에서 계속
> 새로운 Claude Code 탭을 열고 다음 명령 실행:
> /moai:3-plan SPEC-001

방법 2: 컨텍스트 초기화
> /clear 명령 실행 후:
> /moai:3-plan SPEC-001

**Why?**: 각 단계별로 컨텍스트를 초기화하면:
- 이전 단계의 노이즈 제거
- 현재 단계에 집중된 최적 성능
- 메모리 효율성 향상
- 더 정확한 결과 생성
```

## 작업 타입별 처리

### $ARGUMENTS 첫 번째 인자: 작업 타입

#### `req` - 요구사항만 작성

```bash
> /moai:2-spec req "사용자 인증 기능"
```

- requirements.md 집중 작성
- 사용자 스토리와 수용 기준 상세화
- EARS 형식 완벽 적용

#### `design` - 설계 문서만 작성

```bash
> /moai:2-spec design "마이크로서비스 아키텍처"
```

- design.md 집중 작성
- 시스템 아키텍처 다이어그램 포함
- 기술적 의사결정 근거 명시

#### `tasks` - 작업 분해만 수행

```bash
> /moai:2-spec tasks "프로젝트 전체 태스크 분해"
```

- tasks.md 집중 작성
- 상세한 작업 단위 분해
- 의존성과 일정 계획

#### `all` - 전체 SPEC 문서 작성

```bash
> /moai:2-spec all "소셜 미디어 플랫폼 개발"
```

- requirements.md, design.md, tasks.md 모두 생성
- 통합적이고 일관된 명세 작성
- 전체 프로젝트 스코프 정의

### $ARGUMENTS 나머지 인자: 상세 내용

두 번째 인자부터는 작업의 구체적인 내용이나 요구사항을 기술합니다.

#### 예시 1: 상세 기능 명세

```bash
> /moai:2-spec all "실시간 채팅 시스템 - WebSocket 기반, 파일 첨부 지원, 읽음 표시 기능"
```

#### 예시 2: 성능 요구사항 포함

```bash
> /moai:2-spec req "API 응답시간 500ms 이하, 동시접속 10,000명 지원"
```

#### 예시 3: 특정 기술 스택 고려

```bash
> /moai:2-spec design "Next.js App Router, Prisma ORM, PostgreSQL 기반 설계"
```

## EARS 형식 적용 예시

### Before (일반적인 요구사항)

```
사용자가 로그인할 수 있어야 한다.
시스템은 빠르게 응답해야 한다.
```

### After (EARS 형식 적용)

```
WHEN 사용자가 올바른 이메일과 패스워드를 입력하면,
시스템은 3초 이내에 대시보드로 리디렉션해야 한다.

IF 잘못된 인증 정보가 3회 연속 입력되면,
시스템은 해당 계정을 15분간 일시적으로 잠그고 알림 메일을 발송해야 한다.

WHILE 사용자 세션이 활성 상태인 동안,
시스템은 30분마다 자동으로 토큰을 갱신해야 한다.

WHERE 모바일 환경에서,
시스템은 Touch ID 또는 Face ID를 통한 생체 인증을 지원해야 한다.
```

## 완료 시 산출물

```markdown
**완료**: EARS 형식 명세 작성이 완료되었습니다!

**생성된 문서**:
  ├── .moai/specs/SPEC-001/
  │   ├── spec.md           # EARS 형식 요구사항 (15개 항목)
  │   ├── acceptance.md     # Given-When-Then 수락 기준 (45개 시나리오)
  │   └── user-stories.md   # User Stories (US-001~015)
  └── .moai/indexes/
      └── tags.json         # 14-Core TAG 시스템 업데이트

**명세 품질 지표**:
  - User Stories: 15개 생성
  - EARS 요구사항: 45개 (WHEN 18개, IF 11개, WHILE 7개, WHERE 5개, UBIQUITOUS 4개)
  - 수락 기준: 45개 시나리오 (평균 3개/Story)
  - 명확성 점수: 92% ([NEEDS CLARIFICATION] 3개/총 48개 항목)

**14-Core TAG 매핑**:
  - @REQ:BUS-001~008: 비즈니스 요구사항
  - @REQ:SEC-001~003: 보안 요구사항
  - @REQ:PERF-001~002: 성능 요구사항
  - @REQ:UX-001~002: 사용자 경험 요구사항
  - 추적성 매트릭스: 100% 완료

**명확화 필요 항목** (3개):
  1. [NEEDS CLARIFICATION: 사용자 권한 체계 상세 정의 필요]
  2. [NEEDS CLARIFICATION: 데이터 보관 정책 및 GDPR 준수 방안]
  3. [NEEDS CLARIFICATION: 모바일 앱 푸시 알림 정책]

**다음 단계** (4단계 파이프라인):
  1. 명확화 완료 후: /moai:3-plan SPEC-001 (Constitution Check)
  2. 작업 분해: /moai:4-tasks SPEC-001 (TDD 기반)
  3. 구현 시작: /moai:5-dev T001 (Red-Green-Refactor)

**Pro Tips**:
- [NEEDS CLARIFICATION] 해결 전까지 /moai:3-plan 진행 불가
- 각 단계별 컨텍스트 초기화 권장 (/clear 또는 새 탭)
- 모든 변경사항은 14-Core TAG로 자동 추적됩니다
```

## ⚠️ 에러 처리

### Steering 문서 누락 시
```markdown
❌ ERROR: Steering 문서를 찾을 수 없습니다.

Steering 문서는 SPEC 작성의 필수 전제조건입니다.
먼저 다음 명령으로 프로젝트를 초기화해주세요:
> /moai:1-project init

Steering 문서 경로: .moai/steering/
```

### 불완전한 입력 시
```markdown
⚠️ WARNING: 작업 내용이 불충분합니다.

더 구체적인 요구사항을 제공해주세요:
- 핵심 기능은 무엇인가요?
- 성능 요구사항이 있나요?
- 특별한 제약사항이 있나요?

권장 형식: /moai:2-spec all "실시간 채팅 - WebSocket, 파일 첨부, 1000명 동시접속"
```

### [NEEDS CLARIFICATION] 미해결
```markdown
🔴 ERROR: 명세에 불명확한 부분이 있습니다.

미해결 항목:
- [NEEDS CLARIFICATION: 사용자 권한 체계]
- [NEEDS CLARIFICATION: 데이터 보관 정책]

다음 단계 진행을 위해 모든 명확화 마커를 해결해주세요.
```

이 명령어를 통해 체계적이고 완전한 SPEC 문서가 자동으로 생성되며, @TAG 시스템을 통한 완벽한 추적성이 보장됩니다.
