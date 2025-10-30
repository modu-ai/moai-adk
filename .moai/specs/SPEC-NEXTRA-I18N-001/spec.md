---
id: NEXTRA-I18N-001
version: 0.1.0
status: completed
created: 2025-10-31
updated: 2025-10-31
author: @GOOS
priority: high
category: feature
labels:
  - i18n
  - localization
  - next-intl
depends_on:
  - NEXTRA-SITE-001
scope:
  packages:
    - docs-site
  files:
    - next.config.js
    - middleware.ts
    - i18n/routing.ts
    - i18n/request.ts
    - messages/ko.json
    - messages/en.json
    - components/LanguageSwitcher.tsx
    - theme.config.tsx
---

# SPEC-NEXTRA-I18N-001: next-intl을 통한 다국어 지원 (한국어/영어)

## HISTORY

### v0.1.0 (2025-10-31)
- **COMPLETED**: next-intl i18n 구현 완료 및 SPEC 검증
- **CHANGES**:
  - `localePrefix` 설정 변경: `always` → `as-needed` (기본 언어 경로 최적화)
  - 실제 구현된 파일 목록 반영 (i18n/request.ts, components/LanguageSwitcher.tsx 추가)
  - 테스트 작성 완료 (@TEST:NEXTRA-I18N-001/003/005/007)
- **TAG CHAIN**: @SPEC:NEXTRA-I18N-001 → @CODE:NEXTRA-I18N-004/006/008/010/012 → @TEST:NEXTRA-I18N-001/003/005/007
- **STATUS**: draft → completed

### v0.0.1 (2025-10-31)
- **INITIAL**: next-intl 기반 다국어 지원 SPEC 작성
- **AUTHOR**: @GOOS
- **SECTIONS**: Environment, Assumptions, Requirements, Specifications
- **CONTEXT**: MkDocs → Nextra 마이그레이션 Phase 1 (v1.0 Core) 두 번째 SPEC
- **DEPENDENCIES**: SPEC-NEXTRA-SITE-001 (Next.js 14 + Nextra 4.0 기본 구조)

---

## @SPEC:NEXTRA-I18N-001 Overview

**목적**: next-intl 라이브러리를 사용하여 한국어와 영어 2개 언어를 지원하는 다국어 시스템을 구축합니다.

**범위**:
- next-intl 설치 및 설정
- URL 기반 언어 라우팅 (`/ko/`, `/en/`)
- 브라우저 로케일 자동 인식 및 리다이렉트
- 언어 전환 UI 컴포넌트
- 번역 파일 구조 및 관리 (`messages/ko.json`, `messages/en.json`)
- 기본 언어(fallback) 설정

**제외 사항**:
- 콘텐츠 번역 작업 → SPEC-NEXTRA-CONTENT-001
- 검색 엔진 최적화 (SEO) → Phase 2 (v1.1 Enhancement)
- RTL(Right-to-Left) 언어 지원 → Phase 2
- 언어별 날짜/시간 형식 → Phase 2

---

## Environment (환경 조건)

### Development Environment
- **Next.js**: v14.x (NEXTRA-SITE-001에서 설치됨)
- **React**: v18.2.0+
- **Node.js**: v20.x LTS
- **Package Manager**: npm 또는 yarn

### Production Environment
- **Hosting**: Vercel (NEXTRA-SITE-001에서 설정됨)
- **Domain**: adk.mo.ai.kr
- **URL Structure** (`localePrefix: 'as-needed'`):
  - 한국어 (기본): `https://adk.mo.ai.kr/` 또는 `https://adk.mo.ai.kr/ko/`
  - 영어: `https://adk.mo.ai.kr/en/`

### Required Dependencies
- `next-intl@^3.25.0`: Next.js 14 App Router 호환 i18n 라이브러리
- `negotiator@^0.6.3`: Accept-Language 헤더 파싱 (선택사항)

### Configuration Files
- `i18n/routing.ts`: 언어 라우팅 설정 (지원 언어, 기본 언어)
- `middleware.ts`: 언어 감지 및 리다이렉트 미들웨어
- `messages/ko.json`: 한국어 번역 파일
- `messages/en.json`: 영어 번역 파일
- `next.config.js`: next-intl 플러그인 설정

