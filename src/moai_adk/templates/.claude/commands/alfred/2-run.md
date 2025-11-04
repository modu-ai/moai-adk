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

> **Critical Note**: ALWAYS invoke `Skill("moai-alfred-ask-user-questions")` before using `AskUserQuestion` tool. This skill provides up-to-date best practices, field specifications, and validation rules for interactive prompts.
>
> **Batched Design**: All AskUserQuestion calls follow batched design principles (1-4 questions per call) to minimize user interaction turns. See CLAUDE.md section "Alfred Command Completion Pattern" for details.

<!-- @CODE:ALF-WORKFLOW-002:CMD-RUN -->

**4-Step Workflow Integration**: This command implements Step 3 of Alfred's workflow (Task Execution with TodoWrite tracking). See CLAUDE.md for full workflow details.

## 🎯 Command Purpose

Execute planned tasks based on SPEC document analysis. Supports TDD implementation, prototyping, and documentation work.

**Run on**: $ARGUMENTS

## 💡 Execution philosophy: "Plan → Run → Sync"

`/alfred:2-run` performs planned tasks through various execution strategies.

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

## 📋 Execution flow

1. **SPEC Analysis**: Requirements extraction and complexity assessment
2. **Implementation Strategy**: Determine optimized approach (TDD, prototype, documentation)
3. **User Confirmation**: Review and approve action plan
4. **Execute Task**: Perform work according to approved plan
5. **Git Operations**: Create step-by-step commits with git-manager

## 🧠 Associated Skills & Agents

| Agent                  | Core Skill                       | Purpose                                 |
| ---------------------- | -------------------------------- | --------------------------------------- |
| implementation-planner | `moai-alfred-language-detection` | Detect language and design architecture |
| tdd-implementer        | `moai-essentials-debug`          | Implement TDD (RED → GREEN → REFACTOR)  |
| quality-gate           | `moai-alfred-trust-validation`   | Verify TRUST 5 principles               |
| git-manager            | `moai-alfred-git-workflow`       | Commit and manage Git workflows         |

**Note**: TUI Survey Skill is used for user confirmations during the run phase and is shared across all interactive prompts.

## 🔗 Associated Agent

- **Phase 1**: implementation-planner (📋 technical architect) - SPEC analysis and execution strategy
- **Phase 2**: tdd-implementer (🔬 senior developer) - Task execution
- **Phase 2.5**: quality-gate (🛡️ Quality Assurance Engineer) - TRUST principle verification (automatic)
- **Phase 3**: git-manager (🚀 Release Engineer) - Git commits

## 💡 Example Usage

Users can run commands as follows:
- `/alfred:2-run SPEC-001` - Run a specific SPEC
- `/alfred:2-run all` - Run all SPECs in batches
- `/alfred:2-run SPEC-003 --test` - Run only tests

---

## 🔍 YOUR TASK: STEP 1 - SPEC Analysis and Execution Planning

Your task is to analyze SPEC requirements and create an execution plan. Follow these steps:

### STEP 1.1: Determine if Codebase Exploration is Needed

Read the SPEC document at `.moai/specs/SPEC-$ARGUMENTS/spec.md`.

IF the SPEC requires understanding existing code patterns (architecture, similar implementations, test patterns), THEN:
  - Invoke the Explore agent using the Task tool
  - Set subagent_type to "Explore"
  - Set description to "Explore existing code structures and patterns"
  - Pass prompt: "SPEC-$ARGUMENTS와 관련된 기존 코드를 탐색해주세요: 유사한 기능 구현 코드 (src/), 참고할 테스트 패턴 (tests/), 아키텍처 패턴 및 디자인 패턴, 현재 라이브러리 및 버전 (package.json, requirements.txt). 상세도 수준: medium"
  - Store exploration results in memory for next step

IF the SPEC does not require codebase exploration (greenfield implementation, simple feature), THEN:
  - Skip exploration phase
  - Proceed directly to STEP 1.2

### STEP 1.2: Create Implementation Plan

Invoke the implementation-planner agent using the Task tool:
  - Set subagent_type to "implementation-planner"
  - Set description to "SPEC analysis and establishment of execution strategy"
  - Pass prompt including:
    - SPEC ID ($ARGUMENTS)
    - Requirements extraction and complexity assessment
    - Library and tool selection (use WebFetch for latest stable versions)
    - TAG chain design
    - Step-by-step execution plan
    - Risk factors and mitigation strategies
    - (Optional) Exploration results from STEP 1.1 if available
  - Request that the agent creates an action plan report

