# MoAI Skill Reference

Common patterns, flag reference, legacy command mapping, and configuration files used across all MoAI workflows.

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

- DDD Workflow: Phase 1 (planning) feeds Phase 2 (implementation) feeds Phase 2.5 (quality validation)
- SPEC Creation: Explore agent results feed into manager-spec agent for document generation
- Release Pipeline: Quality gates must pass before version selection, which must complete before tagging

Implementation:

- Wait for each Task() to return before invoking the next
- Include previous phase outputs in the next Task() prompt as context
- Ensure semantic continuity: each agent receives sufficient context to operate independently

### Hybrid Execution Pattern

Combine parallel and sequential patterns within a single workflow.

Use Cases:

- Fix Workflow: Parallel diagnostic scan (LSP + linters + AST-grep), then sequential fix application based on combined results
- Alfred Workflow: Parallel exploration phase, then sequential SPEC generation and DDD implementation
- Run Workflow: Parallel quality checks, then sequential implementation tasks

Implementation:

- Identify which operations are independent (parallelize these)
- Identify which operations depend on prior results (sequence these)
- Group parallel operations at the beginning of each phase, followed by sequential dependent operations

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
- run --resume SPEC-XXX: Resume DDD implementation from last completed task
- alfred --resume SPEC-XXX: Resume full autonomous workflow from last phase
- fix --resume: Resume fix cycle from last diagnostic state
- release --resume: Resume release from last completed phase (uses snapshot in .moai/cache/release-snapshots/)

---

## Context Propagation Between Phases

Each phase must pass results forward to the next phase to avoid redundant analysis.

Required Context Elements:

- Exploration Results: File paths, architecture patterns, technology stack, dependency map
- SPEC Data: Requirements list, acceptance criteria, technical approach, scope boundaries
- Implementation Results: Files modified, tests created, coverage metrics, remaining tasks
- Quality Results: Test pass/fail counts, lint errors, type check results, security findings
- Git State: Current branch, commit count since last tag, tag history

Propagation Method:

- Include a structured summary of previous phase outputs in the Task() prompt
- Reference specific file paths rather than inline large content blocks
- Use SPEC document as the canonical source of truth across phases

---

## Flag Reference

### Global Flags (Available Across All Workflows)

- --resume SPEC-XXX: Resume workflow from last checkpoint for the specified SPEC
- --sequential (--seq): Force sequential execution instead of parallel where applicable

### Plan Flags

- --worktree: Create an isolated git worktree for the SPEC implementation
- --branch: Create a feature branch for the SPEC (default branch naming: spec/SPEC-XXX)
- --resume SPEC-XXX: Resume an interrupted plan session

### Run Flags

- --resume SPEC-XXX: Resume DDD implementation from last completed task

### Sync Flags

- auto: Default mode, generate documentation and PR automatically
- force: Regenerate all documentation even if unchanged
- status: Display sync status without performing operations
- project: Sync project documentation only (product.md, structure.md, tech.md)

### Fix Flags

- --dry: Preview detected issues without applying fixes
- --sequential: Process fixes one at a time instead of in parallel
- --level N: Control fix depth (Level 1: auto-fixable, Level 2: simple logic, Level 3: complex, Level 4: architectural)
- --resume: Resume from last diagnostic state

### Loop Flags

- --max N: Maximum iteration count (default: 100)
- --auto: Skip user confirmation between iterations
- --seq: Force sequential processing

### Alfred Flags

- --loop: Enable iterative fixing after implementation
- --max N: Maximum fix iterations when --loop is active
- --sequential: Force sequential agent execution
- --branch: Create feature branch before implementation
- --pr: Create pull request after completion
- --resume SPEC-XXX: Resume from last completed phase

### Release Flags

- VERSION argument: Specify target version directly (e.g., /moai release 0.35.0)
- --resume: Resume release from last saved snapshot
- --status: Check current release state

---

## Legacy Command Mapping

Previous /moai:X-Y command format mapped to new /moai subcommand format:

- /moai:0-project maps to /moai project
- /moai:1-plan maps to /moai plan
- /moai:2-run maps to /moai run
- /moai:3-sync maps to /moai sync
- /moai:9-feedback maps to /moai feedback (aliases: fb, bug, issue)
- /moai:99-release maps to /moai release (aliases: rel)
- /moai:fix maps to /moai fix
- /moai:loop maps to /moai loop
- /moai:alfred maps to /moai alfred (aliases: auto)

All legacy commands remain functional through alias resolution in the Intent Router.

---

## Configuration Files Reference

### Core Configuration

- .moai/config/config.yaml: Main configuration file (merged from section files)
- .moai/config/sections/language.yaml: Language settings (conversation_language, agent_prompt_language, code_comments)
- .moai/config/sections/user.yaml: User identification (name)
- .moai/config/sections/quality.yaml: TRUST 5 framework settings, LSP quality gates, test coverage targets
- .moai/config/sections/system.yaml: System metadata (moai.version)

### Project Documentation

- .moai/project/product.md: Product overview, features, user value
- .moai/project/structure.md: Project architecture and directory organization
- .moai/project/tech.md: Technology stack, dependencies, technical decisions

### SPEC Documents

- .moai/specs/SPEC-XXX/spec.md: Specification document with EARS format requirements
- .moai/specs/SPEC-XXX/plan.md: Execution plan with task breakdown
- .moai/specs/SPEC-XXX/acceptance.md: Acceptance criteria and test plan

### Release Artifacts

- CHANGELOG.md: Bilingual changelog (English + Korean per version)
- .moai/cache/release-snapshots/latest.json: Release state snapshot for recovery

### Version Files (5 files synchronized during release)

- pyproject.toml: Authoritative version source
- src/moai_adk/version.py: Runtime fallback version
- .moai/config/config.yaml: Config display version
- .moai/config/sections/system.yaml: System metadata version
- src/moai_adk/templates/.moai/config/sections/system.yaml: Template version for distribution

---

## Completion Markers

AI adds markers to signal workflow state:

- `<moai>DONE</moai>`: Single task or phase completed
- `<moai>COMPLETE</moai>`: Full workflow completed (all phases finished)

These markers enable automation detection and loop termination in the loop workflow.

---

## Error Handling Delegation

- Quality gate failures: Use expert-debug subagent for diagnosis and resolution
- Agent execution failures: Use expert-debug subagent for investigation
- Token limit errors: Execute /clear, then guide user to resume with --resume flag
- Permission errors: Review .claude/settings.json manually
- Integration errors: Use expert-devops subagent
- MoAI-ADK errors: Suggest /moai feedback to create a GitHub issue

---

Version: 1.0.0
Last Updated: 2026-01-28
