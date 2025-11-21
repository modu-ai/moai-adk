---
name: moai-baas-vercel-ext
description: Performance & Analytics - Vercel caching strategies, monitoring, and analytics
---

## Performance & Analytics

### Advanced Caching Configuration

```typescript
export class EnterpriseVercelManager {
  // Advanced caching configuration
  configureCaching(): CacheConfig {
    return {
      rules: [
        {
          source: '/api/(.*)',
          headers: {
            'Cache-Control': 's-maxage=60, stale-while-revalidate=300',
            'Vercel-CDN-Cache-Control': 'max-age=3600',
          },
        },
        {
          source: '/_next/static/(.*)',
          headers: {
            'Cache-Control': 'public, max-age=31536000, immutable',
          },
        },
        {
          source: '/images/(.*)',
          headers: {
            'Cache-Control': 'public, max-age=86400',
          },
        },
      ],
      revalidate: {
        '/api/products': 3600, // 1 hour
        '/api/users': 60, // 1 minute
        '/blog/(.*)': 86400, // 24 hours
      },
    };
  }
}

interface CacheConfig {
  rules: CacheRule[];
  revalidate: Record<string, number>;
}

interface CacheRule {
  source: string;
  headers: Record<string, string>;
}
```

### Analytics Integration

```typescript
export class VercelAnalytics {
  private collectEndpoint: string = '/api/analytics/collect';

  async trackEvent(event: AnalyticsEvent): Promise<void> {
    try {
      await fetch(this.collectEndpoint, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          ...event,
          timestamp: new Date().toISOString(),
          userAgent: navigator.userAgent,
          url: window.location.href,
        }),
      });
    } catch (error) {
      console.error('Analytics tracking error:', error);
    }
  }

  async trackPageView(page: string, title: string): Promise<void> {
    await this.trackEvent({
      name: 'page_view',
      data: {
        page,
        title,
        referrer: document.referrer,
      },
    });
  }

  async trackUserAction(action: string, data: Record<string, any>): Promise<void> {
    await this.trackEvent({
      name: 'user_action',
      data: {
        action,
        ...data,
      },
    });
  }

  async trackPerformance(metric: string, value: number): Promise<void> {
    await this.trackEvent({
      name: 'performance',
      data: {
        metric,
        value,
        connectionType: (navigator as any).connection?.effectiveType,
      },
    });
  }
}

interface AnalyticsEvent {
  name: string;
  data: Record<string, any>;
}
```

### Performance Monitoring

```typescript
export class VercelMonitoring {
  private vitals: WebVitals = {};

  recordVital(name: string, value: number): void {
    this.vitals[name] = value;
    
    // Send to analytics if value exceeds threshold
    const thresholds: Record<string, number> = {
      LCP: 2500, // Largest Contentful Paint
      FID: 100, // First Input Delay
      CLS: 0.1, // Cumulative Layout Shift
      FCP: 1800, // First Contentful Paint
      TTFB: 800, // Time to First Byte
    };

    if (value > thresholds[name]) {
      // Send performance alert
      this.sendPerformanceAlert(name, value, thresholds[name]);
    }
  }

  private async sendPerformanceAlert(
    metric: string, 
    value: number, 
    threshold: number
  ): Promise<void> {
    try {
      await fetch('/api/monitoring/performance', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          metric,
          value,
          threshold,
          url: window.location.href,
          timestamp: new Date().toISOString(),
          userAgent: navigator.userAgent,
        }),
      });
    } catch (error) {
      console.error('Performance monitoring error:', error);
    }
  }

  getVitals(): WebVitals {
    return { ...this.vitals };
  }
}

interface WebVitals {
  LCP?: number;
  FID?: number;
  CLS?: number;
  FCP?: number;
  TTFB?: number;
}
```

### Image Optimization

```typescript
// Next.js configuration for image optimization
const nextConfig = {
  images: {
    domains: ['yourdomain.com', 'cdn.yourdomain.com'],
    formats: ['image/webp', 'image/avif'],
    deviceSizes: [640, 750, 828, 1080, 1200, 1920, 2048, 3840],
    imageSizes: [16, 32, 48, 64, 96, 128, 256, 384],
  },
  
  // Compiler optimization
  compiler: {
    removeConsole: process.env.NODE_ENV === 'production',
  },
};
```

### Headers Configuration

```typescript
// Headers for performance and security
async headers() {
  return [
    {
      source: '/api/:path*',
      headers: [
        {
          key: 'Cache-Control',
          value: 's-maxage=60, stale-while-revalidate=300',
        },
        {
          key: 'X-Frame-Options',
          value: 'DENY',
        },
        {
          key: 'X-Content-Type-Options',
          value: 'nosniff',
        },
      ],
    },
    {
      source: '/(.*)',
      headers: [
        {
          key: 'X-DNS-Prefetch-Control',
          value: 'on',
        },
      ],
    },
  ];
}
```

---

**Module**: Performance & Analytics
**Parent Skill**: moai-baas-vercel-ext
**Last Updated**: 2025-11-21
