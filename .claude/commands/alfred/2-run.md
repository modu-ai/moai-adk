---
name: alfred:2-run
description: "Execute TDD implementation cycle"
argument-hint: "SPEC-ID - All with SPEC ID to implement (e.g. SPEC-001) or all \"SPEC Implementation\""
allowed-tools:
  - Read
  - Write
  - Edit
  - MultiEdit
  - Bash(python3:*)
  - Bash(pytest:*)
  - Bash(npm:*)
  - Bash(node:*)
  - Bash(git:*)
  - Task
  - WebFetch
  - Grep
  - Glob
  - TodoWrite
---

# ⚒️ MoAI-ADK Step 2: Execute Implementation (Run) - TDD Implementation

> **Critical Note**: ALWAYS invoke `Skill("moai-alfred-ask-user-questions")` before using `AskUserQuestion` tool. This skill provides up-to-date best practices, field specifications, and validation rules for interactive prompts.
>
> **Batched Design**: All AskUserQuestion calls follow batched design principles (1-4 questions per call) to minimize user interaction turns. See CLAUDE.md section "Alfred Command Completion Pattern" for details.

<!-- @CODE:ALF-WORKFLOW-002:CMD-RUN -->

**4-Step Workflow Integration**: This command implements Step 3 of Alfred's workflow (Task Execution with TodoWrite tracking). See CLAUDE.md for full workflow details.

---

## 🎯 Command Purpose

Execute planned tasks based on SPEC document analysis. Supports TDD implementation, prototyping, and documentation work.

**Run on**: $ARGUMENTS

## 💡 Execution Philosophy: "Plan → Run → Sync"

`/alfred:2-run` performs planned tasks through various execution strategies.

### 3 Main Scenarios

#### Scenario 1: TDD Implementation (main method) ⭐
```bash
/alfred:2-run SPEC-AUTH-001
→ RED → GREEN → REFACTOR
→ Implement high-quality code through test-driven development
```

#### Scenario 2: Prototyping
```bash
/alfred:2-run SPEC-PROTO-001
→ Prototype implementation for quick verification
→ Quick feedback with minimal testing
```

#### Scenario 3: Documentation Tasks
```bash
/alfred:2-run SPEC-DOCS-001
→ Writing documentation and generating sample code
→ API documentation, tutorials, guides, etc.
```

## 📋 Execution Flow

1. **SPEC Analysis**: Requirements extraction and complexity assessment
2. **Implementation Strategy**: Determine optimized approach (TDD, prototype, documentation)
3. **User Confirmation**: Review and approve action plan
4. **Execute Task**: Perform work according to approved plan
5. **Git Operations**: Create step-by-step commits with git-manager

---

## 🧠 Associated Skills & Agents

| Agent                  | Core Skill                       | Purpose                                 |
| ---------------------- | -------------------------------- | --------------------------------------- |
| implementation-planner | `moai-alfred-language-detection` | Detect language and design architecture |
| tdd-implementer        | `moai-essentials-debug`          | Implement TDD (RED → GREEN → REFACTOR)  |
| quality-gate           | `moai-alfred-trust-validation`   | Verify TRUST 5 principles               |
| git-manager            | `moai-alfred-git-workflow`       | Commit and manage Git workflows         |

**Note**: TUI Survey Skill is used for user confirmations during the run phase and is shared across all interactive prompts.

---

## 🚀 PHASE 1: Analysis & Planning

**Goal**: Analyze SPEC requirements and create execution plan.

### Step 1.1: Load Skills & Prepare Context

1. **Load TUI Skill immediately**:
   - Invoke: `Skill("moai-alfred-ask-user-questions")`
   - This enables interactive menus for all user interactions

2. **Read SPEC document**:
   - Read: `.moai/specs/SPEC-$ARGUMENTS/spec.md`
   - Determine if codebase exploration is needed (existing patterns, similar implementations)

3. **Optionally invoke Explore agent for codebase analysis**:
   - IF SPEC requires understanding existing code patterns:
     - Use Task tool with `subagent_type: "Explore"`
     - Prompt: "Analyze codebase for SPEC-$ARGUMENTS: Similar implementations, test patterns, architecture, libraries/versions"
     - Thoroughness: "medium"
   - ELSE: Skip and proceed directly to Step 1.2

