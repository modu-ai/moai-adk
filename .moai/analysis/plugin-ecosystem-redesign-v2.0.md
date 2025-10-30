# MoAI-ADK 플러그인 생태계 v2.0 재설계

**문서 상태**: 설계 단계 - 사용자 승인 대기
**작성일**: 2025-10-31
**범위**: 스킬 분류법 정정, UI/UX 플러그인 확장, 블로그 플러그인 추가

---

## 🎯 개요

이 문서는 3가지 주요 재설계를 다룹니다:

1. **스킬 분류 체계 (Skill Taxonomy)** - 10개 신규 스킬의 정확한 이름 변경 및 재분류
2. **UI/UX 플러그인 확장** - 3개 에이전트 → 7개 에이전트로 확대, Figma MCP 통합
3. **블로그 작성 플러그인 신규 추가** - WordPress, 네이버블로그, 티스토리 지원

---

## 📋 Phase 1: 스킬 분류 체계 정정

### 문제 상황

현재 스킬 명명이 일관성이 없습니다:
- ❌ `moai-lang-nextjs-advanced` (프레임워크인데 "lang"이라고 명명)
- ❌ `moai-lang-react-19` (프레임워크인데 "lang"이라고 명명)
- ❌ `moai-lang-fastapi-patterns` (프레임워크인데 "lang"이라고 명명)
- ❌ `moai-deploy-*` (SaaS 플랫폼을 "deploy"라고 명명)

**명확한 정의**:
- `language` = 프로그래밍 언어 (Python, TypeScript, JavaScript, Go, Rust 등)
- `framework` = 웹 프레임워크 (Next.js, React, FastAPI, Express, Django 등)
- `design` = 디자인 시스템 & UI (Tailwind, shadcn/ui, Figma, 디자인 토큰)
- `saas` = 공식 MCP 기반 SaaS 플랫폼 (Vercel, Supabase, Render, WordPress 등)

### 스킬 분류 체계 구조

