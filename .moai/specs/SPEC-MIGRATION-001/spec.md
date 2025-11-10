---
title: "SPEC-MIGRATION-001: Next.js 16 + Nextra 4.6.0 통합 마이그레이션"
version: "1.0.0"
date: "2025-01-10"
author: "Alfred"
status: "draft"
category: "MIGRATION"
priority: "high"
---

# @SPEC:MIGRATION-001: Next.js 16 + Nextra 4.6.0 통합 마이그레이션

## 📋 TAG BLOCK

```
@REQ:MIGRATION-001-001: Next.js 16 업그레이드 요구사항
@REQ:MIGRATION-001-002: Nextra 4.6.0 마이그레이션 요구사항
@REQ:MIGRATION-001-003: React 19 업그레이드 요구사항
@REQ:MIGRATION-001-004: Pages Router → App Router 전환 요구사항
@REQ:MIGRATION-001-005: 다국어 100+ MDX 파일 호환성 요구사항
@REQ:MIGRATION-001-006: 성능 개선 목표 (빌드 50%, LCP 25%)
@REQ:MIGRATION-001-007: FlexSearch → Pagefind 검색 엔진 전환
@REQ:MIGRATION-001-008: Turbopack 빌드 시스템 도입
@REQ:MIGRATION-001-009: Bun 패키지 매니저 도입
@REQ:MIGRATION-001-010: Biome 린터/포매터 도입

@DESIGN:MIGRATION-001-001: 마이그레이션 아키텍처 설계
@DESIGN:MIGRATION-001-002: 12단계 실행 계획
@DESIGN:MIGRATION-001-003: 롤백 전략 설계
@DESIGN:MIGRATION-001-004: 위험 평가 및 완화 계획

@TASK:MIGRATION-001-001: 의존성 업그레이드 준비
@TASK:MIGRATION-001-002: 프로젝트 구조 재설계
@TASK:MIGRATION-001-003: Nextra 설정 마이그레이션
@TASK:MIGRATION-001-004: 다국어 콘텐츠 호환성 검증
@TASK:MIGRATION-001-005: 검색 엔진 전환
@TASK:MIGRATION-001-006: 테스트 및 검증

@TEST:MIGRATION-001-001: 빌드 성능 테스트
@TEST:MIGRATION-001-002: 렌더링 성능 테스트
@TEST:MIGRATION-001-003: 다국어 기능 테스트
@TEST:MIGRATION-001-004: 검색 기능 테스트
@TEST:MIGRATION-001-005: 회귀 테스트

@FEATURE:MIGRATION-001-001: App Router 호환성
@FEATURE:MIGRATION-001-002: React 19 기능 활용
@FEATURE:MIGRATION-001-003: Turbopack 빌드 성능
@FEATURE:MIGRATION-001-004: Pagefind 검색 성능
@FEATURE:MIGRATION-001-005: 자동화된 마이그레이션 검증
@FEATURE:MIGRATION-001-006: Bun 빌드 성능 최적화
@FEATURE:MIGRATION-001-007: Biome 코드 품질 자동화
```

## 🎯 개요

MoAI-ADK 문서 사이트를 현재의 Next.js 14.2.15 + Nextra 3.3.1 (Pages Router)에서 Next.js 16 + Nextra 4.6.0 (App Router)로 통합 마이그레이션하는 프로젝트 명세서입니다. React 18.2.0에서 React 19로의 업그레이드를 포함하며, 4개 언어(ko, en, ja, zh)의 100+ MDX 파일 호환성을 보장하고 성능 목표를 달성하는 것을 목표로 합니다.

## 🏗️ EARS 구조

### Environment (환경)

#### 현재 상태 (Current State)
- **프레임워크**: Next.js 14.2.15 (Pages Router)
- **문서 프레임워크**: Nextra 3.3.1
- **React 버전**: 18.2.0
- **빌드 시스템**: SWC + Webpack
- **검색 엔진**: FlexSearch
- **콘텐츠**: 100+ MDX 파일 (4개 언어: ko, en, ja, zh)
- **호스팅**: Vercel
- **CI/CD**: Vercel 자동 배포

#### 목표 상태 (Target State)
- **프레임워크**: Next.js 16 (App Router)
- **문서 프레임워크**: Nextra 4.6.0
- **React 버전**: 19.0.0
- **빌드 시스템**: Turbopack
- **검색 엔진**: Pagefind
- **콘텐츠**: 100+ MDX 파일 (4개 언어: ko, en, ja, zh) 호환성 보장
- **호스팅**: Vercel (향상된 성능)
- **CI/CD**: Vercel 자동 배포 (향상된 빌드)

