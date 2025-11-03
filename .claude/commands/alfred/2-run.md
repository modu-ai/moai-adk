---
name: alfred:2-run
description: "Execute TDD implementation cycle"
argument-hint: "SPEC-ID - All with SPEC ID to implement (e.g. SPEC-001) or all \"SPEC Implementation\""
allowed-tools:
  - Read
  - Write
  - Edit
  - MultiEdit
  - Bash(python3:*)
  - Bash(pytest:*)
  - Bash(npm:*)
  - Bash(node:*)
  - Bash(git:*)
  - Task
  - WebFetch
  - Grep
  - Glob
  - TodoWrite
---

# ⚒️ MoAI-ADK Phase 2: Run the plan - Flexible implementation strategy
> **Note**: Interactive prompts use `AskUserQuestion tool (documented in moai-alfred-interactive-questions skill)` for TUI selection menus. The skill is loaded on-demand when user interaction is required.
>
> **Batched Design**: All AskUserQuestion calls follow batched design principles (1-4 questions per call) to minimize user interaction turns. See CLAUDE.md section "Alfred Command Completion Pattern" for details.

<!-- @CODE:ALF-WORKFLOW-002:CMD-RUN -->

**4-Step Workflow Integration**: This command implements Step 3 of Alfred's workflow (Task Execution with TodoWrite tracking). See CLAUDE.md for full workflow details.

## 🎯 Command Purpose

Analyze SPEC documents to execute planned tasks. It supports not only TDD implementation but also various execution scenarios such as prototyping and documentation work.

**Run on**: $ARGUMENTS

## 💡 Execution philosophy: “Plan → Run → Sync”

`/alfred:2-run` is a general-purpose command that does not simply "build" code, but **performs** a planned task.

### 3 main scenarios

#### Scenario 1: TDD implementation (main method) ⭐
```bash
/alfred:2-run SPEC-AUTH-001
→ RED → GREEN → REFACTOR
→ Implement high-quality code through test-driven development
```

#### Scenario 2: Prototyping
```bash
/alfred:2-run SPEC-PROTO-001
→ Prototype implementation for quick verification
→ Quick feedback with minimal testing
```

#### Scenario 3: Documentation tasks
```bash
/alfred:2-run SPEC-DOCS-001
→ Writing documentation and generating sample code
→ API documentation, tutorials, guides, etc.
```

> **Standard two-step workflow** (see `CLAUDE.md` - "Alfred Command Execution Pattern" for details)

## 📋 Execution flow

1. **SPEC Analysis**: Requirements extraction and complexity assessment
2. **Establishment of implementation strategy**: Determine the optimized approach for each language (TDD, prototype, documentation, etc.)
3. **User Confirmation**: Review and approve action plan
4. **Execute work**: Perform work according to the approved plan
5. **Git Operations**: Creating step-by-step commits with git-manager

## 🧠 Associated Skills & Agents

| Agent                  | Core Skill                       | Purpose                                 |
| ---------------------- | -------------------------------- | --------------------------------------- |
| implementation-planner | `moai-alfred-language-detection` | Detect language and design architecture |
| tdd-implementer        | `moai-essentials-debug`          | Implement TDD (RED → GREEN → REFACTOR)  |
| quality-gate           | `moai-alfred-trust-validation`   | Verify TRUST 5 principles               |
| git-manager            | `moai-alfred-git-workflow`       | Commit and manage Git workflows         |

**Note**: TUI Survey Skill is used for user confirmations during the run phase and is shared across all interactive prompts.

## 🔗 Associated Agent

- **Phase 1**: implementation-planner (📋 technical architect) - SPEC analysis and establishment of execution strategy
- **Phase 2**: tdd-implementer (🔬 senior developer) - Dedicated to execution work
- **Phase 2.5**: quality-gate (🛡️ Quality Assurance Engineer) - TRUST principle verification (automatically)
- **Phase 3**: git-manager (🚀 Release Engineer) - Dedicated to Git commits

## 💡 Example of use

Users can run commands as follows:
- `/alfred:2-run SPEC-001` - Run a specific SPEC
- `/alfred:2-run all` - Run all SPECs in batches
- `/alfred:2-run SPEC-003 --test` - Run only tests

## 🔍 STEP 1: SPEC analysis and execution plan establishment

STEP 1 consists of **two independent phases** to provide flexible workflow based on task complexity:

### 📋 STEP 1 Workflow Overview

