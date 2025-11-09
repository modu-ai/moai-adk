---
spec_id: SPEC-BAAS-ECOSYSTEM-001
spec_title: BaaS 플랫폼 생태계 통합 (9개 플랫폼 Ultra-comprehensive)
created_date: 2025-11-09
version: 2.0.0
status: active
priority: P0
owner: GoosLab
related_tags:
  - "@SPEC:BAAS-ECOSYSTEM-001"
  - "@CODE:BAAS-FOUNDATION"
  - "@CODE:BAAS-SUPABASE"
  - "@CODE:BAAS-VERCEL"
  - "@CODE:BAAS-NEON"
  - "@CODE:BAAS-CLERK"
  - "@CODE:BAAS-RAILWAY"
  - "@CODE:BAAS-CONVEX"
  - "@CODE:BAAS-FIREBASE"
  - "@CODE:BAAS-CLOUDFLARE"
  - "@CODE:BAAS-AUTH0"
  - "@TEST:BAAS-PLATFORM-DETECTION"
  - "@TEST:BAAS-PATTERN-VALIDATION"
  - "@DOC:BAAS-ARCHITECTURE"
linked_specs: []
implementation_phases: 6
timeline_weeks: 6
estimated_effort_hours: 150
---

# SPEC-BAAS-ECOSYSTEM-001: BaaS 플랫폼 생태계 통합 (9개 플랫폼)

## 📋 개요

MoAI-ADK에 **9개 BaaS 플랫폼** (Supabase, Vercel, Neon, Clerk, Railway, Convex, Firebase, Cloudflare, Auth0)을 심화 통합하여 vibe coder들이 다양한 아키텍처 패턴으로 최적의 플랫폼 조합을 선택하고 설정할 수 있도록 지원합니다.

### 핵심 가치

- **자동 플랫폼 감지**: package.json, vercel.json, .env 분석으로 최대 9개 플랫폼 감지
- **확장된 패턴 추천**: 프로젝트 특성에 따른 **6-8가지 아키텍처 패턴** 제안
- **심화된 Context7 로딩**: 각 플랫폼의 1000+ word 상세 가이드 + 공식 문서
- **엔터프라이즈급 지원**: RLS 디버깅, Realtime sync, 엣지 배포, 엔터프라이즈 인증 등 실제 pain point 해결

---

## 🎯 요구사항 분석

### EARS 구조: 5가지 요구사항 타입

#### 1️⃣ Ubiquitous (항상 적용)

**Given-When-Then**: 모든 상황

1. **When** vibe coder가 새로운 프로젝트를 시작할 때 또는 기존 프로젝트에서 `/alfred:1-plan` 명령어를 실행
   - **Then** MoAI-ADK는 package.json, vercel.json, .env를 분석하여 현재 사용 중인 플랫폼을 자동 감지

2. **When** 플랫폼이 감지되면
   - **Then** 해당 플랫폼의 Context7 문서 자동 로딩 (예: Supabase → RLS, Migrations, Realtime 문서)

3. **When** 플랫폼 감지 결과를 사용자에게 제시할 때
   - **Then** 4가지 아키텍처 패턴 (A/B/C/D) 중 권장 패턴을 명확하게 제시

---

#### 2️⃣ Event-driven (특정 이벤트 발생)

**Given-When-Then**: 사용자 입력 기반

1. **When** 사용자가 `/alfred:1-plan` 명령어 실행 후 AskUserQuestion 선택지를 받음
   - **Then** 다음 8가지 옵션 제공:
     - **A**: Full Supabase (PostgreSQL + RLS + Auth + Storage + Realtime)
     - **B**: Best-of-breed DB+Auth (Neon DB + Clerk Auth + Vercel Deploy)
     - **C**: Railway All-in-one (Railway 단일 플랫폼 통합)
     - **D**: Hybrid Premium (Supabase + Clerk + Railway + Vercel + Cloudflare)
     - **E**: Firebase Full Stack (Firebase Auth + Firestore + Storage + Hosting)
     - **F**: Convex Realtime (Convex Sync + Auth + Database + Hosting)
     - **G**: Cloudflare Edge-first (Cloudflare Workers + D1 DB + Pages)
     - **H**: Enterprise OAuth (Auth0 + 자유 선택 DB/Deploy)

2. **When** 사용자가 패턴을 선택
   - **Then** 해당 패턴에 필요한 Skills와 Agents 자동으로 활성화

3. **When** 플랫폼별 설정이 필요한 경우
   - **Then** 플랫폼 전문가 Agent가 자동으로 조언 제공

---

#### 3️⃣ State-driven (상태 변화)

**Given-When-Then**: Phase 기반 배포 (6주 timeline)

1. **When** Phase 1(2주) 완료되면
   - **Then** Foundation + Supabase + Vercel Skills 활성화 ✅
   - AND backend-expert, database-expert, devops-expert Agents 강화
   - AND `/alfred:1-plan`에 플랫폼 감지 로직 통합

2. **When** Phase 2(1주) 완료되면
   - **Then** Neon + Clerk Skills 추가 활성화
   - AND database-expert, security-expert Agents 강화

