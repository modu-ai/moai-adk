---
name: alfred:2-run
description: "Implement all SPECs with SPEC ID to implement (e.g. SPEC-001) or all - Execute planned work (TDD implementation, prototyping, documentation, etc.)" 
argument-hint: "SPEC-ID - All with SPEC ID to implement (e.g. SPEC-001) or all “SPEC Implementation”
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
> Interactive prompts rely on `Skill("moai-alfred-interactive-questions")` so AskUserQuestion renders TUI selection menus for user surveys and approvals.

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

First, the specified SPEC is analyzed to establish an action plan and receive user confirmation.

**The implementation-planner agent automatically loads and analyzes the required documents.**

### 🔍 Browse the code base (recommended)

**If you need to understand existing code structure or find similar patterns** Use the Explore agent first:

```
Invoking the Task tool (Explore agent):
- subagent_type: "Explore"
- description: "Explore existing code structures and patterns"
- prompt: "Please explore existing code related to SPEC-$ARGUMENTS:
 - Similar function implementation code (src/)
 - Test patterns for reference (tests/)
 - Architectural patterns and design patterns
 - Use Current libraries and versions (package.json, requirements.txt)
 thoroughness level: medium"
```

**When to use the Explore Agent**:
- ✅ When you need to understand the existing code structure/pattern
- ✅ When you need to refer to how a similar function is implemented
- ✅ When you need to understand the architectural rules of the project
- ✅ Check the library and version being used

### ⚙️ How to call an agent

**In STEP 1, we call the implementation-planner agent using the Task tool**:

```
Task tool call example:
- subagent_type: "implementation-planner"
- description: "SPEC analysis and establishment of execution strategy"
- prompt: "Please analyze the SPEC of $ARGUMENTS and establish an execution plan.
 It must include the following:
 1. SPEC requirements extraction and complexity assessment
 2. Library and tool selection (using WebFetch)
          3. TAG chain design
 4. Step-by-step execution plan
 5. Risks and response plans
6. Create action plan and use `Skill("moai-alfred-interactive-questions")` to confirm the next action with the user
 (Optional) Explore results: $EXPLORE_RESULTS"
```

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

## 🚀 STEP 2: Execute task (after user approval)

After user approval (gathered through `Skill("moai-alfred-interactive-questions")`), **call the tdd-implementer agent using the Task tool**.

### ⚙️ How to call an agent

**STEP 2 calls tdd-implementer using the Task tool**:

```
Task tool call example:
- subagent_type: "tdd-implementer"
- description: "Execute task"
- prompt: "Please execute the task according to the plan approved in STEP 1.
 For TDD scenario:
 - Perform RED → GREEN → REFACTOR cycle,
 Perform the following for each TAG:
 1. RED Phase: Write a test that fails with the @TEST:ID tag
 2. GREEN Phase: Minimal implementation with the @CODE:ID tag
 3. REFACTOR Phase: Improve code quality
 4. Verify TAG completion conditions and proceed to the next TAG

Execute on: $ARGUMENTS"
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
5. **Create action plan**: Create a structured plan and, via `Skill("moai-alfred-interactive-questions")`, collect user approval before proceeding

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

> For more information: `.moai/memory/development-guide.md` - see section "Context Engineering"

### Core strategy of this command

**Load first**: `.moai/specs/SPEC-XXX/spec.md` (implementation target requirement)

**Recommendation**: Job execution completed successfully. You can experience better performance and context management by starting a new chat session with the `/clear` or `/new` command before proceeding to the next step (`/alfred:3-sync`).

---

## Next steps

**Recommendation**: For better performance and context management, start a new chat session with the `/clear` or `/new` command before proceeding to the next step.

- After task execution is complete, document synchronization proceeds with `/alfred:3-sync`
- All Git operations are dedicated to the git-manager agent to ensure consistency
- Only command-level orchestration is used without direct calls between agents
