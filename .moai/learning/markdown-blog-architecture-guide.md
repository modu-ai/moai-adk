# GitHub + Vercel 마크다운 블로그 완벽 가이드

**생성일**: 2025-11-22
**작성자**: Yoda - Technical Depth Expert
**난이도**: 초급 → 고급

---

## 📑 목차

1. [전제 조건](#전제-조건)
2. [핵심 개념](#핵심-개념)
3. [기술 스택 선택](#기술-스택-선택)
4. [실제 구현 예시](#실제-구현-예시)
5. [댓글 시스템 통합](#댓글-시스템-통합)
6. [성능 최적화](#성능-최적화)
7. [배포 전략](#배포-전략)
8. [실습 연습](#실습-연습)
9. [추가 학습](#추가-학습)
10. [요약 체크리스트](#요약-체크리스트)

---

## 전제 조건

이 가이드를 학습하기 전에 다음을 이해하고 있어야 합니다:

- [ ] Git 기본 명령어 (`git add`, `git commit`, `git push`)
- [ ] GitHub 계정 및 레포지토리 생성 방법
- [ ] Vercel 계정 (무료) 생성
- [ ] Node.js 18+ 설치
- [ ] 마크다운 기본 문법
- [ ] 터미널/커맨드라인 사용법

---

## 핵심 개념

### JAMstack 아키텍처란?

**JAMstack** = **J**avaScript + **A**PIs + **M**arkup

```
전통적인 블로그 (WordPress)          JAMstack 블로그 (Astro/Next.js)
┌─────────────────────────┐         ┌─────────────────────────┐
│  매 요청마다:            │         │  빌드 타임에 한 번만:    │
│  1. DB 쿼리 (MySQL)     │         │  1. 마크다운 파일 읽기   │
│  2. PHP 템플릿 렌더링   │         │  2. HTML 생성           │
│  3. HTML 동적 생성      │         │  3. CDN에 배포          │
│  4. 사용자에게 전달     │         │                        │
│  = 500-2000ms          │         │  런타임에:             │
│                        │         │  1. CDN에서 HTML 제공  │
└─────────────────────────┘         │  = 100-300ms ⚡        │
                                    └─────────────────────────┘
```

### 왜 데이터베이스가 필요 없는가?

**핵심 원리 4가지**:

1. **컨텐츠 = 파일 시스템**
   ```
   content/blog/
   ├── post-1.md  ← 이것이 DB의 "row"
   ├── post-2.md
   └── post-3.md

   각 마크다운 파일의 frontmatter = DB의 "column"
   ---
   title: "블로그 제목"      ← DB의 "title" 컬럼
   pubDate: 2025-01-15      ← DB의 "created_at" 컬럼
   author: "GOOS"           ← DB의 "author" 컬럼
   tags: ["tech", "web"]    ← DB의 "tags" 관계
   ---
   ```

2. **Git = 버전 관리 시스템**
   ```bash
   # DB의 version history 대신
   git log content/blog/post-1.md

   # 이전 버전 복구 대신
   git checkout HEAD~3 -- content/blog/post-1.md
   ```

3. **빌드 타임 생성 = 쿼리 결과 캐싱**
   ```
   DB 쿼리 (매번):                    빌드 타임 (한 번만):
   SELECT * FROM posts               const posts = await getCollection('blog');
   WHERE draft = false               → HTML 파일 생성
   ORDER BY pubDate DESC             → CDN에 업로드
   ↓                                 ↓
   500ms 쿼리 시간                    0.1ms 파일 읽기 시간
   ```

4. **CDN = 전역 캐시**
   ```
   사용자(서울) → Vercel Edge(서울) → 즉시 HTML 제공 (10ms)
   사용자(뉴욕) → Vercel Edge(뉴욕) → 즉시 HTML 제공 (10ms)

   vs DB 방식:
   사용자(서울) → 서버(미국) → DB 쿼리 → 응답 (500ms)
   ```

---

## 기술 스택 선택

### 의사결정 트리

```
내 블로그의 주요 목적은?
├─ 순수 콘텐츠 발행 (글쓰기 중심)
│  └─ React 컴포넌트 재사용 필요?
│     ├─ 아니오 → Astro 4.x ✅ (최고의 성능)
│     └─ 예 → Next.js 15 (React 생태계)
│
└─ 인터랙티브 기능 필요 (대시보드, 로그인 등)
   └─ Next.js 15 ✅ (API Routes + React)
```

---

## 1순위 추천: Astro 4.x

### 핵심 장점

| 특징 | Astro 4.x | Next.js 15 | Gatsby 5 |
|------|-----------|------------|----------|
| **JavaScript 번들** | 0-10KB ⭐⭐⭐⭐⭐ | 150KB ⭐⭐⭐ | 180KB ⭐⭐ |
| **빌드 속도** (100개 포스트) | 5초 ⭐⭐⭐⭐⭐ | 25초 ⭐⭐⭐⭐ | 45초 ⭐⭐⭐ |
| **Lighthouse 점수** | 100 ⭐⭐⭐⭐⭐ | 95-98 ⭐⭐⭐⭐ | 92-95 ⭐⭐⭐ |
| **학습 곡선** | 쉬움 ⭐⭐⭐⭐⭐ | 중간 ⭐⭐⭐ | 어려움 ⭐⭐ |
| **마크다운 DX** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ |

### Astro의 설계 철학

**"Zero JavaScript by Default"** - 기본적으로 JavaScript를 보내지 않음

```astro
---
// src/pages/blog/[slug].astro
import { getCollection, render } from 'astro:content';

export async function getStaticPaths() {
  const posts = await getCollection('blog');
  return posts.map(post => ({
    params: { slug: post.id },
    props: { post },
  }));
}

const { post } = Astro.props;
const { Content } = await render(post);
---

<!-- ✅ 이 페이지는 순수 HTML + CSS만 전송 -->
<article class="prose">
  <h1>{post.data.title}</h1>
  <Content />
</article>

<!-- ❌ 일반적인 React 블로그는 150KB+ JavaScript 전송 -->
```

### Content Collections - 타입 안전 마크다운

```typescript
// src/content/config.ts - Zod 스키마 정의
import { defineCollection, z } from 'astro:content';

const blog = defineCollection({
  schema: z.object({
    title: z.string(),
    description: z.string(),
    pubDate: z.coerce.date(),
    author: z.string(),
    tags: z.array(z.string()),
    draft: z.boolean().default(false),
    image: z.object({
      url: z.string(),
      alt: z.string()
    }).optional()
  })
});

export const collections = { blog };
```

**타입 안전성의 장점**:

```typescript
// ❌ 일반 마크다운 파싱 (타입 없음)
const post = await readMarkdown('post.md');
post.frontmatter.title  // string | undefined (오타 가능)
post.frontmatter.pubDte // ❌ 오타 발견 못함

// ✅ Astro Content Collections (타입 안전)
const post = await getEntry('blog', 'my-post');
post.data.title  // string (자동 완성 ✨)
post.data.pubDte // ❌ 타입스크립트 오류: Property 'pubDte' does not exist
```

**빌드 타임 유효성 검증**:

```markdown
<!-- content/blog/invalid-post.md -->
---
title: 123  ❌ 숫자인데 string 기대
pubDate: "invalid date"  ❌ 잘못된 날짜 형식
tags: "single-string"  ❌ 배열 필요
---

# 내용

→ 빌드 실패!
Error: [blog] post-1.md frontmatter does not match collection schema.
  title: Expected string, received number.
  pubDate: Invalid date.
  tags: Expected array, received string.
```

---

## 2순위 대안: Next.js 15 (App Router)

### 언제 Next.js가 더 나은가?

**Next.js를 선택해야 하는 경우**:

1. **React 컴포넌트 재사용**
   ```tsx
   // 블로그 + 랜딩 페이지 + 대시보드에서 동일한 컴포넌트 사용
   import { Button } from '@/components/ui/button';

   // Astro에서는 React 컴포넌트 사용이 제한적
   ```

2. **향후 인터랙티브 기능 계획**
   ```tsx
   // 예: 사용자 대시보드, 로그인, 실시간 업데이트
   'use client'  // Next.js에서 클라이언트 컴포넌트

   export function UserDashboard() {
     const [data, setData] = useState();
     // React 생태계의 모든 것 사용 가능
   }
   ```

3. **API Routes 필요**
   ```typescript
   // app/api/subscribe/route.ts
   export async function POST(request: Request) {
     const { email } = await request.json();
     // 뉴스레터 구독 처리
     await subscribeToNewsletter(email);
     return Response.json({ success: true });
   }
   ```

### Next.js 15 구현 예시

```typescript
// app/blog/[slug]/page.tsx
import { getPostBySlug, getAllPosts } from '@/lib/mdx';
import { MDXRemote } from 'next-mdx-remote/rsc';

export async function generateStaticParams() {
  const posts = await getAllPosts();
  return posts.map(post => ({ slug: post.slug }));
}

export default async function BlogPost({
  params
}: {
  params: { slug: string }
}) {
  const post = await getPostBySlug(params.slug);

  return (
    <article className="prose">
      <h1>{post.frontmatter.title}</h1>
      <MDXRemote source={post.content} />
    </article>
  );
}
```

---

## 실제 구현 예시

### Astro 블로그 완전한 구조

```
my-astro-blog/
├── src/
│   ├── content/
│   │   ├── config.ts          # Content Collections 스키마
│   │   └── blog/
│   │       ├── post-1.md      # 블로그 포스트
│   │       └── post-2.md
│   ├── layouts/
│   │   └── BlogLayout.astro   # 블로그 레이아웃
│   ├── pages/
│   │   ├── index.astro        # 홈페이지
│   │   ├── blog/
│   │   │   ├── index.astro    # 블로그 목록
│   │   │   └── [slug].astro   # 블로그 포스트 동적 페이지
│   │   ├── rss.xml.ts         # RSS 피드
│   │   └── sitemap.xml        # (자동 생성)
│   └── styles/
│       └── global.css
├── public/
│   └── images/
├── astro.config.mjs
├── tailwind.config.js
└── package.json
```

### 블로그 포스트 예시

```markdown
---
# content/blog/my-first-post.md
title: "Astro로 블로그 만들기"
description: "초고속 정적 사이트 생성기 Astro 소개"
pubDate: 2025-01-15
author: "GOOS"
tags: ["astro", "web-development", "performance"]
image:
  url: "/images/astro-intro.jpg"
  alt: "Astro 로고"
draft: false
---

# Astro로 블로그 만들기

Astro는 **Zero JavaScript by Default** 철학을 가진 현대적인 정적 사이트 생성기입니다.

## 주요 특징

1. **빠른 성능**: 100/100 Lighthouse 점수
2. **Content Collections**: 타입 안전한 마크다운 관리
3. **Island Architecture**: 필요한 곳에만 JavaScript

```javascript
// 예제 코드
const greeting = "Hello, Astro!";
console.log(greeting);
```

![Astro Architecture](/images/astro-arch.png)
```

### 블로그 포스트 페이지 구현

```astro
---
// src/pages/blog/[slug].astro
import { getCollection, render } from 'astro:content';
import BlogLayout from '../../layouts/BlogLayout.astro';

export async function getStaticPaths() {
  const posts = await getCollection('blog', ({ data }) => !data.draft);
  return posts.map(post => ({
    params: { slug: post.id },
    props: { post },
  }));
}

const { post } = Astro.props;
const { Content, headings } = await render(post);
---

<BlogLayout
  title={post.data.title}
  description={post.data.description}
  pubDate={post.data.pubDate}
  author={post.data.author}
>
  <article class="prose prose-lg max-w-4xl mx-auto px-4 py-12">
    <!-- 헤더 -->
    <header class="mb-8">
      <h1 class="text-4xl font-bold mb-2">{post.data.title}</h1>
      <div class="text-gray-600 text-sm">
        {post.data.pubDate.toLocaleDateString('ko-KR')} · {post.data.author}
      </div>
      <div class="flex gap-2 mt-4">
        {post.data.tags.map(tag => (
          <span class="px-3 py-1 bg-blue-100 text-blue-800 text-sm rounded-full">
            {tag}
          </span>
        ))}
      </div>
    </header>

    <!-- 본문 (마크다운 컨텐츠) -->
    <Content />

    <!-- 댓글 섹션 (Giscus) -->
    <div class="mt-12 border-t pt-8">
      <h2 class="text-2xl font-bold mb-4">댓글</h2>
      <script
        src="https://giscus.app/client.js"
        data-repo="GOOS/my-blog"
        data-repo-id="R_kgDOH..."
        data-category="Comments"
        data-category-id="DIC_kwDOH..."
        data-mapping="pathname"
        data-reactions-enabled="1"
        data-theme="light"
        data-lang="ko"
        crossorigin="anonymous"
        async>
      </script>
    </div>
  </article>
</BlogLayout>
```

### 블로그 목록 페이지

```astro
---
// src/pages/blog/index.astro
import { getCollection } from 'astro:content';
import BaseLayout from '../../layouts/BaseLayout.astro';

const allPosts = await getCollection('blog', ({ data }) => !data.draft);
const sortedPosts = allPosts.sort((a, b) =>
  b.data.pubDate.valueOf() - a.data.pubDate.valueOf()
);
---

<BaseLayout title="블로그" description="모든 블로그 포스트">
  <div class="max-w-4xl mx-auto px-4 py-12">
    <h1 class="text-4xl font-bold mb-8">블로그</h1>

    <div class="grid gap-8">
      {sortedPosts.map(post => (
        <article class="border-b pb-6 hover:bg-gray-50 transition p-4 rounded">
          <a href={`/blog/${post.id}/`} class="group">
            <h2 class="text-2xl font-bold mb-2 group-hover:text-blue-600">
              {post.data.title}
            </h2>
            <p class="text-gray-600 mb-2">{post.data.description}</p>
            <div class="text-sm text-gray-500 mb-2">
              {post.data.pubDate.toLocaleDateString('ko-KR')} · {post.data.author}
            </div>
            <div class="flex gap-2">
              {post.data.tags.map(tag => (
                <span class="px-2 py-1 bg-gray-100 text-sm rounded">
                  #{tag}
                </span>
              ))}
            </div>
          </a>
        </article>
      ))}
    </div>
  </div>
</BaseLayout>
```

---

## 댓글 시스템 통합

### Giscus (GitHub Discussions) - 추천 ⭐

**장점**:
- ✅ **완전 무료** (GitHub 인프라 사용)
- ✅ **GDPR 준수** (프라이버시 친화적)
- ✅ **마크다운 지원** (코드 블록, 이미지 등)
- ✅ **GitHub 계정으로 로그인** (별도 가입 불필요)
- ✅ **이메일 알림** (GitHub 알림 시스템)
- ✅ **이모지 반응** (👍 ❤️ 😄 등)
- ✅ **모더레이션** (GitHub 권한 시스템)

**설정 방법**:

```bash
# 1. GitHub 레포지토리에서 Discussions 활성화
Settings → Features → Discussions (체크)

# 2. Giscus 앱 설치
https://github.com/apps/giscus 방문 → Install

# 3. 설정 생성
https://giscus.app 방문
→ 레포지토리 입력: GOOS/my-blog
→ Discussion 카테고리 선택: Comments
→ 설정 코드 복사
```

**구현 코드**:

```html
<!-- Astro/Next.js 모두 동일 -->
<script src="https://giscus.app/client.js"
        data-repo="GOOS/my-blog"                 <!-- 내 레포지토리 -->
        data-repo-id="R_kgDOH..."                <!-- Giscus 앱에서 받음 -->
        data-category="Comments"                  <!-- Discussion 카테고리 -->
        data-category-id="DIC_kwDOH..."          <!-- Giscus 앱에서 받음 -->
        data-mapping="pathname"                   <!-- URL 경로로 매핑 -->
        data-strict="0"
        data-reactions-enabled="1"                <!-- 이모지 반응 허용 -->
        data-emit-metadata="0"
        data-input-position="top"                 <!-- 댓글 입력창 위치 -->
        data-theme="light"                        <!-- 테마 -->
        data-lang="ko"                            <!-- 언어 -->
        crossorigin="anonymous"
        async>
</script>
```

### 대안 비교

| 시스템 | 백엔드 | 비용 | GDPR | GitHub 필요 | 추천도 |
|--------|--------|------|------|------------|--------|
| **Giscus** | GitHub Discussions | 무료 | ✅ | 필요 | ⭐⭐⭐⭐⭐ |
| **Utterances** | GitHub Issues | 무료 | ✅ | 필요 | ⭐⭐⭐⭐ |
| **Disqus** | 외부 서버 | 무료/유료 | ❌ (광고, 추적) | 불필요 | ⭐⭐ |
| **Commento** | 자체 서버 | $10/월 | ✅ | 불필요 | ⭐⭐⭐ |

**추천 이유**:
- GitHub Discussions가 Issues보다 댓글에 적합 (스레드, 반응 등)
- Disqus는 광고 + 사용자 추적으로 프라이버시 문제
- 자체 서버는 유지보수 부담

---

## 성능 최적화

### 빌드 타임 최적화

#### 1. Syntax Highlighting (Shiki)

**런타임 방식 (Prism.js)**:
```
마크다운: ```python\ndef hello():\n```
↓
HTML: <pre><code>def hello():</code></pre>
+ Prism.js 로드 (20KB)
↓ 브라우저에서 JavaScript 실행
↓
<pre><code><span class="token keyword">def</span>...</code></pre>
```

**빌드 타임 방식 (Shiki - Astro 기본)**:
```
마크다운: ```python\ndef hello():\n```
↓ 빌드 타임에 변환
↓
HTML:
<pre class="shiki github-dark">
  <span style="color:#C678DD">def</span>
  <span style="color:#61AFEF">hello</span>
  <span style="color:#ABB2BF">()</span>
</pre>
+ 0KB JavaScript ✨
```

**성능 비교**:
- Prism.js: 20KB + 런타임 실행 (50-100ms)
- Shiki: 0KB + 0ms (이미 HTML에 포함)

#### 2. 이미지 최적화

```typescript
// ❌ 최적화 안 됨
<img src="/images/photo.jpg" alt="사진" />
// → 5MB JPEG, 원본 크기

// ✅ Astro Image 컴포넌트
import { Image } from 'astro:assets';
import photo from '../assets/photo.jpg';

<Image
  src={photo}
  alt="사진"
  width={800}
  height={600}
  format="webp"
  quality={80}
/>
// → 자동으로:
//   - 800x600 리사이즈
//   - WebP 변환 (5MB → 200KB)
//   - srcset 생성 (반응형)
//   - lazy loading
```

#### 3. CSS 최적화 (Tailwind)

```javascript
// tailwind.config.js
export default {
  content: ['./src/**/*.{astro,html,js,jsx,md,mdx,ts,tsx}'],
  // ✅ 사용하지 않는 CSS 자동 제거
  // 1000KB → 20KB
};
```

### 런타임 최적화

#### 1. Lazy Loading

```astro
<!-- 댓글은 스크롤 시 로드 -->
<div id="comments">
  <script>
    const observer = new IntersectionObserver((entries) => {
      if (entries[0].isIntersecting) {
        loadGiscus();
        observer.disconnect();
      }
    });
    observer.observe(document.getElementById('comments'));
  </script>
</div>
```

#### 2. Font Optimization

```astro
---
// src/layouts/BaseLayout.astro
---
<head>
  <!-- Pretendard 폰트 최적화 로딩 -->
  <link
    rel="preload"
    href="/fonts/Pretendard-Variable.woff2"
    as="font"
    type="font/woff2"
    crossorigin
  />
</head>
```

### Core Web Vitals 목표

| 지표 | 목표 | Astro 실제 | Next.js 실제 |
|------|------|-----------|-------------|
| **LCP** (최대 콘텐츠풀 페인트) | < 2.5s | 1.1s ✅ | 1.8s ✅ |
| **FID** (첫 입력 지연) | < 100ms | 0ms ✅ | 50ms ✅ |
| **CLS** (누적 레이아웃 이동) | < 0.1 | 0.02 ✅ | 0.05 ✅ |
| **TTFB** (첫 바이트까지의 시간) | < 800ms | 300ms ✅ | 500ms ✅ |

---

## 배포 전략

### Vercel 배포 (권장)

**장점**:
- ✅ **GitHub 통합**: Push하면 자동 배포
- ✅ **Preview 배포**: PR마다 미리보기 URL
- ✅ **Edge CDN**: 전 세계 200+ 지역
- ✅ **무료 플랜**: 개인 블로그 충분
- ✅ **Zero Config**: 설정 거의 불필요

**배포 단계**:

```bash
# 1. Vercel CLI 설치 (선택)
npm install -g vercel

# 2. 프로젝트를 GitHub에 푸시
git init
git add .
git commit -m "Initial commit"
git remote add origin https://github.com/GOOS/my-blog.git
git push -u origin main

# 3. Vercel에서 Import
https://vercel.com/new 방문
→ Import Git Repository
→ GOOS/my-blog 선택
→ Deploy 클릭

# 4. 자동 감지
Framework Preset: Astro (자동 감지)
Build Command: npm run build (자동)
Output Directory: dist (자동)

# 5. 배포 완료!
https://my-blog-goos.vercel.app
```

**커스텀 도메인 설정**:

```bash
# Vercel 대시보드
Project → Settings → Domains
→ 도메인 입력: blog.goos.dev
→ DNS 설정 (Vercel 안내에 따름)
→ 자동 HTTPS 인증서
```

### 환경 변수 관리

```bash
# .env (로컬 개발용 - Git에 커밋하지 않음!)
PUBLIC_SITE_URL=http://localhost:4321
GISCUS_REPO_ID=R_kgDOH...
GISCUS_CATEGORY_ID=DIC_kwDOH...

# Vercel 대시보드에서 설정
Project → Settings → Environment Variables
→ PUBLIC_SITE_URL = https://blog.goos.dev
```

### CI/CD 워크플로우

```
로컬에서 작성                     Vercel 배포
┌────────────────┐               ┌────────────────┐
│ 마크다운 작성   │               │ 자동 빌드      │
│ content/blog/  │   git push    │ Astro build    │
│ new-post.md    │ ───────────>  │ 정적 파일 생성  │
└────────────────┘               │ CDN 배포       │
                                 └────────────────┘
                                         │
                                         ↓
                                 ┌────────────────┐
                                 │ Preview URL    │
                                 │ (PR마다 생성)  │
                                 └────────────────┘
                                         │
                                         ↓ Merge
                                 ┌────────────────┐
                                 │ Production     │
                                 │ blog.goos.dev  │
                                 └────────────────┘
```

---

## 실습 연습

### Exercise 1 - 기초: Astro 블로그 설정

**난이도**: ⭐ 초급

**목표**: Astro 블로그 기본 구조 이해

**실습**:
```bash
# 1. Astro 프로젝트 생성
npm create astro@latest my-blog
# → Template: Blog
# → TypeScript: Yes (Strict)
# → Install dependencies: Yes

cd my-blog

# 2. Tailwind CSS 추가
npx astro add tailwind

# 3. 개발 서버 실행
npm run dev
# → http://localhost:4321

# 4. 첫 번째 블로그 포스트 작성
# src/content/blog/my-first-post.md 생성
```

**자가 평가**:
- [ ] Content Collections의 역할을 이해했는가?
- [ ] Frontmatter와 본문의 차이를 아는가?
- [ ] `npm run build`로 정적 사이트를 생성할 수 있는가?

---

### Exercise 2 - 중급: Giscus 댓글 통합

**난이도**: ⭐⭐ 중급

**목표**: 서버리스 댓글 시스템 통합

**실습**:
```bash
# 1. GitHub 레포지토리 Discussions 활성화
# Settings → Features → Discussions

# 2. Giscus 앱 설치
# https://github.com/apps/giscus

# 3. 설정 생성
# https://giscus.app
# → 레포지토리 입력
# → 설정 코드 복사

# 4. 블로그 포스트 템플릿에 추가
# src/pages/blog/[slug].astro
```

**자가 평가**:
- [ ] GitHub Discussions와 Issues의 차이를 아는가?
- [ ] `data-mapping="pathname"`의 의미를 이해했는가?
- [ ] 댓글이 실제로 GitHub Discussions에 저장되는 것을 확인했는가?

---

### Exercise 3 - 고급: Vercel 배포 + 커스텀 도메인

**난이도**: ⭐⭐⭐ 고급

**목표**: 프로덕션 배포 완료

**실습**:
```bash
# 1. GitHub 레포지토리 생성
git init
git add .
git commit -m "Initial commit"
git branch -M main
git remote add origin https://github.com/GOOS/my-blog.git
git push -u origin main

# 2. Vercel 배포
# https://vercel.com/new
# → Import Git Repository
# → 배포 완료

# 3. 커스텀 도메인 설정 (선택)
# Vercel → Domains
# → blog.goos.dev 추가
# → DNS 설정

# 4. 성능 측정
# https://pagespeed.web.dev
# → 내 블로그 URL 입력
# → Lighthouse 100/100 확인
```

**자가 평가**:
- [ ] Git → GitHub → Vercel 워크플로우를 이해했는가?
- [ ] Preview 배포와 Production 배포의 차이를 아는가?
- [ ] Core Web Vitals가 모두 "Good" 범위인가?

---

## 추가 학습

### 추가 기능 구현

1. **RSS 피드**
   ```typescript
   // src/pages/rss.xml.ts
   import rss from '@astrojs/rss';
   import { getCollection } from 'astro:content';

   export async function GET(context) {
     const posts = await getCollection('blog');
     return rss({
       title: 'GOOS의 블로그',
       description: '웹 개발 이야기',
       site: context.site,
       items: posts.map(post => ({
         title: post.data.title,
         pubDate: post.data.pubDate,
         link: `/blog/${post.id}/`,
       })),
     });
   }
   ```

2. **검색 기능 (Pagefind)**
   ```bash
   npm install -D pagefind

   # package.json
   "scripts": {
     "build": "astro build && npx pagefind --site dist"
   }
   ```

3. **다크 모드**
   ```astro
   <script>
     const theme = localStorage.getItem('theme') || 'light';
     document.documentElement.classList.add(theme);
   </script>
   ```

### 관련 고급 주제

- **SEO 최적화**: Open Graph, JSON-LD, Sitemap
- **Analytics 통합**: Vercel Analytics, Plausible
- **Newsletter**: Buttondown, ConvertKit 통합
- **MDX 고급**: Interactive Components in Markdown
- **Internationalization**: 다국어 블로그

---

## 요약 체크리스트

### 핵심 개념
- [ ] JAMstack 아키텍처의 3가지 요소 (JavaScript, APIs, Markup) 이해
- [ ] 빌드 타임 생성 vs 런타임 생성 차이 설명 가능
- [ ] 왜 데이터베이스가 필요 없는지 4가지 이유 설명 가능

### 기술 스택
- [ ] Astro vs Next.js 선택 기준 이해
- [ ] Content Collections의 타입 안전성 장점 이해
- [ ] Giscus vs Utterances vs Disqus 차이 이해

### 실전 기술
- [ ] Astro 프로젝트 생성 및 설정 가능
- [ ] 마크다운 포스트 작성 (frontmatter + 본문)
- [ ] Giscus 댓글 시스템 통합
- [ ] Vercel에 배포 및 커스텀 도메인 설정

### 성능 최적화
- [ ] Shiki 빌드 타임 신택스 하이라이팅 이해
- [ ] 이미지 최적화 (WebP, lazy loading) 적용
- [ ] Core Web Vitals 100/100 달성 방법 이해

---

**다음 학습 주제**:
- **MDX 고급 기능**: Interactive components in markdown
- **SEO 마스터하기**: Structured data, sitemap, Open Graph
- **Headless CMS 통합**: Notion, Contentful, Sanity와 연동

---

**생성일**: 2025-11-22
**버전**: 1.0
**라이선스**: MIT
**작성자**: Yoda - MoAI-ADK Technical Depth Expert
