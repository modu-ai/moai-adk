---
title: "기본 CRUD 작업"
category: "rest-api"
difficulty: "초급"
tags: [fastapi, crud, sqlalchemy, rest-api, python]
---

# 기본 CRUD 작업

## 개요

FastAPI와 SQLAlchemy를 사용하여 사용자(User) 리소스에 대한 CRUD (Create, Read, Update, Delete) 작업을 구현합니다. REST API의 가장 기본이 되는 패턴입니다.

## 사용 사례

- 사용자 관리 시스템
- 블로그 게시글 관리
- 제품 카탈로그 관리
- 모든 리소스 기반 API

## 완전한 코드 예제

### 1. 데이터베이스 설정 (`database.py`)

```python
# SPEC: API-001 - 데이터베이스 연결 설정
# @TAG:API-001

from sqlalchemy import create_engine
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker

# SQLite 데이터베이스 (개발용)
SQLALCHEMY_DATABASE_URL = "sqlite:///./test.db"
# PostgreSQL 예시 (프로덕션)
# SQLALCHEMY_DATABASE_URL = "postgresql://user:password@localhost/dbname"

# 데이터베이스 엔진 생성
engine = create_engine(
    SQLALCHEMY_DATABASE_URL,
    connect_args={"check_same_thread": False}  # SQLite 전용
)

# 세션 팩토리 생성
SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)

# Base 클래스 생성
Base = declarative_base()

def get_db():
    """
    데이터베이스 세션 의존성

    Yields:
        Session: SQLAlchemy 데이터베이스 세션
    """
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()
```

### 2. SQLAlchemy 모델 (`models.py`)

```python
# SPEC: API-002 - 사용자 데이터 모델 정의
# @TAG:API-002

from sqlalchemy import Column, Integer, String, Boolean, DateTime
from sqlalchemy.sql import func
from database import Base

class User(Base):
    """
    사용자 모델

    Attributes:
        id: 기본 키
        email: 이메일 (고유값)
        name: 이름
        is_active: 활성화 상태
        created_at: 생성 시간
        updated_at: 수정 시간
    """
    __tablename__ = "users"

    id = Column(Integer, primary_key=True, index=True)
    email = Column(String, unique=True, index=True, nullable=False)
    name = Column(String, nullable=False)
    is_active = Column(Boolean, default=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), onupdate=func.now())

    def __repr__(self):
        return f"<User(id={self.id}, email={self.email}, name={self.name})>"
```

### 3. Pydantic 스키마 (`schemas.py`)

```python
# SPEC: API-003 - 요청/응답 스키마 정의
# @TAG:API-003

from pydantic import BaseModel, EmailStr, Field
from datetime import datetime
from typing import Optional

class UserBase(BaseModel):
    """사용자 기본 스키마"""
    email: EmailStr = Field(..., description="사용자 이메일")
    name: str = Field(..., min_length=1, max_length=100, description="사용자 이름")

class UserCreate(UserBase):
    """사용자 생성 요청 스키마"""
    pass

class UserUpdate(BaseModel):
    """사용자 수정 요청 스키마"""
    email: Optional[EmailStr] = Field(None, description="수정할 이메일")
    name: Optional[str] = Field(None, min_length=1, max_length=100, description="수정할 이름")
    is_active: Optional[bool] = Field(None, description="활성화 상태")

class UserResponse(UserBase):
    """사용자 응답 스키마"""
    id: int
    is_active: bool
    created_at: datetime
    updated_at: Optional[datetime]

    class Config:
        from_attributes = True  # Pydantic v2
        # orm_mode = True  # Pydantic v1
```

### 4. CRUD 함수 (`crud.py`)

