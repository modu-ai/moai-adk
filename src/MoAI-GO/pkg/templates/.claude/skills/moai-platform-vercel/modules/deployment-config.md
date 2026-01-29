# Deployment Configuration

## Overview

Vercel deployment configuration is managed through `vercel.json` and environment variables. This module covers project settings, function configuration, and CI/CD integration.

## vercel.json Configuration

### Complete Configuration Example

```json
{
  "$schema": "https://openapi.vercel.sh/vercel.json",
  "framework": "nextjs",
  "buildCommand": "pnpm build",
  "installCommand": "pnpm install",
  "outputDirectory": ".next",
  "regions": ["iad1", "sfo1", "fra1", "hnd1"],
  "functions": {
    "app/api/**/*.ts": {
      "memory": 1024,
      "maxDuration": 30
    },
    "app/api/heavy/**/*.ts": {
      "memory": 3008,
      "maxDuration": 60
    }
  },
  "crons": [
    {
      "path": "/api/cron/cleanup",
      "schedule": "0 0 * * *"
    },
    {
      "path": "/api/cron/sync",
      "schedule": "*/15 * * * *"
    }
  ],
  "headers": [
    {
      "source": "/api/(.*)",
      "headers": [
        { "key": "Access-Control-Allow-Origin", "value": "*" },
        { "key": "Cache-Control", "value": "s-maxage=60, stale-while-revalidate" }
      ]
    },
    {
      "source": "/(.*)",
      "headers": [
        { "key": "X-Content-Type-Options", "value": "nosniff" },
        { "key": "X-Frame-Options", "value": "DENY" },
        { "key": "X-XSS-Protection", "value": "1; mode=block" }
      ]
    }
  ],
  "rewrites": [
    { "source": "/blog/:slug", "destination": "/posts/:slug" },
    { "source": "/api/v1/:path*", "destination": "/api/:path*" }
  ],
  "redirects": [
    { "source": "/old-page", "destination": "/new-page", "permanent": true }
  ]
}
```

## Environment Variables

### CLI Management

```bash
# Production environment
vercel env add DATABASE_URL production
vercel env add API_SECRET production
vercel env add NEXT_PUBLIC_API_URL production

# Preview environments (PR deployments)
vercel env add DATABASE_URL preview
vercel env add API_SECRET preview

# Development environment
vercel env add DATABASE_URL development

# Pull environment for local development
vercel env pull .env.local

# List all environment variables
vercel env ls

# Remove environment variable
vercel env rm DATABASE_URL production
```

### Environment-Specific Configuration

```typescript
// next.config.js
module.exports = {
  env: {
    CUSTOM_KEY: process.env.CUSTOM_KEY
  },
  // Public variables (available in browser)
  publicRuntimeConfig: {
    apiUrl: process.env.NEXT_PUBLIC_API_URL
  },
  // Server-only variables
  serverRuntimeConfig: {
    secretKey: process.env.API_SECRET
  }
}
```

## Preview Deployments

### GitHub Integration

```yaml
# .github/workflows/vercel-preview.yml
name: Vercel Preview Deployment
on:
  pull_request:
    types: [opened, synchronize, reopened]

jobs:
  deploy-preview:
    runs-on: ubuntu-latest
    environment:
      name: Preview
      url: ${{ steps.deploy.outputs.url }}
    steps:
      - uses: actions/checkout@v4

      - name: Install Vercel CLI
        run: npm i -g vercel@latest

      - name: Pull Vercel Environment
        run: vercel pull --yes --environment=preview --token=${{ secrets.VERCEL_TOKEN }}

      - name: Build Project
        run: vercel build --token=${{ secrets.VERCEL_TOKEN }}

      - name: Deploy to Vercel
        id: deploy
        run: |
          url=$(vercel deploy --prebuilt --token=${{ secrets.VERCEL_TOKEN }})
          echo "url=$url" >> $GITHUB_OUTPUT

      - name: Comment PR with Preview URL
        uses: actions/github-script@v7
        with:
          script: |
            github.rest.issues.createComment({
              owner: context.repo.owner,
              repo: context.repo.repo,
              issue_number: context.issue.number,
              body: 'Preview deployed: ${{ steps.deploy.outputs.url }}'
            })
```

