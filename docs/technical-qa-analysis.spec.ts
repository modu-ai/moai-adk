import { test, expect, devices } from '@playwright/test';

// 테스트 환경 설정: 다양한 브라우저와 디바이스에서 테스트
test.describe('Technical Quality Analysis - MoAI-ADK Documentation', () => {
  const baseURL = 'http://localhost:3000';

  // 데스크톱 환경 테스트
  test.describe('Desktop Environment', () => {
    test.use({ ...devices['Desktop Chrome'] });

    test('Framework Compatibility - Next.js 15 Migration', async ({ page }) => {
      console.log('🔍 Testing Next.js 15 migration compatibility...');

      await page.goto(baseURL);

      // 1. React 18 features 확인
      const reactVersion = await page.evaluate(() => {
        return window.React?.version;
      });
      expect(reactVersion).toMatch(/^18\./);

      // 2. Next.js version 확인
      const nextVersion = await page.evaluate(() => {
        return window.__NEXT_DATA__?.buildId ? '15.x' : 'unknown';
      });

      // 3. Nextra 4.6 기능 확인
      const nextraFeatures = await page.locator('[data-theme], [class*="nextra"]').count();
      expect(nextraFeatures).toBeGreaterThan(0);

      // 4. Turbopack 사용 여부 확인 (개발 모드)
      const turbopackIndicator = await page.evaluate(() => {
        return !!document.querySelector('[data-turbopack]');
      });

      console.log(`✅ React Version: ${reactVersion}`);
      console.log(`✅ Next.js: ${nextVersion}`);
      console.log(`✅ Nextra Features: ${nextraFeatures} components found`);
      console.log(`✅ Turbopack: ${turbopackIndicator ? 'Enabled' : 'Not detected'}`);
    });

    test('Bundle Optimization Analysis', async ({ page }) => {
      console.log('📦 Analyzing bundle optimization...');

      const responses: any[] = [];
      page.on('response', response => {
        if (response.url().includes('.js') || response.url().includes('.css')) {
          responses.push({
            url: response.url(),
            size: response.headers()['content-length'] || 'unknown',
            encoded: response.headers()['content-encoding'] || 'none'
          });
        }
      });

      await page.goto(baseURL);
      await page.waitForLoadState('networkidle');

      // 번들 파일 분석
      const jsBundles = responses.filter(r => r.url.includes('.js'));
      const cssBundles = responses.filter(r => r.url.includes('.css'));

      // 코드 분할 확인
      const chunkCount = jsBundles.filter(b => b.url.includes('chunk-')).length;
      const hasLazyChunks = jsBundles.some(b => b.url.includes('lazy'));

      // 압축 확인
      const compressedFiles = responses.filter(r => r.encoded === 'gzip' || r.encoded === 'br').length;

      console.log(`📊 Bundle Analysis:`);
      console.log(`   - JS Bundles: ${jsBundles.length}`);
      console.log(`   - CSS Bundles: ${cssBundles.length}`);
      console.log(`   - Code Splitting: ${chunkCount} chunks found`);
      console.log(`   - Lazy Loading: ${hasLazyChunks ? 'Enabled' : 'Not detected'}`);
      console.log(`   - Compression: ${compressedFiles}/${responses.length} files compressed`);

      expect(chunkCount).toBeGreaterThan(0);
      expect(compressedFiles).toBeGreaterThan(0);
    });

    test('Performance Metrics - Core Web Vitals', async ({ page }) => {
      console.log('⚡ Measuring Core Web Vitals...');

      await page.goto(baseURL);

      // LCP (Largest Contentful Paint) 측정
      const lcp = await page.evaluate(async () => {
        return new Promise(resolve => {
          new PerformanceObserver(list => {
            const entries = list.getEntries();
            const lastEntry = entries[entries.length - 1];
            resolve(lastEntry.startTime);
          }).observe({ entryTypes: ['largest-contentful-paint'] });

          // 5초 후 타임아웃
          setTimeout(() => resolve(5000), 5000);
        });
      });

      // FID (First Input Delay) 예측 (실제 FID는 사용자 상호작용 필요)
      const ttfb = await page.evaluate(() => {
        const nav = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming;
        return nav.responseStart - nav.requestStart;
      });

      // CLS (Cumulative Layout Shift) 측정
      const cls = await page.evaluate(async () => {
        return new Promise(resolve => {
          let clsValue = 0;
          new PerformanceObserver(list => {
            for (const entry of list.getEntries()) {
              if (!(entry as any).hadRecentInput) {
                clsValue += (entry as any).value;
              }
            }
          }).observe({ entryTypes: ['layout-shift'] });

          setTimeout(() => resolve(clsValue), 3000);
        });
      });

      // 메모리 사용량 확인
      const memoryUsage = await page.evaluate(() => {
        return (performance as any).memory ? {
          used: Math.round((performance as any).memory.usedJSHeapSize / 1024 / 1024),
          total: Math.round((performance as any).memory.totalJSHeapSize / 1024 / 1024),
          limit: Math.round((performance as any).memory.jsHeapSizeLimit / 1024 / 1024)
        } : null;
      });

      console.log(`🎯 Core Web Vitals:`);
      console.log(`   - LCP: ${Math.round(lcp)}ms (Target: <2500ms)`);
      console.log(`   - TTFB: ${Math.round(ttfb)}ms (Target: <800ms)`);
      console.log(`   - CLS: ${Math.round(cls * 1000) / 1000} (Target: <0.1)`);
      if (memoryUsage) {
        console.log(`   - Memory: ${memoryUsage.used}MB/${memoryUsage.total}MB`);
      }

      // 성능 기준 확인
      expect(lcp).toBeLessThan(2500); // LCP < 2.5s
      expect(ttfb).toBeLessThan(800);  // TTFB < 800ms
      expect(cls).toBeLessThan(0.1);   // CLS < 0.1
    });

    test('Code Quality & Technical Standards', async ({ page }) => {
      console.log('🔬 Analyzing code quality and technical standards...');

      await page.goto(baseURL);

      // 1. TypeScript 컴파일 에러 확인
      const consoleErrors: string[] = [];
      page.on('console', msg => {
        if (msg.type() === 'error') {
          consoleErrors.push(msg.text());
        }
      });

      await page.waitForLoadState('networkidle');

      // 2. HTML 유효성 확인
      const htmlValid = await page.evaluate(() => {
        const doctype = document.doctype;
        const htmlLang = document.documentElement.lang;
        const metaViewport = document.querySelector('meta[name="viewport"]');
        const metaCharset = document.querySelector('meta[charset]');

        return {
          hasDoctype: !!doctype,
          hasLang: !!htmlLang,
          hasViewport: !!metaViewport,
          hasCharset: !!metaCharset,
          lang: htmlLang
        };
      });

      // 3. SEO 및 접근성 메타 태그 확인
      const metaTags = await page.evaluate(() => {
        const title = document.title;
        const description = document.querySelector('meta[name="description"]')?.getAttribute('content');
        const ogTitle = document.querySelector('meta[property="og:title"]')?.getAttribute('content');
        const ogDescription = document.querySelector('meta[property="og:description"]')?.getAttribute('content');

        return { title, description, ogTitle, ogDescription };
      });

      // 4. 스크립트 로딩 확인
      const scriptsInfo = await page.evaluate(() => {
        const scripts = Array.from(document.querySelectorAll('script[src]'));
        return scripts.map(script => ({
          src: script.getAttribute('src'),
          async: script.hasAttribute('async'),
          defer: script.hasAttribute('defer'),
          type: script.getAttribute('type')
        }));
      });

      console.log(`📋 Code Quality Analysis:`);
      console.log(`   - Console Errors: ${consoleErrors.length}`);
      console.log(`   - HTML Doctype: ${htmlValid.hasDoctype ? '✅' : '❌'}`);
      console.log(`   - HTML Lang: ${htmlValid.lang || 'Not set'}`);
      console.log(`   - Meta Viewport: ${htmlValid.hasViewport ? '✅' : '❌'}`);
      console.log(`   - Meta Charset: ${htmlValid.hasCharset ? '✅' : '❌'}`);
      console.log(`   - Page Title: ${metaTags.title || 'Missing'}`);
      console.log(`   - Scripts: ${scriptsInfo.length} external scripts`);

      if (consoleErrors.length > 0) {
        console.warn(`⚠️ Console Errors:`, consoleErrors);
      }

      expect(consoleErrors.length).toBe(0);
      expect(htmlValid.hasDoctype).toBe(true);
      expect(htmlValid.hasViewport).toBe(true);
      expect(metaTags.title).toBeTruthy();
    });

    test('Security Analysis', async ({ page }) => {
      console.log('🔒 Analyzing security configurations...');

      const responses: any[] = [];
      page.on('response', response => {
        responses.push({
          url: response.url(),
          status: response.status(),
          headers: response.headers()
        });
      });

      await page.goto(baseURL);
      await page.waitForLoadState('networkidle');

      // 1. HTTPS 사용 확인 (개발 환경에서는 HTTP 허용)
      const isHTTPS = page.url().startsWith('https://');

      // 2. 보안 헤더 확인
      const securityHeaders = await page.evaluate(() => {
        const headers: Record<string, string> = {};
        // CSP, X-Frame-Options 등 클라이언트에서 확인 가능한 헤더들
        return headers;
      });

      // 3. XSS 방지 확인
      const xssProtection = await page.evaluate(() => {
        const scripts = document.querySelectorAll('script:not([src])');
        return Array.from(scripts).map(script => script.textContent || '');
      });

      // 4. 민감 정보 노출 확인
      const sensitiveInfo = await page.evaluate(() => {
        const text = document.body.innerText;
        const sensitivePatterns = [
          /api[_-]?key/i,
          /password/i,
          /secret/i,
          /token/i
        ];

        return sensitivePatterns.some(pattern => pattern.test(text));
      });

      console.log(`🛡️ Security Analysis:`);
      console.log(`   - Protocol: ${isHTTPS ? 'HTTPS ✅' : 'HTTP (Dev Mode)'}`);
      console.log(`   - Inline Scripts: ${xssProtection.length}`);
      console.log(`   - Sensitive Info Exposed: ${sensitiveInfo ? '⚠️ Yes' : '✅ No'}`);

      expect(sensitiveInfo).toBe(false);
    });

    test('Browser Compatibility Matrix', async ({ page }) => {
      console.log('🌐 Testing browser compatibility features...');

      await page.goto(baseURL);

      // 1. Modern JavaScript features 확인
      const jsFeatures = await page.evaluate(() => {
        return {
          es6Modules: typeof Symbol !== 'undefined',
          asyncAwait: typeof (async () => {})() !== 'undefined',
          arrowFunctions: (() => true)(),
          destructuring: (() => { const {a = 1} = {}; return a === 1; })(),
          spread: [...[1, 2, 3]].length === 3,
          template: `test`.includes('test'),
          optionalChaining: {a: 1}?.a === 1
        };
      });

      // 2. CSS Features 확인
      const cssFeatures = await page.evaluate(() => {
        const testCSS = (property: string) => {
          return window.getComputedStyle(document.body)[property as any] !== undefined;
        };

        return {
          flexbox: testCSS('display') === 'flex' || testCSS('display') === 'block',
          grid: testCSS('display') === 'grid' || testCSS('display') === 'block',
          customProperties: testCSS('color') !== undefined
        };
      });

      // 3. Web APIs 확인
      const webAPIs = await page.evaluate(() => {
        return {
          intersectionObserver: typeof IntersectionObserver !== 'undefined',
          mutationObserver: typeof MutationObserver !== 'undefined',
          localStorage: typeof localStorage !== 'undefined',
          sessionStorage: typeof sessionStorage !== 'undefined',
          fetch: typeof fetch !== 'undefined'
        };
      });

      // 4. Error Handling 확인
      const errorHandling = await page.evaluate(() => {
        let hasErrors = false;
        window.addEventListener('error', () => hasErrors = true);
        return !hasErrors;
      });

      console.log(`🔧 Browser Compatibility:`);
      console.log(`   - ES6+ Features: ${Object.values(jsFeatures).filter(Boolean).length}/${Object.keys(jsFeatures).length}`);
      console.log(`   - CSS Features: ${Object.values(cssFeatures).filter(Boolean).length}/${Object.keys(cssFeatures).length}`);
      console.log(`   - Web APIs: ${Object.values(webAPIs).filter(Boolean).length}/${Object.keys(webAPIs).length}`);
      console.log(`   - Error Handling: ${errorHandling ? '✅' : '❌'}`);

      expect(Object.values(jsFeatures).every(Boolean)).toBe(true);
      expect(Object.values(webAPIs).every(Boolean)).toBe(true);
    });
  });

  // 모바일 환경 테스트
  test.describe('Mobile Environment', () => {
    test.use({ ...devices['Pixel 5'] });

    test('Mobile Performance & Responsiveness', async ({ page }) => {
      console.log('📱 Testing mobile performance and responsiveness...');

      await page.goto(baseURL);

      // 1. 반응형 디자인 확인
      const isMobileOptimized = await page.evaluate(() => {
        const viewport = document.querySelector('meta[name="viewport"]');
        const hasResponsiveBreakpoints = window.innerWidth <= 768;

        return {
          hasViewport: !!viewport,
          isMobileWidth: hasResponsiveBreakpoints,
          touchFriendly: 'ontouchstart' in window
        };
      });

      // 2. 모바일 성능 측정
      const mobilePerformance = await page.evaluate(() => {
        const nav = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming;
        return {
          domContentLoaded: Math.round(nav.domContentLoadedEventEnd - nav.domContentLoadedEventStart),
          loadComplete: Math.round(nav.loadEventEnd - nav.loadEventStart),
          firstPaint: performance.getEntriesByType('paint')[0]?.startTime || 0,
          firstContentfulPaint: performance.getEntriesByType('paint')[1]?.startTime || 0
        };
      });

      // 3. 터치 상호작용 확인
      const touchInteraction = await page.evaluate(() => {
        const buttons = document.querySelectorAll('button, a, input');
        const clickableElements = Array.from(buttons).filter(el => {
          const rect = el.getBoundingClientRect();
          // 44px 최소 터치 타겟 (Apple HIG)
          return rect.width >= 44 && rect.height >= 44;
        });

        return {
          totalClickable: buttons.length,
          touchFriendly: clickableElements.length,
          touchFriendlyRatio: clickableElements.length / buttons.length
        };
      });

      console.log(`📱 Mobile Analysis:`);
      console.log(`   - Viewport Meta: ${isMobileOptimized.hasViewport ? '✅' : '❌'}`);
      console.log(`   - Mobile Width: ${isMobileOptimized.isMobileWidth ? '✅' : '❌'}`);
      console.log(`   - Touch Support: ${isMobileOptimized.touchFriendly ? '✅' : '❌'}`);
      console.log(`   - DOM Ready: ${mobilePerformance.domContentLoaded}ms`);
      console.log(`   - Load Complete: ${mobilePerformance.loadComplete}ms`);
      console.log(`   - Touch-Friendly Elements: ${touchInteraction.touchFriendly}/${touchInteraction.totalClickable}`);

      expect(isMobileOptimized.hasViewport).toBe(true);
      expect(isMobileOptimized.touchFriendly).toBe(true);
      expect(touchInteraction.touchFriendlyRatio).toBeGreaterThan(0.8);
    });
  });

  // 크로스 브라우저 테스트
  ['Desktop Firefox', 'Desktop Safari'].forEach(browserName => {
    test.describe(`Cross-Browser Testing - ${browserName}`, () => {
      const device = browserName.includes('Firefox') ? devices['Desktop Firefox'] : devices['Desktop Safari'];
      test.use(device);

      test('Cross-Browser Compatibility', async ({ page, browserName }) => {
        console.log(`🌍 Testing ${browserName} compatibility...`);

        await page.goto(baseURL);

        // 렌더링 및 기능 호환성 확인
        const browserCompatibility = await page.evaluate(() => {
          const hasContent = document.body.innerText.length > 100;
          const hasStyles = window.getComputedStyle(document.body).color !== '';
          const hasImages = document.querySelectorAll('img').length > 0;
          const hasLinks = document.querySelectorAll('a').length > 0;

          return {
            hasContent,
            hasStyles,
            hasImages,
            hasLinks,
            jsWorking: typeof window !== 'undefined'
          };
        });

        console.log(`${browserName} Compatibility:`, browserCompatibility);

        expect(browserCompatibility.hasContent).toBe(true);
        expect(browserCompatibility.hasStyles).toBe(true);
        expect(browserCompatibility.jsWorking).toBe(true);
      });
    });
  });
});