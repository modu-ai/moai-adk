---
name: alfred:3-sync
description: "Synchronize documentation and finalize PR"
argument-hint: 'Mode target path - Mode: auto (default)|force|status|project, target
  path: Synchronization target path'
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
│ STEP 0: Load Skills (IMMEDIATE)                         │
│  → Load moai-alfred-ask-user-questions Skill            │
│  → Enable TUI menu rendering                            │
└──────────────────────────────────────────────────────────┘
                          ↓
┌──────────────────────────────────────────────────────────┐
│ STEP 1: Analysis & Planning                             │
│  STEP 1.1: Verify prerequisites                         │
│  STEP 1.2: Analyze project status (Git + TAG)           │
│  STEP 1.3: Determine sync scope (mode-specific)         │
│  STEP 1.4: (Optional) TAG chain navigation              │
│  STEP 1.5: Create synchronization plan                  │
│  STEP 1.6: Request user approval (AskUserQuestion)      │
└──────────────────────────────────────────────────────────┘
                          ↓
          ┌───────────────┴───────────────┐
          │                               │
     User chooses                    User chooses
     "Proceed"                       "Abort/Modify"
          │                               │
          ↓                               ↓
┌─────────────────────────┐   ┌──────────────────────┐
│ STEP 2: Execute Sync    │   │ STEP 4: Graceful Exit│
│  STEP 2.1: Safety backup│   │  → Display abort msg │
│  STEP 2.2: Living Doc   │   │  → OR re-analyze     │
│  STEP 2.3: TAG update   │   └──────────────────────┘
│  STEP 2.4: SPEC sync    │
│  STEP 2.5: Domain sync  │
│  STEP 2.6: Completion   │
└─────────────────────────┘
          ↓
