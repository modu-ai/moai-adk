---
title: Vercel Deployment Best Practices
description: Master Vercel deployment for Next.js with preview environments, edge functions, and CI/CD optimization
freedom_level: high
tier: ops
updated: 2025-10-31
---

# Vercel Deployment Best Practices

## Overview

Vercel is the deployment platform built by the creators of Next.js, offering zero-config deployments, automatic preview environments, global edge network, and edge functions. This skill covers deployment strategies, environment management, performance optimization, and CI/CD best practices for 2025.

## Key Patterns

### 1. Automatic Preview Deployments

**Pattern**: Every Git push gets a unique preview URL for testing.

```bash
# Connect GitHub repository
vercel link

# Every PR automatically gets preview URL:
# https://my-app-git-feature-branch-username.vercel.app

# Production deployment
git push origin main  # Auto-deploys to production domain
```

**Best Practices**:
```json
// vercel.json
{
  "git": {
    "deploymentEnabled": {
      "main": true,        // Auto-deploy main branch
      "develop": true      // Auto-deploy develop branch
    }
  },
  "github": {
    "silent": false,       // Comment on PRs with preview URL
    "autoAlias": true      // Alias preview deployments
  }
}
```

**Workflow**:
1. Create feature branch → automatic preview deployment
2. Team reviews live preview (not localhost)
3. Merge PR → automatic production deployment
4. Rollback with one click if needed

### 2. Environment Variables Management

**Pattern**: Separate environment variables by deployment context.

```bash
# Add environment variables via CLI
vercel env add DATABASE_URL production
vercel env add DATABASE_URL preview
vercel env add DATABASE_URL development

# Or via Vercel Dashboard:
# Settings → Environment Variables
```

**Limits & Best Practices**:
- **Total limit**: 64KB per deployment
- **Edge Functions/Middleware**: 5KB per variable
- **Sensitive data**: Use Vercel's encrypted storage (not .env files in repo)

```javascript
// next.config.js
module.exports = {
  env: {
    // ❌ BAD: Hardcoded secrets
    API_KEY: 'abc123'
  },
  
  // ✅ GOOD: Use environment variables
  serverRuntimeConfig: {
    apiKey: process.env.API_KEY
  },
  publicRuntimeConfig: {
    apiUrl: process.env.NEXT_PUBLIC_API_URL
  }
}
```

### 3. Edge Functions Deployment

**Pattern**: Deploy serverless functions to Vercel's edge network.

```typescript
// app/api/geolocation/route.ts
export const runtime = 'edge'

export async function GET(request: Request) {
  const geo = request.headers.get('x-vercel-ip-country') || 'Unknown'
  
  return Response.json({
    country: geo,
    city: request.headers.get('x-vercel-ip-city'),
    latency: Date.now()
  })
}
```

**Performance Comparison**:
- **Edge Runtime**: <50ms cold start, runs in 100+ locations
- **Node.js Runtime**: ~200ms cold start, runs in select regions

**Use Cases for Edge**:
- Geolocation-based routing
- A/B testing logic
- Authentication checks
- Simple data transformations
- API proxying with minimal logic

### 4. Performance Budgets & Monitoring

**Pattern**: Set performance budgets in CI pipeline.

```json
// vercel.json
{
  "builds": [
    {
      "src": "package.json",
      "use": "@vercel/next"
    }
  ],
  "headers": [
    {
      "source": "/(.*)",
      "headers": [
        {
          "key": "X-DNS-Prefetch-Control",
          "value": "on"
        },
        {
          "key": "Strict-Transport-Security",
          "value": "max-age=31536000; includeSubDomains"
        }
      ]
    }
  ],
  "trailingSlash": false
}
```

**Lighthouse CI Integration**:
```yaml
# .github/workflows/lighthouse.yml
name: Lighthouse CI
on: [pull_request]

jobs:
  lighthouse:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run Lighthouse
        uses: treosh/lighthouse-ci-action@v9
        with:
          urls: |
            https://${{ github.event.pull_request.head.sha }}-my-app.vercel.app
          budgetPath: ./budget.json
          uploadArtifacts: true
```

