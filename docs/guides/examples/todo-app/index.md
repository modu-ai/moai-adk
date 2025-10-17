# ToDo 앱 튜토리얼

> **학습 목표**: MoAI-ADK의 SPEC-First TDD 방법론을 학습하고, 실제 풀스택 애플리케이션을 단계별로 구축합니다.

이 튜토리얼은 MoAI-ADK를 사용하여 학습용 ToDo 애플리케이션을 처음부터 끝까지 만드는 과정을 안내합니다. 프로젝트 초기화부터 SPEC 작성, TDD 구현, 문서 동기화까지 전체 워크플로우를 실습합니다.

---

## 📋 프로젝트 개요

### 완성 예상도

```
┌─────────────────────────────────────┐
│  📝 Todo App                        │
├─────────────────────────────────────┤
│  [ ] 장보기 (높음) ✏️ 🗑️             │
│  [✓] 보고서 작성 (중간) ✏️ 🗑️         │
│  [ ] 운동하기 (낮음) ✏️ 🗑️            │
│                                     │
│  [➕ 새 할일 추가]                   │
└─────────────────────────────────────┘
```

### 기술 스택

**Frontend**:

- Vite + React 18
- TypeScript 5.x
- Tailwind CSS

**Backend**:

- FastAPI
- SQLAlchemy 2.0 + Alembic
- Pydantic v2
- Python 3.11+

**Database**:

- SQLite (로컬 개발)

**배포**:

- Docker Compose (로컬)

### 주요 기능

- ✅ Todo CRUD (생성, 조회, 수정, 삭제)
- ✅ 상태 관리 (완료/미완료/진행중)
- ✅ 우선순위 설정 (높음/중간/낮음)
- ✅ 필터링 및 검색

### 학습 포인트

- ✅ **SPEC-First 접근법** - 코드 작성 전 명세 우선
- ✅ **EARS 요구사항 작성** - 체계적인 요구사항 정의
- ✅ **TDD 사이클** - Red-Green-Refactor 실습
- ✅ **@TAG 추적성** - SPEC → 테스트 → 코드 연결
- ✅ **타입 안전성** - TypeScript ↔ Pydantic 타입 계약
- ✅ **DB 마이그레이션** - Alembic 사용법

---

## 🗺️ 튜토리얼 구성

### Part 1: 프로젝트 초기화 (30분)

**[프로젝트 초기화하기](./01-project-init.md)**

- 프로젝트 생성 및 MoAI-ADK 초기화
- `/alfred:0-project` 커맨드 실행
- 프로젝트 문서 작성 (product, structure, tech)

**학습 내용**:

- MoAI-ADK 프로젝트 구조
- Product Discovery 인터뷰
- 기술 스택 선택 전략

**산출물**:

- `.moai/project/product.md`
- `.moai/project/structure.md`
- `.moai/project/tech.md`
- `.moai/config.json`

---

### Part 2: SPEC 작성 (30분)

**[첫 번째 SPEC 작성하기](./02-spec-writing.md)**

- `/alfred:1-spec` 커맨드 실행
- EARS 요구사항 작성법 학습
- TODO-001: Todo CRUD 기능 명세

**학습 내용**:

- SPEC 문서 구조
- EARS 5가지 구문
- Acceptance Criteria 작성법
- @TAG 체계 이해

**산출물**:

- `.moai/specs/SPEC-TODO-001/spec.md`
- `.moai/specs/SPEC-TODO-001/plan.md`
- `.moai/specs/SPEC-TODO-001/acceptance.md`

---

### Part 3: Backend TDD 구현 (2시간)

**[Backend API 구현하기](./03-backend-tdd.md)**

- Alembic 초기 설정
- SQLAlchemy 2.0 모델 정의
- FastAPI REST API 구현
- pytest 테스트 작성

**학습 내용**:

- RED: 실패하는 테스트 작성
- GREEN: 최소 구현
- REFACTOR: 코드 품질 개선
- Alembic 마이그레이션

**산출물**:

- `backend/app/models/todo.py`
- `backend/app/api/todos.py`
- `backend/app/schemas/todo.py`
- `backend/tests/test_todos.py`
- `alembic/versions/xxx_create_todos_table.py`

---

### Part 4: Frontend 구현 (1.5시간)

**[Frontend UI 구현하기](./04-frontend-impl.md)**

- React 컴포넌트 작성
- TypeScript 타입 정의
- Tailwind CSS 스타일링
- Vitest 테스트 작성

**학습 내용**:

- TodoList, TodoForm, TodoItem 컴포넌트
- TypeScript 인터페이스
- API 클라이언트 구현
- 컴포넌트 테스트

**산출물**:

- `frontend/src/components/TodoList.tsx`
- `frontend/src/components/TodoForm.tsx`
- `frontend/src/components/TodoItem.tsx`
- `frontend/src/api/todos.ts`
- `frontend/src/types/todo.ts`

---

### Part 5: 문서 동기화 (20분)

**[문서 동기화 및 배포](./05-sync-deploy.md)**

- `/alfred:3-sync` 커맨드 실행
- Living Document 생성
- @TAG 체인 검증
- 통합 테스트 및 배포 준비

**학습 내용**:

- 문서 자동 생성
- TAG 무결성 검증
- 통합 테스트 전략

**산출물**:

- `.moai/sync-report.md`
- SPEC 버전 업데이트 (v0.0.1 → v0.1.0)
- 실행 가능한 풀스택 애플리케이션

---

## 🚀 시작하기

### 사전 요구사항

