# Alfred Architecture and Sub-Agent Invocation Exploration Report

**Date**: 2025-11-02  
**Status**: Complete  
**Scope**: Alfred's orchestration code, sub-agent invocation patterns, translation layer insertion points  

## Executive Summary

Alfred is a **declarative command orchestrator** that coordinates 10+ sub-agents through explicit `Task()` tool invocations. The architecture uses a **three-layer pattern**: Commands → Sub-agents → Skills. Commands pass user language directly to sub-agents via `{{CONVERSATION_LANGUAGE}}` template variables, enabling native multilingual support without a translation layer.

**Key Finding**: A translation layer is **NOT required** in the current architecture—Alfred already implements pass-through language forwarding. However, improvements can be made in:
1. Centralizing language context management
2. Standardizing prompt parameter passing
3. Creating reusable prompt templates

---

## Core Architecture

### 1. Orchestration Layers

Alfred operates across **four execution layers**:

```
Layer 1: Alfred Commands (User-facing)
  ↓ Task() tool with prompt
Layer 2: Sub-agents (Task-focused specialists)
  ↓ Skill() invocations
Layer 3: Skills (Reusable knowledge capsules)
  ↓ Executes work
Layer 4: Tools (Read, Write, Grep, Bash, etc.)
```

### 2. Current Language Architecture (v0.7.0+)

**Status**: Language localization FULLY IMPLEMENTED ✅

#### Layer 1: Commands → Task() Invocations

Alfred commands (`.claude/commands/alfred/*.md`) invoke sub-agents using **template variable substitution**:

```
{{CONVERSATION_LANGUAGE}}  → Language code (e.g., "ko", "en", "ja")
{{CONVERSATION_LANGUAGE_NAME}} → Display name (e.g., "Korean", "English")
```

**Example from `/alfred:1-plan` (line 570-602)**:

```
prompt: """You are spec-builder agent.

LANGUAGE CONFIGURATION:
- conversation_language: {{CONVERSATION_LANGUAGE}}
- language_name: {{CONVERSATION_LANGUAGE_NAME}}

CRITICAL INSTRUCTION:
All SPEC documents and analysis must be generated in conversation_language.
- If conversation_language is 'ko' (Korean): Generate ALL analysis in Korean
- If conversation_language is 'ja' (Japanese): Generate ALL analysis in Japanese
```

#### Layer 2: Sub-agents Handle Language Natively

Each sub-agent has a **"🌍 Language Handling"** section defining:

1. **Prompt Language**: Receives user's `conversation_language` directly
2. **Output Language**: Generates content in user's language
3. **Always English**: Technical items (@TAGs, Skill names, YAML, code)
4. **Explicit Skill Invocation**: `Skill("moai-foundation-ears")` (never keyword-matching)

**Example from `spec-builder.md` (lines 23-53)**:

```markdown
## 🌍 Language Handling

IMPORTANT: You will receive prompts in the user's **configured conversation_language**.

Alfred passes the user's language directly to you via `Task()` calls.

**Language Guidelines**:
1. Prompt Language: user's conversation_language
2. Output Language: Generate SPEC documents in user's conversation_language
3. Always in English: @TAG identifiers, Skill names, YAML frontmatter
4. Explicit Skill Invocation: Skill("moai-foundation-specs"), Skill("moai-foundation-ears")
```

---

## Sub-Agent Invocation Patterns

### 1. Task() Tool Signature

All Alfred commands invoke sub-agents using the standard **Task() pattern**:

```python
Task(
  subagent_type: str,      # Agent identifier (e.g., "spec-builder", "tdd-implementer")
  description: str,        # Action description
  prompt: str             # Full prompt (includes language config + task instructions)
)
```

### 2. Core Invocation Pattern (3 parts)

Every Task() call follows this structure:

**Part 1: Prompt Header** (lines 1-10)
```
You are [agent_name] agent.
```

**Part 2: Language Configuration** (lines 11-15)
```
LANGUAGE CONFIGURATION:
- conversation_language: {{CONVERSATION_LANGUAGE}}
- language_name: {{CONVERSATION_LANGUAGE_NAME}}

CRITICAL INSTRUCTION:
[Language-specific output requirements]
```

