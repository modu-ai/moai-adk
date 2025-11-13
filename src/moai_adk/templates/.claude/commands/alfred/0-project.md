---
name: alfred:0-project
description: "Initialize project metadata and documentation"
argument-hint: "[setting [tab_ID]|update]"
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

Initialize or update project metadata with **language-first architecture**. Supports four execution modes:
- **INITIALIZATION**: First-time project setup
- **AUTO-DETECT**: Already initialized projects (modify settings or re-initialize)
- **SETTINGS**: Tab-based configuration management (new mode)
- **UPDATE**: Template optimization after moai-adk package update

---

## 🧠 Associated Agents & Skills

| Agent/Skill | Purpose |
|---|---|
| project-manager | Orchestrates language-first initialization and configuration |
| moai-project-language-initializer | Language selection and initialization workflows |
| moai-project-config-manager | Configuration management with language context |
| moai-project-batch-questions | Standardizes user interaction patterns with tab-based system |

---

## 🌐 Language-First Architecture

**Core Principle**: Language selection ALWAYS happens BEFORE any other configuration.

- **Initialization**: Language selection → Project interview → Documentation
- **Auto-Detect**: Language confirmation → Settings options
- **Settings**: Language context → Tab-based configuration
- **Update**: Language confirmation → Template optimization

---

## 💡 Execution Philosophy: "Plan → Configure → Complete"

`/alfred:0-project` performs project setup through complete agent delegation:

```
User Command: /alfred:0-project [setting]
    ↓
/alfred:0-project Command
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

1. **`/alfred:0-project setting [tab_ID]`** → SETTINGS MODE
   - Tab ID examples: `tab_1_user_language`, `tab_2_project_info`, `tab_3_git_strategy`, etc
   - Omit tab_ID for interactive tab selection
2. **`/alfred:0-project update`** → UPDATE MODE
3. **`/alfred:0-project`** (no args):
   - Check if `.moai/config/config.json` exists
   - Exists → AUTO-DETECT MODE
   - Missing → INITIALIZATION MODE
4. **Invalid subcommand** → Show error and exit

### Step 2: Delegate to Project Manager Agent

Use Task tool:
- `subagent_type`: "project-manager"
- `description`: "Route and analyze project setup request"
- `prompt`:
  ```
  You are the project-manager agent.

  **Task**: Analyze project context and route to appropriate mode.

  **Detected Mode**: $MODE (INITIALIZATION/AUTO-DETECT/SETTINGS/UPDATE)
  **Language Context**: Determine from .moai/config.json if exists

  **For INITIALIZATION**:
  - Invoke Skill("moai-project-language-initializer", mode="language_first")
  - Conduct language-aware user interview
  - Generate project documentation
  - Invoke Skill("moai-project-config-manager") for config creation

  **For AUTO-DETECT**:
  - Confirm current language settings
  - If "Change Language" → Invoke Skill("moai-project-language-initializer", mode="language_change_only")
  - Display current configuration
  - Offer: Modify Settings / Review Configuration / Re-initialize / Cancel

  **For SETTINGS**:
  - Load tab schema from .claude/skills/moai-project-batch-questions/tab_schema.json
  - Confirm language context first
  - Execute batch questions via moai-project-batch-questions skill
  - Process responses and update config.json atomically
  - Report changes and validation results

  **For UPDATE**:
  - Confirm language context
  - Invoke Skill("moai-project-template-optimizer") for smart merging
  - Update templates and configuration
  - Auto-translate announcements to current language

  **Output**: Mode-specific completion report with next steps
  ```

**Store**: Response in `$MODE_EXECUTION_RESULT`

---

## 🔧 PHASE 2: Execute Mode

**Goal**: Execute the appropriate mode based on routing decision.

### Mode Handler: project-manager Agent

The project-manager agent handles all mode-specific workflows:

**INITIALIZATION MODE**:
- Language-first user interview (via Skill)
- Project type detection and configuration
- Documentation generation
- Auto-translate announcements to selected language

**AUTO-DETECT MODE**:
- Language confirmation
- Display current configuration
- Offer: Modify Settings / Review Configuration / Re-initialize / Cancel
- Route to selected sub-action

**SETTINGS MODE** (NEW):
- Language confirmation
- Load tab schema for batch-based questions
- Execute batch questions with AskUserQuestion
- Process user responses
- Validate settings at critical checkpoints
- Update `.moai/config/config.json` atomically
- Report changes

**UPDATE MODE**:
- Analyze backup and compare templates
- Perform smart template merging
- Update `.moai/` files with new features
- Auto-translate announcements to current language

### Language-Aware Announcements

After any language selection or change, auto-translate company announcements:
```bash
uv run $CLAUDE_PROJECT_DIR/.claude/hooks/alfred/shared/utils/announcement_translator.py
```

This ensures `.claude/settings.json` contains announcements in the user's selected language.

---

## 🎭 SETTINGS MODE: Tab-Based Configuration (NEW)

> **Version**: v2.1.0 | **Last Updated**: 2025-11-13 | **Changes**: Tab-based UX improvements, auto-processing for locale/language

### Overview

The SETTINGS MODE uses a tab-based batch question system to provide organized, user-friendly configuration management:

- **5 tabs**: Organized by configuration domain
- **12 batches**: Grouped questions within tabs
- **41 settings**: Complete config.json coverage (down from 44 via auto-processing)
- **40 questions**: User-facing questions (down from 43)
- **Atomic updates**: Safe deep merge with backup/rollback

### Initial Entry Point: Tab Selection Screen

When user runs `/alfred:0-project setting` (without tab_ID), present tab selection:

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

**Tab 3: Git Strategy & Workflow** (Recommended with Validation)
- Batch 3.1: Personal checkpoint settings (4 questions)
- Batch 3.2: Personal commit/branch settings (4 questions)
- Batch 3.3: Personal policy & Team PR (4 questions)
- Batch 3.4: Team GitFlow policy (4 questions)
- Setting count: 16 | Critical checkpoint for Git conflicts

**Tab 4: Quality Principles & Reports** (Optional)
- Batch 4.1: Constitution settings (4 questions)
- Batch 4.2: Report generation policy (4 questions)
- Batch 4.3: Report storage location (1 question)
- Setting count: 9

**Tab 5: System & GitHub Integration** (Optional)
- Batch 5.1: MoAI system settings (4 questions)
- Batch 5.2: GitHub automation settings (3 questions)
- Setting count: 8

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

#### Step 5: Atomic Config Update

**Update Pattern** (Safe deep merge):

```markdown
Step 1: Load current config.json
Step 2: Create backup: config.json.backup-{timestamp}
Step 3: Deep merge user updates into current config
  - Preserve existing settings not in update
  - Recursively merge nested objects
  - Validate final config structure
