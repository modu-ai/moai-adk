---
name: "moai-alfred-rules"
version: "4.0.0"
created: 2025-11-02
updated: 2025-11-12
status: stable
tier: Alfred
description: "Alfred SuperAgent의 필수 규칙을 정의합니다. November 2025 enterprise standard 기반. 3-Layer architecture, 4-Step workflow, Agent-first paradigm, Skill invocation rules, AskUserQuestion patterns, TRUST 5 quality gates, TAG chain integrity, commit message standards. 사용: 워크플로우 규칙 검증, 품질 게이트 확인, MoAI-ADK 표준 준수, 아키텍처 규칙 검증."
keywords: ['rules', 'agent-first', 'skill-invocation', 'ask-user-question', 'trust-5', 'tag-chain', 'workflow-compliance', 'quality-gates', '4-step-workflow', 'architecture-rules']
allowed-tools: "Read, Glob, Grep, Bash, AskUserQuestion, TodoWrite"
---

## Skill 개요

**moai-alfred-rules**는 Alfred SuperAgent의 의사결정과 실행을 제어하는 핵심 프레임워크입니다.

| 항목 | 값 |
|------|-----|
| 버전 | 4.0.0 (November 2025 enterprise) |
| 티어 | Alfred (상위 계층) |
| 자동 로드 | 규칙 검증, 품질 게이트, 아키텍처 규칙 필요 시 |
| 아키텍처 패러다임 | Agent-First (Command → Agent → Skill → Hook) |
| 워크플로우 모델 | 4-Step ADAP Workflow |

---

## 무엇을 하는가?

### 핵심 책임

1. **3-Layer Architecture 정의**: Commands → Agents → Skills 계층 분리
2. **4-Step Workflow 규칙**: ADAP (Analyze, Design, Assure, Produce) + Intent
3. **Agent-First Paradigm**: 모든 실행 작업을 agents에 위임
4. **Skill 호출 규칙**: 10+ mandatory patterns, invocation syntax
5. **AskUserQuestion 패턴**: 5가지 필수 사용 시나리오
6. **TRUST 5 Quality Gates**: T/R/U/S/T 각각의 검증 기준
7. **TAG Chain Integrity**: SPEC→TEST→CODE→DOC 추적
8. **Commit Message Standards**: TDD cycle message formats

---

## 언제 사용하는가?

### 필수 시나리오 (MUST use)

| 상황 | 사용 여부 |
|------|---------|
| ✅ Skill() 호출 규칙 검증 | **필수** |
| ✅ Command vs Agent 역할 분리 | **필수** |
| ✅ AskUserQuestion 사용 판단 | **필수** |
| ✅ TRUST 5 준수 확인 | **필수** |
| ✅ TAG 체인 무결성 검증 | **필수** |
| ✅ 커밋 메시지 형식 확인 | **필수** |
| ✅ 워크플로우 compliance 검증 | **필수** |
| ✅ Agent delegation 올바른지 확인 | **필수** |
| ✅ 품질 게이트(quality gate) 통과 | **필수** |
| ✅ 아키텍처 규칙 위반 감지 | **필수** |

---

## Rule 1: 3-Layer Architecture (November 2025 Standard)

### 계층 구조

```
┌─────────────────────────────────────┐
│ Commands (Orchestration Only)       │ ← User-facing entry points
│ /alfred:0-project, /alfred:1-plan   │   No direct execution
│ /alfred:2-run, /alfred:3-sync       │
└──────────┬──────────────────────────┘
           │ Task(subagent_type="...")
           ↓
┌─────────────────────────────────────┐
│ Agents (Domain Expertise)           │ ← Deep reasoning
│ spec-builder, tdd-implementer       │   Complex decisions
│ test-engineer, doc-syncer           │   Plan → Execute
│ git-manager, qa-validator           │
└──────────┬──────────────────────────┘
           │ Skill("skill-name")
           ↓
┌─────────────────────────────────────┐
│ Skills (Knowledge Capsules)         │ ← Reusable patterns
│ 55 specialized Skills               │   Playbooks
│ < 1000 lines each                   │   Best practices
└─────────────────────────────────────┘
```

