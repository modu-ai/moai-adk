---
spec_id: SPEC-BAAS-ECOSYSTEM-001
spec_title: BaaS 플랫폼 생태계 통합 (5개 플랫폼)
created_date: 2025-11-09
version: 1.0.0
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
  - "@TEST:BAAS-PLATFORM-DETECTION"
  - "@TEST:BAAS-PATTERN-VALIDATION"
  - "@DOC:BAAS-ARCHITECTURE"
linked_specs: []
implementation_phases: 4
timeline_weeks: 4
estimated_effort_hours: 80
---

# SPEC-BAAS-ECOSYSTEM-001: BaaS 플랫폼 생태계 통합

## 📋 개요

MoAI-ADK에 5개 BaaS 플랫폼(Supabase, Vercel, Neon, Clerk, Railway)을 통합하여 vibe coder들이 프로젝트에 최적의 플랫폼 조합을 자동으로 선택하고 설정할 수 있도록 지원합니다.

### 핵심 가치

- **자동 플랫폼 감지**: package.json, vercel.json, .env 분석으로 현재 사용 중인 플랫폼 자동 감지
- **최적 패턴 추천**: 프로젝트 특성(MVP/Production, 팀 규모, 예산)에 따른 4가지 표준 패턴 제안
- **Context7 자동 로딩**: 선택된 플랫폼의 공식 문서 자동 로딩
- **상위 문제 해결**: RLS 디버깅, 스키마 설계, 마이그레이션 안전성 등 실제 pain point 해결

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
   - **Then** 다음 옵션 제공:
     - A: Full Supabase (Supabase 모든 기능)
     - B: Best-of-breed (Neon DB + Clerk Auth + Vercel Deploy)
     - C: Railway All-in-one (Railway 단일 플랫폼)
     - D: Hybrid (Supabase + Clerk + Railway + Vercel 조합)

2. **When** 사용자가 패턴을 선택
   - **Then** 해당 패턴에 필요한 Skills와 Agents 자동으로 활성화

3. **When** 플랫폼별 설정이 필요한 경우
   - **Then** 플랫폼 전문가 Agent가 자동으로 조언 제공

---

#### 3️⃣ State-driven (상태 변화)

**Given-When-Then**: Phase 기반 배포

1. **When** Phase 1(2주) 완료되면
   - **Then** Foundation + Supabase + Vercel Skills 활성화
   - AND backend-expert, database-expert, devops-expert Agents 강화
   - AND `/alfred:1-plan`에 플랫폼 감지 로직 통합

2. **When** Phase 2(1주) 완료되면
   - **Then** Neon + Clerk Skills 추가 활성화
   - AND database-expert, security-expert Agents 강화

3. **When** Phase 3(1주) 완료되면
   - **Then** Railway Skill 활성화
   - AND devops-expert Agent 강화

4. **When** Phase 4(1주) 완료되면
   - **Then** 모든 4가지 패턴 완전히 작동
   - AND 실제 프로젝트 테스트 완료

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

### Skills 계층 구조

```
moai-baas-foundation (Foundation)
├── 800 words
├── BaaS 개념 설명
├── 4가지 패턴 (A/B/C/D) 상세 설명
└── 의사결정 행렬

moai-baas-supabase-ext (Extension)
├── 1000 words
├── Postgres, RLS, Edge Functions
├── Migrations, Realtime
└── Context7: supabase.com/docs

moai-baas-vercel-ext (Extension)
├── 600 words
├── Next.js optimization
├── Edge vs. Serverless
└── Context7: vercel.com/docs

moai-baas-neon-ext (Extension)
├── 600 words
├── DB branching, autoscaling
├── Connection pooling
└── Context7: neon.tech/docs

moai-baas-clerk-ext (Extension)
├── 600 words
├── MFA, SSO, Webhooks
├── MAU optimization
└── Context7: clerk.com/docs

moai-baas-railway-ext (Extension)
├── 600 words
├── Full-stack deployment
├── Cost optimization
└── Context7: railway.app/docs
```

