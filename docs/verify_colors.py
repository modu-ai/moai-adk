#!/usr/bin/env python3
"""Verify color theme consistency."""

import asyncio
from playwright.async_api import async_playwright


async def verify_colors():
    """Verify color theme consistency across light and dark modes."""

    base_url = "http://localhost:8000"

    async with async_playwright() as p:
        browser = await p.chromium.launch()

        # Test light mode
        print("Verifying Light Mode Colors...")
        page = await browser.new_page(viewport={"width": 1920, "height": 1080})
        await page.goto(base_url, wait_until="networkidle")

        # Get computed colors from root element
        light_colors = await page.evaluate("""() => {
            const root = document.documentElement;
            const style = getComputedStyle(root);
            return {
                primary: style.getPropertyValue('--md-primary-fg-color').trim(),
                accent: style.getPropertyValue('--md-accent-fg-color').trim(),
                bg: style.getPropertyValue('--md-default-bg-color').trim() || window.getComputedStyle(document.body).backgroundColor,
            };
        }""")

        print(f"  Primary Text: {light_colors['primary']}")
        print(f"  Accent: {light_colors['accent']}")
        print(f"  Background: {light_colors['bg']}")

        # Take screenshot of current mode
        await page.screenshot(path="/Users/goos/MoAI/MoAI-ADK/docs/.ui-review-updated/light_mode.png")
        print("  Screenshot saved: light_mode.png")

        await page.close()

        # Test dark mode by toggling theme
        print("\nVerifying Dark Mode Colors...")
        page = await browser.new_page(viewport={"width": 1920, "height": 1080})
        await page.goto(base_url, wait_until="networkidle")

        # Find and click the theme toggle button
        try:
            # Try to find the theme toggle
            toggle = await page.query_selector("button[aria-label*='Dark'], button[aria-label*='Light']")
            if toggle:
                await toggle.click()
                await page.wait_for_timeout(500)  # Wait for transition
        except:
            print("  Could not toggle theme automatically")

        dark_colors = await page.evaluate("""() => {
            const root = document.documentElement;
            const style = getComputedStyle(root);
            return {
                primary: style.getPropertyValue('--md-primary-fg-color').trim(),
                accent: style.getPropertyValue('--md-accent-fg-color').trim(),
                bg: style.getPropertyValue('--md-default-bg-color').trim() || window.getComputedStyle(document.body).backgroundColor,
            };
        }""")

        print(f"  Primary Text: {dark_colors['primary']}")
        print(f"  Accent: {dark_colors['accent']}")
        print(f"  Background: {dark_colors['bg']}")

        await page.screenshot(path="/Users/goos/MoAI/MoAI-ADK/docs/.ui-review-updated/dark_mode.png")
        print("  Screenshot saved: dark_mode.png")

        await page.close()
        await browser.close()

        print("\nColor verification complete!")
        print("Screenshots saved to /Users/goos/MoAI/MoAI-ADK/docs/.ui-review-updated/")


if __name__ == "__main__":
    asyncio.run(verify_colors())
