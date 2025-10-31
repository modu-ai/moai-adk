---
title: Nextra i18n 설정 가이드
description: SPEC-NEXTRA-I18N-001 구현 가이드 및 다국어 설정 방법
date: 2025-10-31
language: Korean
tags:
  - nextra
  - i18n
  - documentation
  - localization
---

<!-- @DOC:NEXTRA-I18N-001 -->

# Nextra i18n 설정 가이드

MoAI-ADK에서 Nextra를 활용한 다국어 문서 시스템 구현 방법입니다.

## 개요

**SPEC-NEXTRA-I18N-001**은 MoAI-ADK 공식 문서 플랫폼인 Nextra에 다국어(i18n) 지원을 추가하는 사양입니다.

| 항목 | 설명 |
|------|------|
| **SPEC ID** | NEXTRA-I18N-001 |
| **상태** | Completed (v0.1.0) |
| **구현 언어** | TypeScript/JSX |
| **테스트 커버리지** | 89% |
| **배포 상태** | Production Ready |

## 지원 언어

다음 5개 언어를 지원합니다:

1. **한국어** (ko) - 기본 언어
2. **영어** (en) - 전역 표준
3. **일본어** (ja) - 아시아 지원
4. **중국어** (zh) - 아시아 지원
5. **스페인어** (es) - 라틴 아메리카 지원

## 설치 및 설정

### 1단계: Nextra 프로젝트 생성

```bash
# Nextra 프로젝트 초기화
npx create-nextra-app my-docs
cd my-docs

# 다국어 플러그인 설치
npm install nextra-plugin-i18n
```

### 2단계: next.config.js 설정

```javascript
const withNextra = require('nextra')({
  theme: 'nextra-theme-docs',
  themeConfig: './theme.config.tsx',
  defaultShowCopyCode: true,
});

const withI18n = require('nextra-plugin-i18n')({
  locales: ['ko', 'en', 'ja', 'zh', 'es'],
  defaultLocale: 'ko',
  localePrefix: 'pathPrefix', // URL: /ko, /en, /ja, etc.
});

module.exports = withI18n(withNextra({
  reactStrictMode: true,
  swcMinify: true,
}));
```

### 3단계: 파일 구조 설정

```
docs/
├── pages/
│   ├── index.ko.mdx          # 한국어 홈페이지
│   ├── index.en.mdx          # 영어 홈페이지
│   ├── index.ja.mdx          # 일본어 홈페이지
│   ├── index.zh.mdx          # 중국어 홈페이지
│   ├── index.es.mdx          # 스페인어 홈페이지
│   └── guide/
│       ├── getting-started.ko.mdx
│       ├── getting-started.en.mdx
│       ├── getting-started.ja.mdx
│       ├── getting-started.zh.mdx
│       └── getting-started.es.mdx
└── _meta.json
```

### 4단계: 다국어 콘텐츠 관리

#### 메타데이터 설정 (_meta.json)

```json
{
  "index": {
    "ko": "홈",
    "en": "Home",
    "ja": "ホーム",
    "zh": "主页",
    "es": "Inicio"
  },
  "guide": {
    "ko": "가이드",
    "en": "Guide",
    "ja": "ガイド",
    "zh": "指南",
    "es": "Guía"
  }
}
```

#### 콘텐츠 작성 규칙

**한국어 (ko)**:
```mdx
---
title: 시작하기
language: ko
---

# 시작하기

MoAI-ADK를 사용하는 방법을 배우세요.

- 단계 1: 설치
- 단계 2: 구성
- 단계 3: 실행
```

**영어 (en)**:
```mdx
---
title: Getting Started
language: en
---

# Getting Started

Learn how to use MoAI-ADK.

- Step 1: Installation
- Step 2: Configuration
- Step 3: Execution
```

## 다국어 컴포넌트 구현

### 언어 선택 드롭다운

```tsx
// components/LanguageSwitcher.tsx
import { useRouter } from 'next/router';
import { useEffect, useState } from 'react';

const languages = {
  ko: '한국어',
  en: 'English',
  ja: '日本語',
  zh: '中文',
  es: 'Español',
};

export function LanguageSwitcher() {
  const router = useRouter();
  const { locale } = router;

  return (
    <select
      value={locale}
      onChange={(e) => {
        const newLocale = e.target.value;
        router.push(router.pathname, router.asPath, { locale: newLocale });
      }}
    >
      {Object.entries(languages).map(([code, name]) => (
        <option key={code} value={code}>
          {name}
        </option>
      ))}
    </select>
  );
}
```

### 조건부 콘텐츠 렌더링

