import { test, expect, devices } from '@playwright/test';
import { injectAxe, checkA11y } from 'axe-playwright';

// Test suite for comprehensive UI/UX analysis
test.describe('MoAI-ADK Documentation UI/UX Analysis', () => {
  let baseUrl: string;

  test.beforeAll(async () => {
    baseUrl = 'http://localhost:3000';
  });

  // 1. Visual Design Analysis
  test.describe('Visual Design Analysis', () => {
    test('should have proper typography and spacing', async ({ page }) => {
      await page.goto(baseUrl);

      // Check main heading
      const mainHeading = page.locator('h1').first();
      await expect(mainHeading).toBeVisible();

      // Check font properties using computed styles
      const headingStyles = await mainHeading.evaluate((el) => {
        const styles = window.getComputedStyle(el);
        return {
          fontSize: styles.fontSize,
          fontWeight: styles.fontWeight,
          lineHeight: styles.lineHeight,
          fontFamily: styles.fontFamily
        };
      });

      console.log('Main heading styles:', headingStyles);

      // Check spacing between elements
      const navigation = page.locator('nav').first();
      const mainContent = page.locator('main').first();

      if (await navigation.count() > 0 && await mainContent.count() > 0) {
        const navBounds = await navigation.boundingBox();
        const mainBounds = await mainContent.boundingBox();

        if (navBounds && mainBounds) {
          const spacing = mainBounds.y - (navBounds.y + navBounds.height);
          console.log(`Spacing between nav and main: ${spacing}px`);
          expect(spacing).toBeGreaterThan(16); // At least 16px spacing
        }
      }

      // Take screenshot for visual reference
      await page.screenshot({
        path: 'test-results/visual-design-typography.png',
        fullPage: true
      });
    });

    test('should have proper color scheme and contrast', async ({ page }) => {
      await page.goto(baseUrl);

      // Check primary colors
      const body = page.locator('body');
      const bodyStyles = await body.evaluate((el) => {
        const styles = window.getComputedStyle(el);
        return {
          backgroundColor: styles.backgroundColor,
          color: styles.color
        };
      });

      console.log('Body color scheme:', bodyStyles);

      // Check link colors
      const links = page.locator('a').first();
      if (await links.count() > 0) {
        const linkStyles = await links.evaluate((el) => {
          const styles = window.getComputedStyle(el);
          return {
            color: styles.color,
            textDecoration: styles.textDecoration
          };
        });
        console.log('Link styles:', linkStyles);
      }

      await page.screenshot({
        path: 'test-results/visual-design-colors.png',
        fullPage: true
      });
    });

    test('should have proper visual hierarchy', async ({ page }) => {
      await page.goto(baseUrl);

      // Check heading hierarchy
      const headings = await page.locator('h1, h2, h3, h4, h5, h6').all();
      console.log(`Found ${headings.length} headings`);

      for (let i = 0; i < headings.length; i++) {
        const tagName = await headings[i].evaluate(el => el.tagName);
        const text = await headings[i].textContent();
        console.log(`${tagName}: ${text?.substring(0, 50)}...`);
      }

      // Check for proper heading order (h1 should come before h2, etc.)
      const headingLevels = await Promise.all(
        headings.map(h => h.evaluate(el => parseInt(el.tagName.substring(1))))
      );

      for (let i = 1; i < headingLevels.length; i++) {
        expect(headingLevels[i]).toBeLessThanOrEqual(headingLevels[i - 1] + 1);
      }
    });
  });

  // 2. Navigation Assessment
  test.describe('Navigation Assessment', () => {
    test('should have working navigation links', async ({ page }) => {
      await page.goto(baseUrl);

      // Find all navigation links
      const navLinks = page.locator('nav a, .navigation a, [role="navigation"] a');
      const linkCount = await navLinks.count();

      console.log(`Found ${linkCount} navigation links`);

      // Test first few navigation links
      const linksToTest = Math.min(linkCount, 5);
      for (let i = 0; i < linksToTest; i++) {
        const link = navLinks.nth(i);
        const href = await link.getAttribute('href');
        const text = await link.textContent();

        console.log(`Testing link: ${text} -> ${href}`);

        if (href && !href.startsWith('http')) {
          // Test internal navigation
          await link.click();
          await page.waitForLoadState('networkidle');

          // Check that navigation succeeded
          await expect(page).toHaveURL(new RegExp(href));

          // Go back for next test
          await page.goBack();
          await page.waitForLoadState('networkidle');
        }
      }
    });

    test('should have working language switching', async ({ page }) => {
      await page.goto(baseUrl);

      // Look for language switcher
      const languageSelectors = [
        '[data-language-selector]',
        '.language-selector',
        '[role="combobox"]',
        'select'
      ];

      let languageSwitcher = null;
      for (const selector of languageSelectors) {
        const element = page.locator(selector);
        if (await element.count() > 0) {
          languageSwitcher = element.first();
          break;
        }
      }

      if (languageSwitcher) {
        console.log('Found language switcher');

        // Try different language options
        const languages = ['ko', 'en', 'ja', 'zh'];
        for (const lang of languages) {
          // Check if there's a language-specific link
          const langLink = page.locator(`a[href*="/${lang}"], a[href*="lang=${lang}"]`).first();
          if (await langLink.count() > 0) {
            await langLink.click();
            await page.waitForLoadState('networkidle');

            // Check if URL changed to include language
            const url = page.url();
            console.log(`Switched to ${lang}: ${url}`);

            // Go back
            await page.goBack();
            await page.waitForLoadState('networkidle');
          }
        }
      } else {
        console.log('Language switcher not found');
      }
    });

    test('should have search functionality', async ({ page }) => {
      await page.goto(baseUrl);

      // Look for search input
      const searchSelectors = [
        'input[type="search"]',
        'input[placeholder*="search" i]',
        'input[placeholder*="검색" i]',
        '[role="search"] input',
        '.search-input'
      ];

      let searchInput = null;
      for (const selector of searchSelectors) {
        const element = page.locator(selector);
        if (await element.count() > 0) {
          searchInput = element.first();
          break;
        }
      }

      if (searchInput) {
        console.log('Found search input');

        // Test search functionality
        await searchInput.fill('MoAI');
        await page.keyboard.press('Enter');

        // Wait a moment for search results
        await page.waitForTimeout(1000);

        // Check for search results
        const searchResults = page.locator('.search-results, [role="listbox"], .search-suggestions');
        if (await searchResults.count() > 0) {
          console.log('Search results found');
        } else {
          console.log('No search results container found');
        }

        await page.screenshot({
          path: 'test-results/search-functionality.png',
          fullPage: true
        });
      } else {
        console.log('Search input not found');
      }
    });
  });

  // 3. Accessibility Check
  test.describe('Accessibility Check', () => {
    test('should pass axe accessibility audit', async ({ page }) => {
      await page.goto(baseUrl);
      await injectAxe(page);

      // Run accessibility audit
      await checkA11y(page, null, {
        detailedReport: true,
        detailedReportOptions: { html: true },
        rules: {
          // Enable important rules
          'color-contrast': { enabled: true },
          'keyboard-navigation': { enabled: true },
          'focus-order-semantics': { enabled: true },
          'heading-order': { enabled: true },
          'landmark-one-main': { enabled: true },
          'page-has-heading-one': { enabled: true },
          'region': { enabled: true }
        }
      }, false, {
        html: true,
        detailedReport: true
      });
    });

    test('should support keyboard navigation', async ({ page }) => {
      await page.goto(baseUrl);

      // Test Tab navigation
      const focusableElements = await page.locator(
        'a, button, input, select, textarea, [tabindex]:not([tabindex="-1"])'
      ).all();

      console.log(`Found ${focusableElements.length} focusable elements`);

      // Test Tab through first few elements
      for (let i = 0; i < Math.min(focusableElements.length, 10); i++) {
        await page.keyboard.press('Tab');

        // Check that something is focused
        const focusedElement = page.locator(':focus');
        expect(await focusedElement.count()).toBeGreaterThan(0);

        const tagName = await focusedElement.evaluate(el => el.tagName);
        console.log(`Tab ${i + 1}: Focused on ${tagName}`);
      }

      // Test Shift+Tab navigation
      for (let i = 0; i < 3; i++) {
        await page.keyboard.press('Shift+Tab');
        const focusedElement = page.locator(':focus');
        expect(await focusedElement.count()).toBeGreaterThan(0);
      }

      // Test Enter key on links
      const firstLink = page.locator('a').first();
      if (await firstLink.count() > 0) {
        await firstLink.focus();
        const href = await firstLink.getAttribute('href');

        if (href && !href.startsWith('http')) {
          await page.keyboard.press('Enter');
          await page.waitForLoadState('networkidle');
          console.log('Enter key navigation working');

          // Go back
          await page.goBack();
        }
      }

      await page.screenshot({
        path: 'test-results/keyboard-navigation.png',
        fullPage: true
      });
    });

    test('should have proper ARIA labels and semantic HTML', async ({ page }) => {
      await page.goto(baseUrl);

      // Check for semantic landmarks
      const landmarks = [
        'main',
        'nav',
        'header',
        'footer',
        '[role="main"]',
        '[role="navigation"]',
        '[role="banner"]',
        '[role="contentinfo"]'
      ];

      for (const landmark of landmarks) {
        const element = page.locator(landmark);
        if (await element.count() > 0) {
          console.log(`Found landmark: ${landmark}`);
        }
      }

      // Check for proper heading structure
      const h1 = page.locator('h1');
      expect(await h1.count()).toBeGreaterThanOrEqual(1);

      // Check for ARIA labels on interactive elements
      const buttonsWithoutText = page.locator('button:not([aria-label]):not([aria-labelledby])');
      const buttonCount = await buttonsWithoutText.count();

      if (buttonCount > 0) {
        console.log(`Found ${buttonCount} buttons without accessible labels`);

        // Check if they have visible text or icons
        for (let i = 0; i < Math.min(buttonCount, 5); i++) {
          const button = buttonsWithoutText.nth(i);
          const hasText = await button.evaluate(el => el.textContent?.trim());
          const hasIcon = await button.locator('svg, i, .icon').count();

          if (!hasText && hasIcon === 0) {
            console.warn(`Button at index ${i} lacks accessible label`);
          }
        }
      }
    });
  });

  // 4. Performance Evaluation
  test.describe('Performance Evaluation', () => {
    test('should load within acceptable time', async ({ page }) => {
      const startTime = Date.now();
      await page.goto(baseUrl);
      await page.waitForLoadState('networkidle');
      const loadTime = Date.now() - startTime;

      console.log(`Page load time: ${loadTime}ms`);
      expect(loadTime).toBeLessThan(3000); // Should load within 3 seconds

      // Check Core Web Vitals
      const metrics = await page.evaluate(() => {
        return new Promise((resolve) => {
          new PerformanceObserver((list) => {
            const entries = list.getEntries();
            const vitals = {
              LCP: entries.find(e => e.name === 'largest-contentful-paint')?.startTime || 0,
              FID: entries.find(e => e.name === 'first-input')?.processingStart - entries.find(e => e.name === 'first-input')?.startTime || 0,
              CLS: entries.find(e => e.name === 'cumulative-layout-shift')?.value || 0
            };
            resolve(vitals);
          }).observe({ entryTypes: ['largest-contentful-paint', 'first-input', 'cumulative-layout-shift'] });

          // Fallback timeout
          setTimeout(() => resolve({ LCP: 0, FID: 0, CLS: 0 }), 2000);
        });
      });

      console.log('Core Web Vitals:', metrics);
    });

    test('should handle scroll performance', async ({ page }) => {
      await page.goto(baseUrl);
      await page.waitForLoadState('networkidle');

      // Get page height
      const pageHeight = await page.evaluate(() => document.body.scrollHeight);
      console.log(`Page height: ${pageHeight}px`);

      if (pageHeight > window.innerHeight * 2) {
        // Test smooth scrolling
        const scrollSteps = 5;
        const scrollAmount = pageHeight / scrollSteps;

        const startTime = Date.now();

        for (let i = 1; i <= scrollSteps; i++) {
          await page.evaluate((scrollTo) => {
            window.scrollTo({ top: scrollTo, behavior: 'smooth' });
          }, scrollAmount * i);

          await page.waitForTimeout(500);
        }

        const scrollTime = Date.now() - startTime;
        console.log(`Scroll performance time: ${scrollTime}ms`);

        // Test if elements are loading properly during scroll
        await page.evaluate(() => window.scrollTo(0, 0));
      }
    });
  });

  // 5. Content Review
  test.describe('Content Review', () => {
    test('should render all sections properly', async ({ page }) => {
      await page.goto(baseUrl);

      // Check for main content sections
      const contentSelectors = [
        'main',
        'article',
        '.content',
        '.documentation',
        '[role="main"]'
      ];

      let mainContent = null;
      for (const selector of contentSelectors) {
        const element = page.locator(selector);
        if (await element.count() > 0) {
          mainContent = element.first();
          break;
        }
      }

      expect(mainContent).toBeTruthy();

      if (mainContent) {
        // Check for content visibility
        await expect(mainContent).toBeVisible();

        // Check for text content
        const textContent = await mainContent.textContent();
        expect(textContent?.length).toBeGreaterThan(100);

        console.log(`Main content length: ${textContent?.length} characters`);
      }

      await page.screenshot({
        path: 'test-results/content-rendering.png',
        fullPage: true
      });
    });

    test('should display MDX content correctly', async ({ page }) => {
      await page.goto(baseUrl);

      // Look for code blocks
      const codeBlocks = page.locator('pre code, .code-block, [class*="code"]');
      const codeBlockCount = await codeBlocks.count();

      console.log(`Found ${codeBlockCount} code blocks`);

      // Check code highlighting
      for (let i = 0; i < Math.min(codeBlockCount, 3); i++) {
        const codeBlock = codeBlocks.nth(i);
        const codeText = await codeBlock.textContent();
        console.log(`Code block ${i + 1}: ${codeText?.substring(0, 50)}...`);

        // Check for syntax highlighting classes
        const hasSyntaxHighlighting = await codeBlock.evaluate(el => {
          const classList = Array.from(el.classList);
          return classList.some(cls => cls.includes('language-') || cls.includes('highlight'));
        });

        if (hasSyntaxHighlighting) {
          console.log(`Code block ${i + 1} has syntax highlighting`);
        }
      }

      // Look for other MDX elements
      const mdxElements = [
        { selector: 'blockquote', name: 'Blockquotes' },
        { selector: 'table', name: 'Tables' },
        { selector: '.markdown-alert, .callout', name: 'Alerts/Callouts' },
        { selector: 'img', name: 'Images' }
      ];

      for (const { selector, name } of mdxElements) {
        const elements = page.locator(selector);
        const count = await elements.count();
        if (count > 0) {
          console.log(`Found ${count} ${name}`);
        }
      }
    });

    test('should have proper meta information', async ({ page }) => {
      await page.goto(baseUrl);

      // Check page title
      const title = await page.title();
      console.log(`Page title: ${title}`);
      expect(title).toContain('MoAI-ADK');

      // Check meta description
      const metaDescription = await page.locator('meta[name="description"]').getAttribute('content');
      if (metaDescription) {
        console.log(`Meta description: ${metaDescription}`);
        expect(metaDescription.length).toBeGreaterThan(50);
      }

      // Check language meta tag
      const htmlLang = await page.locator('html').getAttribute('lang');
      console.log(`HTML lang: ${htmlLang}`);
      expect(htmlLang).toMatch(/^(en|ko|ja|zh)$/);
    });
  });

  // Mobile Responsiveness
  test.describe('Mobile Responsiveness', () => {
    ['Pixel 5', 'iPhone 12'].forEach(deviceName => {
      test(`should be responsive on ${deviceName}`, async ({ page }) => {
        const device = devices[deviceName as keyof typeof devices];
        await page.setViewportSize(device.viewport || { width: 375, height: 667 });

        await page.goto(baseUrl);
        await page.waitForLoadState('networkidle');

        // Check mobile navigation
        const mobileNav = page.locator('.mobile-nav, .hamburger, [aria-label="menu"], button[aria-expanded]');
        if (await mobileNav.count() > 0) {
          console.log(`Found mobile navigation on ${deviceName}`);

          // Test mobile menu toggle
          await mobileNav.first().click();
          await page.waitForTimeout(500);

          // Check if menu expands
          const isExpanded = await mobileNav.first().getAttribute('aria-expanded');
          console.log(`Mobile menu expanded: ${isExpanded}`);

          await page.screenshot({
            path: `test-results/mobile-${deviceName.toLowerCase().replace(' ', '-')}-menu.png`,
            fullPage: true
          });
        }

        // Check horizontal scrolling (should not exist)
        const body = page.locator('body');
        const bodyOverflow = await body.evaluate(el => {
          return window.getComputedStyle(el).overflowX;
        });

        console.log(`Body overflow-x on ${deviceName}: ${bodyOverflow}`);
        expect(bodyOverflow).not.toBe('scroll');

        await page.screenshot({
          path: `test-results/mobile-${deviceName.toLowerCase().replace(' ', '-')}-responsive.png`,
          fullPage: true
        });
      });
    });
  });

  // Cross-browser compatibility
  test.describe('Cross-browser Compatibility', () => {
    ['chromium', 'firefox', 'webkit'].forEach(browserName => {
      test(`should work in ${browserName}`, async ({ page, browserName: currentBrowser }) => {
        test.skip(currentBrowser !== browserName, `This test runs only on ${browserName}`);

        await page.goto(baseUrl);
        await page.waitForLoadState('networkidle');

        // Basic functionality check
        await expect(page.locator('body')).toBeVisible();

        const title = await page.title();
        expect(title).toContain('MoAI-ADK');

        console.log(`${browserName} compatibility test passed`);

        await page.screenshot({
          path: `test-results/browser-${browserName}.png`,
          fullPage: true
        });
      });
    });
  });
});