---
description: "Run Mode Routing — Execution mode gate integration, team mode routing, context propagation, completion criteria, test scenarios, and custom harness extension"
user-invocable: false
metadata:
  parent: moai-workflow-run
  phase: "Mode Routing: Execution Mode Gate, Team Mode, Completion, and Scenarios"
---

# Execution Mode Gate Integration

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

# Team Mode Routing

When --team flag is provided or auto-selected — Mode 3 (`agent-team`) of the Phase 0.95 catalog (`.claude/rules/moai/workflow/orchestration-mode-selection.md` §A), resolved through the §C.1 capability gate on both paths — the run phase MUST switch to team orchestration:

1. Verify prerequisites: workflow.team.enabled == true AND CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1 env var is set
2. If prerequisites met: Read ${CLAUDE_SKILL_DIR}/team/run.md and execute the team workflow (spawn backend-dev + frontend-dev + tester + quality via Agent(name=...) — the team forms implicitly on first spawn)
3. If prerequisites NOT met: Warn user then fallback to standard sub-agent mode

Team composition: backend-dev (inherit) + frontend-dev (inherit) + tester (inherit) + quality (inherit, read-only)

## Worktree Isolation [HARD]

- [SHOULD] When spawning implementation teammates (backend-dev, frontend-dev, tester) via `Agent(isolation: "worktree")`, Claude Code runtime decides whether to materialize an L1 worktree. MoAI orchestrator does NOT mandate isolation (per the worktree-autonomous user policy).
- [SHOULD] Read-only teammates (quality) typically do not benefit from `isolation: "worktree"`; omit the flag unless a specific reason applies. `permissionMode: plan` is sufficient.
- [HARD] All worktree path rules from context-loading.md "Worktree Path Rules [HARD] (All Modes)" section apply to team mode as well
- After team shutdown, run `git worktree prune` to clean up stale worktree references

For detailed team orchestration steps, see ${CLAUDE_SKILL_DIR}/team/run.md.

---

# Context Propagation

Context flows forward through every phase:

- Phase 1 to Phase 2: Execution plan with architecture decisions guides implementation
- Phase 2 to Phase 2.5: Implementation code plus planning context enables context-aware validation
- Phase 2.5 to Phase 3: Quality findings enable semantically meaningful commit messages
- Phase 2 to /moai sync: Implementation divergence report enables accurate SPEC and project document updates

---

# Completion Criteria

All of the following must be verified:

- Phase 1: manager-spec returned execution plan with requirements and success criteria
- User approval checkpoint blocked Phase 2 until user confirmed
- Phase 1.5: Tasks decomposed with requirement traceability
- Phase 1.8: MX context map built for target files (skipped for greenfield)
- Phase 2: Implementation completed according to development_mode (with MX context)
- Phase 2.5: sync-auditor (or orchestrator verification batch) completed TRUST 5 validation with PASS or WARNING status
- Quality gate blocked Phase 3 if status was CRITICAL
- Phase 3: manager-git created commits (branch or direct) only if quality permitted
- Phase 4: Next step honors the pipeline contract — `full-pipeline` auto-chains into `/moai sync` (announced in the transcript); `single-phase` presents sync as the "(Recommended)" first next-step option (never a silent chain)

---

# Test Scenarios

## Normal Flow
**Prompt**: "/moai run SPEC-AUTH-001"
**Expected Result**:
- Phase 0.9: Detects Go project (go.mod) → references `.claude/rules/moai/languages/go.md`
- Phase 0.95: SPEC has 8 files, 2 domains → Standard Mode selected
- Phase 1: manager-spec creates execution plan with 5 tasks
- Decision Point: User approves plan
- Phase 2: Implementation via manager-develop (DDD mode)
- Phase 2.5: TRUST 5 validation passes
- Phase 3: Commits created on feature branch

## Fix Mode Flow
**Prompt**: "/moai run SPEC-BUG-042" (bug fix SPEC, 2 files affected)
**Expected Result**:
- Phase 0.95: SPEC has 2 files, 1 domain → Fix Mode selected
- Directly spawns manager-develop + orchestrator verification batch (lint + test + coverage)
- Minimal overhead, fast execution
- Quality validation still runs

## Error Flow
**Prompt**: "/moai run SPEC-NONEXISTENT"
**Expected Result**:
- SPEC directory not found in .moai/specs/
- AskUserQuestion: "SPEC not found. Create it with /moai plan?"
- If user confirms, redirect to plan workflow

---

Version: 2.11.0
Updated: 2026-03-30
Changes: Added Phase 0.9 JIT Language Detection, Phase 0.95 Scale-Based Mode Selection, test scenarios.

---

# Custom Harness Extension (Optional)

@.moai/harness/run-extension.md

*(이 파일은 `/moai project --harness`로 생성됩니다. 파일이 없으면 자동으로 skip됩니다.)*