---

## Assumptions (가정)

### Technical Assumptions
1. **next-intl 3.x 안정성**: next-intl 3.25.0+가 Next.js 14 App Router와 완전히 호환됨
2. **파일 기반 번역**: JSON 파일로 번역 관리 (데이터베이스 없음)
3. **정적 라우팅**: 언어별 경로가 빌드 시 생성됨 (`/ko/`, `/en/`)
4. **기본 언어**: 한국어(ko)를 기본 언어로 설정 (번역 누락 시 fallback)

### Process Assumptions
1. **브라우저 로케일 우선**: `Accept-Language` 헤더를 기반으로 언어 자동 인식
2. **URL 기반 언어 전환**: 사용자가 `/ko/` 또는 `/en/` 경로로 언어 전환
3. **SEO 최적화 제외**: Phase 1에서는 `hreflang` 태그 및 Sitemap 제외

### User Assumptions
1. **주요 사용자 언어**: 한국어(ko), 영어(en)
2. **언어 전환 빈도**: 사용자는 세션 중 1회 이하로 언어 전환
3. **콘텐츠 동등성**: 모든 페이지가 한국어와 영어로 동등하게 번역됨

---

## Requirements (요구사항)

### UBIQUITOUS (보편적 요구사항)
- **REQ-I18N-001**: 시스템은 한국어(ko)와 영어(en) 2가지 언어를 기본으로 지원해야 한다
- **REQ-I18N-002**: 시스템은 URL 기반 언어 라우팅을 제공해야 한다 (기본 언어: `/`, 영어: `/en/`)
- **REQ-I18N-003**: 시스템은 모든 UI 텍스트를 번역 파일(`messages/ko.json`, `messages/en.json`)에서 가져와야 한다
- **REQ-I18N-004**: 시스템은 번역 파일 누락 시 기본 언어(한국어)로 폴백해야 한다

### EVENT-DRIVEN (이벤트 기반 요구사항)
- **REQ-I18N-005**: WHEN 사용자가 https://adk.mo.ai.kr에 방문하면, 시스템은 브라우저 `Accept-Language` 헤더를 인식하여 적절한 언어 경로(기본: `/` 또는 `/en/`)로 리다이렉트해야 한다
- **REQ-I18N-006**: WHEN 사용자가 `/` 또는 `/ko/` 경로에 접속하면, 시스템은 한국어 콘텐츠를 표시해야 한다
- **REQ-I18N-007**: WHEN 사용자가 `/en/` 경로에 접속하면, 시스템은 영어 콘텐츠를 표시해야 한다
- **REQ-I18N-008**: WHEN 사용자가 언어 전환 드롭다운에서 다른 언어를 선택하면, 시스템은 현재 페이지를 해당 언어 경로로 전환해야 한다 (예: `/docs` → `/en/docs`)

### STATE-DRIVEN (상태 기반 요구사항)
- **REQ-I18N-009**: WHILE 사용자가 `/` (한국어) 또는 `/en/` (영어) 경로에 있을 때, 시스템은 해당 언어의 번역 파일을 로드하여 UI 텍스트를 표시해야 한다
- **REQ-I18N-010**: WHILE 사용자가 특정 언어 경로에 있을 때, 시스템은 네비게이션 메뉴의 모든 링크에 해당 언어 접두사를 유지해야 한다

### OPTIONAL (선택적 기능)
- **REQ-I18N-011**: WHERE 사용자가 언어 선택기를 클릭하면, 시스템은 현재 페이지를 다른 언어로 표시할 수 있다
- **REQ-I18N-012**: WHERE 번역이 누락된 키가 있으면, 시스템은 개발 모드에서 경고를 표시할 수 있다

### CONSTRAINTS (제약사항)
- **REQ-I18N-013**: IF 번역 파일(`messages/ko.json` 또는 `messages/en.json`)이 누락되면, 시스템은 기본 언어(한국어)로 폴백해야 한다
- **REQ-I18N-014**: IF 번역 파일의 JSON 형식이 잘못되었으면, 시스템은 명확한 오류 메시지를 표시하고 빌드를 중단해야 한다
- **REQ-I18N-015**: IF 사용자가 지원하지 않는 언어 경로(예: `/fr/`)로 접속하면, 시스템은 404 페이지를 표시하거나 기본 언어(`/`)로 리다이렉트해야 한다

