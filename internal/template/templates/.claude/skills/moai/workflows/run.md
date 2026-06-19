---
description: >
  DDD/TDD implementation workflow for SPEC requirements. Second step
  of the Plan-Run-Sync workflow. Routes to manager-develop based
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
  agents: ["manager-develop", "manager-git", "Explore"]
  phases: ["run"]
---

<!-- TRACE PROBE: per SPEC-V3R4-WORKFLOW-SPLIT-001 T0.5 baseline trace mechanism -->
<!-- Activated by MOAI_TRACE_PHASES=1 environment variable -->
<!-- Emits one line per Phase entry/exit to stderr in format: [trace] /moai run Phase <N> <enter|exit> -->

# Run Workflow Entry Router

이 파일은 `/moai run` 진입점 라우터입니다. 각 Phase는 on-demand로 해당 sub-skill을 `Read`하여 로드합니다.

## Phase Owners (per the canonical agent catalog policy)

Phase Owners: `manager-develop` (run-phase implementation — single-spawn per Anthropic's coding-task parallelism caveat "most coding tasks involve fewer truly parallelizable tasks than research"; `cycle_type` ∈ `{tdd, ddd, autofix}` per the canonical cycle-type contract) + `manager-git` (Tier L PR creation OR `--pr` flag per the canonical Tier-based PR routing policy) + `Explore` (read-only investigation when scope discovery needed).

Phase 0.95 Mode Selection: orchestrator autonomous 5-mode decision (autopilot / loop / team / pipeline / background) is logged at `.moai/specs/SPEC-{ID}/progress.md` § Phase 0.95 Mode Selection. Phase 0.95 SHOULD be invoked before any manager-develop spawn for SPECs sized ≥ Tier M.

`cycle_type=autofix` mode: `/moai fix` workflow integration delegates to manager-develop with the utility-class pipeline 3-phase contract (localize → repair → validate per `.claude/rules/moai/workflow/spec-workflow.md` § Subcommand Classification) and the max-3-iteration contract per `.claude/rules/moai/workflow/ci-autofix-protocol.md`.

## Phase Routing Table

| Phase Group | Sub-skill 경로 | 내용 |
|------------|----------------|------|
| Phase 0: Context Loading | `Read workflows/run/context-loading.md` | Mode dispatch, UltraThink, harness level, context loading, worktree path rules |
| Phase 0.5~1.8: Phase Execution | `Read workflows/run/phase-execution.md` | Plan Audit Gate, environment assessment, JIT language detection, scale-based mode, analysis/planning, task decomposition, development mode routing |
| Phase 2~4: Implementation | `Read workflows/run/task-decomposition.md` | DDD/TDD cycles, quality validation (Phase 2.5/2.8), git operations (Phase 3), completion guidance (Phase 4) |
| Mode Routing + Completion | `Read workflows/run/mode-orchestration.md` | Execution mode gate, team mode routing, context propagation, completion criteria, test scenarios |

## Invocation Flow

```
/moai run SPEC-XXX
  ├── [trace] /moai run Phase 0 enter
  │   Read workflows/run/context-loading.md  → Mode dispatch + context setup
  ├── [trace] /moai run Phase 0.5 enter
  │   Read workflows/run/phase-execution.md  → Phase Sequence (0.5~1.8) + mode routing
  ├── [trace] /moai run Phase 2 enter
  │   Read workflows/run/task-decomposition.md → Implementation + quality + git
  └── [trace] /moai run Mode enter
      Read workflows/run/mode-orchestration.md → Team mode + completion criteria
```

## Quick Reference

**Purpose**: SPEC 요구사항을 DDD 또는 TDD 방법론으로 구현합니다.

**Input**: `$ARGUMENTS` = SPEC-ID (예: `SPEC-AUTH-001`)

**Development mode**: `.moai/config/sections/quality.yaml` `development_mode` 설정 (`ddd` 또는 `tdd`)에 따라 자동 선택.

**Mode dispatch** (`--mode` flag):
- `autopilot` (기본): Phase 0.95 scale-based 선택 후 Phase 2A/2B 실행
- `loop`: Ralph engine 위임 (see `loop.md`)
- `team`: Agent Teams 모드 (requires prerequisites)
- `pipeline`: REJECTED — `MODE_PIPELINE_ONLY_UTILITY` 오류 반환

**Harness levels**: `minimal` → skip optional phases | `standard` → all phases | `thorough` → sprint contract + sync-auditor

**Phase 0.5 (Plan Audit Gate)**: 모든 harness level에서 SKIP 불가. SPEC plan 아티팩트 독립 감사 필수.

**Worktree path rules**: [HARD] 모든 에이전트 프롬프트에 절대 경로 금지. project-root-relative 경로 사용.

## On-Demand Sub-skill Loading

각 Phase 진입 시점에 해당 sub-skill을 로드합니다:

```
# Phase 0: Context loading 및 mode dispatch 시작 시
Read .claude/skills/moai/workflows/run/context-loading.md

# Phase 0.5 (Plan Audit Gate) 진입 시
Read .claude/skills/moai/workflows/run/phase-execution.md

# Phase 2 (Implementation) 진입 시
Read .claude/skills/moai/workflows/run/task-decomposition.md

# Team mode 또는 completion criteria 확인 시
Read .claude/skills/moai/workflows/run/mode-orchestration.md
```

## Custom Harness Extension

@.moai/harness/run-extension.md

*(이 파일은 `/moai project --harness`로 생성됩니다. 파일이 없으면 자동으로 skip됩니다.)*

## Sentinel Error Keys

CI guards in `internal/template/agentless_audit_test.go` enforce the literal `MODE_UNKNOWN` sentinel remains present in this skill body (REQ-WF003-010, shared with `design.md`). `MODE_UNKNOWN` is emitted when `--mode <value>` is supplied to `/moai run` but `<value>` is not in the valid set `{autopilot, loop, team, pipeline}` (note: pipeline is itself rejected with the separate `MODE_PIPELINE_ONLY_UTILITY` sentinel — see line 71). The complementary `MODE_PIPELINE_ONLY_UTILITY` and `MODE_TEAM_UNAVAILABLE` sentinels are documented in this skill body and in `design.md`.

Ordering invariant (read before the autonomy section below): the Implementation Kickoff Approval `AskUserQuestion` human gate is always cleared FIRST; any run-phase autonomy set is downstream of it. The next section documents that ordering and the autonomy condition together.

## Run-phase Autonomy (/goal ac_converge)

This section wires the run-phase autonomy mechanisms — the Implementation Kickoff Approval human-gate ordering reference and the `ac_converge` `/goal` condition — into a single co-located place. The two parts are ORDERED: the Implementation Kickoff Approval `AskUserQuestion` human gate is described FIRST (it must be cleared before any autonomy begins), then the `/goal ac_converge` set (entered only after Implementation Kickoff Approval approval).

### 1. Implementation Kickoff Approval ordering (the human gate comes first)

[HARD] Before any run-phase autonomy (a `/goal` set, a Mode 6 Workflow launch, or any autonomous loop), the orchestrator MUST have already obtained explicit Implementation Kickoff Approval approval. Implementation Kickoff Approval is the plan→run HUMAN GATE: a mandatory orchestrator-issued `AskUserQuestion` round (run-phase entry / further review / abort, first option marked "(Recommended)") presented after Phase 0.5 (Plan Audit Gate) and before Phase 0.95 (Mode Selection). The orchestrator emits this gate; it is never embedded inside a subagent body (subagents cannot prompt the user — the asymmetric boundary in `.claude/rules/moai/core/agent-common-protocol.md` § User Interaction Boundary).

[HARD] Implementation Kickoff Approval is **score-independent**: the orchestrator emits the Implementation Kickoff Approval `AskUserQuestion` gate **regardless of the plan-auditor score**, including the high skip-eligible case. Skip-eligibility (a high autonomous-bypass score) applies ONLY to Phase 0.5 plan-auditor verdict re-execution — NOT to Implementation Kickoff Approval. A high plan-auditor score never authorizes skipping the Implementation Kickoff Approval human gate. This is the Implementation Kickoff Approval mandatory-restoration invariant per the Implementation Kickoff Approval mandatory-restoration policy.

Because Implementation Kickoff Approval also drains all user preferences (Tier, mode preference, PR strategy), the orchestrator collects every preference at this gate BEFORE launching any autonomy — `/goal`-turn agents and Mode 6 Workflow agents cannot prompt the user mid-run, so the one decision that must involve the user is taken here.

### 2. The `ac_converge` `/goal` condition (set only after Implementation Kickoff Approval approval)

ONLY after Implementation Kickoff Approval approval is obtained, the orchestrator MAY set the `ac_converge` `/goal` to grant phase-internal autonomy (it removes per-turn STOP prompts so the run-phase loop continues until convergence). The condition is hard-coded inline (no registry dependency) and is transcript-measurable — every predicate references a line the orchestrator surfaces in the conversation, never a file the `/goal` evaluator would have to read:

```text
Every blocking acceptance criterion in
.moai/specs/SPEC-{ID}/acceptance.md has its PASS evidence surfaced in
the conversation (test output, build exit 0, or explicit AC-id: PASS
line); AND `go test ./...` exit 0 is surfaced; AND no test file outside
the SPEC scope was modified (surfaced via git status). Stop when all
hold. Max 20 turns.
On any semantic failure (data race, deadlock, panic, test assertion
failure), clear this goal and escalate via AskUserQuestion — do NOT
auto-fix semantic failures.
[PRECONDITION: Implementation Kickoff Approval user approval already obtained; this goal does
 NOT substitute for or bypass Implementation Kickoff Approval.]
```

### 3. Transcript-measurability (the evaluator judges only surfaced lines)

The reference to `acceptance.md` in the condition is the NAMING of where the AC list lives — it is NOT an instruction for the evaluator to open that path. The measured predicate is "PASS evidence **surfaced in the conversation**". The Haiku evaluator judges only what the orchestrator has surfaced into the transcript; the orchestrator is responsible for surfacing each per-AC PASS line, the `go test ./...` exit code (exit 0), and the `git status` output. No predicate is a file-path read.

### 4. Semantic-failure escalation (HARD)

When a semantic failure — a data race, a deadlock, a panic, or a test assertion failure — is surfaced during the autonomous loop, the orchestrator MUST clear the `/goal` and escalate via `AskUserQuestion` rather than auto-fixing the semantic failure (aligns with the semantic-failure-handling rule). Semantic failures require human approval; the autonomous loop is for convergence on mechanical / test PASS evidence, not for fixing race conditions silently.

### 5. Non-substitution clause (HARD)

While the `ac_converge` goal is active, the orchestrator MUST NOT treat the goal as authorization to bypass Implementation Kickoff Approval, create a PR, or perform any destructive operation. Those remain explicit, separately-surfaced gates: Implementation Kickoff Approval was already cleared before the goal was set; PR creation and any destructive operation are surfaced for explicit decision after convergence. The `/goal` removes per-turn STOP prompts — it does not pre-approve hard-to-reverse or shared-system actions, and the goal never substitutes for the human gate.

### 6. Blocker reports, never user prompts (asymmetric boundary)

A `/goal`-turn agent or a Mode 6 Workflow agent that lacks a required input returns a structured blocker report; the orchestrator runs an `AskUserQuestion` round and re-delegates with the answers injected. Agents never prompt the user directly. This is the asymmetric orchestrator-subagent boundary per `.claude/rules/moai/core/agent-common-protocol.md` § User Interaction Boundary.

### 7. Graceful degradation when `/goal` is unavailable

When the runtime does not support `/goal` (version below v2.1.139, or hooks disabled via `disableAllHooks` / `allowManagedHooksOnly`), run-phase autonomy degrades gracefully: the orchestrator proceeds with the standard manual run-phase flow (per-turn progression, no autonomous loop) rather than failing. The version preflight + capability gate implementation is deferred to the autonomy-config follow-up; this section documents the degradation contract.

### Cross-references (cite, do not restate)

- `.claude/rules/moai/workflow/goal-directive.md` — `/goal` semantics (the evaluator judges the transcript only — it neither runs tools nor opens any path; the `max N turns` bound; clear-on-`/clear`).
- `.claude/rules/moai/workflow/orchestration-mode-selection.md` § C.3 — Mode 6 (Workflow) capability gate (Implementation Kickoff Approval-passed + preferences-collected preconditions; scaling-not-nesting; the named-script-API prohibition).
- `.claude/rules/moai/workflow/dynamic-workflows.md` — the Workflow primitive (no mid-run user input; Implementation Kickoff Approval unaffected).
- the Implementation Kickoff Approval mandatory-restoration policy — the score-independent human gate this section preserves.

---

## Recursive Self-Diagnosis Loop (bounded)

This section wires the IGGDA Phase 2 bounded recursive self-diagnosis loop (D3) — the "self-audit" layer that handles MECHANICAL code failures fast during run-phase autonomy. It is the DIAGNOSE-PATCH-VERIFY pattern inherited from the `cycle_type=autofix` contract (see `.claude/rules/moai/development/manager-develop-prompt-template.md` § cycle_type Mode Reference), projected onto the run-phase autonomy loop. The loop is BOUNDED (max 3 iterations) and SEMANTIC-FAILURE-SAFE (data race / deadlock / panic / test assertion failure escalate immediately, NEVER auto-patched).

### 1. Mechanical vs semantic failure classification

[ZONE:Evolvable] [HARD] When a failure surfaces during the autonomous run-phase loop, the recursive self-diagnosis sub-agent MUST first classify it:

| Failure type | Examples | Loop action |
|--------------|----------|-------------|
| **Mechanical** | lint rule violation, type error, build error, missing import, formatting drift, gofmt divergence | DIAGNOSE-PATCH-VERIFY (max 3 iterations) |
| **Semantic** | data race, deadlock, panic, test assertion failure, concurrency hazard | IMMEDIATE escalate (NEVER auto-fix) |

The classification is grounded in `run.md §4 Semantic-failure escalation (HARD)` (the existing run-phase autonomy section) + `ci-autofix-protocol.md` (the mechanical-autofix max-3 bound) + `.claude/rules/moai/workflow/runtime-recovery-doctrine.md` §3 (the 5 circuit-breaker invariants). **Test assertion failures are SEMANTIC** — a failing assertion indicates a behavior contract was violated, not a mechanical error (REQ-IGGDA-016, AC-IGGDA-023). Mis-classifying a test assertion failure as mechanical and auto-patching it would mask the real behavior bug.

### 2. The DIAGNOSE-PATCH-VERIFY pattern (mechanical failures only)

For a MECHANICAL failure, the recursive self-diagnosis sub-agent executes one iteration:

1. **DIAGNOSE** — read the failing output (lint message, build error, type error). Identify the root cause. Challenge the diagnosed root cause once ("How do we know this is the cause, not a symptom?") before patching — do NOT stop at the first plausible cause.
2. **PATCH** — apply a minimal fix addressing the root cause WITHOUT expanding scope. The patch surface is the SPEC scope ONLY.
3. **VERIFY** — re-run the failing check locally. On exit 0, advance. On still-failing, increment the iteration counter and repeat from DIAGNOSE.

The sub-agent runs in FOREGROUND (`run_in_background: false`) — it patches code, so the background-write prohibition (`agent-common-protocol.md` § Background Agent Execution) binds.

### 3. Max-3-iteration bound (HARD)

[ZONE:Evolvable] [HARD] The loop attempts AT MOST 3 iterations. When iteration 3 completes without a VERIFY exit 0, the loop HALTS — iteration 4 is PROHIBITED. The orchestrator is signaled to run an `AskUserQuestion` escalation presenting the user with at least: (a) continue with manual investigation, (b) revert the offending change and re-plan, (c) abort with structured failure report. There is no auto-resume — the human MUST decide.

The max-3 bound is inherited from two sources:
- `.claude/rules/moai/workflow/ci-autofix-protocol.md` — the autofix loop's per-PR-push max-3-iteration contract.
- `.claude/rules/moai/workflow/runtime-recovery-doctrine.md` §3 invariant 1 — `MAX_CONSECUTIVE_AUTOCOMPACT_FAILURES=3` (book1 ch06 §5). Three iterations is the empirically-validated threshold: a mechanical failure is either patched (iterations 1–2 typically succeed) or is actually semantic-misclassified (iteration 3 fails → escalate, and the human re-classifies). Raising the bound risks the death-spiral; lowering it under-fixes easily-patched mechanical failures.

### 4. Semantic failures escalate IMMEDIATELY (HARD)

[ZONE:Evolvable] [HARD] When the failure classifier determines the failure is SEMANTIC (data race, deadlock, panic, test assertion failure, concurrency hazard), the loop halts IMMEDIATELY — NO DIAGNOSE-PATCH-VERIFY attempt. The orchestrator is signaled to run an `AskUserQuestion` human escalation. Semantic failures are NEVER auto-patched. This inherits `run.md §4 Semantic-failure escalation (HARD)` and `ci-autofix-protocol.md` (CONST-V3R5-010 — semantic failures MUST NOT be auto-fixed without human approval).

### 5. The 5 circuit-breaker invariants (runtime-recovery-doctrine §3 compliance)

The loop complies with all 5 circuit-breaker invariants from `.claude/rules/moai/workflow/runtime-recovery-doctrine.md` §3:

1. **Max-3 same-rung failures → escalate rung**: the max-3-iteration bound IS this invariant's projection. After 3 failed iterations, the loop escalates (does NOT attempt a 4th).
2. **`hasAttemptedReactiveCompact` no-self-loop**: within a single turn, the loop does NOT re-attempt the same DIAGNOSE-PATCH-VERIFY if it already failed this turn. Each iteration is a new turn.
3. **Compact-can-PTL last-resort escape**: if the loop itself hits PTL (the recovery triggers the error it is recovering from), the loop falls to rung-4 abort + preserve (persist state to `progress.md`), NOT another patch.
4. **Abort-closes-ledger**: on halt (iteration 3 OR semantic failure OR PTL), the loop persists its state to `progress.md §E Recursive Self-Diagnosis Log` before the session ends. An abort that leaves `progress.md` stale forces the next session to rediscover in-flight state — a silent restart disguised as a resume.
5. **Narrative-consistency**: across compact/recovery boundaries, the loop's state is reported via the 5-Section Evidence-Bearing Report format (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk per `.claude/rules/moai/core/verification-claim-integrity.md` §3).

### 6. Iteration logging contract

[ZONE:Evolvable] [HARD] Each recursive self-diagnosis iteration MUST be appended to `.moai/specs/<SPEC-ID>/progress.md` under a `## §E Recursive Self-Diagnosis Log` section with: iteration number, failure classification (mechanical/semantic), root-cause summary, patch summary, VERIFY result, and (on halt) the escalation reason. This is the ledger the abort-closes-ledger invariant (§5 #4) persists. The log is grep-verifiable:

```bash
grep -A 10 "Recursive Self-Diagnosis Log" .moai/specs/<SPEC-ID>/progress.md
```

### 7. Forbidden-paths list (PATCH scope discipline)

[ZONE:Evolvable] [HARD] The PATCH step MUST NOT modify:
- `.env`, `.env.*`, credentials files, secrets
- `scripts/ci-watch/run.sh` or any Wave 2 infrastructure scripts
- Any file outside the SPEC's declared run-phase scope (plan.md §A EXTEND envelope)

The patch surface is the SPEC scope ONLY. This inherits CONST-V3R5-011 + CONST-V3R5-013 (the autofix protected-files list) + `manager-develop-prompt-template.md` § cycle_type=autofix PATCH step. A PATCH that reaches outside the SPEC scope is a scope-discipline violation (agent-common-protocol.md § Agent Core Behaviors #5 Maintain Scope Discipline).

### 8. Flat hierarchy (no agent spawns agent)

The recursive self-diagnosis sub-agent is spawned BY THE ORCHESTRATOR (not by manager-develop — per Anthropic's published sub-agent guidance, subagents cannot spawn other subagents). When the sub-agent needs to delegate (e.g., it hits a blocker requiring user input), it returns a structured blocker report; the orchestrator runs `AskUserQuestion` and re-delegates. The sub-agent NEVER calls `AskUserQuestion` directly (the asymmetric orchestrator-subagent boundary).

### 9. Relationship to the IGGDA pipeline

This loop is the Phase 2 "self-audit" layer of the IGGDA 4-phase pipeline (see `.claude/rules/moai/workflow/orchestration-mode-selection.md` §I). It is COMPLEMENTARY to — NOT a replacement for — the independent audits (plan-auditor Phase 1, sync-auditor Phase 3). Self-audit handles mechanical code failures fast (bounded loop, no human in the loop for easy cases); independent audit handles SPEC-quality assurance (fresh-context skeptical evaluation). See `orchestration-mode-selection.md` §J.3 for the disambiguation.

