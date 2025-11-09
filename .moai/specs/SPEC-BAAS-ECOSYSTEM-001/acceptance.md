---
doc_type: acceptance_criteria
spec_id: SPEC-BAAS-ECOSYSTEM-001
created_date: 2025-11-09
version: 1.0.0
---

# 승인 기준: SPEC-BAAS-ECOSYSTEM-001

## 📋 개요

이 문서는 SPEC-BAAS-ECOSYSTEM-001의 완료를 검증하기 위한 Given-When-Then 형식의 승인 기준을 정의합니다.

---

## ✅ Scenario 1: Supabase + Vercel 자동 감지

### Given (초기 상태)
```
새로운 Next.js 프로젝트
├─ package.json
│  ├─ "dependencies": {
│  │   "@supabase/supabase-js": "^2.x",
│  │   "next": "^14.x"
│  └─ }
├─ vercel.json (존재)
└─ .env
   ├─ NEXT_PUBLIC_SUPABASE_URL=https://xxx.supabase.co
   └─ NEXT_PUBLIC_SUPABASE_ANON_KEY=xxx
```

### When (사용자 액션)
```bash
cd my-supabase-vercel-project/
/alfred:1-plan "Add authentication feature"
```

### Then (예상 결과)

#### 1️⃣ Platform Detection
```
✅ Detected Platforms:
   - supabase (from package.json: @supabase/supabase-js)
   - vercel (from vercel.json + package.json: next)

✅ Recommended Pattern: A (Full Supabase + Vercel)
```

#### 2️⃣ Context7 Auto-Loading
```
✅ Loading Context7 documentation:
   - Supabase RLS: https://supabase.com/docs/guides/database/postgres/row-level-security
   - Supabase Auth: https://supabase.com/docs/guides/auth
   - Supabase Realtime: https://supabase.com/docs/guides/realtime
   - Vercel Deployments: https://vercel.com/docs/deployments/overview
   - Vercel Edge Functions: https://vercel.com/docs/functions/edge-functions

✅ Total tokens consumed: ~4500 (within 20,000 budget)
```

#### 3️⃣ AskUserQuestion
```
AskUserQuestion invoked with 4 options:

Pattern A: Full Supabase (Recommended)
├─ DB: Supabase PostgreSQL
├─ Auth: Supabase Auth
├─ Backend: Supabase Edge Functions
├─ Deploy: Vercel
└─ Cost: Low-Medium

Pattern B: Best-of-breed
├─ DB: Neon (Serverless Postgres)
├─ Auth: Clerk (Advanced MFA/SSO)
├─ Backend: Railway
└─ Cost: Medium

Pattern C: Railway All-in-one
├─ Platform: Railway (Full-stack)
├─ Includes: DB, Auth, Backend
└─ Cost: Low

Pattern D: Hybrid
├─ DB: Supabase
├─ Auth: Clerk
├─ Backend: Railway
└─ Cost: Medium-High

User selects: Pattern A (default)
```

#### 4️⃣ Agent Activation
```
✅ Activated Agents:
   - backend-expert (Supabase + Vercel stack recommendation)
   - database-expert (PostgreSQL + RLS guidance)
   - devops-expert (Vercel deployment strategy)

✅ Skills loaded:
   - moai-baas-foundation (global context)
   - moai-baas-supabase-ext (RLS, Migrations, Realtime)
   - moai-baas-vercel-ext (Edge Functions, Deployment)
```

#### 5️⃣ SPEC Creation
```
✅ SPEC document created with:
   - Supabase + Vercel architecture decision
   - RLS policy design
   - Vercel deployment configuration
   - Context7 docs linked
```

### 📊 Acceptance Checklist
- [ ] Platform detection: Supabase + Vercel
- [ ] Recommended pattern: A
- [ ] Context7 documentation loaded (RLS, Auth, Realtime, Deployments, Edge)
- [ ] Total tokens < 5000
- [ ] AskUserQuestion presented correctly
- [ ] Pattern A selected
- [ ] Agents activated (backend, database, devops)
- [ ] SPEC created with full context

---

## ✅ Scenario 2: Neon + Clerk + Vercel (Best-of-breed)

