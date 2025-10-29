---
name: alfred:1-plan
description: "Planning (brainstorming, plan writing, design discussion) + Branch/PR creation"
# Translations:
# - ko: "계획 수립 (브레인스토밍, 설계 논의) + 브랜치/PR 생성"
# - ja: "計画策定（ブレインストーミング、設計議論）+ ブランチ/PR作成"
# - zh: "规划（头脑风暴、设计讨论）+ 分支/PR创建"
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
> **Note**: Interactive prompts use `AskUserQuestion tool (documented in moai-alfred-interactive-questions skill)` for TUI selection menus. The skill is loaded on-demand when user interaction is required.
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

`/alfred:1-plan` is a general-purpose command that **creates a plan**, rather than simply “creating” a SPEC document.

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

## 📋 Execution flow

1. **Project Analysis**: In-depth analysis of product/structure/tech.md
2. **SPEC candidate discovery**: Prioritization based on business requirements
3. **User Verification**: Review and approve writing plan
4. **Plan creation**: Generate specifications of EARS structure (spec.md, plan.md, acceptance.md)
5. **Git operations**: Create branches/PRs via git-manager

## 🧠 Associated Skills & Agents

| Agent        | Core Skill                 | Purpose                     |
| ------------ | -------------------------- | --------------------------- |
| spec-builder | `moai-foundation-ears`     | Write SPEC with EARS syntax |
| git-manager  | `moai-alfred-git-workflow` | Create branch and PR        |

**Note**: TUI Survey Skill is used for user confirmations during the plan phase and is shared across all interactive prompts.

## 🔗 Associated Agent

- **Primary**: spec-builder (🏗️ System Architect) - Dedicated to writing SPEC documents
- **Secondary**: git-manager (🚀 Release Engineer) - Dedicated to creating Git branches/PRs

## 💡 Example of use

Users can run commands like this:
- `/alfred:1-plan` - Auto-suggestion based on project documents
- `/alfred:1-plan "JWT authentication system"` - Manually create a single SPEC
- `/alfred:1-plan SPEC-001 "Security hardening"` - Supplementation of existing SPEC

## 🔍 STEP 1: Project analysis and planning

STEP 1 consists of **two independent phases** to provide flexible workflow based on user request clarity:

### 📋 STEP 1 Workflow Overview

```
┌─────────────────────────────────────────────────────────────┐
│ STEP 1: Project Analysis & Planning                        │
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
│              PROCEED TO STEP 2                              │
└─────────────────────────────────────────────────────────────┘
```

**Key Points**:
- **Phase A is optional** - Skip if user provides clear SPEC title
- **Phase B is required** - Always runs to analyze project and create SPEC
- **Results flow forward** - Exploration results (if any) are passed to spec-builder

---

### 🔍 Phase A: Codebase Exploration (OPTIONAL)

**Use the Explore agent when user request is unclear or needs context.**

#### When to use Phase A:

- ✅ User uses vague keywords ("where is...", "find me...", "related to...")
- ✅ Need to understand existing code structure before planning
- ✅ Feature spans multiple files or modules
- ❌ User provides clear SPEC title (skip to Phase B)

#### How to invoke Explore agent:

```
Invoking the Task tool (Explore agent):
- subagent_type: "Explore"
- description: "Explore related files in the codebase"
- prompt: "Please find all files related to the following keywords: $ARGUMENTS
 - File location (src/, tests/, docs/)
 - Relevant SPEC document (.moai/specs/)
 - Existing implementation code
 thoroughness level: medium"
```

**Note**: If user provides clear SPEC title, skip Phase A and proceed directly to Phase B.

---

### ⚙️ Phase B: SPEC Planning (REQUIRED)

**Call the spec-builder agent to analyze project and create SPEC documents.**

This phase is **always required** regardless of whether Phase A was executed.

#### How to invoke spec-builder:

```
Call the Task tool:
- subagent_type: "spec-builder"
- description: "Analyze the plan and establish a plan"
- prompt: """You are spec-builder agent.

LANGUAGE CONFIGURATION:
- conversation_language: {{CONVERSATION_LANGUAGE}}
- language_name: {{CONVERSATION_LANGUAGE_NAME}}

CRITICAL INSTRUCTION:
All SPEC documents and analysis must be generated in conversation_language.
- If conversation_language is 'ko' (Korean): Generate ALL analysis, plans, and SPEC documents in Korean
- If conversation_language is 'ja' (Japanese): Generate ALL analysis, plans, and SPEC documents in Japanese
- If conversation_language is other language: Follow the specified language

SKILL INVOCATION:
Use explicit Skill() calls when needed:
- Skill("moai-foundation-specs") for SPEC structure guidance
- Skill("moai-foundation-ears") for EARS syntax requirements
- Skill("moai-alfred-spec-metadata-validation") for metadata validation

TASK:
Please analyze the project document and suggest SPEC candidates.
Run in analysis mode, and must include the following:
1. In-depth analysis of product/structure/tech.md
2. Identify SPEC candidates and Determine priorities
3. Design EARS structure
4. Wait for user approval

User input: $ARGUMENTS
(Optional) Explore results: $EXPLORE_RESULTS"""
```

**Note**: If Phase A was executed, pass the exploration results via `$EXPLORE_RESULTS` variable.

### Plan analysis progress

1. **Project document analysis**
 - In-depth analysis of product/structure/tech.md
 - Review existing SPEC list and priorities (.moai/specs/ scan)
 - Evaluate implementation feasibility and complexity
 - (Optional) Identify existing code structure by reflecting the Explore results

2. **Discovering SPEC candidates**
 - Extracting core business requirements
 - Reflecting technical constraints
 - Creating a list of SPEC candidates by priority

3. **Implementation plan report**
 - Present step-by-step plan creation plan
 - Estimated scope of work and dependency analysis
 - Design EARS structure and Acceptance Criteria

### User verification steps

After reviewing your implementation plan, Alfred invokes `AskUserQuestion tool (documented in moai-alfred-interactive-questions skill)` to present the following options:
- **"Go"** or **"Start"**: Start writing the plan as planned
- **"Modify [Content]"**: Request modifications to the plan
- **"Stop"**: Stop writing the plan

---

## 🚀 STEP 2: Create plan document (after user approval)

After user approval (collected via `AskUserQuestion tool (documented in moai-alfred-interactive-questions skill)`), call the spec-builder and git-manager agents using the **Task tool**.

### ⚙️ How to call an agent

```
1. Call spec-builder (create plan):
   - subagent_type: "spec-builder"
   - description: "Create SPEC document"
   - prompt: """You are spec-builder agent.

LANGUAGE CONFIGURATION:
- conversation_language: {{CONVERSATION_LANGUAGE}}
- language_name: {{CONVERSATION_LANGUAGE_NAME}}

CRITICAL INSTRUCTION:
ALL SPEC documents MUST be generated in conversation_language:
- spec.md: Full document in conversation_language
- plan.md: Full document in conversation_language
- acceptance.md: Full document in conversation_language

YAML frontmatter and @TAG identifiers MUST remain in English.
Code examples and technical keywords can be mixed (code in English, narrative in user language).

SKILL INVOCATION:
Use explicit Skill() calls when needed:
- Skill("moai-foundation-specs") for SPEC structure guidance
- Skill("moai-foundation-ears") for EARS syntax requirements
- Skill("moai-alfred-spec-metadata-validation") for metadata validation
- Skill("moai-alfred-tag-scanning") for TAG chain references

TASK:
Please fill out the SPEC document according to the plan approved in STEP 1.
Create a specification for the EARS structure."""

2. Invoke git-manager (Git task):
   - subagent_type: "git-manager"
   - description: "Create Git branch/PR"
   - prompt: "After completing the plan, please create a branch and Draft PR."
```

## function