```python
# SPEC: API-004 - CRUD 비즈니스 로직
# @TAG:API-004

from sqlalchemy.orm import Session
from typing import List, Optional
import models
import schemas

def get_user(db: Session, user_id: int) -> Optional[models.User]:
    """
    ID로 사용자 조회

    Args:
        db: 데이터베이스 세션
        user_id: 사용자 ID

    Returns:
        User 객체 또는 None
    """
    return db.query(models.User).filter(models.User.id == user_id).first()

def get_user_by_email(db: Session, email: str) -> Optional[models.User]:
    """
    이메일로 사용자 조회

    Args:
        db: 데이터베이스 세션
        email: 사용자 이메일

    Returns:
        User 객체 또는 None
    """
    return db.query(models.User).filter(models.User.email == email).first()

def get_users(db: Session, skip: int = 0, limit: int = 100) -> List[models.User]:
    """
    사용자 목록 조회

    Args:
        db: 데이터베이스 세션
        skip: 건너뛸 레코드 수
        limit: 조회할 최대 레코드 수

    Returns:
        User 객체 리스트
    """
    return db.query(models.User).offset(skip).limit(limit).all()

def create_user(db: Session, user: schemas.UserCreate) -> models.User:
    """
    사용자 생성

    Args:
        db: 데이터베이스 세션
        user: 사용자 생성 데이터

    Returns:
        생성된 User 객체
    """
    db_user = models.User(**user.model_dump())
    db.add(db_user)
    db.commit()
    db.refresh(db_user)
    return db_user

def update_user(
    db: Session,
    user_id: int,
    user_update: schemas.UserUpdate
) -> Optional[models.User]:
    """
    사용자 정보 수정

    Args:
        db: 데이터베이스 세션
        user_id: 사용자 ID
        user_update: 수정할 데이터

    Returns:
        수정된 User 객체 또는 None
    """
    db_user = get_user(db, user_id)
    if not db_user:
        return None

    # None이 아닌 값만 업데이트
    update_data = user_update.model_dump(exclude_unset=True)
    for field, value in update_data.items():
        setattr(db_user, field, value)

    db.commit()
    db.refresh(db_user)
    return db_user

def delete_user(db: Session, user_id: int) -> bool:
    """
    사용자 삭제

    Args:
        db: 데이터베이스 세션
        user_id: 사용자 ID

    Returns:
        삭제 성공 여부
    """
    db_user = get_user(db, user_id)
    if not db_user:
        return False

    db.delete(db_user)
    db.commit()
    return True
```

### 5. FastAPI 라우터 (`main.py`)

```python
# SPEC: API-005 - REST API 엔드포인트
# @TAG:API-005

from fastapi import FastAPI, Depends, HTTPException, status
from sqlalchemy.orm import Session
from typing import List

import models
import schemas
import crud
from database import engine, get_db

# 데이터베이스 테이블 생성
models.Base.metadata.create_all(bind=engine)

# FastAPI 앱 생성
app = FastAPI(
    title="User CRUD API",
    description="사용자 관리 REST API",
    version="1.0.0"
)

@app.post("/users/", response_model=schemas.UserResponse, status_code=status.HTTP_201_CREATED)
def create_user(user: schemas.UserCreate, db: Session = Depends(get_db)):
    """
    새 사용자 생성

    Args:
        user: 사용자 생성 데이터
        db: 데이터베이스 세션

    Returns:
        생성된 사용자 정보

    Raises:
        HTTPException: 이메일이 이미 존재하는 경우
    """
    # 이메일 중복 체크
    db_user = crud.get_user_by_email(db, email=user.email)
    if db_user:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="이미 등록된 이메일입니다"
        )

    return crud.create_user(db=db, user=user)

@app.get("/users/", response_model=List[schemas.UserResponse])
def read_users(skip: int = 0, limit: int = 100, db: Session = Depends(get_db)):
    """
    사용자 목록 조회

    Args:
        skip: 건너뛸 레코드 수 (기본값: 0)
        limit: 조회할 최대 레코드 수 (기본값: 100)
        db: 데이터베이스 세션

    Returns:
        사용자 목록
    """
    users = crud.get_users(db, skip=skip, limit=limit)
    return users

@app.get("/users/{user_id}", response_model=schemas.UserResponse)
def read_user(user_id: int, db: Session = Depends(get_db)):
    """
    특정 사용자 조회

    Args:
        user_id: 사용자 ID
        db: 데이터베이스 세션

    Returns:
        사용자 정보

    Raises:
        HTTPException: 사용자를 찾을 수 없는 경우
    """
    db_user = crud.get_user(db, user_id=user_id)
    if db_user is None:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="사용자를 찾을 수 없습니다"
        )
    return db_user

@app.put("/users/{user_id}", response_model=schemas.UserResponse)
def update_user(
    user_id: int,
    user_update: schemas.UserUpdate,
    db: Session = Depends(get_db)
):
    """
    사용자 정보 수정

    Args:
        user_id: 사용자 ID
        user_update: 수정할 데이터
        db: 데이터베이스 세션

    Returns:
        수정된 사용자 정보

    Raises:
        HTTPException: 사용자를 찾을 수 없는 경우
    """
    db_user = crud.update_user(db, user_id=user_id, user_update=user_update)
    if db_user is None:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="사용자를 찾을 수 없습니다"
        )
    return db_user

@app.delete("/users/{user_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_user(user_id: int, db: Session = Depends(get_db)):
    """
    사용자 삭제

    Args:
        user_id: 사용자 ID
        db: 데이터베이스 세션

    Raises:
        HTTPException: 사용자를 찾을 수 없는 경우
    """
    success = crud.delete_user(db, user_id=user_id)
    if not success:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="사용자를 찾을 수 없습니다"
        )
    return None

@app.get("/")
def root():
    """루트 엔드포인트"""
    return {
        "message": "User CRUD API",
        "docs": "/docs",
        "version": "1.0.0"
    }
```

