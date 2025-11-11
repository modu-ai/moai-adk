---
title: "JWT 기본 인증"
category: "authentication"
difficulty: "초급"
tags: [jwt, auth, fastapi, security, python]
---

# JWT 기본 인증

## 개요

JSON Web Token (JWT)을 사용한 기본 인증 구현 예제입니다. 사용자 로그인 시 JWT 토큰을 발급하고, 보호된 엔드포인트 접근 시 토큰을 검증합니다.

## 사용 사례

- 사용자 로그인/로그아웃
- API 엔드포인트 보호
- Stateless 인증
- 모바일 앱 인증

## 완전한 코드 예제

### 1. 의존성 설치

```bash
uv pip install fastapi python-jose[cryptography] passlib[bcrypt] python-multipart uvicorn
```

### 2. 인증 유틸리티 (`auth.py`)

```python
# SPEC: AUTH-001 - JWT 인증 유틸리티
# @TAG:AUTH-001

from datetime import datetime, timedelta
from typing import Optional
from jose import JWTError, jwt
from passlib.context import CryptContext
from fastapi import Depends, HTTPException, status
from fastapi.security import OAuth2PasswordBearer
from pydantic import BaseModel
import os

# 환경 변수에서 설정 로드
SECRET_KEY = os.getenv("SECRET_KEY", "your-secret-key-change-in-production")
ALGORITHM = "HS256"
ACCESS_TOKEN_EXPIRE_MINUTES = 30

# 비밀번호 해싱 설정
pwd_context = CryptContext(schemes=["bcrypt"], deprecated="auto")

# OAuth2 스킴 (Authorization: Bearer {token})
oauth2_scheme = OAuth2PasswordBearer(tokenUrl="auth/login")

class Token(BaseModel):
    """토큰 응답 스키마"""
    access_token: str
    token_type: str = "bearer"

class TokenData(BaseModel):
    """토큰 페이로드 스키마"""
    user_id: Optional[int] = None
    username: Optional[str] = None

def verify_password(plain_password: str, hashed_password: str) -> bool:
    """
    비밀번호 검증

    Args:
        plain_password: 평문 비밀번호
        hashed_password: 해시된 비밀번호

    Returns:
        일치 여부
    """
    return pwd_context.verify(plain_password, hashed_password)

def get_password_hash(password: str) -> str:
    """
    비밀번호 해싱

    Args:
        password: 평문 비밀번호

    Returns:
        해시된 비밀번호
    """
    return pwd_context.hash(password)

def create_access_token(data: dict, expires_delta: Optional[timedelta] = None) -> str:
    """
    JWT 액세스 토큰 생성

    Args:
        data: 토큰에 포함할 데이터
        expires_delta: 만료 시간 (기본: 30분)

    Returns:
        JWT 토큰 문자열
    """
    to_encode = data.copy()

    # 만료 시간 설정
    if expires_delta:
        expire = datetime.utcnow() + expires_delta
    else:
        expire = datetime.utcnow() + timedelta(minutes=ACCESS_TOKEN_EXPIRE_MINUTES)

    to_encode.update({"exp": expire})

    # JWT 인코딩
    encoded_jwt = jwt.encode(to_encode, SECRET_KEY, algorithm=ALGORITHM)
    return encoded_jwt

def decode_access_token(token: str) -> TokenData:
    """
    JWT 토큰 디코딩 및 검증

    Args:
        token: JWT 토큰 문자열

    Returns:
        TokenData 객체

    Raises:
        HTTPException: 토큰이 유효하지 않은 경우
    """
    credentials_exception = HTTPException(
        status_code=status.HTTP_401_UNAUTHORIZED,
        detail="인증 정보를 확인할 수 없습니다",
        headers={"WWW-Authenticate": "Bearer"},
    )

    try:
        payload = jwt.decode(token, SECRET_KEY, algorithms=[ALGORITHM])
        user_id: int = payload.get("sub")
        username: str = payload.get("username")

        if user_id is None:
            raise credentials_exception

        return TokenData(user_id=user_id, username=username)

    except JWTError:
        raise credentials_exception
```

