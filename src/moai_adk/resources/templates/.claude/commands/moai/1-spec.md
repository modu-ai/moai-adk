---
description: EARS 형식 명세 작성 - 비즈니스 요구사항을 구조화된 명세로 변환
argument-hint: <feature-description>|<SPEC-ID>|--project [additional-details...]
allowed-tools: Read, Write, Edit, MultiEdit, Bash, Task
---

# MoAI-ADK  SPEC 작성 (GitFlow 통합)

!@ spec-builder 에이전트를 활용해 비즈니스 요구사항을 EARS 형식의 엔지니어링 명세로 변환하고 GitFlow 워크플로우를 자동 통합합니다.

## 🔀 GitFlow 자동화 실행 코드 (완전 투명)

#### 현재 Git 상태 확인
- Current branch: !`git branch --show-current`
- Git status: !`git status --porcelain`
- Remote branches: !`git branch -r | head -5`
- Last commits: !`git log --oneline -3`
- Existing SPECs: !`ls .moai/specs/ 2>/dev/null | wc -l`

#### GitFlow 전략 (모드별)

**단일 SPEC 모드**:
1. **기본 브랜치 전환**: develop 또는 main으로 자동 전환
2. **SPEC ID 자동 할당**: 기존 SPEC 개수 기반 증가
3. **개별 피처 브랜치**: feature/SPEC-XXX-{feature-name}
4. **4단계 자동 커밋**: SPEC → Stories → 수락기준 → 완성
5. **개별 Draft PR**: 단일 기능에 대한 PR

**--project 모드**:
1. **기본 브랜치 전환**: develop 또는 main으로 자동 전환
2. **통합 브랜치 생성**: feature/project-{timestamp}-initial-specs
3. **SPEC별 순차 커밋**: 각 SPEC마다 개별 커밋
4. **단일 Draft PR**: 전체 프로젝트 명세에 대한 PR

비즈니스 요구사항을 EARS(Easy Approach to Requirements Syntax) 형식의 엔지니어링 명세로 변환합니다. **GitFlow가 완전히 통합**되어 버전 관리가 자동으로 처리됩니다.

## 🚀 빠른 시작

```bash
# 단일 기능 SPEC 작성
/moai:1-spec "JWT 기반 사용자 인증 시스템"

# 프로젝트 전체 SPEC 대화형 생성 (신규 🔥)
/moai:1-spec --project

# 기존 SPEC 수정
/moai:1-spec SPEC-001 "추가 보안 요구사항"
```

## 🔄 실행 흐름 (GitFlow 자동화)

```mermaid
flowchart TD
    A[/moai:1-spec 입력] --> A1[🔀 feature 브랜치 자동 생성]
    A1 --> B{입력 타입 분석}
    B -->|--project| C[대화형 프로젝트 분석]
    B -->|feature-description| E[단일 SPEC 작성]
    B -->|SPEC-ID| F[기존 SPEC 수정]

    C --> G[5단계 질문/답변]
    E --> I[EARS 형식 변환]
    F --> I

    G --> J[다중 SPEC 생성]
    I --> K[품질 검증]
    J --> K
    K --> K1[📝 SPEC 파일 자동 커밋]
    K1 --> K2[🔄 Draft PR 생성]
    K2 --> L[완료: /moai:2-build로 이동]
```

## 🔀 GitFlow 자동 통합 

### 완전 투명한 버전 관리
사용자는 Git 명령어를 전혀 알 필요 없이 모든 버전 관리가 자동 처리됩니다.

#### 1. Feature 브랜치 자동 생성
```bash
# 자동 실행 (사용자 불가시)
git checkout -b feature/SPEC-001-user-auth
git push -u origin feature/SPEC-001-user-auth
```

#### 2. 단계별 자동 커밋
- **SPEC 초안**: `feat(SPEC-001): Add initial requirements draft`
- **User Stories**: `feat(SPEC-001): Add user stories US-001~005`
- **수락 기준**: `feat(SPEC-001): Add acceptance criteria scenarios`
- **최종 검토**: `feat(SPEC-001): Complete EARS specification`

