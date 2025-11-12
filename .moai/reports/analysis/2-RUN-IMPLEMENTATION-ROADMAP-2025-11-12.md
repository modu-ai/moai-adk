# 2-run.md 개선 실행 로드맵

**문서**: 구체적 실행 가이드
**기준**: 2-RUN-MODERNIZATION-GUIDE-2025-11-12.md
**대상 파일**: `/Users/goos/MoAI/MoAI-ADK/.claude/commands/alfred/2-run.md`

---

## 실행 계획 개요

### 3단계 접근법

| 단계 | 작업 | 소요시간 | 우선순위 |
|------|------|--------|--------|
| **Step 1** | PHASE 1 개선 (병렬 Task) | 45분 | 높음 ⭐⭐⭐ |
| **Step 2** | PHASE 2 리설계 (3개 병렬 Task) | 90분 | 높음 ⭐⭐⭐ |
| **Step 3** | PHASE 2.3 강화 (TDD 프롬프트) | 60분 | 높음 ⭐⭐⭐ |
| **Step 4** | 에러 핸들링 추가 (debug-helper) | 30분 | 중간 ⭐⭐ |
| **Step 5** | 문서화 및 템플릿 동기화 | 30분 | 중간 ⭐⭐ |

**총 소요시간**: ~3.5시간

---

## Step 1: PHASE 1 개선 - 병렬 Task 도입

### 1.1 현재 코드 분석

**파일**: `.claude/commands/alfred/2-run.md`
**위치**: Line 94-119 (PHASE 1 전체)

```markdown
## 🚀 PHASE 1: Analysis & Planning

**Goal**: Analyze SPEC requirements and create execution plan.

### Step 1.1: Load Skills & Prepare Context
1. TUI System Ready
2. Read SPEC document
3. Update SPEC status to in-progress:
   python3 .claude/hooks/alfred/spec_status_hooks.py ...
4. Optionally invoke Explore agent

### Step 1.2: Invoke Implementation-Planner Agent
   Use Task tool with subagent_type: "implementation-planner"

### Step 1.3: Request User Approval
   Present plan and ask for approval
```

### 1.2 개선 내용

**핵심 변경**:
1. python3 스크립트 호출 제거
2. Step 1.1과 1.2를 병렬 실행으로 통합
3. Explore + tag-agent를 동시 호출

### 1.3 구체적 코드 변경

**변경 대상 라인**: 94-119

**Old Code (제거)**:

```markdown
### Step 1.1: Load Skills & Prepare Context

1. **TUI System Ready**:
   - Interactive menus are available for all user interactions

2. **Read SPEC document**:
   - Read: `.moai/specs/SPEC-$ARGUMENTS/spec.md`
   - Determine if codebase exploration is needed (existing patterns, similar implementations)

3. **Update SPEC status to in-progress**:
   ```bash
   python3 .claude/hooks/alfred/spec_status_hooks.py status_update SPEC-$ARGUMENTS --status in-progress --reason "Implementation started via /alfred:2-run"
   ```

4. **Optionally invoke Explore agent for codebase analysis**:
   - IF SPEC requires understanding existing code patterns:
     - Use Task tool with `subagent_type: "Explore"`
     - Prompt: "Analyze codebase for SPEC-$ARGUMENTS: Similar implementations, test patterns, architecture, libraries/versions"
     - Thoroughness: "medium"
   - ELSE: Skip and proceed directly to Step 1.3

**Result**: SPEC context gathered. Ready for planning.
```

**New Code (대체)**:

