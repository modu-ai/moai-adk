---
name: moai-workflow-plan
description: >
  Creates comprehensive SPEC documents using EARS format as the first step
  of the Plan-Run-Sync workflow. Handles project exploration, SPEC file
  generation, validation, and optional Git environment setup with worktree
  or branch creation. Use when planning features or creating specifications.
user-invocable: false
metadata:
  version: "2.6.0"
  category: "workflow"
  status: "active"
  updated: "2026-02-23"
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

Create comprehensive SPEC documents using EARS format as the first step of the Plan-Run-Sync workflow.

For phase overview and token budgets, see: .claude/rules/moai/workflow/spec-workflow.md

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
- --team: Enable team-based exploration (see ${CLAUDE_SKILL_DIR}/team/plan.md for parallel research team)
- --no-issue: Skip GitHub Issue creation after SPEC generation
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
- .moai/project/codemaps/ directory listing (architecture maps for existing codebase understanding)
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
- If .moai/project/codemaps/ exists: Use as architecture baseline to accelerate exploration
- Find relevant files by keywords from user request
- Locate existing SPEC documents in .moai/specs/
- Identify implementation patterns and dependencies
- Discover project configuration files
- Read target directories in depth — understand deeply how each module works, its intricacies and side effects
- Study cross-module interactions in great detail — trace data flow through the system
- Go through related test files to understand expected behavior and edge cases
- Report comprehensive results for Phase 1B context

### Phase 0.5: Deep Research (Recommended)

Agent: Explore subagent (deep codebase analysis)

Purpose: Produce a persistent research.md artifact documenting deep codebase understanding. This document serves as a verification surface — MoAI and the user can review it and correct misunderstandings before planning begins.

When to run:
- Feature involves modifying existing code
- Feature has cross-module dependencies
- User explicitly requests research phase

When to skip:
- Simple, isolated additions (new file with no dependencies)
- User provides explicit "skip research" instruction

Tasks for the Explore subagent:
- Read target code areas in depth — understand how they work deeply, their intricacies and specificities
- Study related systems in great detail — trace data flow, identify implicit contracts and side effects
- Discover reference implementations in the existing codebase — find similar patterns that can guide the new implementation
- Search for relevant open-source examples or documented patterns that align with the project's conventions
- Document all findings in a structured research.md file

Research directives (Deep Reading patterns):
- Use language that demands thoroughness: "read deeply", "study in great detail", "understand the intricacies"
- Avoid surface-level scanning — agent must trace through actual execution paths
- Every finding must include specific file paths and line references

Output: `.moai/specs/SPEC-{ID}/research.md` containing:
- Architecture analysis with file paths and dependency maps
- Existing patterns and conventions discovered
- Reference implementations found (internal codebase or documented patterns)
- Risks, constraints, and implicit contracts identified
- Recommendations for the implementation approach

### Phase 1.25: Design Direction (Conditional)

Purpose: Establish design intent and direction for UI/UX-related SPECs before SPEC planning begins. Based on the Intent-First design philosophy from the interface-design methodology.

When to run:
- SPEC description contains 2+ UI/UX keywords: ui, frontend, interface, design, component, page, screen, layout, form, dashboard, button, modal, view, sidebar, navigation, widget, chart, table
- User explicitly requests design direction

When to skip:
- No UI/UX keywords detected in SPEC description
- User explicitly requests "skip design" or uses --prototype flag
- Backend-only, infrastructure, or documentation SPECs

Agent: expert-frontend subagent (with moai-design-craft skill)

Tasks:
1. Check if `.moai/design/system.md` exists and has content
2. If system.md exists: Load as design context, skip Intent-First process
3. If system.md is empty or missing: Execute Intent-First process:
   - Answer: Who is this human? What must they accomplish? What should this feel like?
   - Produce domain exploration: 5+ domain concepts, 5+ color world entries, 1 signature element
   - Identify 3+ defaults to avoid (generic patterns to reject)
4. Generate design direction artifact

Output: `.moai/specs/SPEC-{ID}/design-direction.md` containing:
- Intent statement (who, what, feel)
- Domain concepts and vocabulary
- Color world exploration
- Signature element definition
- Defaults to avoid
- Reference to `.moai/design/system.md` if exists

Design direction guard: [HARD] During Phase 1.25, the agent MUST NOT write implementation code. Focus exclusively on design exploration and direction definition.

After Phase 1.25: Offer to persist design decisions to `.moai/design/system.md` if it was newly created or updated. Use AskUserQuestion: "Save design direction to project-level design memory (.moai/design/system.md)?"

### Phase 1B: SPEC Planning (Required)

Agent: manager-spec subagent

Input: User request plus Phase 1A results (if executed), plus design-direction.md (if Phase 1.25 executed)

