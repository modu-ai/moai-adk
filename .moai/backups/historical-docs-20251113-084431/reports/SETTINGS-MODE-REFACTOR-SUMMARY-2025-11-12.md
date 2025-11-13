# SETTINGS MODE Refactor: Tab-Based Configuration Implementation

## Executive Summary

Successfully refactored `/alfred:0-project` command to implement new tab-based batch question system for configuration management. Command now supports 5 organized tabs covering 44 settings across 12 batches with atomic update capabilities and critical checkpoint validation.

**Refactor Status**: COMPLETE
**Date**: 2025-11-12
**Version**: 1.1.0 (SETTINGS MODE v2.0.0)
**Lines Modified**: 617 (previous 319)

---

## Changes Overview

### 1. Command Metadata Updates

**File**: `.claude/commands/alfred/0-project.md`

#### Removed (PHASE 2.5 Simplification)
- Lines 177-254: Direct Python code execution (ContextManager import, path validation)
- Direct tool usage for context saving
- Manual phase data building

**Why**: Phase 2.5 logic now delegated to project-manager agent using Skill("moai-project-config-manager")

#### Changed
- `allowed-tools`: Removed Read, Write, Edit, MultiEdit, Grep, Glob, Bash
- `allowed-tools`: Added AskUserQuestion (for user interaction)
- `allowed-tools`: Kept Task (for agent orchestration)
- `argument-hint`: Changed from `[setting|update]` to `[setting [tab_ID]|update]`

### 2. Architecture Clarification

**Removed**:
- Ambiguous PHASE 1-3 structure referencing direct file operations
- Configuration instructions mixing command and agent responsibilities

**Added**:
- Clear "Key Principle: Zero Direct Tool Usage" section
- Task flow diagram showing complete delegation
- Explicit emphasis on "Task() and AskUserQuestion() ONLY"

### 3. New SETTINGS MODE Implementation

**Added Section**: "SETTINGS MODE: Tab-Based Configuration (NEW)"

#### Tab Schema Reference
- Location: `.claude/skills/moai-project-batch-questions/tab_schema.json`
- 5 tabs with clear organizational hierarchy:
  - Tab 1: User & Language (REQUIRED, 4 settings)
  - Tab 2: Project Info (RECOMMENDED, 7 settings)
  - Tab 3: Git Strategy (RECOMMENDED with validation, 16 settings)
  - Tab 4: Quality & Reports (OPTIONAL, 9 settings)
  - Tab 5: System & GitHub (OPTIONAL, 8 settings)

#### Batch Execution Flow (5 Steps)

1. **Load Tab Schema**
   - Read tab definition and batches
   - Extract field mappings to config.json paths
   - Load current values from existing config

2. **Execute Batch via AskUserQuestion**
   - Max 4 questions per batch (per API constraints)
   - Show current values in question text
   - Support text_input, select_single, select_multiple, number_input types

3. **Process Responses**
   - Map selected options to config values
   - Handle "Other" custom input option
   - Handle "Keep current" to preserve existing values
   - Build update object from field paths

4. **Validate at Checkpoints**
   - **Tab 1**: Language consistency validation
   - **Tab 3**: Personal/Team Git conflict validation
   - **Final**: Required field checks, type validation

5. **Atomic Config Update**
   - Create backup: `config.json.backup-{timestamp}`
   - Deep merge user updates into current config
   - Write atomically with verification
   - Delete backup on success or restore on failure

### 4. Implementation Details & Examples

#### Single Tab Example: Tab 1 User & Language
```
Step 1: Project-manager loads tab schema
Step 2: Extracts Tab 1 (tab_1_user_language)
Step 3: Gets Batch 1.1 (기본설정)
Step 4: Loads current config values
Step 5: Calls AskUserQuestion with 4 questions
Step 6: Processes responses (map to config fields)
Step 7: Runs Tab 1 validation (language consistency)
Step 8: Creates atomic update (backup + deep merge)
Step 9: Writes config.json and reports success
```

