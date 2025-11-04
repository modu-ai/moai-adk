# Analysis Summary: MoAI-ADK Command Documentation Standards

## Investigation Scope

Analyzed MoAI-ADK's official command documentation format by examining:
1. ✅ 4 existing command files (.claude/commands/alfred/0-project.md, 1-plan.md, 2-run.md, 3-sync.md)
2. ✅ Package template structure (src/moai_adk/templates/.claude/commands/)
3. ✅ CLAUDE.md project guidelines
4. ✅ Global CLAUDE.md standards (/.claude/CLAUDE.md)

## Key Findings

### 1. YAML Frontmatter Format (STRICT)

```yaml
---
name: alfred:COMMAND-NAME
description: "One-sentence present-tense purpose"
argument-hint: "[option1|option2] [--flag] path"
allowed-tools:
  - Read
  - Write
  - Edit
  - MultiEdit
  - Bash(git:*)
  - Task
  - WebFetch
---
```

**Critical Rules**:
- ✅ MUST precede all markdown content
- ✅ Argument-hint uses pipe for mutually exclusive choices
- ✅ Allowed-tools uses specific Bash pattern whitelist (Bash(cmd:*))
- ❌ DO NOT use incomplete tool lists

### 2. Content Architecture (70/30 Principle)

**Balance discovered**:
- **70% Narrative Guidance**: Explain WHY and WHEN
  - High-level workflow descriptions
  - Decision trees and conditionals
  - Best practices and rationale
  - User-facing explanations

- **30% Code Examples**: Show HOW and WHAT
  - Complete Task tool invocations (copy-paste ready)
  - AskUserQuestion structures with response handling
  - Bash command patterns
  - Exact syntax (no pseudo-code)

**Example from 1-plan.md**:
- Lines 1-111: Narrative (command purpose, philosophy, scenarios)
- Lines 112-240: Code examples (Phase A/B invocations)
- Lines 241-481: Mixed (narrative + AskUserQuestion complete blocks)

### 3. Mandatory Section Structure

**Consistent across all commands** (0-project, 1-plan, 2-run, 3-sync):

```
1. YAML Frontmatter
2. Title with emoji (🎯, 🏗️, ⚒️, 📚)
3. Note about interactive prompts
4. @CODE TAG reference
5. 4-Step Workflow Integration statement
6. 🎯 Command Purpose (1-2 paragraphs)
7. 💡 Execution philosophy (3+ scenarios)
8. 📋 Execution flow (numbered list)
9. 🧠 Associated Skills & Agents (table)
10. 🔗 Associated Agent (role list)
11. 💡 Example of use (2-3 examples)
12. 🔍 STEP 1: Analysis & Planning (main content)
13. STEP 2+: Implementation (main content)
14. Final Step: AskUserQuestion for next action
15. Next steps: Guidance for continuation

**Total typical length**: 2,000-2,400 lines
```

### 4. Phase Architecture Pattern

**Discovered consistent pattern** (PHASE A/B paradigm):

```markdown
## 🔍 STEP N: [Description]

STEP N consists of **two independent phases**:

### 📋 STEP N Workflow Overview
[ASCII diagram showing flow]

**Key Points**:
- **Phase A is optional** - [when to skip]
- **Phase B is required** - [when to run]
- **Results flow forward** - [data passing]

---

### 🔍 Phase A: [Name] (OPTIONAL)

**Use [agent] when [condition].**

#### When to use Phase A:
- ✅ Scenario 1
- ✅ Scenario 2
- ❌ Negative case

#### How to invoke [Agent]:

```
Task tool call:
- subagent_type: "agent-name"
- description: "[task]"
- prompt: "[instruction]"
```

---

### ⚙️ Phase B: [Name] (REQUIRED)

[Repeat Phase A structure]
```

**Pattern present in ALL commands**:
- OPTIONAL Phase A for optional exploration
- REQUIRED Phase B for mandatory core logic
- Clear decision points about when to skip Phase A
- Results flowing from A to B

### 5. AskUserQuestion Implementation

**Discovered Best Practices**:

1. **Always includes complete code block**:
   ```python
   AskUserQuestion(
       questions=[
           {
               "question": "...",
               "header": "...",
               "multiSelect": false,
               "options": [...]
           }
       ]
   )
   ```

2. **Response Processing with exact string matching**:
   ```markdown
   - **"Option 1"** (`answers["0"] === "Option 1"`) → Action 1
   - **"Option 2"** (`answers["0"] === "Option 2"`) → Action 2
   ```

3. **Each response maps to specific next step**

4. **Batching pattern**: 2-4 related questions in 1 call
   - Example: 0-project batches language + nickname + GitHub settings
   - UX improvement: 3 questions = 1 turn (vs 3 turns)

**Found in**:
- 0-project.md: Lines 226-627 (3 batched questions)
- 1-plan.md: Lines 453-505 (4 options with detailed responses)
- 2-run.md: Lines 239-267 (4-option approval decision)
- 3-sync.md: Lines 310-344 (4-option sync strategy)

### 6. Task Tool Invocation Format

**Mandatory pattern** (found in ALL commands):

```
Call the Task tool:
- subagent_type: "[agent-name]"
- description: "[brief task description]"
- prompt: """[Multi-line instruction block]

[Can span multiple paragraphs]

Special variables like {{CONVERSATION_LANGUAGE}} preserved
"""
```

