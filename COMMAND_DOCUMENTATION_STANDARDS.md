# MoAI-ADK Command Documentation Standards
## Official Format, Structure & Guidance Writing Guidelines

---

## Executive Summary

MoAI-ADK command files (`.claude/commands/alfred/*.md`) follow a **highly structured** documentation format that balances narrative guidance with executable instructions. This document defines:

1. **YAML Frontmatter Standards** - Metadata structure
2. **Content Architecture** - Section hierarchy and flow
3. **Narrative vs Code Balance** - How to write instructions vs examples
4. **Flags & Parameters Documentation** - Proper format for arguments
5. **Decision Points & AskUserQuestion** - How to design user interactions
6. **Workflow Description** - Best practices for depicting complex flows

---

## 1. YAML Frontmatter Standards

### Required Fields

Every command file MUST begin with YAML frontmatter:

```yaml
---
name: alfred:1-plan
description: "Define specifications and create development branch"
argument-hint: Title 1 Title 2 ... | SPEC-ID modifications
allowed-tools:
  - Read
  - Write
  - Edit
  - MultiEdit
  - Grep
  - Glob
  - TodoWrite
  - Bash(git:*)
  - Bash(gh:*)
  - Bash(rg:*)
  - Bash(mkdir:*)
---
```

### Field Descriptions

| Field | Purpose | Format | Example |
|-------|---------|--------|---------|
| **name** | Unique command identifier | `alfred:COMMAND-NAME` | `alfred:1-plan` |
| **description** | One-sentence purpose | Present tense | "Define specifications and create development branch" |
| **argument-hint** | Parameter format hint | Pipe-separated options with descriptions | `"[patch\|minor\|major] [--dry-run] [--testpypi]"` |
| **allowed-tools** | Tools this command can use | Array of tool patterns | `Bash(git:*)`, `Task`, `WebFetch` |

### Tool Pattern Rules

- **Read**: Allowed without restrictions
- **Write/Edit**: Allowed without restrictions
- **Bash(cmd:*)**: Specific bash command whitelist (e.g., `Bash(git:*)`, `Bash(python3:*)`)
- **Task**: Allowed for sub-agent invocation
- **Specialized**: WebFetch, Grep, Glob, MultiEdit as needed

---

## 2. Content Architecture

### Recommended Section Structure

#### A. Command Overview (lines 3-50)
```markdown
# 🎯 Symbol MoAI-ADK Phase/Step: Title

> **Note**: Interactive prompts use [description]

<!-- @CODE:ALF-WORKFLOW-NNN:CMD-NAME -->

**4-Step Workflow Integration**: [Explanation of workflow integration]

## 🎯 Command Purpose
[1-2 paragraphs explaining command's high-level goal]

## 💡 Execution philosophy
[Design principle or philosophy]

### Main scenarios (3 examples)
#### Scenario 1: [Primary approach] ⭐
#### Scenario 2: [Alternative approach]
#### Scenario 3: [Optional approach]
```

**Guidelines**:
- Use emoji consistently (🎯🔧⚙️📋etc.)
- Keep purpose explanation **user-facing and non-technical**
- Show 3 usage scenarios with execution flow
- End with pointer to related documentation

#### B. Execution Flow Overview
```markdown
## 📋 Execution flow

1. **[Phase Name]**: [Description]
2. **[Phase Name]**: [Description]
3. **[Phase Name]**: [Description]
4. **[Phase Name]**: [Description]

## 🧠 Associated Skills & Agents
[Table showing agents and their skills]

## 🔗 Associated Agent
- **Primary**: [agent-name] (role) - [responsibility]
- **Secondary**: [agent-name] (role) - [responsibility]

## 💡 Example of use
Users can run commands like:
- `/alfred:1-plan` - [description]
- `/alfred:1-plan "title"` - [description]
- `/alfred:1-plan SPEC-001 "title"` - [description]
```

**Guidelines**:
- Keep flow descriptions **high-level and sequential**
- Use consistent bullet points
- Associate specific agents with responsibilities
- Provide 2-3 clear usage examples

#### C. Detailed Phase Implementation (Main Body)

Detailed sections follow consistent structure:
- Phase A (OPTIONAL) with conditions
- Phase B (REQUIRED) with complete invocations
- Decision points with AskUserQuestion
- Response processing with mapped actions

---

## 3. Narrative vs Code Balance

### What Should Be Narrative (Plain English)

Use **narrative explanations** for:
- High-level workflow descriptions
- Rationale and philosophy
- User-facing purpose statements
- Conditional logic and decision trees
- Error handling strategies
- Best practices and guidelines

### What Should Be Code Examples

Use **code examples** for:
- Exact Task tool invocations with all parameters
- AskUserQuestion complete structures
- Bash command syntax and patterns
- File manipulation operations
- Conditional logic with specific syntax

### The 70/30 Rule

**Recommended balance**:
- **70% narrative/guidance** - Explain WHY and WHEN
- **30% code examples** - Show HOW with specific syntax

**Anti-patterns**:
- ❌ Narrative without supporting code → Users don't know exact syntax
- ❌ Code blocks without explanation → Users don't understand rationale
- ❌ Pseudo-code mixed with real code → Creates confusion

---

## 4. Flags & Parameters Documentation

### Format for Argument Hints

```yaml
argument-hint: "[patch|minor|major] [--dry-run] [--testpypi]"
```

