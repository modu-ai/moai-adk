---
title: "Directive Specifications Index & Quick Reference"
version: "1.0.0"
date: "2025-11-19"
audience: "Everyone (architects, developers, QA, leadership)"
scope: "Navigation and quick reference for all directive specifications"
---

# Directive Specifications: Quick Reference & Index

**Complete set of specifications for embedding directives in official Claude Code files.**

This guide helps you find the right specification for your role and task.

---

## What Changed?

### Before: Separate Directive Files
```
❌ .moai/directives/ (disconnected from code)
   ├── 0-project-command-directive.md
   ├── 0-project-error-recovery-guide.md
   └── 0-project-executive-summary.md

❌ .claude/commands/moai/0-project.md (no directives)
❌ .claude/agents/moai/project-manager.md (incomplete directives)
```

### After: Embedded Directives
```
✅ .claude/commands/moai/0-project.md (CONTAINS all command directives)
✅ .claude/agents/moai/project-manager.md (CONTAINS all agent directives)
✅ .claude/skills/moai-*/SKILL.md (CONTAINS skill directives)

Optional Archive:
📦 .moai/directives/ (reference only - not active)
```

---

## Four Specification Documents Delivered

### 1️⃣ Architecture Specification
**File**: `DIRECTIVE-ARCHITECTURE-SPECIFICATION.md`

**Purpose**: Understand HOW directives should be embedded (the architecture)

**Size**: 2,800 lines

**Read This If You**:
- Need to understand the overall architecture
- Are designing new Claude Code files
- Want to know why directives are embedded, not separate
- Need to implement migration from separate files to embedded

**Key Sections**:
- Problem & Solution
- Core Architecture Principle
- Layer 1-3 Structure (command/agent/skill)
- Embedding Standards (how to write)
- Migration Path
- Validation Checklist

**Time to Read**: 45 minutes

---

### 2️⃣ Command Directive Specification
**File**: `COMMAND-DIRECTIVE-SPECIFICATION.md`

**Purpose**: Detailed specification of WHAT goes in `.claude/commands/moai/0-project.md`

**Size**: 3,200 lines

**Read This If You**:
- Are implementing the command file
- Need to understand entry point logic
- Want to know what the command should/shouldn't do
- Need to understand mode detection and routing
- Want to see error handling patterns

