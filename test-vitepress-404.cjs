// E2E Test to diagnose VitePress 404 error
const { chromium } = require('playwright');

(async () => {
  console.log('🚀 Starting VitePress 404 Diagnosis...\n');

  const browser = await chromium.launch({
    headless: false, // Show browser for debugging
    slowMo: 1000 // Slow down actions
  });

  const context = await browser.newContext();
  const page = await context.newPage();

  // Listen to console messages
  page.on('console', msg => {
    const type = msg.type();
    const text = msg.text();
    if (type === 'error' || text.includes('404')) {
      console.log(`❌ Browser Console [${type}]:`, text);
    }
  });

  // Listen to page errors
  page.on('pageerror', error => {
    console.log('❌ Page Error:', error.message);
  });

  // Listen to failed requests
  page.on('requestfailed', request => {
    console.log('❌ Request Failed:', request.url(), request.failure().errorText);
  });

  try {
    console.log('📍 Navigating to http://localhost:5173/...');
    const response = await page.goto('http://localhost:5173/', {
      waitUntil: 'networkidle',
      timeout: 30000
    });

    console.log(`✅ Response Status: ${response.status()}`);
    console.log(`✅ Response URL: ${response.url()}`);

    // Wait for page to load
    await page.waitForTimeout(3000);

    // Check page title
    const title = await page.title();
    console.log(`📄 Page Title: "${title}"`);

    // Check if 404 page is displayed
    const bodyText = await page.textContent('body');
    if (bodyText.includes('404') || bodyText.includes('PAGE NOT FOUND')) {
      console.log('❌ 404 ERROR DETECTED IN PAGE CONTENT');
      console.log('Body preview:', bodyText.substring(0, 500));
    } else {
      console.log('✅ No 404 error in page content');
    }

    // Check for VitePress app
    const appDiv = await page.$('#app');
    if (appDiv) {
      const appContent = await appDiv.textContent();
      console.log('✅ VitePress app div found');
      if (appContent.includes('404')) {
        console.log('❌ 404 detected inside #app div');
        console.log('App content preview:', appContent.substring(0, 300));
      }
    } else {
      console.log('❌ VitePress app div NOT found');
    }

    // Check for hero section (should be on homepage)
    const heroSection = await page.$('.VPHero');
    if (heroSection) {
      console.log('✅ Hero section found - homepage is loading correctly');
      const heroText = await heroSection.textContent();
      console.log('Hero text:', heroText.substring(0, 200));
    } else {
      console.log('❌ Hero section NOT found - homepage may not be loading');
    }

    // Check current URL (might be redirected)
    const currentUrl = page.url();
    console.log(`📍 Current URL: ${currentUrl}`);

    // Take screenshot
    await page.screenshot({ path: '/Users/goos/MoAI/MoAI-ADK/screenshot-404-test.png', fullPage: true });
    console.log('📸 Screenshot saved: screenshot-404-test.png');

    // Check network requests
    const requests = [];
    page.on('request', request => requests.push(request.url()));
    await page.reload({ waitUntil: 'networkidle' });

    console.log('\n📊 Network Requests Summary:');
    const failed = requests.filter(url => url.includes('404'));
    if (failed.length > 0) {
      console.log('❌ Failed requests:', failed);
    } else {
      console.log('✅ All requests successful');
    }

  } catch (error) {
    console.error('❌ Test Error:', error.message);
  } finally {
    console.log('\n🔚 Test completed. Browser will close in 5 seconds...');
    await page.waitForTimeout(5000);
    await browser.close();
  }
})();