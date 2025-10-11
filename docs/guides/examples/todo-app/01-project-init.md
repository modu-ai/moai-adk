# Part 1: 프로젝트 초기화

> **소요시간**: 약 30분
> **학습 목표**: `/alfred:0-project` 커맨드를 사용하여 프로젝트 문서를 작성하고, MoAI-ADK 프로젝트 구조를 이해합니다.

---

## 🎯 이번 Part에서 배울 것

- ✅ MoAI-ADK 프로젝트 생성 및 초기화
- ✅ `/alfred:0-project` 커맨드 사용법
- ✅ Product Discovery 인터뷰 프로세스
- ✅ 프로젝트 문서 구조 (product, structure, tech)
- ✅ 기술 스택 선택 및 조정

---

## 📋 사전 준비

### 필수 도구 확인

```bash
# MoAI-ADK 설치 확인
moai --version

# Node.js 18+ (Frontend 용)
node --version

# Python 3.11+ (Backend 용)
python --version
```

### 프로젝트 디렉토리 생성

```bash
# 프로젝트 디렉토리 생성 및 이동
mkdir my-moai-project
cd my-moai-project

# MoAI-ADK 초기화
moai init .
```

**예상 출력**:

```
✅ MoAI-ADK 프로젝트가 초기화되었습니다!

생성된 파일:
  .moai/
  ├── config.json       # 프로젝트 설정
  ├── project/          # 프로젝트 문서 (템플릿)
  │   ├── product.md
  │   ├── structure.md
  │   └── tech.md
  ├── specs/            # SPEC 문서 저장소
  └── memory/           # 개발 가이드
      ├── development-guide.md
      └── spec-metadata.md

다음 단계:
1. /alfred:0-project - 프로젝트 문서 작성
2. /alfred:1-spec - 첫 번째 SPEC 작성
```

---

## 🚀 Step 1: `/alfred:0-project` 실행

Claude Code를 열고 다음 커맨드를 실행합니다:

```
/alfred:0-project
```

### Alfred의 응답

Alfred가 프로젝트 환경을 분석하고 초기화 계획을 제시합니다.

**분석 보고서 예시**:

```
📊 프로젝트 초기화 계획: my-moai-project

환경 분석 결과:
- 프로젝트 유형: 신규 프로젝트 (템플릿 문서 존재)
- 감지된 언어: 감지되지 않음 (빈 프로젝트)
- 현재 문서 상태: 100% (템플릿)
- 구조 복잡도: 단순

🎯 인터뷰 전략:
- 질문 카테고리: Product Discovery
- 예상 질문 수: 15개 (필수 10개 + 선택 5개)
- 예상 소요시간: 10-15분
- 우선순위 영역:
  a. 핵심 미션 및 비즈니스 가치
  b. 주요 사용자층 및 시나리오
  c. 해결할 핵심 문제
  d. 기술 스택 선택
  e. 품질 정책 및 배포 전략

⚠️ 주의사항:
- 기존 문서: 템플릿 상태 → 실제 정보로 갱신 필요
- 언어 설정: 자동 감지 불가 → 인터뷰 중 확인 필요
- 설정 충돌: config.json 기본 설정 → 언어별 최적화 필요

✅ 예상 산출물:
- product.md: 비즈니스 요구사항 문서
- structure.md: 시스템 아키텍처 문서
- tech.md: 기술 스택 및 정책 문서
- config.json: 프로젝트 설정 파일

📝 인터뷰 흐름 (3단계):
1단계: Product Discovery (product.md 작성)
2단계: Structure Blueprint (structure.md 작성)
3단계: Tech Stack Mapping (tech.md 작성)

---
승인 요청: 위 계획으로 인터뷰를 진행하시겠습니까?

- "진행" 또는 "시작": 인터뷰 시작
- "수정 [내용]": 계획 수정
- "중단": 초기화 중단
```

### 승인하기

