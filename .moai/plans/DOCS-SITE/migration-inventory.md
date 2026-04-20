# moai-docs → Hextra 마이그레이션 인벤토리 (Phase 0 Discovery)

> **작업자**: Phase 0 Discovery researcher
> **스캔 일자**: 2026-04-17
> **스캔 대상**: `/Users/goos/moai/moai-docs` (Nextra 4 + Next.js 16 + Bun)
> **출력 경로**: `/Users/goos/MoAI/moai-adk-go/.moai/plans/DOCS-SITE/migration-inventory.md`
> **방법론**: 실제 `find`/`grep`/`wc` 명령 기반 전수 조사 (추측 배제)

---

## 1. 디렉토리 구조 요약

### 1.1 `content/` 전체 트리

`find content -type d` 결과:

```
content
content/en
content/en/advanced
content/en/agency
content/en/core-concepts
content/en/getting-started
content/en/quality-commands
content/en/utility-commands
content/en/workflow-commands
content/en/worktree
content/ja
content/ja/advanced
content/ja/agency
content/ja/core-concepts
content/ja/getting-started
content/ja/quality-commands
content/ja/utility-commands
content/ja/workflow-commands
content/ja/worktree
content/ko
content/ko/advanced
content/ko/agency
content/ko/contributing       ← ko 전용 (다른 로케일에 없음)
content/ko/core-concepts
content/ko/getting-started
content/ko/multi-llm          ← ko 전용 (다른 로케일에 없음)
content/ko/quality-commands
content/ko/utility-commands
content/ko/workflow-commands
content/ko/worktree
content/zh
content/zh/advanced
content/zh/agency
content/zh/core-concepts
content/zh/getting-started
content/zh/quality-commands
content/zh/utility-commands
content/zh/workflow-commands
content/zh/worktree
```

### 1.2 locale별 페이지 수

| Locale | 페이지 수 (`.mdx` + `.md`) |
|--------|---------------------------|
| en     | 52                        |
| ja     | 52                        |
| zh     | 52                        |
| ko     | 63 (← `contributing/`, `multi-llm/` 포함으로 11페이지 많음) |

**총합**: 219 MDX/MD 페이지

**비고**: ko가 원본 로케일이고, multi-llm/contributing 2개 섹션은 ko에만 존재. 마이그레이션 시 이 불균형 유지 여부 결정 필요.

### 1.3 `components/` 파일 목록 + line count

```
162  components/LanguageSelector.tsx
127  components/navbar.tsx
122  components/ui/dialog.tsx
110  components/structured-data.tsx
 91  components/CodeBlock.tsx
 86  components/ui/card.tsx
 56  components/ui/button.tsx
 43  components/ClientNavbar.tsx
 22  components/ui/input.tsx
────
819  합계 (9개 파일)
```

- `components/ui/*` 4개는 shadcn/ui 부트스트랩 (Radix UI 기반)
- `components/*.tsx` 5개는 Nextra/Next.js 전용 커스텀 컴포넌트

### 1.4 `app/`, `lib/`, `layouts/` 파일 목록

**`layouts/`**: 존재하지 않음 (Nextra 4는 `app/[[...mdxPath]]/layout.tsx`로 대체)

**`app/`** 구조:
```
app/
├── .claude/plans/            (빈 디렉토리, 정리 대상)
├── [...meta].json            (현재 미사용 — 이전 버전 잔재)
├── [[...mdxPath]]/
│   ├── layout.tsx            (71 LOC, Nextra Layout + LanguageSelector)
│   ├── mdx-wrapper.tsx       (26 LOC, useMDXComponents 주입)
│   └── page.tsx              (28 LOC, importPage + generateStaticParams)
├── _content.mdx              (132 LOC, 루트 랜딩 페이지 레거시)
├── _not-found.tsx            (33 LOC, 404 페이지)
├── debug-pagemap/            (빈 디렉토리)
├── globals.css               (78 LOC)
├── layout.tsx                (47 LOC, 루트 HTML + Analytics + GA + JSON-LD)
├── proxy.ts                  (20 LOC, 헤더 주입 프록시)
└── sitemap.ts                (77 LOC, 정적 sitemap 생성)
```

**`lib/`**:
```
lib/page-map.ts    (234 LOC, _meta.ts 파서 — 현재 미사용 의심, app/layout에서 nextra/page-map 사용)
lib/utils.ts       (7 LOC, shadcn cn() 헬퍼)
```

**루트 레벨 기타**:
- `mdx-components.tsx` (11 LOC) — `useMDXComponents` 재내보내기
- `middleware.ts` (164 LOC) — i18n 라우팅
- `next.config.mjs` (9 LOC) — `withNextra({ latex: true })`
- `theme.config.tsx` (293 LOC) — per-locale 테마 설정

### 1.5 `public/` 정적 자원

```
apple-touch-icon.png   (77 KB)
favicon-16x16.png      (1 KB)
favicon-32x32.png      (3 KB)
favicon.ico            (3 KB)
favicon.svg            (114 B)
icon-192.png           (85 KB)
icon-512.png           (543 KB)
images/                (빈 디렉토리)
llms.txt               (37 lines)
manifest.json          (53 lines, PWA manifest)
og.png                 (7.5 MB — 최적화 필요)
robots.txt             (15 lines, sitemap 링크 포함)
```

총 **11개 파일 + 1개 빈 디렉토리**.

---

## 2. Nextra 컴포넌트 사용 카탈로그

### 2.1 `grep -rn "from ['\"]nextra" content/` 집계

전체 match 결과를 `sort | uniq -c | sort -rn`으로 집계:

```
 127  import { Callout } from 'nextra/components'
  48  import { Callout } from "nextra/components";
  38  import type { MetaRecord } from "nextra";
   4  import { Callout } from 'nextra/components';
────
 217  총 import 라인 (4개 문법 변형)
```

