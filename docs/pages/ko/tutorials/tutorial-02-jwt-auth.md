---
title: "Tutorial 2: JWT 인증 시스템 구현"
description: "실무 수준의 JWT 기반 인증 시스템을 1시간 만에 구축합니다"
duration: "1시간"
difficulty: "중급"
tags: [tutorial, jwt, authentication, oauth2, security]
---

# Tutorial 2: JWT 인증 시스템 구현

이 튜토리얼에서는 실무에서 사용하는 JWT(JSON Web Token) 기반 인증 시스템을 처음부터 구축합니다. 회원가입, 로그인, 토큰 갱신, 보호된 라우트 구현까지 전체 인증 플로우를 배웁니다.

## 🎯 학습 목표

이 튜토리얼을 완료하면 다음을 할 수 있습니다:

- ✅ JWT 기반 인증 시스템 구조 이해하기
- ✅ 회원가입, 로그인, 로그아웃 기능 구현하기
- ✅ Access Token과 Refresh Token 관리하기
- ✅ bcrypt로 안전한 비밀번호 해싱하기
- ✅ OAuth2 Password Flow 구현하기
- ✅ Protected routes (인증 필요 엔드포인트) 만들기
- ✅ Senior Engineer Thinking으로 보안 Best Practices 적용하기

## 📋 사전 요구사항

### 필수 설치

- **Python 3.11+**
- **MoAI-ADK v0.23.0+**
- **Tutorial 1 완료** (REST API 기초 이해)
- **PostgreSQL** (선택, 프로덕션에서 권장)

### 선행 지식

- REST API 기본 (Tutorial 1)
- HTTP 헤더 (Authorization header)
- 기본적인 암호학 개념 (해싱, 암호화)
- JSON 포맷

### 설치 확인

```bash
# 프로젝트 디렉토리 생성
mkdir jwt-auth-tutorial
cd jwt-auth-tutorial

# MoAI-ADK 초기화
moai-adk init
```

## 🔐 JWT 인증 시스템 아키텍처

### 인증 플로우

```mermaid
sequenceDiagram
    participant User
    participant API
    participant DB

    User->>API: POST /auth/register (email, password)
    API->>DB: 사용자 저장 (해시된 비밀번호)
    DB-->>API: 사용자 생성 완료
    API-->>User: 201 Created

    User->>API: POST /auth/login (email, password)
    API->>DB: 사용자 조회
    DB-->>API: 사용자 데이터
    API->>API: 비밀번호 검증
    API->>API: JWT 토큰 생성
    API-->>User: access_token, refresh_token

    User->>API: GET /users/me (Authorization: Bearer {token})
    API->>API: JWT 검증
    API->>DB: 사용자 데이터 조회
    DB-->>API: 사용자 정보
    API-->>User: 사용자 프로필

    User->>API: POST /auth/refresh (refresh_token)
    API->>API: Refresh Token 검증
    API->>API: 새 Access Token 생성
    API-->>User: 새 access_token
```

### 보안 계층

```mermaid
graph LR
    A[평문 비밀번호] -->|bcrypt hash| B[해시된 비밀번호]
    B -->|DB 저장| C[안전한 저장소]

    D[로그인 성공] -->|JWT 생성| E[Access Token<br/>15분]
    D -->|JWT 생성| F[Refresh Token<br/>7일]

    E -->|검증| G[Protected Routes]
    F -->|갱신| E
```

## 🚀 프로젝트 구조

```
jwt-auth-tutorial/
├── .moai/
│   ├── config.json
│   └── specs/
│       └── SPEC-JWT-AUTH-001.md
├── src/
│   └── auth_api/
│       ├── __init__.py
│       ├── main.py              # FastAPI 앱
│       ├── config.py            # 설정 (환경변수)
│       ├── models.py            # Pydantic 모델
│       ├── database.py          # DB 연결 및 모델
│       ├── auth.py              # 인증 로직
│       ├── security.py          # JWT 및 암호화
│       ├── dependencies.py      # 의존성 (get_current_user)
│       └── routes/
│           ├── __init__.py
│           ├── auth.py          # 인증 라우트
│           └── users.py         # 사용자 라우트
├── tests/
│   ├── __init__.py
│   ├── test_auth.py             # 인증 테스트
│   └── test_security.py         # 보안 테스트
├── .env.example
├── requirements.txt
└── README.md
```

