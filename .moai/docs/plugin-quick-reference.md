# MoAI-Alfred 플러그인 빠른 참조 가이드

**작성일**: 2025-10-31
**마켓플레이스**: `moai-marketplace` (로컬)

---

## 🎯 플러그인 선택 가이드

어떤 플러그인이 필요한가요? 다음 테이블에서 찾으세요:

| 상황 | 추천 플러그인 | 명령어 | 에이전트 |
|-----|------------|--------|---------|
| **FastAPI 백엔드 시작** | moai-plugin-backend | `/init-fastapi` | Backend Architect |
| **Next.js 프론트엔드 시작** | moai-plugin-frontend | `/init-next` | 없음 |
| **Figma 디자인 → 코드** | moai-plugin-uiux | `/setup-shadcn-ui` | Design Strategist |
| **Vercel/Supabase 배포** | moai-plugin-devops | `/deploy-config` | Deployment Strategist |
| **기술 블로그 작성** | moai-plugin-technical-blog | `/blog-write` | Content Strategist |

---

## 1️⃣ Backend Plugin (moai-plugin-backend)

### 정보
- **버전**: 1.0.0-dev
- **상태**: 개발 중
- **에이전트**: 4개 (Backend Architect, FastAPI Specialist, API Designer, Database Expert)
- **명령어**: 3개

### 설치

```bash
# 마켓플레이스 추가 (처음 1회만)
/plugin marketplace add /Users/goos/MoAI/MoAI-ADK-v1.0/moai-marketplace

# 플러그인 설치
/plugin install moai-plugin-backend@moai-marketplace
```

### 주요 명령어

#### `/init-fastapi` - FastAPI 프로젝트 초기화

```bash
/init-fastapi

# 대화형 입력
Project name: my_api
Python version: 3.13
Database: PostgreSQL
```

**생성물**:
- FastAPI 앱 구조 (app/, models/, schemas/, api/)
- pyproject.toml (uv 설정)
- Alembic 마이그레이션 폴더
- .env 템플릿

#### `/db-setup` - 데이터베이스 설정

```bash
/db-setup

# 선택 옵션
Database: PostgreSQL / MySQL
Host: localhost
Database name: my_db
```

**생성물**:
- Alembic 마이그레이션 설정
- .env 파일 (DB_URL)
- 초기 마이그레이션

#### `/resource-crud` - CRUD 엔드포인트 생성

```bash
/resource-crud

# 입력
Resource name: User
Fields:
  - name: string (required)
  - email: string (unique, required)
  - age: integer (optional)
```

**생성물**:
- SQLAlchemy 모델
- Pydantic 스키마
- CRUD 라우터 (GET, POST, PUT, DELETE)

### 에이전트 활용

```python
# Backend Architect 호출
Task(
  subagent_type="backend-architect",
  prompt="Design FastAPI microservice for e-commerce with users, products, orders"
)

# FastAPI Specialist 호출
Task(
  subagent_type="fastapi-specialist",
  prompt="Optimize API request validation and error handling"
)

# Database Expert 호출
Task(
  subagent_type="database-expert",
  prompt="Design database schema for user relationships"
)

# API Designer 호출
Task(
  subagent_type="api-designer",
  prompt="Create REST API endpoints following OpenAPI 3.1"
)
```

### 스킬

- `moai-lang-fastapi-patterns` - FastAPI 비동기 패턴
- `moai-lang-python` - Python 3.13+ 베스트 프랙티스
- `moai-domain-backend` - 백엔드 아키텍처
- `moai-domain-database` - 데이터베이스 설계

---

## 2️⃣ Frontend Plugin (moai-plugin-frontend)

### 정보
- **버전**: 1.0.0-dev
- **상태**: 개발 중
- **에이전트**: 0개 (명령어 기반)
- **명령어**: 3개 (Playwright-MCP 포함)

### 설치

```bash
/plugin install moai-plugin-frontend@moai-marketplace
```

### 주요 명령어

#### `/init-next` - Next.js 프로젝트 초기화

```bash
/init-next

# 대화형 입력
Project name: my_app
Package manager: npm / pnpm / bun
TypeScript: Yes
Tailwind CSS: Yes
```

