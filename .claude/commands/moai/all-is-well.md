---
name: moai:all-is-well
description: "One-click automation - From SPEC generation to documentation sync"
argument-hint: '"feature description" [--branch] [--pr] | resume SPEC-XXX'
type: workflow
allowed-tools: Task, AskUserQuestion, TodoWrite, Skill, Glob
model: inherit
---

## Pre-execution Context

!git status --porcelain
!git branch --show-current

## Essential Files

@.moai/config/sections/ralph.yaml
@.moai/config/sections/git-strategy.yaml
@.moai/config/sections/quality.yaml

---

# /moai:all-is-well - One-Click Development Automation

User Interaction Architecture: AskUserQuestion must be used at COMMAND level only. Subagents via Task() are stateless and cannot interact with users. Collect all user input BEFORE delegating to agents.

Execute the complete MoAI development workflow with a single command.

## Command Purpose

Automates the full "Plan -> Run -> Sync" workflow:

1. Creates SPEC from description (`/moai:1-plan`)
2. Implements with TDD (`/moai:2-run`)
3. Synchronizes documentation (`/moai:3-sync`)

Feature Description: $ARGUMENTS

## Intelligent Routing Analysis

Before starting the workflow, analyze the user's requirements to determine the optimal execution path.

### Analysis Criteria

Step 1 - Requirement Complexity Assessment:

- Single-domain task: Feature touches only one area (e.g., only backend, only frontend)
- Multi-domain task: Feature spans multiple areas (e.g., API + UI + database)
- Full-stack feature: Requires end-to-end implementation with testing and documentation

Step 2 - Workflow Path Decision:

IF requirement is a full-stack feature requiring:

- SPEC documentation
- TDD implementation with coverage requirements
- Documentation synchronization
- Git workflow integration

THEN: Use Sub-Command Workflow (Plan -> Run -> Sync)

- Phase 1: /moai:1-plan for SPEC generation
- Phase 2: /moai:2-run for TDD implementation
- Phase 3: /moai:3-sync for documentation sync

IF requirement is a targeted task in single domain:

THEN: Delegate to Expert Agent directly

- Backend-only: Use expert-backend subagent
- Frontend-only: Use expert-frontend subagent
- Database-only: Use expert-database subagent
- Security audit: Use expert-security subagent
- Performance optimization: Use expert-performance subagent
- Bug fix: Use expert-debug subagent

### Domain Detection Keywords

Backend domain indicators:

- API, REST, GraphQL, endpoint, server, authentication, JWT, OAuth
- Korean: API, 서버, 인증, 백엔드
- Japanese: API, サーバー, 認証, バックエンド

Frontend domain indicators:

- UI, component, React, Vue, Next.js, CSS, responsive, state management
- Korean: UI, 컴포넌트, 프론트엔드, 상태관리
- Japanese: UI, コンポーネント, フロントエンド

Database domain indicators:

- Schema, migration, query, index, PostgreSQL, MongoDB, Redis
- Korean: 스키마, 쿼리, 데이터베이스, 인덱스
- Japanese: スキーマ, クエリ, データベース

Security domain indicators:

- Security, vulnerability, OWASP, injection, XSS, CSRF, audit
- Korean: 보안, 취약점, 감사
- Japanese: セキュリティ, 脆弱性, 監査

### Routing Decision Flow

Step 1 - Parse user's feature description
Step 2 - Detect domain keywords from description
Step 3 - Count domains involved

IF domain_count == 1 AND no_spec_required:

- Confirm with user: "This appears to be a single-domain task. Delegate to [expert-X] agent directly?"
- IF user confirms: Delegate to expert agent
- IF user declines: Proceed with full workflow

IF domain_count >= 2 OR explicit_spec_requested:

- Proceed with Sub-Command Workflow (Plan -> Run -> Sync)

### User Confirmation for Routing

Use AskUserQuestion to confirm routing decision:

Question: "Based on your requirement, I recommend [Routing Option]. How would you like to proceed?"

Options:

- Full Workflow: Execute complete Plan -> Run -> Sync cycle with SPEC
- Expert Delegation: Delegate directly to [detected expert] agent
- Hybrid: Start with expert implementation, then sync documentation

WHY: User may have context not detectable from description alone
IMPACT: Wrong routing wastes time or produces incomplete deliverables

## Argument Parsing

Step 1 - Parse $ARGUMENTS:

- Detect resume mode: If $ARGUMENTS starts with "resume", extract SPEC-ID
- Extract feature description: Text before any flags (quoted or unquoted)
- Detect --branch flag: Boolean, overrides git-strategy for branch creation
- Detect --pr flag: Boolean, overrides git-strategy for PR creation

Step 2 - Validate Arguments:

- [HARD] Feature description must not be empty (unless resume mode)
  WHY: Empty description produces vague SPEC
  IMPACT: Workflow fails at Phase 1 with unclear requirements

- [HARD] Resume SPEC-ID must exist when resume mode is active
  WHY: Cannot resume non-existent workflow
  IMPACT: Immediate failure with clear error message

Step 3 - Store Parsed Values:

- $FEATURE_DESCRIPTION: User's feature description
- $BRANCH_FLAG: true if --branch present
- $PR_FLAG: true if --pr present
- $RESUME_MODE: true if resume command detected
- $RESUME_SPEC_ID: SPEC-XXX if resume mode active

## Usage Examples

Basic usage (uses git-strategy settings):

```
/moai:all-is-well "User authentication with JWT tokens"
```

With branch creation override:

```
/moai:all-is-well "Shopping cart feature" --branch
```

With PR creation override:

```
/moai:all-is-well "Payment integration" --pr
```

Combined flags:

```
/moai:all-is-well "OAuth2 authentication with Google and GitHub providers" --branch --pr
```

Resume interrupted workflow:

```
/moai:all-is-well resume SPEC-AUTH-001
```

Multi-word description (proper quoting):

```
/moai:all-is-well "Shopping cart with quantity limits and discount codes"
```

## Command Options

- `--branch`: Override git-strategy to create a feature branch
- `--pr`: Override git-strategy to create a pull request after sync
- `resume SPEC-XXX`: Resume workflow from last checkpoint
- Default behavior follows `.moai/config/sections/git-strategy.yaml` settings

## Associated Agents

This command orchestrates three specialized agent workflows:

manager-spec (via /moai:1-plan):

- Responsibility: SPEC generation in EARS format
- Input: Feature description from $ARGUMENTS
- Output: Approved SPEC document with SPEC-ID
- Checkpoint: User must approve SPEC before proceeding

manager-tdd (via /moai:2-run):

- Responsibility: TDD implementation with RED-GREEN-REFACTOR cycle
- Input: Approved SPEC-ID from Phase 1
- Output: Implementation with 85%+ test coverage
- Checkpoint: Quality gate validation before proceeding

manager-docs (via /moai:3-sync):

- Responsibility: Documentation synchronization
- Input: SPEC-ID and implementation context from Phase 2
- Output: Synchronized documentation, commits, optional PR
- Checkpoint: Final review and completion summary

## Agent Invocation Patterns (CLAUDE.md Compliance)

This command uses agent execution patterns defined in CLAUDE.md.

### Sequential Phase-Based Chaining

Command implements sequential chaining through 3 orchestrated phases:

Phase Flow:

- Phase 1: SPEC Generation (delegates to /moai:1-plan)
- Phase 2: TDD Implementation (delegates to /moai:2-run)
- Phase 3: Documentation Sync (delegates to /moai:3-sync)

Each phase receives outputs from previous phases as context.

WHY: Sequential execution ensures SPEC approval before implementation, and implementation before documentation
IMPACT: Parallel execution would create documentation without approved SPEC or implemented code

### Resumable Agent Support

Command supports resume pattern using SPEC ID:

- Invoke: /moai:all-is-well resume SPEC-XXX
- Behavior: Resumes from last successful phase checkpoint
- Storage: Workflow state preserved in `.moai/cache/all-is-well-{spec-id}.json`

WHY: Complex workflows may encounter interruptions or token limits
IMPACT: Resume capability prevents loss of workflow progress

### User Interaction Consolidation

All user decisions are collected at command level before delegation:

- Phase 1: SPEC approval (proceed/modify/draft/cancel)
- Phase 2: Quality gate override (if needed)
- Phase 3: PR creation confirmation (if configured)

WHY: Subagents via Task() are stateless and cannot interact with users
IMPACT: Ensures smooth workflow without blocking on user input

## Workflow Execution

### Phase 1: SPEC Generation

Agent: manager-spec (via /moai:1-plan)

Actions:

- Analyze feature description
- Generate SPEC document in EARS format
- Create acceptance criteria
- Present SPEC for user approval

Checkpoint: User must approve SPEC before proceeding

### Phase 2: TDD Implementation

Agent: manager-tdd (via /moai:2-run)

Actions:

- Create execution plan from SPEC
- Execute RED-GREEN-REFACTOR cycle
- Achieve minimum 85% test coverage
- Validate with TRUST 5 framework

Checkpoint: Quality gate validation before proceeding

### Phase 3: Documentation Sync

Agent: manager-docs (via /moai:3-sync)

Actions:

- Update documentation to match implementation
- Create or update README sections
- Generate API documentation if applicable
- Create PR if configured

