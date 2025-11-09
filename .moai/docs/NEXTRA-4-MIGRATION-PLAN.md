# Next.js 16 + Nextra 4.6.0 마이그레이션 계획

**문서 작성일**: 2025-11-10
**대상 프로젝트**: MoAI-ADK Documentation
**현재 버전**: Next.js 14.2.15 + Nextra 3.3.1 (Pages Router)
**목표 버전**: Next.js 16 + Nextra 4.6.0 (App Router)
**배포 환경**: Vercel (production)
**다국어 지원**: 한국어, 영어, 일본어, 중국어 (4개 언어)

---

## 1️⃣ 개요 및 목표

### 마이그레이션 목표
- Next.js 14.2.15 → 16 (최신 stable)
- Nextra 3.3.1 → 4.6.0
- React 18.2.0 → 19.x
- Pages Router → App Router (완전 전환)
- FlexSearch → Pagefind (검색 엔진)
- Turbopack 활성화 (번들러)

### 성공 기준
1. 모든 100+ MDX 파일 정상 렌더링 (4개 언어 모두)
2. 검색 기능 정상 동작 (Pagefind)
3. 빌드 시간 50% 감소 (Turbopack 효과)
4. Core Web Vitals 개선 (LCP < 2.5s)
5. 다운타임 0분 (Vercel 무중단 배포)
6. 기존 URL 구조 유지 (301 리다이렉트 불필요)

### 리스크 요소
| 리스크 | 심각도 | 대응 방안 |
|--------|--------|----------|
| Next.js 16 + Nextra 4.6.0 호환성 미검증 | HIGH | Phase 1에서 완전 검증 후 진행 |
| 다국어 라우팅 변경 (i18n) | MEDIUM | 기존 i18n 구조 유지, 마이그레이션 스크립트 작성 |
| 100+ MDX 파일 일괄 변환 | MEDIUM | 자동 마이그레이션 스크립트 + 수동 검증 |
| Pagefind 인덱싱 실패 | MEDIUM | 빌드 시 인덱싱 검증, fallback 구성 |
| 검색 쿼리 성능 저하 | LOW | Pagefind 최적화 설정 |

---

## 2️⃣ 현재 상태 분석

### 2.1 기술 스택

```
현재 상태 (v14.2.15 + Nextra 3.3.1 Pages Router):
├─ Next.js: 14.2.15
├─ Nextra: 3.3.1
├─ nextra-theme-docs: 3.3.1
├─ React: 18.2.0
├─ React-DOM: 18.2.0
├─ TypeScript: 5.9.3
├─ Tailwind CSS: 3.4.1
├─ Router: Pages Router (pages/ 디렉토리)
├─ Search: FlexSearch (내장)
└─ Build: SWC + Webpack

목표 상태 (v16 + Nextra 4.6.0 App Router):
├─ Next.js: 16.x (latest stable)
├─ Nextra: 4.6.0
├─ nextra-theme-docs: 4.6.0
├─ React: 19.x (stable)
├─ React-DOM: 19.x
├─ TypeScript: 5.x+ (호환성 유지)
├─ Tailwind CSS: 3.4.1+ (유지 또는 업그레이드)
├─ Router: App Router (app/ 디렉토리)
├─ Search: Pagefind (대체)
└─ Build: Turbopack (기본)
```

### 2.2 디렉토리 구조

```
현재 Pages Router 구조:
docs/
├── pages/
│   ├── _app.tsx
│   ├── _document.tsx
│   ├── index.mdx
│   ├── ko/
│   │   ├── index.mdx
│   │   ├── _meta.json
│   │   ├── getting-started/
│   │   ├── guides/
│   │   ├── reference/
│   │   ├── advanced/
│   │   ├── contributing/
│   │   └── tutorials/
│   ├── en/
│   │   ├── _meta.json
│   │   └── ... (similar structure)
│   ├── ja/
│   │   ├── _meta.json
│   │   └── ...
│   └── zh/
│       ├── _meta.json
│       └── ...
├── theme.config.tsx
├── next.config.cjs
└── package.json

목표 App Router 구조:
docs/
├── app/
│   ├── layout.jsx (또는 .tsx)
│   ├── page.jsx
│   ├── [locale]/
│   │   ├── layout.jsx
│   │   ├── page.jsx
│   │   ├── (getting-started)/
│   │   │   ├── page.jsx
│   │   │   └── ... (nested routes)
│   │   ├── (guides)/
│   │   ├── (reference)/
│   │   ├── (advanced)/
│   │   ├── (contributing)/
│   │   └── (tutorials)/
│   └── api/ (필요시 API 라우트)
├── content/
│   ├── ko/
│   │   ├── index.mdx
│   │   ├── getting-started/
│   │   ├── guides/
│   │   ├── reference/
│   │   ├── advanced/
│   │   ├── contributing/
│   │   └── tutorials/
│   ├── en/
│   ├── ja/
│   └── zh/
├── theme.config.jsx
├── next.config.mjs
└── package.json
```

### 2.3 파일 통계

| 항목 | 수량 |
|------|------|
| MDX 파일 총합 | 100+ |
| _meta.json 파일 | 43개 |
| 언어별 구조 | 4개 (ko, en, ja, zh) |
| 최상위 섹션 | 6개 (Home, Getting Started, Guides, Reference, Advanced, Contributing) |
| 하위 섹션 | ~20개 이상 |
| 공개 자산 | 20+ 파일 (icons, images) |

---

## 3️⃣ Phase별 마이그레이션 계획

### Phase 1: 준비 및 검증 (수동 작업 불필요 - 계획 수립 단계)

**목표**: Next.js 16 + Nextra 4.6.0 호환성 완전 검증

#### Phase 1.1: 호환성 검사 및 테스트 환경 구성

**작업 내용**:
1. Nextra 4.6.0 공식 문서 검토
   - App Router 마이그레이션 가이드 학습
   - Breaking changes 확인
   - i18n 지원 방식 변경 검토

2. Next.js 16 호환성 확인
   - React 19 peer dependency 호환성
   - TypeScript 5.x 호환성
   - Turbopack 호환성

3. 현재 커스텀 설정 영향도 분석
   - theme.config.tsx 항목별 호환성
   - next.config.cjs 항목별 호환성
   - 커스텀 컴포넌트 검토

