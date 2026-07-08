---
description: "Run Phase 2~4 — DDD/TDD implementation cycles, quality validation, git operations, and completion guidance"
user-invocable: false
metadata:
  parent: moai-workflow-run
  phase: "Phase 2-4: Implementation, Quality Validation, and Completion"
---

# Phase 2: Implementation (Mode-Dependent)

**[HARD] Worktree Prompt Construction**: When spawning implementation agents (manager-develop) with `isolation: "worktree"`, the orchestrator MUST construct prompts using project-root-relative paths only. Do NOT embed the current working directory path in the agent prompt. See context-loading.md "Worktree Path Rules [HARD]" section.

## Phase 2A: DDD Implementation (for ddd mode)

Agent: manager-develop subagent

Input: Approved execution plan from Phase 1 plus task decomposition from Phase 1.5. Include `.moai/project/structure.md` and `.moai/project/tech.md` as onboarding context in the agent prompt so the implementation agent understands the project's architecture conventions before writing code.

Requirements:

- Initialize task tracking for progress across refactoring steps
- Execute the complete ANALYZE-PRESERVE-IMPROVE cycle
- Verify all existing tests pass after each transformation
- Create characterization tests for uncovered code paths
- Ensure test coverage meets or exceeds 85%

Output: files_modified list, characterization_tests_created list, test_results (all passing), behavior_preserved flag, structural_metrics comparison, implementation_divergence report.

Implementation Divergence Tracking:

The manager-develop subagent must track deviations from the original SPEC plan during implementation:

- planned_files: Files listed in plan.md that were expected to be created or modified
- actual_files: Files actually created or modified during the DDD cycle
- additional_features: Features or capabilities implemented beyond the original SPEC scope (with rationale)
- scope_changes: Description of any scope adjustments made during implementation (expansions, deferrals, or substitutions)
- new_dependencies: Any new libraries, packages, or external dependencies introduced
- new_directories: Any new directory structures created

This divergence data is consumed by /moai sync for SPEC document updates and project document synchronization.

### Drift Guard Check (DDD)

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

## Phase 2B: TDD Implementation (for tdd mode)

Agent: manager-develop subagent

Input: Approved execution plan from Phase 1 plus task decomposition from Phase 1.5. Include `.moai/project/structure.md` and `.moai/project/tech.md` as onboarding context in the agent prompt so the implementation agent understands the project's architecture conventions before writing code.

Requirements:

- Initialize task tracking for progress across TDD cycles
- Execute the complete RED-GREEN-REFACTOR cycle for each feature
- Write tests before implementation (test-first discipline)
- Ensure minimum 80% coverage per commit (85% recommended for new code)

Output: files_created list, specification_tests_created list, test_results (all passing), coverage percentage, refactoring_improvements list, implementation_divergence report.

Implementation Divergence Tracking:

The manager-develop subagent must track deviations from the original SPEC plan during implementation:

- planned_files: Files listed in plan.md that were expected to be created
- actual_files: Files actually created during the TDD cycle
- additional_features: Features or capabilities implemented beyond the original SPEC scope (with rationale)
- scope_changes: Description of any scope adjustments made during implementation
- new_dependencies: Any new libraries, packages, or external dependencies introduced
- new_directories: Any new directory structures created

This divergence data is consumed by /moai sync for SPEC document updates and project document synchronization.

### Drift Guard Check (TDD)

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

## Phase 2.5: Quality Validation

Agent: sync-auditor subagent (independent quality scoring per `.claude/rules/moai/workflow/archived-agent-rejection.md` §C row 2; OR orchestrator verification batch — lint + test + coverage)

Input: Both Phase 1 planning context and Phase 2 implementation results.

TRUST 5 validation checks:

- Tested: Tests exist and pass before changes. Test-driven design discipline maintained.
- Readable: Code follows project conventions and includes documentation.
- Unified: Implementation follows existing project patterns.
- Secured: No security vulnerabilities introduced. OWASP compliance verified.
- Trackable: All changes logged with clear commit messages. History analysis supported.

Output: trust_5_validation results per pillar, coverage percentage, overall status (PASS, WARNING, or CRITICAL), and issues_found list.

### Extended Quality Checks

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

## Quality Gate Decision

If status is CRITICAL:
- Present quality issues to user via AskUserQuestion
- Option to return to implementation phase for fixes
- Exit current execution flow

If coverage is below target (quality.yaml test_coverage_target):
- Route coverage-gap handling through `go test -cover` + `/moai gate` (the documented coverage replacement path)
- Re-run quality validation after coverage improvement

If status is PASS or WARNING: Continue to Phase 2.8.

## Phase 2.7: Re-planning Gate Check

Purpose: Detect stagnation and trigger re-assessment if implementation is stuck. See .claude/rules/moai/workflow/spec-workflow.md for trigger conditions, communication path, and detection method.

Check `.moai/specs/SPEC-{ID}/progress.md` for stagnation signals. If triggered, return structured stagnation report to MoAI for user escalation.

## Phase 2.75: Pre-Review Quality Gate

