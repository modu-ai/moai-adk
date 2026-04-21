---
name: moai-workflow-run
description: >
  DDD/TDD implementation workflow for SPEC requirements. Second step
  of the Plan-Run-Sync workflow. Routes to manager-ddd or manager-tdd based
  on quality.yaml development_mode setting.
user-invocable: false
metadata:
  version: "2.6.0"
  category: "workflow"
  status: "active"
  updated: "2026-02-23"
  tags: "run, implementation, ddd, tdd, spec"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["run", "implement", "build", "create", "develop", "code"]
  agents: ["manager-ddd", "manager-tdd", "manager-strategy", "manager-quality", "manager-git"]
  phases: ["run"]
---

# Run Workflow Orchestration

## Purpose

Implement SPEC requirements using the configured development methodology.

For methodology details (DDD ANALYZE-PRESERVE-IMPROVE and TDD RED-GREEN-REFACTOR cycles, success criteria, brownfield enhancement), see: .claude/rules/moai/workflow/workflow-modes.md

## Scope

- Implements Step 3 of MoAI's 4-step workflow (Task Execution)
- Receives SPEC documents created by /moai plan
- Hands off to /moai sync for documentation and PR

## Input

- $ARGUMENTS: SPEC-ID to implement (e.g., SPEC-AUTH-001)
- Resume: Re-running /moai run SPEC-XXX resumes from last successful phase checkpoint
- --team: Enable team-based implementation (see ${CLAUDE_SKILL_DIR}/team/run.md for parallel implementation team)

## UltraThink Auto-Activation

When the run phase begins, evaluate whether to activate deep analysis mode for the strategy phase:

**Activation condition** (any of):
- SPEC spans >= 2 distinct domains (backend + frontend, auth + database, etc.)
- SPEC plan.md lists >= 8 files to create or modify
- SPEC involves architectural patterns (new module, service, middleware layer)
- User explicitly includes `ultrathink` keyword

**UltraThink vs --deepthink**:
- `ultrathink`: Extended reasoning within the current agent — deeper strategy analysis, more thorough trade-off evaluation
- `--deepthink`: Sequential Thinking MCP invocation — structured step-by-step analysis via `mcp__sequential-thinking__sequentialthinking`

When activated: Apply to Phase 1 (Strategy) for deeper architectural analysis. Log: "UltraThink mode activated for strategy phase: [reason]"

## Harness Level Routing

At Run phase entry, determine the pipeline depth:

1. Receive harness level from orchestrator (moai.md Complexity Estimator) or default to standard
2. Apply level-specific phase configuration:
   - **minimal**: Skip phases [0, 0.5, 2.0, 2.5, 2.75, 2.8a, 2.9, 2.10]. Direct implementation only.
   - **standard**: Execute all phases. evaluator-active in final-pass mode (Phase 2.8a only).
   - **thorough**: Execute all phases. evaluator-active in per-sprint mode (Phase 2.0 + 2.8a). Sprint contract enabled.
3. Load SPEC context (token-efficient):
   - If `.moai/specs/SPEC-{ID}/spec-compact.md` exists: Load spec-compact.md (~30% token savings)
   - Otherwise: Load full spec.md (backward compatible)
4. Log harness level to progress.md for traceability

Escalation: If a quality gate fails during execution, escalate harness level:
- minimal → standard (on Phase 2.5 fail)
- standard → thorough (on Phase 2.8a CRITICAL)
- Maximum 2 escalations per SPEC run

## Context Loading

Before execution, load these essential files:

- .moai/config/config.yaml (git strategy, automation settings)
- .moai/config/sections/quality.yaml (coverage targets, TRUST 5 settings, development_mode)
- .moai/config/sections/harness.yaml (harness depth levels, auto-detection rules)
- .moai/config/sections/git-strategy.yaml (auto_branch, branch creation policy)
- .moai/config/sections/language.yaml (git_commit_messages setting)
- .moai/specs/SPEC-{ID}/ directory (spec-compact.md preferred, or spec.md, plan.md, acceptance.md)
- .moai/specs/SPEC-{ID}/progress.md (session resume context: if exists, load to identify completed phases and skip them; if absent, will be created at Phase 1 start)
- .moai/specs/SPEC-{ID}/tasks.md (task decomposition with planned files, if exists)
- .moai/project/structure.md (architecture context for implementation decisions)
- .moai/project/tech.md (technology stack context)
- .moai/project/codemaps/ directory listing (architecture maps for dependency and module understanding)

Pre-execution commands: git status, git branch, git log, git diff.

### Lessons Loading (REQ-SLQG-013)

Before spawning implementation agents, load relevant lessons from auto-memory:

1. Read `~/.claude/projects/{project-hash}/memory/lessons.md` if it exists
2. Filter lessons by domain relevance:
   - Match lesson categories against SPEC domain keywords
   - Match lesson tags against modified file paths (from SPEC scope)
   - Limit to top 5 most recent matching lessons
