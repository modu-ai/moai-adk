# Part 2: SPEC 작성

> **소요시간**: 약 30분
> **학습 목표**: `/alfred:1-spec` 커맨드를 사용하여 EARS 구조의 SPEC 문서를 작성하고, @TAG 시스템을 이해합니다.

---

## 🎯 이번 Part에서 배울 것

- ✅ `/alfred:1-spec` 커맨드 사용법
- ✅ EARS 요구사항 작성 방법 (5가지 구문)
- ✅ SPEC 문서 구조 (spec.md, plan.md, acceptance.md)
- ✅ Acceptance Criteria (Given-When-Then) 작성법
- ✅ @TAG 체계 이해 및 적용

---

## 📋 사전 준비

### Part 1 완료 확인

```bash
# 프로젝트 문서 확인
ls .moai/project/

# 예상 출력:
# product.md
# structure.md
# tech.md

# SPEC 후보 확인
grep "TODO:SPEC-BACKLOG-001" .moai/project/product.md -A 7
```

**예상 SPEC 후보**:

```markdown
## TODO:SPEC-BACKLOG-001 다음 단계 SPEC 후보

1. **TODO-001**: Todo 항목 CRUD 기능 (생성, 조회, 수정, 삭제)
2. **TODO-002**: Todo 상태 관리 (완료/미완료/진행중)
3. **TODO-003**: Todo 우선순위 설정 (높음/중간/낮음)
4. **TODO-004**: Todo 필터링 및 검색 기능
5. **TODO-005**: 프론트엔드 UI 컴포넌트 (Tailwind CSS)
6. **TODO-006**: Docker Compose 배포 환경 구성
```

---

## 🚀 Step 1: `/alfred:1-spec` 실행

Claude Code에서 다음 커맨드를 실행합니다:

```
/alfred:1-spec
```

> **참고**: 특정 SPEC ID를 지정하려면 `/alfred:1-spec TODO-001` 형태로 실행할 수도 있습니다.

### Alfred의 응답

Alfred가 프로젝트 문서를 분석하고 SPEC 후보를 제안합니다.

**분석 보고서 예시**:

```
📊 SPEC 작성 계획 보고서: my-moai-project

분석 결과:
- 모드: Personal (로컬 개발, Git 미활성화)
- 프로젝트 유형: 풀스택 ToDo 애플리케이션 (학습용)
- 기술 스택: Vite+React + FastAPI + SQLAlchemy 2.0 + Alembic
- 기존 SPEC: 없음 (첫 번째 SPEC 작성)

발굴된 SPEC 후보 (product.md:133 기반):
1. ⭐ TODO-001: Todo 항목 CRUD 기능 (최우선 - 핵심 기능)
2. TODO-002: Todo 상태 관리 (완료/미완료/진행중)
3. TODO-003: Todo 우선순위 설정 (높음/중간/낮음)
4. TODO-004: Todo 필터링 및 검색 기능
5. TODO-005: 프론트엔드 UI 컴포넌트 (Tailwind CSS)
6. TODO-006: Docker Compose 배포 환경 구성

🎯 작성 전략

선택된 SPEC: TODO-001: Todo 항목 CRUD 기능

선택 이유:
- 가장 기본적이고 필수적인 기능
- 프론트엔드-백엔드 풀스택 연동 학습에 최적
- 타입 안전성(TypeScript ↔ Pydantic) 검증 가능
- Alembic 마이그레이션 학습 기회 (첫 테이블 생성)
- 이후 모든 기능(상태 관리, 우선순위, 필터링)의 기반

EARS 구조 설계:

### Ubiquitous (필수 기능)
- 시스템은 Todo 항목 생성, 조회, 수정, 삭제 기능을 제공해야 한다
- 시스템은 SQLAlchemy 2.0 ORM을 통해 데이터를 관리해야 한다
- 시스템은 Alembic을 통해 데이터베이스 스키마를 버전 관리해야 한다

### Event-driven (이벤트 기반)
- WHEN 사용자가 Todo를 생성하면,
  시스템은 고유 ID를 부여하고 DB에 저장해야 한다
- WHEN 사용자가 Todo를 삭제하면,
  시스템은 해당 항목을 DB에서 영구 삭제해야 한다
- WHEN 스키마가 변경되면,
  시스템은 Alembic 마이그레이션 파일을 생성해야 한다

### State-driven (상태 기반)
- WHILE Todo 목록이 비어있을 때,
  시스템은 안내 메시지를 표시해야 한다

### Constraints (제약사항)
- IF Todo 제목이 비어있으면,
  시스템은 생성/수정을 거부하고 400 에러를 반환해야 한다
- Todo 제목은 최대 200자를 초과할 수 없다
- 모든 DB 변경은 Alembic 마이그레이션을 통해서만 수행해야 한다

Acceptance Criteria (Given-When-Then):
1. Todo 생성 시나리오 (POST /api/todos)
2. Todo 조회 시나리오 (GET /api/todos, GET /api/todos/{id})
3. Todo 수정 시나리오 (PATCH /api/todos/{id})
4. Todo 삭제 시나리오 (DELETE /api/todos/{id})
5. Alembic 마이그레이션 (초기 테이블 생성)

⚠️ 주의사항

기술적 제약:
- TypeScript 인터페이스 ↔ Pydantic 스키마 타입 일치 필수
- SQLAlchemy 2.0 모델 정의 (declarative_base 또는 Mapped 스타일)
- Alembic 초기 설정 (alembic init, env.py 구성)
- 마이그레이션 파일 생성 (alembic revision --autogenerate)
- REST API 엔드포인트 설계 (GET/POST/PATCH/DELETE)
- 프론트엔드 상태 관리 전략 (React State/Context)

✅ 예상 산출물

디렉토리 구조:
.moai/specs/SPEC-TODO-001/
├── spec.md           # EARS 구조의 핵심 명세
├── plan.md           # 구현 계획서 (TDD 단계별 + Alembic)
└── acceptance.md     # 인수 기준 (Given-When-Then)

SPEC 메타데이터:
- ID: TODO-001
- 버전: v0.0.1 (INITIAL)
- 우선순위: critical (핵심 기능)
- 예상 구현:
  a. Alembic 초기 설정
  b. Backend API (SQLAlchemy 2.0 ORM)
  c. Frontend UI (React + TypeScript)

@TAG 체계:
- @SPEC:TODO-001 (.moai/specs/SPEC-TODO-001/spec.md)
- @TEST:TODO-001 (backend/tests/, frontend/src/tests/)
- @CODE:TODO-001 (backend/app/, frontend/src/)
- @CODE:TODO-001:DATA (backend/app/models/, alembic/versions/)

예상 작업 시간:
- SPEC 작성: 15-20분
- TDD 구현: 1.5-2.5시간
- 문서 동기화: 10분

---
승인 요청: 위 계획으로 TODO-001: Todo 항목 CRUD 기능 SPEC 작성을 진행하시겠습니까?

- "진행" 또는 "시작": SPEC 문서 작성 시작
- "수정 [내용]": 계획 수정
- "중단": SPEC 작성 중단
```

---

## 💬 Step 2: 승인 및 기술 스택 조정 (선택)

### 기본 승인

계획을 확인했다면 진행하세요:

```
진행
```

### 기술 스택 조정 (필요시)

SPEC 작성 전에 기술 스택을 수정할 수 있습니다.

**예시: ORM 변경**

