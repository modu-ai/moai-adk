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
> **Note**: Interactive prompts use `AskUserQuestion tool (documented in moai-alfred-interactive-questions skill)` for TUI selection menus. The skill is loaded on-demand when user interaction is required.
>
> **Batched Design**: All AskUserQuestion calls follow batched design principles (1-4 questions per call) to minimize user interaction turns. See CLAUDE.md section "Alfred Command Completion Pattern" for details.

<!-- @CODE:ALF-WORKFLOW-003:CMD-SYNC -->

**4-Step Workflow Integration**: This command implements Step 4 of Alfred's workflow (Report & Commit with conditional report generation). See CLAUDE.md for full workflow details.

## 🚀 START HERE

**CRITICAL**: Load the TUI Survey Skill FIRST before any user interaction:

```
AskUserQuestion tool (documented in moai-alfred-interactive-questions skill)
```

This Skill MUST be loaded at the very beginning to enable TUI menu rendering for AskUserQuestion calls throughout this workflow.

## 🎯 Command Purpose

Synchronize code changes to Living Documents and verify @TAG system to ensure complete traceability.

**Document sync to**: $ARGUMENTS

> **Standard two-step workflow** (see `CLAUDE.md` - "Alfred Command Execution Pattern" for details)

## 📋 Execution flow

**Phase 0: Skill Loading** (IMMEDIATE)
- Load `AskUserQuestion tool (documented in moai-alfred-interactive-questions skill)` at the very start
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

STEP 1 consists of **two independent phases** to provide flexible workflow based on project complexity:

### 📋 STEP 1 Workflow Overview

```
┌─────────────────────────────────────────────────────────────┐
│ STEP 1: Synchronization Analysis & Planning                │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Phase A (OPTIONAL)                                         │
│  ┌─────────────────────────────────────────┐               │
│  │ 🔍 Explore Agent                        │               │
│  │ • Navigate complex TAG chains           │               │
│  │ • Scan entire TAG system                │               │
│  │ • Identify orphan TAGs                  │               │
│  └─────────────────────────────────────────┘               │
│                    ↓                                        │
│          (exploration results)                              │
│                    ↓                                        │
│  Phase B (REQUIRED)                                         │
│  ┌─────────────────────────────────────────┐               │
│  │ ⚙️ tag-agent + doc-syncer Agents        │               │
│  │ • Verify TAG integrity (full project)   │               │
│  │ • Analyze Git changes                   │               │
│  │ • Create synchronization plan           │               │
│  │ • Request user approval                 │               │
│  └─────────────────────────────────────────┘               │
│                    ↓                                        │
│          (user approval via AskUserQuestion)                │
│                    ↓                                        │
│              PROCEED TO STEP 2                              │
└─────────────────────────────────────────────────────────────┘
```

**Key Points**:
- **Phase A is optional** - Skip for simple single-SPEC changes
- **Phase B is required** - Always runs to verify TAGs and plan sync
- **Results flow forward** - Exploration results (if any) are passed to tag-agent
- **⚠️ Important**: tag-agent verifies ENTIRE PROJECT, not just changed files

---

### 🔍 Phase A: TAG Chain Navigation (OPTIONAL)

**Use the Explore agent for complex or extensive TAG chains.**

#### When to use Phase A:

- ✅ Large projects (100+ files)
- ✅ Need comprehensive TAG chain integrity verification
- ✅ Changes span multiple SPECs or modules
- ❌ Simple changes to a single SPEC (skip to Phase B)

#### How to invoke Explore agent:

```
Invoking the Task tool (Explore agent):
- subagent_type: "Explore"
- description: "Scan entire TAG system"
- prompt: "프로젝트 전체에서 @TAG 시스템을 스캔해주세요:
 - @SPEC TAG 위치 (.moai/specs/)
 - @TEST TAG 위치 (tests/)
 - @CODE TAG 위치 (src/)
 - @DOC TAG 위치 (docs/)
 - 고아 TAG 및 끊긴 참조 감지
 상세도 수준: very thorough"
```

**Note**: For simple changes, skip Phase A and proceed directly to Phase B.

---

### ⚙️ Phase B: TAG Verification & Sync Planning (REQUIRED)

