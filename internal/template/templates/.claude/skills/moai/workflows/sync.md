---
description: >
  Synchronizes documentation with code changes, verifies project quality,
  and finalizes pull requests. Third step of the Plan-Run-Sync workflow.
  Includes deep code review with auto-fix, coverage analysis with test generation,
  SPEC divergence analysis, project document updates, and Context Memory generation.
  Use when documentation sync, PR creation, or quality verification is needed.
user-invocable: false
metadata:
  version: "3.8.0"
  category: "workflow"
  status: "active"
  updated: "2026-05-17"
  tags: "sync, documentation, pull-request, quality, verification, pr, context-memory"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["sync", "docs", "pr", "documentation", "pull request", "changelog", "readme"]
  agents: ["manager-docs", "manager-git"]
  phases: ["sync"]
---

<!-- TRACE PROBE: workflow-split baseline trace mechanism -->
<!-- Activated by MOAI_TRACE_PHASES=1 environment variable -->
<!-- Emits one line per Phase entry/exit to stderr in format: [trace] /moai sync Phase <N> <enter|exit> -->

# Sync Workflow Entry Router

이 파일은 `/moai sync` 진입점 라우터입니다. 각 Phase 그룹은 on-demand로 해당 sub-skill을 `Read`하여 로드합니다.

## Phase Owners (per the canonical agent catalog policy)

