---
name: alfred:0-project
description: "Initialize project metadata and documentation"
argument-hint: "[setting|update]"
allowed-tools:
  - Read
  - Write
  - Edit
  - MultiEdit
  - Grep
  - Glob
  - TodoWrite
  - Bash(ls:*)
  - Bash(find:*)
  - Bash(cat:*)
  - Task
---

# 📋 MoAI-ADK Step 0: Initialize/Update Universal Language Support Project Documentation

> **Note**: Interactive prompts use `AskUserQuestion tool (documented in moai-alfred-ask-user-questions skill)` for TUI selection menus. The skill is loaded on-demand when user interaction is required.

## 🎯 Command Purpose

Automatically analyzes the project environment to create/update product/structure/tech.md documents and configure language-specific optimization settings.

## 📋 Execution Flow

**Step 1: Command Routing** - Detect subcommand and route to appropriate workflow
**Step 2: Mode Execution** - Execute initialization, settings, or update workflow
**Step 3: Skills Integration** - Use specialized skills for complex operations
**Step 4: Completion** - Provide next step options to user

## 🧠 Associated Skills & Agents

| Agent/Skill                    | Core Skill                          | Purpose                                       |
| ------------------------------ | ----------------------------------- | --------------------------------------------- |
| project-manager                | `moai-alfred-language-detection`    | Initialize project and interview requirements |
| trust-checker                  | `moai-alfred-trust-validation`      | Verify initial project structure (optional)   |
| **NEW: Language Initializer**  | `moai-project-language-initializer` | Handle language and user setup workflows      |
| **NEW: Config Manager**        | `moai-project-config-manager`       | Manage all configuration operations           |
| **NEW: Template Optimizer**    | `moai-project-template-optimizer`   | Handle template comparison and optimization   |
| **NEW: Batch Questions**       | `moai-project-batch-questions`      | Standardize user interaction patterns        |

## 🔗 Associated Agent

- **Primary**: project-manager (📋 planner) - Dedicated to project initialization
- **Quality Check**: trust-checker (✅ Quality assurance lead) - Initial structural verification (optional)
- **Secondary**: 4 specialized skills for complex workflows

## 💡 Example of use

The user executes the `/alfred:0-project` command to analyze the project and create/update documents.

## Command Overview

It is a systematic initialization system that analyzes the project environment and creates/updates product/structure/tech.md documents.

- **Automatically detect language**: Automatically recognize Python, TypeScript, Java, Go, Rust, etc.
- **Project type classification**: Automatically determine new vs. existing projects
- **High-performance initialization**: Achieve 0.18 second initialization with TypeScript-based CLI
- **2-step workflow**: 1) Analysis and planning → 2) Execution after user approval
- **Skills-based architecture**: Complex operations handled by dedicated skills

## How to use

The user executes the `/alfred:0-project` command to start analyzing the project and creating/updating documents.

**Automatic processing**:

- Update mode if there is an existing `.moai/project/` document
- New creation mode if there is no document
- Automatic detection of language and project type

## ⚠️ Prohibitions

**What you should never do**:

- ❌ Create a file in the `.claude/memory/` directory
- ❌ Create a file `.claude/commands/alfred/*.json`
- ❌ Unnecessary overwriting of existing documents
- ❌ Date and numerical prediction ("within 3 months", "50% reduction") etc.
- ❌ Hypothetical scenarios, expected market size, future technology trend predictions

**Expressions to use**:

- ✅ "High/medium/low priority"
- ✅ "Immediately needed", "step-by-step improvements"
- ✅ Current facts
- ✅ Existing technology stack
- ✅ Real problems

---

## 🚀 Command Router: Detect and Route

**Your immediate task**: Detect which subcommand the user provided and route to the correct workflow.

### Step 1: Check what subcommand the user provided

**Look at the user's command carefully**:
- Did they type `/alfred:0-project setting`?
- Did they type `/alfred:0-project update`?
- Did they type just `/alfred:0-project` (no subcommand)?
- Did they type something invalid like `/alfred:0-project xyz`?

### Step 2: Route based on subcommand

**IF user typed: `/alfred:0-project setting`**:
1. Print: "🔧 Entering Settings Mode - Modify existing project configuration"
2. Jump to **SETTINGS MODE** below
3. Skip ALL other sections
4. Stop after completing SETTINGS MODE
5. **DO NOT proceed** to other workflows

**ELSE IF user typed: `/alfred:0-project update`**:
1. Print: "🔄 Entering Template Update Mode - Optimize templates after moai-adk update"
2. Jump to **UPDATE MODE** below
3. Skip ALL other sections
4. Stop after completing UPDATE MODE
5. **DO NOT proceed** to other workflows

