# @SPEC:ALF-WORKFLOW-001 구현 계획

> **4단계 워크플로우 로직 구현 전략**
>
> 우선순위 기반 마일스톤 및 기술적 접근 방식

---

## 구현 개요

### 목표

Alfred SuperAgent의 워크플로우를 체계화하여:
1. **의도 파악 단계**: 사용자 요청의 명확성 확인 및 `AskUserQuestion` 자동 실행
2. **계획 수립 단계**: Plan Agent를 통한 작업 분석 및 분해
3. **작업 실행 단계**: TodoWrite 기반 투명한 진행 상황 추적
4. **보고/커밋 단계**: 선택적 보고서 생성 및 필수 커밋 생성

### 핵심 가치

- **투명성**: TodoWrite를 통한 실시간 진행 상황 가시화
- **효율성**: Plan Agent의 병렬/순차 작업 판단으로 최적 실행
- **명확성**: AskUserQuestion으로 가정 제거
- **추적성**: Git 커밋 필수화로 모든 변경 이력 보존

---

## 마일스톤

### 🔥 Primary Goal: 핵심 워크플로우 구현

#### Milestone 1: 기초 문서 업데이트
**목표**: CLAUDE.md와 CLAUDE-RULES.md에 4단계 워크플로우 규칙 추가

**작업 항목**:
1. **CLAUDE.md 업데이트**
   - Alfred 역할 섹션에 4단계 워크플로우 설명 추가
   - "Alfred's Core Directives"에 워크플로우 원칙 명시
   - 의사결정 원칙에 AskUserQuestion 사용 기준 추가
   - TodoWrite 사용 지침 추가
   - Plan Agent 통합 설명 추가

2. **CLAUDE-RULES.md 업데이트**
   - Skill 호출 규칙에 Plan Agent 추가
   - Interactive Question Rules 섹션 확장
   - TodoWrite 사용 규칙 추가
   - 보고서 생성 규칙 명확화 (명시적 요청 시에만)
   - Git 커밋 필수화 규칙 추가

**검증 기준**:
- ✅ 두 문서에 4단계 워크플로우 언급
- ✅ AskUserQuestion 사용 기준 5가지 명시
- ✅ TodoWrite 상태 전이 규칙 명시
- ✅ 보고서 자동 생성 금지 규칙 명시

#### Milestone 2: 명령 템플릿 업데이트
**목표**: `/alfred:1-plan`, `/alfred:2-run`, `/alfred:3-sync` 명령에 워크플로우 로직 적용

**작업 항목**:
1. **1-plan.md 업데이트**
   - Intent Understanding 단계 추가
   - AskUserQuestion 조건부 실행 로직
   - Plan Agent 호출 지침
   - TodoWrite 초기 작업 목록 생성

2. **2-run.md 업데이트**
   - TodoWrite 기반 작업 진행 추적
   - 순차/병렬 실행 로직
   - 차단 요인 감지 및 처리
   - 작업 완료 시 즉시 completed 상태 변경

3. **3-sync.md 업데이트**
   - 최종 TAG 검증
   - 선택적 보고서 생성 로직
   - git-manager 호출하여 커밋 생성
   - 작업 완료 확인

**검증 기준**:
- ✅ 각 명령이 4단계 워크플로우를 명시적으로 참조
- ✅ 1-plan에서 Plan Agent 호출 확인
- ✅ 2-run에서 TodoWrite 업데이트 확인
- ✅ 3-sync에서 커밋 생성 필수화 확인

#### Milestone 3: 에이전트 지침 업데이트
**목표**: 5개 핵심 에이전트가 새 워크플로우를 따르도록 지침 업데이트

**작업 항목**:
1. **implementation-planner 업데이트**
   - Plan Agent 통합 패턴 추가
   - 작업 분석 결과를 TodoWrite 형식으로 출력
   - 단일/병렬 작업 판단 기준 명시

2. **tdd-implementer 업데이트**
   - 작업 시작 시 TodoWrite → in_progress 업데이트
   - 작업 완료 시 TodoWrite → completed 업데이트
   - 차단 요인 발생 시 보고 프로토콜

3. **doc-syncer 업데이트**
   - 보고서 생성 규칙: 명시적 요청 시에만
   - 허용 위치: `.moai/docs/`, `.moai/reports/`, `.moai/analysis/`
   - 금지 위치: 프로젝트 루트 (`IMPLEMENTATION_GUIDE.md` 등)

4. **git-manager 업데이트**
   - 모든 작업 완료 후 커밋 생성 필수화
   - TDD 단계별 커밋 (RED/GREEN/REFACTOR)
   - Alfred co-authorship 자동 추가

