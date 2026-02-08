---
name: moai-workflow-plan
description: >
  Creates comprehensive SPEC documents using EARS format as the first step
  of the Plan-Run-Sync workflow. Handles project exploration, SPEC file
  generation, validation, and optional Git environment setup with worktree
  or branch creation. Use when planning features or creating specifications.
license: Apache-2.0
compatibility: Designed for Claude Code
user-invocable: false
metadata:
  version: "1.0.0"
  category: "workflow"
  status: "active"
  updated: "2026-02-03"
  tags: "plan, spec, ears, requirements, specification, design"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["plan", "spec", "design", "architect", "requirements", "feature request"]
  agents: ["manager-spec", "Explore", "manager-git"]
  phases: ["plan"]
---

# Plan Workflow Orchestration

## Purpose

Create comprehensive SPEC documents using EARS format as the first step of the Plan-Run-Sync workflow. This workflow handles everything from project exploration to SPEC file generation and optional Git environment setup.

## Scope

- Implements Steps 1-2 of MoAI's 4-step workflow (Intent Understanding, Plan Creation)
- Steps 3-4 are handled by /moai run and /moai sync respectively

## Input

- $ARGUMENTS: One of three patterns
  - Feature description: "User authentication system"
  - Resume command: resume SPEC-XXX
  - Feature description with flags: "User authentication" --worktree or --branch

## Supported Flags

- --worktree: Create isolated Git worktree environment (highest priority)
- --branch: Create traditional feature branch (second priority)
- No flag: SPEC only by default; user may be prompted based on config
- --team: Enable team-based exploration (see team-plan.md for parallel research team)
- resume SPEC-XXX: Continue from last saved draft state

Flag priority: --worktree takes precedence over --branch, which takes precedence over default.

## Context Loading

Before execution, load these essential files:

- .moai/config/config.yaml (git strategy, language settings)
- .moai/config/sections/git-strategy.yaml (auto_branch, branch creation policy)
- .moai/config/sections/language.yaml (git_commit_messages setting)
- .moai/project/product.md (product context)
- .moai/project/structure.md (architecture context)
- .moai/project/tech.md (technology context)
- .moai/specs/ directory listing (existing SPECs for deduplication)

Pre-execution commands: git status, git branch, git log, git diff, find .moai/specs.

---

## Phase Sequence

### Phase 1A: Project Exploration (Optional)

Agent: Explore subagent (read-only codebase analysis)

When to run:

- User provides vague or unstructured request
- Need to discover existing files and patterns
- Unclear about current project state

When to skip:

- User provides clear SPEC title (e.g., "Add authentication module")
- Resume scenario with existing SPEC context

Tasks for the Explore subagent:

- Find relevant files by keywords from user request
- Locate existing SPEC documents in .moai/specs/
- Identify implementation patterns and dependencies
- Discover project configuration files
- Report comprehensive results for Phase 1B context

### Phase 1B: SPEC Planning (Required)

Agent: manager-spec subagent

Input: User request plus Phase 1A results (if executed)

Tasks for manager-spec:

- Analyze project documents (product.md, structure.md, tech.md)
- Propose 1-3 SPEC candidates with proper naming
- Check for duplicate SPECs in .moai/specs/
- Design EARS structure for each candidate
- Create implementation plan with technical constraints
- Identify library versions (production stable only, no beta/alpha)

Output: Implementation plan with SPEC candidates, EARS structure, and technical constraints.

### Decision Point 1: SPEC Creation Approval

Tool: AskUserQuestion (at orchestrator level only)

Options:

- Proceed with SPEC Creation
- Request Plan Modification
- Save as Draft
- Cancel

If "Proceed": Continue to Phase 1.5 then Phase 2.
If "Modify": Collect feedback, re-run Phase 1B with feedback context.
If "Draft": Save plan.md with status draft, create commit, print resume command, exit.
If "Cancel": Discard plan, exit with no files created.

### Phase 1.5: Pre-Creation Validation Gate

Purpose: Prevent common SPEC creation errors before file generation.

Step 1 - Document Type Classification:

- Detect keywords to classify as SPEC, Report, or Documentation
- Reports route to .moai/reports/, Documentation to .moai/docs/
- Only SPEC-type content proceeds to Phase 2

Step 2 - SPEC ID Validation (all checks must pass):

