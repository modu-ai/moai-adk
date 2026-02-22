---
name: moai-reference
description: >
  Common execution patterns, flag reference, legacy command mapping,
  and resume patterns used across all MoAI workflows. Provides context
  propagation guidance and team execution patterns.
  Use when needing execution patterns, flag details, or configuration reference.
user-invocable: false
metadata:
  version: "2.5.0"
  category: "foundation"
  status: "active"
  updated: "2026-02-22"
  tags: "reference, patterns, flags, configuration, legacy, resume, context"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["reference", "pattern", "flag", "config", "resume", "legacy", "mapping"]
  agents: ["manager-spec", "manager-ddd", "manager-tdd", "manager-docs", "manager-quality", "manager-git"]
  phases: ["plan", "run", "sync"]
---

# MoAI Skill Reference

Common patterns, flag reference, and legacy command mapping used across all MoAI workflows.

For configuration file paths, see: @CLAUDE.md Section 9
For error handling delegation, see: @CLAUDE.md Section 11
For completion markers, see: @.claude/rules/moai/core/moai-constitution.md
For development mode details, see: @.claude/rules/moai/workflow/workflow-modes.md

---

## Execution Patterns

### Parallel Execution Pattern

When multiple operations are independent, invoke them in a single response. Claude Code automatically runs multiple Task() calls in parallel (up to 10 concurrent).

Use Cases:

- Exploration Phase: Launch codebase analysis, documentation research, and quality assessment simultaneously via separate Task() calls
- Diagnostic Scan: Run LSP diagnostics, AST-grep analysis, and linter checks in parallel
- Multi-file Generation: Generate product.md, structure.md, and tech.md simultaneously when analysis is complete

Implementation:

- Include multiple Task() calls in the same response message
- Each Task() targets a different subagent or a different scope within the same agent
- Results are collected when all parallel tasks complete
- Maximum 10 concurrent Task() calls for optimal throughput

### Sequential Execution Pattern

When operations have dependencies, chain them sequentially. Each Task() call receives context from the previous phase results.

Use Cases:

- DDD/TDD Workflow: Phase 1 (planning) feeds Phase 2 (implementation) feeds Phase 2.5 (quality validation)
- SPEC Creation: Explore agent results feed into manager-spec agent for document generation
- Release Pipeline: Quality gates must pass before version selection, which must complete before tagging

### Hybrid Execution Pattern

Combine parallel and sequential patterns within a single workflow.

Use Cases:

- Fix Workflow: Parallel diagnostic scan (LSP + linters + AST-grep), then sequential fix application based on combined results
- MoAI Workflow: Parallel exploration phase, then sequential SPEC generation and DDD/TDD implementation

### Team Execution Pattern

When team mode is enabled, use Agent Teams for persistent parallel coordination.

Use Cases:

- Plan Phase: Parallel research team (researcher + analyst + architect)
- Run Phase: Parallel implementation team (backend-dev + frontend-dev + tester) with file ownership boundaries
- Debug Phase: Competing hypothesis investigation team

Implementation:

- TeamCreate to initialize team structure with shared task list
- Task() with team_name and name parameters to spawn teammates
- SendMessage for inter-teammate coordination and idle handling
- TaskList for self-coordinated work distribution
- TeamDelete after all teammates shut down

---

## Resume Pattern

When a workflow is interrupted or needs to continue from a previous session, use the --resume flag.

Behavior:

- Read existing SPEC document from .moai/specs/SPEC-XXX/
- Determine last completed phase from SPEC status markers
- Skip completed phases and resume from the next pending phase
- Preserve all prior analysis, decisions, and generated artifacts

Applicable Workflows:

- plan --resume SPEC-XXX: Resume SPEC creation from last checkpoint
- run --resume SPEC-XXX: Resume DDD/TDD implementation from last completed task
- moai --resume SPEC-XXX: Resume full autonomous workflow from last phase
- fix --resume: Resume fix cycle from last diagnostic state

---

## Context Propagation Between Phases

Each phase must pass results forward to the next phase to avoid redundant analysis.

Required Context Elements:

- Exploration Results: File paths, architecture patterns, technology stack, dependency map
- SPEC Data: Requirements list, acceptance criteria, technical approach, scope boundaries
- Implementation Results: Files modified, tests created, coverage metrics, remaining tasks
- Quality Results: Test pass/fail counts, lint errors, type check results, security findings
- Implementation Divergence: Planned vs actual files, additional features, scope changes, new dependencies
- Git State: Current branch, commit count since last tag, tag history

Propagation Method:

- Include a structured summary of previous phase outputs in the Task() prompt
- Reference specific file paths rather than inline large content blocks
- Use SPEC document as the canonical source of truth across phases
- Pass implementation divergence report from run phase to sync phase for SPEC/project document updates

---

## Flag Reference

### Global Flags (Available Across All Workflows)

- --resume [ID]: Resume workflow from last checkpoint (SPEC-ID or snapshot ID)
- --seq: Force sequential execution instead of parallel where applicable
- --ultrathink: Activate Sequential Thinking MCP for deep analysis before execution
- --team: Force Agent Teams mode for parallel execution
- --solo: Force sub-agent mode (single agent per phase)

### Plan Flags

- --worktree: Create an isolated git worktree for the SPEC implementation
- --branch: Create a feature branch for the SPEC (default branch naming: feature/SPEC-XXX)
- --resume SPEC-XXX: Resume an interrupted plan session
- --team: Force team-based exploration (researcher + analyst + architect)

### Run Flags

- --resume SPEC-XXX: Resume DDD/TDD implementation from last completed task
- --team: Force team-based implementation (backend-dev + frontend-dev + tester)
- --review: Enable post-implementation review loop

### Sync Flags

- Modes (positional): auto (default), force, status, project
- --merge: Auto-merge PR and clean up branch after sync
- --skip-mx: Skip MX tag validation during sync

### Fix Flags

- --dry: Preview detected issues without applying fixes
- --level N: Control fix depth (Level 1: auto-fixable, Level 2: simple logic, Level 3: complex, Level 4: architectural)
- --security: Include security issues in scan
- --sequential: Run diagnostics sequentially instead of in parallel
- --resume: Resume fix cycle from last diagnostic state
- --team: Force competing hypothesis investigation team

### Loop Flags

- --max N: Maximum iteration count (default: 100)
- --auto-fix: Enable automatic fix application for Level 1-2 issues
- --seq: Force sequential diagnostics

### MX Flags

- --all: Scan entire codebase (not just modified files)
- --dry: Preview tag changes without applying
- --priority P1-P4: Filter by priority level
- --force: Overwrite existing tags
- --team: Force parallel scan by language

### Review Flags

- --staged: Review staged changes only
- --branch: Compare against specified branch
- --security: Focus on security review
- --team: Force parallel multi-perspective review team

### Coverage Flags

- --target N: Set coverage target percentage
- --file PATH: Analyze specific file
- --report: Generate report only (no test generation)

### E2E Flags

- --record: Record browser session as GIF
- --url URL: Target URL for testing
- --journey NAME: Specific user journey to test

### Clean Flags

- --dry: Preview dead code without removing
- --safe-only: Only remove confirmed dead code
- --file PATH: Target specific file for analysis

### MoAI (Default) Flags

- --loop: Enable iterative fixing during run phase
- --max N: Maximum fix iterations when --loop is active
- --branch: Create feature branch before implementation
- --pr: Create pull request after completion

---

## Legacy Command Mapping

Previous /moai:X-Y command format mapped to new /moai subcommand format:

- /moai:0-project maps to /moai project
- /moai:1-plan maps to /moai plan
- /moai:2-run maps to /moai run
- /moai:3-sync maps to /moai sync
- /moai:9-feedback maps to /moai feedback
- /moai:fix maps to /moai fix
- /moai:loop maps to /moai loop
- /moai:moai maps to /moai (default autonomous workflow)
- /moai:review maps to /moai review
- /moai:clean maps to /moai clean
- /moai:codemaps maps to /moai codemaps
- /moai:coverage maps to /moai coverage
- /moai:e2e maps to /moai e2e
- /moai:mx maps to /moai mx

Note: /moai:99-release is a separate local-only command, not part of the /moai skill.

---

Version: 2.5.0
Last Updated: 2026-02-22
