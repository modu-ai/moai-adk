---
name: moai:1-plan
description: "Define specifications and create development branch"
argument-hint: Title 1 Title 2 ... | SPEC-ID modifications
allowed-tools:
  - Task
  - AskUserQuestion
  - Skill
skills:
  - moai-core-issue-labels
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

#### Step 1A.1: Invoke Explore Agent

**If user request lacks clarity or context, use Task to call Explore Agent:**

```
Tool: Task
Parameters:
- subagent_type: "Explore"
- description: "Explore project files and patterns"
- prompt: """You are the Explore agent.

Analyze the current project directory structure and relevant files based on the user request: "{{USER_REQUEST}}"

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
```

**IF user provided clear SPEC title**: Skip Phase A entirely and proceed directly to Phase B.

---

### 📋 PHASE 1B: SPEC Planning (Required)

#### Step 1B.1: Invoke spec-builder for project analysis

Use the Task tool to call the spec-builder agent:

```
Tool: Task
Parameters:
- subagent_type: "spec-builder"
- description: "Analyze project and create SPEC plan"
- prompt: """You are the spec-builder agent.

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

### Step 1: Invoke spec-builder for SPEC creation

Use the Task tool to call the spec-builder agent:

```
Tool: Task
Parameters:
- subagent_type: "spec-builder"
- description: "Create SPEC document"
- prompt: """You are the spec-builder agent.

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

### SPEC Document Creation

1. **Create directory**: `.moai/specs/SPEC-{SPEC_ID}/`

2. **Generate spec.md**:
   - YAML frontmatter with all 7 required fields (id, version, status, created, updated, author, priority)
   - HISTORY section immediately after frontmatter
   - Complete EARS structure with all 5 requirement types:
     * Functional Requirements (MUST)
     * Non-Functional Requirements (SHOULD)
     * Interface Requirements (SHALL)
     * Design Constraints (MUST)
     * Acceptance Criteria (GIVEN/WHEN/THEN)

3. **Generate plan.md**:
   - Implementation plan with detailed steps
   - Task decomposition and dependencies
   - Resource requirements and timeline
   - Technology stack specifications

4. **Generate acceptance.md**:
   - Minimum 2 Given/When/Then scenarios
   - Edge case testing scenarios
   - Success criteria and validation methods

### Quality Assurance:
- Information not in product/structure/tech document supplemented by asking new questions
- Acceptance Criteria written at least 2 times in 3 columns Given/When/Then
- Number of modules reduced by Readable standard (default 5) - if exceeded, include justification in SPEC context section

### ID Integration:
- Follow SPEC ID lifecycle rules

### Git Integration:
- Generate commit messages following conventional commits
- Create appropriate branch names based on git strategy
- Include SPEC identifiers in commit messages
"""
```

---

## 🚀 PHASE 3: Git Branch & PR Setup (STEP 2 continuation)

PHASE 3 automatically executes IF:

1. PHASE 2 completed successfully
2. Git operations enabled in config
3. User has appropriate permissions

### Step 1: Check Git Configuration

Validate git strategy and configuration before proceeding:

```
Tool: Task
Parameters:
- subagent_type: "git-manager"
- description: "Check git configuration and strategy"
- prompt: """You are the git-manager agent.

Check git configuration and determine appropriate strategy:

1. Read `.moai/config/config.json` git_strategy section
2. Determine if Personal or Team mode
3. Validate git user configuration
4. Check current branch and status

Return:
- mode: "personal" or "team"
- current_branch: [current branch name]
- user_configured: true/false
- strategy: [git strategy configuration]
- recommended_actions: [list of actions needed]
"""
```

### Step 2: Create Branch (Personal Mode)

**IF git_strategy.personal mode**:

```
Tool: Task
Parameters:
- subagent_type: "git-manager"
- description: "Create feature branch"
- prompt: """You are the git-manager agent.

Create feature branch for SPEC implementation:

1. Create branch: `feature/SPEC-{SPEC_ID}`
2. Set tracking upstream if remote exists
3. Switch to new branch
4. Create initial commit with stub files

Use conventional commit format: "feat(spec): Add SPEC-{SPEC_ID} specification"
"""
```

### Step 3: Create Draft PR (Team Mode)

**IF git_strategy.team mode**:

```
Tool: Task
Parameters:
- subagent_type: "git-manager"
- description: "Create draft pull request"
- prompt: """You are the git-manager agent.

Create draft PR for SPEC implementation:

1. Create branch: `feature/SPEC-{SPEC_ID}`
2. Push to remote
3. Create draft PR targeting `develop` branch
4. Add appropriate labels and reviewers
5. Include SPEC ID in PR title: "feat(spec): Add SPEC-{SPEC_ID} [DRAFT]"

Set PR as draft and do not auto-merge.
"""
```

### Step 4: Final Status Report

After git operations complete:

```
📊 Phase 3 Complete - Git Setup Finished

✅ **Git Operations Completed:**
- Branch created: feature/SPEC-{SPEC_ID}
- PR created (Team mode) or branch ready (Personal mode)
- Initial commit with SPEC files
- SPEC tracking established

🎯 **Next Steps:**
1. 📝 **Review SPEC**: Check .moai/specs/SPEC-{SPEC_ID}/ files
2. 🔧 **Start Implementation**: Run `/moai:2-run SPEC-{SPEC_ID}`
3. 📋 **Monitor Progress**: Track implementation via TDD cycle
4. 🔄 **Team Collaboration**: Review/modify draft PR if in Team mode

💡 **Tips:**
- Use `/moai:2-run SPEC-{SPEC_ID}` to begin implementation
- Follow RED → GREEN → REFACTOR cycle for TDD
- Commit frequently with meaningful messages
- Review progress regularly
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