Checkpoint: Final review and completion summary

## Quality Gate Decision Points

### Phase 1 to Phase 2 Gate

Condition: User must explicitly approve SPEC

Options presented via AskUserQuestion:

- Proceed: Continue to Phase 2 with approved SPEC
- Modify: Return to SPEC editing
- Save as Draft: Save SPEC for later without implementing
- Cancel: Abort workflow entirely

IF not approved: Workflow halts at Phase 1

### Phase 2 to Phase 3 Gate

Conditions:

- TRUST 5 validation must pass or warn (not CRITICAL)
- Test coverage must be >= 85% (configurable in quality.yaml)
- All tests must pass

IF CRITICAL issues found: Present issues to user with options:

- Retry: Attempt to fix issues automatically
- Override: Proceed despite issues (with warning)
- Abort: Stop workflow and preserve work

### Phase 3 Completion Gate

Conditions:

- Documentation sync successful
- Git commit created (per git-strategy)
- PR created if requested and configured

IF PR creation fails: Offer manual PR creation instructions

## Context Propagation

### Phase 1 to Phase 2

Phase 1 Output:

- Approved SPEC-ID (e.g., SPEC-AUTH-001)
- SPEC file path
- Acceptance criteria summary

Phase 2 Receives:

- SPEC-ID for implementation reference
- SPEC content for TDD planning

WHY: Ensures correct SPEC is implemented
IMPACT: Missing SPEC-ID causes Phase 2 to fail with clear error

### Phase 2 to Phase 3

Phase 2 Output:

- Implementation summary (files changed, lines added)
- Test results (passed/failed count, coverage percentage)
- Branch name and commit hashes

Phase 3 Receives:

- SPEC-ID for documentation reference
- Git status and changed files list
- Test coverage report

WHY: Documentation must accurately reflect implementation
IMPACT: Missing context produces inconsistent or incomplete documentation

### Checkpoint Data Persistence

Storage Location: `.moai/cache/all-is-well-{spec-id}.json`

Contents:

- phase_completed: Last successfully completed phase (1, 2, or 3)
- spec_id: SPEC identifier
- branch_name: Git branch if created
- git_commits: List of commit hashes created
- timestamp: Last update time

WHY: Enables resume after interruption or session timeout
IMPACT: Without persistence, workflow must restart from Phase 1

## Git Strategy Integration

This command respects your git-strategy.yaml configuration:

Manual Mode (default):

- No automatic branch creation
- No automatic PR creation
- All changes on current branch

Personal Mode:

- Auto-creates feature branch
- Commits automatically
- PR creation optional

Team Mode:

- Auto-creates feature branch
- Commits automatically
- Auto-creates draft PR

Override with `--branch` or `--pr` flags when needed.

## Ralph Engine Integration

### Conditional Activation

Check: @.moai/config/sections/ralph.yaml for `ralph.enabled: true`
IF disabled: Skip Ralph integration entirely, proceed with standard workflow

### Phase 2 Integration (TDD Implementation)

When Ralph Engine is enabled:

- After each file change: Trigger LSP diagnostic check
- Before test run: Execute AST-grep security scan
- Issue detection: Categorize as ERROR, WARNING, or INFO

### Feedback Loop Pattern

Execution sequence: Tests -> LSP -> AST-grep

IF any issues found:

- Attempt auto-fix (max 3 iterations per issue type)
- Track fix attempts in `.moai/cache/ralph-fixes.json`
- IF still failing after 3 attempts: Escalate to user

### LSP Diagnostics

- Run on: All modified files in current phase
- Check for: Type errors, undefined references, import issues
- Integration: Works with project's configured language server

### AST-grep Security Scanning

- Run on: All modified source files
- Patterns: OWASP Top 10, injection vulnerabilities, hardcoded secrets
- Output: Security findings with severity levels

## Progress Tracking

Initialize TodoWrite at command start with workflow phases:

1. Parse arguments and validate input
2. Execute Phase 1: SPEC Generation
3. Await user SPEC approval
4. Execute Phase 2: TDD Implementation
5. Validate quality gates
6. Execute Phase 3: Documentation Sync
7. Generate completion summary

Update status as each step completes:

- Mark current step as in_progress before starting
- Mark as completed immediately upon success
- Add sub-tasks if phase reveals additional work

WHY: Visible progress maintains user confidence during long workflows
IMPACT: Enables debugging and recovery if specific steps fail

## Error Recovery

If any phase fails:

- Current progress is preserved in `.moai/cache/`
- User is notified of the failure point
- Recovery options are presented

### Phase 1 Failure (SPEC Generation)

Message: "SPEC generation failed: [specific reason]"

Options:

