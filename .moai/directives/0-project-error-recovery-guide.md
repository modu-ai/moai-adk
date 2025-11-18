---
title: "/moai:0-project Error Recovery Guide"
version: "1.0.0"
updated: "2025-11-19"
audience: "Implementers, support teams, error handlers"
scope: "Complete error classification, recovery strategies, and prevention"
---

# /moai:0-project Error Recovery Guide

**Comprehensive guide to handling, recovering from, and preventing errors in the /moai:0-project command.**

---

## Error Classification System

Errors are classified by **severity**, **recoverability**, and **user impact**.

### Severity Levels

**CRITICAL**: Command cannot complete, user data at risk
- Examples: Config file corruption, backup system failure
- Recovery: Immediate intervention required
- User impact: High

**HIGH**: Command completion blocked, but data safe
- Examples: Invalid user input, skill execution failure
- Recovery: User action required (retry, fix input)
- User impact: Medium-High

**MEDIUM**: Feature unavailable, but command continues
- Examples: Context save failure, optional field invalid
- Recovery: Non-blocking warning, retry available
- User impact: Medium

**LOW**: Minor issue, transparent recovery
- Examples: Skill slow, non-critical validation warning
- Recovery: Automatic retry or silent skip
- User impact: Low

---

## Error Categories & Handling

### Category 1: Configuration File Errors

#### 1.1 Config File Missing

**Error ID**: CONFIG_MISSING

**Severity**: HIGH (triggers INITIALIZATION mode)

**When It Occurs**:
- First time running on new machine
- Config file deleted
- Wrong directory

**Error Message to User**:
```
[User's language]
"No project configuration found.
Let's set up your project from scratch.

This will take 5-10 minutes."
```

**Recovery Steps**:
1. Detect missing config
2. Automatically enter INITIALIZATION mode
3. Ask language first
4. Conduct user interview
5. Create new config file
6. Return to normal flow

**User Action Required**: Answer 4-6 setup questions

**Prevention**:
- Keep config in `.moai/config/config.json` (standard location)
- Don't manually delete config
- Use `/moai:0-project setting` to modify, not manual edits

---

#### 1.2 Config File Invalid JSON

**Error ID**: CONFIG_INVALID_JSON

**Severity**: CRITICAL

**When It Occurs**:
- User manually edited config with syntax error
- File corruption during write
- Incomplete write interrupted

**Error Message to User**:
```
[User's language]
"Configuration file is corrupted and cannot be read.

We have two options:
1. Restore from backup (if available)
2. Start fresh with new configuration

Which would you prefer?"
```

**Detection**:
```
try:
    config = json.load(config_file)
except json.JSONDecodeError as e:
    → CONFIG_INVALID_JSON
    → Show line number and context
```

**Recovery Steps**:
1. Detect invalid JSON
2. Check for backup (`.moai/config/config.json.backup`)
3. If backup exists:
   - Show user: "Backup available. Restore?"
   - If yes: Restore backup, validate
   - If no: Proceed to fresh init
4. If no backup:
   - Show user: "No backup available. Start fresh?"
   - If yes: Reinitialize
   - If no: Exit without changes

**User Action Required**: Choose recovery option

**Backup Strategy**:
- Before every write: `cp config.json config.json.backup`
- Keep 3 versions: current, -1, -2
- Backup located next to original file

---

#### 1.3 Config File Incomplete

**Error ID**: CONFIG_INCOMPLETE

**Severity**: MEDIUM

**When It Occurs**:
- Required fields missing
- User manually deleted a section
- Config from old version missing new fields

**Error Message to User**:
```
[User's language]
"Some configuration settings are missing:
- user.name
- project.name

Would you like to fill these in now?"
```

**Detection**:
```
required_fields = [
    "moai.version",
    "user.name",
    "language.conversation_language",
    "project.name"
]

missing = [f for f in required_fields if f not in config]
if missing:
    → CONFIG_INCOMPLETE
    → List missing fields
```

