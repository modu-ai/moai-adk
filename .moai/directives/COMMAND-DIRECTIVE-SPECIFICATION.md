---
title: "Command Directive Specification for /moai:0-project"
version: "1.0.0"
date: "2025-11-19"
audience: "Architects, Command implementers"
scope: "WHAT content SHOULD BE in .claude/commands/moai/0-project.md"
status: "Complete Specification"
---

# Command Directive Specification

**This specification defines WHAT CONTENT and WHAT STRUCTURE should be embedded in `.claude/commands/moai/0-project.md`**

This is NOT code. This is a detailed specification of what the command file should contain.

---

## File Location & Purpose

**File**: `/Users/goos/MoAI/MoAI-ADK/.claude/commands/moai/0-project.md`

**Purpose**:
- Entry point for project initialization workflow
- User-facing orchestration of project setup
- Specification of command behavior
- Delegation to project-manager agent

**Responsibilities**:
- Parse user arguments ✅
- Detect initialization state ✅
- Load language context ✅
- Route to correct mode ✅
- Delegate to agent ✅
- Display results ✅

**Not Responsible For**:
- ❌ Asking interview questions (that's agent + skills)
- ❌ Writing configuration files (that's skills)
- ❌ Complex mode logic (that's agent)
- ❌ Validation rules (that's agent + skills)

---

## YAML Frontmatter Specification

### Current (Correct)

```yaml
---
name: moai/0-project
description: "Initialize or update project configuration, documentation, and setup"
argument-hint: ["[update|setting [tab_id]]"]
tools: [Task, AskUserQuestion]
model: haiku
permissionMode: auto
---
```

### Required Changes

**Name**:
- Current: `moai/0-project` ✅ CORRECT
- Specification: MUST be `moai/0-project` (slash format)

**Description**:
- Current: Correct but could be more specific
- Specification: "Initialize or update project configuration. Options: no args (auto-detect), setting [tab] (modify settings), update (package update)"

**Argument-hint**:
- Current: `["[update|setting [tab_id]]"]` ✅ CORRECT
- Specification: Must cover all three cases:
  ```yaml
  argument-hint: [
    "[no arguments]",
    "setting [tab_id]",
    "update"
  ]
  ```

**Tools**:
- Current: `[Task, AskUserQuestion]` ✅ CORRECT
- Specification: MUST ONLY include these two:
  - `Task` - for delegating to project-manager agent
  - `AskUserQuestion` - for entry-level user decisions (if any)
  - ❌ MUST NOT include: Read, Write, Edit, Bash, Grep, etc.

**Model**:
- Current: `haiku`
- Specification: Command is lightweight orchestration, so `haiku` is correct

**permissionMode**:
- Current: `auto`
- Specification: `auto` is correct (command doesn't do file ops, only delegates)

---

## Content Structure Specification

### Section 1: Command Overview (User Perspective)

**Purpose**: Users understand what this command does before using it

**Specification**:

```markdown
# /moai:0-project - Project Setup & Configuration

## What This Command Does

Users should understand:
- This command initializes a NEW project or UPDATES an existing one
- It guides through project setup interactively
- It creates/updates `.moai/config/config.json` and project documentation
- It works in their conversation language
- No technical knowledge required

**Do NOT mention**: Agents, skills, technical architecture details

## When Should You Use This?

Present exactly FOUR use cases:

| Scenario | Run This | Expected Result |
|----------|----------|---|
| Starting a completely new project | `/moai:0-project` | Interview + config created + next steps |
| Project already initialized, want to review/modify | `/moai:0-project` | Auto-detect mode, shows current config |
| Want to change language or settings | `/moai:0-project setting` | Tab selection screen |
| Project was just updated to new version | `/moai:0-project update` | Template merge + new features applied |

## Quick Start (30 seconds)

```bash
# 1. Initialize a new project
/moai:0-project

# 2. Review or modify settings
/moai:0-project setting

# 3. After package update
/moai:0-project update
```

## What You Get

**After running this command, you will have**:
- `.moai/config/config.json` - Your project configuration
- `.moai/project/product.md` - What your project is
- `.moai/project/structure.md` - How it's organized
- `.moai/project/tech.md` - What technology it uses
- Clear next steps (usually: write specs with /moai:1-plan)
```

**Style Requirements**:
- Simple, clear language
- No technical jargon
- Benefits-focused (what user gets)
- Action-oriented (what user does)
- No emojis (they don't work in all markdown renderers)

---

### Section 2: Entry Point Directives

**Purpose**: Implementers understand how to parse user input and route correctly

**Specification**:

```markdown
## How It Works (Technical)

### Argument Parsing & Mode Detection

DIRECTIVE: Parse command arguments FIRST, before any other operation

The command MUST handle these cases:

```
Input: /moai:0-project
  ├─ No arguments
  ├─ Check: Does .moai/config/config.json exist?
  │  ├─ YES → AUTO-DETECT mode
  │  └─ NO  → INITIALIZATION mode

Input: /moai:0-project setting
  ├─ "setting" keyword found
  ├─ Check: Is tab_id argument present?
  │  ├─ YES (e.g., "setting 1") → SETTINGS mode for that tab
  │  └─ NO  → Show tab selection screen

Input: /moai:0-project setting [tab_id]
  ├─ Tab ID specified (1, 2, 3, 4, or 5)
  └─ SETTINGS mode for that specific tab

Input: /moai:0-project update
  ├─ "update" keyword found
  └─ UPDATE mode

Input: Invalid argument
  ├─ Example: /moai:0-project invalid
  └─ Show error + usage hint
```

DIRECTIVE: Invalid arguments trigger helpful error:
```
Error: Unknown argument 'invalid'

Usage:
  /moai:0-project              (Auto-detect or initialize)
  /moai:0-project setting      (Modify settings)
  /moai:0-project setting 1    (Modify Tab 1 only)
  /moai:0-project update       (Apply package updates)
```

### Configuration State Validation

DIRECTIVE: Before routing to agent, validate configuration file:

1. Check if `.moai/config/config.json` exists
   - If exists: Validate JSON syntax
     - Valid JSON → Continue
     - Invalid JSON → Offer restore from backup or reinitialize
   - If missing → Proceed to INITIALIZATION mode

2. Load language from config (if exists)
   - Read: config.json → language.conversation_language
   - Use this language for ALL subsequent output
   - If language field missing: Use CLI default

3. Pass loaded language to project-manager agent
   - Agent MUST use this language for all output
   - Agent receives: mode, language, user_request

### Mode Routing Decision

DIRECTIVE: Route to EXACTLY ONE mode:

**Decision Tree**:
```
detected_mode = NONE

If arguments == []:
  If config.json exists AND valid:
    detected_mode = AUTO-DETECT
  Else if config.json missing:
    detected_mode = INITIALIZATION
  Else (invalid):
    detected_mode = ERROR (offer recovery)

Else if arguments[0] == "setting":
  If arguments[1] exists:
    detected_mode = SETTINGS(arguments[1])
  Else:
    detected_mode = SETTINGS(show_tabs)

Else if arguments[0] == "update":
  detected_mode = UPDATE

Else:
  detected_mode = ERROR (invalid)

If detected_mode == ERROR:
  Show error message
  Exit

Else:
  Proceed to Agent Delegation
```

DIRECTIVE: Once mode is detected, DO NOT change it
- Execute that mode completely
- Do not chain modes
- Do not offer mode switching within command

### Language Context Loading

DIRECTIVE: Language MUST be established BEFORE delegating to agent

1. If config.json exists:
   - Read: config.json → language.conversation_language
   - Validate language code exists
   - Use for all output

2. If config.json missing (INITIALIZATION mode):
   - Use CLI default language
   - Agent will ask user to confirm/change
   - Agent will persist selected language

3. Pass language to project-manager agent:
   ```python
   agent_context = {
       "detected_mode": detected_mode,
       "language": loaded_language,
       "config_exists": config_file_exists,
       "user_request": raw_command_arguments
   }
   ```

4. Agent MUST use this language for ALL output
   - No language switching
   - Consistent throughout mode execution
```

**Style Requirements**:
- Decision trees use ASCII art for clarity
- Each DIRECTIVE is numbered/bulleted
- Constraints marked with "MUST" not "should"
- No pseudocode (use numbered steps instead)

---

### Section 3: Tool Usage Constraints

**Purpose**: Implementers understand what tools are available and why

**Specification**:

```markdown
## Tool Usage Constraints

DIRECTIVE: This command uses ONLY these two tools:

### Tool 1: Task()

**Usage**: Delegate all mode execution to project-manager agent

```python
response = Task(
    subagent_type="project-manager",
    prompt=f"""
Mode: {detected_mode}
Language: {loaded_language}
Command arguments: {raw_arguments}

Execute the {detected_mode} mode completely.
Refer to your internal directives for mode-specific workflow.
Return completion status, changes made, and next steps.
"""
)
```

**Why Task() only**:
- Agent has access to all needed tools (Read, Write, Skill, etc.)
- Command stays lightweight (orchestration only)
- Clear separation: command routes, agent executes
- Easier to test (command is pure routing logic)

**Not allowed in command**:
- ❌ Cannot call Skill() directly (go through agent)
- ❌ Cannot use Read/Write/Edit (agent delegates this)
- ❌ Cannot use Bash (security boundary)

### Tool 2: AskUserQuestion()

**Usage**: Only at entry point IF needed for mode selection

**Example** (if tab_id not provided):
```python
response = AskUserQuestion(
    questions=[{
        question: "Which settings tab would you like to modify?",
        header: "Tab Selection",
        multiSelect: false,
        options: [
            {
                label: "Tab 1: User & Language",
                description: "Your name, conversation language, agent prompt language"
            },
            {
                label: "Tab 2: Project Info",
                description: "Project name, description, owner, personal/team mode"
            },
            {
                label: "Tab 3: Git Strategy",
                description: "Git workflow, branch naming, commit conventions"
            },
            {
                label: "Tab 4: Quality Principles",
                description: "TRUST 5 settings, quality standards"
            },
            {
                label: "Tab 5: System Integration",
                description: "GitHub automation, MoAI system settings"
            },
            {
                label: "Modify All Tabs",
                description: "Update settings in recommended order (Tab 1 → 5)"
            }
        ]
    }]
)
```

**When to use**:
- ✅ Tab selection screen (if SETTINGS mode with no tab_id)
- ✅ Error confirmation (offer retry/cancel)

**When NOT to use**:
- ❌ Do NOT ask interview questions (agent's job)
- ❌ Do NOT ask for project details (agent's job)
- ❌ Do NOT ask for git strategy (agent's job)

**Why limited AskUserQuestion**:
- Most interaction happens in agent (not command)
- Command orchestrates, agent interviews
- Keep command fast (haiku model, minimal context)

---

## Restrictions (What NOT to Do)

DIRECTIVE: This command MUST NOT:

❌ **Read configuration files directly**
- Let agent do this via Skill delegation
- Exception: Read config.json ONLY to check existence/validity for mode detection

❌ **Write to any files**
- All file operations through agent → skills
- Config updates, documentation creation all delegated

❌ **Use Bash commands**
- Security boundary: commands don't execute bash
- Any bash operations through agent

❌ **Ask interview questions**
- "What is your project name?" → agent's job
- "What's your tech stack?" → agent's job
- Command only routes, doesn't interview

❌ **Call Skills directly**
- No: Skill("moai-project-config-manager", ...)
- Instead: delegate to agent, agent calls skills

❌ **Handle complex mode logic**
- Each mode has phases (Red-Green-Refactor for /moai:2-run, etc.)
- Command is too simple to handle this complexity
- Agent handles all mode-specific logic

❌ **Validate configuration rules**
- "Language code must be 'en', 'ko', 'ja'" → agent's job
- "Git branch must match pattern X" → agent's job
- Command only validates file existence/JSON syntax

❌ **Translate messages to user's language**
- Well, actually: Command CAN display simple messages in user's language IF it has language from config
- But complex multi-language handling: delegate to agent
- Keep command simple
```

**Style Requirements**:
- Clear "what NOT to do"
- Explanation of why (separation of concerns)
- Examples of violations

---

### Section 4: Error Handling at Entry Point

**Purpose**: Implementers understand what can go wrong at entry point and how to recover

**Specification**:

```markdown
## Entry-Level Error Handling

DIRECTIVE: Errors at entry point are handled BY THE COMMAND, not delegated to agent

### Error Type 1: Invalid Argument

**When it happens**:
```
/moai:0-project invalid_arg
```

**Recovery**:
```
Show error message:
"Unknown argument 'invalid_arg'

Usage:
  /moai:0-project              Initialize or auto-detect
  /moai:0-project setting      Modify settings
  /moai:0-project setting [1-5] Modify specific tab
  /moai:0-project update       Apply package updates

Example:
  /moai:0-project setting 1    Change language
"
Exit command (no further action)
```

### Error Type 2: Configuration File Exists but Invalid JSON

**When it happens**:
- `.moai/config/config.json` exists
- File content is not valid JSON (syntax error)

**Detection**:
```python
try:
    with open(".moai/config/config.json") as f:
        json.load(f)
    config_valid = True
except json.JSONDecodeError:
    config_valid = False
```

**Recovery**:
```
Show warning:
"Configuration file is corrupted (.moai/config/config.json)

Options:
1. Restore from backup (if available)
   /moai:0-project setting 1 (then fix and try again)
2. Re-initialize project
   /moai:0-project (will ask you to confirm before overwriting)
3. Fix manually
   Edit .moai/config/config.json directly
"

If user selects restore:
- Check if backup exists (.moai/config/config.json.bak)
- If backup valid: restore it
- If backup missing: offer reinitialize

If user selects reinitialize:
- Mode = INITIALIZATION
- Proceed with re-init (user will be warned about overwriting)
```

### Error Type 3: Permission Denied

**When it happens**:
- Cannot read `.moai/config/config.json` (permission denied)
- Cannot write to `.moai/config/` (permission denied)

**Recovery**:
```
Show error (in user's language if loaded):
"Permission denied: Cannot access project configuration
  Path: .moai/config/config.json
  Fix: Check file permissions or contact your system administrator

  Details:
  - Verify directory exists: ls -la .moai/config/
  - Check your user can read: stat .moai/config/config.json
  - Restore permissions if needed: chmod 644 .moai/config/config.json
"
Exit command (admin intervention required)
```

### Error Type 4: Directory Structure Missing

**When it happens**:
- `.moai/` directory doesn't exist
- `.moai/config/` directory doesn't exist

**Recovery**:
```
If .moai/ missing:
  → INITIALIZATION mode (will create structure)

If .moai/config/ missing but .moai/ exists:
  → INITIALIZATION mode (will create subdirectory)

These are not errors, they're normal for fresh projects
```

### Entry-Level Error Summary Table

| Error | Detection | Recovery | User Message |
|-------|-----------|----------|---|
| Invalid argument | Parse fails | Show usage | "Unknown argument: X" |
| Config invalid JSON | json.load fails | Offer restore/reinit | "Configuration file is corrupted" |
| Permission denied | OSError | Show fix steps | "Permission denied: Cannot access..." |
| Missing .moai/ | not os.path.exists | Proceed to INIT | (No error, just initialize) |

---

## What Happens After Entry Point

DIRECTIVE: After entry point validation, delegate ALL work to agent

```
Entry Point (command)
  ├─ Parse arguments
  ├─ Detect mode
  ├─ Load language
  ├─ Validate config file (existence/syntax)
  └─ If all valid: Call Task(project-manager, ...)
       │
       └─ project-manager agent takes over
            ├─ Execute mode-specific workflow
            ├─ Ask interview questions (if needed)
            ├─ Call skills for file operations
            ├─ Validate configurations
            └─ Return results
       │
       └─ Back to command
            ├─ Display results
            ├─ Show what changed
            ├─ Present next steps
            └─ Exit
```

DIRECTIVE: Command does NOT:
- ❌ Ask detailed questions (agent does)
- ❌ Validate complex rules (agent does)
- ❌ Write files (skills do)
- ❌ Merge configurations (agent does)

DIRECTIVE: Command ONLY:
- ✅ Parses arguments
- ✅ Validates file existence/syntax
- ✅ Loads language
- ✅ Routes to agent
- ✅ Displays results
```

**Style Requirements**:
- Error table for quick reference
- ASCII flowchart for data flow
- Each error has detection + recovery steps

---

### Section 5: Agent Delegation Specification

**Purpose**: Implementers understand exactly what to pass to agent and what to expect

**Specification**:

```markdown
## Delegation to project-manager Agent

DIRECTIVE: After entry point validation, delegate to project-manager via Task()

### What to Pass to Agent

```python
response = Task(
    subagent_type="project-manager",
    prompt=f"""
EXECUTION CONTEXT:
- Mode: {detected_mode}
  (INITIALIZATION | AUTO-DETECT | SETTINGS | UPDATE)
- Language: {loaded_language}
  (Language code to use for all output, e.g., 'en', 'ko', 'ja')
- User Request: {raw_arguments}
  (Original command arguments)
- Config Exists: {config_file_exists}
  (true/false)
- Command Time: {start_time}
  (ISO 8601 format)

EXECUTION DIRECTIVES:
Execute the {detected_mode} mode completely according to your internal mode directives.
Use the language '{loaded_language}' for ALL user-facing output.
Do NOT change modes mid-execution.
Return structured response with: status, changes, next_steps.

FAILURE PROTOCOL:
If any error occurs:
1. Do not silently fail
2. Report error clearly to user in their language
3. Offer recovery path (retry/cancel)
4. Return error status to command

SUCCESS PROTOCOL:
When mode completes successfully:
1. List exact changes made
2. Show files created/modified (with paths)
3. Present next recommended step
4. Return success status to command
"""
)
```

### What to Expect from Agent

**Success Response**:
```json
{
    "status": "success",
    "mode": "INITIALIZATION",
    "language": "en",
    "duration_seconds": 45,
    "changes": {
        "created": [
            ".moai/config/config.json",
            ".moai/project/product.md",
            ".moai/project/structure.md",
            ".moai/project/tech.md"
        ],
        "modified": [],
        "deleted": []
    },
    "user_input_summary": {
        "project_name": "My Project",
        "owner": "My Team",
        "language": "en"
    },
    "next_steps": [
        "Write specifications: /moai:1-plan",
        "Implement features: /moai:2-run SPEC-001",
        "Sync documentation: /moai:3-sync"
    ],
    "context_saved": true
}
```

**Partial Failure Response**:
```json
{
    "status": "partial_success",
    "mode": "SETTINGS",
    "language": "en",
    "completed_tabs": [1, 2],
    "failed_tabs": [3],
    "error_on_tab": 3,
    "error_message": "Git branch naming conflict detected. Fix: Use consistent prefix.",
    "recovery_options": [
        "Fix the issue and retry: /moai:0-project setting 3",
        "Skip Tab 3 and continue: /moai:0-project setting 4",
        "Cancel and start over: /moai:0-project"
    ]
}
```

**Complete Failure Response**:
```json
{
    "status": "failure",
    "mode": "INITIALIZATION",
    "language": "en",
    "error_type": "config_write_failed",
    "error_message": "Failed to write .moai/config/config.json: Permission denied",
    "recovery_steps": [
        "Check directory permissions: ls -la .moai/",
        "Fix permissions if needed: chmod 755 .moai/config/",
        "Try again: /moai:0-project"
    ],
    "logs_available": true,
    "log_path": ".moai/logs/error-2025-11-19-10-30-45.log"
}
```

### Handling Agent Response

**If response.status == "success"**:
```python
Display in user's language:
"Project setup complete!

Created:
" + format_file_list(response.changes.created) + "

Your next steps:
1. " + response.next_steps[0] + "
2. " + response.next_steps[1] + "

Ready to continue?
"
```

**If response.status == "partial_success"**:
```python
Display in user's language:
"Partial success: " + response.completed_tabs.count + " of " + response.total_tabs + " settings completed.

Issue on Tab " + response.failed_tabs[0] + ":
" + response.error_message + "

Your options:
" + format_recovery_options(response.recovery_options)
```

**If response.status == "failure"**:
```python
Display in user's language:
"Setup failed: " + response.error_message + "

To fix:
" + format_recovery_steps(response.recovery_steps) + "

Need more help?
See logs: " + response.log_path
```

---

## Agent Responsibilities (Not Command's)

DIRECTIVE: These are project-manager agent's responsibilities, NOT command's:

1. **Interview Questions**
   - Asking for project name, description
   - Asking for team size, git mode
   - Asking for tech stack, deployment target

2. **Complex Mode Logic**
   - Phase workflows (INITIALIZATION has 5 phases)
   - Skill coordination (call moai-project-config-manager, moai-project-documentation)
   - Validation rules and checkpoint logic

3. **File Operations**
   - Reading/writing config files
   - Creating project documentation
   - Backup and rollback

4. **Configuration Merging**
   - Merging user responses into config
   - Handling update templates
   - Conflict resolution

5. **Language Processing**
   - Translating documentation to user's language
   - Localizing error messages
   - Managing language-specific questions

The command's job is to route to the agent and display results.
```

**Style Requirements**:
- JSON examples for response format
- Clear "success", "partial", "failure" response types
- Implementation code snippets (not pseudocode)
- Clear table of responsibilities

---

### Section 6: Supporting Skills Reference

**Purpose**: Implementers understand which skills are used and why

**Specification**:

```markdown
## Skills Used by This Command

DIRECTIVE: This command delegates all work to agent. Agent uses these skills:

### Skill 1: moai-project-language-initializer

**When agent uses it**:
- INITIALIZATION mode: Detect/confirm language
- Settings mode Tab 1: Change language

**What it does**:
- Detects system language
- Offers language selection
- Validates language code
- Returns selected language

**Why command cares**:
- Ensures all output is in user's language
- Passed to agent as context

**Reference**: See `.claude/skills/moai-project-language-initializer/SKILL.md`

---

### Skill 2: moai-project-config-manager

**When agent uses it**:
- All modes: Read current config
- INITIALIZATION: Create new config
- SETTINGS: Update config atomically
- UPDATE: Merge new templates with existing config

**What it does**:
- Reads/writes `.moai/config/config.json`
- Creates backups before writes
- Validates configuration
- Implements atomic updates
- Provides rollback on failure

**Why command cares**:
- Agent relies on this for config operations
- Returns success/failure and changes made
- Failures bubble up to command

**Reference**: See `.claude/skills/moai-project-config-manager/SKILL.md`

---

### Skill 3: moai-project-documentation

**When agent uses it**:
- INITIALIZATION: Generate product/structure/tech.md
- AUTO-DETECT: Optionally refresh documentation
- UPDATE: Update documentation to new standard

**What it does**:
- Creates .moai/project/product.md
- Creates .moai/project/structure.md
- Creates .moai/project/tech.md
- Auto-translates to user's language
- Guides interview with prompts

**Why command cares**:
- Part of successful initialization
- User gets complete project documentation
- Sets up next command (/moai:1-plan)

**Reference**: See `.claude/skills/moai-project-documentation/SKILL.md`

---

### Skill 4: moai-project-batch-questions

**When agent uses it**:
- SETTINGS mode: Load tab schema, generate questions, process responses

**What it does**:
- Loads tab definitions from schema
- Generates questions in user's language
- Loads current values for display
- Processes user responses
- Maps responses to config fields

**Why command cares**:
- Agent uses this for all SETTINGS mode questions
- Ensures consistent question structure
- Results are mapped to config updates

**Reference**: See `.claude/skills/moai-project-batch-questions/SKILL.md`

---

### Skill 5: moai-project-template-optimizer

**When agent uses it**:
- UPDATE mode: Merge new templates with existing config

**What it does**:
- Analyzes differences between old and new templates
- Preserves user customizations
- Merges new defaults intelligently
- Validates merged config

**Why command cares**:
- UPDATE mode depends on this
- Ensures updates don't lose user data
- Validates final state

**Reference**: See `.claude/skills/moai-project-template-optimizer/SKILL.md`

---

### Skill Summary

| Skill | Used By Mode | Responsibility |
|-------|---|---|
| moai-project-language-initializer | INIT, SETTINGS | Language selection |
| moai-project-config-manager | All | Config file operations |
| moai-project-documentation | INIT, AUTO-DETECT, UPDATE | Project documentation |
| moai-project-batch-questions | SETTINGS | Interview questions |
| moai-project-template-optimizer | UPDATE | Template merging |

DIRECTIVE: Command does not call these skills directly
- Agent calls them (delegation pattern)
- Agent handles their responses
- Agent reports results back to command
```

**Style Requirements**:
- Each skill has: when used, what it does, why command cares
- Summary table for quick reference
- Links to skill documentation

---

## File Structure Summary

The complete command file should have this structure:

```
.claude/commands/moai/0-project.md

YAML Frontmatter
  ├─ name: moai/0-project
  ├─ description: [specific, covering all modes]
  ├─ argument-hint: [all three cases]
  ├─ tools: [Task, AskUserQuestion]
  └─ model: haiku

# /moai:0-project - Project Setup & Configuration

## What This Command Does
[User-friendly overview]

## When Should You Use This?
[4 use case scenarios]

## Quick Start
[30-second examples]

## What You Get
[Deliverables list]

## How It Works (Technical)

### Argument Parsing & Mode Detection
[Decision tree + directives]

### Configuration State Validation
[Config validation logic]

### Mode Routing Decision
[Routing algorithm]

### Language Context Loading
[Language loading directive]

## Tool Usage Constraints

### Tool 1: Task()
[When/how to use, why]

### Tool 2: AskUserQuestion()
[When/how to use, why]

## Restrictions (What NOT to Do)
[Complete list with explanations]

## Entry-Level Error Handling

### Error Type 1: Invalid Argument
[Detection + Recovery]

### Error Type 2: Invalid JSON
[Detection + Recovery]

### Error Type 3: Permission Denied
[Detection + Recovery]

### Error Type 4: Directory Missing
[Detection + Recovery]

### Entry-Level Error Summary Table
[All errors at a glance]

## What Happens After Entry Point
[Flowchart + delegation logic]

## Delegation to project-manager Agent

### What to Pass to Agent
[Exact Task() call with context]

### What to Expect from Agent
[Success + partial + failure responses]

### Handling Agent Response
[Code snippets for each response type]

## Agent Responsibilities
[Clear list of what agent does, not command]

## Skills Used by This Command

### Skill 1: moai-project-language-initializer
[When + what + why]

### Skill 2: moai-project-config-manager
[When + what + why]

### Skill 3: moai-project-documentation
[When + what + why]

### Skill 4: moai-project-batch-questions
[When + what + why]

### Skill 5: moai-project-template-optimizer
[When + what + why]

### Skill Summary Table
[All skills at a glance]
```

---

## Key Principles

### 1. Progressive Disclosure

**First thing users see**: What does this command do? (simple language)
**Then**: When to use it? (4 scenarios)
**Then**: How to use it? (quick examples)
**Then**: What will happen? (results)

**Technical details at bottom** (not front and center)

### 2. Clear Separation of Concerns

**Command**: Entry point orchestration
**Agent**: Complex mode logic
**Skills**: File operations and specialized tasks

Each layer has clear boundaries.

### 3. Directives Are Executable

Every "DIRECTIVE:" statement can be directly implemented as code.

Not philosophical: "Handle errors gracefully"
But specific: "If JSON parse fails, check for backup"

### 4. One Source of Truth

This file IS the specification.
Not separate from code - it IS the specification that code follows.

### 5. User Experience First

Users read the first half: What/when/how
Implementers read the second half: Technical directives
But both audiences have what they need.

---

## Success Criteria

This command specification is complete when:

- [ ] YAML frontmatter specifies correct tools + model
- [ ] User-facing sections (What/When/Quick Start) are clear
- [ ] Argument parsing decision tree is unambiguous
- [ ] Mode routing is specified for all cases
- [ ] Error handling covers all entry-point errors
- [ ] Delegation to agent is specified with exact Task() call
- [ ] Expected agent responses documented (success/partial/failure)
- [ ] All tool constraints documented with rationale
- [ ] All responsibilities clearly assigned (command/agent/skill)
- [ ] Skill references link to skill documentation
- [ ] No implementation details (code stays out of specification)
- [ ] Language handling is clear (what command does, what agent does)
- [ ] All directives use "DIRECTIVE:" prefix
- [ ] No directives in separate files (all embedded here)

---

## Implementation Checklist

When implementing `.claude/commands/moai/0-project.md`:

- [ ] Read this specification completely
- [ ] Follow structure exactly (user sections first, technical last)
- [ ] Include all YAML frontmatter fields
- [ ] Add all 6 major sections
- [ ] Use decision trees for mode routing
- [ ] Include error recovery for all entry-level errors
- [ ] Show exact Task() call format
- [ ] Document expected agent responses
- [ ] List all tool constraints
- [ ] Reference all supporting skills
- [ ] Test against implementation checklist
- [ ] Verify no tool violations
- [ ] Validate decision tree completeness

---

## Testing Against This Specification

### Automated Tests

When command is implemented, verify:

1. **Argument Parsing**
   - Test: `/moai:0-project` with config → AUTO-DETECT detected ✓
   - Test: `/moai:0-project` without config → INITIALIZATION detected ✓
   - Test: `/moai:0-project setting` → Tab selection ✓
   - Test: `/moai:0-project setting 1` → Tab 1 mode ✓
   - Test: `/moai:0-project update` → UPDATE mode detected ✓
   - Test: `/moai:0-project invalid` → Error shown ✓

2. **Tool Usage**
   - Test: Command uses ONLY Task() and AskUserQuestion() ✓
   - Test: No Read/Write/Edit/Bash calls ✓
   - Test: No Skill() calls (delegated through agent) ✓

3. **Language Loading**
   - Test: Language loaded from config ✓
   - Test: Language passed to agent ✓
   - Test: Output in user's language ✓

4. **Error Handling**
   - Test: Invalid JSON config → recovery offered ✓
   - Test: Permission denied → helpful message ✓
   - Test: Missing directory → initialize mode ✓

5. **Agent Delegation**
   - Test: Correct context passed to agent ✓
   - Test: Agent responses handled correctly ✓
   - Test: Results displayed to user ✓

### Manual Tests

When command is implemented, verify:

1. **User Experience**
   - Can new user initialize project? ✓
   - Does user understand what's happening? ✓
   - Are next steps clear? ✓

2. **Configuration Results**
   - Does config.json have expected fields? ✓
   - Are project docs created? ✓
   - Is language set correctly? ✓

3. **Error Recovery**
   - Can user recover from invalid config? ✓
   - Are recovery instructions clear? ✓
   - Can user retry after error? ✓

---

**Document Version**: 1.0.0 (Specification)
**Status**: Ready for implementation
**Next Step**: Implement `.claude/commands/moai/0-project.md` following this specification
