---
id: NEXTRA-SITE-001
version: 0.0.1
status: draft
created: 2025-10-31
updated: 2025-10-31
author: @GOOS
priority: critical
---

# SPEC-NEXTRA-SITE-001 Implementation Plan

## HISTORY

### v0.0.1 (2025-10-31)
- **INITIAL**: Nextra 4.0 기본 구조 구축 실행 계획 작성
- **AUTHOR**: @GOOS
- **SECTIONS**: Milestones, Technical Approach, Architecture, Risks

---

## Milestones (우선순위 기반)

### Primary Goal: 프로젝트 초기화 및 로컬 개발 환경 구성
**목표**: Next.js 14 + Nextra 4.0 프로젝트를 초기화하고 로컬에서 개발 서버가 정상 동작하는 것을 확인

**완료 조건**:
- ✅ `package.json` 생성 및 의존성 설치 완료
- ✅ `next.config.js`, `theme.config.tsx` 파일 생성 완료
- ✅ `npm run dev` 실행 시 localhost:3000에서 Nextra 홈페이지 표시
- ✅ 파일 수정 시 HMR 동작 확인

**주요 작업**:
1. 프로젝트 디렉토리 생성 (`moai-adk-docs/`)
2. `npm init` 및 의존성 설치 (Next.js, Nextra, React)
3. TypeScript 설정 파일 생성 (`tsconfig.json`)
4. `.gitignore` 파일 생성 (`.next/`, `out/`, `node_modules/` 제외)
5. 기본 페이지 생성 (`app/layout.tsx`, `app/page.tsx`)
6. Nextra 테마 설정 (`theme.config.tsx`)
7. 개발 서버 시작 및 동작 확인

---

### Secondary Goal: 프로덕션 빌드 및 Vercel 배포 환경 구성
**목표**: 프로덕션 빌드가 성공하고 Vercel에 배포하여 adk.mo.ai.kr에서 사이트가 정상 표시되는 것을 확인

**완료 조건**:
- ✅ `npm run build` 실행 시 오류 없이 빌드 완료
- ✅ `out/` 디렉토리에 정적 파일 생성 확인
- ✅ Vercel 프로젝트 생성 및 Git 연동 완료
- ✅ `vercel.json` 설정 파일 생성 완료
- ✅ https://adk.mo.ai.kr에서 사이트 정상 표시

**주요 작업**:
1. `next.config.js`에 `output: 'export'` 설정 추가
2. `vercel.json` 생성 (배포 설정, 보안 헤더)
3. Git 저장소 생성 및 초기 커밋
4. Vercel 대시보드에서 프로젝트 생성 및 Git 저장소 연동
5. 도메인 설정 (adk.mo.ai.kr → Vercel 프로젝트)
6. 배포 및 동작 확인

---

### Final Goal: 품질 게이트 통과 및 문서 작성
**목표**: 빌드 성능, 접근성, 보안 기준을 충족하고 개발 가이드를 작성

**완료 조건**:
- ✅ 빌드 시간 < 3분
- ✅ Lighthouse 성능 점수 90+ (모바일/데스크톱)
- ✅ 보안 헤더 응답 확인 (X-Content-Type-Options, X-Frame-Options 등)
- ✅ 개발 가이드 문서 작성 완료 (`DOC:SITE-GUIDE-001`)
- ✅ 배포 가이드 문서 작성 완료 (`DOC:SITE-DEPLOY-001`)

**주요 작업**:
1. 빌드 시간 측정 및 최적화 (필요시)
2. Lighthouse 성능 테스트 실행
3. 보안 헤더 설정 검증 (curl 명령 또는 브라우저 개발자 도구)
4. 개발 가이드 작성 (로컬 개발 환경 설정, HMR 동작 확인)
5. 배포 가이드 작성 (Vercel 설정, 도메인 연결)

---

## Technical Approach (기술 접근 방식)

### 1. Next.js 14 App Router 선택 이유
- **최신 권장 방식**: Next.js 13.4+에서 App Router가 stable로 전환됨
- **React Server Components**: 서버 컴포넌트를 통한 성능 최적화
- **향후 확장성**: 다국어 지원(i18n), 동적 라우팅, API Routes 통합 시 유리

### 2. Nextra 4.0 선택 이유
- **문서 전용 프레임워크**: 기술 문서 작성에 최적화된 테마 및 플러그인 제공
- **마크다운 기반**: 기존 MkDocs 콘텐츠 마이그레이션 용이
- **검색 기능**: 내장 검색(FlexSearch) 및 코드 블록 검색 지원
- **LaTeX 지원**: 수학 공식 표현 가능 (향후 필요시)

