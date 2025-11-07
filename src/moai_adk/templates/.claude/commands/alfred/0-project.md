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

### Step 2: Delegate to Project Manager Agent
1. **Invoke Agent**:
   ```python
   Task(
       subagent_type="project-manager",
       prompt="Modify project settings in confirmed language",
       parameters={"mode": "settings_modification", "language": confirmed_language}
   )
   ```
2. **Agent Responsibilities**:
   - Display current settings in confirmed language
   - Ask for language change option (via Skill internally)
   - Collect new values (via Skill internally)
   - Update config.json
   - Provide completion report

### Step 3: Agent Handles All Settings Interactions
**Project Manager Agent will internally**:
- Invoke `Skill("moai-project-language-initializer", mode="language_change_only")` if needed
- Invoke `Skill("moai-project-config-manager", language=confirmed_language)`
- Invoke `Skill("moai-project-batch-questions")` for user interaction
- Handle validation and error recovery

### Step 4: Exit after completion (in confirmed language)
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

### Step 2: Delegate to Project Manager Agent
1. **Invoke Agent**:
   ```python
   Task(
       subagent_type="project-manager",
       prompt="Optimize templates after moai-adk update",
       parameters={"mode": "template_update_optimization", "language": confirmed_language}
   )
   ```
2. **Agent Responsibilities**:
   - Analyze update context (backup discovery, template comparison)
   - Invoke Template Optimizer Skill internally
   - Perform smart merging
   - Generate completion report

### Step 4: Update Confirmation and Completion (in confirmed language)
1. **Display Update Results** (in confirmed language):
   ```
   ✅ **템플릿 최적화 완료!**
   📊 **업데이트된 파일**: [number]개
   🔧 **사용자 정의 유지**: [number]개
   📝 **생성된 보고서**: [report location]
   ```

2. **Auto-Translate Announcements** (CRITICAL):
   ```bash
   # Ensure announcements match current language after template update
   uv run $CLAUDE_PROJECT_DIR/.claude/hooks/alfred/shared/utils/announcement_translator.py
   ```

3. **Ask for Next Steps** (in confirmed language):
   - Option 1: "업데이트 내용 검토" → Show detailed changes
   - Option 2: "설정 수정" → Go to settings mode
   - Option 3: "종료" → End command

4. **Exit after completion**
5. **Do NOT proceed** to any other workflows
6. **End command execution**

---

## 🚀 INITIALIZATION MODE: First-time Project Setup

**When to execute**: `/alfred:0-project` with no existing config.json

### Step 1: Language-First Initialization (CRITICAL)
**IMPORTANT**: Language selection MUST happen BEFORE any other configuration.

1. **Display**: "🚀 Starting first-time project initialization..."
2. **Invoke Project Manager Agent**: Delegate to specialized agent
   ```python
   Task(
       subagent_type="project-manager",
       prompt="Initialize new project with language-first flow",
       parameters={"mode": "language_first_initialization"}
   )
   ```
3. **Agent Responsibilities**:
   - Detect/select project language
   - Conduct user interview in selected language
   - Generate project documentation
   - Invoke Skills internally as needed

### Step 2: Project Manager Executes Fresh Install
**The project-manager Agent will**:
- Invoke `Skill("moai-project-language-initializer", mode="language_first")` internally
- Invoke `Skill("moai-project-language-initializer", mode="fresh_install")` internally
- Invoke `Skill("moai-project-documentation")` internally
- Generate all project documentation
- Return completion status

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

5. **Auto-Translate Announcements** (CRITICAL - NEW):
   ```bash
   # After config.json is created, auto-translate companyAnnouncements
   uv run $CLAUDE_PROJECT_DIR/.claude/hooks/alfred/shared/utils/announcement_translator.py
   ```
   - Reads `conversation_language` from `.moai/config.json`
   - Translates 22 announcements to selected language
   - Updates `.claude/settings.json` automatically

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
1. **Delegate to Agent**:
   ```python
   Task(
       subagent_type="project-manager",
       prompt="Change project language settings",
       parameters={"mode": "language_change"}
   )
   ```
2. **Agent will**:
   - Invoke `Skill("moai-project-language-initializer", mode="language_change_only")` internally
   - Update configuration with new language
   - Return completion status

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

---

## 🌍 Language-Specific CompanyAnnouncements

### Auto-Translation Strategy

**Principle**: `.claude/settings.json` contains `companyAnnouncements` automatically translated to the user's selected language.

