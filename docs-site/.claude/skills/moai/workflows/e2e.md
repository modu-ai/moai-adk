---
name: moai-workflow-e2e
description: >
  Create and run E2E tests using Agent Browser, Playwright CLI, or Claude in Chrome.
  Supports auto-detection and installation, user journey mapping, GIF recording,
  and automated test execution with intelligent tool selection.
user-invocable: false
metadata:
  version: "2.7.0"
  category: "workflow"
  status: "active"
  updated: "2026-02-21"
  tags: "e2e, end-to-end, testing, browser, playwright, chrome, agent-browser, user-journey"
  context7-libraries: "microsoft/playwright"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["e2e", "end-to-end", "e2e test", "browser test", "playwright", "agent-browser", "user journey"]
  agents: ["expert-testing", "expert-frontend"]
  phases: ["e2e"]
---

# Workflow: E2E - End-to-End Testing

Purpose: Create and run end-to-end tests that validate complete user flows through the application. Supports three browser automation tools with auto-detection, auto-installation, and intelligent selection.

Flow: Tool Selection -> Installation -> Journey Mapping -> Test Script Creation -> Execution -> Recording -> Report

## Supported Flags

- --tool TOOL: Force browser tool selection, skipping user prompt. Options: agent-browser, playwright, chrome-mcp (default: always ask user)
- --record: Record browser interactions as GIF for visual verification
- --url URL: Target URL for browser-based testing (default: auto-detect from project config)
- --journey NAME: Run a specific named user journey only
- --headless: Run browser tests in headless mode (default: true)
- --browser BROWSER: Browser to use for Playwright (default: chromium). Options: chromium, firefox, webkit
- --timeout N: Test timeout in seconds (default: 30)
- --retry N: Number of retries for flaky tests (default: 1)

## Browser Automation Tools

### Tool Comparison

| Feature | Agent Browser | Playwright CLI | Claude in Chrome |
|---------|--------------|----------------|------------------|
| **Token Cost** | Low (CLI output) | Low (CLI output) | High (MCP round-trips) |
| **Setup** | npm install | npx playwright install | Chrome extension required |
| **Headless** | Yes | Yes | No (requires visible Chrome) |
| **Cross-Browser** | Chromium only | Chromium, Firefox, WebKit | Chrome only |
| **GIF Recording** | Via Playwright trace | Via Playwright trace | Via mcp__claude-in-chrome__gif_creator |
| **AI Navigation** | Built-in AI agent | Script-based | MCP tool-based AI |
| **Best For** | AI-driven exploration | Deterministic test suites | Interactive debugging |
| **Install Size** | ~50MB | ~200MB | N/A (extension) |
| **CI/CD** | Yes | Yes | No |
| **Source** | github.com/vercel-labs/agent-browser | playwright.dev | Built-in MCP |

### Auto-Detection Priority

When --tool flag is not provided, auto-detect in this order:

1. Check if `agent-browser` is installed: `npx agent-browser --version` or `bunx agent-browser --version`
2. Check if Playwright is installed: `npx playwright --version` or `bunx playwright --version`
3. Check if Chrome MCP is available: Verify mcp__claude-in-chrome tools exist

Detection results are used to mark availability status in AskUserQuestion options.

### Recommendation Logic

[HARD] Always present tool selection to user via AskUserQuestion. The recommendation logic below determines which option is marked "(Recommended)":

| Condition | Recommended Tool | Rationale |
|-----------|-----------------|-----------|
| --record flag with no specific tool | Claude in Chrome | Best GIF recording via MCP |
| CI/CD environment detected | Playwright CLI | Most reliable headless support |
| Journey requires AI exploration | Agent Browser | Built-in AI navigation |
| Deterministic test suite needed | Playwright CLI | Most stable, cross-browser |
| Interactive debugging | Claude in Chrome | Visual real-time feedback |
| Default (no special condition) | Playwright CLI | Best balance of features and token efficiency |

## Phase 0: Tool Selection & Installation

### Step 0.1: Tool Detection

[HARD] Delegate tool detection to the expert-testing subagent.

Detection commands (run in parallel):
- Agent Browser: `npx agent-browser --version 2>/dev/null || echo "not-installed"`
- Playwright: `npx playwright --version 2>/dev/null || echo "not-installed"`
- Chrome MCP: Check if mcp__claude-in-chrome tools are available in current session

### Step 0.2: User Selection

[HARD] Always present tool selection to user via AskUserQuestion, regardless of how many tools are detected.

If --tool flag provided: Use specified tool, skip to Step 0.3.

If --tool flag NOT provided: Always present via AskUserQuestion with detection status and recommendation:

- Playwright CLI: Deterministic test execution with cross-browser support. Most stable and token-efficient. Ideal for CI/CD pipelines and comprehensive test suites. (Mark with availability status from detection)
- Agent Browser: AI-powered browser navigation by Vercel. The agent autonomously explores and interacts with web pages. Best for exploratory testing and AI-driven user journey validation. (Mark with availability status from detection)
- Claude in Chrome: Real-time browser automation via MCP tools. Best for interactive debugging, visual verification, and GIF recording. Requires Chrome with Claude extension. (Mark with availability status from detection)

Mark the recommended tool with "(Recommended)" based on the Recommendation Logic above. Include installation status (installed/not installed) in each option description so the user can make an informed choice.

### Step 0.3: Installation (if needed)

If selected tool is not installed:

**Playwright Installation:**
```bash
# Node.js project
npx playwright install --with-deps chromium
# Or with Bun
bunx playwright install --with-deps chromium

# For cross-browser testing
npx playwright install --with-deps
```

**Agent Browser Installation:**
```bash
# Install globally
npm install -g agent-browser
# Or project-local
npm install --save-dev agent-browser

# Verify
npx agent-browser --version
```

**Claude in Chrome:** No installation needed (built-in MCP). If Chrome MCP tools are not available, inform user to:
- Install Claude Code Chrome extension
- Open Chrome browser
- Connect via Claude Code

After installation, verify the tool works with a simple test before proceeding.

### Skill Reference

For Playwright best practices and patterns, reference: Skill("moai-workflow-testing") which includes:
- playwright-best-practices.md: Page Object Model, cross-browser strategy, performance integration
- Context7 integration: `microsoft/playwright` library ID for latest API patterns

## Phase 1: Journey Mapping

[HARD] Delegate journey mapping to the expert-testing subagent.

If --journey flag: Load the specified journey definition and skip to Phase 2.

If no --journey flag: Analyze the application to identify key user journeys.

Journey Discovery:
- Read project documentation (.moai/project/product.md) for feature descriptions
- Analyze route definitions (routes.ts, urls.py, router.go) for available paths
- Identify form elements, authentication flows, and CRUD operations
- Map critical user paths (login, main feature, error handling)

Present discovered journeys via AskUserQuestion:

- Test all journeys (Recommended): Run E2E tests for all discovered user journeys. This provides the most comprehensive coverage but takes longer to execute.
- Select specific journeys: Choose which journeys to test from the discovered list. Useful when you want to focus on recently changed features.
- Define custom journey: Describe a custom user journey to test. MoAI will create the test script based on your description.

Journey Definition Format:
```markdown
Journey: User Login
Steps:
1. Navigate to /login
2. Enter email in #email field
3. Enter password in #password field
4. Click Submit button
5. Verify redirect to /dashboard
6. Verify welcome message displayed
```

## Phase 2: Test Script Creation

[HARD] Delegate test script creation to the expert-testing subagent.

### Playwright CLI Mode

Generate Playwright test files following project conventions:

Test File Naming:
- TypeScript: `e2e/{journey-name}.spec.ts`
- JavaScript: `e2e/{journey-name}.spec.js`
- Python: `e2e/test_{journey_name}.py`

Include:
- Page Object Model pattern (reference playwright-best-practices.md)
- Setup and teardown fixtures
- Step-by-step assertions with descriptive names
- Screenshot capture at key verification points
- Network response validation where applicable
- Accessibility checks via `@axe-core/playwright` if available

### Agent Browser Mode

Generate Agent Browser task definitions:

Task File: `e2e/{journey-name}.agent.ts` or `e2e/{journey-name}.agent.js`

Include:
- Natural language instruction for each journey step
- Validation assertions for expected outcomes
- Fallback instructions for alternative UI layouts
- Timeout configuration per step

### Claude in Chrome Mode

No test script files generated. Instead, create journey step definitions as structured prompts for MCP tool execution in Phase 3.

## Phase 3: Test Execution

[HARD] Delegate test execution to the expert-testing subagent (or expert-frontend for Chrome MCP).

### Playwright CLI Execution

Run via CLI (token-efficient - CLI output only, no MCP round-trips):
```bash
# Run all E2E tests
npx playwright test e2e/

# Run specific journey
npx playwright test e2e/{journey-name}.spec.ts

# With HTML report
npx playwright test --reporter=html

# Headless (default) or headed
npx playwright test --headed  # For debugging
```

Parse CLI output for results. No MCP tokens consumed.

### Agent Browser Execution

Run via CLI:
```bash
# Run with natural language task
npx agent-browser --task "Navigate to {url} and {journey_steps}"

# Run task file
npx agent-browser --file e2e/{journey-name}.agent.ts

# With trace recording
npx agent-browser --task "..." --trace
```