```
Tier 1: Foundation (글로벌 인프라)
├─ moai-foundation-*
│  ├─ moai-foundation-git
│  ├─ moai-foundation-specs
│  ├─ moai-foundation-ears
│  ├─ moai-foundation-tags
│  ├─ moai-foundation-langs
│  └─ moai-foundation-trust

Tier 2: Essentials (MoAI-ADK 핵심)
├─ moai-essentials-*
│  ├─ moai-essentials-review
│  ├─ moai-essentials-debug
│  └─ moai-essentials-perf

Tier 3: Alfred (슈퍼에이전트 인프라)
├─ moai-alfred-*
│  ├─ moai-alfred-git-workflow
│  ├─ moai-alfred-language-detection
│  ├─ moai-alfred-spec-metadata-validation
│  ├─ moai-alfred-ears-authoring
│  ├─ moai-alfred-trust-validation
│  ├─ moai-alfred-tag-scanning
│  └─ moai-alfred-interactive-questions

Tier 4: Domain (비즈니스 도메인 전문성)
├─ moai-domain-*
│  ├─ moai-domain-frontend
│  ├─ moai-domain-backend
│  ├─ moai-domain-database
│  ├─ moai-domain-web-api
│  ├─ moai-domain-security
│  ├─ moai-domain-devops
│  ├─ moai-domain-ml
│  ├─ moai-domain-data-science
│  ├─ moai-domain-mobile-app
│  └─ moai-domain-cli-tool

Tier 5: Language (프로그래밍 언어 패턴)
├─ moai-language-*
│  ├─ moai-language-python
│  ├─ moai-language-typescript
│  ├─ moai-language-javascript
│  ├─ moai-language-go
│  ├─ moai-language-rust
│  ├─ moai-language-java
│  ├─ moai-language-csharp
│  ├─ moai-language-kotlin
│  ├─ moai-language-ruby
│  ├─ moai-language-php
│  ├─ moai-language-swift
│  ├─ moai-language-c
│  ├─ moai-language-cpp
│  ├─ moai-language-r
│  ├─ moai-language-dart
│  ├─ moai-language-scala
│  ├─ moai-language-sql
│  └─ moai-language-shell

Tier 6: Framework (웹 프레임워크 & 라이브러리)
├─ moai-framework-*
│  ├─ moai-framework-nextjs-advanced       ← 변경: moai-lang-nextjs-advanced에서
│  ├─ moai-framework-react-19              ← 변경: moai-lang-react-19에서
│  ├─ moai-framework-fastapi-patterns      ← 변경: moai-lang-fastapi-patterns에서
│  └─ [향후] moai-framework-django-*
│  └─ [향후] moai-framework-nestjs-*
│  └─ [향후] moai-framework-astro-*

Tier 7: Design (디자인 시스템 & 도구)
├─ moai-design-*
│  ├─ moai-design-tailwind-v4
│  ├─ moai-design-shadcn-ui
│  ├─ moai-design-figma-to-code            ← 신규
│  └─ moai-design-figma-mcp                ← 신규

Tier 8: SaaS (공식 MCP 기반 플랫폼)
├─ moai-saas-*
│  ├─ moai-saas-vercel-mcp                 ← 변경: moai-deploy-vercel에서
│  ├─ moai-saas-supabase-mcp               ← 변경: moai-deploy-supabase에서
│  ├─ moai-saas-render-mcp                 ← 변경: moai-deploy-render에서
│  ├─ moai-saas-wordpress-publishing       ← 신규
│  ├─ moai-saas-naver-blog-publishing      ← 신규
│  └─ moai-saas-tistory-publishing         ← 신규

Tier 9: Content (콘텐츠 & 마케팅)
├─ moai-content-*
│  ├─ moai-content-seo-optimization        ← 신규
│  ├─ moai-content-image-generation        ← 신규
│  └─ moai-content-blog-strategy           ← 신규

Tier 10: CloudCode (Claude Code 시스템)
├─ moai-cc-*
│  ├─ moai-cc-agents
│  ├─ moai-cc-commands
│  ├─ moai-cc-skills
│  ├─ moai-cc-hooks
│  ├─ moai-cc-mcp-plugins
│  ├─ moai-cc-claude-md
│  ├─ moai-cc-settings
│  └─ moai-cc-memory

Tier 11: Spec & Authoring
├─ moai-spec-*
│  ├─ moai-spec-authoring
│  └─ moai-skill-factory

❌ 삭제 대상 (PM 플러그인 제거)
├─ moai-pm-charter
└─ moai-pm-risk-matrix
```

### Phase 1 구현 작업

**스킬 재분류 (10개 스킬)**:

| 현재 이름 | 변경될 이름 | Tier | 카테고리 | 유형 |
|---|---|---|---|---|
| moai-lang-nextjs-advanced | **moai-framework-nextjs-advanced** | 6 | Framework | 웹 프레임워크 |
| moai-lang-react-19 | **moai-framework-react-19** | 6 | Framework | UI 라이브러리 |
| moai-lang-fastapi-patterns | **moai-framework-fastapi-patterns** | 6 | Framework | 웹 프레임워크 |
| moai-design-shadcn-ui | ✅ 유지 | 7 | Design | 컴포넌트 라이브러리 |
| moai-design-tailwind-v4 | ✅ 유지 | 7 | Design | CSS 프레임워크 |
| moai-deploy-vercel | **moai-saas-vercel-mcp** | 8 | SaaS | 배포 플랫폼 |
| moai-deploy-supabase | **moai-saas-supabase-mcp** | 8 | SaaS | 데이터베이스 플랫폼 |
| moai-deploy-render | **moai-saas-render-mcp** | 8 | SaaS | 배포 플랫폼 |
| moai-pm-charter | ❌ 삭제 | — | PM | 제거됨 |
| moai-pm-risk-matrix | ❌ 삭제 | — | PM | 제거됨 |

**신규 스킬 생성**:
1. moai-design-figma-to-code (Tier 7)
2. moai-design-figma-mcp (Tier 7)
3. moai-saas-wordpress-publishing (Tier 8)
4. moai-saas-naver-blog-publishing (Tier 8)
5. moai-saas-tistory-publishing (Tier 8)
6. moai-content-seo-optimization (Tier 9)
7. moai-content-image-generation (Tier 9)
8. moai-content-blog-strategy (Tier 9)

