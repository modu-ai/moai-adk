---
title: "Tutorial 1: 첫 REST API 개발"
description: "FastAPI로 REST API를 30분만에 만들어봅니다"
duration: "30분"
difficulty: "초급"
tags: [tutorial, rest-api, fastapi, beginner]
---

# Tutorial 1: 첫 REST API 개발

이 튜토리얼에서는 MoAI-ADK와 FastAPI를 사용하여 사용자 관리 REST API를 처음부터 만들어봅니다. SPEC-first TDD 워크플로우를 통해 체계적으로 API를 구축하는 방법을 배웁니다.

## 🎯 학습 목표

이 튜토리얼을 완료하면 다음을 할 수 있습니다:

- ✅ MoAI-ADK의 `/alfred:1-plan`으로 SPEC 작성하기
- ✅ `/alfred:2-run`으로 TDD 기반 구현하기
- ✅ FastAPI로 CRUD REST API 구축하기
- ✅ Pydantic으로 데이터 검증하기
- ✅ Swagger UI로 자동 문서화 활용하기
- ✅ Pytest로 API 테스트 작성하기

## 📋 사전 요구사항

### 필수 설치

- **Python 3.11+**: `python --version`으로 확인
- **MoAI-ADK v0.23.0+**: `moai-adk --version`으로 확인
- **Git**: 버전 관리를 위해 필요
- **IDE**: VS Code, PyCharm 등 (추천: VS Code + Python extension)

### 선행 지식

- Python 기본 문법 (함수, 클래스, 데코레이터)
- HTTP 프로토콜 기초 (GET, POST, PUT, DELETE)
- JSON 데이터 포맷
- 기본적인 터미널 사용법

### 설치 확인

```bash
# Python 버전 확인
python --version  # Python 3.11.0 이상이어야 함

# MoAI-ADK 설치 확인
moai-adk --version  # v0.23.0 이상

# 프로젝트 디렉토리 생성
mkdir user-api-tutorial
cd user-api-tutorial

# MoAI-ADK 프로젝트 초기화
moai-adk init
```

## 🚀 프로젝트 구조

완성된 프로젝트는 다음과 같은 구조를 가집니다:

```
user-api-tutorial/
├── .moai/
│   ├── config.json
│   └── specs/
│       └── SPEC-USER-API-001.md
├── src/
│   └── user_api/
│       ├── __init__.py
│       ├── main.py           # FastAPI 앱
│       ├── models.py         # Pydantic 모델
│       ├── database.py       # 인메모리 데이터베이스
│       └── routes.py         # API 라우트
├── tests/
│   ├── __init__.py
│   └── test_user_api.py      # API 테스트
├── requirements.txt
└── README.md
```

## 단계별 실습

### Step 1: SPEC 작성으로 시작하기

MoAI-ADK의 핵심은 **SPEC-first** 접근입니다. 코드를 작성하기 전에 무엇을 만들지 명확히 정의합니다.

```bash
# Alfred에게 계획 요청
/alfred:1-plan "사용자 관리 REST API 만들기"
```

Alfred가 대화형으로 요구사항을 물어봅니다:

**Alfred의 질문**:
```
어떤 기능이 필요한가요?
1. 기본 CRUD만
2. 검색 기능 포함
3. 페이지네이션 포함
```

**선택**: "1. 기본 CRUD만"

Alfred가 자동으로 SPEC 문서를 생성합니다:

**생성된 SPEC** (`.moai/specs/SPEC-USER-API-001.md`):

```markdown
# SPEC-USER-API-001: 사용자 관리 REST API

## 요구사항

사용자 정보를 관리하는 REST API를 구현한다.

### 기능 요구사항

- FR-001: 모든 사용자 목록 조회 (GET /users)
- FR-002: 특정 사용자 조회 (GET /users/{id})
- FR-003: 새 사용자 생성 (POST /users)
- FR-004: 사용자 정보 수정 (PUT /users/{id})
- FR-005: 사용자 삭제 (DELETE /users/{id})

### 데이터 모델

User:
- id: int (자동 생성)
- name: str (필수, 2-50자)
- email: str (필수, 이메일 형식)
- age: int (선택, 0-150)
- created_at: datetime (자동 생성)

### API 엔드포인트

| Method | Endpoint | Description | Request | Response |
|--------|----------|-------------|---------|----------|
| GET | /users | 모든 사용자 조회 | - | User[] |
| GET | /users/{id} | 사용자 조회 | - | User |
| POST | /users | 사용자 생성 | UserCreate | User |
| PUT | /users/{id} | 사용자 수정 | UserUpdate | User |
| DELETE | /users/{id} | 사용자 삭제 | - | 204 |

### 검증 규칙

- 이메일은 유효한 형식이어야 함
- 이름은 2-50자 사이여야 함
- 나이는 0-150 사이여야 함 (제공된 경우)
- 중복 이메일 불가

### 에러 처리

- 404: 사용자를 찾을 수 없음
- 400: 잘못된 요청 데이터
- 409: 이메일 중복
```