### 규칙 1.1: Commands - Orchestration ONLY

**금지 사항 (❌)**:
```bash
# ❌ WRONG: 직접 작업 실행
echo "Building application..."
python setup.py build
git commit -m "Build"

# ❌ WRONG: Skill 직접 호출
Skill("moai-alfred-rules")  # Commands에서 금지!

# ❌ WRONG: 복잡한 로직 구현
if feature_type == "backend":
  # 복잡한 비즈니스 로직...
```

**필수 사항 (✅)**:
```bash
# ✅ CORRECT: Agent에 위임
Task(
  subagent_type="tdd-implementer",
  description="Build and test application",
  prompt="Implement feature with RED-GREEN-REFACTOR cycle"
)

# ✅ CORRECT: 의사결정만 하고 위임
if user_approval_needed:
  AskUserQuestion(...)  # 사용자 확인 후 위임
  Task(subagent_type="implementation-planner", ...)
```

### 규칙 1.2: Agents - Domain Expertise Ownership

**agent의 책임**:
- ✅ 복잡한 분석 & 추론 (deep reasoning)
- ✅ 계획 수립 (planning)
- ✅ 의사결정 (decision-making)
- ✅ Skill 호출 및 조율 (orchestration within domain)
- ✅ 작업 실행 (execution)

**예시: tdd-implementer Agent**:
```
Agent receives: "Implement user authentication"
  ↓
1. Analyze: SPEC 검토, 요구사항 분석
2. Design: 아키텍처 설계
3. RED: test-engineer Skill 호출 → 테스트 작성
4. GREEN: 코드 구현
5. REFACTOR: 최적화
6. Commit: git-manager 호출 → commit
  ↓
Returns: Fully tested, documented, committed code
```

### 규칙 1.3: Skills - Knowledge Capsules (Stateless)

**Skill 특성**:
- ✅ 상태가 없음 (stateless)
- ✅ 재사용 가능 (reusable)
- ✅ Agent에 의해 호출됨 (called by agents)
- ✅ 1000줄 이하 (< 1000 lines)
- ✅ 단일 주제 (single topic)

**금지 (❌)**:
```bash
# ❌ WRONG: Skill이 다른 Skill 호출
Skill("moai-foundation-git")  # ← Skills에서 금지!

# ❌ WRONG: Skill이 Task() 실행
Task(subagent_type="...")  # ← Skills에서 금지!

# ❌ WRONG: Skill이 상태 유지
state = {"counter": 0}  # ← Stateless 위반
```

---

## Rule 2: 4-Step Agent-Based Workflow (November 2025)

### Phase Overview

```
Phase 0: INTENT                      Phase 1: ANALYZE
┌──────────────────┐               ┌──────────────────┐
│ User Request     │ ─clarity?─→   │ WebSearch        │
│ Ambiguous?       │ NO             │ WebFetch         │
│ ✅ YES: clarify  │     YES        │ Research         │
│ ✅ NO: continue  │────────────→   │ Best practices   │
└──────────────────┘               └──────────────────┘
                                          ↓
Phase 3: ASSURE                     Phase 2: DESIGN
┌──────────────────┐               ┌──────────────────┐
│ Quality Gate     │←──────────────→│ Architecture     │
│ TRUST 5          │                │ Latest info      │
│ TAG integrity    │                │ Version specs    │
│ Compliance       │                │ Design patterns  │
└──────────────────┘               └──────────────────┘
        ↓
Phase 4: PRODUCE
┌──────────────────┐
│ Skill invocation │
│ File generation  │
│ Commit           │
└──────────────────┘
```

### Phase 0: Intent (사용자 의도 파악)

