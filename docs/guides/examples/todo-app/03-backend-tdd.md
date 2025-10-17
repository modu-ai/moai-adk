# Part 3: Backend TDD 구현

> **소요시간**: 약 2시간
> **학습 목표**: RED-GREEN-REFACTOR 사이클을 실전에서 적용하고, SQLAlchemy 2.0 + FastAPI 기반 REST API를 완벽하게 구현합니다.

---

## 개요

이제 `/alfred:2-build` 커맨드로 실제 TDD 구현을 시작합니다. Alfred는 다음 순서로 작업합니다:

1. **라이브러리 버전 확인** (웹 검색)
2. **프로젝트 구조 생성**
3. **RED**: 실패하는 테스트 작성
4. **GREEN**: 테스트를 통과하는 최소 구현
5. **REFACTOR**: 코드 품질 개선

## Step 1: TDD 구현 시작

### 1.1 `/alfred:2-build` 실행

```bash
/alfred:2-build TODO-001
```

Alfred가 SPEC-TODO-001을 분석하고 구현 계획을 제시합니다:

```
📊 분석 결과

- SPEC ID: TODO-001 (Todo 항목 CRUD 기능)
- 프로젝트 타입: Fullstack (FastAPI + Vite React)
- 복잡도: 중간
- 예상 작업시간: 4-6시간 (TDD 사이클 포함)
- 현재 상태: 빈 프로젝트 (초기 구조 생성 필요)
```

### 1.2 라이브러리 버전 확인 요청

최신 안정 버전을 사용하기 위해 다음과 같이 지시합니다:

```
사용하는 라이브러리는 오늘 기준으로 안정버전만 사용하도록 한다.
웹검색해서 버전을 확인하고 버전은 항상 명시해서 해당 버전 이상만 사용하도록 하자.
```

Alfred가 웹 검색으로 최신 버전을 확인합니다:

| 패키지 | 최신 안정 버전 | 설치 명령 |
|--------|----------------|-----------|
| FastAPI | 0.118.3 | `fastapi>=0.118.3` |
| SQLAlchemy | 2.0.43 | `sqlalchemy>=2.0.43` |
| Alembic | 1.16.5 | `alembic>=1.16.5` |
| Pydantic | 2.12.0 | `pydantic>=2.12.0` |
| pytest | 8.4.2 | `pytest>=8.4.2` |

### 1.3 구현 승인

```
진행
```

Alfred가 즉시 TDD 구현을 시작합니다.

---

## Step 2: 프로젝트 구조 생성

### 2.1 백엔드 디렉토리 구조

Alfred가 자동으로 생성하는 구조:

```
backend/
├── app/
│   ├── __init__.py
│   ├── main.py              # FastAPI 애플리케이션 진입점
│   ├── database.py          # DB 세션 관리
│   ├── models/
│   │   ├── __init__.py
│   │   └── todo.py          # @CODE:TODO-001:DATA
│   ├── schemas/
│   │   ├── __init__.py
│   │   └── todo.py          # Pydantic 스키마
│   ├── services/
│   │   ├── __init__.py
│   │   └── todo_service.py  # @CODE:TODO-001:DOMAIN
│   └── routers/
│       ├── __init__.py
│       └── todos.py         # @CODE:TODO-001:API
├── tests/
│   ├── __init__.py
│   ├── conftest.py          # pytest 설정
│   └── test_todos.py        # @TEST:TODO-001
├── alembic/
│   ├── versions/            # 마이그레이션 파일
│   └── env.py               # Alembic 설정
├── requirements.txt
├── pytest.ini
├── pyproject.toml           # ruff 설정
└── alembic.ini
```

### 2.2 의존성 파일 생성

**`requirements.txt`**:

```python
# Backend Dependencies (Python 3.11+)
# 2025-10-11 기준 최신 안정 버전

# Web Framework
fastapi>=0.118.3
uvicorn[standard]>=0.34.0

# ORM & Database
sqlalchemy>=2.0.43
alembic>=1.16.5

# Data Validation
pydantic>=2.12.0
pydantic-settings>=2.7.0

# Testing
pytest>=8.4.2
pytest-asyncio>=0.24.0
pytest-cov>=6.0.0
httpx>=0.28.0

# Development Tools
ruff>=0.8.0
mypy>=1.14.0
```

**`pytest.ini`**:

```ini
[pytest]
testpaths = tests
python_files = test_*.py
python_classes = Test*
python_functions = test_*

addopts =
    --verbose
    --cov=app
    --cov-report=term-missing
    --cov-report=html
    --cov-fail-under=85
```

**`pyproject.toml`** (ruff 설정):

```toml
[tool.ruff]
line-length = 100
target-version = "py311"

[tool.ruff.lint]
select = ["E", "F", "I", "N", "W", "UP"]
ignore = []

[tool.mypy]
python_version = "3.11"
warn_return_any = true
warn_unused_configs = true
disallow_untyped_defs = true
```

### 2.3 가상환경 및 의존성 설치

```bash
cd backend
python3 -m venv .venv
.venv/bin/pip install --upgrade pip setuptools wheel
.venv/bin/pip install -r requirements.txt
```

출력:

```
Successfully installed fastapi-0.118.3 sqlalchemy-2.0.44 alembic-1.16.5 ...
```

---

## Step 3: 데이터 레이어 구현

### 3.1 데이터베이스 연결 설정

**`app/database.py`**:

```python
# @CODE:TODO-001:DATA | SPEC: SPEC-TODO-001.md
"""
Database connection and session management
"""
from typing import Generator

from sqlalchemy import create_engine
from sqlalchemy.orm import DeclarativeBase, Session, sessionmaker

SQLALCHEMY_DATABASE_URL = "sqlite:///./todos.db"

engine = create_engine(
    SQLALCHEMY_DATABASE_URL,
    connect_args={"check_same_thread": False},  # SQLite only
    echo=True,  # 개발 시 SQL 쿼리 로깅
)

SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)


class Base(DeclarativeBase):
    """Base class for all models"""

    pass


def get_db() -> Generator[Session, None, None]:
    """Dependency for FastAPI routes"""
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()
```

**핵심 포인트**:
- SQLAlchemy 2.0의 `DeclarativeBase` 사용
- `get_db()`는 FastAPI Dependency Injection용
- `Generator` 타입 힌트로 mypy strict 모드 대응

### 3.2 SQLAlchemy 모델 정의

**`app/models/todo.py`**:

```python
# @CODE:TODO-001:DATA | SPEC: SPEC-TODO-001.md
"""
Todo SQLAlchemy Model
"""
from datetime import datetime
from typing import Optional

from sqlalchemy import String, func
from sqlalchemy.orm import Mapped, mapped_column

from app.database import Base


class Todo(Base):
    __tablename__ = "todos"

    id: Mapped[int] = mapped_column(primary_key=True, index=True)
    title: Mapped[str] = mapped_column(String(200), nullable=False)
    description: Mapped[Optional[str]] = mapped_column(String(1000), nullable=True)
    completed: Mapped[bool] = mapped_column(default=False, nullable=False)
    created_at: Mapped[datetime] = mapped_column(
        default=datetime.utcnow, server_default=func.now(), nullable=False
    )
    updated_at: Mapped[datetime] = mapped_column(
        default=datetime.utcnow,
        onupdate=datetime.utcnow,
        server_default=func.now(),
        nullable=False,
    )
```

**핵심 포인트**:
- SQLAlchemy 2.0 `Mapped` 타입 사용
- `mapped_column`으로 컬럼 정의
- `func.now()`로 서버 시간 기본값 설정
- `Optional[str]`로 nullable 명시

### 3.3 Pydantic 스키마 정의

**`app/schemas/todo.py`**:

```python
# @CODE:TODO-001:DATA | SPEC: SPEC-TODO-001.md
"""
Pydantic schemas for Todo validation
"""
from datetime import datetime
from typing import Optional

from pydantic import BaseModel, ConfigDict, Field


class TodoCreate(BaseModel):
    """Schema for creating a new todo"""

    title: str = Field(..., min_length=1, max_length=200, description="Todo title")
    description: Optional[str] = Field(
        None, max_length=1000, description="Todo description"
    )


class TodoUpdate(BaseModel):
    """Schema for updating a todo"""

    title: Optional[str] = Field(None, min_length=1, max_length=200)
    description: Optional[str] = Field(None, max_length=1000)
    completed: Optional[bool] = None


class TodoResponse(BaseModel):
    """Schema for todo response"""

    model_config = ConfigDict(from_attributes=True)

    id: int
    title: str
    description: Optional[str]
    completed: bool
    created_at: datetime
    updated_at: datetime
```

**핵심 포인트**:
- Pydantic v2 `ConfigDict` 사용
- `from_attributes=True`로 SQLAlchemy 모델 변환 지원
- `Field(...)`로 검증 규칙 명시

### 3.4 Alembic 마이그레이션 설정

**`alembic.ini`** (자동 생성):

```ini
[alembic]
script_location = alembic
sqlalchemy.url = sqlite:///./todos.db
```

**`alembic/env.py`**:

```python
# @CODE:TODO-001:DATA | SPEC: SPEC-TODO-001.md
from logging.config import fileConfig
from sqlalchemy import engine_from_config, pool
from alembic import context

# Import Base and models
from app.database import Base
from app.models import todo  # 모델 import 필수!

config = context.config
fileConfig(config.config_file_name)
target_metadata = Base.metadata

# ... (Alembic 기본 코드)
```

**마이그레이션 생성 및 실행**:

```bash
.venv/bin/alembic revision --autogenerate -m "Create todos table"
.venv/bin/alembic upgrade head
```

출력:

```
INFO  [alembic.runtime.migration] Running upgrade  -> 29e260b1a897, Create todos table
```

**확인**:

```bash
sqlite3 todos.db ".schema todos"
```

---

## Step 4: RED - 실패하는 테스트 작성

### 4.1 pytest 설정

**`tests/conftest.py`**:

```python
# @TEST:TODO-001 | SPEC: SPEC-TODO-001.md
"""
Pytest configuration and fixtures
"""
import pytest
from fastapi.testclient import TestClient
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker
from sqlalchemy.pool import StaticPool

from app.database import Base, get_db
from app.main import app

# 테스트용 인메모리 SQLite 데이터베이스
SQLALCHEMY_DATABASE_URL = "sqlite:///:memory:"

engine = create_engine(
    SQLALCHEMY_DATABASE_URL,
    connect_args={"check_same_thread": False},
    poolclass=StaticPool,
)
TestingSessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)


def override_get_db():
    """Override database dependency for testing"""
    try:
        db = TestingSessionLocal()
        yield db
    finally:
        db.close()


@pytest.fixture(scope="function")
def client():
    """Create test client with fresh database"""
    Base.metadata.create_all(bind=engine)
    app.dependency_overrides[get_db] = override_get_db
    with TestClient(app) as test_client:
        yield test_client
    Base.metadata.drop_all(bind=engine)
    app.dependency_overrides.clear()
```

**핵심 포인트**:
- 각 테스트마다 새로운 인메모리 DB 생성 (`scope="function"`)
- `app.dependency_overrides`로 DB 의존성 교체
- 테스트 후 자동 정리

### 4.2 CRUD 테스트 작성

**`tests/test_todos.py`**:

```python
# @TEST:TODO-001 | SPEC: SPEC-TODO-001.md
"""
RED Phase: Integration tests for Todo API endpoints
These tests will fail until we implement the API routes and services
"""
import pytest
from fastapi.testclient import TestClient


class TestTodoCreate:
    """Test POST /api/todos"""

    def test_create_todo_success(self, client: TestClient):
        """GIVEN 유효한 todo 데이터 WHEN POST 요청 THEN 201 응답"""
        payload = {"title": "장보기", "description": "우유, 계란 사기"}
        response = client.post("/api/todos", json=payload)

        assert response.status_code == 201
        data = response.json()
        assert data["title"] == "장보기"
        assert data["description"] == "우유, 계란 사기"
        assert data["completed"] is False
        assert "id" in data
        assert "created_at" in data

    def test_create_todo_missing_title(self, client: TestClient):
        """GIVEN title 누락 WHEN POST 요청 THEN 422 응답"""
        payload = {"description": "설명만 있음"}
        response = client.post("/api/todos", json=payload)

        assert response.status_code == 422

    def test_create_todo_empty_title(self, client: TestClient):
        """GIVEN 빈 title WHEN POST 요청 THEN 422 응답"""
        payload = {"title": "", "description": "빈 제목"}
        response = client.post("/api/todos", json=payload)

        assert response.status_code == 422


class TestTodoRead:
    """Test GET /api/todos"""

    def test_get_all_todos_empty(self, client: TestClient):
        """GIVEN 빈 DB WHEN GET 요청 THEN 빈 배열 반환"""
        response = client.get("/api/todos")

        assert response.status_code == 200
        assert response.json() == []

    def test_get_all_todos_with_data(self, client: TestClient):
        """GIVEN 2개 todo WHEN GET 요청 THEN 2개 반환"""
        # Arrange
        client.post("/api/todos", json={"title": "할일 1"})
        client.post("/api/todos", json={"title": "할일 2"})

        # Act
        response = client.get("/api/todos")

        # Assert
        assert response.status_code == 200
        data = response.json()
        assert len(data) == 2

    def test_get_todo_by_id_success(self, client: TestClient):
        """GIVEN 특정 todo WHEN GET /{id} 요청 THEN 해당 todo 반환"""
        # Arrange
        create_response = client.post("/api/todos", json={"title": "특정 할일"})
        todo_id = create_response.json()["id"]

        # Act
        response = client.get(f"/api/todos/{todo_id}")

        # Assert
        assert response.status_code == 200
        data = response.json()
        assert data["id"] == todo_id
        assert data["title"] == "특정 할일"

    def test_get_todo_by_id_not_found(self, client: TestClient):
        """GIVEN 존재하지 않는 ID WHEN GET 요청 THEN 404 응답"""
        response = client.get("/api/todos/9999")

        assert response.status_code == 404


class TestTodoUpdate:
    """Test PATCH /api/todos/{id}"""

    def test_update_todo_title(self, client: TestClient):
        """GIVEN 기존 todo WHEN 제목 수정 THEN 200 응답"""
        # Arrange
        create_response = client.post("/api/todos", json={"title": "원래 제목"})
        todo_id = create_response.json()["id"]

        # Act
        response = client.patch(
            f"/api/todos/{todo_id}", json={"title": "수정된 제목"}
        )

        # Assert
        assert response.status_code == 200
        data = response.json()
        assert data["title"] == "수정된 제목"

    def test_update_todo_completed(self, client: TestClient):
        """GIVEN 미완료 todo WHEN completed=True THEN 완료 상태 변경"""
        # Arrange
        create_response = client.post("/api/todos", json={"title": "할일"})
        todo_id = create_response.json()["id"]

        # Act
        response = client.patch(f"/api/todos/{todo_id}", json={"completed": True})

        # Assert
        assert response.status_code == 200
        data = response.json()
        assert data["completed"] is True

    def test_update_todo_not_found(self, client: TestClient):
        """GIVEN 존재하지 않는 ID WHEN PATCH 요청 THEN 404 응답"""
        response = client.patch("/api/todos/9999", json={"title": "수정"})

        assert response.status_code == 404


class TestTodoDelete:
    """Test DELETE /api/todos/{id}"""

    def test_delete_todo_success(self, client: TestClient):
        """GIVEN 기존 todo WHEN DELETE 요청 THEN 204 응답"""
        # Arrange
        create_response = client.post("/api/todos", json={"title": "삭제할 할일"})
        todo_id = create_response.json()["id"]

        # Act
        response = client.delete(f"/api/todos/{todo_id}")

        # Assert
        assert response.status_code == 204

        # Verify deleted
        get_response = client.get(f"/api/todos/{todo_id}")
        assert get_response.status_code == 404

    def test_delete_todo_not_found(self, client: TestClient):
        """GIVEN 존재하지 않는 ID WHEN DELETE 요청 THEN 404 응답"""
        response = client.delete("/api/todos/9999")

        assert response.status_code == 404
```

