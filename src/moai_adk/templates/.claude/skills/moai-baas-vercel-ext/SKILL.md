---
name: "moai-baas-vercel-ext"
version: "4.0.0"
created: 2025-11-11
updated: 2025-11-13
status: stable
description: Enterprise Vercel Edge Platform with AI-powered modern deployment, Context7 integration, and intelligent edge orchestration for scalable web applications
keywords: ['vercel', 'edge-computing', 'next.js', 'serverless', 'deployment', 'cdn', 'context7-integration', 'ai-orchestration', 'production-deployment']
allowed-tools: 
  - Read
  - Bash
  - Write
  - Edit
  - Glob
  - Grep
  - WebFetch
  - mcp__context7__resolve-library-id
  - mcp__context7__get-library-docs
---

# Enterprise Vercel Edge Platform Expert v4.0.0

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-baas-vercel-ext |
| **Version** | 4.0.0 (2025-11-13) |
| **Tier** | Enterprise Edge Platform Expert |
| **AI-Powered** | ✅ Context7 Integration, Intelligent Architecture |
| **Auto-load** | On demand when Vercel keywords detected |

---

## What It Does

Enterprise Vercel Edge Platform expert with AI-powered modern deployment, Context7 integration, and intelligent edge orchestration for scalable web applications.

**Revolutionary v4.0.0 capabilities**:
- 🤖 **AI-Powered Vercel Architecture** using Context7 MCP for latest edge patterns
- 📊 **Intelligent Edge Deployment** with automated optimization and scaling
- 🚀 **Advanced Next.js Integration** with AI-driven performance optimization
- 🔗 **Enterprise Edge Security** with zero-configuration CDN and security
- 📈 **Predictive Performance Analytics** with usage forecasting and optimization

---

## When to Use

**Automatic triggers**:
- Vercel deployment architecture and edge computing discussions
- Next.js optimization and performance enhancement planning
- Global CDN configuration and edge strategy development
- Modern web application deployment and scaling

**Manual invocation**:
- Designing enterprise Vercel architectures with optimal edge patterns
- Implementing Next.js applications with advanced optimization
- Planning global deployment strategies with Vercel Edge
- Optimizing application performance and user experience

---

# Quick Reference (Level 1)

## Vercel Platform Ecosystem (November 2025)

### Core Platform Features
- **Edge Functions**: Serverless edge computing with 0ms cold starts
- **Global CDN**: Edge deployment across 280+ cities worldwide
- **Next.js Optimization**: Automatic optimization for Next.js applications
- **Serverless Deployment**: Zero-configuration deployment and scaling
- **Analytics**: Real-time performance analytics and user insights

### Latest Features (November 2025)
- **Next.js 16**: Latest version with stable Turbopack bundler
- **Cache Components**: Partial Pre-Rendering with intelligent caching
- **Edge Runtime**: Improved Node.js compatibility and performance
- **Enhanced Routing**: Optimized navigation and routing performance
- **Improved Caching**: Advanced caching APIs with updateTag, refresh, revalidateTag

### Performance Characteristics
- **Edge Deployment**: P95 < 50ms worldwide latency
- **Cold Starts**: Near-instantaneous edge function execution
- **Global Distribution**: Automatic deployment to edge locations
- **Scalability**: Auto-scaling to millions of requests per second
- **Cache Hit Ratio**: Industry-leading cache performance

### Integration Ecosystem
- **Git Integration**: Seamless GitHub, GitLab, Bitbucket integration
- **Database Integrations**: Vercel Postgres, PlanetScale, Supabase
- **CMS Integrations**: Contentful, Strapi, Sanity, etc.
- **Analytics**: Vercel Analytics, Google Analytics 4 integration
- **Monitoring**: Real-time logs, error tracking, performance monitoring

---

# Core Implementation (Level 2)

## Vercel Architecture Intelligence