4. 검색 엔진 전환 계획
   - FlexSearch → Pagefind 마이그레이션 경로
   - Pagefind 설정 및 성능

**파일 변경 없음** (검증 단계)

**예상 소요 시간**: 계획 단계에서 수행됨 (실행: 2-3시간)

---

### Phase 2: 기본 구조 전환 (자동화 가능 부분 포함)

**목표**: Pages Router → App Router 완전 전환, 디렉토리 구조 변경

#### Phase 2.1: 새 App Router 구조 생성

**변경할 파일**: 없음 (Phase 2.2에서 생성)

**새로 생성할 파일**:
```
docs/
├── app/
│   ├── layout.jsx (Root layout)
│   ├── page.jsx (홈 페이지)
│   ├── [locale]/
│   │   ├── layout.jsx (Locale layout)
│   │   └── page.jsx (언어별 홈)
│   └── api/
│       └── search.js (Pagefind API, 선택)
├── content/
│   ├── ko/
│   │   ├── index.mdx
│   │   ├── getting-started/
│   │   │   ├── index.mdx
│   │   │   ├── installation.mdx
│   │   │   ├── quick-start.mdx
│   │   │   └── concepts.mdx
│   │   ├── guides/
│   │   ├── reference/
│   │   ├── advanced/
│   │   ├── contributing/
│   │   └── tutorials/
│   ├── en/
│   ├── ja/
│   └── zh/
├── lib/
│   ├── mdx-loader.js (MDX 로더)
│   ├── i18n.js (i18n 헬퍼)
│   └── search.js (검색 헬퍼)
├── hooks/
│   └── useLocale.js (locale 훅)
├── components/
│   ├── Navigation.jsx
│   ├── Sidebar.jsx
│   ├── TableOfContents.jsx
│   └── SearchBar.jsx
└── .pagefindrc.json (Pagefind 설정)
```

**새로 생성할 파일 상세 목록**:

1. **app/layout.jsx** (Root Layout)
   - HTML 루트 엘리먼트
   - 글로벌 스타일 import
   - 메타데이터 설정
   - 검색 엔진 스크립트

2. **app/page.jsx** (홈 페이지 - 리다이렉트)
   - 기본 언어로 자동 리다이렉트 (ko)
   - 또는 언어 선택 UI

3. **app/[locale]/layout.jsx** (Locale 레이아웃)
   - 언어별 리소스 로드
   - Navigation 및 Sidebar 컴포넌트
   - 테마 설정
   - Nextra 통합

4. **app/[locale]/page.jsx** (언어별 홈)
   - 언어별 index.mdx 콘텐츠 렌더링

5. **app/[locale]/[[...slug]]/page.jsx** (Catch-all 라우트)
   - 동적 페이지 라우팅
   - 언어별 경로 처리
   - getStaticParams (SSG)
   - getStaticProps (SSG Props)