**How it works**:
1. During 0-project setup, user selects their preferred language
2. After language selection, ALL workflow modes (INITIALIZATION/AUTO-DETECT/UPDATE) trigger auto-translation
3. Announcement strings are translated to selected language
4. Translated announcements written to `.claude/settings.json`
5. Current language stored in `.moai/config.json` → `language.conversation_language`

### Implementation in All Workflow Modes

**Trigger Points** - Auto-translate and update `.claude/settings.json` in:

#### 1. INITIALIZATION MODE
After `Skill("moai-project-language-initializer", mode="language_first")` completes:
```bash
# Step: After language selection, translate and update announcements
uv run $CLAUDE_PROJECT_DIR/.claude/hooks/alfred/shared/utils/announcement_translator.py
```

#### 2. AUTO-DETECT MODE
After language confirmation in Step 2:
```bash
# Apply language-specific announcements from config
uv run $CLAUDE_PROJECT_DIR/.claude/hooks/alfred/shared/utils/announcement_translator.py
```

#### 3. SETTINGS MODE
When user changes language (after Config Manager skill updates config.json):
```bash
# Update announcements to match new language setting
uv run $CLAUDE_PROJECT_DIR/.claude/hooks/alfred/shared/utils/announcement_translator.py
```

#### 4. UPDATE MODE
Final step after template optimization:
```bash
# Ensure announcements match current language
uv run $CLAUDE_PROJECT_DIR/.claude/hooks/alfred/shared/utils/announcement_translator.py
```

### Supported Translation Sources

**Reference Announcements** (English baseline - 22 strings):
```
1. Start with a plan: Write down what you want to build first to avoid confusion (/alfred:1-plan)
2. ✅ 5 promises: Test-first + Easy-to-read code + Clean organization + Secure + Trackable
3. Task list: Continuous progress tracking ensures nothing gets missed
4. Language separation: We communicate in your language, computers understand in English
5. Everything connected: Plan→Test→Code→Docs are all linked together
6. ⚡ Parallel processing: Independent tasks can be handled simultaneously
7. Tools first: Find the right tools before starting any work
8. Step by step: What you want→Plan→Execute→Report results
9. Auto-generated lists: Planning automatically creates task lists
10. ❓ Ask when confused: If something isn't clear, just ask right away
11. 🧪 Automatic quality checks: Code automatically verified against 5 core principles
12. Multi-language support: Automatic validation for Python, JavaScript, and more
13. ⚡ Never stops: Can continue even when tools are unavailable
14. Flexible approach: Choose between team collaboration or individual work as needed
15. 🧹 Auto cleanup: Automatically removes unnecessary items when work is complete
16. ⚡ Quick updates: New versions detected in 3 seconds, only fetch what's needed
17. On-demand loading: Only loads current tools to save memory
18. Complete history: All steps from planning to code are recorded for easy reference
19. Bug reporting: File bug reports to GitHub in 30 seconds
20. 🩺 Health check: Use 'moai-adk doctor' to instantly check current status
21. Safe updates: Use 'moai-adk update' to safely add new features
22. 🧹 When work is done: Use '/clear' to clean up conversation for the next task
```

**Currently Supported Languages**:
- **ko** (Korean/한국어): Culturally localized translations with appropriate verb forms and expressions
- **en** (English): Baseline/reference version
- **ja** (Japanese/日本語): Formal/polite expressions suitable for Japanese audience

### Adding New Languages

When supporting a new language (e.g., **es** for Spanish):

1. **Translation Requirements**:
   - Translate all 22 announcement strings to target language
   - Preserve emoji and special characters (✅, ⚡, 🧪, 🧹, 🩺, →)
   - Maintain tone: Encouraging, action-oriented, user-friendly
   - Keep command references intact: `/alfred:1-plan`, `moai-adk doctor`, `/clear`

2. **Implementation**:
   - Add language mapping in 0-project command or language initializer
   - Create translation dictionary/storage for new language
   - Ensure `translate_announcements("es")` returns Spanish strings

3. **Validation**:
   - Test in INITIALIZATION MODE with new language selection
   - Verify announcements appear in `.claude/settings.json` with correct language
   - Confirm emoji display correctly in Claude Code UI
   - Check command references are readable in context

### User Experience Flow

```
User runs: /alfred:0-project
    ↓
Skill("moai-project-language-initializer")
    → User selects: "Korean (한국어)"
    ↓
translate_announcements("ko") returns Korean strings
    ↓
.claude/settings.json updated with:
    "companyAnnouncements": [
        "계획 우선: 혼란을 피하기 위해...",
        "✅ 5가지 약속: 테스트 우선...",
        ...
    ]
    ↓
User sees Korean announcements on Claude Code startup ✓
```