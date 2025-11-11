const { chromium } = require('playwright');

async function takeScreenshots() {
  const browser = await chromium.launch();
  const context = await browser.newContext({
    viewport: { width: 1920, height: 1080 }
  });
  const page = await context.newPage();

  console.log('Navigating to homepage...');
  await page.goto('http://localhost:3000');
  await page.waitForLoadState('networkidle');
  await page.screenshot({ path: 'homepage.png', fullPage: true });
  console.log('Homepage screenshot saved');

  console.log('Taking Korean language page screenshot...');
  await page.goto('http://localhost:3000/ko');
  await page.waitForLoadState('networkidle');
  await page.screenshot({ path: 'korean-page.png', fullPage: true });
  console.log('Korean page screenshot saved');

  console.log('Taking English language page screenshot...');
  await page.goto('http://localhost:3000/en');
  await page.waitForLoadState('networkidle');
  await page.screenshot({ path: 'english-page.png', fullPage: true });
  console.log('English page screenshot saved');

  console.log('Taking Japanese language page screenshot...');
  await page.goto('http://localhost:3000/ja');
  await page.waitForLoadState('networkidle');
  await page.screenshot({ path: 'japanese-page.png', fullPage: true });
  console.log('Japanese page screenshot saved');

  console.log('Taking Chinese language page screenshot...');
  await page.goto('http://localhost:3000/zh');
  await page.waitForLoadState('networkidle');
  await page.screenshot({ path: 'chinese-page.png', fullPage: true });
  console.log('Chinese page screenshot saved');

  // Mobile viewport test
  await context.setDefaultViewport({ width: 375, height: 667 });
  console.log('Taking mobile homepage screenshot...');
  await page.goto('http://localhost:3000');
  await page.waitForLoadState('networkidle');
  await page.screenshot({ path: 'mobile-homepage.png', fullPage: true });
  console.log('Mobile homepage screenshot saved');

  await browser.close();
  console.log('All screenshots completed!');
}

takeScreenshots().catch(console.error);