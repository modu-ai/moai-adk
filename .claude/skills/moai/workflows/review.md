---
name: moai-workflow-review
description: >
  Multi-perspective code review with security, performance, quality, and UX analysis.
  Supports staged changes, branch comparison, and security-focused review.
  Team mode available for parallel multi-perspective review.
  Use when performing code review, security audit, or quality assessment.
user-invocable: false
metadata:
  version: "2.5.0"
  category: "workflow"
  status: "active"
  updated: "2026-02-21"
  tags: "review, code-review, security, performance, quality, ux, audit"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["review", "code review", "security audit", "quality check", "code analysis"]
  agents: ["manager-quality", "expert-security"]
  phases: ["review"]
---

# Workflow: Review - Code Review

Purpose: Multi-perspective code review analyzing security, performance, quality, and UX dimensions. Produces a consolidated, prioritized report of findings.

Flow: Identify Changes -> Analyze Perspectives -> Consolidate -> Report

## Supported Flags

- --staged: Review only staged (git add) changes
- --branch BRANCH: Compare current branch against BRANCH (default: main)
- --security: Focus primarily on security review (OWASP, injection, auth)
- --file PATH: Review specific file(s) only
- --team: Use parallel multi-perspective review team (see team/review.md)

## Phase 1: Identify Changes

Determine the scope of code to review:

If --staged: Use `git diff --staged` to get staged changes.
If --branch: Use `git diff {BRANCH}...HEAD` to get branch changes.
If --file: Read the specified file(s) directly.
If no flag: Use `git diff HEAD~1` for the most recent commit changes.

Collect:
- List of modified files with change types (added, modified, deleted)
- Diff summary with line counts
- Affected modules and their responsibilities

## Phase 2: Multi-Perspective Analysis

[HARD] Delegate review to the manager-quality subagent with all perspectives.

If --team flag: Route to team/review.md for parallel multi-perspective review with 4 dedicated reviewers.

If no --team flag (default single-agent mode): Delegate to manager-quality subagent with instructions to review from all 4 perspectives sequentially.

### Perspective 1: Security Review

- OWASP Top 10 compliance check
- Input validation and sanitization
- Authentication and authorization logic
- Secrets exposure (API keys, passwords, tokens)
- Injection risks (SQL, command, XSS, CSRF)
- Dependency vulnerability check

If --security flag: This perspective receives primary focus with deeper analysis.

### Perspective 2: Performance Review

- Algorithmic complexity analysis (O(n) considerations)
- Database query efficiency (N+1 queries, missing indexes)
- Memory usage patterns (leaks, excessive allocation)
- Caching opportunities
- Bundle size impact (frontend changes)
- Concurrency safety (race conditions, deadlocks)

### Perspective 3: Quality Review

- TRUST 5 compliance (Tested, Readable, Unified, Secured, Trackable)
- Naming conventions and code readability
- Error handling completeness
- Test coverage for changed code
- Documentation for public APIs
- Consistency with project patterns and conventions

### Perspective 4: UX Review

- User flow integrity (do changes break existing flows?)
- Error states and edge cases from user perspective
- Accessibility compliance (WCAG, ARIA)
- Loading states and feedback mechanisms
- Breaking changes in public interfaces

## Phase 3: MX Tag Compliance Check

After perspective analysis, check @MX tag compliance for changed files:

- New exported functions: Should have @MX:NOTE or @MX:ANCHOR
- High fan_in functions (>=3 callers): Must have @MX:ANCHOR
- Dangerous patterns: Should have @MX:WARN
- Untested public functions: Should have @MX:TODO

Report missing or outdated @MX tags as findings.

## Phase 4: Report Consolidation

Produce a consolidated review report organized by severity:

### Report Structure

```markdown
## Code Review Report - {target}

### Critical Issues (must fix)
- [SECURITY] file:line: Description
- [PERFORMANCE] file:line: Description

### Warnings (should fix)
- [QUALITY] file:line: Description
- [UX] file:line: Description

### Suggestions (nice to have)
- [QUALITY] file:line: Description

### MX Tag Compliance
- Missing tags: N
- Outdated tags: N
- Compliant files: N/M

### Overall Assessment
- Security: PASS/FAIL
- Performance: PASS/WARN
- Quality: PASS/WARN
- UX: PASS/WARN
- TRUST 5 Score: N/5
```

## Phase 5: Next Steps

Present options via AskUserQuestion:

- Auto-fix issues (Recommended): Run /moai fix to automatically resolve Level 1-2 issues found in the review. Critical and complex issues will require manual attention.
- Create fix tasks: Create TaskList items for each finding so they can be addressed individually. Useful for team coordination.
- Export report: Save the review report to .moai/reports/ for future reference and tracking.
- Dismiss: Acknowledge the review without taking immediate action.

## Task Tracking

[HARD] Task management tools mandatory:
- Each critical finding tracked as a pending task via TaskCreate
- Warnings grouped by file as aggregate tasks
- Suggestions listed in report but not tracked as tasks

## Team Mode

When --team flag is provided, review delegates to the team-based multi-perspective review workflow.

Team composition: 4 review agents (security, performance, quality, UX) analyzing in parallel.

For detailed team orchestration steps, see team/review.md.

Fallback: If team mode is unavailable, standard single-agent sequential review continues.

Team Prerequisites:
- workflow.team.enabled: true in .moai/config/sections/workflow.yaml
- CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1 in environment
- If prerequisites not met: Falls back to single-agent review

## Agent Chain Summary

- Phase 1: MoAI orchestrator (change identification via git)
- Phase 2-3: manager-quality subagent (multi-perspective analysis) OR expert-security subagent (if --security)
- Phase 4-5: MoAI orchestrator (consolidation and user interaction)

## Execution Summary

1. Parse arguments (extract flags: --staged, --branch, --security, --file, --team)
2. If --team: Route to team/review.md workflow
3. Identify code changes (git diff based on flags)
4. Delegate multi-perspective review to manager-quality subagent
5. Check @MX tag compliance for changed files
6. Consolidate findings by severity
7. Present report with next step options

---

Version: 1.0.0
