---
name: moai:1-plan
description: "Define specifications and create development branch"
argument-hint: Title 1 Title 2 ... | SPEC-ID modifications
allowed-tools:
  - Task
  - AskUserQuestion
  - Skill
  - TodoWrite
model: sonnet
skills:
  - moai-core-issue-labels
  - moai-spec-intelligent-workflow
  - moai-alfred-ask-user-questions
---

## 📋 Pre-execution Context

!git status --porcelain
!git branch --show-current
!git log --oneline -10
!git diff --name-only HEAD
!find .moai/specs -name "*.md" -type f 2>/dev/null

## 📁 Essential Files

@.moai/config/config.json
@.moai/project/product.md
@.moai/project/structure.md
@.moai/project/tech.md
@.moai/specs/
@CLAUDE.md

---

# 🏗️ MoAI-ADK Step 1: Establish a plan (Plan) - Always make a plan first and then proceed.

> **Batched Design**: All AskUserQuestion calls follow batched design principles (1-4 questions per call) to minimize user interaction turns. See CLAUDE.md section "Alfred Command Completion Pattern" for details.

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

---

## The 4-Step Agent-Based Workflow Command Logic (v5.0.0)

This command implements the first 2 steps of Alfred's 4-step workflow:

1. **STEP 1**: Intent Understanding (Clarify user requirements)
2. **STEP 2**: Plan Creation (Create execution strategy with agent delegation)
3. **STEP 3**: Task Execution (Execute via tdd-implementer - NOT in this command)
4. **STEP 4**: Report & Commit (Documentation and git operations - NOT in this command)

**Command Scope**: Only executes Steps 1-2. Steps 3-4 are executed by `/moai:2-run` and `/moai:3-sync`.

---

## The Command Has THREE Execution Phases:

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
│  📊 Progress Report & User Confirmation                     │
│  • Display analysis results and plan summary                 │
│  • Show next steps and deliverables                         │
│  • Request final user approval                             │
│                    ↓                                        │
│              PROCEED TO PHASE 2                             │
└─────────────────────────────────────────────────────────────┘
```

**Key Points**:

- **Phase A is optional** - Skip if user provides clear SPEC title
- **Phase B is required** - Always runs to analyze project and create SPEC

---

### 📋 PHASE 1A: Project Exploration (Optional - if needed)

#### When to run Phase A:

- User provides only vague/unstructured request
- Need to find existing files and patterns
- Unclear about current project state

#### Step 1A.1: Invoke Explore Agent (Optional)

**Conditional Execution: Run Phase A ONLY if user request lacks clarity**

```
IF user_request_is_vague_or_needs_exploration:
    explore_result = Task(
        subagent_type="Explore",
        description="Explore project files and patterns related to: $ARGUMENTS",
        prompt="""You are the Explore agent.

Analyze the current project directory structure and relevant files based on the user request: "$ARGUMENTS"

Tasks:
1. Find relevant files by keywords from the user request
2. Locate existing SPEC documents (.moai/specs/*.md)
3. Identify implementation patterns and dependencies
4. Discover project configuration files
5. Analyze existing codebase structure

Report back:
- List of relevant files found
- Existing SPEC candidates discovered
- Implementation patterns identified
- Technical constraints and dependencies
- Recommendations for user clarification

Return comprehensive results to guide spec-builder agent.
"""
    )

    # Store agent ID for resume chain
    $EXPLORE_AGENT_ID = explore_result.metadata.agent_id

    # Log Phase 1A checkpoint
    Log to .moai/logs/phase-checkpoints.json:
      phase: "1A"
      agent_id: $EXPLORE_AGENT_ID
      status: "EXPLORATION_COMPLETE"
      timestamp: NOW()
ELSE:
    # User provided clear SPEC title - skip Phase A
    $EXPLORE_AGENT_ID = null

    # Log Phase 1A checkpoint (skipped)
    Log to .moai/logs/phase-checkpoints.json:
      phase: "1A"
      status: "SKIPPED"
      timestamp: NOW()

