---
spec_id: SPEC-SEMAP-001
title: "SEMAP Behavioral Contracts"
created: "2026-04-01"
status: approved
priority: high
module: template
version: "1.0.0"
lifecycle: spec-anchored
---

# SEMAP Behavioral Contracts

## Overview

MoAI 에이전트에 행동 계약(behavioral contract) 시스템을 도입한다. 각 에이전트의 사전 조건(preconditions), 사후 조건(postconditions), 불변 조건(invariants), 금지 행위(forbidden)를 명시적으로 정의하여 에이전트 실행의 예측 가능성과 신뢰성을 높인다. Phase 1에서 manager-ddd와 manager-quality에 먼저 적용하고, 검증 후 전체 매니저/전문가 에이전트로 확대한다.

## Requirements (EARS Format)

### REQ-001: Agent Contract Schema

**When** an agent definition includes a `## Contract` section in its body, the system **shall** recognize the following contract fields:

- `preconditions`: List of conditions that MUST be true before the agent executes (input validation)
- `postconditions`: List of conditions that MUST be true after the agent completes (output guarantees)
- `invariants`: List of conditions that MUST remain true throughout agent execution (behavioral constraints)
- `forbidden`: List of actions the agent MUST NEVER perform (hard prohibitions)

Each condition **shall** be a natural language assertion that can be verified by inspection.

**Acceptance**:
- **Given** an agent definition with a `## Contract` section containing all 4 fields
- **When** the agent definition is parsed
- **Then** preconditions, postconditions, invariants, and forbidden lists are recognized
- **And** each field contains at least one verifiable assertion
- **Given** an agent definition without a `## Contract` section
- **When** the agent definition is parsed
- **Then** no contract verification is performed (backward compatible)

### REQ-002: Contract Verification

**When** a phase completes and the executing agent has a defined contract, the system **shall**:
- Verify all postconditions are satisfied
- Check that no forbidden actions were performed (via output inspection)
- If any postcondition fails, generate a contract violation report containing:
  - Agent name
  - Phase identifier
  - Failed postcondition(s)
  - Actual output vs expected condition
  - Severity level (warning in Phase 1, error in Phase 2+)

**Acceptance**:
- **Given** manager-ddd completes Phase 2A with all postconditions satisfied
- **When** contract verification runs
- **Then** no violation report is generated
- **And** the phase proceeds to the next step
- **Given** manager-ddd completes Phase 2A but "test coverage >= 85%" postcondition fails
- **When** contract verification runs
- **Then** a contract violation report is generated
- **And** the report includes the failed postcondition and actual coverage value
- **And** in Phase 1 rollout, the violation is logged as warning (non-blocking)

### REQ-003: Phased Rollout

**When** SEMAP contracts are deployed, the system **shall** follow a 3-phase rollout:

- **Phase 1**: Apply contracts to manager-ddd and manager-quality only
  - Contract violations produce warnings (non-blocking)
  - Collect violation data for 30 days to tune thresholds
- **Phase 2**: Extend contracts to all manager-* agents
  - Contract violations produce errors (blocking for critical postconditions)
  - Warning-only for non-critical postconditions
- **Phase 3**: Extend contracts to all expert-* agents
  - Full contract enforcement across the agent catalog

Each phase transition requires human approval based on violation data analysis.

**Acceptance**:
- **Given** SEMAP is deployed at Phase 1
- **When** a violation occurs in manager-ddd
- **Then** the violation is logged as a warning (implementation continues)
- **And** violation data is collected for threshold tuning
- **Given** SEMAP is at Phase 2
- **When** a critical postcondition violation occurs in manager-quality
- **Then** the violation is logged as an error
- **And** the phase is blocked until the violation is resolved
- **Given** SEMAP is at Phase 1
- **When** expert-backend violates a hypothetical contract
- **Then** no contract verification occurs (expert-* not yet enrolled)

## Exclusions (What NOT to Build)

- Shall NOT modify Claude Code's native agent system (contracts are MoAI-layer only)
- Shall NOT block implementation on contract violations in Phase 1 (warning only)
- Shall NOT auto-generate contracts from agent behavior (contracts are human-authored)
- Shall NOT apply contracts to builder-* agents (agent creation agents are exempt)
- Shall NOT enforce contracts on Explore subagent (read-only agents are exempt)

## Acceptance Criteria

**Scenario 1: manager-ddd Contract Verification (Phase 1)**
- Given manager-ddd has a contract with postcondition "all new functions have tests"
- When manager-ddd completes implementation with 2 untested functions
- Then a contract violation report is generated
- And the report lists the 2 untested function names
- And the violation severity is "warning" (Phase 1, non-blocking)
- And implementation proceeds to the next phase

**Scenario 2: manager-quality Contract Enforcement (Phase 1)**
- Given manager-quality has a contract with postcondition "LSP errors = 0"
- When manager-quality completes with 3 remaining LSP errors
- Then a contract violation report is generated
- And the report includes error count and file locations
- And the violation is logged as warning (Phase 1)

**Scenario 3: Backward Compatibility**
- Given an existing agent definition (e.g., expert-backend) with no Contract section
- When the agent is invoked and completes
- Then no contract verification occurs
- And the agent behaves exactly as before SEMAP introduction

**Scenario 4: Contract Section Structure**
- Given manager-ddd.md with a Contract section
- When the contract is inspected
- Then preconditions include "SPEC document loaded and parsed"
- And postconditions include "all modified files pass go vet"
- And invariants include "existing tests continue to pass"
- And forbidden includes "deleting test files"

## Technical Approach

1. Define contract section format as Markdown within agent definition body (## Contract heading)
2. Add contract sections to manager-ddd.md and manager-quality.md
3. Implement contract verification step in spec-workflow.md at phase completion points
4. Generate violation reports as structured Markdown appended to progress.md
5. Configure Phase 1 enforcement level (warning-only) in harness.yaml or quality.yaml

## Files to Modify

- `internal/template/templates/.claude/agents/moai/manager-ddd.md` (add ## Contract section)
- `internal/template/templates/.claude/agents/moai/manager-quality.md` (add ## Contract section)
- `internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md` (contract verification at phase completion)
