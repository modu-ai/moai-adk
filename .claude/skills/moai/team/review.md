---
description: >
  Multi-perspective code review using parallel review teammates.
  Four reviewers analyze security, performance, quality, and UX simultaneously.
  Findings are consolidated and deduplicated into a prioritized report.
  Use when comprehensive multi-angle code review is needed.
user-invocable: false
metadata:
  version: "2.5.0"
  category: "workflow"
  status: "active"
  updated: "2026-02-21"
  tags: "review, team, security, performance, quality, ux, parallel"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 3000

# MoAI Extension: Triggers
triggers:
  keywords: ["team review", "multi-perspective review", "parallel review"]
  agents: ["general-purpose"]
  phases: ["review"]
---
# Workflow: Team Review - Multi-Perspective Code Review

Purpose: Review code changes from multiple perspectives simultaneously. Each reviewer focuses on a specific quality dimension.

Flow: Spawn review teammates (implicit team) -> Perspective Assignment -> Parallel Review -> Report Consolidation

## Prerequisites

- workflow.team.enabled: true
- CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1
- Triggered by: /moai review --team OR explicit multi-perspective review request

## Phase 0: Review Setup

1. Identify the code changes to review (diff, PR, or file list)
2. The team forms implicitly on the first teammate spawn (one team per session, no setup step). Teams/tasks are stored under the session-derived name `session-<first8>`.
3. Create review tasks:
   ```
   TaskCreate: "Security review: OWASP compliance, input validation, auth" (no deps)
   TaskCreate: "Performance review: algorithmic complexity, resource usage, caching" (no deps)
   TaskCreate: "Quality review: TRUST 5, patterns, maintainability, test coverage" (no deps)
   TaskCreate: "UX review: user flow validation, error states, edge cases, accessibility" (no deps)
   TaskCreate: "Consolidate review findings" (blocked by above)
   ```

## Phase 1: Spawn Review Team

Use the review team pattern. All spawns MUST use Agent() with the `name` parameter — the team forms implicitly on the first spawn (no setup step; the `team_name` parameter is accepted but ignored as of Claude Code v2.1.178). Launch all four in a single response for parallel execution:

Each spawn prompt MUST include the following finding-stage instruction so every reviewer surfaces complete coverage: "At the finding stage, report every issue you find, including ones you are uncertain about or consider low-severity, each with a confidence level and an estimated severity. Do not filter for importance or confidence while finding — the verdict stage (must-pass thresholds + harmonic scoring) does the filtering downstream. The goal at this stage is coverage: surfacing a finding that later gets filtered out is preferable to silently dropping a real bug."

```
Agent(
  subagent_type: "general-purpose",
  name: "security-reviewer",
  mode: "plan",
  prompt: "You are a security reviewer on team moai-review-{target}.
    Review the following changes for security issues.
    Check OWASP Top 10 compliance, input validation, authentication/authorization,
    secrets exposure, injection risks.
    Changes: {diff_summary}
    When done, mark your task as completed via TaskUpdate and send findings to the team lead via SendMessage."
)

Agent(
  subagent_type: "general-purpose",
  name: "perf-reviewer",
  mode: "plan",
  prompt: "You are a performance reviewer on team moai-review-{target}.
    Review the following changes for performance issues.
    Check algorithmic complexity, database query efficiency, memory usage,
    caching opportunities, bundle size impact.
    Changes: {diff_summary}
    When done, mark your task as completed via TaskUpdate and send findings to the team lead via SendMessage."
)

Agent(
  subagent_type: "general-purpose",
  name: "quality-reviewer",
  mode: "plan",
  prompt: "You are a code quality reviewer on team moai-review-{target}.
    Review the following changes for code quality.
    Check TRUST 5 compliance, naming conventions, error handling,
    test coverage, documentation, consistency with project patterns.
    Changes: {diff_summary}
    When done, mark your task as completed via TaskUpdate and send findings to the team lead via SendMessage."
)

Agent(
  subagent_type: "general-purpose",
  name: "ux-reviewer",
  mode: "plan",
  prompt: "You are a UX reviewer on team moai-review-{target}.
    Review the following changes for user experience impact.
    Validate user flows remain functional, check error states and edge cases
    from the user's perspective, verify accessibility compliance,
    assess whether the changes align with expected user behavior.
    Changes: {diff_summary}
    When done, mark your task as completed via TaskUpdate and send findings to the team lead via SendMessage."
)
```

All four reviewers run in parallel. Messages from teammates are delivered automatically to MoAI.

## Phase 2: Parallel Review

Reviewers work independently (all read-only, 4 perspectives):
- Security: OWASP compliance, injection risks, auth vulnerabilities
- Performance: Algorithmic complexity, resource usage, caching
- Quality: TRUST 5 compliance, patterns, maintainability
- UX: User flow integrity, error states, accessibility
- Each rates findings by severity (critical, warning, suggestion)
- Reports findings to team lead

## Phase 3: Report Consolidation

After all reviews complete:
1. Collect findings from all reviewers
2. Deduplicate overlapping issues
3. Prioritize by severity (critical first)
4. Present consolidated review report to user with:
   - Critical issues requiring immediate attention
   - Warnings that should be addressed
   - Suggestions for improvement
   - Overall quality assessment per TRUST 5 dimension
   - User experience impact assessment

## Phase 4: Cleanup

1. Shutdown all review teammates:
   ```
   SendMessage(type: "shutdown_request", recipient: "security-reviewer", content: "Review complete, shutting down")
   SendMessage(type: "shutdown_request", recipient: "perf-reviewer", content: "Review complete, shutting down")
   SendMessage(type: "shutdown_request", recipient: "quality-reviewer", content: "Review complete, shutting down")
   SendMessage(type: "shutdown_request", recipient: "ux-reviewer", content: "Review complete, shutting down")
   ```
2. Clean up GLM env vars and restore Claude-only operation:
   ```bash
   moai cc
   ```
   This safely removes GLM env vars while preserving ANTHROPIC_AUTH_TOKEN and other settings.
   Do NOT manually Read/Write settings.local.json — use the CLI command which handles JSON merging correctly.
3. Team cleanup is automatic on session exit (no explicit teardown call — the TeamDelete tool was removed in Claude Code v2.1.178)
4. Optionally create fix tasks for critical issues

## Fallback

If team creation fails:
- Fall back to sync-auditor subagent for single-perspective review
- Sequential review of security, performance, then quality

---

Version: 1.1.0
Updated: 2026-02-13
Source: Added ux-reviewer as 4th review perspective for user flow validation, error states, and accessibility.