#### 3. Draft PR 자동 생성
```bash
# gh CLI로 자동 실행
gh pr create --draft \
  --title "SPEC-001: 사용자 인증 시스템" \
  --body "🧭 SPEC 작성 완료

## 📋 생성된 파일
- spec.md (12개 요구사항)
- user-stories.md (5개 스토리)
- acceptance.md (15개 시나리오)

## 🎯 다음 단계
/moai:2-build SPEC-001"
```

### 브랜치 전략
- **feature/SPEC-XXX-{feature-name}**: SPEC 작성용 브랜치
- **develop**: 통합 브랜치 (자동 머지 대상)
- **main**: 릴리스 브랜치 (수동 머지)

## 🤖 spec-builder 에이전트 자동화

**spec-builder 에이전트**가 전체 SPEC 작성 + GitFlow 과정을 완전 자동화:

### 💯 순차 처리 최적화
- **단일 SPEC**: 개별 브랜치 전략 (spec-builder 에이전트 1개)
- **--project 다중 SPEC**: **통합 브랜치 순차 실행**
  - 5개 SPEC → 단일 브랜치에 순차 커밋
  - Git 충돌 0%, 안정성 100%
  - 초보자 친화적 경험

### 에이전트 기능
- **요구사항 분석**: EARS 키워드(WHEN/IF/WHILE/WHERE/UBIQUITOUS) 구조화
- **User Story 생성**: US-XXX 형식의 체계적 사용자 스토리
- **수락 기준 작성**: Given-When-Then 테스트 가능 기준
- **Git 워크플로우**: 브랜치 생성 → 커밋 → PR 생성까지 전자동

### 순차 실행 모델
```markdown
# --project 모드: 대화형 5단계 질문
1. 프로젝트 유형 선택
2. 핵심 기능 정의
3. 사용자 유형 분류
4. 성능 요구사항
5. 보안 요구사항

→ 자동으로 3-5개 SPEC 순차 생성
→ 통합 브랜치에 각각 커밋
→ 단일 Draft PR 생성
```

## 📋 --project 옵션 (대화형 프로젝트 SPEC) 🔥

### 기능 개요
5단계 질문을 통해 프로젝트 전체를 분석하고 모든 핵심 기능의 SPEC을 자동 생성합니다.

### 질문 단계
1. **프로젝트 유형**: 웹앱, API, 모바일앱, 데스크톱앱
2. **핵심 기능**: 사용자 관리, 결제, 알림, 콘텐츠 관리 등
3. **사용자 유형**: 일반 사용자, 관리자, 게스트 등
4. **성능 요구사항**: 응답시간, 동시 접속자, 처리량
5. **보안 요구사항**: 인증 방식, 개인정보, 규정 준수

### 생성 결과 (💯 통합 브랜치 전략)

#### 단일 SPEC vs --project 모드 비교
- **단일 SPEC**: 개별 브랜치 + 개별 PR (2-3분/SPEC)
- **--project 모드**: 통합 브랜치 + 단일 PR (5개 SPEC 순차 생성 시 8-10분)

#### --project 모드 장점
```markdown
🏢 프로젝트 SPEC 통합 생성 완료:

🌿 브랜치: feature/project-20250119-initial-specs
├── SPEC-001: 사용자 인증 시스템 (P0) ✓
├── SPEC-002: 게시글 관리 시스템 (P0) ✓
├── SPEC-003: 댓글 및 좋아요 (P1) ✓
├── SPEC-004: 관리자 대시보드 (P1) ✓
└── SPEC-005: 모니터링 시스템 (P2) ✓

✨ 장점:
- 하나의 브랜치로 간단 관리
- Git 충돌 위험 0%
- 전체 프로젝트 일관성
- 단계별 구현 가능

🎯 다음: /moai:2-build SPEC-001 (첫 번째부터 순차 구현)
```

#### --project 명령어 예시
```bash
# 대화형 프로젝트 분석로 5개 SPEC 생성
/moai:1-spec --project

# 바로 답변하는 경우
Q: 프로젝트 유형? A: 웹앱
Q: 핵심 기능? A: 로그인, 게시판, 댓글
Q: 사용자 유형? A: 일반, 관리자
...
```

## 📝 EARS 형식 변환 자동화

