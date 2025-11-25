---
name: moai-baas-vercel-ext
description: Configuration & Deployment - Vercel deployment strategies and advanced features
---

## Configuration & Deployment

### Next.js Configuration

```typescript
const nextConfig = {
  // Enable experimental features
  experimental: {
    optimizeCss: true,
    optimizePackageImports: ['lucide-react', '@radix-ui/react-icons'],
    turbo: {
      rules: {
        '*.svg': {
          loaders: ['@svgr/webpack'],
          as: '*.js',
        },
      },
    },
  },

  // Webpack configuration
  webpack: (config, { dev, isServer }) => {
    // Custom webpack configuration
    if (!dev && !isServer) {
      Object.assign(config.resolve.alias, {
        'react': 'preact/compat',
        'react-dom': 'preact/compat',
      });
    }

    return config;
  },

  // Redirects and rewrites
  async redirects() {
    return [
      {
        source: '/home',
        destination: '/',
        permanent: true,
      },
      {
        source: '/docs/:path*',
        destination: 'https://docs.yourdomain.com/:path*',
        permanent: true,
      },
    ];
  },

  async rewrites() {
    return [
      {
        source: '/api/analytics/:path*',
        destination: '/api/analytics/:path*',
      },
    ];
  },
};
```

### A/B Testing with Middleware

```typescript
import { NextRequest, NextResponse } from 'next/server';
import crypto from 'crypto';

// Middleware for A/B testing
export function middleware(request: NextRequest) {
  const userId = request.headers.get('user-id') ||
                 request.cookies.get('user-id')?.value ||
                 crypto.randomUUID();

  // Determine A/B variant using consistent hashing
  const variant = determineABVariant(userId);

  // Create response with variant header
  const response = NextResponse.next();
  response.headers.set('X-AB-Variant', variant);

  // Set cookie for client-side tracking
  response.cookies.set('ab-variant', variant, {
    maxAge: 30 * 24 * 60 * 60, // 30 days
    secure: true,
    sameSite: 'strict',
  });

  return response;
}

function determineABVariant(userId: string): string {
  // Use consistent hashing to assign variants
  const hash = crypto
    .createHash('md5')
    .update(userId)
    .digest('hex');

  const hashValue = parseInt(hash.substring(0, 8), 16);

  // Split traffic: 50% control, 50% variant_a
  if (hashValue % 100 < 50) {
    return 'control';
  } else {
    return 'variant_a';
  }
}

export const config = {
  matcher: ['/experiments/:path*'],
};
```

### Server-Side A/B Testing

```typescript
// app/api/ab-test/route.ts
import { NextRequest, NextResponse } from 'next/server';

export async function GET(request: NextRequest) {
  const variant = request.headers.get('x-ab-variant') || 'control';
  const userId = request.cookies.get('user-id')?.value;

  try {
    // Fetch experiment configuration from database
    const experimentConfig = await getExperimentConfig(variant);

    // Record analytics
    await recordExperimentParticipation({
      userId,
      variant,
      timestamp: new Date(),
      userAgent: request.headers.get('user-agent'),
    });

    return NextResponse.json({
      variant,
      config: experimentConfig,
    }, {
      headers: {
        'Cache-Control': 'no-cache',
        'X-AB-Variant': variant,
      },
    });
  } catch (error) {
    console.error('A/B test error:', error);
    return NextResponse.json(
      { error: 'Internal Server Error' },
      { status: 500 }
    );
  }
}

async function getExperimentConfig(variant: string): Promise<Record<string, any>> {
  // Fetch from database, cache, or configuration
  return {
    featureFlags: {
      newUI: variant === 'variant_a',
      betaFeatures: variant === 'variant_a',
    },
    theme: variant === 'variant_a' ? 'modern' : 'classic',
  };
}

async function recordExperimentParticipation(data: any): Promise<void> {
  // Record to analytics service, database, etc.
  console.log('Experiment participation:', data);
}
```

### Geo-Personalization

