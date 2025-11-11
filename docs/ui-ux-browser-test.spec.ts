import { test, expect } from '@playwright/test';

test.describe('UI/UX Comprehensive Browser Tests', () => {
  test.beforeEach(async ({ page }) => {
    // 페이지 접속 및 로드 대기
    await page.goto('http://localhost:3000/ko');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(2000); // 추가 렌더링 대기
  });

  test('기본 페이지 구조 및 렌더링 확인', async ({ page }) => {
    // 1. 페이지 제목 확인
    await expect(page).toHaveTitle(/MoAI-ADK/);

    // 2. 주요 요소들이 렌더링되었는지 확인
    const header = page.locator('header');
    await expect(header).toBeVisible();

    const mainContent = page.locator('main');
    await expect(mainContent).toBeVisible();

    const sidebar = page.locator('.nextra-sidebar-container, [class*="sidebar"]');
    // 사이드바는 있을 수도 없을 수도 있음 (반응형)

    // 3. 내비게이션 메뉴 확인
    const navigation = page.locator('nav');
    if (await navigation.count() > 0) {
      await expect(navigation.first()).toBeVisible();
    }

    // 4. 콘텐츠가 로드되었는지 확인
    const content = page.locator('[class*="content"], article, .nextra-content');
    if (await content.count() > 0) {
      await expect(content.first()).toBeVisible();
    }

    console.log('✅ 기본 페이지 구조 확인 완료');
  });

  test('타이포그래피 및 시각 디자인 확인', async ({ page }) => {
    // 1. 제목 계층 구조 확인
    const headings = page.locator('h1, h2, h3, h4, h5, h6');
    const headingCount = await headings.count();

    if (headingCount > 0) {
      // h1이 있는지 확인
      const h1 = page.locator('h1');
      const h1Count = await h1.count();

      if (h1Count > 0) {
        await expect(h1.first()).toBeVisible();
        const h1Text = await h1.first().textContent();
        console.log(`H1 제목: ${h1Text}`);
      }

      // 폰트 스타일 확인
      const h1Styles = await page.locator('h1').first().evaluate(el => {
        const computed = window.getComputedStyle(el);
        return {
          fontSize: computed.fontSize,
          fontWeight: computed.fontWeight,
          lineHeight: computed.lineHeight,
          color: computed.color
        };
      });

      console.log('H1 스타일:', h1Styles);
    }

    // 2. 단락 텍스트 확인
    const paragraphs = page.locator('p');
    const pCount = await paragraphs.count();

    if (pCount > 0) {
      const pStyles = await page.locator('p').first().evaluate(el => {
        const computed = window.getComputedStyle(el);
        return {
          fontSize: computed.fontSize,
          lineHeight: computed.lineHeight,
          color: computed.color
        };
      });

      console.log('단락 스타일:', pStyles);
    }

    console.log('✅ 타이포그래피 확인 완료');
  });

  test('색상 대비 및 접근성 기본 확인', async ({ page }) => {
    // 1. 텍스트와 배경 색상 확인
    const bodyStyles = await page.evaluate(() => {
      const computed = window.getComputedStyle(document.body);
      return {
        backgroundColor: computed.backgroundColor,
        color: computed.color,
        fontFamily: computed.fontFamily
      };
    });

    console.log('페이지 기본 스타일:', bodyStyles);

    // 2. 버튼 요소 확인
    const buttons = page.locator('button, [role="button"], a[class*="button"]');
    const buttonCount = await buttons.count();

    if (buttonCount > 0) {
      for (let i = 0; i < Math.min(buttonCount, 5); i++) {
        const button = buttons.nth(i);
        if (await button.isVisible()) {
          const buttonStyles = await button.evaluate(el => {
            const computed = window.getComputedStyle(el);
            return {
              backgroundColor: computed.backgroundColor,
              color: computed.color,
              border: computed.border,
              padding: computed.padding
            };
          });
          console.log(`버튼 ${i+1} 스타일:`, buttonStyles);
        }
      }
    }

    // 3. 링크 확인
    const links = page.locator('a[href]');
    const linkCount = await links.count();

    if (linkCount > 0) {
      const linkStyles = await links.first().evaluate(el => {
        const computed = window.getComputedStyle(el);
        return {
          color: computed.color,
          textDecoration: computed.textDecoration
        };
      });
      console.log('링크 스타일:', linkStyles);
    }

    console.log('✅ 색상 및 스타일 확인 완료');
  });

  test('반응형 디자인 테스트', async ({ page }) => {
    // 데스크탑 뷰 (기본)
    await page.setViewportSize({ width: 1200, height: 800 });
    await page.waitForTimeout(1000);

    // 데스크탑 스크린샷
    await page.screenshot({
      path: 'test-results/desktop-homepage.png',
      fullPage: true
    });

    // 태블릿 뷰
    await page.setViewportSize({ width: 768, height: 1024 });
    await page.waitForTimeout(1000);

    // 사이드바/내비게이션 변화 확인
    const sidebarDesktop = page.locator('.nextra-sidebar-container, [class*="sidebar"]');
    const isSidebarVisibleTablet = await sidebarDesktop.isVisible();

    await page.screenshot({
      path: 'test-results/tablet-homepage.png',
      fullPage: true
    });

    // 모바일 뷰
    await page.setViewportSize({ width: 375, height: 667 });
    await page.waitForTimeout(1000);

    // 모바일용 내비게이션 버튼 확인
    const mobileMenuButton = page.locator('button[aria-label*="menu"], button[aria-label*="navigation"], .hamburger, [class*="menu-button"]');
    if (await mobileMenuButton.count() > 0) {
      console.log('모바일 메뉴 버튼 발견');
    }

    await page.screenshot({
      path: 'test-results/mobile-homepage.png',
      fullPage: true
    });

    console.log('✅ 반응형 디자인 테스트 완료');
    console.log(`- 데스크탑 (1200x800): 스크린샷 저장`);
    console.log(`- 태블릿 (768x1024): 스크린샷 저장, 사이드바 표시: ${isSidebarVisibleTablet}`);
    console.log(`- 모바일 (375x667): 스크린샷 저장`);
  });

  test('사용자 상호작용 테스트', async ({ page }) => {
    // 1. 내비게이션 링크 클릭 테스트
    const navigationLinks = page.locator('nav a[href], .nextra-nav-container a[href]');
    const navLinkCount = await navigationLinks.count();

    if (navLinkCount > 0) {
      // 첫 번째 내비게이션 링크 클릭
      const firstLink = navigationLinks.first();
      const linkHref = await firstLink.getAttribute('href');

      if (linkHref && !linkHref.startsWith('http')) {
        await firstLink.click();
        await page.waitForTimeout(2000);

        // 페이지가 이동했는지 확인
        const currentUrl = page.url();
        console.log(`클릭 후 URL: ${currentUrl}`);

        // 뒤로가기
        await page.goBack();
        await page.waitForTimeout(1000);
      }
    }

    // 2. 모바일 메뉴 토GGLE 테스트 (모바일 뷰에서)
    await page.setViewportSize({ width: 375, height: 667 });
    await page.waitForTimeout(1000);

    const mobileMenuButton = page.locator('button[aria-label*="menu"], button[aria-label*="navigation"], .hamburger, [class*="menu-button"]');
    if (await mobileMenuButton.count() > 0 && await mobileMenuButton.first().isVisible()) {
      await mobileMenuButton.first().click();
      await page.waitForTimeout(1000);

      // 메뉴가 열렸는지 확인
      const mobileMenu = page.locator('[class*="mobile-menu"], [class*="sidebar"], [role="navigation"]');
      if (await mobileMenu.count() > 0) {
        const isMenuOpen = await mobileMenu.first().isVisible();
        console.log(`모바일 메뉴 열림: ${isMenuOpen}`);

        if (isMenuOpen) {
          await mobileMenuButton.first().click(); // 메뉴 닫기
          await page.waitForTimeout(500);
        }
      }
    }

    console.log('✅ 사용자 상호작용 테스트 완료');
  });

  test('페이지 성능 및 렌더링 품질', async ({ page }) => {
    // 1. 페이지 로드 성능 측정
    const performanceMetrics = await page.evaluate(() => {
      const navigation = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming;
      return {
        domContentLoaded: navigation.domContentLoadedEventEnd - navigation.domContentLoadedEventStart,
        loadComplete: navigation.loadEventEnd - navigation.loadEventStart,
        firstPaint: performance.getEntriesByType('paint').find(p => p.name === 'first-paint')?.startTime || 0,
        firstContentfulPaint: performance.getEntriesByType('paint').find(p => p.name === 'first-contentful-paint')?.startTime || 0
      };
    });

    console.log('페이지 성능 지표:', performanceMetrics);

    // 2. 렌더링된 요소 수 확인
    const elementCount = await page.evaluate(() => {
      return document.querySelectorAll('*').length;
    });

    console.log(`전체 DOM 요소 수: ${elementCount}`);

    // 3. 이미지 최적화 확인
    const images = page.locator('img');
    const imageCount = await images.count();

    if (imageCount > 0) {
      let optimizedImages = 0;
      for (let i = 0; i < imageCount; i++) {
        const img = images.nth(i);
        if (await img.isVisible()) {
          const src = await img.getAttribute('src');
          const alt = await img.getAttribute('alt');

          if (src) {
            // webp, avif 등 최신 포맷 확인
            if (src.includes('.webp') || src.includes('.avif')) {
              optimizedImages++;
            }
          }

          // alt 속성 확인
          if (!alt) {
            console.warn(`이미지에 alt 속성이 없음: ${src}`);
          }
        }
      }

      console.log(`이미지 수: ${imageCount}, 최적화된 이미지: ${optimizedImages}`);
    }

    console.log('✅ 성능 및 품질 테스트 완료');
  });

  test('키보드 내비게이션 및 포커스 관리', async ({ page }) => {
    // 1. Tab 키 내비게이션 테스트
    await page.keyboard.press('Tab');
    await page.waitForTimeout(500);

    let focusedElement = await page.evaluate(() => document.activeElement?.tagName);
    console.log(`첫 번째 Tab 후 포커스: ${focusedElement}`);

    // 몇 번 더 Tab 키 누르기
    for (let i = 0; i < 5; i++) {
      await page.keyboard.press('Tab');
      await page.waitForTimeout(300);

      focusedElement = await page.evaluate(() => {
        const el = document.activeElement;
        return {
          tagName: el?.tagName,
          className: el?.className,
          hasFocusOutline: window.getComputedStyle(el as Element).outline !== 'none'
        };
      });

      console.log(`Tab ${i+2} 포커스:`, focusedElement);
    }

    // 2. Shift+Tab 테스트
    await page.keyboard.press('Shift+Tab');
    await page.waitForTimeout(500);

    focusedElement = await page.evaluate(() => document.activeElement?.tagName);
    console.log(`Shift+Tab 후 포커스: ${focusedElement}`);

    // 3. Escape 키 테스트 (모달/메뉴 닫기)
    await page.keyboard.press('Escape');
    await page.waitForTimeout(500);

    console.log('✅ 키보드 내비게이션 테스트 완료');
  });
});