```json
// budget.json
{
  "budget": [
    {
      "path": "/*",
      "resourceSizes": [
        {
          "resourceType": "script",
          "budget": 300
        },
        {
          "resourceType": "total",
          "budget": 500
        }
      ],
      "timings": [
        {
          "metric": "first-contentful-paint",
          "budget": 2000
        },
        {
          "metric": "largest-contentful-paint",
          "budget": 2500
        }
      ]
    }
  ]
}
```

### 5. Incremental Static Regeneration (ISR) Strategy

**Pattern**: Configure ISR for optimal caching and freshness.

```typescript
// app/products/[id]/page.tsx
export const revalidate = 3600 // Revalidate every hour

export async function generateStaticParams() {
  const products = await fetch('https://api.example.com/products')
    .then(res => res.json())
  
  // Pre-render top 100 products
  return products.slice(0, 100).map((p: any) => ({
    id: p.id.toString()
  }))
}

export default async function ProductPage({ params }: { params: { id: string } }) {
  const product = await fetch(`https://api.example.com/products/${params.id}`, {
    next: { revalidate: 3600, tags: ['products'] }
  }).then(res => res.json())
  
  return <div>{product.name}</div>
}
```

**Vercel-specific optimizations**:
- ISR pages cached globally on edge network
- On-demand revalidation via API routes
- Automatic stale-while-revalidate behavior

### 6. Branch Protection & Required Checks

**Pattern**: Enforce quality gates before production deployment.

```yaml
# .github/workflows/ci.yml
name: CI Pipeline
on:
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
      - run: npm ci
      - run: npm test
      - run: npm run lint
      - run: npm run type-check

  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
      - run: npm ci
      - run: npm run build
```

**GitHub Branch Protection**:
- Require status checks to pass before merging
- Require review from code owners
- Automatically deploy to preview → review → merge → production

### 7. Rollback & Feature Flags

**Pattern**: Instant rollback with Vercel's deployment history.

```bash
# List recent deployments
vercel ls

# Promote specific deployment to production
vercel promote <deployment-url>

# Instant rollback (one-click in dashboard)
# Or via CLI:
vercel alias set <previous-deployment-url> production.com
```

**Feature Flags Pattern**:
```typescript
// lib/features.ts
export function isFeatureEnabled(feature: string, geo?: string) {
  const flags = {
    'new-checkout': process.env.VERCEL_ENV === 'production' ? 0.5 : 1.0,
    'dark-mode': 1.0,
    'beta-features': geo === 'US' ? 1.0 : 0.0
  }
  
  return Math.random() < (flags[feature] || 0)
}

// app/page.tsx
export default function HomePage() {
  const showNewCheckout = isFeatureEnabled('new-checkout')
  
  return (
    <div>
      {showNewCheckout ? <NewCheckout /> : <OldCheckout />}
    </div>
  )
}
```

## Checklist

- [ ] Connect GitHub repository to Vercel for automatic deployments
- [ ] Configure environment variables for production/preview/development
- [ ] Set up branch protection rules on GitHub (require checks)
- [ ] Add performance budgets (LCP < 2.5s, FID < 100ms, CLS < 0.1)
- [ ] Test Edge Functions: verify <50ms cold start times
- [ ] Enable Vercel Analytics for Core Web Vitals monitoring
- [ ] Configure ISR revalidation times based on content freshness
- [ ] Test preview deployments: share with team for feedback
- [ ] Set up Lighthouse CI for automated performance checks
- [ ] Document rollback procedure for emergency scenarios

## Resources

- **Official Vercel Docs**: https://vercel.com/docs
- **Next.js on Vercel**: https://vercel.com/docs/frameworks/nextjs
- **Environment Variables**: https://vercel.com/docs/projects/environment-variables
- **Edge Functions**: https://vercel.com/docs/functions/edge-functions
- **Deployment Guide (2025)**: https://medium.com/@takafumi.endo/how-vercel-simplifies-deployment-for-developers-beaabe0ada32
- **Performance Best Practices**: https://vercel.com/blog/nextjs-performance-best-practices

---

**Version**: 1.0.0  
**Last Updated**: 2025-10-31  
**Model Recommendation**: Sonnet (deep reasoning for deployment strategy and CI/CD optimization)
