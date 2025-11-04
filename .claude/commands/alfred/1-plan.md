---
name: alfred:1-plan
description: "Define specifications and create development branch"
argument-hint: Title 1 Title 2 ... | SPEC-ID modifications
allowed-tools:
- Read
- Write
- Edit
- MultiEdit
- Grep
- Glob
- TodoWrite
- Bash(git:*)
- Bash(gh:*)
- Bash(rg:*)
- Bash(mkdir:*)
---

# 🏗️ MoAI-ADK Step 1: Establish a plan (Plan) - Always make a plan first and then proceed.

> **Critical Note**: ALWAYS invoke `Skill("moai-alfred-ask-user-questions")` before using `AskUserQuestion` tool. This skill provides up-to-date best practices, field specifications, and validation rules for interactive prompts.
>
> **Batched Design**: All AskUserQuestion calls follow batched design principles (1-4 questions per call) to minimize user interaction turns. See CLAUDE.md section "Alfred Command Completion Pattern" for details.

<!-- @CODE:ALF-WORKFLOW-001:CMD-PLAN -->

**4-Step Workflow Integration**: This command implements Steps 1-2 of Alfred's workflow (Intent Understanding → Plan Creation). See CLAUDE.md for full workflow details.

## 🎯 Command Purpose

**"Plan → Run → Sync"** As the first step in the workflow, it supports the entire planning process from ideation to plan creation.

**Plan for**: $ARGUMENTS

## 🤖 CodeRabbit AI Integration (Local Only)

This local environment includes CodeRabbit AI review integration for SPEC documents:

**Automatic workflows:**
- ✅ SPEC review: CodeRabbit analyzes SPEC metadata and EARS structure
- ✅ GitHub Issue sync: SPEC files automatically create/update GitHub Issues
- ✅ Auto-approval: Draft PRs are approved when quality meets standards (80%+)
- ✅ SPEC quality validation: Checklist for metadata, structure, and content

**Scope:**
- 🏠 **Local environment**: Full CodeRabbit integration with auto-approval
- 📦 **Published packages**: Users get GitHub Issue sync only (no CodeRabbit)

> See `.coderabbit.yaml` for detailed review rules and SPEC validation checklist

## 💡 Planning philosophy: "Always make a plan first and then proceed."

`/alfred:1-plan` is a general-purpose command that **creates a plan**, rather than simply "creating" a SPEC document.

### 3 main scenarios

#### Scenario 1: Creating a Plan (Primary Method) ⭐
```bash
/alfred:1-plan "User authentication function"
→ Refine idea
→ Requirements specification using EARS syntax
→ Create feature/SPEC-XXX branch
→ Create Draft PR
```

#### Scenario 2: Brainstorming
```bash
/alfred:1-plan "Payment system improvement idea"
→ Organizing and structuring ideas
→ Deriving requirements candidates
→ Technical review and risk analysis
```

#### Scenario 3: Improve existing SPEC
```bash
/alfred:1-plan "SPEC-AUTH-001 Security Enhancement"
→ Analyze existing plan
→ Establish improvement direction
→ Create new version plan
```

> **Standard two-step workflow** (see `CLAUDE.md` - "Alfred Command Execution Pattern" for details)

## 📋 Your Task

You are executing the `/alfred:1-plan` command. Your job is to analyze the user's request and create a SPEC document following the EARS (Event-Action-Response-State) structure.

The command has **THREE execution phases**:

1. **PHASE 1**: Project Analysis & SPEC Planning (STEP 1)
2. **PHASE 2**: SPEC Document Creation (STEP 2)
3. **PHASE 3**: Git Branch & PR Setup (STEP 2 continuation)

Each phase contains explicit step-by-step instructions.

---

## 🔍 PHASE 1: Project Analysis & SPEC Planning (STEP 1)

PHASE 1 consists of **two independent sub-phases** to provide flexible workflow based on user request clarity:

### 📋 PHASE 1 Workflow Overview