5. **spec-builder 업데이트**
   - SPEC 작성 시작 시 TodoWrite 업데이트
   - SPEC 완료 시 completed 상태 변경
   - git-manager에 Git 작업 위임

**검증 기준**:
- ✅ 각 에이전트의 main.md에 TodoWrite 업데이트 로직 존재
- ✅ git-manager만 Git 작업 수행
- ✅ doc-syncer에 보고서 생성 규칙 명시
- ✅ 모든 에이전트가 차단 요인 보고 프로토콜 준수

---

### 🚀 Secondary Goal: 통합 및 검증

#### Milestone 4: 통합 테스트
**목표**: 실제 워크플로우 시나리오로 4단계 로직 검증

**작업 항목**:
1. **시나리오 1: 명확한 요청**
   - 입력: "JWT 인증 시스템 만들어줘. 30분 토큰 만료."
   - 예상 동작:
     - AskUserQuestion 건너뛰기 ✅
     - Plan Agent 호출 ✅
     - TodoWrite 6개 작업 생성 ✅
     - 순차 실행 ✅
     - 커밋 3개 생성 (RED/GREEN/REFACTOR) ✅

2. **시나리오 2: 모호한 요청**
   - 입력: "대시보드 추가해줘"
   - 예상 동작:
     - AskUserQuestion 실행 ✅
     - 4-5개 질문 제시 ✅
     - 사용자 응답 수집 ✅
     - Plan Agent 호출 ✅
     - TodoWrite 작업 생성 ✅

3. **시나리오 3: 차단 요인 발생**
   - 입력: "FastAPI 앱에 Redis 캐싱 추가"
   - 예상 동작:
     - Plan Agent 호출 ✅
     - TodoWrite 작업 생성 ✅
     - 구현 중 Redis 미설치 감지 ✅
     - 차단 작업 생성 및 해결 ✅
     - 원래 작업 재개 ✅

4. **시나리오 4: 병렬 작업**
   - 입력: "README.md와 CHANGELOG.md 업데이트"
   - 예상 동작:
     - Plan Agent가 병렬 가능 판단 ✅
     - 두 작업 동시 in_progress ✅
     - 두 doc-syncer 병렬 실행 ✅
     - 두 작업 동시 completed ✅

**검증 기준**:
- ✅ 4개 시나리오 모두 성공
- ✅ TodoWrite 상태 전이가 올바름
- ✅ 커밋이 모든 시나리오에서 생성됨
- ✅ 보고서는 명시적 요청 시에만 생성됨

#### Milestone 5: 문서화 및 마무리
**목표**: 사용자 가이드 및 개발자 문서 업데이트

**작업 항목**:
1. **CLAUDE-PRACTICES.md 업데이트**
   - 4단계 워크플로우 실전 예시 추가
   - AskUserQuestion 사용 패턴 추가
   - TodoWrite 패턴 추가
   - Plan Agent 통합 패턴 추가

2. **CLAUDE-AGENTS-GUIDE.md 업데이트**
   - Plan Agent 설명 추가 (built-in Claude Agent)
   - 에이전트 선택 결정 트리에 Plan Agent 추가
   - 에이전트 협업 패턴에 TodoWrite 통합 추가

3. **README.md 업데이트**
   - "Four-Step Workflow" 섹션 추가
   - 사용자 관점에서의 워크플로우 설명
   - 예시 스크린샷 (가능한 경우)

**검증 기준**:
- ✅ 3개 문서 모두 4단계 워크플로우 언급
- ✅ 실전 예시가 구체적이고 재현 가능
- ✅ 사용자가 이해하기 쉬운 언어로 작성

---

## 기술적 접근 방식

### 1. AskUserQuestion 통합

#### 구현 위치
- **Alfred (CLAUDE.md)**: 의도 명확성 평가 로직
- **1-plan.md**: AskUserQuestion 실행 지점

#### 조건부 실행 로직
```markdown
IF user_intent_clarity == LOW:
  THEN invoke AskUserQuestion
  COLLECT user_responses
  SUMMARIZE selections
  CONFIRM with user
ELSE:
  SKIP to Plan Agent
```

#### 명확성 평가 기준
- **HIGH**: 기술 스택, 요구사항, 범위 모두 명시
- **MEDIUM**: 일부 모호하지만 기본 가정 가능
- **LOW**: 여러 해석 가능, 비즈니스/UX 결정 필요

### 2. Plan Agent 통합