```
사용자: sqlmodel -> sqlalchemy 2.0, alembic 추가 해서 사용하도록 하자.

Alfred: 알겠습니다! 기술 스택을 수정하겠습니다.

🔄 기술 스택 변경사항:

Before → After:
- ❌ SQLModel → ✅ SQLAlchemy 2.0 (성숙한 ORM, 풍부한 생태계)
- ➕ Alembic (DB 마이그레이션 관리 도구 추가)

변경 이유:
- SQLAlchemy 2.0의 성숙한 ORM 및 풍부한 커뮤니티 지원
- Alembic을 통한 체계적인 데이터베이스 스키마 버전 관리
- 학습 목적으로 마이그레이션 관리 경험 추가

(업데이트된 계획 제시...)

승인 요청: 위 계획으로 SPEC 작성을 진행하시겠습니까?
```

수정된 계획을 확인 후 승인:

```
진행
```

---

## 📝 Step 3: SPEC 문서 생성

Alfred가 spec-builder 에이전트를 호출하여 SPEC 문서를 작성합니다.

### Phase 1: 에이전트 계획 수립

spec-builder 에이전트가 한 번 더 계획을 확인합니다:

```
spec-builder 에이전트가 Phase 2 진행 전 승인을 요청하고 있습니다.

다음 단계로 진행하려면:
- "진행" - SPEC 문서 3개 파일 생성 시작

작업 내용:
1. .moai/specs/SPEC-TODO-001/ 디렉토리 생성
2. spec.md, plan.md, acceptance.md 파일 작성
3. YAML Front Matter, HISTORY, EARS 구조 포함
```

### Phase 2: 파일 생성

```
진행
```

입력 후, Alfred가 3개 파일을 생성합니다 (약 2분 소요).

---

## 📄 Step 4: 생성된 SPEC 문서 확인

### 생성 완료 메시지

```
✅ SPEC-TODO-001 생성 완료!

첫 번째 SPEC 문서가 성공적으로 작성되었습니다.

📁 생성된 파일:
.moai/specs/SPEC-TODO-001/
├── spec.md         (6.3 KB) - EARS 구조의 핵심 명세
├── plan.md         (10 KB)  - TDD 구현 계획
└── acceptance.md   (12 KB)  - 상세한 수락 기준 (9개 시나리오)

📋 SPEC 메타데이터:
- ID: TODO-001
- 제목: Todo 항목 CRUD 기능
- 버전: v0.0.1 (INITIAL)
- 상태: draft
- 우선순위: critical
- 작성자: @Goos
- 범위: Backend + Frontend

🎯 핵심 기능:

Backend (FastAPI + SQLAlchemy 2.0 + Alembic):
- RESTful API 엔드포인트 (GET/POST/PATCH/DELETE /api/todos)
- SQLAlchemy 2.0 모델 (Mapped 타입)
- Alembic 마이그레이션 자동화
- Pydantic v2 스키마 검증

Frontend (Vite + React + TypeScript):
- TodoList, TodoForm, TodoItem 컴포넌트
- TypeScript 인터페이스
- Tailwind CSS 스타일링
- API 클라이언트

📝 Acceptance Criteria (9개):
1. AC-1: Todo 생성 (POST /api/todos)
2. AC-2: Todo 전체 조회 (GET /api/todos)
3. AC-3: Todo 단일 조회 (GET /api/todos/{id})
4. AC-4: Todo 수정 (PATCH /api/todos/{id})
5. AC-5: Todo 삭제 (DELETE /api/todos/{id})
6. AC-6: 제목 필수 검증
7. AC-7: Alembic 마이그레이션
8. AC-8: 프론트엔드 UI 테스트
9. AC-9: 엔드투엔드 통합 테스트
```

### 파일 확인

```bash
# SPEC 디렉토리 확인
ls -lh .moai/specs/SPEC-TODO-001/

# 예상 출력:
# spec.md         (6.3 KB)
# plan.md         (10 KB)
# acceptance.md   (12 KB)

# 각 파일 미리보기
head -30 .moai/specs/SPEC-TODO-001/spec.md
```

---

## 📚 Step 5: SPEC 문서 내용 이해

### spec.md - 핵심 명세