**생성물**:
- Next.js 16 App Router 구조
- React 19.2 컴포넌트 예제
- Biome 설정 (linting/formatting)
- shadcn/ui 통합 준비

#### `/biome-setup` - Biome 린터 설정

```bash
/biome-setup

# 설정 자동 생성
# - biome.json
# - package.json scripts
# - lint/format 명령어
```

#### `/playwright-setup` - Playwright-MCP E2E 테스트 초기화

```bash
/playwright-setup

# Playwright-MCP 설정
# - playwright.config.ts 생성
# - 예제 테스트 파일
# - GitHub Actions 통합
# - MCP 서버 설정
```

**생성물**:
- Playwright 설정 및 테스트 디렉토리
- MCP 기반 자동화된 E2E 테스트 프레임워크
- CI/CD 파이프라인 설정

### 스킬

- `moai-framework-nextjs-advanced` - Next.js 16 고급 패턴
- `moai-framework-react-19` - React 19.2 패턴
- `moai-design-shadcn-ui` - shadcn/ui 컴포넌트
- `moai-domain-frontend` - 프론트엔드 아키텍처
- `moai-testing-playwright-mcp` - Playwright-MCP E2E 테스트

---

## 3️⃣ DevOps Plugin (moai-plugin-devops)

### 정보
- **버전**: 2.0.0-dev
- **상태**: 개발 중
- **에이전트**: 4개 (Deployment Strategist, Vercel Specialist, Supabase Specialist, Render Specialist)
- **명령어**: 4개

### 설치

```bash
/plugin install moai-plugin-devops@moai-marketplace
```

### 주요 명령어

#### `/deploy-config` - 배포 설정

```bash
/deploy-config

# 다중 클라우드 선택
Select platforms:
  ☑ Vercel (Frontend)
  ☑ Supabase (Database)
  ☑ Render (Backend)
```

#### `/connect-vercel` - Vercel 연동

```bash
/connect-vercel

# 입력 필요
Vercel Token: (GitHub 계정 인증)
Project name: my_frontend
```

**생성물**:
- vercel.json
- GitHub Actions 워크플로우
- 프리뷰 환경 설정

#### `/connect-supabase` - Supabase 연동

```bash
/connect-supabase

# 입력 필요
Supabase URL: https://xxxx.supabase.co
Supabase Key: your-anon-key
```

**생성물**:
- .env.local (Supabase 설정)
- PostgreSQL 마이그레이션 스크립트
- Row Level Security (RLS) 정책

#### `/connect-render` - Render 연동

```bash
/connect-render

# 입력 필요
Render Service: Backend FastAPI
GitHub Token: (배포용 토큰)
```

**생성물**:
- render.yaml (배포 설정)
- health check 엔드포인트
- 환경 변수 설정

### 에이전트

```python
# Deployment Strategist
Task(
  subagent_type="deployment-strategist",
  prompt="Design multi-cloud deployment architecture"
)

# Vercel Specialist
Task(
  subagent_type="vercel-specialist",
  prompt="Optimize Next.js deployment on Vercel"
)

# Supabase Specialist
Task(
  subagent_type="supabase-specialist",
  prompt="Setup PostgreSQL database and authentication"
)

# Render Specialist
Task(
  subagent_type="render-specialist",
  prompt="Deploy FastAPI backend on Render"
)
```

### 스킬

- `moai-saas-vercel-mcp` - Vercel MCP 통합
- `moai-saas-supabase-mcp` - Supabase MCP 통합
- `moai-saas-render-mcp` - Render MCP 통합

---

## 4️⃣ UI/UX Plugin (moai-plugin-uiux)

### 정보
- **버전**: 2.0.0-dev
- **상태**: 개발 중
- **에이전트**: 7개 (Design Strategist, Design System Architect, Component Builder, Figma Designer, CSS/HTML Generator, Accessibility Specialist, Design Documentation Writer)
- **명령어**: 3개

### 설치

```bash
/plugin install moai-plugin-uiux@moai-marketplace
```

### 주요 명령어

#### `/ui-ux` - 디자인 지시 오케스트레이션