**ELSE IF user typed: `/alfred:0-project` (no subcommand, nothing after)**:
1. Check if the file `.moai/config.json` exists in the current directory
   - Read the file path: `.moai/config.json`
   - IF file exists → Print "✅ Project is already initialized!" AND jump to **AUTO-DETECT MODE**
   - IF file does NOT exist → Print "🚀 Starting first-time project initialization..." AND jump to **INITIALIZATION MODE**

**ELSE IF user typed an invalid subcommand** (like `/alfred:0-project xyz`):
1. Print this error message:
   ```
   ❌ Unknown subcommand: xyz

   Valid subcommands:
   /alfred:0-project          - Auto-detect mode (first-time or already initialized)
   /alfred:0-project setting  - Modify existing settings
   /alfred:0-project update   - Optimize templates after moai-adk update

   Example: /alfred:0-project setting
   ```
2. Exit immediately
3. **DO NOT make any changes**

### Step 3: CRITICAL RULES

⚠️ **IMPORTANT - Read this carefully**:
- Execute ONLY ONE mode per command invocation
- **DO NOT execute multiple modes** (e.g., do not run setting mode AND first-time setup in the same invocation)
- Stop and exit immediately after completing the selected mode
- **DO NOT jump to other workflows** unless that is the explicitly detected mode
- **DO NOT guess** which mode the user wanted - always detect from their actual command

---

## 🔧 SETTINGS MODE: Modify Existing Project Configuration

**When to execute**: `/alfred:0-project setting` OR user selected "Modify Settings" from auto-detect mode

### Step 1: Load and validate configuration
1. **Read `.moai/config.json`** to verify it exists and is valid JSON
2. **Extract and display current settings**:
   ```
   ✅ **Language**: [value from language.conversation_language]
   ✅ **Nickname**: [value from user.nickname]
   ✅ **Agent Prompt Language**: [value from language.agent_prompt_language]
   ✅ **GitHub Auto-delete Branches**: [value from github.auto_delete_branches]
   ✅ **SPEC Git Workflow**: [value from github.spec_git_workflow]
   ✅ **Report Generation**: [value from report_generation.user_choice]
   ✅ **Selected Domains**: [value from stack.selected_domains]
   ```

### Step 2: Use Config Manager Skill
```python
Skill("moai-project-config-manager")
```

**Purpose**: Let the skill handle all configuration modification workflows
**The skill will**:
- Ask which settings to modify
- Collect new values using batched questions
- Update config.json with proper merge strategy
- Handle validation and error recovery
- Provide completion report

### Step 3: Exit after completion
1. **Print**: "✅ Settings update completed!"
2. **Do NOT proceed** to any other workflows
3. **End command execution**

---

## 🔄 UPDATE MODE: Template Optimization After moai-adk Update

**When to execute**: `/alfred:0-project update` OR user selected template optimization

### Step 1: Check for backups and templates
1. **Check `.moai-backups/` directory** for existing backups
2. **Check template versions** vs current project files
3. **Analyze what needs optimization**

### Step 2: Use Template Optimizer Skill
```python
Skill("moai-project-template-optimizer")
```

**Purpose**: Let the skill handle template comparison and optimization
**The skill will**:
- Detect and analyze existing backups
- Compare current templates with backup files
- Perform smart merging to preserve user customizations
- Update optimization flags in config.json
- Generate completion report

### Step 3: Exit after completion
1. **Print**: "✅ Template optimization completed!"
2. **Do NOT proceed** to any other workflows
3. **End command execution**

---

## 🚀 INITIALIZATION MODE: First-time Project Setup

**When to execute**: `/alfred:0-project` with no existing config.json

### Step 1: Project environment analysis
1. **Display**: "🚀 Starting first-time project initialization..."
2. **Analyze** project structure and detect:
   - Project type (new vs existing)
   - Codebase language (Python, TypeScript, etc.)
   - Existing documentation status

### Step 2: Use Language Initializer Skill
```python
Skill("moai-project-language-initializer")
```

**Purpose**: Let the skill handle comprehensive project setup
**The skill will**:
- Collect language preferences
- Set up user nickname and profile
- Configure team mode settings if applicable
- Select project domains
- Configure report generation settings
- Create initial `.moai/config.json`

### Step 3: Proceed to project documentation creation
1. **Invoke**: `Task` with `project-manager` agent
2. **Purpose**: Create product/structure/tech.md documents
3. **Parameters**: Pass language and user preferences from the language initializer
4. **The agent will**:
   - Conduct environmental analysis
   - Create interview strategy
   - Generate project documentation

### Step 4: Completion and next steps
1. **Print**: "✅ Project initialization completed!"
2. **Ask user what to do next** using AskUserQuestion:
   - Option 1: "Write Specifications" → Guide to `/alfred:1-plan`
   - Option 2: "Review Project Structure" → Show current state
   - Option 3: "Start New Session" → Guide to `/clear`