PROCEED TO PHASE 1B
```

**Decision Logic**: If user provided clear SPEC title (like "Add authentication module"), skip Phase A entirely and proceed directly to Phase B.

---

### 📋 PHASE 1B: SPEC Planning (Required)

#### Step 1B.1: Invoke spec-builder for project analysis (Resume from Phase 1A if applicable)

Use the Task tool to call the spec-builder agent with conditional resume:

```
# Phase 1B: SPEC Planning (Resume from Phase 1A if exploration was done)
planning_result = Task(
    subagent_type="spec-builder",
    resume="$EXPLORE_AGENT_ID",  # ⭐ Resume if Phase 1A executed, null if skipped
    description="Analyze project and create SPEC plan for: $ARGUMENTS",
    prompt="""You are the spec-builder agent.

IF $EXPLORE_AGENT_ID is set:
    You are continuing from project exploration in Phase 1A.
    The exploration results (files found, patterns identified, constraints) are automatically available via resume.
ELSE:
    Start fresh analysis based on user request: "$ARGUMENTS"

Language settings:
- conversation_language: {{CONVERSATION_LANGUAGE}}
- language_name: {{CONVERSATION_LANGUAGE_NAME}}

IMPORTANT INSTRUCTIONS:
CRITICAL LANGUAGE CONFIGURATION:
- You receive instructions in agent_prompt_language from config (default: English for global standard)
- You must respond in conversation_language from config (user's preferred language)
- Example: If agent_prompt_language="en" and conversation_language="ko", you receive English instructions but respond in Korean

SPEC DOCUMENT LANGUAGE RULES:
All SPEC documents content must be written in {{CONVERSATION_LANGUAGE}}:
- spec.md: Main content in {{CONVERSATION_LANGUAGE}}
- plan.md: Main content in {{CONVERSATION_LANGUAGE}}
- acceptance.md: Main content in {{CONVERSATION_LANGUAGE}}

ALWAYS ENGLISH (global standards):
- Skill names in invocations: Skill("skill-name")
- Code examples and technical keywords
- Technical terms and function names

SUPPORTED LANGUAGES (50+):
All MoAI-ADK supported languages including: en, ko, ja, es, fr, de, zh, ru, pt, it, ar, hi, th, vi, and many more.
Use conversation_language value directly without hardcoded language checks.

TASK:
Analyze the project based on user request: "{{USER_REQUEST}}"

### PHASE 1B.1: Project Analysis and SPEC Discovery

1. **Document Analysis**: Scan for existing documentation and patterns
   - Product document: Find relevant files
   - Structure document: Identify architectural patterns
   - Tech document: Discover technical constraints

2. **SPEC Candidate Generation**: Create 1-3 SPEC candidates
   - Analyze existing SPECs in `.moai/specs/` for duplicates
   - Check related GitHub issues via `Skill("moai-core-issue-labels")`
   - Generate unique SPEC candidates with proper naming

3. **EARS Structure Design**: For each SPEC candidate:
   - Define clear requirements using EARS grammar
   - Design acceptance criteria with Given/When/Then
   - Identify technical dependencies and constraints

### PHASE 1B.2: Implementation Plan Creation

For the selected SPEC candidate, create a comprehensive implementation plan:

**Technical Constraints & Dependencies:**
- Library versions: Use `WebSearch` to find latest stable versions
- Specify exact versions (e.g., `fastapi>=0.118.3`)
- Exclude beta/alpha versions, select only production stable versions
- Note: Detailed versions finalized in `/moai:2-run` stage

**Precautions:**
- Technical constraints: [Restraints to consider]
- Dependency: [Relevance with other SPECs]
- Branch strategy: [Processing by Personal/Team mode]

**Expected deliverables:**
- spec.md: [Core specifications of the EARS structure]
- plan.md: [Implementation plan]
- acceptance.md: [Acceptance criteria]
- Branches/PR: [Git operations by mode]
"""
)

# Store agent ID for resume chain
$PLANNING_AGENT_ID = planning_result.metadata.agent_id

# Log Phase 1B checkpoint
Log to .moai/logs/phase-checkpoints.json:
  phase: "1B"
  agent_id: $PLANNING_AGENT_ID
  resumed_from: $EXPLORE_AGENT_ID
  status: "PLANNING_COMPLETE"
  timestamp: NOW()
