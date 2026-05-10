---
name: expert-debug
description: |
  Retired (SPEC-V3R2-ORC-001) — use manager-quality with diagnostic-mode.
  Debugging diagnosis has been folded into manager-quality as a Diagnostic Sub-Mode.
  See manager-quality.md § Diagnostic Sub-Mode for the active replacement.
retired: true
retired_replacement: manager-quality
retired_param_hint: "diagnostic-mode"
tools: []
skills: []
---

# expert-debug — Retired Agent

<!-- @MX:NOTE: [AUTO] retirement-pattern — matches SPEC-V3R3-RETIRED-DDD-001 stub migration -->

This agent has been retired as part of SPEC-V3R2-ORC-001 (Agent roster consolidation 22 → 17).

## Replacement

Use **manager-quality** in diagnostic mode instead. The Diagnostic Sub-Mode absorbs
expert-debug's routing logic and CI failure interpretation capability.

## Migration Guide

| Old Invocation | New Invocation |
|----------------|----------------|
| `Use the expert-debug subagent to diagnose this error` | `Use the manager-quality subagent in diagnostic-mode to diagnose this error` |
| `expert-debug: find root cause of test failure` | `manager-quality: run Diagnostic Sub-Mode — find root cause of test failure` |

## Active Agent

See `.claude/agents/moai/manager-quality.md` § Diagnostic Sub-Mode for the full definition.
