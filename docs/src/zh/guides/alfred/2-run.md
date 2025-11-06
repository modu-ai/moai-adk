---
title: 2-run 命令指南
description: 学习如何使用 Alfred 的 2-run 命令执行完整的测试驱动开发（TDD）流程
---

# 2-run 命令指南

`/alfred:2-run` 命令是 MoAI-ADK 开发执行阶段的核心工具，负责基于 SPEC 执行完整的测试驱动开发（TDD）流程，确保代码质量和可测试性。

## 命令概览

### 基本语法
```bash
/alfred:2-run SPEC-ID
```

### 命令目的
- 基于 SPEC 执行 TDD 开发流程
- 实现高质量的代码
- 确保测试覆盖率
- 应用 TRUST 5 原则
- 生成完整的实现方案

### 触发的代理
- **code-builder**：主导 TDD 实现
  - **implementation-planner**：实现策略制定
  - **tdd-implementer**：TDD 循环执行
- **quality-gate**：质量保证和验证
- **trust-checker**：TRUST 5 原则验证
- **domain-experts**：领域专业知识支持

---

## TDD 工作流程详解

### 阶段 1：Implementation Planning（实现规划）

#### SPEC 分析
Alfred 首先深入分析 SPEC 文档：

```python
def analyze_spec(spec_id):
    spec_content = read_spec_file(f".moai/specs/SPEC-{spec_id}/spec.md")

    analysis = {
        "requirements": extract_requirements(spec_content),
        "acceptance_criteria": extract_acceptance_criteria(spec_content),
        "constraints": extract_constraints(spec_content),
        "dependencies": extract_dependencies(spec_content),
        "risks": extract_risks(spec_content),
        "technical_needs": identify_technical_needs(spec_content)
    }

    return analysis
```

#### 架构设计
基于 SPEC 分析，Alfred 设计合适的架构：

```yaml
架构设计示例:
SPEC: USER-AUTH-001 (用户认证系统)

技术栈选择:
- 后端框架: FastAPI (高性能，自动文档生成)
- 数据库: PostgreSQL (ACID 特性，JSON 支持)
- 认证: JWT (无状态，易于扩展)
- 密码加密: bcrypt (安全，慢哈希)
- 验证库: Pydantic (类型安全，自动验证)

目录结构:
src/
├── auth/
│   ├── __init__.py
│   ├── models.py      # 用户数据模型
│   ├── schemas.py     # API 请求/响应模式
│   ├── services.py    # 业务逻辑
│   ├── api.py         # API 端点
│   └── repository.py  # 数据访问层
├── core/
│   ├── __init__.py
│   ├── security.py    # 安全工具
│   ├── config.py      # 配置管理
│   └── exceptions.py  # 自定义异常
└── tests/
    ├── test_auth_api.py
    ├── test_auth_services.py
    └── test_auth_repository.py

设计原则:
- 单一职责原则：每个类/函数只有一个职责
- 依赖注入：便于测试和模块化
- 接口隔离：清晰的模块边界
- 开闭原则：易于扩展新功能
```

#### 技术选型
Alfred 会推荐最适合的技术栈：

```yaml
技术选型决策树:

Web 框架选择:
FastAPI:
  ✅ 自动 API 文档生成
  ✅ 类型提示支持
  ✅ 高性能异步支持
  ✅ 易于测试
  ✅ 丰富的验证功能

数据库选择:
PostgreSQL:
  ✅ 强一致性保证
  ✅ JSON 数据类型支持
  ✅ 丰富的索引类型
  ✅ 成熟的生态系统
  ✅ 良好的 Python 支持

认证方案:
JWT:
  ✅ 无状态认证
  ✅ 易于分布式部署
  ✅ 标准化实现
  ✅ 移动端友好
  ✅ 细粒度权限控制
```

### 阶段 2：TDD 循环执行

#### 🔴 RED 阶段：编写失败测试

##### 测试策略制定
Alfred 首先制定全面的测试策略：

```python
def design_test_strategy(spec_analysis):
    strategy = {
        "unit_tests": design_unit_tests(spec_analysis),
        "integration_tests": design_integration_tests(spec_analysis),
        "api_tests": design_api_tests(spec_analysis),
        "edge_cases": identify_edge_cases(spec_analysis),
        "security_tests": design_security_tests(spec_analysis),
        "performance_tests": design_performance_tests(spec_analysis)
    }
    return strategy
```

##### 测试用例生成
基于 SPEC 的验收标准生成测试用例：

```python
# `@TEST:USER-AUTH-001 | SPEC: SPEC-USER-AUTH-001.md

import pytest
from fastapi.testclient import TestClient
from src.auth.api import app
from src.auth.models import User
from src.auth.services import AuthService
from unittest.mock import patch

client = TestClient(app)

