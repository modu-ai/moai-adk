/**
 * Vercel Performance Monitoring Configuration
 * Sets up Web Vitals tracking and performance analytics
 */

export function reportWebVitals(metric) {
  // Log to custom analytics endpoint
  if (typeof window !== 'undefined') {
    const body = JSON.stringify(metric);

    // Use sendBeacon for reliability
    if (navigator.sendBeacon) {
      navigator.sendBeacon('/api/metrics', body);
    }

    // Log to console in development
    if (process.env.NODE_ENV === 'development') {
      console.log('Web Vital:', metric);
    }
  }
}

/**
 * Performance Metrics to Track:
 * - LCP (Largest Contentful Paint): Measures loading performance
 * - FID (First Input Delay): Measures interactivity
 * - CLS (Cumulative Layout Shift): Measures visual stability
 * - TTFB (Time to First Byte): Measures server response time
 * - FCP (First Contentful Paint): Measures perceived load speed
 */

export const performanceThresholds = {
  LCP: {
    good: 2500,      // 2.5s
    needsImprovement: 4000,  // 4.0s
  },
  FID: {
    good: 100,       // 100ms
    needsImprovement: 300,   // 300ms
  },
  CLS: {
    good: 0.1,
    needsImprovement: 0.25,
  },
  TTFB: {
    good: 600,       // 600ms
    needsImprovement: 1800,  // 1.8s
  },
  FCP: {
    good: 1800,      // 1.8s
    needsImprovement: 3000,  // 3.0s
  },
};

/**
 * Deployment Health Check
 * Runs post-deployment validation
 */
export async function deploymentHealthCheck() {
  const checks = {
    homepage: false,
    korean: false,
    english: false,
    japanese: false,
    chinese: false,
    performance: false,
    security: false,
  };

  try {
    // Check homepage
    const homeResponse = await fetch('/', { method: 'HEAD' });
    checks.homepage = homeResponse.ok;

    // Check language versions
    const locales = ['/', '/en', '/ja', '/zh'];
    const localeResponses = await Promise.all(
      locales.map(locale =>
        fetch(`${locale}`, { method: 'HEAD' }).then(r => r.ok)
      )
    );

    checks.korean = localeResponses[0];
    checks.english = localeResponses[1];
    checks.japanese = localeResponses[2];
    checks.chinese = localeResponses[3];

    // Check performance headers
    const headerResponse = await fetch('/', { method: 'HEAD' });
    const headers = {
      'X-Content-Type-Options': headerResponse.headers.get('X-Content-Type-Options'),
      'X-Frame-Options': headerResponse.headers.get('X-Frame-Options'),
      'Strict-Transport-Security': headerResponse.headers.get('Strict-Transport-Security'),
    };
    checks.security = Object.values(headers).every(Boolean);

    return {
      status: Object.values(checks).every(Boolean) ? 'healthy' : 'degraded',
      checks,
      timestamp: new Date().toISOString(),
    };
  } catch (error) {
    return {
      status: 'unhealthy',
      error: error.message,
      timestamp: new Date().toISOString(),
    };
  }
}

/**
 * Analytics Event Tracking
 */
export function trackEvent(eventName, eventData = {}) {
  if (typeof window !== 'undefined' && window.gtag) {
    window.gtag('event', eventName, {
      ...eventData,
      timestamp: new Date().toISOString(),
    });
  }
}

/**
 * Page Performance Metrics
 */
export function getPageMetrics() {
  if (typeof window === 'undefined') return null;

  const navigation = performance.getEntriesByType('navigation')[0];
  const paintEntries = performance.getEntriesByType('paint');

  return {
    navigationTiming: {
      domainLookup: navigation.domainLookupEnd - navigation.domainLookupStart,
      connect: navigation.connectEnd - navigation.connectStart,
      request: navigation.responseStart - navigation.requestStart,
      response: navigation.responseEnd - navigation.responseStart,
      dom: navigation.domInteractive - navigation.responseEnd,
      load: navigation.loadEventEnd - navigation.loadEventStart,
    },
    paintTiming: {
      fcp: paintEntries.find(e => e.name === 'first-contentful-paint')?.startTime,
      lcp: paintEntries.find(e => e.name === 'largest-contentful-paint')?.startTime,
    },
    memory: performance.memory ? {
      usedJSHeapSize: performance.memory.usedJSHeapSize,
      totalJSHeapSize: performance.memory.totalJSHeapSize,
      jsHeapSizeLimit: performance.memory.jsHeapSizeLimit,
    } : null,
  };
}
