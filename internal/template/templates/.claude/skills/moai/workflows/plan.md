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

### Phase 0.3: Clarity Evaluation (Conditional)

Purpose: Evaluate how clearly the user's request is specified before beginning deep research. A vague request produces a weaker SPEC; this phase detects vagueness early and gathers missing context through a structured interview.

**Skip conditions (any one is sufficient):**
- `--skip-interview` flag is present in $ARGUMENTS
- Input matches `resume SPEC-XXX` pattern (resuming an existing draft)
- Input contains 5 or more distinct technical keywords (e.g., framework names, file paths, function names, domain terms)
- `interview.enabled: false` in `.moai/config/sections/interview.yaml`

**Clarity Scoring (1-10):**

Evaluate the user's input against five dimensions:

1. Technical keyword count: 2+ points for 3-4 keywords; 1 point for 1-2; 0 for none
2. Action verbs specificity: "add CRUD endpoints for user profile" scores higher than "improve the app"
3. File or module mentions: explicit file paths or module names each add 1 point
4. Generic nouns penalty: deduct 1 point for each vague noun like "system", "feature", "thing"
5. Scope boundary clarity: a defined boundary ("only the POST /users endpoint, no auth changes") adds 2 points

**Score-to-rounds mapping:**

| Clarity Score | Interview Rounds |
|---|---|
| 1-3 | 0 (request too vague — ask one broad clarification question instead) |
| 4-6 | 2 rounds maximum |
| 7-10 | 5 rounds maximum |

Log the score: "Clarity score: {N}/10 — proceeding with {M} interview round(s)."

If score is 1-3: Use a single AskUserQuestion asking for a clearer description, then re-evaluate. Do not enter the full interview loop.

### Phase 0.3.1: Deep Interview Loop (Conditional)

Purpose: Gather missing context through a structured, topic-focused interview before research begins. Each round presents curated options so the user can answer quickly.

**Entry condition:** Clarity score 4-10 AND skip conditions not met (from Phase 0.3).

**Guard:** [HARD] During the interview loop, the agent MUST NOT write implementation code or start codebase exploration. The sole output is `.moai/specs/SPEC-{ID}/interview.md`.

**Round topics:**

| Round | Focus Topic | Example Questions |
|---|---|---|
| 1 | Scope | What is included and explicitly excluded? |
| 2 | Constraints | Performance, security, compatibility, technology limits |
| 3 | Success criteria | How do we know when this is done and working correctly? |
| 4 | Edge cases | What unusual or failure scenarios must be handled? |
| 5 | Priority | What is the minimum viable slice if scope must be cut? |

**Per-round execution:**

For each round:

1. Formulate 3 recommended options relevant to the current topic and the user's request context.
2. Present via AskUserQuestion with exactly 4 options:
   - Option 1: [Recommended based on context] (Recommended): [Detailed description of this answer]
   - Option 2: [Alternative]: [Description]
   - Option 3: [Alternative]: [Description]
   - Option 4: Type your own answer: Enter a custom response if none of the above match
