#!/usr/bin/env python3
"""Capture screenshots of MoAI-ADK documentation site for UX analysis."""

import asyncio
import os
from pathlib import Path
from playwright.async_api import async_playwright


async def capture_screenshots():
    """Capture screenshots of various pages."""

    base_url = "http://localhost:8000"
    screenshots_dir = Path("/Users/goos/MoAI/MoAI-ADK/docs/.ui-review")
    screenshots_dir.mkdir(exist_ok=True)

    pages_to_capture = [
        ("home", "/"),
        ("getting-started", "/getting-started/installation/"),
        ("alfred-workflow", "/guides/alfred/"),
        ("quick-start", "/getting-started/quick-start/"),
        ("cli-reference", "/reference/cli/"),
        ("contributing", "/contributing/"),
    ]

    async with async_playwright() as p:
        browser = await p.chromium.launch()

        # Desktop view (1920x1080)
        print("Capturing desktop views (1920x1080)...")
        page = await browser.new_page(viewport={"width": 1920, "height": 1080})

        for name, path in pages_to_capture:
            try:
                url = f"{base_url}{path}"
                await page.goto(url, wait_until="networkidle")
                await page.wait_for_load_state("domcontentloaded")

                # Take full page screenshot
                screenshot_path = screenshots_dir / f"{name}_desktop_full.png"
                await page.screenshot(path=str(screenshot_path), full_page=True)
                print(f"  - {name}: {screenshot_path}")
            except Exception as e:
                print(f"  - {name}: ERROR - {e}")

        await page.close()

        # Mobile view (375x812)
        print("\nCapturing mobile views (375x812)...")
        page = await browser.new_page(viewport={"width": 375, "height": 812})

        for name, path in pages_to_capture[:2]:  # Just first 2 for mobile
            try:
                url = f"{base_url}{path}"
                await page.goto(url, wait_until="networkidle")
                await page.wait_for_load_state("domcontentloaded")

                screenshot_path = screenshots_dir / f"{name}_mobile_full.png"
                await page.screenshot(path=str(screenshot_path), full_page=True)
                print(f"  - {name}: {screenshot_path}")
            except Exception as e:
                print(f"  - {name}: ERROR - {e}")

        await page.close()

        # Tablet view (768x1024)
        print("\nCapturing tablet views (768x1024)...")
        page = await browser.new_page(viewport={"width": 768, "height": 1024})

        for name, path in pages_to_capture[:2]:  # Just first 2 for tablet
            try:
                url = f"{base_url}{path}"
                await page.goto(url, wait_until="networkidle")
                await page.wait_for_load_state("domcontentloaded")

                screenshot_path = screenshots_dir / f"{name}_tablet_full.png"
                await page.screenshot(path=str(screenshot_path), full_page=True)
                print(f"  - {name}: {screenshot_path}")
            except Exception as e:
                print(f"  - {name}: ERROR - {e}")

        await page.close()

        await browser.close()
        print(f"\nScreenshots saved to {screenshots_dir}")


if __name__ == "__main__":
    asyncio.run(capture_screenshots())
