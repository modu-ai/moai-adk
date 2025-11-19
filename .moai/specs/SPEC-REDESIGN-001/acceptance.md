---
spec_id: SPEC-REDESIGN-001
title: 프로젝트 설정 스키마 v3.0.0 - 수용 기준 및 테스트 시나리오
version: 1.0.0
created_at: 2025-11-19T00:00:00Z
---

# 수용 기준 및 테스트 시나리오: SPEC-REDESIGN-001

## 개요

본 문서는 SPEC-REDESIGN-001 구현 완료 후 수용 기준(Acceptance Criteria) 및 상세 테스트 시나리오를 정의합니다.

---

## 수용 기준 (Acceptance Criteria)

### AC-001: Quick Start 모드 설정 시간 (2-3분)

**Requirement**: /moai:0-project 실행 후 Quick Start 모드 선택 시 2-3분 내 설정 완료

**Given**: 신규 사용자가 /moai:0-project 명령어 실행
**When**: Tab 1 (10개 질문) 응답 → Tab 2 "Quick Start - Skip for Now" 선택 → Tab 3 Git 설정 완료
**Then**:
- ✅ 전체 설정 소요 시간: 2-3분 이내
- ✅ config.json 생성: 10개 사용자 입력 + 16개 smart defaults + 5개 auto-detected = 31개 설정값
- ✅ .moai/project/ 디렉토리 생성: product.md, structure.md, tech.md 빈 템플릿 포함
- ✅ 사용자가 즉시 /moai:1-plan 실행 가능 (개발 준비 완료)

**검증 방법**:
1. 신규 프로젝트에서 /moai:0-project 실행
2. 각 배치 실행 시간 기록
3. 전체 소요 시간 측정
4. config.json 생성 확인
5. 빈 템플릿 파일 확인

**성공 기준**:
- 측정된 시간 ≤ 3분 (여유 포함)
- config.json 파일 존재 및 유효
- 31개 설정값 모두 포함

---

### AC-002: Full Documentation + Standard 모드 (15-20분)

**Requirement**: Full Documentation + Standard 깊이 선택 시 15-20분 내 완료

**Given**: 신규 사용자가 Full Documentation 선택, Brainstorming 깊이 "Standard"(15분) 선택
**When**: Brainstorm 질문 10-15개 응답 및 AI 분석 완료
**Then**:
- ✅ 예상 소요 시간: 15-20분 이내
- ✅ product.md 생성:
  - 프로젝트 비전 (Vision Statement)
  - 타겟 사용자 (Target Users)
  - 가치 제안 (Value Proposition)
  - 로드맵 (Roadmap)
  - AI 분석 기반 인사이트 포함
- ✅ structure.md 생성:
  - 시스템 아키텍처 설명
  - 주요 컴포넌트 (Core Components)
  - 컴포넌트 간 관계도
  - 의존성 (Dependencies)
  - 옵션: ASCII 또는 Markdown 다이어그램
- ✅ tech.md 생성:
  - 기술 스택 선택 (Tech Stack Selection)
  - Trade-offs 분석
  - 성능 고려사항
  - 보안 고려사항
  - 초기 설정 가이드
- ✅ 3개 문서 모두 .moai/project/ 에 저장
- ✅ project-manager/tdd-implementer/experts에 컨텍스트 자동 로드

**검증 방법**:
1. Full Documentation + Standard 선택 후 진행
2. 브레인스토밍 세션 시간 기록
3. AI 분석 처리 시간 기록
4. 전체 소요 시간 측정
5. 3개 문서 생성 확인
6. 문서 품질 검증
7. 에이전트 컨텍스트 로드 확인

**성공 기준**:
- 전체 소요 시간 ≤ 20분
- 3개 문서 모두 생성됨
- 각 문서 최소 200자 이상
- AI 분석 포함 확인
- 에이전트 로드 성공

---

### AC-003: 질문 감소율 63% (27개 → 10개)