**Recovery Steps**:
1. Detect missing required fields
2. List which fields are missing
3. Offer: "Fill in missing fields now?"
4. If yes:
   - Open SETTINGS mode for relevant tabs
   - Ask user for missing values
   - Merge with existing config
   - Write atomically
5. If no:
   - Exit with warning

**User Action Required**: Provide missing values or confirm skip

---

#### 1.4 Config File Permission Denied

**Error ID**: CONFIG_PERMISSION_DENIED

**Severity**: CRITICAL

**When It Occurs**:
- File permissions incorrect (read-only when should be writable)
- Running from sandboxed/restricted environment
- Network drive access issue

**Error Message to User**:
```
[User's language]
"Cannot access configuration file.
Permission denied.

This might be because:
- File is read-only
- You don't have write permission
- Running in restricted environment

Contact your administrator for help."
```

**Recovery Steps**:
1. Attempt to read config
2. If read succeeds but write fails:
   - Show: "Config can be read but not written"
   - Show: "In read-only mode, cannot save changes"
   - Offer: "Show current config?" or "Exit"
3. If read also fails:
   - Show: "Cannot access configuration file"
   - Exit cleanly

**User Action Required**: Fix permissions or contact admin

**Prevention**:
- Ensure `.moai/` directory is writable by user
- Don't run as different user than created config
- Don't store config on read-only network share

---

### Category 2: User Input Validation Errors

#### 2.1 Invalid Language Code

**Error ID**: INVALID_LANGUAGE_CODE

**Severity**: HIGH

**When It Occurs**:
- User selects language option that doesn't exist
- Manual edit of config with invalid language code
- Auto-suggestion engine provides invalid code

**Error Message to User**:
```
[Current language or English]
"Language code '[code]' is not recognized.

Supported languages: Korean (ko), English (en),
Spanish (es), Japanese (ja)

Please select a valid language."
```

**Detection**:
```
supported_languages = ["ko", "en", "es", "ja"]
if selected_language not in supported_languages:
    → INVALID_LANGUAGE_CODE
    → Show available options
```

**Recovery Steps**:
1. Detect invalid language
2. Show supported languages
3. Ask user to reselect
4. Validate new selection
5. If valid: Update config
6. If invalid again: Repeat until valid or user cancels

**Checkpoint**: Tab 1 validation (after language selection)

**User Action Required**: Reselect from valid options

**Prevention**:
- Use AskUserQuestion with predefined options (no free text)
- Don't allow manual language code entry
- Validate on every config read (catch manual edits)

---

#### 2.2 Missing Required Field

**Error ID**: MISSING_REQUIRED_FIELD

**Severity**: HIGH

**When It Occurs**:
- User cancels mid-interview without answering required question
- User leaves field empty that requires value
- Tab 2 skipped but project.name not in config

**Error Message to User**:
```
[User's language]
"Required field is missing: [field name]

This is needed to complete the setup.

Would you like to provide it now?"
```

**Checkpoint**: Before final config write

**Detection**:
```
required = ["user.name", "project.name",
            "language.conversation_language"]
missing = [f for f in required if f not in updates
           and f not in existing_config]
if missing:
    → MISSING_REQUIRED_FIELD
    → List missing fields
```

**Recovery Steps**:
1. Identify missing required fields
2. Show which field is missing and why it's required
3. Offer: "Would you like to provide this now?"
4. If yes:
   - Ask specifically for that field (via AskUserQuestion)
   - Validate response
   - Update config
5. If no:
   - Block config write
   - Offer: "Cancel this operation?"

**User Action Required**: Provide missing field or cancel

**Prevention**:
- Always validate before allowing skip
- Make required fields obvious in questions
- Don't allow next step until required field provided

---

#### 2.3 Git Configuration Conflict

**Error ID**: GIT_CONFLICT

**Severity**: HIGH

