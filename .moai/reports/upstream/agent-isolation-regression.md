---
title: "Agent(isolation: \"worktree\") regression — upstream bug report draft"
status: placeholder
created_at: 2026-05-08
spec: SPEC-V3R3-CI-AUTONOMY-001
wave: 5
task: W5-T07
agent_responsible: claude-code-guide
---

# Agent(isolation: "worktree") Regression — Upstream Bug Report

> **Status**: Placeholder. The `claude-code-guide` agent will populate this file
> when the orchestrator first detects an `Agent(isolation: "worktree")` regression
> in this project. Until then, this file serves as a contract that the report
> location is reserved.

_TBD_

## Sections (to be filled by claude-code-guide)

- Environment (claude --version, OS, git version, MoAI version)
- Reproduction Steps
- Expected Behavior
- Actual Behavior (with snapshot ID + log path)
- Workarounds
- Frequency (per-session / sporadic / always)
- MoAI Mitigation (pointer to `moai worktree snapshot|verify|restore`)

## How to file upstream

When this report is populated and reviewed:

1. Visit https://github.com/anthropics/claude-code/issues
2. Click "New issue" → "Bug report"
3. Paste this file's content (excluding this `## How to file upstream` section)
4. Add labels: `bug`, `agent`, `worktree-isolation`
5. Submit

The MoAI orchestrator does NOT file this report automatically — file submission
is a deliberate user action.

## Cross-references

- SPEC: `.moai/specs/SPEC-V3R3-CI-AUTONOMY-001/spec.md` § 3.6 (REQ-CIAUT-035)
- Rule: `.claude/rules/moai/workflow/worktree-state-guard.md`
- Agent: `.claude/agents/moai/claude-code-guide.md`