```bash
/ui-ux

# 입력
Design task: Create a user dashboard layout
Components needed:
  - Header with navigation
  - Sidebar menu
  - Content area with cards
  - Footer
```

#### `/setup-shadcn-ui` - shadcn/ui 초기화

```bash
/setup-shadcn-ui

# 자동 설정
# - Tailwind CSS v4 설정
# - shadcn/ui 컴포넌트 추가
# - 예제 컴포넌트
```

#### `/design-tokens` - 디자인 토큰 관리

```bash
/design-tokens

# Figma 연동 (MCP 필요)
# - 색상 팔레트 추출
# - 타이포그래피 설정
# - 간격 스케일
# - 그림자 설정
```

### 에이전트

```python
# Design Strategist
Task(
  subagent_type="design-strategist",
  prompt="Design a modern SaaS dashboard with accessibility focus"
)

# Component Builder
Task(
  subagent_type="component-builder",
  prompt="Create reusable React components for form inputs"
)

# Figma Designer (MCP 연동 필요)
Task(
  subagent_type="figma-designer",
  prompt="Extract design tokens from Figma and generate CSS"
)
```

### 스킬

- `moai-design-figma-mcp` - Figma MCP 통합
- `moai-design-figma-to-code` - Figma → 코드 변환
- `moai-design-shadcn-ui` - shadcn/ui 패턴
- `moai-design-tailwind-v4` - Tailwind CSS v4
- `moai-domain-frontend` - 프론트엔드 아키텍처

---

## 5️⃣ Technical Blog Plugin (moai-plugin-technical-blog)

### 정보
- **버전**: 2.0.0-dev
- **상태**: 개발 중
- **에이전트**: 7개 (Content Strategist, Technical Writer, SEO Specialist, Code Example Curator, Visual Designer, Markdown Formatter, Template Coordinator)
- **명령어**: 1개 (통합)

### 설치

```bash
/plugin install moai-plugin-technical-blog@moai-marketplace
```

### 주요 명령어

#### `/blog-write` - 기술 블로그 작성

```bash
/blog-write

# 자연어 입력 (자동 템플릿 선택)
Write a blog post about "Getting started with FastAPI and PostgreSQL"

# 또는 템플릿 명시
Template: Tutorial
Topic: FastAPI Database Integration
Audience: Python developers
Code examples: Yes (Python, SQL)
```

**자동 처리**:
- ✅ 템플릿 자동 선택 (Tutorial/Case Study/How-to/Announcement/Comparison)
- ✅ SEO 최적화 (메타 태그, 헤딩 구조)
- ✅ 코드 예제 생성 및 검증
- ✅ 이미지 프롬프트 생성
- ✅ Markdown 포맷팅
- ✅ OG 이미지 스펙 생성

### 5개 템플릿

| 템플릿 | 용도 | 구조 |
|-------|------|------|
| **Tutorial** | 단계별 학습 | 개요 → 전제조건 → 단계 → 결론 |
| **Case Study** | 실제 사례 | 도전과제 → 해결책 → 결과 |
| **How-to** | 문제 해결 | 문제 → 해결책 → 검증 |
| **Announcement** | 뉴스/업데이트 | 뉴스 → 영향 → 행동 |
| **Comparison** | 기술 비교 | 개요 → 비교 → 권장사항 |

### 에이전트

```python
# Content Strategist
Task(
  subagent_type="technical-content-strategist",
  prompt="Plan blog content strategy for Q4 2025 targeting Python developers"
)

# Technical Writer
Task(
  subagent_type="technical-writer",
  prompt="Write a comprehensive FastAPI tutorial"
)

# SEO Specialist
Task(
  subagent_type="seo-specialist",
  prompt="Optimize blog post for search engines"
)

# Code Example Curator
Task(
  subagent_type="code-example-curator",
  prompt="Generate runnable FastAPI code examples"
)

# Visual Designer
Task(
  subagent_type="visual-designer",
  prompt="Create diagrams and OG image prompts"
)
```

### 스킬

- `moai-content-technical-writing` - 기술 글쓰기
- `moai-content-seo-optimization` - SEO 최적화
- `moai-content-code-examples` - 코드 예제
- `moai-content-blog-templates` - 블로그 템플릿
- `moai-content-image-generation` - AI 이미지 생성
- `moai-content-markdown-to-blog` - 마크다운 변환