Purpose: Run lightweight quality gate checks before the full review phase. This connects the gate workflow (workflows/gate.md) into the run pipeline.

Execution: Always runs. Equivalent to `/moai gate --fix` on modified files.

Steps: Run language-specific lint, formatter check, and type-checker as a single-turn multi-Bash parallel batch (per `.claude/rules/moai/core/agent-common-protocol.md` § Parallel Execution) — one Bash tool call per check within the same assistant turn, each redirecting verbatim output to `/tmp/moai-verify/` and surfacing only exit code + bounded-tail summary in context (the file-redirect contract). Auto-fix any fixable issues (`--fix` behavior) after the batch completes. If unfixable errors remain: Report and block (must fix before review).

Output: gate_report with pass/fail per check category. If all pass, continue to Phase 2.8a.

## Phase 2.8a: Active Quality Evaluation (sync-auditor)

**Condition**: Execute when harness level = standard or thorough (evaluator enabled).
**Skip**: When harness level = minimal.

Steps:
1. Invoke sync-auditor with:
   - SPEC acceptance criteria (from spec-compact.md or spec.md)
   - Milestone contract (from contract.md, if thorough harness)
   - Implementation changeset (modified/created files)
2. sync-auditor evaluates all 4 dimensions:
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
- Sub-agent mode: Agent(subagent_type="sync-auditor")
- Team mode: SendMessage to reviewer teammate
- CG mode: Leader performs evaluation inline

Output: evaluation_report with per-dimension PASS/FAIL/UNVERIFIED verdicts and findings list.

<!-- moai:evolvable-start id="gate-run-2" -->
## HUMAN GATE: Implementation Complete

**Previous phase output:** Implementation with TRUST 5 validation passed
**Approval question:** Is the implementation ready for git operations?
**Cannot proceed until:**
- [ ] All tests pass (show evidence)
- [ ] TRUST 5 validation complete
- [ ] @MX tags updated if needed
- [ ] User has reviewed post-implementation issues list
<!-- moai:evolvable-end -->

## Phase 2.8b: TRUST 5 Static Verification (sync-auditor) [MANDATORY]

Purpose: Multi-dimensional review iteration for high-quality output. This phase is ALWAYS executed to ensure consistent code quality.

**Standard review** (always executed via the sync-auditor subagent — independent quality scoring per `.claude/rules/moai/workflow/archived-agent-rejection.md` §C row 2):
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

## Phase 2.9: MX Tag Update [HARD]

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

## LSP Quality Gates

The run phase enforces LSP-based quality gates as configured in quality.yaml:
- Zero LSP errors required (lsp_quality_gates.run.max_errors: 0)
- Zero type errors required (lsp_quality_gates.run.max_type_errors: 0)
- Zero lint errors required (lsp_quality_gates.run.max_lint_errors: 0)
- No regression from baseline allowed (lsp_quality_gates.run.allow_regression: false)

## Phase 3: Git Operations (Conditional)

Input: Full context from Phases 1, 2, and 2.5.

Execution conditions:
- quality_status is PASS or WARNING

Route selection (per `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Phase Discipline — Frozen HARD Route A/B): **Route A** (Hybrid Trunk main-direct) applies to Tier S/M SPECs with no explicit `--pr` flag; **Route B** (PR route) applies to Tier L SPECs OR any SPEC with an explicit `--pr` flag.

**Route A — Tier S/M, no `--pr` (default):**

Agent: manager-develop subagent (the same agent that performed Phase 2 — no separate agent spawn for git operations)

Tasks:
- Stage all relevant implementation and test files
- Commit directly to `main` with conventional commit messages (no feature branch, no PR)
- If SPEC metadata contains `issue_number` (non-zero): Include `Fixes #{issue_number}` in commit message footer
- Push the commit(s) to `main`
- Verify each commit was created and pushed successfully

**Route B — Tier L OR explicit `--pr`:**

Agent: manager-git subagent

Tasks:
- Create feature branch `feat/SPEC-{ID}`
- Stage all relevant implementation and test files
- Create commits with conventional commit messages
- If SPEC metadata contains `issue_number` (non-zero): Include `Fixes #{issue_number}` in commit message footer
- Open a PR per the configured `merge_method` (default `squash`)
- Verify each commit was created successfully

Commit message format when issue_number exists (both routes):
```
feat(scope): description

SPEC: SPEC-{ID}
Fixes #{issue_number}

Co-Authored-By: Claude <noreply@anthropic.com>
```

Output: branch_name (Route B only; n/a on Route A), commits array (sha and message), files_staged count, status, issue_number (from SPEC metadata).

## Phase 4: Completion and Guidance

Tool: AskUserQuestion (at orchestrator level)

Display implementation summary:
- Files created count, tests passing count, coverage percentage, commits count

Options:
- Sync Documentation (recommended): Execute /moai sync to synchronize docs and create PR
- Implement Another Feature: Return to /moai plan for additional SPEC
- Review Results: Examine implementation and test coverage locally
- Finish: Session complete
