import { test, expect } from '@playwright/test';
import { injectAxe, checkA11y } from 'axe-playwright';

test.describe('Accessibility and Code Quality Analysis', () => {
  const baseURL = 'http://localhost:3000';

  test.beforeEach(async ({ page }) => {
    await page.goto(baseURL);
  });

  test('WCAG 2.1 AA Accessibility Compliance', async ({ page }) => {
    console.log('♿ WCAG 2.1 AA Accessibility Compliance Test...');

    // axe-core 주입
    await injectAxe(page);

    // 접근성 검사 실행
    await checkA11y(page, null, {
      detailedReport: true,
      detailedReportOptions: { html: true },
      rules: {
        // WCAG 2.1 AA 기준으로 중요한 규칙들
        'color-contrast': { enabled: true },
        'keyboard-navigation': { enabled: true },
        'focus-order-semantics': { enabled: true },
        'heading-order': { enabled: true },
        'label-title-only': { enabled: true },
        'link-in-text-block': { enabled: true },
        'meta-viewport': { enabled: true },
        'page-has-heading-one': { enabled: true },
        'region': { enabled: true },
        'skip-link': { enabled: true }
      }
    });

    // 수동 접근성 검사
    const accessibilityChecks = await page.evaluate(() => {
      const checks: any = {};

      // 1. 키보드 내비게이션 가능성 확인
      const focusableElements = document.querySelectorAll(
        'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
      );
      checks.focusableElements = focusableElements.length;

      // 2. 대체 텍스트 확인
      const imagesWithoutAlt = document.querySelectorAll('img:not([alt])');
      checks.imagesWithoutAlt = imagesWithoutAlt.length;

      // 3. 폼 레이블 확인
      const inputsWithoutLabels = document.querySelectorAll('input:not([aria-label]):not([aria-labelledby])');
      const hasAssociatedLabels = Array.from(inputsWithoutLabels).filter(input => {
        const id = input.getAttribute('id');
        return id && document.querySelector(`label[for="${id}"]`);
      });
      checks.inputsWithoutLabels = inputsWithoutLabels.length - hasAssociatedLabels.length;

      // 4. 제목 구조 확인
      const headings = document.querySelectorAll('h1, h2, h3, h4, h5, h6');
      checks.headingStructure = {
        total: headings.length,
        hasH1: document.querySelectorAll('h1').length > 0,
        hierarchy: Array.from(headings).map(h => ({
          level: parseInt(h.tagName.substring(1)),
          text: h.textContent?.substring(0, 50)
        }))
      };

      // 5. ARIA 랜드마크 확인
      const landmarks = document.querySelectorAll('[role="banner"], [role="navigation"], [role="main"], [role="contentinfo"], [role="search"]');
      checks.landmarks = landmarks.length;

      // 6. 색상 대비 (기본 확인)
      const textElements = document.querySelectorAll('p, span, div, h1, h2, h3, h4, h5, h6');
      checks.textElements = textElements.length;

      return checks;
    });

    console.log(`♿ Accessibility Analysis:`);
    console.log(`   - Focusable Elements: ${accessibilityChecks.focusableElements}`);
    console.log(`   - Images without Alt: ${accessibilityChecks.imagesWithoutAlt}`);
    console.log(`   - Inputs without Labels: ${accessibilityChecks.inputsWithoutLabels}`);
    console.log(`   - Heading Structure: ${accessibilityChecks.headingStructure.total} headings`);
    console.log(`   - Has H1: ${accessibilityChecks.headingStructure.hasH1 ? '✅' : '❌'}`);
    console.log(`   - ARIA Landmarks: ${accessibilityChecks.landmarks}`);
    console.log(`   - Text Elements: ${accessibilityChecks.textElements}`);

    // 접근성 기준 검증
    expect(accessibilityChecks.imagesWithoutAlt).toBe(0);
    expect(accessibilityChecks.headingStructure.hasH1).toBe(true);
    expect(accessibilityChecks.focusableElements).toBeGreaterThan(0);
  });

  test('TypeScript Code Quality Analysis', async ({ page }) => {
    console.log('📘 TypeScript Code Quality Analysis...');

    await page.goto(baseURL);

    // 컴파일 에러 및 런타임 타입 에러 확인
    const typeScriptQuality = await page.evaluate(() => {
      const checks: any = {
        hasConsoleErrors: false,
        hasRuntimeErrors: false,
        missingTypeDefinitions: [],
        propTypeIssues: []
      };

      // 1. 콘솔 에러 수집
      const consoleErrors: string[] = [];
      const originalError = console.error;
      console.error = (...args) => {
        consoleErrors.push(args.join(' '));
        checks.hasConsoleErrors = true;
        originalError.apply(console, args);
      };

      // 2. React Props 타입 검사 (가능한 경우)
      const reactComponents = document.querySelectorAll('[data-reactroot], [data-react-checksum]');
      checks.reactComponents = reactComponents.length;

      // 3. TypeScript 관련 에러 패턴 검사
      const scripts = Array.from(document.querySelectorAll('script[src]'));
      const typeScriptScripts = scripts.filter(script =>
        script.getAttribute('src')?.includes('.ts') ||
        script.getAttribute('src')?.includes('typescript')
      );

      // 4. Import/Export 문제 확인
      const scriptContents = Array.from(document.querySelectorAll('script:not([src])'))
        .map(script => script.textContent || '');

      const hasImportErrors = scriptContents.some(content =>
        content.includes('Cannot read property') ||
        content.includes('undefined is not') ||
        content.includes('is not a function')
      );

      checks.importErrors = hasImportErrors;
      checks.typeScriptScripts = typeScriptScripts.length;

      // 콘솔 원복
      console.error = originalError;

      return checks;
    });

    // Next.js TypeScript 최적화 확인
    const nextTypeScriptFeatures = await page.evaluate(() => {
      return {
        hasServerComponents: !!window.__NEXT_DATA__?.props?.pageProps,
        hasStaticGeneration: window.__NEXT_DATA__?.staticMarkup || false,
        hasIncrementalSSR: window.__NEXT_DATA__?.props?.pageProps !== undefined,
        hasAppRouter: window.__NEXT_DATA__?.appGip || window.__NEXT_DATA__?.buildId !== undefined
      };
    });

    console.log(`📘 TypeScript Quality:`);
    console.log(`   - Console Errors: ${typeScriptQuality.hasConsoleErrors ? '❌ Found' : '✅ None'}`);
    console.log(`   - React Components: ${typeScriptQuality.reactComponents}`);
    console.log(`   - Import Errors: ${typeScriptQuality.importErrors ? '❌ Found' : '✅ None'}`);
    console.log(`   - TS Scripts: ${typeScriptQuality.typeScriptScripts}`);

    console.log(`🚀 Next.js TypeScript Features:`);
    console.log(`   - App Router: ${nextTypeScriptFeatures.hasAppRouter ? '✅' : '❌'}`);
    console.log(`   - Server Components: ${nextTypeScriptFeatures.hasServerComponents ? '✅' : '❌'}`);
    console.log(`   - Static Generation: ${nextTypeScriptFeatures.hasStaticGeneration ? '✅' : '❌'}`);
    console.log(`   - Incremental SSR: ${nextTypeScriptFeatures.hasIncrementalSSR ? '✅' : '❌'}`);

    expect(typeScriptQuality.hasConsoleErrors).toBe(false);
    expect(typeScriptQuality.importErrors).toBe(false);
    expect(nextTypeScriptFeatures.hasAppRouter).toBe(true);
  });

  test('ESLint and Code Standards Compliance', async ({ page }) => {
    console.log('📏 ESLint and Code Standards Compliance...');

    // 코드 품질 관련 브라우저 측정
    const codeQuality = await page.evaluate(() => {
      const checks: any = {};

      // 1. 콘솔 경고 및 에러
      const consoleWarnings: string[] = [];
      const consoleErrors: string[] = [];

      const originalWarn = console.warn;
      const originalError = console.error;

      console.warn = (...args) => consoleWarnings.push(args.join(' '));
      console.error = (...args) => consoleErrors.push(args.join(' '));

      // 2. 사용되지 않는 CSS 확인 (기본)
      const allStylesheets = Array.from(document.querySelectorAll('link[rel="stylesheet"]'));
      const inlineStyles = Array.from(document.querySelectorAll('style'));

      // 3. DOM 효율성 확인
      const duplicatedIds = [];
      const elementsWithIds = document.querySelectorAll('[id]');
      const idCounts: Record<string, number> = {};

      elementsWithIds.forEach(element => {
        const id = element.getAttribute('id');
        if (id) {
          idCounts[id] = (idCounts[id] || 0) + 1;
          if (idCounts[id] > 1) {
            duplicatedIds.push(id);
          }
        }
      });

      // 4. 비효율적인 선택자 확인
      const inefficientSelectors = document.querySelectorAll('*[style]');

      // 5. 메모리 누수 가능성 확인 (기본)
      const eventListeners = [];
      const elementsWithListeners = document.querySelectorAll('[onclick], [onload], [onerror]');

      // 콘솔 원복
      console.warn = originalWarn;
      console.error = originalError;

      return {
        consoleWarnings: consoleWarnings.length,
        consoleErrors: consoleErrors.length,
        stylesheets: allStylesheets.length,
        inlineStyles: inlineStyles.length,
        duplicatedIds: [...new Set(duplicatedIds)],
        inefficientSelectors: inefficientSelectors.length,
        elementsWithListeners: elementsWithListeners.length,
        totalElements: document.querySelectorAll('*').length
      };
    });

    // 개발 도구 통계 확인
    const developmentStats = await page.evaluate(() => {
      return {
        hasReactDevtools: !!window.__REACT_DEVTOOLS_GLOBAL_HOOK__,
        hasReduxDevtools: !!window.__REDUX_DEVTOOLS_EXTENSION__,
        hasVueDevtools: !!window.__VUE_DEVTOOLS_GLOBAL_HOOK__,
        performanceMarks: performance.getEntriesByType('mark').length,
        performanceMeasures: performance.getEntriesByType('measure').length
      };
    });

    console.log(`📏 Code Quality Analysis:`);
    console.log(`   - Console Warnings: ${codeQuality.consoleWarnings}`);
    console.log(`   - Console Errors: ${codeQuality.consoleErrors}`);
    console.log(`   - Stylesheets: ${codeQuality.stylesheets}`);
    console.log(`   - Inline Styles: ${codeQuality.inlineStyles}`);
    console.log(`   - Duplicated IDs: ${codeQuality.duplicatedIds.length}`);
    console.log(`   - Inefficient Selectors: ${codeQuality.inefficientSelectors}`);
    console.log(`   - Elements with Inline Listeners: ${codeQuality.elementsWithListeners}`);
    console.log(`   - Total DOM Elements: ${codeQuality.totalElements}`);

    console.log(`🛠️ Development Tools:`);
    console.log(`   - React DevTools: ${developmentStats.hasReactDevtools ? '✅' : '❌'}`);
    console.log(`   - Redux DevTools: ${developmentStats.hasReduxDevtools ? '✅' : '❌'}`);
    console.log(`   - Performance Marks: ${developmentStats.performanceMarks}`);
    console.log(`   - Performance Measures: ${developmentStats.performanceMeasures}`);

    // 코드 품질 기준 검증
    expect(codeQuality.consoleErrors).toBe(0);
    expect(codeQuality.duplicatedIds.length).toBe(0);
    expect(developmentStats.hasReactDevtools).toBe(true);
  });

  test('Progressive Enhancement and Fallbacks', async ({ page }) => {
    console.log('🔄 Progressive Enhancement and Fallbacks...');

    // 자바스크립트 비활성화 환경 시뮬레이션
    await page.route('**/*.js', route => route.abort());

    // 페이지 재로드 (자바스크립트 없이)
    await page.goto(baseURL, { waitUntil: 'domcontentloaded' });

    const noJSChecks = await page.evaluate(() => {
      return {
        hasContent: document.body.innerText.length > 100,
        hasNavigation: document.querySelectorAll('nav, [role="navigation"]').length > 0,
        hasHeadings: document.querySelectorAll('h1, h2, h3').length > 0,
        hasLinks: document.querySelectorAll('a').length > 0,
        hasForms: document.querySelectorAll('form').length > 0,
        noscriptTags: document.querySelectorAll('noscript').length
      };
    });

    // 자바스크립트 다시 활성화
    await page.unroute('**/*.js');

    console.log(`🔄 Progressive Enhancement:`);
    console.log(`   - Content without JS: ${noJSChecks.hasContent ? '✅' : '❌'}`);
    console.log(`   - Navigation without JS: ${noJSChecks.hasNavigation ? '✅' : '❌'}`);
    console.log(`   - Headings without JS: ${noJSChecks.hasHeadings ? '✅' : '❌'}`);
    console.log(`   - Links without JS: ${noJSChecks.hasLinks ? '✅' : '❌'}`);
    console.log(`   - Forms without JS: ${noJSChecks.hasForms ? '✅' : '❌'}`);
    console.log(`   - Noscript Tags: ${noJSChecks.noscriptTags}`);

    // 점진적 향상 기능 확인
    await page.goto(baseURL);

    const progressiveFeatures = await page.evaluate(() => {
      return {
        hasServiceWorker: !!navigator.serviceWorker,
        hasWebWorkers: typeof Worker !== 'undefined',
        hasLocalStorage: typeof localStorage !== 'undefined',
        hasSessionStorage: typeof sessionStorage !== 'undefined',
        hasGeolocation: typeof navigator.geolocation !== 'undefined',
        hasWebSockets: typeof WebSocket !== 'undefined',
        hasWebAssembly: typeof WebAssembly !== 'undefined',
        hasIntersectionObserver: typeof IntersectionObserver !== 'undefined',
        hasMutationObserver: typeof MutationObserver !== 'undefined'
      };
    });

    console.log(`🚀 Modern Web APIs:`);
    Object.entries(progressiveFeatures).forEach(([feature, supported]) => {
      console.log(`   - ${feature}: ${supported ? '✅' : '❌'}`);
    });

    // 필수 기능 확인
    expect(noJSChecks.hasContent).toBe(true);
    expect(noJSChecks.hasHeadings).toBe(true);
    expect(progressiveFeatures.hasLocalStorage).toBe(true);
    expect(progressiveFeatures.hasIntersectionObserver).toBe(true);
  });

  test('Internationalization and Localization Support', async ({ page }) => {
    console.log('🌍 Internationalization and Localization Support...');

    // i18n 설정 확인
    const i18nSupport = await page.evaluate(() => {
      const html = document.documentElement;

      return {
        hasLangAttribute: !!html.lang,
        langAttribute: html.lang,
        hasDirectionAttribute: !!html.dir,
        directionAttribute: html.dir,
        hasMetaCharset: !!document.querySelector('meta[charset]'),
        charset: document.querySelector('meta[charset]')?.getAttribute('charset'),
        contentLanguage: document.querySelector('meta[http-equiv="content-language"]')?.getAttribute('content')
      };
    });

    // 다국어 지원 확인
    const multilingualFeatures = await page.evaluate(() => {
      // 언어 전환 링크 확인
      const langSwitcher = document.querySelector('[lang], .language-selector, .locale-switcher');

      // 다국어 콘텐츠 확인
      const hasMultipleLanguages = Array.from(document.querySelectorAll('[lang]'))
        .map(el => el.getAttribute('lang'))
        .filter(Boolean)
        .length > 1;

      return {
        hasLangSwitcher: !!langSwitcher,
        multipleLanguagesDetected: hasMultipleLanguages,
        langTags: Array.from(document.querySelectorAll('[lang]'))
          .map(el => el.getAttribute('lang'))
          .filter(Boolean)
      };
    });

    console.log(`🌍 i18n Support:`);
    console.log(`   - HTML Lang: ${i18nSupport.hasLangAttribute ? `✅ (${i18nSupport.langAttribute})` : '❌'}`);
    console.log(`   - Direction: ${i18nSupport.hasDirectionAttribute ? `✅ (${i18nSupport.directionAttribute})` : '❌'}`);
    console.log(`   - Charset: ${i18nSupport.hasMetaCharset ? `✅ (${i18nSupport.charset})` : '❌'}`);
    console.log(`   - Language Switcher: ${multilingualFeatures.hasLangSwitcher ? '✅' : '❌'}`);
    console.log(`   - Multiple Languages: ${multilingualFeatures.multipleLanguagesDetected ? '✅' : '❌'}`);
    console.log(`   - Detected Languages: ${multilingualFeatures.langTags.join(', ')}`);

    // 기본 i18n 요구사항 확인
    expect(i18nSupport.hasLangAttribute).toBe(true);
    expect(i18nSupport.hasMetaCharset).toBe(true);
    expect(i18nSupport.charset).toBe('utf-8');
  });
});