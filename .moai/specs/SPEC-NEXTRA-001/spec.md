---
id: SPEC-NEXTRA-001
version: "1.0.0"
status: "draft"
created: "2025-11-28"
updated: "2025-11-28"
author: "GOOS"
priority: "HIGH"
---

## HISTORY

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0.0 | 2025-11-28 | GOOS | Nextra 기반 온라인 문서 사이트 초기 SPEC - 6단계 구현 계획 |

# SPEC-NEXTRA-001: MoAI-ADK 온라인 문서 사이트 구축 (Nextra 기반)

## 개요

MoAI-ADK의 모든 기능과 명령어를 문서화하는 Nextra 기반 온라인 문서 사이트를 구축합니다. README.ko.md와 기존 .moai/docs 내용을 기반으로 체계적인 문서 구조를 제공하고, Git Worktree CLI 문서를 포함한 모든 스킬과 명령어에 대한 상세 가이드를 제공합니다.

**핵심 가치**:
- 🌐 **통합 문서**: 모든 MoAI-ADK 기능을 한 곳에서 검색 가능
- 🎨 **일관된 디자인**: MoAI-ADK 브랜딩과 그레이스케일 테마 적용
- 📱 **반응형**: 모바일 최적화된 문서 경험
- 🚀 **고성능**: Next.js 15+ 기반으로 빠른 로딩 속도
- 🔍 **검색 최적화**: 효율적인 내비게이션과 검색 기능

---

## Environment (실행 환경 및 전제 조건)

### 시스템 환경

- **Node.js**: 18.17+ LTS
- **Next.js**: 15.0+ 최신 버전
- **Nextra**: 3.0+ 최신 버전
- **Package Manager**: npm, yarn, 또는 pnpm
- **배포 환경**: Vercel (권장) 또는 Netlify

### 필수 패키지

```json
{
  "dependencies": {
    "next": "^15.0.0",
    "nextra": "^3.0.0",
    "nextra-theme-docs": "^3.0.0",
    "react": "^19.0.0",
    "react-dom": "^19.0.0"
  },
  "devDependencies": {
    "@types/node": "^20.0.0",
    "@types/react": "^19.0.0",
    "typescript": "^5.0.0",
    "eslint": "^8.0.0",
    "prettier": "^3.0.0"
  }
}
```

### 프로젝트 구조

```
docs/
├── pages/                    # Nextra 페이지 구조
│   ├── index.mdx            # 홈페이지
│   ├── getting-started/     # PART A: 시작하기
│   │   ├── _meta.js         # 네비게이션 메타데이터
│   │   ├── overview.mdx     # 개요
│   │   ├── installation.mdx # 설치 가이드
│   │   └── quickstart.mdx   # 빠른 시작
│   ├── core-concepts/       # PART B: 핵심 개념
│   │   ├── _meta.js
│   │   ├── spec-format.mdx  # SPEC과 EARS 포맷
│   │   ├── agents.mdx       # Mr.Alfred와 에이전트
│   │   ├── workflow.mdx     # 개발 워크플로우
│   │   └── commands.mdx     # 핵심 커맨드
│   ├── advanced/            # PART C: 심화 학습
│   │   ├── _meta.js
│   │   ├── agents-guide.mdx # 에이전트 가이드 (26개)
│   │   ├── skills-library.mdx # 스킬 라이브러리 (22개)
│   │   ├── patterns.mdx     # 조합 패턴과 예제
│   │   └── trust5-quality.mdx # TRUST 5 품질보증
│   ├── worktree/            # Git Worktree CLI 문서
│   │   ├── _meta.js
│   │   ├── guide.mdx        # Worktree 가이드
│   │   ├── examples.mdx     # 사용 예제
│   │   └── faq.mdx          # 자주 묻는 질문
│   └── reference/           # API 레퍼런스
│       ├── _meta.js
│       ├── skills.mdx       # 스킬 API 레퍼런스
│       └── commands.mdx     # 명령어 레퍼런스
├── styles/                  # 스타일 파일
│   └── globals.css          # 전역 스타일 (그레이스케일 테마)
├── components/              # React 컴포넌트
├── public/                  # 정적 에셋
├── theme.config.tsx         # Nextra 테마 설정
├── next.config.js           # Next.js 설정
└── package.json             # 패키지 설정
```

---

## Assumptions (가정 사항)

### 기술적 가정

1. **기존 콘텐츠**: README.ko.md (1,865줄)와 .moai/docs/ 내용이 마이그레이션 대상임
2. **스킬 통합**: 22개의 전문 스킬 문서가 자동으로 레퍼런스 페이지에 포함됨
3. **명령어 문서**: 6개의 코어 커맨드(/moai:0-3)가 상세하게 문서화됨
4. **워크트리 통합**: SPEC-WORKTREE-001의 모든 결과물이 통합됨

