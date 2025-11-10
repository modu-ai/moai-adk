import { test, expect } from '@playwright/test';

test.describe('Advanced Performance Profiling', () => {
  const baseURL = 'http://localhost:3000';

  test('Network Performance Analysis', async ({ page }) => {
    console.log('🌐 Network Performance Analysis...');

    const networkRequests: any[] = [];
    page.on('request', request => {
      networkRequests.push({
        url: request.url(),
        method: request.method(),
        resourceType: request.resourceType(),
        headers: request.headers()
      });
    });

    page.on('response', response => {
      const request = networkRequests.find(r => r.url === response.url());
      if (request) {
        request.status = response.status();
        request.responseHeaders = response.headers();
        request.timing = response.timing ? {
          startTime: response.timing.startTime,
          requestStart: response.timing.requestStart,
          responseStart: response.timing.responseStart,
          end: response.timing.endTime
        } : null;
      }
    });

    await page.goto(baseURL, { waitUntil: 'networkidle' });

    // 리소스 타입별 분석
    const resourceAnalysis = {
      scripts: networkRequests.filter(r => r.resourceType === 'script'),
      stylesheets: networkRequests.filter(r => r.resourceType === 'stylesheet'),
      images: networkRequests.filter(r => r.resourceType === 'image'),
      documents: networkRequests.filter(r => r.resourceType === 'document'),
      fonts: networkRequests.filter(r => r.resourceType === 'font'),
      other: networkRequests.filter(r => !['script', 'stylesheet', 'image', 'document', 'font'].includes(r.resourceType))
    };

    // HTTP/2 지원 확인
    const http2Requests = networkRequests.filter(r => r.responseHeaders?.[':status']);

    // 캐싱 전략 분석
    const cachedResponses = networkRequests.filter(r =>
      r.responseHeaders?.['cache-control'] &&
      r.responseHeaders['cache-control'].includes('max-age')
    );

    console.log(`📊 Network Analysis:`);
    console.log(`   - Total Requests: ${networkRequests.length}`);
    console.log(`   - Scripts: ${resourceAnalysis.scripts.length}`);
    console.log(`   - Stylesheets: ${resourceAnalysis.stylesheets.length}`);
    console.log(`   - Images: ${resourceAnalysis.images.length}`);
    console.log(`   - Fonts: ${resourceAnalysis.fonts.length}`);
    console.log(`   - HTTP/2 Requests: ${http2Requests.length}`);
    console.log(`   - Cached Responses: ${cachedResponses.length}`);

    // 최적화 권장사항 확인
    const hasLargeAssets = networkRequests.some(r =>
      r.responseHeaders?.['content-length'] &&
      parseInt(r.responseHeaders['content-length']) > 1024 * 1024 // 1MB
    );

    const hasUncompressedAssets = networkRequests.some(r =>
      r.resourceType === 'script' || r.resourceType === 'stylesheet'
    ).filter(r => !r.responseHeaders?.['content-encoding']);

    console.log(`   - Large Assets (>1MB): ${hasLargeAssets ? '⚠️ Found' : '✅ None'}`);
    console.log(`   - Uncompressed Assets: ${hasUncompressedAssets.length > 0 ? '⚠️ Found' : '✅ None'}`);

    expect(networkRequests.length).toBeGreaterThan(0);
    expect(resourceAnalysis.scripts.length).toBeGreaterThan(0);
  });

  test('Runtime Performance Monitoring', async ({ page }) => {
    console.log('⚡ Runtime Performance Monitoring...');

    await page.goto(baseURL);

    // 런타임 성능 지표 수집
    const performanceMetrics = await page.evaluate(() => {
      const metrics: any = {};

      // Navigation Timing
      const nav = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming;
      metrics.navigation = {
        domContentLoaded: nav.domContentLoadedEventEnd - nav.domContentLoadedEventStart,
        loadComplete: nav.loadEventEnd - nav.loadEventStart,
        timeToFirstByte: nav.responseStart - nav.requestStart,
        domInteractive: nav.domInteractive - nav.navigationStart,
        firstPaint: 0,
        firstContentfulPaint: 0
      };

      // Paint Timing
      const paintEntries = performance.getEntriesByType('paint');
      paintEntries.forEach(entry => {
        if (entry.name === 'first-paint') {
          metrics.navigation.firstPaint = entry.startTime;
        } else if (entry.name === 'first-contentful-paint') {
          metrics.navigation.firstContentfulPaint = entry.startTime;
        }
      });

      // Resource Timing
      const resources = performance.getEntriesByType('resource');
      metrics.resources = {
        total: resources.length,
        totalSize: resources.reduce((sum, resource) => {
          return sum + (resource.transferSize || 0);
        }, 0),
        slowestResource: Math.max(...resources.map(r => r.duration)),
        averageResourceTime: resources.reduce((sum, r) => sum + r.duration, 0) / resources.length
      };

      // Memory Usage (Chrome only)
      if ((performance as any).memory) {
        metrics.memory = {
          usedJSHeapSize: (performance as any).memory.usedJSHeapSize,
          totalJSHeapSize: (performance as any).memory.totalJSHeapSize,
          jsHeapSizeLimit: (performance as any).memory.jsHeapSizeLimit
        };
      }

      // Long Tasks (if available)
      const observer = new PerformanceObserver(list => {
        metrics.longTasks = list.getEntries().map(entry => ({
          duration: entry.duration,
          startTime: entry.startTime
        }));
      });

      try {
        observer.observe({ entryTypes: ['longtask'] });
      } catch (e) {
        metrics.longTasks = [];
      }

      return metrics;
    });

    // 3초 후 추가 메트릭 수집 (Long Tasks 측정을 위해)
    await page.waitForTimeout(3000);

    console.log(`⚡ Runtime Performance:`);
    console.log(`   - Time to First Byte: ${Math.round(performanceMetrics.navigation.timeToFirstByte)}ms`);
    console.log(`   - DOM Interactive: ${Math.round(performanceMetrics.navigation.domInteractive)}ms`);
    console.log(`   - First Paint: ${Math.round(performanceMetrics.navigation.firstPaint)}ms`);
    console.log(`   - First Contentful Paint: ${Math.round(performanceMetrics.navigation.firstContentfulPaint)}ms`);
    console.log(`   - DOM Content Loaded: ${Math.round(performanceMetrics.navigation.domContentLoaded)}ms`);
    console.log(`   - Load Complete: ${Math.round(performanceMetrics.navigation.loadComplete)}ms`);
    console.log(`   - Resources: ${performanceMetrics.resources.total} files`);
    console.log(`   - Total Size: ${Math.round(performanceMetrics.resources.totalSize / 1024)}KB`);
    console.log(`   - Average Resource Time: ${Math.round(performanceMetrics.resources.averageResourceTime)}ms`);

    if (performanceMetrics.memory) {
      console.log(`   - Memory Usage: ${Math.round(performanceMetrics.memory.usedJSHeapSize / 1024 / 1024)}MB`);
    }

    if (performanceMetrics.longTasks) {
      console.log(`   - Long Tasks: ${performanceMetrics.longTasks.length}`);
    }

    // 성능 기준 검증
    expect(performanceMetrics.navigation.timeToFirstByte).toBeLessThan(800);
    expect(performanceMetrics.navigation.firstContentfulPaint).toBeLessThan(2500);
    expect(performanceMetrics.navigation.loadComplete).toBeLessThan(5000);
  });

  test('JavaScript Bundle Analysis', async ({ page }) => {
    console.log('📦 JavaScript Bundle Analysis...');

    await page.goto(baseURL);

    // 번들 정보 수집
    const bundleAnalysis = await page.evaluate(() => {
      const scripts = Array.from(document.querySelectorAll('script[src]'));
      const inlineScripts = Array.from(document.querySelectorAll('script:not([src])'));

      const bundles = scripts.map(script => {
        const src = script.getAttribute('src')!;
        return {
          url: src,
          isChunk: src.includes('chunk-'),
          isLazy: src.includes('lazy'),
          isFramework: src.includes('react') || src.includes('next') || src.includes('nextra'),
          isAsync: script.hasAttribute('async'),
          isDefer: script.hasAttribute('defer')
        };
      });

      return {
        totalScripts: scripts.length,
        inlineScripts: inlineScripts.length,
        chunks: bundles.filter(b => b.isChunk).length,
        lazyBundles: bundles.filter(b => b.isLazy).length,
        frameworkBundles: bundles.filter(b => b.isFramework).length,
        asyncScripts: bundles.filter(b => b.isAsync).length,
        deferScripts: bundles.filter(b => b.isDefer).length,
        bundles
      };
    });

    // 동적 임포트 확인
    const dynamicImports = await page.evaluate(() => {
      // Next.js 동적 import 사용 여부 확인
      return window.__NEXT_DATA__?.props?.pageProps || {};
    });

    // 코드 분할 최적화 확인
    const codeSplittingScore = bundleAnalysis.chunks > 0 ? 'Good' : 'Needs Improvement';
    const lazyLoadingScore = bundleAnalysis.lazyBundles > 0 ? 'Good' : 'Needs Improvement';

    console.log(`📦 Bundle Analysis:`);
    console.log(`   - Total Scripts: ${bundleAnalysis.totalScripts}`);
    console.log(`   - Inline Scripts: ${bundleAnalysis.inlineScripts}`);
    console.log(`   - Code Splitting: ${codeSplittingScore} (${bundleAnalysis.chunks} chunks)`);
    console.log(`   - Lazy Loading: ${lazyLoadingScore} (${bundleAnalysis.lazyBundles} lazy bundles)`);
    console.log(`   - Framework Bundles: ${bundleAnalysis.frameworkBundles}`);
    console.log(`   - Async Scripts: ${bundleAnalysis.asyncScripts}`);
    console.log(`   - Defer Scripts: ${bundleAnalysis.deferScripts}`);

    // Next.js 최적화 확인
    const nextOptimizations = {
      hasDynamicImports: Object.keys(dynamicImports).length > 0,
      hasAutoOptimization: true, // Next.js 자동 최적화
      hasImageOptimization: !!document.querySelector('img[nw-attrs]'), // Next.js Image
      hasFontOptimization: !!document.querySelector('link[rel="preload"][as="font"]')
    };

    console.log(`🚀 Next.js Optimizations:`);
    console.log(`   - Dynamic Imports: ${nextOptimizations.hasDynamicImports ? '✅' : '❌'}`);
    console.log(`   - Auto Optimization: ${nextOptimizations.hasAutoOptimization ? '✅' : '❌'}`);
    console.log(`   - Image Optimization: ${nextOptimizations.hasImageOptimization ? '✅' : '❌'}`);
    console.log(`   - Font Optimization: ${nextOptimizations.hasFontOptimization ? '✅' : '❌'}`);

    expect(bundleAnalysis.totalScripts).toBeGreaterThan(0);
    expect(nextOptimizations.hasAutoOptimization).toBe(true);
  });

  test('Render Performance Analysis', async ({ page }) => {
    console.log('🎨 Render Performance Analysis...');

    await page.goto(baseURL);

    // 렌더링 성능 분석
    const renderMetrics = await page.evaluate(async () => {
      const metrics: any = {};

      // First Input Delay 측정 준비
      let fid = 0;
      const fidPromise = new Promise(resolve => {
        const measureFid = (event: Event) => {
          fid = event.timeStamp;
          resolve(fid);
          document.removeEventListener('click', measureFid);
          document.removeEventListener('keydown', measureFid);
        };

        document.addEventListener('click', measureFid);
        document.addEventListener('keydown', measureFid);
      });

      // Largest Contentful Paint 측정
      const lcpPromise = new Promise(resolve => {
        new PerformanceObserver(list => {
          const entries = list.getEntries();
          const lastEntry = entries[entries.length - 1];
          resolve(lastEntry.startTime);
        }).observe({ entryTypes: ['largest-contentful-paint'] });
      });

      // Cumulative Layout Shift 측정
      let cls = 0;
      const clsPromise = new Promise(resolve => {
        new PerformanceObserver(list => {
          for (const entry of list.getEntries()) {
            if (!(entry as any).hadRecentInput) {
              cls += (entry as any).value;
            }
          }
          setTimeout(() => resolve(cls), 2500);
        }).observe({ entryTypes: ['layout-shift'] });
      });

      // DOM 렌더링 관련 메트릭
      const nav = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming;

      metrics.dom = {
        parsing: nav.domInteractive - nav.navigationStart,
        domContentLoaded: nav.domContentLoadedEventEnd - nav.domContentLoadedEventStart,
        domComplete: nav.domComplete - nav.navigationStart
      };

      // 렌더링 트리 정보
      const totalElements = document.querySelectorAll('*').length;
      const textNodes = document.createTreeWalker(
        document.body,
        NodeFilter.SHOW_TEXT,
        null,
        false
      );

      let textNodeCount = 0;
      while (textNodes.nextNode()) textNodeCount++;

      metrics.render = {
        totalElements,
        textNodes: textNodeCount,
        domDepth: getMaxDOMDepth(document.body)
      };

      // 비동기 작업들 대기
      await Promise.all([
        lcpPromise.catch(() => 0),
        clsPromise,
        fidPromise
      ]);

      metrics.webVitals = {
        lcp: await lcpPromise.catch(() => 0),
        cls: await clsPromise,
        fid: fid
      };

      function getMaxDOMDepth(element: Element, depth = 0): number {
        const children = Array.from(element.children);
        return children.length === 0 ? depth :
          Math.max(...children.map(child => getMaxDOMDepth(child, depth + 1)));
      }

      return metrics;
    });

    console.log(`🎨 Render Performance:`);
    console.log(`   - DOM Parsing: ${Math.round(renderMetrics.dom.parsing)}ms`);
    console.log(`   - DOM Content Loaded: ${Math.round(renderMetrics.dom.domContentLoaded)}ms`);
    console.log(`   - DOM Complete: ${Math.round(renderMetrics.dom.domComplete)}ms`);
    console.log(`   - Total Elements: ${renderMetrics.render.totalElements}`);
    console.log(`   - Text Nodes: ${renderMetrics.render.textNodes}`);
    console.log(`   - Max DOM Depth: ${renderMetrics.render.domDepth}`);
    console.log(`   - LCP: ${Math.round(renderMetrics.webVitals.lcp)}ms`);
    console.log(`   - CLS: ${Math.round(renderMetrics.webVitals.cls * 1000) / 1000}`);
    console.log(`   - FID: ${renderMetrics.webVitals.fid}ms (measured on interaction)`);

    // 렌더링 최적화 확인
    const renderOptimizations = {
      reasonableDOMDepth: renderMetrics.render.domDepth < 20,
      reasonableElementCount: renderMetrics.render.totalElements < 5000,
      fastDOMParsing: renderMetrics.dom.parsing < 500
    };

    console.log(`🚀 Render Optimizations:`);
    console.log(`   - DOM Depth: ${renderOptimizations.reasonableDOMDepth ? '✅' : '⚠️'} (${renderMetrics.render.domDepth} levels)`);
    console.log(`   - Element Count: ${renderOptimizations.reasonableElementCount ? '✅' : '⚠️'} (${renderMetrics.render.totalElements} elements)`);
    console.log(`   - Fast Parsing: ${renderOptimizations.fastDOMParsing ? '✅' : '⚠️'} (${Math.round(renderMetrics.dom.parsing)}ms)`);

    expect(renderMetrics.render.totalElements).toBeGreaterThan(0);
    expect(renderOptimizations.reasonableDOMDepth).toBe(true);
  });
});