3. Record the user's answer.
4. Re-evaluate clarity score after each round.
5. If updated clarity score drops to 3 or below: end the loop early (user's answers added no useful information).
6. If updated clarity score reaches 8 or above: end the loop early (sufficient clarity achieved).
7. Display round counter: "Interview round {N}/{max_rounds}"

**Output:** Write all interview answers to `.moai/specs/SPEC-{ID}/interview.md` with this structure:

```
# Interview: {SPEC Title}

## Round 1: Scope
Question: {question asked}
Answer: {user's answer}

## Round 2: Constraints
...

## Clarity Score
Initial: {N}/10
Final: {N}/10
Rounds completed: {N}
```

**Context passing:** Pass `interview.md` to Phase 0.5 (Deep Research) and Phase 1B (SPEC Planning) as additional context. Both agents MUST read interview.md before proceeding.

### Phase 0.4: UltraThink Auto-Activation (Conditional)

Purpose: Automatically activate deep analysis mode for complex SPECs that benefit from structured reasoning.

**Activation condition**: Evaluate task complexity from Phase 1A exploration results or user request:
- Complexity score >= 7 (multi-domain, cross-cutting concerns)
- Request involves architectural decisions (new module, system redesign, migration)
- Request touches security-critical areas (auth, payment, data isolation)
- User explicitly includes `ultrathink` keyword in request

**UltraThink vs --deepthink distinction**:
- `ultrathink`: Claude Code native deep analysis mode — activates extended reasoning within the current agent context. Triggered by keyword detection in user input.
- `--deepthink`: Sequential Thinking MCP tool invocation — programmatic step-by-step analysis via `mcp__sequential-thinking__sequentialthinking`. Triggered by explicit flag.

When UltraThink auto-activates:
- Log: "UltraThink mode activated: [reason]"
- Apply extended reasoning to Phase 0.5 research and Phase 1B SPEC creation
- Produce deeper analysis in research.md with trade-off comparisons and risk assessments
- Consider alternative approaches and document rejection rationale

When --deepthink flag is present (can combine with UltraThink):
- Invoke Sequential Thinking MCP for structured step-by-step analysis
- Each thinking step documented in research.md

**Skip condition**: Simple, well-scoped features (complexity < 5, single domain, clear requirements). Log: "UltraThink skipped: low complexity task."

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

### Delta Markers for Brownfield Projects

When the SPEC modifies existing code (detected via research.md analysis), apply delta markers:

```
### [DELTA] {Module Name}
- [EXISTING] {description} - unchanged context, characterization tests only
- [MODIFY] {description} - existing code to change, requires characterization tests before modification
- [NEW] {description} - new code to create, full implementation + new tests
- [REMOVE] {description} - code to delete, requires dependency analysis and migration verification
```

Delta markers are OPTIONAL and only suggested for brownfield projects. Greenfield projects skip this.

### spec-compact.md Auto-Generation

After all SPEC files are created, auto-generate `.moai/specs/SPEC-{ID}/spec-compact.md`:

Extract from spec.md:
- All REQ-XXX requirements (EARS format entries)
- All acceptance criteria (Given/When/Then scenarios)
- Files to modify list
- Exclusions (What NOT to Build) section

Exclude: Overview, technical approach, research references, annotation history.

Purpose: Run phase loads spec-compact.md (~30% token savings) instead of full spec.md.
Fallback: If generation fails, Run phase uses full spec.md.

Quality constraints:
- Requirement modules limited to 5 or fewer per SPEC
- Acceptance criteria minimum 2 Given/When/Then scenarios
- Technical terms and function names remain in English
- Exclusions section MUST contain at least 1 entry

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

### Phase 3.5: MX Tag Planning [MANDATORY]

Purpose: Identify code locations that will need @MX annotations during implementation. This information is passed to run workflow agents as context constraints.

Execution conditions: Always executed. Depth varies by scope:
- **Full scan**: SPEC involves modifying existing code OR creating new public APIs
- **Lightweight scan**: New feature with no existing code interaction (scan public API surface only)

Tasks:
- Scan target files for high fan_in functions (potential @MX:ANCHOR)
- Identify dangerous patterns (goroutines, complexity) for @MX:WARN
- List magic constants and business rules for @MX:NOTE
- Document MX tag strategy in `plan.md`
- Output: `mx_plan` section in SPEC document with annotation targets and priorities

### Phase 3.6: SPEC Quality Gate

Purpose: Verify SPEC document quality before proceeding to implementation. Catches incomplete or inconsistent specs early.

Tasks:
- Verify all EARS-format requirements have corresponding acceptance criteria
- Check that affected files list is complete (cross-reference with codebase)
- Validate that MX tag plan covers all high-risk areas (fan_in >= 3, goroutines)
- Run lightweight security check on SPEC scope (flag if auth/crypto/input-validation areas are touched)

Gate decision:
- **PASS**: All checks satisfied. Proceed to Decision Point 2.
- **WARNING**: Minor gaps found (e.g., missing acceptance criteria for edge cases). Present findings and offer fix or continue.
- **FAIL**: Critical gaps (e.g., no acceptance criteria, security-sensitive scope without security considerations). Must fix before proceeding.

Tool: AskUserQuestion (when WARNING or FAIL)
Options:
- Fix SPEC issues (Recommended): Return to SPEC editing with specific gaps highlighted
- Continue with warnings: Proceed knowing gaps exist (WARNING only, not available for FAIL)
- Abort: Exit plan workflow

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

Purpose: After SPEC creation, detect execution environment and present optimal implementation mode.

**Step 1: Detect active LLM mode**
Read `.moai/config/sections/llm.yaml` → `llm.team_mode` field:
- `""` (empty) or `"cc"`: CC mode (Claude-only)
- `"glm"`: GLM mode (GLM-only)
- `"cg"`: CG mode (Claude Leader + GLM Workers)

**Step 2: Detect tmux availability**
Check `$TMUX` environment variable via Bash: `test -n "$TMUX" && echo "tmux" || echo "no-tmux"`

**Step 3: Present options based on detection**

When tmux IS available: AskUserQuestion with 3 options (descriptions adapt to active_mode):
- Option 1 (Recommended): Worktree + {active_mode}
  - CC: "Create MoAI worktree with tmux session. All agents use Claude. Highest quality."
  - GLM: "Create MoAI worktree with tmux session. All agents use GLM. Cost optimized."
  - CG: "Create MoAI worktree with tmux session. Leader=Claude, Workers=GLM. Balanced quality-cost."
- Option 2: Team Mode (in-process): Use Agent Teams for parallel implementation within current session. Best for multi-domain features.
- Option 3: Sub-agent Mode (sequential): Use sequential sub-agents. Best for simple, single-domain tasks.

When tmux is NOT available: AskUserQuestion with 2 options:
- Option 1 (Recommended): Sub-agent Mode: Use sequential sub-agents for implementation. Tmux is not available for session isolation.
- Option 2: Team Mode (in-process): Use Agent Teams for parallel implementation within current session.

**Step 4: Execute selected mode**
- **Worktree mode**: Execute `moai worktree new SPEC-{ID} --tmux` to create worktree with tmux session. The tmux session will:
  - CC mode: Create session, cd to worktree, run `/moai run SPEC-{ID}`
  - GLM mode: Create session, inject GLM env, cd to worktree, run `/moai run SPEC-{ID}`
  - CG mode: Create session, inject GLM env to session, clear GLM from settings.local.json, cd to worktree, run `/moai run SPEC-{ID}`
  - Display: "Implementation started in tmux session: moai-{ProjectName}-{SPEC-ID}"
- **Team mode**: Proceed to `/moai run SPEC-{ID} --team`
- **Sub-agent mode**: Proceed to `/moai run SPEC-{ID} --solo`

**Step 5: Gate result passing**
- Pass the selected execution mode to the run workflow
- If worktree mode: Run workflow executes in the tmux session (no further action needed from plan)
- If team/sub-agent mode: Continue to run workflow in current session

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
- Phase 2: All SPEC files created (spec.md, plan.md, acceptance.md, spec-compact.md)
- Directory naming follows .moai/specs/SPEC-{ID}/ format
- YAML frontmatter contains all 8 required fields (including issue_number)
- EARS structure is complete
- Exclusions section present with at least 1 entry
- Delta markers applied for brownfield requirements (if applicable)
- spec-compact.md auto-generated with requirements + acceptance criteria only
- Phase 2.5: GitHub Issue created and linked (unless --no-issue)
- Phase 3: Appropriate git action taken based on flags and user choice
- If --worktree: SPEC committed before worktree creation
- Next steps presented to user

---

## Test Scenarios

### Normal Flow
**Prompt**: "/moai plan JWT authentication with refresh token rotation"
**Expected Result**:
- Phase 1A: Explore discovers existing auth files if any
- Phase 1B: manager-spec designs EARS requirements for JWT auth
- Annotation cycle: 1-3 iterations refining requirements
- Phase 2: SPEC-AUTH-001 created with spec.md, plan.md, acceptance.md
- Phase 2.5: GitHub Issue created and linked to SPEC
- Phase 3: Feature branch feat/SPEC-AUTH-001-jwt-auth created (if --branch)

### Existing Assets Flow
**Prompt**: "/moai plan add payment gateway" (existing e-commerce codebase)
**Expected Result**:
- Explore discovers existing order, product, user models
- SPEC references existing models as dependencies
- plan.md identifies extension points in existing architecture
- No duplicate functionality proposed

### Error Flow
**Prompt**: "/moai plan" (no description provided)
**Expected Result**:
- AskUserQuestion prompts user for feature description
- After user provides description, normal flow continues
- If user cancels, graceful exit with no files created

---

Version: 2.8.0
Updated: 2026-03-30
Changes: Added test scenarios, Phase 0.9 JIT Language Detection.
