---
description: >
  Full autonomous plan-run-sync pipeline. Default workflow when no subcommand
  is specified. Handles parallel exploration, SPEC generation, DDD/TDD
  implementation with optional auto-fix loop, and documentation sync.
user-invocable: false
metadata:
  version: "3.0.0"
  category: "workflow"
  status: "active"
  updated: "2026-07-07"
  tags: "moai, autonomous, pipeline, plan-run-sync, default"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["moai", "autonomous", "pipeline", "build", "implement", "create"]
  agents: ["moai"]
  phases: ["plan", "run", "sync"]
---

# Workflow: MoAI - Autonomous Development Orchestration

Purpose: Full autonomous workflow. User provides a goal, MoAI autonomously executes plan -> run -> sync pipeline. This is the default workflow when no subcommand is specified.

Flow: Explore -> Plan -> Run -> Sync -> Done

For phase overview, token budgets, and phase transitions, see: .claude/rules/moai/workflow/spec-workflow.md

## Supported Flags

- --loop: Enable auto iterative fixing during run phase
- --max N: Maximum iteration count for loop (default 100)
- --branch: Auto-create feature branch
- --pr: Auto-create pull request after completion
- --resume SPEC-XXX: Resume previous work from existing SPEC
- --team: RETIRED — Agent Teams static layer retired; emits `MODE_TEAM_UNAVAILABLE` and falls back to sub-agent mode
- --solo: Force sub-agent mode (single agent per phase)
- --sequential: Run Phase 1 exploration agents sequentially instead of in parallel
- --issue: Opt-in GitHub Issue creation after SPEC generation (plan phase); absence skips Issue creation per the late-branch opt-in policy

**Default Behavior (no flag)**: The orchestrator auto-selects the execution mode from the Phase 4 6-mode catalog (`.claude/rules/moai/workflow/orchestration-mode-selection.md` §A — trivial / background / agent-team / parallel / sub-agent / workflow). The complexity auto-select thresholds are stated once in that rule's §B.1 (machine source: `workflow.yaml` `auto_selection`) — not restated here.

## Configuration Files

- quality.yaml: TRUST 5 quality thresholds AND development_mode routing
- workflow.yaml: Execution mode, team settings, loop prevention

## Development Mode Routing (CRITICAL)

[HARD] Before Phase 4 implementation, ALWAYS check `.moai/config/sections/quality.yaml`:

```yaml
constitution:
  development_mode: tdd    # or ddd
```

**Routing Logic**:

| Feature Type | Mode: ddd | Mode: tdd |
|--------------|-----------|-----------|
| **New package/module** (no existing file) | DDD* | TDD |
| **New feature in existing file** | DDD | TDD |
| **Refactoring existing code** | DDD | TDD (with brownfield pre-RED analysis) |
| **Bug fix in existing code** | DDD | TDD |

*DDD adapts for greenfield (ANALYZE requirements -> PRESERVE with spec tests -> IMPROVE)

**Agent Selection**:
- **TDD cycle**: `manager-develop` subagent (RED-GREEN-REFACTOR)
- **DDD cycle**: `manager-develop` subagent (ANALYZE-PRESERVE-IMPROVE)

For methodology details, see: .claude/rules/moai/workflow/spec-workflow.md (Run Phase section)

## Phase 1: Parallel Exploration

Launch three agents simultaneously in a single response (Priority High: parallel execution over sequential when the three agents are independent).

Agent 1 - Explore (subagent_type Explore, produces research.md):
- If .moai/project/codemaps/ exists: Use as architecture baseline to accelerate exploration (skip redundant scanning)
- Read target code areas IN DEPTH — understand deeply how each module works, its intricacies and side effects
- Study cross-module interactions IN GREAT DETAIL — trace data flow, identify implicit contracts
- Search for REFERENCE IMPLEMENTATIONS — find similar patterns in the codebase that can guide the new feature
- Document findings with specific file paths and line references
- Output: research.md artifact with architecture analysis, reference implementations, risks, and constraints

Agent 2 - Research (subagent_type Explore with WebSearch/WebFetch focus):
- External documentation and best practices
- API docs, library documentation, similar implementations
- Reference implementations from open-source projects that align with project conventions
- Documented design patterns relevant to the feature being implemented

Agent 3 - Quality (subagent_type sync-auditor — independent quality scoring per `.claude/rules/moai/workflow/archived-agent-rejection.md` §C row 2):
- Current project quality assessment
- Test coverage status, lint status, technical debt

