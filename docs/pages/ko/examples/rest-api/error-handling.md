---
title: "에러 처리 & 검증"
category: "rest-api"
difficulty: "초급"
tags: [fastapi, error-handling, validation, pydantic, exception]
---

# 에러 처리 & 검증

## 개요

안전하고 사용자 친화적인 에러 처리 시스템을 구현합니다. Pydantic 검증, HTTPException, 커스텀 예외 핸들러를 활용합니다.

## 사용 사례

- 잘못된 요청 데이터 처리
- 리소스를 찾을 수 없을 때
- 권한 부족 처리
- 서버 내부 에러 처리
- 일관된 에러 응답 형식

## 완전한 코드 예제

### 1. 커스텀 예외 (`exceptions.py`)

```python
# SPEC: API-030 - 커스텀 예외 정의
# @TAG:API-030

from fastapi import HTTPException, status

class BaseAPIException(HTTPException):
    """기본 API 예외"""

    def __init__(self, detail: str = None):
        super().__init__(
            status_code=self.status_code,
            detail=detail or self.detail
        )

class ResourceNotFoundException(BaseAPIException):
    """리소스를 찾을 수 없음"""
    status_code = status.HTTP_404_NOT_FOUND
    detail = "요청한 리소스를 찾을 수 없습니다"

class DuplicateResourceException(BaseAPIException):
    """중복 리소스"""
    status_code = status.HTTP_409_CONFLICT
    detail = "이미 존재하는 리소스입니다"

class ValidationException(BaseAPIException):
    """검증 실패"""
    status_code = status.HTTP_422_UNPROCESSABLE_ENTITY
    detail = "입력 데이터 검증에 실패했습니다"

class UnauthorizedException(BaseAPIException):
    """인증 실패"""
    status_code = status.HTTP_401_UNAUTHORIZED
    detail = "인증이 필요합니다"

class ForbiddenException(BaseAPIException):
    """권한 부족"""
    status_code = status.HTTP_403_FORBIDDEN
    detail = "접근 권한이 없습니다"

class RateLimitException(BaseAPIException):
    """요청 제한 초과"""
    status_code = status.HTTP_429_TOO_MANY_REQUESTS
    detail = "요청 한도를 초과했습니다. 잠시 후 다시 시도해주세요"
```

### 2. 에러 응답 스키마 (`schemas.py`)

```python
# SPEC: API-031 - 에러 응답 스키마
# @TAG:API-031

from pydantic import BaseModel, Field
from typing import Optional, List, Dict, Any
from datetime import datetime

class ErrorDetail(BaseModel):
    """에러 상세 정보"""
    loc: Optional[List[str]] = Field(None, description="에러 위치")
    msg: str = Field(..., description="에러 메시지")
    type: str = Field(..., description="에러 타입")

class ErrorResponse(BaseModel):
    """표준 에러 응답"""
    status_code: int = Field(..., description="HTTP 상태 코드")
    message: str = Field(..., description="에러 메시지")
    details: Optional[List[ErrorDetail]] = Field(None, description="상세 에러 목록")
    timestamp: datetime = Field(default_factory=datetime.utcnow, description="발생 시각")
    path: Optional[str] = Field(None, description="요청 경로")

    class Config:
        json_schema_extra = {
            "example": {
                "status_code": 404,
                "message": "사용자를 찾을 수 없습니다",
                "details": None,
                "timestamp": "2024-01-01T00:00:00",
                "path": "/users/999"
            }
        }

class ValidationErrorResponse(ErrorResponse):
    """검증 에러 응답"""
    details: List[ErrorDetail] = Field(..., description="검증 에러 목록")

    class Config:
        json_schema_extra = {
            "example": {
                "status_code": 422,
                "message": "입력 데이터 검증 실패",
                "details": [
                    {
                        "loc": ["body", "email"],
                        "msg": "유효한 이메일 주소를 입력해주세요",
                        "type": "value_error.email"
                    }
                ],
                "timestamp": "2024-01-01T00:00:00",
                "path": "/users/"
            }
        }
```

### 3. 예외 핸들러 (`handlers.py`)