3. **End command execution**

---

## 🔍 AUTO-DETECT MODE: Handle Already Initialized Projects

**When to execute**: `/alfred:0-project` with existing config.json

### Step 1: Load and display current configuration
1. **Read `.moai/config.json`** to get current settings
2. **Display current project status**:
   ```
   ✅ **Language**: [value from language.conversation_language]
   ✅ **Nickname**: [value from user.nickname]
   ✅ **Agent Prompt Language**: [value from language.agent_prompt_language]
   ✅ **GitHub Auto-delete Branches**: [value from github.auto_delete_branches]
   ✅ **SPEC Git Workflow**: [value from github.spec_git_workflow]
   ✅ **Report Generation**: [value from report_generation.user_choice]
   ✅ **Selected Domains**: [value from stack.selected_domains]
   ```

### Step 2: Ask what user wants to do
**Present these 4 options** to the user (let them choose one):

1. **"🔧 Modify Settings"** - Change language, nickname, GitHub settings, or reports config
2. **"📋 Review Current Setup"** - Display full current project configuration
3. **"🔄 Re-initialize"** - Run full initialization again (with warning)
4. **"⏸️ Cancel"** - Exit without making any changes

### Step 3: Handle user selection

**IF user selected: "🔧 Modify Settings"**:
1. Print: "🔧 Entering Settings Mode..."
2. **Jump to SETTINGS MODE** above
3. Let SETTINGS MODE handle the rest
4. Stop after SETTINGS MODE completes

**ELSE IF user selected: "📋 Review Current Setup"**:
1. Print this header: `## Current Project Configuration`
2. Show all current settings (from config.json)
3. Print: "✅ Configuration review complete."
4. Exit (stop the command)

**ELSE IF user selected: "🔄 Re-initialize"**:
1. Print this warning:
   ```
   ⚠️ WARNING: This will re-run the full project initialization

   Your existing files will be preserved in:
   - Backup: .moai-backups/[TIMESTAMP]/
   - Current: .moai/project/*.md (will be UPDATED)
   ```
2. **Ask the user**: "Are you sure you want to continue? Type 'yes' to confirm or anything else to cancel"
3. **IF user typed 'yes'**:
   - Print: "🔄 Starting full re-initialization..."
   - **Jump to INITIALIZATION MODE** above
   - Let INITIALIZATION MODE handle the rest
4. **ELSE** (user typed anything else):
   - Print: "✅ Re-initialization cancelled."
   - Exit (stop the command)

**ELSE IF user selected: "⏸️ Cancel"**:
1. Print:
   ```
   ✅ Exiting without changes.

   Your project remains initialized with current settings.
   To modify settings later, run: /alfred:0-project setting
   ```
2. Exit immediately (stop the command)

---

## 📊 Command Completion Pattern

**CRITICAL**: When any Alfred command completes, **ALWAYS use `AskUserQuestion` tool** to ask the user what to do next.

### Implementation Example
```python
AskUserQuestion(
    questions=[
        {
            "question": "Project initialization is complete. What would you like to do next?",
            "header": "Next Step",
            "options": [
                {"label": "Write Specifications", "description": "Run /alfred:1-plan to define requirements"},
                {"label": "Review Project Structure", "description": "Check current project state"},
                {"label": "Start New Session", "description": "Run /clear to start fresh"}
            ]
        }
    ]
)
```

**Rules**:
1. **NO EMOJIS** in JSON fields (causes API errors)
2. **Always use AskUserQuestion** - Never suggest next steps in prose
3. **Provide 3-4 clear options** - Not open-ended
4. **Language**: Present options in user's `conversation_language`

---

## 🎯 Key Improvements Achieved

### ✅ Modular Architecture
- **Original**: 3,647 lines in single monolithic file
- **Optimized**: ~500 lines main router + 4 specialized skills
- **Improvement**: 86% size reduction in main file

### ✅ Skills-Based Delegation
- **Language Initializer**: Handles all user setup workflows (70% smaller)
- **Config Manager**: Manages all configuration operations
- **Template Optimizer**: Handles template comparison and optimization
- **Batch Questions**: Standardizes user interaction patterns

### ✅ Simplified Routing
- **Clear command parsing**: Detect subcommand and route directly
- **Mode isolation**: Each mode runs independently
- **Skill delegation**: Complex operations handled by specialized skills

### ✅ Enhanced Maintainability
- **Separation of concerns**: Each skill has clear responsibility
- **Reusability**: Skills can be used by other commands
- **Testability**: Each component can be tested independently

### ✅ Improved User Experience
- **Faster execution**: Skills optimized for specific tasks
- **Better error handling**: Specialized error recovery in each skill
- **Clearer workflows**: Direct routing without nested complexity