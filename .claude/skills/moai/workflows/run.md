---
description: >
  DDD/TDD implementation workflow for SPEC requirements. Second step
  of the Plan-Run-Sync workflow. Routes to manager-develop based
  on quality.yaml constitution.development_mode setting.
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

<!-- TRACE PROBE: workflow-split baseline trace mechanism -->
<!-- Activated by MOAI_TRACE_PHASES=1 environment variable -->
<!-- Emits one line per Phase entry/exit to stderr in format: [trace] /moai run Phase <N> <enter|exit> -->

# Run Workflow Entry Router

이 파일은 `/moai run` 진입점 라우터입니다. 각 Phase는 on-demand로 해당 sub-skill을 `Read`하여 로드합니다.

## Phase Owners (per the canonical agent catalog policy)

Phase Owners: `manager-develop` (run-phase implementation — single-spawn per Anthropic's coding-task parallelism caveat "most coding tasks involve fewer truly parallelizable tasks than research"; `cycle_type` ∈ `{tdd, ddd, autofix}` per the canonical cycle-type contract) + `manager-git` (Tier L PR creation OR `--pr` flag per the canonical Tier-based PR routing policy) + `Explore` (read-only investigation when scope discovery needed).

Skill injection: at each `manager-develop` spawn the orchestrator injects the cycle_type skill (`moai-workflow-tdd` | `moai-workflow-ddd`) plus 0-3 domain `moai-ref-*` skills matched to the mission domain, as `At start, invoke Skill("<name>") for <reason>` lines, per `.claude/rules/moai/workflow/skill-routing.md` §1 and the delegation map (`.moai/config/sections/delegation.yaml`).

Phase 4 Mode Selection: orchestrator autonomous decision over the 6-mode catalog (trivial / background / agent-team / parallel / sub-agent / workflow) per `.claude/rules/moai/workflow/orchestration-mode-selection.md` §A, logged at `.moai/specs/SPEC-{ID}/progress.md` § Phase 4 Mode Selection. Phase 4 SHOULD be invoked before any manager-develop spawn for SPECs sized ≥ Tier M. The `--mode` dispatch axis below is a SEPARATE axis — see that rule's §G.1 crosswalk (correspondence, not merge).

`cycle_type=autofix` mode: `/moai fix` workflow integration delegates to manager-develop with the utility-class pipeline 3-phase contract (localize → repair → validate per `.claude/rules/moai/workflow/spec-workflow.md` § Subcommand Classification) and the max-3-iteration contract per `.claude/rules/moai/workflow/ci-autofix-protocol.md`.

## Phase Routing Table

| Phase Group | Sub-skill 경로 | 내용 |
|------------|----------------|------|
| Phase 0: Context Loading | `Read workflows/run/context-loading.md` | Mode dispatch, UltraThink, harness level, context loading, worktree path rules |
| Phase 1~1.8: Phase Execution | `Read workflows/run/phase-execution.md` | Plan Audit Gate, environment assessment, JIT language detection, scale-based mode, analysis/planning, task decomposition, development mode routing |
| Phase 11~4: Implementation | `Read workflows/run/task-decomposition.md` | DDD/TDD cycles, quality validation (Phase 13/2.8), git operations (Phase 19), completion guidance (Phase 20) |
| Mode Routing + Completion | `Read workflows/run/mode-orchestration.md` | Execution mode gate, mode dispatch routing, context propagation, completion criteria, verify exit gate (factory contract), test scenarios |

## Fan-Out Index

| Fan-Out ID | Trigger condition | Target file | What is parallelised |
|---|---|---|---|
| `FO-RUN-1` | the MX scan target spans many files across several packages | `workflows/run/phase-execution.md` | Phase 1.8 MX context-map scan — one read-only shard per package |
| `FO-RUN-2` | a milestone's RED stage spans several independent test targets | `workflows/run/task-decomposition.md` | RED-stage test drafting — one read-only drafter per target |
| `FO-RUN-3` | the quality-evidence fan-out script is on disk AND the runtime supports dynamic workflows | `workflows/run/task-decomposition.md` | the Phase 13 / 16 / 17 quality band — one parallel four-dimension evidence pass |
| `FO-RUN-4` | the modified files span several packages | `workflows/run/task-decomposition.md` | Phase 18 MX tag scan — one read-only shard per package; tag application stays single-writer |

## Invocation Flow

```
/moai run SPEC-XXX
  ├── [trace] /moai run Phase 0 enter
  │   Read workflows/run/context-loading.md  → Mode dispatch + context setup
  ├── [trace] /moai run Phase 1 enter
  │   Read workflows/run/phase-execution.md  → Phase Sequence (0.5~1.8) + mode routing
  ├── [trace] /moai run Phase 11 enter
  │   Read workflows/run/task-decomposition.md → Implementation + quality + git
  └── [trace] /moai run Mode enter
      Read workflows/run/mode-orchestration.md → Mode dispatch + completion criteria
```

## Quick Reference

**Purpose**: SPEC 요구사항을 DDD 또는 TDD 방법론으로 구현합니다.

**Input**: `$ARGUMENTS` = SPEC-ID (예: `SPEC-AUTH-001`)

**Development mode**: `.moai/config/sections/quality.yaml` `constitution.development_mode` 설정 (`ddd` 또는 `tdd`)에 따라 자동 선택.

**Mode dispatch** (`--mode` flag):
- `autopilot` (기본): Phase 4 scale-based 선택 후 Phase 11/2B 실행
- `loop`: Ralph engine 위임 (see `loop.md`)
- `team`: RETIRED — `--mode team` emits `MODE_TEAM_UNAVAILABLE` and falls back to `autopilot` (Agent Teams static layer retired)
- `pipeline`: REJECTED — `MODE_PIPELINE_ONLY_UTILITY` 오류 반환

**Harness levels**: `minimal` → skip optional phases | `standard` → all phases | `thorough` → GAN-loop Sprint Contract Protocol + sync-auditor

**Phase 1 (Plan Audit Gate)**: 모든 harness level에서 SKIP 불가. SPEC plan 아티팩트 독립 감사 필수. The audit always runs (once) for every tier — Tier S changes only the re-run loop, never whether the audit runs: for Tier S SPECs the gate applies a single audit pass by default (the iterative verdict re-execution loop defaults OFF; a PASS verdict is final without a score-threshold re-run, while FAIL/INCONCLUSIVE still halts and escalates as today). Tier M/L iterative audit behavior is unchanged. See `workflows/run/phase-execution.md` § Tier S Single-Audit-Pass Default.

**Worktree path rules**: [HARD] 모든 에이전트 프롬프트에 절대 경로 금지. project-root-relative 경로 사용.

**Chaining (single-phase contract)**: an explicit `/moai run` invocation carries a `single-phase` pipeline contract — on run-phase completion, the sync chain is surfaced as the "(Recommended)" first option of the next-step AskUserQuestion; it never fires silently. The `full-pipeline` auto-chain applies only to the default `/moai` route (see `workflows/moai.md` § run→sync chaining policy).

## On-Demand Sub-skill Loading

각 Phase 진입 시점에 해당 sub-skill을 로드합니다:

```
# Phase 0: Context loading 및 mode dispatch 시작 시
Read .claude/skills/moai/workflows/run/context-loading.md

# Phase 1 (Plan Audit Gate) 진입 시
Read .claude/skills/moai/workflows/run/phase-execution.md

# Phase 11 (Implementation) 진입 시
Read .claude/skills/moai/workflows/run/task-decomposition.md

# Mode dispatch 또는 completion criteria 확인 시
Read .claude/skills/moai/workflows/run/mode-orchestration.md
```

## Custom Harness Extension

@.moai/harness/run-extension.md

*(이 파일은 `/moai project --harness`로 생성됩니다. 파일이 없으면 자동으로 skip됩니다.)*

## Sentinel Error Keys

A CI audit verifies the literal `MODE_UNKNOWN` sentinel remains present in this skill body (shared with `design.md`). `MODE_UNKNOWN` is emitted when `--mode <value>` is supplied to `/moai run` but `<value>` is not in the valid set `{autopilot, loop, team, pipeline}` (note: pipeline is itself rejected with the separate `MODE_PIPELINE_ONLY_UTILITY` sentinel — see line 71). The complementary `MODE_PIPELINE_ONLY_UTILITY` and `MODE_TEAM_UNAVAILABLE` sentinels are documented in this skill body and in `design.md`.

Ordering invariant (read before the autonomy section below): the Implementation Kickoff Approval `AskUserQuestion` human gate is always cleared FIRST; any run-phase autonomy set is downstream of it. The next section documents that ordering and the autonomy condition together.

## Run-phase Autonomy (ac_converge)

This section wires the run-phase autonomy mechanisms — the Implementation Kickoff Approval human-gate ordering reference and the `ac_converge` goal condition — into a single co-located place. The two parts are ORDERED: the Implementation Kickoff Approval `AskUserQuestion` human gate is described FIRST (it must be cleared before any autonomy begins), then the `ac_converge` arming (entered only after Implementation Kickoff Approval approval).

> **Progression-mode axis (autonomous vs. semi-autonomous)**: the Implementation Kickoff Approval gate also offers a progression-mode choice — autonomous (default; the loop continues without per-turn prompts) or semi-autonomous (the `stop-goal` hook emits a checkpoint-signal each turn for orchestrator-side `AskUserQuestion` confirmation). This axis selects ONLY post-approval progression; the gate stays mandatory in both modes. See `.claude/skills/moai/workflows/goal.md` § Progression Mode.

### 1. Implementation Kickoff Approval ordering (the human gate comes first)

[HARD] Before any run-phase autonomy (arming a goal, a Mode 6 Workflow launch, or any autonomous loop), the orchestrator MUST have already obtained explicit Implementation Kickoff Approval approval. Implementation Kickoff Approval is the plan→run HUMAN GATE: a mandatory orchestrator-issued `AskUserQuestion` round (run-phase entry / further review / abort, first option marked "(Recommended)") presented after Phase 1 (Plan Audit Gate) and before Phase 4 (Mode Selection). The orchestrator emits this gate; it is never embedded inside a subagent body (subagents cannot prompt the user — the asymmetric boundary in `.claude/rules/moai/core/agent-common-protocol.md` § User Interaction Boundary).

[HARD] Implementation Kickoff Approval is **score-independent**: the orchestrator emits the Implementation Kickoff Approval `AskUserQuestion` gate **regardless of the plan-auditor score**, including the high skip-eligible case. Skip-eligibility (a high autonomous-bypass score) applies ONLY to Phase 1 plan-auditor verdict re-execution — NOT to Implementation Kickoff Approval. A high plan-auditor score never authorizes skipping the Implementation Kickoff Approval human gate. This is the Implementation Kickoff Approval mandatory-restoration invariant per the Implementation Kickoff Approval mandatory-restoration policy.

Because Implementation Kickoff Approval also drains all user preferences (Tier, mode preference, PR strategy), the orchestrator collects every preference at this gate BEFORE launching any autonomy — goal-loop turn agents and Mode 6 Workflow agents cannot prompt the user mid-run, so the one decision that must involve the user is taken here.

### 2. The `ac_converge` goal condition (armed only after Implementation Kickoff Approval approval)

ONLY after Implementation Kickoff Approval approval is obtained, the orchestrator MAY arm the `ac_converge` goal via `/moai goal "<condition>"` to grant phase-internal autonomy (it removes per-turn STOP prompts so the run-phase loop continues until convergence). Because arming is **arm-only** — it records the condition but starts no work — the orchestrator arms it ALONGSIDE the run-phase work it is driving, never in place of that work (`goal-directive.md` § Goal-Presentation Timing). The condition is hard-coded inline (no registry dependency) and is authored entirely as **model conditions** — every predicate references a line the orchestrator surfaces in the conversation, so the evaluator judges it against the transcript rather than by opening a file:

```text
Every blocking acceptance criterion in
.moai/specs/SPEC-{ID}/acceptance.md has its PASS evidence surfaced in
the conversation (test output, build exit 0, or explicit AC-id: PASS
line); AND `go test ./...` exit 0 is surfaced; AND no test file outside
the SPEC scope was modified (surfaced via git status). Stop when all
hold.
On any semantic failure (data race, deadlock, panic, test assertion
failure), clear this goal and escalate via AskUserQuestion — do NOT
auto-fix semantic failures.
[PRECONDITION: Implementation Kickoff Approval user approval already obtained; this goal does NOT substitute for or bypass Implementation Kickoff Approval.]
```
> **Actual turn ceiling:** the legacy "Max 20 turns" clause above is NOT parsed — `parseCondition` matches only a trailing `exits <N>`, so a "Max 20 turns" / "stop after N turns" suffix has no mechanical effect. The `ac_converge` goal therefore arms at the default 30-turn ceiling and actually runs up to 30 turns (subject to `min(30, block cap)` and the stagnation guard). The literal is retained only as the historical reference for the author's intent. To bound a goal at a specific turn count, use `--max-turns N`; for an effectively-unbounded goal, use `--max-turns 0 --max-duration <seconds>`.

### 3. Autonomy invariants (cite, do not restate — full doctrine in canonical rules)

The following HARD invariants govern the `ac_converge` loop. Each is the canonical rule's render surface here; the rule is the SSOT.

- **Transcript-measurability**: the `acceptance.md` reference NAMES where the AC list lives — it is NOT a path the evaluator opens. Because every predicate above is a model condition, the `stop-goal` evaluator judges only what the orchestrator SURFACES into the transcript (per-AC PASS line, `go test ./...` exit 0, `git status`).
- **Semantic-failure escalation (HARD)**: on a data race / deadlock / panic / test assertion failure surfaced during the loop, clear the goal (`/moai goal clear`) and escalate via `AskUserQuestion` — NEVER auto-fix a semantic failure (per `ci-autofix-protocol.md` semantic-failure-handling).
- **Non-substitution (HARD)**: the goal removes per-turn STOP prompts only. It does NOT authorize bypassing Implementation Kickoff Approval (already cleared), PR creation, or any destructive operation — those remain separately-surfaced explicit gates.
- **Blocker reports, never user prompts**: a goal-loop turn or Mode 6 Workflow agent lacking input returns a structured blocker report; the orchestrator runs `AskUserQuestion` and re-delegates (asymmetric boundary per `agent-common-protocol.md` § User Interaction Boundary).
- **Graceful degradation**: the goal engine's evaluator IS a Stop hook (`moai hook stop-goal`), so `/moai goal` is unavailable when hooks are disabled (`disableAllHooks`, or `allowManagedHooksOnly` permitting only managed hooks). It carries no runtime-version floor of its own. When the engine is unavailable, run-phase autonomy degrades to the standard manual per-turn flow rather than failing.

### Cross-references (cite, do not restate)

- `.claude/rules/moai/workflow/goal-directive.md` — `/moai goal` semantics (mechanical + model conditions judged by the `stop-goal` evaluator at each turn-end; the turn-ceiling bound; per-session goal state, so a goal armed in a previous session is not in effect after `/clear`).
- `.claude/rules/moai/workflow/orchestration-mode-selection.md` § C.3 — Mode 6 (Workflow) capability gate (Implementation Kickoff Approval-passed + preferences-collected; scaling-not-nesting; named-script-API prohibition).
- `.claude/rules/moai/workflow/dynamic-workflows.md` — the Workflow primitive (no mid-run user input; Implementation Kickoff Approval unaffected).
- `.claude/rules/moai/workflow/runtime-recovery-doctrine.md` §3 — the 5 circuit-breaker invariants the bounded self-diagnosis loop (below) complies with.
- `.claude/rules/moai/workflow/ci-autofix-protocol.md` + `.claude/rules/moai/development/manager-develop-prompt-template.md` § cycle_type Mode Reference — the DIAGNOSE-PATCH-VERIFY max-3 mechanical-autofix contract the loop inherits.

---

## Recursive Self-Diagnosis Loop (bounded — DIAGNOSE-PATCH-VERIFY)

The bounded self-diagnosis loop handles MECHANICAL run-phase failures fast (DIAGNOSE-PATCH-VERIFY, max 3 iterations) and escalates SEMANTIC failures immediately. It is the run-phase projection of the `cycle_type=autofix` DIAGNOSE-PATCH-VERIFY contract; the canonical doctrine lives in the cross-referenced rules above. Summary contract:

| Item | Contract | Canonical SSOT |
|------|----------|----------------|
| Classification | Mechanical (lint / type / build / import / format) → DIAGNOSE-PATCH-VERIFY; Semantic (data race / deadlock / panic / **test assertion failure**) → IMMEDIATE escalate | `runtime-recovery-doctrine.md` §3 + `ci-autofix-protocol.md` |
| Iteration bound | [HARD] max 3 iterations; iteration 4 PROHIBITED; on iteration-3 fail the orchestrator runs an `AskUserQuestion` escalation (continue / revert+re-plan / abort) with no auto-resume | `ci-autofix-protocol.md` max-3 + `runtime-recovery-doctrine.md` §3 invariant 1 |
| Semantic safety | [HARD] semantic failures NEVER auto-patched (the constitutional rule) | `ci-autofix-protocol.md` |
| PATCH scope | [HARD] SPEC scope ONLY; MUST NOT touch `.env*` / credentials / CI watch infrastructure and workflow definitions / files outside plan.md §A EXTEND envelope (the constitutional rule/013) | `manager-develop-prompt-template.md` § cycle_type=autofix |
| Foreground | sub-agent runs `run_in_background: false` (it patches code; background-write prohibition binds) | `agent-common-protocol.md` § Background Agent Execution |
| Flat hierarchy | spawned BY THE ORCHESTRATOR (not manager-develop — subagents cannot spawn subagents); blocker reports never direct user prompts | `agent-common-protocol.md` § User Interaction Boundary |
| Ledger | [HARD] each iteration appended to `progress.md` `## §E Recursive Self-Diagnosis Log` (iteration #, classification, root-cause, patch, VERIFY result, escalation reason); grep-verifiable via `grep -A 10 "Recursive Self-Diagnosis Log" .moai/specs/<SPEC-ID>/progress.md` | `runtime-recovery-doctrine.md` §3 invariant 4 (abort-closes-ledger) |

This loop is COMPLEMENTARY to the independent audits (plan-auditor Phase 5, sync-auditor Phase 19) — self-audit handles mechanical failures fast; independent audit handles SPEC-quality assurance. See `orchestration-mode-selection.md` §J.3.

## Routing Ledger Recording

At run dispatch, the orchestrator records the routing decision to the routing-ledger via `moai harness ledger record` (per the SKILL.md router recording obligation). As run-phase gates complete, it appends machine evidence via `moai harness ledger evidence` — a terminal gate exit (`--kind gate_exit --value 0 --terminal --ref "go test ./..."`) or a verify-log path (`--kind verify_path --ref <.moai/state/verify/... log>`). Outcome is finalized from that machine evidence only — never supplied as an input. The recording is opt-in and fail-open; it never blocks the run phase.