```markdown
---
id: TODO-001
version: 0.0.1
status: draft
priority: critical
created: 2025-10-11
updated: 2025-10-11
author: @Goos
category: feature
labels: [crud, backend, frontend, database]
scope:
  packages:
    - backend/app
    - frontend/src
---

# @SPEC:TODO-001: Todo 항목 CRUD 기능

## HISTORY
### v0.0.1 (2025-10-11)
- **INITIAL**: Todo CRUD 기능 명세 작성
- **AUTHOR**: @Goos
- **REVIEW**: Alfred (spec-builder)

---

## 개요

Todo 애플리케이션의 핵심 기능인 할일(Todo) 항목의
생성, 조회, 수정, 삭제(CRUD) 기능을 구현합니다.

### 목표
- SQLAlchemy 2.0 ORM 기반 데이터 모델 정의
- Alembic을 통한 DB 마이그레이션 관리
- FastAPI REST API 엔드포인트 구현
- React + TypeScript 기반 UI 컴포넌트 구현
- 프론트엔드-백엔드 타입 안전성 보장

---

## EARS 요구사항

### Ubiquitous Requirements (필수 기능)
- 시스템은 Todo 항목의 생성, 조회, 수정, 삭제 기능을 제공해야 한다
- 시스템은 SQLAlchemy 2.0 ORM을 통해 데이터를 관리해야 한다
- 시스템은 Alembic을 통해 데이터베이스 스키마를 버전 관리해야 한다

### Event-driven Requirements (이벤트 기반)
- WHEN 사용자가 Todo를 생성하면,
  시스템은 고유 ID를 부여하고 데이터베이스에 저장해야 한다
- WHEN 사용자가 Todo를 수정하면,
  시스템은 변경 내용을 검증하고 데이터베이스에 반영해야 한다
- WHEN 사용자가 Todo를 삭제하면,
  시스템은 해당 항목을 데이터베이스에서 영구 삭제해야 한다
- WHEN 스키마가 변경되면,
  시스템은 Alembic 마이그레이션 파일을 생성해야 한다

### State-driven Requirements (상태 기반)
- WHILE Todo 목록이 비어있을 때,
  시스템은 안내 메시지를 표시해야 한다
- WHILE Todo 목록이 존재할 때,
  시스템은 각 Todo의 제목과 ID를 표시해야 한다

### Constraints (제약사항)
- IF Todo 제목이 비어있으면,
  시스템은 생성/수정 요청을 거부하고 400 에러를 반환해야 한다
- Todo 제목은 최대 200자를 초과할 수 없다
- 모든 데이터베이스 변경은 Alembic 마이그레이션을 통해서만 수행해야 한다
- API 응답은 Pydantic 스키마를 통해 검증되어야 한다

---

## 데이터 모델

### Todo 모델 (SQLAlchemy 2.0)

```python
# @CODE:TODO-001:DATA | SPEC: SPEC-TODO-001.md

from sqlalchemy import String
from sqlalchemy.orm import Mapped, mapped_column
from app.database import Base

class Todo(Base):
    __tablename__ = "todos"

    id: Mapped[int] = mapped_column(primary_key=True, index=True)
    title: Mapped[str] = mapped_column(String(200), nullable=False)
    description: Mapped[str | None] = mapped_column(String(1000), nullable=True)
```

### Pydantic 스키마

```python
# @CODE:TODO-001:API | SPEC: SPEC-TODO-001.md

from pydantic import BaseModel, Field

class TodoCreate(BaseModel):
    title: str = Field(..., min_length=1, max_length=200)
    description: str | None = Field(None, max_length=1000)

class TodoResponse(BaseModel):
    id: int
    title: str
    description: str | None

    class Config:
        from_attributes = True
