# Skill: moai-baas-vercel-ext

## Metadata

```yaml
skill_id: moai-baas-vercel-ext
skill_name: Vercel Deployment & Edge Functions (Production Best Practices)
version: 2.0.0
created_date: 2025-11-09
updated_date: 2025-11-09
language: english
triggers:
  - keywords: ["Vercel", "Edge Functions", "Next.js", "Deployment", "ISR", "Serverless", "Production", "Performance"]
  - contexts: ["vercel-detected", "pattern-a", "pattern-b", "pattern-d"]
agents:
  - frontend-expert
  - devops-expert
freedom_level: high
word_count: 1000
context7_references:
  - url: "https://vercel.com/docs/deployments/overview"
    topic: "Deployment Strategy Comparison"
  - url: "https://vercel.com/docs/functions/edge-functions"
    topic: "Edge Functions Guide"
  - url: "https://vercel.com/docs/concepts/image-optimization"
    topic: "Image Optimization"
  - url: "https://vercel.com/docs/deployments/git"
    topic: "Git Integration & Preview Deployments"
  - url: "https://vercel.com/docs/concepts/functions/serverless-functions"
    topic: "Serverless Functions"
spec_reference: "@SPEC:BAAS-ECOSYSTEM-001"
```

---

## 📚 Content

### 1. Vercel Deployment Principles (150 words)

**Vercel** is a cloud deployment platform optimized for Next.js and edge computing.

**Deployment Process**:
```
Git Push (to main/develop)
   ↓
GitHub/GitLab webhook
   ↓
Vercel: Auto-build
   ├─ npm install
   ├─ npm run build (Next.js)
   └─ Optimization & compression
   ↓
Deploy to Edge Network (200+ locations)
   ↓
CDN cache enabled
   ↓
Live! (preview + production)
```

**Next.js Rendering Strategies**:

| Strategy | Build Time | Caching | Use Case |
|----------|---------|---------|---------|
| **SSG** | Build time | Permanent | Blogs, docs, landing pages |
| **ISR** | Background | Time-based | Semi-static content |
| **SSR** | Per request | None | Real-time data, personalized |
| **CSR** | Client-side | None | Dashboards, interactive apps |

**Example: ISR (Incremental Static Regeneration)**
```typescript
// pages/blog/[slug].tsx
export async function getStaticProps({ params }) {
  const post = await getPost(params.slug);

  return {
    props: { post },
    revalidate: 3600 // Regenerate hourly
  };
}
```

---

### 2. Edge Functions (200 words)

**Edge Functions**: Serverless functions running at edge closest to users.

**Serverless vs Edge**:

```
Client Request
   ↓
┌─────────────────────────────────┐
│ Edge Functions (Fast, Global)    │
├─────────────────────────────────┤
│ - Location: Regional edges (200+)│
│ - Response time: <100ms         │
│ - Max duration: 15 minutes       │
│ - Use: Auth, redirects, transforms
└─────────────────────────────────┘
   ↓ (Only when needed)
┌─────────────────────────────────┐
│ Serverless Functions (Central)   │
├─────────────────────────────────┤
│ - Location: Central datacenter   │
│ - Response time: 100-1000ms     │
│ - Cold start: 5 minutes         │
│ - Use: DB queries, compute, APIs│
└─────────────────────────────────┘
```

**Edge Middleware Example**:

```typescript
// middleware.ts - Auth check at edge
import { NextRequest, NextResponse } from 'next/server';

export async function middleware(req: NextRequest) {
  // 1. Verify token at edge (fast)
  const token = req.cookies.get('auth_token');

  if (!token) {
    return NextResponse.redirect(new URL('/login', req.url));
  }

  // 2. Optional: Fetch user from Supabase
  const res = await fetch('https://xxx.supabase.co/rest/v1/users', {
    headers: {
      'Authorization': `Bearer ${token.value}`,
      'apikey': process.env.NEXT_PUBLIC_SUPABASE_KEY
    }
  });

  if (!res.ok) {
    return NextResponse.redirect(new URL('/unauthorized', req.url));
  }

  return NextResponse.next();
}

export const config = {
  matcher: ['/dashboard/:path*', '/api/:path*']
};
```

**When to use Edge Functions**:
```typescript
// ✅ Perfect for Edge
- Authentication token validation
- Geo-based redirects
- A/B testing logic
- Request/response transformation

// ❌ Avoid on Edge
- Database queries (latency)
- File uploads
- Heavy computation
- Realtime subscriptions
```

---

### 3. Environment Variables (100 words)

**Environment setup**:

```bash
# .env.local (local development)
NEXT_PUBLIC_SUPABASE_URL=https://xxx.supabase.co
NEXT_PUBLIC_SUPABASE_ANON_KEY=eyJ...
SUPABASE_SERVICE_KEY=eyJ...  # Server-side only

# vercel.json (production)
{
  "env": {
    "NEXT_PUBLIC_SUPABASE_URL": "@supabase_url",
    "NEXT_PUBLIC_SUPABASE_ANON_KEY": "@supabase_key",
    "SUPABASE_SERVICE_KEY": "@supabase_service_key"
  }
}
```

**Secrets management**:
```bash
# Via Vercel CLI
vercel env add SUPABASE_SERVICE_KEY

# Or Dashboard
Settings → Environment Variables → Add
```

**Best Practices**:
- ✅ `NEXT_PUBLIC_` = safe for client (public data only)
- ✅ Service keys = server-only environment
- ❌ Never expose keys in logs
- ✅ Rotate secrets quarterly

