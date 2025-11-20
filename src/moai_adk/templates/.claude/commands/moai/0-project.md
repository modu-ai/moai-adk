---
name: moai:0-project
description: "Initialize project metadata and documentation"
argument-hint: "[<empty>|setting|update|--glm-on <token>]"
allowed-tools:
  - Task
  - AskUserQuestion
---

# ⚒️ MoAI-ADK Step 0: Initialize/Update Project (Project Setup)

> **Interactive Prompts**: Use `AskUserQuestion` tool for TUI-based user interaction.
> **Architecture**: Commands → Agents → Skills. This command orchestrates ONLY through `Task()` tool.
> **Delegation Model**: Complete agent-first pattern. All execution delegated to project-manager.

**4-Step Workflow Integration**: This command implements Step 0 of Alfred's workflow (Project Bootstrap). See CLAUDE.md for full workflow details.

---

## 🎯 Command Purpose

Initialize or update project metadata with **language-first architecture**. Supports five execution modes:

- **INITIALIZATION**: First-time project setup
- **AUTO-DETECT**: Already initialized projects (modify settings or re-initialize)
- **SETTINGS**: Tab-based configuration management
- **UPDATE**: Template optimization after moai-adk package update
- **GLM Configuration** (`--glm-on <token>`): Configure GLM API integration

---

## 🧠 Associated Agents & Skills

| Agent/Skill                       | Purpose                                                      |
| --------------------------------- | ------------------------------------------------------------ |
| project-manager                   | Orchestrates language-first initialization and configuration |
| moai-project-language-initializer | Language selection and initialization workflows              |
| moai-project-config-manager       | Configuration management with language context               |
| moai-project-batch-questions      | Standardizes user interaction patterns with tab-based system |

---

## 🌐 Language Configuration (Pre-set by moai-adk CLI)

**Core Principle**: Language is already configured by `moai-adk init` or `moai-adk update` CLI commands.

- `/moai:0-project` reads language from `.moai/config/config.json`
- Language change only when explicitly requested (SETTINGS mode, Tab 1)

- **Initialization**: Read language from config → Project interview → Documentation
- **Auto-Detect**: Display current language → Settings options (with language change shortcut)
- **Settings**: Display current language in Tab 1 → Optional language change
- **Update**: Preserve language from config → Template optimization

---

## 💡 Execution Philosophy: "Plan → Configure → Complete"

`/moai:0-project` performs project setup through complete agent delegation:

```
User Command: /moai:0-project [setting]
    ↓
/moai:0-project Command
    └─ Task(subagent_type="project-manager")
        ├─ Phase 1: Route and analyze
        ├─ Phase 2: Execute mode (INIT/AUTO-DETECT/SETTINGS/UPDATE)
        ├─ Phase 2.5: Save phase context
        └─ Phase 3: Completion and next steps
            ↓
        Output: Project configured with language-first principles
```

### Key Principle: Zero Direct Tool Usage

**This command uses ONLY Task() and AskUserQuestion():**

- ❌ No Read (file operations delegated)
- ❌ No Write (file operations delegated)
- ❌ No Edit (file operations delegated)
- ❌ No Bash (all bash commands delegated)
- ❌ No TodoWrite (delegated to project-manager)
- ✅ **Task()** for orchestration
- ✅ **AskUserQuestion()** for user interaction

All complexity is handled by the **project-manager** agent.

---

## 🚀 PHASE 1: Command Routing & Analysis

**Goal**: Detect subcommand and prepare execution context.

### Step 1: Route Based on Subcommand

Analyze the command user provided:

1. **`/moai:0-project --glm-on [api-token]`** → GLM CONFIGURATION MODE
   - Detect if token provided in argument
   - If token missing: Check `.env.glm` (auto-load if exists)
   - If token missing: Check `ANTHROPIC_AUTH_TOKEN` environment variable
   - If all missing: Request token from user
   - Delegate to project-manager with GLM context
   - Call setup-glm.py script to configure GLM

