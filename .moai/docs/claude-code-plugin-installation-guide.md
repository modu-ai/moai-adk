# Claude Code 플러그인 설치 및 테스트 가이드

**작성일**: 2025-10-31
**대상**: MoAI-ADK 플러그인 개발자 및 사용자
**마켓플레이스**: `/Users/goos/MoAI/MoAI-ADK-v1.0/moai-marketplace`

---

## 📊 플러그인 마켓플레이스 개요

### 마켓플레이스 구조

```
moai-marketplace/
├── marketplace.json              # 마켓플레이스 메타데이터 및 플러그인 카탈로그
├── plugins/                      # 플러그인 구현체들
│   ├── moai-plugin-backend       # Backend scaffolding (FastAPI, SQLAlchemy)
│   ├── moai-plugin-frontend      # Frontend scaffolding (Next.js, React 19)
│   ├── moai-plugin-devops        # Multi-cloud deployment (Vercel, Supabase, Render)
│   ├── moai-plugin-uiux          # Design automation (Figma MCP, shadcn/ui)
│   └── moai-plugin-technical-blog# Technical writing excellence
├── docs/                         # 플러그인 개발 문서
└── README.md                     # 마켓플레이스 설명서
```

### 이용 가능한 플러그인 목록

| 플러그인 | 버전 | 상태 | 설명 | 에이전트 수 |
|---------|------|------|------|----------|
| **moai-plugin-backend** | 1.0.0-dev | 개발 중 | FastAPI + uv + SQLAlchemy 2.0 | 4 |
| **moai-plugin-frontend** | 1.0.0-dev | 개발 중 | Next.js 16 + React 19.2 | 0 |
| **moai-plugin-devops** | 2.0.0-dev | 개발 중 | Vercel, Supabase, Render MCP | 4 |
| **moai-plugin-uiux** | 2.0.0-dev | 개발 중 | Figma MCP + Design-to-Code | 7 |
| **moai-plugin-technical-blog** | 2.0.0-dev | 개발 중 | 기술 블로그 작성 자동화 | 7 |

**마켓플레이스 통계**:
- 총 플러그인: 5개
- 전문 에이전트: 23개
- 스킬: 22개
- 지원 언어: 영어 (다국어 지원 예정)

---

## 🔧 플러그인 구조 분석: moai-plugin-backend

### 1. 플러그인 설정 파일 (plugin.json)

**경로**: `plugins/moai-plugin-backend/.claude-plugin/plugin.json`

```json
{
  "id": "moai-plugin-backend",
  "name": "Backend Plugin",
  "version": "1.0.0-dev",
  "status": "development",
  "description": "FastAPI 0.120.2 + uv scaffolding",
  "author": "GOOS🪿",
  "category": "backend",
  "minClaudeCodeVersion": "1.0.0",
  "commands": [
    {
      "name": "init-fastapi",
      "description": "Initialize FastAPI project with uv"
    },
    {
      "name": "db-setup",
      "description": "Setup database with Alembic"
    },
    {
      "name": "resource-crud",
      "description": "Generate CRUD endpoints from SPEC"
    }
  ],
  "agents": [
    {
      "name": "backend-agent",
      "type": "specialist",
      "description": "Backend scaffolding specialist"
    }
  ],
  "skills": [
    "moai-lang-fastapi-patterns",
    "moai-lang-python",
    "moai-domain-backend",
    "moai-domain-database"
  ],
  "permissions": {
    "allowedTools": ["Read", "Write", "Edit", "Bash"],
    "deniedTools": []
  }
}
```

**핵심 요소**:
- **ID**: 플러그인 고유 식별자
- **Commands**: 사용자가 실행할 수 있는 슬래시 명령어 (3개)
- **Agents**: 전문화된 에이전트 (4개 - backend-architect, fastapi-specialist, api-designer, database-expert)
- **Skills**: 재사용 가능한 지식 모듈 (4개)
- **Permissions**: 도구 접근 제어