```

---

## API 엔드포인트

### 1. Todo 생성

- **URL**: `POST /api/todos`
- **Request**: `TodoCreate`
- **Response**: `TodoResponse` (201 Created)

### 2. Todo 목록 조회

- **URL**: `GET /api/todos`
- **Response**: `List[TodoResponse]` (200 OK)

### 3. Todo 단일 조회

- **URL**: `GET /api/todos/{id}`
- **Response**: `TodoResponse` (200 OK)

### 4. Todo 수정

- **URL**: `PATCH /api/todos/{id}`
- **Request**: `TodoUpdate` (부분 수정 허용)
- **Response**: `TodoResponse` (200 OK)
- **참고**: PATCH는 리소스의 부분 업데이트를 의미하며, 제공된 필드만 수정됩니다

### 5. Todo 삭제

- **URL**: `DELETE /api/todos/{id}`
- **Response**: `204 No Content`

```

### plan.md - 구현 계획

```markdown
# @SPEC:TODO-001 구현 계획

## TDD 구현 순서

### Phase 1: Alembic 초기 설정 (20분)

1. Alembic 설치 및 초기화

   ```bash
   pip install alembic
   alembic init alembic
   ```

1. env.py 설정

   - 데이터베이스 URL 설정
   - Base 메타데이터 임포트

1. 첫 마이그레이션 생성

   ```bash
   alembic revision --autogenerate -m "Create todos table"
   alembic upgrade head
   ```

### Phase 2: Backend TDD (1.5시간)

#### RED: 실패하는 테스트 작성 (30분)

```python
# @TEST:TODO-001 | SPEC: SPEC-TODO-001.md

# tests/test_todos.py

def test_create_todo():
    response = client.post("/api/todos", json={"title": "Test Todo"})
    assert response.status_code == 201
    assert response.json()["title"] == "Test Todo"

def test_get_todos():
    response = client.get("/api/todos")
    assert response.status_code == 200
    assert isinstance(response.json(), list)

def test_create_todo_without_title():
    response = client.post("/api/todos", json={})
    assert response.status_code == 400
```

#### GREEN: 최소 구현 (60분)

**1. SQLAlchemy 모델**

```python
# @CODE:TODO-001:DATA | SPEC: SPEC-TODO-001.md
# app/models/todo.py

class Todo(Base):
    __tablename__ = "todos"
    id: Mapped[int] = mapped_column(primary_key=True)
    title: Mapped[str] = mapped_column(String(200))
    description: Mapped[str | None] = mapped_column(String(1000))
```

**2. API 엔드포인트**

```python
# @CODE:TODO-001:API | SPEC: SPEC-TODO-001.md
# app/api/todos.py

@router.post("/", response_model=TodoResponse, status_code=201)
def create_todo(todo: TodoCreate, db: Session = Depends(get_db)):
    db_todo = Todo(**todo.dict())
    db.add(db_todo)
    db.commit()
    db.refresh(db_todo)
    return db_todo
```

#### REFACTOR: 코드 개선 (30분)

- 서비스 레이어 분리
- 중복 코드 제거
- 타입 힌트 개선

### Phase 3: Frontend 구현 (1시간)

#### TypeScript 타입

```typescript
// @CODE:TODO-001:UI | SPEC: SPEC-TODO-001.md
// src/types/todo.ts

export interface Todo {
  id: number;
  title: string;
  description?: string;
}
```

#### React 컴포넌트

```tsx
// @CODE:TODO-001:UI | SPEC: SPEC-TODO-001.md
// src/components/TodoList.tsx

export function TodoList() {
  const [todos, setTodos] = useState<Todo[]>([]);

  useEffect(() => {
    todosApi.getAll().then(setTodos);
  }, []);

  if (todos.length === 0) {
    return <p>할일이 없습니다.</p>;
  }

  return (
    <ul>
      {todos.map(todo => (
        <TodoItem key={todo.id} todo={todo} />
      ))}
    </ul>
  );
}
```

---

## 예상 일정

| 단계 | 작업 | 예상 시간 |
|------|------|----------|
| 1 | Alembic 설정 | 20분 |
| 2 | Backend TDD | 1.5시간 |
| 3 | Frontend 구현 | 1시간 |
| 4 | 테스트 및 검증 | 30분 |
| **총계** | | **3시간 20분** |

```