**Result**: SPEC context gathered. Ready for planning.

---

### Step 1.2: Invoke Implementation-Planner Agent

**Your task**: Call implementation-planner to analyze SPEC and create execution strategy.

Use Task tool:
- `subagent_type`: "implementation-planner"
- `description`: "SPEC analysis and execution strategy establishment"
- `prompt`:
  ```
  You are the implementation-planner agent.

  **Task**: Analyze SPEC and create execution plan.

  SPEC ID: $ARGUMENTS
  Language: [from .moai/config.json]

  **Analyze**:
  1. Requirements extraction and complexity assessment
  2. Library selection (use WebFetch for latest stable versions)
  3. TAG chain design
  4. Step-by-step execution plan
  5. Risk factors and mitigation strategies

  **Consider**: Exploration results if provided

  **Output**: Execution plan report with:
  - Complexity (Low/Medium/High)
  - Estimated work time
  - Selected language and approach
  - Latest library versions
  - Risk factors
  - Quality gate targets
  ```

**Store**: Response in `$EXECUTION_PLAN`

---

### Step 1.3: Request User Approval

Present plan to user:

1. **Display plan report**:
   ```
   ═══════════════════════════════════════════════════════
   📋 Execution Plan: SPEC-$ARGUMENTS
   ═══════════════════════════════════════════════════════

   📊 Analysis Results:
   - Complexity: [Low/Medium/High]
   - Estimated Time: [Time]
   - Key Challenges: [List]

   🎯 Strategy:
   - Language: [Language]
   - Approach: [Approach]
   - Core modules: [List]

   📦 Dependencies:
   - [Package version list]

   ⚠️ Risks:
   - [Risk list]

   ═══════════════════════════════════════════════════════
   ```

2. **Ask for user approval using AskUserQuestion**:
   - `question`: "Implementation plan is ready. How would you like to proceed?"
   - `header`: "Plan Approval"
   - `multiSelect`: false
   - `options`: 4 choices:
     1. "✅ Proceed with TDD" → Start implementation
     2. "🔍 Research First" → Deep dive into codebase
     3. "🔄 Request Modifications" → Change strategy
     4. "⏸️ Postpone" → Save plan for later

3. **Process user response**:
   - IF "Proceed" → Go to PHASE 2
   - IF "Research First" → Re-run Explore agent, update plan, re-ask approval
   - IF "Modifications" → Ask for changes, update plan, re-ask approval
   - IF "Postpone" → Save plan to `.moai/specs/SPEC-$ARGUMENTS/plan.md`, create commit, exit

**Result**: User decision captured. Command proceeds or exits.

---

## 🔧 PHASE 2: Execute Task (TDD Implementation)

**Goal**: Execute approved implementation plan with TDD cycle.

### Step 2.1: Initialize Progress Tracking

Use TodoWrite to track all tasks:

1. **Parse tasks from execution plan**:
   - Extract all TAG IDs and descriptions
   - Create TodoWrite entry for each task

2. **Initialize TodoWrite**:
   - Set all tasks to "pending"
   - Ready for status updates during execution

---

### Step 2.2: Check Domain Readiness (Optional)

For multi-domain SPECs:

1. **Read SPEC metadata** for `domains:` field
2. **For each domain**, invoke Explore agent for readiness check:
   - Domain examples: frontend, backend, devops, database, data-science, mobile
   - Prompt: "Brief readiness check for [domain] implementation of SPEC-$ARGUMENTS (3-4 key points)"
3. **Store feedback** in memory for tdd-implementer

---

### Step 2.3: Invoke TDD-Implementer Agent

**Your task**: Call tdd-implementer to execute the approved plan with TDD cycle.

Use Task tool:
- `subagent_type`: "tdd-implementer"
- `description`: "Execute TDD implementation cycle"
- `prompt`:
  ```
  You are the tdd-implementer agent.

  Language settings:
  - conversation_language: [from config]
  - Code must be in English
  - Code comments: per project language rules

  **Execute Approved Plan**:
  - SPEC ID: $ARGUMENTS
  - Execution plan: [from implementation-planner]
  - Domain expertise: [if available from Step 2.2]

  **TDD Cycle for each TAG**:
  1. RED: Write failing test (@TEST:TAG)
  2. GREEN: Minimal implementation (@CODE:TAG)
  3. REFACTOR: Code quality improvement

  **Skills to use**:
  - Skill("moai-alfred-language-detection") - Language detection
  - Skill("moai-essentials-debug") - Debugging if errors occur

  **Output**: Implementation completion report with TAG status
  ```