#### 기술 제약 조건
- **파이썬 프로젝트**: MoAI-ADK의 문서 사이트는 docs/ 디렉토리에 위치
- **언어 정책**: conversation_language="ko"에 따른 한국어 우선 생성
- **Git 전략**: Personal 모드, 기능 브랜치 기반 개발
- **배포**: Vercel Production 환경

### Assumptions (가정)

#### 기술적 가정
1. **Node.js 호환성**: 현재 Node.js 버전이 Next.js 16 요구사항과 호환됨
2. **의존성 호환성**: 현재 사용 중인 MDX 플러그인과 Nextra 4.6.0 호환 가능
3. **Vercel 지원**: Vercel이 Next.js 16 + App Router를 완벽히 지원함
4. **성능 개선**: Turbopack이 현재 빌드 시스템보다 50% 이상 향상된 성능 제공

#### 비즈니스 가정
1. **사용자 영향**: 마이그레이션 기간 중 서비스 중단 없음
2. **콘텐츠 무결성**: 기존 100+ MDX 파일의 내용과 형식 유지
3. **SEO 영향**: URL 구조 변경 없이 SEO 순위 유지
4. **다국어 지원**: 4개 언어 모두 동일한 수준의 기능 제공

#### 리소스 가정
1. **개발 시간**: 12단계 마이그레이션을 위한 충분한 개발 시간 확보
2. **테스트 환경**: 프로덕션과 동일한 구조의 스테이징 환경
3. **롤백 준비**: 문제 발생 시 즉각적인 롤백을 위한 백업 전략
4. **문서화**: 마이그레이션 과정과 결과의 상세한 문서화

### Requirements (요구사항)

#### 기능적 요구사항

**FR1: 프레임워크 업그레이드**
- FR1.1: Next.js 14.2.15 → 16.0.0으로 업그레이드
- FR1.2: Nextra 3.3.1 → 4.6.0으로 업그레이드
- FR1.3: React 18.2.0 → 19.0.0으로 업그레이드
- FR1.4: Pages Router → App Router 완전 전환

**FR2: 빌드 시스템 현대화**
- FR2.1: SWC + Webpack → Turbopack으로 전환
- FR2.2: 빌드 시간 50% 이상 개선
- FR2.3: 핫 리로드 성능 향상
- FR2.4: 개발 환경 최적화

**FR3: 검색 엔진 마이그레이션**
- FR3.1: FlexSearch → Pagefind로 전환
- FR3.2: 4개 언어 검색 지원 유지
- FR3.3: 검색 성능 30% 이상 개선
- FR3.4: 실시간 검색 인덱싱 기능

**FR4: 콘텐츠 호환성**
- FR4.1: 100+ MDX 파일 완전 호환
- FR4.2: 4개 언어(ko, en, ja, zh) 지원 유지
- FR4.3: 기존 플러그인과 마크다운 확장 호환
- FR4.4: 이미지와 리소스 경로 유지

**FR5: 성능 최적화**
- FR5.1: LCP (Largest Contentful Paint) 25% 개선
- FR5.2: FCP (First Contentful Paint) 30% 개선
- FR5.3: 번들 크기 최적화
- FR5.4: 자동 이미지 최적화 기능 활용

#### 비기능적 요구사항

**NFR1: 호환성**
- NFR1.1: 브라우저 호환성 유지 (Chrome, Firefox, Safari, Edge 최신 버전)
- NFR1.2: 모바일 반응형 디자인 유지
- NFR1.3: 접근성 표준(WCAG 2.1 AA) 준수
- NFR1.4: SEO 구조와 메타데이터 유지

**NFR2: 안정성**
- NFR2.1: 마이그레이션 중 서비스 중단 없음
- NFR2.2: 99.9% 가용성 보장
- NFR2.3: 에러 핸들링과 로깅 강화
- NFR2.4: 장애 복구 시간 10분 내외

**NFR3: 보안**
- NFR3.1: 새로운 보안 취약점 없음
- NFR3.2: CORS 설정 유지
- NFR3.3: CSP(Content Security Policy) 정책 유지
- NFR3.4: 의존성 보안 검증

**NFR4: 유지보수성**
- NFR4.1: 코드 가독성 향상
- NFR4.2: 개발 워크플로우 단순화
- NFR4.3: 자동화된 테스트와 배포
- NFR4.4: 상세한 문서화

### Specifications (명세)

#### SP1: 12단계 마이그레이션 계획

**Phase 1: 준비 및 분석 (1-2일)**
- SP1.1: 현재 프로젝트 구조 및 의존성 분석
- SP1.2: Next.js 16 + Nextra 4.6.0 호환성 검증
- SP1.3: 마이그레이션 계획 상세화
- SP1.4: 롤백 전략 수립

