---
spec_id: SPEC-04-GROUP-D
title: Phase 4 Skill Modularization - Platform/BaaS Skills (10개)
description: 10개 플랫폼 및 BaaS 서비스 스킬의 포괄적인 모듈화 및 표준화
phase: Phase 4
category: PLATFORM
week: 7-8
status: PLANNED
priority: HIGH
owner: GOOS행
created_date: 2025-11-22
updated_date: 2025-11-22
total_skills: 10
modularization_target: 100%
---

# SPEC-04-GROUP-D: Phase 4 Skill Modularization - Platform/BaaS Skills

## Given (알려진 조건)

### 완료된 작업
- **Phase 1-3**: 15개 스킬 모듈화 완료
- **표준 모듈화 패턴**: 5개 파일 구조 확립
- **Context7 Integration**: 모든 스킬에 라이브러리 문서 링크
- **기준 날짜**: 2025-11-22 (최신 버전 기준)

### 대상 스킬 목록 (10개)

#### Authentication Services (2개)
1. **moai-baas-auth0-ext** - Auth0 플랫폼, OAuth, 멀티 테넌시
2. **moai-baas-clerk-ext** - Clerk 서비스, 모던 인증, UX

#### Database Services (3개)
3. **moai-baas-neon-ext** - Neon PostgreSQL, 서버리스, 스케일링
4. **moai-baas-supabase-ext** - Supabase 플랫폼, PostgreSQL, RealtimeDB
5. **moai-baas-firebase-ext** - Firebase 에코시스템, Firestore, Realtime

#### Deployment & Edge (2개)
6. **moai-baas-vercel-ext** - Vercel 플랫폼, Next.js, 엣지 함수
7. **moai-baas-cloudflare-ext** - Cloudflare 서비스, 워커, CDN

#### Real-Time & Specialized (2개)
8. **moai-baas-convex-ext** - Convex 백엔드, RealtimeDB, TypeScript
9. **moai-baas-railway-ext** - Railway 플랫폼, 배포, 환경 관리

#### Foundation
10. **moai-baas-foundation** - BaaS 설계 원칙, 서비스 선택, 통합

---

## When (실행 조건)

### 선행 조건
- SPEC-04-GROUP-A, B, C 완료 후 진행
- 남은 토큰 예산 충분함 (≥150K)
- Skill Factory 에이전트 활용 가능
- Context7 라이브러리 접근 가능
- 각 BaaS 플랫폼 최신 문서 확인

### 실행 가능한 경우
- 서비스 카테고리별 관련 스킬을 묶어서 처리 가능
- 병렬 처리로 효율성 증대 가능

---

## What (명확한 요구사항)

### 각 스킬마다 생성할 파일 구조
- SKILL.md (≤400줄): 개요, 3단계 학습 경로, Best Practices
- examples.md (550-700줄): 10-15개 실제 예제
- modules/advanced-patterns.md (400-500줄): 고급 패턴 및 아키텍처
- modules/optimization.md (300-500줄): 성능 최적화 기법
- reference.md (30-40줄): API 문서, 링크

### 품질 기준
- 모든 파일 마크다운 (.md) 형식
- Context7 Integration 섹션 필수
- 2025-11-22 최신 버전 정보 기재
- 모든 예제 실행 가능해야 함
- 서비스 계정 설정 가이드 포함

### 처리 순서 (우선도 기반)

#### Session 1 (Week 7, 후반)
**대상**: Foundation + Auth Services
- moai-baas-foundation: BaaS 설계 원칙, 서비스 비교
- moai-baas-auth0-ext: Auth0 구현 패턴
- moai-baas-clerk-ext: Clerk 통합

**예상 토큰**: 70-90K

#### Session 2 (Week 8, 초반)
**대상**: Database Services
- moai-baas-neon-ext: Neon PostgreSQL, 서버리스
- moai-baas-supabase-ext: Supabase 플랫폼, RealtimeDB
- moai-baas-firebase-ext: Firebase 서비스, Firestore

