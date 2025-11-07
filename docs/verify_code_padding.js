const { chromium } = require('playwright');

(async () => {
  const browser = await chromium.launch();
  const context = await browser.createBrowserContext();
  const page = await context.newPage();

  try {
    console.log('🔍 Loading Korean documentation homepage...');
    await page.goto('http://localhost:8000/ko/', { waitUntil: 'networkidle' });

    // Find a page with code blocks - let's go to a guide page
    console.log('🔍 Navigating to Alfred guide page...');
    await page.goto('http://localhost:8000/ko/guides/alfred/', { waitUntil: 'networkidle' });

    // Wait for content to load
    await page.waitForTimeout(1000);

    // Check for code blocks (div.highlight)
    const codeBlocks = await page.locator('div.highlight').all();
    console.log(`\n📊 Found ${codeBlocks.length} code blocks on the page\n`);

    if (codeBlocks.length === 0) {
      console.log('❌ No code blocks found. Let\'s check for pre/code elements...');
      const preElements = await page.locator('pre').all();
      console.log(`Found ${preElements.length} <pre> elements\n`);
    }

    // Check the first code block in detail
    if (codeBlocks.length > 0) {
      const firstCodeBlock = codeBlocks[0];

      console.log('✅ Checking first code block styles...\n');

      const styles = await firstCodeBlock.evaluate(el => {
        const computed = window.getComputedStyle(el);
        const rect = el.getBoundingClientRect();
        return {
          padding: computed.padding,
          paddingTop: computed.paddingTop,
          paddingRight: computed.paddingRight,
          paddingBottom: computed.paddingBottom,
          paddingLeft: computed.paddingLeft,
          margin: computed.margin,
          marginTop: computed.marginTop,
          marginRight: computed.marginRight,
          marginBottom: computed.marginBottom,
          marginLeft: computed.marginLeft,
          width: rect.width,
          height: rect.height,
          backgroundColor: computed.backgroundColor,
          fontFamily: computed.fontFamily,
          fontSize: computed.fontSize,
        };
      });

      console.log('📐 Code Block Dimensions:');
      console.log(`   Width: ${styles.width}px, Height: ${styles.height}px`);

      console.log('\n📏 Padding Applied:');
      console.log(`   Top: ${styles.paddingTop}`);
      console.log(`   Right: ${styles.paddingRight}`);
      console.log(`   Bottom: ${styles.paddingBottom}`);
      console.log(`   Left: ${styles.paddingLeft}`);

      console.log('\n📏 Margin Applied:');
      console.log(`   Top: ${styles.marginTop}`);
      console.log(`   Right: ${styles.marginRight}`);
      console.log(`   Bottom: ${styles.marginBottom}`);
      console.log(`   Left: ${styles.marginLeft}`);

      console.log('\n🎨 Style Properties:');
      console.log(`   Background Color: ${styles.backgroundColor}`);
      console.log(`   Font Family: ${styles.fontFamily}`);
      console.log(`   Font Size: ${styles.fontSize}`);

      // Check if Hack font is present
      if (styles.fontFamily.toLowerCase().includes('hack')) {
        console.log('\n✅ Hack font IS being applied!');
      } else {
        console.log('\n❌ Hack font NOT found. Font family: ' + styles.fontFamily);
      }

      // Check padding
      const paddingValues = [
        parseInt(styles.paddingTop),
        parseInt(styles.paddingRight),
        parseInt(styles.paddingBottom),
        parseInt(styles.paddingLeft)
      ];

      if (paddingValues.every(v => v > 15)) { // 1.2em ≈ 19px for 16px base
        console.log('✅ Adequate padding IS being applied!');
      } else {
        console.log('❌ Padding is NOT sufficient. Expected ~19px (1.2em), got:');
        console.log(`   Min padding: ${Math.min(...paddingValues)}px`);
      }

      // Show the raw HTML of the code block
      console.log('\n📝 Code Block HTML:');
      const html = await firstCodeBlock.innerHTML();
      console.log(html.substring(0, 200) + '...');
    }

    // Take a screenshot to verify visually
    console.log('\n📸 Taking screenshot of page...');
    await page.screenshot({ path: '/tmp/code-block-verification.png', fullPage: true });
    console.log('✅ Screenshot saved to /tmp/code-block-verification.png');

  } catch (error) {
    console.error('❌ Error during verification:', error);
  } finally {
    await browser.close();
  }
})();