## 단계별 실습

### Step 1: SPEC 작성

```bash
/alfred:1-plan "JWT 인증 시스템 구현"
```

**생성된 SPEC** (`.moai/specs/SPEC-JWT-AUTH-001.md`):

```markdown
# SPEC-JWT-AUTH-001: JWT 인증 시스템

## 요구사항

JWT 기반 인증 시스템을 구현하여 API 보안을 강화한다.

### 기능 요구사항

#### 인증 (Authentication)

- FR-001: 회원가입 (POST /auth/register)
  - 이메일, 비밀번호, 이름 입력
  - 이메일 중복 확인
  - 비밀번호 bcrypt 해싱

- FR-002: 로그인 (POST /auth/login)
  - 이메일, 비밀번호로 인증
  - Access Token (15분 유효)
  - Refresh Token (7일 유효)

- FR-003: 토큰 갱신 (POST /auth/refresh)
  - Refresh Token으로 새 Access Token 발급

- FR-004: 로그아웃 (POST /auth/logout)
  - Refresh Token 무효화

#### 보호된 리소스

- FR-005: 내 프로필 조회 (GET /users/me)
  - 유효한 Access Token 필요

- FR-006: 프로필 수정 (PUT /users/me)
  - 이름, 이메일 수정 가능

### 보안 요구사항

- SR-001: 비밀번호는 bcrypt로 해싱 (최소 8자)
- SR-002: JWT는 HS256 알고리즘 사용
- SR-003: Access Token은 15분 후 만료
- SR-004: Refresh Token은 7일 후 만료
- SR-005: Refresh Token은 DB에 저장하여 관리

### 데이터 모델

User:
- id: UUID (자동 생성)
- email: str (unique, 이메일 형식)
- hashed_password: str
- full_name: str
- is_active: bool (기본값: True)
- created_at: datetime

RefreshToken:
- id: UUID
- user_id: UUID (FK)
- token: str (unique)
- expires_at: datetime
- created_at: datetime
```

### Step 2: 환경 설정

**requirements.txt**:
```txt
fastapi==0.104.1
uvicorn[standard]==0.24.0
pydantic==2.5.0
pydantic-settings==2.1.0
python-jose[cryptography]==3.3.0
passlib[bcrypt]==1.7.4
python-multipart==0.0.6
sqlalchemy==2.0.23
alembic==1.13.0
psycopg2-binary==2.9.9
pytest==7.4.3
httpx==0.25.2
```

설치:
```bash
pip install -r requirements.txt
```

**.env.example**:
```env
# JWT 설정
SECRET_KEY=your-secret-key-here-change-in-production
ALGORITHM=HS256
ACCESS_TOKEN_EXPIRE_MINUTES=15
REFRESH_TOKEN_EXPIRE_DAYS=7

# Database (개발용 SQLite)
DATABASE_URL=sqlite:///./test.db

# 프로덕션 PostgreSQL (선택)
# DATABASE_URL=postgresql://user:password@localhost/dbname
```

실제 `.env` 파일 생성:
```bash
cp .env.example .env

# SECRET_KEY 생성
python -c "import secrets; print(secrets.token_urlsafe(32))"
# 출력된 값을 .env의 SECRET_KEY에 복사
```

### Step 3: 설정 관리

**src/auth_api/config.py**:

```python
"""
애플리케이션 설정
"""
from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    """애플리케이션 설정"""

    # JWT 설정
    SECRET_KEY: str
    ALGORITHM: str = "HS256"
    ACCESS_TOKEN_EXPIRE_MINUTES: int = 15
    REFRESH_TOKEN_EXPIRE_DAYS: int = 7

    # Database
    DATABASE_URL: str = "sqlite:///./test.db"

    # 보안
    PASSWORD_MIN_LENGTH: int = 8

    class Config:
        env_file = ".env"
        case_sensitive = True


settings = Settings()
```

### Step 4: 데이터베이스 모델

**src/auth_api/database.py**:

```python
"""
데이터베이스 모델 및 세션
"""
import uuid
from datetime import datetime
from sqlalchemy import Boolean, Column, String, DateTime, ForeignKey
from sqlalchemy.dialects.postgresql import UUID
from sqlalchemy.orm import declarative_base, relationship, Session
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker
from .config import settings

Base = declarative_base()


class User(Base):
    """사용자 모델"""
    __tablename__ = "users"

    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    email = Column(String, unique=True, index=True, nullable=False)
    hashed_password = Column(String, nullable=False)
    full_name = Column(String, nullable=False)
    is_active = Column(Boolean, default=True)
    created_at = Column(DateTime, default=datetime.utcnow)

    refresh_tokens = relationship("RefreshToken", back_populates="user", cascade="all, delete-orphan")


class RefreshToken(Base):
    """Refresh Token 모델"""
    __tablename__ = "refresh_tokens"

    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    user_id = Column(UUID(as_uuid=True), ForeignKey("users.id"), nullable=False)
    token = Column(String, unique=True, index=True, nullable=False)
    expires_at = Column(DateTime, nullable=False)
    created_at = Column(DateTime, default=datetime.utcnow)

    user = relationship("User", back_populates="refresh_tokens")


# 데이터베이스 엔진 생성
engine = create_engine(
    settings.DATABASE_URL,
    connect_args={"check_same_thread": False} if "sqlite" in settings.DATABASE_URL else {}
)

# 세션 팩토리
SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)


def get_db():
    """데이터베이스 세션 의존성"""
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()


def init_db():
    """데이터베이스 테이블 생성"""
    Base.metadata.create_all(bind=engine)
```

### Step 5: 보안 유틸리티

**src/auth_api/security.py**:

```python
"""
JWT 및 암호화 유틸리티
"""
from datetime import datetime, timedelta
from typing import Optional
from jose import JWTError, jwt
from passlib.context import CryptContext
from .config import settings

# 비밀번호 해싱 컨텍스트
pwd_context = CryptContext(schemes=["bcrypt"], deprecated="auto")


def verify_password(plain_password: str, hashed_password: str) -> bool:
    """비밀번호 검증"""
    return pwd_context.verify(plain_password, hashed_password)


def get_password_hash(password: str) -> str:
    """비밀번호 해싱"""
    return pwd_context.hash(password)


def create_access_token(data: dict, expires_delta: Optional[timedelta] = None) -> str:
    """
    Access Token 생성

    Args:
        data: JWT payload에 포함할 데이터
        expires_delta: 만료 시간 (기본값: 15분)

    Returns:
        JWT 토큰 문자열
    """
    to_encode = data.copy()

    if expires_delta:
        expire = datetime.utcnow() + expires_delta
    else:
        expire = datetime.utcnow() + timedelta(minutes=settings.ACCESS_TOKEN_EXPIRE_MINUTES)

    to_encode.update({
        "exp": expire,
        "type": "access"
    })

    encoded_jwt = jwt.encode(
        to_encode,
        settings.SECRET_KEY,
        algorithm=settings.ALGORITHM
    )
    return encoded_jwt


def create_refresh_token(data: dict) -> str:
    """
    Refresh Token 생성

    Args:
        data: JWT payload에 포함할 데이터

    Returns:
        JWT 토큰 문자열
    """
    to_encode = data.copy()
    expire = datetime.utcnow() + timedelta(days=settings.REFRESH_TOKEN_EXPIRE_DAYS)

    to_encode.update({
        "exp": expire,
        "type": "refresh"
    })

    encoded_jwt = jwt.encode(
        to_encode,
        settings.SECRET_KEY,
        algorithm=settings.ALGORITHM
    )
    return encoded_jwt


def decode_token(token: str) -> dict:
    """
    JWT 토큰 디코딩

    Args:
        token: JWT 토큰 문자열

    Returns:
        디코딩된 payload

    Raises:
        JWTError: 토큰이 유효하지 않을 때
    """
    payload = jwt.decode(
        token,
        settings.SECRET_KEY,
        algorithms=[settings.ALGORITHM]
    )
    return payload
```

### Step 6: Pydantic 모델

**src/auth_api/models.py**:

```python
"""
Pydantic 모델 (Request/Response)
"""
from datetime import datetime
from typing import Optional
from uuid import UUID
from pydantic import BaseModel, EmailStr, Field


class UserRegister(BaseModel):
    """회원가입 요청"""
    email: EmailStr = Field(..., description="이메일 주소")
    password: str = Field(..., min_length=8, description="비밀번호 (최소 8자)")
    full_name: str = Field(..., min_length=2, max_length=100, description="이름")


class UserLogin(BaseModel):
    """로그인 요청"""
    email: EmailStr
    password: str


class TokenResponse(BaseModel):
    """토큰 응답"""
    access_token: str
    refresh_token: str
    token_type: str = "bearer"


class TokenRefresh(BaseModel):
    """토큰 갱신 요청"""
    refresh_token: str


class UserResponse(BaseModel):
    """사용자 응답"""
    id: UUID
    email: EmailStr
    full_name: str
    is_active: bool
    created_at: datetime

    class Config:
        from_attributes = True


class UserUpdate(BaseModel):
    """사용자 정보 수정"""
    full_name: Optional[str] = Field(None, min_length=2, max_length=100)
    email: Optional[EmailStr] = None
```

### Step 7: 인증 로직

**src/auth_api/auth.py**:

```python
"""
인증 비즈니스 로직
"""
from datetime import datetime, timedelta
from uuid import UUID
from sqlalchemy.orm import Session
from fastapi import HTTPException, status
from .database import User, RefreshToken
from .models import UserRegister, UserLogin, TokenResponse
from .security import (
    verify_password,
    get_password_hash,
    create_access_token,
    create_refresh_token,
    decode_token,
)
from .config import settings


def register_user(db: Session, user_data: UserRegister) -> User:
    """
    새 사용자 등록

    Args:
        db: 데이터베이스 세션
        user_data: 회원가입 데이터

    Returns:
        생성된 사용자 객체

    Raises:
        HTTPException: 이메일이 이미 존재할 때
    """
    # 이메일 중복 확인
    existing_user = db.query(User).filter(User.email == user_data.email).first()
    if existing_user:
        raise HTTPException(
            status_code=status.HTTP_409_CONFLICT,
            detail="Email already registered"
        )

    # 사용자 생성
    user = User(
        email=user_data.email,
        hashed_password=get_password_hash(user_data.password),
        full_name=user_data.full_name,
    )

    db.add(user)
    db.commit()
    db.refresh(user)

    return user


def authenticate_user(db: Session, login_data: UserLogin) -> TokenResponse:
    """
    사용자 인증 및 토큰 발급

    Args:
        db: 데이터베이스 세션
        login_data: 로그인 데이터

    Returns:
        Access Token 및 Refresh Token

    Raises:
        HTTPException: 인증 실패 시
    """
    # 사용자 조회
    user = db.query(User).filter(User.email == login_data.email).first()

    if not user or not verify_password(login_data.password, user.hashed_password):
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Incorrect email or password",
            headers={"WWW-Authenticate": "Bearer"},
        )

    if not user.is_active:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="User account is inactive"
        )

    # 토큰 생성
    access_token = create_access_token(data={"sub": str(user.id), "email": user.email})
    refresh_token_str = create_refresh_token(data={"sub": str(user.id)})

    # Refresh Token DB에 저장
    refresh_token = RefreshToken(
        user_id=user.id,
        token=refresh_token_str,
        expires_at=datetime.utcnow() + timedelta(days=settings.REFRESH_TOKEN_EXPIRE_DAYS)
    )
    db.add(refresh_token)
    db.commit()

    return TokenResponse(
        access_token=access_token,
        refresh_token=refresh_token_str
    )


def refresh_access_token(db: Session, refresh_token_str: str) -> str:
    """
    Refresh Token으로 새 Access Token 발급

    Args:
        db: 데이터베이스 세션
        refresh_token_str: Refresh Token

    Returns:
        새 Access Token

    Raises:
        HTTPException: 토큰이 유효하지 않을 때
    """
    try:
        payload = decode_token(refresh_token_str)

        # 토큰 타입 확인
        if payload.get("type") != "refresh":
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="Invalid token type"
            )

        user_id = payload.get("sub")
        if not user_id:
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="Invalid token payload"
            )

        # DB에서 토큰 확인
        refresh_token = db.query(RefreshToken).filter(
            RefreshToken.token == refresh_token_str,
            RefreshToken.user_id == UUID(user_id)
        ).first()

        if not refresh_token:
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="Refresh token not found"
            )

        # 토큰 만료 확인
        if refresh_token.expires_at < datetime.utcnow():
            db.delete(refresh_token)
            db.commit()
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="Refresh token expired"
            )

        # 새 Access Token 생성
        user = db.query(User).filter(User.id == UUID(user_id)).first()
        access_token = create_access_token(data={"sub": user_id, "email": user.email})

        return access_token

    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Could not validate credentials"
        )


def logout_user(db: Session, refresh_token_str: str):
    """
    사용자 로그아웃 (Refresh Token 무효화)

    Args:
        db: 데이터베이스 세션
        refresh_token_str: Refresh Token
    """
    refresh_token = db.query(RefreshToken).filter(
        RefreshToken.token == refresh_token_str
    ).first()

    if refresh_token:
        db.delete(refresh_token)
        db.commit()
```