### Given (초기 상태)
```
새로운 Next.js 프로젝트
├─ package.json
│  ├─ "dependencies": {
│  │   "@clerk/nextjs": "^4.x",
│  │   "next": "^14.x",
│  │   "@neondatabase/serverless": "^0.x"
│  └─ }
├─ vercel.json (존재)
└─ .env
   ├─ DATABASE_URL=postgresql://user:pass@xxxl.neon.tech/db
   ├─ CLERK_SECRET_KEY=xxx
   └─ NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY=xxx
```

### When (사용자 액션)
```bash
cd my-enterprise-project/
/alfred:1-plan "Implement enterprise authentication"
```

### Then (예상 결과)

#### 1️⃣ Platform Detection
```
✅ Detected Platforms:
   - neon (from .env: DATABASE_URL contains neon.tech)
   - clerk (from package.json: @clerk/nextjs)
   - vercel (from vercel.json + package.json: next)

✅ Recommended Pattern: B (Best-of-breed)
```

#### 2️⃣ Context7 Auto-Loading
```
✅ Loading Context7 documentation:
   - Neon: DB Branching, Connection Pooling, Autoscaling
   - Clerk: OAuth, MFA, SSO, Webhooks
   - Vercel: Deployments, Edge Functions, Environment Variables

   Total docs loaded: 3 platforms
   Total tokens consumed: ~5500 (within 20,000 budget)
```

#### 3️⃣ AskUserQuestion
```
AskUserQuestion invoked:

Pattern B: Best-of-breed (Recommended)
├─ DB: Neon (Serverless Postgres with Branching)
├─ Auth: Clerk (Advanced MFA/SSO)
├─ Backend: Vercel Edge Functions
├─ Deploy: Vercel
└─ Features: DB branching, session management, webhooks

[Alternative patterns also presented]

User selects: Pattern B
```

#### 4️⃣ Agent Activation
```
✅ Activated Agents:
   - database-expert (Neon: connection pooling, branching)
   - security-expert (Clerk: MFA, SSO, session management)
   - devops-expert (Vercel: multi-environment deployment)

✅ Skills loaded:
   - moai-baas-foundation
   - moai-baas-neon-ext
   - moai-baas-clerk-ext
   - moai-baas-vercel-ext
```

#### 5️⃣ Architecture Recommendations
```
✅ backend-expert provides:
   - Neon connection pooling setup (PgBouncer)
   - Clerk session management best practices
   - Vercel environment configuration
   - Secret rotation strategy

✅ database-expert provides:
   - Neon DB branching workflow for development
   - Schema versioning strategy
   - Autoscaling threshold recommendations

✅ security-expert provides:
   - Clerk MFA enforcement
   - OAuth provider configuration (Google, GitHub)
   - Webhook signature validation
```

### 📊 Acceptance Checklist
- [ ] Platform detection: Neon + Clerk + Vercel
- [ ] Recommended pattern: B
- [ ] Context7 documentation loaded (Neon, Clerk, Vercel)
- [ ] Total tokens < 6000
- [ ] AskUserQuestion presented correctly
- [ ] Pattern B selected
- [ ] Agents activated (database, security, devops)
- [ ] Architecture recommendations provided

---

## ✅ Scenario 3: Railway All-in-one 감지

### Given (초기 상태)
```
새로운 Express.js 프로젝트
├─ package.json
│  ├─ "dependencies": {
│  │   "express": "^4.x",
│  │   "pg": "^8.x"
│  └─ }
├─ vercel.json (없음)
└─ .env
   ├─ DATABASE_URL=postgresql://user:pass@...railway.app/db
   └─ PORT=8000
```

### When (사용자 액션)
```bash
cd my-railway-project/
/alfred:1-plan "Deploy backend application"
```

### Then (예상 결과)

#### 1️⃣ Platform Detection
```
✅ Detected Platforms:
   - railway (from .env: DATABASE_URL contains railway.app)

✅ Recommended Pattern: C (Railway All-in-one)
   Rationale: Single platform detected, cost-efficient architecture
```

