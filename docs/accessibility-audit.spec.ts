import { test, expect } from '@playwright/test';

test.describe('WCAG Accessibility Audit', () => {
  test('should perform comprehensive accessibility audit', async ({ page }) => {
    await page.goto('http://localhost:3000', { waitUntil: 'networkidle' });

    console.log('\n=== WCAG 2.1 AA ACCESSIBILITY AUDIT ===');

    // 1. COLOR CONTRAST ANALYSIS
    console.log('\n1. COLOR CONTRAST (WCAG 1.4.3)');

    const colorContrastElements = [
      { selector: 'h1', name: 'Main heading' },
      { selector: 'h2', name: 'Subheadings' },
      { selector: 'body', name: 'Body text' },
      { selector: 'a', name: 'Links' }
    ];

    for (const { selector, name } of colorContrastElements) {
      const elements = page.locator(selector);
      const count = await elements.count();

      if (count > 0) {
        // Get computed styles for the first element
        const styles = await elements.first().evaluate((el) => {
          const styles = window.getComputedStyle(el);
          return {
            backgroundColor: styles.backgroundColor,
            color: styles.color,
            fontSize: styles.fontSize
          };
        });

        console.log(`${name}: Background=${styles.backgroundColor}, Text=${styles.color}, Size=${styles.fontSize}`);

        // Basic color contrast check - look for obvious issues
        if (styles.color === 'rgb(128, 128, 128)' || styles.color === 'rgb(169, 169, 169)') {
          console.warn(`⚠️  ${name} may have insufficient color contrast`);
        }
      }
    }

    // 2. KEYBOARD NAVIGATION (WCAG 2.1.1)
    console.log('\n2. KEYBOARD NAVIGATION');

    // Test Tab navigation
    let tabbableElements = [];
    let currentElement = null;

    for (let i = 0; i < 20; i++) {
      await page.keyboard.press('Tab');
      const focused = page.locator(':focus');
      const hasFocus = await focused.count() > 0;

      if (hasFocus) {
        const tagName = await focused.evaluate(el => el.tagName);
        const text = await focused.evaluate(el => el.textContent?.substring(0, 30));
        const tabindex = await focused.getAttribute('tabindex');

        tabbableElements.push({ tagName, text, tabindex });
        currentElement = focused;
      } else {
        break;
      }
    }

    console.log(`Tabbable elements found: ${tabbableElements.length}`);
    tabbableElements.forEach((elem, i) => {
      console.log(`  ${i + 1}. ${elem.tagName} (${elem.tabindex}): "${elem.text}"`);
    });

    // Test if we can navigate back
    if (tabbableElements.length > 2) {
      await page.keyboard.press('Shift+Tab');
      await page.keyboard.press('Shift+Tab');
      const backFocused = page.locator(':focus');
      const hasBackFocus = await backFocused.count() > 0;

      if (hasBackFocus) {
        console.log('✓ Shift+Tab navigation working');
      } else {
        console.warn('⚠️  Shift+Tab navigation may have issues');
      }
    }

    // 3. FOCUS INDICATORS (WCAG 2.4.7)
    console.log('\n3. FOCUS INDICATORS');

    // Check focus styles
    if (currentElement) {
      const focusStyles = await currentElement.evaluate((el) => {
        const styles = window.getComputedStyle(el, ':focus');
        return {
          outline: styles.outline,
          outlineColor: styles.outlineColor,
          outlineWidth: styles.outlineWidth,
          boxShadow: styles.boxShadow
        };
      });

      console.log('Focus styles:', focusStyles);

      if (focusStyles.outline === 'none' && focusStyles.boxShadow === 'none') {
        console.warn('⚠️  Focus indicator may not be visible');
      } else {
        console.log('✓ Focus indicators present');
      }
    }

    // 4. SEMANTIC HTML (WCAG 1.3.1)
    console.log('\n4. SEMANTIC HTML STRUCTURE');

    const semanticElements = [
      { selector: 'main', role: 'Main content area' },
      { selector: 'nav', role: 'Navigation' },
      { selector: 'header', role: 'Header/banner' },
      { selector: 'footer', role: 'Footer/contentinfo' },
      { selector: 'section', role: 'Section' },
      { selector: 'article', role: 'Article' },
      { selector: 'aside', role: 'Sidebar/complementary' }
    ];

    for (const { selector, role } of semanticElements) {
      const elements = page.locator(selector);
      const count = await elements.count();

      if (count > 0) {
        console.log(`✓ ${role}: ${count} ${selector} element(s) found`);
      } else {
        console.log(`ℹ️  ${role}: No ${selector} elements found`);
      }
    }

    // Check heading structure
    const headings = await page.locator('h1, h2, h3, h4, h5, h6').all();
    console.log(`\nHeading structure (${headings.length} total):`);

    let headingLevels = [];
    for (let i = 0; i < headings.length; i++) {
      const level = await headings[i].evaluate(el => parseInt(el.tagName.substring(1)));
      const text = await headings[i].textContent();
      headingLevels.push(level);
      console.log(`  H${level}: "${text?.substring(0, 40)}..."`);
    }

    // Check for skipped heading levels
    let hasSkippedLevels = false;
    for (let i = 1; i < headingLevels.length; i++) {
      if (headingLevels[i] > headingLevels[i - 1] + 1) {
        console.warn(`⚠️  Skipped heading level: H${headingLevels[i - 1]} to H${headingLevels[i]}`);
        hasSkippedLevels = true;
      }
    }

    if (!hasSkippedLevels) {
      console.log('✓ Heading levels are sequential');
    }

    // 5. ARIA LABELS AND ROLES (WCAG 4.1.2)
    console.log('\n5. ARIA LABELS AND ROLES');

    // Check for elements with ARIA attributes
    const ariaElements = page.locator('[aria-label], [aria-labelledby], [role]');
    const ariaCount = await ariaElements.count();
    console.log(`Elements with ARIA attributes: ${ariaCount}`);

    if (ariaCount > 0) {
      for (let i = 0; i < Math.min(ariaCount, 10); i++) {
        const element = ariaElements.nth(i);
        const ariaLabel = await element.getAttribute('aria-label');
        const ariaLabelledby = await element.getAttribute('aria-labelledby');
        const role = await element.getAttribute('role');
        const tagName = await element.evaluate(el => el.tagName);

        if (ariaLabel || ariaLabelledby || role) {
          console.log(`  ${tagName}: aria-label="${ariaLabel}", aria-labelledby="${ariaLabelledby}", role="${role}"`);
        }
      }
    }

    // Check for buttons without accessible names
    const buttons = page.locator('button');
    const buttonCount = await buttons.count();
    let buttonsWithoutLabels = 0;

    for (let i = 0; i < buttonCount; i++) {
      const button = buttons.nth(i);
      const hasText = await button.evaluate(el => el.textContent?.trim());
      const hasAriaLabel = await button.getAttribute('aria-label');
      const hasAriaLabelledby = await button.getAttribute('aria-labelledby');

      if (!hasText && !hasAriaLabel && !hasAriaLabelledby) {
        buttonsWithoutLabels++;
      }
    }

    if (buttonsWithoutLabels > 0) {
      console.warn(`⚠️  ${buttonsWithoutLabels} button(s) without accessible labels`);
    } else {
      console.log('✓ All buttons have accessible labels');
    }

    // 6. ALTERNATIVE TEXT (WCAG 1.1.1)
    console.log('\n6. IMAGE ALTERNATIVE TEXT');

    const images = page.locator('img');
    const imageCount = await images.count();
    console.log(`Images found: ${imageCount}`);

    let imagesWithoutAlt = 0;
    for (let i = 0; i < imageCount; i++) {
      const img = images.nth(i);
      const alt = await img.getAttribute('alt');
      const src = await img.getAttribute('src');

      if (alt === null) {
        imagesWithoutAlt++;
        console.warn(`⚠️  Image missing alt text: ${src}`);
      }
    }

    if (imagesWithoutAlt === 0) {
      console.log('✓ All images have alt text');
    }

    // 7. FORM LABELS (WCAG 3.3.2)
    console.log('\n7. FORM LABELS');

    const inputs = page.locator('input, select, textarea');
    const inputCount = await inputs.count();
    console.log(`Form controls: ${inputCount}`);

    if (inputCount > 0) {
      let inputsWithoutLabels = 0;
      for (let i = 0; i < inputCount; i++) {
        const input = inputs.nth(i);
        const hasLabel = await input.evaluate(el => {
          const id = el.id;
          const hasAttributeLabel = el.getAttribute('aria-label') || el.getAttribute('aria-labelledby');
          const hasAssociatedLabel = id && document.querySelector(`label[for="${id}"]`);
          const hasParentLabel = el.closest('label');

          return !!(hasAttributeLabel || hasAssociatedLabel || hasParentLabel);
        });

        if (!hasLabel) {
          inputsWithoutLabels++;
          const inputType = await input.getAttribute('type') || input.evaluate(el => el.tagName);
          console.warn(`⚠️  Form control without label: ${inputType}`);
        }
      }

      if (inputsWithoutLabels === 0) {
        console.log('✓ All form controls have labels');
      }
    }

    // 8. RESPONSIVE DESIGN AND ZOOM (WCAG 1.4.10)
    console.log('\n8. RESPONSIVE DESIGN AND ZOOM');

    // Test 200% zoom
    await page.setViewportSize({ width: 640, height: 480 }); // Simulate 200% zoom
    await page.waitForTimeout(1000);

    const hasHorizontalScrollAtZoom = await page.evaluate(() => {
      return document.body.scrollWidth > window.innerWidth;
    });

    if (hasHorizontalScrollAtZoom) {
      console.warn('⚠️  Horizontal scrolling at 200% zoom');
    } else {
      console.log('✓ No horizontal scrolling at 200% zoom');
    }

    // Test mobile viewport
    await page.setViewportSize({ width: 375, height: 667 });
    await page.waitForTimeout(1000);

    const mobileAccessibility = await page.evaluate(() => {
      const touchTargets = document.querySelectorAll('a, button, input, select, textarea');
      let smallTouchTargets = 0;

      touchTargets.forEach(element => {
        const rect = element.getBoundingClientRect();
        const size = Math.min(rect.width, rect.height);
        if (size < 44) { // 44px minimum recommended touch target
          smallTouchTargets++;
        }
      });

      return {
        touchTargets: touchTargets.length,
        smallTouchTargets
      };
    });

    console.log(`Mobile touch targets: ${mobileAccessibility.touchTargets}`);
    if (mobileAccessibility.smallTouchTargets > 0) {
      console.warn(`⚠️  ${mobileAccessibility.smallTouchTargets} touch targets smaller than 44px`);
    } else {
      console.log('✓ Touch targets meet size recommendations');
    }

    // 9. LANGUAGE ATTRIBUTES (WCAG 3.1.1)
    console.log('\n9. LANGUAGE ATTRIBUTES');

    const htmlLang = await page.locator('html').getAttribute('lang');
    if (htmlLang) {
      console.log(`✓ HTML lang attribute: ${htmlLang}`);
    } else {
      console.warn('⚠️  Missing HTML lang attribute');
    }

    // Check for language changes in content
    const koreanText = page.locator('*:not(script):not(style)').filter({ hasText: /[가-힣]/ });
    const koreanCount = await koreanText.count();
    console.log(`Korean text elements: ${koreanCount}`);

    if (koreanCount > 0) {
      console.log('ℹ️  Korean content detected - consider lang attributes if mixed languages');
    }

    // 10. SUMMARY AND RECOMMENDATIONS
    console.log('\n=== ACCESSIBILITY SUMMARY ===');

    const totalChecks = 9;
    let passedChecks = 0;

    // Simple scoring based on our checks
    if (tabbableElements.length > 0) passedChecks++;
    if (headingLevels.length > 0 && !hasSkippedLevels) passedChecks++;
    if (buttonsWithoutLabels === 0) passedChecks++;
    if (imagesWithoutAlt === 0) passedChecks++;
    if (htmlLang) passedChecks++;
    if (!hasHorizontalScrollAtZoom) passedChecks++;
    if (mobileAccessibility.smallTouchTargets === 0) passedChecks++;

    const score = Math.round((passedChecks / totalChecks) * 100);
    console.log(`Accessibility Score: ${score}% (${passedChecks}/${totalChecks} checks passed)`);

    // Capture accessibility screenshot
    await page.setViewportSize({ width: 1280, height: 720 });
    await page.screenshot({
      path: 'test-results/accessibility-audit.png',
      fullPage: true
    });

    console.log('\n=== AUDIT COMPLETE ===');
  });
});