```

#### Step 1B.2: Request user approval

After the spec-builder presents the implementation plan report, use AskUserQuestion tool for explicit approval:

Tool: AskUserQuestion
Parameters:
questions:

- question: "Planning is complete. Would you like to proceed with SPEC creation based on this plan?"
  header: "SPEC Generation"
  multiSelect: false
  options:
  - label: "Proceed with SPEC Creation"
    description: "Create SPEC files in .moai/specs/SPEC-{ID}/ based on approved plan"
  - label: "Request Plan Modification"
    description: "Modify plan content before SPEC creation"
  - label: "Save as Draft"
    description: "Save plan as draft and continue later"
  - label: "Cancel"
    description: "Discard plan and return to planning stage"

**Wait for user response**, then proceed to Step 3.5.

#### Step 3.5: Progress Report and User Confirmation

**This step automatically executes after PHASE 1 completion.**

Display detailed progress report to user and get final approval:

```
📊 Progress Report for PHASE 1 Completion

✅ **Completed Items:**
- Project document analysis completed
- Existing SPEC scan completed
- SPEC candidate generation completed
- Technical constraint analysis completed

📋 **Plan Summary:**
- Selected SPEC: {SPEC ID} - {SPEC Title}
- Priority: {Priority}
- Estimated time: {Time Estimation}
- Main technology stack: {Technology Stack}

🎯 **Next Phase Plan (PHASE 2):**
- spec.md creation: Core specifications with EARS structure
- plan.md creation: Detailed implementation plan
- acceptance.md creation: Acceptance criteria and scenarios
- Directory: .moai/specs/SPEC-{ID}/

