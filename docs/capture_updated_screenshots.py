#!/usr/bin/env python3
"""Capture updated screenshots of MoAI-ADK documentation site."""

import asyncio
from pathlib import Path
from playwright.async_api import async_playwright


async def capture_updated_screenshots():
    """Capture updated screenshots after modifications."""

    base_url = "http://localhost:8000"
    screenshots_dir = Path("/Users/goos/MoAI/MoAI-ADK/docs/.ui-review-updated")
    screenshots_dir.mkdir(exist_ok=True)

    pages_to_capture = [
        ("home-updated", "/"),
        ("getting-started-updated", "/getting-started/installation/"),
    ]

    async with async_playwright() as p:
        browser = await p.chromium.launch()

        # Desktop view (1920x1080)
        print("Capturing updated desktop views...")
        page = await browser.new_page(viewport={"width": 1920, "height": 1080})

        for name, path in pages_to_capture:
            try:
                url = f"{base_url}{path}"
                # Clear cache to get fresh content
                await page.context.clear_cookies()
                await page.goto(url, wait_until="networkidle")
                await page.wait_for_load_state("domcontentloaded")

                screenshot_path = screenshots_dir / f"{name}_full.png"
                await page.screenshot(path=str(screenshot_path), full_page=True)
                print(f"  - {name}: {screenshot_path}")
            except Exception as e:
                print(f"  - {name}: ERROR - {e}")

        await page.close()
        await browser.close()
        print(f"\nUpdated screenshots saved to {screenshots_dir}")


if __name__ == "__main__":
    asyncio.run(capture_updated_screenshots())