**Requirement**: Tab 1 설정 질문이 10개로 제한됨 (기존 27개 대비 63% 감소)

**Given**: Tab 1 스키마 로드
**When**: 모든 배치 계산
**Then**:
- ✅ Tab 1 essential 질문: 10개
  - Batch 1.1 (Identity & Language): 3개
  - Batch 1.2 (Project Basics): 3개
  - Batch 1.3 (Development Mode): 2개
  - Batch 1.4 (Quality Standards): 2개
- ✅ Tab 2 선택 질문: 1개 (필수)
- ✅ Tab 2 조건부 질문: 1개 (full_now 선택 시만)
- ✅ Tab 3 조건부 질문: 0-4개 (모드에 따라)
- ✅ 전체 사용자 질문: 10-16개 (선택과 모드에 따라)
- ✅ 기존 27개 대비 63% 감소 달성

**검증 방법**:
1. tab_schema.json 분석
2. 모든 배치에서 질문 개수 계산
3. 기존 v2.x와 비교
4. 질문 감소율 계산

**성공 기준**:
- Tab 1 질문 = 10개 정확히
- 조건부 질문 포함 시 전체 ≤ 16개
- 감소율 ≥ 63%

---

### AC-004: 100% 설정 커버리지 (31개 설정값)

**Requirement**: 모든 31개 config.json 설정값이 완전히 커버됨

**Given**: /moai:0-project 완료, config.json 생성
**When**: config.json 파일 검증
**Then**:
- ✅ 10개 사용자 입력 설정:
  1. user.name
  2. language.conversation_language
  3. language.agent_prompt_language
  4. project.name
  5. project.owner
  6. project.description
  7. git_strategy.mode
  8. git_strategy.{mode}.workflow
  9. constitution.test_coverage_target
  10. constitution.enforce_tdd
- ✅ 1개 필수 선택 설정:
  11. project.documentation_mode
- ✅ 1개 조건부 설정:
  12. project.documentation_depth
- ✅ 4개 조건부 Git 설정:
  13. git_strategy.personal.auto_checkpoint
  14. git_strategy.personal.push_to_remote
  15. git_strategy.team.auto_pr
  16. git_strategy.team.draft_pr
- ✅ 5개 Auto-Detect 설정:
  17. project.language
  18. project.locale
  19. language.conversation_language_name
  20. project.template_version
  21. moai.version
- ✅ 16개 Smart Defaults (자동 적용):
  22-31. (Git 모드별, 프리셋별 설정)
- ✅ 총 31개 = 100% 커버리지

**검증 방법**:
1. config.json 파일 로드
2. 각 필드 존재 여부 확인
3. 각 필드 값 타입 검증
4. 필수 필드 검증
5. Optional 필드 검증

**성공 기준**:
- 31개 설정값 모두 존재
- 모든 값 올바른 타입
- 필수 필드 모두 채워짐
- 커버리지 = 100%

---

### AC-005: 조건부 배치 렌더링

**Requirement**: git_strategy.mode에 따라 Tab 3 배치가 올바르게 표시/숨겨짐

**Given**: Tab 1 완료, git_strategy.mode 선택 완료
**When**: Tab 3으로 진행
**Then**:

#### Scenario 5a: Personal 모드
- ✅ Tab 3.1_personal 배치 표시 (2개 질문)
  - Auto Checkpoint
  - Push to Remote
- ✅ Tab 3.1_team 배치 숨겨짐 (Auto PR, Draft PR 미표시)
- ✅ 불필요한 질문 0개

#### Scenario 5b: Team 모드
- ✅ Tab 3.1_team 배치 표시 (2개 질문)
  - Auto PR
  - Draft PR
- ✅ Tab 3.1_personal 배치 숨겨짐 (Auto Checkpoint, Push to Remote 미표시)
- ✅ 불필요한 질문 0개

#### Scenario 5c: Hybrid 모드
- ✅ Tab 3.1_personal 배치 표시 (현재는 Personal 활성)
- ✅ Tab 3.1_team 배치 동적 토글 옵션 (팀 협업 시작 시 활성화)
- ✅ 모드 전환 가능