**규칙 0.1**: Intent가 모호하면 AskUserQuestion 사용

```
상황: "데이터 처리 모듈 만들어줘"

Step 1: Intent 평가
  ├─ Clarity: LOW (어떤 데이터? 어떤 처리?)
  └─ Action: AskUserQuestion 사용

Step 2: 명확화
AskUserQuestion({
  question: "어떤 데이터를 처리하나요?",
  options: [
    "CSV 파일",
    "데이터베이스",
    "API 응답"
  ]
})

Step 3: 명확한 요구사항 확보 후 다음 phase로
```

### Phase 1: Analyze (정보 수집 & 연구)

**도구**: WebSearch, WebFetch

```
Task 1: Research latest information
  ├─ Search: "[framework] [version] best practices 2025"
  ├─ Fetch: Official documentation URLs
  └─ Validate: Cross-check multiple sources

Task 2: Collect best practices
  ├─ Official docs
  ├─ Industry standards
  └─ Current patterns (2025)

Task 3: Identify version-specific guidance
  ├─ Current stable version
  ├─ Breaking changes
  └─ Deprecation warnings
```

### Phase 2: Design (아키텍처 설계)

**입력**: Phase 1의 연구 결과
**출력**: November 2025 최신 정보를 기반한 설계

```
Design Activities:
  ├─ 최신 정보 기반 이름 지정
  ├─ 현재 버전 명시
  ├─ 최신 패턴 포함
  ├─ 공식 문서 링크
  └─ 마지막 업데이트 날짜 포함
```

### Phase 3: Assure (품질 검증)

**TRUST 5 Quality Gates**:

| Gate | 검증 기준 |
|------|----------|
| **Test** | 85%+ coverage, 모든 경로 테스트 |
| **Readable** | Clean code, SOLID 원칙, 주석 포함 |
| **Unified** | 일관된 패턴, 중복 제거, 네이밍 표준 |
| **Secured** | OWASP Top 10 확인, 비밀정보 제거 |

### Phase 4: Produce (생성 & 배포)

**책임**: Skill 호출 (예: moai-skill-factory)

```
Actions:
  ├─ 템플릿 적용
  ├─ 파일 생성
  ├─ 메타데이터 추가
  ├─ 최신 정보 embedded
  ├─ 공식 문서 링크 포함
  └─ Version date 표기
```

---

## Rule 3: Agent-First Paradigm (Critical Enforcement)

### 금지된 작업 (❌ ABSOLUTELY FORBIDDEN)

Alfred (또는 Command)가 직접 실행하면 **절대** 안 됩니다:

1. **직접 bash 명령 실행** ❌
   ```bash
   # ❌ WRONG
   bash("git commit -m 'message'")
   os.system("python build.py")
   
   # ✅ CORRECT
   Task(subagent_type="git-manager", prompt="Commit changes")
   ```

2. **파일 읽기/쓰기** ❌
   ```bash
   # ❌ WRONG
   with open("file.py", "w") as f:
       f.write(code)
   
   # ✅ CORRECT
   Task(subagent_type="file-manager", prompt="Create file")
   ```

3. **Git 직접 조작** ❌
   ```bash
   # ❌ WRONG
   subprocess.run(["git", "push", "origin", "main"])
   
   # ✅ CORRECT
   Task(subagent_type="git-manager", prompt="Push changes")
   ```

4. **코드 분석 직접 실행** ❌
   ```bash
   # ❌ WRONG
   lines = len(open("file.py").readlines())
   
   # ✅ CORRECT
   Task(subagent_type="code-analyzer", prompt="Analyze code")
   ```

5. **테스트 직접 실행** ❌
   ```bash
   # ❌ WRONG
   subprocess.run(["pytest", "tests/"])
   
   # ✅ CORRECT
   Task(subagent_type="test-engineer", prompt="Run tests")
   ```

### 의무 위임 (✅ MANDATORY DELEGATION)