Tasks for manager-spec:
- Analyze project documents (product.md, structure.md, tech.md)
- Propose 1-3 SPEC candidates with proper naming
- Check for duplicate SPECs in .moai/specs/
- Design EARS structure for each candidate
- Create implementation plan with technical constraints
- Identify library versions (production stable only, no beta/alpha)
- Search for reference implementations: Identify similar patterns in the existing codebase or well-documented approaches that can guide implementation
- When reference implementations are found, include them in the plan as "Reference: {file_path}:{line_range}" to improve implementation quality

Output: Implementation plan with SPEC candidates, EARS structure, and technical constraints.

Implementation guard: [HARD] During Phases 0.5, 1A, and 1B, all agent prompts MUST include the instruction: "DO NOT write implementation code. Focus exclusively on research, analysis, and planning." This separation of thinking and typing is the foundation of effective AI-assisted development.

### Decision Point 1: Plan Review and Annotation Cycle

Tool: AskUserQuestion (at orchestrator level only)

Options:
- Proceed with SPEC Creation (Recommended): Plan is approved, continue to Phase 1.5 then Phase 2
- Annotate Plan: Add inline notes to plan.md for revision (starts annotation cycle)
- Save as Draft: Save plan.md with status draft, create commit, print resume command, exit
- Cancel: Discard plan, exit with no files created

If "Proceed": Continue to Phase 1.5 then Phase 2.
If "Annotate": Enter Annotation Cycle (see below).
If "Draft": Save plan.md with status draft, create commit, print resume command, exit.
If "Cancel": Discard plan, exit with no files created.

#### Annotation Cycle (1-6 iterations)

Purpose: Allow users to iteratively refine the plan through inline notes before any code is written. This prevents expensive failures by catching architectural misunderstandings, missed conventions, and scope issues early.

Process:
1. User reviews plan.md (and research.md if available) in their editor
2. User adds inline notes directly into the document (e.g., "NOTE: use drizzle:generate for migrations, not raw SQL")
3. User signals completion via AskUserQuestion
4. MoAI delegates to manager-spec subagent: "Address all inline notes in the plan document and update it accordingly. DO NOT implement any code."
5. manager-spec updates plan.md, removing addressed notes and incorporating feedback
6. MoAI presents updated plan to user for another review cycle

Iteration limits:
- Maximum 6 annotation cycles per plan
- After each cycle, present options: Proceed / Annotate Again / Save Draft / Cancel
- Track iteration count and display: "Annotation cycle {N}/6"

Guard rule: [HARD] During annotation cycles, the explicit instruction "DO NOT implement any code — only update the plan document" MUST be included in every agent prompt. This prevents premature code generation.

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

Composite domain rules: Maximum 2 domains recommended, maximum 3 allowed.

### Phase 2: SPEC Document Creation

Agent: manager-spec subagent

Input: Approved plan from Phase 1B, validated SPEC ID from Phase 1.5.

File generation (all three files created simultaneously):

- .moai/specs/SPEC-{ID}/spec.md
  - YAML frontmatter with 8 required fields (id, version, status, created, updated, author, priority, issue_number)
  - issue_number: GitHub Issue number linked to this SPEC (0 if --no-issue or Issue creation skipped)
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

### Phase 2.5: GitHub Issue Creation (Conditional)

Purpose: Create a GitHub Issue linked to the SPEC document for bidirectional traceability between planning artifacts and issue tracker.

Execution conditions:
- `--no-issue` flag is NOT set
- GitHub CLI (`gh`) is available
- Repository has a remote origin

Skip conditions:
- `--no-issue` flag is set
- `gh` CLI not available (log warning and continue)
- No remote origin configured

#### Step 2.5.1: Create GitHub Issue

Agent: manager-git subagent

Create a GitHub Issue from SPEC metadata:

```bash
gh issue create \
  --title "[SPEC-{ID}] {SPEC title}" \
  --body "$(cat <<'EOF'
## SPEC Reference

- **SPEC ID**: SPEC-{ID}
- **Status**: draft
- **Priority**: {priority}
- **Created**: {created_date}

## Requirements Summary

{Brief summary from spec.md EARS requirements}

## Acceptance Criteria

{Summary from acceptance.md}

---

*This issue was automatically created by MoAI from SPEC-{ID}.*
*SPEC location: `.moai/specs/SPEC-{ID}/spec.md`*
EOF
)" \
  --label "spec"
```

#### Step 2.5.2: Update SPEC Metadata

After Issue creation, update the SPEC frontmatter with the issue number:

- Read the issue number from `gh issue create` output
- Update `issue_number` field in `.moai/specs/SPEC-{ID}/spec.md` YAML frontmatter
- Add cross-reference comment in the Issue: `gh issue comment {number} --body "SPEC document: .moai/specs/SPEC-{ID}/spec.md"`

#### Step 2.5.3: Bidirectional Reference

