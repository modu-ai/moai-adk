---
name: moai
description: >
  MoAI super agent - unified orchestrator for autonomous development.
  Routes natural language or explicit subcommands (plan, run, sync, fix,
  loop, mx, project, feedback, review, challenge, clean, codemaps, coverage, e2e)
  to specialized agents.
  Use for any development task from planning to deployment.
allowed-tools: Task, AskUserQuestion, TaskCreate, TaskUpdate, TaskList, TaskGet, Bash, Read, Write, Edit, Glob, Grep
argument-hint: "[subcommand] [args] | \"natural language task\""
---

## Pre-execution Context

!`git status --porcelain 2>/dev/null || true`
!`git branch --show-current 2>/dev/null || true`

## Essential Files

.moai/config/config.yaml

---

## Authority References

Rules and constraints governing all workflows are always loaded from these sources. Do NOT duplicate their content here:

- Core identity, orchestration principles, agent catalog: CLAUDE.md
- Quality gates, security boundaries: .claude/rules/moai/core/moai-constitution.md
- SPEC workflow phases, token budgets: .claude/rules/moai/workflow/spec-workflow.md
- Development methodologies (DDD/TDD): .claude/rules/moai/workflow/workflow-modes.md
- Agent definitions: See CLAUDE.md Section 4. For agent creation, use builder-agent subagent.
- @MX tag rules and protocol: .claude/rules/moai/workflow/mx-tag-protocol.md

---

## Intent Router

### Raw User Input

$ARGUMENTS

### Routing Instructions

[HARD] Route the Raw User Input above using the strict priority order below. Extract the FIRST WORD of the input for subcommand matching. All text after the subcommand keyword is CONTEXT to be passed to the matched workflow — it is NOT a routing signal and MUST NOT influence which workflow is selected.

## Execution Mode Flags (mutually exclusive)

- `--team`: Force Agent Teams mode for parallel execution
- `--solo`: Force sub-agent mode (single agent per phase)
- No flag: System auto-selects based on complexity thresholds (domains >= 3, files >= 10, or complexity score >= 7)

When no flag is provided, the system evaluates task complexity and automatically selects between team mode (for complex, multi-domain tasks) and sub-agent mode (for focused, single-domain tasks).

## Effort Routing (Experimental)

- `--effort auto|low|medium|high`: Controls reasoning depth hints injected into agent prompts
- Default: `auto` (system determines based on task complexity)
- When `auto`: single-file fixes → low, moderate scope → medium, multi-domain/security → high
- When effort != medium: Prepend effort hint to all agent prompts in the workflow
- See `.claude/rules/moai/development/model-policy.md` for effort level definitions and auto-detection logic

### Priority 1: Explicit Subcommand Matching

[HARD] Extract the FIRST WORD from the Raw User Input section above. If it matches any subcommand below (or its alias), route to that workflow IMMEDIATELY. Do NOT analyze the remaining text for routing — it is context for the matched workflow:

- **plan** (aliases: spec): SPEC document creation workflow
- **run** (aliases: impl): DDD/TDD implementation workflow (per quality.yaml development_mode)
- **sync** (aliases: docs, pr): Documentation synchronization and PR creation
- **project** (aliases: init): Project documentation generation
- **feedback** (aliases: fb, bug, issue): GitHub issue creation
- **fix**: Auto-fix errors in a single pass
- **loop**: Iterative auto-fix until completion marker detected
- **mx**: MX tag scan and annotation for codebase
- **review** (aliases: code-review): Code review with security and MX tag compliance
- **challenge** (aliases: critique, devil's-advocate): Multi-perspective SPEC critique with 4 viewpoints
- **resume** (aliases: recover, 이어서): Resume interrupted SPEC work from checkpoint
- **clean** (aliases: dead-code): Identify and safely remove dead code
- **codemaps**: Generate architecture documentation in `.moai/project/codemaps/`
- **coverage** (aliases: cov): Analyze test coverage and generate missing tests
- **e2e** (aliases: e2e-test): Create and run E2E tests
- **context** (aliases: ctx, memory): Extract and display git-based context memory


### Priority 2: SPEC-ID Detection

Only if Priority 1 did not match: Check if the Raw User Input contains a pattern matching SPEC-XXX (such as SPEC-AUTH-001). If found, route to the **run** workflow automatically. The SPEC-ID becomes the target for DDD/TDD implementation.

### Priority 3: Natural Language Classification

Only if BOTH Priority 1 AND Priority 2 did not match: Classify the intent of the ENTIRE Raw User Input as natural language. This priority is NEVER reached when the first word matches a known subcommand.

- Planning and design language (design, architect, plan, spec, requirements, feature request) routes to **plan**
- Error and fix language (fix, error, bug, broken, failing, lint) routes to **fix**
- Iterative and repeat language (keep fixing, until done, repeat, iterate, all errors) routes to **loop**
- Documentation language (document, sync, docs, readme, changelog, PR) routes to **sync** or **project**
- Feedback and bug report language (report, feedback, suggestion, issue) routes to **feedback**
- MX tag language (mx tag, annotation, code context, legacy annotate) routes to **mx**
- Implementation language (implement, build, create, add, develop) with clear scope routes to **moai** (default autonomous)

### Priority 4: Default Behavior

If the intent remains ambiguous after all priority checks, use AskUserQuestion to present the top 2-3 matching workflows and let the user choose.

If the intent is clearly a development task with no specific routing signal, default to the **moai** workflow (plan -> run -> sync pipeline) for full autonomous execution.

---

## Workflow Quick Reference

### plan - SPEC Document Creation

Purpose: Create comprehensive specification documents using EARS format with Research-Plan-Annotate cycle.
Phases: Deep Research (research.md) -> SPEC Planning (with pre-mortem) -> Annotation Cycle (1-6 iterations) -> SPEC Creation -> Handoff Documentation (decisions.md)
Agents: manager-spec (primary), Explore (research), manager-git (conditional)
Flags: --worktree, --branch, --resume SPEC-XXX, --team, --no-issue, --consensus
For detailed orchestration: Read ${CLAUDE_SKILL_DIR}/workflows/plan.md

### run - DDD/TDD Implementation

Purpose: Implement SPEC requirements through configured development methodology.
Agents: manager-strategy, manager-ddd or manager-tdd (per quality.yaml), manager-quality, manager-git
Flags: --resume SPEC-XXX, --team
For detailed orchestration: Read ${CLAUDE_SKILL_DIR}/workflows/run.md

### sync - Documentation Sync and PR

Purpose: Synchronize documentation with code changes and prepare pull requests.
Agents: manager-docs (primary), manager-quality, manager-git
Modes: auto, force, status, project. Flags: --merge, --skip-mx
For detailed orchestration: Read ${CLAUDE_SKILL_DIR}/workflows/sync.md

### fix - Auto-Fix Errors

Purpose: Autonomously detect and fix LSP errors, linting issues, and type errors.
Agents: expert-debug (diagnosis), expert-backend/expert-frontend (fixes)
Flags: --dry, --sequential, --level N, --resume, --team
For detailed orchestration: Read ${CLAUDE_SKILL_DIR}/workflows/fix.md

### loop - Iterative Auto-Fix

Purpose: Repeatedly fix issues until completion marker detected or max iterations reached.
Agents: expert-debug, expert-backend, expert-frontend, expert-testing
Flags: --max N, --auto-fix, --seq
For detailed orchestration: Read ${CLAUDE_SKILL_DIR}/workflows/loop.md

### mx - MX Tag Scan and Annotation

Purpose: Scan codebase and add @MX code-level annotations for AI agent context.
Agents: Explore (scan), expert-backend (annotation)
Flags: --all, --dry, --priority P1-P4, --force, --team
For detailed orchestration: Read ${CLAUDE_SKILL_DIR}/workflows/mx.md

### review - Code Review

Purpose: Multi-perspective code review with security, performance, quality, and UX analysis.
Agents: manager-quality (primary), expert-security
Flags: --staged, --branch, --security, --team
For detailed orchestration: Read ${CLAUDE_SKILL_DIR}/workflows/review.md (team mode: ${CLAUDE_SKILL_DIR}/team/review.md)

### challenge - Multi-Perspective SPEC Critique

Purpose: Generate critical questions from tech, business, user, and ops perspectives before implementation.
Phases: SPEC Validation → Parallel Critic Invocation (4 agents) → Question Consolidation → User Q&A → Report
Agents: manager-challenge (primary), critic-tech, critic-business, critic-user, critic-ops
Flags: --skip-persona tech|business|user|ops, --auto
For detailed orchestration: Read ${CLAUDE_SKILL_DIR}/workflows/challenge.md

### resume - Resume Interrupted Work

Purpose: Detect and resume interrupted SPEC work from journal checkpoints.
Phases: Detect Resumable Work → Build Resume Context → Confirm → Execute
Agents: None (orchestrator-driven)
Flags: [SPEC-XXX] (optional, auto-detect if omitted)
For detailed orchestration: Read ${CLAUDE_SKILL_DIR}/workflows/resume.md

### clean - Dead Code Removal

Purpose: Identify and safely remove unused code with test verification.
Agents: expert-refactoring, expert-testing
Flags: --dry, --safe-only, --file PATH
For detailed orchestration: Read ${CLAUDE_SKILL_DIR}/workflows/clean.md

### codemaps - Architecture Documentation

Purpose: Scan codebase and generate architecture documentation.
Agents: Explore, manager-docs
Flags: --force, --area AREA
For detailed orchestration: Read ${CLAUDE_SKILL_DIR}/workflows/codemaps.md

### coverage - Test Coverage Analysis

Purpose: Analyze test coverage gaps and generate missing tests.
Agents: expert-testing
Flags: --target N, --file PATH, --report
For detailed orchestration: Read ${CLAUDE_SKILL_DIR}/workflows/coverage.md

### e2e - End-to-End Testing

Purpose: Create and run E2E tests using Chrome, Playwright, or Agent Browser.
Agents: expert-testing, expert-frontend
Flags: --record, --url URL, --journey NAME
For detailed orchestration: Read ${CLAUDE_SKILL_DIR}/workflows/e2e.md

### (default) - MoAI Autonomous Workflow

Purpose: Full autonomous research -> plan -> annotate -> run -> sync pipeline.
Phases: Parallel Exploration (research.md) -> SPEC Generation -> Annotation Cycle -> Implementation -> Sync
Agents: Explore, manager-spec, manager-ddd/tdd, manager-quality, manager-docs, manager-git
Flags: --loop, --max N, --branch, --pr, --resume SPEC-XXX, --team, --solo, --no-issue
For detailed orchestration: Read ${CLAUDE_SKILL_DIR}/workflows/moai.md

### project - Project Documentation

Purpose: Generate project documentation by analyzing the existing codebase.
Agents: Explore, manager-docs, expert-devops (optional)
Output: product.md, structure.md, tech.md in .moai/project/
For detailed orchestration: Read ${CLAUDE_SKILL_DIR}/workflows/project.md

### context - Git-Based Context Memory

Purpose: Extract AI-developer interaction context from git commit history for session continuity.
Agents: manager-git (primary)
Flags: --spec SPEC-XXX, --days N, --category CAT, --summary, --inject
For detailed orchestration: Read ${CLAUDE_SKILL_DIR}/workflows/context.md

### feedback - GitHub Issue Creation

Purpose: Collect user feedback and create GitHub issues.
Agents: manager-quality
For detailed orchestration: Read ${CLAUDE_SKILL_DIR}/workflows/feedback.md

---

## Execution Directive

When this skill is activated, execute the following steps in order:

Step 1 - Parse Arguments:
Extract subcommand keywords and flags from the Raw User Input. Recognized global flags: --resume [ID], --seq, --deepthink, --team, --solo, --consensus, --effort [auto|low|medium|high]. When --deepthink is detected, activate Sequential Thinking MCP for deep analysis before execution. When --consensus is detected, pass it to the plan workflow for consensus planning loop.

Step 2 - Route to Workflow:
Apply the Intent Router (Priority 1 through Priority 4) to determine the target workflow. If ambiguous, use AskUserQuestion to clarify with the user.

Step 2.5 - Project Documentation Check:
Before executing plan, run, sync, fix, loop, or default workflows, verify project documentation exists by checking for `.moai/project/product.md`. If product.md does NOT exist, use AskUserQuestion to ask the user (in their conversation_language):

Question: Project documentation not found. Would you like to create it first?
Options:
- Create project documentation (Recommended): Generates product.md, structure.md, tech.md through a guided interview. This helps MoAI understand your project context for better results in all subsequent workflows.
- Skip and continue: Proceed without project documentation. MoAI will have less context about your project.

This check does NOT apply to: project, feedback subcommands.

[HARD] Beginner-Friendly Option Design:
All AskUserQuestion calls throughout MoAI workflows MUST follow these rules:
- The first option MUST always be the recommended choice, clearly marked with "(Recommended)" suffix
- Every option MUST include a detailed description explaining what it does and its implications

Step 3 - Load Workflow Details:
Read the corresponding workflows/<name>.md file for detailed orchestration instructions.

Step 4 - Read Configuration:
Load relevant configuration from .moai/config/config.yaml and section files as needed.

Step 5 - Initialize Task Tracking:
Use TaskCreate to register discovered work items with pending status.

Step 6 - Execute Workflow Phases:
Follow the workflow-specific phase instructions. Delegate all implementation to appropriate agents via Agent(). Collect user approvals at designated checkpoints via AskUserQuestion.

Step 7 - Track Progress:
Update task status using TaskUpdate as work progresses (pending to in_progress to completed).

Step 8 - Present Results:
Display results to the user in their conversation_language using Markdown format.

Step 9 - Add Completion Marker:
When all workflow phases complete successfully, add the appropriate completion marker (`<moai>DONE</moai>` or `<moai>COMPLETE</moai>`).

Step 10 - Guide Next Steps:
Use AskUserQuestion to present the user with logical next actions based on the completed workflow.

---

Version: 2.7.0
Last Updated: 2026-03-19