### 3. 사용자 모델 및 CRUD (`models.py`, `crud.py`)

```python
# SPEC: AUTH-002 - 사용자 모델
# @TAG:AUTH-002

from sqlalchemy import Column, Integer, String, Boolean, DateTime
from sqlalchemy.sql import func
from database import Base

class User(Base):
    """사용자 모델"""
    __tablename__ = "users"

    id = Column(Integer, primary_key=True, index=True)
    username = Column(String, unique=True, index=True, nullable=False)
    email = Column(String, unique=True, index=True, nullable=False)
    hashed_password = Column(String, nullable=False)
    is_active = Column(Boolean, default=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now())

# ===== CRUD 함수 (crud.py) =====

from sqlalchemy.orm import Session
from typing import Optional
import models
import auth

def get_user_by_username(db: Session, username: str) -> Optional[models.User]:
    """사용자명으로 사용자 조회"""
    return db.query(models.User).filter(
        models.User.username == username
    ).first()

def get_user_by_email(db: Session, email: str) -> Optional[models.User]:
    """이메일로 사용자 조회"""
    return db.query(models.User).filter(
        models.User.email == email
    ).first()

def create_user(db: Session, username: str, email: str, password: str) -> models.User:
    """
    새 사용자 생성

    Args:
        db: 데이터베이스 세션
        username: 사용자명
        email: 이메일
        password: 평문 비밀번호

    Returns:
        생성된 User 객체
    """
    hashed_password = auth.get_password_hash(password)

    db_user = models.User(
        username=username,
        email=email,
        hashed_password=hashed_password
    )

    db.add(db_user)
    db.commit()
    db.refresh(db_user)

    return db_user

def authenticate_user(db: Session, username: str, password: str) -> Optional[models.User]:
    """
    사용자 인증

    Args:
        db: 데이터베이스 세션
        username: 사용자명
        password: 평문 비밀번호

    Returns:
        인증된 User 객체 또는 None
    """
    user = get_user_by_username(db, username)

    if not user:
        return None

    if not auth.verify_password(password, user.hashed_password):
        return None

    return user
```

### 4. API 엔드포인트 (`main.py`)

