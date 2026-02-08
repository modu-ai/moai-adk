# Workflow: Team Review - Multi-Perspective Code Review

Purpose: Review code changes from multiple perspectives simultaneously. Each reviewer focuses on a specific quality dimension.

Flow: TeamCreate -> Perspective Assignment -> Parallel Review -> Report Consolidation

## Prerequisites

- workflow.team.enabled: true
- CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1
- Triggered by: /moai review --team OR explicit multi-perspective review request

## Phase 0: Review Setup

1. Identify the code changes to review (diff, PR, or file list)
2. Create team:
   ```
   TeamCreate(team_name: "moai-review-{target}")
   ```
3. Create review tasks:
   ```
   TaskCreate: "Security review: OWASP compliance, input validation, auth" (no deps)
   TaskCreate: "Performance review: algorithmic complexity, resource usage, caching" (no deps)
   TaskCreate: "Quality review: TRUST 5, patterns, maintainability, test coverage" (no deps)
   TaskCreate: "Consolidate review findings" (blocked by above)
   ```

## Phase 1: Spawn Review Team

Use the review team pattern:

Teammate 1 - security-reviewer (team-quality agent, inherit model):
- Prompt: "Review the following changes for security issues. Check OWASP Top 10 compliance, input validation, authentication/authorization, secrets exposure, injection risks. Changes: {diff_summary}"

Teammate 2 - perf-reviewer (team-quality agent, inherit model):
- Prompt: "Review the following changes for performance issues. Check algorithmic complexity, database query efficiency, memory usage, caching opportunities, bundle size impact. Changes: {diff_summary}"

Teammate 3 - quality-reviewer (team-quality agent, inherit model):
- Prompt: "Review the following changes for code quality. Check TRUST 5 compliance, naming conventions, error handling, test coverage, documentation, consistency with project patterns. Changes: {diff_summary}"

## Phase 2: Parallel Review

Reviewers work independently (all read-only):
- Each focuses on their assigned quality dimension
- Reviews all changed files from their perspective
- Rates each finding by severity (critical, warning, suggestion)
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

## Phase 4: Cleanup

1. Shutdown all review teammates
2. TeamDelete to clean up resources
3. Optionally create fix tasks for critical issues

## Fallback

If team creation fails:
- Fall back to manager-quality subagent for single-perspective review
- Sequential review of security, performance, then quality

---

Version: 1.0.0