┌──────────────────────────────────────────────────────────┐
│ STEP 3: Git Operations & PR                             │
│  STEP 3.1: Commit document changes (git-manager)        │
│  STEP 3.2: (Optional) PR Ready transition               │
│  STEP 3.3: (Optional) PR auto-merge (--auto-merge flag) │
│  STEP 3.4: (Optional) Branch cleanup                    │
│  STEP 3.5: Display completion report                    │
└──────────────────────────────────────────────────────────┘
```

---

## 🔧 STEP 0: Load Skills (IMMEDIATE)

**Your task**: Load the TUI Survey Skill at the very beginning to enable interactive menus.

**Steps**:

1. **Load the skill immediately**:
   ```
   Invoke Skill: moai-alfred-ask-user-questions
   ```

2. **Why this matters**:
   - This skill provides TUI menu rendering capabilities
   - Must be loaded BEFORE any user interaction
   - Used in STEP 1.6, STEP 2.6, and STEP 3.5

**Result**: TUI menu system ready for all subsequent user interactions.

**Next step**: Go to STEP 1.1

---

## 📊 STEP 1: Analysis & Planning

### STEP 1.1: Verify Prerequisites

**Your task**: Check that all required components exist before starting synchronization analysis.

**Steps**:

1. **Check MoAI-ADK project structure**:
   - Directory to check: `.moai/` and `.claude/`
   - IF `.moai/` does NOT exist → Print error and exit:
     ```
     ❌ Error: Not a MoAI-ADK project

     This command requires MoAI-ADK structure:
     - .moai/config.json (project configuration)
     - .moai/specs/ (SPEC documents)
     - .claude/ (Claude Code configuration)

     Run `/alfred:0-project init` to initialize a project.
     ```

2. **Check Git repository status**:
   - Execute: `git rev-parse --is-inside-work-tree`
   - IF NOT a Git repository → Print error and exit:
     ```
     ❌ Error: Not a Git repository

     This command requires Git version control.
     Initialize with: git init
     ```

3. **Verify Python environment** (for TAG verification):
   - Execute: `which python3`
   - IF python3 NOT found → Print warning (non-fatal):
     ```
     ⚠️ Warning: Python3 not found

     TAG verification scripts may not run.
     Synchronization will continue with limited TAG checks.
     ```

4. **Print prerequisites status**:
   ```
   ✅ Prerequisites verified:
   - MoAI-ADK structure: OK
   - Git repository: OK
   - Python environment: OK (or WARNING)
   ```

**Result**: All prerequisites validated. Ready for analysis.

**Next step**: Go to STEP 1.2

---

### STEP 1.2: Analyze Project Status

**Your task**: Gather Git changes and project metadata to determine synchronization scope.

**Steps**:

1. **Check Git status**:
   - Execute: `git status --porcelain`
   - Store result in variable: `$GIT_STATUS`

2. **Get list of changed files**:
   - Execute: `git diff --name-only HEAD`
   - Store result in variable: `$CHANGED_FILES`

3. **Count changes by type**:
   - Python files changed: Count files matching `*.py` in `$CHANGED_FILES`
   - Test files changed: Count files in `tests/` directory
   - Document files changed: Count files matching `*.md` in `$CHANGED_FILES`
   - SPEC files changed: Count files in `.moai/specs/` directory

4. **Read project configuration**:
   - Read file: `.moai/config.json`
   - Extract values:
     - `git_strategy.mode` → Store as `$PROJECT_MODE` (Personal/Team)
     - `git_strategy.spec_git_workflow` → Store as `$WORKFLOW` (feature_branch/develop_direct)
     - `language.conversation_language` → Store as `$LANG`

5. **Determine if changes are code-heavy or document-heavy**:
   - IF Python files changed > 50 lines → `$CHANGE_TYPE = "code-heavy"`
   - ELSE IF document files changed > 10 lines → `$CHANGE_TYPE = "doc-heavy"`
   - ELSE → `$CHANGE_TYPE = "mixed"`

6. **Print analysis summary**:
   ```
   📊 Project Status Analysis

   Git Status:
   - Changed files: [count from $CHANGED_FILES]
   - Python files: [count]
   - Test files: [count]
   - Documents: [count]
   - SPEC files: [count]

   Project Configuration:
   - Mode: $PROJECT_MODE (Personal/Team)
   - Workflow: $WORKFLOW
   - Language: $LANG
   - Change type: $CHANGE_TYPE
   ```

**Result**: Project status analyzed and stored in variables.

**Next step**: Go to STEP 1.3

---

### STEP 1.3: Determine Sync Scope (Mode-Specific)

**Your task**: Determine which files and documents need synchronization based on the selected mode.

**Steps**:

1. **Read mode from $ARGUMENTS**:
   - Parse first argument: `$1` → Store as `$MODE`
   - IF `$MODE` is empty → Set `$MODE = "auto"`
   - Valid modes: `auto`, `force`, `status`, `project`

2. **IF mode is "status"**:
   - **Your task**: Perform quick status check only (no synchronization)
   - Execute: `python3 .moai/scripts/tag_scanner.py --status-only`
   - Read file: `.moai/reports/sync-report-latest.md` (if exists)
   - Print status:
     ```
     📊 Synchronization Status

     Last sync: [date from sync-report-latest.md]

     TAG System Health:
     - @SPEC TAGs: [count] found
     - @CODE TAGs: [count] found
     - @TEST TAGs: [count] found
     - @DOC TAGs: [count] found
     - Orphan TAGs: [count] (if any)

     Git Status:
     - Changed files since last sync: [count]
     - Uncommitted changes: [yes/no]

     Recommendation: [Sync needed / Up to date]
     ```
   - **STOP HERE** → Exit command (status mode complete)

3. **IF mode is "force"**:
   - **Your task**: Force full re-synchronization of all documents
   - Set scope variables:
     - `$SYNC_SCOPE = "full"`
     - `$TARGET_DIRS = "src/ tests/ docs/ .moai/specs/"`
     - `$REGENERATE_ALL = true`
   - Print:
     ```
     🔄 Force Mode Activated

     Synchronization scope: FULL PROJECT
     - All source files will be scanned
     - All documents will be regenerated
     - All TAG chains will be re-verified
     ```

4. **IF mode is "project"**:
   - **Your task**: Project-wide integrated synchronization
   - Set scope variables:
     - `$SYNC_SCOPE = "project"`
     - `$TARGET_DIRS = "src/ tests/ docs/ .moai/specs/ README.md"`
     - `$UPDATE_PROJECT_DOCS = true`
   - Print:
     ```
     🏢 Project Mode Activated

     Synchronization scope: INTEGRATED PROJECT
     - README.md will be updated (full feature list)
     - docs/architecture.md will be updated
     - docs/api/ will be unified
     - .moai/indexes/ will be rebuilt
     ```

5. **IF mode is "auto"** (default):
   - **Your task**: Smart selective synchronization based on Git changes
   - Set scope variables:
     - `$SYNC_SCOPE = "selective"`
     - `$TARGET_DIRS = [directories from $CHANGED_FILES]`
     - `$UPDATE_PROJECT_DOCS = false`
   - Determine selective scope:
     - IF `$CHANGE_TYPE = "code-heavy"` → Include `src/`, `tests/`, related SPEC
     - IF `$CHANGE_TYPE = "doc-heavy"` → Include `docs/`, `.moai/specs/`
     - IF `$CHANGE_TYPE = "mixed"` → Include all changed directories
   - Print:
     ```
     🎯 Auto Mode Activated

     Synchronization scope: SELECTIVE
     - Target directories: $TARGET_DIRS
     - Changed files: [count]
     - Estimated sync time: [based on change count]
     ```

6. **Parse additional flags from $ARGUMENTS**:
   - Search for `--auto-merge` flag → Set `$AUTO_MERGE = true` (default: false)
   - Search for `--skip-pre-check` flag → Set `$SKIP_PRE_CHECK = true` (default: false)
   - Search for `--skip-quality-check` flag → Set `$SKIP_QUALITY_CHECK = true` (default: false)

**Result**: Synchronization scope determined and stored in variables.

**Next step**:
- IF mode was "status" → EXIT (already completed)
- ELSE → Go to STEP 1.4

---

### STEP 1.4: (Optional) TAG Chain Navigation

**Your task**: Optionally perform comprehensive TAG chain exploration for large projects.

**Decision point**:

1. **Determine if TAG exploration is needed**:
   - IF `$MODE = "force"` OR `$MODE = "project"` → TAG exploration REQUIRED
   - ELSE IF changed files > 100 → TAG exploration RECOMMENDED
   - ELSE IF `$SYNC_SCOPE = "selective"` → SKIP exploration (go to STEP 1.5)

2. **IF TAG exploration is REQUIRED or RECOMMENDED**:
   - Print:
     ```
     🔍 TAG Chain Navigation

     Performing comprehensive TAG system scan...
     This may take a few moments for large projects.
     ```

3. **Invoke Explore agent for TAG scanning**:
   - **Your task**: Call the Explore agent to scan entire TAG system
   - Use Task tool:
     - `subagent_type`: "Explore"
     - `description`: "Scan entire TAG system across project"
     - `prompt`:
       ```
       프로젝트 전체에서 @TAG 시스템을 스캔해주세요:

       스캔 범위:
       - @SPEC TAG 위치 (.moai/specs/)
       - @TEST TAG 위치 (tests/)
       - @CODE TAG 위치 (src/)
       - @DOC TAG 위치 (docs/)

       검증 항목:
       - 고아 TAG (orphan TAGs) 감지
       - 끊긴 참조 (broken references) 감지
       - 중복 TAG (duplicate TAGs) 감지

       상세도 수준: very thorough

       출력 형식:
       - TAG 인벤토리 전체 목록
       - 문제 TAG 목록 (위치 포함)
       - 추천 수정 사항
       ```

4. **Store Explore agent results**:
   - Read response from Explore agent
   - Store in variable: `$EXPLORE_RESULTS`
   - Print summary:
     ```
     ✅ TAG exploration complete

     TAG inventory:
     - @SPEC TAGs found: [count]
     - @CODE TAGs found: [count]
     - @TEST TAGs found: [count]
     - @DOC TAGs found: [count]

     Issues detected:
     - Orphan TAGs: [count]
     - Broken references: [count]
     - Duplicate TAGs: [count]
     ```

5. **IF TAG exploration was SKIPPED**:
   - Set `$EXPLORE_RESULTS = null`
   - Print:
     ```
     ⏩ TAG exploration skipped (not needed for selective sync)
     ```

**Result**: TAG chain exploration completed (or skipped). Results stored in `$EXPLORE_RESULTS`.

**Next step**: Go to STEP 1.5

---

### STEP 1.5: Create Synchronization Plan

**Your task**: Call tag-agent and doc-syncer to verify TAG integrity and establish a detailed synchronization plan.

**This phase runs TWO agents sequentially**:

1. **Tag-agent call (TAG verification across ENTIRE PROJECT)**:

   - **Your task**: Invoke tag-agent to verify TAG system integrity
   - Use Task tool:
     - `subagent_type`: "tag-agent"
     - `description`: "Verify TAG system across entire project"
     - `prompt`:
       ```
       전체 프로젝트에서 포괄적인 @TAG 시스템 검증을 수행해주세요.

       **필수 범위**: 변경된 파일만이 아니라 모든 소스 파일을 스캔합니다.

       **검증 항목**:
       1. .moai/specs/ 디렉토리의 @SPEC TAG
       2. tests/ 디렉토리의 @TEST TAG
       3. src/ 디렉토리의 @CODE TAG
       4. docs/ 디렉토리의 @DOC TAG

       **고아 감지** (필수):
       - 매칭되는 @SPEC이 없는 @CODE TAG 감지
       - 매칭되는 @CODE가 없는 @SPEC TAG 감지
       - 매칭되는 @SPEC이 없는 @TEST TAG 감지
       - 매칭되는 @SPEC/@CODE가 없는 @DOC TAG 감지

       **출력 형식**:
       - 고아 TAG의 전체 목록을 위치와 함께 제공합니다.
       - TAG 체인 무결성 평가 (Healthy / Issues Detected)

       (선택사항) 탐색 결과: $EXPLORE_RESULTS
       ```

   - **Wait for tag-agent response**
   - Store response in variable: `$TAG_VALIDATION_RESULTS`
   - Print summary:
     ```
     ✅ TAG verification complete

     TAG chain integrity: [Healthy / Issues Detected]

     Issues found (if any):
     - Orphan @CODE TAGs: [list]
     - Orphan @SPEC TAGs: [list]
     - Broken references: [list]
     ```

2. **Doc-syncer call (synchronization plan establishment)**:

   - **Your task**: Invoke doc-syncer to analyze Git changes and create sync plan
   - Use Task tool:
     - `subagent_type`: "doc-syncer"
     - `description`: "Establish a document synchronization plan"
     - `prompt`:
       ```
       당신은 doc-syncer 에이전트입니다.

       언어 설정:
       - 대화_언어: $LANG
       - 언어명: [Korean/English/Japanese based on $LANG]

       중요 지시사항:
       문서 업데이트는 대화_언어를 반드시 존중해야 합니다:
       - 사용자 대면 문서 (README, 가이드): $LANG
       - SPEC 문서 (spec.md, plan.md, acceptance.md): $LANG
       - 코드 주석: $LANG (기술 키워드 제외)
       - 기술 문서 및 YAML 프론트매터: 영어

       스킬 호출:
       필요 시 명시적 Skill() 호출 사용:
       - Skill("moai-foundation-tags") - TAG 체인 검증
       - Skill("moai-foundation-trust") - 품질 게이트 검사
       - Skill("moai-alfred-tag-scanning") - TAG 인벤토리 업데이트

       작업:
       Git 변경사항을 분석하고 문서 동기화 계획을 수립해주세요.
       모든 문서 업데이트가 대화_언어 설정과 일치하는지 확인합니다.

       동기화 모드: $MODE
       동기화 범위: $SYNC_SCOPE
       대상 디렉토리: $TARGET_DIRS
       변경된 파일: $CHANGED_FILES

       (필수) TAG 검증 결과: $TAG_VALIDATION_RESULTS
       (선택사항) 탐색 결과: $EXPLORE_RESULTS
       ```

   - **Wait for doc-syncer response**
   - Store response in variable: `$SYNC_PLAN`
   - Print summary:
     ```
     📋 Synchronization Plan Created

     Documents to update:
     - Living Documents: [list]
     - SPEC documents: [list]
     - TAG indexes: [list]

     Estimated work:
     - Files to update: [count]
     - New files to create: [count]
     - TAG repairs needed: [count]
     - Estimated time: [based on change count]
     ```

**Result**: TAG validation results and synchronization plan stored in variables.

**Next step**: Go to STEP 1.6

---

### STEP 1.6: Request User Approval

**Your task**: Present the synchronization plan to the user and request approval to proceed.

**Steps**:

1. **Display comprehensive plan report**:
   ```
   ═══════════════════════════════════════════════════════
   📚 Document Synchronization Plan Report
   ═══════════════════════════════════════════════════════

   📊 Project Analysis:
   - Mode: $MODE
   - Scope: $SYNC_SCOPE
   - Changed files: [count from $CHANGED_FILES]
   - Project mode: $PROJECT_MODE (Personal/Team)

   🎯 Synchronization Strategy:
   - Living Documents to update: [list from $SYNC_PLAN]
   - SPEC documents to sync: [list]
   - TAG repairs needed: [count from $TAG_VALIDATION_RESULTS]
   - Domain-specific sync: [if applicable]

   ⚠️ TAG System Status:
   - TAG chain integrity: [Healthy / Issues Detected]
   - Orphan TAGs: [count]
   - Broken references: [count]

   ✅ Expected Deliverables:
   - sync-report.md: Summary of synchronization results
   - tags.db: Updated TAG index
   - Living Documents: [list]
   - PR Status: [if Team mode: Draft → Ready transition]

   ═══════════════════════════════════════════════════════
   ```

2. **Ask user for approval using AskUserQuestion**:
   - **Your task**: Use the AskUserQuestion tool to gather user decision
   - Tool call:
     - `questions`: Array with 1 question
     - Question details:
       - `question`: "Synchronization plan is ready. How would you like to proceed?"
       - `header`: "Plan Approval"
       - `multiSelect`: false
       - `options`: Array with 4 choices:
         1. Label: "✅ Proceed with Sync", Description: "Execute document synchronization as planned"
         2. Label: "🔄 Request Modifications", Description: "Specify changes to the synchronization strategy"
         3. Label: "🔍 Review Details", Description: "Re-examine TAG validation results and changes"
         4. Label: "❌ Abort", Description: "Cancel synchronization, keep current state"

3. **Wait for user response**:
   - Store response in variable: `$USER_DECISION`
   - Read value from: `$USER_DECISION["0"]` (first question answer)

4. **Process user response**:

   **IF user chose "✅ Proceed with Sync"**:
   - Print:
     ```
     ✅ User approved synchronization plan

     Proceeding to document synchronization...
     ```
   - **Next step**: Go to STEP 2.1

   **IF user chose "🔄 Request Modifications"**:
   - Print:
     ```
     🔄 User requested modifications to plan

     Please specify what changes you'd like to the synchronization strategy:
     ```
   - Wait for user input (freeform text)
   - Store input in variable: `$MODIFICATION_REQUEST`
   - Print:
     ```
     Re-analyzing with requested modifications...
     ```
   - **Next step**: Go back to STEP 1.5 (re-create plan with modifications)

   **IF user chose "🔍 Review Details"**:
   - Print detailed TAG validation results:
     ```
     🔍 Detailed TAG Validation Results

     $TAG_VALIDATION_RESULTS (full output)

     Detailed synchronization plan:

     $SYNC_PLAN (full output)
     ```
   - After displaying details, re-present the approval question
   - **Next step**: Go back to STEP 1.6 (re-ask approval)

   **IF user chose "❌ Abort"**:
   - Print:
     ```
     ❌ Synchronization aborted by user

     No changes were made to documents or Git history.
     Current branch state maintained.
     ```
   - **Next step**: Go to STEP 4 (Graceful Exit)

**Result**: User decision captured. Command proceeds or exits based on choice.

**Next step**: Based on user decision (see above)

---

## 🚀 STEP 2: Execute Document Synchronization

### STEP 2.1: Create Safety Backup

**Your task**: Create a safety backup of current document state before making any changes.

**Steps**:

1. **Create backup directory with timestamp**:
   - Generate timestamp: `date +%Y-%m-%d-%H%M%S` → Store as `$TIMESTAMP`
   - Create directory: `.moai-backups/sync-$TIMESTAMP/`
   - Execute: `mkdir -p .moai-backups/sync-$TIMESTAMP/`

2. **Copy current documents to backup**:
   - Copy README.md: `cp README.md .moai-backups/sync-$TIMESTAMP/` (if exists)
   - Copy docs/: `cp -r docs/ .moai-backups/sync-$TIMESTAMP/` (if exists)
   - Copy .moai/specs/: `cp -r .moai/specs/ .moai-backups/sync-$TIMESTAMP/`
   - Copy .moai/indexes/: `cp -r .moai/indexes/ .moai-backups/sync-$TIMESTAMP/` (if exists)

3. **Verify backup creation**:
   - Execute: `ls -la .moai-backups/sync-$TIMESTAMP/`
   - IF backup directory is empty → Print error and exit:
     ```
     ❌ Error: Backup creation failed

     Unable to create safety backup at:
     .moai-backups/sync-$TIMESTAMP/

     Synchronization aborted to prevent data loss.
     ```
   - ELSE → Print success:
     ```
     💾 Safety backup created

     Backup location: .moai-backups/sync-$TIMESTAMP/
     Files backed up: [count]

     You can restore from this backup if needed.
     ```

**Result**: Safety backup created. Safe to proceed with document modifications.

**Next step**: Go to STEP 2.2

---

### STEP 2.2: Synchronize Living Documents

**Your task**: Call doc-syncer agent to perform Living Document synchronization and TAG updates.

**Steps**:

1. **Invoke doc-syncer agent for synchronization execution**:
   - **Your task**: Call doc-syncer to execute the approved synchronization plan
   - Use Task tool:
     - `subagent_type`: "doc-syncer"
     - `description`: "Execute Living Document synchronization"
     - `prompt`:
       ```
       당신은 doc-syncer 에이전트입니다.

       언어 설정:
       - 대화_언어: $LANG

       **승인된 동기화 계획 실행**:

       이전 분석 결과:
       - TAG 검증 결과: $TAG_VALIDATION_RESULTS
       - 동기화 계획: $SYNC_PLAN
       - 탐색 결과: $EXPLORE_RESULTS (if available)

       **작업 지시**:

       1. Living Document 동기화:
          - 변경된 코드를 문서에 반영
          - API 문서 자동 생성/업데이트
          - README 업데이트 (if needed)
          - Architecture 문서 동기화

       2. @TAG 시스템 업데이트:
          - TAG 인덱스 갱신 (.moai/indexes/tags.db)
          - 고아 TAG 수리 (if possible)
          - 끊긴 참조 복구

       3. 문서-코드 일관성 검증:
          - SPEC ↔ CODE 동기화 확인
          - TEST ↔ CODE 매칭 확인
          - DOC ↔ CODE 참조 확인

       4. 동기화 리포트 생성:
          - 파일 위치: .moai/reports/sync-report-$TIMESTAMP.md
          - 포함 내용: 업데이트된 파일 목록, TAG 수리 내역, 남은 이슈

       **중요**: 모든 문서 생성/수정은 대화_언어($LANG)를 사용하세요.

       승인된 계획을 정확히 실행하고, 결과를 상세히 보고해주세요.
       ```

2. **Wait for doc-syncer completion**:
   - Monitor doc-syncer agent execution
   - Store results in variable: `$SYNC_RESULTS`

3. **Verify synchronization results**:
   - Read file: `.moai/reports/sync-report-$TIMESTAMP.md`
   - IF file does NOT exist → Print error:
     ```
     ⚠️ Warning: Sync report not generated

     doc-syncer may have encountered an issue.
     Checking for partial sync results...
     ```
   - ELSE → Print success:
     ```
     ✅ Living Document synchronization complete

     Sync report: .moai/reports/sync-report-$TIMESTAMP.md
     ```

4. **Print synchronization summary**:
   ```
   📚 Synchronization Results:

   Documents updated:
   - [list from $SYNC_RESULTS]

   TAG repairs:
   - [count] orphan TAGs fixed
   - [count] broken references repaired

   New files created:
   - [list]

   Sync report: .moai/reports/sync-report-$TIMESTAMP.md
   ```

**Result**: Living Documents synchronized. TAG system updated.

**Next step**: Go to STEP 2.3

---

### STEP 2.3: Update TAG Index

**Your task**: Update the TAG traceability index to reflect current state.

**Steps**:

1. **Check if TAG scanner script exists**:
   - Check file: `.moai/scripts/tag_scanner.py`
   - IF file does NOT exist → Print warning and skip:
     ```
     ⚠️ Warning: TAG scanner script not found

     Skipping TAG index update.
     TAG verification was performed by doc-syncer agent.
     ```
   - IF file exists → Proceed to next step

2. **Run TAG scanner to update index**:
   - Execute: `python3 .moai/scripts/tag_scanner.py --update-index`
   - Store exit code in variable: `$TAG_SCANNER_EXIT`

3. **Verify TAG scanner success**:
   - IF `$TAG_SCANNER_EXIT != 0` → Print warning:
     ```
     ⚠️ Warning: TAG scanner encountered issues

     Exit code: $TAG_SCANNER_EXIT

     TAG index may be incomplete. Check:
     .moai/logs/tag-scanner.log
     ```
   - ELSE → Print success:
     ```
     ✅ TAG index updated

     Index location: .moai/indexes/tags.db
     ```

4. **Verify TAG chain integrity**:
   - Execute: `rg '@TAG' -n src/ tests/` (direct code scan)
   - Store output in variable: `$TAG_SCAN_OUTPUT`
   - Count TAGs by type:
     - `@SPEC` TAGs: Count lines matching `@SPEC`
     - `@CODE` TAGs: Count lines matching `@CODE`
     - `@TEST` TAGs: Count lines matching `@TEST`
     - `@DOC` TAGs: Count lines matching `@DOC`
   - Print:
     ```
     🔗 TAG Chain Verification:
     - @SPEC TAGs: [count]
     - @CODE TAGs: [count]
     - @TEST TAGs: [count]
     - @DOC TAGs: [count]

     TAG chain integrity: [PASS / ISSUES]
     ```

**Result**: TAG index updated and verified.

**Next step**: Go to STEP 2.4

---

### STEP 2.4: SPEC Document Synchronization (CRITICAL)

**Your task**: Ensure that SPEC documents are updated to match code changes.

**Important note**: This step is CRITICAL. Any code changes MUST be reflected in SPEC documents to maintain specification alignment.

**Steps**:

1. **Analyze Git diff for functional changes**:
   - Execute: `git diff HEAD --unified=3`
   - Store output in variable: `$GIT_DIFF`
   - Identify functional impacts:
     - Function signature changes (parameters, return values)
     - Behavior changes (logic flow, edge cases)
     - Performance characteristics (latency, throughput)
     - External dependencies (new APIs, services)

2. **Determine which SPECs need updates**:
   - Parse changed files from `$CHANGED_FILES`
   - For each changed file in `src/`:
     - Search for `@CODE:SPEC-{ID}` TAG in file
     - Extract SPEC ID
     - Add to list: `$SPECS_TO_UPDATE`
   - Print:
     ```
     📋 SPEC Documents Requiring Synchronization:

     $SPECS_TO_UPDATE (list of SPEC IDs)
     ```

3. **For each SPEC in $SPECS_TO_UPDATE**:

   a. **Read current SPEC documents**:
      - Read file: `.moai/specs/SPEC-{ID}/spec.md`
      - Read file: `.moai/specs/SPEC-{ID}/plan.md`
      - Read file: `.moai/specs/SPEC-{ID}/acceptance.md`

   b. **Compare SPEC requirements with actual code implementation**:
      - **Your task**: Verify each SPEC requirement matches code behavior
      - Check:
        - Do EARS statements match function signatures?
        - Are acceptance criteria still valid?
        - Were edge cases discovered during implementation?
        - Did implementation strategy change from plan.md?

   c. **Identify spec-to-code divergence**:
      - IF divergence detected → Mark for update
      - Store divergence details in variable: `$SPEC_DIVERGENCE`

   d. **Update SPEC documents to match implementation**:
      - **IF spec.md needs update**:
        - Update EARS statements to match actual code behavior
        - Add edge cases discovered during implementation
        - Update @CODE TAG references if files moved
      - **IF acceptance.md needs update**:
        - Add new test cases for edge cases
        - Update acceptance criteria for behavior changes
        - Link with @TEST TAGs from test files
      - **IF plan.md needs update**:
        - Document implementation changes
        - Explain deviations from original plan
        - Update technical approach if changed

   e. **Update SPEC metadata if implementation is complete**:
      - Check if SPEC status is `draft`
      - Check if RED → GREEN → REFACTOR commits exist
      - Check if @TEST and @CODE TAGs exist
      - **IF all conditions met**:
        - Update: `status: draft` → `status: completed`
        - Update: `version: 0.0.x` → `version: 0.1.0`
        - Add HISTORY section with completion date

4. **Verify SPEC-Code alignment**:
   - For each updated SPEC:
     - **Verify**: All EARS statements match implementation
     - **Verify**: Acceptance criteria valid for current code
     - **Verify**: @CODE/@TEST TAGs point to correct locations
     - **Verify**: No spec-code divergence remains

5. **Print SPEC synchronization results**:
   ```
   ✅ SPEC Document Synchronization Complete

   SPECs updated: [count]

   Details:
   - SPEC-001: spec.md, acceptance.md updated (edge case added)
   - SPEC-002: plan.md updated (implementation approach changed)
   - SPEC-003: Status → completed (v0.1.0)

   SPEC-Code alignment: VERIFIED
   ```

6. **IF no SPECs required updates**:
   - Print:
     ```
     ℹ️ No SPEC updates needed

     All SPEC documents are aligned with current code.
     ```

**Result**: SPEC documents synchronized with code. Specification alignment maintained.

**Next step**: Go to STEP 2.5

---

### STEP 2.5: Domain-Based Sync Routing (Automatic)

**Your task**: Route documentation sync to domain-specific experts based on changed file patterns.

**Steps**:

1. **Detect domains from changed files**:
   - Read variable: `$CHANGED_FILES`
   - Initialize empty array: `$DETECTED_DOMAINS`

2. **Check for Frontend changes**:
   - Search `$CHANGED_FILES` for patterns:
     - `*.tsx`, `*.jsx`, `*.vue`
     - `src/components/*`, `src/pages/*`
   - IF any match found → Add "frontend" to `$DETECTED_DOMAINS`

3. **Check for Backend changes**:
   - Search `$CHANGED_FILES` for patterns:
     - `src/api/*`, `src/models/*`, `src/routes/*`, `src/services/*`
   - IF any match found → Add "backend" to `$DETECTED_DOMAINS`

4. **Check for DevOps changes**:
   - Search `$CHANGED_FILES` for patterns:
     - `Dockerfile`, `docker-compose.yml`
     - `.github/workflows/*`, `terraform/*`, `k8s/*`
   - IF any match found → Add "devops" to `$DETECTED_DOMAINS`

5. **Check for Database changes**:
   - Search `$CHANGED_FILES` for patterns:
     - `src/database/*`, `migrations/*`, `*.sql`, `schema/*`
   - IF any match found → Add "database" to `$DETECTED_DOMAINS`

6. **Check for Data Science changes**:
   - Search `$CHANGED_FILES` for patterns:
     - `notebooks/*`, `src/pipelines/*`, `*.ipynb`, `src/models/ml/*`
   - IF any match found → Add "datascience" to `$DETECTED_DOMAINS`

7. **Check for Mobile changes**:
   - Search `$CHANGED_FILES` for patterns:
     - `src/mobile/*`, `ios/*`, `android/*`, `*.swift`, `*.kt`
   - IF any match found → Add "mobile" to `$DETECTED_DOMAINS`

8. **IF no domains detected**:
   - Print:
     ```
     ℹ️ No domain-specific sync needed

     Using standard sync templates.
     ```
   - **Skip to STEP 2.6**

9. **IF domains detected**:
   - Print:
     ```
     🎯 Domain-specific sync routing activated

     Detected domains: $DETECTED_DOMAINS

     Generating domain-specific documentation...
     ```

10. **For each domain in $DETECTED_DOMAINS**:

    **IF domain is "frontend"**:
    - Invoke Explore agent:
      - `subagent_type`: "Explore"
      - `prompt`:
        ```
        Act as frontend-expert. Provide sync documentation.

        Changed Files: [list frontend files from $CHANGED_FILES]

        Provide frontend-specific documentation:
        1. Component documentation updates
        2. Storybook story generation (if applicable)
        3. UI architecture diagram updates
        4. Accessibility compliance notes
        5. Component prop documentation

        Output format: Markdown document
        Save to: .moai/reports/sync-frontend-$TIMESTAMP.md
        ```

    **IF domain is "backend"**:
    - Invoke Explore agent:
      - `subagent_type`: "Explore"
      - `prompt`:
        ```
        Act as backend-expert. Provide sync documentation.

        Changed Files: [list backend files from $CHANGED_FILES]

        Provide backend-specific documentation:
        1. OpenAPI spec generation/updates
        2. Schema documentation updates
        3. Error handling documentation
        4. API endpoint examples
        5. Performance characteristics

        Output format: Markdown document
        Save to: .moai/reports/sync-backend-$TIMESTAMP.md
        ```

    **IF domain is "devops"**:
    - Invoke Explore agent:
      - `subagent_type`: "Explore"
      - `prompt`:
        ```
        Act as devops-expert. Provide sync documentation.

        Changed Files: [list devops files from $CHANGED_FILES]

        Provide DevOps-specific documentation:
        1. Deployment documentation updates
        2. CI/CD pipeline changes
        3. Infrastructure diagrams
        4. Configuration management notes
        5. Monitoring/alerting setup

        Output format: Markdown document
        Save to: .moai/reports/sync-devops-$TIMESTAMP.md
        ```

    **IF domain is "database"**:
    - Invoke Explore agent:
      - `subagent_type`: "Explore"
      - `prompt`:
        ```
        Act as database-expert. Provide sync documentation.

        Changed Files: [list database files from $CHANGED_FILES]

        Provide database-specific documentation:
        1. Schema documentation updates
        2. Migration logs
        3. Query optimization notes
        4. Index strategies
        5. Data integrity constraints

        Output format: Markdown document
        Save to: .moai/reports/sync-database-$TIMESTAMP.md
        ```

    **IF domain is "datascience"**:
    - Invoke Explore agent:
      - `subagent_type`: "Explore"
      - `prompt`:
        ```
        Act as datascience-expert. Provide sync documentation.

        Changed Files: [list data science files from $CHANGED_FILES]

        Provide data science-specific documentation:
        1. Pipeline documentation
        2. Model cards (for ML models)
        3. Data validation rules
        4. Feature engineering notes
        5. Experiment tracking

        Output format: Markdown document
        Save to: .moai/reports/sync-datascience-$TIMESTAMP.md
        ```

    **IF domain is "mobile"**:
    - Invoke Explore agent:
      - `subagent_type`: "Explore"
      - `prompt`:
        ```
        Act as mobile-expert. Provide sync documentation.

        Changed Files: [list mobile files from $CHANGED_FILES]

        Provide mobile-specific documentation:
        1. Platform-specific documentation (iOS/Android)
        2. App lifecycle documentation
        3. Native module documentation
        4. Push notification setup
        5. Deep linking configuration

        Output format: Markdown document
        Save to: .moai/reports/sync-mobile-$TIMESTAMP.md
        ```

11. **Wait for all domain-specific sync agents to complete**:
    - Monitor each Explore agent execution
    - Store results in array: `$DOMAIN_SYNC_RESULTS`

12. **Create combined sync report**:
    - **Your task**: Consolidate all domain-specific reports into master report
    - Read all domain-specific report files
    - Write combined report to: `.moai/reports/sync-report-$TIMESTAMP.md`
    - Include sections:
      - Changed Files Summary
      - Domain-Specific Sync Results (with links to detailed reports)
      - @TAG Verification
      - Next Steps

13. **Print domain sync summary**:
    ```
    ✅ Domain-specific sync complete

    Domains processed:
    - Frontend: [count] files → sync-frontend-$TIMESTAMP.md
    - Backend: [count] files → sync-backend-$TIMESTAMP.md
    - DevOps: [count] files → sync-devops-$TIMESTAMP.md

    Combined report: sync-report-$TIMESTAMP.md
    ```

**Result**: Domain-specific documentation generated. Combined sync report created.

**Next step**: Go to STEP 2.6

---

### STEP 2.6: Display Synchronization Completion Report

**Your task**: Present comprehensive synchronization results to the user.

**Steps**:

1. **Read synchronization results**:
   - Read file: `.moai/reports/sync-report-$TIMESTAMP.md`
   - Store content in variable: `$SYNC_REPORT_CONTENT`

2. **Calculate synchronization metrics**:
   - Count updated files from `$SYNC_RESULTS`
   - Count TAG repairs from `$TAG_VALIDATION_RESULTS`
   - Count domain-specific reports from `$DETECTED_DOMAINS`

3. **Print comprehensive completion report**:
   ```
   ═══════════════════════════════════════════════════════
   ✅ Document Synchronization Complete
   ═══════════════════════════════════════════════════════

   📊 Synchronization Summary:
   - Mode: $MODE
   - Scope: $SYNC_SCOPE
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

4. **IF TAG issues remain**:
   - Print warning:
     ```
     ⚠️ Remaining TAG Issues:

     The following issues could not be automatically resolved:
     - Orphan @CODE TAGs: [list]
     - Broken references: [list]

     Manual review recommended.
     ```

5. **Print next step guidance**:
   ```
   🚀 Next Steps:

   1. Review sync report: .moai/reports/sync-report-$TIMESTAMP.md
   2. Commit changes: Git operations will be handled next
   3. PR handling: [if Team mode] Ready for PR transition
   ```

**Result**: User informed of synchronization results. Ready for Git operations.

**Next step**: Go to STEP 3.1

---

## 🔧 STEP 3: Git Operations & PR

### STEP 3.1: Commit Document Changes

**Your task**: Call git-manager agent to commit all document changes with proper commit message.

**Steps**:

1. **Prepare commit summary**:
   - Read variable: `$SYNC_RESULTS`
   - Create summary text:
     ```
     docs: sync documentation with code changes

     Synchronized Living Documents:
     - [list from $SYNC_RESULTS]

     TAG system updates:
     - [TAG repair count] repairs
     - TAG index updated

     SPEC synchronization:
     - [SPEC update count] SPECs updated

     Domain-specific sync:
     - [domain list if applicable]

     🤖 Generated with Claude Code

     Co-Authored-By: 🦄 Alfred@MoAI
     ```
   - Store in variable: `$COMMIT_MESSAGE`

2. **Invoke git-manager agent for commit**:
   - **Your task**: Call git-manager to commit document changes
   - Use Task tool:
     - `subagent_type`: "git-manager"
     - `description`: "Commit document synchronization changes"
     - `prompt`:
       ```
       당신은 git-manager 에이전트입니다.

       **작업**: 문서 동기화 변경사항을 Git 커밋으로 저장하세요.

       **커밋 범위**:
       - 변경된 모든 문서 파일
       - .moai/reports/ 디렉토리
       - .moai/indexes/ 디렉토리 (if changed)
       - README.md (if changed)
       - docs/ 디렉토리 (if changed)

       **커밋 메시지**:
       $COMMIT_MESSAGE

       **중요**:
       - HEREDOC 형식으로 커밋 메시지 전달
       - 모든 변경사항을 하나의 커밋으로 묶기
       - 커밋 후 성공 여부 보고

       **실행 순서**:
       1. git add (변경된 문서 파일들)
       2. git commit -m (HEREDOC)
       3. git log -1 (커밋 확인)
       ```

3. **Wait for git-manager response**:
   - Store response in variable: `$GIT_COMMIT_RESULT`

4. **Verify commit success**:
   - Execute: `git log -1 --oneline`
   - Store output in variable: `$LAST_COMMIT`
   - Print:
     ```
     ✅ Document changes committed

     Commit: $LAST_COMMIT
     ```

5. **IF commit failed**:
   - Print error:
     ```
     ❌ Error: Git commit failed

     git-manager reported an issue.
     Check git status and resolve conflicts if any.

     You can retry commit manually:
     git add .moai/reports/ docs/ README.md
     git commit -m "docs: sync documentation"
     ```
   - **STOP HERE** → Exit with error code

**Result**: Document changes committed to Git history.

**Next step**: Go to STEP 3.2

---

### STEP 3.2: (Optional) PR Ready Transition

**Your task**: Transition PR from Draft to Ready for Review (Team mode only).

**Decision point**:

1. **Check if PR transition is needed**:
   - Read variable: `$PROJECT_MODE`
   - IF `$PROJECT_MODE != "Team"` → Skip this step:
     ```
     ℹ️ PR transition skipped (Personal mode)
     ```
   - **Next step**: Go to STEP 3.5 (skip auto-merge)

2. **IF Team mode**:
   - Continue to PR transition

3. **Check if gh CLI is available**:
   - Execute: `which gh`
   - IF gh NOT found → Print warning and skip:
     ```
     ⚠️ Warning: GitHub CLI (gh) not found

     PR transition skipped.
     To enable PR features, install gh CLI:
     https://cli.github.com/
     ```
   - **Next step**: Go to STEP 3.5

4. **Get current PR number**:
   - Execute: `gh pr view --json number -q '.number'`
   - Store result in variable: `$PR_NUMBER`
   - IF command fails → Print info and skip:
     ```
     ℹ️ No PR found for current branch

     Skipping PR transition.
     You can create a PR manually with: gh pr create
     ```
   - **Next step**: Go to STEP 3.5

5. **Get current PR status**:
   - Execute: `gh pr view $PR_NUMBER --json isDraft -q '.isDraft'`
   - Store result in variable: `$IS_DRAFT`

6. **IF PR is already Ready** (`$IS_DRAFT = false`):
   - Print:
     ```
     ℹ️ PR already in Ready state

     PR #$PR_NUMBER: Ready for Review
     ```
   - **Next step**: Go to STEP 3.3 (check auto-merge)

7. **IF PR is Draft** (`$IS_DRAFT = true`):
   - Transition PR to Ready:
     - Execute: `gh pr ready $PR_NUMBER`
     - Store exit code in variable: `$PR_READY_EXIT`

8. **Verify PR transition success**:
   - IF `$PR_READY_EXIT != 0` → Print error:
     ```
     ❌ Error: PR transition failed

     Unable to transition PR #$PR_NUMBER to Ready.
     Check PR status manually: gh pr view $PR_NUMBER
     ```
   - ELSE → Print success:
     ```
     ✅ PR transitioned to Ready for Review

     PR #$PR_NUMBER: Ready for Review

     View PR: gh pr view $PR_NUMBER --web
     ```

9. **Assign reviewers** (if configured):
   - Read `.moai/config.json` → Extract `github.reviewers`
   - IF reviewers configured:
     - Execute: `gh pr edit $PR_NUMBER --add-reviewer [reviewer-list]`

10. **Add labels** (if configured):
    - Read `.moai/config.json` → Extract `github.pr_labels`
    - IF labels configured:
      - Execute: `gh pr edit $PR_NUMBER --add-label [label-list]`

**Result**: PR transitioned to Ready for Review (Team mode only).

**Next step**: Go to STEP 3.3

---

### STEP 3.3: (Optional) PR Auto-Merge

**Your task**: Automatically merge PR and clean up branch (if --auto-merge flag is set).

**Decision point**:

1. **Check if auto-merge is requested**:
   - Read variable: `$AUTO_MERGE`
   - IF `$AUTO_MERGE = false` → Skip this step:
     ```
     ℹ️ Auto-merge not requested

     PR is Ready for manual review and merge.
     ```
   - **Next step**: Go to STEP 3.5

2. **IF auto-merge requested** (`$AUTO_MERGE = true`):
   - Print:
     ```
     🤖 Auto-merge activated

     Checking PR status and CI/CD...
     ```

3. **Check CI/CD status**:
   - Execute: `gh pr checks $PR_NUMBER`
   - Store exit code in variable: `$CI_STATUS`
   - IF `$CI_STATUS != 0` → Print error and abort:
     ```
     ❌ Error: CI/CD checks not passing

     Cannot auto-merge PR #$PR_NUMBER

     Check CI/CD status: gh pr checks $PR_NUMBER

     Please wait for checks to pass or merge manually.
     ```
   - **STOP HERE** → Go to STEP 3.5

4. **Check for merge conflicts**:
   - Execute: `gh pr view $PR_NUMBER --json mergeable -q '.mergeable'`
   - Store result in variable: `$MERGEABLE`
   - IF `$MERGEABLE != "MERGEABLE"` → Print error and abort:
     ```
     ❌ Error: PR has merge conflicts

     Cannot auto-merge PR #$PR_NUMBER

     Resolve conflicts manually:
     git fetch origin develop
     git merge origin/develop
     ```
   - **STOP HERE** → Go to STEP 3.5

5. **Execute auto-merge**:
   - Print:
     ```
     🚀 Merging PR #$PR_NUMBER...
     ```
   - Execute: `gh pr merge $PR_NUMBER --squash --delete-branch`
   - Store exit code in variable: `$MERGE_EXIT`

6. **Verify merge success**:
   - IF `$MERGE_EXIT != 0` → Print error:
     ```
     ❌ Error: PR merge failed

     Auto-merge encountered an issue.
     Merge manually: gh pr merge $PR_NUMBER --squash
     ```
   - ELSE → Print success:
     ```
     ✅ PR merged successfully

     PR #$PR_NUMBER: Merged to develop
     Remote feature branch: Deleted
     ```

**Result**: PR merged and remote branch deleted (if auto-merge succeeded).

**Next step**: Go to STEP 3.4

---

### STEP 3.4: (Optional) Branch Cleanup

**Your task**: Clean up local branches and checkout develop branch (if auto-merge succeeded).

**Decision point**:

1. **Check if branch cleanup is needed**:
   - Read variable: `$MERGE_EXIT`
   - IF `$MERGE_EXIT != 0` OR `$AUTO_MERGE = false` → Skip cleanup:
     ```
     ℹ️ Branch cleanup skipped
     ```
   - **Next step**: Go to STEP 3.5

2. **IF merge succeeded**:
   - Continue to branch cleanup

3. **Get current branch name**:
   - Execute: `git rev-parse --abbrev-ref HEAD`
   - Store result in variable: `$CURRENT_BRANCH`

4. **Checkout develop branch**:
   - Execute: `git checkout develop`
   - Store exit code in variable: `$CHECKOUT_EXIT`
   - IF `$CHECKOUT_EXIT != 0` → Print error:
     ```
     ⚠️ Warning: Could not checkout develop

     Staying on current branch: $CURRENT_BRANCH
     ```
   - ELSE → Print:
     ```
     ✅ Checked out develop branch
     ```

5. **Synchronize with remote**:
   - Execute: `git pull origin develop`
   - Print:
     ```
     ✅ Develop branch synchronized with remote
     ```

6. **Delete local feature branch**:
   - Execute: `git branch -d $CURRENT_BRANCH`
   - Store exit code in variable: `$DELETE_EXIT`
   - IF `$DELETE_EXIT != 0` → Print warning:
     ```
     ⚠️ Warning: Could not delete local branch

     Branch: $CURRENT_BRANCH

     Delete manually if needed: git branch -D $CURRENT_BRANCH
     ```
   - ELSE → Print:
     ```
     ✅ Local feature branch deleted

     Branch: $CURRENT_BRANCH
     ```

7. **Check for auto-delete-branches config**:
   - Read `.moai/config.json` → Extract `github.auto_delete_branches`
   - IF configured as `true`:
     - Print:
       ```
       ℹ️ Remote branch already deleted by auto-merge
       ```

8. **Print branch cleanup summary**:
   ```
   🧹 Branch Cleanup Complete

   - Current branch: develop
   - Deleted local branch: $CURRENT_BRANCH
   - Remote branch: Deleted

   🎉 Ready for next feature!

   Start next work with: /alfred:1-plan "feature description"
   ```

**Result**: Branches cleaned up. Repository ready for next development cycle.

**Next step**: Go to STEP 3.5

---

### STEP 3.5: Display Completion Report & Next Steps

**Your task**: Present final completion report and ask user what to do next.

**Steps**:

1. **Determine workflow completion status**:
   - Read variables: `$PROJECT_MODE`, `$AUTO_MERGE`, `$MERGE_EXIT`
   - Determine completion type:
     - IF `$AUTO_MERGE = true` AND `$MERGE_EXIT = 0` → Status: "Full Auto-Merge Complete"
     - ELSE IF `$PROJECT_MODE = "Team"` → Status: "PR Ready for Review"
     - ELSE → Status: "Personal Mode Checkpoint"

2. **Print final completion report**:

   **IF Full Auto-Merge Complete**:
   ```
   ═══════════════════════════════════════════════════════
   🎉 MoAI-ADK Workflow Complete (Full Auto-Merge)
   ═══════════════════════════════════════════════════════

   ✅ Completed Steps:
   1. Document synchronization
   2. TAG system verification
   3. Git commit
   4. PR merge to develop
   5. Branch cleanup

   📍 Current Status:
   - Branch: develop
   - PR #$PR_NUMBER: Merged
   - Local feature branch: Deleted
   - Remote feature branch: Deleted

   🚀 Ready for next feature!

   ═══════════════════════════════════════════════════════
   ```

   **IF PR Ready for Review**:
   ```
   ═══════════════════════════════════════════════════════
   ✅ MoAI-ADK Workflow Complete (PR Ready)
   ═══════════════════════════════════════════════════════

   ✅ Completed Steps:
   1. Document synchronization
   2. TAG system verification
   3. Git commit
   4. PR transition to Ready

   📍 Current Status:
   - PR #$PR_NUMBER: Ready for Review
   - Branch: $CURRENT_BRANCH

   ⏳ Next Steps:
   - Review PR: gh pr view $PR_NUMBER --web
   - Merge after approval: gh pr merge $PR_NUMBER --squash

   ═══════════════════════════════════════════════════════
   ```

   **IF Personal Mode**:
   ```
   ═══════════════════════════════════════════════════════
   ✅ Document Synchronization Complete (Personal Mode)
   ═══════════════════════════════════════════════════════

   ✅ Completed Steps:
   1. Document synchronization
   2. TAG system verification
   3. Git commit (checkpoint)

   📍 Current Status:
   - Branch: $CURRENT_BRANCH
   - Changes committed locally

   💡 Personal mode workflow:
   - Continue development on current branch
   - OR merge to main manually when ready

   ═══════════════════════════════════════════════════════
   ```

3. **Ask user for next action using AskUserQuestion**:
   - **Your task**: Use the AskUserQuestion tool to gather user's next step
   - Tool call:
     - `questions`: Array with 1 question
     - Question details:
       - `question`: "Documentation synchronization complete. What would you like to do next?"
       - `header`: "Next Steps"
       - `multiSelect`: false
       - `options`: Array with 3 choices (context-dependent):

         **IF auto-merge completed**:
         1. Label: "📋 Create Next SPEC", Description: "Start new feature planning with /alfred:1-plan"
         2. Label: "🔄 Start New Session", Description: "Execute /clear for fresh session (recommended for performance)"
         3. Label: "🎯 Project Overview", Description: "Review project status and documentation"

         **IF PR Ready (not auto-merged)**:
         1. Label: "📋 Create Next SPEC", Description: "Start new feature planning with /alfred:1-plan"
         2. Label: "📤 Review PR", Description: "View PR and prepare for manual merge"
         3. Label: "🔄 Start New Session", Description: "Execute /clear for fresh session (recommended for performance)"

         **IF Personal Mode**:
         1. Label: "📋 Create Next SPEC", Description: "Start new feature planning with /alfred:1-plan"
         2. Label: "🔧 Continue Development", Description: "Keep working on current branch"
         3. Label: "🔄 Start New Session", Description: "Execute /clear for fresh session (recommended for performance)"

4. **Wait for user response**:
   - Store response in variable: `$NEXT_ACTION`
   - Read value from: `$NEXT_ACTION["0"]`

5. **Process user response**:

   **IF user chose "📋 Create Next SPEC"**:
   - Print:
     ```
     🚀 Ready to start next feature

     Use: /alfred:1-plan "feature description"

     This will create a new SPEC and start a new development cycle.
     ```

   **IF user chose "🔄 Start New Session"**:
   - Print:
     ```
     🔄 Starting fresh session

     Recommendation: Use /clear or /new command to reset context.

     This improves performance for the next development cycle.
     ```

   **IF user chose "📤 Review PR"** (Team mode):
   - Print:
     ```
     📤 Opening PR for review

     PR #$PR_NUMBER: Ready for Review

     View in browser: gh pr view $PR_NUMBER --web

     After approval, merge with: gh pr merge $PR_NUMBER --squash
     ```

   **IF user chose "🔧 Continue Development"** (Personal mode):
   - Print:
     ```
     🔧 Continuing development

     Current branch: $CURRENT_BRANCH

     Your changes are committed. Continue implementing features.
     ```

   **IF user chose "🎯 Project Overview"**:
   - Print:
     ```
     🎯 Project Overview

     Sync reports: .moai/reports/
     SPEC documents: .moai/specs/
     TAG index: .moai/indexes/tags.db

     Quick status check: /alfred:3-sync status
     ```

6. **Print final message**:
   ```
   ✨ Thank you for using MoAI-ADK!

   For more information:
   - Skill("moai-alfred-dev-guide")
   - CLAUDE.md in project root
   ```

**Result**: User informed of completion. Next action suggested.

**Next step**: Command complete. Exit.

---

## 🚨 STEP 4: Graceful Exit (User Aborted or Modified)

**Your task**: Handle user abortion or modification requests gracefully.

**This step is reached when**:
- User chose "❌ Abort" in STEP 1.6
- User chose "🔄 Request Modifications" in STEP 1.6

**Steps**:

1. **Read user decision**:
   - Read variable: `$USER_DECISION["0"]`

2. **IF user chose "❌ Abort"**:
   - Print abort message:
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
   - **Exit command** with exit code 0

3. **IF user chose "🔄 Request Modifications"**:
   - Print:
     ```
     🔄 Re-analyzing with modifications...
     ```
   - **Go back to STEP 1.5** with `$MODIFICATION_REQUEST` applied

**Result**: Command exits gracefully or re-runs analysis.

**Next step**: Exit or STEP 1.5

---

## 📚 Mode-Specific Execution Summary

### Auto Mode (default)

**Command**: `/alfred:3-sync` or `/alfred:3-sync auto`

**Behavior**:
- Smart selective synchronization based on Git changes
- Only changed files and related documents are updated
- Quick execution for daily workflow
- PR Ready transition in Team mode

**Best for**: Daily development workflow

---

### Force Mode

**Command**: `/alfred:3-sync force`

**Behavior**:
- Full project re-synchronization
- All documents regenerated
- All TAG chains re-verified
- Longer execution time

**Best for**: Error recovery, major refactoring, periodic full sync

---

### Status Mode

**Command**: `/alfred:3-sync status`

**Behavior**:
- Quick status check only
- No synchronization performed
- Reports TAG system health
- Lists changed files since last sync

**Best for**: Quick health check before starting work

---

### Project Mode

**Command**: `/alfred:3-sync project`

**Behavior**:
- Integrated project-wide synchronization
- README.md updated with full feature list
- docs/architecture.md updated
- docs/api/ unified
- .moai/indexes/ rebuilt

**Best for**: Milestone completion, periodic integrated sync

---

## 🏗️ Agent Collaboration Architecture

### Separation of Concerns

**doc-syncer Agent** (STEP 2):
- Living Document synchronization
- @TAG system updates
- SPEC-Code alignment
- Domain-specific routing
- Sync report generation

**git-manager Agent** (STEP 3):
- Git commit operations
- PR status transitions
- PR auto-merge (if requested)
- Branch cleanup
- GitHub CLI integration

**Single Responsibility Principle**:
- doc-syncer does NOT touch Git
- git-manager does NOT touch documents
- Clear handoff between STEP 2 and STEP 3

---

## 🔗 TAG System Verification

### Direct Code Scanning

**Method**: `rg '@TAG' -n src/ tests/`

**Why**: Direct code scanning ensures accurate TAG counting without relying on index files.

**Verification points**:
- @SPEC TAG locations (.moai/specs/)
- @TEST TAG locations (tests/)
- @CODE TAG locations (src/)
- @DOC TAG locations (docs/)

**Integrity checks**:
- Orphan @CODE TAGs (no matching @SPEC)
- Orphan @SPEC TAGs (no matching @CODE)
- Broken references
- Duplicate TAGs

---

## 🎯 Integration with MoAI-ADK Workflow

### 4-Step Workflow Position

This command is **STEP 4** (Report & Commit):

1. **/alfred:1-plan** → SPEC creation
2. **/alfred:2-run** → TDD implementation
3. **/alfred:3-sync** → **Document sync + PR (this command)**
4. Cycle complete → Start new SPEC