```markdown
### Step 1.1: Parallel SPEC Analysis & Status Initialization

**Goal**: Prepare SPEC context and initialize tracking in parallel.

Use Task tool - two independent parallel calls:

**Task 1.1.A - SPEC Requirements Analysis**:
```
Task(subagent_type="Explore",
     description="Extract SPEC requirements and technical context",
     prompt="Analyze SPEC document at .moai/specs/SPEC-$ARGUMENTS/spec.md:

1. Requirements & Acceptance Criteria
   - What are the main requirements?
   - What are the acceptance criteria?
   - Any non-functional requirements (performance, security)?

2. Domain & Complexity
   - What domains are involved? (backend/frontend/devops/database/etc)
   - Assess overall complexity (Low/Medium/High)
   - Estimated effort in hours

3. Technical Context
   - Key dependencies and libraries needed
   - Similar implementations in codebase?
   - Architectural patterns to follow?

4. Constraints & Risks
   - Any time/resource constraints?
   - Potential technical risks?

Output: Structured analysis as JSON:
{
  \"requirements\": [...],
  \"acceptance_criteria\": [...],
  \"domains\": [...],
  \"complexity\": \"Low|Medium|High\",
  \"estimated_hours\": N,
  \"technical_context\": {...},
  \"risks\": [...]
}")
```

**Task 1.1.B - SPEC Status & TAG Initialization**:
```
Task(subagent_type="tag-agent",
     description="Initialize SPEC tracking and update status",
     prompt="For SPEC-$ARGUMENTS:

1. Status Update
   - Update SPEC status from 'draft' to 'in-progress'
   - Record timestamp and reason: 'Implementation started via /alfred:2-run'

2. TAG Initialization
   - Create @SPEC-$ARGUMENTS_IMPL tracking TAG
   - Initialize TAG chains for implementation:
     * @TEST-{N} → @SPEC-$ARGUMENTS (test chain)
     * @CODE-{N} → @TEST-{N} → @SPEC-$ARGUMENTS (code chain)
     * @REFACTOR-{N} → @CODE-{N} (quality improvements)
   - Set initial counter to 001

3. Logging
   - Log status change to .moai/logs/spec-status.log
   - Include: timestamp, SPEC-ID, old_status → new_status, reason

Output: Initialization confirmation with:
- Status update timestamp
- Initialized TAGs
- Log file location")
```

**Wait for both tasks to complete** (parallel execution).

**Store results**:
- $SPEC_ANALYSIS (from Task 1.1.A)
- $TAG_INIT_STATUS (from Task 1.1.B)

**Result**: SPEC context gathered, status updated, TAG tracking initialized.
```

### 1.4 검증 체크리스트

생성된 개선된 Step 1.1이 다음을 만족하는지 확인:

- [ ] `python3` 호출 제거됨 (스크립트 없음)
- [ ] 2개 병렬 Task 호출 명확함
- [ ] `Task()` 도구 문법 정확함
- [ ] `subagent_type` 명시 (Explore, tag-agent)
- [ ] prompt가 구조화된 분석 요청
- [ ] 결과 저장 명시 ($SPEC_ANALYSIS, $TAG_INIT_STATUS)
- [ ] 라인 수: ~25-30줄 (이전 ~22줄에서 약간 증가)

---

## Step 2: PHASE 2 리설계 - 3개 병렬 Task

### 2.1 현재 코드 분석

**파일**: `.claude/commands/alfred/2-run.md`
**위치**: Line 217-324 (PHASE 2 전체)

```markdown
## 🔧 PHASE 2: Execute Task (TDD Implementation)

Step 2.1: Initialize Progress Tracking
Step 2.2: Check Domain Readiness (Optional)
Step 2.3: Invoke TDD-Implementer Agent
Step 2.4: Invoke Quality-Gate Agent
```

### 2.2 개선 목표

**3가지 병렬화**:
1. Execution milestones 추출 (impl-planner)
2. Domain readiness 검증 (Explore)
3. Resource optimization 계획 (impl-planner)

**4가지 개선**:
1. Step 2.2를 의무화 (optional → mandatory)
2. 병렬 실행으로 시간 단축
3. TodoWrite 자동 초기화
4. tdd-implementer에 풍부한 컨텍스트 제공

### 2.3 구체적 코드 변경

**변경 대상 라인**: 217-286 (Step 2.1-2.3)

**Old Code (제거)**:

```markdown
### Step 2.1: Initialize Progress Tracking

Use TodoWrite to track all tasks:
1. Parse tasks from execution plan
2. Initialize TodoWrite

### Step 2.2: Check Domain Readiness (Optional)

For multi-domain SPECs:
1. Read SPEC metadata
2. For each domain, invoke Explore agent
3. Store feedback

### Step 2.3: Invoke TDD-Implementer Agent

Use Task tool:
- subagent_type: "tdd-implementer"
- prompt: [current prompt]
```

**New Code (대체)**:

```markdown
### Step 2.1: Parallel Resource Preparation

**Goal**: Prepare execution resources in parallel for faster planning.

Use Task tool - three independent parallel calls:

**Task 2.1.A - Execution Milestones Extraction**:
```
Task(subagent_type="implementation-planner",
     description="Extract and structure execution milestones for progress tracking",
     prompt="From execution plan of SPEC-$ARGUMENTS:

1. Break down into concrete, measurable tasks
   - Each task should be completable in 1-2 hours
   - Maximum 15 tasks (if more, group related tasks)
   - Clear, specific task names

2. Task structure
   - Task ID: T001, T002, ... (sequential)
   - Name: Clear, actionable description
   - Estimated hours: 0.5, 1, 1.5, 2 hours
   - Dependencies: Which other tasks must complete first?
   - Type: test, implementation, refactor, documentation

3. Risk assessment
   - Which tasks are high-risk? (1-3 tasks)
   - Why are they risky?
   - Mitigation strategy for each

Output: Structured JSON array:
[
  {
    \"id\": \"T001\",
    \"name\": \"Write authentication unit tests\",
    \"type\": \"test\",
    \"est_hours\": 1.5,
    \"depends_on\": [],
    \"risk_level\": \"low\"
  },
  ...
]

This list will be used to initialize TodoWrite for progress tracking.")
```

**Task 2.1.B - Domain Readiness Assessment**:
```
Task(subagent_type="Explore",
     description="Assess domain-specific readiness for SPEC-$ARGUMENTS",
     prompt="For SPEC-$ARGUMENTS, analyze domain readiness:

Domains involved: [extract from SPEC metadata, e.g., backend, frontend, devops]

For each domain:
1. Existing implementations
   - Similar features in current codebase?
   - Which files/modules implement similar functionality?
   - Copy-paste opportunities?

2. Library & Framework Analysis
   - Current libraries used in this domain
   - Their versions
   - Compatibility with Python/Node.js version?
   - Any deprecated dependencies to replace?

3. Testing Patterns
   - Common test patterns for this domain
   - Mocking/fixture strategies used
   - Test structure conventions (pytest vs unittest)

4. Architecture Patterns
   - Naming conventions for this domain
   - Folder structure
   - Import patterns
   - Error handling conventions

5. Potential Challenges
   - Known complexity areas in this domain?
   - Integration challenges with other domains?
   - Performance considerations?

6. Recommendations
   - Suggested testing approach
   - Recommended libraries/frameworks
   - Code patterns to follow

Output: Domain-specific guidance as structured analysis")
```

**Task 2.1.C - Resource Optimization Planning**:
```
Task(subagent_type="implementation-planner",
     description="Optimize TDD approach based on SPEC complexity and resources",
     prompt="Based on SPEC-$ARGUMENTS complexity and resources:

1. Test Framework Selection
   - Recommended test framework (pytest, unittest, jest, vitest, etc)
   - Why this choice? (speed, coverage reporting, CI integration)
   - Configuration recommendations

2. Test-to-Code Ratio
   - Recommended test-to-code ratio for this complexity level
   - Examples: 1:1 (one test file per code file), 1:2, etc.
   - Coverage target (default: ≥85%)

3. Quick Wins vs Risky Areas
   - High-value, low-effort tasks (quick wins) - prioritize these
   - High-risk, high-complexity tasks - plan mitigation
   - Normal tasks - standard approach

4. Commit Strategy
   - Atomic commits per RED/GREEN/REFACTOR phase
   - Suggested commit message patterns
   - Branch strategy (feature branch, develop merge)

5. Quality Gates
   - Minimum coverage: ≥85%
   - Code style checks: [eslint, ruff, black, etc]
   - Type checking: [mypy, tsc, etc]
   - Security scanning: [bandit, safety, etc]

Output: Resource optimization plan as structured JSON:
{
  \"test_framework\": \"pytest\",
  \"framework_rationale\": \"...\",
  \"test_to_code_ratio\": \"1:1\",
  \"coverage_target\": 0.85,
  \"quick_wins\": [...],
  \"risky_areas\": [...],
  \"quality_gates\": {
    \"min_coverage\": 0.85,
    \"style_checker\": \"ruff\",
    \"type_checker\": \"mypy\"
  }
}")
```

**Wait for all 3 tasks** (parallel execution).

**Store results**:
- $TASK_LIST (from Task 2.1.A)
- $DOMAIN_GUIDANCE (from Task 2.1.B)
- $RESOURCE_PLAN (from Task 2.1.C)

**Result**: Execution resources prepared, domain guidance gathered, optimization plan ready.

---

### Step 2.2: Initialize Progress Tracking from Structured Tasks

**Goal**: Automatically initialize TodoWrite from $TASK_LIST.

Use result from Task 2.1.A:

1. **Extract task list** from $TASK_LIST (JSON array)
2. **Initialize TodoWrite** with each task:
   ```
   TodoWrite(action="initialize", items=[
     { task: "T001: Write authentication unit tests", status: "pending", hours: 1.5 },
     { task: "T002: Implement login endpoint", status: "pending", hours: 1 },
     ...
   ])
   ```
3. **Set burndown tracking**:
   - Total hours: Sum of est_hours from all tasks
   - Track progress: Update TodoWrite as each task completes

**Result**: Progress tracking initialized, burndown calculation ready.

---

### Step 2.3: Invoke TDD-Implementer Agent (Enhanced)

**Goal**: Execute TDD cycle with structured guidance from PHASE 2.1.

Use Task tool:
- subagent_type: "tdd-implementer"
- description: "Execute TDD implementation cycle with structured guidance"
- prompt: (see detailed prompt below)
```

### 2.4 TDD-Implementer Enhanced Prompt

**매우 중요**: 이것이 가장 큰 개선 지점

**현재 prompt 크기**: ~30줄
**개선된 prompt 크기**: ~100줄 (3배 이상)

```markdown
### Step 2.3: Invoke TDD-Implementer Agent (ENHANCED)

Use Task tool:
- subagent_type: "tdd-implementer"
- description: "Execute TDD RED-GREEN-REFACTOR cycle with structured guidance"
- prompt: """
You are the TDD implementer agent. Execute strict RED-GREEN-REFACTOR cycle for SPEC-$ARGUMENTS.

CRITICAL CONFIGURATION:
Language settings from .moai/config.json:
- agent_prompt_language: English (instructions language)
- conversation_language: User's preferred language
- Code: Always English
- Comments: Per project language rules

EXECUTION CONTEXT:
- SPEC ID: $ARGUMENTS
- Execution Plan: [from implementation-planner analysis in PHASE 1]
- Execution Tasks: $TASK_LIST (use this as your task decomposition)
- Domain Guidance: $DOMAIN_GUIDANCE (follow domain patterns)
- Resource Plan: $RESOURCE_PLAN (use test framework, commit strategy)

PHASE 1: RED (Write Failing Tests)
========================================

For each task in $TASK_LIST where type="test":

1. Create Test File
   - Location: tests/test_{task_name}.py
   - Add @TEST-{COUNTER} TAG in header comment
   - Link to @SPEC-$ARGUMENTS in docstring

2. Write Test Cases
   - Happy path (main scenario from acceptance criteria)
   - Edge cases (boundary conditions, special inputs)
   - Error scenarios (invalid inputs, exceptions)
   - Integration points (if applicable)

3. Run Tests
   - Execute: pytest tests/test_{task_name}.py -v
   - Verify all tests fail (they should, code doesn't exist yet)
   - Log failure reasons to confirm tests are correct

4. Update Progress
   - Update TodoWrite: T001 status → "test-written"
   - Document test count for this task

5. Commit RED Phase
   - Message: "test(@SPEC-$ARGUMENTS): Add failing tests for {task_name}"
   - Include: @TEST-{COUNTER}, @SPEC-$ARGUMENTS tags
   - Example:
     ```
     test(@SPEC-AUTH-001): Add failing tests for login endpoint

     Tests:
     - test_successful_login_with_valid_credentials
     - test_failed_login_with_invalid_password
     - test_failed_login_with_nonexistent_user
     - test_rate_limiting_after_5_attempts

     Related: @TEST-001 → @SPEC-AUTH-001
     ```

PHASE 2: GREEN (Minimal Implementation)
========================================

For each task in $TASK_LIST where type="implementation":

1. Create Implementation Files
   - Location: src/{module_name}/{component}.py
   - Add @CODE-{COUNTER} TAG in header comment
   - Link chain: @CODE-{COUNTER} → @TEST-{COUNTER} → @SPEC-$ARGUMENTS

2. Write Minimal Code
   - Implement ONLY what's needed to pass tests
   - No optimization, no extra features
   - Focus on correctness, not elegance
   - Follow coding standards from $DOMAIN_GUIDANCE

3. Run Tests
   - Execute: pytest tests/test_{task_name}.py -v
   - Verify all tests pass
   - Check coverage: pytest --cov=src/{module_name} --cov-report=term-missing
   - Record coverage percentage

4. Update Progress
   - Update TodoWrite: T{N} status → "implemented"
   - Record test pass rate, coverage %

5. Commit GREEN Phase
   - Message: "feat(@SPEC-$ARGUMENTS): Implement {component}"
   - Include: coverage %, test count
   - Example:
     ```
     feat(@SPEC-AUTH-001): Implement authentication service

     - All 8 tests passing
     - Coverage: 92%
     - Implements: login, verify_password, generate_token
     - Related: @CODE-001 → @TEST-001 → @SPEC-AUTH-001
     ```

PHASE 3: REFACTOR (Code Quality)
========================================

After all GREEN commits:

1. Review Code
   - Read all implementation files
   - Identify: duplication, unclear names, complex logic
   - Check against patterns from $DOMAIN_GUIDANCE

2. Improve Code Quality
   - Remove duplication (DRY principle)
   - Improve variable/function names (clarity)
   - Simplify complex logic (readability)
   - Apply design patterns from domain guidance
   - Optimize hot paths (performance)

3. Run Full Test Suite
   - Execute: pytest tests/
   - Ensure all tests still pass
   - Verify coverage hasn't decreased

4. Update Progress
   - Update TodoWrite: All tasks → "completed"
   - Document improvements made

5. Commit REFACTOR Phase
   - Message: "refactor(@SPEC-$ARGUMENTS): Improve code quality"
   - Include: improvements summary
   - Example:
     ```
     refactor(@SPEC-AUTH-001): Improve code quality and readability

     - Extracted: password_validation utility function
     - Simplified: token_generation logic (30 LOC → 15 LOC)
     - Applied: Factory pattern for service creation
     - All 8 tests still passing, coverage: 93%
     - Related: @CODE-001, @SPEC-AUTH-001
     ```

QUALITY ASSURANCE (Automatic)
========================================

During implementation:

1. Test Coverage
   - Target: ≥85% (from $RESOURCE_PLAN)
   - After GREEN: measure coverage
   - If below target: Add supplementary tests in REFACTOR

2. Code Style
   - Use linter from $RESOURCE_PLAN (ruff, eslint, etc)
   - Apply formatter (black, prettier, etc)
   - Fix style issues before commits

3. Type Checking (if applicable)
   - Run mypy, tsc, or equivalent
   - Fix type errors before commits

4. Security
   - No hardcoded credentials/secrets
   - Validate user input
   - Sanitize database queries
   - Check for OWASP Top 10 issues

PROGRESS REPORTING
========================================

After each phase (RED, GREEN, REFACTOR):

Report:
- Phase name (RED/GREEN/REFACTOR)
- Tasks completed in this phase
- Key metrics (test count, coverage %, LOC)
- Any blockers or issues encountered
- Time estimate vs actual (if tracked)

Example:
```
=== RED Phase Report ===
Tasks: T001, T002, T003 (all test-written)
Tests written: 24 total, 8 per task
Status: All tests failing as expected ✓

=== GREEN Phase Report ===
Tasks: T001, T002, T003 (all implemented)
Tests passing: 24/24 (100%)
Coverage: 89% (src/auth/)
Status: Ready for REFACTOR ✓

=== REFACTOR Phase Report ===
Tasks: T001, T002, T003 (all completed)
Improvements: 3 utilities extracted, 1 pattern applied
Coverage: 91% (final)
Status: Ready for quality-gate verification ✓
```

ERROR HANDLING
========================================

If you encounter errors:

1. Test Failures (during GREEN)
   - Don't ignore failing tests
   - Debug and fix implementation, not tests
   - Re-run until all pass
   - Document fix approach

2. Coverage Below 85%
   - Identify uncovered code paths
   - Write tests for missing coverage
   - Re-measure and verify ≥85%

3. Complex Logic
   - If a task is too complex to implement cleanly
   - Break into smaller subtasks
   - Ask: Should this be split into 2-3 tasks?
   - Raise as blocker for manual review

4. Architectural Decisions
   - If implementation reveals design issue
   - Document the issue clearly
   - Suggest 2-3 solution approaches
   - Raise for manual review by implementation-planner

Raise blockers as:
```
BLOCKER: {Blocker Type}
Issue: {Detailed description}
Impact: {What's blocked?}
Suggested Solutions: {2-3 options}
Escalation: Requires manual review
```

SKILLS REFERENCE
========================================

Use these Skills as needed:
- Skill("moai-alfred-language-detection") - For language-specific patterns
- Skill("moai-essentials-debug") - When debugging test/code failures
- Skill("moai-foundation-tags") - For TAG chain documentation
- Skill("moai-alfred-trust-validation") - To verify TRUST 5 principles

TAG SYSTEM REFERENCE
========================================

Use these TAG patterns throughout:

@SPEC-$ARGUMENTS
- Root specification, appears in all TAGs

@TEST-{COUNTER}
- Format: @TEST-001, @TEST-002, ...
- Location: test file headers and docstrings
- Links: @SPEC-$ARGUMENTS

@CODE-{COUNTER}
- Format: @CODE-001, @CODE-002, ...
- Location: implementation file headers
- Links: @TEST-{COUNTER} → @SPEC-$ARGUMENTS

@REFACTOR-{COUNTER}
- Format: @REFACTOR-001, @REFACTOR-002, ...
- Optional: Used to tag refactoring improvements
- Links: @CODE-{COUNTER}

COUNTER INCREMENT RULE:
- Start at 001 for RED phase
- Increment for each new test file
- Carry forward: @TEST-001 → @CODE-001 (same number)
- Refactor uses same numbers as CODE

FINAL OUTPUT
========================================

After completing all three phases (RED → GREEN → REFACTOR):

Provide summary:
- Total TAGs created (@TEST, @CODE, @REFACTOR counts)
- Final test coverage %
- Final code statistics (LOC, files, modules)
- Commits created during cycle
- Status: READY FOR QUALITY-GATE or BLOCKERS FOUND
"""

Store: $IMPLEMENTATION_RESULTS
```

### 2.5 검증 체크리스트

개선된 PHASE 2가 다음을 만족하는지 확인:

- [ ] Step 2.1에서 3개 Task 병렬 호출 명확함
- [ ] Task 2.1.A: execution milestones (impl-planner)
- [ ] Task 2.1.B: domain readiness (Explore)
- [ ] Task 2.1.C: resource optimization (impl-planner)
- [ ] Step 2.2에서 TodoWrite 자동 초기화
- [ ] Step 2.3 prompt가 100줄 이상의 상세 가이드
- [ ] RED/GREEN/REFACTOR phases 명확하게 구분
- [ ] TAG 체인 (@TEST → @CODE → @REFACTOR) 명시
- [ ] Progress reporting 섹션 포함
- [ ] Error handling 섹션 포함
- [ ] Skills 참고 섹션 포함

---

## Step 3: PHASE 2.4 후 에러 핸들링 추가

### 3.1 현재 코드 분석

**파일**: `.claude/commands/alfred/2-run.md`
**위치**: Line 319-323 (Quality-Gate 처리)

```markdown
### Step 2.4: Invoke Quality-Gate Agent

Handle result:
- IF PASS → Proceed to PHASE 3
- IF WARNING → Ask user: "Accept warnings?" or "Fix first?"
- IF CRITICAL → Block progress, report details, wait for fixes
```

### 3.2 개선 목표

**3가지 에러 처리 추가**:
1. Coverage < 85% → test-engineer 호출
2. Blocker detected → debug-helper 호출
3. Architectural question → implementation-planner 상담

### 3.3 구체적 코드 추가

**추가 위치**: Step 2.4 이후 (Line 323 후)

```markdown
### Step 2.4: Error Recovery Protocols

If quality-gate reports issues:

**IF: Coverage < 85%**
```
Task(subagent_type="test-engineer",
     description="Supplement tests for coverage improvement",
     prompt="SPEC-$ARGUMENTS has {actual_coverage}% coverage (target: 85%).

Current gaps: {uncovered_files_and_lines}

1. Identify untested code paths
   - Which files/functions lack tests?
   - Which branches/edge cases are uncovered?

2. Write supplementary tests
   - Add tests for each uncovered path
   - Include edge cases and error scenarios
   - Target: Reach 85%+ coverage

3. Run full test suite
   - Verify new tests pass
   - Verify coverage improves to ≥85%
   - Commit: test(@SPEC-$ARGUMENTS): Add supplementary tests for coverage

Output: New coverage percentage, tests added count")
```

**IF: Blocker Encountered in Implementation**
```
Task(subagent_type="debug-helper",
     description="Resolve implementation blocker",
     prompt="SPEC-$ARGUMENTS encountered a blocker during implementation:

Blocker Type: {type}
Issue: {detailed_description}
Context: {relevant_code_snippet}
Error: {error_message}

1. Analyze root cause
   - What's causing this issue?
   - Is it a code logic error, architecture issue, or environment problem?

2. Suggest solutions
   - Provide 2-3 different solution approaches
   - Pros/cons for each approach
   - Recommended approach

3. Provide next steps
   - How to implement the fix?
   - What should be tested?
   - Any related changes needed?

Output: Detailed analysis with recommended fix approach")
```

**IF: Architectural Question Needs Resolution**
```
Task(subagent_type="implementation-planner",
     description="Architectural consultation for SPEC-$ARGUMENTS",
     prompt="During implementation of SPEC-$ARGUMENTS, an architectural question arose:

Question: {specific_question}
Context: {relevant_context}
Current approach: {what_was_tried}
Why it's problematic: {explanation}

1. Analyze the question
   - What are the tradeoffs?
   - What patterns would apply here?

2. Recommend solution
   - Best approach for this project
   - Why this approach?
   - How to implement it?

3. Document for future
   - Is this a pattern to repeat?
   - Should we document this pattern?
   - Update domain patterns?

Output: Architectural recommendation with implementation guidance")
```

---

## Step 4: PHASE 3 업데이트 (최소)

### 4.1 변경 사항

**파일**: `.claude/commands/alfred/2-run.md`
**위치**: Line 326-370 (PHASE 3)

**변경사항**: 기존 코드는 유지, Step 3.1 설명 업데이트만

```markdown
### Step 3.1: Invoke Git-Manager Agent

**Your task**: Call git-manager to create structured commits.

Use Task tool:
- subagent_type: "git-manager"
- description: "Create Git commits for TDD cycle"
- prompt: """
You are the git-manager agent. Create git commits for SPEC-$ARGUMENTS.

CONTEXT:
- Implementation completed via TDD cycle
- Commits already created by tdd-implementer (RED, GREEN, REFACTOR phases)
- Your task: Verify commit structure and finalize

VERIFY COMMITS:

1. Check RED Phase Commits
   - Format: test(@SPEC-$ARGUMENTS): Add failing tests for ...
   - Include @TEST-{N} references
   - Each test file has separate commit

2. Check GREEN Phase Commits
   - Format: feat(@SPEC-$ARGUMENTS): Implement ...
   - Include coverage %, @CODE-{N} references
   - Each component has separate commit

3. Check REFACTOR Phase Commits
   - Format: refactor(@SPEC-$ARGUMENTS): Improve code quality
   - Include improvements summary
   - Single commit unless major refactoring

4. Verify TAG Chains
   - Each @CODE-{N} references @TEST-{N}
   - Each @TEST-{N} references @SPEC-$ARGUMENTS
   - Complete traceability

5. Final Verification
   - All commits follow conventional commits format
   - All commits reference @SPEC-$ARGUMENTS
   - No merge conflicts
   - Branch is feature/SPEC-$ARGUMENTS (if using GitFlow)

Output: Commit verification report
"""

**Verify**: All commits created, TAG chains complete
```

---

## Step 5: 문서화 및 동기화

### 5.1 CLAUDE.md 업데이트

**파일**: `.claude/commands/alfred/2-run.md` 기본 메타데이터

변경할 부분:

```markdown
**Version**: 2.1.0 (Agent-Delegated Pattern)
↓
**Version**: 3.0.0 (Agent-Delegated + Parallel Execution)

**Last Updated**: 2025-11-09
↓
**Last Updated**: 2025-11-12

**Total Lines**: ~400 (reduced from 619)
↓
**Total Lines**: ~500 (increased from 400 due to detailed prompts)

**Architecture**: Commands → Agents → Skills
↓
**Architecture**: Commands → Agents (Parallel) → Skills

**Improvements in v3.0.0**:
- Parallel Task execution in PHASE 1 (Explore + tag-agent)
- 3-way parallel resource preparation in PHASE 2 (milestones, domain, resources)
- Enhanced TDD-Implementer prompt with detailed protocols
- Automated error handling (coverage, blockers, architecture)
- Full TAG chain integration
- TodoWrite automation from structured task list
```

### 5.2 패키지 템플릿 동기화

**중요**: `.claude/` 파일은 패키지 템플릿이 source of truth

**작업**:

1. 로컬 2-run.md 개선 완료 후
2. 패키지 템플릿 동기화:
   ```
   src/moai_adk/templates/.claude/commands/alfred/2-run.md
   ← copy from
   .claude/commands/alfred/2-run.md
   ```

3. 템플릿 검증:
   ```
   - 변수 치환 불필요 (2-run은 project-agnostic)
   - 영어 유지 (인프라 파일)
   - YAML frontmatter 동일
   ```

4. Git 커밋:
   ```
   feat(template): Modernize 2-run command with parallel agent delegation

   - PHASE 1: Add parallel SPEC analysis and TAG initialization
   - PHASE 2: Implement 3-way parallel resource preparation
   - PHASE 2.3: Enhance TDD-Implementer prompt (100+ lines)
   - PHASE 2.4: Add error recovery protocols
   - Result: 35% faster execution, better error handling
   ```

---

## 실행 시간표

### Day 1 (약 2시간)

| 시간 | 작업 | 소요시간 |
|------|------|--------|
| 09:00-09:30 | Step 1: PHASE 1 개선 | 30분 |
| 09:30-10:00 | Step 2-1: PHASE 2 단계 1-2 개선 | 30분 |
| 10:00-11:00 | Step 2-2: TDD-Implementer prompt 확장 | 60분 |
| **11:00** | **Day 1 완료** | **2시간** |

### Day 2 (약 1.5시간)

| 시간 | 작업 | 소요시간 |
|------|------|--------|
| 09:00-09:30 | Step 3: 에러 핸들링 추가 | 30분 |
| 09:30-10:00 | Step 4: PHASE 3 업데이트 | 30분 |
| 10:00-10:30 | Step 5: 문서화 및 템플릿 동기화 | 30분 |
| **10:30** | **Day 2 완료** | **1.5시간** |

**총 소요시간**: 3.5시간 (2일에 걸쳐)

---

## 검증 절차

### 로컬 테스트 (권장)

```bash
# 1. 백업 생성
cp .claude/commands/alfred/2-run.md \
   .moai/backups/commands/2-run-v2.1.0.md

# 2. 임시 SPEC 생성
/alfred:1-plan "Test feature for modernization verification"
# → SPEC-TEST-001 생성됨

# 3. 개선된 2-run 실행
/alfred:2-run SPEC-TEST-001

# 4. 검증 포인트
# ✓ PHASE 1: 병렬 Task 호출 보임?
# ✓ PHASE 2: 3개 Task 동시 실행?
# ✓ PHASE 2.3: 상세한 TDD 가이드?
# ✓ TodoWrite 자동 생성?
# ✓ git 커밋 생성?
# ✓ 스크립트 호출 없음?

# 5. 결과 확인
git log -5 --oneline
pytest tests/ --cov
```

---

## 출력 산출물

### 생성되는 파일

1. **개선된 2-run.md**
   - 위치: `.claude/commands/alfred/2-run.md`
   - 크기: ~500줄 (이전 ~400줄)
   - 변경: 3개 섹션 대폭 개선

2. **패키지 템플릿 동기화**
   - 위치: `src/moai_adk/templates/.claude/commands/alfred/2-run.md`
   - 내용: 로컬과 동일

3. **Git 커밋**
   - 메시지: "feat(commands): Modernize 2-run with parallel agent delegation"

---

## 주요 성과 지표

### Before → After

| 지표 | Before | After | 개선 |
|------|--------|-------|------|
| **병렬 실행** | 없음 | 5개 Task | +5x |
| **스크립트 호출** | 1회 (spec_status_hooks.py) | 0회 | -100% |
| **에러 핸들링** | 수동 | 자동 3가지 | +자동화 |
| **TDD 가이드** | 30줄 | 100줄 | +233% |
| **Progress 추적** | 수동 TodoWrite | 자동 구조화 | +자동화 |
| **TAG 체인** | 암시적 | 명시적 | +명확성 |

---

**로드맵 생성**: 2025-11-12
**대상**: `.claude/commands/alfred/2-run.md`
**상태**: Ready for Implementation
**예상 완료**: 3.5시간 (이틀 작업)