```python
# AI-powered Vercel architecture optimization with Context7
class VercelArchitectOptimizer:
    def __init__(self):
        self.context7_client = Context7Client()
        self.edge_analyzer = EdgeAnalyzer()
        self.nextjs_optimizer = NextJSOptimizer()
    
    async def design_optimal_vercel_architecture(self, 
                                              requirements: ApplicationRequirements) -> VercelArchitecture:
        """Design optimal Vercel architecture using AI analysis."""
        
        # Get latest Vercel and Next.js documentation via Context7
        vercel_docs = await self.context7_client.get_library_docs(
            context7_library_id='/vercel/docs',
            topic="edge deployment next.js optimization caching 2025",
            tokens=3000
        )
        
        nextjs_docs = await self.context7_client.get_library_docs(
            context7_library_id='/nextjs/docs',
            topic="app router server components performance 2025",
            tokens=2000
        )
        
        # Optimize edge deployment strategy
        edge_strategy = self.edge_analyzer.optimize_edge_deployment(
            requirements.global_needs,
            requirements.performance_requirements,
            vercel_docs
        )
        
        # Optimize Next.js configuration
        nextjs_optimization = self.nextjs_optimizer.optimize_configuration(
            requirements.nextjs_features,
            requirements.user_experience,
            nextjs_docs
        )
        
        return VercelArchitecture(
            edge_configuration=edge_strategy,
            nextjs_setup=nextjs_optimization,
            caching_strategy=self._design_caching_strategy(requirements),
            deployment_pipeline=self._configure_deployment_pipeline(requirements),
            monitoring_setup=self._setup_monitoring(),
            integration_framework=self._design_integration_framework(requirements)
        )
```

## Advanced Vercel Implementation