### Conditional Report Generation

**Follows CLAUDE.md guidance**:
- Reports generated ONLY when explicitly requested
- Sync report is REQUIRED (always generated)
- Domain-specific reports OPTIONAL (only if domains detected)

---

## ⚙️ Environment Dependencies

**Required**:
- Git repository
- MoAI-ADK project structure (.moai/, .claude/)
- Python3 (for TAG verification scripts)

**Optional**:
- gh CLI (for GitHub PR integration)
- Domain-specific experts (for domain routing)

**Graceful degradation**:
- Works without gh CLI (skips PR features)
- Works without Python3 (limited TAG checks)
- Works without domain experts (uses generic templates)

---

## 🎓 Best Practices

### When to Use Each Mode

**Use auto mode** (default):
- Daily development workflow
- Single SPEC implementation complete
- Quick sync after code changes

**Use force mode**:
- After major refactoring
- Error recovery (TAG system broken)
- Periodic full re-sync (weekly/monthly)

**Use status mode**:
- Before starting work (health check)
- Quick overview of sync needs
- No changes to files

**Use project mode**:
- Milestone completion
- Release preparation
- Integrated documentation update

### Performance Tips

**Use --auto-merge for fast iteration**:
- Automatically merges PR
- Cleans up branches
- Returns to develop
- Ready for next /alfred:1-plan immediately