class TestUserRegistration:
    """用户注册功能测试"""

    def test_register_with_valid_data_should_create_user(self):
        """当提供有效数据时，系统必须创建用户并返回 201"""
        user_data = {
            "email": "test@example.com",
            "password": "SecurePass123!",
            "full_name": "Test User"
        }

        response = client.post("/auth/register", json=user_data)

        assert response.status_code == 201
        data = response.json()
        assert data["email"] == user_data["email"]
        assert data["full_name"] == user_data["full_name"]
        assert "id" in data
        assert "password" not in data  # 确保密码不在响应中

    def test_register_with_duplicate_email_should_return_400(self):
        """当使用重复邮箱时，系统必须返回 400 错误"""
        # 先创建一个用户
        existing_email = "existing@example.com"
        create_test_user(email=existing_email)

        user_data = {
            "email": existing_email,
            "password": "SecurePass123!",
            "full_name": "Another User"
        }

        response = client.post("/auth/register", json=user_data)

        assert response.status_code == 400
        assert "email already exists" in response.json()["detail"].lower()

    def test_register_with_invalid_email_should_return_422(self):
        """当邮箱格式无效时，系统必须返回 422 错误"""
        user_data = {
            "email": "invalid-email",
            "password": "SecurePass123!",
            "full_name": "Test User"
        }

        response = client.post("/auth/register", json=user_data)

        assert response.status_code == 422
        assert "email" in response.json()["detail"][0]["loc"]

    def test_register_with_weak_password_should_return_422(self):
        """当密码强度不足时，系统必须返回 422 错误"""
        user_data = {
            "email": "test@example.com",
            "password": "weak",  # 太短的密码
            "full_name": "Test User"
        }

        response = client.post("/auth/register", json=user_data)

        assert response.status_code == 422
        assert "password" in response.json()["detail"][0]["loc"]

class TestUserLogin:
    """用户登录功能测试"""

    def test_login_with_valid_credentials_should_return_token(self):
        """当提供有效凭证时，系统必须返回 JWT 令牌"""
        # 创建测试用户
        user = create_test_user(
            email="test@example.com",
            password="CorrectPass123!"
        )

        login_data = {
            "email": "test@example.com",
            "password": "CorrectPass123!"
        }

        response = client.post("/auth/login", json=login_data)

        assert response.status_code == 200
        data = response.json()
        assert "access_token" in data
        assert data["token_type"] == "bearer"
        assert isinstance(data["expires_in"], int)

    def test_login_with_invalid_email_should_return_401(self):
        """当邮箱无效时，系统必须返回 401 错误"""
        login_data = {
            "email": "nonexistent@example.com",
            "password": "SomePass123!"
        }

        response = client.post("/auth/login", json=login_data)

        assert response.status_code == 401
        assert "invalid credentials" in response.json()["detail"].lower()

    def test_login_with_invalid_password_should_return_401(self):
        """当密码错误时，系统必须返回 401 错误"""
        user = create_test_user(
            email="test@example.com",
            password="CorrectPass123!"
        )

        login_data = {
            "email": "test@example.com",
            "password": "WrongPass123!"
        }

        response = client.post("/auth/login", json=login_data)

        assert response.status_code == 401
        assert "invalid credentials" in response.json()["detail"].lower()

class TestTokenValidation:
    """令牌验证功能测试"""

    def test_valid_token_should_allow_access(self):
        """当令牌有效时，系统必须允许访问受保护资源"""
        user = create_and_login_user()
        token = user["access_token"]

        headers = {"Authorization": f"Bearer {token}"}
        response = client.get("/auth/me", headers=headers)

        assert response.status_code == 200
        data = response.json()
        assert data["email"] == user["email"]

    def test_invalid_token_should_return_401(self):
        """当令牌无效时，系统必须返回 401 错误"""
        headers = {"Authorization": "Bearer invalid_token"}
        response = client.get("/auth/me", headers=headers)

        assert response.status_code == 401
        assert "invalid token" in response.json()["detail"].lower()

    def test_expired_token_should_return_401(self):
        """当令牌过期时，系统必须返回 401 错误"""
        # 创建过期的令牌
        expired_token = create_expired_token()

        headers = {"Authorization": f"Bearer {expired_token}"}
        response = client.get("/auth/me", headers=headers)

        assert response.status_code == 401
        assert "token has expired" in response.json()["detail"].lower()

# 测试辅助函数
def create_test_user(email: str, password: str) -> User:
    """创建测试用户"""
    user_service = AuthService()
    return user_service.create_user(email, password)

def create_and_login_user() -> dict:
    """创建并登录测试用户"""
    user = create_test_user("test@example.com", "CorrectPass123!")
    auth_service = AuthService()
    return auth_service.authenticate_user("test@example.com", "CorrectPass123!")

def create_expired_token() -> str:
    """创建过期的 JWT 令牌"""
    # 实现创建过期令牌的逻辑
    pass
```

##### 测试执行验证
Alfred 验证测试能够正确失败：

```bash
# 运行测试（预期全部失败）
pytest tests/test_auth.py -v