---

## 🎨 Phase 2: UI/UX 플러그인 확장

### 현재 상태 (단순화됨)

**UI/UX 플러그인 v1.0**:
- 3개 에이전트 (Design System Architect, Component Builder, Accessibility Specialist)
- 1개 명령어: `/setup-shadcn-ui`
- 기본 Tailwind + shadcn/ui 초기화

### 목표 상태: 디자인 자동화 플러그인 v2.0

**UI/UX 플러그인 v2.0 (확장됨)**:
- 7개 에이전트 (3개에서 확대)
- 6개 이상의 명령어 및 내부 오케스트레이션
- Figma MCP 통합으로 디자인-투-코드 자동화
- 플러그인 내부 오케스트레이션 흐름

### 에이전트 팀 구조 (7개 총합)

#### 기존 에이전트 (업그레이드)
1. **Design System Architect** (Sonnet)
   - 역할: 디자인 토큰 및 테마 전략 정의
   - 스킬: `moai-domain-frontend`, `moai-design-tailwind-v4`, `moai-design-shadcn-ui`
   - 책임:
     - 디자인 요구사항 분석 및 토큰 스펙 작성
     - 색상 시스템, 타이포그래피 스케일, 스페이싱 그리드 정의
     - Tailwind 설정 및 테마 파일 생성

2. **Component Builder** (Haiku)
   - 역할: 재사용 가능한 접근성 컴포넌트 생성
   - 스킬: `moai-domain-frontend`, `moai-design-shadcn-ui`, `moai-essentials-perf`
   - 책임:
     - 디자인 스펙에서 React 컴포넌트 구축
     - 컴포넌트 재사용성 및 조합 보장
     - 렌더링 성능 최적화

3. **Accessibility Specialist** (Haiku)
   - 역할: WCAG 2.1 AA 준수 보장
   - 스킬: `moai-domain-security`, `moai-domain-frontend`
   - 책임:
     - WCAG 2.1 AA 기준 검증
     - 키보드 네비게이션 및 스크린 리더 테스트
     - 색상 대비 및 시맨틱 HTML 감시

#### 신규 에이전트 (Figma 통합)
4. **Design Strategist** (Sonnet)
   - 역할: 사용자 지시사항 분석 및 디자인 스펙 작성
   - 스킬: `moai-design-figma-mcp`, `moai-domain-frontend`, `moai-essentials-review`
   - 책임:
     - `/ui-ux "사용자 지시사항"` 요청 파싱
     - 자연어에서 디자인 스펙 작성
     - 구현 전략 계획
     - 적절한 에이전트로 위임

5. **Figma Designer** (Haiku)
   - 역할: Figma MCP 통합 및 디자인 파일 관리
   - 스킬: `moai-design-figma-mcp`, `moai-design-figma-to-code`
   - 책임:
     - Figma MCP에 연결
     - 디자인 파일 생성/파싱
     - 디자인 토큰 및 컴포넌트 추출
     - 디자인 변경사항과 코드 동기화

6. **CSS/HTML Generator** (Haiku)
   - 역할: 디자인에서 프로덕션 코드 자동 생성
   - 스킬: `moai-design-figma-to-code`, `moai-language-typescript`, `moai-framework-react-19`
   - 책임:
     - Figma 디자인을 React 컴포넌트로 변환
     - Tailwind CSS 스타일 생성
     - 정적 페이지의 HTML/CSS 생성
     - 생성된 코드 검증

7. **Design Documentation Writer** (Haiku)
   - 역할: 디자인 시스템 문서 작성
   - 스킬: `moai-domain-frontend`, `moai-essentials-review`
   - 책임:
     - 컴포넌트 가이드 및 API 문서 생성
     - 디자인 토큰 레퍼런스 작성
     - 접근성 구현 문서화
     - 디자인 시스템 변경로그 관리

### 명령어 구조 및 내부 오케스트레이션