**Part 3: Skill Invocation Block** (lines 16-20)
```
SKILL INVOCATION:
Use explicit Skill() calls when needed:
- Skill("moai-foundation-ears") for EARS syntax
- Skill("moai-lang-python") for Python best practices
```

**Part 4: Task Instructions** (lines 21+)
```
TASK:
[Actual work instructions]
User input: $ARGUMENTS
```

### 3. Current Sub-Agent Invocation Locations

| Command | Agent | File Location | Prompt Lines |
|---------|-------|---------------|--------------|
| `/alfred:0-project` | project-manager | `.claude/commands/alfred/0-project.md` | 567-602 |
| `/alfred:1-plan` (Phase A) | Explore | `.claude/commands/alfred/1-plan.md` | 172-181 |
| `/alfred:1-plan` (Phase B) | spec-builder | `.claude/commands/alfred/1-plan.md` | 195-227 |
| `/alfred:1-plan` (Phase 2) | spec-builder | `.claude/commands/alfred/1-plan.md` | 267-292 |
| `/alfred:1-plan` (Phase 2) | git-manager | `.claude/commands/alfred/1-plan.md` | 294-320 |
| `/alfred:2-run` (Phase A) | Explore | `.claude/commands/alfred/2-run.md` | 159-169 |
| `/alfred:2-run` (Phase B) | implementation-planner | `.claude/commands/alfred/2-run.md` | 183-198 |
| `/alfred:2-run` (Phase 2) | tdd-implementer | `.claude/commands/alfred/2-run.md` | 246-279 |
| `/alfred:3-sync` | doc-syncer | `.claude/commands/alfred/3-sync.md` | TBD |

---

## Where Language Injection Happens

### Current Implementation

1. **Alfred reads config**: `config.json` → `language.conversation_language`
2. **Template substitution**: Command template applies `{{CONVERSATION_LANGUAGE}}`
3. **Task invocation**: Alfred calls `Task(prompt="...", subagent_type="...")`
4. **Sub-agent receives**: Full prompt in user's language
5. **Sub-agent processes**: Generates output in user's language

### Files Involved

| File | Purpose | Language Injection |
|------|---------|-------------------|
| `.moai/config.json` | Source of truth | Contains `language.conversation_language` |
| `.claude/commands/alfred/0-project.md` | Command definition | Uses `{{CONVERSATION_LANGUAGE}}` template |
| `.claude/commands/alfred/1-plan.md` | Command definition | Uses `{{CONVERSATION_LANGUAGE}}` template |
| `.claude/commands/alfred/2-run.md` | Command definition | Uses `{{CONVERSATION_LANGUAGE}}` template |
| `.claude/commands/alfred/3-sync.md` | Command definition | Uses `{{CONVERSATION_LANGUAGE}}` template |
| `.claude/agents/alfred/*.md` | Agent specifications | Documents language handling expectations |

---

## Key Patterns for Sub-Agent Communication

### Pattern 1: Language-Aware Prompt Structure

```markdown
# Template Pattern (from /alfred:1-plan, line 199)

Call the Task tool:
- subagent_type: "spec-builder"
- description: "Analyze the plan and establish a plan"
- prompt: """You are spec-builder agent.

LANGUAGE CONFIGURATION:
- conversation_language: {{CONVERSATION_LANGUAGE}}
- language_name: {{CONVERSATION_LANGUAGE_NAME}}

CRITICAL INSTRUCTION:
All SPEC documents and analysis must be generated in conversation_language.

SKILL INVOCATION:
Use explicit Skill() calls when needed:
- Skill("moai-foundation-specs")
- Skill("moai-foundation-ears")

TASK:
[Work instructions]

User input: $ARGUMENTS"""
```

### Pattern 2: Explicit Skill Invocation

Every agent must use explicit `Skill()` calls, NOT auto-triggering:

```markdown
# CORRECT (from tdd-implementer.md, line 44-46)
- Always use explicit syntax: Skill("moai-alfred-language-detection")
- Do NOT rely on keyword matching or auto-triggering

# Code example
Agent receives: "Write Python unit tests"
Agent calls: Skill("moai-lang-python")
Agent writes: Code in English + Status updates in user language
```

### Pattern 3: AskUserQuestion for Interaction

When user input is needed:

```python
AskUserQuestion(
    questions=[
        {
            "question": "Which language would you like to use?",
            "header": "Language",
            "options": [...]
        }
    ]
)
```

**Files**: 
- Invoked in commands at lines where user approval is needed
- Documented in `moai-alfred-interactive-questions` skill (loaded on-demand)

---

## Critical Rules (from Agent Specifications)

### Rule 1: Explicit Skill Invocation (MANDATORY)

From `spec-builder.md` (line 44-46):
```
4. Explicit Skill Invocation:
   - Always use explicit syntax: Skill("moai-foundation-specs")
   - Do NOT rely on keyword matching or auto-triggering
   - Skill names are always English
```

### Rule 2: Language Passes Directly (NO TRANSLATION)

From CLAUDE.md (Layer 1: User Conversation):
```
ALWAYS use user's conversation_language for ALL user-facing content:
- Responses to user: User's configured language
- Task prompts: User's language (passed directly to Sub-agents)
- Sub-agent communication: User's language
```

### Rule 3: English-Only Infrastructure

From CLAUDE.md (Layer 2: Static Infrastructure):
```
MoAI-ADK package and templates stay in English:
- Skill("skill-name") → Skill names always English
- .claude/skills/ → Skill content in English
- Code comments → English
- Git commit messages → English
```

### Rule 4: Batched AskUserQuestion Design

From CLAUDE.md (Alfred Command Completion Pattern):
```
Use batched AskUserQuestion calls (1-4 questions per call) to reduce user interaction turns:
- ✅ BATCHED (RECOMMENDED): 2-4 related questions in 1 call
- ❌ SEQUENTIAL (AVOID): Multiple calls for independent questions
```

---

## Recommended Enhancement Points

### Enhancement 1: Centralized Language Context Manager

**Location**: New file `src/moai_adk/core/language/context.py`

**Purpose**: Standardize how language configuration is read and passed

**Current**: Template variable substitution at command level  
**Proposed**: Python class to manage language state

```python
class LanguageContext:
    def __init__(self, config_path: Path):
        self.config = self._load_config(config_path)
        self.conversation_language = self.config.get("language", {}).get("conversation_language", "en")
    
    def get_prompt_variables(self) -> dict:
        return {
            "CONVERSATION_LANGUAGE": self.conversation_language,
            "CONVERSATION_LANGUAGE_NAME": self._get_display_name()
        }
    
    def build_language_section(self) -> str:
        """Returns the LANGUAGE CONFIGURATION section for prompts"""
        return f"""LANGUAGE CONFIGURATION:
- conversation_language: {self.conversation_language}
- language_name: {self._get_display_name()}"""
```

### Enhancement 2: Reusable Prompt Templates

**Location**: New file `.claude/templates/prompt-templates.md`

**Purpose**: Standardize Task() invocation across commands

```markdown
## Task Invocation Template

Call the Task tool:
- subagent_type: {{AGENT_TYPE}}
- description: {{DESCRIPTION}}
- prompt: """You are {{AGENT_NAME}} agent.

{{LANGUAGE_SECTION}}

CRITICAL INSTRUCTION:
{{CRITICAL_INSTRUCTION}}

SKILL INVOCATION:
{{SKILL_LIST}}

TASK:
{{TASK_DESCRIPTION}}

User input: $ARGUMENTS"""
```

### Enhancement 3: Documented Prompt Parameter Passing

**Location**: `.moai/memory/prompt-parameter-standard.md` (new file)

**Purpose**: Standardize variable names across all commands

