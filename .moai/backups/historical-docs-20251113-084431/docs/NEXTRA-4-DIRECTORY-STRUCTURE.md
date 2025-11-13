# Nextra 4 마이그레이션 - 디렉토리 구조 상세 가이드

## 1. 현재 구조 (Pages Router + Nextra 3.3.1)

```
MoAI-ADK/docs/
├── public/                          # 정적 자산
│   ├── icons/
│   │   ├── check-circle.svg
│   │   ├── performance.svg
│   │   ├── language.svg
│   │   └── ... (15+ 아이콘)
│   ├── apple-icon-180x180.png
│   ├── favicon.ico
│   ├── og.png
│   └── ...
│
├── pages/                           # Pages Router (삭제 예정)
│   ├── _app.tsx                     # 앱 래퍼 (X 삭제)
│   ├── _document.tsx                # HTML 문서 (X 삭제)
│   ├── index.mdx                    # 루트 인덱스 (X 이동)
│   │
│   ├── ko/                          # 한국어 콘텐츠
│   │   ├── _meta.json               # → content/ko/meta.json
│   │   ├── index.mdx                # → content/ko/index.mdx
│   │   ├── getting-started/
│   │   │   ├── _meta.json           # → content/ko/getting-started/meta.json
│   │   │   ├── index.mdx            # → content/ko/getting-started/index.mdx
│   │   │   ├── installation.mdx
│   │   │   ├── quick-start.mdx
│   │   │   ├── concepts.mdx
│   │   │   └── glossary.mdx
│   │   ├── guides/
│   │   │   ├── _meta.json
│   │   │   ├── index.mdx
│   │   │   ├── alfred/
│   │   │   ├── specs/
│   │   │   ├── tdd/
│   │   │   ├── project/
│   │   │   └── contributing-translations.mdx
│   │   ├── reference/
│   │   │   ├── _meta.json
│   │   │   ├── agents/
│   │   │   ├── cli/
│   │   │   ├── hooks/
│   │   │   ├── skills/
│   │   │   └── tags/
│   │   ├── advanced/
│   │   │   ├── _meta.json
│   │   │   ├── architecture.mdx
│   │   │   ├── extensions.mdx
│   │   │   ├── i18n.mdx
│   │   │   ├── performance.mdx
│   │   │   └── security.mdx
│   │   ├── contributing/
│   │   │   ├── _meta.json
│   │   │   ├── index.mdx
│   │   │   ├── development.mdx
│   │   │   ├── releases.mdx
│   │   │   └── style.mdx
│   │   ├── tutorials/
│   │   │   ├── _meta.json
│   │   │   ├── index.mdx
│   │   │   └── hello-world-api.mdx
│   │   ├── troubleshooting/
│   │   │   ├── _meta.json
│   │   │   └── index.mdx
│   │   └── translation-status.mdx
│   │
│   ├── en/                          # 영어 콘텐츠
│   │   ├── _meta.json               # → content/en/meta.json
│   │   ├── index.mdx                # → content/en/index.mdx
│   │   ├── getting-started/
│   │   ├── guides/
│   │   ├── reference/
│   │   ├── advanced/
│   │   ├── contributing/
│   │   └── ... (유사한 구조)
│   │
│   ├── ja/                          # 일본어 콘텐츠
│   │   ├── _meta.json
│   │   ├── guides/
│   │   ├── reference/
│   │   ├── advanced/
│   │   ├── contributing/
│   │   ├── tutorials/
│   │   └── troubleshooting/
│   │
│   └── zh/                          # 중국어 콘텐츠
│       ├── _meta.json
│       ├── contributing/
│       ├── reference/
│       └── tutorials/
│
├── theme.config.tsx                 # Nextra 테마 설정 (수정)
├── next.config.cjs                  # Next.js 설정 (수정)
├── package.json                     # 의존성 (수정)
├── tsconfig.json                    # TypeScript 설정 (유지)
├── tailwind.config.js               # Tailwind 설정 (유지)
├── postcss.config.js                # PostCSS 설정 (유지)
└── README.md                        # 프로젝트 문서 (유지)
```