### Step 2: 환경 설정

필요한 패키지를 설치합니다.

**requirements.txt**:
```txt
fastapi==0.104.1
uvicorn[standard]==0.24.0
pydantic==2.5.0
pytest==7.4.3
httpx==0.25.2
```

설치:
```bash
pip install -r requirements.txt
```

### Step 3: 데이터 모델 정의

**src/user_api/models.py**:

```python
"""
사용자 데이터 모델 정의
"""
from datetime import datetime
from typing import Optional
from pydantic import BaseModel, EmailStr, Field


class UserBase(BaseModel):
    """사용자 기본 모델"""
    name: str = Field(..., min_length=2, max_length=50, description="사용자 이름")
    email: EmailStr = Field(..., description="이메일 주소")
    age: Optional[int] = Field(None, ge=0, le=150, description="나이")


class UserCreate(UserBase):
    """사용자 생성 요청 모델"""
    pass


class UserUpdate(UserBase):
    """사용자 수정 요청 모델"""
    name: Optional[str] = Field(None, min_length=2, max_length=50)
    email: Optional[EmailStr] = None
    age: Optional[int] = Field(None, ge=0, le=150)


class User(UserBase):
    """사용자 응답 모델"""
    id: int = Field(..., description="사용자 ID")
    created_at: datetime = Field(..., description="생성 시각")

    class Config:
        from_attributes = True


class UserListResponse(BaseModel):
    """사용자 목록 응답"""
    users: list[User]
    total: int
```

**포인트**:
- `Pydantic`의 `Field`로 상세한 검증 규칙 정의
- `EmailStr`로 이메일 형식 자동 검증
- `Optional`로 선택적 필드 표현
- 요청/응답 모델 분리 (보안 및 명확성)

### Step 4: 인메모리 데이터베이스 구현

**src/user_api/database.py**:

```python
"""
인메모리 사용자 데이터베이스
"""
from datetime import datetime
from typing import Optional
from .models import User, UserCreate, UserUpdate


class UserDatabase:
    """사용자 데이터를 관리하는 인메모리 데이터베이스"""

    def __init__(self):
        self._users: dict[int, dict] = {}
        self._next_id: int = 1

    def create_user(self, user_data: UserCreate) -> User:
        """새 사용자 생성"""
        # 이메일 중복 확인
        if any(u["email"] == user_data.email for u in self._users.values()):
            raise ValueError("Email already exists")

        user_dict = {
            "id": self._next_id,
            "name": user_data.name,
            "email": user_data.email,
            "age": user_data.age,
            "created_at": datetime.now(),
        }

        self._users[self._next_id] = user_dict
        self._next_id += 1

        return User(**user_dict)

    def get_user(self, user_id: int) -> Optional[User]:
        """특정 사용자 조회"""
        user_dict = self._users.get(user_id)
        if user_dict:
            return User(**user_dict)
        return None

    def get_all_users(self) -> list[User]:
        """모든 사용자 조회"""
        return [User(**u) for u in self._users.values()]

    def update_user(self, user_id: int, user_data: UserUpdate) -> Optional[User]:
        """사용자 정보 수정"""
        if user_id not in self._users:
            return None

        user_dict = self._users[user_id]

        # 제공된 필드만 업데이트
        update_data = user_data.model_dump(exclude_unset=True)

        # 이메일 변경 시 중복 확인
        if "email" in update_data:
            if any(
                u["email"] == update_data["email"] and uid != user_id
                for uid, u in self._users.items()
            ):
                raise ValueError("Email already exists")

        user_dict.update(update_data)
        return User(**user_dict)

    def delete_user(self, user_id: int) -> bool:
        """사용자 삭제"""
        if user_id in self._users:
            del self._users[user_id]
            return True
        return False

    def clear(self):
        """모든 데이터 삭제 (테스트용)"""
        self._users.clear()
        self._next_id = 1


# 전역 데이터베이스 인스턴스
db = UserDatabase()
```