---

## Specifications (세부 설계)

### 1. next-intl Installation & Configuration

#### Installation
```bash
npm install next-intl@^3.25.0
```

#### Directory Structure
```
moai-adk-docs/
├── i18n/
│   └── routing.ts         # 언어 라우팅 설정
├── messages/
│   ├── ko.json           # 한국어 번역
│   └── en.json           # 영어 번역
├── middleware.ts          # 언어 감지 미들웨어
├── app/
│   ├── [locale]/         # 언어별 동적 라우팅
│   │   ├── layout.tsx    # 언어별 레이아웃
│   │   └── page.tsx      # 언어별 홈페이지
│   └── layout.tsx        # Root layout
└── next.config.js        # next-intl 플러그인 설정
```

---

### 2. Routing Configuration

#### i18n/routing.ts
```typescript
import { defineRouting } from 'next-intl/routing';

export const routing = defineRouting({
  // 지원 언어 목록
  locales: ['ko', 'en'],

  // 기본 언어 (폴백)
  defaultLocale: 'ko',

  // URL 경로 접두사 전략
  // 'as-needed': 기본 언어(/ko/)는 접두사 생략, 비기본 언어(/en/)만 표시
  // 'always': 모든 언어가 접두사 필요 (/ko/, /en/)
  localePrefix: 'as-needed'
});
```

---

### 3. Middleware for Language Detection

#### middleware.ts
```typescript
import createMiddleware from 'next-intl/middleware';
import { routing } from './i18n/routing';

export default createMiddleware(routing);

export const config = {
  // 미들웨어가 실행될 경로 패턴
  matcher: [
    // 모든 경로에 대해 실행 (/_next, /api, /_vercel, 정적 파일 제외)
    '/((?!_next|_vercel|.*\\..*).*)',
    // 루트 경로
    '/',
    // 언어별 경로
    '/(ko|en)/:path*'
  ]
};
```

**동작 방식**:
1. 사용자가 `/`에 접속 → `Accept-Language` 헤더 읽기
2. 브라우저 언어가 `ko-KR` → `/ko/`로 리다이렉트
3. 브라우저 언어가 `en-US` → `/en/`으로 리다이렉트
4. 브라우저 언어가 없거나 지원하지 않는 경우 → `/ko/`로 리다이렉트 (기본 언어)

---

### 4. Translation Files

#### messages/ko.json (한국어)
```json
{
  "Navigation": {
    "home": "홈",
    "docs": "문서",
    "api": "API 레퍼런스",
    "guide": "가이드",
    "about": "소개"
  },
  "Common": {
    "learnMore": "자세히 보기",
    "getStarted": "시작하기",
    "search": "검색",
    "language": "언어"
  },
  "Home": {
    "title": "MoAI-ADK 문서",
    "subtitle": "SPEC-First TDD 개발 프레임워크",
    "description": "AI 기반의 체계적인 개발 워크플로우로 고품질 코드를 작성하세요."
  },
  "Footer": {
    "copyright": "© 2025 Modu AI. All rights reserved.",
    "github": "GitHub",
    "discord": "Discord"
  }
}
```

#### messages/en.json (영어)
```json
{
  "Navigation": {
    "home": "Home",
    "docs": "Documentation",
    "api": "API Reference",
    "guide": "Guide",
    "about": "About"
  },
  "Common": {
    "learnMore": "Learn More",
    "getStarted": "Get Started",
    "search": "Search",
    "language": "Language"
  },
  "Home": {
    "title": "MoAI-ADK Documentation",
    "subtitle": "SPEC-First TDD Development Framework",
    "description": "Write high-quality code with AI-powered systematic development workflows."
  },
  "Footer": {
    "copyright": "© 2025 Modu AI. All rights reserved.",
    "github": "GitHub",
    "discord": "Discord"
  }
}
```