# 输出示例：
# test_register_with_valid_data_should_create_user FAILED
# test_register_with_duplicate_email_should_return_400 FAILED
# test_login_with_valid_credentials_should_return_token FAILED
# ...
# 15 tests failed, 0 passed
```

**Git 提交 RED 阶段**：
```bash
git add tests/test_auth.py
git commit -m "🔴 test(USER-AUTH-001): add failing authentication tests"
```

#### 🟢 GREEN 阶段：最小实现

##### 实现策略
Alfred 制定最简单的实现策略：

```python
def design_minimal_implementation(test_requirements):
    """设计最小实现策略"""
    strategy = {
        "models": design_data_models(test_requirements),
        "schemas": design_api_schemas(test_requirements),
        "services": design_business_services(test_requirements),
        "repositories": design_data_access(test_requirements),
        "api_endpoints": design_api_endpoints(test_requirements)
    }
    return strategy
```

##### 数据模型实现
```python
# `@CODE:USER-AUTH-001:MODEL | SPEC: SPEC-USER-AUTH-001.md | TEST: tests/test_auth.py

from sqlalchemy import Column, Integer, String, DateTime, Boolean
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.sql import func
from datetime import datetime

Base = declarative_base()

class User(Base):
    """@CODE:USER-AUTH-001:MODEL - 用户数据模型"""
    __tablename__ = "users"

    id = Column(Integer, primary_key=True, index=True)
    email = Column(String(255), unique=True, index=True, nullable=False)
    password_hash = Column(String(255), nullable=False)
    full_name = Column(String(255), nullable=False)
    is_active = Column(Boolean, default=True)
    is_verified = Column(Boolean, default=False)
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), onupdate=func.now())

    def verify_password(self, password: str) -> bool:
        """验证密码"""
        from src.core.security import verify_password
        return verify_password(password, self.password_hash)

    @property
    def is_authenticated(self) -> bool:
        """用户是否已认证"""
        return self.is_active and self.is_verified
```

##### API 模式定义
```python
# `@CODE:USER-AUTH-001:SCHEMA | SPEC: SPEC-USER-AUTH-001.md | TEST: tests/test_auth.py

from pydantic import BaseModel, EmailStr, validator
from typing import Optional

class UserRegisterRequest(BaseModel):
    """@CODE:USER-AUTH-001:SCHEMA - 用户注册请求"""
    email: EmailStr
    password: str
    full_name: str

    @validator('password')
    def validate_password(cls, v):
        """验证密码强度"""
        if len(v) < 8:
            raise ValueError('Password must be at least 8 characters long')
        if not any(c.isupper() for c in v):
            raise ValueError('Password must contain at least one uppercase letter')
        if not any(c.islower() for c in v):
            raise ValueError('Password must contain at least one lowercase letter')
        if not any(c.isdigit() for c in v):
            raise ValueError('Password must contain at least one digit')
        return v

class UserLoginRequest(BaseModel):
    """@CODE:USER-AUTH-001:SCHEMA - 用户登录请求"""
    email: EmailStr
    password: str

class UserResponse(BaseModel):
    """@CODE:USER-AUTH-001:SCHEMA - 用户响应"""
    id: int
    email: str
    full_name: str
    is_active: bool
    is_verified: bool
    created_at: datetime

    class Config:
        from_attributes = True

class TokenResponse(BaseModel):
    """@CODE:USER-AUTH-001:SCHEMA - 令牌响应"""
    access_token: str
    token_type: str
    expires_in: int
```

##### 业务服务实现
```python
# `@CODE:USER-AUTH-001:SERVICE | SPEC: SPEC-USER-AUTH-001.md | TEST: tests/test_auth.py

from sqlalchemy.orm import Session
from src.auth.models import User
from src.auth.schemas import UserRegisterRequest, UserLoginRequest
from src.core.security import hash_password, verify_password, create_access_token
from src.core.exceptions import DuplicateUserError, InvalidCredentialsError
from typing import Optional

class AuthService:
    """@CODE:USER-AUTH-001:SERVICE - 认证业务服务"""

    def __init__(self, db: Session):
        self.db = db

    def create_user(self, user_data: UserRegisterRequest) -> User:
        """创建新用户"""
        # 检查邮箱是否已存在
        existing_user = self.db.query(User).filter(User.email == user_data.email).first()
        if existing_user:
            raise DuplicateUserError("Email already exists")

        # 创建用户
        hashed_password = hash_password(user_data.password)
        db_user = User(
            email=user_data.email,
            password_hash=hashed_password,
            full_name=user_data.full_name
        )

        self.db.add(db_user)
        self.db.commit()
        self.db.refresh(db_user)

        return db_user

    def authenticate_user(self, email: str, password: str) -> dict:
        """认证用户并返回令牌"""
        user = self.db.query(User).filter(User.email == email).first()

        if not user or not user.verify_password(password):
            raise InvalidCredentialsError("Invalid email or password")

        if not user.is_active:
            raise InvalidCredentialsError("Account is inactive")

        # 创建访问令牌
        token_data = {
            "sub": user.email,
            "user_id": user.id,
            "is_active": user.is_active,
            "is_verified": user.is_verified
        }

        access_token = create_access_token(data=token_data)

        return {
            "access_token": access_token,
            "token_type": "bearer",
            "expires_in": 3600,  # 1 hour
            "user": {
                "id": user.id,
                "email": user.email,
                "full_name": user.full_name
            }
        }

    def get_user_by_id(self, user_id: int) -> Optional[User]:
        """根据 ID 获取用户"""
        return self.db.query(User).filter(User.id == user_id).first()