**Critical**:
- ✅ Triple backticks for multi-line prompts
- ✅ All parameters specified: subagent_type, description, prompt
- ✅ Variables like {{CONVERSATION_LANGUAGE}} preserved as-is
- ✅ Language setting instructions included in prompt
- ✅ Explicit Skill() references when needed

### 7. Flag Documentation Pattern

**From release-new.md**:

```markdown
### Dry-Run Mode Guide

**Dry-Run mode** simulates execution without actual changes.

### Usage Method

```bash
/awesome:release-new [patch|minor|major] --dry-run
```

### Dry-Run Execution (What runs):
✅ **Actually runs**: pytest, ruff, mypy, bandit
❌ **Does NOT run**: File modifications, Git commits, PyPI deploy

### Comparison Table

| Mode | Flag | Changed Files | Dry-Run Only |
|------|------|----------------|--------------|
| Immediate | none | YES | NO |
| Dry-Run | --dry-run | NO | YES |
```

**Pattern**:
1. Explain what flag does (narrative)
2. Show usage syntax
3. List execution differences (✅/❌)
4. Use comparison table
5. Real example invocations

### 8. Workflow Visualization Techniques

**Three methods used**:

1. **ASCII Box Diagrams** (6-10 lines):
   ```
   ┌─────────────────────┐
   │ STEP 1: Analysis    │
   ├─────────────────────┤
   │ Phase A (optional)  │
   │ ↓                   │
   │ Phase B (required)  │
   └─────────────────────┘
   ```

2. **Flow Trees** (conditional branches):
   ```
   Command Input
     ↓
   Check condition
     ├─ Branch A → Action A
     ├─ Branch B → Action B
     └─ Branch C → Action C
   ```

3. **Step-by-step flows** (with indentation):
   ```
   /alfred:1-plan SPEC-001
     ↓
   config.json check
     ├─ feature_branch → Create PR
     ├─ develop_direct → Direct commit
     └─ per_spec → Ask user
   ```

### 9. Language & Tone Standards

**Discovered Rules**:

1. **User-facing content**: User's `conversation_language` (Korean for local)
2. **Code & infrastructure**: English (global standard)
3. **Section headers**: Emoji + Verb + Noun
   - ✅ "🚀 STEP 2: Execute Task"
   - ❌ "Run the task code section"

4. **Tense**: Present continuous or imperative
   - "Creating SPEC document..." (continuous)
   - "Create SPEC document" (imperative)

5. **Address**: Direct to user
   - "You can now run..." (good)
   - "The user should run..." (avoid)

### 10. Special Patterns Identified

**Pattern 1: Batched Questions UX**
- 0-project bundles language + nickname + GitHub settings
- Reduces 5-6 turns to 1-2 turns
- Documented with "(N turns → M turns reduction)"

**Pattern 2: Optional → Required Phase Structure**
- Phase A always marked "(OPTIONAL)" with skip conditions
- Phase B always marked "(REQUIRED)" with always-run conditions
- Clear explanation of when Phase A can be skipped

**Pattern 3: EARS Syntax References**
- Links to `Skill("moai-foundation-ears")` for SPEC writing
- Example EARS requirements shown in actual examples
- References maintain traceability

**Pattern 4: Response → Action Mapping**
- Each AskUserQuestion response has **exactly mapped action**
- String matching shown: `answers["0"] === "Label"`
- Action outcome described: "Proceed to Phase 2", "Display message", etc.

## File Statistics

| File | Lines | Sections | AskUserQuestions | Tasks | Phases |
|------|-------|----------|------------------|-------|--------|
| 0-project.md | 2,202 | 25+ | 8 | 4 | 5 |
| 1-plan.md | 880 | 20+ | 3 | 2 | 4 |
| 2-run.md | 793 | 22+ | 3 | 2 | 4 |
| 3-sync.md | 1,085 | 25+ | 4 | 2 | 4 |

## Recommendations

### For New Commands

1. **Start with YAML frontmatter** - Define tools and arguments first
2. **Write purpose section** (lines 3-50) - Keep high-level
3. **Show 3 usage scenarios** - Demonstrate flexibility
4. **Use PHASE A/B pattern** - Clear optional vs required
5. **Always include AskUserQuestion** - Show complete structure with response mapping
6. **Batch related questions** - Improve UX by 50%+
7. **Provide copy-paste code** - All Task invocations must be complete
8. **Use ASCII diagrams** - Help visualize complex flows
9. **End with next steps** - Ask user what to do next

### For Documentation

✅ **DO**:
- Explain WHY in narrative (70%)
- Show HOW in code (30%)
- Use tables for comparisons
- Batch user questions
- Link to Skills explicitly
- Document all flags
- Provide real examples

❌ **DON'T**:
- Mix pseudo-code with real syntax
- Leave code examples incomplete
- Write pseudo-code alone
- Explain syntax in narrative when showing code
- Skip response processing for AskUserQuestion
- Use inconsistent emoji
- Make sections >200 lines

## Deliverable

Created: **COMMAND_DOCUMENTATION_STANDARDS.md**

This comprehensive guide documents:
- YAML frontmatter standards
- Content architecture (70/30 narrative/code)
- Phase A/B optional/required pattern
- AskUserQuestion complete structure
- Batched design for UX improvement
- Task tool invocation format
- Workflow visualization techniques
- Language and tone guidelines
- Validation checklist
- 10 best practices (DO/DON'T)

