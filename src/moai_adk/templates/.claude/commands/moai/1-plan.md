---
name: moai:1-plan
description: "Define specifications and create development branch or worktree"
argument-hint: Title 1 Title 2 ... | SPEC-ID modifications [--worktree | --branch]
allowed-tools: Task, AskUserQuestion, TodoWrite
model: inherit
---

##  Pre-execution Context

!git status --porcelain
!git branch --show-current
!git log --oneline -10
!git diff --name-only HEAD
!find .moai/specs -name "\*.md" -type f 2>/dev/null

##  Essential Files

@.moai/config/config.yaml
@.moai/project/product.md
@.moai/project/structure.md
@.moai/project/tech.md
.moai/specs/

---

#  MoAI-ADK Step 1: Establish a plan (Plan) - Always make a plan first and then proceed

User Interaction Architecture: AskUserQuestion must be used at COMMAND level only. Subagents invoked via Task() operate in isolated, stateless contexts and cannot interact with users. Collect all user input BEFORE delegating to agents.

Batched Design: All AskUserQuestion calls follow batched design principles (1-4 questions per call, max 4 options per question) to minimize user interaction turns. See CLAUDE.md section "User Interaction Architecture" for details.

4-Step Workflow Integration: This command implements Steps 1-2 of Alfred's workflow (Intent Understanding → Plan Creation). See CLAUDE.md for full workflow details.

##  Command Purpose

"Plan → Run → Sync" As the first step in the workflow, it supports the entire planning process from ideation to plan creation.

Plan for: $ARGUMENTS

###  Usage Scenarios (3 Execution Patterns)

Scenario 1: SPEC Only (Default)

```bash
/moai:1-plan "User authentication system"
```

- Creates SPEC documents only
- Follows existing branch creation logic

Scenario 2: SPEC + Branch (Legacy)

```bash
/moai:1-plan "User authentication system" --branch
```

- Creates SPEC documents + Git branch
- Traditional feature branch workflow

Scenario 3: SPEC + Worktree (NEW)

```bash
/moai:1-plan "User authentication system" --worktree
```

- Creates SPEC documents + Git worktree
- Isolated development environment for parallel SPEC work
- Displays guidance messages for worktree navigation

Flag Priority: `--worktree` > `--branch` > default (SPEC only)

##  CodeRabbit AI Integration (Local Only)

This local environment includes CodeRabbit AI review integration for SPEC documents:

Automatic workflows:

-  SPEC review: CodeRabbit analyzes SPEC metadata and EARS structure
-  GitHub Issue sync: SPEC files automatically create/update GitHub Issues
-  Auto-approval: Draft PRs are approved when quality meets standards (80%+)
-  SPEC quality validation: Checklist for metadata, structure, and content

Scope:

-  Local environment: Full CodeRabbit integration with auto-approval
-  Published packages: Users get GitHub Issue sync only (no CodeRabbit)

> See `.coderabbit.yaml` for detailed review rules and SPEC validation checklist

---

##  Associated Agents & Skills

Associated Agents for SPEC Planning and Creation:

- Explore: Codebase exploration and file system analysis
  WHY: Fast, focused discovery without blocking agent
  IMPACT: Reduces manual project discovery time

- manager-spec: SPEC generation in EARS format and planning
  WHY: Specialized domain knowledge for structured requirements
  IMPACT: Ensures consistent SPEC document quality

- manager-git: Git workflow and branch management
  WHY: Encapsulates git operations with proper error handling
  IMPACT: Prevents manual git errors and ensures consistency

### Agent Delegation Strategy

Phase 1A: Research & Analysis

- Use built-in Explore agent for fast codebase analysis (read-only)
- Use Plan agent (auto-invoked in plan mode) for SPEC research
- Use MoAI manager-spec agent for SPEC generation

Phase 1B: Specialized Analysis

- Use MoAI domain agents (expert-backend, expert-database, etc.) for specialized decisions
- Use mcp-context7 for API documentation research
- Use mcp-sequential-thinking for complex architectural decisions

---

## Agent Invocation Patterns (CLAUDE.md Compliance)

This command uses agent execution patterns defined in CLAUDE.md (lines 96-120).

### Sequential Phase-Based Chaining PASS

Command implements sequential chaining through 4 distinct phases:

Phase Flow:
- Phase 1A (Optional): Project Exploration via Explore subagent
- Phase 1B (Required): SPEC Planning via manager-spec subagent
- Phase 2: SPEC Document Creation via manager-spec subagent
- Phase 3: Git Branch Setup via manager-git subagent (conditional)

Each phase receives context and outputs from previous phases.

WHY: Sequential execution ensures proper dependency management
- Phase 1B needs exploration results from 1A (if applicable)
- Phase 2 requires approved plan from Phase 1B
- Phase 3 depends on created SPEC files from Phase 2

IMPACT: Skipping phases or parallel execution would violate dependencies and create incomplete specifications

### Parallel Execution FAIL

Not applicable - phases have explicit dependencies

WHY: Each phase depends on outputs from previous phase
- Cannot create SPEC documents before plan approval
- Cannot create git branch before SPEC files exist

IMPACT: Parallel execution would cause file system inconsistencies and incomplete workflows

### Resumable Agent Support PASS

Command supports resume pattern for draft SPECs:

Resume Command:
- `/moai:1-plan resume SPEC-XXX`
- Continues from last saved draft state
- Preserves user input and planning context

WHY: Complex planning sessions may require multiple iterations
IMPACT: Resume capability prevents loss of planning work and enables iterative refinement

---

Refer to CLAUDE.md "Agent Chaining Patterns" (lines 96-120) for complete pattern architecture.

---

##  Execution Philosophy: "Always make a plan first and then proceed."

`/moai:1-plan` performs SPEC planning through complete agent delegation:

```
User Command: /moai:1-plan "description"
    ↓
/moai:1-plan Command
    └─ Task(subagent_type="Explore" or "manager-spec")
        ├─ Phase 1A (Optional): Project Exploration
        ├─ Phase 1B (Required): SPEC Planning
        ├─ Phase 2: SPEC Document Creation
        └─ Phase 3: Git Branch & PR Setup
            ↓
        Output: SPEC documents + branch (conditional) + next steps
```

### Key Principle: Zero Direct Tool Usage

[HARD] Complete Delegation Model:

This command uses ONLY Task() and AskUserQuestion():
WHY: Specialized agents encapsulate domain logic and ensure quality control
IMPACT: Direct tool usage bypasses expert review and quality gates

[HARD] AskUserQuestion at Command Level Only:
WHY: Subagents via Task() are stateless and cannot interact with users
IMPACT: Expecting agents to use AskUserQuestion causes workflow failures
Correct Pattern: Command collects user input, passes choices to Task() as parameters

Requirement: All file operations delegated to agents
- No Read (file operations delegated)
  WHY: Agents handle context-aware file discovery
  IMPACT: Direct read loses architectural context

- No Write (file operations delegated)
  WHY: Agents implement validation before file creation
  IMPACT: Direct write skips quality checks

- No Edit (file operations delegated)
  WHY: Agents coordinate multi-file updates atomically
  IMPACT: Direct edit risks partial updates

- No Bash (all bash commands delegated)
  WHY: Agents handle error recovery and state consistency
  IMPACT: Direct bash loses error handling context

Required Tool Usage:
- Task() for agent orchestration
  WHY: Task ensures structured agent coordination
  IMPACT: Uncoordinated agent calls produce inconsistent results

- AskUserQuestion() for user interaction
  WHY: Structured questions ensure consistent user experience
  IMPACT: Informal interaction produces ambiguous responses

All complexity is handled by specialized agents, not by direct command execution.

---

## The 4-Step Agent-Based Workflow Command Logic (v5.0.0)

This command implements the first 2 steps of Alfred's 4-step workflow:

1. STEP 1: Intent Understanding (Clarify user requirements)
2. STEP 2: Plan Creation (Create execution strategy with agent delegation)
3. STEP 3: Task Execution (Execute via manager-tdd - NOT in this command)
4. STEP 4: Report & Commit (Documentation and git operations - NOT in this command)

Command Scope: Only executes Steps 1-2. Steps 3-4 are executed by `/moai:2-run` and `/moai:3-sync`.

---

## The Command Has THREE Execution Phases

1. PHASE 1: Project Analysis & SPEC Planning (STEP 1)
2. PHASE 2: SPEC Document Creation (STEP 2)
3. PHASE 3: Git Branch & PR Setup (STEP 2 continuation)

Each phase contains explicit step-by-step instructions.

---

##  PHASE 1: Project Analysis & SPEC Planning (STEP 1)

PHASE 1 consists of two independent sub-phases to provide flexible workflow based on user request clarity:

###  PHASE 1 Workflow Overview

```
┌─────────────────────────────────────────────────────────────┐
│ PHASE 1: Project Analysis & SPEC Planning                  │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Phase A (OPTIONAL)                                         │
│  ┌─────────────────────────────────────────┐               │
│  │  Explore Agent                        │               │
│  │ • Find relevant files by keywords       │               │
│  │ • Locate existing SPEC documents        │               │
│  │ • Identify implementation patterns      │               │
│  └─────────────────────────────────────────┘               │
│                    ↓                                        │
│          (exploration results)                              │
│                    ↓                                        │
│  Phase B (REQUIRED)                                         │
│  ┌─────────────────────────────────────────┐               │
│  │ ⚙ manager-spec Agent                   │               │
│  │ • Analyze project documents             │               │
│  │ • Propose SPEC candidates               │               │
│  │ • Design EARS structure                 │               │
│  │ • Request user approval                 │               │
│  └─────────────────────────────────────────┘               │
│                    ↓                                        │
│   Progress Report & User Confirmation                     │
│  • Display analysis results and plan summary                 │
│  • Show next steps and deliverables                         │
│  • Request final user approval                             │
│                    ↓                                        │
│              PROCEED TO PHASE 2                             │
└─────────────────────────────────────────────────────────────┘
```

Key Points:

- Phase A is optional - Skip if user provides clear SPEC title
- Phase B is required - Always runs to analyze project and create SPEC

---

###  PHASE 1A: Project Exploration (Optional - if needed)

#### When to run Phase A

This phase executes conditionally based on request clarity:

- [SOFT] User provides only vague/unstructured request
  WHY: Vague requests require exploration to identify relevant context
  IMPACT: Skipping exploration for vague requests produces unfocused SPECs

- [SOFT] Need to find existing files and patterns
  WHY: Discovery prevents duplicate work and informs architecture decisions
  IMPACT: Missing patterns leads to inconsistent implementation approaches

- [SOFT] Unclear about current project state
  WHY: Project context shapes technical constraints and dependencies
  IMPACT: Uninformed SPECs fail to account for existing architecture

#### Step 1A.1: Invoke Explore Agent (Optional)

Conditional Execution: Run Phase A ONLY if user request lacks clarity

If user request is vague or needs exploration:

Use the Explore subagent to:

Analyze the current project directory structure and relevant files based on the user request: "$ARGUMENTS"

Tasks:

1. Find relevant files by keywords from the user request
2. Locate existing SPEC documents (.moai/specs/\*.md)
3. Identify implementation patterns and dependencies
4. Discover project configuration files
5. Analyze existing codebase structure

Report back:

- List of relevant files found
- Existing SPEC candidates discovered
- Implementation patterns identified
- Technical constraints and dependencies
- Recommendations for user clarification

Return comprehensive results to guide manager-spec agent.

Phase 1A Completion:

- Log exploration completion status
- Proceed to Phase 1B with exploration context

Else (user provided clear SPEC title):

- Skip Phase A
- Log Phase 1A as skipped
- Proceed directly to Phase 1B

Decision Logic: If user provided clear SPEC title (like "Add authentication module"), skip Phase A entirely and proceed directly to Phase B.

---

