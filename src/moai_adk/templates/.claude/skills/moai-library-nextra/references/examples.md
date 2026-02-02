# Nextra Documentation Framework Examples

실용적인 예시를 통해 Nextra 기반 문서 사이트 구축 패턴을 학습합니다.

---

## Example 1: 기본 Nextra 프로젝트 설정

**Scenario**: 새로운 문서 사이트를 Nextra로 초기화하는 상황

**Input**:
```bash
# Nextra 프로젝트 생성
npx create-nextra-app@latest my-docs --template docs
cd my-docs

# 프로젝트 구조 확인
tree -L 2
```

**Output**:
```
my-docs/
├── pages/
│   ├── _app.tsx
│   ├── _meta.json
│   ├── index.mdx
│   └── docs/
│       ├── _meta.json
│       ├── getting-started.mdx
│       └── guide.mdx
├── public/
│   └── favicon.ico
├── theme.config.tsx
├── next.config.js
├── package.json
└── tsconfig.json
```

**Explanation**: Nextra는 Next.js 기반의 문서 프레임워크로, 파일 시스템 라우팅을 사용합니다. pages/ 디렉토리의 MDX 파일이 자동으로 문서 페이지가 되며, _meta.json으로 네비게이션을 구성합니다.

---

## Example 2: 테마 설정 커스터마이징

**Scenario**: 브랜딩과 기능을 포함한 완전한 테마 설정

**Input**:
```typescript
// theme.config.tsx
import { DocsThemeConfig } from 'nextra-theme-docs';
import { useRouter } from 'next/router';
import { useConfig } from 'nextra-theme-docs';

const config: DocsThemeConfig = {
  // 브랜딩
  logo: (
    <span style={{ fontWeight: 800 }}>
      <svg width="24" height="24" viewBox="0 0 24 24">
        {/* 로고 SVG */}
      </svg>
      My Documentation
    </span>
  ),
  logoLink: '/',

  // 프로젝트 링크
  project: {
    link: 'https://github.com/myorg/myproject'
  },
  docsRepositoryBase: 'https://github.com/myorg/myproject/tree/main/docs',

  // 채팅/지원 링크
  chat: {
    link: 'https://discord.gg/myproject'
  },

  // 네비게이션
  navigation: {
    prev: true,
    next: true
  },

  // 사이드바
  sidebar: {
    defaultMenuCollapseLevel: 1,
    toggleButton: true,
    autoCollapse: true
  },

  // 목차 (Table of Contents)
  toc: {
    backToTop: true,
    float: true,
    title: 'On This Page'
  },

  // 피드백
  feedback: {
    content: 'Question? Give us feedback',
    labels: 'feedback'
  },

  // 편집 링크
  editLink: {
    text: 'Edit this page on GitHub'
  },

  // 푸터
  footer: {
    text: (
      <span>
        MIT {new Date().getFullYear()} My Project.
        Built with Nextra.
      </span>
    )
  },

  // SEO
  head: function useHead() {
    const { title } = useConfig();
    const { route } = useRouter();
    const socialCard = 'https://myproject.com/og.png';

    return (
      <>
        <meta name="viewport" content="width=device-width, initial-scale=1.0" />
        <meta property="og:title" content={title ? title + ' - My Docs' : 'My Docs'} />
        <meta property="og:image" content={socialCard} />
        <meta property="og:url" content={`https://docs.myproject.com${route}`} />
        <meta name="twitter:card" content="summary_large_image" />
        <link rel="icon" href="/favicon.ico" />
      </>
    );
  },

  // 타이틀 템플릿
  useNextSeoProps() {
    const { asPath } = useRouter();
    if (asPath === '/') {
      return { titleTemplate: 'My Documentation' };
    }
    return { titleTemplate: '%s - My Docs' };
  },

  // 다크 모드
  darkMode: true,
  nextThemes: {
    defaultTheme: 'system'
  }
};

export default config;
```

**Output**:
```
문서 사이트 기능:
- 커스텀 로고와 브랜딩
- GitHub 연동 (소스 보기, 편집 링크)
- Discord 채팅 지원
- 반응형 사이드바 (자동 축소)
- 페이지 내 목차 (플로팅)
- SEO 최적화 (Open Graph, Twitter Cards)
- 다크/라이트 테마 자동 전환
- 피드백 버튼
- 이전/다음 페이지 네비게이션
```

**Explanation**: theme.config.tsx는 Nextra 사이트의 모든 설정을 담당합니다. 브랜딩, 네비게이션, SEO, 피드백 기능 등을 한 곳에서 관리할 수 있습니다.

---

## Example 3: MDX 컴포넌트와 인터랙티브 문서

**Scenario**: React 컴포넌트를 활용한 인터랙티브 문서 작성

**Input**:
```mdx
// pages/docs/components.mdx
---
title: Component Examples
description: Interactive component documentation
---

