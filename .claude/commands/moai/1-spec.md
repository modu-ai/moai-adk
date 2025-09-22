---
name: moai:1-spec
description: EARS 형식 명세 작성 - 비즈니스 요구사항을 구조화된 명세로 변환
argument-hint: <feature-description>|<SPEC-ID>|--project [additional-details...]
allowed-tools: Read, Write, Edit, MultiEdit, Bash(git:*), Bash(gh:*), Bash(ls:*), Bash(mkdir:*), Bash(python3:*), Task, Grep, Glob
---

# MoAI-ADK 1단계: SPEC 작성

spec-builder 에이전트를 활용해 비즈니스 요구사항을 EARS 형식 명세로 변환하고, GitFlow 자동화를 지원합니다.

## 빠른 시작

```bash
# 단일 기능 SPEC 작성
/moai:1-spec "JWT 기반 사용자 인증 시스템"

# 프로젝트 전체 SPEC 대화형 생성
/moai:1-spec --project

# 기존 SPEC 수정
/moai:1-spec SPEC-001 "추가 보안 요구사항"
```

## GitFlow 워크플로우 (자동화)

### 현재 상태 확인
- Current branch: !`git branch --show-current`
- Git status: !`git status --porcelain`
- Existing SPECs: !`ls .moai/specs/ 2>/dev/null | wc -l`

### 브랜치 전략

**단일 SPEC 모드**:
1. develop/main 브랜치로 전환
2. SPEC-XXX ID 자동 할당
3. feature/SPEC-XXX-{name} 브랜치 생성
4. 4단계 구조화 커밋
5. Draft PR 자동 생성 (gh CLI)

**--project 모드**:
1. 통합 브랜치 생성: feature/project-{timestamp}
2. 5단계 대화형 질문으로 다중 SPEC 생성
3. 각 SPEC별 순차 커밋
4. 단일 통합 PR 생성

## EARS 명세 구조

### 핵심 키워드 활용
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
```

## 프로젝트 모드 (--project)

### 대화형 5단계 질문
1. **프로젝트 유형**: 웹앱, API, 모바일앱, 데스크톱앱
2. **핵심 기능**: 사용자 관리, 결제, 알림, 콘텐츠 관리
3. **사용자 유형**: 일반 사용자, 관리자, 게스트
4. **성능 요구사항**: 응답시간, 동시 접속자, 처리량
5. **보안 요구사항**: 인증 방식, 개인정보, 규정 준수

### 생성 결과
```markdown
🏢 프로젝트 SPEC 통합 생성 완료:

🌿 브랜치: feature/project-20250119-initial-specs
├── SPEC-001: 사용자 인증 시스템 (P0) ✓
├── SPEC-002: 게시글 관리 시스템 (P0) ✓
├── SPEC-003: 댓글 및 좋아요 (P1) ✓
├── SPEC-004: 관리자 대시보드 (P1) ✓
└── SPEC-005: 모니터링 시스템 (P2) ✓

🎯 다음: /moai:2-build SPEC-001
```

## User Stories & 수락 기준

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

## 4단계 커밋 패턴

1. **SPEC 초안**: `📝 SPEC-001: 초기 요구사항 명세 작성`
2. **User Stories**: `📖 SPEC-001: User Stories 및 시나리오 추가`
3. **수락 기준**: `✅ SPEC-001: 수락 기준 및 테스트 시나리오 정의`
4. **최종 완성**: `🎯 SPEC-001: 명세 완성 및 프로젝트 구조 생성`

## 품질 검증

### 검증 기준
- 모든 User Story에 수락 기준 존재 (최소 3개)
- EARS 요구사항의 테스트 가능성 확인
- [NEEDS CLARIFICATION] 마커 10% 이하
- @REQ 태그를 통한 추적성 완성

### 검증 결과 예시
```markdown
📊 SPEC 품질 지표:
- User Stories: 12개 생성
- EARS 요구사항: 35개
- 수락 기준: 36개 시나리오
- 명확성 점수: 94%
- 추적성 매트릭스: 완료
```

## 완료 후 다음 단계

```bash
✅ 1단계 SPEC 작성 + GitFlow 완료!

🔀 Git 작업 (자동 시도):
├── feature/SPEC-001-user-auth 브랜치 생성
├── 4단계 커밋 완료
└── Draft PR #123 생성: "SPEC-001: 사용자 인증 시스템"

📁 생성된 파일:
└── .moai/specs/SPEC-001/spec.md

🎯 다음 단계:
> /moai:2-build SPEC-001  # TDD 구현
> /moai:3-sync           # 문서 동기화
```

## 에러 처리

### Git index.lock 감지
```bash
원인: 이전 git 명령 비정상 종료
해결: rm -f .git/index.lock 후 재실행
```

### 불완전한 입력
```bash
⚠️ 더 구체적인 요구사항 필요
예: /moai:1-spec "JWT 인증 - 소셜 로그인, 토큰 갱신, 권한 관리"
```

모든 SPEC 작성은 Constitution 5원칙을 준수하며, 16-Core TAG 시스템으로 완전한 추적성을 보장합니다.