```
┌─────────────────────────────────────────────────────────────┐
│ PHASE 1: Project Analysis & SPEC Planning                  │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Phase A (OPTIONAL)                                         │
│  ┌─────────────────────────────────────────┐               │
│  │ 🔍 Explore Agent                        │               │
│  │ • Find relevant files by keywords       │               │
│  │ • Locate existing SPEC documents        │               │
│  │ • Identify implementation patterns      │               │
│  └─────────────────────────────────────────┘               │
│                    ↓                                        │
│          (exploration results)                              │
│                    ↓                                        │
│  Phase B (REQUIRED)                                         │
│  ┌─────────────────────────────────────────┐               │
│  │ ⚙️ spec-builder Agent                   │               │
│  │ • Analyze project documents             │               │
│  │ • Propose SPEC candidates               │               │
│  │ • Design EARS structure                 │               │
│  │ • Request user approval                 │               │
│  └─────────────────────────────────────────┘               │
│                    ↓                                        │
│          (user approval via AskUserQuestion)                │
│                    ↓                                        │
│              PROCEED TO PHASE 2                             │
└─────────────────────────────────────────────────────────────┘
```

**Key Points**:
- **Phase A is optional** - Skip if user provides clear SPEC title
- **Phase B is required** - Always runs to analyze project and create SPEC
- **Results flow forward** - Exploration results (if any) are passed to spec-builder

---

### 🔍 Phase A: Codebase Exploration (OPTIONAL)

**When to execute Phase A:**

You should execute Phase A ONLY IF the user's request meets one of these criteria:

- ✅ User uses vague keywords ("where is...", "find me...", "related to...")
- ✅ Need to understand existing code structure before planning
- ✅ Feature spans multiple files or modules
- ❌ User provides clear SPEC title (skip to Phase B immediately)

**Decision rule**: If user typed a clear SPEC title (e.g., "JWT authentication system"), skip Phase A and proceed directly to Phase B.

#### How to execute Phase A:

**Step 1**: Determine IF you need exploration

1. Read the user's `$ARGUMENTS` input
2. Check if the input contains vague keywords:
   - "where is..."
   - "find me..."
   - "related to..."
   - "somewhere..."
   - "I think there's..."
3. IF the input is vague → proceed to Step 2
4. IF the input is clear → skip to Phase B

**Step 2**: Invoke the Explore agent

Use the Task tool to call the Explore agent:

```
Tool: Task
Parameters:
- subagent_type: "Explore"
- description: "Explore related files in the codebase"
- prompt: "다음 키워드와 관련된 모든 파일을 찾아주세요: $ARGUMENTS
  - 파일 위치 (src/, tests/, docs/)
  - 관련 SPEC 문서 (.moai/specs/)
  - 기존 구현 코드
  상세도 수준: medium"
```

**Step 3**: Wait for exploration results

1. The Explore agent will search the codebase
2. It will return a list of relevant files and locations
3. Store these results in a variable called `$EXPLORE_RESULTS`
4. Proceed to Phase B with this variable

---

### ⚙️ Phase B: SPEC Planning (REQUIRED)

**This phase ALWAYS runs** regardless of whether Phase A was executed.

Your task is to analyze the project documents and propose SPEC candidates to the user.

#### Step 1: Invoke the spec-builder agent

Use the Task tool to call the spec-builder agent:

```
Tool: Task
Parameters:
- subagent_type: "spec-builder"
- description: "Analyze the plan and establish a plan"
- prompt: """당신은 spec-builder 에이전트입니다.

언어 설정:
- 대화_언어: {{CONVERSATION_LANGUAGE}}
- 언어명: {{CONVERSATION_LANGUAGE_NAME}}

중요 지시사항:
SPEC 문서는 이중 언어 구조를 따라야 합니다 (사용자 언어 + 영어 요약):

conversation_language == 'ko' (한국어)인 경우:
- YAML 메타데이터: 영어만 사용
- 제목 (@SPEC 태그): 한국어 주요, 영어 버전은 하단에 기재
- 주요 내용 (분석, 요구사항, EARS): 한국어
- SUMMARY 섹션: 영어 (국제 기여자를 위해 100-200단어)
- HISTORY: 한국어 (새로운 항목), 주요 버전에는 영어 요약

conversation_language == 'ja' (일본어)인 경우:
- 한국어와 동일한 이중 언어 패턴 사용
- 주요 내용: 일본어
- SUMMARY: 영어

다른 언어인 경우:
- 주요 내용: 사용자 지정 언어
- SUMMARY: 영어 (항상)

스킬 호출:
필요 시 명시적 Skill() 호출 사용:
- Skill("moai-foundation-specs") - SPEC 구조 가이드
- Skill("moai-foundation-ears") - EARS 문법 요구사항
- Skill("moai-alfred-spec-metadata-validation") - 메타데이터 검증

작업:
프로젝트 문서를 분석하여 SPEC 후보자를 제시해주세요.
분석 모드로 실행하며, 다음을 포함해야 합니다:
1. product/structure/tech.md의 심층 분석
2. SPEC 후보자 식별 및 우선순위 결정
3. EARS 구조 설계
4. 사용자 승인 대기

사용자 입력: $ARGUMENTS
(선택사항) 탐색 결과: $EXPLORE_RESULTS"""
```

