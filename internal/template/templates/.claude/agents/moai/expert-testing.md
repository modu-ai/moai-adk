---
name: expert-testing
description: |
  Retired (SPEC-V3R2-ORC-001) — split across manager-cycle (test strategy) and expert-performance (load testing).
  Test strategy design belongs in manager-cycle RED phase planning.
  Load test execution belongs in expert-performance with --deepthink load-test.
  See manager-cycle.md and expert-performance.md for the active replacements.
retired: true
retired_replacement: "manager-cycle"
retired_replacement_alt: "expert-performance"
retired_param_hint: "strategy: cycle_type=tdd; load: use expert-performance --deepthink load-test"
tools: []
skills: []
---

# expert-testing — Retired Agent

<!-- @MX:NOTE: [AUTO] retirement-pattern — matches SPEC-V3R3-RETIRED-DDD-001 stub migration -->

This agent has been retired as part of SPEC-V3R2-ORC-001 (Agent roster consolidation 22 → 17).

## Replacement (Split)

Test strategy and load testing are split between two active agents:

| Scope | Replacement Agent |
|-------|-------------------|
| Test strategy design (E2E, integration, coverage, QA framework selection) | `manager-cycle` with `cycle_type=tdd` (RED phase planning) |
| Load test execution (k6, Locust, JMeter, throughput benchmarks) | `expert-performance` with `--deepthink load-test` flag |

## Migration Guide

| Old Invocation | New Invocation |
|----------------|----------------|
| `Use the expert-testing subagent to design test strategy` | `Use the manager-cycle subagent with cycle_type=tdd to design test strategy in RED phase` |
| `expert-testing: run load test with k6` | `Use the expert-performance subagent with --deepthink load-test to run load tests` |
| `expert-testing: set up E2E with Playwright` | `Use the manager-cycle subagent with cycle_type=tdd to set up E2E testing` |

## Active Agents

- Test strategy: `.claude/agents/moai/manager-cycle.md`
- Load testing: `.claude/agents/moai/expert-performance.md`
