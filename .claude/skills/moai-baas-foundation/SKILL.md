# Skill: moai-baas-foundation

## Metadata

```yaml
skill_id: moai-baas-foundation
skill_name: BaaS Platform Foundation & 9-Platform Decision Framework (Ultra-comprehensive)
version: 2.0.0
created_date: 2025-11-09
updated_date: 2025-11-09
language: english
triggers:
  - keywords: ["BaaS", "backend-as-a-service", "platform selection", "architecture", "9 platforms", "Convex", "Firebase", "Cloudflare", "Auth0"]
  - contexts: ["/alfred:1-plan", "platform-selection", "architecture-decision", "pattern-a-h"]
agents:
  - spec-builder
  - backend-expert
  - database-expert
  - devops-expert
  - security-expert
  - frontend-expert
freedom_level: high
word_count: 1400
spec_reference: "@SPEC:BAAS-ECOSYSTEM-001"
```

---

## 📚 Content

### 1. BaaS (Backend-as-a-Service) Concepts & 9-Platform Overview (150 words)

**Backend-as-a-Service** is a cloud service model providing backend functionality without requiring server infrastructure management.

**Core Characteristics**:
- No infrastructure management needed (serverless)
- Immediately usable features (Auth, DB, Storage, Realtime)
- Automatic scaling
- Pay-as-you-go pricing model

**MoAI-ADK Supported 9-Platform Comparison**:

| Platform | Strengths | Weaknesses | Cost | Type |
|----------|-----------|-----------|------|------|
| **Supabase** | PostgreSQL + RLS + Auth | Single stack | Low | Postgres |
| **Vercel** | Edge Functions + Deployment | Limited monitoring | Medium | Deploy |
| **Neon** | DB branching + auto-scale | DB only | Medium | Postgres |
| **Clerk** | MFA + SSO + Security | Auth only | High | Auth |
| **Railway** | Full-stack integration | Limited customization | Low | Full-stack |
| **Convex** | Realtime Sync + Auth | Small community | Medium | Realtime |
| **Firebase** | Fully managed | High vendor lock-in | Low-Med | Full-stack |
| **Cloudflare** | Edge Workers + Speed | Learning curve | Low | Edge |
| **Auth0** | Enterprise authentication | High cost | High | Auth |

---

### 2. Eight Architecture Patterns (700 words)

#### **Pattern A: Full Supabase (Postgres Integration)**
```
PostgreSQL + RLS + Auth + Storage + Realtime + Vercel
```
- **Target**: MVP, small teams (< 5 people), rapid development
- **Cost**: Low ($0-100/month)
- **Strengths**: Best integration, RLS security, realtime features
- **Weaknesses**: PostgreSQL dependent, limited advanced auth

#### **Pattern B: Best-of-breed (Postgres + Enterprise Auth)**
```
Neon (DB) + Clerk (Auth) + Vercel (Deploy)
```
- **Target**: Production, large teams (5-50 people), advanced auth
- **Cost**: Medium ($100-500/month)
- **Strengths**: Peak performance, MFA/SSO, DB branching
- **Weaknesses**: 3-platform orchestration, increased complexity

#### **Pattern C: Railway All-in-one (Single Platform)**
```
Railway (PostgreSQL + Backend + Monitoring)
```
- **Target**: Solo developers, low-budget, Monolith preference
- **Cost**: Low ($5-50/month)
- **Strengths**: Simplicity, fast deployment, lowest cost
- **Weaknesses**: No advanced auth, limited flexibility

#### **Pattern D: Hybrid Premium (Postgres + Edge + Cloud)**
```
Supabase (DB) + Clerk (Auth) + Railway (Backend) + Vercel (Edge) + Cloudflare (CDN)
```
- **Target**: Complex requirements, maximum flexibility
- **Cost**: High ($200-1000+/month)
- **Strengths**: Maximum flexibility, all features, high security
- **Weaknesses**: 5-platform orchestration, operational complexity

#### **Pattern E: Firebase Full Stack (Google Ecosystem)**
```
Firebase (Auth + Firestore + Storage + Hosting + Functions)
```
- **Target**: Google ecosystem preference, rapid prototyping
- **Cost**: Low-Medium ($0-500/month)
- **Strengths**: Fully managed, integration, scalability
- **Weaknesses**: Firestore learning curve, lock-in, NoSQL