```

##### API 端点实现
```python
# `@CODE:USER-AUTH-001:API | SPEC: SPEC-USER-AUTH-001.md | TEST: tests/test_auth.py

from fastapi import APIRouter, Depends, HTTPException, status
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
from sqlalchemy.orm import Session
from src.database import get_db
from src.auth.services import AuthService
from src.auth.schemas import UserRegisterRequest, UserLoginRequest, UserResponse, TokenResponse
from src.core.exceptions import DuplicateUserError, InvalidCredentialsError

router = APIRouter(prefix="/auth", tags=["authentication"])
security = HTTPBearer()

@router.post("/register", response_model=UserResponse, status_code=status.HTTP_201_CREATED)
def register(user_data: UserRegisterRequest, db: Session = Depends(get_db)):
    """@CODE:USER-AUTH-001:API - 用户注册端点"""
    try:
        auth_service = AuthService(db)
        user = auth_service.create_user(user_data)
        return UserResponse.from_orm(user)
    except DuplicateUserError as e:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=str(e)
        )

@router.post("/login", response_model=TokenResponse)
def login(login_data: UserLoginRequest, db: Session = Depends(get_db)):
    """@CODE:USER-AUTH-001:API - 用户登录端点"""
    try:
        auth_service = AuthService(db)
        token_data = auth_service.authenticate_user(login_data.email, login_data.password)
        return TokenResponse(**token_data)
    except InvalidCredentialsError as e:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail=str(e),
            headers={"WWW-Authenticate": "Bearer"},
        )

@router.get("/me", response_model=UserResponse)
def get_current_user(
    credentials: HTTPAuthorizationCredentials = Depends(security),
    db: Session = Depends(get_db)
):
    """@CODE:USER-AUTH-001:API - 获取当前用户信息"""
    try:
        # 验证令牌并获取用户信息
        from src.core.security import verify_token
        token_data = verify_token(credentials.credentials)

        auth_service = AuthService(db)
        user = auth_service.get_user_by_id(token_data["user_id"])

        if not user:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail="User not found"
            )

        return UserResponse.from_orm(user)
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Invalid token",
            headers={"WWW-Authenticate": "Bearer"},
        )
```

##### 测试验证
```bash
# 运行测试（预期全部通过）
pytest tests/test_auth.py -v

# 输出示例：
# test_register_with_valid_data_should_create_user PASSED
# test_register_with_duplicate_email_should_return_400 PASSED
# test_login_with_valid_credentials_should_return_token PASSED
# ...
# 15 tests passed, 0 failed

# 测试覆盖率
pytest --cov=src.auth --cov-report=term-missing
# 输出示例：
# src/auth/models.py      100%  15/15
# src/auth/schemas.py     100%  25/25
# src/auth/services.py    100%  35/35
# src/auth/api.py         100%  20/20
# TOTAL                   100%  95/95
```

**Git 提交 GREEN 阶段**：
```bash
git add src/auth/ tests/test_auth.py
git commit -m "🟢 feat(USER-AUTH-001): implement authentication API"
```

#### ♻️ REFACTOR 阶段：代码改进

##### 代码质量分析
Alfred 分析代码质量并识别改进机会：

```python
def analyze_code_quality(implementation):
    analysis = {
        "complexity": calculate_complexity(implementation),
        "duplication": detect_duplication(implementation),
        "design_patterns": identify_design_opportunities(implementation),
        "performance": identify_performance_issues(implementation),
        "security": identify_security_improvements(implementation),
        "maintainability": assess_maintainability(implementation)
    }
    return analysis
```

##### 重构实施
基于分析结果进行代码重构：

###### 1. 抽象通用功能
```python
# `@CODE:USER-AUTH-001:REPOSITORY | SPEC: SPEC-USER-AUTH-001.md | TEST: tests/test_auth.py

from abc import ABC, abstractmethod
from typing import Generic, TypeVar, Type, List, Optional
from sqlalchemy.orm import Session

ModelType = TypeVar("ModelType", bound=Base)