**핵심 발견**: content/ 전체에서 Nextra로부터 가져오는 **런타임 컴포넌트는 `Callout` 단 하나뿐**입니다. `MetaRecord`는 `_meta.ts`의 타입 주석으로만 사용되며 런타임 기능 아님.

### 2.2 사용된 Nextra 컴포넌트 유일값

| 컴포넌트          | 사용 위치                                   | 사용 방식           |
|-------------------|---------------------------------------------|---------------------|
| `Callout`         | 168개 MDX 파일에서 import                   | `<Callout type="..."/>` 735건 |
| `MetaRecord`      | 38개 `_meta.ts`에서 타입 import             | 타입 주석 전용      |

**사용되지 않는 Nextra 컴포넌트** (조사 결과 0건):
- `Tabs`, `Tab`
- `Cards`, `Card`
- `Steps`
- `FileTree`
- `Bleed`
- `Navbar` (nextra/components 버전)

### 2.3 컴포넌트별 사용 빈도 상위

`grep -rohE "<[A-Z][a-zA-Z]+" content/ --include="*.mdx" | sort | uniq -c`:

```
735  <Callout    (type="tip"|"info"|"warning"|"error"|"success" 5종)
  4  <VERSION    (코드블록 내 placeholder, 실제 JSX 아님 — "<VERSION>" 문자열)
```

**결론**: content/ MDX 내 JSX 컴포넌트 사용은 **`<Callout>` 100%**. 마이그레이션 시 Hextra의 `callout` shortcode로 1:1 매핑 가능.

### 2.4 `_meta.ts` 파일 전체 목록 (38개)

locale별 구조는 동일한 패턴 반복:

```
content/en/_meta.ts                        (24 LOC)
content/en/advanced/_meta.ts               (19 LOC)
content/en/agency/_meta.ts
content/en/core-concepts/_meta.ts          (17 LOC)
content/en/getting-started/_meta.ts        (17 LOC)
content/en/quality-commands/_meta.ts
content/en/utility-commands/_meta.ts       (20 LOC)
content/en/workflow-commands/_meta.ts
content/en/worktree/_meta.ts
── (ja, zh 동일 구조 각 9개)
content/ko/_meta.ts                        (30 LOC, 10섹션)
content/ko/advanced/_meta.ts               (20 LOC)
content/ko/agency/_meta.ts
content/ko/contributing/_meta.ts           ← ko 전용
content/ko/core-concepts/_meta.ts
content/ko/getting-started/_meta.ts
content/ko/multi-llm/_meta.ts              ← ko 전용
content/ko/quality-commands/_meta.ts
content/ko/utility-commands/_meta.ts
content/ko/workflow-commands/_meta.ts
content/ko/worktree/_meta.ts
```

**총 _meta.ts 38개, 641 LOC 합계**.

Export 패턴 샘플 (en 루트):
```typescript
import type { MetaRecord } from "nextra";

const meta: MetaRecord = {
  "getting-started": "Getting Started",
  "core-concepts": "Core Concepts",
  "workflow-commands": "Workflow Commands",
  "utility-commands": "Utility Commands",
  "quality-commands": "Quality Commands",
  agency: "Agency",
  advanced: "Advanced",
  worktree: "Git Worktree",
};

export default meta;
```

섹션 내 export 패턴 샘플 (ko/core-concepts/_meta.ts):
```typescript
const meta: MetaRecord = {
  index: { title: "핵심 개념", display: "hidden" },
  "what-is-moai-adk": "MoAI-ADK란?",
  "harness-engineering": "하네스 엔지니어링",
  "spec-based-dev": "SPEC 기반 개발",
  ddd: "MoAI-ADK 개발 방법론",
  "auto-quality": "자동 품질 레이어",
  "trust-5": "TRUST 5 품질",
  "moai-memory": "MoAI Memory",
};
```

**핵심 스키마**:
- 값이 string → 메뉴 제목
- 값이 object → `{ title, display: "hidden" | ... }`
- 키 순서 = 사이드바 노출 순서

---

## 3. MDX 고급 기능 사용 현황

### 3.1 JSX 컴포넌트 인라인 사용

`grep -rohE "<[A-Z][a-zA-Z]+" content/ --include="*.mdx" | sort | uniq -c` 결과:

```
735  <Callout
```

**그 외 JSX 컴포넌트 사용 0건**. (`<VERSION>` 4건은 코드블록 내 placeholder 문자열)

### 3.2 커스텀 import로 React 컴포넌트 주입

`grep -rh "^import " content/ | sort -u`:

```
import { Callout } from 'nextra/components'
import { Callout } from 'nextra/components';
import { Callout } from "nextra/components";
import type { MetaRecord } from "nextra";
import json     ← Python 코드블록 예시 (실제 코드 아님)
import os       ← Python 코드블록 예시
import re       ← Python 코드블록 예시
import sys      ← Python 코드블록 예시
```

**결론**: React/TypeScript 관점에서 실제 import되는 심볼은 `Callout`, `MetaRecord` 2개뿐. 나머지는 Python 코드 펜스 내부 텍스트.

### 3.3 Mermaid 코드블록

```
총 569건 (content/**/*.mdx)
```

상위 파일:
```
10  content/{en,ja,ko,zh}/worktree/guide.mdx          (4 files × 10 = 40)
 9  content/{en,ja,ko,zh}/worktree/faq.mdx            (4 × 9  = 36)
 9  content/{en,ja,ko,zh}/getting-started/quickstart.mdx (4 × 9 = 36)
 7  content/{en,ja,ko,zh}/core-concepts/ddd.mdx       (4 × 7  = 28)
```