#### **Pattern F: Convex Realtime (Sync-first Architecture)**
```
Convex (Database + Sync + Auth + Functions + Hosting)
```
- **Target**: Realtime apps, modern frontend development
- **Cost**: Medium ($50-500/month)
- **Strengths**: Native realtime sync, TypeScript-first
- **Weaknesses**: Smaller community, emerging platform

#### **Pattern G: Cloudflare Edge-first (Performance Priority)**
```
Cloudflare Workers (Edge Functions) + D1 (Database) + Pages (Hosting)
```
- **Target**: Edge performance critical, global deployment
- **Cost**: Low ($0-200/month)
- **Strengths**: Ultra-low latency, edge deployment, low cost
- **Weaknesses**: Learning curve, sparse documentation

#### **Pattern H: Enterprise OAuth (Auth0 + Flexible Backend)**
```
Auth0 (Advanced Auth) + Free Choice (DB/Deploy/Backend)
```
- **Target**: Enterprise auth required, SAML/OIDC mandatory
- **Cost**: High ($1000+/month)
- **Strengths**: Enterprise features, SAML, Hooks
- **Weaknesses**: High cost, complex configuration

---

### 3. Decision Matrix (V2 - 9-Platform Based) (250 words)

**Level 1: Project Stage Classification**

```
MVP (Fast Launch Priority)
├─ Pattern A (Full Supabase) ← Recommended
├─ Pattern C (Railway) ← Minimal setup
└─ Pattern E (Firebase) ← Google ecosystem

Growth (Scalability + Features)
├─ Pattern B (Best-of-breed) ← Recommended
├─ Pattern F (Convex) ← Realtime priority
└─ Pattern D (Hybrid) ← Maximum flexibility

Scale (Enterprise + High Availability)
├─ Pattern D (Hybrid Premium) ← Recommended
├─ Pattern H (Auth0 + Free) ← Enterprise auth
└─ Pattern G (Cloudflare) ← Edge performance
```

**Level 2: Team Size vs Features**

```
Solo (1 person) → Pattern C (Railway) or Pattern A (Supabase)
Small (2-4 people) → Pattern A (Supabase) or Pattern E (Firebase)
Medium (5-15 people) → Pattern B (Best-of-breed) or Pattern F (Convex)
Large (15+ people) → Pattern D (Hybrid) or Pattern H (Enterprise)
```

**Level 3: Special Requirements**

```
Realtime app required → Pattern F (Convex) or Pattern A (Supabase Realtime)
Edge performance critical → Pattern G (Cloudflare) or Pattern D (with Vercel Edge)
Enterprise auth → Pattern H (Auth0) or Pattern D (Clerk)
Google ecosystem → Pattern E (Firebase)
Maximum control needed → Pattern D (Hybrid Premium)
```

**Priority Weighting**:
1. **Team size** (40%): Largest impact
2. **Project stage** (30%)
3. **Special requirements** (20%)
4. **Budget** (10%)

---

### 4. Real-World Pain Points & Solutions (150 words)

| Pain Point | Pattern Solution | Implementation |
|-----------|------------------|-----------------|
| **RLS Debugging** | Pattern A, D | Supabase Logs, pgTAP tests |
| **Data Sync** | Pattern F, A | Convex Sync or Supabase Realtime |
| **Global Latency** | Pattern G | Cloudflare Workers + Pages |
| **Enterprise Auth** | Pattern H, D | Auth0 + SAML/OIDC |
| **DB Branching/Dev** | Pattern B | Neon development instances |
| **Cost Optimization** | Pattern C | Railway single platform |
| **Type Safety** | Pattern F | Convex TypeScript definitions |
| **Lock-in Avoidance** | Pattern D | Multi-platform approach |

---

### 5. Real Project Scenarios & Pattern Selection (100 words)

**Scenario 1: SaaS MVP (Early Stage Startup)**
```
Requirements: User auth, database, file uploads, billing
Timeline: Launch in 4 weeks
Budget: $0-200/month
Team: 2 developers

RECOMMENDED PATTERN: A (Full Supabase)
├─ Supabase: Auth + DB + RLS + Storage + Realtime
├─ Vercel: Frontend deployment
├─ Stripe: Billing integration
└─ Cost: ~$50/month
```

