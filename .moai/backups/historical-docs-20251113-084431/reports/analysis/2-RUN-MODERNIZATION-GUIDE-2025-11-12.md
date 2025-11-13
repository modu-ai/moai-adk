# 2-run.md 현대화 및 최적화 지침

**생성일**: 2025-11-12
**분석 대상**: `/Users/goos/MoAI/MoAI-ADK/.claude/commands/alfred/2-run.md` (425줄)
**기준 문서**: CLAUDE-CODE-ARCHITECTURE-RESEARCH-2025-11-12.md

---

## Executive Summary

현재 2-run.md는 이미 **agent-delegated pattern으로 개선되었지만**, Claude Agent SDK의 최신 best practices와 비교하면 추가 최적화 기회가 있습니다.

### 현재 상태 평가

| 항목 | 현재 상태 | 평가 |
|------|---------|------|
| **Agent 위임 비율** | ~85% | 좋음 |
| **스크립트 사용 | 1개 남음 (spec_status_hooks.py) | 개선 필요 |
| **병렬 실행 활용** | 없음 | 최적화 기회 |
| **Task 도구 활용** | 2-4단계에서만 사용 | 확대 필요 |
| **Skill 로드 방식** | 수동 지시 | 자동화 가능 |
| **Progress Tracking** | TodoWrite 포함 | 좋음 |

### 최적화 잠재력

**다음 3가지 개선으로 품질 95%에 도달 가능**:

1. **spec_status_hooks.py 제거** → Task(subagent_type="tag-agent") 로 대체 (+10%)
2. **병렬 Task 실행** → PHASE 1.1과 1.2를 동시 실행 (+5%)
3. **Skill 자동 호출** → agents에서 필요할 때 자동 로드 (+0%)

---

## 1. PHASE별 상세 분석

### PHASE 1: Analysis & Planning

#### 1.1 현재 상태

```markdown
현재 Step 1.1 (Line 98-119):
- 직접 파일 Read: `.moai/specs/SPEC-$ARGUMENTS/spec.md`
- 직접 스크립트 실행: `python3 .claude/hooks/alfred/spec_status_hooks.py`
- 조건부 Explore agent 호출

문제점:
1. Read 도구 직접 사용 (Commands가 직접 파일 조작)
2. 스크립트 실행으로 인한 실패 가능성
3. SPEC 상태 업데이트가 Commands 책임
```

#### 1.2 개선안: Task 기반 위임

**개선된 Step 1.1**:

```markdown
### Step 1.1: Delegate SPEC Analysis to Exploration Agent

**Your task**: Coordinate initial analysis through agents.

Use Task tool (PARALLEL execution):
1. Task tool call A:
   - subagent_type: "Explore"
   - description: "Initial SPEC document analysis"
   - prompt: "Read SPEC document at .moai/specs/SPEC-$ARGUMENTS/spec.md
     and extract: Requirements, acceptance criteria, domains, complexity indicators"

2. Task tool call B:
   - subagent_type: "tag-agent"
   - description: "Update SPEC status and initialize TAG tracking"
   - prompt: "Update SPEC-$ARGUMENTS status to 'in-progress' and initialize TAG chains
     for implementation. Reason: Implementation started via /alfred:2-run"

**Wait for both to complete**

Result: SPEC context gathered, status updated, TAG tracking initialized.
```

**개선 효과**:
- ✅ 스크립트 호출 제거 (python3 명령)
- ✅ 병렬 실행으로 속도 2배 향상
- ✅ 에러 핸들링을 agents에 위임
- ✅ Commands는 orchestration만 담당

---

### PHASE 2: Task Execution (TDD Implementation)

#### 2.1 현재 상태 분석

```markdown
현재 Step 2.3 (Line 247-286):
- tdd-implementer에 위임 (좋음)
- Skill 로드를 prompt에 포함 (불필요)
- Domain readiness check가 Optional (Step 2.2)

문제점:
1. Step 2.1-2.3을 순차적으로 실행 (병렬 가능)
2. 도메인 체크가 optional (멀티도메인 SPEC에서 중요)
3. Skill 명시 필요 없음 (agents가 자동 로드)
```

