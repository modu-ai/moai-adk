---
name: moai-specialized-playwright
description: Playwright E2E testing framework for cross-browser automation and web testing
version: 1.0.0
modularized: false
last_updated: 2025-11-24
allowed_tools: [Context7, Skill, Task]
compliance_score: 100%
category_tier: 6
auto_trigger_keywords: [playwright, e2e, testing, automation, browser, cross-browser, api-testing, visual]
agent_coverage: [test-engineer, quality-gate, tdd-implementer]
context7_references: [/microsoft/playwright]
invocation_api_version: 1.0
dependencies: [moai-domain-testing, moai-foundation-trust]
deprecated: false
modules: null
successor: null
---

# moai-specialized-playwright: Cross-Browser E2E Testing

## Quick Reference (Level 1)

**Playwright Automation**: Cross-browser E2E testing with full API and visual testing support for modern web applications.

**Key Capabilities**:
- Multi-browser testing (Chromium, Firefox, WebKit)
- Headless and headed execution
- Network and API mocking
- Screenshot/video recording
- Accessibility testing

**When to Use**: Comprehensive web application testing, critical user journeys, visual regression testing.

---

## Implementation Guide (Level 2)

### Test Configuration

```python
from playwright.async_api import async_playwright

async def run_tests():
    async with async_playwright() as p:
        browser = await p.chromium.launch(headless=True)
        page = await browser.new_page()
        await page.goto("https://example.com")
        assert "Example" in await page.title()
```

### Element Interaction

```python
# Fill form fields
await page.fill('input[type="email"]', 'user@example.com')
await page.fill('input[type="password"]', 'secret')
await page.click('button[type="submit"]')

# Wait for navigation
await page.wait_for_navigation()
```

### API Testing

```python
async def test_api():
    async with async_playwright() as p:
        async with p.request.new_context() as context:
            response = await context.post('/api/users', data={'name': 'John'})
            assert response.ok
            user = await response.json()
```

---

## Advanced Patterns (Level 3)

### Network Interception

Mock API responses:

```python
async def handle_route(route):
    await route.abort(error_code="failed")

await page.route("**/*.json", handle_route)
```

### Visual Testing

```python
await page.screenshot(path="screenshot.png", full_page=True)
expect(page).to_have_screenshot()
```

### Accessibility Testing

```python
from axe_playwright.async_api import inject_axe, check

await inject_axe(page)
violations = await check(page)
assert len(violations) == 0
```

---

**Status**: Production Ready
**Best for**: Critical user journey testing, visual regression, cross-browser compatibility
