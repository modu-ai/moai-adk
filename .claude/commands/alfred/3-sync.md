---
name: alfred:3-sync
description: Document synchronization + PR Ready conversion
argument-hint: "Mode target path - Mode: auto (default)|force|status|project, target path: Synchronization target path"
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
> Interactive prompts rely on `Skill("moai-alfred-interactive-questions")` so AskUserQuestion renders TUI selection menus for user surveys and approvals.

## 🚀 START HERE

**CRITICAL**: Load the TUI Survey Skill FIRST before any user interaction:

```
Skill("moai-alfred-interactive-questions")
```

This Skill MUST be loaded at the very beginning to enable TUI menu rendering for AskUserQuestion calls throughout this workflow.

## 🎯 Command Purpose

Synchronize code changes to Living Documents and verify @TAG system to ensure complete traceability.

**Document sync to**: $ARGUMENTS

> **Standard two-step workflow** (see `CLAUDE.md` - "Alfred Command Execution Pattern" for details)

## 📋 Execution flow

**Phase 0: Skill Loading** (IMMEDIATE)
- Load `Skill("moai-alfred-interactive-questions")` at the very start
- This enables TUI menu rendering for all user interactions

**Phase 1: Analysis & Planning**
1. **Project status analysis**: Git changes and TAG system verification
2. **Determine the scope of synchronization**: Full/partial/selective synchronization strategy
3. **User Confirmation**: Review and approve synchronization plan via AskUserQuestion (TUI menu)

**Phase 2: Conditional Execution** (based on user choice)
4. **Document Synchronization**: Living Document updates and TAG integrity guaranteed (IF user selects "Proceed")
5. **Git operations**: Commit and PR state transitions via git-manager (IF user selects "Proceed")
   - OR abort workflow (IF user selects "Abort")
   - OR revise plan (IF user selects "Modify")

## 🧠 Associated Skills & Agents

| Agent        | Core Skill                     | Purpose                        |
| ------------ | ------------------------------ | ------------------------------ |
| tag-agent    | `moai-alfred-tag-scanning`     | Verify TAG system integrity    |
| quality-gate | `moai-alfred-trust-validation` | Check code quality before sync |
| doc-syncer   | `moai-alfred-tag-scanning`     | Synchronize Living Documents   |
| git-manager  | `moai-alfred-git-workflow`     | Handle Git operations          |

**Note**: TUI Survey Skill is loaded once at Phase 0 and reused throughout all user interactions.

## 🔗 Associated Agent

- **Phase 1**: quality-gate (🛡️ Quality Assurance Engineer) - Quality verification before synchronization (conditional)
- **Primary**: doc-syncer (📖 Technical Writer) - Dedicated to document synchronization
- **Secondary**: git-manager (🚀 Release Engineer) - Dedicated to Git commits/PR

## 💡 Example of use

Users can run the command as follows:
- `/alfred:3-sync` - Auto-sync (PR Ready only)
- `/alfred:3-sync --auto-merge` - PR auto-merge + branch cleanup
- `/alfred:3-sync force` - Force full synchronization
- `/alfred:3-sync status` - Check synchronization status
- `/alfred:3-sync project` - Integrated project synchronization

### 🚀 Fully automated GitFlow (--auto-merge)

**Automatically performs the following actions when used in Team mode**:
1. Document synchronization complete
2. Switch to PR Ready
3. Check CI/CD status
4. PR automatic merge (squash)
5. Develop checkout and synchronization
6. Organizing local feature branches
7. **Ready for next task** ✅

**Recommended use time**: When you want to complete the merge in one go after completing TDD implementation.

**Personal mode**: Automate local main/develop merges and branch cleanups

## 🔍 STEP 1: Analyze synchronization scope and establish plan

Analyze project status to determine synchronization scope, develop a systematic synchronization plan, and receive user confirmation.

**The doc-syncer agent automatically scans the TAG chain and identifies and analyzes Git changes.**

### 🔍 TAG chain navigation (optional)

**If your TAG chain is complex or extensive**, utilize the Explore agent first:

```
Invoking the Task tool (Explore agent):
- subagent_type: "Explore"
- description: "Scan entire TAG system"
- prompt: "Please scan @TAG system throughout the project:
 - @SPEC TAG location (.moai/specs/)
 - @TEST TAG location (tests/)
 - @CODE TAG location (src/)
          - @DOC TAG location (docs/)
 - Detect orphan TAGs and broken references
 Thoroughness level: very thorough"
```

**Explore Agent When to Use**:
- ✅ Large projects (100+ files)
- ✅ When TAG chain integrity verification is required
- ✅ Changes across multiple SPECs
- ❌ Simple changes to a single SPEC

### ⚙️ How to call an agent

**In STEP 1, call doc-syncer and tag-agent using the Task tool**:

```
1. Tag-agent call (TAG verification):
   - subagent_type: "tag-agent"
- description: "Verify TAG system"
 - prompt: "Please verify the integrity of the entire TAG chain.
 Please verify the integrity of @SPEC, @TEST, @CODE, @DOC TAGs
 and orphan TAGs.
 (Optional) Explore results: $EXPLORE_RESULTS"

2. doc-syncer call (synchronization plan):
   - subagent_type: "doc-syncer"
- description: "Establish a document synchronization plan"
 - prompt: "Please analyze Git changes and establish a document synchronization plan.
             $ARGUMENTS
(Optional) TAG validation results: $TAG_VALIDATION_RESULTS"
```

### Synchronization analysis in progress

1. **Check project status**
 - Git status and changed file list
 - Code-document consistency check
 - @TAG system verification (using tag-agent or Explore)
 - (Optional) Extensive TAG scan based on Explore results

2. **Determine the scope of synchronization**
 - Living Document area requiring update
 - TAG index need to be updated
 - PR status transition possibility (team mode)

3. **Establish a synchronization strategy**
 - Synchronization approach for each mode
 - Estimated work time and priorities
 - Identify potential risks

### Phase 1 Details: Quality pre-verification (conditional automatic execution)

Quickly check code quality before synchronization.

**Differences from Phase 3 (2-build)**:
- **Phase 3**: In-depth verification after completion of TDD implementation (test coverage, code quality, security)
- **Phase 1**: Quick scan before synchronization (file corruption, critical issues only)

**Purpose**: Prevent documentation of code with quality issues

**Execution conditions (automatic judgment)**:
- Check the number of code change lines with Git diff
- Changed lines > 50 lines: Automatically run
- Changed lines ≤ 50 lines: Skip
- Change only document: Skip

**Verification items**:
- **Verify only changed files**: File targets verified by Git diff
- **TRUST principle verification**: Run trust-checker script
- **Code style**: Run linter (changed files only)
- **TAG chain**: Verify changed TAG integrity

**How ​​it works**:
Alfred automatically calls the quality-gate agent when there are a lot of code changes to perform quick quality verification before document synchronization.

**Handling verification results**:

✅ **PASS (0 Critical)**: Synchronization in progress

⚠️ **WARNING (0 Critical, Warning included)**: Synchronization proceeds after displaying warning.

❌ **CRITICAL (1 or more Critical)**: Synchronization stopped, correction recommended
- Critical issue found: Synchronization stopped, correction recommended
- User selection: “Retry after modification” or “Force proceed”

**Skip verification option**:
To skip pre-verification, use the `/alfred:3-sync --skip-pre-check` option.

---

### User verification steps

After reviewing your sync plan, `Skill("moai-alfred-interactive-questions")` presents the following options for user decision:
- **"Proceed"** or **"Start"**: Start synchronization as planned
- **"Modify [Contents]"**: Request modifications to your sync plan
- **"Abort"**: Abort the sync operation

---

## 🚀 STEP 2: Execute document synchronization (after user approval)

After user approval (collected via `Skill("moai-alfred-interactive-questions")`), the doc-syncer agent performs **Living Document synchronization and @TAG updates**, and optionally executes PR Ready transitions only in team mode.

### Phase 2 Details: SPEC Completion Processing (Automatic)

The doc-syncer agent automatically determines whether TDD implementation is complete and updates SPEC metadata.