class BaseRepository(Generic[ModelType], ABC):
    """@CODE:USER-AUTH-001:REPOSITORY - 基础仓库模式"""

    def __init__(self, db: Session, model: Type[ModelType]):
        self.db = db
        self.model = model

    def get_by_id(self, id: int) -> Optional[ModelType]:
        """根据 ID 获取实体"""
        return self.db.query(self.model).filter(self.model.id == id).first()

    def get_by_field(self, field: str, value: any) -> Optional[ModelType]:
        """根据字段值获取实体"""
        filter_kwargs = {field: value}
        return self.db.query(self.model).filter_by(**filter_kwargs).first()

    def create(self, obj_in: dict) -> ModelType:
        """创建新实体"""
        db_obj = self.model(**obj_in)
        self.db.add(db_obj)
        self.db.commit()
        self.db.refresh(db_obj)
        return db_obj

    def update(self, db_obj: ModelType, obj_in: dict) -> ModelType:
        """更新实体"""
        for field, value in obj_in.items():
            setattr(db_obj, field, value)
        self.db.add(db_obj)
        self.db.commit()
        self.db.refresh(db_obj)
        return db_obj

    def delete(self, id: int) -> ModelType:
        """删除实体"""
        obj = self.get_by_id(id)
        self.db.delete(obj)
        self.db.commit()
        return obj

class UserRepository(BaseRepository[User]):
    """@CODE:USER-AUTH-001:REPOSITORY - 用户仓库实现"""

    def __init__(self, db: Session):
        super().__init__(db, User)

    def get_by_email(self, email: str) -> Optional[User]:
        """根据邮箱获取用户"""
        return self.get_by_field("email", email)

    def get_active_users(self) -> List[User]:
        """获取活跃用户列表"""
        return self.db.query(self.model).filter(self.model.is_active == True).all()
```

###### 2. 改进错误处理
```python
# `@CODE:USER-AUTH-001:EXCEPTIONS | SPEC: SPEC-USER-AUTH-001.md | TEST: tests/test_auth.py

from src.core.exceptions import BaseException

class AuthenticationError(BaseException):
    """认证相关错误基类"""
    pass

class UserAlreadyExistsError(AuthenticationError):
    """用户已存在错误"""
    def __init__(self, email: str):
        self.email = email
        super().__init__(f"User with email {email} already exists")

class InvalidCredentialsError(AuthenticationError):
    """无效凭证错误"""
    def __init__(self, message: str = "Invalid email or password"):
        super().__init__(message)

class AccountInactiveError(AuthenticationError):
    """账户未激活错误"""
    def __init__(self, email: str):
        self.email = email
        super().__init__(f"Account {email} is inactive")

class TokenExpiredError(AuthenticationError):
    """令牌过期错误"""
    def __init__(self):
        super().__init__("Token has expired")

class InvalidTokenError(AuthenticationError):
    """无效令牌错误"""
    def __init__(self, message: str = "Invalid token"):
        super().__init__(message)
```

###### 3. 添加配置管理
```python
# `@CODE:USER-AUTH-001:CONFIG | SPEC: SPEC-USER-AUTH-001.md | TEST: tests/test_auth.py

from pydantic import BaseSettings
from typing import Optional

class AuthSettings(BaseSettings):
    """@CODE:USER-AUTH-001:CONFIG - 认证配置"""

    # JWT 配置
    secret_key: str
    algorithm: str = "HS256"
    access_token_expire_minutes: int = 60

    # 密码配置
    password_min_length: int = 8
    password_require_uppercase: bool = True
    password_require_lowercase: bool = True
    password_require_digits: bool = True
    password_require_special: bool = False

    # 邮箱配置
    email_from: Optional[str] = None
    email_verify_token_expire_minutes: int = 1440  # 24 hours

    # 安全配置
    max_login_attempts: int = 5
    lockout_duration_minutes: int = 15

    class Config:
        env_file = ".env"
        case_sensitive = True

# 全局配置实例
auth_settings = AuthSettings()
```

###### 4. 添加缓存支持
```python
# `@CODE:USER-AUTH-001:CACHE | SPEC: SPEC-USER-AUTH-001.md | TEST: tests/test_auth.py

import redis
import json
from typing import Optional, Any
from src.core.config import get_settings

class CacheService:
    """@CODE:USER-AUTH-001:CACHE - 缓存服务"""

    def __init__(self):
        settings = get_settings()
        self.redis_client = redis.Redis(
            host=settings.redis_host,
            port=settings.redis_port,
            decode_responses=True
        )

    def get(self, key: str) -> Optional[Any]:
        """获取缓存值"""
        try:
            value = self.redis_client.get(key)
            return json.loads(value) if value else None
        except Exception:
            return None

    def set(self, key: str, value: Any, expire_seconds: int = 3600) -> bool:
        """设置缓存值"""
        try:
            return self.redis_client.setex(
                key,
                expire_seconds,
                json.dumps(value, default=str)
            )
        except Exception:
            return False

    def delete(self, key: str) -> bool:
        """删除缓存"""
        try:
            return bool(self.redis_client.delete(key))
        except Exception:
            return False

    def clear_user_cache(self, user_id: int) -> bool:
        """清除用户相关缓存"""
        patterns = [
            f"user:{user_id}:*",
            f"auth:user:{user_id}:*"
        ]

        for pattern in patterns:
            keys = self.redis_client.keys(pattern)
            if keys:
                self.redis_client.delete(*keys)

        return True