**Call tag-agent and doc-syncer to verify TAG integrity and plan synchronization.**

This phase is **always required** and runs **two agents sequentially**:

#### How to invoke agents:

```
1. Tag-agent call (TAG verification - FULL PROJECT SCOPE):
   - subagent_type: "tag-agent"
- description: "Verify TAG system across entire project"
 - prompt: "전체 프로젝트에서 포괄적인 @TAG 시스템 검증을 수행해주세요.

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

 **출력 형식**: 고아 TAG의 전체 목록을 위치와 함께 제공합니다.

 (선택사항) 탐색 결과: $EXPLORE_RESULTS"

2. doc-syncer call (synchronization plan):
   - subagent_type: "doc-syncer"
   - description: "Establish a document synchronization plan"
   - prompt: """당신은 doc-syncer 에이전트입니다.

언어 설정:
- 대화_언어: {{CONVERSATION_LANGUAGE}}
- 언어명: {{CONVERSATION_LANGUAGE_NAME}}

중요 지시사항:
문서 업데이트는 대화_언어를 반드시 존중해야 합니다:
- 사용자 대면 문서 (README, 가이드): {{CONVERSATION_LANGUAGE}}
- SPEC 문서 (spec.md, plan.md, acceptance.md): {{CONVERSATION_LANGUAGE}}
- 코드 주석: {{CONVERSATION_LANGUAGE}} (기술 키워드 제외)
- 기술 문서 및 YAML 프론트매터: 영어

스킬 호출:
필요 시 명시적 Skill() 호출 사용:
- Skill("moai-foundation-tags") - TAG 체인 검증
- Skill("moai-foundation-trust") - 품질 게이트 검사
- Skill("moai-alfred-tag-scanning") - TAG 인벤토리 업데이트

작업:
Git 변경사항을 분석하고 문서 동기화 계획을 수립해주세요.
모든 문서 업데이트가 대화_언어 설정과 일치하는지 확인합니다.

$ARGUMENTS
(선택사항) TAG 검증 결과: $TAG_VALIDATION_RESULTS"""
```

**Note**:
- **Sequential execution**: Run tag-agent first, then doc-syncer
- **Results flow**: TAG validation results from tag-agent are passed to doc-syncer via `$TAG_VALIDATION_RESULTS`
- **Phase A results**: If Phase A was executed, exploration results are passed to tag-agent via `$EXPLORE_RESULTS`

---

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

### 🎯 Synchronization Plan Approval (DECISION POINT 1)

After completing synchronization analysis and establishing a plan, Alfred invokes AskUserQuestion to gather user approval:

```python
AskUserQuestion(
    questions=[
        {
            "question": "Synchronization plan is ready. How would you like to proceed?",
            "header": "Plan Approval",
            "multiSelect": false,
            "options": [
                {
                    "label": "✅ Proceed with Sync",
                    "description": "Execute document synchronization as planned"
                },
                {
                    "label": "🔄 Request Modifications",
                    "description": "Specify changes to the synchronization strategy"
                },
                {
                    "label": "🔍 Review Details",
                    "description": "Re-examine TAG validation results and changes"
                },
                {
                    "label": "❌ Abort",
                    "description": "Cancel synchronization, keep current state"
                }
            ]
        }
    ]
)
```

**Response Processing**:
- **✅ Proceed with Sync** (`answers["0"] === "Proceed"`) → Execute Phase 2 (document synchronization)
- **🔄 Request Modifications** (`answers["0"] === "Modifications"`) → Collect feedback and re-analyze
- **🔍 Review Details** (`answers["0"] === "Review"`) → Display TAG validation results, then re-present decision
- **❌ Abort** (`answers["0"] === "Abort"`) → Stop synchronization, maintain current branches

---

## 🚀 STEP 2: Execute document synchronization (after user approval)

After user approval (collected via `AskUserQuestion tool (documented in moai-alfred-interactive-questions skill)`), the doc-syncer agent performs **Living Document synchronization and @TAG updates**, and optionally executes PR Ready transitions only in team mode.

---

### 2.0.5 Domain-Based Sync Routing (Automatic - During Sync)

**Purpose**: Route documentation sync to domain-specific experts based on changed files

**When to run**: During doc-syncer execution, after analyzing git changes

