/**
 * Simple Playwright screenshot test for MoAI-ADK Documentation
 */

const { chromium } = require('playwright');
const fs = require('fs');

async function captureScreenshots() {
  console.log('🎨 Starting MoAI-ADK Documentation Screenshot Analysis...');

  const browser = await chromium.launch();
  const context = await browser.newContext();
  const page = await context.newPage();

  try {
    // Set up monitoring for console errors
    const consoleErrors = [];
    page.on('console', msg => {
      if (msg.type() === 'error') {
        consoleErrors.push(msg.text());
      }
    });

    // Network monitoring
    const networkErrors = [];
    page.on('response', response => {
      if (response.status() >= 400) {
        networkErrors.push(`${response.url()} - ${response.status()}`);
      }
    });

    console.log('📄 Loading documentation page...');

    // Navigate to the page with extended timeout
    const response = await page.goto('http://localhost:3000', {
      waitUntil: 'networkidle',
      timeout: 30000
    });

    console.log(`Page loaded with status: ${response.status()}`);

    // Wait for content to potentially load
    await page.waitForTimeout(5000);

    // Get page information
    const title = await page.title();
    const url = page.url();

    console.log(`Page title: "${title}"`);
    console.log(`Final URL: ${url}`);

    // Check for any content on the page
    const contentAnalysis = await page.evaluate(() => {
      const bodyText = document.body.innerText || document.body.textContent || '';
      const htmlContent = document.documentElement.outerHTML;

      return {
        bodyTextLength: bodyText.length,
        hasVisibleContent: bodyText.trim().length > 0,
        htmlLength: htmlContent.length,
        elementCount: document.querySelectorAll('*').length,
        bodyClasses: document.body.className,
        bodyStyle: document.body.getAttribute('style') || '',
        computedStyles: getComputedStyle(document.body),
        hasNextraElements: {
          container: !!document.querySelector('.nextra-container'),
          sidebar: !!document.querySelector('.nextra-sidebar-container'),
          nav: !!document.querySelector('.nextra-nav-container'),
          content: !!document.querySelector('.nextra-content'),
          search: !!document.querySelector('.nextra-search')
        }
      };
    });

    console.log('📊 Content Analysis:');
    console.log(`  Body text length: ${contentAnalysis.bodyTextLength}`);
    console.log(`  Has visible content: ${contentAnalysis.hasVisibleContent}`);
    console.log(`  Total elements: ${contentAnalysis.elementCount}`);
    console.log(`  Body classes: "${contentAnalysis.bodyClasses}"`);

    // Check for loaded stylesheets
    const stylesheets = await page.evaluate(() => {
      const links = Array.from(document.querySelectorAll('link[rel="stylesheet"]'));
      return links.map(link => ({
        href: link.href,
        loaded: link.sheet !== null
      }));
    });

    console.log('🎨 Stylesheets:');
    stylesheets.forEach((sheet, index) => {
      console.log(`  ${index + 1}. ${sheet.href} - Loaded: ${sheet.loaded}`);
    });

    // Check for script errors
    if (consoleErrors.length > 0) {
      console.log('\n❌ Console Errors:');
      consoleErrors.forEach(error => console.log(`  • ${error}`));
    }

    if (networkErrors.length > 0) {
      console.log('\n🌐 Network Errors:');
      networkErrors.forEach(error => console.log(`  • ${error}`));
    }

    // Capture screenshots at different viewport sizes
    const viewports = [
      { name: 'mobile', width: 375, height: 667 },
      { name: 'tablet', width: 768, height: 1024 },
      { name: 'desktop', width: 1920, height: 1080 }
    ];

    for (const viewport of viewports) {
      await page.setViewportSize({ width: viewport.width, height: viewport.height });
      await page.waitForTimeout(1000); // Wait for any responsive changes

      const screenshotPath = `docs-screenshot-${viewport.name}-${viewport.width}x${viewport.height}.png`;
      await page.screenshot({
        path: screenshotPath,
        fullPage: true
      });

      console.log(`📸 Screenshot saved: ${screenshotPath}`);
    }

    // Generate a simple HTML report
    const report = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>MoAI-ADK Documentation Analysis Report</title>
    <style>
        body { font-family: system-ui, sans-serif; margin: 2rem; line-height: 1.6; }
        .header { background: #f0f0f0; padding: 1rem; border-radius: 8px; margin-bottom: 2rem; }
        .section { margin: 2rem 0; padding: 1rem; border-left: 4px solid #007acc; background: #f9f9f9; }
        .error { border-left-color: #e74c3c; background: #fdf2f2; }
        .success { border-left-color: #27ae60; background: #f2fdf2; }
        .warning { border-left-color: #f39c12; background: #fef9e7; }
        .screenshot { margin: 1rem 0; max-width: 100%; }
        .screenshot img { max-width: 100%; border: 1px solid #ddd; border-radius: 4px; }
        pre { background: #f4f4f4; padding: 1rem; overflow-x: auto; border-radius: 4px; }
    </style>
</head>
<body>
    <div class="header">
        <h1>🎨 MoAI-ADK Documentation Analysis Report</h1>
        <p>Generated on: ${new Date().toLocaleString()}</p>
        <p>Analysis URL: ${url}</p>
        <p>Page Title: "${title}"</p>
    </div>

    <div class="section ${contentAnalysis.hasVisibleContent ? 'success' : 'error'}">
        <h2>📄 Content Analysis</h2>
        <ul>
            <li>Body text length: ${contentAnalysis.bodyTextLength} characters</li>
            <li>Has visible content: ${contentAnalysis.hasVisibleContent ? '✅ Yes' : '❌ No'}</li>
            <li>Total elements: ${contentAnalysis.elementCount}</li>
            <li>Body classes: "${contentAnalysis.bodyClasses}"</li>
        </ul>
    </div>

    <div class="section">
        <h2>🎨 Stylesheets (${stylesheets.length})</h2>
        <ul>
            ${stylesheets.map(sheet =>
              `<li>${sheet.href} - ${sheet.loaded ? '✅ Loaded' : '❌ Not loaded'}</li>`
            ).join('')}
        </ul>
    </div>

    ${consoleErrors.length > 0 ? `
    <div class="section error">
        <h2>❌ Console Errors (${consoleErrors.length})</h2>
        <ul>
            ${consoleErrors.map(error => `<li>${error}</li>`).join('')}
        </ul>
    </div>
    ` : ''}

    ${networkErrors.length > 0 ? `
    <div class="section error">
        <h2>🌐 Network Errors (${networkErrors.length})</h2>
        <ul>
            ${networkErrors.map(error => `<li>${error}</li>`).join('')}
        </ul>
    </div>
    ` : ''}

    <div class="section">
        <h2>📸 Screenshots</h2>
        ${viewports.map(vp => `
            <div class="screenshot">
                <h3>${vp.name} (${vp.width}x${vp.height})</h3>
                <img src="docs-screenshot-${vp.name}-${vp.width}x${vp.height}.png" alt="${vp.name} screenshot">
            </div>
        `).join('')}
    </div>

    <div class="section">
        <h2>🔍 Nextra Element Detection</h2>
        <ul>
            <li>nextra-container: ${contentAnalysis.hasNextraElements.container ? '✅ Found' : '❌ Missing'}</li>
            <li>nextra-sidebar-container: ${contentAnalysis.hasNextraElements.sidebar ? '✅ Found' : '❌ Missing'}</li>
            <li>nextra-nav-container: ${contentAnalysis.hasNextraElements.nav ? '✅ Found' : '❌ Missing'}</li>
            <li>nextra-content: ${contentAnalysis.hasNextraElements.content ? '✅ Found' : '❌ Missing'}</li>
            <li>nextra-search: ${contentAnalysis.hasNextraElements.search ? '✅ Found' : '❌ Missing'}</li>
        </ul>
    </div>

    <div class="section">
        <h2>💡 Recommendations</h2>
        <ul>
            ${!contentAnalysis.hasVisibleContent ? '<li>❌ Page appears to be empty - check if documentation content is properly loaded</li>' : ''}
            ${!contentAnalysis.hasNextraElements.container ? '<li>❌ Nextra container not found - Nextra theme may not be properly initialized</li>' : ''}
            ${stylesheets.length === 0 ? '<li>❌ No stylesheets loaded - CSS may not be properly configured</li>' : ''}
            ${consoleErrors.length > 0 ? '<li>🔧 Fix JavaScript console errors</li>' : ''}
            ${networkErrors.length > 0 ? '<li>🌐 Fix network loading errors</li>' : ''}
            <li>🎯 Ensure Nextra 4.6.0 is properly configured and all dependencies are installed</li>
        </ul>
    </div>
</body>
</html>`;

    fs.writeFileSync('docs-visual-analysis-report.html', report);
    console.log('\n📊 Visual analysis report saved to: docs-visual-analysis-report.html');

  } catch (error) {
    console.error('❌ Screenshot analysis failed:', error.message);

    // Try to capture a screenshot even if loading failed
    try {
      await page.screenshot({
        path: 'error-screenshot.png',
        fullPage: true
      });
      console.log('📸 Error screenshot saved: error-screenshot.png');
    } catch (screenshotError) {
      console.error('❌ Could not capture error screenshot:', screenshotError.message);
    }

    throw error;
  } finally {
    await browser.close();
  }
}

// Run the analysis
if (require.main === module) {
  captureScreenshots()
    .then(() => console.log('\n✅ Screenshot analysis completed!'))
    .catch(error => {
      console.error('\n❌ Analysis failed:', error.message);
      process.exit(1);
    });
}

module.exports = captureScreenshots;