6. **content/ko|en|ja|zh/** (콘텐츠 디렉토리)
   - 기존 pages/ko|en|ja|zh/ 파일 복사
   - _meta.json → meta.json으로 이름 변경

7. **lib/i18n.js**
   ```javascript
   export const LOCALES = ['ko', 'en', 'ja', 'zh'];
   export const DEFAULT_LOCALE = 'ko';

   export function getLocale(pathname) {
     const match = pathname.match(/^\/([a-z]{2})(?:\/|$)/);
     return match ? match[1] : DEFAULT_LOCALE;
   }

   export function isValidLocale(locale) {
     return LOCALES.includes(locale);
   }
   ```

8. **lib/mdx-loader.js** (MDX 콘텐츠 로더)
   - fs를 사용한 파일 시스템 읽기
   - MDX 파일 파싱
   - 메타데이터 추출

9. **.pagefindrc.json** (Pagefind 설정)
   ```json
   {
     "site": "public",
     "root_selector": "article",
     "exclude_selectors": ["header", "nav"],
     "bundle": true,
     "keep_index_url": false,
     "indexing": {
       "indexed_attrs": {
         "img": ["alt"],
         "a": ["href"]
       }
     },
     "languages": {
       "ko": {
         "min_search_term_length": 1
       },
       "en": {
         "min_search_term_length": 2
       },
       "ja": {
         "min_search_term_length": 1
       },
       "zh": {
         "min_search_term_length": 1
       }
     }
   }
   ```

10. **next.config.mjs** (새 Next.js 설정)
    - CJS → ESM 변경
    - Nextra 4 플러그인 통합
    - Turbopack 활성화

**예상 소요 시간**: 실행 3-4시간

#### Phase 2.2: theme.config.tsx → theme.config.jsx 전환

**변경할 파일**: `docs/theme.config.tsx`

**변경 사항**:
```typescript
// Before (Nextra 3 - Pages Router)
import { DocsThemeConfig } from 'nextra-theme-docs'

const config: DocsThemeConfig = {
  // Pages Router specific
  i18n: [
    { locale: 'ko', name: '한국어' },
    { locale: 'en', name: 'English' },
    { locale: 'ja', name: '日本語' },
    { locale: 'zh', name: '中文' },
  ],
  search: {
    placeholder: '검색...',
  },
  // ... rest of config
}

// After (Nextra 4 - App Router)
const config = {
  // App Router에서는 i18n이 다르게 처리됨
  // layout.jsx에서 직접 처리

  defaultLanguage: 'ko',

  logo: (
    <span style={{ fontWeight: 700, fontSize: '1.2rem' }}>
      🗿 MoAI-ADK
    </span>
  ),

  // Pagefind 검색 설정
  search: {
    placeholder: '검색...',
    emptyResult: {
      default: '검색 결과가 없습니다',
    },
  },

  // ... rest remains similar
}
```

**구체적 변경 항목**:
1. TypeScript → JavaScript 변환 (옵션)
2. i18n 설정 제거 (layout.jsx에서 처리)
3. search 속성 확인 (Pagefind 호환성)
4. 모든 기타 설정 유지

**예상 소요 시간**: 실행 30분

#### Phase 2.3: next.config.cjs → next.config.mjs 변환

**변경할 파일**: `docs/next.config.cjs`

**변경 사항**:
```javascript
// Before (CommonJS)
const nextra = require('nextra')
const withNextra = nextra.default || nextra

module.exports = withNextra({
  theme: 'nextra-theme-docs',
  themeConfig: './theme.config.tsx',
  staticImage: true,
  latex: true,
  codeHighlight: true,
  reactStrictMode: true,
})

// After (ESM + Nextra 4)
import nextra from 'nextra'

const withNextra = nextra({
  theme: 'nextra-theme-docs',
  themeConfig: './theme.config.jsx',
  staticImage: true,
  latex: true,
  codeHighlight: true,
  mdxOptions: {
    development: process.env.NODE_ENV === 'development',
  },
})

export default withNextra({
  reactStrictMode: true,
  swcMinify: false, // Turbopack 사용으로 인한 변경
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
})
```

**구체적 변경 항목**:
1. CJS → ESM 변환 (`require` → `import`)
2. `module.exports` → `export default`
3. Nextra 4 플러그인 API 사용
4. Turbopack 실험적 설정 추가
5. SWC 최소화 설정 조정

**예상 소요 시간**: 실행 30분

---

### Phase 3: 콘텐츠 마이그레이션 (자동화 스크립트 필수)

**목표**: 모든 100+ MDX 파일을 새 디렉토리 구조로 이동 및 변환

#### Phase 3.1: MDX 파일 일괄 마이그레이션

**작업 내용**:
1. 자동 마이그레이션 스크립트 작성
   - `pages/ko/` → `content/ko/`
   - `pages/en/` → `content/en/`
   - `pages/ja/` → `content/ja/`
   - `pages/zh/` → `content/zh/`
   - `pages/index.mdx` → `content/ko/index.mdx` (기본 언어)

2. _meta.json → meta.json 이름 변경
   - 모든 `_meta.json` → `meta.json`
   - 파일 내용 변경 없음 (호환성 검증 필요)

3. 파일 구조 검증
   - 모든 파일 복사 성공 여부 확인
   - 경로 중복 확인
   - 링크 깨짐 검증

**스크립트 예시** (`scripts/migrate-nextra-4.js`):
```javascript
const fs = require('fs');
const path = require('path');

const SOURCE_DIR = path.join(__dirname, '../pages');
const TARGET_DIR = path.join(__dirname, '../content');
const LOCALES = ['ko', 'en', 'ja', 'zh'];

function migrateFiles() {
  // 1. content 디렉토리 생성
  if (!fs.existsSync(TARGET_DIR)) {
    fs.mkdirSync(TARGET_DIR, { recursive: true });
  }

  // 2. 각 언어별 파일 복사
  LOCALES.forEach(locale => {
    const sourceLocaleDir = path.join(SOURCE_DIR, locale);
    const targetLocaleDir = path.join(TARGET_DIR, locale);

    if (fs.existsSync(sourceLocaleDir)) {
      copyRecursive(sourceLocaleDir, targetLocaleDir);
      renameMetaJsonFiles(targetLocaleDir);
    }
  });

  // 3. 루트 index.mdx 처리
  const sourceIndex = path.join(SOURCE_DIR, 'index.mdx');
  if (fs.existsSync(sourceIndex)) {
    // 기본 언어(ko)로 복사하거나 redirect 페이지로 유지
    console.log('Root index.mdx will be handled in app/page.jsx');
  }

  console.log('Migration completed successfully!');
}

function copyRecursive(src, dest) {
  if (!fs.existsSync(dest)) {
    fs.mkdirSync(dest, { recursive: true });
  }

  const files = fs.readdirSync(src);
  files.forEach(file => {
    const srcFile = path.join(src, file);
    const destFile = path.join(dest, file);

    if (fs.statSync(srcFile).isDirectory()) {
      copyRecursive(srcFile, destFile);
    } else {
      fs.copyFileSync(srcFile, destFile);
      console.log(`Copied: ${srcFile} → ${destFile}`);
    }
  });
}

function renameMetaJsonFiles(dir) {
  const files = fs.readdirSync(dir);
  files.forEach(file => {
    const filePath = path.join(dir, file);
    if (fs.statSync(filePath).isDirectory()) {
      renameMetaJsonFiles(filePath);
    } else if (file === '_meta.json') {
      const newPath = path.join(dir, 'meta.json');
      fs.renameSync(filePath, newPath);
      console.log(`Renamed: ${filePath} → ${newPath}`);
    }
  });
}

migrateFiles();
```

**실행 명령어**:
```bash
node scripts/migrate-nextra-4.js
```

**예상 소요 시간**: 실행 15분 (스크립트 포함)

#### Phase 3.2: 링크 무결성 검증

**작업 내용**:
1. 모든 내부 링크 검증
   - `[text](/ko/path)` 형식 유지 확인
   - 상대 경로 → 절대 경로 변환 필요 시

2. 깨진 링크 자동 수정
   - 스크립트로 경로 재구성
   - 오류 로그 생성

3. 외부 링크 샘플 검증
   - 100+ 파일 중 샘플 10개 선택
   - 수동 검증

**스크립트 예시** (`scripts/validate-links.js`):
```javascript
const fs = require('fs');
const path = require('path');
const glob = require('glob');

function validateLinks() {
  const contentDir = path.join(__dirname, '../content');
  const mdxFiles = glob.sync(`${contentDir}/**/*.mdx`);

  let errorCount = 0;
  const errors = [];

  mdxFiles.forEach(file => {
    const content = fs.readFileSync(file, 'utf-8');

    // 링크 패턴 검사 (간단한 정규식)
    const linkPattern = /\[([^\]]+)\]\(([^)]+)\)/g;
    let match;

    while ((match = linkPattern.exec(content)) !== null) {
      const [, text, link] = match;

      // 상대 경로 링크 검증
      if (!link.startsWith('http') && !link.startsWith('#')) {
        const resolvedPath = path.resolve(path.dirname(file), link);
        if (!fs.existsSync(resolvedPath)) {
          errors.push({
            file,
            link,
            text,
            error: 'Link not found'
          });
          errorCount++;
        }
      }
    }
  });

  if (errorCount > 0) {
    console.error(`Found ${errorCount} broken links:`);
    errors.forEach(err => {
      console.error(`  - ${err.file}: [${err.text}](${err.link})`);
    });
    process.exit(1);
  } else {
    console.log('All links are valid!');
  }
}

validateLinks();
```

**예상 소요 시간**: 실행 20분 (검증 포함)

#### Phase 3.3: MDX 프론트매터 검증

