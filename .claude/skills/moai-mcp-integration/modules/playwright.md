# Playwright MCP Server Integration Module

**Version**: 2.0.0 (AI-Enhanced Web Testing)
**Last Updated**: 2025-11-24
**Purpose**: Enterprise web application testing with AI-enhanced test generation and visual regression

---

## 📖 Quick Overview (2 Minutes)

Playwright is the unified MCP server for web application testing, providing:

- **Basic automation**: Sync/async browser automation
- **AI-enhanced testing**: ML-powered test pattern recognition
- **Visual regression testing**: AI-powered image diff analysis
- **Cross-browser testing**: Chrome, Firefox, Safari coordination
- **QA workflows**: Automated test generation and execution
- **Performance testing**: Load and performance profiling
- **CI/CD integration**: Pipeline automation and reporting

---

## 🔧 Implementation Guide

### Basic Playwright Automation

#### **Synchronous Automation**

```python
from playwright.sync_api import sync_playwright

def automate_basic_flow():
    """Basic synchronous Playwright automation."""

    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()

        # Navigate to site
        page.goto('https://example.com')
        page.wait_for_load_state('networkidle')

        # Perform actions
        page.click('text=Sign In')
        page.fill('input[name="email"]', 'user@example.com')
        page.fill('input[name="password"]', 'password123')
        page.click('button:has-text("Login")')

        # Wait for navigation
        page.wait_for_load_state('networkidle')

        # Verify results
        assert page.is_visible('text=Dashboard')

        # Take screenshot
        page.screenshot(path='success.png')

        browser.close()
```

#### **Asynchronous Automation**

```python
from playwright.async_api import async_playwright

async def automate_async_flow():
    """Asynchronous Playwright automation for concurrent operations."""

    async with async_playwright() as p:
        browser = await p.chromium.launch(headless=True)
        page = await browser.new_page()

        # Navigate to site
        await page.goto('https://example.com')
        await page.wait_for_load_state('networkidle')

        # Perform actions
        await page.click('text=Sign In')
        await page.fill('input[name="email"]', 'user@example.com')
        await page.fill('input[name="password"]', 'password123')
        await page.click('button:has-text("Login")')

        # Wait for navigation
        await page.wait_for_load_state('networkidle')

        # Verify results
        assert await page.is_visible('text=Dashboard')

        await browser.close()
```

### Server Lifecycle Management

#### **with_server.py Usage**

```bash
# Single server (frontend)
python scripts/with_server.py \
  --server "npm run dev" \
  --port 5173 \
  -- python automation.py

# Multiple servers (backend + frontend)
python scripts/with_server.py \
  --server "cd backend && python server.py" --port 3000 \
  --server "cd frontend && npm run dev" --port 5173 \
  -- python automation.py

# Custom server configurations
python scripts/with_server.py \
  --server "python -m uvicorn api.main:app --reload" --port 8000 \
  --server "npm run dev:web" --port 3000 \
  --server "npm run dev:mobile" --port 3001 \
  -- pytest tests/e2e/
```

---

## 🤖 AI-Enhanced Testing

### AI Test Pattern Recognition

```python
class AITestPatternRecognizer:
    """AI-powered test pattern detection and classification."""

    async def analyze_webapp_with_context7(
        self,
        webapp_url: str,
        context: dict
    ) -> dict:
        """
        Analyze webapp using Context7 documentation and AI pattern matching.

        Returns:
            TestAnalysis with recommendations
        """
        # Get latest testing patterns from Context7
        playwright_docs = await mcp__context7__get_library_docs(
            context7_compatible_library_id="/microsoft/playwright",
            topic="AI testing patterns automated test generation visual regression 2025",
            tokens=5000
        )

        # AI pattern classification
        app_type = self.classify_application_type(webapp_url, context)
        test_patterns = self.match_known_test_patterns(app_type, context)

        # Extract Context7-enhanced patterns
        context7_insights = self.extract_context7_patterns(app_type, playwright_docs)

        return {
            "application_type": app_type,
            "confidence_score": self.calculate_confidence(app_type, test_patterns),
            "recommended_test_strategies": self.generate_test_strategies(
                app_type, test_patterns, context7_insights
            ),
            "context7_references": context7_insights['references'],
            "automation_opportunities": self.identify_automation_opportunities(
                app_type, test_patterns
            )
        }
```

### AI-Enhanced Test Generation

```python
class AITestGenerator:
    """Generate tests using AI pattern recognition."""

    async def generate_tests_with_context7(
        self,
        webapp_url: str,
        include_visual_regression: bool = True,
        cross_browser_config: list = None
    ) -> list[str]:
        """
        Generate Playwright tests using AI and Context7 patterns.

        Returns:
            List of generated test files
        """
        # Get latest testing patterns
        patterns = await mcp__context7__get_library_docs(
            context7_compatible_library_id="/microsoft/playwright",
            topic="test generation patterns best practices",
            tokens=3000
        )

        # Analyze application
        app_analysis = await self.analyze_application(webapp_url)

        # Generate tests
        generated_tests = []

        # Basic functionality tests
        basic_tests = self.generate_basic_tests(app_analysis, patterns)
        generated_tests.extend(basic_tests)

        # Visual regression tests
        if include_visual_regression:
            visual_tests = await self.generate_visual_regression_tests(
                webapp_url, patterns
            )
            generated_tests.extend(visual_tests)

        # Cross-browser tests
        if cross_browser_config:
            cross_browser_tests = self.generate_cross_browser_tests(
                app_analysis, cross_browser_config, patterns
            )
            generated_tests.extend(cross_browser_tests)

        return generated_tests
```