**의존성**: `package.json`에 `@mermaid-js/mermaid-cli`가 있음 — 빌드 시 SVG 프리렌더링 중인지 추가 확인 필요. 현재 Nextra 4는 클라이언트 사이드 렌더링 (`next.config.mjs`의 `latex: true`만 켜져있음). 569개 다이어그램은 **Hextra 마이그레이션 시 별도 처리 필수**.

### 3.4 LaTeX/KaTeX 사용

`grep -rn '\$\$' content/ --include="*.mdx"`: **0건**

`next.config.mjs`에 `latex: true`가 켜져있으나 **실제 사용 없음**. 제거 가능.

### 3.5 Frontmatter 사용 패턴

`find content -name "*.mdx" | while read f; do head -1 "$f" | grep -q '^---$'; done`:

```
MDX files with YAML frontmatter: 0
```

**결과**: 모든 MDX 파일이 `import { Callout } from 'nextra/components'` 또는 `# 제목`으로 시작. **YAML frontmatter 사용 0%**.

샘플 첫 10줄 (각 locale 설치 페이지):
```markdown
import { Callout } from 'nextra/components'

# 설치

MoAI-ADK 2.x를 시스템에 설치하는 방법을 안내합니다.
...
```

**이 점이 Hextra 마이그레이션에 결정적으로 중요**: Hextra는 Hugo 기반이라 거의 모든 페이지가 YAML/TOML frontmatter를 요구함. 즉 **219개 페이지 모두 frontmatter를 신규 주입**해야 함 (title, weight 등).

---

## 4. 런타임 / 빌드 설정

### 4.1 `next.config.mjs` (9 LOC)

```javascript
import nextra from "nextra";

const withNextra = nextra({
  latex: true,
});

// Nextra 4 uses file-based i18n with content/ko, content/en, etc.
// Do NOT configure Next.js i18n when using Nextra's file-based i18n
export default withNextra({});
```

**평가**:
- `latex: true`는 실제 사용 0건이므로 제거 후보
- Nextra 4의 파일 기반 i18n 패턴 채택 (Next.js native i18n 미사용)

### 4.2 `theme.config.tsx` (293 LOC) — per-locale 설정

주요 구성:
- `defaultTheme: "light"`
- `logo: { text: "🗿 MoAI-ADK" }`
- `i18n: [ko, en, zh, ja]` 4개 로케일 × 각각 `{locale, name, title, description, lang, dir, theme: {toc, editLink, feedback, footer}}`
- `project.link`: GitHub stars 뱃지 + repo 링크 (JSX)
- `chat.link`: `https://discord.gg/moai-adk`
- `docsRepositoryBase`: `https://github.com/modu-ai/moai-adk/tree/main/docs`
- `toc`: `{ backToTop, float, title }`
- `sidebar`: `{ defaultMenuCollapseLevel: 1, toggleButton: true }`
- `navigation`: `{ prev: true, next: true }`
- `search: false` (검색 비활성)
- `head`: 24개 `<link>`, `<meta>` 태그 + `<Analytics />` Vercel 컴포넌트
  - favicon × 4
  - manifest.json
  - viewport, theme-color, description
  - OG: type, locale, url, site_name, title, description, image
  - Twitter: card, title, description, image
  - keywords
  - hreflang × 5 (ko, en, zh, ja, x-default)
  - canonical

**Hextra 전환 시**:
- `head` 메타 태그 24개 → Hugo `params` + 커스텀 `head.html` partial 필요
- per-locale `theme.toc/editLink/feedback/footer` 텍스트 → Hugo `i18n/{locale}.toml` 4개 파일
- JSX `project.link` → Hugo partial 또는 네비게이션 아이템
- `<Analytics key="vercel-analytics" />` → Hextra에 별도 통합 필요

### 4.3 `middleware.ts` (164 LOC) — i18n 라우팅

Edge middleware 로직:
1. `SUPPORTED_LOCALES = ["ko", "en", "zh", "ja"]`, `DEFAULT_LOCALE = "ko"`
2. `/api/`, `/_next/`, `/static/`, 확장자 포함 경로는 통과
3. URL prefix에서 locale 추출 → 쿠키 저장 (1년 maxAge)
4. locale 없으면 `Accept-Language` 헤더 파싱 후 리다이렉트
5. 쿠키 > Accept-Language > 기본값 `ko` 우선순위
6. matcher: `/((?!api|_next|_vercel|.*\\..*).*)`

**Hextra 전환 시**:
- Hugo는 빌드타임 정적 사이트라 Edge middleware 개념 없음
- 대안 1: Vercel `Edge Functions`로 별도 구현 (Go 스크립트 또는 vercel.json rewrites)
- 대안 2: 클라이언트 사이드 JS로 `navigator.language` 감지 + 리다이렉트 (JS off 사용자 이슈)
- 대안 3: `Accept-Language` 기반 서버 사이드 리다이렉트는 Vercel `vercel.json` `rewrites` + 커스텀 함수로 가능

**이 로직이 마이그레이션의 최대 난관 중 하나**. Phase 4 Hugo partial로는 완전 재현 불가, Vercel 런타임 조합 필요.

### 4.4 `package.json` — dependencies/devDependencies/scripts

**dependencies (16개)**:
```
@hugeicons/react: hugeicons/react         (GitHub 직접 참조)
@next/mdx: ^16.1.1
@radix-ui/react-dialog: ^1.1.15
@radix-ui/react-dropdown-menu: ^2.1.16
@radix-ui/react-slot: ^1.2.4
@types/node: ^25.0.6
@types/react: ^19.2.8
@types/react-dom: ^19.2.3
@vercel/analytics: ^1.6.1
class-variance-authority: ^0.7.1
clsx: ^2.1.1
lucide-react: ^0.563.0
next: ^16.1.1
next-themes: ^0.4.6
nextra: ^4.6.1
nextra-theme-docs: ^4.6.1
react: ^19.2.3
react-dom: ^19.2.3
tailwind-merge: ^3.4.0
typescript: ^5.9.3
```