**When It Occurs**:
- User selects Personal mode but sets base_branch to "develop"
- User sets PR strategy for team but team mode disabled
- Branch naming conflicts (e.g., main/master inconsistency)

**Error Message to User**:
```
[User's language]
"Configuration conflict detected:

You selected 'Personal' mode but set PR base to 'develop'.
In Personal mode, changes should go directly to 'main'.

Which would you prefer?
1. Change to Team mode (for develop-based workflow)
2. Change PR base to main (for Personal mode)
3. Keep as is and accept the mismatch"
```

**Checkpoint**: Tab 3 validation (after git settings)

**Detection**:
```
mode = config["project"]["mode"]
pr_base = config["git"]["personal"]["pr_base"]

if mode == "personal" and pr_base == "develop":
    → GIT_CONFLICT
    → Suggest fixes

conflict_rules = [
    ("personal", "pr_base", "develop", "should be main"),
    ("team", "use_gitflow", False, "should be True for team")
]
```

**Recovery Steps**:
1. Detect conflict
2. Explain what's wrong in user's language
3. Show suggested fixes
4. Offer options:
   - Auto-apply suggestion (recommend)
   - Manually fix (let user choose)
   - Keep current (with warning)
   - Cancel and restart
5. If user selects option: Update config and continue
6. If cancel: Discard changes and return to tab

**User Action Required**: Resolve conflict or cancel tab

**Prevention**:
- Ask mode-aware questions (show/hide based on selected mode)
- Validate as user answers each question
- Show implications when user changes mode

---

### Category 3: Skill Execution Errors

#### 3.1 Skill Not Found

**Error ID**: SKILL_NOT_FOUND

**Severity**: HIGH (non-blocking)

**When It Occurs**:
- moai-project-config-manager not installed
- moai-project-batch-questions missing
- Skill directory corrupted

**Error Message to User**:
```
[User's language]
"A system skill is missing: [skill name]

This is required for configuration management.

Options:
1. Retry (might be temporary)
2. Exit and try again later
3. Show system details (for debugging)"
```

**Recovery Steps**:
1. Attempt to load skill
2. If fails: Show error message
3. Offer retry (in case of transient issue)
4. If retry also fails:
   - Show: "Skill unavailable"
   - Offer: "Try again later or exit?"
5. If user retries and skill appears: Continue
6. If user exits: Save any partial progress, exit cleanly

**User Action Required**: Retry or exit

**Prevention**:
- Verify skill installation before command
- Keep skill files in `.claude/skills/` with correct structure
- Document skill dependencies

---

#### 3.2 Skill Execution Timeout

**Error ID**: SKILL_TIMEOUT

**Severity**: MEDIUM

**When It Occurs**:
- Skill takes longer than 60 seconds
- System is slow or overloaded
- Skill has infinite loop or deadlock

**Error Message to User**:
```
[User's language]
"System operation took too long and timed out.

This might be temporary. Would you like to try again?

1. Retry (same operation)
2. Skip (continue without this step)
3. Cancel (exit completely)"
```

**Timeout Values**:
- Config read: 5 seconds
- Config write: 10 seconds
- Skill execution: 60 seconds
- Question answering: No timeout (user-driven)

**Recovery Steps**:
1. Start operation with timeout
2. If timeout occurs:
   - Cancel operation
   - Show error message
   - Offer retry, skip, or cancel
3. If user retries:
   - Attempt operation again
   - If succeeds: Continue
   - If fails again: Offer skip or cancel
4. If user skips:
   - Skip this operation
   - Continue with next step
5. If user cancels:
   - Discard any partial changes
   - Exit cleanly

**User Action Required**: Retry, skip, or cancel

**Prevention**:
- Monitor skill performance
- Log slow operations
- Increase timeout if skills are consistently slow

---

#### 3.3 Skill Validation Error

**Error ID**: SKILL_VALIDATION_ERROR

**Severity**: HIGH

