---
id: SPEC-V3R5-HARNESS-AUTONOMY-001
title: "Harness Autonomy — Compact Extract"
version: "0.1.0"
status: draft
created: 2026-05-20
updated: 2026-05-20
author: GOOS Kim
priority: P0
phase: "v3.0.0 — Round 5"
module: "internal/harness + .moai/harness + .claude/skills/moai-harness-learner + internal/hook"
lifecycle: spec-anchored
tags: "harness, autonomy, self-evolution, 4-tier, 5-layer-safety, compact, mega-sprint, w3"
---

> Auto-extracted compact summary from spec.md / plan.md / acceptance.md. For full context, see those three files. Vision source: `.moai/research/harness-autonomy-vision-2026-05-18.md` (iter3).

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-05-20 | GOOS Kim (via MoAI orchestrator) | Initial compact extract |
| 0.1.0 | 2026-05-20 | GOOS Kim (via MoAI orchestrator) | Iteration 2 revision per plan-auditor iter 1 BLOCKING (B1-B5) + SHOULD (S1/S3/S4/S5/S6/S10) defects. Counts updated: 38 REQs (was 36, +037/+038), 14 ACs (was 12, +008b/+013/+014), 1 C-binary (C-HRA-008). Architecture caption updated — L1 = W3 IMPLEMENTS hook + reads W1 zone-registry data (not "W1 client"). L3 noted as blocker-report pattern, not direct AskUserQuestion. Brownfield reality: `internal/harness/` is NOT absent (30+ existing files; W3 extends `safety/` subdir + preserves layer*.go). |

---

## Mission

Mega-Sprint **W3**: My Harness (EVOLVABLE) layer가 워크플로우 실행 중 지속적으로 학습하고 자가 진화하는 메커니즘을 구현. Core MoAI는 변경하지 않음 (W1 Frozen Guard 보호). 사용자 directive (2026-05-18): "하네스만 자율 진화 하도록 한다."

---

## Architecture (5 Subsystems)

| Subsystem | Function | Module path |
|-----------|----------|-------------|
| M1 Lesson Capture | SubagentStop hook → heuristic match (<500ms) → observation emit | `internal/harness/capture/` |
| M2 Tier Engine | observations.yaml + state machine (1x→3x→5x→10x) + anti-pattern flag | `internal/harness/tier/` |
| M3 5-Layer Safety | L1 Frozen Guard (W3 implements hook + reads W1 zone-registry as data per B1) + L2 Canary (async, has veto) + L3 Contradiction (blocker-report pattern per B5) + L4 Rate Limiter (only purely-internal layer) + L5 Human Oversight (blocker-report pattern) | `internal/harness/safety/` (EXTEND existing) + `internal/hook/pre_tool.go` (EXTEND existing) |
| M4 Throttling + CLI | 4 modes (immediate/batch/quiet/mute) + 6 CLI verbs | `internal/harness/throttle/` + `internal/cli/harness_*.go` |
| M5 Cold-Start Seed | schema + load hook (content = W4 scope) | `internal/harness/seeds/` |

End-to-end test (M6) integrates all subsystems.

---

## Tier Engine

| Observations | Status | Action |
|---|---|---|
| 1x | `observation` | logged only |
| 3x | `heuristic` | manager-develop hint |
| 5x | `rule` | Sprint Contract candidate |
| 10x | `high-confidence` | AskUserQuestion auto-propose (throttling-aware) |
| 1x critical | `anti-pattern` | FROZEN immediate flag |
| Seed | `rule` (Tier 3 start) | cold-start mitigation |

Thresholds [1,3,5,10] FROZEN per REQ-HRA-007 (v3.5.0 fixed).

---

## 5-Layer Safety + Canary Veto Policy (E5)