After all agents complete:
- Collect outputs from each agent response
- Extract key findings from Explore (research.md with files, patterns, reference implementations), Research (external knowledge, documented patterns), Quality (coverage baseline)
- Synthesize into unified exploration report including research.md artifact
- Save research.md to .moai/specs/SPEC-{ID}/research.md when SPEC ID is determined
- Generate execution plan with files to create/modify and test strategy

Error handling: If any agent fails, continue with results from successful agents. Note missing information in plan.

If --sequential flag: Run Explore, then Research, then Quality sequentially instead.

## Phase 1 Completion: Routing Decision

Single-domain routing:
- If task is single-domain (e.g., "SQL optimization"): Delegate directly to expert agent, skip SPEC generation
- If task is multi-domain: Proceed to full workflow with SPEC generation

User approval checkpoint via AskUserQuestion:
- Options: Proceed to SPEC creation, Modify approach, Cancel

## Phase 2: SPEC Generation

- Delegate to manager-spec subagent
- Output: GEARS-format SPEC document at .moai/specs/SPEC-XXX/spec.md
- Includes requirements, acceptance criteria, technical approach

## Phase 3: Plan Annotation Cycle (1-6 iterations)

After SPEC generation and before implementation:
1. Present SPEC document and research.md to user for review
2. User adds inline annotations/corrections to plan
3. MoAI delegates to manager-spec: "Address all inline notes. DO NOT implement any code."
4. Repeat until user approves (maximum 6 iterations)
5. Track iteration count: "Annotation cycle {N}/6"

This iterative refinement catches architectural misunderstandings before implementation begins.

## Pipeline Gates (named, in order)

The default pipeline declares these gates explicitly. Each is implemented by its owning sub-skill (plan.md / run.md / sync.md); they are named here so the pipeline body carries the full gate sequence:

1. **Plan-audit gate (plan-auditor)** — after Phase 2/1.5: the plan-auditor subagent independently audits the SPEC plan artifacts in a fresh context (bias prevention). FAIL/INCONCLUSIVE halts the pipeline and surfaces to the user.
2. **Implementation Kickoff Approval (plan→run HUMAN GATE)** — presented exactly once per pipeline entry, at the plan→run boundary, via an orchestrator AskUserQuestion round. Score-independent: a plan-auditor PASS or skip-eligible score never bypasses it. A derived completion condition (router Step 2.8) does NOT authorize run-phase entry — only this gate does. All user preferences (tier, mode preference, PR strategy, chain scope) are drained at this gate.
3. **Phase 4 Mode Selection (6-mode catalog)** — autonomous selection per `orchestration-mode-selection.md` §A, logged to progress.md; strictly downstream of Implementation Kickoff Approval. **Mode 6 (workflow fan-out) operational entry**: selectable ONLY when the §C.3 capability gate holds — Implementation Kickoff Approval passed + all preferences collected + scope ≥ ~30 files with one uniform mechanical transform and no inter-file dependency + runtime ≥ v2.1.154 with workflows not disabled. Before launch, record the selection + gate confirmations in `progress.md` §F Phase 4 Mode Selection; then launch the fan-out from the orchestrator (scaling, not nesting) — workflow agents cannot prompt the user, so every needed decision must already be drained.
4. **Sync-audit gate (sync-auditor)** — after Phase 5: the sync-auditor subagent scores the sync output in a fresh context (4-dimension). FAIL/INCONCLUSIVE halts the chain — the pipeline never auto-completes past a failing gate. On FAIL, the sync-auditor verdict carries a structured defect-list (finding id / file+location / severity / required fix); the orchestrator routes fixes directly (orchestrator-direct edit or a single re-delegation) and the confirming re-audit is scoped to the enumerated defect delta rather than a from-scratch full re-audit — within the existing iteration ceilings. Verdict authority stays with the sync-auditor: the delta scope reduces re-audit cost, and it never substitutes an orchestrator self-assessment for an auditor verdict.

## Phase 4: Implementation (TDD or DDD based on development_mode)

[HARD] Agent delegation mandate: ALL implementation tasks MUST be delegated to specialized agents. NEVER execute implementation directly, even after auto compact.

[HARD] Methodology selection based on `.moai/config/sections/quality.yaml`:

- **development_mode: tdd** (default): Use `manager-develop` (RED-GREEN-REFACTOR)
- **development_mode: ddd**: Use `manager-develop` (ANALYZE-PRESERVE-IMPROVE)