```python
# SPEC: API-032 - 예외 핸들러
# @TAG:API-032

from fastapi import Request, status
from fastapi.responses import JSONResponse
from fastapi.exceptions import RequestValidationError
from pydantic import ValidationError
import logging

from exceptions import BaseAPIException
from schemas import ErrorResponse, ErrorDetail

logger = logging.getLogger(__name__)

async def custom_exception_handler(
    request: Request,
    exc: BaseAPIException
) -> JSONResponse:
    """커스텀 예외 핸들러"""

    error_response = ErrorResponse(
        status_code=exc.status_code,
        message=exc.detail,
        path=str(request.url.path)
    )

    # 로깅
    logger.warning(
        f"API Exception: {exc.status_code} - {exc.detail} - Path: {request.url.path}"
    )

    return JSONResponse(
        status_code=exc.status_code,
        content=error_response.model_dump()
    )

async def validation_exception_handler(
    request: Request,
    exc: RequestValidationError
) -> JSONResponse:
    """Pydantic 검증 에러 핸들러"""

    # Pydantic 에러를 ErrorDetail 형식으로 변환
    errors = []
    for error in exc.errors():
        errors.append(ErrorDetail(
            loc=error.get("loc"),
            msg=error.get("msg"),
            type=error.get("type")
        ))

    error_response = ErrorResponse(
        status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
        message="입력 데이터 검증에 실패했습니다",
        details=errors,
        path=str(request.url.path)
    )

    # 로깅
    logger.warning(
        f"Validation Error: {request.url.path} - {len(errors)} errors"
    )

    return JSONResponse(
        status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
        content=error_response.model_dump()
    )

async def general_exception_handler(
    request: Request,
    exc: Exception
) -> JSONResponse:
    """일반 예외 핸들러 (500 Internal Server Error)"""

    error_response = ErrorResponse(
        status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
        message="서버 내부 오류가 발생했습니다",
        path=str(request.url.path)
    )

    # 로깅 (스택 트레이스 포함)
    logger.error(
        f"Internal Server Error: {request.url.path}",
        exc_info=exc
    )

    return JSONResponse(
        status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
        content=error_response.model_dump()
    )
```

### 4. FastAPI 앱 설정 (`main.py`)

