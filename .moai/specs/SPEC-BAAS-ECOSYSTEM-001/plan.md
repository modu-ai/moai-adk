---
doc_type: implementation_plan
spec_id: SPEC-BAAS-ECOSYSTEM-001
created_date: 2025-11-09
version: 1.0.0
---

# 구현 계획: SPEC-BAAS-ECOSYSTEM-001

## 📋 개요

4주 동안 5개 BaaS 플랫폼(Supabase, Vercel, Neon, Clerk, Railway) 통합을 단계적으로 진행합니다.

**총 노력**: 80시간 | **기간**: 4주 | **팀**: 6명 (Alfred + 5 specialists)

---

## Phase 1: Foundation + Supabase + Vercel (2주, 40시간)

### 🎯 목표

- Foundation Skill 생성 (모든 에이전트 기초)
- Supabase Skill 생성 (RLS, Migrations, Realtime)
- Vercel Skill 생성 (Edge Functions, Deployment)
- `/alfred:1-plan` 플랫폼 감지 로직 추가

### 📦 Deliverables

#### 1. Skills 생성 (3개)

**A. `.claude/skills/moai-baas-foundation/SKILL.md` (800 words)**

목차:
```
1. BaaS 개념 (100w)
   - Backend-as-a-Service 정의
   - 5가지 플랫폼 비교

2. 4가지 패턴 설명 (400w)

   Pattern A: Full Supabase (Supabase + Vercel)
   - 대상: MVP, 작은 팀
   - 장점: 통합성, 빠른 개발
   - 단점: Postgres 제약

   Pattern B: Best-of-breed (Neon + Clerk + Vercel)
   - 대상: Production, 큰 팀
   - 장점: 각 영역 최고의 도구
   - 단점: 통합 복잡도

   Pattern C: Railway (Railway all-in-one)
   - 대상: MVP, 저예산
   - 장점: 단순성, 저비용
   - 단점: 제한된 기능

   Pattern D: Hybrid (Supabase + Clerk + Railway + Vercel)
   - 대상: Production, 유연성 중시
   - 장점: 최고의 유연성
   - 단점: 관리 복잡도

3. 의사결정 행렬 (200w)
   - 프로젝트 특성별 패턴 선택
   - 예산, 팀 규모, 성숙도 기준

4. Common Pain Points (100w)
   - RLS 디버깅 팁
   - 마이그레이션 안전성
   - 성능 최적화
```

**B. `.claude/skills/moai-baas-supabase-ext/SKILL.md` (1000 words)**

목차:
```
1. Supabase 아키텍처 (150w)
   - PostgreSQL + RLS + Auth + Storage + Realtime
   - Edge Functions vs. Database Functions

2. RLS (Row Level Security) 깊이 있게 (300w)
   - Policy 작성 방법
   - 500 에러 디버깅
   - Policy 테스트 (pgTAP)
   - 보안 Best Practices

3. Database Functions (200w)
   - PostgreSQL 함수 작성
   - RPC 호출
   - 트리거 및 알림

4. Migrations (200w)
   - 버전 관리 전략
   - 마이그레이션 안전성
   - Rollback 전략

5. Realtime (100w)
   - Broadcast vs. Postgres Changes
   - Presence System

6. Common Issues & Solutions (50w)
   - Auth 토큰 관리
   - 동시성 문제
   - 성능 튜닝
```

Context7 링크:
- https://supabase.com/docs/guides/database/postgres/row-level-security
- https://supabase.com/docs/guides/database/migrations
- https://supabase.com/docs/guides/realtime

**C. `.claude/skills/moai-baas-vercel-ext/SKILL.md` (600 words)**

목차:
```
1. Vercel 배포 (150w)
   - Next.js 최적화
   - ISR vs. SSR vs. SSG
   - Image Optimization

2. Edge Functions (200w)
   - Edge Runtime 특성
   - Supabase와의 통합
   - 성능 vs. 비용 트레이드오프

3. Environment Variables (100w)
   - 환경별 설정
   - Secrets 관리

4. Monitoring & Analytics (150w)
   - Web Vitals
   - Error Tracking
   - Performance Monitoring
```

Context7 링크:
- https://vercel.com/docs/deployments/overview
- https://vercel.com/docs/functions/edge-functions

#### 2. Agents 강화 (3개)

**A. `spec-builder.md` 수정**

