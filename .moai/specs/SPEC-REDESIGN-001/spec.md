---
spec_id: SPEC-REDESIGN-001
title: 프로젝트 설정 스키마 v3.0.0 리디자인 및 문서 통합
version: 1.0.0
status: ready_for_implementation
priority: high
created_at: 2025-11-19T00:00:00Z
updated_at: 2025-11-19T00:00:00Z
author: spec-builder
---

# SPEC-REDESIGN-001: 프로젝트 설정 스키마 v3.0.0 리디자인 및 문서 통합

## 개요

### 목적
/moai:0-project 명령어의 사용자 경험을 획기적으로 단순화하고, 프로젝트 문서 생성 기능을 통합하여 AI 에이전트가 프로젝트 맥락을 이해할 수 있도록 하는 새로운 설정 시스템 구축.

### 핵심 문제점
- 기존: 사용자 질문 27개 → 설정 시간 15-20분 (신규 사용자 진입 장벽 높음)
- 설정 과정에서 프로젝트 문서 미생성 → AI 에이전트가 프로젝트 이해 부족
- Tab 탭 구조 비효율 (5개 탭) → 네비게이션 복잡

### 성공 기준
- 설정 시간: 2-3분 (Quick Start), 5-30분 (Full Documentation)
- 질문 감소율: 63% (27개 → 10개 essential)
- 설정 커버리지: 100% (31개 설정 전부)
- 문서 생성 기능: 자동화 (product.md, structure.md, tech.md)
- AI 에이전트 맥락 통합: 완료

### 범위
1. Tab 구조 리디자인 (5개 → 3개)
2. 사용자 질문 최소화 및 smart defaults 도입
3. 프로젝트 문서 생성 워크플로우 통합
4. AI 에이전트 컨텍스트 자동 로딩
5. 하위호환성 유지 (v2.1.0 설정)

---

## 필수 요구사항 (MUST)

### 질문 및 설정 요구사항
1. **MUST** 사용자 질문을 27개에서 10개 essential 질문으로 축소
   - Tab 1 Quick Start: 10개 질문
   - Tab 2 Documentation Choice: 1개 선택 질문
   - Tab 3 Git Automation: 조건부 2개 (Personal) 또는 2개 (Team)

2. **MUST** 16개 설정값에 smart defaults 자동 적용
   - 사용자 개입 없이 자동 구성
   - 사용자 선택 기반 지능형 기본값
   - 예: 개발 모드 선택 시 관련 Git 설정 자동 적용

3. **MUST** 5개 필드 자동 감지 및 설정
   - project.language: 코드베이스에서 감지
   - project.locale: conversation_language에서 매핑
   - language.conversation_language_name: 언어 코드→이름 변환
   - project.template_version: 시스템 관리
   - moai.version: 시스템 관리

4. **MUST** 100% 설정 커버리지 유지 (31개 설정)
   - 사용자 입력 + smart defaults + auto-detected = 31개 설정 완전 커버

### 기능 요구사항
5. **MUST** Quick Start 모드 지원 (2-3분)
   - 최소한의 질문 (10개)
   - 문서 생성 건너뛰기 옵션
   - 개발 즉시 시작 가능

6. **MUST** Full Documentation 모드 지원
   - 문서 생성 선택: Full / Skip / Minimal
   - Brainstorming 깊이: Quick(5분) / Standard(15분) / Deep(30분)
   - AI 분석 기반 문서 자동 생성

7. **MUST** 프로젝트 문서 자동 생성
   - product.md: 프로젝트 비전 및 가치 제안
   - structure.md: 아키텍처 및 시스템 설계
   - tech.md: 기술 스택 및 도구 선택
   - 저장 위치: .moai/project/

8. **MUST** AI 에이전트 컨텍스트 자동 로딩
   - project-manager: product.md 로드 (프로젝트 비전 이해)
   - tdd-implementer: structure.md 로드 (아키텍처 참고)
   - backend-expert/frontend-expert: tech.md 로드 (기술 스택 정보)