**핵심 포인트**:
- Given-When-Then 패턴 사용
- 각 HTTP 메서드별로 클래스 분리
- 성공/실패 케이스 모두 작성
- 경계값 테스트 (빈 문자열, 존재하지 않는 ID)

### 4.3 RED 확인 (테스트 실패)

테스트를 실행하면 당연히 실패합니다 (아직 API를 구현하지 않았으므로):

```bash
.venv/bin/pytest tests/ -v
```

출력 (예상):

```
tests/test_todos.py::TestTodoCreate::test_create_todo_success FAILED
... (404 Not Found 에러)
```

이제 GREEN 단계로 넘어갑니다!

---

## Step 5: GREEN - 최소 구현

### 5.1 서비스 레이어 구현

**`app/services/todo_service.py`**:

```python
# @CODE:TODO-001:DOMAIN | SPEC: SPEC-TODO-001.md
"""
Todo business logic service
"""
from typing import List, Optional

from sqlalchemy.orm import Session

from app.models.todo import Todo
from app.schemas.todo import TodoCreate, TodoUpdate


class TodoService:
    """Service layer for Todo CRUD operations"""

    @staticmethod
    def create_todo(db: Session, todo_data: TodoCreate) -> Todo:
        """Create a new todo"""
        todo = Todo(title=todo_data.title, description=todo_data.description)
        db.add(todo)
        db.commit()
        db.refresh(todo)
        return todo

    @staticmethod
    def get_all_todos(db: Session) -> List[Todo]:
        """Get all todos"""
        return db.query(Todo).order_by(Todo.created_at.desc()).all()

    @staticmethod
    def get_todo_by_id(db: Session, todo_id: int) -> Optional[Todo]:
        """Get a todo by ID"""
        return db.query(Todo).filter(Todo.id == todo_id).first()

    @staticmethod
    def update_todo(
        db: Session, todo_id: int, todo_data: TodoUpdate
    ) -> Optional[Todo]:
        """Update a todo"""
        todo = TodoService.get_todo_by_id(db, todo_id)
        if not todo:
            return None

        update_data = todo_data.model_dump(exclude_unset=True)
        for key, value in update_data.items():
            setattr(todo, key, value)

        db.commit()
        db.refresh(todo)
        return todo

    @staticmethod
    def delete_todo(db: Session, todo_id: int) -> bool:
        """Delete a todo. Returns True if deleted, False if not found"""
        todo = TodoService.get_todo_by_id(db, todo_id)
        if not todo:
            return False

        db.delete(todo)
        db.commit()
        return True
```

**핵심 포인트**:
- 비즈니스 로직을 서비스 레이어로 분리
- `@staticmethod`로 상태 없는 서비스 구현
- `model_dump(exclude_unset=True)`로 부분 업데이트 지원

### 5.2 FastAPI 라우터 구현

**`app/routers/todos.py`**:

```python
# @CODE:TODO-001:API | SPEC: SPEC-TODO-001.md
"""
Todo API endpoints
"""
from typing import List

from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session

from app.database import get_db
from app.schemas.todo import TodoCreate, TodoResponse, TodoUpdate
from app.services.todo_service import TodoService

router = APIRouter(prefix="/api/todos", tags=["todos"])


@router.post("/", response_model=TodoResponse, status_code=status.HTTP_201_CREATED)
def create_todo(todo: TodoCreate, db: Session = Depends(get_db)) -> TodoResponse:
    """
    Create a new todo item

    - **title**: Todo title (1-200 characters)
    - **description**: Optional description (max 1000 characters)
    """
    return TodoService.create_todo(db, todo)


@router.get("/", response_model=List[TodoResponse])
def get_all_todos(db: Session = Depends(get_db)) -> List[TodoResponse]:
    """Get all todo items, ordered by creation date (newest first)"""
    return TodoService.get_all_todos(db)


@router.get("/{todo_id}", response_model=TodoResponse)
def get_todo(todo_id: int, db: Session = Depends(get_db)) -> TodoResponse:
    """Get a specific todo item by ID"""
    todo = TodoService.get_todo_by_id(db, todo_id)
    if not todo:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND, detail="Todo not found"
        )
    return todo


@router.patch("/{todo_id}", response_model=TodoResponse)
def update_todo(
    todo_id: int, todo_data: TodoUpdate, db: Session = Depends(get_db)
) -> TodoResponse:
    """
    Update a todo item (partial update)

    - Only provided fields will be updated
    - Unset fields remain unchanged
    """
    todo = TodoService.update_todo(db, todo_id, todo_data)
    if not todo:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND, detail="Todo not found"
        )
    return todo


@router.delete("/{todo_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_todo(todo_id: int, db: Session = Depends(get_db)) -> None:
    """Delete a todo item"""
    success = TodoService.delete_todo(db, todo_id)
    if not success:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND, detail="Todo not found"
        )
```

**핵심 포인트**:
- `Depends(get_db)`로 DB 세션 주입
- `response_model`로 응답 스키마 명시
- `HTTPException`으로 에러 처리
- OpenAPI 문서 자동 생성 (docstring)

### 5.3 FastAPI 애플리케이션 설정

**`app/main.py`**:

```python
# @CODE:TODO-001:API | SPEC: SPEC-TODO-001.md
"""
FastAPI application entry point
"""
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from app.routers import todos

app = FastAPI(
    title="Todo API",
    description="SPEC-First TDD Todo Application",
    version="0.1.0",
)

# CORS 설정 (프론트엔드 연동용)
app.add_middleware(
    CORSMiddleware,
    allow_origins=["http://localhost:5173"],  # Vite 기본 포트
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# 라우터 등록
app.include_router(todos.router)


@app.get("/")
def read_root() -> dict[str, str]:
    """Health check endpoint"""
    return {"status": "healthy", "message": "Todo API is running"}
```

### 5.4 GREEN 확인 (테스트 통과)

```bash
.venv/bin/pytest tests/ -v
```

출력:

```
tests/test_todos.py::TestTodoCreate::test_create_todo_success PASSED
tests/test_todos.py::TestTodoCreate::test_create_todo_missing_title PASSED
tests/test_todos.py::TestTodoCreate::test_create_todo_empty_title PASSED
tests/test_todos.py::TestTodoRead::test_get_all_todos_empty PASSED
tests/test_todos.py::TestTodoRead::test_get_all_todos_with_data PASSED
tests/test_todos.py::TestTodoRead::test_get_todo_by_id_success PASSED
tests/test_todos.py::TestTodoRead::test_get_todo_by_id_not_found PASSED
tests/test_todos.py::TestTodoUpdate::test_update_todo_title PASSED
tests/test_todos.py::TestTodoUpdate::test_update_todo_completed PASSED
tests/test_todos.py::TestTodoUpdate::test_update_todo_not_found PASSED
tests/test_todos.py::TestTodoDelete::test_delete_todo_success PASSED
tests/test_todos.py::TestTodoDelete::test_delete_todo_not_found PASSED

============================== 14 passed in 0.16s ==============================
```

🎉 **모든 테스트 통과!**

---

## Step 6: REFACTOR - 코드 품질 개선

