const { chromium } = require('playwright');

(async () => {
  console.log('🎭 Playwright 브라우저 테스트 시작');

  // 브라우저 시작
  const browser = await chromium.launch({
    headless: false, // GUI 모드로 실행
    slowMo: 1000 // 1초 지연으로 사용자가 볼 수 있게
  });

  const context = await browser.newContext({
    viewport: { width: 1280, height: 720 }
  });

  const page = await context.newPage();

  try {
    // 1. 문서 서버에 접속
    console.log('📖 MoAI-ADK 문서 서버 접속 중...');
    await page.goto('http://localhost:8080');
    await page.waitForLoadState('networkidle');

    // 스크린샷 찍기
    await page.screenshot({
      path: '/Users/goos/Moai/MoAI-ADK/docs-homepage.png',
      fullPage: true
    });
    console.log('✅ 홈페이지 스크린샷 저장 완료');

    // 2. 페이지 제목 확인
    const title = await page.title();
    console.log(`📄 페이지 제목: ${title}`);

    // 3. 주요 요소 확인
    const heading = await page.locator('h1').first().textContent();
    console.log(`🎯 주요 제목: ${heading}`);

    // 4. 내비게이션 테스트
    console.log('🧭 내비게이션 메뉴 확인...');
    const navLinks = await page.locator('nav a').count();
    console.log(`📋 내비게이션 링크 수: ${navLinks}`);

    // 5. 페이지 내용 스크롤
    console.log('📜 페이지 스크롤 테스트...');
    await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight));
    await page.waitForTimeout(2000);

    // 6. 페이지 하단 스크린샷
    await page.screenshot({
      path: '/Users/goos/Moai/MoAI-ADK/docs-bottom.png',
      fullPage: true
    });

    // 7. 링크 클릭 테스트 (있는 경우)
    const firstLink = page.locator('a').first();
    if (await firstLink.count() > 0) {
      console.log('🔗 첫 번째 링크 클릭 테스트...');
      await firstLink.click();
      await page.waitForLoadState('networkidle');

      await page.screenshot({
        path: '/Users/goos/Moai/MoAI-ADK/docs-first-link.png',
        fullPage: true
      });
      console.log('✅ 링크 클릭 테스트 완료');
    }

    console.log('✨ 테스트 완료!');

  } catch (error) {
    console.error('❌ 테스트 중 오류 발생:', error);
    await page.screenshot({
      path: '/Users/goos/Moai/MoAI-ADK/docs-error.png',
      fullPage: true
    });
  } finally {
    await browser.close();
    console.log('🔚 브라우저 닫기');
  }
})();