**When It Occurs**:
- Config manager rejects proposed config as invalid
- Batch questions skill returns malformed response
- Template optimizer fails to merge configs

**Error Message to User**:
```
[User's language]
"Validation failed: [detailed error message from skill]

This usually means:
- Your answers don't match required format
- Configuration has conflicting values
- System encountered unexpected data

Details:
[Show specific validation error]

Would you like to:
1. Review the error details
2. Try again with different answers
3. Cancel this operation"
```

**Recovery Steps**:
1. Skill detects validation error
2. Pass detailed error message to command
3. Show error in user's language
4. Extract actionable advice from error
5. Offer recovery options:
   - Show details (technical info)
   - Retry (go back to relevant questions)
   - Cancel (discard changes)
6. If retry: Return to last unanswered question
7. If cancel: Discard changes, return to main menu

**User Action Required**: Retry or cancel

**Prevention**:
- Validate user input BEFORE passing to skill
- Check data types and ranges
- Provide clear error messages from skills

---

### Category 4: Persistence & Backup Errors

#### 4.1 Config Write Failed

**Error ID**: CONFIG_WRITE_FAILED

**Severity**: CRITICAL

**When It Occurs**:
- Disk is full
- File permissions changed mid-operation
- Filesystem became read-only
- Interrupted write (power loss, crash)

**Error Message to User**:
```
[User's language]
"Could not save configuration to disk.

This might be because:
- Disk is full
- Permission problem
- Network disconnection

Your changes are not saved.

Options:
1. Try again (might be transient)
2. Show system status
3. Exit without saving"
```

**Recovery Steps**:
1. Attempt to write config atomically
2. If write fails:
   - Verify backup still exists (rollback available)
   - Show error message to user
   - Offer retry or exit
3. If retry succeeds:
   - Verify written content (read back and compare)
   - Show: "Configuration saved successfully"
   - Continue
4. If retry fails again:
   - Show: "Cannot write to disk"
   - Offer: "Exit without saving changes?"
5. If user exits:
   - Confirm: "Changes will be lost. Are you sure?"
   - If confirmed: Restore from backup (if any changes made)
   - Exit cleanly

**Atomic Write Pattern**:
```
1. Create temporary file: config.json.tmp
2. Write to temporary file
3. Read back and verify (checksum)
4. Rename temp to actual: config.json
5. Remove temporary file

If any step fails: Rollback
```

**User Action Required**: Retry or exit without saving

**Prevention**:
- Check disk space before write
- Verify file permissions before write
- Use atomic write pattern
- Keep recent backup always available

---

#### 4.2 Backup Creation Failed

**Error ID**: BACKUP_CREATION_FAILED

**Severity**: MEDIUM (non-critical)

**When It Occurs**:
- Cannot create backup file
- Backup location has no write permission
- Disk space insufficient for backup

**Error Message to User**:
```
[User's language]
"Could not create backup of current configuration.

This is a warning. Configuration update can continue,
but we won't have a backup to restore if something fails.

Would you like to:
1. Continue without backup (risky)
2. Try backup again
3. Cancel update (safest)"
```

**Recovery Steps**:
1. Attempt to create backup
2. If fails:
   - Log error
   - Show non-critical warning
   - Offer continue, retry, or cancel
3. If user chooses continue:
   - Proceed with config update
   - User accepts risk
   - Log this decision
4. If retry:
   - Attempt backup again
   - If succeeds: Proceed normally
   - If fails again: Repeat offer
5. If cancel:
   - Discard changes
   - Exit cleanly

**Backup Strategy**:
```
Current: config.json
Backup 1: config.json.backup-1
Backup 2: config.json.backup-2
Backup 3: config.json.backup-3
```

Keep 3 versions rotating (oldest deleted).

**User Action Required**: Continue, retry, or cancel

**Prevention**:
- Verify backup directory is writable at session start
- Monitor disk space
- Rotate backups to keep space usage low