추가 기능:
```python
def detect_platforms_and_recommend():
    """
    1. 프로젝트 분석
       - package.json: @supabase/supabase-js, @clerk/nextjs, next 확인
       - vercel.json: 존재 여부
       - .env: neon.tech, railway.app, supabase.co 확인

    2. 감지된 플랫폼 목록 생성

    3. Context7 자동 로딩
       - 각 플랫폼의 권장 문서 로드

    4. AskUserQuestion
       - 4가지 패턴 선택지 제시
       - 사용자 선택 수집
    """
    pass
```

수정 위치: `.claude/agents/spec-builder.md` → `/alfred:1-plan` 섹션

**B. `backend-expert.md` 수정**

추가 기능:
```python
def recommend_stack(answers):
    """
    사용자 답변 기반 패턴 추천
    - MVP vs. Production
    - Team size (small/large)
    - Budget (low/high)
    - Flexibility required (yes/no)
    """
    pass
```

수정 위치: `.claude/agents/backend-expert.md` → Architecture recommendation

**C. `devops-expert.md` 수정**

추가 기능:
```python
def deployment_strategy(platform_stack):
    """
    각 플랫폼별 배포 전략
    - Supabase + Vercel
    - Neon + Railway + Vercel
    - etc.
    """
    pass
```

수정 위치: `.claude/agents/devops-expert.md` → Deployment section

#### 3. Integration 작업

**A. `/alfred:1-plan` 개선**

변경 사항:
```bash
# 기존
/alfred:1-plan "feature name"

# 변경 후 (플랫폼 감지 추가)
/alfred:1-plan "feature name"
  ↓ (자동)
  Platform Detection
  ├─ package.json 분석
  ├─ vercel.json 확인
  ├─ .env 파싱
  └─ 감지된 플랫폼 목록
  ↓ (자동)
  Context7 로딩
  ├─ Supabase docs (if detected)
  ├─ Vercel docs (if detected)
  └─ ...
  ↓ (사용자 선택)
  AskUserQuestion: 4가지 패턴 선택
  ├─ Pattern A (Full Supabase)
  ├─ Pattern B (Best-of-breed)
  ├─ Pattern C (Railway)
  └─ Pattern D (Hybrid)
  ↓ (자동)
  Agent Activation
  └─ 선택된 패턴에 필요한 Agents만 활성화
```

### ✅ Phase 1 성공 기준

1. ✅ 3개 Skills 생성 (Foundation 800w + Supabase 1000w + Vercel 600w)
2. ✅ 3개 Agents 강화 (spec-builder, backend-expert, devops-expert)
3. ✅ `/alfred:1-plan` 플랫폼 감지 로직 통합
4. ✅ Context7 Supabase + Vercel 자동 로딩
5. ✅ AskUserQuestion 패턴 선택 UI
6. ✅ 실제 프로젝트 테스트 (Supabase + Vercel)

### 🧪 테스트 계획

**Test Case 1: Supabase + Vercel 감지**
```bash
cd test-project-supabase-vercel/
# package.json: @supabase/supabase-js, next
# vercel.json: 존재
# .env: supabase.co

/alfred:1-plan "Add auth feature"
# Expected: Pattern A 추천
```

**Test Case 2: 새로운 프로젝트 (플랫폼 없음)**
```bash
cd test-project-new/
# package.json: 기본
# vercel.json: 없음
# .env: 비어있음

/alfred:1-plan "Setup backend"
# Expected: 4가지 패턴 모두 제시
```

---

## Phase 2: Neon + Clerk (1주, 20시간)

### 🎯 목표

- Neon Skill 생성 (DB branching, autoscaling)
- Clerk Skill 생성 (MFA, SSO, Webhooks)
- Agents 강화

### 📦 Deliverables

#### 1. Skills 생성 (2개)

**A. `.claude/skills/moai-baas-neon-ext/SKILL.md` (600 words)**

Topics:
- Serverless Postgres
- DB branching workflow
- Connection pooling
- Autoscaling
- Cost optimization

**B. `.claude/skills/moai-baas-clerk-ext/SKILL.md` (600 words)**

Topics:
- OAuth & SSO integration
- Multi-factor authentication (MFA)
- Session management
- Webhooks & events
- MAU optimization

#### 2. Agents 강화 (2개)