계획을 확인했다면 다음과 같이 입력하세요:

```
진행
```

---

## 💬 Step 2: Product Discovery 인터뷰

Alfred가 프로젝트 목적과 요구사항을 파악하기 위한 질문을 시작합니다.

### Phase 1: 기본 정보

**Alfred의 질문**:

```
Phase 1: Product Discovery

다음 질문에 답변해주세요:

1. 프로젝트의 핵심 목적은 무엇인가요?
   - 해결하려는 문제
   - 주요 사용자

2. 프로젝트 유형은 무엇인가요?
   - 웹/모바일/API/CLI/라이브러리 등

3. 현재 프로젝트 단계는 어디인가요?
   - POC/초기개발/활성개발/유지보수 등
```

### 답변 예시 1: 상세 답변

프로젝트에 대해 명확히 알고 있다면 다음과 같이 답변하세요:

```
1. 학습용 ToDo-App 프로젝트 예제
   - MoAI-ADK SPEC-First TDD 방법론 학습
   - 풀스택 개발 경험 쌓기

2. 웹 애플리케이션 (PC/모바일)
   - Frontend: TypeScript 기반
   - Backend: Python FastAPI 기반

3. 초기 개발 단계
```

### 답변 예시 2: 간단히 답변 (권장)

빠르게 진행하고 싶다면:

```
1. 학습용 ToDo-App 프로젝트 예제
2. 웹앱 (pc/모바일) Next.js fastapi, sqlmodel, pydantic2
```

Alfred가 나머지 정보를 자동으로 추론합니다.

### 기술 스택 조정 (필요시)

Alfred가 제안한 기술 스택을 수정하고 싶다면:

**프레임워크 변경**:

```
사용자: 스택: Next.js -> 말고 간단하게 테스트 할 수 있는 정도의 프레임워크면 좋겠다.
       그리고 tailwind css + 추가

Alfred: 알겠습니다! 더 간단한 구성으로 수정하겠습니다.

최종 종합 답변 (수정):

기술 스택:
- 프론트엔드: Vite + React, TypeScript, Tailwind CSS
- 백엔드: FastAPI, SQLModel, Pydantic v2, Python 3.11+
- 데이터베이스: SQLite
```

**ORM 변경**:

```
사용자: sqlmodel -> sqlalchemy 2.0, alembic 추가 해서 사용하도록 하자.

Alfred: 알겠습니다! 기술 스택을 수정하겠습니다.

변경 이유:
- SQLAlchemy 2.0의 성숙한 ORM 및 풍부한 커뮤니티 지원
- Alembic을 통한 체계적인 데이터베이스 스키마 버전 관리
- 학습 목적으로 마이그레이션 관리 경험 추가
```

**배포 전략**:

```
사용자: 배포는 로컬 docker로 배포

Alfred: 배포 정보를 수정하겠습니다.

배포:
- 로컬 Docker Compose
- Docker Compose (frontend + backend + db)
```

### 자동 답변 요청 (시간 절약)

복잡한 질문에 시간을 쓰고 싶지 않다면:

```
사용자: 네 나머지도 알아서 답변을 해주세요

Alfred: 알겠습니다. 학습용 ToDo-App 프로젝트 특성을 고려하여
        합리적인 답변을 작성하겠습니다.

---
종합 답변 정보:

기본 정보:
- 핵심 목적: MoAI-ADK SPEC-First TDD 방법론 학습 및 풀스택 개발 연습
- 사용자: 개인 학습용 (본인)
- 프로젝트 유형: 웹 애플리케이션 (풀스택: Vite+React + FastAPI)
- 현재 단계: 초기 개발

기술 스택:
- 프론트엔드: Vite + React, TypeScript, Tailwind CSS
- 백엔드: FastAPI, SQLAlchemy 2.0 + Alembic, Pydantic v2, Python 3.11+
- 데이터베이스: SQLite

주요 기능:
- 할일(Todo) CRUD 기능
- 상태 관리 (완료/미완료/진행중)
- 우선순위 설정 (높음/중간/낮음)
- 필터링 및 검색

아키텍처:
- 풀스택 모노레포 구조
- REST API 기반 통신
- 타입 안전성 (TypeScript + Pydantic)

품질 목표:
- 테스트 커버리지: 85% 이상
- TRUST 5원칙 준수
- @TAG 기반 추적성

배포:
- 로컬 Docker 배포
- Docker Compose (프론트엔드 + 백엔드 + DB)
```