**Phase 2: 의존성 업그레이드 (1일)**
- SP2.1: package.json 의존성 버전 업데이트
- SP2.2: 호환성 문제 해결
- SP2.3: 개발 환경 설정
- SP2.4: 초기 빌드 테스트

**Phase 3: 프로젝트 구조 재설계 (2-3일)**
- SP3.1: Pages Router → App Router 구조 변환
- SP3.2: 레이아웃 컴포넌트 마이그레이션
- SP3.3: 페이지 라우팅 설정
- SP3.4: 미들웨어 설정 업데이트

**Phase 4: Nextra 설정 마이그레이션 (2일)**
- SP4.1: next.config.cjs → next.config.mjs 변환
- SP4.2: Nextra 4.6.0 설정 적용
- SP4.3: 테마 설정 업데이트
- SP4.4: 플러그인 설정 마이그레이션

**Phase 5: 다국어 콘텐츠 호환성 (2-3일)**
- SP5.1: MDX 파일 형식 검증
- SP5.2: 언어별 레이아웃 설정
- SP5.3: 다국어 라우팅 설정
- SP5.4: 콘텐츠 렌더링 테스트

**Phase 6: 검색 엔진 전환 (2일)**
- SP6.1: FlexSearch → Pagefind 전환
- SP6.2: 검색 인덱싱 설정
- SP6.3: 다국어 검색 기능 구현
- SP6.4: 검색 UI 컴포넌트 업데이트

**Phase 7: Turbopack 도입 (1-2일)**
- SP7.1: Turbopack 빌드 설정
- SP7.2: 개발 환경 최적화
- SP7.3: 프로덕션 빌드 설정
- SP7.4: 빌드 성능 테스트

**Phase 8: 성능 최적화 (2-3일)**
- SP8.1: 이미지 최적화 설정
- SP8.2: 코드 스플리팅 최적화
- SP8.3: 캐싱 전략 적용
- SP8.4: Core Web Vitals 최적화

**Phase 9: 테스트 및 검증 (3-4일)**
- SP9.1: 기능 테스트 수행
- SP9.2: 성능 테스트 수행
- SP9.3: 회귀 테스트 수행
- SP9.4: 사용자 테스트 수행

**Phase 10: 배포 준비 (1일)**
- SP10.1: 프로덕션 빌드 검증
- SP10.2: 환경 변수 설정
- SP10.3: Vercel 설정 업데이트
- SP10.4: 배포 스크립트 준비

**Phase 11: 점진적 배포 (1-2일)**
- SP11.1: 스테이징 환경 배포
- SP11.2: 최종 검증 및 테스트
- SP11.3: 프로덕션 환경 배포
- SP11.4: 모니터링 설정

**Phase 12: 안정화 및 문서화 (1-2일)**
- SP12.1: 성능 모니터링
- SP12.2: 문제 해결 및 최적화
- SP12.3: 마이그레이션 문서화
- SP12.4: 운영 가이드 업데이트

#### SP2: 기술 명세

**SP2.1: 의존성 버전 명세**
```json
{
  "dependencies": {
    "next": "^16.0.0",
    "react": "^19.0.0",
    "react-dom": "^19.0.0",
    "nextra": "^4.6.0",
    "nextra-theme-docs": "^4.6.0"
  },
  "devDependencies": {
    "@types/node": "^22.0.0",
    "@types/react": "^19.0.0",
    "@types/react-dom": "^19.0.0",
    "typescript": "^5.0.0",
    "pagefind": "^1.0.0",
    "bun-types": "^1.0.0",
    "@biomejs/biome": "^1.8.0"
  },
  "packageManager": "bun@1.1.0"
}
```

**SP2.2: Next.js 16 App Router 구조**
```
docs/
├── app/
│   ├── layout.tsx          # 루트 레이아웃
│   ├── page.tsx            # 홈페이지
│   ├── [lang]/
│   │   ├── layout.tsx      # 언어별 레이아웃
│   │   └── page.tsx        # 언어별 홈
│   │   └── [...slug]/
│   │       └── page.tsx    # 동적 라우팅
│   └── api/
│       └── search/         # 검색 API
├── components/
│   ├── search/
│   ├── theme/
│   └── layout/
├── lib/
│   ├── pagefind/
│   └── i18n/
└── src/
    └── ko/
    └── en/
    └── ja/
    └── zh/
```

