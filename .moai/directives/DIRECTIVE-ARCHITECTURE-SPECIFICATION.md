---
title: "Directive Architecture Specification for /moai:0-project"
version: "2.0.0"
date: "2025-11-19"
audience: "Architects, Implementers"
scope: "How directives SHOULD BE EMBEDDED in official Claude Code files"
status: "Complete Specification (Not Code)"
---

# Directive Architecture Specification

**This document specifies WHERE, HOW, and WHAT directives should be embedded in official Claude Code files.**

---

## Problem Statement

**Current (WRONG)**:
- Directives stored in separate `.moai/directives/` directory
- Directives disconnected from actual command/agent implementation
- Multiple sources of truth
- Maintenance burden (directives + files get out of sync)

**Target (CORRECT)**:
- Directives embedded directly IN official files
- `.claude/commands/moai/0-project.md` contains command directives
- `.claude/agents/moai/project-manager.md` contains agent directives
- Supporting skills contain skill-specific directives
- Single source of truth: the official files
- Directives are executable specifications

---

## Core Architecture Principle

**Official Claude Code files ARE the directives.**

The `.claude/` directory files are not just implementation - they ARE the specification. When you read `.claude/commands/moai/0-project.md`, you are reading what the command SHOULD DO.

### Consequences

1. **Directives are authoritative**: When command implementation differs from command directive, the directive wins
2. **Directives are living documents**: They evolve with the code
3. **No separate directive documents** in `.moai/directives/` - that's a violation of DRY principle
4. **Single source of truth**: The official `.claude/` files

---

## Architecture Overview

### Layer 1: Command File (.claude/commands/moai/0-project.md)

**Purpose**: User-facing specification and entry point orchestration

**Structure**:
```
YAML Frontmatter (standard)
  ├── name
  ├── description
  ├── argument-hint
  ├── tools (Task, AskUserQuestion only)
  └── model

User Directives (THIS IS NEW)
  ├── What this command does
  ├── When to use it
  ├── What NOT to do
  └── Entry point logic

Execution Directives (THIS IS NEW)
  ├── Mode detection
  ├── Initial validation
  ├── Language loading
  └── Agent delegation pattern

Error Handling (THIS IS NEW)
  ├── Entry point errors
  ├── Recovery patterns
  └── User-facing messages

Tool Constraints (THIS IS NEW)
  ├── Only Task() and AskUserQuestion()
  ├── No Read/Write/Edit/Bash
  └── Why these constraints matter

Content (markdown body)
  ├── Command overview
  ├── Use cases
  ├── Quick start
  └── Skill references
```

### Layer 2: Agent File (.claude/agents/moai/project-manager.md)

**Purpose**: Implementation logic and mode-based workflow

**Structure**:
```
YAML Frontmatter (standard)
  ├── name
  ├── description (includes "Use PROACTIVELY for")
  ├── tools
  └── skills

Agent Directives (THIS IS NEW)
  ├── Agent responsibility
  ├── Proactive trigger conditions
  └── Success criteria

Mode-Based Directives (THIS IS NEW)
  ├── INITIALIZATION mode
  │  ├── When triggered
  │  ├── Workflow steps
  │  ├── Language handling
  │  ├── Skill delegation
  │  ├── Validation checkpoints
  │  └── Failure recovery
  │
  ├── AUTO-DETECT mode
  │  ├── When triggered
  │  ├── Workflow steps
  │  └── ...
  │
  ├── SETTINGS mode
  │  ├── When triggered
  │  ├── Tab-based workflow
  │  ├── Multi-tab handling
  │  └── ...
  │
  └── UPDATE mode
     ├── When triggered
     ├── Template merging
     └── ...

Skill Delegation Directives (THIS IS NEW)
  ├── Which skills for which modes
  ├── How to invoke skills
  ├── What to expect from skills
  └── Error handling from skills

Language Handling Directives (THIS IS NEW)
  ├── How to read language from config
  ├── How to use language in output
  ├── When to update language
  └── Language consistency rules

Validation Directives (THIS IS NEW)
  ├── Checkpoint 1: After Tab 1 (language)
  ├── Checkpoint 2: After Tab 3 (git)
  ├── Checkpoint 3: Before final write
  └── Validation failure recovery

State Management Directives (THIS IS NEW)
  ├── Context accumulation
  ├── Configuration merging
  ├── Error state handling
  └── Context saving

User Interaction Directives (THIS IS NEW)
  ├── AskUserQuestion usage rules
  ├── Question structure
  ├── Response processing
  └── Language in questions

Content (markdown body)
  ├── Workflow overview
  ├── Phase-by-phase instructions
  ├── Integration with skills
  ├── Error recovery paths
  └── Success patterns
```

