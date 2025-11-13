# SETTINGS MODE: Quick Reference Guide

**Document**: Quick reference for refactored `/alfred:0-project setting` command
**Date**: 2025-11-12
**Version**: 1.0.0

---

## File Location & Changes

**Modified File**: `.claude/commands/alfred/0-project.md`
- Previous: 319 lines (Phases 1-3 with direct Python code)
- Current: 617 lines (Phases 1-3 + SETTINGS MODE section)
- Change: +298 lines (+93%)
- Type: Major feature addition with architectural refactoring

---

## What Changed

### Removed (PHASE 2.5)
```python
# REMOVED: Lines 177-254
from moai_adk.core.context_manager import ContextManager
context_mgr = ContextManager(project_root)
phase_data = {...}  # Built manually in command
context_mgr.save_phase_result(phase_data)
```

**Why**: Delegated to project-manager agent

### Added (SETTINGS MODE)
```markdown
NEW: "🎭 SETTINGS MODE: Tab-Based Configuration (NEW)" section
- 5 tabs with 12 batches covering 44 settings
- Atomic config updates with backup/rollback
- Validation at 3 checkpoints
- Complete agent delegation
```

---

## Tab Structure Overview

### Quick Tabs Map

| # | Tab Name | Settings | Batches | Status | Validation |
|---|----------|----------|---------|--------|-----------|
| 1 | User & Language | 4 | 1 | REQUIRED | Language check ✓ |
| 2 | Project Info | 7 | 2 | RECOMMENDED | - |
| 3 | Git Strategy | 16 | 4 | RECOMMENDED | Git conflicts ✓ |
| 4 | Quality & Reports | 9 | 3 | OPTIONAL | - |
| 5 | System & GitHub | 8 | 2 | OPTIONAL | - |
| **TOTAL** | **5 tabs** | **44 settings** | **12 batches** | - | **3 checkpoints** |

### Tab Details

#### Tab 1: User & Language (REQUIRED)
```
Batch 1.1: Basic settings (4 questions)
  1. User name
  2. Conversation language (ko, en, ja, es)
  3. Language display name
  4. Agent prompt language

Field paths:
  - user.name
  - language.conversation_language
  - language.conversation_language_name
  - language.agent_prompt_language

Checkpoint: Language consistency validation
```

#### Tab 2: Project Info (RECOMMENDED)
```
Batch 2.1: Project metadata (4 questions)
  - project.name
  - project.description
  - project.owner
  - project.mode (personal|team)

Batch 2.2: Locale settings (3 questions)
  - project.locale
  - project.default_language
  - project.optimized_for_language

No validation checkpoint
```

#### Tab 3: Git Strategy (RECOMMENDED)
```
Batch 3.1: Personal checkpoint settings (4 questions)
  - checkpoint_enabled, checkpoint_interval
  - auto_commit, auto_push

Batch 3.2: Personal commit/branch settings (4 questions)
  - use_signed_commits, main_branch
  - feature_branch_prefix, bugfix_branch_prefix

Batch 3.3: Personal policy & Team PR (4 questions)
  - use_develop_branch, require_detailed_pr_description
  - team pr_base_branch, team develop_branch

Batch 3.4: Team GitFlow policy (4 questions)
  - use_gitflow, auto_merge_pr
  - require_code_review, version_strategy

Checkpoint: Git conflict detection
  - Personal vs Team mode validation
  - Branch naming consistency check
```

#### Tab 4: Quality & Reports (OPTIONAL)
```
Batch 4.1: Constitution settings (4 questions)
  - minimum_test_coverage
  - linting_policy, type_checking
  - require_tag_tracing

Batch 4.2: Report generation policy (4 questions)
  - report enabled, auto_create
  - format, include (multiselect)

Batch 4.3: Report storage location (1 question)
  - allowed_locations (multiselect)

No validation checkpoint
```

#### Tab 5: System & GitHub (OPTIONAL)
```
Batch 5.1: MoAI system settings (4 questions)
  - auto_update, enable_telemetry
  - debug_mode, log_level

Batch 5.2: GitHub automation settings (3 questions)
  - auto_create_pr
  - use_issue_template, use_pr_template

No validation checkpoint
```

---

## Batch Execution Flow (5-Step Pattern)