The implementation-planner agent will:
  1. Analyze SPEC requirements and assess complexity
  2. Check technical constraints and dependencies
  3. Detect project language and optimize execution strategy
  4. Search for latest stable library versions using WebFetch
  5. Design TAG chain and execution sequence
  6. Create step-by-step action plan with risk assessment
  7. Generate execution plan report

### STEP 1.3: Present Plan to User for Approval

After the implementation-planner completes the execution plan, invoke `Skill("moai-alfred-ask-user-questions")` first.

Then use the AskUserQuestion tool to obtain explicit user approval:
  - Present the implementation plan in plain text format
  - Ask "Implementation plan is ready. How would you like to proceed?"
  - Provide these options:
    - "Proceed with TDD" - Start RED → GREEN → REFACTOR cycle
    - "Research First" - Invoke Explore agent to study existing code patterns
    - "Modify Strategy" - Request changes to implementation approach
    - "Postpone" - Save plan and return later
  - Wait for user response

**Response handling:**

IF user selects "Proceed with TDD", THEN:
  - Display "Starting RED phase..."
  - Proceed to STEP 2 (TDD Implementation)

IF user selects "Research First", THEN:
  - Display "Codebase exploration complete. Plan updated."
  - Invoke Explore agent to analyze existing codebase
  - Pass exploration results back to implementation-planner
  - Re-generate updated plan with research insights
  - Present updated plan for approval again (recursive - return to STEP 1.3)

IF user selects "Modify Strategy", THEN:
  - Display "Plan modified. Please review updated strategy."
  - Collect strategy modification requests from user
  - Update implementation plan with requested changes
  - Present updated plan for approval again (recursive - return to STEP 1.3)

IF user selects "Postpone", THEN:
  - Display "Plan saved. Resume with `/alfred:2-run SPEC-{ID}`"
  - Save plan to `.moai/specs/SPEC-{ID}/plan.md`
  - Create git commit with message "plan(spec): Save implementation plan for SPEC-{ID}"
  - Stop execution
  - User can resume later by running `/alfred:2-run SPEC-{ID}`

---

## 🚀 YOUR TASK: STEP 2 - Execute Task (After User Approval)

Your task is to execute the approved implementation plan. Follow these steps:

### STEP 2.1: Check Domain Readiness (Optional - Before Implementation)

Read the SPEC metadata to identify required domains.

IF SPEC frontmatter contains `domains:` field OR `.moai/config.json` has `stack.selected_domains`, THEN:
  - For each domain in the list:
    - IF domain is "frontend", THEN invoke Explore agent with prompt: "You are consulting as frontend-expert for TDD implementation. SPEC: [SPEC-{ID}]. Provide implementation readiness check: Component structure recommendations, State management approach, Testing strategy, Accessibility requirements, Performance optimization tips. Output: Brief advisory for tdd-implementer (3-4 key points)"
    - IF domain is "backend", THEN invoke Explore agent with prompt: "You are consulting as backend-expert for TDD implementation. SPEC: [SPEC-{ID}]. Provide implementation readiness check: API contract validation, Database schema requirements, Authentication/authorization patterns, Error handling strategy, Async processing considerations. Output: Brief advisory for tdd-implementer (3-4 key points)"
    - IF domain is "devops", THEN invoke Explore agent with devops-specific readiness check
    - IF domain is "database", THEN invoke Explore agent with database-specific readiness check
    - IF domain is "data-science", THEN invoke Explore agent with data-science-specific readiness check
    - IF domain is "mobile", THEN invoke Explore agent with mobile-specific readiness check
  - Store all domain expert feedback in memory
  - Save advisory to `.moai/specs/SPEC-{ID}/plan.md` under "## Domain Expert Advisory (Implementation Phase)" section

IF no domains specified OR domain expert unavailable, THEN:
  - Skip advisory phase
  - Continue to STEP 2.2 (implementation proceeds regardless)

### STEP 2.2: Invoke TDD Implementer Agent