#### Complex Tab Example: Tab 3 Git Strategy
```
Step 1: Load Tab 3 with 4 batches
Step 2-5: Execute each batch (Personal checkpoints, commit, policy, Team GitFlow)
Step 6: Run Tab 3 validation checkpoint
  - Detect Personal/Team conflicts
  - Check branch naming consistency
  - Highlight issues for user confirmation
Step 7: Merge all 4 batches into single update
Step 8: Create atomic update
Step 9: Report all 16 settings changes
```

#### Multi-Tab Workflow: All Tabs
```
Recommended flow:
  1. Tab 1 (REQUIRED) → Language foundation
  2. Tab 2 (RECOMMENDED) → Project metadata
  3. Tab 3 (RECOMMENDED with validation) → Git workflow
  4. Ask: "Continue to optional tabs?"
  5. Tab 4 (OPTIONAL) → Quality settings
  6. Tab 5 (OPTIONAL) → System settings

Each tab independent:
  - If user cancels mid-tab: changes not saved
  - If validation fails: user can retry
  - Final atomic update only after all selected tabs complete
```

### 5. Phase 2.5 Refactoring: Delegation Pattern

**Previous Approach** (Direct execution in command):
```python
# Lines 177-254: Direct Python code
from moai_adk.core.context_manager import ContextManager
context_mgr = ContextManager(project_root)
phase_data = {...}  # Built manually
context_mgr.save_phase_result(phase_data)
```

**New Approach** (Agent delegation):
```markdown
Agent delegates to Skill("moai-project-config-manager"):
  - Save context via ContextManager
  - Handle file path validation
  - Implement error recovery (non-blocking)
  - Report success/failure

Error Handling:
  - Context save failures should NOT block command completion
  - Log clear warning messages
  - Allow user to retry manually if needed
```

**Benefits**:
- Command remains thin orchestration layer
- All complexity in specialized agents
- Consistent with 2-run pattern
- Easier to test and maintain

### 6. Language Architecture Updates

**Tab 1 is Critical for Language**:
- User confirms conversation_language FIRST
- All subsequent user-facing questions use this language
- AskUserQuestion must use user's conversation_language for ALL fields

**No Emojis in AskUserQuestion**:
- Removed all emoji usage in question/header/label/description fields
- Clear text only (per JSON encoding constraints)
- Examples use Korean (ko): "사용자 이름", "대화 언어", etc.

### 7. Critical Rules Section

**Added explicit guidance**:
- Execute ONLY ONE mode per invocation
- ALWAYS confirm language context first
- Run validation at Tab 1, Tab 3, final update
- Create atomic config update with backup/rollback
- Report all changes
- Use AskUserQuestion for ALL user interaction

**No Direct Tool Usage**:
```
Forbidden:
  ❌ NO Read (delegated)
  ❌ NO Write (delegated)
  ❌ NO Edit (delegated)
  ❌ NO Bash (delegated)
  ❌ NO TodoWrite (delegated)

Allowed:
  ✅ Task() for orchestration
  ✅ AskUserQuestion() for interaction
```

---

## Integration Points

### 1. Tab Schema Integration

**Location**: `.claude/skills/moai-project-batch-questions/tab_schema.json`

**Used by**: Project-manager agent via Skill("moai-project-batch-questions")

**Data Flow**:
```
Command: /alfred:0-project setting tab_1
    ↓
Task(subagent_type="project-manager")
    ↓
Load tab_schema.json (via skill)
    ↓
Execute Batch 1.1 (4 questions)
    ↓
AskUserQuestion (user interaction)
    ↓
Process responses + validate + atomic update
    ↓
Report success
```

### 2. Project-Manager Agent Integration

**Responsibilities**:
- Route command to correct mode (INIT/AUTO-DETECT/SETTINGS/UPDATE)
- Load tab schema from skill
- Execute batch questions
- Process responses and map to config fields
- Validate at checkpoints
- Create atomic config updates
- Report changes and next steps

