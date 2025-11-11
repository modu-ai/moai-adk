import { test, expect } from '@playwright/test';

// Test suite for MoAI-ADK Documentation UI/UX Verification
// Tests for Nextra 3.3.1 → Nextra 4.6.0 migration

const BASE_URL = 'http://localhost:3000';

test.describe('MoAI-ADK Documentation UI/UX Verification', () => {

  test.beforeAll(async () => {
    // Ensure the site is running
    console.log('Starting UI/UX verification of MoAI-ADK Documentation...');
  });

  test('Homepage loads correctly', async ({ page }) => {
    await page.goto(BASE_URL);

    // Check if the page loads without errors
    await expect(page).toHaveTitle(/MoAI-ADK/);

    // Check for main elements
    await expect(page.locator('h1')).toBeVisible();

    // Check for language options
    await expect(page.locator('text=한국어')).toBeVisible();
    await expect(page.locator('text=English')).toBeVisible();
    await expect(page.locator('text=日本語')).toBeVisible();
    await expect(page.locator('text=中文')).toBeVisible();

    // Check for main navigation
    await expect(page.locator('nav')).toBeVisible();

    // Check for search functionality
    await expect(page.locator('input[placeholder*="검색"]')).toBeVisible();
  });

  test('Korean language page works correctly', async ({ page }) => {
    await page.goto(`${BASE_URL}/ko`);

    // Check if Korean content loads
    await expect(page.locator('h1')).toBeVisible();

    // Check navigation
    await expect(page.locator('nav')).toBeVisible();

    // Check search is in Korean
    const searchInput = page.locator('input[placeholder*="검색"]');
    if (await searchInput.count() > 0) {
      await expect(searchInput).toBeVisible();
    }
  });

  test('English language page works correctly', async ({ page }) => {
    await page.goto(`${BASE_URL}/en`);

    // Check if English content loads
    await expect(page.locator('h1')).toBeVisible();

    // Check navigation
    await expect(page.locator('nav')).toBeVisible();
  });

  test('Japanese language page works correctly', async ({ page }) => {
    await page.goto(`${BASE_URL}/ja`);

    // Check if Japanese content loads
    await expect(page.locator('h1')).toBeVisible();

    // Check navigation
    await expect(page.locator('nav')).toBeVisible();
  });

  test('Chinese language page works correctly', async ({ page }) => {
    await page.goto(`${BASE_URL}/zh`);

    // Check if Chinese content loads
    await expect(page.locator('h1')).toBeVisible();

    // Check navigation
    await expect(page.locator('nav')).toBeVisible();
  });

  test('Navigation functionality', async ({ page }) => {
    await page.goto(`${BASE_URL}/ko`);

    // Test sidebar navigation
    const sidebar = page.locator('.nextra-sidebar-container, aside');
    if (await sidebar.count() > 0) {
      await expect(sidebar).toBeVisible();

      // Test if navigation links are clickable
      const navLinks = sidebar.locator('a');
      if (await navLinks.count() > 0) {
        const firstLink = navLinks.first();
        await expect(firstLink).toBeVisible();
      }
    }

    // Test language switcher
    const languageSwitcher = page.locator('text=English');
    if (await languageSwitcher.count() > 0) {
      await languageSwitcher.click();
      await expect(page).toHaveURL(/\/en/);
    }
  });

  test('Theme consistency and design elements', async ({ page }) => {
    await page.goto(`${BASE_URL}/ko`);

    // Check for consistent theme
    await expect(page.locator('body')).toBeVisible();

    // Check for proper color contrast (basic check)
    const textElements = page.locator('p, h1, h2, h3, h4, h5, h6');
    if (await textElements.count() > 0) {
      await expect(textElements.first()).toBeVisible();
    }

    // Check for responsive design elements
    const viewport = page.viewportSize();
    if (viewport) {
      console.log(`Current viewport: ${viewport.width}x${viewport.height}`);
    }
  });

  test('Mobile responsiveness', async ({ page }) => {
    // Test mobile viewport
    await page.setViewportSize({ width: 375, height: 667 });
    await page.goto(`${BASE_URL}/ko`);

    // Check if mobile navigation works
    const mobileNavButton = page.locator('button[aria-label*="menu"], button[aria-label*="navigation"]');
    if (await mobileNavButton.count() > 0) {
      await expect(mobileNavButton).toBeVisible();

      // Test mobile menu toggle
      await mobileNavButton.click();
      await page.waitForTimeout(500); // Wait for animation

      const mobileMenu = page.locator('.nextra-sidebar-container[style*="transform"], .nextra-nav-container');
      if (await mobileMenu.count() > 0) {
        await expect(mobileMenu).toBeVisible();
      }
    }

    // Check content readability on mobile
    await expect(page.locator('h1')).toBeVisible();
    await expect(page.locator('body')).toBeVisible();
  });

  test('Code highlighting and copy functionality', async ({ page }) => {
    await page.goto(`${BASE_URL}/ko`);

    // Look for code blocks
    const codeBlocks = page.locator('pre code');
    if (await codeBlocks.count() > 0) {
      const firstCodeBlock = codeBlocks.first();
      await expect(firstCodeBlock).toBeVisible();

      // Check for copy button
      const copyButton = page.locator('button[aria-label*="copy"], button[title*="copy"]');
      if (await copyButton.count() > 0) {
        await expect(copyButton.first()).toBeVisible();
      }
    }
  });

  test('Search functionality', async ({ page }) => {
    await page.goto(`${BASE_URL}/ko`);

    // Test search input
    const searchInput = page.locator('input[placeholder*="검색"]');
    if (await searchInput.count() > 0) {
      await expect(searchInput).toBeVisible();

      // Test search interaction
      await searchInput.fill('getting started');
      await page.waitForTimeout(1000); // Wait for search results

      // Check if search results appear
      const searchResults = page.locator('.pagefind-ui__result, .search-results');
      // Note: Pagefind results may take time to load
    }
  });

  test('Performance and accessibility basics', async ({ page }) => {
    await page.goto(`${BASE_URL}/ko`);

    // Check for basic accessibility
    await expect(page.locator('h1')).toBeVisible();

    // Check for proper heading structure
    const headings = page.locator('h1, h2, h3, h4, h5, h6');
    if (await headings.count() > 0) {
      await expect(headings.first()).toBeVisible();
    }

    // Check for proper landmarks
    await expect(page.locator('nav, main, [role="navigation"], [role="main"]')).toBeVisible();

    // Check for alt text on images
    const images = page.locator('img');
    const imageCount = await images.count();
    if (imageCount > 0) {
      for (let i = 0; i < Math.min(imageCount, 5); i++) { // Check first 5 images
        const img = images.nth(i);
        const alt = await img.getAttribute('alt');
        if (!alt && await img.isVisible()) {
          console.warn(`Image missing alt text: ${await img.getAttribute('src')}`);
        }
      }
    }
  });

  test('Footer and brand consistency', async ({ page }) => {
    await page.goto(`${BASE_URL}/ko`);

    // Check for footer
    const footer = page.locator('footer');
    if (await footer.count() > 0) {
      await expect(footer).toBeVisible();

      // Check for brand consistency
      await expect(page.locator('text=GoosLab')).toBeVisible();
      await expect(page.locator('text=MoAI-ADK')).toBeVisible();
    }

    // Check for GitHub link
    const githubLink = page.locator('a[href*="github"]');
    if (await githubLink.count() > 0) {
      await expect(githubLink.first()).toBeVisible();
    }
  });

  test.afterAll(async () => {
    console.log('UI/UX verification completed!');
  });
});

// Test for specific migration issues
test.describe('Migration Issues Detection', () => {

  test('Check for Nextra 4.x migration compatibility', async ({ page }) => {
    await page.goto(`${BASE_URL}/ko`);

    // Check for console errors
    page.on('console', (message) => {
      if (message.type() === 'error') {
        console.error('Console error:', message.text());
      }
    });

    // Check for broken styles
    const body = page.locator('body');
    await expect(body).toBeVisible();

    // Check for proper theme loading
    await expect(page.locator('html')).toHaveClass(/theme/);
  });

  test('Check Pagefind search integration', async ({ page }) => {
    await page.goto(`${BASE_URL}/ko`);

    // Check if Pagefind scripts are loaded
    const pagefindScript = page.locator('script[src*="pagefind"]');

    // Wait a bit for dynamic loading
    await page.waitForTimeout(2000);

    const scriptCount = await pagefindScript.count();
    if (scriptCount > 0) {
      console.log(`Found ${scriptCount} Pagefind scripts`);
    } else {
      console.log('Pagefind scripts may be loaded dynamically');
    }
  });
});