**검증 방법**:
1. 각 모드(Personal/Team/Hybrid) 선택
2. Tab 3 진행
3. 표시되는 배치 확인
4. 조건부 로직 검증

**성공 기준**:
- Personal: Personal 배치만 표시
- Team: Team 배치만 표시
- Hybrid: Personal 기본 + 동적 전환 가능

---

### AC-006: Smart Defaults 자동 적용

**Requirement**: 16개 설정값에 smart defaults 자동 적용

**Given**: Tab 1-2 사용자 입력 완료
**When**: config.json 저장
**Then**:
- ✅ 다음 설정값들이 사용자 입력 없이 자동 적용:
  1. git_strategy.personal.workflow (기본값: "github-flow")
  2. git_strategy.team.workflow (기본값: "git-flow")
  3. git_strategy.personal.auto_checkpoint (기본값: "event-driven")
  4. git_strategy.personal.push_to_remote (기본값: false)
  5. git_strategy.team.auto_pr (기본값: false)
  6. git_strategy.team.draft_pr (기본값: false)
  7. 기타 10개 smart defaults (상황별)
- ✅ 기본값 모두 선택적(conditional)이거나 권장값

**검증 방법**:
1. 사용자 입력 최소화하고 저장
2. config.json 확인
3. 각 기본값 검증

**성공 기준**:
- 16개 설정값 모두 자동 적용됨
- 모든 기본값 적절함

---

### AC-007: Auto-Detect 필드 자동 감지

**Requirement**: 5개 필드가 자동으로 감지 및 설정됨

**Given**: /moai:0-project 실행
**When**: Auto-Detect 프로세스 실행
**Then**:
- ✅ project.language: 코드베이스 분석하여 감지
  - Python 프로젝트 → "python"
  - TypeScript 프로젝트 → "typescript"
- ✅ project.locale: conversation_language에서 자동 매핑
  - "ko" → "ko_KR"
  - "en" → "en_US"
  - "ja" → "ja_JP"
- ✅ language.conversation_language_name: 언어 코드 → 이름 변환
  - "ko" → "Korean"
  - "en" → "English"
  - "ja" → "Japanese"
- ✅ project.template_version: 시스템 관리값
  - 현재 MoAI-ADK 템플릿 버전으로 설정
- ✅ moai.version: 시스템 관리값
  - 현재 MoAI-ADK 버전으로 설정

**검증 방법**:
1. /moai:0-project 실행
2. config.json 생성
3. Auto-Detect 필드 확인
4. 값 정확성 검증

**성공 기준**:
- 5개 필드 모두 감지됨
- 모든 값 정확함
- 사용자 개입 불필요

---

### AC-008: 설정값 저장 원자성 (All-or-Nothing)

**Requirement**: config.json 저장이 원자적으로 수행됨 (부분 저장 없음)

**Given**: 모든 Tab 완료, config.json 저장 대기 중
**When**: 저장 프로세스 시작
**Then**:
- ✅ 전체 31개 설정값 동시 저장
- ✅ 부분 저장 상태 절대 없음
- ✅ 저장 실패 시 기존 config.json 유지
- ✅ 저장 성공 시 .moai/config/config.json 최신화
- ✅ 트랜잭션 로그 기록

**검증 방법**:
1. 저장 프로세스 모니터링
2. 중간에 프로세스 중단 시도
3. config.json 상태 확인
4. 트랜잭션 로그 확인

**성공 기준**:
- 모든 또는 없음 (부분 저장 불가)
- 실패 시 롤백 완료
- 성공 시 모든 값 저장됨

---

### AC-009: Template 변수 동적 해석

**Requirement**: {{user.name}}, {{project.owner}} 등 변수가 Runtime에 해석됨