#### 2️⃣ Context7 Auto-Loading
```
✅ Loading Context7 documentation:
   - Railway: Full-stack deployment, environment management, monitoring
   - PostgreSQL: Basic database operations

   Total tokens consumed: ~2000 (very efficient)
```

#### 3️⃣ AskUserQuestion
```
AskUserQuestion invoked:

Pattern C: Railway All-in-one (Recommended)
├─ Platform: Railway (unified)
├─ Includes: PostgreSQL DB, Backend, Monitoring
├─ Deployment: Git push → Railway
└─ Cost: Low

[Alternative patterns also presented]

User selects: Pattern C
```

#### 4️⃣ Agent Activation
```
✅ Activated Agents:
   - devops-expert (Railway full-stack deployment)
   - backend-expert (Railway environment setup)

✅ Skills loaded:
   - moai-baas-foundation
   - moai-baas-railway-ext
```

#### 5️⃣ Deployment Configuration
```
✅ devops-expert provides:
   - Railway environment variables setup
   - PostgreSQL connection pooling
   - Logging and monitoring configuration
   - Cost tracking recommendations

✅ Deployment checklist:
   - [ ] Railway project created
   - [ ] Environment variables configured
   - [ ] PostgreSQL plugin attached
   - [ ] Health check endpoint configured
   - [ ] Monitoring alerts set
```

### 📊 Acceptance Checklist
- [ ] Platform detection: Railway
- [ ] Recommended pattern: C
- [ ] Context7 documentation loaded (Railway)
- [ ] Total tokens < 3000
- [ ] AskUserQuestion presented correctly
- [ ] Pattern C selected
- [ ] Agents activated (devops, backend)
- [ ] Deployment configuration provided

---

## ✅ Scenario 4: 새로운 프로젝트 (플랫폼 미감지)

### Given (초기 상태)
```
새로운 프로젝트
├─ package.json (기본값, BaaS 의존성 없음)
├─ vercel.json (없음)
└─ .env (비어있음)
```

### When (사용자 액션)
```bash
cd new-empty-project/
/alfred:1-plan "Setup backend infrastructure"
```

### Then (예상 결과)

#### 1️⃣ Platform Detection
```
⚠️ No platforms detected

Message: "No existing BaaS platforms detected.
          Let's choose the best architecture for your project."
```

#### 2️⃣ AskUserQuestion
```
AskUserQuestion invoked: Which architecture do you prefer?

Pattern A: Full Supabase (Integrated, fast development)
├─ Best for: MVPs, small teams
├─ Cost: Low-Medium
└─ Setup time: 15 minutes

Pattern B: Best-of-breed (Modular, scalable)
├─ Best for: Production systems, large teams
├─ Cost: Medium
└─ Setup time: 30 minutes

Pattern C: Railway All-in-one (Simple, cost-effective)
├─ Best for: Solo developers, low-budget startups
├─ Cost: Low
└─ Setup time: 10 minutes

Pattern D: Hybrid (Maximum flexibility)
├─ Best for: Complex requirements, multi-region
├─ Cost: Medium-High
└─ Setup time: 45 minutes

User selects: Pattern A (Full Supabase)
```

#### 3️⃣ Context7 Auto-Loading
```
✅ Loading Context7 documentation based on user choice:
   - Supabase: Complete stack documentation
   - Vercel: Deployment documentation

   Total tokens consumed: ~4500
```

#### 4️⃣ Project Initialization
```
✅ Project setup suggestions:
   1. Install dependencies
      npm install @supabase/supabase-js

   2. Create Supabase project
      https://supabase.com/dashboard

   3. Set environment variables in .env
      NEXT_PUBLIC_SUPABASE_URL=...
      NEXT_PUBLIC_SUPABASE_ANON_KEY=...

   4. Deploy to Vercel
      vercel deploy
```

### 📊 Acceptance Checklist
- [ ] Platform detection: None (correctly identified)
- [ ] AskUserQuestion presented all 4 patterns
- [ ] User selection captured (Pattern A)
- [ ] Context7 documentation loaded for Supabase + Vercel
- [ ] Total tokens < 5000
- [ ] Project initialization guidance provided
- [ ] Documentation links provided

---

## 🎯 Cross-Scenario Requirements

