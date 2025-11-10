import { test, expect, devices } from '@playwright/test';

test.describe('Comprehensive UI/UX Manual Analysis', () => {
  let baseUrl: string;

  test.beforeAll(async () => {
    baseUrl = 'http://localhost:3000';
  });

  test('should capture comprehensive UI analysis data', async ({ page }) => {
    // Go to the documentation site
    await page.goto(baseUrl, { waitUntil: 'networkidle' });

    // 1. BASIC PAGE ANALYSIS
    console.log('\n=== BASIC PAGE ANALYSIS ===');

    const title = await page.title();
    console.log(`Page Title: "${title}"`);

    const url = page.url();
    console.log(`Current URL: ${url}`);

    // 2. VISUAL DESIGN ANALYSIS
    console.log('\n=== VISUAL DESIGN ANALYSIS ===');

    // Get body styles
    const bodyStyles = await page.locator('body').evaluate((el) => {
      const styles = window.getComputedStyle(el);
      return {
        backgroundColor: styles.backgroundColor,
        color: styles.color,
        fontSize: styles.fontSize,
        fontFamily: styles.fontFamily,
        lineHeight: styles.lineHeight,
        fontWeight: styles.fontWeight
      };
    });
    console.log('Body Styles:', bodyStyles);

    // Check headings
    const headings = await page.locator('h1, h2, h3, h4, h5, h6').all();
    console.log(`Total headings: ${headings.length}`);

    for (let i = 0; i < Math.min(headings.length, 10); i++) {
      const heading = headings[i];
      const tagName = await heading.evaluate(el => el.tagName);
      const text = await heading.textContent();
      const styles = await heading.evaluate(el => {
        const styles = window.getComputedStyle(el);
        return {
          fontSize: styles.fontSize,
          fontWeight: styles.fontWeight,
          color: styles.color
        };
      });

      console.log(`${tagName}: "${text?.substring(0, 50)}..."`);
      console.log(`  Style: ${styles.fontSize}, ${styles.fontWeight}, ${styles.color}`);
    }

    // 3. NAVIGATION ANALYSIS
    console.log('\n=== NAVIGATION ANALYSIS ===');

    // Find navigation elements
    const navSelectors = [
      'nav', '.navigation', '.nav', '[role="navigation"]',
      '.sidebar', '.menu', '.header'
    ];

    let navigationElement = null;
    for (const selector of navSelectors) {
      const element = page.locator(selector);
      if (await element.count() > 0) {
        navigationElement = element.first();
        console.log(`Navigation found: ${selector}`);
        break;
      }
    }

    if (navigationElement) {
      const navLinks = navigationElement.locator('a');
      const linkCount = await navLinks.count();
      console.log(`Navigation links: ${linkCount}`);

      // Show first few navigation links
      for (let i = 0; i < Math.min(linkCount, 8); i++) {
        const link = navLinks.nth(i);
        const text = await link.textContent();
        const href = await link.getAttribute('href');
        console.log(`  ${i + 1}. "${text?.trim()}" -> ${href}`);
      }
    }

    // 4. LANGUAGE SWITCHING ANALYSIS
    console.log('\n=== LANGUAGE SWITCHING ANALYSIS ===');

    // Check for language indicators
    const langSelectors = [
      'html[lang]', '[data-language]', '.language-selector',
      '.lang', '[aria-label*="language"]', 'select'
    ];

    for (const selector of langSelectors) {
      const element = page.locator(selector);
      if (await element.count() > 0) {
        const attr = await element.first().getAttribute(selector.includes('[') ? selector.match(/(\w+)=/)![1] : 'class');
        console.log(`Language element: ${selector} = ${attr}`);
      }
    }

    // Look for language-specific links
    const languageLinks = page.locator('a[href*="/ko/"], a[href*="/en/"], a[href*="/ja/"], a[href*="/zh/"]');
    const langLinkCount = await languageLinks.count();
    if (langLinkCount > 0) {
      console.log(`Language-specific links: ${langLinkCount}`);
      for (let i = 0; i < langLinkCount; i++) {
        const href = await languageLinks.nth(i).getAttribute('href');
        console.log(`  - ${href}`);
      }
    }

    // 5. SEARCH FUNCTIONALITY
    console.log('\n=== SEARCH FUNCTIONALITY ===');

    const searchSelectors = [
      'input[type="search"]', 'input[placeholder*="search" i]',
      'input[placeholder*="검색" i]', '.search input',
      '[role="search"] input', '#search'
    ];

    let searchInput = null;
    for (const selector of searchSelectors) {
      const element = page.locator(selector);
      if (await element.count() > 0) {
        searchInput = element.first();
        console.log(`Search input found: ${selector}`);
        break;
      }
    }

    if (searchInput) {
      const placeholder = await searchInput.getAttribute('placeholder');
      console.log(`Search placeholder: "${placeholder}"`);
    }

    // 6. ACCESSIBILITY ANALYSIS
    console.log('\n=== ACCESSIBILITY ANALYSIS ===');

    // Check for semantic landmarks
    const landmarks = [
      { selector: 'main', name: 'Main content' },
      { selector: 'nav', name: 'Navigation' },
      { selector: 'header', name: 'Header' },
      { selector: 'footer', name: 'Footer' },
      { selector: '[role="main"]', name: 'Main (role)' },
      { selector: '[role="navigation"]', name: 'Navigation (role)' },
      { selector: '[role="banner"]', name: 'Banner (role)' },
      { selector: '[role="contentinfo"]', name: 'Contentinfo (role)' }
    ];

    for (const { selector, name } of landmarks) {
      const element = page.locator(selector);
      const count = await element.count();
      if (count > 0) {
        console.log(`✓ ${name}: ${count} found`);
      }
    }

    // Check for focusable elements
    const focusableElements = await page.locator(
      'a, button, input, select, textarea, [tabindex]:not([tabindex="-1"])'
    ).all();
    console.log(`Focusable elements: ${focusableElements.length}`);

    // Test keyboard navigation
    console.log('Testing keyboard navigation...');
    let focusedCount = 0;

    for (let i = 0; i < 5; i++) {
      await page.keyboard.press('Tab');
      const focusedElement = page.locator(':focus');
      const hasFocus = await focusedElement.count() > 0;

      if (hasFocus) {
        focusedCount++;
        const tagName = await focusedElement.evaluate(el => el.tagName);
        const text = await focusedElement.evaluate(el => el.textContent?.substring(0, 30));
        console.log(`  Tab ${i + 1}: ${tagName} - "${text}"`);
      } else {
        console.log(`  Tab ${i + 1}: No focus`);
      }
    }

    console.log(`Successful keyboard navigation: ${focusedCount}/5`);

    // 7. PERFORMANCE METRICS
    console.log('\n=== PERFORMANCE METRICS ===');

    // Resource loading analysis
    const resources = await page.evaluate(() => {
      return performance.getEntriesByType('resource').map(entry => ({
        name: entry.name.split('/').pop(),
        type: entry.initiatorType,
        duration: entry.duration,
        size: entry.transferSize || 0
      }));
    });

    console.log(`Total resources: ${resources.length}`);

    const jsFiles = resources.filter(r => r.type === 'script');
    const cssFiles = resources.filter(r => r.type === 'stylesheet');
    const images = resources.filter(r => r.type === 'img');

    console.log(`JavaScript files: ${jsFiles.length}`);
    console.log(`CSS files: ${cssFiles.length}`);
    console.log(`Images: ${images.length}`);

    // Calculate total sizes
    const totalJS = jsFiles.reduce((sum, file) => sum + file.size, 0);
    const totalCSS = cssFiles.reduce((sum, file) => sum + file.size, 0);
    const totalImages = images.reduce((sum, file) => sum + file.size, 0);

    console.log(`JS total size: ${(totalJS / 1024).toFixed(2)} KB`);
    console.log(`CSS total size: ${(totalCSS / 1024).toFixed(2)} KB`);
    console.log(`Images total size: ${(totalImages / 1024).toFixed(2)} KB`);

    // 8. CONTENT ANALYSIS
    console.log('\n=== CONTENT ANALYSIS ===');

    // Check for code blocks
    const codeBlocks = page.locator('pre code, .code-block, [class*="code"], .highlight');
    const codeBlockCount = await codeBlocks.count();
    console.log(`Code blocks: ${codeBlockCount}`);

    // Check for tables
    const tables = page.locator('table');
    const tableCount = await tables.count();
    console.log(`Tables: ${tableCount}`);

    // Check for lists
    const lists = page.locator('ul, ol');
    const listCount = await lists.count();
    console.log(`Lists: ${listCount}`);

    // Check for blockquotes
    const blockquotes = page.locator('blockquote');
    const blockquoteCount = await blockquotes.count();
    console.log(`Blockquotes: ${blockquoteCount}`);

    // 9. SCREENSHOTS
    console.log('\n=== CAPTURING SCREENSHOTS ===');

    // Desktop view
    await page.setViewportSize({ width: 1280, height: 720 });
    await page.screenshot({
      path: 'test-results/desktop-full-view.png',
      fullPage: true
    });
    console.log('✓ Desktop screenshot captured');

    // Tablet view
    await page.setViewportSize({ width: 768, height: 1024 });
    await page.waitForTimeout(1000);
    await page.screenshot({
      path: 'test-results/tablet-view.png',
      fullPage: true
    });
    console.log('✓ Tablet screenshot captured');

    // Mobile view
    await page.setViewportSize({ width: 375, height: 667 });
    await page.waitForTimeout(1000);
    await page.screenshot({
      path: 'test-results/mobile-view.png',
      fullPage: true
    });
    console.log('✓ Mobile screenshot captured');

    // Test mobile navigation
    const mobileMenuButton = page.locator('button[aria-expanded], .hamburger, .menu-toggle, button[aria-label*="menu"]');
    if (await mobileMenuButton.count() > 0) {
      await mobileMenuButton.first().click();
      await page.waitForTimeout(1000);

      await page.screenshot({
        path: 'test-results/mobile-menu-open.png',
        fullPage: true
      });
      console.log('✓ Mobile menu screenshot captured');
    }

    // 10. RESPONSIVE DESIGN TEST
    console.log('\n=== RESPONSIVE DESIGN TEST ===');

    const viewports = [
      { name: 'Desktop', width: 1280, height: 720 },
      { name: 'Laptop', width: 1024, height: 768 },
      { name: 'Tablet', width: 768, height: 1024 },
      { name: 'Mobile', width: 375, height: 667 }
    ];

    for (const viewport of viewports) {
      await page.setViewportSize({ width: viewport.width, height: viewport.height });
      await page.waitForTimeout(500);

      // Check for horizontal scrolling
      const hasHorizontalScroll = await page.evaluate(() => {
        return document.body.scrollWidth > window.innerWidth;
      });

      console.log(`${viewport.name} (${viewport.width}x${viewport.height}): ${hasHorizontalScroll ? '❌ Horizontal scroll' : '✓ No horizontal scroll'}`);
    }

    console.log('\n=== ANALYSIS COMPLETE ===');
  });
});