**Given**: config.json 저장, template 변수 사용
**When**: project-manager 에이전트 로드
**Then**:
- ✅ {{user.name}} → "GOOS행" (실제 사용자명)
- ✅ {{project.owner}} → "GoosLab" (설정된 owner)
- ✅ {{project.name}} → "MoAI-ADK" (설정된 프로젝트명)
- ✅ 모든 변수 완전히 해석됨
- ✅ Agent task_prompt에 구체적 값 전달

**검증 방법**:
1. config.json에 template 변수 포함된 설정 생성
2. project-manager 로드
3. Task prompt 확인
4. 변수 해석 결과 검증

**성공 기준**:
- 모든 변수 올바르게 해석됨
- 에이전트가 구체적 값 수신

---

### AC-010: AI 에이전트 컨텍스트 자동 로드

**Requirement**: 생성된 문서가 해당 에이전트에 자동으로 로드됨

**Given**: Full Documentation 선택, 3개 문서 생성 완료
**When**: 각 에이전트 로드
**Then**:
- ✅ project-manager 로드: product.md 자동 로드
  - system_context에 추가됨
  - 모든 task에 자동 주입됨
  - 프로젝트 비전 이해 가능
- ✅ tdd-implementer 로드: structure.md 자동 로드
  - architecture reference로 사용됨
  - SPEC 구현 시 참고 가능
- ✅ backend-expert 로드: tech.md 자동 로드
  - 기술 스택 정보 접근 가능
  - 라이브러리 선택 시 참고
- ✅ frontend-expert 로드: tech.md 자동 로드
  - 기술 스택 정보 접근 가능
  - 프레임워크 선택 시 참고

**검증 방법**:
1. Full Documentation + Standard 완료
2. 각 에이전트 로드
3. system_context/task_prompt 확인
4. 문서 컨텐츠 확인

**성공 기준**:
- 3개 에이전트 모두 해당 문서 로드
- 컨텍스트 자동 주입됨
- 에이전트가 문서 정보 활용 가능

---

### AC-011: Backward Compatibility (v2.1.0 마이그레이션)

**Requirement**: v2.1.0 config.json이 자동 마이그레이션되며 호환성 유지

**Given**: v2.1.0 config.json 기존 파일 존재
**When**: /moai:0-project 다시 실행
**Then**:
- ✅ 기존 설정값 자동 감지
- ✅ Tab 1에서 현재값 표시:
  - user.name: "GOOS행" 표시
  - git_strategy.mode: "hybrid" 표시
  - 기타 설정값 모두 표시
- ✅ 마이그레이션 옵션 제시:
  - "Keep Current": 기존 값 유지
  - "Update": 새로운 값 입력
- ✅ v3.0.0 schema로 업그레이드
- ✅ 호환되지 않는 필드는 smart default로 마이그레이션

**검증 방법**:
1. v2.1.0 config.json 준비
2. /moai:0-project 실행
3. 마이그레이션 과정 확인
4. 업그레이드된 config.json 검증

**성공 기준**:
- 기존 설정값 모두 유지
- 호환되지 않는 필드 자동 처리
- 업그레이드 성공

---

### AC-012: AskUserQuestion API 제약 준수

**Requirement**: AskUserQuestion API 제약을 모두 준수

**Given**: tab_schema.json v3.0.0
**When**: 모든 배치 분석
**Then**:
- ✅ 배치당 최대 4개 질문 준수
  - Batch 1.1: 3개 (< 4) ✅
  - Batch 1.2: 3개 (< 4) ✅
  - Batch 1.3: 2개 (< 4) ✅
  - Batch 1.4: 2개 (< 4) ✅
  - Batch 2.1: 1개 (< 4) ✅
  - Batch 2.2: 1개 (< 4) ✅
  - 기타 모든 배치: ≤ 4개 ✅
- ✅ 이모지 불포함:
  - 질문 필드: 이모지 없음 ✅
  - 헤더 필드: 이모지 없음 ✅
- ✅ Header 길이 ≤ 12자:
  - 모든 header 검증