### acceptance.md - 인수 기준

```markdown
# @SPEC:TODO-001 Acceptance Criteria

## AC-1: Todo 생성 (POST /api/todos)

### Given-When-Then

**Scenario 1: 유효한 Todo 생성**

```gherkin
Given 사용자가 Todo 생성 요청을 보낸다
  And 요청 본문에 제목 "장보기"가 포함되어 있다
When POST /api/todos 엔드포인트를 호출한다
Then 응답 상태 코드는 201이어야 한다
  And 응답 본문에 생성된 Todo의 id가 포함되어야 한다
  And 응답 본문의 title은 "장보기"여야 한다
```

**테스트 코드**:

```python
# @TEST:TODO-001 | SPEC: SPEC-TODO-001.md

def test_create_todo_success():
    # Given
    payload = {"title": "장보기", "description": "우유, 계란 사기"}

    # When
    response = client.post("/api/todos", json=payload)

    # Then
    assert response.status_code == 201
    data = response.json()
    assert "id" in data
    assert data["title"] == "장보기"
    assert data["description"] == "우유, 계란 사기"
```

**Scenario 2: 제목 없이 Todo 생성 (실패)**

```gherkin
Given 사용자가 Todo 생성 요청을 보낸다
  And 요청 본문에 제목이 없다
When POST /api/todos 엔드포인트를 호출한다
Then 응답 상태 코드는 400이어야 한다
  And 에러 메시지는 "제목은 필수입니다"를 포함해야 한다
```

---

## AC-2: Todo 전체 조회 (GET /api/todos)

**Scenario 1: 빈 목록 조회**

```gherkin
Given 데이터베이스에 Todo가 하나도 없다
When GET /api/todos 엔드포인트를 호출한다
Then 응답 상태 코드는 200이어야 한다
  And 응답 본문은 빈 배열 []이어야 한다
```

**Scenario 2: Todo가 있는 목록 조회**

```gherkin
Given 데이터베이스에 Todo가 2개 있다
When GET /api/todos 엔드포인트를 호출한다
Then 응답 상태 코드는 200이어야 한다
  And 응답 본문은 2개의 Todo를 포함한 배열이어야 한다
```

---

(나머지 AC-3 ~ AC-9 시나리오...)

```

---

## 📚 EARS 및 @TAG 시스템 참고

### EARS 요구사항 작성법

**EARS (Easy Approach to Requirements Syntax)**는 체계적인 요구사항 작성을 위한 5가지 구문을 제공합니다:

- **Ubiquitous**: 시스템은 [기능]을 제공해야 한다
- **Event-driven**: WHEN [조건]이면, 시스템은 [동작]해야 한다
- **State-driven**: WHILE [상태]일 때, 시스템은 [동작]해야 한다
- **Optional**: WHERE [조건]이면, 시스템은 [동작]할 수 있다
- **Constraints**: IF [조건]이면, 시스템은 [제약]해야 한다

> **상세 가이드**: [EARS 가이드](../../concepts/ears-guide.md)에서 실전 예제와 작성 패턴을 확인하세요.

### @TAG 추적성 시스템

**@TAG**는 SPEC → TEST → CODE → DOC의 추적성을 보장합니다:

```
@SPEC:ID → @TEST:ID → @CODE:ID → @DOC:ID
```

**주요 TAG**:
- `@SPEC:TODO-001` - SPEC 문서 (.moai/specs/)
- `@TEST:TODO-001` - 테스트 코드 (tests/)
- `@CODE:TODO-001` - 구현 코드 (src/)
- `@DOC:TODO-001` - 문서 (docs/)

> **상세 가이드**: [TAG 시스템](../../concepts/tag-system.md)에서 서브 카테고리와 검증 방법을 확인하세요.

---

## ✅ 완료 확인

