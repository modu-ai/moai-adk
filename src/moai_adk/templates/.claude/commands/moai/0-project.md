---
name: moai:0-project
description: "Initialize project metadata and documentation"
argument-hint: "[<empty>|setting|update|--glm-on <token>]"
allowed-tools: Read, Write, Edit, Grep, Glob, WebFetch, WebSearch, Bash, TodoWrite, AskUserQuestion, Task, Skill
model: inherit
---

## Pre-execution Context

!git status --porcelain
!git config --get user.name
!git config --get user.email
!git branch --show-current

## Essential Files

@.moai/config/config.yaml
@.moai/project/product.md
@.moai/project/structure.md
@.moai/project/tech.md

---

# MoAI-ADK Step 0: Initialize/Update Project (Project Setup)

User Interaction Architecture: AskUserQuestion tool must be used at COMMAND level only, not within subagents. Subagents invoked via Task() operate in isolated, stateless contexts and cannot interact with users directly.

Correct Pattern: Command collects user input via AskUserQuestion BEFORE delegating to agents. Pass collected choices as parameters to Task() calls.

Architecture: Commands delegate to Agents, which coordinate Skills. This command orchestrates exclusively through Task() tool.

Delegation Model: Complete agent-first pattern. All execution delegated to manager-project agent. Agents receive pre-collected user choices and execute without further user interaction.

Workflow Integration: This command implements Step 0 of Alfred's three-step execution model (Understand-Plan-Execute). See CLAUDE.md for complete workflow details.

---

##  Command Purpose

Initialize or update project metadata using language-first architecture. The system supports five execution modes:

INITIALIZATION Mode: First-time project setup and configuration
AUTO-DETECT Mode: Existing projects with optional modification or re-initialization
SETTINGS Mode: Interactive tab-based configuration management with validation
UPDATE Mode: Template optimization after moai-adk package update
GLM Configuration Mode: GLM API integration setup via --glm-on parameter

WHY: Multi-mode design accommodates diverse user scenarios from fresh installs to updates.

IMPACT: Users can navigate project lifecycle without manual intervention.

---

##  Associated Agents and Skills

The following agents and skills support this command:

manager-project agent orchestrates language-first initialization and configuration workflows.

moai-workflow-project skill provides unified project management including language initialization, atomic config operations, template merging, and tab-based batch question execution.

moai-workflow-templates skill manages template generation and customization.

WHY: Distributed responsibility enables specialized expertise and focused tool access.

IMPACT: Each agent optimizes its domain while maintaining system coherence.

---

##  Language Configuration

Core Principle: Language configuration originates from moai-adk CLI initialization or update commands.

[HARD] Read language from .moai/config/config.yaml before starting any mode.

WHY: Preserving existing language settings prevents disruption to user experience.

IMPACT: Missing language context causes mode selection ambiguity.

Language is preserved across modes except when:

- [SOFT] User explicitly requests language change via SETTINGS mode, Tab 1
- [SOFT] Update mode detects language-compatible improvements

Execution sequence varies by mode:

Initialization mode: Read language from config if present, conduct project interview, generate documentation.

Auto-Detect mode: Display current language, offer settings modification with language change shortcut.

Settings mode: Display current language in Tab 1, allow optional language modification.

Update mode: Preserve language from backup, perform template optimization.

WHY: Mode-specific language handling respects existing configuration while enabling user choice.

---

## YAML-Based Question System

### Question Definition Files

All configuration questions are defined in YAML files under .moai/config/questions/:

- _schema.yaml: Schema definition and constraints
- tab1-user.yaml: User and language settings (3 questions)
- tab2-project.yaml: Project metadata (2 questions)
- tab3-git.yaml: Git strategy and workflow (26 questions, includes GitHub Profile)
- tab4-quality.yaml: Quality principles and reports (7 questions)
- tab5-system.yaml: System and GitHub integration (7 questions)

WHY: YAML-based questions enable consistent structure and easy maintenance.

IMPACT: Question changes require only YAML updates, not code modifications.

### Language-Aware Question Execution

