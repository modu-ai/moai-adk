---
name: moai-workflow-run
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
