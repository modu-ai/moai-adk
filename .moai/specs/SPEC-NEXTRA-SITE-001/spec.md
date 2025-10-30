---
id: NEXTRA-SITE-001
version: 0.0.1
status: draft
created: 2025-10-31
updated: 2025-10-31
author: @GOOS
priority: critical
category: feature
labels:
  - nextra
  - nextjs
  - infrastructure
  - deployment
scope:
  packages:
    - "/"
  files:
    - next.config.js
    - vercel.json
    - package.json
---

# SPEC-NEXTRA-SITE-001: Nextra 4.0 + Next.js 14 기본 구조 구축

## HISTORY

### v0.0.1 (2025-10-31)
- **INITIAL**: Nextra 4.0 기반 문서 사이트 초기 구조 SPEC 작성
- **AUTHOR**: @GOOS
- **SECTIONS**: Environment, Assumptions, Requirements, Specifications
- **CONTEXT**: MkDocs → Nextra 마이그레이션 Phase 1 (v1.0 Core) 첫 번째 SPEC

---

## @SPEC:NEXTRA-SITE-001 Overview

**목적**: Next.js 14 App Router 기반의 Nextra 4.0 사이트 초기 구조를 구축하고, Vercel 배포 환경을 설정합니다.

**범위**:
- Next.js 14 프로젝트 초기화 (App Router 사용)
- Nextra 4.0 설치 및 기본 설정
- Vercel 배포 환경 구성 (adk.mo.ai.kr)
- 개발 환경 구성 (Hot Module Replacement)

**제외 사항**:
- 다국어(i18n) 설정 → SPEC-NEXTRA-I18N-001
- 콘텐츠 마이그레이션 → SPEC-NEXTRA-CONTENT-001
- SEO 최적화 → Phase 2 (v1.1 Enhancement)

---

## Environment (환경 조건)

### Development Environment
- **Node.js**: v20.x LTS (v18.x 이상 지원)
- **Package Manager**: npm (v10.x) 또는 yarn (v1.22.x+)
- **OS**: macOS, Linux, Windows (WSL2 권장)

### Production Environment
- **Hosting**: Vercel (automatic deployment from Git)
- **Domain**: adk.mo.ai.kr (기존 도메인 재사용)
- **Build Time Target**: < 3분
- **Runtime**: Node.js 20.x (Vercel default)

### Required Dependencies
- `next@^14.0.0`: Next.js 14 (App Router)
- `nextra@^4.0.0`: Nextra 4.0
- `nextra-theme-docs@^4.0.0`: Nextra 문서 테마
- `react@^18.2.0`: React 18
- `react-dom@^18.2.0`: React DOM

### Optional Dependencies (Phase 1 제외)
- `next-intl`: 다국어 지원 (Phase 1에서는 제외)
- `mermaid`: 다이어그램 (Phase 2)

---

## Assumptions (가정)

### Technical Assumptions
1. **Next.js 14 App Router**: Pages Router가 아닌 App Router 사용 (최신 권장 사항)
2. **Nextra 4.0 안정성**: 2025년 10월 기준 Nextra 4.0이 production-ready 상태
3. **Vercel 무료 플랜**: Hobby 플랜으로 충분 (상업적 사용 시 Pro 플랜 전환)
4. **도메인 소유권**: adk.mo.ai.kr 도메인에 대한 DNS 설정 권한 보유

### Process Assumptions
1. **Git 기반 배포**: main 브랜치 푸시 시 Vercel 자동 배포
2. **환경 변수**: `.env.local` 파일로 로컬 환경 관리 (Vercel에서는 대시보드 설정)
3. **빌드 최적화**: Incremental Static Regeneration (ISR) 미사용 (모두 정적 페이지)

### User Assumptions
1. **개발자 숙련도**: Next.js 기본 개념 이해 (App Router, RSC)
2. **콘텐츠 작성자**: 마크다운 문법 숙지
3. **배포 권한**: Vercel 프로젝트에 대한 배포 권한 보유