**Start new session after sync**:
- Use /clear or /new command
- Reduces context size
- Improves performance for next cycle

---

## 🔍 Troubleshooting

### TAG Verification Issues

**Problem**: Orphan TAGs detected

**Solution**:
1. Review sync report: `.moai/reports/sync-report-latest.md`
2. Manually fix orphan TAGs in code
3. Re-run: `/alfred:3-sync force`

### PR Transition Failed

**Problem**: gh CLI error

**Solution**:
1. Check gh authentication: `gh auth status`
2. Verify PR exists: `gh pr view`
3. Manual transition: `gh pr ready [PR_NUMBER]`

### Merge Conflicts

**Problem**: Cannot auto-merge due to conflicts

**Solution**:
1. Fetch latest: `git fetch origin develop`
2. Merge develop: `git merge origin/develop`
3. Resolve conflicts
4. Re-run: `/alfred:3-sync`

---

## 📖 Related Documentation

**Skills**:
- `Skill("moai-alfred-tag-scanning")` - TAG system details
- `Skill("moai-alfred-git-workflow")` - Git operations
- `Skill("moai-alfred-trust-validation")` - Quality gates
- `Skill("moai-alfred-ask-user-questions")` - TUI interactions

**Workflows**:
- CLAUDE.md - Alfred 4-step workflow
- `.moai/docs/workflows/alfred-3-sync.md` - Detailed workflow

**Configuration**:
- `.moai/config.json` - Project settings
- `.claude/settings.json` - Claude Code settings

---

**Version**: 3.0.0 (Fully Imperative)
**Last Updated**: 2025-01-04
**Pattern**: Pure command-driven, zero Python pseudo-code, step-by-step execution flow