```python
# SPEC: API-033 - 에러 처리 통합
# @TAG:API-033

from fastapi import FastAPI, Depends, HTTPException, status
from fastapi.exceptions import RequestValidationError
from sqlalchemy.orm import Session
from pydantic import BaseModel, EmailStr, Field, validator

import exceptions
import handlers
from database import get_db
import models

app = FastAPI(
    title="Error Handling API",
    description="체계적인 에러 처리 시스템"
)

# 예외 핸들러 등록
app.add_exception_handler(
    exceptions.BaseAPIException,
    handlers.custom_exception_handler
)
app.add_exception_handler(
    RequestValidationError,
    handlers.validation_exception_handler
)
app.add_exception_handler(
    Exception,
    handlers.general_exception_handler
)

# ===== 스키마 =====

class UserCreate(BaseModel):
    """사용자 생성 스키마 (검증 포함)"""
    email: EmailStr = Field(..., description="이메일 주소")
    name: str = Field(..., min_length=2, max_length=50, description="이름")
    age: int = Field(..., ge=0, le=150, description="나이")
    password: str = Field(..., min_length=8, description="비밀번호 (최소 8자)")

    @validator('password')
    def validate_password(cls, v):
        """비밀번호 강도 검증"""
        if not any(char.isdigit() for char in v):
            raise ValueError('비밀번호는 최소 1개의 숫자를 포함해야 합니다')
        if not any(char.isupper() for char in v):
            raise ValueError('비밀번호는 최소 1개의 대문자를 포함해야 합니다')
        return v

    @validator('name')
    def validate_name(cls, v):
        """이름 검증"""
        if v.isdigit():
            raise ValueError('이름은 숫자만으로 구성될 수 없습니다')
        return v.strip()

class UserResponse(BaseModel):
    """사용자 응답 스키마"""
    id: int
    email: str
    name: str
    age: int

    class Config:
        from_attributes = True

# ===== API 엔드포인트 =====

@app.post("/users/", response_model=UserResponse, status_code=status.HTTP_201_CREATED)
def create_user(user: UserCreate, db: Session = Depends(get_db)):
    """
    사용자 생성

    **검증 규칙**:
    - 이메일: 유효한 이메일 형식
    - 이름: 2-50자, 숫자만으로 구성 불가
    - 나이: 0-150
    - 비밀번호: 최소 8자, 숫자 1개 이상, 대문자 1개 이상

    **에러 응답**:
    - 422: 검증 실패
    - 409: 이미 존재하는 이메일
    - 500: 서버 에러
    """
    # 이메일 중복 체크
    existing = db.query(models.User).filter(
        models.User.email == user.email
    ).first()

    if existing:
        raise exceptions.DuplicateResourceException(
            detail=f"이메일 '{user.email}'은 이미 사용 중입니다"
        )

    # 사용자 생성
    db_user = models.User(**user.model_dump(exclude={'password'}))
    # 실제로는 비밀번호 해시화 필요
    db.add(db_user)
    db.commit()
    db.refresh(db_user)

    return db_user

@app.get("/users/{user_id}", response_model=UserResponse)
def get_user(user_id: int, db: Session = Depends(get_db)):
    """
    사용자 조회

    **에러 응답**:
    - 404: 사용자를 찾을 수 없음
    """
    user = db.query(models.User).filter(models.User.id == user_id).first()

    if not user:
        raise exceptions.ResourceNotFoundException(
            detail=f"ID {user_id}인 사용자를 찾을 수 없습니다"
        )

    return user

@app.delete("/users/{user_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_user(user_id: int, db: Session = Depends(get_db)):
    """
    사용자 삭제

    **권한**: 관리자만 가능

    **에러 응답**:
    - 401: 인증 실패
    - 403: 권한 부족
    - 404: 사용자를 찾을 수 없음
    """
    # 권한 체크 (간단한 예시)
    # 실제로는 JWT 토큰 검증 필요
    is_admin = False  # 실제 권한 체크 로직

    if not is_admin:
        raise exceptions.ForbiddenException(
            detail="관리자만 사용자를 삭제할 수 있습니다"
        )

    user = db.query(models.User).filter(models.User.id == user_id).first()

    if not user:
        raise exceptions.ResourceNotFoundException()

    db.delete(user)
    db.commit()

@app.get("/test/error")
def test_internal_error():
    """서버 에러 테스트"""
    # 의도적으로 에러 발생
    raise Exception("테스트 에러입니다")

@app.get("/test/rate-limit")
def test_rate_limit():
    """요청 제한 테스트"""
    # 실제로는 Redis 등으로 요청 횟수 체크
    raise exceptions.RateLimitException()
```

## 단계별 설명

### 1. 에러 응답 예시

```bash
# 검증 실패 (422)
curl -X POST "http://localhost:8000/users/" \
  -H "Content-Type: application/json" \
  -d '{"email":"invalid-email","name":"A","age":200,"password":"weak"}'

# 응답
{
  "status_code": 422,
  "message": "입력 데이터 검증에 실패했습니다",
  "details": [
    {
      "loc": ["body", "email"],
      "msg": "유효한 이메일 주소를 입력해주세요",
      "type": "value_error.email"
    },
    {
      "loc": ["body", "name"],
      "msg": "최소 길이는 2입니다",
      "type": "value_error.any_str.min_length"
    }
  ],
  "timestamp": "2024-01-01T00:00:00",
  "path": "/users/"
}
```

### 2. 리소스 없음 (404)

```bash
curl "http://localhost:8000/users/999"

# 응답
{
  "status_code": 404,
  "message": "ID 999인 사용자를 찾을 수 없습니다",
  "details": null,
  "timestamp": "2024-01-01T00:00:00",
  "path": "/users/999"
}
```

## 테스트 코드

