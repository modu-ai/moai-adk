---
spec_id: SPEC-EVAL-001
title: "evaluator-active Agent & Sprint Contract"
created: "2026-04-01"
status: planned
priority: high
module: template
version: "1.0.0"
lifecycle: spec-anchored
---

# evaluator-active Agent & Sprint Contract

## Overview

evaluator-active는 MoAI-ADK의 독립적 품질 평가 에이전트로, 기존 manager-quality를 보완한다. 회의론적(skeptical) 평가 관점을 도입하여 구현 결과물을 기능성, 보안, 장인정신, 일관성 4개 축으로 평가한다. Sprint Contract 협상 메커니즘을 통해 구현 전 계획 검증을 수행하고, Phase 2.8을 2.8a(능동 평가)와 2.8b(정적 검증)로 분리하여 품질 게이트를 강화한다.

## Requirements (EARS Format)

### REQ-001: evaluator-active Agent Definition

**When** the MoAI-ADK template is deployed, the system **shall** include an `evaluator-active.md` agent definition in `internal/template/templates/.claude/agents/moai/` with the following configuration:
- model: sonnet
- permissionMode: plan (read-only)
- Skeptical evaluator prompt that NEVER rationalizes acceptance and requires concrete evidence
- 4 evaluation dimensions with weights: Functionality(40%), Security(25%), Craft(20%), Consistency(15%)
- Security FAIL results in overall FAIL regardless of other dimension scores (hard threshold)

**Acceptance**:
- **Given** a deployed MoAI-ADK project
- **When** evaluator-active is invoked
- **Then** the agent operates in read-only mode (permissionMode: plan) and evaluates against all 4 dimensions
- **And** a Security FAIL verdict overrides all other passing dimensions to produce an overall FAIL

### REQ-002: Phase 2.0 Sprint Contract Negotiation

**When** harness level is thorough and the Run phase begins, the system **shall** insert a Phase 2.0 before Phase 2.1 where evaluator-active reviews manager-ddd's implementation plan before coding starts.
- The contract negotiation result is recorded in `.moai/specs/SPEC-{ID}/contract.md`
- Phase 2.0 is active only in thorough harness level

**Acceptance**:
- **Given** a SPEC in Run phase with harness level = thorough
- **When** Phase 2.0 executes
- **Then** evaluator-active reviews the implementation plan produced by manager-ddd
- **And** the negotiation result is persisted to `.moai/specs/SPEC-{ID}/contract.md`
- **And** Phase 2.1 does not start until contract is finalized

### REQ-003: Phase 2.8a/2.8b Split

**When** the Run phase reaches the quality gate, the system **shall** split Phase 2.8 into two sub-phases:
- Phase 2.8a: evaluator-active tests against acceptance criteria with a maximum of 3 retry cycles
- Phase 2.8b: manager-quality performs TRUST 5 static verification (existing behavior preserved)

**Acceptance**:
- **Given** a SPEC implementation completing the build phase
- **When** Phase 2.8a executes
- **Then** evaluator-active evaluates against SPEC acceptance criteria
- **And** if FAIL, the implementation agent retries (max 3 cycles)
- **And** after Phase 2.8a PASS (or max retries), Phase 2.8b executes existing manager-quality checks

### REQ-004: Mode-Specific Deployment

**When** the evaluator-active is invoked, the system **shall** deploy it according to the current execution mode:
- Sub-agent mode: Agent(evaluator-active) call
- Team mode: reviewer role teammate via SendMessage
- CG mode: Leader(Claude) performs evaluator role directly (no separate agent spawn)

**Acceptance**:
- **Given** MoAI running in sub-agent mode
- **When** evaluator-active is needed
- **Then** it is invoked via Agent(evaluator-active)
- **Given** MoAI running in team mode
- **When** evaluator-active is needed
- **Then** a reviewer role teammate receives the evaluation task via SendMessage
- **Given** MoAI running in CG mode
- **When** evaluator-active is needed
- **Then** the Leader (Claude) performs the evaluation directly without spawning a separate agent

### REQ-005: Evaluator Intervention Modes

**When** evaluator-active is configured, the system **shall** support two intervention modes:
- final-pass: Single evaluation at Phase 2.8a only (standard harness, Opus 4.6+ default)
- per-sprint: Phase 2.0 pre-negotiation + Phase 2.8a post-verification (thorough harness)

**Acceptance**:
- **Given** harness level = standard
- **When** Run phase executes
- **Then** evaluator-active runs only at Phase 2.8a (final-pass mode)
- **Given** harness level = thorough
- **When** Run phase executes
- **Then** evaluator-active runs at Phase 2.0 (contract negotiation) AND Phase 2.8a (post-verification)

## Exclusions (What NOT to Build)

- Shall NOT replace manager-quality (evaluator-active supplements, not replaces)
- Shall NOT modify existing test files (evaluator-active is read-only evaluation)
- Shall NOT be active in minimal harness level
- Shall NOT write code or apply fixes (evaluation only, never implementation)

## Acceptance Criteria

**Scenario 1: Standard Harness Evaluation**
- Given a SPEC with harness level = standard
- When the Run phase reaches Phase 2.8a
- Then evaluator-active evaluates the implementation against acceptance criteria
- And produces a PASS/FAIL verdict with per-dimension scores
- And Phase 2.8b (manager-quality) runs after 2.8a completes

**Scenario 2: Thorough Harness Full Cycle**
- Given a SPEC with harness level = thorough
- When the Run phase starts
- Then Phase 2.0 sprint contract negotiation executes before Phase 2.1
- And the contract is recorded in contract.md
- And Phase 2.8a evaluates the final implementation
- And a maximum of 3 retry cycles is enforced on failure

**Scenario 3: Security Hard Threshold**
- Given an implementation with a security vulnerability
- When evaluator-active evaluates the implementation
- Then the Security dimension returns FAIL
- And the overall verdict is FAIL regardless of other dimension scores

## Technical Approach

1. Create `evaluator-active.md` agent definition with skeptical prompt and scoring rubric
2. Add Phase 2.0 section to `run.md` skill, gated by harness level check
3. Split existing Phase 2.8 into 2.8a (evaluator-active) and 2.8b (manager-quality)
4. Implement mode-specific deployment logic in run.md routing
5. Define intervention mode selection based on harness.yaml configuration

## Files to Modify

- `internal/template/templates/.claude/agents/moai/evaluator-active.md` (NEW)
- `internal/template/templates/.claude/skills/moai/workflows/run.md` (Phase 2.0, 2.8a/b split)