- Retry: Re-attempt SPEC generation with same description
- Edit description: Modify feature description and retry
- Abort: Exit workflow, preserve nothing

### Phase 2 Failure (TDD Implementation)

Message: "Implementation failed: [specific reason]"

Details provided:

- Test results: X tests passing, Y tests failing
- Coverage: Current percentage vs 85% threshold
- TRUST 5 status: Validation results

Options:

- Retry: Attempt to fix failing tests automatically
- Skip quality gate: Proceed despite issues (creates warning annotation)
- Manual fix: Pause workflow for manual intervention, then resume
- Abort: Exit workflow, preserve implemented code

### Phase 3 Failure (Documentation Sync)

Message: "Documentation sync failed: [specific reason]"

Common failure reasons:

- Merge conflict detected
- PR creation failed (permissions, network)
- Documentation generation error

Options:

- Resolve conflicts: Open conflict resolution flow
- Force sync: Override conflicts (with backup)
- Manual PR: Skip auto-PR, provide manual instructions
- Abort: Exit workflow, preserve all work

## Success Criteria

Workflow is considered successful when:

- SPEC document created and approved by user
- Implementation complete with 85%+ test coverage
- All tests passing (or user override with warning)
- Documentation synchronized with implementation
- Git operations complete per configuration

## Output Format

Phase completion reports use Markdown formatting:

```markdown
## All-Is-Well Workflow Complete

### Summary

- SPEC: SPEC-XXX created and approved
- Implementation: 12 files, 88% coverage
- Tests: 24/24 passing
- Documentation: Updated

### Git Status

- Branch: feature/SPEC-XXX (or current branch)
- Commits: 3 commits created
- PR: #123 created (if applicable)

### Next Steps

1. Review the implementation
2. Run manual testing if needed
3. Merge when ready
```

## Quick Reference

Command Syntax:

- New workflow: `/moai:all-is-well "description" [--branch] [--pr]`
- Resume workflow: `/moai:all-is-well resume SPEC-XXX`

Phase Summary:

- Phase 1: SPEC -> User Approval -> Phase 2
- Phase 2: TDD -> Quality Gate -> Phase 3
- Phase 3: Docs -> PR (optional) -> Complete

Key Files:

- SPEC storage: `.moai/specs/SPEC-XXX.md`
- Cache storage: `.moai/cache/all-is-well-{spec-id}.json`
- Config files: `.moai/config/sections/*.yaml`

Override Flags:

- `--branch`: Force feature branch creation
- `--pr`: Force PR creation after sync

---

## Implementation Notes

Tool Usage: This command orchestrates through Task() delegation only. Direct file operations are prohibited except for Glob (SPEC validation).

User Interaction: All AskUserQuestion calls happen at command level before delegation. Subagents receive decisions as parameters.

Context Propagation: Each phase receives context from previous phases via structured handoff data.

Interruption Recovery: Loop state preserved in `.moai/cache/` for resume capability. Use `resume SPEC-XXX` to continue.

Skill Loading: Load `moai-foundation-core` and `moai-workflow-project` skills at command initialization for workflow context.

---

## EXECUTION DIRECTIVE

You must NOW execute the command following the workflow described above.

1. Start by parsing $ARGUMENTS to extract feature description and detect flags.
2. IF resume mode: Load cached state and skip to appropriate phase.
3. Perform Intelligent Routing Analysis:
   - Detect domain keywords in feature description
   - Count domains involved
   - Determine optimal execution path
4. Use AskUserQuestion to confirm routing decision with user:
   - Full Workflow (Plan -> Run -> Sync)
   - Expert Delegation (single domain task)
   - Hybrid (expert implementation + doc sync)
5. IF Full Workflow selected:
   - Execute Phase 1 by invoking Skill("moai:1-plan") with feature description
   - Wait for user SPEC approval via AskUserQuestion
   - Execute Phase 2 by invoking Skill("moai:2-run") with approved SPEC-ID
   - Validate quality gates and present results
   - Execute Phase 3 by invoking Skill("moai:3-sync") with SPEC-ID
6. IF Expert Delegation selected:
   - Invoke appropriate expert subagent (expert-backend, expert-frontend, etc.)
   - Report completion directly without SPEC generation
7. IF Hybrid selected:
   - Invoke expert subagent for implementation
   - Execute Phase 3 by invoking Skill("moai:3-sync") for documentation
8. Generate completion summary and present next steps.
9. Do NOT just describe what you will do. DO IT.

---

Version: 2.1.0
Last Updated: 2026-01-10
Pattern: Intelligent Routing with Sequential Phase Orchestration
Integration: /moai:1-plan, /moai:2-run, /moai:3-sync, Expert Agents, Ralph Engine, Git Strategy, TRUST 5