import { Callout, Tabs, Tab, Cards, Card, Steps } from 'nextra/components';

# Component Library

<Callout type="info">
  This page demonstrates interactive documentation features.
</Callout>

## Installation

<Tabs items={['npm', 'yarn', 'pnpm']}>
  <Tab>
    ```bash
    npm install @myproject/components
    ```
  </Tab>
  <Tab>
    ```bash
    yarn add @myproject/components
    ```
  </Tab>
  <Tab>
    ```bash
    pnpm add @myproject/components
    ```
  </Tab>
</Tabs>

## Quick Start

<Steps>
### Import the component

Import the Button component from the library:

```tsx
import { Button } from '@myproject/components';
```

### Use in your app

Add the Button to your JSX:

```tsx
function App() {
  return <Button variant="primary">Click me</Button>;
}
```

### Customize as needed

Adjust props to match your design:

```tsx
<Button variant="secondary" size="lg" disabled>
  Large Secondary Button
</Button>
```
</Steps>

## Component Cards

<Cards>
  <Card title="Button" href="/docs/components/button">
    Primary interaction component
  </Card>
  <Card title="Input" href="/docs/components/input">
    Text input with validation
  </Card>
  <Card title="Modal" href="/docs/components/modal">
    Overlay dialog component
  </Card>
</Cards>

## Live Example

export const LiveButton = () => {
  const [count, setCount] = React.useState(0);
  return (
    <button
      onClick={() => setCount(c => c + 1)}
      className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
    >
      Clicked {count} times
    </button>
  );
};

Try the interactive button below:

<LiveButton />

## Callout Types

<Callout type="default">
  Default callout for general information.
</Callout>

<Callout type="info">
  Info callout for helpful tips.
</Callout>

<Callout type="warning">
  Warning callout for important notices.
</Callout>

<Callout type="error">
  Error callout for critical information.
</Callout>
```

**Output**:
```
렌더링된 문서 페이지:
1. Callout 박스 - 정보 강조
2. Tabs - 패키지 매니저별 설치 명령
3. Steps - 단계별 가이드 (번호 자동 부여)
4. Cards - 관련 문서 링크 카드
5. Live Example - 클릭 카운터 인터랙티브 버튼
6. 다양한 Callout 타입 시연
```

**Explanation**: Nextra의 MDX 지원으로 마크다운과 React 컴포넌트를 혼합할 수 있습니다. 내장 컴포넌트(Callout, Tabs, Steps, Cards)와 커스텀 컴포넌트를 활용하여 풍부한 문서를 작성합니다.

---

## Common Patterns

### Pattern 1: 네비게이션 구조화 (_meta.json)

사이드바 메뉴와 페이지 순서를 정의하는 패턴입니다.

```json
// pages/_meta.json (루트 레벨)
{
  "index": {
    "title": "Home",
    "type": "page",
    "display": "hidden"
  },
  "docs": {
    "title": "Documentation",
    "type": "page"
  },
  "blog": {
    "title": "Blog",
    "type": "page",
    "theme": {
      "layout": "full"
    }
  },
  "about": {
    "title": "About",
    "type": "page"
  }
}

// pages/docs/_meta.json (섹션 레벨)
{
  "index": "Overview",
  "getting-started": "Getting Started",
  "---": {
    "type": "separator",
    "title": "Guide"
  },
  "installation": "Installation",
  "configuration": "Configuration",
  "advanced": {
    "title": "Advanced Topics",
    "type": "menu",
    "items": {
      "performance": "Performance",
      "security": "Security"
    }
  },
  "api-reference": {
    "title": "API Reference",
    "href": "/api"
  },
  "github": {
    "title": "GitHub",
    "href": "https://github.com/myorg/repo",
    "newWindow": true
  }
}
```

### Pattern 2: 검색 최적화

FlexSearch 기반 전문 검색을 활성화하는 패턴입니다.

```javascript
// next.config.js
const withNextra = require('nextra')({
  theme: 'nextra-theme-docs',
  themeConfig: './theme.config.tsx',
  search: {
    codeblocks: true  // 코드 블록도 검색 대상에 포함
  },
  defaultShowCopyCode: true,
  flexsearch: {
    codeblocks: true
  }
});