```
STEP 1: Load Tab Schema
  ├─ Read .claude/skills/moai-project-batch-questions/tab_schema.json
  ├─ Extract tab definition and batches
  ├─ Load current values from config.json
  └─ Prepare field mappings

STEP 2: Execute Batch via AskUserQuestion
  ├─ Max 4 questions per batch (API constraint)
  ├─ Show current values in questions
  ├─ Support: text_input, select_single, select_multiple, number_input
  └─ Automatic "Other" option for custom input

STEP 3: Process Responses
  ├─ Map selected options to config values
  ├─ Handle "Other" → custom user input
  ├─ Handle "Keep current" → preserve existing value
  └─ Build update object from field paths

STEP 4: Validate at Checkpoints
  ├─ Tab 1: Language consistency
  ├─ Tab 3: Git conflicts
  └─ Final: Required fields, type validation

STEP 5: Atomic Config Update
  ├─ Create backup: config.json.backup-{timestamp}
  ├─ Deep merge user updates
  ├─ Write atomically with verification
  ├─ On success: Delete backup, report
  └─ On failure: Restore backup, report error
```

---

## Entry Points (Command Examples)

### Entry Point 1: All Tabs (Recommended Flow)
```bash
/alfred:0-project setting

Executes:
  1. Tab 1 → REQUIRED foundation (language)
  2. Tab 2 → RECOMMENDED (project info)
  3. Tab 3 → RECOMMENDED (git strategy)
  4. Asks: "Continue to optional tabs?"
  5. Tab 4 → OPTIONAL (quality)
  6. Tab 5 → OPTIONAL (system)

Each tab validates independently
Final atomic update after all complete
```

### Entry Point 2: Specific Tab
```bash
/alfred:0-project setting tab_1_user_language
→ Execute Tab 1 only (4 questions, 1 batch)

/alfred:0-project setting tab_3_git_strategy
→ Execute Tab 3 only (16 questions, 4 batches, validation)
```

### Entry Point 3: Existing Modes (Unchanged)
```bash
/alfred:0-project
→ INITIALIZATION (no config.json) or AUTO-DETECT (config exists)

/alfred:0-project update
→ UPDATE MODE (template merge)
```

---

## Validation Checkpoints

### Checkpoint 1: Tab 1 (Language)
```
Check language.conversation_language is valid:
  ✓ ko, en, ja, es, (other ISO 639-1 codes)

Check agent_prompt_language consistency

Error handling:
  ❌ If invalid → Re-ask Tab 1 questions
  ❌ If conflict → Show explanation, offer fix
```

### Checkpoint 2: Tab 3 (Git Strategy)
```
Check Personal/Team mode conflicts:
  - If Personal: main_branch ≠ "develop"
  - If Team: pr_base_branch in [develop, main]

Check branch naming consistency:
  - feature_branch_prefix format valid
  - bugfix_branch_prefix format valid

Error handling:
  ⚠️ Highlight conflicts
  ⚠️ Suggest auto-fix options
  ⚠️ Let user confirm or retry
```

### Checkpoint 3: Final (Before Update)
```
Check required fields set:
  - user.name (if tab_1 executed)
  - project.name (if tab_2 executed)
  - language.conversation_language (if tab_1 executed)

Check field value types:
  - Booleans: true/false
  - Numbers: valid integers
  - Strings: non-empty if required
  - Arrays: valid format

Error handling:
  ❌ Report specific validation errors
  ❌ Don't proceed until all errors fixed
```

---

## Update Process (Atomic Pattern)

```
Load current config.json
    ↓
Create backup: config.json.backup-20251112-143025
    ↓
Deep merge user updates into current
    Preserve: Existing settings not in update
    Merge: Nested objects (git_strategy.personal.*)
    Handle: Array fields (report.include=[...])
    ↓
Validate merged config structure
    ↓
Write updated config.json atomically
    ↓
Verify write success
    ├─ SUCCESS: Delete backup, report completion
    └─ FAILURE: Restore from backup, report error
```

---

## Language Architecture

### Language First (Tab 1)
```
User runs: /alfred:0-project setting

Tab 1 asks: "Alfred와 대화할 때 사용할 언어는?"
User selects: "한국어 (ko)"

All subsequent questions in Korean:
  - Tab 2: "프로젝트 이름은?"
  - Tab 3: "기본 브랜치는?"
  - Etc.
```

### No Emojis Rule
```
❌ BAD:  question: "설정을 선택해주세요 🎯"
✅ GOOD: question: "설정을 선택해주세요"

❌ BAD:  header: "기본 정보 ⚙️"
✅ GOOD: header: "기본 정보"

Reason: JSON encoding constraints in AskUserQuestion
```