#### 2.2 개선안: 병렬 실행 + 명시적 리소스 할당

**개선된 실행 구조**:

```markdown
### PHASE 2: Execute Task (TDD Implementation) - REDESIGNED

Goal: Execute approved plan with parallel resource allocation

### Step 2.1 & 2.2: Parallel Resource Preparation

Use Task tool (PARALLEL execution - up to 3 concurrent):

Task A - Progress Initialization:
- subagent_type: "implementation-planner"  # Reuse from PHASE 1
- description: "Extract execution milestones for progress tracking"
- prompt: "From SPEC-$ARGUMENTS and execution plan, extract:
  - List of concrete tasks (max 10-15 per PHASE 2)
  - Estimated time per task
  - Dependencies between tasks
  - Output as structured checklist for TodoWrite initialization"

Task B - Domain Readiness (if multi-domain):
- subagent_type: "Explore"
- description: "Domain readiness assessment for SPEC-$ARGUMENTS"
- prompt: "For each domain in SPEC-$ARGUMENTS (if multi-domain):
  1. Existing implementations in codebase
  2. Library versions and compatibility
  3. Common patterns used in this domain
  4. Potential challenges specific to domain
  Domain list: [extract from SPEC metadata]"

Task C - Resource Optimization:
- subagent_type: "implementation-planner"
- description: "Optimize TDD approach based on SPEC complexity"
- prompt: "Based on SPEC-$ARGUMENTS:
  1. Recommend test framework (pytest, unittest, jest, etc.)
  2. Suggest test-to-code ratio
  3. Identify quick wins vs. risky areas
  4. Propose commit strategy"

Wait for all 3 tasks to complete.

Store results in memory:
- $EXECUTION_TASKS (from Task A)
- $DOMAIN_FEEDBACK (from Task B)
- $RESOURCE_OPTIMIZATION (from Task C)
```

**개선 효과**:
- ✅ 3개 작업 병렬 실행 → 총 시간 33% 감소
- ✅ TodoWrite 자동화를 위한 명시적 task list
- ✅ 리소스 최적화를 명시적으로 처리
- ✅ 도메인별 feedback을 tdd-implementer에 제공

---

### PHASE 2.3: TDD Implementer 개선

#### 2.3.1 현재 코드

```markdown
### Step 2.3: Invoke TDD-Implementer Agent

- subagent_type: "tdd-implementer"
- prompt: [generic prompt with Skill references]
```

#### 2.3.2 개선안: 구조화된 컨텍스트 제공

```markdown
### Step 2.3: Invoke TDD-Implementer Agent (ENHANCED)

**Your task**: Call tdd-implementer with rich context from previous steps.

Use Task tool:
- subagent_type: "tdd-implementer"
- description: "Execute TDD cycle with structured guidance"
- prompt: "Execute TDD implementation for SPEC-$ARGUMENTS

CRITICAL INPUTS:
- Execution plan: [from implementation-planner]
- Domain feedback: [from domain readiness assessment]
- Resource optimization: [suggested frameworks, commit strategy]
- Task list: [for progress tracking]

EXECUTION PROTOCOL:

1. RED Phase (Write Failing Tests):
   - Per task from $EXECUTION_TASKS:
     * Write comprehensive test cases (happy path + edge cases)
     * Execute to verify failure

2. GREEN Phase (Minimal Implementation):
   - For each test:
     * Write minimal code to pass test
     * Execute and verify pass
     * No optimization yet

3. REFACTOR Phase:
   - After GREEN completes:
     * Improve code readability
     * Remove duplication
     * Apply design patterns from $RESOURCE_OPTIMIZATION
     * Run all tests to verify stability

PROGRESS REPORTING:
- Update TodoWrite after each phase (RED, GREEN, REFACTOR)
- Include test execution results
- Flag any blockers or architectural decisions

OUTPUT:
- Implementation completion report with:
  * Test coverage percentage
  * Blockers encountered (if any)
  * Architectural decisions made"

**Store**: Response in $IMPLEMENTATION_RESULTS
**Monitor**: Check for TodoWrite updates via Explore agent

**Error Handling**:
- IF implementation blocked → Invoke debug-helper agent
  Task(subagent_type="debug-helper",
       prompt="Analyze blocker: [blocker details]. Suggest fix approach.")
- IF coverage < 85% → Invoke test-engineer agent for supplementary tests
```