module.exports = withNextra({
  // Next.js 설정
  reactStrictMode: true,
  images: {
    domains: ['example.com']
  }
});
```

```typescript
// theme.config.tsx - 검색 설정
const config: DocsThemeConfig = {
  search: {
    placeholder: 'Search documentation...',
    emptyResult: (
      <span className="block p-8 text-center text-gray-400">
        No results found.
      </span>
    )
  }
};
```

### Pattern 3: 다국어 (i18n) 설정

국제화를 지원하는 패턴입니다.

```javascript
// next.config.js
const withNextra = require('nextra')({
  theme: 'nextra-theme-docs',
  themeConfig: './theme.config.tsx'
});

module.exports = withNextra({
  i18n: {
    locales: ['en', 'ko', 'ja'],
    defaultLocale: 'en'
  }
});
```

```typescript
// theme.config.tsx
const config: DocsThemeConfig = {
  i18n: [
    { locale: 'en', text: 'English' },
    { locale: 'ko', text: '한국어' },
    { locale: 'ja', text: '日本語' }
  ]
};
```

```
프로젝트 구조:
pages/
├── index.mdx           # 영어 (기본)
├── index.ko.mdx        # 한국어
├── index.ja.mdx        # 일본어
└── docs/
    ├── guide.mdx       # 영어
    ├── guide.ko.mdx    # 한국어
    └── guide.ja.mdx    # 일본어
```

---

## Anti-Patterns (피해야 할 패턴)

### Anti-Pattern 1: _meta.json 누락

**Problem**: _meta.json 없이 페이지를 추가하면 네비게이션 순서가 알파벳 순

```
pages/docs/
├── advanced.mdx
├── getting-started.mdx  # 알파벳 순으로 advanced 다음에 표시됨!
└── installation.mdx
```

**Solution**: 모든 디렉토리에 _meta.json 추가

```json
// pages/docs/_meta.json
{
  "getting-started": "Getting Started",  // 첫 번째
  "installation": "Installation",        // 두 번째
  "advanced": "Advanced"                 // 세 번째
}
```

### Anti-Pattern 2: 과도한 중첩 구조

**Problem**: 5단계 이상 깊은 중첩은 사용자 경험 저하

```
pages/docs/guide/api/v2/endpoints/users/create.mdx
# URL: /docs/guide/api/v2/endpoints/users/create
# 너무 깊음!
```

**Solution**: 2-3단계 이내로 구조화

```
pages/
├── docs/
│   ├── guide/
│   │   └── getting-started.mdx
│   └── api/
│       ├── overview.mdx
│       └── endpoints.mdx  # 모든 엔드포인트를 한 페이지에
```

### Anti-Pattern 3: 대용량 이미지 직접 포함

**Problem**: public/ 폴더에 최적화되지 않은 대용량 이미지

```mdx
<!-- 잘못된 예시 -->
![Large Image](/images/screenshot-4k.png)
<!-- 4K 이미지 그대로 로드 → 성능 저하 -->
```

**Solution**: Next.js Image 컴포넌트 사용

```mdx
<!-- 올바른 예시 -->
import Image from 'next/image';

<Image
  src="/images/screenshot.png"
  alt="Screenshot"
  width={800}
  height={450}
  placeholder="blur"
  blurDataURL="/images/screenshot-blur.png"
/>
```

### Anti-Pattern 4: 하드코딩된 링크

**Problem**: 절대 경로를 하드코딩하면 배포 환경 변경 시 문제

```mdx
<!-- 잘못된 예시 -->
Visit [our API](https://docs.myproject.com/api) for more info.
<!-- 도메인 변경 시 모든 링크 수정 필요 -->
```

**Solution**: 상대 경로 사용

```mdx
<!-- 올바른 예시 -->
Visit [our API](/api) for more info.

<!-- 또는 컴포넌트 활용 -->
import Link from 'next/link';

<Link href="/api">our API</Link>
```

---

## Deployment Checklist

Nextra 사이트 배포 전 확인 사항:

| 항목 | 확인 |
|------|------|
| theme.config.tsx에 모든 필수 설정 완료 | |
| 모든 디렉토리에 _meta.json 존재 | |
| 이미지 최적화 (WebP, 적절한 크기) | |
| SEO 메타데이터 설정 완료 | |
| 404 페이지 커스터마이징 | |
| 검색 기능 테스트 | |
| 다크 모드 테스트 | |
| 모바일 반응형 확인 | |
| 빌드 성공 (npm run build) | |
| 링크 깨짐 검사 | |

---

## Quick Reference

```bash
# 개발 서버
npm run dev

# 프로덕션 빌드
npm run build

# 정적 내보내기
npm run build && npm run export

# Vercel 배포
npx vercel
```

---

Version: 1.0.0
Last Updated: 2025-12-06
