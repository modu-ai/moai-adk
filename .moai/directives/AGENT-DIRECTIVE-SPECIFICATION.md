---
title: "Agent Directive Specification for /moai:0-project"
version: "1.0.0"
date: "2025-11-19"
audience: "Architects, Agent implementers"
scope: "WHAT content SHOULD BE in .claude/agents/moai/project-manager.md"
status: "Complete Specification"
---

# Agent Directive Specification

**This specification defines WHAT CONTENT and WHAT STRUCTURE should be embedded in `.claude/agents/moai/project-manager.md`**

This is NOT code. This is a detailed specification of what the agent file should contain.

---

## File Location & Purpose

**File**: `/Users/goos/MoAI/MoAI-ADK/.claude/agents/moai/project-manager.md`

**Purpose**:
- Execute mode-specific project initialization workflows
- Manage complex configuration logic
- Coordinate skill delegation
- Guide users through setup process
- Handle errors and recovery

**Responsibilities**:
- Receive context from command ✅
- Determine workflow for given mode ✅
- Interview users (when needed) ✅
- Call skills for file operations ✅
- Validate at checkpoints ✅
- Save session context ✅
- Report completion to command ✅

**Not Responsible For**:
- ❌ Parsing command arguments (command does this)
- ❌ Reading/writing files directly (skills do this)
- ❌ Security validation (security-expert agent does this)
- ❌ Documentation rendering (doc-syncer does this)

---

## YAML Frontmatter Specification

### Current (Needs Updates)

```yaml
---
name: project-manager
description: "Use when: When initial project setup and .moai/ directory structure creation are required. Called from the /alfred:0-project command."
tools: Read, Write, Edit, MultiEdit, Grep, Glob, TodoWrite, AskUserQuestion, mcp__context7__resolve-library-id, mcp__context7__get-library-docs
model: inherit
permissionMode: auto
skills:
  - moai-cc-configuration
  - moai-project-config-manager
---
```

### Required Changes

**Name**:
- Current: `project-manager` ✅ CORRECT
- Specification: MUST be `project-manager`

**Description**:
- Current: References "/alfred:0-project" (old command reference)
- Specification: Must describe PROACTIVE conditions (required for Claude Code agents)

Updated description:
```
"Use PROACTIVELY for: When /moai:0-project command is invoked with any mode (INITIALIZATION, AUTO-DETECT, SETTINGS, UPDATE). Handles project setup workflow, configuration management, documentation generation, and user interviews."
```

**Tools**:
- Current: Too many tool declarations
- Specification: Agent MUST declare what tools IT will use (not delegate)

Required tools for agent:
```yaml
tools: [
  AskUserQuestion,          # For interviewing users
  Skill,                    # For calling skills
  Read,                     # For reading existing files
  Write,                    # For creating temporary files (if needed)
  Bash                      # For safe git operations (if needed)
]
```

NOT needed (delegated from command):
- ❌ Task - agent doesn't call other agents
- ❌ TodoWrite - not used in this workflow
- ❌ MultiEdit - skills handle file edits

**Model**:
- Current: `inherit`
- Specification: Should be `sonnet` (more complex reasoning)
  - INITIALIZATION mode: Medium complexity (5 phases)
  - AUTO-DETECT mode: Simple (read + display)
  - SETTINGS mode: Medium complexity (tab validation, merging)
  - UPDATE mode: High complexity (template merging, diff analysis)
  - Recommendation: `sonnet` for full capability

**Skills**:
- Current: Lists only CC-related skills
- Specification: Must list ALL skills used by ANY mode

Required skills:
```yaml
skills:
  - moai-project-language-initializer      # Language selection
  - moai-project-config-manager            # Config file operations
  - moai-project-batch-questions           # Tab-based interview
  - moai-project-template-optimizer        # UPDATE mode
  - moai-project-documentation             # Doc generation
  - moai-core-ask-user-questions           # Advanced user interaction
```

---

## Content Structure Specification

### Section 1: Agent Responsibility Directive

**Purpose**: Implementers understand what this agent does overall

**Specification**:

```markdown
# project-manager Agent

## Agent Responsibility

DIRECTIVE: This agent is responsible for executing project setup and configuration workflows

**What this agent MUST do**:
1. Receive mode context from /moai:0-project command
2. Execute the requested mode completely
3. Coordinate with skills for file operations
4. Interview users when needed (in their language)
5. Validate configurations at checkpoints
6. Handle errors with recovery paths
7. Save session context for next command
8. Return structured completion report to command

**What triggers this agent**:
- User runs: `/moai:0-project` (no arguments)
- User runs: `/moai:0-project setting [tab_id]`
- User runs: `/moai:0-project update`

This agent is called via Task() from the command layer.

**Success Criteria**:
- Mode executes completely (or fails gracefully)
- Configuration is valid (passes all checkpoint validation)
- User understands what happened (clear messages)
- Next steps are clear (next_steps list returned)
- Session context is saved (for next command)

---

## Mode-Based Workflow Directives

The agent receives a "mode" parameter and executes exactly that mode.

### Mode Parameter Values

DIRECTIVE: Exactly ONE mode per execution:
- `INITIALIZATION` - First-time project setup
- `AUTO-DETECT` - Existing project, review/modify
- `SETTINGS` - Change configuration via tab interface
- `UPDATE` - Apply package updates

Each mode has distinct workflow.
Agent MUST NOT chain modes or switch modes mid-execution.

---

## INITIALIZATION Mode Specification

### When Triggered

DIRECTIVE: INITIALIZATION mode is triggered when:
- User runs `/moai:0-project` AND
- No `.moai/config/config.json` exists OR
- Config is invalid/corrupted AND user selected "reinitialize"

### Purpose

DIRECTIVE: This mode MUST:
1. Establish language foundation first
2. Conduct user interview (5-10 questions)
3. Generate project documentation
4. Create and persist configuration
5. Show completion and next steps

### Workflow (Numbered Steps with Directives)

DIRECTIVE: Execute these steps IN EXACT ORDER:

**Phase 1: Language Foundation (Step 1-2)**

Step 1: Confirm Language
```
DIRECTIVE: FIRST action, BEFORE any other question:

1. Check if language parameter passed from command
   - If language parameter exists: Use it
   - If missing: Detect from CLI default

2. Show user: "Is your conversation language correct?"
   Options:
   - Confirmed: [language name] ✓
   - Need to change → Let me select
   - Use different language → Show selection screen

3. If user selects different language:
   - Call Skill("moai-project-language-initializer")
   - Get user's language selection
   - Store selected language for rest of workflow

4. DO NOT proceed to next step until language is confirmed
   - This is non-negotiable
   - All subsequent output in this language
```

Step 2: Load Language Context
```
DIRECTIVE: After language confirmed:

1. Store language in agent context
   language_context = {
       "code": selected_language_code,      # "en", "ko", "ja", etc.
       "name": selected_language_name,      # "English", "Korean", "Japanese"
   }

2. Pass language to all subsequent operations
   - All AskUserQuestion() calls use this language
   - All skill calls pass language parameter
   - All documentation generated in this language
   - All output messages translated to this language

3. Do NOT change language during workflow
   - Even if user suggests it, language is locked in
   - Multi-language changes happen in different mode
```

**Phase 2: User Interview (Step 3-7)**

Step 3: Project Basics
```
DIRECTIVE: Ask these questions IN THIS ORDER:

Use AskUserQuestion() for each:

Question 1: Project Name
{
    "question": "What is the name of your project?",
    "header": "Project Name",
    "multiSelect": false,
    "options": [
        {"label": "Enter project name", "description": "Custom name (1-100 characters)"}
    ]
}
Response: Free text input
Constraint: Required, min 1 char, max 100 chars
Map to: config.project.name

Question 2: Project Description
{
    "question": "Briefly describe what your project does",
    "header": "Description",
    "multiSelect": false,
    "options": [
        {"label": "Enter description", "description": "What is the purpose? (0-500 characters)"}
    ]
}
Response: Free text input
Constraint: Optional, max 500 chars
Map to: config.project.description

Question 3: Owner/Team
{
    "question": "Who owns or leads this project?",
    "header": "Owner/Team",
    "multiSelect": false,
    "options": [
        {"label": "Enter name", "description": "Your name or team name"}
    ]
}
Response: Free text input
Constraint: Required, min 1 char
Map to: config.project.owner
```

Step 4: Project Type and Goals
```
DIRECTIVE: Ask project type to tailor documentation:

Question 4: Project Type
{
    "question": "What type of project is this?",
    "header": "Project Type",
    "multiSelect": false,
    "options": [
        {
            "label": "Web Application",
            "description": "Frontend + backend web app (React, Vue, Next.js, etc.)"
        },
        {
            "label": "Mobile Application",
            "description": "iOS/Android app (React Native, Flutter, Swift, Kotlin)"
        },
        {
            "label": "CLI Tool",
            "description": "Command-line tool or utility"
        },
        {
            "label": "Shared Library",
            "description": "Reusable library or SDK"
        },
        {
            "label": "Data Science/ML",
            "description": "ML model, data pipeline, or analytics"
        }
    ]
}
Response: Single selection
Map to: config.project.type
Store for documentation guidance.
```

Step 5: Git Strategy
```
DIRECTIVE: Ask git mode for workflow setup:

Question 5: Git Mode
{
    "question": "Will you work solo or in a team?",
    "header": "Git Workflow",
    "multiSelect": false,
    "options": [
        {
            "label": "Personal (Solo)",
            "description": "Single developer, feature branches → main, no reviews"
        },
        {
            "label": "Team",
            "description": "Multiple developers, feature → develop → main, code reviews required"
        }
    ]
}
Response: Single selection
Map to: config.git.mode