### 3. 정적 사이트 생성(SSG) 방식
- **이유**: 모든 페이지가 문서이므로 서버 렌더링(SSR) 불필요
- **장점**:
  - 빠른 초기 로딩 속도
  - CDN 캐싱 최적화
  - Vercel Edge Network를 통한 글로벌 배포
- **설정**: `next.config.js`에 `output: 'export'` 추가

### 4. Vercel 배포 선택 이유
- **Next.js 최적화**: Vercel이 Next.js를 개발한 회사로 최적화 지원
- **자동 배포**: Git 푸시 시 자동 배포 (CI/CD 별도 설정 불필요)
- **Edge Network**: 글로벌 CDN을 통한 빠른 콘텐츠 제공
- **무료 플랜**: Hobby 플랜으로 충분 (상업적 사용 시 Pro 플랜 전환)

---

## Architecture Design (아키텍처 설계)

### High-Level Architecture
```
┌─────────────────────────────────────────────────────────────┐
│                     User (Browser)                          │
└──────────────────────┬──────────────────────────────────────┘
                       │ HTTPS Request
                       ↓
┌─────────────────────────────────────────────────────────────┐
│                  Vercel Edge Network                        │
│  (CDN, SSL, Security Headers, DDoS Protection)              │
└──────────────────────┬──────────────────────────────────────┘
                       │ Cache Miss
                       ↓
┌─────────────────────────────────────────────────────────────┐
│               Static Files (out/ directory)                 │
│  - HTML: Pre-rendered pages                                 │
│  - CSS: Tailwind + Nextra theme styles                      │
│  - JS: React 18 + Next.js 14 runtime                        │
│  - Assets: Images, fonts, favicon                           │
└─────────────────────────────────────────────────────────────┘
```

### Component Architecture (Nextra 테마 구조)
```
app/layout.tsx (Root Layout)
    ↓
┌────────────────────────────────────────┐
│  Nextra Theme (nextra-theme-docs)     │
│  ┌──────────────────────────────────┐ │
│  │  Navbar (로고, 검색, GitHub 링크) │ │
│  └──────────────────────────────────┘ │
│  ┌──────────────────────────────────┐ │
│  │  Sidebar (페이지 네비게이션)      │ │
│  └──────────────────────────────────┘ │
│  ┌──────────────────────────────────┐ │
│  │  Main Content Area               │ │
│  │  - MDX 렌더링                    │ │
│  │  - 코드 블록 (Syntax Highlight)  │ │
│  │  - 목차 (TOC)                    │ │
│  └──────────────────────────────────┘ │
│  ┌──────────────────────────────────┐ │
│  │  Footer (저작권, 소셜 링크)       │ │
│  └──────────────────────────────────┘ │
└────────────────────────────────────────┘
```

### Directory Structure Rationale
```
moai-adk-docs/
├── app/                    # Next.js 14 App Router
│   ├── layout.tsx         # Root layout (Nextra 테마 래퍼)
│   └── page.tsx           # Home page (MDX import 가능)
├── pages/                 # Nextra 문서 페이지 (옵션)
│   └── _meta.json         # 페이지 메타데이터 (순서, 제목)
├── public/                # 정적 자산
│   ├── favicon.ico
│   └── logo.png
├── theme.config.tsx       # Nextra 테마 설정 (로고, 링크, SEO)
├── next.config.js         # Next.js 설정 (Nextra 플러그인)
├── vercel.json            # Vercel 배포 설정
├── package.json           # 의존성 관리
└── tsconfig.json          # TypeScript 설정
```

**디렉토리 선택 기준**:
- `app/`: Next.js 14 App Router 사용 (Root Layout, 전역 스타일)
- `pages/`: Nextra가 자동으로 라우팅 생성 (파일 기반 라우팅)
- `public/`: 정적 자산 (이미지, 폰트) → `/` 경로로 접근 가능

---

## Dependency Management (의존성 관리)

### Core Dependencies
| Package | Version | Purpose |
|---------|---------|---------|
| `next` | `^14.2.0` | Next.js 14 (App Router) |
| `nextra` | `^4.0.0` | Nextra 코어 |
| `nextra-theme-docs` | `^4.0.0` | Nextra 문서 테마 |
| `react` | `^18.2.0` | React 18 |
| `react-dom` | `^18.2.0` | React DOM |

