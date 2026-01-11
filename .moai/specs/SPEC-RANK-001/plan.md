# SPEC-RANK-001: Implementation Plan

## Overview

This document outlines the step-by-step implementation plan for MoAI Token Rank service.

---

## Phase 1: Infrastructure Setup (Week 1)

### 1.1 Project Initialization
- [ ] Initialize pnpm monorepo with Turborepo
- [ ] Create apps/web directory for Next.js
- [ ] Create packages/shared for shared types
- [ ] Configure TypeScript, ESLint, Prettier

### 1.2 Next.js 16 Setup
- [ ] Initialize Next.js 16 with App Router
- [ ] Install dependencies: tailwindcss, shadcn/ui
- [ ] Configure Tailwind with custom theme
- [ ] Set up shadcn/ui components

### 1.3 Database Setup (Neon)
- [ ] Create Neon project and database
- [ ] Install Drizzle ORM
- [ ] Create initial schema migration
- [ ] Set up Row Level Security policies

### 1.4 Authentication
- [ ] Install NextAuth.js v5
- [ ] Configure GitHub OAuth provider
- [ ] Create auth API routes
- [ ] Implement session management

### 1.5 Rate Limiting
- [ ] Create Upstash Redis project
- [ ] Install @upstash/ratelimit
- [ ] Configure rate limiters for each endpoint

### 1.6 Deployment
- [ ] Configure vercel.json
- [ ] Set up environment variables
- [ ] Deploy to Vercel
- [ ] Configure custom domain (rank.mo.ai.kr)

---

## Phase 2: Backend API (Week 2)

### 2.1 Core API Routes
- [ ] GET /api/leaderboard - Public leaderboard
- [ ] GET /api/users/[username] - User profile
- [ ] GET /api/stats/global - Global statistics

### 2.2 Protected API Routes
- [ ] GET /api/me - Current user info
- [ ] GET /api/me/stats - User statistics
- [ ] PATCH /api/me/settings - Update settings
- [ ] POST /api/me/regenerate-key - Regenerate API key

### 2.3 CLI API Routes (v1)
- [ ] POST /api/v1/sessions - Submit session data
- [ ] GET /api/v1/rank - Get current rank
- [ ] GET /api/v1/status - API status check

### 2.4 Security Implementation
- [ ] HMAC signature validation middleware
- [ ] Request timestamp validation (5-minute window)
- [ ] Server-side session hash recalculation
- [ ] Audit logging for security events

### 2.5 Score Calculation
- [ ] Implement composite score algorithm
- [ ] Create daily aggregation cron job
- [ ] Implement ranking update logic

---

## Phase 3: Frontend (Week 3)

### 3.1 Layout & Navigation
- [ ] Create root layout with header/footer
- [ ] Implement navigation with auth state
- [ ] Add theme toggle (dark/light)
- [ ] Create loading skeletons

### 3.2 Leaderboard Page
- [ ] Leaderboard table component
- [ ] Period filter (daily/weekly/monthly/all)
- [ ] Pagination or infinite scroll
- [ ] Current user highlight

### 3.3 User Profile Page
- [ ] Public profile display
- [ ] Token usage statistics
- [ ] Session history (if public)
- [ ] Rank badge display

### 3.4 Dashboard (Authenticated)
- [ ] Personal statistics overview
- [ ] Token usage chart (Recharts)
- [ ] API key management
- [ ] Privacy mode toggle

### 3.5 Authentication UI
- [ ] Sign in page with GitHub button
- [ ] Sign out confirmation
- [ ] Error handling

---

## Phase 4: CLI Integration (Week 4)

*Note: This phase is implemented in moai-adk repository*

### 4.1 CLI Commands (moai-adk/src/moai_adk/rank/)
- [ ] `moai-adk rank register` - OAuth registration
- [ ] `moai-adk rank status` - Current rank display
- [ ] `moai-adk rank leaderboard` - CLI leaderboard view
- [ ] `moai-adk rank enable/disable` - Toggle collection

### 4.2 Secure Storage
- [ ] Keyring integration (macOS/Linux/Windows)
- [ ] Fallback encrypted file storage
- [ ] API key management utilities

### 4.3 Session Collection
- [ ] JSONL file parser
- [ ] MoAI project detection (.moai folder)
- [ ] Token usage extraction
- [ ] HMAC request signing

### 4.4 Hook Integration
- [ ] Session end hook script
- [ ] settings.json registration
- [ ] Retry queue for failed submissions

---

## Phase 5: Polish & Launch (Week 5)

### 5.1 Security Testing
- [ ] OWASP ZAP scan
- [ ] Rate limit testing
- [ ] HMAC validation testing
- [ ] SQL injection testing

### 5.2 Performance Optimization
- [ ] Database query optimization
- [ ] Add database indexes
- [ ] Implement caching strategy
- [ ] Lighthouse performance audit

### 5.3 Documentation
- [ ] README.md
- [ ] API documentation
- [ ] CLI usage guide
- [ ] Privacy policy

### 5.4 Launch Preparation
- [ ] Final security review
- [ ] Backup strategy
- [ ] Monitoring setup
- [ ] Launch announcement

---

## Technical Decisions

### Why Next.js 16?
- Latest stable version with App Router improvements
- Server Components for optimal performance
- Built-in API routes eliminate need for separate backend
- Vercel native integration

### Why Neon over Supabase?
- Cost-effective for variable workloads (scale-to-zero)
- Pure PostgreSQL focus (no feature overhead)
- Better Vercel Edge integration
- Lower starting cost ($7.66/mo vs $25/mo)

### Why Drizzle over Prisma?
- Better performance with serverless
- TypeScript-first with full type inference
- Simpler migration workflow
- No binary dependencies

### Why pnpm + Turborepo?
- Efficient dependency management
- Shared packages across monorepo
- Incremental builds
- Cache optimization

---

## Dependencies

### Production Dependencies
```json
{
  "next": "^16.0.0",
  "react": "^19.0.0",
  "next-auth": "^5.0.0",
  "drizzle-orm": "^0.36.0",
  "@neondatabase/serverless": "^0.10.0",
  "@upstash/ratelimit": "^2.0.0",
  "@upstash/redis": "^1.34.0",
  "zod": "^3.23.0",
  "recharts": "^2.13.0"
}
```

### Dev Dependencies
```json
{
  "typescript": "^5.7.0",
  "tailwindcss": "^4.0.0",
  "drizzle-kit": "^0.28.0",
  "turbo": "^2.3.0",
  "@types/react": "^19.0.0",
  "@types/node": "^22.0.0"
}
```

---

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| API key leakage | Hash-only storage, regeneration option |
| Rate limit bypass | IP + user compound limiting |
| Data integrity | Server-side hash recalculation |
| Downtime | Offline CLI queue, graceful degradation |
| Privacy concerns | Clear data policy, privacy mode |

---

## Success Criteria

- [ ] All FR and NFR requirements met
- [ ] Security review findings addressed
- [ ] 90%+ test coverage on critical paths
- [ ] Lighthouse performance score > 90
- [ ] Documentation complete

---

**Last Updated**: 2026-01-11
**Status**: Ready for Execution