```markdown
# Standard Prompt Parameters

## Required Parameters
- `{{CONVERSATION_LANGUAGE}}`: Language code from config
- `{{CONVERSATION_LANGUAGE_NAME}}`: Display name of language
- `$ARGUMENTS`: User input or SPEC ID
- `$EXPLORE_RESULTS`: Optional exploration results (Phase A)

## Optional Parameters
- `$SPEC_ID`: When running specific SPEC (e.g., SPEC-AUTH-001)
- `$PREVIOUS_PHASE_RESULTS`: Results from prior phase
- `$PROJECT_LANGUAGE`: Detected codebase language
```

---

## Actual Code Example: Full Invocation Flow

### /alfred:1-plan → spec-builder Invocation

**Step 1: Read Config**  
File: `.moai/config.json`
```json
{
  "language": {
    "conversation_language": "ko",
    "conversation_language_name": "한국어"
  }
}
```

**Step 2: Substitute Template Variables**  
File: `.claude/commands/alfred/1-plan.md` (line 199-227)
```
{{CONVERSATION_LANGUAGE}} → "ko"
{{CONVERSATION_LANGUAGE_NAME}} → "한국어"
```

**Step 3: Call Task Tool**
```python
Task(
  subagent_type="spec-builder",
  description="Analyze the plan and establish a plan",
  prompt="""You are spec-builder agent.

LANGUAGE CONFIGURATION:
- conversation_language: ko
- language_name: 한국어

CRITICAL INSTRUCTION:
All SPEC documents and analysis must be generated in conversation_language.
- If conversation_language is 'ko' (Korean): Generate ALL analysis in Korean

SKILL INVOCATION:
Use explicit Skill() calls when needed:
- Skill("moai-foundation-specs") for SPEC structure guidance
- Skill("moai-foundation-ears") for EARS syntax requirements

TASK:
Please analyze the project document and suggest SPEC candidates in Korean.

User input: $ARGUMENTS"""
)
```

**Step 4: Sub-agent Processes**
- Receives prompt with language config
- Reads `spec-builder.md` (lines 23-53)
- Recognizes `conversation_language: ko`
- Generates all output in Korean
- Calls Skills explicitly: `Skill("moai-foundation-ears")`

**Step 5: Output to User**
```
한국어로 SPEC 후보를 분석했습니다:
- 사용자 인증 시스템 (SPEC-AUTH-001)
- API 엔드포인트 설계 (SPEC-API-001)
...
```

---

## Files Touched by Language Configuration

### Direct Dependencies
- `/Users/goos/MoAI/MoAI-ADK/.moai/config.json` - Source of truth
- `/Users/goos/MoAI/MoAI-ADK/.claude/commands/alfred/*.md` - 4 command files
- `/Users/goos/MoAI/MoAI-ADK/.claude/agents/alfred/*.md` - 10 agent files

### Supporting Infrastructure
- `/Users/goos/MoAI/MoAI-ADK/src/moai_adk/core/template/config.py` - Config loading
- `/Users/goos/MoAI/MoAI-ADK/src/moai_adk/core/template/processor.py` - Template substitution
- `/Users/goos/MoAI/MoAI-ADK/src/moai_adk/core/project/phase_executor.py` - Initialization

### Documentation
- `/Users/goos/MoAI/MoAI-ADK/CLAUDE.md` (lines 1-400) - Language architecture definition
- `/Users/goos/MoAI/MoAI-ADK/.moai/memory/CLAUDE-AGENTS-GUIDE.md` - Agent language handling

---

## Conclusion

Alfred's language architecture is **production-ready** with direct pass-through of user language to sub-agents. No translation layer is needed. The system works by:

1. ✅ Reading user's `conversation_language` from config
2. ✅ Substituting `{{CONVERSATION_LANGUAGE}}` in command templates
3. ✅ Passing language config to sub-agents via Task() prompt
4. ✅ Sub-agents generating output in user's language
5. ✅ All infrastructure (Skills, @TAGs, code) remaining in English

**Recommended next steps**:
1. Implement `LanguageContext` class for centralized management
2. Create reusable prompt templates to reduce duplication
3. Add `.moai/memory/prompt-parameter-standard.md` documentation
4. Consider async parallel Task() invocations for performance