2. **`/moai:0-project setting`** → SETTINGS MODE
   - Always uses interactive tab selection via AskUserQuestion
   - User selects specific tab or "Modify All Tabs" option

3. **`/moai:0-project update`** → UPDATE MODE

4. **`/moai:0-project`** (no args):
   - Check if `.moai/config/config.json` exists
   - Exists → AUTO-DETECT MODE
   - Missing → INITIALIZATION MODE

5. **Invalid subcommand** → Show error and exit

### Step 2: Delegate to Project Manager Agent

Use Task tool:

- `subagent_type`: "project-manager"
- `description`: "Route and analyze project setup request"
- `prompt`:

  ```
  You are the project-manager agent.

  **Task**: Analyze project context and route to appropriate mode.

  **Detected Mode**: $MODE (INITIALIZATION/AUTO-DETECT/SETTINGS/UPDATE/GLM_CONFIGURATION)
  **Language Context**: Read from .moai/config.json if exists
  **GLM Token** (if GLM mode): $GLM_TOKEN

  **For INITIALIZATION**:
  - Check .moai/config.json for language setting
  - If missing: Invoke Skill("moai-project-language-initializer", mode="language_first")
  - If present: Use existing language, skip language selection
  - Conduct language-aware user interview
  - Generate project documentation
  - Invoke Skill("moai-project-config-manager") for config creation

  **For AUTO-DETECT**:
  - Read current language from .moai/config.json
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

  **For SETTINGS**:
  - Load current language from .moai/config.json
  - Load tab schema from .claude/skills/moai-project-batch-questions/tab_schema.json
  - Execute batch questions via moai-project-batch-questions skill
  - Process responses and update config.json atomically via Skill("moai-project-config-manager")
  - Report changes and validation results

  **For UPDATE**:
  - Read language from config backup (preserve existing setting)
  - Invoke Skill("moai-project-template-optimizer") for smart merging
  - Update templates and configuration
  - Auto-translate announcements to current language if needed

  **For GLM_CONFIGURATION**:
  - Receive GLM API token from parameter (or detect from environment)
  - Check token resolution sequence:
    1. Use provided token from `--glm-on <token>` argument (if not empty)
    2. Auto-load from existing `.env.glm` file (if exists and token missing)
    3. Auto-load from `ANTHROPIC_AUTH_TOKEN` environment variable (if set)
    4. Request from user via AskUserQuestion (if all above missing)
  - Execute GLM setup script: `uv run .moai/scripts/setup-glm.py <GLM_TOKEN>`
  - Verify configuration in .claude/settings.local.json:
    * ANTHROPIC_AUTH_TOKEN: <api_token> (stored in "env" section)
    * ANTHROPIC_BASE_URL: https://api.z.ai/api/anthropic
    * ANTHROPIC_DEFAULT_HAIKU_MODEL: glm-4.5-air
    * ANTHROPIC_DEFAULT_SONNET_MODEL: glm-4.6
    * ANTHROPIC_DEFAULT_OPUS_MODEL: glm-4.6
  - Verify .env.glm created with secure permissions (0o600)
  - Verify .gitignore includes .env.glm entry
  - Report GLM configuration success to user with all configured keys
  - Remind user: "Claude Code를 재시작하면 새 설정이 자동으로 로드됩니다"

  **Output**: Mode-specific completion report with next steps
  ```

**Store**: Response in `$MODE_EXECUTION_RESULT`

---

## 🔧 PHASE 2: Execute Mode

**Goal**: Execute the appropriate mode based on routing decision.

### Mode Handler: project-manager Agent

The project-manager agent handles all mode-specific workflows:

**INITIALIZATION MODE**:

- Read language from config.json (or use CLI default if missing)
- Conduct language-aware user interview (via Skill)
- Project type detection and configuration
- Documentation generation
- Auto-translate announcements to selected language

**AUTO-DETECT MODE**:

- Read current language from config.json
- **CRITICAL CHECK**: Detect partial initialization state
  - Check if project documentation exists in `.moai/project/`:
    * product.md, structure.md, tech.md
  - If ANY doc missing → Use AskUserQuestion (in user's language)
    * Question: "Your configuration exists but project documentation is missing. Would you like to complete the initialization?"
    * Options: "Yes, complete initialization" / "No, review configuration" / "Cancel"
    * If "Yes" → Switch to INITIALIZATION workflow
- Display current configuration (including language, initialization status)
- Offer: Modify Settings / Change Language Only / Review Configuration / Re-initialize / Cancel
- "Change Language Only" shortcut → SETTINGS mode, Tab 1 only
- Route to selected sub-action
- **Language-Aware**: All AskUserQuestion calls in user's conversation_language (NO EMOJIS)

**SETTINGS MODE** (NEW):

- Read current language from config.json
- Load tab schema for batch-based questions
- Execute batch questions with AskUserQuestion
- Process user responses
- Validate settings at critical checkpoints
- Delegate config update to `moai-project-config-manager` Skill
- Report changes

**UPDATE MODE**:

- Preserve language from config backup
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

## 🎭 SETTINGS MODE: Tab-Based Configuration (NEW)

> **Version**: v2.0.0 | **Last Updated**: 2025-11-19 | **Changes**: Removed [tab_ID] arg, added git_strategy.mode selection, expanded Tab 3 with conditional batches, fixed 26 field name errors, +16 settings

### Overview

The SETTINGS MODE uses a tab-based batch question system to provide organized, user-friendly configuration management:

- **5 tabs**: Organized by configuration domain
- **17 batches**: Grouped questions within tabs (added 5 batches: Batch 3.0, 3.3, 3.5, 3.6, improved organization)
- **57 settings**: Complete config.json v0.26.0 coverage (+39% from v1.0.0)
- **54 questions**: User-facing questions (+14 from v1.0.0)
- **Conditional batches**: Tab 3 shows Personal/Team/Hybrid batches based on mode selection
- **Atomic updates**: Safe deep merge with backup/rollback

### Initial Entry Point: Tab Selection Screen

When user runs `/moai:0-project setting` (without tab_ID), present tab selection:

```markdown
Which settings tab would you like to modify?

Options:

1. Tab 1: User & Language

   - Configure user name, conversation language, agent prompt language

2. Tab 2: Project Basic Information

   - Configure project name, description, owner, mode

3. Tab 3: Git Strategy & Workflow

   - Configure Personal/Team Git settings, commit/branch strategy

4. Tab 4: Quality Principles & Reports

   - Configure TRUST 5, report generation, storage location

5. Tab 5: System & GitHub Integration

   - Configure MoAI system, GitHub automation

6. Modify All Tabs
   - Recommended: Tab 1 → Tab 2 → Tab 3 → Others
```

**After Tab Completion**:

```markdown
Would you like to modify another settings tab?

1. No, finish settings
2. Select another tab
```

### Tab Schema Reference

Location: `.claude/skills/moai-project-batch-questions/tab_schema.json`

**Tab 1: User & Language** (Required Foundation)

- Batch 1.1: Basic settings (3 questions - UPDATED: removed conversation_language_name)
  - User name, conversation language, agent prompt language
  - NOTE: conversation_language_name is auto-updated when conversation_language changes
- Setting count: 3 | Critical checkpoint

**Tab 2: Project Basic Information** (Recommended)

- Batch 2.1: Project metadata (4 questions)
  - Project name, description, owner, mode
- Batch 2.2: Auto-processed locale settings (0 questions - UPDATED: internal analysis only)
  - project.locale, default_language, optimized_for_language (auto-determined from conversation_language)
  - NOTE: No user input needed. These 3 fields update automatically when conversation_language changes
- Setting count: 4

**Tab 3: Git Strategy & Workflow** (Recommended with Validation - REDESIGNED v2.0.0)

- Batch 3.0: Workflow mode selection (1 question - Personal/Team/Hybrid) → Controls visibility of subsequent batches
- Batch 3.1: Personal core settings (4 questions) - CONDITIONAL (Personal/Hybrid only)
- Batch 3.2: Personal branch & cleanup (4 questions) - CONDITIONAL (Personal/Hybrid only)
- Batch 3.3: Personal protection & merge (4 questions) - CONDITIONAL (Personal/Hybrid only)
- Batch 3.4: Team core settings (4 questions) - CONDITIONAL (Team/Hybrid only)
- Batch 3.5: Team branch & protection (4 questions) - CONDITIONAL (Team/Hybrid only)
- Batch 3.6: Team safety & merge (2 questions) - CONDITIONAL (Team/Hybrid only)
- Setting count: 29 (+13 from v1.0.0) | Critical checkpoint for Git conflicts & mode consistency

**Tab 4: Quality Principles & Reports** (Optional - UPDATED v2.0.0)

- Batch 4.1: Constitution settings (3 questions - reduced from 4, renamed minimum_test_coverage→test_coverage_target)
- Batch 4.2: Report generation policy (4 questions - expanded, added warn_user & user_choice)
- Setting count: 9 (same count, better fields)

**Tab 5: System & GitHub Integration** (Optional - UPDATED v2.0.0)

- Batch 5.1: MoAI system settings (3 questions - updated, aligned with config.json v0.26.0)
- Batch 5.2: GitHub automation settings (5 questions - expanded from 3, added templates & spec_workflow fields)
- Setting count: 11 (+3 from v1.0.0)

### Batch Execution Flow

#### Step 1: Load Tab Schema

```markdown
Load: .claude/skills/moai-project-batch-questions/tab_schema.json
Extract:

- Tab definition (label, batches)
- Batch questions (max 4 per batch)
- Field mappings to config.json paths
- Current values from existing config
- Validation rules
```

#### Step 2: Execute Batch via AskUserQuestion

**Single Batch Execution Example** (Tab 1, Batch 1.1):

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
{label: "Korean (ko)", description: "All content will be generated in Korean"},
{label: "English (en)", description: "All content will be generated in English"},
{label: "Japanese (ja)", description: "All content will be generated in Japanese"},
{label: "Spanish (es)", description: "All content will be generated in Spanish"}
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

Wait for user responses, then process each response into config update:
user.name → user_input_or_keep_current
language.conversation_language → selected_value
language.conversation_language_name → user_input_or_keep_current
language.agent_prompt_language → selected_value
```

#### Step 3: Process Responses

**Mapping Logic**:

```markdown
For each question in batch:

1. Get field path from schema (e.g., "user.name")
2. Get user's response (selected option or custom input)
3. Convert to config.json value:
   - "Other" option → Use custom input from user
   - Selected option → Use option's mapped value
   - "Keep current" → Use existing value
4. Build update object: {field_path: new_value}
5. Collect all updates from batch
```

#### Step 4: Validate at Checkpoints

**Checkpoint Locations** (from tab_schema navigation_flow):

1. **After Tab 1** (Language settings):

   - Verify conversation_language is valid (ko, en, ja, es, etc)
   - Verify agent_prompt_language consistency
   - Error recovery: Re-ask Tab 1 if validation fails

2. **After Tab 3** (Git strategy):

   - Validate Personal/Team mode conflicts
     - If Personal: main_branch should not be "develop"
     - If Team: PR base must be develop or main (never direct to main)
   - Validate branch naming consistency
   - Error recovery: Highlight conflicts, offer fix suggestions

3. **Before Config Update** (Final validation):
   - Check all required fields are set (marked required: true in schema)
   - Verify no conflicting settings
   - Validate field value types (string, bool, number, array)
   - Report validation results to user

#### Step 5: Delegate Atomic Config Update to Skill

**Update Pattern** (Skill-delegated):

```markdown
Delegate ALL config update operations to Skill("moai-project-config-manager"):

- Skill handles backup/rollback logic internally
- Skill performs deep merge with validation
- Skill writes atomically to config.json
- Skill reports success/failure

Agent responsibilities:

- Collect user responses from AskUserQuestion
- Map responses to config field paths
- Pass update map to Skill
- Report results to user
```

**Skill Responsibilities**:

- Skill("moai-project-config-manager") handles ALL file operations
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
Step 4: Loads current values from config.json
  - user.name: "GoosLab"
  - language.conversation_language: "ko"
  - language.agent_prompt_language: "ko"
Step 5: Calls AskUserQuestion with 3 questions (UPDATED: removed language_display_name)
  - Question 1: "The user name is currently set to 'GoosLab'. Is this correct?"
  - Question 2: "What language should Alfred use in conversations? (current: Korean/ko)"
  - Question 3: "The agent internal prompt language is currently set to Korean(ko). How would you like to configure this?"
Step 6: Receives user responses
Step 7: Processes each response (map to config fields)
  - user.name response → user.name
  - conversation_language response → language.conversation_language
  - Auto-update: conversation_language_name (ko → Korean, en → English, ja → Japanese, es → Spanish)
  - agent_prompt_language response → language.agent_prompt_language
Step 8: Runs Tab 1 validation checkpoint
  - Check language is valid
  - Verify consistency
Step 9: Delegates atomic update to Skill("moai-project-config-manager")
  - Skill handles backup/rollback internally
  - Skill performs deep merge (including auto-updated conversation_language_name)
  - Skill verifies final structure
Step 10: Receives result from Skill
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
Step 6: Delegate atomic update to Skill("moai-project-config-manager")
  - Skill handles backup/rollback internally
  - Skill performs deep merge with final validation

Step 7: Report all 29 settings changes (or 16-20 depending on mode)
```

#### Multi-Tab Workflow Example

User runs: `/moai:0-project setting` (without tab_ID) → Tab Selection Screen

```
Flow:
1. Show Tab Selection Screen (Which settings tab would you like to modify?)
2. User selects tab or "Modify All Tabs"
3. Execute selected tab
   - Tab 1 (REQUIRED): User & Language (1 batch, 3 questions)
   - Tab 2 (RECOMMENDED): Project Info (2 batches, 4 questions in batch 2.1 + 0 questions auto-processing in batch 2.2)
   - Tab 3 (RECOMMENDED): Git Strategy (6 batches, 23 questions total, conditional by mode)
     * Batch 3.0: Mode selection (1 question)
     * Personal mode: Batches 3.1-3.3 (12 questions)
     * Team mode: Batches 3.4-3.6 (10 questions)
     * Hybrid mode: All batches (22 questions)
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
  - Tab 3: Git workflow strategy
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

**MANDATORY**:

- Execute ONLY ONE tab per command invocation (unless user specifies "all tabs")
- READ language context from config.json before starting SETTINGS MODE
- Run validation at Tab 1, Tab 3, and before final update
- Delegate config update to `moai-project-config-manager` Skill (no direct backup in command)
- Report all changes made
- Use AskUserQuestion for ALL user interaction

**Configuration Priority**:

- `.moai/config/config.json` settings ALWAYS take priority
- Existing language settings respected unless user explicitly requests change in Tab 1
- Fresh installs: Language already set by moai-adk CLI, skip language selection

**Language**:

- Tab schema stored in English (technical field names)
- All user-facing questions in user's conversation_language
- AskUserQuestion must use user's conversation_language for ALL fields

---

## 💾 PHASE 2.5: Save Phase Context

**Goal**: Persist phase execution results for explicit context passing to subsequent commands.

### Step 1: Extract Context from Agent Response

After project-manager agent completes, extract the following information:

- **Project metadata**: name, mode, owner, language
- **Files created**: List of generated files with absolute paths
- **Tech stack**: Primary codebase language
- **Next phase**: Recommended next command (1-plan)

### Step 2: Delegate Context Saving to project-manager

The project-manager agent handles all context saving:

```markdown
Context data to persist:

- Phase: "0-project"
- Mode: INITIALIZATION|AUTO-DETECT|SETTINGS|UPDATE
- Timestamp: ISO8601 UTC
- Status: completed|failed
- Outputs:
  - project_name
  - mode (personal|team)
  - language (conversation_language)
  - tech_stack (detected primary language)
- Files created: [list of absolute paths]
- Next phase: "1-plan"

Agent delegates to Skill("moai-project-config-manager"):

- Save context via ContextManager
- Handle file path validation
- Implement error recovery (non-blocking)
- Report success/failure
```

**Error Handling Strategy**:

- Context save failures should NOT block command completion
- Log clear warning messages for debugging
- Allow user to retry manually if needed

---

## 🔒 PHASE 3: Completion & Next Steps

**Goal**: Guide user to next action in their selected language.

### Step 1: Display Completion Status

Show mode-specific completion message in user's language:

- **INITIALIZATION**: "Project initialization complete"
- **AUTO-DETECT**: Configuration review/modification complete
- **SETTINGS**: "Settings updated successfully"
- **UPDATE**: "Templates optimized and updated"

### Step 2: Offer Next Steps

Use AskUserQuestion in user's language:

- **From Initialization**: Write SPEC / Review Structure / New Session
- **From Settings**: Continue Settings / Sync Documentation / Exit
- **From Update**: Review Changes / Modify Settings / Exit

**Critical**: NO EMOJIS in AskUserQuestion fields. Use clear text only.

---

## 📋 Critical Rules

**MANDATORY**:

- Execute ONLY ONE mode per invocation
- Never skip language confirmation/selection
- Always use user's `conversation_language` for all output
- Auto-translate announcements after language changes
- Route to correct mode based on command analysis
- Delegate ALL execution to project-manager agent
- Use AskUserQuestion for ALL user interaction
- NO EMOJIS in AskUserQuestion fields

**No Direct Tool Usage**:

- ❌ NO Read (file operations)
- ❌ NO Write (file operations)
- ❌ NO Edit (file operations)
- ❌ NO Bash (delegated to agents)
- ❌ NO TodoWrite (delegated to agents)
- ✅ ONLY Task() and AskUserQuestion()

**Configuration Priority**:

- `.moai/config/config.json` settings ALWAYS take priority
- Existing language settings respected unless user requests change
- Fresh installs: Language selection FIRST, then all other config

---

## 📚 Quick Reference

| Scenario             | Mode           | Entry Point                       | Key Phases                                                     |
| -------------------- | -------------- | --------------------------------- | -------------------------------------------------------------- |
| First-time setup     | INITIALIZATION | `/moai:0-project` (no config)     | Read language → Interview → Docs                               |
| Existing project     | AUTO-DETECT    | `/moai:0-project` (config exists) | Read language → Display → Options                              |
| Modify config        | SETTINGS       | `/moai:0-project setting`         | Interactive tab selection → Conditional batches → Skill update |
| After package update | UPDATE         | `/moai:0-project update`          | Preserve language → Template merge → Announce                  |

**Associated Skills**:

- `Skill("moai-project-language-initializer")` - Language selection/change
- `Skill("moai-project-config-manager")` - Config operations (atomic updates, backup/rollback)
- `Skill("moai-project-template-optimizer")` - Template merging
- `Skill("moai-project-batch-questions")` - Tab-based batch questions

**Project Documentation Directory**:

- **Location**: `.moai/project/` (singular, NOT `.moai/projects/`)
- **Files**: `product.md`, `structure.md`, `tech.md` (auto-generated or interactive)
- **Language**: Auto-translated to user's conversation_language

**Version**: 2.0.0 (Tab-based Configuration with Conditional Batches & Fixed Field Alignment)
**Last Updated**: 2025-11-19
**Architecture**: Commands → Agents → Skills (Complete delegation, no direct backup in command)
**Tab Schema**: `.claude/skills/moai-project-batch-questions/tab_schema.json` (v2.0.0)
**Improvements in v2.0.0**:

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

**Important**:

- Use conversation language from config
- No emojis in any AskUserQuestion fields
- Always provide clear next step options

## ⚡️ EXECUTION DIRECTIVE

**You must NOW execute the command following the "Execution Philosophy" described above.**

1. Analyze the subcommand/context.
2. Call the `Task` tool with `subagent_type="project-manager"` immediately.
3. Do NOT just describe what you will do. DO IT.