### Dev Dependencies
| Package | Version | Purpose |
|---------|---------|---------|
| `typescript` | `^5.3.0` | TypeScript 컴파일러 |
| `@types/node` | `^20.0.0` | Node.js 타입 정의 |
| `@types/react` | `^18.2.0` | React 타입 정의 |
| `@types/react-dom` | `^18.2.0` | React DOM 타입 정의 |

**버전 고정 전략**:
- `^` 접두사 사용: Minor 버전 업데이트 자동 허용 (호환성 유지)
- Major 버전 업데이트는 수동 검토 후 적용

---

## Risks & Mitigation (위험 요소 및 대응 방안)

### Risk 1: Nextra 4.0이 아직 베타/알파 버전일 경우
- **확률**: Medium
- **영향**: High (프로젝트 전체 지연)
- **대응**:
  - Nextra 공식 GitHub 릴리스 페이지에서 안정성 확인
  - 필요시 Nextra 3.x로 롤백 (3.x는 production-ready 확인됨)
  - Nextra Discord 커뮤니티에서 베타 사용자 피드백 확인

### Risk 2: Next.js 14 App Router 학습 곡선
- **확률**: Medium
- **영향**: Medium (개발 속도 감소)
- **대응**:
  - Next.js 공식 문서 학습 (App Router 섹션 집중)
  - Nextra 예제 프로젝트 분석 (nextra.site 소스 코드)
  - 필요시 Pages Router 사용 (Nextra는 양쪽 모두 지원)

### Risk 3: Vercel 배포 실패 (빌드 오류)
- **확률**: Low
- **영향**: High (배포 불가)
- **대응**:
  - 로컬에서 `npm run build` 성공 확인 후 배포
  - Vercel 빌드 로그 실시간 모니터링
  - 오류 발생 시 이전 배포 버전으로 즉시 롤백

### Risk 4: 도메인(adk.mo.ai.kr) DNS 설정 오류
- **확률**: Medium
- **영향**: Medium (배포 완료 후 접근 불가)
- **대응**:
  - DNS 설정 사전 검증: `nslookup adk.mo.ai.kr`, `dig adk.mo.ai.kr`
  - Vercel 대시보드에서 CNAME 레코드 확인
  - 필요시 DNS 전파 대기 (최대 24시간, 일반적으로 1-2시간)

---

## Testing Strategy (테스트 전략)

### Unit Tests (Phase 1에서는 제외)
- **이유**: 정적 사이트이므로 복잡한 로직 없음
- **Phase 2 고려 사항**: 커스텀 컴포넌트 추가 시 Jest + React Testing Library 도입

### Integration Tests
- **빌드 프로세스 테스트**:
  - `npm run build` 실행 후 `out/` 디렉토리 존재 확인
  - `out/index.html` 파일 존재 확인
  - 빌드 시간 < 3분 확인

### E2E Tests (Phase 1에서는 수동)
- **로컬 개발 서버 테스트**:
  - `npm run dev` 실행 → localhost:3000 접속
  - 홈페이지 렌더링 확인
  - 파일 수정 → HMR 동작 확인 (3초 이내)

- **배포 후 프로덕션 테스트**:
  - https://adk.mo.ai.kr 접속
  - 홈페이지 렌더링 확인
  - Lighthouse 성능 테스트 (Chrome DevTools)
  - 보안 헤더 확인 (curl 또는 브라우저 개발자 도구)

---

## Performance Optimization (성능 최적화)

### Build Time Optimization
- **목표**: 빌드 시간 < 3분
- **전략**:
  - 불필요한 의존성 제거 (`package.json` 정리)
  - Next.js 빌드 캐시 활성화 (`.next/cache/`)
  - Vercel에서 자동 빌드 캐싱 활용

### Runtime Performance
- **First Load JS**: < 150 KB (Next.js 권장)
- **전략**:
  - 코드 스플리팅: Next.js 자동 지원
  - 이미지 최적화: `next/image` 사용 (Phase 2)
  - Lazy Loading: 불필요한 컴포넌트 지연 로딩

### CDN Optimization
- **Vercel Edge Network**: 자동으로 글로벌 CDN에 배포
- **캐싱 전략**:
  - 정적 파일 (HTML, CSS, JS): 브라우저 캐시 1년
  - 이미지: 브라우저 캐시 1년

---

## Rollback Strategy (롤백 전략)