Phase Owners: `manager-docs` (sync-phase artifact authoring — CHANGELOG.md + README.md + docs-site + progress.md §F.3 + frontmatter `in-progress → implemented` transition for all SPEC artifacts; MUST NOT modify spec.md/plan.md/acceptance.md body content per `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix) + `manager-git` (PR creation per branching strategy when Tier L OR `--pr` flag per the canonical Tier-based PR routing policy).

Sync-phase quality gate (per the canonical sync-phase quality gate policy) is enforced by the `.claude/hooks/moai/sync-phase-quality-gate.sh` Stop hook — lint + test + coverage delta verification + dependency manifest audit. The hook returns exit 2 to block sync completion on lint/test failure or coverage regression > 5pp. The hook replaces the prior pattern of spawning an inline quality agent for coverage and security analysis during sync (that agent is archived per `.claude/rules/moai/workflow/archived-agent-rejection.md` §C row 2; the Stop hook is its canonical replacement).

## Phase Routing Table

| Phase Group | Sub-skill 경로 | 내용 |
|------------|----------------|------|
| Phase 1: Pre-Sync Context + Deployment Readiness | `Read workflows/sync/quality-gates-context.md` | Purpose/Scope/Input/Mode/Flags/Context Loading, Phase 1 HUMAN GATE 1, Phase 2 DB Check, Phase 3 Deployment Readiness |
| Phase 7~Phase 10: Quality Verification | `Read workflows/sync/quality-gates-quality.md` | Phase 7 Quality Check, Phase 8 Security Scan, Phase 9 MX Tag Validation, Phase 10 Coverage Analysis |
| Phase 11~Phase 12: Analysis + Doc Sync | `Read workflows/sync/doc-execution.md` | Phase 11 Analysis, HUMAN GATE 2 Documentation Scope, Phase 12 Execute Doc Synchronization |
| Phase 13~Phase 14: Git Delivery + Completion | `Read workflows/sync/delivery.md` | Phase 13 Git Operations, Phase 14 Completion, GitStrategy PR-ready transition, Graceful Exit, Test Scenarios |

## HUMAN GATE Map

| GATE | Location | Formal ID | Trigger |
|------|----------|-----------|---------|
| GATE 1: Pre-Sync Quality | `sync/quality-gates-context.md` | `gate-sync-1` | Working tree + all tests pass check before entering Phase 3 |
| GATE 2: Documentation Scope | `sync/doc-execution.md` | `gate-sync-2` | User reviews divergence report and approves doc regeneration scope |

> Note: Additional AskUserQuestion decision points exist in Phase 1 (gate failure), Phase 3 (test failure), Phase 6 (breaking changes), Phase 7 (test failure), Phase 8 (security critical), Phase 13 (CI mirror failure), and Phase 14 (next steps). These are inline decision gates, not named evolvable GATEs.

## Invocation Flow

```
/moai sync [mode] [--pr] [--merge] [--skip-mx]
  ├── [trace] /moai sync Phase 1 enter
  │   Read workflows/sync/quality-gates-context.md  → HUMAN GATE 1 + Deployment Readiness
  ├── [trace] /moai sync Phase 7 enter
  │   Read workflows/sync/quality-gates-quality.md  → Quality/Security/MX/Coverage
  ├── [trace] /moai sync Phase 11 enter
  │   Read workflows/sync/doc-execution.md          → Divergence Analysis + HUMAN GATE 2 + Doc Sync
  └── [trace] /moai sync Phase 13 enter
      Read workflows/sync/delivery.md               → Git Ops + CI Mirror + PR + Auto-Merge + Completion
```

## Quick Reference

**Purpose**: 코드 변경사항과 문서를 동기화하고, PR을 생성하여 SPEC 사이클을 완료합니다.

**Input**: `$ARGUMENTS` = `[mode] [path] [flags]`

**Modes**: `auto` (기본) | `force` | `status` | `project`

**Flags**: `--pr` (PR 생성) | `--merge` (deprecated, auto-merge) | `--skip-mx` (MX 검증 스킵)

**HUMAN GATEs**: GATE 1 (quality-gates-context.md §Phase 1) → GATE 2 (doc-execution.md §Step 1.6)

**Status mode early exit**: Phase 9 완료 후 보고서 출력 및 종료 (Phase 11+ 실행 안 함)

**Chain entry**: sync may be entered via auto-chain from run-phase completion (a `full-pipeline` contract — announced in the transcript, no extra approval round at the run→sync phase boundary) or via explicit `/moai sync` invocation (a `single-phase` contract). Either way the sync-internal HUMAN GATEs (`gate-sync-1`, `gate-sync-2`) fire unchanged, and a FAIL/INCONCLUSIVE sync-audit verdict or a blocking sync quality gate HALTS the chain — no auto-completion past a failing gate.

## On-Demand Sub-skill Loading

각 Phase 진입 시점에 해당 sub-skill을 로드합니다:

```
# Phase 1 (Pre-Sync Quality Gate) 진입 시
Read .claude/skills/moai/workflows/sync/quality-gates-context.md

# Phase 7 (Quality Verification) 진입 시
Read .claude/skills/moai/workflows/sync/quality-gates-quality.md

# Phase 11 (Analysis and Planning) 진입 시
Read .claude/skills/moai/workflows/sync/doc-execution.md

# Phase 13 (Git Operations) 진입 시
Read .claude/skills/moai/workflows/sync/delivery.md
```

## Custom Harness Extension

@.moai/harness/sync-extension.md

*(이 파일은 `/moai project --harness`로 생성됩니다. 파일이 없으면 자동으로 skip됩니다.)*

## Sentinel Error Keys

A CI audit verifies the literal `MODE_PIPELINE_ONLY_UTILITY` sentinel remains present in this skill body (shared with `design.md`). Passing `--mode pipeline` to `/moai sync` is rejected because sync is a Multi-Agent subcommand; pipeline mode is reserved for utility subcommands.

## Routing Ledger Recording

At sync dispatch, the orchestrator records the routing decision to the routing-ledger via `moai harness ledger record` (per the SKILL.md router recording obligation). At the sync-phase quality gate, it appends the gate result as machine evidence via `moai harness ledger evidence --kind gate_exit --value <exit> --terminal --ref "sync-phase-quality-gate"`. Outcome is derived from machine evidence only — never supplied as an input. The recording is opt-in and fail-open; it never blocks the sync phase.
