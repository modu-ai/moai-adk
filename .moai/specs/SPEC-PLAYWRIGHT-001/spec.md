---
spec_id: SPEC-PLAYWRIGHT-001
title: "Playwright Active Testing"
created: "2026-04-01"
status: approved
priority: high
module: template
version: "1.0.0"
lifecycle: spec-anchored
---

# Playwright Active Testing

## Overview

evaluator-active 에이전트에 Playwright/claude-in-chrome MCP 통합을 추가하여 정적 코드 분석을 넘어 실행 중인 웹 인터페이스를 직접 탐색하고 테스트하는 능동적 평가 역량을 부여한다. 이 기능은 thorough harness 수준에서만 활성화되며, 웹 프론트엔드가 있는 프로젝트에서만 동작한다. Phase 2.8a의 평가 프로세스에 통합되어 PASS/FAIL 판정에 반영된다.

## Requirements (EARS Format)

### REQ-001: Playwright/claude-in-chrome MCP Integration

**When** evaluator-active is invoked in a web project with thorough harness level, the system **shall** enable Playwright testing capabilities through the claude-in-chrome MCP server.

The evaluator-active agent **shall** be able to:
- Navigate to URLs of the running web application
- Interact with UI elements (click, type, scroll)
- Capture and evaluate visual states
- Verify interactive behavior matches acceptance criteria
- Detect accessibility issues (missing alt text, broken tab order, insufficient contrast)

**Acceptance**:
- **Given** evaluator-active is invoked with chrome MCP tools available
- **When** a web application is running locally
- **Then** evaluator-active can navigate to the application URL
- **And** interact with UI elements programmatically
- **And** evaluate visual and interactive states against acceptance criteria

### REQ-002: Conditional Activation

**When** evaluator-active is invoked, Playwright testing **shall** activate only when ALL of the following conditions are met:
- Harness level = thorough
- Project has a web frontend (detected via `.moai/project/tech.md` technology list, or presence of `package.json` with frontend dependencies, or `index.html` in project root)

**When** any condition is not met, evaluator-active **shall** skip Playwright testing and proceed with static evaluation only.

**Acceptance**:
- **Given** harness level = thorough AND project has package.json with "react" dependency
- **When** evaluator-active is invoked
- **Then** Playwright testing capabilities are activated
- **Given** harness level = standard AND project has web frontend
- **When** evaluator-active is invoked
- **Then** Playwright testing is skipped (harness level condition not met)
- **Given** harness level = thorough AND project is a CLI tool (no web frontend)
- **When** evaluator-active is invoked
- **Then** Playwright testing is skipped (web frontend condition not met)

### REQ-003: Integration with Phase 2.8a

**When** Playwright testing is active during Phase 2.8a, the system **shall**:
- Include Playwright test results in the evaluator-active assessment
- Weight Playwright results within the Functionality dimension (interactive subset)
- Include Playwright findings in the PASS/FAIL verdict
- Report specific UI/UX issues found during interactive testing

Playwright test failures **shall** contribute to the Functionality dimension score but **shall not** independently override the overall verdict (unlike Security FAIL).

**Acceptance**:
- **Given** Phase 2.8a runs with Playwright testing active
- **When** Playwright detects a broken navigation flow
- **Then** the Functionality dimension score is reduced
- **And** the specific navigation issue is documented in the evaluation report
- **Given** Phase 2.8a runs with Playwright testing active
- **When** all Playwright interactive tests pass
- **Then** the Functionality dimension receives a positive contribution
- **And** the evaluation report notes successful interactive verification

## Exclusions (What NOT to Build)

- Shall NOT be active for CLI-only or backend-only projects
- Shall NOT require Playwright installation (uses existing claude-in-chrome MCP server)
- Shall NOT install or manage browser binaries
- Shall NOT run Playwright tests in CI/CD pipelines (local evaluation only)
- Shall NOT modify the application under test (read-only interaction)

## Acceptance Criteria

**Scenario 1: Full Playwright Evaluation (Thorough + Web)**
- Given a React application with login form and dashboard
- And harness level = thorough
- When Phase 2.8a evaluator-active runs
- Then evaluator-active navigates to the login page
- And submits test credentials via the form
- And verifies dashboard loads with expected components
- And reports interactive test results in the evaluation

**Scenario 2: Graceful Skip (No Web Frontend)**
- Given a Go CLI application with no web interface
- And harness level = thorough
- When Phase 2.8a evaluator-active runs
- Then Playwright testing is skipped with informational log
- And static code evaluation proceeds normally
- And no error is raised for missing web frontend

**Scenario 3: Graceful Skip (Standard Harness)**
- Given a web application with harness level = standard
- When Phase 2.8a evaluator-active runs
- Then Playwright testing is skipped
- And only static evaluation and acceptance criteria testing occur

**Scenario 4: Accessibility Detection**
- Given a web form with missing label associations and no focus indicators
- And harness level = thorough
- When evaluator-active runs Playwright testing
- Then accessibility issues are detected and reported
- And the Functionality dimension score reflects the accessibility gap

## Technical Approach

1. Add claude-in-chrome MCP tools to evaluator-active.md agent definition allowed tools
2. Implement web project detection heuristic (tech.md, package.json, index.html)
3. Add Playwright testing section to Phase 2.8a in run.md, gated by harness level + web detection
4. Define Playwright test flow: navigate, interact, capture, evaluate
5. Integrate Playwright results into Functionality dimension scoring

## Files to Modify

- `internal/template/templates/.claude/agents/moai/evaluator-active.md` (add chrome MCP tools to allowed tools)
- `internal/template/templates/.claude/skills/moai/workflows/run.md` (Playwright testing in Phase 2.8a)