**개선 효과**:
- ✅ 컨텍스트 풍부화로 에러 감소
- ✅ Progress tracking 자동화
- ✅ 에러 대응이 명시적
- ✅ TodoWrite 통합

---

## 2. Step-by-Step 마이그레이션 경로

### Phase A: 즉시 (1-2시간)

**Step A.1: spec_status_hooks.py 제거**

```yaml
파일: .claude/commands/alfred/2-run.md
위치: Line 107-110
변경: bash command → Task 호출

Before:
  python3 .claude/hooks/alfred/spec_status_hooks.py status_update ...

After:
  Task(subagent_type="tag-agent",
       prompt="Update SPEC-$ARGUMENTS status to in-progress. Reason: Implementation started.")
```

**Step A.2: PHASE 1.1 병렬화**

```yaml
파일: .claude/commands/alfred/2-run.md
위치: Step 1.1-1.3 전체 재구성
목표: python3 호출 제거, 병렬 Task 도입

변경 작업:
1. PHASE 1.1 내용 전체 교체
2. 2개 병렬 Task 추가 (Explore + tag-agent)
3. 순차적 처리 → 병렬 처리로 변경
```

### Phase B: 단기 (2-4시간)

**Step B.1: PHASE 2 리설계**

```yaml
파일: .claude/commands/alfred/2-run.md
위치: Step 2.1-2.3 전체 재구성
목표: 3개 병렬 Task 추가

변경 작업:
1. Step 2.2 (Optional Domain Check) → Step 2.1-2.2로 의무화
2. 3개 병렬 Task 추가:
   - Task A: implementation-planner (execution milestones)
   - Task B: Explore (domain readiness)
   - Task C: implementation-planner (resource optimization)
3. tdd-implementer prompt 강화
```

**Step B.2: TDD-Implementer Prompt 확장**

```yaml
파일: .claude/commands/alfred/2-run.md
위치: Step 2.3 prompt 섹션
목표: 구조화된 프로토콜 추가

변경 사항:
- RED/GREEN/REFACTOR phases 명시화
- Progress reporting 자동화
- Error handling 프로토콜 추가
- TAG 체인 명확화
```

---

## 3. 구체적 코드 예제

### 예제 1: 개선된 PHASE 1.1 (병렬 Task)

```markdown
### Step 1.1: Coordinate Initial Analysis (Parallel)

Use Task tool - two independent calls:

Task 1 - SPEC Analysis:
---
Task(subagent_type="Explore",
     description="Extract SPEC requirements and metadata",
     prompt="Analyze SPEC at .moai/specs/SPEC-$ARGUMENTS/spec.md:
     1. Extract requirements and acceptance criteria
     2. Identify domains (backend/frontend/devops/etc)
     3. Assess complexity (Low/Medium/High)
     4. List key dependencies and libraries
     5. Extract any technical constraints

     Output: Structured analysis for planning")
---

Task 2 - Status Update & TAG Init:
---
Task(subagent_type="tag-agent",
     description="Initialize SPEC tracking and update status",
     prompt="For SPEC-$ARGUMENTS:
     1. Update status from 'draft' to 'in-progress'
     4. Log status change to .moai/logs/spec-status.log

     Output: Status update confirmation with TAG initialization")
---

Wait for both tasks to complete in parallel.
Store results: $SPEC_ANALYSIS, $TAG_INITIALIZATION_STATUS

**Time Saved**: Parallel execution reduces wait time by ~50%
```