```
┌─────────────────────────────────────────────────────────────┐
│ STEP 1: SPEC Analysis & Planning                           │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Phase A (OPTIONAL)                                         │
│  ┌─────────────────────────────────────────┐               │
│  │ 🔍 Explore Agent                        │               │
│  │ • Browse existing codebase              │               │
│  │ • Find similar implementations          │               │
│  │ • Identify patterns & architecture      │               │
│  └─────────────────────────────────────────┘               │
│                    ↓                                        │
│          (exploration results)                              │
│                    ↓                                        │
│  Phase B (REQUIRED)                                         │
│  ┌─────────────────────────────────────────┐               │
│  │ ⚙️ implementation-planner Agent         │               │
│  │ • Analyze SPEC requirements             │               │
│  │ • Design execution strategy             │               │
│  │ • Create implementation plan            │               │
│  │ • Request user approval                 │               │
│  └─────────────────────────────────────────┘               │
│                    ↓                                        │
│          (user approval via AskUserQuestion)                │
│                    ↓                                        │
│              PROCEED TO STEP 2                              │
└─────────────────────────────────────────────────────────────┘
```

**Key Points**:
- **Phase A is optional** - Skip if you don't need to explore existing code
- **Phase B is required** - Always runs to analyze SPEC and create execution plan
- **Results flow forward** - Exploration results (if any) are passed to implementation-planner

---

### 🔍 Phase A: Codebase Exploration (OPTIONAL)

**Use the Explore agent when you need to understand existing code before planning.**

#### When to use Phase A:

- ✅ Need to understand existing code structure/patterns
- ✅ Need to find similar function implementations for reference
- ✅ Need to understand project architectural rules
- ✅ Need to check libraries and versions being used

#### How to invoke Explore agent:

```
Invoking the Task tool (Explore agent):
- subagent_type: "Explore"
- description: "Explore existing code structures and patterns"
- prompt: "SPEC-$ARGUMENTS와 관련된 기존 코드를 탐색해주세요:
 - 유사한 기능 구현 코드 (src/)
 - 참고할 테스트 패턴 (tests/)
 - 아키텍처 패턴 및 디자인 패턴
 - 현재 라이브러리 및 버전 (package.json, requirements.txt)
 상세도 수준: medium"
```

**Note**: If you skip Phase A, proceed directly to Phase B.

---

### ⚙️ Phase B: Execution Planning (REQUIRED)

**Call the implementation-planner agent to analyze SPEC and establish execution strategy.**

This phase is **always required** regardless of whether Phase A was executed.

#### How to invoke implementation-planner:

```
Task tool call:
- subagent_type: "implementation-planner"
- description: "SPEC analysis and establishment of execution strategy"
- prompt: "$ARGUMENTS의 SPEC을 분석하고 실행 계획을 수립해주세요.
 다음을 포함해야 합니다:
 1. SPEC 요구사항 추출 및 복잡도 평가
 2. 라이브러리 및 도구 선택 (WebFetch 사용)
 3. TAG 체인 설계
 4. 단계별 실행 계획
 5. 위험 요소 및 대응 계획
 6. 행동 계획을 작성하고 `AskUserQuestion 도구 (moai-alfred-interactive-questions 스킬 참고)`로 사용자와 다음 단계를 확인합니다
 (선택사항) 탐색 결과: $EXPLORE_RESULTS"
```

**Note**: If Phase A was executed, pass the exploration results via `$EXPLORE_RESULTS` variable.

### SPEC analysis in progress

1. **SPEC document analysis**
 - Requirements extraction and complexity assessment
 - Check technical constraints
 - Dependency and impact scope analysis
 - (Optional) Identify existing code structure based on Explore results

2. **Establish execution strategy**
 - Detect project language and optimize execution strategy
 - Determine approach (TDD, prototyping, documentation, etc.)
 - Estimate expected work scope and time

3. **Check and specify library versions (required)**
 - **Web search**: Check the latest stable versions of all libraries to be used through `WebSearch`
 - **Specify versions**: Specify the exact version for each library in the implementation plan report (e.g. `fastapi>=0.118.3`)
 - **Stability priority**: Exclude beta/alpha versions, select only production stable versions
 - **Check compatibility**: Verify version compatibility between libraries
 - **Search keyword examples**:
     - `"FastAPI latest stable version 2025"`
     - `"SQLAlchemy 2.0 latest stable version 2025"`
     - `"React 18 latest stable version 2025"`

