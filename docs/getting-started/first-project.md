# First Project Tutorial: Todo App

> **실전 Todo 앱 만들기**
>
> MoAI-ADK의 SPEC-First TDD 워크플로우를 실제 프로젝트에 적용

---

## 📋 Table of Contents

- [튜토리얼 개요](#튜토리얼-개요)
- [프로젝트 요구사항](#프로젝트-요구사항)
- [Step 1: 프로젝트 설계](#step-1-프로젝트-설계)
- [Step 2: 사용자 인증 구현](#step-2-사용자-인증-구현-auth-001)
- [Step 3: Todo CRUD 구현](#step-3-todo-crud-구현-todo-001)
- [Step 4: API 문서 자동 생성](#step-4-api-문서-자동-생성)
- [Step 5: 프로덕션 배포 준비](#step-5-프로덕션-배포-준비)
- [전체 코드 리뷰](#전체-코드-리뷰)
- [다음 단계](#다음-단계)

---

## 튜토리얼 개요

### 🎯 학습 목표

이 튜토리얼을 완료하면 다음을 배울 수 있습니다:

- ✅ SPEC-First TDD 워크플로우 실전 적용
- ✅ 다중 SPEC 관리 (USER, AUTH, TODO)
- ✅ RESTful API 설계 및 구현
- ✅ JWT 기반 인증 시스템
- ✅ 데이터베이스 통합 (SQLAlchemy)
- ✅ API 문서 자동 생성
- ✅ TRUST 5원칙 적용
- ✅ @TAG 추적성 관리

### 🛠️ 기술 스택

| 계층            | 기술                       |
| --------------- | -------------------------- |
| **언어**        | Python 3.13+               |
| **프레임워크**  | FastAPI 0.104+             |
| **데이터베이스** | SQLite (개발), PostgreSQL (프로덕션) |
| **ORM**         | SQLAlchemy 2.0+            |
| **인증**        | JWT (PyJWT)                |
| **테스트**      | pytest, pytest-asyncio     |
| **린터**        | ruff                       |
| **타입 체크**   | mypy                       |

### ⏱️ 예상 소요 시간

- **초급자**: 60-90분
- **중급자**: 30-45분
- **고급자**: 15-30분

---

## 프로젝트 요구사항

### 기능 요구사항

**사용자 관리**:
- 사용자 회원가입 (이메일, 비밀번호)
- 사용자 로그인 (JWT 토큰 발급)
- 사용자 프로필 조회

**Todo 관리**:
- Todo 생성 (제목, 내용, 마감일)
- Todo 목록 조회 (사용자별 필터링)
- Todo 상세 조회
- Todo 수정 (제목, 내용, 완료 상태)
- Todo 삭제

### 비기능 요구사항

**보안**:
- 비밀번호 bcrypt 해싱
- JWT 토큰 기반 인증
- API 엔드포인트 권한 검증

**품질**:
- 테스트 커버리지 ≥85%
- 타입 안전성 (mypy strict)
- 코드 품질 (ruff 통과)

**성능**:
- API 응답 시간 <200ms
- 데이터베이스 쿼리 최적화

---

## Step 1: 프로젝트 설계

### 1-1. 프로젝트 생성 및 초기화

```bash
# 1. 프로젝트 생성
moai init todo-app
cd todo-app

# 2. Python 가상환경 생성 (uv 사용)
uv venv
source .venv/bin/activate  # macOS/Linux
.venv\Scripts\activate     # Windows

# 3. 의존성 설치
uv pip install fastapi[all] sqlalchemy pyjwt bcrypt pytest pytest-asyncio
```

---

### 1-2. Claude Code에서 프로젝트 컨텍스트 생성

```bash
# Claude Code 실행
claude
```

**Claude Code에서**:

```text
/alfred:0-project
```

**Alfred의 응답 (Phase 1)**:

```
📖 Alfred SuperAgent 시작합니다.

🔍 Phase 1: 프로젝트 분석

1️⃣ 프로젝트 구조:
  - 루트: /Users/goos/todo-app
  - 언어: Python 3.13
  - 프레임워크: FastAPI (감지됨)

2️⃣ 생성할 문서:
  - product.md (Todo 앱 개요)
  - structure.md (API 구조, 모듈 설계)
  - tech.md (FastAPI, SQLAlchemy 도구 체인)

진행하시겠습니까? (진행/수정/중단)
```

**사용자**: `진행`

**Alfred의 응답 (Phase 2)**:

```
✅ Phase 2: 프로젝트 초기화

📝 생성된 문서:
  - .moai/project/product.md (220 lines)
  - .moai/project/structure.md (180 lines)
  - .moai/project/tech.md (200 lines)

🎉 프로젝트 초기화 완료!
```

---

### 1-3. 프로젝트 구조 설계

**Claude Code에서 product.md 확인**:

```bash
# product.md 내용 미리보기
cat .moai/project/product.md
```

**생성된 product.md 예시** (일부):

```markdown
# Todo App - Product Overview

## 제품 개요

REST API 기반 Todo 관리 애플리케이션

## 핵심 기능

### 사용자 인증
- 회원가입 (이메일/비밀번호)
- 로그인 (JWT 토큰)
- 프로필 조회

### Todo 관리
- CRUD 작업
- 사용자별 필터링
- 완료 상태 관리

## 기술 목표

- RESTful API 설계 원칙 준수
- JWT 기반 무상태 인증
- 테스트 커버리지 ≥85%
```

---

## Step 2: 사용자 인증 구현 (AUTH-001)

### 2-1. SPEC 작성: 사용자 회원가입 및 로그인

**Claude Code에서**:

```text
/alfred:1-spec "사용자 회원가입 및 로그인 API"
```

**Alfred의 응답 (Phase 1)**:

```
🏗️ spec-builder 에이전트 시작

📋 Phase 1: SPEC 분석

1️⃣ SPEC ID 생성:
  - 제안 ID: AUTH-001
  - 디렉토리: .moai/specs/SPEC-AUTH-001/

2️⃣ 중복 확인:
  ✅ AUTH-001 사용 가능

3️⃣ EARS 명세 구조:
  - Ubiquitous: 회원가입, 로그인 기능
  - Event-driven: 인증 성공/실패 처리
  - State-driven: 인증 상태 관리
  - Constraints: 비밀번호 보안 정책

진행하시겠습니까? (진행/수정/중단)
```

**사용자**: `진행`

**Alfred의 응답 (Phase 2)**:

```
✅ Phase 2: SPEC 작성

📝 SPEC 문서 생성:
  - .moai/specs/SPEC-AUTH-001/spec.md (320 lines)

📊 SPEC 메타데이터:
  ---
  id: AUTH-001
  version: 0.0.1
  status: draft
  created: 2025-10-14
  updated: 2025-10-14
  author: @Goos
  priority: high
  category: feature
  labels: [authentication, security, api]
  ---

🌿 Git 브랜치 생성:
  ✅ feature/SPEC-AUTH-001

📤 Git 커밋:
  ✅ 🔴 RED: SPEC-AUTH-001 사용자 인증 API 명세

🎉 SPEC 작성 완료!
```

---

### 2-2. SPEC 문서 확인

**생성된 SPEC-AUTH-001/spec.md** (주요 내용):

```yaml
---
id: AUTH-001
version: 0.0.1
status: draft
created: 2025-10-14
updated: 2025-10-14
author: @Goos
priority: high
category: feature
labels: [authentication, security, api]
---

# @SPEC:AUTH-001: 사용자 회원가입 및 로그인 API

## HISTORY
### v0.0.1 (2025-10-14)
- **INITIAL**: 사용자 인증 API 명세 작성
- **AUTHOR**: @Goos

## Overview
JWT 기반 사용자 인증 시스템을 구현한다.

## Requirements (EARS)

### Ubiquitous Requirements
- 시스템은 이메일/비밀번호 기반 회원가입 기능을 제공해야 한다
- 시스템은 JWT 토큰 기반 로그인 기능을 제공해야 한다
- 시스템은 사용자 프로필 조회 기능을 제공해야 한다

### Event-driven Requirements
- WHEN 사용자가 회원가입을 요청하면, 시스템은 이메일 중복을 확인해야 한다
- WHEN 이메일이 중복되면, 시스템은 400 에러를 반환해야 한다
- WHEN 유효한 회원가입 정보가 제공되면, 시스템은 비밀번호를 bcrypt로 해싱하여 저장해야 한다
- WHEN 사용자가 로그인을 요청하면, 시스템은 자격증명을 검증해야 한다
- WHEN 자격증명이 유효하면, 시스템은 JWT 토큰을 발급해야 한다
- WHEN 자격증명이 무효하면, 시스템은 401 에러를 반환해야 한다

### State-driven Requirements
- WHILE 사용자가 인증된 상태일 때, 시스템은 보호된 엔드포인트 접근을 허용해야 한다
- WHILE JWT 토큰이 유효한 동안, 시스템은 재로그인을 요구하지 않아야 한다

### Constraints
- 비밀번호는 최소 8자 이상이어야 한다
- 비밀번호는 영문 대소문자, 숫자, 특수문자를 각각 1개 이상 포함해야 한다
- JWT 토큰 만료시간은 24시간이어야 한다
- bcrypt 해싱 라운드는 12 이상이어야 한다

## API Endpoints

### POST /api/v1/auth/signup
**Request**:
```json
{
  "email": "user@example.com",
  "password": "SecurePass123!",
  "username": "johndoe"
}
```

**Response (201)**:

```json
{
  "id": "uuid",
  "email": "user@example.com",
  "username": "johndoe",
  "created_at": "2025-10-14T10:00:00Z"
}
```

### POST /api/v1/auth/login
**Request**:

```json
{
  "email": "user@example.com",
  "password": "SecurePass123!"
}
```

**Response (200)**:

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "token_type": "bearer",
  "expires_in": 86400
}
```

### GET /api/v1/auth/me
**Headers**:

```
Authorization: Bearer {access_token}
```

**Response (200)**:

```json
{
  "id": "uuid",
  "email": "user@example.com",
  "username": "johndoe",
  "created_at": "2025-10-14T10:00:00Z"
}
```

## Database Schema

### users 테이블
| Column       | Type         | Constraints       |
| ------------ | ------------ | ----------------- |
| id           | UUID         | PRIMARY KEY       |
| email        | VARCHAR(255) | UNIQUE, NOT NULL  |
| username     | VARCHAR(100) | NOT NULL          |
| password_hash | VARCHAR(255) | NOT NULL          |
| created_at   | TIMESTAMP    | NOT NULL          |
| updated_at   | TIMESTAMP    | NOT NULL          |

## Security Requirements
- 비밀번호 bcrypt 해싱 (rounds=12)
- JWT 토큰 HS256 알고리즘
- SQL Injection 방어 (SQLAlchemy ORM)
- XSS 방어 (입력 검증)

## Testing Requirements
- 회원가입 성공 케이스
- 회원가입 실패 케이스 (이메일 중복, 약한 비밀번호)
- 로그인 성공 케이스
- 로그인 실패 케이스 (잘못된 자격증명)
- JWT 토큰 검증 테스트
- 보호된 엔드포인트 접근 테스트

```

---

### 2-3. TDD 구현: RED-GREEN-REFACTOR

**Claude Code에서**:
```text
/alfred:2-build AUTH-001
```

**Alfred의 응답 (Phase 1)**:

```
💎 code-builder 에이전트 시작

📋 Phase 1: TDD 계획

1️⃣ SPEC 읽기:
  ✅ .moai/specs/SPEC-AUTH-001/spec.md
  - 요구사항: 15개
  - API 엔드포인트: 3개
  - 보안 제약사항: 4개

2️⃣ TDD 사이클:
  🔴 RED: 실패하는 테스트
    - tests/auth/test_signup.py (회원가입 테스트)
    - tests/auth/test_login.py (로그인 테스트)
    - tests/auth/test_me.py (프로필 조회 테스트)

  🟢 GREEN: 최소 구현
    - src/models/user.py (User 모델)
    - src/auth/service.py (인증 서비스)
    - src/auth/router.py (API 라우터)
    - src/auth/jwt.py (JWT 유틸)

  🔵 REFACTOR: 품질 개선
    - 타입 힌트 강화
    - 에러 처리 개선
    - TRUST 5원칙 적용

진행하시겠습니까? (진행/수정/중단)
```

**사용자**: `진행`

---

### 🔴 RED: 실패하는 테스트 작성

**Alfred의 응답**:

```
🔴 RED: 테스트 작성

📝 1. tests/auth/test_signup.py (95 lines)
📝 2. tests/auth/test_login.py (85 lines)
📝 3. tests/auth/test_me.py (60 lines)

🧪 테스트 실행:
❌ FAILED tests/auth/test_signup.py::test_signup_success
  ModuleNotFoundError: No module named 'src.auth.service'

✅ RED 단계 완료

📤 Git 커밋:
✅ 🔴 RED: SPEC-AUTH-001 인증 API 테스트 작성
```

**생성된 tests/auth/test_signup.py** (예시):

```python
# @TEST:AUTH-001 | SPEC: SPEC-AUTH-001/spec.md

import pytest
from fastapi.testclient import TestClient
from src.main import app
from src.database import get_db, Base, engine

client = TestClient(app)

@pytest.fixture(autouse=True)
def setup_db():
    """각 테스트 전 데이터베이스 초기화"""
    Base.metadata.create_all(bind=engine)
    yield
    Base.metadata.drop_all(bind=engine)

class TestSignup:
    def test_signup_success(self):
        """유효한 회원가입 요청"""
        response = client.post("/api/v1/auth/signup", json={
            "email": "test@example.com",
            "password": "SecurePass123!",
            "username": "testuser"
        })
        assert response.status_code == 201
        data = response.json()
        assert data["email"] == "test@example.com"
        assert data["username"] == "testuser"
        assert "id" in data
        assert "password" not in data  # 비밀번호 노출 금지

    def test_signup_duplicate_email(self):
        """이메일 중복 시 실패"""
        # 첫 번째 회원가입
        client.post("/api/v1/auth/signup", json={
            "email": "test@example.com",
            "password": "SecurePass123!",
            "username": "testuser1"
        })
        # 두 번째 회원가입 (동일 이메일)
        response = client.post("/api/v1/auth/signup", json={
            "email": "test@example.com",
            "password": "AnotherPass456!",
            "username": "testuser2"
        })
        assert response.status_code == 400
        assert "already exists" in response.json()["detail"].lower()

    def test_signup_weak_password(self):
        """약한 비밀번호로 회원가입 실패"""
        response = client.post("/api/v1/auth/signup", json={
            "email": "test@example.com",
            "password": "weak",
            "username": "testuser"
        })
        assert response.status_code == 400
        assert "password" in response.json()["detail"].lower()

    def test_signup_invalid_email(self):
        """유효하지 않은 이메일 형식"""
        response = client.post("/api/v1/auth/signup", json={
            "email": "invalid-email",
            "password": "SecurePass123!",
            "username": "testuser"
        })
        assert response.status_code == 422  # FastAPI validation error
```

---

#### 🟢 GREEN: 최소 구현으로 테스트 통과

**Alfred의 응답**:

```
🟢 GREEN: 구현 코드 작성

📝 1. src/models/user.py (60 lines)
📝 2. src/auth/service.py (120 lines)
📝 3. src/auth/router.py (80 lines)
📝 4. src/auth/jwt.py (45 lines)
📝 5. src/database.py (30 lines)
📝 6. src/main.py (25 lines)

🧪 테스트 실행:
✅ PASSED tests/auth/test_signup.py::test_signup_success
✅ PASSED tests/auth/test_signup.py::test_signup_duplicate_email
✅ PASSED tests/auth/test_signup.py::test_signup_weak_password
✅ PASSED tests/auth/test_login.py::test_login_success
✅ PASSED tests/auth/test_login.py::test_login_invalid_credentials
✅ PASSED tests/auth/test_me.py::test_get_profile_authenticated
✅ PASSED (12/12 tests)

📊 커버리지:
  - src/auth/service.py: 94%
  - src/auth/router.py: 92%

✅ GREEN 단계 완료

📤 Git 커밋:
✅ 🟢 GREEN: SPEC-AUTH-001 인증 API 구현
```

**생성된 src/auth/service.py** (주요 코드):

```python
# @CODE:AUTH-001 | SPEC: SPEC-AUTH-001/spec.md | TEST: tests/auth/test_*.py

"""
@CODE:AUTH-001: 사용자 인증 서비스

TDD 이력:
- RED: tests/auth/ 테스트 작성
- GREEN: bcrypt + PyJWT 구현
- REFACTOR: 타입 안전성 및 에러 처리 강화
"""

from dataclasses import dataclass
from datetime import datetime, timedelta
import bcrypt
import jwt
from sqlalchemy.orm import Session
from src.models.user import User

@dataclass
class SignupRequest:
    email: str
    password: str
    username: str

@dataclass
class LoginRequest:
    email: str
    password: str

@dataclass
class TokenResponse:
    access_token: str
    token_type: str = "bearer"
    expires_in: int = 86400  # 24시간

class AuthService:
    def __init__(self, db: Session, secret_key: str):
        self.db = db
        self.secret_key = secret_key

    # @CODE:AUTH-001:API: 회원가입
    def signup(self, request: SignupRequest) -> User:
        """사용자 회원가입"""
        # @CODE:AUTH-001:DOMAIN: 이메일 중복 확인
        existing_user = self.db.query(User).filter(
            User.email == request.email
        ).first()
        if existing_user:
            raise ValueError("Email already exists")

        # @CODE:AUTH-001:DOMAIN: 비밀번호 강도 검증
        if not self._validate_password(request.password):
            raise ValueError("Password must be at least 8 characters with uppercase, lowercase, number, and special character")

        # @CODE:AUTH-001:DATA: 비밀번호 해싱
        password_hash = bcrypt.hashpw(
            request.password.encode(),
            bcrypt.gensalt(rounds=12)
        ).decode()

        # @CODE:AUTH-001:DATA: 사용자 생성
        user = User(
            email=request.email,
            username=request.username,
            password_hash=password_hash
        )
        self.db.add(user)
        self.db.commit()
        self.db.refresh(user)
        return user

    # @CODE:AUTH-001:API: 로그인
    def login(self, request: LoginRequest) -> TokenResponse:
        """사용자 로그인 및 JWT 토큰 발급"""
        # @CODE:AUTH-001:DATA: 사용자 조회
        user = self.db.query(User).filter(
            User.email == request.email
        ).first()
        if not user:
            raise ValueError("Invalid credentials")

        # @CODE:AUTH-001:DOMAIN: 비밀번호 검증
        if not bcrypt.checkpw(
            request.password.encode(),
            user.password_hash.encode()
        ):
            raise ValueError("Invalid credentials")

        # @CODE:AUTH-001:API: JWT 토큰 생성
        token = self._generate_token(user)
        return TokenResponse(access_token=token)

    # @CODE:AUTH-001:DOMAIN: 비밀번호 강도 검증
    def _validate_password(self, password: str) -> bool:
        """비밀번호 강도 검증 (최소 8자, 대소문자, 숫자, 특수문자)"""
        if len(password) < 8:
            return False
        has_upper = any(c.isupper() for c in password)
        has_lower = any(c.islower() for c in password)
        has_digit = any(c.isdigit() for c in password)
        has_special = any(c in "!@#$%^&*" for c in password)
        return has_upper and has_lower and has_digit and has_special

    # @CODE:AUTH-001:API: JWT 토큰 생성
    def _generate_token(self, user: User) -> str:
        """JWT 토큰 생성 (만료: 24시간)"""
        payload = {
            "user_id": str(user.id),
            "email": user.email,
            "exp": datetime.utcnow() + timedelta(days=1)
        }
        return jwt.encode(payload, self.secret_key, algorithm="HS256")
```

---

#### 🔵 REFACTOR: 코드 품질 개선

**Alfred의 응답**:

```
🔵 REFACTOR: 코드 품질 개선

📝 1. 타입 힌트 강화 (mypy strict 통과)
✅ mypy src/ --strict: Success

📝 2. 에러 처리 개선 (커스텀 예외)
✅ src/auth/exceptions.py 추가

📝 3. TRUST 5원칙 적용:
  ✅ Test First: 테스트 커버리지 93%
  ✅ Readable: 함수당 평균 12 LOC
  ✅ Unified: 타입 안전성 100%
  ✅ Secured: bcrypt rounds=12, JWT HS256
  ✅ Trackable: @TAG 체인 완전

📝 4. 코드 포맷팅 (ruff)
✅ ruff format src/

🧪 테스트 재실행:
✅ PASSED (12/12 tests)

✅ REFACTOR 단계 완료

📤 Git 커밋:
✅ ♻️ REFACTOR: SPEC-AUTH-001 코드 품질 개선

🎉 TDD 구현 완료!

📊 최종 통계:
  - 테스트 파일: 3개 (240 lines)
  - 구현 파일: 6개 (360 lines)
  - 테스트 커버리지: 93%
  - 타입 안전성: 100%

🎯 다음 단계:
  /alfred:3-sync
```

---

### 2-4. 문서 동기화

**Claude Code에서**:

```text
/alfred:3-sync
```

**Alfred의 응답**:

```
📖 doc-syncer 에이전트 시작

✅ Phase 2: 문서 동기화

📝 1. Living Document 생성:
  ✅ docs/api/auth.md (180 lines)

📝 2. SPEC 버전 업데이트:
  ✅ SPEC-AUTH-001: v0.0.1 → v0.1.0 (completed)

📝 3. TAG 체인 검증:
  ✅ @SPEC:AUTH-001 → @TEST:AUTH-001 → @CODE:AUTH-001 → @DOC:AUTH-001

📊 동기화 리포트:
  ✅ .moai/reports/sync-report-2025-10-14.md

📤 Git 커밋:
✅ 📝 DOCS: SPEC-AUTH-001 문서 동기화

🎉 문서 동기화 완료!
```

---

## Step 3: Todo CRUD 구현 (TODO-001)

### 3-1. SPEC 작성: Todo 관리 API

**Claude Code에서**:

```text
/alfred:1-spec "Todo CRUD API (생성, 조회, 수정, 삭제)"
```

**Alfred의 응답**:

```
🏗️ spec-builder 에이전트 시작

✅ SPEC 작성 완료:
  - .moai/specs/SPEC-TODO-001/spec.md (380 lines)

📊 SPEC 메타데이터:
  id: TODO-001
  version: 0.0.1
  status: draft
  depends_on: [AUTH-001]  # 인증 필요
  priority: high

🌿 Git 브랜치:
  ✅ feature/SPEC-TODO-001

🎉 SPEC 작성 완료!
```

**생성된 SPEC-TODO-001/spec.md** (주요 내용):

```yaml
---
id: TODO-001
version: 0.0.1
status: draft
created: 2025-10-14
updated: 2025-10-14
author: @Goos
priority: high
category: feature
labels: [todo, crud, api]
depends_on: [AUTH-001]  # JWT 인증 필요
---

# @SPEC:TODO-001: Todo CRUD API

## HISTORY
### v0.0.1 (2025-10-14)
- **INITIAL**: Todo CRUD API 명세 작성
- **AUTHOR**: @Goos

## Overview
사용자별 Todo 항목을 관리하는 CRUD API를 구현한다.

## Requirements (EARS)

### Ubiquitous Requirements
- 시스템은 Todo 생성 기능을 제공해야 한다
- 시스템은 Todo 목록 조회 기능을 제공해야 한다
- 시스템은 Todo 상세 조회 기능을 제공해야 한다
- 시스템은 Todo 수정 기능을 제공해야 한다
- 시스템은 Todo 삭제 기능을 제공해야 한다

### Event-driven Requirements
- WHEN 사용자가 Todo를 생성하면, 시스템은 완료 상태를 false로 초기화해야 한다
- WHEN 사용자가 자신의 Todo를 조회하면, 시스템은 해당 Todo를 반환해야 한다
- WHEN 사용자가 다른 사용자의 Todo를 조회하려 하면, 시스템은 403 에러를 반환해야 한다
- WHEN 존재하지 않는 Todo를 조회하면, 시스템은 404 에러를 반환해야 한다

### State-driven Requirements
- WHILE 사용자가 인증된 상태일 때만, 시스템은 Todo API 접근을 허용해야 한다
- WHILE Todo가 삭제되지 않은 상태일 때만, 시스템은 조회를 허용해야 한다

### Constraints
- Todo 제목은 최소 1자, 최대 200자여야 한다
- Todo 내용은 최대 5000자여야 한다
- 사용자는 자신의 Todo만 수정/삭제할 수 있어야 한다

## API Endpoints

### POST /api/v1/todos
**Headers**: `Authorization: Bearer {token}`
**Request**:
```json
{
  "title": "장보기",
  "description": "우유, 빵, 계란 구매",
  "due_date": "2025-10-20T18:00:00Z"
}
```

**Response (201)**:

```json
{
  "id": "uuid",
  "title": "장보기",
  "description": "우유, 빵, 계란 구매",
  "is_completed": false,
  "due_date": "2025-10-20T18:00:00Z",
  "created_at": "2025-10-14T10:00:00Z"
}
```

### GET /api/v1/todos
**Headers**: `Authorization: Bearer {token}`
**Query Parameters**:
- `is_completed`: boolean (옵션)
- `skip`: int (페이지네이션)
- `limit`: int (페이지네이션)

**Response (200)**:

```json
{
  "items": [
    {
      "id": "uuid",
      "title": "장보기",
      "is_completed": false,
      "due_date": "2025-10-20T18:00:00Z"
    }
  ],
  "total": 1
}
```

### GET /api/v1/todos/{todo_id}
### PUT /api/v1/todos/{todo_id}
### DELETE /api/v1/todos/{todo_id}

## Database Schema

### todos 테이블
| Column       | Type         | Constraints           |
| ------------ | ------------ | --------------------- |
| id           | UUID         | PRIMARY KEY           |
| user_id      | UUID         | FOREIGN KEY, NOT NULL |
| title        | VARCHAR(200) | NOT NULL              |
| description  | TEXT         | NULL                  |
| is_completed | BOOLEAN      | DEFAULT false         |
| due_date     | TIMESTAMP    | NULL                  |
| created_at   | TIMESTAMP    | NOT NULL              |
| updated_at   | TIMESTAMP    | NOT NULL              |

```

---

### 3-2. TDD 구현

**Claude Code에서**:
```text
/alfred:2-build TODO-001
```

**Alfred의 응답** (요약):

```
💎 code-builder 에이전트 시작

🔴 RED: 테스트 작성
  ✅ tests/todos/test_create.py (80 lines)
  ✅ tests/todos/test_list.py (90 lines)
  ✅ tests/todos/test_update.py (85 lines)
  ✅ tests/todos/test_delete.py (60 lines)

🟢 GREEN: 구현
  ✅ src/models/todo.py (55 lines)
  ✅ src/todos/service.py (150 lines)
  ✅ src/todos/router.py (120 lines)

🔵 REFACTOR: 품질 개선
  ✅ 타입 힌트, TRUST 5원칙 적용

📊 최종 통계:
  - 테스트: 18/18 통과
  - 커버리지: 91%

🎉 TDD 구현 완료!
```

---

### 3-3. 문서 동기화

```text
/alfred:3-sync
```

**Alfred의 응답**:

```
📖 doc-syncer 에이전트 시작

✅ 문서 동기화 완료:
  - docs/api/todos.md (220 lines)
  - SPEC-TODO-001: v0.1.0 (completed)

🎉 동기화 완료!
```

---

## Step 4: API 문서 자동 생성

### 4-1. FastAPI Swagger UI 확인

```bash
# 개발 서버 실행
uvicorn src.main:app --reload
```

**브라우저에서 접속**:
- Swagger UI: `http://localhost:8000/docs`
- ReDoc: `http://localhost:8000/redoc`

**생성된 API 문서** (자동):
- `GET /api/v1/auth/signup`
- `POST /api/v1/auth/login`
- `GET /api/v1/auth/me`
- `POST /api/v1/todos`
- `GET /api/v1/todos`
- `GET /api/v1/todos/{todo_id}`
- `PUT /api/v1/todos/{todo_id}`
- `DELETE /api/v1/todos/{todo_id}`

---

### 4-2. Living Document 확인

**Alfred가 자동 생성한 문서**:
- `docs/api/auth.md` (인증 API)
- `docs/api/todos.md` (Todo API)

---

## Step 5: 프로덕션 배포 준비

### 5-1. 환경 변수 설정

```bash
# .env 파일 생성
cat > .env << EOF
DATABASE_URL=postgresql://user:password@localhost/todo_db
SECRET_KEY=your-secret-key-here
ENVIRONMENT=production
EOF
```

---

### 5-2. 품질 검증

```bash
# 타입 체크
mypy src/ --strict

# 린터
ruff check src/

# 테스트
pytest --cov=src --cov-report=html

# 커버리지 확인
open htmlcov/index.html
```

---

### 5-3. Docker 컨테이너화 (선택)

```dockerfile
# Dockerfile
FROM python:3.13-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY src/ src/
COPY .env .env

CMD ["uvicorn", "src.main:app", "--host", "0.0.0.0", "--port", "8000"]
```

---

## 전체 코드 리뷰

### 📊 프로젝트 통계

| 항목                | 개수/값 |
| ------------------- | ------- |
| **SPEC 문서**       | 2개     |
| **테스트 파일**     | 7개     |
| **구현 파일**       | 10개    |
| **API 엔드포인트**  | 8개     |
| **테스트 케이스**   | 30개    |
| **테스트 커버리지** | 92%     |
| **총 라인 수**      | ~1,800  |

---

### 📂 최종 디렉토리 구조

```
todo-app/
├── .moai/
│   ├── config.json
│   ├── project/
│   │   ├── product.md
│   │   ├── structure.md
│   │   └── tech.md
│   ├── specs/
│   │   ├── SPEC-AUTH-001/
│   │   │   └── spec.md (v0.1.0, completed)
│   │   └── SPEC-TODO-001/
│   │       └── spec.md (v0.1.0, completed)
│   └── reports/
│       └── sync-report-2025-10-14.md
│
├── src/
│   ├── main.py
│   ├── database.py
│   ├── models/
│   │   ├── user.py
│   │   └── todo.py
│   ├── auth/
│   │   ├── service.py
│   │   ├── router.py
│   │   ├── jwt.py
│   │   └── exceptions.py
│   └── todos/
│       ├── service.py
│       └── router.py
│
├── tests/
│   ├── auth/
│   │   ├── test_signup.py
│   │   ├── test_login.py
│   │   └── test_me.py
│   └── todos/
│       ├── test_create.py
│       ├── test_list.py
│       ├── test_update.py
│       └── test_delete.py
│
├── docs/
│   └── api/
│       ├── auth.md
│       └── todos.md
│
├── .env
├── requirements.txt
├── pyproject.toml
└── README.md
```

---

### 🏷️ TAG 추적성 확인

```bash
# TAG 체인 검증
rg '@(SPEC|TEST|CODE|DOC):' -n .moai/specs/ tests/ src/ docs/
```

**출력 예시**:

```
.moai/specs/SPEC-AUTH-001/spec.md:7:# @SPEC:AUTH-001
.moai/specs/SPEC-TODO-001/spec.md:7:# @SPEC:TODO-001

tests/auth/test_signup.py:1:# @TEST:AUTH-001
tests/todos/test_create.py:1:# @TEST:TODO-001

src/auth/service.py:1:# @CODE:AUTH-001
src/todos/service.py:1:# @CODE:TODO-001

docs/api/auth.md:1:# @DOC:AUTH-001
docs/api/todos.md:1:# @DOC:TODO-001
```

**TAG 체인 완전성**: ✅ 100%

---

## 다음 단계

### 🎓 심화 학습

#### 1. Alfred SuperAgent 마스터
➡️ **[Alfred SuperAgent Guide](https://moai-adk.vercel.app/guides/alfred-superagent/)**

**다루는 내용**:
- 10개 AI 에이전트 조율 전략
- Sequential vs Parallel 실행 패턴
- 온디맨드 에이전트 활용

#### 2. SPEC-First TDD 방법론
➡️ **[SPEC-First TDD Guide](https://moai-adk.vercel.app/guides/spec-first-tdd/)**

**다루는 내용**:
- EARS 5가지 구문 심화
- 복잡한 비즈니스 로직 명세화
- 의존성 그래프 관리

#### 3. TAG 시스템 고급 기법
➡️ **[TAG System Guide](https://moai-adk.vercel.app/guides/tag-system/)**

**다루는 내용**:
- TAG 서브 카테고리 활용 (API, UI, DATA, DOMAIN, INFRA)
- 고아 TAG 자동 복구
- 대규모 프로젝트 TAG 관리

---

### 🚀 실전 프로젝트 아이디어

#### 초급 (1-2주)
- **블로그 플랫폼**: 게시글 CRUD, 댓글, 좋아요
- **북마크 관리**: URL 수집, 태그 분류
- **일정 관리**: 캘린더, 알림

#### 중급 (2-4주)
- **전자상거래**: 상품, 장바구니, 주문, 결제
- **SNS 플랫폼**: 팔로우, 피드, 메시지
- **프로젝트 관리 도구**: 칸반 보드, 스프린트

#### 고급 (4-8주)
- **실시간 채팅**: WebSocket, 멀티룸
- **동영상 스트리밍**: 업로드, 트랜스코딩, CDN
- **검색 엔진**: 크롤링, 인덱싱, 랭킹

---

### 💡 추가 리소스

#### 공식 문서
- **[전체 문서 사이트](https://moai-adk.vercel.app)**
- **[CLI Reference](https://moai-adk.vercel.app/cli/)**
- **[API Reference](https://moai-adk.vercel.app/api/)**

#### 커뮤니티
- **[GitHub Repository](https://github.com/modu-ai/moai-adk)**
- **[GitHub Issues](https://github.com/modu-ai/moai-adk/issues)**
- **[GitHub Discussions](https://github.com/modu-ai/moai-adk/discussions)**

#### 패키지
- **[PyPI Package](https://pypi.org/project/moai-adk/)**

---

## 🎉 축하합니다!

이 튜토리얼을 완료하셨습니다! 이제 당신은:

- ✅ SPEC-First TDD 워크플로우를 실전에 적용할 수 있습니다
- ✅ Alfred와 9개 AI 에이전트를 활용할 수 있습니다
- ✅ RESTful API를 설계하고 구현할 수 있습니다
- ✅ JWT 기반 인증 시스템을 구축할 수 있습니다
- ✅ TRUST 5원칙을 적용한 고품질 코드를 작성할 수 있습니다
- ✅ @TAG 시스템으로 완벽한 추적성을 보장할 수 있습니다

**다음 프로젝트에서 MoAI-ADK를 활용해보세요!** 🚀

---

**마지막 업데이트**: 2025-10-14
**버전**: v0.3.0
