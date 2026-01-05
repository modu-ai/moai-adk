# Vercel Platform Reference

## Vercel CLI Commands

### Project Management

```bash
# Initialize project
vercel init

# Link to existing project
vercel link

# Deploy to preview
vercel

# Deploy to production
vercel --prod

# List deployments
vercel ls

# Inspect deployment
vercel inspect <deployment-url>

# Rollback to previous deployment
vercel rollback
```

### Environment Variables

```bash
# Add environment variable
vercel env add <name> <environment>
# Environments: production, preview, development

# List environment variables
vercel env ls

# Remove environment variable
vercel env rm <name> <environment>

# Pull environment variables to local
vercel env pull .env.local
```

### Domains

```bash
# Add custom domain
vercel domains add <domain>

# List domains
vercel domains ls

# Remove domain
vercel domains rm <domain>

# Inspect domain
vercel domains inspect <domain>
```

### DNS Management

```bash
# List DNS records
vercel dns ls <domain>

# Add DNS record
vercel dns add <domain> <subdomain> <type> <value>

# Remove DNS record
vercel dns rm <record-id>
```

### Storage

```bash
# Create storage
vercel storage create <type> <name>
# Types: kv, blob, postgres

# List storage
vercel storage ls

# Remove storage
vercel storage rm <name>
```

### Secrets (Legacy)

```bash
# Add secret
vercel secrets add <name> <value>

# List secrets
vercel secrets ls

# Remove secret
vercel secrets rm <name>
```

## vercel.json Schema

### Build Configuration

```json
{
  "framework": "nextjs",
  "buildCommand": "npm run build",
  "installCommand": "npm install",
  "outputDirectory": ".next",
  "devCommand": "npm run dev"
}
```

### Regions

Available regions for functions:
- `iad1` - Washington, D.C. (default)
- `sfo1` - San Francisco
- `fra1` - Frankfurt
- `hnd1` - Tokyo
- `syd1` - Sydney
- `gru1` - Sao Paulo
- `cle1` - Cleveland
- `pdx1` - Portland
- `dub1` - Dublin
- `arn1` - Stockholm
- `sin1` - Singapore
- `bom1` - Mumbai
- `cdg1` - Paris
- `lhr1` - London

### Function Configuration

```json
{
  "functions": {
    "api/**/*.ts": {
      "memory": 1024,
      "maxDuration": 30,
      "runtime": "nodejs20.x"
    }
  }
}
```

Memory options: 128, 256, 512, 1024, 2048, 3008 MB

Duration limits by plan:
- Hobby: 10 seconds
- Pro: 60 seconds
- Enterprise: 900 seconds

### Cron Expression Syntax

```
*    *    *    *    *
|    |    |    |    |
|    |    |    |    └── Day of week (0-7, Sunday = 0 or 7)
|    |    |    └────── Month (1-12)
|    |    └─────────── Day of month (1-31)
|    └──────────────── Hour (0-23)
└───────────────────── Minute (0-59)
```

Examples:
- `0 0 * * *` - Every day at midnight
- `*/15 * * * *` - Every 15 minutes
- `0 */6 * * *` - Every 6 hours
- `0 9 * * 1-5` - 9 AM weekdays
- `0 0 1 * *` - First day of each month

### Headers Configuration

```json
{
  "headers": [
    {
      "source": "/(.*)",
      "headers": [
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

### Rewrites and Redirects

```json
{
  "rewrites": [
    { "source": "/blog/:slug", "destination": "/posts/:slug" },
    { "source": "/api/v1/:path*", "destination": "/api/:path*" },
    { "source": "/proxy/:match*", "destination": "https://external.com/:match*" }
  ],
  "redirects": [
    { "source": "/old-page", "destination": "/new-page", "permanent": true },
    { "source": "/temp-redirect", "destination": "/other", "permanent": false },
    { "source": "/external", "destination": "https://example.com", "statusCode": 301 }
  ]
}
```

## Next.js Route Segment Config

### Export Options

```typescript
// Force dynamic rendering
export const dynamic = 'force-dynamic'

// Force static rendering
export const dynamic = 'force-static'

// Automatic (default)
export const dynamic = 'auto'

// Error on dynamic usage
export const dynamic = 'error'
```

### Revalidation Options

```typescript
// Time-based revalidation (seconds)
export const revalidate = 60

// Disable revalidation (always fresh)
export const revalidate = 0

// Never revalidate (static)
export const revalidate = false
```

### Runtime Options

```typescript
// Edge runtime
export const runtime = 'edge'

// Node.js runtime (default)
export const runtime = 'nodejs'
```

### Preferred Region

```typescript
// Single region
export const preferredRegion = 'iad1'

// Multiple regions
export const preferredRegion = ['iad1', 'sfo1', 'fra1']

// All regions
export const preferredRegion = 'auto'

// Home region only
export const preferredRegion = 'home'
```

## Fetch Cache Options

```typescript
// Force cache (default for GET in Server Components)
fetch(url, { cache: 'force-cache' })

// No cache (always fresh)
fetch(url, { cache: 'no-store' })

// Time-based revalidation
fetch(url, { next: { revalidate: 3600 } })

// Tag-based revalidation
fetch(url, { next: { tags: ['posts', 'featured'] } })
```

## Response Headers

### Cache-Control Directives

```typescript
return Response.json(data, {
  headers: {
    // Public caching
    'Cache-Control': 'public, max-age=3600',

    // CDN-only caching
    'Cache-Control': 'public, s-maxage=3600, max-age=0',

    // Stale-while-revalidate
    'Cache-Control': 'public, s-maxage=60, stale-while-revalidate=300',

    // No caching
    'Cache-Control': 'private, no-cache, no-store, must-revalidate',

    // Immutable assets
    'Cache-Control': 'public, max-age=31536000, immutable'
  }
})
```

### CDN-Specific Headers

```typescript
return Response.json(data, {
  headers: {
    // Vercel-specific CDN cache
    'CDN-Cache-Control': 'public, max-age=86400',

    // Surrogate control (Varnish/Fastly compatible)
    'Surrogate-Control': 'max-age=3600',

    // Vary header for content negotiation
    'Vary': 'Accept-Encoding, Accept-Language'
  }
})
```

## Edge Middleware Matchers

```typescript
// Match all paths except specific patterns
export const config = {
  matcher: ['/((?!api|_next/static|_next/image|favicon.ico).*)']
}

// Match specific paths
export const config = {
  matcher: ['/dashboard/:path*', '/api/:path*']
}

// Match with regex
export const config = {
  matcher: [
    {
      source: '/api/:path*',
      regexp: '^/api/(.*)'
    }
  ]
}
```

## GitHub Actions Secrets

Required secrets for CI/CD integration:
- `VERCEL_TOKEN` - Vercel API token
- `VERCEL_ORG_ID` - Organization ID (optional for teams)
- `VERCEL_PROJECT_ID` - Project ID

Get these values:
1. `VERCEL_TOKEN`: vercel.com/account/tokens
2. `VERCEL_ORG_ID`: Run `vercel link` and check `.vercel/project.json`
3. `VERCEL_PROJECT_ID`: Same as above