---

## 2. 목표 구조 (App Router + Nextra 4.6.0)

```
MoAI-ADK/docs/
├── public/                          # 정적 자산 (변경 없음)
│   ├── icons/
│   ├── apple-icon-180x180.png
│   ├── favicon.ico
│   └── ... (모두 유지)
│
├── app/                             # App Router (신규)
│   ├── layout.jsx                   # 루트 레이아웃 (신규)
│   │   └── 구성:
│   │       - HTML DOCTYPE 및 기본 메타
│   │       - 글로벌 스타일 import
│   │       - 상단바, 네비게이션
│   │       - 하단바
│   │
│   ├── page.jsx                     # 루트 페이지 (신규)
│   │   └── 기능: 기본 언어(ko)로 리다이렉트
│   │
│   ├── [locale]/                    # 동적 언어 라우트 (신규)
│   │   ├── layout.jsx               # 언어별 레이아웃 (신규)
│   │   │   └── 구성:
│   │   │       - 언어 변수 확인
│   │   │       - 언어별 리소스 로드
│   │   │       - Navigation 및 Sidebar
│   │   │       - 테마 제공자
│   │   │       - Nextra 통합
│   │   │
│   │   ├── page.jsx                 # 언어별 홈 (신규)
│   │   │   └── 기능: 언어별 index.mdx 렌더링
│   │   │
│   │   ├── (getting-started)/       # 라우트 그룹
│   │   │   ├── page.jsx             # /ko/getting-started
│   │   │   ├── installation/page.jsx
│   │   │   ├── quick-start/page.jsx
│   │   │   ├── concepts/page.jsx
│   │   │   └── glossary/page.jsx
│   │   │
│   │   ├── (guides)/                # 라우트 그룹
│   │   │   ├── page.jsx             # /ko/guides
│   │   │   ├── alfred/page.jsx
│   │   │   ├── specs/page.jsx
│   │   │   ├── tdd/page.jsx
│   │   │   ├── project/page.jsx
│   │   │   └── contributing-translations/page.jsx
│   │   │
│   │   ├── (reference)/             # 라우트 그룹
│   │   │   ├── page.jsx
│   │   │   ├── agents/page.jsx
│   │   │   ├── cli/page.jsx
│   │   │   ├── hooks/page.jsx
│   │   │   ├── skills/page.jsx
│   │   │   └── tags/page.jsx
│   │   │
│   │   ├── (advanced)/              # 라우트 그룹
│   │   │   ├── page.jsx
│   │   │   ├── architecture/page.jsx
│   │   │   ├── extensions/page.jsx
│   │   │   ├── i18n/page.jsx
│   │   │   ├── performance/page.jsx
│   │   │   └── security/page.jsx
│   │   │
│   │   ├── (contributing)/          # 라우트 그룹
│   │   │   ├── page.jsx
│   │   │   ├── development/page.jsx
│   │   │   ├── releases/page.jsx
│   │   │   └── style/page.jsx
│   │   │
│   │   ├── (tutorials)/             # 라우트 그룹
│   │   │   ├── page.jsx
│   │   │   └── hello-world-api/page.jsx
│   │   │
│   │   ├── (troubleshooting)/       # 라우트 그룹
│   │   │   └── page.jsx
│   │   │
│   │   ├── translation-status/page.jsx
│   │   │
│   │   └── [[...slug]]/page.jsx     # Catch-all (선택) - 특정 패턴 처리용
│   │
│   ├── not-found.jsx                # 404 페이지 (신규)
│   ├── error.jsx                    # 에러 경계 (신규)
│   └── layout-alternative/          # 선택적 다른 레이아웃
│
├── content/                         # MDX 콘텐츠 저장소 (신규)
│   ├── ko/                          # pages/ko → content/ko
│   │   ├── meta.json                # _meta.json → meta.json
│   │   ├── index.mdx
│   │   ├── getting-started/
│   │   │   ├── meta.json
│   │   │   ├── index.mdx
│   │   │   ├── installation.mdx
│   │   │   ├── quick-start.mdx
│   │   │   ├── concepts.mdx
│   │   │   └── glossary.mdx
│   │   ├── guides/
│   │   │   ├── meta.json
│   │   │   ├── index.mdx
│   │   │   ├── alfred/
│   │   │   │   ├── meta.json
│   │   │   │   ├── 1-plan.mdx
│   │   │   │   ├── 2-run.mdx
│   │   │   │   ├── 3-sync.mdx
│   │   │   │   └── 9-feedback.mdx
│   │   │   ├── specs/
│   │   │   │   ├── meta.json
│   │   │   │   ├── basics.mdx
│   │   │   │   ├── ears.mdx
│   │   │   │   ├── examples.mdx
│   │   │   │   └── tags.mdx
│   │   │   ├── tdd/
│   │   │   │   ├── meta.json
│   │   │   │   ├── index.mdx
│   │   │   │   ├── red.mdx
│   │   │   │   ├── green.mdx
│   │   │   │   └── refactor.mdx
│   │   │   ├── project/
│   │   │   │   ├── meta.json
│   │   │   │   ├── index.mdx
│   │   │   │   ├── config.mdx
│   │   │   │   ├── deploy.mdx
│   │   │   │   └── init.mdx
│   │   │   └── contributing-translations.mdx
│   │   ├── reference/
│   │   │   ├── meta.json
│   │   │   ├── index.mdx
│   │   │   ├── agents/
│   │   │   │   ├── meta.json
│   │   │   │   ├── index.mdx
│   │   │   │   ├── core.mdx
│   │   │   │   ├── experts.mdx
│   │   │   │   └── ...
│   │   │   ├── cli/
│   │   │   ├── hooks/
│   │   │   ├── skills/
│   │   │   └── tags/
│   │   ├── advanced/
│   │   │   ├── meta.json
│   │   │   ├── architecture.mdx
│   │   │   ├── extensions.mdx
│   │   │   ├── i18n.mdx
│   │   │   ├── performance.mdx
│   │   │   └── security.mdx
│   │   ├── contributing/
│   │   │   ├── meta.json
│   │   │   ├── index.mdx
│   │   │   ├── development.mdx
│   │   │   ├── releases.mdx
│   │   │   └── style.mdx
│   │   ├── tutorials/
│   │   │   ├── meta.json
│   │   │   ├── index.mdx
│   │   │   └── hello-world-api.mdx
│   │   ├── troubleshooting/
│   │   │   ├── meta.json
│   │   │   └── index.mdx
│   │   └── translation-status.mdx
│   │
│   ├── en/                          # pages/en → content/en
│   │   ├── meta.json
│   │   ├── index.mdx
│   │   └── ... (유사한 구조)
│   │
│   ├── ja/                          # pages/ja → content/ja
│   │   ├── meta.json
│   │   └── ... (완전한 구조)
│   │
│   └── zh/                          # pages/zh → content/zh
│       ├── meta.json
│       └── ... (부분 구조)
│
├── lib/                             # 유틸리티 라이브러리 (신규)
│   ├── i18n.js                      # 국제화 헬퍼
│   │   └── 함수:
│   │       - getLocale()
│   │       - isValidLocale()
│   │       - getLocaleConfig()
│   │       - getAvailableLocales()
│   │
│   ├── mdx-loader.js                # MDX 파일 로더
│   │   └── 함수:
│   │       - loadMDXFile()
│   │       - getMDXFrontmatter()
│   │       - getMDXContent()
│   │       - listMDXFiles()
│   │
│   ├── navigation.js                # 네비게이션 구조
│   │   └── 함수:
│   │       - getNavigation()
│   │       - getBreadcrumbs()
│   │       - getNextPrevious()
│   │
│   └── search.js                    # 검색 헬퍼
│       └── 함수:
│           - initPagefind()
│           - searchPages()
│           - getSearchResults()
│
├── hooks/                           # React 커스텀 훅 (신규)
│   ├── useLocale.js                 # 언어 정보 훅
│   ├── useNavigation.js             # 네비게이션 훅
│   ├── useTheme.js                  # 테마 훅
│   └── useSearch.js                 # 검색 훅
│
├── components/                      # React 컴포넌트 (신규/수정)
│   ├── Navigation.jsx               # 상단 네비게이션
│   ├── Sidebar.jsx                  # 좌측 사이드바
│   ├── TableOfContents.jsx          # 목차
│   ├── SearchBar.jsx                # 검색 바
│   ├── LanguageSwitcher.jsx         # 언어 선택기
│   ├── Footer.jsx                   # 하단 푸터
│   └── MobileMenu.jsx               # 모바일 메뉴
│
├── theme.config.jsx                 # Nextra 테마 설정 (수정)
│   └── 변경사항:
│       - TypeScript → JavaScript
│       - i18n 설정 제거/조정
│       - Pagefind 검색 설정
│
├── next.config.mjs                  # Next.js 설정 (수정)
│   └── 변경사항:
│       - CommonJS → ESM
│       - Nextra 4 플러그인
│       - Turbopack 설정
│
├── .pagefindrc.json                 # Pagefind 설정 (신규)
│   └── 설정:
│       - 인덱싱 경로
│       - 언어별 설정
│       - 제외 선택자
│
├── middleware.js                    # 언어 감지 미들웨어 (신규/선택)
│   └── 기능:
│       - URL 언어 감지
│       - 자동 리다이렉트
│
├── postcss.config.js                # PostCSS 설정 (유지)
├── tailwind.config.js               # Tailwind CSS 설정 (유지)
├── tsconfig.json                    # TypeScript 설정 (수정)
├── package.json                     # 의존성 (수정)
│
├── scripts/                         # 마이그레이션 스크립트 (신규)
│   ├── migrate-nextra-4.js          # 파일 마이그레이션
│   ├── validate-links.js            # 링크 검증
│   ├── validate-frontmatter.js      # 프론트매터 검증
│   ├── fix-meta-json.js             # meta.json 변환
│   ├── generate-sitemap.js          # Sitemap 생성
│   └── validate-build.js            # 빌드 검증
│
└── README.md                        # 프로젝트 문서 (수정)
```