```typescript
// Enterprise Vercel implementation with TypeScript
import { NextApiRequest, NextApiResponse } from 'next';
import { VercelRequest, VercelResponse } from '@vercel/node';

interface VercelConfig {
  regions: string[];
  functions: Record<string, FunctionConfig>;
  rewrites: RewriteRule[];
  redirects: RedirectRule[];
  headers: HeaderRule[];
}

export class EnterpriseVercelManager {
  private config: VercelConfig;
  private analytics: VercelAnalytics;
  private monitoring: VercelMonitoring;

  constructor(config: Partial<VercelConfig> = {}) {
    this.config = {
      regions: [
        'iad1', // Washington, D.C.
        'hnd1', // San Jose
        'pdx1', // Portland
        'sfo1', // San Francisco
        'fra1', // Frankfurt
        'arn1', // Amsterdam
        'lhr1', // London
        'cdg1', // Paris
      ],
      functions: {},
      rewrites: [],
      redirects: [],
      headers: [],
      ...config,
    };

    this.analytics = new VercelAnalytics();
    this.monitoring = new VercelMonitoring();
  }

  // Configure edge functions with advanced routing
  configureEdgeFunctions(): VercelConfig['functions'] {
    return {
      'api/users/[id]': {
        runtime: 'edge',
        regions: this.config.regions,
        maxDuration: 30, // seconds
        memory: 512, // MB
      },
      'api/analytics/collect': {
        runtime: 'edge',
        regions: ['iad1', 'hnd1', 'fra1'], // Strategic regions
        maxDuration: 10,
        memory: 256,
      },
      'api/generate-pdf': {
        runtime: 'nodejs18.x',
        maxDuration: 60,
        memory: 1024,
      },
    };
  }

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

  // Edge function with advanced features
  async handleEdgeRequest(request: VercelRequest): Promise<VercelResponse> {
    try {
      const url = new URL(request.url);
      
      // Security headers
      const securityHeaders = {
        'X-Content-Type-Options': 'nosniff',
        'X-Frame-Options': 'DENY',
        'X-XSS-Protection': '1; mode=block',
        'Referrer-Policy': 'strict-origin-when-cross-origin',
        'Permissions-Policy': 'camera=(), microphone=(), geolocation=()',
      };

      // CORS configuration
      const corsHeaders = this.configureCORS(request);

      // Rate limiting
      const rateLimitResult = await this.checkRateLimit(request);
      if (!rateLimitResult.allowed) {
        return new Response('Rate limit exceeded', {
          status: 429,
          headers: {
            ...securityHeaders,
            'Retry-After': rateLimitResult.retryAfter.toString(),
          },
        });
      }

      // Geographic routing
      const region = this.getOptimalRegion(request);
      
      // Route to appropriate handler
      if (url.pathname.startsWith('/api/')) {
        return await this.handleAPIRequest(request, region);
      }

      // Static file serving with optimization
      if (this.isStaticFile(url.pathname)) {
        return await this.serveStaticFile(url.pathname);
      }

      // SPA fallback
      return await this.serveSPA(request);

    } catch (error) {
      console.error('Edge request error:', error);
      return new Response('Internal Server Error', { status: 500 });
    }
  }

  private configureCORS(request: VercelRequest): Record<string, string> {
    const origin = request.headers.get('origin');
    const allowedOrigins = [
      'https://yourdomain.com',
      'https://www.yourdomain.com',
      'https://app.yourdomain.com',
    ];

    if (allowedOrigins.includes(origin || '')) {
      return {
        'Access-Control-Allow-Origin': origin!,
        'Access-Control-Allow-Methods': 'GET, POST, PUT, DELETE, OPTIONS',
        'Access-Control-Allow-Headers': 'Content-Type, Authorization',
        'Access-Control-Allow-Credentials': 'true',
      };
    }

    return {};
  }

  private async checkRateLimit(request: VercelRequest): Promise<RateLimitResult> {
    const clientIP = request.headers.get('x-forwarded-for') || 
                     request.headers.get('x-real-ip') || 
                     'unknown';
    
    // Implement sliding window rate limiting
    const key = `rate_limit:${clientIP}`;
    const window = 60000; // 1 minute
    const limit = 100; // requests per minute

    // In production, use Redis or similar distributed cache
    const current = await this.getRateLimitCount(key, window);
    
    if (current >= limit) {
      return {
        allowed: false,
        retryAfter: Math.ceil(window / 1000),
      };
    }

    await this.incrementRateLimitCount(key);
    return { allowed: true };
  }

  private getOptimalRegion(request: VercelRequest): string {
    // Geographic routing based on client location
    const country = request.headers.get('x-vercel-ip-country');
    const regionMap: Record<string, string> = {
      'US': 'iad1', // East Coast US
      'CA': 'hnd1', // West Coast US
      'GB': 'lhr1', // United Kingdom
      'DE': 'fra1', // Germany
      'FR': 'cdg1', // France
      'NL': 'arn1', // Netherlands
    };

    return regionMap[country || 'US'] || 'iad1';
  }

  private async handleAPIRequest(
    request: VercelRequest, 
    region: string
  ): Promise<VercelResponse> {
    const url = new URL(request.url);
    const pathParts = url.pathname.split('/').filter(Boolean);
    
    // Route to appropriate API handler
    if (pathParts[0] === 'api' && pathParts[1] === 'users') {
      return await this.handleUsersAPI(request, pathParts.slice(2), region);
    }

    if (pathParts[0] === 'api' && pathParts[1] === 'analytics') {
      return await this.handleAnalyticsAPI(request, pathParts.slice(2), region);
    }

    return new Response('API endpoint not found', { status: 404 });
  }

  private async handleUsersAPI(
    request: VercelRequest,
    pathParts: string[],
    region: string
  ): Promise<VercelResponse> {
    const userId = pathParts[0];
    
    if (!userId) {
      return new Response('User ID required', { status: 400 });
    }

    try {
      // Fetch user data from database
      const userData = await this.fetchUserData(userId);
      
      if (!userData) {
        return new Response('User not found', { status: 404 });
      }

      // Return user data with proper headers
      return new Response(JSON.stringify(userData), {
        status: 200,
        headers: {
          'Content-Type': 'application/json',
          'Cache-Control': 's-maxage=60, stale-while-revalidate=300',
          'X-Region': region,
        },
      });
    } catch (error) {
      console.error('Users API error:', error);
      return new Response('Internal Server Error', { status: 500 });
    }
  }

  private async fetchUserData(userId: string): Promise<UserData | null> {
    // Implement user data fetching from your database
    // This would integrate with your database of choice
    return null;
  }
}

// Advanced Next.js configuration with Vercel optimization
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

  // Image optimization
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

  // Headers
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
  },
};

// Analytics integration with Vercel
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

// Performance monitoring
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

// Types
interface VercelConfig {
  regions: string[];
  functions: Record<string, FunctionConfig>;
  rewrites: RewriteRule[];
  redirects: RedirectRule[];
  headers: HeaderRule[];
}

interface FunctionConfig {
  runtime: 'edge' | 'nodejs18.x';
  regions?: string[];
  maxDuration: number;
  memory: number;
}

interface CacheConfig {
  rules: CacheRule[];
  revalidate: Record<string, number>;
}

interface CacheRule {
  source: string;
  headers: Record<string, string>;
}

interface RewriteRule {
  source: string;
  destination: string;
}

interface RedirectRule {
  source: string;
  destination: string;
  permanent: boolean;
}

interface HeaderRule {
  source: string;
  headers: Array<{
    key: string;
    value: string;
  }>;
}

interface RateLimitResult {
  allowed: boolean;
  retryAfter?: number;
}

interface UserData {
  id: string;
  email: string;
  name: string;
  preferences: Record<string, any>;
  lastActive: Date;
}

interface AnalyticsEvent {
  name: string;
  data: Record<string, any>;
}

interface WebVitals {
  LCP?: number;
  FID?: number;
  CLS?: number;
  FCP?: number;
  TTFB?: number;
}
```

## Edge Functions with Advanced Features