**Important**: IF Phase A was executed, include the `$EXPLORE_RESULTS` variable in the prompt. IF Phase A was skipped, omit the last line.

#### Step 2: Wait for spec-builder analysis

The spec-builder agent will:

1. **Read project documents**:
   - `.moai/project/product.md` (business requirements)
   - `.moai/project/structure.md` (architecture constraints)
   - `.moai/project/tech.md` (technical stack and policies)

2. **Scan existing SPECs**:
   - List all directories in `.moai/specs/`
   - Check for existing SPEC IDs to prevent duplicates
   - Identify current priorities and gaps

3. **Evaluate feasibility**:
   - Implementation complexity
   - Dependencies on other SPECs
   - Technical constraints
   - Resource requirements

4. **Propose SPEC candidates**:
   - Extract core business requirements
   - Reflect technical constraints
   - Create prioritized list of SPEC candidates

5. **Present implementation plan report**:

The spec-builder will generate a report in this format:

```
## Plan Creation Plan Report: [TARGET]

### Analysis Results
- **Discovered SPEC Candidates**: [Number and Category]
- **High Priority**: [List of Core SPECs]
- **Estimated Work Time**: [Time Estimation]

### Writing Strategy
- **Selected SPEC**: [SPEC ID and Title to Write]
- **EARS Structure**: [Event-Action-Response-State Design]
- **Acceptance Criteria**: [Given-When-Then Scenario]

### Technology stack and library versions (optional)
**Included only if technology stack is determined during planning stage**:
- **Web search**: Use `WebSearch` to find the latest stable versions of key libraries to use
- **Specify versions**: Specify exact versions for each library, e.g. `fastapi>=0.118.3`)
- **Stability priority**: Exclude beta/alpha versions, select only production stable versions
- **Note**: Detailed version is finalized in `/alfred:2-run` stage

### Precautions
- **Technical constraints**: [Restraints to consider]
- **Dependency**: [Relevance with other SPECs]
- **Branch strategy**: [Processing by Personal/Team mode]

### Expected deliverables
- **spec.md**: [Core specifications of the EARS structure]
- **plan.md**: [Implementation plan]
- **acceptance.md**: [Acceptance criteria]
- **Branches/PR**: [Git operations by mode]
```

#### Step 3: Request user approval

After the spec-builder presents the implementation plan report, you MUST ask the user for explicit approval before proceeding to PHASE 2.

**Ask the user this question**:

"Plan development is complete. Would you like to proceed with SPEC creation based on this plan?"

**Present these options**:

1. **Proceed with SPEC Creation** - Create SPEC files in `.moai/specs/SPEC-{ID}/` based on approved plan
2. **Request Modifications** - Specify changes to the plan before SPEC creation
3. **Save as Draft** - Save plan as draft without creating SPEC files yet
4. **Cancel** - Discard plan and return to planning phase

**Wait for the user to answer**.

#### Step 4: Process user's answer

Based on the user's choice:

**IF user selected "Proceed with SPEC Creation"**:
1. Store approval confirmation
2. Proceed to PHASE 2 (SPEC Document Creation)

**IF user selected "Request Modifications"**:
1. Ask the user: "What changes would you like to make to the plan?"
2. Wait for user's feedback
3. Pass feedback to spec-builder agent
4. spec-builder updates the plan
5. Return to Step 3 (request approval again with updated plan)

**IF user selected "Save as Draft"**:
1. Create directory: `.moai/specs/SPEC-{ID}/`
2. Save plan to `.moai/specs/SPEC-{ID}/plan.md` with status: draft
3. Create commit: `draft(spec): WIP SPEC-{ID} - {title}`
4. Print to user: "Draft saved. Resume with: `/alfred:1-plan resume SPEC-{ID}`"
5. End command execution (stop here)

