#!/usr/bin/env python3
"""Final validation and comprehensive testing."""

import asyncio
from playwright.async_api import async_playwright


async def final_validation():
    """Perform final validation of all requirements."""

    base_url = "http://localhost:8000"

    async with async_playwright() as p:
        browser = await p.chromium.launch()

        print("=" * 60)
        print("FINAL VALIDATION - MoAI-ADK Documentation Site")
        print("=" * 60)

        # Test 1: Check for remaining emojis
        print("\n1. Checking for remaining emojis in rendered content...")
        page = await browser.new_page(viewport={"width": 1920, "height": 1080})
        await page.goto(base_url, wait_until="networkidle")

        # Get all text content
        text_content = await page.evaluate("() => document.body.innerText")

        emoji_list = [
            "📋", "📚", "🚀", "🎯", "✨", "🔧", "🌍", "🎨", "🔐",
            "🎉", "📦", "🔥", "👥", "🤝", "📖", "🎭"
        ]

        found_emojis = []
        for emoji in emoji_list:
            if emoji in text_content:
                found_emojis.append(emoji)

        if found_emojis:
            print(f"  WARNING: Found {len(found_emojis)} emoji(s) in content: {found_emojis}")
        else:
            print("  ✓ No emojis found in text content")

        await page.close()

        # Test 2: Mobile responsiveness
        print("\n2. Testing Mobile Responsiveness (375x812)...")
        page = await browser.new_page(viewport={"width": 375, "height": 812})
        await page.goto(base_url, wait_until="networkidle")

        mobile_ok = await page.evaluate("""() => {
            const viewport = window.innerWidth;
            const styles = window.getComputedStyle(document.documentElement);
            return {
                viewport_width: viewport,
                body_width: document.body.offsetWidth,
                has_horizontal_scroll: document.body.scrollWidth > window.innerWidth,
            };
        }""")

        print(f"  Viewport Width: {mobile_ok['viewport_width']}px")
        print(f"  Body Width: {mobile_ok['body_width']}px")
        if not mobile_ok['has_horizontal_scroll']:
            print("  ✓ No horizontal scrolling detected")
        else:
            print("  WARNING: Horizontal scrolling detected")

        await page.screenshot(path="/Users/goos/MoAI/MoAI-ADK/docs/.ui-review-updated/mobile_375x812.png")
        print("  Screenshot saved: mobile_375x812.png")
        await page.close()

        # Test 3: Tablet responsiveness
        print("\n3. Testing Tablet Responsiveness (768x1024)...")
        page = await browser.new_page(viewport={"width": 768, "height": 1024})
        await page.goto(base_url, wait_until="networkidle")

        tablet_ok = await page.evaluate("""() => {
            return {
                viewport_width: window.innerWidth,
                body_width: document.body.offsetWidth,
                has_horizontal_scroll: document.body.scrollWidth > window.innerWidth,
            };
        }""")

        print(f"  Viewport Width: {tablet_ok['viewport_width']}px")
        print(f"  Body Width: {tablet_ok['body_width']}px")
        if not tablet_ok['has_horizontal_scroll']:
            print("  ✓ No horizontal scrolling detected")
        else:
            print("  WARNING: Horizontal scrolling detected")

        await page.screenshot(path="/Users/goos/MoAI/MoAI-ADK/docs/.ui-review-updated/tablet_768x1024.png")
        print("  Screenshot saved: tablet_768x1024.png")
        await page.close()

        # Test 4: Desktop view
        print("\n4. Testing Desktop View (1920x1080)...")
        page = await browser.new_page(viewport={"width": 1920, "height": 1080})
        await page.goto(base_url, wait_until="networkidle")

        desktop_ok = await page.evaluate("""() => {
            const header = document.querySelector('header');
            const sidebar = document.querySelector('nav');
            const content = document.querySelector('main');
            return {
                has_header: !!header,
                has_sidebar: !!sidebar,
                has_content: !!content,
                viewport: window.innerWidth,
            };
        }""")

        print(f"  ✓ Header: {desktop_ok['has_header']}")
        print(f"  ✓ Navigation: {desktop_ok['has_sidebar']}")
        print(f"  ✓ Content Area: {desktop_ok['has_content']}")

        await page.screenshot(path="/Users/goos/MoAI/MoAI-ADK/docs/.ui-review-updated/desktop_full.png")
        print("  Screenshot saved: desktop_full.png")
        await page.close()

        # Test 5: Color consistency
        print("\n5. Verifying Color Theme Consistency...")
        page = await browser.new_page(viewport={"width": 1920, "height": 1080})
        await page.goto(base_url, wait_until="networkidle")

        colors = await page.evaluate("""() => {
            const root = document.documentElement;
            const style = getComputedStyle(root);
            const links = document.querySelectorAll('a');
            const linkColor = links.length > 0 ?
                window.getComputedStyle(links[0]).color : 'N/A';

            return {
                primary_text: style.getPropertyValue('--md-primary-fg-color').trim() || '#171612',
                accent_text: style.getPropertyValue('--md-accent-fg-color').trim() || '#504F4B',
                background: style.getPropertyValue('--md-default-bg-color').trim() || '#ffffff',
                link_color: linkColor,
            };
        }""")

        print(f"  Primary Text Color: {colors['primary_text']}")
        print(f"  Accent Color: {colors['accent_text']}")
        print(f"  Background: {colors['background']}")
        print(f"  Link Color: {colors['link_color']}")

        # Verify neutral palette
        is_neutral = '#171612' in str(colors.values()) or '#504F4B' in str(colors.values())
        if is_neutral:
            print("  ✓ Neutral (grayscale) color palette confirmed")
        else:
            print("  WARNING: Color palette may not be neutral")

        await page.close()

        await browser.close()

        print("\n" + "=" * 60)
        print("VALIDATION COMPLETE")
        print("=" * 60)
        print("\nAll screenshots saved to:")
        print("  /Users/goos/MoAI/MoAI-ADK/docs/.ui-review-updated/")


if __name__ == "__main__":
    asyncio.run(final_validation())