9. **MUST** Personal/Team/Hybrid Git 모드 조건부 배치
   - Tab 3.1_personal: git_strategy.mode == 'personal' 또는 'hybrid'일 때만 표시
   - Tab 3.1_team: git_strategy.mode == 'team'일 때만 표시
   - 불필요한 질문 숨김 (UX 개선)

10. **MUST** Template 변수 시스템 지원
    - {{user.name}}, {{project.name}}, {{project.owner}} 등 동적 값
    - Runtime에 해석되어 config.json에 저장
    - 파라미터화된 설정값 활용

11. **MUST** AskUserQuestion API 제약 준수
    - 배치당 최대 4개 질문 (API 제한)
    - 이모지 불포함 (field names, headers)
    - 자동 "Other" 옵션 지원 (5번째 선택지)

12. **MUST** Backward Compatibility 유지
    - v2.1.0 설정값 자동 마이그레이션
    - 기존 config.json 형식 호환
    - 설정 스키마 버전 관리

---

## 비필수 요구사항 (SHOULD)

### 성능 및 사용성
1. **SHOULD** 설정 시간 획기적 단축
   - 기존: 15-20분 → 2-30분 (선택에 따라)
   - Quick Start: 2-3분 (즉시 개발 시작)
   - Full Deep: 30분 (포괄적 문서 생성)

2. **SHOULD** 사용자 경험 개선
   - 명확한 단계별 가이드
   - 진행도 표시 및 예상 시간 안내
   - 다음 단계 자동 제시

3. **SHOULD** 스마트 기본값 제시
   - 개발 모드 선택 시 관련 설정 자동 제안
   - Git 모드에 따른 조건부 기본값
   - 사용자 선택 프리셋 (Personal Dev / Team Collaboration)

4. **SHOULD** 다국어 지원
   - 50+ MoAI-ADK 언어 지원
   - 질문 및 옵션 다국어화
   - conversation_language에 따른 자동 번역

### 신뢰성
5. **SHOULD** 입력값 검증 및 에러 처리
   - 필수 필드 검증
   - 설정값 범위 확인
   - 상충되는 설정 감지 및 경고

6. **SHOULD** 설정 저장 원자성
   - 전체 설정 동시 저장 또는 전혀 저장 안 함
   - 부분 저장 상태 방지
   - config.json 무결성 보장

---

## 인터페이스 요구사항 (SHALL)

### UI/UX 설계
1. **SHALL** 3-Tab 인터페이스 제공
   - Tab 1: Quick Start (Essential Setup) - 10개 질문
   - Tab 2: Project Documentation (Documentation Choice) - 1-2개 질문
   - Tab 3: Git Automation (Conditional) - 0-4개 질문 (모드에 따라)

2. **SHALL** AskUserQuestion 배치 기반 상호작용
   - 한 번에 4개 이하 질문만 표시
   - 배치 완료 후 다음 배치로 진행
   - 진행도 추적 가능

3. **SHALL** 각 선택지별 실행 시간 표시
   - "Quick Start - Skip for Now (30초)"
   - "Full Documentation - Now (5-30분 brainstorming 깊이에 따라)"
   - "Minimal - Auto-generate (1분)"

4. **SHALL** 명확한 진행 지표 및 다음 단계
   - 현재 진행도: "Tab 1/3 - Batch 1/4"
   - 예상 남은 시간
   - 다음 단계 미리보기

### 입력 옵션
5. **SHALL** 조건부 렌더링 지원
   - show_if: conditional 필드 (예: project.documentation_mode == 'full_now')
   - Tab 3 선택적 표시 (git_strategy.mode 기반)
   - 동적 질문 show/hide

6. **SHALL** 다양한 입력 타입 지원
   - text_input: 사용자 정의 입력 (name, project_name 등)
   - select_single: 단일 선택 (language, mode 등)
   - number_input: 수치 입력 (test_coverage_target 등)
   - 각 타입별 검증 규칙

---

## 설계 제약사항 (MUST)

