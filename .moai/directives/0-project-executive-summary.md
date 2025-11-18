---
title: "/moai:0-project Executive Summary"
version: "1.0.0"
updated: "2025-11-19"
audience: "Project managers, decision makers, quick reference"
scope: "High-level overview of command behavior and success criteria"
---

# /moai:0-project Command: Executive Summary

**One-page overview of what the command does, why it matters, and how to know when it succeeds.**

---

## What Is This Command?

The `/moai:0-project` command initializes or manages project configuration for MoAI-ADK.

**In 30 seconds**: User runs this command once to set up their project, or later to update settings. The command asks 4-16 questions in their preferred language, saves answers to configuration, and guides them to the next step.

---

## When Should Users Run It?

| Scenario | Command | Result |
|----------|---------|--------|
| First time setting up project | `/moai:0-project` | Full project initialization |
| Project exists, want to modify settings | `/moai:0-project setting` | Tab-based configuration editor |
| Want to change language only | `/moai:0-project setting tab_1_user_language` | Language change + auto-update dependent fields |
| After moai-adk package update | `/moai:0-project update` | Smart template merging + preservation of user changes |

---

## How Does It Work?

### User Experience Flow

```
User runs command
    ↓
Command detects what user wants
    ↓
Asks 4-16 questions (in user's language)
    ↓
Validates answers (catches conflicts/errors)
    ↓
Saves configuration (atomically, with backup)
    ↓
Shows what changed
    ↓
Offers next steps
    ↓
User selects what to do next
```

**Time**: 5-20 minutes depending on mode

### Language-First Principle

**Language is the PRIMARY setting:**
- User's language is confirmed/chosen FIRST
- ALL other questions asked in that language
- Changing language auto-updates dependent settings
- No re-asking of questions when language changes

### Smart Entry Point

The command is smart about what to do:
- **No config exists?** → Full initialization (INITIALIZATION mode)
- **Config exists?** → Show summary, offer options (AUTO-DETECT mode)
- **`setting` argument?** → Tab-based editor (SETTINGS mode)
- **`update` argument?** → Smart package update (UPDATE mode)

---

## Key Features

### 1. Language-First Architecture
Every decision in the user's chosen language. No language mixing.

### 2. Progressive Disclosure
- Essential questions first (User & Language, Project Info)
- Recommended questions next (Git Strategy)
- Advanced questions optional (Quality, System)
- User never overwhelmed upfront

### 3. Validation at Checkpoints
- Language consistency checked after Tab 1
- Git strategy conflicts caught after Tab 3
- All required fields verified before save
- User sees errors in their language with clear solutions

### 4. Atomic Configuration
- All changes saved together or none
- Backup created before any write
- Rollback available on error
- User sees final state, not intermediate states

### 5. Clear Change Reporting
User knows exactly what changed:
- Which fields were modified
- Before/after values
- How many total changes
- What this affects (if user wants to know)

### 6. Guided Next Steps
After completing, user is offered clear options:
- "Write specification" → `/moai:1-plan`
- "Sync documentation" → `/moai:3-sync`
- "Modify more settings" → `/moai:0-project setting`
- "Exit" → Clean shutdown

---

## Configuration Outcomes

### What Gets Configured

**Essential** (Always):
- User name
- Conversation language
- Project name
- Git mode (personal/team)

**Recommended** (Usually):
- Project description
- Project owner
- Git branching strategy
- Quality principles (TRUST 5)

**Optional** (If needed):
- GitHub automation
- MoAI system settings
- Report generation
- Custom report location

### What Gets Auto-Processed (Never Asked)

These fields update automatically when user changes language:
- `conversation_language_name` (Korean ← from "ko")
- `project.locale` (derived from language)
- `default_language` (same as conversation_language)
- `optimized_for_language` (same as conversation_language)

---

## Success Criteria: How to Know It Worked

### User-Level Success