**IF user selected "Cancel"**:
1. Print to user: "Plan discarded. No files created."
2. End command execution (stop here)

---

## 🚀 PHASE 2: SPEC Document Creation (STEP 2 - After Approval)

This phase ONLY executes IF the user selected "Proceed with SPEC Creation" in Phase B Step 4.

Your task is to create the SPEC document files in the correct directory structure.

### ⚠️ Critical Rule: Directory Naming Convention

**Format that MUST be followed**: `.moai/specs/SPEC-{ID}/`

**Correct Examples**:
- ✅ `SPEC-AUTH-001/`
- ✅ `SPEC-REFACTOR-001/`
- ✅ `SPEC-UPDATE-REFACTOR-001/`

**Incorrect examples**:
- ❌ `AUTH-001/` (missing SPEC- prefix)
- ❌ `SPEC-001-auth/` (additional text after ID)
- ❌ `SPEC-AUTH-001-jwt/` (additional text after ID)

**Duplicate check required**: Before creating a new SPEC ID, search existing TAG IDs to prevent duplication:

```bash
rg "@SPEC:{ID}" -n .moai/specs/
```

**Composite Domain Rules**:
- ✅ Allow: `UPDATE-REFACTOR-001` (2 domains)
- ⚠️ Caution: `UPDATE-REFACTOR-FIX-001` (3+ domains, simplification recommended)

### Step 1: Invoke spec-builder for SPEC creation

Use the Task tool to call the spec-builder agent:

```
Tool: Task
Parameters:
- subagent_type: "spec-builder"
- description: "Create SPEC document"
- prompt: """당신은 spec-builder 에이전트입니다.

언어 설정:
- 대화_언어: {{CONVERSATION_LANGUAGE}}
- 언어명: {{CONVERSATION_LANGUAGE_NAME}}

중요 지시사항:
모든 SPEC 문서는 대화_언어로 작성되어야 합니다:
- spec.md: 전체 문서를 대화_언어로 작성
- plan.md: 전체 문서를 대화_언어로 작성
- acceptance.md: 전체 문서를 대화_언어로 작성

YAML 프론트매터와 @TAG 식별자는 반드시 영어로 유지합니다.
코드 예제와 기술 키워드는 혼합 가능 (코드는 영어, 설명은 사용자 언어).

스킬 호출:
필요 시 명시적 Skill() 호출 사용:
- Skill("moai-foundation-specs") - SPEC 구조 가이드
- Skill("moai-foundation-ears") - EARS 문법 요구사항
- Skill("moai-alfred-spec-metadata-validation") - 메타데이터 검증
- Skill("moai-alfred-tag-scanning") - TAG 체인 참조

작업:
STEP 1에서 승인된 계획에 따라 SPEC 문서를 작성해주세요.
EARS 구조에 대한 명세를 작성합니다."""
```

### Step 2: Wait for spec-builder to create files

The spec-builder agent will:

1. **Create directory structure**:
   ```bash
   mkdir -p .moai/specs/SPEC-{ID}/
   ```

2. **Create spec.md** (main SPEC document):

   **YAML Front Matter** (top of file):
   ```yaml
   ---
   id: AUTH-001
   version: 0.0.1
   status: draft
   created: 2025-09-15
   updated: 2025-09-15
   author: @Goos
   priority: high
   ---
   ```

   **Required fields**:
   - `id`: Same as TAG ID (`<domain>-<3 digits>`) - Never change after creation
   - `version`: v0.0.1 (INITIAL) → v0.1.0 (Implementation Completed) → v1.0.0 (Stable)
   - `status`: draft | in_progress | completed | deprecated
   - `created`: YYYY-MM-DD
   - `updated`: YYYY-MM-DD
   - `author`: GitHub @ prefix + ID (e.g. `@Goos`)
   - `priority`: critical | high | medium | low

   **Optional fields** (9 total):
   - `category`: Feature type
   - `labels`: Tags for grouping
   - `depends_on`: List of SPEC IDs this depends on
   - `blocks`: List of SPEC IDs blocked by this
   - `related_specs`: List of related SPEC IDs
   - `related_issue`: GitHub Issue number
   - `scope`: Implementation scope estimate

   **HISTORY section** (immediately after YAML):
   ```markdown
   # @SPEC:DOMAIN-NNN: [SPEC title]

   ## HISTORY

   ### v0.0.1 (2025-09-15)
   - **INITIAL**: Initial creation of [feature name] specification
   - **AUTHOR**: @Goos
   - **SCOPE**: [Brief scope description]
   - **CONTEXT**: [Why this SPEC was created]
   ```

   **EARS Requirements sections**:
   ```markdown
   ## Requirements

   ### Ubiquitous
   - The system must provide [feature]

   ### Event-driven (event-driven)
   - WHEN [condition], the system must [operate]

   ### State-driven
   - WHILE When in [state], the system must [operate]

   ### Optional (Optional function)
   - If WHERE [condition], the system can [operate]

   ### Unwanted Behaviors
   - IF [condition], the system must [respond appropriately with error handling or quality gates]
   ```

   **Traceability section**:
   ```markdown
   ## Traceability (@TAG)
   - **SPEC**: @SPEC:DOMAIN-NNN
   - **TEST**: tests/[domain]/test_[feature].py
   - **CODE**: src/[domain]/[feature].py
   - **DOC**: docs/[domain]/[feature].md
   ```