## 단계별 설명

### 1. 프로젝트 설정

```bash
# 가상환경 생성
uv venv

# 의존성 설치
uv pip install fastapi sqlalchemy pydantic uvicorn[standard]
```

### 2. 디렉토리 구조

```
crud-api/
├── database.py      # DB 연결 설정
├── models.py        # SQLAlchemy 모델
├── schemas.py       # Pydantic 스키마
├── crud.py          # CRUD 로직
├── main.py          # FastAPI 앱
└── test.db          # SQLite DB (자동 생성)
```

### 3. 서버 실행

```bash
# 개발 서버 시작
uvicorn main:app --reload

# 다른 포트로 실행
uvicorn main:app --reload --port 8080
```

### 4. API 테스트

```bash
# 사용자 생성
curl -X POST "http://localhost:8000/users/" \
  -H "Content-Type: application/json" \
  -d '{"email":"john@example.com","name":"John Doe"}'

# 사용자 목록 조회
curl "http://localhost:8000/users/"

# 특정 사용자 조회
curl "http://localhost:8000/users/1"

# 사용자 수정
curl -X PUT "http://localhost:8000/users/1" \
  -H "Content-Type: application/json" \
  -d '{"name":"Jane Doe"}'

# 사용자 삭제
curl -X DELETE "http://localhost:8000/users/1"
```

## 테스트 코드

```python
# SPEC: TEST-API-001 - CRUD API 테스트
# @TAG:TEST-API-001

import pytest
from fastapi.testclient import TestClient
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker

from database import Base, get_db
from main import app
import models

# 테스트용 인메모리 데이터베이스
SQLALCHEMY_DATABASE_URL = "sqlite:///./test_db.db"
engine = create_engine(SQLALCHEMY_DATABASE_URL, connect_args={"check_same_thread": False})
TestingSessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)

# 테스트 DB 설정
Base.metadata.create_all(bind=engine)

def override_get_db():
    """테스트용 DB 세션"""
    try:
        db = TestingSessionLocal()
        yield db
    finally:
        db.close()

app.dependency_overrides[get_db] = override_get_db
client = TestClient(app)

@pytest.fixture(autouse=True)
def setup_database():
    """각 테스트 전후로 DB 초기화"""
    Base.metadata.create_all(bind=engine)
    yield
    Base.metadata.drop_all(bind=engine)

def test_create_user():
    """사용자 생성 테스트"""
    response = client.post(
        "/users/",
        json={"email": "test@example.com", "name": "Test User"}
    )
    assert response.status_code == 201
    data = response.json()
    assert data["email"] == "test@example.com"
    assert data["name"] == "Test User"
    assert "id" in data

def test_create_user_duplicate_email():
    """중복 이메일 생성 테스트"""
    # 첫 번째 사용자 생성
    client.post("/users/", json={"email": "test@example.com", "name": "User 1"})

    # 같은 이메일로 두 번째 시도
    response = client.post(
        "/users/",
        json={"email": "test@example.com", "name": "User 2"}
    )
    assert response.status_code == 400
    assert "이미 등록된 이메일" in response.json()["detail"]

def test_read_users():
    """사용자 목록 조회 테스트"""
    # 사용자 2명 생성
    client.post("/users/", json={"email": "user1@example.com", "name": "User 1"})
    client.post("/users/", json={"email": "user2@example.com", "name": "User 2"})

    # 목록 조회
    response = client.get("/users/")
    assert response.status_code == 200
    data = response.json()
    assert len(data) == 2

def test_read_user():
    """특정 사용자 조회 테스트"""
    # 사용자 생성
    create_response = client.post(
        "/users/",
        json={"email": "test@example.com", "name": "Test User"}
    )
    user_id = create_response.json()["id"]

    # 조회
    response = client.get(f"/users/{user_id}")
    assert response.status_code == 200
    data = response.json()
    assert data["id"] == user_id
    assert data["email"] == "test@example.com"

def test_read_user_not_found():
    """존재하지 않는 사용자 조회 테스트"""
    response = client.get("/users/999")
    assert response.status_code == 404
    assert "찾을 수 없습니다" in response.json()["detail"]

def test_update_user():
    """사용자 수정 테스트"""
    # 사용자 생성
    create_response = client.post(
        "/users/",
        json={"email": "test@example.com", "name": "Old Name"}
    )
    user_id = create_response.json()["id"]

    # 수정
    response = client.put(
        f"/users/{user_id}",
        json={"name": "New Name"}
    )
    assert response.status_code == 200
    data = response.json()
    assert data["name"] == "New Name"
    assert data["email"] == "test@example.com"  # 이메일은 그대로

def test_delete_user():
    """사용자 삭제 테스트"""
    # 사용자 생성
    create_response = client.post(
        "/users/",
        json={"email": "test@example.com", "name": "Test User"}
    )
    user_id = create_response.json()["id"]

    # 삭제
    response = client.delete(f"/users/{user_id}")
    assert response.status_code == 204

    # 삭제 확인
    get_response = client.get(f"/users/{user_id}")
    assert get_response.status_code == 404

def test_delete_user_not_found():
    """존재하지 않는 사용자 삭제 테스트"""
    response = client.delete("/users/999")
    assert response.status_code == 404
```

