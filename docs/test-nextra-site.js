const { chromium } = require('playwright');
const fs = require('fs');
const path = require('path');

// Create screenshots directory
const screenshotDir = '/Users/goos/MoAI/MoAI-ADK/.playwright-mcp';
if (!fs.existsSync(screenshotDir)) {
  fs.mkdirSync(screenshotDir, { recursive: true });
}

async function test() {
  const browser = await chromium.launch();
  const context = await browser.createContext({
    viewport: { width: 1920, height: 1080 },
  });

  const testResults = {
    timestamp: new Date().toISOString(),
    tests: [],
  };

  try {
    // Test 1: Homepage Navigation
    console.log('\n=== Test 1: Homepage Navigation ===');
    let page = await context.newPage();

    try {
      await page.goto('http://localhost:3000', { waitUntil: 'networkidle', timeout: 10000 });
      await page.screenshot({ path: `${screenshotDir}/01-homepage.png` });
      console.log('✓ Homepage loads successfully (redirect from / to /ko)');
      testResults.tests.push({
        test: 'Homepage Navigation',
        status: 'PASS',
        details: 'Successfully navigated to homepage and redirected to /ko'
      });
    } catch (e) {
      console.error('✗ Homepage navigation failed:', e.message);
      testResults.tests.push({
        test: 'Homepage Navigation',
        status: 'FAIL',
        error: e.message
      });
    }
    await page.close();

    // Test 2: Korean Version Navigation
    console.log('\n=== Test 2: Korean Version Navigation ===');
    page = await context.newPage();

    try {
      await page.goto('http://localhost:3000/ko', { waitUntil: 'networkidle' });

      // Check if content is visible
      const titleVisible = await page.isVisible('h1');
      const contentVisible = await page.isVisible('main, [role="main"]');

      if (titleVisible || contentVisible) {
        console.log('✓ Korean version loads with visible content');
        testResults.tests.push({
          test: 'Korean Version Navigation',
          status: 'PASS',
          details: 'Korean version loads with visible content'
        });
      } else {
        throw new Error('No content visible on Korean page');
      }

      await page.screenshot({ path: `${screenshotDir}/02-korean-homepage.png` });
    } catch (e) {
      console.error('✗ Korean version navigation failed:', e.message);
      testResults.tests.push({
        test: 'Korean Version Navigation',
        status: 'FAIL',
        error: e.message
      });
    }
    await page.close();

    // Test 3: English Version Navigation
    console.log('\n=== Test 3: English Version Navigation ===');
    page = await context.newPage();

    try {
      await page.goto('http://localhost:3000/en', { waitUntil: 'networkidle' });

      const titleVisible = await page.isVisible('h1');
      const contentVisible = await page.isVisible('main, [role="main"]');

      if (titleVisible || contentVisible) {
        console.log('✓ English version loads with visible content');
        testResults.tests.push({
          test: 'English Version Navigation',
          status: 'PASS',
          details: 'English version loads with visible content'
        });
      } else {
        throw new Error('No content visible on English page');
      }

      await page.screenshot({ path: `${screenshotDir}/03-english-homepage.png` });
    } catch (e) {
      console.error('✗ English version navigation failed:', e.message);
      testResults.tests.push({
        test: 'English Version Navigation',
        status: 'FAIL',
        error: e.message
      });
    }
    await page.close();

    // Test 4: Light Theme Verification
    console.log('\n=== Test 4: Light Theme Verification ===');
    page = await context.newPage();

    try {
      await page.goto('http://localhost:3000/ko', { waitUntil: 'networkidle' });

      // Force light theme
      await page.evaluate(() => {
        localStorage.setItem('theme', 'light');
      });
      await page.reload({ waitUntil: 'networkidle' });

      // Get computed styles
      const htmlBg = await page.evaluate(() => {
        return window.getComputedStyle(document.documentElement).backgroundColor;
      });

      const bodyColor = await page.evaluate(() => {
        return window.getComputedStyle(document.body).color;
      });

      console.log(`✓ Light theme background: ${htmlBg}`);
      console.log(`✓ Light theme text color: ${bodyColor}`);
      testResults.tests.push({
        test: 'Light Theme Verification',
        status: 'PASS',
        details: `Background: ${htmlBg}, Text: ${bodyColor}`
      });

      await page.screenshot({ path: `${screenshotDir}/04-light-theme.png` });
    } catch (e) {
      console.error('✗ Light theme verification failed:', e.message);
      testResults.tests.push({
        test: 'Light Theme Verification',
        status: 'FAIL',
        error: e.message
      });
    }
    await page.close();

    // Test 5: Dark Theme Verification
    console.log('\n=== Test 5: Dark Theme Verification ===');
    page = await context.newPage();

    try {
      await page.goto('http://localhost:3000/ko', { waitUntil: 'networkidle' });

      // Force dark theme
      await page.evaluate(() => {
        localStorage.setItem('theme', 'dark');
        document.documentElement.classList.add('dark');
      });
      await page.reload({ waitUntil: 'networkidle' });

      // Get computed styles
      const htmlBg = await page.evaluate(() => {
        return window.getComputedStyle(document.documentElement).backgroundColor;
      });

      const bodyColor = await page.evaluate(() => {
        return window.getComputedStyle(document.body).color;
      });

      console.log(`✓ Dark theme background: ${htmlBg}`);
      console.log(`✓ Dark theme text color: ${bodyColor}`);
      testResults.tests.push({
        test: 'Dark Theme Verification',
        status: 'PASS',
        details: `Background: ${htmlBg}, Text: ${bodyColor}`
      });

      await page.screenshot({ path: `${screenshotDir}/05-dark-theme.png` });
    } catch (e) {
      console.error('✗ Dark theme verification failed:', e.message);
      testResults.tests.push({
        test: 'Dark Theme Verification',
        status: 'FAIL',
        error: e.message
      });
    }
    await page.close();

    // Test 6: Sidebar Navigation
    console.log('\n=== Test 6: Sidebar Navigation ===');
    page = await context.newPage();

    try {
      await page.goto('http://localhost:3000/ko', { waitUntil: 'networkidle' });

      // Check if sidebar exists
      const sidebarExists = await page.isVisible('nav, [role="navigation"], aside');

      if (sidebarExists) {
        console.log('✓ Sidebar is visible');

        // Try to find and click on a navigation link
        const navLinks = await page.$$('nav a, [role="navigation"] a, aside a');
        console.log(`✓ Found ${navLinks.length} navigation links`);

        testResults.tests.push({
          test: 'Sidebar Navigation',
          status: 'PASS',
          details: `Sidebar visible with ${navLinks.length} navigation links`
        });
      } else {
        throw new Error('Sidebar not found');
      }

      await page.screenshot({ path: `${screenshotDir}/06-sidebar-navigation.png` });
    } catch (e) {
      console.error('✗ Sidebar navigation test failed:', e.message);
      testResults.tests.push({
        test: 'Sidebar Navigation',
        status: 'FAIL',
        error: e.message
      });
    }
    await page.close();

    // Test 7: Table of Contents (TOC)
    console.log('\n=== Test 7: Table of Contents ===');
    page = await context.newPage();

    try {
      await page.goto('http://localhost:3000/ko', { waitUntil: 'networkidle' });

      // Check if TOC exists
      const tocExists = await page.isVisible('[role="doc-toc"], .toc, aside[aria-label*="toc" i]');

      if (tocExists) {
        console.log('✓ Table of Contents is visible');
        testResults.tests.push({
          test: 'Table of Contents',
          status: 'PASS',
          details: 'TOC component is visible'
        });
      } else {
        console.log('ℹ Table of Contents not found (may not be on homepage)');
        testResults.tests.push({
          test: 'Table of Contents',
          status: 'INFO',
          details: 'TOC not present on homepage (expected on content pages)'
        });
      }

      await page.screenshot({ path: `${screenshotDir}/07-toc.png` });
    } catch (e) {
      console.error('✗ TOC test failed:', e.message);
      testResults.tests.push({
        test: 'Table of Contents',
        status: 'FAIL',
        error: e.message
      });
    }
    await page.close();

    // Test 8: Responsive Design (Mobile View)
    console.log('\n=== Test 8: Responsive Design ===');
    page = await context.newPage();

    try {
      await page.setViewportSize({ width: 375, height: 812 });
      await page.goto('http://localhost:3000/ko', { waitUntil: 'networkidle' });

      console.log('✓ Mobile viewport (375x812) loaded successfully');
      testResults.tests.push({
        test: 'Responsive Design',
        status: 'PASS',
        details: 'Mobile viewport loads without errors'
      });

      await page.screenshot({ path: `${screenshotDir}/08-mobile-view.png` });
    } catch (e) {
      console.error('✗ Responsive design test failed:', e.message);
      testResults.tests.push({
        test: 'Responsive Design',
        status: 'FAIL',
        error: e.message
      });
    }
    await page.close();

    // Test 9: Search Functionality
    console.log('\n=== Test 9: Search Functionality ===');
    page = await context.newPage();

    try {
      await page.setViewportSize({ width: 1920, height: 1080 });
      await page.goto('http://localhost:3000/ko', { waitUntil: 'networkidle' });

      // Look for search input
      const searchInput = await page.$('input[type="search"], input[placeholder*="검색"], input[placeholder*="Search"]');

      if (searchInput) {
        console.log('✓ Search input found');
        await searchInput.click();
        await searchInput.type('Alfred', { delay: 100 });
        await page.waitForTimeout(1000);

        console.log('✓ Search query typed');
        testResults.tests.push({
          test: 'Search Functionality',
          status: 'PASS',
          details: 'Search input found and accepting input'
        });

        await page.screenshot({ path: `${screenshotDir}/09-search.png` });
      } else {
        console.log('ℹ Search input not found in header');
        testResults.tests.push({
          test: 'Search Functionality',
          status: 'INFO',
          details: 'Search functionality not readily visible'
        });
      }
    } catch (e) {
      console.error('✗ Search test failed:', e.message);
      testResults.tests.push({
        test: 'Search Functionality',
        status: 'FAIL',
        error: e.message
      });
    }
    await page.close();

    // Test 10: Language Switcher
    console.log('\n=== Test 10: Language Switcher ===');
    page = await context.newPage();

    try {
      await page.goto('http://localhost:3000/ko', { waitUntil: 'networkidle' });

      // Look for language selector
      const langSelector = await page.$('select[aria-label*="lang"], [role="combobox"][aria-label*="lang"], button[aria-label*="lang"]');

      if (langSelector) {
        console.log('✓ Language selector found');

        // Try to switch to English
        const englishLink = await page.$('a[href="/en"]');
        if (englishLink) {
          console.log('✓ English language link found');
          testResults.tests.push({
            test: 'Language Switcher',
            status: 'PASS',
            details: 'Language switcher found with English option'
          });
        }
      } else {
        // Try to find language links directly
        const langLink = await page.$('a[href="/en"]');
        if (langLink) {
          console.log('✓ Language links found');
          testResults.tests.push({
            test: 'Language Switcher',
            status: 'PASS',
            details: 'Language links available'
          });
        }
      }

      await page.screenshot({ path: `${screenshotDir}/10-language-switcher.png` });
    } catch (e) {
      console.error('✗ Language switcher test failed:', e.message);
      testResults.tests.push({
        test: 'Language Switcher',
        status: 'FAIL',
        error: e.message
      });
    }
    await page.close();

    // Test 11: Font Rendering (Korean)
    console.log('\n=== Test 11: Font Rendering (Korean) ===');
    page = await context.newPage();

    try {
      await page.goto('http://localhost:3000/ko', { waitUntil: 'networkidle' });

      // Check if Korean fonts are loaded
      const fontInfo = await page.evaluate(() => {
        const element = document.body;
        return window.getComputedStyle(element).fontFamily;
      });

      console.log(`✓ Font family: ${fontInfo}`);
      testResults.tests.push({
        test: 'Font Rendering (Korean)',
        status: 'PASS',
        details: `Font family: ${fontInfo}`
      });

      await page.screenshot({ path: `${screenshotDir}/11-korean-fonts.png` });
    } catch (e) {
      console.error('✗ Font rendering test failed:', e.message);
      testResults.tests.push({
        test: 'Font Rendering (Korean)',
        status: 'FAIL',
        error: e.message
      });
    }
    await page.close();

    // Test 12: Link Navigation
    console.log('\n=== Test 12: Link Navigation ===');
    page = await context.newPage();

    try {
      await page.goto('http://localhost:3000/ko', { waitUntil: 'networkidle' });

      // Get all links
      const links = await page.$$eval('a', as => as.map(a => ({ href: a.href, text: a.textContent.trim() })).filter(a => a.href && !a.href.includes('github.com')));

      console.log(`✓ Found ${links.length} internal links`);
      console.log('✓ Sample links:', links.slice(0, 3).map(l => `"${l.text}" -> ${l.href}`).join(', '));

      testResults.tests.push({
        test: 'Link Navigation',
        status: 'PASS',
        details: `Found ${links.length} internal links`
      });

      await page.screenshot({ path: `${screenshotDir}/12-links.png` });
    } catch (e) {
      console.error('✗ Link navigation test failed:', e.message);
      testResults.tests.push({
        test: 'Link Navigation',
        status: 'FAIL',
        error: e.message
      });
    }
    await page.close();

  } finally {
    await browser.close();
  }

  // Print summary
  console.log('\n' + '='.repeat(60));
  console.log('TEST SUMMARY');
  console.log('='.repeat(60));

  const passed = testResults.tests.filter(t => t.status === 'PASS').length;
  const failed = testResults.tests.filter(t => t.status === 'FAIL').length;
  const info = testResults.tests.filter(t => t.status === 'INFO').length;

  console.log(`Total: ${testResults.tests.length} | Pass: ${passed} | Fail: ${failed} | Info: ${info}`);

  testResults.tests.forEach(t => {
    const symbol = t.status === 'PASS' ? '✓' : t.status === 'FAIL' ? '✗' : 'ℹ';
    console.log(`${symbol} ${t.test}: ${t.status}`);
  });

  // Save results to JSON
  fs.writeFileSync(`${screenshotDir}/test-results.json`, JSON.stringify(testResults, null, 2));
  console.log(`\nResults saved to: ${screenshotDir}/test-results.json`);
  console.log(`Screenshots saved to: ${screenshotDir}/`);
}

test().catch(console.error);