**Automatic update conditions**:
- SPEC with status `draft`
- RED → GREEN → REFACTOR commit exists
- @TEST and @CODE TAG exist

**Update details**:
- `status: draft` → `status: completed`
- `version: 0.0.x` → `version: 0.1.0`
- Automatic addition of HISTORY section

**If conditions are not met**: Phase 2 detailed work is automatically skipped

## function

- **Automatic Document Synchronization**: The doc-syncer agent performs Living Document synchronization and @TAG updates. Optionally implements the PR Ready transition only in team mode.

## Synchronization output

- `.moai/reports/sync-report.md` creation/update
- TAG chain verification: Direct code scan (`rg '@TAG' -n src/ tests/`)

## Execution method by mode

## 📋 STEP 1 Implementation Guide: Analyzing the scope of synchronization and establishing a plan

### 1. Project status analysis

Alfred calls the doc-syncer agent to analyze synchronization targets and scopes.

#### Analysis Checklist

- [ ] **Git status**: Changed files, branch status, commit history
- [ ] **Document consistency**: Need for code-to-document synchronization
- [ ] **TAG system**: @TAG scheme verification and broken links
- [ ] **Sync scope**: Full vs partial vs specific path synchronization

### 2. Determine synchronization strategy

#### Mode-specific synchronization approach

| mode         | Synchronization range           | PR processing          | Key Features           |
| ------------ | ------------------------------- | ---------------------- | ---------------------- |
| **Personal** | Local document synchronization  | checkpoint only        | Focus on personal work |
| **Team**     | Full Sync + TAG                 | PR Ready conversion    | Collaboration support  |
| **Auto**     | Intelligent automatic selection | Decisions by situation | Optimal strategy       |
| **Force**    | Force full sync                 | Full regeneration      | For error recovery     |

#### Expected scope of work

- **Living Document**: API documentation, README, architecture document
- **TAG index**: Update `.moai/indexes/tags.db`
- **Sync report**: `.moai/reports/sync-report.md`
- **PR status**: Draft → Ready for Review transition

### 3. Generate synchronization plan report

Present your plan in the following format:

```
## Document Synchronization Plan Report: [TARGET]

### 📊 Health Analysis Results
- **Changed Files**: [Number and Type]
- **Synchronization Required**: [High/Medium/Low]
- **TAG System Status**: [Healthy/Problem Detected]

### 🎯 Sync Strategy
- **Selected Mode**: [auto/force/status/project]
- **Sync Scope**: [Full/Partial/Selective]
- **PR Handling**: [Maintain/Switch Ready/Create New PR]

### ⚠️ Notes
- **Potential conflicts**: [Possible document conflicts]
- **TAG issues**: [Broken links, duplicate TAGs]
- **Performance impact**: [Estimated time for large synchronization]

### ✅ Expected deliverables
- **sync-report.md**: [Summary of sync results]
- **tags.db**: [Updated TAG index]
- **Living Documents**: [Updated document list]
- **PR Status**: [PR transition in team mode]

---
**Approval Request**: Do you want to proceed with synchronization using the above plan?
 (select “Proceed”, “Modify [Content]”, or “Abort”)
```

---

## 🚀 STEP 2 Implementation Guide: Document Synchronization (After Approval)

Only when the user selects **"Proceed"** or **"Start"** will Alfred call the doc-syncer agent to perform Living Document synchronization and TAG updates.

### Sync step-by-step guide

1. **Living Document Synchronization**: Code → Document automatically reflected
2. **TAG System Verification**: @TAG System Integrity Verification
3. **Index Update**: Traceability Matrix Update
4. **Create Report**: Create a summary of synchronization results

### Agent collaboration structure

- **Step 1**: The `doc-syncer` agent is dedicated to Living Document synchronization and @TAG management.
- **Step 2**: The `git-manager` agent is dedicated to all Git commits, PR state transitions, and synchronization.
- **Single Responsibility Principle**: doc-syncer only performs document tasks, and git-manager only performs Git tasks.
- **Sequential execution**: Executes in the order doc-syncer → git-manager to maintain clear dependencies.
- **No inter-agent calls**: Each agent does not directly call other agents, and executes commands. Runs sequentially in levels only.