```yaml
# 플러그인 레벨 명령어 라우팅: /ui-ux "지시사항" → 내부 위임

명령어:
  - 이름: "ui-ux"
    설명: "UI/UX 지시사항 처리기"
    파라미터:
      - 이름: "directive"
        타입: "string"
        설명: "사용자의 디자인 요청 (자연어)"
    내부_오케스트레이션: true
    위임_로직:
      패턴: "지시사항 파싱 → 전문 에이전트로 라우팅"

  # 전문화된 명령어
  - 이름: "design"
    설명: "요구사항에서 디자인 스펙 작성"
    위임_대상: ["Design Strategist"]
    예시: "/design 분석 대시보드 레이아웃 만들기"

  - 이름: "figma-to-code"
    설명: "Figma 파일을 React 컴포넌트로 변환"
    위임_대상: ["Figma Designer", "CSS/HTML Generator"]
    예시: "/figma-to-code https://figma.com/file/..."

  - 이름: "setup-design-system"
    설명: "디자인 토큰 및 테마 초기화"
    위임_대상: ["Design System Architect"]
    예시: "/setup-design-system Tailwind 토큰 초기화"

  - 이름: "add-component"
    설명: "디자인 스펙에서 새 컴포넌트 생성"
    위임_대상: ["Component Builder", "Accessibility Specialist"]
    예시: "/add-component Button primary 변형"

  - 이름: "generate-design-guide"
    설명: "디자인 시스템 문서 생성"
    위임_대상: ["Design Documentation Writer"]
    예시: "/generate-design-guide 컴포넌트 API 레퍼런스"

  - 이름: "create-prototype"
    설명: "인터랙티브 프로토타입 생성"
    위임_대상: ["Design Strategist", "CSS/HTML Generator"]
    예시: "/create-prototype 랜딩 페이지 프로토타입 만들기"

  - 이름: "setup-shadcn-ui"
    설명: "shadcn/ui 초기화 (레거시 호환성)"
    위임_대상: ["Design System Architect", "Component Builder"]
    예시: "/setup-shadcn-ui 커스텀 테마로 초기화"
```

### 내부 오케스트레이션 흐름

```
사용자 입력: /ui-ux "반응형 대시보드 컴포넌트 만들기"
    ↓
UI/UX 플러그인 진입점
    ↓
Design Strategist (에이전트 4) 지시사항 분석
    ├─ 이는 디자인 + 컴포넌트 작업임을 판단
    └─ 분해: 디자인 스펙 + 컴포넌트 코드
    ↓
병렬 실행:
├─ Design System Architect (에이전트 1) 토큰 & 스타일 작성
├─ Component Builder (에이전트 2) 컴포넌트 생성
└─ Accessibility Specialist (에이전트 3) 준수 검증
    ↓
CSS/HTML Generator (에이전트 6) 프로덕션 코드 출력
    ↓
Design Documentation Writer (에이전트 7) 컴포넌트 문서화
    ↓
Figma Designer (에이전트 5) 선택적으로 Figma 파일과 동기화
    ↓
완성된 디자인 → 코드 출력을 Alfred로 반환
```

### UI/UX 플러그인 설정 업데이트

```json
{
  "id": "moai-plugin-uiux",
  "name": "UI/UX 플러그인",
  "version": "2.0.0-dev",
  "description": "디자인 시스템 자동화 - Figma MCP + 디자인-투-코드 + 디자인 토큰 + 컴포넌트 라이브러리",
  "tags": ["design-system", "tailwind", "shadcn-ui", "figma", "design-automation", "accessibility"],
  "commands": [
    {
      "name": "ui-ux",
      "description": "일반 UI/UX 지시사항을 내부 오케스트레이션으로 처리"
    },
    {
      "name": "design",
      "description": "요구사항에서 디자인 스펙 작성"
    },
    {
      "name": "figma-to-code",
      "description": "Figma 파일을 React 컴포넌트로 변환"
    },
    {
      "name": "setup-design-system",
      "description": "디자인 토큰 및 테마 초기화"
    },
    {
      "name": "add-component",
      "description": "디자인 스펙에서 새 컴포넌트 생성"
    },
    {
      "name": "generate-design-guide",
      "description": "디자인 시스템 문서 생성"
    },
    {
      "name": "create-prototype",
      "description": "인터랙티브 프로토타입 생성"
    },
    {
      "name": "setup-shadcn-ui",
      "description": "shadcn/ui 초기화 (레거시 지원)"
    }
  ],
  "agents": 7,
  "mcp_integrations": ["figma"],
  "permissions": {
    "allowedTools": ["Read", "Write", "Edit", "Bash", "Task"],
    "deniedTools": ["DeleteFile"]
  }
}
```