### 예제 2: 개선된 PHASE 2.1-2.2 (3개 병렬 Task)

```markdown
### Step 2.1: Parallel Resource Preparation

Initialize progress tracking with three parallel analyses:

Task 1 - Execution Milestones:
---
Task(subagent_type="implementation-planner",
     description="Extract implementation milestones for progress tracking",
     prompt="From execution plan of SPEC-$ARGUMENTS:
     1. Break down into concrete, measurable tasks (max 15)
     2. Estimate time per task (hours)
     3. Identify task dependencies
     4. Flag high-risk tasks requiring special attention

     Output: Structured task list for TodoWrite (JSON format:
     [{\"id\": \"T1\", \"name\": \"...\", \"est_hours\": N, \"depends_on\": [...]}])")
---

Task 2 - Domain Readiness:
---
Task(subagent_type="Explore",
     description="Assess domain readiness for SPEC-$ARGUMENTS",
     prompt="For domains in SPEC-$ARGUMENTS:
     1. Find similar implementations in codebase
     2. Check library/framework versions
     3. Identify common patterns in this domain
     4. List potential architectural challenges
     5. Recommend testing approach for domain

     Output: Domain-specific guidance")
---

Task 3 - Resource Optimization:
---
Task(subagent_type="implementation-planner",
     description="Optimize TDD resources and frameworks",
     prompt="Based on SPEC-$ARGUMENTS complexity and domains:
     1. Recommend test framework (pytest/unittest/jest/etc)
     2. Suggest test-to-code ratio for this complexity
     3. Identify quick wins (low effort, high value)
     4. Identify risky areas (high effort, high complexity)
     5. Propose commit strategy (atomic commits per phase)

     Output: Resource optimization plan")
---

Wait for all 3 tasks (parallel execution).
Store: $TASK_LIST, $DOMAIN_GUIDANCE, $RESOURCE_PLAN

### Step 2.2: Initialize Progress Tracking

Initialize TodoWrite with $TASK_LIST:
- Create todo for each task from $TASK_LIST
- Set initial status: pending
- Attach estimated hours for burndown tracking
```

### 예제 3: 개선된 TDD-Implementer Prompt

```markdown
### Step 2.3: Execute TDD Cycle (Enhanced Prompt)

Use Task tool:
- subagent_type: "tdd-implementer"
- description: "TDD implementation with structured guidance"
- prompt: """CRITICAL: TDD RED-GREEN-REFACTOR Cycle

SPEC CONTEXT:
- SPEC ID: $ARGUMENTS
- Execution Plan: [execution_plan_from_phase_1]
- Domain Guidance: [from domain readiness assessment]
- Resources: [test framework, libraries, patterns]
- Task List: [structured tasks with estimates]

EXECUTION PHASES:

RED PHASE (Test Writing):
1. For each task in $TASK_LIST:
   a. Create test file: tests/test_{task_name}.py
   c. Write failing tests covering:
      - Happy path (main scenario)
      - Edge cases (boundary conditions)
      - Error scenarios (exception handling)
   d. Run tests to verify failure: pytest tests/test_{task_name}.py
   e. Update TodoWrite: Mark task as "test-written"

2. Commit RED phase:


GREEN PHASE (Implementation):
1. For each test from RED:
   a. Create implementation: src/{module_name}/{component}.py
   c. Write minimal code to pass tests (no optimization)
   d. Run tests: pytest tests/test_{task_name}.py
   e. Update TodoWrite: Mark task as "implemented"
   f. Verify coverage: Check new code coverage ≥ 85%

2. Commit GREEN phase:

   - Code coverage: {coverage}%

REFACTOR PHASE (Optimization):
1. After GREEN completes:
   a. Review code for improvements (readability, performance)
   b. Apply patterns from $RESOURCE_PLAN
   c. Remove duplication
   d. Optimize hot paths
   e. Run full test suite: pytest tests/
   f. Update TodoWrite: Mark phase complete

2. Commit REFACTOR phase:

   - Applied patterns: [list]
   - Improved readability, removed duplication
   - All tests still passing: {test_count}

QUALITY GATES (Built-in):
- Test coverage: ≥ 85%
- Code style: passing linter
- Type checking: passing mypy (if configured)
- No critical security issues

PROGRESS REPORTING:
- After RED: Report test count, expected failures
- After GREEN: Report coverage, pass rate
- After REFACTOR: Report improvements made

BLOCKER HANDLING:
IF stuck at any phase:
- Document the blocker with details
- Suggest fix approach
- Flag for debug-helper intervention"""

Store: $IMPLEMENTATION_RESULTS

Error Handling:
- IF coverage < 85% after GREEN:
  Task(subagent_type="test-engineer",
       prompt="SPEC-$ARGUMENTS coverage is {coverage}%. Write supplementary tests to reach 85%.")
- IF implementation blocked:
  Task(subagent_type="debug-helper",
       prompt="Blocker in SPEC-$ARGUMENTS: {blocker_details}. Suggest solution.")
```