**포인트**:
- 단순한 딕셔너리 기반 저장소 (실무에서는 PostgreSQL, MongoDB 등 사용)
- 이메일 중복 검증 로직
- `model_dump(exclude_unset=True)`로 제공된 필드만 업데이트

### Step 5: API 라우트 구현

**src/user_api/routes.py**:

```python
"""
사용자 API 라우트
"""
from fastapi import APIRouter, HTTPException, status
from .models import User, UserCreate, UserUpdate, UserListResponse
from .database import db

router = APIRouter(prefix="/users", tags=["users"])


@router.post("/", response_model=User, status_code=status.HTTP_201_CREATED)
def create_user(user: UserCreate):
    """
    새 사용자 생성

    - **name**: 사용자 이름 (2-50자)
    - **email**: 이메일 주소 (유효한 형식)
    - **age**: 나이 (선택, 0-150)
    """
    try:
        return db.create_user(user)
    except ValueError as e:
        raise HTTPException(
            status_code=status.HTTP_409_CONFLICT,
            detail=str(e)
        )


@router.get("/", response_model=UserListResponse)
def get_users():
    """모든 사용자 조회"""
    users = db.get_all_users()
    return UserListResponse(users=users, total=len(users))


@router.get("/{user_id}", response_model=User)
def get_user(user_id: int):
    """특정 사용자 조회"""
    user = db.get_user(user_id)
    if not user:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"User {user_id} not found"
        )
    return user


@router.put("/{user_id}", response_model=User)
def update_user(user_id: int, user_data: UserUpdate):
    """사용자 정보 수정"""
    try:
        user = db.update_user(user_id, user_data)
        if not user:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail=f"User {user_id} not found"
            )
        return user
    except ValueError as e:
        raise HTTPException(
            status_code=status.HTTP_409_CONFLICT,
            detail=str(e)
        )


@router.delete("/{user_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_user(user_id: int):
    """사용자 삭제"""
    if not db.delete_user(user_id):
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"User {user_id} not found"
        )
```

**포인트**:
- `APIRouter`로 라우트 그룹화
- `response_model`로 응답 타입 명시
- `status_code`로 HTTP 상태 코드 지정
- 명확한 에러 처리 (404, 409)

### Step 6: FastAPI 앱 생성

**src/user_api/main.py**:

```python
"""
User API FastAPI 애플리케이션
"""
from fastapi import FastAPI
from .routes import router

app = FastAPI(
    title="User Management API",
    description="간단한 사용자 관리 REST API",
    version="1.0.0",
    docs_url="/docs",
    redoc_url="/redoc",
)

# 라우트 등록
app.include_router(router)


@app.get("/")
def root():
    """API 루트"""
    return {
        "message": "User Management API",
        "docs": "/docs",
        "version": "1.0.0"
    }


@app.get("/health")
def health_check():
    """헬스 체크 엔드포인트"""
    return {"status": "healthy"}
```

### Step 7: 테스트 작성 (TDD)

**tests/test_user_api.py**:

```python
"""
User API 테스트
"""
import pytest
from fastapi.testclient import TestClient
from src.user_api.main import app
from src.user_api.database import db


@pytest.fixture(autouse=True)
def reset_database():
    """각 테스트 전에 데이터베이스 초기화"""
    db.clear()
    yield
    db.clear()


client = TestClient(app)


def test_root_endpoint():
    """루트 엔드포인트 테스트"""
    response = client.get("/")
    assert response.status_code == 200
    assert response.json()["message"] == "User Management API"


def test_create_user():
    """사용자 생성 테스트"""
    user_data = {
        "name": "Alice",
        "email": "alice@example.com",
        "age": 30
    }
    response = client.post("/users/", json=user_data)

    assert response.status_code == 201
    data = response.json()
    assert data["name"] == "Alice"
    assert data["email"] == "alice@example.com"
    assert data["age"] == 30
    assert "id" in data
    assert "created_at" in data


def test_create_user_invalid_email():
    """잘못된 이메일 형식 테스트"""
    user_data = {
        "name": "Bob",
        "email": "invalid-email",
        "age": 25
    }
    response = client.post("/users/", json=user_data)
    assert response.status_code == 422  # Validation error


def test_create_user_duplicate_email():
    """이메일 중복 테스트"""
    user_data = {
        "name": "Alice",
        "email": "alice@example.com"
    }
    client.post("/users/", json=user_data)

    # 같은 이메일로 다시 생성 시도
    response = client.post("/users/", json=user_data)
    assert response.status_code == 409
    assert "already exists" in response.json()["detail"]


def test_get_users():
    """사용자 목록 조회 테스트"""
    # 사용자 2명 생성
    client.post("/users/", json={"name": "Alice", "email": "alice@example.com"})
    client.post("/users/", json={"name": "Bob", "email": "bob@example.com"})

    response = client.get("/users/")
    assert response.status_code == 200

    data = response.json()
    assert data["total"] == 2
    assert len(data["users"]) == 2


def test_get_user():
    """특정 사용자 조회 테스트"""
    # 사용자 생성
    create_response = client.post("/users/", json={
        "name": "Alice",
        "email": "alice@example.com"
    })
    user_id = create_response.json()["id"]

    # 조회
    response = client.get(f"/users/{user_id}")
    assert response.status_code == 200
    assert response.json()["name"] == "Alice"


def test_get_user_not_found():
    """존재하지 않는 사용자 조회 테스트"""
    response = client.get("/users/9999")
    assert response.status_code == 404
    assert "not found" in response.json()["detail"]


def test_update_user():
    """사용자 정보 수정 테스트"""
    # 사용자 생성
    create_response = client.post("/users/", json={
        "name": "Alice",
        "email": "alice@example.com",
        "age": 30
    })
    user_id = create_response.json()["id"]

    # 수정
    update_data = {"name": "Alice Updated", "age": 31}
    response = client.put(f"/users/{user_id}", json=update_data)

    assert response.status_code == 200
    data = response.json()
    assert data["name"] == "Alice Updated"
    assert data["age"] == 31
    assert data["email"] == "alice@example.com"  # 변경 안 됨


def test_delete_user():
    """사용자 삭제 테스트"""
    # 사용자 생성
    create_response = client.post("/users/", json={
        "name": "Alice",
        "email": "alice@example.com"
    })
    user_id = create_response.json()["id"]

    # 삭제
    response = client.delete(f"/users/{user_id}")
    assert response.status_code == 204

    # 조회 시 404
    get_response = client.get(f"/users/{user_id}")
    assert get_response.status_code == 404
```

### Step 8: 애플리케이션 실행

```bash
# 개발 서버 실행
uvicorn src.user_api.main:app --reload --port 8000
```

출력:
```
INFO:     Uvicorn running on http://127.0.0.1:8000 (Press CTRL+C to quit)
INFO:     Started reloader process [12345] using StatReload
INFO:     Started server process [12346]
INFO:     Waiting for application startup.
INFO:     Application startup complete.
```

### Step 9: API 테스트 (수동)

브라우저에서 Swagger UI 열기:
```
http://localhost:8000/docs
```

**cURL로 테스트**:

```bash
# 1. 사용자 생성
curl -X POST "http://localhost:8000/users/" \
  -H "Content-Type: application/json" \
  -d '{"name": "Alice", "email": "alice@example.com", "age": 30}'

# 응답:
# {
#   "id": 1,
#   "name": "Alice",
#   "email": "alice@example.com",
#   "age": 30,
#   "created_at": "2024-01-15T10:30:00.123456"
# }

# 2. 모든 사용자 조회
curl "http://localhost:8000/users/"

# 3. 특정 사용자 조회
curl "http://localhost:8000/users/1"

# 4. 사용자 수정
curl -X PUT "http://localhost:8000/users/1" \
  -H "Content-Type: application/json" \
  -d '{"name": "Alice Updated", "age": 31}'

# 5. 사용자 삭제
curl -X DELETE "http://localhost:8000/users/1"
```

## ✅ 테스트 및 검증

### 자동화된 테스트 실행

```bash
# 모든 테스트 실행
pytest tests/ -v

# 커버리지와 함께 실행
pytest tests/ --cov=src --cov-report=html

# 특정 테스트만 실행
pytest tests/test_user_api.py::test_create_user -v
```