### EARS 키워드 활용
- **WHEN**: 조건 발생 시 → 명확한 트리거 정의
- **IF**: 특정 상태 → 조건부 동작 정의
- **WHILE**: 진행 중 → 지속적 처리 정의
- **WHERE**: 특정 환경 → 컨텍스트별 동작
- **UBIQUITOUS**: 항상 → 전역 규칙 정의

### 변환 예시
**Before**: "사용자가 로그인할 수 있어야 한다"

**After**:
```markdown
WHEN 사용자가 올바른 이메일과 패스워드를 입력하면,
시스템은 3초 이내에 JWT 토큰을 생성하고 대시보드로 리디렉션해야 한다.

IF 잘못된 인증 정보가 3회 연속 입력되면,
시스템은 해당 계정을 15분간 잠그고 관리자에게 알림을 발송해야 한다.
```

## 📋 User Stories & 수락 기준 생성

### US-XXX 형식 템플릿
```markdown
US-001: 사용자 로그인
As a 일반 사용자
I want to 이메일과 패스워드로 로그인
So that 개인화된 서비스를 이용할 수 있다

수락 기준:
- 올바른 이메일 형식 검증
- 패스워드 최소 8자리 이상
- 3회 실패 시 계정 임시 잠금
- 성공 시 대시보드 리다이렉트
```

### Given-When-Then 시나리오
```markdown
**Scenario: 성공적인 로그인**
Given 등록된 사용자 "user@example.com"이 존재하고
When 올바른 이메일과 패스워드를 입력하고 "로그인"을 클릭하면
Then 3초 이내에 JWT 토큰을 생성하고
  And 대시보드로 리디렉션하며
  And "환영합니다, [사용자명]님" 메시지를 표시한다
```

## 📁 생성 파일 구조

```
.moai/specs/SPEC-XXX/
└── spec.md           # EARS 요구사항 + User Stories + 수락 기준 통합
```

## ✅ 품질 검증 자동화

### 검증 기준
- 모든 User Story에 수락 기준 존재 (최소 3개)
- EARS 요구사항의 테스트 가능성 확인
- [NEEDS CLARIFICATION] 마커 10% 이하
- @REQ 태그를 통한 추적성 보장

### 검증 결과
```markdown
📊 SPEC 품질 지표:
- User Stories: 12개 생성
- EARS 요구사항: 35개
- 수락 기준: 36개 시나리오
- 명확성 점수: 94%
- 추적성 매트릭스: 100% 완료
```

## 🔄 완료 후 다음 단계

###  GitFlow 통합 워크플로우
```bash
✅ SPEC 작성 + GitFlow 완료!

🔀 Git 작업 (자동 완료):
├── feature/SPEC-001-user-auth 브랜치 생성
├── 4단계 커밋 완료 (SPEC → Stories → 수락기준 → 최종)
└── Draft PR #123 생성: "SPEC-001: 사용자 인증 시스템"

📁 생성된 파일:
└── .moai/specs/SPEC-001/
    └── spec.md (EARS 요구사항 + User Stories + 수락 기준 통합)

🎯 다음 단계 (2단계 파이프라인):
> /moai:2-build SPEC-001  # TDD 구현 (자동 PR 업데이트)
> /moai:3-sync           # 문서 동기화 + PR Ready
```

## ⚠️ 에러 처리

### Steering 문서 누락
```bash
❌ Steering 문서가 필요합니다.
먼저: moai init 실행하여 프로젝트 초기화
```

### 불완전한 입력
```bash
⚠️ 더 구체적인 요구사항 필요:
예: /moai:1-spec "JWT 인증 - 소셜 로그인, 토큰 갱신, 권한 관리"
```

### [NEEDS CLARIFICATION] 미해결
```bash
🔴 명세에 불명확한 부분이 있습니다.
모든 명확화 마커 해결 후 /moai:2-build 진행 가능
```

## 🔁 응답 구조

출력은 반드시 3단계 구조를 따릅니다:
1. **Phase 1 Results**: SPEC 생성 결과
2. **Phase 2 Plan**: 다음 단계 계획
3. **Phase 3 Implementation**: 구체적 실행 안내

이 명령어는 MoAI-ADK 0.2.0의 핵심으로, 단순화된 인터페이스로 강력한 SPEC 작성을 제공합니다.