3. **Create plan.md** (implementation plan):
   - Step-by-step implementation roadmap
   - Technical approach
   - Resource requirements
   - Timeline estimates

4. **Create acceptance.md** (acceptance criteria):
   - Given-When-Then scenarios
   - Test cases
   - Success criteria

### Step 3: Verify SPEC files were created

After spec-builder completes:

1. Check that directory exists:
   ```bash
   ls -la .moai/specs/SPEC-{ID}/
   ```

2. Verify all 3 files exist:
   - `spec.md`
   - `plan.md`
   - `acceptance.md`

3. IF any file is missing:
   - Print error: "SPEC file creation failed: missing {filename}"
   - End command execution (stop here)

4. IF all files exist:
   - Proceed to PHASE 3

---

## 🚀 PHASE 3: Git Branch & PR Setup (STEP 2 continuation)

This phase ONLY executes IF PHASE 2 completed successfully and all SPEC files were created.

Your task is to create a Git branch and GitHub Pull Request for the SPEC.

### Step 1: Invoke git-manager agent

Use the Task tool to call the git-manager agent:

```
Tool: Task
Parameters:
- subagent_type: "git-manager"
- description: "Create Git branch/PR with duplicate prevention"
- prompt: """당신은 git-manager 에이전트입니다.

언어 설정:
- 대화_언어: {{CONVERSATION_LANGUAGE}}
- 언어명: {{CONVERSATION_LANGUAGE_NAME}}

중요 지시사항 (팀 모드 중복 방지):
GitHub Issue 또는 PR을 만들기 전에:
1. 항상 제목에 SPEC-ID가 있는 기존 Issue를 확인하세요
2. 항상 feature/SPEC-{ID} 브랜치명의 기존 PR을 확인하세요
3. Issue가 존재하면 → 업데이트, 중복 생성 금지
4. PR이 존재하면 → 업데이트, 중복 생성 금지
5. 둘 다 존재하면 → 최신 SPEC 버전으로 모두 업데이트
6. 레이블 필터 실패 시 대체 검색 사용 (일부 Issue는 레이블 없을 수 있음)
7. 항상 레이블 추가: "spec", "planning", + 우선순위 레이블

git-manager.md의 "SPEC 작성 시" 섹션에서 자세한 중복 방지 프로토콜과 코드 예제를 참고하세요.

작업:
완성된 SPEC 문서에 대해 기능 브랜치(feature/SPEC-{SPEC_ID})와 Draft PR(→ develop)을 생성합니다.
GitHub 엔티티를 생성하기 전에 중복 방지 프로토콜을 구현합니다.

출력 언어: {{CONVERSATION_LANGUAGE}}"""
```

### Step 2: Wait for git-manager to complete

The git-manager agent will:

1. **Check project mode** (Personal or Team):
   - Read `.moai/config.json`
   - Check `project.mode` field
   - IF mode == "Personal" → create local branch only
   - IF mode == "Team" → create branch + GitHub Issue + PR

2. **Create Git branch**:

   **Personal mode**:
   - Branch from `main` or `develop` (based on `git_strategy.personal.branch_from` in config)
   - Branch name: `feature/SPEC-{ID}`
   - Example: `git checkout -b feature/SPEC-AUTH-001`

   **Team mode** (CRITICAL - GitFlow enforcement):
   - **ALWAYS branch from `develop`** (GitFlow standard)
   - Branch name: `feature/SPEC-{ID}`
   - Example: `git checkout -b feature/SPEC-AUTH-001 develop`

