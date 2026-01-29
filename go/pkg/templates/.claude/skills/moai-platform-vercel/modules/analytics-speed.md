# Analytics and Speed Insights

## Overview

Vercel provides built-in analytics and performance monitoring through Vercel Analytics (page views and visitors) and Speed Insights (Core Web Vitals). Both integrate seamlessly with Next.js applications.

## Vercel Analytics

### Installation

```bash
npm install @vercel/analytics
```

### Basic Integration

```typescript
// app/layout.tsx
import { Analytics } from '@vercel/analytics/react'

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>
        {children}
        <Analytics />
      </body>
    </html>
  )
}
```

### Configuration Options

```typescript
// app/layout.tsx
import { Analytics } from '@vercel/analytics/react'

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>
        {children}
        <Analytics
          mode="production" // or "development"
          debug={process.env.NODE_ENV === 'development'}
          beforeSend={(event) => {
            // Filter out sensitive paths
            if (event.url.includes('/admin')) {
              return null
            }
            return event
          }}
        />
      </body>
    </html>
  )
}
```

### Custom Events

```typescript
// components/Button.tsx
import { track } from '@vercel/analytics'

export function BuyButton({ productId }: { productId: string }) {
  return (
    <button
      onClick={() => {
        track('purchase_clicked', {
          productId,
          timestamp: new Date().toISOString()
        })
      }}
    >
      Buy Now
    </button>
  )
}
```

## Speed Insights

### Installation

```bash
npm install @vercel/speed-insights
```

### Basic Integration

```typescript
// app/layout.tsx
import { Analytics } from '@vercel/analytics/react'
import { SpeedInsights } from '@vercel/speed-insights/next'

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>
        {children}
        <Analytics />
        <SpeedInsights />
      </body>
    </html>
  )
}
```

### Configuration Options

```typescript
// app/layout.tsx
import { SpeedInsights } from '@vercel/speed-insights/next'

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>
        {children}
        <SpeedInsights
          sampleRate={0.5} // Sample 50% of page views
          debug={process.env.NODE_ENV === 'development'}
          route={null} // Override detected route
        />
      </body>
    </html>
  )
}
```

## Custom Web Vitals Reporting

### Client-Side Implementation

```typescript
// app/components/WebVitals.tsx
'use client'

import { useReportWebVitals } from 'next/web-vitals'

export function WebVitals() {
  useReportWebVitals((metric) => {
    const body = JSON.stringify({
      name: metric.name,
      value: metric.value,
      rating: metric.rating,
      delta: metric.delta,
      id: metric.id,
      navigationType: metric.navigationType
    })

    // Send to custom analytics endpoint
    if (navigator.sendBeacon) {
      navigator.sendBeacon('/api/vitals', body)
    } else {
      fetch('/api/vitals', { body, method: 'POST', keepalive: true })
    }
  })

  return null
}
```

### API Endpoint for Custom Analytics

```typescript
// app/api/vitals/route.ts
import { NextRequest } from 'next/server'

export async function POST(request: NextRequest) {
  const metric = await request.json()

  // Log to console in development
  if (process.env.NODE_ENV === 'development') {
    console.log('Web Vital:', metric)
  }

  // Send to external analytics service
  await fetch('https://analytics.example.com/vitals', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      ...metric,
      page: request.headers.get('referer'),
      userAgent: request.headers.get('user-agent'),
      timestamp: new Date().toISOString()
    })
  })

  return Response.json({ received: true })
}
```

## Core Web Vitals

### Metrics Tracked

LCP (Largest Contentful Paint):
- Measures loading performance
- Good: under 2.5 seconds
- Needs improvement: 2.5-4 seconds
- Poor: over 4 seconds

FID (First Input Delay):
- Measures interactivity
- Good: under 100ms
- Needs improvement: 100-300ms
- Poor: over 300ms

CLS (Cumulative Layout Shift):
- Measures visual stability
- Good: under 0.1
- Needs improvement: 0.1-0.25
- Poor: over 0.25

INP (Interaction to Next Paint):
- Measures responsiveness
- Good: under 200ms
- Needs improvement: 200-500ms
- Poor: over 500ms

TTFB (Time to First Byte):
- Measures server response time
- Good: under 800ms
- Needs improvement: 800-1800ms
- Poor: over 1800ms

### Optimization Strategies

```typescript
// Optimize LCP with priority images
import Image from 'next/image'

export function Hero() {
  return (
    <Image
      src="/hero.jpg"
      alt="Hero"
      width={1200}
      height={600}
      priority // Load immediately
      fetchPriority="high"
    />
  )
}
```

```typescript
// Optimize CLS with size hints
export function ProductCard({ product }) {
  return (
    <div className="aspect-square"> {/* Reserve space */}
      <Image
        src={product.image}
        alt={product.name}
        width={300}
        height={300}
        className="object-cover"
      />
    </div>
  )
}
```

## Real User Monitoring

### Aggregate Metrics Dashboard

Access via Vercel Dashboard:
1. Navigate to your project
2. Click on "Analytics" tab
3. View page views, unique visitors, and top pages
4. Click on "Speed Insights" for Web Vitals

### Performance Budgets

```typescript
// next.config.js
module.exports = {
  experimental: {
    webVitalsAttribution: ['CLS', 'LCP', 'FID', 'INP', 'TTFB']
  }
}
```

## Third-Party Analytics Integration

### Google Analytics 4

```typescript
// app/layout.tsx
import Script from 'next/script'

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>
        {children}
        <Script
          src="https://www.googletagmanager.com/gtag/js?id=G-XXXXXXXXXX"
          strategy="afterInteractive"
        />
        <Script id="google-analytics" strategy="afterInteractive">
          {`
            window.dataLayer = window.dataLayer || [];
            function gtag(){dataLayer.push(arguments);}
            gtag('js', new Date());
            gtag('config', 'G-XXXXXXXXXX');
          `}
        </Script>
      </body>
    </html>
  )
}
```

### Combining with Vercel Analytics

```typescript
// app/layout.tsx
import { Analytics } from '@vercel/analytics/react'
import { SpeedInsights } from '@vercel/speed-insights/next'
import Script from 'next/script'

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>
        {children}
        {/* Vercel Analytics */}
        <Analytics />
        <SpeedInsights />
        {/* Google Analytics */}
        <Script src="https://www.googletagmanager.com/gtag/js?id=G-XXXXXXXXXX" />
      </body>
    </html>
  )
}
```

## Context7 Integration

For latest analytics and Speed Insights documentation, use:
- Library: `/vercel/vercel`
- Topics: analytics, speed-insights, web-vitals
- Token allocation: 5000-8000 for comprehensive coverage
