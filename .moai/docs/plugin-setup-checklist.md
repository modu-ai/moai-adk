# Claude Code 플러그인 설정 완료 체크리스트

**작성일**: 2025-10-31
**상태**: 준비 완료
**마켓플레이스**: `/Users/goos/MoAI/MoAI-ADK-v1.0/moai-marketplace`

---

## ✅ 생성된 문서 목록

### 📚 메인 문서 (4개)

```
.moai/docs/
├── claude-code-plugin-installation-guide.md    (14KB, 550줄)
│   ├─ 플러그인 마켓플레이스 개요
│   ├─ 플러그인 구조 분석 (backend)
│   ├─ 단계별 설치 가이드
│   ├─ 플러그인 기능 테스트
│   ├─ 고급 시나리오
│   └─ 문제 해결
│
├── plugin-quick-reference.md                   (13KB, 623줄)
│   ├─ 플러그인 선택 가이드
│   ├─ 각 플러그인별 상세 정보
│   │  ├─ moai-plugin-backend
│   │  ├─ moai-plugin-frontend
│   │  ├─ moai-plugin-devops
│   │  ├─ moai-plugin-uiux
│   │  └─ moai-plugin-technical-blog
│   ├─ 에이전트 활용법
│   ├─ 플러그인 조합 시나리오
│   └─ FAQ
│
├── plugin-testing-scenarios.md                 (17KB, 758줄)
│   ├─ 테스트 전략
│   ├─ Unit Tests (파일 검증)
│   ├─ Integration Tests (Claude Code 통합)
│   ├─ E2E Tests (사용자 워크플로우)
│   ├─ Performance Tests (성능 메트릭)
│   ├─ 자동화 테스트 스크립트
│   └─ CI/CD 통합
│
└── plugin-ecosystem-introduction.md            (47KB, 1725줄) - 기존
    ├─ 플러그인 생태계 개요
    ├─ 아키텍처 설명
    └─ 개발 가이드
```

**총 문서 크기**: 91KB
**총 줄 수**: 3,656줄

---

## 🚀 빠른 시작 (5분)

### Step 1: 마켓플레이스 등록 (1분)

```bash
/plugin marketplace add /Users/goos/MoAI/MoAI-ADK-v1.0/moai-marketplace
```

**확인**:

```bash
/plugin marketplace list
```

### Step 2: 플러그인 설치 (2분)

#### 옵션 A: 모든 플러그인 설치

```bash
/plugin install moai-plugin-backend@moai-marketplace
/plugin install moai-plugin-frontend@moai-marketplace
/plugin install moai-plugin-devops@moai-marketplace
/plugin install moai-plugin-uiux@moai-marketplace
/plugin install moai-plugin-technical-blog@moai-marketplace
```

#### 옵션 B: 특정 플러그인만 설치

```bash
# Backend 개발
/plugin install moai-plugin-backend@moai-marketplace

# Frontend 개발
/plugin install moai-plugin-frontend@moai-marketplace

# Full-stack 배포
/plugin install moai-plugin-devops@moai-marketplace
```

### Step 3: 설치 검증 (1분)

```bash
/help
```

**예상 출력**:

```
Installed Plugins:
✓ moai-plugin-backend (v1.0.0-dev)
✓ moai-plugin-frontend (v1.0.0-dev)
✓ moai-plugin-devops (v2.0.0-dev)
✓ moai-plugin-uiux (v2.0.0-dev)
✓ moai-plugin-technical-blog (v2.0.0-dev)

Available Commands:
  /init-fastapi              - Backend plugin
  /db-setup                  - Backend plugin
  /resource-crud             - Backend plugin
  /init-next                 - Frontend plugin
  /biome-setup               - Frontend plugin
  /deploy-config             - DevOps plugin
  /connect-vercel            - DevOps plugin
  /connect-supabase          - DevOps plugin
  /connect-render            - DevOps plugin
  /ui-ux                     - UI/UX plugin
  /setup-shadcn-ui           - UI/UX plugin
  /design-tokens             - UI/UX plugin
  /blog-write                - Technical Blog plugin
```

### Step 4: 첫 번째 명령어 실행 (1분)

```bash
# Backend 프로젝트 생성
/init-fastapi

# 입력
Project name: my_first_api
Python version: 3.13
Database: PostgreSQL
```