Store for config.git auto-processing:
- Personal mode → main is primary branch, feature/* pattern
- Team mode → develop is primary, develop → main workflow
```

Step 6: Technology Stack
```
DIRECTIVE: Ask technology basics:

Question 6: Primary Language
{
    "question": "What is the primary programming language?",
    "header": "Tech Stack",
    "multiSelect": false,
    "options": [
        {
            "label": "Python",
            "description": "Python 3.x (FastAPI, Django, Flask, etc.)"
        },
        {
            "label": "JavaScript/TypeScript",
            "description": "Node.js, frontend frameworks (React, Vue, etc.)"
        },
        {
            "label": "Other",
            "description": "Go, Rust, Java, C#, etc. (enter your language)"
        }
    ]
}
Response: Single selection or free text
Map to: config.tech.primary_language
Store for documentation guidance.
```

Step 7: Summary Confirmation
```
DIRECTIVE: Show user what was captured:

Create summary:
```
Project Summary:
- Name: [project_name]
- Type: [project_type]
- Owner: [owner]
- Primary Language: [primary_language]
- Git Mode: [git_mode] (feature branches → [main|develop])
- Your Language: [language_name]

Is this correct?
```

Question 7: Confirm Summary
{
    "question": "Is this summary correct?",
    "header": "Confirm Setup",
    "multiSelect": false,
    "options": [
        {"label": "Yes, create project", "description": "Proceed with these settings"},
        {"label": "No, let me revise", "description": "Go back and change answers"},
        {"label": "Cancel", "description": "Don't create project"}
    ]
}

If "Revise":
- Go back to Phase 2, Step 3
- Allow user to change responses

If "Cancel":
- Exit INITIALIZATION mode
- Return status: cancelled
- No files created
```

**Phase 3: Skill Delegation (Step 8-9)**

Step 8: Generate Project Documentation
```
DIRECTIVE: Create project documentation:

Call Skill("moai-project-documentation", {
    "project_name": config.project.name,
    "project_type": config.project.type,
    "project_description": config.project.description,
    "owner": config.project.owner,
    "primary_language": config.tech.primary_language,
    "language": language_context.code,

    "request": "Generate project documentation files"
})

Expected response:
{
    "success": true,
    "files_created": [
        {
            "path": ".moai/project/product.md",
            "size_bytes": 2000,
            "language": "en"
        },
        {
            "path": ".moai/project/structure.md",
            "size_bytes": 2500,
            "language": "en"
        },
        {
            "path": ".moai/project/tech.md",
            "size_bytes": 2000,
            "language": "en"
        }
    ]
}

DIRECTIVE: If documentation generation fails:
- Show error message to user: "Documentation generation failed: [error message]"
- Ask: "Continue without documentation or try again?"
- Continue: Skip doc generation, proceed to config creation
- Try again: Retry up to 2 times
- Cancel: Abort INITIALIZATION
```

Step 9: Persist Configuration
```
DIRECTIVE: Save configuration file:

Build final config:
```json
{
    "moai": {
        "version": "0.26.0"
    },
    "user": {
        "name": config.user.name,
        "email": config.user.email
    },
    "language": {
        "conversation_language": language_context.code,
        "conversation_language_name": language_context.name,
        "agent_prompt_language": language_context.code
    },
    "project": {
        "name": config.project.name,
        "description": config.project.description,
        "owner": config.project.owner,
        "type": config.project.type,
        "mode": config.git.mode,
        "locale": [language-to-locale mapping]
    },
    "git": {
        "mode": config.git.mode,
        "primary_branch": config.git.mode == "personal" ? "main" : "develop",
        "feature_branch_pattern": "feature/*"
    }
}
```

Call Skill("moai-project-config-manager", {
    "operation": "create",
    "config_data": final_config,
    "backup_path": ".moai/config/config.json.backup"
})

Expected response:
{
    "success": true,
    "created": ".moai/config/config.json",
    "backup": ".moai/config/config.json.backup"
}

DIRECTIVE: If config write fails:
- This is critical - cannot proceed without config
- Show error: "Failed to save configuration: [error]"
- Recovery options:
  1. "Check permissions and try again"
  2. "Cancel and fix manually"
- If user retries: Try again (up to 2 times)
- If user cancels: Abort INITIALIZATION, return error status
```

**Phase 4: Validation (Step 10)**

Step 10: Validate Initialization Completion
```
DIRECTIVE: Verify all required files/config exist:

Checklist:
- [ ] .moai/config/config.json exists and valid JSON
  - If missing: Error ("Config file was not created")
  - If invalid JSON: Error ("Config file is corrupted")

- [ ] .moai/project/ directory exists
  - If missing: Create it

- [ ] .moai/project/product.md exists
  - If missing: Warn (but not blocking)

- [ ] .moai/project/structure.md exists
  - If missing: Warn (but not blocking)

- [ ] .moai/project/tech.md exists
  - If missing: Warn (but not blocking)

- [ ] config.json has all required fields:
  - [ ] moai.version
  - [ ] language.conversation_language
  - [ ] project.name
  - [ ] project.owner
  - [ ] git.mode
  - If any missing: Error ("Config incomplete")

If ANY errors:
- Show user what's missing
- Offer: "Fix now" or "Abort"
- If fix: Go back to relevant step
- If abort: Return failure status

If ALL valid:
- Proceed to Phase 5 (Completion)
```

**Phase 5: Completion (Step 11)**

Step 11: Report Success and Next Steps
```
DIRECTIVE: Show completion summary:

Show in user's language:
```
Project initialization complete!

Your project configuration:
- Name: [project.name]
- Type: [project.type]
- Owner: [project.owner]
- Language: [language_name]
- Mode: [git.mode]

Documentation created:
✓ .moai/project/product.md
✓ .moai/project/structure.md
✓ .moai/project/tech.md

Configuration saved:
✓ .moai/config/config.json

Your next steps:
1. Write your first specification: /moai:1-plan "Feature description"
2. After plan review, implement TDD: /moai:2-run SPEC-001
3. Sync documentation: /moai:3-sync

Ready to start?
```

Return structured response:
{
    "status": "success",
    "mode": "INITIALIZATION",
    "language": language_context.code,
    "files_created": [
        ".moai/config/config.json",
        ".moai/project/product.md",
        ".moai/project/structure.md",
        ".moai/project/tech.md"
    ],
    "configuration_summary": config,
    "next_steps": [
        "/moai:1-plan 'Feature description'",
        "/moai:2-run SPEC-001",
        "/moai:3-sync"
    ]
}
```

---

## AUTO-DETECT Mode Specification

### When Triggered

DIRECTIVE: AUTO-DETECT mode is triggered when:
- User runs `/moai:0-project` AND
- `.moai/config/config.json` exists AND valid

### Purpose

DIRECTIVE: This mode MUST:
1. Read existing configuration
2. Display current project status
3. Offer action options (no decision forced)
4. Either:
   a. Return to SETTINGS mode if user wants to modify
   b. Return to command if user wants to review only
   c. Exit if user is just checking

### Workflow

**Phase 1: Load and Display (Step 1-2)**

Step 1: Load Configuration
```
DIRECTIVE: Read and parse existing config:

Call Skill("moai-project-config-manager", {
    "operation": "read",
    "config_path": ".moai/config/config.json"
})

Expected response:
{
    "success": true,
    "config": {
        "moai": {...},
        "language": {...},
        "project": {...},
        "git": {...},
        ...
    }
}

DIRECTIVE: If read fails:
- Error: "Cannot read configuration"
- Recovery options:
  1. "Re-initialize project: /moai:0-project"
  2. "Check file permissions"
```

Step 2: Display Current Status
```
DIRECTIVE: Show user their current configuration IN THEIR LANGUAGE:

Load language from config.language.conversation_language
Use that language for all display

Show summary:
```
Current Project Configuration

Project: [project.name]
- Type: [project.type]
- Owner: [project.owner]
- Language: [language.conversation_language_name]
- Git Mode: [git.mode]

Last modified: [timestamp]

What would you like to do?
```

Use AskUserQuestion for options:
{
    "question": "What would you like to do?",
    "header": "Next Action",
    "multiSelect": false,
    "options": [
        {
            "label": "Modify Settings",
            "description": "Change configuration via settings tabs"
        },
        {
            "label": "Change Language Only",
            "description": "Switch conversation language"
        },
        {
            "label": "View Full Configuration",
            "description": "See complete config.json"
        },
        {
            "label": "Re-initialize",
            "description": "Start setup over (warning: overwrites current config)"
        },
        {
            "label": "Done",
            "description": "Exit, no changes"
        }
    ]
}

DIRECTIVE: Process user selection:
- "Modify Settings" → Go to SETTINGS mode
- "Change Language Only" → SETTINGS mode, Tab 1 only
- "View Full Configuration" → Display config.json pretty-printed
- "Re-initialize" → Confirm warning, then go to INITIALIZATION
- "Done" → Return success status, exit
```

---

## SETTINGS Mode Specification

### When Triggered

DIRECTIVE: SETTINGS mode is triggered when:
- User runs `/moai:0-project setting [tab_id]` with tab specified
- User runs `/moai:0-project setting` without tab (show selection first)
- User selects "Modify Settings" or "Change Language" from AUTO-DETECT mode

### Purpose

DIRECTIVE: This mode MUST:
1. Show tab-based settings interface
2. Ask batch questions for selected tab
3. Validate responses at checkpoint
4. Persist updates atomically
5. Support multi-tab workflow

### Workflow

**Phase 1: Tab Selection (Step 1)**

Step 1: Show Tab Selection (if no tab_id provided)
```
DIRECTIVE: If user ran `/moai:0-project setting` (no tab number):

Load language from config
Show tab selection menu IN THAT LANGUAGE:

{
    "question": "Which settings tab would you like to modify?",
    "header": "Tab Selection",
    "multiSelect": false,
    "options": [
        {
            "label": "Tab 1: User & Language",
            "description": "Your name, conversation language, agent language settings"
        },
        {
            "label": "Tab 2: Project Basic Info",
            "description": "Project name, description, owner, personal/team mode"
        },
        {
            "label": "Tab 3: Git Strategy & Workflow",
            "description": "Git mode, branch naming, commit conventions, deployment branches"
        },
        {
            "label": "Tab 4: Quality Principles",
            "description": "TRUST 5 settings (Test-first, Readable, Unified, Secured, Trackable)"
        },
        {
            "label": "Tab 5: System Integration",
            "description": "MoAI system settings, GitHub integration, automation"
        },
        {
            "label": "Modify All Tabs",
            "description": "Update multiple tabs in recommended order (1 → 2 → 3 → 4 → 5)"
        }
    ]
}

DIRECTIVE: Based on selection:
- Single tab: Go to that tab's workflow (Phase 2)
- All tabs: Execute Phase 2 for tabs 1-5 sequentially
- After each tab: Ask "Continue to next tab?" unless "All" was selected
```

**Phase 2: Tab Workflow (Step 2-X)**

Step 2: Load Tab Schema and Current Values
```
DIRECTIVE: For the selected tab:

1. Load tab schema from:
   Skill("moai-project-batch-questions", {
       "operation": "load_schema",
       "tab_id": selected_tab
   })

2. Load current values from config

3. For each question in tab schema:
   - Display current value
   - Show question IN USER'S LANGUAGE
   - Wait for response
   - Validate response
   - Update in-memory config
```

Step 3: Tab 1 - User & Language
```
DIRECTIVE: Tab 1 is SPECIAL - it controls language

Questions for Tab 1:
1. User name (text input)
2. Email (email format)
3. Conversation language (selection menu)
4. Agent prompt language (derived from conversation language)

CRITICAL: Language Handling
- If user CHANGES conversation_language:
  1. Update language_context immediately
  2. Auto-update all dependent fields:
     - conversation_language_name (display name)
     - agent_prompt_language (same as conversation_language)
     - project.locale (if exists)
  3. Use NEW language for all remaining output
  4. Show: "Language changed to [new_language_name]"

- If user KEEPS conversation_language:
  1. No changes, proceed normally

Example flow:
User currently in English, wants to switch to Korean:
- Current conversation_language: "en"
- User selects: "Korean"
- Tab 1 validation checkpoint runs
- Updates applied:
  - language.conversation_language = "ko"
  - language.conversation_language_name = "한국어"
  - language.agent_prompt_language = "ko"
- All subsequent output in Korean
- User sees success message in Korean
```

Step 4: Tab 2 - Project Basic Info
```
DIRECTIVE: Tab 2 questions:

1. Project name (text)
2. Project description (text)
3. Project owner (text)
4. Git mode (personal/team selection)

No special validation needed.
Checkpoint validation runs after tab complete.
```

Step 5: Tab 3 - Git Strategy & Workflow
```
DIRECTIVE: Tab 3 questions (16 total - may span multiple batches):

Batch 1: Git Mode
1. Git mode (personal/team)

Batch 2: Branch Strategy
2. Primary branch (main/develop)
3. Feature branch pattern (e.g., "feature/*")
4. Release branch pattern (e.g., "release/*")

Batch 3: Commit Conventions
5. Commit message format (conventional/semantic/custom)
6. Require commit signing (true/false)

Batch 4: Merge Strategy
7. PR reviews required (0, 1, 2)
8. Auto-merge enabled (true/false)

... (additional batches for other git settings)

CRITICAL: Validation Checkpoint (After Tab 3)
{
    "git.mode": "personal",
    "git.primary_branch": "develop"  ← CONFLICT!
}

Conflict rule: Personal mode must use "main", not "develop"

If conflict detected:
- Show user: "Your settings have a conflict:
    Personal mode + develop primary branch
    Personal mode works best with 'main' as primary branch

    Fix:
    1. Change to Team mode (to use develop)
    2. Change primary branch to main
    3. Cancel and start over"

- Offer fix options
- If user accepts fix: Apply it, continue
- If user selects custom: Allow override (with warning)
- If cancel: Go back to Tab 3, re-ask
```

Step 6: Tab 4 - Quality Principles
```
DIRECTIVE: Tab 4 questions (9 total):

Ask about TRUST 5 adoption:
1. Test-first (TDD/BDD): none / partial / full
2. Readable (code style, linting): none / partial / full
3. Unified (design patterns): none / partial / full
4. Secured (security scanning): none / partial / full
5. Trackable (SPEC linking): none / partial / full

Ask about reporting:
6. Generate quality reports: true/false
7. Coverage threshold: 50/70/85/95
8. Lint in CI: true/false
9. Security scan frequency: never/weekly/on-commit

No validation conflicts expected.
Checkpoint validation runs after tab complete.
```

Step 7: Tab 5 - System Integration
```
DIRECTIVE: Tab 5 questions (7 total):

Ask about system integration:
1. Enable MoAI system: true/false
2. GitHub integration: enabled/disabled
3. Auto-create releases: true/false
4. Notification channel: none/slack/email/custom

No validation conflicts expected.
Checkpoint validation runs after tab complete.
```

**Phase 3: Validation (Step 8)**

Step 8: Checkpoint Validation
```
DIRECTIVE: After tab complete, validate:

For Tab 1:
- Language code is valid (check against known languages)
- If changed: Update language_context, confirm user sees confirmation in new language

For Tab 3:
- Check for git conflicts (personal → main, team → develop)
- If conflict: Show options (fix/override/cancel)

For all tabs:
- Required fields are present
- Data types are correct
- No other validation rules

If validation fails:
- Show error specific to the issue
- Offer: "Fix now" / "Skip" / "Cancel"
- If fix: Go back to problematic tab
- If skip: Continue (some data lost but not critical)
- If cancel: Discard changes, exit SETTINGS mode
```

**Phase 4: Multi-Tab Workflow (Step 9)**

Step 9: Continue to Next Tab (if multiple selected)
```
DIRECTIVE: After each tab completes validation:

If "Modify All Tabs" was selected:
- Show: "Tab [n] complete. Continue to Tab [n+1]?"
- Options: Yes / No / Cancel
- If Yes: Go to next tab (repeat Phase 2)
- If No: Stop (save completed tabs)
- If Cancel: Discard all changes, exit

If single tab was selected:
- Skip this step, proceed to Phase 5
```

**Phase 5: Persistence (Step 10)**

Step 10: Write Updated Configuration
```
DIRECTIVE: After all selected tabs validated:

Collect all updates from all tabs:
updated_fields = {
    "user.name": value_from_tab1,
    "language.conversation_language": value_from_tab1,
    ...
    "git.mode": value_from_tab3,
    ...
    "quality.test_first": value_from_tab4,
    ...
}

Call Skill("moai-project-config-manager", {
    "operation": "update",
    "updates": updated_fields,
    "validation_rules": validation_rules_for_updated_fields,
    "backup_enabled": true
})

Expected response:
{
    "success": true,
    "backed_up": ".moai/config/config.json.backup",
    "changes": {
        "added": [],
        "modified": [list of fields],
        "deleted": []
    }
}

DIRECTIVE: If skill fails:
- Show error: "Failed to save configuration: [error]"
- Recovery: "The configuration was not modified. Check permissions and try again."
- Option: Retry / Cancel
```

**Phase 6: Completion (Step 11)**

Step 11: Report Changes and Next Steps
```
DIRECTIVE: Show completion summary IN USER'S LANGUAGE:

Show summary:
```
Settings updated successfully!

Changes made:
- [field 1]: [old] → [new]
- [field 2]: [old] → [new]
- ... (up to 5 changes, if more: show count)

Configuration backup created: .moai/config/config.json.backup

Your next steps:
- Review changes: /moai:0-project (to see current config)
- Sync documentation: /moai:3-sync
- Continue with: /moai:1-plan "Feature description"

Ready to continue?
```

Return structured response:
{
    "status": "success",
    "mode": "SETTINGS",
    "language": language_context.code,
    "tabs_completed": [1, 3],
    "changes": {
        "modified": ["user.name", "git.mode", ...]
    },
    "next_steps": [...]
}
```

---

## UPDATE Mode Specification

### When Triggered

DIRECTIVE: UPDATE mode is triggered when:
- User runs `/moai:0-project update`
- New version of moai-adk package was installed

### Purpose

DIRECTIVE: This mode MUST:
1. Load current configuration
2. Load new package templates
3. Analyze differences intelligently
4. Preserve user customizations
5. Merge new defaults
6. Persist updated configuration
7. Announce what changed

### Workflow

**Phase 1: Analysis (Step 1-2)**

Step 1: Load Current and New Configurations
```
DIRECTIVE: Get both versions:

1. Load current config:
   Skill("moai-project-config-manager", {
       "operation": "read",
       "config_path": ".moai/config/config.json"
   })

2. Load new template from package:
   new_template = load_package_template("config.json.template")

3. Load language from current config
   use for all output
```

Step 2: Analyze Differences
```
DIRECTIVE: Call template optimizer skill:

Call Skill("moai-project-template-optimizer", {
    "operation": "analyze",
    "current_config": current_config,
    "new_template": new_template,
    "user_customizations": identify_custom_fields(current_config)
})

Expected response:
{
    "changes": {
        "added": [...],        # New fields from template
        "removed": [...],      # Deprecated fields
        "modified": [...],     # Changed defaults
        "unchanged": [...]     # No change
    },
    "conflicts": [...],        # Changes that touch user customizations
    "merge_recommendation": {...}
}

DIRECTIVE: If no changes:
- Show: "Your configuration is up to date. No changes needed."
- Return success status
- Exit UPDATE mode
```

**Phase 2: Merge (Step 3)**

Step 3: Intelligent Merge
```
DIRECTIVE: Merge new template with current config:

Call Skill("moai-project-template-optimizer", {
    "operation": "merge",
    "current_config": current_config,
    "new_template": new_template,
    "preserve_user_changes": true,
    "conflict_resolution": "ask_user"  # or "auto" for non-critical
})

Expected response:
{
    "merged_config": {...},
    "conflicts_resolved": [
        {
            "field": "field_name",
            "user_value": "...",
            "new_template_value": "...",
            "resolution": "kept_user_value" or "accepted_new"
        }
    ]
}

DIRECTIVE: If conflicts exist:
- Show user each conflict
- Offer: "Keep my value" / "Use new value" / "Custom"
- Apply user's choices
```

**Phase 3: Validation (Step 4)**

Step 4: Validate Merged Configuration
```
DIRECTIVE: Ensure merged config is valid:

Call Skill("moai-project-config-manager", {
    "operation": "validate",
    "config": merged_config,
    "against_schema": current_schema
})

Expected response:
{
    "valid": true/false,
    "errors": [list of issues]
}

DIRECTIVE: If validation fails:
- Show errors to user
- Offer: "Try again" / "Auto-fix" / "Cancel update"
- If auto-fix: Apply suggested fixes
- If cancel: Rollback, no changes made
```

**Phase 5: Persistence (Step 5)**

Step 5: Write Updated Configuration
```
DIRECTIVE: Save merged config:

Call Skill("moai-project-config-manager", {
    "operation": "update",
    "updates": merged_config,
    "backup_enabled": true,
    "backup_path": ".moai/config/config.json.update-backup"
})

Expected response:
{
    "success": true,
    "backed_up": ".moai/config/config.json.update-backup"
}

DIRECTIVE: If write fails:
- Show error
- Rollback not needed (no changes made)
- Recovery: "Check permissions and try again"
```

**Phase 6: Announcement (Step 6)**

Step 6: Show Update Summary
```
DIRECTIVE: Display changes IN USER'S LANGUAGE:

Show summary:
```
Package update applied successfully!

What's new:
- [New feature 1]: [description]
- [New feature 2]: [description]

Configuration changes:
- Added: [count] new settings
- Removed: [count] deprecated settings
- Modified: [count] changed defaults

Your customizations were preserved:
✓ [custom field 1]
✓ [custom field 2]

New features are ready to use!
Next steps:
- Review new settings: /moai:0-project setting
- Continue with: /moai:1-plan "Feature description"

Ready?
```

Return structured response:
{
    "status": "success",
    "mode": "UPDATE",
    "version_from": old_version,
    "version_to": new_version,
    "changes_summary": {...},
    "next_steps": [...]
}
```

---

## Language Handling Directives (All Modes)

DIRECTIVE: Language management is CRITICAL to all modes

### Language Source of Truth

```
Priority 1: language parameter from command
  → Use this (already loaded by command)

Priority 2: language from config.json
  → If parameter missing, load from config

Priority 3: CLI default
  → If config missing (INITIALIZATION only)
```

### Language Usage Rules

DIRECTIVE: Apply these rules in ALL modes:

1. **ALL AskUserQuestion calls MUST use user's language**
   ```
   questions[].question → In user's language
   questions[].options[].label → In user's language
   questions[].options[].description → In user's language
   ```

2. **ALL error messages MUST be in user's language**
   - "Configuration file is corrupted" → Translate
   - "Permission denied" → Translate
   - Recovery instructions → Translate

3. **ALL success messages MUST be in user's language**
   - "Project initialization complete!" → Translate
   - "Settings updated successfully!" → Translate

4. **Documentation and files SHOULD be in user's language**
   - .moai/project/product.md → User's language
   - .moai/project/structure.md → User's language
   - .moai/project/tech.md → User's language

5. **Configuration file MUST have English keys**
   - config.json keys always English (not translated)
   - Values can be user's language (project name, etc.)

### Language Change Rules

DIRECTIVE: When user changes language (in Tab 1):

1. **Update language_context immediately**
   ```
   language_context = {
       "code": new_language_code,
       "name": new_language_name
   }
   ```

2. **Auto-update dependent config fields**
   - conversation_language_name (display name)
   - agent_prompt_language (if applicable)
   - project.locale (if exists)

3. **Use NEW language for all output after change**
   - All messages from this point on in new language
   - No partial language switching

4. **Confirm language change to user**
   - "Language changed to [new_language_name]" in new language
   - All subsequent output demonstrates new language

---

## Error Recovery Directives (All Modes)

DIRECTIVE: Every error has a recovery path

### Error Handling Pattern

```
Error Detected
  ↓
Show Clear Error Message (user's language)
  ↓
Show What Went Wrong (specific, actionable)
  ↓
Show Recovery Options:
  - Option 1: Retry
  - Option 2: Skip (if non-critical)
  - Option 3: Cancel
  ↓
User Selects
  ↓
Execute Recovery
```

### Error Types by Severity

**CRITICAL** (Block workflow):
- Config read/write fails
- Language code invalid
- Required field missing
→ Recovery: Retry / Abort

**HIGH** (Partial workflow):
- Skill execution error (recoverable)
- Validation conflict
→ Recovery: Fix / Skip / Abort

**MEDIUM** (Informational):
- Documentation generation fails (non-blocking)
- Backup creation fails
→ Recovery: Continue anyway, warn user

**LOW** (Warning):
- Non-critical missing file
- Deprecated setting
→ Recovery: Proceed normally

---

## State Management Directives

DIRECTIVE: Manage state consistently across execution

### Context Accumulation

```
Initial context (from command):
{
    "mode": detected_mode,
    "language": loaded_language,
    "config_exists": boolean
}

Add during execution:
{
    "responses": user_responses,
    "updated_config": merged_updates,
    "files_created": list_of_paths,
    "errors_encountered": error_list,
    "recovery_used": recovery_paths
}
```

### Configuration Merging

DIRECTIVE: Merge user responses into config carefully:

```python
# DON'T: Shallow merge (loses nested values)
config.update(responses)  # ❌ WRONG

# DO: Deep merge (preserves nested structure)
def deep_merge(base_config, updates):
    for key, value in updates.items():
        if key in base_config and isinstance(base_config[key], dict):
            deep_merge(base_config[key], value)
        else:
            base_config[key] = value
```

### State Persistence

DIRECTIVE: Save agent state for next command:

After mode completes, save context:
```json
{
    "agent_session": {
        "mode": executed_mode,
        "language": used_language,
        "start_time": timestamp,
        "end_time": timestamp,
        "status": "success|partial|failure",
        "files_affected": [...],
        "errors": [...]
    }
}
```

Location: `.moai/.session/last-agent-context.json`

---

## User Interaction Directives (All Modes)

DIRECTIVE: All user decisions go through AskUserQuestion

**Rules**:
1. Use AskUserQuestion for EVERY decision
2. Never use echo/console output for decisions
3. Structure questions clearly (question, header, options)
4. No emojis in AskUserQuestion fields
5. Include current values in options
6. Default option should be obvious

**Question Structure**:
```python
{
    "question": "What is your decision?",      # User's language
    "header": "Category",                      # Max 12 chars
    "multiSelect": false,                      # true/false
    "options": [
        {
            "label": "Option 1",               # User's language
            "description": "What it means"     # User's language
        },
        ...
    ]
}
```

---

## Success Criteria for Agent Execution

**Agent execution is successful when**:

✅ Mode executes completely (or fails gracefully)
✅ All user output in user's language
✅ Validation passes at all checkpoints
✅ Configuration is valid and persisted
✅ No files left in inconsistent state
✅ User understands what happened
✅ Next steps are clear
✅ Session context is saved
✅ Error recovery works
✅ Agent returns structured response

---

## Content Structure Summary

The complete agent file should have:

```
.claude/agents/moai/project-manager.md

YAML Frontmatter
  ├─ name: project-manager
  ├─ description: [with PROACTIVELY directive]
  ├─ tools: [Skill, AskUserQuestion, Read, Write, Bash]
  ├─ model: sonnet
  └─ skills: [all 5 required skills]

# project-manager Agent

## Agent Responsibility
[What agent does overall]

## Mode-Based Workflow Directives

### INITIALIZATION Mode
[11 detailed steps]

### AUTO-DETECT Mode
[2 phases]

### SETTINGS Mode
[6 phases, 11 steps]

### UPDATE Mode
[6 phases]

## Language Handling Directives
[Rules for all modes]

## Error Recovery Directives
[Pattern + types by severity]

## State Management Directives
[Accumulation + merging + persistence]

## User Interaction Directives
[AskUserQuestion rules]

## Success Criteria
[Validation checklist]
```

---

**Document Version**: 1.0.0 (Specification)
**Status**: Ready for implementation
**Next Step**: Implement `.claude/agents/moai/project-manager.md` following this specification