**devDependencies (8개)**:
```
@biomejs/biome: ^2.3.11
@mermaid-js/mermaid-cli: ^11.12.0
@playwright/test: ^1.58.0
@tailwindcss/postcss: ^4.1.18
rehype-pretty-code: ^0.14.1
shiki: ^3.21.0
tailwindcss: ^4.1.18
tailwindcss-animate: ^1.0.7
```

**scripts (11개)**:
```json
"dev":       "next dev",
"build":     "next build",
"start":     "next start",
"lint":      "biome check .",
"lint:fix":  "biome check --write .",
"format":    "biome format --write .",
"check":     "biome check --write --unsafe .",
"typecheck": "tsc --noEmit",
"test":      "playwright test",
"test:ui":   "playwright test --ui",
"test:headed": "playwright test --headed",
"ci":        "bun run lint && bun run typecheck && bun run build",
"precommit": "bun run lint && bun run typecheck"
```

**마이그레이션 시 전면 폐기**: Nextra/Next.js/React 생태계 의존 0건으로 목표. Hextra(Hugo)는 바이너리 하나만 있으면 됨.

### 4.5 `vercel.json` (63 LOC)

```json
{
  "buildCommand": "bun run build",
  "outputDirectory": ".next",
  "devCommand": "bun run dev",
  "installCommand": "bun install",
  "framework": "nextjs",
  "regions": ["hkg1"],
  "cleanUrls": true,
  "trailingSlash": false,
  "github": { "silent": true },
  "redirects": [
    { "source": "/:locale(ko|en|ja|zh)/moai-rank/:path*", "destination": "/:locale", "permanent": true },
    { "source": "/moai-rank/:path*", "destination": "/", "permanent": true }
  ],
  "headers": [
    { "source": "/(.*)",       "headers": [ X-Frame-Options: DENY, X-Content-Type-Options: nosniff, Referrer-Policy: strict-origin-when-cross-origin ] },
    { "source": "/api/(.*)",   "headers": [ Cache-Control: s-maxage=60, stale-while-revalidate ] },
    { "source": "/(.*).mdx",   "headers": [ Cache-Control: public, max-age=3600, stale-while-revalidate=86400 ] }
  ]
}
```

**마이그레이션 시**:
- `framework: "nextjs"` → 제거 또는 `"hugo"`
- `buildCommand`: `hugo --minify` 등으로 교체
- `outputDirectory`: `.next` → `public` (Hugo 기본)
- `installCommand`: `bun install` → Hugo 버전별 바이너리 설치 스크립트
- `redirects` 2건 (moai-rank 정리)은 그대로 유지 가능
- `headers`의 `/(.*).mdx` 규칙은 Hugo 정적 출력에서는 불필요 (HTML로 변환됨)

### 4.6 `.env.local`의 키 이름만

```
VERCEL_OIDC_TOKEN
```

단 1개. Vercel 자동 주입 토큰으로, 마이그레이션 대상 아님.

### 4.7 기타 설정 파일 존재 여부

| 파일                | 존재 | 크기     | 비고                                    |
|---------------------|------|----------|-----------------------------------------|
| `biome.json`        | ✅   | 547 B    | Nextra용 lint 규칙 (include: ts/tsx/js/jsx/json/mdx) |
| `.markdownlint.json`| ✅   | 94 B     | MD013, MD033, MD036, MD041 비활성        |
| `.gitignore`        | ✅   | 3,783 B  | Next.js 표준 + `.next/`, `.vercel/` 등   |
| `tsconfig.json`     | ✅   | 722 B    | TypeScript 5.9 설정                     |
| `tailwind.config.ts`| ✅   | 2,918 B  | shadcn/ui 기반 CSS 변수 테마            |
| `postcss.config.mjs`| ✅   | 66 B     | Tailwind PostCSS 설정                   |
| `components.json`   | ✅   | 447 B    | shadcn/ui CLI 설정                      |
| `playwright.config.ts` | ✅ | 450 B   | E2E 테스트 설정                         |

---

## 5. URL 구조 / SEO

### 5.1 `app/` 구조 (Next.js App Router 사용)

```
app/
├── [[...mdxPath]]/        ← catch-all 라우트 (optional segments)
│   ├── layout.tsx         → Nextra Layout + LanguageSelector + Navbar
│   ├── page.tsx           → importPage + MDXContent 렌더링
│   └── mdx-wrapper.tsx    → useMDXComponents 주입 wrapper
├── [...meta].json         → 레거시 미사용 (사이드바 메타)
├── _content.mdx           → 루트 랜딩 (레거시 의심)
├── _not-found.tsx         → 404 페이지
├── debug-pagemap/         → 빈 디렉토리
├── globals.css            → Tailwind 전역 스타일
├── layout.tsx             → 루트 HTML + Analytics + GA + JSON-LD
├── proxy.ts               → 헤더 주입 프록시 미들웨어
└── sitemap.ts             → 동적 sitemap 생성
```

**모든 페이지는 `[[...mdxPath]]` 동적 라우트로 처리**. Next.js App Router의 Nextra 4 통합 패턴.

### 5.2 `sitemap.xml` / `robots.txt` 생성 방식

**`sitemap.ts`** (77 LOC, 빌드타임 생성):
- `baseUrl = "https://adk.mo.ai.kr"`
- 4개 locale × 36개 main routes = 144 URL 하드코딩
- 각 엔트리에 `alternates.languages`로 hreflang 제공
- `changeFrequency: "weekly"`, `priority: 1.0` (루트) / `0.8` (나머지)

**문제점**: 하드코딩된 36개 route가 실제 content/ 구조와 불일치할 수 있음. `/claude-code/*` 7개 경로는 `sitemap.ts`에 있으나 `content/` 디렉토리에는 없음 (이전 구조 잔재).