---

#### 4.3 Rollback Failed

**Error ID**: ROLLBACK_FAILED

**Severity**: CRITICAL

**When It Occurs**:
- Backup file corrupted
- Cannot write backup back to original location
- Backup location no longer accessible
- Concurrent modification during rollback

**Error Message to User**:
```
[User's language]
"Automatic recovery failed. Configuration may be corrupted.

An error occurred:
- Could not restore backup
- Current state is unknown
- Manual intervention required

Please contact support with this information:
[Timestamp, error details]

We recommend:
1. Do NOT modify configuration manually
2. Contact support immediately
3. Have your backup files available for recovery"
```

**Recovery Steps**:
1. Detect write failure (or any critical failure)
2. Attempt automatic rollback
3. If rollback succeeds:
   - Verify restored state
   - Show: "Rolled back to previous configuration"
   - Continue or exit
4. If rollback fails:
   - This is CRITICAL
   - Show detailed error
   - Recommend user contact support
   - Log all details for debugging
   - Exit without further changes

**User Action Required**: Contact support

**Prevention**:
- Test rollback mechanism regularly
- Keep multiple backup copies
- Verify backups are readable after creation
- Monitor backup file integrity

---

### Category 5: Context & Session Errors

#### 5.1 Context Save Failed

**Error ID**: CONTEXT_SAVE_FAILED

**Severity**: LOW (non-blocking)

**When It Occurs**:
- Cannot write to `.moai/logs/`
- Cannot serialize context to JSON
- Session context directory missing

**Error Message to User**:
```
[User's language]
"Context could not be saved for next command.

This is a non-critical warning. Your configuration
is saved, but context for the next step may be lost.

You can run /moai:0-project again to restore context."
```