3. **Create initial commit**:
   ```bash
   git add .moai/specs/SPEC-{ID}/
   git commit -m "spec(SPEC-{ID}): Initial SPEC creation

   🤖 Generated with Claude Code

   Co-Authored-By: 🎩 Alfred@MoAI"
   ```

4. **Push branch to remote** (Team mode only):
   ```bash
   git push -u origin feature/SPEC-{ID}
   ```

5. **Create GitHub Issue** (Team mode only):
   - Title: `[SPEC-{ID}] {SPEC title}`
   - Body: Summary of SPEC content
   - Labels: `spec`, `planning`, `{priority}`
   - Check for duplicates BEFORE creating
   - IF duplicate exists → update existing Issue

6. **Create Draft PR** (Team mode only):
   - Source: `feature/SPEC-{ID}`
   - Target: **ALWAYS `develop`** (GitFlow rule)
   - Title: `[SPEC-{ID}] {SPEC title}`
   - Body: Link to Issue + SPEC summary
   - Status: Draft (not ready for review)
   - Check for duplicates BEFORE creating
   - IF duplicate exists → update existing PR

### Step 3: Verify Git operations completed

After git-manager completes:

1. **Check branch was created**:
   ```bash
   git branch --list feature/SPEC-{ID}
   ```
   - IF branch exists → success
   - IF branch missing → print error and stop

2. **Personal mode verification**:
   - Check local commit exists
   - Print success message

3. **Team mode verification**:
   - Check remote branch exists: `git ls-remote origin feature/SPEC-{ID}`
   - Check GitHub Issue was created: `gh issue list --label spec`
   - Check Draft PR was created: `gh pr list --state open --head feature/SPEC-{ID}`
   - IF any verification fails → print error and stop

### Step 4: CodeRabbit SPEC Review (Local Only - Automatic)

**This step happens automatically in the background. You DO NOT need to execute anything.**

After Draft PR is created, CodeRabbit automatically triggers SPEC review:

**What CodeRabbit reviews**:
- ✅ YAML frontmatter validation (7 required fields)
- ✅ HISTORY section structure and completeness
- ✅ EARS requirements clarity (Ubiquitous/Event-driven/State-driven/Optional/Unwanted Behaviors)
- ✅ Acceptance criteria quality (Given-When-Then scenarios)
- ✅ @TAG system compliance (SPEC/TEST/CODE/DOC traceability)
- ✅ Documentation and formatting

**Expected timeline**: 1-2 minutes

**IF you are running in local environment**:
1. Print to user: "🤖 CodeRabbit is reviewing SPEC PR (1-2 minutes)..."
2. Print to user: "→ PR will be auto-approved if quality meets standards (80%+)"
3. Print to user: "→ Check `.coderabbit.yaml` for detailed review checklist"

**IF you are running in published package**:
1. Print to user: "✅ Draft PR created"
2. Print to user: "→ Manual review required (CodeRabbit not available)"

---

## ✅ Command Completion & Next Steps

After PHASE 3 completes successfully, you MUST ask the user what to do next.

### Ask the user this question:

"SPEC creation is complete. What would you like to do next?"

### Present these options:

1. **Start Implementation** - Proceed to `/alfred:2-run SPEC-XXX` for TDD implementation
2. **Review SPEC** - Review and modify SPEC documents before implementation
3. **New Session** - Execute `/clear` for better context management (recommended)
4. **Cancel** - Return to planning phase

### Wait for the user to answer

### Process user's answer:

**IF user selected "Start Implementation"**:
1. Print: "Starting TDD implementation workflow..."
2. Print: "You can execute: `/alfred:2-run SPEC-XXX`"
3. End command execution (user will manually run next command)

**IF user selected "Review SPEC"**:
1. Print: "📁 SPEC files created in `.moai/specs/SPEC-XXX/`"
2. Print: "Files: spec.md, plan.md, acceptance.md"
3. Print: "After review, run: `/alfred:2-run SPEC-XXX`"
4. End command execution

**IF user selected "New Session"**:
1. Print: "⏳ Clearing session for better context management..."
2. Print: "Note: This improves performance for large projects"
3. Print: "Next session: Run `/alfred:2-run SPEC-XXX`"
4. End command execution (user will manually run /clear)

