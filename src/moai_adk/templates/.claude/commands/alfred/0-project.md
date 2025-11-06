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

Automatically analyzes the project environment to create/update product/structure/tech.md documents and configure language-specific optimization settings with **language-first contextual flows**.

## 🌐 Language-First Architecture (CRITICAL)

**Core Principle**: Language selection ALWAYS happens BEFORE any other configuration or operations.

### Language-First Flow Rules
1. **Language First**: Language selection is the VERY FIRST step in ANY flow
2. **Context Persistence**: Once selected, ALL subsequent interactions use that language
3. **Flow Adaptation**: Each flow (fresh install/update/settings) adapts based on language context
4. **Settings Respect**: Existing language settings are confirmed before any operations

### Contextual Flow Differentiation
| Context | Language Handling | Flow Type | Key Features |
|---------|-------------------|-----------|--------------|
| **Fresh Install** | Language selection FIRST | Installation questionnaire | Complete setup, language-aware documentation |
| **Update Mode** | Language confirmation FIRST | Update/merge questionnaire | Template optimization, language-aware updates |
| **Existing Project** | Language confirmation FIRST | Settings modification options | Language change priority, contextual settings |

## 📋 Execution Flow

**Step 1: Command Routing** - Detect subcommand and route to appropriate workflow
**Step 2: Language Context Establishment** - ALWAYS determine/confirm language FIRST
**Step 3: Contextual Flow Execution** - Execute appropriate contextual flow
**Step 4: Skills Integration** - Use specialized skills with language context
**Step 5: Completion** - Provide next step options in selected language

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
2. **IMPORTANT**: Language context will be established in SETTINGS MODE
3. Jump to **SETTINGS MODE** below
4. Skip ALL other sections
5. Stop after completing SETTINGS MODE
6. **DO NOT proceed** to other workflows

**ELSE IF user typed: `/alfred:0-project update`**:
1. Print: "🔄 Entering Template Update Mode - Optimize templates after moai-adk update"
2. **IMPORTANT**: Language context will be established FIRST in UPDATE MODE
3. Jump to **UPDATE MODE** below
4. Skip ALL other sections
5. Stop after completing UPDATE MODE
6. **DO NOT proceed** to other workflows

**ELSE IF user typed: `/alfred:0-project` (no subcommand, nothing after)**:
1. Check if the file `.moai/config.json` exists in the current directory
   - Read the file path: `.moai/config.json`
   - IF file exists → Print "✅ Project is already initialized!" AND jump to **AUTO-DETECT MODE**
   - IF file does NOT exist → Print "🚀 Starting first-time project initialization..." AND jump to **INITIALIZATION MODE**
   - **CRITICAL**: Both modes will establish language context FIRST

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

### Step 1: Language-First Settings Context
**IMPORTANT**: Always establish language context BEFORE any settings modifications.

1. **Check `.moai/config.json`** for existing language settings
2. **Language Confirmation** (in current language):
   - If no config exists → **STOP** and redirect to INITIALIZATION MODE
   - If config exists → Display current language and confirm
3. **Set Settings Language Context**: ALL settings interactions in confirmed language

### Step 2: Load and Display Current Configuration (in confirmed language)
1. **Read `.moai/config.json`** to verify it exists and is valid JSON
2. **Extract and display current settings** (in confirmed language):
   ```
   ✅ **언어**: [language.conversation_language_name]
   ✅ **닉네임**: [user.nickname]
   ✅ **에이전트 프롬프트 언어**: [language.agent_prompt_language]
   ✅ **GitHub 자동 브랜치 삭제**: [github.auto_delete_branches]
   ✅ **SPEC Git 워크플로우**: [github.spec_git_workflow]
   ✅ **보고서 생성**: [report_generation.user_choice]
   ✅ **선택된 도메인**: [stack.selected_domains]
   ```

### Step 3: Language Change Option (CRITICAL)
**Before showing other settings, offer language change first** (in confirmed language):

1. **Language Priority Question**:
   - "언어 설정을 변경하시겠습니까?" (in Korean)
   - "Would you like to change language settings?" (in English)
   - Options: "Change Language" | "Keep Current" | "Show All Settings"

2. **IF user selects "Change Language"**:
   ```python
   Skill("moai-project-language-initializer", mode="language_change_only")
   ```
   - Update language context
   - Restart settings mode in new language

3. **IF user selects "Keep Current" or "Show All Settings"**:
   - Continue with current language context
   - Proceed to Step 4

### Step 4: Use Config Manager Skill (Language-Aware)
```python
Skill("moai-project-config-manager", language=confirmed_language)
```