## Best Practices

### API 설계
- ✅ **RESTful 규칙 준수**: `/users/` (컬렉션), `/users/{id}` (개별 리소스)
- ✅ **적절한 HTTP 메서드**: GET (조회), POST (생성), PUT (수정), DELETE (삭제)
- ✅ **표준 상태 코드**: 200 (성공), 201 (생성), 204 (삭제 성공), 404 (없음), 400 (잘못된 요청)
- ✅ **명확한 응답 구조**: Pydantic 스키마로 응답 형식 정의
- ✅ **API 문서 자동 생성**: FastAPI의 `/docs` (Swagger UI), `/redoc`

### 데이터베이스
- ✅ **의존성 주입**: `Depends(get_db)`로 세션 관리
- ✅ **트랜잭션 관리**: `commit()` 후 `refresh()` 호출
- ✅ **리소스 정리**: `finally` 블록에서 `db.close()`
- ✅ **인덱스 사용**: 자주 조회하는 필드에 `index=True`
- ✅ **제약 조건**: `unique=True`, `nullable=False` 등

### 코드 품질
- ✅ **Type Hints**: 모든 함수에 타입 명시
- ✅ **Docstring**: 한국어로 명확한 설명
- ✅ **에러 처리**: 명확한 에러 메시지
- ✅ **검증**: Pydantic으로 입력 검증
- ✅ **테스트**: 모든 엔드포인트 테스트 작성

## 주의사항

### 보안
- ❌ **비밀번호 평문 저장 금지**: 해시화 필요 (예: bcrypt)
- ❌ **SQL 인젝션 주의**: SQLAlchemy ORM 사용 (Raw query 지양)
- ❌ **민감한 정보 로깅 금지**: 비밀번호, 토큰 등

### 성능
- ❌ **N+1 쿼리 문제**: `joinedload()` 사용
- ❌ **무제한 조회 금지**: `limit` 파라미터 필수
- ❌ **대량 삭제 주의**: 배치 처리 고려

### 데이터 무결성
- ❌ **중복 데이터 허용 금지**: Unique 제약 조건
- ❌ **NULL 값 주의**: `nullable=False` 설정
- ❌ **트랜잭션 누락 금지**: CRUD 함수에서 commit 필수

## 관련 예제

- [페이지네이션 & 정렬](/ko/examples/rest-api/pagination) - 대용량 데이터 처리
- [필터링 & 검색](/ko/examples/rest-api/filtering) - 동적 쿼리
- [에러 처리 & 검증](/ko/examples/rest-api/error-handling) - 안전한 API
- [SQLAlchemy 관계](/ko/examples/database/relationships) - 테이블 관계
- [Pytest 단위 테스트](/ko/examples/testing/unit-tests) - TDD 개발

## 다음 단계

1. [에러 처리 & 검증](/ko/examples/rest-api/error-handling) - 더 안전한 API 만들기
2. [JWT 인증](/ko/examples/authentication/jwt-basic) - API 보호하기
3. [Tutorial 01: FastAPI 프로젝트](/ko/tutorials/tutorial-01-fastapi) - 전체 프로젝트 구축