###  PHASE 1B: SPEC Planning (Required)

#### Step 1B.1: Invoke manager-spec for project analysis

Use the manager-spec subagent to:

Analyze project and create SPEC plan for: $ARGUMENTS

Context Handling:

- If Phase 1A was executed: Continue from project exploration results
- If Phase 1A was skipped: Start fresh analysis based on user request: "$ARGUMENTS"

Language Configuration:

- conversation_language: {{CONVERSATION_LANGUAGE}}
- language_name: {{CONVERSATION_LANGUAGE_NAME}}

Critical Language Rules:

- Receive instructions in agent_prompt_language from config (default: English)
- Respond in conversation_language from config (user's preferred language)
- All SPEC documents content must be written in {{CONVERSATION_LANGUAGE}}
- Code examples and technical keywords remain in English (global standard)

Task Instructions:

PHASE 1B.1: Project Analysis and SPEC Discovery

1. Document Analysis: Scan for existing documentation and patterns

   - Product document: Find relevant files
   - Structure document: Identify architectural patterns
   - Tech document: Discover technical constraints

2. SPEC Candidate Generation: Create 1-3 SPEC candidates

   - Analyze existing SPECs in `.moai/specs/` for duplicates
   - Check related GitHub issues via appropriate tools
   - Generate unique SPEC candidates with proper naming

3. EARS Structure Design: For each SPEC candidate:
   - Define clear requirements using EARS grammar
   - Design acceptance criteria with Given/When/Then
   - Identify technical dependencies and constraints

PHASE 1B.2: Implementation Plan Creation
For the selected SPEC candidate, create a comprehensive implementation plan:

Technical Constraints & Dependencies:

- Library versions: Find latest stable versions
- Specify exact versions (e.g., `fastapi>=0.118.3`)
- Exclude beta/alpha versions, select only production stable versions
- Note: Detailed versions finalized in `/moai:2-run` stage

Precautions:

- Technical constraints: [Restraints to consider]
- Dependency: [Relevance with other SPECs]
- Branch strategy: [Processing by Personal/Team mode]

Expected deliverables:

- spec.md: [Core specifications of the EARS structure]
- plan.md: [Implementation plan]
- acceptance.md: [Acceptance criteria]
- Branches/PR: [Git operations by mode]

Phase 1B Completion:

- Log planning completion status
- Store context for subsequent phases

#### Step 1B.2: Request user approval

After the manager-spec presents the implementation plan report, use AskUserQuestion tool for explicit approval:

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

After AskUserQuestion returns user selection, proceed to Step 3.5.

#### Step 3.5: Progress Report and User Confirmation

This step automatically executes after PHASE 1 completion.

Display detailed progress report to user and get final approval:

```
 Progress Report for PHASE 1 Completion

 Completed Items:
- Project document analysis completed
- Existing SPEC scan completed
- SPEC candidate generation completed
- Technical constraint analysis completed

 Plan Summary:
- Selected SPEC: {SPEC ID} - {SPEC Title}
- Priority: {Priority}
- Estimated time: {Time Estimation}
- Main technology stack: {Technology Stack}

 Next Phase Plan (PHASE 2):
- spec.md creation: Core specifications with EARS structure
- plan.md creation: Detailed implementation plan
- acceptance.md creation: Acceptance criteria and scenarios
- Directory: .moai/specs/SPEC-{ID}/

 Important Notes:
- Existing files may be overwritten
- Dependencies: {Dependencies}
- Resource requirements: {Resource Requirements}
```

Tool: AskUserQuestion
Parameters:
questions:

- question: "Plan completion and progress report\n\nAnalysis results:\n- SPEC candidates found: [Number]\n- Priority: [Priority]\n- Estimated work time: [Time Estimation]\n\nNext steps:\n1. PHASE 2: SPEC file creation\n - .moai/specs/SPEC-{ID}/\n - spec.md, plan.md, acceptance.md creation\n\nProceed with the plan?"
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

Process user response from AskUserQuestion in Step 4.

#### Step 4: Process user's answer

Based on the user's choice:

IF user selected "Proceed":

1. Store approval confirmation
2. Print: " Plan approved. Proceeding to PHASE 2."
3. Proceed to PHASE 2 (SPEC Document Creation)

IF user selected "Detailed Revision":

1. Ask the user: "What changes would you like to make to the plan?"
2. Collect user's feedback via AskUserQuestion
3. Pass feedback to manager-spec agent
4. manager-spec updates the plan
5. Return to Step 3.5 (request approval again with updated plan)

IF user selected "Save as Draft":

1. Create directory: `.moai/specs/SPEC-{ID}/`
2. Save plan to `.moai/specs/SPEC-{ID}/plan.md` with status: draft
3. Create commit: `draft(spec): WIP SPEC-{ID} - {title}`
4. Print to user: "Draft saved. Resume with: `/moai:1-plan resume SPEC-{ID}`"
5. End command execution (stop here)

IF user selected "Cancel":

1. Print to user: "Plan discarded. No files created."
2. End command execution (stop here)

---

##  PHASE 2: SPEC Document Creation (STEP 2 - After Approval)

This phase ONLY executes IF the user selected "Proceed" in Step 3.5.

Your task is to create the SPEC document files in the correct directory structure.

###  Critical Rule: Directory Naming Convention

[HARD] SPEC Directory Structure Requirement:

Format requirement: `.moai/specs/SPEC-{ID}/`
WHY: Standardized directory structure enables automated discovery and tooling
IMPACT: Non-standard naming breaks automation and causes deployment failures

Correct Examples of Required Format:

- `SPEC-AUTH-001/` - Domain single, properly formatted
- `SPEC-REFACTOR-001/` - Domain single, properly formatted
- `SPEC-UPDATE-REFACTOR-001/` - Composite domains (2), properly formatted

Examples of Incorrect Formats to Avoid:

- `AUTH-001/` - Missing SPEC- prefix breaks automation
- `SPEC-001-auth/` - Additional text after ID violates convention
- `SPEC-AUTH-001-jwt/` - Additional text after ID violates convention

[HARD] ID Uniqueness Verification Requirement:

Verification scope: Search entire .moai/specs/ directory before creation
WHY: Duplicate IDs cause merge conflicts and implementation uncertainty
IMPACT: Duplicate IDs create version ambiguity and maintenance chaos

Verification output must include:

- exists: Boolean indicating whether ID already exists
- locations: Array of all conflicting file paths (empty if no conflicts)
- recommendation: Text assessment ("safe to create" or "duplicate found - use ID-XXX instead")

[SOFT] Composite Domain Naming Guidance:

- Allow: `UPDATE-REFACTOR-001` (2 domains maximum)
  WHY: Two domains indicate coordinated work without excessive complexity
  IMPACT: Two-domain SPECs maintain focus and scope

- Caution: `UPDATE-REFACTOR-FIX-001` (3+ domains not recommended)
  WHY: Three or more domains indicate scope creep and mixed concerns
  IMPACT: Complex domain names signal SPECs that should be split

### Step 1: Invoke manager-spec for SPEC creation

Use the manager-spec subagent to:

Create SPEC document files for approved plan

Context Continuity:

- Continue from the SPEC planning phase in Phase 1B
- Use full planning context (project analysis, SPEC candidates, implementation plan) to generate comprehensive SPEC document files

Language Configuration:

- conversation_language: {{CONVERSATION_LANGUAGE}}
- language_name: {{CONVERSATION_LANGUAGE_NAME}}

Critical Language Rules:

- Receive instructions in agent_prompt_language from config (default: English)
- Respond in conversation_language from config (user's preferred language)
- All SPEC documents content must be written in {{CONVERSATION_LANGUAGE}}
- Technical terms and function names remain in English (global standard)

SPEC File Generation Rules (MANDATORY):

[HARD] Create Directory Structure, Not Single Files:

Requirement: Always create folder structure for SPEC documents
WHY: Directory structure enables multi-file organization and metadata storage
IMPACT: Single .md files prevent future tool integration and violate structure assumptions

Correct approach: Create `.moai/specs/SPEC-AUTH-001/` as directory
Incorrect approach: Create `.moai/specs/SPEC-AUTH-001.md` as single file

[HARD] Verify Before Creation:

Requirement: Check directory name format and ID duplicates before writing files
WHY: Pre-flight verification prevents invalid states and merge conflicts
IMPACT: Writing invalid files creates cleanup burden and workflow disruption

[HARD] Quality Gate Compliance:

Requirement: Follow these rules exactly for quality gate to pass
WHY: Quality gates ensure system reliability and consistency
IMPACT: Non-compliance causes pipeline failures and deployment blocks

SPEC Document Creation (Step-by-Step):

Step 1: Verify SPEC ID Format

- Format: SPEC-{DOMAIN}-{NUMBER}
- Examples:  SPEC-AUTH-001, SPEC-REFACTOR-001, SPEC-UPDATE-REFACTOR-001
- Wrong:  AUTH-001, SPEC-001-auth, SPEC-AUTH-001-jwt

Step 2: Verify ID Uniqueness

- Search .moai/specs/ for existing SPEC files
- If duplicate ID found → Change ID or update existing SPEC
- If ID is unique → Proceed to Step 3

Step 3: Create Directory Structure

- Create directory: .moai/specs/SPEC-{SPEC_ID}/
- Directory creation completes synchronously before Step 4

Step 4: Generate 3 SPEC Files (SIMULTANEOUS - Required)

- Create all 3 files at once:
  - .moai/specs/SPEC-{SPEC_ID}/spec.md
  - .moai/specs/SPEC-{SPEC_ID}/plan.md
  - .moai/specs/SPEC-{SPEC_ID}/acceptance.md

File Requirements:

spec.md Requirements:

- YAML frontmatter with all 7 required fields:
  - id: SPEC-{SPEC_ID}
  - version: "1.0.0"
  - status: "draft"
  - created: "{{YYYY-MM-DD}}"
  - updated: "{{YYYY-MM-DD}}"
  - author: "{{AUTHOR_NAME}}"
  - priority: "{{HIGH|MEDIUM|LOW}}"
- HISTORY section immediately after frontmatter
- Complete EARS structure with all 5 requirement types

plan.md Requirements:

- Implementation plan with detailed steps
- Task decomposition and dependencies
- Resource requirements and timeline
- Technology stack specifications
- Risk analysis and mitigation strategies

acceptance.md Requirements:

- Minimum 2 Given/When/Then test scenarios
- Edge case testing scenarios
- Success criteria and validation methods
- Performance/quality gate criteria

Quality Assurance:

- Information not in product/structure/tech document supplemented by asking new questions
- Acceptance Criteria written at least 2 times in Given/When/Then format
- Number of requirement modules ≤ 5 (if exceeded, include justification in SPEC)

Git Integration:

- Generate commit messages following conventional commits
- Create appropriate branch names based on git strategy
- Include SPEC identifiers in commit messages

---

##  PHASE 3: Git Branch & PR Setup (STEP 2 continuation)

###  CRITICAL: PHASE 3 Execution is Conditional on Config AND Flags

PHASE 3 executes ONLY IF:

1. PHASE 2 completed successfully
2. One of these conditions is met:
   - `--worktree` flag is provided (NEW: Worktree creation)
   - `--branch` flag is provided OR user chose branch creation
   - Configuration permits branch creation
   - `github.spec_git_workflow` is explicitly configured

PHASE 3 Branch Logic:

- If `--worktree` flag: Skip traditional branch creation, create worktree instead
- If `--branch` flag: Follow traditional branch creation logic
- If no flags: Follow existing AskUserQuestion flow for branch choice

PHASE 3 is SKIPPED IF:

- `github.spec_git_workflow == "develop_direct"` (Direct commits, no branches)
- Configuration validation fails
- User permissions insufficient
- No branch/worktree creation flags provided AND user chooses "no branch"

---

### Step 1: Read and Validate Git Configuration

MANDATORY: Read configuration BEFORE any git operations

Execute the following config validation (this is pseudo-code representing the actual decision logic):

```python
# Step 1A: Read configuration from .moai/config/config.yaml
config = read_json(".moai/config/config.yaml")
git_mode = config.get("git_strategy", {}).get("mode")  # "personal" or "team"
spec_workflow = config.get("github", {}).get("spec_git_workflow")  # Required

# Step 1B: Validate spec_git_workflow value
valid_workflows = ["develop_direct", "feature_branch", "per_spec"]
if spec_workflow not in valid_workflows:
    ERROR: f"Invalid spec_git_workflow: {spec_workflow}"
    ERROR: f"Must be one of: {valid_workflows}"
    SKIP_PHASE_3 = True
    ABORT_GIT_OPERATIONS()

# Step 1C: Validate consistency
if git_mode == "personal" and spec_workflow == "develop_direct":
    CONSISTENCY_OK = True  #  Consistent
elif git_mode == "personal" and spec_workflow in ["feature_branch", "per_spec"]:
    WARN: "Personal mode with branch creation is non-standard but allowed"
    CONSISTENCY_OK = True
elif git_mode == "team" and spec_workflow in ["feature_branch", "per_spec"]:
    CONSISTENCY_OK = True
else:
    ERROR: "Inconsistent git configuration"
    ABORT_GIT_OPERATIONS()

# Step 1D: Determine PHASE 3 routing
log(f"Git Config: mode={git_mode}, spec_workflow={spec_workflow}")
```

Visual: Configuration validation checkpoint

```
git_mode = "personal" ?
    ├─ spec_workflow = "develop_direct"    → PHASE 3 SKIPPED (ROUTE A)
    ├─ spec_workflow = "feature_branch"    → PHASE 3 EXECUTES (ROUTE B)
    └─ spec_workflow = "per_spec"          → PHASE 3 WITH USER ASK (ROUTE C)
git_mode = "team" ?
    └─ (spec_workflow value ignored)       → PHASE 3 EXECUTES (ROUTE D - Team Mode)
```

---

### Step 2: Branch Creation Logic (All 3 Modes)

All modes use common `branch_creation.prompt_always` configuration

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

CONDITION: `branch_creation.prompt_always == true`

ACTION: Ask user for branch/worktree creation preference

**Step 1: Check auto_branch configuration**

Read configuration value from config.yaml:
- Path: git_strategy.automation.auto_branch
- Default: true

**Step 2: Early exit if auto_branch is disabled**

If auto_branch equals false:
- Set ROUTE to USE_CURRENT_BRANCH
- Skip to Step 2.4 immediately
- Do NOT ask user any questions

**Step 3: Ask user if auto_branch is enabled**

Use AskUserQuestion tool with the following parameters:
- Question: "Create a development environment for this SPEC?"
- Header: "Development Environment"
- MultiSelect: false
- Options:
  1. "Create Worktree" - Create isolated worktree environment (recommended for parallel SPEC development)
  2. "Create Branch" - Create feature/SPEC-{SPEC_ID} branch (traditional workflow)
  3. "Use current branch" - Work directly on current branch

**Step 4: Determine route based on user choice**

Based on user selection:
- If "Create Worktree" selected: Set ROUTE to CREATE_WORKTREE
- If "Create Branch" selected: Set ROUTE to CREATE_BRANCH
- If "Use current branch" selected: Set ROUTE to USE_CURRENT_BRANCH

Next Step: Go to Step 2.5 (worktree), 2.3 (branch), or 2.4 (current) based on route

---

#### Step 2.3: Create Feature Branch (After User Choice OR Auto-Creation)

CONDITION:
- User selected "Create Branch"
- OR (`prompt_always: false` AND git_mode in [personal, team])
- AND `git_strategy.automation.auto_branch == true`

If `auto_branch: false`: Skip to Step 2.4 (Use current branch)

ACTION: Invoke manager-git to create feature branch

Use the manager-git subagent to:

Create feature branch for SPEC implementation

Instructions:

- MODE: {git_mode} (manual/personal/team)
- BRANCH_CREATION: prompt_always = {prompt_always}

Tasks:

1. Create branch: `feature/SPEC-{SPEC_ID}-{description}`
2. Set tracking upstream if remote exists
3. Switch to new branch
4. Create initial commit (if appropriate for mode)

Validation:

- Verify branch was created and checked out
- Verify current branch is feature/SPEC-{SPEC_ID}
- Return branch creation status

Note: PR creation is handled separately in /moai:2-run or /moai:3-sync (Team mode only)

Expected Outcome:

```
 Feature branch created: feature/SPEC-{SPEC_ID}-description
 Current branch switched to feature branch
 Ready for implementation in /moai:2-run
```

---

#### Step 2.4: Skip Branch Creation (After User Choice OR Manual Mode)

CONDITION: User selected "Use current branch" OR (`prompt_always: false` AND git_mode == manual)

ACTION: Skip branch creation, continue with current branch

```
 Branch creation skipped

Behavior:
- SPEC files created on current branch
- NO manager-git agent invoked
- Ready for /moai:2-run implementation
- Commits will be made directly to current branch during TDD cycle
```

---

#### Step 2.5: Worktree Creation (NEW - When --worktree flag provided)

CONDITION: `--worktree` flag is provided in user command

ACTION: Create Git worktree using WorktreeManager

```python
# Step 2.5: Create worktree (when --worktree flag provided)
# Parse command arguments to check for --worktree flag
import sys
import shlex
from pathlib import Path

# Parse command arguments to detect --worktree flag
command_args = shlex.split("$ARGUMENTS" if "$ARGUMENTS" else "")
has_worktree_flag = "--worktree" in command_args
has_branch_flag = "--branch" in command_args

if has_worktree_flag:
    # Worktree creation logic
    try:
        # Import WorktreeManager from the implemented CLI module
        from moai_adk.cli.worktree.manager import WorktreeManager

        # Determine paths
        project_root = Path.cwd()  # Current working directory
        worktree_root = Path.home() / "worktrees" / "MoAI-ADK"  # Default worktree root

        # Initialize worktree manager
        worktree_manager = WorktreeManager(project_root, worktree_root)

        # Create worktree for the SPEC
        worktree_info = worktree_manager.create(
            spec_id="SPEC-{SPEC_ID}",
            branch_name=f"feature/SPEC-{SPEC_ID}",
            base_branch="main"
        )

        # Display success messages
        print(f"\n SPEC created: SPEC-{SPEC_ID}")
        print(f" Worktree created: {worktree_info.path}")
        print(f"\n Next steps:")
        print(f"  1. Switch to worktree: moai-worktree switch SPEC-{SPEC_ID}")
        print(f"  2. Or use shell eval: eval $(moai-worktree go SPEC-{SPEC_ID})")
        print(f"  3. Then run: /moai:2-run SPEC-{SPEC_ID}")

    except Exception as e:
        # Handle worktree creation errors gracefully
        print(f"\n Worktree creation failed: {e}")
        print(f" SPEC created: SPEC-{SPEC_ID}")
        print(f" You can manually create worktree later with:")
        print(f"   moai-worktree new SPEC-{SPEC_ID}")
```

Expected Success Outcome:

```
 SPEC created: SPEC-AUTH-001
 Worktree created: ~/worktrees/MoAI-ADK/SPEC-AUTH-001

 Next steps:
  1. Switch to worktree: moai-worktree switch SPEC-AUTH-001
  2. Or use shell eval: eval $(moai-worktree go SPEC-AUTH-001)
  3. Then run: /moai:2-run SPEC-AUTH-001
```

Error Handling:

- If worktree creation fails: SPEC is still created, show manual worktree creation instructions
- If worktree already exists: Show switch instructions
- If WorktreeManager not available: Show installation/dependency instructions

---

#### Step 2.6: Team Mode - Create Draft PR (After Branch Creation)

CONDITION: `git_mode == "team"` AND branch was created (Step 2.3) AND NOT `--worktree` flag

ACTION: Create draft PR for team review

Use the manager-git subagent to:

Create draft pull request for SPEC implementation (Team mode only)

Critical Config: git_strategy.mode == "team"
→ Team mode REQUIRES draft PRs for review coordination

Tasks:

1. Create draft PR: feature/SPEC-{SPEC_ID} → main/develop branch
2. PR title: "feat(spec): Add SPEC-{SPEC_ID} [DRAFT]"
3. PR body: Include SPEC ID, description, and checklist
4. Add appropriate labels (spec, draft, etc.)
5. Assign reviewers from team config (if configured)
6. Set PR as DRAFT (do NOT auto-merge)

Validation:

- Verify PR was created in draft status
- Return PR URL and status

Expected Outcome:

```
 Feature branch: feature/SPEC-{SPEC_ID}
 Draft PR created for team review
 Ready for /moai:2-run implementation
```

---

### Step 3: Conditional Status Report

Display status based on configuration and execution result:

#### Case 1: Branch Creation Prompted (`prompt_always: true`) - User Selected "Auto create"

```
 Phase 3 Status: Feature Branch Created (User Choice)

 Configuration: git_strategy.mode = "{git_mode}"
 Branch Creation: prompt_always = true → User chose "Auto create"

 Feature Branch Created:
- Branch: `feature/SPEC-{SPEC_ID}`
- Current branch switched to feature branch
- Ready for implementation on isolated branch

{IF TEAM MODE:
 Draft PR Created (Team Mode):
- PR Title: "feat(spec): Add SPEC-{SPEC_ID} [DRAFT]"
- Target Branch: develop/main
- Status: DRAFT (awaiting review)
}

 Next Steps:
1.  Review SPEC in `.moai/specs/SPEC-{SPEC_ID}/`
2.  Execute `/moai:2-run SPEC-{SPEC_ID}` to begin implementation
3. 🌿 All commits will be made to feature branch
{IF TEAM MODE:
4.  Share draft PR with team for early review (already created)
5.  Team can comment during development
6.  Finalize PR in `/moai:3-sync` when complete
:ELSE:
4.  Create PR in `/moai:3-sync` when implementation complete
}
```

---

#### Case 2: Branch Creation Prompted (`prompt_always: true`) - User Selected "Use current branch"

```
 Phase 3 Status: Direct Commit Mode (User Choice)

 Configuration: git_strategy.mode = "{git_mode}"
 Branch Creation: prompt_always = true → User chose "Use current branch"

 No Branch Created:
- SPEC files created on current branch
- Ready for direct implementation
- Commits will be made directly to current branch

 Next Steps:
1.  Review SPEC in `.moai/specs/SPEC-{SPEC_ID}/`
2.  Execute `/moai:2-run SPEC-{SPEC_ID}` to begin implementation
3.  All commits will be made directly to current branch
4.  Follow TDD: RED → GREEN → REFACTOR cycles
```

---

#### Case 3: Branch Creation Auto-Skipped (Manual Mode + `prompt_always: false`)

```
 Phase 3 Status: Direct Commit Mode (Configuration)

 Configuration: git_strategy.mode = "manual"
 Branch Creation: prompt_always = false → Auto-skipped

 No Branch Created (Manual Mode Default):
- SPEC files created on current branch
- NO manager-git invoked (as configured)
- Ready for direct implementation
- Commits will be made directly to current branch

 Next Steps:
1.  Review SPEC in `.moai/specs/SPEC-{SPEC_ID}/`
2.  Execute `/moai:2-run SPEC-{SPEC_ID}` to begin implementation
3.  Make commits directly to current branch
4.  Follow TDD: RED → GREEN → REFACTOR cycles
```

---

#### Case 4: Branch Creation Skipped with Auto-Enable Prompt (Personal/Team + `prompt_always: false` + `auto_enabled: false`)

```
 Phase 3 Status: Direct Commit Mode (Manual Default for Personal/Team)

 Configuration: git_strategy.mode = "{git_mode}" (personal or team)
 Branch Creation: prompt_always = false, auto_enabled = false → Manual Default

 Branch Creation: Not created yet (pending user approval)
- SPEC files created on current branch
- Ready for implementation
- Commits will be made directly to current branch initially

 Automation Approval Offered:
─────────────────────────────────────────
Would you like to enable automatic branch creation for future SPEC creations?
(This will update your config.yaml)

 Yes  → Set branch_creation.auto_enabled = true
        → Next SPEC will auto-create feature/SPEC-XXX branch

 No   → Keep manual mode
        → Continue working on current branch for this SPEC
        → No config changes made
─────────────────────────────────────────

 Next Steps:
1.  Review SPEC in `.moai/specs/SPEC-{SPEC_ID}/`
2.  Execute `/moai:2-run SPEC-{SPEC_ID}` to begin implementation
3.  Make commits directly to current branch
4.  Follow TDD: RED → GREEN → REFACTOR cycles
5.  Create PR in `/moai:3-sync` when implementation complete
```

---

#### Case 5: Branch Creation Auto-Enabled (Personal/Team + `prompt_always: false` + `auto_enabled: true`)

```
 Phase 3 Status: Feature Branch Created (Auto-Enabled)

 Configuration: git_strategy.mode = "{git_mode}" (personal or team)
 Branch Creation: prompt_always = false, auto_enabled = true → Auto-enabled

 Feature Branch Created:
- Branch: `feature/SPEC-{SPEC_ID}`
- Current branch switched to feature branch
- Ready for implementation on isolated branch

{IF TEAM MODE:
 Draft PR Created (Team Mode):
- PR Title: "feat(spec): Add SPEC-{SPEC_ID} [DRAFT]"
- Target Branch: develop/main
- Status: DRAFT (awaiting review)
}

 Next Steps:
1.  Review SPEC in `.moai/specs/SPEC-{SPEC_ID}/`
2.  Execute `/moai:2-run SPEC-{SPEC_ID}` to begin implementation
3. 🌿 All commits will be made to feature branch
{IF TEAM MODE:
4.  Share draft PR with team for early review
5.  Team can comment on draft PR during development
6.  Finalize PR in `/moai:3-sync` when complete
:ELSE:
4.  Create PR in `/moai:3-sync` when implementation complete
}
```

---

#### Case 6: Worktree Creation (--worktree flag or user choice)

```
 Phase 3 Status: Worktree Created (Isolated Development Environment)

 Worktree Creation: --worktree flag provided OR user chose "Create Worktree"
 SPEC Created: SPEC-{SPEC_ID} documents generated successfully

 Isolated Worktree Created:
- Path: ~/worktrees/MoAI-ADK/SPEC-{SPEC_ID}/
- Branch: feature/SPEC-{SPEC_ID}
- Status: Ready for parallel development

 Next Steps:
1.  Switch to worktree: `moai-worktree switch SPEC-{SPEC_ID}`
2.  Or use shell eval: `eval $(moai-worktree go SPEC-{SPEC_ID})`
3.  Review SPEC documents in worktree: `.moai/specs/SPEC-{SPEC_ID}/`
4.  Execute `/moai:2-run SPEC-{SPEC_ID}` to begin TDD implementation
5.  Work on isolated environment without affecting other SPECs

 Benefits of Worktree Development:
-  Complete isolation from other SPEC work
- 🔀 Easy switching between multiple SPECs
- 🧹 Automatic cleanup when SPEC is completed
-  Lower memory usage than full repository clones
```

---

## Output Format

All command execution phases must produce structured output with semantic XML sections:

Analysis Output Format:

Project analysis results structured as:
- Context: Current project state and relevant files discovered
- Findings: SPEC candidates identified with rationale
- Assessment: Technical constraints and implementation feasibility
- Recommendations: Next steps and decision options

Plan Output Format:

SPEC planning results structured as:
- Requirements: Approved SPEC title, ID, priority, and scope
- Architecture: Technical stack, dependencies, and integration points
- Decomposition: Task breakdown and implementation sequence
- Validation: Quality criteria and acceptance conditions

Implementation Output Format:

SPEC creation results structured as:
- Status: Phase completion status and artifacts created
- Artifacts: Location and format of created SPEC files
- Validation: Quality gate results and compliance verification
- NextSteps: User guidance for proceeding to implementation phase

WHY: Structured output enables parsing for automated workflows and tool integration
IMPACT: Unstructured output prevents downstream automation and creates manual overhead

---

##  Summary: Your Execution Checklist

Before you consider this command complete, verify:

- [ ] PHASE 1 executed: manager-spec analyzed project and proposed SPEC candidates
- [ ] Progress report displayed: User shown detailed progress report with analysis results
- [ ] User approval obtained: User explicitly approved SPEC creation (via enhanced AskUserQuestion)
- [ ] PHASE 2 executed: manager-spec created all 3 SPEC files (spec.md, plan.md, acceptance.md)
- [ ] Directory naming correct: `.moai/specs/SPEC-{ID}/` format followed
- [ ] YAML frontmatter valid: All 7 required fields present
- [ ] HISTORY section present: Immediately after YAML frontmatter
- [ ] EARS structure complete: All 5 requirement types included
- [ ] PHASE 3 executed: Appropriate action taken based on flags/user choice:
  - [ ] If --worktree: WorktreeManager created isolated worktree environment
  - [ ] If --branch: manager-git created feature branch
  - [ ] If prompt: User choice implemented (worktree/branch/current)
  - [ ] If Team mode: Draft PR created (when branch created, not worktree)
- [ ] Branch/Worktree naming correct: `feature/SPEC-{ID}` format for branches, `SPEC-{ID}` for worktrees
- [ ] Next steps presented: User shown appropriate guidance for worktree navigation or branch development
- [ ] Worktree guidance displayed: Worktree switch/eval instructions shown (when applicable)

IF all checkboxes are checked → Command execution successful

IF any checkbox is unchecked → Identify missing step and complete it before ending

---

##  Quick Reference

| Scenario               | Mode               | Entry Point                                                         | Key Phases                                                   | Expected Outcome                             |
| ---------------------- | ------------------ | ------------------------------------------------------------------- | ------------------------------------------------------------ | -------------------------------------------- |
| Clear feature request  | Direct to Planning | `/moai:1-plan "feature description"`                                | Phase 1B → Phase 2 → Phase 3                                 | SPEC created + branch/worktree (conditional) |
| Vague user request     | Exploration First  | `/moai:1-plan "vague request"`                                      | Phase 1A → Phase 1B → Phase 2 → Phase 3                      | Exploration → SPEC + branch/worktree         |
| Resume draft SPEC      | Resume Existing    | `/moai:1-plan resume SPEC-XXX`                                      | Phase 1B → Phase 2 → Phase 3                                 | Complete existing SPEC                       |
| Worktree creation  | NEW            | `/moai:1-plan "feature" --worktree`                                 | Phase 1B → Phase 2 → Phase 3 (worktree)                      | SPEC + isolated worktree environment     |
| Branch creation prompt | User Choice        | `/moai:1-plan "feature"` (prompt_always: true)                      | Phase 1-2 → User chooses (worktree/branch/current) → Phase 3 | SPEC + user-selected strategy                |
| Auto branch creation   | Automated          | `/moai:1-plan "feature"` (prompt_always: false, auto_enabled: true) | Phase 1-2 → Auto branch creation → Phase 3                   | SPEC + auto branch (Personal/Team)           |

###  New Worktree Workflow Examples

Basic Worktree Creation:

```bash
/moai:1-plan "User authentication system" --worktree
# Output:
#  SPEC created: SPEC-AUTH-001
#  Worktree created: ~/worktrees/MoAI-ADK/SPEC-AUTH-001
#
#  Next steps:
#   1. Switch to worktree: moai-worktree switch SPEC-AUTH-001
#   2. Or use shell eval: eval $(moai-worktree go SPEC-AUTH-001)
```

Interactive Environment Selection:

```bash
/moai:1-plan "Payment integration"
# User prompted to choose:
# - Create Worktree (recommended for parallel development)
# - Create Branch (traditional workflow)
# - Use current branch
```

Associated Agents & Components:

- `Explore` - Project exploration and file discovery (Phase 1A, optional)
- `manager-spec` - SPEC planning and document creation (Phase 1B-2, required)
- `manager-git` - Branch and PR creation (Phase 3, conditional)
- WorktreeManager - Worktree creation and management (Phase 3, when --worktree flag used)

Key Integration Points:

- WorktreeManager Import: `from moai_adk.cli.worktree.manager import WorktreeManager`
- Worktree Registry: Automatic registration in `~/worktrees/MoAI-ADK/.moai-worktree-registry.json`
- Git Integration: Creates feature branch `feature/SPEC-{ID}` and associated worktree
- Error Handling: Graceful fallback if worktree creation fails

SPEC Documents Directory:

- Location: `.moai/specs/SPEC-{ID}/` (directory format, NOT single .md file)
- Files: `spec.md`, `plan.md`, `acceptance.md` (created simultaneously via MultiEdit)
- Format: EARS structure with YAML frontmatter + HISTORY section
- Language: All content in user's conversation_language

Version: 5.1.0 (4-Step Agent-Based Workflow + Worktree Integration)
Last Updated: 2025-11-28
Architecture: Commands → Agents → Skills (Complete delegation)
NEW: WorktreeManager integration for parallel SPEC development

---

##  SPEC-WORKTREE-001 Integration Status

Status:  COMPLETE - Full integration achieved on 2025-11-28

### What Was Implemented

1. --worktree Flag Support: Added argument parsing for `--worktree` flag in `/moai:1-plan`
2. WorktreeManager Integration: Automatic worktree creation using existing `src/moai_adk/cli/worktree/manager.py`
3. Guidance Messages: Clear next-step instructions for worktree navigation
4. Interactive Flow: AskUserQuestion integration for worktree/branch/current choice
5. Error Handling: Graceful fallback when worktree creation fails

### Expected Behavior

```bash
# Command execution:
/moai:1-plan "User authentication" --worktree

# Expected output:
 SPEC created: SPEC-AUTH-001
 Worktree created: ~/worktrees/MoAI-ADK/SPEC-AUTH-001

 Next steps:
  1. Switch to worktree: moai-worktree switch SPEC-AUTH-001
  2. Or use shell eval: eval $(moai-worktree go SPEC-AUTH-001)
  3. Then run: /moai:2-run SPEC-AUTH-001
```

### Integration Points

- Import: `from moai_adk.cli.worktree.manager import WorktreeManager`
- Worktree Registry: Automatic registration in `~/worktrees/MoAI-ADK/.moai-worktree-registry.json`
- Branch Creation: Creates feature branch `feature/SPEC-{SPEC_ID}` automatically
- Documentation: Updated all examples, checklists, and status reports

### Completion Criteria (All Met)

-  Flag Parsing: `--worktree` flag detected and processed correctly
-  Worktree Creation: WorktreeManager.create() called with correct parameters
-  User Guidance: Next steps displayed in user-friendly format
-  Error Handling: Fallback messages when worktree creation fails
-  Documentation: All references updated with worktree scenarios
-  Backward Compatibility: Existing --branch and default behavior preserved

SPEC-WORKTREE-001: 100% Complete - All 85% existing implementation + 15% missing integration now complete

---

## End of command execution guide

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

Important:

- Use conversation language from config
- No emojis in any AskUserQuestion fields
- Always provide clear next step options

##  EXECUTION DIRECTIVE

You must NOW execute the command following the "The 4-Step Agent-Based Workflow Command Logic" described above.

1. Start PHASE 1: Project Analysis & SPEC Planning immediately.
2. Use the manager-spec subagent (or Explore subagent as appropriate).
3. Do NOT just describe what you will do. DO IT.