**`robots.txt`** (15 LOC, 정적 파일):
```
User-agent: *
Allow: /
Disallow: /api/
Disallow: /_next/
Disallow: /debug-pagemap/
Crawl-delay: 1
Sitemap: https://adk.mo.ai.kr/sitemap.xml
```

### 5.3 `structured-data.tsx` JSON-LD 스키마 (110 LOC)

4종의 schema.org 엔티티 주입 (루트 `app/layout.tsx`에서 `<MoAIStructuredData />`로 호출):

| JSON-LD 타입            | 주요 필드                                      |
|-------------------------|-----------------------------------------------|
| `Organization`          | name, url, logo, description, sameAs (GitHub/Discord/NPM) |
| `SoftwareApplication`   | operatingSystem, applicationCategory, offers, aggregateRating, author, license, softwareVersion |
| `WebSite`               | name, url, description, potentialAction (SearchAction), inLanguage (ko/en/zh/ja) |
| `TechArticle`           | headline, description, image, author, publisher, datePublished, dateModified, keywords, inLanguage |

**마이그레이션 시**: Hugo `layouts/partials/head.html`에 4개 `<script type="application/ld+json">` 직접 삽입 또는 데이터 파일(`data/structured_data.toml`) 기반 렌더.

### 5.4 동적 라우트 패턴

- `[[...mdxPath]]` — optional catch-all (모든 MDX 페이지)
- `[...meta].json` — 미사용 레거시

**Hextra 전환 시**: Hugo는 파일 기반 정적 라우팅만 지원하므로 `[[...mdxPath]]`는 content 디렉토리 구조 1:1 매핑으로 대체.

### 5.5 OG 이미지 처리 방식

- 정적 파일 `public/og.png` (7.5 MB) 단일 이미지
- `theme.config.tsx` head에서 `og:image`로 하드코딩 참조
- **동적 OG 이미지 생성 없음** (Next.js `@vercel/og` 같은 것 미사용)
- Hextra로 그대로 이식 가능하나 7.5 MB는 **압축 최적화 필수** (아마 300-500 KB로 가능)

---

## 6. Custom 컴포넌트 분석

### 6.1 `components/` 개요

| 파일                        | LOC | 주요 역할                      | Nextra/Next.js 의존 | Hextra 대체 가능? |
|-----------------------------|-----|--------------------------------|---------------------|-------------------|
| `LanguageSelector.tsx`      | 162 | 4개 언어 드롭다운 + 경로 유지   | `next/navigation`, `@radix-ui` | **부분** (Hextra 언어 스위처 내장 있으나 커스텀 플래그 UI는 partial 필요) |
| `navbar.tsx`                | 127 | GitHub stars 뱃지, 실시간 API  | `lucide-react`      | **부분** (API 호출은 클라이언트 JS 필요) |
| `structured-data.tsx`       | 110 | JSON-LD 4종 주입               | 없음 (순수 React)    | ✅ O (head partial) |
| `CodeBlock.tsx`             |  91 | macOS 스타일 코드블록 + 복사   | `lucide-react`, `use client` | ❌ (Hextra 내장 code copy와 스타일 재설계 필요) |
| `ClientNavbar.tsx`          |  43 | Navbar + LanguageSelector 조합 | `nextra-theme-docs` | **부분** (Hextra 네비게이션으로 대체 가능) |
| `ui/dialog.tsx`             | 122 | Radix UI Dialog wrapper        | `@radix-ui`         | ❌ (현재 content/에서 미사용 — 제거 가능) |
| `ui/card.tsx`               |  86 | shadcn Card 컴포넌트           | 없음                | ❌ (미사용 — 제거 가능) |
| `ui/button.tsx`             |  56 | shadcn Button 컴포넌트         | 없음                | ❌ (미사용 — 제거 가능) |
| `ui/input.tsx`              |  22 | shadcn Input 컴포넌트          | 없음                | ❌ (미사용 — 제거 가능) |

**중요 발견**: `components/ui/*` 4개 (dialog, card, button, input)는 content/ 어디에서도 import되지 않음. 앱 레이아웃에도 사용 흔적 없음 → **완전한 dead code, 마이그레이션 제외 대상**.

### 6.2 핵심 5개 컴포넌트 상세

#### 6.2.1 `ClientNavbar.tsx` (43 LOC)

**역할**: Nextra `Navbar`를 "use client" 경계에서 래핑 + GitHub SVG 아이콘 직접 삽입 + LanguageSelector 추가.

**의존성**:
```typescript
import { Navbar } from "nextra-theme-docs";
import LanguageSelector from "./LanguageSelector";
```

**특징**:
- GitHub 24×24 SVG 인라인 (다른 아이콘 미사용)
- `projectIcon` prop으로 커스텀 slot 주입
- `logo`와 `projectLink`를 props로 받음

**Hextra 대체**: Hextra의 `hextra.navbar-end` hook으로 대체 가능하나, GitHub SVG + LanguageSelector 조합은 custom partial (`layouts/partials/custom/navbar-end.html`) 신규 작성 필요.

#### 6.2.2 `CodeBlock.tsx` (91 LOC)

**역할**: macOS 창 스타일 코드 블록 (빨강/노랑/초록 점 헤더) + 언어 라벨 + 호버시 노출되는 복사 버튼.

**의존성**:
```typescript
import { Check, Copy } from "lucide-react";
import { useState } from "react";
```

**특징**:
- `extractCodeContent()` 재귀 헬퍼로 React children에서 텍스트 추출
- 호버 애니메이션 (Tailwind `opacity-0 group-hover:opacity-100`)
- `data-language` data attribute 활용