```python
# SPEC: AUTH-003 - 인증 API
# @TAG:AUTH-003

from fastapi import FastAPI, Depends, HTTPException, status
from fastapi.security import OAuth2PasswordRequestForm
from sqlalchemy.orm import Session
from pydantic import BaseModel, EmailStr
from typing import Optional
from datetime import timedelta

import models
import auth
import crud
from database import engine, get_db

# 테이블 생성
models.Base.metadata.create_all(bind=engine)

app = FastAPI(
    title="JWT Authentication API",
    description="JWT 기반 인증 시스템"
)

# ===== 스키마 =====

class UserRegister(BaseModel):
    """사용자 등록 스키마"""
    username: str
    email: EmailStr
    password: str

class UserResponse(BaseModel):
    """사용자 응답 스키마"""
    id: int
    username: str
    email: str
    is_active: bool

    class Config:
        from_attributes = True

# ===== 의존성: 현재 사용자 =====

async def get_current_user(
    token: str = Depends(auth.oauth2_scheme),
    db: Session = Depends(get_db)
) -> models.User:
    """
    현재 인증된 사용자 조회

    Args:
        token: JWT 액세스 토큰
        db: 데이터베이스 세션

    Returns:
        User 객체

    Raises:
        HTTPException: 토큰이 유효하지 않은 경우
    """
    # 토큰 디코딩
    token_data = auth.decode_access_token(token)

    # 사용자 조회
    user = db.query(models.User).filter(
        models.User.id == token_data.user_id
    ).first()

    if user is None:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="사용자를 찾을 수 없습니다"
        )

    if not user.is_active:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="비활성화된 사용자입니다"
        )

    return user

# ===== 인증 엔드포인트 =====

@app.post("/auth/register", response_model=UserResponse, status_code=status.HTTP_201_CREATED)
def register(user: UserRegister, db: Session = Depends(get_db)):
    """
    사용자 등록

    Returns:
        생성된 사용자 정보 (비밀번호 제외)
    """
    # 중복 체크
    if crud.get_user_by_username(db, user.username):
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="이미 사용 중인 사용자명입니다"
        )

    if crud.get_user_by_email(db, user.email):
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="이미 사용 중인 이메일입니다"
        )

    # 사용자 생성
    db_user = crud.create_user(
        db=db,
        username=user.username,
        email=user.email,
        password=user.password
    )

    return db_user

@app.post("/auth/login", response_model=auth.Token)
def login(
    form_data: OAuth2PasswordRequestForm = Depends(),
    db: Session = Depends(get_db)
):
    """
    로그인 - JWT 토큰 발급

    Args:
        form_data: username과 password

    Returns:
        JWT 액세스 토큰
    """
    # 사용자 인증
    user = crud.authenticate_user(db, form_data.username, form_data.password)

    if not user:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="잘못된 사용자명 또는 비밀번호입니다",
            headers={"WWW-Authenticate": "Bearer"},
        )

    # 토큰 생성
    access_token_expires = timedelta(minutes=auth.ACCESS_TOKEN_EXPIRE_MINUTES)
    access_token = auth.create_access_token(
        data={"sub": user.id, "username": user.username},
        expires_delta=access_token_expires
    )

    return {"access_token": access_token, "token_type": "bearer"}

# ===== 보호된 엔드포인트 =====

@app.get("/users/me", response_model=UserResponse)
def read_users_me(current_user: models.User = Depends(get_current_user)):
    """
    현재 사용자 정보 조회 (보호된 엔드포인트)

    **인증 필요**: Bearer Token

    Returns:
        현재 로그인한 사용자 정보
    """
    return current_user

@app.get("/users/me/items")
def read_own_items(current_user: models.User = Depends(get_current_user)):
    """사용자의 아이템 조회 (예시)"""
    return [
        {"item_id": 1, "owner": current_user.username},
        {"item_id": 2, "owner": current_user.username}
    ]

@app.get("/")
def root():
    """API 정보"""
    return {
        "message": "JWT Authentication API",
        "docs": "/docs",
        "endpoints": {
            "register": "POST /auth/register",
            "login": "POST /auth/login",
            "me": "GET /users/me (인증 필요)"
        }
    }
```

## 사용 방법

### 1. 사용자 등록

```bash
curl -X POST "http://localhost:8000/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john_doe",
    "email": "john@example.com",
    "password": "SecurePassword123"
  }'
```

### 2. 로그인 (토큰 발급)

```bash
curl -X POST "http://localhost:8000/auth/login" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "username=john_doe&password=SecurePassword123"

# 응답
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "bearer"
}
```

### 3. 보호된 엔드포인트 접근

```bash
# 토큰을 Authorization 헤더에 포함
curl -X GET "http://localhost:8000/users/me" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

## 테스트 코드

```python
# SPEC: TEST-AUTH-001 - JWT 인증 테스트
# @TAG:TEST-AUTH-001

import pytest
from fastapi.testclient import TestClient
from main import app

client = TestClient(app)

def test_register_user():
    """사용자 등록 테스트"""
    response = client.post(
        "/auth/register",
        json={
            "username": "testuser",
            "email": "test@example.com",
            "password": "TestPassword123"
        }
    )
    assert response.status_code == 201
    data = response.json()
    assert data["username"] == "testuser"
    assert data["email"] == "test@example.com"
    assert "hashed_password" not in data  # 비밀번호 노출 방지

