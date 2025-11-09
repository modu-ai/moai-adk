---
name: alfred:3-sync
description: "Synchronize documentation and finalize PR"
argument-hint: 'Mode target path - Mode: auto (default)|force|status|project, target path: Synchronization target path'
allowed-tools:
- Read
- Write
- Edit
- MultiEdit
- Bash(git:*)
- Bash(gh:*)
- Bash(python3:*)
- Task
- Grep
- Glob
- TodoWrite
---

# 📚 MoAI-ADK Step 3: Document Synchronization (+Optional PR Ready)

> **Note**: Interactive prompts use `AskUserQuestion tool (documented in moai-alfred-ask-user-questions skill)` for TUI selection menus. The skill is loaded on-demand when user interaction is required.
>
> **Batched Design**: All AskUserQuestion calls follow batched design principles (1-4 questions per call) to minimize user interaction turns. See CLAUDE.md section "Alfred Command Completion Pattern" for details.

<!-- @CODE:ALF-WORKFLOW-003:CMD-SYNC -->

**4-Step Workflow Integration**: This command implements Step 4 of Alfred's workflow (Report & Commit with conditional report generation). See CLAUDE.md for full workflow details.

---

## 🎯 Command Purpose

Synchronize code changes to Living Documents and verify @TAG system to ensure complete traceability.

**Document sync to**: $ARGUMENTS

> **Standard workflow**: STEP 1 (Analysis & Planning) → User Approval → STEP 2 (Document Sync) → STEP 3 (Git Commit & PR)

---

## 📋 Execution Modes

This command supports **4 operational modes**:

| Mode | Scope | PR Processing | Use Case |
|------|-------|---------------|----------|
| **auto** (default) | Smart selective sync | PR Ready conversion | Daily development workflow |
| **force** | Full project re-sync | Full regeneration | Error recovery, major refactoring |
| **status** | Status check only | Report only | Quick health check |
| **project** | Integrated project-wide | Project-level updates | Milestone completion, periodic sync |

**Command usage examples**:
- `/alfred:3-sync` → Auto-sync (PR Ready only)
- `/alfred:3-sync --auto-merge` → PR auto-merge + branch cleanup
- `/alfred:3-sync force` → Force full synchronization
- `/alfred:3-sync status` → Check synchronization status
- `/alfred:3-sync project` → Integrated project synchronization
- `/alfred:3-sync auto src/auth/` → Specific path synchronization

---

## 🧠 Associated Skills & Agents

| Agent | Core Skill | Purpose |
| ------------ | ------------------------------ | ------------------------------ |
| tag-agent | `moai-alfred-tag-scanning` | Verify TAG system integrity |
| quality-gate | `moai-alfred-trust-validation` | Check code quality before sync |
| doc-syncer | `moai-alfred-tag-scanning` | Synchronize Living Documents |
| git-manager | `moai-alfred-git-workflow` | Handle Git operations |

**Note**: TUI Survey Skill is loaded once at Phase 0 and reused throughout all user interactions.

---

## 🚀 OVERALL WORKFLOW STRUCTURE

```
┌──────────────────────────────────────────────────────────┐
│ PHASE 1: Analysis & Planning (tag-agent + doc-syncer)   │
│  - Verify prerequisites                                  │
│  - Analyze project status (Git + TAG)                    │
│  - Request user approval                                 │
└──────────────────────────────────────────────────────────┘
                          ↓
          ┌───────────────┴───────────────┐
          │                               │
     User approves                   User aborts
          │                               │
          ↓                               ↓
┌─────────────────────────┐   ┌──────────────────────┐
│ PHASE 2: Execute Sync   │   │ PHASE 4: Graceful    │
│ (doc-syncer + quality)  │   │ Exit (no changes)    │
│  - Create backup        │   └──────────────────────┘
│  - Sync documents       │
│  - Verify TAGs          │
└─────────────────────────┘
          ↓
┌──────────────────────────────────────────────────────────┐
│ PHASE 3: Git Operations & PR (git-manager)               │
│  - Commit document changes                               │
│  - Transition PR (Team mode)                             │
│  - Auto-merge (if requested)                             │
│  - Branch cleanup                                        │
│  - Next steps guidance                                   │
└──────────────────────────────────────────────────────────┘
```

---

## 🔧 PHASE 1: Analysis & Planning

**Goal**: Gather project context, verify TAG system, and get user approval.

### Step 1.1: Verify Prerequisites & Load Skills

Execute these verification steps:

1. **Load the TUI Skill immediately**:
   - Invoke: `Skill("moai-alfred-ask-user-questions")`
   - This enables interactive menus for all user interactions

2. **Verify MoAI-ADK structure**:
   - Check: `.moai/` directory exists
   - Check: `.claude/` directory exists
   - IF missing → Print error and exit

3. **Verify Git repository**:
   - Execute: `git rev-parse --is-inside-work-tree`
   - IF not a Git repo → Print error and exit

4. **Verify Python environment** (optional, non-fatal):
   - Execute: `which python3`
   - IF not found → Print warning but continue

**Result**: Prerequisites verified. TUI system ready.

---

### Step 1.2: Analyze Project Status

Gather context for synchronization planning:

1. **Analyze Git changes**:
   - Execute: `git status --porcelain`
   - Execute: `git diff --name-only HEAD`
   - Count: Python files, test files, documents, SPEC files

2. **Read project configuration**:
   - Read: `.moai/config.json`
   - Extract: `git_strategy.mode` (Personal/Team)
   - Extract: `language.conversation_language` (for document updates)
   - Extract: `git_strategy.spec_git_workflow`

3. **Determine synchronization mode**:
   - Parse $ARGUMENTS for mode: `auto`, `force`, `status`, `project`
   - IF empty → Default to `auto`
   - Parse flags: `--auto-merge`, `--skip-pre-check`, `--skip-quality-check`

4. **Handle status mode early exit**:
   - IF mode is `status` → Execute quick check only:
     - Print current TAG health
     - Print changed files count
     - Print recommendation
     - EXIT command (no further processing)

**Result**: Project status analyzed and mode determined.

---

### Step 1.3: Invoke Tag-Agent for TAG Verification

**Your task**: Call tag-agent to verify TAG system integrity across entire project.

Use Task tool:
- `subagent_type`: "tag-agent"
- `description`: "Verify TAG system across entire project"
- `prompt`:
  ```
  Please perform comprehensive @TAG system verification across the entire project.

  **Required Scope**: Scan ALL source files, not just changed files.

  **Verification Items**:
  1. @SPEC TAGs in .moai/specs/ directory
  2. @TEST TAGs in tests/ directory
  3. @CODE TAGs in src/ directory
  4. @DOC TAGs in docs/ directory

  **Orphan Detection** (Required):
  - Detect @CODE TAGs with no matching @SPEC
  - Detect @SPEC TAGs with no matching @CODE
  - Detect @TEST TAGs with no matching @SPEC
  - Detect @DOC TAGs with no matching @SPEC/@CODE

  **Output Format**:
  - Provide complete list of orphan TAGs with locations
  - TAG chain integrity assessment (Healthy / Issues Detected)
  ```

**Store**: Response in `$TAG_VALIDATION_RESULTS`

---

### Step 1.4: Invoke Doc-Syncer for Synchronization Plan

**Your task**: Call doc-syncer to analyze Git changes and create synchronization strategy.

Use Task tool:
- `subagent_type`: "doc-syncer"
- `description`: "Establish a document synchronization plan"
- `prompt`:
  ```
  You are the doc-syncer agent.

  Language settings:
  - conversation_language: [from .moai/config.json]

  Task: Analyze Git changes and create a synchronization plan.

  Synchronization mode: [auto/force/status/project]
  Changed files: [from git diff]

  TAG verification results: [from tag-agent]

  Required output:
  1. Summary of documents to update
  2. SPEC documents requiring synchronization
  3. TAG repairs needed
  4. Estimated work scope

  Ensure all document updates align with conversation_language setting.
  ```

**Store**: Response in `$SYNC_PLAN`

---

### Step 1.5: Request User Approval

Present synchronization plan and get user decision:

1. **Display comprehensive plan report**:
   ```
   ═══════════════════════════════════════════════════════
   📚 Document Synchronization Plan Report
   ═══════════════════════════════════════════════════════

   📊 Project Analysis:
   - Mode: [mode]
   - Scope: [scope]
   - Changed files: [count]
   - Project mode: [Personal/Team]

   🎯 Synchronization Strategy:
   - Living Documents: [list]
   - SPEC documents: [list]
   - TAG repairs needed: [count]

   ⚠️ TAG System Status:
   - TAG chain integrity: [Healthy / Issues]
   - Orphan TAGs: [count]
   - Broken references: [count]

   ═══════════════════════════════════════════════════════
   ```