**Detection Logic**:

Alfred analyzes changed files to determine which domains were modified:

```bash
# Get list of changed files
git diff --name-only HEAD~1 HEAD

# Domain detection by file patterns
```

**File Pattern → Domain Mapping**:

| File Patterns | Domain | Sync Focus |
|---------------|--------|-----------|
| `src/components/*`, `src/pages/*`, `*.tsx`, `*.jsx`, `*.vue` | Frontend | Component documentation, Storybook, UI architecture |
| `src/api/*`, `src/models/*`, `src/routes/*`, `src/services/*` | Backend | API documentation (OpenAPI), schema docs, error handling |
| `Dockerfile`, `docker-compose.yml`, `.github/workflows/*`, `terraform/*`, `k8s/*` | DevOps | Deployment docs, CI/CD status, infrastructure diagrams |
| `src/database/*`, `migrations/*`, `*.sql`, `schema/*` | Database | Schema documentation, migration logs, query optimization |
| `notebooks/*`, `src/pipelines/*`, `*.ipynb`, `src/models/ml/*` | Data Science | Pipeline documentation, model cards, data validation |
| `src/mobile/*`, `ios/*`, `android/*`, `*.swift`, `*.kt` | Mobile | Platform-specific docs, app lifecycle, native modules |

**Automatic Invocation Pattern**:

```python
# Pseudo-code for domain detection from git changes
changed_files = git_diff("HEAD~1", "HEAD")

detected_domains = []
if any(".tsx" in f or ".jsx" in f or "src/components/" in f for f in changed_files):
    detected_domains.append("frontend")
if any("src/api/" in f or "src/models/" in f or "src/routes/" in f for f in changed_files):
    detected_domains.append("backend")
if any("Dockerfile" in f or ".github/workflows/" in f or "docker-compose" in f for f in changed_files):
    detected_domains.append("devops")
# ... repeat for all domains

# Invoke domain-specific sync for each detected domain
for domain in detected_domains:
    Task(
        subagent_type="Explore",
        prompt=f"Generate domain-specific sync report for {domain} changes"
    )
```

**Example Invocation** (Frontend + Backend detected):

```python
# Frontend sync
Task(
    subagent_type="Explore",
    prompt="""You are frontend-expert providing sync documentation.

Changed Files: [src/components/Dashboard.tsx, src/pages/Home.vue]

Provide frontend-specific documentation:
1. Component documentation updates
2. Storybook story generation
3. UI architecture diagram updates
4. Accessibility compliance notes

Output format: Markdown document for .moai/reports/sync-frontend.md"""
)

# Backend sync
Task(
    subagent_type="Explore",
    prompt="""You are backend-expert providing sync documentation.

Changed Files: [src/api/auth.py, src/models/user.py, src/routes/users.py]

Provide backend-specific documentation:
1. OpenAPI spec generation
2. Schema documentation updates
3. Error handling documentation
4. API endpoint examples

Output format: Markdown document for .moai/reports/sync-backend.md"""
)
```

**Output Storage Structure**:

```
.moai/reports/
├── sync-report-2025-10-23.md          # Combined sync report
├── sync-frontend-2025-10-23.md        # Frontend-specific sync
├── sync-backend-2025-10-23.md         # Backend-specific sync
└── sync-devops-2025-10-23.md          # DevOps-specific sync
```

**Combined Sync Report Format**:

```markdown
## 📚 Documentation Sync Report - 2025-10-23

### Changed Files Summary
- Frontend: 3 files (components, pages)
- Backend: 5 files (api, models, routes)
- DevOps: 1 file (Dockerfile)

### Domain-Specific Sync Results

#### 🎨 Frontend Sync
- ✅ Component documentation: Dashboard.tsx documented
- ✅ Storybook stories: 2 stories generated
- ✅ UI architecture: Component hierarchy diagram updated
- 📄 Details: [sync-frontend-2025-10-23.md](./sync-frontend-2025-10-23.md)

#### ⚙️ Backend Sync
- ✅ OpenAPI spec: /api/auth endpoints documented
- ✅ Schema documentation: User model fields updated
- ✅ Error handling: 401/403 response examples added
- 📄 Details: [sync-backend-2025-10-23.md](./sync-backend-2025-10-23.md)

#### 🚀 DevOps Sync
- ✅ Dockerfile: Multi-stage build documented
- ✅ Deployment: Railway configuration updated
- 📄 Details: [sync-devops-2025-10-23.md](./sync-devops-2025-10-23.md)

### @TAG Verification
- ✅ All changed files have @TAG references
- ✅ SPEC → CODE → TEST → DOC chain intact

### Next Steps
- Review domain-specific sync reports
- Update README.md with new features
- Create PR for documentation changes
```