---

## 3. 주요 변경사항 상세 설명

### 3.1 라우트 구조 변경

**Pages Router (현재)**:
```
pages/
├── index.mdx → URL: /
├── ko/
│   ├── index.mdx → URL: /ko
│   └── guides/index.mdx → URL: /ko/guides
└── en/
    ├── index.mdx → URL: /en
    └── guides/index.mdx → URL: /en/guides
```

**App Router (목표)**:
```
app/
├── page.jsx → URL: / (리다이렉트 to /ko)
├── [locale]/
│   ├── page.jsx → URL: /ko, /en, /ja, /zh
│   ├── (guides)/
│   │   ├── page.jsx → URL: /ko/guides, /en/guides, ...
│   │   └── ...
│   └── ...
└── not-found.jsx → URL: /* (매칭 안 되는 경로)
```

### 3.2 콘텐츠 저장 위치 변경

**Pages Router**: `pages/{locale}/*.mdx` (라우트와 콘텐츠 동시)
**App Router**: `content/{locale}/*.mdx` (콘텐츠 분리)

**장점**:
- 라우트 로직과 콘텐츠 분리 (관심사의 분리)
- 더 명확한 구조
- 나중에 콘텐츠 마이그레이션 용이

### 3.3 메타데이터 파일 변경

**현재**: `_meta.json` (private 파일 관례)
**목표**: `meta.json` (Nextra 4 표준)