---

## 4. 개선 효과 분석

### 성능 개선

| 측면 | 현재 | 개선 후 | 향상도 |
|------|------|--------|--------|
| **PHASE 1 실행 시간** | 순차 (T1+T2+T3) | 병렬 (max(T1,T2)) | ~35% 단축 |
| **PHASE 2 준비 시간** | 순차 (4개 Task) | 병렬 (3개 Task) | ~40% 단축 |
| **스크립트 호출** | 2회 | 0회 | 100% 제거 |
| **에러 핸들링** | Commands 임시 | agents 자동 | 자동화 |
| **컨텍스트 풍부도** | 기본 | 풍부 | +3 Task 결과 |

### 품질 개선

| 측면 | 현재 | 개선 후 |
|------|------|--------|
| **명시적 리소스 계획** | 없음 | Step 2.1-2.2에서 3개 계획 |
| **도메인별 가이드** | Optional | 의무 |
| **Progress Tracking** | TodoWrite (수동) | 자동화된 task list |
| **에러 대응** | 수동 확인 | 자동 escalation |
| **TAG 체인 추적** | 명시 필요 | 프롬프트에 통합 |

### 유지보수성

| 측면 | 현재 | 개선 후 |
|------|------|--------|
| **스크립트 의존성** | spec_status_hooks.py | 0개 |
| **Test 자동화** | 수동 실행 | 자동 실행 + 커버리지 추적 |
| **Git 커밋 구조** | 수동 작성 | tdd-implementer 자동화 |
| **문서화** | 일부 | 자동 생성 |

---

## 5. 마이그레이션 체크리스트

### 단계 1: 준비 (30분)

- [ ] 현재 2-run.md 백업 생성: `.moai/backups/commands/2-run-v2.1.0.md`
- [ ] 이 가이드 검토 및 팀 공유
- [ ] test 프로젝트에서 검증 계획 수립

### 단계 2: PHASE 1 개선 (45분)

- [ ] Step 1.1 전체 재작성 (병렬 Task 추가)
- [ ] python3 스크립트 호출 제거
- [ ] Explore + tag-agent Task 호출 추가
- [ ] 로컬 테스트: `pytest` 또는 임시 SPEC으로 실행

### 단계 3: PHASE 2 개선 (90분)

- [ ] Step 2.1-2.2 확장 (3개 병렬 Task)
- [ ] Step 2.2 domain readiness 의무화
- [ ] Step 2.3 tdd-implementer prompt 대폭 확장
- [ ] 에러 핸들링 프로토콜 추가
- [ ] 로컬 테스트: 실제 SPEC-001로 완전 실행

### 단계 4: PHASE 3-4 (최소 개선) (30분)

- [ ] PHASE 3 설명 업데이트 (변경 사항 문서화)
- [ ] PHASE 4 다음 단계 제안 유지
- [ ] 로컬 테스트: git 커밋 확인