**SP2.3: Nextra 4.6.0 설정 명세**
```typescript
// next.config.mjs
const withNextra = require('nextra')({
  theme: 'nextra-theme-docs',
  themeConfig: './theme.config.tsx',
  latex: true,
  flexsearch: {
    codeblocks: false
  },
  defaultShowCopyCode: true,
  readingTime: true,
  mdxOptions: {
    remarkPlugins: [
      // remark-gfm, remark-math 등
    ],
    rehypePlugins: [
      // rehype-highlight, rehype-katex 등
    ]
  }
})

module.exports = withNextra({
  experimental: {
    turbo: {
      rules: {
        '*.svg': {
          loaders: ['@svgr/webpack'],
          as: '*.js',
        },
      },
    },
  },
  images: {
    domains: ['cdn.vercel.app'],
  },
  i18n: {
    locales: ['ko', 'en', 'ja', 'zh'],
    defaultLocale: 'ko',
    domains: [
      {
        domain: 'moai-docs.vercel.app',
        defaultLocale: 'en',
      },
      {
        domain: 'ko.moai-docs.vercel.app',
        defaultLocale: 'ko',
      },
    ],
  },
})
```

**SP2.4: Pagefind 검색 명세**
```typescript
// lib/pagefind/config.ts
export interface PagefindConfig {
  rootSelector: 'html';
  excerptLength: 30;
  filters: {
    language: ['ko', 'en', 'ja', 'zh'];
    category: ['guide', 'reference', 'tutorial'];
  };
  forceLanguage: 'ko';
  ranking: {
    pageLength: 1.5;
    termFrequency: 2.0,
    termSimilarity: 1.0
  };
}
```

**SP2.5: Bun 패키지 매니저 명세**
```json
{
  "packageManager": "bun@1.1.0",
  "scripts": {
    "dev": "bun next dev",
    "build": "bun next build",
    "start": "bun next start",
    "lint": "bun biome check --write",
    "type-check": "bun tsc --noEmit",
    "test": "bun test",
    "clean": "rm -rf .next out dist"
  },
  "engines": {
    "bun": ">=1.1.0",
    "node": ">=18.0.0"
  }
}
```

**SP2.6: Biome 린터/포매터 명세**
```json
{
  "biome": {
    "extends": ["@biomejs/biome"],
    "formatter": {
      "enabled": true,
      "formatWithErrors": false,
      "indentStyle": "space",
      "indentWidth": 2,
      "lineWidth": 100
    },
    "linter": {
      "enabled": true,
      "rules": {
        "recommended": true,
        "complexity": {
          "noExtraBooleanCast": "error",
          "noMultipleSpacesInRegularExpressionLiterals": "error"
        },
        "correctness": {
          "noUnusedVariables": "error",
          "useExhaustiveDependencies": "error"
        },
        "style": {
          "noNegationElse": "error",
          "useShorthandArrayType": "error"
        }
      }
    },
    "javascript": {
      "formatter": {
        "jsxQuoteStyle": "double",
        "quoteProperties": "asNeeded"
      }
    },
    "typescript": {
      "formatter": {
        "quoteStyle": "double"
      }
    }
  }
}
```

## 🎯 수락 기준

### 성능 기준
- 빌드 시간: 현재 대비 50% 이상 개선 (Bun + Turbopack 효과)
- LCP: 2.5초 → 1.9초 이하 (25% 개선)
- FCP: 1.8초 → 1.3초 이하 (30% 개선)
- 번들 크기: 현재와 동일하거나 감소
- 패키지 설치 속도: npm 대비 200% 향상 (Bun 효과)
- 린트/포맷팅 속도: ESLint/Prettier 대비 100% 향상 (Biome 효과)

### 기능성 기준
- 100+ MDX 파일 완전 호환성
- 4개 언어 모두 정상 동작
- 검색 기능 모든 언어에서 작동
- 모든 링크와 라우팅 정상 작동

### 품질 기준
- 0개의 빌드 에러
- 0개의 콘솔 에러
- 95% 이상의 Lighthouse 점수
- 모든 테스트 케이스 통과
- Biome 린트 검사 100% 통과
- Biome 포맷팅 일관성 100% 달성
- Bun 패키지 관리 안정성 확보

## 🔍 추적성

### 관련 SPEC
- @SPEC:DOCS-001: 기존 문서 사이트 설정
- @SPEC:BUILD-001: 빌드 프로세스 명세

### 관련 코드
- @CODE:NEXT-CONFIG-001: Next.js 설정 파일
- @CODE:NEXTRA-THEME-001: Nextra 테마 설정
- @CODE:SEARCH-001: 검색 기능 구현

### 관련 테스트
- @TEST:BUILD-001: 빌드 프로세스 테스트
- @TEST:SEARCH-001: 검색 기능 테스트
- @TEST:I18N-001: 다국어 기능 테스트