- ✅ 옵션 2-4개:
  - 모든 질문 2개 이상 4개 이하 옵션

**검증 방법**:
1. JSON Schema 검증
2. 배치당 질문 수 계산
3. 이모지 검색
4. Header 길이 검산
5. 옵션 개수 확인

**성공 기준**:
- 모든 API 제약 준수
- 검증 에러 0개

---

### AC-013: 설정 완료 후 개발 즉시 시작 가능

**Requirement**: Quick Start 완료 후 /moai:1-plan 명령어 즉시 실행 가능

**Given**: Quick Start 설정 완료
**When**: /moai:1-plan 명령어 실행
**Then**:
- ✅ config.json이 모든 필요한 정보 포함
- ✅ project-manager가 프로젝트 정보 접근 가능
- ✅ /moai:1-plan이 에러 없이 실행됨
- ✅ SPEC 생성 시작 가능
- ✅ 추가 설정 불필요

**검증 방법**:
1. Quick Start 완료
2. /moai:1-plan "테스트 기능" 실행
3. 실행 성공/실패 확인
4. 에러 메시지 확인

**성공 기준**:
- /moai:1-plan 성공 실행
- 에러 메시지 0개
- SPEC 생성 시작

---

## 테스트 시나리오 및 Given-When-Then

### 시나리오 1: 완전한 Quick Start 워크플로우 (2-3분)

**시나리오 이름**: Quick Start - 신규 사용자 신규 프로젝트

**Background**:
- 신규 사용자 (첫 프로젝트)
- /moai:0-project 명령어 미경험
- 빠른 개발 시작이 목표

**Scenario S1a: Tab 1 - Identity & Language 배치**

Given:
- /moai:0-project 실행
- Tab 1 시작

When:
- Batch 1.1 질문 3개 응답:
  1. "What name should Alfred use?" → "철수"
  2. "Conversation language?" → "Korean (ko)"
  3. "Agent prompt language?" → "English"

Then:
- ✅ Batch 1.1 완료 (3개 질문 응답)
- ✅ 예상 시간: 30-45초
- ✅ Batch 1.2로 자동 진행

**Scenario S1b: Tab 1 - Project Basics 배치**

Given:
- Batch 1.1 완료
- Batch 1.2 시작

When:
- Batch 1.2 질문 3개 응답:
  1. "Project Name?" → "MyApp"
  2. "Project Owner?" → "철수" (자동 제시)
  3. "Project Description?" → "Skip"

Then:
- ✅ Batch 1.2 완료 (3개 질문 응답)
- ✅ 예상 시간: 30-45초
- ✅ Batch 1.3으로 자동 진행

**Scenario S1c: Tab 1 - Development Mode 배치**

Given:
- Batch 1.2 완료
- Batch 1.3 시작

When:
- Batch 1.3 질문 2개 응답:
  1. "Git workflow mode?" → "Hybrid"
  2. "Branching workflow type?" → "GitHub Flow"

Then:
- ✅ Batch 1.3 완료 (2개 질문 응답)
- ✅ 예상 시간: 20-30초
- ✅ Batch 1.4로 자동 진행

**Scenario S1d: Tab 1 - Quality Standards 배치**

Given:
- Batch 1.3 완료
- Batch 1.4 시작

When:
- Batch 1.4 질문 2개 응답:
  1. "Test coverage target?" → "90" (기본값 추천)
  2. "Enforce TDD?" → "Yes" (기본값 추천)

Then:
- ✅ Batch 1.4 완료 (2개 질문 응답)
- ✅ 예상 시간: 20-30초
- ✅ Tab 1 완료, Tab 2로 진행

**Scenario S1e: Tab 2 - Documentation Choice**

Given:
- Tab 1 완료
- Tab 2 시작

When:
- Batch 2.1 질문 1개 응답:
  1. "Documentation strategy?" → "Quick Start - Skip for Now"

