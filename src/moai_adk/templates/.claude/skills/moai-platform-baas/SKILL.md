---
name: moai-platform-baas
description: Comprehensive unified BaaS (Backend-as-a-Service) integration hub for 9 major providers: Auth0, Clerk, Firebase Auth, Supabase, Neon, Convex, Firebase Firestore, Vercel, and Railway with AI-powered provider selection, cross-provider patterns, and migration guides.
version: 2.0.0
aliases: [moai-platform-baas]
category: platform
modularized: false
updated: 2025-11-27
status: active
tags:
  - platform
  - baas
  - backend
  - integration
  - firebase
  - supabase
  - auth0
  - clerk
  - vercel
  - railway
allowed-tools: Read, Write, Bash, Grep, Glob
deprecated_names:
  moai-platform-baas:
    deprecated_in: v0.32.0
    remove_in: v0.35.0
    message: "Use moai-platform-baas instead"
---

# moai-platform-baas: Comprehensive BaaS Integration Hub

## Quick Reference (30 seconds)

**Enterprise BaaS Provider Unification**: One skill to manage authentication, database, and deployment across 9 major BaaS providers with AI-powered provider selection, cross-provider integration patterns, and automated migration assistance.

### Core Provider Matrix

| Category | Providers | Key Strength |
|----------|-----------|--------------|
| **Authentication** | Auth0, Clerk, Firebase Auth | SSO, Modern Auth, Google Integration |
| **Database** | Supabase, Neon, Convex, Firestore | PostgreSQL 16, Serverless, Real-time |
| **Deployment** | Vercel, Railway | Edge Optimization, Full-stack |

### Key Capabilities
- ✅ **AI Provider Selection**: Automatically choose optimal providers based on requirements
- ✅ **Cross-Provider Integration**: Seamless setup between auth, database, and deployment
- ✅ **Migration Guides**: Step-by-step migration between any providers
- ✅ **Production Patterns**: Enterprise-ready configurations and security
- ✅ **Cost Optimization**: Provider-specific cost analysis and recommendations
- ✅ **Context7 Integration**: Latest API documentation and best practices

### When to Use
- **New Project**: AI-driven provider selection for optimal stack
- **Migration**: Smooth transition between BaaS providers
- **Multi-Provider**: Complex integrations across different services
- **Cost Optimization**: Analyze and reduce BaaS infrastructure costs
- **Security Audits**: Ensure compliance across all providers

### Quick Provider Selection
```python
# AI-powered provider recommendation
ai_selector = BaaSProviderSelector()
recommendation = ai_selector.select_optimal_stack({
    "requirements": ["scalable", "real-time", "enterprise"],
    "tech_stack": "Next.js",
    "team_size": 5,
    "budget": "$500/month"
})
# Returns: { auth: "Clerk", database: "Supabase", deployment: "Vercel" }
```

---

## Implementation Guide

### Phase 1: AI Provider Selection & Setup

**Requirements Analysis & Provider Selection**:
```python
async def analyze_baas_requirements(project_context: ProjectContext) -> ProviderRecommendation:
    """AI-powered provider selection based on project requirements."""
    
    # Get latest Context7 documentation for all providers
    context7_docs = await fetch_all_provider_docs()
    
    # Analyze requirements against provider capabilities
    analysis = ProviderAnalyzer().analyze_requirements(project_context, context7_docs)
    
    # Generate weighted recommendation
    return ProviderRecommender().generate_recommendation(analysis)
```

**Unified Provider Configuration**:
```python
# Unified provider setup (works for all 9 providers)
provider_manager = UnifiedBaaSManager()

# Authentication setup
auth_config = provider_manager.configure_auth({
    "provider": "clerk",  # or "auth0", "firebase-auth"
    "features": ["social_auth", "mfa", "organizations"]
})

# Database setup  
db_config = provider_manager.configure_database({
    "provider": "supabase",  # or "neon", "convex", "firestore"
    "schema_path": "./schema.sql",
    "migrations": True
})

# Deployment setup
deploy_config = provider_manager.configure_deployment({
    "provider": "vercel",  # or "railway"
    "framework": "nextjs",
    "environment": "production"
})
```

### Phase 2: Authentication Providers

**Auth0 (Enterprise SSO Focus)**:
- Enterprise SSO with 50+ connections (SAML, OIDC, ADFS)
- B2B SaaS with organizations and RBAC
- Custom database connections and advanced security

**Clerk (Modern Auth Focus)**:
- Modern WebAuthn and passkey support
- Built-in organization management
- Multi-platform SDKs and beautiful UI components

**Firebase Auth (Google Integration)**:
- Deep Google services integration
- Firebase Analytics and Cloud Functions
- Mobile-first design with Google Cloud integration

### Phase 3: Database Providers

**Supabase (PostgreSQL 16+ Focus)**:
- PostgreSQL 16 with pgvector and AI extensions
- Row-Level Security for multi-tenant apps
- Real-time subscriptions and Edge Functions

**Neon (Serverless PostgreSQL)**:
- Auto-scaling serverless PostgreSQL
- Instant database branching
- 30-day Point-in-Time Recovery

