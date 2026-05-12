# spec-compact — SPEC-V3R2-HRN-002

One-page reference card. For full text see `spec.md`, `research.md`, `plan.md`, `acceptance.md`, `tasks.md`.

## Identity

| Field | Value |
|-------|-------|
| ID | SPEC-V3R2-HRN-002 |
| Title | Evaluator Memory Scope Amendment (per-iteration fresh judgment) |
| Phase | v3.0.0 — Phase 5 — Harness + Evaluator |
| Priority | P0 Critical |
| Status (spec.md) | draft |
| Status (plan.md) | audit-ready |
| Breaking | true (BC-V3R2-010) |
| Lifecycle | spec-anchored |

## Goal in 3 lines

Insert FROZEN-zone amendment §11.4.1 into `.claude/rules/moai/design/constitution.md` declaring per-iteration ephemeral evaluator judgment memory + durable Sprint Contract state, per R1 §9 Agent-as-a-Judge anti-pattern (Zhuge et al. 2024). Wire `evaluator.memory_scope: per_iteration` (FROZEN value) into both design.yaml and harness.yaml with loader validator rejecting any other value via `HRN_EVAL_MEMORY_FROZEN`. Land via CON-002 5-layer graduation protocol mirroring SPC-001 PR #870 pattern.

## REQ Index (19 total)

| Modality | REQ IDs |
|----------|---------|
| Ubiquitous (5.1) | 001, 002, 003, 004, 005, 006, 007 |
| Event-Driven (5.2) | 008, 009, 010, 011 |
| State-Driven (5.3) | 012, 013, 014 |
| Optional (5.4) | 015, 016 |
| Unwanted (5.5) | 017, 018, 019 |

## AC Index (13 total; 3 self-demonstrate hierarchy with depth-2 nesting on AC-07)

| AC ID | Hierarchical? | Mapped REQs |
|-------|---------------|-------------|
| AC-HRN-002-01 | YES (3 children .a/.b/.c) | 001 |
| AC-HRN-002-02 | flat | 002 |
| AC-HRN-002-03 | flat | 003, 004 |
| AC-HRN-002-04 | YES (3 children .a/.b/.c) | 005, 011, 014, 018 |
| AC-HRN-002-05 | flat | 006 |
| AC-HRN-002-06 | flat | 007 |
| AC-HRN-002-07 | YES (3 children + 2 grandchildren on .a) | 009, 013, 017 |
| AC-HRN-002-08 | flat | 012 |
| AC-HRN-002-09 | flat | 008, 010, 015 |
| AC-HRN-002-10 | flat | 019 |
| AC-HRN-002-11 | flat | 014 (BC-V3R2-010 session upgrade) |
| AC-HRN-002-12 | flat | 010 (dual evolution-log) |
| AC-HRN-002-13 | flat | 016 (solo-mode alias) |

## Files Touched in Run Phase

| File | Purpose | M |
|------|---------|---|
| `.claude/rules/moai/design/constitution.md` | §11.4.1 + version 3.4.0→3.5.0 + HISTORY | M2 |
| `internal/config/types.go` + `loader.go` + `errors.go` | EvaluatorConfig.MemoryScope + HRN_EVAL_MEMORY_FROZEN | M2 |
| `internal/config/loader_test.go` + 3 fixtures | AC-HRN-002-04 leaf tests | M4 |
| `.claude/agents/moai/evaluator-active.md` | Cross-reference line | M3 |
| `.claude/skills/moai-workflow-gan-loop/SKILL.md` | Phase 4b + Solo Mode subsection | M3 |
| `internal/harness/evaluator_leak.go` + `_test.go` | Leak detection (REQ-017) | M3 |
| `.moai/config/sections/{design,harness}.yaml` (+ template mirrors) | evaluator.memory_scope: per_iteration | M4 |
| `.moai/specs/SPEC-V3R2-HRN-002/canary-fresh-memory-eval.txt` + `con-002-amendment-evidence.md` | Evidence | M5 |
| `.moai/research/evolution-log.md` + `.moai/design/v3-research/evolution-log.md` | Dual EVO-HRN-002 | M5 |
| `.claude/rules/moai/core/zone-registry.md` + `progress.md` | CONST entry + status | M5 |

## Adjacent SPECs

| Type | SPEC | Relationship |
|------|------|--------------|
| Blocked by | SPEC-V3R2-CON-001 | FROZEN zone model + zone-registry infrastructure |
| Gates | SPEC-V3R2-CON-002 | 5-layer amendment safety gate (mirrors SPC-001 PR #870) |
| Blocked by | SPEC-V3R2-HRN-001 | HarnessConfig struct extended here |
| Blocks | SPEC-V3R2-HRN-003 | Hierarchical per-leaf scoring needs fresh-judgment semantics |
| Blocks | SPEC-V3R2-WF-003 | Thorough harness mode uses Sprint Contract per this SPEC |
| Pattern source | SPEC-V3R2-SPC-001 | PR #870 — CON-002 paperwork + hierarchical AC schema reused verbatim |

## Risks (top 4)

| Risk | Severity | Mitigation milestone |
|------|----------|-----------------------|
| Human reviewer rubber-stamps without reading R1 §9 evidence | HIGH | M5 — bundle 4-component evidence in AskUser Option 1 |
| Subtle prior-context leak via Sprint Contract serialization | HIGH | M3 — leak detection test (REQ-017) |
| <3 design projects in Canary corpus | MEDIUM | M5 — fall back to v2-legacy or CanaryUnavailable |
| Dual evolution-log path disambiguation | MEDIUM | M5 — Decision D5 dual-write codified |

End of compact.