Invoke the tdd-implementer agent using the Task tool:
  - Set subagent_type to "tdd-implementer"
  - Set description to "Execute task with TDD implementation"
  - Pass prompt including:
    - SPEC ID ($ARGUMENTS)
    - Language settings (conversation_language, conversation_language_name)
    - Code and technical output must be in English
    - Code comments language rules (local project vs package code)
    - Test descriptions and documentation language
    - Skill invocation instructions (moai-alfred-language-detection, language-specific skills, debug, refactor)
    - Approved plan from STEP 1
    - Domain expert advisory from STEP 2.1 (if available)
    - TDD cycle instructions: RED → GREEN → REFACTOR
    - For each TAG: Write failing test (@TEST:ID) → Minimal implementation (@CODE:ID) → Code quality improvement
    - TAG completion verification and progression

The tdd-implementer agent will:
  1. Detect project language and select optimal TDD tools
  2. Execute RED phase: Write failing tests with @TEST:ID tags, verify test failure
  3. Execute GREEN phase: Write minimal implementation with @CODE:ID tags, verify test pass
  4. Execute REFACTOR phase: Improve code quality (remove duplication, explicit naming, structured logging/exception handling)
  5. Verify TAG completion conditions and proceed to next TAG
  6. Repeat cycle for all TAGs in the plan

### STEP 2.3: Monitor TDD Cycle Progress

During tdd-implementer execution:
  - Initialize TodoWrite with all tasks from the implementation plan
  - Set initial status to "pending" for all tasks
  - For each task:
    - Update TodoWrite status from "pending" to "in_progress" (exactly ONE task at a time)
    - Monitor tdd-implementer progress
    - Update TodoWrite status from "in_progress" to "completed" when task finishes
  - Track RED → GREEN → REFACTOR cycle completion for each TAG

IF tdd-implementer encounters errors or test failures, THEN:
  - Keep task in "in_progress" status
  - Invoke Skill("moai-essentials-debug") for troubleshooting
  - Log error details
  - Attempt fix and retry
  - IF error persists after retry, THEN escalate to user for manual intervention

### STEP 2.4: Quality Gate Verification (Automatic)

After tdd-implementer completes all tasks, automatically invoke the quality-gate agent:
  - Set subagent_type to "quality-gate"
  - Set description to "TRUST principle verification and quality validation"
  - Pass prompt to verify:
    - TRUST principles (Test coverage ≥ 85%, Readable code, Unified architecture, Secured, Traceable @TAG chain)
    - Code style (run linter: ESLint/Pylint)
    - Test coverage (run language-specific coverage tools)
    - TAG chain integrity (check orphan TAGs, missing TAGs)
    - Dependency security (check vulnerabilities)

The quality-gate agent will generate a verification report.

**Handle verification results:**

IF verification result is PASS (0 Critical, ≤5 Warnings), THEN:
  - Display "Quality gate passed."
  - Create quality report
  - Proceed to STEP 3 (Git operations)

IF verification result is WARNING (0 Critical, ≥6 Warnings), THEN:
  - Display warning message with details
  - Use AskUserQuestion to ask: "Quality gate has warnings. How would you like to proceed?"
    - Option 1: "Continue to Git commit" - Accept warnings and proceed to STEP 3
    - Option 2: "Fix warnings first" - Return to tdd-implementer to address warnings
  - Wait for user response
  - IF user selects "Continue", THEN proceed to STEP 3
  - IF user selects "Fix warnings first", THEN return to STEP 2.2 with warning details

IF verification result is CRITICAL (≥1 Critical), THEN:
  - Display "Quality gate blocked. Critical issues found."
  - Block Git commits
  - Generate detailed report with file:line information
  - Display report to user
  - Recommend re-invoking tdd-implementer to fix critical issues
  - Stop execution
  - Wait for user to fix issues and re-run command

IF user passed `--skip-quality-check` flag, THEN:
  - Skip quality gate verification
  - Display "Quality gate skipped (user requested)."
  - Proceed directly to STEP 3

---

## 🚀 YOUR TASK: STEP 3 - Git Operations (After Quality Gate)

Your task is to create Git commits for all completed work. Follow these steps:

### STEP 3.1: Invoke Git Manager Agent