**IF user selected "Cancel"**:
1. Print: "Returning to planning phase..."
2. Print: "SPEC files preserved for future use"
3. Print: "Create more SPECs with: `/alfred:1-plan`"
4. End command execution

---

## 📚 Reference Information

The following sections provide reference information for understanding SPEC structure and requirements. **You do not need to memorize these - they are available when needed.**

### EARS Specification Writing Guide

When creating SPEC requirements, follow the EARS (Event-Action-Response-State) structure:

1. **Event**: Define trigger events that occur in the system
2. **Action**: Specification of the system's action for an event
3. **Response**: Defining a response as a result of an action
4. **State**: Specifies system state changes and side effects

**Example**:
```markdown
### Ubiquitous Requirements
- The system must provide user authentication functionality

### Event-driven Requirements
- WHEN the user logs in with valid credentials, the system must issue a JWT token

### State-driven Requirements
- WHILE the token is in an unexpired state, the system must allow access to the protected resource

### Unwanted Behaviors
- IF the token has expired, the system must return a 401 Unauthorized response
```

For complete EARS syntax and examples, invoke: `Skill("moai-foundation-ears")`

### SPEC Metadata Standard

For complete metadata field descriptions, validation rules, and version system guide, invoke: `Skill("moai-alfred-spec-metadata-extended")`

**Quick reference**:
- **7 required fields**: id, version, status, created, updated, author, priority
- **9 optional fields**: category, labels, depends_on, blocks, related_specs, related_issue, scope

### Agent Role Separation

**spec-builder dedicated area**:
- Analysis of project documents and discovery of SPEC candidates
- Preparation of EARS structure specifications
- Preparation of Acceptance Criteria (Given-When-Then)
- Verification of SPEC document quality
- Application of @TAG system

**git-manager dedicated area**:
- Create and manage all Git branches
- Apply branch strategy for each mode (Personal: branch from main/develop, Team: ALWAYS branch from develop)
- Create GitHub Issue/PR (Team Mode: Create Draft PR `feature/SPEC-{ID}` → `develop`)
- Create initial commit and tags
- Handle remote synchronization

**Single Responsibility Principle**: spec-builder only writes plans, git-manager only performs Git/GitHub operations.

**Sequential execution**: Executes in the order spec-builder → git-manager to maintain clear dependencies.

**No inter-agent calls**: Each agent does NOT call other agents directly. They are executed sequentially only at the command level.

### Context Management Strategy

**Load first**: `.moai/project/product.md` (business requirement)

**Recommendation after completion**: The plan is complete. You can experience better performance and context management by starting a new chat session with the `/clear` or `/new` command before proceeding to the next step (`/alfred:2-run`).

For complete context engineering strategy, invoke: `Skill("moai-alfred-dev-guide")`

### Writing Tips

- Information that is not in the product/structure/tech document is supplemented by asking a new question
- Acceptance Criteria is encouraged to be written at least 2 times in 3 columns Given/When/Then
- The number of modules is reduced due to the relaxation of the Readable standard among the TRUST principles. If the recommended value (default 5) is exceeded, include justification in the SPEC `context` section

---

## 🎯 Summary: Your Execution Checklist

Before you consider this command complete, verify:

- [ ] **PHASE 1 executed**: spec-builder analyzed project and proposed SPEC candidates
- [ ] **User approval obtained**: User explicitly approved SPEC creation (via AskUserQuestion)
- [ ] **PHASE 2 executed**: spec-builder created all 3 SPEC files (spec.md, plan.md, acceptance.md)
- [ ] **Directory naming correct**: `.moai/specs/SPEC-{ID}/` format followed
- [ ] **YAML frontmatter valid**: All 7 required fields present
- [ ] **HISTORY section present**: Immediately after YAML frontmatter
- [ ] **EARS structure complete**: All 5 requirement types included
- [ ] **PHASE 3 executed**: git-manager created branch and PR (if Team mode)
- [ ] **Branch naming correct**: `feature/SPEC-{ID}` format
- [ ] **GitFlow enforced**: PR targets `develop` branch (not `main`)
- [ ] **Next steps presented**: User asked what to do next (via AskUserQuestion)

IF all checkboxes are checked → Command execution successful

IF any checkbox is unchecked → Identify missing step and complete it before ending

---

**End of command execution guide**