### Requirement 1: Token Budget Management
```
✅ For all scenarios:
   - Foundation Skill: ~800 tokens (always)
   - Extension Skills: ~600-1000 tokens each (as needed)
   - Context7 docs: ~1500 tokens per platform (max 4)

   Maximum case (4 platforms):
   800 + (1000 + 600 + 600 + 600) + (4 × 1500) = 8,600 tokens

   ✅ Well within 20,000 token budget (43% utilization max)
```

### Requirement 2: No Breaking Changes
```
✅ Backward compatibility:
   - Existing projects without BaaS still work
   - `/alfred:1-plan` behaves identically for non-BaaS projects
   - No global Hooks (no side effects)
   - All changes are agent-internal
```

### Requirement 3: Learning Curve
```
✅ Minimal learning curve:
   - Platform detection is automatic
   - 4 patterns are clear and simple
   - No new commands to learn
   - Extends existing `/alfred:1-plan` workflow
```

### Requirement 4: Documentation Quality
```
✅ Documentation standards:
   - Each Skill includes 5-6 major topics
   - Context7 links to official documentation
   - Examples for common use cases
   - Troubleshooting section in each Skill
```

### Requirement 5: Error Handling
```
✅ Graceful degradation:
   - If platform detection fails: Show all 4 patterns
   - If Context7 fails: Continue without docs
   - If agent fails: Fallback to generic guidance
   - All errors logged and reported
```

---

## 📊 Success Metrics

| Metric | Target | Method |
|--------|--------|--------|
| Platform Detection Accuracy | > 95% | Test with 20+ projects |
| Context7 Load Success | 100% | Verify all platform docs load |
| Token Usage | < 20,000 | Measure max case (4 platforms) |
| User Selection Time | < 2 minutes | Time from `/alfred:1-plan` to SPEC creation |
| Backward Compatibility | 100% | Test with existing projects |
| Documentation Completeness | > 90% | Coverage checklist |

---

## 🚦 Signoff

### Phase 1 Completion (Week 2)
- [ ] Scenario 1 (Supabase + Vercel): PASS
- [ ] Scenario 4 (New project): PASS
- [ ] Token budget verified: < 6,000
- [ ] Backward compatibility verified: All tests pass

### Phase 2 Completion (Week 3)
- [ ] Scenario 2 (Neon + Clerk + Vercel): PASS
- [ ] Token budget verified: < 7,000
- [ ] All 4 agents activated correctly

### Phase 3 Completion (Week 4)
- [ ] Scenario 3 (Railway): PASS
- [ ] All 4 patterns working: A, B, C, D

### Final Signoff (Week 5)
- [ ] All scenarios: PASS
- [ ] Token budget: PASS (< 8,600)
- [ ] Documentation: PASS (> 90% complete)
- [ ] Backward compatibility: PASS (100%)
- [ ] Ready for production deployment: YES

---

## 📝 테스트 환경

### Test Project A: Supabase + Vercel
```bash
cd test-projects/test-supabase-vercel/
# package.json: @supabase/supabase-js, next
# vercel.json: ✓
# .env: NEXT_PUBLIC_SUPABASE_URL, NEXT_PUBLIC_SUPABASE_ANON_KEY
```

### Test Project B: Neon + Clerk + Vercel
```bash
cd test-projects/test-neon-clerk-vercel/
# package.json: @clerk/nextjs, next, @neondatabase/serverless
# vercel.json: ✓
# .env: DATABASE_URL (neon.tech), CLERK_SECRET_KEY
```

### Test Project C: Railway
```bash
cd test-projects/test-railway/
# package.json: express, pg
# vercel.json: ✗
# .env: DATABASE_URL (railway.app)
```

### Test Project D: Empty (No BaaS)
```bash
cd test-projects/test-empty/
# package.json: (basic, no BaaS)
# vercel.json: ✗
# .env: (empty)
```

---

## 🔗 Related Documents

- **Main SPEC**: `.moai/specs/SPEC-BAAS-ECOSYSTEM-001/spec.md`
- **Implementation Plan**: `.moai/specs/SPEC-BAAS-ECOSYSTEM-001/plan.md`