Invoke the git-manager agent using the Task tool:
  - Set subagent_type to "git-manager"
  - Set description to "Create Git commits for completed implementation"
  - Pass prompt including:
    - SPEC ID ($ARGUMENTS)
    - List of all completed tasks from TodoWrite
    - TDD cycle phases (RED → GREEN → REFACTOR)
    - Quality gate results
    - Git strategy from `.moai/config.json` (team mode: GitFlow with develop branch)
    - Request structured commits with proper tagging

The git-manager agent will:
  1. Create checkpoint backup before starting Git operations
  2. Create structured commits for each TDD phase:
     - RED commit: "test(SPEC-{ID}): Add failing tests for {feature}"
     - GREEN commit: "feat(SPEC-{ID}): Implement {feature}"
     - REFACTOR commit: "refactor(SPEC-{ID}): Improve {aspect}"
  3. Apply Git strategy based on mode (team: feature branch → develop, solo: direct to main)
  4. Verify commits created successfully
  5. Display commit summary to user

### STEP 3.2: Handle Git Strategy (Team Mode)

Read Git configuration from `.moai/config.json`:
  - Read `git_strategy.team.use_gitflow` field
  - Read `git_strategy.team.develop_branch` field (default: "develop")

IF use_gitflow is true, THEN:
  - Verify current branch is feature branch (feature/SPEC-*)
  - Verify develop branch exists
  - Do NOT create PR to main (forbidden)
  - Wait for user to manually merge develop → main when ready
  - Display: "Commits created on feature branch. Run `/alfred:3-sync` to create PR targeting develop."

IF use_gitflow is false (solo mode), THEN:
  - Commit directly to current branch
  - Push to remote if configured
  - Display: "Commits created and pushed."

### STEP 3.3: Verify Commit Success

After git-manager completes:
  - Run `git log --oneline -5` to verify commits were created
  - Check commit messages follow the format:
    ```
    {type}(SPEC-{ID}): {description}

    {detailed explanation if needed}

    🤖 Generated with Claude Code

    Co-Authored-By: 🎩 Alfred@MoAI
    ```
  - Verify commit author and timestamp
  - Display commit summary to user

IF commits were not created successfully, THEN:
  - Display error message
  - Show git status output
  - Recommend manual Git operations
  - Stop execution

---

## 🎯 YOUR TASK: STEP 4 - Next Steps (Final)

Your task is to guide the user on what to do next. Follow these steps:

### STEP 4.1: Present Next Steps Options

After STEP 3 completes successfully, invoke `Skill("moai-alfred-ask-user-questions")` first.

Then use the AskUserQuestion tool to ask the user:
  - Ask "Implementation is complete. What would you like to do next?"
  - Provide these options:
    - "Synchronize Documentation" - Proceed to /alfred:3-sync for documentation synchronization
    - "Implement More Features" - Continue with /alfred:2-run SPEC-XXX for next feature
    - "New Session" - Execute /clear for better context management (recommended after large implementations)
    - "Complete" - Finish current session
  - Wait for user response

**Response handling:**

IF user selects "Synchronize Documentation", THEN:
  - Display "Starting documentation synchronization..."
  - Inform user to run: `/alfred:3-sync auto`
  - Explain that this will verify TAGs, update docs, and prepare for PR merge

IF user selects "Implement More Features", THEN:
  - Display "Ready for next feature implementation..."
  - Inform user to run: `/alfred:2-run SPEC-YYY` for another feature
  - Note that current session context will be maintained

IF user selects "New Session", THEN:
  - Display "Clearing session for better context management..."
  - Explain that this is recommended after large implementations
  - Inform user they can run any command in the next session

IF user selects "Complete", THEN:
  - Display "Implementation workflow complete!"
  - Recommend running `/alfred:3-sync` manually when ready
  - Suggest reviewing work or planning next features

### STEP 4.2: Display Summary

After user responds, display a plain text summary:
  - SPEC ID implemented
  - Number of TAGs completed
  - Number of commits created
  - Quality gate result (PASS/WARNING/CRITICAL)
  - Next recommended action based on user selection

---

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
 (Choose between "Proceed," "Modify [Content]," or "Abort")
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

## Next steps

**Recommendation**: For better performance and context management, start a new chat session with the `/clear` or `/new` command before proceeding to the next step.

- After task execution is complete, document synchronization proceeds with `/alfred:3-sync`
- All Git operations are dedicated to the git-manager agent to ensure consistency
- Only command-level orchestration is used without direct calls between agents
