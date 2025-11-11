---
title: "역할 기반 접근 제어 (RBAC)"
category: "authentication"
difficulty: "고급"
tags: [rbac, permissions, authorization, security]
---

# 역할 기반 접근 제어 (RBAC)

## 개요

Role-Based Access Control (RBAC)을 구현하여 사용자 역할에 따라 API 접근을 제어합니다.

## 핵심 개념

```python
User → Role → Permissions

예시:
- admin: 모든 권한
- editor: read, write
- viewer: read만
```

## 간단한 구현

```python
# SPEC: AUTH-030 - RBAC 구현
# @TAG:AUTH-030

from enum import Enum
from fastapi import Depends, HTTPException, status

class Role(str, Enum):
    ADMIN = "admin"
    EDITOR = "editor"
    VIEWER = "viewer"

def require_role(required_role: Role):
    """역할 확인 데코레이터"""
    def role_checker(current_user: User = Depends(get_current_user)):
        if current_user.role != required_role:
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail="권한이 없습니다"
            )
        return current_user
    return role_checker

@app.delete("/posts/{post_id}")
def delete_post(
    post_id: int,
    user: User = Depends(require_role(Role.ADMIN))
):
    """관리자만 삭제 가능"""
    # ...
```

## 관련 예제

- [JWT 인증](/ko/examples/authentication/jwt-basic)