- User understands what the command did
- Configuration matches user's intent
- User knows what changed (or didn't)
- User has clear next action
- User didn't need to edit files manually

### System-Level Success

- Configuration file written correctly
- All changes validated
- Backup created successfully
- Context saved for next command
- Zero errors or warnings

### Observable Indicators

✅ Command exits cleanly with no errors

✅ `.moai/config/config.json` is updated

✅ User sees summary of changes made

✅ User is offered next steps

✅ All output was in user's language

---

## Error Handling Philosophy

**Errors should be recoverable, not fatal:**

| What Goes Wrong | What User Sees | What Happens |
|---|---|---|
| Invalid language code | "Language not recognized. Try again?" | User can re-select language |
| Git strategy conflict | "Personal mode + develop base conflict. Fix?" | User sees suggestion, can accept/retry |
| Required field missing | "Project name required. Please enter it:" | User can provide missing value |
| Config file corrupted | "Config file invalid. Restore backup?" | User can restore or reinitialize |
| Backup fails (non-critical) | Warning message | Command completes anyway, no blocker |

**Key principle**: Non-critical failures don't block completion. User is informed, can retry if needed.

---

## Technical Architecture (Implementer View)

### Tool Usage
- **ONLY**: Task() and AskUserQuestion()
- **NEVER**: Read, Write, Edit, Bash (all delegated)

### Delegation Model
```
Command (orchestration only)
    ↓ Task()
project-manager agent (mode logic)
    ↓ Skill()
Specialized skills (file operations, validation)
    ↓ File system
Config files, documentation, backups
```

### Mode Logic
1. **Routing**: Detect command arguments, detect config state
2. **Mode Execution**: Execute selected mode
3. **Validation**: Check results at critical points
4. **Persistence**: Save config atomically
5. **Completion**: Guide next steps

### Skill Dependencies

| Skill | When Used |
|---|---|
| moai-project-language-initializer | Language selection/change |
| moai-project-config-manager | All config file operations |
| moai-project-batch-questions | Tab-based question execution |
| moai-project-template-optimizer | Package update (UPDATE mode) |
| moai-project-documentation | Auto-generate product/tech docs |

---

## Configuration Priority (Rule Hierarchy)

**When settings conflict, this is the priority**:

1. **User's explicit choice** (highest priority)
   - User selects option → That's the answer
2. **Existing config** (if user doesn't change)
   - Current settings preserved unless modified
3. **Auto-processed values** (computed, not user choices)
   - Language name, locale, etc. computed from language
4. **Package defaults** (lowest priority)
   - New defaults added only if field missing

---

## Critical Rules (Non-Negotiable)

**1. Language First**
- Confirm language before asking other questions
- Never change language without explicit user action

**2. One Mode Per Invocation**
- Detect mode, execute it, complete it
- No mode switching within single command

**3. Delegate Everything**
- No direct file operations in command
- All work delegated to agents and skills

**4. Use AskUserQuestion Always**
- Every user decision goes through AskUserQuestion
- No other mechanism for interactive choices

**5. Atomic Updates Only**
- All changes together or none
- Backup before write
- Rollback available

**6. No Emojis in Interactive Fields**
- Clean, professional appearance
- Text-only questions, options, headers

---

## User Journey Examples

### Example 1: Brand New Project

```
User: /moai:0-project
System: "This is your first time. Let's set up your project."

Q1: "What language would you like to use?"
User: "Korean"

Q2: "What's your project name?"
User: "user-auth-system"

Q3: "Brief description?"
User: "Authentication service with JWT"

Q4: "Who's the owner?"
User: "GoosLab"

Q5: "Personal or team development?"
User: "Personal"

System: "Setup complete! Here's what we configured:
- Language: Korean
- Project: user-auth-system
- Owner: GoosLab
- Mode: Personal Git workflow

Next steps: Write specification (/moai:1-plan)?"

User: "Yes" → Exits, ready for next command
```

### Example 2: Update Settings Later

```
User: /moai:0-project setting tab_2_project_info
System: "Which project settings would you like to modify?"

Q1: "Project name? (current: user-auth-system)"
User: "Keep"

Q2: "Description? (current: Authentication service...)"
User: "Change" → "Enhanced JWT authentication with refresh tokens"

Q3: "Owner? (current: GoosLab)"
User: "Keep"

System: "Updated 1 setting:
- project.description: Changed

What next: Modify another tab or exit?"

User: "Exit"
```

### Example 3: Change Language

```
User: /moai:0-project setting tab_1_user_language
System: "Language settings"

Q1: "Current language is Korean. Change?"
User: "Yes" → Select English

System: "Updated 4 settings (3 auto):
- language.conversation_language: ko → en
- language.conversation_language_name: Korean → English (auto)
- language.default_language: ko → en (auto)
- language.optimized_for_language: ko → en (auto)

All future output will be in English."
```

---

## Troubleshooting Quick Reference

| Issue | Diagnosis | Fix |
|---|---|---|
| "Config not found" | First time running | Run `/moai:0-project` to initialize |
| "Language not recognized" | User selected invalid language | Re-select from offered options |
| "Git conflict" | Personal mode + develop base | Change base to main or change mode |
| "Settings lost" | Closed before completing | Changes weren't saved. Try again. |
| "Can't edit tab" | Tab locked or permission issue | Try `/moai:0-project` without tab ID |
| "Backup failed" | Non-critical warning | Can continue. Config still updated. |

---

## Time Investment

| Mode | Setup Time | Questions | Difficulty |
|---|---|---|---|
| INITIALIZATION | 5-10 min | 4-6 | Easy |
| AUTO-DETECT | 2-3 min | 0-1 | Very easy |
| SETTINGS (1 tab) | 5-10 min | 3-4 | Easy |
| SETTINGS (all tabs) | 20-30 min | 40+ | Medium |
| UPDATE | 5-10 min | 0-1 | Easy |

---

## What Happens Behind the Scenes

**Phases**:

1. **Phase 1: Smart Entry Point (2 min)**
   - Parse command arguments
   - Detect what mode user wants
   - Load current language
   - Prepare context for execution

2. **Phase 2: Execute Mode (10-20 min)**
   - Ask user questions (in their language)
   - Validate answers
   - Delegate to skills for file operations
   - Create backup before writing
   - Save configuration atomically

3. **Phase 3: Completion & Next Steps (2 min)**
   - Show what changed
   - Present next action options
   - User selects continuation path

---

## Integration with Other Commands

**Dependency chain**:
```
/moai:0-project (Project setup)
    ↓ Configures language, project info
/moai:1-plan (Create specifications)
    ↓ Uses language, project info from config
/moai:2-run (Implement & test)
    ↓ Uses git strategy from config
/moai:3-sync (Generate documentation)
    ↓ Uses language, quality principles from config
```

All downstream commands depend on this command's configuration being correct.

---

## Summary Table

| Aspect | Detail |
|---|---|
| **Purpose** | Initialize and manage project configuration |
| **User Time** | 5-30 minutes depending on mode |
| **Questions** | 4-40 (mode dependent) |
| **Language** | User's chosen language for ALL output |
| **Validation** | At 2-3 critical checkpoints |
| **Persistence** | Atomic, with backup/rollback |
| **Error Recovery** | All errors have recovery paths |
| **Next Steps** | User guided to /moai:1-plan |
| **Success Indicator** | Config file updated, changes reported |
| **Failure Recovery** | Rollback available, can retry |

---

**Document Version**: 1.0.0
**Audience**: Non-technical stakeholders, quick reference users
**Last Updated**: 2025-11-19
**Status**: Ready for distribution

For detailed implementation guidance, see: `/moai/directives/0-project-command-directive.md`