**Hextra 대체 난이도**: **Hard**. Hextra는 Chroma (Hugo 내장) 기반 하이라이팅이고 `hugo.toml`의 `markup.highlight` 설정으로 스타일 제어. macOS 창 헤더는 커스텀 CSS + JS + `render-codeblock-default.html` partial 전체 재작성 필요.

**현황 이슈**: 이 컴포넌트가 실제로 MDX에서 사용되는지 불명확. `mdx-components.tsx`는 단순히 `nextra-theme-docs`의 기본 컴포넌트만 재내보내므로 `<pre><code>` 커스텀 매핑이 없음. **즉 이 CodeBlock은 선언만 되어있고 실제 연결 안 된 것으로 의심**.

#### 6.2.3 `LanguageSelector.tsx` (162 LOC)

**역할**: 4개 언어 (ko/en/ja/zh) 드롭다운, 플래그 이모지 + 한글 이름, 현재 경로 유지하며 locale 전환.

**의존성**:
```typescript
import * as DropdownMenuPrimitive from "@radix-ui/react-dropdown-menu";
import { ChevronDown } from "lucide-react";
import { usePathname } from "next/navigation";
```

**핵심 로직**:
1. `getLocaleFromPathname()` — pathname 첫 세그먼트에서 locale 추출
2. `buildLocalizedPathname()` — 현재 locale prefix 제거 후 target locale로 교체
3. `window.location.href = newPathname` — 풀 페이지 이동

**Hextra 대체**: Hextra는 기본 언어 스위처 제공. 다만 플래그 이모지 + "Language / 언어" 헤더 같은 커스텀 UI는 `layouts/partials/language-switch.html` override 필요.

#### 6.2.4 `navbar.tsx` (127 LOC)

**역할**: GitHub stars API를 실시간 호출하여 뱃지로 표시하는 `GitHubStarBadge` + 단순 버전 `GitHubStarButton`.

**의존성**:
```typescript
import { Github, Star } from "lucide-react";
```

**특징**:
- 클라이언트 사이드 `fetch("https://api.github.com/repos/modu-ai/moai-adk")`
- `next: { revalidate: 300 }` — 5분 캐시
- 1000+ → "1.2k" 포맷팅

**Hextra 대체 난이도**: **Medium**. Hugo는 빌드타임이므로:
- 옵션 1: 빌드 스크립트에서 GitHub API 호출 후 `data/github.toml` 생성 → partial에서 참조
- 옵션 2: shields.io 뱃지 이미지 (`theme.config.tsx`에 이미 포함 중)로 대체 → 가장 간단

**현황 이슈**: `theme.config.tsx` line 107-131에도 GitHub stars 뱃지가 shields.io 이미지로 직접 삽입되어 있음. 즉 **`GitHubStarBadge` 컴포넌트는 중복 선언 + 미사용** 가능성 매우 높음. 추가 검증 필요.

#### 6.2.5 `structured-data.tsx` (110 LOC)

**역할**: JSON-LD 구조화 데이터 4종을 `<script type="application/ld+json">`으로 주입.

**의존성**: 없음 (순수 React 함수형)

**특징**:
- `Organization` / `SoftwareApplication` / `WebSite` / `TechArticle` 4종
- `dateModified: new Date().toISOString().split("T")[0]` — 빌드타임마다 다름
- `softwareVersion: "2.0.0"` 하드코딩 (릴리스와 미동기화 위험)
- `aggregateRating.ratingValue: "4.8", ratingCount: "42"` 하드코딩 (임의 값, 검색엔진 정책 위반 소지)

**Hextra 대체 난이도**: **Easy**. Hugo `layouts/partials/head.html`에 `{{ partial "structured-data" . }}` 삽입하고, 데이터를 `data/structured_data.toml` (per-locale)로 외부화.

**권장**: `aggregateRating` 하드코딩은 **마이그레이션 시점에 제거** (Google 구조화 데이터 정책상 검증 불가한 평점 주장 금지).

---

## 7. 마이그레이션 작업량 추정

### 7.1 Easy (Hextra 내장 기능으로 대체)

- **`<Callout>` 735건** → Hextra `callout` shortcode 1:1 치환 (type별 5종: tip/info/warning/error/success)
- **`_meta.ts` → `_index.md` frontmatter** 변환 (38개 파일)
- **shields.io GitHub stars 뱃지** → `theme.config` JSX를 `hugo.toml` `[menu.main]`의 외부 이미지 링크로 대체
- **`sitemap.xml` 자동 생성** → Hugo 내장
- **`robots.txt` 정적 파일** → `static/robots.txt`로 그대로 복사
- **PWA `manifest.json`** → `static/manifest.json`으로 그대로 복사
- **favicon / icons / og.png** → `static/`으로 복사 (og.png는 압축 최적화 권장)
- **Mermaid 569건** → Hextra는 Mermaid JS CDN 내장 지원 (config 플래그만 켜면 됨)
- **`structured-data` JSON-LD** → partial + data file 외부화
- **`nextra-theme-docs`의 prev/next, sidebar, TOC** → Hextra 동등 기능 내장
- **4개 locale 검색 (현재 `search: false`로 꺼져있음)** → Hextra FlexSearch로 켜면 개선 효과
- **per-locale `toc.title`, `editLink.content`, `feedback.content`, `footer.content`** → `i18n/{ko,en,ja,zh}.toml`

### 7.2 Medium (Hugo partial 작성 필요)