**Key Sections**:
- YAML Frontmatter (what's required)
- User Directives (overview sections)
- Entry Point Directives (parsing, mode detection)
- Tool Usage Constraints (Task + AskUserQuestion only)
- Error Handling at Entry Point
- Agent Delegation
- Skills Reference

**Key Content**:
- Decision tree for argument parsing
- All error types at entry point + recovery
- Exact Task() call format to agent
- Expected response formats from agent
- Complete skills reference

**Time to Read**: 60 minutes (40 for overview, 20 for deep dive)

---

### 3️⃣ Agent Directive Specification
**File**: `AGENT-DIRECTIVE-SPECIFICATION.md`

**Purpose**: Detailed specification of WHAT goes in `.claude/agents/moai/project-manager.md`

**Size**: 4,500 lines

**Read This If You**:
- Are implementing the agent file
- Need to understand mode-specific workflows
- Want to see how to handle the 4 modes
- Need language handling directives
- Want to understand skill integration
- Need error recovery patterns

**Key Sections**:
- Agent Responsibility (overall role)
- INITIALIZATION Mode (11 steps, 5 phases)
- AUTO-DETECT Mode
- SETTINGS Mode (5 tabs, 6 phases)
- UPDATE Mode (6 phases)
- Language Handling Directives
- Error Recovery Directives
- State Management Directives
- User Interaction Directives

**Key Content**:
- Exact numbered steps for INITIALIZATION (11 steps)
- All 4 modes with complete workflows
- Tab-based SETTINGS structure
- Language change handling
- Skill delegation patterns
- Validation checkpoints and rules
- Error recovery patterns
- AskUserQuestion() structures

**Time to Read**: 90 minutes (60 for overview, 30 for deep dive)

---

### 4️⃣ Specification Delivery Summary
**File**: `SPECIFICATION-DELIVERY.md`

**Purpose**: Executive summary of what was delivered

**Size**: 1,500 lines

**Read This If You**:
- Are deciding whether to approve these specifications
- Need a high-level overview before diving in
- Want to understand the problem and solution
- Need to know how to use the specifications
- Want implementation timeline

**Key Sections**:
- Problem Solved (before vs. after)
- Deliverables Overview
- How Specifications Work Together
- Key Highlights
- How to Use (5-phase implementation guide)
- Success Criteria

**Time to Read**: 20 minutes

---

## Quick Navigation Guide

### By Role

**👔 Project Leadership**
1. Read: `SPECIFICATION-DELIVERY.md` (20 min)
2. Review: Problem/Solution section
3. Decision: Approve for implementation
4. Monitor: 5-phase implementation timeline

**🏗️ Software Architect**
1. Read: `DIRECTIVE-ARCHITECTURE-SPECIFICATION.md` (45 min)
2. Review: All architecture principles
3. Understand: Command ↔ Agent ↔ Skill layers
4. Validate: Against existing Claude Code patterns

**⚙️ Command Implementer**
1. Read: `DIRECTIVE-ARCHITECTURE-SPECIFICATION.md` (20 min, focus on overview)
2. Read: `COMMAND-DIRECTIVE-SPECIFICATION.md` (60 min, full detail)
3. Implement: Following specification exactly
4. Test: Against success criteria

**🤖 Agent Implementer**
1. Read: `DIRECTIVE-ARCHITECTURE-SPECIFICATION.md` (20 min, focus on overview)
2. Read: `AGENT-DIRECTIVE-SPECIFICATION.md` (90 min, full detail)
3. Implement: Following specification exactly
4. Test: Against success criteria

**🧪 QA Engineer**
1. Read: `SPECIFICATION-DELIVERY.md` (20 min)
2. Read: Error handling sections (from other specs)
3. Plan: Test cases based on directives
4. Test: Against specifications

**📚 Documentation**
1. Read: `SPECIFICATION-DELIVERY.md` (20 min)
2. Skim: Other specs for terminology
3. Write: User guides explaining directives
4. Reference: Specs as authoritative source

---

### By Task

**Understanding the overall design**
→ Read: `DIRECTIVE-ARCHITECTURE-SPECIFICATION.md`

**Implementing the command file**
→ Read: `COMMAND-DIRECTIVE-SPECIFICATION.md`

**Implementing the agent file**
→ Read: `AGENT-DIRECTIVE-SPECIFICATION.md`

**Getting executive approval**
→ Read: `SPECIFICATION-DELIVERY.md`

**Planning implementation timeline**
→ Read: `SPECIFICATION-DELIVERY.md` → Phase-based guide

**Testing the implementation**
→ Read: Success Criteria in relevant spec

**Migrating from old approach**
→ Read: DIRECTIVE-ARCHITECTURE-SPECIFICATION.md → Migration Path section

**Understanding error handling**
→ Read: COMMAND spec (entry errors) + AGENT spec (mode errors)

**Understanding language handling**
→ Read: AGENT-DIRECTIVE-SPECIFICATION.md → Language Handling Directives section

**Understanding skill integration**
→ Read: COMMAND spec (which skills) + AGENT spec (how to call + what to expect)

---

## Key Concepts Summary

### 1. Official Files ARE Directives

**Principle**: The `.claude/` files don't HAVE directives - they ARE directives

```
.claude/commands/moai/0-project.md
  = Specification of what command should do
  = Specification of what command shouldn't do
  = Specification of tool constraints
  = Specification of entry-point error handling
```

**Consequence**: If implementation differs from directive, the directive wins.

---

### 2. Layered Architecture

```
User Input
    ↓
Command (.claude/commands/moai/0-project.md)
  ├─ Parses arguments
  ├─ Detects mode
  ├─ Loads language
  └─ Delegates to agent
    ↓
Agent (.claude/agents/moai/project-manager.md)
  ├─ Executes mode workflow
  ├─ Asks interview questions
  └─ Delegates to skills
    ↓
Skills (.claude/skills/moai-*/SKILL.md)
  ├─ File operations
  ├─ Validations
  └─ Transformations
    ↓
User Output (in user's language)
```

Each layer has clear directives for what it does/doesn't do.

---

### 3. Tool Constraints

**Command** (lightweight orchestration):
- ✅ Task() - delegate to agent
- ✅ AskUserQuestion() - entry-level decisions only
- ❌ No Read/Write/Edit/Bash

**Agent** (complex logic):
- ✅ Skill() - call skills
- ✅ AskUserQuestion() - user interviews
- ✅ Bash - safe operations (if needed)
- ❌ No direct file operations (delegate to skills)

**Skills** (execution):
- ✅ Read/Write/Edit - file operations
- ✅ Bash - all needed operations
- ✅ Any tools needed for task

---

### 4. Language is FIRST

**Critical Principle**: Language is confirmed BEFORE any other action

```
INITIALIZATION Mode, Step 1:
  "Is your conversation language correct?"
  ↓
  User confirms language
  ↓
  ALL subsequent output in that language
```

Why? Because language affects everything:
- All user-facing messages
- All interview questions
- All documentation generation
- All error messages

---

### 5. Four Modes

**Mode**: Operational context (what the user wants)

| Mode | Trigger | Purpose |
|------|---------|---------|
| INITIALIZATION | `/moai:0-project` + no config | First-time setup |
| AUTO-DETECT | `/moai:0-project` + config exists | Review/modify existing |
| SETTINGS | `/moai:0-project setting` | Change configuration |
| UPDATE | `/moai:0-project update` | Apply package updates |

Each mode has distinct workflow with 3-6 phases.

---

### 6. Five Skills

**Skill**: Reusable capability for specific operations

| Skill | Mode | Purpose |
|-------|------|---------|
| moai-project-language-initializer | INIT, SETTINGS | Language selection |
| moai-project-config-manager | All | Config file operations |
| moai-project-batch-questions | SETTINGS | Interview framework |
| moai-project-template-optimizer | UPDATE | Template merging |
| moai-project-documentation | INIT, AUTO-DETECT, UPDATE | Doc generation |

Agent calls skills when needed. Command never calls skills directly.

---

## Reading Checklists

### Before Implementation Starts

- [ ] Leadership reads `SPECIFICATION-DELIVERY.md`
- [ ] Architect reads `DIRECTIVE-ARCHITECTURE-SPECIFICATION.md`
- [ ] Command dev reads `COMMAND-DIRECTIVE-SPECIFICATION.md`
- [ ] Agent dev reads `AGENT-DIRECTIVE-SPECIFICATION.md`
- [ ] Team discusses any questions/clarifications
- [ ] Team approves specifications

### During Implementation

- [ ] Command dev references `COMMAND-DIRECTIVE-SPECIFICATION.md` daily
  - [ ] Entry point directives (when parsing)
  - [ ] Tool constraints (when coding)
  - [ ] Error handling (when error handling)
  - [ ] Agent delegation (when calling Task())

- [ ] Agent dev references `AGENT-DIRECTIVE-SPECIFICATION.md` daily
  - [ ] Mode directives (when implementing mode)
  - [ ] Language directives (when handling language)
  - [ ] Skill directives (when calling skills)
  - [ ] Validation directives (when validating)
  - [ ] Error directives (when handling errors)

### During Code Review

- [ ] Reviewer checks command against `COMMAND-DIRECTIVE-SPECIFICATION.md`
  - [ ] Tool usage correct?
  - [ ] Argument parsing complete?
  - [ ] Error handling matches spec?

- [ ] Reviewer checks agent against `AGENT-DIRECTIVE-SPECIFICATION.md`
  - [ ] All modes implemented?
  - [ ] All phases complete?
  - [ ] Language handling correct?
  - [ ] Skill calls correct?

### During Testing

- [ ] QA references success criteria in each spec
- [ ] QA tests each error recovery path
- [ ] QA tests with multiple languages
- [ ] QA validates specification compliance

---

## Common Questions

### Q: Why embed directives instead of keeping them separate?

**A**:
- Directives get out of sync easily with separate files
- Developers often ignore separate documentation
- Official files ARE the specification (single source of truth)
- Easier to maintain when directives live with code
- Forces directives to be current

See: `DIRECTIVE-ARCHITECTURE-SPECIFICATION.md` → Problem Statement section

---

### Q: What if implementation differs from specification?

**A**: The specification wins.
- Fix the implementation to match the spec
- If spec needs change: update spec + implementation together
- Never ignore specifications to "get it done"
- Specs are part of the code contract

---

### Q: Can I use different tools than specified?

**A**: No.
- Command: ONLY Task() and AskUserQuestion()
- Agent: ONLY Skill() and AskUserQuestion() (+ Read/Write if needed)
- Skills: Any tools needed for task

Why? Clear separation of concerns, security boundaries, testability.

See: `COMMAND-DIRECTIVE-SPECIFICATION.md` → Tool Usage Constraints

---

### Q: What if user speaks a language not in config?

**A**: Default to English, ask user to select.

See: `AGENT-DIRECTIVE-SPECIFICATION.md` → Language Handling Directives

---

### Q: How do I handle errors during skill execution?

**A**:
- Catch skill error
- Show user clear error message in their language
- Offer recovery path (retry/skip/cancel)
- Never silently fail

See: `AGENT-DIRECTIVE-SPECIFICATION.md` → Error Recovery Directives

---

### Q: How many skills can I call?

**A**: Call only the skills needed for your mode. Don't call unnecessary skills.

See: `COMMAND-DIRECTIVE-SPECIFICATION.md` → Skills Reference

---

### Q: Can I change language mid-workflow?

**A**: No. Language is locked after confirmation (Step 1 of INITIALIZATION, or loaded from config).

If user wants different language: Use SETTINGS mode Tab 1.

See: `AGENT-DIRECTIVE-SPECIFICATION.md` → Language Handling Directives

---

### Q: How do I validate user input?

**A**:
- At checkpoints: After key steps
- Three validation points: Tab 1 (language), Tab 3 (git conflicts), before write

See: `AGENT-DIRECTIVE-SPECIFICATION.md` → Validation Directives (various modes)

---

## Implementation Timeline (High Level)

| Phase | Duration | What Happens |
|-------|----------|---|
| Review | 1 week | Team reads specs, clarifies questions |
| Planning | 1 week | Detailed task list, effort estimates |
| Command | 1 week | Implement entry point, mode detection |
| Agent | 2 weeks | Implement all 4 modes with phases |
| Testing | 1 week | QA validates against specifications |
| **Total** | **6 weeks** | Ready for release |

See: `SPECIFICATION-DELIVERY.md` → Phase-based guide

---

## Success Criteria

Implementation is complete when:

✅ Command implementation matches `COMMAND-DIRECTIVE-SPECIFICATION.md` exactly
✅ Agent implementation matches `AGENT-DIRECTIVE-SPECIFICATION.md` exactly
✅ Skill implementations include directives per `DIRECTIVE-ARCHITECTURE-SPECIFICATION.md`
✅ All tests pass
✅ User experience matches specifications
✅ Error recovery works for all error paths
✅ Language handling correct for multiple languages
✅ Specification compliance verified
✅ No separate directive files active (all embedded)

---

## File Manifest

All specifications in: `/Users/goos/MoAI/MoAI-ADK/.moai/directives/`

| File | Lines | Purpose |
|------|-------|---------|
| DIRECTIVE-ARCHITECTURE-SPECIFICATION.md | 2,800 | How directives should be embedded |
| COMMAND-DIRECTIVE-SPECIFICATION.md | 3,200 | What goes in command file |
| AGENT-DIRECTIVE-SPECIFICATION.md | 4,500 | What goes in agent file |
| SPECIFICATION-DELIVERY.md | 1,500 | Executive summary + implementation guide |
| README-SPECIFICATIONS.md | This file | Quick reference + navigation |
| **TOTAL** | **12,000** | **Complete specification package** |

---

## Getting Started

### Right Now (Next 5 minutes)
1. ✅ You're reading this guide
2. Understand the 4 documents
3. Identify which one you need

### Next (Next hour)
1. Read the appropriate spec for your role
2. Make a list of questions
3. Share questions with team

### This Week
1. Clarify any specification questions
2. Get team approval
3. Create detailed implementation plan

### Next Week
1. Start implementation
2. Reference specifications daily
3. Code review against specifications

---

## Questions or Clarifications?

Each specification includes:
- Introduction explaining scope and audience
- Detailed sections on specific topics
- Examples and code snippets
- Implementation checklists
- Success criteria

If you can't find answer:
1. Check the index of relevant specification
2. Search for topic across all specs
3. Ask someone who reads that specification
4. File clarification issue for architecture lead

---

## Quick Links Within Specs

**For Entry Point Logic**:
→ `COMMAND-DIRECTIVE-SPECIFICATION.md` → Section 2

**For Mode Execution**:
→ `AGENT-DIRECTIVE-SPECIFICATION.md` → Mode sections

**For Tool Constraints**:
→ `COMMAND-DIRECTIVE-SPECIFICATION.md` → Section 3

**For Language Handling**:
→ `AGENT-DIRECTIVE-SPECIFICATION.md` → Language Directives section

**For Error Recovery**:
→ `COMMAND-DIRECTIVE-SPECIFICATION.md` → Section 4
→ `AGENT-DIRECTIVE-SPECIFICATION.md` → Error Recovery section

**For Skill Integration**:
→ `COMMAND-DIRECTIVE-SPECIFICATION.md` → Section 6
→ `AGENT-DIRECTIVE-SPECIFICATION.md` → Each mode's Skill Delegation section

**For Implementation Timeline**:
→ `SPECIFICATION-DELIVERY.md` → How to Use section

---

## Version & Maintenance

**Version**: 1.0.0 (Initial Release)
**Released**: 2025-11-19
**Status**: APPROVED - Ready for Implementation

**Updates**:
- If specifications need clarification: Document in issue
- If specifications need revision: Update all related specs
- If implementation reveals issues: Update specs first, then code

**Maintenance**: Architecture lead

---

## Next Phase

After these specifications are approved:

→ **Create detailed implementation issues**
→ **Assign to command + agent developers**
→ **Begin implementation following phased timeline**
→ **QA tests against specifications**
→ **Code review validates specification compliance**
→ **Deploy when specification compliance 100%**

---

**Navigation Guide Version**: 1.0.0
**Last Updated**: 2025-11-19
**Status**: COMPLETE - Ready for Use

For detailed information, select the appropriate specification document from the table above.
