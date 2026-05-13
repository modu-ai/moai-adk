---
spec_id: SPEC-PLAYWRIGHT-001
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

SPEC-PLAYWRIGHT-001 adds Playwright/claude-in-chrome MCP integration to evaluator-active for active testing of web interfaces. This enables evaluator-active to navigate, interact with, and evaluate running web applications during Phase 2.8a, gated by thorough harness level and web frontend detection.

Codebase analysis shows this is **not yet implemented**. The evaluator-active agent currently uses read-only tools (`Read, Grep, Glob, Bash, mcp__sequential-thinking__sequentialthinking`) and does not include chrome MCP tools. Playwright testing is not referenced in run.md Phase 2.8a.

## 2. Gap Analysis

### Current State

| REQ | Status | Evidence |
|-----|--------|----------|
| REQ-001 | NOT DONE | evaluator-active.md tools field lacks chrome MCP tools |
| REQ-002 | NOT DONE | No web project detection logic in evaluator-active or run.md |
| REQ-003 | NOT DONE | No Playwright testing integration in Phase 2.8a |

### Implementation Required

1. Add chrome MCP tools to evaluator-active.md allowed tools
2. Add web project detection heuristic
3. Add Playwright testing section to Phase 2.8a, gated by thorough + web detection
4. Define Playwright test flow and scoring integration

## 3. Milestone Breakdown

### M1 -- Add Chrome MCP Tools to evaluator-active -- Priority P0

Add claude-in-chrome MCP tools to evaluator-active agent definition:
- Add `mcp__chrome-devtools__*` tools to the tools field
- These tools enable: navigate_page, take_snapshot, take_screenshot, click, fill, press_key
- Maintain backward compatibility: if chrome MCP is unavailable, skip gracefully

Files:
- `internal/template/templates/.claude/agents/moai/evaluator-active.md` line 12 (tools field)

### M2 -- Web Project Detection Heuristic -- Priority P0

Implement web project detection logic:
- Check `.moai/project/tech.md` technology list for frontend frameworks
- Check for `package.json` with frontend dependencies (react, vue, next, etc.)
- Check for `index.html` in project root
- Return detection result as boolean flag for gating

Files:
- `internal/template/templates/.claude/skills/moai/workflows/run.md` (add detection step before Phase 2.8a)

### M3 -- Playwright Testing in Phase 2.8a -- Priority P0

Add Playwright testing section to Phase 2.8a, gated by conditions:
- Gate 1: harness level = thorough
- Gate 2: web frontend detected (from M2 heuristic)
- Test flow: navigate -> interact -> capture -> evaluate
- Score integration: Playwright results contribute to Functionality dimension
- Playwright failures do NOT independently override overall verdict (unlike Security)

Files:
- `internal/template/templates/.claude/skills/moai/workflows/run.md` (Phase 2.8a section)

### M4 -- Accessibility Detection Integration -- Priority P1

Add accessibility checks to Playwright testing:
- Missing alt text detection
- Broken tab order detection
- Insufficient contrast detection
- Missing focus indicators
- Report accessibility findings in evaluation report

Files:
- `internal/template/templates/.claude/agents/moai/evaluator-active.md` (add accessibility evaluation section)
- `internal/template/templates/.claude/skills/moai/workflows/run.md` (add accessibility check step)

### M5 -- Graceful Skip Documentation -- Priority P2

Document graceful skip behavior for non-web or non-thorough projects:
- Informational log when Playwright testing is skipped
- No error raised for missing web frontend
- Static evaluation proceeds normally
- Skip reason included in evaluation report

Files:
- `internal/template/templates/.claude/agents/moai/evaluator-active.md` (skip conditions)
- `internal/template/templates/.claude/skills/moai/workflows/run.md` (skip branch in Phase 2.8a)

## 4. File:line Anchors

| File | Line(s) | Action | Purpose |
|------|---------|--------|---------|
| `internal/template/templates/.claude/agents/moai/evaluator-active.md` | 12 | Edit | Add chrome MCP tools |
| `internal/template/templates/.claude/agents/moai/evaluator-active.md` | 46-55 | Edit | Update dimensions for Playwright results |
| `internal/template/templates/.claude/skills/moai/workflows/run.md` | 818-858 | Edit | Add Playwright testing in Phase 2.8a |

## 5. Quality Gates

- evaluator-active can use chrome MCP tools when available
- Playwright testing activates only when thorough + web frontend detected
- Playwright failures contribute to Functionality dimension but do not override verdict
- Graceful skip with informational log when conditions not met
- No regression in static evaluation when Playwright is unavailable

## 6. Dependencies

- SPEC-EVAL-001: evaluator-active agent definition (must exist first)
- SPEC-EVALLIB-001: frontend.md evaluator profile (Playwright scoring integration)
- claude-in-chrome MCP server: Must be configured and available at runtime