**Recovery Steps**:
1. After command completes, attempt context save
2. If fails:
   - Log error (but don't crash)
   - Show warning to user
   - Offer: "Retry context save?"
3. If retry succeeds:
   - Continue normally
4. If retry fails:
   - Show: "Context save failed. You can retry manually."
   - Complete command successfully anyway

**Context Data**:
```
{
    "phase": "0-project",
    "mode": "INITIALIZATION",
    "timestamp": "2025-11-19T10:30:00Z",
    "project_name": "user-auth-system",
    "language": "ko",
    "status": "completed"
}
```

**User Action Required**: None (non-blocking)

**Prevention**:
- Ensure `.moai/logs/` exists and is writable
- Handle JSON serialization errors gracefully
- Implement fallback logging mechanism

---

#### 5.2 Session Interrupted

**Error ID**: SESSION_INTERRUPTED

**Severity**: HIGH

**When It Occurs**:
- User closes terminal/Claude Code mid-command
- Network disconnection during operation
- System shutdown or crash
- User sends Ctrl+C

**Error Message to User**:
```
[When reconnecting]

"Your previous session was interrupted.

Your configuration:
- Last saved state: [timestamp]
- Mode: [mode name]
- Progress: [% complete]

Would you like to:
1. Resume from where you left off
2. Start fresh
3. Show what was saved"
```

**Recovery Steps**:
1. Detect session interruption (via saved context)
2. Show user where they were
3. Offer resume, restart, or review
4. If resume:
   - Load saved context
   - Jump to next unanswered question
   - Continue from there
5. If restart:
   - Start from beginning
   - Show only unsaved changes would be lost
6. If review:
   - Show saved configuration status

**Session Persistence**:
- Save context every 2 questions
- Save config after every change
- Keep session state in `.moai/sessions/`

**User Action Required**: Resume or restart

**Prevention**:
- Save progress frequently
- Make resumption explicit (show resume option)
- Keep detailed session logs

---

### Category 6: Agent & Task Execution Errors

#### 6.1 Project Manager Agent Timeout

**Error ID**: AGENT_TIMEOUT

**Severity**: HIGH

**When It Occurs**:
- project-manager agent takes > 5 minutes
- Agent blocked on I/O operation
- Agent infinite loop

**Error Message to User**:
```
[User's language]
"System operation took too long.

The command started 5 minutes ago and hasn't completed.
This might be a system issue.

Options:
1. Wait a bit longer (may still complete)
2. Try again from the beginning
3. Exit and try later"
```

**Timeout Thresholds**:
- INITIALIZATION mode: 5 minutes
- AUTO-DETECT mode: 2 minutes
- SETTINGS mode: 3 minutes
- UPDATE mode: 5 minutes

**Recovery Steps**:
1. Monitor agent execution time
2. If timeout threshold exceeded:
   - Show warning to user
   - Offer wait, retry, or exit
3. If user chooses wait:
   - Wait additional 2 minutes
   - If completes: Continue normally
   - If fails: Offer retry or exit
4. If user retries:
   - Start mode execution again
   - Monitor same thresholds
5. If user exits:
   - Save any partial context
   - Exit cleanly

**User Action Required**: Wait, retry, or exit

**Prevention**:
- Monitor agent execution in agent code
- Log checkpoints during long operations
- Optimize slow skill calls

---

#### 6.2 Agent Execution Error

**Error ID**: AGENT_ERROR

**Severity**: HIGH

**When It Occurs**:
- Agent encounters unexpected exception
- Agent skill delegation fails
- Agent subprocess crashes

**Error Message to User**:
```
[User's language]
"An internal system error occurred.

This shouldn't happen. Details logged for debugging.

Error type: [error class]
Timestamp: [ISO timestamp]

Would you like to:
1. Try the operation again
2. Exit and try later
3. View error details (technical)"
```

**Recovery Steps**:
1. Catch exception in agent
2. Extract error details
3. Show user-friendly message
4. Log full error stack trace
5. Offer retry or exit
6. If retry: Restart mode execution
7. If exit: Save context, exit cleanly

**Logging**:
```
Log location: .moai/logs/agent-errors-{date}.log
Include: timestamp, error type, stack trace,
         user language, mode, phase
```

**User Action Required**: Retry or exit

**Prevention**:
- Comprehensive error handling in agents
- Detailed logging of failures
- Regular testing of error paths

---

## Error Prevention Strategies

### Strategy 1: Validation at Entry Points

**When**: Before accepting any user input

**What to Check**:
- Language code validity
- Required field presence
- Data type correctness
- Range/length validation

**Example**:
```
Before asking "Git strategy":
✓ User language is valid
✓ Project name is provided
✓ Git mode is selected (personal/team)

Then ask Git strategy questions (mode-specific)
```

### Strategy 2: Checkpoint Validation

**Checkpoints**:
- Tab 1 complete: Language consistency
- Tab 3 complete: Git conflict detection
- Before config write: All required fields

**Pattern**:
```
1. Collect user answers
2. Run validation rules
3. If error: Explain, re-ask affected fields
4. If pass: Continue to next tab or write
```

### Strategy 3: Atomic Operations

**Pattern**:
```
1. Backup current state
2. Attempt operation
3. If error: Rollback to backup
4. If success: Remove backup (after verification)
```

**Never leave partial state**:
- All changes together or none
- User sees final state only

### Strategy 4: Graceful Degradation

**Principle**: Less critical features fail silently

| Feature | Severity | Behavior |
|---|---|---|
| Config write | CRITICAL | Block, error, retry |
| Backup creation | MEDIUM | Warning, continue |
| Context save | LOW | Warning, continue |
| Documentation gen | LOW | Skip, warn, continue |

### Strategy 5: User Confirmation

**When to require confirmation**:
- Breaking changes (mode switch)
- Data loss (reinitialize)
- Risky operations (no backup)

**Pattern**:
```
"This action will [impact].
Are you sure? (Type 'yes' to confirm)"
```

---

## Troubleshooting Decision Tree

```
Error occurs
    ↓
Is it a validation error?
├─ YES: Show user what's wrong, let them fix it
└─ NO: Is it a file error?
    ├─ YES: Check config access, offer restore/reinit
    └─ NO: Is it a skill error?
        ├─ YES: Show skill error, offer retry/skip
        └─ NO: Is it a timeout?
            ├─ YES: Offer retry or exit
            └─ NO: Is it an interruption?
                ├─ YES: Offer resume
                └─ NO: Contact support (log details)
```

---

## Error Message Quality Standards

**Every error message MUST**:

1. **Be in user's language**
   - Not English, not technical jargon
   - Translated accurately

2. **Explain what went wrong**
   - "Configuration file is corrupted"
   - Not "json.JSONDecodeError"

3. **Suggest why it happened**
   - "This might be because..."
   - "Common causes are..."

4. **Provide recovery path**
   - "Would you like to...?"
   - Always offer options

5. **Be actionable**
   - Tell user what to do
   - Not just diagnose problem

**Good**: "Project name is required. What would you like to name your project?"

**Bad**: "ValueError: 'project.name' key not found in config dict"

---

## Testing Error Paths

### Test Cases Required

- [ ] Config file missing (first time)
- [ ] Config file corrupted (manual edit)
- [ ] Config file incomplete (missing fields)
- [ ] Invalid language code
- [ ] Missing required field
- [ ] Git conflict (personal + develop)
- [ ] Skill not found (missing moai-project-config-manager)
- [ ] Skill timeout (60+ seconds)
- [ ] Config write fails (disk full simulation)
- [ ] Backup fails
- [ ] Rollback fails
- [ ] Context save fails
- [ ] Agent timeout
- [ ] Agent execution error
- [ ] Session interruption (Ctrl+C)
- [ ] Resume from interrupted session

### Test Environment

- Mock filesystem errors
- Mock skill failures
- Mock timeout conditions
- Mock corrupted config states
- Test recovery from each error state

---

## Support Escalation Path

**For users unable to recover**:

1. **Level 1: Self-help**
   - Read error message carefully
   - Try suggested recovery step
   - Wait and retry

2. **Level 2: Documentation**
   - Check CLAUDE.md
   - Check .moai/memory/ files
   - Review examples

3. **Level 3: Community**
   - Check GitHub issues
   - Ask in community forum
   - Provide error details/logs

4. **Level 4: Direct support**
   - Provide `.moai/logs/` files
   - Provide config backup (sanitized)
   - Provide exact steps to reproduce
   - Provide system information

---

## Logging & Debugging

### Log Files

**Location**: `.moai/logs/`

**Files**:
- `agent-errors-{date}.log` - Agent execution errors
- `config-operations-{date}.log` - Config reads/writes
- `validation-{date}.log` - Validation failures
- `command-trace-{date}.log` - Full command trace

**Log Level**:
- ERROR: Fatal conditions
- WARNING: Recovery attempted
- INFO: Normal operation milestones
- DEBUG: Detailed operation steps

### Enable Debug Mode

```
/moai:0-project --debug
```

**Debug output includes**:
- Config state at each step
- User inputs received
- Skill calls and responses
- Validation results
- File operations

---

## Compliance & Standards

**Error handling MUST comply with**:

- TRUST 5 principle (Trackable)
  - All errors logged with full context
  - Support can trace user journey

- Security standards
  - Never expose system paths in user messages
  - Never expose credentials
  - Sanitize error messages for user language

- Accessibility standards
  - Error messages readable
  - Recovery options clear
  - No critical info in colors only

- User privacy
  - Don't collect personal data in error logs
  - Allow user to opt out of detailed logging

---

**Document Version**: 1.0.0
**Last Updated**: 2025-11-19
**Status**: Ready for implementation
**Audience**: Implementers, support teams, QA engineers

For implementation, reference: `/moai/directives/0-project-command-directive.md`