**Delegation to Skills**:
- `Skill("moai-project-batch-questions")` - Tab schema access
- `Skill("moai-project-config-manager")` - Config operations
- `Skill("moai-project-language-initializer")` - Language selection
- `Skill("moai-project-template-optimizer")` - Template merging

### 3. Backup & Rollback Strategy

**Safe Update Pattern**:
```
1. Load current config.json
2. Create backup: config.json.backup-20251112-143025
3. Deep merge user updates
4. Validate merged config structure
5. Write updated config.json atomically
6. Verify write success
   - Success: Delete backup, report completion
   - Failure: Restore backup, report error
```

**User-Facing**:
- No manual backup management required
- Automatic rollback on write failure
- Backup timestamp for debugging

### 4. Atomic Update with Deep Merge

**Merge Logic**:
- Preserve existing settings not in update
- Recursively merge nested objects (e.g., git_strategy.personal.*)
- Support array fields (e.g., report_generation.include)
- Validate final config matches schema

**Example Merge**:
```json
Current config:
{
  "user": {"name": "GoosLab"},
  "project": {"name": "moai-adk", "owner": "GoosLab"}
}

User update:
{"user": {"name": "NewName"}}

Result:
{
  "user": {"name": "NewName"},
  "project": {"name": "moai-adk", "owner": "GoosLab"}
}
```

---

## Command Entry Points

### Existing Entry Points (Preserved)

1. **First-time setup**:
   ```bash
   /alfred:0-project
   # No config.json → INITIALIZATION MODE
   ```

2. **Existing project**:
   ```bash
   /alfred:0-project
   # config.json exists → AUTO-DETECT MODE
   ```

3. **Package update**:
   ```bash
   /alfred:0-project update
   # → UPDATE MODE
   ```

### New Entry Points (SETTINGS MODE)

4. **Interactive all tabs**:
   ```bash
   /alfred:0-project setting
   # → Load all 5 tabs, recommended flow with validation
   ```

5. **Specific tab**:
   ```bash
   /alfred:0-project setting tab_1_user_language
   # → Execute Tab 1 (language settings)

   /alfred:0-project setting tab_3_git_strategy
   # → Execute Tab 3 (Git strategy with 4 batches)
   ```

---

## Validation & Error Handling

### Checkpoint Validation

#### Tab 1 Checkpoint (Language Settings)
```
Validations:
  - conversation_language: Valid codes (ko, en, ja, es, etc)
  - agent_prompt_language: Consistent with conversation
  - Language display name: Matches selected language

Error Recovery:
  - If invalid: Re-ask Tab 1 questions
  - If conflict: Show explanation, offer fix
```

#### Tab 3 Checkpoint (Git Strategy)
```
Validations:
  - Personal mode: main_branch != "develop"
  - Team mode: PR base in [develop, main]
  - Branch naming: Prefix consistency (feature/, bugfix/)
  - GitFlow enforcement: use_gitflow alignment

Error Recovery:
  - Highlight conflicts
  - Show auto-fix suggestions
  - Let user confirm or retry
```

#### Final Validation (Before Config Update)
```
Validations:
  - All required fields set (from schema)
  - No conflicting settings
  - Field value types correct (string, bool, number, array)
  - Config structure matches expected format

Error Recovery:
  - Report specific validation errors
  - Don't proceed with update until fixed
```

---

## File Modifications Summary

### Primary File
- **Path**: `/Users/goos/MoAI/MoAI-ADK/.claude/commands/alfred/0-project.md`
- **Lines**: 617 (previous: 319)
- **Change Type**: Major refactor
- **Breaking Changes**: None (backwards compatible entry points)

### Related Files (No Changes Required)
- `.claude/skills/moai-project-batch-questions/tab_schema.json` (exists, used by command)
- `.moai/config/config.json` (template, updated via command)
- Project-manager agent (existing, enhanced with new SETTINGS flow)

---

## Testing Recommendations

### Entry Point Testing
1. `/alfred:0-project` with existing config → AUTO-DETECT mode
2. `/alfred:0-project setting` → Load all tabs
3. `/alfred:0-project setting tab_1_user_language` → Tab 1 only
4. `/alfred:0-project setting tab_3_git_strategy` → Tab 3 with validation
5. `/alfred:0-project update` → UPDATE mode