### 사용자 가정

1. **주요 사용자**: 한국어 사용자, MoAI-ADK 개발자
2. **기술 수준**: 중급 이상 개발자 (CLI, Git, 기본 웹 기술 이해)
3. **접근성**: 모바일 기기와 데스크톱 모두에서 최적화된 경험 필요
4. **검색**: 키워드 기반 검색이 효율적으로 작동해야 함

---

## Requirements (요구사항)

### Requirements (Ubiquitous) - 보편적 요구사항

- 사이트는 Next.js 15+와 Nextra 3.0+로 구현되어야 한다.
- 모든 페이지는 MDX 형식으로 작성되어야 한다.
- 사이트는 라이트 모드와 다크 모드를 모두 지원해야 한다.
- 모든 콘텐츠는 한국어로 제공되어야 한다.
- MoAI-ADK 브랜딩과 일관된 그레이스케일 테마를 적용해야 한다.

### Requirements (Event-driven) - 이벤트 기반 요구사항

- WHEN 사용자가 홈페이지에 방문했을 때 THEN 핵심 기능과 빠른 시작 가이드가 표시되어야 한다.
- WHEN 사용자가 검색어를 입력했을 때 THEN 관련 페이지와 섹션이 표시되어야 한다.
- WHEN 사용자가 워크트리 섹션을 탐색할 때 THEN 모든 워크트리 관련 문서로 이동할 수 있어야 한다.
- WHEN 사용자가 스킬 레퍼런스를 조회할 때 THEN 22개 스킬의 상세 정보를 볼 수 있어야 한다.

### Requirements (State-driven) - 상태 기반 요구사항

- 사이트는 Lighthouse 90+ 점수를 달성해야 한다.
- 첫 페이지 로딩 시간은 2.5초 이내여야 한다.
- 모든 이미지는 최적화되어 Core Web Vitals를 만족해야 한다.
- 테마 전환(라이트/다크)은 즉시 적용되어야 한다.
- 검색 기능은 300ms 이내에 결과를 표시해야 한다.

### Requirements (Interface) - 인터페이스 요구사항

- 모든 내부 링크는 자동으로 생성되어야 한다 (Table of Contents).
- 이전/다음 페이지 네비게이션이 제공되어야 한다.
- 헤더와 푸터는 접근성 가이드라인(WCAG 2.1+)을 준수해야 한다.
- 코드 블록은 구문 강조와 복사 기능을 제공해야 한다.
- 모바일에서는 햄버거 메뉴가 제공되어야 한다.

---

## Specifications (상세 명세)

### 1. CSS 스타일 시스템

#### 1.1 그레이스케일 테마 변수

```css
:root {
  /* 웹폰트 설정 */
  --font-sans: 'Pretendard', 'Inter', system-ui, sans-serif;
  --font-en: 'Inter', 'Roboto', 'Helvetica Neue', Arial, sans-serif;
  --font-ko: 'Pretendard', 'Noto Sans KR', 'Apple SD Gothic Neo', sans-serif;
  --font-code: 'JetBrains Mono', 'Hack', 'Consolas', 'Monaco', monospace;

  /* 라이트 테마 */
  --color-primary-fg: #000000;
  --color-primary-bg: #FFFFFF;
  --color-accent-fg: #666666;
  --color-bg: #FFFFFF;
  --color-bg-light: #F9F9F9;
  --color-surface: #F5F5F5;
  --color-text: #000000;
  --color-text-secondary: #666666;
  --color-border: #DDDDDD;
  --color-code-bg: #F0F0F0;
  --color-brand-primary: #1976d2;
  --color-brand-accent: #dc2626;
  --color-brand-success: #059669;
}

/* 다크 테마 오버라이드 */
[data-theme="dark"] {
  --color-primary-fg: #FFFFFF;
  --color-primary-bg: #121212;
  --color-accent-fg: #BBBBBB;
  --color-bg: #121212;
  --color-bg-light: #1E1E1E;
  --color-surface: #1E1E1E;
  --color-text: #FFFFFF;
  --color-text-secondary: #BBBBBB;
  --color-border: #333333;
  --color-code-bg: #1E1E1E;
  --color-brand-primary: #4dabf7;
}
```

#### 1.2 한글 타이포그래피

```css
/* 한글 최적화 */
.ko-typography {
  font-family: var(--font-ko);
  letter-spacing: -0.02em;
  line-height: 1.7;
  word-break: keep-word;
}

/* 코드 블록 스타일링 */
.code-block {
  font-family: var(--font-code);
  background: var(--color-code-bg);
  border: 1px solid var(--color-border);
  border-radius: 6px;
  padding: 1rem;
  overflow-x: auto;
}
```