```

###### 5. 性能优化
```python
# `@CODE:USER-AUTH-001:PERFORMANCE | SPEC: SPEC-USER-AUTH-001.md | TEST: tests/test_auth.py

from functools import lru_cache
from typing import Optional
from src.core.cache import CacheService

class OptimizedAuthService:
    """@CODE:USER-AUTH-001:PERFORMANCE - 优化的认证服务"""

    def __init__(self, db: Session):
        self.db = db
        self.cache = CacheService()

    @lru_cache(maxsize=1000)
    def get_password_requirements(self) -> dict:
        """获取密码要求（缓存结果）"""
        return {
            "min_length": auth_settings.password_min_length,
            "require_uppercase": auth_settings.password_require_uppercase,
            "require_lowercase": auth_settings.password_require_lowercase,
            "require_digits": auth_settings.password_require_digits,
            "require_special": auth_settings.password_require_special,
        }

    def get_user_with_cache(self, user_id: int) -> Optional[User]:
        """从缓存获取用户信息"""
        cache_key = f"user:{user_id}:profile"

        # 尝试从缓存获取
        cached_user = self.cache.get(cache_key)
        if cached_user:
            return User(**cached_user)

        # 从数据库获取并缓存
        user = self.db.query(User).filter(User.id == user_id).first()
        if user:
            user_data = {
                "id": user.id,
                "email": user.email,
                "full_name": user.full_name,
                "is_active": user.is_active,
                "is_verified": user.is_verified
            }
            self.cache.set(cache_key, user_data, expire_seconds=300)  # 5分钟缓存

        return user

    def invalidate_user_cache(self, user_id: int) -> None:
        """使用户缓存失效"""
        self.cache.clear_user_cache(user_id)
        # 清除 LRU 缓存
        self.get_user_with_cache.cache_clear()
```

##### 重构后测试验证
```bash
# 运行完整测试套件
pytest tests/test_auth.py -v --cov=src.auth

# 输出示例：
# 15 tests passed, 0 failed
# Coverage: src/auth 95%
#
# 性能测试
# tests/test_performance.py::test_login_performance PASSED               [  0.0123s]
# tests/test_performance.py::test_registration_performance PASSED         [  0.0156s]
#
# 安全测试
# tests/test_security.py::test_sql_injection_protection PASSED            [  0.0089s]
# tests/test_security.py::test_xss_protection PASSED                     [  0.0067s]
```

**Git 提交 REFACTOR 阶段**：
```bash
git add src/auth/ tests/
git commit -m "♻️ refactor(USER-AUTH-001): improve code quality and performance"
```

### 阶段 3：质量保证与验证

#### TRUST 5 原则验证
Alfred 自动验证代码是否符合 TRUST 5 原则：

```yaml
TRUST 验证结果:
✅ Test First: 测试覆盖率 95% (≥85%)
✅ Readable: 代码风格检查通过
   - 函数长度平均 15 行 (<50)
   - 类复杂度适中
   - 命名清晰明确
✅ Unified: 架构一致性验证通过
   - 遵循仓库模式
   - 统一的错误处理
   - 一致的 API 设计
✅ Secured: 安全检查通过
   - 密码加密存储
   - 输入验证完整
   - JWT 安全实现
✅ Trackable: @TAG 完整性验证通过
   - 所有代码都有 @TAG 标记
   - TAG 链完整无断裂
   - 提交信息规范

TRUST 总分: 96/100 🎉
```

#### 性能基准测试
```python
def run_performance_benchmarks():
    """运行性能基准测试"""

    # 登录性能测试
    login_time = measure_average_time(
        lambda: client.post("/auth/login", json=test_login_data),
        iterations=100
    )

    # 注册性能测试
    registration_time = measure_average_time(
        lambda: client.post("/auth/register", json=test_registration_data),
        iterations=100
    )

    # 令牌验证性能测试
    token_validation_time = measure_average_time(
        lambda: client.get("/auth/me", headers=auth_headers),
        iterations=1000
    )

    return {
        "login_avg_ms": login_time * 1000,
        "registration_avg_ms": registration_time * 1000,
        "token_validation_avg_ms": token_validation_time * 1000,
    }

# 性能测试结果
performance_results = run_performance_benchmarks()
print(f"Login average: {performance_results['login_avg_ms']:.2f}ms")
print(f"Registration average: {performance_results['registration_avg_ms']:.2f}ms")
print(f"Token validation average: {performance_results['token_validation_avg_ms']:.2f}ms")

# 输出：
# Login average: 12.34ms (< 200ms ✅)
# Registration average: 45.67ms (< 500ms ✅)
# Token validation average: 2.89ms (< 10ms ✅)
```

#### 安全扫描
```bash
# 运行安全扫描工具
bandit -r src/auth/
# 输出：无高危漏洞发现

safety check
# 输出：无已知漏洞依赖