| Layer | Sync | Budget | Veto |
|---|---|---|---|
| L1 Frozen Guard | sync | <10ms p99 (REQ-HRA-037) | YES (Vision §3.4 8 sentinels — catalog NOT W1) |
| L2 Canary | async ~30s | 30s soft / 60s hard | **YES (post-L5 veto, auto-rollback per E5 + B4)** |
| L3 Contradiction | sync | <1s | YES (blocker report → orchestrator AskUserQuestion per B5) |
| L4 Rate Limiter | sync | <100ms | YES (3/wk + 24h cd + 50 active + 48h post-veto cd per B4) |
| L5 Human Oversight | sync user-paced | unbounded | YES (blocker report → orchestrator AskUserQuestion final decision) |

Provisional apply when L5 approves before L2 completes; L2 FAIL = auto-rollback.

---

## File Touch List (modules to modify/create in run-phase)

### NEW Go packages

- `internal/harness/capture/{capture.go,capture_test.go}`
- `internal/harness/tier/{tier.go,observations.go,tier_test.go}`
- `internal/harness/safety/{pipeline.go,l1.go,l2_canary.go,l3.go,l4_ratelimit.go,l5_oversight.go,canary_veto.go,*_test.go}`
- `internal/harness/throttle/{throttle.go,throttle_test.go}`
- `internal/harness/seeds/{schema.go,loader.go,loader_test.go}`
- `internal/harness/integration_test.go`
- `internal/cli/harness.go` (parent) + `internal/cli/harness_{status,apply,rollback,disable,mute,verify}.go` (verb-per-file)

### MODIFIED existing

- `internal/hook/subagent_stop.go` (extend with harness-learner invocation)
- `.moai/config/sections/harness.yaml` (extend with `proposal.*` + `seeds.library_path`)
- `.moai/config/sections/workflow.yaml` (extend with `harness.proposal.*`)
- `.claude/skills/moai-harness-learner/SKILL.md` (W3 documents extension; actual edit in run-phase)
- `.claude/skills/moai-meta-harness/SKILL.md` (W3 documents seed integration point)

### NEW runtime data

- `.moai/harness/{observations.yaml,evolution-log.md,anti-patterns.yaml,proposal-queue.yaml,revert/}`
- `.claude/skills/moai-meta-harness/seeds/.gitkeep` (W4 fills content)

### NEW CI guard

- `internal/harness/sentinel_catalog_test.go` (verify HARNESS_LEARNING_* catalog matches plan.md §7)

---

## AC Count + Dependency Status

| Metric | Value |
|--------|-------|
| EARS REQs | 38 (REQ-HRA-001..036 + REQ-HRA-037 S6 + REQ-HRA-038 B2) |
| Binary ACs | 14 (AC-HRA-001..012 + AC-HRA-008b B4 + AC-HRA-013 S3 + AC-HRA-014 S6) |
| Edge Cases | 6 (EC-HRA-001..006) |
| Risk Mitigations | 5 (R-HRA-001..005) |
| Binary Constraint Verifications | 1 (C-HRA-008 S5 subagent boundary grep) |
| Exclusions | 10 (EXCL-HRA-001..010) |
| Total verification surface | 26 (14 + 6 + 5 + 1) |
| REQ ↔ AC traceability | 100% (every REQ has ≥1 AC, via primary or secondary mapping) |
| Open Questions (Vision §9) | 9 (5 RESOLVED in plan.md §10, 2 N/A, 2 deferred) |
| Iteration 2 BLOCKING resolved | 5 (B1 W1 mis-citation, B2 brownfield blindness, B3 seed dual-path, B4 Canary veto finality, B5 L3/L5 subagent boundary) |
| Iteration 2 SHOULD resolved | 6 (S1 field naming policy, S3 REQ-007 enforcement, S4 R11 timeout, S5 subagent grep, S6 L1 latency NFR, S10 M5 stub-only) |

Dependencies:
- **Hard**: W1 (CONSTITUTION-DUAL-001, COMPLETE) — zone-registry 111 entries (DATA SSOT only per W1 EXCL-001); W3 implements PreToolUse hook + Vision §3.4 8 HARNESS_FROZEN_* sentinels
- **Soft**: W2 (CORE-SLIM-001, COMPLETE) — moai-foundation-quality preload
- **Unblocks**: W4 (PROJECT-MEGA-001) — seed library content + meta-harness 7-Phase

