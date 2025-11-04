# MoAI-ADK Command Documentation Standards - Investigation Index

## Overview

Complete analysis of MoAI-ADK's official documentation format and command writing standards based on examination of 4 existing command files, package templates, and project guidelines.

## Deliverables

### 1. COMMAND_DOCUMENTATION_STANDARDS.md (399 lines)
**Comprehensive standards guide covering:**
- YAML frontmatter format (required fields, tool patterns)
- Content architecture (recommended section structure)
- Narrative vs Code balance (70/30 principle)
- Flags & parameters documentation
- Batched design pattern for UX
- Workflow visualization techniques
- AskUserQuestion pattern structure
- Best practices (DO/DON'T checklist)
- Language & tone guidelines
- Validation checklist

**Use this when:** Creating new commands, refining documentation style, establishing team standards.

### 2. ANALYSIS_SUMMARY.md (357 lines)
**Detailed investigation findings covering:**
- Investigation scope (4 commands examined)
- YAML frontmatter format standards
- Content architecture (70/30 principle) with examples
- Mandatory section structure (15-section pattern)
- Phase A/B optional/required paradigm
- AskUserQuestion implementation details
- Task tool invocation format
- Flag documentation pattern
- Workflow visualization techniques
- Special patterns identified (batching, EARS, response mapping)
- File statistics table
- Recommendations for new commands

**Use this when:** Understanding the reasoning behind standards, training new team members, validating existing commands.

### 3. COMMAND_EXAMPLES.md (495 lines)
**Real examples from actual MoAI-ADK commands with annotations:**
- Example 1: YAML frontmatter (1-plan.md) with field explanations
- Example 2: Batched questions (0-project.md, lines 226-350)
- Example 3: Phase A/B pattern (2-run.md, lines 99-195)
- Example 4: Complete decision point (2-run.md, lines 234-294)
- Example 5: Comparison table (1-plan.md)
- Example 6: Narrative + code balance (1-plan.md)
- Key takeaways (10 patterns that work, 10 that fail)

**Use this when:** Seeing practical examples, learning by example, validating your implementation.

---

## Key Findings Summary

### YAML Frontmatter (Critical)
```yaml
---
name: alfred:COMMAND-NAME
description: "One-sentence present-tense purpose"
argument-hint: "[option1|option2] [--flag] path"
allowed-tools: [specific whitelist]
---
```

### Content Architecture (70/30)
- **70% Narrative**: WHY and WHEN (flow, philosophy, conditions)
- **30% Code**: HOW and WHAT (complete invocations, exact syntax)

### Phase A/B Pattern (Universal)
- **Phase A (OPTIONAL)**: For optional exploration with skip conditions
- **Phase B (REQUIRED)**: For core logic that always runs
- Data flows from A to B with `$RESULT_VARIABLES`

### AskUserQuestion (Always Complete)
Every question must include:
1. Full code block (copy-paste ready)
2. Response processing with exact string matching
3. Action mapping for each response option

### Batched Questions (UX Best Practice)
- Combine 2-4 related questions in ONE AskUserQuestion call
- Example: 0-project batches language + nickname + GitHub settings
- UX improvement: 50-66% reduction in interaction turns

### Task Tool Format (Mandatory)
```
Call the Task tool:
- subagent_type: "agent-name"
- description: "[task]"
- prompt: """[Multi-line instruction]"""
```

### Special Variables (Preserved)
- `{{CONVERSATION_LANGUAGE}}` - User's language setting
- `$ARGUMENTS` - User input
- `$EXPLORE_RESULTS` - Phase A results flowing to Phase B

---

## Standard Section Structure

All commands follow this 15-section pattern:

1. YAML frontmatter
2. Title with emoji (🎯, 🏗️, ⚒️, 📚)
3. Interactive prompts note
4. @CODE TAG reference
5. 4-Step Workflow Integration
6. Command Purpose (1-2 paragraphs)
7. Execution philosophy (3+ scenarios)
8. Execution flow (numbered)
9. Associated Skills & Agents (table)
10. Associated Agent (role list)
11. Example of use (2-3 examples)
12. STEP 1: Analysis & Planning (Phase A/B)
13. STEP 2+: Implementation (Phase A/B)
14. Final step: AskUserQuestion for next action
15. Next steps guidance

**Typical length**: 2,000-2,400 lines

---

## Documentation Patterns Identified

### Pattern 1: Batched Questions UX
```
Sequential ❌: 3 questions = 3 turns
Batched ✅: 3 questions = 1 turn (66% reduction)
```

### Pattern 2: Optional → Required Phases
```
Phase A: "[OPTIONAL]" with clear skip conditions
Phase B: "[REQUIRED]" with always-run explanation
Results: Phase A → $VARIABLE → Phase B
```

### Pattern 3: Response → Action Mapping
```
- **"Option 1"** (`answers["0"] === "Option 1"`) → Action 1
- **"Option 2"** (`answers["0"] === "Option 2"`) → Action 2
```

### Pattern 4: Complete Task Invocations
```python
Task tool:
- subagent_type: "agent-name"
- description: "[clear task]"
- prompt: """[full instruction block]"""
```

### Pattern 5: Workflow Diagrams
```
ASCII boxes (6-10 lines)
Flow trees (with branches)
Step-by-step flows (with indentation)
```

---

## Best Practices (Quick Reference)

### ✅ DO
- Explain WHY in narrative (70%)
- Show HOW in code (30%)
- Use tables for comparisons
- Batch user questions (2-4 per call)
- Link to Skills explicitly: `Skill("skill-name")`
- Document all flags
- Provide real, copy-paste examples

### ❌ DON'T
- Mix pseudo-code with real syntax
- Leave code examples incomplete
- Write sections >200 lines without breaks
- Use placeholder values
- Skip response processing for AskUserQuestion
- Use inconsistent emoji
- Forget optional/required phase distinction

---

## File Statistics

| File | Lines | Content | Sections |
|------|-------|---------|----------|
| COMMAND_DOCUMENTATION_STANDARDS.md | 399 | Complete standards guide | 10 major |
| ANALYSIS_SUMMARY.md | 357 | Investigation findings | 10 major |
| COMMAND_EXAMPLES.md | 495 | Real examples annotated | 6 examples |
| TOTAL | 1,251 | Complete documentation | - |

---

## How to Use These Documents

### For Command Authors
1. Read **COMMAND_DOCUMENTATION_STANDARDS.md** section-by-section
2. Refer to **COMMAND_EXAMPLES.md** for real code patterns
3. Use **validation checklist** before publishing

### For Code Reviewers
1. Check YAML frontmatter against standards
2. Verify 70/30 narrative/code balance
3. Confirm all AskUserQuestion blocks have response processing
4. Ensure Phase A/B distinction is clear

### For Training/Onboarding
1. Start with **ANALYSIS_SUMMARY.md** overview
2. Show **COMMAND_EXAMPLES.md** for practical patterns
3. Review **COMMAND_DOCUMENTATION_STANDARDS.md** for details

### For Quick Reference
- Bookmark "Key Findings Summary" section (this document)
- Keep "Best Practices" checklist handy
- Refer to COMMAND_EXAMPLES.md while writing

---

## Investigation Methodology

### Sources Examined
1. ✅ 4 existing command files:
   - 0-project.md (2,202 lines)
   - 1-plan.md (880 lines)
   - 2-run.md (793 lines)
   - 3-sync.md (1,085 lines)

2. ✅ Package template structure
   - src/moai_adk/templates/.claude/commands/

3. ✅ Project guidelines
   - CLAUDE.md (local, 1,000+ lines)
   - .claude/CLAUDE.md (global, 500+ lines)

### Analysis Depth
- Line-by-line examination of command files
- Pattern identification across all 4 commands
- Consistent structure validation
- Real example extraction and annotation
- Standards derivation from observed patterns

---

## Standards Validation Checklist

Before publishing a new command, verify:

- [ ] YAML frontmatter complete (name, description, argument-hint, allowed-tools)
- [ ] First 100 lines explain purpose clearly
- [ ] 3+ usage scenarios shown
- [ ] Phase A marked OPTIONAL with skip conditions
- [ ] Phase B marked REQUIRED
- [ ] All AskUserQuestion blocks include:
  - [ ] Full code block
  - [ ] Response processing with `answers["0"] === "Label"`
  - [ ] Action mapping for each response
- [ ] All Task tool calls complete with subagent_type, description, prompt
- [ ] Variables preserved: {{CONVERSATION_LANGUAGE}}, $ARGUMENTS, etc.
- [ ] Tables use consistent columns
- [ ] ASCII diagrams under 10 lines
- [ ] Sections under 200 lines
- [ ] Emoji consistent throughout
- [ ] Links to Skills use `Skill("skill-name")` format
- [ ] Code blocks copy-paste ready (no pseudo-code)
- [ ] 70/30 narrative/code balance maintained
- [ ] Final section ends with AskUserQuestion for next action

---

## Version History

Created: 2025-11-04
Investigation Type: Comprehensive standards analysis
Scope: MoAI-ADK official command documentation format
Coverage: 4 commands, 4,960 total lines analyzed
Standards Derived: 15-section structure, 70/30 balance, Phase A/B pattern, batching pattern