---

## Agents & Skills Integration

### Project-Manager Agent
```
Responsibilities:
  ├─ Route command to correct mode
  ├─ Load tab schema
  ├─ Execute batches via AskUserQuestion
  ├─ Process responses and validate
  └─ Create atomic config updates

Delegates to Skills:
  ├─ Skill("moai-project-batch-questions")
  ├─ Skill("moai-project-config-manager")
  ├─ Skill("moai-project-language-initializer")
  └─ Skill("moai-project-template-optimizer")
```

### Tab Schema Location
```
File: .claude/skills/moai-project-batch-questions/tab_schema.json
Version: 1.0.0
Size: 1045 lines
Contents:
  ├─ 5 tabs (with descriptions, batches)
  ├─ 12 batches (with questions, field paths)
  ├─ 44 settings (with current values, validation)
  └─ Navigation flow (completion order, checkpoints)
```

---

## Critical Rules Cheat Sheet

### MUST DO
```
✅ Execute ONLY ONE mode per invocation
✅ Confirm language context FIRST (Tab 1)
✅ Use user's conversation_language for ALL output
✅ Run validation at Tab 1, Tab 3, and Final
✅ Create atomic updates with backup/rollback
✅ Use AskUserQuestion for ALL user interaction
✅ Report all changes made
```

### MUST NOT DO
```
❌ Skip language confirmation
❌ Use direct tools (Read, Write, Edit, Bash)
❌ Execute multiple modes in one invocation
❌ Save config without atomic pattern
❌ Use emojis in AskUserQuestion
❌ Skip validation checkpoints
```

---

## Troubleshooting

### Issue: "Invalid language code"
```
Solution:
  1. Check language.conversation_language field
  2. Must be valid ISO 639-1 code (ko, en, ja, es, etc)
  3. Run Tab 1 again to fix
```

### Issue: "Git strategy conflict detected"
```
Solution:
  1. Check Personal vs Team mode setting
  2. Check pr_base_branch for Team mode (must be develop/main)
  3. Check branch naming consistency
  4. Tab 3 validation will show conflict explanation
  5. Accept auto-fix suggestion or retry
```

### Issue: "Write failed, config rolled back"
```
Solution:
  1. Check disk space
  2. Check file permissions on .moai/config/
  3. Check config.json syntax validity
  4. Backup file available: config.json.backup-{timestamp}
  5. Try again, agent will retry
```

### Issue: "Required field missing"
```
Solution:
  1. Identify missing field from error message
  2. Re-run corresponding tab
  3. Don't skip "Keep current" for required fields
  4. Some fields marked required: true in schema
```

---

## File References

### Modified Files
- `.claude/commands/alfred/0-project.md` (617 lines)

### Related Files (Integration)
- `.claude/skills/moai-project-batch-questions/tab_schema.json` (1045 lines)
- `.moai/config/config.json` (template reference)

### Documentation Files Created
- `.moai/reports/SETTINGS-MODE-REFACTOR-SUMMARY-2025-11-12.md`
- `.moai/reports/SETTINGS-MODE-VALIDATION-CHECKLIST-2025-11-12.md`
- `.moai/reports/SETTINGS-MODE-QUICK-REFERENCE-2025-11-12.md` (this file)

---

## Version Information

**Command Version**: 1.1.0
**SETTINGS MODE Version**: 2.0.0
**Tab Schema Version**: 1.0.0
**Last Updated**: 2025-11-12
**Status**: PRODUCTION READY

---

## Quick Links (In Command File)

| Section | Lines | Purpose |
|---------|-------|---------|
| Command Metadata | 1-8 | allowed-tools, argument-hint |
| Command Purpose | 22-28 | 4 execution modes |
| Execution Philosophy | 54-82 | Agent delegation diagram |
| Phase 1: Routing | 85-147 | Subcommand analysis |
| Phase 2: Execute | 150-192 | Mode handlers |
| **SETTINGS MODE** | **196-498** | **NEW: Tab-based config** |
| Phase 2.5: Context | 502-544 | Delegated to agents |
| Phase 3: Completion | 548-568 | Next steps guidance |
| Critical Rules | 571-595 | Mandatory requirements |
| Quick Reference | 598-616 | Command entry points |

---

**Quick Reference Completed**: 2025-11-12
**Status**: READY FOR DEPLOYMENT ✅