4. **Report action plan**
 - Present step-by-step action plan
 - Identify potential risk factors
 - Set quality gate checkpoints
 - **Specify library version (required)**

### User verification steps

After reviewing the action plan, select one of the following:
- **"Proceed"** or **"Start"**: Start executing the task as planned
- **"Modify [Content]"**: Request a plan modification
- **"Abort"**: Stop the task

---

## Implementation Strategy Approval

After the execution plan is ready, Alfred uses `AskUserQuestion` tool (documented in moai-alfred-interactive-questions skill) to obtain explicit user approval before proceeding to TDD implementation.

**Example AskUserQuestion Call**:
```python
AskUserQuestion(
    questions=[
        {
            "question": "Implementation plan is ready. How would you like to proceed?",
            "header": "Implementation Approval",
            "multiSelect": false,
            "options": [
                {
                    "label": "✅ Proceed with TDD",
                    "description": "Start RED → GREEN → REFACTOR cycle"
                },
                {
                    "label": "🔍 Research First",
                    "description": "Invoke Explore agent to study existing code patterns"
                },
                {
                    "label": "🔄 Modify Strategy",
                    "description": "Request changes to implementation approach"
                },
                {
                    "label": "⏸️ Postpone",
                    "description": "Save plan and return later"
                }
            ]
        }
    ]
)
```

**Response Processing**:
- **"✅ Proceed with TDD"** (`answers["0"] === "✅ Proceed with TDD"`) → Execute Phase 2
  - Proceed directly to STEP 2 (TDD implementation)
  - Invoke tdd-implementer agent with approved plan
  - Begin RED phase (write failing tests)
  - Display: "🔴 Starting RED phase..."

- **"🔍 Research First"** (`answers["0"] === "🔍 Research First"`) → Run exploration first
  - Invoke Explore agent to analyze existing codebase
  - Pass exploration results to implementation-planner
  - Re-generate plan with research insights
  - Re-present plan for approval
  - Display: "🔍 Codebase exploration complete. Plan updated."

- **"🔄 Modify Strategy"** (`answers["0"] === "🔄 Modify Strategy"`) → Revise plan
  - Collect strategy modification requests from user
  - Update implementation plan with changes
  - Re-present for approval (recursive)
  - Display: "🔄 Plan modified. Please review updated strategy."

- **"⏸️ Postpone"** (`answers["0"] === "⏸️ Postpone"`) → Save and resume later
  - Save plan to `.moai/specs/SPEC-{ID}/plan.md`
  - Commit with message "plan(spec): Save implementation plan for SPEC-{ID}"
  - User can resume with `/alfred:2-run SPEC-{ID}`
  - Display: "⏸️ Plan saved. Resume with `/alfred:2-run SPEC-{ID}`"

---

## 🚀 STEP 2: Execute task (after user approval)

After user approval (collected via the AskUserQuestion decision point above), **call the tdd-implementer agent using the Task tool**.

---

### 2.0.5 Domain Readiness Check (Automatic - Before Implementation)

**Purpose**: Load domain-expert agents as "implementation advisors" based on SPEC.stack metadata

**When to run**: After user approval, BEFORE invoking tdd-implementer

**Detection Logic**:

Alfred reads the SPEC metadata to identify required domains:

```bash
# Read SPEC metadata
spec_metadata=$(grep "^domains:" .moai/specs/SPEC-{ID}/spec.md)

# Or read from config.json
selected_domains=$(jq -r '.stack.selected_domains[]' .moai/config.json)
```

**Domain Expert Advisory Pattern**:

| Domain | Readiness Check | Advisory Focus |
|--------|----------------|----------------|
| **Frontend** | Component structure, testing strategy, state management | Component hierarchy, React/Vue best practices, UI testing patterns |
| **Backend** | API contract, database schema, async patterns | RESTful design, database indexing, error handling, authentication |
| **DevOps** | Docker readiness, environment variables, health checks | Containerization, CI/CD integration, deployment strategies |
| **Database** | Schema design, migration strategy, indexing | Data modeling, query optimization, migration safety |
| **Data Science** | Data pipeline design, notebook structure | ETL patterns, data validation, model versioning |
| **Mobile** | Platform-specific requirements, app lifecycle | Native integration, state management, offline support |

**Example Invocation** (Frontend + Backend detected):