**번역 파일 구조 규칙**:
- 네임스페이스 기반 그룹핑 (Navigation, Common, Home, Footer 등)
- 키 이름은 camelCase 사용
- 값은 실제 표시될 텍스트
- 모든 번역 파일은 동일한 키 구조를 유지

---

### 5. Next.js Configuration

#### next.config.js (업데이트)
```javascript
const withNextra = require('nextra')({
  theme: 'nextra-theme-docs',
  themeConfig: './theme.config.tsx',
});

const withNextIntl = require('next-intl/plugin')(
  // 번역 파일 로드 경로
  './i18n/request.ts'
);

module.exports = withNextIntl(withNextra({
  reactStrictMode: true,
  output: 'export',
  images: {
    unoptimized: true
  }
}));
```

#### i18n/request.ts
```typescript
import { getRequestConfig } from 'next-intl/server';
import { routing } from './routing';

export default getRequestConfig(async ({ locale }) => {
  // 지원하지 않는 언어는 기본 언어로 폴백
  if (!routing.locales.includes(locale as any)) {
    locale = routing.defaultLocale;
  }

  return {
    messages: (await import(`../messages/${locale}.json`)).default
  };
});
```

---

### 6. App Router Integration

#### app/[locale]/layout.tsx
```typescript
import { NextIntlClientProvider } from 'next-intl';
import { getMessages } from 'next-intl/server';
import { notFound } from 'next/navigation';
import { routing } from '@/i18n/routing';

export function generateStaticParams() {
  return routing.locales.map((locale) => ({ locale }));
}

export default async function LocaleLayout({
  children,
  params: { locale }
}: {
  children: React.ReactNode;
  params: { locale: string };
}) {
  // 지원하지 않는 언어는 404
  if (!routing.locales.includes(locale as any)) {
    notFound();
  }

  // 번역 파일 로드
  const messages = await getMessages();

  return (
    <html lang={locale}>
      <body>
        <NextIntlClientProvider messages={messages}>
          {children}
        </NextIntlClientProvider>
      </body>
    </html>
  );
}
```

#### app/[locale]/page.tsx
```typescript
import { useTranslations } from 'next-intl';

export default function HomePage() {
  const t = useTranslations('Home');

  return (
    <div>
      <h1>{t('title')}</h1>
      <p>{t('subtitle')}</p>
      <p>{t('description')}</p>
    </div>
  );
}
```

---

### 7. Language Switcher Component

#### components/LanguageSwitcher.tsx
```typescript
'use client';

import { useLocale, useTranslations } from 'next-intl';
import { usePathname, useRouter } from 'next/navigation';
import { routing } from '@/i18n/routing';

export default function LanguageSwitcher() {
  const locale = useLocale();
  const router = useRouter();
  const pathname = usePathname();
  const t = useTranslations('Common');

  const handleLanguageChange = (newLocale: string) => {
    // 현재 경로에서 언어 접두사 제거
    const pathWithoutLocale = pathname.replace(`/${locale}`, '');

    // 새로운 언어 접두사로 교체
    const newPath = `/${newLocale}${pathWithoutLocale}`;

    router.push(newPath);
  };

  return (
    <select
      value={locale}
      onChange={(e) => handleLanguageChange(e.target.value)}
      className="border rounded px-2 py-1"
      aria-label={t('language')}
    >
      <option value="ko">한국어</option>
      <option value="en">English</option>
    </select>
  );
}
```

---

### 8. Nextra Theme Integration

#### theme.config.tsx (업데이트)
```tsx
import { DocsThemeConfig } from 'nextra-theme-docs';
import LanguageSwitcher from './components/LanguageSwitcher';

const config: DocsThemeConfig = {
  logo: <span>MoAI-ADK Documentation</span>,
  project: {
    link: 'https://github.com/modu-ai/moai-adk'
  },
  docsRepositoryBase: 'https://github.com/modu-ai/moai-adk/tree/main/docs',
  footer: {
    text: 'MoAI-ADK © 2025 Modu AI'
  },
  navbar: {
    extraContent: () => <LanguageSwitcher />
  }
};

export default config;
```

---

## Traceability (추적성)