**변경 규칙**:
- 모든 `_meta.json` → `meta.json`
- 파일 내용: 변경 없음 (호환성 유지)

**변환 스크립트 실행 예**:
```javascript
// 모든 _meta.json을 meta.json으로 이름 변경
find content/ -name "_meta.json" -exec rename '_meta' 'meta' {} \;
```

### 3.4 라우트 그룹 활용

**라우트 그룹** `(group-name)`: URL에 포함되지 않음

**예시**:
```
app/[locale]/(guides)/
├── page.jsx → URL: /ko/guides (O)
├── alfred/page.jsx → URL: /ko/guides/alfred (O)
└── NOT /ko/(guides)/guides ❌
```

**목적**:
- 논리적 그룹화
- 사이드바 구조와 일치
- 깔끔한 URL 유지

---

## 4. 마이그레이션 순서

### Step 1: 디렉토리 생성
```bash
mkdir -p docs/app/{[locale]/{(getting-started),(guides),(reference),(advanced),(contributing),(tutorials),(troubleshooting)}}
mkdir -p docs/content/{ko,en,ja,zh}
mkdir -p docs/lib docs/hooks docs/components docs/scripts
```

### Step 2: 파일 복사
```bash
# 콘텐츠 복사
cp -r docs/pages/ko/* docs/content/ko/
cp -r docs/pages/en/* docs/content/en/
cp -r docs/pages/ja/* docs/content/ja/
cp -r docs/pages/zh/* docs/content/zh/
```