3. **When** Phase 3(2주) 완료되면
   - **Then** Convex + Firebase Skills 추가 활성화 (신규)
   - AND backend-expert, frontend-expert Agents 강화

4. **When** Phase 4(2주) 완료되면
   - **Then** Cloudflare + Auth0 Skills 추가 활성화 (신규)
   - AND devops-expert, security-expert Agents 강화

5. **When** Phase 5(1주) 완료되면
   - **Then** Railway Skill 최종 추가 (이동)
   - AND 모든 9개 플랫폼 Skills 활성화

6. **When** Phase 6(1주) 완료되면
   - **Then** 모든 8가지 패턴 (A-H) 완전히 작동
   - AND 모든 아키텍처 패턴 실제 프로젝트 테스트 완료
   - AND 배포 준비 완료

---

#### 4️⃣ Optional (선택적 기능)

1. **Platform auto-combination**: 감지된 플랫폼 조합에서 최적 구성 자동 생성
2. **Extensibility for new platforms**: 새로운 BaaS 플랫폼 추가 시 Skill 확장 지점 제공
3. **Cost calculator**: 각 패턴별 월간 예상 비용 계산
4. **Migration guide**: 한 패턴에서 다른 패턴으로의 마이그레이션 가이드

---

#### 5️⃣ Unwanted Behaviors (방지할 행동)

❌ **Should NOT**:
1. Global Hooks 사용 (플랫폼 미사용 프로젝트에서도 검사)
2. 자동 파일 수정 (사용자 승인 없이)
3. 과도한 Context7 로딩 (토큰 예산 20,000 초과)
4. 사용자 학습 곡선 증가 (기존 Alfred 워크플로우 방해)

---

## 🏗️ 아키텍처 설계

### Skills 계층 구조 (1 Base + 8 Extensions)

**Phase별 활성화**:

```
Phase 1 (2주):
├── moai-baas-foundation (1000w+)
│   ├── 9개 플랫폼 개요
│   ├── 8가지 패턴 (A-H) 상세
│   └── 의사결정 행렬
├── moai-baas-supabase-ext (1000w) ✅
│   ├── Postgres, RLS, Functions, Migrations, Realtime
│   └── Context7: supabase.com/docs
└── moai-baas-vercel-ext (600w) ✅
    ├── Next.js, Edge Functions, Serverless
    └── Context7: vercel.com/docs

Phase 2 (1주):
├── moai-baas-neon-ext (1000w)
│   ├── Serverless Postgres, DB branching, pooling
│   └── Context7: neon.tech/docs
└── moai-baas-clerk-ext (1000w)
    ├── OAuth, MFA, SSO, Webhooks, session
    └── Context7: clerk.com/docs

Phase 3 (2주):
├── moai-baas-convex-ext (1000w) [신규]
│   ├── Realtime Sync, Functions, Database, Auth
│   └── Context7: convex.dev/docs
└── moai-baas-firebase-ext (1000w) [신규]
    ├── Auth, Firestore, Storage, Hosting, Functions
    └── Context7: firebase.google.com/docs

Phase 4 (2주):
├── moai-baas-cloudflare-ext (1000w) [신규]
│   ├── Workers, D1 Database, Pages, Analytics Engine
│   └── Context7: developers.cloudflare.com/docs
└── moai-baas-auth0-ext (1000w) [신규]
    ├── Enterprise Auth, SAML, MFA, Hooks, Rules
    └── Context7: auth0.com/docs

Phase 5 (1주):
└── moai-baas-railway-ext (600w)
    ├── Full-stack deployment, monitoring
    └── Context7: railway.app/docs
```

### Agents 강화 (6 Agents + 8 Patterns)

| Agent | 강화 항목 | Phase | 목적 |
|-------|---------|-------|------|
| spec-builder | Platform detection (9개) | Phase 1 | `/alfred:1-plan` 실행 시 자동 감지 |
| backend-expert | Stack recommendation (8개 패턴) | Phase 1-4 | 패턴 A-H 선택지 제공 |
| database-expert | DB selection & optimization | Phase 2-3 | Postgres vs. Neon vs. Firestore vs. Convex |
| security-expert | Auth comparison & enterprise | Phase 2-4 | Supabase vs. Clerk vs. Auth0 심화 비교 |
| devops-expert | Deployment strategy (9개) | Phase 1-5 | 각 플랫폼/패턴별 배포 전략 |
| frontend-expert | Edge/Client-side integration | Phase 1-3 | Vercel Edge, Cloudflare Workers, Convex 활용 |

### Platform Detection 알고리즘 (9개 플랫폼)