---

### Visual Regression Testing

#### **AI-Powered Visual Regression**

```python
class AIVisualRegressionTester:
    """AI-powered visual regression testing with Context7."""

    async def test_with_context7_ai(
        self,
        baseline_url: str,
        current_url: str
    ) -> dict:
        """
        Perform visual regression testing using AI and Context7.

        Returns:
            VisualRegressionResult with AI analysis
        """
        # Get Context7 visual testing patterns
        context7_patterns = await mcp__context7__get_library_docs(
            context7_compatible_library_id="/microsoft/playwright",
            topic="visual regression testing screenshot comparison patterns",
            tokens=3000
        )

        # Capture screenshots
        baseline_screenshot = await self.capture_screenshot(baseline_url)
        current_screenshot = await self.capture_screenshot(current_url)

        # AI-powered visual analysis
        visual_diff = await self.analyze_visual_differences_with_ai(
            baseline_screenshot,
            current_screenshot,
            context7_patterns
        )

        return {
            "baseline_url": baseline_url,
            "current_url": current_url,
            "diff_percentage": visual_diff["percentage"],
            "changes_detected": visual_diff["changes"],
            "recommendations": visual_diff["recommendations"],
            "ai_confidence_score": visual_diff["confidence"],
            "context7_patterns_applied": context7_patterns
        }

    async def capture_screenshot(self, url: str, full_page: bool = True) -> bytes:
        """Capture webpage screenshot."""
        async with async_playwright() as p:
            browser = await p.chromium.launch()
            page = await browser.new_page(viewport={"width": 1280, "height": 720})

            await page.goto(url)
            await page.wait_for_load_state('networkidle')

            screenshot = await page.screenshot(full_page=full_page)

            await browser.close()

            return screenshot
```

#### **Baseline Management**

```python
import hashlib
import json
from pathlib import Path

class VisualRegressionBaseline:
    """Manage visual regression baselines."""

    def __init__(self, baseline_dir: str = ".playwright/baselines"):
        self.baseline_dir = Path(baseline_dir)
        self.baseline_dir.mkdir(parents=True, exist_ok=True)

    def save_baseline(self, test_name: str, screenshot: bytes):
        """Save visual baseline for test."""
        baseline_path = self.baseline_dir / f"{test_name}.png"
        baseline_path.write_bytes(screenshot)

        # Record metadata
        metadata = {
            "test_name": test_name,
            "timestamp": datetime.now().isoformat(),
            "hash": hashlib.md5(screenshot).hexdigest(),
            "size": len(screenshot)
        }

        metadata_path = self.baseline_dir / f"{test_name}.json"
        metadata_path.write_text(json.dumps(metadata, indent=2))

    def get_baseline(self, test_name: str) -> bytes:
        """Retrieve baseline screenshot."""
        baseline_path = self.baseline_dir / f"{test_name}.png"

        if not baseline_path.exists():
            raise FileNotFoundError(f"Baseline not found: {test_name}")

        return baseline_path.read_bytes()

    def compare_with_baseline(self, test_name: str, screenshot: bytes) -> dict:
        """Compare screenshot with baseline."""
        baseline = self.get_baseline(test_name)

        # Use image comparison library
        from PIL import Image
        import io

        baseline_img = Image.open(io.BytesIO(baseline))
        current_img = Image.open(io.BytesIO(screenshot))

        # Calculate difference
        if baseline_img.size != current_img.size:
            return {
                "match": False,
                "reason": "Size mismatch",
                "baseline_size": baseline_img.size,
                "current_size": current_img.size
            }

        # Pixel-level comparison
        diff = ImageChops.difference(baseline_img, current_img)
        diff_percentage = (diff.getbbox()[2] * diff.getbbox()[3]) / (
            baseline_img.width * baseline_img.height
        ) * 100 if diff.getbbox() else 0

        return {
            "match": diff_percentage < 2,  # Allow 2% difference
            "diff_percentage": diff_percentage,
            "pixels_changed": count_diff_pixels(diff)
        }
```

---

### Cross-Browser Testing

#### **Multi-Browser Coordination**