### API 및 기술 제약
1. **MUST** AskUserQuestion API 제약 준수
   - 배치당 최대 4개 질문 (하드 리미트)
   - 최소 2개, 최대 4개 옵션 (질문당)
   - 5번째 옵션은 자동 "Other" (auto_other_option: true)

2. **MUST** 이모지 금지
   - 질문 필드에 이모지 사용 금지
   - 헤더에 이모지 사용 금지
   - Header는 최대 12자 (칩/태그 디스플레이)

3. **MUST** Template 변수 시스템 사용
   - {{user.name}}, {{project.name}}, {{project.owner}} 등
   - {{field.path}} 문법 준수
   - Runtime 해석 및 동적 값 설정

4. **MUST** Conditional 렌더링 로직 지원
   - show_if: "field == 'value'" 형식
   - 중첩 조건 지원 (AND, OR 로직)
   - 실행 시 조건 평가

### 스키마 및 데이터
5. **MUST** JSON 스키마 정합성
   - tab_schema.json v3.0.0 준수
   - 모든 필수 필드 포함 (id, label, batch_count 등)
   - Schema 검증 도구로 정합성 확인

6. **MUST** Smart Defaults 정의
   - 모든 선택지에 smart_default 값 지정
   - 사용자 입력 없을 때 자동 적용
   - Fallback 기본값 제공

7. **MUST** 설정값 완결성
   - 31개 설정값 모두 정의
   - User input / Smart defaults / Auto-detected 분류
   - 중복 또는 누락 방지

### 통합 및 호환성
8. **MUST** Skill 통합
   - moai-project-batch-questions: 질문 표시
   - moai-project-documentation-generator: 문서 생성 (NEW)
   - moai-project-config-manager: 원자적 저장
   - moai-project-language-initializer: 다국어 처리

9. **MUST** MCP 통합
   - Context7 MCP: 최신 문서 및 라이브러리 정보
   - WebSearch MCP (Deep 모드): 경쟁사 분석, 트렌드 조사

10. **MUST** 에이전트 컨텍스트 로딩
    - product.md → project-manager.task_prompt 추가
    - structure.md → tdd-implementer.system_context 추가
    - tech.md → backend-expert/frontend-expert.context 추가

---

## 수용 기준 (Given-When-Then)

### 시나리오 1: Quick Start 모드 (2-3분 설정)
**Given**: 신규 사용자가 /moai:0-project 명령어 실행
**When**: Tab 1 완료 후 Tab 2에서 "Quick Start - Skip for Now" 선택 및 Tab 3 Git 설정 완료
**Then**:
- 설정 소요 시간: 2-3분 이내
- config.json 생성: 10개 사용자 입력 + 16개 smart defaults + 5개 auto-detected = 31개 완전 설정
- 개발 준비 완료: 즉시 /moai:1-plan 명령어로 SPEC 생성 가능
- Minimal templates 생성: product.md, structure.md, tech.md 빈 템플릿만 생성

### 시나리오 2: Full Documentation 모드 (Standard 깊이, 15분)
**Given**: 신규 사용자가 Full Documentation 선택, Brainstorming 깊이 "Standard"(15분) 선택
**When**: Brainstorming 질문 10-15개 응답 및 AI 분석 완료
**Then**:
- 문서 자동 생성 완료:
  - product.md: 프로젝트 비전, 타겟 사용자, 가치 제안 (AI 분석 포함)
  - structure.md: 전체 아키텍처, 주요 컴포넌트, 의존성 (다이어그램 포함)
  - tech.md: 기술 스택 선택, Trade-offs, 초기 설정 가이드
- AI 에이전트 컨텍스트 자동 로드: 3개 문서가 해당 에이전트에 로드됨
- 예상 소요 시간: 15분 (사용자 입력 + AI 분석)