def test_login_success():
    """로그인 성공 테스트"""
    # 먼저 사용자 등록
    client.post(
        "/auth/register",
        json={"username": "logintest", "email": "login@example.com", "password": "Pass123"}
    )

    # 로그인
    response = client.post(
        "/auth/login",
        data={"username": "logintest", "password": "Pass123"}
    )
    assert response.status_code == 200
    data = response.json()
    assert "access_token" in data
    assert data["token_type"] == "bearer"

def test_login_wrong_password():
    """잘못된 비밀번호 테스트"""
    client.post(
        "/auth/register",
        json={"username": "wrongpass", "email": "wrong@example.com", "password": "Correct123"}
    )

    response = client.post(
        "/auth/login",
        data={"username": "wrongpass", "password": "Wrong123"}
    )
    assert response.status_code == 401
    assert "잘못된" in response.json()["detail"]

def test_access_protected_endpoint():
    """보호된 엔드포인트 접근 테스트"""
    # 사용자 등록 및 로그인
    client.post(
        "/auth/register",
        json={"username": "protected", "email": "protected@example.com", "password": "Pass123"}
    )
    login_response = client.post(
        "/auth/login",
        data={"username": "protected", "password": "Pass123"}
    )
    token = login_response.json()["access_token"]

    # 보호된 엔드포인트 접근
    response = client.get(
        "/users/me",
        headers={"Authorization": f"Bearer {token}"}
    )
    assert response.status_code == 200
    data = response.json()
    assert data["username"] == "protected"

def test_access_without_token():
    """토큰 없이 보호된 엔드포인트 접근 테스트"""
    response = client.get("/users/me")
    assert response.status_code == 401
```

## Best Practices

### 보안
- ✅ **SECRET_KEY는 환경 변수로**: 코드에 하드코딩 금지
- ✅ **HTTPS 사용 필수**: 토큰이 평문으로 전송됨
- ✅ **짧은 만료 시간**: 15-30분 권장
- ✅ **bcrypt 사용**: 비밀번호 해싱에 bcrypt 또는 argon2
- ✅ **비밀번호 검증**: 최소 길이, 복잡도 요구

### 토큰 관리
- ✅ **페이로드 최소화**: 민감한 정보 제외
- ✅ **만료 시간 검증**: `exp` 클레임 확인
- ✅ **리프레시 토큰 사용**: 장기 세션
- ✅ **토큰 블랙리스트**: 로그아웃 시 토큰 무효화

### 코드 품질
- ✅ **의존성 주입**: `Depends(get_current_user)`
- ✅ **에러 처리**: 명확한 에러 메시지
- ✅ **Type Hints**: 모든 함수에 타입 지정
- ✅ **테스트 커버리지**: 80% 이상

## 주의사항

### 절대 금지
- ❌ **SECRET_KEY를 git에 커밋하지 마세요**
- ❌ **JWT에 비밀번호 저장 금지** (JWT는 디코딩 가능)
- ❌ **HTTP로 JWT 전송 금지** (반드시 HTTPS)
- ❌ **만료 시간을 너무 길게 설정 금지**

### 권장 사항
- ⚠️ **프로덕션에서는 Redis로 토큰 관리**
- ⚠️ **Rate Limiting 추가**: 무차별 대입 공격 방지
- ⚠️ **2FA (Two-Factor Auth) 고려**: 보안 강화
- ⚠️ **감사 로그**: 로그인 시도 기록

## 관련 예제

- [리프레시 토큰](/ko/examples/authentication/refresh-tokens) - 장기 세션 관리
- [OAuth2 인증](/ko/examples/authentication/oauth2) - 소셜 로그인
- [RBAC](/ko/examples/authentication/rbac) - 권한 관리
- [속도 제한](/ko/examples/security/rate-limiting) - API 남용 방지

## 참고 자료

- [FastAPI Security](https://fastapi.tiangolo.com/tutorial/security/)
- [JWT.io](https://jwt.io/) - JWT 디버거
- [OWASP Authentication Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)