**Purpose**: Let the skill handle all configuration modification workflows with language context
**The skill will** (in confirmed language):
- Ask which settings to modify (using batched questions in confirmed language)
- Collect new values using batched questions in confirmed language
- Update config.json with proper merge strategy
- Handle validation and error recovery with language-appropriate messages
- Provide completion report in confirmed language

### Step 5: Exit after completion (in confirmed language)
1. **Print**: "✅ 설정 업데이트 완료!" (or equivalent in confirmed language)
2. **Offer Next Steps** (in confirmed language):
   - Option 1: "추가 설정 수정" → Continue settings mode
   - Option 2: "프로젝트 문서 생성" → Guide to `/alfred:1-plan`
   - Option 3: "종료" → End command
3. **Do NOT proceed** to any other workflows
4. **End command execution**

---

## 🔄 UPDATE MODE: Template Optimization After moai-adk Update

**When to execute**: `/alfred:0-project update` OR user selected template optimization

### Step 1: Language-First Update Context Detection
**IMPORTANT**: Always establish language context BEFORE any update operations.

1. **Check `.moai/config.json`** for existing language settings
2. **Language Confirmation** (in current language):
   - If no config exists → Run language selection FIRST
   - If config exists → Confirm current language settings
3. **Set Update Language Context**: ALL update interactions in confirmed language

### Step 2: Contextual Update Analysis
**Analyze the update context** (in confirmed language):

1. **Update Type Detection**:
   ```
   🔍 **업데이트 유형 분석 중...**
   ✅ **moai-adk 버전 변경 감지**: [version detection]
   ✅ **백업 파일 발견**: [backup analysis]
   ✅ **템플릿 변경 사항**: [template differences]
   ```

2. **Backup Discovery**:
   - Check `.moai-backups/` directory for existing backups
   - Analyze backup versions and completeness
   - Identify which backup to use for comparison

3. **Template Comparison**:
   - Check template versions vs current project files
   - Analyze what needs optimization
   - Detect user customizations vs template defaults

### Step 3: Use Template Optimizer Skill (Language-Aware)
```python
Skill("moai-project-template-optimizer", mode="update", language=confirmed_language)
```

**Purpose**: Let the skill handle template comparison and optimization with language context
**The skill will** (in confirmed language):
- Detect and analyze existing backups
- Compare current templates with backup files
- Perform smart merging to preserve user customizations
- Update optimization flags in config.json
- Generate completion report in confirmed language

### Step 4: Update Confirmation and Completion (in confirmed language)
1. **Display Update Results** (in confirmed language):
   ```
   ✅ **템플릿 최적화 완료!**
   📊 **업데이트된 파일**: [number]개
   🔧 **사용자 정의 유지**: [number]개
   📝 **생성된 보고서**: [report location]
   ```

2. **Ask for Next Steps** (in confirmed language):
   - Option 1: "업데이트 내용 검토" → Show detailed changes
   - Option 2: "설정 수정" → Go to settings mode
   - Option 3: "종료" → End command

3. **Exit after completion**
4. **Do NOT proceed** to any other workflows
5. **End command execution**

---

## 🚀 INITIALIZATION MODE: First-time Project Setup

**When to execute**: `/alfred:0-project` with no existing config.json

### Step 1: Language-First Initialization (CRITICAL)
**IMPORTANT**: Language selection MUST happen BEFORE any other configuration.

1. **Display**: "🚀 Starting first-time project initialization..."
2. **Immediate Language Selection**: Use Language Initializer Skill FIRST
   ```python
   Skill("moai-project-language-initializer", mode="language_first")
   ```
3. **Language Detection Strategy**:
   - Check environment variables (LANG, locale)
   - Detect from system settings
   - Present language options immediately
4. **Language Confirmation**: Display selected language and confirm
5. **Set Language Context**: ALL subsequent interactions MUST use selected language

### Step 2: Contextual Fresh Install Flow
**After language selection, proceed with fresh install workflow**:

```python
Skill("moai-project-language-initializer", mode="fresh_install", language=selected_language)
```

**Fresh Install Process**:
1. **User Profile Collection** (in selected language):
   - Nickname and user preferences
   - Experience level and role
   - Team vs personal mode selection

2. **Project Analysis** (language-aware):
   - Detect project type and codebase language
   - Analyze existing structure (if any)
   - Identify technology stack

3. **Comprehensive Configuration** (in selected language):
   - Team settings (if team mode)
   - Domain selection
   - Report generation preferences
   - GitHub and Git workflow configuration

4. **Create Initial Configuration**:
   - Generate complete `.moai/config.json`
   - Validate all settings
   - Set up language-specific configurations