Step 4: Write updated config.json atomically
Step 5: Verify write success
  - If success: Delete backup, report completion
  - If failure: Restore from backup, report error
```

**Backup/Rollback Strategy**:
```markdown
Success flow:
  config.json.backup → (deleted after verification)

Error flow:
  config.json.backup → (restored as config.json)
  Report: "Configuration update failed, rolled back to previous version"
```

### Implementation Details

#### Tab 1 Execution Example

User runs: `/alfred:0-project setting tab_1_user_language`

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
Step 9: Creates atomic update
  - Backup current config
  - Deep merge updates (including auto-updated conversation_language_name)
  - Verify final structure
Step 10: Write updated config.json
Step 11: Report success and changes made (4 fields: user.name, conversation_language, conversation_language_name [auto], agent_prompt_language)
```

#### Tab 3 Validation Example (Complex)

User runs: `/alfred:0-project setting tab_3_git_strategy`

```
Step 1: Load Tab 3 (tab_3_git_strategy) - 4 batches
Step 2: Execute Batch 3.1 (Personal checkpoint settings)
  - Get user responses, validate
Step 3: Execute Batch 3.2 (Personal commit/branch)
  - Get user responses, validate
Step 4: Execute Batch 3.3 (Personal policy & Team PR)
  - Get user responses, validate
Step 5: Execute Batch 3.4 (Team GitFlow policy)
  - Get user responses, validate
Step 6: Run Tab 3 validation checkpoint
  - Check for Personal/Team conflicts
  - Example: If Personal mode but PR base is develop → Warn
  - Example: If Team mode but use_gitflow is false → Suggest fix
  - Let user confirm or retry
Step 7: Merge all 4 batches into single update object
Step 8: Create atomic update (backup + deep merge)
Step 9: Report all 16 settings changes
```

#### Multi-Tab Workflow Example

User runs: `/alfred:0-project setting` (without tab_ID) → Tab Selection Screen

```
Flow:
1. Show Tab Selection Screen (Which settings tab would you like to modify?)
2. User selects tab or "Modify All Tabs"
3. Execute selected tab
   - Tab 1 (REQUIRED): User & Language (3 questions)
   - Tab 2 (RECOMMENDED): Project Info (4 questions in batch 2.1 + auto-processing in batch 2.2)
   - Tab 3 (RECOMMENDED): Git Strategy (4 batches, 16 questions with validation)
   - Tab 4 (OPTIONAL): Quality & Reports (3 batches, 9 questions)
   - Tab 5 (OPTIONAL): System & GitHub (2 batches, 7 questions)
4. After tab completion, ask: "Would you like to modify another settings tab?"
   - No, finish settings (exit)
   - Select another tab (select another tab)
5. Final atomic update after user finishes

Each tab completes independently:
  - If user cancels mid-tab, changes not saved
  - If tab validation fails, user can retry
  - Final atomic update only after ALL selected tabs complete
  - Auto-processing happens during atomic update (e.g., conversation_language_name, locale)
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
- ALWAYS confirm language context before starting SETTINGS MODE
- Run validation at Tab 1, Tab 3, and before final update
- Create atomic config update with backup/rollback support
- Report all changes made
- Use AskUserQuestion for ALL user interaction

**Configuration Priority**:
- `.moai/config/config.json` settings ALWAYS take priority
- Existing language settings respected unless user requests change
- Fresh installs: Language selection FIRST (Tab 1), then all other config

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

| Scenario | Mode | Entry Point | Key Phases |
|---|---|---|---|
| First-time setup | INITIALIZATION | `/alfred:0-project` (no config) | Language → Interview → Docs |
| Existing project | AUTO-DETECT | `/alfred:0-project` (config exists) | Language → Display → Options |
| Modify config | SETTINGS | `/alfred:0-project setting [tab]` | Language → Tab batches → Atomic update |
| After package update | UPDATE | `/alfred:0-project update` | Language → Template merge → Announce |

**Associated Skills**:
- `Skill("moai-project-language-initializer")` - Language selection
- `Skill("moai-project-config-manager")` - Config operations
- `Skill("moai-project-template-optimizer")` - Template merging
- `Skill("moai-project-batch-questions")` - Tab-based batch questions

**Version**: 1.1.0 (Tab-Based SETTINGS MODE v2.0.0)
**Last Updated**: 2025-11-12
**Architecture**: Commands → Agents → Skills (Complete delegation)
**Tab Schema**: `.claude/skills/moai-project-batch-questions/tab_schema.json`

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
                "description": "Execute /alfred:1-plan to define feature specifications"
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