| 작업 | 위임 대상 | 패턴 |
|------|---------|------|
| 계획 수립 | plan-agent | `Task(subagent_type="plan-agent", ...)` |
| 코드 개발 | tdd-implementer | `Task(subagent_type="tdd-implementer", ...)` |
| 테스트 작성 | test-engineer | `Task(subagent_type="test-engineer", ...)` |
| 문서화 | doc-syncer | `Task(subagent_type="doc-syncer", ...)` |
| Git 작업 | git-manager | `Task(subagent_type="git-manager", ...)` |
| 품질 검증 | qa-validator | `Task(subagent_type="qa-validator", ...)` |
| 사용자 질문 | ask-user-questions | `AskUserQuestion(...)` |

---

## Rule 4: 10 Mandatory Skill Invocations

### 규칙 4.1: Skill Invocation Pattern

**Syntax**:
```python
Skill("skill-name")  # Explicit invocation only
```

### 10가지 필수 Skill

| # | Skill | 용도 | Invocation |
|---|-------|------|-----------|
| 1 | moai-foundation-trust | TRUST 5 검증 | `Skill("moai-foundation-trust")` |
| 2 | moai-foundation-tags | TAG 검증 & 추적 | `Skill("moai-foundation-tags")` |
| 3 | moai-foundation-specs | SPEC 작성 & 검증 | `Skill("moai-foundation-specs")` |
| 4 | moai-foundation-ears | EARS 요구사항 형식 | `Skill("moai-foundation-ears")` |
| 5 | moai-foundation-git | Git 워크플로우 | `Skill("moai-foundation-git")` |
| 6 | moai-foundation-langs | 언어 & 스택 감지 | `Skill("moai-foundation-langs")` |
| 7 | moai-essentials-debug | 디버깅 & 에러 분석 | `Skill("moai-essentials-debug")` |
| 8 | moai-essentials-refactor | 리팩토링 & 개선 | `Skill("moai-essentials-refactor")` |
| 9 | moai-essentials-perf | 성능 최적화 | `Skill("moai-essentials-perf")` |
| 10 | moai-essentials-review | 코드 리뷰 & 품질 | `Skill("moai-essentials-review")` |

### 규칙 4.2: Skill Invocation 예제

```python
# Context 1: TRUST 5 검증 필요
if validation_required:
    Skill("moai-foundation-trust")
    # → 반환: TRUST score, violations, recommendations

# Context 2: TAG 체인 검증
if tag_integrity_check:
    Skill("moai-foundation-tags")
    # → 반환: 고아 TAG, 깨진 체인, 제안

# Context 3: SPEC 작성
if spec_needed:
    Skill("moai-foundation-specs")
    # → 반환: SPEC template, validation results

# Context 4: Git 워크플로우
if git_decision_needed:
    Skill("moai-foundation-git")
    # → 반환: branch strategy, commit format, merge rules
```

---

## Rule 5: AskUserQuestion Patterns (5가지 필수 시나리오)

### 규칙 5.1: MANDATORY Scenarios

**Scenario 1: 기술 스택 모호**
```
상황: "Python 웹 프레임워크 추천해줄래?"

AskUserQuestion({
  question: "어떤 유형의 애플리케이션?",
  header: "Application Type",
  options: [
    { label: "REST API", description: "High performance APIs" },
    { label: "Web Application", description: "Traditional MVC" },
    { label: "Microservice", description: "Event-driven" }
  ]
})
```

**Scenario 2: 아키텍처 결정**
```
상황: "데이터베이스 모델을 어떻게 설계?"

AskUserQuestion({
  question: "어떤 데이터 특성?",
  header: "Data Model",
  options: [
    { label: "Relational", description: "Structured, ACID" },
    { label: "Document", description: "Flexible schema" },
    { label: "Graph", description: "Relationships" }
  ]
})
```