```typescript
// app/api/geo-personalization/route.ts
import { NextRequest, NextResponse } from 'next/server';

export async function GET(request: NextRequest) {
  try {
    // Get geolocation from Vercel headers
    const country = request.headers.get('x-vercel-ip-country') || 'US';
    const region = request.headers.get('x-vercel-ip-region') || '';
    const city = request.headers.get('x-vercel-ip-city') || '';

    // Get personalized content based on location
    const content = getGeoPersonalizedContent(country, region, city);

    const responseData = {
      location: {
        country,
        region,
        city,
      },
      personalized_content: content,
      timestamp: new Date().toISOString(),
    };

    return NextResponse.json(responseData, {
      headers: {
        'Cache-Control': 'public, max-age=3600',
        'X-Geo-Country': country,
      },
    });
  } catch (error) {
    console.error('Geo-personalization error:', error);
    return NextResponse.json(
      { error: 'Internal Server Error' },
      { status: 500 }
    );
  }
}

function getGeoPersonalizedContent(
  country: string,
  region: string,
  city: string
): Record<string, any> {
  // Content personalization based on geolocation
  const geoConfig: Record<string, Record<string, any>> = {
    'US': {
      currency: 'USD',
      language: 'en',
      promotions: ['free_shipping', 'local_deals'],
      shipping_options: ['standard', 'express', 'overnight'],
      tax_rate: 0.08,
    },
    'GB': {
      currency: 'GBP',
      language: 'en',
      promotions: ['free_shipping_uk', 'eu_deals'],
      shipping_options: ['standard_uk', 'express_uk'],
      vat_rate: 0.20,
    },
    'DE': {
      currency: 'EUR',
      language: 'de',
      promotions: ['free_shipping_de', 'eu_deals'],
      shipping_options: ['standard_eu', 'express_eu'],
      vat_rate: 0.19,
    },
    'CA': {
      currency: 'CAD',
      language: 'en',
      promotions: ['canada_free_shipping'],
      shipping_options: ['standard_ca', 'express_ca'],
      tax_rate: 0.05,
    },
  };

  // Return country-specific config or default
  return geoConfig[country] || {
    currency: 'USD',
    language: 'en',
    promotions: ['international_shipping'],
    shipping_options: ['standard_international'],
  };
}
```

### Edge-Based Geo-Personalization

```typescript
// lib/geo-personalization.ts
import { NextRequest, NextResponse } from 'next/server';

// Use in middleware for ultra-fast geo-routing
export function geoPersonalizationMiddleware(request: NextRequest) {
  const country = request.geo?.country || 'US';
  const city = request.geo?.city || 'Unknown';

  // Clone and modify response with geo headers
  const response = NextResponse.next();
  response.headers.set('X-User-Country', country);
  response.headers.set('X-User-City', city);

  // Set preference cookie based on geo
  const preference = getGeoPreference(country);
  response.cookies.set('geo-preference', preference, {
    maxAge: 7 * 24 * 60 * 60, // 7 days
  });

  return response;
}

function getGeoPreference(country: string): string {
  const preferences: Record<string, string> = {
    'US': 'en-USD',
    'GB': 'en-GBP',
    'DE': 'de-EUR',
    'FR': 'fr-EUR',
    'JP': 'ja-JPY',
  };

  return preferences[country] || 'en-USD';
}
```

### Function Configuration

```typescript
interface FunctionConfig {
  runtime: 'edge' | 'nodejs18.x';
  regions?: string[];
  maxDuration: number;
  memory: number;
}

// Example configurations
const functionConfigs = {
  'api/users/[id]': {
    runtime: 'edge',
    regions: ['iad1', 'hnd1', 'fra1'],
    maxDuration: 30,
    memory: 512,
  },
  'api/generate-pdf': {
    runtime: 'nodejs18.x',
    maxDuration: 60,
    memory: 1024,
  },
};
```

---

**Module**: Configuration & Deployment
**Parent Skill**: moai-baas-vercel-ext
**Last Updated**: 2025-11-21