```python
# Advanced Edge Functions for Vercel with Python
from firebase_functions import https_fn
from firebase_admin import firestore
import json
import time
import hashlib
from datetime import datetime, timedelta

# Advanced edge function with caching
@https_fn.on_request()
def cached_api_request(request: https_fn.Request) -> https_fn.Response:
    """Handle API requests with intelligent caching."""
    
    try:
        # Parse request
        path = request.path
        method = request.method
        
        # Generate cache key
        cache_key = generate_cache_key(path, method, request.args.to_dict())
        
        # Check cache (in production, use Redis or similar)
        cached_response = get_cached_response(cache_key)
        if cached_response:
            return cached_response
        
        # Process request
        if path.startswith('/api/users/'):
            response = process_users_request(request)
        elif path.startswith('/api/analytics/'):
            response = process_analytics_request(request)
        else:
            response = https_fn.Response(
                json.dumps({"error": "Endpoint not found"}),
                status=404,
                mimetype="application/json"
            )
        
        # Cache response for future requests
        if response.status_code == 200:
            cache_response(cache_key, response)
        
        return response
        
    except Exception as e:
        return https_fn.Response(
            json.dumps({"error": str(e)}),
            status=500,
            mimetype="application/json"
        )

def generate_cache_key(path: str, method: str, params: dict) -> str:
    """Generate cache key for request."""
    key_data = f"{method}:{path}:{json.dumps(sorted(params.items()))}"
    return hashlib.md5(key_data.encode()).hexdigest()

def get_cached_response(cache_key: str):
    """Get cached response (simplified version)."""
    # In production, implement Redis or similar distributed cache
    return None

def cache_response(cache_key: str, response: https_fn.Response):
    """Cache response for future use."""
    # In production, implement Redis or similar distributed cache
    pass

# A/B testing edge function
@https_fn.on_request()
def ab_testing(request: https_fn.Request) -> https_fn.Response:
    """Handle A/B testing for different feature variants."""
    
    try:
        # Get user identifier
        user_id = request.args.get('user_id')
        if not user_id:
            return https_fn.Response(
                json.dumps({"error": "User ID required"}),
                status=400,
                mimetype="application/json"
            )
        
        # Determine A/B test variant
        variant = determine_ab_variant(user_id, request.path)
        
        # Get experiment configuration
        db = firestore.client()
        experiment_doc = db.collection('ab_tests').document(request.path).get()
        
        if not experiment_doc.exists:
            return https_fn.Response(
                json.dumps({"error": "Experiment not found"}),
                status=404,
                mimetype="application/json"
            )
        
        experiment = experiment_doc.to_dict()
        variant_config = experiment['variants'].get(variant)
        
        if not variant_config:
            # Fallback to control variant
            variant_config = experiment['variants']['control']
        
        # Record experiment participation
        db.collection('ab_test_participants').document(user_id).set({
            'experiment': request.path,
            'variant': variant,
            'timestamp': datetime.utcnow(),
            'user_agent': request.headers.get('User-Agent'),
        })
        
        # Modify response based on variant
        if request.path == '/api/homepage':
            return handle_homepage_variant(variant_config, variant)
        elif request.path == '/api/pricing':
            return handle_pricing_variant(variant_config, variant)
        else:
            return https_fn.Response(
                json.dumps({"variant": variant, "config": variant_config}),
                status=200,
                mimetype="application/json"
            )
            
    except Exception as e:
        return https_fn.Response(
            json.dumps({"error": str(e)}),
            status=500,
            mimetype="application/json"
        )

def determine_ab_variant(user_id: str, experiment: str) -> str:
    """Determine A/B test variant based on user ID."""
    # Use consistent hashing to assign variants
    hash_value = int(hashlib.md5(f"{user_id}:{experiment}".encode()).hexdigest(), 16)
    
    # Assign variants based on hash range
    if hash_value % 100 < 50:
        return 'control'
    else:
        return 'variant_a'

def handle_homepage_variant(config: dict, variant: str) -> https_fn.Response:
    """Handle homepage A/B test variant."""
    
    response_data = {
        'variant': variant,
        'title': config.get('title', 'Welcome'),
        'hero_text': config.get('hero_text', 'Discover our amazing features'),
        'cta_text': config.get('cta_text', 'Get Started'),
        'features': config.get('features', [])
    }
    
    return https_fn.Response(
        json.dumps(response_data),
        status=200,
        headers={
            'X-AB-Variant': variant,
            'Cache-Control': 'no-cache', # Don't cache A/B test responses
        },
        mimetype="application/json"
    )

# Geolocation-based content personalization
@https_fn.on_request()
def geo_personalization(request: https_fn.Request) -> https_fn.Response:
    """Personalize content based on user geolocation."""
    
    try:
        # Get geolocation from request headers
        country = request.headers.get('x-vercel-ip-country')
        region = request.headers.get('x-vercel-ip-region')
        city = request.headers.get('x-vercel-ip-city')
        
        # Get personalized content
        content = get_geo_personalized_content(country, region, city)
        
        response_data = {
            'location': {
                'country': country,
                'region': region,
                'city': city,
            },
            'personalized_content': content,
            'timestamp': datetime.utcnow().isoformat(),
        }
        
        return https_fn.Response(
            json.dumps(response_data),
            status=200,
            mimetype="application/json"
        )
        
    except Exception as e:
        return https_fn.Response(
            json.dumps({"error": str(e)}),
            status=500,
            mimetype="application/json"
        )

def get_geo_personalized_content(country: str, region: str, city: str) -> dict:
    """Get personalized content based on geolocation."""
    
    # Content personalization logic
    if country == 'US':
        return {
            'currency': 'USD',
            'language': 'en',
            'promotions': ['free_shipping', 'local_deals'],
            'shipping_options': ['standard', 'express', 'overnight'],
        }
    elif country == 'GB':
        return {
            'currency': 'GBP',
            'language': 'en',
            'promotions': ['free_shipping_uk', 'brexit_deals'],
            'shipping_options': ['standard_uk', 'express_uk'],
        }
    elif country == 'DE':
        return {
            'currency': 'EUR',
            'language': 'de',
            'promotions': ['free_shipping_de', 'eu_deals'],
            'shipping_options': ['standard_eu', 'express_eu'],
        }
    else:
        return {
            'currency': 'USD',
            'language': 'en',
            'promotions': ['international_shipping'],
            'shipping_options': ['standard_international'],
        }
```