**Scenario 2: Realtime Collaboration App**
```
Requirements: Live sync, presence, conflict resolution
Timeline: MVP in 8 weeks
Budget: $100-500/month
Team: 4 developers

RECOMMENDED PATTERN: F (Convex Realtime)
├─ Convex: TypeScript-first realtime sync + auth
├─ Vercel: Edge deployment
└─ Cost: ~$200/month
```

**Scenario 3: Enterprise Dashboard**
```
Requirements: SAML/OIDC, MFA, audit logs, complex roles
Timeline: 12 weeks
Budget: $1000+/month
Team: 10+ developers

RECOMMENDED PATTERN: D or H (Hybrid Premium + Auth0)
├─ Supabase or Neon: Database with audit
├─ Clerk or Auth0: Enterprise auth
├─ Vercel: Frontend + Edge functions
└─ Cost: $500-2000/month
```

**Scenario 4: Global Edge-First App**
```
Requirements: <100ms latency globally, AI/ML APIs
Timeline: 6 weeks
Budget: Low cost
Team: 2-3 developers

RECOMMENDED PATTERN: G (Cloudflare Edge-first)
├─ Cloudflare Workers: Global edge functions
├─ D1: Distributed database
├─ Vercel: Static frontend
└─ Cost: $50-300/month
```

---

### 6. Platform Migration Strategy & Best Practices (100 words)

**Migration Path: Pattern A → Pattern B (Supabase → Best-of-breed)**

```
Phase 1: Parallel Setup (1-2 weeks)
├─ Set up Neon database with PostgreSQL backup
├─ Migrate Supabase schema via pg_dump/psql
├─ Configure Clerk for new auth system
└─ Run data validation tests

Phase 2: Gradual Migration (2-4 weeks)
├─ Route 10% new users to Clerk + Neon
├─ Monitor error rates & performance
├─ Gradually increase to 50% → 100%
└─ Keep Supabase auth as fallback

Phase 3: Cleanup (1 week)
├─ Archive old Supabase data
├─ Remove legacy auth routes
└─ Document new architecture
```

**Zero-downtime Migration Checklist**:
- ✅ Database schema compatibility verified
- ✅ Data migration tested in staging
- ✅ Dual-write setup for consistency
- ✅ Rollback plan documented
- ✅ Performance baseline established
- ✅ Team training completed

**Cost-Benefit Analysis Template**:

| Factor | Supabase | Neon + Clerk | Break-even |
|--------|----------|--------------|-----------|
| Monthly cost | $100 | $250 | ~3 months |
| Development time | 40h | 60h | +20h investment |
| Scaling capacity | Medium | High | 10x users |
| Team velocity | +15% | +25% | 2x productivity |

**Key Migration Decisions**:
1. **Hot vs Cold migration**: Hot = zero downtime, Cold = downtime required
2. **Big-bang vs Gradual**: Gradual safer for production apps
3. **Rollback timing**: Plan for 24-48 hours post-migration rollback window
4. **Data validation**: Mandatory checksums on critical data

---

## 🎯 Usage

### Invocation from Agents
```python
Skill("moai-baas-foundation")
# Result: Clear understanding of all 9 platforms and 8 patterns
```

### Usage Scenarios
```
User: /alfred:1-plan "Add backend"
↓
spec-builder: Load moai-baas-foundation
↓
Detect 1-9 platforms in project
↓
AskUserQuestion: Present 8 patterns (A-H)
↓
User: Select pattern
↓
Load extension Skills (moai-baas-supabase-ext, etc.)
```

---

## 📚 Reference Materials

- SPEC-BAAS-ECOSYSTEM-001 (main specification)
- moai-baas-supabase-ext, moai-baas-vercel-ext (existing)
- moai-baas-neon-ext, moai-baas-clerk-ext (Phase 2)
- moai-baas-convex-ext, moai-baas-firebase-ext (Phase 3)
- moai-baas-cloudflare-ext, moai-baas-auth0-ext (Phase 4)
- moai-baas-railway-ext (Phase 5)

---

## ✅ Validation Checklist

- [x] 9-platform overview
- [x] 8 architecture patterns (A-H)
- [x] Decision matrix (V2) with 3-level classification
- [x] Pain points & solutions (expanded)
- [x] Real project scenarios (4 detailed use cases)
- [x] Platform migration strategy (zero-downtime guide)
- [x] 1400-word target (from 1200)
- [x] English language (policy compliant)