### Step 8: 의존성 (get_current_user)

**src/auth_api/dependencies.py**:

```python
"""
FastAPI 의존성
"""
from uuid import UUID
from fastapi import Depends, HTTPException, status
from fastapi.security import OAuth2PasswordBearer
from sqlalchemy.orm import Session
from .database import User, get_db
from .security import decode_token

oauth2_scheme = OAuth2PasswordBearer(tokenUrl="/auth/login")


async def get_current_user(
    token: str = Depends(oauth2_scheme),
    db: Session = Depends(get_db)
) -> User:
    """
    현재 인증된 사용자 조회

    Args:
        token: JWT Access Token
        db: 데이터베이스 세션

    Returns:
        현재 사용자 객체

    Raises:
        HTTPException: 토큰이 유효하지 않거나 사용자를 찾을 수 없을 때
    """
    credentials_exception = HTTPException(
        status_code=status.HTTP_401_UNAUTHORIZED,
        detail="Could not validate credentials",
        headers={"WWW-Authenticate": "Bearer"},
    )

    try:
        payload = decode_token(token)

        # 토큰 타입 확인
        if payload.get("type") != "access":
            raise credentials_exception

        user_id: str = payload.get("sub")
        if user_id is None:
            raise credentials_exception

    except Exception:
        raise credentials_exception

    # 사용자 조회
    user = db.query(User).filter(User.id == UUID(user_id)).first()
    if user is None:
        raise credentials_exception

    if not user.is_active:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="User account is inactive"
        )

    return user
```

### Step 9: 인증 라우트

**src/auth_api/routes/auth.py**:

```python
"""
인증 관련 라우트
"""
from fastapi import APIRouter, Depends, status
from sqlalchemy.orm import Session
from ..database import get_db
from ..models import UserRegister, UserLogin, UserResponse, TokenResponse, TokenRefresh
from ..auth import register_user, authenticate_user, refresh_access_token, logout_user

router = APIRouter(prefix="/auth", tags=["authentication"])


@router.post("/register", response_model=UserResponse, status_code=status.HTTP_201_CREATED)
def register(user_data: UserRegister, db: Session = Depends(get_db)):
    """
    회원가입

    - **email**: 유효한 이메일 주소
    - **password**: 최소 8자 이상
    - **full_name**: 사용자 이름
    """
    user = register_user(db, user_data)
    return user


@router.post("/login", response_model=TokenResponse)
def login(login_data: UserLogin, db: Session = Depends(get_db)):
    """
    로그인

    성공 시 Access Token과 Refresh Token을 반환합니다.
    - Access Token: 15분 유효
    - Refresh Token: 7일 유효
    """
    return authenticate_user(db, login_data)


@router.post("/refresh")
def refresh_token(token_data: TokenRefresh, db: Session = Depends(get_db)):
    """
    Access Token 갱신

    Refresh Token으로 새로운 Access Token을 발급받습니다.
    """
    new_access_token = refresh_access_token(db, token_data.refresh_token)
    return {"access_token": new_access_token, "token_type": "bearer"}


@router.post("/logout", status_code=status.HTTP_204_NO_CONTENT)
def logout(token_data: TokenRefresh, db: Session = Depends(get_db)):
    """
    로그아웃

    Refresh Token을 무효화합니다.
    """
    logout_user(db, token_data.refresh_token)
```