3. Include filtered lessons in agent spawn prompt as "Previous lessons learned" context
4. Maximum 2000 tokens for lesson injection
5. If lessons.md does not exist or no relevant lessons found, skip silently

### Resume Check

Before Phase 1, check if `.moai/specs/SPEC-{ID}/progress.md` exists:
- If it exists: Load content, identify last completed phase checkpoint, skip all completed phases, resume from the next pending phase. Log: "Resuming SPEC-{ID} from Phase {N}"
- If it does not exist: Create the file now with initial entry:
  ```
  ## SPEC-{ID} Progress

  - Started: {current timestamp}
  ```
- The progress.md file persists across sessions and enables seamless resume after interruption.

---

## Worktree Path Rules [HARD] (All Modes)

When delegating to ANY agent with `isolation: "worktree"` (sub-agent mode or team mode):

- [HARD] Reference all write-target files by project-root-relative paths (e.g., `src/auth/handler.go`)
- [HARD] Do NOT include absolute paths (e.g., `/Users/.../project/src/auth/handler.go`) in agent prompts
- [HARD] Do NOT include `cd /absolute/path &&` in any Bash commands within agent prompts
- [HARD] SPEC files: use `.moai/specs/SPEC-XXX/spec.md` (relative), not absolute paths
- [HARD] The agent's CWD is automatically set to the worktree root by Claude Code — all relative paths resolve correctly

Anti-patterns that bypass worktree isolation:
```
# WRONG: Absolute path bypasses worktree
"Read /Users/user/project/src/auth/handler.go and fix the bug"

# WRONG: cd to main project in Bash command
"Run: cd /Users/user/project && go test ./..."

# CORRECT: Relative path — agent resolves from its own CWD (worktree root)
"The bug is in src/auth/handler.go. Read the file and fix it."

# CORRECT: No cd prefix — agent CWD is already worktree root
"Run: go test ./..."
```

See `.claude/rules/moai/workflow/worktree-integration.md` for complete path rules.

---

## Phase Sequence

All phases execute sequentially. Each phase receives outputs from all previous phases as context.

### Phase 0.5: Environment Assessment (Conditional)

Condition: Only executes when `memory_guard.enabled: true` in quality.yaml.
If memory_guard is not enabled or not present, skip to Phase 1.

Purpose: Detect available system memory and determine test execution strategy to prevent OOM.

Steps:
1. Read memory_guard configuration from quality.yaml
2. Detect available system memory:
   - Linux: `free -m | awk '/^Mem:/{print $7}'` (MemAvailable)
   - macOS: `sysctl -n hw.memsize` (total memory in bytes, divide by 1048576 for MB) and `vm_stat | awk '/Pages free/{print $3}'` (approximate available)
3. Compare available memory against thresholds:
   - Below emergency_threshold_mb: BLOCK test execution, warn user, suggest closing other applications or increasing memory
   - Below adaptive_threshold_mb: Set test_execution_strategy to memory_guard.test_split_strategy (default: "module")
   - Above adaptive_threshold_mb: Set test_execution_strategy to "full" (normal execution)
4. Pass test_execution_strategy as context to all subsequent phases via agent prompt

Output: test_execution_strategy ("full", "module", "changed") passed to Phase 1+ as binding context.

Progress update: Append to `.moai/specs/SPEC-{ID}/progress.md`:
```
- Phase 0.5 complete: memory_guard={enabled|disabled}, available_mb={N}, strategy={full|module|changed}
```

### Phase 0.9: JIT Language Skill Detection

Purpose: Detect the project's primary language and prepare the appropriate language skill reference for agent spawn prompts. Since language skills are not statically bound to agents, the orchestrator must inject them at spawn time.

Steps:
1. Check project root for language indicator files:
   - go.mod → moai-lang-go
   - package.json with "typescript" in devDependencies → moai-lang-typescript
   - package.json without typescript → moai-lang-javascript
   - pyproject.toml or requirements.txt → moai-lang-python
   - Cargo.toml → moai-lang-rust
   - pom.xml or build.gradle → moai-lang-java
   - build.gradle.kts → moai-lang-kotlin
   - *.csproj or *.sln → moai-lang-csharp
   - Gemfile → moai-lang-ruby
   - mix.exs → moai-lang-elixir
   - build.sbt → moai-lang-scala
   - Package.swift → moai-lang-swift
   - pubspec.yaml → moai-lang-flutter
   - DESCRIPTION (with R content) → moai-lang-r
   - CMakeLists.txt or *.cpp → moai-lang-cpp
2. Store the detected language skill name(s) as context for subsequent phases
3. When spawning any expert or manager agent, include in the prompt: "Load Skill({detected-language-skill}) for language-specific patterns and conventions."
4. If multiple languages detected (e.g., monorepo), include all relevant language skills

Output: detected_language_skills list passed to all subsequent agent spawn prompts.