### 2. 플러그인 에이전트들

**경로**: `plugins/moai-plugin-backend/agents/`

| 에이전트 | 역할 | 설명 |
|--------|------|------|
| **backend-architect.md** | Server Architecture | FastAPI 앱 구조 설계, 의존성 주입, 마이크로서비스 아키텍처 |
| **fastapi-specialist.md** | FastAPI Expert | 비동기 패턴, 라우터 설계, 요청/응답 처리 |
| **api-designer.md** | API Design | OpenAPI 스펙, RESTful 설계, 엔드포인트 최적화 |
| **database-expert.md** | Database | SQLAlchemy 모델, Alembic 마이그레이션, 쿼리 최적화 |

### 3. 플러그인 스킬들

**경로**: `plugins/moai-plugin-backend/skills/`

- `moai-lang-fastapi-patterns.md` - FastAPI 비동기 패턴
- `moai-lang-python.md` - Python 3.13+ 베스트 프랙티스
- `moai-domain-backend.md` - 백엔드 아키텍처
- `moai-domain-database.md` - 데이터베이스 설계

---

## 📋 플러그인 설치 단계별 가이드

### 단계 1: 마켓플레이스 등록 (로컬 환경)

로컬 마켓플레이스를 Claude Code에 등록합니다:

```bash
# Claude Code에서 실행
/plugin marketplace add /Users/goos/MoAI/MoAI-ADK-v1.0/moai-marketplace

# 또는 상대 경로 사용 (프로젝트 루트에서)
/plugin marketplace add ./moai-marketplace
```

**결과**:
- 마켓플레이스가 등록되면 "moai-marketplace"가 활성화됨
- 모든 5개의 플러그인이 설치 가능한 상태

### 단계 2: 플러그인 설치

#### 옵션 A: 대화형 메뉴 (권장)

```
/plugin
```

메뉴가 표시됩니다:

```
❯ 1. Browse and install plugins
  2. Manage and uninstall plugins
  3. Add marketplace
  4. Manage marketplaces
```

**1번 선택** → "Browse and install plugins"

**마켓플레이스 선택**:
```
Select marketplace:
❯ moai-marketplace
  (other marketplaces if available)
```

**플러그인 선택**:
```
Select plugin:
  moai-plugin-backend
  moai-plugin-frontend
  moai-plugin-devops
  moai-plugin-uiux
❯ moai-plugin-technical-blog
```

테스트용으로 `moai-plugin-backend` 선택 권장

#### 옵션 B: 직접 명령어 (개발자용)

```bash
# Backend 플러그인 설치
/plugin install moai-plugin-backend@moai-marketplace

# 모든 플러그인 설치
/plugin install moai-plugin-frontend@moai-marketplace
/plugin install moai-plugin-devops@moai-marketplace
/plugin install moai-plugin-uiux@moai-marketplace
/plugin install moai-plugin-technical-blog@moai-marketplace
```

#### 옵션 C: 설정 파일 기반 (팀 협업용)

`.claude/settings.json`에 추가:

```json
{
  "enabledPlugins": [
    "moai-plugin-backend@moai-marketplace",
    "moai-plugin-frontend@moai-marketplace",
    "moai-plugin-devops@moai-marketplace",
    "moai-plugin-uiux@moai-marketplace"
  ],
  "extraKnownMarketplaces": [
    "/Users/goos/MoAI/MoAI-ADK-v1.0/moai-marketplace"
  ]
}
```

### 단계 3: 설치 검증

설치 후 다음 명령어로 확인:

```bash
# 설치된 플러그인 및 명령어 확인
/help

# 또는 플러그인 관리 메뉴
/plugin
```

**예상 출력** (moai-plugin-backend 설치 시):