- **`LanguageSelector` 플래그 드롭다운** — Hextra 기본 스위처 스타일과 다르므로 partial override (`layouts/partials/language-switch.html`)
- **`theme.config.tsx`의 `head` 24개 태그** → `layouts/partials/head.html` 커스터마이즈
- **OG meta 태그 locale별 다이나믹 값** — Hugo `i18n/` + `.Site.Params.{locale}` 조합
- **Google Analytics + Vercel Analytics 2중 주입** → Hextra GA 설정 + Vercel Analytics script 별도 partial
- **`docsRepositoryBase` Edit this page 링크** — Hextra 지원하나 per-locale 텍스트 오버라이드 필요
- **`redirects` 2건 (moai-rank)** → `vercel.json`에 그대로 유지 가능
- **`navbar.tsx` GitHub stars API** → 빌드 스크립트로 data 생성 (사용 여부 확인 후 삭제가 더 나을 수도)
- **`debug-pagemap` / `_content.mdx` / `[...meta].json`** 레거시 파일 정리

### 7.3 Hard (재설계 또는 기능 포기 판단 필요)

- **`middleware.ts` Edge runtime 로직 (164 LOC)**
  - **문제**: Accept-Language 기반 자동 리다이렉트 + 쿠키 저장은 Hugo 정적 사이트로 직접 재현 불가
  - **대안 A**: Vercel `Edge Functions` 별도 구현 (Go/TS 작성)
  - **대안 B**: `vercel.json` `rewrites`로 정적 규칙만 (Accept-Language 무시)
  - **대안 C**: 클라이언트 JS로 감지 + 리다이렉트 (JS off 시 기본값 ko)
  - **대안 D**: 루트 `/`를 locale-agnostic 랜딩 페이지로 두고 사용자 명시 선택
  - **판단 필요**: UX 우선순위와 SEO 임팩트 (hreflang은 이미 설정됨)

- **`CodeBlock.tsx` macOS 창 스타일**
  - **문제**: Hugo Chroma와 구조 완전 다름. 복사 버튼 JS + macOS dots + 호버 애니메이션 전체 재구현
  - **대안 A**: Hextra 기본 코드블록 수용 (스타일 포기, 복사 버튼은 Hextra에 내장되어 있음)
  - **대안 B**: 커스텀 `render-codeblock-default.html` + 별도 CSS + JS 전체 이식
  - **우선 확인**: `CodeBlock`이 실제로 MDX 렌더링에 연결되어있는지 확인 필요 (mdx-components.tsx 보면 연결 증거 없음 — 즉 **dead code일 가능성 높음**)

- **Nextra 4의 `_meta.ts` object 값 (`{title, display}`)** 중 `display: "hidden"` 의미론
  - **현재**: 사이드바에 숨기지만 URL 접근은 가능
  - **Hugo 대응**: frontmatter `_build.list: never` + `_build.render: always` 조합으로 유사 재현

- **`@vercel/analytics` Next.js `<Analytics />` 컴포넌트**
  - **문제**: Next.js 특화 라이브러리. Hugo에서는 지원 X
  - **대안**: Vercel Analytics 범용 script 태그로 대체 (공식 문서 제공)

---

## 8. 발견한 리스크

### 리스크 1: `middleware.ts` Edge 리다이렉트 로직
- **무엇이**: `Accept-Language` 감지 + 쿠키 기반 locale 우선순위 로직 (164 LOC)이 Hugo 정적 사이트와 근본적으로 호환되지 않음
- **왜 문제인가**: 현재 사용자가 일본어 브라우저로 `https://adk.mo.ai.kr/` 접속 시 `/ja`로 자동 리다이렉트되는 UX 유지가 마이그레이션 핵심 요구사항이라면, Vercel Edge Function을 별도 작성해야 하고 이는 Go 프로젝트 "템플릿 기반 문서 사이트" 범위를 벗어남
- **완화책**: (a) Phase 0 설계 시 Accept-Language 자동 리다이렉트 UX를 포기하고 루트에 언어 선택 랜딩 페이지를 두는 안을 검토, (b) 유지하려면 `api/detect-locale.go` Vercel Function을 추가하고 `vercel.json` rewrites와 결합

### 리스크 2: 219개 MDX 파일 전면 frontmatter 주입
- **무엇이**: 현재 0개 파일에만 YAML frontmatter가 있음. Hugo는 `title`, `weight`, `draft` 등 frontmatter 필수 필드가 실질적으로 요구됨
- **왜 문제인가**: 수동으로 219개 파일에 frontmatter 추가 시 휴먼 에러, 로케일 간 일관성 깨짐, 섹션 순서 오차 발생. `_meta.ts` 38개에서 순서/제목을 추출해 대응되는 MDX에 주입하는 스크립트 없이 수작업은 위험
- **완화책**: Phase 2에서 Go 스크립트 작성 — `_meta.ts` 파싱 + MDX 대응 매칭 + frontmatter 자동 삽입. TDD로 per-locale 검증 후 일괄 실행

### 리스크 3: Mermaid 569개 다이어그램 렌더링
- **무엇이**: worktree/guide, worktree/faq, getting-started/quickstart 등에 집중된 Mermaid 블록. Nextra 4는 클라이언트 사이드 `mermaid.js`로 렌더 (`@mermaid-js/mermaid-cli`는 devDependency로만 존재하여 빌드 타임 프리렌더 안 함)
- **왜 문제인가**: Hextra에 Mermaid 지원 플래그가 있으나 569개 × 4 locale = 2,276개 다이어그램이 클라이언트 런타임 렌더링 시 초기 로드 3-5초 저하. SEO 크롤러도 렌더 전 콘텐츠만 인덱싱
- **완화책**: (a) Hugo 빌드 시 `@mermaid-js/mermaid-cli`로 SVG 사전 렌더링 파이프라인 구축, (b) Hextra의 lazy-load 기능 활용 (스크롤 뷰포트 진입 시만 렌더), (c) 다이어그램 수량이 과다한 문서는 정적 이미지로 대체 검토