### Related SPECs
- **@SPEC:NEXTRA-SITE-001**: Next.js 14 + Nextra 4.0 기본 구조 (선행 조건)
- **@SPEC:NEXTRA-CONTENT-001**: 콘텐츠 마이그레이션 (후속 작업)

### Related Tasks
- **@TASK:I18N-SETUP-001**: next-intl 설치 및 설정
- **@TASK:I18N-ROUTING-001**: 언어 라우팅 및 미들웨어 구성
- **@TASK:I18N-TRANSLATION-001**: 번역 파일 구조 생성 (ko.json, en.json)
- **@TASK:I18N-SWITCHER-001**: 언어 전환 UI 컴포넌트 구현

### Related Tests
- **@TEST:I18N-ROUTE-001**: `/ko/`, `/en/` 경로 접근 테스트
- **@TEST:I18N-REDIRECT-001**: 루트 경로(`/`) 리다이렉트 테스트
- **@TEST:I18N-TRANSLATE-001**: 번역 파일 로드 및 표시 테스트
- **@TEST:I18N-SWITCHER-001**: 언어 전환 동작 테스트

### Related Documentation
- **@DOC:I18N-GUIDE-001**: 다국어 개발 가이드 (번역 파일 작성법)
- **@DOC:I18N-TROUBLESHOOTING-001**: i18n 트러블슈팅 가이드

---

## Risks & Mitigation (위험 요소 및 대응 방안)

### Risk 1: next-intl 3.x와 Nextra 4.0 호환성 이슈
- **확률**: Medium
- **영향**: High
- **대응**:
  - next-intl GitHub Issues에서 Nextra 관련 이슈 검색
  - 필요시 next-intl 2.x로 다운그레이드
  - Nextra 커뮤니티에 문의

### Risk 2: 번역 파일 동기화 문제
- **확률**: High
- **영향**: Medium
- **대응**:
  - 번역 키 누락 감지 스크립트 작성 (npm script)
  - CI/CD에서 번역 파일 검증 자동화
  - 개발 모드에서 누락된 번역 키 경고 표시

### Risk 3: URL 구조 변경으로 인한 SEO 영향
- **확률**: Low
- **영향**: Medium
- **대응**:
  - 기존 URL에서 새 URL로 301 리다이렉트 설정
  - Google Search Console에 새 URL 구조 등록
  - Phase 2에서 `hreflang` 태그 추가

### Risk 4: 정적 빌드 시 언어별 페이지 생성 실패
- **확률**: Medium
- **영향**: High
- **대응**:
  - `generateStaticParams()` 함수 검증
  - 로컬 빌드에서 `out/ko/`, `out/en/` 디렉토리 확인
  - 빌드 오류 발생 시 next-intl 문서 재검토

---

## Success Criteria (성공 기준)

### Definition of Done
- ✅ next-intl 설치 완료 및 `package.json`에 의존성 추가됨
- ✅ `/` (한국어), `/en/` (영어) 경로 접근 시 해당 언어 콘텐츠 표시됨
- ✅ 루트 경로(`/`) 접속 시 브라우저 언어에 따라 적절한 언어로 표시됨
- ✅ 언어 전환 드롭다운에서 언어 선택 시 즉시 전환됨
- ✅ 번역 파일(`messages/ko.json`, `messages/en.json`) 정상 로드됨
- ✅ 번역 누락 시 기본 언어(한국어)로 폴백됨
- ✅ 빌드 성공: 정적 사이트 생성 확인
- ✅ Vercel 배포 후 https://adk.mo.ai.kr/, https://adk.mo.ai.kr/en/ 정상 접속

---

## Appendix

### Reference Documents
- [next-intl Documentation](https://next-intl.dev)
- [Next.js 14 Internationalization Guide](https://nextjs.org/docs/app/building-your-application/routing/internationalization)
- [Nextra i18n Support](https://nextra.site/docs/guide/i18n)

### Related GitHub Issues
- TBD (이슈 생성 후 링크 추가)

### Notes
- Phase 1에서는 한국어/영어만 지원하며, 추가 언어는 Phase 2에서 고려
- SEO 최적화(`hreflang` 태그, Sitemap)는 Phase 2에서 추가
- RTL(Right-to-Left) 언어 지원은 추후 검토