- **Project document analysis**: Analyzes `.moai/project/{product,structure,tech}.md` to suggest implementation candidates and generates SPEC after user approval.
- **Personal mode**: Create a `.moai/specs/SPEC-{ID}/` directory and a template document (**Directory name format required**: `SPEC-` prefix + TAG ID).
- **Team mode**: Create a GitHub Issue (or Discussion) Associate it with a branch template.

## How to use

The user executes the command in the form:
- `/alfred:1-plan` - Auto-suggestion based on project documents (recommended)
- `/alfred:1-plan "JWT Authentication System"` - Manually create a single SPEC
- `/alfred:1-plan SPEC-001 "Security Reinforcement"` - Supplementation of existing SPEC

If not entered, 3 to 5 priorities will be suggested based on the Q&A results, and only the approved items will be confirmed as actual SPECs.

## Summary of processing by mode

| mode     | output                                                                     | Branch Strategy                                     | Additional Actions                                  |
| -------- | -------------------------------------------------------------------------- | --------------------------------------------------- | --------------------------------------------------- |
| Personal | Templates `.moai/specs/SPEC-XXX/spec.md`, `plan.md`, `acceptance.md`, etc. | Branch from `main` or `develop` (based on settings) | git-manager agent automatically creates checkpoints |
| Team     | GitHub Issue (`[SPEC-XXX] Title`), Draft PR (optional)                     | **Always branch from `develop`** (GitFlow standard) | `gh` CLI stay logged in, Draft PR → develop created |

## Input options

- **Automatic suggestion**: `/alfred:1-plan` → Create a list of candidates based on the core bullet of the project document
- **Manual creation**: Pass the title as an argument → Create only 1 case, Acceptance template is supplemented after reply
- **Supplementation mode**: `SPEC-ID Delivered in “memo” format → Update existing SPEC document/Issue

## 📋 STEP 1 Execution Guide: Project Analysis and Planning

### ⚠️ Essential rules: Directory naming convention

**Format that must be followed**: `.moai/specs/SPEC-{ID}/`

**Correct Example**:
- ✅ `SPEC-AUTH-001/`
- ✅ `SPEC-REFACTOR-001/`
- ✅ `SPEC-UPDATE-REFACTOR-001/`

**Incorrect example**:
- ❌ `AUTH-001/` (missing SPEC- prefix)
- ❌ `SPEC-001-auth/` (additional text after ID)
- ❌ `SPEC-AUTH-001-jwt/` (additional text after ID)

**Duplicate check required**: Before creating a new SPEC ID, be sure to search the existing TAG ID to prevent duplication.

**Composite Domain Rules**:
- ✅ Allow: `UPDATE-REFACTOR-001` (2 domains)
- ⚠️ Caution: `UPDATE-REFACTOR-FIX-001` (3+ domains, simplification recommended)

---

### 1. Analysis of project documents

Alfred calls the spec-builder agent to perform project document-based planning analysis and planning.

#### Analysis Checklist

- [ ] **Requirements extraction**: Identify key business requirements in product.md
- [ ] **Architectural constraints**: Identify system design constraints in structure.md
- [ ] **Technical constraints**: Technology stack and quality policy in tech.md
- [ ] **Existing SPEC**: Review current SPEC list and priorities

### 2. SPEC candidate discovery strategy

#### Prioritization criteria

| Priority   | standards                   | SPEC Candidate Type                         |
| ---------- | --------------------------- | ------------------------------------------- |
| **High**   | Core Business Values ​​     | User core functions, API design             |
| **Medium** | System Stability            | Authentication/Security, Data Management    |
| **Low**    | Improvements and expansions | UI/UX improvement, performance optimization |

#### Approach by SPEC type

- **API/Backend**: Endpoint design, data model, authentication
- **Frontend**: User interface, state management, routing
- **Infrastructure**: Deployment, monitoring, security policy
- **Quality**: Test strategy, performance criteria, documentation

### 3. Create a plan Create a plan report

Present your plan in the following format:

```
## Plan Creation Plan Report: [TARGET]

### 📊 Analysis Results
- **Discovered SPEC Candidates**: [Number and Category]
- **High Priority**: [List of Core SPECs]
- **Estimated Work Time**: [Time Estimation]