### 최종 확인

Alfred가 종합 답변을 제시하면 확인 후 진행하세요:

```
사용자: 네 이렇게 진행을 해주세요.
```

---

## 📄 Step 3: 문서 자동 생성

Alfred가 인터뷰 내용을 바탕으로 3개 문서를 자동 생성합니다.

### 생성 결과

```
✅ 프로젝트 초기화 완료!

📁 생성/갱신된 문서:
┌──────────────────────────────┬────────┬────────┬──────────┐
│ 문서                         │ 버전   │ 크기   │ 상태     │
├──────────────────────────────┼────────┼────────┼──────────┤
│ .moai/project/product.md     │ v0.2.0 │ 6.8 KB │ ✅ 갱신   │
│ .moai/project/structure.md   │ v0.2.0 │ 8.2 KB │ ✅ 갱신   │
│ .moai/project/tech.md        │ v0.2.0 │ 9.5 KB │ ✅ 갱신   │
│ .moai/config.json            │ v0.2.0 │ -      │ ✅ 최적화 │
└──────────────────────────────┴────────┴────────┴──────────┘

🔍 품질 검증 결과:
✅ @TAG 체계 검증 (15개 TAG 확인)
✅ YAML Front Matter 형식
✅ HISTORY 섹션
✅ config.json 유효한 JSON 형식

🚀 감지된 환경 정보:
프로젝트 유형: 풀스택 모노레포 (학습용 ToDo 애플리케이션)

Frontend:
- 언어: TypeScript 5.x
- 프레임워크: Vite + React 18
- 스타일링: Tailwind CSS
- 테스트: Vitest + Testing Library
- 도구: pnpm, ESLint, Prettier

Backend:
- 언어: Python 3.11+
- 프레임워크: FastAPI
- ORM: SQLAlchemy 2.0
- 마이그레이션: Alembic
- 검증: Pydantic v2
- 테스트: pytest + pytest-cov
- 도구: pip, Ruff, mypy

데이터베이스: SQLite (로컬)

배포: Docker Compose (로컬)
```

---

## 📚 Step 4: 생성된 문서 확인

각 문서의 내용을 확인해봅시다.

### product.md - 비즈니스 요구사항

```bash
cat .moai/project/product.md
```

**주요 섹션**:

```markdown
---
id: PRODUCT-001
version: 0.2.0
status: active
created: 2025-10-01
updated: 2025-10-11
author: @Goos
priority: high
---

# my-moai-project Product Definition

## HISTORY

### v0.2.0 (2025-10-11)
- **CHANGED**: 학습용 ToDo-App 프로젝트로 실제 정보 갱신
- **AUTHOR**: @Goos
- **REVIEW**: Alfred (project-manager)

## @DOC:MISSION-001 핵심 미션

**MoAI-ADK SPEC-First TDD 방법론을 학습하고 실습하기 위한
ToDo 애플리케이션 예제 프로젝트**

이 프로젝트는 학습 목적으로 설계되었으며, 다음을 목표로 합니다:
- SPEC 우선 개발 프로세스 체득
- TDD (Red-Green-Refactor) 사이클 실습
- @TAG 시스템 기반 추적성 이해
- 풀스택 개발 경험 (TypeScript + Python)

### 핵심 가치 제안
- **실전 학습 환경**: 실무와 유사한 풀스택 구조에서 TDD 방법론 학습
- **타입 안전성**: TypeScript + Pydantic을 통한 엔드투엔드 타입 안전성
- **추적 가능성**: @TAG 시스템으로 SPEC → 테스트 → 코드 → 문서 연결
- **재현 가능한 환경**: Docker 기반 로컬 배포로 일관된 개발 환경

## @SPEC:USER-001 주요 사용자층

### 1차 사용자
- **대상**: 본인 (개인 학습자)
- **핵심 니즈**:
  - MoAI-ADK 개발 방법론 실습
  - SPEC-First TDD 워크플로우 이해
  - 풀스택 개발 능력 향상
  - @TAG 추적성 시스템 체득

## TODO:SPEC-BACKLOG-001 다음 단계 SPEC 후보

1. **TODO-001**: Todo 항목 CRUD 기능 (생성, 조회, 수정, 삭제)
2. **TODO-002**: Todo 상태 관리 (완료/미완료/진행중)
3. **TODO-003**: Todo 우선순위 설정 (높음/중간/낮음)
4. **TODO-004**: Todo 필터링 및 검색 기능
5. **TODO-005**: 프론트엔드 UI 컴포넌트 (Tailwind CSS)
6. **TODO-006**: Docker Compose 배포 환경 구성
```

### structure.md - 시스템 아키텍처

```bash
cat .moai/project/structure.md
```

**주요 섹션**:

```markdown
# my-moai-project Structure Design

## @DOC:ARCHITECTURE-001 전체 아키텍처

### 아키텍처 전략

**풀스택 모노레포 구조 (프론트엔드 + 백엔드 분리)**

```

my-moai-project/
├── frontend/              # React + TypeScript 프론트엔드
│   ├── src/
│   │   ├── components/   # UI 컴포넌트 (Tailwind CSS)
│   │   ├── api/          # 백엔드 API 호출 레이어
│   │   ├── types/        # TypeScript 타입 정의
│   │   └── tests/        # 프론트엔드 테스트
│   └── vite.config.ts
│
├── backend/               # FastAPI + Python 백엔드
│   ├── app/
│   │   ├── api/          # REST API 엔드포인트
│   │   ├── models/       # SQLAlchemy 모델
│   │   ├── schemas/      # Pydantic 스키마
│   │   └── services/     # 비즈니스 로직
│   ├── tests/            # 백엔드 테스트
│   └── requirements.txt
│
├── docker-compose.yml     # 로컬 배포 설정
└── .moai/                # MoAI-ADK 프로젝트 문서

```

**선택 이유**:
- **모노레포**: 학습 편의성 (하나의 프로젝트에서 풀스택 전체 관리)
- **명확한 분리**: 프론트-백엔드 책임 분리, 독립적 테스트 가능
- **타입 안전성**: TypeScript (클라이언트) ↔ Pydantic (서버) 타입 계약
- **단순성**: 마이크로서비스보다 학습 및 배포가 단순

## @DOC:MODULES-001 모듈별 책임 구분

### 1. Frontend 모듈 (Vite + React + TypeScript)

| 컴포넌트 | 역할 | 주요 기능 |
|----------|------|-----------|
| `TodoList` | Todo 목록 표시 | 전체/완료/미완료 필터링, 상태 토글 |
| `TodoForm` | Todo 생성 폼 | 제목/우선순위 입력, 유효성 검증 |
| `TodoItem` | 개별 Todo 표시 | 수정/삭제/상태변경 액션 |
| `api/client` | API 통신 레이어 | Fetch/Axios 기반 REST API 호출 |

### 2. Backend 모듈 (FastAPI + SQLAlchemy + Pydantic)

| 컴포넌트 | 역할 | 주요 기능 |
|----------|------|-----------|
| `api/todos` | Todo REST API | GET/POST/PATCH/DELETE 엔드포인트 |
| `models/todo` | Todo 데이터 모델 | SQLAlchemy 기반 ORM 모델 |
| `schemas/todo` | Todo 스키마 | Pydantic 기반 요청/응답 검증 |
| `services/todo` | Todo 비즈니스 로직 | CRUD 작업, 필터링, 정렬 |
```