```python
# Read SPEC metadata
spec_domains = ["frontend", "backend"]  # from SPEC frontmatter

# Invoke domain experts BEFORE tdd-implementer
for domain in spec_domains:
    if domain == "frontend":
        Task(
            subagent_type="Explore",
            prompt="""You are consulting as frontend-expert for TDD implementation.

SPEC: [SPEC-UI-001 - User Dashboard Component]

Provide implementation readiness check:
1. Component structure recommendations
2. State management approach (Redux/Zustand/Context)
3. Testing strategy (Jest + Testing Library)
4. Accessibility requirements
5. Performance optimization tips

Output: Brief advisory for tdd-implementer (3-4 key points)"""
        )

    if domain == "backend":
        Task(
            subagent_type="Explore",
            prompt="""You are consulting as backend-expert for TDD implementation.

SPEC: [SPEC-API-001 - Authentication Endpoints]

Provide implementation readiness check:
1. API contract validation
2. Database schema requirements
3. Authentication/authorization patterns
4. Error handling strategy
5. Async processing considerations

Output: Brief advisory for tdd-implementer (3-4 key points)"""
        )
```

**Output Format** (Stored in SPEC plan.md):

```markdown
## Domain Expert Advisory (Implementation Phase)

### Frontend Readiness
- Component structure: Use compound component pattern for Dashboard
- State management: Recommend Zustand for lightweight state
- Testing: Prioritize user interaction tests over implementation details
- Performance: Implement React.memo for expensive components

### Backend Readiness
- API contract: OpenAPI 3.0 spec generated from FastAPI
- Database schema: Add index on user_id and created_at columns
- Authentication: Use JWT with refresh token rotation
- Async: Use background tasks for email notifications
```

**Integration with tdd-implementer**:

```python
# Pass domain expert feedback to tdd-implementer
Task(
    subagent_type="tdd-implementer",
    prompt="""You are tdd-implementer agent.

SPEC: SPEC-{ID}

DOMAIN EXPERT ADVISORY:
{domain_expert_feedback}

Execute TDD implementation considering domain expert guidance.
Follow RED → GREEN → REFACTOR cycle with domain best practices.

$ARGUMENTS"""
)
```

**Graceful Degradation**:
- If SPEC.stack.domains missing → Skip advisory (greenfield implementation)
- If domain expert unavailable → Continue with tdd-implementer only
- Advisory is non-blocking (implementation proceeds regardless)

---

### ⚙️ How to call an agent

**STEP 2 calls tdd-implementer using the Task tool**:

```
Call the Task tool:
- subagent_type: "tdd-implementer"
- description: "Execute task with TDD implementation"
- prompt: """당신은 tdd-implementer 에이전트입니다.

언어 설정:
- 대화_언어: {{CONVERSATION_LANGUAGE}}
- 언어명: {{CONVERSATION_LANGUAGE_NAME}}

중요 지시사항:
**Code and technical output MUST be in English.** This ensures global compatibility and maintainability.

코드 문법 및 키워드: 영어 (고정).
코드 주석:
- 로컬 프로젝트 코드: 반드시 {{CONVERSATION_LANGUAGE}}로 작성
- 패키지 코드 (src/moai_adk/): 반드시 영어로 작성 (글로벌 배포용)
테스트 설명 및 문서: 반드시 {{CONVERSATION_LANGUAGE}}로 작성.

스킬 호출:
필요 시 명시적 Skill() 호출 사용:
- Skill("moai-alfred-language-detection") - 프로젝트 언어 감지
- Skill("moai-lang-python") 또는 언어별 스킬 - 베스트 프랙티스
- Skill("moai-essentials-debug") - 테스트 실패 시
- Skill("moai-essentials-refactor") - REFACTOR 단계에서

작업: STEP 1에서 승인된 계획에 따라 작업을 실행합니다.

TDD 시나리오의 경우:
- RED → GREEN → REFACTOR 사이클 수행
- 각 TAG에 대해 다음을 수행:
  1. RED 단계: @TEST:ID 태그로 실패하는 테스트 작성
  2. GREEN 단계: @CODE:ID 태그로 최소한의 구현
  3. REFACTOR 단계: 코드 품질 개선
  4. TAG 완료 조건 확인 및 다음 TAG 진행

실행 대상: $ARGUMENTS"""
```

## 🔗 TDD optimization for each language

### Project language detection and optimal routing

`tdd-implementer` automatically detects the language of your project and selects the optimal TDD tools and workflow:

- **Language detection**: Analyze project files (package.json, pyproject.toml, go.mod, etc.)
- **Tool selection**: Automatically select the optimal test framework for each language
- **TAG application**: Write @TAG annotations directly in code files
- **Run cycle**: RED → GREEN → REFACTOR sequential process

### TDD tool mapping

#### Backend/System

| SPEC Type           | Implementation language | Test Framework         | Performance Goals | Coverage Goals |
| ------------------- | ----------------------- | ---------------------- | ----------------- | -------------- |
| **CLI/System**      | TypeScript              | jest + ts-node         | < 18ms            | 95%+           |
| **API/Backend**     | TypeScript              | Jest + SuperTest       | < 50ms            | 90%+           |
| **Frontend**        | TypeScript              | Jest + Testing Library | < 100ms           | 85%+           |
| **Data Processing** | TypeScript              | Jest + Mock            | < 200ms           | 85%+           |
| **Python Project**  | Python                  | pytest + mypy          | Custom            | 85%+           |

#### Mobile Framework

| SPEC Type        | Implementation language | Test Framework             | Performance Goals | Coverage Goals |
| ---------------- | ----------------------- | -------------------------- | ----------------- | -------------- |
| **Flutter App**  | Dart                    | flutter test + widget test | < 100ms           | 85%+           |
| **React Native** | TypeScript              | Jest + RN Testing Library  | < 100ms           | 85%+           |
| **iOS App**      | Swift                   | XCTest + XCUITest          | < 150ms           | 80%+           |
| **Android App**  | Kotlin                  | JUnit + Espresso           | < 150ms           | 80%+           |

## 🚀 Optimized agent collaboration structure

- **Phase 1**: `implementation-planner` agent analyzes SPEC and establishes execution strategy
- **Phase 2**: `tdd-implementer` agent executes tasks (TDD cycle, prototyping, documentation, etc.)
- **Phase 2.5**: `quality-gate` agent verifies TRUST principle and quality verification (automatically)
- **Phase 3**: `git-manager` agent processes all commits at once after task completion
- **Single responsibility principle**: Each agent is responsible only for its own area of expertise
- **Inter-agent call prohibited**: Each agent runs independently, sequential calls are made only at the command level

## 🔄 Step 2 Workflow Execution Order

### Phase 1: Analysis and planning phase

The `implementation-planner` agent does the following:

1. **SPEC document analysis**: Requirements extraction and complexity assessment of specified SPEC ID
2. **Library selection**: Check the latest stable version and verify compatibility through WebFetch
3. **TAG chain design**: Determine TAG order and dependency
4. **Establishment of implementation strategy**: Step-by-step implementation plan and risk identification
5. **Create action plan**: Create a structured plan and, via `AskUserQuestion tool (documented in moai-alfred-interactive-questions skill)`, collect user approval before proceeding

### Phase 2: Task execution phase (after approval)

The `tdd-implementer` agent performs **TAG-by-TAG** after user approval (based on TDD scenario):

1. **RED Phase**: Write a failing test (add @TEST:ID tag) and check for failure
2. **GREEN Phase**: Write minimal code that passes the test (add @CODE:ID tag)
3. **REFACTOR Phase**: Improve code quality (without changing functionality)
4. **TAG completion confirmation**: Verify the completion conditions of each TAG and proceed to the next TAG

### Phase 2.5: Quality verification gate (automatic execution)

After the job execution is complete, the `quality-gate` agent **automatically** performs quality verification.

**Automatic execution conditions**:
- Automatically invoked upon completion of task execution
- Manually invoked upon user request

**Verification items**:
- **TRUST principle verification**: Trust-checker script execution and result parsing
 - T (Testable): Test coverage ≥ 85%
 - R (Readable): Code readability (file≤300 LOC, function≤50 LOC, Complexity≤10)
 - U (Unified): Architectural integrity
 - S (Secured): No security vulnerabilities
 - T (Traceable): @TAG chain integrity
- **Code style**: Run and verify linter (ESLint/Pylint)
- **Test Coverage**: Run language-specific coverage tools and verify goal achievement
- **TAG chain verification**: Check orphan TAGs, missing TAGs
- **Dependency verification**: Check security vulnerabilities

**How ​​it works**: When Alfred completes job execution, it automatically calls the quality-gate agent to perform quality verification.

**Handling verification results**:

