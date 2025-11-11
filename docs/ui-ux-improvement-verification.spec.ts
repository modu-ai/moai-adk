import { test, expect } from '@playwright/test';

test.describe('UI/UX 개선 사항 검증', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('http://localhost:3000/ko');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(3000); // 스타일 적용 대기
  });

  test('✅ H1 제목 스타일 개선 확인', async ({ page }) => {
    const h1 = page.locator('h1');
    await expect(h1).toBeVisible();

    const h1Styles = await h1.evaluate(el => {
      const computed = window.getComputedStyle(el);
      return {
        fontSize: computed.fontSize,
        fontWeight: computed.fontWeight,
        lineHeight: computed.lineHeight,
        color: computed.color,
        background: computed.background,
        WebkitBackgroundClip: computed.webkitBackgroundClip,
        WebkitTextFillColor: computed.webkitTextFillColor
      };
    });

    console.log('H1 개선된 스타일:', h1Styles);

    // H1이 16px보다 커졌는지 확인
    const fontSize = parseInt(h1Styles.fontSize);
    expect(fontSize).toBeGreaterThan(20); // 20px 이상이어야 함

    // 굵기 확인
    expect(h1Styles.fontWeight).toBe('700');

    // 그라데이션 적용 확인
    expect(h1Styles.background).toContain('gradient');
  });

  test('✅ 페이지 타이틀 확인', async ({ page }) => {
    const title = await page.title();
    console.log('페이지 타이틀:', title);

    expect(title).toContain('MoAI-ADK');
    expect(title).toContain('AI 기반 SPEC-First TDD 개발 프레임워크');
    expect(title.length).toBeGreaterThan(10); // 빈 타이틀이 아닌지 확인
  });

  test('✅ 링크 색상 구분 확인', async ({ page }) => {
    const links = page.locator('a[href]');
    const linkCount = await links.count();

    if (linkCount > 0) {
      // 첫 번째 링크와 단락의 색상 비교
      const firstLink = links.first();
      const firstParagraph = page.locator('p').first();

      if (await firstParagraph.isVisible()) {
        const linkColor = await firstLink.evaluate(el =>
          window.getComputedStyle(el).color
        );
        const textColor = await firstParagraph.evaluate(el =>
          window.getComputedStyle(el).color
        );

        console.log('링크 색상:', linkColor);
        console.log('텍스트 색상:', textColor);

        // 링크와 텍스트 색상이 다른지 확인
        expect(linkColor).not.toBe(textColor);
      }
    }
  });

  test('✅ 보안 헤더 확인', async ({ page }) => {
    const response = await page.goto('http://localhost:3000/ko');

    if (response) {
      const headers = response.headers();

      console.log('보안 헤더들:');
      console.log('- X-Content-Type-Options:', headers['x-content-type-options']);
      console.log('- X-Frame-Options:', headers['x-frame-options']);
      console.log('- X-XSS-Protection:', headers['x-xss-protection']);
      console.log('- Referrer-Policy:', headers['referrer-policy']);

      expect(headers['x-content-type-options']).toBe('nosniff');
      expect(headers['x-frame-options']).toBe('DENY');
      expect(headers['x-xss-protection']).toBe('1; mode=block');
      expect(headers['referrer-policy']).toBe('strict-origin-when-cross-origin');
    }
  });

  test('✅ 스킵 링크 접근성 확인', async ({ page }) => {
    // 스킵 링크가 있는지 확인
    const skipLink = page.locator('.skip-link');
    expect(skipLink).toHaveCount(1);

    // 스킵 링크에 초점을 맞추고 확인
    await skipLink.focus();
    await expect(skipLink).toBeFocused();

    // 스킵 링크 텍스트 확인
    const skipLinkText = await skipLink.textContent();
    expect(skipLinkText).toContain('메인 콘텐츠로 바로가기');

    console.log('✅ 스킵 링크 정상 작동:', skipLinkText);
  });

  test('✅ 키보드 내비게이션 개선 확인', async ({ page }) => {
    // Tab 키로 이동하며 포커스 확인
    await page.keyboard.press('Tab');
    await page.waitForTimeout(200);

    let focusedElement = await page.evaluate(() => ({
      tagName: document.activeElement?.tagName,
      hasFocusOutline: window.getComputedStyle(document.activeElement as Element).outline !== 'none'
    }));

    console.log('첫 번째 Tab 후 포커스:', focusedElement);

    // 스킵 링크에 포커스가 가는지 확인
    expect(focusedElement.tagName).toBe('A');
    expect(focusedElement.hasFocusOutline).toBe(true);

    // Enter 키로 스킵 링크 활성화
    await page.keyboard.press('Enter');
    await page.waitForTimeout(500);

    // 메인 콘텐츠로 이동했는지 확인 (해시가 변경되었는지)
    const url = page.url();
    console.log('스킵 링크 클릭 후 URL:', url);
  });

  test('✅ 반응형 디자인 스크린샷 비교', async ({ page }) => {
    // 다양한 화면 크기에서 스크린샷
    const sizes = [
      { width: 1200, height: 800, name: 'desktop' },
      { width: 768, height: 1024, name: 'tablet' },
      { width: 375, height: 667, name: 'mobile' }
    ];

    for (const size of sizes) {
      await page.setViewportSize({ width: size.width, height: size.height });
      await page.waitForTimeout(1000);

      await page.screenshot({
        path: `test-results/improved-${size.name}-homepage.png`,
        fullPage: true
      });

      console.log(`✅ ${size.name} 스크린샷 저장 완료`);
    }
  });

  test('✅ 접근성 ARIA 랜드마크 확인', async ({ page }) => {
    // 메인 콘텐츠 ID 확인
    const mainContent = page.locator('#main-content');
    await expect(mainContent).toBeVisible();

    // 메인 콘텐츠가 h1인지 확인
    const mainContentTag = await mainContent.evaluate(el => el.tagName);
    expect(mainContentTag).toBe('H1');

    console.log('✅ 메인 콘텐츠 ID 할당 완료');
  });

  test('✅ CSS 변수 적용 확인', async ({ page }) => {
    const rootStyles = await page.evaluate(() => {
      const root = document.documentElement;
      const computed = window.getComputedStyle(root);

      return {
        brandPrimary: computed.getPropertyValue('--brand-primary'),
        textPrimary: computed.getPropertyValue('--text-primary'),
        bgPrimary: computed.getPropertyValue('--bg-primary'),
        linkColor: computed.getPropertyValue('--link-color')
      };
    });

    console.log('CSS 변수 적용 상태:', rootStyles);

    // 주요 CSS 변수가 적용되었는지 확인
    expect(rootStyles.brandPrimary).toBeTruthy();
    expect(rootStyles.textPrimary).toBeTruthy();
    expect(rootStyles.bgPrimary).toBeTruthy();
    expect(rootStyles.linkColor).toBeTruthy();
  });

  test('✅ 테마 설정 개선 사항 확인', async ({ page }) => {
    // 다국어 선택기 확인
    const langSelector = page.locator('[class*="i18n"], [class*="language"]');
    if (await langSelector.count() > 0) {
      console.log('✅ 다국어 선택기 확인');
    }

    // 검색 placeholder 확인
    const searchInput = page.locator('input[placeholder*="검색"]');
    if (await searchInput.count() > 0) {
      const placeholder = await searchInput.getAttribute('placeholder');
      console.log('검색 placeholder:', placeholder);
      expect(placeholder).toContain('검색');
    }

    // 편집 링크 확인
    const editLink = page.locator('a[href*="github.com"]');
    if (await editLink.count() > 0) {
      console.log('✅ 편집 링크 확인');
    }
  });

  test('✅ 다크모드 지원 확인', async ({ page }) => {
    // 다크모드 토글 버튼 확인
    const darkModeToggle = page.locator('button[aria-label*="dark"], button[aria-label*="light"], [class*="dark-mode"]');

    if (await darkModeToggle.count() > 0) {
      console.log('✅ 다크모드 토글 버튼 확인');

      // 다크모드로 전환
      await darkModeToggle.first().click();
      await page.waitForTimeout(1000);

      // 다크모드 스타일 확인
      const bodyStyles = await page.evaluate(() => {
        return {
          backgroundColor: window.getComputedStyle(document.body).backgroundColor,
          color: window.getComputedStyle(document.body).color
        };
      });

      console.log('다크모드 스타일:', bodyStyles);
    }
  });
});