---

## 🔄 플러그인 조합 시나리오

### 시나리오 1: 전체 스택 애플리케이션

```bash
# 1. 백엔드 프로젝트 생성
/plugin install moai-plugin-backend@moai-marketplace
/init-fastapi
/db-setup
/resource-crud (User, Product, Order)

# 2. 프론트엔드 프로젝트 생성
/plugin install moai-plugin-frontend@moai-marketplace
/init-next

# 3. UI/UX 디자인 시스템 구축
/plugin install moai-plugin-uiux@moai-marketplace
/setup-shadcn-ui
/design-tokens

# 4. 배포 설정
/plugin install moai-plugin-devops@moai-marketplace
/deploy-config
/connect-vercel
/connect-supabase
/connect-render

# 5. 문서화
/plugin install moai-plugin-technical-blog@moai-marketplace
/blog-write (API 가이드)
/blog-write (배포 가이드)
```

### 시나리오 2: 빠른 프로토타입

```bash
# 1분 안에 완전한 API 구축
/init-fastapi
/db-setup
/resource-crud (User)

# 1분 안에 프론트엔드 구축
/init-next
/setup-shadcn-ui

# 배포 준비
/deploy-config
```

### 시나리오 3: 기술 블로그 출판 자동화

```bash
# 월간 기술 블로그 콘텐츠 생성
/blog-write "FastAPI v0.120 새 기능 설명"
/blog-write "React 19 마이그레이션 가이드"
/blog-write "PostgreSQL 성능 튜닝"

# 각 포스트마다:
# - SEO 최적화
# - 코드 예제 생성
# - OG 이미지 생성
# - 해시태그 자동 생성
```

---

## 📋 플러그인 설치 체크리스트

### 초기 설정 (1회)

```bash
# ✓ Step 1: 마켓플레이스 추가
/plugin marketplace add /Users/goos/MoAI/MoAI-ADK-v1.0/moai-marketplace

# ✓ Step 2: 마켓플레이스 확인
/plugin marketplace list

# ✓ Step 3: 플러그인 목록 확인
/plugin browse
```

### 플러그인별 테스트

**moai-plugin-backend**:
- [ ] `/init-fastapi` 실행
- [ ] `/db-setup` 실행
- [ ] `/resource-crud` 실행
- [ ] Backend Architect 에이전트 호출

**moai-plugin-frontend**:
- [ ] `/init-next` 실행
- [ ] `/biome-setup` 실행
- [ ] React 컴포넌트 생성

**moai-plugin-devops**:
- [ ] `/deploy-config` 실행
- [ ] `/connect-vercel` 실행
- [ ] Supabase 연동 설정

**moai-plugin-uiux**:
- [ ] `/setup-shadcn-ui` 실행
- [ ] 컴포넌트 빌더 에이전트 호출

**moai-plugin-technical-blog**:
- [ ] `/blog-write` 실행
- [ ] 5개 템플릿 테스트

---

## 🆘 자주 묻는 질문

### Q1: 플러그인을 어떻게 제거하나요?

```bash
/plugin uninstall moai-plugin-backend
```

### Q2: 여러 플러그인을 한 번에 설치하려면?

```bash
/plugin install moai-plugin-backend@moai-marketplace
/plugin install moai-plugin-frontend@moai-marketplace
/plugin install moai-plugin-devops@moai-marketplace
/plugin install moai-plugin-uiux@moai-marketplace
/plugin install moai-plugin-technical-blog@moai-marketplace
```

또는 `.claude/settings.json`에 추가

### Q3: 플러그인 업데이트는?

```bash
# 자동 업데이트 확인
/plugin update

# 특정 플러그인 업데이트
/plugin update moai-plugin-backend
```

### Q4: 로컬 플러그인 개발은?

1. 플러그인 폴더 생성
2. `plugin.json` 작성
3. `.claude-plugin/` 디렉토리 생성
4. 명령어/에이전트 추가
5. 로컬 마켓플레이스에 추가
6. 테스트

---

**최종 업데이트**: 2025-10-31
**플러그인 버전**: 1.0.0-2.0.0-dev