semgrep --config=auto src/auth/
# 输出：无安全问题发现
```

---

## 使用示例

### 示例 1：简单 CRUD 功能

#### 用户输入
```bash
/alfred:2-run PRODUCT-001
```

#### Alfred 处理过程
1. **SPEC 分析**：产品管理 CRUD 需求
2. **架构设计**：FastAPI + SQLAlchemy
3. **TDD 执行**：完整的 RED → GREEN → REFACTOR
4. **质量验证**：TRUST 5 原则检查

```yaml
输出结果:
✅ SPEC: PRODUCT-001 分析完成
✅ 架构设计：RESTful API + 仓储模式
✅ TDD 循环：15 个测试用例全部通过
✅ 代码覆盖率：92%
✅ TRUST 评分：94/100

实现内容:
- 产品模型 (Product)
- API 端点 (CRUD)
- 数据验证 (Pydantic)
- 错误处理 (HTTPException)
- 单元测试 (pytest)
- 集成测试 (TestClient)

性能指标:
- 创建产品: 15ms (< 100ms)
- 查询产品: 8ms (< 50ms)
- 更新产品: 12ms (< 100ms)
- 删除产品: 5ms (< 50ms)
```

### 示例 2：复杂业务逻辑

#### 用户输入
```bash
/alfred:2-run ORDER-002
```

#### Alfred 处理过程
1. **复杂度分析**：涉及多个业务实体
2. **专家激活**：backend-expert 参与
3. **分阶段实现**：逐步构建复杂功能
4. **全面测试**：单元测试 + 集成测试

```yaml
输出结果:
✅ SPEC: ORDER-002 分析完成
✅ 专家参与: backend-expert
✅ 复杂度评估: 中等
✅ 实现策略: 分阶段开发
✅ 测试策略: 多层次测试

实现阶段:
阶段 1: 基础订单模型 (已完成)
阶段 2: 订单状态管理 (进行中)
阶段 3: 支付集成 (待开始)
阶段 4: 库存管理 (待开始)

测试覆盖:
- 单元测试: 85%
- 集成测试: 78%
- 端到端测试: 65%
- 总体覆盖率: 76%
```

### 示例 3：性能优化

#### 用户输入
```bash
/alfred:2-run SEARCH-003 --optimize-performance
```

#### Alfred 处理过程
1. **性能需求分析**：高性能搜索功能
2. **技术选型**：Elasticsearch + Redis 缓存
3. **优化策略**：数据库索引、查询优化、缓存策略
4. **性能测试**：负载测试和基准测试

```yaml
输出结果:
✅ SPEC: SEARCH-003 分析完成
✅ 性能目标: 10万 QPS
✅ 技术选型: Elasticsearch + Redis
✅ 优化策略: 多层缓存 + 数据库优化
✅ 性能测试: 通过所有基准

性能优化措施:
- 数据库索引优化
- 查询结果缓存
- 分页和懒加载
- 异步处理
- 连接池优化

基准测试结果:
- 单次搜索: 23ms (< 100ms)
- 并发搜索: 5000 QPS (> 1000 QPS)
- 内存使用: 512MB (< 1GB)
- CPU 使用率: 45% (< 80%)
```

---

## 高级功能

### 1. 增量开发

#### 语法
```bash
# 在现有实现基础上添加功能
/alfred:2-run SPEC-001 --incremental

# 基于特定提交进行开发
/alfred:2-run SPEC-001 --from-commit=abc123
```

#### 处理方式
Alfred 会：
1. 分析现有实现
2. 识别需要修改的部分
3. 保留现有功能
4. 添加新功能
5. 更新相关测试

### 2. 性能优化模式

#### 语法
```bash
# 专注于性能优化
/alfred:2-run SPEC-001 --optimize-performance

# 指定性能目标
/alfred:2-run SPEC-001 --performance-target="1000 QPS"
```

#### 优化策略
- 数据库查询优化
- 缓存策略实施
- 异步处理
- 算法优化
- 资源使用优化

### 3. 安全强化模式

#### 语法
```bash
# 专注于安全性提升
/alfred:2-run SPEC-001 --security-hardening

# 指定安全标准
/alfred:2-run SPEC-001 --security-standard="OWASP"
```

#### 安全措施
- 输入验证强化
- 认证和授权改进
- 数据加密
- 安全审计日志
- 漏洞扫描

### 4. 测试驱动重构

#### 语法
```bash
# 基于测试重构现有代码
/alfred:2-run SPEC-001 --refactor-with-tests

# 指定重构目标
/alfred:2-run SPEC-001 --refactor-target="improve-maintainability"
```

#### 重构流程
1. 分析现有代码
2. 编写缺失测试
3. 逐步重构
4. 保持测试通过
5. 验证改进效果

---

## 最佳实践

### 1. 准备工作

#### 确保 SPEC 完整
```bash
# 在运行 2-run 前验证 SPEC
/alfred:3-sync --verify-specs

# 检查 SPEC 质量
cat .moai/specs/SPEC-XXX/spec.md
```

#### 环境准备
```bash
# 检查项目环境
moai-adk doctor

