---
name: moai:0-project
description: "Initialize project metadata and documentation"
argument-hint: "[<empty>|setting|update|--glm-on <token>]"
allowed-tools:
  - Task
  - AskUserQuestion
model: inherit
skills:
  - moai-project-language-initializer
  - moai-project-config-manager
  - moai-project-batch-questions
---

## Pre-execution Context

!git status --porcelain
!git config --get user.name
!git config --get user.email
!git branch --show-current

## Essential Files

@.moai/config/config.json
@.moai/project/product.md
@.moai/project/structure.md
@.moai/project/tech.md

---

# MoAI-ADK Step 0: Initialize/Update Project (Project Setup)

Commands → Agents → Skills. All execution delegated to project-manager agent.

**Critical Rule**: NO EMOJIS in AskUserQuestion (question, header, options fields).

**4-Step Workflow Integration**: Step 0 of Alfred's workflow (Project Bootstrap).

---

## Command Purpose

Initialize or update project metadata with language-first architecture. Five execution modes:

- **INITIALIZATION**: First-time project setup
- **AUTO-DETECT**: Existing initialized projects with modification/re-initialization options
- **SETTINGS**: Tab-based configuration management
- **UPDATE**: Template optimization after moai-adk package update
- **GLM Configuration** (`--glm-on <token>`): Configure GLM API integration

---

## PHASE 1: Command Routing & Analysis

**Step 1: Route Based on Subcommand**

1. **`/moai:0-project --glm-on [api-token]`** → GLM CONFIGURATION MODE
   - Detect token from argument, `.env.glm`, or `ANTHROPIC_AUTH_TOKEN` env var
   - If all missing: Request from user via AskUserQuestion
   - Delegate to project-manager with GLM context
   - Execute setup-glm.py script

2. **`/moai:0-project setting`** → SETTINGS MODE
   - Interactive tab selection via AskUserQuestion

3. **`/moai:0-project update`** → UPDATE MODE

4. **`/moai:0-project`** (no args):
   - `.moai/config/config.json` exists → AUTO-DETECT MODE
   - Missing → INITIALIZATION MODE

5. **Invalid subcommand** → Show error and exit

**Step 2: Delegate to Project Manager Agent**

```
Task(
  subagent_type="project-manager",
  prompt="Route and analyze project setup request",
  context={
    "mode": $MODE,
    "language": (from config.json if exists),
    "glm_token": $GLM_TOKEN (if GLM mode)
  }
)
```

---

## PHASE 2: Execute Mode

**INITIALIZATION MODE**:
- Read language from config.json (or use CLI default)
- Conduct language-aware user interview
- Generate project documentation
- Invoke moai-project-config-manager for config creation