## 🚀 Optimized parallel/sequential hybrid workflow

### Phase 1: Quick status check (parallel execution)

Do the following **simultaneously**:

```
Task 1 (haiku): Check Git status
├── Collect list of changed files
├── Check branch status
└── Determine need for synchronization

Task 2 (sonnet): Analyze document structure
├── Detect project type
├── Collect TAG list
└── Determine synchronization scope
```

### Phase 2: Document synchronization (sequential execution)

The `doc-syncer` agent (sonnet) handles intensive processing:

- Living Document synchronization
- @TAG system verification and update
- Document-code consistency check
- TAG traceability matrix update

### Phase 3: Git task processing (sequential execution)

Final processing by the `git-manager` agent (haiku):

- Commit document changes
- Apply synchronization strategy for each mode
- Switch PR Ready in Team mode
- Automatically assign reviewers (using gh CLI)

### Phase 4: PR merge and branch cleanup (optional)

Additional processing by `git-manager` when using the `--auto-merge` flag:

**Team mode (GitFlow)**:
1. Check PR status (CI/CD pass check)
2. PR automatic merge (to develop branch)
3. Delete remote feature branch
4. Local develop checkout and synchronization
5. Organizing local feature branches
6. Notification that the next task is ready

**Personal Mode**:
1. Local main/develop merge
2. Delete feature branch
3. Check out the base branch
4. Notification that the next task is ready

**Performance improvements**: Minimize latency by parallelizing the initial verification step

### Argument handling

- **$1 (mode)**: `$1` → `auto` (default)|`force`|`status`|`project`
- **$2 (path)**: `$2` → Sync target path (optional)
- **flags**:
 - `--auto-merge`: Enable PR automatic merge and branch cleanup (Team mode)
 - `--skip-pre-check`: Skip pre-quality check
 - `--skip-quality-check`: Skip final quality check

**Command usage example**:
- `/alfred:3-sync` - Basic automatic synchronization (optimized by mode)
- `/alfred:3-sync --auto-merge` - PR automatic merge + branch cleanup (Team mode recommended)
- `/alfred:3-sync force` - Force full synchronization
- `/alfred:3-sync status` - Check synchronization status
- `/alfred:3-sync project` - Integrated project synchronization
- `/alfred:3-sync auto src/auth/` - Specific path Synchronization
- `/alfred:3-sync --auto-merge --skip-pre-check` - Fast merge

### Agent role separation

#### doc-syncer dedicated area

- Living Document synchronization (code ↔ document)
- @TAG system verification and update
- Automatic creation/update of API document
- README and architecture document synchronization
- Verification of document-code consistency

#### git-manager dedicated area

- All Git commit operations (add, commit, push)
- Apply synchronization strategy for each mode
- PR status transition (Draft → Ready)
- **PR auto merge** (when --auto-merge flag)
 - Check CI/CD status
 - Conflict verification
 - Execute Squash merge
  - Remote branch deletion
- **Branch cleanup and conversion**
 - Local develop checkout
 - Remote synchronization (git pull)
 - Local feature branch deletion
- Automatic assignment and labeling of reviewers
- GitHub CLI integration and remote synchronization

### 🧪 Personal Mode

- The git-manager agent automatically creates checkpoints before and after synchronization
- The README, in-depth documentation, and PR body are organized manually according to the checklist.

### 🏢 Team Mode

- Full synchronization of Living Document + @TAG verification/correction
- Optionally perform PR Ready conversion and labeling only when gh CLI is set
- Fully automated when using **--auto-merge flag**:
 1. Document synchronization complete.
  2. git push origin feature/SPEC-{ID}
  3. gh pr ready {PR_NUMBER}
4. Check CI/CD status (gh pr checks)
  5. gh pr merge --squash --delete-branch
  6. git checkout develop && git pull origin develop
7. Notification that the next task is ready

**Important**: All Git operations (commit, sync, PR management) are handled by the git-manager agent, so this command does not run Git operations directly.