---

## M1-M6 Milestone List

| Milestone | Priority | Dependencies | Deliverables |
|-----------|----------|--------------|--------------|
| M1 Lesson Capture | P0 | W1 | capture/ package + SubagentStop integration + observations.yaml |
| M2 Tier Engine | P0 | M1 | tier/ package + state machine + anti-pattern flag |
| M3 5-Layer Safety | P0 | M2+W1 | safety/ package (per-layer files) + Canary veto + provisional apply |
| M4 Throttling+CLI | P0 | M3 | throttle/ + 6 CLI verbs + workflow.yaml extension |
| M5 Cold-Start Seed | P1 | M2 | seeds/ loader + harness.yaml extension + empty seed dir |
| M6 Integration Test | P0 | M1-M5 | integration_test.go covering all ACs/ECs/Rs |

Sequential within single run-phase (manager-develop cycle_type=tdd, RED→GREEN→REFACTOR per milestone). ~25 files, ~2850 LOC (code + tests).

---

## Sentinel Catalog Extension

W3 introduces 10 new HARNESS_LEARNING_* sentinels (additive to the Vision §3.4-defined 8 HARNESS_FROZEN_* catalog):

`HARNESS_LEARNING_FROZEN_BLOCKED` (L1), `HARNESS_LEARNING_CANARY_FAILED` (L2 sync), `HARNESS_LEARNING_CANARY_VETO` (L2 post-L5), `HARNESS_LEARNING_CONTRADICTION` (L3), `HARNESS_LEARNING_RATELIMIT_EXCEEDED` (L4), `HARNESS_LEARNING_USER_REJECTED` (L5), `HARNESS_LEARNING_TIER_VIOLATION` (tier engine), `HARNESS_LEARNING_SCHEMA_DRIFT` (parse), `HARNESS_LEARNING_SEED_INVALID` (cold-start), `HARNESS_LEARNING_MUTE_INVALID_CATEGORY` (CLI).

CI guard (`internal/harness/sentinel_catalog_test.go`) verifies catalog ↔ plan.md §7 alignment.

---

## Key Constraints

- C-HRA-001: Frontmatter 12-field canonical, `created:`/`updated:`/`tags:` strict.
- C-HRA-002: 16-language neutrality — sentinel keys language-agnostic.
- C-HRA-003: Performance — capture <500ms p95, L1 <10ms p99, L3 <1s, L4 <100ms.
- C-HRA-004: No external dependencies (Go stdlib + yaml.v3 only).
- C-HRA-005: TRUST 5 — `internal/harness/` coverage ≥85%.
- C-HRA-006: `harness.yaml` `tier_thresholds: [1,3,5,10]` FROZEN.
- C-HRA-007: Vision §3.4 8 HARNESS_FROZEN_* catalog preserved unchanged (W3 only adds 10 HARNESS_LEARNING_* additive sentinels; Vision catalog itself is not modified).
- C-HRA-008: Subagent boundary — harness-learner NEVER invokes AskUserQuestion (blocker report only).

---

## Out of Scope (10 Exclusions)

EXCL-HRA-001 Determinism (W4), -002 Seed library content (W4), -003 `/moai project --refresh` (W4), -004 meta-harness 7-Phase (W4), -005 Project-specific generation (W4), -006 Migration tooling, -007 Backward migration tool (non-goal), -008 LLM-based capture (deferred), -009 Cross-project sharing (deferred), -010 Real-time canary streaming (out of scope).

---

## Recommended Next Action

Invoke **plan-auditor subagent** for Phase 2.3 independent review of this SPEC's plan-phase artifacts (4 files). Anticipated audit dimensions per W1 iter1 pattern: D1 Brief Quality, D2 Phase Decomposition, D3 Risk Management, D4 Frontmatter Compliance, D5 Exclusion Discipline, D6 Lint Baseline. Target: PASS ≥ 0.85 (W1 iter2 achieved 0.96).