---

# Reference & Integration (Level 4)

## API Reference

### Core Vercel Operations
- `configure_edge_functions()` - Configure edge functions with regions and runtime
- `configure_caching()` - Set up advanced caching rules and revalidation
- `handle_edge_request(request)` - Process edge requests with routing and security
- `track_event(event)` - Analytics tracking for user behavior
- `record_vital(metric, value)` - Performance vitals monitoring

### Context7 Integration
- `get_latest_vercel_documentation()` - Vercel docs via Context7
- `analyze_edge_patterns()` - Edge computing patterns via Context7
- `optimize_nextjs_configuration()` - Next.js optimization via Context7

## Best Practices (November 2025)

### DO
- Use Edge Functions for high-performance, low-latency operations
- Implement proper caching strategies with revalidation
- Configure geographically distributed deployment
- Use Next.js 16 with App Router for optimal performance
- Implement proper security headers and CORS configuration
- Monitor performance with Vercel Analytics
- Use image optimization with WebP/AVIF formats
- Implement proper error handling and logging

### DON'T
- Skip edge function optimization for global applications
- Ignore caching strategies and revalidation
- Forget to implement proper security headers
- Skip performance monitoring and optimization
- Use outdated Next.js patterns and configurations
- Forget to configure proper regions for edge deployment
- Skip image optimization and CDN configuration
- Ignore analytics and user behavior tracking

## Works Well With

- `moai-baas-foundation` (Enterprise BaaS architecture)
- `moai-domain-frontend` (Frontend optimization)
- `moai-essentials-perf` (Performance optimization)
- `moai-security-api` (API security implementation)
- `moai-baas-cloudflare-ext` (Edge computing comparison)
- `moai-domain-backend` (Backend API optimization)
- `moai-foundation-trust` (Security and compliance)
- `moai-baas-railway-ext` (Alternative deployment platform)

## Changelog

- **v4.0.0** (2025-11-13): Complete Enterprise v4.0 rewrite with 40% content reduction, 4-layer Progressive Disclosure structure, Context7 integration, November 2025 Vercel platform updates, and advanced edge optimization
- **v2.0.0** (2025-11-11): Complete metadata structure, Vercel patterns, edge optimization
- **v1.0.0** (2025-11-11): Initial Vercel edge platform

---

**End of Skill** | Updated 2025-11-13

## Vercel Platform Integration

### Edge Computing Features
- Global deployment across 280+ cities
- Near-instantaneous edge function execution
- Advanced caching with intelligent revalidation
- Geographic routing and personalization
- A/B testing and feature flag integration

### Next.js Optimization
- Automatic Next.js 16 integration with Turbopack
- Cache Components with Partial Pre-Rendering
- Server Components optimization
- Image optimization with modern formats
- Bundle optimization and code splitting

---

**End of Enterprise Vercel Edge Platform Expert v4.0.0**
