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

# ⚒️ MoAI-ADK Step 0: Initialize/Update Project (Project Setup)

> **Interactive Prompts**: Use `AskUserQuestion` tool (invoked via `Skill("moai-alfred-ask-user-questions")`) for TUI-based user interaction. The skill provides all question specifications and validation.

<!-- @CODE:ALF-WORKFLOW-000:CMD-PROJECT -->

**4-Step Workflow Integration**: This command implements Step 0 of Alfred's workflow (Project Bootstrap). See CLAUDE.md for full workflow details.

---

## 🎯 Command Purpose

Initialize or update project metadata with **language-first architecture**. Supports three execution modes:
- **INITIALIZATION**: First-time project setup
- **AUTO-DETECT**: Already initialized projects (modify settings or re-initialize)
- **UPDATE**: Template optimization after moai-adk package update

---

## 🧠 Associated Skills & Agents

| Agent/Skill | Purpose |
|---|---|
| project-manager | Orchestrates language-first initialization and configuration |
| moai-project-language-initializer | Language selection and initialization workflows |
| moai-project-config-manager | Configuration management with language context |
| moai-project-batch-questions | Standardizes user interaction patterns |

---

## 🌐 Language-First Architecture

**Core Principle**: Language selection ALWAYS happens BEFORE any other configuration.

- **Initialization**: Language selection → Project interview → Documentation
- **Auto-Detect**: Language confirmation → Settings options
- **Update**: Language confirmation → Template optimization
- **Settings**: Language context → Configuration modification

---

## 🚀 PHASE 1: Command Routing & Analysis

**Goal**: Detect subcommand and prepare execution context.

### Step 1: Route Based on Subcommand

**Analyze the command user provided**:

1. **`/alfred:0-project setting`** → SETTINGS MODE
2. **`/alfred:0-project update`** → UPDATE MODE
3. **`/alfred:0-project`** (no args):
   - Check if `.moai/config.json` exists
   - Exists → AUTO-DETECT MODE
   - Missing → INITIALIZATION MODE
4. **Invalid subcommand** → Show error and exit

### Step 2: Invoke Project Manager Agent

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

  **For SETTINGS**:
  - Confirm language context first
  - Present modification options in user's language
  - Invoke Skill("moai-project-config-manager") for updates

  **For UPDATE**:
  - Confirm language context
  - Invoke Skill("moai-project-template-optimizer") for smart merging
  - Update templates and configuration

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

**SETTINGS MODE**:
- Language confirmation
- Present modification options
- Handle language change if requested
- Update `.moai/config.json`

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

## 🔒 PHASE 3: Completion & Next Steps

**Goal**: Guide user to next action in their selected language.

### Step 1: Display Completion Status

Show mode-specific completion message in user's language:
- **INITIALIZATION**: "✅ Project initialization complete"
- **AUTO-DETECT**: Configuration review/modification complete
- **SETTINGS**: "✅ Settings updated successfully"
- **UPDATE**: "✅ Templates optimized and updated"

### Step 2: Offer Next Steps

Use AskUserQuestion in user's language:
- **From Initialization**: Write SPEC / Review Structure / New Session
- **From Settings**: Continue Settings / Sync Documentation / Exit
- **From Update**: Review Changes / Modify Settings / Exit

---

## 📋 Critical Rules

**MANDATORY**:
- ❌ Execute ONLY ONE mode per invocation
- ❌ Never skip language confirmation/selection
- ✅ Always use user's `conversation_language` for all output
- ✅ Auto-translate announcements after language changes
- ✅ Route to correct mode based on command analysis

**Configuration Priority**:
- `.moai/config.json` settings ALWAYS take priority
- Existing language settings respected unless user requests change
- Fresh installs: Language selection FIRST, then all other config

---

## 📚 Quick Reference

| Scenario | Mode | Entry Point | Key Agent |
|---|---|---|---|
| First-time setup | INITIALIZATION | No config.json | project-manager |
| Existing project | AUTO-DETECT | /alfred:0-project | project-manager |
| Modify config | SETTINGS | /alfred:0-project setting | project-manager |
| After package update | UPDATE | /alfred:0-project update | project-manager |

**Associated Skills**:
- `Skill("moai-project-language-initializer")` - Language selection
- `Skill("moai-project-config-manager")` - Config operations
- `Skill("moai-project-template-optimizer")` - Template merging
- `Skill("moai-project-batch-questions")` - User interaction

**Version**: 1.0.0 (Optimized Agent-Delegated Pattern)
**Last Updated**: 2025-11-09
**Total Lines**: ~380 (reduced from 637)
**Architecture**: Commands → Agents → Skills