### Production Deployment

```yaml
# .github/workflows/vercel-production.yml
name: Vercel Production Deployment
on:
  push:
    branches: [main]

jobs:
  deploy-production:
    runs-on: ubuntu-latest
    environment:
      name: Production
      url: https://your-domain.com
    steps:
      - uses: actions/checkout@v4

      - name: Install Vercel CLI
        run: npm i -g vercel@latest

      - name: Pull Vercel Environment
        run: vercel pull --yes --environment=production --token=${{ secrets.VERCEL_TOKEN }}

      - name: Build Project
        run: vercel build --prod --token=${{ secrets.VERCEL_TOKEN }}

      - name: Deploy to Vercel
        run: vercel deploy --prebuilt --prod --token=${{ secrets.VERCEL_TOKEN }}
```

## Function Configuration

### Memory and Duration

```json
{
  "functions": {
    "app/api/light/**/*.ts": {
      "memory": 512,
      "maxDuration": 10
    },
    "app/api/standard/**/*.ts": {
      "memory": 1024,
      "maxDuration": 30
    },
    "app/api/heavy/**/*.ts": {
      "memory": 3008,
      "maxDuration": 60
    }
  }
}
```

Memory options: 128, 256, 512, 1024, 2048, 3008 MB
Duration limits: Hobby (10s), Pro (60s), Enterprise (900s)

### Cron Jobs

```json
{
  "crons": [
    {
      "path": "/api/cron/daily-cleanup",
      "schedule": "0 0 * * *"
    },
    {
      "path": "/api/cron/hourly-sync",
      "schedule": "0 * * * *"
    },
    {
      "path": "/api/cron/every-15-min",
      "schedule": "*/15 * * * *"
    }
  ]
}
```

## Monorepo Configuration

### Turborepo Setup

```json
{
  "buildCommand": "cd ../.. && pnpm turbo build --filter=web",
  "installCommand": "cd ../.. && pnpm install",
  "framework": "nextjs",
  "outputDirectory": "apps/web/.next"
}
```

### Root Configuration

```json
{
  "version": 2,
  "projects": [
    {
      "name": "web",
      "src": "apps/web"
    },
    {
      "name": "docs",
      "src": "apps/docs"
    }
  ]
}
```

## Domain Configuration

### Custom Domains

```bash
# Add custom domain
vercel domains add example.com

# Add subdomain
vercel domains add api.example.com

# List domains
vercel domains ls

# Remove domain
vercel domains rm example.com
```

### DNS Configuration

```bash
# View DNS records
vercel dns ls example.com

# Add DNS record
vercel dns add example.com @ A 76.76.21.21

# Add CNAME for subdomain
vercel dns add example.com www CNAME cname.vercel-dns.com
```

## Security Configuration

### Security Headers

```json
{
  "headers": [
    {
      "source": "/(.*)",
      "headers": [
        { "key": "Strict-Transport-Security", "value": "max-age=63072000; includeSubDomains" },
        { "key": "X-Content-Type-Options", "value": "nosniff" },
        { "key": "X-Frame-Options", "value": "DENY" },
        { "key": "X-XSS-Protection", "value": "1; mode=block" },
        { "key": "Referrer-Policy", "value": "strict-origin-when-cross-origin" },
        { "key": "Permissions-Policy", "value": "camera=(), microphone=(), geolocation=()" }
      ]
    }
  ]
}
```

### CORS Configuration

```json
{
  "headers": [
    {
      "source": "/api/(.*)",
      "headers": [
        { "key": "Access-Control-Allow-Credentials", "value": "true" },
        { "key": "Access-Control-Allow-Origin", "value": "https://example.com" },
        { "key": "Access-Control-Allow-Methods", "value": "GET,OPTIONS,PATCH,DELETE,POST,PUT" },
        { "key": "Access-Control-Allow-Headers", "value": "X-CSRF-Token, X-Requested-With, Accept, Accept-Version, Content-Length, Content-MD5, Content-Type, Date, X-Api-Version" }
      ]
    }
  ]
}
```

## Context7 Integration

For latest deployment configuration documentation, use:
- Library: `/vercel/vercel`
- Topics: vercel.json, environment-variables, deployment
- Token allocation: 5000-8000 for comprehensive coverage