---

### 4. Monitoring & Analytics (150 words)

**Web Vitals tracking**:

```typescript
// app/layout.tsx
import { Analytics } from '@vercel/analytics/react';

export default function RootLayout({ children }) {
  return (
    <html>
      <body>
        {children}
        <Analytics /> {/* Auto-tracking enabled */}
      </body>
    </html>
  );
}
```

**Key Metrics**:
- **LCP** (Largest Contentful Paint): Content load time (<2.5s target)
- **INP** (Interaction to Next Paint): Input responsiveness (<200ms target)
- **CLS** (Cumulative Layout Shift): Visual stability (<0.1 target)

**Performance optimization**:

```typescript
// 1. Code splitting with dynamic imports
const HeavyComponent = dynamic(() => import('./Heavy'), {
  loading: () => <Skeleton />
});

// 2. Image optimization (automatic)
import Image from 'next/image';

export default function Page() {
  return (
    <Image
      src="/photo.jpg"
      width={400}
      height={300}
      priority // LCP optimization
      // Vercel auto-optimizes:
      // - WebP conversion
      // - Responsive images
      // - Lazy loading
    />
  );
}

// 3. Font optimization
import { Inter } from 'next/font/google';
const inter = Inter({ subsets: ['latin'] });
```

**Error tracking & cost**:
- Dashboard → Logs → Errors for debugging
- Free tier: 100 builds/month, limited functions

---

### 5. Production Deployment Workflow (200 words)

**Branching strategy**:

```bash
# Feature development
git checkout -b feature/new-feature
npm run dev

# Build locally before pushing
npm run build

# Create preview deployment
git push origin feature/new-feature
# Vercel auto-creates preview URL

# Preview testing
# Visit: https://project-[random].vercel.app

# Merge to main for production
git checkout main
git merge feature/new-feature
git push origin main
# Auto-deploys to production
```

**Pre-deployment checklist**:
```typescript
// 1. Environment secrets set
vercel env list

// 2. Build succeeds locally
npm run build

// 3. Analytics enabled
import { Analytics } from '@vercel/analytics/react';

// 4. Error monitoring ready
Sentry.init({ dsn: process.env.NEXT_PUBLIC_SENTRY_DSN });

// 5. Database connections verified
const supabase = createClient(
  process.env.NEXT_PUBLIC_SUPABASE_URL,
  process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY
);
```

**Monitoring post-deployment**:

```bash
# Check deployment status
vercel deploy --prod

# View real-time logs
vercel logs

# Monitor Web Vitals
# Dashboard → Analytics → Web Vitals
# Target: LCP <2.5s, INP <200ms, CLS <0.1

# Error monitoring
# Dashboard → Logs → Errors
```

**Rollback procedure**:

```bash
# If deployment fails, Vercel auto-rollback to previous stable
# Or manual rollback:
vercel rollback
```

---

### 6. Performance & Cost Optimization (200 words)

**Performance optimization**:

```typescript
// 1. Implement incremental static regeneration
export async function getStaticProps() {
  return {
    props: { /* data */ },
    revalidate: 3600 // Regenerate hourly
  };
}

// 2. Use Edge Functions for auth/redirects (FREE)
// vs Serverless for DB queries (paid per invocation)

// 3. Compress images automatically
// Vercel handles: WebP, AVIF, responsive sizes

// 4. Code splitting for large bundles
const HeavyModal = dynamic(() => import('./HeavyModal'));

// 5. Font optimization prevents CLS
import { Inter } from 'next/font/google';
const inter = Inter({ display: 'swap' });
```

**Cost optimization strategies**:

| Item | Cost | Optimization |
|------|------|------|
| **Builds** | Free (100/month) | Merge carefully, use preview |
| **Serverless** | $0.50/1M requests | Use Edge Functions instead |
| **Edge** | Free (included) | Offload auth/redirects |
| **Data** | Included in plan | Monitor with Analytics |

**Monitoring costs**:

```bash
# Check usage dashboard
vercel analytics

# Review function invocations
vercel logs --follow

# Estimate monthly costs
# Free: up to $20 value
# Pro: $20/month for 1M serverless requests
```

---

## 🎯 Usage

### Agent Invocation

```python
# From frontend-expert or devops-expert
Skill("moai-baas-vercel-ext")

# Auto-loaded when Vercel patterns detected
```

### Context7 Auto-loading

When Vercel detected:
- Deployment strategy comparison (SSG vs ISR vs SSR)
- Edge Functions detailed guide
- Performance optimization checklist
- Production deployment workflow

---

## 📚 Reference Materials

- [Vercel Deployment Guide](https://vercel.com/docs/deployments/overview)
- [Edge Functions Documentation](https://vercel.com/docs/functions/edge-functions)
- [Image Optimization](https://vercel.com/docs/concepts/image-optimization)
- [Git Integration & Preview](https://vercel.com/docs/deployments/git)
- [Serverless Functions](https://vercel.com/docs/concepts/functions/serverless-functions)

---

## ✅ Validation Checklist

- [x] Deployment principles (SSG/ISR/SSR)
- [x] Edge Functions best practices
- [x] Environment variable management
- [x] Monitoring & Web Vitals analytics
- [x] Production deployment workflow
- [x] Performance optimization patterns
- [x] Cost monitoring & optimization
- [x] 1000+ word target (from 600)
- [x] English language (policy compliant)