Domain-specialist selection (for domain-specific work) — per `.claude/rules/moai/workflow/archived-agent-rejection.md` §C, domain expertise is injected at delegation time via a per-spawn `Agent(general-purpose)` with the domain whitelist + domain instructions, NOT a static expert agent file. Each spawn also carries 0-3 injected `moai-ref-*`/`moai-domain-*` skills per the delegation map (`.moai/config/sections/delegation.yaml` domain_skills; skill-routing.md §1):
- Backend logic: manager-develop (or per-spawn `Agent(general-purpose)` backend specialist) → Skill("moai-ref-api-patterns")
- Frontend components: manager-develop (or per-spawn `Agent(general-purpose)` frontend specialist) → Skill("moai-ref-react-patterns")
- Test creation: manager-develop subagent → Skill("moai-ref-testing-pyramid")
- Bug fixing: manager-develop + orchestrator verification batch (lint + test + coverage)
- Refactoring: manager-develop (cycle_type=ddd) or per-spawn `Agent(general-purpose)` refactoring specialist → Skill("moai-workflow-ddd")
- Security fixes: per-spawn `Agent(general-purpose)` security reviewer (or Stop hook dependency-manifest audit) → Skill("moai-ref-owasp-checklist")

Loop behavior (when --loop flag or workflow.yaml loop_prevention settings enabled) — this is the run-phase diagnostic fix-loop (Ralph-style, bounded by `loop_prevention.max_iterations`), DISTINCT from the pipeline-level § Agentic Completion Loop below (bounded by `agentic_loop.max_iterations`):
- While issues exist AND iteration less than max:
  - Execute diagnostics (parallel by default)
  - Delegate fix to appropriate expert agent
  - Verify fix results
  - Check whether completion conditions are satisfied
  - If satisfied: Break loop

## Phase 5: Documentation Sync

- Delegate to manager-docs subagent
- Synchronize documentation with implementation
- Detect SPEC-implementation divergence and update SPEC documents accordingly
- Conditionally update project documents (.moai/project/) when structural changes detected
- Respect SPEC lifecycle level for update strategy (spec-first, spec-anchored, spec-as-source)
- Signal completion in the Completion Report on success

## Agentic Completion Loop (post-kickoff autonomous iteration)

When the router recorded a completion condition (router Step 2.8) and the pipeline contract is `full-pipeline`, the orchestrator drives an autonomous completion loop AFTER Implementation Kickoff Approval. The loop operates ONLY downstream of that gate — entering the loop never substitutes for it.

**Lifecycle**:

- **Entry**: post-kickoff only. The completion condition is set via `/goal` when the runtime supports it (goal-directive transcript-measurable form); otherwise the orchestrator evaluates the identical condition text per-turn (graceful degradation — no parallel evaluator).
- **Iteration cycle**: run → sync → verify ONLY. While the condition is unmet, the loop re-enters the failing RUN or SYNC phase and iterates. Plan-phase re-entry is NEVER an autonomous loop step — it occurs solely via the no-progress escalation path below with explicit user approval, and the revised plan re-crosses the Implementation Kickoff Approval gate before any run-phase re-entry.
- **Termination** (any of four): (1) the completion condition evaluates met; (2) the iteration ceiling is reached — `workflow.agentic_loop.max_iterations` in `.moai/config/sections/workflow.yaml` (default 10; pipeline-level iterations, DISTINCT from `loop_prevention.max_iterations`, the per-operation diagnostic fix-loop bound); (3) an escalation fires (no-progress or semantic failure); (4) context-threshold suspension.

**Loop safety rules**:

- **Iteration-ceiling verdict** (cause 2): when the ceiling is reached, halt and emit the same structured 5-section evidence report (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk, per `verification-claim-integrity.md` §3) that the Ralph engine emits at its own ceiling exit (`workflows/loop.md` § Ceiling-Exit Verdict Contract) — persist remaining issues to `.moai/state/loop-verdict-<id>.json` and propose a lesson-capture entry before ending the session. This closes the protocol gap relative to causes 3 (no-progress escalation) and 4 (context-threshold suspension), which already carry their own structured reports below.
- **No-progress escalation**: when the same failure signature (identical failing check + same error class) is observed in two consecutive iterations, halt and escalate via a structured report; the orchestrator runs an AskUserQuestion round (continue with manual investigation / revert + re-plan / abort). Revert + re-plan re-crosses Implementation Kickoff Approval before any run-phase re-entry. No third identical iteration is attempted.
- **Dark-flow guard**: every iteration surfaces a per-iteration visible report in the conversation (iteration #, phase executed, evidence delta, condition-evaluation result). Silent iterations are prohibited — the transcript evidence is also what keeps the `/goal` evaluator functional (transcript-measurability per goal-directive).
- **Semantic-failure escalation**: on a semantic failure (data race, deadlock, panic, test assertion failure), clear the active completion condition and escalate immediately via AskUserQuestion — the loop never auto-fixes a semantic failure.
- **Context-threshold suspension**: when context usage crosses the model-specific handoff threshold (`context-window-management.md` § Context Window Targets), suspend at the current iteration boundary, persist state to progress.md, and emit the paste-ready resume message per `session-handoff.md`. The loop does not start a new iteration past the threshold.
- **Boundary**: subagents and workflow agents operating inside the loop never prompt the user; all mid-loop user decisions ride structured blocker reports → orchestrator AskUserQuestion (`agent-common-protocol.md` § User Interaction Boundary).
- **Relationship to the Ralph engine**: the Ralph engine (`/moai loop`) remains the specialized diagnostic fix-loop; this pipeline-level loop MAY invoke it during the verify step for mechanical convergence. The two differ in granularity — this loop iterates over phases (phase-granular); the Ralph engine iterates over diagnostics. DISTINCT again from both is the task-granular goal engine (`/moai goal`), which evaluates a condition-declared completion state each turn-end via the `stop-goal` Stop hook — see `.claude/skills/moai/workflows/goal.md`.

**run→sync chaining policy (pipeline contract)**:

- `full-pipeline` contract: run-phase completion auto-chains into sync, announced in the transcript — no additional approval round at the run→sync phase boundary (sync doc work is non-destructive; PR creation still follows Tier-based PR routing and its own gates). The HUMAN GATEs preserved INSIDE the sync workflow (`gate-sync-1` pre-sync quality, `gate-sync-2` documentation scope) still fire unchanged within the chained sync phase.
- `single-phase` contract (explicit `run`/`sync` invocation): phase completion surfaces the chain as the "(Recommended)" first option of the existing next-step AskUserQuestion — the chain never fires silently.
- Failing gates halt the chain: when the sync-audit gate returns FAIL/INCONCLUSIVE or the sync-phase quality gate blocks, the chain halts and escalates — the loop never auto-completes past a failing gate.

## Mode Selection (team dispatch retired)

The `--team` flag and Mode 3 (`agent-team`) are RETIRED with the Agent Teams
static layer. A forced `--team` emits `MODE_TEAM_UNAVAILABLE` and falls back to
sub-agent mode; the native `moai cg` GLM teammate runtime is unaffected.

Mode selection:
- `--team`: RETIRED — emits `MODE_TEAM_UNAVAILABLE`, falls back to Mode 5 (sub-agent).
- `--solo`: Force Mode 5 (sub-agent).
- No flag (default): Auto-select per the Phase 4 6-mode catalog; thresholds stated once in `orchestration-mode-selection.md` §B.1.

## Execution Summary

1. Parse arguments (extract flags: --loop, --max, --sequential, --branch, --pr, --resume, --team, --solo, --issue)
2. If --resume with SPEC ID: Load existing SPEC and continue from last state
3. Detect development_mode from quality.yaml (ddd/tdd)
4. **Mode decision**: determine execution mode
   - If `--team` flag: RETIRED — emit `MODE_TEAM_UNAVAILABLE` and fall back to sub-agent mode (Agent Teams static layer retired)
   - If `--solo` flag: Force sub-agent mode
   - If no flag (default): Auto-select per the Phase 4 6-mode catalog (thresholds per `orchestration-mode-selection.md` §B.1)
5. Execute Phase 1 (parallel or sequential exploration)
6. Routing decision (single-domain direct delegation vs full workflow)
7. TaskCreate for discovered tasks
8. User confirmation via AskUserQuestion
9. **Phase 1 (Research)**: Save research.md from Phase 1 Explore findings to SPEC directory
10. **Phase 2 (Plan)**: manager-spec sub-agent (team orchestration retired)
10.5. **Phase 2 (Issue)**: Create GitHub Issue linked to SPEC (only when the `--issue` opt-in flag is set; default skips per the late-branch opt-in policy). See plan.md Phase 13.
11. **Phase 3 (Annotate)**: Run annotation cycle (1-6 iterations) until user approves plan
11.2. **Plan-audit gate**: plan-auditor independent audit of the plan artifacts (Pipeline Gates #1); FAIL/INCONCLUSIVE halts
11.3. **Implementation Kickoff Approval**: plan→run HUMAN GATE — exactly once per pipeline entry, score-independent (Pipeline Gates #2). Merged round: presented in a single AskUserQuestion call carrying both this Kickoff question AND the Step 11.5 execution-shape question (multi-question, ≤4 questions per call). The Kickoff question offers run-phase entry (Recommended) / additional review / abort; merging co-locates the two questions into one blocking round-trip and never removes, weakens, or auto-bypasses the Kickoff gate — declining Kickoff halts run-phase entry exactly as a standalone round would
11.5. **Execution Mode Selection Gate**: co-located with Step 11.3 in the same single AskUserQuestion call (see 11.3) — shape preferences collected here feed Phase 4 mode selection (6-mode catalog, Pipeline Gates #3)
   - If `--team` flag: RETIRED — emit `MODE_TEAM_UNAVAILABLE`, fall back to execution_mode="sub-agent"
   - If `--solo` flag: Skip the execution-shape question (auto-select execution_mode="sub-agent"); the Kickoff question still rides its own round
   - Otherwise (no flag):
     - Read .moai/config/sections/llm.yaml → team_mode ("" = cc, "glm" = glm, "cg" = cg)
     - Bash: test -n "$TMUX" && echo "tmux" || echo "no-tmux"
     - Merged AskUserQuestion (single call, with Step 11.3): Q1 Kickoff — run-phase entry (Recommended) / additional review / abort; Q2 execution shape — worktree+{mode} (Recommended if tmux available) | sub-agent
   - Worktree selected: Launch new tmux session in worktree dir, terminate current pipeline
   - Sub-agent selected: Pass execution_mode + active_mode to Phase 2
   - See plan.md Decision Point 3.5 for full option details
12. **Phase 3 (Harness Level Auto-Detection)**: Determine pipeline depth before Run
   - Load `.moai/config/sections/harness.yaml` (if not found, default to standard)
   - CG mode: Always thorough (natural Generator-Evaluator split)
   - Solo/Team: Run Complexity Estimator:
     - Count distinct domains in SPEC requirements (domain_count)
     - Count total files to modify (file_count, from plan.md)
     - Check for security/payment/critical keywords
     - Compute complexity_score = domain_count * 2 + file_count / 3 (integer, rounded down)
   - Apply auto_detection rules (evaluated in order, first match wins):
     - security/payment keywords OR spec_priority == critical → thorough
     - file_count >= 10 AND multi_domain (domain_count >= 2) → thorough
     - file_count > 3 OR multi_domain → standard
     - file_count <= 3 AND single_domain AND no security keywords → minimal
   - Record detected harness level in progress.md
   - Pass harness level to Run phase
13. **Phase 4 (Run)**: Route based on Gate result (execution_mode parameter)
   - worktree: Already running in isolated tmux+worktree session (Gate handled transition)
   - sub-agent: manager-develop (cycle_type=ddd or tdd, per quality.yaml development_mode)
   - Harness level determines phase skipping and evaluator involvement
14. **Phase 5 (Sync)**: Always manager-docs sub-agent (sync phase is always sub-agent) — entered via auto-chain on a `full-pipeline` contract, or via the "(Recommended)" next-step option on a `single-phase` contract
14.5. **Sync-audit gate**: sync-auditor independent 4-dimension scoring (Pipeline Gates #4); FAIL/INCONCLUSIVE halts the chain
15. Terminate with the Completion Report completion signal (or continue the Agentic Completion Loop while the completion condition is unmet)
   - Full-pipeline completion close: when a `full-pipeline` contract completes successfully with no genuine pending decision, close with a clean completion statement and NO manufactured next-step question — the askuser-protocol § Completion-Report Next-Step Discipline "close with NO question" clause is the full-pipeline default. A genuine next-step decision, when one actually exists, still rides AskUserQuestion.
   - `single-phase` contract completions keep the existing "(Recommended)" next-step chain question unchanged (Step 14 — the chain never fires silently)

---

Version: 3.0.1
Updated: 2026-07-09
Source: SPEC-MOAI-001. Named pipeline gates + agentic completion loop + chaining policy (v3.0.0). Added the iteration-ceiling verdict protocol for Agentic Completion Loop termination cause 2, closing its parity gap with causes 3/4 (v3.0.1). Previous: --team/--solo flag Gate auto-skip (v2.9.0), Harness auto-detection (v2.8.0).