```
Installed Plugins:
✓ moai-plugin-backend (v1.0.0-dev)

Available Commands:
  /init-fastapi       - Initialize FastAPI project with uv
  /db-setup           - Setup database with Alembic
  /resource-crud      - Generate CRUD endpoints from SPEC

Available Agents:
  - backend-architect (Specialist, Sonnet)
  - fastapi-specialist (Specialist, Haiku)
  - api-designer (Specialist, Haiku)
  - database-expert (Specialist, Haiku)
```

---

## 🧪 플러그인 기능 테스트

### 테스트 1: Backend 플러그인 명령어 테스트

#### 1.1 FastAPI 프로젝트 초기화

```bash
# 새로운 테스트 디렉토리 생성
mkdir -p /tmp/test-backend-plugin
cd /tmp/test-backend-plugin

# Claude Code에서 실행
/init-fastapi
```

**사용자 입력 예상**:
- 프로젝트 이름: `my_api`
- Python 버전: `3.13`
- 데이터베이스: `PostgreSQL`

**생성되는 파일**:
```
my_api/
├── pyproject.toml
├── app/
│   ├── __init__.py
│   ├── main.py
│   ├── api/
│   │   ├── __init__.py
│   │   └── v1/
│   │       ├── endpoints/
│   │       └── dependencies.py
│   ├── models/
│   ├── schemas/
│   └── core/
├── migrations/
├── tests/
└── README.md
```

#### 1.2 데이터베이스 설정

```bash
/db-setup
```

**기능**:
- PostgreSQL 또는 MySQL 연결 설정
- Alembic 마이그레이션 초기화
- `.env` 파일 생성 (데이터베이스 URL)
- 초기 마이그레이션 생성

#### 1.3 CRUD 엔드포인트 생성

```bash
/resource-crud
```

**입력 예시**:
```
Resource name: User
Fields:
- name (string, required)
- email (string, required, unique)
- age (integer, optional)
```

**생성 내용**:
- SQLAlchemy 모델 (`models/user.py`)
- Pydantic 스키마 (`schemas/user.py`)
- CRUD 라우터 (`api/v1/endpoints/user.py`)
- 자동 엔드포인트:
  - `GET /api/v1/users` - 전체 조회
  - `POST /api/v1/users` - 생성
  - `GET /api/v1/users/{id}` - 단일 조회
  - `PUT /api/v1/users/{id}` - 수정
  - `DELETE /api/v1/users/{id}` - 삭제

### 테스트 2: 에이전트 상호작용 테스트

각 플러그인의 에이전트를 Task 도구로 호출:

#### 2.1 Backend Architect 에이전트

```
Task(
  subagent_type="backend-architect",
  prompt="Design a scalable FastAPI microservice architecture for an e-commerce platform with users, products, and orders"
)
```

**예상 출력**:
- 마이크로서비스 분리 제안
- 의존성 주입 패턴 설명
- 데이터베이스 스키마 제안
- 비동기 태스크 큐 구조

#### 2.2 FastAPI Specialist 에이전트

```
Task(
  subagent_type="fastapi-specialist",
  prompt="Optimize the user authentication flow for JWT token management"
)
```

**예상 기능**:
- 토큰 생성/검증 로직
- 미들웨어 구현
- 에러 핸들링

#### 2.3 Database Expert 에이전트

```
Task(
  subagent_type="database-expert",
  prompt="Design database schema for user profile with relationships"
)
```

### 테스트 3: 스킬 로딩 테스트

플러그인이 로드한 스킬들을 사용:

```bash
# Python 스킬 활용
Skill("moai-lang-python")

# FastAPI 패턴 스킬
Skill("moai-framework-fastapi-patterns")

# 데이터베이스 설계 스킬
Skill("moai-domain-database")
```

---

## 🚀 고급 플러그인 테스트 시나리오

### 시나리오 1: 전체 백엔드 프로젝트 구축

1. `moai-plugin-backend` 설치
2. `/init-fastapi` 실행 → 프로젝트 스캐폴딩
3. `/db-setup` 실행 → PostgreSQL 설정
4. `/resource-crud` 실행 (여러 리소스)
   - User 리소스
   - Product 리소스
   - Order 리소스