### 시나리오 3: Personal 모드 조건부 배치
**Given**: Tab 1 완료, git_strategy.mode = "personal" 선택
**When**: Tab 3로 진행
**Then**:
- Tab 3.1_personal 배치만 표시 (Auto Checkpoint, Push to Remote)
- Tab 3.1_team 배치는 숨겨짐 (Auto PR, Draft PR)
- 불필요한 질문 0개 (UX 개선)
- Personal 모드 기본값 적용 (event-driven, false)

### 시나리오 4: Team 모드 조건부 배치
**Given**: Tab 1 완료, git_strategy.mode = "team" 선택
**When**: Tab 3로 진행
**Then**:
- Tab 3.1_team 배치만 표시 (Auto PR, Draft PR)
- Tab 3.1_personal 배치는 숨겨짐 (Auto Checkpoint, Push to Remote)
- Team 기본값 적용 (false, false)

### 시나리오 5: Hybrid 모드 조건부 배치
**Given**: Tab 1 완료, git_strategy.mode = "hybrid" 선택
**When**: Tab 3로 진행
**Then**:
- Tab 3.1_personal 배치 표시 (현재는 Personal 모드가 활성)
- 추가 질문: "팀 협업 시작 시기?" (Dynamic 토글)
- 전환 시 Team 배치로 자동 전환

### 시나리오 6: 설정값 원자성
**Given**: 모든 Tab 완료, config.json 저장 대기 중
**When**: 저장 프로세스 시작
**Then**:
- 전체 31개 설정값 동시 저장
- 부분 저장 상태 없음 (all-or-nothing)
- 저장 실패 시 기존 config.json 유지
- 저장 성공 시 ".moai/config/config.json" 최신화

### 시나리오 7: Backward Compatibility
**Given**: v2.1.0 config.json 기존 파일 존재
**When**: /moai:0-project 다시 실행
**Then**:
- 기존 설정값 자동 감지 및 현재값 표시
- 마이그레이션 옵션 제시 (Keep Current / Update)
- v3.0.0 schema로 업그레이드
- 호환되지 않는 필드는 smart default로 마이그레이션

### 시나리오 8: Template 변수 해석
**Given**: config.json 저장, {{user.name}} 변수 사용
**When**: project-manager 에이전트 로드
**Then**:
- {{user.name}} → "GOOS행" (실제 값으로 해석)
- {{project.owner}} → "GoosLab" (설정값으로 해석)
- 동적 값 완전히 평가됨
- Agent task_prompt에 구체적 값 전달

### 시나리오 9: 문서 생성 중단
**Given**: Full Documentation + Deep 모드 선택 후 10분 경과
**When**: 사용자가 "완료" 버튼 클릭 (최소 50% 질문 응답)
**Then**:
- 현재까지의 응답으로 즉시 문서 생성
- "부분 완성" 마크 표시
- 나머지는 템플릿으로 자동 채움
- 사용자가 나중에 완성 가능

---

## 설정 커버리지 매트릭스

### 전체 31개 설정값