**예상 출력**:
```
tests/test_user_api.py::test_root_endpoint PASSED                   [ 10%]
tests/test_user_api.py::test_create_user PASSED                     [ 20%]
tests/test_user_api.py::test_create_user_invalid_email PASSED       [ 30%]
tests/test_user_api.py::test_create_user_duplicate_email PASSED     [ 40%]
tests/test_user_api.py::test_get_users PASSED                       [ 50%]
tests/test_user_api.py::test_get_user PASSED                        [ 60%]
tests/test_user_api.py::test_get_user_not_found PASSED              [ 70%]
tests/test_user_api.py::test_update_user PASSED                     [ 80%]
tests/test_user_api.py::test_delete_user PASSED                     [ 90%]

============================== 9 passed in 0.45s ==============================
```

### Swagger UI로 검증

1. 브라우저에서 `http://localhost:8000/docs` 열기
2. "POST /users/" 엔드포인트 클릭
3. "Try it out" 클릭
4. Request body 입력:
   ```json
   {
     "name": "Test User",
     "email": "test@example.com",
     "age": 25
   }
   ```
5. "Execute" 클릭
6. Response 확인 (Status 201)

## 🔧 문제 해결

### 문제 1: ModuleNotFoundError

**증상**:
```
ModuleNotFoundError: No module named 'fastapi'
```

**원인**: 패키지가 설치되지 않음

**해결**:
```bash
pip install -r requirements.txt
```

### 문제 2: Email validation error

**증상**:
```
pydantic.error_wrappers.ValidationError: email
```

**원인**: `email-validator` 패키지 미설치

**해결**:
```bash
pip install email-validator
```

### 문제 3: Port already in use

**증상**:
```
ERROR: [Errno 48] Address already in use
```

**원인**: 8000 포트가 이미 사용 중

**해결**:
```bash
# 다른 포트 사용
uvicorn src.user_api.main:app --reload --port 8001

# 또는 기존 프로세스 종료
lsof -ti:8000 | xargs kill -9
```

### 문제 4: CORS 에러 (브라우저에서 호출 시)

**증상**:
```
Access to fetch at 'http://localhost:8000' from origin 'http://localhost:3000'
has been blocked by CORS policy
```

**해결**: CORS 미들웨어 추가

**src/user_api/main.py**:
```python
from fastapi.middleware.cors import CORSMiddleware

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  # 프로덕션에서는 구체적으로 지정
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)
```

## 💡 Best Practices

### 1. 데이터 검증을 Pydantic에 위임

```python
# ❌ 나쁜 예: 수동 검증
if len(name) < 2 or len(name) > 50:
    raise ValueError("Name must be 2-50 characters")

# ✅ 좋은 예: Pydantic Field
name: str = Field(..., min_length=2, max_length=50)
```

### 2. 에러 응답 표준화

```python
# HTTPException으로 일관된 에러 응답
raise HTTPException(
    status_code=status.HTTP_404_NOT_FOUND,
    detail="User not found"
)
```

### 3. 응답 모델 분리

```python
# 요청과 응답 모델을 분리하여 보안 강화
class UserCreate(BaseModel):  # 요청
    name: str
    email: str

class User(UserCreate):  # 응답
    id: int
    created_at: datetime
```

## 🚀 다음 단계

축하합니다! 첫 REST API를 완성했습니다. 이제 다음을 시도해보세요:

### 학습 확장

1. **[Tutorial 2: JWT 인증 구현](/ko/tutorials/tutorial-02-jwt-auth)**
   - 이 API에 인증 시스템 추가하기

2. **데이터베이스 연결**
   - PostgreSQL 또는 MongoDB 연결
   - SQLAlchemy ORM 사용

3. **페이지네이션 추가**
   - Query parameters로 limit, offset 구현
   - 성능 최적화

### 실전 적용

- **자신의 프로젝트에 적용**: 이 패턴을 실제 프로젝트에 활용
- **API 확장**: 추가 리소스 (posts, comments 등) 구현
- **배포**: Vercel, Railway, Fly.io 등에 배포

### 추가 학습 자료

- [FastAPI 공식 문서](https://fastapi.tiangolo.com/)
- [Pydantic 가이드](https://docs.pydantic.dev/)
- [REST API 디자인 Best Practices](/ko/guides/api-design)

## 📚 참고 자료

- [MoAI-ADK SPEC 가이드](/ko/guides/spec-writing)
- [TDD 워크플로우](/ko/guides/tdd-workflow)
- [FastAPI 모범 사례](/ko/guides/fastapi-best-practices)

---

**질문이 있으신가요?** [Discord](https://discord.gg/moai-adk)에서 커뮤니티에 물어보세요!