**Scenario 3: 의도 모호**
```
상황: "코드 개선해줄래?"

AskUserQuestion({
  question: "어떤 측면을 개선?",
  header: "Improvement Focus",
  options: [
    { label: "Performance", description: "Speed & efficiency" },
    { label: "Readability", description: "Code clarity" },
    { label: "Security", description: "Vulnerability fixes" }
  ],
  multiSelect: true  # 복수 선택 허용
})
```

**Scenario 4: 기존 컴포넌트 영향**
```
상황: "패키지 업그레이드하려는데 호환성?"

AskUserQuestion({
  question: "기존 코드 호환성 유지 필요?",
  header: "Compatibility",
  options: [
    { label: "Full compatibility", description: "Maintain all APIs" },
    { label: "Deprecation path", description: "Gradual migration" },
    { label: "Breaking OK", description: "Version bump allowed" }
  ]
})
```

**Scenario 5: 자원 제약**
```
상황: "시스템 리팩토링 계획 수립"

AskUserQuestion({
  question: "예상 개발 기간?",
  header: "Timeline",
  options: [
    { label: "1 week", description: "Focused scope" },
    { label: "2-4 weeks", description: "Medium refactor" },
    { label: "1+ months", description: "Comprehensive" }
  ]
})
```

### 규칙 5.2: 올바른 사용법

**❌ WRONG** (평문 질문):
```
사용자: "뭘 선호해?"
응답: "음... 생각해봤는데 아마도..."
```

**✅ CORRECT** (AskUserQuestion):
```
AskUserQuestion({
  question: "어떤 접근을 선호하시나요?",
  header: "Approach",
  multiSelect: false,
  options: [
    { label: "Option A", description: "Benefit A, Cost B" },
    { label: "Option B", description: "Benefit C, Cost D" }
  ]
})
```

---

## Rule 6: TRUST 5 Quality Gates (November 2025 Enterprise)

### 각 Gate의 검증 기준

#### T: Test First (85%+ Coverage)
```yaml
requirements:
  coverage: "≥ 85%"
  coverage_tools: ["pytest-cov", "coverage.py"]
  test_types:
    - unit_tests: "각 함수/메서드"
    - integration_tests: "모듈 간 상호작용"
    - edge_cases: "경계값, 에러 처리"
  
validation:
  - All code paths executed
  - Error conditions tested
  - Mock external dependencies
  - No skipped tests (×skip, ×xfail)
```

#### R: Readable (Clean Code)
```yaml
requirements:
  code_standards:
    - SOLID 원칙 준수
    - DRY (Don't Repeat Yourself)
    - KISS (Keep It Simple, Stupid)
  
  documentation:
    - Function docstrings
    - Complex logic comments
    - Type hints (Python 3.10+)
  
  naming:
    - Descriptive variable names
    - Consistent conventions
    - No single-letter vars (except i, j, k in loops)
```

#### U: Unified (Consistent Patterns)
```yaml
requirements:
  consistency:
    - Same patterns across codebase
    - No duplicate logic
    - Shared utilities for common tasks
  
  conventions:
    - Consistent indentation (4 spaces)
    - Consistent naming (snake_case, PascalCase)
    - Consistent import order
  
  validation:
    - Linting (flake8, black, isort)
    - Static analysis (pylint, mypy)
    - No code duplication (DRY violations)
```

#### S: Secured (OWASP Top 10)
```yaml
requirements:
  security_checks:
    - No hardcoded credentials
    - No SQL injection vectors
    - No XXE vulnerabilities
    - Input validation
    - Output encoding
  
  tools:
    - bandit (Python security linter)
    - safety (dependency vulnerabilities)
    - SAST scanning
  
  validation:
    - No secrets committed
    - Dependencies scanned
    - Known CVEs checked
```

```yaml
requirements:
  tag_chain:
    - Complete history traceability
  
  documentation:
  
  validation:
    - Bidirectional references
```

---

## Rule 7: TAG Chain Integrity Rules