---

## Requirements (요구사항)

### UBIQUITOUS (보편적 요구사항)
- **REQ-SITE-001**: 시스템은 Next.js 14 App Router 기반의 Nextra 4.0 사이트를 제공해야 한다
- **REQ-SITE-002**: 시스템은 모든 페이지를 정적 사이트 생성(SSG) 방식으로 빌드해야 한다
- **REQ-SITE-003**: 시스템은 반응형 디자인을 지원하여 모바일/태블릿/데스크톱에서 정상 표시되어야 한다

### EVENT-DRIVEN (이벤트 기반 요구사항)
- **REQ-SITE-004**: WHEN 개발자가 `npm run dev`를 실행하면, 시스템은 localhost:3000에서 개발 서버를 시작해야 한다
- **REQ-SITE-005**: WHEN 개발자가 파일을 수정하면, 시스템은 3초 이내에 Hot Module Replacement를 수행해야 한다
- **REQ-SITE-006**: WHEN 개발자가 `npm run build`를 실행하면, 시스템은 `.next/` 디렉토리에 최적화된 프로덕션 빌드를 생성해야 한다
- **REQ-SITE-007**: WHEN Vercel이 Git 푸시를 감지하면, 시스템은 자동으로 배포 프로세스를 시작해야 한다

### STATE-DRIVEN (상태 기반 요구사항)
- **REQ-SITE-008**: WHILE 사이트가 로컬 개발 모드에서 실행 중일 때, 시스템은 파일 변경 시 자동 새로고침을 지원해야 한다
- **REQ-SITE-009**: WHILE 빌드가 진행 중일 때, 시스템은 빌드 진행 상태를 콘솔에 출력해야 한다

### OPTIONAL (선택적 기능)
- **REQ-SITE-010**: WHERE 개발자가 Vercel에 배포하려면, 시스템은 `vercel.json` 설정 파일을 제공할 수 있다
- **REQ-SITE-011**: WHERE 개발자가 로컬 빌드를 미리보려면, 시스템은 `npm run start` 명령을 지원할 수 있다

### CONSTRAINTS (제약사항)
- **REQ-SITE-012**: IF 빌드 시간이 3분을 초과하면, 시스템은 경고를 표시하고 최적화를 권장해야 한다
- **REQ-SITE-013**: IF Node.js 버전이 18 미만이면, 시스템은 오류 메시지를 표시하고 설치를 중단해야 한다
- **REQ-SITE-014**: IF 필수 의존성이 누락되면, 시스템은 명확한 오류 메시지와 함께 설치 명령을 안내해야 한다

---

## Specifications (세부 설계)

### 1. Project Initialization

#### Directory Structure
```
moai-adk-docs/
├── app/                    # Next.js 14 App Router
│   ├── layout.tsx         # Root layout (Nextra 통합)
│   └── page.tsx           # Home page
├── pages/                 # Nextra 문서 페이지 (옵션)
├── public/                # 정적 자산 (이미지, 파비콘)
│   ├── favicon.ico
│   └── logo.png
├── theme.config.tsx       # Nextra 테마 설정
├── next.config.js         # Next.js 설정 (Nextra 플러그인)
├── vercel.json            # Vercel 배포 설정
├── package.json           # 의존성 관리
├── tsconfig.json          # TypeScript 설정
└── .gitignore             # Git 무시 파일
```

#### package.json Configuration
```json
{
  "name": "moai-adk-docs",
  "version": "1.0.0",
  "private": true,
  "scripts": {
    "dev": "next dev",
    "build": "next build",
    "start": "next start",
    "lint": "next lint"
  },
  "dependencies": {
    "next": "^14.2.0",
    "nextra": "^4.0.0",
    "nextra-theme-docs": "^4.0.0",
    "react": "^18.2.0",
    "react-dom": "^18.2.0"
  },
  "devDependencies": {
    "@types/node": "^20.0.0",
    "@types/react": "^18.2.0",
    "@types/react-dom": "^18.2.0",
    "typescript": "^5.3.0"
  }
}
```