- `database-expert.md`: Neon 특화 최적화
- `security-expert.md`: Clerk auth comparison

### ✅ Phase 2 성공 기준

1. ✅ 2개 Skills 생성
2. ✅ 2개 Agents 강화
3. ✅ Pattern B (Best-of-breed) 완전 작동 테스트

### 🧪 테스트 계획

**Test Case: Neon + Clerk + Vercel 감지**
```bash
cd test-project-best-of-breed/
# package.json: @clerk/nextjs, next
# vercel.json: 존재
# .env: neon.tech

/alfred:1-plan "Add authentication"
# Expected: Pattern B 추천 + Neon docs + Clerk docs 로드
```

---

## Phase 3: Railway (1주, 10시간)

### 🎯 목표

- Railway Skill 생성
- Agent 강화
- Pattern C 테스트

### 📦 Deliverables

#### 1. Skills 생성 (1개)

**A. `.claude/skills/moai-baas-railway-ext/SKILL.md` (600 words)**

Topics:
- Railway 플랫폼 개요
- Full-stack deployment
- Environment management
- Monitoring & logging
- Cost tracking

#### 2. Agents 강화 (1개)

- `devops-expert.md`: Railway 배포 전략

### ✅ Phase 3 성공 기준

1. ✅ 1개 Skill 생성
2. ✅ 1개 Agent 강화
3. ✅ Pattern C (Railway) 완전 작동 테스트

### 🧪 테스트 계획

**Test Case: Railway all-in-one**
```bash
cd test-project-railway/
# package.json: next
# vercel.json: 없음
# .env: railway.app

/alfred:1-plan "Deploy application"
# Expected: Pattern C 추천 + Railway docs 로드
```

---

## Phase 4: Testing & Documentation (1주, 10시간)

### 🎯 목표

- 모든 4가지 패턴 (A/B/C/D) 실제 프로젝트 검증
- 문서 작성
- 토큰 예산 검증

### 📦 Deliverables

#### 1. 문서 작성 (2개)

**A. `docs/troubleshooting/baas-platforms.md`**

구조:
```
1. Supabase Troubleshooting
   - RLS policy errors
   - Auth token issues
   - Real-time connection

2. Vercel Troubleshooting
   - Edge Function errors
   - Environment variable issues
   - Build optimization

3. Neon Troubleshooting
   - Connection pooling
   - Autoscaling issues
   - Data branching

4. Clerk Troubleshooting
   - SSO configuration
   - Session management
   - Webhook delivery

5. Railway Troubleshooting
   - Environment variables
   - Logging
   - Cost monitoring
```

**B. `README.md` 수정 (BaaS 섹션 추가)**

추가 내용:
```markdown
## BaaS Platform Support

MoAI-ADK supports 5 BaaS platforms integrated into `/alfred:1-plan`:

### Supported Patterns

- **Pattern A**: Full Supabase (Supabase + Vercel)
- **Pattern B**: Best-of-breed (Neon + Clerk + Vercel)
- **Pattern C**: Railway all-in-one
- **Pattern D**: Hybrid (Supabase + Clerk + Railway + Vercel)

### Quick Start

```bash
/alfred:1-plan "Setup backend"
# MoAI-ADK will auto-detect your platforms
# and recommend the best pattern
```

See [BaaS Platforms Guide](docs/troubleshooting/baas-platforms.md)
```

#### 2. 토큰 예산 검증

**검증 항목**:
- Foundation Skill 로드: ~800 tokens
- Extension Skills 로드 (최악의 경우 5개): ~3500 tokens
- Context7 docs (최대 4개 플랫폼): ~6000 tokens
- **총합**: ~10,300 tokens (20,000 한계 내)

#### 3. 실제 프로젝트 검증

**Pattern A 테스트** (Supabase + Vercel):
- [ ] 프로젝트 생성
- [ ] 플랫폼 자동 감지
- [ ] Context7 문서 로드
- [ ] 아키텍처 패턴 추천
- [ ] 실제 기능 구현 (auth, DB, realtime)

**Pattern B 테스트** (Neon + Clerk + Vercel):
- [ ] 프로젝트 생성
- [ ] 플랫폼 자동 감지
- [ ] 3개 플랫폼 문서 동시 로드
- [ ] 아키텍처 패턴 추천
- [ ] 실제 기능 구현 (auth, DB, deployment)