---

## 📝 Phase 3: 블로그 작성 플러그인 (신규)

### 플러그인 목적

다중 플랫폼에서 블로그 작성 워크플로우 전체를 관리하는 콘텐츠 생성 플러그인입니다. SEO 최적화 및 이미지 생성 지원을 포함합니다.

### 지원 플랫폼

1. **WordPress** - 국제 블로깅 플랫폼
2. **네이버블로그** - 한국 블로깅 플랫폼
3. **티스토리** - 한국 블로깅 플랫폼

### 에이전트 팀 구조 (7개 에이전트)

1. **Content Strategist** (Sonnet)
   - 역할: 전략적 계획 및 콘텐츠 방향 설정
   - 스킬: `moai-content-blog-strategy`, `moai-domain-data-science`
   - 책임:
     - 블로그 전략 및 대상 읽자층 분석
     - 콘텐츠 캘린더 계획
     - 콘텐츠 공백 식별
     - 발행 일정 설정

2. **SEO Specialist** (Haiku)
   - 역할: 검색 엔진 최적화
   - 스킬: `moai-content-seo-optimization`, `moai-domain-web-api`
   - 책임:
     - 키워드 및 검색 의도 연구
     - 메타 태그 및 제목 최적화
     - SEO 친화적인 URL 작성
     - 경쟁사 콘텐츠 분석
     - 스키마 마크업 생성

3. **Image Prompt Generator** (Haiku)
   - 역할: AI 이미지 생성 프롬프트 작성
   - 스킬: `moai-content-image-generation`
   - 책임:
     - 상세한 이미지 생성 프롬프트 작성
     - 다중 이미지 생성기 지원:
       - GPT-Image-1 (OpenAI)
       - 나노바나나 (한국 AI 이미지)
       - Midjourney (상용 AI 이미지)
     - 이미지 스타일 일관성 유지
     - 여러 변형 생성

4. **Content Writer** (Haiku)
   - 역할: 블로그 포스트 작성 및 편집
   - 스킬: `moai-language-typescript`, `moai-domain-web-api`
   - 책임:
     - 매력적인 블로그 포스트 작성
     - 일관된 톤 및 스타일 유지
     - 다중 버전 작성 (영어, 한국어)
     - SEO 권장사항 구현
     - 읽기 편의성을 위한 콘텐츠 포맷팅

5. **Platform Publisher** (Haiku)
   - 역할: 다중 플랫폼 발행 자동화
   - 스킬: `moai-saas-wordpress-publishing`, `moai-saas-naver-blog-publishing`, `moai-saas-tistory-publishing`
   - 책임:
     - WordPress REST API를 통한 발행
     - 네이버블로그 발행
     - 티스토리 발행
     - 플랫폼 간 포스트 스케줄링
     - 플랫폼별 메타데이터 관리

6. **Knowledge Manager** (Haiku)
   - 역할: llms.txt 및 지식 베이스 관리
   - 스킬: `moai-content-blog-strategy`, `moai-essentials-review`
   - 책임:
     - llms.txt 지식 베이스 유지보수
     - 콘텐츠 참고자료 및 인용문 관리
     - 발행된 콘텐츠 추적
     - 콘텐츠 인덱스 작성
     - 발견 및 재사용 지원

7. **Content Curator** (Haiku)
   - 역할: 콘텐츠 소스 및 영감 수집
   - 스킬: `moai-content-blog-strategy`, `moai-domain-data-science`
   - 책임:
     - 업계 트렌드 모니터링
     - 연구 자료 수집
     - 콘텐츠 소스 관리
     - 콘텐츠 브리프 작성
     - 아이디어 발상 및 계획 지원