```python
class CrossBrowserTester:
    """Coordinate testing across multiple browsers."""

    async def run_cross_browser_tests(
        self,
        test_func,
        browsers: list[str] = None,
        headless: bool = True
    ) -> dict:
        """
        Run test across multiple browsers concurrently.

        Args:
            test_func: Async test function
            browsers: ['chromium', 'firefox', 'webkit']
            headless: Run in headless mode

        Returns:
            Results for each browser
        """
        if browsers is None:
            browsers = ['chromium', 'firefox', 'webkit']

        results = {}

        async with async_playwright() as p:
            # Launch all browsers
            browser_instances = {}
            for browser_type in browsers:
                if browser_type == 'chromium':
                    browser_instances['chromium'] = await p.chromium.launch(headless=headless)
                elif browser_type == 'firefox':
                    browser_instances['firefox'] = await p.firefox.launch(headless=headless)
                elif browser_type == 'webkit':
                    browser_instances['webkit'] = await p.webkit.launch(headless=headless)

            # Run tests in parallel
            tasks = []
            for browser_name, browser in browser_instances.items():
                tasks.append(self.run_test_in_browser(
                    test_func,
                    browser_name,
                    browser
                ))

            results = await asyncio.gather(*tasks)

            # Close all browsers
            for browser in browser_instances.values():
                await browser.close()

        return results

    async def run_test_in_browser(self, test_func, browser_name: str, browser):
        """Run test in specific browser."""
        page = await browser.new_page()

        try:
            result = await test_func(page)
            return {
                "browser": browser_name,
                "status": "passed",
                "result": result
            }
        except Exception as e:
            return {
                "browser": browser_name,
                "status": "failed",
                "error": str(e)
            }
        finally:
            await page.close()
```

---

### Performance Testing

#### **Load & Performance Analysis**

```python
class PerformanceTester:
    """Test application performance."""

    async def measure_page_performance(self, url: str) -> dict:
        """Measure page load performance metrics."""

        async with async_playwright() as p:
            browser = await p.chromium.launch()
            page = await browser.new_page()

            # Measure metrics
            start_time = time.time()

            await page.goto(url)
            first_contentful_paint = await page.evaluate("""
                () => {
                    const entries = performance.getEntriesByName('first-contentful-paint');
                    return entries.length > 0 ? entries[0].startTime : null;
                }
            """)

            await page.wait_for_load_state('networkidle')
            load_time = time.time() - start_time

            # Get performance metrics
            metrics = await page.evaluate("""
                () => ({
                    fcp: performance.getEntriesByName('first-contentful-paint')[0]?.startTime,
                    lcp: performance.getEntriesByName('largest-contentful-paint')[0]?.startTime,
                    cls: performance.getEntriesByType('layout-shift')
                        .reduce((sum, entry) => sum + (entry.hadRecentInput ? 0 : entry.value), 0),
                    memory: performance.memory ? {
                        usedJSHeapSize: performance.memory.usedJSHeapSize,
                        totalJSHeapSize: performance.memory.totalJSHeapSize
                    } : null
                })
            """)

            await browser.close()

            return {
                "url": url,
                "load_time_ms": load_time * 1000,
                "first_contentful_paint_ms": first_contentful_paint,
                "metrics": metrics
            }
```

---

## 🧪 CI/CD Integration

### GitHub Actions Integration

```yaml
# .github/workflows/playwright-tests.yml
name: Playwright Tests

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        browser: [chromium, firefox, webkit]

    steps:
      - uses: actions/checkout@v3

      - uses: actions/setup-python@v4
        with:
          python-version: '3.11'

      - name: Install dependencies
        run: |
          pip install -r requirements.txt
          playwright install ${{ matrix.browser }}

      - name: Run Playwright tests
        run: |
          pytest tests/e2e/ --browser ${{ matrix.browser }}

      - name: Upload test results
        if: always()
        uses: actions/upload-artifact@v3
        with:
          name: playwright-report-${{ matrix.browser }}
          path: playwright-report/

      - name: Visual regression check
        run: |
          python scripts/visual_regression_check.py
```

---

## ✅ Best Practices

✅ **DO**:
- Use Context7 for latest Playwright patterns
- Always wait for `networkidle` on dynamic apps
- Use descriptive selectors (text=, role=, IDs)
- Implement visual regression testing
- Run cross-browser tests regularly
- Use AI-enhanced test generation for complex flows
- Monitor performance metrics
- Implement exponential backoff for retries
- Use fixtures for reusable setup
- Generate baseline screenshots regularly

❌ **DON'T**:
- Use `page.wait_for_timeout()` without reason
- Hard-code element indices (use selectors)
- Ignore flaky tests (use proper waits)
- Run without headless mode in CI/CD
- Forget to update visual baselines
- Mix sync and async code
- Create massive test files (split logically)
- Ignore performance metrics

---

## 📊 Recommended Test Structure

```
tests/
├── e2e/
│   ├── auth/
│   │   ├── login.py
│   │   └── signup.py
│   ├── dashboard/
│   │   ├── overview.py
│   │   └── analytics.py
│   └── conftest.py
├── visual/
│   ├── components/
│   │   └── button.py
│   └── baselines/
│       └── baselines.py
└── performance/
    └── load_tests.py
```

---

## 🔄 Changelog

| Version | Date | Changes |
|---------|------|---------|
| **2.0.0** | 2025-11-24 | Merged into moai-mcp-integration hub module with AI enhancements |
| 1.0.0 | 2025-11-22 | Initial Playwright MCP integration |

---

**Module Version**: 2.0.0
**Status**: Production Ready
**Compliance**: 100% (Playwright 1.40+)