Then:
- ✅ Batch 2.1 완료 (1개 질문)
- ✅ Batch 2.2 조건부 숨겨짐 (full_now 미선택)
- ✅ 예상 시간: 10-15초
- ✅ Tab 3로 진행

**Scenario S1f: Tab 3 - Personal Git Settings (조건부)**

Given:
- Tab 2 완료
- git_strategy.mode = "hybrid" 선택됨
- Tab 3 시작

When:
- Batch 3.1_personal 표시됨 (hybrid = personal 포함)
- Batch 3.1_personal 질문 2개 응답:
  1. "Auto checkpoint?" → "Event-Driven" (기본값)
  2. "Push to remote?" → "No" (기본값)

Then:
- ✅ Batch 3.1_personal 완료 (2개 질문)
- ✅ Batch 3.1_team 숨겨짐 (hybrid 모드에서 나중 활성화 가능)
- ✅ 예상 시간: 20-30초
- ✅ Tab 3 완료

**Scenario S1g: 최종 설정 저장 및 완료**

Given:
- 모든 Tab 완료
- 사용자 입력: 10개

When:
- "완료" 버튼 클릭
- config.json 저장 프로세스 시작

Then:
- ✅ config.json 생성:
  - 10개 사용자 입력
  - 16개 smart defaults
  - 5개 auto-detected
  - 총 31개 설정값
- ✅ .moai/project/ 디렉토리 생성
- ✅ product.md, structure.md, tech.md 빈 템플릿 생성
- ✅ 저장 완료 메시지 표시
- ✅ 예상 시간: 10-15초
- ✅ 전체 소요 시간: 2-3분

**Scenario S1 최종 검증**:
- ✅ 설정 시간: 2-3분 달성
- ✅ 질문 개수: 10개 (기본) + 1개 (문서 선택) = 11개
- ✅ config.json: 31개 설정값
- ✅ 프로젝트 준비: 개발 즉시 시작 가능

---

### 시나리오 2: Full Documentation + Standard 모드 (15-20분)

**시나리오 이름**: Full Documentation - 포괄적 프로젝트 계획

**Background**:
- 신규 스타트업 프로젝트
- 포괄적 프로젝트 문서 필요
- Architecture 설계 문서 원함
- 기술 선택 가이드 필요

**Scenario S2a: Tab 1-2 기본 설정 (3-5분)**

Given:
- /moai:0-project 실행
- Tab 1-2 진행

When:
- Tab 1 완료 (10개 질문)
- Tab 2 Batch 2.1 응답: "Full Documentation - Now"
- Tab 2 Batch 2.2 응답: "Standard (10-15 questions, ~15 min)"

Then:
- ✅ 기본 설정 완료
- ✅ Documentation mode: "full_now"
- ✅ Documentation depth: "standard"
- ✅ 예상 시간: 3-5분

**Scenario S2b: Brainstorming 세션 (10-15분)**

Given:
- Tab 2 문서 생성 옵션 설정 완료
- Brainstorming 시작

When:
- moai-project-documentation-generator Skill 위임
- Standard 깊이 Brainstorm 질문 10-15개 제시
- 사용자가 각 질문에 응답:
  1. 프로젝트 비전 (1-2분)
  2. 타겟 사용자 (1-2분)
  3. 아키텍처 개요 (2-3분)
  4. 기술 스택 (2-3분)
  5. 성능/보안 요구사항 (2-3분)
  6. 기타 Trade-off (1-2분)

Then:
- ✅ Brainstorm 질문 10-15개 응답
- ✅ 각 응답 평균 100-200자
- ✅ AI 분석 시작
- ✅ 예상 시간: 10-15분

**Scenario S2c: AI 분석 및 문서 생성 (2-3분)**

Given:
- Brainstorm 세션 완료
- 사용자 응답 수집됨

When:
- AI 분석 엔진 실행:
  - Context7 MCP: 최신 기술 트렌드 조회
  - WebSearch: 경쟁사 분석
  - Claude (Sonnet): 포괄적 분석 수행
