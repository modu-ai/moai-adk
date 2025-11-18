---
title: "/moai:0-project Command Directive Guide"
version: "1.0.0"
updated: "2025-11-19"
audience: "Users + Implementers"
scope: "Command orchestration, configuration philosophy, user experience"
---

# /moai:0-project Command Directive Guide

**A comprehensive guide defining HOW the /moai:0-project command SHOULD WORK, from user perspective and implementer perspective.**

---

## Section 1: Core Philosophy & Principles

### 1.1 Fundamental Directives

**DIRECTIVE 1: Language-First Architecture**
- Language is the PRIMARY configuration dimension
- All other settings are secondary to language choice
- Language MUST be established before asking domain-specific questions
- Language change MUST immediately trigger auto-processing of dependent settings
- User's conversation language drives ALL output language

**DIRECTIVE 2: Zero Direct Tool Usage**
- This command uses ONLY Task() and AskUserQuestion()
- NO Read, Write, Edit, or Bash operations
- ALL file operations delegated to agents
- Complex operations delegated to specialized skills
- Command acts as pure orchestrator, not executor

**DIRECTIVE 3: Complete Agent Delegation**
- Command → project-manager agent (orchestration)
- project-manager → specialized skills (execution)
- Skills → file operations, config updates, validation
- No work happens in the command itself
- Agent response becomes the completion status

**DIRECTIVE 4: Progressive Disclosure**
- Essential configuration first (User & Language)
- Recommended configuration second (Project Info, Git Strategy)
- Advanced configuration last (Quality, System)
- User never overwhelmed with options upfront
- Each phase builds understanding for next phase

**DIRECTIVE 5: User-Centric Language**
- All questions and responses in user's conversation language
- No technical jargon unless explained
- Clear benefit statement for each setting
- "Why this matters" included implicitly or explicitly
- No emojis in AskUserQuestion fields

**DIRECTIVE 6: Atomic Configuration Updates**
- Each configuration change is atomic (all-or-nothing)
- Backup created before any update
- Rollback available if validation fails
- User sees exact changes made
- No partial updates visible to user

### 1.2 Success Criteria

**A successful /moai:0-project execution MUST deliver**:
- Language properly configured and persisted
- Project metadata aligned with user intent
- Configuration validated at critical checkpoints
- Clear next steps presented to user
- Session context saved for next command

---

## Section 2: Smart Entry Point Directive (Phase 1)

### 2.1 Command Analysis & Routing

**DIRECTIVE: Analyze user input and route to exactly ONE mode**

The command receives:
- No arguments → Auto-detect mode
- `setting [tab_id]` → Settings mode
- `setting` (no tab_id) → Tab selection screen
- `update` → Update mode
- Invalid arguments → Clear error message

**DIRECTIVE: Detect initialization state**

Before routing, check:
1. Does `.moai/config/config.json` exist?
   - EXISTS → Auto-detect mode
   - MISSING → Initialization mode
2. Is it valid JSON?
   - YES → Continue to mode execution
   - NO → Offer recovery (restore backup or reinit)

**DIRECTIVE: Load language context early**

BEFORE any user interaction:
- Read current language from config.json (if exists)
- Pass language to project-manager agent
- Project-manager uses this language for ALL output
- If config missing, use CLI default (from moai-adk init)

### 2.2 Route Execution

**DIRECTIVE: Delegate to project-manager agent with mode context**

Call Task() with:
```
subagent_type: "project-manager"
prompt: """
Mode: [INITIALIZATION|AUTO-DETECT|SETTINGS|UPDATE]
Language context: [language code from config]
User request: [parsed command arguments]

Execute this mode with complete workflow:
[mode-specific instructions]
"""
```

Store response in execution context for Phase 2.

---

## Section 3: Essential Configuration Directive (Phase 2)

### 3.1 Initialization Mode

**When**: User runs `/moai:0-project` and no config exists

**DIRECTIVE: Establish language foundation first**

The project-manager agent MUST:
1. Detect language from CLI (default from moai-adk)
2. Confirm with user: "Is your conversation language correct?"
3. Offer language selection if needed
4. DO NOT proceed to other questions until language confirmed
5. Store language before asking anything else