#### 호출 패턴
```markdown
Alfred → Plan Agent:
  - Task description (user's language)
  - Project context (language, architecture, existing code)
  - Expected output: task breakdown, dependencies, execution order

Plan Agent → Alfred:
  - Task list (structured format)
  - Execution strategy (sequential | parallel)
  - Estimated file changes
```

#### Plan Agent 출력 형식
```yaml
analysis:
  task_type: "single_feature" | "multiple_features" | "refactoring"
  complexity: "low" | "medium" | "high"
  execution_mode: "sequential" | "parallel"

tasks:
  - name: "Write SPEC document"
    agent: "spec-builder"
    dependencies: []
    estimated_files: [".moai/specs/SPEC-AUTH-001/*"]

  - name: "Write failing tests (RED)"
    agent: "tdd-implementer"
    dependencies: ["Write SPEC document"]
    estimated_files: ["tests/auth/test_jwt_service.py"]
```

### 3. TodoWrite 통합

#### 상태 전이 규칙
```
pending → in_progress: 작업 시작 시 (에이전트가 호출되기 직전)
in_progress → completed: 작업 완료 시 (에이전트가 성공 응답)
in_progress → pending: 차단 요인 발생 시 (새 차단 작업 생성)
```

#### 동시 in_progress 제약
- **규칙**: 정확히 하나의 작업만 in_progress
- **예외**: Plan Agent가 병렬 가능 판단한 경우에만 여러 작업 동시 in_progress 허용
- **검증**: TodoWrite 업데이트 전 상태 검사

### 4. 선택적 보고서 생성

#### 명시적 요청 감지 패턴
```python
REPORT_REQUEST_KEYWORDS = [
    "보고서 만들어",
    "분석 문서 작성",
    "구현 가이드 문서",
    "create report",
    "write analysis document",
    "implementation guide"
]

def should_generate_report(user_request: str) -> bool:
    return any(keyword in user_request.lower() for keyword in REPORT_REQUEST_KEYWORDS)
```

#### 허용/금지 위치
- ✅ **허용**: `.moai/docs/`, `.moai/reports/`, `.moai/analysis/`, `.moai/specs/SPEC-*/`
- ❌ **금지**: 프로젝트 루트 (`IMPLEMENTATION_GUIDE.md`, `EXPLORATION_REPORT.md`, `*_ANALYSIS.md`)

### 5. Git 커밋 필수화

#### 커밋 생성 시점
- **Trigger**: 모든 TodoWrite 작업이 completed 상태
- **Executor**: git-manager (다른 에이전트는 Git 작업 안 함)
- **Format**: TDD 단계별 (test/feat/refactor)

#### Alfred Co-authorship
```
Co-Authored-By: 🎩 Alfred@[MoAI](https://adk.mo.ai.kr)
```

---

## 아키텍처 설계

### 데이터 흐름

```
User Request
    ↓
Alfred (Intent Clarity Check)
    ↓
[LOW clarity]→ AskUserQuestion → User Responses → Alfred
    ↓
Plan Agent (Task Analysis)
    ↓
TodoWrite (Task List Creation)
    ↓
Task Executor (Sub-agents)
    ├─ spec-builder
    ├─ tdd-implementer
    ├─ doc-syncer
    └─ git-manager
    ↓
TodoWrite (Status Updates: pending → in_progress → completed)
    ↓
git-manager (Commit Creation)
    ↓
User (Completion Notification)
```

### 모듈 책임

| 모듈 | 책임 | 입력 | 출력 |
|------|------|------|------|
| **Alfred** | 의도 파악, Plan Agent 호출, TodoWrite 관리, 에이전트 조율 | 사용자 요청 | 작업 완료 알림 |
| **Plan Agent** | 작업 분석, 의존성 파악, 실행 전략 결정 | 작업 설명, 프로젝트 맥락 | 구조화된 작업 목록 |
| **AskUserQuestion** | 대화형 질문, 사용자 응답 수집 | 질문 목록 | 사용자 선택 |
| **TodoWrite** | 작업 목록 추적, 상태 관리 | 작업 항목, 상태 변경 | 현재 진행 상황 |
| **Sub-agents** | 실제 작업 수행 (SPEC, TDD, 문서, Git) | 작업 지시 | 완료/차단 보고 |

### 오류 처리 전략

| 오류 유형 | 감지 | 처리 |
|-----------|------|------|
| **의도 불명확** | Alfred (명확성 평가) | AskUserQuestion 실행 |
| **차단 요인** | Sub-agent (실행 중 오류) | 새 차단 작업 생성, 원래 작업 in_progress 유지 |
| **TodoWrite 상태 충돌** | Alfred (상태 업데이트 전 검증) | 오류 로그, 작업 중단, 사용자 알림 |
| **Plan Agent 실패** | Alfred (Plan 호출 후 검증) | 수동 작업 분해로 폴백, 사용자 알림 |
| **Git 커밋 실패** | git-manager | 재시도 1회, 실패 시 사용자에게 수동 커밋 요청 |