### 블로그 플러그인 아키텍처

```
사용자 입력: /blog "React 19 베스트 프랙티스 작성"
    ↓
Content Strategist (에이전트 1)
├─ 주제 및 대상 읽자층 분석
├─ 플랫폼 결정 (WordPress, 네이버블로그, 티스토리)
└─ 콘텐츠 브리프 작성
    ↓
병렬 처리:
├─ SEO Specialist (에이전트 2)
│  ├─ 키워드 연구: ["React 19", "hooks", "server components"]
│  ├─ 경쟁사 콘텐츠 분석
│  └─ SEO 전략 작성
├─ Image Prompt Generator (에이전트 3)
│  ├─ 3-5개 이미지 프롬프트 생성
│  └─ 지원: GPT-Image-1, 나노바나나, Midjourney
└─ Content Writer (에이전트 4)
   ├─ 블로그 포스트 초안 작성
   ├─ SEO 권장사항 구현
   └─ 버전 생성: 영어 & 한국어
    ↓
Content Curator (에이전트 7)
├─ 참고 자료 및 소스 수집
├─ 지식 베이스에 추가
└─ 인용문 목록 작성
    ↓
Knowledge Manager (에이전트 6)
├─ 신규 콘텐츠로 llms.txt 업데이트
├─ 향후 참고를 위해 인덱싱
└─ 콘텐츠 링크 생성
    ↓
Platform Publisher (에이전트 5)
├─ WordPress에 발행
├─ 네이버블로그에 발행
├─ 티스토리에 발행
└─ 플랫폼 간 스케줄 설정
    ↓
Alfred로 발행 확인 반환
```

### 명령어 구조

```yaml
명령어:
  - 이름: "blog"
    설명: "블로그 작성 요청 처리"
    파라미터:
      - 이름: "topic"
        타입: "string"
        설명: "블로그 주제 또는 요구사항"

  - 이름: "blog-strategy"
    설명: "콘텐츠 캘린더 및 전략 계획"
    위임_대상: ["Content Strategist"]

  - 이름: "blog-write"
    설명: "주제에서 블로그 포스트 작성"
    위임_대상: ["Content Writer", "SEO Specialist"]

  - 이름: "blog-optimize-seo"
    설명: "포스트 SEO 최적화"
    위임_대상: ["SEO Specialist"]

  - 이름: "blog-generate-images"
    설명: "이미지 프롬프트 생성"
    위임_대상: ["Image Prompt Generator"]

  - 이름: "blog-publish"
    설명: "플랫폼으로 발행"
    위임_대상: ["Platform Publisher"]

  - 이름: "blog-manage-knowledge"
    설명: "llms.txt 지식 베이스 업데이트"
    위임_대상: ["Knowledge Manager"]

  - 이름: "blog-curate"
    설명: "콘텐츠 영감 수집"
    위임_대상: ["Content Curator"]
```

### 블로그 플러그인 설정

```json
{
  "id": "moai-plugin-blog",
  "name": "블로그 작성 플러그인",
  "version": "1.0.0-dev",
  "status": "development",
  "description": "콘텐츠 생성 및 다중 플랫폼 발행 - WordPress, 네이버블로그, 티스토리, SEO, 이미지 생성, llms.txt",
  "author": "GOOS🪿",
  "category": "content",
  "tags": ["blogging", "wordpress", "korean-blogs", "seo", "image-generation", "content-creation"],
  "repository": "https://github.com/moai-adk/moai-alfred-marketplace/tree/main/plugins/moai-plugin-blog",
  "documentation": "https://github.com/moai-adk/moai-alfred-marketplace/blob/main/plugins/moai-plugin-blog/README.md",
  "minClaudeCodeVersion": "1.0.0",
  "commands": [
    {
      "name": "blog",
      "description": "블로그 작성 요청 처리"
    },
    {
      "name": "blog-strategy",
      "description": "콘텐츠 캘린더 및 전략 계획"
    },
    {
      "name": "blog-write",
      "description": "주제에서 블로그 포스트 작성"
    },
    {
      "name": "blog-optimize-seo",
      "description": "포스트 SEO 최적화"
    },
    {
      "name": "blog-generate-images",
      "description": "이미지 프롬프트 생성"
    },
    {
      "name": "blog-publish",
      "description": "플랫폼으로 발행"
    },
    {
      "name": "blog-manage-knowledge",
      "description": "llms.txt 지식 베이스 업데이트"
    },
    {
      "name": "blog-curate",
      "description": "콘텐츠 영감 수집"
    }
  ],
  "agents": 7,
  "platforms": ["wordpress", "naver-blog", "tistory"],
  "image_generators": ["gpt-image-1", "nanobanana", "midjourney"],
  "permissions": {
    "allowedTools": ["Read", "Write", "Edit", "Bash", "Task", "WebSearch", "WebFetch"],
    "deniedTools": []
  },
  "dependencies": [],
  "installCommand": "/plugin install moai-plugin-blog",
  "releaseNotes": "초기 v1.0.0-dev - 다중 플랫폼 지원"
}
```