[HARD] When executing questions from .moai/config/questions/*.yaml, translate to user's conversation_language at runtime.

Question files are written in English (source of truth). At execution time:
- Read user's conversation_language from sections/language.yaml
- Translate question text, options, and descriptions to user's language
- Present AskUserQuestion in user's language
- Store answer values in English (field values remain English)

Example Translation Flow:
- Question YAML (English): "What is your name?"
- User Language: Korean (ko)
- AskUserQuestion presented: Korean translation of question
- Answer stored: user.name = "GOOS" (value unchanged)

WHY: Single-source English questions with runtime translation reduces maintenance burden.

IMPACT: Adding new languages requires only translation logic, not question file duplication.

### Question Loading Priority

1. Load question definitions from .moai/config/questions/tab*.yaml
2. Read current values from .moai/config/sections/*.yaml
3. Present questions with current values as defaults
4. Store updated values back to sections/*.yaml

### Section File Updates

Configuration values are stored in modular section files:
- sections/user.yaml: User name (loaded by CLAUDE.md)
- sections/language.yaml: All language settings (loaded by CLAUDE.md)
- sections/project.yaml: Project metadata
- sections/git-strategy.yaml: Git workflow configuration
- sections/quality.yaml: TDD and quality settings
- sections/system.yaml: MoAI system settings

WHY: Modular sections enable token-efficient CLAUDE.md loading.

IMPACT: CLAUDE.md loads only user.yaml and language.yaml (~17 lines vs 400+ full config).

---

## Agent Invocation Patterns (CLAUDE.md Compliance)

This command uses agent execution patterns defined in CLAUDE.md (lines 96-120).

### Sequential Phase-Based Chaining PASS

Command implements sequential chaining through 3 main phases:

Phase Flow:
- Phase 1: Mode Detection & Language Configuration (determines operation mode and loads language)
- Phase 2: User Intent Collection (gathers project metadata and preferences via AskUserQuestion)
- Phase 3: Configuration & Documentation Generation (manager-project generates all project files)

Each phase receives outputs from previous phases as context.

WHY: Sequential execution ensures proper configuration and file generation
- Phase 2 requires mode and language from Phase 1 to display appropriate questions
- Phase 3 requires user responses from Phase 2 to generate correct configuration
- File generation requires complete configuration before creating documentation

IMPACT: Skipping phases would create incomplete or misconfigured project setup

### Parallel Execution FAIL

Not applicable - configuration requires sequential processing

WHY: Project initialization has strict ordering requirements
- Language must be determined before asking questions
- User responses must be collected before generating files
- Configuration must be validated before documentation generation

IMPACT: Parallel execution would create configuration conflicts and invalid project state

### Resumable Agent Support FAIL

Not applicable - command completes in single execution

WHY: Project setup is fast atomic operation
- Most setups complete in under 1 minute
- Configuration operations are atomic and transactional
- No long-running processes requiring checkpoints

IMPACT: Resume pattern unnecessary for typical project initialization

---

Refer to CLAUDE.md "Agent Chaining Patterns" (lines 96-120) for complete pattern architecture.

---

##  Execution Philosophy

The project setup follows explicit delegation pattern: Understand the user intent, Plan the execution, Execute through specialized agents.

The command delegates all functional work to manager-project agent through Task() invocation.

The command maintains zero direct tool usage except for Task() orchestration and AskUserQuestion() user interaction.

[HARD] Tool Usage Restrictions:

- DO NOT use Read for file operations (delegated to manager-project)
- DO NOT use Write for file operations (delegated to manager-project)
- DO NOT use Edit for file modifications (delegated to manager-project)
- DO NOT use Bash for command execution (delegated to manager-project)
- DO NOT use TodoWrite for task management (delegated to manager-project)
- DO use Task() for agent orchestration
- DO use AskUserQuestion() for user interaction

WHY: Delegation enables specialized agent tool access and consistent execution context.

IMPACT: Direct tool usage would bypass agent expertise and validation layers.

Manager-project agent handles all implementation complexity including file operations, configuration management, and validation logic.

---

##  PHASE 1: Command Routing and Analysis

Goal: Detect subcommand intent and prepare execution context for delegation.

### Step 1: Analyze User Command Arguments

[HARD] Parse user command to determine execution mode.

WHY: Correct mode detection ensures appropriate workflow execution.

IMPACT: Incorrect routing causes wrong execution flow.

The system routes based on provided arguments:

GLM Configuration: Command includes --glm-on with optional API token
- Detect token in --glm-on parameter
- If token missing: Attempt auto-load from .env.glm file
- If token missing: Attempt auto-load from ANTHROPIC_AUTH_TOKEN environment variable
- If all sources missing: Request token from user via AskUserQuestion

SETTINGS Mode: Command includes setting argument
- Always use interactive tab selection via AskUserQuestion
- User selects specific tab or "Modify All Tabs" option

UPDATE Mode: Command includes update argument
- No additional argument parsing required

INITIALIZATION or AUTO-DETECT: Command has no arguments
- [HARD] Check if .moai/config/config.yaml exists
- File exists: Route to AUTO-DETECT MODE
- File missing: Route to INITIALIZATION MODE

Invalid Arguments: Unrecognized command format
- Display error message explaining valid command syntax
- Exit with error state

WHY: Argument-driven routing prevents mode confusion.

IMPACT: Ambiguous routing leads to wrong workflow execution.

### Step 2: Delegate to Manager-Project Agent

[HARD] Invoke manager-project subagent with detected mode and context.

WHY: Specialized agent handles mode-specific complexity.

IMPACT: Direct execution would bypass validation and expertise layers.

Pass the following context to manager-project agent:

- Detected Mode value (INITIALIZATION, AUTO-DETECT, SETTINGS, UPDATE, or GLM_CONFIGURATION)
- Language Context read from .moai/config/config.yaml if present
- GLM Token value if GLM_CONFIGURATION mode selected
- User command arguments for reference

For INITIALIZATION:

- Check .moai/config.yaml for language setting
- If missing: Use moai-workflow-project skill for language detection
- If present: Use existing language, skip language selection
- Conduct language-aware user interview
- Generate project documentation
- Use moai-workflow-project skill for config creation

For AUTO-DETECT:

- Read current language from .moai/config.yaml
- Check if project documentation exists (.moai/project/product.md, structure.md, tech.md)
- If docs missing → PARTIAL INITIALIZATION state detected
  - Use AskUserQuestion to ask user: "Your configuration exists but project documentation is missing. Would you like to complete the initialization now?"
  - Options: "Yes, complete initialization" / "No, review configuration" / "Cancel"
  - If user selects "Yes" → Switch to INITIALIZATION workflow
  - Otherwise → Continue with regular AUTO-DETECT options
- Display current configuration (including language)
- Offer: Modify Settings / Change Language Only / Review Configuration / Re-initialize / Cancel
- If "Change Language Only" → Go to Tab 1 in SETTINGS mode
- Otherwise route to selected sub-action

For SETTINGS:

- Load current language from .moai/config.yaml
- Load tab schema from appropriate skill schema
- Execute batch questions via moai-workflow-project skill
- Process responses and update config.yaml atomically via moai-workflow-project skill
- Report changes and validation results

For UPDATE:

- Read language from config backup (preserve existing setting)
- Use moai-workflow-project skill for smart merging
- Update templates and configuration
- Auto-translate announcements to current language if needed

For GLM_CONFIGURATION:

- Receive GLM API token from parameter (or detect from environment)
- Check token resolution sequence:
  1. Use provided token from `--glm-on <token>` argument (if not empty)
  2. Auto-load from existing `.env.glm` file (if exists and token missing)
  3. Auto-load from `ANTHROPIC_AUTH_TOKEN` environment variable (if set)
  4. Request from user via AskUserQuestion (if all above missing)
- Execute GLM setup script: `uv run .moai/scripts/setup-glm.py <GLM_TOKEN>`
- Verify configuration in .claude/settings.local.json with proper keys
- Verify .env.glm created with secure permissions (0o600)
- Verify .gitignore includes .env.glm entry
- Report GLM configuration success to user with all configured keys
- Remind user: "Restart Claude Code to automatically load the new settings"

Output: Mode-specific completion report with next steps

Store: Response in `$MODE_EXECUTION_RESULT`

---

##  PHASE 1.5: Smart Question System (INITIALIZATION Mode Only)

Goal: Collect minimal user input, research best practices via web search, and propose optimized project configuration.

WHY: Reduce user burden from 40+ questions to 3 essential questions while providing expert-level recommendations based on current industry standards.

IMPACT: Non-developers can create production-ready projects with minimal technical knowledge.

### Step 1: Collect Essential Information (3 Questions Only)

[HARD] Use AskUserQuestion to collect ONLY 3 required fields at COMMAND level.

WHY: Minimal questions reduce user friction while capturing essential project identity.

IMPACT: More questions would overwhelm non-developers and slow down project creation.

Question 1 - Project Name:

- question: Translate to user's conversation_language with friendly guidance:
  - Korean: "프로젝트 이름을 정해주세요. 영어 소문자와 하이픈(-)만 사용할 수 있어요."
  - English: "Choose a project name. Use only lowercase letters and hyphens (-)."
- header: "Project Name"
- type: text input (user selects "Other" to type)
- examples shown to user:
  - "snake-game" (게임 프로젝트)
  - "my-first-app" (첫 번째 앱)
  - "todo-list" (할 일 목록)
- validation: lowercase, no spaces, valid directory name
- helper: If user enters Korean or spaces, suggest alternatives automatically

Question 2 - Project Type:

- question: Translate to user's conversation_language with beginner-friendly tone:
  - Korean: "어떤 종류의 프로젝트를 만들고 싶으세요?"
  - English: "What kind of project would you like to create?"
- header: "Project Type"
- multiSelect: false
- options (max 4, beginner-friendly labels):
  - label: "Web Browser App"
    korean_label: "웹 브라우저에서 보는 것"
    description: "크롬, 사파리 등 브라우저에서 열어보는 프로그램 (게임, 웹사이트 등)"
    tech_hint: "HTML, CSS, JavaScript 또는 React, Vue 사용"
  - label: "Mobile App"
    korean_label: "스마트폰 앱"
    description: "아이폰이나 안드로이드에서 설치해서 사용하는 앱"
    tech_hint: "React Native, Flutter 사용"
  - label: "Command Line Tool"
    korean_label: "터미널에서 실행하는 도구"
    description: "명령어로 실행하는 프로그램 (자동화 스크립트 등)"
    tech_hint: "Python, Node.js, Go 사용"
  - label: "Backend API"
    korean_label: "서버/백엔드 서비스"
    description: "데이터를 저장하고 처리하는 서버 프로그램"
    tech_hint: "FastAPI, Express, Django 사용"

Question 3 - Project Description:

- question: Translate to user's conversation_language with examples:
  - Korean: "만들고 싶은 것을 한 문장으로 설명해주세요."
  - English: "Describe what you want to build in one sentence."
- header: "Description"
- type: text input (user selects "Other" to type)
- examples shown to user (in conversation_language):
  - "방향키로 뱀을 조종해서 음식을 먹는 게임" (스네이크 게임)
  - "할 일을 추가하고 완료 표시하는 앱" (투두 리스트)
  - "날씨 정보를 보여주는 웹페이지" (날씨 앱)
  - "사진을 업로드하고 공유하는 서비스" (갤러리)
- validation: non-empty string
- helper: Encourage natural language description, no technical terms required

Store responses in: $PROJECT_ESSENTIAL_INFO

### Step 2: Web Search for Best Practices

[HARD] After collecting essential info, perform targeted web searches using WebSearch tool.

WHY: Real-time research ensures recommendations reflect current industry standards (2025).

IMPACT: Outdated recommendations would create technical debt from project start.

Based on collected project_type, execute 3 parallel searches:

Search Query 1: "[project_type] best practices 2025"
- Extract: Design patterns, architecture recommendations, common pitfalls

Search Query 2: "[project_type] recommended tech stack 2025"
- Extract: Framework comparisons, library recommendations, version requirements

Search Query 3: "[project_type] [project_description keywords] architecture patterns"
- Extract: Domain-specific patterns, similar project architectures

Analyze search results and compile:

- Recommended Frameworks: Top 2-3 options with pros/cons
- UI Library: Best options for project type
- Testing Framework: Industry standard for project type
- Build Tools: Current best practices
- Deployment Target: Common deployment platforms

Store analysis in: $BEST_PRACTICES_ANALYSIS

### Step 3: Generate Smart Proposal

[HARD] Create structured proposal based on web search analysis.

WHY: Expert-curated recommendations reduce decision paralysis for non-developers.

IMPACT: Poor recommendations would lead to project failure or rework.

Proposal Structure:

Complexity Detection (before proposing):
- Simple Projects (recommend NO framework):
  - Keywords: "게임", "game", "간단한", "simple", "첫", "first", "연습", "practice"
  - Project types: Browser games, simple utilities, learning projects
  - Recommended stack: HTML5 + CSS3 + Vanilla JavaScript (+ Canvas for games)
  - Rationale: "프레임워크 없이 기본 기술만으로 빠르게 만들 수 있어요"

- Medium Projects (recommend lightweight framework):
  - Keywords: "블로그", "포트폴리오", "랜딩페이지"
  - Recommended stack: Next.js or Astro with minimal dependencies

- Complex Projects (recommend full framework):
  - Keywords: "대시보드", "관리자", "SaaS", "플랫폼"
  - Recommended stack: Full framework with state management, testing, etc.

Technology Stack Proposal:
- For Simple Projects (games, learning):
  - Framework: "프레임워크 없음 - 순수 HTML/CSS/JavaScript로 충분해요"
  - UI: "CSS만 사용 (Canvas for games)"
  - Testing: "브라우저 개발자 도구로 확인"
  - Build: "빌드 도구 없음 - 파일을 바로 브라우저에서 열면 됩니다"

- For Medium/Complex Projects:
  - Framework: Primary recommendation with rationale
  - UI Library: Recommended option
  - Testing: Standard for project type
  - Build Tools: Current best practice

Architecture Proposal:
- Directory Structure: Recommended layout for project type
- Key Patterns: Component patterns, state management approach
- Data Flow: How data moves through the application

Quality Standards:
- Test Coverage Target: Industry standard percentage
- TRUST 5 Level: Recommended validation level
- Accessibility: WCAG level appropriate for project

Store proposal in: $SMART_PROPOSAL

### Step 4: Present Proposal and Get User Approval

[HARD] Use AskUserQuestion to present proposal and get explicit user approval.

WHY: User must consent before any files are created or configuration applied.

IMPACT: Automatic execution without approval violates user autonomy and may create unwanted configuration.

Proposal Presentation Format:

Present in user's conversation_language with clear sections:

```markdown
Based on your project "{project_name}" ({project_type}), here is my recommendation:

**Technology Stack**
- Framework: {framework} - {brief_rationale}
- UI: {ui_library} - {brief_rationale}
- Testing: {test_framework}
- Build: {build_tools}

**Architecture**
- Structure: {directory_pattern}
- Patterns: {key_patterns}

**Quality Standards**
- Coverage: {coverage_target}%
- TRUST Level: {trust_level}

Sources: [List actual WebSearch sources used]
```

Approval Question:

- question: Translate "Do you approve this configuration? You can modify specific parts or request a different approach." to user's conversation_language
- header: "Approval"
- multiSelect: false
- options:
  - label: "Approve and Continue"
    description: "Create project with these settings"
  - label: "Modify Tech Stack"
    description: "Keep architecture, change framework/libraries"
  - label: "Modify Architecture"
    description: "Keep tech stack, change structure/patterns"
  - label: "Start Over"
    description: "Re-run search with different criteria"

If user selects "Modify":
- Use additional AskUserQuestion to collect specific changes
- Update $SMART_PROPOSAL with modifications
- Re-present for approval

Loop until user selects "Approve and Continue"

Store approval in: $USER_APPROVAL

### Step 5: Delegate to Manager-Project Agent

[HARD] After user approval, invoke manager-project agent with complete context.

WHY: Agent receives all necessary information to generate files without further user interaction.

IMPACT: Missing context would force agent to make assumptions or fail.

Pass to manager-project via Task():

- Mode: INITIALIZATION
- Language: User's conversation_language
- Project Essential Info: $PROJECT_ESSENTIAL_INFO (name, type, description)
- Approved Proposal: $SMART_PROPOSAL (tech stack, architecture, quality)
- User Approval Status: $USER_APPROVAL
- WebSearch Sources: URLs used for recommendations

Agent responsibilities:
- Generate .moai/project/product.md from project description + approved proposal
- Generate .moai/project/structure.md from architecture proposal
- Generate .moai/project/tech.md from technology stack proposal
- Create .moai/config/config.yaml with all settings
- Apply language settings throughout

Output: Complete project initialization with documentation

---

##  PHASE 2: Execute Mode

Goal: Execute the appropriate mode based on routing decision.

### Mode Handler: manager-project Agent

The manager-project agent handles all mode-specific workflows:

INITIALIZATION MODE:

- Read language from config.yaml (or use CLI default if missing)
- Conduct language-aware user interview (via Skill)
- Project type detection and configuration
- Documentation generation
- Auto-translate announcements to selected language

AUTO-DETECT MODE:

- Read current language from config.yaml
- CRITICAL CHECK: Detect partial initialization state
  - Check if project documentation exists in `.moai/project/`:
    - product.md, structure.md, tech.md
  - If ANY doc missing → Use AskUserQuestion (in user's language)
    - Question: "Your configuration exists but project documentation is missing. Would you like to complete the initialization?"
    - Options: "Yes, complete initialization" / "No, review configuration" / "Cancel"
    - If "Yes" → Switch to INITIALIZATION workflow
- Display current configuration (including language, initialization status)
- Offer: Modify Settings / Change Language Only / Review Configuration / Re-initialize / Cancel
- "Change Language Only" shortcut → SETTINGS mode, Tab 1 only
- Route to selected sub-action
- Language-Aware: All AskUserQuestion calls in user's conversation_language (NO EMOJIS)

SETTINGS MODE (NEW):

- Read current language from config.yaml
- Load tab schema for batch-based questions
- Execute batch questions with AskUserQuestion
- Process user responses
- Validate settings at critical checkpoints
- Delegate config update to UnifiedConfigManager from moai-workflow-project
- Report changes

UPDATE MODE:

- Preserve language from config backup
- **Config Format Migration (v0.32.0+):**
  - Check if `.moai/config/config.yaml` exists (legacy format)
  - If exists: Convert to `config.yaml` with intelligent optimization:
    - Preserve all user settings and customizations
    - Add meaningful comments explaining each section
    - Optimize structure (remove dead fields, consolidate duplicates)
  - Use YAML's native features:
    - Multi-line strings for long descriptions
    - Inline comments for field explanations
    - Better readability with proper indentation
  - Report migration success to user
- Analyze backup and compare templates
- Perform smart template merging
- Update `.moai/` files with new features
- Auto-translate announcements to current language if needed

### Language-Aware Announcements

After any language selection or change, auto-translate company announcements:

```bash
uv run $CLAUDE_PROJECT_DIR/.claude/hooks/moai/shared/utils/announcement_translator.py
```

This ensures `.claude/settings.json` contains announcements in the user's selected language.

---

##  SETTINGS MODE: Tab-Based Configuration (NEW)

> Version: v2.0.0 | Last Updated: 2025-11-19 | Changes: Removed [tab_ID] arg, added git_strategy.mode selection, expanded Tab 3 with conditional batches, fixed 26 field name errors, +16 settings

### Overview

The SETTINGS MODE uses a tab-based batch question system to provide organized, user-friendly configuration management:

- 5 tabs: Organized by configuration domain
- 17 batches: Grouped questions within tabs (added 5 batches: Batch 3.0, 3.3, 3.5, 3.6, improved organization)
- 57 settings: Complete config.yaml v0.28.0 coverage (+39% from v1.0.0)
- 54 questions: User-facing questions (+14 from v1.0.0)
- Conditional batches: Tab 3 shows Personal/Team/Hybrid batches based on mode selection
- Atomic updates: Safe deep merge with backup/rollback

### Initial Entry Point: Tab Selection Screen

When user runs `/moai:0-project setting` (without tab_ID), present tab selection:

```markdown
Which settings tab would you like to modify?

Options:

1. Tab 1: User & Language

   - Configure user name, conversation language, agent prompt language

2. Tab 2: Project Settings

   - Configure project name, description

3. Tab 3: Git Strategy & Workflow

   - Configure GitHub profile, Manual/Personal/Team Git settings, commit/branch strategy

4. Tab 4: Quality Principles & Reports

   - Configure TRUST 5, report generation, storage location

5. Tab 5: System & GitHub Integration

   - Configure MoAI system, GitHub automation

6. Modify All Tabs
   - Recommended: Tab 1 → Tab 2 → Tab 3 → Others
```

After Tab Completion:

```markdown
Would you like to modify another settings tab?

1. No, finish settings
2. Select another tab
```

### Tab Schema Reference

Location: `.claude/skills/moai-workflow-project/schemas/tab_schema.json`

Tab 1: User & Language (Required Foundation)

- Batch 1.1: Basic settings (3 questions - UPDATED: removed conversation_language_name)
  - User name, conversation language, agent prompt language
  - NOTE: conversation_language_name is auto-updated when conversation_language changes
- Setting count: 3 | Critical checkpoint

Tab 2: Project Settings (Recommended)

- Batch 2.1: Project metadata (2 questions)
  - Project name, description
  - NOTE: GitHub Profile moved to Tab 3, Project Mode removed
- Batch 2.2: Auto-processed locale settings (0 questions - internal analysis only)
  - project.locale, default_language, optimized_for_language (auto-determined from conversation_language)
  - NOTE: No user input needed. These 3 fields update automatically when conversation_language changes
  - Auto-Processing Delegation: Command does NOT perform auto-processing. manager-project agent receives user selections, determines derived fields, then delegates atomic update to UnifiedConfigManager skill.
- Setting count: 2

Tab 3: Git Strategy & Workflow (Recommended with Validation - REDESIGNED v2.0.0)

- Batch 3.0a: GitHub Profile (1 question) - Always shown, required for all modes
  - GitHub username (e.g., @GoosLab) - moved from Tab 2
- Batch 3.0: Workflow mode selection (1 question - Manual/Personal/Team) - Controls visibility of subsequent batches
- Batch 3.1: Manual core settings (4 questions) - CONDITIONAL (Manual only)
- Batch 3.2: Manual push settings (1 question) - CONDITIONAL (Manual only)
- Batch 3.3: Personal core settings (4 questions) - CONDITIONAL (Personal only)
- Batch 3.4: Personal commit & push settings (3 questions) - CONDITIONAL (Personal only)
- Batch 3.5: Personal branch settings (2 questions) - CONDITIONAL (Personal only)
- Batch 3.6: Team core settings (4 questions) - CONDITIONAL (Team only)
- Batch 3.7: Team commit & PR settings (3 questions) - CONDITIONAL (Team only)
- Batch 3.8: Team branch & protection settings (3 questions) - CONDITIONAL (Team only)
- Batch 3.9: Team branch naming settings (2 questions) - CONDITIONAL (Team only)
- Setting count: 26 | Critical checkpoint for Git conflicts & mode consistency

Tab 4: Quality Principles & Reports (Optional - UPDATED v2.0.0)

- Batch 4.1: Constitution settings (3 questions - reduced from 4, renamed minimum_test_coverage→test_coverage_target)
- Batch 4.2: Report generation policy (4 questions - expanded, added warn_user & user_choice)
- Setting count: 9 (same count, better fields)

Tab 5: System & GitHub Integration (Optional - UPDATED v2.0.0)

- Batch 5.1: MoAI system settings (3 questions - updated, aligned with config.yaml v0.28.0)
- Batch 5.2: GitHub automation settings (5 questions - expanded from 3, added templates & spec_workflow fields)
- Setting count: 11 (+3 from v1.0.0)

### Batch Execution Flow

#### Step 1: Load Tab Schema

```markdown
Load: .claude/skills/moai-workflow-project/schemas/tab_schema.json
Extract:

- Tab definition (label, batches)
- Batch questions (max 4 per batch)
- Field mappings to config.yaml paths
- Current values from existing config
- Validation rules
```

#### Step 2: Execute Batch via AskUserQuestion

Single Batch Execution Example (Tab 1, Batch 1.1):

```markdown
Call: AskUserQuestion(
questions: [
{
question: "How would you like to configure the user name? (current: GoosLab)",
header: "User Name",
multiSelect: false,
options: [
{label: "Keep Current Value", description: "Continue using GoosLab"},
{label: "Change", description: "Select Other to enter a new name"}
]
},
{
question: "What language should Alfred use in conversations? (current: Korean/ko)",
header: "Conversation Language",
multiSelect: false,
options: [
{label: "Korean (한국어)", description: "모든 콘텐츠를 한국어로 생성"},
{label: "English (English)", description: "All content will be generated in English"},
{label: "Japanese (日本語)", description: "日本語で全てのコンテンツを生成"},
{label: "Chinese (中文)", description: "中文内容生成"},
{label: "Custom (기타)", description: "직접 언어 코드 입력 (예: th, ar, vi, nl)"}
]
},
{
question: "What is the display name for the selected language? (current: Korean)",
header: "Language Display Name",
multiSelect: false,
options: [...]
},
{
question: "What language should agent prompts use? (current: same as conversation)",
header: "Agent Prompt Language",
multiSelect: false,
options: [...]
}
]
)

Process user responses from AskUserQuestion into config update:
user.name → user_input_or_keep_current
language.conversation_language → selected_value
language.conversation_language_name → user_input_or_keep_current
language.agent_prompt_language → selected_value
```

#### Step 3: Process Responses

Mapping Logic:

```markdown
For each question in batch:

1. Get field path from schema (e.g., "user.name")
2. Get user's response (selected option or custom input)
3. Convert to config.yaml value:
   - "Other" option → Use custom input from user
   - Selected option → Use option's mapped value
   - "Keep current" → Use existing value
4. Build update object: {field_path: new_value}
5. Collect all updates from batch
```

#### Step 4: Validate at Checkpoints

Checkpoint Locations (from tab_schema navigation_flow):

1. After Tab 1 (Language settings):

   - Verify conversation_language is valid (ko, en, ja, es, etc)
   - Verify agent_prompt_language consistency
   - Error recovery: Re-ask Tab 1 if validation fails

2. After Tab 3 (Git strategy):

   - Validate Personal/Team mode conflicts
     - If Personal: main_branch should not be "develop"
     - If Team: PR base must be develop or main (never direct to main)
   - Validate branch naming consistency
   - Error recovery: Highlight conflicts, offer fix suggestions

3. Before Config Update (Final validation):
   - Check all required fields are set (marked required: true in schema)
   - Verify no conflicting settings
   - Validate field value types (string, bool, number, array)
   - Report validation results to user

#### Step 5: Delegate Atomic Config Update to Skill

Update Pattern (Skill-delegated):

```markdown
Delegate ALL config update operations to UnifiedConfigManager from moai-workflow-project:

- Manager handles backup/rollback logic internally
- Manager performs deep merge with validation
- Manager writes atomically to config.yaml
- Manager reports success/failure

Agent responsibilities:

- Collect user responses from AskUserQuestion
- Map responses to config field paths
- Pass update map to Skill
- Report results to user
```

UnifiedConfigManager Responsibilities:

- UnifiedConfigManager from moai-workflow-project handles ALL file operations
- Internal backup/rollback if needed
- Atomic write and validation
- Error reporting

### Implementation Details

#### Tab 1 Execution Example

User runs: `/moai:0-project setting tab_1_user_language`

```
Step 1: Project-manager loads tab schema
Step 2: Extracts Tab 1 (tab_1_user_language)
Step 3: Gets Batch 1.1 (基本設定)
Step 4: Loads current values from config.yaml including extended language settings:
  - User configuration: user.name value
  - Conversation language: language.conversation_language (ko, en, ja, zh, etc.)
  - Agent prompt language: language.agent_prompt_language
  - Additional language settings: git_commit_messages, code_comments, documentation, error_messages
Step 5: Calls AskUserQuestion with 3 questions for core language settings
  - Question 1: "The user name is currently set to [current value]. Is this correct?"
  - Question 2: "What language should Alfred use in conversations? (current: [language display name]/[code])"
  - Question 3: "The agent internal prompt language is currently set to [display name]/[code]. How would you like to configure this?"
Step 6: Receives user responses and maps them to configuration fields
Step 7: Processes each response (map to config fields)
  - user.name response → user.name
  - conversation_language response → language.conversation_language
  - Auto-update: conversation_language_name (ko → Korean, en → English, ja → Japanese, es → Spanish)
  - agent_prompt_language response → language.agent_prompt_language
Step 8: Runs Tab 1 validation checkpoint
  - Check language is valid
  - Verify consistency
Step 9: Delegates atomic update to UnifiedConfigManager from moai-workflow-project
  - Manager handles backup/rollback internally
  - Manager performs deep merge (including auto-updated conversation_language_name)
  - Manager verifies final structure
Step 10: Receives result from Manager
  - Success: Report changes made (4 fields: user.name, conversation_language, conversation_language_name [auto], agent_prompt_language)
  - Failure: Report error from Skill with recovery suggestions
```

#### Tab 3 Validation Example (Complex - NEW v2.0.0)

User runs: `/moai:0-project setting` (or `/moai:0-project setting tab_3_git_strategy`)

```
Step 1: Load Tab 3 (tab_3_git_strategy) - 6 batches total
Step 2: Execute Batch 3.0 (Workflow Mode Selection)
  - User selects: Personal, Team, or Hybrid
  - Validation: Confirm mode selection
Step 3: CONDITIONAL LOGIC - Based on mode:

  IF mode = Personal:
    - Execute Batch 3.1 (Personal core settings)
    - Execute Batch 3.2 (Personal branch & cleanup)
    - Execute Batch 3.3 (Personal protection & merge)
    - Skip Batches 3.4, 3.5, 3.6 (Team batches)

  ELSE IF mode = Team:
    - Skip Batches 3.1, 3.2, 3.3 (Personal batches)
    - Execute Batch 3.4 (Team core settings)
    - Execute Batch 3.5 (Team branch & protection)
    - Execute Batch 3.6 (Team safety & merge)

  ELSE IF mode = Hybrid:
    - Execute ALL batches (3.1-3.6) for full flexibility

Step 4: Run Tab 3 validation checkpoint
  - Validate mode selection consistency
  - Check Personal/Team conflicts
    - Example: If Personal: base_branch should be "main" (not "develop")
    - Example: If Team: prevent_main_direct_merge should be true
    - Example: If Team: default_pr_base must be "develop" or "main"
  - Branch naming consistency
  - Let user confirm or retry if conflicts found

Step 5: Merge all executed batches into single update object
Step 6: Delegate atomic update to UnifiedConfigManager from moai-workflow-project
  - Manager handles backup/rollback internally
  - Manager performs deep merge with final validation

Step 7: Report all 29 settings changes (or 16-20 depending on mode)
```

#### Multi-Tab Workflow Example

User runs: `/moai:0-project setting` (always interactive, no tab_ID) → Tab Selection Screen

```
Flow:
1. Show Tab Selection Screen (Which settings tab would you like to modify?)
2. User selects tab or "Modify All Tabs"
3. Execute selected tab
   - Tab 1 (REQUIRED): User & Language (1 batch, 3 questions)
   - Tab 2 (RECOMMENDED): Project Info (2 batches, 2 questions in batch 2.1 + 0 questions auto-processing in batch 2.2)
   - Tab 3 (RECOMMENDED): Git Strategy (10 batches, 26 questions total, conditional by mode)
     * Batch 3.0a: GitHub Profile (1 question) - always shown
     * Batch 3.0: Mode selection (1 question)
     * Manual mode: Batches 3.1-3.2 (5 questions)
     * Personal mode: Batches 3.3-3.5 (9 questions)
     * Team mode: Batches 3.6-3.9 (12 questions)
   - Tab 4 (OPTIONAL): Quality & Reports (2 batches, 7 questions)
   - Tab 5 (OPTIONAL): System & GitHub (2 batches, 8 questions)
4. After tab completion, ask: "Would you like to modify another settings tab?"
   - No, finish settings (exit)
   - Select another tab (return to step 1)
5. Final atomic update after user finishes all selected tabs

Tab-level behavior:
  - If user cancels mid-tab, changes NOT saved
  - If tab validation fails, user can retry or skip tab
  - After ALL selected tabs complete successfully, perform final atomic update
  - Auto-processing happens during atomic update (e.g., conversation_language_name, locale)
  - Tab 3 conditional batches respect mode selection (shown/hidden based on git_strategy.mode)

Tab completion order (recommended):
  - Tab 1 (REQUIRED): Foundation language settings
  - Tab 2: Project metadata
  - Tab 3: Git workflow strategy (includes GitHub Profile)
  - Tab 4: Quality principles
  - Tab 5: System integration
```

### Tab Schema Structure

```json
{
  "version": "1.0.0",
  "tabs": [
    {
      "id": "tab_1_user_language",
      "label": "Tab 1: User & Language",
      "batches": [
        {
          "batch_id": "1.1",
          "questions": [
            {
              "question": "...",
              "header": "...",
              "field": "user.name",
              "type": "text_input|select_single|select_multiple|number_input",
              "multiSelect": false,
              "options": [...],
              "current_value": "...",
              "required": true
            }
          ]
        }
      ]
    }
  ],
  "navigation_flow": {
    "completion_order": ["tab_1", "tab_2", "tab_3", "tab_4", "tab_5"],
    "validation_sequence": [
      "Tab 1 checkpoint",
      "Tab 3 checkpoint",
      "Final validation"
    ]
  }
}
```

### Critical Rules

MANDATORY:

- Execute ONLY ONE tab per command invocation (unless user specifies "all tabs")
- READ language context from config.yaml before starting SETTINGS MODE
- Run validation at Tab 1, Tab 3, and before final update
- Delegate config update to UnifiedConfigManager from moai-workflow-project (no direct backup in command)
- Report all changes made
- Use AskUserQuestion for ALL user interaction

Configuration Priority:

- `.moai/config/config.yaml` settings ALWAYS take priority
- Existing language settings respected unless user explicitly requests change in Tab 1
- Fresh installs: Language already set by moai-adk CLI, skip language selection

Language:

- Tab schema stored in English (technical field names)
- All user-facing questions in user's conversation_language
- AskUserQuestion must use user's conversation_language for ALL fields

---

##  PHASE 2.5: Save Phase Context

Goal: Persist phase execution results for explicit context passing to subsequent commands.

### Step 1: Extract Context from Agent Response

After manager-project agent completes, extract the following information:

- Project metadata: name, description, language
- Files created: List of generated files with absolute paths
- Tech stack: Primary codebase language
- Next phase: Recommended next command (1-plan)

### Step 2: Delegate Context Saving to manager-project

The manager-project agent handles all context saving:

```markdown
Context data to persist:

- Phase: "0-project"
- Mode: INITIALIZATION|AUTO-DETECT|SETTINGS|UPDATE
- Timestamp: ISO8601 UTC
- Status: completed|failed
- Outputs:
  - project_name
  - project_description
  - language (conversation_language)
  - tech_stack (detected primary language)
  - github_profile (from Tab 3)
- Files created: [list of absolute paths]
- Next phase: "1-plan"

Agent delegates to UnifiedConfigManager from moai-workflow-project:

- Save context via ContextManager
- Handle file path validation
- Implement error recovery (non-blocking)
- Report success/failure
```

Error Handling Strategy:

- Context save failures should NOT block command completion
- Log clear warning messages for debugging
- Allow user to retry manually if needed

---

##  PHASE 3: Completion & Next Steps

Goal: Guide user to next action in their selected language.

### Step 1: Display Completion Status

Show mode-specific completion message in user's language:

- INITIALIZATION: "Project initialization complete"
- AUTO-DETECT: Configuration review/modification complete
- SETTINGS: "Settings updated successfully"
- UPDATE: "Templates optimized and updated"

### Step 2: Offer Next Steps

Use AskUserQuestion in user's language:

- From Initialization: Write SPEC / Review Structure / New Session
- From Settings: Continue Settings / Sync Documentation / Exit
- From Update: Review Changes / Modify Settings / Exit

Critical: NO EMOJIS in AskUserQuestion fields. Use clear text only.

---

##  Critical Rules

### Mode Execution

[HARD] Execute exactly ONE mode per command invocation.

WHY: Multi-mode execution causes state inconsistency.

IMPACT: Executing multiple modes corrupts configuration state.

Exception: [SOFT] Allow mode switching only when user explicitly selects different mode in SETTINGS tab selection screen.

### Language Handling

[HARD] Always use user's conversation_language for all output and prompts.

WHY: User comprehension requires language consistency.

IMPACT: Using wrong language reduces accessibility.

[HARD] Never skip language confirmation in INITIALIZATION mode.

WHY: Language choice determines all subsequent output.

IMPACT: Missing language selection blocks localized experience.

[SOFT] Auto-translate announcements after language changes using announcement_translator.py.

WHY: Announcements should reflect user's selected language.

### Agent Delegation

[HARD] Delegate ALL execution to manager-project agent.

WHY: Agent provides specialized expertise and validation.

IMPACT: Direct execution bypasses error recovery and consistency checks.

[HARD] Route to correct mode based on command argument analysis.

WHY: Correct routing enables appropriate workflow.

IMPACT: Wrong routing causes unexpected behavior.

### User Interaction

[HARD] Use AskUserQuestion for ALL user interaction at COMMAND level only.

WHY: Subagents invoked via Task() operate in isolated, stateless contexts and cannot interact with users. AskUserQuestion requires real-time user response which subagents cannot receive.

IMPACT: Attempting AskUserQuestion in subagents causes workflow failures or silent failures.

[HARD] Collect all user choices via AskUserQuestion BEFORE delegating to agents.

WHY: Subagents need pre-collected user decisions to execute without interaction.

IMPACT: Missing parameters force subagents to make assumptions or fail.

[HARD] Pass user choices as parameters when invoking Task().

WHY: Agents must receive all necessary context at invocation time since they cannot request more.

IMPACT: Incomplete context causes agents to fail or produce incorrect results.

[HARD] Never include EMOJI characters in AskUserQuestion fields.

WHY: Emoji parsing varies across platforms and may cause display issues.

IMPACT: Platform inconsistencies create confusing user experience.

[HARD] Maximum 4 options per AskUserQuestion question.

WHY: Tool constraint limits options to 4 per question.

IMPACT: Exceeding 4 options causes tool execution failure. Use multi-step questions for more choices.

### Tool Usage Constraints

[HARD] Use ONLY Task() for agent orchestration and AskUserQuestion() for interaction.

WHY: Direct tool usage bypasses agent expertise.

IMPACT: Tool misuse causes validation and consistency failures.

Prohibited Direct Tools:

- NO Read for file operations
- NO Write for file creation or modification
- NO Edit for configuration changes
- NO Bash for command execution
- NO TodoWrite for task management

All tool-based operations delegate to manager-project agent.

### Configuration Management

[HARD] .moai/config/config.yaml settings ALWAYS take priority over defaults.

WHY: Existing configuration represents user intent.

IMPACT: Ignoring existing config causes destructive overwrites.

[SOFT] Respect existing language settings unless user explicitly requests change via SETTINGS Tab 1.

WHY: Unexpected language changes disrupt user workflow.

IMPACT: Automatic changes without user consent reduce trust.

[SOFT] For fresh installs: language selection occurs FIRST before other configuration.

WHY: Language choice affects all subsequent interactions.

IMPACT: Deferred language selection complicates initial setup.

---

##  Output Format

Responses and status reports must follow structured XML format for clarity and automated processing:

<analysis>
Context assessment including detected mode, language context, and user command arguments.
Example: "Detected AUTO-DETECT mode, user language is Korean (ko), existing config found at .moai/config/config.yaml"
</analysis>

<approach>
Execution strategy selected based on analysis, including manager-project agent invocation parameters.
Example: "Delegating to manager-project agent with AUTO-DETECT mode context, language: ko, requesting settings review"
</approach>

<phase>
Current execution phase (Phase 1, Phase 2, Phase 2.5, or Phase 3) with step-by-step progress.
Example: "Phase 2: Execute AUTO-DETECT Mode - Displaying current configuration and offering modification options"
</phase>

<verification>
Quality checks and validation results at checkpoints, including language validation and configuration consistency checks.
Example: "Language validation passed (ko is valid), configuration loaded successfully, ready for user interaction"
</verification>

<completion>
Mode-specific completion summary with files created, settings modified, and next recommended action.
Example: "Settings updated successfully. Modified 4 fields: user.name, conversation_language, language_name [auto], agent_prompt_language. Recommend next step: /moai:1-plan"
</completion>

WHY: Structured output enables automated parsing and consistent status tracking across command executions.

IMPACT: Unstructured output reduces ability to track execution state and causes user confusion about command progress.

---

##  Quick Reference

| Scenario             | Mode           | Entry Point                       | Key Phases                                                     |
| -------------------- | -------------- | --------------------------------- | -------------------------------------------------------------- |
| First-time setup     | INITIALIZATION | `/moai:0-project` (no config)     | Read language → Interview → Docs                               |
| Existing project     | AUTO-DETECT    | `/moai:0-project` (config exists) | Read language → Display → Options                              |
| Modify config        | SETTINGS       | `/moai:0-project setting`         | Interactive tab selection → Conditional batches → Skill update |
| After package update | UPDATE         | `/moai:0-project update`          | Preserve language → Template merge → Announce                  |

Associated Skills:

- moai-workflow-project - Unified project management skill providing:
  - Language selection/change
  - Config operations (atomic updates, backup/rollback)
  - Template merging
  - Tab-based batch questions

Project Documentation Directory:

- Location: `.moai/project/` (singular, NOT `.moai/projects/`)
- Files: `product.md`, `structure.md`, `tech.md` (auto-generated or interactive)
- Language: Auto-translated to user's conversation_language

Version: 2.0.0 (Tab-based Configuration with Conditional Batches & Fixed Field Alignment)
Last Updated: 2025-11-19
Architecture: Commands → Agents → Skills (Complete delegation, no direct backup in command)
Tab Schema: Available in moai-workflow-project skill (v2.0.0)
Improvements in v2.0.0:

- Removed `[tab_ID]` argument → Always use interactive tab selection
- Added git_strategy.mode selection (Batch 3.0) with Personal/Team/Hybrid conditional logic
- Expanded Tab 3: 16 → 29 settings (+81%)
- Fixed 26 outdated/incorrect field names (checkpoint_enabled→auto_checkpoint, etc)
- Enhanced validation checkpoints: 3 → 6 rules
- Total coverage: 41 → 57 settings (+39%)

---

## Final Step: Next Action Selection

After command execution completes, use AskUserQuestion tool to guide user to next action:

```python
AskUserQuestion({
    "questions": [{
        "question": "Project setup is complete. What would you like to do next?",
        "header": "Next Steps",
        "multiSelect": false,
        "options": [
            {
                "label": "Write Specification",
                "description": "Execute /moai:1-plan to define feature specifications"
            },
            {
                "label": "Review Project Structure",
                "description": "Check current project status and settings"
            },
            {
                "label": "Start New Session",
                "description": "Initialize workspace and start fresh"
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

You must NOW execute the command following the "Execution Philosophy" described above.

1. Analyze the subcommand/context.
2. Use the manager-project subagent to handle project setup.
3. Do NOT just describe what you will do. DO IT.