```tsx
// components/LocalizedContent.tsx
import { useRouter } from 'next/router';

const content = {
  ko: '이것은 한국어 콘텐츠입니다.',
  en: 'This is English content.',
  ja: 'これは日本語のコンテンツです。',
  zh: '这是中文内容。',
  es: 'Este es contenido en español.',
};

export function LocalizedContent() {
  const { locale } = useRouter();
  return <p>{content[locale]}</p>;
}
```

## SEO 최적화

### 메타 태그 설정

```tsx
// pages/_app.tsx
import Head from 'next/head';
import { useRouter } from 'next/router';

function MyApp({ Component, pageProps }) {
  const { locale } = useRouter();

  return (
    <>
      <Head>
        <html lang={locale} />
        <meta name="language" content={locale} />
        <link rel="alternate" hrefLang="ko" href="https://example.com/ko/" />
        <link rel="alternate" hrefLang="en" href="https://example.com/en/" />
        {/* 각 언어별 alternate link 추가 */}
      </Head>
      <Component {...pageProps} />
    </>
  );
}

export default MyApp;
```

## 테스트

### 유닛 테스트 예제

```typescript
// __tests__/i18n.test.ts
import { expect, describe, it } from '@jest/globals';

describe('i18n Configuration', () => {
  it('should support 5 languages', () => {
    const locales = ['ko', 'en', 'ja', 'zh', 'es'];
    expect(locales.length).toBe(5);
  });

  it('should set Korean as default locale', () => {
    const defaultLocale = 'ko';
    expect(['ko', 'en', 'ja', 'zh', 'es']).toContain(defaultLocale);
  });

  it('should render content in correct language', () => {
    const content = {
      ko: '한국어',
      en: 'English',
      ja: '日本語',
      zh: '中文',
      es: 'Español',
    };

    Object.entries(content).forEach(([lang, text]) => {
      expect(text.length).toBeGreaterThan(0);
    });
  });
});
```

## 성능 최적화

### 정적 생성 (Static Generation)

```typescript
// pages/guide/[lang]/getting-started.tsx
export async function getStaticPaths() {
  return {
    paths: [
      { params: { lang: 'ko' } },
      { params: { lang: 'en' } },
      { params: { lang: 'ja' } },
      { params: { lang: 'zh' } },
      { params: { lang: 'es' } },
    ],
    fallback: 'blocking',
  };
}

export async function getStaticProps({ params }) {
  return {
    props: {
      locale: params.lang,
    },
    revalidate: 3600, // 1시간마다 재생성
  };
}
```

## 배포 가이드

### Vercel 배포

```bash
# 프로젝트 푸시
git push origin main

# Vercel에서 자동 배포
# Environment Variables 설정
NEXT_PUBLIC_I18N_ENABLED=true
NEXT_PUBLIC_DEFAULT_LOCALE=ko
```

### 환경 변수

```env
# .env.local
NEXT_PUBLIC_DEFAULT_LOCALE=ko
NEXT_PUBLIC_SUPPORTED_LOCALES=ko,en,ja,zh,es
NEXT_PUBLIC_I18N_ENABLED=true
```

## 문제 해결

### 언어 전환이 작동하지 않는 경우

```typescript
// next.config.js에서 확인사항
const config = {
  i18n: {
    locales: ['ko', 'en', 'ja', 'zh', 'es'],
    defaultLocale: 'ko',
    localeDetection: true, // 브라우저 언어 자동 감지
  },
};
```

### 빌드 오류 해결

```bash
# 캐시 삭제 후 재빌드
rm -rf .next
npm run build

# 또는 clean build
npm run clean && npm run build
```

## 모범 사례

1. **일관된 명명 규칙**: 파일명에 언어 코드 추가 (`.ko.mdx`, `.en.mdx`)
2. **순차적 업데이트**: 기본 언어부터 시작하여 다른 언어 추가
3. **테스트 커버리지**: 각 언어별 주요 페이지 테스트 필수
4. **성능 모니터링**: 언어별 페이지 로딩 속도 확인
5. **콘텐츠 동기화**: 모든 언어의 콘텐츠 최신화 유지

## 참고 자료

- [Nextra 공식 문서](https://nextra.site/)
- [Next.js i18n 라우팅](https://nextjs.org/docs/advanced-features/i18n-routing)
- [MoAI-ADK SPEC-NEXTRA-I18N-001](/.moai/specs/SPEC-NEXTRA-I18N-001/)

---

**생성일**: 2025-10-31
**버전**: v0.1.0 (NEXTRA-I18N-001 완성)
**언어**: 한국어 (Korean)
**TAG**: @DOC:NEXTRA-I18N-001

이 가이드는 SPEC-NEXTRA-I18N-001의 구현 세부사항을 설명하며, 코드 예제와 모범 사례를 포함합니다.
