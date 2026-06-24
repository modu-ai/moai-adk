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

<!-- TRACE PROBE: workflow-split baseline trace mechanism -->
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

**Harness levels**: `minimal` → skip optional phases | `standard` → all phases | `thorough` → GAN-loop Sprint Contract Protocol + sync-auditor

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

A CI audit verifies the literal `MODE_UNKNOWN` sentinel remains present in this skill body (shared with `design.md`). `MODE_UNKNOWN` is emitted when `--mode <value>` is supplied to `/moai run` but `<value>` is not in the valid set `{autopilot, loop, team, pipeline}` (note: pipeline is itself rejected with the separate `MODE_PIPELINE_ONLY_UTILITY` sentinel — see line 71). The complementary `MODE_PIPELINE_ONLY_UTILITY` and `MODE_TEAM_UNAVAILABLE` sentinels are documented in this skill body and in `design.md`.

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

### 3. Autonomy invariants (cite, do not restate — full doctrine in canonical rules)

The following HARD invariants govern the `ac_converge` loop. Each is the canonical rule's render surface here; the rule is the SSOT.

- **Transcript-measurability**: the `acceptance.md` reference NAMES where the AC list lives — it is NOT a path the `/goal` evaluator opens. The Haiku evaluator judges only what the orchestrator SURFACES into the transcript (per-AC PASS line, `go test ./...` exit 0, `git status`).
- **Semantic-failure escalation (HARD)**: on a data race / deadlock / panic / test assertion failure surfaced during the loop, clear the `/goal` and escalate via `AskUserQuestion` — NEVER auto-fix a semantic failure (per `ci-autofix-protocol.md` semantic-failure-handling).
- **Non-substitution (HARD)**: the goal removes per-turn STOP prompts only. It does NOT authorize bypassing Implementation Kickoff Approval (already cleared), PR creation, or any destructive operation — those remain separately-surfaced explicit gates.
- **Blocker reports, never user prompts**: a `/goal`-turn or Mode 6 Workflow agent lacking input returns a structured blocker report; the orchestrator runs `AskUserQuestion` and re-delegates (asymmetric boundary per `agent-common-protocol.md` § User Interaction Boundary).
- **Graceful degradation**: when `/goal` is unavailable (runtime < v2.1.139, or hooks disabled), run-phase autonomy degrades to the standard manual per-turn flow rather than failing.

### Cross-references (cite, do not restate)

- `.claude/rules/moai/workflow/goal-directive.md` — `/goal` semantics (transcript-only evaluation; `max N turns` bound; clear-on-`/clear`).
- `.claude/rules/moai/workflow/orchestration-mode-selection.md` § C.3 — Mode 6 (Workflow) capability gate (Implementation Kickoff Approval-passed + preferences-collected; scaling-not-nesting; named-script-API prohibition) + §I IGGDA 4-phase pipeline + §J.3 self-audit-vs-independent-audit disambiguation.
- `.claude/rules/moai/workflow/dynamic-workflows.md` — the Workflow primitive (no mid-run user input; Implementation Kickoff Approval unaffected).
- `.claude/rules/moai/workflow/runtime-recovery-doctrine.md` §3 — the 5 circuit-breaker invariants the bounded self-diagnosis loop (below) complies with.
- `.claude/rules/moai/workflow/ci-autofix-protocol.md` + `.claude/rules/moai/development/manager-develop-prompt-template.md` § cycle_type Mode Reference — the DIAGNOSE-PATCH-VERIFY max-3 mechanical-autofix contract the loop inherits.

---

## Recursive Self-Diagnosis Loop (bounded — DIAGNOSE-PATCH-VERIFY)

The IGGDA Phase 2 bounded self-diagnosis loop handles MECHANICAL run-phase failures fast (DIAGNOSE-PATCH-VERIFY, max 3 iterations) and escalates SEMANTIC failures immediately. It is the run-phase projection of the `cycle_type=autofix` DIAGNOSE-PATCH-VERIFY contract; the canonical doctrine lives in the cross-referenced rules above. Summary contract:

| Item | Contract | Canonical SSOT |
|------|----------|----------------|
| Classification | Mechanical (lint / type / build / import / format) → DIAGNOSE-PATCH-VERIFY; Semantic (data race / deadlock / panic / **test assertion failure**) → IMMEDIATE escalate | `runtime-recovery-doctrine.md` §3 + `ci-autofix-protocol.md` |
| Iteration bound | [HARD] max 3 iterations; iteration 4 PROHIBITED; on iteration-3 fail the orchestrator runs an `AskUserQuestion` escalation (continue / revert+re-plan / abort) with no auto-resume | `ci-autofix-protocol.md` max-3 + `runtime-recovery-doctrine.md` §3 invariant 1 |
| Semantic safety | [HARD] semantic failures NEVER auto-patched (the constitutional rule) | `ci-autofix-protocol.md` |
| PATCH scope | [HARD] SPEC scope ONLY; MUST NOT touch `.env*` / credentials / `scripts/ci-watch/run.sh` / files outside plan.md §A EXTEND envelope (the constitutional rule/013) | `manager-develop-prompt-template.md` § cycle_type=autofix |
| Foreground | sub-agent runs `run_in_background: false` (it patches code; background-write prohibition binds) | `agent-common-protocol.md` § Background Agent Execution |
| Flat hierarchy | spawned BY THE ORCHESTRATOR (not manager-develop — subagents cannot spawn subagents); blocker reports never direct user prompts | `agent-common-protocol.md` § User Interaction Boundary |
| Ledger | [HARD] each iteration appended to `progress.md` `## §E Recursive Self-Diagnosis Log` (iteration #, classification, root-cause, patch, VERIFY result, escalation reason); grep-verifiable via `grep -A 10 "Recursive Self-Diagnosis Log" .moai/specs/<SPEC-ID>/progress.md` | `runtime-recovery-doctrine.md` §3 invariant 4 (abort-closes-ledger) |

This loop is COMPLEMENTARY to the independent audits (plan-auditor Phase 1, sync-auditor Phase 3) — self-audit handles mechanical failures fast; independent audit handles SPEC-quality assurance. See `orchestration-mode-selection.md` §J.3.