✅ **PASS (0 Critical, 5 or less Warnings)**:
- Proceed to Phase 3 (Git work)
- Create a quality report

⚠️ **WARNING (0 Critical, 6 or more Warnings)**:
- Display warning
- User choice: "Continue" or "Re-verify after modification"

❌ **CRITICAL (1 or more Critical)**:
- Block Git commits
- Detailed report on items requiring improvement (including file: line information)
- Recommended tdd-implementer re-invocation

**Skip verification option**: To skip quality verification, use the `--skip-quality-check` option.

### Phase 3: Git operations (git-manager)

After the `git-manager` agent completes the task **at once**:

1. **Create checkpoint**: Backup point before starting work
2. **Structured Commit**: Step-by-step commit creation (RED→GREEN→REFACTOR for TDD)
3. **Final synchronization**: Apply Git strategy for each mode and remote synchronization


## 📋 STEP 1 Execution Guide: SPEC Analysis and Planning

### 1. SPEC document analysis

Alfred calls the implementation-planner agent to check the SPEC document and create an execution plan.

#### Analysis Checklist

- [ ] **Requirements clarity**: Are the functional requirements in the SPEC specific?
- [ ] **Technical constraints**: Check performance, compatibility, and security requirements
- [ ] **Dependency analysis**: Connection points with existing code and scope of impact
- [ ] **Complexity assessment**: Implementation difficulty and expected workload

### 2. Determine implementation strategy

#### TypeScript execution criteria

| SPEC characteristics | execution language  | Reason                                                    |
| -------------------- | ------------------- | --------------------------------------------------------- |
| CLI/System Tools     | TypeScript          | High performance (18ms), type safety, SQLite3 integration |
| API/Backend          | TypeScript          | Node.js ecosystem, Express/Fastify compatibility          |
| Frontend             | TypeScript          | React/Vue native support                                  |
| data processing      | TypeScript          | High-performance asynchronous processing, type safety     |
| User Python Project  | Python tool support | MoAI-ADK provides Python project development tools        |

#### Approach

- **Bottom-up**: Utility → Service → API
- **Top-down**: API → Service → Utility
- **Middle-out**: Core logic → Bidirectional expansion

### 3. Generate action plan report

Present your plan in the following format:

```
## Execution Plan Report: [SPEC-ID]

### 📊 Analysis Results
- **Complexity**: [Low/Medium/High]
- **Estimated Work Time**: [Time Estimation]
- **Key Technical Challenges**: [Technical Difficulties]

### 🎯 Execution Strategy
- **Language of choice**: [Python/TypeScript + Reason]
- **Approach**: [Bottom-up/Top-down/Middle-out or Prototype/Documentation]
- **Core module**: [Major work target]

### 📦 Library version (required - based on web search)
**Backend dependencies** (example):
| package    | Latest stable version | installation command |
| ---------- | --------------------- | -------------------- |
| FastAPI    | 0.118.3               | fastapi>=0.118.3     |
| SQLAlchemy | 2.0.43                | sqlalchemy>=2.0.43   |

**Frontend dependency** (example):
| package | Latest stable version | installation command |
| ------- | --------------------- | -------------------- |
| React   | 18.3.1                | react@^18.3.1        |
| Vite    | 7.1.9                 | vite@^7.1.9          |

**Important Compatibility Information**:
- [Specific Version Requirements]
- [Known Compatibility Issues]

### ⚠️ Risk Factors
- **Technical Risk**: [Expected Issues]
- **Dependency Risk**: [External Dependency Issues]
- **Schedule Risk**: [Possible Delay]

### ✅ Quality Gates
- **Test Coverage**: [Goal %]
- **Performance Goals**: [Specific Metrics]
- **Security Checkpoints**: [Verification Items]

---
**Approval Request**: Do you want to proceed with the above plan?
 (Choose between “Proceed,” “Modify [Content],” or “Abort”)
```

---

## 🚀 STEP 2 Execution Guide: Execute Task (After Approval)

Only if the user selects **"Proceed"** or **"Start"** will Alfred call the tdd-implementer agent to start the task.

### TDD step-by-step guide

1. **RED**: Writing failure tests with Given/When/Then structure. Follow test file rules for each language and simply record failure logs. 
2. **GREEN**: Add only the minimal implementation that makes the tests pass. Optimization is postponed to the REFACTOR stage.
3. **REFACTOR**: Removal of duplication, explicit naming, structured logging/exception handling enhancements. Split into additional commits if necessary.