2. **Ask for user approval using AskUserQuestion**:
   - `question`: "Synchronization plan is ready. How would you like to proceed?"
   - `header`: "Plan Approval"
   - `multiSelect`: false
   - `options`: 4 choices:
     1. "✅ Proceed with Sync" → Execute synchronization
     2. "🔄 Request Modifications" → Modify strategy
     3. "🔍 Review Details" → See full TAG results
     4. "❌ Abort" → Cancel (no changes made)

3. **Process user response**:
   - IF "Proceed" → Go to PHASE 2
   - IF "Modifications" → Ask for changes, re-run PHASE 1
   - IF "Review Details" → Show TAG results, re-ask approval
   - IF "Abort" → Go to PHASE 4 (graceful exit)

**Result**: User decision captured. Command proceeds or exits.

---

## 🚀 PHASE 2: Execute Document Synchronization

**Goal**: Synchronize documents with code changes, update TAGs, verify quality.

### Step 2.1: Create Safety Backup

Before making any changes:

1. **Generate timestamp**:
   - Execute: `date +%Y-%m-%d-%H%M%S` → Store as `$TIMESTAMP`

2. **Create backup directory**:
   - Execute: `mkdir -p .moai-backups/sync-$TIMESTAMP/`

3. **Backup critical files**:
   - Copy: `README.md` (if exists)
   - Copy: `docs/` directory (if exists)
   - Copy: `.moai/specs/` directory
   - Copy: `.moai/indexes/` directory (if exists)

4. **Verify backup**:
   - Execute: `ls -la .moai-backups/sync-$TIMESTAMP/`
   - IF empty → Print error and exit
   - ELSE → Print success message

**Result**: Safety backup created.

---

### Step 2.2: Invoke Doc-Syncer for Document Synchronization

**Your task**: Call doc-syncer to execute the approved synchronization plan.

Use Task tool:
- `subagent_type`: "doc-syncer"
- `description`: "Execute Living Document synchronization"
- `prompt`:
  ```
  You are the doc-syncer agent.

  Language settings:
  - conversation_language: [from config]

  **Execute the approved synchronization plan**:

  Previous analysis results:
  - TAG verification: [from tag-agent]
  - Synchronization strategy: [from doc-syncer analysis]

  **Task Instructions**:

  1. Living Document synchronization:
     - Reflect changed code in documentation
     - Auto-generate/update API documentation
     - Update README (if needed)
     - Synchronize Architecture documents

  2. @TAG system updates:
     - Update TAG index (.moai/indexes/tags.db)
     - Repair orphan TAGs (if possible)
     - Restore broken references

  3. SPEC synchronization:
     - Ensure SPEC documents match implementation
     - Update EARS statements if needed
     - Link @CODE/@TEST TAGs correctly

  4. Domain-based documentation:
     - Detect changed domains (frontend/backend/devops/database/ml/mobile)
     - Generate domain-specific documentation updates

  5. Generate synchronization report:
     - File location: .moai/reports/sync-report-$TIMESTAMP.md
     - Include: Updated file list, TAG repairs, results summary

  **Important**: Use conversation_language for all document updates.

  Execute the plan precisely and report results in detail.
  ```

**Store**: Response in `$SYNC_RESULTS`

---

### Step 2.3: Invoke Quality-Gate for Verification

**Your task**: Call quality-gate to verify synchronization quality.

Use Task tool:
- `subagent_type`: "quality-gate"
- `description`: "Verify document synchronization quality"
- `prompt`:
  ```
  You are the quality-gate agent.

  **Task**: Verify that document synchronization meets TRUST 5 principles.

  Synchronization results: [from doc-syncer]
  TAG validation: [from tag-agent]

  **Verification checks**:
  1. Test First: Are all TAG chains complete?
  2. Readable: Are documents well-formatted?
  3. Unified: Are all documents consistent?
  4. Secured: Are no credentials exposed?
  5. Trackable: Are all TAGs properly linked?

  **Output**: PASS / FAIL with details
  ```

**Result**: Quality verification complete.

---

## 🔧 PHASE 3: Git Operations & PR

**Goal**: Commit changes, transition PR (if Team mode), optionally auto-merge.

### Step 3.1: Invoke Git-Manager for Commit

**Your task**: Call git-manager to commit all document changes.