### 🎯 Writing Strategy
- **Selected SPEC**: [SPEC ID and Title to Write]
- **EARS Structure**: [Event-Action-Response-State Design]
- **Acceptance Criteria**: [Given-When-Then Scenario]

### 📦 Technology stack and library versions (optional)
**Included only if technology stack is determined during planning stage**:
- **Web search**: Use `WebSearch` to find the latest stable versions of key libraries to use
- **Specify versions**: Specify exact versions for each library, e.g. `fastapi>=0.118.3`)
- **Stability priority**: Exclude beta/alpha versions, select only production stable versions
- **Note**: Detailed version is finalized in `/alfred:2-run` stage

### ⚠️ Precautions
- **Technical constraints**: [Restraints to consider]
- **Dependency**: [Relevance with other SPECs]
- **Branch strategy**: [Processing by Personal/Team mode]

### ✅ Expected deliverables
- **spec.md**: [Core specifications of the EARS structure]
- **plan.md**: [Implementation plan]
- **acceptance.md**: [Acceptance criteria]
- **Branches/PR**: [Git operations by mode]

---
**Approval Request**: Would you like to proceed with creating a plan with the above plan?
 (Choose between “Proceed,” “Modify [Content],” or “Abort”)
```

---

## 🚀 STEP 2 Implementation Guide: Create a Plan (After Approval)

Only if the user selects **"Proceed"** or **"Start"** will Alfred call the spec-builder agent to begin building the SPEC document.

### EARS specification writing guide

1. **Event**: Define trigger events that occur in the system
2. **Action**: Specification of the system's action for an event
3. **Response**: Defining a response as a result of an action
4. **State**: Specifies system state changes and side effects

**Example** (see `development-guide.md` for details):
```markdown
### Ubiquitous Requirements
- The system must provide user authentication functionality

### Event-driven Requirements
- WHEN the user logs in with valid credentials, the system must issue a JWT token

### State-driven Requirements
- When the WHILE token is in an unexpired state, the system must allow access to the protected resource.

### Unwanted Behaviors
- If the IF token has expired, the system must return a 401 Unauthorized response.
```

### 📄 SPEC Document Template

#### YAML Front Matter Schema

> **📋 SPEC Metadata Standard (SSOT)**: `.moai/memory/spec-metadata.md`

**Metadata that must be included** at the top of the spec.md file:
- **7 required fields**: id, version, status, created, updated, author, priority
- **9 optional fields**: category, labels, depends_on, blocks, related_specs, related_issue, scope

**Simple reference example**:
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

**Core rules**:
- **id**: Same as TAG ID (`<domain>-<3 digits>`) - Never change after creation
 - **Directory name**: `.moai/specs/SPEC-{ID}/` (e.g. `SPEC-AUTH-001/`)
  - **Duplicate Check**: `rg "@SPEC:{ID}" -n .moai/specs/` Required
- **version**: v0.0.1 (INITIAL) → v0.1.0 (Implementation Completed) → v1.0.0 (Stable)
- **author**: GitHub @ prefix is required before ID (e.g. `@Goos`)
- **priority**: critical | high | medium | low

**Full field description and validation methods**: see `.moai/memory/spec-metadata.md`

#### HISTORY section (required)

You must include a HISTORY section **right after the YAML Front Matter**:

```markdown
# @SPEC:DOMAIN-NNN: JWT-based authentication system

## HISTORY

### v0.0.1 (2025-09-15)
- **INITIAL**: Initial creation of JWT-based authentication system specification
- **AUTHOR**: @Goos
- **SCOPE**: Token issuance, verification, and renewal logic
- **CONTEXT**: Reflects requirements for strengthening user authentication

### v0.0.2 (2025-09-20)
- **ADDED**: Added social login requirements (Draft modification)
- **AUTHOR**: @Goos
- **REVIEW**: @security-team (approved)
- **CHANGES**:
- OAuth2 integration requirements
 - Google/GitHub login support