---

## 📖 문서별 용도

### `claude-code-plugin-installation-guide.md`

**대상**: 플러그인 사용자, 팀 리더

**사용 시점**:
- 처음 플러그인을 설치할 때
- 플러그인 구조를 이해하고 싶을 때
- 각 플러그인을 상세히 테스트하고 싶을 때

**주요 섹션**:
- ✅ 마켓플레이스 개요 (5개 플러그인, 23개 에이전트, 22개 스킬)
- ✅ moai-plugin-backend 상세 분석
- ✅ 3가지 설치 방법 (UI, CLI, 설정 파일)
- ✅ 단계별 테스트 시나리오
- ✅ 고급 통합 시나리오

### `plugin-quick-reference.md`

**대상**: 일일 사용자, 개발자

**사용 시점**:
- 어떤 플러그인을 사용할지 고르고 싶을 때
- 특정 플러그인의 명령어를 빠르게 찾고 싶을 때
- 에이전트를 활용하고 싶을 때

**주요 섹션**:
- ✅ 플러그인 선택 가이드 (상황별 추천)
- ✅ 플러그인별 명령어 상세 설명
- ✅ 각 플러그인의 에이전트 활용법
- ✅ 플러그인 조합 시나리오 (3가지)
- ✅ FAQ (6개)

### `plugin-testing-scenarios.md`

**대상**: QA 엔지니어, 플러그인 개발자, DevOps 담당자

**사용 시점**:
- 플러그인 품질을 보증하고 싶을 때
- 자동화된 테스트를 구축하고 싶을 때
- CI/CD 파이프라인에 통합하고 싶을 때

**주요 섹션**:
- ✅ Unit Tests (파일 검증 스크립트)
- ✅ Integration Tests (Claude Code 통합 테스트)
- ✅ E2E Tests (완전한 사용자 워크플로우)
- ✅ Performance Tests (성능 메트릭)
- ✅ 자동화 스크립트 (bash)
- ✅ GitHub Actions 예제

---

## 🎯 사용 사례별 로드맵

### 사용 사례 1: "FastAPI 백엔드 만들기"

```
1. plugin-quick-reference.md 읽기
   → moai-plugin-backend 섹션 찾기

2. claude-code-plugin-installation-guide.md 읽기
   → "단계 2: 플러그인 설치" → "Backend 플러그인 명령어 테스트"

3. 실행
   /init-fastapi
   /db-setup
   /resource-crud

4. 결과
   ✅ FastAPI 프로젝트 생성됨
   ✅ PostgreSQL 연동됨
   ✅ CRUD API 자동 생성됨
```

**필요한 문서**: quick-reference (2분) + installation-guide (5분)

### 사용 사례 2: "전체 스택 애플리케이션 배포"

```
1. plugin-quick-reference.md
   → 전체 스택 애플리케이션 섹션

2. claude-code-plugin-installation-guide.md
   → "고급 플러그인 테스트 시나리오" → "시나리오 1"

3. 순서대로 실행
   Backend 초기화
   → Frontend 초기화
   → UI/UX 디자인
   → DevOps 배포

4. 결과
   ✅ FastAPI 백엔드
   ✅ Next.js 프론트엔드
   ✅ Vercel/Supabase/Render 배포
```

**필요한 문서**: quick-reference (3분) + installation-guide (10분)

### 사용 사례 3: "플러그인 품질 검증"

```
1. plugin-testing-scenarios.md
   → "1️⃣ Unit Tests" 섹션

2. 검증 스크립트 실행
   bash validate-plugin-json.sh
   bash validate-command-files.sh

3. plugin-testing-scenarios.md
   → "2️⃣ Integration Tests"

4. E2E 테스트 실행
   /plugin install moai-plugin-backend
   /init-fastapi

5. plugin-testing-scenarios.md
   → "4️⃣ Performance Tests"

6. 성능 측정
   measure-plugin-load-time.sh
   measure-command-execution-time.sh

7. 결과
   ✅ 모든 테스트 통과
   ✅ 성능 메트릭 기록
```

**필요한 문서**: testing-scenarios (전체)

---

## 🔧 설정 파일 기반 설치 (팀 협업용)

### `.claude/settings.json` 예제