### 규칙 7.1: TAG Naming Convention

```
Format: @<DOMAIN>-<###>

Examples:
```

### 규칙 7.2: TAG Lifecycle

```

  └─ Example: test_auth.py - test_auth_success()

  └─ Example: auth.py - authenticate_user()

  └─ Example: CHANGELOG.md

```

### 규칙 7.3: TAG Validation Rules

**❌ 위반**:
```python
# 1. Orphan TAG (TEST/CODE 없음)

# 2. 깨진 체인 (missing step)

# 3. 불일치 (다른 번호)
```

**✅ 정확**:
```python
# Complete chain
```

---

## Rule 8: Commit Message Standards (TDD Cycle)

### 규칙 8.1: Commit Format

```
Format:
<type>(<tag>): <subject>

<body>

<footer>
```

### 규칙 8.2: TDD Cycle Commits

**RED Commit** (테스트 작성):
```

- test_successful_login()
- test_invalid_credentials()
- test_expired_token()

Status: RED (Tests fail as expected)
```

**GREEN Commit** (구현):
```

- Implement authenticate_user() function
- Add token generation
- Add error handling

Status: GREEN (All tests pass)
```

**REFACTOR Commit** (최적화):
```

- Extract token validation to separate function
- Add caching for user lookups
- Improve error messages

Status: PASSING (Tests still pass, code improved)
```

### 규칙 8.3: Commit 타입 (Conventional Commits)

| 타입 | 설명 | 예시 |
|------|------|------|
| **chore** | 빌드/설정 | `chore: Update dependencies` |

---

## Rule 9: Workflow Compliance Validation

### 규칙 9.1: Compliance Checklist

**Before Commit**:
- [ ] 테스트 통과 (85%+ coverage)
- [ ] Linting 통과 (black, flake8)
- [ ] Security scan 통과 (bandit, safety)
- [ ] 비밀정보 제거 (no secrets)
- [ ] 커밋 메시지 형식 맞음

**Before Merge**:
- [ ] TAG 체인 완전
- [ ] TRUST 5 통과
- [ ] Code review 완료
- [ ] CI/CD 통과
- [ ] PR 설명 포함

### 규칙 9.2: Violation Response

**Violation Detected**:
```
1. 자동 탐지 (hook)
2. 사용자에게 알림
3. 수정 요청
4. 재검증
5. Pass/Fail 결정
```

---

## 3-Level Progressive Disclosure

### Level 1: 빠른 시작 (Beginner - 10분)

**당신이 알아야 할 것**:
1. Command = 조율만, Agent = 실행
2. Skill = 전문 도구
3. 규칙 위반 → 에러

### Level 2: 실무 패턴 (Intermediate - 30분)

**당신이 해야 할 것**:
1. AskUserQuestion 사용 시점 판단
2. TRUST 5 검증
3. TAG 체인 유지
4. 커밋 메시지 형식

### Level 3: 심화 (Advanced - 1시간)

**당신이 최적화할 것**:
1. Agent delegation 전략
2. Skill 조합 활용
3. Workflow 자동화
4. 규칙 예외 관리

---

## 공식 자료 & 참고 링크

### Architecture References
- Command-Agent-Skill 패턴: Internal moai-adk documentation
- ADAP Workflow: Internal workflow definition

### Quality Standards
- TRUST 5 Framework: Skill("moai-foundation-trust")
- TAG System: Skill("moai-foundation-tags")
- Commit Standards: Skill("moai-foundation-git")

### Enterprise Standards (November 2025)
- Agent-First Architecture: InfoQ (agentic-ai-architecture-framework)
- TDD Best Practices: Official pytest documentation
- Code Quality: OWASP Top 10, SOLID principles

---

**버전**: 4.0.0 (November 2025 Enterprise Standard)
**최종 업데이트**: 2025-11-12
**유지보수자**: GoosLab (Alfred SuperAgent Framework)
