---
title: "입력 검증"
category: "security"
difficulty: "초급"
tags: [pydantic, validation, security, fastapi]
---

# 입력 검증

## 개요

Pydantic을 사용하여 모든 사용자 입력을 검증하고 보안 위협을 방지합니다.

## 기본 검증

```python
from pydantic import BaseModel, EmailStr, Field, validator

class UserCreate(BaseModel):
    username: str = Field(..., min_length=3, max_length=50, pattern="^[a-zA-Z0-9_]+$")
    email: EmailStr
    age: int = Field(..., ge=0, le=150)

    @validator('username')
    def username_alphanumeric(cls, v):
        """사용자명은 영숫자와 언더스코어만"""
        if not v.replace('_', '').isalnum():
            raise ValueError('영숫자와 언더스코어만 허용됩니다')
        return v
```

## 고급 검증

```python
from pydantic import validator

class PasswordUpdate(BaseModel):
    old_password: str
    new_password: str = Field(..., min_length=8)
    confirm_password: str

    @validator('new_password')
    def password_strength(cls, v):
        """비밀번호 강도 검증"""
        if not any(char.isdigit() for char in v):
            raise ValueError('최소 1개의 숫자 필요')
        if not any(char.isupper() for char in v):
            raise ValueError('최소 1개의 대문자 필요')
        return v

    @validator('confirm_password')
    def passwords_match(cls, v, values):
        """비밀번호 일치 확인"""
        if 'new_password' in values and v != values['new_password']:
            raise ValueError('비밀번호가 일치하지 않습니다')
        return v
```

## 관련 예제

- [에러 처리](/ko/examples/rest-api/error-handling)
- [SQL 인젝션 방지](/ko/examples/security/sql-injection-prevention)