```python
# SPEC: TEST-API-030 - 에러 처리 테스트
# @TAG:TEST-API-030

import pytest
from fastapi.testclient import TestClient
from main import app

client = TestClient(app)

def test_validation_error_invalid_email():
    """잘못된 이메일 검증 테스트"""
    response = client.post(
        "/users/",
        json={
            "email": "invalid-email",
            "name": "John Doe",
            "age": 30,
            "password": "Password1"
        }
    )

    assert response.status_code == 422
    data = response.json()
    assert "이메일" in data["message"] or "email" in str(data["details"])

def test_validation_error_short_name():
    """짧은 이름 검증 테스트"""
    response = client.post(
        "/users/",
        json={
            "email": "test@example.com",
            "name": "A",  # 너무 짧음
            "age": 30,
            "password": "Password1"
        }
    )

    assert response.status_code == 422
    data = response.json()
    assert any("name" in str(detail.get("loc", [])) for detail in data["details"])

def test_validation_error_weak_password():
    """약한 비밀번호 검증 테스트"""
    response = client.post(
        "/users/",
        json={
            "email": "test@example.com",
            "name": "John Doe",
            "age": 30,
            "password": "weak"  # 대문자, 숫자 없음
        }
    )

    assert response.status_code == 422

def test_duplicate_resource():
    """중복 리소스 에러 테스트"""
    user_data = {
        "email": "duplicate@example.com",
        "name": "John Doe",
        "age": 30,
        "password": "Password1"
    }

    # 첫 번째 생성
    response1 = client.post("/users/", json=user_data)
    assert response1.status_code == 201

    # 두 번째 생성 시도
    response2 = client.post("/users/", json=user_data)
    assert response2.status_code == 409
    assert "이미 사용 중" in response2.json()["message"]

def test_resource_not_found():
    """리소스 없음 에러 테스트"""
    response = client.get("/users/99999")
    assert response.status_code == 404
    data = response.json()
    assert "찾을 수 없습니다" in data["message"]

def test_forbidden_error():
    """권한 부족 에러 테스트"""
    response = client.delete("/users/1")
    assert response.status_code == 403
    assert "권한" in response.json()["message"]

def test_internal_server_error():
    """서버 에러 테스트"""
    response = client.get("/test/error")
    assert response.status_code == 500
    assert "서버 내부 오류" in response.json()["message"]

def test_error_response_structure():
    """에러 응답 구조 테스트"""
    response = client.get("/users/999")
    data = response.json()

    # 필수 필드 확인
    assert "status_code" in data
    assert "message" in data
    assert "timestamp" in data
    assert "path" in data
```

## Best Practices

### 에러 메시지
- ✅ **사용자 친화적**: 기술 용어 지양, 명확한 한국어
- ✅ **구체적**: "에러 발생" 대신 "이메일 형식이 올바르지 않습니다"
- ✅ **해결 방법 제시**: "다시 시도해주세요"
- ✅ **일관된 형식**: ErrorResponse 스키마 사용

### 보안
- ✅ **민감 정보 숨김**: 내부 구조, SQL 쿼리 노출 금지
- ✅ **스택 트레이스 제거**: 프로덕션에서는 로그만 기록
- ✅ **에러 레벨 구분**: 로그 레벨 적절히 사용 (INFO, WARNING, ERROR)

### 로깅
- ✅ **구조화된 로깅**: JSON 형식
- ✅ **컨텍스트 포함**: 요청 경로, 사용자 ID, 타임스탬프
- ✅ **에러 추적**: Sentry, Datadog 등 통합

## 주의사항

- ❌ **에러 메시지에 SQL 쿼리 포함 금지**
- ❌ **스택 트레이스를 사용자에게 노출 금지**
- ❌ **비밀번호, 토큰 등 민감 정보 로깅 금지**
- ❌ **에러 처리 누락**: 모든 예외 처리 필수

## 관련 예제

- [기본 CRUD 작업](/ko/examples/rest-api/basic-crud)
- [입력 검증](/ko/examples/security/input-validation)
- [JWT 인증](/ko/examples/authentication/jwt-basic)

## 참고 자료

- [FastAPI Error Handling](https://fastapi.tiangolo.com/tutorial/handling-errors/)
- [Pydantic Validation](https://docs.pydantic.dev/latest/concepts/validators/)