**예상 토큰**: 80-100K

#### Session 3 (Week 8, 중반)
**대상**: Deployment & Real-Time
- moai-baas-vercel-ext: Vercel 배포, 엣지 함수
- moai-baas-cloudflare-ext: Cloudflare 워커, CDN
- moai-baas-railway-ext: Railway 배포, 환경 관리

**예상 토큰**: 80-100K

#### Session 4 (Week 8, 후반)
**대상**: Real-Time & 최종 검수
- moai-baas-convex-ext: Convex RealtimeDB
- 기존 스킬 검수 및 개선

**예상 토큰**: 40-50K

---

## Then (완료 기준)

### SPEC 완료 후 상태
- 10개 스킬 100% 모듈화 완료
- 각 스킬당 5개 파일 생성/수정
- 모든 검증 항목 완료

### 누적 진행률
```
Phase 4 Overall Progress:
- Week 1-2: 15개 (11.1%)
- Week 4-5: +18개 (GROUP-A) (13.3%)
- Week 5-6: +17개 (GROUP-B) (12.6%)
- Week 6-7: +20개 (GROUP-C) (14.8%)
- Week 7-8: +10개 (GROUP-D) (7.4%) ← 현재
- 누계: 80개 (59.3%)
- 남은 작업: 55개 (40.7%)
```

---

## 자동화 지시사항

### Skill Factory 배치 모듈화 명령어
```bash
# Session 1: Foundation + Auth
Skill("moai-cc-skill-factory", {
  "action": "batch_modularize",
  "skills": [
    "moai-baas-foundation",
    "moai-baas-auth0-ext",
    "moai-baas-clerk-ext"
  ],
  "target_version": "2025-11-22",
  "category": "authentication"
})

# Session 2: Database Services
Skill("moai-cc-skill-factory", {
  "action": "batch_modularize",
  "skills": [
    "moai-baas-neon-ext",
    "moai-baas-supabase-ext",
    "moai-baas-firebase-ext"
  ],
  "target_version": "2025-11-22",
  "category": "database"
})

# Session 3: Deployment
Skill("moai-cc-skill-factory", {
  "action": "batch_modularize",
  "skills": [
    "moai-baas-vercel-ext",
    "moai-baas-cloudflare-ext",
    "moai-baas-railway-ext"
  ],
  "target_version": "2025-11-22",
  "category": "deployment"
})
```

---

## BaaS 서비스 비교 매트릭스

| 서비스 | 특성 | 최적 사용 사례 |
|--------|------|----------------|
| Auth0 | 엔터프라이즈 인증 | 복잡한 권한, 멀티 테넌시 |
| Clerk | 개발자 친화적 | 스타트업, 모던 UX |
| Neon | 서버리스 PostgreSQL | 확장 가능한 관계형 DB |
| Supabase | 오픈소스 Firebase | PostgreSQL + RealtimeDB |
| Firebase | 구글 에코시스템 | 빠른 프로토타이핑 |
| Vercel | 프론트엔드 배포 | Next.js, 엣지 함수 |
| Cloudflare | 글로벌 엣지 | CDN, 워커, 보안 |
| Railway | 간단한 배포 | 풀스택 애플리케이션 |
| Convex | RealtimeDB | 실시간 협업 앱 |

---

## 리소스 예산

| Session | 카테고리 | 스킬 수 | 예상 토큰 |
|---------|---------|--------|----------|
| Session 1 | Auth | 3 | 70-90K |
| Session 2 | Database | 3 | 80-100K |
| Session 3 | Deployment | 3 | 80-100K |
| Session 4 | Real-time | 1 | 40-50K |
| **합계** | **BaaS** | **10개** | **270-340K** |

---

**SPEC ID**: SPEC-04-GROUP-D
**생성일**: 2025-11-22
**상태**: PLANNED
**우선도**: HIGH