### tech.md - 기술 스택 및 정책

```bash
cat .moai/project/tech.md
```

**주요 섹션**:

```markdown
# my-moai-project Technology Stack

## @DOC:STACK-001 언어 & 런타임

### Frontend 언어 선택
- **언어**: TypeScript
- **버전**: TypeScript 5.x, ES2022+
- **런타임**: Node.js 18+ (LTS)
- **패키지 매니저**: pnpm (빠른 설치, 디스크 효율성)

### Backend 언어 선택
- **언어**: Python
- **버전**: Python 3.11+
- **선택 이유**:
  - FastAPI/Pydantic의 강력한 타입 시스템
  - SQLAlchemy 2.0의 성숙한 ORM 및 풍부한 생태계
  - Alembic을 통한 체계적인 DB 마이그레이션 관리
  - pytest 기반 우수한 테스트 환경
- **패키지 매니저**: pip + requirements.txt

## @DOC:FRAMEWORK-001 핵심 프레임워크 & 라이브러리

### Frontend 주요 의존성
```json
{
  "dependencies": {
    "react": "^18.3.0",
    "react-dom": "^18.3.0"
  },
  "devDependencies": {
    "vite": "^5.0.0",
    "typescript": "^5.3.0",
    "tailwindcss": "^3.4.0",
    "@vitejs/plugin-react": "^4.2.0",
    "vitest": "^1.0.0",
    "@testing-library/react": "^14.0.0"
  }
}
```

### Backend 주요 의존성

```txt
# requirements.txt
fastapi==0.109.0
uvicorn[standard]==0.27.0
sqlalchemy==2.0.25
alembic==1.13.1
pydantic==2.5.0
pytest==7.4.0
pytest-cov==4.1.0
httpx==0.26.0  # 테스트용 HTTP 클라이언트
```

## @DOC:QUALITY-001 품질 게이트 & 정책

### 테스트 커버리지

- **목표**: 85% 이상 (Frontend & Backend 각각)
- **측정 도구**:
  - Frontend: Vitest + @vitest/coverage-v8
  - Backend: pytest-cov
- **실패 시 대응**: PR 차단, 커버리지 미달 영역 우선 수정

### 정적 분석

| 도구 | 역할 | 설정 파일 | 실패 시 조치 |
|------|------|-----------|--------------|
| **TypeScript** | 타입 검증 | `tsconfig.json` | 타입 오류 수정 필수 |
| **ESLint** | 코드 품질 (TS) | `.eslintrc.json` | 린트 경고 해결 권장 |
| **Prettier** | 코드 포맷 (TS) | `.prettierrc` | 자동 포맷팅 적용 |
| **Ruff** | 린터+포매터 (Python) | `ruff.toml` | 자동 수정 가능 |
| **mypy** | 타입 검증 (Python) | `mypy.ini` | 타입 힌트 추가 필수 |

```

### config.json - 프로젝트 설정

```bash
cat .moai/config.json
```

**내용**:

```json
{
  "version": "0.2.0",
  "mode": "personal",
  "projectName": "my-moai-project",
  "projectType": "fullstack",
  "projectDescription": "MoAI-ADK SPEC-First TDD 학습용 ToDo 애플리케이션",
  "locale": "ko",
  "frontend": {
    "language": "typescript",
    "framework": "vite-react",
    "runtime": "node",
    "runtimeVersion": "18+",
    "packageManager": "pnpm",
    "testFramework": "vitest",
    "linter": "eslint",
    "formatter": "prettier",
    "typeChecker": "typescript",
    "coverageTarget": 85,
    "buildTool": "vite",
    "styling": "tailwindcss"
  },
  "backend": {
    "language": "python",
    "framework": "fastapi",
    "runtime": "python",
    "runtimeVersion": "3.11+",
    "packageManager": "pip",
    "testFramework": "pytest",
    "linter": "ruff",
    "formatter": "ruff",
    "typeChecker": "mypy",
    "coverageTarget": 85,
    "orm": "sqlalchemy2",
    "migration": "alembic",
    "validation": "pydantic2"
  },
  "database": {
    "type": "sqlite",
    "location": "local"
  },
  "deployment": {
    "strategy": "docker-compose",
    "ports": {
      "frontend": 3000,
      "backend": 8000
    }
  },
  "backup": {
    "enabled": true,
    "retentionDays": 30
  }
}
```