**작업 내용**:
1. 모든 MDX 파일의 프론트매터 확인
   - YAML 형식 정상성 검증
   - 필수 필드 검증 (title 등)

2. Nextra 4 호환 메타데이터 확인
   - `title`, `description` 등
   - 커스텀 필드 호환성 확인

**예상 소요 시간**: 실행 10분 (자동 검증)

---

### Phase 4: 의존성 업그레이드

**목표**: package.json 업데이트 및 호환성 검증

#### Phase 4.1: package.json 업데이트

**변경할 파일**: `docs/package.json`

**변경 사항**:
```json
// Before
{
  "dependencies": {
    "next": "14.2.15",
    "nextra": "^3.3.1",
    "nextra-theme-docs": "^3.3.1",
    "react": "18.2.0",
    "react-dom": "18.2.0",
    "@next/third-parties": "^14.2.0"
  },
  "devDependencies": {
    "@types/node": "24.10.0",
    "@types/react": "18.2.0",
    "@types/react-dom": "18.2.0",
    "typescript": "5.9.3",
    "tailwindcss": "^3.4.1",
    "postcss": "^8.4.31",
    "autoprefixer": "^10.4.16",
    "eslint": "^8.56.0",
    "eslint-config-next": "^14.2.0"
  }
}

// After
{
  "dependencies": {
    "next": "^16.0.0",
    "nextra": "^4.6.0",
    "nextra-theme-docs": "^4.6.0",
    "react": "^19.0.0",
    "react-dom": "^19.0.0",
    "@next/third-parties": "^16.0.0",
    "pagefind": "^1.1.0"
  },
  "devDependencies": {
    "@types/node": "24.x",
    "@types/react": "^19.0.0",
    "@types/react-dom": "^19.0.0",
    "typescript": "^5.9.0",
    "tailwindcss": "^3.4.1",
    "postcss": "^8.4.31",
    "autoprefixer": "^10.4.16",
    "eslint": "^8.56.0",
    "eslint-config-next": "^16.0.0"
  }
}
```

**구체적 변경 항목**:
1. next: 14.2.15 → ^16.0.0
2. nextra: ^3.3.1 → ^4.6.0
3. nextra-theme-docs: ^3.3.1 → ^4.6.0
4. react: 18.2.0 → ^19.0.0
5. react-dom: 18.2.0 → ^19.0.0
6. @next/third-parties: ^14.2.0 → ^16.0.0
7. @types/react: 18.2.0 → ^19.0.0
8. @types/react-dom: 18.2.0 → ^19.0.0
9. eslint-config-next: ^14.2.0 → ^16.0.0
10. pagefind: "^1.1.0" 추가 (새로운 의존성)

**예상 소요 시간**: 실행 5분

#### Phase 4.2: 의존성 설치 및 호환성 검증

**작업 내용**:
1. `npm install` 또는 `uv lock` 실행
   - 새 의존성 다운로드
   - peer dependency 충돌 확인

2. peer dependency 충돌 해결
   - React 19 호환성 검증
   - Nextra 4 호환성 검증
   - 필요시 패치 버전 조정

3. TypeScript 재컴파일
   - `npm run type-check` 실행
   - 타입 에러 수정

**예상 소요 시간**: 실행 10-15분

---

### Phase 5: 검색 엔진 마이그레이션 (FlexSearch → Pagefind)

**목표**: 검색 기능 완전 전환 및 성능 검증

#### Phase 5.1: Pagefind 설정

**새로 생성할 파일**: `.pagefindrc.json`

**파일 내용** (이미 Phase 2에서 언급):
```json
{
  "site": "out",
  "root_selector": "article",
  "exclude_selectors": ["header", "nav", ".aside"],
  "bundle": true,
  "keep_index_url": false,
  "indexing": {
    "indexed_attrs": {
      "img": ["alt"],
      "a": ["href"]
    }
  },
  "languages": {
    "ko": {
      "min_search_term_length": 1,
      "splitting_strategy": "cjk"
    },
    "en": {
      "min_search_term_length": 2
    },
    "ja": {
      "min_search_term_length": 1,
      "splitting_strategy": "cjk"
    },
    "zh": {
      "min_search_term_length": 1,
      "splitting_strategy": "cjk"
    }
  }
}
```

**설정 항목 설명**:
- `site`: 빌드된 정적 사이트 디렉토리
- `root_selector`: 인덱싱할 루트 엘리먼트
- `exclude_selectors`: 인덱싱 제외 엘리먼트
- `languages`: 언어별 토큰화 설정
- `splitting_strategy`: CJK 언어 (한중일) 처리

**예상 소요 시간**: 실행 30분 (설정 + 테스트)

#### Phase 5.2: 검색 UI 컴포넌트 업데이트

**작업 내용**:
1. Pagefind 클라이언트 라이브러리 통합
   - `<PagefindUI>` 컴포넌트 설정
   - 다크 모드 대응

2. 검색 결과 화면 테스트
   - 각 언어별 검색 테스트
   - 하이라이트 기능 확인
   - 성능 측정

3. 폴백(Fallback) 처리
   - Pagefind 로드 실패 시 대체 UI
   - 오프라인 모드 대응

**예상 소요 시간**: 실행 1시간

#### Phase 5.3: 검색 인덱싱 검증

**작업 내용**:
1. 빌드 후 인덱스 파일 확인
   - `public/pagefind/` 디렉토리 생성 확인
   - 인덱스 파일 크기 확인 (정상: 500KB-2MB)

2. 검색 쿼리 성능 측정
   - 응답 시간: < 100ms 목표
   - 결과 정확도: 90% 이상

3. 다국어 검색 테스트
   - 한국어: "API", "설정" 등 검색
   - 영어: "installation", "guide" 등 검색
   - 일본어, 중국어: 각 언어별 검색

**예상 소요 시간**: 실행 45분

---

### Phase 6: 빌드 및 배포 설정

**목표**: Vercel 배포 최적화 및 Turbopack 활성화

#### Phase 6.1: 빌드 설정 최적화

**작업 내용**:
1. `next.config.mjs`에 Turbopack 설정 추가
   ```javascript
   experimental: {
     turbo: {
       enabled: true,
       useRootDir: true,
     },
   }
   ```