### Layer 3: Skill Files (.claude/skills/moai-*/SKILL.md)

**Purpose**: Reusable capability specifications

**Structure**:
```
YAML Frontmatter (standard)
  ├── name
  ├── description
  └── model

Skill Directives (THIS IS NEW)
  ├── What this skill does
  ├── When to call it
  ├── Input specification
  ├── Output specification
  └── Error cases

Integration Directives (THIS IS NEW)
  ├── How agent calls this skill
  ├── Required context
  ├── Response format
  └── Error handling expected

Language Directives (if applicable)
  ├── What gets translated
  ├── What stays English
  └── Language parameter handling

Content (markdown body)
  ├── Detailed implementation guide
  ├── Examples
  └── Reference documentation
```

---

## Directive Categories by Location

### Command Directives (.claude/commands/moai/0-project.md)

These directives live IN the command file as structured markdown sections:

**Section 1: User Directives**
- What is this command? (user perspective)
- When should user run it? (4 use cases)
- What does user get? (deliverables)
- What language? (user's language)
- No technical jargon

**Section 2: Entry Point Directives**
- What arguments does it accept?
- How does it detect mode?
- What validation happens first?
- When does it delegate to agent?
- How is language loaded?

**Section 3: Tool Usage Constraints**
- Only Task() and AskUserQuestion() allowed
- NO Read/Write/Edit/Bash
- Why these constraints exist
- How to request agent capability if needed

**Section 4: Error Handling at Entry**
- What can go wrong at entry point?
- How are entry-level errors handled?
- Where does error recovery happen?
- What does user see?

**Section 5: Integration with project-manager Agent**
- When is agent called?
- What context is passed?
- What response is expected?
- How are agent failures handled?

**Section 6: Skill References**
- Which skills does this command use?
- When are they invoked?
- Direct reference links
- Why each skill matters

**Format**: Regular markdown with clear section headers and structured lists

**Embedded Example**:
```markdown
## 2. Entry Point Directives

### 2.1 Argument Detection

DIRECTIVE: The command MUST detect user intent from arguments:
- No arguments → auto-detect mode
- "setting" → settings mode
- "update" → update mode

DIRECTIVE: Before routing anywhere:
1. Verify config.json exists/valid
2. Load language from config
3. Validate arguments
4. Route to exactly ONE mode

### 2.2 Mode Delegation

DIRECTIVE: Call project-manager agent with:
```python
Task(
    subagent_type="project-manager",
    prompt=f"Mode: {detected_mode}\nLanguage: {loaded_language}\n..."
)
```

### 2.3 Entry Point Failures

| Error | Recovery |
|-------|----------|
| Config missing | INITIALIZATION mode |
| Config invalid | Offer restore or reinit |
| ...
```

---

## Agent Directives (.claude/agents/moai/project-manager.md)

These directives are EMBEDDED IN the agent file as structured sections:

**Section 1: Agent Responsibility Directive**
- What is this agent responsible for?
- When should it be called?
- Success criteria for this agent
- Failure modes and recovery

**Section 2-5: Mode-Specific Directives** (one per mode)

**For each mode (INITIALIZATION, AUTO-DETECT, SETTINGS, UPDATE)**:

```markdown
## Mode: [NAME]

### When This Mode Is Triggered

DIRECTIVE: This mode executes when:
- [condition 1]
- [condition 2]
- [not when: condition 3]

### Mode Workflow

DIRECTIVE: Execute these steps IN ORDER:

1. Language Setup
   - Read/confirm language
   - Use for all output
   - Validate language code

2. User Interview
   - Ask questions (AskUserQuestion only)
   - In user's language
   - Process responses

3. Skill Delegation
   - Call moai-project-config-manager Skill
   - Pass updates
   - Get success/failure response

4. Validation
   - Check validation rules
   - Report results
   - If failure: recovery

5. Completion
   - Confirm what changed
   - Next steps
   - Context save

### Skill Delegation Rules

DIRECTIVE: When calling skills:
- Skill("moai-project-config-manager") for file ops
- Skill("moai-project-batch-questions") for interview
- Skill("moai-project-language-initializer") for language
- Always wait for skill response
- Handle skill errors gracefully

### Validation Checkpoints

DIRECTIVE: Validate at these points:
- After [event 1]: [validation rule 1]
- After [event 2]: [validation rule 2]
- Before write: [validation rule 3]

If validation fails: [recovery step]

### Error Recovery

DIRECTIVE: When errors occur:
- [error type 1] → [recovery step 1]
- [error type 2] → [recovery step 2]

### User Interaction

DIRECTIVE: Use AskUserQuestion for:
- [decision 1]
- [decision 2]
- All confirmations
- Never use other output method

Structure questions as:
```python
{
    "question": "...",  # in user's language
    "header": "...",    # max 12 chars
    "options": [...]    # clear descriptions
}
```
```

**Section 6: Language Handling Directive**
- How to read language from config
- How to use it for output
- When to update it
- Language consistency checks

**Section 7: State Management Directive**
- How to accumulate context
- How to merge configurations
- How to save state
- Error state handling

**Section 8: Integration with Skills Directive**
- Which skills are required
- When to call each skill
- What to pass to skills
- How to handle skill responses
- Skill failure recovery

---

## Skill Directives (.claude/skills/moai-*/SKILL.md)

**Skill Directive Template**:

```markdown
---
name: moai-project-config-manager
description: Manages configuration file operations with atomic updates and rollback
model: haiku
---

# moai-project-config-manager Skill

## Skill Directives

### What This Skill Does

DIRECTIVE: This skill is responsible for:
- Reading .moai/config/config.json
- Creating backups before writes
- Validating configuration updates
- Writing changes atomically
- Implementing rollback on failure
- Reporting exact changes made

### When to Call This Skill

DIRECTIVE: Call this skill when:
- User has completed configuration changes
- Updates are ready to be persisted
- Atomic write is required
- Backup/rollback safety needed

**Do NOT call when**:
- Reading configuration (use direct read)
- Single value lookup (unnecessary overhead)

### Input Specification

DIRECTIVE: Pass the following to skill:

```python
Skill(
    "moai-project-config-manager",
    instruction="atomic_update",
    config_path=".moai/config/config.json",
    updates={
        "user.name": "new_name",
        "language.conversation_language": "en",
        ...
    },
    backup_enabled=true,
    validation_rules={
        "user.name": { "type": "string", "required": true },
        ...
    }
)
```

### Output Specification

DIRECTIVE: Expect the following response:

```python
{
    "success": true/false,
    "operation": "update",
    "backed_up": "path/to/backup",
    "changes": {
        "added": {...},
        "modified": {...},
        "deleted": {...}
    },
    "validation": {
        "passed": true/false,
        "errors": [...]
    }
}
```

### Error Cases

DIRECTIVE: Handle these errors:
- Backup failure: Continue with warning
- Write failure: Rollback and error
- Validation failure: Return errors, don't write

### Language Handling

DIRECTIVE: This skill:
- Does NOT translate config keys (English only)
- Does translate error messages (user's language)
- Receives language parameter from agent
- Uses language for all user-facing messages

## Implementation Guide

[Detailed implementation content...]

## Examples

[Code examples...]

## Reference

[Technical reference...]
```

---

## How Information Flows Through Layers

```
User Input
    ↓
Command (.claude/commands/moai/0-project.md)
  - Parse arguments
  - Detect mode
  - Load language
  - Validate entry
    ↓ (via Task())
Agent (.claude/agents/moai/project-manager.md)
  - Read/confirm language
  - Execute mode-specific workflow
  - Call skills
    ↓ (via Skill())
Skills (.claude/skills/moai-*/SKILL.md)
  - Execute specific operations
  - Report results
    ↓
Back to Agent
  - Process skill results
  - Validate
  - Recover on errors
    ↓
Back to Command
  - Report completion
  - Show next steps
    ↓
User Output (in user's language)
```

**Key Point**: Each layer has directives embedded IN the file that defines it.

---

## Directive Embedding Standards

### For Commands

**Format**: Markdown sections with clear headers

**Example**:
```markdown
## 2. Entry Point Directives

### 2.1 Mode Detection

DIRECTIVE: Detect user intent by analyzing arguments...

**Decision Tree**:
- No args → Check config state
  - Config exists → AUTO-DETECT
  - Config missing → INITIALIZATION
- "setting" → SETTINGS
- ...

### 2.2 Language Loading

DIRECTIVE: Load language before any user interaction...

Steps:
1. Read .moai/config/config.json
2. Extract conversation_language field
3. Use for ALL output
4. Pass to project-manager agent
5. If config missing, use CLI default
```

**Do NOT use**:
- ❌ Code blocks for directives (use markdown structure)
- ❌ Tables for complex logic (use decision trees)
- ❌ Separate DIRECTIVES.md files (embed in place)

### For Agents

**Format**: Structured markdown with clear sections per mode

**Each mode section includes**:
```markdown
## Mode: INITIALIZATION

### When Triggered
[conditions]

### Workflow Steps
[numbered list with directives]

### Skill Delegation
[skill calls with directives]

### Validation
[validation rules]

### Error Recovery
[error → recovery mapping]

### User Interaction
[AskUserQuestion directives]
```

**Do NOT use**:
- ❌ Single large workflow section (break by mode)
- ❌ Pseudocode (use numbered steps with directives)
- ❌ Separate error handling docs (embed in each mode)

### For Skills

**Format**: Structured SKILL.md with clear sections

**Always include**:
```markdown
## Skill Directives

### What This Skill Does
[DIRECTIVE: responsibility statement]

### When to Call This Skill
[DIRECTIVE: when + when NOT to call]

### Input Specification
[DIRECTIVE: expected input structure + examples]

### Output Specification
[DIRECTIVE: expected output structure + examples]

### Error Cases
[DIRECTIVE: error handling rules]

### Language Handling
[DIRECTIVE: language-specific rules]
```

---

## Directive Style Guide

### Writing Directives

**Each directive MUST:**

1. **Start with DIRECTIVE**: "DIRECTIVE: [statement]"
2. **Be actionable**: Specifies WHAT to do, not WHY (unless critical)
3. **Be unambiguous**: No interpretation needed
4. **Include constraints**: What NOT to do (marked with ❌)
5. **Be specific**: Not "validate input" but "validate language code matches [list]"

**Examples**:

```markdown
DIRECTIVE: Before delegating to project-manager agent, verify:
- config.json exists OR trigger INITIALIZATION
- If config exists and invalid JSON → offer restore or reinit
- Load conversation_language from config
- Never proceed without valid language

DIRECTIVE: In INITIALIZATION mode, ask these questions IN ORDER:
1. Project name (required, min 1 char, max 100 chars)
2. Project description (optional, max 500 chars)
3. Owner/team (required)
4. Git mode (required, options: personal/team)

DIRECTIVE: Language MUST be confirmed before any other configuration questions
- Do NOT skip this step
- Do NOT assume user's language
- Use AskUserQuestion with clear language options
- Pass confirmed language to all subsequent interactions
```

**Do NOT write**:
```markdown
❌ Validate user input
❌ Handle errors gracefully
❌ Make sure language is set
❌ Ask user for configuration
```

### Organizing Directives

**Hierarchy**:
1. **Core Directives** (MUST do)
   - Non-negotiable requirements
   - Labeled as "DIRECTIVE:"

2. **Supporting Directives** (SHOULD do)
   - Best practices
   - Labeled as "GUIDELINE:"

3. **Context** (explanatory)
   - Why directives matter
   - Regular markdown paragraphs

**Example structure**:
```markdown
## Section Title

### Subsection

DIRECTIVE: [unambiguous requirement]
[numbered steps or decision rules]

GUIDELINE: [best practice]
[suggestions for implementation]

**Context**: [explanation of why this matters]
[regular paragraph explaining the "why"]
```

---

## Validation: Are Directives Properly Embedded?

### Checklist for Command File

- [ ] YAML frontmatter includes tool constraints
- [ ] User Directives section (what/when/why)
- [ ] Entry Point Directives section (mode detection)
- [ ] Tool Constraints section (only Task/AskUserQuestion)
- [ ] Error Handling section (entry-level errors)
- [ ] Integration section (agent delegation)
- [ ] Skill References section (which skills)
- [ ] No directives in separate `.moai/directives/` files
- [ ] All directives use "DIRECTIVE:" prefix

### Checklist for Agent File

- [ ] YAML frontmatter with skill list
- [ ] Agent Responsibility section
- [ ] Separate section for EACH mode
  - [ ] When Triggered subsection
  - [ ] Workflow Steps subsection
  - [ ] Skill Delegation subsection
  - [ ] Validation Checkpoints subsection
  - [ ] Error Recovery subsection
  - [ ] User Interaction subsection
- [ ] Language Handling Directives section
- [ ] State Management Directives section
- [ ] Integration with Skills section
- [ ] All mode workflows use numbered steps with directives
- [ ] Error recovery is embedded in each mode (not separate)

### Checklist for Skill File

- [ ] Skill Directives section at top
  - [ ] What This Skill Does
  - [ ] When to Call It
  - [ ] Input Specification
  - [ ] Output Specification
  - [ ] Error Cases
  - [ ] Language Handling (if applicable)
- [ ] Implementation Guide section
- [ ] Examples section
- [ ] Reference section
- [ ] All directives use "DIRECTIVE:" prefix

---

## Migration Path: From Separate Files to Embedded

**Current State**:
```
.moai/directives/
├── README.md
├── 0-project-command-directive.md
├── 0-project-error-recovery-guide.md
└── 0-project-executive-summary.md

.claude/commands/moai/
└── 0-project.md (missing directives)

.claude/agents/moai/
└── project-manager.md (partial directives)
```

**Target State**:
```
.moai/directives/
└── (empty - all directives migrated to .claude/)

.claude/commands/moai/
└── 0-project.md (CONTAINS all command directives)

.claude/agents/moai/
└── project-manager.md (CONTAINS all agent directives)

.claude/skills/moai-*/
└── SKILL.md (CONTAINS skill directives)
```

### Migration Steps

**Phase 1: Prepare**
- [ ] Review current directives in `.moai/directives/`
- [ ] Understand current command/agent structure
- [ ] Identify which directives belong where

**Phase 2: Extract Directives by Layer**

From current documents, extract:
1. **Command Directives** → embedding in `.claude/commands/moai/0-project.md`
   - Entry point logic
   - Mode detection
   - Tool constraints
   - Error handling at entry
   - Skill references

2. **Agent Directives** → embedding in `.claude/agents/moai/project-manager.md`
   - Agent responsibility
   - Mode-specific workflows
   - Skill delegation
   - Language handling
   - Validation logic
   - Error recovery

3. **Skill Directives** → embedding in respective skill files
   - For each skill used, extract its directive section

**Phase 3: Embed Directives**

For each target file:
1. Identify appropriate sections
2. Add new sections following directive style guide
3. Convert extracted directives to embedded format
4. Verify consistency with other layers
5. Test that directives make sense in context

**Phase 4: Delete Separate Directive Files**

Once embedded:
- [ ] Archive `.moai/directives/` for reference
- [ ] Document that directives live in `.claude/`
- [ ] Update CLAUDE.md to reference official files
- [ ] Delete separate directive files

**Phase 5: Maintain Going Forward**

- [ ] When command changes, update its directives
- [ ] When agent changes, update its directives
- [ ] When skill changes, update its directives
- [ ] No separate directive documents created
- [ ] Official files ARE the directives

---

## Directive Consistency Across Layers

### Language Consistency Directive

**Consistency Rule**: All three layers MUST agree on language handling

**In Command file**:
```markdown
### Tool Constraints

DIRECTIVE: Load user's language from config and pass to agent
- Use only AskUserQuestion() for user input
- All output must be in user's conversation_language
- Pass language to project-manager agent
```

**In Agent file**:
```markdown
## Language Handling Directives

DIRECTIVE: Use language passed from command for ALL output
- Read from passed language parameter
- Use for all AskUserQuestion() calls
- Use for all error messages
- Pass to skills that need it
```

**In Skill file**:
```markdown
### Language Handling

DIRECTIVE: Receive language from agent
- Use for error messages (user-facing)
- Keep config keys English only
- Translate success/failure messages
```

**Validation**: If one layer changes language handling, all must be updated together.

### Tool Usage Consistency Directive

**Consistency Rule**: All layers use same tools in same way

**In Command**:
```markdown
## Tool Usage Constraints

DIRECTIVE: Use ONLY these tools:
- Task() for delegating to agent
- AskUserQuestion() for user interaction

NO Read, Write, Edit, Bash usage allowed
```

**In Agent**:
```markdown
## Tool Usage Constraints

DIRECTIVE: Use ONLY these tools:
- Skill() to call skills (which internally use proper tools)
- AskUserQuestion() for user interaction

Delegate file operations to skills (don't do directly)
```

**In Skill**:
```markdown
### Tool Usage

DIRECTIVE: You may use any tools needed:
- Read/Write/Edit for file operations
- Bash for safe operations
- These are delegated from agent for this purpose
```

---

## Success Criteria: Properly Embedded Directives

**Specification is complete when**:

1. **All command directives are in `.claude/commands/moai/0-project.md`**
   - No separate command directive file
   - Clear entry point directives
   - Tool constraints documented
   - Error handling specified

2. **All agent directives are in `.claude/agents/moai/project-manager.md`**
   - One section per mode
   - Skill delegation specified
   - Language handling embedded
   - Validation rules clear
   - Error recovery documented

3. **All skill directives are in respective skill files**
   - Skill Directives section at top
   - Input/output specs clear
   - Error cases specified
   - Language handling (if applicable)

4. **Consistency across layers**
   - Language handling consistent
   - Tool usage consistent
   - Skill delegation consistent

5. **No directives in `.moai/directives/`**
   - All migrated to official files
   - `.moai/directives/` archive only (for reference)
   - CLAUDE.md references official files

6. **Directives are actionable**
   - Every "DIRECTIVE:" has clear what/when/how
   - No interpretation needed
   - Can be directly implemented
   - Can be directly tested

---

## Implementation Notes

### For Developers

When implementing the command/agent:
1. Read the directive sections FIRST
2. Every directive is a requirement
3. If directive and code differ → directive wins (code needs fixing)
4. Test against directives (not just code review)

### For Code Reviewers

When reviewing changes:
1. Check that changes match directives
2. If directives need updating, require them to be updated too
3. Directives are part of the code contract

### For QA Engineers

When planning tests:
1. Each directive becomes test case(s)
2. Validation checkpoints are test gates
3. Error recovery is test scenario
4. Success criteria are test acceptance rules

### For Documentation

When writing user guides:
1. Reference the directives as specification
2. Directives are authoritative
3. User guides explain directives in user language
4. Official files are the source of truth

---

## Summary

**Directive Architecture v2.0 Specification**

| Aspect | Specification |
|--------|---|
| Where directives live | IN official `.claude/` files, not separate |
| Command directives | `.claude/commands/moai/0-project.md` sections |
| Agent directives | `.claude/agents/moai/project-manager.md` sections |
| Skill directives | Each skill file's SKILL Directives section |
| Format | Structured markdown with "DIRECTIVE:" prefix |
| Style | Actionable, unambiguous, testable statements |
| Consistency | All layers must be synchronized |
| Single source of truth | The official files ARE the directives |
| Migration | Move from separate files to embedded |
| Going forward | No separate directive documents |

**This specification ensures that**:
- Directives are always current (no sync issues)
- Developers can't ignore directives (they're right there)
- Specification is testable and reviewable
- One source of truth
- Reduced maintenance burden

---

**Document Version**: 2.0.0 (Specification)
**Status**: Ready for implementation
**Next Step**: Implement embedded directives in official files following this specification