- ID Format: Must match SPEC-{DOMAIN}-{NUMBER} pattern (e.g., SPEC-AUTH-001)
- Domain Name: Must be from the approved domain list (AUTH, API, UI, DB, REFACTOR, FIX, UPDATE, PERF, TEST, DOCS, INFRA, DEVOPS, SECURITY, and others)
- ID Uniqueness: Search .moai/specs/ to confirm no duplicates exist
- Directory Structure: Must create directory, never flat files

Composite domain rules: Maximum 2 domains recommended (e.g., UPDATE-REFACTOR-001), maximum 3 allowed.

### Phase 2: SPEC Document Creation

Agent: manager-spec subagent

Input: Approved plan from Phase 1B, validated SPEC ID from Phase 1.5.

File generation (all three files created simultaneously):

- .moai/specs/SPEC-{ID}/spec.md
  - YAML frontmatter with 7 required fields (id, version, status, created, updated, author, priority)
  - HISTORY section immediately after frontmatter
  - Complete EARS structure with all 5 requirement types
  - Content written in conversation_language

- .moai/specs/SPEC-{ID}/plan.md
  - Implementation plan with task decomposition
  - Technology stack specifications and dependencies
  - Risk analysis and mitigation strategies

- .moai/specs/SPEC-{ID}/acceptance.md
  - Minimum 2 Given/When/Then test scenarios
  - Edge case testing scenarios
  - Performance and quality gate criteria

Quality constraints:

- Requirement modules limited to 5 or fewer per SPEC
- Acceptance criteria minimum 2 Given/When/Then scenarios
- Technical terms and function names remain in English

### Phase 3: Git Environment Setup (Conditional)

Execution conditions: Phase 2 completed successfully AND one of the following:

- --worktree flag provided
- --branch flag provided or user chose branch creation
- Configuration permits branch creation (git_strategy settings)

Skipped when: develop_direct workflow, no flags and user chooses "Use current branch".

#### Worktree Path (--worktree flag)

Prerequisite: SPEC files MUST be committed before worktree creation.

- Stage SPEC files: git add .moai/specs/SPEC-{ID}/
- Create commit: feat(spec): Add SPEC-{ID} - {title}
- Create worktree via WorktreeManager with branch feature/SPEC-{ID}
- Display worktree path and navigation instructions

#### Branch Path (--branch flag or user choice)

Agent: manager-git subagent

- Create branch: feature/SPEC-{ID}-{description}
- Set tracking upstream if remote exists
- Switch to new branch
- Team mode: Create draft PR via manager-git subagent

#### Current Branch Path (no flag or user choice)

- No branch creation, no manager-git invocation
- SPEC files remain on current branch

### Decision Point 2: Development Environment Selection

Tool: AskUserQuestion (when prompt_always config is true and auto_branch is true)

Options:

- Create Worktree (recommended for parallel SPEC development)
- Create Branch (traditional workflow)
- Use current branch

### Decision Point 3: Next Action Selection

Tool: AskUserQuestion (after SPEC creation completes)

Options:

- Start Implementation (execute /moai run SPEC-{ID})
- Modify Plan
- Add New Feature (create additional SPEC)

---

## Team Mode Routing

When --team flag is provided or auto-selected, the plan phase MUST switch to team orchestration:

1. Verify prerequisites: workflow.team.enabled == true AND CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1 env var is set
2. If prerequisites met: Read workflows/team-plan.md and execute the team workflow (TeamCreate with researcher + analyst + architect)
3. If prerequisites NOT met: Warn user with message "Team mode requires workflow.team.enabled: true in workflow.yaml and CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1 env var" then fallback to standard sub-agent mode (manager-spec)

Team composition: researcher (haiku) + analyst (inherit) + architect (inherit)

For detailed team orchestration steps, see workflows/team-plan.md.

---

## Completion Criteria

All of the following must be verified:

- Phase 1: manager-spec analyzed project and proposed SPEC candidates
- User approval obtained via AskUserQuestion before SPEC creation
- Phase 2: All 3 SPEC files created (spec.md, plan.md, acceptance.md)
- Directory naming follows .moai/specs/SPEC-{ID}/ format
- YAML frontmatter contains all 7 required fields
- EARS structure is complete
- Phase 3: Appropriate git action taken based on flags and user choice
- If --worktree: SPEC committed before worktree creation
- Next steps presented to user

---

Version: 2.0.0
Updated: 2026-02-07
Source: Extracted from .claude/commands/moai/1-plan.md v5.1.0. Added team mode support and --team flag.