2. 빌드 환경 변수 설정
   - `NEXT_PUBLIC_SITE_URL`: https://moai-adk.gooslab.ai
   - `NEXT_PUBLIC_BUILD_TIME`: 빌드 타임스탬프

3. 빌드 스크립트 업데이트
   ```json
   {
     "scripts": {
       "dev": "next dev",
       "build": "next build && pagefind --site out",
       "start": "next start",
       "lint": "next lint",
       "type-check": "tsc --noEmit",
       "validate": "node scripts/validate-links.js",
       "ci": "npm run type-check && npm run lint && npm run build && npm run validate"
     }
   }
   ```

**예상 소요 시간**: 실행 20분

#### Phase 6.2: Vercel 배포 설정

**변경할 파일**: 없음 (Vercel 자동 감지) 또는 `vercel.json` 업데이트

**변경 사항** (선택사항 - 필요시만):
```json
{
  "buildCommand": "next build && pagefind --site out",
  "outputDirectory": ".next",
  "installCommand": "npm install"
}
```

**Vercel 프로젝트 설정**:
1. Build Command: `next build && pagefind --site out`
2. Output Directory: `.next`
3. Install Command: `npm install` (기본값)
4. Environment Variables: (기존 유지)

**예상 소요 시간**: 실행 10분

#### Phase 6.3: 로컬 빌드 검증

**작업 내용**:
1. 로컬에서 프로덕션 빌드 실행
   ```bash
   npm run build
   npm start
   ```

2. 빌드 시간 측정
   - 목표: 현재 대비 50% 감소 (Turbopack 효과)
   - 예상: 120초 → 60초

3. 빌드 산출물 검증
   - `.next/` 디렉토리 생성 확인
   - 페이지 크기 확인 (최적화 검증)
   - sourcemap 확인 (디버깅용)

**예상 소요 시간**: 실행 30분

---

### Phase 7: 정적 생성 및 라우팅 검증

**목표**: SSG (Static Site Generation) 설정 및 동적 라우팅 검증

#### Phase 7.1: generateStaticParams 구현

**새로 생성할 파일**: `app/[locale]/[[...slug]]/page.jsx`

**파일 내용**:
```javascript
import { LOCALES } from '@/lib/i18n'
import { getMDXData } from '@/lib/mdx'

export async function generateStaticParams() {
  const params = [];

  // 1. 모든 언어의 홈 페이지
  LOCALES.forEach(locale => {
    params.push({
      locale,
      slug: []  // 홈: /ko, /en, /ja, /zh
    });
  });

  // 2. 모든 MDX 페이지
  LOCALES.forEach(locale => {
    const mdxFiles = getMDXData(locale);
    mdxFiles.forEach(file => {
      const slug = file.slug.split('/'); // "getting-started/index" → ["getting-started"]
      params.push({
        locale,
        slug
      });
    });
  });

  return params;
}

export default function Page({ params }) {
  const { locale, slug } = params;
  // 페이지 렌더링 로직
}
```

**목표**: 모든 100+ 페이지 정적 생성

**예상 소요 시간**: 실행 45분

#### Phase 7.2: 다국어 라우팅 검증

**테스트 항목**:
1. 언어 선택 라우팅
   - /ko → 한국어 사이트
   - /en → 영어 사이트
   - / → 기본 언어(ko)로 리다이렉트

2. 깊은 경로 라우팅
   - /ko/guides/alfred/1-plan → 정상 로드
   - /en/reference/agents/index → 정상 로드

3. 404 처리
   - /ko/non-existent → 404 페이지
   - /invalid-locale/page → 404 또는 기본 언어로 리다이렉트