# 安装必要依赖
uv sync

# 启动开发服务器（如需要）
uvicorn src.main:app --reload
```

### 2. 交互最佳实践

#### 提供明确的指导
```bash
# 指定实现重点
/alfred:2-run AUTH-001 --focus="security"

# 排除特定功能
/alfred:2-run AUTH-001 --exclude="social-login"

# 指定技术约束
/alfred:2-run AUTH-001 --tech-stack="FastAPI, PostgreSQL"
```

#### 及时反馈
```bash
# 对实现方案提供反馈
"这个架构设计很好，但请添加更多的错误处理"
"测试覆盖率足够，但需要增加边界条件测试"
"性能优化方案很全面，请继续实施"
```

### 3. 质量保证

#### 定期验证
```bash
# 每个阶段完成后验证
/alfred:3-sync --trust-check

# 性能基准测试
/alfred:3-sync --performance-test

# 安全扫描
/alfred:3-sync --security-scan
```

#### 代码审查
```bash
# 生成代码审查报告
/alfred:3-sync --code-review

# 检查代码质量指标
/alfred:3-sync --quality-metrics
```

---

## 故障排除

### 常见问题

#### 1. 测试无法通过
**症状**：GREEN 阶段测试仍然失败

**解决方案**：
```bash
# 检查测试错误信息
pytest tests/test_spec.py -v

# 检查实现逻辑
/alfred:2-run SPEC-001 --debug

# 寻求 Alfred 帮助
"测试失败，请帮我检查实现逻辑"
```

#### 2. 性能不达标
**症状**：性能测试未通过基准

**解决方案**：
```bash
# 运行性能分析
/alfred:2-run SPEC-001 --performance-analysis

# 优化建议
/alfred:2-run SPEC-001 --optimize-performance

# 重新运行性能测试
/alfred:3-sync --performance-test
```

#### 3. 安全问题
**症状**：安全扫描发现漏洞

**解决方案**：
```bash
# 安全加固
/alfred:2-run SPEC-001 --security-hardening

# 重新扫描
/alfred:3-sync --security-scan

# 生成安全报告
/alfred:3-sync --security-report
```

### 调试技巧

#### 1. 启用详细日志
```bash
# 启用调试模式
export ALFRED_DEBUG=true
/alfred:2-run SPEC-001 --debug

# 保存调试信息
/alfred:2-run SPEC-001 --debug --output=debug.log
```

#### 2. 分步执行
```bash
# 只执行 RED 阶段
/alfred:2-run SPEC-001 --red-only

# 只执行 GREEN 阶段
/alfred:2-run SPEC-001 --green-only

# 只执行 REFACTOR 阶段
/alfred:2-run SPEC-001 --refactor-only
```

#### 3. 跳过某些步骤
```bash
# 跳过性能测试
/alfred:2-run SPEC-001 --skip-performance

# 跳过安全扫描
/alfred:2-run SPEC-001 --skip-security

# 跳过重构
/alfred:2-run SPEC-001 --skip-refactor
```

---

## 与其他工具的集成

### 与 CI/CD 集成
```yaml
# .github/workflows/tdd.yml
name: TDD Workflow

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  tdd:
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v3

    - name: Set up Python
      uses: actions/setup-python@v4
      with:
        python-version: '3.13'

    - name: Install dependencies
      run: |
        pip install -r requirements.txt
        pip install -r requirements-dev.txt

    - name: Run Alfred TDD
      run: |
        claude --non-interactive "/alfred:2-run ${{ github.event.inputs.spec_id }}"

    - name: Run tests
      run: pytest

    - name: Check coverage
      run: pytest --cov=src --cov-fail-under=85

    - name: Security scan
      run: bandit -r src/
```

### 与代码质量工具集成
```bash
# 集成 pre-commit hooks
pre-commit run --all-files

# 集成代码覆盖率检查
pytest --cov=src --cov-report=xml

# 集成代码质量检查
ruff check src/
mypy src/
```

---

## 总结

`/alfred:2-run` 命令是 MoAI-ADK 开发执行阶段的核心工具，它能够：

- **执行完整 TDD 流程**：RED → GREEN → REFACTOR
- **保证代码质量**：应用 TRUST 5 原则
- **自动化最佳实践**：代码重构、性能优化、安全加固
- **提供全面测试**：单元测试、集成测试、性能测试

### 关键要点

1. **SPEC 驱动**：始终基于 SPEC 进行开发
2. **测试优先**：先写测试，再写实现
3. **持续重构**：不断改进代码质量
4. **全面验证**：确保功能、性能、安全都达标
5. **文档同步**：保持代码和文档的一致性

### 下一步

- [学习 3-sync 命令](3-sync.md)
- [理解测试策略](../tdd/)
- [掌握代码重构](../essentials/refactor.md)
- [查看性能优化](../performance/)

通过熟练使用 `/alfred:2-run` 命令，您可以确保开发出高质量、可测试、可维护的代码，满足所有业务需求和技术标准。