**AUTO-DETECT MODE**:
- Read current language from config.json
- Detect partial initialization state (check `.moai/project/` for product.md, structure.md, tech.md)
- If docs missing: Ask user via AskUserQuestion (in user's language): "Complete initialization now?"
  - Options: "Yes, complete" / "No, review" / "Cancel"
  - If "Yes" → Switch to INITIALIZATION workflow
- Display current configuration
- Offer: Modify Settings / Change Language / Review / Re-initialize / Cancel
- "Change Language Only" → SETTINGS mode, Tab 1 only

**SETTINGS MODE (Tab-Based Configuration)**:
- Load current language from config.json
- Load tab schema from `.claude/skills/moai-project-batch-questions/tab_schema.json`
- Execute batch questions via moai-project-batch-questions skill
- Process responses and update config.json atomically via moai-project-config-manager
- Report changes

**UPDATE MODE**:
- Preserve language from config backup
- Invoke moai-project-template-optimizer for smart merging
- Auto-translate announcements to current language if needed

**GLM_CONFIGURATION MODE**:
- Receive GLM API token from parameter or detect from environment (sequence: --glm-on arg → .env.glm → ANTHROPIC_AUTH_TOKEN env var → user input)
- Execute: `uv run .moai/scripts/setup-glm.py <GLM_TOKEN>`
- Verify configuration in .claude/settings.local.json:
  - ANTHROPIC_AUTH_TOKEN in "env" section
  - ANTHROPIC_BASE_URL: https://api.z.ai/api/anthropic
  - ANTHROPIC_DEFAULT_HAIKU_MODEL: glm-4.5-air
  - ANTHROPIC_DEFAULT_SONNET_MODEL: glm-4.6
  - ANTHROPIC_DEFAULT_OPUS_MODEL: glm-4.6
- Verify .env.glm created with 0o600 permissions
- Verify .gitignore includes .env.glm
- Report success with configured keys

---

## SETTINGS Mode: Tab-Based Configuration

**Version**: 2.0.0 | **Last Updated**: 2025-11-19

### Tab Structure

**Tab 1: User & Language** (Required)
- Batch 1.1: Basic settings (3 questions)
  - User name, conversation language, agent prompt language
  - Note: conversation_language_name auto-updated from language code

**Tab 2: Project Basic Information** (Recommended)
- Batch 2.1: Project metadata (4 questions)
  - Project name, description, owner, mode
- Batch 2.2: Auto-processing (0 questions, internal only)
  - project.locale, default_language auto-determined from conversation_language

**Tab 3: Git Strategy & Workflow** (Recommended - Redesigned v2.0.0)
- Batch 3.0: Workflow mode selection (1 question: Personal/Team/Hybrid) → Controls visibility of subsequent batches
- Batch 3.1: Personal core settings (4 questions) - CONDITIONAL (Personal/Hybrid)
- Batch 3.2: Personal branch & cleanup (4 questions) - CONDITIONAL (Personal/Hybrid)
- Batch 3.3: Personal protection & merge (4 questions) - CONDITIONAL (Personal/Hybrid)
- Batch 3.4: Team core settings (4 questions) - CONDITIONAL (Team/Hybrid)
- Batch 3.5: Team branch & protection (4 questions) - CONDITIONAL (Team/Hybrid)
- Batch 3.6: Team safety & merge (2 questions) - CONDITIONAL (Team/Hybrid)
- Total: 29 settings with conditional visibility based on mode

**Tab 4: Quality Principles & Reports** (Optional)
- Batch 4.1: Constitution settings (3 questions)
  - Note: minimum_test_coverage renamed to test_coverage_target
- Batch 4.2: Report generation policy (4 questions)
- Total: 7 settings

**Tab 5: System & GitHub Integration** (Optional)
- Batch 5.1: MoAI system settings (3 questions)
- Batch 5.2: GitHub automation settings (5 questions)
- Total: 8 settings

### Batch Execution Flow

**Step 1: Load Tab Schema**

Load: `.claude/skills/moai-project-batch-questions/tab_schema.json`

Extract: Tab definition, batch questions, field mappings, current values, validation rules

**Step 2: Execute Batch via AskUserQuestion**

Call AskUserQuestion with batch questions (max 4 per batch, plain text only - NO EMOJIS).

Example (Tab 1, Batch 1.1):

```
AskUserQuestion(
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
...
]
)
```

**Step 3: Process Responses**

For each question: Get field path from schema → Get user response → Convert to config.json value (Other option = custom input, selected option = mapped value, Keep current = existing value) → Build update object

**Step 4: Validate at Checkpoints**

1. **After Tab 1**: Verify conversation_language valid, verify agent_prompt_language consistency, re-ask if validation fails
2. **After Tab 3**: Validate Personal/Team mode conflicts (Personal: main_branch != "develop", Team: PR base = develop or main), validate branch naming consistency
3. **Before Config Update**: Check all required fields set, verify no conflicting settings, validate field types, report results

**Step 5: Delegate Atomic Config Update to Skill**

Delegate ALL config updates to Skill("moai-project-config-manager"):
- Skill handles backup/rollback internally
- Skill performs deep merge with validation
- Skill writes atomically to config.json
- Skill reports success/failure

**After Tab Completion**:

Ask: "Would you like to modify another settings tab?"
- No, finish settings (exit)
- Select another tab (return to Tab Selection Screen)

### Tab 3 Validation Example (Complex)

User selects Tab 3:

1. Execute Batch 3.0 (Workflow Mode Selection) → User selects Personal, Team, or Hybrid
2. CONDITIONAL LOGIC:
   - IF Personal: Execute Batches 3.1-3.3, skip 3.4-3.6
   - IF Team: Skip Batches 3.1-3.3, execute 3.4-3.6
   - IF Hybrid: Execute ALL batches 3.1-3.6
3. Validate Tab 3 checkpoint (mode consistency, branch naming)
4. Merge all executed batches into single update object
5. Delegate atomic update to Skill("moai-project-config-manager")

### Multi-Tab Workflow

User runs: `/moai:0-project setting` → Tab Selection Screen

```
Flow:
1. Show Tab Selection Screen
2. User selects tab or "Modify All Tabs"
3. Execute selected tab(s)
4. After tab completion: "Would you like to modify another settings tab?"
   - No, finish settings (exit)
   - Select another tab (loop to step 2)
5. Final atomic update after user finishes all selected tabs

Behavior:
- If user cancels mid-tab: Changes NOT saved
- If tab validation fails: User can retry or skip tab
- After ALL tabs complete: Single final atomic update
- Auto-processing during atomic update (e.g., conversation_language_name auto-update, locale auto-determination)
- Tab 3 conditional batches respect mode selection
```

---

## Error Handling Strategy

**AskUserQuestion Errors**:

1. User Cancels Mid-Batch → Discard all changes in current batch, return to Tab Selection Screen, no partial updates to config.json
2. Timeout (> 5 minutes) → Log warning, notify user in conversation_language, ask user to resume or restart batch, preserve context for retry
3. Invalid Custom Input → Validate on "Other" option selection, show error message, re-ask question (example: "xyz" for language → reject, show valid options)

**Config Update Errors**:

1. File Write Failure → Skill reports error (no changes written), check file permissions, retry after user confirms, fallback: save to `.moai/temp/config-pending.json` with recovery instructions
2. Validation Failure After Merge → Rollback via Skill backup mechanism, show conflict details, ask user to retry batch, log error to `.moai/logs/config-errors.log`
3. Batch Interruption → If agent crashes mid-batch, state not persisted, user must restart batch from beginning, no automatic resume to prevent partial configuration states

---

## PHASE 2.5: Save Phase Context

**Step 1: Extract Context from Agent Response**

After project-manager completes, extract:
- Project metadata: name, mode, owner, language
- Files created: List with absolute paths
- Tech stack: Primary codebase language
- Next phase: Recommended next command (1-plan)

**Step 2: Delegate Context Saving to project-manager**

Agent delegates to Skill("moai-project-config-manager"):
- Save context via ContextManager
- Handle file path validation
- Implement error recovery (non-blocking)
- Report success/failure

Context save failures MUST NOT block command completion.

---

## PHASE 3: Completion & Next Steps

**Step 1: Display Completion Status**

Mode-specific completion message in user's language:
- INITIALIZATION: "Project initialization complete"
- AUTO-DETECT: "Configuration review/modification complete"
- SETTINGS: "Settings updated successfully"
- UPDATE: "Templates optimized and updated"

**Step 2: Offer Next Steps**

Use AskUserQuestion in user's language (NO EMOJIS):
- From Initialization: Write SPEC / Review Structure / New Session
- From Settings: Continue Settings / Sync Documentation / Exit
- From Update: Review Changes / Modify Settings / Exit

---

## Critical Rules

**MANDATORY**:
- Execute ONLY ONE mode per invocation
- Never skip language confirmation/selection
- Always use user's conversation_language for all output
- Auto-translate announcements after language changes
- Route to correct mode based on command analysis
- Delegate ALL execution to project-manager agent
- Use AskUserQuestion for ALL user interaction
- NO EMOJIS in ANY AskUserQuestion field (question, header, options)

**No Direct Tool Usage**:
- NO Read (file operations delegated)
- NO Write (file operations delegated)
- NO Edit (file operations delegated)
- NO Bash (delegated to agents)
- ONLY Task() and AskUserQuestion()

**Configuration Priority**:
- `.moai/config/config.json` settings ALWAYS take priority
- Existing language settings respected unless user requests change
- Fresh installs: Language selection FIRST, then other config

---

## Quick Reference

| Scenario | Mode | Entry Point | Key Phases |
|----------|------|-------------|-----------|
| First-time setup | INITIALIZATION | `/moai:0-project` (no config) | Read language → Interview → Docs |
| Existing project | AUTO-DETECT | `/moai:0-project` (config exists) | Read language → Display → Options |
| Modify config | SETTINGS | `/moai:0-project setting` | Tab selection → Conditional batches → Skill update |
| After package update | UPDATE | `/moai:0-project update` | Preserve language → Template merge → Announce |

**Associated Skills**:
- `Skill("moai-project-language-initializer")` - Language selection/change
- `Skill("moai-project-config-manager")` - Config operations (atomic updates, backup/rollback)
- `Skill("moai-project-template-optimizer")` - Template merging
- `Skill("moai-project-batch-questions")` - Tab-based batch questions

**Project Documentation Directory**:
- Location: `.moai/project/` (singular)
- Files: product.md, structure.md, tech.md (auto-generated or interactive)
- Language: Auto-translated to user's conversation_language

**Version**: 2.0.0 (Tab-based Configuration with Conditional Batches)
**Last Updated**: 2025-11-19
**Tab Schema**: `.claude/skills/moai-project-batch-questions/tab_schema.json` (v2.0.0)

**v2.0.0 Improvements**:
- Removed [tab_ID] argument → Always interactive tab selection
- Added git_strategy.mode selection (Batch 3.0) with Personal/Team/Hybrid conditional logic
- Expanded Tab 3: 16 → 29 settings
- Fixed 26 field names aligned with config.json v0.26.0
- Enhanced validation: 3 → 6 checkpoint rules
- Total coverage: 41 → 57 settings

**Field Name Corrections (v2.0.0)**: conversation_language_name (auto-updated), test_coverage_target (was minimum_test_coverage), auto_checkpoint (was checkpoint_enabled), github.spec_git_workflow (was github.spec_workflow), branch_protection (was prevent_main_direct_merge), and 21 additional field alignments. Full list: `.moai/docs/field-migration-guide.md`

---

## EXECUTION DIRECTIVE

Execute the command following the "Execution Philosophy" described above:

1. Analyze the subcommand/context.
2. Call the `Task` tool with `subagent_type="project-manager"` immediately.
3. DO NOT describe what you will do. DO IT.