### 2. Nextra Configuration

#### next.config.js
```javascript
const withNextra = require('nextra')({
  theme: 'nextra-theme-docs',
  themeConfig: './theme.config.tsx',
  latex: true,
  search: {
    codeblocks: true
  },
  defaultShowCopyCode: true
});

module.exports = withNextra({
  reactStrictMode: true,
  output: 'export', // 정적 사이트 생성
  images: {
    unoptimized: true // 정적 배포를 위한 이미지 최적화 비활성화
  }
});
```

#### theme.config.tsx (기본 설정)
```tsx
import { DocsThemeConfig } from 'nextra-theme-docs';

const config: DocsThemeConfig = {
  logo: <span>MoAI-ADK Documentation</span>,
  project: {
    link: 'https://github.com/modu-ai/moai-adk'
  },
  docsRepositoryBase: 'https://github.com/modu-ai/moai-adk/tree/main/docs',
  footer: {
    text: 'MoAI-ADK © 2025 Modu AI'
  },
  useNextSeoProps() {
    return {
      titleTemplate: '%s – MoAI-ADK'
    };
  }
};

export default config;
```

### 3. Vercel Deployment Configuration

#### vercel.json
```json
{
  "framework": "nextjs",
  "buildCommand": "npm run build",
  "devCommand": "npm run dev",
  "installCommand": "npm install",
  "outputDirectory": "out",
  "git": {
    "deploymentEnabled": {
      "main": true,
      "develop": false
    }
  },
  "github": {
    "autoAlias": true
  },
  "headers": [
    {
      "source": "/(.*)",
      "headers": [
        {
          "key": "X-Content-Type-Options",
          "value": "nosniff"
        },
        {
          "key": "X-Frame-Options",
          "value": "DENY"
        },
        {
          "key": "X-XSS-Protection",
          "value": "1; mode=block"
        }
      ]
    }
  ]
}
```

### 4. Development Workflow

#### Step 1: Project Setup
```bash
# 1. 프로젝트 디렉토리 생성
mkdir moai-adk-docs && cd moai-adk-docs

# 2. package.json 생성 및 의존성 설치
npm init -y
npm install next@^14.2.0 nextra@^4.0.0 nextra-theme-docs@^4.0.0 react@^18.2.0 react-dom@^18.2.0

# 3. TypeScript 설치 (선택사항)
npm install -D typescript @types/node @types/react @types/react-dom
```

#### Step 2: Local Development
```bash
# 개발 서버 시작
npm run dev

# 브라우저에서 http://localhost:3000 접속
# HMR 동작 확인: 파일 수정 시 자동 새로고침
```

#### Step 3: Production Build
```bash
# 프로덕션 빌드 생성
npm run build

# 빌드 결과 확인: out/ 디렉토리 생성 확인
ls -la out/

# 로컬 프리뷰 (선택사항)
npm run start
```

#### Step 4: Vercel Deployment
```bash
# Vercel CLI 설치 (선택사항)
npm install -g vercel

# Vercel 배포 (Git 푸시만으로도 자동 배포됨)
git push origin main

# Vercel 대시보드에서 배포 상태 확인
```

### 5. Quality Gates

#### Build Performance
- **빌드 시간**: < 3분 (초과 시 경고)
- **번들 크기**: First Load JS < 150 KB (Next.js 권장)
- **Lighthouse 점수**: Performance 90+ (모바일/데스크톱)

#### Accessibility
- **ARIA 속성**: 모든 인터랙티브 요소에 적절한 ARIA 레이블
- **키보드 네비게이션**: Tab/Enter 키로 모든 기능 접근 가능
- **색상 대비**: WCAG AA 기준 (최소 4.5:1)

#### Security Headers (vercel.json 설정)
- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `X-XSS-Protection: 1; mode=block`

---