Use Task tool:
- `subagent_type`: "git-manager"
- `description`: "Commit document synchronization changes"
- `prompt`:
  ```
  You are the git-manager agent.

  **Task**: Commit document synchronization changes to Git.

  **Commit Scope**:
  - All changed document files
  - .moai/reports/ directory
  - .moai/indexes/ directory (if changed)
  - README.md (if changed)
  - docs/ directory (if changed)

  **Commit Message Template**:
  docs: sync documentation with code changes

  Synchronized Living Documents:
  - [list from synchronization results]

  TAG system updates:
  - [count] repairs completed
  - TAG index updated

  SPEC synchronization:
  - [count] SPECs updated

  Domain-specific sync:
  - [domain list if applicable]

  Generated with Claude Code

  Co-Authored-By: Alfred <alfred@mo.ai.kr>

  **Important**:
  - Pass commit message in HEREDOC format
  - Bundle all changes into a single commit
  - Report success after commit

  **Execution Order**:
  1. git add (changed document files)
  2. git commit -m (HEREDOC)
  3. git log -1 (verify commit)
  ```

**Verify**:
- Execute: `git log -1 --oneline`
- Print commit info
- IF commit failed → Exit with error code

---

### Step 3.2: (Optional) PR Ready Transition

For Team mode projects only:

1. **Check if Team mode**:
   - Read: `git_strategy.mode` from config
   - IF Personal → Skip to next phase

2. **Transition PR to Ready**:
   - Use Task tool:
     - `subagent_type`: "git-manager"
     - `description`: "Transition PR to Ready for Review"
     - `prompt`: "Transition PR from Draft to Ready. Execute: `gh pr ready`"

3. **Assign reviewers and labels** (if configured)

---

### Step 3.3: (Optional) PR Auto-Merge

If `--auto-merge` flag is set:

1. **Check CI/CD status**:
   - Execute: `gh pr checks`
   - IF failing → Print warning and skip merge

2. **Check merge conflicts**:
   - Execute: `gh pr view --json mergeable`
   - IF conflicts exist → Print warning and skip merge

3. **Execute auto-merge**:
   - Execute: `gh pr merge --squash --delete-branch`

4. **Branch cleanup**:
   - Checkout: `git checkout develop`
   - Pull: `git pull origin develop`
   - Delete local branch if merge succeeded

---

## 🎯 PHASE 4: Completion & Next Steps

**Goal**: Report results and guide user to next action.

### Step 4.1: Display Completion Report

Print comprehensive summary:

```
═══════════════════════════════════════════════════════
✅ Document Synchronization Complete
═══════════════════════════════════════════════════════

📊 Synchronization Summary:
- Mode: [mode]
- Scope: [scope]
- Files updated: [count]
- Files created: [count]
- TAG repairs: [count]

📚 Documents Updated:
- Living Documents: [list]
- SPEC documents: [list]
- Domain-specific reports: [count]

🔗 TAG System Status:
- @SPEC TAGs: [count]
- @CODE TAGs: [count]
- @TEST TAGs: [count]
- @DOC TAGs: [count]
- TAG chain integrity: [PASS / WARNING]

📄 Reports Generated:
- Master sync report: .moai/reports/sync-report-$TIMESTAMP.md
- Domain reports: [list if any]

💾 Backup Location:
- Safety backup: .moai-backups/sync-$TIMESTAMP/

═══════════════════════════════════════════════════════
```

---

### Step 4.2: Ask for Next Action

Use AskUserQuestion to guide next steps:

- `question`: "Documentation synchronization complete. What would you like to do next?"
- `header`: "Next Steps"
- `multiSelect`: false
- `options`: 3-4 choices depending on context:
  - "📋 Create Next SPEC" → /alfred:1-plan
  - "🔄 Start New Session" → /clear for fresh context
  - "📤 Review PR" (Team mode) → gh pr view --web
  - "🔧 Continue Development" (Personal mode)
  - "🎯 Project Overview" → Review reports and docs

---

## 🚨 Graceful Exit (User Aborts)

If user chooses to abort in PHASE 1:

```
═══════════════════════════════════════════════════════
❌ Synchronization Aborted
═══════════════════════════════════════════════════════

No changes were made to:
- Documents
- Git history
- Branch state

Your project remains in its current state.

You can retry synchronization anytime with:
/alfred:3-sync [mode]

═══════════════════════════════════════════════════════
```

Exit command with code 0.

---

## 📚 Quick Reference

**For synchronization details, consult**:
- `Skill("moai-alfred-tag-scanning")` - TAG system
- `Skill("moai-alfred-git-workflow")` - Git operations
- `Skill("moai-alfred-trust-validation")` - Quality gates
- CLAUDE.md - Full workflow documentation

**Version**: 3.1.0 (Agent-Delegated Pattern)
**Last Updated**: 2025-11-09
**Total Lines**: ~800 (reduced from 2,096)
**Architecture**: Commands → Agents → Skills
