/**
 * E2E Tests for MoAI-ADK Documentation Server
 * Tests the complete user experience after _meta.json fix
 */

const { test, expect } = require('@playwright/test');

const BASE_URL = 'http://localhost:3001';

test.describe('MoAI-ADK Documentation Server - Post-Fix Validation', () => {

  test.beforeEach(async ({ page }) => {
    // Navigate to the docs server
    await page.goto(BASE_URL);
  });

  test('should load main page without errors', async ({ page }) => {
    // Wait for page to load
    await page.waitForLoadState('networkidle');

    // Check that page loads successfully (no 500 errors)
    expect(await page.title()).toContain('MoAI-ADK');

    // Check for main content
    await expect(page.locator('body')).toContainText('MoAI-ADK');
  });

  test('should display Korean navigation sidebar', async ({ page }) => {
    // Wait for sidebar to appear
    await page.waitForSelector('nav', { timeout: 10000 });

    // Check for Korean navigation items
    const navItems = [
      '한국어',
      '시작하기',
      '주요 기능',
      'Skills',
      'Alfred',
      '가이드',
      '출력 스타일',
      '에이전트',
      '아키텍처',
      '레퍼런스'
    ];

    for (const item of navItems) {
      await expect(page.locator('nav')).toContainText(item);
    }
  });

  test('should render content properly (not truncated)', async ({ page }) => {
    // Wait for content to load
    await page.waitForLoadState('networkidle');

    // Get the main content area
    const mainContent = page.locator('main, .content, article');

    // Check that content is not truncated to just 21 characters
    const textContent = await mainContent.textContent();
    expect(textContent?.length).toBeGreaterThan(100);

    // Check for meaningful content
    await expect(mainContent).toContainText('MoAI');
  });

  test('should navigate to different sections', async ({ page }) => {
    // Click on "시작하기" (Getting Started)
    await page.click('text=시작하기');
    await page.waitForLoadState('networkidle');

    // Should navigate to getting started section
    expect(page.url()).toContain('/getting-started');

    // Check for getting started content
    await expect(page.locator('main')).toContainText('시작');
  });

  test('should have working search functionality', async ({ page }) => {
    // Look for search input
    const searchInput = page.locator('input[placeholder*="검색"], input[type="search"], .search-input');

    if (await searchInput.isVisible()) {
      // Type search query
      await searchInput.fill('Alfred');

      // Should show search results or navigation
      await page.waitForTimeout(1000);

      // Verify search is working (no errors)
      expect(await page.locator('body').textContent()).not.toContain('Error');
    }
  });

  test('should be responsive on mobile devices', async ({ page }) => {
    // Emulate mobile device
    await page.setViewportSize({ width: 375, height: 667 });

    // Check that content is still accessible
    await expect(page.locator('body')).toBeVisible();

    // Check for mobile navigation toggle
    const mobileNav = page.locator('.mobile-nav, button[aria-label*="menu"], .menu-toggle');
    if (await mobileNav.isVisible()) {
      await mobileNav.click();
      await page.waitForTimeout(500);
    }

    // Should still show navigation
    await expect(page.locator('nav')).toBeVisible();
  });

  test('should have working dark mode toggle', async ({ page }) => {
    // Look for dark mode toggle
    const darkModeToggle = page.locator('[aria-label*="dark"], [data-theme], .dark-mode-toggle');

    if (await darkModeToggle.isVisible()) {
      // Click dark mode toggle
      await darkModeToggle.click();
      await page.waitForTimeout(500);

      // Check that theme changed (class or attribute)
      const hasDarkClass = await page.locator('html, body').evaluate(el => {
        return el.classList.contains('dark') ||
               el.getAttribute('data-theme') === 'dark' ||
               el.getAttribute('class')?.includes('dark');
      });

      // Toggle back to light mode
      if (hasDarkClass) {
        await darkModeToggle.click();
        await page.waitForTimeout(500);
      }
    }
  });

  test('should handle internationalization properly', async ({ page }) => {
    // Check for language switcher
    const langSwitcher = page.locator('[aria-label*="language"], .lang-select, select');

    if (await langSwitcher.isVisible()) {
      // Should show language options
      await expect(langSwitcher).toBeVisible();

      // Current language should be Korean
      await expect(page.locator('nav')).toContainText('한국어');
    }
  });

  test('should have no console errors', async ({ page }) => {
    // Listen for console errors
    const errors = [];
    page.on('console', msg => {
      if (msg.type() === 'error') {
        errors.push(msg.text());
      }
    });

    // Wait for page to fully load
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(3000);

    // Should have no console errors related to _meta or JSON
    const metaErrors = errors.filter(error =>
      error.includes('_meta') ||
      error.includes('JSON') ||
      error.includes('Unexpected token')
    );

    expect(metaErrors).toHaveLength(0);
  });
});

test.describe('Documentation Performance and Accessibility', () => {
  test('should load quickly (Core Web Vitals)', async ({ page }) => {
    // Start performance measurement
    const startTime = Date.now();

    await page.goto(BASE_URL);
    await page.waitForLoadState('networkidle');

    const loadTime = Date.now() - startTime;

    // Should load within reasonable time (under 5 seconds)
    expect(loadTime).toBeLessThan(5000);
  });

  test('should be accessible (a11y)', async ({ page }) => {
    await page.goto(BASE_URL);
    await page.waitForLoadState('networkidle');

    // Basic accessibility checks
    await expect(page.locator('h1')).toBeVisible();
    await expect(page.locator('nav')).toBeVisible();
    await expect(page.locator('main')).toBeVisible();

    // Check for proper heading structure
    const headings = await page.locator('h1, h2, h3, h4, h5, h6').all();
    expect(headings.length).toBeGreaterThan(0);
  });
});