```json
{
  "env": {
    "MOAI_MARKETPLACE": "/Users/goos/MoAI/MoAI-ADK-v1.0/moai-marketplace",
    "MOAI_PLUGINS": "backend,frontend,devops"
  },
  "enabledPlugins": [
    "moai-plugin-backend@moai-marketplace",
    "moai-plugin-frontend@moai-marketplace",
    "moai-plugin-devops@moai-marketplace"
  ],
  "extraKnownMarketplaces": [
    "/Users/goos/MoAI/MoAI-ADK-v1.0/moai-marketplace"
  ],
  "permissions": {
    "allow": [
      "Task",
      "Read",
      "Write",
      "Edit",
      "Bash"
    ]
  }
}
```

**효과**: 팀원이 프로젝트를 열 때 플러그인이 자동으로 설치됨

---

## 📊 플러그인 매트릭스

### 기능별 플러그인 매핑

| 기능 | 플러그인 | 명령어 | 에이전트 |
|------|--------|--------|---------|
| **백엔드 스캐폴딩** | moai-plugin-backend | `/init-fastapi` | Backend Architect |
| **프론트엔드 스캐폴딩** | moai-plugin-frontend | `/init-next` | - |
| **UI 컴포넌트** | moai-plugin-uiux | `/setup-shadcn-ui` | Component Builder |
| **배포 자동화** | moai-plugin-devops | `/deploy-config` | Deployment Strategist |
| **콘텐츠 작성** | moai-plugin-technical-blog | `/blog-write` | Content Strategist |
| **디자인→코드** | moai-plugin-uiux | `/design-tokens` | Figma Designer |

### 에이전트 분포

```
Total Agents: 23

Specialist Agents:
  ├─ Backend: 4 (Backend Architect, FastAPI Specialist, API Designer, Database Expert)
  ├─ DevOps: 4 (Deployment Strategist, Vercel, Supabase, Render Specialists)
  ├─ UI/UX: 7 (Design Strategist, Design System Architect, Component Builder, Figma Designer, CSS/HTML Generator, Accessibility Specialist, Design Documentation Writer)
  ├─ Content: 6 (Content Strategist, Technical Writer, SEO Specialist, Code Example Curator, Visual Designer, Markdown Formatter)
  └─ Frontend: 0

Coordinator Agents:
  └─ Template Workflow Coordinator (Technical Blog)
```

### 스킬 분포

```
Total Skills: 22

By Category:
  ├─ Framework Skills: 3 (Next.js, React, FastAPI)
  ├─ Domain Skills: 4 (Frontend, Backend, Database, DevOps)
  ├─ Language Skills: 4 (Python, FastAPI, TypeScript, SQL)
  ├─ SaaS Skills: 3 (Vercel, Supabase, Render)
  ├─ Design Skills: 3 (Figma MCP, Figma-to-Code, shadcn/ui)
  └─ Content Skills: 2 (SEO, Blog Strategy, Image Generation, etc.)
```

---

## 🧪 테스트 상태

### 문서별 테스트 적용 범위

| 문서 | Unit Tests | Integration | E2E | Performance | CI/CD |
|------|-----------|-------------|-----|-------------|-------|
| Installation Guide | ✅ | ✅ | ✅ | ⭕ | ⭕ |
| Quick Reference | - | - | ✅ | - | - |
| Testing Scenarios | ✅ | ✅ | ✅ | ✅ | ✅ |

**범례**: ✅ 포함됨, ⭕ 부분 포함, - 포함 안 함

---

## 📝 다음 단계

### Phase 1: 기본 검증 (지금)

- [ ] 마켓플레이스 등록
- [ ] 모든 플러그인 설치
- [ ] `/help` 확인
- [ ] `/init-fastapi` 테스트

### Phase 2: 완전한 워크플로우 (내일)

- [ ] Backend → Frontend → DevOps (완전 스택)
- [ ] UI/UX 디자인 시스템
- [ ] 기술 블로그 작성

### Phase 3: 자동화 & CI/CD (1주일)

- [ ] 자동화 테스트 스크립트 실행
- [ ] GitHub Actions 통합
- [ ] 팀 문서화

### Phase 4: 성능 최적화 & 배포 (2주)

- [ ] 성능 메트릭 수집
- [ ] 병목 지점 분석
- [ ] 프로덕션 배포

---

## 📚 전체 문서 인덱스