```
Input: Project root directory
├─ Step 1: package.json 분석
│  ├─ "@supabase/supabase-js" → add "supabase"
│  ├─ "@clerk/nextjs" → add "clerk"
│  ├─ "convex" → add "convex"
│  ├─ "firebase" → add "firebase"
│  ├─ "wrangler" → add "cloudflare"
│  ├─ "auth0" → add "auth0"
│  ├─ "next" → add "vercel"
│  └─ "@neondatabase/serverless" → add "neon"
├─ Step 2: Configuration files 확인
│  ├─ vercel.json 존재 → add "vercel"
│  ├─ convex.json 존재 → add "convex"
│  ├─ firebase.json 존재 → add "firebase"
│  ├─ wrangler.toml 존재 → add "cloudflare"
│  └─ .firebaserc 존재 → add "firebase"
├─ Step 3: .env 분석
│  ├─ "neon.tech" → add "neon"
│  ├─ "railway.app" → add "railway"
│  ├─ "supabase.co" → add "supabase"
│  ├─ "clerk.com" → add "clerk"
│  ├─ "convex.cloud" → add "convex"
│  ├─ "firebase" → add "firebase"
│  ├─ "cloudflare" → add "cloudflare"
│  └─ "auth0.com" → add "auth0"
└─ Output: List of detected platforms (1-9개) + recommended pattern (A-H)
```

---

## 📊 기술 스택

| 계층 | 기술 | 목적 | 크기 |
|-----|-----|------|------|
| Skills | Progressive Disclosure (1 Base + 8 Ext) | 토큰 효율성 | ~10,000w |
| Agents | 6개 Domain Expert Agents | 9개 플랫폼 전문 조언 | 확대 |
| Context7 | 9개 공식 문서 (모든 플랫폼) | 최신 정보 유지 | ~9000 tokens |
| Integration | `/alfred:1-plan` 개선 (9개 감지) | 워크플로우 통합 | 확대 |
| Detection | Python 스크립트 (3-step 분석) | 9개 플랫폼 자동 감지 | 강화 |
| Patterns | 8가지 아키텍처 패턴 (A-H) | 다양한 프로젝트 대응 | 신규 |

---

## 🎯 성공 기준

### Functional Requirements

1. ✅ Platform auto-detection (9개 플랫폼, 3-step 분석)
2. ✅ Context7 auto-loading (감지된 플랫폼 문서 9개)
3. ✅ AskUserQuestion integration (8가지 패턴 선택)
4. ✅ Agent recommendations (9개 플랫폼별 전문가 조언)
5. ✅ Pattern recommendation (감지된 플랫폼 기반 최적 패턴 제시)
6. ✅ 토큰 예산 관리 (<20,000 tokens max)

### Quality Requirements

1. ✅ No global Hooks (플랫폼 미사용 프로젝트 영향 없음)
2. ✅ No learning curve increase (기존 워크플로우 확장만)
3. ✅ Backward compatibility (모든 기존 프로젝트 호환)

---

## 📅 구현 타임라인 (6주, 150시간)

### Phase 1 (2주, 40시간) - 기초 구축 및 Postgres 계열
- **Skills**: Foundation (1000w+) ✅ + Supabase (1000w) ✅ + Vercel (600w) ✅
- **Agents**: backend-expert, database-expert, devops-expert 강화
- **Integration**: `/alfred:1-plan` 플랫폼 감지 로직 추가 (9개 플랫폼)
- **Patterns**: A (Full Supabase) 완성

### Phase 2 (1주, 20시간) - 고급 DB 및 인증
- **Skills**: Neon (1000w) + Clerk (1000w)
- **Agents**: database-expert, security-expert 강화
- **Patterns**: B (Best-of-breed) 완성

### Phase 3 (2주, 40시간) - Realtime 및 Firebase 계열 [신규]
- **Skills**: Convex (1000w) + Firebase (1000w)
- **Agents**: backend-expert, frontend-expert 강화
- **Patterns**: E (Firebase Full Stack) + F (Convex Realtime) 완성

### Phase 4 (2주, 40시간) - 엣지 컴퓨팅 및 엔터프라이즈 인증 [신규]
- **Skills**: Cloudflare (1000w) + Auth0 (1000w)
- **Agents**: devops-expert, security-expert 강화
- **Patterns**: G (Cloudflare Edge) + H (Enterprise OAuth) 완성

### Phase 5 (1주, 15시간) - Full-stack 통합
- **Skills**: Railway (600w)
- **Agents**: devops-expert 최종 강화
- **Patterns**: C (Railway All-in-one) + D (Hybrid Premium) 완성

### Phase 6 (1주, 15시간) - 테스트 및 배포 준비
- 모든 8가지 패턴 (A-H) 실제 프로젝트 테스트
- docs/troubleshooting/baas-platforms.md 작성 (9개 플랫폼)
- README.md BaaS 섹션 추가 (8가지 패턴)
- 토큰 예산 최종 검증
- 배포 준비 완료

---

## 🔗 Related Documents

- `.moai/specs/SPEC-BAAS-ECOSYSTEM-001/plan.md` - 6주 상세 구현 계획
- `.moai/specs/SPEC-BAAS-ECOSYSTEM-001/acceptance.md` - 승인 기준 (Given-When-Then, 확장)
- `CLAUDE.md` - Alfred 핵심 지침

---

## 📝 변경 이력

| 버전 | 날짜 | 변경사항 |
|-----|-----|---------|
| 2.0 | 2025-11-09 | 9개 플랫폼 Ultra-comprehensive 확장 |
| 1.0 | 2025-11-09 | 초기 생성 (5개 플랫폼) |