**Store**: Response in `$IMPLEMENTATION_RESULTS`

---

### Step 2.4: Invoke Quality-Gate Agent

After tdd-implementer completes, call quality-gate for TRUST 5 verification:

Use Task tool:
- `subagent_type`: "quality-gate"
- `description`: "TRUST principle verification"
- `prompt`:
  ```
  You are the quality-gate agent.

  **Verify TRUST 5 principles**:
  1. Test First: Coverage ≥ 85% (from .moai/config.json)
  2. Readable: File ≤ 300 LOC, function ≤ 50 LOC
  3. Unified: Consistent architecture and patterns
  4. Secured: No exposed credentials
  5. Trackable: Complete TAG chain

  **Also verify**:
  - Code style (linter/formatter)
  - No critical issues

  **Output**: PASS / WARNING / CRITICAL with details
  ```

**Handle result**:
- IF PASS → Proceed to PHASE 3
- IF WARNING → Ask user: "Accept warnings?" or "Fix first?"
- IF CRITICAL → Block progress, report details, wait for fixes

---

## 🚀 PHASE 3: Git Operations

**Goal**: Create Git commits for all completed work.

### Step 3.1: Invoke Git-Manager Agent

**Your task**: Call git-manager to create structured commits.

Use Task tool:
- `subagent_type`: "git-manager"
- `description`: "Create Git commits for TDD cycle"
- `prompt`:
  ```
  You are the git-manager agent.

  **Create commits**:
  - SPEC ID: $ARGUMENTS
  - Completed tasks: [from TodoWrite]
  - TDD phases: RED → GREEN → REFACTOR

  **Commit structure**:
  - RED: test(SPEC-{ID}): Add failing tests
  - GREEN: feat(SPEC-{ID}): Implement feature
  - REFACTOR: refactor(SPEC-{ID}): Improve code quality

  **Git strategy**: Use GitFlow if team mode (feature → develop)

  **Output**: Commit summary
  ```

**Verify**: Commits were created successfully

---

### Step 3.2: Verify and Complete

1. **Execute**: `git log -1 --oneline`
2. **Display** commit summary to user
3. **Next guidance**: "Commits created on feature branch. Run `/alfred:3-sync` to create PR."

---

## 🎯 PHASE 4: Next Steps

**Goal**: Guide user to next action.

### Step 4.1: Ask for Next Action

Use AskUserQuestion:
- `question`: "Implementation is complete. What would you like to do next?"
- `header`: "Next Steps"
- `multiSelect`: false
- `options`: 4 choices:
  - "📄 Synchronize Documentation" → `/alfred:3-sync auto`
  - "🚀 Implement More Features" → `/alfred:2-run SPEC-XXX`
  - "🔄 Start New Session" → `/clear` (recommended)
  - "✅ Complete" → End workflow

### Step 4.2: Display Summary

```
═══════════════════════════════════════════════════════
✅ Implementation Complete
═══════════════════════════════════════════════════════

SPEC: SPEC-$ARGUMENTS
TAGs: [count] completed
Commits: [count] created
Quality: [PASS/WARNING/CRITICAL]

Next: [Based on user selection]
═══════════════════════════════════════════════════════
```

---

## 📚 Quick Reference

**For implementation details, consult**:
- `Skill("moai-alfred-language-detection")` - Language-specific TDD tools
- `Skill("moai-essentials-debug")` - Debugging strategies
- `Skill("moai-alfred-trust-validation")` - TRUST 5 principles
- CLAUDE.md - Full workflow documentation

**Quality Gate Checklist**:
- ✅ Test coverage ≥ 85%
- ✅ Code style compliance
- ✅ TAG chain completeness
- ✅ No security vulnerabilities
- ✅ TRUST 5 principles met

**Version**: 2.1.0 (Agent-Delegated Pattern)
**Last Updated**: 2025-11-09
**Total Lines**: ~400 (reduced from 619)
**Architecture**: Commands → Agents → Skills