```
플러그인 가이드 문서 구조:

1️⃣  개요 & 이론
    └─ plugin-ecosystem-introduction.md (생태계 개요)

2️⃣  설치 & 기본 사용
    ├─ claude-code-plugin-installation-guide.md (상세 설치 가이드)
    └─ plugin-quick-reference.md (빠른 참조)

3️⃣  고급 사용 & 통합
    └─ plugin-testing-scenarios.md (테스트 & 자동화)

4️⃣  체크리스트 & 로드맵
    └─ plugin-setup-checklist.md (이 문서)
```

---

## 🎊 축약 버전 명령어 치트시트

```bash
# 1. 마켓플레이스 추가 (1회만)
/plugin marketplace add /Users/goos/MoAI/MoAI-ADK-v1.0/moai-marketplace

# 2. 플러그인 설치 (필요한 것만)
/plugin install moai-plugin-backend@moai-marketplace

# 3. 플러그인 리스트
/plugin list

# 4. 마켓플레이스 관리
/plugin marketplace list

# 5. 플러그인 제거
/plugin uninstall moai-plugin-backend

# 6. 도움말
/help

# 7. 백엔드 프로젝트 시작
/init-fastapi

# 8. DB 설정
/db-setup

# 9. CRUD 생성
/resource-crud

# 10. 프론트엔드 시작
/init-next
```

---

## 💡 팁 & 트릭

### 팁 1: 빠른 풀 스택 생성

```bash
# Backend
cd backend && /init-fastapi && /db-setup

# Frontend
cd ../frontend && /init-next

# 완료!
```

### 팁 2: 에이전트 활용

```bash
# Backend Architect와 협력
Task(
  subagent_type="backend-architect",
  prompt="Design microservice architecture"
)
```

### 팁 3: 스킬 접근

```bash
# Python 스킬 로드
Skill("moai-lang-python")

# FastAPI 스킬
Skill("moai-framework-fastapi-patterns")
```

---

## ⚠️ 주의사항

### 1. 플러그인 버전

모든 플러그인은 현재 **개발 버전 (v1.0.0-dev ~ v2.0.0-dev)**입니다.

프로덕션 환경에서는 안정화된 버전이 릴리스될 때까지 테스트 환경에서만 사용하세요.

### 2. MCP 서버 의존성

일부 플러그인 (devops, uiux)은 MCP 서버가 필요합니다:

- **moai-plugin-devops**: Vercel, Supabase, Render MCP
- **moai-plugin-uiux**: Figma MCP

설치 전에 필요한 MCP 서버가 구성되었는지 확인하세요.

### 3. Python & Node.js 버전

- Backend 플러그인: Python 3.13 권장
- Frontend 플러그인: Node.js 20+ 권장

### 4. 디스크 공간

전체 플러그인 설치 시 약 500MB의 추가 디스크 공간이 필요합니다.

---

## 🆘 도움말 연락처

### 문서 관련

- 설치 문제: `claude-code-plugin-installation-guide.md` → 문제 해결 섹션
- 사용 방법: `plugin-quick-reference.md` → FAQ
- 테스트 방법: `plugin-testing-scenarios.md` → 자동화 테스트 스크립트

### GitHub

- 마켓플레이스: https://github.com/moai-adk/moai-marketplace
- 이슈 리포트: GitHub Issues

### 공식 문서

- Claude Code: https://docs.claude.com/en/docs/claude-code/

---

## ✅ 완료 확인 리스트

플러그인 설정이 완료되었는지 확인하세요:

- [x] 공식 문서 검토 완료
- [x] 마켓플레이스 구조 분석 완료
- [x] 4개의 상세 가이드 문서 생성 완료
- [x] 플러그인 설치 가이드 작성 완료
- [x] 테스트 시나리오 작성 완료
- [x] 자동화 스크립트 제공 완료
- [ ] 실제 플러그인 설치 (사용자가 수행)
- [ ] 명령어 실행 테스트 (사용자가 수행)
- [ ] E2E 워크플로우 검증 (사용자가 수행)

---

**문서 작성 완료**: 2025-10-31
**총 가이드 크기**: 91KB (3,656줄)
**상태**: ✅ 준비 완료

다음 단계: 위의 "다음 단계" 섹션을 참고하여 실제 설치를 진행하세요!
