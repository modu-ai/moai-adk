# SPEC-V3R5-HARNESS-AUTONOMY-001 — Compact

> Token-efficient summary of full spec.md / plan.md / acceptance.md. Mega-Sprint W3 — T2 Standard. Issue #1022.

## What

Self-evolution mechanism for My Harness (EVOLVABLE zone, vision §3.1). Captures workflow observations → 4-Tier progression → 5-Layer Safety gating → user-approved my-harness-* file writes.

## Why

W1 (constitution dual-zone) provides classification SSOT but no runtime enforcement. Without W3 PreToolUse hook, a hallucinating harness-learner could overwrite Core MoAI paths. W3 makes the FROZEN/EVOLVABLE boundary actually safe at runtime.

## 4-Tier Pipeline (M2)

| Observations | Tier / Status | Action |
|--------------|---------------|--------|
| 1x | `observation` | logged only |
| 3x | `heuristic` | suggestion (manager-develop hints) |
| 5x | `rule` | graduation candidate |
| 10x | `high-confidence` | Tier 4: AskUserQuestion auto-propose (throttled) |
| 1x critical | `anti-pattern` | instant FROZEN |
| seed | `tier: 3` start | meta-harness inject (W4 content) |

## 5-Layer Safety (M3 — brownfield EXTEND)

Sequential L1 → L3 → L4 → L5 (sync), L2 async with VETO power post-L5:

- **L1 Frozen Guard** (PreToolUse hook, sync <10ms p99) — 8 sentinel `HARNESS_FROZEN_*_VIOLATION`. Wires to W1 zone-registry data SSoT.
- **L2 Canary** (async ~30s shadow eval on last 3 projects) — Canary Veto Policy E5: provisional apply on L5 approve → on FAIL auto-rollback + AskUserQuestion notification. Canary is the FINAL gate.
- **L3 Contradiction Detector** (sync <1s) — emits blocker report; subagent does NOT call AskUserQuestion (boundary).
- **L4 Rate Limiter** (sync <100ms) — 3/week + 24h cooldown + 50 active max.
- **L5 Human Oversight** (user-paced) — blocker report → orchestrator AskUserQuestion → re-delegate.

## L1 Frozen Guard 8 Sentinels (M3.1)

| Deny path | Sentinel |
|-----------|----------|
| `.claude/agents/moai/**` | `HARNESS_FROZEN_AGENT_VIOLATION` |
| `.claude/skills/moai-*/**` (EXCEPT `my-harness-*`) | `HARNESS_FROZEN_SKILL_VIOLATION` |
| `.claude/rules/moai/**` | `HARNESS_FROZEN_RULE_VIOLATION` |
| `.claude/commands/moai/**` | `HARNESS_FROZEN_COMMAND_VIOLATION` |
| `.claude/hooks/moai/**` | `HARNESS_FROZEN_HOOK_VIOLATION` |
| `.claude/output-styles/moai/**` | `HARNESS_FROZEN_OUTPUTSTYLE_VIOLATION` |
| `CLAUDE.md` | `HARNESS_FROZEN_INSTRUCTION_VIOLATION` |
| `.moai/config/sections/*.yaml` | `HARNESS_FROZEN_CONFIG_VIOLATION` |

Bypass: `MOAI_FROZEN_GUARD_BYPASS=moai-update-internal` env (CLI internal only).

## Throttling (M4) — 4 modes

`immediate` (default) / `batch` (window+max) / `quiet` (hour range) / `mute` (per-category). R11 timeout: 7-day auto-defer to quiet.

## CLI (M6) — 6 + 2 verbs

`status`, `apply`, `rollback` (idempotent), `disable`, `mute`, `verify --determinism` (stub) + `mute-list`, `unmute`. All support `--json`.

## Brownfield Strategy (§1.5)

- **PRESERVE**: `internal/harness/{applier,layer1..5 (trigger verifier — different concern),...}` 20+ files
- **EXTEND**: `internal/harness/safety/{pipeline,canary,frozen_guard,contradiction,rate_limit,oversight}.go` (stubs → production); `internal/harness/{observer,learner,types}.go`; `internal/hook/pre_tool.go` (653 LOC, +8 sentinels)
- **NEW**: `internal/harness/{capture,tier,throttle,seeds}/` packages; `internal/cli/harness.go`; `.claude/agents/moai/harness-learner.md`; 1 dummy seed fixture

## Dependencies

- **Hard**: W1 CONSTITUTION-DUAL-001 (COMPLETE, main `7bd23bb69`) — data SSoT only per W1 EXCL-001
- **Parallel** (no race): W2 CORE-SLIM-001 (COMPLETE)
- **Unblocks**: W4 PROJECT-MEGA-001 — needs W3 substrate before seed content + meta-harness 7-Phase + deterministic generation

## Scope Exclusions (10 EXCL-HRA-*)

Deterministic generation (W4) · 8 baseline seed content (W4) · `/moai project --refresh` (W4) · meta-harness 7-Phase (W4) · project-specific my-harness gen (W4) · migration tooling · LLM-based capture · cross-repo sync · recursive self-introspection · past-30-day rollback extension.

## Estimate

~25 files · ~2850 LOC (code + tests) · 6 milestones M1-M6 · single SPEC, single run-phase PR (W1/W2 precedent).

## Counts at a glance

- 38 EARS REQs (REQ-HRA-001..038)
- 14 binary ACs (AC-HRA-001..014) + 6 EC + 5 Risk + 1 Constraint AC (C-HRA-008)
- 10 NFRs (R-HRA-S1..S4, Q1..Q2, A1, O1, I1, T1)
- 14 sentinels total: 8 `HARNESS_FROZEN_*` + 6 `HARNESS_LEARNING_*` (plan §3)
- 6 milestones M1..M6
- 10 EXCL-HRA-001..010

## Plan-Auditor target

- iter1: REVISE ≥0.75 (identify BLOCKING/SHOULD defects)
- iter2: **PASS ≥0.90** (W1 precedent 0.71 → 0.96; W3 expected ≥0.902 per issue #1022)
