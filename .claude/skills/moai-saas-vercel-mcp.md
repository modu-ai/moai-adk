# moai-saas-vercel-mcp

Master Vercel for Next.js deployment with preview environments, edge functions, and CI/CD optimization.

## Quick Start

Vercel is the official Next.js hosting platform designed for performance and developer experience. Use this skill when deploying Next.js applications, configuring edge functions, managing preview environments, or optimizing performance monitoring.

## Core Patterns

### Pattern 1: Project Deployment Setup

**Pattern**: Configure a Vercel project with environment variables, build settings, and deployment hooks.

```typescript
// vercel.json - Project configuration
{
  "buildCommand": "npm run build",
  "outputDirectory": ".next",
  "env": {
    "DATABASE_URL": "@db_url",
    "API_SECRET": "@api_secret"
  },
  "regions": ["sfo1", "iad1"],
  "functions": {
    "api/**": {
      "memory": 3008,
      "maxDuration": 60
    }
  }
}
```

**When to use**:
- Setting up a new Next.js project on Vercel
- Configuring environment-specific variables for dev/staging/production
- Optimizing function memory and timeout settings
- Specifying deployment regions for global performance

**Key benefits**:
- Declarative infrastructure configuration
- Automatic redeploy on git push
- Environment variable management without hardcoding
- Region-aware deployment for latency optimization

### Pattern 2: Edge Functions & Middleware

**Pattern**: Deploy serverless functions at edge locations for low-latency request processing.

```typescript
// pages/api/redirect.ts - Vercel Edge Function
import { NextRequest, NextResponse } from 'next/server';

export const config = {
  runtime: 'edge',
  regions: ['sfo1', 'iad1'], // Deploy to specific regions
};

export default async function handler(request: NextRequest) {
  const { searchParams } = new URL(request.url);
  const redirect = searchParams.get('to');

  // Validate and redirect at edge
  if (redirect && isValidUrl(redirect)) {
    return NextResponse.redirect(redirect, { status: 301 });
  }

  return NextResponse.json(
    { error: 'Invalid redirect target' },
    { status: 400 }
  );
}

function isValidUrl(url: string): boolean {
  try {
    new URL(url);
    return true;
  } catch {
    return false;
  }
}
```

**When to use**:
- Authentication checks before rendering pages
- Request/response transformations (redirects, headers)
- Geographic routing based on visitor location
- Rate limiting or bot protection at edge

**Key benefits**:
- Sub-10ms latency (runs at edge, not origin server)
- No cold starts (optimized runtime)
- Simple deployment (same as regular functions)

### Pattern 3: Preview Deployments & Environment Promotion

**Pattern**: Manage multiple deployment environments with automatic PR previews.

```bash
# GitHub Actions workflow for Vercel deployments
name: Vercel Preview Deployment
on:
  pull_request:
    branches: [main]

jobs:
  Deploy-Preview:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - uses: vercel/action@main
        with:
          vercel-token: ${{ secrets.VERCEL_TOKEN }}
          vercel-org-id: ${{ secrets.VERCEL_ORG_ID }}
          vercel-project-id: ${{ secrets.VERCEL_PROJECT_ID }}
          scope: ${{ secrets.VERCEL_ORG_ID }}

  Deploy-Production:
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - uses: vercel/action@main
        with:
          vercel-token: ${{ secrets.VERCEL_TOKEN }}
          vercel-org-id: ${{ secrets.VERCEL_ORG_ID }}
          vercel-project-id: ${{ secrets.VERCEL_PROJECT_ID }}
          vercel-args: --prod
```

**When to use**:
- Setting up automatic deployments for pull requests
- Managing staging and production environments
- Controlling when changes go live to users
- Rolling back to previous deployments

**Key benefits**:
- Automatic PR preview links
- Zero-downtime deployments
- Instant rollback capability
- Environment-specific variable configuration

## Progressive Disclosure

### Level 1: Basic Deployment
- Create Vercel account and connect GitHub
- Deploy Next.js project with one click
- View deployment logs and errors
- Access live URL immediately

### Level 2: Advanced Configuration
- Configure custom domains
- Set environment variables per environment
- Manage team members and permissions
- Monitor Web Vitals and performance

### Level 3: Expert Optimization
- Use edge functions for low-latency processing
- Implement incremental static regeneration (ISR)
- Configure image optimization settings
- Optimize bundle size and Core Web Vitals

## Works Well With

- **Next.js 16+**: Official hosting platform (built by same team)
- **React 19**: Server Components and use() hook support
- **GitHub**: Automatic deployments on git push
- **Postgres/Supabase**: Database integration via environment variables
- **Middleware**: Request/response interception at edge
- **ISR & Streaming**: Optimal performance with Vercel's infrastructure

## References

- **Official Documentation**: https://vercel.com/docs
- **Next.js Deployment**: https://nextjs.org/docs/deployment
- **Edge Functions**: https://vercel.com/docs/functions/edge-functions
- **Web Vitals**: https://vercel.com/docs/speed-insights
- **Environment Variables**: https://vercel.com/docs/concepts/projects/environment-variables