The SPEC ↔ Issue link enables:
- SPEC spec.md frontmatter contains `issue_number: {N}` for downstream workflows
- GitHub Issue body contains SPEC-ID and file path for human navigation
- run.md Phase 3 uses `issue_number` to include `Fixes #{N}` in commits/PRs
- sync.md leverages `Fixes #{N}` in PR for automatic Issue closure on merge

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
- Create worktree: `moai worktree new SPEC-{ID}`
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

### Phase 3.5: MX Tag Planning (Optional)

Purpose: Identify code locations that will need @MX annotations during implementation.

Execution conditions: SPEC involves modifying existing code OR creating new public APIs.

Tasks:
- Scan target files for high fan_in functions (potential @MX:ANCHOR)
- Identify dangerous patterns (goroutines, complexity) for @MX:WARN
- List magic constants and business rules for @MX:NOTE
- Document MX tag strategy in plan.md

Skip conditions: New feature with no existing code interaction.

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

### Decision Point 3.5: Execution Mode Selection Gate

Triggered when: User selects "Start Implementation" in Decision Point 3.

**Step 1 — Detect active mode:**
Read `.moai/config/sections/llm.yaml` → `llm.team_mode` field:
- `""` (empty) = CC mode (all agents use Claude)
- `"glm"` = GLM mode (all agents use GLM)
- `"cg"` = CG mode (Leader=Claude, Workers=GLM)

**Step 2 — Detect tmux availability:**
Bash: `test -n "$TMUX" && echo "tmux" || echo "no-tmux"`

**Step 3 — Present options when tmux is available:**
AskUserQuestion with 3 options (descriptions adapt to active_mode):
- Option 1 (Recommended): Worktree + {active_mode}
  - CC: "독립 worktree에서 CC 모드 실행. 모든 에이전트 Claude. 최고 품질."
  - GLM: "독립 worktree에서 GLM 모드 실행. 모든 에이전트 GLM. 비용 최적화."
  - CG: "독립 worktree에서 CG 모드 실행. Leader=Claude, Workers=GLM. 품질-비용 균형."
- Option 2: Team Mode — 현재 세션에서 Agent Teams 실행. Worktree 없이 직접 실행.
- Option 3: Sub-agent Mode — 순차 실행. 가장 안정적이고 토큰 효율적.

**Step 3 (tmux unavailable):** AskUserQuestion with 2 options:
- Option 1 (Recommended): Sub-agent Mode — 순차 실행. tmux 없이 가장 안정적.
- Option 2: Team Mode (in-process) — 현재 세션에서 Agent Teams 실행.

**Step 4 — Worktree 선택 시 실행:**
- CC: 추가 env 설정 불필요. worktree 생성 후 새 tmux 세션에서 claude 실행.
- GLM: 새 tmux 세션에 injectTmuxSessionEnv()로 GLM env 주입 후 실행.
- CG: 새 tmux 세션에 injectTmuxSessionEnv() 적용 + settings.local.json에서 GLM env 제거(Leader 격리).
- 새 tmux 세션에서 worktree 디렉터리로 이동 후 `/moai run SPEC-{ID}` 실행.
- 현재 세션 종료 (worktree 세션이 독립적으로 실행됨).

**Step 5 — Gate 결과를 run 워크플로우에 전달:**
- `execution_mode`: worktree | team | sub-agent
- `active_mode`: cc | glm | cg
- `tmux_available`: true | false

---

## Team Mode Routing

When --team flag is provided or auto-selected, the plan phase MUST switch to team orchestration:

1. Verify prerequisites: workflow.team.enabled == true AND CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1 env var is set
2. If prerequisites met: Read ${CLAUDE_SKILL_DIR}/team/plan.md and execute the team workflow (TeamCreate with researcher + analyst + architect)
3. If prerequisites NOT met: Warn user then fallback to standard sub-agent mode (manager-spec)

Team composition: researcher (haiku) + analyst (inherit) + architect (inherit)

For detailed team orchestration steps, see ${CLAUDE_SKILL_DIR}/team/plan.md.

---

## Completion Criteria

All of the following must be verified:

- Phase 1: manager-spec analyzed project and proposed SPEC candidates
- User approval obtained via AskUserQuestion before SPEC creation
- Phase 2: All 3 SPEC files created (spec.md, plan.md, acceptance.md)
- Directory naming follows .moai/specs/SPEC-{ID}/ format
- YAML frontmatter contains all 8 required fields (including issue_number)
- EARS structure is complete
- Phase 2.5: GitHub Issue created and linked (unless --no-issue)
- Phase 3: Appropriate git action taken based on flags and user choice
- If --worktree: SPEC committed before worktree creation
- Next steps presented to user

---

Version: 2.7.0
Updated: 2026-03-11
Changes: Added Phase 2.5 GitHub Issue creation with bidirectional SPEC-Issue linking, --no-issue flag, issue_number SPEC frontmatter field.
