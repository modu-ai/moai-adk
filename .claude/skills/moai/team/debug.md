---
name: moai-workflow-team-debug
description: >
  Debug complex issues through parallel competing hypothesis investigation.
  Each teammate explores a different theory independently using haiku model.
  Evidence is synthesized to identify root cause before fix implementation.
  Use when debugging issues with multiple potential root causes.
user-invocable: false
metadata:
  version: "2.5.0"
  category: "workflow"
  status: "active"
  updated: "2026-02-21"
  tags: "debug, team, hypothesis, investigation, parallel"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 3000

# MoAI Extension: Triggers
triggers:
  keywords: ["debug team", "hypothesis", "investigation", "parallel debug"]
  agents: ["expert-debug"]
  phases: ["fix"]
---
# Workflow: Team Debug - Investigation Team

Purpose: Debug complex issues through parallel competing hypothesis investigation. Each teammate explores a different theory independently.

Flow: TeamCreate -> Hypothesis Assignment -> Parallel Investigation -> Evidence Synthesis -> Fix

## Prerequisites

- workflow.team.enabled: true
- CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1
- Triggered by: /moai fix --team OR auto-detected when issue has multiple potential root causes

## Phase 0: Issue Analysis

1. Analyze the error/issue to identify potential root causes
2. Formulate 2-3 competing hypotheses
3. Create team:
   ```
   TeamCreate(team_name: "moai-debug-{issue-slug}")
   ```
4. Create investigation tasks:
   ```
   TaskCreate: "Investigate hypothesis 1: {description}" (no deps)
   TaskCreate: "Investigate hypothesis 2: {description}" (no deps)
   TaskCreate: "Investigate hypothesis 3: {description}" (no deps)
   TaskCreate: "Synthesize findings and implement fix" (blocked by above)
   ```

## Phase 1: Spawn Investigation Team

Use the investigation team pattern:

Teammate 1 - hypothesis-1 (team-researcher agent, haiku model):
- Prompt: "Investigate whether the issue is caused by {hypothesis_1}. Look for evidence supporting or contradicting this theory. Report your findings with confidence level."

Teammate 2 - hypothesis-2 (team-researcher agent, haiku model):
- Prompt: "Investigate whether the issue is caused by {hypothesis_2}. Look for evidence supporting or contradicting this theory. Report your findings with confidence level."

Teammate 3 - hypothesis-3 (team-researcher agent, haiku model):
- Prompt: "Investigate whether the issue is caused by {hypothesis_3}. Look for evidence supporting or contradicting this theory. Report your findings with confidence level."

## Phase 2: Parallel Investigation

Teammates work independently (all haiku, fast and cheap):
- Each explores their hypothesis
- Searches codebase for evidence
- Checks logs, tests, configuration
- Reports findings with confidence level (high/medium/low)

MoAI monitors:
- Receive findings as teammates complete
- If one hypothesis gets high confidence early, may redirect others

## Phase 3: Evidence Synthesis

After all investigations complete:
1. Compare evidence across hypotheses
2. Identify the most likely root cause
3. Delegate fix to expert-debug subagent (NOT a teammate) with:
   - All investigation findings
   - Identified root cause
   - Reproduction steps from evidence
4. Follow reproduction-first bug fix protocol (write failing test first)

## Phase 4: Cleanup

1. Shutdown all investigation teammates
2. Clean up GLM env vars and restore Claude-only operation:
   ```bash
   moai cc
   ```
   This safely removes GLM env vars while preserving ANTHROPIC_AUTH_TOKEN and other settings.
   Do NOT manually Read/Write settings.local.json — use the CLI command which handles JSON merging correctly.
3. TeamDelete to clean up resources
4. Report diagnosis and fix to user

## Fallback

If team creation fails:
- Fall back to sub-agent fix workflow (workflows/fix.md)
- Use sequential hypothesis investigation instead

---

Version: 1.0.0