### 파일 생성 확인

```bash
# SPEC 디렉토리 확인
ls -lh .moai/specs/SPEC-TODO-001/

# 예상 출력:
# spec.md         (6.3 KB)  - 핵심 명세
# plan.md         (10 KB)   - 구현 계획
# acceptance.md   (12 KB)   - 인수 기준

# YAML Front Matter 확인
head -20 .moai/specs/SPEC-TODO-001/spec.md
```

### 품질 체크리스트

- [ ] `.moai/specs/SPEC-TODO-001/` 디렉토리 생성
- [ ] `spec.md` 파일 존재 및 @SPEC:TODO-001 TAG 포함
- [ ] `plan.md` 파일 존재 및 TDD 단계 포함
- [ ] `acceptance.md` 파일 존재 및 Given-When-Then 시나리오 포함
- [ ] YAML Front Matter 메타데이터 완성 (id, version, status, priority)
- [ ] HISTORY 섹션 포함 (v0.0.1 INITIAL)
- [ ] EARS 5가지 구문 모두 포함
- [ ] API 엔드포인트 정의 완성
- [ ] 데이터 모델 정의 완성

---

## 🎓 학습 정리

### 핵심 개념

1. **SPEC-First 개발**: 코드 작성 전에 명세를 먼저 정의
2. **EARS 방법론**: 체계적인 요구사항 작성 (5가지 구문)
3. **Acceptance Criteria**: Given-When-Then 형식의 명확한 인수 기준
4. **@TAG 시스템**: SPEC → 테스트 → 코드의 추적성 보장

### SPEC 문서 구조

- **spec.md**: 핵심 명세 (EARS, 데이터 모델, API)
- **plan.md**: 구현 계획 (TDD 단계별 작업)
- **acceptance.md**: 인수 기준 (Given-When-Then 시나리오)

### 메타데이터 필드

- `id`: SPEC 고유 ID (TODO-001)
- `version`: 버전 (v0.0.1 = INITIAL, draft)
- `status`: 상태 (draft → active → completed)
- `priority`: 우선순위 (critical, high, medium, low)
- `category`: 변경 유형 (feature, bugfix, refactor)
- `labels`: 분류 태그 (검색용)

---

## 🔍 트러블슈팅

### 문제 1: SPEC ID가 중복되었어요

**확인**:

```bash
rg "@SPEC:TODO-001" -n .moai/specs/
```

**해결책**:

- 중복되지 않는 새 ID 사용
- Alfred에게 중복 확인 요청

### 문제 2: EARS 구문을 어떻게 적용해야 할지 모르겠어요

**해결책**:

- spec.md의 예시 참고
- [EARS 가이드](../../concepts/ears-guide.md) 문서 참고
- Alfred에게 "EARS 예시 더 보여줘" 요청

### 문제 3: acceptance.md가 너무 길어요

**해결책**:

- 핵심 시나리오만 포함 (2-3개)
- 나머지는 구현 단계에서 추가

---

## 🚀 다음

SPEC 문서 작성이 완료되었습니다! 이제 TDD로 구현할 준비가 되었습니다.

**다음**: [Part 3: Backend TDD 구현하기](./03-backend-tdd.md)

**이전**: [Part 1: 프로젝트 초기화](./01-project-init.md)

**커맨드**:

```
/alfred:2-build TODO-001
```

---

## 📚 참고 자료

- [EARS 가이드](../../concepts/ears-guide.md)
- [TAG 시스템](../../concepts/tag-system.md)
- [SPEC 메타데이터 표준](https://github.com/modu-ai/moai-adk/blob/main/.moai/memory/spec-metadata.md)
- [워크플로우 가이드: 1-spec](../../workflow/1-spec.md)

---

**💡 Tip**: SPEC 문서는 살아있는 문서(Living Document)입니다. 구현 중 요구사항이 변경되면 SPEC을 업데이트하고 HISTORY에 기록하세요!