| 번호 | 설정명 | 소스 | 방식 | 참고 |
|------|--------|------|------|------|
| 1 | user.name | Tab 1.1 Q1 | User Input | 필수, smart_default="User" |
| 2 | language.conversation_language | Tab 1.1 Q2 | User Select | 필수, 4개 언어 옵션 |
| 3 | language.agent_prompt_language | Tab 1.1 Q3 | User Select | 필수, English 권장 |
| 4 | project.name | Tab 1.2 Q1 | User Input | 필수, smart_default="my-project" |
| 5 | project.owner | Tab 1.2 Q2 | User Input | 필수, smart_default={{user.name}} |
| 6 | project.description | Tab 1.2 Q3 | User Input | 선택, smart_default="" |
| 7 | git_strategy.mode | Tab 1.3 Q1 | User Select | 필수, 3개 옵션 |
| 8 | git_strategy.{mode}.workflow | Tab 1.3 Q2 | User Select | 필수, conditional mapping |
| 9 | constitution.test_coverage_target | Tab 1.4 Q1 | User Number | 필수, smart_default=90 |
| 10 | constitution.enforce_tdd | Tab 1.4 Q2 | User Select | 필수, smart_default=true |
| 11 | project.documentation_mode | Tab 2.1 Q1 | User Select | 필수, 3개 옵션 |
| 12 | project.documentation_depth | Tab 2.2 Q1 | Conditional | 조건부 (full_now일 때만) |
| 13 | git_strategy.personal.auto_checkpoint | Tab 3.1_personal Q1 | Conditional | Personal/Hybrid 모드 |
| 14 | git_strategy.personal.push_to_remote | Tab 3.1_personal Q2 | Conditional | Personal/Hybrid 모드 |
| 15 | git_strategy.team.auto_pr | Tab 3.1_team Q1 | Conditional | Team 모드 |
| 16 | git_strategy.team.draft_pr | Tab 3.1_team Q2 | Conditional | Team 모드 |
| 17 | project.language | Auto-Detect | 코드베이스 분석 | system-managed |
| 18 | project.locale | Auto-Detect | conversation_language 매핑 | system-managed |
| 19 | language.conversation_language_name | Auto-Detect | 언어코드→이름 변환 | system-managed |
| 20 | project.template_version | Auto-Detect | 시스템 관리 | system-managed |
| 21 | moai.version | Auto-Detect | 시스템 관리 | system-managed |
| 22 | git_strategy.personal.workflow (smart default) | Smart Default | mode=='personal'일 때 | github-flow |
| 23 | git_strategy.team.workflow (smart default) | Smart Default | mode=='team'일 때 | git-flow |
| 24 | constitution.test_coverage_target (alt) | Smart Default | 90 미만일 때 | 90 추천 |
| 25-31 | 기타 smart defaults | Smart Default | 16개 설정값 자동 적용 | 상황별 |

### 설정 커버리지 검증
- 사용자 질문: 10개
- 조건부 질문: 2개 (선택 기반)
- Auto-Detect: 5개
- Smart Defaults: 16개
- **총합: 31개 (100% 커버리지)**

---

## 구현 범위

### 신규 생성 파일
1. `.claude/skills/moai-project-documentation-generator/SKILL.md` - 문서 생성 Skill
2. `.claude/skills/moai-project-documentation-generator/brainstorm_questions.json` - Brainstorm 질문 세트
3. `.moai/specs/SPEC-REDESIGN-001/` - 본 SPEC 및 구현 계획

### 수정 파일
1. `.claude/skills/moai-project-batch-questions/tab_schema.json` - v3.0.0 (COMPLETED)
2. `.claude/commands/moai/0-project.md` - /moai:0-project 명령어 업데이트
3. `.claude/agents/moai/project-manager.md` - 문서 로딩 통합
4. `.claude/agents/moai/tdd-implementer.md` - 구조 문서 컨텍스트 추가
5. `.claude/agents/moai/backend-expert.md` - tech 문서 컨텍스트 추가
6. `.claude/agents/moai/frontend-expert.md` - tech 문서 컨텍스트 추가
7. `CLAUDE.md` - 문서 사용법 추가 섹션

### 관련 SPEC
- **SPEC-UPDATE-CONFIG**: config.json 구조 정의 및 마이그레이션
- **SPEC-DOC-GEN**: 프로젝트 문서 생성 엔진 구현
- **SPEC-CONTEXT-LOAD**: AI 에이전트 컨텍스트 자동 로딩

---

## 히스토리

| 버전 | 날짜 | 변경 내용 |
|------|------|---------|
| 1.0.0 | 2025-11-19 | 초안 작성 및 SPEC 완성 |

---

## 태그 및 추적

**관련 Tag IDs**:
- REDESIGN-001-001: tab_schema.json v3.0.0
- REDESIGN-001-002: moai-project-documentation-generator Skill
- REDESIGN-001-003: AI 에이전트 컨텍스트 통합

**의존성**:
- ✓ tab_schema.json v3.0.0 (완료)
- ⏳ moai-project-documentation-generator Skill (in progress)
- ⏳ AI 에이전트 컨텍스트 로딩 (pending)