- 3개 문서 자동 생성:
  1. product.md (프로젝트 비전 + AI 인사이트)
  2. structure.md (아키텍처 다이어그램 포함)
  3. tech.md (기술 선택 이유 + Trade-offs)

Then:
- ✅ product.md 생성:
  - Vision Statement (200+ 자)
  - Target Users (150+ 자)
  - Value Proposition (200+ 자)
  - Roadmap (200+ 자)
  - AI 분석 인사이트 포함
- ✅ structure.md 생성:
  - System Architecture (300+ 자)
  - Core Components (250+ 자)
  - Dependencies (150+ 자)
  - 옵션: 아스키 다이어그램
- ✅ tech.md 생성:
  - Technology Selection (250+ 자)
  - Trade-offs Analysis (250+ 자)
  - Performance Considerations (150+ 자)
  - Security Considerations (150+ 자)
  - Setup Guide (200+ 자)
- ✅ 3개 문서 저장: .moai/project/
- ✅ 예상 시간: 2-3분

**Scenario S2d: 에이전트 컨텍스트 자동 로드**

Given:
- 3개 문서 생성 완료
- 문서가 .moai/project/에 저장됨

When:
- 에이전트 초기화:
  - project-manager 로드
  - tdd-implementer 로드
  - backend-expert 로드
  - frontend-expert 로드

Then:
- ✅ product-manager.system_context: product.md 로드
- ✅ tdd-implementer.context: structure.md 로드
- ✅ backend-expert.context: tech.md 로드
- ✅ frontend-expert.context: tech.md 로드
- ✅ 컨텍스트 자동 주입 완료

**Scenario S2e: 최종 설정 및 완료**

Given:
- AI 분석 완료
- 에이전트 컨텍스트 로드 완료
- Tab 3 Git 설정 대기 중

When:
- Tab 3 진행
- Git 설정 완료 (2개 질문, 30초)
- "완료" 버튼 클릭

Then:
- ✅ config.json 저장
- ✅ 3개 프로젝트 문서 저장
- ✅ 완료 메시지 표시
- ✅ /moai:1-plan 다음 단계 제시

**Scenario S2 최종 검증**:
- ✅ 전체 소요 시간: 15-20분 (실제)
  - 기본 설정: 3-5분
  - Brainstorm: 10-15분
  - AI 분석: 2-3분
  - Git 설정 + 저장: 1분
- ✅ 3개 문서 생성됨
- ✅ 각 문서 200+ 자 (충분한 내용)
- ✅ AI 분석 기반 인사이트 포함
- ✅ 에이전트 컨텍스트 로드 완료
- ✅ 프로젝트 준비 완료

---

### 시나리오 3: Conditional Rendering - Personal vs Team 모드

**시나리오 이름**: 조건부 배치 렌더링 검증

**Scenario S3a: Personal 모드 - 불필요한 질문 없음**

Given:
- Tab 1.3에서 git_strategy.mode = "personal" 선택

When:
- Tab 3로 진행

Then:
- ✅ Tab 3.1_personal 배치 표시:
  - "Auto checkpoint?" 표시 ✅
  - "Push to remote?" 표시 ✅
- ✅ Tab 3.1_team 배치 숨겨짐:
  - "Auto PR?" 숨겨짐 ✅
  - "Draft PR?" 숨겨짐 ✅
- ✅ 불필요한 질문 0개
- ✅ Personal 모드 기본값 자동 적용:
  - auto_checkpoint: "event-driven"
  - push_to_remote: false

**Scenario S3b: Team 모드 - Team 배치만 표시**

Given:
- Tab 1.3에서 git_strategy.mode = "team" 선택

When:
- Tab 3로 진행

Then:
- ✅ Tab 3.1_team 배치 표시:
  - "Auto PR?" 표시 ✅
  - "Draft PR?" 표시 ✅
- ✅ Tab 3.1_personal 배치 숨겨짐:
  - "Auto checkpoint?" 숨겨짐 ✅
  - "Push to remote?" 숨겨짐 ✅