### Scenario Testing
- **Tab 1**: Change language from ko to en, verify all subsequent interactions in English
- **Tab 3**: Set Personal mode, then try to set PR base to develop → Should warn about conflict
- **Multi-tab**: Execute Tab 1 + Tab 2 + Tab 3, then skip Tab 4/5 → Should save all three
- **Rollback**: Simulate write failure, verify backup is restored

### Validation Testing
- **Language**: Invalid language code → Should re-ask Tab 1
- **Git conflicts**: Team mode with GitFlow disabled → Should highlight conflict
- **Required fields**: Skip required field → Should block update until set
- **Type validation**: Set number field to string → Should reject or convert

---

## Documentation & Communication

### Updates to CLAUDE.md Project Template
```
No updates needed - command references existing patterns:
  - Task(subagent_type="...") delegation
  - AskUserQuestion for user interaction
  - Language architecture enforcement
  - PHASES pattern
```

### User-Facing Documentation
All guidance built into command via:
1. Tab schema descriptions (easy_description field)
2. Validation error messages (user-friendly)
3. Next steps guidance (AskUserQuestion)
4. Change reports (what was updated)

### Developer Documentation
- Tab execution flow documented with examples
- Batch processing logic explained step-by-step
- Checkpoint validation rules detailed
- Atomic update pattern with backup/rollback documented

---

## Alignment with CLAUDE.md Requirements

### Requirement 1: Language-First Architecture
✅ **Met**: Tab 1 executes FIRST, user confirms language, all subsequent output uses that language

### Requirement 2: AskUserQuestion Tool
✅ **Met**: All user interaction through AskUserQuestion in Korean (ko)

### Requirement 3: Agent Delegation
✅ **Met**: All execution delegated to project-manager via Task()

### Requirement 4: No Direct Tool Usage
✅ **Met**: Only Task() and AskUserQuestion() used in command

### Requirement 5: Phase 2.5 Context Saving
✅ **Met**: Delegated to project-manager agent using Skill("moai-project-config-manager")

### Requirement 6: Configuration Priority
✅ **Met**: .moai/config/config.json always takes priority, merged with deep merge logic

---

## Version Information

**Command Version**: 1.1.0
**SETTINGS MODE Version**: 2.0.0
**Tab Schema Version**: 1.0.0
**Last Updated**: 2025-11-12
**Architecture**: Commands → Agents → Skills (Complete delegation)

**Components**:
- Command: `.claude/commands/alfred/0-project.md`
- Tab Schema: `.claude/skills/moai-project-batch-questions/tab_schema.json`
- Supporting Skills: moai-project-batch-questions, moai-project-config-manager, etc.
- Primary Agent: project-manager

---

## Future Enhancements

### Potential v2.1.0 Updates
1. Tab skip/optional fields based on project mode (Personal vs Team)
2. Conditional batch visibility (show batch X only if setting Y = Z)
3. Settings preview before atomic update
4. Partial tab retry (don't lose all batches on single validation failure)
5. Setting import/export (JSON config sharing)

### Potential v3.0.0 (Major)
1. Real-time setting recommendations (suggest best practices)
2. Settings wizard (step-by-step guided config)
3. Multi-language prompt translation (prompts auto-translated to user language)
4. Interactive validation with auto-fix suggestions
5. Settings sync across team members

---

## Conclusion

Successfully refactored `/alfred:0-project` command to implement tab-based batch question system for organized configuration management. Command maintains backward compatibility while adding powerful new SETTINGS MODE with:

- 5 organized tabs covering 44 settings
- 12 batches for grouped questions
- Atomic updates with backup/rollback
- Critical checkpoint validation
- Complete agent delegation (no direct tool usage)
- Language-first architecture enforcement

All changes follow established Alfred patterns from 2-run reference implementation and maintain consistency with MoAI-ADK architectural principles.