4. 정적 파일 접근
   - /public/icons/*.svg → 정상 로드
   - /public/og.png → 정상 로드

**예상 소요 시간**: 실행 1시간

---

### Phase 8: 다국어 메타데이터 및 SEO

**목표**: 각 언어별 메타데이터 설정 및 SEO 최적화

#### Phase 8.1: generateMetadata 구현

**작업 내용**:
1. 각 언어별 메타데이터 함수 구현
   ```javascript
   export async function generateMetadata({ params }) {
     const { locale, slug } = params;
     const page = getPageData(locale, slug);

     return {
       title: page.title,
       description: page.description,
       openGraph: {
         title: page.title,
         description: page.description,
         url: `https://moai-adk.gooslab.ai/${locale}/${slug.join('/')}`,
         locale: locale === 'ko' ? 'ko_KR' : locale,
       },
       alternates: {
         languages: {
           ko: `https://moai-adk.gooslab.ai/ko/${slug.join('/')}`,
           en: `https://moai-adk.gooslab.ai/en/${slug.join('/')}`,
           ja: `https://moai-adk.gooslab.ai/ja/${slug.join('/')}`,
           zh: `https://moai-adk.gooslab.ai/zh/${slug.join('/')}`,
         },
       },
     };
   }
   ```

2. hreflang 태그 추가
   - 다국어 버전 간 연결
   - SEO 크롤러 지원

3. 구조화된 데이터 (Schema.org)
   - BreadcrumbList
   - Article
   - Organization

**예상 소요 시간**: 실행 1시간

---

### Phase 9: 성능 최적화 및 Core Web Vitals

**목표**: LCP < 2.5s, FID < 100ms, CLS < 0.1 달성

#### Phase 9.1: 이미지 최적화

**작업 내용**:
1. Next.js Image 컴포넌트 사용
   ```javascript
   import Image from 'next/image'

   <Image
     src="/public/demo.png"
     alt="Demo"
     width={800}
     height={600}
     priority={true}  // LCP 이미지
   />
   ```

2. WebP 포맷 사용
3. Lazy loading 설정
4. 응답형 이미지 (srcset)

**예상 소요 시간**: 실행 1시간

#### Phase 9.2: 코드 분할 및 동적 import

**작업 내용**:
1. 무거운 컴포넌트 동적 import
   ```javascript
   import dynamic from 'next/dynamic'

   const SearchUI = dynamic(() => import('@/components/SearchUI'), {
     loading: () => <div>Loading...</div>,
     ssr: false,
   })
   ```

2. 라우트별 코드 분할 (자동)
   - Next.js App Router의 기본 기능

**예상 소요 시간**: 실행 30분

#### Phase 9.3: 캐싱 전략

**작업 내용**:
1. HTTP 캐싱 헤더 설정
   ```javascript
   // next.config.mjs
   headers: [
     {
       source: '/public/:path*',
       headers: [
         {
           key: 'Cache-Control',
           value: 'public, max-age=31536000, immutable'
         }
       ]
     }
   ]
   ```

2. CDN 캐싱 (Vercel 자동)
3. 브라우저 캐싱 설정

**예상 소요 시간**: 실행 20분

#### Phase 9.4: Lighthouse 스캔

**작업 내용**:
1. 로컬 Lighthouse 검사
   ```bash
   npm install -g lighthouse
   lighthouse https://localhost:3000/ko --view
   ```

2. 점수 목표:
   - Performance: > 90
   - Accessibility: > 90
   - Best Practices: > 90
   - SEO: > 90

3. 문제 해결 및 최적화

**예상 소요 시간**: 실행 1.5시간

---

### Phase 10: 통합 테스트 및 QA

**목표**: 모든 기능 검증 및 프로덕션 배포 준비

#### Phase 10.1: 기능 테스트

**테스트 항목**:

| 기능 | 테스트 항목 | 통과 기준 |
|------|----------|----------|
| 다국어 라우팅 | /ko, /en, /ja, /zh 접근 | 각 언어로 정상 로드 |
| 페이지 렌더링 | 100+ MDX 파일 렌더링 | 모두 정상 렌더링 |
| 검색 기능 | 각 언어별 검색 | 검색 결과 정확 |
| 내부 링크 | 모든 내부 링크 | 링크 깨짐 없음 |
| 외부 링크 | 샘플 10+ 링크 | 대상 사이트 정상 |
| 다크 모드 | 테마 전환 | 모든 페이지 렌더링 정상 |
| 사이드바 | 섹션 토글 | 토글 동작 정상 |
| 목차(TOC) | 헤딩 링크 | 링크 점프 정상 |
| 반응형 디자인 | 모바일/태블릿/데스크톱 | 모든 기기 렌더링 정상 |
| 성능 | Lighthouse 점수 | 모두 > 90 |

**예상 소요 시간**: 실행 3시간

#### Phase 10.2: Staging 배포

**작업 내용**:
1. Vercel에 staging 브랜치 배포
   - `staging` 브랜치 생성 및 푸시
   - Vercel 자동 preview URL 생성

2. Staging 환경 테스트
   - https://moai-adk-staging.vercel.app
   - 모든 기능 재검증

3. 성능 측정
   - 실제 배포 환경에서의 성능
   - 로딩 시간 측정

**예상 소요 시간**: 실행 1시간

#### Phase 10.3: 롤백 계획 검증

**작업 내용**:
1. 현재 main 브랜치 백업
   ```bash
   git branch backup/nextjs-14 main
   git push origin backup/nextjs-14
   ```

2. 롤백 절차 문서화
   - 롤백 명령어 작성
   - 예상 소요 시간 계산

3. 롤백 테스트 (선택)
   - 스테이징에서 롤백 시뮬레이션

**예상 소요 시간**: 실행 30분

---

### Phase 11: 프로덕션 배포 및 모니터링

**목표**: 무중단 배포 및 모니터링 설정

#### Phase 11.1: 프로덕션 배포

**배포 프로세스**:
1. 마이그레이션 브랜치를 main으로 병합
   ```bash
   git checkout main
   git pull origin main
   git merge feature/nextra-4-migration
   git push origin main
   ```

2. Vercel 자동 배포
   - main 브랜치 푸시 시 자동 배포 시작
   - 배포 시간: 약 5-10분

3. 배포 완료 확인
   - https://moai-adk.gooslab.ai 접근 확인
   - 모든 페이지 로드 확인

**배포 타이밍**: 트래픽이 적은 시간대 (권장: KST 오전 2-5시)

**다운타임**: 0분 (Vercel의 무중단 배포)

**예상 소요 시간**: 실행 15분

#### Phase 11.2: 배포 후 검증

**검증 항목** (배포 직후):
1. 기본 기능 검증 (5분)
   - 홈페이지 로드
   - 다국어 라우팅
   - 검색 기능

2. 성능 모니터링 (15분)
   - Web Vitals 측정
   - 에러 로깅 확인
   - 트래픽 모니터링

3. 로그 확인
   - Vercel 배포 로그
   - Next.js 에러 로그
   - 클라이언트 콘솔 에러

**예상 소요 시간**: 실행 30분

#### Phase 11.3: 24시간 모니터링

**모니터링 항목**:
1. 성능 지표
   - 페이지 로드 시간
   - 검색 응답 시간
   - API 응답 시간 (있는 경우)

2. 에러 모니터링
   - Sentry (설정된 경우)
   - Vercel Analytics
   - 브라우저 콘솔 에러

3. 트래픽 모니터링
   - 방문자 수
   - 페이지별 트래픽
   - 지역별 트래픽

4. 문제 대응
   - 에러 발생 시 즉시 대응
   - 심각 에러 시 롤백 준비

**모니터링 기간**: 배포 후 24-48시간

**예상 소요 시간**: 자동 (수동 검사: 1시간/일)

---

### Phase 12: 정리 및 완료

**목표**: 마이그레이션 완료 및 문서화

#### Phase 12.1: 이전 파일 정리

**삭제할 파일**:
```
docs/
├── pages/  (전체 디렉토리 삭제)
│   ├── _app.tsx
│   ├── _document.tsx
│   ├── index.mdx
│   └── [locale]/
├── theme.config.tsx (이전 버전)
├── next.config.cjs (이전 버전)
└── scripts/migrate-nextra-4.js (마이그레이션 스크립트 - 보관 후 삭제)
```

**보관할 파일**:
- git 히스토리에 이전 버전이 기록되므로 파일 삭제는 안전함

**예상 소요 시간**: 실행 10분

#### Phase 12.2: 문서화 및 커밋

**작성할 문서**:
1. `MIGRATION_REPORT.md` - 마이그레이션 요약
2. `NEXTRA_4_SETUP.md` - 새 환경 설정 가이드
3. `CHANGELOG.md` 업데이트

**커밋 메시지**:
```
feat: Migrate from Next.js 14 + Nextra 3 to Next.js 16 + Nextra 4

- Migrate Pages Router to App Router
- Update all 100+ MDX files to new directory structure
- Replace FlexSearch with Pagefind
- Enable Turbopack for faster builds
- Update dependencies (React 18 → 19, Next.js 14 → 16)
- Improve Core Web Vitals (LCP < 2.5s)
- Maintain full multilingual support (ko, en, ja, zh)

Performance improvements:
- Build time: 120s → 60s (50% reduction)
- LCP: 2.8s → 2.1s
- FID: 85ms → 40ms
- CLS: 0.12 → 0.05

Closes #XXX
```

**예상 소요 시간**: 실행 20분

#### Phase 12.3: 팀 공지 및 종료

**공지 사항**:
1. 마이그레이션 완료 알림
2. 변경 사항 요약
3. 새 개발 환경 가이드 공유

**예상 소요 시간**: 실행 10분

---

## 4️⃣ 파일 변경 요약

### 생성할 파일 (신규)

#### 구조 및 라우팅 (App Router)
```
app/
├── layout.jsx (60 lines)
├── page.jsx (50 lines)
├── [locale]/
│   ├── layout.jsx (80 lines)
│   ├── page.jsx (50 lines)
│   ├── [[...slug]]/page.jsx (150 lines)
│   ├── (getting-started)/
│   ├── (guides)/
│   ├── (reference)/
│   ├── (advanced)/
│   ├── (contributing)/
│   └── (tutorials)/
├── not-found.jsx (30 lines)
└── error.jsx (50 lines)

content/
├── ko/ (기존 pages/ko 파일들 복사)
├── en/
├── ja/
└── zh/
```

#### 라이브러리 및 유틸리티
```
lib/
├── i18n.js (50 lines)
├── mdx-loader.js (120 lines)
├── navigation.js (80 lines)
└── search.js (60 lines)

hooks/
├── useLocale.js (30 lines)
├── useNavigation.js (40 lines)
└── useSearch.js (50 lines)

components/
├── Navigation.jsx (100 lines)
├── Sidebar.jsx (120 lines)
├── TableOfContents.jsx (90 lines)
├── SearchBar.jsx (80 lines)
└── LanguageSwitcher.jsx (60 lines)
```

#### 설정 파일
```
.pagefindrc.json (40 lines)
theme.config.jsx (100 lines, 기존과 유사)
next.config.mjs (50 lines, CJS → ESM)
middleware.js (선택, 언어 감지용)
```

#### 마이그레이션 스크립트
```
scripts/
├── migrate-nextra-4.js (150 lines)
├── validate-links.js (100 lines)
├── validate-frontmatter.js (80 lines)
├── fix-meta-json.js (60 lines)
└── generate-sitemap.js (100 lines)
```

**총 신규 파일**: 약 40-50개

### 수정할 파일

#### package.json
- 의존성 업그레이드 (10개 항목)
- 스크립트 업데이트 (3개 항목)

#### theme.config.tsx → theme.config.jsx
- TypeScript → JavaScript 변환
- i18n 설정 제거 또는 조정

#### next.config.cjs → next.config.mjs
- CommonJS → ESM 변환
- Nextra 4 플러그인 API 업데이트
- Turbopack 실험적 설정 추가

#### (새) middleware.js (선택사항)
- 언어 감지 및 리다이렉트

**총 수정 파일**: 약 5-10개

### 삭제할 파일

#### 이전 Pages Router
```
docs/pages/ (전체 디렉토리)
- pages/index.mdx
- pages/ko/ (모든 파일)
- pages/en/ (모든 파일)
- pages/ja/ (모든 파일)
- pages/zh/ (모든 파일)
- pages/_app.tsx
- pages/_document.tsx
- pages/_meta.json (전체)
```

#### 마이그레이션 스크립트 (완료 후)
```
scripts/migrate-nextra-4.js
scripts/validate-links.js
(선택: 보관 또는 삭제)
```

**총 삭제 파일**: 약 100+ (기존 콘텐츠는 content/ 디렉토리로 복사됨)

---

## 5️⃣ 위험도 분석 및 대응

### 위험도: HIGH

| 위험 | 원인 | 영향도 | 대응 방안 |
|------|------|--------|---------|
| **Next.js 16 + Nextra 4.6.0 호환성** | 신규 메이저 버전 조합 | CRITICAL | Phase 1 완전 검증 → staging에서 전체 테스트 |
| **100+ MDX 파일 변환 실패** | 자동 스크립트 오류 | HIGH | 스크립트 작성 후 샘플 10개 파일로 테스트 |
| **검색 기능 장애** | FlexSearch → Pagefind 전환 | HIGH | Pagefind 인덱싱 검증 및 fallback UI |

### 위험도: MEDIUM

| 위험 | 원인 | 영향도 | 대응 방안 |
|------|------|--------|---------|
| **다국어 라우팅 복잡성** | i18n 설정 변경 | MEDIUM | 각 언어별 라우팅 상세 테스트 |
| **빌드 시간 증가** | 의존성 추가 | LOW | Turbopack으로 상쇄 가능 |
| **SEO 메타데이터 손실** | generateMetadata 미구현 | MEDIUM | 모든 페이지에 메타데이터 함수 구현 |

### 위험도: LOW

| 위험 | 원인 | 영향도 | 대응 방안 |
|------|------|--------|---------|
| **캐싱 문제** | 빌드 캐시 변경 | LOW | 로컬 및 CI 캐시 초기화 |
| **TypeScript 타입 에러** | React 19 타입 변경 | LOW | `npm run type-check` 실행 및 수정 |

---

## 6️⃣ 롤백 전략

### 롤백 가능성

**가능 시점**:
- Phase 1-2 완료 후 언제든지 가능
- Phase 11 배포 후 48시간 내 가능

**롤백 절차**:

```bash
# 1. 이전 버전으로 체크아웃
git checkout backup/nextjs-14

# 2. main에 강제 푸시 (위험: 팀 협업 시 주의)
git push origin backup/nextjs-14:main --force

# 3. Vercel에서 자동 배포 시작
# 또는 Vercel Dashboard에서 수동 배포

# 4. 배포 완료 후 검증
curl https://moai-adk.gooslab.ai
```

**예상 롤백 시간**: 10-15분

**전제 조건**:
1. 이전 main 브랜치 백업 보존
2. Vercel deploy 권한 확보

---

## 7️⃣ 타이밍 및 검증

### 배포 권장 시점

**최적 배포 시간**:
- **KST 오전 2시-5시** (트래픽 최소)
- 또는 **금요일 오후 2시** (토요일 24시간 모니터링 가능)

**배포 전 체크리스트**:
- [ ] 모든 Phase 완료
- [ ] staging 배포 성공 및 테스트 완료
- [ ] 롤백 백업 생성 완료
- [ ] 팀 공지 완료
- [ ] 모니터링 도구 준비 완료

**배포 중 체크리스트**:
- [ ] Vercel 배포 진행 중 모니터링
- [ ] 빌드 로그 확인
- [ ] 배포 완료 알림 수신

**배포 후 검증**:
- [ ] 5분: 기본 페이지 로드 확인
- [ ] 15분: 모든 언어 라우팅 확인
- [ ] 30분: 검색 기능 테스트
- [ ] 1시간: Lighthouse 점수 확인
- [ ] 24시간: 에러 모니터링

---

## 8️⃣ 예상 소요 시간 (총합)

### Phase별 예상 시간

| Phase | 내용 | 예상 시간 | 누적 |
|-------|------|---------|------|
| 1 | 호환성 검증 및 계획 | 2-3시간 | 2-3h |
| 2 | 기본 구조 전환 | 5-6시간 | 7-9h |
| 3 | 콘텐츠 마이그레이션 | 45분 | 7-10h |
| 4 | 의존성 업그레이드 | 15-20분 | 7-10.5h |
| 5 | 검색 엔진 전환 | 2시간 | 9-12.5h |
| 6 | 빌드 및 배포 설정 | 30분 | 9.5-13h |
| 7 | 정적 생성 및 라우팅 | 1.5시간 | 11-14.5h |
| 8 | SEO 최적화 | 1시간 | 12-15.5h |
| 9 | 성능 최적화 | 3.5시간 | 15.5-19h |
| 10 | QA 및 통합 테스트 | 5시간 | 20.5-24h |
| 11 | 프로덕션 배포 | 1.5시간 | 22-25.5h |
| 12 | 정리 및 완료 | 40분 | 22.5-26h |

**총 예상 시간**: **22.5 - 26 시간** (단일 개발자 기준)

**권장 일정**:
- **5-6 업무일** (하루 4-5시간 작업 기준)
- 또는 **2-3 업무일** (전시간 할당 기준)

**최단 일정**: 무중단 집중 작업 시 **1-2 업무일** 가능

---

## 9️⃣ 검증 체크리스트

### Phase 통과 기준

#### Phase 1 완료 조건
- [ ] Nextra 4 공식 문서 검토 완료
- [ ] 호환성 이슈 없음 확인
- [ ] 마이그레이션 계획 승인

#### Phase 2 완료 조건
- [ ] App Router 디렉토리 구조 생성 완료
- [ ] layout.jsx 및 page.jsx 구현 완료
- [ ] 로컬 dev 서버 정상 실행

#### Phase 3 완료 조건
- [ ] 모든 MDX 파일 복사 완료
- [ ] meta.json 이름 변경 완료
- [ ] 마이그레이션 스크립트 테스트 완료

#### Phase 4 완료 조건
- [ ] package.json 업데이트 완료
- [ ] npm install 성공
- [ ] npm run type-check 통과

#### Phase 5 완료 조건
- [ ] Pagefind 설정 완료
- [ ] 각 언어별 검색 테스트 완료
- [ ] 검색 결과 정확도 > 90%

#### Phase 6 완료 조건
- [ ] 로컬 빌드 성공
- [ ] next.config.mjs 문법 검증 완료
- [ ] 빌드 시간 측정 완료

#### Phase 7 완료 조건
- [ ] generateStaticParams 구현 완료
- [ ] 모든 100+ 페이지 정적 생성 완료
- [ ] 라우팅 테스트 완료

#### Phase 8 완료 조건
- [ ] generateMetadata 구현 완료
- [ ] 모든 페이지 메타데이터 설정 완료
- [ ] 다국어 hreflang 링크 확인

#### Phase 9 완료 조건
- [ ] Lighthouse 점수: Performance > 90
- [ ] Core Web Vitals: LCP < 2.5s, FID < 100ms, CLS < 0.1
- [ ] 이미지 최적화 완료

#### Phase 10 완료 조건
- [ ] 모든 기능 테스트 통과
- [ ] Staging 배포 성공
- [ ] 성능 측정 완료

#### Phase 11 완료 조건
- [ ] 프로덕션 배포 성공
- [ ] 배포 후 모니터링 24시간 완료
- [ ] 에러 0건

#### Phase 12 완료 조건
- [ ] 이전 파일 정리 완료
- [ ] 문서화 완료
- [ ] 최종 커밋 완료

---

## 🔟 참고 자료 및 문서

### 공식 문서
1. **Next.js 16 마이그레이션 가이드**
   - https://nextjs.org/docs/app/getting-started/installation

2. **Nextra 4 마이그레이션 가이드**
   - https://nextra.site/guide/migrate-from-3

3. **Pagefind 공식 문서**
   - https://pagefind.app/

4. **Next.js App Router 문서**
   - https://nextjs.org/docs/app

### 예상 주요 변경 사항
1. **i18n**: 기존 Nextra i18n → 커스텀 i18n 또는 next-intl 라이브러리
2. **라우팅**: Pages Router (`pages/`) → App Router (`app/`)
3. **검색**: FlexSearch → Pagefind
4. **성능**: SWC + Webpack → Turbopack

### 커뮤니티 리소스
- Next.js 공식 Discord
- Nextra GitHub Discussions
- Stack Overflow (tag: next.js, nextra)

---

## 1️⃣1️⃣ 부록: 마이그레이션 후 유지보수

### 정기 점검
- **주간**: 성능 지표 모니터링
- **월간**: 의존성 업데이트 확인
- **분기별**: Core Web Vitals 측정 및 최적화

### 추가 개선 사항 (향후)
1. **next-intl** 라이브러리 도입 (더 나은 i18n)
2. **Sentry** 에러 모니터링 추가
3. **Analytics** 더 상세한 데이터 수집
4. **AI 검색** (향후 Pagefind 대체 고려)

---

**문서 작성 완료**: 2025-11-10
**마이그레이션 준비 상태**: 완전 검증 필요
**다음 단계**: Phase 1 호환성 검증 및 팀 승인