- ✅ 불필요한 질문 0개
- ✅ Team 모드 기본값 자동 적용:
  - auto_pr: false
  - draft_pr: false

**Scenario S3c: Hybrid 모드 - Personal 우선, 동적 전환 가능**

Given:
- Tab 1.3에서 git_strategy.mode = "hybrid" 선택

When:
- Tab 3로 진행

Then:
- ✅ Tab 3.1_personal 배치 초기 표시:
  - "Auto checkpoint?" 표시 ✅
  - "Push to remote?" 표시 ✅
- ✅ Tab 3.1_team 동적 토글 옵션:
  - "팀 협업 시작?" 옵션 제시
  - 선택 시 Team 배치로 전환 가능
- ✅ Hybrid 모드 기본값 자동 적용:
  - auto_checkpoint: "event-driven"
  - push_to_remote: false

**Scenario S3 최종 검증**:
- ✅ Personal: Personal 배치만 2개 질문
- ✅ Team: Team 배치만 2개 질문
- ✅ Hybrid: 유연한 전환 가능
- ✅ 모든 모드에서 조건부 로직 동작

---

### 시나리오 4: Edge Case - 빠른 중단 및 부분 완성

**시나리오 이름**: Deep 모드 중도 중단 후 부분 문서 생성

**Scenario S4a: Deep 모드 선택 후 10분 경과**

Given:
- Full Documentation + Deep 모드 선택
- Brainstorm 질문 20-25개 중 10개만 응답
- 사용자가 "완료" 버튼 클릭

When:
- 부분 응답으로 문서 생성 시작
- AI 분석 실행 (응답한 10개 기반)

Then:
- ✅ 부분 완성 상태로 문서 생성:
  - product.md: 부분 작성됨 (100-150자)
  - structure.md: 기본 정보만 포함
  - tech.md: 부분 작성됨
- ✅ "[작성 중]" 마크 표시
- ✅ 사용자가 나중에 완성 가능
- ✅ 완료 메시지 명확히 표시

**Scenario S4 최종 검증**:
- ✅ 중단 시에도 부분 문서 생성
- ✅ "작성 중" 상태 명확
- ✅ 나중에 재개 가능

---

## 테스트 실행 계획

### 테스트 단계

| 단계 | 테스트 유형 | 테스트 케이스 | 예상 기간 | 담당 |
|-----|-----------|----------|---------|------|
| 1 | 단위 테스트 | 스키마 검증, API 제약 검증 | 2시간 | backend-expert |
| 2 | 통합 테스트 | 명령어 실행, 에이전트 로드 | 3시간 | backend-expert |
| 3 | 기능 테스트 | 각 시나리오 1-4 실행 | 4시간 | qa-engineer |
| 4 | 성능 테스트 | 시간 측정 (2-3분, 15-20분 등) | 2시간 | qa-engineer |
| 5 | 호환성 테스트 | v2.1.0 마이그레이션 | 2시간 | backend-expert |
| 6 | 사용자 수용 테스트 | 실제 사용자 3-5명 | 3시간 | qa-engineer |

### 테스트 기준

**Pass 기준**:
- AC-001 ~ AC-013 모두 만족
- 예상 시간 달성 (±10% 허용)
- 에러 메시지 0개
- 성공 확률 100%

**Fail 기준**:
- AC 미달성
- 예상 시간 초과 (+15% 초과)
- 예상하지 못한 에러 발생
- 동일 이슈 반복 (2회 이상)

---

## 결론

본 수용 기준 및 테스트 시나리오는 SPEC-REDESIGN-001 구현 완료 후 설정 시간 단축, 질문 감소, 문서 자동 생성, AI 에이전트 컨텍스트 통합이 모두 달성되었음을 검증합니다.

특히 시나리오 1(Quick Start)과 시나리오 2(Full Documentation)를 통해 사용자의 다양한 필요에 대응하는 유연한 설정 경험을 제공할 수 있음을 보여줍니다.