### Agents 강화

| Agent | 강화 항목 | 목적 |
|-------|---------|------|
| spec-builder | Platform detection logic | `/alfred:1-plan` 실행 시 자동 감지 |
| backend-expert | Stack recommendation | 패턴 A/B/C/D 선택지 제공 |
| database-expert | Platform-specific DB selection | Postgres vs. Neon vs. Railway 선택 |
| security-expert | Auth comparison | Supabase Auth vs. Clerk 비교 |
| devops-expert | Deployment strategy | 각 패턴별 배포 전략 |
| frontend-expert | Vercel Edge Functions | Edge Functions 활용 |

### Platform Detection 알고리즘

```
Input: Project root directory
├─ Step 1: package.json 분석
│  ├─ "@supabase/supabase-js" → add "supabase"
│  ├─ "@clerk/nextjs" → add "clerk"
│  └─ "next" → add "vercel"
├─ Step 2: vercel.json 확인
│  ├─ 존재 → add "vercel"
│  └─ 미존재 → skip
├─ Step 3: .env 분석
│  ├─ "neon.tech" in content → add "neon"
│  ├─ "railway.app" in content → add "railway"
│  └─ "supabase.co" in content → add "supabase"
└─ Output: List of detected platforms + recommended pattern
```

---

## 📊 기술 스택

| 계층 | 기술 | 목적 |
|-----|-----|------|
| Skills | Progressive Disclosure (1 Base + 5 Ext) | 토큰 효율성 |
| Agents | 6개 Domain Expert Agents | 플랫폼별 전문 조언 |
| Context7 | 5개 공식 문서 (Supabase, Vercel, Neon, Clerk, Railway) | 최신 정보 유지 |
| Integration | `/alfred:1-plan` 개선 | 워크플로우 통합 |
| Detection | Python 스크립트 (package.json, .env 분석) | 자동 감지 |

---

## 🎯 성공 기준

### Functional Requirements

1. ✅ Platform auto-detection (4가지 표준 조합)
2. ✅ Context7 auto-loading (감지된 플랫폼 문서)
3. ✅ AskUserQuestion integration (4가지 패턴 선택)
4. ✅ Agent recommendations (플랫폼별 전문가 조언)
5. ✅ 토큰 예산 관리 (<20,000 tokens)

### Quality Requirements

1. ✅ No global Hooks (플랫폼 미사용 프로젝트 영향 없음)
2. ✅ No learning curve increase (기존 워크플로우 확장)
3. ✅ Backward compatibility (기존 프로젝트 호환)

---

## 📅 구현 타임라인

### Phase 1 (2주) - Foundation + Supabase + Vercel
- Skills: Foundation (800w) + Supabase (1000w) + Vercel (600w)
- Agents: backend-expert, database-expert, devops-expert 강화
- Integration: `/alfred:1-plan` 플랫폼 감지 로직 추가

### Phase 2 (1주) - Neon + Clerk
- Skills: Neon (600w) + Clerk (600w)
- Agents: database-expert, security-expert 강화

### Phase 3 (1주) - Railway
- Skills: Railway (600w)
- Agents: devops-expert 강화

### Phase 4 (1주) - Testing & Documentation
- 모든 4가지 패턴 (A/B/C/D) 실제 프로젝트 테스트
- docs/troubleshooting/baas-platforms.md 작성
- README.md BaaS 섹션 추가

---

## 🔗 Related Documents

- `.moai/specs/SPEC-BAAS-ECOSYSTEM-001/plan.md` - 상세 구현 계획
- `.moai/specs/SPEC-BAAS-ECOSYSTEM-001/acceptance.md` - 승인 기준 (Given-When-Then)
- `CLAUDE.md` - Alfred 핵심 지침

---

## 📝 변경 이력

| 버전 | 날짜 | 변경사항 |
|-----|-----|---------|
| 1.0 | 2025-11-09 | 초기 생성 |