### 6.1 린터 및 포매터 실행

**Ruff로 자동 수정**:

```bash
.venv/bin/ruff check app/ --fix
.venv/bin/ruff format app/
```

출력:

```
Found 16 errors (16 fixed, 0 remaining).
6 files reformatted, 5 files left unchanged
```

### 6.2 커버리지 확인

```bash
.venv/bin/pytest tests/ --cov=app --cov-report=term-missing
```

출력:

```
Name                             Stmts   Miss  Cover   Missing
--------------------------------------------------------------
app/__init__.py                      1      0   100%
app/database.py                     15      0   100%
app/main.py                          9      0   100%
app/models/__init__.py               1      0   100%
app/models/todo.py                   8      0   100%
app/routers/__init__.py              0      0   100%
app/routers/todos.py                30      0   100%
app/schemas/__init__.py              1      0   100%
app/schemas/todo.py                 12      0   100%
app/services/__init__.py             1      0   100%
app/services/todo_service.py        28      1    96%   47
--------------------------------------------------------------
TOTAL                              106      1    95%

============================== 14 passed in 0.16s ==============================
```

✅ **커버리지 95% 달성! (목표 85% 초과)**

### 6.3 타입 검사 (mypy)

나중에 `/alfred:3-sync` 후 추가 개선 시 mypy strict 모드를 적용할 예정입니다.

---

## Step 7: 백엔드 실행 확인

### 7.1 서버 실행

```bash
cd backend
.venv/bin/uvicorn app.main:app --reload
```

출력:

```
INFO:     Uvicorn running on http://127.0.0.1:8000 (Press CTRL+C to quit)
INFO:     Started reloader process [12345] using StatReload
INFO:     Started server process [12346]
INFO:     Waiting for application startup.
INFO:     Application startup complete.
```

### 7.2 API 문서 확인

브라우저에서 접속:

- Swagger UI: `http://localhost:8000/docs`
- ReDoc: `http://localhost:8000/redoc`

### 7.3 수동 테스트 (curl)

**Todo 생성**:

```bash
curl -X POST http://localhost:8000/api/todos \
  -H "Content-Type: application/json" \
  -d '{"title": "API 테스트", "description": "curl로 생성"}'
```

응답:

```json
{
  "id": 1,
  "title": "API 테스트",
  "description": "curl로 생성",
  "completed": false,
  "created_at": "2025-10-11T10:30:00",
  "updated_at": "2025-10-11T10:30:00"
}
```

**Todo 목록 조회**:

```bash
curl http://localhost:8000/api/todos
```

---

## 핵심 학습 포인트

### ✅ TDD 사이클 완벽 적용

1. **RED**: 14개 테스트 작성 → 실패 확인
1. **GREEN**: API 구현 → 모든 테스트 통과
1. **REFACTOR**: 린터/포매터 실행 → 95% 커버리지

### ✅ 최신 Python 기술 스택

- SQLAlchemy 2.0: `Mapped`, `mapped_column`
- Pydantic v2: `ConfigDict`, `model_dump()`
- FastAPI: Dependency Injection, 자동 OpenAPI 문서
- pytest: fixture, TestClient, 인메모리 DB

### ✅ 코드 품질 도구

- ruff: 빠른 린터 + 포매터
- pytest-cov: 95% 커버리지
- mypy: strict 타입 검사 (후속 개선)

### ✅ @TAG 추적성

- `@CODE:TODO-001:DATA` - 모델/스키마
- `@CODE:TODO-001:DOMAIN` - 서비스 레이어
- `@CODE:TODO-001:API` - FastAPI 라우터
- `@TEST:TODO-001` - pytest 테스트

---

## 🚀 다음

백엔드 TDD 구현이 완료되었습니다! 이제 프론트엔드로 넘어갑니다.

**다음**: [Part 4: Frontend 구현하기](./04-frontend-impl.md)

**이전**: [Part 2: SPEC 작성하기](./02-spec-writing.md)