---

## ✅ 완료 확인

다음 항목들을 확인하세요:

### 파일 생성 확인

```bash
# 프로젝트 문서 확인
ls -lh .moai/project/

# 예상 출력:
# product.md     (6.8 KB)
# structure.md   (8.2 KB)
# tech.md        (9.5 KB)

# 설정 파일 확인
cat .moai/config.json
```

### 품질 체크리스트

- [ ] `.moai/project/product.md` 생성 완료
- [ ] `.moai/project/structure.md` 생성 완료
- [ ] `.moai/project/tech.md` 생성 완료
- [ ] `.moai/config.json` 업데이트 완료
- [ ] 각 문서에 YAML Front Matter 포함
- [ ] 각 문서에 HISTORY 섹션 포함
- [ ] @TAG 체계 확인 (예: @DOC:MISSION-001)

---

## 🎓 학습 정리

### 핵심 개념

1. **Product Discovery**: 프로젝트의 비즈니스 요구사항을 명확히 정의
2. **Structure Blueprint**: 시스템 아키텍처 및 모듈 설계
3. **Tech Stack Mapping**: 기술 스택 및 품질 정책 수립

### @TAG 체계

product.md, structure.md, tech.md에서 발견한 TAG들:

- `@DOC:MISSION-001` - 핵심 미션
- `@SPEC:USER-001` - 주요 사용자
- `@DOC:ARCHITECTURE-001` - 아키텍처
- `@DOC:MODULES-001` - 모듈 구분
- `@DOC:STACK-001` - 언어 및 런타임
- `@DOC:QUALITY-001` - 품질 게이트
- `TODO:SPEC-BACKLOG-001` - SPEC 후보

이 TAG들은 나중에 SPEC 작성 시 참조됩니다.

---

## 🔍 트러블슈팅

### 문제 1: 인터뷰 질문이 너무 많아요

**해결책**:

```
사용자: 네 나머지도 알아서 답변을 해주세요
```

### 문제 2: 기술 스택을 잘못 선택했어요

**해결책**:

- Phase 1에서는 언제든 수정 가능합니다.
- 예: "sqlmodel → sqlalchemy 2.0으로 변경해주세요"

### 문제 3: config.json이 업데이트되지 않았어요

**해결책**:

```bash
# config.json 검증
cat .moai/config.json | python -m json.tool

# Alfred에게 재생성 요청
사용자: config.json을 다시 생성해주세요
```

---

## 🚀 다음

프로젝트 문서 작성이 완료되었습니다! 이제 첫 번째 SPEC을 작성할 준비가 되었습니다.

**다음**: [Part 2: SPEC 작성하기](./02-spec-writing.md)

**이전**: [튜토리얼 개요](./index.md)

---

## 📚 참고 자료

- [MoAI-ADK 개발 가이드](https://github.com/modu-ai/moai-adk/blob/main/.moai/memory/development-guide.md)
- [워크플로우 가이드: 0-project](../../workflow/0-project.md)
- [CLAUDE.md](https://github.com/modu-ai/moai-adk/blob/main/CLAUDE.md)

---

**💡 Tip**: 프로젝트 문서는 언제든 수정 가능합니다. 프로젝트가 진행되면서 요구사항이 변경되면 이 문서들을 업데이트하세요!
