---
spec_id: SPEC-SEMAP-001
plan_version: "0.1.0"
plan_date: 2026-05-14
plan_author: manager-spec (auto)
plan_status: draft
---

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-14 | manager-spec | Initial plan |

## 1. Plan Overview

SPEC-SEMAP-001 introduces behavioral contracts (preconditions, postconditions, invariants, forbidden) for MoAI agents. Phase 1 targets manager-ddd (now retired/consolidated into manager-develop) and manager-quality. The codebase has evolved since this SPEC was written: manager-ddd and manager-tdd have been retired into manager-develop (SPEC-V3R3-RETIRED-DDD-001).

Key consideration: Since the original Phase 1 targets (manager-ddd, manager-quality) have changed, the plan must be updated to target the current agent roster: **manager-develop** (the consolidated implementation agent) and **manager-quality**.

## 2. Gap Analysis

### Current State

| REQ | Status | Evidence |
|-----|--------|----------|
| REQ-001 | NOT DONE | No agent has a `## Contract` section |
| REQ-002 | NOT DONE | No contract verification at phase completion points |
| REQ-003 | NOT DONE | No phased rollout enforcement level in config |

### Agent Roster Change

- manager-ddd -> RETIRED, replaced by manager-develop (cycle_type=ddd)
- manager-tdd -> RETIRED, replaced by manager-develop (cycle_type=tdd)
- Phase 1 targets must be updated: manager-develop + manager-quality

## 3. Milestone Breakdown

### M1 -- Define Contract Schema and Template -- Priority P0

Define the Markdown-based contract section format for agent definitions:
- `## Contract` heading in agent body
- `preconditions`: List of conditions before agent executes
- `postconditions`: List of conditions after agent completes
- `invariants`: Conditions that hold throughout execution
- `forbidden`: Hard prohibitions
- Each condition is a natural language assertion verifiable by inspection

Files:
- No file changes needed (schema definition is documentation-first)

### M2 -- Add Contract Sections to Phase 1 Agents -- Priority P0

Add `## Contract` sections to the two Phase 1 agents:
- **manager-develop.md**: Preconditions (SPEC loaded, worktree active), Postconditions (all tests pass, coverage >= 85%), Invariants (existing tests pass), Forbidden (delete test files)
- **manager-quality.md**: Preconditions (implementation files exist), Postconditions (LSP errors = 0, TRUST 5 pass), Invariants (no code modified), Forbidden (write code, apply fixes)

Files:
- `internal/template/templates/.claude/agents/moai/manager-develop.md` (add ## Contract section)
- `internal/template/templates/.claude/agents/moai/manager-quality.md` (add ## Contract section)

### M3 -- Contract Verification at Phase Completion -- Priority P1

Add contract verification step to spec-workflow.md at phase completion points:
- After manager-develop completes (Phase 2 implementation): verify postconditions
- After manager-quality completes (Phase 2.8b): verify postconditions
- Generate violation report as structured Markdown appended to progress.md
- Phase 1 enforcement: violations are warnings (non-blocking)

Files:
- `internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md` (add contract verification step)

### M4 -- Configure Phase 1 Enforcement Level -- Priority P1

Configure warning-only enforcement for Phase 1:
- Add `contract_enforcement: warning` to quality.yaml or harness.yaml
- Contract violations produce warnings, not errors
- Violation data collected for threshold tuning
- Phase 2 escalation to blocking enforcement requires separate SPEC amendment

Files:
- `internal/template/templates/.moai/config/sections/quality.yaml` (add contract_enforcement field)

### M5 -- Backward Compatibility Verification -- Priority P2

Verify backward compatibility:
- Agents without `## Contract` sections behave exactly as before
- No contract verification runs for agents without contracts
- No performance impact on agents without contracts
- Existing tests pass without modification

Files:
- No file changes (verification only)

## 4. File:line Anchors

| File | Line(s) | Action | Purpose |
|------|---------|--------|---------|
| `internal/template/templates/.claude/agents/moai/manager-develop.md` | body (end) | Edit | Add ## Contract section |
| `internal/template/templates/.claude/agents/moai/manager-quality.md` | body (end) | Edit | Add ## Contract section |
| `internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md` | Phase completion | Edit | Add contract verification step |
| `internal/template/templates/.moai/config/sections/quality.yaml` | end | Edit | Add contract_enforcement field |

## 5. Quality Gates

- Contract section parseable (4 fields: preconditions, postconditions, invariants, forbidden)
- Phase 1 agents (manager-develop, manager-quality) have contracts
- Contract violations in Phase 1 produce warnings only (non-blocking)
- Agents without contracts are unaffected (backward compatible)
- No contract auto-generation (contracts are human-authored per exclusion)

## 6. Dependencies

- SPEC-V3R3-RETIRED-DDD-001: manager-ddd retired into manager-develop (changes Phase 1 targets)
- SPEC-EVAL-001: evaluator-active performs active evaluation (contract verification complements)
- SPEC-DRIFT-001: Drift guard checks file scope (contract postconditions check behavior scope)