### 블로그 스킬 (8개 신규 스킬)

1. **moai-content-blog-strategy** - 블로그 전략, 계획, 대상 읽자층 분석
2. **moai-content-seo-optimization** - SEO 기법, 키워드 연구, 최적화
3. **moai-content-image-generation** - GPT-Image-1, 나노바나나, Midjourney용 이미지 프롬프트
4. **moai-saas-wordpress-publishing** - WordPress REST API 통합
5. **moai-saas-naver-blog-publishing** - 네이버블로그 API 통합
6. **moai-saas-tistory-publishing** - 티스토리 API 통합
7. **moai-content-markdown-to-blog** - 다양한 플랫폼용 포맷 변환
8. **moai-content-llms-txt-management** - llms.txt 지식 베이스 관리

---

## 📊 변경 사항 요약

### 스킬 변경 (18개 총합)

**재분류 (8개)**:
- 3개 Framework 스킬: nextjs-advanced, react-19, fastapi-patterns
- 3개 SaaS 스킬: vercel, supabase, render
- 2개 삭제: pm-charter, pm-risk-matrix

**신규 스킬 (10개)**:
- 2개 Design 스킬: figma-mcp, figma-to-code
- 3개 SaaS 스킬: wordpress, naver-blog, tistory
- 3개 Content 스킬: seo-optimization, image-generation, blog-strategy
- 2개 Content/Knowledge 스킬: markdown-to-blog, llms-txt-management

### 플러그인 변경

**업데이트 플러그인 (1개)**:
- **UI/UX 플러그인**: 3개 에이전트 → 7개 에이전트, 1개 명령어 → 8개 명령어

**신규 플러그인 (1개)**:
- **블로그 플러그인**: 7개 에이전트, 8개 명령어, 3개 플랫폼, 3개 이미지 생성기

**플러그인 수**: 4개 플러그인 → 5개 플러그인

### 에이전트 수

**총 에이전트**: 15개 → 19개 (Phase 3 + 확장)
- Frontend: 4개 에이전트
- Backend: 4개 에이전트
- UI/UX: 7개 에이전트 (3개에서 확대)
- DevOps: 4개 에이전트
- Blog: 7개 에이전트

---

## ✅ 승인 체크리스트

구현 전에 다음을 확인하시기 바랍니다:

- [ ] **스킬 분류법**: 정확한 분류 승인? (Framework vs Language vs Design vs SaaS)
- [ ] **UI/UX 플러그인**: 7개 에이전트 Figma MCP 통합 설계 승인?
- [ ] **블로그 플러그인**: WordPress/한국 블로그/SEO/이미지 5번째 플러그인 승인?
- [ ] **공식 MCP**: Vercel, Supabase, Render MCP 확인?
- [ ] **명명 규칙**: 모든 Tier 명명 규칙 승인?

---

**다음 단계** (승인 후):
1. 기존 10개 스킬 이름 변경/재조직
2. 신규 10개 스킬 생성
3. UI/UX 에이전트 팀 설계 및 구현
4. 블로그 플러그인 전체 구조 구현
5. marketplace.json 업데이트
6. Phase 4: 통합 테스트