### Step 3: Project Documentation Creation (Language-Aware)
1. **Invoke**: `Task` with `project-manager` agent
2. **Pass Language Context**: Ensure all documentation in selected language
3. **Parameters**: Language, user preferences, project context
4. **The agent will**:
   - Conduct environmental analysis
   - Create interview strategy in selected language
   - Generate project documentation in selected language

### Step 4: Completion and Next Steps (in selected language)
1. **Print**: "✅ 프로젝트 초기화가 완료되었습니다!" (or equivalent in selected language)
2. **Ask user what to do next** using AskUserQuestion (in selected language):
   - Option 1: "사양서 작성" → Guide to `/alfred:1-plan`
   - Option 2: "프로젝트 구조 검토" → Show current state
   - Option 3: "새 세션 시작" → Guide to `/clear`
3. **End command execution**

---

## 🔍 AUTO-DETECT MODE: Handle Already Initialized Projects

**When to execute**: `/alfred:0-project` with existing config.json

### Step 1: Language-First Context Detection
**IMPORTANT**: Always confirm/establish language context FIRST.

1. **Read `.moai/config.json`** to get current language settings
2. **Display Language Confirmation** (in current language):
   ```
   ✅ **현재 언어 설정**: [language.conversation_language_name]
   ✅ **대화 언어**: [language.conversation_language]
   ```
3. **Language Confirmation Question** (in current language):
   - "현재 언어 설정을 계속 사용하시겠습니까?" (in Korean)
   - "Continue using current language settings?" (in English)
   - Options: "Continue" | "Change Language" | "Show Current Settings"

### Step 2: Language Context Handling
**IF user selects "Change Language"**:
1. **Immediate Language Selection**:
   ```python
   Skill("moai-project-language-initializer", mode="language_change_only")
   ```
2. **Update Language Context**: Switch ALL subsequent interactions to new language
3. **Update Configuration**: Save new language settings
4. **Continue with new language context**

**IF user selects "Continue" or "Show Current Settings"**:
1. **Maintain Current Language Context**
2. **Proceed to Step 3** with confirmed language

### Step 3: Display Current Configuration (in confirmed language)
1. **Read `.moai/config.json`** to get all current settings
2. **Display current project status** (in confirmed language):
   ```
   ✅ **언어**: [language.conversation_language_name]
   ✅ **닉네임**: [user.nickname]
   ✅ **에이전트 프롬프트 언어**: [language.agent_prompt_language]
   ✅ **GitHub 자동 브랜치 삭제**: [github.auto_delete_branches]
   ✅ **SPEC Git 워크플로우**: [github.spec_git_workflow]
   ✅ **보고서 생성**: [report_generation.user_choice]
   ✅ **선택된 도메인**: [stack.selected_domains]
   ```

### Step 4: Ask what user wants to do (in confirmed language)
**Present these 4 options** to the user (in confirmed language):

1. **"🔧 설정 수정"** - Change language, nickname, GitHub settings, or reports config
2. **"📋 현재 설정 검토"** - Display full current project configuration
3. **"🔄 다시 초기화"** - Run full initialization again (with warning)
4. **"⏸️ 취소"** - Exit without making any changes

### Step 6: Handle user selection

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

### ✅ Language-First Architecture
- **Core Principle**: Language selection ALWAYS happens before any other configuration
- **Context Persistence**: Once selected, ALL subsequent interactions use that language
- **Flow Adaptation**: Each flow (fresh install/update/settings) adapts based on language context
- **Improvement**: Eliminates language confusion and ensures consistent user experience

### ✅ Contextual Flow Differentiation
- **Fresh Install**: Language selection → Installation questionnaire → Setup completion
- **Update Mode**: Language confirmation → Update/merge options → Optimization
- **Existing Project**: Language confirmation → Settings options or re-initialization
- **Improvement**: Clear separation between installation types with appropriate workflows

### ✅ Modular Architecture
- **Original**: 3,647 lines in single monolithic file
- **Optimized**: ~600 lines main router + 4 specialized skills
- **Improvement**: 83% size reduction in main file with enhanced functionality

### ✅ Skills-Based Delegation
- **Language Initializer**: Handles language-first project setup workflows
- **Config Manager**: Manages all configuration operations with language context
- **Template Optimizer**: Handles template comparison and optimization
- **Batch Questions**: Standardizes user interaction patterns with language support

### ✅ Enhanced User Experience
- **Language-First Interactions**: All user-facing content respects language selection
- **Contextual Workflows**: Each flow type provides appropriate options and guidance
- **Faster Execution**: Skills optimized for specific tasks with language awareness
- **Better Error Handling**: Specialized error recovery with language-appropriate messages