**DIRECTIVE: Conduct language-aware user interview**

After language confirmed:
- ALL questions must be in selected language
- Use AskUserQuestion for ALL user interaction
- Maximum 3-4 questions per batch
- Questions ordered: essential → recommended
- Skip advanced questions in initialization mode

**Minimum essential interview questions**:
1. Project name
2. Project description
3. Project owner/team
4. Preferred Git mode (personal/team)

**DIRECTIVE: Auto-process dependent settings**

After user confirms language:
- Auto-set: `project.locale` (derived from language)
- Auto-set: `language.default_language` (same as conversation_language)
- Auto-set: `language.optimized_for_language` (same as conversation_language)
- DO NOT ask user for these (they're computed)

**DIRECTIVE: Generate project documentation**

After interview complete:
- Create `.moai/project/` directory structure
- Generate `product.md`, `structure.md`, `tech.md`
- Auto-translate documentation to user's language
- Save to config for reference

**DIRECTIVE: Save completion context**

Before returning:
- Extract mode result (INITIALIZATION completed)
- Extract created files (list with absolute paths)
- Extract next recommended action (usually /moai:1-plan)
- Delegate context save to project-manager agent

### 3.2 Auto-Detect Mode

**When**: User runs `/moai:0-project` and config exists

**DIRECTIVE: Display current configuration in user's language**

The project-manager agent MUST:
1. Read language from config.json
2. Display current project status in that language:
   - Project name
   - Current language
   - Git mode (personal/team)
   - Last modified timestamp
3. Show configuration summary

**DIRECTIVE: Offer clear action options**

Present user with menu in their language:
- "Modify Settings" → Go to SETTINGS mode
- "Change Language Only" → Go directly to Tab 1 settings
- "Review Configuration" → Show detailed config
- "Re-initialize" → Start over (warning: loses current config)
- "Cancel" → Exit without changes

**DIRECTIVE: Execute user's selection**

- "Modify Settings" → Ask which tab to modify
- "Change Language Only" → Execute SETTINGS mode, Tab 1 only
- "Review Configuration" → Display full config.json
- "Re-initialize" → Confirm, then reinitialize
- "Cancel" → Exit gracefully

**DIRECTIVE: No breaking changes without confirmation**

Any action that changes config MUST:
- Show preview of changes
- Ask for explicit confirmation
- Create backup before writing
- Show success/failure clearly

### 3.3 Settings Mode (Tab-Based Configuration)

**When**: User runs `/moai:0-project setting [tab_id]`

**DIRECTIVE: Read language before starting**

The project-manager agent MUST:
1. Read current language from config.json
2. Load tab schema from `.claude/skills/moai-project-batch-questions/tab_schema.json`
3. Translate all questions to user's language
4. Load current values from config for display

**DIRECTIVE: Tab Selection (if no tab_id provided)**

If user runs `/moai:0-project setting` (no tab specified):
1. Show tab selection screen in user's language
2. Display 5 tabs with descriptions
3. Add "Modify All Tabs" option
4. Wait for user selection
5. Execute selected tab(s)

**Tab Selection Screen**:
```
Which settings tab would you like to modify?

1. Tab 1: User & Language (REQUIRED)
   - User name, conversation language, agent prompt language

2. Tab 2: Project Basic Information (RECOMMENDED)
   - Project name, description, owner, mode

3. Tab 3: Git Strategy & Workflow (RECOMMENDED)
   - Personal/Team settings, commit strategy

4. Tab 4: Quality Principles & Reports (OPTIONAL)
   - TRUST 5, report generation

5. Tab 5: System & GitHub Integration (OPTIONAL)
   - MoAI system, GitHub automation

6. Modify All Tabs (RECOMMENDED ORDER: Tab 1 → 2 → 3 → 4 → 5)
```

**DIRECTIVE: Execute batch-based questions**

For each selected tab:
1. Load batch definition from tab schema
2. Call AskUserQuestion with batch questions
3. Maximum 4 questions per batch
4. Wait for user responses
5. Process responses to config field mappings
6. Move to next batch in same tab
7. Run validation checkpoint for tab

**DIRECTIVE: Validate at critical checkpoints**

After Tab 1 (Language):
- Verify language code is valid
- Verify agent_prompt_language consistency
- If invalid: explain error, re-ask Tab 1

After Tab 3 (Git Strategy):
- Check Personal/Team conflict (e.g., Personal with develop main)
- Check branch naming consistency
- If invalid: highlight conflict, offer fix suggestions

Before Final Update:
- Verify all required fields present
- Verify no conflicting settings
- Verify correct data types
- Report validation results

**DIRECTIVE: Delegate atomic config update to skill**

After all selected tabs complete:
1. Collect all updates from all tabs
2. Call Skill("moai-project-config-manager"):
   - Pass collected updates
   - Request atomic deep merge
   - Include backup/rollback logic
3. Skill handles file operations
4. Command receives success/failure
5. Report changes to user

**DIRECTIVE: Handle multi-tab workflow**

After tab completion, ask:
"Would you like to modify another settings tab?"
- Option 1: No, finish settings
- Option 2: Select another tab
- Option 3: Cancel (discard changes)

If Option 2: Return to tab selection screen

**DIRECTIVE: Report all changes**

When settings complete:
- List exact fields changed
- Show before/after values
- Count total changes
- Confirm no data loss

### 3.4 Update Mode

**When**: User runs `/moai:0-project update`

**DIRECTIVE: Preserve language from config**

The project-manager agent MUST:
1. Read current language from config.json
2. Preserve it throughout update
3. Use it for all announcements
4. DO NOT ask user for language again

**DIRECTIVE: Smart template merging**

- Analyze backup of current config
- Compare with new package templates
- Identify changes (additions, removals, modifications)
- Merge intelligently (keep user changes, add new defaults)
- Run validation on merged config

**DIRECTIVE: Auto-translate announcements**

- Read user's language from config
- Translate update announcements to that language
- Include "What's new" summary
- Include "Configuration changed" details

**DIRECTIVE: Show update summary**

- List new features added
- List deprecated settings removed
- List changed defaults
- Ask user to review

**DIRECTIVE: Save update context**

- Mark phase as UPDATE completed
- Record update timestamp
- List files modified
- Suggest next action (/moai:1-plan)

---

## Section 4: User Interaction Model (Phase 3)

### 4.1 AskUserQuestion Directive

**DIRECTIVE: Every user decision goes through AskUserQuestion**

- NEVER use echo, console output for decisions
- ALWAYS use AskUserQuestion for choices
- ALWAYS use AskUserQuestion for confirmations
- ALWAYS use AskUserQuestion for follow-ups

**DIRECTIVE: Structure questions clearly**

Each question MUST have:
- `question`: What the user is deciding
- `header`: Short label (max 12 chars)
- `options`: 2-4 mutually exclusive choices
- Clear descriptions of what each option means

**DIRECTIVE: No emojis in interactive fields**

- ❌ Emojis in `question` field
- ❌ Emojis in `label` field
- ❌ Emojis in `description` field
- ✅ Emojis only in natural conversation text

**DIRECTIVE: Default values should be obvious**

- Show current value in question text
- Highlight "Keep current" as default option
- Make clear what changes and what stays same

**DIRECTIVE: Language consistency**

- All fields must be in user's conversation language
- No mixing languages in single question
- Respect user's language preference absolutely

### 4.2 Error Recovery Directive

**DIRECTIVE: Validation failures are non-blocking opportunities**

When validation fails:
1. Clearly explain what went wrong
2. Show which field caused issue
3. Suggest how to fix it
4. Offer: "Try again" or "Skip" or "Cancel"
5. DO NOT exit unless user explicitly cancels

**DIRECTIVE: Configuration errors have standard recovery**

| Error Type | Recovery |
|---|---|
| Invalid language code | Re-ask language selection |
| Git conflict | Highlight conflict, offer fix suggestions |
| Missing required field | Show which field, re-ask |
| Invalid data type | Show expected format, re-ask |
| Backup failure | Show error, offer: retry/skip/cancel |

**DIRECTIVE: Context save failures are non-critical**

- If context save fails, complete command successfully
- Show clear warning about context save
- Offer manual retry option
- Log error for debugging

### 4.3 Completion & Next Steps Directive

**DIRECTIVE: Every execution must end with next action**

After any mode completes:
1. Show completion status (success/partial/failure)
2. List what was accomplished
3. Present next step options
4. Use AskUserQuestion for navigation
5. Include clear descriptions

**Completion patterns**:

**After INITIALIZATION**:
```
Project setup is complete. What would you like to do next?

1. Write Specification (Execute /moai:1-plan)
2. Review Project Structure (Show generated files)
3. Modify Settings (Execute /moai:0-project setting)
4. Start New Session (Exit and prepare for next step)
```

**After AUTO-DETECT**:
```
Configuration review complete. Next steps:

1. Continue Settings (Modify another tab)
2. Sync Documentation (Execute /moai:3-sync)
3. View Full Configuration (Display config.json)
4. Exit (No further action)
```

**After SETTINGS**:
```
Settings updated successfully. Next steps:

1. Modify Another Tab (Return to tab selection)
2. Review Changes (Show detailed change log)
3. Sync Documentation (Execute /moai:3-sync)
4. Exit (Finish for now)
```

**After UPDATE**:
```
Package update complete. Next steps:

1. Review Changes (Show detailed diff)
2. Modify Settings (Execute /moai:0-project setting)
3. Start Implementation (Execute /moai:1-plan)
4. Exit (Done for now)
```

---

## Section 5: Configuration Philosophy Directives

### 5.1 Single Source of Truth

**DIRECTIVE: `.moai/config/config.json` is always authoritative**

- This file is the ONLY source of truth
- Never override with in-memory defaults
- Always read from file at start
- Always write changes back to file atomically
- Users who edit file manually should have their edits respected

**DIRECTIVE: Language settings have cascading priority**

```
1. User's conversation_language (set in Tab 1)
   ↓ (derives)
2. conversation_language_name (display name, auto-updated)
   ↓ (same as)
3. default_language (auto-set)
   ↓ (same as)
4. optimized_for_language (auto-set)
```

If user changes conversation_language in Tab 1:
- ALL 4 fields update atomically
- Other tabs don't need to re-ask these questions
- Validation confirms consistency

### 5.2 Configuration Scope

**DIRECTIVE: Clear boundaries of responsibility**

**User settings** (in config):
- User name
- Conversation language
- Agent prompt language
- Project info (name, description, owner, mode)
- Git strategy and branch names
- Quality principles (TRUST 5)
- GitHub automation preferences

**Auto-generated settings** (NOT asked, auto-processed):
- conversation_language_name (from language code)
- project.locale (from conversation_language)
- default_language (from conversation_language)
- optimized_for_language (from conversation_language)

**System settings** (managed by package):
- MoAI version
- Package version
- Template version
- Schema version

### 5.3 Validation Philosophy

**DIRECTIVE: Validation happens at decision points, not constantly**

Validation checkpoints:
1. After Tab 1: Language consistency check
2. After Tab 3: Git strategy conflict check
3. Before config write: Required field check
4. After config write: Integrity check

**DIRECTIVE: Validation errors are teaching moments**

When validation fails:
- Explain WHY it failed in user's language
- Show specific conflicting values
- Suggest what would fix it
- Don't force user to guess

---

## Section 6: Phase-Based Guidelines

### 6.1 Phase 1: Smart Entry Point (5 minutes)

**Goal**: Understand what user wants and prepare execution context

**Key Activities**:
- Analyze command arguments
- Detect initialization state
- Load language context
- Route to correct mode
- Delegate to project-manager agent

**Success Criteria**:
- User intent clearly understood
- Correct mode identified
- Language context loaded
- Agent receives complete context
- Ready for Phase 2 execution

**Error Recovery**:
- Invalid config detected → Offer restore or reinit
- Language missing → Use CLI default
- Arguments invalid → Show clear error and exit

### 6.2 Phase 2: Execute Mode (10-20 minutes)

**Goal**: Execute the selected mode completely

**Key Activities**:
- Mode-specific workflow execution
- User interaction via AskUserQuestion
- Validation at checkpoints
- Delegation to skills for file operations
- Context state management

**Success Criteria**:
- Configuration complete and validated
- All user responses processed
- Changes persisted atomically
- Context saved for next phase
- User aware of what changed

**Error Recovery**:
- Validation failure → Explain, re-ask affected questions
- Skill execution failure → Clear error, offer retry
- Context save failure → Warning (non-blocking)

### 6.3 Phase 3: Completion & Next Steps (2-3 minutes)

**Goal**: Confirm successful execution and guide next action

**Key Activities**:
- Display completion status
- List accomplished changes
- Show available next steps
- User selects what to do
- Prepare context for next command

**Success Criteria**:
- User understands what happened
- User knows what changed
- User has clear path forward
- Session context saved
- Execution logged

**Error Recovery**:
- User cancels next step → Graceful exit
- Context save fails → Warning, offer manual context

---

## Section 7: Critical Rules (Non-Negotiable)

### 7.1 Execution Rules

**RULE 1: Execute exactly ONE mode per invocation**
- Mode detected, executed, completed
- Multi-tab workflows are ONE mode execution
- No mode mixing or chaining within single command

**RULE 2: Never skip language confirmation**
- First action in INITIALIZATION: confirm language
- First action in AUTO-DETECT: read language from config
- First action in SETTINGS: read language from config
- All output must be in that language

**RULE 3: Delegate ALL work to agents and skills**
- Command orchestrates only
- Agents execute
- Skills do file operations
- NO direct Read/Write/Edit/Bash in command

**RULE 4: Use AskUserQuestion for ALL user interaction**
- Every choice → AskUserQuestion
- Every confirmation → AskUserQuestion
- Every next step → AskUserQuestion
- No other output mechanism for decisions

**RULE 5: Atomic config updates only**
- All changes together or none
- Backup created before write
- Rollback available on failure
- User sees final state, not intermediate states

### 7.2 Language Rules

**RULE 1: Config language is authoritative**
- Read language from `.moai/config/config.json`
- Use it for ALL user-facing output
- Pass it to agents as context
- Don't change without explicit user request

**RULE 2: All output in conversation language**
- Questions in user's language
- Options in user's language
- Success messages in user's language
- Error messages in user's language

**RULE 3: No language ambiguity**
- One conversation language per session
- One agent prompt language per session
- Don't switch languages mid-interaction
- Confirm language before decisions

### 7.3 Tool Usage Rules

**RULE 1: ONLY Task() and AskUserQuestion()**
- ❌ NO Read tool usage
- ❌ NO Write tool usage
- ❌ NO Edit tool usage
- ❌ NO Bash tool usage
- ✅ Task() for delegation
- ✅ AskUserQuestion() for interaction

**RULE 2: Task tool usage pattern**
```python
Task(
    subagent_type="project-manager",
    prompt="Complete mode execution",
    [optional: context={...}]
)
```

**RULE 3: AskUserQuestion usage pattern**
```python
AskUserQuestion(
    questions=[{
        question: "...",
        header: "...",
        multiSelect: false,
        options: [...]
    }]
)
```

### 7.4 Configuration Priority Rules

**RULE 1: Existing config has priority**
- If config exists, read it
- Use existing language unless user changes it
- Use existing settings unless user changes them
- Only ask for NEW settings, not re-confirm old ones

**RULE 2: User input overrides everything**
- If user explicitly changes setting, that's final
- No "are you sure?" for valid changes
- Clear confirmation for breaking changes only

**RULE 3: Auto-processing is silent**
- conversation_language_name, locale, etc auto-update
- Don't ask user for these
- Don't announce auto-updates
- Include in change report at end

---

## Section 8: Documentation Structure

### 8.1 Output Documentation

**Project documentation MUST include**:
- `product.md` - What is this project
- `structure.md` - How is it organized
- `tech.md` - What technology does it use

**Documentation MUST be**:
- In user's conversation language
- Generated automatically
- Stored in `.moai/project/` directory
- Referenced in config for quick access

### 8.2 User Guidance Documentation

**Users should have access to**:
- CLAUDE.md - Quick reference (5/15/30 min levels)
- .moai/memory/ - Extended memory files
- .moai/directives/ - Directive guides (like this one)
- Skill documentation - For specific features
- Agent documentation - For execution patterns

**Guidance should be**:
- Progressive (essential → recommended → advanced)
- Practical (show what to do, not theory)
- Example-driven (concrete scenarios)
- Language-aware (in user's language where relevant)

---

## Section 9: Error Handling Directives

### 9.1 Configuration File Errors

**Error: config.json missing**
- Mode: INITIALIZATION
- Recovery: Create new config from interview
- User sees: "Setting up new project"
- Rollback: Not applicable (new file)

**Error: config.json invalid JSON**
- Mode: AUTO-DETECT
- Recovery: Try to restore backup, or reinitialize
- User sees: Clear explanation of what's wrong
- Rollback: Restore from backup if available

**Error: config.json incomplete**
- Mode: AUTO-DETECT
- Recovery: Highlight missing fields, offer settings fix
- User sees: "Some settings are incomplete"
- Rollback: Not applicable (just missing data)

### 9.2 Skill Execution Errors

**Error: Skill not found**
- Recovery: Check skill installation, clear error message
- User sees: "Skill unavailable, try again later"
- Action: Offer to retry or skip this operation

**Error: Skill execution timeout**
- Recovery: Timeout after 60s, offer retry
- User sees: "Operation took too long"
- Action: Retry with increased timeout or skip

**Error: Skill validation failure**
- Recovery: Get detailed error from skill, explain to user
- User sees: Clear explanation in their language
- Action: Offer to fix the issue or cancel

### 9.3 Validation Errors

**Error: Invalid language code**
- When: Tab 1 language selection
- Recovery: Re-ask with clear error message
- User sees: "Language code 'xx' is not recognized"
- Options: Select again or cancel

**Error: Git conflict (Personal mode + develop base)**
- When: Tab 3 completion
- Recovery: Highlight conflict, suggest fixes
- User sees: "These settings conflict. Here's why..."
- Options: Auto-fix suggestion or manual override

**Error: Required field missing**
- When: Before final config write
- Recovery: Show which field, re-ask that field
- User sees: "Project name is required"
- Options: Provide value or cancel

### 9.4 Backup & Recovery Errors

**Error: Backup creation failed**
- Severity: Non-blocking warning
- Recovery: Log error, continue without backup
- User sees: Warning message
- Action: User can manually backup config before continuing

**Error: Context save failed**
- Severity: Non-blocking warning
- Recovery: Log error, complete command successfully
- User sees: "Context save failed, you can retry manually"
- Action: User can run /moai:0-project again to re-save

---

## Section 10: Implementation Checklist

### 10.1 Pre-Implementation Verification

- [ ] All phases understand mode detection logic
- [ ] All phases use only Task() and AskUserQuestion()
- [ ] Language handling is consistent across all modes
- [ ] Error messages translate to user's language
- [ ] Skill dependencies are documented
- [ ] Validation checkpoints are in correct places
- [ ] No emojis in AskUserQuestion fields
- [ ] Tab schema format is validated
- [ ] Backup/rollback logic is tested
- [ ] Context save mechanism is specified

### 10.2 Agent Implementation Checklist

**project-manager agent MUST**:
- [ ] Detect and route to correct mode
- [ ] Read language from config before starting
- [ ] Use language for all output
- [ ] Delegate to correct skills
- [ ] Handle skill errors gracefully
- [ ] Validate at checkpoints
- [ ] Save context atomically
- [ ] Report completion clearly
- [ ] Guide next steps
- [ ] Handle cancellation gracefully

### 10.3 Skill Integration Checklist

**moai-project-language-initializer**:
- [ ] Offer language selection
- [ ] Validate language codes
- [ ] Return selected language

**moai-project-config-manager**:
- [ ] Read current config
- [ ] Create backup before write
- [ ] Merge updates deeply
- [ ] Validate final config
- [ ] Write atomically
- [ ] Implement rollback
- [ ] Report changes made

**moai-project-batch-questions**:
- [ ] Load tab schema
- [ ] Generate questions in user's language
- [ ] Load current values
- [ ] Map responses to config paths

**moai-project-template-optimizer**:
- [ ] Load existing config
- [ ] Load new templates
- [ ] Detect changes intelligently
- [ ] Preserve user changes
- [ ] Validate merged config

### 10.4 Testing Checklist

- [ ] Test INITIALIZATION mode (new project)
- [ ] Test AUTO-DETECT mode (existing project)
- [ ] Test SETTINGS mode (all tabs)
- [ ] Test UPDATE mode (after package update)
- [ ] Test language switching
- [ ] Test validation checkpoints
- [ ] Test error recovery paths
- [ ] Test context save
- [ ] Test multi-tab workflow
- [ ] Test cancellation at various points
- [ ] Test with different languages
- [ ] Test config file corruption recovery

---

## Section 11: Success Metrics

### 11.1 User Experience Success

**Metric: User completes setup without confusion**
- Success: User can answer 4-6 questions and project is ready
- Measure: No support requests about command purpose
- Target: 90% of users understand what command does

**Metric: Configuration aligns with user intent**
- Success: Resulting config matches what user wants
- Measure: No manual edits needed after command
- Target: 95% of configs correct on first try

**Metric: User knows what changed**
- Success: User can list all changes made
- Measure: Change report is clear and complete
- Target: 100% of changes reported

**Metric: User knows next step**
- Success: User selects next action from clear options
- Measure: User doesn't ask "what do I do now?"
- Target: 95% of users select a next step

### 11.2 System Success

**Metric: Command completes successfully**
- Success: Config written, context saved, exit clean
- Measure: No errors or warnings
- Target: 99% success rate

**Metric: Language handling is correct**
- Success: All output in user's language
- Measure: No language mixing or omissions
- Target: 100% language consistency

**Metric: Validation catches errors**
- Success: Invalid configs caught before write
- Measure: All errors caught at checkpoints
- Target: 100% validation coverage

**Metric: Backup/rollback works**
- Success: Failures are recoverable
- Measure: Rollback restores previous state
- Target: 100% recovery success

---

## Section 12: Quick Reference for Implementers

### 12.1 Mode Decision Tree

```
User runs: /moai:0-project [args]
    |
    +-- No args?
    |   +-- Config exists? → AUTO-DETECT
    |   +-- No config? → INITIALIZATION
    |
    +-- "setting"?
    |   +-- Tab ID specified? → SETTINGS (that tab)
    |   +-- No tab ID? → Tab selection → SETTINGS
    |
    +-- "update"? → UPDATE
    |
    +-- Invalid? → Error message, exit
```

### 12.2 Phase Checklist (Implementer View)

**Phase 1: Entry Point**
1. Parse command arguments
2. Detect initialization state
3. Load language context
4. Route to mode
5. Call project-manager agent

**Phase 2: Mode Execution**
1. project-manager receives mode
2. Executes mode-specific workflow
3. Delegates to skills
4. Validates at checkpoints
5. Saves context

**Phase 3: Completion**
1. Display status
2. List changes
3. Present next steps
4. User selects action
5. Exit cleanly

### 12.3 Critical Integration Points

**Integration 1: project-manager agent**
- Receives: Mode, language context, user request
- Returns: Mode result, changes made, next steps
- Responsible for: All mode logic, skill delegation, error handling

**Integration 2: moai-project-config-manager skill**
- Receives: Config updates, validation rules
- Returns: Success/failure, detailed changes
- Responsible for: File operations, backup/rollback, atomic writes

**Integration 3: AskUserQuestion tool**
- Receives: Question structure, current values
- Returns: User response
- Responsible for: UX, language rendering, option presentation

**Integration 4: Skill("moai-project-batch-questions")**
- Receives: Tab ID, language, current config
- Returns: User responses mapped to config fields
- Responsible for: Tab schema, question generation, response processing

---

## Section 13: Language-Specific Directives

### 13.1 Message Formatting by Language

**When translating user-facing strings**:
- Korean: Respect formal/informal tone conventions
- English: Use clear, simple vocabulary
- Spanish: Account for gender and number agreement
- Japanese: Use appropriate politeness level

**Structure for messages**:
```
Problem: State clearly what went wrong
Impact: Why this matters to the user
Solution: How to fix it
Action: What user should do next
```

### 13.2 Auto-Processing by Language

**When language changes**:
1. Update `conversation_language` immediately
2. Auto-update `conversation_language_name` (display name)
3. Auto-update `project.locale` (if present)
4. Auto-update `default_language` (if present)
5. Auto-update `optimized_for_language` (if present)
6. Translate any announcement text to new language
7. Do NOT re-ask other questions

**When language is confirmed**:
1. Use this language for ALL subsequent output
2. Pass language to all agents
3. Load skills in this language
4. Generate documentation in this language

---

## Appendix A: Tab Schema Reference

**Location**: `.claude/skills/moai-project-batch-questions/tab_schema.json`

**Version**: 1.0.0 (as of 2025-11-19)

**Tabs**:
1. Tab 1: User & Language (REQUIRED)
   - Questions: 3
   - Checkpoint: Language validation
2. Tab 2: Project Basic Information (RECOMMENDED)
   - Questions: 4 (+ auto-processing)
   - Checkpoint: None
3. Tab 3: Git Strategy & Workflow (RECOMMENDED)
   - Questions: 16
   - Checkpoint: Git conflict detection
4. Tab 4: Quality Principles & Reports (OPTIONAL)
   - Questions: 9
   - Checkpoint: None
5. Tab 5: System & GitHub Integration (OPTIONAL)
   - Questions: 7
   - Checkpoint: None

---

## Appendix B: Associated Skills

| Skill | Purpose | Responsibility |
|---|---|---|
| moai-project-language-initializer | Language selection | Offer options, validate, return choice |
| moai-project-config-manager | Config operations | Read, write, backup, rollback, validate |
| moai-project-batch-questions | Tab-based questions | Load schema, generate questions, process responses |
| moai-project-template-optimizer | Template merging | Analyze changes, merge intelligently, validate |
| moai-project-documentation | Doc generation | Create product/structure/tech docs, auto-translate |

---

## Appendix C: Configuration File Structure

**Location**: `.moai/config/config.json`

**Key sections**:
- `moai`: Version information
- `user`: User name, email
- `language`: Conversation language, prompt language
- `project`: Name, description, owner, mode, locale
- `git`: Personal/Team strategy, branch names
- `quality`: TRUST 5 settings
- `system`: GitHub, MoAI settings

**Auto-generated fields** (DO NOT ask user):
- `project.locale`: Derived from language
- `language.conversation_language_name`: Display name
- `language.default_language`: Same as conversation language
- `language.optimized_for_language`: Same as conversation language

---

## Final Summary

**The /moai:0-project command is successful when**:

1. User understands what just happened
2. User knows what changed (or didn't)
3. User has a clear next step
4. Configuration is validated and persisted
5. Context is saved for next command
6. Language is consistent throughout
7. Error recovery paths work
8. User can get back to previous state if needed

**This command represents the first interaction** in the MoAI-ADK workflow. Everything that follows depends on correct initialization. Therefore, this command must be:

- **Reliable**: Always succeeds or recovers gracefully
- **Clear**: User never confused about what's happening
- **Respectful**: Honors user's language and preferences
- **Guiding**: Shows what to do next
- **Recoverable**: Mistakes are not permanent

---

**Document Version**: 1.0.0
**Last Updated**: 2025-11-19
**Status**: Ready for implementation
**Audience**: Project-manager agent, skills developers, users
**Maintenance**: Update when mode logic changes or new validation added
