---
name: moai-baas-foundation
version: 2.0.0
created: 2025-10-22
updated: 2025-11-11
status: active
description: Backend-as-a-Service foundation patterns, BaaS architecture principles, and service integration guidelines. Use when designing BaaS solutions, selecting providers, or implementing backend services.
keywords: ['baas', 'backend', 'service', 'integration', 'architecture']
allowed-tools:
  - Read
  - Bash
  - WebSearch
  - WebFetch
---

# BaaS Foundation Skill

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-baas-foundation |
| **Version** | 2.0.0 (2025-11-11) |
| **Allowed tools** | Read, Bash, WebSearch, WebFetch |
| **Auto-load** | On demand when BaaS patterns detected |
| **Tier** | Foundation |

---

## What It Does

Backend-as-a-Service foundation patterns, BaaS architecture principles, and service integration guidelines.

**Key capabilities**:
- ✅ BaaS architecture best practices
- ✅ Provider selection criteria
- ✅ Integration patterns
- ✅ Security considerations
- ✅ Scalability principles

---

## When to Use

- ✅ Designing BaaS solutions
- ✅ Selecting BaaS providers
- ✅ Implementing backend services
- ✅ Planning service integrations

---

## Core BaaS Patterns

### Provider Selection Matrix
- **Auth0**: Authentication & authorization
- **Firebase**: Real-time database & storage
- **Supabase**: Open-source Firebase alternative
- **Clerk**: Modern authentication
- **Railway**: Deployment platform
- **Vercel**: Frontend deployment
- **Cloudflare**: Edge computing & security
- **Neon**: PostgreSQL serverless
- **Convex**: Real-time database

---

## Integration Architecture

### Service Composition
```
Frontend Application
    ↓
Authentication Layer (Auth0/Clerk)
    ↓
Business Logic Layer
    ↓
Data Layer (Supabase/Firebase/Convex)
    ↓
Infrastructure Layer (Vercel/Railway)
```

---

## Dependencies

- BaaS provider documentation
- Integration patterns
- Security best practices
- Performance considerations

---

## Works Well With

- `moai-baas-auth0-ext` (Authentication)
- `moai-baas-firebase-ext` (Database)
- `moai-baas-supabase-ext` (Open-source alternative)
- `moai-domain-backend` (Backend patterns)

---

## Changelog

- **v2.0.0** (2025-11-11): Complete metadata structure, provider matrix, integration patterns
- **v1.0.0** (2025-10-22): Initial BaaS foundation

---

**End of Skill** | Updated 2025-11-11