**Branch Policy**:
- Base branch: `develop` (GitFlow standard)
- After merge: automatically checkout `develop`
- Next `/alfred:1-plan` automatically starts in `develop`

## Synchronization Details (Summary)

1. Project analysis and TAG verification → Check broken/duplicate/orphaned TAG
2. Code ↔ Document synchronization → API/README/architecture document update, SPEC ↔ Code TODO synchronization
3. TAG chain verification → `rg '@TAG' -n src/ tests/` (scan code directly)

## Next steps

**Recommendation**: For better performance and context management, start a new chat session with the `/clear` or `/new` command before proceeding to the next step.

- The entire MoAI-ADK workflow is completed after document synchronization is completed
- All Git operations are dedicated to the git-manager agent to ensure consistency
- Only command-level orchestration is used without direct calls between agents

## Report results

Report synchronization results in a structured format:

### Successful synchronization (summary example)

✅ Document synchronization complete — Update N, Create M, TAG Modify K, Verification passed

### Partial synchronization (problem detected)

```
⚠️ Partial sync completed (issue found)

❌ Problems that need solving:
├── Broken links: X (specific list)
├── Duplicate TAG: X
└── Orphan TAG: X

🛠️ Auto-correction recommendations:
1. Broken link recovery
2. Merge duplicate TAGs
3. Orphan TAG cleanup
```

## Next steps guidance

### Development cycle complete

**Default mode (PR Ready only)**:
```
🔄 MoAI-ADK 3-step workflow completion:
✅ /alfred:1-plan → Create EARS specification (feature/SPEC-{ID} branch)
✅ /alfred:2-run → TDD implementation
✅ /alfred:3-sync → Document synchronization + PR Ready

⏳ Next steps: PR review and manual merge required
> gh pr view (check PR)
> gh pr merge --squash (merge after review)
```

**Auto Merge Mode (Recommended)**:
```
🔄 Fully automated GitFlow workflow:
✅ /alfred:1-plan → EARS specification creation (from develop)
✅ /alfred:2-run → TDD implementation
✅ /alfred:3-sync --auto-merge → Document synchronization + PR Merge + branch cleanup

🎉 Automatic switch to develop branch done!
📍 You are here: develop (ready for next work)
> /alfred:1-plan "Describe next feature" # Create new branch in develop
```

### Integrated project mode

**When to use**:
- When the implementation of multiple SPECs has been completed and the entire project documentation needs to be updated
- When periodic synchronization of the entire document in Personal mode is required.

**Differences from Personal/Team mode**:
- **Personal/Team mode**: Synchronize only specific SPEC-related documents
- **Project mode**: Synchronize README, architecture documentation, and entire API documentation

**Output**:
- README.md (updated complete feature list)
- docs/architecture.md (updated system design)
- docs/api/ (unified API documentation)
- .moai/indexes/ (rebuilt full TAG index)

```
🏢 Integrated branch sync complete!

📋 Entire project synchronization:
├── README.md (full feature list)
├── docs/architecture.md (system design)
├── docs/api/ (unified API documentation)
└── .moai/indexes/ (full TAG index)

🎯 PR conversion support completed
```

## Constraints and Assumptions

**Environment Dependency:**

- Git repository required
- gh CLI (required for GitHub integration)
- Python3 (TAG verification script)

**Prerequisites:**

- MoAI-ADK project structure (.moai/, .claude/)
- TDD implementation completion status
- Compliance with TRUST 5 principles

**Limitations:**

- TAG verification is based on file existence
- PR automatic conversion only works in gh CLI environment
- Coverage figures need to be measured separately

---

## 🧠 Context Management

> For more information: `.moai/memory/development-guide.md` - see section "Context Engineering"

### Core strategy of this command

**Load first**: `.moai/reports/sync-report-latest.md` (old sync state)

**Recommendation**: Document synchronization is complete. Now that the entire MoAI-ADK cycle (1-spec → 2-build → 3-sync) has been completed, start a new conversation session with the `/clear` or `/new` command before developing the next feature.

---

**Aims to improve code-document consistency and ensure @TAG traceability by linking with the doc-syncer subagent.**