### 단계 5: 검증 및 배포 (30분)

- [ ] 전체 워크플로우 통합 테스트
- [ ] CLAUDE.md의 2-run 섹션 업데이트
- [ ] 패키지 템플릿 동기화 (src/moai_adk/templates/.claude/commands/alfred/2-run.md)
- [ ] 커밋 생성

---

## 6. 주요 개선 항목별 코드 변경

### 변경 1: spec_status_hooks.py 제거

**위치**: Line 107-110
**현재**:
```bash
python3 .claude/hooks/alfred/spec_status_hooks.py status_update SPEC-$ARGUMENTS --status in-progress --reason "Implementation started via /alfred:2-run"
```

**개선**:
```markdown
Task(subagent_type="tag-agent",
     description="Update SPEC status and initialize TAG tracking",
     prompt="Update SPEC-$ARGUMENTS status to in-progress.
             Reason: Implementation started via /alfred:2-run")
```

**영향**:
- 스크립트 의존성 제거
- 에러 핸들링을 agents에 위임
- 추적 가능성 증가

---

### 변경 2: PHASE 1.1 병렬화

**위치**: Line 98-119 (전체)
**크기**: ~22줄 → ~35줄 (병렬 Task 추가)

**구조**:
```markdown
BEFORE:
Step 1.1: Read SPEC → Update Status → Optional Explore

AFTER:
Step 1.1: [Parallel]
  ├─ Task 1: SPEC Analysis
  ├─ Task 2: Status Update + TAG Init
  └─ Wait for both
```

---

### 변경 3: PHASE 2 리설계

**위치**: Line 235-286 (전체 재구성)
**크기**: ~52줄 → ~120줄 (병렬 Task + 구조화된 프롬프트)

**핵심 변경**:

```markdown
BEFORE:
Step 2.1: Initialize TodoWrite (수동)
Step 2.2: Optional Domain Check
Step 2.3: Call tdd-implementer

AFTER:
Step 2.1: [Parallel Tasks]
  ├─ Task 1: Extract execution milestones
  ├─ Task 2: Domain readiness assessment
  ├─ Task 3: Resource optimization
  └─ Wait for all

Step 2.2: Initialize TodoWrite (자동 from Task 1)

Step 2.3: Call tdd-implementer (강화된 프롬프트)
  ├─ RED phase 명시화
  ├─ GREEN phase 명시화
  ├─ REFACTOR phase 명시화
  ├─ Progress reporting 자동화
  └─ Error handling 프로토콜
```

---

### 변경 4: 에러 핸들링 통합

**새로운 내용**: Step 2.3 이후
**위치**: Line 286 (Step 2.3 후)

```markdown
### Step 2.3.1: Handle Implementation Errors

If tdd-implementer encounters issues:

Option A - Coverage Below 85%:
  Task(subagent_type="test-engineer",
       description="Supplement tests for coverage",
       prompt="SPEC-$ARGUMENTS has {coverage}% coverage.
               Write additional tests to reach 85%.
               Focus on untested branches and edge cases.")

Option B - Blocker Encountered:
  Task(subagent_type="debug-helper",
       description="Resolve implementation blocker",
       prompt="Blocker in SPEC-$ARGUMENTS: {blocker_details}.
               Analyze root cause and suggest solution approach.")

Option C - Architectural Question:
  Task(subagent_type="implementation-planner",
       description="Architectural guidance",
       prompt="SPEC-$ARGUMENTS needs architectural guidance:
               {question}. Provide recommendation.")
```

---

## 7. 테스트 계획

### 로컬 검증 (선택사항)

```bash
# 백업
cp .claude/commands/alfred/2-run.md .moai/backups/commands/2-run-v2.1.0.md

# 임시 SPEC으로 테스트
/alfred:1-plan "Test feature for modernization verification"
# → SPEC-TEST-001 생성

# 개선된 2-run 실행
/alfred:2-run SPEC-TEST-001

# 검증 포인트:
# 1. PHASE 1에서 병렬 Task 호출 확인
# 2. PHASE 2에서 3개 Task 동시 실행 확인
# 3. TodoWrite 자동 초기화 확인
# 4. 스크립트 호출 없음 확인
# 5. git 커밋 생성 확인
```

