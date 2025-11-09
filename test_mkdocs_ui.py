#!/usr/bin/env python3
"""
MkDocs UI/UX Visual Testing with Playwright
Tests desktop, mobile, and dark mode rendering
"""

import asyncio
import os
import sys
import subprocess
import time
from pathlib import Path
from datetime import datetime

try:
    from playwright.async_api import async_playwright, Browser, Page, BrowserContext
except ImportError:
    print("ERROR: Playwright not installed. Run: pip install playwright")
    print("Then run: playwright install")
    sys.exit(1)


class MkDocsUITester:
    """Visual testing suite for MkDocs documentation"""

    def __init__(self, base_url: str = "http://localhost:8000", docs_dir: str = "/Users/goos/MoAI/MoAI-ADK/docs"):
        self.base_url = base_url
        self.docs_dir = docs_dir
        self.mkdocs_process = None
        self.screenshots_dir = Path(docs_dir) / "screenshots"
        self.screenshots_dir.mkdir(exist_ok=True)
        self.results = []

    async def start_mkdocs_server(self) -> bool:
        """Start mkdocs serve in the background"""
        print(f"\n[1/6] Starting MkDocs server at {self.base_url}...")

        try:
            # Check if server is already running
            import urllib.request
            try:
                urllib.request.urlopen(self.base_url, timeout=2)
                print("✓ MkDocs server already running")
                return True
            except Exception:
                pass

            # Start mkdocs serve
            os.chdir(self.docs_dir)
            self.mkdocs_process = subprocess.Popen(
                ["mkdocs", "serve", "--dev-addr", "127.0.0.1:8000"],
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE
            )

            # Wait for server to start
            for attempt in range(30):
                try:
                    import urllib.request
                    urllib.request.urlopen(self.base_url, timeout=1)
                    print(f"✓ MkDocs server started successfully")
                    return True
                except Exception:
                    await asyncio.sleep(0.5)

            print("✗ Failed to start MkDocs server")
            return False

        except Exception as e:
            print(f"✗ Error starting MkDocs: {e}")
            return False

    def stop_mkdocs_server(self):
        """Stop the mkdocs server"""
        if self.mkdocs_process:
            print("\nStopping MkDocs server...")
            self.mkdocs_process.terminate()
            try:
                self.mkdocs_process.wait(timeout=5)
            except subprocess.TimeoutExpired:
                self.mkdocs_process.kill()

    async def test_desktop_view(self, browser, context: BrowserContext, page: Page):
        """Test desktop view (1400px width)"""
        print("\n[2/6] Testing Desktop View (1400x900)...")

        await page.set_viewport_size({"width": 1400, "height": 900})
        await page.goto(self.base_url)
        await page.wait_for_load_state("networkidle")

        # Wait for page rendering
        await asyncio.sleep(1)

        # Take screenshots
        screenshot_path = self.screenshots_dir / f"desktop-homepage-{self._timestamp()}.png"
        await page.screenshot(path=str(screenshot_path), full_page=True)
        print(f"✓ Homepage screenshot: {screenshot_path}")
        self.results.append(f"Desktop homepage: {screenshot_path}")

        # Test TOC visibility
        toc_element = await page.query_selector(".md-toc")
        if toc_element:
            print("✓ TOC (Table of Contents) visible on desktop")
            self.results.append("TOC visible on desktop: YES")
        else:
            print("⚠ TOC not found on desktop")
            self.results.append("TOC visible on desktop: NO")

        # Check navigation
        nav = await page.query_selector(".md-nav")
        if nav:
            print("✓ Navigation sidebar visible")
            self.results.append("Navigation sidebar: YES")

        return screenshot_path

    async def test_mobile_view(self, browser, context: BrowserContext, page: Page):
        """Test mobile view (750px width)"""
        print("\n[3/6] Testing Mobile View (750x1000)...")

        await page.set_viewport_size({"width": 750, "height": 1000})
        await page.goto(self.base_url)
        await page.wait_for_load_state("networkidle")
        await asyncio.sleep(1)

        # Take screenshot
        screenshot_path = self.screenshots_dir / f"mobile-homepage-{self._timestamp()}.png"
        await page.screenshot(path=str(screenshot_path), full_page=True)
        print(f"✓ Mobile screenshot: {screenshot_path}")
        self.results.append(f"Mobile homepage: {screenshot_path}")

        # Check mobile menu
        mobile_menu = await page.query_selector(".md-header__button")
        if mobile_menu:
            print("✓ Mobile menu button visible")
            self.results.append("Mobile menu button: YES")

        return screenshot_path

    async def test_dark_mode(self, browser, context: BrowserContext, page: Page):
        """Test dark mode toggle"""
        print("\n[4/6] Testing Dark Mode...")

        await page.set_viewport_size({"width": 1400, "height": 900})
        await page.goto(self.base_url)
        await page.wait_for_load_state("networkidle")

        # Find and click dark mode toggle
        toggle_button = await page.query_selector("[data-md-color-scheme]")
        if toggle_button:
            # Click theme toggle
            theme_toggle = await page.query_selector(".md-header__button[aria-label*='mode']")
            if theme_toggle:
                await theme_toggle.click()
                await asyncio.sleep(1)

                screenshot_path = self.screenshots_dir / f"dark-mode-{self._timestamp()}.png"
                await page.screenshot(path=str(screenshot_path), full_page=True)
                print(f"✓ Dark mode screenshot: {screenshot_path}")
                self.results.append(f"Dark mode: {screenshot_path}")
                return True

        print("⚠ Dark mode toggle not found")
        self.results.append("Dark mode toggle: NOT FOUND")
        return False

    async def test_navigation_and_toc(self, browser, context: BrowserContext, page: Page):
        """Test navigation menu and TOC links"""
        print("\n[5/6] Testing Navigation and TOC...")

        await page.set_viewport_size({"width": 1400, "height": 900})
        await page.goto(self.base_url)
        await page.wait_for_load_state("networkidle")

        tests = {
            "Navigation visible": ".md-nav",
            "TOC visible": ".md-toc",
            "Search bar": ".md-search__input",
            "Language selector": ".md-select",
            "Edit button": "button[title*='Edit']",
            "Code copy button": "button[data-md-clipboard]",
        }

        for test_name, selector in tests.items():
            element = await page.query_selector(selector)
            status = "✓" if element else "⚠"
            print(f"{status} {test_name}: {'Found' if element else 'Not found'}")
            self.results.append(f"{test_name}: {'Found' if element else 'Not found'}")

        # Test navigation click
        nav_item = await page.query_selector(".md-nav__item a")
        if nav_item:
            href = await nav_item.get_attribute("href")
            print(f"✓ Navigation item found: {href}")
            self.results.append(f"Navigation link: {href}")

    async def test_console_errors(self, browser, context: BrowserContext, page: Page):
        """Capture console errors and warnings"""
        print("\n[6/6] Checking Console for Errors/Warnings...")

        console_messages = []
        page.on("console", lambda msg: console_messages.append({
            "type": msg.type,
            "text": msg.text,
            "location": str(msg.location)
        }))

        await page.set_viewport_size({"width": 1400, "height": 900})
        await page.goto(self.base_url)
        await page.wait_for_load_state("networkidle")
        await asyncio.sleep(1)

        errors = [m for m in console_messages if m["type"] == "error"]
        warnings = [m for m in console_messages if m["type"] == "warning"]

        print(f"✓ Console messages captured:")
        print(f"  - Errors: {len(errors)}")
        print(f"  - Warnings: {len(warnings)}")

        if errors:
            print("\nFound console errors:")
            for error in errors[:5]:  # Show first 5
                print(f"  ✗ {error['text']}")
        else:
            print("✓ No console errors detected")

        if warnings:
            print("\nFound console warnings:")
            for warning in warnings[:5]:  # Show first 5
                print(f"  ⚠ {warning['text']}")

        self.results.append(f"Console errors: {len(errors)}")
        self.results.append(f"Console warnings: {len(warnings)}")

        return len(errors) == 0

    def _timestamp(self) -> str:
        """Generate timestamp for screenshot filename"""
        return datetime.now().strftime("%Y%m%d-%H%M%S")

    async def run_all_tests(self):
        """Run all tests"""
        print("\n" + "="*70)
        print("MkDocs UI/UX Visual Testing with Playwright")
        print("="*70)

        # Start server
        if not await self.start_mkdocs_server():
            print("✗ Failed to start MkDocs server")
            return False

        try:
            async with async_playwright() as p:
                browser = await p.chromium.launch()
                context = await browser.new_context(
                    viewport={"width": 1400, "height": 900}
                )
                page = await context.new_page()

                # Run all tests
                await self.test_desktop_view(browser, context, page)
                await self.test_mobile_view(browser, context, page)
                await self.test_dark_mode(browser, context, page)
                await self.test_navigation_and_toc(browser, context, page)
                await self.test_console_errors(browser, context, page)

                # Cleanup
                await context.close()
                await browser.close()

                return True

        except Exception as e:
            print(f"\n✗ Test error: {e}")
            import traceback
            traceback.print_exc()
            return False

        finally:
            self.stop_mkdocs_server()

    def print_summary(self):
        """Print test summary"""
        print("\n" + "="*70)
        print("TEST SUMMARY")
        print("="*70)
        print(f"Screenshots saved to: {self.screenshots_dir}")
        print("\nTest Results:")
        for result in self.results:
            print(f"  • {result}")
        print("="*70 + "\n")


async def main():
    """Main entry point"""
    tester = MkDocsUITester()
    success = await tester.run_all_tests()
    tester.print_summary()

    return 0 if success else 1


if __name__ == "__main__":
    exit_code = asyncio.run(main())
    sys.exit(exit_code)