**Syntax**:
- `[option1|option2]` = Mutually exclusive choices (required)
- `[--flag]` = Optional flag
- `path` = File/directory path
- Separate multiple with spaces

### In-Document Flag Explanation

Flags should be documented with:
1. **Purpose** - What does the flag do?
2. **Effect** - How does it change execution?
3. **Comparison table** - Side-by-side flag options
4. **Example invocations** - Real usage

---

## 5. Batched Design Pattern

### What Is Batching?

**Batching** = Combining 2+ related questions in **ONE AskUserQuestion call** to reduce user interaction turns.

**UX Improvement**:
- Sequential: 3 questions = 3 turns
- Batched: 3 questions = 1 turn (66% reduction)

### Implementation

```python
AskUserQuestion(
    questions=[
        {
            "question": "Question 1?",
            "header": "Category 1",
            "multiSelect": false,
            "options": [...]
        },
        {
            "question": "Question 2?",
            "header": "Category 2",
            "multiSelect": false,
            "options": [...]
        }
    ]
)
```

---

## 6. Workflow Visualization

### ASCII Flow Diagrams

Use box drawings for clarity:
- `┌`, `─`, `├`, `└`, `│` for structure
- `↓` for sequential flow
- Indent clearly with 2 spaces
- Keep under 10 lines

**Example**:
```
┌─────────────────────────────┐
│ STEP 1: Analysis            │
├─────────────────────────────┤
│ Phase A (OPTIONAL)          │
│ ↓                           │
│ Phase B (REQUIRED)          │
│ ↓                           │
│ Decision Point              │
└─────────────────────────────┘
```

### Decision Trees

Show conditional branches with alignment:
```
Command Input
  ↓
Check condition
  ├─ Branch A → Action A
  ├─ Branch B → Action B
  └─ Branch C → Action C
```

---

## 7. AskUserQuestion Pattern

### Complete Structure

Every AskUserQuestion must include:
1. **Full code block** - Copy-paste ready
2. **Response Processing** - Exact string matching
3. **Action mapping** - What happens for each response

```python
AskUserQuestion(
    questions=[
        {
            "question": "User-facing question?",
            "header": "Category",
            "multiSelect": false,
            "options": [
                {
                    "label": "Option 1",
                    "description": "What happens if selected"
                },
                {
                    "label": "Option 2",
                    "description": "What happens if selected"
                }
            ]
        }
    ]
)
```

**Response Processing**:
```markdown
- **"Option 1"** (`answers["0"] === "Option 1"`) → Execute Action 1
  - Step 1
  - Step 2
  - Display: "Message"

- **"Option 2"** (`answers["0"] === "Option 2"`) → Execute Action 2
  - Step 1
  - Step 2
  - Display: "Message"
```

---

## 8. Best Practices Summary

### ✅ DO

1. Use narrative for "why" and "when" decisions
2. Use code for "how" and "what" technical details
3. Provide complete Task tool invocations - Copy-paste ready
4. Show AskUserQuestion with exact response handling
5. Use tables for comparison and decision matrices
6. Batch related questions - Minimize user interaction turns
7. Link to Skills using `Skill("skill-name")` format
8. Document all flags with purpose and effect
9. Show example commands with realistic output
10. Keep sections under 200 lines

### ❌ DON'T

1. Mix pseudo-code with real code
2. Write incomplete code examples
3. Explain code in narrative when syntax matters
4. Use placeholder values
5. Nest conditionals too deeply
6. Document flags only in argument-hint
7. Forget AskUserQuestion flow
8. Make sections too long
9. Use inconsistent emoji
10. Reference undefined terms

---

## 9. Language & Tone Guidelines

### User-Facing Content
- **Language**: User's selected `conversation_language`
- **Tone**: Professional, instructional, action-oriented
- **Tense**: Present continuous or imperative
- **Address**: Speak to user directly

### Code & Technical Content
- **Language**: English (global technical standard)
- **Syntax**: Exact, copy-paste ready
- **Comments**: Explain WHY, not WHAT

### Section Headers
Use: **emoji** + **imperative verb** + **noun**

✅ Good:
- "🚀 STEP 2: Execute Task"
- "📋 Execution flow"
- "⚙️ Phase B: Implementation (REQUIRED)"

❌ Bad:
- "Execute task code sections"
- "Flow and execution"

---

## 10. Validation Checklist

Before publishing:

- [ ] YAML frontmatter complete and valid
- [ ] All allowed-tools properly whitelisted
- [ ] First 100 lines explain purpose clearly
- [ ] Optional vs Required phases clearly marked
- [ ] All AskUserQuestion blocks complete with response handling
- [ ] All Task tool invocations copy-paste ready
- [ ] Tables use consistent column format
- [ ] Diagrams use consistent box drawing
- [ ] Links to Skills use `Skill("skill-name")` format
- [ ] Code blocks use triple backticks
- [ ] No pseudo-code mixed with real syntax
- [ ] Sections under 200 lines
- [ ] Emoji consistent throughout
- [ ] Examples show realistic invocations
- [ ] Error handling documented
- [ ] Final section includes "Next steps"

---

## Conclusion

MoAI-ADK command documentation balances:
- **70% narrative** for understanding workflow and rationale
- **30% code** for precise syntax and execution
- **User-centered design** with clear decision points
- **Batched interactions** to minimize turns
- **Copy-paste readiness** for all code
- **Emoji consistency** for visual scanning

This ensures **clarity, consistency, and maintainability** across all Alfred commands.