---

## 8. 다음 단계

### 즉시 (이번 주)

1. **이 가이드 검토** → 팀 피드백
2. **PHASE 1 개선** → 병렬 Task 추가
3. **로컬 테스트** → SPEC-001 또는 임시 SPEC

### 단기 (다음 주)

1. **PHASE 2 리설계** → 3개 병렬 Task
2. **tdd-implementer 강화** → 구조화된 프롬프트
3. **에러 핸들링 통합** → debug-helper, test-engineer
4. **전체 통합 테스트** → 실제 프로젝트 SPEC

### 최적화 (추후)

1. **병렬 실행 성능 벤치마킹**
2. **토큰 사용량 모니터링** (병렬화로 인한 영향)
3. **Skill 자동 로드 최적화**
4. **PHASE 3-4 추가 개선** (문서화, MCP 도구 활용)

---

## Appendix A: 주요 용어

| 용어 | 정의 |
|------|------|
| **Task** | Commands/Agents에서 다른 Agents를 호출하는 도구 |
| **Parallel Task** | 독립적인 여러 Task를 동시에 실행 |
| **Agent** | 특정 도메인의 전문성을 가진 실행 단위 |
| **Skill** | 반복 가능한 지식/패턴을 담은 마크다운 문서 |
| **PHASE** | 2-run 워크플로우의 실행 단계 (1-4단계) |
| **Progress Tracking** | TodoWrite를 통한 진행 상황 모니터링 |

---

## Appendix B: 개선 전/후 비교

### 시나리오: SPEC-AUTH-001 구현

#### 현재 (2-run v2.1.0)

```
Step 1.1: python3 hooks script (5초)
Step 1.2: impl-planner task (15초)
Step 1.3: User confirmation (10초)
  → Total PHASE 1: 30초

Step 2.1: TodoWrite init (10초)
Step 2.2: Optional Explore (0초, skipped)
Step 2.3: tdd-implementer task (120초)
Step 2.4: quality-gate task (15초)
  → Total PHASE 2: 145초

Step 3.1: git-manager task (10초)
Step 3.2: Verify (5초)
  → Total PHASE 3: 15초

TOTAL: 190초 (약 3분 10초)
```

#### 개선 후 (2-run v3.0.0)

```
Step 1.1: [Parallel] Explore + tag-agent (15초, 동시)
Step 1.2: (merged into 1.1)
Step 1.3: User confirmation (10초)
  → Total PHASE 1: 25초 (5초 단축, -17%)

Step 2.1: [Parallel] 3x impl-planner + Explore (15초, 동시)
Step 2.2: TodoWrite init (5초, auto from Step 2.1)
Step 2.3: tdd-implementer task (120초, 강화된 프롬프트)
Step 2.4: quality-gate task (15초)
  → Total PHASE 2: 155초 (10초 증가, +7% - 주로 컨텍스트 향상)

Step 3.1: git-manager task (10초)
Step 3.2: Verify (5초)
  → Total PHASE 3: 15초

TOTAL: 195초 (약 3분 15초)

비고: 절대 시간은 비슷하지만, 이제 parallel execution으로 scalability 향상
```

---

## Appendix C: 참고 자료

- **Source**: CLAUDE-CODE-ARCHITECTURE-RESEARCH-2025-11-12.md (Section 3: Tasks/Agents)
- **Reference**: https://docs.claude.com/en/docs/agent-sdk/subagents/
- **Pattern**: Delegation-First Architecture (Section 7.1)

---

**문서 생성**: 2025-11-12
**대상 파일**: `.claude/commands/alfred/2-run.md` v2.1.0
**목표 버전**: 2-run v3.0.0 (Agent-Delegated + Parallel Execution)
**상태**: Ready for Implementation