### Step 10: 사용자 라우트 (Protected)

**src/auth_api/routes/users.py**:

```python
"""
사용자 관련 라우트 (인증 필요)
"""
from fastapi import APIRouter, Depends
from sqlalchemy.orm import Session
from ..database import User, get_db
from ..models import UserResponse, UserUpdate
from ..dependencies import get_current_user

router = APIRouter(prefix="/users", tags=["users"])


@router.get("/me", response_model=UserResponse)
def get_my_profile(current_user: User = Depends(get_current_user)):
    """
    내 프로필 조회

    인증된 사용자의 프로필 정보를 반환합니다.
    **Authorization 헤더에 Bearer 토큰 필요**
    """
    return current_user


@router.put("/me", response_model=UserResponse)
def update_my_profile(
    user_data: UserUpdate,
    current_user: User = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """
    내 프로필 수정

    이름, 이메일을 수정할 수 있습니다.
    **Authorization 헤더에 Bearer 토큰 필요**
    """
    update_data = user_data.model_dump(exclude_unset=True)

    for field, value in update_data.items():
        setattr(current_user, field, value)

    db.commit()
    db.refresh(current_user)

    return current_user
```

### Step 11: FastAPI 앱

**src/auth_api/main.py**:

```python
"""
JWT 인증 API 애플리케이션
"""
from fastapi import FastAPI
from .database import init_db
from .routes import auth, users

# 데이터베이스 초기화
init_db()

app = FastAPI(
    title="JWT Authentication API",
    description="JWT 기반 인증 시스템",
    version="1.0.0",
)

# 라우트 등록
app.include_router(auth.router)
app.include_router(users.router)


@app.get("/")
def root():
    """API 루트"""
    return {
        "message": "JWT Authentication API",
        "docs": "/docs",
        "version": "1.0.0"
    }
```

### Step 12: 테스트 작성

**tests/test_auth.py**:

```python
"""
인증 API 테스트
"""
import pytest
from fastapi.testclient import TestClient
from src.auth_api.main import app
from src.auth_api.database import Base, engine, get_db, SessionLocal

# 테스트용 데이터베이스
@pytest.fixture(scope="function")
def db_session():
    Base.metadata.create_all(bind=engine)
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()
        Base.metadata.drop_all(bind=engine)


@pytest.fixture
def client(db_session):
    def override_get_db():
        try:
            yield db_session
        finally:
            pass

    app.dependency_overrides[get_db] = override_get_db
    return TestClient(app)


def test_register_user(client):
    """회원가입 테스트"""
    response = client.post("/auth/register", json={
        "email": "test@example.com",
        "password": "password123",
        "full_name": "Test User"
    })

    assert response.status_code == 201
    data = response.json()
    assert data["email"] == "test@example.com"
    assert data["full_name"] == "Test User"
    assert "id" in data


def test_register_duplicate_email(client):
    """이메일 중복 테스트"""
    user_data = {
        "email": "test@example.com",
        "password": "password123",
        "full_name": "Test User"
    }

    client.post("/auth/register", json=user_data)
    response = client.post("/auth/register", json=user_data)

    assert response.status_code == 409


def test_login_success(client):
    """로그인 성공 테스트"""
    # 회원가입
    client.post("/auth/register", json={
        "email": "test@example.com",
        "password": "password123",
        "full_name": "Test User"
    })

    # 로그인
    response = client.post("/auth/login", json={
        "email": "test@example.com",
        "password": "password123"
    })

    assert response.status_code == 200
    data = response.json()
    assert "access_token" in data
    assert "refresh_token" in data
    assert data["token_type"] == "bearer"


def test_login_wrong_password(client):
    """잘못된 비밀번호로 로그인 테스트"""
    client.post("/auth/register", json={
        "email": "test@example.com",
        "password": "password123",
        "full_name": "Test User"
    })

    response = client.post("/auth/login", json={
        "email": "test@example.com",
        "password": "wrongpassword"
    })

    assert response.status_code == 401


def test_get_profile_with_token(client):
    """토큰으로 프로필 조회 테스트"""
    # 회원가입
    client.post("/auth/register", json={
        "email": "test@example.com",
        "password": "password123",
        "full_name": "Test User"
    })

    # 로그인
    login_response = client.post("/auth/login", json={
        "email": "test@example.com",
        "password": "password123"
    })
    token = login_response.json()["access_token"]

    # 프로필 조회
    response = client.get(
        "/users/me",
        headers={"Authorization": f"Bearer {token}"}
    )

    assert response.status_code == 200
    data = response.json()
    assert data["email"] == "test@example.com"


def test_get_profile_without_token(client):
    """토큰 없이 프로필 조회 테스트"""
    response = client.get("/users/me")
    assert response.status_code == 401


def test_refresh_token(client):
    """토큰 갱신 테스트"""
    # 회원가입 및 로그인
    client.post("/auth/register", json={
        "email": "test@example.com",
        "password": "password123",
        "full_name": "Test User"
    })

    login_response = client.post("/auth/login", json={
        "email": "test@example.com",
        "password": "password123"
    })
    refresh_token = login_response.json()["refresh_token"]

    # 토큰 갱신
    response = client.post("/auth/refresh", json={
        "refresh_token": refresh_token
    })

    assert response.status_code == 200
    assert "access_token" in response.json()
```