This phase always executes and does NOT require user approval.

### Phase 0.95: Scale-Based Execution Mode Selection

Purpose: Automatically select the optimal execution mode based on task scope, preventing over-engineering for simple tasks and under-resourcing for complex ones.

Mode Selection Rules:

| Request Pattern | Detection Criteria | Execution Mode | Agents |
|----------------|-------------------|---------------|--------|
| Bug fix / error fix | SPEC scope ≤ 3 files, single domain | **Fix Mode** | expert-debug + expert-testing |
| Single endpoint / function | SPEC scope ≤ 5 files, single domain | **Focused Mode** | relevant expert + expert-testing |
| Feature across 1 domain | SPEC scope 5-10 files, single domain | **Standard Mode** | manager-strategy + relevant expert + manager-quality |
| Multi-domain feature | SPEC scope ≥ 10 files OR ≥ 3 domains | **Full Pipeline** | All agents (strategy + backend + frontend + testing + quality + docs) |
| Large cross-cutting change | complexity score ≥ 7 AND --team flag | **Team Mode** | 3-4 parallel teammates |

Detection Steps:
1. Count files referenced in SPEC requirements and plan
2. Identify domains touched (backend, frontend, database, infra, docs)
3. Assess complexity from SPEC priority and acceptance criteria count
4. Select mode based on the table above
5. Log selected mode: "Scale-based mode: {mode} (files: {N}, domains: {N})"

This phase auto-selects and does NOT require user approval. The user can override with --team or --solo flags.

### Phase 1: Analysis and Planning

Agent: manager-strategy subagent

Input: SPEC document content from the provided SPEC-ID. If research.md exists in the SPEC directory (.moai/specs/SPEC-{ID}/research.md), include it as additional context for deeper understanding of the codebase architecture, reference implementations, and identified risks.

Tasks for manager-strategy:

- Read and fully analyze the SPEC document
- Extract requirements and success criteria
- Identify implementation phases and individual tasks
- Determine tech stack and dependencies required
- Estimate complexity and effort
- Create detailed execution strategy with phased approach

Output: Execution plan containing plan_summary, requirements list, success_criteria, and effort_estimate.

Implementation guard: [HARD] During Phase 1 (Analysis and Planning), the manager-strategy subagent MUST NOT write any implementation code. The explicit instruction "DO NOT implement any code — focus exclusively on analysis and planning" MUST be included in the agent prompt. This separation of thinking and execution prevents premature implementation and ensures the plan is reviewed before any code is written.

### Decision Point 1: Plan Approval

<!-- moai:evolvable-start id="gate-run-1" -->
### HUMAN GATE: Plan Approval

**Previous phase output:** Analysis and implementation plan with task decomposition
**Approval question:** Is the implementation plan correct and complete?
**Cannot proceed until:**
- [ ] Plan covers all SPEC acceptance criteria
- [ ] Task decomposition respects Multi-File Decomposition rule (>3 files = split)
- [ ] User has approved the approach
<!-- moai:evolvable-end -->

Tool: AskUserQuestion (at orchestrator level)

Before presenting options, verify the plan against these criteria:

- Proportionality: Is the plan proportional to the requirements? Flag plans with excessive abstraction layers, unnecessary patterns, or scope creep beyond SPEC requirements.
- Code Reuse: Has the plan identified existing code, libraries, or patterns that can be reused? Flag plans that reinvent existing functionality.
- Reference Implementations: Has the plan leveraged reference implementations from research.md? Flag plans that ignore available reference code in the codebase.
- Simplicity: Does the plan follow YAGNI (You Aren't Gonna Need It)? Flag speculative features not in the SPEC.

Options:

- Proceed with plan (continue to Phase 1.5)
- Modify plan (collect feedback, re-run Phase 1)
- Postpone (exit, continue later)

If user does not select "Proceed": Exit execution.

### Phase 1.5: Task Decomposition

Agent: manager-strategy subagent (continuation)

Purpose: Decompose the approved execution plan into atomic, reviewable tasks following SDD 2025 standard.

Tasks for manager-strategy:

- Decompose plan into atomic implementation tasks
- Each task must be completable in a single DDD/TDD cycle
- Assign priority and dependencies for each task
- Generate task tracking entries for progress visibility
- Verify task coverage matches all SPEC requirements

Task structure for each decomposed task:

- Task ID: Sequential within SPEC (TASK-001, TASK-002, etc.)
- Description: Clear action statement
- Requirement Mapping: Which SPEC requirement it fulfills
- Dependencies: List of prerequisite tasks
- Acceptance Criteria: How to verify completion

Constraints: Decompose into atomic tasks where each task completes in a single DDD/TDD cycle. No artificial limit on task count. If the SPEC itself is too complex, consider splitting the SPEC.

Output: Task list with coverage_verified flag set to true.

#### tasks.md Generation (Persistent Artifact)

After task decomposition, generate `.moai/specs/SPEC-{ID}/tasks.md` for audit trail:

```
## Task Decomposition
SPEC: {SPEC-ID}

| Task ID | Description | Requirement | Dependencies | Planned Files | Status |
|---------|-------------|-------------|--------------|---------------|--------|
| T-001 | {desc} | REQ-001 | - | file1.go, file2.go | pending |
| T-002 | {desc} | REQ-002 | T-001 | file3.go | pending |
```

This file is git-tracked. Update task status as implementation progresses.
The planned_files column is used by the Drift Guard (Phase 2A/2B) to detect scope drift.

### Phase 1.6: Acceptance Criteria Initialization (Failing Checklist)

Purpose: Convert all SPEC acceptance criteria into explicit pending TaskList entries. This creates a visible "failing checklist" — all items start as pending and are marked completed (passing) as implementation progresses, following the Harness Engineering pattern.

Action:
- Read spec.md acceptance criteria for SPEC-{ID}
- For each acceptance criterion, execute TaskCreate:
  - subject: `[AC-N] <acceptance criterion statement>`
  - description: Requirement reference, expected behavior, verification method
  - status: pending (starts as "failing")
- Verify all SPEC requirements are covered by at least one task

Output: TaskList populated with all acceptance criteria as pending items.

Progress update: Append to `.moai/specs/SPEC-{ID}/progress.md`:
```
- Phase 1.6 complete: {N} acceptance criteria registered as pending tasks
```

### Phase 1.7: File Structure Scaffolding

Purpose: Create empty file stubs for all planned new files before implementation begins. This prevents entropy by establishing structure before coding, following the Harness Engineering "Blueprint" pattern.

Condition: Execute only for planned new files (files that do not yet exist in the codebase). Skip if all planned files already exist (modification-only SPEC).

Action:
- Identify all planned new files from Phase 1.5 task decomposition
- For each planned new file that does not yet exist:
  - Create empty stub with minimal required structure matching the project's language conventions (e.g., package declaration for Go, module header for Python, empty class for TypeScript)
  - Do NOT add any implementation logic — stubs only
- After stub creation: Capture LSP baseline (this is the clean baseline before any implementation)

Output: List of stub files created, LSP baseline diagnostics captured.

Progress update: Append to `.moai/specs/SPEC-{ID}/progress.md`:
```
- Phase 1.7 complete: {N} stub files created, LSP baseline captured
```

### Phase 1.8: Pre-Implementation MX Context Scan

Purpose: Scan files that will be modified during implementation to build an MX context map for implementation agents.

**Scan Target:** All existing files listed in the task decomposition (from Phase 1.5).

**MX Context Extraction:**
- @MX:ANCHOR: Identify invariant contracts. Pass to implementation agents as "do not break" constraints with fan_in counts.
- @MX:WARN: Identify danger zones. Alert agents to approach these areas with extra caution.
- @MX:NOTE: Collect business logic context. Include in agent prompts for informed implementation.
- @MX:TODO: Match against SPEC requirements. If a TODO aligns with a task, the implementation resolves it.
- @MX:LEGACY: Identify legacy code without SPEC. Flag for careful handling during modifications.

**Output:** MX context map included in Phase 2 agent prompts. The map is structured per-file:
- file_path: list of tags with type, line, description, and constraints

**Skip Condition:** If target files do not exist (greenfield implementation), skip this phase.

See .claude/rules/moai/workflow/mx-tag-protocol.md for tag type definitions.

### Development Mode Routing

Before Phase 2, determine the development methodology by reading `.moai/config/sections/quality.yaml`:

**If development_mode is "ddd":**
- Route all tasks to manager-ddd subagent
- Use ANALYZE-PRESERVE-IMPROVE cycle (see @workflow-modes.md for details)

**If development_mode is "tdd":**
- Route all tasks to manager-tdd subagent
- Use RED-GREEN-REFACTOR cycle (see @workflow-modes.md for details)

### Phase 2.0: Sprint Contract Negotiation

**Condition**: Execute only when harness level = thorough.
**Skip**: When harness level = minimal or standard.

Steps:
1. Load implementation plan from Phase 1.5 task decomposition
2. Invoke evaluator-active to review the plan:
   - Identify missing edge cases in proposed test coverage
   - Flag security concerns in the implementation approach
   - Verify acceptance criteria are specific and testable
3. evaluator-active produces contract proposal with:
   - Done criteria (specific test cases that must pass)
   - Edge cases identified for coverage
   - Hard thresholds (coverage %, performance targets, security requirements)
4. Record agreed contract in `.moai/specs/SPEC-{ID}/contract.md`
5. Maximum 2 negotiation rounds. If no agreement after 2 rounds, proceed with evaluator's recommendations as the contract.

Mode-specific deployment:
- Sub-agent mode: Agent(subagent_type="evaluator-active")
- Team mode: SendMessage to reviewer teammate
- CG mode: Leader performs contract negotiation inline

**Output**: `.moai/specs/SPEC-{ID}/contract.md`

### Delta Marker Detection (Brownfield Pre-Check)

Before routing to Phase 2A or 2B, scan the loaded SPEC for `[DELTA]` section markers:

1. Check spec.md (or spec-compact.md) for any line matching `[EXISTING]`, `[MODIFY]`, `[NEW]`, or `[REMOVE]`
2. If NO delta markers found: skip this section, proceed to Phase 2A/2B normally (greenfield path)
3. If delta markers found: activate delta-aware routing as follows

**Delta-aware routing rules (applied within DDD or TDD mode):**

| Marker | Treatment | Action |
|--------|-----------|--------|
| `[EXISTING]` | Context only — do not modify | Generate characterization tests to document current behavior; no code changes |
| `[MODIFY]` | Modify with safety net | Generate characterization tests FIRST, verify they pass, THEN apply modifications |
| `[NEW]` | Full implementation | Apply complete DDD ANALYZE-PRESERVE-IMPROVE or TDD RED-GREEN-REFACTOR cycle |
| `[REMOVE]` | Safe deletion | Check all callers and dependents; confirm no active references; then remove |

**Delta processing order** (prevents regression):
1. Process all `[EXISTING]` items — characterization tests only
2. Process all `[MODIFY]` items — characterization tests → modification → verify tests still pass
3. Process all `[NEW]` items — full implementation cycle
4. Process all `[REMOVE]` items — dependency analysis → safe deletion

If no delta markers are present in the SPEC, delta processing is silently skipped and the standard implementation flow proceeds unchanged (backward compatible with greenfield SPECs).

### Phase 2: Implementation (Mode-Dependent)

**[HARD] Worktree Prompt Construction**: When spawning implementation agents (manager-ddd, manager-tdd) with `isolation: "worktree"`, the orchestrator MUST construct prompts using project-root-relative paths only. Do NOT embed the current working directory path in the agent prompt. See "Worktree Path Rules [HARD]" section above.

#### Phase 2A: DDD Implementation (for ddd mode)

Agent: manager-ddd subagent

Input: Approved execution plan from Phase 1 plus task decomposition from Phase 1.5. Include `.moai/project/structure.md` and `.moai/project/tech.md` as onboarding context in the agent prompt so the implementation agent understands the project's architecture conventions before writing code.

Requirements:

- Initialize task tracking for progress across refactoring steps
- Execute the complete ANALYZE-PRESERVE-IMPROVE cycle
- Verify all existing tests pass after each transformation
- Create characterization tests for uncovered code paths
- Ensure test coverage meets or exceeds 85%

Output: files_modified list, characterization_tests_created list, test_results (all passing), behavior_preserved flag, structural_metrics comparison, implementation_divergence report.

Implementation Divergence Tracking:

The manager-ddd subagent must track deviations from the original SPEC plan during implementation:

- planned_files: Files listed in plan.md that were expected to be created or modified
- actual_files: Files actually created or modified during the DDD cycle
- additional_features: Features or capabilities implemented beyond the original SPEC scope (with rationale)
- scope_changes: Description of any scope adjustments made during implementation (expansions, deferrals, or substitutions)
- new_dependencies: Any new libraries, packages, or external dependencies introduced
- new_directories: Any new directory structures created

This divergence data is consumed by /moai sync for SPEC document updates and project document synchronization.

#### Drift Guard Check (DDD)

After each DDD IMPROVE cycle completion, compare planned vs actual:

1. Read planned_files from `.moai/specs/SPEC-{ID}/tasks.md`
2. Compare against actual_files from divergence tracking above
3. Calculate drift: (unplanned_new_files / total_planned_files) * 100
4. Log to `.moai/specs/SPEC-{ID}/progress.md`:
   - Cycle number, planned count, actual count, drift percentage
   - List any unplanned files
5. Alert thresholds:
   - drift <= 20%: Informational only
   - 20% < drift <= 30%: Warning in progress.md
   - drift > 30% (cumulative): Trigger Phase 2.7 re-planning gate

#### Phase 2B: TDD Implementation (for tdd mode)

Agent: manager-tdd subagent

Input: Approved execution plan from Phase 1 plus task decomposition from Phase 1.5. Include `.moai/project/structure.md` and `.moai/project/tech.md` as onboarding context in the agent prompt so the implementation agent understands the project's architecture conventions before writing code.

Requirements:

- Initialize task tracking for progress across TDD cycles
- Execute the complete RED-GREEN-REFACTOR cycle for each feature
- Write tests before implementation (test-first discipline)
- Ensure minimum 80% coverage per commit (85% recommended for new code)

Output: files_created list, specification_tests_created list, test_results (all passing), coverage percentage, refactoring_improvements list, implementation_divergence report.

Implementation Divergence Tracking:

The manager-tdd subagent must track deviations from the original SPEC plan during implementation:

- planned_files: Files listed in plan.md that were expected to be created
- actual_files: Files actually created during the TDD cycle
- additional_features: Features or capabilities implemented beyond the original SPEC scope (with rationale)
- scope_changes: Description of any scope adjustments made during implementation
- new_dependencies: Any new libraries, packages, or external dependencies introduced
- new_directories: Any new directory structures created

This divergence data is consumed by /moai sync for SPEC document updates and project document synchronization.

#### Drift Guard Check (TDD)

After each TDD REFACTOR cycle completion, compare planned vs actual:

1. Read planned_files from `.moai/specs/SPEC-{ID}/tasks.md`
2. Compare against actual_files from divergence tracking above
3. Calculate drift: (unplanned_new_files / total_planned_files) * 100
4. Log to `.moai/specs/SPEC-{ID}/progress.md`:
   - Cycle number, planned count, actual count, drift percentage
   - List any unplanned files
5. Alert thresholds:
   - drift <= 20%: Informational only
   - 20% < drift <= 30%: Warning in progress.md
   - drift > 30% (cumulative): Trigger Phase 2.7 re-planning gate

### Phase 2.5: Quality Validation

Agent: manager-quality subagent

Input: Both Phase 1 planning context and Phase 2 implementation results.

TRUST 5 validation checks:

- Tested: Tests exist and pass before changes. Test-driven design discipline maintained.
- Readable: Code follows project conventions and includes documentation.
- Unified: Implementation follows existing project patterns.
- Secured: No security vulnerabilities introduced. OWASP compliance verified.
- Trackable: All changes logged with clear commit messages. History analysis supported.

Output: trust_5_validation results per pillar, coverage percentage, overall status (PASS, WARNING, or CRITICAL), and issues_found list.

#### Extended Quality Checks

Code Complexity Analysis:
- Function size: Flag functions exceeding 50 lines (suggest splitting)
- File size: Flag files exceeding 500 lines (suggest decomposition)
- Cyclomatic complexity: Flag functions with complexity > 10
- Nesting depth: Flag code with nesting > 4 levels

Dead Code Detection:
- Unused imports, functions, variables, and orphaned files
- Auto-removal: When confirmed, delegate to clean workflow (workflows/clean.md)

Side Effect Analysis:
- Caller impact: For each modified function, identify all callers and assess impact
- Interface changes: Flag signature changes that affect downstream consumers
- State mutations: Identify unexpected state changes in modified code paths
- Dependency chain: Trace changes through dependency graph to detect cascading effects

Code Reuse Opportunities:
- Duplication detection, library overlap, pattern consolidation, shared abstraction

### Quality Gate Decision

If status is CRITICAL:
- Present quality issues to user via AskUserQuestion
- Option to return to implementation phase for fixes
- Exit current execution flow

If coverage is below target (quality.yaml test_coverage_target):
- Auto-route to coverage workflow (workflows/coverage.md)
- Re-run quality validation after coverage improvement

If status is PASS or WARNING: Continue to Phase 2.8.

### Phase 2.7: Re-planning Gate Check

Purpose: Detect stagnation and trigger re-assessment if implementation is stuck. See .claude/rules/moai/workflow/spec-workflow.md for trigger conditions, communication path, and detection method.

Check `.moai/specs/SPEC-{ID}/progress.md` for stagnation signals. If triggered, return structured stagnation report to MoAI for user escalation.

### Phase 2.75: Pre-Review Quality Gate

Purpose: Run lightweight quality gate checks before the full review phase. This connects the gate workflow (workflows/gate.md) into the run pipeline.

Execution: Always runs. Equivalent to `/moai gate --fix` on modified files.

Steps:
1. Run language-specific lint on modified files
2. Run formatter check on modified files
3. Run type-checker on modified files
4. Auto-fix any fixable issues (--fix behavior)
5. If unfixable errors remain: Report and block (must fix before review)

Output: gate_report with pass/fail per check category. If all pass, continue to Phase 2.8a.

### Phase 2.8a: Active Quality Evaluation (evaluator-active)

**Condition**: Execute when harness level = standard or thorough (evaluator enabled).
**Skip**: When harness level = minimal.

Steps:
1. Invoke evaluator-active with:
   - SPEC acceptance criteria (from spec-compact.md or spec.md)
   - Sprint contract (from contract.md, if thorough harness)
   - Implementation changeset (modified/created files)
2. evaluator-active evaluates all 4 dimensions:
   - Functionality (40%): Run tests, verify each acceptance criterion
   - Security (25%): OWASP check (HARD: Security FAIL = overall FAIL)
   - Craft (20%): Coverage >= 85%, error handling review
   - Consistency (15%): Pattern adherence check
3. Verdict handling:
   - PASS: Proceed to Phase 2.8b
   - FAIL: Return specific findings to implementation agent for targeted fix
   - Maximum 3 fix-evaluate cycles
   - After 3 FAIL cycles: Present findings to user via AskUserQuestion

Mode-specific deployment:
- Sub-agent mode: Agent(subagent_type="evaluator-active")
- Team mode: SendMessage to reviewer teammate
- CG mode: Leader performs evaluation inline

Output: evaluation_report with per-dimension PASS/FAIL/UNVERIFIED verdicts and findings list.

<!-- moai:evolvable-start id="gate-run-2" -->
### HUMAN GATE: Implementation Complete

**Previous phase output:** Implementation with TRUST 5 validation passed
**Approval question:** Is the implementation ready for git operations?
**Cannot proceed until:**
- [ ] All tests pass (show evidence)
- [ ] TRUST 5 validation complete
- [ ] @MX tags updated if needed
- [ ] User has reviewed post-implementation issues list
<!-- moai:evolvable-end -->

### Phase 2.8b: TRUST 5 Static Verification (manager-quality) [MANDATORY]

Purpose: Multi-dimensional review iteration for high-quality output. This phase is ALWAYS executed to ensure consistent code quality.

**Standard review** (always executed via manager-quality subagent):
- Purpose alignment: Do changes match SPEC requirements?
- Improvement safety: Are existing behaviors preserved?
- Side effect verification: Any unintended impacts?
- Full change review: All modified files reviewed
- Dead code cleanup: No orphaned code left behind
- User flow validation: End-to-end correctness

**Security/Performance review** (conditional, triggered when changes affect security/performance/UX domains OR --review flag):
- Invoke review workflow explicitly: Read `${CLAUDE_SKILL_DIR}/workflows/review.md` and execute its multi-perspective analysis (security, performance, quality, UX reviewers)
- This replaces the previous vague "delegate to review workflow" with an explicit skill invocation

Iteration behavior:
- Each review dimension generates findings with severity (critical, warning, suggestion)
- Critical findings trigger a fix cycle: delegate to appropriate expert agent, then re-review
- Maximum 3 review iterations to prevent infinite loops
- If all dimensions pass with no critical findings: Continue to Phase 2.9

Output: review_findings per dimension, iterations_completed count, final review status.

### Phase 2.9: MX Tag Update [HARD]

Purpose: Update @MX code annotations for modified files. See .claude/rules/moai/workflow/mx-tag-protocol.md for tag rules.

[HARD] This phase is MANDATORY. MoAI MUST scan all files modified during Phase 2 and verify @MX tag coverage before proceeding to Phase 3. If implementation agents did not add required tags during their work, MoAI adds them here.

**Validation criteria (blocking):**
- P1: Every new exported function with fan_in >= 3 MUST have `@MX:ANCHOR`
- P2: Every new goroutine/async pattern MUST have `@MX:WARN`
- P1/P2 violations block Phase 3 until resolved

**TDD Mode:**
- Remove `@MX:TODO` tags for tests that now pass
- Add `@MX:NOTE` for complex logic added during GREEN phase
- Review `@MX:WARN` tags if dangerous patterns were improved

**DDD Mode:**
- Run 3-Pass scan if codebase has zero @MX tags
- Update `@MX:ANCHOR` tags if fan_in changed
- Add `@MX:NOTE` for business rules discovered during ANALYZE
- Convert `@MX:LEGACY` to `@MX:SPEC` if SPEC retroactively created

Output: MX_TAG_REPORT with tags added, updated, removed by type.

### LSP Quality Gates

The run phase enforces LSP-based quality gates as configured in quality.yaml:
- Zero LSP errors required (lsp_quality_gates.run.max_errors: 0)
- Zero type errors required (lsp_quality_gates.run.max_type_errors: 0)
- Zero lint errors required (lsp_quality_gates.run.max_lint_errors: 0)
- No regression from baseline allowed (lsp_quality_gates.run.allow_regression: false)

### Phase 3: Git Operations (Conditional)

Agent: manager-git subagent

Input: Full context from Phases 1, 2, and 2.5.

Execution conditions:
- quality_status is PASS or WARNING
- If config git_strategy.automation.auto_branch is true: Create feature branch feature/SPEC-{ID}
- If auto_branch is false: Commit directly to current branch

Tasks for manager-git:
- Create feature branch (if auto_branch enabled)
- Stage all relevant implementation and test files
- Create commits with conventional commit messages
- If SPEC metadata contains `issue_number` (non-zero): Include `Fixes #{issue_number}` in commit message footer
- Verify each commit was created successfully

Commit message format when issue_number exists:
```
feat(scope): description

SPEC: SPEC-{ID}
Fixes #{issue_number}

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>
```

Output: branch_name, commits array (sha and message), files_staged count, status, issue_number (from SPEC metadata).

### Phase 4: Completion and Guidance

Tool: AskUserQuestion (at orchestrator level)

Display implementation summary:
- Files created count, tests passing count, coverage percentage, commits count

Options:
- Sync Documentation (recommended): Execute /moai sync to synchronize docs and create PR
- Implement Another Feature: Return to /moai plan for additional SPEC
- Review Results: Examine implementation and test coverage locally
- Finish: Session complete

---

## Execution Mode Gate Integration

When the run phase is invoked from plan.md Decision Point 3.5 or moai.md Phase 11.5, the gate passes these parameters:
- `execution_mode`: worktree | team | sub-agent
- `active_mode`: cc | glm | cg
- `tmux_available`: true | false

**If execution_mode == "worktree":**
This run invocation is already inside the isolated tmux session and worktree.
Proceed with standard sub-agent run phase in the current environment.
No additional routing needed — CC/GLM/CG env is already configured by the Gate.

**If execution_mode == "team":**
Apply Team Mode Routing below. The active_mode determines worker model selection:
- CC: All teammates use Claude (default behavior)
- GLM: All teammates inherit GLM env from tmux session
- CG: Leader=Claude (clean session), Workers=GLM (tmux env injected)

**If execution_mode == "sub-agent":**
Skip Team Mode Routing. Proceed directly to Phase 1 (Strategy).

**If no execution_mode provided (direct `/moai run` invocation):**
Apply existing --team/--solo flag logic in Team Mode Routing below.

---

## Team Mode Routing

When --team flag is provided or auto-selected, the run phase MUST switch to team orchestration:

1. Verify prerequisites: workflow.team.enabled == true AND CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1 env var is set
2. If prerequisites met: Read ${CLAUDE_SKILL_DIR}/team/run.md and execute the team workflow (TeamCreate with backend-dev + frontend-dev + tester + quality)
3. If prerequisites NOT met: Warn user then fallback to standard sub-agent mode

Team composition: backend-dev (inherit) + frontend-dev (inherit) + tester (inherit) + quality (inherit, read-only)

### Worktree Isolation [HARD]

- [HARD] Implementation teammates (backend-dev, frontend-dev, tester) MUST use `isolation: "worktree"` when spawned via Agent()
- [HARD] Read-only teammates (quality) MUST NOT use `isolation: "worktree"` — permissionMode: plan is sufficient
- [HARD] All worktree path rules from "Worktree Path Rules [HARD] (All Modes)" section above apply to team mode as well
- After team shutdown, run `git worktree prune` to clean up stale worktree references

For detailed team orchestration steps, see ${CLAUDE_SKILL_DIR}/team/run.md.

---

## Context Propagation

Context flows forward through every phase:

- Phase 1 to Phase 2: Execution plan with architecture decisions guides implementation
- Phase 2 to Phase 2.5: Implementation code plus planning context enables context-aware validation
- Phase 2.5 to Phase 3: Quality findings enable semantically meaningful commit messages
- Phase 2 to /moai sync: Implementation divergence report enables accurate SPEC and project document updates

---

## Completion Criteria

All of the following must be verified:

- Phase 1: manager-strategy returned execution plan with requirements and success criteria
- User approval checkpoint blocked Phase 2 until user confirmed
- Phase 1.5: Tasks decomposed with requirement traceability
- Phase 1.8: MX context map built for target files (skipped for greenfield)
- Phase 2: Implementation completed according to development_mode (with MX context)
- Phase 2.5: manager-quality completed TRUST 5 validation with PASS or WARNING status
- Quality gate blocked Phase 3 if status was CRITICAL
- Phase 3: manager-git created commits (branch or direct) only if quality permitted
- Phase 4: User presented with next step options

---

## Test Scenarios

### Normal Flow
**Prompt**: "/moai run SPEC-AUTH-001"
**Expected Result**:
- Phase 0.9: Detects Go project (go.mod) → loads moai-lang-go
- Phase 0.95: SPEC has 8 files, 2 domains → Standard Mode selected
- Phase 1: manager-strategy creates execution plan with 5 tasks
- Decision Point: User approves plan
- Phase 2: Implementation via manager-ddd (DDD mode)
- Phase 2.5: TRUST 5 validation passes
- Phase 3: Commits created on feature branch

### Fix Mode Flow
**Prompt**: "/moai run SPEC-BUG-042" (bug fix SPEC, 2 files affected)
**Expected Result**:
- Phase 0.95: SPEC has 2 files, 1 domain → Fix Mode selected
- Directly spawns expert-debug + expert-testing
- Minimal overhead, fast execution
- Quality validation still runs

### Error Flow
**Prompt**: "/moai run SPEC-NONEXISTENT"
**Expected Result**:
- SPEC directory not found in .moai/specs/
- AskUserQuestion: "SPEC not found. Create it with /moai plan?"
- If user confirms, redirect to plan workflow

---

Version: 2.11.0
Updated: 2026-03-30
Changes: Added Phase 0.9 JIT Language Detection, Phase 0.95 Scale-Based Mode Selection, test scenarios.