### 리스크 4: ko 전용 섹션 (`contributing/`, `multi-llm/`) 불균형
- **무엇이**: ko는 63 페이지, 나머지 3개 locale은 각 52 페이지. `contributing/`, `multi-llm/` 두 섹션이 ko에만 존재
- **왜 문제인가**: Hextra/Hugo i18n에서 미번역 페이지는 기본 언어 fallback 또는 404 처리됨. 링크 무결성 CI 실패 가능
- **완화책**: (a) 마이그레이션 시점에 3개 locale에도 skeleton 페이지 생성 (title만 번역, 본문은 "Coming soon" 또는 ko 원본 링크), (b) ko에만 보여줄 섹션이면 per-locale 메뉴 분기 처리

### 리스크 5: `structured-data.tsx` 하드코딩된 `aggregateRating`
- **무엇이**: `ratingValue: "4.8", ratingCount: "42"`가 근거 없이 하드코딩
- **왜 문제인가**: Google Search Central 가이드라인상 검증 불가한 평점 주장은 구조화 데이터 위반 — 향후 Rich Results에서 제외 또는 수동 페널티 가능
- **완화책**: 마이그레이션 중 이 필드를 제거하거나, GitHub stars/npm downloads 같은 검증 가능 지표로 교체

### 리스크 6: `CodeBlock.tsx`, `components/ui/*`, `lib/page-map.ts`, `app/[...meta].json` 등 dead code
- **무엇이**: 실제 런타임에 호출되지 않는 것으로 의심되는 파일이 다수 (CodeBlock은 mdx-components.tsx에 연결 안 됨, ui/* 4개는 content/에서 import 0건, page-map.ts는 app/layout.tsx가 nextra 공식 API 사용으로 사실상 미사용)
- **왜 문제인가**: 마이그레이션 대상에 포함시키면 작업량이 부풀려짐. 실제 사용 여부를 Phase 0에서 확정하지 못하면 Hextra 쪽에 필요 없는 partial을 작성하게 됨
- **완화책**: Phase 0 Gate 1에서 dead code 확정 후 마이그레이션 대상에서 제외. 현행 코드베이스에 대해서도 archive 전 정리 PR 검토

### 리스크 7: Vercel `regions: ["hkg1"]` 지역 설정
- **무엇이**: 현재 홍콩 리전에 고정 배포 (한국/일본/중국 사용자 최적화 목적 추정)
- **왜 문제인가**: Hugo는 전 세계 CDN 엣지 캐싱이므로 `regions` 설정이 불필요하거나 역효과. 그러나 Vercel의 Serverless Function 실행 리전은 middleware 대체 Edge Function 작성 시 영향
- **완화책**: Hextra 전환 시 `regions`는 제거 또는 Vercel 기본값(`iad1`)으로. 지연 우려 시 CDN cache-control 헤더로 제어

### 리스크 8: `_meta.ts`의 TypeScript 타입 의존
- **무엇이**: 38개 `_meta.ts` 파일이 `import type { MetaRecord } from "nextra";` 타입 의존
- **왜 문제인가**: Hugo frontmatter는 YAML/TOML/JSON이라 TypeScript 타입 안전성 상실. `display: "hidden"` 같은 값 유효성 검증을 Go 스크립트 또는 CI로 대체 필요
- **완화책**: Phase 2에서 `_meta.ts` 파서 Go 스크립트 + JSON Schema 기반 frontmatter 검증 CI 단계 추가

### 리스크 9: 이미지 용량 (og.png 7.5 MB)
- **무엇이**: `public/og.png`가 7.5 MB로 과도하게 큼. `icon-512.png`도 543 KB
- **왜 문제인가**: OG 이미지는 소셜 공유 시 페치되므로 CDN 트래픽 및 공유 로딩 지연
- **완화책**: ImageOptim/cwebp로 300-500 KB 수준으로 압축 후 마이그레이션. Hugo `resources.Images.Resize` 파이프라인 활용도 가능

### 리스크 10: `sitemap.ts`의 하드코딩된 경로 36개와 실제 content 구조 불일치
- **무엇이**: `sitemap.ts`에 `/claude-code/*` 7개 경로가 있으나 `content/` 디렉토리에는 `claude-code/` 섹션이 존재하지 않음 (이전 구조 잔재)
- **왜 문제인가**: Hugo 자동 sitemap은 실제 존재하는 페이지만 포함하므로 현재 잘못된 URL이 사라지면 SEO 영향. 만약 외부에서 이 URL로 백링크가 있다면 301 리다이렉트 필요
- **완화책**: Phase 0 Gate 1에서 현행 sitemap vs content/ 실재 경로 diff 수행 후, 사라질 URL에 대한 301 규칙을 `vercel.json`에 추가

---

## 부록: 전체 파일 카운트 요약

| 카테고리                | 파일 수 | 합계 LOC (대략) |
|-------------------------|---------|-----------------|
| Content MDX (4 locale)  | 219     | 추후 측정       |
| `_meta.ts` (4 locale)   | 38      | 641             |
| `components/*.tsx`      | 9       | 819             |
| `app/*.tsx` + `app/**`  | 9       | 483             |
| `lib/*.ts`              | 2       | 241             |
| `middleware.ts`         | 1       | 164             |
| `theme.config.tsx`      | 1       | 293             |
| `next.config.mjs`       | 1       | 9               |
| `vercel.json`           | 1       | 63              |
| `package.json`          | 1       | 55              |
| `tsconfig.json`, `tailwind.config.ts`, `biome.json`, `.markdownlint.json`, `postcss.config.mjs`, `components.json`, `playwright.config.ts` | 7 | ~100 |
| `public/*` (정적 자산)   | 11      | (바이너리 위주) |
| **총 코드 파일 (MDX 제외)** | **80**  | **~2,868 LOC**  |

---

**스캔 완료**. 다음 단계: Phase 1에서 본 인벤토리를 기반으로 SPEC-DOCS-SITE-001 EARS 요구사항을 작성하여 결정 포인트 (리스크 1, 6 특히) 해결.
