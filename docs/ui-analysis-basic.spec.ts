import { test, expect } from '@playwright/test';

test.describe('Basic UI Analysis', () => {
  test('should load main page and check basic elements', async ({ page }) => {
    // Start from the documentation site
    await page.goto('http://localhost:3000');

    // Check page title
    const title = await page.title();
    console.log(`Page title: ${title}`);
    expect(title).toContain('MoAI-ADK');

    // Check main heading
    const h1 = page.locator('h1').first();
    await expect(h1).toBeVisible();
    const h1Text = await h1.textContent();
    console.log(`Main heading: ${h1Text}`);

    // Check navigation elements
    const nav = page.locator('nav').first();
    if (await nav.count() > 0) {
      console.log('Navigation found');
      const navLinks = nav.locator('a');
      const linkCount = await navLinks.count();
      console.log(`Navigation links: ${linkCount}`);

      for (let i = 0; i < Math.min(linkCount, 5); i++) {
        const linkText = await navLinks.nth(i).textContent();
        const linkHref = await navLinks.nth(i).getAttribute('href');
        console.log(`  Link ${i + 1}: ${linkText} -> ${linkHref}`);
      }
    }

    // Check main content
    const main = page.locator('main').first();
    if (await main.count() === 0) {
      const mainContent = page.locator('[role="main"], .content, .nextra-container').first();
      await expect(mainContent).toBeVisible();
      console.log('Main content area found');
    } else {
      await expect(main).toBeVisible();
      console.log('Main element found');
    }

    // Check language indicators
    const langSelectors = [
      'html[lang]',
      '[data-language]',
      '.language-selector'
    ];

    for (const selector of langSelectors) {
      const element = page.locator(selector);
      if (await element.count() > 0) {
        console.log(`Language element found: ${selector}`);
        break;
      }
    }

    // Take a screenshot
    await page.screenshot({
      path: 'test-results/basic-ui-analysis.png',
      fullPage: true
    });

    // Check responsive design by resizing viewport
    await page.setViewportSize({ width: 375, height: 667 }); // Mobile
    await page.waitForTimeout(1000);

    await page.screenshot({
      path: 'test-results/mobile-view.png',
      fullPage: true
    });

    console.log('Mobile viewport screenshot captured');

    // Back to desktop
    await page.setViewportSize({ width: 1280, height: 720 });
    await page.waitForTimeout(1000);
  });

  test('should check color contrast and accessibility basics', async ({ page }) => {
    await page.goto('http://localhost:3000');

    // Get computed styles for body
    const bodyStyles = await page.locator('body').evaluate((el) => {
      const styles = window.getComputedStyle(el);
      return {
        backgroundColor: styles.backgroundColor,
        color: styles.color,
        fontSize: styles.fontSize,
        fontFamily: styles.fontFamily
      };
    });

    console.log('Body styles:', bodyStyles);

    // Check text color contrast (basic check)
    const textColor = bodyStyles.color;
    const bgColor = bodyStyles.backgroundColor;

    console.log(`Text color: ${textColor}`);
    console.log(`Background color: ${bgColor}`);

    // Check for proper heading structure
    const headings = await page.locator('h1, h2, h3, h4, h5, h6').all();
    console.log(`Total headings: ${headings.length}`);

    for (let i = 0; i < Math.min(headings.length, 5); i++) {
      const tagName = await headings[i].evaluate(el => el.tagName);
      const text = await headings[i].textContent();
      console.log(`  ${tagName}: ${text?.substring(0, 50)}...`);
    }

    // Check for focusable elements
    const focusableElements = await page.locator('a, button, input, select, textarea').all();
    console.log(`Focusable elements: ${focusableElements.length}`);

    // Test keyboard navigation
    await page.keyboard.press('Tab');
    const focusedElement = page.locator(':focus');
    expect(await focusedElement.count()).toBeGreaterThan(0);

    const focusedTag = await focusedElement.evaluate(el => el.tagName);
    console.log(`First focused element: ${focusedTag}`);

    await page.screenshot({
      path: 'test-results/keyboard-focus.png',
      fullPage: true
    });
  });

  test('should check performance and loading', async ({ page }) => {
    const startTime = Date.now();

    // Listen for network requests
    const requests: string[] = [];
    page.on('request', request => {
      requests.push(request.url());
    });

    await page.goto('http://localhost:3000', { waitUntil: 'networkidle' });

    const loadTime = Date.now() - startTime;
    console.log(`Page load time: ${loadTime}ms`);

    // Check number of resources loaded
    console.log(`Total requests: ${requests.length}`);

    // Analyze request types
    const cssRequests = requests.filter(url => url.includes('.css'));
    const jsRequests = requests.filter(url => url.includes('.js'));
    const imageRequests = requests.filter(url => url.match(/\.(jpg|jpeg|png|gif|webp|svg)$/i));

    console.log(`CSS files: ${cssRequests.length}`);
    console.log(`JS files: ${jsRequests.length}`);
    console.log(`Images: ${imageRequests.length}`);

    // Check if page loads within acceptable time
    expect(loadTime).toBeLessThan(5000); // 5 seconds max

    // Get page dimensions
    const dimensions = await page.evaluate(() => ({
      width: document.documentElement.scrollWidth,
      height: document.documentElement.scrollHeight,
      viewport: {
        width: window.innerWidth,
        height: window.innerHeight
      }
    }));

    console.log('Page dimensions:', dimensions);

    // Check for scroll performance
    if (dimensions.height > dimensions.viewport.height * 1.5) {
      console.log('Page is scrollable, testing scroll performance');

      const scrollStart = Date.now();
      await page.evaluate(() => {
        window.scrollTo({ top: document.body.scrollHeight / 2, behavior: 'smooth' });
      });

      await page.waitForTimeout(1000);
      const scrollTime = Date.now() - scrollStart;

      console.log(`Scroll performance time: ${scrollTime}ms`);
    }
  });
});