### Step 3: meta.json 이름 변경
```bash
find docs/content -name "_meta.json" -exec rename 's/_meta/meta/' {} \;
```

### Step 4: 라우트 파일 생성
```bash
# app/layout.jsx
# app/page.jsx
# app/[locale]/layout.jsx
# app/[locale]/page.jsx
# app/[locale]/(getting-started)/page.jsx
# ... 등등
```

### Step 5: 라이브러리/훅/컴포넌트 생성
```bash
# lib/i18n.js
# hooks/useLocale.js
# components/Navigation.jsx
# ... 등등
```

### Step 6: 설정 파일 업데이트
```bash
# next.config.mjs (cjs → mjs)
# theme.config.jsx (tsx → jsx)
# package.json (의존성 업그레이드)
# .pagefindrc.json (신규)
```

### Step 7: 테스트 및 검증
```bash
npm install
npm run dev
npm run build
```

---

## 5. 파일 개수 요약

| 분류 | 현재 | 목표 | 변화 |
|------|------|------|------|
| **콘텐츠 파일 (MDX)** | 100+ | 100+ | 위치만 변경 |
| **_meta.json / meta.json** | 43 | 43 | 이름 변경 |
| **app/ 라우트 파일** | 0 | 40+ | 신규 생성 |
| **lib/ 유틸리티** | 0 | 5 | 신규 생성 |
| **hooks/ 커스텀 훅** | 0 | 4 | 신규 생성 |
| **components/ 컴포넌트** | 0 | 7 | 신규 생성 |
| **마이그레이션 스크립트** | 0 | 6 | 신규 생성 |
| **설정 파일** | 3 | 4 | 1개 추가 |
| **pages/ 디렉토리** | 143+ | 0 | 삭제 |

---

## 6. 체크리스트

- [ ] 신규 디렉토리 구조 생성 (app/, content/, lib/, hooks/, components/)
- [ ] MDX 파일 모두 content/ 디렉토리로 복사
- [ ] _meta.json 모두 meta.json으로 이름 변경
- [ ] 라우트 파일(page.jsx) 모두 생성
- [ ] 레이아웃 파일(layout.jsx) 생성
- [ ] 유틸리티 함수 구현 (lib/i18n.js, lib/mdx-loader.js 등)
- [ ] 커스텀 훅 구현 (useLocale.js, useNavigation.js 등)
- [ ] React 컴포넌트 구현 (Navigation, Sidebar, SearchBar 등)
- [ ] 설정 파일 업데이트 (next.config.mjs, theme.config.jsx, .pagefindrc.json)
- [ ] package.json 의존성 업그레이드
- [ ] 로컬 빌드 테스트
- [ ] 모든 언어 라우팅 테스트
- [ ] 검색 기능 테스트
- [ ] 성능 측정 (Lighthouse)
- [ ] pages/ 디렉토리 삭제