## ✅ 테스트 및 검증

### 애플리케이션 실행

```bash
uvicorn src.auth_api.main:app --reload
```

### 수동 테스트 (cURL)

```bash
# 1. 회원가입
curl -X POST "http://localhost:8000/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "alice@example.com",
    "password": "secret123",
    "full_name": "Alice"
  }'

# 2. 로그인
curl -X POST "http://localhost:8000/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "alice@example.com",
    "password": "secret123"
  }'

# 응답에서 access_token 복사

# 3. 프로필 조회 (토큰 필요)
curl -X GET "http://localhost:8000/users/me" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"

# 4. 토큰 갱신
curl -X POST "http://localhost:8000/auth/refresh" \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "YOUR_REFRESH_TOKEN"
  }'
```

### 자동화된 테스트

```bash
pytest tests/test_auth.py -v
```

## 🔧 문제 해결

### 문제 1: SECRET_KEY not set

**증상**:
```
pydantic.error_wrappers.ValidationError: SECRET_KEY
```

**해결**:
```bash
# .env 파일에 SECRET_KEY 추가
echo "SECRET_KEY=$(python -c 'import secrets; print(secrets.token_urlsafe(32))')" >> .env
```

### 문제 2: Token expired

**증상**:
```
401 Unauthorized: Token has expired
```

**해결**: Refresh Token으로 새 Access Token 발급

```python
# /auth/refresh 엔드포인트 사용
```

## 💡 Best Practices & Security

### 1. 비밀번호 저장

```python
# ❌ 나쁜 예: 평문 저장
user.password = "secret123"

# ✅ 좋은 예: bcrypt 해싱
user.hashed_password = get_password_hash("secret123")
```

### 2. 토큰 만료 시간

```python
# Access Token: 짧게 (15분)
ACCESS_TOKEN_EXPIRE_MINUTES = 15

# Refresh Token: 길게 (7일)
REFRESH_TOKEN_EXPIRE_DAYS = 7
```

### 3. HTTPS 사용 (프로덕션)

```python
# 프로덕션에서는 반드시 HTTPS
# HTTP로 JWT 전송 금지
```

### 4. Refresh Token 관리

```python
# DB에 저장하여 무효화 가능하게
# 로그아웃 시 DB에서 삭제
```

## 🚀 다음 단계

축하합니다! JWT 인증 시스템을 완성했습니다.

### 추가 기능 구현

1. **이메일 인증**: 회원가입 시 이메일 확인
2. **비밀번호 재설정**: 이메일로 재설정 링크 전송
3. **OAuth2 소셜 로그인**: Google, GitHub 연동
4. **2FA (Two-Factor Authentication)**: TOTP 기반 2단계 인증

### 다음 튜토리얼

- **[Tutorial 3: 데이터베이스 최적화](/ko/tutorials/tutorial-03-database-optimization)**
  - 성능 개선 및 캐싱 전략

## 📚 참고 자료

- [JWT.io](https://jwt.io/)
- [OWASP Authentication Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)
- [FastAPI Security](https://fastapi.tiangolo.com/tutorial/security/)

---

**Happy Coding! 🚀**