Parse CLI output for results. Agent Browser handles AI navigation internally.

### Claude in Chrome Execution

Execute via MCP tools (higher token cost but real-time interaction):
- mcp__claude-in-chrome__navigate: Page navigation
- mcp__claude-in-chrome__find: Element location
- mcp__claude-in-chrome__form_input: Form interaction
- mcp__claude-in-chrome__computer: Click/keyboard interactions
- mcp__claude-in-chrome__read_page: Content verification
- mcp__claude-in-chrome__get_page_text: Full page text extraction

Execution flow per journey step:
1. Navigate to URL
2. Wait for page load (read_page to verify)
3. Interact with elements (form_input, computer)
4. Verify expected outcome (read_page, get_page_text)
5. Capture screenshot if verification point

### Common Options (all tools)

- If --headless flag: Force headless mode (Playwright/Agent Browser only)
- If --timeout N: Set per-test timeout
- If --retry N: Retry failed tests up to N times

## Phase 4: Recording (Optional)

If --record flag:

### Playwright Recording
```bash
# Generate trace (includes screenshots, network, console)
npx playwright test --trace on

# View trace
npx playwright show-trace trace.zip
```
Trace files stored in `e2e/traces/` directory.

### Agent Browser Recording
```bash
# With trace flag
npx agent-browser --task "..." --trace --output e2e/recordings/
```

### Claude in Chrome Recording
- Use mcp__claude-in-chrome__gif_creator to record browser interactions
- Capture extra frames before and after actions for smooth playback
- Generate GIF with meaningful filename: `e2e/recordings/{journey-name}.gif`

Recording Best Practices:
- Name files descriptively (e.g., "login_flow.gif", "checkout_process.gif")
- Include timestamp in filename for versioning
- Store all recordings in `e2e/recordings/` directory

## Phase 5: Report

Display E2E test results in user's conversation_language:

```markdown
## E2E Test Report

### Tool Used: {Playwright CLI | Agent Browser | Claude in Chrome}

### Results Summary
| Journey | Status | Duration | Screenshots |
|---------|--------|----------|-------------|
| Login | PASS | 2.3s | 3 captured |
| Checkout | FAIL | 5.1s | 4 captured |

### Failures
- Checkout (Step 4): Expected redirect to /confirmation but got /error
  - Screenshot: e2e/screenshots/checkout-step4.png
  - Error: TimeoutError: Navigation timeout of 30000ms exceeded

### Recordings (if --record)
- e2e/recordings/login_flow.gif
- e2e/recordings/checkout_process.gif

### Coverage
- User journeys tested: 5/7
- Critical paths covered: 3/3
- Error scenarios tested: 2/4
```

Next Steps (AskUserQuestion):

- Fix failing tests (Recommended): Debug and fix the failing E2E tests. MoAI will analyze the failure screenshots and error messages to identify root causes.
- Rerun failed tests: Retry only the failed tests with increased timeout. Useful for flaky tests that may pass on retry.
- Add more journeys: Define additional user journeys to improve E2E coverage. MoAI will suggest uncovered critical paths.
- Switch tool: Try a different browser automation tool for comparison. Useful when current tool has limitations for specific test scenarios.

## Task Tracking

[HARD] Task management tools mandatory:
- Each user journey tracked as a pending task via TaskCreate
- Before test execution: change to in_progress via TaskUpdate
- After test passes: change to completed via TaskUpdate
- Failed tests remain in_progress with failure details

## Agent Chain Summary

- Phase 0: expert-testing subagent (tool detection and installation)
- Phase 1: expert-testing subagent (journey mapping)
- Phase 2: expert-testing subagent (test script creation)
- Phase 3: expert-testing or expert-frontend subagent (test execution)
- Phase 4: expert-frontend subagent (GIF recording via Chrome MCP) or expert-testing (Playwright/Agent Browser traces)
- Phase 5: MoAI orchestrator (report and user interaction)

## Execution Summary

1. Parse arguments (extract flags: --tool, --record, --url, --journey, --headless, --browser, --timeout, --retry)
2. Phase 0: Detect available tools, always prompt user with recommendation, install if needed
3. If --journey: Load specific journey, skip to Phase 2
4. Phase 1: Delegate journey mapping to expert-testing subagent
5. Present journey options to user via AskUserQuestion
6. Phase 2: Delegate test script creation to expert-testing subagent
7. Phase 3: Delegate test execution to expert-testing/expert-frontend subagent
8. Phase 4: If --record, capture recordings via selected tool's mechanism
9. TaskCreate/TaskUpdate for all journeys
10. Phase 5: Report results with next step options

---

Version: 2.1.0