### 2. 콘텐츠 마이그레이션 명세

#### 2.1 README.ko.md 구조화

| README 섹션 | Nextra 페이지 | 우선순위 |
|-------------|--------------|----------|
| PART A (시작하기) | getting-started/* | HIGH |
| PART B (핵심 개념) | core-concepts/* | HIGH |
| PART C (심화 학습) | advanced/* | MEDIUM |
| PART D (고급 & 참고) | reference/* | MEDIUM |
| 워크트리 관련 | worktree/* | HIGH |

#### 2.2 .moai/docs/ 통합

- WORKTREE_GUIDE.md → worktree/guide.mdx
- WORKTREE_FAQ.md → worktree/faq.mdx
- WORKTREE_EXAMPLES.md → worktree/examples.mdx

### 3. Nextra 설정 명세

#### 3.1 theme.config.tsx

```typescript
export default {
  logo: <span>🗿 MoAI-ADK</span>,
  project: {
    link: 'https://github.com/moai-ai/moai-adk',
  },
  docsRepositoryBase: 'https://github.com/moai-ai/moai-adk/tree/main/docs',
  footer: {
    text: 'MIT License © 2025 MoAI-ADK Contributors',
  },
  toc: {
    backToTop: true,
    extraContent: (
      <div style={{ marginTop: '1rem' }}>
        <a href="/worktree">Git Worktree CLI →</a>
      </div>
    ),
  },
  sidebar: {
    defaultMenuCollapseLevel: 1,
  },
}
```

### 4. 성능 최적화 명세

#### 4.1 이미지 최적화

- 모든 이미지는 WebP 형식으로 제공
- 반응형 이미지 (srcset) 적용
- 지연 로딩 (lazy loading) 적용
- Core Web Vitals 준수

#### 4.2 번들 최적화

- 코드 스플리팅 (페이지별 분리)
- 트리 쉐이킹 (사용하지 않는 코드 제거)
- gzip 압축 활성화
- CDN을 통한 정적 에셋 전송

### 5. 검색 기능 명세

- 전체 텍스트 검색 지원
- 페이지 섹션별 검색
- 실시간 검색 결과 표시
- 검색 기록 저장
- 키보드 단축키 (Cmd/Ctrl + K)

### 6. 접근성 명세

- WCAG 2.1 AA 준수
- 키보드 내비게이션 지원
- 스크린 리더 최적화
- 고대비 모드 지원
- 포커스 관리

---

## Constraints (제약 조건)

### 기술적 제약

- Next.js 15+와 Nextra 3.0+를 반드시 사용해야 한다.
- 모든 색상은 CSS 변수를 통해서만 정의해야 한다 (하드코딩 금지).
- 웹폰트는 Pretendard, Inter, JetBrains Mono만 사용해야 한다.
- 배포는 Vercel로 제한한다.
- 빌드 타임은 3분 이내여야 한다.

### 콘텐츠 제약

- 모든 문서는 한국어로 작성되어야 한다.
- 코드 예제는 실제로 동작하는 코드여야 한다.
- 외부 링크는 모두 https를 사용해야 한다.
- 이미지는 alt 속성을 반드시 포함해야 한다.

### 디자인 제약

- 그레이스케일 컬러 팔레트만 사용해야 한다.
- 브랜드 색상은 보조색으로만 제한적으로 사용해야 한다.
- 일관된 간격과 타이포그래피를 유지해야 한다.
- 다크 모드는 라이트 모드의 반전이 아닌 최적화된 색상을 사용해야 한다.

### 시간 제약

- MVP 구현: 2주 이내
- 전체 기능 구현: 4주 이내
- 성능 최적화: 1주 이내
- 배포 및 안정화: 1주 이내

---

## Traceability (추적성)

### 관련 SPEC

- **SPEC-WORKTREE-001**: 워크트리 CLI 문서 통합 (HIGH)
- **SPEC-UPDATE-001**: 기존 문서 업데이트 패턴 참고 (MEDIUM)

### 의존성

- **manager-docs**: 문서 생성 및 관리 에이전트
- **moai-library-nextra**: Nextra 통합 스킬
- **docs/styles/globals.css**: 기존 스타일 시스템 참고

### 성공 기준

- **문서 완성도**: 100% (모든 기능 문서화)
- **성능 점수**: Lighthouse 90+
- **접근성**: WCAG 2.1 AA 준수
- **사용자 만족도**: 4.5/5.0 이상 (내부 테스트)

### 리스크 관리

- **HIGH**: 대용량 콘텐츠 마이그레이션 → 자동화 스크립트 개발
- **MEDIUM**: 성능 최적화 → 사전 프로파일링 및 테스트
- **LOW**: 브라우저 호환성 → 모던 브라우저만 지원으로 범위 축소