**TRUST 5 Principles Linkage** (Details: `development-guide.md` - "TRUST 5 Principles"):
- **T (Test First)**: Writing SPEC-based tests in the RED stage
- **R (Readable)**: Readability in the REFACTOR stage Improvement (file≤300 LOC, function≤50 LOC)
- **T (Trackable)**: Maintain @TAG traceability at all stages.

> TRUST 5 principles provide only basic recommendations, so if you need a structure that exceeds `simplicity_threshold`, proceed with the basis in SPEC or ADR.

## Agent role separation

### implementation-planner dedicated area

- SPEC document analysis and requirements extraction
- Library selection and version management
- TAG chain design and sequence decision
- Establishment of implementation strategy and identification of risks
- Creation of execution plan

### tdd-implementer dedicated area

- Execute tasks (TDD, prototyping, documentation, etc.) 
 - Write and run tests (TDD scenarios) 
 - Add and manage TAG comments 
 - Improve code quality (refactoring) 
 - Run language-specific linters/formatters

### Quality-gate dedicated area

- TRUST principle verification
- Code style verification
- Test coverage verification
- TAG chain integrity verification
- Dependency security verification

### git-manager dedicated area

- All Git commit operations (add, commit, push)
- Checkpoint creation for each task stage
- Apply commit strategy for each mode
- Git branch/tag management
- Remote synchronization processing

## Quality Gate Checklist

- Test coverage ≥ `.moai/config.json.test_coverage_target` (default 85%)
- Pass linter/formatter (`ruff`, `eslint --fix`, `gofmt`, etc.)
- Check presence of structured logging or observation tool call
- @TAG update needed changes note (used by doc-syncer in next step)

---

## 🧠 Context Management

> For more information: Skill("moai-alfred-dev-guide") - see section "Context Engineering"

### Core strategy of this command

**Load first**: `.moai/specs/SPEC-XXX/spec.md` (implementation target requirement)

**Recommendation**: Job execution completed successfully. You can experience better performance and context management by starting a new chat session with the `/clear` or `/new` command before proceeding to the next step (`/alfred:3-sync`).

---

## Final Step

### After STEP 3 (git-manager) Completes

Alfred calls AskUserQuestion to collect user's next action:

**Example AskUserQuestion Call**:
```python
AskUserQuestion(
    questions=[
        {
            "question": "Implementation is complete. What would you like to do next?",
            "header": "Next Steps",
            "multiSelect": false,
            "options": [
                {
                    "label": "📚 Synchronize Documentation",
                    "description": "Proceed to /alfred:3-sync for documentation synchronization"
                },
                {
                    "label": "🔨 Implement More Features",
                    "description": "Continue with /alfred:2-run SPEC-XXX for next feature"
                },
                {
                    "label": "🔄 New Session",
                    "description": "Execute /clear for better context management (recommended)"
                },
                {
                    "label": "✅ Complete",
                    "description": "Finish current session"
                }
            ]
        }
    ]
)
```

**Response Processing**:
- **"📚 Synchronize Documentation"** (`answers["0"] === "📚 Synchronize Documentation"`) → Proceed to `/alfred:3-sync`
  - Display: "Starting documentation synchronization..."
  - User can execute: `/alfred:3-sync auto`
  - This verifies TAGs, updates docs, and prepares for PR merge

- **"🔨 Implement More Features"** (`answers["0"] === "🔨 Implement More Features"`) → Continue implementation
  - Display: "Ready for next feature implementation..."
  - User can run: `/alfred:2-run SPEC-YYY` for another feature
  - Maintains current session context

- **"🔄 New Session"** (`answers["0"] === "🔄 New Session"`) → Clear and restart
  - Display: "⏳ Clearing session for better context management..."
  - Recommended after large implementations
  - Next session: Can run any command

- **"✅ Complete"** (`answers["0"] === "✅ Complete"`) → End current workflow
  - Display: "Implementation workflow complete!"
  - Recommend next manual steps via `/alfred:3-sync`
  - User can review work or plan next features

---

## Next steps

**Recommendation**: For better performance and context management, start a new chat session with the `/clear` or `/new` command before proceeding to the next step.

- After task execution is complete, document synchronization proceeds with `/alfred:3-sync`
- All Git operations are dedicated to the git-manager agent to ensure consistency
- Only command-level orchestration is used without direct calls between agents