### v0.1.0 (2025-10-01)
- **IMPLEMENTATION COMPLETED**: TDD implementation completed (status: draft → completed)
- **TDD CYCLE**: RED → GREEN → REFACTOR
- **COMMITS**: [Implementation commit hash list]
- **FILES**: [Created/modified file list]
```

**HISTORY writing rules**:
- **Version system**: v0.0.1 (INITIAL) → v0.1.0 (implementation complete) → v1.0.0 (stabilization)
 - Detailed version system: See `.moai/memory/spec-metadata.md#version-system`
- **Version order**: Latest version on top (reverse order)
- **Change type tag**: INITIAL, ADDED, CHANGED, IMPLEMENTATION COMPLETED, BREAKING, DEPRECATED, REMOVED, FIXED
 - Detailed description: See `.moai/memory/spec-metadata.md#history-writing-guide`
- **Required items**: Version, date, AUTHOR, changes
- **Optional items**: REVIEW, SCOPE, CONTEXT, MIGRATION

#### SPEC document overall structure

```markdown
---
id: AUTH-001
version: 1.0.0
status: draft
created: 2025-09-15
updated: 2025-09-15
author: @username
---

# @SPEC:DOMAIN-NNN: [SPEC title]

## HISTORY
[Change history by version – see example above]

## Environment
[System environment and prerequisites]

## Assumptions
[Design assumptions]

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

## Traceability (@TAG)
- **SPEC**: @SPEC:DOMAIN-NNN
- **TEST**: tests/auth/test_service.py
- **CODE**: src/auth/service.py
- **DOC**: docs/api/authentication.md
```

### Agent collaboration structure

- **Step 1**: The `spec-builder` agent is dedicated to analyzing project documents and creating SPEC documents.
- **Step 2**: The `git-manager` agent is dedicated to branch creation and GitHub Issue/PR creation.
- **Single Responsibility Principle**: spec-builder only writes plans, git-manager only performs Git/GitHub operations. 
- **Sequential execution**: Executes in the order spec-builder → git-manager to maintain clear dependencies.
- **No inter-agent calls**: Each agent calls the other agents. It is not called directly, but is executed sequentially only at the command level.

## 🚀 Optimized workflow execution order

### Phase 1: Parallel project analysis (performance optimization)

**Perform simultaneously**:

```
Task 1 (haiku): Scan project structure
├── Detect languages/frameworks
├── Collect list of existing SPECs
└── Draft priority backlog

Task 2 (sonnet): In-depth document analysis
├── product.md requirements extraction
├── structure.md architecture analysis
└── tech.md technical constraints
```

**Performance improvements**: Parallelize basic scans and deep analysis to minimize latency

### Phase 2: Create SPEC document integration

The `spec-builder` agent (sonnet) integrates the results of the parallel analysis:

- Proposal of function candidates based on project document
- Creation of SPEC document after user approval (using MultiEdit)
- Simultaneous creation of 3 files (spec.md, plan.md, acceptance.md)

### Phase 3: Git task processing

Final processing by the `git-manager` agent (haiku):

- **Branch creation**: Apply strategy for each mode
 - **Personal mode**: Branch from `main` or `develop` (based on project settings)
 - **Team mode**: **Always branch from `develop`** (GitFlow standard)
 - Branch name: `feature/SPEC-{ID}` format
- **Create GitHub Issue**: Create SPEC Issue in Team mode
- **Create Draft PR**: `feature/SPEC-{ID}` → `develop` in Team mode Create PR
- **Initial Commit**: Commit SPEC document and create tags

**Important**: Each agent runs independently, and direct calls between agents are prohibited.

## Agent role separation

### spec-builder dedicated area

- Analysis of project documents and discovery of SPEC candidates
- Preparation of EARS structure specifications
- Preparation of Acceptance Criteria (Given-When-Then)
- Verification of SPEC document quality
- Application of @TAG system

### git-manager dedicated area

- Create and manage all Git branches
- **Apply branch strategy for each mode**
 - Personal: Branch from `main` or `develop`
 - Team: **Always branch from `develop`** (GitFlow)