## Traceability (추적성)

### Related SPECs
- **@SPEC:NEXTRA-I18N-001**: 다국어 지원 (의존: NEXTRA-SITE-001 완료 후)
- **@SPEC:NEXTRA-CONTENT-001**: 콘텐츠 마이그레이션 (의존: NEXTRA-SITE-001 완료 후)

### Related Tasks
- **@TASK:SITE-SETUP-001**: Next.js 14 프로젝트 초기화
- **@TASK:SITE-NEXTRA-001**: Nextra 4.0 설정 및 테마 커스터마이징
- **@TASK:SITE-DEPLOY-001**: Vercel 배포 환경 구성

### Related Tests
- **`TEST:NEXTRA-BUILD-001`**: 빌드 프로세스 성공 여부 테스트
- **`TEST:NEXTRA-CONFIG-001`**: 개발 서버 시작 및 HMR 동작 테스트
- **`TEST:NEXTRA-THEME-001`**: Vercel 배포 성공 여부 테스트

### Related Documentation
- **`DOC:SITE-GUIDE-001`**: Nextra 4.0 개발 가이드
- **`DOC:SITE-DEPLOY-001`**: Vercel 배포 가이드

---

## Risks & Mitigation (위험 요소 및 대응 방안)

### Risk 1: Nextra 4.0 호환성 이슈
- **확률**: Medium
- **영향**: High
- **대응**:
  - Nextra 4.0 공식 문서 및 GitHub Issues 모니터링
  - 문제 발생 시 Nextra 3.x로 롤백 옵션 준비
  - 커뮤니티 포럼에서 유사 사례 검색

### Risk 2: 빌드 시간 초과 (> 3분)
- **확률**: Low
- **영향**: Medium
- **대응**:
  - Incremental Static Regeneration (ISR) 고려
  - 불필요한 의존성 제거
  - Next.js 빌드 캐시 활성화

### Risk 3: Vercel 배포 실패
- **확률**: Low
- **영향**: High
- **대응**:
  - Vercel 로그 실시간 모니터링
  - 로컬 빌드 성공 확인 후 배포
  - Rollback 전략: 이전 배포 버전으로 즉시 복구

### Risk 4: 도메인(adk.mo.ai.kr) 설정 오류
- **확률**: Medium
- **영향**: Medium
- **대응**:
  - DNS 설정 사전 검증 (nslookup, dig 명령)
  - Vercel 대시보드에서 도메인 연결 상태 확인
  - CNAME 레코드 올바른 설정 확인

---

## Success Criteria (성공 기준)

### Definition of Done
- ✅ Next.js 14 프로젝트가 초기화되고 `package.json`에 모든 의존성이 설치됨
- ✅ `npm run dev` 실행 시 localhost:3000에서 Nextra 홈페이지가 표시됨
- ✅ 파일 수정 시 3초 이내에 HMR이 동작하여 변경사항이 반영됨
- ✅ `npm run build` 실행 시 오류 없이 빌드가 완료되고 `out/` 디렉토리가 생성됨
- ✅ Vercel에 배포 후 https://adk.mo.ai.kr에서 사이트가 정상 표시됨
- ✅ 빌드 시간이 3분 이내로 완료됨
- ✅ Lighthouse 성능 점수가 90+ (모바일/데스크톱)
- ✅ 보안 헤더가 응답에 포함됨 (X-Content-Type-Options, X-Frame-Options 등)

---

## Appendix

### Reference Documents
- [Next.js 14 Documentation](https://nextjs.org/docs)
- [Nextra 4.0 Documentation](https://nextra.site)
- [Vercel Deployment Guide](https://vercel.com/docs)

### Related GitHub Issues
- TBD (이슈 생성 후 링크 추가)

### Notes
- Phase 1 (v1.0 Core)에서는 다국어 지원을 제외하고 기본 구조만 구축
- SEO 최적화 및 고급 기능은 Phase 2 (v1.1 Enhancement)에서 추가