```bash
# 필수
✅ Node.js 18+ (Frontend)
✅ Python 3.11+ (Backend)
✅ MoAI-ADK 설치 완료

# 권장
🔲 Docker Desktop (배포 테스트용)
🔲 Visual Studio Code + Claude Code
```

### 설치 확인

```bash
# MoAI-ADK CLI
moai --version

# Node.js
node --version  # v18 이상

# Python
python --version  # 3.11 이상

# Docker (선택)
docker --version
```

### 프로젝트 생성

```bash
# 1. 프로젝트 디렉토리 생성
mkdir my-moai-project
cd my-moai-project

# 2. MoAI-ADK 초기화
moai init .
```

**예상 출력**:

```
✅ MoAI-ADK 프로젝트가 초기화되었습니다!

생성된 파일:
  .moai/
  ├── config.json       # 프로젝트 설정
  ├── project/          # 프로젝트 문서 (템플릿)
  ├── specs/            # SPEC 문서 저장소
  └── memory/           # 개발 가이드

다음 단계:
1. /alfred:0-project - 프로젝트 문서 작성
2. /alfred:1-spec - 첫 번째 SPEC 작성
```

---

## 📖 학습 순서

### 추천 학습 경로

1. **[Part 1: 프로젝트 초기화](./01-project-init.md)** ⬅️ 여기서 시작!
2. [Part 2: SPEC 작성](./02-spec-writing.md)
3. [Part 3: Backend TDD 구현](./03-backend-tdd.md)
4. [Part 4: Frontend 구현](./04-frontend-impl.md)
5. [Part 5: 문서 동기화 및 배포](./05-sync-deploy.md)

### 예상 소요 시간

| Part | 내용 | 시간 |
|------|------|------|
| Part 1 | 프로젝트 초기화 | 30분 |
| Part 2 | SPEC 작성 | 30분 |
| Part 3 | Backend TDD | 2시간 |
| Part 4 | Frontend 구현 | 1.5시간 |
| Part 5 | 동기화 및 배포 | 20분 |
| **총계** | | **약 5시간** |

---

## 💡 학습 팁

### 처음 학습하시는 분

- ⏸️ 각 Part를 하나씩 완료한 후 다음으로 넘어가세요
- 📝 코드를 직접 타이핑하면서 따라하세요 (복사 붙여넣기 X)
- 🤔 각 단계에서 "왜?"를 생각하며 학습하세요
- ❓ 막히는 부분은 [트러블슈팅 섹션](#-트러블슈팅)을 참고하세요

### 경험이 있으신 분

- 🚀 관심 있는 Part부터 골라서 학습하세요
- 🎯 자신의 프로젝트에 맞게 커스터마이징하세요
- 🔧 더 나은 구현 방법을 실험해보세요

---

## 🎯 학습 체크리스트

완성 후 다음 항목들을 확인할 수 있습니다:

### 프로젝트 구조

- [ ] `.moai/project/` - 프로젝트 문서 (product, structure, tech)
- [ ] `.moai/specs/` - SPEC 문서들
- [ ] `backend/` - FastAPI 백엔드
- [ ] `frontend/` - React 프론트엔드
- [ ] (선택) Docker 배포 설정

### 기능 구현

- [ ] Todo 생성 (POST /api/todos)
- [ ] Todo 조회 (GET /api/todos, GET /api/todos/:id)
- [ ] Todo 수정 (PUT /api/todos/:id)
- [ ] Todo 삭제 (DELETE /api/todos/:id)
- [ ] 프론트엔드 UI 동작

### 품질 검증

- [ ] Backend 테스트 커버리지 85% 이상
- [ ] Frontend 컴포넌트 테스트 작성
- [ ] @TAG 체계 완성
- [ ] 통합 테스트 통과 (Backend + Frontend)

### 학습 목표

- [ ] SPEC-First 개발 프로세스 이해
- [ ] EARS 요구사항 작성법 숙지
- [ ] TDD Red-Green-Refactor 사이클 체득
- [ ] @TAG 시스템 활용 능력
- [ ] 풀스택 타입 안전성 보장 방법 습득

---

## 🔍 트러블슈팅

### 자주 묻는 질문

**Q: Personal 모드에서 Git 작업이 필요한가요?**

A: Personal 모드는 로컬 개발에 집중하도록 설계되어 Git이 선택사항입니다. 원하시면 수동으로 커밋을 만들거나, `.moai/config.json`에서 `mode: "team"`으로 변경하세요.

**Q: 인터뷰 질문이 너무 많아요**

A: "네 나머지도 알아서 답변을 해주세요"라고 입력하면 Alfred가 합리적인 답변을 자동으로 작성합니다.

**Q: 기술 스택을 바꾸고 싶어요**

A: `/alfred:1-spec` Phase 1에서 기술 스택을 수정할 수 있습니다. Alfred가 관련 문서를 자동으로 업데이트합니다.

**Q: Docker 없이도 실습 가능한가요?**

A: 네! Part 1-4는 로컬 환경에서 실습 가능합니다. Part 5의 Docker 배포는 선택사항입니다.

### 추가 도움말

- [설치 가이드](../../installation.md)
- [빠른 시작](../../quick-start.md)
- [워크플로우 가이드](../../workflow/overview.md)
- [GitHub Issues](https://github.com/modu-ai/moai-adk/issues)

---

## 🚀 다음 단계

준비가 되셨나요? 첫 번째 단계부터 시작하세요!

**👉 [Part 1: 프로젝트 초기화하기](./01-project-init.md)**