- Create GitHub Issue/PR
 - Team Mode: Create Draft PR (`feature/SPEC-{ID}` → `develop`)
- Create initial commit and tags
- Handle remote synchronization

## Step 2 workflow execution sequence

### Phase 1: Analysis and planning phase

**Plan Analyzer** does the following:

1. **Loading project document**: In-depth analysis of product/structure/tech.md
2. **SPEC candidate discovery**: Prioritization based on business requirements
3. **Establishment of implementation strategy**: EARS structure and acceptance design
4. **Creating a Writing Plan**: Presents a step-by-step approach to writing a plan
5. **Awaiting user approval**: Review plan and gather feedback

### Phase 2: Plan preparation phase (after approval)

The `spec-builder` agent **continuously** performs after user approval:

1. **Writing EARS specification**: Event-Action-Response-State structuring
2. **Acceptance Criteria**: Given-When-Then Scenario Writing
3. **Document quality verification**: Apply TRUST principles and @TAG
4. **Template creation**: Simultaneous creation of spec.md, plan.md, acceptance.md

### Phase 3: Git operations (git-manager)

The `git-manager` agent does **all at once** after the SPEC is complete:

1. **Create branch**: Apply branch strategy for each mode
2. **GitHub Issue**: Create SPEC Issue in Team mode
3. **Initial commit**: Commit SPEC document and create tags
4. **Remote Sync**: Apply synchronization strategy for each mode

### Phase 3.5: CodeRabbit SPEC Review (Local Only)

**After Draft PR is created, CodeRabbit automatically:**

```bash
echo "🤖 Waiting for CodeRabbit SPEC review..."

# CodeRabbit triggers automatically on Draft PR creation
# Review includes:
# - SPEC metadata validation (YAML frontmatter)
# - EARS structure completeness check
# - Acceptance criteria quality (Given-When-Then)
# - @TAG system traceability
# - Documentation clarity

# Expected time: 1-2 minutes
for i in {1..12}; do
    sleep 10

    # Check PR review status
    approval=$(gh pr view $pr_num --json reviewDecision --jq '.reviewDecision')

    if [ "$approval" = "APPROVED" ]; then
        echo "✅ CodeRabbit approved SPEC PR!"
        echo "→ Ready for development with /alfred:2-run SPEC-$spec_id"
        break
    fi

    echo "⏳ CodeRabbit reviewing... ($i/12)"
done
```

**CodeRabbit review includes:**
- ✅ YAML frontmatter validation (7 required fields)
- ✅ HISTORY section structure and completeness
- ✅ EARS requirements clarity (Ubiquitous/Event-driven/State-driven/Optional/Unwanted Behaviors)
- ✅ Acceptance criteria quality (Given-When-Then scenarios)
- ✅ @TAG system compliance (SPEC/TEST/CODE/DOC traceability)
- ✅ Documentation and formatting

See `.coderabbit.yaml` for detailed SPEC review checklist.

## Writing Tips

- Information that is not in the product/structure/tech document is supplemented by asking a new question. 
- Acceptance Criteria is encouraged to be written at least 2 times in 3 columns Given/When/Then. 
- The number of modules is reduced due to the relaxation of the Readable standard among the TRUST principles. If the recommended value (default 5) is exceeded, please include justification in the SPEC `context` section.

---

## 🧠 Context Management

> For more information: `.moai/memory/development-guide.md` - see section "Context Engineering"

### Core strategy of this command

**Load first**: `.moai/project/product.md` (business requirement)

**Recommendation**: The plan is complete. You can experience better performance and context management by starting a new chat session with the `/clear` or `/new` command before proceeding to the next step (`/alfred:2-run`).

---

## Next steps

**Recommendation**: For better performance and context management, start a new chat session with the `/clear` or `/new` command before proceeding to the next step.

- Start implementing TDD with `/alfred:2-run SPEC-XXX`
- Team mode: After creating an issue, the git-manager agent automatically creates a branch.