**Integration with doc-syncer**:

```python
# doc-syncer orchestrates domain-specific sync
Task(
    subagent_type="doc-syncer",
    prompt="""You are doc-syncer agent.

DOMAIN SYNC RESULTS:
{domain_sync_results}

Consolidate all domain-specific sync reports into master sync report.
Ensure @TAG chain integrity across all domains.
Update .moai/reports/sync-report-{date}.md

$ARGUMENTS"""
)
```

**Graceful Degradation**:
- If no domain detected → Standard sync (no domain-specific reports)
- If domain expert unavailable → Use generic sync templates
- Multi-domain changes → Generate separate reports, combine into master

---

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

---

### Phase 2-1: SPEC Document Synchronization (CRITICAL)

**IMPORTANT**: Any code or file changes MUST be reflected in SPEC documents to maintain specification alignment.

#### When to synchronize SPEC documents:

1. **After code modifications**:
   - Functional changes to implemented features
   - Bug fixes that alter expected behavior
   - Performance optimizations with observable changes
   - API/function signature changes
   - New dependencies or external integrations

2. **After requirement clarifications**:
   - Acceptance criteria refinements
   - Edge case discoveries during implementation
   - User feedback incorporation
   - Security/compliance adjustments

3. **After structural changes**:
   - File organization or module restructuring
   - New configuration options
   - Breaking API changes
   - Database schema modifications

#### SPEC documents requiring update:

All files in `.moai/specs/SPEC-{ID}/` must be synchronized:

- **spec.md**: Update EARS requirements if implementation differs from specification
- **plan.md**: Revise implementation strategy if approach changed
- **acceptance.md**: Update acceptance criteria if new test cases or edge cases discovered

#### Synchronization rules:

**Code ↔ SPEC Comparison**:
```
1. Review Git diff for changed files
2. Identify functional impacts:
   ├─ Signature changes (parameters, return values)
   ├─ Behavior changes (logic flow, edge cases)
   ├─ Performance characteristics (latency, throughput changes)
   └─ External dependencies (new APIs, services)
3. Map changes to SPEC requirements:
   ├─ Verify each changed function matches EARS statement
   ├─ Check if acceptance criteria still valid
   └─ Identify any spec-to-code divergence
4. Update SPEC documents:
   ├─ Correct EARS statements to match actual implementation
   ├─ Add discovered edge cases to acceptance criteria
   ├─ Update plan.md with implementation changes
   └─ Maintain TAG references (@SPEC, @CODE, @TEST consistency)
```

#### Example: When synchronization is needed

**Scenario 1: Bug Fix Changes Behavior**
```
Git change: Fixed database connection retry logic
- Was: Max 3 retries with 1-second delay
- Now: Max 5 retries with exponential backoff

SPEC update required:
- spec.md: Update EARS statement for retry behavior
- acceptance.md: Add test case for exponential backoff
- Update @CODE TAG location if function moved
```

**Scenario 2: API Signature Changes**
```
Git change: Refactored authentication function signature
- Was: validate_token(token: str) -> bool
- Now: validate_token(token: str, ttl: int = 3600) -> dict

SPEC update required:
- spec.md: Update function requirements for new TTL parameter
- acceptance.md: Add test cases for TTL validation
- plan.md: Document reason for signature change
```

**Scenario 3: New Edge Cases Discovered**
```
Git change: Added null-check validation during testing
- Discovered: Special handling needed for empty strings

SPEC update required:
- spec.md: Add EARS statement for empty string edge case
- acceptance.md: Add test case for empty string handling
- Link with @TEST TAG from test file
```

#### SPEC-Code Divergence Detection:

**Anti-pattern: Code without matching SPEC**
```
❌ WRONG: Code changes exist but SPEC documents unchanged
- Function behavior diverges from specification
- Acceptance criteria becomes inaccurate
- @TAG chain breaks (CODE exists without matching SPEC reference)

✅ CORRECT: Code changes synchronized to SPEC
- SPEC documents updated to match implementation
- All EARS statements verified against actual code
- @TAG chain maintained: SPEC ↔ CODE ↔ TEST ↔ DOC
```

#### SPEC Synchronization Checklist (doc-syncer responsibility):

Before marking sync as complete:
- [ ] All changed code files reviewed against SPEC
- [ ] EARS statements match implementation behavior
- [ ] Acceptance criteria valid for current code
- [ ] Edge cases discovered during implementation added to SPEC
- [ ] @CODE/@TEST TAGs point to correct locations
- [ ] @SPEC TAG references updated if files reorganized
- [ ] HISTORY section updated if version changed
- [ ] No spec-code divergence remains

---

## function

- **Automatic Document Synchronization**: The doc-syncer agent performs Living Document synchronization and @TAG updates. Optionally implements the PR Ready transition only in team mode.
- **SPEC-Code Alignment**: doc-syncer verifies that SPEC documents match implemented code and updates them when changes are detected.

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

## 🎯 PR Merge Strategy Selection (DECISION POINT 2)

After document synchronization completes successfully, Alfred checks team mode and invokes AskUserQuestion for PR merge strategy:

```python
AskUserQuestion(
    questions=[
        {
            "question": "Document synchronization complete. How would you like to handle the PR?",
            "header": "PR Merge Strategy",
            "multiSelect": false,
            "options": [
                {
                    "label": "🤖 Auto-Merge (Recommended)",
                    "description": "Automatically merge PR and clean up branch (team mode)"
                },
                {
                    "label": "📋 Manual Review",
                    "description": "Keep PR in Ready state for manual review and merge"
                },
                {
                    "label": "✏️ Keep as Draft",
                    "description": "Maintain PR in draft state for further refinement"
                },
                {
                    "label": "🔄 New Cycle",
                    "description": "Save changes and start new feature (skip PR for now)"
                }
            ]
        }
    ]
)
```

**Response Processing**:
- **🤖 Auto-Merge** (`answers["0"] === "Auto-Merge"`) → Execute PR auto-merge with CI checks, update develop branch
- **📋 Manual Review** (`answers["0"] === "Manual"`) → Transition PR to Ready state, notify reviewers
- **✏️ Keep as Draft** (`answers["0"] === "Draft"`) → Leave PR in draft, ready for refinement
- **🔄 New Cycle** (`answers["0"] === "New"`) → Skip PR handling, proceed to next feature planning

---

## Final Step

After PR strategy is confirmed, Alfred invokes AskUserQuestion to ask the user what to do next:

```python
AskUserQuestion(
    questions=[
        {
            "question": "Documentation synchronization complete. What would you like to do next?",
            "header": "Next Steps",
            "multiSelect": false,
            "options": [
                {
                    "label": "📋 Create Next SPEC",
                    "description": "Start new feature planning with /alfred:1-plan"
                },
                {
                    "label": "📤 Merge PR",
                    "description": "Review and merge PR to develop branch"
                },
                {
                    "label": "🔄 Start New Session",
                    "description": "Execute /clear for fresh session (recommended for performance)"
                }
            ]
        }
    ]
)
```

**User Responses**:
- **📋 Create Next SPEC**: Proceed to `/alfred:1-plan` for creating next SPEC
- **📤 Merge PR**: Manual PR review and merge on GitHub
- **🔄 Start New Session**: Execute `/clear` to start fresh session (recommended for performance)

---

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

> For more information: Skill("moai-alfred-dev-guide") - see section "Context Engineering"

### Core strategy of this command

**Load first**: `.moai/reports/sync-report-latest.md` (old sync state)

**Recommendation**: Document synchronization is complete. Now that the entire MoAI-ADK cycle (1-spec → 2-build → 3-sync) has been completed, start a new conversation session with the `/clear` or `/new` command before developing the next feature.

---

**Aims to improve code-document consistency and ensure @TAG traceability by linking with the doc-syncer subagent.**