⚠️ **Important Notes:**
- Existing files may be overwritten
- Dependencies: {Dependencies}
- Resource requirements: {Resource Requirements}
```

Tool: AskUserQuestion
Parameters:
questions:

- question: "Plan completion and progress report\n\n**Analysis results:**\n- SPEC candidates found: [Number]\n- Priority: [Priority]\n- Estimated work time: [Time Estimation]\n\n**Next steps:**\n1. PHASE 2: SPEC file creation\n - .moai/specs/SPEC-{ID}/\n - spec.md, plan.md, acceptance.md creation\n\nProceed with the plan?"
  header: "Plan Confirmation"
  multiSelect: false
  options:
  - label: "Proceed"
    description: "Start SPEC creation according to plan"
  - label: "Detailed Revision"
    description: "Revise plan content then proceed"
  - label: "Save as Draft"
    description: "Save plan and continue later"
  - label: "Cancel"
    description: "Cancel operation and discard plan"

**Wait for user response**, then proceed to Step 4.

#### Step 4: Process user's answer

Based on the user's choice:

**IF user selected "Proceed"**:

1. Store approval confirmation
2. Print: "✅ Plan approved. Proceeding to PHASE 2."
3. Proceed to PHASE 2 (SPEC Document Creation)

**IF user selected "Detailed Revision"**:

1. Ask the user: "What changes would you like to make to the plan?"
2. Wait for user's feedback
3. Pass feedback to spec-builder agent
4. spec-builder updates the plan
5. Return to Step 3.5 (request approval again with updated plan)

**IF user selected "Save as Draft"**:

1. Create directory: `.moai/specs/SPEC-{ID}/`
2. Save plan to `.moai/specs/SPEC-{ID}/plan.md` with status: draft
3. Create commit: `draft(spec): WIP SPEC-{ID} - {title}`
4. Print to user: "Draft saved. Resume with: `/moai:1-plan resume SPEC-{ID}`"
5. End command execution (stop here)

**IF user selected "Cancel"**:

1. Print to user: "Plan discarded. No files created."
2. End command execution (stop here)

---

## 🚀 PHASE 2: SPEC Document Creation (STEP 2 - After Approval)

This phase ONLY executes IF the user selected "Proceed" in Step 3.5.

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

**Duplicate check required**: Verify SPEC ID uniqueness before creation

Search scope:

- Primary: .moai/specs/ directory

Return:

- exists: true/false
- locations: [] (if exists, list all conflicting file paths)
- recommendation: "safe to create" or "duplicate found - suggest different ID"

**Composite Domain Rules**:

- ✅ Allow: `UPDATE-REFACTOR-001` (2 domains)
- ⚠️ Caution: `UPDATE-REFACTOR-FIX-001` (3+ domains, simplification recommended)

### Step 1: Invoke spec-builder for SPEC creation (Resume from Phase 1B)

Use the Task tool to call the spec-builder agent with resume to maintain context:

```
# Phase 2: SPEC Document Creation (Resume from Phase 1B)
spec_result = Task(
    subagent_type="spec-builder",
    resume="$PLANNING_AGENT_ID",  # ⭐ Resume: Inherit full planning context
    description="Create SPEC document files for approved plan",
    prompt="""You are the spec-builder agent.

You are continuing from the SPEC planning phase in Phase 1B.

The full planning context (project analysis, SPEC candidates, implementation plan) is automatically available via resume.
Use this context to generate comprehensive SPEC document files.

You are the spec-builder agent.

Language settings:
- conversation_language: {{CONVERSATION_LANGUAGE}}
- language_name: {{CONVERSATION_LANGUAGE_NAME}}

IMPORTANT INSTRUCTIONS:
CRITICAL LANGUAGE CONFIGURATION:
- You receive instructions in agent_prompt_language from config (default: English for global standard)
- You must respond in conversation_language from config (user's preferred language)
- Example: If agent_prompt_language="en" and conversation_language="ko", you receive English instructions but respond in Korean

SPEC DOCUMENT LANGUAGE RULES:
All SPEC documents content must be written in {{CONVERSATION_LANGUAGE}}:
- spec.md: Main content in {{CONVERSATION_LANGUAGE}}
- plan.md: Main content in {{CONVERSATION_LANGUAGE}}
- acceptance.md: Main content in {{CONVERSATION_LANGUAGE}}

ALWAYS ENGLISH (global standards):
- Skill names in invocations: Skill("skill-name")
- Code examples and technical keywords
- Technical terms and function names

SUPPORTED LANGUAGES (50+):
All MoAI-ADK supported languages including: en, ko, ja, es, fr, de, zh, ru, pt, it, ar, hi, th, vi, and many more.
Use conversation_language value directly without hardcoded language checks.

TASK:
Create SPEC-{SPEC_ID} with the following requirements:

### CRITICAL: SPEC File Generation Rules (MANDATORY)

⚠️ **YOU MUST FOLLOW THESE RULES EXACTLY OR QUALITY GATE WILL FAIL:**

1. **NEVER create single .md file**: ❌ WRONG: .moai/specs/SPEC-AUTH-001.md
2. **ALWAYS create folder structure**: ✅ CORRECT: .moai/specs/SPEC-AUTH-001/ (directory)
3. **ALWAYS use Bash + MultiEdit combo**: ❌ WRONG: Write to .moai/specs/SPEC-{ID}/spec.md separately
4. **ALWAYS verify before creation**: Check directory name format and ID duplicates

### SPEC Document Creation (Step-by-Step)

**Step 1: Verify SPEC ID Format**
- Format: SPEC-{DOMAIN}-{NUMBER}
- Examples: ✅ SPEC-AUTH-001, SPEC-REFACTOR-001, SPEC-UPDATE-REFACTOR-001
- Wrong: ❌ AUTH-001, SPEC-001-auth, SPEC-AUTH-001-jwt

**Step 2: Verify ID Uniqueness**
- Search .moai/specs/ for existing SPEC files
- If duplicate ID found → Change ID or update existing SPEC
- If ID is unique → Proceed to Step 3

**Step 3: Create Directory Structure**
- Use Bash tool: mkdir -p /Users/goos/MoAI/MoAI-ADK/.moai/specs/SPEC-{SPEC_ID}/
- Wait for directory creation to complete
- Proceed to Step 4 ONLY AFTER directory exists

**Step 4: Generate 3 SPEC Files (SIMULTANEOUS - Required)**
- Use MultiEdit tool to create all 3 files at once
- DO NOT use Write tool for individual files
- Create files:
  * .moai/specs/SPEC-{SPEC_ID}/spec.md
  * .moai/specs/SPEC-{SPEC_ID}/plan.md
  * .moai/specs/SPEC-{SPEC_ID}/acceptance.md

### spec.md Requirements:
- YAML frontmatter with all 7 required fields:
  * id: SPEC-{SPEC_ID}
  * version: "1.0.0"
  * status: "draft"
  * created: "{{YYYY-MM-DD}}"
  * updated: "{{YYYY-MM-DD}}"
  * author: "{{AUTHOR_NAME}}"
  * priority: "{{HIGH|MEDIUM|LOW}}"
- HISTORY section immediately after frontmatter
- Complete EARS structure with all 5 requirement types:
  * Functional Requirements (MUST)
  * Non-Functional Requirements (SHOULD)
  * Interface Requirements (SHALL)
  * Design Constraints (MUST)
  * Acceptance Criteria (GIVEN/WHEN/THEN format)

### plan.md Requirements:
- Implementation plan with detailed steps
- Task decomposition and dependencies
- Resource requirements and timeline
- Technology stack specifications
- Risk analysis and mitigation strategies

### acceptance.md Requirements:
- Minimum 2 Given/When/Then test scenarios
- Edge case testing scenarios
- Success criteria and validation methods
- Performance/quality gate criteria

### Quality Assurance:
- Information not in product/structure/tech document supplemented by asking new questions
- Acceptance Criteria written at least 2 times in Given/When/Then format
- Number of requirement modules ≤ 5 (if exceeded, include justification in SPEC)

### Git Integration:
- Generate commit messages following conventional commits
- Create appropriate branch names based on git strategy
- Include SPEC identifiers in commit messages
"""
```

---

## 🚀 PHASE 3: Git Branch & PR Setup (STEP 2 continuation)

### ⚠️ CRITICAL: PHASE 3 Execution is Conditional on Config

**PHASE 3 executes ONLY IF**:

1. PHASE 2 completed successfully
2. `github.spec_git_workflow` is explicitly configured
3. Configuration permits branch creation

**PHASE 3 is SKIPPED IF**:
- `github.spec_git_workflow == "develop_direct"` (Direct commits, no branches)
- Configuration validation fails
- User permissions insufficient

---

### Step 1: Read and Validate Git Configuration (GitHub Flow 3-Mode)

**MANDATORY: Read configuration BEFORE any git operations**

Execute the following config validation (this is pseudo-code representing the actual decision logic):

```python
# Step 1A: Read configuration from .moai/config/config.json
config = read_json(".moai/config/config.json")
git_mode = config.get("git_strategy", {}).get("mode")  # "manual", "personal", or "team"
branch_creation = config.get("git_strategy", {}).get("branch_creation", {})
prompt_always = branch_creation.get("prompt_always", True)  # Default: true

# Step 1B: Validate mode value
valid_modes = ["manual", "personal", "team"]
if git_mode not in valid_modes:
    ERROR: f"Invalid git_strategy.mode: {git_mode}"
    ERROR: f"Must be one of: {valid_modes}"
    ABORT_GIT_OPERATIONS()

# Step 1C: Validate environment
environment = config.get("git_strategy", {}).get("environment")  # "local" or "github"
github_integration = config.get("git_strategy", {}).get("github_integration", False)

if git_mode == "manual" and environment != "local":
    WARN: "Manual mode should use local environment"

# Step 1D: Determine branch creation behavior
log(f"Git Config: mode={git_mode}, prompt_always={prompt_always}")

if prompt_always == True:
    # All modes ask user for branch creation
    ACTION = "ASK_USER_FOR_BRANCH_CREATION"
else:
    # Auto-create branch based on mode
    if git_mode in ["personal", "team"]:
        ACTION = "AUTO_CREATE_BRANCH"
    else:  # manual mode
        ACTION = "SKIP_BRANCH_CREATION"
```

**Visual**: Configuration validation checkpoint (GitHub Flow 3-Mode)
```
git_strategy.mode = "manual" ?
    ├─ prompt_always = true  → ASK USER: "Create branch?" (Yes/No)
    └─ prompt_always = false → SKIP branch creation (use main/current)

git_strategy.mode = "personal" ?
    ├─ prompt_always = true  → ASK USER: "Create branch?" (Yes/No)
    └─ prompt_always = false → AUTO create feature/SPEC-XXX

git_strategy.mode = "team" ?
    ├─ prompt_always = true  → ASK USER: "Create branch?" (Yes/No)
    └─ prompt_always = false → AUTO create feature/SPEC-XXX + GitHub Issue
```

---

### Step 2: Branch Creation Logic (All 3 Modes)

**All modes use common `branch_creation.prompt_always` configuration**

#### Step 2.1: Determine Branch Creation Behavior

Based on config `git_strategy.branch_creation.prompt_always`:

```python
# Step 2.1: Read branch creation configuration
prompt_always = config.get("git_strategy", {})
                       .get("branch_creation", {})
                       .get("prompt_always", True)  # Default: true

if prompt_always == True:
    ACTION = "ASK_USER_FOR_BRANCH_CREATION"
elif prompt_always == False:
    if git_mode == "manual":
        ACTION = "SKIP_BRANCH_CREATION"
    else:  # personal or team
        ACTION = "AUTO_CREATE_BRANCH"
```

---

#### Step 2.2: Route A - Ask User (When `prompt_always: true`)

**CONDITION**: `branch_creation.prompt_always == true`

**ACTION**: Ask user for branch creation preference

```python
AskUserQuestion({
    "questions": [{
        "question": "브랜치를 생성하시겠습니까?",
        "header": "Branch Strategy",
        "multiSelect": false,
        "options": [
            {
                "label": "자동 생성",
                "description": "feature/SPEC-{SPEC_ID} 브랜치 자동 생성"
            },
            {
                "label": "현재 브랜치 사용",
                "description": "현재 브랜치에서 직접 작업"
            }
        ]
    }]
})

# Based on user choice:
if user_choice == "자동 생성":
    ROUTE = "CREATE_BRANCH"
else:
    ROUTE = "USE_CURRENT_BRANCH"
```

**Next Step**: Go to Step 2.3 or 2.4 based on user choice

---

#### Step 2.3: Create Feature Branch (After User Choice OR Auto-Creation)

**CONDITION**: User selected "자동 생성" OR (`prompt_always: false` AND git_mode in [personal, team])

**ACTION**: Invoke git-manager to create feature branch

```python
# Step 2.3: Create feature branch
Task(
    subagent_type="git-manager",
    description="Create feature branch for SPEC implementation",
    prompt="""You are the git-manager agent.

INSTRUCTION: Create feature branch for SPEC implementation.

MODE: {git_mode} (manual/personal/team)
BRANCH_CREATION: prompt_always = {prompt_always}

TASKS:
1. Create branch: `feature/SPEC-{SPEC_ID}-{description}`
2. Set tracking upstream if remote exists
3. Switch to new branch
4. Create initial commit (if appropriate for mode)

VALIDATION:
- Verify branch was created and checked out
- Verify current branch is feature/SPEC-{SPEC_ID}
- Return branch creation status

NOTE: PR creation is handled separately in /moai:2-run or /moai:3-sync (Team mode only)
"""
)
```

**Expected Outcome**:
```
✅ Feature branch created: feature/SPEC-{SPEC_ID}-description
✅ Current branch switched to feature branch
✅ Ready for implementation in /moai:2-run
```

---

#### Step 2.4: Skip Branch Creation (After User Choice OR Manual Mode)

**CONDITION**: User selected "현재 브랜치 사용" OR (`prompt_always: false` AND git_mode == manual)

**ACTION**: Skip branch creation, continue with current branch

```
✅ Branch creation skipped

Behavior:
- SPEC files created on current branch
- NO git-manager agent invoked
- Ready for /moai:2-run implementation
- Commits will be made directly to current branch during TDD cycle
```

---

#### Step 2.5: Team Mode - Create Draft PR (After Branch Creation)

**CONDITION**: `git_mode == "team"` AND branch was created (Step 2.3)

**ACTION**: Create draft PR for team review

```python
# Step 2.5: Create draft PR (Team mode only)
Task(
    subagent_type="git-manager",
    description="Create draft PR for SPEC (Team mode)",
    prompt="""You are the git-manager agent.

INSTRUCTION: Create draft pull request for SPEC implementation.

CRITICAL CONFIG: git_strategy.mode == "team"
→ Team mode REQUIRES draft PRs for review coordination

TASKS:
1. Create draft PR: feature/SPEC-{SPEC_ID} → main/develop branch
2. PR title: "feat(spec): Add SPEC-{SPEC_ID} [DRAFT]"
3. PR body: Include SPEC ID, description, and checklist
4. Add appropriate labels (spec, draft, etc.)
5. Assign reviewers from team config (if configured)
6. Set PR as DRAFT (do NOT auto-merge)

VALIDATION:
- Verify PR was created in draft status
- Return PR URL and status
"""
)
```

**Expected Outcome**:
```
✅ Feature branch: feature/SPEC-{SPEC_ID}
✅ Draft PR created for team review
✅ Ready for /moai:2-run implementation
```

---

### Step 3: Conditional Status Report

Display status based on configuration and execution result:

#### Case 1: Branch Creation Prompted (`prompt_always: true`) - User Selected "자동 생성"

```
📊 Phase 3 Status: Feature Branch Created (User Choice)

✅ **Configuration**: git_strategy.mode = "{git_mode}"
✅ **Branch Creation**: prompt_always = true → User chose "자동 생성"

✅ **Feature Branch Created**:
- Branch: `feature/SPEC-{SPEC_ID}`
- Current branch switched to feature branch
- Ready for implementation on isolated branch

{IF TEAM MODE:
✅ **Draft PR Created** (Team Mode):
- PR Title: "feat(spec): Add SPEC-{SPEC_ID} [DRAFT]"
- Target Branch: develop/main
- Status: DRAFT (awaiting review)
}

🎯 **Next Steps:**
1. 📝 Review SPEC in `.moai/specs/SPEC-{SPEC_ID}/`
2. 🔧 Execute `/moai:2-run SPEC-{SPEC_ID}` to begin implementation
3. 🌿 All commits will be made to feature branch
{IF TEAM MODE:
4. 👥 Share draft PR with team for early review (already created)
5. 💬 Team can comment during development
6. ✅ Finalize PR in `/moai:3-sync` when complete
:ELSE:
4. 🔄 Create PR in `/moai:3-sync` when implementation complete
}
```

---

#### Case 2: Branch Creation Prompted (`prompt_always: true`) - User Selected "현재 브랜치 사용"

```
📊 Phase 3 Status: Direct Commit Mode (User Choice)

✅ **Configuration**: git_strategy.mode = "{git_mode}"
✅ **Branch Creation**: prompt_always = true → User chose "현재 브랜치 사용"

✅ **No Branch Created**:
- SPEC files created on current branch
- Ready for direct implementation
- Commits will be made directly to current branch

🎯 **Next Steps:**
1. 📝 Review SPEC in `.moai/specs/SPEC-{SPEC_ID}/`
2. 🔧 Execute `/moai:2-run SPEC-{SPEC_ID}` to begin implementation
3. 💾 All commits will be made directly to current branch
4. 🧪 Follow TDD: RED → GREEN → REFACTOR cycles
```

---

#### Case 3: Branch Creation Auto-Skipped (Manual Mode + `prompt_always: false`)

```
📊 Phase 3 Status: Direct Commit Mode (Configuration)

✅ **Configuration**: git_strategy.mode = "manual"
✅ **Branch Creation**: prompt_always = false → Auto-skipped

✅ **No Branch Created** (Manual Mode Default):
- SPEC files created on current branch
- NO git-manager invoked (as configured)
- Ready for direct implementation
- Commits will be made directly to current branch

🎯 **Next Steps:**
1. 📝 Review SPEC in `.moai/specs/SPEC-{SPEC_ID}/`
2. 🔧 Execute `/moai:2-run SPEC-{SPEC_ID}` to begin implementation
3. 💾 Make commits directly to current branch
4. 🧪 Follow TDD: RED → GREEN → REFACTOR cycles
```

---

#### Case 4: Branch Creation Skipped with Auto-Enable Prompt (Personal/Team + `prompt_always: false` + `auto_enabled: false`)

```
📊 Phase 3 Status: Direct Commit Mode (Manual Default for Personal/Team)

✅ **Configuration**: git_strategy.mode = "{git_mode}" (personal or team)
✅ **Branch Creation**: prompt_always = false, auto_enabled = false → Manual Default

⚠️ **Branch Creation**: Not created yet (waiting for approval)
- SPEC files created on current branch
- Ready for implementation
- Commits will be made directly to current branch initially

💡 **Automation Approval Offered:**
─────────────────────────────────────────
Would you like to enable automatic branch creation for future SPEC creations?
(This will update your config.json)

🤖 Yes  → Set branch_creation.auto_enabled = true
        → Next SPEC will auto-create feature/SPEC-XXX branch

❌ No   → Keep manual mode
        → Continue working on current branch for this SPEC
        → No config changes made
─────────────────────────────────────────

🎯 **Next Steps:**
1. 📝 Review SPEC in `.moai/specs/SPEC-{SPEC_ID}/`
2. 🔧 Execute `/moai:2-run SPEC-{SPEC_ID}` to begin implementation
3. 💾 Make commits directly to current branch
4. 🧪 Follow TDD: RED → GREEN → REFACTOR cycles
5. 🔄 Create PR in `/moai:3-sync` when implementation complete
```

---

#### Case 5: Branch Creation Auto-Enabled (Personal/Team + `prompt_always: false` + `auto_enabled: true`)

```
📊 Phase 3 Status: Feature Branch Created (Auto-Enabled)

✅ **Configuration**: git_strategy.mode = "{git_mode}" (personal or team)
✅ **Branch Creation**: prompt_always = false, auto_enabled = true → Auto-enabled

✅ **Feature Branch Created**:
- Branch: `feature/SPEC-{SPEC_ID}`
- Current branch switched to feature branch
- Ready for implementation on isolated branch

{IF TEAM MODE:
✅ **Draft PR Created** (Team Mode):
- PR Title: "feat(spec): Add SPEC-{SPEC_ID} [DRAFT]"
- Target Branch: develop/main
- Status: DRAFT (awaiting review)
}

🎯 **Next Steps:**
1. 📝 Review SPEC in `.moai/specs/SPEC-{SPEC_ID}/`
2. 🔧 Execute `/moai:2-run SPEC-{SPEC_ID}` to begin implementation
3. 🌿 All commits will be made to feature branch
{IF TEAM MODE:
4. 👥 Share draft PR with team for early review
5. 💬 Team can comment on draft PR during development
6. ✅ Finalize PR in `/moai:3-sync` when complete
:ELSE:
4. 🔄 Create PR in `/moai:3-sync` when implementation complete
}
```

---

## 🎯 Summary: Your Execution Checklist

Before you consider this command complete, verify:

- [ ] **PHASE 1 executed**: spec-builder analyzed project and proposed SPEC candidates
- [ ] **Progress report displayed**: User shown detailed progress report with analysis results
- [ ] **User approval obtained**: User explicitly approved SPEC creation (via enhanced AskUserQuestion)
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

## **End of command execution guide**

## Final Step: Next Action Selection

After SPEC creation completes, use AskUserQuestion tool to guide user to next action:

```python
AskUserQuestion({
    "questions": [{
        "question": "SPEC document creation is complete. What would you like to do next?",
        "header": "Next Steps",
        "multiSelect": false,
        "options": [
            {
                "label": "Start Implementation",
                "description": "Execute /moai:2-run to begin TDD development"
            },
            {
                "label": "Modify Plan",
                "description": "Modify and enhance SPEC content"
            },
            {
                "label": "Add New Feature",
                "description": "Create additional SPEC document"
            }
        ]
    }]
})
```

**Important**:

- Use conversation language from config
- No emojis in any AskUserQuestion fields
- Always provide clear next step options

## ⚡️ EXECUTION DIRECTIVE

**You must NOW execute the command following the "The 4-Step Agent-Based Workflow Command Logic" described above.**

1. Start PHASE 1: Project Analysis & SPEC Planning immediately.
2. Call the `Task` tool with `subagent_type="spec-builder"` (or `Explore` as appropriate).
3. Do NOT just describe what you will do. DO IT.