5. 에이전트 호출 → 인증/인가 로직 추가
6. 테스트 실행 → 모든 엔드포인트 검증

### 시나리오 2: UI/UX 디자인 플러그인 테스트

```bash
/plugin install moai-plugin-uiux@moai-marketplace

# 설치 후
/setup-shadcn-ui

# Figma 토큰 설정
/design-tokens

# 컴포넌트 자동 생성
# (Figma MCP 연동 필요)
```

### 시나리오 3: DevOps 배포 파이프라인 테스트

```bash
/plugin install moai-plugin-devops@moai-marketplace

# Vercel 프론트엔드 배포
/connect-vercel

# Supabase 데이터베이스 연동
/connect-supabase

# Render 백엔드 배포
/connect-render
```

---

## 📊 플러그인 성능 메트릭

설치 후 성능을 모니터링하려면:

```bash
# Claude Code 상태 확인
/help

# 플러그인 상태
/plugin

# 설치된 플러그인 목록
/plugin list

# 플러그인 제거
/plugin uninstall moai-plugin-backend
```

### 기대되는 성능

| 메트릭 | 값 |
|------|-----|
| 플러그인 로드 시간 | < 500ms |
| 명령어 실행 시간 | 1-3초 (초기화) |
| 에이전트 응답 시간 | 2-5초 |
| 스킬 로드 시간 | < 200ms |

---

## 🔍 문제 해결

### 문제 1: 마켓플레이스 인식 안 됨

```bash
# 전체 경로 사용
/plugin marketplace add /Users/goos/MoAI/MoAI-ADK-v1.0/moai-marketplace

# 또는 확인
/plugin marketplace list
```

### 문제 2: 플러그인 설치 실패

**확인 사항**:
- `plugin.json` 파일이 존재하는가?
- JSON 형식이 올바른가?
- 필수 필드가 있는가? (id, name, version, description)

```bash
# plugin.json 검증
cat /Users/goos/MoAI/MoAI-ADK-v1.0/moai-marketplace/plugins/moai-plugin-backend/.claude-plugin/plugin.json | jq .
```

### 문제 3: 명령어 인식 안 됨

```bash
# 플러그인 다시 로드
/plugin uninstall moai-plugin-backend
/plugin install moai-plugin-backend@moai-marketplace

# Claude Code 재시작
/clear
```

---

## 📚 추가 리소스

### 플러그인 개발 문서

- **마켓플레이스 가이드**: `marketplace.json` 스키마
- **에이전트 템플릿**: `docs/agent-template-guide.md`
- **커맨드 템플릿**: `docs/command-template-guide.md`
- **플러그인 스키마**: `docs/plugin-json-schema.md`

### 공식 Claude Code 문서

- https://docs.claude.com/en/docs/claude-code/plugins.md
- https://docs.claude.com/en/docs/claude-code/plugin-marketplaces.md

### MoAI-ADK 마켓플레이스

- **GitHub**: https://github.com/moai-adk/moai-marketplace
- **로컬 경로**: `/Users/goos/MoAI/MoAI-ADK-v1.0/moai-marketplace`

---

## ✅ 테스트 체크리스트

설치 및 테스트 완료 확인:

- [ ] 마켓플레이스 등록 완료
- [ ] `moai-plugin-backend` 설치 완료
- [ ] `/help` 명령어에서 새 플러그인 명령어 확인
- [ ] `/init-fastapi` 명령어 실행 테스트
- [ ] `/db-setup` 명령어 실행 테스트
- [ ] 에이전트 호출 테스트
- [ ] 스킬 로드 테스트
- [ ] 설정 파일 기반 설치 테스트
- [ ] 다른 플러그인 설치 테스트
- [ ] 플러그인 제거/재설치 테스트

---

**마지막 업데이트**: 2025-10-31
**문서 버전**: 1.0.0
**상태**: 완성