**Pattern C 테스트** (Railway):
- [ ] 프로젝트 생성
- [ ] 플랫폼 자동 감지
- [ ] Railway 문서 로드
- [ ] Full-stack deployment
- [ ] 비용 추적

**Pattern D 테스트** (Hybrid):
- [ ] 프로젝트 생성
- [ ] 4개 플랫폼 모두 감지
- [ ] 아키텍처 권장사항
- [ ] 통합 테스트

### ✅ Phase 4 성공 기준

1. ✅ 모든 4가지 패턴 실제 프로젝트 테스트 완료
2. ✅ docs/troubleshooting/baas-platforms.md 작성
3. ✅ README.md BaaS 섹션 추가
4. ✅ 토큰 예산 < 20,000 확인
5. ✅ 배포 준비 완료

---

## 📊 통합 요약

### Skills 총 6개
| Skill | 크기 | 활성화 시점 |
|-------|------|----------|
| moai-baas-foundation | 800w | Phase 1 |
| moai-baas-supabase-ext | 1000w | Phase 1 |
| moai-baas-vercel-ext | 600w | Phase 1 |
| moai-baas-neon-ext | 600w | Phase 2 |
| moai-baas-clerk-ext | 600w | Phase 2 |
| moai-baas-railway-ext | 600w | Phase 3 |
| **Total** | **4200w** | Phase 4 |

### Agents 강화 (6개)
| Agent | 강화 사항 | Phase |
|-------|---------|-------|
| spec-builder | Platform detection | Phase 1 |
| backend-expert | Stack recommendation | Phase 1 |
| devops-expert | Deployment strategy | Phase 1-3 |
| database-expert | DB selection | Phase 1-2 |
| security-expert | Auth comparison | Phase 2 |
| frontend-expert | Edge Functions | Phase 1 |

### Context7 통합 (5개)
| Platform | Docs | Phase |
|----------|------|-------|
| Supabase | RLS, Migrations, Realtime | Phase 1 |
| Vercel | Deployments, Edge Functions | Phase 1 |
| Neon | Branching, Autoscaling | Phase 2 |
| Clerk | OAuth, MFA, Webhooks | Phase 2 |
| Railway | Full-stack, Monitoring | Phase 3 |

---

## 🎯 Risk Management

### Risk 1: Token Budget Overflow
**Mitigation**: Progressive Disclosure 구현
- Foundation 로드 필수
- Extension은 감지된 플랫폼만 로드
- 최대값 테스트 (4개 플랫폼 동시)

### Risk 2: Learning Curve
**Mitigation**: 자동화 우선
- 플랫폼 자동 감지
- 기존 `/alfred:1-plan` 확장 (새 명령어 없음)
- AskUserQuestion으로 직관적 선택

### Risk 3: Compatibility with Existing Projects
**Mitigation**: Backward compatibility 검증
- 플랫폼 미감지 프로젝트에도 작동
- Hooks 미사용 (Agent 내부 검증)

---

## 📅 주간 체크포인트

### Week 1-2 (Phase 1)
- [ ] Day 1-2: Skills 구조 설계
- [ ] Day 3-4: Foundation Skill 작성
- [ ] Day 5: Supabase Skill 작성
- [ ] Day 6: Vercel Skill 작성
- [ ] Day 7-8: Agents 강화
- [ ] Day 9-10: `/alfred:1-plan` 통합
- [ ] Day 11-14: 테스트 및 버그 수정

### Week 3 (Phase 2)
- [ ] Day 1-2: Neon Skill 작성
- [ ] Day 3-4: Clerk Skill 작성
- [ ] Day 5: Agents 강화
- [ ] Day 6-7: 테스트 및 버그 수정

### Week 4 (Phase 3)
- [ ] Day 1-2: Railway Skill 작성
- [ ] Day 3: Agents 강화
- [ ] Day 4-5: 테스트

### Week 5 (Phase 4)
- [ ] Day 1-2: 문서 작성
- [ ] Day 3-4: 최종 테스트
- [ ] Day 5: 배포 준비

---

## 🔗 Related Resources

- **SPEC**: `.moai/specs/SPEC-BAAS-ECOSYSTEM-001/spec.md`
- **Acceptance**: `.moai/specs/SPEC-BAAS-ECOSYSTEM-001/acceptance.md`
- **Documentation**: TBD in Phase 4