**Convex (Real-time Backend)**:
- Real-time reactive queries and optimistic updates
- Instant database branching for development
- TypeScript-first design with built-in caching

**Firebase Firestore (Mobile Focus)**:
- Real-time synchronization with offline caching
- Mobile-first SDKs for iOS and Android
- Google ecosystem integration

### Phase 4: Deployment Platforms

**Vercel (Edge Deployment)**:
- Global edge network with Next.js optimization
- Edge Functions and zero-config deployments
- Analytics and built-in observability

**Railway (Full-stack Containers)**:
- Full-stack container deployment with Docker
- Multi-region support and built-in CI/CD
- Environment variables management

### Phase 5: Cross-Provider Integration

**Modern Web Stack (Vercel + Clerk + Supabase)**:
```python
class ModernWebStack:
    """Vercel + Clerk + Supabase integration."""
    
    def setup_integration(self):
        """Seamless integration setup."""
        return {
            "authentication": "clerk",
            "database": "supabase", 
            "deployment": "vercel",
            "real_time_features": True,
            "server_functions": True,
            "edge_optimization": True
        }
```

**Enterprise Stack (Auth0 + Supabase + Vercel)**:
- Enterprise SSO with 50+ connections
- PostgreSQL 16 with Row-Level Security
- Global edge performance

**Real-time Stack (Clerk + Convex + Vercel)**:
- Modern authentication with organizations
- Real-time collaborative features
- Edge performance for global users

---

## Provider Selection Decision Trees

### Authentication Provider Selection
```
Need Enterprise SSO with 50+ connections?
  ├─ YES → Auth0 (Enterprise grade)
  │        ├─ B2B SaaS focus? → Auth0 Organizations
  │        └─ General enterprise? → Auth0 Enterprise
  └─ NO → Need modern WebAuthn?
           ├─ YES → Clerk (Passwordless)
           │        ├─ Organizations needed? → Clerk Pro
           │        └─ Simple auth only? → Clerk Starter
           └─ NO → Google ecosystem integration?
                    ├─ YES → Firebase Auth
                    └─ NO → Clerk (default)
```

### Database Provider Selection
```
Need PostgreSQL with advanced features?
  ├─ YES → Real-time subscriptions required?
  │        ├─ YES → Supabase (PostgreSQL 16 + pgvector)
  │        └─ NO → Neon (Serverless PostgreSQL)
  └─ NO → Real-time collaborative features?
           ├─ YES → Convex (Real-time backend)
           └─ NO → Mobile-first app?
                    ├─ YES → Firestore (Mobile optimized)
                    └─ NO → Supabase (default)
```

### Deployment Platform Selection
```
Edge performance critical?
  ├─ YES → Vercel (Edge optimization)
  │        ├─ Next.js app? → Vercel (optimized)
  │        └─ Other framework? → Vercel (universal)
  └─ NO → Full-stack container needed?
           ├─ YES → Railway (Container optimized)
           │        ├─ Multi-region? → Railway Pro
           │        └─ Single region? → Railway Standard
           └─ NO → Vercel (default)
```

---

## Real-World Integration Examples

### Example 1: Enterprise SaaS Application
- **Stack**: Auth0 + Supabase + Vercel
- **Features**: Multi-tenant architecture, enterprise SSO, global edge performance
- **Cost**: $800-1200/month
- **Setup Time**: 2-3 days

### Example 2: Modern Web Application
- **Stack**: Clerk + Neon + Vercel
- **Features**: Passwordless auth, serverless database, edge functions
- **Cost**: $200-400/month
- **Setup Time**: 1-2 days

### Example 3: Real-time Collaborative Platform
- **Stack**: Clerk + Convex + Vercel
- **Features**: Real-time sync, database branching, collaborative editing
- **Cost**: $300-600/month
- **Setup Time**: 2-4 days

---

## Advanced Patterns & Migration

**Migration Engine**: Automated migration between any providers with data transformation and verification

**Cost Optimization**: AI-powered cost analysis and recommendations for optimal provider configurations

**Security Compliance**: Unified security framework supporting GDPR, HIPAA, and enterprise compliance requirements

**Multi-Region Deployment**: Global deployment strategies with automatic failover and data residency

*For detailed implementation patterns, migration scripts, and cost optimization examples, see:*
- [reference.md](reference.md) - Comprehensive provider documentation
- [examples.md](examples.md) - Production-ready implementation examples

---

## Works Well With

- `moai-context7-integration` - Latest BaaS provider documentation and API patterns
- `moai-domain-frontend` - Frontend integration patterns for BaaS providers
- `moai-domain-backend` - Backend architecture patterns for BaaS integration
- `moai-security-api` - BaaS security best practices and compliance
- `moai-cloud-aws-advanced` - AWS integration with BaaS providers
- `moai-foundation-trust` - Quality validation for BaaS implementations

---

**Status**: Production Ready (Enterprise)
**Generated with**: MoAI-ADK Skill Factory v2.0
**Last Updated**: 2025-11-25
**Providers Covered**: 9 major BaaS services (Auth0, Clerk, Firebase Auth, Supabase, Neon, Convex, Firestore, Vercel, Railway)