---

## 위험 및 대응 방안

### Risk 1: Plan Agent 의존성
**위험**: Plan Agent가 작업을 잘못 분석하거나 실패할 경우 전체 워크플로우 중단
**확률**: 낮음 (Claude built-in agent의 안정성)
**영향**: 높음 (워크플로우 진행 불가)
**대응**:
- Plan Agent 실패 시 수동 작업 분해로 폴백
- Alfred가 간단한 작업은 Plan Agent 없이 직접 분해
- 사용자에게 명시적 작업 목록 요청 옵션 제공

### Risk 2: TodoWrite 상태 불일치
**위험**: 여러 에이전트가 동시에 TodoWrite 업데이트 시 충돌
**확률**: 중간 (병렬 작업 시 발생 가능)
**영향**: 중간 (진행 상황 추적 오류)
**대응**:
- TodoWrite 업데이트 전 상태 검증 로직 추가
- 병렬 작업 시 각 에이전트에 별도 작업 할당 (충돌 방지)
- 충돌 감지 시 즉시 작업 중단 및 사용자 알림

### Risk 3: AskUserQuestion 남용
**위험**: 너무 많은 질문으로 사용자 피로 증가
**확률**: 중간 (명확성 평가 기준이 너무 엄격할 경우)
**영향**: 중간 (사용자 경험 저하)
**대응**:
- 질문 수를 3-5개로 제한
- 기본값 제공하여 사용자가 Enter만 눌러도 진행 가능
- "이전 선택 기억" 기능 (유사한 요청 시 재사용)

### Risk 4: 보고서 생성 규칙 혼란
**위험**: "명시적 요청"의 정의가 모호하여 일관성 없는 동작
**확률**: 중간 (자연어 해석의 어려움)
**영향**: 낮음 (보고서 과다/누락 생성)
**대응**:
- 명확한 키워드 목록 정의 ("보고서", "문서", "분석", "가이드")
- 애매한 경우 사용자에게 확인 질문
- 보고서 생성 여부를 AskUserQuestion에 포함 옵션 제공

---

## 성공 기준

### 기능적 기준
- ✅ 4단계 워크플로우가 모든 Alfred 명령에서 실행됨
- ✅ AskUserQuestion이 모호한 요청에서 자동 실행됨
- ✅ Plan Agent가 작업을 올바르게 분석하고 TodoWrite 형식으로 출력함
- ✅ TodoWrite가 실시간으로 작업 진행 상황을 반영함
- ✅ 보고서가 명시적 요청 시에만 생성됨
- ✅ Git 커밋이 모든 작업 완료 후 자동 생성됨

### 품질 기준
- ✅ 문서 일관성: 모든 문서가 4단계 워크플로우를 동일하게 설명
- ✅ TAG 체인 무결성: @SPEC:ALF-WORKFLOW-001이 모든 관련 파일에 존재
- ✅ 테스트 커버리지: 통합 시나리오 4개 모두 성공
- ✅ 사용자 경험: AskUserQuestion이 3-5개 질문으로 제한됨
- ✅ 성능: Plan Agent 호출이 5초 이내 완료

### 추적성 기준
- ✅ 10개 파일 모두 수정 완료
- ✅ Git 커밋이 TDD 단계별로 분리 (RED/GREEN/REFACTOR)
- ✅ PR #118이 Ready for Review 상태로 변경
- ✅ SPEC 문서가 v0.1.0으로 업데이트 (구현 완료 후)

---

## 다음 단계

1. **`/alfred:2-run SPEC-ALF-WORKFLOW-001` 실행**
   - implementation-planner가 이 plan.md를 분석
   - TDD 사이클로 구현 (RED → GREEN → REFACTOR)
   - 모든 수정이 feature/SPEC-ALF-WORKFLOW-001 브랜치에 커밋

2. **`/alfred:3-sync` 실행**
   - TAG 체인 검증
   - 문서 동기화
   - PR #118을 Ready for Review로 변경

3. **통합 테스트**
   - 4개 시나리오 실행
   - 문제 발견 시 수정 및 재테스트

4. **PR 리뷰 및 머지**
   - 코드 리뷰 요청
   - 승인 후 main 브랜치에 머지
   - v0.8.3 릴리스 준비

---

**마지막 업데이트**: 2025-10-29
**문서 버전**: v0.0.1
**작성자**: @Goos