### Scenario 1: Vercel 배포 실패
- **대응**:
  1. Vercel 대시보드에서 "Rollback to Previous Deployment" 클릭
  2. 이전 배포 버전으로 즉시 복구 (1분 이내)
  3. 오류 원인 분석 후 수정

### Scenario 2: 빌드 성공했으나 사이트 동작 안 함
- **대응**:
  1. 로컬에서 `npm run build && npm run start` 재현
  2. 브라우저 콘솔 및 네트워크 탭에서 오류 확인
  3. 문제 수정 후 재배포

### Scenario 3: DNS 설정 오류
- **대응**:
  1. Vercel 기본 도메인 (예: `moai-adk-docs.vercel.app`)으로 임시 접근
  2. DNS 설정 재검토 (CNAME 레코드 올바른지 확인)
  3. DNS 전파 대기 (최대 24시간)

---

## Documentation Plan (문서 작성 계획)

### 1. 개발 가이드 (`DOC:SITE-GUIDE-001`)
**내용**:
- 로컬 개발 환경 설정 (Node.js, npm 설치)
- 프로젝트 클론 및 의존성 설치
- 개발 서버 시작 방법 (`npm run dev`)
- HMR 동작 확인 방법
- 빌드 프로세스 설명 (`npm run build`)

### 2. 배포 가이드 (`DOC:SITE-DEPLOY-001`)
**내용**:
- Vercel 프로젝트 생성 방법
- Git 저장소 연동 방법
- 도메인 설정 방법 (adk.mo.ai.kr)
- 환경 변수 설정 (필요시)
- 배포 상태 모니터링 방법

### 3. 트러블슈팅 가이드 (`@DOC:SITE-TROUBLESHOOTING-001`)
**내용**:
- 빌드 오류 해결 방법
- 배포 실패 대응 방법
- DNS 설정 오류 해결 방법
- 성능 최적화 팁

---

## Success Criteria (성공 기준)

### Milestone 1 완료 기준
- ✅ `npm run dev` 실행 시 localhost:3000에서 Nextra 홈페이지 표시
- ✅ 파일 수정 시 HMR 동작 (3초 이내)

### Milestone 2 완료 기준
- ✅ `npm run build` 실행 시 오류 없이 빌드 완료
- ✅ https://adk.mo.ai.kr에서 사이트 정상 표시

### Milestone 3 완료 기준
- ✅ 빌드 시간 < 3분
- ✅ Lighthouse 성능 점수 90+
- ✅ 보안 헤더 응답 확인

### 전체 SPEC 완료 기준
- ✅ 모든 Milestone 완료
- ✅ 개발 가이드 및 배포 가이드 작성 완료
- ✅ Git 커밋 및 PR 생성 완료

---

## Next Steps (다음 단계)

### 의존 관계
- **선행 작업**: 없음 (첫 번째 SPEC)
- **후속 작업**:
  - **`SPEC:NEXTRA-I18N-001`**: 다국어 지원 (NEXTRA-SITE-001 완료 후)
  - **`SPEC:NEXTRA-CONTENT-001`**: 콘텐츠 마이그레이션 (NEXTRA-SITE-001 완료 후)

### Handoff Points
- **구현 담당**: TDD-implementer (via `/alfred:2-run SPEC-NEXTRA-SITE-001`)
- **문서 작성**: doc-syncer (via `/alfred:3-sync`)
- **Git 관리**: git-manager (자동)

---

## Appendix

### Useful Commands
```bash
# 프로젝트 초기화
npm init -y
npm install next@^14.2.0 nextra@^4.0.0 nextra-theme-docs@^4.0.0 react@^18.2.0 react-dom@^18.2.0

# 개발 서버 시작
npm run dev

# 프로덕션 빌드
npm run build

# 빌드 결과 미리보기
npm run start

# 빌드 시간 측정
time npm run build

# Lighthouse 성능 테스트 (Chrome DevTools 사용)
# 브라우저에서 F12 → Lighthouse → Generate Report

# 보안 헤더 확인
curl -I https://adk.mo.ai.kr

# DNS 설정 확인
nslookup adk.mo.ai.kr
dig adk.mo.ai.kr
```

### Reference Links
- [Next.js 14 App Router 가이드](https://nextjs.org/docs/app)
- [Nextra 4.0 문서](https://nextra.site)
- [Vercel 배포 가이드](https://vercel.com/docs/deployments/overview)
- [React Server Components 설명](https://nextjs.org/docs/app/building-your-application/rendering/server-components)
